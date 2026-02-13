// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"time"

	"github.com/google/uuid"
)

//go:generate mockgen -source ../interfaces/catalog_code_access.go -destination ../interfaces/mock/mock_catalog_code_access.go
type CatalogCodeAccess interface {
	// 生成指定数量的编码
	Generate(ctx context.Context, id uuid.UUID, count int) (*CodeList, error)
	GetTimestampBlacklist(ctx context.Context) ([]string, error)

	GetUserRoles(ctx context.Context) ([]*RoleEntry, error)
	GetUserInfos(ctx context.Context, userId string) (*GetUserInfoRes, error)
}

// 编码列表
type CodeList struct {
	// 编码列表
	Entries []string `json:"entries,omitempty"`
	// 编码的数量
	TotalCount int `json:"total_count,omitempty"`
}

type RoleEntry struct {
	ID   string `json:"id"`   // 角色ID
	Name string `json:"name"` // 角色名称
}

type GetUserInfoRes struct {
	Id        string      `json:"id"`
	CreatedAt interface{} `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	CreatedBy struct {
	} `json:"created_by"`
	UpdatedBy struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	} `json:"updated_by"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	LoginName   string `json:"login_name"`
	Scope       string `json:"scope"`
	UserType    int    `json:"user_type"`
	Status      int    `json:"status"`
	ParentDeps  []struct {
		PathId string `json:"path_id"`
		Path   string `json:"path"`
	} `json:"parent_deps"`
	Roles []struct {
		Id        string    `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		CreatedBy string    `json:"created_by"`
		Name      string    `json:"name"`
		Type      string    `json:"type"`
		Color     string    `json:"color"`
		Scope     string    `json:"scope"`
		Icon      string    `json:"icon"`
		System    int       `json:"System"`
	} `json:"roles"`
	Permissions []struct {
		Scope     string    `json:"scope"`
		Id        string    `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Name      string    `json:"name"`
		Category  string    `json:"category"`
	} `json:"permissions"`
}
