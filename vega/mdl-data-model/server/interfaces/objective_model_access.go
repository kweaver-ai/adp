// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"database/sql"
)

//go:generate mockgen -source ../interfaces/objective_model_access.go -destination ../interfaces/mock/mock_objective_model_access.go
type ObjectiveModelAccess interface {
	CheckObjectiveModelExistByID(ctx context.Context, modelID string) (string, bool, error)
	CheckObjectiveModelExistByName(ctx context.Context, modelName string) (string, bool, error)

	CreateObjectiveModel(ctx context.Context, tx *sql.Tx, objectiveModel ObjectiveModel) error
	ListObjectiveModels(ctx context.Context, modelsQuery ObjectiveModelsQueryParams) ([]ObjectiveModel, error)
	GetObjectiveModelsTotal(ctx context.Context, modelsQuery ObjectiveModelsQueryParams) (int, error)
	GetObjectiveModelsByModelIDs(ctx context.Context, modelIDs []string) ([]ObjectiveModel, error)
	UpdateObjectiveModel(ctx context.Context, tx *sql.Tx, objectiveModel ObjectiveModel) error
	DeleteObjectiveModels(ctx context.Context, tx *sql.Tx, modelIDs []string) (int64, error)
}
