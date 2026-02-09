// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
)

//go:generate mockgen -source ../interfaces/vega_calculate_access.go -destination ../interfaces/mock/mock_vega_calculate_access.go
type VegaCalculateAccess interface {
	StatementQuery(ctx context.Context, sql string) (*VegaCalculateData, error)
	NextUriQuery(ctx context.Context, nextUri string) (*VegaCalculateData, error)
}

type VegaCalculateData struct {
	NextUri string    `json:"nextUri"`
	Columns []*Column `json:"columns"`
	Data    []*[]any  `json:"data"` // data的每个数组里面是一行数据,每行数据的位置与columns里的定义的字段顺序一致
	Stats   *Stats    `json:"stats"`
}

type Stats struct {
	State string `json:"state"`
}

type VegaCalculateError struct {
	Stats *Stats `json:"stats"`
	Error *Error `json:"error"`
}

type Error struct {
	Message   string `json:"message"`
	ErrorName string `json:"errorName"`
}
