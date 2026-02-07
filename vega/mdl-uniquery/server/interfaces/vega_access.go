// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
)

//go:generate mockgen -source ../interfaces/vega_access.go -destination ../interfaces/mock/mock_vega_access.go
type VegaAccess interface {
	GetVegaViewFieldsByID(ctx context.Context, viewID string) (VegaViewWithFields, error)
	FetchDatasFromVega(ctx context.Context, nextUri string, sql string) (VegaFetchData, error)
}
