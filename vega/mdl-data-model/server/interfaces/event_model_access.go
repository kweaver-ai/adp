// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"database/sql"
)

//go:generate mockgen -source ../interfaces/event_model_access.go -destination ../interfaces/mock/mock_event_model_access.go
type EventModelAccess interface {
	CreateEventModels(tx *sql.Tx, eventModels []EventModel) ([]map[string]any, error)
	GetEventModelByID(modelID string) (EventModel, error)
	UpdateEventModel(tx *sql.Tx, eventModel EventModel) error
	DeleteEventModels(tx *sql.Tx, eventModels []EventModel) error
	QueryEventModels(ctx context.Context, params EventModelQueryRequest) ([]EventModel, error)
	QueryTotalNumberEventModels(params EventModelQueryRequest) (int, error)
	GetEventModelNamesByIDs(params []string) ([]string, error)
	// UpdateDetectRule(detectRule DetectRule) error

	GetEventModelMapByNames(modelNames []string) (map[string]string, error)
	GetEventModelMapByIDs(modelIDs []string) (map[string]string, error)
	GetEventModelRefsByID(modelID string) (int, error)
	GetEventModelDependenceByID(modelID string) (int, error)

	//NOTE event model task
	CreateEventTask(ctx context.Context, tx *sql.Tx, metricTask EventTask) error
	UpdateEventTask(ctx context.Context, tx *sql.Tx, task EventTask) error
	// SetTaskSyncStatusByTaskID(ctx context.Context, tx *sql.Tx, taskSyncStatus EventTaskSyncStatus) error
	// SetTaskSyncStatusByModelID(ctx context.Context, tx *sql.Tx, taskSyncStatus EventTaskSyncStatus) error

	GetEventTaskByTaskID(ctx context.Context, taskID string) (EventTask, error)
	GetEventTaskByModelID(ctx context.Context, modelID string) (EventTask, bool, error)
	GetEventTaskIDByModelIDs(ctx context.Context, modelID []string) ([]string, error)
	DeleteEventTaskByTaskIDs(ctx context.Context, tx *sql.Tx, taskIDs []string) error
	UpdateEventTaskStatusInFinish(ctx context.Context, task EventTask) error
	// GetProcessingEventTasks(ctx context.Context) ([]EventTask, error)

	UpdateEventTaskAttributes(ctx context.Context, task EventTask) error
}
