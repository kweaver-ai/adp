// Package dataset provides dataset data access implementations.
package dataset

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"

	"vega-backend/common"
	"vega-backend/interfaces"
	"vega-backend/logics/filter_condition"
)

var (
	daOnce sync.Once
	da     interfaces.DatasetAccess
)

type datasetAccess struct {
	appSetting *common.AppSetting
	osClient   *opensearch.Client
}

// NewDatasetAccess creates a new DatasetAccess.
func NewDatasetAccess(appSetting *common.AppSetting) interfaces.DatasetAccess {
	daOnce.Do(func() {
		da = &datasetAccess{
			appSetting: appSetting,
		}
	})
	return da
}

// Create a new Dataset.
func (da *datasetAccess) Create(ctx context.Context, name string, schemaDefinition []*interfaces.Property) error {
	// 获取 OpenSearch 客户端
	client, err := da.getOpenSearchClient()
	if err != nil {
		return err
	}

	// 构建索引映射
	mappings := map[string]any{
		"properties": map[string]any{},
	}

	mapping := map[string]any{
		"mappings": mappings,
	}

	mapping["settings"] = map[string]any{
		"index": map[string]any{
			"number_of_shards":   1,
			"number_of_replicas": 0,
		},
	}

	// 检查是否有vector字段
	hasVectorField := false
	// 根据 schemaDefinition 添加字段映射
	properties := mapping["mappings"].(map[string]any)["properties"].(map[string]any)
	for _, column := range schemaDefinition {
		fieldType := column.Type
		switch column.Type {
		case "integer":
			fieldType = "long"
		case "unsigned_integer":
			fieldType = "unsigned_long"
		case "float":
			fieldType = "double"
		case "decimal":
			fieldType = "scaled_float"
		case "string":
			fieldType = "keyword"
		case "datetime":
			fieldType = "date"
		case "time":
			fieldType = "keyword"
		case "json":
			fieldType = "object"
		case "vector":
			hasVectorField = true
			fieldType = "knn_vector"
		case "point":
			fieldType = "geo_point"
		case "shape":
			fieldType = "geo_shape"
		default:
			// 保持 fieldType 不变
		}
		// 创建字段属性映射
		fieldProps := map[string]any{
			"type": fieldType,
		}
		// 为decimal类型添加scaling_factor参数
		if column.Type == "decimal" {
			fieldProps["scaling_factor"] = 1000000000000000000.0 // 18位小数
		}
		if len(column.Features) > 0 {
			for _, feature := range column.Features {
				if feature.Config != nil {
					for k, v := range feature.Config {
						switch feature.FeatureType {
						case "keyword":
							if column.Type == "text" {
								// 添加子字段
								fieldProps["fields"] = map[string]any{
									"keyword": map[string]any{
										"type": "keyword",
									},
								}
							} else {
								fieldProps[k] = v
							}
						case "vector":
							fieldProps[k] = v
						case "fulltext":
							continue
						default:
							return fmt.Errorf("unsupported feature type: %s", feature.FeatureType)
						}
					}
				}
			}
		}
		properties[column.Name] = fieldProps
	}

	// 如果有vector字段，开启knn
	if hasVectorField {
		indexSettings := mapping["settings"].(map[string]any)["index"].(map[string]any)
		indexSettings["knn"] = true
	}

	// 检查索引是否存在
	existsReq := opensearchapi.IndicesExistsRequest{
		Index: []string{name},
	}

	existsResp, err := existsReq.Do(ctx, client)
	if err != nil {
		return err
	}
	defer existsResp.Body.Close()

	// 如果索引不存在，创建索引
	if existsResp.StatusCode == 404 {
		data, err := json.Marshal(mapping)
		if err != nil {
			return err
		}
		bodyReader := strings.NewReader(string(data))
		da.printRequestBody(bodyReader)
		createReq := opensearchapi.IndicesCreateRequest{
			Index: name,
			Body:  bodyReader,
		}

		createResp, err := createReq.Do(ctx, client)
		if err != nil {
			return err
		}
		defer createResp.Body.Close()

		if createResp.IsError() {
			return fmt.Errorf("failed to create index: %s", createResp.String())
		}
	}

	return nil
}

// Update updates a Dataset.
func (da *datasetAccess) Update(ctx context.Context, name string, schemaDefinition []*interfaces.Property) error {
	// 获取 OpenSearch 客户端
	client, err := da.getOpenSearchClient()
	if err != nil {
		return err
	}
	// 检查索引是否存在
	existsReq := opensearchapi.IndicesExistsRequest{
		Index: []string{name},
	}

	existsResp, err := existsReq.Do(ctx, client)
	if err != nil {
		return err
	}
	defer existsResp.Body.Close()

	if existsResp.StatusCode != 200 {
		return fmt.Errorf("dataset %s does not exist", name)
	}

	// 构建properties映射
	mappings := map[string]any{
		"properties": map[string]any{},
	}

	// 根据 schemaDefinition 添加字段映射
	properties := mappings["properties"].(map[string]any)
	for _, column := range schemaDefinition {
		fieldType := column.Type
		switch column.Type {
		case "integer":
			fieldType = "long"
		case "unsigned_integer":
			fieldType = "unsigned_long"
		case "float":
			fieldType = "double"
		case "decimal":
			fieldType = "scaled_float"
		case "string":
			fieldType = "keyword"
		case "datetime":
			fieldType = "date"
		case "time":
			fieldType = "keyword"
		case "json":
			fieldType = "object"
		case "vector":
			fieldType = "knn_vector"
		case "point":
			fieldType = "geo_point"
		case "shape":
			fieldType = "geo_shape"
		default:
			// 保持 fieldType 不变
		}
		// 创建字段属性映射
		fieldProps := map[string]any{
			"type": fieldType,
		}
		// 为decimal类型添加scaling_factor参数
		if column.Type == "decimal" {
			fieldProps["scaling_factor"] = 1000000000000000000.0 // 18位小数
		}
		properties[column.Name] = fieldProps
		// 如果有 column.Features, 则添加到 properties[column.Name] 中
		if column.Features != nil {
			for _, feature := range column.Features {
				if feature.Config != nil {
					for k, v := range feature.Config {
						// 处理嵌套的配置键，如 "method.engine"
						keys := strings.Split(k, ".")
						if len(keys) > 1 {
							// 创建嵌套对象结构
							current := properties[column.Name].(map[string]any)
							for i := 0; i < len(keys)-1; i++ {
								if _, ok := current[keys[i]]; !ok {
									current[keys[i]] = map[string]any{}
								}
								current = current[keys[i]].(map[string]any)
							}
							current[keys[len(keys)-1]] = v
						} else {
							// 直接添加顶层配置
							properties[column.Name].(map[string]any)[k] = v
						}
					}
				}
			}
		}
	}

	// 构建 JSON 字符串
	data, err := json.Marshal(mappings)
	if err != nil {
		return err
	}
	bodyReader := strings.NewReader(string(data))
	da.printRequestBody(bodyReader)
	updateReq := opensearchapi.IndicesPutMappingRequest{
		Index: []string{name},
		Body:  bodyReader,
	}
	updateResp, err := updateReq.Do(ctx, client)
	if err != nil {
		return err
	}
	defer updateResp.Body.Close()

	if updateResp.IsError() {
		return fmt.Errorf("failed to update index mapping: %s", updateResp.String())
	}

	return nil
}

// Delete a Dataset.
func (da *datasetAccess) Delete(ctx context.Context, name string) error {
	// 获取 OpenSearch 客户端
	client, err := da.getOpenSearchClient()
	if err != nil {
		return err
	}

	// 检查索引是否存在
	existsReq := opensearchapi.IndicesExistsRequest{
		Index: []string{name},
	}

	existsResp, err := existsReq.Do(ctx, client)
	if err != nil {
		return err
	}
	defer existsResp.Body.Close()

	// 如果索引存在，删除索引
	if existsResp.StatusCode == 200 {
		deleteReq := opensearchapi.IndicesDeleteRequest{
			Index: []string{name},
		}

		deleteResp, err := deleteReq.Do(ctx, client)
		if err != nil {
			return err
		}
		defer deleteResp.Body.Close()

		if deleteResp.IsError() {
			return fmt.Errorf("failed to delete index: %s", deleteResp.String())
		}
	}

	return nil
}

// CheckExist 检查 dataset 是否存在
func (da *datasetAccess) CheckExist(ctx context.Context, name string) (bool, error) {
	// 获取 OpenSearch 客户端
	client, err := da.getOpenSearchClient()
	if err != nil {
		return false, err
	}

	// 检查索引是否存在
	existsReq := opensearchapi.IndicesExistsRequest{
		Index: []string{name},
	}

	existsResp, err := existsReq.Do(ctx, client)
	if err != nil {
		return false, err
	}
	defer existsResp.Body.Close()

	return existsResp.StatusCode == 200, nil
}

// ListDocuments lists dataset documents.
func (da *datasetAccess) ListDocuments(ctx context.Context, name string, params *interfaces.ResourceDataQueryParams, schemaDefinition []*interfaces.Property) ([]map[string]any, int64, error) {
	// 获取 OpenSearch 客户端
	client, err := da.getOpenSearchClient()
	if err != nil {
		return nil, 0, err
	}

	// 构建查询条件
	query := map[string]any{
		"query": map[string]any{
			"match_all": map[string]any{},
		},
		"from": 0,
		"size": 100,
	}

	// 处理输出字段（_source）
	if params != nil && len(params.OutputFields) > 0 {
		// 过滤掉_score字段，因为它不是源字段而是查询计算的分数
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
		// 确保track_scores为true，以便在需要时获取_score
		if includeScore {
			query["track_scores"] = true
		}
	}

	// 处理排序
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

	// 处理分页
	if params != nil {
		if params.Offset > 0 && params.SearchAfter == nil {
			query["from"] = params.Offset
		}

		if params.Limit > 0 {
			query["size"] = params.Limit
		}

		// 处理 search_after
		if len(params.SearchAfter) > 0 {
			query["search_after"] = params.SearchAfter
		}
	}

	// 处理过滤器
	if params != nil && params.ActualFilterCond != nil {
		// 构建过滤条件查询
		filterQuery, err := da.ConvertFilterCondition(params.ActualFilterCond, schemaDefinition)
		if err != nil {
			return nil, 0, err
		}
		if filterQuery != nil {
			query["query"] = filterQuery
		}
	}

	// 序列化查询条件
	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, 0, err
	}

	// 打印请求body
	bodyReader := bytes.NewReader(queryJSON)
	da.printRequestBody(bodyReader)

	// 列出文档
	req := opensearchapi.SearchRequest{
		Index: []string{name},
		Body:  bodyReader,
	}

	resp, err := req.Do(ctx, client)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return nil, 0, fmt.Errorf("failed to search documents: %s", resp.String())
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, err
	}

	hits, ok := result["hits"].(map[string]any)
	if !ok {
		return nil, 0, fmt.Errorf("invalid search result format")
	}

	total, ok := hits["total"].(map[string]any)["value"].(float64)
	if !ok {
		total = 0
	}

	hitsArray, ok := hits["hits"].([]any)
	if !ok {
		return []map[string]any{}, int64(total), nil
	}

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
		// 添加_score字段（如果存在）
		if score, ok := hitMap["_score"].(float64); ok {
			source["_score"] = score
		}
		documents = append(documents, source)
	}

	return documents, int64(total), nil
}

// CreateDocuments 批量创建 dataset 文档
func (da *datasetAccess) CreateDocuments(ctx context.Context, name string, documents []map[string]any) ([]string, error) {
	// 获取 OpenSearch 客户端
	client, err := da.getOpenSearchClient()
	if err != nil {
		return nil, err
	}

	// 构建批量请求
	var bulkBody strings.Builder
	for _, doc := range documents {
		// 写入操作元数据
		opMeta := map[string]map[string]string{
			"index": {
				"_index": name,
			},
		}
		// 如果文档中包含 _id 字段，则使用该字段作为文档ID
		if docID, ok := doc["_id"].(string); ok {
			opMeta["index"]["_id"] = docID
			delete(doc, "_id") // 从文档数据中移除 _id 字段，避免重复索引
		}

		if err := json.NewEncoder(&bulkBody).Encode(opMeta); err != nil {
			return nil, err
		}
		// 写入文档数据
		if err := json.NewEncoder(&bulkBody).Encode(doc); err != nil {
			return nil, err
		}
	}

	// 打印请求body
	bodyReader := strings.NewReader(bulkBody.String())
	da.printRequestBody(bodyReader)

	// 执行批量请求
	req := opensearchapi.BulkRequest{
		Body:    bodyReader,
		Refresh: "true",
	}

	resp, err := req.Do(ctx, client)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return nil, fmt.Errorf("failed to create documents: %s", resp.String())
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// 提取文档ID
	var docIDs []string
	if items, ok := result["items"].([]interface{}); ok {
		for _, item := range items {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if indexResult, ok := itemMap["index"].(map[string]interface{}); ok {
					if docID, ok := indexResult["_id"].(string); ok {
						docIDs = append(docIDs, docID)
					}
				}
			}
		}
	}

	return docIDs, nil
}

// GetDocument 获取 dataset 文档
func (da *datasetAccess) GetDocument(ctx context.Context, name string, docID string) (map[string]any, error) {
	// 获取 OpenSearch 客户端
	client, err := da.getOpenSearchClient()
	if err != nil {
		return nil, err
	}

	// 获取文档
	req := opensearchapi.GetRequest{
		Index:      name,
		DocumentID: docID,
	}

	resp, err := req.Do(ctx, client)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return nil, fmt.Errorf("failed to get document: %s", resp.String())
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	source, ok := result["_source"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("document not found")
	}

	source["_id"] = result["_id"]

	return source, nil
}

// UpdateDocument 更新 dataset 文档
func (da *datasetAccess) UpdateDocument(ctx context.Context, name string, docID string, document map[string]any) error {
	// 获取 OpenSearch 客户端
	client, err := da.getOpenSearchClient()
	if err != nil {
		return err
	}

	// 更新文档
	data, err := json.Marshal(map[string]any{"doc": document})
	if err != nil {
		return err
	}

	bodyReader := strings.NewReader(string(data))
	da.printRequestBody(bodyReader)

	req := opensearchapi.UpdateRequest{
		Index:      name,
		DocumentID: docID,
		Body:       bodyReader,
	}

	resp, err := req.Do(ctx, client)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return fmt.Errorf("failed to update document: %s", resp.String())
	}

	return nil
}

// DeleteDocument 删除 dataset 文档
func (da *datasetAccess) DeleteDocument(ctx context.Context, name string, docID string) error {
	// 获取 OpenSearch 客户端
	client, err := da.getOpenSearchClient()
	if err != nil {
		return err
	}

	// 删除文档
	req := opensearchapi.DeleteRequest{
		Index:      name,
		DocumentID: docID,
	}

	resp, err := req.Do(ctx, client)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return fmt.Errorf("failed to delete document: %s", resp.String())
	}

	return nil
}

// UpdateDocuments 批量更新 dataset 文档
func (da *datasetAccess) UpdateDocuments(ctx context.Context, name string, updateRequests []map[string]any) error {
	// 获取 OpenSearch 客户端
	client, err := da.getOpenSearchClient()
	if err != nil {
		return err
	}

	// 构建批量更新请求
	var bulkBody bytes.Buffer
	for _, updateReq := range updateRequests {
		docID, ok := updateReq["id"].(string)
		if !ok {
			continue
		}
		document := updateReq["document"]
		if document == nil {
			continue
		}

		// 写入更新操作的元数据
		metadata := map[string]map[string]string{
			"update": {
				"_index": name,
				"_id":    docID,
			},
		}
		if err := json.NewEncoder(&bulkBody).Encode(metadata); err != nil {
			return err
		}

		// 写入更新操作的文档
		updateDoc := map[string]any{
			"doc": document,
		}
		if err := json.NewEncoder(&bulkBody).Encode(updateDoc); err != nil {
			return err
		}
	}

	// 打印请求body
	da.printRequestBody(&bulkBody)

	// 执行批量更新请求
	req := opensearchapi.BulkRequest{
		Body: &bulkBody,
	}

	resp, err := req.Do(ctx, client)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return fmt.Errorf("failed to update documents: %s", resp.String())
	}

	return nil
}

// DeleteDocuments 批量删除 dataset 文档
func (da *datasetAccess) DeleteDocuments(ctx context.Context, name string, docIDs string) error {
	// 获取 OpenSearch 客户端
	client, err := da.getOpenSearchClient()
	if err != nil {
		return err
	}

	// 解析文档 ID 列表（逗号分隔）
	docIDList := strings.Split(docIDs, ",")

	// 构建批量删除请求
	var bulkBody bytes.Buffer
	for _, docID := range docIDList {
		docID = strings.TrimSpace(docID)
		if docID == "" {
			continue
		}

		// 写入删除操作的元数据
		metadata := map[string]map[string]string{
			"delete": {
				"_index": name,
				"_id":    docID,
			},
		}
		if err := json.NewEncoder(&bulkBody).Encode(metadata); err != nil {
			return err
		}
	}

	// 打印请求body
	da.printRequestBody(&bulkBody)

	// 执行批量删除请求
	req := opensearchapi.BulkRequest{
		Body: &bulkBody,
	}

	resp, err := req.Do(ctx, client)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return fmt.Errorf("failed to delete documents: %s", resp.String())
	}

	return nil
}

// DeleteDocumentsByQuery 批量删除 dataset 文档
func (da *datasetAccess) DeleteDocumentsByQuery(ctx context.Context, name string, params *interfaces.ResourceDataQueryParams, schemaDefinition []*interfaces.Property) error {
	// 获取 OpenSearch 客户端
	client, err := da.getOpenSearchClient()
	if err != nil {
		return err
	}

	// 构建查询条件
	query := map[string]any{
		"query": map[string]any{
			"match_all": map[string]any{},
		},
	}

	// 处理过滤条件
	if params != nil && params.ActualFilterCond != nil {
		// 构建过滤条件查询
		filterQuery, err := da.ConvertFilterCondition(params.ActualFilterCond, schemaDefinition)
		if err != nil {
			return err
		}
		if filterQuery != nil {
			query["query"] = filterQuery
		}
	}

	// 将查询条件转换为JSON
	queryBytes, err := json.Marshal(query)
	if err != nil {
		return err
	}

	// 打印请求body
	bodyReader := bytes.NewReader(queryBytes)
	da.printRequestBody(bodyReader)

	// 设置Refresh参数为true
	refresh := true
	req := opensearchapi.DeleteByQueryRequest{
		Index:   []string{name},
		Body:    bodyReader,
		Refresh: &refresh, // 立即刷新索引，确保后续查询能看到删除结果
	}

	resp, err := req.Do(ctx, client)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return fmt.Errorf("failed to delete documents: %s", resp.String())
	}

	return nil
}

// ConvertFilterCondition converts a FilterCondition to OpenSearch DSL.
func (da *datasetAccess) ConvertFilterCondition(condition interfaces.FilterCondition, schemaDefinition []*interfaces.Property) (map[string]any, error) {

	switch condition.GetOperation() {
	case filter_condition.OperationAnd:
		return da.ConvertFilterConditionAnd(condition, schemaDefinition)

	case filter_condition.OperationOr:
		return da.ConvertFilterConditionOr(condition, schemaDefinition)

	default:
		return da.ConvertFilterConditionWithOpr(condition, schemaDefinition)
	}
}

// ConvertFilterConditionAnd converts an AndCond to OpenSearch DSL.
func (da *datasetAccess) ConvertFilterConditionAnd(condition interfaces.FilterCondition, schemaDefinition []*interfaces.Property) (map[string]any, error) {

	condAnd, ok := condition.(*filter_condition.AndCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.AndCond")
	}

	must := make([]map[string]any, 0, len(condAnd.SubConds))
	for _, subCond := range condAnd.SubConds {
		convertedCond, err := da.ConvertFilterCondition(subCond, schemaDefinition)
		if err != nil {
			return nil, err
		}
		must = append(must, convertedCond)
	}

	return map[string]any{
		"bool": map[string]any{
			"must": must,
		},
	}, nil
}

// ConvertFilterConditionOr converts an OrCond to OpenSearch DSL.
func (da *datasetAccess) ConvertFilterConditionOr(condition interfaces.FilterCondition, schemaDefinition []*interfaces.Property) (map[string]any, error) {

	condOr, ok := condition.(*filter_condition.OrCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.OrCond")
	}

	should := make([]map[string]any, 0, len(condOr.SubConds))
	for _, subCond := range condOr.SubConds {
		convertedCond, err := da.ConvertFilterCondition(subCond, schemaDefinition)
		if err != nil {
			return nil, err
		}
		should = append(should, convertedCond)
	}

	return map[string]any{
		"bool": map[string]any{
			"should":               should,
			"minimum_should_match": 1,
		},
	}, nil
}

// ConvertFilterConditionWithOpr converts a FilterCondition with operation to OpenSearch DSL.
func (da *datasetAccess) ConvertFilterConditionWithOpr(condition interfaces.FilterCondition, schemaDefinition []*interfaces.Property) (map[string]any, error) {

	switch condition.GetOperation() {
	case filter_condition.OperationEqual, filter_condition.OperationEqual2:
		return da.ConvertFilterConditionEqual(condition, schemaDefinition)
	case filter_condition.OperationNotEqual, filter_condition.OperationNotEqual2:
		return da.ConvertFilterConditionNotEqual(condition, schemaDefinition)
	case filter_condition.OperationGt, filter_condition.OperationGt2:
		return da.ConvertFilterConditionGt(condition)
	case filter_condition.OperationGte, filter_condition.OperationGte2:
		return da.ConvertFilterConditionGte(condition)
	case filter_condition.OperationLt, filter_condition.OperationLt2:
		return da.ConvertFilterConditionLt(condition)
	case filter_condition.OperationLte, filter_condition.OperationLte2:
		return da.ConvertFilterConditionLte(condition)
	case filter_condition.OperationIn:
		return da.ConvertFilterConditionIn(condition, schemaDefinition)
	case filter_condition.OperationNotIn:
		return da.ConvertFilterConditionNotIn(condition, schemaDefinition)
	case filter_condition.OperationLike:
		return da.ConvertFilterConditionLike(condition, schemaDefinition)
	case filter_condition.OperationNotLike:
		return da.ConvertFilterConditionNotLike(condition, schemaDefinition)
	case filter_condition.OperationContain:
		return da.ConvertFilterConditionContain(condition)
	case filter_condition.OperationNotContain:
		return da.ConvertFilterConditionNotContain(condition)
	case filter_condition.OperationRange:
		return da.ConvertFilterConditionRange(condition)
	case filter_condition.OperationOutRange:
		return da.ConvertFilterConditionOutRange(condition)
	case filter_condition.OperationNull:
		return da.ConvertFilterConditionNull(condition)
	case filter_condition.OperationNotNull:
		return da.ConvertFilterConditionNotNull(condition)
	case filter_condition.OperationEmpty:
		return da.ConvertFilterConditionEmpty(condition)
	case filter_condition.OperationNotEmpty:
		return da.ConvertFilterConditionNotEmpty(condition)
	case filter_condition.OperationPrefix:
		return da.ConvertFilterConditionPrefix(condition)
	case filter_condition.OperationNotPrefix:
		return da.ConvertFilterConditionNotPrefix(condition)
	case filter_condition.OperationBetween:
		return da.ConvertFilterConditionBetween(condition)
	case filter_condition.OperationExist:
		return da.ConvertFilterConditionExist(condition)
	case filter_condition.OperationNotExist:
		return da.ConvertFilterConditionNotExist(condition)
	case filter_condition.OperationRegex:
		return da.ConvertFilterConditionRegex(condition)
	case filter_condition.OperationMatch:
		return da.ConvertFilterConditionMatch(condition)
	case filter_condition.OperationMatchPhrase:
		return da.ConvertFilterConditionMatchPhrase(condition)
	case filter_condition.OperationTrue:
		return da.ConvertFilterConditionTrue(condition)
	case filter_condition.OperationFalse:
		return da.ConvertFilterConditionFalse(condition)
	case filter_condition.OperationBefore:
		return da.ConvertFilterConditionBefore(condition)
	case filter_condition.OperationCurrent:
		return da.ConvertFilterConditionCurrent(condition)
	case filter_condition.OperationMultiMatch:
		return da.ConvertFilterConditionMultiMatch(condition)
	case filter_condition.OperationKnnVector:
		return da.ConvertFilterConditionKnnVector(condition)
	default:
		return nil, fmt.Errorf("operation %s is not supported", condition.GetOperation())
	}
}

// ConvertFilterConditionMultiMatch converts a MultiMatchCond to OpenSearch DSL.
func (da *datasetAccess) ConvertFilterConditionMultiMatch(condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.MultiMatchCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.MultiMatchCond")
	}

	value := cond.Cfg.ValueOptCfg.Value
	fields := make([]string, 0, len(cond.Fields))
	for _, field := range cond.Fields {
		fields = append(fields, field.Name)
	}

	multiMatchQuery := map[string]any{
		"query":  value,
		"fields": fields,
	}

	if cond.MatchType != "" {
		multiMatchQuery["type"] = cond.MatchType
	}

	return map[string]any{
		"multi_match": multiMatchQuery,
	}, nil
}

// ConvertFilterConditionEqual converts an EqualCond to OpenSearch DSL.
func (da *datasetAccess) ConvertFilterConditionEqual(condition interfaces.FilterCondition, schemaDefinition []*interfaces.Property) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.EqualCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.EqualCond")
	}

	keyword := ""
	fieldName := cond.Lfield.OriginalName
	if fieldName == "" {
		fieldName = cond.Lfield.Name
	}
	// 检查字段类型，如果是 text 类型，使用 .keyword 子字段来比较
	for _, prop := range schemaDefinition {
		if prop.OriginalName == fieldName && prop.Type == "text" {
			keyword = ".keyword"
			break
		}
	}
	switch cond.Cfg.ValueFrom {
	case interfaces.ValueFrom_Const:
		return map[string]any{
			"term": map[string]any{
				fieldName + keyword: cond.Value,
			},
		}, nil
	case interfaces.ValueFrom_Field:
		return map[string]any{
			"script": map[string]any{
				"source": fmt.Sprintf("doc['%s'].value == doc['%s'].value", fieldName+keyword, cond.Rfield.OriginalName+keyword),
			},
		}, nil
	default:
		return nil, fmt.Errorf("value_from %s is not supported", cond.Cfg.ValueFrom)
	}
}

// ConvertFilterConditionNotEqual converts a NotEqualCond to OpenSearch DSL.
func (da *datasetAccess) ConvertFilterConditionNotEqual(condition interfaces.FilterCondition, schemaDefinition []*interfaces.Property) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.NotEqualCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NotEqualCond")
	}

	keyword := ""
	fieldName := cond.Lfield.OriginalName
	if fieldName == "" {
		fieldName = cond.Lfield.Name
	}
	// 检查字段类型，如果是 text 类型，使用 .keyword 子字段来比较
	for _, prop := range schemaDefinition {
		if prop.OriginalName == fieldName && prop.Type == "text" {
			keyword = ".keyword"
			break
		}
	}
	switch cond.Cfg.ValueFrom {
	case interfaces.ValueFrom_Const:
		return map[string]any{
			"bool": map[string]any{
				"must_not": map[string]any{
					"term": map[string]any{
						fieldName + keyword: cond.Value,
					},
				},
			},
		}, nil
	case interfaces.ValueFrom_Field:
		return map[string]any{
			"script": map[string]any{
				"source": fmt.Sprintf("doc['%s'].value != doc['%s'].value", fieldName+keyword, cond.Rfield.OriginalName+keyword),
			},
		}, nil
	default:
		return nil, fmt.Errorf("value_from %s is not supported", cond.Cfg.ValueFrom)
	}
}

// ConvertFilterConditionGt converts a GtCond to OpenSearch DSL.
func (da *datasetAccess) ConvertFilterConditionGt(condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.GtCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.GtCond")
	}

	switch cond.Cfg.ValueFrom {
	case interfaces.ValueFrom_Const:
		return map[string]any{
			"range": map[string]any{
				cond.Lfield.OriginalName: map[string]any{
					"gt": cond.Value,
				},
			},
		}, nil
	case interfaces.ValueFrom_Field:
		return map[string]any{
			"script": map[string]any{
				"source": fmt.Sprintf("doc['%s'].value > doc['%s'].value", cond.Lfield.OriginalName, cond.Rfield.OriginalName),
			},
		}, nil
	default:
		return nil, fmt.Errorf("value_from %s is not supported", cond.Cfg.ValueFrom)
	}
}

// ConvertFilterConditionGte converts a GteCond to OpenSearch DSL.
func (da *datasetAccess) ConvertFilterConditionGte(condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.GteCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.GteCond")
	}

	switch cond.Cfg.ValueFrom {
	case interfaces.ValueFrom_Const:
		return map[string]any{
			"range": map[string]any{
				cond.Lfield.OriginalName: map[string]any{
					"gte": cond.Value,
				},
			},
		}, nil
	case interfaces.ValueFrom_Field:
		return map[string]any{
			"script": map[string]any{
				"source": fmt.Sprintf("doc['%s'].value >= doc['%s'].value", cond.Lfield.OriginalName, cond.Rfield.OriginalName),
			},
		}, nil
	default:
		return nil, fmt.Errorf("value_from %s is not supported", cond.Cfg.ValueFrom)
	}
}

// ConvertFilterConditionLt converts a LtCond to OpenSearch DSL.
func (da *datasetAccess) ConvertFilterConditionLt(condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.LtCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.LtCond")
	}

	switch cond.Cfg.ValueFrom {
	case interfaces.ValueFrom_Const:
		return map[string]any{
			"range": map[string]any{
				cond.Lfield.OriginalName: map[string]any{
					"lt": cond.Value,
				},
			},
		}, nil
	case interfaces.ValueFrom_Field:
		return map[string]any{
			"script": map[string]any{
				"source": fmt.Sprintf("doc['%s'].value < doc['%s'].value", cond.Lfield.OriginalName, cond.Rfield.OriginalName),
			},
		}, nil
	default:
		return nil, fmt.Errorf("value_from %s is not supported", cond.Cfg.ValueFrom)
	}
}

// ConvertFilterConditionLte converts a LteCond to OpenSearch DSL.
func (da *datasetAccess) ConvertFilterConditionLte(condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.LteCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.LteCond")
	}

	switch cond.Cfg.ValueFrom {
	case interfaces.ValueFrom_Const:
		return map[string]any{
			"range": map[string]any{
				cond.Lfield.OriginalName: map[string]any{
					"lte": cond.Value,
				},
			},
		}, nil
	case interfaces.ValueFrom_Field:
		return map[string]any{
			"script": map[string]any{
				"source": fmt.Sprintf("doc['%s'].value <= doc['%s'].value", cond.Lfield.OriginalName, cond.Rfield.OriginalName),
			},
		}, nil
	default:
		return nil, fmt.Errorf("value_from %s is not supported", cond.Cfg.ValueFrom)
	}
}

// ConvertFilterConditionIn converts an InCond to OpenSearch DSL.
func (da *datasetAccess) ConvertFilterConditionIn(condition interfaces.FilterCondition, schemaDefinition []*interfaces.Property) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.InCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.InCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [in] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	keyword := ""
	fieldName := cond.Lfield.OriginalName
	if fieldName == "" {
		fieldName = cond.Lfield.Name
	}
	// 检查字段类型，如果是 text 类型，使用 .keyword 子字段来比较
	for _, prop := range schemaDefinition {
		if prop.OriginalName == fieldName && prop.Type == "text" {
			keyword = ".keyword"
			break
		}
	}

	return map[string]any{
		"terms": map[string]any{
			fieldName + keyword: cond.Value,
		},
	}, nil
}

// ConvertFilterConditionNotIn converts a NotInCond to OpenSearch DSL.
func (da *datasetAccess) ConvertFilterConditionNotIn(condition interfaces.FilterCondition, schemaDefinition []*interfaces.Property) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.NotInCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NotInCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [not_in] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	keyword := ""
	fieldName := cond.Lfield.OriginalName
	if fieldName == "" {
		fieldName = cond.Lfield.Name
	}
	// 检查字段类型，如果是 text 类型，使用 .keyword 子字段来比较
	for _, prop := range schemaDefinition {
		if prop.OriginalName == fieldName && prop.Type == "text" {
			keyword = ".keyword"
			break
		}
	}

	return map[string]any{
		"bool": map[string]any{
			"must_not": map[string]any{
				"terms": map[string]any{
					fieldName + keyword: cond.Value,
				},
			},
		},
	}, nil
}

// ConvertFilterConditionLike converts a LikeCond to OpenSearch DSL.
func (da *datasetAccess) ConvertFilterConditionLike(condition interfaces.FilterCondition, schemaDefinition []*interfaces.Property) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.LikeCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.LikeCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [like] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	fieldName := cond.Lfield.OriginalName
	// 检查字段类型，如果是 text 类型，使用 .keyword 后缀
	for _, prop := range schemaDefinition {
		if prop.OriginalName == fieldName && prop.Type == "text" {
			fieldName = fieldName + ".keyword"
			break
		}
	}

	vStr := da.replaceLikeWildcards(cond.Value)
	return map[string]any{
		"regexp": map[string]any{
			fieldName: vStr,
		},
	}, nil
}

// ConvertFilterConditionNotLike converts a NotLikeCond to OpenSearch DSL.
func (da *datasetAccess) ConvertFilterConditionNotLike(condition interfaces.FilterCondition, schemaDefinition []*interfaces.Property) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.NotLikeCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NotLikeCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [not_like] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	fieldName := cond.Lfield.OriginalName
	// 检查字段类型，如果是 text 类型，使用 .keyword 后缀
	for _, prop := range schemaDefinition {
		if prop.OriginalName == fieldName && prop.Type == "text" {
			fieldName = fieldName + ".keyword"
			break
		}
	}

	vStr := da.replaceLikeWildcards(cond.Value)
	return map[string]any{
		"bool": map[string]any{
			"must_not": map[string]any{
				"regexp": map[string]any{
					fieldName: vStr,
				},
			},
		},
	}, nil
}

// ConvertFilterConditionContain converts a ContainCond to OpenSearch DSL.
func (da *datasetAccess) ConvertFilterConditionContain(condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.ContainCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.ContainCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [contain] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	values := cond.Value
	should := make([]map[string]any, len(values))
	for i, v := range values {
		should[i] = map[string]any{
			"term": map[string]any{
				cond.Lfield.OriginalName: v,
			},
		}
	}

	return map[string]any{
		"bool": map[string]any{
			"should":               should,
			"minimum_should_match": 1,
		},
	}, nil
}

// ConvertFilterConditionNotContain converts a NotContainCond to OpenSearch DSL.
func (da *datasetAccess) ConvertFilterConditionNotContain(condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.NotContainCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NotContainCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [not_contain] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	values := cond.Value
	mustNot := make([]map[string]any, len(values))
	for i, v := range values {
		mustNot[i] = map[string]any{
			"term": map[string]any{
				cond.Lfield.OriginalName: v,
			},
		}
	}

	return map[string]any{
		"bool": map[string]any{
			"must_not": mustNot,
		},
	}, nil
}

// ConvertFilterConditionRange converts a RangeCond to OpenSearch DSL.
func (da *datasetAccess) ConvertFilterConditionRange(condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.RangeCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.RangeCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [range] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	values := cond.Value
	if len(values) != 2 {
		return nil, fmt.Errorf("range condition requires exactly 2 values")
	}

	return map[string]any{
		"range": map[string]any{
			cond.Lfield.OriginalName: map[string]any{
				"gte": values[0],
				"lte": values[1],
			},
		},
	}, nil
}

// ConvertFilterConditionOutRange converts an OutRangeCond to OpenSearch DSL.
func (da *datasetAccess) ConvertFilterConditionOutRange(condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.OutRangeCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.OutRangeCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [out_range] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	values := cond.Value
	if len(values) != 2 {
		return nil, fmt.Errorf("out_range condition requires exactly 2 values")
	}

	return map[string]any{
		"bool": map[string]any{
			"should": []map[string]any{
				{
					"range": map[string]any{
						cond.Lfield.OriginalName: map[string]any{
							"lt": values[0],
						},
					},
				},
				{
					"range": map[string]any{
						cond.Lfield.OriginalName: map[string]any{
							"gt": values[1],
						},
					},
				},
			},
			"minimum_should_match": 1,
		},
	}, nil
}

// ConvertFilterConditionNull converts a NullCond to OpenSearch DSL.
func (da *datasetAccess) ConvertFilterConditionNull(condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.NullCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NullCond")
	}

	return map[string]any{
		"bool": map[string]any{
			"must_not": map[string]any{
				"exists": map[string]any{
					"field": cond.Lfield.OriginalName,
				},
			},
		},
	}, nil
}

// ConvertFilterConditionMatch converts a MatchCond to OpenSearch DSL.
func (da *datasetAccess) ConvertFilterConditionMatch(condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.MatchCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.MatchCond")
	}

	value := cond.Cfg.ValueOptCfg.Value

	// 如果是全部字段匹配
	if len(cond.Fields) > 1 {
		should := make([]map[string]any, 0, len(cond.Fields))
		for _, field := range cond.Fields {
			should = append(should, map[string]any{
				"match": map[string]any{
					field.Name: value,
				},
			})
		}
		return map[string]any{
			"bool": map[string]any{
				"should":               should,
				"minimum_should_match": 1,
			},
		}, nil
	} else if len(cond.Fields) == 1 {
		// 单个字段匹配
		field := cond.Fields[0]
		return map[string]any{
			"match": map[string]any{
				field.Name: value,
			},
		}, nil
	}

	return nil, fmt.Errorf("match condition has no fields")
}

// ConvertFilterConditionMatchPhrase converts a MatchPhraseCond to OpenSearch DSL.
func (da *datasetAccess) ConvertFilterConditionMatchPhrase(condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.MatchPhraseCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.MatchPhraseCond")
	}

	value := cond.Cfg.ValueOptCfg.Value

	// 如果是全部字段匹配
	if len(cond.Fields) > 1 {
		should := make([]map[string]any, 0, len(cond.Fields))
		for _, field := range cond.Fields {
			should = append(should, map[string]any{
				"match_phrase": map[string]any{
					field.Name: value,
				},
			})
		}
		return map[string]any{
			"bool": map[string]any{
				"should":               should,
				"minimum_should_match": 1,
			},
		}, nil
	} else if len(cond.Fields) == 1 {
		// 单个字段匹配
		field := cond.Fields[0]
		return map[string]any{
			"match_phrase": map[string]any{
				field.Name: value,
			},
		}, nil
	}

	return nil, fmt.Errorf("match_phrase condition has no fields")
}

// ConvertFilterConditionNotNull converts a NotNullCond to OpenSearch DSL.
func (da *datasetAccess) ConvertFilterConditionNotNull(condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.NotNullCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NotNullCond")
	}

	return map[string]any{
		"exists": map[string]any{
			"field": cond.Lfield.OriginalName,
		},
	}, nil
}

// ConvertFilterConditionEmpty converts an EmptyCond to OpenSearch DSL.
func (da *datasetAccess) ConvertFilterConditionEmpty(condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.EmptyCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.EmptyCond")
	}

	return map[string]any{
		"bool": map[string]any{
			"should": []map[string]any{
				{
					"term": map[string]any{
						cond.Lfield.OriginalName: "",
					},
				},
				{
					"bool": map[string]any{
						"must_not": map[string]any{
							"exists": map[string]any{
								"field": cond.Lfield.OriginalName,
							},
						},
					},
				},
			},
			"minimum_should_match": 1,
		},
	}, nil
}

// ConvertFilterConditionNotEmpty converts a NotEmptyCond to OpenSearch DSL.
func (da *datasetAccess) ConvertFilterConditionNotEmpty(condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.NotEmptyCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NotEmptyCond")
	}

	return map[string]any{
		"bool": map[string]any{
			"must": []map[string]any{
				{
					"exists": map[string]any{
						"field": cond.Lfield.OriginalName,
					},
				},
				{
					"bool": map[string]any{
						"must_not": map[string]any{
							"term": map[string]any{
								cond.Lfield.OriginalName: "",
							},
						},
					},
				},
			},
		},
	}, nil
}

// ConvertFilterConditionPrefix converts a PrefixCond to OpenSearch DSL.
func (da *datasetAccess) ConvertFilterConditionPrefix(condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.PrefixCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.PrefixCond")
	}

	vStr := cond.Value
	return map[string]any{
		"prefix": map[string]any{
			cond.Lfield.OriginalName: vStr,
		},
	}, nil
}

// ConvertFilterConditionNotPrefix converts a NotPrefixCond to OpenSearch DSL.
func (da *datasetAccess) ConvertFilterConditionNotPrefix(condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.NotPrefixCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NotPrefixCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [not_prefix] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	vStr := cond.Value
	return map[string]any{
		"bool": map[string]any{
			"must_not": map[string]any{
				"prefix": map[string]any{
					cond.Lfield.OriginalName: vStr,
				},
			},
		},
	}, nil
}

// ConvertFilterConditionBetween converts a BetweenCond to OpenSearch DSL.
func (da *datasetAccess) ConvertFilterConditionBetween(condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.BetweenCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.BetweenCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [between] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	values := cond.Value
	if len(values) != 2 {
		return nil, fmt.Errorf("between condition requires exactly 2 values")
	}

	return map[string]any{
		"range": map[string]any{
			cond.Lfield.OriginalName: map[string]any{
				"gte": values[0],
				"lte": values[1],
			},
		},
	}, nil
}

// ConvertFilterConditionExist converts an ExistCond to OpenSearch DSL.
func (da *datasetAccess) ConvertFilterConditionExist(condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.ExistCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.ExistCond")
	}

	return map[string]any{
		"exists": map[string]any{
			"field": cond.Lfield.OriginalName,
		},
	}, nil
}

// ConvertFilterConditionNotExist converts a NotExistCond to OpenSearch DSL.
func (da *datasetAccess) ConvertFilterConditionNotExist(condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.NotExistCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NotExistCond")
	}

	return map[string]any{
		"bool": map[string]any{
			"must_not": map[string]any{
				"exists": map[string]any{
					"field": cond.Lfield.OriginalName,
				},
			},
		},
	}, nil
}

// ConvertFilterConditionRegex converts a RegexCond to OpenSearch DSL.
func (da *datasetAccess) ConvertFilterConditionRegex(condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.RegexCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.RegexCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [regex] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	return map[string]any{
		"regexp": map[string]any{
			cond.Lfield.OriginalName: cond.Value,
		},
	}, nil
}

// ConvertFilterConditionTrue converts a TrueCond to OpenSearch DSL.
func (da *datasetAccess) ConvertFilterConditionTrue(condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.TrueCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.TrueCond")
	}

	return map[string]any{
		"term": map[string]any{
			cond.Lfield.OriginalName: true,
		},
	}, nil
}

// ConvertFilterConditionFalse converts a FalseCond to OpenSearch DSL.
func (da *datasetAccess) ConvertFilterConditionFalse(condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.FalseCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.FalseCond")
	}

	return map[string]any{
		"term": map[string]any{
			cond.Lfield.OriginalName: false,
		},
	}, nil
}

// ConvertFilterConditionBefore converts a BeforeCond to OpenSearch DSL.
func (da *datasetAccess) ConvertFilterConditionBefore(condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.BeforeCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.BeforeCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [before] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	values := cond.Value
	if len(values) != 2 {
		return nil, fmt.Errorf("before condition requires exactly 2 values")
	}

	interval, ok := values[0].(float64)
	if !ok {
		return nil, fmt.Errorf("condition [before] interval value should be a number")
	}
	_, ok = values[1].(string)
	if !ok {
		return nil, fmt.Errorf("condition [before] unit value should be a string")
	}

	return map[string]any{
		"range": map[string]any{
			cond.Lfield.OriginalName: map[string]any{
				"lt": map[string]any{
					"now": fmt.Sprintf("-%dh", int(interval)),
				},
			},
		},
	}, nil
}

// ConvertFilterConditionCurrent converts a CurrentCond to OpenSearch DSL.
func (da *datasetAccess) ConvertFilterConditionCurrent(condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.CurrentCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.CurrentCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [current] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	var format string
	switch cond.Value {
	case filter_condition.CurrentYear:
		format = "yyyy"
	case filter_condition.CurrentMonth:
		format = "yyyy-MM"
	case filter_condition.CurrentWeek:
		format = "yyyy-ww"
	case filter_condition.CurrentDay:
		format = "yyyy-MM-dd"
	case filter_condition.CurrentHour:
		format = "yyyy-MM-dd HH"
	case filter_condition.CurrentMinute:
		format = "yyyy-MM-dd HH:mm"
	default:
		return nil, fmt.Errorf("condition [current] unsupported format: %s", cond.Value)
	}

	return map[string]any{
		"script": map[string]any{
			"source": fmt.Sprintf("doc['%s'].value.toString('${format}') == new Date().toString('${format}')", cond.Lfield.OriginalName),
			"params": map[string]any{
				"format": format,
			},
		},
	}, nil
}

// printRequestBody 打印请求body
func (da *datasetAccess) printRequestBody(body interface{}) {
	switch v := body.(type) {
	case *bytes.Buffer:
		logger.Debugf("OpenSearch Request Body: %s", v.String())
	case *strings.Reader:
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(v)
		bodyStr := buf.String()
		logger.Debugf("OpenSearch Request Body: %s", bodyStr)
		// 重置Reader位置
		v.Reset(bodyStr)
	case *bytes.Reader:
		buf := new(bytes.Buffer)
		// 保存原始位置
		pos, _ := v.Seek(0, io.SeekCurrent)
		_, _ = buf.ReadFrom(v)
		bodyStr := buf.String()
		logger.Debugf("OpenSearch Request Body: %s", bodyStr)
		// 重置Reader位置
		_, _ = v.Seek(pos, io.SeekStart)
	default:
		// 其他类型尝试转换为JSON
		if data, err := json.Marshal(body); err == nil {
			logger.Debugf("OpenSearch Request Body: %s", string(data))
		}
	}
}

// getOpenSearchClient 获取或创建 OpenSearch 客户端
func (da *datasetAccess) getOpenSearchClient() (*opensearch.Client, error) {
	if da.osClient != nil {
		return da.osClient, nil
	}

	// 从配置获取 OpenSearch 连接信息
	osConfig := common.GetOpenSearchSetting()

	// 创建 OpenSearch 客户端
	client, err := opensearch.NewClient(opensearch.Config{
		Addresses: []string{fmt.Sprintf("%s://%s:%d", osConfig.Protocol, osConfig.Host, osConfig.Port)},
		Username:  osConfig.Username,
		Password:  osConfig.Password,
	})
	if err != nil {
		return nil, err
	}

	da.osClient = client
	return client, nil
}

// ConvertFilterConditionKnnVector converts a KnnVectorCond to OpenSearch DSL.
func (da *datasetAccess) ConvertFilterConditionKnnVector(condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.KnnVectorCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.KnnVectorCond")
	}

	value := cond.Cfg.ValueOptCfg.Value

	// 构建 knn 查询
	knnQuery := map[string]any{
		cond.FilterFieldName: map[string]any{
			"vector": value,
		},
	}

	// 添加 limit_key 和 limit_value
	if limitKey, ok := cond.Cfg.RemainCfg["limit_key"].(string); ok && limitKey != "" {
		if limitValue, ok := cond.Cfg.RemainCfg["limit_value"]; ok {
			knnQuery[cond.FilterFieldName].(map[string]any)[limitKey] = limitValue
		}
	} else {
		// 使用默认值
		knnQuery[cond.FilterFieldName].(map[string]any)["k"] = 10
	}

	// 添加子条件
	if len(cond.SubConds) > 0 {
		filterQueries := make([]map[string]any, 0, len(cond.SubConds))
		for _, subCond := range cond.SubConds {
			subQuery, err := da.ConvertFilterCondition(subCond, nil)
			if err != nil {
				return nil, err
			}
			filterQueries = append(filterQueries, subQuery)
		}

		return map[string]any{
			"knn": knnQuery,
			"filter": map[string]any{
				"bool": map[string]any{
					"must": filterQueries,
				},
			},
		}, nil
	}

	return map[string]any{
		"knn": knnQuery,
	}, nil
}

// replaceLikeWildcards，把 like 的通配符替换成正则表达式里的字符
func (da *datasetAccess) replaceLikeWildcards(input string) string {
	if input == "" {
		return input
	}

	var result strings.Builder
	escaped := false
	runes := []rune(input)

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		if escaped {
			// 转义字符后的字符
			switch r {
			case '%', '_', '\\':
				result.WriteRune(r)
			default:
				// 如果转义了非特殊字符，保留转义符和字符
				result.WriteRune('\\')
				result.WriteRune(r)
			}
			escaped = false
		} else if r == '\\' {
			// 遇到转义符，检查是否是最后一个字符
			if i == len(runes)-1 {
				// 转义符在末尾，直接输出
				result.WriteRune(r)
			} else {
				// 标记转义状态，但不立即输出转义符
				escaped = true
			}
		} else if r == '%' {
			result.WriteString(".*")
		} else if r == '_' {
			result.WriteString(".")
		} else {
			result.WriteRune(r)
		}
	}

	// 处理以转义符结尾的情况
	if escaped {
		result.WriteRune('\\')
	}

	return result.String()
}
