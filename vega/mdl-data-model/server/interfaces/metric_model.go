// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	dcond "data-model/interfaces/condition"
)

const (
	//模块名称
	METRIC_MODEL_MODULE       = "MetricModel"
	METRIC_MODEL_GROUP_MODULE = "MetricModelGroup"
	METRIC_MODEL_TASK         = "MetricModelTask"

	//模块类型
	MODULE_TYPE_METRIC_MODEL       = "metric_model"
	MODULE_TYPE_METRIC_MODEL_GROUP = "metric_model_group"

	//对象类型
	OBJECTTYPE_METRIC_MODEL       = "ID_AUDIT_METRIC_MODEL"
	OBJECTTYPE_METRIC_MODEL_GROUP = "ID_AUDIT_METRIC_MODEL_GROUP"

	DEFAULT_METRIC_MODEL_GROUP_SORT = "group_name"

	//模块名称
	METRIC_MODEL_MEASURE = "MetricMeasureName"

	// 数据源类型
	// DATA_SOURCE_DATA_VIEW       = "data_view"
	// DATA_SOURCE_VEGA_LOGIC_VIEW = "vega_logic_view"

	// measure prefix
	MEASURE_PREFIX = "__m."

	// 度量名称正则, 只允许以 __m. 开头，且自定义部分只允许字母、数字、下划线
	MEASURE_NAME_RULE = `^__m.[0-9a-zA-Z][\w]*$`

	// metric type
	ATOMIC_METRIC     string = "atomic"
	DERIVED_METRIC    string = "derived"
	COMPOSITED_METRIC string = "composite"

	// query type
	PROMQL     string = "promql"
	DSL        string = "dsl"
	DSL_CONFIG string = "dsl_config"
	SQL        string = "sql"

	// promql dateField & metricField
	PROMQL_DATEFIELD   string = "@timestamp"
	PROMQL_METRICFIELD string = "value"

	// sql 默认的度量字段, as 的别名
	SQL_METRICFIELD string = "__sql_value"

	// 值字段
	VALUE_FIELD_NAME         string = "__value"
	VALUE_FIELD_DISPLAY_NAME string = "指标值"

	// unit type
	UNIT_NUM        string = "numUnit"
	UNIT_STORE      string = "storeUnit"
	UNIT_TRANS_RATE string = "transmissionRate"
	UNIT_TIME       string = "timeUnit"
	CURRENCY_UNIT   string = "currencyUnit"
	PERCENTAGE_UNIT string = "percentageUnit"
	COUNT_UNIT      string = "countUnit" // 对象量词
	WEIGHT_UNIT     string = "weightUnit"

	// 内置变量
	VARIABLE_INTERVAL string = "__interval"

	// 时间区间常量
	INTERVAL_5MIN string = "5m"

	// AGGS, AGGREGATIONS string = "aggs", "aggregations"
	AGGS string = "aggs"

	// 聚合类型
	TERMS      string = "terms"
	FILTERS    string = "filters"
	RANGE      string = "range"
	DATE_RANGE string = "date_range"

	// MULTI_TERMS    string = "multi_terms"
	DATE_HISTOGRAM string = "date_histogram"
	VALUE_COUNT    string = "value_count"
	CARDINALITY    string = "cardinality"
	SUM            string = "sum"
	AVG            string = "avg"
	MAX            string = "max"
	MIN            string = "min"
	TOP_HITS       string = "top_hits"

	// 允许追溯的最大的数据时间点数
	MAX_RETRACE_POINTS_NUM float64 = 10000

	SCHEDULE_TYPE_FIXED = "FIX_RATE" // 与 xxljob 的类型保持一致
	SCHEDULE_TYPE_CRON  = "CRON"     // 与 xxljob 的类型保持一致

	// 前一个时间单位的窗口
	PREVIOUS_HOUR string = "previous_hour"
	PREVIOUS_DAY  string = "previous_day"
	PREVIOUS_WEEK string = "previous_week"
	// PREVIOUS_MONTH string = "previous_month"

	// 任务同步状态
	// SCHEDULE_SYNC_STATUS_CREATE  = 0 // 创建中
	// SCHEDULE_SYNC_STATUS_UPDATE  = 1 // 更新中
	// SCHEDULE_SYNC_STATUS_DELETE  = 2 // 删除中
	SCHEDULE_SYNC_STATUS_FINISH  = 3 // 执行中
	SCHEDULE_SYNC_STATUS_SUCCESS = 4 // 执行成功
	SCHEDULE_SYNC_STATUS_FAILED  = 5 // 执行失败

	// 任务类型 calendar 和 fixed
	// TASK_TYPE_CALENDAR             = "calendar"
	// TASK_TYPE_FIXED                = "fixed"
	// CALENDAR_TASK_NAME             = "日历步长持久化任务"
	// CALENDAR_SCHEDULE_EXPR         = "0 0 0 * * ?"
	// CALNEDAR_STEP_DAY              = "day"
	// CALNEDAR_STEP_WEEK             = "week"
	// CALNEDAR_STEP_MONTH            = "month"
	// CALNEDAR_STEP_QUARTER          = "quarter"
	// CALNEDAR_STEP_YEAR             = "year"
	// CALENDAR_TASK_RETRACE_DURATION = "30d"

	// xxl-job任务执行超时时间.单位是秒.设置大一点,在高基序列,高追溯时长的情况下,第一次任务触发的耗时比较大.设置为24h
	XXL_JOB_TASK_EXEC_TIMEOUT int = 86400
)

var (
	METRIC_MODEL_SORT = map[string]string{
		"group_name":  "f_group_name",
		"model_name":  "f_model_name",
		"update_time": "f_update_time",
	}

	METRIC_MODEL_GROUP_SORT = map[string]string{
		"group_name": "f_group_name",
	}

	MEASURE_FIELD_TYPE = map[string]string{
		"long":          "long",
		"integer":       "integer",
		"short":         "short",
		"byte":          "byte",
		"double":        "double",
		"float":         "float",
		"half_float":    "half_float",
		"scaled_float":  "scaled_float",
		"unsigned_long": "unsigned_long",
	}

	UnitTypeMap = map[string]map[string]string{
		UNIT_NUM: {
			"none": "none",
			"K":    "K",
			"Mil":  "Mil",
			"Bil":  "Bil",
			"Tri":  "Tri",
		},
		UNIT_STORE: {
			"bit":  "Byte",
			"Byte": "Byte",
			"KB":   "KiB",
			"KiB":  "KiB",
			"MB":   "MiB",
			"MiB":  "MiB",
			"GB":   "GiB",
			"GiB":  "GiB",
			"TB":   "TiB",
			"TiB":  "TiB",
			"PB":   "PiB",
			"PiB":  "PiB",
		},
		UNIT_TRANS_RATE: {
			"bps":   "B/s",
			"B/s":   "B/s",
			"Kbps":  "KiB/s",
			"KiB/s": "KiB/s",
			"Mbps":  "MiB/s",
			"MiB/s": "MiB/s",
		},
		UNIT_TIME: {
			"ns":      "ns",
			"μs":      "μs",
			"ms":      "ms",
			"s":       "s",
			"m":       "m",
			"h":       "h",
			"d":       "d",
			"day":     "d",
			"week":    "d",
			"month":   "d",
			"year":    "d",
			"quarter": "d",
		},
		CURRENCY_UNIT: {
			"Fen":      "Fen",
			"Jiao":     "Jiao",
			"CNY":      "CNY",
			"10K_CNY":  "10K_CNY",
			"1M_CNY":   "1M_CNY",
			"100M_CNY": "100M_CNY",
			"US_Cent":  "US_Cent",
			"USD":      "USD",
			"EUR_Cent": "EUR_Cent",
		},
		PERCENTAGE_UNIT: {
			"%": "%",
			"‰": "‰",
		},
		COUNT_UNIT: {
			"household":   "household",
			"transaction": "transaction",
			"piece":       "piece",
			"item":        "item",
			"times":       "times",
			"man_day":     "man_day",
			"family":      "family",
			"hand":        "hand",
			"sheet":       "sheet",
			"packet":      "packet",
		},
		WEIGHT_UNIT: {
			"ton": "ton",
			"kg":  "kg",
		},
	}

	OPERATION_TYPE = map[string]string{
		dcond.Operation_IN:        dcond.Operation_IN,
		dcond.Operation_EQ:        dcond.Operation_EQ,
		dcond.Operation_NE:        dcond.Operation_NE,
		dcond.Operation_RANGE:     dcond.Operation_RANGE,
		dcond.Operation_OUT_RANGE: dcond.Operation_OUT_RANGE,
		dcond.Operation_LIKE:      dcond.Operation_LIKE,
		dcond.Operation_NOT_LIKE:  dcond.Operation_NOT_LIKE,
		dcond.Operation_GT:        dcond.Operation_GT,
		dcond.Operation_GTE:       dcond.Operation_GTE,
		dcond.Operation_LT:        dcond.Operation_LT,
		dcond.Operation_LTE:       dcond.Operation_LTE,
	}

	PREVIOUS_TIMEWINDOW = map[string]string{
		PREVIOUS_HOUR: PREVIOUS_HOUR,
		PREVIOUS_DAY:  PREVIOUS_DAY,
		PREVIOUS_WEEK: PREVIOUS_WEEK,
		// PREVIOUS_MONTH: PREVIOUS_MONTH,
	}

	// 系统预留分组 ID, value 代表list groups返回的顺序
	// RESERVED_METRIC_MODEL_GROUP_ID = map[string]int{
	// 	"":                             0,
	// 	Metric_GroupID_ARObservability: 1,
	// 	Metric_GroupID_Event:           2,
	// }
)

// 给结构模型使用，待结构模型下线，此对象删除
// 模型基础信息, 主要结构模型使用，先注释掉
// type ModelBasicInfo struct {
// 	ModelID string `json:"id"`
// 	ModelSimpleInfo
// }

// 模型除模型ID外的基础信息, 包括分组名称和模型名称
type ModelSimpleInfo struct {
	ModelName string `json:"name"`
	GroupName string `json:"group_name"`
	UnitType  string `json:"unit_type,omitempty"`
	Unit      string `json:"unit,omitempty"`
}

type CreateMetricModel struct {
	ModelID         string                  `json:"id"`
	ModelName       string                  `json:"name"`
	CatalogID       string                  `json:"catalog_id"`
	CatalogContent  string                  `json:"catalog_content"`
	MeasureName     string                  `json:"measure_name"`
	GroupName       string                  `json:"group_name"`
	Tags            []string                `json:"tags"`
	Comment         string                  `json:"comment"`
	MetricType      string                  `json:"metric_type"`
	DataViewID      string                  `json:"data_view_id"`
	DataSource      *CreateMetricDataSource `json:"data_source"`
	QueryType       string                  `json:"query_type"`
	Formula         string                  `json:"formula"`
	FormulaConfig   any                     `json:"formula_config"`
	OrderByFields   []OrderField            `json:"order_by_fields,omitempty"`
	HavingCondition *CondCfg                `json:"having_condition,omitempty"`
	AnalysisDims    []Field                 `json:"analysis_dimensions"`
	DateField       string                  `json:"date_field"`
	MeasureField    string                  `json:"measure_field"`
	UnitType        string                  `json:"unit_type"`
	Unit            string                  `json:"unit"`
	Builtin         bool                    `json:"builtin"`
	Task            *CreateMetricTask       `json:"task"`
	ModuleType      string                  `json:"module_type"`
}

type OrderField struct {
	Field
	Direction string `json:"direction"` // 排序方向
}

type Field struct {
	Name        string  `json:"name"` // 技术名
	Type        string  `json:"type"`
	DisplayName string  `json:"display_name"`      // 显示名
	Comment     *string `json:"comment,omitempty"` // 从视图中获取到的字段的备注信息
}

type CreateMetricTask struct {
	TaskName        string   `json:"name"`
	Schedule        Schedule `json:"schedule"`
	TimeWindows     []string `json:"time_windows"`
	Steps           []string `json:"steps"`
	IndexBase       string   `json:"index_base"`
	RetraceDuration string   `json:"retrace_duration"`
	Comment         string   `json:"comment"`
}

type CreateMetricDataSource struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type SQLConfig struct {
	Condition           *CondCfg  `json:"condition,omitempty"`
	ConditionStr        string    `json:"condition_str,omitempty"`
	AggrExpr            *AggrExpr `json:"aggr_expression,omitempty"`
	AggrExprStr         string    `json:"aggr_expression_str,omitempty"`
	GroupByFields       []string  `json:"group_by_fields,omitempty"`
	GroupByFieldsDetail []Field   `json:"group_by_fields_detail,omitempty"`
}

type AggrExpr struct {
	Field string `json:"field"`
	Aggr  string `json:"aggr"`
}

// 衍生指标配置项
type DerivedConfig struct {
	DependMetricModel *DependMetricModel `json:"depend_metric_model"`
	DateCondition     *CondCfg           `json:"date_condition,omitempty"`
	BusinessCondition *CondCfg           `json:"business_condition,omitempty"`
	ConditionStr      string             `json:"condition_str,omitempty"`
}

type DependMetricModel struct {
	ID        string `json:"id"`
	GroupName string `json:"group_name,omitempty"`
	Name      string `json:"name,omitempty"`
}

// 简单指标模型结构体
type SimpleMetricModel struct {
	ModelID        string   `json:"id"`
	ModelName      string   `json:"name"`
	CatalogID      string   `json:"catalog_id"`
	CatalogContent string   `json:"catalog_content"`
	MeasureName    string   `json:"measure_name"`
	GroupID        string   `json:"group_id"`
	GroupName      string   `json:"group_name"`
	Tags           []string `json:"tags"`
	Comment        string   `json:"comment"`
	MetricType     string   `json:"metric_type"`
	DataViewID     string   `json:"data_view_id"`

	QueryType          string       `json:"query_type"`
	Formula            string       `json:"formula"`
	FormulaConfig      any          `json:"formula_config,omitempty"`
	OrderByFields      []OrderField `json:"order_by_fields,omitempty"`
	HavingCondition    *CondCfg     `json:"having_condition,omitempty"`
	AnalysisDims       []Field      `json:"analysis_dimensions,omitempty"`
	DateField          string       `json:"date_field"`
	MeasureField       string       `json:"measure_field"`
	UnitType           string       `json:"unit_type"`
	Unit               string       `json:"unit"`
	Builtin            bool         `json:"builtin"`
	IsCalendarInterval int          `json:"is_calendar_interval"` // 默认值为0，0为非日历步长，1为日历步长
	Creator            AccountInfo  `json:"creator"`
	CreateTime         int64        `json:"create_time"`
	UpdateTime         int64        `json:"update_time"`

	// 操作权限
	Operations []string `json:"operations"`
}

// 指标模型结构体
type MetricModel struct {
	SimpleMetricModel
	DataSource   *MetricDataSource `json:"data_source"`
	DataViewName string            `json:"data_view_name"`
	Task         *MetricTask       `json:"task,omitempty"`
	FieldsMap    map[string]Field  `json:"fields_map"` // 字段集
	ModuleType   string            `json:"module_type"`

	IfContainTopHits bool `json:"-"`
	IfNameModify     bool `json:"-"`
}

type MetricDataSource struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

type MetricModelQueryConfig struct {
	Bucket      string
	Aggregation string
}

// 包含了数据视图过滤条件的指标模型结构体
type MetricModelWithFilters struct {
	MetricModel
	DataView *DataView `json:"data_view,omitempty"`
}

// 组合名称结构体
type CombinationName struct {
	GroupName string
	ModelName string
}

// 指标模型分组结构体
type MetricModelGroup struct {
	GroupID          string `json:"id"`
	GroupName        string `json:"name"`
	Comment          string `json:"comment"`
	CreateTime       int64  `json:"create_time"`
	UpdateTime       int64  `json:"update_time"`
	Builtin          bool   `json:"builtin"`
	MetricModelCount int    `json:"metric_model_count"`
}

// 接受请求体参数
type MetricModelGroupName struct {
	GroupName string `json:"group_name"`
}

// 指标模型列表查询参数
type MetricModelsQueryParams struct {
	PaginationQueryParameters
	NamePattern string
	Name        string
	MetricType  string
	QueryType   string
	Tag         string
	GroupID     string
}

// 指标模型列表查询参数
type ListMetricGroupQueryParams struct {
	Builtin []bool
	PaginationQueryParameters
}

type DataSourceField struct {
	Name string      `json:"name"`
	Type interface{} `json:"type"`
}

type MetricTask struct {
	TaskID             string      `json:"id"`
	TaskName           string      `json:"name,omitempty"`
	ModuleType         string      `json:"module_type"`
	ModelID            string      `json:"model_id"`
	MeasureName        string      `json:"measure_name,omitempty"`
	Schedule           Schedule    `json:"schedule"`
	TimeWindows        []string    `json:"time_windows,omitempty"`
	Steps              []string    `json:"steps"`
	IndexBase          string      `json:"index_base"`
	IndexBaseName      string      `json:"index_base_name"`
	RetraceDuration    string      `json:"retrace_duration"`
	Comment            string      `json:"comment"`
	ScheduleSyncStatus int         `json:"schedule_sync_status"`
	ExeccuteStatus     int         `json:"execute_status"`
	CreateTime         int64       `json:"create_time"`
	UpdateTime         int64       `json:"update_time"`
	PlanTime           int64       `json:"plan_time"`
	Creator            AccountInfo `json:"creator"`
}

type Schedule struct {
	Type       string `json:"type"`
	Expression string `json:"expression"`
}

type TaskSyncStatus struct {
	SyncStatus int
	UpdateTime int64
	TaskIDs    []string
	// ModelIDs   []string
}

func IsValidMetricType(m string) bool {
	return m == ATOMIC_METRIC || m == DERIVED_METRIC || m == COMPOSITED_METRIC
}

func IsValidQueryType(m string) bool {
	return m == PROMQL || m == DSL || m == DSL_CONFIG || m == SQL
}

func IsValidUnitType(m string) bool {
	return m == UNIT_NUM || m == UNIT_STORE || m == UNIT_TRANS_RATE || m == UNIT_TIME ||
		m == CURRENCY_UNIT || m == PERCENTAGE_UNIT || m == COUNT_UNIT || m == WEIGHT_UNIT
}

func IsValidUnit(unitType string, unit string) (bool, string) {
	units, ok := UnitTypeMap[unitType]
	if !ok {
		return false, ""
	}

	newUnit, ok := units[unit]
	if !ok {
		return false, ""
	}

	return true, newUnit
}

// func IsValidDataSourceType(m string) bool {
// 	return m == DATA_SOURCE_DATA_VIEW || m == DATA_SOURCE_VEGA_LOGIC_VIEW
// }
