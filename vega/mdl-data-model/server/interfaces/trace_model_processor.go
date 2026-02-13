// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"

	dtype "data-model/interfaces/data_type"
)

var (
	SPAN_METADATA = []TraceModelField{
		{
			Name: "__duration",
			Type: dtype.DataType_Integer,
		},
		{
			Name: "__end_time",
			Type: dtype.DataType_Integer,
		},
		{
			Name: "__kind",
			Type: dtype.DataType_Text,
		},
		{
			Name: "__name",
			Type: dtype.DataType_Text,
		},
		{
			Name: "__parent_span_id",
			Type: dtype.DataType_Text,
		},
		{
			Name: "__service_name",
			Type: dtype.DataType_Text,
		},
		{
			Name: "__span_id",
			Type: dtype.DataType_Text,
		},
		{
			Name: "__start_time",
			Type: dtype.DataType_Integer,
		},
		{
			Name: "__status",
			Type: dtype.DataType_Text,
		},
		{
			Name: "__trace_id",
			Type: dtype.DataType_Text,
		},
	}
	RELATED_LOG_METADATA = []TraceModelField{
		{
			Name: "__span_id",
			Type: dtype.DataType_Text,
		},
		{
			Name: "__trace_id",
			Type: dtype.DataType_Text,
		},
	}
)

//go:generate mockgen -source ../interfaces/trace_model_processor.go -destination ../interfaces/mock/mock_trace_model_processor.go
type TraceModelProcessor interface {
	// 获取Span的字段信息
	GetSpanFieldInfo(ctx context.Context, model TraceModel) (fieldInfos []TraceModelField, err error)
	// 获取RelatedLog的字段信息
	GetRelatedLogFieldInfo(ctx context.Context, model TraceModel) (fieldInfos []TraceModelField, err error)
}
