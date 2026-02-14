// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package mariadb

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

// TestMariaDBCatalogCreate MariaDB Catalog创建AT测试
// 编号规则：MD1xx
func TestMariaDBCatalogCreate(t *testing.T) {
	var (
		ctx     context.Context
		config  *setup.TestConfig
		client  *testutil.HTTPClient
		builder *MariaDBPayloadBuilder
	)

	Convey("MariaDB Catalog创建AT测试 - 初始化", t, func() {
		ctx = context.Background()

		var err error
		config, err = setup.LoadTestConfig()
		So(err, ShouldBeNil)
		So(config, ShouldNotBeNil)
		So(config.TargetMariaDB.Host, ShouldNotBeEmpty)

		client = testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)
		t.Logf("✓ AT测试环境就绪，VEGA Manager: %s", config.VegaBackend.BaseURL)

		builder = NewMariaDBPayloadBuilder(config.TargetMariaDB)
		builder.SetTestConfig(config)

		cataloghelpers.CleanupCatalogs(client, t)

		// ========== 正向测试（MD101-MD110） ==========

		Convey("MD101: 创建MariaDB catalog - 基本场景", func() {
			payload := builder.BuildCreatePayload()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
			So(resp.Body["id"], ShouldNotBeEmpty)
		})

		Convey("MD102: 创建后验证connector_type为mariadb", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			catalog := cataloghelpers.ExtractFromEntriesResponse(getResp)
			So(catalog["connector_type"], ShouldEqual, "mariadb")
		})

		Convey("MD103: 创建后验证type为physical", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			catalog := cataloghelpers.ExtractFromEntriesResponse(getResp)
			So(catalog["type"], ShouldEqual, cataloghelpers.CatalogTypePhysical)
		})

		Convey("MD104: 创建MariaDB catalog - 完整字段", func() {
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

		Convey("MD105: 创建带MariaDB特定options（charset/timeout）", func() {
			options := map[string]any{
				"charset": "utf8mb4",
				"timeout": "10s",
			}
			payload := builder.BuildCreatePayloadWithOptions(options)
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("MD106: 创建后立即查询", func() {
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

		Convey("MD107: MariaDB连接测试成功", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			testResp := client.POST("/api/vega-backend/v1/catalogs/"+catalogID+"/test-connection", nil)
			So(testResp.StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("MD108: 获取MariaDB catalog健康状态", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			statusResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID + "/health-status")
			So(statusResp.StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("MD109: 创建实例级MariaDB catalog（不指定database）", func() {
			payload := builder.BuildCreatePayloadWithoutDatabase()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("MD110: 实例级MariaDB catalog连接测试成功", func() {
			payload := builder.BuildCreatePayloadWithoutDatabase()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			testResp := client.POST("/api/vega-backend/v1/catalogs/"+catalogID+"/test-connection", nil)
			So(testResp.StatusCode, ShouldEqual, http.StatusOK)
		})

		// ========== connector_config负向测试（MD121-MD129） ==========

		Convey("MD121: 缺少host字段", func() {
			mariadbConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           cataloghelpers.GenerateUniqueName("missing-host"),
				"connector_type": "mariadb",
				"connector_config": map[string]any{
					"port":      mariadbConfig.Port,
					"databases": []string{mariadbConfig.Database},
					"username":  mariadbConfig.Username,
					"password":  builder.GetEncryptedPassword(),
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MD122: 缺少port字段", func() {
			mariadbConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           cataloghelpers.GenerateUniqueName("missing-port"),
				"connector_type": "mariadb",
				"connector_config": map[string]any{
					"host":      mariadbConfig.Host,
					"databases": []string{mariadbConfig.Database},
					"username":  mariadbConfig.Username,
					"password":  builder.GetEncryptedPassword(),
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MD123: 缺少user字段", func() {
			mariadbConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           cataloghelpers.GenerateUniqueName("missing-user"),
				"connector_type": "mariadb",
				"connector_config": map[string]any{
					"host":      mariadbConfig.Host,
					"port":      mariadbConfig.Port,
					"databases": []string{mariadbConfig.Database},
					"password":  builder.GetEncryptedPassword(),
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MD124: 空用户名", func() {
			mariadbConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           cataloghelpers.GenerateUniqueName("empty-user"),
				"connector_type": "mariadb",
				"connector_config": map[string]any{
					"host":      mariadbConfig.Host,
					"port":      mariadbConfig.Port,
					"databases": []string{mariadbConfig.Database},
					"username":  "",
					"password":  mariadbConfig.Password,
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MD125: 错误密码", func() {
			payload := builder.BuildCreatePayloadWithWrongCredentials()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MD126: 不存在的数据库", func() {
			payload := builder.BuildCreatePayloadWithNonExistentDB()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MD127: 无效端口（非数字）", func() {
			payload := builder.BuildCreatePayloadWithInvalidPort()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MD128: 超出范围端口（65536）", func() {
			mariadbConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           cataloghelpers.GenerateUniqueName("overflow-port"),
				"connector_type": "mariadb",
				"connector_config": map[string]any{
					"host":      mariadbConfig.Host,
					"port":      65536,
					"databases": []string{mariadbConfig.Database},
					"username":  mariadbConfig.Username,
					"password":  mariadbConfig.Password,
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MD129: 负数端口", func() {
			mariadbConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           cataloghelpers.GenerateUniqueName("negative-port"),
				"connector_type": "mariadb",
				"connector_config": map[string]any{
					"host":      mariadbConfig.Host,
					"port":      -1,
					"databases": []string{mariadbConfig.Database},
					"username":  mariadbConfig.Username,
					"password":  mariadbConfig.Password,
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		// ========== 边界测试（MD131-MD138） ==========

		Convey("MD131: port边界值（1）", func() {
			mariadbConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           cataloghelpers.GenerateUniqueName("port-1"),
				"connector_type": "mariadb",
				"connector_config": map[string]any{
					"host":      mariadbConfig.Host,
					"port":      1,
					"databases": []string{mariadbConfig.Database},
					"username":  mariadbConfig.Username,
					"password":  builder.GetEncryptedPassword(),
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldBeIn, []int{http.StatusCreated, http.StatusBadRequest})
		})

		Convey("MD132: port边界值（65535）", func() {
			mariadbConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           cataloghelpers.GenerateUniqueName("port-65535"),
				"connector_type": "mariadb",
				"connector_config": map[string]any{
					"host":      mariadbConfig.Host,
					"port":      65535,
					"databases": []string{mariadbConfig.Database},
					"username":  mariadbConfig.Username,
					"password":  builder.GetEncryptedPassword(),
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldBeIn, []int{http.StatusCreated, http.StatusBadRequest})
		})

		Convey("MD133: database名称最大长度（64字符）", func() {
			mariadbConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           cataloghelpers.GenerateUniqueName("long-db"),
				"connector_type": "mariadb",
				"connector_config": map[string]any{
					"host":      mariadbConfig.Host,
					"port":      mariadbConfig.Port,
					"databases": []string{strings.Repeat("d", 64)},
					"username":  mariadbConfig.Username,
					"password":  builder.GetEncryptedPassword(),
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldBeIn, []int{http.StatusCreated, http.StatusBadRequest})
		})

		Convey("MD134: database名称超过最大长度", func() {
			mariadbConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           cataloghelpers.GenerateUniqueName("too-long-db"),
				"connector_type": "mariadb",
				"connector_config": map[string]any{
					"host":      mariadbConfig.Host,
					"port":      mariadbConfig.Port,
					"databases": []string{strings.Repeat("d", 65)},
					"username":  mariadbConfig.Username,
					"password":  builder.GetEncryptedPassword(),
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MD135: host为IP地址", func() {
			payload := builder.BuildCreatePayload()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("MD136: host为域名", func() {
			mariadbConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           cataloghelpers.GenerateUniqueName("domain-host"),
				"connector_type": "mariadb",
				"connector_config": map[string]any{
					"host":      "localhost",
					"port":      mariadbConfig.Port,
					"databases": []string{mariadbConfig.Database},
					"username":  mariadbConfig.Username,
					"password":  builder.GetEncryptedPassword(),
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldBeIn, []int{http.StatusCreated, http.StatusBadRequest})
		})

		Convey("MD137: 不指定database（实例级连接）", func() {
			payload := builder.BuildCreatePayloadWithoutDatabase()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("MD138: password为空（无密码连接）", func() {
			mariadbConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           cataloghelpers.GenerateUniqueName("no-password"),
				"connector_type": "mariadb",
				"connector_config": map[string]any{
					"host":      mariadbConfig.Host,
					"port":      mariadbConfig.Port,
					"databases": []string{mariadbConfig.Database},
					"username":  mariadbConfig.Username,
					"password":  "",
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldBeIn, []int{http.StatusCreated, http.StatusBadRequest})
		})
	})

	_ = ctx
}

// TestMariaDBCatalogRead MariaDB Catalog读取AT测试
// 编号规则：MD2xx
func TestMariaDBCatalogRead(t *testing.T) {
	var (
		ctx     context.Context
		config  *setup.TestConfig
		client  *testutil.HTTPClient
		builder *MariaDBPayloadBuilder
	)

	Convey("MariaDB Catalog读取AT测试 - 初始化", t, func() {
		ctx = context.Background()

		var err error
		config, err = setup.LoadTestConfig()
		So(err, ShouldBeNil)

		client = testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		builder = NewMariaDBPayloadBuilder(config.TargetMariaDB)
		builder.SetTestConfig(config)

		cataloghelpers.CleanupCatalogs(client, t)

		// ========== 读取测试（MD201-MD205） ==========

		Convey("MD201: 获取存在的MariaDB catalog", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			So(getResp.StatusCode, ShouldEqual, http.StatusOK)

			catalog := cataloghelpers.ExtractFromEntriesResponse(getResp)
			So(catalog["id"], ShouldEqual, catalogID)
		})

		Convey("MD202: 列表查询 - 按type过滤physical", func() {
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

		Convey("MD203: 列表查询 - 按connector_type过滤mariadb", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			listResp := client.GET("/api/vega-backend/v1/catalogs?connector_type=mariadb&offset=0&limit=100")
			So(listResp.StatusCode, ShouldEqual, http.StatusOK)

			if entries, ok := listResp.Body["entries"].([]any); ok {
				So(len(entries), ShouldBeGreaterThanOrEqualTo, 1)
				for _, entry := range entries {
					So(entry.(map[string]any)["connector_type"], ShouldEqual, "mariadb")
				}
			}
		})

		Convey("MD204: 查询catalog - 验证所有字段返回", func() {
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
			So(catalog["connector_type"], ShouldEqual, "mariadb")
			So(catalog["create_time"], ShouldNotBeZeroValue)
			So(catalog["update_time"], ShouldNotBeZeroValue)
		})

		Convey("MD205: 验证connector_config.password不返回", func() {
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

// TestMariaDBCatalogUpdate MariaDB Catalog更新AT测试
// 编号规则：MD3xx
func TestMariaDBCatalogUpdate(t *testing.T) {
	var (
		ctx     context.Context
		config  *setup.TestConfig
		client  *testutil.HTTPClient
		builder *MariaDBPayloadBuilder
	)

	Convey("MariaDB Catalog更新AT测试 - 初始化", t, func() {
		ctx = context.Background()

		var err error
		config, err = setup.LoadTestConfig()
		So(err, ShouldBeNil)

		client = testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		builder = NewMariaDBPayloadBuilder(config.TargetMariaDB)
		builder.SetTestConfig(config)

		cataloghelpers.CleanupCatalogs(client, t)

		// ========== 更新测试（MD301-MD305） ==========

		Convey("MD301: 整体更新connector_config", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			originalData := cataloghelpers.ExtractFromEntriesResponse(getResp)

			mariadbConfig := builder.GetConfig()
			updatePayload := cataloghelpers.BuildUpdatePayload(originalData, map[string]any{
				"connector_config": map[string]any{
					"host":      mariadbConfig.Host,
					"port":      mariadbConfig.Port,
					"databases": []string{mariadbConfig.Database},
					"username":  mariadbConfig.Username,
					"password":  builder.GetEncryptedPassword(),
					"options": map[string]any{
						"charset": "utf8mb4",
					},
				},
			})
			updateResp := client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("MD302: 更新connector_config后连接测试", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			originalData := cataloghelpers.ExtractFromEntriesResponse(getResp)

			mariadbConfig := builder.GetConfig()
			updatePayload := cataloghelpers.BuildUpdatePayload(originalData, map[string]any{
				"connector_config": map[string]any{
					"host":      mariadbConfig.Host,
					"port":      mariadbConfig.Port,
					"databases": []string{mariadbConfig.Database},
					"username":  mariadbConfig.Username,
					"password":  builder.GetEncryptedPassword(),
				},
			})
			updateResp := client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)

			testResp := client.POST("/api/vega-backend/v1/catalogs/"+catalogID+"/test-connection", nil)
			So(testResp.StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("MD303: 更新host为无效地址", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			originalData := cataloghelpers.ExtractFromEntriesResponse(getResp)

			mariadbConfig := builder.GetConfig()
			updatePayload := cataloghelpers.BuildUpdatePayload(originalData, map[string]any{
				"connector_config": map[string]any{
					"host":      "invalid-host-12345.example.com",
					"port":      mariadbConfig.Port,
					"databases": []string{mariadbConfig.Database},
					"username":  mariadbConfig.Username,
					"password":  builder.GetEncryptedPassword(),
				},
			})
			updateResp := client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MD304: 更新port为无效值", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			originalData := cataloghelpers.ExtractFromEntriesResponse(getResp)

			mariadbConfig := builder.GetConfig()
			updatePayload := cataloghelpers.BuildUpdatePayload(originalData, map[string]any{
				"connector_config": map[string]any{
					"host":      mariadbConfig.Host,
					"port":      65536,
					"databases": []string{mariadbConfig.Database},
					"username":  mariadbConfig.Username,
					"password":  builder.GetEncryptedPassword(),
				},
			})
			updateResp := client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MD305: 更新password", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			originalData := cataloghelpers.ExtractFromEntriesResponse(getResp)

			mariadbConfig := builder.GetConfig()
			updatePayload := cataloghelpers.BuildUpdatePayload(originalData, map[string]any{
				"connector_config": map[string]any{
					"host":      mariadbConfig.Host,
					"port":      mariadbConfig.Port,
					"databases": []string{mariadbConfig.Database},
					"username":  mariadbConfig.Username,
					"password":  builder.GetEncryptedPassword(),
				},
			})
			updateResp := client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)
		})
	})

	_ = ctx
}

// TestMariaDBCatalogDelete MariaDB Catalog删除AT测试
// 编号规则：MD4xx
func TestMariaDBCatalogDelete(t *testing.T) {
	var (
		ctx     context.Context
		config  *setup.TestConfig
		client  *testutil.HTTPClient
		builder *MariaDBPayloadBuilder
	)

	Convey("MariaDB Catalog删除AT测试 - 初始化", t, func() {
		ctx = context.Background()

		var err error
		config, err = setup.LoadTestConfig()
		So(err, ShouldBeNil)

		client = testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		builder = NewMariaDBPayloadBuilder(config.TargetMariaDB)
		builder.SetTestConfig(config)

		cataloghelpers.CleanupCatalogs(client, t)

		// ========== 删除测试（MD401-MD402） ==========

		Convey("MD401: 删除MariaDB catalog后健康状态不可查", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			deleteResp := client.DELETE("/api/vega-backend/v1/catalogs/" + catalogID)
			So(deleteResp.StatusCode, ShouldEqual, http.StatusNoContent)

			statusResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID + "/health-status")
			So(statusResp.StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("MD402: 删除MariaDB catalog后不能测试连接", func() {
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

// TestMariaDBSpecificOptions MariaDB特有选项测试
// 编号规则：MD5xx
func TestMariaDBSpecificOptions(t *testing.T) {
	var (
		ctx     context.Context
		config  *setup.TestConfig
		client  *testutil.HTTPClient
		builder *MariaDBPayloadBuilder
	)

	Convey("MariaDB特有选项AT测试 - 初始化", t, func() {
		ctx = context.Background()

		var err error
		config, err = setup.LoadTestConfig()
		So(err, ShouldBeNil)

		client = testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		builder = NewMariaDBPayloadBuilder(config.TargetMariaDB)
		builder.SetTestConfig(config)

		cataloghelpers.CleanupCatalogs(client, t)

		// ========== MariaDB特有选项测试（MD501-MD506） ==========

		Convey("MD501: MariaDB charset选项测试（utf8mb4）", func() {
			options := map[string]any{"charset": "utf8mb4"}
			payload := builder.BuildCreatePayloadWithOptions(options)
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("MD502: MariaDB parseTime选项测试", func() {
			options := map[string]any{"parseTime": "true"}
			payload := builder.BuildCreatePayloadWithOptions(options)
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("MD503: MariaDB loc选项测试（时区）", func() {
			options := map[string]any{"loc": "Local"}
			payload := builder.BuildCreatePayloadWithOptions(options)
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("MD504: MariaDB timeout选项测试", func() {
			options := map[string]any{"timeout": "10s"}
			payload := builder.BuildCreatePayloadWithOptions(options)
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("MD505: MariaDB SSL连接测试", func() {
			options := map[string]any{"tls": "skip-verify"}
			payload := builder.BuildCreatePayloadWithOptions(options)
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldBeIn, []int{http.StatusCreated, http.StatusBadRequest})
		})

		Convey("MD506: MariaDB collation选项测试", func() {
			options := map[string]any{"collation": "utf8mb4_unicode_ci"}
			payload := builder.BuildCreatePayloadWithOptions(options)
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})
	})

	_ = ctx
}
