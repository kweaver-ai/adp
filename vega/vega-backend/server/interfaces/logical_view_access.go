// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"fmt"
)

// 导入视图的结构体，condition 为 any 类型，兼容新旧过滤器格式
// builtin 为 any 类型,兼容数字 0, 1 和 bool 类型
// 添加 loggroup_filters 字段，防止由日志分组升级上来的视图导出后，分组条件丢失
// type CreateDataView struct {
// 	ViewID        string           `json:"id"`
// 	ViewName      string           `json:"name"`
// 	TechnicalName string           `json:"technical_name"`
// 	GroupID       string           `json:"group_id"`
// 	GroupName     string           `json:"group_name"`
// 	Type          string           `json:"type"`
// 	QueryType     string           `json:"query_type"`
// 	Tags          []string         `json:"tags"`
// 	Comment       string           `json:"comment"`
// 	Builtin       any              `json:"builtin"`
// 	DataSourceID  string           `json:"data_source_id"`
// 	FileName      string           `json:"file_name"`
// 	ExcelConfig   *ExcelConfig     `json:"excel_config"`
// 	DataScope     []*DataScopeNode `json:"data_scope"`
// 	Fields        []*ViewField     `json:"fields"`
// 	PrimaryKeys   []string         `json:"primary_keys"`
// 	ModuleType    string           `json:"module_type"`
// 	DataSource    map[string]any   `json:"data_source"` // TODO 暂时为索引库创建视图保留，后续删掉
// }

// // 数据视图结构体
// type DataView struct {
// 	SimpleDataView
// 	Fields         []*ViewField          `json:"fields"`
// 	FieldTypeMap   map[string]string     `json:"-"`
// 	FieldsMap      map[string]*ViewField `json:"fields_map"` // todo: 指标模型需要的字段,文月确认结构体,临时加的
// 	ModuleType     string                `json:"module_type"`
// 	Creator        AccountInfo           `json:"creator"`
// 	Updater        AccountInfo           `json:"updater"`
// 	DataScope      []*DataScopeNode      `json:"data_scope,omitempty"`
// 	ExcelConfig    *ExcelConfig          `json:"excel_config,omitempty"`
// 	MetadataFormID string                `json:"metadata_form_id,omitempty"`
// 	PrimaryKeys    []string              `json:"primary_keys"`
// 	SQLStr         string                `json:"sql_str,omitempty"`
// 	MetaTableName  string                `json:"meta_table_name,omitempty"`
// 	VegaDataSource *DataSource           `json:"-"`
// }

// // 简单的视图结构，列表查询接口使用
// type SimpleDataView struct {
// 	ViewID            string         `json:"id"`
// 	ViewName          string         `json:"name"`
// 	TechnicalName     string         `json:"technical_name"`
// 	GroupID           string         `json:"group_id"`
// 	GroupName         string         `json:"group_name"`
// 	Type              string         `json:"type" binding:"required,oneof=atomic custom"`
// 	QueryType         string         `json:"query_type" binding:"required,oneof=SQL DSL IndexBase"`
// 	Tags              []string       `json:"tags"`
// 	Comment           string         `json:"comment"`
// 	Builtin           bool           `json:"builtin"`
// 	DataSource        map[string]any `json:"data_source"` // TODO 暂时为索引库创建视图保留，后续删掉
// 	CreateTime        int64          `json:"create_time"`
// 	UpdateTime        int64          `json:"update_time"`
// 	DeleteTime        int64          `json:"delete_time"`
// 	DataSourceType    string         `json:"data_source_type,omitempty"`
// 	DataSourceID      string         `json:"data_source_id,omitempty"`
// 	DataSourceName    string         `json:"data_source_name,omitempty"`
// 	DataSourceCatalog string         `json:"data_source_catalog,omitempty"`
// 	FileName          string         `json:"file_name,omitempty"`
// 	Status            string         `json:"status,omitempty"`

// 	// 操作权限
// 	Operations []string `json:"operations"`
// }

// DataScopeNode 表示图中的节点
type DataScopeNode struct {
	ID           string         `json:"id"`
	Title        string         `json:"title"`
	Type         string         `json:"type"`
	InputNodes   []string       `json:"input_nodes"`
	Config       map[string]any `json:"config"`
	OutputFields []*ViewField   `json:"output_fields"`
}

// 节点类型为view的节点配置
type ViewNodeCfg struct {
	ViewID   string         `json:"view_id" mapstructure:"view_id"`
	Filters  *FilterCondCfg `json:"filters,omitempty" mapstructure:"filters"`
	Distinct Distinct       `json:"distinct" mapstructure:"distinct"`
	View     *Resource      `json:"view,omitempty" mapstructure:"view"`
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

// 数据视图字段
type ViewField struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	DisplayName  string `json:"display_name"`
	OriginalName string `json:"original_name"`
	Description  string `json:"description"`
	SrcNodeID    string `json:"src_node_id,omitempty"`
	SrcNodeName  string `json:"src_node_name,omitempty"`

	Features []FieldFeature `json:"features,omitempty"`
}

// 字段特征
type FieldFeature struct {
	FeatureName string           `json:"name"`       // 特征名称
	FeatureType FieldFeatureType `json:"type"`       // 特征类型：keyword, fulltext, vector
	Comment     string           `json:"comment"`    // 特征描述
	RefField    string           `json:"ref_field"`  // 核心：引用的字段名
	IsDefault   bool             `json:"is_default"` //  同类型下只能有一个为 true
	IsNative    bool             `json:"is_native"`  // 是否为底层物理同步生成的（true:系统, false:手动）
	Config      map[string]any   `json:"config"`     // 特有配置（如分词器、权重、向量维度）
}

type KeywordConfig struct {
	IgnoreAboveLen int `json:"ignore_above_len" mapstructure:"ignore_above_len"`
}

type FulltextConfig struct {
	Analyzer string `json:"analyzer" mapstructure:"analyzer"`
}

type VectorConfig struct {
	ModelID string `json:"model_id" mapstructure:"model_id"`

	//Model *SmallModel `json:"-"`
}

func (v *ViewField) String() string {
	return fmt.Sprintf("ViewField{name: %s, type: %s, description: %s, display_name: %s, original_name: %s}",
		v.Name, v.Type, v.Description, v.DisplayName, v.OriginalName)
}

type ListViewQueryParams struct {
	Type            string
	QueryType       string
	DataSourceType  string
	DataSourceID    string
	FileName        string
	Keyword         string
	Name            string
	NamePattern     string
	TechnicalName   string
	GroupID         string
	GroupName       string
	Status          []string
	CreateTimeStart int64
	CreateTimeEnd   int64
	UpdateTimeStart int64
	UpdateTimeEnd   int64
	Tag             string
	Builtin         []bool
	Operations      []string
	PaginationQueryParams
}

type DataSource struct {
	ID           string  `json:"id"`                   // 数据源业务id
	Name         string  `json:"name"`                 // 数据源名称
	Type         string  `json:"type"`                 // 数据库类型名称
	BinData      BinData `json:"bin_data"`             // 数据源配置信息
	Comment      string  `json:"comment"`              // 描述
	LastScanTime int64   `json:"last_scan_time"`       // 上一次扫描时间
	Status       string  `json:"status"`               // 数据源状态：扫描中、可用
	CreatorID    string  `json:"created_by_uid"`       // 创建人id
	CreatorType  string  `json:"created_by_user_type"` // 创建人类型
	CreateTime   int64   `json:"created_at"`           // 创建时间
	UpdaterID    string  `json:"updated_by_uid"`       // 更新人id
	UpdaterType  string  `json:"updated_by_user_type"` // 更新人类型
	UpdateTime   int64   `json:"updated_at"`           // 更新时间

}

type BinData struct {
	CatalogName     string `json:"catalog_name"`
	DataBaseName    string `json:"database_name"`
	ConnectProtocol string `json:"connect_protocol"`
	Schema          string `json:"schema"`
	Host            string `json:"host"`
	Port            int    `json:"port"`
	Account         string `json:"account"`
	Password        string `json:"password"`
	Token           string `json:"token"`
	StorageProtocol string `json:"storage_protocol"`
	StorageBase     string `json:"storage_base"`
	ReplicaSet      string `json:"replica_set"`
}

type DataSourceStatus struct {
	Status int32 `gorm:"column:status" json:"status"`
}

type ListDataSourcesResult struct {
	Entries    []*DataSource `json:"entries"`
	TotalCount int           `json:"total_count"`
}

//go:generate mockgen -source ../interfaces/data_source_access.go -destination ../interfaces/mock/mock_data_source_access.go
type DataSourceAccess interface {
	GetDataSourceByID(ctx context.Context, dataSourceID string) (*DataSource, error)
	ListDataSources(ctx context.Context) (*ListDataSourcesResult, error)
}
