// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"net/http"
	"testing"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/robfig/cron/v3"
	. "github.com/smartystreets/goconvey/convey"

	derrors "data-model/errors"
	"data-model/interfaces"
)

func Test_ValidateMetricModel_ValidateMetricModel(t *testing.T) {
	Convey("Test ValidateMetricModel", t, func() {

		Convey("Validate failed, because model name is null", func() {

			_, res := ValidateMetricModel(testCtx, &interfaces.MetricModel{})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_NullParameter_ModelName)
		})

		Convey("Validate failed, because measure name length > 40", func() {
			_, res := ValidateMetricModel(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:   "a",
					MeasureName: "1111111111111111111111111111111111111111111111111111",
				},
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_LengthExceeded_MeasureName)
		})

		Convey("Validate failed, because metric type is null", func() {
			_, res := ValidateMetricModel(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:   "a",
					MeasureName: "__m.a",
				},
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_NullParameter_MetricType)
		})

		Convey("Validate failed, because metric type unsupported", func() {
			_, res := ValidateMetricModel(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:   "a",
					MetricType:  "ato",
					MeasureName: "__m.a",
				},
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_UnsupportMetricType)
		})

		Convey("Validate failed, because data source is null", func() {
			_, res := ValidateMetricModel(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:   "a",
					MetricType:  interfaces.ATOMIC_METRIC,
					MeasureName: "__m.a",
				},
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_NullParameter_DataSource)
		})

		Convey("Validate failed, because data source is is empty", func() {
			_, res := ValidateMetricModel(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:   "a",
					MetricType:  interfaces.ATOMIC_METRIC,
					MeasureName: "__m.a",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.QueryType_SQL,
				},
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_NullParameter_DataSourceID)
		})

		Convey("Validate failed, because query type is null", func() {
			_, res := ValidateMetricModel(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:   "a",
					MetricType:  interfaces.ATOMIC_METRIC,
					MeasureName: "__m.a",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			})

			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_NullParameter_QueryType)
		})

		Convey("Validate failed, because query type unsupport", func() {
			_, res := ValidateMetricModel(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:   "a",
					MetricType:  interfaces.ATOMIC_METRIC,
					QueryType:   "a",
					MeasureName: "__m.a",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			})

			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_UnsupportQueryType)
		})

		// Convey("Validate failed, when query type is sql and data source is not vega", func() {
		// 	_, res := ValidateMetricModel(testCtx, &interfaces.MetricModel{
		// 		SimpleMetricModel: interfaces.SimpleMetricModel{
		// 			ModelName:   "a",
		// 			MetricType:  interfaces.ATOMIC_METRIC,
		// 			QueryType:   interfaces.SQL,
		// 			MeasureName: "__m.a",
		// 		},
		// 		DataSource: &interfaces.MetricDataSource{
		// 			Type: interfaces.QueryType_SQL,
		// 			ID:   "a",
		// 		},
		// 	})
		// 	So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
		// 	So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_DataSourceType)
		// })

		Convey("Validate failed, when query type is sql and formula config is null", func() {
			_, res := ValidateMetricModel(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:   "a",
					MetricType:  interfaces.ATOMIC_METRIC,
					QueryType:   interfaces.SQL,
					MeasureName: "__m.a",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.QueryType_SQL,
					ID:   "a",
				},
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_NullParameter_FormulaConfig)
		})

		Convey("Validate failed, when query type is sql and marshal formula config failed", func() {
			_, res := ValidateMetricModel(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:     "a",
					MetricType:    interfaces.ATOMIC_METRIC,
					QueryType:     interfaces.SQL,
					MeasureName:   "__m.a",
					FormulaConfig: make(chan int),
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.QueryType_SQL,
					ID:   "a",
				},
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_FormulaConfig)
		})

		Convey("Validate failed, when query type is sql and Unmarshal formula config failed", func() {
			_, res := ValidateMetricModel(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:     "a",
					MetricType:    interfaces.ATOMIC_METRIC,
					QueryType:     interfaces.SQL,
					MeasureName:   "__m.a",
					FormulaConfig: "",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.QueryType_SQL,
					ID:   "a",
				},
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_FormulaConfig)
		})

		Convey("Validate failed, when query type is sql and aggregation is empty", func() {
			_, res := ValidateMetricModel(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:     "a",
					MetricType:    interfaces.ATOMIC_METRIC,
					QueryType:     interfaces.SQL,
					MeasureName:   "__m.a",
					FormulaConfig: interfaces.SQLConfig{},
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.QueryType_SQL,
					ID:   "a",
				},
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_NullParameter_SQLAggrExpression)
		})

		Convey("Validate failed, when query type is sql and AggrExpr is not nil but empty && AggrExprStr is empty ", func() {
			_, res := ValidateMetricModel(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:   "a",
					MetricType:  interfaces.ATOMIC_METRIC,
					QueryType:   interfaces.SQL,
					MeasureName: "__m.a",
					FormulaConfig: interfaces.SQLConfig{
						AggrExpr: &interfaces.AggrExpr{},
					},
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.QueryType_SQL,
					ID:   "a",
				},
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_NullParameter_SQLAggrExpression)
		})

		Convey("Validate failed, when query type is sql and AggrExpr is not empty && AggrExprStr is not empty", func() {
			_, res := ValidateMetricModel(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:   "a",
					MetricType:  interfaces.ATOMIC_METRIC,
					QueryType:   interfaces.SQL,
					MeasureName: "__m.a",
					FormulaConfig: interfaces.SQLConfig{
						AggrExpr: &interfaces.AggrExpr{
							Aggr:  "avg",
							Field: "f1",
						},
						AggrExprStr: "avg(a)",
					},
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.QueryType_SQL,
					ID:   "a",
				},
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_SQLAggrExpression)
		})

		Convey("Validate failed, when query type is sql and both Condition and ConditionStr are not empty", func() {
			_, res := ValidateMetricModel(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:   "a",
					MetricType:  interfaces.ATOMIC_METRIC,
					QueryType:   interfaces.SQL,
					MeasureName: "__m.a",
					FormulaConfig: interfaces.SQLConfig{
						AggrExpr:     &interfaces.AggrExpr{},
						AggrExprStr:  "avg(a)",
						ConditionStr: "a=1",
						Condition:    &interfaces.CondCfg{},
					},
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.QueryType_SQL,
					ID:   "a",
				},
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_SQLCondition)
		})

		// Convey("Validate failed, when query type is sql and date field is empty", func() {
		// 	_, res := ValidateMetricModel(testCtx, &interfaces.MetricModel{
		// 		SimpleMetricModel: interfaces.SimpleMetricModel{
		// 			ModelName:   "a",
		// 			MetricType:  interfaces.ATOMIC_METRIC,
		// 			QueryType:   interfaces.SQL,
		// 			MeasureName: "__m.a",
		// 			FormulaConfig: interfaces.SQLConfig{
		// 				AggrExpr: &interfaces.AggrExpr{
		// 					Aggr:  "avg",
		// 					Field: "f1",
		// 				},
		// 				AggrExprStr:  "",
		// 				ConditionStr: "a=1",
		// 			},
		// 		},
		// 		DataSource: &interfaces.MetricDataSource{
		// 			Type: interfaces.DATA_SOURCE_VEGA_LOGIC_VIEW,
		// 			ID:   "a",
		// 		},
		// 	})
		// 	So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
		// 	So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_NullParameter_DateField)
		// })

		Convey("Validate failed, because formula is null", func() {
			_, res := ValidateMetricModel(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:   "a",
					MetricType:  interfaces.ATOMIC_METRIC,
					QueryType:   interfaces.PROMQL,
					MeasureName: "__m.a",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_NullParameter_Formula)
		})

		Convey("Validate failed, because date field of promql is invalid", func() {
			_, res := ValidateMetricModel(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:   "a",
					MetricType:  interfaces.ATOMIC_METRIC,
					QueryType:   interfaces.PROMQL,
					Formula:     "a",
					MeasureName: "__m.a",
					DateField:   "@timestamp1",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_DateField)
		})

		Convey("Validate failed, because metric field of promql is invalid", func() {
			_, res := ValidateMetricModel(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:    "a",
					MetricType:   interfaces.ATOMIC_METRIC,
					QueryType:    interfaces.PROMQL,
					Formula:      "a",
					MeasureName:  "__m.a",
					DateField:    "",
					MeasureField: "a",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_MeasureField)
		})

		Convey("Validate failed, because unit type is null", func() {
			_, res := ValidateMetricModel(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:    "a",
					MetricType:   interfaces.ATOMIC_METRIC,
					QueryType:    interfaces.PROMQL,
					Formula:      "a",
					MeasureName:  "__m.a",
					DateField:    "",
					MeasureField: "",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_NullParameter_UnitType)
		})

		Convey("Validate failed, because unit type is invalid", func() {
			_, res := ValidateMetricModel(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:    "a",
					MetricType:   interfaces.ATOMIC_METRIC,
					QueryType:    interfaces.PROMQL,
					Formula:      "a",
					UnitType:     "a",
					MeasureName:  "__m.a",
					DateField:    interfaces.PROMQL_DATEFIELD,
					MeasureField: interfaces.PROMQL_METRICFIELD,
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_UnitType)
		})

		Convey("Validate failed, because unit type is not numUnit && unit is empty", func() {
			_, res := ValidateMetricModel(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:    "a",
					MetricType:   interfaces.ATOMIC_METRIC,
					QueryType:    interfaces.PROMQL,
					UnitType:     "storeUnit",
					Formula:      "a",
					MeasureName:  "__m.a",
					DateField:    interfaces.PROMQL_DATEFIELD,
					MeasureField: interfaces.PROMQL_METRICFIELD,
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_NullParameter_Unit)
		})

		Convey("Validate failed, because unit is invalid", func() {
			_, res := ValidateMetricModel(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:    "a",
					MetricType:   interfaces.ATOMIC_METRIC,
					QueryType:    interfaces.PROMQL,
					Formula:      "a",
					UnitType:     "storeUnit",
					Unit:         "a",
					MeasureName:  "__m.a",
					DateField:    interfaces.PROMQL_DATEFIELD,
					MeasureField: interfaces.PROMQL_METRICFIELD,
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_UnitType)
		})

		Convey("Validate failed, because tags number > 5", func() {
			_, res := ValidateMetricModel(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:    "a",
					MetricType:   interfaces.ATOMIC_METRIC,
					QueryType:    interfaces.PROMQL,
					Tags:         []string{"a", "b", "c", "d", "e", "f"},
					UnitType:     "storeUnit",
					Unit:         "Byte",
					Formula:      "a",
					MeasureName:  "__m.a",
					DateField:    interfaces.PROMQL_DATEFIELD,
					MeasureField: interfaces.PROMQL_METRICFIELD,
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			})

			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_CountExceeded_TagTotal)
		})

		Convey("Validate failed, because tag name invalid", func() {
			_, res := ValidateMetricModel(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:    "a",
					MetricType:   interfaces.ATOMIC_METRIC,
					QueryType:    interfaces.PROMQL,
					Tags:         []string{"a", "b", "c", "d", "e/"},
					UnitType:     "storeUnit",
					Unit:         "Byte",
					Formula:      "a",
					MeasureName:  "__m.a",
					DateField:    interfaces.PROMQL_DATEFIELD,
					MeasureField: interfaces.PROMQL_METRICFIELD,
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_InvalidParameter_DataTagName)
		})

		Convey("Validate failed, because comment length > 255", func() {
			_, res := ValidateMetricModel(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "a",
					MetricType: interfaces.ATOMIC_METRIC,
					QueryType:  interfaces.PROMQL,
					Tags:       []string{"a", "b", "c", "d", "e"},
					Comment: `aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
				aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
				aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
				aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa`,
					Formula:      "a",
					UnitType:     "storeUnit",
					Unit:         "Byte",
					MeasureName:  "__m.a",
					DateField:    interfaces.PROMQL_DATEFIELD,
					MeasureField: interfaces.PROMQL_METRICFIELD,
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_LengthExceeded_Comment)
		})

		Convey("Validate succeed", func() {
			_, res := ValidateMetricModel(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:    "a",
					MetricType:   interfaces.ATOMIC_METRIC,
					QueryType:    interfaces.PROMQL,
					Tags:         []string{"a", "b", "c", "d", "e"},
					Comment:      `aaaaaaaaaaaaaaaaaaaaa`,
					Formula:      "a",
					UnitType:     "storeUnit",
					Unit:         "Byte",
					MeasureName:  "__m.a",
					DateField:    interfaces.PROMQL_DATEFIELD,
					MeasureField: interfaces.PROMQL_METRICFIELD,
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			})
			So(res, ShouldEqual, nil)
		})
	})
}

func Test_ValidateMetricModel_ValidateDSL(t *testing.T) {

	Convey("Test ValidateDSL", t, func() {

		var containTopHits bool

		Convey("Validate failed, because dsl Unmarshal err", func() {

			res := validateDSL(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "a",
					MetricType: interfaces.ATOMIC_METRIC,
					QueryType:  interfaces.DSL,
					Formula:    `"a": {{__interval}}`,
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			}, &containTopHits)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_Formula)
		})

		Convey("Validate failed, because missing dsl size", func() {

			res := validateDSL(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "a",
					MetricType: interfaces.ATOMIC_METRIC,
					QueryType:  interfaces.DSL,
					Formula:    `{"a": 1}`,
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			}, &containTopHits)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "The size of dsl expected 0, actual is <nil>")
		})

		Convey("Validate failed, because dsl size != 0", func() {

			res := validateDSL(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "a",
					MetricType: interfaces.ATOMIC_METRIC,
					QueryType:  interfaces.DSL,
					Formula:    `{"a": 1,"size":1}`,
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			}, &containTopHits)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "The size of dsl expected 0, actual is 1")
		})

		Convey("Validate failed, because dsl aggs not exists", func() {

			res := validateDSL(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "a",
					MetricType: interfaces.ATOMIC_METRIC,
					QueryType:  interfaces.DSL,
					Formula:    `{"a": 1,"size":0}`,
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			}, &containTopHits)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "dsl missing aggregation.")
		})

		Convey("Validate-validAggs failed, because dsl aggs is not a map", func() {

			res := validateDSL(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "a",
					MetricType: interfaces.ATOMIC_METRIC,
					QueryType:  interfaces.DSL,
					Formula:    `{"aggs": 1,"size":0}`,
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			}, &containTopHits)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "The aggregation of dsl is not a map")
		})

		Convey("Validate-validAggs failed, because dsl aggs contain multi aggs", func() {

			res := validateDSL(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "a",
					MetricType: interfaces.ATOMIC_METRIC,
					QueryType:  interfaces.DSL,
					Formula:    `{"size":0,"aggs":{"terms1":{},"terms2":{}}}`,
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			}, &containTopHits)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "Multiple aggregation is not supported")
		})

		Convey("Validate-validAggs failed, because aggs name are same", func() {

			res := validateDSL(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "a",
					MetricType: interfaces.ATOMIC_METRIC,
					QueryType:  interfaces.DSL,
					Formula: `{"size":0,"aggs": {
						"name1": {
						  "terms": {
							"field": "b",
							"size": 10
						  },
						  "aggs": {
							"name1": {
							  "terms": {
								"field": "a",
								"size": 10
							  }
							}
						  }
						}
					}}`,
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			}, &containTopHits)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "The aggregation names of each layer aggregation cannot be the same")
		})

		Convey("Validate-validAggs failed, because aggs is not a map", func() {

			res := validateDSL(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "a",
					MetricType: interfaces.ATOMIC_METRIC,
					QueryType:  interfaces.DSL,
					Formula: `{"size":0,"aggs": {
						"name1": 0
					}}`,
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			}, &containTopHits)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "The aggregation of dsl is not a map")
		})

		Convey("Validate-validAggs failed, because unsurpport agg type", func() {

			res := validateDSL(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "a",
					MetricType: interfaces.ATOMIC_METRIC,
					QueryType:  interfaces.DSL,
					Formula: `{"size":0,"aggs": {
						"name1": {
						  "histogram": {
							"field": "price",
							"fixed_interval": 50
						  }
						}
					}}`,
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			}, &containTopHits)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "Unsupport aggregation type in dsl.")
		})

		Convey("Validate-validAggs failed, because use interval", func() {

			res := validateDSL(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "a",
					MetricType: interfaces.ATOMIC_METRIC,
					QueryType:  interfaces.DSL,
					Formula: `{"size":0,"aggs": {
						"name1": {
						  "date_histogram": {
							"field": "price",
							"interval": 50
						  }
						}
					}}`,
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			}, &containTopHits)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "The interval has been abandoned")
		})

		Convey("Validate-validAggs failed, because fixed_interval and calendar_interval of dsl cannot exist simultaneously", func() {

			res := validateDSL(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "a",
					MetricType: interfaces.ATOMIC_METRIC,
					QueryType:  interfaces.DSL,
					Formula: `{"size":0,"aggs": {
						"name1": {
						  "date_histogram": {
							"field": "price",
							"fixed_interval": 50,
							"calendar_interval": "1d"
						  }
						}
					}}`,
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			}, &containTopHits)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "The date_histogram aggregation statement is incorrect")
		})

		Convey("Validate-validAggs failed, because Neither fixed_interval nor calendar_interval  exists in the DSL statement", func() {

			res := validateDSL(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "a",
					MetricType: interfaces.ATOMIC_METRIC,
					QueryType:  interfaces.DSL,
					Formula: `{"size":0,"aggs": {
						"name1": {
						  "date_histogram": {
							"field": "price"
						  }
						}
					}}`,
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			}, &containTopHits)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "The date_histogram aggregation statement is incorrect")
		})

		Convey("Validate-validAggs-validTopHits failed, because top hits unmarshal error", func() {

			res := validateDSL(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "a",
					MetricType: interfaces.ATOMIC_METRIC,
					QueryType:  interfaces.DSL,
					Formula: `{"size":0,"aggs": {
						"name1": {
						  "top_hits": {
							"size": "1"
						  }
						}
					}}`,
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			}, &containTopHits)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "TopHits Unmarshal error: Mismatch type int64 with value string \"at index 8: mismatched type with value\\n\\n\\t{\\\"size\\\":\\\"1\\\"}\\n\\t........^...\\n\"")
		})

		Convey("Validate-validAggs-validTopHits failed, because top hits size lte 0 ", func() {

			res := validateDSL(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "a",
					MetricType: interfaces.ATOMIC_METRIC,
					QueryType:  interfaces.DSL,
					Formula: `{"size":0,"aggs": {
						"name1": {
						  "top_hits": {
							"size": -1
						  }
						}
					}}`,
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			}, &containTopHits)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "TopHits's numHits must be > 0")
		})

		Convey("Validate-validAggs-validTopHits failed, because top hits includes is empty ", func() {

			res := validateDSL(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "a",
					MetricType: interfaces.ATOMIC_METRIC,
					QueryType:  interfaces.DSL,
					Formula: `{"size":0,"aggs": {
						"name1": {
						  "top_hits": {
							"size": 1
						  }
						}
					}}`,
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			}, &containTopHits)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "TopHits's includes must not be empty")
		})

		Convey("Validate-validAggs-validTopHits failed, because measure field is not one of top hits includes ", func() {

			res := validateDSL(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "a",
					MetricType: interfaces.ATOMIC_METRIC,
					QueryType:  interfaces.DSL,
					Formula: `{"size":0,"aggs": {
						"name1": {
						  "top_hits": {
							"size": 1,
							"_source": {
								"includes": [ "labels.cpu","metrics.node_cpu_seconds_total" ]
							  }
						  }
						}
					}}`,
					MeasureField: "a",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			}, &containTopHits)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "The measure field must be one of the includes in top_hits")
		})

		Convey("Validate failed, because measure field is not one of top hits includes ", func() {

			res := validateDSL(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "a",
					MetricType: interfaces.ATOMIC_METRIC,
					QueryType:  interfaces.DSL,
					Formula: `{"size":0,"aggs": {
						"name1": {
						  "top_hits": {
							"size": 1,
							"_source": {
								"includes": [ "labels.cpu","metrics.node_cpu_seconds_total" ]
							  }
						  }
						}
					}}`,
					MeasureField: "a",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			}, &containTopHits)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "The measure field must be one of the includes in top_hits")
		})

		Convey("Validate failed, because agg layers gte 7 ", func() {

			res := validateDSL(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "a",
					MetricType: interfaces.ATOMIC_METRIC,
					QueryType:  interfaces.DSL,
					Formula: `{"size":0,"aggs": {
						"name1": {
						  "terms": {
							"field": "2",
							"size": 10
						  },
						  "aggs": {
							"NAME2": {
							  "terms": {
								"field": "2",
								"size": 10
							  },
							  "aggs": {
								"NAME3": {
								  "terms": {
									"field": "3",
									"size": 10
								  },
								  "aggs": {
									"NAME4": {
									  "terms": {
										"field": "2",
										"size": 10
									  },
									  "aggs": {
										"NAME5": {
										  "terms": {
											"field": "4",
											"size": 10
										  },
										  "aggs": {
											"NAME6": {
											  "terms": {
												"field": "5",
												"size": 10
											  },
											  "aggs": {
												"NAME7": {
												  "terms": {
													"field": "3",
													"size": 10
												  },
												  "aggs": {
													"NAME8": {
													  "terms": {
														"field": "4",
														"size": 10
													  }
													}
												  }
												}
											  }
											}
										  }
										}
									  }
									}
								  }
								}
							  }
							}
						  }
						}
					  }}`,
					MeasureField: "a"},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			}, &containTopHits)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "The number of aggregation layers in DSL does not exceed 7.")
		})

		Convey("Validate failed, because missing metric aggregation ", func() {

			res := validateDSL(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "a",
					MetricType: interfaces.ATOMIC_METRIC,
					QueryType:  interfaces.DSL,
					Formula: `{"size":0,"aggs": {
						"name1": {
							"terms": {
							"field": "2",
							"size": 10
							},
							"aggs": {
							"NAME2": {
								"terms": {
								"field": "2",
								"size": 10
								}
							}
							}
						}
						}}`,
					MeasureField: "a",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			}, &containTopHits)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "A metric aggregation should be included in dsl")
		})

		Convey("Validate failed, because measure field is empty ", func() {

			res := validateDSL(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "a",
					MetricType: interfaces.ATOMIC_METRIC,
					QueryType:  interfaces.DSL,
					Formula: `{"size":0,"aggs": {
						"name1": {
						  "terms": {
							"field": "2",
							"size": 10
						  },
						  "aggs": {
							"NAME2": {
							  "terms": {
								"field": "2",
								"size": 10
							  },
							  "aggs": {
								"NAME": {
								  "value_count": {
									"field": "1"
								  }
								}
							  }
							}
						  }
						}
					  }}`,
					MeasureField: "",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			}, &containTopHits)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_NullParameter_MeasureField)
		})

		Convey("Validate failed, because measure field not aggs name ", func() {

			res := validateDSL(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "a",
					MetricType: interfaces.ATOMIC_METRIC,
					QueryType:  interfaces.DSL,
					Formula: `{"size":0,"aggs": {
						"name1": {
						  "terms": {
							"field": "2",
							"size": 10
						  },
						  "aggs": {
							"NAME2": {
							  "terms": {
								"field": "2",
								"size": 10
							  },
							  "aggs": {
								"NAME": {
								  "value_count": {
									"field": "1"
								  }
								}
							  }
							}
						  }
						}
					  }}`,
					MeasureField: "a",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			}, &containTopHits)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_MeasureField)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "Measure Field should be the name of metric aggregation in dsl")
		})

		Convey("Validate failed, because not contain date histogram ", func() {

			res := validateDSL(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "a",
					MetricType: interfaces.ATOMIC_METRIC,
					QueryType:  interfaces.DSL,
					Formula: `{"size":0,"aggs": {
						"name1": {
						  "terms": {
							"field": "2",
							"size": 10
						  },
						  "aggs": {
							"NAME2": {
							  "terms": {
								"field": "2",
								"size": 10
							  },
							  "aggs": {
								"NAME": {
								  "value_count": {
									"field": "1"
								  }
								}
							  }
							}
						  }
						}
					  }}`,
					MeasureField: "NAME",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			}, &containTopHits)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "Only one date_histogram aggregation should be included in dsl")
		})

		Convey("Validate failed, because contain more than one date histogram ", func() {

			res := validateDSL(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "a",
					MetricType: interfaces.ATOMIC_METRIC,
					QueryType:  interfaces.DSL,
					Formula: `{"size":0,"aggs": {
						"name1": {
						  "terms": {
							"field": "2",
							"size": 10
						  },
						  "aggs": {
							"NAME": {
							  "date_histogram": {
								"field": "@timestamp",
								"fixed_interval": "1d"
							  },
							  "aggs": {
								"NAME1": {
								  "date_histogram": {
									"field": "@timestamp",
									"fixed_interval": "1d"
								  },
								  "aggs": {
									"NAMEv": {
									  "sum": {
										"field": "a"
									  }
									}
								  }
								}
							  }
							}
						  }
						}
					  }}`,
					MeasureField: "NAMEv",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			}, &containTopHits)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "Only one date_histogram aggregation should be included in dsl")
		})

		Convey("Validate failed, because date field is not date histogram's name ", func() {

			res := validateDSL(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "a",
					MetricType: interfaces.ATOMIC_METRIC,
					QueryType:  interfaces.DSL,
					Formula: `{"size":0,"aggs": {
						"name1": {
						  "terms": {
							"field": "2",
							"size": 10
						  },
						  "aggs": {
							"NAME": {
							  "date_histogram": {
								"field": "@timestamp",
								"fixed_interval": "1d"
							  },
							  "aggs": {
								"NAME1": {
								  "avg": {
									"field": "a"
								  }
								}
							  }
							}
						  }
						}
					  }}`,
					MeasureField: "NAME1",
					DateField:    "a",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			}, &containTopHits)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_DateField)
		})

		Convey("Validate failed, because there is no metric aggs under the date histogram ", func() {

			res := validateDSL(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "a",
					MetricType: interfaces.ATOMIC_METRIC,
					QueryType:  interfaces.DSL,
					Formula: `{"size":0,"aggs": {
						"cpu2": {
							"terms": {
								"field": "labels.cpu.keyword",
								"size": 10000
							},
							"aggs": {
								"mode": {
									"terms": {
										"field": "labels.mode.keyword",
										"size": 100000
									},
									"aggs": {
										"time2": {
											"date_histogram": {
												"field": "@timestamp",
												"fixed_interval": "5d",
												"format": "yyyy-MM-dd HH:mm:ss.SSS",
												"min_doc_count": 1,
												"order": {
													"_key": "asc"
												}
											},
											"aggs": {
												"value": {
													"terms": {
													  "field": "1",
													  "size": 10
													},
													"aggs": {
													  "NAME": {
														"value_count": {
														  "field": "1"
														}
													  }
													}
												}
											}
										}
									}
								}
							}
						}
					}}`,
					MeasureField: "NAME",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			}, &containTopHits)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "The sub aggregation of date_histogram aggregation needs to be a metric aggregation.")
		})

		Convey("Validate success ", func() {

			res := validateDSL(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "a",
					MetricType: interfaces.ATOMIC_METRIC,
					QueryType:  interfaces.DSL,
					Formula: `{"size":0,"aggs": {
						"NAME1": {
						  "terms": {
							"field": "labels.cpu.keyword",
							"size": 10
						  },
						  "aggs": {
							"NAME2": {
							  "filters": {
								"filters": {
								  "system": {
									"term": {
									  "labels.mode": "system"
									}
								  },
								  "user": {
									"term": {
									  "labels.mode": "user"
									}
								  }
								}
							  },
							  "aggs": {
								"NAME3": {
								  "range": {
									"field": "metrics.node_cpu_seconds_total",
									"ranges": [
									  {
										"from": 0,
										"to": 1000000
									  },
									  {
										"from": 1000000,
										"to": 20000000
									  }
									]
								  },
								  "aggs": {
									"NAME4": {
									  "date_range": {
										"field": "@timestamp",
										"ranges": [
										  {
											"from": "now-5d/d",
											"to": "now"
										  },
										  {
											"from": "now-10d/d",
											"to": "now-5d/d"
										  }
										]
									  },
									  "aggs": {
										"NAME5": {
										  "date_histogram": {
											"field": "@timestamp",
											"fixed_interval": "1d"
										  },
										  "aggs": {
											"NAME6": {
											  "max": {
												"field": "metrics.node_cpu_seconds_total"
											  }
											}
										  }
										}
									  }
									}
								  }
								}
							  }
							}
						  }
						}
					  }}`,
					MeasureField: "NAME6",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			}, &containTopHits)
			So(res, ShouldBeNil)
		})

		Convey("Validate success when interval is calendar_interval", func() {

			res := validateDSL(testCtx, &interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "a",
					MetricType: interfaces.ATOMIC_METRIC,
					QueryType:  interfaces.DSL,
					Formula: `{"size":0,"aggs": {
						"NAME1": {
						  "terms": {
							"field": "labels.cpu.keyword",
							"size": 10
						  },
						  "aggs": {
							"NAME2": {
							  "filters": {
								"filters": {
								  "system": {
									"term": {
									  "labels.mode": "system"
									}
								  },
								  "user": {
									"term": {
									  "labels.mode": "user"
									}
								  }
								}
							  },
							  "aggs": {
								"NAME3": {
								  "range": {
									"field": "metrics.node_cpu_seconds_total",
									"ranges": [
									  {
										"from": 0,
										"to": 1000000
									  },
									  {
										"from": 1000000,
										"to": 20000000
									  }
									]
								  },
								  "aggs": {
									"NAME4": {
									  "date_range": {
										"field": "@timestamp",
										"ranges": [
										  {
											"from": "now-5d/d",
											"to": "now"
										  },
										  {
											"from": "now-10d/d",
											"to": "now-5d/d"
										  }
										]
									  },
									  "aggs": {
										"NAME5": {
										  "date_histogram": {
											"field": "@timestamp",
											"calendar_interval": "day"
										  },
										  "aggs": {
											"NAME6": {
											  "max": {
												"field": "metrics.node_cpu_seconds_total"
											  }
											}
										  }
										}
									  }
									}
								  }
								}
							  }
							}
						  }
						}
					  }}`,
					MeasureField: "NAME6",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			}, &containTopHits)
			So(res, ShouldBeNil)
		})

	})
}

func Test_ValidateMetricModel_Cron2(t *testing.T) {
	// "* * * 1W * ?", "* * * L * ?","* * 1-2 1W * ?","* * 1-2 L * ?","* * 0/1 1W * ?","* * 0/1 L * ?","1-2 * * * * 1L" 不支持
	exprs := []string{
		"* 1/100 * * * ?", "0 0 10,14,16 * * ?",
		"* * * * * ?", "1-2 * * * * ?", "0/1 * * * * ?", "0,1 * * * * ?",
		"* 1-2 * * * ?", "* 1/2 * * * ?", "* 0,1 * * * ?",
		"1-2 1-2 * * * ?", "1-2 1/2 * * * ?", "1-2 0,1,2 * * * ?",
		"0/1 1-2 * * * ?", "0/1 1/2 * * * ?", "0/1 0,1,2 * * * ?",
		"0,1 1-2 * * * ?", "0,1 1/2 * * * ?", "0,1 0,10,11 * * * ?",
		"* * 1-2 * * ?",
		"* * 0/1 * * ?", "* * 0,1,2 * * ?",
		"* 1-2 1-2 * * ?", "* 1-2 0/1 * * ?", "* 1-2 0,1 * * ?",
		"* 1/2 1-2 * * ?", "* 1/2 0/1 * * ?", "* 1/2 0,1 * * ?",
		"* 0,1 1-2 * * ?", "* 0,1 0/1 * * ?", "* 0,1 0,1 * * ?",
		"1-2 * 1-2 * * ?", "1-2 * 0/1 * * ?", "1-2 * 0,1 * * ?",
		"1-2 1-2 1-2 * * ?", "1-2 1-2 0/1 * * ?", "1-2 1-2 0,1 * * ?",
		"1-2 */2 * * * ?", "1-2 1/2 1-2 * * ?", "1-2 1/2 0/1 * * ?", "1-2 1/2 0,1 * * ?",
		"1-2 0,1 * * * ?", "1-2 0,1 1-2 * * ?", "1-2 0,1 0/1 * * ?", "1-2 0,1 0,1 * * ?",
		"0/1 * 1-2 * * ?", "0/1 * 0/1 * * ?", "0/1 * 0,1 * * ?",
		"0/1 1-2 1-2 * * ?", "0/1 1-2 0/1 * * ?", "0/1 1-2 0,1 * * ?",
		"0/1 1/2 1-2 * * ?", "0/1 1/2 0/1 * * ?", "0/1 1/2 0,1 * * ?",
		"0/1 0,1 * * * ?", "0/1 0,1 1-2 * * ?", "0/1 0,1 0/1 * * ?", "0/1 0,1 0,1 * * ?",
		"0,1 * 1-2 * * ?", "0,1 * 0/1 * * ?", "0,1 * 0,1 * * ?",
		"0,1 1-2 1-2 * * ?", "0,1 1-2 0/1 * * ?", "0,1 1-2 0,1 * * ?",
		"0,1 1/2 1-2 * * ?", "0,1 1/2 0/1 * * ?", "0,1 1/2 0,1 * * ?",
		"0,1 0,1 1-2 * * ?", "0,1 0,1 0/1 * * ?", "0,1 0,1 0,1 * * ?",
		"* * * ? * ?", "* * * 1-2 * ?", "* * * 1/1 * ?", "* * * 1,2 * ?",
		"* * 1-2 ? * ?", "* * 1-2 1-2 * ?", "* * 1-2 1/1 * ?", "* * 1-2 1,2 * ?",
		"* * 0/1 ? * ?", "* * 0/1 1-2 * ?", "* * 0/1 1/1 * ?", "* * 0/1 1,2 * ?",
		"* * 0,1 ? * ?", "* * 0,1 1-2 * ?", "* * 0,1 1/1 * ?", "* * 0,1 1,2 * ?",
		"* 1-2 * ? * ?", "* 1-2 * 1-2 * ?", "* 1-2 * 1/1 * ?", "* 1-2 * 1,2 * ?",
		"* 1-2 1-2 ? * ?", "* 1-2 1-2 1-2 * ?", "* 1-2 1-2 1/1 * ?", "* 1-2 1-2 1,2 * ?",
		"* 1-2 0/1 ? * ?", "* 1-2 0/1 1-2 * ?", "* 1-2 0/1 1/1 * ?", "* 1-2 0/1 1,2 * ?",
		"* 1-2 0,1 ? * ?", "* 1-2 0,1 1-2 * ?", "* 1-2 0,1 1/1 * ?", "* 1-2 0,1 1,2 * ?",
		"* 1/2 * ? * ?", "* 1/2 * 1-2 * ?", "* 1/2 * 1/1 * ?", "* 1/2 * 1,2 * ?",
		"* 1/2 1-2 ? * ?", "* 1/2 1-2 1-2 * ?", "* 1/2 1-2 1/1 * ?", "* 1/2 1-2 1,2 * ?",
		"* 1/2 0/1 ? * ?", "* 1/2 0/1 1-2 * ?", "* 1/2 0/1 1/1 * ?", "* 1/2 0/1 1,2 * ?",
		"* 1/2 0,1 ? * ?", "* 1/2 0,1 1-2 * ?", "* 1/2 0,1 1/1 * ?", "* 1/2 0,1 1,2 * ?",
		"* 0,1 * ? * ?", "* 0,1 * 1-2 * ?", "* 0,1 * 1/1 * ?", "* 0,1 * 1,2 * ?",
		"* 0,1 1-2 ? * ?", "* 0,1 1-2 1-2 * ?", "* 0,1 1-2 1/1 * ?", "* 0,1 1-2 1,2 * ?",
		"* 0,1 0/1 ? * ?", "* 0,1 0/1 1-2 * ?", "* 0,1 0/1 1/1 * ?", "* 0,1 0/1 1,2 * ?",
		"* 0,1 0,1 ? * ?", "* 0,1 0,1 1-2 * ?", "* 0,1 0,1 1/1 * ?", "* 0,1 0,1 1,2 * ?",
		"1-2 * * ? * ?", "1-2 * * 1-2 * ?", "1-2 * * 1/1 * ?", "1-2 * * 1,2 * ?",
		"1-2 * * * ? ?", "1-2 * * * 1-2 ?", "1-2 * * * 1/1 ?", "1-2 * * * 1,2 ?",
		"1-2 * * * * ?", "1-2 * * * * 1-2", "1-2 * * * * 1/1", "1-2 * * * * 1,2",
		"1-2 * * * * ?",
	}
	// cronParser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	secondParser := cron.NewParser(cron.Second | cron.Minute |
		cron.Hour | cron.Dom | cron.Month | cron.DowOptional | cron.Descriptor)

	// cronParser := cron.New(cron.WithSeconds())
	Convey("Test robfig/cron", t, func() {
		for _, expr := range exprs {
			_, err := secondParser.Parse(expr)
			So(err, ShouldBeNil)
			// cron.Parse(expr)
		}
	})

	// Convey("Test jakecron", t, func() {
	// 	for _, expr := range exprs {
	// 		// _, err := secondParser.Parse(expr)
	// 		err := ParseSpec(expr)
	// 		So(err, ShouldBeNil)
	// 		// cron.Parse(expr)
	// 	}
	// })

	// Convey("Test gorhill/cronexpr", t, func() {
	// 	// 如果只有6个字段，则会预加一个0秒字段。与xxljob中的表达式不匹配，xxljob中第一个是秒。这样表达式就差了一位.
	// 	// 此库的表达式的范围是正确校验的
	// 	for _, expr := range exprs {
	// 		_, err := cronexpr.Parse(expr)
	// 		So(err, ShouldBeNil)
	// 		// cron.Parse(expr)
	// 	}
	// })
}
