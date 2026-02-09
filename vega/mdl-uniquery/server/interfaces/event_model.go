// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"encoding/json"
)

const (
	EVENT_MODEL_LEVEL_NORMAL        = 0
	EVENT_MODEL_LEVEL_CRITICAL      = 1
	EVENT_MODEL_LEVEL_MAJOR         = 2
	EVENT_MODEL_LEVEL_MINOR         = 3
	EVENT_MODEL_LEVEL_WARNING       = 4
	EVENT_MODEL_LEVEL_INDETERMINATE = 5
	EVENT_MODEL_LEVEL_CLEARED       = 6

	GENERATE_TYPE_BATCH     = "batch"
	GENERATE_TYPE_STREAMING = "streaming"

	DEFAULT_SUBSCRIBE_TOPIC = "%s.mdl.view"
	ATOMIC_EVENT_DATA_TOPIC = "%s.mdl.atomic_event"
)

var (
	FieldNameMap = map[string]string{
		"level": "Level",
		"tags":  "Tags",
		"type":  "EventType",
		"id":    "Id",
	}

	EVENT_MODEL_LEVEL_ZH_CN = map[int]string{
		1: "紧急",
		2: "主要",
		3: "次要",
		4: "提示",
		5: "不明确",
		6: "清除",
	}

	EVENT_MODEL_LEVEL_EN_US = map[int]string{
		1: "Critical",
		2: "Major",
		3: "Minor",
		4: "Warning",
		5: "Indeterminate",
		6: "Cleared",
	}
)

// 事件查询的接口查询参数
type EventQueryReq struct {
	Querys    []EventQuery `json:"querys" binding:"required,dive,required"`
	SortKey   string       `json:"sort" form:"sort" binding:"omitempty"`
	Limit     int64        `json:"limit" form:"limit" binding:"omitempty"`
	Offset    int64        `json:"offset" form:"offset" binding:"omitempty,number,gte=0"`
	Direction string       `json:"direction" form:"direction" binding:"omitempty,oneof=asc desc"`
}

// 事件数据预览的查询请求体
type EventQuery struct {
	QueryType string `json:"query_type" form:"query_type" binding:"required,oneof=instant_query range_query class_query"`
	Id        string `json:"id,omitempty" form:"id"`
	Start     int64  `json:"start" form:"end" binding:"omitempty"`
	End       int64  `json:"end" form:"end" binding:"omitempty"`
	Step      string `json:"step" form:"step" binding:"omitempty"`
	// Time          int64    `json:"time" form:"time"`
	Filters    []Filter `json:"filters" form:"filters" binding:"omitempty"`
	Extraction []Filter `json:"extraction" form:"extraction" binding:"omitempty"`
	// LookBackDelta string   `json:"look_back_delta" form:"look_back_delta"`
	Preview   int    `json:"preview" form:"preview" binding:"number,oneof=1 0"`
	SortKey   string `json:"sort" form:"sort" binding:"omitempty"`
	Limit     int64  `json:"limit" form:"limit" binding:"omitempty,number,gte=-1"`
	Offset    int64  `json:"offset" form:"offset" binding:"omitempty,number,gte=0"`
	Direction string `json:"direction" form:"direction" binding:"omitempty,oneof=asc desc"`
	//NOTE 新建预览时额外传递的参数
	EventModelName      string        `json:"name" binding:"required_without=Id"`
	EventModelType      string        `json:"type" binding:"required_without=Id"`
	EventModelTags      []string      `json:"tags"`
	DataSourceType      string        `json:"data_source_type"`
	DataSource          []string      `json:"data_source"`
	DetectRule          DetectRule    `json:"detect_rule" binding:"required_if=EventModelType atomic"`
	AggregateRule       AggregateRule `json:"aggregate_rule" binding:"required_if=EventModelType aggregate"`
	Comment             string        `json:"comment"`
	DefaultTimeWindow   TimeInterval  `json:"default_time_window"`
	EnableMessageFilter bool          `json:"enable_message_filter" form:"enable_message_filter"`
}

type EventDetailsQueryReq struct {
	EventModelID string `uri:"event_model_id" binding:"required"`
	EventID      string `uri:"event_id" binding:"required"`
	Start        int64  `form:"start" binding:"number"`
	End          int64  `form:"end" binding:"number"`
}

type DetectRule struct {
	DetectRuleId string            `json:"id,omitempty"`
	Priority     int               `json:"priority,omitempty"`
	Type         string            `json:"type,omitempty" binding:"omitempty"`
	Formula      Formula           `json:"formula,omitempty" binding:"omitempty"`
	UpdateTime   int64             `json:"update_time,omitempty" binding:"omitempty"`
	DetectAlgo   string            `json:"detect_algo"`
	AnalysisAlgo map[string]string `json:"analysis_algo" binding:"omitempty"`
}

type FormulaItem struct {
	Level  int         `json:"level" binding:"required_with=Filter,omitempty,oneof=1 2 3 4 5 6"`
	Filter LogicFilter `json:"filter" binging:"required_with=Level,omitempty"`
}

type Formula []FormulaItem

func (f Formula) Len() int           { return len(f) }
func (f Formula) Less(i, j int) bool { return f[i].Level < f[j].Level }
func (f Formula) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }

type LogicFilter struct {
	LogicOperator string        `json:"logic_operator" binding:"omitempty"`
	FilterExpress FilterExpress `json:"filter_express"`
	Children      []LogicFilter `json:"children" binding:"required_with=LogicOperator,omitempty"`
}

type FilterExpress struct {
	Name      string `json:"name" binding:"required_with=Value,omitempty"`
	Value     any    `json:"value" binding:"required_with=Name,omitempty"`
	Operation string `json:"operation" binding:"required_with=Name,omitempty"`
}

type AggregateRule struct {
	AggregateRuleId string `json:"id,omitempty"`
	Priority        int    `json:"priority"`
	Type            string `json:"type"`
	AggregateAlgo   string `json:"aggregate_algo" `

	GroupFields  []string          `json:"group_fields"`
	UpdateTime   int64             `json:"update_time,omitempty" binding:"omitempty"`
	AnalysisAlgo map[string]string `json:"analysis_algo" binding:"omitempty"`
}

type TimeInterval struct {
	Interval int    `json:"interval" binding:"omitempty,number,required_with=Unit"`
	Unit     string `json:"unit" binding:"omitempty,oneof=d h m,required_with=Interval"`
}

type Records []map[string]any

type SourceRecords struct {
	Records Records `json:"source_records" form:"source_records"`
}

type Record map[string]any

// 事件模型结构体
type EventModel struct {
	EventModelID             string        `json:"id"`
	EventModelName           string        `json:"name" validate:"min=5,max=40"`
	EventModelType           string        `json:"type" validate:"oneof=AtomicEvent AggregatedEvent"`
	UpdateTime               int64         `json:"update_time"`
	EventModelTags           []string      `json:"tags" validate:"max<=5,dive,string"`
	DataSourceType           string        `json:"data_source_type" validate:"oneof=metric_model data_view event_model"`
	DataSource               []string      `json:"data_source"`
	DataSourceName           []string      `json:"data_source_name,omitempty"`
	DataSourceGroupName      []string      `json:"data_source_group_name,omitempty"`
	DetectRule               DetectRule    `json:"detect_rule"`
	AggregateRule            AggregateRule `json:"aggregate_rule"`
	Comment                  string        `json:"comment" validate:"string"`
	DefaultTimeWindow        TimeInterval  `json:"default_time_window"`
	Task                     EventTask     `json:"persist_task_config"`
	IsActive                 int           `json:"is_active" validate:"oneof=0 1"`
	IsCustom                 int           `json:"is_custom" validate:"oneof=0 1"`
	EnableSubscribe          int           `json:"enable_subscribe" validate:"oneof=0 1"`
	DownstreamDependentModel []string      `json:"downstream_dependent_model" binding:"omitempty"`
	Status                   int           `json:"status" validate:"oneof=0 1"`
}

// 事件模型持久化任务结构体
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
	DownstreamDependentTask []string       `json:"downstream_dependent_task" binding:"omitempty"`
	UpdateTime              int64          `json:"update_time" binding:"omitempty"`
}

type StorageConfig struct {
	IndexBase    string `json:"index_base"`
	DataViewId   string `json:"data_view_id"`
	DataViewName string `json:"data_view_name" binding:"omitempty"`
}

type DispatchConfig struct {
	TimeOut        int    `json:"timeout"`
	RouteStrategy  string `json:"route_strategy"`
	BlockStrategy  string `json:"block_strategy"`
	FailRetryCount int    `json:"fail_retry_count"`
}

type BaseEvent struct {
	Id                  string            `json:"id" form:"id"`
	Title               string            `json:"title" form:"title"`
	EventModelId        string            `json:"event_model_id"`
	EventModelName      string            `json:"event_model_name" form:"event_model_name"`
	EventType           string            `json:"type" form:"type"`
	Level               int               `json:"level" form:"level"`
	LevelName           string            `json:"level_name" form:"level_name"`
	Tags                []string          `json:"tags" form:"tags"`
	GenerateType        string            `json:"generate_type" form:"generate_type"`
	CreateTime          int64             `json:"@timestamp" form:"@timestamp"`
	DataSource          []string          `json:"data_source" form:"data_source"`
	DataSourceName      []string          `json:"data_source_name" form:"data_source_name"`
	DataSourceGroupName []string          `json:"data_source_group_name" form:"data_source_group_name"`
	DataSourceType      string            `json:"data_source_type" form:"data_source_type"`
	DefaultTimeWindow   TimeInterval      `json:"default_time_window" form:"default_time_window"`
	Schedule            Schedule          `json:"schedule" form:"schedule"`
	Labels              map[string]string `json:"labels" form:"labels"`
	Relations           map[string]any    `json:"relations,omitempty" form:"relations"`
	// EventForward        EventForward      `json:"event_forward,omitempty"`
	// AlgoAppId           int               `json:"algo_app_id"`
	// AlgoAppName         string            `json:"algo_app_name"`

	// LLMApp    string `json:"llm_app"`
	// KnwName   string `json:"knw_name"`
	// GraphName string `json:"graph_name"`
	//NOTE: 接收索引文档id
	DocId string `json:"__id,omitempty"`
}

type EventContext struct {
	Score         float64       `json:"score" form:"score"`
	Level         int           `json:"level" form:"level"`
	SourceRecords Records       `json:"source_records" form:"source_records"`
	GroupFields   []string      `json:"group_fields" form:"group_fields"`
	PreOrderEvent PreOrderEvent `json:"pre_order_event" form:"pre_order_event"`
	When          string        `json:"when,omitempty"`
	Where         string        `json:"where,omitempty"`
	What          string        `json:"what,omitempty"`
	HowMuch       string        `json:"how_much,omitempty"`
	Who           string        `json:"who,omitempty"`
	How           string        `json:"how,omitempty"`
	// Kn                EntityData     `json:"kn,omitempty"`
	// AnalysisStartTime int64      `json:"analysis_start_time,omitempty"`
	// AnalysisEndTime   int64      `json:"analysis_end_time,omitempty"`
	// RunId             string         `json:"run_id,omitempty" form:"run_id"`
	// Suggestion        map[string]any `json:"suggestion,omitempty"`
	IncidentId string `json:"incident_id,omitempty"`
}

type PreOrderEvent struct {
	Level int    `json:"level" form:"level"`
	Id    string `json:"id" form:"id"`
}

// 从索引库接收持久化信息
type EventData struct {
	BaseEvent
	EventForwardStr      string         `json:"event_forward_str"`
	DefaultTimeWindowStr string         `json:"default_time_window_str"`
	LabelsStr            string         `json:"labels_str"`
	Relations            map[string]any `json:"relations,omitempty"`
	RelationsStr         string         `json:"relations_str,omitempty"`
	Message              string         `json:"message"`
	EventMessage         string         `json:"event_message"`
	ContextStr           string         `json:"context_str"`
	Context              EventContext   `json:"context"`
	TriggerTime          int64          `json:"trigger_time"`
	TriggerDataStr       string         `json:"trigger_data_str"`
	TriggerData          Records        `json:"trigger_data"`
	ScheduleStr          string         `json:"schedule_str"`
	AggregateType        string         `json:"aggregate_type"`
	DetectType           string         `json:"detect_type"`
	AggregateAlgo        string         `json:"aggregate_algo"`
	DetectAlgo           string         `json:"detect_algo"`
	// RelationEvents       []RelationEvent `json:"relation_events"`
	// RelationEventsStr    string          `json:"relation_events_str"`
}

func (e EventData) SetDefaultTimeWindow() (TimeInterval, error) {
	var timeWindow TimeInterval
	if e.DefaultTimeWindowStr == "" {
		return TimeInterval{}, nil
	}
	err := json.Unmarshal([]byte(e.DefaultTimeWindowStr), &timeWindow)
	if err != nil {
		// common.GLogger.Errorf("ContextTransferToMessage error: %v", err)
		return TimeInterval{}, err
	}
	return timeWindow, nil
}

func (e EventData) SetLabels() (map[string]string, error) {
	var data map[string]string
	if e.LabelsStr == "" {
		return map[string]string{}, nil
	}
	err := json.Unmarshal([]byte(e.LabelsStr), &data)
	if err != nil {
		return map[string]string{}, err
	}
	return data, nil
}

func (e EventData) SetRelations() (map[string]any, error) {
	if e.RelationsStr == "" {
		return map[string]any{}, nil
	}
	var data map[string]any
	err := json.Unmarshal([]byte(e.RelationsStr), &data)
	if err != nil {
		return map[string]any{}, err
	}
	return data, nil
}

func (e EventData) SetTriggerData() (Records, error) {
	var data Records
	if e.TriggerDataStr == "" {
		return Records{}, nil
	}
	err := json.Unmarshal([]byte(e.TriggerDataStr), &data)
	if err != nil {
		return Records{}, err
	}
	return data, nil
}

func (e EventData) SetSchedule() (Schedule, error) {
	var data Schedule
	if e.ScheduleStr == "" {
		return Schedule{}, nil
	}
	err := json.Unmarshal([]byte(e.ScheduleStr), &data)
	if err != nil {
		// common.GLogger.Errorf("ContextTransferToMessage error: %v", err)
		return Schedule{}, err
	}
	return data, nil
}

func (e EventData) SetContext() (EventContext, error) {
	var msg EventContext
	if e.ContextStr == "" {
		return EventContext{}, nil
	}
	err := json.Unmarshal([]byte(e.ContextStr), &msg)
	if err != nil {
		// common.GLogger.Errorf("ContextTransferToMessage error: %v", err)
		return EventContext{}, err
	}
	return msg, nil
}

type IEvents []IEvent

type IEvent interface {
	GenerateMessage() any
	GetLevel() int
	GetTag() []string
	GetTriggerTime() int64
	GetTriggerData() Records
	GetBaseEvent() BaseEvent
	GetType() string
	GetContext() EventContext
}

// 返回持久化信息 (目前和聚合事件结构相同)
type EventRespData struct {
	BaseEvent
	Message       string         `json:"message"`
	Context       EventContext   `json:"context,omitempty"`
	TriggerTime   int64          `json:"trigger_time,omitempty"`
	TriggerData   Records        `json:"trigger_data,omitempty"`
	AggregateType string         `json:"aggregate_type,omitempty"`
	AggregateAlgo string         `json:"aggregate_algo,omitempty"`
	DetectType    string         `json:"detect_type,omitempty"`
	DetectAlgo    string         `json:"detect_algo,omitempty"`
	Relations     map[string]any `json:"relations,omitempty"`
}

func (e EventRespData) GenerateMessage() any     { return e.Message }
func (e EventRespData) GetLevel() int            { return e.Level }
func (e EventRespData) GetTag() []string         { return e.Tags }
func (e EventRespData) GetTriggerTime() int64    { return e.TriggerTime }
func (e EventRespData) GetTriggerData() Records  { return e.TriggerData }
func (e EventRespData) GetBaseEvent() BaseEvent  { return e.BaseEvent }
func (e EventRespData) GetType() string          { return e.EventType }
func (e EventRespData) GetContext() EventContext { return e.Context }

type AtomicEvent struct {
	BaseEvent
	TriggerTime int64        `json:"trigger_time" form:"trigger_time"`
	TriggerData Records      `json:"trigger_data" form:"trigger_data"`
	Message     string       `json:"message" form:"message"`
	Context     EventContext `json:"context" form:"context"`
	DetectType  string       `json:"detect_type,omitempty" binding:"omitempty"`
	DetectAlgo  string       `json:"detect_algo"`
}

func (e AtomicEvent) GenerateMessage() any     { return e.Message }
func (e AtomicEvent) GetLevel() int            { return e.Level }
func (e AtomicEvent) GetTag() []string         { return e.Tags }
func (e AtomicEvent) GetTriggerTime() int64    { return e.TriggerTime }
func (e AtomicEvent) GetTriggerData() Records  { return e.TriggerData }
func (e AtomicEvent) GetBaseEvent() BaseEvent  { return e.BaseEvent }
func (e AtomicEvent) GetType() string          { return e.EventType }
func (e AtomicEvent) GetContext() EventContext { return e.Context }

type AggregateEvent struct {
	BaseEvent
	Message       string       `json:"message" form:"message"`
	Context       EventContext `json:"context" form:"context"`
	AggregateType string       `json:"aggregate_type,omitempty"`
	AggregateAlgo string       `json:"aggregate_algo,omitempty"`
	TriggerTime   int64        `json:"trigger_time" form:"trigger_time"`
	TriggerData   Records      `json:"trigger_data" form:"trigger_data"`
}

func (e AggregateEvent) GenerateMessage() any     { return e.Message }
func (e AggregateEvent) GetLevel() int            { return e.Level }
func (e AggregateEvent) GetTag() []string         { return e.Tags }
func (e AggregateEvent) GetTriggerTime() int64    { return e.TriggerTime }
func (e AggregateEvent) GetTriggerData() Records  { return e.TriggerData }
func (e AggregateEvent) GetBaseEvent() BaseEvent  { return e.BaseEvent }
func (e AggregateEvent) GetType() string          { return e.EventType }
func (e AggregateEvent) GetContext() EventContext { return e.Context }
