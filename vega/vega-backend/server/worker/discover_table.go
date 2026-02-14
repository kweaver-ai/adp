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

type tableDiscoverItem struct {
	resource  *interfaces.Resource
	tableMeta *interfaces.TableMeta
}

// discoverTableResources discovers table resources from a table connector.
// 分步执行：1. 获取表名列表 2. 创建/更新 Resource 3. 逐个补齐详细元数据
func (dw *discoverWorker) discoverTableResources(ctx context.Context,
	catalog *interfaces.Catalog, connector connectors.Connector) (*interfaces.DiscoverResult, error) {

	tableConnector, ok := connector.(connectors.TableConnector)
	if !ok {
		return nil, fmt.Errorf("connector does not support table discover")
	}

	// Step 1: 获取表名列表
	sourceTables, err := tableConnector.ListTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}
	logger.Infof("Discovered %d tables from source", len(sourceTables))

	// Step 2: 获取现有 Resources
	existingResources, err := dw.rs.GetByCatalogID(ctx, catalog.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing resources: %w", err)
	}

	// Step 3: 对比并创建/更新 Resource（基础信息）
	result, items, err := dw.reconcileTableResources(ctx, catalog, sourceTables, existingResources)
	if err != nil {
		return nil, fmt.Errorf("failed to reconcile resources: %w", err)
	}

	// Step 4: 逐个补齐详细元数据
	if err := dw.enrichTableMetadata(ctx, tableConnector, items); err != nil {
		return nil, fmt.Errorf("failed to enrich table metadata: %w", err)
	}

	logger.Infof("Discover completed for catalog %s: new=%d, stale=%d, unchanged=%d",
		catalog.ID, result.NewCount, result.StaleCount, result.UnchangedCount)

	return result, nil
}

// enrichTableMetadata enriches table resources with detailed metadata.
func (dw *discoverWorker) enrichTableMetadata(ctx context.Context,
	tableConnector connectors.TableConnector, items []tableDiscoverItem) error {

	for _, item := range items {
		table := item.tableMeta
		resource := item.resource

		// 获取详细元数据
		err := tableConnector.GetTableMeta(ctx, table)
		if err != nil {
			logger.Warnf("Failed to get metadata for table %s: %v", table.Name, err)
			return err
		}

		// 填充 Resource 元数据
		resource.Database = table.Database
		resource.SchemaDefinition = []*interfaces.Property{}
		for _, column := range table.Columns {
			resource.SchemaDefinition = append(resource.SchemaDefinition, &interfaces.Property{
				Name:         column.Name,
				Type:         column.Type,
				DisplayName:  column.Name,
				OriginalName: column.Name,
				Description:  column.Description,
			})
		}

		sourceMetadata := make(map[string]any)
		if resource.SourceMetadata != nil {
			sourceMetadata = resource.SourceMetadata
		}
		sourceMetadata["columns"] = table.Columns
		if table.TableType != "" {
			sourceMetadata["table_type"] = table.TableType
		}
		if len(table.Properties) > 0 {
			sourceMetadata["properties"] = table.Properties
		}
		if len(table.PKs) > 0 {
			sourceMetadata["primary_keys"] = table.PKs
		}
		if len(table.Indices) > 0 {
			sourceMetadata["indices"] = table.Indices
		}
		if len(table.ForeignKeys) > 0 {
			sourceMetadata["foreign_keys"] = table.ForeignKeys
		}
		resource.SourceMetadata = sourceMetadata

		// 更新 Resource
		if err := dw.rs.UpdateResource(ctx, resource); err != nil {
			logger.Errorf("Failed to update metadata for table %s: %v", table.Name, err)
			return err
		}

		logger.Infof("Enriched table %s: properties=%v, columns=%d, indices=%d, foreign_keys=%d",
			table.Name, table.Properties, len(table.Columns), len(table.Indices), len(table.ForeignKeys))
	}
	return nil
}

// reconcileTableResources reconciles source tables with existing resources.
func (dw *discoverWorker) reconcileTableResources(ctx context.Context,
	catalog *interfaces.Catalog, sourceTables []*interfaces.TableMeta,
	existingResources []*interfaces.Resource) (*interfaces.DiscoverResult, []tableDiscoverItem, error) {

	result := &interfaces.DiscoverResult{
		CatalogID: catalog.ID,
	}

	// 用于返回的 Discover Items
	var items []tableDiscoverItem

	// 构建现有资源的 map（按 SourceIdentifier 索引）
	existingMap := make(map[string]*interfaces.Resource)
	for _, r := range existingResources {
		existingMap[r.SourceIdentifier] = r
	}

	// 构建源端表的 map
	sourceMap := make(map[string]*interfaces.TableMeta)
	for _, t := range sourceTables {
		sourceIdentifier := dw.buildSourceIdentifier(t)
		sourceMap[sourceIdentifier] = t
	}

	// 处理新增和保持的资源
	for _, table := range sourceTables {
		sourceIdentifier := dw.buildSourceIdentifier(table)

		if resource, ok := existingMap[sourceIdentifier]; ok {
			// 已存在，检查状态
			if resource.Status == interfaces.ResourceStatusStale {
				// 之前标记为 stale，现在重新激活
				if err := dw.rs.UpdateStatus(ctx, resource.ID, interfaces.ResourceStatusActive, ""); err != nil {
					logger.Errorf("Failed to reactivate resource %s: %v", resource.ID, err)
				}
			}
			result.UnchangedCount++
			items = append(items, tableDiscoverItem{
				resource:  resource,
				tableMeta: table,
			})
		} else {
			// 新增资源
			resource, err := dw.createResource(ctx, catalog, table, sourceIdentifier)
			if err != nil {
				logger.Errorf("Failed to create resource %s: %v", sourceIdentifier, err)
			} else {
				result.NewCount++
				items = append(items, tableDiscoverItem{
					resource:  resource,
					tableMeta: table,
				})
			}
		}
	}

	// 处理已删除的资源（标记为 stale）
	for sourceIdentifier, existing := range existingMap {
		if _, ok := sourceMap[sourceIdentifier]; !ok {
			// 源端不存在，标记为 stale
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

// buildSourceIdentifier builds the source identifier for a table.
func (dw *discoverWorker) buildSourceIdentifier(table *interfaces.TableMeta) string {
	if table.Database != "" {
		return fmt.Sprintf("%s.%s", table.Database, table.Name)
	}
	return table.Name
}

// createResource creates a new resource.
func (dw *discoverWorker) createResource(ctx context.Context, catalog *interfaces.Catalog,
	table *interfaces.TableMeta, sourceIdentifier string) (*interfaces.Resource, error) {

	req := &interfaces.ResourceRequest{
		CatalogID:        catalog.ID,
		Name:             sourceIdentifier,
		Description:      table.Description,
		Category:         interfaces.ResourceCategoryTable,
		Status:           interfaces.ResourceStatusActive,
		Database:         table.Database,
		SourceIdentifier: sourceIdentifier,
	}
	id, err := dw.rs.Create(ctx, req)
	if err != nil {
		return nil, err
	}

	// 获取刚创建的resource
	resource, err := dw.rs.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return resource, nil
}
