// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"net/url"
)

//go:generate mockgen -source ../interfaces/data_connection_access.go -destination ../interfaces/mock/mock_data_connection_access.go
type SearchAccess interface {
	SearchSubmit(ctx context.Context, queryBody any, userID string, host string, accessToken string) (any, error)
	SearchFetch(ctx context.Context, jobID string, queryParams url.Values, host string, accessToken string) (any, error)
	SearchFetchFields(ctx context.Context, jobID string, queryParams url.Values, host string, accessToken string) (any, error)
	SearchFetchSameFields(ctx context.Context, jobID string, host string, accessToken string) (any, error)
	SearchContext(ctx context.Context, queryParams url.Values, userID string, host string, accessToken string) (any, error)
}
