// Package types defines data structures used across the application.
package interfaces

// ConnectorConfig holds data source connection configuration.
type ConnectorConfig struct {
	Type     string         `json:"type" mapstructure:"type"`
	Host     string         `json:"host" mapstructure:"host"`
	Port     int            `json:"port" mapstructure:"port"`
	Username string         `json:"username" mapstructure:"username"`
	Password string         `json:"password" mapstructure:"password"`
	Database string         `json:"database" mapstructure:"database"`
	Options  map[string]any `json:"options" mapstructure:"options"`
}

// TableMeta represents table/asset metadata.
type TableMeta struct {
	Name        string           `json:"name"`
	Comment     string           `json:"comment"`
	Database    string           `json:"database"`   // 所属数据库名称（实例级连接时使用）
	SubType     string           `json:"sub_type"`   // table | view | materialized_view
	Properties  map[string]any   `json:"properties"` // 扩展属性：charset, collation, engine, row_count 等
	Columns     []ColumnMeta     `json:"columns"`
	PKs         []string         `json:"primary_keys"`
	Indexes     []IndexInfo      `json:"indexes"`      // 索引列表
	ForeignKeys []ForeignKeyInfo `json:"foreign_keys"` // 外键列表

}

// ForeignKeyInfo represents foreign key metadata.
type ForeignKeyInfo struct {
	Name       string   `json:"name"`
	Columns    []string `json:"columns"`
	RefTable   string   `json:"ref_table"`
	RefColumns []string `json:"ref_columns"`
	OnDelete   string   `json:"on_delete,omitempty"`
	OnUpdate   string   `json:"on_update,omitempty"`
}

// IndexInfo represents index metadata.
type IndexInfo struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Unique  bool     `json:"unique"`
	Primary bool     `json:"primary"`
}

// ColumnMeta represents column metadata.
type ColumnMeta struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	NativeType   string `json:"native_type"`
	Nullable     bool   `json:"nullable"`
	Comment      string `json:"comment"`
	DefaultValue string `json:"default_value,omitempty"` // 默认值
	CharMaxLen   int    `json:"char_max_len,omitempty"`  // 字符最大长度
	NumPrecision int    `json:"num_precision,omitempty"` // 数值精度
	NumScale     int    `json:"num_scale,omitempty"`     // 数值小数位
	Position     int    `json:"position"`                // 列位置（从1开始）
}

// QueryResult represents query execution result.
type QueryResult struct {
	Columns []string         `json:"columns"`
	Rows    []map[string]any `json:"rows"`
	Total   int64            `json:"total"`
}

// FileMeta represents file metadata.
type FileMeta struct {
	Path         string `json:"path"`
	Name         string `json:"name"`
	Size         int64  `json:"size"`
	LastModified int64  `json:"last_modified"`
	ContentType  string `json:"content_type"`
}

// TopicMeta represents message topic metadata.
type TopicMeta struct {
	Name       string `json:"name"`
	Partitions int    `json:"partitions"`
	Replicas   int    `json:"replicas"`
}

// MetricResult represents time-series query result.
type MetricResult struct {
	Metric string            `json:"metric"`
	Values []MetricValue     `json:"values"`
	Labels map[string]string `json:"labels"`
}

// MetricValue represents a single metric data point.
type MetricValue struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

// IndexMeta represents search index metadata.
type IndexMeta struct {
	Name       string               `json:"name"`
	Properties map[string]any       `json:"properties"`
	Mapping    map[string]FieldMeta `json:"mapping"`
}

// FieldMeta represents index field metadata.
type FieldMeta struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Analyzer   string `json:"analyzer,omitempty"`
	Searchable bool   `json:"searchable"`
}

// HealthStatus represents connection health status.
type HealthStatus struct {
	Status    string `json:"status"` // green, yellow, red, unknown
	Message   string `json:"message"`
	LatencyMs int64  `json:"latency_ms"`
}
