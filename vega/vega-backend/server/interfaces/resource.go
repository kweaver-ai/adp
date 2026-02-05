// Package interfaces defines entities, DTOs, and service interfaces.
package interfaces

const (
	ResourceCategoryTable     string = "table"
	ResourceCategoryFile      string = "file"
	ResourceCategoryFileset   string = "fileset"
	ResourceCategoryAPI       string = "api"
	ResourceCategoryMetric    string = "metric"
	ResourceCategoryTopic     string = "topic"
	ResourceCategoryIndex     string = "index"
	ResourceCategoryLogicView string = "logicview"
	ResourceCategoryDataset   string = "dataset"
)

const (
	ResourceStatusActive     string = "active"
	ResourceStatusDisabled   string = "disabled"
	ResourceStatusDeprecated string = "deprecated"
	ResourceStatusStale      string = "stale"
)

// Resource represents a Data Resource entity.
type Resource struct {
	ID          string   `json:"id"`
	CatalogID   string   `json:"catalog_id"`
	Name        string   `json:"name"`
	Tags        []string `json:"tags"`
	Description string   `json:"description"`

	Category string `json:"category"` // 资源类别：table/file/fileset/...

	Status        string `json:"status"`         // 状态：active/stale/disabled
	StatusMessage string `json:"status_message"` // 状态消息

	// 新增字段：支持自动发现
	Database         string         `json:"database,omitempty"`          // 所属数据库（实例级 Catalog 时填充）
	SourceIdentifier string         `json:"source_identifier"`           // 源端标识（原始表名/路径）
	SourceMetadata   map[string]any `json:"source_metadata,omitempty"`   // 源端配置（JSON）
	SchemaDefinition []Property     `json:"schema_definition,omitempty"` // Schema定义

	Creator    AccountInfo `json:"creator"`
	CreateTime int64       `json:"create_time"`
	Updater    AccountInfo `json:"updater"`
	UpdateTime int64       `json:"update_time"`
}

type Property struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	DisplayName  string `json:"display_name"`
	OriginalName string `json:"original_name"`
	Description  string `json:"description"`
}

// ResourcesQueryParams holds resource list query parameters.
type ResourcesQueryParams struct {
	PaginationParams
	CatalogID string
	Category  string
	Status    string
	Database  string
}

// ResourceCreateRequest represents create resource request.
type ResourceRequest struct {
	CatalogID   string   `json:"catalog_id"`
	Name        string   `json:"name"`
	Tags        []string `json:"tags"`
	Description string   `json:"description"`

	Category string `json:"category"`

	Status string `json:"status"`

	Database         string `json:"database,omitempty"` // 所属数据库（实例级 Catalog 时填充）
	SourceIdentifier string `json:"source_identifier"`  // 源端标识（原始表名/路径）

	OriginResource *Resource `json:"-"`
}
