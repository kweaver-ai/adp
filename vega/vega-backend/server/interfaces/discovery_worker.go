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

// DiscoveryWorker interface defines discovery execution functionality.
// This worker is called by the task management service to execute the actual discovery.
type DiscoveryWorker interface {
	Start()
	// ExecuteDiscovery executes discovery for a specific catalog.
	ExecuteDiscovery(ctx context.Context, taskID string) error
}
