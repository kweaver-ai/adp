package interfaces

import (
	"context"
)

const (
	VegaQueryType_Sync   = 0
	VegaQueryType_Stream = 1
	VegaQueryType_DSL    = 2

	// 单源直接下推查询类型
	VegaDataSourceQueryType_Sync   = 1
	VegaDataSourceQueryType_Stream = 2
)

type FetchVegaDataParams struct {
	IsSingleDataSource bool
	QueryType          string
	DataSourceID       string
	NextUri            string
	SqlStr             string
	UseSearchAfter     bool
	Limit              int
	Timeout            int64

	Dsl         DSLCfg
	TableNames  []string
	CatalogName string
}

//go:generate mockgen -source ../interfaces/vega_gateway_access.go -destination ../interfaces/mock/mock_vega_gateway_access.go
type VegaGatewayAccess interface {
	// FetchData(ctx context.Context, nextUri, sqlStr string, sync bool) (*FetchDataRes, error)
	FetchDataNoUnmarshal(ctx context.Context, params *FetchVegaDataParams) ([]byte, error)
	// GetMetadataViewFields(ctx context.Context, viewID string) ([]*cond.ViewField, error)
}

// // 流式查询
// type StreamingFetchData struct {
// 	Type      int    `json:"type"`
// 	Sql       string `json:"sql"`
// 	Timeout   int    `json:"timeout"`
// 	BatchSize int    `json:"batch_size"`
// }

// // 同步查询
// type SyncFetchData struct {
// 	Type int    `json:"type"`
// 	Sql  string `json:"sql"`
// }

// // DSL 查询
// type DslFetchData struct {
// 	Type        int      `json:"type"`
// 	CatalogName string   `json:"catalog_name"`
// 	TableName   []string `json:"table_name"`
// 	Dsl         DSLCfg   `json:"dsl"`
// }

type DataConnFetchDataRes struct {
	TotalCount int       `json:"total_count"`
	Columns    []*Column `json:"columns"`
	Data       [][]any   `json:"data"`
	NextUri    string    `json:"nextUri"`
}

type VegaGatewayProFetchDataRes struct {
	TotalCount int       `json:"total_count"`
	Columns    []*Column `json:"columns"`
	Entries    [][]any   `json:"entries"`
	NextUri    string    `json:"next_uri"`
}

type Column struct {
	Name string `json:"name"`
	Type string `json:"type"`
}
