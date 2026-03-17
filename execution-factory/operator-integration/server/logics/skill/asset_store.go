package skill

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	restinfra "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/rest"
)

//go:generate mockgen -source=asset_store.go -destination=../../mocks/skill_asset_store.go -package=mocks

type skillAssetStore interface {
	Write(ctx context.Context, skillID, relPath string, content []byte) (storageKey string, checksum string, err error)
	Read(ctx context.Context, storageKey string) ([]byte, error)
	DeleteFile(ctx context.Context, storageKey string) error
	DeleteSkill(ctx context.Context, skillID string) error
}

type s3SkillAssetStore struct {
	client *s3.S3
	bucket string
	prefix string
}

type localSkillAssetStore struct{}

var loadFormalSkillAssetStoreConfig = func() (config.S3Config, bool) {
	if !formalConfigFilesExist() {
		return config.S3Config{}, false
	}
	return config.NewConfigLoader().S3Config, true
}

func newSkillAssetStore() skillAssetStore {
	if store, err := newSkillAssetStoreFromFormalConfig(); err == nil && store != nil {
		return store
	}
	if store, err := newS3SkillAssetStoreFromEnv(); err == nil && store != nil {
		return store
	}
	return &localSkillAssetStore{}
}

func newSkillAssetStoreFromFormalConfig() (skillAssetStore, error) {
	cfg, ok := loadFormalSkillAssetStoreConfig()
	if !ok {
		return nil, nil
	}
	return newSkillAssetStoreWithConfig(cfg)
}

func formalConfigFilesExist() bool {
	configPaths := []string{
		"/sysvol/config/agent-operator-integration.yaml",
		"/sysvol/secret/agent-operator-integration-secret.yaml",
	}
	for _, path := range configPaths {
		if _, err := os.Stat(path); err != nil {
			return false
		}
	}
	return true
}

func newSkillAssetStoreWithConfig(cfg config.S3Config) (skillAssetStore, error) {
	if cfg.Endpoint == "" || cfg.AccessID == "" || cfg.AccessSecretKey == "" || cfg.Bucket == "" {
		return nil, nil
	}
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}
	if cfg.StoragePrefix == "" {
		cfg.StoragePrefix = "aoi_skill_assets"
	}

	sess, err := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(cfg.AccessID, cfg.AccessSecretKey, ""),
		Endpoint:         aws.String(cfg.Endpoint),
		Region:           aws.String(cfg.Region),
		DisableSSL:       aws.Bool(!cfg.UseSSL),
		S3ForcePathStyle: aws.Bool(true),
		HTTPClient:       restinfra.NewRawHTTPClient(),
	})
	if err != nil {
		return nil, err
	}
	return &s3SkillAssetStore{
		client: s3.New(sess),
		bucket: cfg.Bucket,
		prefix: cfg.StoragePrefix,
	}, nil
}

func (s *localSkillAssetStore) Write(_ context.Context, skillID, relPath string, content []byte) (string, string, error) {
	storageKey := buildStorageKey(skillID, relPath)
	if err := os.MkdirAll(filepath.Dir(storageKey), 0o755); err != nil {
		return "", "", err
	}
	if err := os.WriteFile(storageKey, content, 0o600); err != nil {
		return "", "", err
	}
	return storageKey, checksumSHA256(content), nil
}

func (s *localSkillAssetStore) Read(_ context.Context, storageKey string) ([]byte, error) {
	return os.ReadFile(storageKey)
}

func (s *localSkillAssetStore) DeleteFile(_ context.Context, storageKey string) error {
	err := os.Remove(storageKey)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (s *localSkillAssetStore) DeleteSkill(_ context.Context, skillID string) error {
	root := buildStorageKey(skillID, "")
	root = filepath.Clean(root)
	return os.RemoveAll(root)
}

func newS3SkillAssetStoreFromEnv() (*s3SkillAssetStore, error) {
	if stringsEqualFold(os.Getenv("SKILL_ASSET_BACKEND"), "local") {
		return nil, nil
	}

	endpoint := os.Getenv("SKILL_S3_ENDPOINT")
	accessID := os.Getenv("SKILL_S3_ACCESS_ID")
	accessSecretKey := os.Getenv("SKILL_S3_ACCESS_SECRET_KEY")
	bucket := os.Getenv("SKILL_S3_BUCKET")
	if endpoint == "" || accessID == "" || accessSecretKey == "" || bucket == "" {
		return nil, nil
	}

	useSSL := false
	if raw := os.Getenv("SKILL_S3_USE_SSL"); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return nil, fmt.Errorf("parse SKILL_S3_USE_SSL failed: %w", err)
		}
		useSSL = parsed
	}
	region := os.Getenv("SKILL_S3_REGION")
	if region == "" {
		region = "us-east-1"
	}
	prefix := os.Getenv("SKILL_S3_PREFIX")
	if prefix == "" {
		prefix = "aoi_skill_assets"
	}
	store, err := newSkillAssetStoreWithConfig(config.S3Config{
		Endpoint:        endpoint,
		AccessID:        accessID,
		AccessSecretKey: accessSecretKey,
		Bucket:          bucket,
		Region:          region,
		UseSSL:          useSSL,
		StoragePrefix:   prefix,
	})
	if err != nil || store == nil {
		return nil, err
	}
	return store.(*s3SkillAssetStore), nil
}

func (s *s3SkillAssetStore) Write(ctx context.Context, skillID, relPath string, content []byte) (string, string, error) {
	key := s.buildObjectKey(skillID, relPath)
	_, err := s.client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   aws.ReadSeekCloser(strings.NewReader(string(content))),
	})
	if err != nil {
		return "", "", err
	}
	return key, checksumSHA256(content), nil
}

func (s *s3SkillAssetStore) Read(ctx context.Context, storageKey string) ([]byte, error) {
	resp, err := s.client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(storageKey),
	})
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	return io.ReadAll(resp.Body)
}

func (s *s3SkillAssetStore) DeleteFile(ctx context.Context, storageKey string) error {
	_, err := s.client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(storageKey),
	})
	return err
}

func (s *s3SkillAssetStore) DeleteSkill(ctx context.Context, skillID string) error {
	prefix := s.buildObjectKey(skillID, "")
	resp, err := s.client.ListObjectsV2WithContext(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return err
	}
	if len(resp.Contents) == 0 {
		return nil
	}
	objects := make([]*s3.ObjectIdentifier, 0, len(resp.Contents))
	for _, obj := range resp.Contents {
		objects = append(objects, &s3.ObjectIdentifier{Key: obj.Key})
	}
	_, err = s.client.DeleteObjectsWithContext(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(s.bucket),
		Delete: &s3.Delete{Objects: objects, Quiet: aws.Bool(true)},
	})
	return err
}

func (s *s3SkillAssetStore) buildObjectKey(skillID, relPath string) string {
	if relPath == "" {
		return filepath.ToSlash(filepath.Join(s.prefix, skillID))
	}
	return filepath.ToSlash(filepath.Join(s.prefix, skillID, relPath))
}

func checksumSHA256(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

func buildStorageKey(skillID, relPath string) string {
	return filepath.Join(os.TempDir(), "aoi_skill_assets", skillID, filepath.FromSlash(relPath))
}

func stringsEqualFold(a, b string) bool {
	return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
}
