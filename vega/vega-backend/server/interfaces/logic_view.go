// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "fmt"

// 特征类型
type FieldFeatureType string

const (
	FieldFeatureType_Keyword  FieldFeatureType = "keyword"
	FieldFeatureType_Fulltext FieldFeatureType = "fulltext"
	FieldFeatureType_Vector   FieldFeatureType = "vector"

	FieldProperty_Type        = "type"
	FieldProperty_IgnoreAbove = "ignore_above"
	FieldProperty_Analyzer    = "analyzer"
	FieldProperty_Fields      = "fields"
	FieldProperty_Dimension   = "dimension"

	DataScopeNodeType_Resource = "resource"
	DataScopeNodeType_Join     = "join"
	DataScopeNodeType_Union    = "union"
	DataScopeNodeType_Sql      = "sql"
	DataScopeNodeType_Output   = "output"

	// join的类型
	JoinType_Inner     = "inner"
	JoinType_Left      = "left"
	JoinType_Right     = "right"
	JoinType_FullOuter = "full outer"

	// union的类型
	UnionType_All      = "all"
	UnionType_Distinct = "distinct"
)

const (
	// 模块名称
	MODULE_TYPE_DATA_VIEW                 = "data_view"
	MODULE_TYPE_DATA_VIEW_ROW_COLUMN_RULE = "data_view_row_column_rule"
	INDEX_BASE                            = "index_base"

	OBJECTTYPE_DATA_VIEW = "ID_AUDIT_DATA_VIEW"

	// 视图字段名称、字段显示名、字段备注、字段特征备注的最大长度
	MaxLength_ViewFieldName           = 255
	MaxLength_ViewFieldDisplayName    = 255
	MaxLength_ViewFieldFeatureName    = 255
	MaxLength_ViewFieldComment        = 1000
	MaxLength_ViewFieldFeatureComment = 1000

	QueryParam_ImportMode = "import_mode"

	AttrFields_GroupName     = "group_name"
	AttrFields_OpenStreaming = "open_streaming"
	AttrFields_Fields        = "fields"
	AttrFields_Name          = "name"
	AttrFields_Comment       = "comment"

	RegexPattern_Builtin_ViewID    = "^[a-z0-9_][a-z0-9_-]{0,39}$"
	RegexPattern_NonBuiltin_ViewID = "^[a-z0-9][a-z0-9_-]{0,39}$"

	RegexPattern_TechnicalName = "^[a-z_][a-z0-9_]{0,39}$"
)

var (
	AttrFieldsMap = map[string]struct{}{
		AttrFields_GroupName: {},
		AttrFields_Fields:    {},
		AttrFields_Name:      {},
		AttrFields_Comment:   {},
	}

	DataScopeNodeTypeMap = map[string]struct{}{
		DataScopeNodeType_Resource: {},
		DataScopeNodeType_Join:     {},
		DataScopeNodeType_Union:    {},
		DataScopeNodeType_Sql:      {},
		DataScopeNodeType_Output:   {},
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

	FieldFeatureTypeMap = map[FieldFeatureType]struct{}{
		FieldFeatureType_Keyword:  {},
		FieldFeatureType_Fulltext: {},
		FieldFeatureType_Vector:   {},
	}

	DATA_VIEW_SORT = map[string]string{
		"update_time":    "f_update_time",
		"name":           "f_view_name",
		"technical_name": "f_technical_name",
		"group_name":     "f_group_name",
	}

	META_FIELDS = map[string]string{
		"@timestamp":    DataType_Datetime,
		"__write_time":  DataType_Datetime,
		"__data_type":   DataType_String,
		"__index_base":  DataType_String,
		"__category":    DataType_String,
		"__id":          DataType_String,
		"__routing":     DataType_String,
		"__tsid":        DataType_String,
		"__pipeline_id": DataType_String,
		"tags":          DataType_String,
	}
)

type LogicView struct {
	Resource
	FieldsMap map[string]*ViewProperty `json:"fields_map,omitempty" mapstructure:"-"`
}

// DataScopeNode 表示图中的节点
type DataScopeNode struct {
	ID           string          `json:"id"`
	Title        string          `json:"title"`
	Type         string          `json:"type"`
	InputNodes   []string        `json:"input_nodes"`
	Config       map[string]any  `json:"config"`
	OutputFields []*ViewProperty `json:"output_fields"`
}

// 节点类型为view的节点配置
type ResourceNodeCfg struct {
	ResourceID string         `json:"resource_id" mapstructure:"resource_id"`
	Filters    *FilterCondCfg `json:"filters,omitempty" mapstructure:"filters"`
	Distinct   Distinct       `json:"distinct" mapstructure:"distinct"`
	Resource   *Resource      `json:"resource,omitempty" mapstructure:"resource"`
}

type Distinct struct {
	Enable bool     `json:"enable" mapstructure:"enable"`
	Fields []string `json:"fields,omitempty" mapstructure:"fields"`
}

// 节点类型为join的节点配置
type JoinNodeCfg struct {
	JoinType string         `json:"join_type" mapstructure:"join_type"`
	JoinOn   []*JoinOn      `json:"join_on" mapstructure:"join_on"`
	Filters  *FilterCondCfg `json:"filters,omitempty" mapstructure:"filters"`
	Distinct Distinct       `json:"distinct,omitempty" mapstructure:"distinct"`
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
	Filters     *FilterCondCfg `json:"filters,omitempty" mapstructure:"filters"`
}

type UnionField struct {
	Field     string `json:"field" mapstructure:"field"`
	ValueFrom string `json:"value_from" mapstructure:"value_from"` // "field" 或 "const"
}

type SQLNodeCfg struct {
	SQLExpression string `json:"sql_expression" mapstructure:"sql_expression"`
}

// 逻辑视图字段
type ViewProperty struct {
	Property
	SrcNodeID   string `json:"src_node_id,omitempty"`
	SrcNodeName string `json:"src_node_name,omitempty"`
}

func (v *ViewProperty) String() string {
	return fmt.Sprintf("ViewProperty{name: %s, type: %s, description: %s, display_name: %s, original_name: %s}",
		v.Name, v.Type, v.Description, v.DisplayName, v.OriginalName)
}
