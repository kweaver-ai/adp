// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"net/url"
)

//go:generate mockgen -source ../interfaces/log_group_service.go -destination ../interfaces/mock/mock_log_group_service.go
type LogGroupService interface {
	GetLogGroupRootsByConn(ctx context.Context, conn *DataConnection) (any, error)
	GetLogGroupTreeByConn(ctx context.Context, userID string, conn *DataConnection) (any, error)
	GetLogGroupChildrenByConn(ctx context.Context, logGroupID string, conn *DataConnection) (any, error)

	SearchSubmitByConn(ctx context.Context, queryBody any, userID string, conn *DataConnection) (any, error)
	SearchFetchByConn(ctx context.Context, jobID string, queryParams url.Values, conn *DataConnection) (any, error)
	SearchFetchFieldsByConn(ctx context.Context, jobID string, queryParams url.Values, conn *DataConnection) (any, error)
	SearchFetchSameFieldsByConn(ctx context.Context, jobID string, conn *DataConnection) (any, error)
	SearchContextByConn(ctx context.Context, queryParams url.Values, userID string, conn *DataConnection) (any, error)
}
