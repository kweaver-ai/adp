// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

type JobInfo struct {
	JobId            string         `json:"job_id"`      // data-view对应的是job_id字段，指标、事件、目标对应的是task_id字段
	JobType          string         `json:"job_type"`    // 任务类型：流式任务(streaming)、定时任务（scheduled）
	ModuleType       string         `json:"module_type"` // 模块类型：data_view, metric_model, event_model, objective_model
	JobConfig        map[string]any `json:"job_config"`
	JobStatus        string         `json:"job_status"`
	JobStatusDetails string         `json:"job_status_details"`
	CreateTime       int64          `json:"create_time"`
	UpdateTime       int64          `json:"update_time"`
	Creator          AccountInfo    `json:"creator"`

	DataView   `json:"data_view,omitempty"`
	MetricTask *MetricTask `josn:"metric_task,omitempty"` // 指标模型、目标模型的持久化任务信息
	EventTask  *EventTask  `josn:"event_task,omitempty"`  // 事件模型的持久化任务信息
	Schedule   `json:"schedule,omitempty"`

	Ticker   *time.Ticker  // 固定频率的计时器
	StopChan chan struct{} `json:"-"` // 用于停止固定频率任务
	CronID   cron.EntryID  // Cron 任务的 ID
}

// job 信息打印
func (j *JobInfo) String() string {
	return fmt.Sprintf("{job_id = %s, job_type = %s, job_status = %s, job_status_details = %s, job_view_id = %s}",
		j.JobId, j.JobType, j.JobStatus, j.JobStatusDetails, j.ViewId)
}

//go:generate mockgen -source ../interfaces/job_access.go -destination ../interfaces/mock/mock_job_access.go
type JobAccess interface {
	ListViewJobs() ([]JobInfo, error)
	UpdateJobStatus(job JobInfo) error

	ListMetricJobs() ([]JobInfo, error)
	ListObjectiveJobs() ([]JobInfo, error)
	ListEventJobs() ([]JobInfo, error)
}
