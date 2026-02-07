// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
)

//go:generate mockgen -source ../interfaces/metric_model_access.go -destination ../interfaces/mock/mock_metric_model_access.go
type MetricModelAccess interface {
	GetTaskPlanTimeById(ctx context.Context, taskId string) (int64, error)
	UpdateTaskAttributesById(ctx context.Context, taskId string, task MetricTask) error
}
