// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
)

//go:generate mockgen -source ../interfaces/objective_model_service.go -destination ../interfaces/mock/mock_objective_model_service.go
type ObjectiveModelService interface {
	Simulate(ctx context.Context, query ObjectiveModelQuery) (ObjectiveModelUniResponse, error)
	Exec(ctx context.Context, query ObjectiveModelQuery) (ObjectiveModelUniResponse, error)
}
