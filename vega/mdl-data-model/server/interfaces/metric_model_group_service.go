// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
)

//go:generate mockgen -source ../interfaces/metric_model_group_service.go -destination ../interfaces/mock/mock_metric_model_group_service.go
type MetricModelGroupService interface {
	CheckMetricModelGroupExist(ctx context.Context, groupName string) (bool, error)

	// 若未找到即MetricModelGroup为空，需要报错，404错误
	GetMetricModelGroupByID(ctx context.Context, groupID string) (MetricModelGroup, error)
	CreateMetricModelGroup(ctx context.Context, metricModelGroup MetricModelGroup) (string, error)
	UpdateMetricModelGroup(ctx context.Context, metricModelGroup MetricModelGroup) error
	ListMetricModelGroups(ctx context.Context, parameter ListMetricGroupQueryParams) ([]*MetricModelGroup, int, error)
	DeleteMetricModelGroup(ctx context.Context, groupID string, force bool) (int64, []MetricModel, error)
	DeleteMetricModelGroupAndModels(ctx context.Context, groupID string, modelIDs []string) (int64, error)
}
