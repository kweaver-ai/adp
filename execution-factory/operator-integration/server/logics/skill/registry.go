package skill

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/dbaccess"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/auth"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/business_domain"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
)

type skillRegistry struct {
	parser                *skillParser
	skillRepo             model.ISkillRepository
	fileRepo              model.ISkillFileIndex
	assetStore            skillAssetStore
	dbTx                  model.DBTx
	AuthService           interfaces.IAuthorizationService
	BusinessDomainService interfaces.IBusinessDomainService
}

var (
	registryOnce sync.Once
	registryInst interfaces.SkillRegistry
)

func NewSkillRegistry() interfaces.SkillRegistry {
	registryOnce.Do(func() {
		registryInst = &skillRegistry{
			parser:                newSkillParser(),
			skillRepo:             dbaccess.NewSkillRepositoryDB(),
			fileRepo:              dbaccess.NewSkillFileIndexDB(),
			assetStore:            newSkillAssetStore(),
			dbTx:                  dbaccess.NewBaseTx(),
			AuthService:           auth.NewAuthServiceImpl(),
			BusinessDomainService: business_domain.NewBusinessDomainService(),
		}
	})
	return registryInst
}

func (r *skillRegistry) resourceType() interfaces.AuthResourceType {
	return interfaces.AuthResourceTypeSkill
}

func (r *skillRegistry) associateBusinessDomain(ctx context.Context, businessDomainID, skillID string) error {
	return r.BusinessDomainService.AssociateResource(ctx, businessDomainID, skillID, r.resourceType())
}

func (r *skillRegistry) batchDisassociateBusinessDomain(ctx context.Context, businessDomainID string, skillIDs []string) error {
	return r.BusinessDomainService.BatchDisassociateResource(ctx, businessDomainID, skillIDs, r.resourceType())
}

func (r *skillRegistry) listMarketVisibleIDs(ctx context.Context, accessor *interfaces.AuthAccessor) ([]string, error) {
	return r.AuthService.ResourceListIDs(ctx, accessor, r.resourceType(), interfaces.AuthOperationTypePublicAccess)
}

func (r *skillRegistry) filterViewableIDs(ctx context.Context, accessor *interfaces.AuthAccessor, skillIDs []string) ([]string, error) {
	return r.AuthService.ResourceFilterIDs(ctx, accessor, skillIDs, r.resourceType(), interfaces.AuthOperationTypeView)
}

func (r *skillRegistry) RegisterSkill(ctx context.Context, req *interfaces.RegisterSkillReq) (*interfaces.RegisterSkillResp, error) {
	accessor, err := r.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if err = r.AuthService.CheckCreatePermission(ctx, accessor, r.resourceType()); err != nil {
		return nil, err
	}

	skill, files, assets, err := r.parser.parseRegisterReq(req)
	if err != nil {
		return nil, err
	}
	skill.FileManifest = utils.ObjectToJSON(files)

	tx, err := r.dbTx.GetTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("get tx failed: %w", err)
	}
	defer func() {
		if tx == nil {
			return
		}
		if err != nil {
			_ = tx.Rollback()
			return
		}
		_ = tx.Commit()
	}()

	skillID, err := r.skillRepo.InsertSkill(ctx, tx, skill)
	if err != nil {
		return nil, err
	}
	skill.Status = model.SkillStatusActive
	if err = r.skillRepo.UpdateSkillStatus(ctx, tx, skillID, model.SkillStatusActive, req.UserID); err != nil {
		return nil, err
	}
	if len(assets) > 0 {
		fileIndices, buildErr := r.persistSkillAssets(ctx, skillID, assets)
		if buildErr != nil {
			err = buildErr
			return nil, err
		}
		if err = r.fileRepo.BatchInsertSkillFiles(ctx, tx, fileIndices); err != nil {
			return nil, err
		}
	}
	if err = r.associateBusinessDomain(ctx, req.BusinessDomainID, skillID); err != nil {
		return nil, err
	}

	filePaths := make([]string, 0, len(files))
	for _, file := range files {
		filePaths = append(filePaths, file.RelPath)
	}
	return &interfaces.RegisterSkillResp{
		SkillID:     skillID,
		Name:        skill.Name,
		Description: skill.Description,
		Version:     skill.Version,
		Status:      model.SkillStatusActive,
		Files:       filePaths,
	}, nil
}

func (r *skillRegistry) DeleteSkill(ctx context.Context, req *interfaces.DeleteSkillReq) error {
	skill, err := r.skillRepo.SelectSkillByID(ctx, nil, req.SkillID)
	if err != nil {
		return err
	}
	if skill == nil {
		return fmt.Errorf("skill not found: %s", req.SkillID)
	}
	if !isSkillDeletable(skill.Status) {
		return fmt.Errorf("skill can not be deleted in status: %s", skill.Status)
	}
	accessor, err := r.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return err
	}
	if err = r.AuthService.CheckDeletePermission(ctx, accessor, req.SkillID, r.resourceType()); err != nil {
		return err
	}

	tx, err := r.dbTx.GetTx(ctx)
	if err != nil {
		return fmt.Errorf("get tx failed: %w", err)
	}
	defer func() {
		if tx == nil {
			return
		}
		if err != nil {
			_ = tx.Rollback()
			return
		}
		_ = tx.Commit()
	}()

	if err = r.skillRepo.UpdateSkillStatus(ctx, tx, req.SkillID, model.SkillStatusDeleting, req.UserID); err != nil {
		return err
	}

	files, err := r.fileRepo.SelectSkillFileBySkillID(ctx, tx, req.SkillID)
	if err != nil {
		return err
	}
	for _, file := range files {
		if removeErr := r.assetStore.DeleteFile(ctx, file.StorageKey); removeErr != nil {
			return removeErr
		}
	}
	if err = r.assetStore.DeleteSkill(ctx, req.SkillID); err != nil {
		return err
	}
	if err = r.fileRepo.DeleteSkillFileBySkillID(ctx, tx, req.SkillID); err != nil {
		return err
	}
	if err = r.skillRepo.DeleteSkillByID(ctx, tx, req.SkillID); err != nil {
		return err
	}
	if err = r.batchDisassociateBusinessDomain(ctx, req.BusinessDomainID, []string{req.SkillID}); err != nil {
		return err
	}
	return r.AuthService.DeletePolicy(ctx, []string{req.SkillID}, r.resourceType())
}

func (r *skillRegistry) DownloadSkill(ctx context.Context, req *interfaces.DownloadSkillReq) (*interfaces.DownloadSkillResp, error) {
	skill, err := r.skillRepo.SelectSkillByID(ctx, nil, req.SkillID)
	if err != nil {
		return nil, err
	}
	if skill == nil || skill.Status == model.SkillStatusDeleting {
		return nil, fmt.Errorf("skill not found: %s", req.SkillID)
	}
	accessor, err := r.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if err = r.AuthService.CheckViewPermission(ctx, accessor, req.SkillID, r.resourceType()); err != nil {
		return nil, err
	}

	files, err := r.fileRepo.SelectSkillFileBySkillID(ctx, nil, req.SkillID)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	writeFile := func(name string, content []byte) error {
		w, createErr := zw.Create(name)
		if createErr != nil {
			return createErr
		}
		_, writeErr := io.Copy(w, bytes.NewReader(content))
		return writeErr
	}

	skillMarkdown := fmt.Sprintf("---\nname: %s\ndescription: %s\nversion: %s\n---\n%s",
		skill.Name, skill.Description, skill.Version, skill.SkillContent)
	if err = writeFile("SKILL.md", []byte(skillMarkdown)); err != nil {
		_ = zw.Close()
		return nil, err
	}
	for _, file := range files {
		content, readErr := r.assetStore.Read(ctx, file.StorageKey)
		if readErr != nil {
			_ = zw.Close()
			return nil, readErr
		}
		if err = writeFile(file.RelPath, content); err != nil {
			_ = zw.Close()
			return nil, err
		}
	}
	if err = zw.Close(); err != nil {
		return nil, err
	}

	return &interfaces.DownloadSkillResp{
		SkillID:  skill.SkillID,
		FileName: fmt.Sprintf("%s.zip", skill.Name),
		Content:  buf.Bytes(),
	}, nil
}

func (r *skillRegistry) QuerySkillList(ctx context.Context, req *interfaces.QuerySkillListReq) (*interfaces.QuerySkillListResp, error) {
	accessor, err := r.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	filter := map[string]interface{}{
		"name":   req.Name,
		"source": req.Source,
		"all":    req.All,
		"limit":  req.PageSize,
		"offset": (req.Page - 1) * req.PageSize,
	}
	if req.CreateUser != "" {
		filter["create_user"] = req.CreateUser
	}
	if req.Status != "" {
		filter["status"] = string(req.Status)
	}

	sortField := "f_update_time"
	switch req.SortBy {
	case "create_time":
		sortField = "f_create_time"
	case "name":
		sortField = "f_name"
	}
	sortOrder := ormhelper.SortOrderDesc
	if req.SortOrder == "asc" {
		sortOrder = ormhelper.SortOrderAsc
	}

	sort := &ormhelper.SortParams{Fields: []ormhelper.SortField{{Field: sortField, Order: sortOrder}}}
	total, err := r.skillRepo.CountByWhereClause(ctx, nil, filter)
	if err != nil {
		return nil, err
	}
	skills, err := r.skillRepo.SelectSkillListPage(ctx, nil, filter, sort, nil)
	if err != nil {
		return nil, err
	}

	viewableCandidates := make([]string, 0, len(skills))
	visibleSkillMap := make(map[string]*model.SkillRepositoryDB, len(skills))
	for _, skill := range skills {
		if skill.Status == model.SkillStatusDeleting {
			continue
		}
		viewableCandidates = append(viewableCandidates, skill.SkillID)
		visibleSkillMap[skill.SkillID] = skill
	}
	viewableIDs, err := r.filterViewableIDs(ctx, accessor, viewableCandidates)
	if err != nil {
		return nil, err
	}

	data := make([]*interfaces.SkillSummary, 0, len(skills))
	for _, skillID := range viewableIDs {
		skill, ok := visibleSkillMap[skillID]
		if !ok {
			continue
		}
		data = append(data, convertSkillSummary(skill))
	}

	pageResult := ormhelper.CalculateQueryResult(total, &ormhelper.PaginationParams{
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	return &interfaces.QuerySkillListResp{
		CommonPageResult: interfaces.CommonPageResult{
			TotalCount: int(pageResult.Total),
			Page:       pageResult.Page,
			PageSize:   pageResult.PageSize,
			TotalPage:  pageResult.TotalPages,
			HasNext:    pageResult.HasNext,
			HasPrev:    pageResult.HasPrev,
		},
		Data: data,
	}, nil
}

func (r *skillRegistry) QuerySkillMarketList(ctx context.Context, req *interfaces.QuerySkillMarketListReq) (*interfaces.QuerySkillMarketListResp, error) {
	accessor, err := r.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	publicIDs, err := r.listMarketVisibleIDs(ctx, accessor)
	if err != nil {
		return nil, err
	}
	resourceToBdMap, err := r.BusinessDomainService.BatchResourceList(ctx, strings.Split(req.BusinessDomainID, ","), r.resourceType())
	if err != nil {
		return nil, err
	}

	filter := map[string]interface{}{
		"name":   req.Name,
		"source": req.Source,
		"all":    true,
	}
	sort := &ormhelper.SortParams{Fields: []ormhelper.SortField{{Field: "f_update_time", Order: ormhelper.SortOrderDesc}}}
	skills, err := r.skillRepo.SelectSkillListPage(ctx, nil, filter, sort, nil)
	if err != nil {
		return nil, err
	}

	publicSet := make(map[string]struct{}, len(publicIDs))
	for _, skillID := range publicIDs {
		publicSet[skillID] = struct{}{}
	}

	filtered := make([]*interfaces.SkillSummary, 0, len(skills))
	for _, skill := range skills {
		if skill.Status == model.SkillStatusDeleting {
			continue
		}
		if _, ok := publicSet[skill.SkillID]; !ok {
			continue
		}
		if _, ok := resourceToBdMap[skill.SkillID]; !ok {
			continue
		}
		filtered = append(filtered, convertSkillSummary(skill))
	}

	start := (req.Page - 1) * req.PageSize
	if start < 0 {
		start = 0
	}
	end := start + req.PageSize
	if req.All {
		start = 0
		end = len(filtered)
	}
	if start > len(filtered) {
		start = len(filtered)
	}
	if end > len(filtered) {
		end = len(filtered)
	}
	pageData := filtered[start:end]

	pageResult := ormhelper.CalculateQueryResult(int64(len(filtered)), &ormhelper.PaginationParams{
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	return &interfaces.QuerySkillMarketListResp{
		CommonPageResult: interfaces.CommonPageResult{
			TotalCount: int(pageResult.Total),
			Page:       pageResult.Page,
			PageSize:   pageResult.PageSize,
			TotalPage:  pageResult.TotalPages,
			HasNext:    pageResult.HasNext,
			HasPrev:    pageResult.HasPrev,
		},
		Data: pageData,
	}, nil
}

func (r *skillRegistry) GetSkillMarketDetail(ctx context.Context, req *interfaces.GetSkillMarketDetailReq) (*interfaces.SkillInfo, error) {
	skill, err := r.skillRepo.SelectSkillByID(ctx, nil, req.SkillID)
	if err != nil {
		return nil, err
	}
	if skill == nil || skill.Status == model.SkillStatusDeleting {
		return nil, fmt.Errorf("skill not found: %s", req.SkillID)
	}
	accessor, err := r.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if err = r.AuthService.CheckPublicAccessPermission(ctx, accessor, req.SkillID, r.resourceType()); err != nil {
		return nil, err
	}
	resourceToBdMap, err := r.BusinessDomainService.BatchResourceList(ctx, strings.Split(req.BusinessDomainID, ","), r.resourceType())
	if err != nil {
		return nil, err
	}
	if _, ok := resourceToBdMap[req.SkillID]; !ok {
		return nil, fmt.Errorf("skill not found: %s", req.SkillID)
	}
	return convertSkillDetail(skill), nil
}

func (r *skillRegistry) GetSkillDetail(ctx context.Context, req *interfaces.GetSkillDetailReq) (*interfaces.SkillInfo, error) {
	skill, err := r.skillRepo.SelectSkillByID(ctx, nil, req.SkillID)
	if err != nil {
		return nil, err
	}
	if skill == nil {
		return nil, fmt.Errorf("skill not found: %s", req.SkillID)
	}
	if skill.Status == model.SkillStatusDeleting {
		return nil, fmt.Errorf("skill not found: %s", req.SkillID)
	}
	accessor, err := r.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if err = r.AuthService.CheckViewPermission(ctx, accessor, req.SkillID, r.resourceType()); err != nil {
		return nil, err
	}
	return convertSkillDetail(skill), nil
}

func convertSkillDetail(skill *model.SkillRepositoryDB) *interfaces.SkillInfo {
	return &interfaces.SkillInfo{
		SkillID:      skill.SkillID,
		Name:         skill.Name,
		Description:  skill.Description,
		Version:      skill.Version,
		Status:       skill.Status,
		Source:       skill.Source,
		Dependencies: utils.JSONToObject[map[string]interface{}](skill.Dependencies),
		ExtendInfo:   utils.JSONToObject[map[string]interface{}](skill.ExtendInfo),
		CreateUser:   skill.CreateUser,
		CreateTime:   skill.CreateTime,
		UpdateUser:   skill.UpdateUser,
		UpdateTime:   skill.UpdateTime,
	}
}

func convertSkillSummary(skill *model.SkillRepositoryDB) *interfaces.SkillSummary {
	return &interfaces.SkillSummary{
		SkillID:     skill.SkillID,
		Name:        skill.Name,
		Description: skill.Description,
		Version:     skill.Version,
		Status:      skill.Status,
		Source:      skill.Source,
		CreateUser:  skill.CreateUser,
		CreateTime:  skill.CreateTime,
		UpdateUser:  skill.UpdateUser,
		UpdateTime:  skill.UpdateTime,
	}
}

func (r *skillRegistry) persistSkillAssets(ctx context.Context, skillID string, assets []*skillAsset) ([]*model.SkillFileIndexDB, error) {
	indices := make([]*model.SkillFileIndexDB, 0, len(assets))
	for _, asset := range assets {
		storageKey, checksum, err := r.assetStore.Write(ctx, skillID, asset.RelPath, asset.Content)
		if err != nil {
			return nil, err
		}
		indices = append(indices, &model.SkillFileIndexDB{
			SkillID:       skillID,
			RelPath:       asset.RelPath,
			PathHash:      utils.MD5(asset.RelPath),
			StorageKey:    storageKey,
			FileType:      asset.FileType,
			ContentSHA256: checksum,
			MimeType:      asset.MimeType,
			AccessLevel:   string(interfaces.SkillFileAccessLevelRuntimeRead),
			Size:          int64(len(asset.Content)),
		})
	}
	return indices, nil
}
