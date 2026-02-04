// Package worker provides background workers for VEGA Manager.
package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/segmentio/kafka-go"

	"vega-backend/common"
	kafka_access "vega-backend/drivenadapters/kafka"
	"vega-backend/interfaces"
	logicsCatalog "vega-backend/logics/catalog"
	"vega-backend/logics/connectors"
	"vega-backend/logics/connectors/factory"
	"vega-backend/logics/discovery_task"
	"vega-backend/logics/resource"
)

var (
	dWorkerOnce sync.Once
	dWorker     interfaces.DiscoveryWorker
)

// discoveryWorker provides resource discovery functionality.
type discoveryWorker struct {
	appSetting *common.AppSetting
	ka         interfaces.KafkaAccess
	rs         interfaces.ResourceService
	cs         interfaces.CatalogService
	dts        interfaces.DiscoveryTaskService
	reader     *kafka.Reader
}

// NewDiscoveryWorker creates or returns the singleton DiscoveryWorker.
func NewDiscoveryWorker(appSetting *common.AppSetting) interfaces.DiscoveryWorker {
	dWorkerOnce.Do(func() {
		dWorker = &discoveryWorker{
			appSetting: appSetting,
			ka:         kafka_access.NewKafkaAccess(appSetting),
			rs:         resource.NewResourceService(appSetting),
			cs:         logicsCatalog.NewCatalogService(appSetting),
			dts:        discovery_task.NewDiscoveryTaskService(appSetting),
		}
	})
	return dWorker
}

func (dw *discoveryWorker) Start() {
	ctx := context.Background()
	err := dw.ka.CreateTopic(ctx, interfaces.DiscoveryTaskTopic)
	if err != nil {
		logger.Errorf("Failed to create topic: %v", err)
		return
	}

	go func() {
		dw.Run(ctx)
	}()
}

func (dw *discoveryWorker) Run(ctx context.Context) {
	defer func() {
		if panicErr := recover(); panicErr != nil {
			logger.Errorf("Discovery worker panic: %v", panicErr)
		}
	}()

	reader, err := dw.ka.NewReader(ctx, interfaces.DiscoveryTaskTopic, "discovery_worker")
	if err != nil {
		logger.Errorf("Failed to create reader: %v", err)
		return
	}
	dw.reader = reader
	defer dw.ka.CloseReader(reader)

	logger.Infof("Discovery worker started, listening on topic: %s", interfaces.DiscoveryTaskTopic)

	for {
		msg, err := dw.ka.ReadMessage(ctx, reader)
		if err != nil {
			logger.Errorf("Failed to read message: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		taskMsg := &interfaces.DiscoveryTaskMessage{}
		if err := sonic.Unmarshal(msg.Value, taskMsg); err != nil {
			logger.Errorf("Failed to unmarshal message: %v", err)
			// 消息解析失败也需要提交位移，避免重复消费无效消息
			_ = dw.ka.CommitMessages(ctx, reader, msg)
			continue
		}

		err = dw.ExecuteDiscovery(ctx, taskMsg.TaskID)
		if err != nil {
			logger.Errorf("Failed to execute discovery: %v", err)
			// 业务失败不提交位移，下次重试
			continue
		}

		// 业务成功后手动提交位移
		if err := dw.ka.CommitMessages(ctx, reader, msg); err != nil {
			logger.Errorf("Failed to commit message: %v", err)
		}
	}
}

// ExecuteDiscovery executes discovery for a specific catalog.
// This method is called by the task management service.
func (dw *discoveryWorker) ExecuteDiscovery(ctx context.Context, taskID string) error {
	logger.Infof("Starting discovery for task: %s", taskID)

	task, err := dw.dts.GetByID(ctx, taskID)
	if err != nil {
		logger.Errorf("Failed to get task info for task %s: %v", taskID, err)
		return err
	}

	catalog, err := dw.cs.GetByID(ctx, task.CatalogID, true)
	if err != nil {
		logger.Errorf("Failed to get catalog for task %s: %v", taskID, err)
		return err
	}

	// Update task status to running and set start time
	now := time.Now().UnixMilli()
	if err := dw.dts.UpdateStatus(ctx, taskID, interfaces.DiscoveryTaskStatusRunning, "", now); err != nil {
		logger.Errorf("Failed to set start time for task %s: %v", taskID, err)
	}

	// Execute discovery
	result, err := dw.discoverCatalog(ctx, catalog)
	if err != nil {
		// Update task status to failed
		now = time.Now().UnixMilli()
		_ = dw.dts.UpdateStatus(ctx, taskID, interfaces.DiscoveryTaskStatusFailed, err.Error(), now)
		return err
	}

	// Update task result
	now = time.Now().UnixMilli()
	if err := dw.dts.UpdateResult(ctx, taskID, result, now); err != nil {
		logger.Errorf("Failed to update result for task %s: %v", taskID, err)
	}

	logger.Infof("Discovery completed for task: %s, catalog: %s", taskID, catalog.ID)
	return nil
}

// discoverCatalog discovers resources for a specific catalog.
func (dw *discoveryWorker) discoverCatalog(ctx context.Context,
	catalog *interfaces.Catalog) (*interfaces.DiscoveryResult, error) {

	logger.Infof("Starting discovery for catalog: %s", catalog.ID)

	// 验证 catalog 类型
	if catalog.Type != interfaces.CatalogTypePhysical {
		return nil, fmt.Errorf("discovery only supports physical catalogs")
	}

	// 1. 创建 Connector 并连接
	connector, err := dw.createAndConnectConnector(ctx, catalog)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to data source: %w", err)
	}
	defer connector.Close(ctx)

	// Update catalog metadata
	if meta, err := connector.GetMetadata(ctx); err == nil {
		if err := dw.cs.UpdateMetadata(ctx, catalog.ID, meta); err != nil {
			logger.Errorf("Failed to update catalog metadata: %v", err)
		}
	} else {
		logger.Warnf("Failed to get metadata: %v", err)
	}

	// 2. 根据 connector category 分发到不同的发现函数
	category := connector.GetCategory()
	switch category {
	case interfaces.ConnectorCategoryTable:
		return dw.discoverTableResources(ctx, catalog, connector)
	case interfaces.ConnectorCategoryIndex:
		return dw.discoverIndexResources(ctx, catalog, connector)
	case interfaces.ConnectorCategoryFile, interfaces.ConnectorCategoryFileset:
		return dw.discoverFileResources(ctx, catalog, connector)
	default:
		return nil, fmt.Errorf("unsupported connector category for discovery: %s", category)
	}
}

// discoverFileResources discovers file resources from a file connector.
func (dw *discoveryWorker) discoverFileResources(ctx context.Context,
	catalog *interfaces.Catalog, connector connectors.Connector) (*interfaces.DiscoveryResult, error) {
	// TODO: 实现文件资源发现逻辑
	return nil, fmt.Errorf("file resource discovery not implemented yet")
}

// createAndConnectConnector creates and connects a connector for the catalog.
func (dw *discoveryWorker) createAndConnectConnector(ctx context.Context,
	catalog *interfaces.Catalog) (connectors.Connector, error) {

	// 使用 mapstructure 反序列化 ConnectorConfig
	cfg := interfaces.ConnectorConfig(catalog.ConnectorConfig)

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
