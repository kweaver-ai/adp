// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package catalog

import (
	"fmt"
	"testing"

	"vega-backend/tests/at/fixtures"
	catalogfixtures "vega-backend/tests/at/fixtures/catalog"
	"vega-backend/tests/at/setup"
	"vega-backend/tests/testutil"
)

// TestSuite Catalog测试套件
// 封装通用测试所需的配置、客户端和构建器
type TestSuite struct {
	Config  *setup.TestConfig
	Client  *testutil.HTTPClient
	Builder catalogfixtures.PayloadBuilder
	T       *testing.T
}

// NewTestSuite 创建新的Catalog测试套件
func NewTestSuite(t *testing.T, connectorType string) (*TestSuite, error) {
	// 加载测试配置
	config, err := setup.LoadTestConfig()
	if err != nil {
		return nil, err
	}

	// 创建HTTP客户端
	client := testutil.NewHTTPClient(config.VegaManager.BaseURL)

	// 创建对应类型的PayloadBuilder
	builder := catalogfixtures.NewPayloadBuilder(connectorType, config)
	if builder == nil {
		return nil, fmt.Errorf("不支持的connector类型: %s", connectorType)
	}

	return &TestSuite{
		Config:  config,
		Client:  client,
		Builder: builder,
		T:       t,
	}, nil
}

// Setup 初始化测试环境
func (s *TestSuite) Setup() error {
	// 验证服务可用性
	if err := s.Client.CheckHealth(); err != nil {
		return err
	}

	s.T.Logf("✓ AT测试环境就绪，VEGA Manager: %s", s.Config.VegaManager.BaseURL)
	s.T.Logf("✓ 测试Connector类型: %s", s.Builder.GetConnectorType())

	// 清理现有catalog
	fixtures.CleanupCatalogs(s.Client, s.T)

	return nil
}

// Cleanup 清理测试环境
func (s *TestSuite) Cleanup() {
	fixtures.CleanupCatalogs(s.Client, s.T)
}

// GetConnectorType 获取当前测试的connector类型
func (s *TestSuite) GetConnectorType() string {
	return s.Builder.GetConnectorType()
}

// ========== CRUD操作封装 ==========

// CreateCatalog 创建catalog并返回ID和完整响应
func (s *TestSuite) CreateCatalog(payload map[string]any) (string, *testutil.HTTPResponse) {
	resp := s.Client.POST("/api/vega-backend/v1/catalogs", payload)
	if resp.StatusCode == 201 {
		if id, ok := resp.Body["id"].(string); ok {
			return id, &resp
		}
	}
	return "", &resp
}

// GetCatalog 获取catalog详情
func (s *TestSuite) GetCatalog(catalogID string) *testutil.HTTPResponse {
	resp := s.Client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
	return &resp
}

// GetCatalogData 获取catalog数据（从entries中提取）
func (s *TestSuite) GetCatalogData(catalogID string) map[string]any {
	resp := s.GetCatalog(catalogID)
	return fixtures.ExtractFromEntriesResponse(*resp)
}

// UpdateCatalog 更新catalog
func (s *TestSuite) UpdateCatalog(catalogID string, payload map[string]any) *testutil.HTTPResponse {
	resp := s.Client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, payload)
	return &resp
}

// DeleteCatalog 删除catalog
func (s *TestSuite) DeleteCatalog(catalogID string) *testutil.HTTPResponse {
	resp := s.Client.DELETE("/api/vega-backend/v1/catalogs/" + catalogID)
	return &resp
}

// ListCatalogs 列表查询catalogs
func (s *TestSuite) ListCatalogs(offset, limit int) *testutil.HTTPResponse {
	url := fmt.Sprintf("/api/vega-backend/v1/catalogs?offset=%d&limit=%d", offset, limit)
	resp := s.Client.GET(url)
	return &resp
}

// ListCatalogsWithParams 带参数列表查询catalogs
func (s *TestSuite) ListCatalogsWithParams(params string) *testutil.HTTPResponse {
	resp := s.Client.GET("/api/vega-backend/v1/catalogs?" + params)
	return &resp
}

// TestConnection 测试连接
func (s *TestSuite) TestConnection(catalogID string) *testutil.HTTPResponse {
	resp := s.Client.POST("/api/vega-backend/v1/catalogs/"+catalogID+"/test-connection", nil)
	return &resp
}

// GetHealthStatus 获取健康状态
func (s *TestSuite) GetHealthStatus(catalogID string) *testutil.HTTPResponse {
	resp := s.Client.GET("/api/vega-backend/v1/catalogs/" + catalogID + "/health-status")
	return &resp
}

// ========== 便捷方法 ==========

// CreateAndGetCatalog 创建catalog并获取完整数据
func (s *TestSuite) CreateAndGetCatalog(payload map[string]any) (string, map[string]any, error) {
	catalogID, createResp := s.CreateCatalog(payload)
	if createResp.StatusCode != 201 {
		return "", nil, fmt.Errorf("创建catalog失败，状态码: %d", createResp.StatusCode)
	}

	catalogData := s.GetCatalogData(catalogID)
	if catalogData == nil {
		return "", nil, fmt.Errorf("获取catalog数据失败")
	}

	return catalogID, catalogData, nil
}

// BuildUpdatePayload 基于现有catalog数据构建更新payload
// 自动回填加密后的password（因为GET不再返回敏感字段）
func (s *TestSuite) BuildUpdatePayload(catalogID string, updates map[string]any) (map[string]any, error) {
	// 获取当前catalog数据
	catalogData := s.GetCatalogData(catalogID)
	if catalogData == nil {
		return nil, fmt.Errorf("获取catalog数据失败")
	}

	// 构建更新payload
	payload := catalogfixtures.BuildUpdatePayload(catalogData, updates)

	// 回填加密后的password（GET响应不再包含敏感字段）
	if s.Builder.GetConnectorType() != "" {
		catalogfixtures.InjectEncryptedPassword(payload, s.Builder.GetEncryptedPassword())
	}

	return payload, nil
}
