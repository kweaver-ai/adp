package driveradapters

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	"uniquery/common"
	cond "uniquery/common/condition"
	vopt "uniquery/common/value_opt"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
)

var (
	now    = time.Now()
	start1 = now.Add(-30 * time.Minute).UnixMilli()
	end1   = now.UnixMilli()

	start2 = int64(1646671470123)
	end2   = int64(1646360670123)

	start3 = int64(1678412100123)
	end3   = int64(1678412210123)

	step_5m      = "5m"
	step_30m     = "30m"
	step_1800000 = int64(1800000)
)

func TestValidatePaginationQueryParams(t *testing.T) {
	Convey("Test validatePaginationQueryParams", t, func() {

		Convey("Validate failed, caused by the invalid offset", func() {
			_, err := validatePaginationQueryParams(testCtx, interfaces.PaginationQueryParams{
				Offset: -1,
			}, interfaces.QUERY_CATEGORY_SPAN_LIST)

			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_InvalidParameter_Offset)
		})

		Convey("Validate failed, caused by the invalid limit", func() {
			_, err := validatePaginationQueryParams(testCtx, interfaces.PaginationQueryParams{
				Offset: 0,
				Limit:  -1,
			}, interfaces.QUERY_CATEGORY_SPAN_LIST)

			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_InvalidParameter_Limit)
		})

		Convey("Validate failed, caused by the invalid offset+limit", func() {
			_, err := validatePaginationQueryParams(testCtx, interfaces.PaginationQueryParams{
				Offset: 9001,
				Limit:  1000,
			}, interfaces.QUERY_CATEGORY_SPAN_LIST)

			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_InvalidParameter_OffestAndLimitSum)
		})

		Convey("Validate failed, caused by the invalid sort", func() {
			_, err := validatePaginationQueryParams(testCtx, interfaces.PaginationQueryParams{
				Offset: 0,
				Limit:  1,
				Sort:   "invalid",
			}, interfaces.QUERY_CATEGORY_SPAN_LIST)

			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_InvalidParameter_Sort)
		})

		Convey("Validate failed, caused by the invalid direction", func() {
			_, err := validatePaginationQueryParams(testCtx, interfaces.PaginationQueryParams{
				Offset:    0,
				Limit:     1,
				Sort:      "@timestamp",
				Direction: "invalid",
			}, interfaces.QUERY_CATEGORY_SPAN_LIST)

			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_InvalidParameter_Direction)
		})

		Convey("Validate succeed", func() {
			_, err := validatePaginationQueryParams(testCtx, interfaces.PaginationQueryParams{
				Offset:    0,
				Limit:     1,
				Sort:      "@timestamp",
				Direction: "desc",
			}, interfaces.QUERY_CATEGORY_SPAN_LIST)

			So(err, ShouldBeNil)
		})
	})
}

func TestValidateDataViewID(t *testing.T) {
	Convey("Test ValidateDataViewID", t, func() {
		c := context.Background()

		Convey("The trace_data_view_id parameter is null", func() {
			err := ValidateDataView(c, "", "trace")
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_Trace_InvalidParameter_TraceDataViewID)
		})

		Convey("The trace_data_view_id parameter is valid", func() {
			err := ValidateDataView(c, "1", "trace")
			So(err, ShouldBeNil)
		})

		Convey("The log_data_view_id parameter is null", func() {
			err := ValidateDataView(c, "", "log")
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_Trace_InvalidParameter_LogDataViewID)
		})

		Convey("The log_data_view_id parameter is valid", func() {
			err := ValidateDataView(c, "1", "log")
			So(err, ShouldBeNil)
		})
	})
}

// func TestValidateSpanStatuses(t *testing.T) {
// 	Convey("Test ValidateSpanStatuses", t, func() {
// 		c := context.Background()

// 		Convey("The span_statuses parameter is null", func() {
// 			expectedErr := &rest.HTTPError{
// 				HTTPCode: http.StatusBadRequest,
// 				BaseError: rest.BaseError{
// 					ErrorCode:    uerrors.Uniquery_Trace_InvalidParameter_SpanStatuses,
// 					Description:  "指定的跨度状态无效",
// 					Solution:     "",
// 					ErrorLink:    "",
// 					ErrorDetails: "The span_statuses is null",
// 				},
// 			}
// 			patch := ApplyFunc(rest.NewHTTPError,
// 				func(httpCode int, errorCode string) *rest.HTTPError {
// 					return expectedErr
// 				},
// 			)
// 			defer patch.Reset()

// 			_, err := ValidateSpanStatuses(c, "")
// 			So(err, ShouldResemble, expectedErr)
// 		})

// 		Convey("The span_statuses parameter is invaild", func() {
// 			expectedErr := &rest.HTTPError{
// 				HTTPCode: http.StatusBadRequest,
// 				BaseError: rest.BaseError{
// 					ErrorCode:    uerrors.Uniquery_Trace_InvalidParameter_SpanStatuses,
// 					Description:  "指定的跨度状态无效",
// 					Solution:     "",
// 					ErrorLink:    "",
// 					ErrorDetails: fmt.Sprintf("The span_statuses is not in the set of [%s]", interfaces.DEFAULT_SPAN_STATUSES),
// 				},
// 			}
// 			patch := ApplyFunc(rest.NewHTTPError,
// 				func(httpCode int, errorCode string) *rest.HTTPError {
// 					return expectedErr
// 				},
// 			)
// 			defer patch.Reset()

// 			_, err := ValidateSpanStatuses(c, "abc")
// 			So(err, ShouldResemble, expectedErr)
// 		})

// 		Convey("The span_statuses parameter is vaild", func() {
// 			_, err := ValidateSpanStatuses(c, "Ok,Error")
// 			So(err, ShouldBeNil)
// 		})
// 	})
// }

// func TestValidateSpanQueryTime(t *testing.T) {
// 	Convey("Test ValidateSpanQueryTime", t, func() {
// 		c := context.Background()

// 		Convey("The start_time parameter is null", func() {
// 			expectedErr := &rest.HTTPError{
// 				HTTPCode: http.StatusBadRequest,
// 				BaseError: rest.BaseError{
// 					ErrorCode:    uerrors.Uniquery_Trace_InvalidParameter_StartTime,
// 					Description:  "指定的开始时间无效",
// 					Solution:     "",
// 					ErrorLink:    "",
// 					ErrorDetails: "The start_time is null",
// 				},
// 			}
// 			patch := ApplyFunc(rest.NewHTTPError,
// 				func(httpCode int, errorCode string) *rest.HTTPError {
// 					return expectedErr
// 				},
// 			)
// 			defer patch.Reset()

// 			_, _, err := ValidateSpanQueryTime(c, "", "")
// 			So(err, ShouldResemble, expectedErr)
// 		})

// 		Convey("The start_time parameter is invaild, its length is not equal to 13", func() {
// 			expectedErr := &rest.HTTPError{
// 				HTTPCode: http.StatusBadRequest,
// 				BaseError: rest.BaseError{
// 					ErrorCode:    uerrors.Uniquery_Trace_InvalidParameter_StartTime,
// 					Description:  "指定的开始时间无效",
// 					Solution:     "",
// 					ErrorLink:    "",
// 					ErrorDetails: fmt.Sprintf("The start_time is not a Unix millisecond timestamp because its length is not equal to %d", interfaces.UNIX_MILLISECOND_TIMESTAMP_STR_LENGTH),
// 				},
// 			}
// 			patch := ApplyFunc(rest.NewHTTPError,
// 				func(httpCode int, errorCode string) *rest.HTTPError {
// 					return expectedErr
// 				},
// 			)
// 			defer patch.Reset()

// 			_, _, err := ValidateSpanQueryTime(c, "1", "2")
// 			So(err, ShouldResemble, expectedErr)
// 		})

// 		Convey("The start_time parameter is invaild bacause it cannot be converted to decimal int64", func() {
// 			err := errors.New("strconv.ParseInt: parsing \"abc\": invalid syntax")
// 			expectedErr := &rest.HTTPError{
// 				HTTPCode: http.StatusBadRequest,
// 				BaseError: rest.BaseError{
// 					ErrorCode:    uerrors.Uniquery_Trace_InvalidParameter_StartTime,
// 					Description:  "指定的开始时间无效",
// 					Solution:     "",
// 					ErrorLink:    "",
// 					ErrorDetails: fmt.Sprintf("The start_time cannot be converted to decimal int64, err: %v", err.Error()),
// 				},
// 			}
// 			patch := ApplyFunc(rest.NewHTTPError,
// 				func(httpCode int, errorCode string) *rest.HTTPError {
// 					return expectedErr
// 				},
// 			)
// 			defer patch.Reset()

// 			_, _, err = ValidateSpanQueryTime(c, "abc", "abc")
// 			So(err, ShouldResemble, expectedErr)
// 		})

// 		Convey("The end_time parameter is null", func() {
// 			expectedErr := &rest.HTTPError{
// 				HTTPCode: http.StatusBadRequest,
// 				BaseError: rest.BaseError{
// 					ErrorCode:    uerrors.Uniquery_Trace_InvalidParameter_EndTime,
// 					Description:  "指定的结束时间无效",
// 					Solution:     "",
// 					ErrorLink:    "",
// 					ErrorDetails: "The end_time is null",
// 				},
// 			}
// 			patch := ApplyFunc(rest.NewHTTPError,
// 				func(httpCode int, errorCode string) *rest.HTTPError {
// 					return expectedErr
// 				},
// 			)
// 			defer patch.Reset()

// 			_, _, err := ValidateSpanQueryTime(c, "", "")
// 			So(err, ShouldResemble, expectedErr)
// 		})

// 		Convey("The end_time parameter is invaild, its length is not equal to 13", func() {
// 			expectedErr := &rest.HTTPError{
// 				HTTPCode: http.StatusBadRequest,
// 				BaseError: rest.BaseError{
// 					ErrorCode:    uerrors.Uniquery_Trace_InvalidParameter_EndTime,
// 					Description:  "指定的结束时间无效",
// 					Solution:     "",
// 					ErrorLink:    "",
// 					ErrorDetails: fmt.Sprintf("The end_time is not a Unix millisecond timestamp because its length is not equal to %d", interfaces.UNIX_MILLISECOND_TIMESTAMP_STR_LENGTH),
// 				},
// 			}
// 			patch := ApplyFunc(rest.NewHTTPError,
// 				func(httpCode int, errorCode string) *rest.HTTPError {
// 					return expectedErr
// 				},
// 			)
// 			defer patch.Reset()

// 			_, _, err := ValidateSpanQueryTime(c, "1614136611000", "abc")
// 			So(err, ShouldResemble, expectedErr)
// 		})

// 		Convey("The end_time parameter is invaild bacause it cannot be converted to decimal int64", func() {
// 			err := errors.New("strconv.ParseInt: parsing \"abc\": invalid syntax")
// 			expectedErr := &rest.HTTPError{
// 				HTTPCode: http.StatusBadRequest,
// 				BaseError: rest.BaseError{
// 					ErrorCode:    uerrors.Uniquery_Trace_InvalidParameter_EndTime,
// 					Description:  "指定的结束时间无效",
// 					Solution:     "",
// 					ErrorLink:    "",
// 					ErrorDetails: fmt.Sprintf("The end_time cannot be converted to decimal int64, err: %v", err.Error()),
// 				},
// 			}
// 			patch := ApplyFunc(rest.NewHTTPError,
// 				func(httpCode int, errorCode string) *rest.HTTPError {
// 					return expectedErr
// 				},
// 			)
// 			defer patch.Reset()

// 			_, _, err = ValidateSpanQueryTime(c, "1614136611000", "abc")
// 			So(err, ShouldResemble, expectedErr)
// 		})

// 		// Convey("The end_time is invalid because it is greater than current time", func() {
// 		// 	current := 1677208611000

// 		// 	patchTimeNow := ApplyFunc(time.Now,
// 		// 		func() time.Time {
// 		// 			return time.UnixMilli(int64(current))
// 		// 		},
// 		// 	)
// 		// 	defer patchTimeNow.Reset()
// 		// 	expectedErr := &rest.HTTPError{
// 		// 		HTTPCode: http.StatusBadRequest,
// 		// 		BaseError: rest.BaseError{
// 		// 			ErrorCode:    uerrors.Uniquery_InvalidParameter_EndTime,
// 		// 			Description:  "指定的结束时间无效",
// 		// 			Solution:     "",
// 		// 			ErrorLink:    "",
// 		// 			ErrorDetails: fmt.Sprintf("The end_time is greater than current time, current timestamp is %v.", current),
// 		// 		},
// 		// 	}
// 		// 	patchNewHttpError := ApplyFunc(rest.NewHTTPError,
// 		// 		func(httpCode int, errorCode string) *rest.HTTPError {
// 		// 			return expectedErr
// 		// 		},
// 		// 	)
// 		// 	defer patchNewHttpError.Reset()

// 		// 	_, _, err := ValidateSpanQueryTime("1677208511000", "1677208911000")
// 		// 	So(err, ShouldResemble, expectedErr)
// 		// })

// 		Convey("The end_time parameter is invalid because it is lesser than start_time", func() {
// 			current := 1677208611000

// 			patchTimeNow := ApplyFunc(time.Now,
// 				func() time.Time {
// 					return time.UnixMilli(int64(current))
// 				},
// 			)
// 			defer patchTimeNow.Reset()
// 			expectedErr := &rest.HTTPError{
// 				HTTPCode: http.StatusBadRequest,
// 				BaseError: rest.BaseError{
// 					ErrorCode:    uerrors.Uniquery_Trace_InvalidParameter_EndTime,
// 					Description:  "指定的结束时间无效",
// 					Solution:     "",
// 					ErrorLink:    "",
// 					ErrorDetails: "The end_time is before start_time",
// 				},
// 			}
// 			patchNewHttpError := ApplyFunc(rest.NewHTTPError,
// 				func(httpCode int, errorCode string) *rest.HTTPError {
// 					return expectedErr
// 				},
// 			)
// 			defer patchNewHttpError.Reset()

// 			_, _, err := ValidateSpanQueryTime(c, "1677208011000", "1677207511000")
// 			So(err, ShouldResemble, expectedErr)
// 		})

// 		Convey("The end_time and start_time parameters are valid", func() {
// 			current := 1677208611000

// 			patchTimeNow := ApplyFunc(time.Now,
// 				func() time.Time {
// 					return time.UnixMilli(int64(current))
// 				},
// 			)
// 			defer patchTimeNow.Reset()

// 			_, _, err := ValidateSpanQueryTime(c, "1677208111000", "1677208511000")
// 			So(err, ShouldBeNil)
// 		})
// 	})
// }

// func TestValidateOffsetAndLimit(t *testing.T) {
// 	Convey("Test ValidateOffsetAndLimit", t, func() {
// 		c := context.Background()

// 		Convey("Validate failed, because the offset cannot be converted to int", func() {
// 			expectedErr := &rest.HTTPError{
// 				HTTPCode: http.StatusBadRequest,
// 				BaseError: rest.BaseError{
// 					ErrorCode:    uerrors.Uniquery_InvalidParameter_Offset,
// 					Description:  "指定的偏移量无效",
// 					Solution:     "",
// 					ErrorLink:    "",
// 					ErrorDetails: "strconv.Atoi: parsing \"a\": invalid syntax",
// 				},
// 			}
// 			patch := ApplyFunc(rest.NewHTTPError,
// 				func(httpCode int, errorCode string) *rest.HTTPError {
// 					return expectedErr
// 				},
// 			)
// 			defer patch.Reset()

// 			offsetStr := "a"
// 			limitStr := "1000"

// 			_, _, err := ValidateOffsetAndLimit(c, offsetStr, limitStr)
// 			So(err, ShouldEqual, expectedErr)
// 		})

// 		Convey("Validate failed, because the offset is not greater than MIN_OFFSET", func() {
// 			expectedErr := &rest.HTTPError{
// 				HTTPCode: http.StatusBadRequest,
// 				BaseError: rest.BaseError{
// 					ErrorCode:    uerrors.Uniquery_InvalidParameter_Offset,
// 					Description:  "指定的偏移量无效",
// 					Solution:     "",
// 					ErrorLink:    "",
// 					ErrorDetails: fmt.Sprintf("The offset is not greater than %d", interfaces.MIN_OFFSET),
// 				},
// 			}
// 			patch := ApplyFunc(rest.NewHTTPError,
// 				func(httpCode int, errorCode string) *rest.HTTPError {
// 					return expectedErr
// 				},
// 			)
// 			defer patch.Reset()

// 			offsetStr := "-2"
// 			limitStr := "1000"

// 			_, _, res := ValidateOffsetAndLimit(c, offsetStr, limitStr)
// 			So(res, ShouldEqual, expectedErr)
// 		})

// 		Convey("Validate failed, because the limit cannot be converted to int", func() {
// 			expectedErr := &rest.HTTPError{
// 				HTTPCode: http.StatusBadRequest,
// 				BaseError: rest.BaseError{
// 					ErrorCode:    uerrors.Uniquery_InvalidParameter_Limit,
// 					Description:  "指定的每页数量限制无效",
// 					Solution:     "",
// 					ErrorLink:    "",
// 					ErrorDetails: "strconv.Atoi: parsing \"a\": invalid syntax",
// 				},
// 			}
// 			patch := ApplyFunc(rest.NewHTTPError,
// 				func(httpCode int, errorCode string) *rest.HTTPError {
// 					return expectedErr
// 				},
// 			)
// 			defer patch.Reset()

// 			offsetStr := "0"
// 			limitStr := "a"

// 			_, _, res := ValidateOffsetAndLimit(c, offsetStr, limitStr)
// 			So(res, ShouldEqual, expectedErr)
// 		})

// 		Convey("Validate failed, because the limit is not in the range of [MIN_LIMIT,MAX_LIMIT]", func() {
// 			expectedErr := &rest.HTTPError{
// 				HTTPCode: http.StatusBadRequest,
// 				BaseError: rest.BaseError{
// 					ErrorCode:    uerrors.Uniquery_InvalidParameter_Limit,
// 					Description:  "指定的每页数量限制无效",
// 					Solution:     "",
// 					ErrorLink:    "",
// 					ErrorDetails: fmt.Sprintf("The number per page is not in the range of [%d,%d]", interfaces.MIN_LIMIT, interfaces.MAX_LIMIT),
// 				},
// 			}
// 			patch := ApplyFunc(rest.NewHTTPError,
// 				func(httpCode int, errorCode string) *rest.HTTPError {
// 					return expectedErr
// 				},
// 			)
// 			defer patch.Reset()

// 			offsetStr := "0"
// 			limitStr := "1100"

// 			_, _, err := ValidateOffsetAndLimit(c, offsetStr, limitStr)
// 			So(err, ShouldEqual, expectedErr)
// 		})

// 		Convey("Validate succeed", func() {
// 			offsetStr := "0"
// 			limitStr := "800"

// 			_, _, err := ValidateOffsetAndLimit(c, offsetStr, limitStr)
// 			So(err, ShouldBeNil)
// 		})
// 	})
// }

func TestValidateMetricModelSimulate(t *testing.T) {
	Convey("Test ValidateMetricModelSimulate", t, func() {
		c := context.Background()
		common.FixedStepsMap = StepsMap

		Convey("Validate failed, because MetricType is null", func() {

			res := ValidateMetricModelSimulate(c, &interfaces.MetricModelQuery{})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_NullParameter_MetricType)
		})

		Convey("Validate failed, because MetricType unsupport", func() {

			res := ValidateMetricModelSimulate(c, &interfaces.MetricModelQuery{MetricType: "A"})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_UnsupportMetricType)
		})

		Convey("Validate failed, because atomic datasource is null", func() {

			res := ValidateMetricModelSimulate(c, &interfaces.MetricModelQuery{MetricType: "atomic"})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_NullParameter_DataSource)
		})

		Convey("Validate failed, because atomic datasource id is empty", func() {

			res := ValidateMetricModelSimulate(c, &interfaces.MetricModelQuery{
				MetricType: "atomic",
				DataSource: &interfaces.MetricDataSource{
					Type: "data_view",
				},
			})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_NullParameter_DataSourceID)
		})

		Convey("Validate failed, because DataViewID is null", func() {

			res := ValidateMetricModelSimulate(c, &interfaces.MetricModelQuery{MetricType: "atomic"})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_NullParameter_DataSource)
		})

		Convey("Validate failed, because QueryType is null", func() {

			res := ValidateMetricModelSimulate(c, &interfaces.MetricModelQuery{
				MetricType: "atomic",
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
			})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_NullParameter_QueryType)
		})

		Convey("Validate failed, because QueryType unsupport", func() {

			res := ValidateMetricModelSimulate(c, &interfaces.MetricModelQuery{
				MetricType: "atomic",
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
				QueryType: "A",
			})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_UnsupportQueryType)
		})

		Convey("Validate failed, because Formula config is null", func() {
			res := ValidateMetricModelSimulate(c, &interfaces.MetricModelQuery{
				MetricType: "atomic",
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
				QueryType: "sql",
			})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_NullParameter_FormulaConfig)
		})

		Convey("Validate failed, when query type is sql and AggrExpr is not nil but empty && AggrExprStr is empty ", func() {
			res := ValidateMetricModelSimulate(c, &interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.SQL,
				FormulaConfig: interfaces.SQLConfig{
					AggrExpr: &interfaces.AggrExpr{},
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.QueryType_SQL,
					ID:   "a",
				},
			},
			)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_NullParameter_SQLAggrExpression)
		})

		Convey("Validate failed, when query type is sql and AggrExpr is not empty && AggrExprStr is not empty", func() {
			res := ValidateMetricModelSimulate(c, &interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.SQL,
				FormulaConfig: interfaces.SQLConfig{
					AggrExpr: &interfaces.AggrExpr{
						Aggr:  "avg",
						Field: "f1",
					},
					AggrExprStr: "avg(a)",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.QueryType_SQL,
					ID:   "a",
				},
			},
			)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_SQLAggrExpression)
		})

		Convey("Validate failed, when query type is sql and both Condition and ConditionStr are not empty", func() {
			res := ValidateMetricModelSimulate(c, &interfaces.MetricModelQuery{
				ModelName:  "a",
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.SQL,
				FormulaConfig: interfaces.SQLConfig{
					AggrExpr:     &interfaces.AggrExpr{},
					AggrExprStr:  "avg(a)",
					ConditionStr: "a=1",
					Condition:    &cond.CondCfg{},
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.QueryType_SQL,
					ID:   "a",
				},
			},
			)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_SQLCondition)
		})

		// Convey("Validate failed, when query type is sql and date field is empty", func() {
		// 	res := ValidateMetricModelSimulate(c, &interfaces.MetricModelQuery{
		// 		ModelName:  "a",
		// 		MetricType: interfaces.ATOMIC_METRIC,
		// 		QueryType:  interfaces.SQL,
		// 		FormulaConfig: interfaces.SQLConfig{
		// 			AggrExpr: &interfaces.AggrExpr{
		// 				Aggr:  "avg",
		// 				Field: "f1",
		// 			},
		// 			AggrExprStr:  "",
		// 			ConditionStr: "a=1",
		// 		},
		// 		DataSource: &interfaces.MetricDataSource{
		// 			Type: interfaces.QueryType_SQL,
		// 			ID:   "a",
		// 		},
		// 	},
		// 	)
		// 	So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
		// 	So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_NullParameter_DateField)
		// })

		// Convey("Validate failed, when query type is sql and date source is not vega", func() {
		// 	res := ValidateMetricModelSimulate(c, &interfaces.MetricModelQuery{
		// 		ModelName:  "a",
		// 		MetricType: interfaces.ATOMIC_METRIC,
		// 		QueryType:  interfaces.SQL,
		// 		FormulaConfig: interfaces.SQLConfig{
		// 			AggrExpr: &interfaces.AggrExpr{
		// 				Aggr:  "avg",
		// 				Field: "f1",
		// 			},
		// 			AggrExprStr:  "",
		// 			ConditionStr: "a=1",
		// 		},
		// 		DataSource: &interfaces.MetricDataSource{
		// 			Type: interfaces.QueryType_DSL,
		// 			ID:   "a",
		// 		},
		// 		DateField: "f1",
		// 	},
		// 	)
		// 	So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
		// 	So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_DataSourceType)
		// })

		Convey("Validate failed, because Formula is null", func() {
			res := ValidateMetricModelSimulate(c, &interfaces.MetricModelQuery{
				MetricType: "atomic",
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
				QueryType: "promql",
				Formula:   "",
			})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_NullParameter_Formula)
		})

		Convey("Validate failed, because validate dsl err", func() {
			res := ValidateMetricModelSimulate(c, &interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
				QueryType: interfaces.DSL,
				Formula:   `"a": {{__interval}}`,
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_Formula)
		})

		Convey("Validate failed, because filter name is null ", func() {
			res := ValidateMetricModelSimulate(c, &interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
				QueryType: interfaces.PROMQL,
				Formula:   `a`,
				Filters:   []interfaces.Filter{{}},
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &start1,
					End:     &end1,
					StepStr: &step_5m,
				},
			})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_NullParameter_FilterName)
		})

		Convey("Validate failed, because start is null", func() {
			endi := int64(123)
			res := ValidateMetricModelSimulate(c, &interfaces.MetricModelQuery{
				MetricType: "atomic",
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
				QueryType: "promql",
				Formula:   "1+2",
				QueryTimeParams: interfaces.QueryTimeParams{
					IsInstantQuery: false,
					End:            &endi,
				},
			})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_InvalidParameter_Start)
		})

		Convey("Validate failed, because query.RequestMetrics.Type is null", func() {
			res := ValidateMetricModelSimulate(c, &interfaces.MetricModelQuery{
				MetricType: "atomic",
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
				QueryType: "promql",
				Formula:   "1+2",
				QueryTimeParams: interfaces.QueryTimeParams{
					IsInstantQuery: false,
					Start:          &start1,
					End:            &end1,
					StepStr:        &step_5m,
				},
				RequestMetrics: &interfaces.RequestMetrics{},
			})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_NullParameter_RequestMetricsType)
		})

		Convey("Validate success, because start end step all null", func() {
			res := ValidateMetricModelSimulate(c, &interfaces.MetricModelQuery{
				MetricType: "atomic",
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
				QueryType: "promql",
				Formula:   "1+2",
				QueryTimeParams: interfaces.QueryTimeParams{
					IsInstantQuery: false,
					Start:          &start1,
					End:            &end1,
					StepStr:        &step_5m,
				},
			})
			So(res, ShouldBeNil)
		})

		Convey("Validate success, because calendar_interval && step is invalid", func() {
			res := ValidateMetricModelSimulate(c, &interfaces.MetricModelQuery{
				MetricType: "atomic",
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "a",
				},
				QueryType:          "promql",
				Formula:            "1+2",
				IsCalendarInterval: 1,
				QueryTimeParams: interfaces.QueryTimeParams{
					IsInstantQuery: false,
					Start:          &start1,
					End:            &end1,
				},
			})
			So(res, ShouldBeNil)
		})
	})
}

func TestValidateTimes(t *testing.T) {
	Convey("Test validateTimes", t, func() {
		c := context.Background()

		// Convey("Validate failed, because start > now ", func() {
		// 	now := time.Now()
		// 	hh, _ := time.ParseDuration("1h")
		// 	start := now.Add(hh)

		// 	_, res := validateTimes(c, start, now)
		// 	So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_InvalidParameter_Start)
		// })

		// Convey("Validate success, because end > now ", func() {
		// 	now := time.Now()
		// 	hh, _ := time.ParseDuration("1h")
		// 	end := now.Add(hh)
		// 	h, _ := time.ParseDuration("-1h")
		// 	start := now.Add(h)

		// 	actualEnd, res := validateTimes(c, start, end)
		// 	So(actualEnd.UnixNano()/1e9, ShouldEqual, now.UnixNano()/1e9)
		// 	So(res, ShouldBeNil)
		// })

		Convey("Validate failed, because start > end ", func() {
			now := time.Now()
			hh, _ := time.ParseDuration("-1h")
			end := now.Add(hh)
			h, _ := time.ParseDuration("1h")
			start := now.Add(h)

			_, res := validateTimes(c, start, end)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_InvalidParameter_Start)
		})

	})
}

func TestValidateMetricModelData(t *testing.T) {
	Convey("Test validateMetricModelData", t, func() {
		c := context.Background()

		// Convey("Validate failed, because method is null", func() {
		// 	res := validateMetricModelData(c, &([]*interfaces.MetricModelQuery{{}}))
		// 	So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_NullParameter_MethodOverride)
		// })

		// Convey("Validate failed, because method is not get", func() {
		// 	res := validateMetricModelData(c, &([]*interfaces.MetricModelQuery{{Method: "GETa"}}))
		// 	So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_MethodOverride)
		// })
		common.FixedStepsMap = StepsMap
		Convey("Validate failed, because start is null ", func() {
			res := validateMetricModelData(c, &interfaces.MetricModelQuery{})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_InvalidParameter_Start)
		})

		Convey("Validate failed, because start is invalid ", func() {
			starti := int64(-1)
			res := validateMetricModelData(c, &interfaces.MetricModelQuery{QueryTimeParams: interfaces.QueryTimeParams{Start: &starti}})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_InvalidParameter_Start)
		})

		Convey("Validate failed, because end is null ", func() {
			starti := int64(1)
			res := validateMetricModelData(c, &interfaces.MetricModelQuery{QueryTimeParams: interfaces.QueryTimeParams{Start: &starti}})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_InvalidParameter_End)
		})

		Convey("Validate failed, because end is invalid ", func() {
			starti := int64(1)
			endi := int64(-1)
			res := validateMetricModelData(c, &interfaces.MetricModelQuery{
				QueryTimeParams: interfaces.QueryTimeParams{Start: &starti, End: &endi}})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_InvalidParameter_End)
		})

		Convey("Validate failed, because start > end ", func() {
			res := validateMetricModelData(c, &interfaces.MetricModelQuery{
				QueryTimeParams: interfaces.QueryTimeParams{Start: &start2, End: &end2}})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_InvalidParameter_Start)
		})

		Convey("Validate failed, because step is null ", func() {
			res := validateMetricModelData(c, &interfaces.MetricModelQuery{
				QueryTimeParams: interfaces.QueryTimeParams{Start: &end2, End: &start2}})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_NullParameter_Step)
		})

		Convey("Validate failed, because step is invalid ", func() {
			stepi := "1a"
			res := validateMetricModelData(c, &interfaces.MetricModelQuery{
				QueryTimeParams: interfaces.QueryTimeParams{Start: &end2, End: &start2, StepStr: &stepi}})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_Step)
		})

		Convey("Validate failed, because step < 0 ", func() {
			stepi := "-1"
			res := validateMetricModelData(c, &interfaces.MetricModelQuery{
				QueryTimeParams: interfaces.QueryTimeParams{Start: &end2, End: &start2, StepStr: &stepi}})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_Step)
		})

		Convey("Validate failed, because step should a multiple of 5 minutes when step > 30min ", func() {
			stepi := "31m"
			res := validateMetricModelData(c, &interfaces.MetricModelQuery{
				QueryTimeParams: interfaces.QueryTimeParams{Start: &end2, End: &start2, StepStr: &stepi}})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_Step)
		})

		Convey("Validate failed, because exceeded maximum resolution of 11,000 points per timeseries.", func() {
			stepi := "1ms"
			res := validateMetricModelData(c, &interfaces.MetricModelQuery{
				QueryTimeParams: interfaces.QueryTimeParams{Start: &start3, End: &end3, StepStr: &stepi}})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_Step)
		})

		Convey("Validate failed, because filter name is null ", func() {
			stepi := "1m"
			res := validateMetricModelData(c, &interfaces.MetricModelQuery{
				QueryTimeParams: interfaces.QueryTimeParams{Start: &end2, End: &start2, StepStr: &stepi},
				Filters:         []interfaces.Filter{{}}})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_NullParameter_FilterName)
		})

		Convey("Validate failed, because query.RequestMetrics.Type is empty", func() {
			query := interfaces.MetricModelQuery{QueryTimeParams: interfaces.QueryTimeParams{Start: &start3, End: &end3, StepStr: &step_30m},
				RequestMetrics: &interfaces.RequestMetrics{}}
			err := validateMetricModelData(c, &query)
			So(query, ShouldResemble, interfaces.MetricModelQuery{
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &start3,
					End:     &end3,
					StepStr: &step_30m,
					Step:    &step_1800000,
				},
				RequestMetrics: &interfaces.RequestMetrics{},
			})
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_NullParameter_RequestMetricsType)
		})

		Convey("Validate success", func() {
			query := interfaces.MetricModelQuery{QueryTimeParams: interfaces.QueryTimeParams{Start: &start3, End: &end3, StepStr: &step_30m}}
			err := validateMetricModelData(c, &query)
			So(query, ShouldResemble, interfaces.MetricModelQuery{
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &start3,
					End:     &end3,
					StepStr: &step_30m,
					Step:    &step_1800000,
				},
			})
			So(err, ShouldBeNil)
		})

		Convey("Validate success with calendar interval", func() {
			starti := int64(1668412100123)
			query := interfaces.MetricModelQuery{QueryTimeParams: interfaces.QueryTimeParams{Start: &starti, End: &end3, StepStr: &step_5m}}
			err := validateMetricModelData(c, &query)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_Step)
		})

		Convey("Validate failed, because instant query && time le 0 ", func() {
			res := validateMetricModelData(c, &interfaces.MetricModelQuery{QueryTimeParams: interfaces.QueryTimeParams{IsInstantQuery: true, Time: -1},
				Filters: []interfaces.Filter{{}}})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_Time)
		})

		Convey("Validate failed, because instant query && lookbackDelta invalid ", func() {
			res := validateMetricModelData(c, &interfaces.MetricModelQuery{QueryTimeParams: interfaces.QueryTimeParams{IsInstantQuery: true, LookBackDeltaStr: "a"},
				Filters: []interfaces.Filter{{}}})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_LookBackDelta)
		})

		Convey("Validate failed, because instant query && lookbackDelta le 0 ", func() {
			res := validateMetricModelData(c, &interfaces.MetricModelQuery{QueryTimeParams: interfaces.QueryTimeParams{IsInstantQuery: true, LookBackDeltaStr: "-1"},
				Filters: []interfaces.Filter{{}}})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_LookBackDelta)
		})

		Convey("Validate success with instant query", func() {
			starti := int64(1683700342518)
			endi := int64(1683700642518)
			query := interfaces.MetricModelQuery{
				QueryTimeParams: interfaces.QueryTimeParams{
					IsInstantQuery: true,
					// LookBackDeltaStr: &step_5m,
					Start: &starti,
					End:   &endi}}
			err := validateMetricModelData(c, &query)
			So(query, ShouldResemble, interfaces.MetricModelQuery{
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:          &starti,
					End:            &endi,
					StepStr:        nil,
					Step:           nil,
					IsInstantQuery: true,
					// Time:             1683700642518,
					// LookBackDeltaStr: &step_5m,
					// LookBackDelta:    300000,
				},
			})
			So(err, ShouldBeNil)
		})

	})
}

func TestValidateFilters(t *testing.T) {
	Convey("Test validateFilters", t, func() {
		c := context.Background()

		Convey("Validate success, because filters is empty ", func() {
			res := validateFilters(c, []interfaces.Filter{})
			So(res, ShouldBeNil)
		})

		Convey("Validate failed, because filter name is null ", func() {
			res := validateFilters(c, []interfaces.Filter{{}})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_NullParameter_FilterName)
		})

		Convey("Validate failed, because filter operation is null ", func() {
			res := validateFilters(c, []interfaces.Filter{{Name: "a", Value: []interface{}{"a", "b"}}})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_NullParameter_FilterOperation)
		})

		Convey("Validate failed, because filter operation is unsupoort ", func() {
			res := validateFilters(c, []interfaces.Filter{{Name: "a", Value: []interface{}{"a", "b"}, Operation: "A"}})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_UnsupportFilterOperation)
		})

		Convey("Validate failed, because operation is in and filter value is null ", func() {
			res := validateFilters(c, []interfaces.Filter{{Name: "a", Operation: "in"}})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_NullParameter_FilterValue)
		})

		Convey("Validate failed, because operation is in and filter value length lte 0 ", func() {
			res := validateFilters(c, []interfaces.Filter{{Name: "a", Value: []interface{}{}, Operation: interfaces.OPERATION_IN}})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_InvalidParameter_FilterValue)
		})

		Convey("Validate failed, because operation is in and filter value is not array ", func() {
			res := validateFilters(c, []interfaces.Filter{{Name: "a", Value: "a", Operation: interfaces.OPERATION_IN}})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_InvalidParameter_FilterValue)
		})

		Convey("Validate failed, because operation is = and filter value is array ", func() {
			res := validateFilters(c, []interfaces.Filter{{Name: "a", Value: []interface{}{}, Operation: interfaces.OPERATION_EQ}})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_InvalidParameter_FilterValue)
		})

		Convey("Validate failed, because operation is != and filter value is array ", func() {
			res := validateFilters(c, []interfaces.Filter{{Name: "a", Value: []interface{}{}, Operation: cond.OperationNotEq}})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_InvalidParameter_FilterValue)
		})

		Convey("Validate failed, because operation is range and filter value is not an array ", func() {
			res := validateFilters(c, []interfaces.Filter{{Name: "a", Value: "a", Operation: cond.OperationRange}})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_InvalidParameter_FilterValue)
		})

		Convey("Validate failed, because operation is range and filter value length ne 2 ", func() {
			res := validateFilters(c, []interfaces.Filter{{Name: "a", Value: []interface{}{}, Operation: cond.OperationRange}})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_InvalidParameter_FilterValue)
		})

		Convey("Validate failed, because operation is out_range and filter value is not an array ", func() {
			res := validateFilters(c, []interfaces.Filter{{Name: "a", Value: "a", Operation: cond.OperationOutRange}})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_InvalidParameter_FilterValue)
		})

		Convey("Validate failed, because operation is out_range and filter value length ne 2 ", func() {
			res := validateFilters(c, []interfaces.Filter{{Name: "a", Value: []interface{}{}, Operation: cond.OperationOutRange}})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_InvalidParameter_FilterValue)
		})

		Convey("Validate success ", func() {
			res := validateFilters(c, []interfaces.Filter{{Name: "a", Value: []interface{}{"a", "b"}, Operation: interfaces.OPERATION_IN}})
			So(res, ShouldBeNil)
		})

	})
}

func TestValidateDSL(t *testing.T) {

	Convey("Test ValidateDSL", t, func() {
		c := context.Background()

		Convey("Validate failed, because dsl Unmarshal err", func() {

			res := validateDSL(c, &interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.DSL,
				Formula:    `"a": {{__interval}}`,
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_Formula)
		})

		Convey("Validate failed, because missing dsl size", func() {

			res := validateDSL(c, &interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.DSL,
				Formula:    `{"a": 1}`,
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "The size of dsl expected 0, actual is <nil>")
		})

		Convey("Validate failed, because dsl size != 0", func() {

			res := validateDSL(c, &interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.DSL,
				Formula:    `{"a": 1,"size":1}`,
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "The size of dsl expected 0, actual is 1")
		})

		Convey("Validate failed, because dsl aggs not exists", func() {

			res := validateDSL(c, &interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.DSL,
				Formula:    `{"a": 1,"size":0}`,
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "dsl missing aggregation.")
		})

		Convey("Validate-validAggs failed, because dsl aggs is not a map", func() {
			res := validateDSL(c, &interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.DSL,
				Formula:    `{"aggs": 1,"size":0}`,
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "The aggregation of dsl is not a map")
		})

		Convey("Validate-validAggs failed, because dsl aggs contain multi aggs", func() {
			res := validateDSL(c, &interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.DSL,
				Formula:    `{"size":0,"aggs":{"terms1":{},"terms2":{}}}`,
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "Multiple aggregation is not supported")
		})

		Convey("Validate-validAggs failed, because aggs name are same", func() {

			res := validateDSL(c, &interfaces.MetricModelQuery{
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
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "The aggregation names of each layer aggregation cannot be the same")
		})

		Convey("Validate-validAggs failed, because aggs is not a map", func() {

			res := validateDSL(c, &interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.DSL,
				Formula: `{"size":0,"aggs": {
					"name1": 0
				}}`,
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "The aggregation of dsl is not a map")
		})

		Convey("Validate-validAggs failed, because unsurpport agg type", func() {

			res := validateDSL(c, &interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.DSL,
				Formula: `{"size":0,"aggs": {
					"name1": {
					  "histogram": {
						"field": "price",
						"interval": 50
					  }
					}
				}}`,
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "Unsupport aggregation type in dsl.")
		})

		Convey("Validate-validAggs-validTopHits failed, because top hits unmarshal error", func() {

			res := validateDSL(c, &interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.DSL,
				Formula: `{"size":0,"aggs": {
					"name1": {
					  "top_hits": {
						"size": "1"
					  }
					}
				}}`,
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldContainSubstring, "TopHits Unmarshal error")
		})

		Convey("Validate-validAggs-validTopHits failed, because top hits size lte 0 ", func() {

			res := validateDSL(c, &interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.DSL,
				Formula: `{"size":0,"aggs": {
					"name1": {
					  "top_hits": {
						"size": -1
					  }
					}
				}}`,
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "TopHits's numHits must be > 0")
		})

		Convey("Validate-validAggs-validTopHits failed, because top hits includes is empty ", func() {

			res := validateDSL(c, &interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.DSL,
				Formula: `{"size":0,"aggs": {
					"name1": {
					  "top_hits": {
						"size": 1
					  }
					}
				}}`,
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "TopHits's includes must not be empty")
		})

		Convey("Validate-validAggs-validTopHits failed, because measure field is not one of top hits includes ", func() {

			res := validateDSL(c, &interfaces.MetricModelQuery{
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
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "The measure field must be one of the includes in top_hits.")
		})

		Convey("Validate-validAggs-validTopHits failed, because date_histogram used obsolete interval ", func() {

			res := validateDSL(c, &interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.DSL,
				Formula: `{
					"size": 0,
					"aggs": {
					  "date_histogram_name": {
						"date_histogram": {
						  "field": "@timestamp",
						  "interval": "6m"
						},
						"aggs": {
						  "metric-aggs-name": {
							"value_count": {
							  "field": "@timestamp"
							}
						  }
						}
					  }
					}
				  }`,
				MeasureField: "metric-aggs-name",
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "The interval has been abandoned")
		})

		Convey("Validate-validAggs-validTopHits failed, because date_histogram used fixed_interval and calendar_interval at the same time ", func() {

			res := validateDSL(c, &interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.DSL,
				Formula: `{
					"size": 0,
					"aggs": {
					  "date_histogram_name": {
						"date_histogram": {
						  "field": "@timestamp",
						  "fixed_interval": "6m",
						  "calendar_interval": "6m"
						},
						"aggs": {
						  "metric-aggs-name": {
							"value_count": {
							  "field": "@timestamp"
							}
						  }
						}
					  }
					}
				  }`,
				MeasureField: "metric-aggs-name",
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "The date_histogram aggregation statement is incorrect")
		})

		Convey("Validate failed, because measure field is not one of top hits includes ", func() {

			res := validateDSL(c, &interfaces.MetricModelQuery{
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
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "The measure field must be one of the includes in top_hits.")
		})

		Convey("Validate failed, because agg layers gte 7 ", func() {

			res := validateDSL(c, &interfaces.MetricModelQuery{
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
				MeasureField: "a",
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "The number of aggregation layers in DSL does not exceed 7.")
		})

		Convey("Validate failed, because missing metric aggregation ", func() {

			res := validateDSL(c, &interfaces.MetricModelQuery{
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
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "A metric aggregation should be included in dsl")
		})

		Convey("Validate failed, because measure field is empty ", func() {

			res := validateDSL(c, &interfaces.MetricModelQuery{
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
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_NullParameter_MeasureField)
		})

		Convey("Validate failed, because measure field not aggs name ", func() {

			res := validateDSL(c, &interfaces.MetricModelQuery{
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
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_MeasureField)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "Measure Field should be the name of metric aggregation in dsl")
		})

		Convey("Validate failed, because not contain date histogram ", func() {

			res := validateDSL(c, &interfaces.MetricModelQuery{
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
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "Only one date_histogram aggregation should be included in dsl")
		})

		Convey("Validate failed, because contain more than one date histogram ", func() {

			res := validateDSL(c, &interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,

				QueryType: interfaces.DSL,
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
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "Only one date_histogram aggregation should be included in dsl")
		})

		Convey("Validate failed, because there is no metric aggs under the date histogram ", func() {

			res := validateDSL(c, &interfaces.MetricModelQuery{
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
			})
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_Formula)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "The sub aggregation of date_histogram aggregation needs to be a metric aggregation.")
		})

		Convey("Validate success ", func() {

			res := validateDSL(c, &interfaces.MetricModelQuery{
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
										  "top_hits": {
											"size": 3,
												"sort": [
													{
														"@timestamp": {
														"order": "desc"
														}
													}
												],
													"_source": {
													"includes": [ "labels.cpu","metrics.node_cpu_seconds_total" ]
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
				MeasureField: "metrics.node_cpu_seconds_total",
			})
			So(res, ShouldBeNil)
		})

		Convey("Validate success when calendar_interval", func() {

			res := validateDSL(c, &interfaces.MetricModelQuery{
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
										  "top_hits": {
											"size": 3,
												"sort": [
													{
														"@timestamp": {
														"order": "desc"
														}
													}
												],
													"_source": {
													"includes": [ "labels.cpu","metrics.node_cpu_seconds_total" ]
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
				MeasureField: "metrics.node_cpu_seconds_total",
			})
			So(res, ShouldBeNil)
		})
	})
}

func TestValidateDataViewSimulate(t *testing.T) {
	Convey("Test ValidateDataViewSimulate", t, func() {
		c := context.Background()

		Convey("ValidateDataViewSimulate failed, validateViewDataSource error", func() {
			query := &interfaces.DataViewSimulateQuery{
				ViewQueryCommonParams: interfaces.ViewQueryCommonParams{
					Format: interfaces.Format_Flat,
				},
				// DataSource: map[string]any{"type": "af"},
			}

			err := ValidateDataViewSimulate(c, query)
			So(err, ShouldNotBeNil)
		})

		Convey("ValidateDataViewSimulate failed, validatePaginationSortParams error", func() {
			query := &interfaces.DataViewSimulateQuery{
				// DataSource: map[string]any{
				// 	"type": "index_base",
				// 	"index_base": []any{
				// 		interfaces.SimpleIndexBase{
				// 			BaseType: "x",
				// 		},
				// 	},
				// },
				ViewQueryCommonParams: interfaces.ViewQueryCommonParams{
					Offset: -1,
					Format: interfaces.Format_Flat,
				},
			}

			err := ValidateDataViewSimulate(c, query)
			So(err, ShouldNotBeNil)
		})

		Convey("ValidateDataViewSimulate failed, validateViewFields error", func() {
			query := &interfaces.DataViewSimulateQuery{
				// DataSource: map[string]any{
				// 	"type": "index_base",
				// 	"index_base": []any{
				// 		interfaces.SimpleIndexBase{
				// 			BaseType: "x",
				// 		},
				// 	},
				// },
				// FieldScope: 2,
				ViewQueryCommonParams: interfaces.ViewQueryCommonParams{
					Format: interfaces.Format_Flat,
				},
			}

			err := ValidateDataViewSimulate(c, query)
			So(err, ShouldNotBeNil)
		})

		Convey("ValidateDataViewSimulate failed, validateViewTime error", func() {
			query := &interfaces.DataViewSimulateQuery{
				// DataSource: map[string]any{
				// 	"type": "index_base",
				// 	"index_base": []any{
				// 		interfaces.SimpleIndexBase{
				// 			BaseType: "x",
				// 		},
				// 	},
				// },
				// FieldScope: 1,
				Fields: []*cond.ViewField{},
				ViewQueryCommonParams: interfaces.ViewQueryCommonParams{
					Start:  -1,
					Format: interfaces.Format_Flat,
				},
			}

			err := ValidateDataViewSimulate(c, query)
			So(err, ShouldNotBeNil)
		})

		Convey("ValidateDataViewSimulate failed, filters count exceeds 10", func() {
			query := &interfaces.DataViewSimulateQuery{
				// DataSource: map[string]any{
				// 	"type": "index_base",
				// 	"index_base": []any{
				// 		interfaces.SimpleIndexBase{
				// 			BaseType: "x",
				// 		},
				// 	},
				// },
				// FieldScope: 1,
				// Fields:     []*cond.ViewField{},
				// Condition: &cond.CondCfg{
				// 	Operation: cond.OperationAnd,
				// 	SubConds: []*cond.CondCfg{
				// 		{Name: "xxx", Operation: cond.OperationEq, ValueOptCfg: vopt.ValueOptCfg{Value: 2, ValueFrom: vopt.ValueFrom_Const}},
				// 		{Name: "xxx", Operation: cond.OperationEq, ValueOptCfg: vopt.ValueOptCfg{Value: 2, ValueFrom: vopt.ValueFrom_Const}},
				// 		{Name: "xxx", Operation: cond.OperationEq, ValueOptCfg: vopt.ValueOptCfg{Value: 2, ValueFrom: vopt.ValueFrom_Const}},
				// 		{Name: "xxx", Operation: cond.OperationEq, ValueOptCfg: vopt.ValueOptCfg{Value: 2, ValueFrom: vopt.ValueFrom_Const}},
				// 		{Name: "xxx", Operation: cond.OperationEq, ValueOptCfg: vopt.ValueOptCfg{Value: 2, ValueFrom: vopt.ValueFrom_Const}},
				// 		{Name: "xxx", Operation: cond.OperationEq, ValueOptCfg: vopt.ValueOptCfg{Value: 2, ValueFrom: vopt.ValueFrom_Const}},
				// 		{Name: "xxx", Operation: cond.OperationEq, ValueOptCfg: vopt.ValueOptCfg{Value: 2, ValueFrom: vopt.ValueFrom_Const}},
				// 		{Name: "xxx", Operation: cond.OperationEq, ValueOptCfg: vopt.ValueOptCfg{Value: 2, ValueFrom: vopt.ValueFrom_Const}},
				// 		{Name: "xxx", Operation: cond.OperationEq, ValueOptCfg: vopt.ValueOptCfg{Value: 2, ValueFrom: vopt.ValueFrom_Const}},
				// 		{Name: "xxx", Operation: cond.OperationEq, ValueOptCfg: vopt.ValueOptCfg{Value: 2, ValueFrom: vopt.ValueFrom_Const}},
				// 		{Name: "xxx", Operation: cond.OperationEq, ValueOptCfg: vopt.ValueOptCfg{Value: 2, ValueFrom: vopt.ValueFrom_Const}},
				// 	},
				// },
				// ViewQueryCommonParams: interfaces.ViewQueryCommonParams{
				// 	Start:  1672502400000,
				// 	End:    1672502500000,
				// 	Format: interfaces.Format_Flat,
				// },
			}

			err := ValidateDataViewSimulate(c, query)
			So(err, ShouldNotBeNil)
		})

		Convey("ValidateDataViewSimulate failed, validateFilters error", func() {
			query := &interfaces.DataViewSimulateQuery{
				// DataSource: map[string]any{
				// 	"type": "index_base",
				// 	"index_base": []any{
				// 		interfaces.SimpleIndexBase{
				// 			BaseType: "x",
				// 		},
				// 	},
				// },
				// FieldScope: 1,
				// Fields:     []*cond.ViewField{},
				// Condition: &cond.CondCfg{
				// 	Name: "xxx",
				// },
				ViewQueryCommonParams: interfaces.ViewQueryCommonParams{
					Start:  1672502400000,
					End:    1672502500000,
					Format: interfaces.Format_Flat,
				},
			}

			err := ValidateDataViewSimulate(c, query)
			So(err, ShouldNotBeNil)
		})

		Convey("ValidateDataViewSimulate success", func() {
			query := &interfaces.DataViewSimulateQuery{
				// DataSource: map[string]any{
				// 	"type": "index_base",
				// 	"index_base": []any{
				// 		interfaces.SimpleIndexBase{BaseType: "x"},
				// 	},
				// },
				// FieldScope: 1,
				Fields: []*cond.ViewField{},
				ViewQueryCommonParams: interfaces.ViewQueryCommonParams{
					Start:  1672502400000,
					End:    1672502500000,
					Format: interfaces.Format_Flat,
					Limit:  10,
				},
			}

			err := ValidateDataViewSimulate(c, query)
			So(err, ShouldBeNil)
		})
	})
}

func TestValidateDataViewQueryV1(t *testing.T) {
	Convey("Test ValidateDataViewQueryV1", t, func() {
		c := context.Background()

		Convey("ValidateDataViewQuery failed, validateViewTime error", func() {
			query := &interfaces.DataViewQueryV1{
				ViewQueryCommonParams: interfaces.ViewQueryCommonParams{
					Start:  1735574400000,
					End:    1735573400000,
					Format: interfaces.Format_Flat,
				},
			}

			err := ValidateDataViewQueryV1(c, query)
			So(err, ShouldNotBeNil)
		})

		Convey("ValidateDataViewQuery failed, validateViewFilters error", func() {
			query := &interfaces.DataViewQueryV1{
				GlobalFilters: &cond.CondCfg{
					Name: "x",
				},
				ViewQueryCommonParams: interfaces.ViewQueryCommonParams{
					Start:  1672502400000,
					End:    1672502500000,
					Format: interfaces.Format_Flat,
				},
			}

			err := ValidateDataViewQueryV1(c, query)
			So(err, ShouldNotBeNil)
		})

		Convey("ValidateDataViewQuery failed, validatePaginationSortParams error", func() {
			query := &interfaces.DataViewQueryV1{
				ViewQueryCommonParams: interfaces.ViewQueryCommonParams{
					Offset: -1,
					Start:  1672502400000,
					End:    1672502500000,
					Format: interfaces.Format_Flat,
				},
			}

			err := ValidateDataViewQueryV1(c, query)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestValidateDataViewQueryV2(t *testing.T) {
	Convey("Test ValidateDataViewQueryV2", t, func() {
		c := context.Background()

		Convey("ValidateDataViewQuery failed, validateViewTime error", func() {
			query := &interfaces.DataViewQueryV2{
				ViewQueryCommonParams: interfaces.ViewQueryCommonParams{
					Start:  1735574400000,
					End:    1735573400000,
					Format: interfaces.Format_Flat,
				},
			}

			err := ValidateDataViewQueryV2(c, query)
			So(err, ShouldNotBeNil)
		})

		Convey("ValidateDataViewQuery failed, validateViewFilters error", func() {
			query := &interfaces.DataViewQueryV2{
				GlobalFilters: map[string]any{
					"field": "x",
				},
				// GlobalFilters: &cond.CondCfg{
				// 	Name: "x",
				// },
				ViewQueryCommonParams: interfaces.ViewQueryCommonParams{
					Start:  1672502400000,
					End:    1672502500000,
					Format: interfaces.Format_Flat,
				},
			}

			err := ValidateDataViewQueryV2(c, query)
			So(err, ShouldNotBeNil)
		})

		Convey("ValidateDataViewQuery failed, validatePaginationSortParams error", func() {
			query := &interfaces.DataViewQueryV2{
				ViewQueryCommonParams: interfaces.ViewQueryCommonParams{
					Offset: -1,
					Start:  1672502400000,
					End:    1672502500000,
					Format: interfaces.Format_Flat,
				},
			}

			err := ValidateDataViewQueryV2(c, query)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestValidateViewFormat(t *testing.T) {
	Convey("Test ValidateViewFormat", t, func() {

		Convey("Validate failed, because format is not original or flat", func() {
			format := "plus"
			res := validateFormat(testCtx, format)
			So(res, ShouldNotBeNil)
		})

		Convey("Validate success with format flat", func() {
			format := interfaces.Format_Flat
			res := validateFormat(testCtx, format)
			So(res, ShouldBeNil)
		})

		Convey("Validate success with format original", func() {
			format := interfaces.Format_Original
			res := validateFormat(testCtx, format)
			So(res, ShouldBeNil)
		})
	})
}

func TestValidateViewDataSource(t *testing.T) {
	Convey("Test ValidateViewDataSource", t, func() {
		c := context.Background()

		Convey("Validate failed, because datasource type is null", func() {
			dataSource := map[string]any{}

			res := validateViewDataSource(c, dataSource, interfaces.FieldScope_All)

			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_DataView_InvalidParameter_DataSource)
		})

		Convey("Validate failed, because datasource type is not string", func() {
			dataSource := map[string]any{"type": 2}
			res := validateViewDataSource(c, dataSource, interfaces.FieldScope_All)

			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_DataView_InvalidParameter_DataSource)
		})

		Convey("Validate failed, because no dataSource index_base", func() {
			dataSource := map[string]any{
				"type": "index_base",
			}

			res := validateViewDataSource(c, dataSource, interfaces.FieldScope_All)

			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_DataView_InvalidParameter_DataSource)
		})

		Convey("Validate failed, because dataSource index_base count is 0", func() {
			dataSource := map[string]any{
				"type":       "index_base",
				"index_base": []any{},
			}

			res := validateViewDataSource(c, dataSource, interfaces.FieldScope_All)

			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_DataView_InvalidParameter_DataSource)
		})

		Convey("Validate failed, because field_scope is 1 and only support one index base", func() {
			dataSource := map[string]any{
				"type": "index_base",
				"index_base": []any{
					interfaces.SimpleIndexBase{Name: "x"},
					interfaces.SimpleIndexBase{Name: "y"},
				},
			}

			res := validateViewDataSource(c, dataSource, interfaces.FieldScope_All)

			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_DataView_InvalidParameter_DataSource)
		})

		Convey("Validate failed, because index base names is not a list", func() {
			dataSource := map[string]any{
				"type":       "index_base",
				"index_base": map[string]string{},
			}

			res := validateViewDataSource(c, dataSource, interfaces.FieldScope_All)

			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_DataView_InvalidParameter_DataSource)
		})

		Convey("Validate failed, because unsupport datasource type", func() {
			dataSource := map[string]any{"type": "af"}

			res := validateViewDataSource(c, dataSource, interfaces.FieldScope_All)

			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_DataView_UnsupportDataSourceType)
		})

		Convey("Validate success", func() {
			dataSource := map[string]any{
				"type": "index_base",
				"index_base": []any{
					interfaces.SimpleIndexBase{Name: "x"},
				},
			}

			res := validateViewDataSource(c, dataSource, interfaces.FieldScope_All)

			So(res, ShouldBeNil)
		})
	})
}

func TestValidateViewTime(t *testing.T) {
	Convey("Test ValidateViewTime", t, func() {

		Convey("Validate failed, because start < 0", func() {
			start := int64(-1)
			end := int64(1672502500000)
			res := validateViewTime(testCtx, start, end)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_InvalidParameter_Start)
		})

		Convey("Validate failed, because end < 0", func() {
			start := int64(1672502400000)
			end := int64(-2)
			res := validateViewTime(testCtx, start, end)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_InvalidParameter_End)
		})

		Convey("Validate failed, because start after current", func() {
			start := int64(2672502400000)
			end := int64(1672502500000)
			res := validateViewTime(testCtx, start, end)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_InvalidParameter_Start)
		})

		Convey("Validate failed, because end after current", func() {
			start := int64(1672502400000)
			end := int64(2672502500000)
			res := validateViewTime(testCtx, start, end)
			So(res, ShouldBeNil)
		})

		Convey("Validate failed, because end before start", func() {
			start := int64(1672502400000)
			end := int64(1572502500000)
			res := validateViewTime(testCtx, start, end)

			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_InvalidParameter_Start)

		})
	})
}

func TestValidatePaginationParams(t *testing.T) {
	Convey("Test ValidatePaginationParams", t, func() {

		Convey("Validate failed, because offset <= 0", func() {
			offset := -1
			limit := 10
			res := validatePaginationParams(testCtx, offset, limit)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_InvalidParameter_Offset)
		})

		Convey("Validate failed, because limit out of range 1", func() {
			offset := 0
			limit := -1
			res := validatePaginationParams(testCtx, offset, limit)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_InvalidParameter_Limit)
		})

		Convey("Validate failed, because limit out of range 2", func() {
			offset := 0
			limit := -1
			res := validatePaginationParams(testCtx, offset, limit)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_InvalidParameter_Limit)
		})

		Convey("Validate failed, because limit out of range 3", func() {
			offset := 0
			limit := 10001
			res := validatePaginationParams(testCtx, offset, limit)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_InvalidParameter_Limit)
		})

		Convey("Validate failed, because limit+offset > 10000", func() {
			offset := 500
			limit := 9600
			res := validatePaginationParams(testCtx, offset, limit)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_InvalidParameter_Limit)
		})

		Convey("Validate success", func() {
			offset := 0
			limit := 10
			res := validatePaginationParams(testCtx, offset, limit)

			So(res, ShouldBeNil)
		})
	})
}

func TestValidateSortParamsV1(t *testing.T) {
	Convey("Test ValidateSortParamsV1", t, func() {

		Convey("Validate failed, because direction not asc or desc", func() {
			direction := "plus"
			err := validateSortParamsV1(testCtx, direction)
			So(err, ShouldNotBeNil)
		})

		Convey("Validate success, asc", func() {
			direction := "asc"
			err := validateSortParamsV1(testCtx, direction)
			So(err, ShouldBeNil)
		})

		Convey("Validate success, desc", func() {
			direction := "desc"
			err := validateSortParamsV1(testCtx, direction)
			So(err, ShouldBeNil)
		})
	})
}

func TestValidateSortParamsV2(t *testing.T) {
	Convey("Test ValidateSortParamsV2", t, func() {

		Convey("Validate failed, because direction not asc or desc", func() {
			sort := []*interfaces.SortParamsV2{
				{Direction: "plus"},
			}
			err := validateSortParamsV2(testCtx, sort)
			So(err, ShouldNotBeNil)
		})

		Convey("Validate success, asc", func() {
			sort := []*interfaces.SortParamsV2{
				{Direction: "asc"},
			}
			err := validateSortParamsV2(testCtx, sort)
			So(err, ShouldBeNil)
		})

		Convey("Validate success, desc", func() {
			sort := []*interfaces.SortParamsV2{
				{Direction: "desc"},
			}
			err := validateSortParamsV2(testCtx, sort)
			So(err, ShouldBeNil)
		})
	})
}

func TestValidateCond(t *testing.T) {
	Convey("Test Validate condition", t, func() {

		Convey("Validate failed, because operation is null", func() {
			cfg := &cond.CondCfg{
				Name:      "xxx",
				Operation: "",
			}

			err := validateCond(testCtx, cfg)

			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_NullParameter_FilterOperation)
		})

		Convey("unsupported operation", func() {
			cfg := &cond.CondCfg{
				Operation: "unsupported",
			}
			err := validateCond(testCtx, cfg)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("subConds recur validate recursively", func() {
			cfg := &cond.CondCfg{
				Operation: cond.OperationAnd,
				SubConds: []*cond.CondCfg{
					{
						Name:      "",
						Operation: cond.OperationExist,
					},
				},
			}
			err := validateCond(testCtx, cfg)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("value_from is not const", func() {
			cfg := &cond.CondCfg{
				Operation: cond.OperationRange,
				Name:      "name",
				ValueOptCfg: vopt.ValueOptCfg{
					Value:     "value",
					ValueFrom: vopt.ValueFrom_Field,
				},
			}
			err := validateCond(testCtx, cfg)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("invalid value length for lte operation", func() {
			cfg := &cond.CondCfg{
				Operation: cond.OperationLte,
				Name:      "name",
				ValueOptCfg: vopt.ValueOptCfg{
					Value:     []interface{}{1},
					ValueFrom: vopt.ValueFrom_Const,
				},
			}
			err := validateCond(testCtx, cfg)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("invalid value type for like operation", func() {
			cfg := &cond.CondCfg{
				Operation: cond.OperationLike,
				Name:      "name",
				ValueOptCfg: vopt.ValueOptCfg{
					Value:     true,
					ValueFrom: vopt.ValueFrom_Const,
				},
			}
			err := validateCond(testCtx, cfg)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("invalid value type for regex operation", func() {
			cfg := &cond.CondCfg{
				Operation: cond.OperationRegex,
				Name:      "name",
				ValueOptCfg: vopt.ValueOptCfg{
					Value:     true,
					ValueFrom: vopt.ValueFrom_Const,
				},
			}
			err := validateCond(testCtx, cfg)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("invalid regex expression for regex operation", func() {
			cfg := &cond.CondCfg{
				Operation: cond.OperationRegex,
				Name:      "name",
				ValueOptCfg: vopt.ValueOptCfg{
					Value:     "*",
					ValueFrom: vopt.ValueFrom_Const,
				},
			}
			err := validateCond(testCtx, cfg)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("invalid value type for in operation", func() {
			cfg := &cond.CondCfg{
				Operation: cond.OperationIn,
				Name:      "name",
				ValueOptCfg: vopt.ValueOptCfg{
					Value:     3,
					ValueFrom: vopt.ValueFrom_Const,
				},
			}
			err := validateCond(testCtx, cfg)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("invalid value length for in operation", func() {
			cfg := &cond.CondCfg{
				Operation: cond.OperationIn,
				Name:      "name",
				ValueOptCfg: vopt.ValueOptCfg{
					Value:     []any{},
					ValueFrom: vopt.ValueFrom_Const,
				},
			}
			err := validateCond(testCtx, cfg)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("invalid value type for range operation", func() {
			cfg := &cond.CondCfg{
				Operation: cond.OperationRange,
				Name:      "name",
				ValueOptCfg: vopt.ValueOptCfg{
					Value:     1,
					ValueFrom: vopt.ValueFrom_Const,
				},
			}
			err := validateCond(testCtx, cfg)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("invalid value length for out_range operation", func() {
			cfg := &cond.CondCfg{
				Operation: cond.OperationOutRange,
				Name:      "name",
				ValueOptCfg: vopt.ValueOptCfg{
					Value:     []any{1, 2, 3},
					ValueFrom: vopt.ValueFrom_Const,
				},
			}
			err := validateCond(testCtx, cfg)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
		})
	})
}

func TestValidateParamsWhenGetSpanList(t *testing.T) {
	Convey("Test validateParamsWhenGetSpanList", t, func() {

		Convey("Validate failed, caused by the error from func 'validatePaginationQueryParams'", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_Offset)

			patch := ApplyFuncReturn(validatePaginationQueryParams, interfaces.PaginationQueryParams{}, expectedErr)
			defer patch.Reset()

			_, err := validateParamsWhenGetSpanList(testCtx, interfaces.SpanListQueryParams{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Validate failed, caused by the error from func 'validateCond'", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_Offset)

			patch1 := ApplyFuncReturn(validatePaginationQueryParams, interfaces.PaginationQueryParams{}, nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateCond, expectedErr)
			defer patch2.Reset()

			_, err := validateParamsWhenGetSpanList(testCtx, interfaces.SpanListQueryParams{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Validate succeed", func() {
			patch1 := ApplyFuncReturn(validatePaginationQueryParams, interfaces.PaginationQueryParams{}, nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateCond, nil)
			defer patch2.Reset()

			_, err := validateParamsWhenGetSpanList(testCtx, interfaces.SpanListQueryParams{})
			So(err, ShouldBeNil)
		})
	})
}

func TestValidateParamsWhenGetRelatedLogList(t *testing.T) {
	Convey("Test validateParamsWhenGetRelatedLogList", t, func() {

		Convey("Validate failed, caused by the error from func 'validatePaginationQueryParams'", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_Offset)

			patch := ApplyFuncReturn(validatePaginationQueryParams, interfaces.PaginationQueryParams{}, expectedErr)
			defer patch.Reset()

			_, err := validateParamsWhenGetRelatedLogList(testCtx, interfaces.RelatedLogListQueryParams{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Validate failed, caused by the error from func 'validateCond'", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_Offset)

			patch1 := ApplyFuncReturn(validatePaginationQueryParams, interfaces.PaginationQueryParams{}, nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateCond, expectedErr)
			defer patch2.Reset()

			_, err := validateParamsWhenGetRelatedLogList(testCtx, interfaces.RelatedLogListQueryParams{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Validate succeed", func() {
			patch1 := ApplyFuncReturn(validatePaginationQueryParams, interfaces.PaginationQueryParams{}, nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateCond, nil)
			defer patch2.Reset()

			_, err := validateParamsWhenGetRelatedLogList(testCtx, interfaces.RelatedLogListQueryParams{})
			So(err, ShouldBeNil)
		})
	})
}

func TestValidateSearchAfterAndPit(t *testing.T) {
	Convey("Test validateSearchAfterAndPit", t, func() {

		Convey("Validate failed, because use search_after and offset is not 0", func() {
			offset := 1
			err := validateSearchAfterAndPit(testCtx, interfaces.SearchAfterParams{
				SearchAfter: []any{"1", "2"},
			}, offset)
			So(err, ShouldNotBeNil)
		})

		Convey("Validate succeed, because pit_keep_alive is valid", func() {
			offset := 0
			err := validateSearchAfterAndPit(testCtx, interfaces.SearchAfterParams{
				PitKeepAlive: "2haha",
			}, offset)
			So(err, ShouldNotBeNil)
		})

		Convey("Validate failed, because pit_keep_alive is invalid", func() {
			offset := 0
			err := validateSearchAfterAndPit(testCtx, interfaces.SearchAfterParams{
				PitKeepAlive: "25h",
			}, offset)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestValidateObjectiveModelimulate(t *testing.T) {
	Convey("Test ValidateObjectiveModelimulate", t, func() {
		c := context.Background()
		common.FixedStepsMap = StepsMap

		Convey("Validate failed, because objective type is null", func() {
			res := ValidateObjectiveModelSimulate(c, &interfaces.ObjectiveModelQuery{})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_ObjectiveModel_NullParameter_ObjectiveType)
		})

		Convey("Validate failed, because objective type is unsupported", func() {
			res := ValidateObjectiveModelSimulate(c, &interfaces.ObjectiveModelQuery{
				ObjectiveType: "invalid",
			})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_ObjectiveModel_UnsupportObjectiveType)
		})

		Convey("Validate failed, because objective config is null", func() {
			res := ValidateObjectiveModelSimulate(c, &interfaces.ObjectiveModelQuery{
				ObjectiveType: interfaces.SLO,
			})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_ObjectiveModel_NullParameter_ObjectiveConfig)
		})

		Convey("Validate failed, because objective is null", func() {
			res := ValidateObjectiveModelSimulate(c, &interfaces.ObjectiveModelQuery{
				ObjectiveType:   interfaces.SLO,
				ObjectiveConfig: interfaces.SLOObjective{},
			})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_ObjectiveModel_NullParameter_Objective)
		})

		Convey("Validate failed, because objective is invalid", func() {
			objective := 0.0
			res := ValidateObjectiveModelSimulate(c, &interfaces.ObjectiveModelQuery{
				ObjectiveType: interfaces.SLO,
				ObjectiveConfig: interfaces.SLOObjective{
					Objective: &objective,
				},
			})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_ObjectiveModel_InvalidParameter_Objective)
		})

		Convey("Validate failed, because period is null", func() {
			objective := 99.9
			res := ValidateObjectiveModelSimulate(c, &interfaces.ObjectiveModelQuery{
				ObjectiveType: interfaces.SLO,
				ObjectiveConfig: interfaces.SLOObjective{
					Objective: &objective,
				},
			})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_ObjectiveModel_NullParameter_Period)
		})

		Convey("Validate failed, because good metric model is null", func() {
			objective := 99.9
			period := int64(90)
			res := ValidateObjectiveModelSimulate(c, &interfaces.ObjectiveModelQuery{
				ObjectiveType: interfaces.SLO,
				ObjectiveConfig: interfaces.SLOObjective{
					Objective: &objective,
					Period:    &period,
				},
			})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_ObjectiveModel_NullParameter_GoodMetricModel)
		})

		Convey("Validate failed, because good metric model id is null", func() {
			objective := 99.9
			period := int64(90)
			res := ValidateObjectiveModelSimulate(c, &interfaces.ObjectiveModelQuery{
				ObjectiveType: interfaces.SLO,
				ObjectiveConfig: interfaces.SLOObjective{
					Objective:       &objective,
					Period:          &period,
					GoodMetricModel: &interfaces.BundleMetricModel{},
				},
			})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_ObjectiveModel_NullParameter_GoodMetricModelID)
		})

		Convey("Validate failed, because total metric model is null", func() {
			objective := 99.9
			period := int64(90)
			res := ValidateObjectiveModelSimulate(c, &interfaces.ObjectiveModelQuery{
				ObjectiveType: interfaces.SLO,
				ObjectiveConfig: interfaces.SLOObjective{
					Objective:       &objective,
					Period:          &period,
					GoodMetricModel: &interfaces.BundleMetricModel{Id: "1"},
				},
			})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_ObjectiveModel_NullParameter_TotalMetricModel)
		})

		Convey("Validate failed, because total metric model id is null", func() {
			objective := 99.9
			period := int64(90)
			res := ValidateObjectiveModelSimulate(c, &interfaces.ObjectiveModelQuery{
				ObjectiveType: interfaces.SLO,
				ObjectiveConfig: interfaces.SLOObjective{
					Objective:        &objective,
					Period:           &period,
					GoodMetricModel:  &interfaces.BundleMetricModel{Id: "1"},
					TotalMetricModel: &interfaces.BundleMetricModel{},
				},
			})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_ObjectiveModel_NullParameter_TotalMetricModelID)
		})

		Convey("Validate failed, because status config total count exceeded", func() {
			objective := 99.9
			period := int64(90)
			res := ValidateObjectiveModelSimulate(c, &interfaces.ObjectiveModelQuery{
				ObjectiveType: interfaces.SLO,
				ObjectiveConfig: interfaces.SLOObjective{
					Objective:        &objective,
					Period:           &period,
					GoodMetricModel:  &interfaces.BundleMetricModel{Id: "1"},
					TotalMetricModel: &interfaces.BundleMetricModel{Id: "2"},
					StatusConfig: &interfaces.ObjectiveStatusConfig{
						Ranges: []interfaces.Range{
							{}, {}, {}, {}, {}, {}, {}, {}, {}, {}, // 10 items
							{}, {}, {}, {}, {}, {}, {}, {}, {}, {}, // 20 items
							{}, // 21 items exceeds limit
						},
					},
				},
			})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_ObjectiveModel_CountExceeded_StatusTotal)
		})

		Convey("Validate failed, because objective config marshal error", func() {
			res := ValidateObjectiveModelSimulate(c, &interfaces.ObjectiveModelQuery{
				ObjectiveType:   interfaces.SLO,
				ObjectiveConfig: make(chan int), // Invalid type that can't be marshaled to JSON
			})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_ObjectiveModel_InvalidParameter_ObjectiveConfig)
		})

		Convey("Validate failed, because objective config unmarshal error", func() {
			res := ValidateObjectiveModelSimulate(c, &interfaces.ObjectiveModelQuery{
				ObjectiveType: interfaces.SLO,
				ObjectiveConfig: map[string]interface{}{
					"objective":          "invalid", // Invalid type that can't be unmarshaled to float64 pointer
					"period":             90,
					"good_metric_model":  map[string]string{"id": "1"},
					"total_metric_model": map[string]string{"id": "2"},
				},
			})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_ObjectiveModel_InvalidParameter_ObjectiveConfig)
		})

		Convey("Validate failed, because KPI objective config marshal error", func() {
			res := ValidateObjectiveModelSimulate(c, &interfaces.ObjectiveModelQuery{
				ObjectiveType:   interfaces.KPI,
				ObjectiveConfig: make(chan int), // Invalid type that can't be marshaled to JSON
			})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_ObjectiveModel_InvalidParameter_ObjectiveConfig)
		})

		Convey("Validate failed, because KPI objective config unmarshal error", func() {
			res := ValidateObjectiveModelSimulate(c, &interfaces.ObjectiveModelQuery{
				ObjectiveType: interfaces.KPI,
				ObjectiveConfig: map[string]interface{}{
					"objective": "invalid", // Invalid type that can't be unmarshaled to float64 pointer
					"unit":      interfaces.UNIT_NUM_NONE,
					"comprehensive_metric_models": []map[string]interface{}{
						{
							"id":     "1",
							"weight": 100,
						},
					},
				},
			})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_ObjectiveModel_InvalidParameter_ObjectiveConfig)
		})

		Convey("Validate failed, because KPI objective is null", func() {
			res := ValidateObjectiveModelSimulate(c, &interfaces.ObjectiveModelQuery{
				ObjectiveType: interfaces.KPI,
				ObjectiveConfig: interfaces.KPIObjective{
					Unit: interfaces.UNIT_NUM_NONE,
				},
			})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_ObjectiveModel_NullParameter_Objective)
		})

		Convey("Validate failed, because KPI objective is invalid", func() {
			objective := 0.0
			res := ValidateObjectiveModelSimulate(c, &interfaces.ObjectiveModelQuery{
				ObjectiveType: interfaces.KPI,
				ObjectiveConfig: interfaces.KPIObjective{
					Objective: &objective,
					Unit:      interfaces.UNIT_NUM_NONE,
				},
			})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_ObjectiveModel_InvalidParameter_Objective)
		})

		Convey("Validate failed, because KPI objective unit is invalid", func() {
			objective := 99.9
			res := ValidateObjectiveModelSimulate(c, &interfaces.ObjectiveModelQuery{
				ObjectiveType: interfaces.KPI,
				ObjectiveConfig: interfaces.KPIObjective{
					Objective: &objective,
					Unit:      "invalid_unit",
				},
			})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_ObjectiveModel_InvalidParameter_ObjectiveUnit)
		})

		Convey("Validate failed, because KPI objective unit is null", func() {
			objective := 99.9
			res := ValidateObjectiveModelSimulate(c, &interfaces.ObjectiveModelQuery{
				ObjectiveType: interfaces.KPI,
				ObjectiveConfig: interfaces.KPIObjective{
					Objective: &objective,
				},
			})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_ObjectiveModel_NullParameter_ObjectiveUnit)
		})

		Convey("Validate failed, because KPI comprehensive_metric_models is null", func() {
			objective := 99.9
			res := ValidateObjectiveModelSimulate(c, &interfaces.ObjectiveModelQuery{
				ObjectiveType: interfaces.KPI,
				ObjectiveConfig: interfaces.KPIObjective{
					Objective: &objective,
					Unit:      interfaces.UNIT_NUM_NONE,
				},
			})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_ObjectiveModel_NullParameter_ComprehensiveMetricModels)
		})

		Convey("Validate failed, because KPI comprehensive_metric_models count exceeded", func() {
			objective := 99.9
			weight := int64(10)
			models := make([]interfaces.ComprehensiveMetricModel, 11) // Create 11 models to exceed limit of 10
			for i := range models {
				models[i] = interfaces.ComprehensiveMetricModel{
					Id:     fmt.Sprintf("id_%d", i),
					Weight: &weight,
				}
			}
			res := ValidateObjectiveModelSimulate(c, &interfaces.ObjectiveModelQuery{
				ObjectiveType: interfaces.KPI,
				ObjectiveConfig: interfaces.KPIObjective{
					Objective:                 &objective,
					Unit:                      interfaces.UNIT_NUM_NONE,
					ComprehensiveMetricModels: models,
				},
			})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_ObjectiveModel_CountExceeded_ComprehensiveMetricModelsTotal)
		})

		Convey("Validate failed, because KPI comprehensive_metric_model id is null", func() {
			objective := 99.9
			weight := int64(10)
			res := ValidateObjectiveModelSimulate(c, &interfaces.ObjectiveModelQuery{
				ObjectiveType: interfaces.KPI,
				ObjectiveConfig: interfaces.KPIObjective{
					Objective: &objective,
					Unit:      interfaces.UNIT_NUM_NONE,
					ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
						{
							Id:     "",
							Weight: &weight,
						},
					},
				},
			})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_ObjectiveModel_NullParameter_ComprehensiveMetricModelID)
		})

		Convey("Validate failed, because KPI comprehensive_metric_model weight is null", func() {
			objective := 99.9
			res := ValidateObjectiveModelSimulate(c, &interfaces.ObjectiveModelQuery{
				ObjectiveType: interfaces.KPI,
				ObjectiveConfig: interfaces.KPIObjective{
					Objective: &objective,
					Unit:      interfaces.UNIT_NUM_NONE,
					ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
						{
							Id:     "test_id",
							Weight: nil,
						},
					},
				},
			})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_ObjectiveModel_NullParameter_ComprehensiveWeight)
		})

		Convey("Validate failed, because KPI comprehensive_metric_model weight is invalid", func() {
			objective := 99.9
			weight := int64(0)
			res := ValidateObjectiveModelSimulate(c, &interfaces.ObjectiveModelQuery{
				ObjectiveType: interfaces.KPI,
				ObjectiveConfig: interfaces.KPIObjective{
					Objective: &objective,
					Unit:      interfaces.UNIT_NUM_NONE,
					ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
						{
							Id:     "test_id",
							Weight: &weight,
						},
					},
				},
			})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_ObjectiveModel_InvalidParameter_ComprehensiveWeight)
		})

		Convey("Validate failed, because KPI comprehensive_metric_model weights sum is not 100", func() {
			objective := 99.9
			weight1 := int64(30)
			weight2 := int64(50)
			res := ValidateObjectiveModelSimulate(c, &interfaces.ObjectiveModelQuery{
				ObjectiveType: interfaces.KPI,
				ObjectiveConfig: interfaces.KPIObjective{
					Objective: &objective,
					Unit:      interfaces.UNIT_NUM_NONE,
					ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
						{
							Id:     "test_id1",
							Weight: &weight1,
						},
						{
							Id:     "test_id2",
							Weight: &weight2,
						},
					},
				},
			})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_ObjectiveModel_InvalidParameter_ComprehensiveWeight)
		})

		Convey("Validate failed, because KPI comprehensive_metric_model count exceeds limit", func() {
			objective := 99.9
			weight := int64(20)
			models := make([]interfaces.ComprehensiveMetricModel, 11)
			for i := 0; i < 11; i++ {
				models[i] = interfaces.ComprehensiveMetricModel{
					Id:     fmt.Sprintf("test_id%d", i),
					Weight: &weight,
				}
			}
			res := ValidateObjectiveModelSimulate(c, &interfaces.ObjectiveModelQuery{
				ObjectiveType: interfaces.KPI,
				ObjectiveConfig: interfaces.KPIObjective{
					Objective:                 &objective,
					Unit:                      interfaces.UNIT_NUM_NONE,
					ComprehensiveMetricModels: models,
				},
			})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_ObjectiveModel_CountExceeded_ComprehensiveMetricModelsTotal)
		})

		Convey("Validate failed, because KPI additional_metric_model id is null", func() {
			objective := 99.9
			weight := int64(100)
			res := ValidateObjectiveModelSimulate(c, &interfaces.ObjectiveModelQuery{
				ObjectiveType: interfaces.KPI,
				ObjectiveConfig: interfaces.KPIObjective{
					Objective: &objective,
					Unit:      interfaces.UNIT_NUM_NONE,
					ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
						{
							Id:     "test_id",
							Weight: &weight,
						},
					},
					AdditionalMetricModels: []interfaces.BundleMetricModel{
						{}, // Empty model with no ID
					},
				},
			})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_ObjectiveModel_NullParameter_AdditionalMetricModelID)
		})

		Convey("Validate failed, because KPI status ranges are invalid", func() {
			objective := 99.9
			weight := int64(100)
			res := ValidateObjectiveModelSimulate(c, &interfaces.ObjectiveModelQuery{
				ObjectiveType: interfaces.KPI,
				ObjectiveConfig: interfaces.KPIObjective{
					Objective: &objective,
					Unit:      interfaces.UNIT_NUM_NONE,
					ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
						{
							Id:     "test_id",
							Weight: &weight,
						},
					},
					StatusConfig: &interfaces.ObjectiveStatusConfig{
						Ranges: []interfaces.Range{
							{
								Status: "",
							},
						},
					},
				},
			})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_ObjectiveModel_InvalidParameter_StatusRanges)
		})

		Convey("Validate failed, because KPI look_back_delta is invalid", func() {
			objective := 99.9
			weight := int64(100)
			res := ValidateObjectiveModelSimulate(c, &interfaces.ObjectiveModelQuery{
				ObjectiveType: interfaces.KPI,
				ObjectiveConfig: interfaces.KPIObjective{
					Objective: &objective,
					Unit:      interfaces.UNIT_NUM_NONE,
					ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
						{
							Id:     "test_id",
							Weight: &weight,
						},
					},
				},
				QueryTimeParams: interfaces.QueryTimeParams{
					IsInstantQuery:   true,
					LookBackDeltaStr: "invalid",
				},
			})
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_LookBackDelta)
		})

		Convey("Validate success", func() {
			objective := 99.9
			period := int64(90)
			query := interfaces.ObjectiveModelQuery{
				ObjectiveType: interfaces.SLO,
				ObjectiveConfig: interfaces.SLOObjective{
					Objective:        &objective,
					Period:           &period,
					GoodMetricModel:  &interfaces.BundleMetricModel{Id: "1"},
					TotalMetricModel: &interfaces.BundleMetricModel{Id: "2"},
				},
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &start3,
					End:     &end3,
					StepStr: &step_5m,
				},
			}
			err := ValidateObjectiveModelSimulate(c, &query)
			So(err, ShouldBeNil)
		})
	})
}

func TestValidateRequestMetrics(t *testing.T) {
	Convey("Test validateRequestMetrics", t, func() {
		ctx := context.Background()

		Convey("RequestMetrics is nil", func() {
			query := &interfaces.MetricModelQuery{}
			err := validateRequestMetrics(ctx, query)
			So(err, ShouldBeNil)
		})

		Convey("RequestMetrics.Type is empty", func() {
			query := &interfaces.MetricModelQuery{
				RequestMetrics: &interfaces.RequestMetrics{},
			}
			err := validateRequestMetrics(ctx, query)
			So(err, ShouldNotBeNil)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_NullParameter_RequestMetricsType)
		})

		Convey("RequestMetrics.Type is invalid", func() {
			query := &interfaces.MetricModelQuery{
				RequestMetrics: &interfaces.RequestMetrics{
					Type: "invalid_type",
				},
			}
			err := validateRequestMetrics(ctx, query)
			So(err, ShouldNotBeNil)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_RequestMetricsType)
		})

		Convey("RequestMetrics.Type is METRICS_SAMEPERIOD but SamePeriodCfg is nil", func() {
			query := &interfaces.MetricModelQuery{
				RequestMetrics: &interfaces.RequestMetrics{
					Type: interfaces.METRICS_SAMEPERIOD,
				},
			}
			err := validateRequestMetrics(ctx, query)
			So(err, ShouldNotBeNil)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_NullParameter_SamePeriodCfg)
		})

		Convey("RequestMetrics.Type is METRICS_SAMEPERIOD but SamePeriodCfg.Method is empty", func() {
			query := &interfaces.MetricModelQuery{
				RequestMetrics: &interfaces.RequestMetrics{
					Type: interfaces.METRICS_SAMEPERIOD,
					SamePeriodCfg: &interfaces.SamePeriodCfg{
						Method:          []string{},
						Offset:          1,
						TimeGranularity: interfaces.METRICS_SAMEPERIOD_TIME_GRANULARITY_DAY,
					},
				},
			}
			err := validateRequestMetrics(ctx, query)
			So(err, ShouldBeNil)
			So(query.RequestMetrics.SamePeriodCfg.Method, ShouldEqual, []string{interfaces.METRICS_SAMEPERIOD_METHOD_GROWTH_VALUE, interfaces.METRICS_SAMEPERIOD_METHOD_GROWTH_RATE})
		})

		Convey("RequestMetrics.Type is METRICS_SAMEPERIOD but SamePeriodCfg.Method has invalid value", func() {
			query := &interfaces.MetricModelQuery{
				RequestMetrics: &interfaces.RequestMetrics{
					Type: interfaces.METRICS_SAMEPERIOD,
					SamePeriodCfg: &interfaces.SamePeriodCfg{
						Method: []string{"invalid_method"},
					},
				},
			}
			err := validateRequestMetrics(ctx, query)
			So(err, ShouldNotBeNil)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_SamePeriodMethod)
		})

		Convey("RequestMetrics.Type is METRICS_SAMEPERIOD but SamePeriodCfg.Offset <= 0", func() {
			query := &interfaces.MetricModelQuery{
				RequestMetrics: &interfaces.RequestMetrics{
					Type: interfaces.METRICS_SAMEPERIOD,
					SamePeriodCfg: &interfaces.SamePeriodCfg{
						Method: []string{interfaces.METRICS_SAMEPERIOD_METHOD_GROWTH_VALUE},
						Offset: 0,
					},
				},
			}
			err := validateRequestMetrics(ctx, query)
			So(err, ShouldNotBeNil)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_SamePeriodOffset)
		})

		Convey("RequestMetrics.Type is METRICS_SAMEPERIOD but SamePeriodCfg.TimeGranularity is invalid", func() {
			query := &interfaces.MetricModelQuery{
				RequestMetrics: &interfaces.RequestMetrics{
					Type: interfaces.METRICS_SAMEPERIOD,
					SamePeriodCfg: &interfaces.SamePeriodCfg{
						Method:          []string{interfaces.METRICS_SAMEPERIOD_METHOD_GROWTH_VALUE},
						Offset:          1,
						TimeGranularity: "invalid_granularity",
					},
				},
			}
			err := validateRequestMetrics(ctx, query)
			So(err, ShouldNotBeNil)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InvalidParameter_SamePeriodTimeGranularity)
		})

		Convey("RequestMetrics.Type is METRICS_SAMEPERIOD and all params valid", func() {
			query := &interfaces.MetricModelQuery{
				RequestMetrics: &interfaces.RequestMetrics{
					Type: interfaces.METRICS_SAMEPERIOD,
					SamePeriodCfg: &interfaces.SamePeriodCfg{
						Method:          []string{interfaces.METRICS_SAMEPERIOD_METHOD_GROWTH_VALUE, interfaces.METRICS_SAMEPERIOD_METHOD_GROWTH_RATE},
						Offset:          1,
						TimeGranularity: interfaces.METRICS_SAMEPERIOD_TIME_GRANULARITY_DAY,
					},
				},
			}
			err := validateRequestMetrics(ctx, query)
			So(err, ShouldBeNil)
		})

		Convey("RequestMetrics.Type is METRICS_PROPORTION", func() {
			query := &interfaces.MetricModelQuery{
				RequestMetrics: &interfaces.RequestMetrics{
					Type: interfaces.METRICS_PROPORTION,
				},
			}
			err := validateRequestMetrics(ctx, query)
			So(err, ShouldBeNil)
		})
	})
}
