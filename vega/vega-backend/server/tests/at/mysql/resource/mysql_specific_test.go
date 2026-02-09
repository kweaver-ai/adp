// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package resource

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"vega-backend/tests/at/fixtures"
	catalogfixtures "vega-backend/tests/at/fixtures/catalog"
	resourcefixtures "vega-backend/tests/at/fixtures/resource"
	"vega-backend/tests/at/setup"
	"vega-backend/tests/testutil"
)

// TestMySQLResourceSpecificCreate MySQL Resource特定功能AT测试 - 创建
// 编号规则：MR1xx为MySQL Resource特定创建测试
func TestMySQLResourceSpecificCreate(t *testing.T) {
	var (
		ctx       context.Context
		config    *setup.TestConfig
		client    *testutil.HTTPClient
		builder   *catalogfixtures.MySQLPayloadBuilder
		catalogID string
	)

	Convey("MySQL Resource特定功能AT测试 - 创建 - 初始化", t, func() {
		ctx = context.Background()

		// 加载测试配置
		var err error
		config, err = setup.LoadTestConfig()
		So(err, ShouldBeNil)
		So(config, ShouldNotBeNil)

		// 验证MySQL配置
		So(config.TargetMySQL.Host, ShouldNotBeEmpty)
		So(config.TargetMySQL.Port, ShouldBeGreaterThan, 0)

		// 创建HTTP客户端
		client = testutil.NewHTTPClient(config.VegaManager.BaseURL)

		// 验证服务可用性
		err = client.CheckHealth()
		So(err, ShouldBeNil)
		t.Logf("✓ AT测试环境就绪，VEGA Manager: %s", config.VegaManager.BaseURL)

		// 创建MySQL payload构建器
		builder = catalogfixtures.NewMySQLPayloadBuilder(config.TargetMySQL)

		// 清理现有数据
		fixtures.CleanupResources(client, t)
		fixtures.CleanupCatalogs(client, t)

		// 创建前置Catalog
		catalogPayload := builder.BuildCreatePayload()
		catalogResp := client.POST("/api/vega-backend/v1/catalogs", catalogPayload)
		So(catalogResp.StatusCode, ShouldEqual, http.StatusCreated)
		catalogID = catalogResp.Body["id"].(string)
		t.Logf("✓ 前置Catalog创建成功，ID: %s", catalogID)

		// ========== MySQL Resource正向测试（MR101-MR120） ==========

		Convey("MR101: 创建resource后验证catalog_id关联", func() {
			payload := resourcefixtures.BuildCreatePayload(catalogID)
			createResp := client.POST("/api/vega-backend/v1/resources", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			resourceID := createResp.Body["id"].(string)

			// 查询验证
			getResp := client.GET("/api/vega-backend/v1/resources/" + resourceID)
			So(getResp.StatusCode, ShouldEqual, http.StatusOK)

			resource := fixtures.ExtractFromEntriesResponse(getResp)
			So(resource, ShouldNotBeNil)
			So(resource["catalog_id"], ShouldEqual, catalogID)
		})

		Convey("MR102: 查询resource - 验证所有字段返回", func() {
			// 创建完整字段的resource
			payload := resourcefixtures.BuildFullCreatePayload(catalogID)
			createResp := client.POST("/api/vega-backend/v1/resources", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			resourceID := createResp.Body["id"].(string)

			// 查询resource
			getResp := client.GET("/api/vega-backend/v1/resources/" + resourceID)
			So(getResp.StatusCode, ShouldEqual, http.StatusOK)

			// 从响应中提取resource对象
			resource := fixtures.ExtractFromEntriesResponse(getResp)
			So(resource, ShouldNotBeNil)

			// 验证基本字段
			So(resource["id"], ShouldNotBeEmpty)
			So(resource["name"], ShouldNotBeEmpty)
			So(resource["catalog_id"], ShouldEqual, catalogID)
			So(resource["description"], ShouldEqual, payload["description"])
			So(resource["create_time"], ShouldNotBeZeroValue)
			So(resource["update_time"], ShouldNotBeZeroValue)

			// 验证tags
			tags, ok := resource["tags"].([]any)
			So(ok, ShouldBeTrue)
			So(tags, ShouldNotBeEmpty)

			// 验证creator和updater
			So(resource["creator"], ShouldNotBeNil)
			So(resource["updater"], ShouldNotBeNil)
		})

		Convey("MR103: 按catalog_id过滤resource列表", func() {
			// 创建1个resource
			payload := resourcefixtures.BuildCreatePayload(catalogID)
			createResp := client.POST("/api/vega-backend/v1/resources", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			// 按catalog_id过滤
			filterResp := client.GET("/api/vega-backend/v1/resources?catalog_id=" + catalogID + "&offset=0&limit=100")
			So(filterResp.StatusCode, ShouldEqual, http.StatusOK)

			if filterResp.Body != nil && filterResp.Body["entries"] != nil {
				entries := filterResp.Body["entries"].([]any)
				So(len(entries), ShouldBeGreaterThanOrEqualTo, 1)

				// 验证都属于当前catalog
				for _, entry := range entries {
					resourceEntry := entry.(map[string]any)
					So(resourceEntry["catalog_id"], ShouldEqual, catalogID)
				}
			}
		})

		Convey("MR104: 列表查询 - 分页测试", func() {
			// 创建5个resource
			for i := 0; i < 5; i++ {
				payload := resourcefixtures.BuildCreatePayload(catalogID)
				resp := client.POST("/api/vega-backend/v1/resources", payload)
				So(resp.StatusCode, ShouldEqual, http.StatusCreated)
			}

			// 第一页
			page1Resp := client.GET("/api/vega-backend/v1/resources?offset=0&limit=2")
			So(page1Resp.StatusCode, ShouldEqual, http.StatusOK)
			entries1 := page1Resp.Body["entries"].([]any)
			So(len(entries1), ShouldBeLessThanOrEqualTo, 2)

			// 第二页
			page2Resp := client.GET("/api/vega-backend/v1/resources?offset=2&limit=2")
			So(page2Resp.StatusCode, ShouldEqual, http.StatusOK)
			entries2 := page2Resp.Body["entries"].([]any)
			So(len(entries2), ShouldBeLessThanOrEqualTo, 2)
		})

		Convey("MR105: 列表查询 - 默认分页参数", func() {
			defaultResp := client.GET("/api/vega-backend/v1/resources")
			So(defaultResp.StatusCode, ShouldEqual, http.StatusOK)
			So(defaultResp.Body, ShouldNotBeNil)
			So(defaultResp.Body["entries"], ShouldNotBeNil)
			_, hasTotalCount := defaultResp.Body["total_count"]
			So(hasTotalCount, ShouldBeTrue)
		})
	})

	_ = ctx
	_ = builder
}

// TestMySQLResourceSpecificUpdate MySQL Resource特定功能AT测试 - 更新
// 编号规则：MR3xx为MySQL Resource特定更新测试
func TestMySQLResourceSpecificUpdate(t *testing.T) {
	var (
		ctx       context.Context
		config    *setup.TestConfig
		client    *testutil.HTTPClient
		builder   *catalogfixtures.MySQLPayloadBuilder
		catalogID string
	)

	Convey("MySQL Resource特定功能AT测试 - 更新 - 初始化", t, func() {
		ctx = context.Background()

		// 加载测试配置
		var err error
		config, err = setup.LoadTestConfig()
		So(err, ShouldBeNil)

		// 创建HTTP客户端
		client = testutil.NewHTTPClient(config.VegaManager.BaseURL)

		// 验证服务可用性
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		// 创建MySQL payload构建器
		builder = catalogfixtures.NewMySQLPayloadBuilder(config.TargetMySQL)

		// 清理现有数据
		fixtures.CleanupResources(client, t)
		fixtures.CleanupCatalogs(client, t)

		// 创建前置Catalog
		catalogPayload := builder.BuildCreatePayload()
		catalogResp := client.POST("/api/vega-backend/v1/catalogs", catalogPayload)
		So(catalogResp.StatusCode, ShouldEqual, http.StatusCreated)
		catalogID = catalogResp.Body["id"].(string)

		// ========== MySQL Resource更新测试（MR301-MR320） ==========

		Convey("MR301: 同时更新多个字段", func() {
			// 创建resource
			payload := resourcefixtures.BuildCreatePayload(catalogID)
			createResp := client.POST("/api/vega-backend/v1/resources", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			resourceID := createResp.Body["id"].(string)

			// 获取原始数据
			getResp := client.GET("/api/vega-backend/v1/resources/" + resourceID)
			originalData := fixtures.ExtractFromEntriesResponse(getResp)

			// 基于原数据构建更新payload
			updatePayload := resourcefixtures.BuildUpdatePayload(originalData, map[string]any{
				"name":        fixtures.GenerateUniqueName("multi-update-resource"),
				"description": "同时更新多个字段的测试",
				"tags":        []string{"multi-update", "test"},
			})

			updateResp := client.PUT("/api/vega-backend/v1/resources/"+resourceID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("MR302: 验证update_time更新", func() {
			// 创建resource
			payload := resourcefixtures.BuildCreatePayload(catalogID)
			createResp := client.POST("/api/vega-backend/v1/resources", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			resourceID := createResp.Body["id"].(string)

			// 等待1秒确保时间戳不同
			time.Sleep(1 * time.Second)

			// 获取原始数据
			getResp := client.GET("/api/vega-backend/v1/resources/" + resourceID)
			originalData := fixtures.ExtractFromEntriesResponse(getResp)
			originalUpdateTime := originalData["update_time"].(float64)

			// 基于原数据构建更新payload
			updatePayload := resourcefixtures.BuildUpdatePayload(originalData, map[string]any{
				"description": "验证update_time更新",
			})
			updateResp := client.PUT("/api/vega-backend/v1/resources/"+resourceID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)

			// 验证update_time已更新
			newGetResp := client.GET("/api/vega-backend/v1/resources/" + resourceID)
			newData := fixtures.ExtractFromEntriesResponse(newGetResp)
			newUpdateTime := newData["update_time"].(float64)
			So(newUpdateTime, ShouldBeGreaterThan, originalUpdateTime)
		})

		Convey("MR303: 验证create_time不变", func() {
			// 创建resource
			payload := resourcefixtures.BuildCreatePayload(catalogID)
			createResp := client.POST("/api/vega-backend/v1/resources", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			resourceID := createResp.Body["id"].(string)

			// 获取原始数据
			getResp := client.GET("/api/vega-backend/v1/resources/" + resourceID)
			originalData := fixtures.ExtractFromEntriesResponse(getResp)
			originalCreateTime := originalData["create_time"].(float64)

			// 基于原数据构建更新payload
			updatePayload := resourcefixtures.BuildUpdatePayload(originalData, map[string]any{
				"description": "验证create_time不变",
			})
			updateResp := client.PUT("/api/vega-backend/v1/resources/"+resourceID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)

			// 验证create_time不变
			newGetResp := client.GET("/api/vega-backend/v1/resources/" + resourceID)
			newData := fixtures.ExtractFromEntriesResponse(newGetResp)
			newCreateTime := newData["create_time"].(float64)
			So(newCreateTime, ShouldEqual, originalCreateTime)
		})
	})

	_ = ctx
	_ = builder
}

// TestMySQLResourceSpecificDelete MySQL Resource特定功能AT测试 - 删除
// 编号规则：MR4xx为MySQL Resource特定删除测试
func TestMySQLResourceSpecificDelete(t *testing.T) {
	var (
		ctx       context.Context
		config    *setup.TestConfig
		client    *testutil.HTTPClient
		builder   *catalogfixtures.MySQLPayloadBuilder
		catalogID string
	)

	Convey("MySQL Resource特定功能AT测试 - 删除 - 初始化", t, func() {
		ctx = context.Background()

		// 加载测试配置
		var err error
		config, err = setup.LoadTestConfig()
		So(err, ShouldBeNil)

		// 创建HTTP客户端
		client = testutil.NewHTTPClient(config.VegaManager.BaseURL)

		// 验证服务可用性
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		// 创建MySQL payload构建器
		builder = catalogfixtures.NewMySQLPayloadBuilder(config.TargetMySQL)

		// 清理现有数据
		fixtures.CleanupResources(client, t)
		fixtures.CleanupCatalogs(client, t)

		// 创建前置Catalog
		catalogPayload := builder.BuildCreatePayload()
		catalogResp := client.POST("/api/vega-backend/v1/catalogs", catalogPayload)
		So(catalogResp.StatusCode, ShouldEqual, http.StatusCreated)
		catalogID = catalogResp.Body["id"].(string)

		// ========== MySQL Resource删除测试（MR401-MR420） ==========

		Convey("MR401: 删除后不能更新", func() {
			// 创建resource
			payload := resourcefixtures.BuildCreatePayload(catalogID)
			createResp := client.POST("/api/vega-backend/v1/resources", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			resourceID := createResp.Body["id"].(string)

			// 删除resource
			deleteResp := client.DELETE("/api/vega-backend/v1/resources/" + resourceID)
			So(deleteResp.StatusCode, ShouldEqual, http.StatusNoContent)

			// 尝试更新已删除的resource
			updateResp := client.PUT("/api/vega-backend/v1/resources/"+resourceID, payload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("MR402: 删除后可以创建同名resource", func() {
			// 创建resource
			payload := resourcefixtures.BuildCreatePayload(catalogID)
			resourceName := payload["name"]
			createResp1 := client.POST("/api/vega-backend/v1/resources", payload)
			So(createResp1.StatusCode, ShouldEqual, http.StatusCreated)

			resourceID1 := createResp1.Body["id"].(string)

			// 删除resource
			deleteResp := client.DELETE("/api/vega-backend/v1/resources/" + resourceID1)
			So(deleteResp.StatusCode, ShouldEqual, http.StatusNoContent)

			// 创建同名resource
			payload2 := resourcefixtures.BuildCreatePayload(catalogID)
			payload2["name"] = resourceName
			createResp2 := client.POST("/api/vega-backend/v1/resources", payload2)

			So(createResp2.StatusCode, ShouldEqual, http.StatusCreated)

			// 新创建的resource应该有不同的ID
			resourceID2 := createResp2.Body["id"].(string)
			So(resourceID2, ShouldNotEqual, resourceID1)
		})

		Convey("MR403: 删除后列表中不再显示", func() {
			// 创建resource
			payload := resourcefixtures.BuildCreatePayload(catalogID)
			resourceName := payload["name"]
			createResp := client.POST("/api/vega-backend/v1/resources", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			resourceID := createResp.Body["id"].(string)

			// 删除resource
			deleteResp := client.DELETE("/api/vega-backend/v1/resources/" + resourceID)
			So(deleteResp.StatusCode, ShouldEqual, http.StatusNoContent)

			// 查询列表
			listResp := client.GET("/api/vega-backend/v1/resources?offset=0&limit=1000")
			So(listResp.StatusCode, ShouldEqual, http.StatusOK)

			if listResp.Body != nil && listResp.Body["entries"] != nil {
				entries := listResp.Body["entries"].([]any)

				// 验证删除的resource不在列表中
				found := false
				for _, entry := range entries {
					resourceEntry := entry.(map[string]any)
					if resourceEntry["name"] == resourceName {
						found = true
						break
					}
				}

				So(found, ShouldBeFalse)
			}
		})

		Convey("MR404: 使用无效ID删除", func() {
			invalidIDs := []string{
				"invalid-id-format",
				"../../../etc/passwd",
				"<script>alert('xss')</script>",
			}

			for _, invalidID := range invalidIDs {
				Convey("无效ID: "+invalidID, func() {
					deleteResp := client.DELETE("/api/vega-backend/v1/resources/" + invalidID)
					So(deleteResp.StatusCode, ShouldEqual, http.StatusNotFound)
				})
			}
		})

		Convey("MR405: 批量删除多个resource", func() {
			// 创建3个resource
			resourceIDs := make([]string, 3)
			for i := 0; i < 3; i++ {
				payload := resourcefixtures.BuildCreatePayload(catalogID)
				createResp := client.POST("/api/vega-backend/v1/resources", payload)
				So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
				resourceIDs[i] = createResp.Body["id"].(string)
			}

			// 批量删除
			idsStr := strings.Join(resourceIDs, ",")
			deleteResp := client.DELETE("/api/vega-backend/v1/resources/" + idsStr)
			So(deleteResp.StatusCode, ShouldEqual, http.StatusNoContent)

			// 验证所有resource都已删除
			for _, resourceID := range resourceIDs {
				getResp := client.GET("/api/vega-backend/v1/resources/" + resourceID)
				So(getResp.StatusCode, ShouldEqual, http.StatusNotFound)
			}
		})
	})

	_ = ctx
	_ = builder
}
