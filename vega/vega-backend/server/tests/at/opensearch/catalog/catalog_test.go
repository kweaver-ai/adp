// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package catalog

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	commoncatalog "vega-backend/tests/at/common/catalog"
	"vega-backend/tests/at/fixtures"
	"vega-backend/tests/at/setup"
	"vega-backend/tests/testutil"
)

// TestOpenSearchCatalogCreate OpenSearch Catalog创建AT测试
// 使用通用测试套件运行所有标准测试
func TestOpenSearchCatalogCreate(t *testing.T) {
	var (
		ctx    context.Context
		config *setup.TestConfig
		client *testutil.HTTPClient
		suite  *commoncatalog.TestSuite
	)

	Convey("OpenSearch Catalog创建AT测试 - 初始化", t, func() {
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

		// 创建测试套件
		suite, err = commoncatalog.NewTestSuite(t, "opensearch")
		So(err, ShouldBeNil)

		// 清理现有catalog
		fixtures.CleanupCatalogs(client, t)

		// ========== 运行通用测试 ==========

		Convey("通用创建测试（CM101-CM111）", func() {
			commoncatalog.RunCommonCreateTests(suite)
		})

		Convey("通用负向测试（CM121-CM136）", func() {
			commoncatalog.RunCommonNegativeTests(suite)
		})

		Convey("通用边界测试（CM141-CM149）", func() {
			commoncatalog.RunCommonBoundaryTests(suite)
		})

		Convey("通用安全测试（CM161-CM163）", func() {
			commoncatalog.RunCommonSecurityTests(suite)
		})
	})

	_ = ctx
}

// TestOpenSearchCatalogRead OpenSearch Catalog读取AT测试
func TestOpenSearchCatalogRead(t *testing.T) {
	var (
		ctx   context.Context
		suite *commoncatalog.TestSuite
	)

	Convey("OpenSearch Catalog读取AT测试 - 初始化", t, func() {
		ctx = context.Background()

		// 创建测试套件
		var err error
		suite, err = commoncatalog.NewTestSuite(t, "opensearch")
		So(err, ShouldBeNil)

		err = suite.Setup()
		So(err, ShouldBeNil)

		// 运行通用读取测试
		Convey("通用读取测试（CM201-CM203）", func() {
			commoncatalog.RunCommonReadTests(suite)
		})
	})

	_ = ctx
}

// TestOpenSearchCatalogUpdate OpenSearch Catalog更新AT测试
func TestOpenSearchCatalogUpdate(t *testing.T) {
	var (
		ctx   context.Context
		suite *commoncatalog.TestSuite
	)

	Convey("OpenSearch Catalog更新AT测试 - 初始化", t, func() {
		ctx = context.Background()

		// 创建测试套件
		var err error
		suite, err = commoncatalog.NewTestSuite(t, "opensearch")
		So(err, ShouldBeNil)

		err = suite.Setup()
		So(err, ShouldBeNil)

		// 运行通用更新测试
		Convey("通用更新测试（CM301-CM304）", func() {
			commoncatalog.RunCommonUpdateTests(suite)
		})
	})

	_ = ctx
}

// TestOpenSearchCatalogDelete OpenSearch Catalog删除AT测试
func TestOpenSearchCatalogDelete(t *testing.T) {
	var (
		ctx   context.Context
		suite *commoncatalog.TestSuite
	)

	Convey("OpenSearch Catalog删除AT测试 - 初始化", t, func() {
		ctx = context.Background()

		// 创建测试套件
		var err error
		suite, err = commoncatalog.NewTestSuite(t, "opensearch")
		So(err, ShouldBeNil)

		err = suite.Setup()
		So(err, ShouldBeNil)

		// 运行通用删除测试
		Convey("通用删除测试（CM401-CM403）", func() {
			commoncatalog.RunCommonDeleteTests(suite)
		})
	})

	_ = ctx
}
