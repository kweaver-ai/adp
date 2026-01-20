// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package utils

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// TestGenerateMCPKey 测试 GenerateMCPKey 函数
func TestGenerateMCPKey(t *testing.T) {
	Convey("TestGenerateMCPKey", t, func() {
		Convey("正常参数", func() {
			result := GenerateMCPKey("mcp-001", 1)
			So(result, ShouldEqual, "mcp-001-1")
		})

		Convey("版本号为 0", func() {
			result := GenerateMCPKey("mcp-001", 0)
			So(result, ShouldEqual, "mcp-001-0")
		})

		Convey("空 mcpID", func() {
			result := GenerateMCPKey("", 1)
			So(result, ShouldEqual, "-1")
		})
	})
}

// TestGenerateMCPServerVersion 测试 GenerateMCPServerVersion 函数
func TestGenerateMCPServerVersion(t *testing.T) {
	Convey("TestGenerateMCPServerVersion", t, func() {
		Convey("版本号 1", func() {
			result := GenerateMCPServerVersion(1)
			So(result, ShouldEqual, "1.0.0")
		})

		Convey("版本号 10", func() {
			result := GenerateMCPServerVersion(10)
			So(result, ShouldEqual, "10.0.0")
		})

		Convey("版本号 0", func() {
			result := GenerateMCPServerVersion(0)
			So(result, ShouldEqual, "0.0.0")
		})
	})
}
