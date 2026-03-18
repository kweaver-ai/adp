package skill

import (
	"context"
	"fmt"
	"sync"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/dbaccess"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
)

type skillRegistry struct {
	parser     *skillParser
	skillRepo  model.ISkillRepository
	fileRepo   model.ISkillFileIndex
	assetStore skillAssetStore
	dbTx       model.DBTx
}

var (
	registryOnce sync.Once
	registryInst interfaces.SkillRegistry
)

func NewSkillRegistry() interfaces.SkillRegistry {
	registryOnce.Do(func() {
		registryInst = &skillRegistry{
			parser:     newSkillParser(),
			skillRepo:  dbaccess.NewSkillRepositoryDB(),
			fileRepo:   dbaccess.NewSkillFileIndexDB(),
			assetStore: newSkillAssetStore(),
			dbTx:       dbaccess.NewBaseTx(),
		}
	})
	return registryInst
}

func (r *skillRegistry) RegisterSkill(ctx context.Context, req *interfaces.RegisterSkillReq) (*interfaces.RegisterSkillResp, error) {
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
	return r.skillRepo.DeleteSkillByID(ctx, tx, req.SkillID)
}

func (r *skillRegistry) QuerySkillList(ctx context.Context, req *interfaces.QuerySkillListReq) (*interfaces.QuerySkillListResp, error) {
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

	data := make([]*interfaces.SkillSummary, 0, len(skills))
	for _, skill := range skills {
		if skill.Status == model.SkillStatusDeleting {
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
