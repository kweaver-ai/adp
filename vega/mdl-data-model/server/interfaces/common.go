// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"math"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

type contextKey string // 自定义专属的key类型

const (
	CONTENT_TYPE_NAME = "Content-Type"
	CONTENT_TYPE_JSON = "application/json"
	CONTENT_TYPE_FORM = "application/x-www-form-urlencoded"

	HTTP_HEADER_METHOD_OVERRIDE = "x-http-method-override"
	HTTP_HEADER_REQUEST_TOOK    = "x-request-took"
	HTTP_HEADER_ACCOUNT_ID      = "x-account-id"
	HTTP_HEADER_ACCOUNT_TYPE    = "x-account-type"

	ACCOUNT_INFO_KEY contextKey = "x-account-info" // 避免直接使用string

	ADMIN_ID   string = "266c6a42-6131-4d62-8f39-853e7093701c"
	ADMIN_TYPE        = string(rest.VisitorType_User)

	OBJECT_NAME_MAX_LENGTH = 40
	DEFAULT_NAME_PATTERN   = ""
	DEFAULT_OFFEST         = "0"
	DEFAULT_LIMIT          = "10" // LIMIT=-1, 不分页
	DEFAULT_SORT           = "update_time"
	DEFAULT_DIRECTION      = "desc"
	DESC_DIRECTION         = "desc"
	ASC_DIRECTION          = "asc"
	MIN_OFFSET             = 0
	MIN_LIMIT              = 1
	MAX_LIMIT              = 1000
	NO_LIMIT               = "-1"
	DEFAULT_SIMPLE_INFO    = "false"
	COMMENT_MAX_LENGTH     = 255
	NAME_INVALID_CHARACTER = "/:?\\\"<>|：？‘’“”！《》,#[]{}%&*$^!=.'"

	TAGS_MAX_NUMBER = 5

	DEFAULT_FORCE    = "false"
	DEFAULT_GROUP_ID = ""

	DEFAULT_INCLUDE_VIEW = "false"

	// 对象的导入模式
	ImportMode_Normal    = "normal"
	ImportMode_Ignore    = "ignore"
	ImportMode_Overwrite = "overwrite"

	// 对象id的校验
	RegexPattern_Builtin_ID    = "^[a-z0-9_][a-z0-9_-]{0,39}$"
	RegexPattern_NonBuiltin_ID = "^[a-z0-9][a-z0-9_-]{0,39}$"

	// 任务类型
	// 提交的时候提交job_type字段，扫描的时候是metric单独扫描，在扫描metric的时候，构造的jobInfo的job_type赋值为 metric_mdoel
	JOB_TYPE_STREAM   = "stream"   // 流式订阅的任务类型为 stream
	JOB_TYPE_SCHEDULE = "schedule" // 定时任务的类型为 schedule

	// 未分组中英文
	UNGROUPED_ZH_CN = "未分组"
	UNGROUPED_EN_US = "Ungrouped"
)

var (
	// 正负无穷
	NEG_INF float64 = math.Inf(-1)
	POS_INF float64 = math.Inf(1)
)

// 分页查询参数
type PaginationQueryParameters struct {
	Offset    int
	Limit     int
	Sort      string
	Direction string
}

type CondCfg struct {
	Name        string     `json:"field,omitempty" mapstructure:"field"` // 接口传递的是name
	Operation   string     `json:"operation,omitempty" mapstructure:"operation"`
	SubConds    []*CondCfg `json:"sub_conditions,omitempty" mapstructure:"sub_conditions"`
	ValueOptCfg `mapstructure:",squash"`

	NameField *ViewField `json:"-" mapstructure:"-"`
}

type ValueOptCfg struct {
	ValueFrom string `json:"value_from,omitempty" mapstructure:"value_from"`
	Value     any    `json:"value,omitempty" mapstructure:"value"`
}

// data-model-job任务的配置
type DataModelJobCfg struct {
	JobID      string `json:"job_id"`
	JobType    string `json:"job_type"`
	ModuleType string `json:"module_type"`
	ViewCfg    `json:"data_view,omitempty"`
	MetricTask *MetricTask `josn:"metric_task,omitempty"` // 指标模型、目标模型的持久化任务信息
	EventTask  *EventTask  `josn:"event_task,omitempty"`  // 事件模型的持久化任务信息
	Schedule   `json:"schedule,omitempty"`
}

// type Field struct {
// 	Name     string `json:"name"`
// 	Type     string `json:"type"`
// 	Hidden   bool   `json:"hidden"`
// 	Comment  string `json:"comment"`
// 	Format   string `json:"format,omitempty"`
// 	Analyzer string `json:"analyzer,omitempty"`

// 	Path []string `json:"-"`
// }

type Filter struct {
	Name      string      `json:"name"`
	Value     interface{} `json:"value"`
	Operation string      `json:"operation"`
}

type AccountInfo struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}
