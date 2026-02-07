// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package catalog

import (
	"fmt"
	"net/http"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"vega-backend/tests/at/fixtures"
	catalogfixtures "vega-backend/tests/at/fixtures/catalog"
)

// RunCommonCreateTests 运行通用创建测试
// 测试编号前缀: CM1xx (Common Create)
func RunCommonCreateTests(suite *TestSuite) {
	connectorType := suite.GetConnectorType()

	// ========== 正向测试（CM101-CM120） ==========

	Convey(fmt.Sprintf("CM101: 创建%s physical catalog - 基本场景", connectorType), func() {
		payload := suite.Builder.BuildCreatePayload()
		resp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		So(resp.Body["id"], ShouldNotBeEmpty)
	})

	Convey("CM102: 创建catalog - 最小字段（仅name）", func() {
		payload := catalogfixtures.BuildMinimalPayload()
		resp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		So(resp.Body["id"], ShouldNotBeEmpty)
	})

	Convey(fmt.Sprintf("CM103: 创建%s catalog - 完整字段", connectorType), func() {
		payload := suite.Builder.BuildFullCreatePayload()
		resp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		So(resp.Body["id"], ShouldNotBeEmpty)
	})

	Convey("CM104: 创建后立即查询", func() {
		payload := suite.Builder.BuildCreatePayload()
		createResp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
		So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

		catalogID := createResp.Body["id"].(string)

		// 立即查询
		getResp := suite.Client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
		So(getResp.StatusCode, ShouldEqual, http.StatusOK)
		catalog := fixtures.ExtractFromEntriesResponse(getResp)
		So(catalog["id"], ShouldEqual, catalogID)
		So(catalog["name"], ShouldEqual, payload["name"])
	})

	Convey(fmt.Sprintf("CM105: 创建带options的%s catalog", connectorType), func() {
		options := map[string]any{
			"timeout":     "10s",
			"max_retries": 3,
		}
		payload := suite.Builder.BuildCreatePayloadWithOptions(options)
		resp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		So(resp.Body["id"], ShouldNotBeEmpty)
	})

	Convey("CM106: 创建logical类型catalog", func() {
		payload := catalogfixtures.BuildLogicalCatalogPayload()
		resp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		So(resp.Body["id"], ShouldNotBeEmpty)
	})

	// 仅在支持连接测试的connector类型上运行
	if suite.Builder.SupportsTestConnection() {
		Convey("CM107: 创建后测试连接成功", func() {
			payload := suite.Builder.BuildCreatePayload()
			createResp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
			So(createResp.Body["id"], ShouldNotBeEmpty)

			catalogID := createResp.Body["id"].(string)

			// 测试连接
			testResp := suite.Client.POST("/api/vega-backend/v1/catalogs/"+catalogID+"/test-connection", nil)
			So(testResp.StatusCode, ShouldEqual, http.StatusOK)

			if testResp.Body != nil {
				_, hasStatus := testResp.Body["health_check_status"]
				So(hasStatus, ShouldBeTrue)
			}
		})

		Convey("CM108: 获取catalog健康状态", func() {
			payload := suite.Builder.BuildCreatePayload()
			createResp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			// 获取状态
			statusResp := suite.Client.GET("/api/vega-backend/v1/catalogs/" + catalogID + "/health-status")
			So(statusResp.StatusCode, ShouldEqual, http.StatusOK)

			if statusResp.Body != nil {
				_, hasStatus := statusResp.Body["health_check_status"]
				So(hasStatus, ShouldBeTrue)
			}
		})
	}

	Convey("CM109: 创建多个catalog，列表查询", func() {
		// 创建3个catalog
		for i := 0; i < 3; i++ {
			payload := suite.Builder.BuildCreatePayload()
			resp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		}

		// 列表查询
		listResp := suite.Client.GET("/api/vega-backend/v1/catalogs?offset=0&limit=10")
		So(listResp.StatusCode, ShouldEqual, http.StatusOK)

		if listResp.Body != nil {
			if entries, ok := listResp.Body["entries"].([]any); ok {
				So(len(entries), ShouldBeGreaterThanOrEqualTo, 3)
			}

			if totalCount, ok := listResp.Body["total_count"].(float64); ok {
				So(totalCount, ShouldBeGreaterThanOrEqualTo, 3.0)
			}
		}
	})

	Convey("CM110: Tags数组测试", func() {
		Convey("空tags数组", func() {
			payload := catalogfixtures.BuildPayloadWithEmptyTags()
			resp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
			So(resp.Body["id"], ShouldNotBeEmpty)
		})

		Convey("单个tag", func() {
			payload := suite.Builder.BuildCreatePayload()
			payload["tags"] = []string{"single-tag"}
			resp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
			So(resp.Body["id"], ShouldNotBeEmpty)
		})

		Convey("多个tags", func() {
			payload := suite.Builder.BuildCreatePayload()
			payload["tags"] = []string{"tag1", "tag2", "tag3", "tag4", "tag5"}
			resp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
			So(resp.Body["id"], ShouldNotBeEmpty)
		})
	})

	Convey("CM111: 特殊字符名称测试", func() {
		testCases := []struct {
			name        string
			catalogName string
			expectCode  int
		}{
			{"中文名称", "测试目录", 201},
			{"连字符", "test-catalog", 201},
			{"下划线", "test_catalog", 201},
			{"点号", "test.catalog", 201},
			{"混合", "测试-catalog_01.test", 201},
		}

		for _, tc := range testCases {
			Convey(tc.name, func() {
				payload := catalogfixtures.BuildPayloadWithSpecialCharsName(tc.catalogName)
				resp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
				So(resp.StatusCode, ShouldEqual, tc.expectCode)
			})
		}
	})
}

// RunCommonNegativeTests 运行通用负向测试
// 测试编号前缀: CM1xx (121-140)
func RunCommonNegativeTests(suite *TestSuite) {
	Convey("CM121: 缺少必填字段 - name", func() {
		payload := catalogfixtures.BuildPayloadWithMissingName()
		resp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
	})

	Convey("CM122: 重复的catalog名称", func() {
		fixedName := fmt.Sprintf("duplicate-test-%d", time.Now().Unix())
		payload1 := suite.Builder.BuildCreatePayload()
		payload1["name"] = fixedName

		// 第一次创建
		resp1 := suite.Client.POST("/api/vega-backend/v1/catalogs", payload1)
		So(resp1.StatusCode, ShouldEqual, http.StatusCreated)

		// 第二次创建相同名称
		payload2 := suite.Builder.BuildCreatePayload()
		payload2["name"] = fixedName

		resp2 := suite.Client.POST("/api/vega-backend/v1/catalogs", payload2)
		So(resp2.StatusCode, ShouldEqual, http.StatusConflict)
		So(resp2.Error, ShouldNotBeNil)
	})

	Convey("CM123: 无效JSON格式", func() {
		invalidJSON := `{"name": "test", "type": }`
		resp := suite.Client.POST("/api/vega-backend/v1/catalogs", invalidJSON)
		So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
	})

	Convey("CM124: 错误的Content-Type", func() {
		payload := suite.Builder.BuildCreatePayload()
		suite.Client.SetHeader("Content-Type", "text/plain")
		resp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
		suite.Client.SetHeader("Content-Type", "application/json")
		So(resp.StatusCode, ShouldEqual, http.StatusNotAcceptable)
	})

	Convey("CM125: 超长name字段（>128字符）", func() {
		payload := catalogfixtures.BuildPayloadWithLongName(129)
		resp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
	})

	Convey("CM126: 超长description字段（>1000字符）", func() {
		payload := catalogfixtures.BuildPayloadWithLongDescription(1001)
		resp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
	})

	// 仅在支持连接测试的connector类型上运行凭证测试
	if suite.Builder.SupportsTestConnection() {
		Convey("CM127: 错误的连接凭证", func() {
			payload := suite.Builder.BuildCreatePayloadWithWrongCredentials()
			resp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("CM128: 无效的连接配置", func() {
			payload := suite.Builder.BuildCreatePayloadWithInvalidConfig()
			resp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})
	}

	Convey("CM129: name为空字符串", func() {
		payload := catalogfixtures.BuildPayloadWithEmptyName()
		resp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
	})

	Convey("CM130: name只有空格", func() {
		payload := catalogfixtures.BuildPayloadWithWhitespaceName()
		resp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
		// name只有长度限制，空格也是有效字符
		So(resp.StatusCode, ShouldEqual, http.StatusCreated)
	})

	Convey("CM131: tags包含空字符串", func() {
		payload := catalogfixtures.BuildPayloadWithEmptyTagInArray()
		resp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
	})

	Convey("CM132: tags包含非法字符", func() {
		payload := catalogfixtures.BuildPayloadWithInvalidCharTag()
		resp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
	})

	Convey("CM133: 单个tag超长（41字符）", func() {
		payload := catalogfixtures.BuildPayloadWithLongTag(41)
		resp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
	})

	Convey("CM134: 无效的connector_type", func() {
		payload := catalogfixtures.BuildPayloadWithInvalidConnectorType()
		resp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
	})

	Convey("CM135: connector_config缺少必要字段", func() {
		payload := catalogfixtures.BuildPayloadWithMissingConnectorFields()
		resp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
	})

	Convey("CM136: 请求体为空", func() {
		resp := suite.Client.POST("/api/vega-backend/v1/catalogs", nil)
		So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
	})
}

// RunCommonBoundaryTests 运行通用边界测试
// 测试编号前缀: CM1xx (141-160)
func RunCommonBoundaryTests(suite *TestSuite) {
	Convey("CM141: name长度边界 - 1字符（最小）", func() {
		payload := catalogfixtures.BuildPayloadWithMinName()
		resp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		So(resp.Body["id"], ShouldNotBeEmpty)
	})

	Convey("CM142: name长度边界 - 128字符（最大允许）", func() {
		payload := catalogfixtures.BuildPayloadWithLongName(128)
		resp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		So(resp.Body["id"], ShouldNotBeEmpty)
	})

	Convey("CM143: name长度边界 - 129字符（超出）", func() {
		payload := catalogfixtures.BuildPayloadWithLongName(129)
		resp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
	})

	Convey("CM144: description长度边界 - 1000字符（最大允许）", func() {
		payload := catalogfixtures.BuildPayloadWithExactDescription(1000)
		resp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		So(resp.Body["id"], ShouldNotBeEmpty)
	})

	Convey("CM145: description长度边界 - 1001字符（超出）", func() {
		payload := catalogfixtures.BuildPayloadWithExactDescription(1001)
		resp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
	})

	Convey("CM146: tags数量边界 - 5个（最大允许）", func() {
		payload := catalogfixtures.BuildPayloadWithManyTags(5)
		resp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		So(resp.Body["id"], ShouldNotBeEmpty)
	})

	Convey("CM147: tags数量边界 - 6个（超出）", func() {
		payload := catalogfixtures.BuildPayloadWithManyTags(6)
		resp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
	})

	Convey("CM148: 单个tag长度边界 - 40字符（最大允许）", func() {
		payload := catalogfixtures.BuildPayloadWithLongTag(40)
		resp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		So(resp.Body["id"], ShouldNotBeEmpty)
	})

	Convey("CM149: 单个tag长度边界 - 41字符（超出）", func() {
		payload := catalogfixtures.BuildPayloadWithLongTag(41)
		resp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
	})
}

// RunCommonSecurityTests 运行通用安全测试
// 测试编号前缀: CM1xx (161-170)
func RunCommonSecurityTests(suite *TestSuite) {
	// 注：name字段只有长度限制，没有内容限制，因此特殊字符会被接受
	// 这些测试验证服务端能安全处理这些特殊输入（不会导致SQL注入、XSS等）

	Convey("CM161: SQL注入尝试", func() {
		payload := catalogfixtures.BuildPayloadWithSQLInjection()
		resp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
		// name没有内容限制，创建成功；验证系统能安全处理SQL注入payload
		So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		So(resp.Body["id"], ShouldNotBeEmpty)
	})

	Convey("CM162: XSS尝试", func() {
		payload := catalogfixtures.BuildPayloadWithXSS()
		resp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
		// name没有内容限制，创建成功；验证系统能安全处理XSS payload
		So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		So(resp.Body["id"], ShouldNotBeEmpty)
	})

	Convey("CM163: 路径遍历尝试", func() {
		payload := catalogfixtures.BuildPayloadWithPathTraversal()
		resp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
		// name没有内容限制，创建成功；验证系统能安全处理路径遍历payload
		So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		So(resp.Body["id"], ShouldNotBeEmpty)
	})
}

// RunCommonReadTests 运行通用读取测试
// 测试编号前缀: CM2xx
func RunCommonReadTests(suite *TestSuite) {
	Convey("CM201: 获取存在的catalog", func() {
		// 先创建
		payload := suite.Builder.BuildCreatePayload()
		createResp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
		So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
		catalogID := createResp.Body["id"].(string)

		// 查询
		getResp := suite.Client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
		So(getResp.StatusCode, ShouldEqual, http.StatusOK)

		catalog := fixtures.ExtractFromEntriesResponse(getResp)
		So(catalog, ShouldNotBeNil)
		So(catalog["id"], ShouldEqual, catalogID)
	})

	Convey("CM202: 获取不存在的catalog", func() {
		resp := suite.Client.GET("/api/vega-backend/v1/catalogs/non-existent-id-12345")
		So(resp.StatusCode, ShouldEqual, http.StatusNotFound)
	})

	Convey("CM203: 列表分页测试", func() {
		// 创建5个catalog
		for i := 0; i < 5; i++ {
			payload := suite.Builder.BuildCreatePayload()
			resp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		}

		// 分页查询
		listResp := suite.Client.GET("/api/vega-backend/v1/catalogs?offset=0&limit=2")
		So(listResp.StatusCode, ShouldEqual, http.StatusOK)

		if entries, ok := listResp.Body["entries"].([]any); ok {
			So(len(entries), ShouldEqual, 2)
		}
	})
}

// RunCommonUpdateTests 运行通用更新测试（修正版：基于原数据更新）
// 测试编号前缀: CM3xx
func RunCommonUpdateTests(suite *TestSuite) {
	Convey("CM301: 更新catalog名称", func() {
		// 创建
		payload := suite.Builder.BuildCreatePayload()
		catalogID, catalogData, err := suite.CreateAndGetCatalog(payload)
		So(err, ShouldBeNil)

		// 基于原数据构建更新payload
		newName := catalogfixtures.GenerateUniqueName("updated-catalog")
		updatePayload := catalogfixtures.BuildUpdatePayload(catalogData, map[string]any{
			"name": newName,
		})
		// 回填加密密码（GET不再返回敏感字段）
		catalogfixtures.InjectEncryptedPassword(updatePayload, suite.Builder.GetEncryptedPassword())
		updateResp := suite.Client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, updatePayload)
		So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)

		// 验证
		getResp := suite.Client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
		catalog := fixtures.ExtractFromEntriesResponse(getResp)
		So(catalog["name"], ShouldEqual, newName)
	})

	Convey("CM302: 更新catalog注释", func() {
		// 创建
		payload := suite.Builder.BuildCreatePayload()
		catalogID, catalogData, err := suite.CreateAndGetCatalog(payload)
		So(err, ShouldBeNil)

		// 基于原数据构建更新payload
		newDescription := "更新后的注释内容"
		updatePayload := catalogfixtures.BuildUpdatePayload(catalogData, map[string]any{
			"description": newDescription,
		})
		// 回填加密密码（GET不再返回敏感字段）
		catalogfixtures.InjectEncryptedPassword(updatePayload, suite.Builder.GetEncryptedPassword())
		updateResp := suite.Client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, updatePayload)
		So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)

		// 验证
		getResp := suite.Client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
		catalog := fixtures.ExtractFromEntriesResponse(getResp)
		So(catalog["description"], ShouldEqual, newDescription)
	})

	Convey("CM303: 更新不存在的catalog", func() {
		// 构建一个基本payload
		updatePayload := map[string]any{
			"name": "new-name",
		}
		resp := suite.Client.PUT("/api/vega-backend/v1/catalogs/non-existent-id-12345", updatePayload)
		So(resp.StatusCode, ShouldEqual, http.StatusNotFound)
	})

	Convey("CM304: 更新tags", func() {
		// 创建
		payload := suite.Builder.BuildCreatePayload()
		catalogID, catalogData, err := suite.CreateAndGetCatalog(payload)
		So(err, ShouldBeNil)

		// 基于原数据构建更新payload
		updatePayload := catalogfixtures.BuildUpdatePayload(catalogData, map[string]any{
			"tags": []string{"updated", "test"},
		})
		// 回填加密密码（GET不再返回敏感字段）
		catalogfixtures.InjectEncryptedPassword(updatePayload, suite.Builder.GetEncryptedPassword())
		updateResp := suite.Client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, updatePayload)
		So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)
	})
}

// RunCommonDeleteTests 运行通用删除测试
// 测试编号前缀: CM4xx
func RunCommonDeleteTests(suite *TestSuite) {
	Convey("CM401: 删除存在的catalog", func() {
		// 创建
		payload := suite.Builder.BuildCreatePayload()
		createResp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
		So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
		catalogID := createResp.Body["id"].(string)

		// 删除
		deleteResp := suite.Client.DELETE("/api/vega-backend/v1/catalogs/" + catalogID)
		So(deleteResp.StatusCode, ShouldEqual, http.StatusNoContent)

		// 验证已删除
		getResp := suite.Client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
		So(getResp.StatusCode, ShouldEqual, http.StatusNotFound)
	})

	Convey("CM402: 删除不存在的catalog", func() {
		resp := suite.Client.DELETE("/api/vega-backend/v1/catalogs/non-existent-id-12345")
		So(resp.StatusCode, ShouldEqual, http.StatusNotFound)
	})

	Convey("CM403: 重复删除同一catalog", func() {
		// 创建
		payload := suite.Builder.BuildCreatePayload()
		createResp := suite.Client.POST("/api/vega-backend/v1/catalogs", payload)
		So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
		catalogID := createResp.Body["id"].(string)

		// 第一次删除
		deleteResp1 := suite.Client.DELETE("/api/vega-backend/v1/catalogs/" + catalogID)
		So(deleteResp1.StatusCode, ShouldEqual, http.StatusNoContent)

		// 第二次删除
		deleteResp2 := suite.Client.DELETE("/api/vega-backend/v1/catalogs/" + catalogID)
		So(deleteResp2.StatusCode, ShouldEqual, http.StatusNotFound)
	})
}

// RunAllCommonTests 运行所有通用测试
func RunAllCommonTests(suite *TestSuite) {
	RunCommonCreateTests(suite)
	RunCommonNegativeTests(suite)
	RunCommonBoundaryTests(suite)
	RunCommonSecurityTests(suite)
	RunCommonReadTests(suite)
	RunCommonUpdateTests(suite)
	RunCommonDeleteTests(suite)
}
