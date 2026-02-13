// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"database/sql"
)

//go:generate mockgen -source ../interfaces/metric_model_group_access.go -destination ../interfaces/mock/mock_metric_model_group_access.go
type MetricModelGroupAccess interface {
	CreateMetricModelGroup(ctx context.Context, tx *sql.Tx, metricModelGroup MetricModelGroup) error
	DeleteMetricModelGroup(ctx context.Context, tx *sql.Tx, groupID string) (int64, error)
	UpdateMetricModelGroup(ctx context.Context, metricModelGroup MetricModelGroup) error
	ListMetricModelGroups(ctx context.Context, params ListMetricGroupQueryParams) ([]*MetricModelGroup, error)
	GetMetricModelGroupsTotal(ctx context.Context, params ListMetricGroupQueryParams) (int, error)

	//根据分组名称判断分组是否存在（创建重复）
	GetMetricModelGroupByName(ctx context.Context, tx *sql.Tx, groupName string) (MetricModelGroup, bool, error)
	CheckMetricModelGroupExist(ctx context.Context, groupName string) (MetricModelGroup, bool, error)
	//判断分组是否存在
	GetMetricModelGroupByID(ctx context.Context, groupID string) (MetricModelGroup, bool, error)
}
