// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

//go:generate mockgen -source ../interfaces/metric_model_access.go -destination ../interfaces/mock/mock_metric_model_access.go
type MetricModelAccess interface {
	GetMetricModel(ctx context.Context, modelId string) ([]MetricModel, bool, error)
	GetMetricModelIDByName(ctx context.Context, groupName, modelName string) (string, bool, error)

	GetMetricModels(ctx context.Context, modelId []string) ([]MetricModel, error)
}
