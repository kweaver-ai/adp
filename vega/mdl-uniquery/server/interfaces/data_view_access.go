package interfaces

import (
	"context"
	cond "uniquery/common/condition"
)

var Headers = map[string]string{
	"Content-Type": "application/json",
}

type DataView struct {
	ViewID            string                     `json:"id" mapstructure:"id"`
	ViewName          string                     `json:"name" mapstructure:"name"`
	TechnicalName     string                     `json:"technical_name" mapstructure:"technical_name"`
	GroupID           string                     `json:"group_id" mapstructure:"group_id"`
	GroupName         string                     `json:"group_name" mapstructure:"group_name"`
	Type              string                     `json:"type" binding:"required,oneof=atomic custom" mapstructure:"type"`
	QueryType         string                     `json:"query_type" binding:"required,oneof=SQL DSL" mapstructure:"query_type"`
	Tags              []string                   `json:"tags" mapstructure:"tags"`
	Comment           string                     `json:"comment" mapstructure:"comment"`
	Builtin           bool                       `json:"builtin" mapstructure:"builtin"`
	CreateTime        int64                      `json:"create_time" mapstructure:"create_time"`
	UpdateTime        int64                      `json:"update_time" mapstructure:"update_time"`
	DataSourceType    string                     `json:"data_source_type,omitempty" mapstructure:"data_source_type"`
	DataSourceID      string                     `json:"data_source_id,omitempty" mapstructure:"data_source_id"`
	DataSourceName    string                     `json:"data_source_name,omitempty" mapstructure:"data_source_name"`
	DataSourceCatalog string                     `json:"data_source_catalog,omitempty" mapstructure:"data_source_catalog"`
	FileName          string                     `json:"file_name,omitempty" mapstructure:"file_name"`
	Status            string                     `json:"status,omitempty" mapstructure:"status"`
	Operations        []string                   `json:"operations" mapstructure:"operations"`
	Fields            []*cond.ViewField          `json:"fields" mapstructure:"fields"`
	FieldScope        string                     `json:"field_scope" mapstructure:"field_scope"`
	FieldsMap         map[string]*cond.ViewField `json:"fields_map" mapstructure:"fields_map"`
	ModuleType        string                     `json:"module_type" mapstructure:"module_type"`
	Creator           AccountInfo                `json:"creator" mapstructure:"creator"`
	Updater           AccountInfo                `json:"updater" mapstructure:"updater"`
	DataScope         []*DataScopeNode           `json:"data_scope,omitempty" mapstructure:"data_scope"`
	ExcelConfig       *ExcelConfig               `json:"excel_config,omitempty" mapstructure:"excel_config"`
	MetadataFormID    string                     `json:"metadata_form_id,omitempty" mapstructure:"metadata_form_id"`
	PrimaryKeys       []string                   `json:"primary_keys" mapstructure:"primary_keys"`
	SQLStr            string                     `json:"sql_str,omitempty" mapstructure:"sql_str"`
	MetaTableName     string                     `json:"meta_table_name,omitempty" mapstructure:"meta_table_name"`
	DataScopeAdvancedParams
}

// 简单的视图结构，列表查询接口使用
type DataScopeAdvancedParams struct {
	HasDataScopeSQLNode   bool   `json:"-"` // 是否包含sql节点
	HasStar               bool   `json:"-"` // 是否有 *
	IsSingleSource        bool   `json:"-"` // 自定义视图的原子视图是否来自同一个数据源
	DataScopeDataSourceID string `json:"-"` // 自定义视图数据源ID
}

// DataScopeNode 表示数据作用域图中的节点
type DataScopeNode struct {
	ID              string                     `json:"id"`
	Title           string                     `json:"title"`
	Type            string                     `json:"type"`
	InputNodes      []string                   `json:"input_nodes"`
	Config          map[string]any             `json:"config"`
	OutputFields    []*cond.ViewField          `json:"output_fields"`
	OutputFieldsMap map[string]*cond.ViewField `json:"-"` // 存储输出字段列表（对应metadata全部字段）
}

// // 节点类型为view的节点配置
// type ViewNodeCfg struct {
// 	ViewID        string                     `json:"view_id" mapstructure:"view_id"`
// 	TechnicalName string                     `json:"technical_name" mapstructure:"technical_name"`
// 	Filters       *cond.CondCfg              `json:"filters,omitempty" mapstructure:"filters"`
// 	Distinct      Distinct                   `json:"distinct" mapstructure:"distinct"`
// 	MetaTableName string                     `json:"meta_table_name" mapstructure:"meta_table_name"`
// 	FieldsMap     map[string]*cond.ViewField `json:"fields_map" mapstructure:"fields_map"` // 存储原子视图的字段列表（对应metadata全部字段）
// }

// 节点类型为view的节点配置
type ViewNodeCfg struct {
	ViewID   string        `json:"view_id" mapstructure:"view_id"`
	Filters  *cond.CondCfg `json:"filters,omitempty" mapstructure:"filters"`
	Distinct Distinct      `json:"distinct" mapstructure:"distinct"`
	View     *DataView     `json:"view,omitempty" mapstructure:"view"`
}

type Distinct struct {
	Enable bool     `json:"enable" mapstructure:"enable"`
	Fields []string `json:"fields,omitempty" mapstructure:"fields"`
}

// 节点类型为join的节点配置
type JoinNodeCfg struct {
	JoinType string        `json:"join_type" mapstructure:"join_type"`
	JoinOn   []*JoinOn     `json:"join_on" mapstructure:"join_on"`
	Filters  *cond.CondCfg `json:"filters,omitempty" mapstructure:"filters"`
	Distinct Distinct      `json:"distinct,omitempty" mapstructure:"distinct"`
}

// join on 配置
type JoinOn struct {
	LeftField  string `json:"left_field" mapstructure:"left_field"`   //传递 name
	RightField string `json:"right_field" mapstructure:"right_field"` //传递 name
	Operator   string `json:"operator" mapstructure:"operator"`
}

// 节点类型为union的节点配置
type UnionNodeCfg struct {
	UnionType   string         `json:"union_type" mapstructure:"union_type"`
	UnionFields [][]UnionField `json:"union_fields" mapstructure:"union_fields"`
	Filters     *cond.CondCfg  `json:"filters,omitempty" mapstructure:"filters"`
}

type UnionField struct {
	Field     string `json:"field" mapstructure:"field"`
	ValueFrom string `json:"value_from" mapstructure:"value_from"` // "field" 或 "const"
}

type SQLNodeCfg struct {
	SQLExpression string `json:"sql_expression" mapstructure:"sql_expression"`
}

type ExcelConfig struct {
	SheetName        string `json:"sheet_name"`          // sheet页，逗号分隔
	StartCell        string `json:"start_cell"`          // 起始单元格
	EndCell          string `json:"end_cell"`            // 结束单元格
	HasHeaders       bool   `json:"has_headers"`         // 是否首行作为列名
	SheetAsNewColumn bool   `json:"sheet_as_new_column"` // 是否将sheet作为新列
}

//go:generate mockgen -source ../interfaces/data_view_access.go -destination ../interfaces/mock/mock_data_view_access.go
type DataViewAccess interface {
	GetDataViewsByIDs(ctx context.Context, ids string, includeDataScopeView bool) ([]*DataView, error)
	GetDataViewIDByName(ctx context.Context, viewName string) (string, error)
}
