// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

// 索引库基础信息，供视图的数据源使用
type SimpleIndexBase struct {
	BaseType string `mapstructure:"base_type" json:"base_type"`
	Name     string `mapstructure:"name" json:"name"`
	Comment  string `mapstructure:"comment" json:"comment"`
}

// 索引库信息
type IndexBase struct {
	SimpleIndexBase
	DataType string           `json:"data_type"`
	Category string           `json:"category"`
	Indices  []string         `json:"indices"`
	Mappings Mappings         `json:"mappings"`
	Fields   []IndexBaseField `json:"-"`
}

// 索引库的字段信息
type Mappings struct {
	UserDefinedMappings []IndexBaseField `json:"user_defined_mappings"`
	MetaMappings        []IndexBaseField `json:"meta_mappings"`
	DynamicMappings     []IndexBaseField `json:"dynamic_mappings"`
}

type IndexBaseField struct {
	Field       string `json:"field"`
	Type        string `json:"type"`
	DisplayName string `json:"display_name"`
}

//go:generate mockgen -source ../interfaces/index_base_access.go -destination ../interfaces/mock/mock_index_base_access.go
type IndexBaseAccess interface {
	GetIndexBasesByTypes(ctx context.Context, baseTypes []string) ([]IndexBase, error)
	GetManyIndexBasesByTypes(ctx context.Context, types []string) ([]IndexBase, error)
	GetSimpleIndexBasesByTypes(ctx context.Context, baseTypes []string) ([]SimpleIndexBase, error)
	ListIndexBases(ctx context.Context) ([]SimpleIndexBase, error)
}
