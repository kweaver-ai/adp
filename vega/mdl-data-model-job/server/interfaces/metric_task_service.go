// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

const (
	APP_NAME        = "metric-persist-jobs"
	ExecutorHandler = "metricTaskhandler"

	// 前一个时间单位的窗口
	PREVIOUS_HOUR  string = "previous_hour"
	PREVIOUS_DAY   string = "previous_day"
	PREVIOUS_WEEK  string = "previous_week"
	PREVIOUS_MONTH string = "previous_month"

	// 任务同步状态
	SCHEDULE_SYNC_STATUS_SUCCESS = 4 // 执行成功
	SCHEDULE_SYNC_STATUS_FAILED  = 5 // 执行失败

	SCHEDULE_RUNNING_STATUS_SUCCESS = 4 // 执行成功
	SCHEDULE_RUNNING_STATUS_FAILED  = 5 // 执行失败

	// 最小固定步长，5min
	MIN_STEP = 300000

	DEFAULT_TIME_ZONE = "Asia/Shanghai"
)

type MetricTask struct {
	TaskID             string      `json:"id"`
	TaskName           string      `json:"name"`
	ModuleType         string      `json:"module_type"`
	ModelID            string      `json:"model_id"`
	MeasureName        string      `json:"measure_name"`
	Schedule           Schedule    `json:"schedule"`
	TimeWindows        []string    `json:"time_windows"`
	Steps              []string    `json:"steps"`
	IndexBase          string      `json:"index_base"`
	IndexBaseName      string      `json:"index_base_name"`
	RetraceDuration    string      `json:"retrace_duration"`
	Comment            string      `json:"comment"`
	ScheduleSyncStatus int         `json:"schedule_sync_status"`
	ExeccuteStatus     int         `json:"execute_status"`
	UpdateTime         int64       `json:"update_time"`
	PlanTime           int64       `json:"plan_time"`
	Creator            AccountInfo `json:"creator"`
}

type Schedule struct {
	Type       string `json:"type"`
	Expression string `json:"expression"`
}

type Filter struct {
	Name      string      `json:"name"`
	Operation string      `json:"operation"`
	Value     interface{} `json:"value"`
}

//go:generate mockgen -source ../interfaces/metric_task_service.go -destination ../interfaces/mock/mock_metric_task_service.go
type MetricTaskService interface {
	// 参数为jobconfig，其内是任务信息
	MetricTaskExecutor(cxt context.Context, metricTask MetricTask) (msg string)
}
