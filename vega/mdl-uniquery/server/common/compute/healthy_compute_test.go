// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package compute

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"uniquery/interfaces"
)

var sourceRecords = interfaces.Records{
	map[string]any{
		"level":            1,
		"type":             "atomic",
		"event_model_name": "yyy",
		"event_model_id":   uint64(222222),
		"title":            "yyy_紧急",
	},
	map[string]any{
		"level":            4,
		"type":             "atomic",
		"event_model_name": "xxx",
		"event_model_id":   uint64(1111111),

		"title": "xxx_提示",
	},
	map[string]any{
		"level":            4,
		"type":             "atomic",
		"event_model_name": "zzz",
		"event_model_id":   uint64(333333),

		"title": "zzz_提示",
	},
}
var group_fields = []string{
	"level",
}

func TestMaxLevelMap(t *testing.T) {
	Convey("Test MaxLevelMap", t, func() {
		Convey("Success", func() {
			expectRecord := interfaces.Records{
				map[string]any{
					"level":            1,
					"type":             "atomic",
					"event_model_name": "yyy",
					"event_model_id":   uint64(222222),
					"title":            "yyy_紧急",
				},
			}
			records, level, _ := MaxLevelMap(sourceRecords)
			So(level, ShouldEqual, 1)
			So(records, ShouldResemble, expectRecord)
		})

		Convey("sourceRecords is empty", func() {
			records, level, _ := MaxLevelMap(interfaces.Records{})
			So(level, ShouldEqual, 0)
			So(records, ShouldBeNil)
		})
	})

}

func TestEventDataGroupAggregation(t *testing.T) {
	Convey("Test EventDataGroupAggregation", t, func() {
		Convey("Success", func() {
			expectRecord := map[string]interfaces.Records{
				"1": {
					map[string]any{
						"level": 1,
						"title": "yyy_紧急",

						"type":             "atomic",
						"event_model_name": "yyy",
						"event_model_id":   uint64(222222),
					},
				},
				"4": {
					map[string]any{
						"level":            4,
						"title":            "xxx_提示",
						"type":             "atomic",
						"event_model_name": "xxx",
						"event_model_id":   uint64(1111111),
					},
					map[string]any{
						"level":            4,
						"title":            "zzz_提示",
						"type":             "atomic",
						"event_model_name": "zzz",
						"event_model_id":   uint64(333333),
					},
				},
			}
			groupRecords := EventDataGroupAggregation(sourceRecords, group_fields)
			So(groupRecords, ShouldResemble, expectRecord)
		})
		Convey("sourceRecords is empty", func() {
			expectRecord := map[string]interfaces.Records{}
			groupRecords := EventDataGroupAggregation(sourceRecords, []string{})
			So(groupRecords, ShouldResemble, expectRecord)
		})
	})

}

func TestGetGroupFieldValue(t *testing.T) {
	Convey("Test GetGroupFieldValue", t, func() {
		Convey("Success", func() {
			groupFieldValueStr := GetGroupFieldValue(map[string]string{
				"level":            "4",
				"title":            "zzz_提示",
				"type":             "atomic",
				"event_model_name": "zzz",
				"event_model_id":   "333333",
			}, group_fields)
			So(groupFieldValueStr, ShouldEqual, "4")
		})
	})

}

func TestSourceDataGroupAggregation(t *testing.T) {
	Convey("Test SourceDataGroupAggregation", t, func() {
		Convey("Success", func() {
			var sourceRecords = interfaces.Records{
				map[string]any{
					"level":            1,
					"type":             "atomic",
					"event_model_name": "yyy",
					"event_model_id":   uint64(222222),
					"title":            "yyy_紧急",
					"trigger_data": interfaces.Records{
						map[string]any{
							"level":            1,
							"type":             "atomic",
							"event_model_name": "aaa",
							"event_model_id":   uint64(5555),
							"title":            "aaa_紧急",
						},
					},
				},
			}
			expectRecord := map[string]interfaces.Records{
				"1": {
					map[string]any{
						"level":            "1",
						"type":             "atomic",
						"event_model_name": "aaa",
						"title":            "aaa_紧急",
					},
				},
			}
			groupRecords := SourceDataGroupAggregation(sourceRecords, group_fields)

			So(groupRecords, ShouldResemble, expectRecord)
		})
		Convey("sourceRecords is empty", func() {
			var sourceRecords = interfaces.Records{}
			expectRecord := map[string]interfaces.Records{}
			groupRecords := SourceDataGroupAggregation(sourceRecords, group_fields)
			So(groupRecords, ShouldResemble, expectRecord)
		})
		Convey("group_fields is empty", func() {
			var sourceRecords = interfaces.Records{}
			expectRecord := map[string]interfaces.Records{}
			groupRecords := SourceDataGroupAggregation(sourceRecords, []string{})
			So(groupRecords, ShouldResemble, expectRecord)
		})
	})

}
