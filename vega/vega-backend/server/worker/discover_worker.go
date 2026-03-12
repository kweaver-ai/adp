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
	"vega-backend/logics/discover_task"
	"vega-backend/logics/resource"
)

var (
	dWorkerOnce sync.Once
	dWorker     interfaces.DiscoverWorker
)

// discoverWorker provides resource discover functionality.
type discoverWorker struct {
	appSetting *common.AppSetting
	aqa        interfaces.AsynqAccess
	rs         interfaces.ResourceService
	cs         interfaces.CatalogService
	dts        interfaces.DiscoverTaskService
}

// NewDiscoverWorker creates or returns the singleton DiscoverWorker.
func NewDiscoverWorker(appSetting *common.AppSetting) interfaces.DiscoverWorker {
	dWorkerOnce.Do(func() {
		dWorker = &discoverWorker{
			appSetting: appSetting,
			aqa:        asynq_access.NewAsynqAccess(appSetting),
			rs:         resource.NewResourceService(appSetting),
			cs:         logicsCatalog.NewCatalogService(appSetting),
			dts:        discover_task.NewDiscoverTaskService(appSetting),
		}
	})
	return dWorker
}

func (dw *discoverWorker) Start() {
	// Start server in a goroutine
	go func() {
		for {
			if err := dw.Run(context.Background()); err != nil {
				logger.Errorf("Discover worker failed: %v", err)
			}
			time.Sleep(1 * time.Second)
		}
	}()
}

func (dw *discoverWorker) Run(ctx context.Context) error {
	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("Discover worker failed: %v", err)
		}
	}()

	srv := dw.aqa.CreateServer(ctx)

	// Register task handler
	mux := asynq.NewServeMux()
	mux.Handle(interfaces.DiscoverTaskType, dw)

	logger.Infof("Discover worker starting, listening for task type: %s", interfaces.DiscoverTaskType)
	if err := srv.Run(mux); err != nil {
		logger.Errorf("Discover worker failed: %v", err)
		return err
	}
	return nil
}

// handleDiscoverTask handles a discover task from the queue.
func (dw *discoverWorker) ProcessTask(ctx context.Context, event *asynq.Task) error {
	var msg interfaces.DiscoverTaskMessage
	if err := sonic.Unmarshal(event.Payload(), &msg); err != nil {
		logger.Errorf("Failed to unmarshal task message: %v", err)
		return err
	}

	taskID := msg.TaskID
	logger.Infof("Starting discover for task: %s", taskID)

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
	if err := dw.dts.UpdateStatus(ctx, taskID, interfaces.DiscoverTaskStatusRunning, "", now); err != nil {
		logger.Errorf("Failed to set start time for task %s: %v", taskID, err)
	}
	// Execute discover : 元数据采集主要逻辑
	//首先根据 catalog ID 获取 catalog 信息，
	//然后根据 catalog 信息获取 connector 信息，
	//然后根据 connector 信息获取 connector 实例，
	//然后根据 connector 实例获取 catalog 的元数据，
	//然后根据 catalog 的元数据获取 catalog 的资源信息：元数据
	result, err := dw.discoverCatalog(ctx, catalog)
	if err != nil {
		// Update task status to failed
		now = time.Now().UnixMilli()
		_ = dw.dts.UpdateStatus(ctx, taskID, interfaces.DiscoverTaskStatusFailed, err.Error(), now)
		return err
	}

	// Update task result
	now = time.Now().UnixMilli()
	if err := dw.dts.UpdateResult(ctx, taskID, result, now); err != nil {
		logger.Errorf("Failed to update result for task %s: %v", taskID, err)
	}

	logger.Infof("Discover completed for task: %s, catalog: %s", taskID, catalog.ID)
	return nil
}

// discoverCatalog discovers resources for a specific catalog.
// discoverCatalog 是一个发现目录资源的方法
// 它接收上下文和目录信息，返回发现结果或错误
// 参数:
//   - ctx: 上下文信息，用于控制请求的超时和取消
//   - catalog: 目录信息，包含目录ID和类型等
//
// 返回值:
//   - *interfaces.DiscoverResult: 发现结果，包含发现的资源信息
//   - error: 错误信息，如果发现过程中出现错误
func (dw *discoverWorker) discoverCatalog(ctx context.Context,
	catalog *interfaces.Catalog) (*interfaces.DiscoverResult, error) {

	logger.Infof("Starting discover for catalog: %s", catalog.ID)

	// 验证 catalog 类型
	if catalog.Type != interfaces.CatalogTypePhysical {
		return nil, fmt.Errorf("discover only supports physical catalogs")
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
	// 2. 根据 connector category 分发到不同的发现函数：例如mysql会到mysql.go下面进行元数据的采集，里面会有具体的实现
	category := connector.GetCategory()
	switch category {
	// table类型的会到这里，例如mysql
	case interfaces.ConnectorCategoryTable:
		return dw.discoverTableResources(ctx, catalog, connector)
	// index类型的会到这里，例如open search
	case interfaces.ConnectorCategoryIndex:
		return dw.discoverIndexResources(ctx, catalog, connector)
	case interfaces.ConnectorCategoryFile, interfaces.ConnectorCategoryFileset:
		return dw.discoverFileResources(ctx, catalog, connector)
	default:
		return nil, fmt.Errorf("unsupported connector category for discover: %s", category)
	}
}

// discoverFileResources discovers file resources from a file connector.
func (dw *discoverWorker) discoverFileResources(ctx context.Context,
	catalog *interfaces.Catalog, connector connectors.Connector) (*interfaces.DiscoverResult, error) {
	// TODO: 实现文件资源发现逻辑
	return nil, fmt.Errorf("file resource discover not implemented yet")
}

// createAndConnectConnector creates and connects a connector for the catalog.
func (dw *discoverWorker) createAndConnectConnector(ctx context.Context,
	catalog *interfaces.Catalog) (connectors.Connector, error) {

	// 创建 connector
	connector, err := factory.GetFactory().CreateConnectorInstance(ctx, catalog.ConnectorType, catalog.ConnectorCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create connector: %w", err)
	}

	// 连接
	if err := connector.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return connector, nil
}
