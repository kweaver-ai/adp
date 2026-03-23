package skill

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"path/filepath"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/drivenadapters"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
)

//go:generate mockgen -source=asset_store.go -destination=../../mocks/skill_asset_store.go -package=mocks

type skillAssetStore interface {
	Upload(ctx context.Context, skillID, relPath string, content []byte) (object *interfaces.OssObject, checksum string, err error)
	Download(ctx context.Context, object *interfaces.OssObject) ([]byte, error)
	Delete(ctx context.Context, object *interfaces.OssObject) error
	GetDownloadURL(ctx context.Context, object *interfaces.OssObject) (string, error)
}

type ossGatewaySkillAssetStore struct {
	client interfaces.OSSGatewayBackendClient
}

const skillAssetObjectPrefix = "aoi_skill_assets"

func newOSSGatewaySkillAssetStore() skillAssetStore {
	return &ossGatewaySkillAssetStore{client: drivenadapters.NewOSSGatewayBackendClient()}
}

// Upload 上传技能资产到 OSS 网关后端
func (s *ossGatewaySkillAssetStore) Upload(ctx context.Context, skillID, relPath string, content []byte) (object *interfaces.OssObject, checksum string, err error) {
	// 检查服务是否ready，否则返回报错
	if !s.client.IsReady() {
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, "oss gateway backend is not ready")
		return
	}
	key := buildObjectKey(skillID, relPath)
	storageID, err := s.client.CurrentStorageID(ctx)
	if err != nil {
		return
	}
	object = &interfaces.OssObject{
		StorageID:  storageID,
		StorageKey: key,
	}
	if err = s.client.UploadFile(ctx, object, content); err != nil {
		return
	}
	return object, checksumSHA256(content), nil
}

func (s *ossGatewaySkillAssetStore) Download(ctx context.Context, object *interfaces.OssObject) ([]byte, error) {
	// 检查服务是否ready，否则返回报错
	if !s.client.IsReady() {
		return nil, errors.DefaultHTTPError(ctx, http.StatusInternalServerError, "oss gateway backend is not ready")
	}
	return s.client.DownloadFile(ctx, object)
}

func (s *ossGatewaySkillAssetStore) Delete(ctx context.Context, object *interfaces.OssObject) error {
	// 检查服务是否ready，否则返回报错
	if !s.client.IsReady() {
		return errors.DefaultHTTPError(ctx, http.StatusInternalServerError, "oss gateway backend is not ready")
	}
	return s.client.DeleteFile(ctx, object)
}

func (s *ossGatewaySkillAssetStore) GetDownloadURL(ctx context.Context, object *interfaces.OssObject) (string, error) {
	// 检查服务是否ready，否则返回报错
	if !s.client.IsReady() {
		return "", errors.DefaultHTTPError(ctx, http.StatusInternalServerError, "oss gateway backend is not ready")
	}
	return s.client.GetDownloadURL(ctx, object)
}

func buildObjectKey(skillID, relPath string) string {
	if relPath == "" {
		return filepath.ToSlash(filepath.Join(skillAssetObjectPrefix, skillID))
	}
	return filepath.ToSlash(filepath.Join(skillAssetObjectPrefix, skillID, relPath))
}

func checksumSHA256(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}
