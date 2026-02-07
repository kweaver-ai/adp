// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"database/sql"
)

const (
	// 审计日志对象类型
	OBJECTTYPE_DATA_DICT_ITEM = "ID_AUDIT_DATA_DICT_ITEM"

	// 使用0值字节作为分隔符
	ITEM_KEY_SEPARATOR = "\x00"

	// 创建数据字典项批量大小
	CREATE_DATA_DICT_ITEM_SIZE = 1000

	DATA_DICT_ITEM_MAX_LENGTH = 3000

	DATA_DICT_ITEM_DEFAULT_SORT = "id"
)

var (
	DATA_DICT_ITEM_SORT = map[string]string{
		"id":    "f_item_id",
		"key":   "f_item_key",
		"value": "f_item_value",
	}
)

// logics操作接口
//
//go:generate mockgen -source ../interfaces/data_dict_item_service.go -destination ../interfaces/mock/mock_data_dict_item_service.go
type DataDictItemsService interface {
	GetDictItemsByItemIDs(ctx context.Context, dictStore string, itemIDs []string) ([]map[string]string, error)
	ListDataDictItems(ctx context.Context, dict DataDict, listDictItemsQuery DataDictItemQueryParams) ([]map[string]string, int, error)

	CreateDataDictItem(ctx context.Context, dict DataDict, dimension Dimension) (string, error)
	UpdateDataDictItem(ctx context.Context, dict DataDict, itemID string, dimension Dimension) error
	DeleteDataDictItem(ctx context.Context, dict DataDict, itemID string) error
	DeleteDataDictItems(ctx context.Context, dict DataDict) error

	ImportDataDictItems(ctx context.Context, dict *DataDict, items []map[string]string, mode string) error

	CreateKVDictItems(ctx context.Context, tx *sql.Tx, dictID string, items []map[string]string) error
	CreateDimensionDictItems(ctx context.Context, tx *sql.Tx, dictID string, dictStore string, dimension Dimension, items []map[string]string) error

	GetKVDictItems(ctx context.Context, dictID string) ([]map[string]string, error)
	GetDimensionDictItems(ctx context.Context, dictID string, dictStore string, dimension Dimension) ([]map[string]string, error)
	UpdateDimension(ctx context.Context, dictStore string, prefix string, new []DimensionItem, old []DimensionItem) ([]DimensionItem, bool, error)
}
