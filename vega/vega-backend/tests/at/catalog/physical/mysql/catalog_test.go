// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package mysql

import (
	"context"
	"net/http"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	cataloghelpers "vega-backend-tests/at/catalog/helpers"
	"vega-backend-tests/at/setup"
	"vega-backend-tests/testutil"
)

// TestMySQLCatalogCreate MySQL Catalog创建AT测试
// 编号规则：MY1xx
func TestMySQLCatalogCreate(t *testing.T) {
	var (
		ctx     context.Context
		config  *setup.TestConfig
		client  *testutil.HTTPClient
		builder *MySQLPayloadBuilder
	)

	Convey("MySQL Catalog创建AT测试 - 初始化", t, func() {
		ctx = context.Background()

		var err error
		config, err = setup.LoadTestConfig()
		So(err, ShouldBeNil)
		So(config, ShouldNotBeNil)
		So(config.TargetMySQL.Host, ShouldNotBeEmpty)

		client = testutil.NewHTTPClient(config.VegaManager.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)
		t.Logf("✓ AT测试环境就绪，VEGA Manager: %s", config.VegaManager.BaseURL)

		builder = NewMySQLPayloadBuilder(config.TargetMySQL)
		builder.SetTestConfig(config)

		cataloghelpers.CleanupCatalogs(client, t)

		// ========== 正向测试（MY101-MY110） ==========

		Convey("MY101: 创建MySQL catalog - 基本场景", func() {
			payload := builder.BuildCreatePayload()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
			So(resp.Body["id"], ShouldNotBeEmpty)
		})

		Convey("MY102: 创建后验证connector_type为mysql", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			catalog := cataloghelpers.ExtractFromEntriesResponse(getResp)
			So(catalog["connector_type"], ShouldEqual, "mysql")
		})

		Convey("MY103: 创建后验证type为physical", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			catalog := cataloghelpers.ExtractFromEntriesResponse(getResp)
			So(catalog["type"], ShouldEqual, cataloghelpers.CatalogTypePhysical)
		})

		Convey("MY104: 创建MySQL catalog - 完整字段", func() {
			payload := builder.BuildFullCreatePayload()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
			So(resp.Body["id"], ShouldNotBeEmpty)

			catalogID := resp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			catalog := cataloghelpers.ExtractFromEntriesResponse(getResp)
			So(catalog["description"], ShouldNotBeEmpty)
			tags, ok := catalog["tags"].([]any)
			So(ok, ShouldBeTrue)
			So(len(tags), ShouldBeGreaterThan, 0)
		})

		Convey("MY105: 创建带MySQL特定options（charset/timeout）", func() {
			options := map[string]any{
				"charset": "utf8mb4",
				"timeout": "10s",
			}
			payload := builder.BuildCreatePayloadWithOptions(options)
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("MY106: 创建后立即查询", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			So(getResp.StatusCode, ShouldEqual, http.StatusOK)

			catalog := cataloghelpers.ExtractFromEntriesResponse(getResp)
			So(catalog["id"], ShouldEqual, catalogID)
			So(catalog["name"], ShouldEqual, payload["name"])
		})

		Convey("MY107: MySQL连接测试成功", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			testResp := client.POST("/api/vega-backend/v1/catalogs/"+catalogID+"/test-connection", nil)
			So(testResp.StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("MY108: 获取MySQL catalog健康状态", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			statusResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID + "/health-status")
			So(statusResp.StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("MY109: 创建实例级MySQL catalog（不指定database）", func() {
			payload := builder.BuildCreatePayloadWithoutDatabase()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("MY110: 实例级MySQL catalog连接测试成功", func() {
			payload := builder.BuildCreatePayloadWithoutDatabase()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			testResp := client.POST("/api/vega-backend/v1/catalogs/"+catalogID+"/test-connection", nil)
			So(testResp.StatusCode, ShouldEqual, http.StatusOK)
		})

		// ========== connector_config负向测试（MY121-MY129） ==========

		Convey("MY121: 缺少host字段", func() {
			mysqlConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           cataloghelpers.GenerateUniqueName("missing-host"),
				"connector_type": "mysql",
				"connector_config": map[string]any{
					"port":      mysqlConfig.Port,
					"databases": []string{mysqlConfig.Database},
					"username":  mysqlConfig.Username,
					"password":  builder.GetEncryptedPassword(),
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MY122: 缺少port字段", func() {
			mysqlConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           cataloghelpers.GenerateUniqueName("missing-port"),
				"connector_type": "mysql",
				"connector_config": map[string]any{
					"host":      mysqlConfig.Host,
					"databases": []string{mysqlConfig.Database},
					"username":  mysqlConfig.Username,
					"password":  builder.GetEncryptedPassword(),
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MY123: 缺少user字段", func() {
			mysqlConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           cataloghelpers.GenerateUniqueName("missing-user"),
				"connector_type": "mysql",
				"connector_config": map[string]any{
					"host":      mysqlConfig.Host,
					"port":      mysqlConfig.Port,
					"databases": []string{mysqlConfig.Database},
					"password":  builder.GetEncryptedPassword(),
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MY124: 空用户名", func() {
			mysqlConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           cataloghelpers.GenerateUniqueName("empty-user"),
				"connector_type": "mysql",
				"connector_config": map[string]any{
					"host":      mysqlConfig.Host,
					"port":      mysqlConfig.Port,
					"databases": []string{mysqlConfig.Database},
					"username":  "",
					"password":  mysqlConfig.Password,
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MY125: 错误密码", func() {
			payload := builder.BuildCreatePayloadWithWrongCredentials()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MY126: 不存在的数据库", func() {
			payload := builder.BuildCreatePayloadWithNonExistentDB()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MY127: 无效端口（非数字）", func() {
			payload := builder.BuildCreatePayloadWithInvalidPort()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MY128: 超出范围端口（65536）", func() {
			mysqlConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           cataloghelpers.GenerateUniqueName("overflow-port"),
				"connector_type": "mysql",
				"connector_config": map[string]any{
					"host":      mysqlConfig.Host,
					"port":      65536,
					"databases": []string{mysqlConfig.Database},
					"username":  mysqlConfig.Username,
					"password":  mysqlConfig.Password,
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MY129: 负数端口", func() {
			mysqlConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           cataloghelpers.GenerateUniqueName("negative-port"),
				"connector_type": "mysql",
				"connector_config": map[string]any{
					"host":      mysqlConfig.Host,
					"port":      -1,
					"databases": []string{mysqlConfig.Database},
					"username":  mysqlConfig.Username,
					"password":  mysqlConfig.Password,
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		// ========== 边界测试（MY131-MY138） ==========

		Convey("MY131: port边界值（1）", func() {
			mysqlConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           cataloghelpers.GenerateUniqueName("port-1"),
				"connector_type": "mysql",
				"connector_config": map[string]any{
					"host":      mysqlConfig.Host,
					"port":      1,
					"databases": []string{mysqlConfig.Database},
					"username":  mysqlConfig.Username,
					"password":  builder.GetEncryptedPassword(),
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldBeIn, []int{http.StatusCreated, http.StatusBadRequest})
		})

		Convey("MY132: port边界值（65535）", func() {
			mysqlConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           cataloghelpers.GenerateUniqueName("port-65535"),
				"connector_type": "mysql",
				"connector_config": map[string]any{
					"host":      mysqlConfig.Host,
					"port":      65535,
					"databases": []string{mysqlConfig.Database},
					"username":  mysqlConfig.Username,
					"password":  builder.GetEncryptedPassword(),
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldBeIn, []int{http.StatusCreated, http.StatusBadRequest})
		})

		Convey("MY133: database名称最大长度（64字符）", func() {
			mysqlConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           cataloghelpers.GenerateUniqueName("long-db"),
				"connector_type": "mysql",
				"connector_config": map[string]any{
					"host":      mysqlConfig.Host,
					"port":      mysqlConfig.Port,
					"databases": []string{strings.Repeat("d", 64)},
					"username":  mysqlConfig.Username,
					"password":  builder.GetEncryptedPassword(),
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldBeIn, []int{http.StatusCreated, http.StatusBadRequest})
		})

		Convey("MY134: database名称超过最大长度", func() {
			mysqlConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           cataloghelpers.GenerateUniqueName("too-long-db"),
				"connector_type": "mysql",
				"connector_config": map[string]any{
					"host":      mysqlConfig.Host,
					"port":      mysqlConfig.Port,
					"databases": []string{strings.Repeat("d", 65)},
					"username":  mysqlConfig.Username,
					"password":  builder.GetEncryptedPassword(),
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MY135: host为IP地址", func() {
			payload := builder.BuildCreatePayload()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("MY136: host为域名", func() {
			mysqlConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           cataloghelpers.GenerateUniqueName("domain-host"),
				"connector_type": "mysql",
				"connector_config": map[string]any{
					"host":      "localhost",
					"port":      mysqlConfig.Port,
					"databases": []string{mysqlConfig.Database},
					"username":  mysqlConfig.Username,
					"password":  builder.GetEncryptedPassword(),
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldBeIn, []int{http.StatusCreated, http.StatusBadRequest})
		})

		Convey("MY137: 不指定database（实例级连接）", func() {
			payload := builder.BuildCreatePayloadWithoutDatabase()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("MY138: password为空（无密码连接）", func() {
			mysqlConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           cataloghelpers.GenerateUniqueName("no-password"),
				"connector_type": "mysql",
				"connector_config": map[string]any{
					"host":      mysqlConfig.Host,
					"port":      mysqlConfig.Port,
					"databases": []string{mysqlConfig.Database},
					"username":  mysqlConfig.Username,
					"password":  "",
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldBeIn, []int{http.StatusCreated, http.StatusBadRequest})
		})
	})

	_ = ctx
}

// TestMySQLCatalogRead MySQL Catalog读取AT测试
// 编号规则：MY2xx
func TestMySQLCatalogRead(t *testing.T) {
	var (
		ctx     context.Context
		config  *setup.TestConfig
		client  *testutil.HTTPClient
		builder *MySQLPayloadBuilder
	)

	Convey("MySQL Catalog读取AT测试 - 初始化", t, func() {
		ctx = context.Background()

		var err error
		config, err = setup.LoadTestConfig()
		So(err, ShouldBeNil)

		client = testutil.NewHTTPClient(config.VegaManager.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		builder = NewMySQLPayloadBuilder(config.TargetMySQL)
		builder.SetTestConfig(config)

		cataloghelpers.CleanupCatalogs(client, t)

		// ========== 读取测试（MY201-MY205） ==========

		Convey("MY201: 获取存在的MySQL catalog", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			So(getResp.StatusCode, ShouldEqual, http.StatusOK)

			catalog := cataloghelpers.ExtractFromEntriesResponse(getResp)
			So(catalog["id"], ShouldEqual, catalogID)
		})

		Convey("MY202: 列表查询 - 按type过滤physical", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			listResp := client.GET("/api/vega-backend/v1/catalogs?type=physical&offset=0&limit=100")
			So(listResp.StatusCode, ShouldEqual, http.StatusOK)

			if entries, ok := listResp.Body["entries"].([]any); ok {
				So(len(entries), ShouldBeGreaterThanOrEqualTo, 1)
				for _, entry := range entries {
					So(entry.(map[string]any)["type"], ShouldEqual, "physical")
				}
			}
		})

		Convey("MY203: 列表查询 - 按connector_type过滤mysql", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			listResp := client.GET("/api/vega-backend/v1/catalogs?connector_type=mysql&offset=0&limit=100")
			So(listResp.StatusCode, ShouldEqual, http.StatusOK)

			if entries, ok := listResp.Body["entries"].([]any); ok {
				So(len(entries), ShouldBeGreaterThanOrEqualTo, 1)
				for _, entry := range entries {
					So(entry.(map[string]any)["connector_type"], ShouldEqual, "mysql")
				}
			}
		})

		Convey("MY204: 查询catalog - 验证所有字段返回", func() {
			payload := builder.BuildFullCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			So(getResp.StatusCode, ShouldEqual, http.StatusOK)

			catalog := cataloghelpers.ExtractFromEntriesResponse(getResp)
			So(catalog["id"], ShouldNotBeEmpty)
			So(catalog["name"], ShouldNotBeEmpty)
			So(catalog["type"], ShouldEqual, cataloghelpers.CatalogTypePhysical)
			So(catalog["connector_type"], ShouldEqual, "mysql")
			So(catalog["create_time"], ShouldNotBeZeroValue)
			So(catalog["update_time"], ShouldNotBeZeroValue)
		})

		Convey("MY205: 验证connector_config.password不返回", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			So(getResp.StatusCode, ShouldEqual, http.StatusOK)

			catalog := cataloghelpers.ExtractFromEntriesResponse(getResp)
			if connCfg, ok := catalog["connector_config"].(map[string]any); ok {
				_, hasPassword := connCfg["password"]
				So(hasPassword, ShouldBeFalse)
			}
		})
	})

	_ = ctx
}

// TestMySQLCatalogUpdate MySQL Catalog更新AT测试
// 编号规则：MY3xx
func TestMySQLCatalogUpdate(t *testing.T) {
	var (
		ctx     context.Context
		config  *setup.TestConfig
		client  *testutil.HTTPClient
		builder *MySQLPayloadBuilder
	)

	Convey("MySQL Catalog更新AT测试 - 初始化", t, func() {
		ctx = context.Background()

		var err error
		config, err = setup.LoadTestConfig()
		So(err, ShouldBeNil)

		client = testutil.NewHTTPClient(config.VegaManager.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		builder = NewMySQLPayloadBuilder(config.TargetMySQL)
		builder.SetTestConfig(config)

		cataloghelpers.CleanupCatalogs(client, t)

		// ========== 更新测试（MY301-MY305） ==========

		Convey("MY301: 整体更新connector_config", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			originalData := cataloghelpers.ExtractFromEntriesResponse(getResp)

			mysqlConfig := builder.GetConfig()
			updatePayload := cataloghelpers.BuildUpdatePayload(originalData, map[string]any{
				"connector_config": map[string]any{
					"host":      mysqlConfig.Host,
					"port":      mysqlConfig.Port,
					"databases": []string{mysqlConfig.Database},
					"username":  mysqlConfig.Username,
					"password":  builder.GetEncryptedPassword(),
					"options": map[string]any{
						"charset": "utf8mb4",
					},
				},
			})
			updateResp := client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("MY302: 更新connector_config后连接测试", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			originalData := cataloghelpers.ExtractFromEntriesResponse(getResp)

			mysqlConfig := builder.GetConfig()
			updatePayload := cataloghelpers.BuildUpdatePayload(originalData, map[string]any{
				"connector_config": map[string]any{
					"host":      mysqlConfig.Host,
					"port":      mysqlConfig.Port,
					"databases": []string{mysqlConfig.Database},
					"username":  mysqlConfig.Username,
					"password":  builder.GetEncryptedPassword(),
				},
			})
			updateResp := client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)

			testResp := client.POST("/api/vega-backend/v1/catalogs/"+catalogID+"/test-connection", nil)
			So(testResp.StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("MY303: 更新host为无效地址", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			originalData := cataloghelpers.ExtractFromEntriesResponse(getResp)

			mysqlConfig := builder.GetConfig()
			updatePayload := cataloghelpers.BuildUpdatePayload(originalData, map[string]any{
				"connector_config": map[string]any{
					"host":      "invalid-host-12345.example.com",
					"port":      mysqlConfig.Port,
					"databases": []string{mysqlConfig.Database},
					"username":  mysqlConfig.Username,
					"password":  builder.GetEncryptedPassword(),
				},
			})
			updateResp := client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MY304: 更新port为无效值", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			originalData := cataloghelpers.ExtractFromEntriesResponse(getResp)

			mysqlConfig := builder.GetConfig()
			updatePayload := cataloghelpers.BuildUpdatePayload(originalData, map[string]any{
				"connector_config": map[string]any{
					"host":      mysqlConfig.Host,
					"port":      65536,
					"databases": []string{mysqlConfig.Database},
					"username":  mysqlConfig.Username,
					"password":  builder.GetEncryptedPassword(),
				},
			})
			updateResp := client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MY305: 更新password", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			originalData := cataloghelpers.ExtractFromEntriesResponse(getResp)

			mysqlConfig := builder.GetConfig()
			updatePayload := cataloghelpers.BuildUpdatePayload(originalData, map[string]any{
				"connector_config": map[string]any{
					"host":      mysqlConfig.Host,
					"port":      mysqlConfig.Port,
					"databases": []string{mysqlConfig.Database},
					"username":  mysqlConfig.Username,
					"password":  builder.GetEncryptedPassword(),
				},
			})
			updateResp := client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)
		})
	})

	_ = ctx
}

// TestMySQLCatalogDelete MySQL Catalog删除AT测试
// 编号规则：MY4xx
func TestMySQLCatalogDelete(t *testing.T) {
	var (
		ctx     context.Context
		config  *setup.TestConfig
		client  *testutil.HTTPClient
		builder *MySQLPayloadBuilder
	)

	Convey("MySQL Catalog删除AT测试 - 初始化", t, func() {
		ctx = context.Background()

		var err error
		config, err = setup.LoadTestConfig()
		So(err, ShouldBeNil)

		client = testutil.NewHTTPClient(config.VegaManager.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		builder = NewMySQLPayloadBuilder(config.TargetMySQL)
		builder.SetTestConfig(config)

		cataloghelpers.CleanupCatalogs(client, t)

		// ========== 删除测试（MY401-MY402） ==========

		Convey("MY401: 删除MySQL catalog后健康状态不可查", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			deleteResp := client.DELETE("/api/vega-backend/v1/catalogs/" + catalogID)
			So(deleteResp.StatusCode, ShouldEqual, http.StatusNoContent)

			statusResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID + "/health-status")
			So(statusResp.StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("MY402: 删除MySQL catalog后不能测试连接", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			deleteResp := client.DELETE("/api/vega-backend/v1/catalogs/" + catalogID)
			So(deleteResp.StatusCode, ShouldEqual, http.StatusNoContent)

			testResp := client.POST("/api/vega-backend/v1/catalogs/"+catalogID+"/test-connection", nil)
			So(testResp.StatusCode, ShouldEqual, http.StatusNotFound)
		})
	})

	_ = ctx
}

// TestMySQLSpecificOptions MySQL特有选项测试
// 编号规则：MY5xx
func TestMySQLSpecificOptions(t *testing.T) {
	var (
		ctx     context.Context
		config  *setup.TestConfig
		client  *testutil.HTTPClient
		builder *MySQLPayloadBuilder
	)

	Convey("MySQL特有选项AT测试 - 初始化", t, func() {
		ctx = context.Background()

		var err error
		config, err = setup.LoadTestConfig()
		So(err, ShouldBeNil)

		client = testutil.NewHTTPClient(config.VegaManager.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		builder = NewMySQLPayloadBuilder(config.TargetMySQL)
		builder.SetTestConfig(config)

		cataloghelpers.CleanupCatalogs(client, t)

		// ========== MySQL特有选项测试（MY501-MY506） ==========

		Convey("MY501: MySQL charset选项测试（utf8mb4）", func() {
			options := map[string]any{"charset": "utf8mb4"}
			payload := builder.BuildCreatePayloadWithOptions(options)
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("MY502: MySQL parseTime选项测试", func() {
			options := map[string]any{"parseTime": "true"}
			payload := builder.BuildCreatePayloadWithOptions(options)
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("MY503: MySQL loc选项测试（时区）", func() {
			options := map[string]any{"loc": "Local"}
			payload := builder.BuildCreatePayloadWithOptions(options)
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("MY504: MySQL timeout选项测试", func() {
			options := map[string]any{"timeout": "10s"}
			payload := builder.BuildCreatePayloadWithOptions(options)
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("MY505: MySQL SSL连接测试", func() {
			options := map[string]any{"tls": "skip-verify"}
			payload := builder.BuildCreatePayloadWithOptions(options)
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldBeIn, []int{http.StatusCreated, http.StatusBadRequest})
		})

		Convey("MY506: MySQL collation选项测试", func() {
			options := map[string]any{"collation": "utf8mb4_unicode_ci"}
			payload := builder.BuildCreatePayloadWithOptions(options)
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})
	})

	_ = ctx
}
