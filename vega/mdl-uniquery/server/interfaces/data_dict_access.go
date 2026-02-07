// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
)

// 数据字典结构体
type DataDict struct {
	DictID     string `json:"id"`
	DictName   string `json:"name"`
	UniqueKey  bool   `json:"unique_key"`
	CreateTime int64  `json:"create_time"`
	UpdateTime int64  `json:"update_time"`
	// DictType   string   `json:"type"`
	// Comment    string   `json:"comment"`
	// Tags       []string `json:"tags"`
	// DictStore  string   `json:"-"`

	Dimension Dimension           `json:"dimension"`
	DictItems []map[string]string `json:"items"`

	// map的键为字典维度键对应的值拼接的字符串
	DictRecords map[string][]map[string]string `json:"-"`
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

//go:generate mockgen -source ../interfaces/data_dict_access.go -destination ../interfaces/mock/mock_data_dict_access.go
type DataDictAccess interface {
	GetDictInfo(ctx context.Context, dictName string) (DataDict, error)
	GetDictIteams(ctx context.Context, dictID string) (map[string][]map[string]string, error)
}
