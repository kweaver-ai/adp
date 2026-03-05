// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

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

const (
	// 模块名称
	MODULE_TYPE_DATA_VIEW                 = "data_view"
	MODULE_TYPE_DATA_VIEW_ROW_COLUMN_RULE = "data_view_row_column_rule"
	INDEX_BASE                            = "index_base"

	ALLOW     = "ID_AUDIT_ACTION_ALLOW"
	NOT_ALLOW = "ID_AUDIT_ACTION_NOT_ALLOW"

	OBJECTTYPE_DATA_VIEW           = "ID_AUDIT_DATA_VIEW"
	OBJECTTYPE_REAL_TIME_STREAMING = "ID_AUDIT_REAL_TIME_STREAMING"

	// 视图名称、视图备注、视图分组名最大长度
	MaxLength_ViewName      = 255
	MaxLength_ViewComment   = 255
	MaxLength_ViewGroupName = 40
	// 视图字段名称、字段显示名、字段备注、字段特征备注的最大长度
	MaxLength_ViewFieldName           = 255
	MaxLength_ViewFieldDisplayName    = 255
	MaxLength_ViewFieldFeatureName    = 255
	MaxLength_ViewFieldComment        = 1000
	MaxLength_ViewFieldFeatureComment = 1000

	Non_Builtin             = 0
	Builtin                 = 1
	Default_Include_Builtin = "false"

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
		// AttrFields_OpenStreaming: {},
		AttrFields_GroupName: {},
		AttrFields_Fields:    {},
		AttrFields_Name:      {},
		AttrFields_Comment:   {},
	}

	DataScopeNodeTypeMap = map[string]struct{}{
		DataScopeNodeType_View:   {},
		DataScopeNodeType_Join:   {},
		DataScopeNodeType_Union:  {},
		DataScopeNodeType_Sql:    {},
		DataScopeNodeType_Output: {},
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
