package resource

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	commonresource "vega-backend/tests/at/common/resource"
	"vega-backend/tests/at/setup"
	"vega-backend/tests/testutil"
)

// TestMySQLResourceCreate MySQL Resource创建AT测试
// 使用通用测试套件运行所有标准测试
func TestMySQLResourceCreate(t *testing.T) {
	var (
		ctx    context.Context
		config *setup.TestConfig
		client *testutil.HTTPClient
		suite  *commonresource.TestSuite
	)

	Convey("MySQL Resource创建AT测试 - 初始化", t, func() {
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

		// 创建测试套件
		suite, err = commonresource.NewTestSuite(t, "mysql")
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

// TestMySQLResourceRead MySQL Resource读取AT测试
func TestMySQLResourceRead(t *testing.T) {
	var (
		ctx   context.Context
		suite *commonresource.TestSuite
	)

	Convey("MySQL Resource读取AT测试 - 初始化", t, func() {
		ctx = context.Background()

		// 创建测试套件
		var err error
		suite, err = commonresource.NewTestSuite(t, "mysql")
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

// TestMySQLResourceUpdate MySQL Resource更新AT测试
func TestMySQLResourceUpdate(t *testing.T) {
	var (
		ctx   context.Context
		suite *commonresource.TestSuite
	)

	Convey("MySQL Resource更新AT测试 - 初始化", t, func() {
		ctx = context.Background()

		// 创建测试套件
		var err error
		suite, err = commonresource.NewTestSuite(t, "mysql")
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

// TestMySQLResourceDelete MySQL Resource删除AT测试
func TestMySQLResourceDelete(t *testing.T) {
	var (
		ctx   context.Context
		suite *commonresource.TestSuite
	)

	Convey("MySQL Resource删除AT测试 - 初始化", t, func() {
		ctx = context.Background()

		// 创建测试套件
		var err error
		suite, err = commonresource.NewTestSuite(t, "mysql")
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

// TestMySQLResourceNameUniqueness MySQL Resource名称唯一性AT测试
func TestMySQLResourceNameUniqueness(t *testing.T) {
	var (
		ctx   context.Context
		suite *commonresource.TestSuite
	)

	Convey("MySQL Resource名称唯一性AT测试 - 初始化", t, func() {
		ctx = context.Background()

		// 创建测试套件
		var err error
		suite, err = commonresource.NewTestSuite(t, "mysql")
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
