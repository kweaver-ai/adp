// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package dataset

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"vega-backend-tests/at/setup"
	"vega-backend-tests/testutil"
)

// TestDatasetResourceCreate Dataset资源创建AT测试
// 测试编号前缀: DS1xx (Dataset Create)
func TestDatasetResourceCreate(t *testing.T) {
	var (
		ctx    context.Context
		config *setup.TestConfig
		client *testutil.HTTPClient
	)

	Convey("Dataset资源创建AT测试 - 初始化", t, func() {
		ctx = context.Background()

		// 加载测试配置
		var err error
		config, err = setup.LoadTestConfig()
		So(err, ShouldBeNil)
		So(config, ShouldNotBeNil)

		// 创建HTTP客户端
		client = testutil.NewHTTPClient(config.VegaBackend.BaseURL)

		// 验证服务可用性
		err = client.CheckHealth()
		So(err, ShouldBeNil)
		t.Logf("✓ AT测试环境就绪，VEGA Manager: %s", config.VegaBackend.BaseURL)

		// 清理现有dataset资源
		cleanupResources(client, t)

		// ========== 正向测试（DS101-DS120） ==========

		Convey("DS101: 创建dataset资源 - 基本场景", func() {
			payload := buildDatasetResourcePayload()
			resp := client.POST("/api/vega-backend/v1/resources", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
			So(resp.Body["id"], ShouldNotBeEmpty)
		})

		Convey("DS102: 创建dataset资源 - 完整字段", func() {
			payload := buildFullDatasetResourcePayload()
			resp := client.POST("/api/vega-backend/v1/resources", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
			So(resp.Body["id"], ShouldNotBeEmpty)
		})

		Convey("DS103: 创建后验证category为dataset", func() {
			payload := buildDatasetResourcePayload()
			createResp := client.POST("/api/vega-backend/v1/resources", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			resourceID := createResp.Body["id"].(string)

			// 查询验证
			getResp := client.GET("/api/vega-backend/v1/resources/" + resourceID)
			So(getResp.StatusCode, ShouldEqual, http.StatusOK)

			resource := extractFromEntriesResponse(getResp)
			So(resource, ShouldNotBeNil)
			So(resource["category"], ShouldEqual, "dataset")
		})

		Convey("DS104: 创建后立即查询", func() {
			payload := buildDatasetResourcePayload()
			createResp := client.POST("/api/vega-backend/v1/resources", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			resourceID := createResp.Body["id"].(string)

			// 立即查询
			getResp := client.GET("/api/vega-backend/v1/resources/" + resourceID)
			So(getResp.StatusCode, ShouldEqual, http.StatusOK)
			resource := extractFromEntriesResponse(getResp)
			So(resource["id"], ShouldEqual, resourceID)
			So(resource["name"], ShouldEqual, payload["name"])
		})

		// ========== 负向测试（DS121-DS127） ==========

		Convey("DS121: 重复的resource名称", func() {
			fixedName := generateUniqueName("duplicate-dataset")
			payload1 := buildDatasetResourcePayloadWithName(fixedName)

			// 第一次创建
			resp1 := client.POST("/api/vega-backend/v1/resources", payload1)
			So(resp1.StatusCode, ShouldEqual, http.StatusCreated)

			// 第二次创建相同名称
			payload2 := buildDatasetResourcePayloadWithName(fixedName)
			resp2 := client.POST("/api/vega-backend/v1/resources", payload2)
			So(resp2.StatusCode, ShouldEqual, http.StatusConflict)
		})

		Convey("DS122: 缺少必填字段 - name", func() {
			payload := map[string]any{
				"category":       "dataset",
				"connector_type": "mariadb",
				"config": map[string]any{
					"host":     "localhost",
					"port":     3306,
					"username": "root",
					"password": "Password123",
					"database": "test",
				},
			}
			resp := client.POST("/api/vega-backend/v1/resources", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})
	})

	_ = ctx
}

// ========== 辅助函数 ==========

// 初始化随机数生成器
func init() {
	rand.Seed(time.Now().UnixNano())
}

// generateUniqueName 生成唯一名称
func generateUniqueName(prefix string) string {
	suffix := rand.Intn(10000)
	return fmt.Sprintf("%s-%d-%d", prefix, time.Now().Unix(), suffix)
}

// cleanupResources 清理现有资源
func cleanupResources(client *testutil.HTTPClient, t *testing.T) {
	resp := client.GET("/api/vega-backend/v1/resources?category=dataset&offset=0&limit=100")
	if resp.StatusCode == http.StatusOK {
		if entries, ok := resp.Body["entries"].([]any); ok {
			for _, entry := range entries {
				if entryMap, ok := entry.(map[string]any); ok {
					if id, ok := entryMap["id"].(string); ok {
						deleteResp := client.DELETE("/api/vega-backend/v1/resources/" + id)
						if deleteResp.StatusCode != http.StatusNoContent {
							t.Logf("清理资源失败 %s: %d", id, deleteResp.StatusCode)
						}
					}
				}
			}
		}
	}
}

// buildDatasetResourcePayload 构建基本的dataset资源payload
func buildDatasetResourcePayload() map[string]any {
	return map[string]any{
		"name":           generateUniqueName("test-dataset"),
		"category":       "dataset",
		"connector_type": "mariadb",
		"config": map[string]any{
			"host":     "localhost",
			"port":     3306,
			"username": "root",
			"password": "Password123",
			"database": "test",
		},
		"source_identifier": "at_db",
		"schema_definition": []map[string]any{
			{"name": "_id", "type": "keyword"},
			{"name": "@timestamp", "type": "long"},
			{"name": "name", "type": "text"},
			{"name": "age", "type": "integer"},
			{"name": "content", "type": "vector[768]"},
		},
	}
}

// buildFullDatasetResourcePayload 构建完整的dataset资源payload
func buildFullDatasetResourcePayload() map[string]any {
	return map[string]any{
		"name":           generateUniqueName("test-dataset-full"),
		"category":       "dataset",
		"connector_type": "mariadb",
		"description":    "测试数据集资源",
		"tags":           []string{"test", "dataset"},
		"config": map[string]any{
			"host":     "localhost",
			"port":     3306,
			"username": "root",
			"password": "Password123",
			"database": "test",
		},
		"source_identifier": "at_db",
		"schema_definition": []map[string]any{
			{"name": "_id", "type": "keyword"},
			{"name": "@timestamp", "type": "long"},
			{"name": "name", "type": "text"},
			{"name": "age", "type": "integer"},
			{"name": "content", "type": "vector[768]"},
		},
	}
}

// buildDatasetResourcePayloadWithName 构建指定名称的dataset资源payload
func buildDatasetResourcePayloadWithName(name string) map[string]any {
	payload := buildDatasetResourcePayload()
	payload["name"] = name
	return payload
}

// extractFromEntriesResponse 从响应中提取资源数据
func extractFromEntriesResponse(resp testutil.HTTPResponse) map[string]any {
	if resp.Body != nil {
		if entries, ok := resp.Body["entries"].([]any); ok {
			if len(entries) > 0 {
				if entry, ok := entries[0].(map[string]any); ok {
					return entry
				}
			}
		}
	}
	return nil
}

// buildUpdatePayload 构建更新payload
func buildUpdatePayload(originalData map[string]any, updates map[string]any) map[string]any {
	// 基于原始数据创建更新payload
	payload := make(map[string]any)
	for k, v := range originalData {
		payload[k] = v
	}

	// 应用更新
	for k, v := range updates {
		payload[k] = v
	}

	return payload
}

// buildDatasetDocumentPayload 构建dataset文档payload
func buildDatasetDocumentPayload() map[string]any {
	return map[string]any{
		"@timestamp": int(time.Now().UnixMilli()),
		"name":       "Test User",
		"age":        30,
		"content":    generateVector(768),
	}
}

// generateVector 生成指定维度的向量
func generateVector(dims int) []float64 {
	vector := make([]float64, dims)
	for i := range vector {
		vector[i] = rand.Float64()*2 - 1 // 生成 [-1, 1] 范围内的随机数
	}
	return vector
}

// TestDatasetResourceRead Dataset资源读取AT测试
// 测试编号前缀: DS2xx
func TestDatasetResourceRead(t *testing.T) {
	var (
		ctx    context.Context
		config *setup.TestConfig
		client *testutil.HTTPClient
	)

	Convey("Dataset资源读取AT测试 - 初始化", t, func() {
		ctx = context.Background()

		var err error
		config, err = setup.LoadTestConfig()
		So(err, ShouldBeNil)

		client = testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		cleanupResources(client, t)

		// ========== 读取测试（DS201-DS210） ==========

		Convey("DS201: 获取存在的dataset资源", func() {
			// 先创建
			payload := buildDatasetResourcePayload()
			createResp := client.POST("/api/vega-backend/v1/resources", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
			resourceID := createResp.Body["id"].(string)

			// 查询
			getResp := client.GET("/api/vega-backend/v1/resources/" + resourceID)
			So(getResp.StatusCode, ShouldEqual, http.StatusOK)

			resource := extractFromEntriesResponse(getResp)
			So(resource, ShouldNotBeNil)
			So(resource["id"], ShouldEqual, resourceID)
		})

		Convey("DS202: 获取不存在的resource", func() {
			resp := client.GET("/api/vega-backend/v1/resources/non-existent-id-12345")
			So(resp.StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("DS203: 列表查询 - 按category过滤dataset", func() {
			// 创建1个dataset resource
			payload := buildDatasetResourcePayload()
			createResp := client.POST("/api/vega-backend/v1/resources", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			// 查询dataset类型
			datasetResp := client.GET("/api/vega-backend/v1/resources?category=dataset&offset=0&limit=10")
			So(datasetResp.StatusCode, ShouldEqual, http.StatusOK)

			if datasetResp.Body != nil && datasetResp.Body["entries"] != nil {
				entries := datasetResp.Body["entries"].([]any)
				So(len(entries), ShouldBeGreaterThanOrEqualTo, 1)
			}
		})
	})

	_ = ctx
}

// TestDatasetResourceUpdate Dataset资源更新AT测试
// 测试编号前缀: DS3xx
func TestDatasetResourceUpdate(t *testing.T) {
	var (
		ctx    context.Context
		config *setup.TestConfig
		client *testutil.HTTPClient
	)

	Convey("Dataset资源更新AT测试 - 初始化", t, func() {
		ctx = context.Background()

		var err error
		config, err = setup.LoadTestConfig()
		So(err, ShouldBeNil)

		client = testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		cleanupResources(client, t)

		// ========== 更新测试（DS301-DS310） ==========

		Convey("DS301: 更新dataset资源名称", func() {
			// 创建
			payload := buildDatasetResourcePayload()
			createResp := client.POST("/api/vega-backend/v1/resources", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
			resourceID := createResp.Body["id"].(string)

			// 获取原始数据
			getResp := client.GET("/api/vega-backend/v1/resources/" + resourceID)
			resourceData := extractFromEntriesResponse(getResp)

			// 基于原数据构建更新payload
			newName := generateUniqueName("updated-dataset")
			updatePayload := buildUpdatePayload(resourceData, map[string]any{
				"name": newName,
			})
			updateResp := client.PUT("/api/vega-backend/v1/resources/"+resourceID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)

			// 验证
			verifyResp := client.GET("/api/vega-backend/v1/resources/" + resourceID)
			resource := extractFromEntriesResponse(verifyResp)
			So(resource["name"], ShouldEqual, newName)
		})

		Convey("DS302: 更新dataset资源schema", func() {
			// 创建
			payload := buildDatasetResourcePayload()
			createResp := client.POST("/api/vega-backend/v1/resources", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
			resourceID := createResp.Body["id"].(string)

			// 获取原始数据
			getResp := client.GET("/api/vega-backend/v1/resources/" + resourceID)
			resourceData := extractFromEntriesResponse(getResp)

			// 更新schema
			newSchema := []map[string]any{
				{"name": "address", "type": "text"},
			}
			updatePayload := buildUpdatePayload(resourceData, map[string]any{
				"schema_definition": newSchema,
			})
			updateResp := client.PUT("/api/vega-backend/v1/resources/"+resourceID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)
		})
	})

	_ = ctx
}

// TestDatasetResourceDelete Dataset资源删除AT测试
// 测试编号前缀: DS4xx
func TestDatasetResourceDelete(t *testing.T) {
	var (
		ctx    context.Context
		config *setup.TestConfig
		client *testutil.HTTPClient
	)

	Convey("Dataset资源删除AT测试 - 初始化", t, func() {
		ctx = context.Background()

		var err error
		config, err = setup.LoadTestConfig()
		So(err, ShouldBeNil)

		client = testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		cleanupResources(client, t)

		// ========== 删除测试（DS401-DS410） ==========

		Convey("DS401: 删除存在的dataset资源", func() {
			// 创建
			payload := buildDatasetResourcePayload()
			client.SetHeader("Content-Type", "application/json")
			createResp := client.POST("/api/vega-backend/v1/resources", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
			resourceID := createResp.Body["id"].(string)

			// 删除
			deleteResp := client.DELETE("/api/vega-backend/v1/resources/" + resourceID)
			So(deleteResp.StatusCode, ShouldEqual, http.StatusNoContent)

			// 验证已删除
			getResp := client.GET("/api/vega-backend/v1/resources/" + resourceID)
			So(getResp.StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("DS402: 删除不存在的resource", func() {
			resp := client.DELETE("/api/vega-backend/v1/resources/non-existent-id-12345")
			So(resp.StatusCode, ShouldEqual, http.StatusNotFound)
		})
	})

	_ = ctx
}

// TestDatasetDocumentsCreate 测试批量创建dataset文档
func TestDatasetDocumentsCreate(t *testing.T) {
	Convey("DD101: 批量创建dataset文档", t, func() {
		var err error
		config, err := setup.LoadTestConfig()
		So(err, ShouldBeNil)

		client := testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		cleanupResources(client, t)

		// 创建测试用的dataset resource
		payload := buildDatasetResourcePayload()
		createResp := client.POST("/api/vega-backend/v1/resources", payload)
		So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
		resourceID := createResp.Body["id"].(string)

		// 构建批量创建文档的payload
		documentsPayload := []map[string]any{
			{
				"@timestamp": time.Now().UnixMilli(),
				"name":       "User 1",
				"age":        25,
				"content":    generateVector(768),
			},
			{
				"@timestamp": time.Now().UnixMilli(),
				"name":       "User 2",
				"age":        35,
				"content":    generateVector(768),
			},
			{
				"@timestamp": time.Now().UnixMilli(),
				"name":       "User 3",
				"age":        40,
				"content":    generateVector(768),
			},
		}
		resp := client.POST("/api/vega-backend/v1/resources/dataset/"+resourceID+"/docs", documentsPayload)
		So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		So(resp.Body["ids"], ShouldNotBeEmpty)
		ids, ok := resp.Body["ids"].([]interface{})
		So(ok, ShouldBeTrue)
		So(len(ids), ShouldEqual, 3)

		// 清理资源
		cleanupResources(client, t)
	})
}

// TestDatasetDocumentsList 测试列出dataset文档
func TestDatasetDocumentsList(t *testing.T) {
	Convey("DD102: 列出dataset文档", t, func() {
		var err error
		config, err := setup.LoadTestConfig()
		So(err, ShouldBeNil)

		client := testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		cleanupResources(client, t)

		// 创建测试用的dataset resource
		payload := buildDatasetResourcePayload()
		createResp := client.POST("/api/vega-backend/v1/resources", payload)
		So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
		resourceID := createResp.Body["id"].(string)

		// 先创建一些文档
		documentsPayload := []map[string]any{
			{
				"@timestamp": time.Now().UnixMilli(),
				"name":       "User 1",
				"age":        25,
				"content":    generateVector(768),
			},
		}
		createDocsResp := client.POST("/api/vega-backend/v1/resources/dataset/"+resourceID+"/docs", documentsPayload)
		So(createDocsResp.StatusCode, ShouldEqual, http.StatusCreated)

		// 构建查询条件
		queryPayload := map[string]any{
			"start": time.Now().UnixMilli() - (24 * 3600 * 1000),
			"end":   time.Now().UnixMilli(),
			"sort": []map[string]any{
				{
					"field":     "@timestamp",
					"direction": "asc",
				},
			},
			"offset":           0,
			"limit":            10,
			"need_total":       true,
			"use_search_after": false,
		}
		// 使用POST请求到/data端点（method override GET）
		client.SetHeader("X-HTTP-Method-Override", "GET")
		resp := client.POST("/api/vega-backend/v1/resources/"+resourceID+"/data", queryPayload)
		client.RemoveHeader("X-HTTP-Method-Override")
		So(resp.StatusCode, ShouldEqual, http.StatusOK)
		So(resp.Body["entries"], ShouldNotBeEmpty)

		// 清理资源
		cleanupResources(client, t)
	})
}

// TestDatasetDocumentGet 测试获取dataset文档
func TestDatasetDocumentGet(t *testing.T) {
	Convey("DD103: 获取dataset文档", t, func() {
		var err error
		config, err := setup.LoadTestConfig()
		So(err, ShouldBeNil)

		client := testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		cleanupResources(client, t)

		// 创建测试用的dataset resource
		payload := buildDatasetResourcePayload()
		createResp := client.POST("/api/vega-backend/v1/resources", payload)
		So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
		resourceID := createResp.Body["id"].(string)

		// 先创建一个文档（使用批量创建接口）
		docPayload := buildDatasetDocumentPayload()
		createDocResp := client.POST("/api/vega-backend/v1/resources/dataset/"+resourceID+"/docs", []map[string]any{docPayload})
		So(createDocResp.StatusCode, ShouldEqual, http.StatusCreated)
		So(createDocResp.Body["ids"], ShouldNotBeEmpty)
		ids, ok := createDocResp.Body["ids"].([]interface{})
		So(ok, ShouldBeTrue)
		So(len(ids), ShouldBeGreaterThan, 0)

		// 获取文档（使用POST /:id/data端点，method override GET）
		queryPayload := map[string]any{
			"start":      time.Now().UnixMilli() - (24 * 3600 * 1000),
			"end":        time.Now().UnixMilli(),
			"offset":     0,
			"limit":      10,
			"need_total": true,
		}
		client.SetHeader("X-HTTP-Method-Override", "GET")
		resp := client.POST("/api/vega-backend/v1/resources/"+resourceID+"/data", queryPayload)
		client.RemoveHeader("X-HTTP-Method-Override")
		So(resp.StatusCode, ShouldEqual, http.StatusOK)
		So(resp.Body["entries"], ShouldNotBeEmpty)

		// 清理资源
		cleanupResources(client, t)
	})
}

// TestDatasetDocumentUpdate 测试更新dataset文档
func TestDatasetDocumentUpdate(t *testing.T) {
	Convey("DD104: 更新dataset文档", t, func() {
		var err error
		config, err := setup.LoadTestConfig()
		So(err, ShouldBeNil)

		client := testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		cleanupResources(client, t)

		// 创建测试用的dataset resource
		payload := buildDatasetResourcePayload()
		createResp := client.POST("/api/vega-backend/v1/resources", payload)
		So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
		resourceID := createResp.Body["id"].(string)

		// 先创建一个文档（使用批量创建接口）
		docPayload := buildDatasetDocumentPayload()
		createDocResp := client.POST("/api/vega-backend/v1/resources/dataset/"+resourceID+"/docs", []map[string]any{docPayload})
		So(createDocResp.StatusCode, ShouldEqual, http.StatusCreated)
		So(createDocResp.Body["ids"], ShouldNotBeEmpty)
		ids, ok := createDocResp.Body["ids"].([]interface{})
		So(ok, ShouldBeTrue)
		So(len(ids), ShouldBeGreaterThan, 0)
		docID := ids[0].(string)

		// 更新文档（使用批量更新接口）
		updatePayload := []map[string]any{
			{
				"id": docID,
				"document": map[string]any{
					"title":   "Updated Test Document",
					"content": generateVector(768),
					"metadata": map[string]any{
						"author":     "Updated Test User",
						"updated_at": "2024-01-02T00:00:00Z",
					},
				},
			},
		}

		resp := client.PUT("/api/vega-backend/v1/resources/dataset/"+resourceID+"/docs", updatePayload)
		So(resp.StatusCode, ShouldEqual, http.StatusNoContent)

		// 清理资源
		cleanupResources(client, t)
	})
}

// TestDatasetDocumentDelete 测试删除dataset文档
func TestDatasetDocumentDelete(t *testing.T) {
	Convey("DD105: 删除dataset文档", t, func() {
		var err error
		config, err := setup.LoadTestConfig()
		So(err, ShouldBeNil)

		client := testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		cleanupResources(client, t)

		// 创建测试用的dataset resource
		payload := buildDatasetResourcePayload()
		createResp := client.POST("/api/vega-backend/v1/resources", payload)
		So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
		resourceID := createResp.Body["id"].(string)

		// 先创建一个文档（使用批量创建接口）
		docPayload := buildDatasetDocumentPayload()
		createDocResp := client.POST("/api/vega-backend/v1/resources/dataset/"+resourceID+"/docs", []map[string]any{docPayload})
		So(createDocResp.StatusCode, ShouldEqual, http.StatusCreated)
		So(createDocResp.Body["ids"], ShouldNotBeEmpty)
		ids, ok := createDocResp.Body["ids"].([]interface{})
		So(ok, ShouldBeTrue)
		So(len(ids), ShouldBeGreaterThan, 0)
		docID := ids[0].(string)

		// 删除文档（使用批量删除接口）
		resp := client.DELETE("/api/vega-backend/v1/resources/dataset/" + resourceID + "/docs/" + docID)
		So(resp.StatusCode, ShouldEqual, http.StatusNoContent)

		// 清理资源
		cleanupResources(client, t)
	})
}
