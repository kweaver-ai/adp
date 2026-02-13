// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"database/sql"
)

type DataConnection struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Tags       []string `json:"tags"`
	Comment    string   `json:"comment"`
	CreateTime int64    `json:"create_time"`
	UpdateTime int64    `json:"update_time"`

	DataSourceType      string   `json:"data_source_type"`
	DataSourceConfig    any      `json:"config"`
	DataSourceConfigMD5 string   `json:"-"`
	ApplicationScope    []string `json:"application_scope"`

	DataConnectionStatus `json:",inline"`
}

type DataConnectionStatus struct {
	ID            string `json:"-"`
	Status        string `json:"status"`
	DetectionTime int64  `json:"detection_time"`
}

type DataConnectionListEntry struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Tags       []string `json:"tags"`
	Comment    string   `json:"comment"`
	CreateTime int64    `json:"create_time"`
	UpdateTime int64    `json:"update_time"`

	DataSourceType   string   `json:"data_source_type"`
	ApplicationScope []string `json:"application_scope"`
}

//go:generate mockgen -source ../interfaces/data_connection_access.go -destination ../interfaces/mock/mock_data_connection_access.go
type DataConnectionAccess interface {
	CreateDataConnection(ctx context.Context, tx *sql.Tx, conn *DataConnection) error
	DeleteDataConnections(ctx context.Context, tx *sql.Tx, connIDs []string) error
	UpdateDataConnection(ctx context.Context, tx *sql.Tx, conn *DataConnection) error
	GetDataConnection(ctx context.Context, connID string) (*DataConnection, bool, error)
	ListDataConnections(ctx context.Context, queryParams DataConnectionListQueryParams) ([]*DataConnectionListEntry, error)
	GetDataConnectionTotal(ctx context.Context, queryParams DataConnectionListQueryParams) (int, error)
	GetMapAboutName2ID(ctx context.Context, connNames []string) (map[string]string, error)
	GetMapAboutID2Name(ctx context.Context, connIDs []string) (map[string]string, error)
	GetDataConnectionsByConfigMD5(ctx context.Context, configMD5 string) (connMap map[string]*DataConnection, err error)
	GetDataConnectionSourceType(ctx context.Context, connID string) (string, bool, error)

	CreateDataConnectionStatus(ctx context.Context, tx *sql.Tx, status DataConnectionStatus) error
	DeleteDataConnectionStatuses(ctx context.Context, tx *sql.Tx, connIDs []string) error
	UpdateDataConnectionStatus(ctx context.Context, tx *sql.Tx, status DataConnectionStatus) error
}
