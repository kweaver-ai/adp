// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"bytes"
	"context"
	"time"

	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/ast"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

const (
	ViewType_Atomic = "atomic"
	ViewType_Custom = "custom"

	QueryType_DSL       = "DSL"
	QueryType_SQL       = "SQL"
	QueryType_IndexBase = "IndexBase"

	DataScopeNodeType_View   = "view"
	DataScopeNodeType_Join   = "join"
	DataScopeNodeType_Union  = "union"
	DataScopeNodeType_Sql    = "sql"
	DataScopeNodeType_Output = "output"

	// join的类型
	JoinType_Inner     = "inner"
	JoinType_Left      = "left"
	JoinType_Right     = "right"
	JoinType_FullOuter = "full outer"

	// union的类型
	UnionType_All      = "all"
	UnionType_Distinct = "distinct"
)

// 字段范围
const (
	FieldScope_Partial = "partial"
	FieldScope_All     = "all"
)

const (
	INDEX_BASE         = "index_base"
	DEFAULT_OFFEST     = "0"
	DEFAULT_LIMIT      = 10
	DefaultSchema      = "default"
	SearchAfter_Limit  = 10000
	FILTERS_MAX_NUMBER = 5

	Format_Original = "original"
	Format_Flat     = "flat"

	Version_V1 = "v1"
	Version_V2 = "v2"

	MetaField_Timestamp = "@timestamp"
	MetaField_ID        = "__id"
	MetaField_WriteTime = "__write_time"

	All_Pits_DataView   = "__all"
	All_Pits_OpenSearch = "_all"

	Headers_MethodOverride = "x-http-method-override"

	QueryParam_AllowNonExistField = "allow_non_exist_field"
	QueryParam_IncludeView        = "include_view"
	QueryParam_Timeout            = "timeout"

	SearchError_SearchContextMissingException = "search_context_missing_exception"
)

var (
	REQUIRED_META_FIELDS = []string{
		"@timestamp",
		"__data_type",
		"__index_base",
		"__write_time",
		"__id",
		"__tsid",
		"__routing",
		"__category",
		"__pipeline_id",
		"tags",
	}

	JoinTypeMap = map[string]struct{}{
		JoinType_Inner:     {},
		JoinType_Left:      {},
		JoinType_Right:     {},
		JoinType_FullOuter: {},
	}

	UnionTypeMap = map[string]struct{}{
		UnionType_All:      {},
		UnionType_Distinct: {},
	}
)

// 视图查询外部接口统一返回结构
type ViewUniResponseV1 struct {
	ScrollId string     `json:"scroll_id,omitempty"`
	View     *DataView  `json:"view"`
	Datas    []ViewData `json:"datas"`
}

// 视图查询外部接口统一返回结构 V1 的 Data 结构
type ViewData struct {
	Total  *int64           `json:"total,omitempty"`
	Values []map[string]any `json:"values"`
}

// 视图查询外部接口统一返回结构 V2
type ViewUniResponseV2 struct {
	PitID          string           `json:"pit_id,omitempty"`
	SearchAfter    []any            `json:"search_after,omitempty"`
	View           *DataView        `json:"view,omitempty"`
	Entries        []map[string]any `json:"entries"`
	TotalCount     *int64           `json:"total_count,omitempty"`
	VegaDurationMs int64            `json:"vega_duration_ms,omitempty"`
	OverallMs      int64            `json:"overall_ms,omitempty"`

	// 提供给 v1 接口
	ScrollId string `json:"-"`
}

// 内部模块调用视图查询返回
type ViewInternalResponse struct {
	ScrollId string
	View     *DataView
	Total    int64
	Datas    []*ast.Node
}

// OpenSearch 每条文档信息
type Document struct {
	Id     string         `json:"_id"`
	Index  string         `json:"_index"`
	Source map[string]any `json:"_source"`
}

type DeletePits struct {
	PitIDs []string `json:"pit_ids" binding:"required"`
}

type DeletePitsResp = opensearchapi.PointInTimeDeleteResp

// OpenSearch DSL 结构体
type DSLCfg struct {
	From           int              `json:"from"`
	Size           int              `json:"size"`
	Sort           []map[string]any `json:"sort,omitempty"`
	TrackScores    bool             `json:"track_scores,omitempty"`
	TrackTotalHits bool             `json:"track_total_hits,omitempty"`
	SearchAfter    []any            `json:"search_after,omitempty"`
	Query          struct {
		Bool struct {
			Should         []any `json:"should,omitempty"`
			Filter         []any `json:"filter,omitempty"`
			Must           []any `json:"must,omitempty"`
			MinShouldMatch int   `json:"minimum_should_match,omitempty"`
		} `json:"bool"`
	} `json:"query"`
	Pit *struct {
		ID        string `json:"id,omitempty"`
		KeepAlive string `json:"keep_alive,omitempty"`
	} `json:"pit,omitempty"`
}

func (dsl DSLCfg) String() string {
	bytes, _ := sonic.MarshalIndent(dsl, "", "  ")
	return string(bytes)
}

// 视图的查询信息-给指标模型使用
type ViewQuery4Metric struct {
	BaseTypes   []string       `json:"-"` // 视图对应的索引库类型
	IndexShards []*IndexShards `json:"-"` // 视图对应的索引列表的索引分片信息
	Indices     []string       `json:"-"` // 视图对应的索引列表
	QueryStr    string         `json:"-"` // 视图本身构造的过滤条件，sql或dsl
	Catalog     string         `json:"-"` // 外置opensearch视图对应的数据源catalog
}

//go:generate mockgen -source ../interfaces/data_view_service.go -destination ../interfaces/mock/mock_data_view_service.go
type DataViewService interface {
	// 视图 handler 层调用
	Simulate(ctx context.Context, query *DataViewSimulateQuery) (*ViewUniResponseV2, error)
	GetSingleViewData(ctx context.Context, viewID string, query ViewQueryInterface) (*ViewUniResponseV2, error)
	DeleteDataViewPits(ctx context.Context, pits *DeletePits) (*DeletePitsResp, error)

	// 服务内部调用，视图提供给内部模块的方法返回的error是httpErr
	// GetDataViewIDByName(ctx context.Context, viewName string) (string, error)
	RetrieveSingleViewData(ctx context.Context, viewID string, query *DataViewQueryV1) (*ViewInternalResponse, error)
	CountMultiFields(ctx context.Context, viewID string, query *DataViewQueryV1, fields []string, sep string) (map[string]int64, error)
	LoadIndexShards(ctx context.Context, indices string) ([]byte, int, error)
	GetIndices(ctx context.Context, baseTypes []string, start int64, end int64) ([]*IndexShards, []string, int, error)
	GetDataViewByID(ctx context.Context, viewID string, includeDataScopeView bool) (*DataView, error)
	GetDataFromOpenSearch(ctx context.Context, query map[string]any, indices []string,
		scroll time.Duration, preference string, trackTotalHits bool) ([]byte, int, error)
	GetDataFromOpenSearchWithBuffer(ctx context.Context, query bytes.Buffer, indices []string,
		scroll time.Duration, preference string) ([]byte, int, error)

	BuildViewQuery4MetricModel(ctx context.Context, start, end int64, view *DataView) (ViewQuery4Metric, error)
}
