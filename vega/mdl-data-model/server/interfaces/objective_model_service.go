// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"database/sql"
)

//go:generate mockgen -source ../interfaces/objective_model_service.go -destination ../interfaces/mock/mock_objective_model_service.go
type ObjectiveModelService interface {
	CheckObjectiveModelExistByID(ctx context.Context, modelID string) (string, bool, error)
	CheckObjectiveModelExistByName(ctx context.Context, modelName string) (string, bool, error)
	CreateObjectiveModels(ctx context.Context, objectiveModel []*ObjectiveModel, mode string) ([]string, error)
	ListObjectiveModels(ctx context.Context, parameter ObjectiveModelsQueryParams) ([]ObjectiveModel, int, error)
	GetObjectiveModels(ctx context.Context, modelIDs []string) ([]ObjectiveModel, error)
	UpdateObjectiveModel(ctx context.Context, tx *sql.Tx, objectiveModel ObjectiveModel) error
	DeleteObjectiveModels(ctx context.Context, modelIDs []string) (int64, error)

	// 获取指标模型的资源实例列表
	ListObjectiveModelSrcs(ctx context.Context, parameter ObjectiveModelsQueryParams) ([]Resource, int, error)
}
