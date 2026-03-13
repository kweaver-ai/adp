// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package opensearch provides OpenSearch/ElasticSearch connector implementation.
package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"vega-backend/logics/filter_condition"

	"github.com/mitchellh/mapstructure"
	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"

	"vega-backend/interfaces"
	"vega-backend/logics/connectors"
)

type opensearchConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	IndexPattern string `mapstructure:"index_pattern"`
}

// OpenSearchConnector implements IndexConnector for OpenSearch/ElasticSearch.
type OpenSearchConnector struct {
	enabled bool
	Config  *opensearchConfig
	client  *opensearch.Client
}

// NewOpenSearchConnector 创建 OpenSearch connector 构建器
func NewOpenSearchConnector() connectors.IndexConnector {
	return &OpenSearchConnector{}
}

// GetType returns the data source type.
func (c *OpenSearchConnector) GetType() string {
	return "opensearch"
}

// GetName returns the data source name.
func (c *OpenSearchConnector) GetName() string {
	return "opensearch"
}

// GetMode returns the connector mode.
func (c *OpenSearchConnector) GetMode() string {
	return interfaces.ConnectorModeLocal
}

// GetCategory returns the connector category.
func (c *OpenSearchConnector) GetCategory() string {
	return interfaces.ConnectorCategoryIndex
}

// GetEnabled returns the enabled status.
func (c *OpenSearchConnector) GetEnabled() bool {
	return c.enabled
}

// SetEnabled sets the enabled status.
func (c *OpenSearchConnector) SetEnabled(enabled bool) {
	c.enabled = enabled
}

// GetSensitiveFields returns the sensitive fields for OpenSearch connector.
func (c *OpenSearchConnector) GetSensitiveFields() []string {
	return []string{"password"}
}

// GetFieldConfig returns the field configuration for OpenSearch connector.
func (c *OpenSearchConnector) GetFieldConfig() map[string]interfaces.ConnectorFieldConfig {
	return map[string]interfaces.ConnectorFieldConfig{
		"host":          {Name: "主机地址", Type: "string", Description: "OpenSearch 服务器主机地址", Required: true, Encrypted: false},
		"port":          {Name: "端口号", Type: "integer", Description: "OpenSearch 服务器端口", Required: true, Encrypted: false},
		"username":      {Name: "用户名", Type: "string", Description: "认证用户名", Required: false, Encrypted: false},
		"password":      {Name: "密码", Type: "string", Description: "认证密码", Required: false, Encrypted: true},
		"index_pattern": {Name: "索引模式", Type: "string", Description: "索引匹配模式（可选，如 log-*）", Required: false, Encrypted: false},
	}
}

// New creates a new OpenSearch connector.
func (c *OpenSearchConnector) New(cfg interfaces.ConnectorConfig) (connectors.Connector, error) {
	var osCfg opensearchConfig
	if err := mapstructure.Decode(cfg, &osCfg); err != nil {
		return nil, fmt.Errorf("failed to decode opensearch config: %w", err)
	}

	return &OpenSearchConnector{
		Config: &osCfg,
	}, nil
}

// Connect establishes connection to OpenSearch.
func (c *OpenSearchConnector) Connect(ctx context.Context) error {
	if c.client != nil {
		return nil
	}

	cfg := opensearch.Config{
		Addresses: []string{fmt.Sprintf("http://%s:%d", c.Config.Host, c.Config.Port)},
		Username:  c.Config.Username,
		Password:  c.Config.Password,
	}
	// TODO: Handle SSL/TLS options if needed

	client, err := opensearch.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create opensearch client: %w", err)
	}

	c.client = client
	return nil
}

// Close closes the connection.
func (c *OpenSearchConnector) Close(ctx context.Context) error {
	c.client = nil
	return nil
}

// Ping checks the connection.
func (c *OpenSearchConnector) Ping(ctx context.Context) error {
	if err := c.Connect(ctx); err != nil {
		return err
	}

	req := opensearchapi.InfoRequest{}
	resp, err := req.Do(ctx, c.client)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return fmt.Errorf("ping failed: %s", resp.String())
	}
	return nil
}

// TestConnection tests the connection to OpenSearch.
func (c *OpenSearchConnector) TestConnection(ctx context.Context) error {
	if err := c.Connect(ctx); err != nil {
		return err
	}

	return c.Ping(ctx)
}

// GetMetadata returns the metadata for the catalog.
// GetMetadata 方法用于获取OpenSearch的元数据信息
// 参数:
//   - ctx: 上下文，用于控制请求的超时和取消
//
// 返回值:
//   - map[string]any: 包含OpenSearch元数据的键值对映射
//   - error: 如果操作过程中发生错误，返回相应的错误信息
func (c *OpenSearchConnector) GetMetadata(ctx context.Context) (map[string]any, error) {
	// 检查客户端是否已初始化连接
	if c.client == nil {
		return nil, fmt.Errorf("connector not connected")
	}

	// 创建OpenSearch信息请求
	req := opensearchapi.InfoRequest{}
	// 发送请求到OpenSearch服务器
	resp, err := req.Do(ctx, c.client)
	if err != nil {
		return nil, err
	}
	// 确保响应体被关闭，以释放资源
	defer resp.Body.Close()
	// 检查响应是否包含错误
	if resp.IsError() {
		return nil, fmt.Errorf("get metadata failed: %s", resp.String())
	}

	// 用于存储解析后的元数据信息
	var info map[string]any
	// 将响应体中的JSON数据解码到info变量中
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	// 返回解析后的元数据信息
	return info, nil
}

// ListIndexes lists all indices.
func (c *OpenSearchConnector) ListIndexes(ctx context.Context) ([]*interfaces.IndexMeta, error) {
	if err := c.Connect(ctx); err != nil {
		return nil, err
	}

	req := opensearchapi.CatIndicesRequest{
		Format: "json",
	}
	if c.Config.IndexPattern != "" {
		req.Index = []string{c.Config.IndexPattern}
	}

	resp, err := req.Do(ctx, c.client)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return nil, fmt.Errorf("failed to list indices: %s", resp.String())
	}

	var catIndices []struct {
		Index     string `json:"index"`
		DocsCount string `json:"docs.count"`
		StoreSize string `json:"store.size"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&catIndices); err != nil {
		return nil, err
	}

	var indices []*interfaces.IndexMeta
	for _, idx := range catIndices {
		if strings.HasPrefix(idx.Index, ".") {
			continue // Skip system indices
		}

		indices = append(indices, &interfaces.IndexMeta{
			Name: idx.Index,
			Properties: map[string]any{
				"docs.count": idx.DocsCount,
				"store.size": idx.StoreSize,
			},
		})
	}
	return indices, nil
}

// GetIndexMeta retrieves index metadata (mappings, settings).
// GetIndexMeta 获取指定索引的元数据信息，包括映射和设置
// 参数:
//   - ctx: 上下文信息，用于控制请求的超时和取消
//   - index: 指向接口 IndexMeta 的指针，用于存储获取到的元数据
//
// 返回值:
//   - error: 如果操作过程中发生错误，则返回错误信息
func (c *OpenSearchConnector) GetIndexMeta(ctx context.Context, index *interfaces.IndexMeta) error {
	// 首先确保连接器已连接到 OpenSearch 服务
	if err := c.Connect(ctx); err != nil {
		return err
	}

	// 检查索引的属性映射是否为空，如果为空则初始化一个空的 map
	if index.Properties == nil {
		index.Properties = make(map[string]any)
	}

	// 1. Get Mappings
	if err := c.fetchMappings(ctx, index); err != nil {
		return fmt.Errorf("failed to fetch mappings: %w", err)
	}

	// 2. Get Settings
	if err := c.fetchSettings(ctx, index); err != nil {
		return fmt.Errorf("failed to fetch settings: %w", err)
	}

	return nil
}

// fetchMappings retrieves and parses index mappings.
func (c *OpenSearchConnector) fetchMappings(ctx context.Context, index *interfaces.IndexMeta) error {
	req := opensearchapi.IndicesGetMappingRequest{
		Index: []string{index.Name},
	}
	resp, err := req.Do(ctx, c.client)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return fmt.Errorf("opensearch API error: %s", resp.String())
	}

	var mappingResp map[string]struct {
		Mappings struct {
			Properties map[string]struct {
				Type string `json:"type"`
				// Add other fields as needed
			} `json:"properties"`
		} `json:"mappings"`
	}
	//{
	//	"test-index" : {
	//	"mappings" : {
	//		"properties" : {
	//			"ip_address" : {
	//				"type" : "ip",
	//				"ignore_malformed" : true
	//			}
	//		}
	//	}
	//}
	//}
	if err := json.NewDecoder(resp.Body).Decode(&mappingResp); err != nil {
		return err
	}
	// Parse mappings:这里是存储的字段元数据，包括type映射
	fieldMap := make(map[string]interfaces.FieldMeta)
	if idxData, ok := mappingResp[index.Name]; ok {
		for fieldName, props := range idxData.Mappings.Properties {
			fieldMap[fieldName] = interfaces.FieldMeta{
				Name:       fieldName,
				Type:       MapType(props.Type),
				OrigType:   props.Type,
				Searchable: true, // Default to true for now
			}
		}
	}
	index.Mapping = fieldMap
	return nil
}

// fetchSettings retrieves index settings.
func (c *OpenSearchConnector) fetchSettings(ctx context.Context, index *interfaces.IndexMeta) error {
	flatSettings := true
	req := opensearchapi.IndicesGetSettingsRequest{
		Index:        []string{index.Name},
		FlatSettings: &flatSettings,
	}
	resp, err := req.Do(ctx, c.client)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return fmt.Errorf("opensearch API error: %s", resp.String())
	}

	var settingsResp map[string]struct {
		Settings map[string]any `json:"settings"`
	}
	//{
	//	"test-index" : {
	//	"settings" : {
	//		"index.creation_date" : "1772682337114",
	//			"index.number_of_replicas" : "1",
	//			"index.number_of_shards" : "1",
	//			"index.provided_name" : "test-index",
	//			"index.uuid" : "2G4vPna8SIC0vTEzZ0NK3Q",
	//			"index.version.created" : "136287827"
	//	}
	//}
	//}
	if err := json.NewDecoder(resp.Body).Decode(&settingsResp); err != nil {
		return err
	}
	if idxData, ok := settingsResp[index.Name]; ok {
		for k, v := range idxData.Settings {
			index.Properties[k] = v
		}
	}
	return nil
}

// ExecuteIndexQuery 执行索引查询
func (c *OpenSearchConnector) ExecuteIndexQuery(ctx context.Context, catalog *interfaces.Catalog, params *interfaces.IndexQueryParams) (*interfaces.QueryResult, error) {
	if err := c.Connect(ctx); err != nil {
		return nil, err
	}

	// 1. 构建 OpenSearch 查询
	osQuery := c.buildOpenSearchQuery(params)
	dslBytes, err := json.Marshal(osQuery)

	// 2. 执行查询
	indexName := params.Resources[0].Name
	searchReq := opensearchapi.SearchRequest{
		Index: []string{indexName},
		Body:  bytes.NewReader(dslBytes),
		Size:  &params.Limit,
		From:  &params.Offset,
	}

	// 处理深度分页
	// if len(params.SearchAfter) > 0 {
	// 	searchReq.SearchAfter = params.SearchAfter
	// }

	resp, err := searchReq.Do(ctx, c.client)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return nil, fmt.Errorf("opensearch query failed: %s", resp.String())
	}

	// 3. 解析响应
	var searchResp map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, err
	}

	// 4. 转换结果
	result := &interfaces.QueryResult{
		Rows:         make([]map[string]any, 0),
		Aggregations: make(map[string]any),
	}

	// 解析 hits
	if hits, ok := searchResp["hits"].(map[string]any); ok {
		if total, ok := hits["total"].(map[string]any); ok {
			if value, ok := total["value"].(float64); ok {
				result.Total = int64(value)
			}
		}

		if hitsList, ok := hits["hits"].([]any); ok {
			for _, hit := range hitsList {
				if hitMap, ok := hit.(map[string]any); ok {
					row := make(map[string]any)
					if source, ok := hitMap["_source"].(map[string]any); ok {
						for k, v := range source {
							row[k] = v
						}
					}
					// 添加 _id 字段
					if id, ok := hitMap["_id"].(string); ok {
						row["_id"] = id
					}
					result.Rows = append(result.Rows, row)
				}
			}
		}
	}

	// 解析聚合结果
	if aggs, ok := searchResp["aggregations"].(map[string]any); ok {
		result.Aggregations = aggs
	}

	return result, nil
}

// buildOpenSearchQuery 构建 OpenSearch 查询 DSL
func (c *OpenSearchConnector) buildOpenSearchQuery(params *interfaces.IndexQueryParams) map[string]any {
	query := make(map[string]any)

	// 1. 构建 query/filter
	if params.ActualFilterCond != nil {
		query["query"] = c.buildFilterQuery(params.ActualFilterCond)
	} else {
		query["query"] = map[string]any{
			"match_all": map[string]any{},
		}
	}

	// 2. 构建排序
	if len(params.Sort) > 0 {
		sortClauses := make([]map[string]any, 0)
		for _, sf := range params.Sort {
			sortClauses = append(sortClauses, map[string]any{
				sf.Field: map[string]any{"order": sf.Direction},
			})
		}
		query["sort"] = sortClauses
	}

	// 3. 构建聚合
	if len(params.Aggregations) > 0 {
		query["aggs"] = params.Aggregations
	}

	// 4. 处理输出字段
	if len(params.OutputFields) > 0 && !contains(params.OutputFields, "*") {
		fields := make([]string, 0)
		for _, f := range params.OutputFields {
			// 移除表别名前缀
			if idx := strings.LastIndex(f, "."); idx > 0 {
				fields = append(fields, f[idx+1:])
			} else {
				fields = append(fields, f)
			}
		}
		query["_source"] = fields
	}

	return query
}

// buildFilterQuery 根据过滤条件构建 OpenSearch 查询
func (c *OpenSearchConnector) buildFilterQuery(cond interfaces.FilterCondition) map[string]any {
	if cond == nil {
		return map[string]any{
			"match_all": map[string]any{},
		}
	}

	// 获取过滤条件的配置
	cfg := cond.GetConfig()
	if cfg == nil {
		return map[string]any{
			"match_all": map[string]any{},
		}
	}

	// 处理 _id 字段的特殊查询
	if cfg.Name == "_id" || strings.HasSuffix(cfg.Name, "._id") {
		return c.buildIdQuery(cfg)
	}

	// 处理其他字段的查询
	return c.buildFieldQuery(cond)
}

// buildIdQuery 构建基于 _id 的查询
func (c *OpenSearchConnector) buildIdQuery(cfg *interfaces.FilterCondCfg) map[string]any {
	switch cfg.Operation {
	case "eq":
		// 使用 term 查询 _id
		return map[string]any{
			"term": map[string]any{
				"_id": cfg.Value,
			},
		}
	case "in":
		// 使用 ids 查询
		return map[string]any{
			"ids": map[string]any{
				"values": cfg.Value,
			},
		}
	default:
		// 默认使用 term 查询
		return map[string]any{
			"term": map[string]any{
				"_id": cfg.Value,
			},
		}
	}
}

// buildFieldQuery 构建普通字段的查询
func (c *OpenSearchConnector) buildFieldQuery(cond interfaces.FilterCondition) map[string]any {
	cfg := cond.GetConfig()
	if cfg == nil {
		return map[string]any{
			"match_all": map[string]any{},
		}
	}

	// 移除字段名中的表别名前缀
	fieldName := cfg.Name
	if idx := strings.LastIndex(fieldName, "."); idx > 0 {
		fieldName = fieldName[idx+1:]
	}

	switch cfg.Operation {
	case "eq":
		return map[string]any{
			"term": map[string]any{
				fieldName: cfg.Value,
			},
		}
	case "ne":
		return map[string]any{
			"bool": map[string]any{
				"must_not": []map[string]any{
					{
						"term": map[string]any{
							fieldName: cfg.Value,
						},
					},
				},
			},
		}
	case "gt":
		return map[string]any{
			"range": map[string]any{
				fieldName: map[string]any{
					"gt": cfg.Value,
				},
			},
		}
	case "gte":
		return map[string]any{
			"range": map[string]any{
				fieldName: map[string]any{
					"gte": cfg.Value,
				},
			},
		}
	case "lt":
		return map[string]any{
			"range": map[string]any{
				fieldName: map[string]any{
					"lt": cfg.Value,
				},
			},
		}
	case "lte":
		return map[string]any{
			"range": map[string]any{
				fieldName: map[string]any{
					"lte": cfg.Value,
				},
			},
		}
	case "in":
		return map[string]any{
			"terms": map[string]any{
				fieldName: cfg.Value,
			},
		}
	case "not_in":
		return map[string]any{
			"bool": map[string]any{
				"must_not": []map[string]any{
					{
						"terms": map[string]any{
							fieldName: cfg.Value,
						},
					},
				},
			},
		}
	case "and":
		if len(cfg.SubConds) > 0 {
			mustClauses := make([]map[string]any, 0, len(cfg.SubConds))
			for _, subCond := range cfg.SubConds {
				subFilterCond, err := filter_condition.NewFilterCondition(context.Background(), subCond, nil)
				if err != nil {
					continue
				}
				mustClauses = append(mustClauses, c.buildFilterQuery(subFilterCond))
			}
			return map[string]any{
				"bool": map[string]any{
					"must": mustClauses,
				},
			}
		}
	case "or":
		if len(cfg.SubConds) > 0 {
			shouldClauses := make([]map[string]any, 0, len(cfg.SubConds))
			for _, subCond := range cfg.SubConds {
				subFilterCond, err := filter_condition.NewFilterCondition(context.Background(), subCond, nil)
				if err != nil {
					continue
				}
				shouldClauses = append(shouldClauses, c.buildFilterQuery(subFilterCond))
			}
			return map[string]any{
				"bool": map[string]any{
					"should":               shouldClauses,
					"minimum_should_match": 1,
				},
			}
		}
	}

	// 默认返回 match_all
	return map[string]any{
		"match_all": map[string]any{},
	}
}

// contains 检查字符串切片是否包含某个元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
