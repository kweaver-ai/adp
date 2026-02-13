// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

const (
	DATA_CONNECTION_OBJECT_TYPE = "ID_AUDIT_DATA_CONNECTION"
	DATA_CONNECTION             = "data connection"
	ADDRESS_PATTERN             = `^[^,]*:([0-9]|[1-9]\d{1,3}|[1-5]\d{4}|6[0-4]\d{3}|65[0-4]\d{2}|655[0-2]\d|6553[0-5])$`
)

var (
	DATA_CONNECTION_SORT = map[string]string{
		"update_time": "f_update_time",
		"name":        "f_connection_name",
	}
)

// 通用列表查询参数
type CommonListQueryParams struct {
	NamePattern string
	Name        string
	Tag         string
	PaginationQueryParameters
}

// 链路模型列表查询参数结构体
type DataConnectionListQueryParams struct {
	CommonListQueryParams
	ApplicationScope []string
}

//go:generate mockgen -source ../interfaces/data_connection_service.go -destination ../interfaces/mock/mock_data_connection_service.go
type DataConnectionService interface {
	CreateDataConnection(ctx context.Context, conn *DataConnection) (string, error)
	DeleteDataConnections(ctx context.Context, connIDs []string) error
	UpdateDataConnection(ctx context.Context, conn *DataConnection, preConn *DataConnection) error
	GetDataConnection(ctx context.Context, connID string, withAuthInfo bool) (*DataConnection, bool, error)
	ListDataConnections(ctx context.Context, queryParams DataConnectionListQueryParams) ([]*DataConnectionListEntry, int, error)

	GetMapAboutName2ID(ctx context.Context, connNames []string) (map[string]string, error)
	GetMapAboutID2Name(ctx context.Context, connIDs []string) (map[string]string, error)
	GetDataConnectionSourceType(ctx context.Context, connID string) (string, bool, error)
}
