package resource

import (
	"fmt"
	"strings"
	"testing"

	"vega-backend/tests/at/fixtures"
	catalogfixtures "vega-backend/tests/at/fixtures/catalog"
	resourcefixtures "vega-backend/tests/at/fixtures/resource"
	"vega-backend/tests/at/setup"
	"vega-backend/tests/testutil"
)

// TestSuite Resource测试套件
// 封装通用测试所需的配置、客户端、前置Catalog和构建器
type TestSuite struct {
	Config         *setup.TestConfig
	Client         *testutil.HTTPClient
	CatalogBuilder catalogfixtures.PayloadBuilder // 用于创建前置Catalog
	CatalogID      string                         // 前置Catalog ID
	ConnectorType  string
	T              *testing.T
}

// NewTestSuite 创建新的Resource测试套件
func NewTestSuite(t *testing.T, connectorType string) (*TestSuite, error) {
	// 加载测试配置
	config, err := setup.LoadTestConfig()
	if err != nil {
		return nil, err
	}

	// 创建HTTP客户端
	client := testutil.NewHTTPClient(config.VegaManager.BaseURL)

	// 创建对应类型的CatalogBuilder
	catalogBuilder := catalogfixtures.NewPayloadBuilder(connectorType, config)
	if catalogBuilder == nil {
		return nil, fmt.Errorf("不支持的connector类型: %s", connectorType)
	}

	return &TestSuite{
		Config:         config,
		Client:         client,
		CatalogBuilder: catalogBuilder,
		ConnectorType:  connectorType,
		T:              t,
	}, nil
}

// Setup 初始化测试环境
// 清理顺序：先清理Resources，再清理Catalogs（避免依赖问题）
func (s *TestSuite) Setup() error {
	// 验证服务可用性
	if err := s.Client.CheckHealth(); err != nil {
		return err
	}

	s.T.Logf("✓ AT测试环境就绪，VEGA Manager: %s", s.Config.VegaManager.BaseURL)
	s.T.Logf("✓ 测试Connector类型: %s", s.ConnectorType)

	// 清理现有资源（先Resource后Catalog）
	fixtures.CleanupResources(s.Client, s.T)
	fixtures.CleanupCatalogs(s.Client, s.T)

	// 创建前置Catalog
	catalogID, err := s.CreatePrerequisiteCatalog()
	if err != nil {
		return err
	}
	s.CatalogID = catalogID

	s.T.Logf("✓ 前置Catalog创建成功，ID: %s", s.CatalogID)
	return nil
}

// Cleanup 清理测试环境
func (s *TestSuite) Cleanup() {
	fixtures.CleanupResources(s.Client, s.T)
	fixtures.CleanupCatalogs(s.Client, s.T)
}

// CreatePrerequisiteCatalog 创建前置Catalog
// 公开方法，支持跨Catalog唯一性测试等场景
func (s *TestSuite) CreatePrerequisiteCatalog() (string, error) {
	payload := s.CatalogBuilder.BuildCreatePayload()
	resp := s.Client.POST("/api/vega-backend/v1/catalogs", payload)
	if resp.StatusCode != 201 {
		return "", fmt.Errorf("创建前置Catalog失败，状态码: %d", resp.StatusCode)
	}

	catalogID, ok := resp.Body["id"].(string)
	if !ok || catalogID == "" {
		return "", fmt.Errorf("创建前置Catalog失败，无法获取ID")
	}

	return catalogID, nil
}

// ========== CRUD操作封装 ==========

// CreateResource 创建resource并返回ID和完整响应
func (s *TestSuite) CreateResource(payload map[string]any) (string, *testutil.HTTPResponse) {
	resp := s.Client.POST("/api/vega-backend/v1/resources", payload)
	if resp.StatusCode == 201 {
		if id, ok := resp.Body["id"].(string); ok {
			return id, &resp
		}
	}
	return "", &resp
}

// GetResource 获取resource详情
func (s *TestSuite) GetResource(resourceID string) *testutil.HTTPResponse {
	resp := s.Client.GET("/api/vega-backend/v1/resources/" + resourceID)
	return &resp
}

// GetResources 批量获取resources（逗号分隔IDs）
func (s *TestSuite) GetResources(ids []string) *testutil.HTTPResponse {
	idsStr := strings.Join(ids, ",")
	resp := s.Client.GET("/api/vega-backend/v1/resources/" + idsStr)
	return &resp
}

// GetResourceData 获取resource数据（从entries中提取）
func (s *TestSuite) GetResourceData(resourceID string) map[string]any {
	resp := s.GetResource(resourceID)
	return fixtures.ExtractFromEntriesResponse(*resp)
}

// UpdateResource 更新resource
func (s *TestSuite) UpdateResource(resourceID string, payload map[string]any) *testutil.HTTPResponse {
	resp := s.Client.PUT("/api/vega-backend/v1/resources/"+resourceID, payload)
	return &resp
}

// DeleteResource 删除resource
func (s *TestSuite) DeleteResource(resourceID string) *testutil.HTTPResponse {
	resp := s.Client.DELETE("/api/vega-backend/v1/resources/" + resourceID)
	return &resp
}

// DeleteResources 批量删除resources（逗号分隔IDs）
func (s *TestSuite) DeleteResources(ids []string) *testutil.HTTPResponse {
	idsStr := strings.Join(ids, ",")
	resp := s.Client.DELETE("/api/vega-backend/v1/resources/" + idsStr)
	return &resp
}

// ListResources 列表查询resources
func (s *TestSuite) ListResources(offset, limit int) *testutil.HTTPResponse {
	url := fmt.Sprintf("/api/vega-backend/v1/resources?offset=%d&limit=%d", offset, limit)
	resp := s.Client.GET(url)
	return &resp
}

// ListResourcesWithParams 带参数列表查询resources
func (s *TestSuite) ListResourcesWithParams(params string) *testutil.HTTPResponse {
	resp := s.Client.GET("/api/vega-backend/v1/resources?" + params)
	return &resp
}

// ========== 便捷方法 ==========

// CreateAndGetResource 创建resource并获取完整数据
func (s *TestSuite) CreateAndGetResource(payload map[string]any) (string, map[string]any, error) {
	resourceID, createResp := s.CreateResource(payload)
	if createResp.StatusCode != 201 {
		return "", nil, fmt.Errorf("创建resource失败，状态码: %d", createResp.StatusCode)
	}

	resourceData := s.GetResourceData(resourceID)
	if resourceData == nil {
		return "", nil, fmt.Errorf("获取resource数据失败")
	}

	return resourceID, resourceData, nil
}

// BuildUpdatePayload 基于现有resource数据构建更新payload
func (s *TestSuite) BuildUpdatePayload(resourceID string, updates map[string]any) (map[string]any, error) {
	// 获取当前resource数据
	resourceData := s.GetResourceData(resourceID)
	if resourceData == nil {
		return nil, fmt.Errorf("获取resource数据失败")
	}

	// 构建更新payload
	return resourcefixtures.BuildUpdatePayload(resourceData, updates), nil
}
