// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"database/sql"
)

//go:generate mockgen -source ../interfaces/data_dict_item_access.go -destination ../interfaces/mock/mock_data_dict_item_access.go
type DataDictItemAccess interface {
	CreateDataDictItem(ctx context.Context, dictID string, itemID string, dictStore string, dimension Dimension) error
	UpdateDataDictItem(ctx context.Context, dictID string, itemID string, dictStore string, dimension Dimension) error
	DeleteDataDictItem(ctx context.Context, dictID string, itemID string, dictStore string) (int64, error)
	// 事务内批量删除某个字典的字典项
	DeleteDataDictItemsByItemIDs(ctx context.Context, tx *sql.Tx, dictID string, itemIDs []string, dictStore string) error

	CreateKVDictItems(ctx context.Context, tx *sql.Tx, dictID string, items []KvDictItem) error
	CreateDimensionDictItems(ctx context.Context, tx *sql.Tx, dictID string, dictStore string, dimension []Dimension) error

	ListDataDictItems(ctx context.Context, dictID DataDict, listDictItemsQuery DataDictItemQueryParams) ([]map[string]string, error)
	GetDictItemTotal(ctx context.Context, dict DataDict, listDictItemsQuery DataDictItemQueryParams) (int, error)
	GetKVDictItems(ctx context.Context, dictID string) ([]map[string]string, error)
	GetDimensionDictItems(ctx context.Context, dictStore string, dimension Dimension) ([]map[string]string, error)

	AddDimensionColumn(ctx context.Context, dictStore string, new DimensionItem) error
	DropDimensionColumn(ctx context.Context, dictStore string, new DimensionItem) error

	DeleteDataDictItems(ctx context.Context, dictID string) error
	DeleteDimensionTable(ctx context.Context, dictStore string) error

	CountDictItemByKey(ctx context.Context, dictID string, dictStore string, keys []DimensionItem) (int, error)
	GetDictItemIDByKey(ctx context.Context, dictID string, dictStore string, keys []DimensionItem) ([]string, error)
	GetDictItemByItemID(ctx context.Context, dictStore string, itemID string) (map[string]string, error)
}
