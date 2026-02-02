package interfaces

import (
	"context"
	"database/sql"
)

//go:generate mockgen -source ../interfaces/action_type_service.go -destination ../interfaces/mock/mock_action_type_service.go
type ActionTypeService interface {
	CheckActionTypeExistByID(ctx context.Context, knID string, branch string, atID string) (string, bool, error)
	CheckActionTypeExistByName(ctx context.Context, knID string, branch string, atName string) (string, bool, error)
	CreateActionTypes(ctx context.Context, tx *sql.Tx, actionTypes []*ActionType, mode string) ([]string, error)
	ListActionTypes(ctx context.Context, query ActionTypesQueryParams) ([]*ActionType, int, error)
	GetActionTypesByIDs(ctx context.Context, knID string, branch string, atIDs []string) ([]*ActionType, error)
	UpdateActionType(ctx context.Context, tx *sql.Tx, actionType *ActionType) error
	DeleteActionTypesByIDs(ctx context.Context, tx *sql.Tx, knID string, branch string, atIDs []string) error

	GetActionTypeIDsByKnID(ctx context.Context, knID string, branch string) ([]string, error)
	DeleteActionTypesByKnID(ctx context.Context, tx *sql.Tx, knID string, branch string) error

	SearchActionTypes(ctx context.Context, query *ConceptsQuery) (ActionTypes, error)

	// 写行动类到索引
	InsertOpenSearchData(ctx context.Context, actionTypes []*ActionType) error
}
