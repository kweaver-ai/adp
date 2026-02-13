// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package catalog provides Catalog management business logic.
package catalog

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	kwcrypto "github.com/kweaver-ai/kweaver-go-lib/crypto"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/rs/xid"
	"go.opentelemetry.io/otel/codes"

	"vega-backend/common"
	catalogAccess "vega-backend/drivenadapters/catalog"
	oerrors "vega-backend/errors"
	"vega-backend/interfaces"
	"vega-backend/logics/connectors/factory"
)

const (
	// EncryptedPrefix is the prefix for encrypted values.
	EncryptedPrefix = "ENC:"
)

var (
	cServiceOnce sync.Once
	cService     interfaces.CatalogService
)

type catalogService struct {
	appSetting *common.AppSetting
	ca         interfaces.CatalogAccess
	cipher     kwcrypto.Cipher
}

// NewCatalogService creates a new CatalogService.
func NewCatalogService(appSetting *common.AppSetting) interfaces.CatalogService {
	cServiceOnce.Do(func() {
		var cipher kwcrypto.Cipher
		if appSetting.CryptoSetting.Enabled {
			var err error
			cipher, err = kwcrypto.NewRSACipher(appSetting.CryptoSetting.PrivateKey, appSetting.CryptoSetting.PublicKey)
			if err != nil {
				logger.Fatalf("Failed to create RSA cipher: %v", err)
			}
		}
		cService = &catalogService{
			appSetting: appSetting,
			ca:         catalogAccess.NewCatalogAccess(appSetting),
			cipher:     cipher,
		}
	})
	return cService
}

// Create creates a new Catalog.
func (cs *catalogService) Create(ctx context.Context, req *interfaces.CatalogRequest) (string, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Create catalog")
	defer span.End()

	// Get account info from context
	accountInfo := interfaces.AccountInfo{}
	if v := ctx.Value(interfaces.ACCOUNT_INFO_KEY); v != nil {
		accountInfo = v.(interfaces.AccountInfo)
	}

	catalogType := interfaces.CatalogTypePhysical
	if req.ConnectorType == "" {
		catalogType = interfaces.CatalogTypeLogical
	} else {
		// 验证敏感字段是否为合法 RSA 密文，获取明文用于连接测试
		sensitiveFields := factory.GetFactory().GetSensitiveFields(req.ConnectorType)
		decryptedConfig, err := cs.validateAndDecryptSensitiveFields(sensitiveFields, req.ConnectorConfig)
		if err != nil {
			logger.Errorf("Failed to validate sensitive fields: %v", err)
			o11y.Error(ctx, fmt.Sprintf("Failed to validate sensitive fields: %v", err))
			span.SetStatus(codes.Error, "Validate sensitive fields failed")
			return "", rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.VegaManager_Catalog_InvalidParameter_SensitiveFieldNotEncrypted).
				WithErrorDetails(err.Error())
		}

		// 用解密后的明文 config 创建 connector 并测试连接
		connectorCfg := interfaces.ConnectorConfig(decryptedConfig)
		connector, err := factory.GetFactory().CreateConnectorInstance(ctx, req.ConnectorType, connectorCfg)
		if err != nil {
			logger.Errorf("Failed to create connector: %v", err)
			o11y.Error(ctx, fmt.Sprintf("Failed to create connector: %v", err))
			span.SetStatus(codes.Error, "Create connector failed")
			return "", rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.VegaManager_Catalog_InternalError_CreateFailed).
				WithErrorDetails(err.Error())
		}

		if err := connector.TestConnection(ctx); err != nil {
			logger.Errorf("Failed to test connection to data source: %v", err)
			o11y.Error(ctx, fmt.Sprintf("Failed to test connection to data source: %v", err))
			span.SetStatus(codes.Error, "Connection failed")
			connector.Close(ctx)
			return "", rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.VegaManager_Catalog_InternalError_TestConnectionFailed).
				WithErrorDetails(err.Error())
		}
		defer connector.Close(ctx)
	}

	now := time.Now().UnixMilli()
	catalog := &interfaces.Catalog{
		ID:                 xid.New().String(),
		Name:               req.Name,
		Tags:               req.Tags,
		Description:        req.Description,
		Type:               catalogType,
		ConnectorType:      req.ConnectorType,
		ConnectorConfig:    req.ConnectorConfig,
		HealthCheckEnabled: true,
		CatalogHealthCheckStatus: interfaces.CatalogHealthCheckStatus{
			HealthCheckStatus: interfaces.CatalogHealthStatusHealthy,
			LastCheckTime:     now,
		},
		Creator:    accountInfo,
		CreateTime: now,
		Updater:    accountInfo,
		UpdateTime: now,
	}

	if err := cs.ca.Create(ctx, catalog); err != nil {
		logger.Errorf("Create catalog failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Create catalog failed: %v", err))
		span.SetStatus(codes.Error, "Create catalog failed")
		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_Catalog_InternalError_CreateFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return catalog.ID, nil
}

// Get retrieves a Catalog by ID.
func (cs *catalogService) GetByID(ctx context.Context, id string, withSensitiveFields bool) (*interfaces.Catalog, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Get catalog")
	defer span.End()

	catalog, err := cs.ca.GetByID(ctx, id)
	if err != nil {
		span.SetStatus(codes.Error, "Get catalog failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_Catalog_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}
	if catalog == nil {
		span.SetStatus(codes.Error, "Catalog not found")
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, oerrors.VegaManager_Catalog_NotFound)
	}

	if !withSensitiveFields {
		// 移除敏感字段，不返回给前端
		cs.removeSensitiveFields(catalog)
	} else {
		// 验证敏感字段是否为合法 RSA 密文，获取明文用于连接测试
		sensitiveFields := factory.GetFactory().GetSensitiveFields(catalog.ConnectorType)
		decryptedConfig, err := cs.decryptSensitiveFields(sensitiveFields, catalog.ConnectorConfig)
		if err != nil {
			logger.Errorf("Failed to validate sensitive fields: %v", err)
			o11y.Error(ctx, fmt.Sprintf("Failed to validate sensitive fields: %v", err))
			span.SetStatus(codes.Error, "Validate sensitive fields failed")
			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.VegaManager_Catalog_InvalidParameter_SensitiveFieldNotEncrypted).
				WithErrorDetails(err.Error())
		}
		catalog.ConnectorConfig = decryptedConfig
	}

	span.SetStatus(codes.Ok, "")
	return catalog, nil
}

// GetByIDs retrieves a Catalog by IDs.
func (cs *catalogService) GetByIDs(ctx context.Context, ids []string) ([]*interfaces.Catalog, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Get catalogs")
	defer span.End()

	catalogs, err := cs.ca.GetByIDs(ctx, ids)
	if err != nil {
		span.SetStatus(codes.Error, "Get catalog failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_Catalog_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}

	// 移除敏感字段，不返回给前端
	for _, c := range catalogs {
		cs.removeSensitiveFields(c)
	}

	span.SetStatus(codes.Ok, "")
	return catalogs, nil
}

// List lists Catalogs with filters.
func (cs *catalogService) List(ctx context.Context, params interfaces.CatalogsQueryParams) ([]*interfaces.Catalog, int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "List catalogs")
	defer span.End()

	catalogs, total, err := cs.ca.List(ctx, params)
	if err != nil {
		span.SetStatus(codes.Error, "List catalogs failed")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_Catalog_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}

	// 移除敏感字段，不返回给前端
	for _, c := range catalogs {
		cs.removeSensitiveFields(c)
	}

	span.SetStatus(codes.Ok, "")
	return catalogs, total, nil
}

// Update updates a Catalog.
func (cs *catalogService) Update(ctx context.Context, id string, req *interfaces.CatalogRequest) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Update catalog")
	defer span.End()

	catalog := req.OriginCatalog
	if catalog == nil {
		span.SetStatus(codes.Error, "Catalog not found")
		return rest.NewHTTPError(ctx, http.StatusNotFound, oerrors.VegaManager_Catalog_NotFound)
	}

	// Apply updates
	if req.Name != catalog.Name {
		exists, err := cs.CheckExistByName(ctx, req.Name)
		if err != nil {
			return err
		}
		if exists {
			span.SetStatus(codes.Error, "Catalog name exists")
			return rest.NewHTTPError(ctx, http.StatusConflict, oerrors.VegaManager_Catalog_NameExists)
		}
		catalog.Name = req.Name
	}
	catalog.Tags = req.Tags
	catalog.Description = req.Description

	if catalog.ConnectorType != req.ConnectorType {
		span.SetStatus(codes.Error, "can not change connector type")
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.VegaManager_Catalog_InvalidParameter_ConnectorType)
	} else if req.ConnectorType != "" {
		// 验证敏感字段是否为合法 RSA 密文，获取明文用于连接测试
		sensitiveFields := factory.GetFactory().GetSensitiveFields(req.ConnectorType)
		decryptedConfig, err := cs.validateAndDecryptSensitiveFields(sensitiveFields, req.ConnectorConfig)
		if err != nil {
			logger.Errorf("Failed to validate sensitive fields: %v", err)
			o11y.Error(ctx, fmt.Sprintf("Failed to validate sensitive fields: %v", err))
			span.SetStatus(codes.Error, "Validate sensitive fields failed")
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.VegaManager_Catalog_InvalidParameter_SensitiveFieldNotEncrypted).
				WithErrorDetails(err.Error())
		}

		// 用解密后的明文 config 创建 connector 并测试连接
		connectorCfg := interfaces.ConnectorConfig(decryptedConfig)
		connector, err := factory.GetFactory().CreateConnectorInstance(ctx, req.ConnectorType, connectorCfg)
		if err != nil {
			logger.Errorf("Failed to create connector: %v", err)
			o11y.Error(ctx, fmt.Sprintf("Failed to create connector: %v", err))
			span.SetStatus(codes.Error, "Create connector failed")
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.VegaManager_Catalog_InternalError_CreateFailed).
				WithErrorDetails(err.Error())
		}

		if err := connector.TestConnection(ctx); err != nil {
			logger.Errorf("Failed to test connection to data source: %v", err)
			o11y.Error(ctx, fmt.Sprintf("Failed to test connection to data source: %v", err))
			span.SetStatus(codes.Error, "Connection failed")
			connector.Close(ctx)
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.VegaManager_Catalog_InternalError_TestConnectionFailed).
				WithErrorDetails(err.Error())
		}
		defer connector.Close(ctx)

		// req.ConnectorConfig 已在 validateAndDecryptSensitiveFields 中加上 ENC: 前缀
		catalog.ConnectorConfig = req.ConnectorConfig
	}

	// Get account info
	accountInfo := interfaces.AccountInfo{}
	if v := ctx.Value(interfaces.ACCOUNT_INFO_KEY); v != nil {
		accountInfo = v.(interfaces.AccountInfo)
	}

	now := time.Now().UnixMilli()
	catalog.Updater = accountInfo
	catalog.UpdateTime = now
	catalog.CatalogHealthCheckStatus = interfaces.CatalogHealthCheckStatus{
		HealthCheckStatus: interfaces.CatalogHealthStatusHealthy,
		LastCheckTime:     now,
	}

	if err := cs.ca.Update(ctx, catalog); err != nil {
		span.SetStatus(codes.Error, "Update catalog failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_Catalog_InternalError_UpdateFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// DeleteByIDs deletes Catalogs by IDs.
func (cs *catalogService) DeleteByIDs(ctx context.Context, ids []string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete catalogs")
	defer span.End()

	if len(ids) == 0 {
		span.SetStatus(codes.Ok, "")
		return nil
	}

	if err := cs.ca.DeleteByIDs(ctx, ids); err != nil {
		span.SetStatus(codes.Error, "Delete catalogs failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_Catalog_InternalError_DeleteFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// CheckExistByID checks if a Catalog exists by ID.
func (cs *catalogService) CheckExistByID(ctx context.Context, id string) (bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Check catalog exist by ID")
	defer span.End()

	catalog, err := cs.ca.GetByID(ctx, id)
	if err != nil {
		span.SetStatus(codes.Error, "GetByID failed")
		return false, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_Catalog_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return catalog != nil, nil
}

// CheckExistByName checks if a Catalog exists by name.
func (cs *catalogService) CheckExistByName(ctx context.Context, name string) (bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Check catalog exist by name")
	defer span.End()

	catalog, err := cs.ca.GetByName(ctx, name)
	if err != nil {
		span.SetStatus(codes.Error, "GetByName failed")
		return false, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_Catalog_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return catalog != nil, nil
}

// TestConnection tests catalog connection.
func (cs *catalogService) TestConnection(ctx context.Context, catalog *interfaces.Catalog) (*interfaces.CatalogHealthCheckStatus, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Test catalog connection")
	defer span.End()

	if catalog == nil {
		span.SetStatus(codes.Error, "Catalog not found")
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, oerrors.VegaManager_Catalog_NotFound)
	}

	result := catalog.CatalogHealthCheckStatus
	span.SetStatus(codes.Ok, "")
	return &result, nil
}

// validateAndDecryptSensitiveFields 验证敏感字段是否为合法 RSA 密文，
// 返回解密后的明文 config（用于连接测试），同时在原始 config 中加上 ENC: 前缀（用于存储）。
// 如果 cipher 为 nil（加密未启用），直接返回原始 config 的拷贝作为 decryptedConfig，不做验证。
func (cs *catalogService) validateAndDecryptSensitiveFields(
	sensitiveFields []string, config map[string]any,
) (decryptedConfig map[string]any, err error) {
	// 拷贝 config 作为 decryptedConfig
	decryptedConfig = make(map[string]any, len(config))
	for k, v := range config {
		decryptedConfig[k] = v
	}

	if cs.cipher == nil {
		return decryptedConfig, nil
	}

	for _, field := range sensitiveFields {
		val, ok := config[field].(string)
		if !ok || val == "" {
			continue
		}
		// 尝试用私钥解密，验证是否为合法密文
		decrypted, decryptErr := cs.cipher.Decrypt(val)
		if decryptErr != nil {
			return nil, fmt.Errorf("field %s: %w", field, decryptErr)
		}
		// 解密成功：明文放入 decryptedConfig，原始 config 加上 ENC: 前缀
		decryptedConfig[field] = decrypted
		config[field] = EncryptedPrefix + val
	}
	return decryptedConfig, nil
}

// removeSensitiveFields 从 ConnectorConfig 中移除敏感字段，用于 GET/List 返回
func (cs *catalogService) removeSensitiveFields(catalog *interfaces.Catalog) {
	if catalog == nil || catalog.ConnectorType == "" {
		return
	}
	sensitiveFields := factory.GetFactory().GetSensitiveFields(catalog.ConnectorType)
	for _, field := range sensitiveFields {
		delete(catalog.ConnectorConfig, field)
	}
}

// decryptSensitiveFields 验证敏感字段是否为合法 RSA 密文，
// 返回解密后的明文 config（用于连接），数据从数据库获取而来，需要先去除ENC前缀，再解密
// 如果 cipher 为 nil（加密未启用），直接返回原始 config 的拷贝作为 decryptedConfig，不做验证。
func (cs *catalogService) decryptSensitiveFields(sensitiveFields []string,
	config map[string]any) (decryptedConfig map[string]any, err error) {

	// 拷贝 config 作为 decryptedConfig
	decryptedConfig = make(map[string]any, len(config))
	for k, v := range config {
		decryptedConfig[k] = v
	}

	if cs.cipher == nil {
		return decryptedConfig, nil
	}

	for _, field := range sensitiveFields {
		val, ok := config[field].(string)
		if !ok || val == "" {
			continue
		}
		// 尝试用私钥解密，验证是否为合法密文
		if !strings.HasPrefix(val, EncryptedPrefix) {
			return nil, fmt.Errorf("field %s: %w", field, errors.New("not encrypted"))
		} else {
			val = val[len(EncryptedPrefix):]
		}
		decrypted, decryptErr := cs.cipher.Decrypt(val)
		if decryptErr != nil {
			return nil, fmt.Errorf("field %s: %w", field, decryptErr)
		}
		// 解密成功：明文放入 decryptedConfig，原始 config 加上 ENC: 前缀
		decryptedConfig[field] = decrypted
		config[field] = EncryptedPrefix + val
	}
	return decryptedConfig, nil
}

func (cs *catalogService) UpdateMetadata(ctx context.Context, id string, metadata map[string]any) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "UpdateMetadata")
	defer span.End()

	err := cs.ca.UpdateMetadata(ctx, id, metadata)
	if err != nil {
		logger.Errorf("Update metadata failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Update metadata failed: %v", err))
		span.SetStatus(codes.Error, "Update metadata failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_Catalog_InternalError_UpdateFailed).
			WithErrorDetails(err.Error())
	}

	return nil
}
