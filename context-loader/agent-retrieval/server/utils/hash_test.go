// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package utils

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// TestMD5 测试 MD5 函数
func TestMD5(t *testing.T) {
	Convey("TestMD5", t, func() {
		Convey("正常字符串", func() {
			result := MD5("hello world")
			So(result, ShouldEqual, "5eb63bbbe01eeed093cb22bb8f5acdc3")
		})

		Convey("空字符串", func() {
			result := MD5("")
			So(result, ShouldEqual, "d41d8cd98f00b204e9800998ecf8427e")
		})

		Convey("相同输入产生相同输出", func() {
			result1 := MD5("test")
			result2 := MD5("test")
			So(result1, ShouldEqual, result2)
		})

		Convey("不同输入产生不同输出", func() {
			result1 := MD5("test1")
			result2 := MD5("test2")
			So(result1, ShouldNotEqual, result2)
		})
	})
}

// TestObjectMD5Hash 测试 ObjectMD5Hash 函数
func TestObjectMD5Hash(t *testing.T) {
	Convey("TestObjectMD5Hash", t, func() {
		Convey("简单对象", func() {
			obj := map[string]string{"key": "value"}
			result, err := ObjectMD5Hash(obj)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeEmpty)
		})

		Convey("结构体对象", func() {
			type TestStruct struct {
				Name  string `json:"name"`
				Value int    `json:"value"`
			}
			obj := TestStruct{Name: "test", Value: 123}
			result, err := ObjectMD5Hash(obj)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeEmpty)
		})

		Convey("相同对象产生相同哈希", func() {
			obj1 := map[string]int{"a": 1, "b": 2}
			obj2 := map[string]int{"a": 1, "b": 2}
			result1, _ := ObjectMD5Hash(obj1)
			result2, _ := ObjectMD5Hash(obj2)
			So(result1, ShouldEqual, result2)
		})

		Convey("nil 对象", func() {
			result, err := ObjectMD5Hash(nil)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeEmpty)
		})
	})
}
