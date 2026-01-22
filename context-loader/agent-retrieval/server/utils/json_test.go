// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package utils

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// TestJSONToObject_Success 测试 JSONToObject 成功场景
func TestJSONToObject_Success(t *testing.T) {
	Convey("TestJSONToObject_Success", t, func() {
		type TestStruct struct {
			Name  string `json:"name"`
			Value int    `json:"value"`
		}

		jsonStr := `{"name": "test", "value": 123}`
		result := JSONToObject[TestStruct](jsonStr)
		So(result.Name, ShouldEqual, "test")
		So(result.Value, ShouldEqual, 123)
	})
}

// TestJSONToObject_EmptyString 测试 JSONToObject 空字符串
func TestJSONToObject_EmptyString(t *testing.T) {
	Convey("TestJSONToObject_EmptyString", t, func() {
		type TestStruct struct {
			Name string `json:"name"`
		}

		result := JSONToObject[TestStruct]("")
		So(result.Name, ShouldEqual, "")
	})
}

// TestJSONToObject_InvalidJSON 测试 JSONToObject 无效 JSON
func TestJSONToObject_InvalidJSON(t *testing.T) {
	Convey("TestJSONToObject_InvalidJSON", t, func() {
		type TestStruct struct {
			Name string `json:"name"`
		}

		result := JSONToObject[TestStruct]("invalid json")
		So(result.Name, ShouldEqual, "")
	})
}

// TestJSONToObjectWithError_Success 测试 JSONToObjectWithError 成功场景
func TestJSONToObjectWithError_Success(t *testing.T) {
	Convey("TestJSONToObjectWithError_Success", t, func() {
		type TestStruct struct {
			Name string `json:"name"`
		}

		result, err := JSONToObjectWithError[TestStruct](`{"name": "test"}`)
		So(err, ShouldBeNil)
		So(result.Name, ShouldEqual, "test")
	})
}

// TestJSONToObjectWithError_EmptyString 测试 JSONToObjectWithError 空字符串
func TestJSONToObjectWithError_EmptyString(t *testing.T) {
	Convey("TestJSONToObjectWithError_EmptyString", t, func() {
		type TestStruct struct {
			Name string `json:"name"`
		}

		result, err := JSONToObjectWithError[TestStruct]("")
		So(err, ShouldBeNil)
		So(result.Name, ShouldEqual, "")
	})
}

// TestJSONToObjectWithError_InvalidJSON 测试 JSONToObjectWithError 无效 JSON
func TestJSONToObjectWithError_InvalidJSON(t *testing.T) {
	Convey("TestJSONToObjectWithError_InvalidJSON", t, func() {
		type TestStruct struct {
			Name string `json:"name"`
		}

		_, err := JSONToObjectWithError[TestStruct]("invalid json")
		So(err, ShouldNotBeNil)
	})
}

// TestAnyToObject_Success 测试 AnyToObject 成功场景
func TestAnyToObject_Success(t *testing.T) {
	Convey("TestAnyToObject_Success", t, func() {
		type TestStruct struct {
			Name  string `json:"name"`
			Value int    `json:"value"`
		}

		source := map[string]interface{}{
			"name":  "test",
			"value": 123,
		}

		var result TestStruct
		err := AnyToObject(source, &result)
		So(err, ShouldBeNil)
		So(result.Name, ShouldEqual, "test")
		So(result.Value, ShouldEqual, 123)
	})
}

// TestAnyToObject_SliceToStruct 测试 AnyToObject 数组转换
func TestAnyToObject_SliceToStruct(t *testing.T) {
	Convey("TestAnyToObject_SliceToStruct", t, func() {
		source := []map[string]interface{}{
			{"name": "item1"},
			{"name": "item2"},
		}

		type Item struct {
			Name string `json:"name"`
		}

		var result []Item
		err := AnyToObject(source, &result)
		So(err, ShouldBeNil)
		So(len(result), ShouldEqual, 2)
		So(result[0].Name, ShouldEqual, "item1")
	})
}
