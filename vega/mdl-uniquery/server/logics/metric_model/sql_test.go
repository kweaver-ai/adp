// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package metric_model

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	"uniquery/common"
	"uniquery/common/condition"
	"uniquery/common/value_opt"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
)

var (
	day    = interfaces.CALENDAR_STEP_DAY
	hour   = interfaces.CALENDAR_STEP_HOUR
	minute = interfaces.CALENDAR_STEP_MINUTE
	start  = int64(1706716800000)
	end    = int64(1706716920000)

//	vegaViewFields2 = interfaces.VegaViewWithFields{
//		Fields: []interfaces.VegaViewField{
//			{
//				Name: "f1",
//				Type: "varchar",
//			},
//			{
//				Name: "f2",
//				Type: "varchar",
//			},
//			{
//				Name: "f3",
//				Type: "date",
//			},
//			{
//				Name: "f4",
//				Type: "float",
//			},
//		},
//		VegaFieldMap: map[string]interfaces.VegaViewField{
//			"f1": {
//				Name: "f1",
//				Type: "varchar",
//			},
//			"f2": {
//				Name: "f2",
//				Type: "varchar",
//			},
//			"f3": {
//				Name: "f3",
//				Type: "date",
//			},
//			"f4": {
//				Name: "f4",
//				Type: "float",
//			},
//		},
//	}
)

func TestGenerateSQL(t *testing.T) {
	Convey("Test generateSQL", t, func() {
		sqlMetric := interfaces.MetricModel{
			QueryType: interfaces.SQL,
		}
		Convey("generateSQL failed, because new condition error ", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusBadRequest,
				Language: "en-US",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_DataView_InvalidParameter_Filters,
					Description:  "Invalid Filters",
					Solution:     "Please check whether the parameter is correct.",
					ErrorLink:    "None",
					ErrorDetails: "New condition failed, condition config field name '' must in view original fields",
				},
			}

			_, httpErr := generateSQL(testENCtx, interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.SQL,
				QueryTimeParams: interfaces.QueryTimeParams{
					StepStr: &day,
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.QueryType_SQL,
					ID:   "1",
				},
				FormulaConfig: interfaces.SQLConfig{
					Condition: &condition.CondCfg{
						Operation: "eq",
						ValueOptCfg: value_opt.ValueOptCfg{
							ValueFrom: "const",
							Value:     "1",
						},
					},
				},
			}, &interfaces.DataView{}, interfaces.ViewQuery4Metric{}, sqlMetric)
			So(httpErr, ShouldResemble, expectedErr)
		})

		// Convey("generateSQL failed, because transFilter2SQL error ", func() {
		// 	expectedErr := &rest.HTTPError{
		// 		HTTPCode: http.StatusBadRequest,
		// 		Language: "en-US",
		// 		BaseError: rest.BaseError{
		// 			ErrorCode:    uerrors.Uniquery_DataView_InvalidParameter_Filters,
		// 			Description:  "Invalid Filters",
		// 			Solution:     "Please check whether the parameter is correct.",
		// 			ErrorLink:    "None",
		// 			ErrorDetails: "filter field[f5] not exists in vega view[]",
		// 		},
		// 	}

		// 	_, httpErr := generateSQL(testENCtx, interfaces.MetricModelQuery{
		// 		MetricType: interfaces.ATOMIC_METRIC,
		// 		QueryType:  interfaces.SQL,
		// 		QueryTimeParams: interfaces.QueryTimeParams{
		// 			StepStr: "day",
		// 		},
		// 		DataSource: &interfaces.MetricDataSource{
		// 			Type: interfaces.QueryType_SQL,
		// 			ID:   "1",
		// 		},
		// 		FormulaConfig: interfaces.SQLConfig{
		// 			Condition: &condition.CondCfg{
		// 				Operation: "==",
		// 				ValueOptCfg: value_opt.ValueOptCfg{
		// 					ValueFrom: "const",
		// 					Value:     "1",
		// 				},
		// 				Name: "f1",
		// 			},
		// 		},
		// 		Filters: []interfaces.Filter{
		// 			{
		// 				Operation: "eq",
		// 				Name:      "f5",
		// 				Value:     "1",
		// 			},
		// 		},
		// 	})
		// 	So(httpErr, ShouldResemble, expectedErr)
		// })

		// Convey("generateSQL success", func() {

		// 	_, httpErr := generateSQL(testENCtx, interfaces.MetricModelQuery{
		// 		MetricType: interfaces.ATOMIC_METRIC,
		// 		QueryType:  interfaces.SQL,
		// 		QueryTimeParams: interfaces.QueryTimeParams{
		// 			StepStr: interfaces.CALENDAR_STEP_MINUTE,
		// 		},
		// 		DataSource: &interfaces.MetricDataSource{
		// 			Type: interfaces.QueryType_SQL,
		// 			ID:   "1",
		// 		},
		// 		FormulaConfig: interfaces.SQLConfig{
		// 			Condition: &condition.CondCfg{
		// 				Operation: "==",
		// 				ValueOptCfg: value_opt.ValueOptCfg{
		// 					ValueFrom: "const",
		// 					Value:     "1",
		// 				},
		// 				Name: "f1",
		// 			},
		// 			GroupByFields: []string{"f1"},
		// 			AggrExpr: &interfaces.AggrExpr{
		// 				Aggr:  "avg",
		// 				Field: "f4",
		// 			},
		// 		},
		// 		Filters: []interfaces.Filter{
		// 			{
		// 				Operation: interfaces.OPERATION_EQ,
		// 				Name:      "f1",
		// 				Value:     "1",
		// 			},
		// 			{
		// 				Operation: condition.OperationRange,
		// 				Name:      "f2",
		// 				Value:     "1",
		// 			},
		// 			{
		// 				Operation: condition.OperationOutRange,
		// 				Name:      "f3",
		// 				Value:     "1",
		// 			},
		// 		},
		// 		DateField:      "f3",
		// 		AnalysisDims:   []string{"f1", "f2"},
		// 		IsModelRequest: true,
		// 	})
		// 	So(httpErr, ShouldBeNil)
		// })

	})
}

func TestParseVegaResult2Uniresponse(t *testing.T) {
	Convey("Test generateSQL", t, func() {
		ctx := context.Background()
		loc, _ := time.LoadLocation(os.Getenv("TZ"))
		common.APP_LOCATION = loc

		sqlMetric := interfaces.MetricModel{
			QueryType: interfaces.SQL,
		}

		Convey("failed with instant query when getProportionTotal error", func() {
			errE := rest.NewHTTPError(ctx, http.StatusInternalServerError,
				uerrors.Uniquery_InternalError_AssertFloat64Failed).
				WithErrorDetails("err: strconv.ParseFloat: parsing \"abc\": invalid syntax, vega metric value is abc")

			vegaData := interfaces.VegaFetchData{
				Data: [][]any{
					{"a", "2024-02-01", 2.2},
					{"b", "2024-02-01", "abc"},
				},
				Columns: []condition.ViewField{
					{
						Name: "f1",
						Type: "varchar",
					},
					{
						Name: "__time",
						Type: "varchar",
					},
					{
						Name: "__value",
						Type: "float",
					},
				},
			}
			samePeriodDatas := interfaces.VegaFetchData{}
			query := interfaces.MetricModelQuery{
				QueryTimeParams: interfaces.QueryTimeParams{
					IsInstantQuery: true,
				},
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.SQL,
				FormulaConfig: interfaces.SQLConfig{
					GroupByFields: []string{"dim1"},
					AggrExpr: &interfaces.AggrExpr{
						Aggr:  "sum",
						Field: "metric1",
					},
				},
				AnalysisDims: []string{"dim1"},
				RequestMetrics: &interfaces.RequestMetrics{
					Type: interfaces.METRICS_PROPORTION,
				},
			}
			vegaDuration := int64(123)

			_, err := parseVegaResult2Uniresponse(ctx, vegaData, samePeriodDatas, query, vegaDuration, sqlMetric)
			So(err, ShouldResemble, errE)

		})

		Convey("success with instant query when Proportion", func() {
			vegaData := interfaces.VegaFetchData{
				Data: [][]any{
					{"a", "2024-02-01", 2.2},
					{"b", "2024-02-01", nil},
				},
				Columns: []condition.ViewField{
					{
						Name: "f1",
						Type: "varchar",
					},
					{
						Name: "__time",
						Type: "varchar",
					},
					{
						Name: "__value",
						Type: "float",
					},
				},
			}
			samePeriodDatas := interfaces.VegaFetchData{}
			query := interfaces.MetricModelQuery{
				QueryTimeParams: interfaces.QueryTimeParams{
					IsInstantQuery: true,
				},
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.SQL,
				FormulaConfig: interfaces.SQLConfig{
					GroupByFields: []string{"dim1"},
					AggrExpr: &interfaces.AggrExpr{
						Aggr:  "sum",
						Field: "metric1",
					},
				},
				AnalysisDims: []string{"dim1"},
				RequestMetrics: &interfaces.RequestMetrics{
					Type: interfaces.METRICS_PROPORTION,
				},
			}
			vegaDuration := int64(123)

			res, err := parseVegaResult2Uniresponse(ctx, vegaData, samePeriodDatas, query, vegaDuration, sqlMetric)
			So(err, ShouldBeNil)
			So(len(res.Datas), ShouldEqual, 2)
			So(res.Datas[0].Proportions, ShouldResemble, []any{float64(100)})
			So(res.Datas[1].Proportions, ShouldResemble, []any{float64(0)})
		})

		Convey("failed with instant query when getFloat64Value error", func() {
			errE := rest.NewHTTPError(ctx, http.StatusInternalServerError,
				uerrors.Uniquery_InternalError_AssertFloat64Failed).
				WithErrorDetails("err: strconv.ParseFloat: parsing \"abc\": invalid syntax, vega metric value is abc")

			vegaData := interfaces.VegaFetchData{
				Data: [][]any{
					{"a", "2024-02-01", 2.2},
					{"b", "2024-02-01", "abc"},
				},
				Columns: []condition.ViewField{
					{
						Name: "f1",
						Type: "varchar",
					},
					{
						Name: "__time",
						Type: "varchar",
					},
					{
						Name: "__value",
						Type: "float",
					},
				},
			}
			samePeriodDatas := interfaces.VegaFetchData{Data: [][]any{
				{"a", "2023-02-01", 2.2},
				{"b", "2023-02-01", "abc"},
			},
				Columns: []condition.ViewField{
					{
						Name: "f1",
						Type: "varchar",
					},
					{
						Name: "__time",
						Type: "varchar",
					},
					{
						Name: "__value",
						Type: "float",
					},
				}}
			query := interfaces.MetricModelQuery{
				QueryTimeParams: interfaces.QueryTimeParams{
					IsInstantQuery: true,
				},
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.SQL,
				FormulaConfig: interfaces.SQLConfig{
					GroupByFields: []string{"dim1"},
					AggrExpr: &interfaces.AggrExpr{
						Aggr:  "sum",
						Field: "metric1",
					},
				},
				AnalysisDims: []string{"dim1"},
				RequestMetrics: &interfaces.RequestMetrics{
					Type: interfaces.METRICS_SAMEPERIOD,
					SamePeriodCfg: &interfaces.SamePeriodCfg{
						Method: []string{interfaces.METRICS_SAMEPERIOD_METHOD_GROWTH_RATE,
							interfaces.METRICS_SAMEPERIOD_METHOD_GROWTH_VALUE},
						Offset:          1,
						TimeGranularity: "year",
					},
				},
			}
			vegaDuration := int64(123)

			_, err := parseVegaResult2Uniresponse(ctx, vegaData, samePeriodDatas, query, vegaDuration, sqlMetric)
			So(err, ShouldResemble, errE)

		})

		Convey("success with instant query when METRICS_SAMEPERIOD", func() {
			vegaData := interfaces.VegaFetchData{
				Data: [][]any{
					{"a", "2024-02-01", 2.2},
					{"b", "2024-02-01", 2.3},
					{"c", "2024-02-01", 3.3},
				},
				Columns: []condition.ViewField{
					{
						Name: "f1",
						Type: "varchar",
					},
					{
						Name: "__time",
						Type: "varchar",
					},
					{
						Name: "__value",
						Type: "float",
					},
				},
			}
			samePeriodDatas := interfaces.VegaFetchData{Data: [][]any{
				{"a", "2024-02-01", 2.2},
				{"b", "2024-02-01", nil},
			},
				Columns: []condition.ViewField{
					{
						Name: "f1",
						Type: "varchar",
					},
					{
						Name: "__time",
						Type: "varchar",
					},
					{
						Name: "__value",
						Type: "float",
					},
				}}
			query := interfaces.MetricModelQuery{
				QueryTimeParams: interfaces.QueryTimeParams{
					IsInstantQuery: true,
				},
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.SQL,
				FormulaConfig: interfaces.SQLConfig{
					GroupByFields: []string{"dim1"},
					AggrExpr: &interfaces.AggrExpr{
						Aggr:  "sum",
						Field: "metric1",
					},
				},
				AnalysisDims: []string{"dim1"},
				RequestMetrics: &interfaces.RequestMetrics{
					Type: interfaces.METRICS_SAMEPERIOD,
					SamePeriodCfg: &interfaces.SamePeriodCfg{
						Method: []string{interfaces.METRICS_SAMEPERIOD_METHOD_GROWTH_RATE,
							interfaces.METRICS_SAMEPERIOD_METHOD_GROWTH_VALUE},
						Offset:          1,
						TimeGranularity: "year",
					},
				},
			}
			vegaDuration := int64(123)

			res, err := parseVegaResult2Uniresponse(ctx, vegaData, samePeriodDatas, query, vegaDuration, sqlMetric)
			So(err, ShouldBeNil)
			So(res.Datas[0].GrowthValues, ShouldResemble, []any{float64(0)})
			So(res.Datas[0].GrowthRates, ShouldResemble, []any{float64(0)})
			So(res.Datas[1].GrowthValues, ShouldResemble, []any{float64(2.3)})
			So(res.Datas[1].GrowthRates, ShouldResemble, []any{nil})
			So(res.Datas[2].GrowthValues, ShouldResemble, []any{nil})
			So(res.Datas[2].GrowthRates, ShouldResemble, []any{nil})
		})
	})
}

func TestConvertVegaDatas2TimeSeries(t *testing.T) {
	Convey("Test convertVegaDatas2TimeSeries", t, func() {
		ctx := context.Background()
		loc, _ := time.LoadLocation(os.Getenv("TZ"))
		common.APP_LOCATION = loc

		Convey("should convert vega data to time series correctly for normal case", func() {
			vegaData := interfaces.VegaFetchData{
				Data: [][]any{
					{"a", "2024-02-01 00:00", 1.1},
					{"a", "2024-02-01 00:01", 2.2},
					{"b", "2024-02-01 00:00", 3.3},
				},
				Columns: []condition.ViewField{
					{Name: "dim1", Type: "varchar"},
					{Name: "__time", Type: "varchar"},
					{Name: "__value", Type: "float"},
				},
			}
			samePeriodDatas := interfaces.VegaFetchData{
				Data: [][]any{
					{"a", "2023-02-01 00:00", 0.5},
					{"a", "2023-02-01 00:01", 1.0},
					{"b", "2023-02-01 00:00", 2.0},
				},
				Columns: []condition.ViewField{
					{Name: "dim1", Type: "varchar"},
					{Name: "__time", Type: "varchar"},
					{Name: "__value", Type: "float"},
				},
			}
			// metricModel := interfaces.MetricModel{}
			query := interfaces.MetricModelQuery{
				QueryTimeParams: interfaces.QueryTimeParams{
					StepStr:        &minute,
					IsInstantQuery: false,
					Start:          &start,
					End:            &end,
				},
				IsCalendar: true,
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.SQL,
				FormulaConfig: interfaces.SQLConfig{
					GroupByFields: []string{"dim1"},
					AggrExpr: &interfaces.AggrExpr{
						Aggr:  "sum",
						Field: "metric1",
					},
				},
				AnalysisDims: []string{"dim1"},
			}
			vegaDuration := int64(100)
			resp, err := convertVegaDatas2TimeSeries(ctx, vegaData, samePeriodDatas, query, vegaDuration)
			So(err, ShouldBeNil)
			So(resp.Step, ShouldEqual, &minute)
			So(resp.VegaDurationMs, ShouldEqual, vegaDuration)
			So(len(resp.Datas), ShouldEqual, 2) // two series: "a" and "b"
			So(resp.SeriesTotal, ShouldEqual, 2)
			// Check time series for "a"
			var foundA bool
			for _, ts := range resp.Datas {
				if ts.Labels["dim1"] == "a" {
					foundA = true
					So(ts.TimeStrs, ShouldResemble, []string{"2024-02-01 00:00", "2024-02-01 00:01"})
					So(ts.Values, ShouldResemble, []any{1.1, 2.2})
				}
			}
			So(foundA, ShouldBeTrue)
		})

		Convey("should handle METRICS_PROPORTION type", func() {
			vegaData := interfaces.VegaFetchData{
				Data: [][]any{
					{"a", "2024-02-01 00", 2.0},
					{"b", "2024-02-01 00", 3.0},
					{"c", "2024-02-01 00", nil},
				},
				Columns: []condition.ViewField{
					{Name: "dim1", Type: "varchar"},
					{Name: "__time", Type: "varchar"},
					{Name: "__value", Type: "float"},
				},
			}
			samePeriodDatas := interfaces.VegaFetchData{}
			// metricModel := interfaces.MetricModel{}
			query := interfaces.MetricModelQuery{
				QueryTimeParams: interfaces.QueryTimeParams{
					StepStr:        &hour,
					IsInstantQuery: false,
					Start:          &start,
					End:            &end,
				},
				IsCalendar: true,
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.SQL,
				FormulaConfig: interfaces.SQLConfig{
					GroupByFields: []string{"dim1"},
					AggrExpr: &interfaces.AggrExpr{
						Aggr:  "sum",
						Field: "metric1",
					},
				},
				AnalysisDims: []string{"dim1"},
				RequestMetrics: &interfaces.RequestMetrics{
					Type: interfaces.METRICS_PROPORTION,
				},
			}
			vegaDuration := int64(200)
			resp, err := convertVegaDatas2TimeSeries(ctx, vegaData, samePeriodDatas, query, vegaDuration)
			So(err, ShouldBeNil)
			So(len(resp.Datas), ShouldEqual, 3)
			for _, ts := range resp.Datas {
				So(len(ts.Proportions), ShouldEqual, 1)
			}
			// proportions: a=2/5*100=40, b=3/5*100=60
			for _, ts := range resp.Datas {
				if ts.Labels["dim1"] == "a" {
					So(ts.Proportions[0], ShouldAlmostEqual, 40)
				}
				if ts.Labels["dim1"] == "b" {
					So(ts.Proportions[0], ShouldAlmostEqual, 60)
				}
				if ts.Labels["dim1"] == "c" {
					So(ts.Proportions[0], ShouldAlmostEqual, float64(0))
				}
			}
		})

		Convey("convert2TimeSeries samePeriodDatas error", func() {
			vegaData := interfaces.VegaFetchData{
				Data: [][]any{
					{"a", "2024-02-01", 5.0},
					{"a", "2024-02-02", 10.0},
				},
				Columns: []condition.ViewField{
					{Name: "dim1", Type: "varchar"},
					{Name: "__time", Type: "varchar"},
					{Name: "__value", Type: "float"},
				},
			}
			samePeriodDatas := interfaces.VegaFetchData{
				Data: [][]any{
					{"a", "2023-02-01", 2.0},
					{"a", "2023-02-02", "a"},
				},
				Columns: []condition.ViewField{
					{Name: "dim1", Type: "varchar"},
					{Name: "__time", Type: "varchar"},
					{Name: "__value", Type: "float"},
				},
			}
			// metricModel := interfaces.MetricModel{}
			query := interfaces.MetricModelQuery{
				QueryTimeParams: interfaces.QueryTimeParams{
					StepStr:        &day,
					IsInstantQuery: false,
					Start:          &start,
					End:            &end,
				},
				IsCalendar: true,
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.SQL,
				FormulaConfig: interfaces.SQLConfig{
					GroupByFields: []string{"dim1"},
					AggrExpr: &interfaces.AggrExpr{
						Aggr:  "sum",
						Field: "metric1",
					},
				},
				AnalysisDims: []string{"dim1"},
				RequestMetrics: &interfaces.RequestMetrics{
					Type: interfaces.METRICS_SAMEPERIOD,
					SamePeriodCfg: &interfaces.SamePeriodCfg{
						Method: []string{interfaces.METRICS_SAMEPERIOD_METHOD_GROWTH_RATE,
							interfaces.METRICS_SAMEPERIOD_METHOD_GROWTH_VALUE},
						Offset:          1,
						TimeGranularity: "year",
					},
				},
			}
			vegaDuration := int64(300)
			_, err := convertVegaDatas2TimeSeries(ctx, vegaData, samePeriodDatas, query, vegaDuration)
			So(err, ShouldResemble, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				uerrors.Uniquery_InternalError_AssertFloat64Failed).
				WithErrorDetails("err: strconv.ParseFloat: parsing \"a\": invalid syntax, vega metric value is a"))
		})

		Convey("should handle METRICS_SAMEPERIOD type", func() {
			vegaData := interfaces.VegaFetchData{
				Data: [][]any{
					{"a", "2024-02-01", 5.0},
					{"a", "2024-02-02", 10.0},
				},
				Columns: []condition.ViewField{
					{Name: "dim1", Type: "varchar"},
					{Name: "__time", Type: "varchar"},
					{Name: "__value", Type: "float"},
				},
			}
			samePeriodDatas := interfaces.VegaFetchData{
				Data: [][]any{
					{"a", "2023-02-01", 2.0},
					{"a", "2023-02-02", 0},
				},
				Columns: []condition.ViewField{
					{Name: "dim1", Type: "varchar"},
					{Name: "__time", Type: "varchar"},
					{Name: "__value", Type: "float"},
				},
			}
			// metricModel := interfaces.MetricModel{}
			query := interfaces.MetricModelQuery{
				QueryTimeParams: interfaces.QueryTimeParams{
					StepStr:        &day,
					IsInstantQuery: false,
					Start:          &start,
					End:            &end,
				},
				IsCalendar: true,
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.SQL,
				FormulaConfig: interfaces.SQLConfig{
					GroupByFields: []string{"dim1"},
					AggrExpr: &interfaces.AggrExpr{
						Aggr:  "sum",
						Field: "metric1",
					},
				},
				AnalysisDims: []string{"dim1"},
				RequestMetrics: &interfaces.RequestMetrics{
					Type: interfaces.METRICS_SAMEPERIOD,
					SamePeriodCfg: &interfaces.SamePeriodCfg{
						Method: []string{interfaces.METRICS_SAMEPERIOD_METHOD_GROWTH_RATE,
							interfaces.METRICS_SAMEPERIOD_METHOD_GROWTH_VALUE},
						Offset:          1,
						TimeGranularity: "year",
					},
				},
			}
			vegaDuration := int64(300)
			resp, err := convertVegaDatas2TimeSeries(ctx, vegaData, samePeriodDatas, query, vegaDuration)
			So(err, ShouldBeNil)
			So(len(resp.Datas), ShouldEqual, 1)
			ts := resp.Datas[0]
			So(ts.Labels["dim1"], ShouldEqual, "a")
			So(ts.GrowthValues, ShouldResemble, []any{float64(3), nil})
			So(ts.GrowthRates, ShouldResemble, []any{float64(150), nil})
		})

		Convey("should handle METRICS_SAMEPERIOD type and fillnull", func() {
			vegaData := interfaces.VegaFetchData{
				Data: [][]any{
					{"a", "2024-02-01", 5.0},
					{"a", "2024-02-02", 10.0},
				},
				Columns: []condition.ViewField{
					{Name: "dim1", Type: "varchar"},
					{Name: "__time", Type: "varchar"},
					{Name: "__value", Type: "float"},
				},
			}
			samePeriodDatas := interfaces.VegaFetchData{
				Data: [][]any{
					{"a", "2023-02-01", 2.0},
					{"a", "2023-02-02", 0},
				},
				Columns: []condition.ViewField{
					{Name: "dim1", Type: "varchar"},
					{Name: "__time", Type: "varchar"},
					{Name: "__value", Type: "float"},
				},
			}
			// metricModel := interfaces.MetricModel{}
			endI := int64(1706889720000)
			query := interfaces.MetricModelQuery{
				QueryTimeParams: interfaces.QueryTimeParams{
					StepStr:        &day,
					IsInstantQuery: false,
					Start:          &start,
					End:            &endI,
				},
				IsCalendar: true,
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.SQL,
				FormulaConfig: interfaces.SQLConfig{
					GroupByFields: []string{"dim1"},
					AggrExpr: &interfaces.AggrExpr{
						Aggr:  "sum",
						Field: "metric1",
					},
				},
				AnalysisDims: []string{"dim1"},
				RequestMetrics: &interfaces.RequestMetrics{
					Type: interfaces.METRICS_SAMEPERIOD,
					SamePeriodCfg: &interfaces.SamePeriodCfg{
						Method: []string{interfaces.METRICS_SAMEPERIOD_METHOD_GROWTH_RATE,
							interfaces.METRICS_SAMEPERIOD_METHOD_GROWTH_VALUE},
						Offset:          1,
						TimeGranularity: "year",
					},
				},
				MetricModelQueryParameters: interfaces.MetricModelQueryParameters{
					FillNull: true,
				},
			}
			vegaDuration := int64(300)
			resp, err := convertVegaDatas2TimeSeries(ctx, vegaData, samePeriodDatas, query, vegaDuration)
			So(err, ShouldBeNil)
			So(len(resp.Datas), ShouldEqual, 1)
			ts := resp.Datas[0]
			So(ts.Labels["dim1"], ShouldEqual, "a")
			So(ts.GrowthValues, ShouldResemble, []any{float64(3), nil, nil})
			So(ts.GrowthRates, ShouldResemble, []any{float64(150), nil, nil})
		})

		Convey("should return error if getFloat64Value fails", func() {
			vegaData := interfaces.VegaFetchData{
				Data: [][]any{
					{"a", "2024-02-01 00:00:00", "not-a-float"},
				},
				Columns: []condition.ViewField{
					{Name: "dim1", Type: "varchar"},
					{Name: "__time", Type: "varchar"},
					{Name: "__value", Type: "float"},
				},
			}
			samePeriodDatas := interfaces.VegaFetchData{}
			// metricModel := interfaces.MetricModel{}
			query := interfaces.MetricModelQuery{
				QueryTimeParams: interfaces.QueryTimeParams{
					StepStr:        &minute,
					IsInstantQuery: false,
					Start:          &start,
					End:            &end,
				},
				IsCalendar: true,
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.SQL,
				FormulaConfig: interfaces.SQLConfig{
					GroupByFields: []string{"dim1"},
					AggrExpr: &interfaces.AggrExpr{
						Aggr:  "sum",
						Field: "metric1",
					},
				},
				AnalysisDims: []string{"dim1"},
			}
			vegaDuration := int64(400)
			_, err := convertVegaDatas2TimeSeries(ctx, vegaData, samePeriodDatas, query, vegaDuration)
			So(err, ShouldNotBeNil)
		})
	})
}
