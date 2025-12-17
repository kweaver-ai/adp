package interfaces

import (
	"context"
	"database/sql"
	"fmt"
)

// 导入视图的结构体，condition 为 any 类型，兼容新旧过滤器格式
// builtin 为 any 类型,兼容数字 0, 1 和 bool 类型
// 添加 loggroup_filters 字段，防止由日志分组升级上来的视图导出后，分组条件丢失
type CreateDataView struct {
	ViewID        string           `json:"id"`
	ViewName      string           `json:"name"`
	TechnicalName string           `json:"technical_name"`
	GroupID       string           `json:"group_id"`
	GroupName     string           `json:"group_name"`
	Type          string           `json:"type"`
	QueryType     string           `json:"query_type"`
	Tags          []string         `json:"tags"`
	Comment       string           `json:"comment"`
	Builtin       any              `json:"builtin"`
	DataSourceID  string           `json:"data_source_id"`
	FileName      string           `json:"file_name"`
	ExcelConfig   *ExcelConfig     `json:"excel_config"`
	DataScope     []*DataScopeNode `json:"data_scope"`
	Fields        []*ViewField     `json:"fields"`
	PrimaryKeys   []string         `json:"primary_keys"`
	ModuleType    string           `json:"module_type"`
	DataSource    map[string]any   `json:"data_source"` // TODO 暂时为索引库创建视图保留，后续删掉
}

// 数据视图结构体
type DataView struct {
	SimpleDataView
	Fields         []*ViewField          `json:"fields"`
	FieldTypeMap   map[string]string     `json:"-"`
	FieldsMap      map[string]*ViewField `json:"fields_map"` // todo: 指标模型需要的字段,文月确认结构体,临时加的
	ModuleType     string                `json:"module_type"`
	Creator        AccountInfo           `json:"creator"`
	Updater        AccountInfo           `json:"updater"`
	DataScope      []*DataScopeNode      `json:"data_scope,omitempty"`
	ExcelConfig    *ExcelConfig          `json:"excel_config,omitempty"`
	MetadataFormID string                `json:"metadata_form_id,omitempty"`
	PrimaryKeys    []string              `json:"primary_keys"`
	SQLStr         string                `json:"sql_str,omitempty"`
	MetaTableName  string                `json:"meta_table_name,omitempty"`
	VegaDataSource *DataSource           `json:"-"`
	// FieldScope       uint8             `json:"field_scope"`
	// Condition        *CondCfg          `json:"filters"`
	// LogGroupFilters  string            `json:"loggroup_filters"`
	// JobID            string            `json:"job_id"`
	// JobStatus        string            `json:"job_status"`
	// JobStatusDetails string            `json:"job_status_details"`
	// StreamingTopic   string            `json:"streaming_topic,omitempty"`
}

// 简单的视图结构，列表查询接口使用
type SimpleDataView struct {
	ViewID            string         `json:"id"`
	ViewName          string         `json:"name"`
	TechnicalName     string         `json:"technical_name"`
	GroupID           string         `json:"group_id"`
	GroupName         string         `json:"group_name"`
	Type              string         `json:"type" binding:"required,oneof=atomic custom"`
	QueryType         string         `json:"query_type" binding:"required,oneof=SQL DSL IndexBase"`
	Tags              []string       `json:"tags"`
	Comment           string         `json:"comment"`
	Builtin           bool           `json:"builtin"`
	DataSource        map[string]any `json:"data_source"` // TODO 暂时为索引库创建视图保留，后续删掉
	CreateTime        int64          `json:"create_time"`
	UpdateTime        int64          `json:"update_time"`
	DeleteTime        int64          `json:"delete_time"`
	DataSourceType    string         `json:"data_source_type,omitempty"`
	DataSourceID      string         `json:"data_source_id,omitempty"`
	DataSourceName    string         `json:"data_source_name,omitempty"`
	DataSourceCatalog string         `json:"data_source_catalog,omitempty"`
	FileName          string         `json:"file_name,omitempty"`
	Status            string         `json:"status,omitempty"`

	// 操作权限
	Operations []string `json:"operations"`
}

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
	ViewID   string    `json:"view_id" mapstructure:"view_id"`
	Filters  *CondCfg  `json:"filters,omitempty" mapstructure:"filters"`
	Distinct Distinct  `json:"distinct" mapstructure:"distinct"`
	View     *DataView `json:"view,omitempty" mapstructure:"view"`
}

type Distinct struct {
	Enable bool     `json:"enable" mapstructure:"enable"`
	Fields []string `json:"fields,omitempty" mapstructure:"fields"`
}

// 节点类型为join的节点配置
type JoinNodeCfg struct {
	JoinType string    `json:"join_type" mapstructure:"join_type"`
	JoinOn   []*JoinOn `json:"join_on" mapstructure:"join_on"`
	Filters  *CondCfg  `json:"filters,omitempty" mapstructure:"filters"`
	Distinct Distinct  `json:"distinct,omitempty" mapstructure:"distinct"`
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
	Filters     *CondCfg       `json:"filters,omitempty" mapstructure:"filters"`
}

type UnionField struct {
	Field     string `json:"field" mapstructure:"field"`
	ValueFrom string `json:"value_from" mapstructure:"value_from"` // "field" 或 "const"
}

type SQLNodeCfg struct {
	SQLExpression string `json:"sql_expression" mapstructure:"sql_expression"`
}

type ExcelConfig struct {
	Sheet            string `json:"sheet"`               // sheet页，逗号分隔
	StartCell        string `json:"start_cell"`          // 起始单元格
	EndCell          string `json:"end_cell"`            // 结束单元格
	HasHeaders       bool   `json:"has_headers"`         // 是否首行作为列名
	SheetAsNewColumn bool   `json:"sheet_as_new_column"` // 是否将sheet作为新列
}

// 数据视图字段
type ViewField struct {
	Name              string       `json:"name"`
	Type              string       `json:"type"`
	Comment           string       `json:"comment"`
	DisplayName       string       `json:"display_name"`
	OriginalName      string       `json:"original_name"`
	DataLength        int32        `json:"data_length"`
	DataAccuracy      int32        `json:"data_accuracy"`
	Status            string       `json:"status"`
	IsNullable        string       `json:"is_nullable"`
	BusinessTimestamp bool         `json:"business_timestamp"`
	SrcNodeID         string       `json:"src_node_id,omitempty"`
	SrcNodeName       string       `json:"src_node_name,omitempty"`
	IsPrimaryKey      sql.NullBool `json:"is_primary_key,omitempty"`

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
	return fmt.Sprintf("ViewField{name: %s, type: %s, comment: %s, display_name: %s, original_name: %s}",
		v.Name, v.Type, v.Comment, v.DisplayName, v.OriginalName)
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
	PaginationQueryParameters
}

// 允许/不允许实时订阅的配置
type RealTimeStreaming struct {
	OpenStreaming bool   `json:"open_streaming"`
	CreateTime    int64  `json:"-"`
	UpdateTime    int64  `json:"-"`
	ViewID        string `json:"-"`
	JobID         string `json:"-"`
}

// 更新视图状态
type UpdateViewStatus struct {
	ViewStatus string      `json:"status"`
	DeleteTime int64       `json:"delete_time"`
	UpdateTime int64       `json:"update_time"`
	Updater    AccountInfo `json:"-"`
}

// 视图分组
type ViewGroupReq struct {
	GroupID   string `json:"group_id"`
	GroupName string `json:"group_name"`
	Builtin   bool   `json:"builtin"`
}

// 原子视图更新的信息
type AtomicViewUpdateReq struct {
	ViewName   string       `json:"name"`
	Fields     []*ViewField `json:"fields"`
	Comment    string       `json:"comment"`
	ViewID     string       `json:"-"`
	Updater    AccountInfo  `json:"-"`
	UpdateTime int64        `json:"-"`
}

type MarkViewDeletedParams struct {
	ViewIDs    []string
	DeleteTime int64
	ViewStatus string
}

//go:generate mockgen -source ../interfaces/data_view_access.go -destination ../interfaces/mock/mock_data_view_access.go
type DataViewAccess interface {
	CreateDataViews(ctx context.Context, tx *sql.Tx, views []*DataView) error
	DeleteDataViews(ctx context.Context, tx *sql.Tx, viewIDs []string) error
	UpdateDataView(ctx context.Context, tx *sql.Tx, view *DataView) error
	GetDataViews(ctx context.Context, viewID []string) ([]*DataView, error)
	ListDataViews(ctx context.Context, viewsQuery *ListViewQueryParams) ([]*SimpleDataView, error)
	GetDataViewsTotal(ctx context.Context, viewsQuery *ListViewQueryParams) (int, error)

	CheckDataViewExistByName(ctx context.Context, tx *sql.Tx, viewName, groupName string) (string, bool, error)
	CheckDataViewExistByTechnicalName(ctx context.Context, tx *sql.Tx, viewTechnicalName, groupName string) (string, bool, error)
	CheckDataViewExistByID(ctx context.Context, tx *sql.Tx, viewID string) (string, bool, error)

	GetDetailedDataViewMapByIDs(ctx context.Context, viewIDs []string) (map[string]*DataView, error)
	GetSimpleDataViewMapByIDs(ctx context.Context, viewIDs []string) (map[string]*DataView, error)

	// 更新视图分组
	UpdateDataViewsGroup(ctx context.Context, tx *sql.Tx, viewIDs []string, groupID string) error
	// 允许、不允许实时订阅
	UpdateDataViewRealTimeStreaming(ctx context.Context, tx *sql.Tx, realTimeStreaming *RealTimeStreaming) error

	UpdateDataViewsAttrs(ctx context.Context, attrs *AtomicViewUpdateReq) error
	UpdateViewStatus(ctx context.Context, tx *sql.Tx, viewIDs []string, param *UpdateViewStatus) error

	// 批量根据视图id获取实时订阅任务id
	GetJobsByDataViewIDs(ctx context.Context, viewID []string) ([]*JobInfo, error)

	// 根据分组id获取视图
	GetSimpleDataViewsByGroupID(ctx context.Context, tx *sql.Tx, groupID string) ([]*SimpleDataView, error)
	GetDataViewsByGroupID(ctx context.Context, groupID string) ([]*DataView, error)

	// 根据数据源id获取视图
	GetDataViewsBySourceID(ctx context.Context, sourceID string) ([]*DataView, error)

	// 批量标记删除视图
	MarkDataViewsDeleted(ctx context.Context, tx *sql.Tx, params *MarkViewDeletedParams) error
}
