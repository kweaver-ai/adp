// Package worker provides background workers for VEGA Manager.
package worker

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-viper/mapstructure/v2"
	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"vega-manager/common"
	"vega-manager/interfaces"
	"vega-manager/logics/connectors"
	"vega-manager/logics/connectors/factory"
)

var (
	dServiceOnce sync.Once
	dService     interfaces.DiscoveryService
)

// discoveryService provides resource discovery functionality.
type discoveryService struct {
	appSetting *common.AppSetting
	rs         interfaces.ResourceService
}

// NewDiscoveryService creates or returns the singleton DiscoveryService.
func NewDiscoveryService(appSetting *common.AppSetting,
	rs interfaces.ResourceService) interfaces.DiscoveryService {
	dServiceOnce.Do(func() {
		dService = &discoveryService{
			appSetting: appSetting,
			rs:         rs,
		}
	})
	return dService
}

// DiscoverCatalog discovers resources for a specific catalog.
func (ds *discoveryService) DiscoverCatalog(ctx context.Context,
	catalog *interfaces.Catalog) (*interfaces.DiscoveryResult, error) {

	logger.Infof("Starting discovery for catalog: %s", catalog.ID)

	// 验证 catalog 类型
	if catalog.Type != interfaces.CatalogTypePhysical {
		return nil, fmt.Errorf("discovery only supports physical catalogs")
	}

	// 1. 创建 Connector 并连接
	connector, err := ds.createAndConnectConnector(ctx, catalog)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to data source: %w", err)
	}
	defer connector.Close(ctx)

	// 2. 根据 connector category 分发到不同的发现函数
	category := connector.GetCategory()
	switch category {
	case interfaces.ConnectorCategoryTable:
		return ds.discoverTableResources(ctx, catalog, connector)
	case interfaces.ConnectorCategoryIndex:
		return ds.discoverIndexResources(ctx, catalog, connector)
	case interfaces.ConnectorCategoryFile, interfaces.ConnectorCategoryFileset:
		return ds.discoverFileResources(ctx, catalog, connector)
	default:
		return nil, fmt.Errorf("unsupported connector category for discovery: %s", category)
	}
}

// discoverFileResources discovers file resources from a file connector.
func (ds *discoveryService) discoverFileResources(ctx context.Context,
	catalog *interfaces.Catalog, connector connectors.Connector) (*interfaces.DiscoveryResult, error) {
	// TODO: 实现文件资源发现逻辑
	return nil, fmt.Errorf("file resource discovery not implemented yet")
}

// createAndConnectConnector creates and connects a connector for the catalog.
func (ds *discoveryService) createAndConnectConnector(ctx context.Context,
	catalog *interfaces.Catalog) (connectors.Connector, error) {

	// 使用 mapstructure 反序列化 ConnectorConfig
	cfg := &interfaces.ConnectorConfig{}
	if err := mapstructure.Decode(catalog.ConnectorConfig, cfg); err != nil {
		return nil, fmt.Errorf("failed to decode connector config: %w", err)
	}

	// 创建 connector
	connector, err := factory.GetFactory().CreateConnectorInstance(ctx, catalog.ConnectorType, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create connector: %w", err)
	}

	// 连接
	if err := connector.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return connector, nil
}
