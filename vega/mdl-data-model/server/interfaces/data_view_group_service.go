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
	//模块名称
	DATA_VIEW_GROUP = "data_view_group"

	OBJECTTYPE_DATA_VIEW_GROUP = "ID_AUDIT_DATA_VIEW_GROUP"

	DEFAULT_DATA_VIEW_GROUP_SORT = "name"
)

const (
	GroupID_All         = "__all"
	GroupName_All       = "__all"
	GroupID_IndexBase   = "__index_base"
	GroupName_IndexBase = "index_base"
)

var (
	DATA_VIEW_GROUP_SORT = map[string]string{
		"update_time": "f_update_time",
		"name":        "f_group_name",
	}
)

//go:generate mockgen -source ../interfaces/data_view_group_service.go -destination ../interfaces/mock/mock_data_view_group_service.go
type DataViewGroupService interface {
	CreateDataViewGroup(ctx context.Context, tx *sql.Tx, group *DataViewGroup) (string, error)
	DeleteDataViewGroup(ctx context.Context, groupID string, includeViews bool) ([]*SimpleDataView, error)
	UpdateDataViewGroup(ctx context.Context, group *DataViewGroup) error
	ListDataViewGroups(ctx context.Context, params *ListViewGroupQueryParams, includeViews bool) ([]*DataViewGroup, int, error)

	GetDataViewGroupByID(ctx context.Context, groupID string) (*DataViewGroup, error)
	CheckDataViewGroupExistByName(ctx context.Context, tx *sql.Tx, name string, builtin bool) (*DataViewGroup, bool, error)

	// 标记删除分组
	MarkDataViewGroupDeleted(ctx context.Context, groupID string, includeViews bool) error
}
