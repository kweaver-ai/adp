// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"fmt"
	"time"

	cond "data-model-job/common/condition"
)

const (
	DATA_VIEW  = "data_view"
	Index_Base = "index_base"

	// tenant.mdl.process.base_type
	Process_Topic = "%s.mdl.process.%s"
	// tenant.mdl.view.view_id
	Sink_Topic = "%s.mdl.view.%s"
)

// 字段范围
const (
	CUSTOM uint8 = iota
	ALL
)

// 任务状态
const (
	JobStatus_Running = "running"
	JobStatus_Error   = "error"
)

const (
	TaskStatus_Error    = "error"
	TaskStatus_Running  = "running"
	TaskStatus_Stopping = "stopping"
	TaskStatus_Stopped  = "stopped"
)

const (
	DataType_Keyword = "keyword"
	DataType_Text    = "text"
	DataType_Binary  = "binary"

	DataType_Short     = "short"
	DataType_Integer   = "integer"
	DataType_Long      = "long"
	DataType_Float     = "float"
	DataType_Double    = "double"
	DataType_HalfFloat = "half_float"
	DataType_Byte      = "byte"

	DataType_Boolean = "boolean"

	DataType_Date = "date"

	DataType_Ip       = "ip"
	DataType_GeoPoint = "geo_point"
	DataType_GeoShape = "geo_shape"
)

type ViewJobCfg struct {
	DataSource map[string]any
	FieldScope uint8
	Fields     []*cond.Field
	Condition  *cond.CondCfg
	JobStatus  string
}

func (vjCfg *ViewJobCfg) String() string {
	return fmt.Sprintf("{data_source = %v, field_scope = %d, fields = %v, condition = %v, job_status = %s}",
		vjCfg.DataSource, vjCfg.FieldScope, vjCfg.Fields, vjCfg.Condition, vjCfg.JobStatus,
	)
}

//go:generate mockgen -source ../interfaces/job_service.go -destination ../interfaces/mock/mock_job_service.go
type JobService interface {
	StartJob(ctx context.Context, jobInfo *JobInfo) error
	UpdateJob(ctx context.Context, jobInfo *JobInfo) error
	StopJob(ctx context.Context, jobId string) error
	WatchJob(interval time.Duration)
	ListenToErrChan()
}
