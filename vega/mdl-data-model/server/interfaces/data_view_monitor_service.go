// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"time"
)

const (
	SyncType_Full        = "full"
	SyncType_Incremental = "incremental"
)

// SyncResult 同步结果
type SyncResult struct {
	ViewName  string    `json:"view_name"`
	Action    string    `json:"action"` // "created", "updated", "skipped", "error"
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// BatchResult 批次处理结果
type BatchResult struct {
	BatchID               int          `json:"batch_id"`                 // 批次ID
	TotalMetaTableCount   int          `json:"total_meta_table_count"`   // 总元数据表数量
	InvalidMetaTableCount int          `json:"invalid_meta_table_count"` // 无效元数据表数量
	TotalCount            int          `json:"total_count"`              // 总处理数量
	SuccessCount          int          `json:"success_count"`            // 成功处理数量
	NeedCreatedCount      int          `json:"need_created_count"`       // 待创建视图数量
	NeedUpdatedCount      int          `json:"need_updated_count"`       // 待更新视图数量
	ErrorCount            int          `json:"error_count"`              // 错误数量
	StartTime             time.Time    `json:"start_time"`               // 开始时间
	EndTime               time.Time    `json:"end_time"`                 // 结束时间
	Results               []SyncResult `json:"results"`                  // 同步结果列表
}

//go:generate mockgen -source ../interfaces/data_view_monitor_service.go -destination ../interfaces/mock/mock_data_view_monitor_service.go
type DataViewMonitorService interface {
	// PollingMetadata 轮询元数据管理接口
	PollingMetadata(ctx context.Context)
}
