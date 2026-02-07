// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

const (
	DATA_DICT = "data dict"

	// 审计日志对象类型
	OBJECTTYPE_DATA_DICT = "ID_AUDIT_DATA_DICT"
)

var (
	DATA_DICT_NAME_MAX_LENGTH      = 40
	DATA_DICT_DIMENSION_MAX_LENGTH = 15
	DATA_DICT_SORT                 = map[string]string{
		"update_time": "f_update_time",
		"name":        "f_dict_name",
	}

	DATA_DICT_TYPE_KV        = "kv_dict"
	DATA_DICT_TYPE_DIMENSION = "dimension_dict"

	DATA_DICT_STORE_DEFAULT = "t_data_dict_item"

	DATA_DICT_KV_DIMENSION = Dimension{
		Keys: []DimensionItem{
			{
				ID:   "item_key",
				Name: "key",
			},
		},
		Values: []DimensionItem{
			{
				ID:   "item_value",
				Name: "value",
			},
		},
	}

	DATA_DICT_DIMENSION_PREFIX_TABLE = "t_data_dict_dimension"
	DATA_DICT_DIMENSION_PREFIX_KEY   = "f_key"
	DATA_DICT_DIMENSION_PREFIX_VALUE = "f_value"
	DATA_DICT_DIMENSION_NAME_ID      = "id"
	DATA_DICT_DIMENSION_NAME_COMMENT = "comment"
)

// 数据字典列表查询参数
type DataDictQueryParams struct {
	NamePattern string
	Name        string
	Tag         string
	Type        string
	PaginationQueryParameters
}

// 数据字典项列表查询参数
type DataDictItemQueryParams struct {
	Patterns []DataDictItemQueryPattern
	PaginationQueryParameters
}

// 数据字典项列表查询参数
type DataDictItemQueryPattern struct {
	QueryField   string
	QueryPattern string
}

// logics操作接口
//
//go:generate mockgen -source ../interfaces/data_dict_service.go -destination ../interfaces/mock/mock_data_dict_service.go
type DataDictService interface {
	ListDataDicts(ctx context.Context, listDictsQuery DataDictQueryParams) ([]DataDict, int64, error)
	GetDataDicts(ctx context.Context, dictIDs []string) ([]DataDict, error)
	CreateDataDict(ctx context.Context, dict DataDict) (string, error)
	UpdateDataDict(ctx context.Context, dict DataDict) error
	DeleteDataDict(ctx context.Context, dict DataDict) (int64, error)

	GetDataDictByID(ctx context.Context, dictID string) (DataDict, error)
	CheckDictExistByName(ctx context.Context, dictName string) (bool, error)

	ListDataDictSrcs(ctx context.Context, listDictsQuery DataDictQueryParams) ([]Resource, int, error)
}
