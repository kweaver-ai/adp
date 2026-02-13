// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
)

const (
	AR_ADMIN_ID string = "f5a08cb056c539d76f373bbc94458e39"
)

//go:generate mockgen -source ../interfaces/data_connection_access.go -destination ../interfaces/mock/mock_data_connection_access.go
type DataManagerAccess interface {
	GetLogGroupRoots(ctx context.Context, host string, accessToken string) (any, error)
	GetLogGroupTree(ctx context.Context, userID string, host string, accessToken string) (any, error)
	GetLogGroupChildren(ctx context.Context, logGroupID string, host string, accessToken string) (any, error)
}
