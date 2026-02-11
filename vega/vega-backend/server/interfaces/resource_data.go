// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "time"

// SortField represents a field to sort by.
type SortField struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

// ResourceDataQueryParams represents query parameters for data retrieval.
type ResourceDataQueryParams struct {
	Offset         int         `json:"offset,omitempty"`
	Limit          int         `json:"limit,omitempty"`
	Sort           []SortField `json:"sort,omitempty"`
	UseSearchAfter bool        `json:"use_search_after,omitempty"`
	SearchAfter    []any       `json:"search_after,omitempty"`

	Filters any `json:"filters,omitempty"`

	OutputFields []string `json:"output_fields"` // 指定输出的字段列表

	NeedTotal bool          `json:"need_total,omitempty"`
	Format    string        `json:"-"`
	Timeout   time.Duration `json:"-"` // 超时时间，查询参数

	//ActualCondition *cond.CondCfg `json:"-"`
}
