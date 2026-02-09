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

	"vega-backend/interfaces"
	"vega-backend/tests/at/fixtures"
	catalogfixtures "vega-backend/tests/at/fixtures/catalog"
	"vega-backend/tests/at/setup"
	"vega-backend/tests/testutil"
)

// TestMySQLSpecificCreate MySQL特定功能AT测试 - 创建
// 编号规则：MY1xx为MySQL特定创建测试
// - MY101-MY120: 正向测试
// - MY121-MY140: 负向测试
func TestMySQLSpecificCreate(t *testing.T) {
	var (
		ctx     context.Context
		config  *setup.TestConfig
		client  *testutil.HTTPClient
		builder *catalogfixtures.MySQLPayloadBuilder
	)

	Convey("MySQL特定功能AT测试 - 创建 - 初始化", t, func() {
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
		builder.SetTestConfig(config)

		// 清理现有catalog
		fixtures.CleanupCatalogs(client, t)

		// ========== MySQL正向测试（MY101-MY120） ==========

		Convey("MY101: 创建MySQL catalog后验证connector_type", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			// 查询验证
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			So(getResp.StatusCode, ShouldEqual, http.StatusOK)

			catalog := fixtures.ExtractFromEntriesResponse(getResp)
			So(catalog, ShouldNotBeNil)
			So(catalog["connector_type"], ShouldEqual, "mysql")
			So(catalog["type"], ShouldEqual, interfaces.CatalogTypePhysical)
		})

		Convey("MY102: 创建带MySQL特定options的catalog", func() {
			options := map[string]any{
				"charset":   "utf8mb4",
				"parseTime": "true",
				"loc":       "Local",
				"timeout":   "10s",
			}
			payload := builder.BuildCreatePayloadWithOptions(options)
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
			So(resp.Body["id"], ShouldNotBeEmpty)
		})

		Convey("MY103: 创建完整字段的MySQL catalog", func() {
			payload := builder.BuildFullCreatePayload()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
			So(resp.Body["id"], ShouldNotBeEmpty)

			// 验证返回的字段
			catalogID := resp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			catalog := fixtures.ExtractFromEntriesResponse(getResp)

			So(catalog["description"], ShouldEqual, "完整的测试catalog，包含所有可选字段")
			tags, ok := catalog["tags"].([]any)
			So(ok, ShouldBeTrue)
			So(len(tags), ShouldBeGreaterThan, 0)
		})

		Convey("MY104: MySQL连接测试成功", func() {
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

		Convey("MY105: 获取MySQL catalog健康状态", func() {
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

		Convey("MY106: 创建实例级MySQL catalog（不指定database）", func() {
			payload := builder.BuildCreatePayloadWithoutDatabase()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
			So(createResp.Body["id"], ShouldNotBeEmpty)

			catalogID := createResp.Body["id"].(string)

			// 验证connector_type
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			So(getResp.StatusCode, ShouldEqual, http.StatusOK)
			catalog := fixtures.ExtractFromEntriesResponse(getResp)
			So(catalog["connector_type"], ShouldEqual, "mysql")
		})

		Convey("MY107: 实例级MySQL catalog连接测试成功", func() {
			payload := builder.BuildCreatePayloadWithoutDatabase()
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

		// ========== MySQL负向测试（MY121-MY140） ==========

		Convey("MY121: 无效端口测试（非数字）", func() {
			payload := builder.BuildCreatePayloadWithInvalidPort()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MY122: 错误密码测试", func() {
			payload := builder.BuildCreatePayloadWithWrongCredentials()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MY123: 不存在的数据库测试", func() {
			payload := builder.BuildCreatePayloadWithNonExistentDB()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MY124: 缺少host字段测试", func() {
			mysqlConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           catalogfixtures.GenerateUniqueName("missing-host-catalog"),
				"connector_type": "mysql",
				"connector_config": map[string]any{
					// 缺少host
					"port":     mysqlConfig.Port,
					"database": mysqlConfig.Database,
					"username": mysqlConfig.Username,
					"password": builder.GetEncryptedPassword(),
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MY125: 缺少port字段测试", func() {
			mysqlConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           catalogfixtures.GenerateUniqueName("missing-port-catalog"),
				"connector_type": "mysql",
				"connector_config": map[string]any{
					"host": mysqlConfig.Host,
					// 缺少port
					"database": mysqlConfig.Database,
					"username": mysqlConfig.Username,
					"password": builder.GetEncryptedPassword(),
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MY126: 不指定database字段测试（实例级连接）", func() {
			// database 为可选字段，不指定时创建实例级连接，应成功
			payload := builder.BuildCreatePayloadWithoutDatabase()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
			So(resp.Body["id"], ShouldNotBeEmpty)
		})

		Convey("MY127: 空用户名测试", func() {
			mysqlConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           catalogfixtures.GenerateUniqueName("empty-username-catalog"),
				"connector_type": "mysql",
				"connector_config": map[string]any{
					"host":     mysqlConfig.Host,
					"port":     mysqlConfig.Port,
					"database": mysqlConfig.Database,
					"username": "",
					"password": mysqlConfig.Password,
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MY128: 超出范围端口测试（65536）", func() {
			mysqlConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           catalogfixtures.GenerateUniqueName("overflow-port-catalog"),
				"connector_type": "mysql",
				"connector_config": map[string]any{
					"host":     mysqlConfig.Host,
					"port":     65536,
					"database": mysqlConfig.Database,
					"username": mysqlConfig.Username,
					"password": mysqlConfig.Password,
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MY129: 负数端口测试", func() {
			mysqlConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           catalogfixtures.GenerateUniqueName("negative-port-catalog"),
				"connector_type": "mysql",
				"connector_config": map[string]any{
					"host":     mysqlConfig.Host,
					"port":     -1,
					"database": mysqlConfig.Database,
					"username": mysqlConfig.Username,
					"password": mysqlConfig.Password,
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})
	})

	_ = ctx
}

// TestMySQLSpecificRead MySQL特定功能AT测试 - 读取
// 编号规则：MY2xx为MySQL特定读取测试
func TestMySQLSpecificRead(t *testing.T) {
	var (
		ctx     context.Context
		config  *setup.TestConfig
		client  *testutil.HTTPClient
		builder *catalogfixtures.MySQLPayloadBuilder
	)

	Convey("MySQL特定功能AT测试 - 读取 - 初始化", t, func() {
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
		builder.SetTestConfig(config)

		// 清理现有catalog
		fixtures.CleanupCatalogs(client, t)

		// ========== MySQL读取测试（MY201-MY220） ==========

		Convey("MY201: 查询catalog - 验证所有字段返回", func() {
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
			So(catalog["type"], ShouldEqual, interfaces.CatalogTypePhysical)
			So(catalog["connector_type"], ShouldEqual, "mysql")
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

		Convey("MY202: 列表查询 - 按type过滤physical", func() {
			// 创建1个physical catalog
			physicalPayload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", physicalPayload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			// 查询physical类型
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

		Convey("MY203: 列表查询 - 按type过滤logical", func() {
			// 创建1个logical catalog
			logicalPayload := catalogfixtures.BuildLogicalCatalogPayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", logicalPayload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			// 查询logical类型
			logicalResp := client.GET("/api/vega-backend/v1/catalogs?type=logical&offset=0&limit=100")
			So(logicalResp.StatusCode, ShouldEqual, http.StatusOK)

			if logicalResp.Body != nil && logicalResp.Body["entries"] != nil {
				entries := logicalResp.Body["entries"].([]any)
				So(len(entries), ShouldBeGreaterThanOrEqualTo, 1)

				// 验证都是logical类型
				for _, entry := range entries {
					catalogEntry := entry.(map[string]any)
					So(catalogEntry["type"], ShouldEqual, "logical")
				}
			}
		})

		Convey("MY204: 列表查询 - 分页测试", func() {
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

		Convey("MY205: 列表查询 - 默认分页参数", func() {
			defaultResp := client.GET("/api/vega-backend/v1/catalogs")
			So(defaultResp.StatusCode, ShouldEqual, http.StatusOK)
			So(defaultResp.Body, ShouldNotBeNil)
			So(defaultResp.Body["entries"], ShouldNotBeNil)
			_, hasTotalCount := defaultResp.Body["total_count"]
			So(hasTotalCount, ShouldBeTrue)
		})
	})

	_ = ctx
}

// TestMySQLSpecificUpdate MySQL特定功能AT测试 - 更新
// 编号规则：MY3xx为MySQL特定更新测试
func TestMySQLSpecificUpdate(t *testing.T) {
	var (
		ctx     context.Context
		config  *setup.TestConfig
		client  *testutil.HTTPClient
		builder *catalogfixtures.MySQLPayloadBuilder
	)

	Convey("MySQL特定功能AT测试 - 更新 - 初始化", t, func() {
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
		builder.SetTestConfig(config)

		// 清理现有catalog
		fixtures.CleanupCatalogs(client, t)

		// ========== MySQL更新测试（MY301-MY320）- 基于原数据更新 ==========

		Convey("MY301: 整体更新catalog connector_config", func() {
			// 创建catalog
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			// 获取原始数据
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			originalData := fixtures.ExtractFromEntriesResponse(getResp)

			// 基于原数据构建更新payload
			mysqlConfig := builder.GetConfig()
			updatePayload := catalogfixtures.BuildUpdatePayload(originalData, map[string]any{
				"connector_config": map[string]any{
					"host":     mysqlConfig.Host,
					"port":     mysqlConfig.Port,
					"database": mysqlConfig.Database,
					"username": mysqlConfig.Username,
					"password": builder.GetEncryptedPassword(), // 使用加密密码
					"options": map[string]any{
						"charset":   "utf8mb4",
						"parseTime": "true",
						"timeout":   "30s",
					},
				},
			})
			updateResp := client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("MY302: 同时更新多个字段", func() {
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
				"name":        catalogfixtures.GenerateUniqueName("multi-update-catalog"),
				"description": "同时更新多个字段的测试",
				"tags":        []string{"multi-update", "test"},
			})
			// GET响应不返回敏感字段，需要注入加密密码
			catalogfixtures.InjectEncryptedPassword(updatePayload, builder.GetEncryptedPassword())

			updateResp := client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("MY303: 更新name超长", func() {
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
			// GET响应不返回敏感字段，需要注入加密密码
			catalogfixtures.InjectEncryptedPassword(updatePayload, builder.GetEncryptedPassword())
			updateResp := client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MY304: 验证update_time更新", func() {
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
			// GET响应不返回敏感字段，需要注入加密密码
			catalogfixtures.InjectEncryptedPassword(updatePayload, builder.GetEncryptedPassword())
			updateResp := client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)

			// 验证update_time已更新
			newGetResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			newData := fixtures.ExtractFromEntriesResponse(newGetResp)
			newUpdateTime := newData["update_time"].(float64)
			So(newUpdateTime, ShouldBeGreaterThan, originalUpdateTime)
		})

		Convey("MY305: 验证create_time不变", func() {
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
			// GET响应不返回敏感字段，需要注入加密密码
			catalogfixtures.InjectEncryptedPassword(updatePayload, builder.GetEncryptedPassword())
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

// TestMySQLSpecificDelete MySQL特定功能AT测试 - 删除
// 编号规则：MY4xx为MySQL特定删除测试
func TestMySQLSpecificDelete(t *testing.T) {
	var (
		ctx     context.Context
		config  *setup.TestConfig
		client  *testutil.HTTPClient
		builder *catalogfixtures.MySQLPayloadBuilder
	)

	Convey("MySQL特定功能AT测试 - 删除 - 初始化", t, func() {
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
		builder.SetTestConfig(config)

		// 清理现有catalog
		fixtures.CleanupCatalogs(client, t)

		// ========== MySQL删除测试（MY401-MY420） ==========

		Convey("MY401: 删除后不能更新", func() {
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

		Convey("MY402: 删除后可以创建同名catalog", func() {
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

		Convey("MY403: 删除包含完整字段的catalog", func() {
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

		Convey("MY404: 批量删除多个catalog", func() {
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

		Convey("MY405: 删除后列表中不再显示", func() {
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

		Convey("MY406: 删除catalog后健康状态不可查", func() {
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

		Convey("MY407: 删除catalog后不能测试连接", func() {
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

		Convey("MY408: 使用无效ID删除", func() {
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
