// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"time"

	"uniquery/logics/promql/labels"
)

const (
	ALL_LABELS_FLAG              = "__all"
	LABELS_STR                   = "__labels_str"
	TSID                         = "__tsid"
	SAMPLING_AGG                 = "sampling"
	IRATE_AGG                    = "irate"
	RATE_AGG                     = "rate"
	INCREASE_AGG                 = "increase"
	CHANGES_AGG                  = "changes"
	AVG_OVER_TIME                = "avg_over_time"
	SUM_OVER_TIME                = "sum_over_time"
	MAX_OVER_TIME                = "max_over_time"
	MIN_OVER_TIME                = "min_over_time"
	COUNT_OVER_TIME              = "count_over_time"
	DELTA_AGG                    = "delta"
	CUMULATIVE_SUM               = "cumulative_sum"
	HISTOGRAM_QUANTILE           = "histogram_quantile"
	K_MINUTE_DOWNTIME            = "continuous_k_minute_downtime"
	DICT_LABELS                  = "dict_labels"
	DICT_VALUES                  = "dict_values"
	LABEL_JOIN                   = "label_join"
	LABEL_REPLACE                = "label_replace"
	FUNC_METRIC_MODEL            = "metric_model"
	KEYWORD_SUFFIX               = "keyword"
	FORM_URLENCODED_CONTENT_TYPE = "application/x-www-form-urlencoded"
	METRICS_PREFIX               = "metrics"
	LABELS_PREFIX                = "labels"
	DESENSITIZE_FIELD_SUFFIX     = "_desensitize"
	TEXT_TYPE                    = "text"
	DEFAULT_LOOK_BACK_DELTA      = time.Minute * time.Duration(5)
	SHARD_ROUTING_30M            = 30 * time.Minute
	SHARD_ROUTING_2H             = 2 * time.Hour
	DEFAULT_LOOK_BACK_DELTA_STR  = "5m"
	DEFAULT_STEP_DIVISOR         = 5 * time.Minute
	KMINUTE_DOWNTTIME_STEP       = 60000
)

type Query struct {
	QueryStr             string
	Start                int64
	End                  int64
	Interval             int64
	IntervalStr          string // 为了把指标模型请求的步长字符串往下传，不只传值
	FixedStart           int64
	FixedEnd             int64
	IsInstantQuery       bool // 用于标记instant query
	IsCalendar           bool
	SubIntervalWith30min int64
	SubIntervalWith2h    int64
	LogGroupId           string // 日志分组参数
	// DataViewId           string               // 数据视图id
	IsMetricModel    bool             // 是否是基于指标模型的查询
	LogGroup         LogGroup         // 日志分组过滤条件等信息
	DataView         DataView         // 对于指标模型的查询来说，dataview 已经在模型信息中返回了
	ViewQuery4Metric ViewQuery4Metric // 对于指标模型的查询来说，dataview 已经在模型信息中返回了
	Filters          []Filter         // 用于接受仪表盘的全局过滤器
	// LookBackDelta       int64                // 瞬时查询时从 time 往前回退的时间区间
	IsPersistMetric     bool         // 当前请求查询的指标是否是持久化指标
	Offset              int64        // 开始相应的序列的偏移量
	Limit               int64        // 每页最多可返回的序列数
	IgnoringHCTS        bool         // 是否忽略高基查询
	IgnoringMemoryCache bool         // 是否忽略内存缓存
	IfNeedAllSeries     bool         // 表达式是否需要查询全部数据，序列聚合、二元运算、histogram_quantile需要全部序列，默认是false
	ModelId             string       // 请求的指标模型id
	ModelUpdateTime     int64        // 模型更新时间
	MaxSearchSeriesSize int          // 向opensearch发起一次查询的最大序列数.当是计算公式有效性检查时,赋值1
	IsModelRequest      bool         // 是否是不需要再次校验dsl公式中的规则的请求，不需要是 true，此时无需校验 dsl 公式是否符合规则，预览时需要校验。
	NotNeedFilling      bool         // 叶子节点是否不需要补点,默认为false，需要补点（当前只针对sampling，其他叶子节点不补点）
	AnalysisDims        []string     // 分析维度，复合指标需要往下穿
	OrderByFields       []OrderField // 排序字段，复合指标的排序字段需要下沉

	VegaDurationMs int64 // 查询时vega的耗时，应用在复合指标的场景
}

type Matchers struct {
	MatcherSet [][]*labels.Matcher
	Start      int64
	End        int64
	LogGroupId string // 日志分组参数
}

// 为了兼容 grafana 的展示，封装返回结果。
type PromQLResponse struct {
	Status         status      `json:"status"`
	Data           interface{} `json:"data,omitempty"`
	SeriesTotal    int         `json:"-"`
	VegaDurationMs int64       `json:"-"` // 查询时vega的耗时，应用在复合指标的场景
}
type status string

//go:generate mockgen -source ../interfaces/promql_service.go -destination ../interfaces/mock/mock_promql_service.go
type PromQLService interface {
	Exec(ctx context.Context, query Query) (PromQLResponse, []byte, int, error)
	Series(matchers Matchers) ([]byte, int, error)

	GetFields(ctx context.Context, query Query) (map[string]bool, int, error)
	GetFieldValues(ctx context.Context, query Query, fieldName string) (map[string]bool, int, error)
	GetLabels(ctx context.Context, query Query) (map[string]bool, int, error)
}
