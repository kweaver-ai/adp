// Package interfaces defines entities, DTOs, and service interfaces.
package interfaces

import "context"

// DiscoveryResult represents the result of a discovery operation.
type DiscoveryResult struct {
	CatalogID      string `json:"catalog_id"`
	NewCount       int    `json:"new_count"`
	StaleCount     int    `json:"stale_count"`
	UnchangedCount int    `json:"unchanged_count"`
	Message        string `json:"message"`
}

// DiscoveryService interface defines discovery functionality.
type DiscoveryService interface {
	// DiscoverCatalog discovers resources for a specific catalog.
	DiscoverCatalog(ctx context.Context, catalog *Catalog) (*DiscoveryResult, error)
}
