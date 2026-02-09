// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"

	dcond "data-model/interfaces/condition"
	dtype "data-model/interfaces/data_type"
)

const (
	TRACE_MODEL_OBJECT_TYPE     string = "ID_AUDIT_TRACE_MODEL"
	TRACE_MODEL                 string = "trace model"
	UNIX_MILLIS                 string = "unix_millis"
	UNIX_MICROS                 string = "unix_micros"
	UNIX_NANOS                  string = "unix_nanos"
	MS                          string = "ms"
	US                          string = "us"
	NS                          string = "ns"
	RELATED_LOG_OPEN            uint8  = 1
	RELATED_LOG_CLOSE           uint8  = 0
	SOURCE_TYPE_DATA_VIEW       string = "data_view"
	SOURCE_TYPE_DATA_CONNECTION string = "data_connection"
	QUERY_CATEGORY_SPAN         string = "span"
	QUERY_CATEGORY_RELATED_LOG  string = "related_log"
)

var (
	TRACE_MODEL_SORT = map[string]string{
		"name":        "f_model_name",
		"update_time": "f_update_time",
	}

	VALID_PRECONDITION_OPERATIONS        = []string{dcond.OperationAnd, dcond.OperationOr, dcond.OperationEq, dcond.OperationNotEq}
	VALID_TIME_FORMATS                   = []string{UNIX_MILLIS, UNIX_MICROS, UNIX_NANOS}
	VALID_DURATION_UNITS                 = []string{MS, US, NS}
	VALID_FIELD_TYPES_FOR_TRACE_ID       = []string{dtype.DataType_Text, dtype.DataType_String}
	VALID_FIELD_TYPES_FOR_SPAN_ID        = []string{dtype.DataType_Text, dtype.DataType_String, dtype.DataType_Integer, dtype.DataType_Integer}
	VALID_FIELD_TYPES_FOR_PARENT_SPAN_ID = []string{dtype.DataType_Text, dtype.DataType_String, dtype.DataType_Integer, dtype.DataType_Integer}
	VALID_FIELD_TYPES_FOR_NAME           = []string{dtype.DataType_Text, dtype.DataType_String}
	VALID_FIELD_TYPES_FOR_START_TIME     = []string{dtype.DataType_Integer, dtype.DataType_Integer}
	VALID_FIELD_TYPES_FOR_END_TIME       = []string{dtype.DataType_Integer, dtype.DataType_Integer}
	VALID_FIELD_TYPES_FOR_DURATION       = []string{dtype.DataType_Integer, dtype.DataType_Integer, dtype.DataType_Float, dtype.DataType_Float}
	VALID_FIELD_TYPES_FOR_KIND           = []string{dtype.DataType_Text, dtype.DataType_String}
	VALID_FIELD_TYPES_FOR_STATUS         = []string{dtype.DataType_Text, dtype.DataType_String}
	VALID_FIELD_TYPES_FOR_SERVICE_NAME   = []string{dtype.DataType_Text, dtype.DataType_String}
	VALID_PRECONDITION_VALUE_FROM        = []string{dcond.ValueFrom_Const, dcond.ValueFrom_Field}
)

// 链路模型列表项
type TraceModelListEntry struct {
	ModelID        string      `json:"id"`
	ModelName      string      `json:"name"`
	SpanSourceType string      `json:"span_source_type"`
	Tags           []string    `json:"tags"`
	Comment        string      `json:"comment"`
	CreateTime     int64       `json:"create_time"`
	UpdateTime     int64       `json:"update_time"`
	Creator        AccountInfo `json:"creator"`

	// 操作权限
	Operations []string `json:"operations"`
}

// 链路模型列表查询参数结构体
type TraceModelListQueryParams struct {
	SpanSourceTypes []string
	CommonListQueryParams
}

// 链路模型结构体
type TraceModel struct {
	ID                    string      `json:"id"`
	Name                  string      `json:"name"`
	Tags                  []string    `json:"tags"`
	TagsStr               string      `json:"-"`
	Comment               string      `json:"comment"`
	CreateTime            int64       `json:"create_time"`
	UpdateTime            int64       `json:"update_time"`
	SpanSourceType        string      `json:"span_source_type"`
	SpanConfig            any         `json:"span_config"`
	SpanConfigBytes       []byte      `json:"-"`
	EnabledRelatedLog     uint8       `json:"enabled_related_log"`
	RelatedLogSourceType  string      `json:"related_log_source_type"`
	RelatedLogConfig      any         `json:"related_log_config"`
	RelatedLogConfigBytes []byte      `json:"-"`
	Creator               AccountInfo `json:"creator"`

	Operations []string `json:"operations"`
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
	Precond    *CondCfg `json:"condition,omitempty"`
	FieldNames []string `json:"field_names"`
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

type TraceModelField struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type TraceModelFieldInfo struct {
	Span       []TraceModelField `json:"span"`
	RelatedLog []TraceModelField `json:"related_log"`
}

//go:generate mockgen -source ../interfaces/trace_model_service.go -destination ../interfaces/mock/mock_trace_model_service.go
type TraceModelService interface {
	CreateTraceModels(ctx context.Context, models []TraceModel) ([]string, error)
	SimulateCreateTraceModel(ctx context.Context, model TraceModel) (TraceModel, error)
	DeleteTraceModels(ctx context.Context, modelIDs []string) error
	UpdateTraceModel(ctx context.Context, model TraceModel) error
	SimulateUpdateTraceModel(ctx context.Context, model TraceModel) (TraceModel, error)
	GetTraceModels(ctx context.Context, modelIDs []string) ([]TraceModel, error)
	ListTraceModels(ctx context.Context, queryParams TraceModelListQueryParams) ([]TraceModelListEntry, int, error)
	GetTraceModelFieldInfo(ctx context.Context, modelID string) (TraceModelFieldInfo, error)

	GetSimpleTraceModelMapByIDs(ctx context.Context, modelIDs []string) (map[string]TraceModel, error)
	GetSimpleTraceModelMapByNames(ctx context.Context, modelNames []string) (map[string]TraceModel, error)

	ListTraceModelSrcs(ctx context.Context, queryParams TraceModelListQueryParams) ([]Resource, int, error)
}
