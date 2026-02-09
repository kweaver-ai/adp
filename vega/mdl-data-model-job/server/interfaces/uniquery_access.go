// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
)

type MetricModelQuery struct {
	Time           int64  `json:"time"`            // 瞬时查询的时间
	LookBackDelta  string `json:"look_back_delta"` // 瞬时查询时从 time 往前回退的时间区间
	IsInstantQuery bool   `json:"instant"`         // 用于标记 instant query，默认为 false，即默认是 query_range
}

type TimeInterval struct {
	Interval int    `json:"interval"`
	Unit     string `json:"unit"`
}
type EventQuery struct {
	QueryType  string   `json:"query_type,omitempty"`
	Id         string   `json:"id,omitempty"`
	Start      int64    `json:"start,omitempty"`
	End        int64    `json:"end,omitempty"`
	SortKey    string   `json:"sort,omitempty"`
	Limit      int64    `json:"limit,omitempty"`
	Offset     int64    `json:"offset,omitempty"`
	Direction  string   `json:"direction,omitempty"`
	Filters    []Filter `json:"filters" form:"filters"`
	Extraction []Filter `json:"extraction" form:"extraction" binding:"omitempty"`
}

type EventModelQueryRequest struct {
	Querys    []EventQuery `json:"querys" `
	SortKey   string       `json:"sort,omitempty" `
	Limit     int64        `json:"limit,omitempty" `
	Offset    int64        `json:"offset,omitempty" `
	Direction string       `json:"direction,omitempty" `
}

type Records []map[string]any

// Uniquery查询返回结果

type PreOrderEvent struct {
	Level int    `json:"level" form:"level"`
	Id    string `json:"id" form:"id"`
}

type Context struct {
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
	// AnalysisStartTime time.Time     `json:"analysis_start_time,omitempty"`
	// AnalysisEndTime   time.Time     `json:"analysis_end_time,omitempty"`
	// RunId             string        `json:"run_id,omitempty" form:"run_id"`
}
type EventModelData struct {
	Id                string            `json:"id"`
	Title             string            `json:"title"`
	EventModelId      string            `json:"event_model_id"`
	EventType         string            `json:"type"`
	Level             int               `json:"level"`
	GenerateType      string            `json:"generate_type"`
	LevelName         string            `json:"level_name,omitempty"`
	Message           string            `json:"message"`
	Context           Context           `json:"context"`
	TriggerTime       int64             `json:"trigger_time"`
	TriggerData       Records           `json:"trigger_data"`
	Tags              []string          `json:"tags"`
	EventModelName    string            `json:"event_model_name"`
	CreateTime        int64             `json:"@timestamp"`
	DataSource        []string          `json:"data_source"`
	DataSourceName    []string          `json:"data_source_name"`
	DataSourceType    string            `json:"data_source_type"`
	DefaultTimeWindow TimeInterval      `json:"default_time_window"`
	Schedule          Schedule          `json:"schedule"`
	Labels            map[string]string `json:"labels"`
	Relations         map[string]any    `json:"relations" form:"relations"`
	AggregateType     string            `json:"aggregate_type"`
	AggregateAlgo     string            `json:"aggregate_algo"`
	DetectType        string            `json:"detect_type"`
	DetectAlgo        string            `json:"detect_algo"`
	// LLMApp            string            `json:"llm_app,omitempty"`
	// Ekn               string            `json:"ekn,omitempty"`
}

type EventModelResponse struct {
	Entries    []EventModelData `json:"entries"`
	TotalCount int              `json:"total_count"`
	Entities   []Records        `json:"entities"`
}

type UniResponse struct {
	Datas      []Data `json:"datas"`
	Step       string `json:"step"`
	IsVariable bool   `json:"is_variable"`
	IsCalendar bool   `json:"is_calendar"`
}
type Data struct {
	Labels map[string]string `json:"labels"`
	Times  []interface{}     `json:"times"`
	Values []interface{}     `json:"values"`
}

//go:generate mockgen -source ../interfaces/uniquery_access.go -destination ../interfaces/mock/mock_uniquery_access.go
type UniqueryAccess interface {
	GetMetricModelData(ctx context.Context, modelId string, query MetricModelQuery) (UniResponse, error)
	GetEventModelData(ctx context.Context, query EventModelQueryRequest) (EventModelResponse, error)
	GetObjectiveModelData(ctx context.Context, modelId string, query MetricModelQuery) (ObjectiveModelUniResponse, error)
}
