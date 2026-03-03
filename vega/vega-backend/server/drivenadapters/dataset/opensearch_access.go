// Package dataset provides dataset data access implementations.
package dataset

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"

	"vega-backend/common"
	"vega-backend/interfaces"
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
	mapping := map[string]any{
		"mappings": map[string]any{
			"properties": map[string]any{},
		},
	}
	mapping["settings"] = map[string]any{
		"index": map[string]any{
			"number_of_shards":   1,
			"number_of_replicas": 0,
		},
	}

	// 根据 schemaDefinition 添加字段映射
	properties := mapping["mappings"].(map[string]any)["properties"].(map[string]any)
	for _, column := range schemaDefinition {
		if column.Name == "_id" {
			continue
		}
		// "type": "vector[768]"，表示向量字段，需要特殊处理
		if strings.HasPrefix(column.Type, "vector") {
			properties[column.Name] = map[string]any{
				"type":      "knn_vector",
				"dimension": strings.TrimSuffix(strings.TrimPrefix(column.Type, "vector["), "]"),
				// "method": map[string]any{
				// 	"name": "hnsw",
				// 	"space_type": "l2",
				// 	"engine": "mnslib",
				// },
			}
			continue
		}
		properties[column.Name] = map[string]any{
			"type": column.Type,
		}
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
		createReq := opensearchapi.IndicesCreateRequest{
			Index: name,
			Body:  strings.NewReader(string(data)),
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
		// "type": "vector[768]"，表示向量字段，需要特殊处理
		if strings.HasPrefix(column.Type, "vector") {
			properties[column.Name] = map[string]any{
				"type": "vector",
				"dims": strings.TrimSuffix(strings.TrimPrefix(column.Type, "vector["), "]"),
			}
			continue
		}
		// 映射 'string' 类型到 OpenSearch 的 'text' 类型
		fieldType := column.Type
		if fieldType == "string" {
			fieldType = "text"
		}
		properties[column.Name] = map[string]any{
			"type": fieldType,
		}
	}

	// 构建 JSON 字符串
	data, err := json.Marshal(mappings)
	if err != nil {
		return err
	}
	updateReq := opensearchapi.IndicesPutMappingRequest{
		Index: []string{name},
		Body:  strings.NewReader(string(data)),
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

// ListDocuments 列出 dataset 中的文档
func (da *datasetAccess) ListDocuments(ctx context.Context, name string, params *interfaces.ResourceDataQueryParams) ([]map[string]any, int64, error) {
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
		if params.Offset > 0 {
			query["from"] = params.Offset
		}

		if params.Limit > 0 {
			query["size"] = params.Limit
		}
	}

	// 处理过滤器
	if params != nil && params.FilterCondition != nil {
		// 这里可以根据实际需要处理复杂的过滤条件
		// 暂时先保持简单实现
	}

	// 序列化查询条件
	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, 0, err
	}

	// 列出文档
	req := opensearchapi.SearchRequest{
		Index: []string{name},
		Body:  bytes.NewReader(queryJSON),
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
	// 打印 resp.Body
	println("---------------------------")
	if resultJson, err := json.MarshalIndent(result, "", "  "); err == nil {
		println(string(resultJson))
	} else {
		println("Failed to marshal result")
	}
	println("---------------------------")

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
		if err := json.NewEncoder(&bulkBody).Encode(opMeta); err != nil {
			return nil, err
		}
		// 写入文档数据
		if err := json.NewEncoder(&bulkBody).Encode(doc); err != nil {
			return nil, err
		}
	}

	// 执行批量请求
	req := opensearchapi.BulkRequest{
		Body:    strings.NewReader(bulkBody.String()),
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

	req := opensearchapi.UpdateRequest{
		Index:      name,
		DocumentID: docID,
		Body:       strings.NewReader(string(data)),
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
