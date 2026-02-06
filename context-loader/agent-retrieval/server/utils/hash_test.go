// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package utils

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

// TestMD5 测试 MD5 函数
func TestMD5(t *testing.T) {
	convey.Convey("TestMD5", t, func() {
		convey.Convey("正常字符串", func() {
			result := MD5("hello world")
			convey.So(result, convey.ShouldEqual, "5eb63bbbe01eeed093cb22bb8f5acdc3")
		})

		convey.Convey("空字符串", func() {
			result := MD5("")
			convey.So(result, convey.ShouldEqual, "d41d8cd98f00b204e9800998ecf8427e")
		})

		convey.Convey("相同输入产生相同输出", func() {
			result1 := MD5("test")
			result2 := MD5("test")
			convey.So(result1, convey.ShouldEqual, result2)
		})

		convey.Convey("不同输入产生不同输出", func() {
			result1 := MD5("test1")
			result2 := MD5("test2")
			convey.So(result1, convey.ShouldNotEqual, result2)
		})
	})
}

// TestObjectMD5Hash 测试 ObjectMD5Hash 函数
func TestObjectMD5Hash(t *testing.T) {
	convey.Convey("TestObjectMD5Hash", t, func() {
		convey.Convey("简单对象", func() {
			obj := map[string]string{"key": "value"}
			result, err := ObjectMD5Hash(obj)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldNotBeEmpty)
		})

		convey.Convey("结构体对象", func() {
			type TestStruct struct {
				Name  string `json:"name"`
				Value int    `json:"value"`
			}
			obj := TestStruct{Name: "test", Value: 123}
			result, err := ObjectMD5Hash(obj)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldNotBeEmpty)
		})

		convey.Convey("相同对象产生相同哈希", func() {
			obj1 := map[string]int{"a": 1, "b": 2}
			obj2 := map[string]int{"a": 1, "b": 2}
			result1, _ := ObjectMD5Hash(obj1)
			result2, _ := ObjectMD5Hash(obj2)
			convey.So(result1, convey.ShouldEqual, result2)
		})

		convey.Convey("nil 对象", func() {
			result, err := ObjectMD5Hash(nil)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldNotBeEmpty)
		})
	})
}
