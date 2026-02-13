// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
)

const (
	EVENT_APP_NAME          = "event-persist-jobs"
	EVENT_EXECUTOR_HANDLER  = "eventTaskhandler"
	GENERATE_TYPE_FOR_BATCH = "batch"
)

type FilterExpress struct {
	Name      string `json:"name" binding:"required_with=Value,omitempty"`
	Value     any    `json:"value" binding:"required_with=Name,omitempty"`
	Operation string `json:"operation" binding:"required_with=Name,omitempty"`
}

type LogicFilter struct {
	LogicOperator string        `json:"logic_operator" binding:"omitempty"`
	FilterExpress FilterExpress `json:"filter_express"`
	Children      []LogicFilter `json:"children" binding:"required_with=LogicOperator,omitempty"`
}

type FormulaItem struct {
	Level  int         `json:"level" binding:"required_with=Filter,omitempty,oneof=1 2 3 4 5 6"`
	Filter LogicFilter `json:"filter" binging:"required_with=Level,omitempty"`
}
type DetectRule struct {
	DetectRuleId string        `json:"id,omitempty"`
	Priority     int           `json:"priority"`
	Type         string        `json:"type" binding:"required_with=Formula,omitempty,oneof=range_detect status_detect"`
	Formula      []FormulaItem `json:"formula" binding:"required_with=Type,dive,required,omitempty"`
	UpdateTime   int64         `json:"update_time,omitempty" binding:"omitempty"`
}

type AggregateRule struct {
	AggregateRuleId string   `json:"id,omitempty"`
	Priority        int      `json:"priority"`
	Type            string   `json:"type" binding:"required_with=AggregateAlgo,omitempty"`
	AggregateAlgo   string   `json:"aggregate_algo" binding:"required_with=Type,omitempty"`
	GroupFields     []string `json:"group_fields" binding:"required_if=Type group_aggregation"`
	UpdateTime      int64    `json:"update_time,omitempty" binding:"omitempty"`
}

type EventModel struct {
	EventModelID        string        `json:"id"`
	EventModelName      string        `json:"name" binding:"min=1,max=40"`
	EventModelType      string        `json:"type" binding:"omitempty,oneof=atomic aggregate"`
	CreateTime          int64         `json:"create_time" binding:"omitempty"`
	UpdateTime          int64         `json:"update_time" binding:"omitempty"`
	EventModelTags      []string      `json:"tags" binding:"max=5,dive,string"`
	DataSourceType      string        `json:"data_source_type" binding:"oneof=metric_model data_view event_model"`
	DataSource          []string      `json:"data_source" binding:"required,max=100,min=1,dive,max=40"`
	DataSourceName      []string      `json:"data_source_name,omitempty"`
	DataSourceGroupName []string      `json:"data_source_group_name,omitempty"`
	DetectRule          DetectRule    `json:"detect_rule" binding:"required_if=EventModelType atomic"`
	AggregateRule       AggregateRule `json:"aggregate_rule" binding:"required_if=EventModelType aggregate"`
	EventModelComment   string        `json:"comment"`
	DefaultTimeWindow   TimeInterval  `json:"default_time_window" binding:"required"`
	Task                EventTask     `json:"persist_task_config"`
	IsActive            int           `json:"is_active" binding:"oneof=0 1"`
	IsCustom            int           `json:"is_custom" binding:"oneof=0 1"`
	EnableSubscribe     int           `json:"enable_subscribe" binding:"oneof=0 1"`
	Status              int           `json:"status" binding:"oneof=0 1"`
}
type EventTask struct {
	TaskID                  string         `json:"id"`
	ModelID                 string         `json:"model_id"`
	Schedule                Schedule       `json:"schedule"`
	StorageConfig           StorageConfig  `json:"storage_config"`
	DispatchConfig          DispatchConfig `json:"dispatch_config"`
	ExecuteParameter        map[string]any `json:"execute_parameter"`
	ScheduleSyncStatus      int            `json:"schedule_sync_status"`
	StatusUpdateTime        int64          `json:"status_update_time" binding:"omitempty"`
	TaskStatus              int            `json:"task_status"`
	ErrorDetails            string         `json:"error_details"`
	UpdateTime              int64          `json:"update_time" binding:"omitempty"`
	DownstreamDependentTask []string       `json:"downstream_dependent_task" binding:"omitempty"`
	Creator                 AccountInfo    `json:"creator"`
}
type DispatchConfig struct {
	TimeOut        int    `json:"timeout"`
	RouteStrategy  string `json:"route_strategy"`
	BlockStrategy  string `json:"block_strategy"`
	FailRetryCount int    `json:"fail_retry_count"`
}

type StorageConfig struct {
	IndexBase    string `json:"index_base"`
	DataViewId   string `json:"data_view_id"`
	DataViewName string `json:"data_view_name"`
}

//go:generate mockgen -source ../interfaces/event_task_service.go -destination ../interfaces/mock/mock_event_task_service.go
type EventTaskService interface {
	EventTaskExecutor(cxt context.Context, eventTask EventTask) (msg string)
}
