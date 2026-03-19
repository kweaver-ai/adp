package skill

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/dbaccess"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/auth"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/business_domain"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
)

type skillReader struct {
	skillRepo              model.ISkillRepository
	fileRepo               model.ISkillFileIndex
	assetStore             skillAssetStore
	AuthService           interfaces.IAuthorizationService
	BusinessDomainService interfaces.IBusinessDomainService
}

var (
	readerOnce sync.Once
	readerInst interfaces.SkillReader
)

func NewSkillReader() interfaces.SkillReader {
	readerOnce.Do(func() {
		readerInst = &skillReader{
			skillRepo:              dbaccess.NewSkillRepositoryDB(),
			fileRepo:               dbaccess.NewSkillFileIndexDB(),
			assetStore:             newSkillAssetStore(),
			AuthService:            auth.NewAuthServiceImpl(),
			BusinessDomainService: business_domain.NewBusinessDomainService(),
		}
	})
	return readerInst
}

func (r *skillReader) resourceType() interfaces.AuthResourceType {
	return interfaces.AuthResourceTypeSkill
}

func (r *skillReader) ensureBusinessDomainVisible(ctx context.Context, businessDomainID, skillID string) error {
	resourceToBdMap, err := r.BusinessDomainService.BatchResourceList(ctx, strings.Split(businessDomainID, ","), r.resourceType())
	if err != nil {
		return err
	}
	if _, ok := resourceToBdMap[skillID]; !ok {
		return fmt.Errorf("skill not found: %s", skillID)
	}
	return nil
}

func (r *skillReader) GetSkillContent(ctx context.Context, req *interfaces.GetSkillContentReq) (*interfaces.GetSkillContentResp, error) {
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
	if err = r.ensureBusinessDomainVisible(ctx, req.BusinessDomainID, req.SkillID); err != nil {
		return nil, err
	}
	return &interfaces.GetSkillContentResp{
		SkillID:      skill.SkillID,
		SkillContent: skill.SkillContent,
		Files:        utils.JSONToObject[[]*interfaces.SkillFileSummary](skill.FileManifest),
		Status:       skill.Status,
	}, nil
}

func (r *skillReader) ReadSkillFile(ctx context.Context, req *interfaces.ReadSkillFileReq) (*interfaces.ReadSkillFileResp, error) {
	skill, err := r.skillRepo.SelectSkillByID(ctx, nil, req.SkillID)
	if err != nil {
		return nil, err
	}
	if skill == nil || skill.Status != model.SkillStatusPublished {
		return nil, fmt.Errorf("skill not found: %s", req.SkillID)
	}
	accessor, err := r.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if err = r.AuthService.CheckExecutePermission(ctx, accessor, req.SkillID, r.resourceType()); err != nil {
		return nil, err
	}
	if err = r.ensureBusinessDomainVisible(ctx, req.BusinessDomainID, req.SkillID); err != nil {
		return nil, err
	}

	relPath, err := normalizeZipPath(req.RelPath)
	if err != nil {
		return nil, err
	}
	file, err := r.fileRepo.SelectSkillFileByPath(ctx, nil, req.SkillID, relPath)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, fmt.Errorf("skill file not found: %s", relPath)
	}

	content, err := r.assetStore.Read(ctx, file.StorageKey)
	if err != nil {
		return nil, err
	}
	if checksumSHA256(content) != file.ContentSHA256 {
		return nil, fmt.Errorf("skill file checksum mismatch: %s", relPath)
	}

	return &interfaces.ReadSkillFileResp{
		SkillID:  req.SkillID,
		RelPath:  relPath,
		Content:  string(content),
		MimeType: file.MimeType,
		FileType: file.FileType,
	}, nil
}

func isSkillDeletable(status model.SkillStatus) bool {
	return status == model.SkillStatusUnpublish || status == model.SkillStatusPublished || status == model.SkillStatusOffline
}
