package skill

import (
	"context"
	"fmt"
	"sync"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/dbaccess"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
)

type skillReader struct {
	skillRepo  model.ISkillRepository
	fileRepo   model.ISkillFileIndex
	assetStore skillAssetStore
}

var (
	readerOnce sync.Once
	readerInst interfaces.SkillReader
)

func NewSkillReader() interfaces.SkillReader {
	readerOnce.Do(func() {
		readerInst = &skillReader{
			skillRepo:  dbaccess.NewSkillRepositoryDB(),
			fileRepo:   dbaccess.NewSkillFileIndexDB(),
			assetStore: newSkillAssetStore(),
		}
	})
	return readerInst
}

func (r *skillReader) GetSkillGuide(ctx context.Context, req *interfaces.GetSkillGuideReq) (*interfaces.GetSkillGuideResp, error) {
	skill, err := r.skillRepo.SelectSkillByID(ctx, nil, req.SkillID)
	if err != nil {
		return nil, err
	}
	if skill == nil || skill.OwnerID != req.BusinessDomainID || skill.Status == model.SkillStatusDeleted {
		return nil, fmt.Errorf("skill not found: %s", req.SkillID)
	}
	return &interfaces.GetSkillGuideResp{
		SkillID: skill.SkillID,
		Content: skill.Instructions,
		Files:   utils.JSONToObject[[]*interfaces.SkillFileSummary](skill.FileManifest),
		Status:  skill.Status,
	}, nil
}

func (r *skillReader) ReadSkillFile(ctx context.Context, req *interfaces.ReadSkillFileReq) (*interfaces.ReadSkillFileResp, error) {
	skill, err := r.skillRepo.SelectSkillByID(ctx, nil, req.SkillID)
	if err != nil {
		return nil, err
	}
	if skill == nil || skill.OwnerID != req.BusinessDomainID || skill.Status != model.SkillStatusActive {
		return nil, fmt.Errorf("skill not found: %s", req.SkillID)
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
	if file.AccessLevel == string(interfaces.SkillFileAccessLevelRestricted) {
		return nil, fmt.Errorf("skill file access denied: %s", relPath)
	}

	content, err := r.assetStore.Read(ctx, file.StorageKey)
	if err != nil {
		return nil, err
	}
	if checksumSHA256(content) != file.ContentSHA256 {
		return nil, fmt.Errorf("skill file checksum mismatch: %s", relPath)
	}

	return &interfaces.ReadSkillFileResp{
		SkillID:     req.SkillID,
		RelPath:     relPath,
		Content:     string(content),
		MimeType:    file.MimeType,
		FileType:    file.FileType,
		AccessLevel: file.AccessLevel,
	}, nil
}

func isSkillDeletable(status model.SkillStatus) bool {
	return status == model.SkillStatusDraft || status == model.SkillStatusActive || status == model.SkillStatusError
}
