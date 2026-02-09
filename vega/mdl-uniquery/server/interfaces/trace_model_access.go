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
	RELATED_LOG_OPEN                uint8  = 1
	RELATED_LOG_CLOSE               uint8  = 0
	SOURCE_TYPE_DATA_VIEW           string = "data_view"
	SOURCE_TYPE_DATA_CONNECTION     string = "data_connection"
	QUERY_CATEGORY_SPAN             string = "span"
	QUERY_CATEGORY_RELATED_LOG      string = "related_log"
	QUERY_CATEGORY_SPAN_LIST        string = "span_list"
	QUERY_CATEGORY_RELATED_LOG_LIST string = "related_log_list"
)

// 链路模型结构体
type TraceModel struct {
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	Tags                 []string `json:"tags"`
	Comment              string   `json:"comment"`
	CreateTime           int64    `json:"create_time"`
	UpdateTime           int64    `json:"update_time"`
	SpanSourceType       string   `json:"span_source_type"`
	SpanConfig           any      `json:"span_config"`
	EnabledRelatedLog    uint8    `json:"enabled_related_log"`
	RelatedLogSourceType string   `json:"related_log_source_type"`
	RelatedLogConfig     any      `json:"related_log_config"`
}

// span配置
type SpanConfigWithDataView struct {
	DataView     DataViewConfig       `json:"data_view"`
	TraceID      TraceIDConfig        `json:"trace_id"`
	SpanID       SpanIDConfig         `json:"span_id"`
	ParentSpanID []ParentSpanIDConfig `json:"parent_span_id"`
	Name         NameConfig           `json:"name"`
	StartTime    StartTimeConfig      `json:"start_time"`
	EndTime      EndTimeConfig        `json:"end_time"`
	Duration     DurationConfig       `json:"duration"`
	Kind         KindConfig           `json:"kind"`
	Status       StatusConfig         `json:"status"`
	ServiceName  ServiceNameConfig    `json:"service_name"`
}

type SpanConfigWithDataConnection struct {
	DataConnection DataConnectionConfig `json:"data_connection"`
}

// span关联日志配置
type RelatedLogConfigWithDataView struct {
	DataView DataViewConfig `json:"data_view"`
	TraceID  TraceIDConfig  `json:"trace_id"`
	SpanID   SpanIDConfig   `json:"span_id"`
}

// 数据视图配置
type DataViewConfig struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// TraceID配置
type TraceIDConfig struct {
	FieldName string `json:"field_name"`
}

// SpanID配置
type SpanIDConfig struct {
	FieldNames []string `json:"field_names"`
}

// 父SpanID配置
type ParentSpanIDConfig struct {
	Precond    *cond.CondCfg `json:"condition"`
	FieldNames []string      `json:"field_names"`
}

// Span名称配置
type NameConfig struct {
	FieldName string `json:"field_name"`
}

// Span开始时间配置
type StartTimeConfig struct {
	FieldName   string `json:"field_name"`
	FieldFormat string `json:"field_format"`
}

// Span结束时间配置
type EndTimeConfig struct {
	FieldName   string `json:"field_name"`
	FieldFormat string `json:"field_format"`
}

// Span耗时配置
type DurationConfig struct {
	FieldName string `json:"field_name"`
	FieldUnit string `json:"field_unit"`
}

// Span类型配置
type KindConfig struct {
	FieldName string `json:"field_name"`
}

// Span状态配置
type StatusConfig struct {
	FieldName string `json:"field_name"`
}

// Span所处服务名称配置
type ServiceNameConfig struct {
	FieldName string `json:"field_name"`
}

// 数据连接配置
type DataConnectionConfig struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

//go:generate mockgen -source ../interfaces/trace_model_access.go -destination ../interfaces/mock/mock_trace_model_access.go
type TraceModelAccess interface {
	GetTraceModelByID(ctx context.Context, modelID string) (TraceModel, bool, error)
	SimulateCreateTraceModel(ctx context.Context, model TraceModel) (TraceModel, error)
	SimulateUpdateTraceModel(ctx context.Context, modelID string, model TraceModel) (TraceModel, error)
}
