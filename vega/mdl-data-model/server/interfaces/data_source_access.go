// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
)

type DataSource struct {
	ID           string  `json:"id"`                   // 数据源业务id
	Name         string  `json:"name"`                 // 数据源名称
	Type         string  `json:"type"`                 // 数据库类型名称
	BinData      BinData `json:"bin_data"`             // 数据源配置信息
	Comment      string  `json:"comment"`              // 描述
	LastScanTime int64   `json:"last_scan_time"`       // 上一次扫描时间
	Status       string  `json:"status"`               // 数据源状态：扫描中、可用
	CreatorID    string  `json:"created_by_uid"`       // 创建人id
	CreatorType  string  `json:"created_by_user_type"` // 创建人类型
	CreateTime   int64   `json:"created_at"`           // 创建时间
	UpdaterID    string  `json:"updated_by_uid"`       // 更新人id
	UpdaterType  string  `json:"updated_by_user_type"` // 更新人类型
	UpdateTime   int64   `json:"updated_at"`           // 更新时间

}

type BinData struct {
	CatalogName     string `json:"catalog_name"`
	DataBaseName    string `json:"database_name"`
	ConnectProtocol string `json:"connect_protocol"`
	Schema          string `json:"schema"`
	Host            string `json:"host"`
	Port            int    `json:"port"`
	Account         string `json:"account"`
	Password        string `json:"password"`
	Token           string `json:"token"`
	StorageProtocol string `json:"storage_protocol"`
	StorageBase     string `json:"storage_base"`
	ReplicaSet      string `json:"replica_set"`
	// DataViewSource  string `json:"data_view_source"`
}

type DataSourceStatus struct {
	Status int32 `gorm:"column:status" json:"status"`
}

type ScanRecord struct {
	RecordID         string `json:"id"`                 // 扫描记录id
	DataSourceID     string `json:"data_source_id"`     // 数据源id
	Scanner          string `json:"scanner"`            // 扫描者
	ScanTime         int64  `json:"scan_time"`          // 扫描时间
	DataSourceStatus string `json:"data_source_status"` // 数据源状态
	MetadataTaskID   string `json:"metadata_task_id"`
}

type ListDataSourcesResult struct {
	Entries    []*DataSource `json:"entries"`
	TotalCount int           `json:"total_count"`
}

//go:generate mockgen -source ../interfaces/data_source_access.go -destination ../interfaces/mock/mock_data_source_access.go
type DataSourceAccess interface {
	GetDataSourceByID(ctx context.Context, dataSourceID string) (*DataSource, error)
	ListDataSources(ctx context.Context) (*ListDataSourcesResult, error)
}
