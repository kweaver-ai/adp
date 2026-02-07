// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package worker

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"vega-backend/interfaces"
	"vega-backend/logics/connectors"
)

type indexDiscoveryItem struct {
	resource  *interfaces.Resource
	indexMeta *interfaces.IndexMeta
}

// discoverIndexResources discovers index resources from an index connector.
func (dw *discoveryWorker) discoverIndexResources(ctx context.Context,
	catalog *interfaces.Catalog, connector connectors.Connector) (*interfaces.DiscoveryResult, error) {

	indexConnector, ok := connector.(connectors.IndexConnector)
	if !ok {
		return nil, fmt.Errorf("connector does not support index discovery")
	}

	// Step 1: List Indices
	sourceIndices, err := indexConnector.ListIndexes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list indices: %w", err)
	}
	logger.Infof("Discovered %d indices from source", len(sourceIndices))

	// Step 2: Get Existing Resources
	existingResources, err := dw.rs.GetByCatalogID(ctx, catalog.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing resources: %w", err)
	}

	// Step 3: Reconcile
	result, items, err := dw.reconcileIndexResources(ctx, catalog, sourceIndices, existingResources)
	if err != nil {
		return nil, fmt.Errorf("failed to reconcile resources: %w", err)
	}

	// Step 4: Enrich
	if err := dw.enrichIndexMetadata(ctx, indexConnector, items); err != nil {
		return nil, fmt.Errorf("failed to enrich index metadata: %w", err)
	}

	logger.Infof("Discovery completed for catalog %s: new=%d, stale=%d, unchanged=%d",
		catalog.ID, result.NewCount, result.StaleCount, result.UnchangedCount)

	return result, nil
}

// reconcileIndexResources reconciles source indices with existing resources.
func (dw *discoveryWorker) reconcileIndexResources(ctx context.Context,
	catalog *interfaces.Catalog, sourceIndices []*interfaces.IndexMeta,
	existingResources []*interfaces.Resource) (*interfaces.DiscoveryResult, []indexDiscoveryItem, error) {

	result := &interfaces.DiscoveryResult{
		CatalogID: catalog.ID,
	}

	var items []indexDiscoveryItem

	existingMap := make(map[string]*interfaces.Resource)
	for _, r := range existingResources {
		existingMap[r.SourceIdentifier] = r
	}

	sourceMap := make(map[string]*interfaces.IndexMeta)
	for _, idx := range sourceIndices {
		sourceMap[idx.Name] = idx
	}

	// Handle new and existing
	for _, idx := range sourceIndices {
		sourceIdentifier := idx.Name

		if resource, ok := existingMap[sourceIdentifier]; ok {
			if resource.Status == interfaces.ResourceStatusStale {
				if err := dw.rs.UpdateStatus(ctx, resource.ID, interfaces.ResourceStatusActive, ""); err != nil {
					logger.Errorf("Failed to reactivate resource %s: %v", resource.ID, err)
				}
			}
			result.UnchangedCount++
			items = append(items, indexDiscoveryItem{
				resource:  resource,
				indexMeta: idx,
			})
		} else {
			resource, err := dw.createIndexResource(ctx, catalog, idx)
			if err != nil {
				logger.Errorf("Failed to create resource %s: %v", sourceIdentifier, err)
			} else {
				result.NewCount++
				items = append(items, indexDiscoveryItem{
					resource:  resource,
					indexMeta: idx,
				})
			}
		}
	}

	// Handle stale
	for sourceIdentifier, existing := range existingMap {
		if _, ok := sourceMap[sourceIdentifier]; !ok {
			if existing.Status != interfaces.ResourceStatusStale {
				if err := dw.rs.UpdateStatus(ctx, existing.ID, interfaces.ResourceStatusStale, ""); err != nil {
					logger.Errorf("Failed to mark resource %s as stale: %v", existing.ID, err)
				} else {
					result.StaleCount++
				}
			}
		}
	}

	result.Message = fmt.Sprintf("Discovery completed: %d new, %d stale, %d unchanged",
		result.NewCount, result.StaleCount, result.UnchangedCount)

	return result, items, nil
}

// createIndexResource creates a new resource for an index.
func (dw *discoveryWorker) createIndexResource(ctx context.Context, catalog *interfaces.Catalog,
	index *interfaces.IndexMeta) (*interfaces.Resource, error) {

	req := &interfaces.ResourceRequest{
		CatalogID: catalog.ID,
		Name:      index.Name,
		Category:  interfaces.ResourceCategoryIndex,
		Status:    interfaces.ResourceStatusActive,
	}
	id, err := dw.rs.Create(ctx, req)
	if err != nil {
		return nil, err
	}

	return dw.rs.GetByID(ctx, id)
}

// enrichIndexMetadata enriches index resources with detailed metadata.
func (dw *discoveryWorker) enrichIndexMetadata(ctx context.Context,
	indexConnector connectors.IndexConnector, items []indexDiscoveryItem) error {

	for _, item := range items {
		idx := item.indexMeta
		resource := item.resource

		// Get detailed metadata (mappings)
		if err := indexConnector.GetIndexMeta(ctx, idx); err != nil {
			logger.Warnf("Failed to get metadata for index %s: %v", idx.Name, err)
			return err
		}

		// Map fields to SchemaDefinition
		var columns []interfaces.Property
		for _, field := range idx.Mapping {
			columns = append(columns, interfaces.Property{
				Name:         field.Name,
				Type:         field.Type,
				DisplayName:  field.Name,
				OriginalName: field.Name,
				Description:  "",
			})
		}
		resource.SchemaDefinition = columns

		// Populate SourceMetadata
		sourceMetadata := make(map[string]any)
		if resource.SourceMetadata != nil {
			sourceMetadata = resource.SourceMetadata
		}

		sourceMetadata["properties"] = idx.Properties
		sourceMetadata["mapping"] = idx.Mapping
		resource.SourceMetadata = sourceMetadata

		// Update Resource
		if err := dw.rs.UpdateResource(ctx, resource); err != nil {
			logger.Errorf("Failed to update metadata for index %s: %v", idx.Name, err)
			return err
		}

		// Wait a bit to avoid overwhelming the server? No, it's fine for now.
		// Just logging
		logger.Infof("Enriched index %s: fields=%d", idx.Name, len(columns))
	}
	return nil
}
