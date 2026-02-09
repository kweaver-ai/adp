// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"database/sql"
)

//go:generate mockgen -source ../interfaces/metric_model_access.go -destination ../interfaces/mock/mock_metric_model_access.go
type MetricModelAccess interface {
	CheckMetricModelExistByID(ctx context.Context, modelID string) (string, bool, error)
	// CheckMetricModelExistByName(ctx context.Context, combinationName CombinationName) (string, bool, error)
	CheckMetricModelByMeasureName(ctx context.Context, measureName string) (string, bool, error)

	CreateMetricModel(ctx context.Context, tx *sql.Tx, metricModel MetricModel) error
	UpdateMetricModel(ctx context.Context, tx *sql.Tx, metricModel MetricModel) error
	DeleteMetricModels(ctx context.Context, tx *sql.Tx, modelIDs []string) (int64, error)
	// ListMetricModels(ctx context.Context, modelsQuery MetricModelsQueryParams) ([]MetricModel, error)
	ListSimpleMetricModels(ctx context.Context, modelsQuery MetricModelsQueryParams) ([]SimpleMetricModel, error)
	GetMetricModelsTotal(ctx context.Context, modelsQuery MetricModelsQueryParams) (int, error)

	GetMetricModelByModelID(ctx context.Context, modelID string) (MetricModel, bool, error)
	GetMetricModelIDByName(ctx context.Context, groupName, modelName string) (string, bool, error)
	GetMetricModelsByModelIDs(ctx context.Context, modelIDs []string) ([]MetricModel, error)
	GetMetricModelSimpleInfosByIDs(ctx context.Context, modelIDs []string) (map[string]SimpleMetricModel, error)

	// 仅给结构模型使用，待结构模型下线，删除下面2个函数
	// GetMetricModelIDsBySimpleInfos(ctx context.Context, simpleInfos []ModelSimpleInfo) (map[ModelSimpleInfo]string, error)
	// GetMetricModelSimpleInfosByIDs2(ctx context.Context, modelIDs []string) (map[string]ModelSimpleInfo, error)

	// 获取一个分组内所有指标模型（可以用来批量删除）
	GetMetricModelsByGroupID(ctx context.Context, groupID string) ([]MetricModel, error)

	// 修改指标模型的groupID （批量修改分组）
	UpdateMetricModelsGroupID(ctx context.Context, tx *sql.Tx, modelIDs []string, groupID string) (int64, error)
}
