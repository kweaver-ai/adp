// Package interfaces defines entities, DTOs, and service interfaces.
package interfaces

const (
	CatalogTypePhysical string = "physical"
	CatalogTypeLogical  string = "logical"
)

const (
	CatalogHealthStatusHealthy   string = "healthy"
	CatalogHealthStatusDegraded  string = "degraded"
	CatalogHealthStatusUnhealthy string = "unhealthy"
	CatalogHealthStatusOffline   string = "offline"
	CatalogHealthStatusDisabled  string = "disabled"
)

type CatalogHealthCheckStatus struct {
	HealthCheckStatus string `json:"health_check_status"`
	LastCheckTime     int64  `json:"last_check_time"`
	HealthCheckResult string `json:"health_check_result"`
}

// Catalog represents a Catalog entity.
type Catalog struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Tags        []string `json:"tags"`
	Description string   `json:"description"`

	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`

	ConnectorType   string         `json:"connector_type"`
	ConnectorConfig map[string]any `json:"connector_config"`
	Metadata        map[string]any `json:"metadata"`

	HealthCheckEnabled bool `json:"health_check_enabled"`
	CatalogHealthCheckStatus

	Creator    AccountInfo `json:"creator"`
	CreateTime int64       `json:"create_time"`
	Updater    AccountInfo `json:"updater"`
	UpdateTime int64       `json:"update_time"`
}

// CatalogsQueryParams holds catalog list query parameters.
type CatalogsQueryParams struct {
	PaginationParams
	Type              string
	HealthCheckStatus string
}

// CatalogCreateRequest represents create catalog request.
type CatalogRequest struct {
	Name            string         `json:"name"`
	Tags            []string       `json:"tags"`
	Description     string         `json:"description"`
	ConnectorType   string         `json:"connector_type"`
	ConnectorConfig map[string]any `json:"connector_config"`

	OriginCatalog *Catalog `json:"-"`
}
