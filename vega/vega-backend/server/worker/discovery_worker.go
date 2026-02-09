// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package worker provides background workers for VEGA Manager.
package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/hibiken/asynq"
	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"vega-backend/common"
	asynq_access "vega-backend/drivenadapters/asynq"
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
	aqa        interfaces.AsynqAccess
	rs         interfaces.ResourceService
	cs         interfaces.CatalogService
	dts        interfaces.DiscoveryTaskService
}

// NewDiscoveryWorker creates or returns the singleton DiscoveryWorker.
func NewDiscoveryWorker(appSetting *common.AppSetting) interfaces.DiscoveryWorker {
	dWorkerOnce.Do(func() {
		dWorker = &discoveryWorker{
			appSetting: appSetting,
			aqa:        asynq_access.NewAsynqAccess(appSetting),
			rs:         resource.NewResourceService(appSetting),
			cs:         logicsCatalog.NewCatalogService(appSetting),
			dts:        discovery_task.NewDiscoveryTaskService(appSetting),
		}
	})
	return dWorker
}

func (dw *discoveryWorker) Start() {
	// Start server in a goroutine
	go func() {
		for {
			if err := dw.Run(context.Background()); err != nil {
				logger.Errorf("Discovery worker failed: %v", err)
			}
			time.Sleep(1 * time.Second)
		}
	}()
}

func (dw *discoveryWorker) Run(ctx context.Context) error {
	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("Discovery worker failed: %v", err)
		}
	}()

	srv := dw.aqa.CreateServer(ctx)

	// Register task handler
	mux := asynq.NewServeMux()
	mux.Handle(interfaces.DiscoveryTaskType, dw)

	logger.Infof("Discovery worker starting, listening for task type: %s", interfaces.DiscoveryTaskType)
	if err := srv.Run(mux); err != nil {
		logger.Errorf("Discovery worker failed: %v", err)
		return err
	}
	return nil
}

// handleDiscoveryTask handles a discovery task from the queue.
func (dw *discoveryWorker) ProcessTask(ctx context.Context, event *asynq.Task) error {
	var msg interfaces.DiscoveryTaskMessage
	if err := sonic.Unmarshal(event.Payload(), &msg); err != nil {
		logger.Errorf("Failed to unmarshal task message: %v", err)
		return err
	}

	taskID := msg.TaskID
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
