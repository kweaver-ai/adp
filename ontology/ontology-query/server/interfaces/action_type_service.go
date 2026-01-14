package interfaces

import "context"

//go:generate mockgen -source ../interfaces/action_type_service.go -destination ../interfaces/mock/mock_action_type_service.go
type ActionTypeService interface {
	GetActionsByActionTypeID(ctx context.Context, query *ActionQuery) (Actions, error)
}
