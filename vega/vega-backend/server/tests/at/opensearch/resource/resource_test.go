// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package resource

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	commonresource "vega-backend/tests/at/common/resource"
	"vega-backend/tests/at/setup"
	"vega-backend/tests/testutil"
)

// TestOpenSearchResourceCreate OpenSearch Resource创建AT测试
// 使用通用测试套件运行所有标准测试
func TestOpenSearchResourceCreate(t *testing.T) {
	var (
		ctx    context.Context
		config *setup.TestConfig
		client *testutil.HTTPClient
		suite  *commonresource.TestSuite
	)

	Convey("OpenSearch Resource创建AT测试 - 初始化", t, func() {
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
		suite, err = commonresource.NewTestSuite(t, "opensearch")
		So(err, ShouldBeNil)

		// 初始化测试环境（清理 + 创建前置Catalog）
		err = suite.Setup()
		So(err, ShouldBeNil)

		// ========== 运行通用测试 ==========

		Convey("通用创建测试（RM101-RM109）", func() {
			commonresource.RunCommonCreateTests(suite)
		})

		Convey("通用负向测试（RM121-RM134）", func() {
			commonresource.RunCommonNegativeTests(suite)
		})

		Convey("通用边界测试（RM141-RM149）", func() {
			commonresource.RunCommonBoundaryTests(suite)
		})

		Convey("通用安全测试（RM161-RM163）", func() {
			commonresource.RunCommonSecurityTests(suite)
		})
	})

	_ = ctx
	_ = client
}

// TestOpenSearchResourceRead OpenSearch Resource读取AT测试
func TestOpenSearchResourceRead(t *testing.T) {
	var (
		ctx   context.Context
		suite *commonresource.TestSuite
	)

	Convey("OpenSearch Resource读取AT测试 - 初始化", t, func() {
		ctx = context.Background()

		// 创建测试套件
		var err error
		suite, err = commonresource.NewTestSuite(t, "opensearch")
		So(err, ShouldBeNil)

		err = suite.Setup()
		So(err, ShouldBeNil)

		// 运行通用读取测试
		Convey("通用读取测试（RM201-RM206）", func() {
			commonresource.RunCommonReadTests(suite)
		})
	})

	_ = ctx
}

// TestOpenSearchResourceUpdate OpenSearch Resource更新AT测试
func TestOpenSearchResourceUpdate(t *testing.T) {
	var (
		ctx   context.Context
		suite *commonresource.TestSuite
	)

	Convey("OpenSearch Resource更新AT测试 - 初始化", t, func() {
		ctx = context.Background()

		// 创建测试套件
		var err error
		suite, err = commonresource.NewTestSuite(t, "opensearch")
		So(err, ShouldBeNil)

		err = suite.Setup()
		So(err, ShouldBeNil)

		// 运行通用更新测试
		Convey("通用更新测试（RM301-RM305）", func() {
			commonresource.RunCommonUpdateTests(suite)
		})
	})

	_ = ctx
}

// TestOpenSearchResourceDelete OpenSearch Resource删除AT测试
func TestOpenSearchResourceDelete(t *testing.T) {
	var (
		ctx   context.Context
		suite *commonresource.TestSuite
	)

	Convey("OpenSearch Resource删除AT测试 - 初始化", t, func() {
		ctx = context.Background()

		// 创建测试套件
		var err error
		suite, err = commonresource.NewTestSuite(t, "opensearch")
		So(err, ShouldBeNil)

		err = suite.Setup()
		So(err, ShouldBeNil)

		// 运行通用删除测试
		Convey("通用删除测试（RM401-RM406）", func() {
			commonresource.RunCommonDeleteTests(suite)
		})
	})

	_ = ctx
}

// TestOpenSearchResourceNameUniqueness OpenSearch Resource名称唯一性AT测试
func TestOpenSearchResourceNameUniqueness(t *testing.T) {
	var (
		ctx   context.Context
		suite *commonresource.TestSuite
	)

	Convey("OpenSearch Resource名称唯一性AT测试 - 初始化", t, func() {
		ctx = context.Background()

		// 创建测试套件
		var err error
		suite, err = commonresource.NewTestSuite(t, "opensearch")
		So(err, ShouldBeNil)

		err = suite.Setup()
		So(err, ShouldBeNil)

		// 运行名称唯一性测试
		Convey("名称唯一性测试（RM501-RM503）", func() {
			commonresource.RunCommonNameUniquenessTests(suite)
		})
	})

	_ = ctx
}
