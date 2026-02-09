// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	cond "uniquery/common/condition"
)

const (
	DEFAULT_SORT = "@timestamp"

	MIN_OFFSET_NUM_OF_LIST = 0
	MIN_LIMIT_NUM_OF_LIST  = 1
	MAX_LIMIT_NUM_OF_LIST  = 1000
	// NO_LIMIT_NUM_OF_LIST      = -1
	DEFAULT_LIMIT_NUM_OF_LIST = 10

	DEFAULT_OFFSET_STR_OF_LIST = "0"
	DEFAULT_LIMIT_STR_OF_LIST  = "10"

	// DEFAULT_SORT_OF_SPAN_LIST        = "@timestamp"
	// DEFAULT_SORT_OF_RELATED_LOG_LIST = "@timestamp"

	UNIX_MILLIS = "unix_millis"
	UNIX_MICROS = "unix_micros"
	UNIX_NANOS  = "unix_nanos"
	MS          = "ms"
	US          = "us"
	NS          = "ns"

	SPAN_KIND_UNSPECIFIED = "unspecified"
	SPAN_KIND_INTERNAL    = "internal"
	SPAN_KIND_SERVER      = "server"
	SPAN_KIND_CLIENT      = "client"
	SPAN_KIND_PRODUCER    = "producer"
	SPAN_KIND_CONSUMER    = "consumer"

	SPAN_STATUS_OK    = "ok"
	SPAN_STATUS_ERROR = "error"
	SPAN_STATUS_UNSET = "unset"
	DEFAULT_SEPARATOR = "$_$"
)

var (
	// DEFAULT_LIST_SORT = map[string]struct{}{
	// 	"@timestamp": {},
	// }

	SPAN_LIST_SORT = map[string]string{
		"@timestamp":   DEFAULT_SORT,
		"__start_time": DEFAULT_SORT,
	}

	RELATED_LOG_LIST_SORT = map[string]string{
		"@timestamp": DEFAULT_SORT,
	}

	SPAN_KIND_MAP = map[string]string{
		"unspecified": SPAN_KIND_UNSPECIFIED,
		"internal":    SPAN_KIND_INTERNAL,
		"server":      SPAN_KIND_SERVER,
		"client":      SPAN_KIND_CLIENT,
		"producer":    SPAN_KIND_PRODUCER,
		"consumer":    SPAN_KIND_CONSUMER,
	}

	SPAN_STATUS_MAP = map[string]string{
		"ok":    SPAN_STATUS_OK,
		"error": SPAN_STATUS_ERROR,
		"unset": SPAN_STATUS_UNSET,
	}
)

type SpanListPreviewParams struct {
	SpanListQueryParams
	TraceModel `json:"trace_model"`
}

type SpanListQueryParams struct {
	TraceID   string        `json:"-"`
	Condition *cond.CondCfg `json:"condition"`
	PaginationQueryParams
}

type TracePreviewParams struct {
	TraceQueryParams
	TraceModel `json:"trace_model"`
}

type TraceQueryParams struct {
	TraceID string         `json:"-"`
	Context map[string]any `json:"context"`
}

type SpanPreviewParams struct {
	TraceModel `json:"trace_model"`
}

type SpanQueryParams struct {
	TraceID string `json:"-"`
	SpanID  string `json:"-"`
}

type RelatedLogListQueryParams struct {
	TraceID   string        `json:"-"`
	SpanID    string        `json:"-"`
	Condition *cond.CondCfg `json:"condition"`
	PaginationQueryParams
}

// 分页查询参数
type PaginationQueryParams struct {
	Offset    int    `json:"offset"`
	Limit     int    `json:"limit"`
	Sort      string `json:"sort"`
	Direction string `json:"direction"`
}

type RelatedLogListPreviewParams struct {
	RelatedLogListQueryParams
	TraceModel `json:"trace_model"`
}

type AbstractSpan struct {
	TraceID      string `json:"__trace_id"`
	SpanID       string `json:"__span_id"`
	ParentSpanID string `json:"__parent_span_id"`
	Name         string `json:"__name"`
	StartTime    int64  `json:"__start_time"`
	EndTime      int64  `json:"__end_time"`
	Duration     int64  `json:"__duration"`
	Kind         string `json:"__kind"`
	Status       string `json:"__status"`
	ServiceName  string `json:"__service_name"`
}

type SpanListEntry map[string]any

// type SpanListEntry struct {
// 	AbstractSpan
// 	RawData `json:",inline"`
// }

type SpanDetail SpanListEntry

// trace详情返回格式
type TraceDetail_ struct {
	TraceID     string                `json:"trace_id"`
	StartTime   int64                 `json:"start_time"`
	EndTime     int64                 `json:"end_time"`
	Duration    int64                 `json:"duration"`
	StatusStats map[string]int64      `json:"status_stats"`
	Depth       int64                 `json:"depth"`
	Detail      *BriefSpan_           `json:"detail"`
	Spans       map[string]SpanDetail `json:"spans"`
	Services    []string              `json:"services"`
}

// type TraceDetailQueryParams struct {
// 	Mu    *sync.Mutex
// 	Wg    *sync.WaitGroup
// 	ErrCh chan error

// 	TraceModel TraceModel
// 	// TraceStats TraceStats
// }

type TraceStatisticData struct {
	StartTime   int64
	EndTime     int64
	Duration    int64
	Depth       int64
	StatusStats map[string]int64
	Services    []string
}

type BriefSpan_ struct {
	Key             string        `json:"key"`
	Name            string        `json:"__name"`
	SpanID          string        `json:"__span_id"`
	ParentSpanID    string        `json:"-"`
	StartTime       int64         `json:"__start_time"`
	EndTime         int64         `json:"__end_time"`
	Duration        int64         `json:"__duration"`
	Kind            string        `json:"__kind"`
	Status          string        `json:"__status"`
	ServiceName     string        `json:"__service_name"`
	RelatedLogCount int64         `json:"related_log_count"`
	Children        []*BriefSpan_ `json:"children"`
}

type AbstractRelatedLog struct {
	TraceID string `json:"__trace_id"`
	SpanID  string `json:"__span_id"`
}

type RelatedLogListEntry map[string]any

// type RelatedLogListEntry struct {
// 	AbstractRelatedLog `json:"model_data"`
// 	RawData            map[string]interface{} `json:"raw_data"`
// }

//go:generate mockgen -source ../interfaces/trace_model_service.go -destination ../interfaces/mock/mock_trace_model_service.go
type TraceModelService interface {
	GetSpanList(ctx context.Context, model TraceModel, params SpanListQueryParams) ([]SpanListEntry, int64, error)
	GetTrace(ctx context.Context, model TraceModel, params TraceQueryParams) (TraceDetail_, error)
	GetSpan(ctx context.Context, model TraceModel, params SpanQueryParams) (SpanDetail, error)
	GetSpanRelatedLogList(ctx context.Context, model TraceModel, params RelatedLogListQueryParams) ([]RelatedLogListEntry, int64, error)

	GetTraceModelByID(ctx context.Context, modelID string) (TraceModel, error)
	SimulateCreateTraceModel(ctx context.Context, model TraceModel) (TraceModel, error)
	SimulateUpdateTraceModel(ctx context.Context, modelID string, model TraceModel) (TraceModel, error)
}
