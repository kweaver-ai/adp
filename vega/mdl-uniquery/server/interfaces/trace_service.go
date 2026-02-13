// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"time"
)

const (
	MIN_OFFSET            = 0
	DEFAULT_OFFSET_STR    = "0"
	MIN_LIMIT             = 1
	MAX_LIMIT             = 1000
	DEFAULT_LIMIT_STR     = "10"
	DEFAULT_SPAN_STATUSES = "Ok,Error,Unset"

	UNIX_MILLISECOND_TIMESTAMP_STR_LENGTH = 13

	// 最大查询长度设置为10000
	MAX_SEARCH_SIZE = 10000

	// 默认Scroll保留时间为1分钟
	DEFAULT_SEARCH_SCROLL_DURATION = time.Minute
	DEFAULT_SEARCH_SCROLL_STR      = "1m"

	// 查询单条trace的关联日志时的最大分桶数
	MAX_SEARCH_RELATED_LOGS_BUCKET = 65000
)

// SPANKIND_MAPPING_TABLE为SpanKind数字与字母映射表
var SPANKIND_MAPPING_TABLE = map[int]string{
	0: "UNSPECIFIED",
	1: "INTERNAL",
	2: "SERVER",
	3: "CLIENT",
	4: "PRODUCER",
	5: "CONSUMER",
}

// span列表查询参数
type SpanListQuery struct {
	DataViewId    string
	Offset        int
	Limit         int
	StartTime     int64
	EndTime       int64
	SpanStatusMap map[string]int
}

// span列表返回格式
type SpanList struct {
	Name        string      `json:"name"`
	SpanID      string      `json:"span_id"`
	TraceID     string      `json:"trace_id"`
	SpanKind    interface{} `json:"span_kind"`
	SpanStatus  string      `json:"span_status"`
	StartTime   interface{} `json:"start_time"`
	Duration    int         `json:"duration"`
	ServiceName string      `json:"service_name"`
}

type SpanRelatedLogStats map[string]int64

// trace详情返回格式
type TraceDetail struct {
	TraceID     string           `json:"trace_id"`
	TraceStatus string           `json:"trace_status"`
	StartTime   int64            `json:"start_time"`
	EndTime     int64            `json:"end_time"`
	Duration    int64            `json:"duration"`
	SpanStats   map[string]int32 `json:"span_stats"`
	Depth       int32            `json:"depth"`
	Spans       *BriefSpan       `json:"spans"`
	// SpanMap: spanID与span的映射表, 不转换成json
	SpanMap  map[string]*BriefSpan `json:"-"`
	Services []Service             `json:"services"`
	// ServiceStats: span所属服务统计, 不转换成json
	ServiceStats map[Service]int32 `json:"-"`
}

type Span struct {
	Name         string                 `json:"Name"`
	SpanContext  SpanContext            `json:"SpanContext"`
	Parent       SpanContext            `json:"Parent"`
	SpanKind     int32                  `json:"SpanKind"`
	SpanKindDesc string                 `json:"SpanKindDesc"`
	StartTime    int64                  `json:"StartTime"`
	EndTime      int64                  `json:"EndTime"`
	Timestamp    string                 `json:"@timestamp"`
	Duration     int64                  `json:"Duration"`
	Attributes   map[string]interface{} `json:"Attributes"`
	Events       []Event                `json:"Events"`
	Links        []Link                 `json:"Links"`
	Status       Status                 `json:"Status"`
	Resource     map[string]interface{} `json:"Resource"`
}

type BriefSpan struct {
	Key             string           `json:"key"`
	Name            interface{}      `json:"Name"`
	SpanContext     BriefSpanContext `json:"SpanContext"`
	Parent          BriefSpanContext `json:"Parent"`
	SpanKindDesc    interface{}      `json:"SpanKindDesc"`
	StartTime       int64            `json:"StartTime"`
	EndTime         int64            `json:"EndTime"`
	Duration        int64            `json:"Duration"`
	Status          BriefStatus      `json:"Status"`
	Resource        BriefResource    `json:"Resource"`
	RelatedLogCount int64            `json:"RelatedLogCount"`
	Children        []*BriefSpan     `json:"children"`
}

type BriefSpanContext struct {
	SpanID string `json:"SpanID"`
}

type BriefStatus struct {
	CodeDesc string `json:"CodeDesc"`
}

type BriefResource struct {
	Service Service `json:"service"`
}

type SpanContext struct {
	TraceID    string      `json:"TraceID"`
	SpanID     string      `json:"SpanID"`
	TraceFlags string      `json:"TraceFlags"`
	TraceState interface{} `json:"TraceState"`
}

type Event struct {
	Name       string                 `json:"Name"`
	Time       int64                  `json:"Time"`
	Attributes map[string]interface{} `json:"Attributes"`
}

type Link struct {
	TraceID    string                 `json:"TraceID"`
	SpanID     string                 `json:"SpanID"`
	TraceState interface{}            `json:"TraceState"`
	Attributes map[string]interface{} `json:"Attributes"`
}

type Status struct {
	Code        int    `json:"Code"`
	CodeDesc    string `json:"CodeDesc"`
	Description string `json:"Description"`
}

type Service struct {
	Name interface{} `json:"name"`
}

// type Bucket struct {
// 	Key      string `json:"key"`
// 	DocCount int    `json:"doc_count"`
// }

// type BriefScrollSearchResponse struct {
// 	ScrollID       string         `json:"_scroll_id"`
// 	BriefOuterHits BriefOuterHits `json:"hits"`
// }

// type BriefOuterHits struct {
// 	Total Total `json:"total"`
// }

type ScrollSearchResponse struct {
	// ScrollID  string    `json:"_scroll_id"`
	OuterHits OuterHits `json:"hits"`
}

type OuterHits struct {
	// Total     Total `json:"total"`
	InnerHits []Hit `json:"hits"`
}

// type Total struct {
// 	Value int32 `json:"value"`
// }

type Hit struct {
	Fields map[string]interface{} `json:"fields"`
}

type LogAggResponse struct {
	Aggs LogAgg `json:"aggregations"`
}

type LogAgg struct {
	GroupBy GroupBySpanID `json:"group_by_SpanID"`
}

type GroupBySpanID struct {
	Buckets []LogStatBucket `json:"buckets"`
}

type LogStatBucket struct {
	Key      string `json:"key"`
	DocCount int64  `json:"doc_count"`
}

//go:generate mockgen -source ../interfaces/trace_service.go -destination ../interfaces/mock/mock_trace_service.go
type TraceService interface {
	// GetSpanList(traceListQuery SpanListQuery) ([]SpanList, int, error)
	GetTraceDetail(ctx context.Context, traceDataViewId, logDataViewId, traceId string) (*TraceDetail, error)
}
