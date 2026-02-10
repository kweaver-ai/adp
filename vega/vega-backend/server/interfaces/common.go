// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package interfaces defines entities, DTOs, and service interfaces.
package interfaces

type contextKey string // 自定义专属的key类型

const (
	CONTENT_TYPE_NAME = "Content-Type"
	CONTENT_TYPE_JSON = "application/json"

	HTTP_HEADER_METHOD_OVERRIDE = "x-http-method-override"
	HTTP_HEADER_ACCOUNT_ID      = "x-account-id"
	HTTP_HEADER_ACCOUNT_TYPE    = "x-account-type"
	HTTP_HEADER_BUSINESS_DOMAIN = "x-business-domain"

	ACCOUNT_INFO_KEY contextKey = "x-account-info" // 避免直接使用string

	NAME_MAX_LENGTH        = 255
	DESCRIPTION_MAX_LENGTH = 1000
	TAGS_MAX_NUMBER        = 5
	TAG_MAX_LENGTH         = 40
	TAG_INVALID_CHARACTER  = "/:?\\\"<>|：？‘’“”！《》,#[]{}%&*$^!=.'"
)

// AccountInfo represents user/account information.
type AccountInfo struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name,omitempty"`
}

// PaginationParams holds common pagination parameters.
type PaginationParams struct {
	Offset    int
	Limit     int
	Sort      string
	Direction string
}

// ListResult wraps list response with pagination info.
type ListResult[T any] struct {
	Entries    []T   `json:"entries"`
	TotalCount int64 `json:"total_count"`
}

// SortField represents a field to sort by.
type SortField struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

// QueryParams represents query parameters for data retrieval.
type QueryParams struct {
	Start           int64       `json:"start,omitempty"`
	End             int64       `json:"end,omitempty"`
	Sort            []SortField `json:"sort,omitempty"`
	Offset          int         `json:"offset,omitempty"`
	Limit           int         `json:"limit,omitempty"`
	NeedTotal       bool        `json:"need_total,omitempty"`
	UseSearchAfter  bool        `json:"use_search_after,omitempty"`
	SearchAfter     []any       `json:"search_after,omitempty"`
	Format          string      `json:"format,omitempty"`
	Filters         interface{} `json:"filters,omitempty"`
}

// Default pagination values
const (
	DefaultOffset = 0
	DefaultLimit  = 20
	MaxLimit      = 100

	SortDirectionASC  = "asc"
	SortDirectionDESC = "desc"
)
