// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"database/sql"
)

//go:generate mockgen -source ../interfaces/metric_model_task_service.go -destination ../interfaces/mock/mock_metric_model_task_service.go
type MetricModelTaskService interface {
	CreateMetricTask(ctx context.Context, tx *sql.Tx, tasks MetricTask) error
	GetMetricTaskIDsByModelIDs(ctx context.Context, modelIDs []string) ([]string, error)
	GetMetricTasksByTaskIDs(ctx context.Context, taskIDs []string) ([]MetricTask, error)
	GetMetricTasksByModelIDs(ctx context.Context, modelIDs []string) (map[string]MetricTask, error)
	// GetProcessingMetricTasks(ctx context.Context) ([]MetricTask, error)
	UpdateMetricTask(ctx context.Context, tx *sql.Tx, task MetricTask) error
	UpdateMetricTaskStatusInFinish(ctx context.Context, task MetricTask) error
	UpdateMetricTaskAttributes(ctx context.Context, task MetricTask) error
	// SetTaskSyncStatusByTaskIDs(ctx context.Context, tx *sql.Tx, taskSyncStatus TaskSyncStatus) error
	// SetTaskSyncStatusByModelIDs(ctx context.Context, tx *sql.Tx, taskSyncStatus TaskSyncStatus) error
	DeleteMetricTaskByTaskIDs(ctx context.Context, tx *sql.Tx, taskIDs []string) error

	CheckMetricModelTaskExistByName(ctx context.Context, task MetricTask, deleteTaskIDs []string) (bool, error)
}
