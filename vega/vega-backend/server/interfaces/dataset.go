// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

// SortField represents a field to sort by.
type SortField struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

// DatasetQueryParams represents query parameters for data retrieval.
type DatasetQueryParams struct {
	Start          int64       `json:"start,omitempty"`
	End            int64       `json:"end,omitempty"`
	Sort           []SortField `json:"sort,omitempty"`
	Offset         int         `json:"offset,omitempty"`
	Limit          int         `json:"limit,omitempty"`
	NeedTotal      bool        `json:"need_total,omitempty"`
	UseSearchAfter bool        `json:"use_search_after,omitempty"`
	SearchAfter    []any       `json:"search_after,omitempty"`
	Format         string      `json:"format,omitempty"`
	Filters        interface{} `json:"filters,omitempty"`
}
