// Package interfaces defines entities, DTOs, and service interfaces.
package interfaces

import "context"

// CatalogService defines catalog business logic interface.
type CatalogService interface {
	// Create creates a new Catalog.
	Create(ctx context.Context, req *CatalogRequest) (string, error)
	// Get retrieves a Catalog by ID.
	GetByID(ctx context.Context, id string, withSensitiveFields bool) (*Catalog, error)
	// Get retrieves a Catalog by IDs.
	GetByIDs(ctx context.Context, ids []string) ([]*Catalog, error)
	// List lists Catalogs with filters.
	List(ctx context.Context, params CatalogsQueryParams) ([]*Catalog, int64, error)
	// Update updates a Catalog.
	Update(ctx context.Context, id string, req *CatalogRequest) error
	// DeleteByIDs deletes Catalogs by IDs.
	DeleteByIDs(ctx context.Context, ids []string) error
	// CheckExistByID checks if a Catalog exists by ID.
	CheckExistByID(ctx context.Context, id string) (bool, error)
	// CheckExistByName checks if a Catalog exists by name.
	CheckExistByName(ctx context.Context, name string) (bool, error)
	// TestConnection tests catalog connection.
	TestConnection(ctx context.Context, catalog *Catalog) (*CatalogHealthCheckStatus, error)

	// UpdateMetadata updates a Catalog metadata.
	UpdateMetadata(ctx context.Context, id string, metadata map[string]any) error
}
