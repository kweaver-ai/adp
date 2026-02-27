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

type indexDiscoverItem struct {
	resource  *interfaces.Resource
	indexMeta *interfaces.IndexMeta
}

// discoverIndexResources discovers index resources from an index connector.
// discoverIndexResources 是一个发现索引资源的方法，它会从连接器中获取索引列表，
// 然后与现有资源进行协调，最后丰富索引的元数据信息
// 参数:
//   - ctx: 上下文信息，用于控制请求的超时和取消
//   - catalog: 目录接口，包含目录的相关信息
//   - connector: 连接器接口，用于与数据源进行交互
//
// 返回值:
//   - *interfaces.DiscoverResult: 发现结果，包含新资源、过期资源和未变化资源的统计信息
//   - error: 错误信息，如果在发现过程中出现错误则返回
func (dw *discoverWorker) discoverIndexResources(ctx context.Context,
	catalog *interfaces.Catalog, connector connectors.Connector) (*interfaces.DiscoverResult, error) {

	// 检查连接器是否实现了IndexConnector接口
	indexConnector, ok := connector.(connectors.IndexConnector)
	if !ok {
		return nil, fmt.Errorf("connector does not support index discover")
	}

	// Step 1: List Indices：获取所有的index
	sourceIndices, err := indexConnector.ListIndexes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list indices: %w", err)
	}
	logger.Infof("Discovered %d indices from source", len(sourceIndices))

	// Step 2: Get Existing Resources：查出db是否已存在，然后做比对
	existingResources, err := dw.rs.GetByCatalogID(ctx, catalog.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing resources: %w", err)
	}

	// Step 3: Reconcile:将index数据获取并插入：
	result, items, err := dw.reconcileIndexResources(ctx, catalog, sourceIndices, existingResources)
	if err != nil {
		return nil, fmt.Errorf("failed to reconcile resources: %w", err)
	}

	// Step 4: Enrich ： 为索引项丰富元数据信息
	if err := dw.enrichIndexMetadata(ctx, indexConnector, items); err != nil {
		return nil, fmt.Errorf("failed to enrich index metadata: %w", err)
	}

	logger.Infof("Discover completed for catalog %s: new=%d, stale=%d, unchanged=%d",
		catalog.ID, result.NewCount, result.StaleCount, result.UnchangedCount)

	return result, nil
}

// reconcileIndexResources reconciles source indices with existing resources.
// reconcileIndexResources 协调索引资源，处理新资源、现有资源和过期资源
// 参数:
//   - ctx: 上下文信息，用于控制请求的超时和取消
//   - catalog: 目录信息，包含ID等元数据
//   - sourceIndices: 源索引元数据列表
//   - existingResources: 已存在的资源列表
//
// 返回值:
//   - *interfaces.DiscoverResult: 发现结果，包含目录ID和各类资源的统计信息
//   - []indexDiscoverItem: 索引发现项目列表，包含资源和索引元数据
//   - error: 错误信息，如果处理过程中出现错误则返回
func (dw *discoverWorker) reconcileIndexResources(ctx context.Context,
	catalog *interfaces.Catalog, sourceIndices []*interfaces.IndexMeta,
	existingResources []*interfaces.Resource) (*interfaces.DiscoverResult, []indexDiscoverItem, error) {

	// 初始化发现结果，设置目录ID
	result := &interfaces.DiscoverResult{
		CatalogID: catalog.ID,
	}

	var items []indexDiscoverItem // 索引发现项目列表

	// 创建已存在资源的映射，以源标识符为键
	existingMap := make(map[string]*interfaces.Resource)
	for _, r := range existingResources {
		existingMap[r.SourceIdentifier] = r
	}

	// 创建源索引映射，以索引名为键
	sourceMap := make(map[string]*interfaces.IndexMeta)
	for _, idx := range sourceIndices {
		sourceMap[idx.Name] = idx
	}

	// Handle new and existing
	for _, idx := range sourceIndices {
		sourceIdentifier := idx.Name //test-index

		if resource, ok := existingMap[sourceIdentifier]; ok {
			if resource.Status == interfaces.ResourceStatusStale {
				if err := dw.rs.UpdateStatus(ctx, resource.ID, interfaces.ResourceStatusActive, ""); err != nil {
					logger.Errorf("Failed to reactivate resource %s: %v", resource.ID, err)
				}
			}
			result.UnchangedCount++
			items = append(items, indexDiscoverItem{
				resource:  resource,
				indexMeta: idx,
			})
		} else {
			resource, err := dw.createIndexResource(ctx, catalog, idx)
			if err != nil {
				logger.Errorf("Failed to create resource %s: %v", sourceIdentifier, err)
			} else {
				result.NewCount++
				items = append(items, indexDiscoverItem{
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

	result.Message = fmt.Sprintf("Discover completed: %d new, %d stale, %d unchanged",
		result.NewCount, result.StaleCount, result.UnchangedCount)

	return result, items, nil
}

// createIndexResource creates a new resource for an index.
func (dw *discoverWorker) createIndexResource(ctx context.Context, catalog *interfaces.Catalog,
	index *interfaces.IndexMeta) (*interfaces.Resource, error) {

	req := &interfaces.ResourceRequest{
		CatalogID:        catalog.ID,
		Name:             index.Name,
		Category:         interfaces.ResourceCategoryIndex,
		Status:           interfaces.ResourceStatusActive,
		SourceIdentifier: index.Name,
	}
	id, err := dw.rs.Create(ctx, req)
	if err != nil {
		return nil, err
	}

	return dw.rs.GetByID(ctx, id)
}

// enrichIndexMetadata enriches index resources with detailed metadata.
// enrichIndexMetadata 为索引项丰富元数据信息
// 参数:
//   - ctx: 上下文信息，用于控制请求的超时和取消
//   - indexConnector: 索引连接器，用于获取索引的元数据
//   - items: 需要丰富元数据的索引项列表
//
// 返回值:
//   - error: 如果在处理过程中发生错误，则返回错误信息
func (dw *discoverWorker) enrichIndexMetadata(ctx context.Context,
	indexConnector connectors.IndexConnector, items []indexDiscoverItem) error {

	// 遍历所有需要处理的索引项
	for _, item := range items {
		idx := item.indexMeta
		resource := item.resource

		// Get detailed metadata (mappings) : 获取index详细的信息
		if err := indexConnector.GetIndexMeta(ctx, idx); err != nil {
			logger.Warnf("Failed to get metadata for index %s: %v", idx.Name, err)
			return err
		}

		// Map fields to SchemaDefinition
		var columns []*interfaces.Property
		for _, field := range idx.Mapping {
			columns = append(columns, &interfaces.Property{
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
