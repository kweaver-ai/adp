// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

type contextKey string // 自定义专属的key类型

const (
	CONTENT_TYPE_NAME = "Content-Type"
	CONTENT_TYPE_JSON = "application/json"

	HTTP_HEADER_METHOD_OVERRIDE = "x-http-method-override"
	HTTP_HEADER_ACCOUNT_ID      = "x-account-id"
	HTTP_HEADER_ACCOUNT_TYPE    = "x-account-type"

	ACCOUNT_INFO_KEY contextKey = "x-account-info" // 避免直接使用string

	ADMIN_ID   string = "266c6a42-6131-4d62-8f39-853e7093701c"
	ADMIN_TYPE        = string(rest.VisitorType_User)

	// 模块类型
	MODULE_TYPE_METRIC_MODEL    = "metric_model"
	MODULE_TYPE_OBJECTIVE_MODEL = "objective_model"
	MODULE_TYPE_EVENT_MODEL     = "event_model"

	// 任务类型
	// 提交的时候提交job_type字段，扫描的时候是metric单独扫描，在扫描metric的时候，构造的jobInfo的job_type赋值为 metric_mdoel
	JOB_TYPE_STREAM   = "stream"   // 流式订阅的任务类型为 stream
	JOB_TYPE_SCHEDULE = "schedule" // 定时任务的类型为 schedule

	// 调度类型
	SCHEDULE_TYPE_FIXED = "FIX_RATE"
	SCHEDULE_TYPE_CRON  = "CRON"

	// 事件模型,指标模型持久化管道的topic
	MODEL_PERSIST_INPUT = "%s.sdp.mdl-model-persistence.input"
)

type AccountInfo struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}
