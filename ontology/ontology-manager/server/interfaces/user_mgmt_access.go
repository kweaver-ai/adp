package interfaces

import (
	"context"
)

// UserMgmtAccess 定义用户管理相关的访问接口
//
//go:generate mockgen -source ../interfaces/user_mgmt_access.go -destination ../interfaces/mock/mock_user_mgmt_access.go
type UserMgmtAccess interface {
	GetAccountNames(ctx context.Context, accountInfos []*AccountInfo) error
}
