package skill

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/dbaccess"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/auth"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/business_domain"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
)

type skillReader struct {
	skillRepo             model.ISkillRepository
	fileRepo              model.ISkillFileIndex
	assetStore            skillAssetStore
	AuthService           interfaces.IAuthorizationService
	BusinessDomainService interfaces.IBusinessDomainService
}

var (
	readerOnce sync.Once
	readerInst interfaces.SkillReader
)

// NewSkillReader 创建技能读取服务对象
func NewSkillReader() interfaces.SkillReader {
	readerOnce.Do(func() {
		readerInst = &skillReader{
			skillRepo:             dbaccess.NewSkillRepositoryDB(),
			fileRepo:              dbaccess.NewSkillFileIndexDB(),
			assetStore:            newSkillAssetStore(),
			AuthService:           auth.NewAuthServiceImpl(),
			BusinessDomainService: business_domain.NewBusinessDomainService(),
		}
	})
	return readerInst
}

// ensureBusinessDomainVisible 确保技能在业务域中可见
// func (r *skillReader) ensureBusinessDomainVisible(ctx context.Context, businessDomainID, skillID string) error {
// 	resourceToBdMap, err := r.BusinessDomainService.BatchResourceList(ctx, strings.Split(businessDomainID, ","), interfaces.AuthResourceTypeSkill)
// 	if err != nil {
// 		return err
// 	}
// 	if _, ok := resourceToBdMap[skillID]; !ok {
// 		err = errors.DefaultHTTPError(ctx, http.StatusNotFound, fmt.Sprintf("skill not found: %s", skillID))
// 		return err
// 	}
// 	return nil
// }

// GetSkillContent 获取技能内容
func (r *skillReader) GetSkillContent(ctx context.Context, req *interfaces.GetSkillContentReq) (resp *interfaces.GetSkillContentResp, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	skill, err := r.skillRepo.SelectSkillByID(ctx, nil, req.SkillID)
	if err != nil {
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	if skill == nil || skill.Status == model.SkillStatusDeleting {
		err = errors.DefaultHTTPError(ctx, http.StatusNotFound, fmt.Sprintf("skill not found: %s", req.SkillID))
		return
	}
	accessor, err := r.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return
	}
	if err = r.AuthService.CheckViewPermission(ctx, accessor, req.SkillID, interfaces.AuthResourceTypeSkill); err != nil {
		return
	}
	// if err = r.ensureBusinessDomainVisible(ctx, req.BusinessDomainID, req.SkillID); err != nil {
	// 	return nil, err
	// }
	// TODO: 待接入审计日志
	return &interfaces.GetSkillContentResp{
		SkillID:      skill.SkillID,
		SkillContent: skill.SkillContent,
		Files:        utils.JSONToObject[[]*interfaces.SkillFileSummary](skill.FileManifest),
		Status:       skill.Status,
	}, nil
}

// ReadSkillFile 读取技能文件内容
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
	if err = r.AuthService.CheckExecutePermission(ctx, accessor, req.SkillID, interfaces.AuthResourceTypeSkill); err != nil {
		return nil, err
	}
	// if err = r.ensureBusinessDomainVisible(ctx, req.BusinessDomainID, req.SkillID); err != nil {
	// 	return nil, err
	// }

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

// func isSkillDeletable(status model.SkillStatus) bool {
// 	// 只有未发布、已发布、已下线的技能才可以删除
// 	return status == model.SkillStatusUnpublish || status == model.SkillStatusPublished || status == model.SkillStatusOffline
// }
