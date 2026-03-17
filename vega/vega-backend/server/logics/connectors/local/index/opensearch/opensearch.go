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
	"reflect"
	"strings"
	"time"

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

// ExecuteQuery executes a query on the OpenSearch index.
// ExecuteQuery 执行OpenSearch查询并返回结果
// 参数:
//   - ctx: 上下文信息
//   - resource: 资源信息，包含索引名称等
//   - params: 查询参数，包括输出字段、排序、分页等
//
// 返回值:
//   - *interfaces.QueryResult: 查询结果，包含行数据和总数
//   - error: 错误信息
func (c *OpenSearchConnector) ExecuteQuery(ctx context.Context, resource *interfaces.Resource,
	params *interfaces.ResourceDataQueryParams) (*interfaces.QueryResult, error) {

	// Ensure we have a connection
	if err := c.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to OpenSearch: %w", err)
	}

	// Get the index name from the resource
	indexName := resource.SourceIdentifier
	if indexName == "" {
		return nil, fmt.Errorf("index name is empty in resource")
	}

	// Build the OpenSearch query
	query := map[string]any{
		"query": map[string]any{
			"match_all": map[string]any{},
		},
		"from": 0,
		"size": 100,
	}

	// Handle output fields (_source)
	if params != nil && len(params.OutputFields) > 0 {
		// Filter out _score field as it's not a source field but a calculated score
		sourceFields := []string{}
		includeScore := false
		for _, field := range params.OutputFields {
			if field != "_score" {
				sourceFields = append(sourceFields, field)
			} else {
				includeScore = true
			}
		}
		if len(sourceFields) > 0 {
			query["_source"] = sourceFields
		}
		// Ensure track_scores is true to get _score when needed
		if includeScore {
			query["track_scores"] = true
		}
	}

	// Handle sorting
	if params != nil && len(params.Sort) > 0 {
		sort := make([]map[string]any, 0, len(params.Sort))
		for _, s := range params.Sort {
			sort = append(sort, map[string]any{
				s.Field: map[string]any{
					"order": s.Direction,
				},
			})
		}
		query["sort"] = sort
	}

	// Handle pagination
	if params != nil {
		if params.Offset > 0 && params.SearchAfter == nil {
			query["from"] = params.Offset
		}

		if params.Limit > 0 {
			query["size"] = params.Limit
		}

		// Handle search_after
		if len(params.SearchAfter) > 0 {
			query["search_after"] = params.SearchAfter
		}
	}

	// Handle filter conditions
	if params != nil && params.ActualFilterCond != nil {
		// Build filter condition query
		filterQuery, err := c.buildFilterQuery(params.ActualFilterCond, resource.SchemaDefinition)
		if err != nil {
			return nil, fmt.Errorf("failed to build filter query: %w", err)
		}
		if filterQuery != nil {
			query["query"] = filterQuery
		}
	}

	// Serialize query
	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize query: %w", err)
	}

	// Execute search request
	req := opensearchapi.SearchRequest{
		Index: []string{indexName},
		Body:  bytes.NewReader(queryJSON),
	}

	resp, err := req.Do(ctx, c.client)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return nil, fmt.Errorf("search failed: %s", resp.String())
	}

	// Parse response
	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode search result: %w", err)
	}

	hits, ok := result["hits"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid search result format: missing hits")
	}

	total, ok := hits["total"].(map[string]any)["value"].(float64)
	if !ok {
		total = 0
	}

	hitsArray, ok := hits["hits"].([]any)
	if !ok {
		return &interfaces.QueryResult{
			Rows:  []map[string]any{},
			Total: int64(total),
		}, nil
	}

	// Extract documents from hits
	documents := make([]map[string]any, 0, len(hitsArray))
	for _, hit := range hitsArray {
		hitMap, ok := hit.(map[string]any)
		if !ok {
			continue
		}

		source, ok := hitMap["_source"].(map[string]any)
		if !ok {
			continue
		}

		source["_id"] = hitMap["_id"]
		// Add _score field if present
		if score, ok := hitMap["_score"].(float64); ok {
			source["_score"] = score
		}
		documents = append(documents, source)
	}

	return &interfaces.QueryResult{
		Rows:  documents,
		Total: int64(total),
	}, nil
}

// buildFilterQuery builds an OpenSearch filter query from filter conditions
func (c *OpenSearchConnector) buildFilterQuery(filterCond interfaces.FilterCondition, schemaDefinition []*interfaces.Property) (map[string]any, error) {
	if filterCond == nil {
		return nil, nil
	}

	// Create a field map from schema definition
	fieldsMap := make(map[string]*interfaces.Property)
	for _, field := range schemaDefinition {
		fieldsMap[field.Name] = field
	}

	// Convert filter condition to OpenSearch query format based on operation type
	operation := filterCond.GetOperation()

	switch operation {
	case "and":
		return c.buildAndQuery(filterCond, fieldsMap)
	case "or":
		return c.buildOrQuery(filterCond, fieldsMap)
	case "==", "eq":
		return c.buildEqualQuery(filterCond, fieldsMap)
	case "!=", "not_eq":
		return c.buildNotEqualQuery(filterCond, fieldsMap)
	case ">", "gt":
		return c.buildGtQuery(filterCond, fieldsMap)
	case ">=", "gte":
		return c.buildGteQuery(filterCond, fieldsMap)
	case "<", "lt":
		return c.buildLtQuery(filterCond, fieldsMap)
	case "<=", "lte":
		return c.buildLteQuery(filterCond, fieldsMap)
	case "in":
		return c.buildInQuery(filterCond, fieldsMap)
	case "not_in":
		return c.buildNotInQuery(filterCond, fieldsMap)
	case "like":
		return c.buildLikeQuery(filterCond, fieldsMap)
	case "not_like":
		return c.buildNotLikeQuery(filterCond, fieldsMap)
	case "contain":
		return c.buildContainQuery(filterCond, fieldsMap)
	case "not_contain":
		return c.buildNotContainQuery(filterCond, fieldsMap)
	case "range":
		return c.buildRangeQuery(filterCond, fieldsMap)
	case "out_range":
		return c.buildOutRangeQuery(filterCond, fieldsMap)
	case "exist":
		return c.buildExistQuery(filterCond, fieldsMap)
	case "not_exist":
		return c.buildNotExistQuery(filterCond, fieldsMap)
	case "empty":
		return c.buildEmptyQuery(filterCond, fieldsMap)
	case "not_empty":
		return c.buildNotEmptyQuery(filterCond, fieldsMap)
	case "regex":
		return c.buildRegexQuery(filterCond, fieldsMap)
	case "match":
		return c.buildMatchQuery(filterCond, fieldsMap)
	case "match_phrase":
		return c.buildMatchPhraseQuery(filterCond, fieldsMap)
	case "prefix":
		return c.buildPrefixQuery(filterCond, fieldsMap)
	case "not_prefix":
		return c.buildNotPrefixQuery(filterCond, fieldsMap)
	case "null":
		return c.buildNullQuery(filterCond, fieldsMap)
	case "not_null":
		return c.buildNotNullQuery(filterCond, fieldsMap)
	case "true":
		return c.buildTrueQuery(filterCond, fieldsMap)
	case "false":
		return c.buildFalseQuery(filterCond, fieldsMap)
	case "before":
		return c.buildBeforeQuery(filterCond, fieldsMap)
	case "current":
		return c.buildCurrentQuery(filterCond, fieldsMap)
	case "between":
		return c.buildBetweenQuery(filterCond, fieldsMap)
	case "knn_vector":
		return c.buildKnnVectorQuery(filterCond, fieldsMap)
	case "multi_match":
		return c.buildMultiMatchQuery(filterCond, fieldsMap)
	default:
		// Default to match_all for unsupported operations
		return map[string]any{
			"match_all": map[string]any{},
		}, nil
	}
}

// getSubConditions extracts sub-conditions from a filter condition
func (c *OpenSearchConnector) getSubConditions(filterCond interfaces.FilterCondition) ([]interfaces.FilterCondition, error) {
	// Use reflection to access SubConds field
	val := reflect.ValueOf(filterCond).Elem()
	field := val.FieldByName("SubConds")
	if !field.IsValid() {
		return nil, fmt.Errorf("filter condition does not have SubConds field")
	}

	subConds := make([]interfaces.FilterCondition, field.Len())
	for i := 0; i < field.Len(); i++ {
		subCond, ok := field.Index(i).Interface().(interfaces.FilterCondition)
		if !ok {
			return nil, fmt.Errorf("sub-condition at index %d is not a FilterCondition", i)
		}
		subConds[i] = subCond
	}

	return subConds, nil
}

// getFieldNameAndValue extracts field name and value from a filter condition
func (c *OpenSearchConnector) getFieldNameAndValue(filterCond interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (string, any, error) {
	// Use reflection to access Cfg field
	val := reflect.ValueOf(filterCond).Elem()
	cfgField := val.FieldByName("Cfg")
	if !cfgField.IsValid() {
		return "", nil, fmt.Errorf("filter condition does not have Cfg field")
	}

	cfg, ok := cfgField.Interface().(*interfaces.FilterCondCfg)
	if !ok {
		return "", nil, fmt.Errorf("Cfg field is not a FilterCondCfg")
	}

	fieldName := cfg.Name
	if fieldName == "" {
		return "", nil, fmt.Errorf("field name is empty")
	}

	// Use RealValue if available, otherwise use Value
	value := cfg.RealValue
	if value == nil {
		value = cfg.Value
	}

	return fieldName, value, nil
}

// buildAndQuery builds an AND query
func (c *OpenSearchConnector) buildAndQuery(filterCond interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {
	subConds, err := c.getSubConditions(filterCond)
	if err != nil {
		return nil, err
	}

	if len(subConds) == 0 {
		return map[string]any{
			"match_all": map[string]any{},
		}, nil
	}

	mustClauses := make([]map[string]any, 0, len(subConds))
	for _, subCond := range subConds {
		clause, err := c.buildFilterQuery(subCond, nil)
		if err != nil {
			return nil, err
		}
		mustClauses = append(mustClauses, clause)
	}

	return map[string]any{
		"bool": map[string]any{
			"must": mustClauses,
		},
	}, nil
}

// buildOrQuery builds an OR query
func (c *OpenSearchConnector) buildOrQuery(filterCond interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {
	subConds, err := c.getSubConditions(filterCond)
	if err != nil {
		return nil, err
	}

	if len(subConds) == 0 {
		return map[string]any{
			"match_all": map[string]any{},
		}, nil
	}

	shouldClauses := make([]map[string]any, 0, len(subConds))
	for _, subCond := range subConds {
		clause, err := c.buildFilterQuery(subCond, nil)
		if err != nil {
			return nil, err
		}
		shouldClauses = append(shouldClauses, clause)
	}

	return map[string]any{
		"bool": map[string]any{
			"should":               shouldClauses,
			"minimum_should_match": 1,
		},
	}, nil
}

// buildEqualQuery builds an equality query
func (c *OpenSearchConnector) buildEqualQuery(filterCond interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {
	fieldName, value, err := c.getFieldNameAndValue(filterCond, fieldsMap)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"term": map[string]any{
			fieldName: value,
		},
	}, nil
}

// buildNotEqualQuery builds a not equality query
func (c *OpenSearchConnector) buildNotEqualQuery(filterCond interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {
	fieldName, value, err := c.getFieldNameAndValue(filterCond, fieldsMap)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"bool": map[string]any{
			"must_not": []map[string]any{
				{
					"term": map[string]any{
						fieldName: value,
					},
				},
			},
		},
	}, nil
}

// buildGtQuery builds a greater than query
func (c *OpenSearchConnector) buildGtQuery(filterCond interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {
	fieldName, value, err := c.getFieldNameAndValue(filterCond, fieldsMap)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"range": map[string]any{
			fieldName: map[string]any{
				"gt": value,
			},
		},
	}, nil
}

// buildGteQuery builds a greater than or equal query
func (c *OpenSearchConnector) buildGteQuery(filterCond interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {
	fieldName, value, err := c.getFieldNameAndValue(filterCond, fieldsMap)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"range": map[string]any{
			fieldName: map[string]any{
				"gte": value,
			},
		},
	}, nil
}

// buildLtQuery builds a less than query
func (c *OpenSearchConnector) buildLtQuery(filterCond interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {
	fieldName, value, err := c.getFieldNameAndValue(filterCond, fieldsMap)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"range": map[string]any{
			fieldName: map[string]any{
				"lt": value,
			},
		},
	}, nil
}

// buildLteQuery builds a less than or equal query
func (c *OpenSearchConnector) buildLteQuery(filterCond interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {
	fieldName, value, err := c.getFieldNameAndValue(filterCond, fieldsMap)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"range": map[string]any{
			fieldName: map[string]any{
				"lte": value,
			},
		},
	}, nil
}

// buildInQuery builds an IN query
func (c *OpenSearchConnector) buildInQuery(filterCond interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {
	fieldName, value, err := c.getFieldNameAndValue(filterCond, fieldsMap)
	if err != nil {
		return nil, err
	}

	// Ensure value is a slice
	values, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("IN query requires an array of values")
	}

	return map[string]any{
		"terms": map[string]any{
			fieldName: values,
		},
	}, nil
}

// buildNotInQuery builds a NOT IN query
func (c *OpenSearchConnector) buildNotInQuery(filterCond interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {
	fieldName, value, err := c.getFieldNameAndValue(filterCond, fieldsMap)
	if err != nil {
		return nil, err
	}

	// Ensure value is a slice
	values, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("NOT IN query requires an array of values")
	}

	return map[string]any{
		"bool": map[string]any{
			"must_not": []map[string]any{
				{
					"terms": map[string]any{
						fieldName: values,
					},
				},
			},
		},
	}, nil
}

// buildLikeQuery builds a LIKE query
func (c *OpenSearchConnector) buildLikeQuery(filterCond interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {
	fieldName, value, err := c.getFieldNameAndValue(filterCond, fieldsMap)
	if err != nil {
		return nil, err
	}

	// Convert SQL LIKE pattern to wildcard pattern
	valueStr, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("LIKE query requires a string value")
	}

	// Convert SQL LIKE wildcards to OpenSearch wildcards
	// SQL: % -> OpenSearch: *
	// SQL: _ -> OpenSearch: ?
	pattern := strings.ReplaceAll(valueStr, "%", "*")
	pattern = strings.ReplaceAll(pattern, "_", "?")

	return map[string]any{
		"wildcard": map[string]any{
			fieldName: pattern,
		},
	}, nil
}

// buildNotLikeQuery builds a NOT LIKE query
func (c *OpenSearchConnector) buildNotLikeQuery(filterCond interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {
	fieldName, value, err := c.getFieldNameAndValue(filterCond, fieldsMap)
	if err != nil {
		return nil, err
	}

	// Convert SQL LIKE pattern to wildcard pattern
	valueStr, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("NOT LIKE query requires a string value")
	}

	// Convert SQL LIKE wildcards to OpenSearch wildcards
	pattern := strings.ReplaceAll(valueStr, "%", "*")
	pattern = strings.ReplaceAll(pattern, "_", "?")

	return map[string]any{
		"bool": map[string]any{
			"must_not": []map[string]any{
				{
					"wildcard": map[string]any{
						fieldName: pattern,
					},
				},
			},
		},
	}, nil
}

// buildContainQuery builds a CONTAIN query
func (c *OpenSearchConnector) buildContainQuery(filterCond interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {
	fieldName, value, err := c.getFieldNameAndValue(filterCond, fieldsMap)
	if err != nil {
		return nil, err
	}

	// Check if value is an array/slice
	values, isArray := value.([]any)
	if isArray && len(values) > 0 {
		// For array values, use a bool query with should clauses
		// This matches documents that contain any of the values
		shouldClauses := make([]map[string]any, 0, len(values))
		for _, v := range values {
			shouldClauses = append(shouldClauses, map[string]any{
				"match": map[string]any{
					fieldName: v,
				},
			})
		}
		return map[string]any{
			"bool": map[string]any{
				"should":               shouldClauses,
				"minimum_should_match": 1,
			},
		}, nil
	}

	// For single value, use simple match query
	return map[string]any{
		"match": map[string]any{
			fieldName: value,
		},
	}, nil
}

// buildNotContainQuery builds a NOT CONTAIN query
func (c *OpenSearchConnector) buildNotContainQuery(filterCond interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {
	fieldName, value, err := c.getFieldNameAndValue(filterCond, fieldsMap)
	if err != nil {
		return nil, err
	}

	// Check if value is an array/slice
	values, isArray := value.([]any)
	if isArray && len(values) > 0 {
		// For array values, use a bool query with must_not clauses
		// This matches documents that do not contain any of the values
		mustNotClauses := make([]map[string]any, 0, len(values))
		for _, v := range values {
			mustNotClauses = append(mustNotClauses, map[string]any{
				"match": map[string]any{
					fieldName: v,
				},
			})
		}
		return map[string]any{
			"bool": map[string]any{
				"must_not": mustNotClauses,
			},
		}, nil
	}

	// For single value, use simple bool query with must_not
	return map[string]any{
		"bool": map[string]any{
			"must_not": []map[string]any{
				{
					"match": map[string]any{
						fieldName: value,
					},
				},
			},
		},
	}, nil
}

// buildRangeQuery builds a RANGE query
func (c *OpenSearchConnector) buildRangeQuery(filterCond interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {
	fieldName, value, err := c.getFieldNameAndValue(filterCond, fieldsMap)
	if err != nil {
		return nil, err
	}

	// Ensure value is a slice with 2 elements
	values, ok := value.([]any)
	if !ok || len(values) != 2 {
		return nil, fmt.Errorf("RANGE query requires an array with 2 values")
	}

	return map[string]any{
		"range": map[string]any{
			fieldName: map[string]any{
				"gte": values[0],
				"lte": values[1],
			},
		},
	}, nil
}

// buildOutRangeQuery builds an OUT RANGE query
func (c *OpenSearchConnector) buildOutRangeQuery(filterCond interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {
	fieldName, value, err := c.getFieldNameAndValue(filterCond, fieldsMap)
	if err != nil {
		return nil, err
	}

	// Ensure value is a slice with 2 elements
	values, ok := value.([]any)
	if !ok || len(values) != 2 {
		return nil, fmt.Errorf("OUT RANGE query requires an array with 2 values")
	}

	return map[string]any{
		"bool": map[string]any{
			"should": []map[string]any{
				{
					"range": map[string]any{
						fieldName: map[string]any{
							"lt": values[0],
						},
					},
				},
				{
					"range": map[string]any{
						fieldName: map[string]any{
							"gt": values[1],
						},
					},
				},
			},
			"minimum_should_match": 1,
		},
	}, nil
}

// buildExistQuery builds an EXIST query
func (c *OpenSearchConnector) buildExistQuery(filterCond interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {
	fieldName, _, err := c.getFieldNameAndValue(filterCond, fieldsMap)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"exists": map[string]any{
			"field": fieldName,
		},
	}, nil
}

// buildNotExistQuery builds a NOT EXIST query
func (c *OpenSearchConnector) buildNotExistQuery(filterCond interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {
	fieldName, _, err := c.getFieldNameAndValue(filterCond, fieldsMap)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"bool": map[string]any{
			"must_not": []map[string]any{
				{
					"exists": map[string]any{
						"field": fieldName,
					},
				},
			},
		},
	}, nil
}

// buildEmptyQuery builds an EMPTY query
func (c *OpenSearchConnector) buildEmptyQuery(filterCond interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {
	fieldName, _, err := c.getFieldNameAndValue(filterCond, fieldsMap)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"bool": map[string]any{
			"must_not": []map[string]any{
				{
					"exists": map[string]any{
						"field": fieldName,
					},
				},
			},
		},
	}, nil
}

// buildNotEmptyQuery builds a NOT EMPTY query
func (c *OpenSearchConnector) buildNotEmptyQuery(filterCond interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {
	fieldName, _, err := c.getFieldNameAndValue(filterCond, fieldsMap)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"exists": map[string]any{
			"field": fieldName,
		},
	}, nil
}

// buildRegexQuery builds a REGEX query
func (c *OpenSearchConnector) buildRegexQuery(filterCond interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {
	fieldName, value, err := c.getFieldNameAndValue(filterCond, fieldsMap)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"regexp": map[string]any{
			fieldName: value,
		},
	}, nil
}

// buildMatchQuery builds a MATCH query
func (c *OpenSearchConnector) buildMatchQuery(filterCond interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {
	fieldName, value, err := c.getFieldNameAndValue(filterCond, fieldsMap)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"match": map[string]any{
			fieldName: value,
		},
	}, nil
}

// buildMatchPhraseQuery builds a MATCH PHRASE query
func (c *OpenSearchConnector) buildMatchPhraseQuery(filterCond interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {
	fieldName, value, err := c.getFieldNameAndValue(filterCond, fieldsMap)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"match_phrase": map[string]any{
			fieldName: value,
		},
	}, nil
}

// buildPrefixQuery builds a PREFIX query
func (c *OpenSearchConnector) buildPrefixQuery(filterCond interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {
	fieldName, value, err := c.getFieldNameAndValue(filterCond, fieldsMap)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"prefix": map[string]any{
			fieldName: value,
		},
	}, nil
}

// buildNotPrefixQuery builds a NOT PREFIX query
func (c *OpenSearchConnector) buildNotPrefixQuery(filterCond interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {
	fieldName, value, err := c.getFieldNameAndValue(filterCond, fieldsMap)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"bool": map[string]any{
			"must_not": []map[string]any{
				{
					"prefix": map[string]any{
						fieldName: value,
					},
				},
			},
		},
	}, nil
}

// buildNullQuery builds a NULL query
func (c *OpenSearchConnector) buildNullQuery(filterCond interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {
	fieldName, _, err := c.getFieldNameAndValue(filterCond, fieldsMap)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"bool": map[string]any{
			"must_not": []map[string]any{
				{
					"exists": map[string]any{
						"field": fieldName,
					},
				},
			},
		},
	}, nil
}

// buildNotNullQuery builds a NOT NULL query
func (c *OpenSearchConnector) buildNotNullQuery(filterCond interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {
	fieldName, _, err := c.getFieldNameAndValue(filterCond, fieldsMap)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"exists": map[string]any{
			"field": fieldName,
		},
	}, nil
}

// buildTrueQuery builds a TRUE query
func (c *OpenSearchConnector) buildTrueQuery(filterCond interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {
	fieldName, _, err := c.getFieldNameAndValue(filterCond, fieldsMap)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"term": map[string]any{
			fieldName: true,
		},
	}, nil
}

// buildFalseQuery builds a FALSE query
func (c *OpenSearchConnector) buildFalseQuery(filterCond interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {
	fieldName, _, err := c.getFieldNameAndValue(filterCond, fieldsMap)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"term": map[string]any{
			fieldName: false,
		},
	}, nil
}

// buildBeforeQuery builds a BEFORE query
func (c *OpenSearchConnector) buildBeforeQuery(filterCond interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {
	fieldName, value, err := c.getFieldNameAndValue(filterCond, fieldsMap)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"range": map[string]any{
			fieldName: map[string]any{
				"lt": value,
			},
		},
	}, nil
}

// buildCurrentQuery builds a CURRENT query
func (c *OpenSearchConnector) buildCurrentQuery(filterCond interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {
	// CURRENT query is typically used for date/time fields
	// It matches records where the field value equals the current time/date
	// For date fields, it matches records with today's date
	// For datetime fields, it matches records with the current time

	fieldName, _, err := c.getFieldNameAndValue(filterCond, fieldsMap)
	if err != nil {
		return nil, err
	}

	// Get the field type from schema definition
	var fieldType string
	if field, ok := fieldsMap[fieldName]; ok {
		fieldType = field.Type
	}

	// Build the query based on field type
	switch fieldType {
	case "date":
		// For date fields, match today's date
		today := time.Now().Format("2006-01-02")
		return map[string]any{
			"term": map[string]any{
				fieldName: today,
			},
		}, nil
	case "datetime", "timestamp":
		// For datetime fields, match the current time within a reasonable range
		now := time.Now()
		// Match records within the last minute
		oneMinuteAgo := now.Add(-time.Minute).Format(time.RFC3339)
		nowStr := now.Format(time.RFC3339)
		return map[string]any{
			"range": map[string]any{
				fieldName: map[string]any{
					"gte": oneMinuteAgo,
					"lte": nowStr,
				},
			},
		}, nil
	default:
		// For other field types, just match the current value
		// This might not make sense for non-date fields, but we'll handle it anyway
		now := time.Now().Unix()
		return map[string]any{
			"term": map[string]any{
				fieldName: now,
			},
		}, nil
	}
}

// buildBetweenQuery builds a BETWEEN query
func (c *OpenSearchConnector) buildBetweenQuery(filterCond interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {
	fieldName, value, err := c.getFieldNameAndValue(filterCond, fieldsMap)
	if err != nil {
		return nil, err
	}

	// Ensure value is a slice with 2 elements
	values, ok := value.([]any)
	if !ok || len(values) != 2 {
		return nil, fmt.Errorf("BETWEEN query requires an array with 2 values")
	}

	return map[string]any{
		"range": map[string]any{
			fieldName: map[string]any{
				"gte": values[0],
				"lte": values[1],
			},
		},
	}, nil
}

// buildKnnVectorQuery builds a KNN VECTOR query
func (c *OpenSearchConnector) buildKnnVectorQuery(filterCond interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {
	fieldName, value, err := c.getFieldNameAndValue(filterCond, fieldsMap)
	if err != nil {
		return nil, err
	}

	// KNN vector query requires a vector value and k (number of neighbors)
	// The value should be a map with "vector" and "k" keys
	valueMap, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("KNN VECTOR query requires a map with vector and k keys")
	}

	vector, ok := valueMap["vector"].([]any)
	if !ok {
		return nil, fmt.Errorf("KNN VECTOR query requires a vector array")
	}

	k, ok := valueMap["k"].(float64)
	if !ok {
		return nil, fmt.Errorf("KNN VECTOR query requires a k value")
	}

	// Build the KNN query
	knnQuery := map[string]any{
		"field":          fieldName,
		"query_vector":   vector,
		"k":              int(k),
		"num_candidates": int(k) * 10, // Use 10x candidates for better accuracy
	}

	// Check if there are sub-conditions (filter conditions for KNN)
	subConds, err := c.getSubConditions(filterCond)
	if err == nil && len(subConds) > 0 {
		// Build filter clauses for each sub-condition
		filterClauses := make([]map[string]any, 0, len(subConds))
		for _, subCond := range subConds {
			clause, err := c.buildFilterQuery(subCond, nil)
			if err != nil {
				return nil, err
			}
			filterClauses = append(filterClauses, clause)
		}

		// Add filter to KNN query
		if len(filterClauses) == 1 {
			knnQuery["filter"] = filterClauses[0]
		} else {
			knnQuery["filter"] = map[string]any{
				"bool": map[string]any{
					"must": filterClauses,
				},
			}
		}
	}

	return map[string]any{
		"knn": knnQuery,
	}, nil
}

// buildMultiMatchQuery builds a MULTI MATCH query
func (c *OpenSearchConnector) buildMultiMatchQuery(filterCond interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {
	// Use reflection to access Cfg field
	val := reflect.ValueOf(filterCond).Elem()
	cfgField := val.FieldByName("Cfg")
	if !cfgField.IsValid() {
		return nil, fmt.Errorf("filter condition does not have Cfg field")
	}

	cfg, ok := cfgField.Interface().(*interfaces.FilterCondCfg)
	if !ok {
		return nil, fmt.Errorf("Cfg field is not a FilterCondCfg")
	}

	// Use RealValue if available, otherwise use Value
	value := cfg.RealValue
	if value == nil {
		value = cfg.Value
	}

	// Get the fields to search in from the RemainCfg
	fields, ok := cfg.RemainCfg["fields"].([]any)
	if !ok {
		return nil, fmt.Errorf("MULTI MATCH query requires a fields array in RemainCfg")
	}

	// Convert fields to strings
	fieldStrings := make([]string, 0, len(fields))
	for _, field := range fields {
		fieldStr, ok := field.(string)
		if !ok {
			return nil, fmt.Errorf("MULTI MATCH query fields must be strings")
		}
		fieldStrings = append(fieldStrings, fieldStr)
	}

	return map[string]any{
		"multi_match": map[string]any{
			"query":  value,
			"fields": fieldStrings,
			"type":   "best_fields",
		},
	}, nil
}
