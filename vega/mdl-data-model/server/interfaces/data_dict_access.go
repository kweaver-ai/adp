// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"database/sql"
)

// 数据字典结构体
type DataDict struct {
	DictID     string   `json:"id"`
	DictName   string   `json:"name"`
	DictType   string   `json:"type"`
	UniqueKey  bool     `json:"unique_key"`
	Tags       []string `json:"tags"`
	Comment    string   `json:"comment"`
	CreateTime int64    `json:"create_time"`
	UpdateTime int64    `json:"update_time"`
	DictStore  string   `json:"-"`

	Dimension Dimension           `json:"dimension"`
	DictItems []map[string]string `json:"items"`

	// old format, compatibility
	OldDictName  string              `json:"dict_name,omitempty"`
	OldDictItems []map[string]string `json:"dict_items,omitempty"`
	// 操作权限
	Operations []string `json:"operations"`
}

// 数据字典表 dimension字段 结构体
type Dimension struct {
	Keys    []DimensionItem `json:"keys"`
	Values  []DimensionItem `json:"values"`
	Comment string          `json:"-"`
	ItemID  string          `json:"-"`
}
type DimensionItem struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Value string `json:",omitempty"`
}

// kv字典 字典项结构体
type KvDictItem struct {
	DictID  string `json:"dict_id,omitempty"`
	ItemID  string `json:"item_id,omitempty"`
	Key     string `json:"key"`
	Value   string `json:"value"`
	Comment string `json:"comment"`
}

// 数据库操作接口
//
//go:generate mockgen -source ../interfaces/data_dict_access.go -destination ../interfaces/mock/mock_data_dict_access.go
type DataDictAccess interface {
	ListDataDicts(ctx context.Context, dictQuery DataDictQueryParams) ([]DataDict, error)
	GetDictTotal(ctx context.Context, dictQuery DataDictQueryParams) (int64, error)
	GetDataDictByID(ctx context.Context, dictID string) (DataDict, error)
	CheckDictExistByName(ctx context.Context, dictName string) (bool, error)

	CreateDataDict(ctx context.Context, tx *sql.Tx, dict DataDict) error
	UpdateDataDict(ctx context.Context, dict DataDict) error
	DeleteDataDict(ctx context.Context, dictID string) (int64, error)

	CreateDimensionDictStore(ctx context.Context, dictStore string, dimension Dimension) error
	AddDimensionIndex(ctx context.Context, dictStore string, keys []DimensionItem) error
	DropDimensionIndex(ctx context.Context, dictStore string) error
	UpdateDictUpdateTime(ctx context.Context, dictID string, updateTime int64) error
}
