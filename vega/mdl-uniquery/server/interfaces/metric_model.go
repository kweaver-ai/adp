// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"regexp"
	"time"
	"uniquery/common/condition"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

const (
	// 指标数据查询使用缓存（持久化）数据的默认值
	DEFAULT_IGNORING_STORE_CACHE = "false"
	// 指标数据查询使用缓存（持久化）数据的默认值
	DEFAULT_IGNORING_MEMORY_CACHE = "false"
	// 指标数据查询忽略高基查询的默认值
	DEFAULT_IGNORING_HCTS = "false"
	// 指标数据查询忽略高基查询的默认值
	DEFAULT_FILL_NULL = "true"

	// 指标数据查询是否包含模型信息
	DEFAULT_INCLUDE_MODEL = "false"
	// 指标数据查询默认的序列数 limit
	DEFAULT_SERIES_LIMIT     = "-1" // LIMI
	DEFAULT_SERIES_LIMIT_INT = -1

	// metric type
	ATOMIC_METRIC     string = "atomic"
	DERIVED_METRIC    string = "derived"
	COMPOSITED_METRIC string = "composite"

	// query type
	PROMQL     string = "promql"
	DSL        string = "dsl"
	DSL_CONFIG string = "dsl_config"
	SQL        string = "sql"

	// 数据源类型
	// DATA_SOURCE_DATA_VIEW       = "data_view"
	// DATA_SOURCE_VEGA_LOGIC_VIEW = "vega_logic_view"

	// filters operation
	OPERATION_IN string = "in"
	OPERATION_EQ string = "="

	// AGGS, AGGREGATIONS string = "aggs", "aggregations"
	AGGS string = "aggs"

	// 分桶类型
	BUCKET_TYPE_TERMS          string = "terms"
	BUCKET_TYPE_RANGE          string = "range"
	BUCKET_TYPE_FILTERS        string = "filters"
	BUCKET_TYPE_GEOHASH_GRID   string = "geohash_grid"
	BUCKET_TYPE_DATE_RANGE     string = "date_range"
	BUCKET_TYPE_DATE_HISTOGRAM string = "date_histogram"
	// MULTI_TERMS    string = "multi_terms"

	// 聚合类型
	AGGR_TYPE_DOC_COUNT      string = "doc_count"
	AGGR_TYPE_VALUE_COUNT    string = "value_count"
	AGGR_TYPE_CARDINALITY    string = "cardinality"
	AGGR_TYPE_SUM            string = "sum"
	AGGR_TYPE_AVG            string = "avg"
	AGGR_TYPE_MAX            string = "max"
	AGGR_TYPE_MIN            string = "min"
	AGGR_TYPE_PERCENTILES    string = "percentiles"
	AGGR_TYPE_TOP_HITS       string = "top_hits"
	AGGR_TYPE_COUNT_DISTINCT string = "count_distinct"

	DEFAULT_CARDINALITY_PRECISION_THRESHOLD int = 40000
	BUCKET_GEOHASH_GRID_MIN_PRECISION       int = 1
	BUCKET_GEOHASH_GRID_MAX_PRECISION       int = 12
	MAX_DATE_HISTOGRAM_BUCKET_SIZE          int = 10000

	// 时间间隔类型
	INTERVAL_TYPE_CALENDAR string = "calendar"
	INTERVAL_TYPE_FIXED    string = "fixed"

	// 时间间隔值
	AUTO_INTERVAL     string = "auto"
	VARIABLE_INTERVAL string = "{{__interval}}"

	DATE_HISTOGRAM_FIELD string = "__date_histogram"
	PERCENT_FIELD        string = "__percent"
	OTHER_FIELD          string = "__other"
	VALUE_FIELD          string = "__value"
	TIME_FIELD           string = "__time"

	DEFAULT_DATE_FIELD string = "@timestamp"

	TERMS_ORDER_TYPE_FIELD string = "field"
	TERMS_ORDER_TYPE_VALUE string = "value"
	TERMS_ORDER_TYPE_COUNT string = "count"

	CALENDAR_STEP_MINUTE  string = "minute"
	CALENDAR_STEP_HOUR    string = "hour"
	CALENDAR_STEP_DAY     string = "day"
	CALENDAR_STEP_WEEK    string = "week"
	CALENDAR_STEP_MONTH   string = "month"
	CALENDAR_STEP_QUARTER string = "quarter"
	CALENDAR_STEP_YEAR    string = "year"
	// CalendarStep_1m      string = "1m"
	// CalendarStep_1h      string = "1h"
	// CalendarStep_1d      string = "1d"
	// CalendarStep_1w      string = "1w"
	// CalendarStep_1M      string = "1M"
	// CalendarStep_1q      string = "1q"
	// CalendarStep_1y      string = "1y"

	// 度量名称正则, 只允许以 __m. 开头，且自定义部分只允许字母、数字、下划线
	MEASURE_NAME_RULE = `^__m\.[0-9a-zA-Z][\w]*$`

	// 时间区间常量
	INTERVAL_5MIN string = "5m"

	// continuous_k_minute_downtime 的 k 最大值是60
	MAX_K_MINUTE = 60

	// 过滤模式
	FILTER_MODE_NORMAL = "normal"
	FILTER_MODE_ERROR  = "error"
	FILTER_MODE_IGNORE = "ignore"
)

var (
	CALENDAR_INTERVALS = map[string]string{
		CALENDAR_STEP_MINUTE:  CALENDAR_STEP_MINUTE,
		CALENDAR_STEP_HOUR:    CALENDAR_STEP_HOUR,
		CALENDAR_STEP_DAY:     CALENDAR_STEP_DAY,
		CALENDAR_STEP_WEEK:    CALENDAR_STEP_WEEK,
		CALENDAR_STEP_MONTH:   CALENDAR_STEP_MONTH,
		CALENDAR_STEP_QUARTER: CALENDAR_STEP_QUARTER,
		CALENDAR_STEP_YEAR:    CALENDAR_STEP_YEAR,

		// CalendarStep_1m: CALENDAR_STEP_MINUTE,
		// CalendarStep_1h: CALENDAR_STEP_HOUR,
		// CalendarStep_1d: CALENDAR_STEP_DAY,
		// CalendarStep_1w: CALENDAR_STEP_WEEK,
		// CalendarStep_1M: CALENDAR_STEP_MONTH,
		// CalendarStep_1q: CALENDAR_STEP_QUARTER,
		// CalendarStep_1y: CALENDAR_STEP_YEAR,
	}

	MEASURE_REGEX = regexp.MustCompile(MEASURE_NAME_RULE)

	INDEX_BASE_SPLIT_TIME = make(map[string]time.Time)

	INDEX_PATTERN_SPLIT_TIME = make(map[string]time.Time)

	// DEFAULT_QUERY_TIME_ZONE, _ = time.LoadLocation(os.Getenv("TZ"))
	DEFAULT_QUERY_TIME_ZONE = time.Local

	DEFAULT_MAX_QUERY_POINTS        int64 = 10000 // promql一批默认的查询最大点数，默认值为10000
	DEFAULT_DSL_MAX_QUERY_POINTS    int64 = 30000 // dsl一批默认查询的最大点数，默认值为30000
	DEFAULT_GET_SERIES_NUM_BY_BATCH int64 = 5000  // 大于5000的序列认为是高基,需并发查询，兼顾dsl的多个字段聚合
	DEFAULT_SERIES_NUM              int64 = 1000  // 默认需并发查询的序列基数,大于1000的序列认为是高基
	PROMQL_BATCH_MAX_SERIES_SIZE    int64 = 1000  // promql一批次查询的最大数

	DEFAULT_SCROLL time.Duration = 0

	// dsl按文档数排序方式: __count, __key
	DSL_TERMS_ORDER_BY_COUNT string = "_count"
	DSL_TERMS_ORDER_BY_KEY   string = "_key"

	// 同环比\占比类型
	METRICS_SAMEPERIOD string = "sameperiod"
	METRICS_PROPORTION string = "proportion"

	// 同环比计算函数
	METRICS_SAMEPERIOD_METHOD_GROWTH_VALUE string = "growth_value"
	METRICS_SAMEPERIOD_METHOD_GROWTH_RATE  string = "growth_rate"

	// 同环比计算的时间粒度
	METRICS_SAMEPERIOD_TIME_GRANULARITY_DAY     string = "day"
	METRICS_SAMEPERIOD_TIME_GRANULARITY_MONTH   string = "month"
	METRICS_SAMEPERIOD_TIME_GRANULARITY_QUARTER string = "quarter"
	METRICS_SAMEPERIOD_TIME_GRANULARITY_YEAR    string = "year"
)

// 指标数据预览的查询请求体
type MetricModelQuery struct {
	QueryTimeParams
	RequestMetrics  *RequestMetrics    `json:"metrics,omitempty"`
	MetricType      string             `json:"metric_type"`
	DataViewID      string             `json:"data_view_id"`
	DataSource      *MetricDataSource  `json:"data_source"`
	QueryType       string             `json:"query_type"`
	Formula         string             `json:"formula"`
	FormulaConfig   any                `json:"formula_config"`
	AnalysisDims    []string           `json:"analysis_dimensions,omitempty"`
	OrderByFields   []OrderField       `json:"order_by_fields,omitempty"`
	HavingCondition *condition.CondCfg `json:"having_condition,omitempty"`
	DateField       string             `json:"date_field"`
	MeasureField    string             `json:"measure_field"`
	IsModelRequest  bool               `json:"is_model_request"` // 是否是来自data-model的请求.data-model的请求是做计算公式有效性检查的, 其不需要再次校验dsl公式中的规则的请求，此时无需校验 dsl 公式是否符合规则，且获取序列时只获取一个序列,且用深度优先的方式收据,预览时需要校验。
	Filters         []Filter           `json:"filters"`
	MetricModelID   string             `json:"-"` // query参数, 请求的指标模型id
	// DataView            DataView          `json:"-"`          // 指标模型的数据视图信息(预览时没有模型信息,需要请求数据源信息,所以用query来接这个参数)
	IsVariable          bool   `json:"-"`          // 模型中的interval是否使用了变量(dsl中解析出来的)
	IsCalendar          bool   `json:"-"`          // 模型中的分桶是否是使用calendar_interval
	HasMatchPersist     bool   `json:"-"`          // 是否已经匹配了持久化查询
	ContainTopHits      bool   `json:"-"`          // 模型中的聚合是否使用来top_hits
	ModelName           string `json:"model_name"` // 指标模型名称
	IsCalendarInterval  int    `json:"-"`          // 默认值为0，0为非日历步长，1为日历步长
	MaxSearchSeriesSize int    `json:"-"`          // 向opensearch发起一次查询的最大序列数
	ModelUpdateTime     int64  `json:"-"`          // 模型更新时间
	QueryTimeNum        int64  `json:"-"`          // 当前查询预估的时间点数
	MetricModelQueryParameters
	Condition    *condition.CondCfg `json:"-"` // 衍生指标的过滤条件在往原子指标传
	ConditionStr string             `json:"-"` // 衍生指标的过滤条件在往原子指标传

	TraceTotalHits bool  `json:"-"` // 是否需要获取数据TotalHits
	FixedStart     int64 `json:"-"`
	FixedEnd       int64 `json:"-"`

	ViewQuery4Metric // 视图-索引的相关信息
}

type QueryIndicesInfo struct {
	IndexShards []*IndexShards `json:"-"` // 视图对应的索引列表的索引分片信息
	Indices     []string       `json:"-"` // 视图对应的索引列表
	QueryStr    string         `json:"-"` // 视图本身构造的过滤条件，sql或dsl
}

// 指标模型和目标模型请求中的时间参数
type QueryTimeParams struct {
	Start          *int64  `json:"start"`
	End            *int64  `json:"end"`
	StepStr        *string `json:"step"`
	IsInstantQuery bool    `json:"instant"` // 用于标记 instant query，默认为 false，即默认是 query_range
	Step           *int64  `json:"-"`

	// 历史兼容所需
	Time             int64  `json:"time"`            // 瞬时查询的时间
	LookBackDeltaStr string `json:"look_back_delta"` // 瞬时查询时从 time 往前回退的时间区间
	// LookBackDelta    int64  `json:"-"`
}

type RequestMetrics struct {
	Type          string         `json:"type"`
	SamePeriodCfg *SamePeriodCfg `json:"sameperiod_config,omitempty"`
}

type SamePeriodCfg struct {
	Method          []string `json:"method"`
	Offset          int      `json:"offset"`
	TimeGranularity string   `json:"time_granularity"`
}

type OrderField struct {
	Name        string `json:"name"` // 视图字段名
	Type        string `json:"type"`
	DisplayName string `json:"display_name"`
	Direction   string `json:"direction"` // 排序方向
}

type MetricModelQueryParameters struct {
	Offset              int64
	Limit               int64
	IncludeModel        bool
	IgnoringHCTS        bool
	IgnoringStoreCache  bool
	IgnoringMemoryCache bool
	FilterMode          string // query参数, 查询过滤模式
	FillNull            bool
}

type Filter struct {
	Name          string `json:"name"`
	IsResultField bool   `json:"is_result_field"`
	Operation     string `json:"operation"`
	Value         any    `json:"value"`
}

type MetricModelUniResponse struct {
	Model           interface{}       `json:"model,omitempty"`
	Datas           []MetricModelData `json:"datas"`
	Step            *string           `json:"step,omitempty"`
	IsVariable      bool              `json:"is_variable"`
	IsCalendar      bool              `json:"is_calendar"`
	IsQueryByBatch  bool              `json:"is_query_by_batch,omitempty"`
	HasMatchPersist bool              `json:"has_match_persist"`
	CurrSeriesNum   int               `json:"-"`
	PointTotal      int               `json:"-"`
	StatusCode      int               `json:"status_code"`
	SeriesTotal     int               `json:"series_total,omitempty"`
	VegaDurationMs  int64             `json:"vega_duration_ms,omitempty"`
	OverallMs       int64             `json:"overall_ms,omitempty"`
	Err             error             `json:"-"`
}

type MetricModelData struct {
	Labels       map[string]string `json:"labels"`
	Times        []any             `json:"times"`
	TimeStrs     []string          `json:"time_strs,omitempty"`
	Values       []any             `json:"values"`
	GrowthValues []any             `json:"growth_values,omitempty"`
	GrowthRates  []any             `json:"growth_rates,omitempty"`
	Proportions  []any             `json:"proportions,omitempty"`
}

type MetricModel struct {
	ModelID            string             `json:"id"`
	ModelName          string             `json:"name"`
	MeasureName        string             `json:"measure_name"`
	MetricType         string             `json:"metric_type"`
	DataSource         *MetricDataSource  `json:"data_source"`
	QueryType          string             `json:"query_type"`
	Formula            string             `json:"formula"`
	FormulaConfig      any                `json:"formula_config,omitempty"`
	OrderByFields      []OrderField       `json:"order_by_fields,omitempty"`
	HavingCondition    *condition.CondCfg `json:"having_condition,omitempty"`
	AnalysisDims       []Field            `json:"analysis_dimensions,omitempty"`
	DateField          string             `json:"date_field"`
	MeasureField       string             `json:"measure_field"`
	UnitType           string             `json:"unit_type"`
	Unit               string             `json:"unit"`
	GroupID            string             `json:"group_id"`
	GroupName          string             `json:"group_name"`
	Tags               []string           `json:"tags"`
	Comment            string             `json:"comment"`
	CreateTime         int64              `json:"create_time"`
	UpdateTime         int64              `json:"update_time"`
	Builtin            bool               `json:"builtin"`
	IsCalendarInterval int                `json:"is_calendar_interval"` // 默认值为0，0为非日历步长，1为日历步长
	Task               *MetricTask        `json:"task,omitempty"`
	FieldsMap          map[string]Field   `json:"fields_map,omitempty"` // 字段集.预览时暂时不返回字段集信息，省略掉。
}

type Field struct {
	Name        string  `json:"name"` // 技术名
	Type        string  `json:"type"`
	DisplayName string  `json:"display_name"`      // 显示名
	Comment     *string `json:"comment,omitempty"` // 显示名
}

type FieldValues struct {
	FieldName string   `json:"-"`
	Type      string   `json:"type"`
	Values    []string `json:"values"`
}

type MetricDataSource struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type SQLConfig struct {
	Condition           *condition.CondCfg `json:"condition,omitempty"`
	ConditionStr        string             `json:"condition_str,omitempty"`
	AggrExpr            *AggrExpr          `json:"aggr_expression,omitempty"`
	AggrExprStr         string             `json:"aggr_expression_str,omitempty"`
	GroupByFields       []string           `json:"group_by_fields,omitempty"`
	GroupByFieldsDetail []Field            `json:"group_by_fields_detail,omitempty"`
}

type AggrExpr struct {
	Field string `json:"field"`
	Aggr  string `json:"aggr"`
}

// 衍生指标配置项
type DerivedConfig struct {
	DependMetricModel *DependMetricModel `json:"depend_metric_model"`
	DateCondition     *condition.CondCfg `json:"date_condition,omitempty"`
	BusinessCondition *condition.CondCfg `json:"business_condition,omitempty"`
	ConditionStr      string             `json:"condition_str,omitempty"`
}

type DependMetricModel struct {
	ID        string `json:"id"`
	GroupName string `json:"group_name,omitempty"`
	Name      string `json:"name,omitempty"`
}

type MetricModelFormulaConfig struct {
	Buckets       []*MetricModelFormulaConfigBucket      `json:"buckets"`
	DateHistogram *MetricModelFormulaConfigDateHistogram `json:"date_histogram"`
	Aggregation   *MetricModelFormulaConfigAggregation   `json:"aggregation"`
	QueryString   string                                 `json:"query_string"`
}

type MetricModelFormulaConfigBucket struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	BktName string `json:"-"`
	Field   string `json:"field"`

	// 词条分桶
	Size      int    `json:"size"`
	Order     string `json:"order"`
	Direction string `json:"direction"`

	// 范围分桶
	Ranges []MetricModelFormulaConfigBucketRange `json:"ranges"`

	// 条件分桶
	OtherBucket bool                                            `json:"other_bucket"`
	Filters     map[string]MetricModelFormulaConfigBucketFilter `json:"filters"`

	// 地理位置分桶
	Precision int `json:"precision"`
}

// 时间直方图分桶
type MetricModelFormulaConfigDateHistogram struct {
	Field         string `json:"field"`
	IntervalType  string `json:"interval_type"`
	IntervalValue string `json:"interval_value"`
}

type MetricModelFormulaConfigBucketRange struct {
	Key  string `json:"key"`
	From any    `json:"from"`
	To   any    `json:"to"`
}

type MetricModelFormulaConfigBucketFilter struct {
	QueryString string `json:"query_string"`
}

type MetricModelFormulaConfigAggregation struct {
	Type     string    `json:"type"`
	Field    string    `json:"field"`
	Percents []float64 `json:"percents"` // 仅用于百分位数
}

type MetricTask struct {
	TaskID             string   `json:"id"`
	TaskName           string   `json:"name"`
	ModelID            string   `json:"model_id"`
	MeasureName        string   `json:"measure_name"`
	Schedule           Schedule `json:"schedule"`
	TimeWindows        []string `json:"time_windows"`
	Steps              []string `json:"steps"`
	IndexBase          string   `json:"index_base"`
	IndexBaseName      string   `json:"index_base_name"`
	RetraceDuration    string   `json:"retrace_duration"`
	Comment            string   `json:"comment"`
	ScheduleSyncStatus int      `json:"schedule_sync_status"`
	ExeccuteStatus     int      `json:"execute_status"`
	UpdateTime         int64    `json:"update_time"`
	PlanTimes          []int64  `json:"plan_times"`
}

type Schedule struct {
	Type       string `json:"type"`
	Expression string `json:"expression"`
}

type AggInfo struct {
	AggName       string
	AggType       string
	TermsFields   []string
	IncludeFields []string
	IsDateField   bool
	TermsField    string         // terms 聚合的字段
	ConfigSize    int64          // 配置的terms分桶的大小
	Sort          string         // 排序字段
	Direction     string         // 排序方向
	EvalSize      int64          // 实际计算的terms分桶的大小
	SeriesTerms                  // 获取序列时,字段的批次信息
	IntervalType  string         // 计算公式date_histogram中的间隔类型:fixed_interval,calendar_interval
	IntervalValue string         // 计算公式中interval配置的值
	ZoneLocation  *time.Location // 根据模型中的时区参数转换的Loaction
}

type SeriesTerms struct {
	BatchSize     int64 // 获取序列时terms字段一批的大小
	BatchNum      int64 // 获取序列时terms字段的批次数量
	LastBatchSize int64 // 最后一批查询的字段的size
}

type TopHits struct {
	Size   int         `json:"size"`
	Sort   interface{} `json:"sort"`
	Source Source      `json:"_source"`
}

type Source struct {
	Includes []string `json:"includes"`
}

type DslInfo struct {
	AggInfos          map[int]AggInfo // 从dsl中解析出来的聚合信息,key为聚合的顺序序号,从0开始
	RangeQueryDSL     map[string]any  // 范围查询的dsl语句的map格式(aggs中含有date_histogram)
	InstantQueryDSL   map[string]any  // 即时查询的dsl语句的map格式(aggs中不含有date_histogram)
	DSLQuery          []byte          // dsl语句的query的bytes。 todo:对于dsl的结构使用的是map，后续合并到dslConfig时考虑dsl的map结构的修改，或者考虑json_raw_data
	TermsInfos        []AggInfo       // terms聚合的信息,key为terms的序号
	TermsToAggs       []int           // 记录terms在原有dsl aggs中的位置
	BucketSeriesNum   int64           // terms的目标序列个数
	NotTermsSeriesNum int64           // filters,range,date_range所占的序列大小
	DateHistogram     AggInfo         // 计算公式中配置的date_histogram信息
}

type UniResponseError struct {
	StatusCode int `json:"status_code"`
	rest.BaseError
}

func IsValidMetricType(m string) bool {
	return m == ATOMIC_METRIC || m == DERIVED_METRIC || m == COMPOSITED_METRIC
}

func IsValidQueryType(m string) bool {
	return m == PROMQL || m == DSL || m == DSL_CONFIG || m == SQL
}

func IsPersistMetric(metricName string) bool {
	if regexMatch := MEASURE_REGEX.MatchString(metricName); !regexMatch {
		return false
	} else {
		return true
	}
}

func IsValidDataSourceType(m string) bool {
	return m == QueryType_SQL || m == QueryType_DSL || m == QueryType_IndexBase
}

func IsValidTimeGranularity(m string) bool {
	return m == METRICS_SAMEPERIOD_TIME_GRANULARITY_DAY || m == METRICS_SAMEPERIOD_TIME_GRANULARITY_MONTH ||
		m == METRICS_SAMEPERIOD_TIME_GRANULARITY_QUARTER || m == METRICS_SAMEPERIOD_TIME_GRANULARITY_YEAR
}
