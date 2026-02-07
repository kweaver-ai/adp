// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package catalog

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"vega-backend/tests/at/fixtures"
	catalogfixtures "vega-backend/tests/at/fixtures/catalog"
	"vega-backend/tests/at/setup"
	"vega-backend/tests/testutil"
)

// TestOpenSearchSpecificCreate OpenSearch特定功能AT测试 - 创建
// 编号规则：OS1xx为OpenSearch特定创建测试
// - OS101-OS120: 正向测试
// - OS121-OS140: 负向测试
func TestOpenSearchSpecificCreate(t *testing.T) {
	var (
		ctx     context.Context
		config  *setup.TestConfig
		client  *testutil.HTTPClient
		builder *catalogfixtures.OpenSearchPayloadBuilder
	)

	Convey("OpenSearch特定功能AT测试 - 创建 - 初始化", t, func() {
		ctx = context.Background()

		// 加载测试配置
		var err error
		config, err = setup.LoadTestConfig()
		So(err, ShouldBeNil)
		So(config, ShouldNotBeNil)

		// 验证OpenSearch配置
		So(config.TargetOpenSearch.Host, ShouldNotBeEmpty)
		So(config.TargetOpenSearch.Port, ShouldBeGreaterThan, 0)

		// 创建HTTP客户端
		client = testutil.NewHTTPClient(config.VegaManager.BaseURL)

		// 验证服务可用性
		err = client.CheckHealth()
		So(err, ShouldBeNil)
		t.Logf("✓ AT测试环境就绪，VEGA Manager: %s", config.VegaManager.BaseURL)

		// 创建OpenSearch payload构建器
		builder = catalogfixtures.NewOpenSearchPayloadBuilder(config.TargetOpenSearch)

		// 清理现有catalog
		fixtures.CleanupCatalogs(client, t)

		// ========== OpenSearch正向测试（OS101-OS120） ==========

		Convey("OS101: 创建带SSL配置的OpenSearch catalog（SSL禁用）", func() {
			payload := builder.BuildCreatePayloadWithSSL(false)
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
			So(resp.Body["id"], ShouldNotBeEmpty)
		})

		Convey("OS102: 创建带自定义options的OpenSearch catalog", func() {
			options := map[string]any{
				"timeout":            "30s",
				"max_retries":        5,
				"compress":           true,
				"discovery_interval": "10m",
			}
			payload := builder.BuildCreatePayloadWithOptions(options)
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
			So(resp.Body["id"], ShouldNotBeEmpty)
		})

		Convey("OS103: 创建OpenSearch catalog后验证connector_type", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			// 查询验证
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			So(getResp.StatusCode, ShouldEqual, http.StatusOK)

			catalog := fixtures.ExtractFromEntriesResponse(getResp)
			So(catalog, ShouldNotBeNil)
			So(catalog["connector_type"], ShouldEqual, "opensearch")
		})

		Convey("OS104: 创建完整字段的OpenSearch catalog", func() {
			payload := builder.BuildFullCreatePayload()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
			So(resp.Body["id"], ShouldNotBeEmpty)

			// 验证返回的字段
			catalogID := resp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			catalog := fixtures.ExtractFromEntriesResponse(getResp)

			So(catalog["description"], ShouldEqual, "完整的OpenSearch测试catalog")
			tags, ok := catalog["tags"].([]any)
			So(ok, ShouldBeTrue)
			So(len(tags), ShouldBeGreaterThan, 0)
		})

		Convey("OS105: OpenSearch连接测试成功", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			// 测试连接
			testResp := client.POST("/api/vega-backend/v1/catalogs/"+catalogID+"/test-connection", nil)
			So(testResp.StatusCode, ShouldEqual, http.StatusOK)

			if testResp.Body != nil {
				_, hasStatus := testResp.Body["health_check_status"]
				So(hasStatus, ShouldBeTrue)
			}
		})

		Convey("OS106: 获取OpenSearch catalog健康状态", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			// 获取健康状态
			statusResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID + "/health-status")
			So(statusResp.StatusCode, ShouldEqual, http.StatusOK)

			if statusResp.Body != nil {
				_, hasStatus := statusResp.Body["health_check_status"]
				So(hasStatus, ShouldBeTrue)
			}
		})

		// ========== OpenSearch负向测试（OS121-OS140） ==========

		Convey("OS121: 无效端口测试（非数字）", func() {
			payload := builder.BuildCreatePayloadWithInvalidPort()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("OS122: 缺少认证信息测试", func() {
			payload := builder.BuildCreatePayloadWithMissingAuth()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("OS123: 错误凭证测试", func() {
			payload := builder.BuildCreatePayloadWithWrongCredentials()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("OS124: 无效host测试", func() {
			payload := builder.BuildCreatePayloadWithInvalidHost()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("OS125: 超出范围端口测试", func() {
			payload := builder.BuildCreatePayloadWithOutOfRangePort()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("OS126: 缺少host字段测试", func() {
			payload := map[string]any{
				"name":           catalogfixtures.GenerateUniqueName("missing-host-catalog"),
				"connector_type": "opensearch",
				"connector_config": map[string]any{
					// 缺少host
					"port":     config.TargetOpenSearch.Port,
					"username": config.TargetOpenSearch.Username,
					"password": config.TargetOpenSearch.Password,
					"use_ssl":  config.TargetOpenSearch.UseSSL,
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("OS127: 缺少port字段测试", func() {
			payload := map[string]any{
				"name":           catalogfixtures.GenerateUniqueName("missing-port-catalog"),
				"connector_type": "opensearch",
				"connector_config": map[string]any{
					"host": config.TargetOpenSearch.Host,
					// 缺少port
					"username": config.TargetOpenSearch.Username,
					"password": config.TargetOpenSearch.Password,
					"use_ssl":  config.TargetOpenSearch.UseSSL,
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		// ========== OpenSearch边界测试（OS141-OS160） ==========

		Convey("OS141: 端口边界测试 - 最小有效端口（1）", func() {
			payload := map[string]any{
				"name":           catalogfixtures.GenerateUniqueName("min-port-catalog"),
				"connector_type": "opensearch",
				"connector_config": map[string]any{
					"host":     config.TargetOpenSearch.Host,
					"port":     1,
					"username": config.TargetOpenSearch.Username,
					"password": config.TargetOpenSearch.Password,
					"use_ssl":  config.TargetOpenSearch.UseSSL,
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			// 端口1有效但可能无法连接，期望BadRequest（连接失败）
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("OS142: 端口边界测试 - 最大有效端口（65535）", func() {
			payload := map[string]any{
				"name":           catalogfixtures.GenerateUniqueName("max-port-catalog"),
				"connector_type": "opensearch",
				"connector_config": map[string]any{
					"host":     config.TargetOpenSearch.Host,
					"port":     65535,
					"username": config.TargetOpenSearch.Username,
					"password": config.TargetOpenSearch.Password,
					"use_ssl":  config.TargetOpenSearch.UseSSL,
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			// 端口65535有效但可能无法连接，期望BadRequest（连接失败）
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("OS143: 端口边界测试 - 超出最大端口（65536）", func() {
			payload := map[string]any{
				"name":           catalogfixtures.GenerateUniqueName("overflow-port-catalog"),
				"connector_type": "opensearch",
				"connector_config": map[string]any{
					"host":     config.TargetOpenSearch.Host,
					"port":     65536,
					"username": config.TargetOpenSearch.Username,
					"password": config.TargetOpenSearch.Password,
					"use_ssl":  config.TargetOpenSearch.UseSSL,
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("OS144: 端口边界测试 - 负数端口", func() {
			payload := map[string]any{
				"name":           catalogfixtures.GenerateUniqueName("negative-port-catalog"),
				"connector_type": "opensearch",
				"connector_config": map[string]any{
					"host":     config.TargetOpenSearch.Host,
					"port":     -1,
					"username": config.TargetOpenSearch.Username,
					"password": config.TargetOpenSearch.Password,
					"use_ssl":  config.TargetOpenSearch.UseSSL,
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("OS145: 空host测试", func() {
			payload := map[string]any{
				"name":           catalogfixtures.GenerateUniqueName("empty-host-catalog"),
				"connector_type": "opensearch",
				"connector_config": map[string]any{
					"host":     "",
					"port":     config.TargetOpenSearch.Port,
					"username": config.TargetOpenSearch.Username,
					"password": config.TargetOpenSearch.Password,
					"use_ssl":  config.TargetOpenSearch.UseSSL,
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("OS146: 空用户名测试", func() {
			payload := map[string]any{
				"name":           catalogfixtures.GenerateUniqueName("empty-username-catalog"),
				"connector_type": "opensearch",
				"connector_config": map[string]any{
					"host":     config.TargetOpenSearch.Host,
					"port":     config.TargetOpenSearch.Port,
					"username": "",
					"password": config.TargetOpenSearch.Password,
					"use_ssl":  config.TargetOpenSearch.UseSSL,
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			// 空用户名可能被接受（取决于业务规则），但连接会失败
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("OS147: 空密码测试", func() {
			payload := map[string]any{
				"name":           catalogfixtures.GenerateUniqueName("empty-password-catalog"),
				"connector_type": "opensearch",
				"connector_config": map[string]any{
					"host":     config.TargetOpenSearch.Host,
					"port":     config.TargetOpenSearch.Port,
					"username": config.TargetOpenSearch.Username,
					"password": "",
					"use_ssl":  config.TargetOpenSearch.UseSSL,
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			// 空密码可能被接受（取决于业务规则），但连接会失败
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})
	})

	_ = ctx
}

// TestOpenSearchSpecificRead OpenSearch特定功能AT测试 - 读取
// 编号规则：OS2xx为OpenSearch特定读取测试
func TestOpenSearchSpecificRead(t *testing.T) {
	var (
		ctx     context.Context
		config  *setup.TestConfig
		client  *testutil.HTTPClient
		builder *catalogfixtures.OpenSearchPayloadBuilder
	)

	Convey("OpenSearch特定功能AT测试 - 读取 - 初始化", t, func() {
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

		// 创建OpenSearch payload构建器
		builder = catalogfixtures.NewOpenSearchPayloadBuilder(config.TargetOpenSearch)

		// 清理现有catalog
		fixtures.CleanupCatalogs(client, t)

		// ========== OpenSearch读取测试（OS201-OS220） ==========

		Convey("OS201: 查询catalog - 验证所有字段返回", func() {
			// 创建完整字段的catalog
			payload := builder.BuildFullCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			// 查询catalog
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			So(getResp.StatusCode, ShouldEqual, http.StatusOK)

			// 从响应中提取catalog对象
			catalog := fixtures.ExtractFromEntriesResponse(getResp)
			So(catalog, ShouldNotBeNil)

			// 验证基本字段
			So(catalog["id"], ShouldNotBeEmpty)
			So(catalog["name"], ShouldNotBeEmpty)
			So(catalog["connector_type"], ShouldEqual, "opensearch")
			So(catalog["description"], ShouldEqual, payload["description"])
			So(catalog["create_time"], ShouldNotBeZeroValue)
			So(catalog["update_time"], ShouldNotBeZeroValue)

			// 验证tags
			tags := catalog["tags"].([]any)
			So(tags, ShouldNotBeEmpty)

			// 验证creator和updater
			So(catalog["creator"], ShouldNotBeNil)
			So(catalog["updater"], ShouldNotBeNil)
		})

		Convey("OS202: 列表查询 - 按connector_type过滤opensearch", func() {
			// 创建1个opensearch catalog
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			// 查询physical类型（opensearch是physical类型）
			physicalResp := client.GET("/api/vega-backend/v1/catalogs?type=physical&offset=0&limit=100")
			So(physicalResp.StatusCode, ShouldEqual, http.StatusOK)

			if physicalResp.Body != nil && physicalResp.Body["entries"] != nil {
				entries := physicalResp.Body["entries"].([]any)
				So(len(entries), ShouldBeGreaterThanOrEqualTo, 1)

				// 验证都是physical类型
				for _, entry := range entries {
					catalogEntry := entry.(map[string]any)
					So(catalogEntry["type"], ShouldEqual, "physical")
				}
			}
		})

		Convey("OS203: 列表查询 - 分页测试", func() {
			// 创建5个catalog
			for i := 0; i < 5; i++ {
				payload := builder.BuildCreatePayload()
				resp := client.POST("/api/vega-backend/v1/catalogs", payload)
				So(resp.StatusCode, ShouldEqual, http.StatusCreated)
			}

			// 第一页
			page1Resp := client.GET("/api/vega-backend/v1/catalogs?offset=0&limit=2")
			So(page1Resp.StatusCode, ShouldEqual, http.StatusOK)
			entries1 := page1Resp.Body["entries"].([]any)
			So(len(entries1), ShouldBeLessThanOrEqualTo, 2)

			// 第二页
			page2Resp := client.GET("/api/vega-backend/v1/catalogs?offset=2&limit=2")
			So(page2Resp.StatusCode, ShouldEqual, http.StatusOK)
			entries2 := page2Resp.Body["entries"].([]any)
			So(len(entries2), ShouldBeLessThanOrEqualTo, 2)
		})

		Convey("OS204: 列表查询 - 默认分页参数", func() {
			defaultResp := client.GET("/api/vega-backend/v1/catalogs")
			So(defaultResp.StatusCode, ShouldEqual, http.StatusOK)
			So(defaultResp.Body, ShouldNotBeNil)
			So(defaultResp.Body["entries"], ShouldNotBeNil)
			_, hasTotalCount := defaultResp.Body["total_count"]
			So(hasTotalCount, ShouldBeTrue)
		})

		Convey("OS205: 查询验证connector_config字段", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			So(getResp.StatusCode, ShouldEqual, http.StatusOK)

			catalog := fixtures.ExtractFromEntriesResponse(getResp)
			So(catalog, ShouldNotBeNil)

			// 验证connector_config包含OpenSearch特定字段
			connectorConfig, ok := catalog["connector_config"].(map[string]any)
			So(ok, ShouldBeTrue)
			So(connectorConfig["host"], ShouldNotBeEmpty)
			So(connectorConfig["port"], ShouldNotBeNil)
		})
	})

	_ = ctx
}

// TestOpenSearchSpecificUpdate OpenSearch特定功能AT测试 - 更新
// 编号规则：OS3xx为OpenSearch特定更新测试
func TestOpenSearchSpecificUpdate(t *testing.T) {
	var (
		ctx     context.Context
		config  *setup.TestConfig
		client  *testutil.HTTPClient
		builder *catalogfixtures.OpenSearchPayloadBuilder
	)

	Convey("OpenSearch特定功能AT测试 - 更新 - 初始化", t, func() {
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

		// 创建OpenSearch payload构建器
		builder = catalogfixtures.NewOpenSearchPayloadBuilder(config.TargetOpenSearch)

		// 清理现有catalog
		fixtures.CleanupCatalogs(client, t)

		// ========== OpenSearch更新测试（OS301-OS320）- 基于原数据更新 ==========

		Convey("OS301: 整体更新catalog connector_config", func() {
			// 创建catalog
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			// 获取原始数据
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			originalData := fixtures.ExtractFromEntriesResponse(getResp)

			// 基于原数据构建更新payload
			osConfig := builder.GetConfig()
			updatePayload := catalogfixtures.BuildUpdatePayload(originalData, map[string]any{
				"connector_config": map[string]any{
					"host":     osConfig.Host,
					"port":     osConfig.Port,
					"username": osConfig.Username,
					"password": osConfig.Password,
					"use_ssl":  osConfig.UseSSL,
					"options": map[string]any{
						"timeout":     "60s",
						"max_retries": 5,
						"compress":    true,
					},
				},
			})
			updateResp := client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("OS302: 同时更新多个字段", func() {
			// 创建catalog
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			// 获取原始数据
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			originalData := fixtures.ExtractFromEntriesResponse(getResp)

			// 基于原数据构建更新payload
			updatePayload := catalogfixtures.BuildUpdatePayload(originalData, map[string]any{
				"name":        catalogfixtures.GenerateUniqueName("multi-update-os-catalog"),
				"description": "同时更新多个字段的OpenSearch测试",
				"tags":        []string{"multi-update", "opensearch", "test"},
			})

			updateResp := client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("OS303: 更新name超长", func() {
			// 创建catalog
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			// 获取原始数据
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			originalData := fixtures.ExtractFromEntriesResponse(getResp)

			// 基于原数据构建更新payload（超长name）
			updatePayload := catalogfixtures.BuildUpdatePayload(originalData, map[string]any{
				"name": strings.Repeat("a", 129),
			})
			updateResp := client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("OS304: 验证update_time更新", func() {
			// 创建catalog
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			// 等待1秒确保时间戳不同
			time.Sleep(1 * time.Second)

			// 获取原始数据
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			originalData := fixtures.ExtractFromEntriesResponse(getResp)
			originalUpdateTime := originalData["update_time"].(float64)

			// 基于原数据构建更新payload
			updatePayload := catalogfixtures.BuildUpdatePayload(originalData, map[string]any{
				"description": "验证update_time更新",
			})
			updateResp := client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)

			// 验证update_time已更新
			newGetResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			newData := fixtures.ExtractFromEntriesResponse(newGetResp)
			newUpdateTime := newData["update_time"].(float64)
			So(newUpdateTime, ShouldBeGreaterThan, originalUpdateTime)
		})

		Convey("OS305: 验证create_time不变", func() {
			// 创建catalog
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			// 获取原始数据
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			originalData := fixtures.ExtractFromEntriesResponse(getResp)
			originalCreateTime := originalData["create_time"].(float64)

			// 基于原数据构建更新payload
			updatePayload := catalogfixtures.BuildUpdatePayload(originalData, map[string]any{
				"description": "验证create_time不变",
			})
			updateResp := client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)

			// 验证create_time不变
			newGetResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			newData := fixtures.ExtractFromEntriesResponse(newGetResp)
			newCreateTime := newData["create_time"].(float64)
			So(newCreateTime, ShouldEqual, originalCreateTime)
		})
	})

	_ = ctx
}

// TestOpenSearchSpecificDelete OpenSearch特定功能AT测试 - 删除
// 编号规则：OS4xx为OpenSearch特定删除测试
func TestOpenSearchSpecificDelete(t *testing.T) {
	var (
		ctx     context.Context
		config  *setup.TestConfig
		client  *testutil.HTTPClient
		builder *catalogfixtures.OpenSearchPayloadBuilder
	)

	Convey("OpenSearch特定功能AT测试 - 删除 - 初始化", t, func() {
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

		// 创建OpenSearch payload构建器
		builder = catalogfixtures.NewOpenSearchPayloadBuilder(config.TargetOpenSearch)

		// 清理现有catalog
		fixtures.CleanupCatalogs(client, t)

		// ========== OpenSearch删除测试（OS401-OS420） ==========

		Convey("OS401: 删除后不能更新", func() {
			// 创建catalog
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			// 删除catalog
			deleteResp := client.DELETE("/api/vega-backend/v1/catalogs/" + catalogID)
			So(deleteResp.StatusCode, ShouldEqual, http.StatusNoContent)

			// 尝试更新已删除的catalog
			updateResp := client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, payload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("OS402: 删除后可以创建同名catalog", func() {
			// 创建catalog
			payload := builder.BuildCreatePayload()
			catalogName := payload["name"]
			createResp1 := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp1.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID1 := createResp1.Body["id"].(string)

			// 删除catalog
			deleteResp := client.DELETE("/api/vega-backend/v1/catalogs/" + catalogID1)
			So(deleteResp.StatusCode, ShouldEqual, http.StatusNoContent)

			// 创建同名catalog
			payload2 := builder.BuildCreatePayload()
			payload2["name"] = catalogName
			createResp2 := client.POST("/api/vega-backend/v1/catalogs", payload2)

			So(createResp2.StatusCode, ShouldEqual, http.StatusCreated)

			// 新创建的catalog应该有不同的ID
			catalogID2 := createResp2.Body["id"].(string)
			So(catalogID2, ShouldNotEqual, catalogID1)
		})

		Convey("OS403: 删除包含完整字段的catalog", func() {
			// 创建包含所有字段的catalog
			payload := builder.BuildFullCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			// 删除
			deleteResp := client.DELETE("/api/vega-backend/v1/catalogs/" + catalogID)
			So(deleteResp.StatusCode, ShouldEqual, http.StatusNoContent)

			// 验证删除成功
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			So(getResp.StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("OS404: 批量删除多个catalog", func() {
			// 创建3个catalog
			catalogIDs := make([]string, 3)
			for i := 0; i < 3; i++ {
				payload := builder.BuildCreatePayload()
				createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
				So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
				catalogIDs[i] = createResp.Body["id"].(string)
			}

			// 依次删除
			for _, catalogID := range catalogIDs {
				deleteResp := client.DELETE("/api/vega-backend/v1/catalogs/" + catalogID)
				So(deleteResp.StatusCode, ShouldEqual, http.StatusNoContent)
			}

			// 验证所有catalog都已删除
			for _, catalogID := range catalogIDs {
				getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
				So(getResp.StatusCode, ShouldEqual, http.StatusNotFound)
			}
		})

		Convey("OS405: 删除后列表中不再显示", func() {
			// 创建catalog
			payload := builder.BuildCreatePayload()
			catalogName := payload["name"]
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			// 删除catalog
			deleteResp := client.DELETE("/api/vega-backend/v1/catalogs/" + catalogID)
			So(deleteResp.StatusCode, ShouldEqual, http.StatusNoContent)

			// 查询列表
			listResp := client.GET("/api/vega-backend/v1/catalogs?offset=0&limit=1000")
			So(listResp.StatusCode, ShouldEqual, http.StatusOK)

			if listResp.Body != nil && listResp.Body["entries"] != nil {
				entries := listResp.Body["entries"].([]any)

				// 验证删除的catalog不在列表中
				found := false
				for _, entry := range entries {
					catalogEntry := entry.(map[string]any)
					if catalogEntry["name"] == catalogName {
						found = true
						break
					}
				}

				So(found, ShouldBeFalse)
			}
		})

		Convey("OS406: 删除catalog后健康状态不可查", func() {
			// 创建catalog
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			// 删除catalog
			deleteResp := client.DELETE("/api/vega-backend/v1/catalogs/" + catalogID)
			So(deleteResp.StatusCode, ShouldEqual, http.StatusNoContent)

			// 尝试查询健康状态
			statusResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID + "/health-status")
			So(statusResp.StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("OS407: 删除catalog后不能测试连接", func() {
			// 创建catalog
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			// 删除catalog
			deleteResp := client.DELETE("/api/vega-backend/v1/catalogs/" + catalogID)
			So(deleteResp.StatusCode, ShouldEqual, http.StatusNoContent)

			// 尝试测试连接
			testResp := client.POST("/api/vega-backend/v1/catalogs/"+catalogID+"/test-connection", nil)
			So(testResp.StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("OS408: 使用无效ID删除", func() {
			invalidIDs := []string{
				"invalid-id-format",
				"../../../etc/passwd",
				"<script>alert('xss')</script>",
			}

			for _, invalidID := range invalidIDs {
				Convey("无效ID: "+invalidID, func() {
					deleteResp := client.DELETE("/api/vega-backend/v1/catalogs/" + invalidID)
					So(deleteResp.StatusCode, ShouldEqual, http.StatusNotFound)
				})
			}
		})
	})

	_ = ctx
}

// TestOpenSearchConnectionVariations OpenSearch连接变体测试
func TestOpenSearchConnectionVariations(t *testing.T) {
	var (
		ctx     context.Context
		config  *setup.TestConfig
		client  *testutil.HTTPClient
		builder *catalogfixtures.OpenSearchPayloadBuilder
	)

	Convey("OpenSearch连接变体测试 - 初始化", t, func() {
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

		// 创建OpenSearch payload构建器
		builder = catalogfixtures.NewOpenSearchPayloadBuilder(config.TargetOpenSearch)

		// 清理现有catalog
		fixtures.CleanupCatalogs(client, t)

		// ========== 连接选项变体测试 ==========

		Convey("OS501: 创建带超时选项的catalog", func() {
			options := map[string]any{
				"timeout": "60s",
			}
			payload := builder.BuildCreatePayloadWithOptions(options)
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("OS502: 创建带重试选项的catalog", func() {
			options := map[string]any{
				"max_retries": 10,
			}
			payload := builder.BuildCreatePayloadWithOptions(options)
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("OS503: 创建带压缩选项的catalog", func() {
			options := map[string]any{
				"compress": true,
			}
			payload := builder.BuildCreatePayloadWithOptions(options)
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("OS504: 创建带多个选项的catalog", func() {
			options := map[string]any{
				"timeout":            "30s",
				"max_retries":        5,
				"compress":           true,
				"discovery_interval": "5m",
			}
			payload := builder.BuildCreatePayloadWithOptions(options)
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("OS505: 创建带空options的catalog", func() {
			payload := builder.BuildCreatePayloadWithOptions(map[string]any{})
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})
	})

	_ = ctx
}
