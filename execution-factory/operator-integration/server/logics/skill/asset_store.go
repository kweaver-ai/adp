package skill

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	restinfra "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/rest"
)

//go:generate mockgen -source=asset_store.go -destination=../../mocks/skill_asset_store.go -package=mocks

type skillAssetStore interface {
	Write(ctx context.Context, skillID, relPath string, content []byte) (storageKey string, checksum string, err error)
	Read(ctx context.Context, storageKey string) ([]byte, error)
	DeleteFile(ctx context.Context, storageKey string) error
	DeleteSkill(ctx context.Context, skillID string) error
}

type ossGatewaySkillAssetStore struct {
	client          *http.Client
	baseURL         string
	storageID       string
	internalRequest bool
	expires         int64
}

type localSkillAssetStore struct{}

const skillAssetObjectPrefix = "aoi_skill_assets"

var loadFormalSkillAssetStoreConfig = func() (config.OSSGatewayConfig, bool) {
	if !formalConfigFilesExist() {
		return config.OSSGatewayConfig{}, false
	}
	return config.NewConfigLoader().OSSGatewayConfig, true
}

func newSkillAssetStore() skillAssetStore {
	if store, err := newSkillAssetStoreFromFormalConfig(); err == nil && store != nil {
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

func newSkillAssetStoreWithConfig(cfg config.OSSGatewayConfig) (skillAssetStore, error) {
	if cfg.BaseURL == "" || cfg.StorageID == "" {
		return nil, nil
	}
	if cfg.Expires <= 0 {
		cfg.Expires = 3600
	}
	return &ossGatewaySkillAssetStore{
		client:          restinfra.NewRawHTTPClient(),
		baseURL:         strings.TrimRight(cfg.BaseURL, "/"),
		storageID:       cfg.StorageID,
		internalRequest: cfg.InternalRequest,
		expires:         cfg.Expires,
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

func (s *ossGatewaySkillAssetStore) Write(ctx context.Context, skillID, relPath string, content []byte) (string, string, error) {
	key := s.buildObjectKey(skillID, relPath)
	authReq, err := s.requestPresigned(ctx, http.MethodGet, s.objectActionURL("upload", key, url.Values{
		"request_method":   []string{http.MethodPut},
		"expires":          []string{fmt.Sprintf("%d", s.expires)},
		"internal_request": []string{fmt.Sprintf("%t", s.internalRequest)},
	}))
	if err != nil {
		return "", "", err
	}
	if err = s.doSignedRequest(ctx, authReq.Method, authReq.URL, authReq.Headers, bytes.NewReader(content)); err != nil {
		return "", "", err
	}
	return key, checksumSHA256(content), nil
}

func (s *ossGatewaySkillAssetStore) Read(ctx context.Context, storageKey string) ([]byte, error) {
	authReq, err := s.requestPresigned(ctx, http.MethodGet, s.objectActionURL("download", storageKey, url.Values{
		"expires":          []string{fmt.Sprintf("%d", s.expires)},
		"internal_request": []string{fmt.Sprintf("%t", s.internalRequest)},
	}))
	if err != nil {
		return nil, err
	}
	resp, err := s.doSignedRequestRaw(ctx, authReq.Method, authReq.URL, authReq.Headers, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	return io.ReadAll(resp.Body)
}

func (s *ossGatewaySkillAssetStore) DeleteFile(ctx context.Context, storageKey string) error {
	authReq, err := s.requestPresigned(ctx, http.MethodGet, s.objectActionURL("delete", storageKey, url.Values{
		"expires":          []string{fmt.Sprintf("%d", s.expires)},
		"internal_request": []string{fmt.Sprintf("%t", s.internalRequest)},
	}))
	if err != nil {
		return err
	}
	return s.doSignedRequest(ctx, authReq.Method, authReq.URL, authReq.Headers, nil)
}

func (s *ossGatewaySkillAssetStore) DeleteSkill(_ context.Context, _ string) error {
	// Gateway API only provides single-object delete. DeleteSkill is kept as a
	// no-op because registry already deletes indexed files one by one before
	// calling this method. The registry cleanup path will be simplified later.
	return nil
}

func (s *ossGatewaySkillAssetStore) buildObjectKey(skillID, relPath string) string {
	if relPath == "" {
		return filepath.ToSlash(filepath.Join(skillAssetObjectPrefix, skillID))
	}
	return filepath.ToSlash(filepath.Join(skillAssetObjectPrefix, skillID, relPath))
}

type gatewayAuthResponse struct {
	Code    int                `json:"code"`
	Message string             `json:"message"`
	Data    gatewayAuthRequest `json:"data"`
}

type gatewayAuthRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}

func (s *ossGatewaySkillAssetStore) objectActionURL(action, key string, query url.Values) string {
	if query == nil {
		query = url.Values{}
	}
	encodedKey := url.PathEscape(key)
	return fmt.Sprintf("%s/api/v1/%s/%s/%s?%s", s.baseURL, action, s.storageID, encodedKey, query.Encode())
}

func (s *ossGatewaySkillAssetStore) requestPresigned(ctx context.Context, method, reqURL string) (*gatewayAuthRequest, error) {
	req, err := http.NewRequestWithContext(ctx, method, reqURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, errors.DefaultHTTPError(ctx, http.StatusInternalServerError, fmt.Sprintf("request oss gateway failed: %s", strings.TrimSpace(string(body))))
	}
	var payload gatewayAuthResponse
	if err = json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if payload.Code != 0 || payload.Data.URL == "" || payload.Data.Method == "" {
		return nil, errors.DefaultHTTPError(ctx, http.StatusInternalServerError, fmt.Sprintf("invalid oss gateway response: code=%d message=%s", payload.Code, payload.Message))
	}
	return &payload.Data, nil
}

func (s *ossGatewaySkillAssetStore) doSignedRequest(ctx context.Context, method, reqURL string, headers map[string]string, body io.Reader) error {
	resp, err := s.doSignedRequestRaw(ctx, method, reqURL, headers, body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return nil
}

func (s *ossGatewaySkillAssetStore) doSignedRequestRaw(ctx context.Context, method, reqURL string, headers map[string]string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		defer func() { _ = resp.Body.Close() }()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, errors.DefaultHTTPError(ctx, http.StatusInternalServerError, fmt.Sprintf("request object storage failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(bodyBytes))))
	}
	return resp, nil
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
