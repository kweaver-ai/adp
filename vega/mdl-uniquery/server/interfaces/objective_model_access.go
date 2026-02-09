// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

//go:generate mockgen -source ../interfaces/objective_model_access.go -destination ../interfaces/mock/mock_objective_model_access.go
type ObjectiveModelAccess interface {
	GetObjectiveModel(ctx context.Context, modelId string) (ObjectiveModel, bool, error)
}
