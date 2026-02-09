// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package event

import (
	"context"
	"testing"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	"uniquery/interfaces"
)

var (
	testENCtx = context.WithValue(context.Background(), rest.XLangKey, rest.AmericanEnglish)
)

var sourceRecords = interfaces.Records{
	map[string]any{
		"level":            "1",
		"type":             "atomic",
		"event_model_name": "yyy",
		"event_model_id":   uint64(222222),
		"title":            "yyy_紧急",
	},
	map[string]any{
		"level":            "4",
		"type":             "atomic",
		"event_model_name": "xxx",
		"event_model_id":   uint64(1111111),

		"title": "xxx_提示",
	},
	map[string]any{
		"level":            "4",
		"type":             "atomic",
		"event_model_name": "zzz",
		"event_model_id":   uint64(333333),

		"title": "zzz_提示",
	},
}
var group_fields = []string{
	"level",
}

func TestFlatteFieldMap(t *testing.T) {
	Convey("Test flatteFieldMap", t, func() {
		element := map[string]any{
			"level": "4",
		}
		Convey("success", func() {
			expectedStr := "'level':'4'"
			res := flatteFieldMap("metric_model", element)
			So(res, ShouldEqual, expectedStr)
		})
		Convey("success,float64", func() {
			element := map[string]any{
				"level": 0.1666666666,
			}
			expectedStr := "'level':'0.17'"
			res := flatteFieldMap("data_view", element)
			So(res, ShouldEqual, expectedStr)
		})
	})

}

func TestFlatteFilters(t *testing.T) {
	Convey("Test flatteFilters", t, func() {

		Convey("success,LogicOperator case and ", func() {
			f := interfaces.LogicFilter{
				LogicOperator: "and",
				FilterExpress: interfaces.FilterExpress{
					Name:      "level",
					Value:     1,
					Operation: "=",
				},
				Children: []interfaces.LogicFilter{
					{
						LogicOperator: "",
						FilterExpress: interfaces.FilterExpress{
							Name:      "title",
							Value:     "xxx",
							Operation: "=",
						},
						Children: []interfaces.LogicFilter{},
					},
				},
			}
			expectedStr := "(('title' = 'xxx'))"
			res := flatteFilters(f)
			So(res, ShouldEqual, expectedStr)
		})
		Convey("success,LogicOperator case or ", func() {
			f := interfaces.LogicFilter{
				LogicOperator: "or",
				FilterExpress: interfaces.FilterExpress{
					Name:      "level",
					Value:     1,
					Operation: "=",
				},
				Children: []interfaces.LogicFilter{
					{
						LogicOperator: "",
						FilterExpress: interfaces.FilterExpress{
							Name:      "title",
							Value:     "xxx",
							Operation: "=",
						},
						Children: []interfaces.LogicFilter{},
					},
				},
			}
			expectedStr := "(('title' = 'xxx'))"
			res := flatteFilters(f)
			So(res, ShouldEqual, expectedStr)
		})
	})

}
func TestGetKeysFromFilters(t *testing.T) {
	Convey("Test GetKeysFromFilters", t, func() {
		Convey("success,LogicOperator case and ", func() {
			f := interfaces.LogicFilter{
				LogicOperator: "and",

				Children: []interfaces.LogicFilter{
					{
						LogicOperator: "",
						FilterExpress: interfaces.FilterExpress{
							Name:      "title",
							Value:     "xxx",
							Operation: "=",
						},
						Children: []interfaces.LogicFilter{},
					},
					{
						LogicOperator: "",
						FilterExpress: interfaces.FilterExpress{
							Name:      "level",
							Value:     1,
							Operation: "=",
						},
						Children: []interfaces.LogicFilter{},
					},
				},
			}
			res := GetKeysFromFilters(f)
			So(res, ShouldResemble, []string{"title", "level"})
		})
	})

}

func TestCompose(t *testing.T) {
	Convey("Test Compose", t, func() {

		Convey("level = 6 with labels ,zh-CN", func() {
			model := interfaces.EventModel{
				EventModelID:   "11",
				EventModelName: "xxx",
				DataSourceType: "metric_model",
				DataSourceName: []string{"系统一分钟负载"},
			}
			formula := interfaces.FormulaItem{
				Level: 6,
				Filter: interfaces.LogicFilter{
					LogicOperator: "or",
					FilterExpress: interfaces.FilterExpress{
						Name:      "level",
						Value:     1,
						Operation: "=",
					},
					Children: []interfaces.LogicFilter{
						{
							LogicOperator: "",
							FilterExpress: interfaces.FilterExpress{
								Name:      "title",
								Value:     "xxx",
								Operation: "=",
							},
							Children: []interfaces.LogicFilter{},
						},
					},
				},
			}
			element := map[string]any{
				"title": "xxx",
			}
			record := map[string]any{
				"labels.host_ip": "localhost",
			}

			res := Compose(testCtx, model, formula, element, record)
			So(res, ShouldEqual, "监控对象(localhost)的监控项([系统一分钟负载])已经恢复正常('title':'xxx')")

		})
		Convey("level = 6 without labels", func() {
			model := interfaces.EventModel{
				EventModelID:   "11",
				EventModelName: "xxx",
				DataSourceType: "metric_model",
				DataSourceName: []string{"系统一分钟负载"},
			}
			formula := interfaces.FormulaItem{
				Level: 6,
				Filter: interfaces.LogicFilter{
					LogicOperator: "or",
					FilterExpress: interfaces.FilterExpress{
						Name:      "level",
						Value:     1,
						Operation: "=",
					},
					Children: []interfaces.LogicFilter{
						{
							LogicOperator: "",
							FilterExpress: interfaces.FilterExpress{
								Name:      "title",
								Value:     "xxx",
								Operation: "=",
							},
							Children: []interfaces.LogicFilter{},
						},
					},
				},
			}
			element := map[string]any{
				"value": "6",
			}
			record := map[string]any{}

			res := Compose(testCtx, model, formula, element, record)
			So(res, ShouldEqual, "监控对象(未知)的监控项([系统一分钟负载])已经恢复正常('value':'6')")

		})
		Convey("level != 6 without labels ,en-US", func() {
			model := interfaces.EventModel{
				EventModelID:   "11",
				EventModelName: "xxx",
				DataSourceType: "metric_model",
				DataSourceName: []string{"系统一分钟负载"},
			}
			formula := interfaces.FormulaItem{
				Level: 1,
				Filter: interfaces.LogicFilter{
					LogicOperator: "or",
					FilterExpress: interfaces.FilterExpress{
						Name:      "level",
						Value:     1,
						Operation: "=",
					},
					Children: []interfaces.LogicFilter{
						{
							LogicOperator: "",
							FilterExpress: interfaces.FilterExpress{
								Name:      "title",
								Value:     "xxx",
								Operation: "=",
							},
							Children: []interfaces.LogicFilter{},
						},
					},
				},
			}
			element := map[string]any{
				"title": "xxx",
			}
			record := map[string]any{}

			res := Compose(testENCtx, model, formula, element, record)
			So(res, ShouldEqual, "The monitored object (unknown) monitoring item([系统一分钟负载]) is abnormal('title':'xxx')")
		})
		Convey("level != 6 with labels ,en-US", func() {
			model := interfaces.EventModel{
				EventModelID:   "11",
				EventModelName: "xxx",
				DataSourceType: "metric_model",
				DataSourceName: []string{"系统一分钟负载"},
			}
			formula := interfaces.FormulaItem{
				Level: 1,
				Filter: interfaces.LogicFilter{
					LogicOperator: "or",
					FilterExpress: interfaces.FilterExpress{
						Name:      "level",
						Value:     1,
						Operation: "=",
					},
					Children: []interfaces.LogicFilter{
						{
							LogicOperator: "",
							FilterExpress: interfaces.FilterExpress{
								Name:      "title",
								Value:     "xxx",
								Operation: "=",
							},
							Children: []interfaces.LogicFilter{},
						},
					},
				},
			}
			element := map[string]any{
				"title": "xxx",
			}
			record := map[string]any{"labels.host_ip": "localhost"}

			res := Compose(testENCtx, model, formula, element, record)
			So(res, ShouldEqual, "The monitored object (localhost) monitoring item([系统一分钟负载]) is abnormal('title':'xxx')")
		})
	})
}

func TestCombine(t *testing.T) {
	Convey("Test Combine", t, func() {

		Convey("success", func() {
			records := interfaces.Records{
				map[string]any{
					"level":            1,
					"type":             "atomic",
					"event_model_name": "yyy",
					"event_model_id":   uint64(222222),
					"title":            "yyy_紧急",
				},
			}
			expected := interfaces.EventContext{
				Level:         1,
				Score:         90.0,
				GroupFields:   []string{},
				SourceRecords: records,
			}

			res := Combine(records, 1, 90.0)
			So(res, ShouldResemble, expected)
		})
	})

}

func TestGroupCombine(t *testing.T) {
	Convey("Test GroupCombine", t, func() {

		Convey("success", func() {
			expected := interfaces.EventContext{
				Score:         39,
				Level:         1,
				SourceRecords: sourceRecords,
				GroupFields:   group_fields,
			}
			res := GroupCombine(sourceRecords, group_fields)
			So(res, ShouldResemble, expected)
		})
	})

}

func TestGenerateMessage(t *testing.T) {
	Convey("Test GenerateMessage", t, func() {

		Convey("healthy_compute", func() {
			context := interfaces.EventContext{
				Score:         39,
				Level:         1,
				SourceRecords: sourceRecords,
				GroupFields:   group_fields,
			}
			res := GenerateMessage("healthy_compute", context, []string{}, "主要")
			So(res, ShouldResemble, "基于1个紧急事件 2个提示事件 ,生成了一个等级为主要的聚合事件")
		})
		Convey("group_aggregation", func() {
			context := interfaces.EventContext{
				Score:         39,
				Level:         1,
				SourceRecords: sourceRecords,
				GroupFields:   group_fields,
			}
			res := GenerateMessage("group_aggregation", context, []string{"labels.host_ip", "labels.host_name"}, "紧急")
			So(res, ShouldResemble, "基于1个紧急事件 2个提示事件 ,分组字段为[labels.host_ip,labels.host_name],生成了一个等级为紧急的聚合事件")
		})
	})
}
func TestRemoveDuplicates(t *testing.T) {
	Convey("Test RemoveDuplicates", t, func() {
		Convey("success,RemoveDuplicates ", func() {
			key := []string{"level", "level", "level", "level"}
			res := RemoveDuplicates(key)
			So(res, ShouldResemble, []string{"level"})
		})
	})
}
