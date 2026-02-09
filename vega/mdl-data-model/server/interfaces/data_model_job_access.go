// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"database/sql"
)

// 数据库的job信息
type JobInfo struct {
	JobID            string         `json:"id"`
	JobType          string         `json:"job_type"`
	JobConfig        map[string]any `json:"job_config"`
	JobStatus        string         `json:"job_status"`
	JobStatusDetails string         `json:"job_status_details"`
	CreateTime       int64          `json:"create_time"`
	UpdateTime       int64          `json:"update_time"`
	Creator          AccountInfo    `json:"creator"`
}

// 视图实时订阅任务的配置
type ViewCfg struct {
	ViewID     string         `json:"id"`
	DataSource map[string]any `json:"data_source"`
	FieldScope uint8          `json:"field_scope"`
	Fields     []*ViewField   `json:"fields"`
	Condition  *CondCfg       `json:"filters"`
	Creator    AccountInfo    `json:"creator"`
}

//go:generate mockgen -source ../interfaces/data_model_job_access.go -destination ../interfaces/mock/mock_data_model_job_access.go
type DataModelJobAccess interface {
	// 数据库操作
	CreateDataModelJob(ctx context.Context, tx *sql.Tx, job *JobInfo) error
	DeleteDataModelJobs(ctx context.Context, tx *sql.Tx, jobIDs []string) error

	// 请求外部服务
	StartJob(ctx context.Context, job *DataModelJobCfg) error
	StopJobs(ctx context.Context, jobIDs []string) error
	UpdateJob(ctx context.Context, job *DataModelJobCfg) error
}
