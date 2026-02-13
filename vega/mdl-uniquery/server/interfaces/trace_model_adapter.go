// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
)

const (
	SOURCE_TYPE_TINGYUN string = "tingyun"
)

//go:generate mockgen -source ../interfaces/trace_model_adapter.go -destination ../interfaces/mock/mock_trace_model_adapter.go
type TraceModelAdapter interface {
	GetSpanList(ctx context.Context, model TraceModel, params SpanListQueryParams) ([]SpanListEntry, int64, error)
	GetSpan(ctx context.Context, model TraceModel, params SpanQueryParams) (SpanDetail, error)
	GetSpanMap(ctx context.Context, model TraceModel, params TraceQueryParams) (map[string]*BriefSpan_, map[string]SpanDetail, error)
	GetRelatedLogCountMap(ctx context.Context, model TraceModel, params TraceQueryParams) (map[string]int64, error)
	GetSpanRelatedLogList(ctx context.Context, model TraceModel, params RelatedLogListQueryParams) ([]RelatedLogListEntry, int64, error)
}
