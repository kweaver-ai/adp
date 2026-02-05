package interfaces

import (
	"context"
	"database/sql"
)

// 数据视图分组结构体
type DataViewGroup struct {
	GroupID       string `json:"id"`
	GroupName     string `json:"name"`
	CreateTime    int64  `json:"create_time"`
	UpdateTime    int64  `json:"update_time"`
	DeleteTime    int64  `json:"delete_time"`
	Builtin       bool   `json:"builtin"`
	DataViewCount int    `json:"data_view_count"`
}

type ListViewGroupQueryParams struct {
	Builtin        []bool
	IncludeDeleted bool
	PaginationQueryParameters
}

type MarkViewGroupDeletedParams struct {
	GroupID    string
	DeleteTime int64
}

//go:generate mockgen -source ../interfaces/data_view_group_access.go -destination ../interfaces/mock/mock_data_view_group_access.go
type DataViewGroupAccess interface {
	CreateDataViewGroup(ctx context.Context, tx *sql.Tx, dataViewGroup *DataViewGroup) error
	DeleteDataViewGroup(ctx context.Context, tx *sql.Tx, groupID string) error
	UpdateDataViewGroup(ctx context.Context, dataViewGroup *DataViewGroup) error
	ListDataViewGroups(ctx context.Context, params *ListViewGroupQueryParams) ([]*DataViewGroup, error)
	GetDataViewGroupsTotal(ctx context.Context, params *ListViewGroupQueryParams) (int, error)

	//判断分组是否存在
	GetDataViewGroupByID(ctx context.Context, groupID string) (*DataViewGroup, bool, error)
	//根据分组名称判断分组是否存在（创建重复）, 用于创建导入数据视图时，更新groupID
	CheckDataViewGroupExistByName(ctx context.Context, tx *sql.Tx, groupName string, builtin bool) (*DataViewGroup, bool, error)
	// 标记删除分组
	MarkDataViewGroupDeleted(ctx context.Context, tx *sql.Tx, params *MarkViewGroupDeletedParams) error
}
