// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package discovery_task provides DiscoveryTask business logic.
package discovery_task

import (
	"context"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/hibiken/asynq"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/rs/xid"
	"go.opentelemetry.io/otel/trace"

	"vega-backend/common"
	asynq_access "vega-backend/drivenadapters/asynq"
	discoverytaskaccess "vega-backend/drivenadapters/discovery_task"
	"vega-backend/interfaces"
)

var (
	dtsOnce    sync.Once
	dtsService interfaces.DiscoveryTaskService
)

type discoveryTaskService struct {
	appSetting *common.AppSetting
	client     *asynq.Client
	dta        interfaces.DiscoveryTaskAccess
}

// NewDiscoveryTaskService creates or returns the singleton DiscoveryTaskService.
func NewDiscoveryTaskService(appSetting *common.AppSetting) interfaces.DiscoveryTaskService {
	dtsOnce.Do(func() {
		asynqAccess := asynq_access.NewAsynqAccess(appSetting)
		dtsService = &discoveryTaskService{
			appSetting: appSetting,
			client:     asynqAccess.CreateClient(context.Background()),
			dta:        discoverytaskaccess.NewDiscoveryTaskAccess(appSetting),
		}
	})
	return dtsService
}

// Create creates a new DiscoveryTask and enqueues it to the task queue.
func (dts *discoveryTaskService) Create(ctx context.Context, catalogID string) (string, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "DiscoveryTaskService.Create",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// Get account info from context
	accountInfo := interfaces.AccountInfo{}
	if ai, ok := ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo); ok {
		accountInfo = ai
	}

	now := time.Now().UnixMilli()
	task := &interfaces.DiscoveryTask{
		ID:          xid.New().String(),
		CatalogID:   catalogID,
		TriggerType: interfaces.DiscoveryTaskTriggerManual,
		Status:      interfaces.DiscoveryTaskStatusPending,
		Progress:    0,
		Message:     "",
		Creator:     accountInfo,
		CreateTime:  now,
	}

	// 1. Write to database
	if err := dts.dta.Create(ctx, task); err != nil {
		logger.Errorf("Failed to create discovery task: %v", err)
		o11y.Error(ctx, "Failed to create discovery task")
		return "", err
	}

	// 2. Enqueue task to task queue
	payload, err := sonic.Marshal(&interfaces.DiscoveryTaskMessage{
		TaskID: task.ID,
	})
	if err != nil {
		logger.Errorf("Failed to marshal discovery task: %v", err)
		o11y.Error(ctx, "Failed to marshal discovery task")
		return "", err
	}

	asynqTask := asynq.NewTask(interfaces.DiscoveryTaskType, payload)
	info, err := dts.client.Enqueue(asynqTask,
		asynq.Queue("high"),
		asynq.MaxRetry(3),
		asynq.Timeout(30*time.Minute),
	)
	if err != nil {
		logger.Errorf("Failed to enqueue discovery task: %v", err)
		o11y.Error(ctx, "Failed to enqueue discovery task")
		return "", err
	}

	logger.Infof("Enqueued task: id=%s, type=%s, queue=%s", info.ID, info.Type, info.Queue)
	return task.ID, nil
}

// GetByID retrieves a DiscoveryTask by ID.
func (dts *discoveryTaskService) GetByID(ctx context.Context, id string) (*interfaces.DiscoveryTask, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "DiscoveryTaskService.GetByID",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	return dts.dta.GetByID(ctx, id)
}

// List lists DiscoveryTasks for a catalog.
func (dts *discoveryTaskService) List(ctx context.Context, params interfaces.DiscoveryTaskQueryParams) ([]*interfaces.DiscoveryTask, int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "DiscoveryTaskService.List",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	return dts.dta.List(ctx, params)
}

// UpdateStatus updates a DiscoveryTask's status.
func (dts *discoveryTaskService) UpdateStatus(ctx context.Context, id, status, message string, stime int64) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "DiscoveryTaskService.UpdateStatus",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	return dts.dta.UpdateStatus(ctx, id, status, message, stime)
}

// UpdateResult updates a DiscoveryTask's result.
func (dts *discoveryTaskService) UpdateResult(ctx context.Context, id string, result *interfaces.DiscoveryResult, stime int64) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "DiscoveryTaskService.UpdateResult",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	return dts.dta.UpdateResult(ctx, id, result, stime)
}
