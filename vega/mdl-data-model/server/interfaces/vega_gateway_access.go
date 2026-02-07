// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

//go:generate mockgen -source ../interfaces/vega_gateway_access.go -destination ../interfaces/mock/mock_vega_gateway_access.go
type VegaGatewayAccess interface {
	// CreateVegaExcelView(ctx context.Context, req *CreateVegaExcelViewReq) (*CreateVegaExcelViewRes, error) // 新增 excel视图
	// DeleteVegaExcelView(ctx context.Context, req *DeleteExcelViewReq) error                                //删除Excel视图

	// FetchData(ctx context.Context, statement string) (*FetchDataRes, error)
	// FetchAuthorizedData(ctx context.Context, statement string, req *FetchReq) (*FetchDataRes, error)
	// GetConnectors(ctx context.Context) (*GetConnectorsRes, error)

	// CreateExcelView(ctx context.Context, req *CreateExcelViewReq) (*CreateExcelViewRes, error) //新增Excel视图
	// DeleteExcelView(ctx context.Context, req *DeleteExcelViewReq) (*DeleteExcelViewRes, error) //删除Excel视图
	// GetPreview(ctx context.Context, req *ViewEntries) (*FetchDataRes, error)
}

type GetViewReq struct {
	CatalogName string `json:"catalogName"`
	SchemaName  string `json:"schemaName"`
	ViewName    string `json:"viewName"`
}
type GetViewRes struct {
	Total   int            `json:"total"`
	Pages   int            `json:"pages"`
	Entries []*ViewEntries `json:"entries"`
}
type ViewEntries struct {
	CatalogName string `json:"catalogName"`
	ViewName    string `json:"viewName"`
	Schema      string `json:"schema"`
	Limit       int    `json:"limit"`
	UserId      string `json:"user_id"`
}

type CreateViewReq struct {
	CatalogName string `json:"catalogName"` // 数据源catalog
	Query       string `json:"query"`
	ViewName    string `json:"viewName"`
}

type CreateVegaExcelViewReq struct {
	Catalog          string         `json:"catalog"`
	FileName         string         `json:"file_name"`
	TableName        string         `json:"table_name"`
	Columns          []*ExcelColumn `json:"columns"`
	StartCell        string         `json:"start_cell"`
	EndCell          string         `json:"end_cell"`
	Sheet            string         `json:"sheet"`
	AllSheet         bool           `json:"all_sheet"`
	SheetAsNewColumn bool           `json:"sheet_as_new_column"`
	HasHeaders       bool           `json:"has_headers"`
	// VDMCatalog       string         `json:"vdm_catalog"`
}

type CreateVegaExcelViewRes struct {
	TableName string `json:"tableName"`
}

type ExcelColumn struct {
	Column string `json:"column"`
	Type   string `json:"type"`
}

//region DeleteView

type DeleteViewReq struct {
	CatalogName string `json:"catalogName"` // 数据源catalog
	ViewName    string `json:"viewName"`
}

//endregion

//region ModifyView

type ModifyViewReq struct {
	CatalogName string `json:"catalogName"` // 数据源catalog
	Query       string `json:"query"`
	ViewName    string `json:"viewName"`
}

//region DeleteDataSource

type FetchDataRes struct {
	TotalCount int       `json:"total_count"`
	Columns    []*Column `json:"columns"`
	Data       [][]any   `json:"data"`
}
type Column struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// endregion

// type StreamFetchResp struct {
// 	NextURI string `json:"nextUri"`
// 	FetchDataRes
// }

// type DownloadDataReq struct {
// 	Catalog  string `json:"catalog"`             // catalog名称
// 	Schema   string `json:"schema"`              // schema名称
// 	Table    string `json:"table"`               // 表名称
// 	Columns  string `json:"columns"`             // 要下载的列名称（多个列以;隔开）
// 	RowRules string `json:"row_rules,omitempty"` // 行限制条件（多个条件以;隔开）
// 	OrderBy  string `json:"order_by,omitempty"`  // 排序字句（多个排序子句以;隔开）如"a desc;b asc;c desc"
// 	Offset   uint64 `json:"offset"`              // 偏移量
// 	Limit    uint64 `json:"limit"`               // 下载行数限制
// 	Action   string `json:"action"`              // 动作类型
// }

//region  CreateExcelView 新增Excel视图

// type CreateExcelViewRes struct {
// 	TableName string `json:"tableName"`
// 	ViewName  string `json:"viewName"`
// }

//endregion

//region  DeleteExcelView 删除Excel视图

// type DeleteExcelViewReq struct {
// 	Catalog string `json:"catalog"`
// 	Schema  string `json:"schema"`
// 	View    string `json:"view"`
// }

// type DeleteExcelViewRes struct {
// 	ViewName string `json:"viewName"`
// }

//endregion

// type FetchReq struct {
// 	UserID string `json:"user_id"`
// 	Action string `json:"action"` // 动作类型
// }

//region GetDataTables

type GetDataTablesReq struct {
	Offset       int    `json:"offset"`
	Limit        int    `json:"limit"`
	Keyword      string `json:"keyword"`
	DataSourceId uint64 `json:"data_source_id"`
	SchemaId     string `json:"schema_id"`
	Ids          string `json:"ids"`
	Sort         string `json:"sort"`
	Direction    string `json:"direction"`
	CheckField   bool   `json:"checkField"`
}
type GetDataTablesRes struct {
	Code        string                  `json:"code"`
	Description string                  `json:"description"`
	Solution    string                  `json:"solution"`
	TotalCount  int                     `json:"total_count"`
	Data        []*GetDataTablesDataRes `json:"data"`
}
type GetDataTablesDataRes struct {
	DataSourceType     int            `json:"data_source_type"`
	DataSourceTypeName string         `json:"data_source_type_name"`
	DataSourceId       string         `json:"data_source_id"`
	DataSourceName     string         `json:"data_source_name"`
	SchemaId           string         `json:"schema_id"`
	SchemaName         string         `json:"schema_name"`
	Id                 string         `json:"id"`
	Name               string         `json:"name"`
	AdvancedParams     string         `json:"advanced_params"`
	AdvancedDataSlice  []AdvancedData `json:"advanced_data_slice"`
	CreateTime         string         `json:"create_time"`
	CreateTimeStamp    string         `json:"create_time_stamp"`
	UpdateTime         string         `json:"update_time"`
	UpdateTimeStamp    string         `json:"update_time_stamp"`
	TableRows          string         `json:"table_rows"`
	HaveField          bool           `json:"have_field"`
}

type AdvancedData struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

//endregion

//region GetDataTableDetail

type GetDataTableDetailReq struct {
	DataSourceId uint64 `json:"data_source_id"`
	SchemaId     string `json:"schema_id"`
	TableId      string `json:"table_id"`
}

// type GetDataTableDetailRes struct {
// 	Code        string                       `json:"code"`
// 	Description string                       `json:"description"`
// 	Solution    string                       `json:"solution"`
// 	Data        []*GetDataTableDetailDataRes `json:"data"`
// }
// type GetDataTableDetailDataRes struct {
// 	Name           string            `json:"name"`
// 	AdvancedParams []*AdvancedParams `json:"advanced_params"`
// 	Fields         []*Fields         `json:"fields"`
// }

type AdvancedParams struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

// type Fields struct {
// 	ID             string `json:"id"`
// 	FieldName      string `json:"field_name"`
// 	FieldLength    int    `json:"field_length"`
// 	FieldPrecision int    `json:"field_precision"`
// 	FieldComment   string `json:"field_comment"`
// 	AdvancedParams int    `json:"advanced_params"`
// 	FieldTypeName  string `json:"field_type_name"`
// }

//endregion
