// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package mysql

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"vega-backend-tests/at/resource/internal"
)

// TestMySQLResourceCommon MySQL Resource通用AT测试
// 使用resource/internal包中的通用测试用例
func TestMySQLResourceCommon(t *testing.T) {
	Convey("MySQL Resource通用AT测试 - 初始化", t, func() {
		// 创建测试套件
		suite, err := internal.NewTestSuite(t, "mysql")
		So(err, ShouldBeNil)
		So(suite, ShouldNotBeNil)

		// 初始化测试环境
		err = suite.Setup()
		So(err, ShouldBeNil)

		// 测试结束后清理
		defer suite.Cleanup()

		// ========== 创建测试（RM1xx） ==========
		Convey("创建测试（RM1xx）", func() {
			internal.RunCommonCreateTests(suite)
		})

		// ========== 负向测试（RM1xx 121-140） ==========
		Convey("负向测试（RM1xx 121-140）", func() {
			internal.RunCommonNegativeTests(suite)
		})

		// ========== 边界测试（RM1xx 141-160） ==========
		Convey("边界测试（RM1xx 141-160）", func() {
			internal.RunCommonBoundaryTests(suite)
		})

		// ========== 安全测试（RM1xx 161-170） ==========
		Convey("安全测试（RM1xx 161-170）", func() {
			internal.RunCommonSecurityTests(suite)
		})
	})
}

// TestMySQLResourceRead MySQL Resource读取AT测试
func TestMySQLResourceRead(t *testing.T) {
	Convey("MySQL Resource读取AT测试 - 初始化", t, func() {
		suite, err := internal.NewTestSuite(t, "mysql")
		So(err, ShouldBeNil)

		err = suite.Setup()
		So(err, ShouldBeNil)
		defer suite.Cleanup()

		// ========== 读取测试（RM2xx） ==========
		Convey("读取测试（RM2xx）", func() {
			internal.RunCommonReadTests(suite)
		})
	})
}

// TestMySQLResourceUpdate MySQL Resource更新AT测试
func TestMySQLResourceUpdate(t *testing.T) {
	Convey("MySQL Resource更新AT测试 - 初始化", t, func() {
		suite, err := internal.NewTestSuite(t, "mysql")
		So(err, ShouldBeNil)

		err = suite.Setup()
		So(err, ShouldBeNil)
		defer suite.Cleanup()

		// ========== 更新测试（RM3xx） ==========
		Convey("更新测试（RM3xx）", func() {
			internal.RunCommonUpdateTests(suite)
		})
	})
}

// TestMySQLResourceDelete MySQL Resource删除AT测试
func TestMySQLResourceDelete(t *testing.T) {
	Convey("MySQL Resource删除AT测试 - 初始化", t, func() {
		suite, err := internal.NewTestSuite(t, "mysql")
		So(err, ShouldBeNil)

		err = suite.Setup()
		So(err, ShouldBeNil)
		defer suite.Cleanup()

		// ========== 删除测试（RM4xx） ==========
		Convey("删除测试（RM4xx）", func() {
			internal.RunCommonDeleteTests(suite)
		})

		// ========== 名称唯一性测试（RM5xx） ==========
		Convey("名称唯一性测试（RM5xx）", func() {
			internal.RunCommonNameUniquenessTests(suite)
		})
	})
}
