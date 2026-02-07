// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
)

type DataSource struct {
	ID      string  `json:"id"`       // 数据源业务id
	Name    string  `json:"name"`     // 数据源名称
	Type    string  `json:"type"`     // 数据库类型名称
	BinData BinData `json:"bin_data"` // 数据源配置信息
	Comment string  `json:"comment"`  // 描述

	// CreatedByUID string    `json:"created_by_uid"`                   // 创建人id
	// CreatedAt    time.Time `json:"created_at"` // 创建时间
	// UpdatedByUID string    `json:"updated_by_uid"`                   // 更新人id
	// UpdatedAt    time.Time `json:"updated_at"` // 更新时间
}

type BinData struct {
	CatalogName     string `json:"catalog_name"`
	DataBaseName    string `json:"database_name"`
	Schema          string `json:"schema"`
	ConnectProtocol string `json:"connect_protocol"`
	Host            string `json:"host"`
	Port            int    `json:"port"`
	Account         string `json:"account"`
	Password        string `json:"password"`
	Token           string `json:"token"`
	StorageProtocol string `json:"storage_protocol"`
	StorageBase     string `json:"storage_base"`
	ReplicaSet      string `json:"replica_set"`
}

//go:generate mockgen -source ../interfaces/data_connection_access.go -destination ../interfaces/mock/mock_data_connection_access.go
type DataConnectionAccess interface {
	GetDataSourceById(ctx context.Context, dataSourceId string) (*DataSource, error)
}
