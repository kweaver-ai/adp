// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"database/sql"
)

//go:generate mockgen -source ../interfaces/metric_model_service.go -destination ../interfaces/mock/mock_metric_model_service.go
type MetricModelService interface {
	// CreateMetricModel(ctx context.Context, metricModel MetricModel) (string, error)
	CreateMetricModels(ctx context.Context, metricModel []*MetricModel, mode string) ([]string, error)
	UpdateMetricModel(ctx context.Context, tx *sql.Tx, metricModel MetricModel) error
	DeleteMetricModels(ctx context.Context, tx *sql.Tx, modelIDs []string) (int64, error)
	// ListMetricModels(ctx context.Context, parameter MetricModelsQueryParams) ([]MetricModel, int, error)
	ListSimpleMetricModels(ctx context.Context, parameter MetricModelsQueryParams) ([]SimpleMetricModel, int, error)
	GetMetricModels(ctx context.Context, modelIDs []string, includeView bool) ([]MetricModelWithFilters, error)
	UpdateMetricModelsGroup(ctx context.Context, models map[string]SimpleMetricModel, group MetricModelGroupName) (int64, error) // 批量修改模型degroupID
	GetMetricModelIDByName(ctx context.Context, groupName, modelName string) (string, error)                                     // 根据名称获取到ID ，可以用于根据名称获取到指标模型对象信息
	GetMetricModelSourceFields(ctx context.Context, modelID string) ([]*ViewField, error)
	GetMetricModelOrderByFields(ctx context.Context, modelIDs []string) ([][]*ViewField, error) // 根据模型id获取模型所绑定的数据源的字段列表

	// 根据groupID 获取指标模型信息。
	GetMetricModelsDetailByGroupID(ctx context.Context, groupID string) ([]MetricModel, error) // 按分组导出接口
	GetMetricModelsByGroupID(ctx context.Context, groupID string) ([]MetricModel, error)       // (用于批量删除分组内指标模型),批量修改指标模型的分组时使用

	// 创建、更新、批量修改模型的分组时需要更新
	RetriveGroupIDByGroupName(ctx context.Context, tx *sql.Tx, groupName string) (MetricModelGroup, error)

	// 校验存在性
	CheckMetricModelExistByID(ctx context.Context, modelID string) (string, bool, error)
	CheckMetricModelExistByName(ctx context.Context, groupName, modelName string) (string, bool, error)
	CheckMetricModelByMeasureName(ctx context.Context, measureName string) (string, bool, error) // 度量名称全局唯一，创建、更新时校验度量名称是否存在

	// 按id或者模型信息。更新模型、删除模型、批量修改模型分组、指标模型任务同步、事件模型时调用此函数
	GetMetricModelByModelID(ctx context.Context, modelID string) (MetricModel, error)
	// 事件模型、目标模型调用此函数根据id获取模型简单信息。创建时校验id的重复性也用了此函数
	GetMetricModelSimpleInfosByIDs(ctx context.Context, modelIDs []string) (map[string]SimpleMetricModel, error)

	// 计算公式有效性校验
	CheckFormula(ctx context.Context, metricModel MetricModel) error

	// 获取指标模型的资源实例列表
	ListMetricModelSrcs(ctx context.Context, parameter MetricModelsQueryParams) ([]Resource, int, error)
}
