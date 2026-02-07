// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	cond "uniquery/common/condition"
)

//go:generate mockgen -source ../interfaces/metric_model_service.go -destination ../interfaces/mock/mock_metric_model_service.go
type MetricModelService interface {
	Simulate(ctx context.Context, query MetricModelQuery) (MetricModelUniResponse, error)
	Exec(ctx context.Context, query *MetricModelQuery) (MetricModelUniResponse, int, int, error)
	GetMetricModelIDByName(ctx context.Context, groupName, modelName string) (string, error)
	GetIndexBaseSplitTime()
	GetMetricModelFields(ctx context.Context, modelID string) ([]Field, error)
	GetMetricModelFieldValues(ctx context.Context, modelID, fieldName string) (FieldValues, error)
	GetMetricModelLabels(ctx context.Context, modelID string) ([]*cond.ViewField, error)
}
