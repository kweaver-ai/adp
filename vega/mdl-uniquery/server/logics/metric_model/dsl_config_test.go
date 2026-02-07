// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package metric_model

import (
	"context"
	"errors"
	"testing"
	"time"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	cond "uniquery/common/condition"
	"uniquery/interfaces"
	dtype "uniquery/interfaces/data_type"
)

var (
	testCtx      = context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage)
	dslCfgStart1 = int64(1724515200000)
	dslCfgEnd1   = int64(1727193600000)
)

func Test_generateDslByConfig(t *testing.T) {

	Convey("Test generateDslByConfig, IsInstantQuery=true", t, func() {

		starti := int64(1727057677000)
		endi := int64(1727057977000)
		view := &interfaces.DataView{
			ViewID:        "1",
			QueryType:     interfaces.QueryType_DSL,
			Type:          interfaces.ViewType_Atomic,
			TechnicalName: "test",
			FieldsMap: map[string]*cond.ViewField{
				"__index_base": {
					Name: "__index_base",
					Type: dtype.DataType_String,
				},
				"category": {
					Name: "category",
					Type: dtype.DataType_Text,
				},
				"value": {
					Name: "value",
					Type: dtype.DataType_Integer,
				},
				"geo.location": {
					Name: "geo.location",
					Type: dtype.DataType_Point,
				},
				"@timestamp": {
					Name: "@timestamp",
					Type: dtype.DataType_Datetime,
				},
			},
		}
		query := interfaces.MetricModelQuery{
			MetricType: interfaces.ATOMIC_METRIC,
			QueryType:  interfaces.DSL_CONFIG,
			// DataView:   view,
			FormulaConfig: interfaces.MetricModelFormulaConfig{
				Buckets: []*interfaces.MetricModelFormulaConfigBucket{
					{
						Type:      interfaces.BUCKET_TYPE_TERMS,
						Name:      "index_base",
						Field:     "__index_base",
						Size:      1000,
						Order:     interfaces.TERMS_ORDER_TYPE_FIELD,
						Direction: interfaces.ASC_DIRECTION,
					}, {
						Type:      interfaces.BUCKET_TYPE_TERMS,
						Name:      "category",
						Field:     "category",
						Order:     interfaces.TERMS_ORDER_TYPE_VALUE,
						Direction: interfaces.DESC_DIRECTION,
					}, {
						Type:      interfaces.BUCKET_TYPE_TERMS,
						Name:      "category",
						Field:     "category",
						Order:     interfaces.TERMS_ORDER_TYPE_COUNT,
						Direction: interfaces.DESC_DIRECTION,
					}, {
						Type:  interfaces.BUCKET_TYPE_RANGE,
						Name:  "value",
						Field: "value",
						Ranges: []interfaces.MetricModelFormulaConfigBucketRange{
							{},
						},
					}, {
						Type: interfaces.BUCKET_TYPE_FILTERS,
						Filters: map[string]interfaces.MetricModelFormulaConfigBucketFilter{
							"value_1": {
								QueryString: "value:1",
							},
							"value_12": {
								QueryString: "value:12",
							},
							"value_123": {
								QueryString: "value:123",
							},
						},
					}, {
						Type:      interfaces.BUCKET_TYPE_GEOHASH_GRID,
						Field:     "geo.location",
						Precision: 10,
					}, {
						Type:  interfaces.BUCKET_TYPE_DATE_RANGE,
						Field: "@timestamp",
						Ranges: []interfaces.MetricModelFormulaConfigBucketRange{
							{},
						},
					},
				},
				DateHistogram: &interfaces.MetricModelFormulaConfigDateHistogram{
					Field:         "@timestamp",
					IntervalType:  interfaces.INTERVAL_TYPE_CALENDAR,
					IntervalValue: interfaces.CALENDAR_STEP_DAY,
				},
				Aggregation: &interfaces.MetricModelFormulaConfigAggregation{
					Type: interfaces.AGGR_TYPE_DOC_COUNT,
				},
				QueryString: "category:log",
			},
			QueryTimeParams: interfaces.QueryTimeParams{
				IsInstantQuery: true,
				Start:          &starti,
				End:            &endi,
			},
		}
		termsBktIdx := 0
		rangeBktIdx := 3
		filtersBktIdx := 4
		geohashGridBktIdx := 5
		dateRangeBktIdx := 6

		Convey("FormulaConfig is nil", func() {
			query.FormulaConfig = nil
			_, err := generateDslByConfig(testCtx, &query, view)
			So(err, ShouldNotBeNil)
		})
		Convey("too many buckets", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			bktInfos := dslConfig.Buckets
			bktInfos = append(bktInfos, dslConfig.Buckets...)
			bktInfos = append(bktInfos, dslConfig.Buckets...)
			bktInfos = append(bktInfos, dslConfig.Buckets...)
			dslConfig.Buckets = bktInfos
			query.FormulaConfig = dslConfig
			_, err := generateDslByConfig(testCtx, &query, view)
			So(err, ShouldNotBeNil)
		})
		Convey("aggrInfo is nil", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.Aggregation = nil
			query.FormulaConfig = dslConfig
			_, err := generateDslByConfig(testCtx, &query, view)
			So(err, ShouldNotBeNil)
		})
		Convey("range_query, but date_histogram is nil", func() {
			query.IsInstantQuery = false
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.DateHistogram = nil
			query.FormulaConfig = dslConfig
			_, err := generateDslByConfig(testCtx, &query, view)
			So(err, ShouldNotBeNil)
		})
		Convey("bucket type is incorrect", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.Buckets[termsBktIdx].Type = "invalid"
			query.FormulaConfig = dslConfig
			_, err := generateDslByConfig(testCtx, &query, view)
			So(err, ShouldNotBeNil)
		})

		// bucket terms
		Convey("bucket terms, field not exist", func() {
			delete(view.FieldsMap, "__index_base")
			_, err := generateDslByConfig(testCtx, &query, view)
			So(err, ShouldNotBeNil)
		})
		Convey("bucket terms, direction is incorrect", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.Buckets[termsBktIdx].Direction = "invalid"
			query.FormulaConfig = dslConfig
			_, err := generateDslByConfig(testCtx, &query, view)
			So(err, ShouldNotBeNil)
		})
		Convey("bucket terms, order is incorrect", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.Buckets[termsBktIdx].Order = "invalid"
			query.FormulaConfig = dslConfig
			_, err := generateDslByConfig(testCtx, &query, view)
			So(err, ShouldNotBeNil)
		})

		// bucket range
		Convey("bucket range, field not exist", func() {
			delete(view.FieldsMap, "value")
			_, err := generateDslByConfig(testCtx, &query, view)
			So(err, ShouldNotBeNil)
		})
		Convey("bucket range, field type is not number", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.Buckets[rangeBktIdx].Field = "__index_base"
			query.FormulaConfig = dslConfig
			_, err := generateDslByConfig(testCtx, &query, view)
			So(err, ShouldNotBeNil)
		})
		Convey("bucket range, missing ranges", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.Buckets[rangeBktIdx].Ranges = []interfaces.MetricModelFormulaConfigBucketRange{}
			query.FormulaConfig = dslConfig
			_, err := generateDslByConfig(testCtx, &query, view)
			So(err, ShouldNotBeNil)
		})

		// bucket filters
		Convey("bucket filters, missing filters", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.Buckets[filtersBktIdx].Filters = map[string]interfaces.MetricModelFormulaConfigBucketFilter{}
			query.FormulaConfig = dslConfig
			_, err := generateDslByConfig(testCtx, &query, view)
			So(err, ShouldNotBeNil)
		})
		Convey("bucket filters, missing QueryString", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.Buckets[filtersBktIdx].Filters["value_1"] = interfaces.MetricModelFormulaConfigBucketFilter{
				QueryString: "",
			}
			query.FormulaConfig = dslConfig
			_, err := generateDslByConfig(testCtx, &query, view)
			So(err, ShouldNotBeNil)
		})
		Convey("bucket filters, missing Name, ok", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.Buckets[filtersBktIdx].Name = ""
			query.FormulaConfig = dslConfig
			_, err := generateDslByConfig(testCtx, &query, view)
			So(err, ShouldBeNil)
		})

		// bucket geohash_grid
		Convey("bucket geohash_grid, field not exist", func() {
			delete(view.FieldsMap, "geo.location")
			_, err := generateDslByConfig(testCtx, &query, view)
			So(err, ShouldNotBeNil)
		})
		Convey("bucket geohash_grid, field type is not geo_point", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.Buckets[geohashGridBktIdx].Field = "__index_base"
			query.FormulaConfig = dslConfig
			_, err := generateDslByConfig(testCtx, &query, view)
			So(err, ShouldNotBeNil)
		})
		Convey("bucket geohash_grid, precision is not in range", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.Buckets[geohashGridBktIdx].Precision = interfaces.BUCKET_GEOHASH_GRID_MAX_PRECISION + 1
			query.FormulaConfig = dslConfig
			_, err := generateDslByConfig(testCtx, &query, view)
			So(err, ShouldNotBeNil)
		})

		// bucket date_range
		Convey("bucket date_range, field not exist", func() {
			delete(view.FieldsMap, "@timestamp")
			_, err := generateDslByConfig(testCtx, &query, view)
			So(err, ShouldNotBeNil)
		})
		Convey("bucket date_range, field type is not date", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.Buckets[dateRangeBktIdx].Field = "__index_base"
			query.FormulaConfig = dslConfig
			_, err := generateDslByConfig(testCtx, &query, view)
			So(err, ShouldNotBeNil)
		})
		Convey("bucket date_range, missing ranges", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.Buckets[dateRangeBktIdx].Ranges = []interfaces.MetricModelFormulaConfigBucketRange{}
			query.FormulaConfig = dslConfig
			_, err := generateDslByConfig(testCtx, &query, view)
			So(err, ShouldNotBeNil)
		})
	})

	Convey("Test generateDslByConfig, IsInstantQuery=false", t, func() {
		viewtmp := &interfaces.DataView{
			ViewID:        "1",
			QueryType:     interfaces.QueryType_DSL,
			Type:          interfaces.ViewType_Atomic,
			TechnicalName: "test",
			FieldsMap: map[string]*cond.ViewField{
				"__index_base": {
					Name: "__index_base",
					Type: dtype.DataType_String,
				},
				"category": {
					Name: "category",
					Type: dtype.DataType_Text,
				},
				"value": {
					Name: "value",
					Type: dtype.DataType_Integer,
				},
				"geo.location": {
					Name: "geo.location",
					Type: dtype.DataType_Point,
				},
				"@timestamp": {
					Name: "@timestamp",
					Type: dtype.DataType_Datetime,
				},
			},
		}
		query := interfaces.MetricModelQuery{
			MetricType: interfaces.ATOMIC_METRIC,
			QueryType:  interfaces.DSL_CONFIG,
			// DataView:   viewtmp,
			FormulaConfig: interfaces.MetricModelFormulaConfig{
				Buckets: []*interfaces.MetricModelFormulaConfigBucket{
					{
						Type:      interfaces.BUCKET_TYPE_TERMS,
						Name:      "index_base",
						Field:     "__index_base",
						Size:      1000,
						Order:     interfaces.TERMS_ORDER_TYPE_FIELD,
						Direction: interfaces.ASC_DIRECTION,
					}, {
						Type:      interfaces.BUCKET_TYPE_TERMS,
						Name:      "category",
						Field:     "category",
						Order:     interfaces.TERMS_ORDER_TYPE_VALUE,
						Direction: interfaces.DESC_DIRECTION,
					}, {
						Type:      interfaces.BUCKET_TYPE_TERMS,
						Name:      "value",
						Field:     "value",
						Order:     interfaces.TERMS_ORDER_TYPE_COUNT,
						Direction: interfaces.DESC_DIRECTION,
					}, {
						Type:  interfaces.BUCKET_TYPE_RANGE,
						Name:  "value",
						Field: "value",
						Ranges: []interfaces.MetricModelFormulaConfigBucketRange{
							{},
						},
					}, {
						Type: interfaces.BUCKET_TYPE_FILTERS,
						Filters: map[string]interfaces.MetricModelFormulaConfigBucketFilter{
							"value_1": {
								QueryString: "value:1",
							},
							"value_12": {
								QueryString: "value:12",
							},
							"value_123": {
								QueryString: "value:123",
							},
						},
					}, {
						Type:      interfaces.BUCKET_TYPE_GEOHASH_GRID,
						Field:     "geo.location",
						Precision: 10,
					}, {
						Type:  interfaces.BUCKET_TYPE_DATE_RANGE,
						Field: "@timestamp",
						Ranges: []interfaces.MetricModelFormulaConfigBucketRange{
							{},
						},
					},
				},
				DateHistogram: &interfaces.MetricModelFormulaConfigDateHistogram{
					Field:         "@timestamp",
					IntervalType:  interfaces.INTERVAL_TYPE_CALENDAR,
					IntervalValue: interfaces.CALENDAR_STEP_DAY,
				},
				Aggregation: &interfaces.MetricModelFormulaConfigAggregation{
					Type: interfaces.AGGR_TYPE_DOC_COUNT,
				},
				QueryString: "category:log",
			},
			QueryTimeParams: interfaces.QueryTimeParams{
				IsInstantQuery: false,
				Start:          &dslCfgStart1,
				End:            &dslCfgEnd1,
				StepStr:        &day,
			},
		}

		// date_histogram calendar
		Convey("date_histogram field is not in date view", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.DateHistogram.Field = "invalid"
			query.FormulaConfig = dslConfig
			_, err := generateDslByConfig(testCtx, &query, viewtmp)
			So(err, ShouldNotBeNil)
		})
		Convey("date_histogram field is not date type", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.DateHistogram.Field = "__index_base"
			query.FormulaConfig = dslConfig
			_, err := generateDslByConfig(testCtx, &query, viewtmp)
			So(err, ShouldNotBeNil)
		})
		Convey("date_histogram interval_type is incorrect", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.DateHistogram.IntervalType = "invalid"
			query.FormulaConfig = dslConfig
			_, err := generateDslByConfig(testCtx, &query, viewtmp)
			So(err, ShouldNotBeNil)
		})
		Convey("date_histogram interval_type is calendar, interval_value is incorrect", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.DateHistogram.IntervalValue = "invalid"
			query.FormulaConfig = dslConfig
			_, err := generateDslByConfig(testCtx, &query, viewtmp)
			So(err, ShouldNotBeNil)
		})
		Convey("date_histogram interval_type is calendar, interval_value is auto, StepStr is incorrect", func() {
			stepi := "invalid"
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.DateHistogram.IntervalValue = interfaces.AUTO_INTERVAL
			query.FormulaConfig = dslConfig
			query.StepStr = &stepi
			_, err := generateDslByConfig(testCtx, &query, viewtmp)
			So(err, ShouldNotBeNil)
		})
		Convey("date_histogram interval_type is calendar, ok", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.DateHistogram.IntervalType = interfaces.INTERVAL_TYPE_CALENDAR
			dslConfig.DateHistogram.IntervalValue = interfaces.CALENDAR_STEP_DAY
			query.FormulaConfig = dslConfig

			patches := ApplyFuncReturn(ParseQuery, nil)
			defer patches.Reset()

			_, err := generateDslByConfig(testCtx, &query, viewtmp)
			So(err, ShouldBeNil)
		})

		// date_histogram fixed
		Convey("date_histogram interval_type is fixed, interval_value is incorrect", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.DateHistogram.IntervalType = interfaces.INTERVAL_TYPE_FIXED
			dslConfig.DateHistogram.IntervalValue = "invalid"
			query.FormulaConfig = dslConfig
			_, err := generateDslByConfig(testCtx, &query, viewtmp)
			So(err, ShouldNotBeNil)
		})
		Convey("date_histogram interval_type is fixed, interval_value is auto, StepStr is incorrect", func() {
			stepi := "invalid"
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.DateHistogram.IntervalType = interfaces.INTERVAL_TYPE_FIXED
			dslConfig.DateHistogram.IntervalValue = interfaces.AUTO_INTERVAL
			query.FormulaConfig = dslConfig
			query.StepStr = &stepi
			_, err := generateDslByConfig(testCtx, &query, viewtmp)
			So(err, ShouldNotBeNil)
		})
		Convey("date_histogram interval_type is fixed, ok", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.DateHistogram.IntervalType = interfaces.INTERVAL_TYPE_FIXED
			dslConfig.DateHistogram.IntervalValue = "1d"
			query.FormulaConfig = dslConfig

			patches := ApplyFuncReturn(ParseQuery, nil)
			defer patches.Reset()

			_, err := generateDslByConfig(testCtx, &query, viewtmp)
			So(err, ShouldBeNil)
		})

		// doc_count with no bucket
		Convey("doc_count with no bucket, but ParseQuery failed", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.Buckets = []*interfaces.MetricModelFormulaConfigBucket{}
			query.FormulaConfig = dslConfig

			patches := ApplyFuncReturn(ParseQuery, errors.New("error"))
			defer patches.Reset()

			_, err := generateDslByConfig(testCtx, &query, viewtmp)
			So(err, ShouldNotBeNil)
		})
		Convey("doc_count with no bucket, ok", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.Buckets = []*interfaces.MetricModelFormulaConfigBucket{}
			query.FormulaConfig = dslConfig

			patches := ApplyFuncReturn(ParseQuery, nil)
			defer patches.Reset()

			_, err := generateDslByConfig(testCtx, &query, viewtmp)
			So(err, ShouldBeNil)
		})

		// value_count
		Convey("value_count field is not in data view", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.Aggregation.Type = interfaces.AGGR_TYPE_VALUE_COUNT
			dslConfig.Aggregation.Field = "invalid"
			query.FormulaConfig = dslConfig
			_, err := generateDslByConfig(testCtx, &query, viewtmp)
			So(err, ShouldNotBeNil)
		})
		Convey("value_count, ok", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.Aggregation.Type = interfaces.AGGR_TYPE_VALUE_COUNT
			dslConfig.Aggregation.Field = "__index_base"
			query.FormulaConfig = dslConfig

			patches := ApplyFuncReturn(ParseQuery, nil)
			defer patches.Reset()

			_, err := generateDslByConfig(testCtx, &query, viewtmp)
			So(err, ShouldBeNil)
		})

		// cardinality
		Convey("cardinality, ok", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.Aggregation.Type = interfaces.AGGR_TYPE_CARDINALITY
			dslConfig.Aggregation.Field = "__index_base"
			query.FormulaConfig = dslConfig

			patches := ApplyFuncReturn(ParseQuery, nil)
			defer patches.Reset()

			_, err := generateDslByConfig(testCtx, &query, viewtmp)
			So(err, ShouldBeNil)
		})

		// sum
		Convey("sum, field is not number", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.Aggregation.Type = interfaces.AGGR_TYPE_SUM
			dslConfig.Aggregation.Field = "__index_base"
			query.FormulaConfig = dslConfig
			_, err := generateDslByConfig(testCtx, &query, viewtmp)
			So(err, ShouldNotBeNil)
		})
		Convey("sum, ok", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.Aggregation.Type = interfaces.AGGR_TYPE_SUM
			dslConfig.Aggregation.Field = "value"
			query.FormulaConfig = dslConfig

			patches := ApplyFuncReturn(ParseQuery, nil)
			defer patches.Reset()

			_, err := generateDslByConfig(testCtx, &query, viewtmp)
			So(err, ShouldBeNil)
		})

		// avg
		Convey("avg, field is not number", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.Aggregation.Type = interfaces.AGGR_TYPE_AVG
			dslConfig.Aggregation.Field = "__index_base"
			query.FormulaConfig = dslConfig
			_, err := generateDslByConfig(testCtx, &query, viewtmp)
			So(err, ShouldNotBeNil)
		})
		Convey("avg, ok", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.Aggregation.Type = interfaces.AGGR_TYPE_AVG
			dslConfig.Aggregation.Field = "value"
			query.FormulaConfig = dslConfig

			patches := ApplyFuncReturn(ParseQuery, nil)
			defer patches.Reset()

			_, err := generateDslByConfig(testCtx, &query, viewtmp)
			So(err, ShouldBeNil)
		})

		// max
		Convey("max, field is not number or date", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.Aggregation.Type = interfaces.AGGR_TYPE_MAX
			dslConfig.Aggregation.Field = "__index_base"
			query.FormulaConfig = dslConfig
			_, err := generateDslByConfig(testCtx, &query, viewtmp)
			So(err, ShouldNotBeNil)
		})
		Convey("max, ok", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.Aggregation.Type = interfaces.AGGR_TYPE_MAX
			dslConfig.Aggregation.Field = "value"
			query.FormulaConfig = dslConfig

			patches := ApplyFuncReturn(ParseQuery, nil)
			defer patches.Reset()

			_, err := generateDslByConfig(testCtx, &query, viewtmp)
			So(err, ShouldBeNil)
		})

		// min
		Convey("min, field is not number or date", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.Aggregation.Type = interfaces.AGGR_TYPE_MIN
			dslConfig.Aggregation.Field = "__index_base"
			query.FormulaConfig = dslConfig
			_, err := generateDslByConfig(testCtx, &query, viewtmp)
			So(err, ShouldNotBeNil)
		})
		Convey("min, ok", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.Aggregation.Type = interfaces.AGGR_TYPE_MIN
			dslConfig.Aggregation.Field = "value"
			query.FormulaConfig = dslConfig

			patches := ApplyFuncReturn(ParseQuery, nil)
			defer patches.Reset()

			_, err := generateDslByConfig(testCtx, &query, viewtmp)
			So(err, ShouldBeNil)
		})

		// percentiles
		Convey("percentiles, field is not number", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.Aggregation.Type = interfaces.AGGR_TYPE_PERCENTILES
			dslConfig.Aggregation.Field = "__index_base"
			query.FormulaConfig = dslConfig
			_, err := generateDslByConfig(testCtx, &query, viewtmp)
			So(err, ShouldNotBeNil)
		})
		Convey("percentiles, percents is empty", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.Aggregation.Type = interfaces.AGGR_TYPE_PERCENTILES
			dslConfig.Aggregation.Field = "value"
			dslConfig.Aggregation.Percents = []float64{}
			query.FormulaConfig = dslConfig
			_, err := generateDslByConfig(testCtx, &query, viewtmp)
			So(err, ShouldNotBeNil)
		})
		Convey("percentiles, ok", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.Aggregation.Type = interfaces.AGGR_TYPE_PERCENTILES
			dslConfig.Aggregation.Field = "value"
			dslConfig.Aggregation.Percents = []float64{50, 90}
			query.FormulaConfig = dslConfig

			patches := ApplyFuncReturn(ParseQuery, nil)
			defer patches.Reset()

			_, err := generateDslByConfig(testCtx, &query, viewtmp)
			So(err, ShouldBeNil)
		})
	})
}

func Test_parseDSLResult2UniresponseForDslConfig(t *testing.T) {
	Convey("Test parseDSLResult2UniresponseForDslConfig", t, func() {

		model := interfaces.MetricModel{}
		dslRes := []byte(`{
			"hits": {
				"total": {
					"value": 12
				}
			},
			"aggregations": {
				"index_base": {
					"buckets": [
						{
							"key": "test",
							"doc_count": 12,
							"category": {
								"buckets": [
									{
										"key": "log",
										"doc_count": 12,
										"value": {
											"buckets": {
												"value_1": {
													"doc_count": 1
												},
												"__other": {
													"doc_count": 11
												}
											}
										}
									}
								]
							}
						}
					]
				}
			}
		}`)

		// viewtmp := &interfaces.DataView{
		// 	ViewID:        "1",
		// 	QueryType:     interfaces.QueryType_DSL,
		// 	Type:          interfaces.ViewType_Atomic,
		// 	TechnicalName: "test",
		// 	FieldsMap: map[string]*cond.ViewField{
		// 		"__index_base": {
		// 			Name: "__index_base",
		// 			Type: dtype.DATATYPE_KEYWORD,
		// 		},
		// 		"category": {
		// 			Name: "category",
		// 			Type: dtype.DATATYPE_TEXT,
		// 		},
		// 		"value": {
		// 			Name: "value",
		// 			Type: dtype.DATATYPE_LONG,
		// 		},
		// 		"geo.location": {
		// 			Name: "geo.location",
		// 			Type: dtype.DATATYPE_GEO_POINT,
		// 		},
		// 		"@timestamp": {
		// 			Name: "@timestamp",
		// 			Type: dtype.DATATYPE_DATETIME,
		// 		},
		// 	},
		// }
		query := interfaces.MetricModelQuery{
			MetricType: interfaces.ATOMIC_METRIC,
			QueryType:  interfaces.DSL_CONFIG,
			// DataView:   viewtmp,
			FormulaConfig: interfaces.MetricModelFormulaConfig{
				Buckets: []*interfaces.MetricModelFormulaConfigBucket{
					{
						Type:      interfaces.BUCKET_TYPE_TERMS,
						Name:      "index_base",
						Field:     "__index_base",
						Size:      1000,
						Order:     interfaces.TERMS_ORDER_TYPE_FIELD,
						Direction: interfaces.ASC_DIRECTION,
					}, {
						Type:      interfaces.BUCKET_TYPE_TERMS,
						Name:      "category",
						Field:     "category",
						Order:     interfaces.TERMS_ORDER_TYPE_VALUE,
						Direction: interfaces.DESC_DIRECTION,
					}, {
						Type: interfaces.BUCKET_TYPE_FILTERS,
						Filters: map[string]interfaces.MetricModelFormulaConfigBucketFilter{
							"value_1": {
								QueryString: "value:1",
							},
							"value_12": {
								QueryString: "value:12",
							},
							"value_123": {
								QueryString: "value:123",
							},
						},
					},
				},
				DateHistogram: &interfaces.MetricModelFormulaConfigDateHistogram{
					Field:         "@timestamp",
					IntervalType:  interfaces.INTERVAL_TYPE_CALENDAR,
					IntervalValue: interfaces.CALENDAR_STEP_DAY,
				},
				Aggregation: &interfaces.MetricModelFormulaConfigAggregation{
					Type: interfaces.AGGR_TYPE_DOC_COUNT,
				},
				QueryString: "category:log",
			},
			IsCalendar: true,
			QueryTimeParams: interfaces.QueryTimeParams{
				IsInstantQuery: false,
				Start:          &dslCfgStart1,
				End:            &dslCfgEnd1,
				StepStr:        &day,
			},
			FixedStart: 1724515200000,
			FixedEnd:   1727193600000,
			MetricModelQueryParameters: interfaces.MetricModelQueryParameters{
				IncludeModel: true,
			},
		}

		Convey("dslRes is incorrect", func() {
			dslRes = []byte("")
			_, err := parseDSLResult2UniresponseForDslConfig(testCtx, dslRes, model, query)
			So(err, ShouldNotBeNil)
		})
		Convey("dslRes missing aggregations", func() {
			dslRes = []byte(`{}`)
			_, err := parseDSLResult2UniresponseForDslConfig(testCtx, dslRes, model, query)
			So(err, ShouldNotBeNil)
		})
		Convey("doc_count with no bucket, but hits total value is invalid", func() {
			dslRes = []byte(`{}`)
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.Buckets = []*interfaces.MetricModelFormulaConfigBucket{}
			query.FormulaConfig = dslConfig
			_, err := parseDSLResult2UniresponseForDslConfig(testCtx, dslRes, model, query)
			So(err, ShouldNotBeNil)
		})
		Convey("TraversalBucket is incorrect", func() {
			patches := ApplyFuncReturn(TraversalBucket, errors.New("error"))
			defer patches.Reset()

			_, err := parseDSLResult2UniresponseForDslConfig(testCtx, dslRes, model, query)
			So(err, ShouldNotBeNil)
		})
		Convey("doc_count with no bucket, ok", func() {
			dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.Buckets = []*interfaces.MetricModelFormulaConfigBucket{}
			query.FormulaConfig = dslConfig
			dslRes := []byte(`{
				"hits": {
					"total": {
						"value": 12
					}
				}
			}`)

			patches := ApplyFuncReturn(TraversalBucket, nil)
			defer patches.Reset()

			_, err := parseDSLResult2UniresponseForDslConfig(testCtx, dslRes, model, query)
			So(err, ShouldBeNil)
		})
		Convey("ok", func() {
			patches := ApplyFuncReturn(TraversalBucket, nil)
			defer patches.Reset()

			_, err := parseDSLResult2UniresponseForDslConfig(testCtx, dslRes, model, query)
			So(err, ShouldBeNil)
		})
	})
}

func Test_TraversalBucket(t *testing.T) {

	Convey("Test TraversalBucket", t, func() {
		// dataview := interfaces.DataView{
		// 	ViewID:        "1",
		// 	QueryType:     interfaces.QueryType_DSL,
		// 	Type:          interfaces.ViewType_Atomic,
		// 	TechnicalName: "test",
		// 	FieldsMap: map[string]*cond.ViewField{
		// 		"__index_base": {
		// 			Name: "__index_base",
		// 			Type: dtype.DATATYPE_KEYWORD,
		// 		},
		// 		"category": {
		// 			Name: "category",
		// 			Type: dtype.DATATYPE_TEXT,
		// 		},
		// 		"value": {
		// 			Name: "value",
		// 			Type: dtype.DATATYPE_LONG,
		// 		},
		// 		"geo.location": {
		// 			Name: "geo.location",
		// 			Type: dtype.DATATYPE_GEO_POINT,
		// 		},
		// 		"@timestamp": {
		// 			Name: "@timestamp",
		// 			Type: dtype.DATATYPE_DATETIME,
		// 		},
		// 	},
		// }

		formula_config := interfaces.MetricModelFormulaConfig{
			Buckets: []*interfaces.MetricModelFormulaConfigBucket{
				{
					Type:      interfaces.BUCKET_TYPE_TERMS,
					Name:      "index_base",
					BktName:   "index_base",
					Field:     "__index_base",
					Size:      1000,
					Order:     interfaces.TERMS_ORDER_TYPE_FIELD,
					Direction: interfaces.ASC_DIRECTION,
				}, {
					Type:      interfaces.BUCKET_TYPE_TERMS,
					Name:      "category",
					BktName:   "category",
					Field:     "category",
					Order:     interfaces.TERMS_ORDER_TYPE_VALUE,
					Direction: interfaces.DESC_DIRECTION,
				}, {
					Type:    interfaces.BUCKET_TYPE_FILTERS,
					Name:    "value",
					BktName: "value",
					Filters: map[string]interfaces.MetricModelFormulaConfigBucketFilter{
						"value_1": {
							QueryString: "value:1",
						},
					},
					OtherBucket: true,
				},
			},
			DateHistogram: &interfaces.MetricModelFormulaConfigDateHistogram{
				Field:         "@timestamp",
				IntervalType:  interfaces.INTERVAL_TYPE_CALENDAR,
				IntervalValue: interfaces.CALENDAR_STEP_DAY,
			},
			Aggregation: &interfaces.MetricModelFormulaConfigAggregation{
				Type: interfaces.AGGR_TYPE_DOC_COUNT,
			},
			QueryString: "category:log",
		}
		starti := int64(1727057677000)
		endi := int64(1727057977000)
		instant_query := interfaces.MetricModelQuery{
			MetricType: interfaces.ATOMIC_METRIC,
			QueryType:  interfaces.DSL_CONFIG,
			// DataView:      dataview,
			FormulaConfig: formula_config,
			QueryTimeParams: interfaces.QueryTimeParams{
				IsInstantQuery: true,
				Start:          &starti,
				End:            &endi,
				// LookBackDelta:  300000,
			},
		}
		range_query := interfaces.MetricModelQuery{
			MetricType: interfaces.ATOMIC_METRIC,
			QueryType:  interfaces.DSL_CONFIG,
			// DataView:      dataview,
			FormulaConfig: formula_config,
			IsCalendar:    true,
			QueryTimeParams: interfaces.QueryTimeParams{
				IsInstantQuery: false,
				Start:          &dslCfgStart1,
				End:            &dslCfgEnd1,
				StepStr:        &day,
			},
			FixedStart: 1724515200000,
			FixedEnd:   1727193600000,
		}

		labels := make(map[string]string)
		datas := make([]interfaces.MetricModelData, 0)

		Convey("instant_query, aggregation is doc_count, ok", func() {
			dslRes := []byte(`{
				"aggregations":{
					"index_base":{
						"buckets":[
							{
								"key":"test",
								"doc_count":12,
								"category":{
									"buckets":[
										{
											"key":"log",
											"doc_count":12,
											"value":{
												"buckets":{
													"value_1":{
														"doc_count":1
													},
													"__other":{
														"doc_count":11
													}
												}
											}
										}
									]
								}
							}
						]
					}
				}
			}`)
			rootNode, _ := sonic.Get(dslRes)
			aggrNode := rootNode.Get("aggregations")
			err := TraversalBucket(aggrNode, instant_query, 0, labels, &datas)
			So(err, ShouldBeNil)
		})
		Convey("range_query, aggregation is doc_count, ok", func() {
			dslRes := []byte(`{
				"aggregations": {
					"index_base": {
						"buckets": [
							{
								"key": "test",
								"doc_count": 12,
								"category": {
									"buckets": [
										{
											"key": "log",
											"doc_count": 12,
											"value": {
												"buckets": {
													"value_1": {
														"doc_count": 1,
														"__date_histogram": {
															"buckets": [
																{
																	"key_as_string": "2024-09-09T00:00:00.000Z",
																	"key": 1725840000000,
																	"doc_count": 1
																}
															]
														}
													},
													"__other": {
														"doc_count": 11,
														"__date_histogram": {
															"buckets": [
																{
																	"key_as_string": "2024-09-09T00:00:00.000Z",
																	"key": 1725840000000,
																	"doc_count": 11
																}
															]
														}
													}
												}
											}
										}
									]
								}
							}
						]
					}
				}
			}`)
			rootNode, _ := sonic.Get(dslRes)
			aggrNode := rootNode.Get("aggregations")
			err := TraversalBucket(aggrNode, range_query, 0, labels, &datas)
			So(err, ShouldBeNil)
		})
		Convey("instant_query, aggregation is value_count, ok", func() {
			dslConfig := instant_query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.Aggregation = &interfaces.MetricModelFormulaConfigAggregation{
				Type:  interfaces.AGGR_TYPE_VALUE_COUNT,
				Field: "value",
			}
			instant_query.FormulaConfig = dslConfig
			dslRes := []byte(`{
				"aggregations": {
					"index_base": {
						"buckets": [
							{
								"key": "test",
								"doc_count": 12,
								"category": {
									"buckets": [
										{
											"key": "log",
											"doc_count": 12,
											"value": {
												"buckets": {
													"value_1": {
														"doc_count": 1,
														"__value": {
															"value": 1
														}
													},
													"__other": {
														"doc_count": 11,
														"__value": {
															"value": 11
														}
													}
												}
											}
										}
									]
								}
							}
						]
					}
				}
			}`)
			rootNode, _ := sonic.Get(dslRes)
			aggrNode := rootNode.Get("aggregations")
			err := TraversalBucket(aggrNode, instant_query, 0, labels, &datas)
			So(err, ShouldBeNil)
		})
		Convey("range_query, aggregation is value_count, ok", func() {
			dslConfig := range_query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.Aggregation = &interfaces.MetricModelFormulaConfigAggregation{
				Type:  interfaces.AGGR_TYPE_VALUE_COUNT,
				Field: "value",
			}
			range_query.FormulaConfig = dslConfig
			dslRes := []byte(`{
				"aggregations": {
					"index_base": {
						"buckets": [
							{
								"key": "test",
								"doc_count": 12,
								"category": {
									"buckets": [
										{
											"key": "log",
											"doc_count": 12,
											"value": {
												"buckets": {
													"value_1": {
														"doc_count": 1,
														"__date_histogram": {
															"buckets": [
																{
																	"key_as_string": "2024-09-09T00:00:00.000Z",
																	"key": 1725840000000,
																	"doc_count": 1,
																	"__value": {
																		"value": 1
																	}
																}
															]
														}
													},
													"__other": {
														"doc_count": 11,
														"__date_histogram": {
															"buckets": [
																{
																	"key_as_string": "2024-09-09T00:00:00.000Z",
																	"key": 1725840000000,
																	"doc_count": 11,
																	"__value": {
																		"value": 11
																	}
																}
															]
														}
													}
												}
											}
										}
									]
								}
							}
						]
					}
				}
			}`)
			rootNode, _ := sonic.Get(dslRes)
			aggrNode := rootNode.Get("aggregations")
			err := TraversalBucket(aggrNode, range_query, 0, labels, &datas)
			So(err, ShouldBeNil)
		})
		Convey("instant_query, aggregation is percentiles, ok", func() {
			dslConfig := instant_query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.Aggregation = &interfaces.MetricModelFormulaConfigAggregation{
				Type:     interfaces.AGGR_TYPE_PERCENTILES,
				Field:    "value",
				Percents: []float64{10, 20, 50},
			}
			instant_query.FormulaConfig = dslConfig
			dslRes := []byte(`{
    		"aggregations": {
        	"index_base": {
						"buckets": [
							{
								"key": "test",
								"doc_count": 12,
								"category": {
									"buckets": [
										{
											"key": "log",
											"doc_count": 12,
											"value": {
												"buckets": {
													"value_1": {
														"doc_count": 1,
														"__value": {
															"values": {
																"10.0": 1.0,
																"20.0": 1.0,
																"50.0": 1.0
															}
														}
													},
													"__other": {
														"doc_count": 11,
														"__value": {
															"values": {
																"10.0": 78.60000000000001,
																"20.0": 900.7000000000002,
																"50.0": 1234567.0
															}
														}
													}
												}
											}
										}
									]
								}
							}
            ]
        	}
    		}
			}`)
			rootNode, _ := sonic.Get(dslRes)
			aggrNode := rootNode.Get("aggregations")
			err := TraversalBucket(aggrNode, instant_query, 0, labels, &datas)
			So(err, ShouldBeNil)
		})
		Convey("IsInstantQuery=false, aggregation is percentiles, ok", func() {
			dslConfig := range_query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
			dslConfig.Aggregation = &interfaces.MetricModelFormulaConfigAggregation{
				Type:     interfaces.AGGR_TYPE_PERCENTILES,
				Field:    "value",
				Percents: []float64{10, 20, 50},
			}
			range_query.FormulaConfig = dslConfig
			dslRes := []byte(`{
    		"aggregations": {
					"index_base": {
						"buckets": [
							{
								"key": "test",
								"doc_count": 12,
								"category": {
									"buckets": [
										{
											"key": "log",
											"doc_count": 12,
											"value": {
												"buckets": {
													"value_1": {
														"doc_count": 1,
														"__date_histogram": {
															"buckets": [
																{
																	"key_as_string": "2024-09-09T00:00:00.000Z",
																	"key": 1725840000000,
																	"doc_count": 1,
																	"__value": {
																		"values": {
																			"10.0": 1.0,
																			"20.0": 1.0,
																			"50.0": 1.0
																		}
																	}
																}
															]
														}
													},
													"__other": {
														"doc_count": 11,
														"__date_histogram": {
															"buckets": [
																{
																	"key_as_string": "2024-09-09T00:00:00.000Z",
																	"key": 1725840000000,
																	"doc_count": 11,
																	"__value": {
																		"values": {
																			"10.0": 78.60000000000001,
																			"20.0": 900.7000000000002,
																			"50.0": 1234567.0
																		}
																	}
																}
															]
														}
													}
												}
											}
										}
									]
								}
							}
						]
					}
				}
			}`)
			rootNode, _ := sonic.Get(dslRes)
			aggrNode := rootNode.Get("aggregations")
			err := TraversalBucket(aggrNode, range_query, 0, labels, &datas)
			So(err, ShouldBeNil)
		})
	})
}

func Test_getPointTimeIndex(t *testing.T) {

	Convey("Test getPointTimeIndex", t, func() {

		Convey("minute", func() {
			query := interfaces.MetricModelQuery{
				IsCalendar: true,
				QueryTimeParams: interfaces.QueryTimeParams{
					StepStr: &minute,
				},
			}
			t1, _ := time.Parse(time.RFC3339, "2024-09-11T00:00:00+08:00")
			t2, _ := time.Parse(time.RFC3339, "2024-09-11T00:01:00+08:00")
			startTime := t1.UnixMilli()
			currentTime := t2.UnixMilli()

			pointTimeIndex := getPointTimeIndex(query, startTime, currentTime)
			So(pointTimeIndex, ShouldEqual, 1)
		})

		Convey("hour", func() {
			query := interfaces.MetricModelQuery{
				IsCalendar: true,
				QueryTimeParams: interfaces.QueryTimeParams{
					StepStr: &hour,
				},
			}
			t1, _ := time.Parse(time.RFC3339, "2024-09-11T00:00:00+08:00")
			t2, _ := time.Parse(time.RFC3339, "2024-09-11T01:00:00+08:00")
			startTime := t1.UnixMilli()
			currentTime := t2.UnixMilli()

			pointTimeIndex := getPointTimeIndex(query, startTime, currentTime)
			So(pointTimeIndex, ShouldEqual, 1)
		})

		Convey("day", func() {
			query := interfaces.MetricModelQuery{
				IsCalendar: true,
				QueryTimeParams: interfaces.QueryTimeParams{
					StepStr: &day,
				},
			}
			t1, _ := time.Parse(time.RFC3339, "2024-09-11T00:00:00+08:00")
			t2, _ := time.Parse(time.RFC3339, "2024-09-12T00:00:00+08:00")
			startTime := t1.UnixMilli()
			currentTime := t2.UnixMilli()

			pointTimeIndex := getPointTimeIndex(query, startTime, currentTime)
			So(pointTimeIndex, ShouldEqual, 1)
		})

		Convey("week", func() {
			query := interfaces.MetricModelQuery{
				IsCalendar: true,
				QueryTimeParams: interfaces.QueryTimeParams{
					StepStr: &setp_week,
				},
			}
			t1, _ := time.Parse(time.RFC3339, "2024-09-11T00:00:00+08:00")
			t2, _ := time.Parse(time.RFC3339, "2024-09-18T00:00:00+08:00")
			startTime := t1.UnixMilli()
			currentTime := t2.UnixMilli()

			pointTimeIndex := getPointTimeIndex(query, startTime, currentTime)
			So(pointTimeIndex, ShouldEqual, 1)
		})

		Convey("month", func() {
			query := interfaces.MetricModelQuery{
				IsCalendar: true,
				QueryTimeParams: interfaces.QueryTimeParams{
					StepStr: &step_month,
				},
			}
			t1, _ := time.Parse(time.RFC3339, "2024-09-11T00:00:00+08:00")
			t2, _ := time.Parse(time.RFC3339, "2024-10-11T00:00:00+08:00")
			startTime := t1.UnixMilli()
			currentTime := t2.UnixMilli()

			pointTimeIndex := getPointTimeIndex(query, startTime, currentTime)
			So(pointTimeIndex, ShouldEqual, 1)
		})

		Convey("quarter", func() {
			query := interfaces.MetricModelQuery{
				IsCalendar: true,
				QueryTimeParams: interfaces.QueryTimeParams{
					StepStr: &step_quarter,
				},
			}
			t1, _ := time.Parse(time.RFC3339, "2024-09-11T00:00:00+08:00")
			t2, _ := time.Parse(time.RFC3339, "2024-12-11T00:00:00+08:00")
			startTime := t1.UnixMilli()
			currentTime := t2.UnixMilli()

			pointTimeIndex := getPointTimeIndex(query, startTime, currentTime)
			So(pointTimeIndex, ShouldEqual, 1)
		})

		Convey("year", func() {
			query := interfaces.MetricModelQuery{
				IsCalendar: true,
				QueryTimeParams: interfaces.QueryTimeParams{
					StepStr: &step_year,
				},
			}
			t1, _ := time.Parse(time.RFC3339, "2024-09-11T00:00:00+08:00")
			t2, _ := time.Parse(time.RFC3339, "2025-09-11T00:00:00+08:00")
			startTime := t1.UnixMilli()
			currentTime := t2.UnixMilli()

			pointTimeIndex := getPointTimeIndex(query, startTime, currentTime)
			So(pointTimeIndex, ShouldEqual, 1)
		})

		Convey("IsCalendar is false", func() {
			query := interfaces.MetricModelQuery{
				IsCalendar: false,
				QueryTimeParams: interfaces.QueryTimeParams{
					StepStr: &step_1min,
					Step:    &step_60000,
				},
			}
			t1, _ := time.Parse(time.RFC3339, "2024-09-11T00:00:00+08:00")
			t2, _ := time.Parse(time.RFC3339, "2024-09-11T00:01:00+08:00")
			startTime := t1.UnixMilli()
			currentTime := t2.UnixMilli()

			pointTimeIndex := getPointTimeIndex(query, startTime, currentTime)
			So(pointTimeIndex, ShouldEqual, 1)
		})
	})
}

func Test_calculateMonthDifference(t *testing.T) {

	Convey("Test calculateMonthDifference", t, func() {

		Convey("t2 month >= t1 month", func() {
			t1, _ := time.Parse(time.RFC3339, "2023-07-11T00:00:00+08:00")
			t2, _ := time.Parse(time.RFC3339, "2024-09-11T00:00:00+08:00")
			diff := calculateMonthDifference(t1, t2)
			So(diff, ShouldEqual, 14)
		})

		Convey("t2 month < t1 month", func() {
			t1, _ := time.Parse(time.RFC3339, "2023-11-11T00:00:00+08:00")
			t2, _ := time.Parse(time.RFC3339, "2024-09-11T00:00:00+08:00")
			diff := calculateMonthDifference(t1, t2)
			So(diff, ShouldEqual, 10)
		})
	})
}
