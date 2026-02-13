// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

// 事件模型列表查询参数
// type EventModelsQueryParams struct {
// 	ID          string `json:"id"`
// 	NamePattern string
// 	Name        string
// 	Type        string
// 	Tags        []string
// 	IsActive    int `default:"-1"`
// 	IsCustom    int `default:"-1"`
// }

const (
	//模块类型
	MODULE_TYPE_EVENT_MODEL = "event_model"

	// 事件模型类型
	EVENT_MODEL_TYPE_ATOMIC   = "atomic"
	EVENT_MODEL_TYPE_AGGR     = "aggregate"
	EVENT_MODEL_TYPE_ANALYSIS = "analysis"
)

var (
	DEFAULT_AGGREGATE_TYPE_FOR_ROOT_CAUSE_ANALYSIS = "1004"
	EVENT_MODEL_DATA_SOURCE_TYPE_FOR_METRIC_MODEL  = "metric_model"
	EVENT_MODEL_DATA_SOURCE_TYPE_FOR_EVENT_MODEL   = "event_model"
	EVENT_MODEL_DATA_SOURCE_TYPE_FOR_DATE_VIEW     = "data_view"
	DETECT_RULE_TYPE_FOR_RANGE_DETECT              = "range_detect"
	DETECT_RULE_TYPE_FOR_STATUS_DETECT             = "status_detect"
	DEFAULT_INDEX_BASE                             = "dip_event_model_data"
	DEFAULT_DATA_VIEW_NAME                         = "dip_event_model_data"
	DEFAULT_DATA_VIEW_ID                           = "__dip_event_model_data"
)

var (
	OBJECTTYPE_EVENT_MODEL = "ID_AUDIT_EVENT_MODEL"
	EXECUTE                = "ID_AUDIT_ACTION_EXECUTE"

	BlockStrategy = map[string]string{
		"SERIAL_EXECUTION": "SERIAL_EXECUTION",
		"DISCARD_LATER":    "DISCARD_LATER",
		"COVER_EARLY":      "COVER_EARLY",
	}
	RouteStrategy = map[string]string{
		"FIRST":                 "FIRST",
		"LAST":                  "LAST",
		"ROUND":                 "ROUND",
		"RANDOM":                "RANDOM",
		"CONSISTENT_HASH":       "CONSISTENT_HASH",
		"LEAST_FREQUENTLY_USED": "LEAST_FREQUENTLY_USED",
		"LEAST_RECENTLY_USED":   "LEAST_RECENTLY_USED",
		"FAILOVER":              "FAILOVER",
		"BUSYOVER":              "BUSYOVER",
		"SHARDING_BROADCAST":    "SHARDING_BROADCAST",
	}
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
	DetectRuleID string            `json:"id,omitempty"`
	Priority     int               `json:"priority"`
	Type         string            `json:"type" binding:"omitempty,required,oneof=range_detect status_detect agi_detect"`
	Formula      []FormulaItem     `json:"formula" binding:"dive,omitempty"`
	CreateTime   int64             `json:"create_time,omitempty" binding:"omitempty"`
	UpdateTime   int64             `json:"update_time,omitempty" binding:"omitempty"`
	DetectAlgo   string            `json:"detect_algo" binding:"omitempty,required_if=Type agi_detect"`
	AnalysisAlgo map[string]string `json:"analysis_algo" binding:"omitempty"`
}

type AggregateRule struct {
	AggregateRuleID string            `json:"id,omitempty"`
	Priority        int               `json:"priority"`
	Type            string            `json:"type" binding:"required_with=AggregateAlgo,omitempty"`
	AggregateAlgo   string            `json:"aggregate_algo" binding:"required_with=Type,omitempty"`
	AnalysisAlgo    map[string]string `json:"analysis_algo" binding:"omitempty"`
	GroupFields     []string          `json:"group_fields" binding:"required_if=Type group_aggregation"`
	CreateTime      int64             `json:"create_time,omitempty" binding:"omitempty"`
	UpdateTime      int64             `json:"update_time,omitempty" binding:"omitempty"`
}

type TimeInterval struct {
	Interval int    `json:"interval" binding:"number"`
	Unit     string `json:"unit" binding:"omitempty,oneof=d h m"`
}

// type DataSource struct {
// 	DataSourceType string   `json:"data_source_type" binding:"required,oneof=metric_model data_view event_model"`
// 	DataSource     []string `json:"data_source" binding:"required,max=100,min=1,dive,max=40"`
// }

type EventModelRequest struct {
	EventModelName           string   `json:"name"  binding:"required,min=1,max=40"`
	EventModelType           string   `json:"type"  binding:"required,oneof=atomic aggregate"`
	EventModelTags           []string `json:"tags" binding:"omitempty,max=5,dive,max=40"`
	DataSourceType           string   `json:"data_source_type" binding:"required,oneof=metric_model data_view event_model"`
	DataSource               []string `json:"data_source" binding:"required_if=EventModelType atomic,max=100,dive,max=40"`
	DataSourceName           []string `json:"data_source_name,omitempty"`
	DataSourceGroupName      []string `json:"data_source_group_name,omitempty"`
	DetectRule               `json:"detect_rule" binding:"required_if=EventModelType atomic"`
	AggregateRule            `json:"aggregate_rule" binding:"required_if=EventModelType aggregate"`
	EventModelComment        string       `json:"comment" binding:"omitempty,max=255"`
	DefaultTimeWindow        TimeInterval `json:"default_time_window"`
	EventTaskRequest         `json:"persist_task_config"`
	DownstreamDependentModel []string `json:"downstream_dependent_model" binding:"omitempty"`
}

type EventTaskRequest struct {
	Schedule                Schedule       `json:"schedule"`
	ExecuteParameter        map[string]any `json:"execute_parameter"`
	StorageConfig           StorageConfig  `json:"storage_config"`
	DispatchConfig          DispatchConfig `json:"dispatch_config"`
	DownstreamDependentTask []string       `json:"downstream_dependent_task"`
}

type EventModelCreateRequest struct {
	EventModelRequest
	IsActive        int `json:"is_active" binding:"omitempty,number,oneof=0 1"`
	IsCustom        int `json:"is_custom" binding:"omitempty,number"`
	Status          int `json:"status" binding:"omitempty,number,oneof=0 1"`
	EnableSubscribe int `json:"enable_subscribe" binding:"omitempty,number,oneof=0 1"`
}

type EventModelUpateRequest struct {
	EventModelRequest
	EventModelID    string `json:"id" uri:"event_model_id" binding:"omitempty,required"`
	IsActive        int    `json:"is_active" form:"is_active" binding:"number,oneof=0 1"` // 定时任务开关
	EnableSubscribe int    `json:"enable_subscribe" form:"enable_subscribe" binding:"number,oneof=0 1"`
	Status          int    `json:"status" form:"status" binding:"number,oneof=0 1"` // 模型是否启用
}

type EventModelItemQueryRequest struct {
	EventModelIDs string `json:"event_model_ids" uri:"event_model_ids"`
}

type EventModelDeleteRequest struct {
	EventModelIDs string `json:"event_model_ids" uri:"event_model_ids"`
}

type EventModelQueryRequest struct {
	EventModelName        string `json:"name" form:"name" binding:"omitempty,max=40,excluded_unless=EventModelNamePattern '' " `
	EventModelNamePattern string `json:"name_pattern" form:"name_pattern" binding:"omitempty,max=40,excluded_unless=EventModelName ''"`
	EventModelType        string `json:"type" form:"type" binding:"omitempty"`
	EventModelTag         string `json:"tag" form:"tag" binding:"omitempty"`
	DataSourceType        string `json:"data_source_type" form:"data_source_type"`
	DataSource            string `json:"data_source" form:"data_source"`
	SortKey               string `json:"sort" form:"sort" binding:"omitempty,oneof=name update_time"`
	Direction             string `json:"direction" form:"direction" binding:"omitempty,oneof=asc desc"`
	Offset                int    `json:"offset" form:"offset" binding:"omitempty,number,min=0"`
	Limit                 int    `json:"limit" form:"limit" binding:"omitempty,number,max=50"`
	IsActive              string `json:"is_active" form:"is_active" `
	IsCustom              int    `json:"is_custom" form:"is_custom" binding:"omitempty,number"`
	EnableSubscribe       string `json:"enable_subscribe" form:"enable_subscribe"`
	Status                string `json:"status" form:"status"`
	ScheduleSyncStatus    string `json:"schedule_sync_status" form:"schedule_sync_status"`
	TaskStatus            string `json:"task_status" form:"task_status"`
	DetectType            string `json:"detect_type" form:"detect_type" binding:"omitempty"`
	AggregateType         string `json:"aggregate_type" form:"aggregate_type" binding:"omitempty"`
}

// 事件模型结构体
// 事件模型结构体
type EventModel struct {
	EventModelID             string        `json:"id"`
	EventModelName           string        `json:"name" binding:"min=1,max=40"`
	EventModelType           string        `json:"type" binding:"omitempty,oneof=atomic aggregate"`
	Creator                  AccountInfo   `json:"creator" binding:"omitempty"`
	CreateTime               int64         `json:"create_time" binding:"omitempty"`
	UpdateTime               int64         `json:"update_time" binding:"omitempty"`
	EventModelTags           []string      `json:"tags" binding:"max=5,dive,string"`
	DataSourceType           string        `json:"data_source_type" binding:"oneof=metric_model data_view event_model"`
	DataSource               []string      `json:"data_source" binding:"required_if=EventModelType atomic,max=100,dive,max=40"`
	DataSourceName           []string      `json:"data_source_name,omitempty"`
	DataSourceGroupName      []string      `json:"data_source_group_name,omitempty"`
	DetectRule               DetectRule    `json:"detect_rule" binding:"required_if=EventModelType atomic"`
	AggregateRule            AggregateRule `json:"aggregate_rule" binding:"required_if=EventModelType aggregate"`
	EventModelComment        string        `json:"comment"`
	DefaultTimeWindow        TimeInterval  `json:"default_time_window"`
	Task                     EventTask     `json:"persist_task_config"`
	IsActive                 int           `json:"is_active" binding:"oneof=0 1"` // 定时任务开关
	IsCustom                 int           `json:"is_custom" binding:"oneof=0 1"`
	EnableSubscribe          int           `json:"enable_subscribe" binding:"oneof=0 1"` // 是否订阅
	DownstreamDependentModel []string      `json:"downstream_dependent_model" binding:"omitempty"`
	Status                   int           `json:"status" binding:"oneof=0 1"` // 模型是否启用

	// 操作权限
	Operations []string `json:"operations"`
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
	Creator                 AccountInfo    `json:"creator" binding:"omitempty"`
	CreateTime              int64          `json:"create_time" binding:"omitempty"`
	UpdateTime              int64          `json:"update_time" binding:"omitempty"`
	DownstreamDependentTask []string       `json:"downstream_dependent_task" binding:"omitempty"`
}

type DispatchConfig struct {
	TimeOut        int    `json:"timeout"`
	RouteStrategy  string `json:"route_strategy"`
	BlockStrategy  string `json:"block_strategy"`
	FailRetryCount int    `json:"fail_retry_count"`
}

type StorageConfig struct {
	IndexBase    string `json:"index_base"`
	DataViewID   string `json:"data_view_id"`
	DataViewName string `json:"data_view_name"`
}

// type EventTaskSyncStatus struct {
// 	SyncStatus int
// 	UpdateTime string
// 	TaskID     string
// 	ModelID    string
// }

//go:generate mockgen -source ../interfaces/event_model_service.go -destination ../interfaces/mock/mock_event_model_service.go
type EventModelService interface {
	CreateEventModels(ctx context.Context, eventModels []EventModel) ([]map[string]any, error)
	UpdateEventModel(ctx context.Context, eventModel EventModelUpateRequest) error
	DeleteEventModels(ctx context.Context, modelIDs []string) ([]EventModel, error)
	QueryEventModels(ctx context.Context, params EventModelQueryRequest) ([]EventModel, int, error)
	GetEventModelByID(ctx context.Context, modelID string) (EventModel, *rest.HTTPError)
	CheckEventModelExistByName(ctx context.Context, modelName string) (bool, error)
	EventModelCreateValidate(ctx context.Context, eventModel EventModel) (EventModel, *rest.HTTPError)
	EventModelUpdateValidate(ctx context.Context, eventModelRequest EventModelUpateRequest) *rest.HTTPError
	BatchValidateDataSources(ctx context.Context, dataSource []string, event_model_type string, dataSourceType string, DataSourceName []string, DataSourceGroupName []string) ([]string, *rest.HTTPError)
	ValidateEventModelDataSource(ctx context.Context, dataSource string, dataSourceType string, DataSourceName string, DataSourceGroupName string) (string, error)
	ValidateEventModelDetectRule(ctx context.Context, detectRule DetectRule, detectRuleType string) *rest.HTTPError
	ValidateEventModelAggregateRule(ctx context.Context, aggregateRule AggregateRule, aggregateRuleType string) bool
	// ValidateEventModel(ctx context.Context, eventModel EventModel) error
	GetEventModelRefs(ctx context.Context, modelID string) (int, *rest.HTTPError)
	GetEventModelMapByNames(modelNames []string) (map[string]string, error)
	GetEventModelMapByIDs(modelIDs []string) (map[string]string, error)

	// SyncSchedule()
	//NOTE：event model task
	CreateEventTask(ctx context.Context, tx *sql.Tx, tasks EventTask) error
	UpdateEventTask(ctx context.Context, tx *sql.Tx, task EventTask) error
	// SetTaskSyncStatusByTaskID(ctx context.Context, tx *sql.Tx, taskSyncStatus EventTaskSyncStatus) error
	// SetTaskSyncStatusByModelID(ctx context.Context, tx *sql.Tx, taskSyncStatus EventTaskSyncStatus) error

	GetEventTaskIDByModelIDs(ctx context.Context, modelID []string) ([]string, error)
	GetEventTaskByTaskID(ctx context.Context, taskID string) (EventTask, error)
	GetEventTaskByModelID(ctx context.Context, modelID string) (EventTask, bool, error)
	DeleteEventTaskByTaskIDs(ctx context.Context, tx *sql.Tx, taskID []string) error
	// GetProcessingEventTasks(ctx context.Context) ([]EventTask, error)
	UpdateEventTaskStatusInFinish(ctx context.Context, task EventTask) error
	ValidateExecuteParam(ctx context.Context, executeParam map[string]any) (bool, error)
	UpdateEventTaskAttributes(ctx context.Context, task EventTask) error

	ListEventModelSrcs(ctx context.Context, params EventModelQueryRequest) ([]Resource, int, error)
}
