// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package trace

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sync"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	jsoniter "github.com/json-iterator/go"
	rest "github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	"uniquery/common"
	"uniquery/common/convert"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	umock "uniquery/interfaces/mock"
	"uniquery/logics"
)

var (
	testCtx = context.WithValue(context.Background(), rest.XLangKey, rest.AmericanEnglish)
)

func MockNewTraceService(appSetting *common.AppSetting, openSearchAccess interfaces.OpenSearchAccess, logGroupAccess interfaces.LogGroupAccess) (ts *traceService) {
	ts = &traceService{
		osClient:   openSearchAccess,
		lgAccess:   logGroupAccess,
		appSetting: appSetting,
	}
	return ts
}

func TestNewTraceService(t *testing.T) {
	Convey("Test NewTraceService", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mockOpenSearchAccess := umock.NewMockOpenSearchAccess(mockCtrl)
		mockLogGroupAccess := umock.NewMockLogGroupAccess(mockCtrl)

		logics.OSAccess = mockOpenSearchAccess
		// logics.DataViewAccess = mockDataViewAccess
		logics.LGAccess = mockLogGroupAccess

		expectedRes := &traceService{
			appSetting: appSetting,
			osClient:   mockOpenSearchAccess,
			lgAccess:   mockLogGroupAccess,
		}

		res := NewTraceService(appSetting)
		So(res, ShouldResemble, expectedRes)
	})
}

func TestGetTraceDetail(t *testing.T) {
	Convey("Test GetTraceDetail", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		appSetting := &common.AppSetting{}
		osAccessMock := umock.NewMockOpenSearchAccess(ctrl)
		dvAccessMock := umock.NewMockLogGroupAccess(ctrl)

		tsMock := MockNewTraceService(appSetting, osAccessMock, dvAccessMock)
		traceDataViewId := "fe5b7f96-443a-11e7-a467-000c29253e90"
		logDataViewId := "ae4b7f84-443a-11e7-a467-000c29253e90"
		traceId := "fe5b7f961154646"

		Convey("get failed, caused by errCh received error", func() {
			expectedErr := errors.New("error")
			patch1 := ApplyFunc(scrollSearchTraceDetail,
				func(ts *traceService, ctx context.Context, wg *sync.WaitGroup, errCh chan<- error, traceDataViewId string, traceDetail *interfaces.TraceDetail) {
					defer wg.Done()
					errCh <- expectedErr
				},
			)
			defer patch1.Reset()

			patch2 := ApplyFunc(scrollSearchRelatedLogCount,
				func(ts *traceService, ctx context.Context, wg *sync.WaitGroup, errCh chan<- error, logDataViewId string, traceId string, spanRelatedLogStats interfaces.SpanRelatedLogStats) {
					wg.Done()
				},
			)
			defer patch2.Reset()

			_, err := tsMock.GetTraceDetail(testCtx, traceDataViewId, logDataViewId, traceId)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("get succeed, and traceStatus is equal to ok", func() {
			patch1 := ApplyFunc(scrollSearchTraceDetail,
				func(ts *traceService, ctx context.Context, wg *sync.WaitGroup, errCh chan<- error, traceDataViewId string, traceDetail *interfaces.TraceDetail) {
					wg.Done()
				},
			)
			defer patch1.Reset()

			patch2 := ApplyFunc(scrollSearchRelatedLogCount,
				func(ts *traceService, ctx context.Context, wg *sync.WaitGroup, errCh chan<- error, logDataViewId string, traceId string, spanRelatedLogStats interfaces.SpanRelatedLogStats) {
					wg.Done()
				},
			)
			defer patch2.Reset()

			patch3 := ApplyFunc(buildTraceTree,
				func(spanMap map[string]*interfaces.BriefSpan) *interfaces.BriefSpan {
					return nil
				},
			)
			defer patch3.Reset()

			patch4 := ApplyFunc(getTraceTreeDepth,
				func(rootSpan *interfaces.BriefSpan) int32 {
					return int32(0)
				},
			)
			defer patch4.Reset()

			res, err := tsMock.GetTraceDetail(testCtx, traceDataViewId, logDataViewId, traceId)
			So(res, ShouldResemble, &interfaces.TraceDetail{
				TraceID:     traceId,
				TraceStatus: "ok",
				StartTime:   math.MaxInt64,
				EndTime:     math.MinInt64,
				Duration:    int64(1),
				Spans:       nil,
				Services:    make([]interfaces.Service, 0),
				SpanStats: map[string]int32{
					"Ok":    0,
					"Error": 0,
					"Unset": 0,
				},
				ServiceStats: make(map[interfaces.Service]int32, 0),
			})

			So(err, ShouldBeNil)
		})

		Convey("get succeed, and traceStatus is equal to error", func() {
			patch1 := ApplyFunc(scrollSearchTraceDetail,
				func(ts *traceService, ctx context.Context, wg *sync.WaitGroup, errCh chan<- error, traceDataViewId string, traceDetail *interfaces.TraceDetail) {
					wg.Done()
					traceDetail.SpanStats["Error"] = 1
				},
			)
			defer patch1.Reset()

			patch2 := ApplyFunc(scrollSearchRelatedLogCount,
				func(ts *traceService, ctx context.Context, wg *sync.WaitGroup, errCh chan<- error, logDataViewId string, traceId string, spanRelatedLogStats interfaces.SpanRelatedLogStats) {
					wg.Done()
				},
			)
			defer patch2.Reset()

			patch3 := ApplyFunc(buildTraceTree,
				func(spanMap map[string]*interfaces.BriefSpan) *interfaces.BriefSpan {
					return nil
				},
			)
			defer patch3.Reset()

			patch4 := ApplyFunc(getTraceTreeDepth,
				func(rootSpan *interfaces.BriefSpan) int32 {
					return int32(0)
				},
			)
			defer patch4.Reset()

			res, err := tsMock.GetTraceDetail(testCtx, traceDataViewId, logDataViewId, traceId)
			So(res, ShouldResemble, &interfaces.TraceDetail{
				TraceID:     traceId,
				TraceStatus: "error",
				StartTime:   math.MaxInt64,
				EndTime:     math.MinInt64,
				Duration:    int64(1),
				Spans:       nil,
				Services:    make([]interfaces.Service, 0),
				SpanStats: map[string]int32{
					"Ok":    0,
					"Error": 1,
					"Unset": 0,
				},
				ServiceStats: make(map[interfaces.Service]int32, 0),
			})

			So(err, ShouldBeNil)
		})
	})
}

func TestScrollSearchTraceDetail(t *testing.T) {
	Convey("Test ScrollSearchTraceDetail", t, func() {

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		appSetting := &common.AppSetting{}
		osAccessMock := umock.NewMockOpenSearchAccess(ctrl)
		dvAccessMock := umock.NewMockLogGroupAccess(ctrl)

		tsMock := MockNewTraceService(appSetting, osAccessMock, dvAccessMock)
		traceDataViewId := "fe5b7f96-443a-11e7-a467-000c29253e90"
		traceDetail := &interfaces.TraceDetail{
			TraceID: "67a3935c108bbd4dea534b5ca1be946d",
		}

		wg := &sync.WaitGroup{}
		errCh := make(chan error)

		Convey("scroll failed, casued by getIndicesAndMustFilters error", func() {
			expectedErr := errors.New("the must_filters field cannot be converted to an interface array")
			patch := ApplyFunc(getIndicesAndMustFilters,
				func(ts *traceService, dataViewId string) ([]string, []interface{}, bool, error) {
					return []string{}, []interface{}{}, false, expectedErr
				},
			)
			defer patch.Reset()

			expectedHttpErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: "en-US",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_InternalError,
					Description:  "Internal Error",
					Solution:     "Please try this operation again, if the error occurs again, submit the work order or contact technical support engineers.",
					ErrorLink:    "None",
					ErrorDetails: "Get indices and must filters failed by trace_data_view_id: " + expectedErr.Error(),
				},
			}

			wg.Add(1)
			go scrollSearchTraceDetail(tsMock, testCtx, wg, errCh, traceDataViewId, traceDetail)
			err := <-errCh

			So(err, ShouldResemble, expectedHttpErr)

			wg.Wait()
		})

		Convey("scroll failed, casued by getIndicesAndMustFilters return nil logIndices", func() {
			patch := ApplyFunc(getIndicesAndMustFilters,
				func(ts *traceService, dataViewId string) ([]string, []interface{}, bool, error) {
					return []string{}, []interface{}{}, false, nil
				},
			)
			defer patch.Reset()

			expectedErr := errors.New("The trace_data_view whose id equals " + traceDataViewId + " was not found!")
			expectedHttpErr := &rest.HTTPError{
				HTTPCode: http.StatusBadRequest,
				Language: "en-US",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_Trace_TraceDataViewNotFound,
					Description:  "The Trace Data View Does Not Exist",
					Solution:     "Please check whether the parameter is correct.",
					ErrorLink:    "None",
					ErrorDetails: expectedErr.Error(),
				},
			}

			wg.Add(1)
			go scrollSearchTraceDetail(tsMock, testCtx, wg, errCh, traceDataViewId, traceDetail)
			err := <-errCh

			So(err, ShouldResemble, expectedHttpErr)

			wg.Wait()
		})

		Convey("scroll failed, casued by searchSubmit return error", func() {
			patch := ApplyFunc(getIndicesAndMustFilters,
				func(ts *traceService, dataViewId string) ([]string, []interface{}, bool, error) {
					return []string{"trace-*"}, []interface{}{}, true, nil
				},
			)
			defer patch.Reset()

			expectedErr := errors.New("error")
			osAccessMock.EXPECT().SearchSubmit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(convert.MapToByte(interfaces.DslResult), http.StatusInternalServerError, expectedErr)

			expectedHttpErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: "en-US",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_InternalError,
					Description:  "Internal Error",
					Solution:     "Please try this operation again, if the error occurs again, submit the work order or contact technical support engineers.",
					ErrorLink:    "None",
					ErrorDetails: expectedErr.Error(),
				},
			}

			wg.Add(1)
			go scrollSearchTraceDetail(tsMock, testCtx, wg, errCh, traceDataViewId, traceDetail)
			err := <-errCh

			So(err, ShouldResemble, expectedHttpErr)

			wg.Wait()
		})

		Convey("scroll failed, casued by total count is equal to zero", func() {
			patch := ApplyFunc(getIndicesAndMustFilters,
				func(ts *traceService, dataViewId string) ([]string, []interface{}, bool, error) {
					return []string{"trace-*"}, []interface{}{}, true, nil
				},
			)
			defer patch.Reset()

			resJson := `{
				"_scroll_id" : "FGluY2x1ZGVfY29udGV4dF91dWlkDnF1ZXJ5VGhlbkZldGNoAxZiNlktNGhBOVNzLTdCdldaaWhla05RAAAAAAABT_cWVXpXb2hVUndRMW1YUlN2N0hxdlJidxZiNlktNGhBOVNzLTdCdldaaWhla05RAAAAAAABT_gWVXpXb2hVUndRMW1YUlN2N0hxdlJidxZiNlktNGhBOVNzLTdCdldaaWhla05RAAAAAAABT_kWVXpXb2hVUndRMW1YUlN2N0hxdlJidw==",
				"took" : 2,
				"timed_out" : false,
				"_shards" : {
				  "total" : 3,
				  "successful" : 3,
				  "skipped" : 0,
				  "failed" : 0
				},
				"hits" : {
				  "total" : {
					"value" : 0,
					"relation" : "eq"
				  },
				  "max_score" : 0.0,
				  "hits" : []
				}
			}`

			osAccessMock.EXPECT().SearchSubmit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return([]byte(resJson), http.StatusOK, nil)

			expectedHttpErr := &rest.HTTPError{
				HTTPCode: http.StatusNotFound,
				Language: "en-US",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_Trace_TraceNotFound,
					Description:  "The Trace Does Not Exist",
					Solution:     "Please check whether the parameter is correct.",
					ErrorLink:    "None",
					ErrorDetails: "The trace whose traceId equals " + traceDetail.TraceID + " was not found!",
				},
			}

			wg.Add(1)
			go scrollSearchTraceDetail(tsMock, testCtx, wg, errCh, traceDataViewId, traceDetail)
			err := <-errCh

			So(err, ShouldResemble, expectedHttpErr)

			wg.Wait()
		})

		Convey("scroll succeed with once query", func() {
			patch1 := ApplyFunc(getIndicesAndMustFilters,
				func(ts *traceService, dataViewId string) ([]string, []interface{}, bool, error) {
					return []string{"trace-*"}, []interface{}{}, true, nil
				},
			)
			defer patch1.Reset()

			patch2 := ApplyFunc(processSpanArray,
				func(resBytes []byte, traceDetail *interfaces.TraceDetail, wg *sync.WaitGroup, ctx context.Context, errCh chan<- error, mutex *sync.RWMutex) {
					wg.Done()
				},
			)
			defer patch2.Reset()

			resJson := `{
				"_scroll_id" : "FGluY2x1ZGVfY29udGV4dF91dWlkDnF1ZXJ5VGhlbkZldGNoAxZiNlktNGhBOVNzLTdCdldaaWhla05RAAAAAAABT_cWVXpXb2hVUndRMW1YUlN2N0hxdlJidxZiNlktNGhBOVNzLTdCdldaaWhla05RAAAAAAABT_gWVXpXb2hVUndRMW1YUlN2N0hxdlJidxZiNlktNGhBOVNzLTdCdldaaWhla05RAAAAAAABT_kWVXpXb2hVUndRMW1YUlN2N0hxdlJidw==",
				"took" : 2,
				"timed_out" : false,
				"_shards" : {
				  "total" : 3,
				  "successful" : 3,
				  "skipped" : 0,
				  "failed" : 0
				},
				"hits" : {
				  "total" : {
					"value" : 1,
					"relation" : "eq"
				  },
				  "max_score" : 0.0,
				  "hits" : [
					{
						"_index" : "json_opentelemetry_trace-2023.03-0",
						"_type" : "_doc",
						"_id" : "BIFg7oYBtvsb2kpvD0EJ",
						"_score" : 0.0,
						"_routing" : "67a3935c108bbd4dea534b5ca1be946d",
						"fields" : {
						  "Resource.service.name.keyword" : [
							"customerA"
						  ],
						  "EndTime" : [
							1679036503480967400
						  ],
						  "SpanContext.SpanID.keyword" : [
							"ff80190fb7961b4f"
						  ],
						  "Status.CodeDesc.keyword" : [
							"Ok"
						  ],
						  "StartTime" : [
							1679036503469671500
						  ],
						  "Duration" : [
							11295900
						  ],
						  "Parent.SpanID.keyword" : [
							"0000000000000000"
						  ],
						  "SpanKindDesc.keyword" : [
							"INTERNAL"
						  ],
						  "Name.keyword" : [
							"customerA-前端入口函数"
						  ]
						}
					}
				  ]
				}
			}`

			osAccessMock.EXPECT().SearchSubmit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return([]byte(resJson), http.StatusOK, nil)

			osAccessMock.EXPECT().DeleteScroll(gomock.Any(), gomock.Any()).
				AnyTimes().Return([]byte{}, http.StatusOK, nil)

			wg.Add(1)
			go scrollSearchTraceDetail(tsMock, testCtx, wg, errCh, traceDataViewId, traceDetail)
			wg.Wait()

			So(traceDetail, ShouldResemble, &interfaces.TraceDetail{
				TraceID: "67a3935c108bbd4dea534b5ca1be946d",
				SpanMap: make(map[string]*interfaces.BriefSpan, 1),
			})
		})

	})
}

func TestScrollSearchRelatedLogCount(t *testing.T) {
	Convey("Test ScrollSearchRelatedLogCount", t, func() {

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		appSetting := &common.AppSetting{}
		osAccessMock := umock.NewMockOpenSearchAccess(ctrl)
		dvAccessMock := umock.NewMockLogGroupAccess(ctrl)

		tsMock := MockNewTraceService(appSetting, osAccessMock, dvAccessMock)
		logDataViewId := "fe5b7f96-443a-11e7-a467-000c29253e90"
		traceId := "67a3935c108bbd4dea534b5ca1be946d"
		spanRelatedLogStats := interfaces.SpanRelatedLogStats{}

		wg := &sync.WaitGroup{}
		errCh := make(chan error)

		Convey("scroll failed, casued by getIndicesAndMustFilters error", func() {
			expectedErr := errors.New("the must_filters field cannot be converted to an interface array")
			patch := ApplyFunc(getIndicesAndMustFilters,
				func(ts *traceService, dataViewId string) ([]string, []interface{}, bool, error) {
					return []string{}, []interface{}{}, false, expectedErr
				},
			)
			defer patch.Reset()

			expectedHttpErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: "en-US",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_InternalError,
					Description:  "Internal Error",
					Solution:     "Please try this operation again, if the error occurs again, submit the work order or contact technical support engineers.",
					ErrorLink:    "None",
					ErrorDetails: "Get indices and must filters failed by log_data_view_id: " + expectedErr.Error(),
				},
			}

			wg.Add(1)
			go scrollSearchRelatedLogCount(tsMock, testCtx, wg, errCh, logDataViewId, traceId, spanRelatedLogStats)
			err := <-errCh

			So(err, ShouldResemble, expectedHttpErr)

			wg.Wait()
		})

		Convey("scroll failed, casued by getIndicesAndMustFilters return nil logIndices", func() {
			patch := ApplyFunc(getIndicesAndMustFilters,
				func(ts *traceService, dataViewId string) ([]string, []interface{}, bool, error) {
					return []string{}, []interface{}{}, false, nil
				},
			)
			defer patch.Reset()

			expectedErr := errors.New("The log_data_view whose id equals " + logDataViewId + " was not found!")
			expectedHttpErr := &rest.HTTPError{
				HTTPCode: http.StatusBadRequest,
				Language: "en-US",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_Trace_LogDataViewNotFound,
					Description:  "The Log Data View Does Not Exist",
					Solution:     "Please check whether the parameter is correct.",
					ErrorLink:    "None",
					ErrorDetails: expectedErr.Error(),
				},
			}

			wg.Add(1)
			go scrollSearchRelatedLogCount(tsMock, testCtx, wg, errCh, logDataViewId, traceId, spanRelatedLogStats)
			err := <-errCh

			So(err, ShouldResemble, expectedHttpErr)

			wg.Wait()
		})

		Convey("scroll failed, casued by searchSubmit return error", func() {
			patch := ApplyFunc(getIndicesAndMustFilters,
				func(ts *traceService, dataViewId string) ([]string, []interface{}, bool, error) {
					return []string{"trace-*"}, []interface{}{}, true, nil
				},
			)
			defer patch.Reset()

			expectedErr := errors.New("error")
			osAccessMock.EXPECT().SearchSubmit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(convert.MapToByte(interfaces.DslResult), http.StatusInternalServerError, expectedErr)

			expectedHttpErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: "en-US",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_InternalError,
					Description:  "Internal Error",
					Solution:     "Please try this operation again, if the error occurs again, submit the work order or contact technical support engineers.",
					ErrorLink:    "None",
					ErrorDetails: expectedErr.Error(),
				},
			}

			wg.Add(1)
			go scrollSearchRelatedLogCount(tsMock, testCtx, wg, errCh, logDataViewId, traceId, spanRelatedLogStats)
			err := <-errCh

			So(err, ShouldResemble, expectedHttpErr)

			wg.Wait()
		})

		Convey("scroll failed, casued by decode error", func() {
			patch := ApplyFunc(getIndicesAndMustFilters,
				func(ts *traceService, dataViewId string) ([]string, []interface{}, bool, error) {
					return []string{"trace-*"}, []interface{}{}, true, nil
				},
			)
			defer patch.Reset()

			osAccessMock.EXPECT().SearchSubmit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(convert.MapToByte(interfaces.DslResult), http.StatusOK, nil)

			expectedErr := errors.New("error")
			patches := ApplyMethodReturn(&jsoniter.Decoder{}, "Decode", expectedErr)
			defer patches.Reset()

			expectedHttpErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: "en-US",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_InternalError,
					Description:  "Internal Error",
					Solution:     "Please try this operation again, if the error occurs again, submit the work order or contact technical support engineers.",
					ErrorLink:    "None",
					ErrorDetails: expectedErr.Error(),
				},
			}

			wg.Add(1)
			go scrollSearchRelatedLogCount(tsMock, testCtx, wg, errCh, logDataViewId, traceId, spanRelatedLogStats)
			err := <-errCh

			So(err, ShouldResemble, expectedHttpErr)

			wg.Wait()
		})

		Convey("scroll succeed with once query", func() {
			patch := ApplyFunc(getIndicesAndMustFilters,
				func(ts *traceService, dataViewId string) ([]string, []interface{}, bool, error) {
					return []string{"trace-*"}, []interface{}{}, true, nil
				},
			)
			defer patch.Reset()

			resJson := `{
				"took" : 177,
				"timed_out" : false,
				"_shards" : {
				  	"total" : 9,
				  	"successful" : 9,
				  	"skipped" : 0,
				  	"failed" : 0
				},
				"hits" : {
				  	"total" : {
						"value" : 10000,
						"relation" : "gte"
				  	},
				  	"max_score" : null,
				  	"hits" : []
				},
				"aggregations" : {
				  	"group_by_SpanID" : {
						"doc_count_error_upper_bound" : 0,
						"sum_other_doc_count" : 0,
						"buckets" : [
							{
								"key" : "%{[Resource][Tingyun-spanId]}_0",
								"doc_count" : 166308
							},
							{
								"key" : "dfc58f6bbfa06f8b_0",
								"doc_count" : 1000
							},
							{
								"key" : "018c71485492b98f_0",
								"doc_count" : 42
							}
						]
					}
				}
			}`

			osAccessMock.EXPECT().SearchSubmit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return([]byte(resJson), http.StatusOK, nil)

			wg.Add(1)
			go scrollSearchRelatedLogCount(tsMock, testCtx, wg, errCh, logDataViewId, traceId, spanRelatedLogStats)
			wg.Wait()

			So(spanRelatedLogStats, ShouldResemble, interfaces.SpanRelatedLogStats{
				"%{[Resource][Tingyun-spanId]}_0": int64(166308),
				"dfc58f6bbfa06f8b_0":              int64(1000),
				"018c71485492b98f_0":              int64(42),
			})
		})

	})
}

func TestGetIndicesAndMustFilters(t *testing.T) {
	Convey("Test GetIndicesAndMustFilters", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		appSetting := &common.AppSetting{}
		osAccessMock := umock.NewMockOpenSearchAccess(ctrl)
		dvAccessMock := umock.NewMockLogGroupAccess(ctrl)

		tsMock := MockNewTraceService(appSetting, osAccessMock, dvAccessMock)
		dataViewId := "fe5b7f96-443a-11e7-a467-000c29253e90"

		Convey("get succeed", func() {
			expectedDataView := interfaces.LogGroup{
				IndexPattern: []string{"trace-*"},
				MustFilter:   []interface{}{},
			}
			dvAccessMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				expectedDataView, true, nil)

			indices, mustFilter, isExist, err := getIndicesAndMustFilters(tsMock, dataViewId)
			So(indices, ShouldResemble, expectedDataView.IndexPattern)
			So(mustFilter, ShouldResemble, expectedDataView.MustFilter)
			So(isExist, ShouldBeTrue)
			So(err, ShouldBeNil)
		})

		Convey("get failed, caused by data manager with error response", func() {
			expectedErr := errors.New("get queryfilters failed")
			dvAccessMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				interfaces.LogGroup{}, false, expectedErr)

			indices, mustFilter, isExist, err := getIndicesAndMustFilters(tsMock, dataViewId)
			So(indices, ShouldResemble, []string{})
			So(mustFilter, ShouldResemble, []interface{}{})
			So(isExist, ShouldBeFalse)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("get failed, caused by error must_filter type", func() {
			expectedErr := errors.New("the must_filters field cannot be converted to an interface array")
			expectedDataView := interfaces.LogGroup{
				IndexPattern: []string{"trace-*"},
				MustFilter:   []string{"must_filter"},
			}

			dvAccessMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				expectedDataView, true, nil)

			indices, mustFilter, isExist, err := getIndicesAndMustFilters(tsMock, dataViewId)
			So(indices, ShouldResemble, []string{})
			So(mustFilter, ShouldResemble, []interface{}{})
			So(isExist, ShouldBeFalse)
			So(err, ShouldResemble, expectedErr)
		})
	})
}

func TestProcessSpanArray(t *testing.T) {
	Convey("Test ProcessSpanArray", t, func() {

		traceDetail := &interfaces.TraceDetail{}
		wg := &sync.WaitGroup{}
		errCh := make(chan error)
		mutex := &sync.RWMutex{}

		correctJson := `{
			"_scroll_id" : "FGluY2x1ZGVfY29udGV4dF91dWlkDnF1ZXJ5VGhlbkZldGNoAxZiNlktNGhBOVNzLTdCdldaaWhla05RAAAAAAABI_gWVXpXb2hVUndRMW1YUlN2N0hxdlJidxZiNlktNGhBOVNzLTdCdldaaWhla05RAAAAAAABI_kWVXpXb2hVUndRMW1YUlN2N0hxdlJidxZiNlktNGhBOVNzLTdCdldaaWhla05RAAAAAAABI_oWVXpXb2hVUndRMW1YUlN2N0hxdlJidw==",
			"took" : 66,
			"timed_out" : false,
			"_shards" : {
			  "total" : 3,
			  "successful" : 3,
			  "skipped" : 0,
			  "failed" : 0
			},
			"hits" : {
			  "total" : {
				"value" : 1,
				"relation" : "eq"
			  },
			  "max_score" : 0.0,
			  "hits" : [
				{
				  "_index" : "json_opentelemetry_trace-2023.03-0",
				  "_type" : "_doc",
				  "_id" : "BIFg7oYBtvsb2kpvD0EJ",
				  "_score" : 0.0,
				  "_routing" : "67a3935c108bbd4dea534b5ca1be946d",
				  "fields" : {
					"Resource.service.name.keyword" : [
					  "customerA"
					],
					"EndTime" : [
					  1679036503480967400
					],
					"SpanContext.SpanID.keyword" : [
					  "ff80190fb7961b4f"
					],
					"Status.CodeDesc.keyword" : [
					  "Ok"
					],
					"StartTime" : [
					  1679036503469671500
					],
					"Duration" : [
					  11295900
					],
					"Parent.SpanID.keyword" : [
					  "0000000000000000"
					],
					"SpanKindDesc.keyword" : [
					  "INTERNAL"
					],
					"Name.keyword" : [
					  "customerA-前端入口函数"
					]
				  }
				}
			  ]
			}
		  }`

		errorJsonWithErrorSpanId := `
		{
			"_scroll_id" : "FGluY2x1ZGVfY29udGV4dF91dWlkDnF1ZXJ5VGhlbkZldGNoAxZiNlktNGhBOVNzLTdCdldaaWhla05RAAAAAAABI_gWVXpXb2hVUndRMW1YUlN2N0hxdlJidxZiNlktNGhBOVNzLTdCdldaaWhla05RAAAAAAABI_kWVXpXb2hVUndRMW1YUlN2N0hxdlJidxZiNlktNGhBOVNzLTdCdldaaWhla05RAAAAAAABI_oWVXpXb2hVUndRMW1YUlN2N0hxdlJidw==",
			"took" : 66,
			"timed_out" : false,
			"_shards" : {
				"total" : 3,
				"successful" : 3,
				"skipped" : 0,
				"failed" : 0
			},
			"hits" : {
				"total" : {
					"value" : 1,
					"relation" : "eq"
				},
				"max_score" : 0.0,
				"hits" : [
					{
						"_index" : "json_opentelemetry_trace-2023.03-0",
						"_type" : "_doc",
						"_id" : "BIFg7oYBtvsb2kpvD0EJ",
						"_score" : 0.0,
						"_routing" : "67a3935c108bbd4dea534b5ca1be946d",
						"fields" : {
							"Resource.service.name.keyword" : [
								"customerA"
							],
							"EndTime" : [
								1679036503480967400
							],
							"SpanContext.SpanID.keyword" : [
								1
							],
							"Status.CodeDesc.keyword" : [
								"Ok"
							],
							"StartTime" : [
								1679036503469671500
							],
							"Duration" : [
								11295900
							],
							"Parent.SpanID.keyword" : [
								"0000000000000000"
							],
							"SpanKindDesc.keyword" : [
								"INTERNAL"
							],
							"Name.keyword" : [
								"customerA-前端入口函数"
							]
						}
					}
				]
			}
		}
		`

		errorJsonWithErrorParentSpanId := `
		{
			"_scroll_id" : "FGluY2x1ZGVfY29udGV4dF91dWlkDnF1ZXJ5VGhlbkZldGNoAxZiNlktNGhBOVNzLTdCdldaaWhla05RAAAAAAABI_gWVXpXb2hVUndRMW1YUlN2N0hxdlJidxZiNlktNGhBOVNzLTdCdldaaWhla05RAAAAAAABI_kWVXpXb2hVUndRMW1YUlN2N0hxdlJidxZiNlktNGhBOVNzLTdCdldaaWhla05RAAAAAAABI_oWVXpXb2hVUndRMW1YUlN2N0hxdlJidw==",
			"took" : 66,
			"timed_out" : false,
			"_shards" : {
				"total" : 3,
				"successful" : 3,
				"skipped" : 0,
				"failed" : 0
			},
			"hits" : {
				"total" : {
					"value" : 1,
					"relation" : "eq"
				},
				"max_score" : 0.0,
				"hits" : [
					{
						"_index" : "json_opentelemetry_trace-2023.03-0",
						"_type" : "_doc",
						"_id" : "BIFg7oYBtvsb2kpvD0EJ",
						"_score" : 0.0,
						"_routing" : "67a3935c108bbd4dea534b5ca1be946d",
						"fields" : {
							"Resource.service.name.keyword" : [
								"customerA"
							],
							"EndTime" : [
								1679036503480967400
							],
							"SpanContext.SpanID.keyword" : [
								"ff80190fb7961b4f"
							],
							"Status.CodeDesc.keyword" : [
								"Ok"
							],
							"StartTime" : [
								1679036503469671500
							],
							"Duration" : [
								11295900
							],
							"Parent.SpanID.keyword" : [
								1
							],
							"SpanKindDesc.keyword" : [
								"INTERNAL"
							],
							"Name.keyword" : [
								"customerA-前端入口函数"
							]
						}
					}
				]
			}
		}
		`

		errorJsonWithErrorStartTime := `
		{
			"_scroll_id" : "FGluY2x1ZGVfY29udGV4dF91dWlkDnF1ZXJ5VGhlbkZldGNoAxZiNlktNGhBOVNzLTdCdldaaWhla05RAAAAAAABI_gWVXpXb2hVUndRMW1YUlN2N0hxdlJidxZiNlktNGhBOVNzLTdCdldaaWhla05RAAAAAAABI_kWVXpXb2hVUndRMW1YUlN2N0hxdlJidxZiNlktNGhBOVNzLTdCdldaaWhla05RAAAAAAABI_oWVXpXb2hVUndRMW1YUlN2N0hxdlJidw==",
			"took" : 66,
			"timed_out" : false,
			"_shards" : {
				"total" : 3,
				"successful" : 3,
				"skipped" : 0,
				"failed" : 0
			},
			"hits" : {
				"total" : {
					"value" : 1,
					"relation" : "eq"
				},
				"max_score" : 0.0,
				"hits" : [
					{
						"_index" : "json_opentelemetry_trace-2023.03-0",
						"_type" : "_doc",
						"_id" : "BIFg7oYBtvsb2kpvD0EJ",
						"_score" : 0.0,
						"_routing" : "67a3935c108bbd4dea534b5ca1be946d",
						"fields" : {
							"Resource.service.name.keyword" : [
								"customerA"
							],
							"EndTime" : [
								1679036503480967400
							],
							"SpanContext.SpanID.keyword" : [
								"ff80190fb7961b4f"
							],
							"Status.CodeDesc.keyword" : [
								"Ok"
							],
							"StartTime" : [
								"1679036503469671500"
							],
							"Duration" : [
								11295900
							],
							"Parent.SpanID.keyword" : [
								"0000000000000000"
							],
							"SpanKindDesc.keyword" : [
								"INTERNAL"
							],
							"Name.keyword" : [
								"customerA-前端入口函数"
							]
						}
					}
				]
			}
		}
		`

		errorJsonWithErrorEndTime := `
		{
			"_scroll_id" : "FGluY2x1ZGVfY29udGV4dF91dWlkDnF1ZXJ5VGhlbkZldGNoAxZiNlktNGhBOVNzLTdCdldaaWhla05RAAAAAAABI_gWVXpXb2hVUndRMW1YUlN2N0hxdlJidxZiNlktNGhBOVNzLTdCdldaaWhla05RAAAAAAABI_kWVXpXb2hVUndRMW1YUlN2N0hxdlJidxZiNlktNGhBOVNzLTdCdldaaWhla05RAAAAAAABI_oWVXpXb2hVUndRMW1YUlN2N0hxdlJidw==",
			"took" : 66,
			"timed_out" : false,
			"_shards" : {
				"total" : 3,
				"successful" : 3,
				"skipped" : 0,
				"failed" : 0
			},
			"hits" : {
				"total" : {
					"value" : 1,
					"relation" : "eq"
				},
				"max_score" : 0.0,
				"hits" : [
					{
						"_index" : "json_opentelemetry_trace-2023.03-0",
						"_type" : "_doc",
						"_id" : "BIFg7oYBtvsb2kpvD0EJ",
						"_score" : 0.0,
						"_routing" : "67a3935c108bbd4dea534b5ca1be946d",
						"fields" : {
							"Resource.service.name.keyword" : [
								"customerA"
							],
							"EndTime" : [
								"1679036503480967400"
							],
							"SpanContext.SpanID.keyword" : [
								"ff80190fb7961b4f"
							],
							"Status.CodeDesc.keyword" : [
								"Ok"
							],
							"StartTime" : [
								1679036503469671500
							],
							"Duration" : [
								11295900
							],
							"Parent.SpanID.keyword" : [
								"0000000000000000"
							],
							"SpanKindDesc.keyword" : [
								"INTERNAL"
							],
							"Name.keyword" : [
								"customerA-前端入口函数"
							]
						}
					}
				]
			}
		}
		`

		Convey("process failed, casued by unmarshal error", func() {
			expectedErr := errors.New("error")
			patches := ApplyMethodReturn(&jsoniter.Decoder{}, "Decode", expectedErr)
			defer patches.Reset()

			expectedHttpErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: "en-US",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_InternalError,
					Description:  "Internal Error",
					Solution:     "Please try this operation again, if the error occurs again, submit the work order or contact technical support engineers.",
					ErrorLink:    "None",
					ErrorDetails: expectedErr.Error(),
				},
			}

			wg.Add(1)
			go processSpanArray([]byte(correctJson), traceDetail, wg, testCtx, errCh, mutex)

			err := <-errCh
			So(err, ShouldResemble, expectedHttpErr)

			wg.Wait()
		})

		Convey("process succeed, but span with error spanId type", func() {
			wg.Add(1)
			processSpanArray([]byte(errorJsonWithErrorSpanId), traceDetail, wg, testCtx, errCh, mutex)

			So(traceDetail, ShouldResemble, &interfaces.TraceDetail{})
		})

		Convey("process succeed, but span with error parent spanId type", func() {
			wg.Add(1)
			processSpanArray([]byte(errorJsonWithErrorParentSpanId), traceDetail, wg, testCtx, errCh, mutex)

			So(traceDetail, ShouldResemble, &interfaces.TraceDetail{})
		})

		Convey("process succeed, but span with error startTime type", func() {
			spanId := "ff80190fb7961b4f"
			expectedErr := fmt.Errorf("an error occurred while converting the field type of the span whose spanId is %s, err: StartTime is not a json.Number, so it can not be converted to an int64", spanId)
			expectedHttpErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: "en-US",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_InternalError,
					Description:  "Internal Error",
					Solution:     "Please try this operation again, if the error occurs again, submit the work order or contact technical support engineers.",
					ErrorLink:    "None",
					ErrorDetails: expectedErr.Error(),
				},
			}

			wg.Add(1)
			go processSpanArray([]byte(errorJsonWithErrorStartTime), traceDetail, wg, testCtx, errCh, mutex)

			err := <-errCh
			So(err, ShouldResemble, expectedHttpErr)

			wg.Wait()
		})

		Convey("process succeed, but span with error endTime type", func() {
			spanId := "ff80190fb7961b4f"
			expectedErr := fmt.Errorf("an error occurred while converting the field type of the span whose spanId is %s, err: EndTime is not a json.Number, so it can not be converted to an int64", spanId)
			expectedHttpErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: "en-US",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_InternalError,
					Description:  "Internal Error",
					Solution:     "Please try this operation again, if the error occurs again, submit the work order or contact technical support engineers.",
					ErrorLink:    "None",
					ErrorDetails: expectedErr.Error(),
				},
			}

			wg.Add(1)
			go processSpanArray([]byte(errorJsonWithErrorEndTime), traceDetail, wg, testCtx, errCh, mutex)

			err := <-errCh
			So(err, ShouldResemble, expectedHttpErr)

			wg.Wait()
		})

		Convey("process succeed", func() {
			traceDetail := &interfaces.TraceDetail{
				// Spans:    &interfaces.BriefSpan{},
				// Services: make([]interfaces.Service, 0),
				SpanMap: make(map[string]*interfaces.BriefSpan, 0),
				SpanStats: map[string]int32{
					"Ok":    0,
					"Error": 0,
					"Unset": 0,
				},
				ServiceStats: make(map[interfaces.Service]int32, 0),
			}
			wg.Add(1)
			go processSpanArray([]byte(correctJson), traceDetail, wg, testCtx, errCh, mutex)
			wg.Wait()

			So(traceDetail, ShouldResemble, &interfaces.TraceDetail{
				EndTime: 1679036503480967400,
				SpanStats: map[string]int32{
					"Error": 0,
					"Ok":    1,
					"Unset": 0,
				},
				SpanMap: map[string]*interfaces.BriefSpan{
					"ff80190fb7961b4f": {
						Key:  "ff80190fb7961b4f",
						Name: "customerA-前端入口函数",
						SpanContext: interfaces.BriefSpanContext{
							SpanID: "ff80190fb7961b4f",
						},
						Parent: interfaces.BriefSpanContext{
							SpanID: "0000000000000000",
						},
						SpanKindDesc: "INTERNAL",
						StartTime:    1679036503469671500,
						EndTime:      1679036503480967400,
						Duration:     11295900,
						Status: interfaces.BriefStatus{
							CodeDesc: "Ok",
						},
						Resource: interfaces.BriefResource{
							Service: interfaces.Service{
								Name: "customerA",
							},
						},
						Children: make([]*interfaces.BriefSpan, 0),
					},
				},
				ServiceStats: map[interfaces.Service]int32{
					{Name: "customerA"}: 1,
				},
			})
		})

	})
}

// func TestProcessLogBuckets(t *testing.T) {
// 	Convey("Test ProcessLogBuckets", t, func() {
// 		Convey("process succeed", func() {
// 			resJson := `{
// 				"buckets": [
// 					{
// 						"key": "1",
// 						"doc_count": 1
// 					},
// 					{
// 						"key": "2",
// 						"doc_count": 4
// 					},
// 					{
// 						"key": "3",
// 						"doc_count": 5
// 					}
// 				]
// 			}`

// 			resBytes, _ := jsoniter.Marshal(resJson)
// 			ctx := context.WithValue(context.Background(), rest.XLangKey, "zh-CN")
// 			errCh := make(chan error)
// 			spanRelatedLogStats := interfaces.SpanRelatedLogStats{}

// 			processLogBuckets(resBytes, testCtx,errCh, spanRelatedLogStats)
// 			So(spanRelatedLogStats["1"], ShouldEqual, 1)
// 			So(spanRelatedLogStats["2"], ShouldEqual, 4)
// 			So(spanRelatedLogStats["3"], ShouldEqual, 5)
// 			So(len(spanRelatedLogStats), ShouldEqual, 3)
// 		})
// 	})
// }

func TestBuildTraceTree(t *testing.T) {
	Convey("Test BuildTraceTree", t, func() {
		Convey("build succeed", func() {
			spanMap := map[string]*interfaces.BriefSpan{
				"1": {
					StartTime: int64(1),
					Parent: interfaces.BriefSpanContext{
						SpanID: "0",
					},
					Children: []*interfaces.BriefSpan{},
				},
				"2": {
					StartTime: int64(2),
					Parent: interfaces.BriefSpanContext{
						SpanID: "1",
					},
					Children: []*interfaces.BriefSpan{},
				},
				"3": {
					StartTime: int64(4),
					Parent: interfaces.BriefSpanContext{
						SpanID: "1",
					},
					Children: []*interfaces.BriefSpan{},
				},
				"4": {
					StartTime: int64(3),
					Parent: interfaces.BriefSpanContext{
						SpanID: "1",
					},
					Children: []*interfaces.BriefSpan{},
				},
				"5": {
					StartTime: int64(5),
					Parent: interfaces.BriefSpanContext{
						SpanID: "2",
					},
					Children: []*interfaces.BriefSpan{},
				},
			}

			expectedRes := &interfaces.BriefSpan{
				StartTime: int64(1),
				Parent: interfaces.BriefSpanContext{
					SpanID: "0",
				},
				Children: []*interfaces.BriefSpan{
					{
						StartTime: int64(2),
						Parent: interfaces.BriefSpanContext{
							SpanID: "1",
						},
						Children: []*interfaces.BriefSpan{
							{
								StartTime: int64(5),
								Parent: interfaces.BriefSpanContext{
									SpanID: "2",
								},
								Children: []*interfaces.BriefSpan{},
							},
						},
					},
					{
						StartTime: int64(3),
						Parent: interfaces.BriefSpanContext{
							SpanID: "1",
						},
						Children: []*interfaces.BriefSpan{},
					},
					{
						StartTime: int64(4),
						Parent: interfaces.BriefSpanContext{
							SpanID: "1",
						},
						Children: []*interfaces.BriefSpan{},
					},
				},
			}

			res := buildTraceTree(spanMap)
			So(res, ShouldResemble, expectedRes)

		})
	})
}

func TestGetTraceTreeDepth(t *testing.T) {
	Convey("Test GetTraceTreeDepth", t, func() {
		Convey("rootSpan is null, return 0", func() {
			res := getTraceTreeDepth(nil)
			So(res, ShouldEqual, 0)
		})
		Convey("rootSpan is not null, return correct value", func() {
			rootSpan := interfaces.BriefSpan{
				SpanContext: interfaces.BriefSpanContext{
					SpanID: "1",
				},
				Children: []*interfaces.BriefSpan{
					{
						SpanContext: interfaces.BriefSpanContext{
							SpanID: "2",
						},
						Children: []*interfaces.BriefSpan{
							{
								SpanContext: interfaces.BriefSpanContext{
									SpanID: "4",
								},
								Children: []*interfaces.BriefSpan{},
							},
						},
					},
					{
						SpanContext: interfaces.BriefSpanContext{
							SpanID: "3",
						},
						Children: []*interfaces.BriefSpan{},
					},
				},
			}

			res := getTraceTreeDepth(&rootSpan)
			So(res, ShouldEqual, 3)
		})

	})
}

func TestMinInt64(t *testing.T) {
	Convey("Test MinInt64", t, func() {
		Convey("a < b, return a", func() {
			var a, b int64 = 5, 15
			res := minInt64(a, b)
			So(res, ShouldEqual, a)
		})
		Convey("a > b, return b", func() {
			var a, b int64 = 18, 8
			res := minInt64(a, b)
			So(res, ShouldEqual, b)
		})
	})
}

func TestMaxInt64(t *testing.T) {
	Convey("Test MaxInt64", t, func() {
		Convey("a > b, return a", func() {
			var a, b int64 = 10, 5
			res := maxInt64(a, b)
			So(res, ShouldEqual, a)
		})
		Convey("a < b, return b", func() {
			var a, b int64 = 10, 20
			res := maxInt64(a, b)
			So(res, ShouldEqual, b)
		})
	})
}

// func TestGetSpanList(t *testing.T) {
// 	Convey("Test GetSpanList", t, func() {
// 		ctrl := gomock.NewController(t)
// 		defer ctrl.Finish()

// 		osAccessMock := umock.NewMockOpenSearchAccess(ctrl)
// 		dvAccessMock := umock.NewMockLogGroupAccess(ctrl)

// 		tsMock := MockNewTraceService(osAccessMock, dvAccessMock)

// 		expectedDataView := interfaces.DataView{
// 			IndexPattern: []string{"trace-*"},
// 			MustFilter:   []interface{}{},
// 		}

// 		listSpanQuery := interfaces.SpanListQuery{
// 			DataViewId: "69bbd3de-ac32-11ed-89e8-ca9b4576213469bbd3de-ac32-11ed-89e8-ca9b45762134",
// 			SpanStatusMap: map[string]int{
// 				"ok":    1,
// 				"error": 1,
// 			},
// 			Limit:     interfaces.MAX_LIMIT,
// 			Offset:    interfaces.MIN_OFFSET,
// 			StartTime: 1614136611000,
// 			EndTime:   1677208611000,
// 		}

// 		drivenOkRes := []interfaces.SpanList{
// 			{
// 				Name:       "Span1-1",
// 				TraceID:    "ab4fe1de2d9c951411bbdbc3533d747f",
// 				SpanStatus: "ok",
// 				StartTime:  "2022-10-17T18:51:58.114304Z",
// 				Duration:   131,
// 			},
// 			{
// 				Name:       "Span2-1",
// 				TraceID:    "ab4fe1de2d9c951411bbdbc3533d747e",
// 				SpanStatus: "ok",
// 				StartTime:  "2022-09-02T18:52:58.114304Z",
// 				Duration:   131,
// 			},
// 		}
// 		Convey("Get trace list failed, caused by data manager with error response", func() {
// 			err := errors.New("get queryfilters failed")
// 			dvAccessMock.EXPECT().GetDataViewQueryFilters(gomock.Any()).AnyTimes().Return(
// 				interfaces.DataView{}, http.StatusInternalServerError, err)

// 			expectedErr := &rest.HTTPError{
// 				HTTPCode: http.StatusInternalServerError,
// 				BaseError: rest.BaseError{
// 					ErrorCode:    uerrors.Uniquery_InternalError,
// 					Description:  "服务器内部错误",
// 					Solution:     "",
// 					ErrorLink:    "",
// 					ErrorDetails: err.Error(),
// 				},
// 			}
// 			patch := ApplyFunc(rest.NewHTTPError,
// 				func(httpCode int, errorCode string) *rest.HTTPError {
// 					return expectedErr
// 				},
// 			)
// 			defer patch.Reset()

// 			res, total, httpErr := tsMock.GetSpanList(listSpanQuery)
// 			So(res, ShouldResemble, []interfaces.SpanList{})
// 			So(total, ShouldEqual, 0)
// 			So(httpErr, ShouldEqual, expectedErr)
// 		})

// 		Convey("Get trace list failed, caused by the dataview was not found", func() {
// 			dvAccessMock.EXPECT().GetDataViewQueryFilters(gomock.Any()).AnyTimes().Return(
// 				interfaces.DataView{}, http.StatusOK, nil)

// 			expectedErr := &rest.HTTPError{
// 				HTTPCode: http.StatusNotFound,
// 				BaseError: rest.BaseError{
// 					ErrorCode:    uerrors.Uniquery_NoSuchARDataView,
// 					Description:  "指定的日志分组不存在",
// 					Solution:     "",
// 					ErrorLink:    "",
// 					ErrorDetails: "The dataView whose dataViewId equals " + listSpanQuery.DataViewId + " was not found!",
// 				},
// 			}
// 			patch := ApplyFunc(rest.NewHTTPError,
// 				func(httpCode int, errorCode string) *rest.HTTPError {
// 					return expectedErr
// 				},
// 			)
// 			defer patch.Reset()

// 			res, total, httpErr := tsMock.GetSpanList(listSpanQuery)
// 			So(res, ShouldResemble, []interfaces.SpanList{})
// 			So(total, ShouldEqual, 0)
// 			So(httpErr, ShouldEqual, expectedErr)
// 		})

// 		Convey("Get trace list failed, caused by opensearch with error response", func() {
// 			dvAccessMock.EXPECT().GetDataViewQueryFilters(gomock.Any()).AnyTimes().Return(
// 				expectedDataView, http.StatusOK, nil)

// 			err := errors.New("opensearch error")
// 			patch1 := ApplyFunc(searchSpanList,
// 				func(ts *traceService, query map[string]interface{}, indices []string) ([]interfaces.SpanList, int, error) {
// 					return []interfaces.SpanList{}, 0, err
// 				})
// 			defer patch1.Reset()

// 			expectedErr := &rest.HTTPError{
// 				HTTPCode: http.StatusInternalServerError,
// 				BaseError: rest.BaseError{
// 					ErrorCode:    uerrors.Uniquery_InternalError,
// 					Description:  "服务器内部错误",
// 					Solution:     "",
// 					ErrorLink:    "",
// 					ErrorDetails: err.Error(),
// 				},
// 			}
// 			patch2 := ApplyFunc(rest.NewHTTPError,
// 				func(httpCode int, errorCode string) *rest.HTTPError {
// 					return expectedErr
// 				},
// 			)
// 			defer patch2.Reset()

// 			res, total, httpErr := tsMock.GetSpanList(listSpanQuery)
// 			So(res, ShouldResemble, []interfaces.SpanList{})
// 			So(total, ShouldEqual, 0)
// 			So(httpErr, ShouldEqual, expectedErr)
// 		})

// 		Convey("Get trace detail successed", func() {
// 			dvAccessMock.EXPECT().GetDataViewQueryFilters(gomock.Any()).AnyTimes().Return(
// 				expectedDataView, http.StatusOK, nil)

// 			patch := ApplyFunc(searchSpanList,
// 				func(ts *traceService, query map[string]interface{}, indices []string) ([]interfaces.SpanList, int, error) {
// 					return drivenOkRes, 2, nil
// 				})
// 			defer patch.Reset()

// 			res, total, httpErr := tsMock.GetSpanList(listSpanQuery)
// 			So(res, ShouldResemble, drivenOkRes)
// 			So(total, ShouldEqual, 2)
// 			So(httpErr, ShouldBeNil)
// 		})
// 	})
// }

// to delete
// func TestGetTraceDetail(t *testing.T) {
// 	Convey("Test GetTraceDetail", t, func() {
// 		ctrl := gomock.NewController(t)
// 		defer ctrl.Finish()

// 		appSetting := &common.AppSetting{
// 			TraceSetting: common.TraceSetting{
// 				MaxSearchSpanSize: 1000,
// 			},
// 		}
// 		osAccessMock := umock.NewMockOpenSearchAccess(ctrl)
// 		dvAccessMock := umock.NewMockLogGroupAccess(ctrl)

// 		tsMock := MockNewTraceService(appSetting, osAccessMock, dvAccessMock)

// 		expectedDataView := interfaces.DataView{
// 			IndexPattern: []string{"trace-*"},
// 			MustFilter:   []interface{}{},
// 		}
// 		dataViewId := "69bbd3de-ac32-11ed-89e8-ca9b4576213469bbd3de-ac32-11ed-89e8-ca9b45762134"
// 		traceId := "ab4fe1de2d9c951411bbdbc3533d747e"

// 		spanJson0 := `{
// 			"Name" : "Span2-1",
// 			"SpanContext" : {
// 			  "TraceID" : "ab4fe1de2d9c951411bbdbc3533d747e",
// 			  "SpanID" : "44c39ef193f3e08b",
// 			  "TraceFlags" : "01",
// 			  "TraceState" : {
// 				"rojo" : "00f067aa0ba902b7",
// 				"congo" : "t61rcWkgMzE"
// 			  }
// 			},
// 			"Parent" : {
// 			  "TraceID" : "00000000000000000000000000000000",
// 			  "SpanID" : "0000000000000000",
// 			  "TraceFlags" : "00",
// 			  "TraceState" : null
// 			},
// 			"SpanKind" : 1,
// 			"StartTime" : 1662144778114304000,
// 			"EndTime" : 1662144838114435000,
// 			"@timestamp" : "2022-09-02T18:52:58.114Z",
// 			"Duration" : 131000,
// 			"Attributes" : {
// 			  "num" : "45"
// 			},
// 			"Events" : [
// 			  {
// 				"Name" : "Nice operation!",
// 				"Attributes" : {
// 				  "bogons" : 100
// 				},
// 				"Time" : 1662058438114304000
// 			  }
// 			],
// 			"Links" : [
// 			  {
// 				"TraceID" : "ab4fe1de2d9c951411bbdbc3533d747f",
// 				"SpanID" : "44c39ef193f3e08G",
// 				"TraceState" : {
// 				  "rojo" : "00f067aa0ba902b7",
// 				  "congo" : "t61rcWkgMzE"
// 				},
// 				"Attributes" : {
// 				  "num1" : 100
// 				}
// 			  }
// 			],
// 			"Status" : {
// 			  "Code": 1,
// 			  "CodeDesc" : "Error",
// 			  "Description" : ""
// 			},
// 			"Resource" : {
// 			  "job_id" : "13bf135ce8b1481e9329a5e3b62171ae",
// 			  "service" : {
// 				"name" : "服务A",
// 				"version" : "版本1.0.1"
// 			  },
// 			  "telemetry" : {
// 				"sdk" : {
// 				  "language" : "go",
// 				  "name" : "opentelemetry",
// 				  "version" : "1.9.0"
// 				}
// 			  }
// 			},
// 			"catagory" : "trace"
// 		  }`

// 		spanJson1 := `{
// 			"Name" : "Span2-1",
// 			"SpanContext" : {
// 			  "TraceID" : "ab4fe1de2d9c951411bbdbc3533d747e",
// 			  "SpanID" : "44c39ef193f3e08b",
// 			  "TraceFlags" : "01",
// 			  "TraceState" : {
// 				"rojo" : "00f067aa0ba902b7",
// 				"congo" : "t61rcWkgMzE"
// 			  }
// 			},
// 			"Parent" : {
// 			  "TraceID" : "00000000000000000000000000000000",
// 			  "SpanID" : "0000000000000000",
// 			  "TraceFlags" : "00",
// 			  "TraceState" : null
// 			},
// 			"SpanKind" : 1,
// 			"StartTime" : 1662144778114304000,
// 			"EndTime" : 1662144838114435000,
// 			"@timestamp" : "2022-09-02T18:52:58.114Z",
// 			"Duration" : 131000,
// 			"Attributes" : {
// 			  "num" : "45"
// 			},
// 			"Events" : [
// 			  {
// 				"Name" : "Nice operation!",
// 				"Attributes" : {
// 				  "bogons" : 100
// 				},
// 				"Time" : 1662058438114304000
// 			  }
// 			],
// 			"Links" : [
// 			  {
// 				"TraceID" : "ab4fe1de2d9c951411bbdbc3533d747f",
// 				"SpanID" : "44c39ef193f3e08G",
// 				"TraceState" : {
// 				  "rojo" : "00f067aa0ba902b7",
// 				  "congo" : "t61rcWkgMzE"
// 				},
// 				"Attributes" : {
// 				  "num1" : 100
// 				}
// 			  }
// 			],
// 			"Status" : {
// 			  "Code": 0,
// 			  "CodeDesc" : "Unset",
// 			  "Description" : ""
// 			},
// 			"Resource" : {
// 			  "job_id" : "13bf135ce8b1481e9329a5e3b62171ae",
// 			  "service" : {
// 				"name" : "服务A",
// 				"version" : "版本1.0.1"
// 			  },
// 			  "telemetry" : {
// 				"sdk" : {
// 				  "language" : "go",
// 				  "name" : "opentelemetry",
// 				  "version" : "1.9.0"
// 				}
// 			  }
// 			},
// 			"catagory" : "trace"
// 		  }`

// 		spanJson2 := `{
// 			"Name" : "Span2-2",
// 			"SpanContext" : {
// 			  "TraceID" : "ab4fe1de2d9c951411bbdbc3533d747e",
// 			  "SpanID" : "44c39ef193f3e07b",
// 			  "TraceFlags" : "01",
// 			  "TraceState" : {
// 				"rojo" : "00f067aa0ba902b7",
// 				"congo" : "t61rcWkgMzE"
// 			  }
// 			},
// 			"Parent" : {
// 			  "TraceID" : "ab4fe1de2d9c951411bbdbc3533d747e",
// 			  "SpanID" : "44c39ef193f3e08b",
// 			  "TraceFlags" : "01",
// 			  "TraceState" : {
// 				"rojo" : "00f067aa0ba902b7",
// 				"congo" : "t61rcWkgMzE"
// 			  }
// 			},
// 			"SpanKind" : 1,
// 			"StartTime" : 1662144838114304000,
// 			"EndTime" : 1662144898114435000,
// 			"@timestamp" : "2022-09-02T18:53:58.114Z",
// 			"Duration" : 131000,
// 			"Attributes" : {
// 			  "num" : "45"
// 			},
// 			"Events" : [
// 			  {
// 				"Name" : "Nice operation!",
// 				"Attributes" : {
// 				  "bogons" : 100
// 				},
// 				"Time" : 1662058438114304000
// 			  }
// 			],
// 			"Links" : [
// 			  {
// 				"TraceID" : "ab4fe1de2d9c951411bbdbc3533d747f",
// 				"SpanID" : "44c39ef193f3e08G",
// 				"TraceState" : {
// 				  "rojo" : "00f067aa0ba902b7",
// 				  "congo" : "t61rcWkgMzE"
// 				},
// 				"Attributes" : {
// 				  "num1" : 100
// 				}
// 			  }
// 			],
// 			"Status" : {
// 			  "Code": 0,
// 			  "CodeDesc" : "Unset",
// 			  "Description" : ""
// 			},
// 			"Resource" : {
// 			  "job_id" : "13bf135ce8b1481e9329a5e3b62171ae",
// 			  "service" : {
// 				"name" : "服务B",
// 				"version" : "版本1.0.1"
// 			  },
// 			  "telemetry" : {
// 				"sdk" : {
// 				  "language" : "go",
// 				  "name" : "opentelemetry",
// 				  "version" : "1.9.0"
// 				}
// 			  }
// 			},
// 			"catagory" : "trace"
// 		  }`

// 		spanJson3 := `{
// 			"Name" : "Span2-3",
// 			"SpanContext" : {
// 			  "TraceID" : "ab4fe1de2d9c951411bbdbc3533d747e",
// 			  "SpanID" : "44c39ef193f3e06b",
// 			  "TraceFlags" : "01",
// 			  "TraceState" : {
// 				"rojo" : "00f067aa0ba902b7",
// 				"congo" : "t61rcWkgMzE"
// 			  }
// 			},
// 			"Parent" : {
// 			  "TraceID" : "ab4fe1de2d9c951411bbdbc3533d747e",
// 			  "SpanID" : "44c39ef193f3e07b",
// 			  "TraceFlags" : "01",
// 			  "TraceState" : {
// 				"rojo" : "00f067aa0ba902b7",
// 				"congo" : "t61rcWkgMzE"
// 			  }
// 			},
// 			"SpanKind" : 1,
// 			"StartTime" : 1662144898114304000,
// 			"EndTime" : 1662144958114435000,
// 			"@timestamp" : "2022-09-02T18:54:58.114Z",
// 			"Duration" : 131000,
// 			"Attributes" : {
// 			  "num" : "45"
// 			},
// 			"Events" : [
// 			  {
// 				"name" : "Nice operation!",
// 				"attributes" : {
// 				  "bogons" : 100
// 				},
// 				"time" : 1662058438114304000
// 			  }
// 			],
// 			"Links" : [
// 			  {
// 				"TraceID" : "ab4fe1de2d9c951411bbdbc3533d747f",
// 				"SpanID" : "44c39ef193f3e08G",
// 				"TraceState" : {
// 				  "rojo" : "00f067aa0ba902b7",
// 				  "congo" : "t61rcWkgMzE"
// 				},
// 				"Attributes" : {
// 				  "num1" : 100
// 				}
// 			  }
// 			],
// 			"Status" : {
// 			  "Code": 0,
// 			  "CodeDesc" : "Unset",
// 			  "Description" : ""
// 			},
// 			"Resource" : {
// 			  "job_id" : "13bf135ce8b1481e9329a5e3b62171ae",
// 			  "service" : {
// 				"name" : "服务D",
// 				"version" : "版本1.0.1"
// 			  },
// 			  "telemetry" : {
// 				"sdk" : {
// 				  "language" : "go",
// 				  "name" : "opentelemetry",
// 				  "version" : "1.9.0"
// 				}
// 			  }
// 			},
// 			"catagory" : "trace"
// 		  }`

// 		span0 := make(map[string]interface{}, 0)
// 		err := sonic.Unmarshal([]byte(spanJson0), &span0)
// 		So(err, ShouldBeNil)

// 		span1 := make(map[string]interface{}, 0)
// 		err = sonic.Unmarshal([]byte(spanJson1), &span1)
// 		So(err, ShouldBeNil)

// 		span2 := make(map[string]interface{}, 0)
// 		err = sonic.Unmarshal([]byte(spanJson2), &span2)
// 		So(err, ShouldBeNil)

// 		span3 := make(map[string]interface{}, 0)
// 		err = sonic.Unmarshal([]byte(spanJson3), &span3)
// 		So(err, ShouldBeNil)

// 		ctx := context.WithValue(context.Background(), rest.XLangKey, "zh-CN")
// 		Convey("Get trace detail failed, caused by data manager with error response", func() {
// 			err := errors.New("get queryfilters failed")
// 			dvAccessMock.EXPECT().GetDataViewQueryFilters(gomock.Any()).AnyTimes().Return(
// 				interfaces.DataView{}, http.StatusInternalServerError, err)

// 			expectedErr := &rest.HTTPError{
// 				HTTPCode: http.StatusInternalServerError,
// 				Language: "zh-CN",
// 				BaseError: rest.BaseError{
// 					ErrorCode:    uerrors.Uniquery_InternalError,
// 					Description:  "服务器内部错误",
// 					Solution:     "请重试该操作，若再次出现该错误请提交工单或联系技术支持工程师。",
// 					ErrorLink:    "暂无",
// 					ErrorDetails: err.Error(),
// 				},
// 			}

// 			res, httpErr := tsMock.GetTraceDetail(testCtx,dataViewId, traceId)
// 			So(res, ShouldResemble, interfaces.TraceDetail{})
// 			So(httpErr, ShouldResemble, expectedErr)
// 		})

// 		Convey("Get trace detail failed, caused by the dataview was not found", func() {
// 			dvAccessMock.EXPECT().GetDataViewQueryFilters(gomock.Any()).AnyTimes().Return(
// 				interfaces.DataView{}, http.StatusOK, nil)

// 			expectedErr := &rest.HTTPError{
// 				HTTPCode: http.StatusBadRequest,
// 				Language: "zh-CN",
// 				BaseError: rest.BaseError{
// 					ErrorCode:    uerrors.Uniquery_Trace_LogGroupNotFound,
// 					Description:  "指定的日志分组不存在",
// 					Solution:     "请检查参数是否正确。",
// 					ErrorLink:    "暂无",
// 					ErrorDetails: "The dataView whose dataViewId equals " + dataViewId + " was not found!",
// 				},
// 			}

// 			res, httpErr := tsMock.GetTraceDetail(testCtx,dataViewId, traceId)
// 			So(res, ShouldResemble, interfaces.TraceDetail{})
// 			So(httpErr, ShouldResemble, expectedErr)
// 		})

// 		Convey("Get trace detail failed, caused by opensearch with error response", func() {
// 			dvAccessMock.EXPECT().GetDataViewQueryFilters(gomock.Any()).AnyTimes().Return(
// 				expectedDataView, http.StatusOK, nil)

// 			err := errors.New("opensearch error")
// 			patch1 := ApplyFunc(searchTraceDetail,
// 				func(ts *traceService, query map[string]interface{}, indices []string) (interfaces.TraceDetail, error) {
// 					return interfaces.TraceDetail{}, err
// 				})
// 			defer patch1.Reset()

// 			expectedErr := &rest.HTTPError{
// 				HTTPCode: http.StatusInternalServerError,
// 				Language: "zh-CN",
// 				BaseError: rest.BaseError{
// 					ErrorCode:    uerrors.Uniquery_InternalError,
// 					Description:  "服务器内部错误",
// 					Solution:     "请重试该操作，若再次出现该错误请提交工单或联系技术支持工程师。",
// 					ErrorLink:    "暂无",
// 					ErrorDetails: err.Error(),
// 				},
// 			}

// 			res, httpErr := tsMock.GetTraceDetail(testCtx,dataViewId, traceId)
// 			So(res, ShouldResemble, interfaces.TraceDetail{})
// 			So(httpErr, ShouldResemble, expectedErr)
// 		})

// 		Convey("Get trace detail failed, the trace was not found", func() {
// 			dvAccessMock.EXPECT().GetDataViewQueryFilters(gomock.Any()).AnyTimes().Return(
// 				expectedDataView, http.StatusOK, nil)

// 			patch1 := ApplyFunc(searchTraceDetail,
// 				func(ts *traceService, query map[string]interface{}, indices []string) (interfaces.TraceDetail, error) {
// 					return interfaces.TraceDetail{}, nil
// 				})
// 			defer patch1.Reset()

// 			expectedErr := &rest.HTTPError{
// 				HTTPCode: http.StatusNotFound,
// 				Language: "zh-CN",
// 				BaseError: rest.BaseError{
// 					ErrorCode:    uerrors.Uniquery_Trace_TraceNotFound,
// 					Description:  "指定的链路不存在",
// 					Solution:     "请检查参数是否正确。",
// 					ErrorLink:    "暂无",
// 					ErrorDetails: "The trace whose traceId equals " + traceId + " was not found!",
// 				},
// 			}

// 			res, httpErr := tsMock.GetTraceDetail(testCtx,dataViewId, traceId)
// 			So(res, ShouldResemble, interfaces.TraceDetail{})
// 			So(httpErr, ShouldResemble, expectedErr)
// 		})

// 		Convey("Get trace detail successed, and trace status is ok", func() {
// 			dvAccessMock.EXPECT().GetDataViewQueryFilters(gomock.Any()).AnyTimes().Return(
// 				expectedDataView, http.StatusOK, nil)

// 			expectedTraceDetail := interfaces.TraceDetail{
// 				TraceStatus: "ok",
// 				SpanStats: map[string]int{
// 					"Ok":    0,
// 					"Error": 0,
// 					"Unset": 3,
// 				},
// 				Spans: []map[string]interface{}{span1, span2, span3},
// 				Services: []interfaces.Service{
// 					{
// 						Name: "服务D",
// 					},
// 					{
// 						Name: "服务B",
// 					},
// 					{
// 						Name: "服务A",
// 					},
// 				},
// 			}
// 			patch := ApplyFunc(searchTraceDetail,
// 				func(ts *traceService, query map[string]interface{}, indices []string) (interfaces.TraceDetail, error) {
// 					return expectedTraceDetail, nil
// 				})
// 			defer patch.Reset()

// 			res, httpErr := tsMock.GetTraceDetail(testCtx,dataViewId, traceId)
// 			So(res.Spans, ShouldResemble, expectedTraceDetail.Spans)
// 			So(res.TraceStatus, ShouldEqual, "ok")
// 			So(httpErr, ShouldBeNil)
// 		})

// 		Convey("Get trace detail successed, and trace status is error", func() {
// 			dvAccessMock.EXPECT().GetDataViewQueryFilters(gomock.Any()).AnyTimes().Return(
// 				expectedDataView, http.StatusOK, nil)

// 			expectedTraceDetail := interfaces.TraceDetail{
// 				TraceStatus: "error",
// 				SpanStats: map[string]int{
// 					"Ok":    0,
// 					"Error": 1,
// 					"Unset": 2,
// 				},
// 				Spans: []map[string]interface{}{span0, span2, span3},
// 				Services: []interfaces.Service{
// 					{
// 						Name: "服务D",
// 					},
// 					{
// 						Name: "服务B",
// 					},
// 					{
// 						Name: "服务A",
// 					},
// 				},
// 			}
// 			patch := ApplyFunc(searchTraceDetail,
// 				func(ts *traceService, query map[string]interface{}, indices []string) (interfaces.TraceDetail, error) {
// 					return expectedTraceDetail, nil
// 				})
// 			defer patch.Reset()

// 			res, httpErr := tsMock.GetTraceDetail(testCtx,dataViewId, traceId)
// 			So(res.Spans, ShouldResemble, expectedTraceDetail.Spans)
// 			So(res.TraceStatus, ShouldEqual, "error")
// 			So(httpErr, ShouldBeNil)
// 		})

// 	})
// }

// func TestSearchSpanList(t *testing.T) {
// 	Convey("Test GetSpanList", t, func() {
// 		ctrl := gomock.NewController(t)
// 		defer ctrl.Finish()

// 		osAccessMock := umock.NewMockOpenSearchAccess(ctrl)
// 		dvAccessMock := umock.NewMockLogGroupAccess(ctrl)

// 		tsMock := MockNewTraceService(osAccessMock, dvAccessMock)

// 		okResp := `{
// 			"took" : 5,
// 			"timed_out" : false,
// 			"_shards" : {
// 			  "total" : 2,
// 			  "successful" : 2,
// 			  "skipped" : 0,
// 			  "failed" : 0
// 			},
// 			"hits" : {
// 			  "total" : {
// 				"value" : 2,
// 				"relation" : "eq"
// 			  },
// 			  "max_score" : null,
// 			  "hits" : [
// 				{
// 				  "_index" : "trace-01",
// 				  "_type" : "_doc",
// 				  "_id" : "N9sJfYYBfz_Qitm7ZCmp",
// 				  "_score" : null,
// 				  "_source" : {
// 					"Status" : {
// 					  "code" : "Ok"
// 					},
// 					"SpanContext" : {
// 					  "TraceID" : "ab4fe1de2d9c951411bbdbc3533d747f"
// 					},
// 					"StartTime" : "2022-10-17T18:51:58.114304Z",
// 					"Duration" : 131,
// 					"Name" : "Span1-1"
// 				  },
// 				  "sort" : [
// 					1666032718114
// 				  ]
// 				},
// 				{
// 				  "_index" : "trace-02",
// 				  "_type" : "_doc",
// 				  "_id" : "M9sJfYYBfz_Qitm7Vylt",
// 				  "_score" : null,
// 				  "_source" : {
// 					"Status" : {
// 					  "code" : "Error"
// 					},
// 					"SpanContext" : {
// 					  "TraceID" : "ab4fe1de2d9c951411bbdbc3533d747e"
// 					},
// 					"StartTime" : "2022-09-02T18:52:58.114304Z",
// 					"Duration" : 131,
// 					"Name" : "Span2-1"
// 				  },
// 				  "sort" : [
// 					1662144778114
// 				  ]
// 				}
// 			  ]
// 			}
// 		  }
// 		  `
// 		errResp := `{
// 			"status": 400,
// 			"error": {
// 				"root_cause": [{
// 					"type": "illegal_argument_exception",
// 					"reason": "this node does not have the remote_cluster_client role"
// 				}],
// 				"type": "illegal_argument_exception",
// 				"reason": "this node does not have the remote_cluster_client role"
// 			}
// 		}`

// 		dsl := map[string]interface{}{
// 			"_source": map[string]interface{}{
// 				"includes": []string{"Name", "StartTime", "SpanContext.TraceID", "Duration", "Status.Code"},
// 			},
// 			"track_total_hits": true,
// 			"sort": []map[string]interface{}{
// 				{
// 					"StartTime": map[string]string{
// 						"order": "desc",
// 					},
// 				},
// 			},
// 		}

// 		// 构造时间范围过滤器
// 		timeRangeFilter := map[string]interface{}{
// 			"range": map[string]interface{}{
// 				"StartTime": map[string]interface{}{
// 					"gte": "2020-10-17T18:51:58.114304Z",
// 					"lte": "2023-02-20T18:51:58.114304Z",
// 				},
// 			},
// 		}

// 		// 构造父TraceID过滤器
// 		parentTraceIDFilter := map[string]interface{}{
// 			"term": map[string]interface{}{
// 				"Parent.TraceID.keyword": map[string]string{
// 					"value": "00000000000000000000000000000000",
// 				},
// 			},
// 		}

// 		indices := []string{"trace-*"}

// 		Convey("Get trace list failed, caused by opensearch response err", func() {
// 			osErrResBody := io.NopCloser(strings.NewReader(errResp))
// 			resBytes, _ := ioutil.ReadAll(osErrResBody)

// 			oerr := errors.New(string(resBytes))
// 			osAccessMock.EXPECT().SearchSpans(gomock.Any(), gomock.Any()).AnyTimes().Return(
// 				nil, oerr)

// 			mustFilter := append([]interface{}{}, timeRangeFilter, parentTraceIDFilter)
// 			dsl["query"] = map[string]interface{}{
// 				"bool": map[string]interface{}{
// 					"must": mustFilter,
// 				},
// 			}

// 			res, total, err := searchSpanList(tsMock, dsl, indices)
// 			So(res, ShouldResemble, []interfaces.SpanList{})
// 			So(total, ShouldEqual, 0)
// 			So(err, ShouldResemble, oerr)
// 		})

// 		Convey("Get all trace list successed", func() {
// 			osOkResBody := io.NopCloser(strings.NewReader(okResp))
// 			resBytes, _ := ioutil.ReadAll(osOkResBody)

// 			osAccessMock.EXPECT().SearchSpans(gomock.Any(), gomock.Any()).AnyTimes().Return(
// 				resBytes, nil)

// 			expectedRes := []interfaces.SpanList{
// 				{
// 					Name:       "Span1-1",
// 					TraceID:    "ab4fe1de2d9c951411bbdbc3533d747f",
// 					SpanStatus: "Ok",
// 					StartTime:  int64(1666032718114304),
// 					Duration:   131,
// 					SpanKind:   "UNSPECIFIED",
// 				},
// 				{
// 					Name:       "Span2-1",
// 					TraceID:    "ab4fe1de2d9c951411bbdbc3533d747e",
// 					SpanStatus: "Error",
// 					StartTime:  int64(1662144778114304),
// 					Duration:   131,
// 					SpanKind:   "UNSPECIFIED",
// 				},
// 			}

// 			res, total, err := searchSpanList(tsMock, dsl, indices)
// 			So(res, ShouldResemble, expectedRes)
// 			So(total, ShouldEqual, 2)
// 			So(err, ShouldBeNil)
// 		})
// 	})
// }

// to delete
// func TestSearchTraceDetail(t *testing.T) {
// 	Convey("Test searchTraceDetail", t, func() {
// 		ctrl := gomock.NewController(t)
// 		defer ctrl.Finish()

// 		appSetting := &common.AppSetting{
// 			TraceSetting: common.TraceSetting{
// 				MaxSearchSpanSize: 1000,
// 			},
// 		}
// 		osAccessMock := umock.NewMockOpenSearchAccess(ctrl)
// 		dvAccessMock := umock.NewMockLogGroupAccess(ctrl)

// 		tsMock := MockNewTraceService(appSetting, osAccessMock, dvAccessMock)

// 		okResp := `{
// 			"took" : 5,
// 			"timed_out" : false,
// 			"_shards" : {
// 			  "total" : 2,
// 			  "successful" : 2,
// 			  "skipped" : 0,
// 			  "failed" : 0
// 			},
// 			"hits" : {
// 			  "total" : {
// 				"value" : 4,
// 				"relation" : "eq"
// 			  },
// 			  "max_score" : null,
// 			  "hits" : [
// 				{
// 				  "_index" : "trace-02",
// 				  "_type" : "_doc",
// 				  "_id" : "M9sJfYYBfz_Qitm7Vylt",
// 				  "_score" : null,
// 				  "_source" : {
// 					"Name" : "Span2-1",
// 					"SpanContext" : {
// 					  "TraceID" : "ab4fe1de2d9c951411bbdbc3533d747e",
// 					  "SpanID" : "44c39ef193f3e08b",
// 					  "TraceFlags" : "01",
// 					  "TraceState" : {
// 						"rojo" : "00f067aa0ba902b7",
// 						"congo" : "t61rcWkgMzE"
// 					  }
// 					},
// 					"Parent" : {
// 					  "TraceID" : "00000000000000000000000000000000",
// 					  "SpanID" : "0000000000000000",
// 					  "TraceFlags" : "00",
// 					  "TraceState" : null
// 					},
// 					"SpanKind" : 1,
// 					"StartTime" : 1662144778114304000,
// 					"EndTime" : 1662144838114435000,
// 					"@timestamp" : "2022-09-02T18:52:58.114Z",
// 					"Duration" : 131000,
// 					"Attributes" : {
// 					  "num" : "45"
// 					},
// 					"Events" : [
// 					  {
// 						"Name" : "Nice operation!",
// 						"Attributes" : {
// 						  "bogons" : 100
// 						},
// 						"Time" : 1662058378114304000
// 					  }
// 					],
// 					"Links" : [
// 					  {
// 						"TraceID" : "ab4fe1de2d9c951411bbdbc3533d747f",
// 						"SpanID" : "44c39ef193f3e08G",
// 						"TraceState" : {
// 						  "rojo" : "00f067aa0ba902b7",
// 						  "congo" : "t61rcWkgMzE"
// 						},
// 						"Attributes" : {
// 						  "num1" : 100
// 						}
// 					  }
// 					],
// 					"Status" : {
// 					  "Code": 0,
// 					  "CodeDesc" : "Unset",
// 					  "Description" : ""
// 					},
// 					"Resource" : {
// 					  "job_id" : "13bf135ce8b1481e9329a5e3b62171ae",
// 					  "service" : {
// 						"name" : "服务A",
// 						"version" : "版本1.0.1"
// 					  },
// 					  "telemetry" : {
// 						"sdk" : {
// 						  "language" : "go",
// 						  "name" : "opentelemetry",
// 						  "version" : "1.9.0"
// 						}
// 					  }
// 					},
// 					"catagory" : "trace"
// 				  },
// 				  "sort" : [
// 					1662144778114
// 				  ]
// 				},
// 				{
// 				  "_index" : "trace-02",
// 				  "_type" : "_doc",
// 				  "_id" : "NNsJfYYBfz_Qitm7Vylt",
// 				  "_score" : null,
// 				  "_source" : {
// 					"Name" : "Span2-2",
// 					"SpanContext" : {
// 					  "TraceID" : "ab4fe1de2d9c951411bbdbc3533d747e",
// 					  "SpanID" : "44c39ef193f3e07b",
// 					  "TraceFlags" : "01",
// 					  "TraceState" : {
// 						"rojo" : "00f067aa0ba902b7",
// 						"congo" : "t61rcWkgMzE"
// 					  }
// 					},
// 					"Parent" : {
// 					  "TraceID" : "ab4fe1de2d9c951411bbdbc3533d747e",
// 					  "SpanID" : "44c39ef193f3e08b",
// 					  "TraceFlags" : "01",
// 					  "TraceState" : {
// 						"rojo" : "00f067aa0ba902b7",
// 						"congo" : "t61rcWkgMzE"
// 					  }
// 					},
// 					"SpanKind" : 1,
// 					"StartTime" : 1662144838114304000,
// 					"EndTime" : 1662144898114435000,
// 					"@timestamp" : "2022-09-02T18:53:58.114Z",
// 					"Duration" : 131,
// 					"Attributes" : {
// 					  "num" : "45"
// 					},
// 					"Events" : [
// 					  {
// 						"Name" : "Nice operation!",
// 						"Attributes" : {
// 						  "bogons" : 100
// 						},
// 						"Time" : 1662058438114304000
// 					  }
// 					],
// 					"Links" : [
// 					  {
// 						"TraceID" : "ab4fe1de2d9c951411bbdbc3533d747f",
// 						"SpanID" : "44c39ef193f3e08G",
// 						"TraceState" : {
// 						  "rojo" : "00f067aa0ba902b7",
// 						  "congo" : "t61rcWkgMzE"
// 						},
// 						"Attributes" : {
// 						  "num1" : 100
// 						}
// 					  }
// 					],
// 					"Status" : {
// 					  "Code": 0,
// 					  "CodeDesc" : "Unset",
// 					  "Description" : ""
// 					},
// 					"Resource" : {
// 					  "job_id" : "13bf135ce8b1481e9329a5e3b62171ae",
// 					  "service" : {
// 						"name" : "服务B",
// 						"version" : "版本1.0.1"
// 					  },
// 					  "telemetry" : {
// 						"sdk" : {
// 						  "language" : "go",
// 						  "name" : "opentelemetry",
// 						  "version" : "1.9.0"
// 						}
// 					  }
// 					},
// 					"catagory" : "trace"
// 				  },
// 				  "sort" : [
// 					1662144838114
// 				  ]
// 				},
// 				{
// 				  "_index" : "trace-02",
// 				  "_type" : "_doc",
// 				  "_id" : "NdsJfYYBfz_Qitm7Vylt",
// 				  "_score" : null,
// 				  "_source" : {
// 					"Name" : "Span2-3",
// 					"SpanContext" : {
// 					  "TraceID" : "ab4fe1de2d9c951411bbdbc3533d747e",
// 					  "SpanID" : "44c39ef193f3e06b",
// 					  "TraceFlags" : "01",
// 					  "TraceState" : {
// 						"rojo" : "00f067aa0ba902b7",
// 						"congo" : "t61rcWkgMzE"
// 					  }
// 					},
// 					"Parent" : {
// 					  "TraceID" : "ab4fe1de2d9c951411bbdbc3533d747e",
// 					  "SpanID" : "44c39ef193f3e07b",
// 					  "TraceFlags" : "01",
// 					  "TraceState" : {
// 						"rojo" : "00f067aa0ba902b7",
// 						"congo" : "t61rcWkgMzE"
// 					  }
// 					},
// 					"SpanKind" : 1,
// 					"StartTime" : 1662144898114304000,
// 					"EndTime" : 1662144958114435000,
// 					"@timestamp" : "2022-09-02T18:54:58.114Z",
// 					"Duration" : 131,
// 					"Attributes" : {
// 					  "num" : "45"
// 					},
// 					"Events" : [
// 					  {
// 						"name" : "Nice operation!",
// 						"attributes" : {
// 						  "bogons" : 100
// 						},
// 						"time" : 1662058438114304000
// 					  }
// 					],
// 					"Links" : [
// 					  {
// 						"TraceID" : "ab4fe1de2d9c951411bbdbc3533d747f",
// 						"SpanID" : "44c39ef193f3e08G",
// 						"TraceState" : {
// 						  "rojo" : "00f067aa0ba902b7",
// 						  "congo" : "t61rcWkgMzE"
// 						},
// 						"Attributes" : {
// 						  "num1" : 100
// 						}
// 					  }
// 					],
// 					"Status" : {
// 					  "Code": 0,
// 					  "CodeDesc" : "Unset",
// 					  "Description" : ""
// 					},
// 					"Resource" : {
// 					  "job_id" : "13bf135ce8b1481e9329a5e3b62171ae",
// 					  "service" : {
// 						"name" : "服务D",
// 						"version" : "版本1.0.1"
// 					  },
// 					  "telemetry" : {
// 						"sdk" : {
// 						  "language" : "go",
// 						  "name" : "opentelemetry",
// 						  "version" : "1.9.0"
// 						}
// 					  }
// 					},
// 					"catagory" : "trace"
// 				  },
// 				  "sort" : [
// 					1662144898114
// 				  ]
// 				},
// 				{
// 				  "_index" : "trace-02",
// 				  "_type" : "_doc",
// 				  "_id" : "NtsJfYYBfz_Qitm7Vylt",
// 				  "_score" : null,
// 				  "_source" : {
// 					"Name" : "Span2-4",
// 					"SpanContext" : {
// 					  "TraceID" : "ab4fe1de2d9c951411bbdbc3533d747e",
// 					  "SpanID" : "44c39ef193f3e05b",
// 					  "TraceFlags" : "01",
// 					  "TraceState" : {
// 						"rojo" : "00f067aa0ba902b7",
// 						"congo" : "t61rcWkgMzE"
// 					  }
// 					},
// 					"Parent" : {
// 					  "TraceID" : "ab4fe1de2d9c951411bbdbc3533d747e",
// 					  "SpanID" : "44c39ef193f3e06b",
// 					  "TraceFlags" : "01",
// 					  "TraceState" : {
// 						"rojo" : "00f067aa0ba902b7",
// 						"congo" : "t61rcWkgMzE"
// 					  }
// 					},
// 					"SpanKind" : 1,
// 					"StartTime" : 1662144958114304000,
// 					"EndTime" : 1662145018114435000,
// 					"@timestamp" : "2022-09-02T18:55:58.114Z",
// 					"Duration" : 131,
// 					"Attributes" : {
// 					  "num" : "45"
// 					},
// 					"Events" : [
// 					  {
// 						"Name" : "Nice operation!",
// 						"Attributes" : {
// 						  "bogons" : 100
// 						},
// 						"Time" : 1662058438114304000
// 					  }
// 					],
// 					"Links" : [
// 					  {
// 						"Trace_id" : "ab4fe1de2d9c951411bbdbc3533d747f",
// 						"Span_id" : "44c39ef193f3e08G",
// 						"Trace_state" : {
// 						  "rojo" : "00f067aa0ba902b7",
// 						  "congo" : "t61rcWkgMzE"
// 						},
// 						"Attributes" : {
// 						  "num1" : 100
// 						}
// 					  }
// 					],
// 					"Status" : {
// 					  "Code": 0,
// 					  "CodeDesc" : "Unset",
// 				      "Description" : ""
// 					},
// 					"Resource" : {
// 					  "job_id" : "13bf135ce8b1481e9329a5e3b62171ae",
// 					  "service" : {
// 						"name" : "服务E",
// 						"version" : "版本1.0.1"
// 					  },
// 					  "telemetry" : {
// 						"sdk" : {
// 						  "language" : "go",
// 						  "name" : "opentelemetry",
// 						  "version" : "1.9.0"
// 						}
// 					  }
// 					},
// 					"catagory" : "trace"
// 				  },
// 				  "sort" : [
// 					1662144958114
// 				  ]
// 				}
// 			  ]
// 			},
// 			"aggregations" : {
// 			  "group_by_status" : {
// 				"doc_count_error_upper_bound" : 0,
// 				"sum_other_doc_count" : 0,
// 				"buckets" : [
// 				  {
// 					"key" : "Unset",
// 					"doc_count" : 4
// 				  }
// 				]
// 			  }
// 			}
// 		  }`

// 		nonstandardResp := `{
// 			"took" : 5,
// 			"timed_out" : false,
// 			"_shards" : {
// 			  "total" : 2,
// 			  "successful" : 2,
// 			  "skipped" : 0,
// 			  "failed" : 0
// 			},
// 			"hits" : {
// 			  "total" : {
// 				"value" : 4,
// 				"relation" : "eq"
// 			  },
// 			  "max_score" : null,
// 			  "hits" : [
// 				{
// 				  "_index" : "trace-02",
// 				  "_type" : "_doc",
// 				  "_id" : "M9sJfYYBfz_Qitm7Vylt",
// 				  "_score" : null,
// 				  "_source" : {
// 					"Name" : "Span2-1",
// 					"SpanContext" : {
// 					  "TraceID" : "ab4fe1de2d9c951411bbdbc3533d747e",
// 					  "SpanID" : "44c39ef193f3e08b",
// 					  "TraceFlags" : "01",
// 					  "TraceState" : {
// 						"rojo" : "00f067aa0ba902b7",
// 						"congo" : "t61rcWkgMzE"
// 					  }
// 					},
// 					"Parent" : {
// 					  "TraceID" : "00000000000000000000000000000000",
// 					  "SpanID" : "0000000000000000",
// 					  "TraceFlags" : "00",
// 					  "TraceState" : null
// 					},
// 					"SpanKind" : 1,
// 					"StartTime" : "1662144778114304000",
// 					"EndTime" : "1662144838114435000",
// 					"@timestamp" : "2022-09-02T18:52:58.114Z",
// 					"Duration" : 131000,
// 					"Attributes" : {
// 					  "num" : "45"
// 					},
// 					"Events" : [
// 					  {
// 						"Name" : "Nice operation!",
// 						"Attributes" : {
// 						  "bogons" : 100
// 						},
// 						"Time" : 1662058378114304000
// 					  }
// 					],
// 					"Links" : [
// 					  {
// 						"TraceID" : "ab4fe1de2d9c951411bbdbc3533d747f",
// 						"SpanID" : "44c39ef193f3e08G",
// 						"TraceState" : {
// 						  "rojo" : "00f067aa0ba902b7",
// 						  "congo" : "t61rcWkgMzE"
// 						},
// 						"Attributes" : {
// 						  "num1" : 100
// 						}
// 					  }
// 					],
// 					"Status" : {
// 					  "Code": 0,
// 					  "CodeDesc" : "Unset",
// 					  "Description" : ""
// 					},
// 					"Resource" : {
// 					  "job_id" : "13bf135ce8b1481e9329a5e3b62171ae",
// 					  "service" : {
// 						"name" : "服务A",
// 						"version" : "版本1.0.1"
// 					  },
// 					  "telemetry" : {
// 						"sdk" : {
// 						  "language" : "go",
// 						  "name" : "opentelemetry",
// 						  "version" : "1.9.0"
// 						}
// 					  }
// 					},
// 					"catagory" : "trace"
// 				  },
// 				  "sort" : [
// 					1662144778114
// 				  ]
// 				},
// 				{
// 				  "_index" : "trace-02",
// 				  "_type" : "_doc",
// 				  "_id" : "NNsJfYYBfz_Qitm7Vylt",
// 				  "_score" : null,
// 				  "_source" : {
// 					"Name" : "Span2-2",
// 					"SpanContext" : {
// 					  "TraceID" : "ab4fe1de2d9c951411bbdbc3533d747e",
// 					  "SpanID" : "44c39ef193f3e07b",
// 					  "TraceFlags" : "01",
// 					  "TraceState" : {
// 						"rojo" : "00f067aa0ba902b7",
// 						"congo" : "t61rcWkgMzE"
// 					  }
// 					},
// 					"Parent" : {
// 					  "TraceID" : "ab4fe1de2d9c951411bbdbc3533d747e",
// 					  "SpanID" : "44c39ef193f3e08b",
// 					  "TraceFlags" : "01",
// 					  "TraceState" : {
// 						"rojo" : "00f067aa0ba902b7",
// 						"congo" : "t61rcWkgMzE"
// 					  }
// 					},
// 					"SpanKind" : 1,
// 					"StartTime" : 1662144838114304000,
// 					"EndTime" : 1662144898114435000,
// 					"@timestamp" : "2022-09-02T18:53:58.114Z",
// 					"Duration" : 131,
// 					"Attributes" : {
// 					  "num" : "45"
// 					},
// 					"Events" : [
// 					  {
// 						"Name" : "Nice operation!",
// 						"Attributes" : {
// 						  "bogons" : 100
// 						},
// 						"Time" : 1662058438114304000
// 					  }
// 					],
// 					"Links" : [
// 					  {
// 						"TraceID" : "ab4fe1de2d9c951411bbdbc3533d747f",
// 						"SpanID" : "44c39ef193f3e08G",
// 						"TraceState" : {
// 						  "rojo" : "00f067aa0ba902b7",
// 						  "congo" : "t61rcWkgMzE"
// 						},
// 						"Attributes" : {
// 						  "num1" : 100
// 						}
// 					  }
// 					],
// 					"Status" : {
// 					  "Code": 0,
// 					  "CodeDesc" : "Unset",
// 					  "Description" : ""
// 					},
// 					"Resource" : {
// 					  "job_id" : "13bf135ce8b1481e9329a5e3b62171ae",
// 					  "service" : {
// 						"name" : "服务B",
// 						"version" : "版本1.0.1"
// 					  },
// 					  "telemetry" : {
// 						"sdk" : {
// 						  "language" : "go",
// 						  "name" : "opentelemetry",
// 						  "version" : "1.9.0"
// 						}
// 					  }
// 					},
// 					"catagory" : "trace"
// 				  },
// 				  "sort" : [
// 					1662144838114
// 				  ]
// 				},
// 				{
// 				  "_index" : "trace-02",
// 				  "_type" : "_doc",
// 				  "_id" : "NdsJfYYBfz_Qitm7Vylt",
// 				  "_score" : null,
// 				  "_source" : {
// 					"Name" : "Span2-3",
// 					"SpanContext" : {
// 					  "TraceID" : "ab4fe1de2d9c951411bbdbc3533d747e",
// 					  "SpanID" : "44c39ef193f3e06b",
// 					  "TraceFlags" : "01",
// 					  "TraceState" : {
// 						"rojo" : "00f067aa0ba902b7",
// 						"congo" : "t61rcWkgMzE"
// 					  }
// 					},
// 					"Parent" : {
// 					  "TraceID" : "ab4fe1de2d9c951411bbdbc3533d747e",
// 					  "SpanID" : "44c39ef193f3e07b",
// 					  "TraceFlags" : "01",
// 					  "TraceState" : {
// 						"rojo" : "00f067aa0ba902b7",
// 						"congo" : "t61rcWkgMzE"
// 					  }
// 					},
// 					"SpanKind" : 1,
// 					"StartTime" : 1662144898114304000,
// 					"EndTime" : 1662144958114435000,
// 					"@timestamp" : "2022-09-02T18:54:58.114Z",
// 					"Duration" : 131,
// 					"Attributes" : {
// 					  "num" : "45"
// 					},
// 					"Events" : [
// 					  {
// 						"name" : "Nice operation!",
// 						"attributes" : {
// 						  "bogons" : 100
// 						},
// 						"time" : 1662058438114304000
// 					  }
// 					],
// 					"Links" : [
// 					  {
// 						"TraceID" : "ab4fe1de2d9c951411bbdbc3533d747f",
// 						"SpanID" : "44c39ef193f3e08G",
// 						"TraceState" : {
// 						  "rojo" : "00f067aa0ba902b7",
// 						  "congo" : "t61rcWkgMzE"
// 						},
// 						"Attributes" : {
// 						  "num1" : 100
// 						}
// 					  }
// 					],
// 					"Status" : {
// 					  "Code": 0,
// 					  "CodeDesc" : "Unset",
// 					  "Description" : ""
// 					},
// 					"Resource" : {
// 					  "job_id" : "13bf135ce8b1481e9329a5e3b62171ae",
// 					  "service" : {
// 						"name" : "服务D",
// 						"version" : "版本1.0.1"
// 					  },
// 					  "telemetry" : {
// 						"sdk" : {
// 						  "language" : "go",
// 						  "name" : "opentelemetry",
// 						  "version" : "1.9.0"
// 						}
// 					  }
// 					},
// 					"catagory" : "trace"
// 				  },
// 				  "sort" : [
// 					1662144898114
// 				  ]
// 				},
// 				{
// 				  "_index" : "trace-02",
// 				  "_type" : "_doc",
// 				  "_id" : "NtsJfYYBfz_Qitm7Vylt",
// 				  "_score" : null,
// 				  "_source" : {
// 					"Name" : "Span2-4",
// 					"SpanContext" : {
// 					  "TraceID" : "ab4fe1de2d9c951411bbdbc3533d747e",
// 					  "SpanID" : "44c39ef193f3e05b",
// 					  "TraceFlags" : "01",
// 					  "TraceState" : {
// 						"rojo" : "00f067aa0ba902b7",
// 						"congo" : "t61rcWkgMzE"
// 					  }
// 					},
// 					"Parent" : {
// 					  "TraceID" : "ab4fe1de2d9c951411bbdbc3533d747e",
// 					  "SpanID" : "44c39ef193f3e06b",
// 					  "TraceFlags" : "01",
// 					  "TraceState" : {
// 						"rojo" : "00f067aa0ba902b7",
// 						"congo" : "t61rcWkgMzE"
// 					  }
// 					},
// 					"SpanKind" : 1,
// 					"StartTime" : 1662144958114304000,
// 					"EndTime" : 1662145018114435000,
// 					"@timestamp" : "2022-09-02T18:55:58.114Z",
// 					"Duration" : 131,
// 					"Attributes" : {
// 					  "num" : "45"
// 					},
// 					"Events" : [
// 					  {
// 						"Name" : "Nice operation!",
// 						"Attributes" : {
// 						  "bogons" : 100
// 						},
// 						"Time" : 1662058438114304000
// 					  }
// 					],
// 					"Links" : [
// 					  {
// 						"Trace_id" : "ab4fe1de2d9c951411bbdbc3533d747f",
// 						"Span_id" : "44c39ef193f3e08G",
// 						"Trace_state" : {
// 						  "rojo" : "00f067aa0ba902b7",
// 						  "congo" : "t61rcWkgMzE"
// 						},
// 						"Attributes" : {
// 						  "num1" : 100
// 						}
// 					  }
// 					],
// 					"Status" : {
// 					  "Code": 0,
// 					  "CodeDesc" : "Unset",
// 				      "Description" : ""
// 					},
// 					"Resource" : {
// 					  "job_id" : "13bf135ce8b1481e9329a5e3b62171ae",
// 					  "service" : {
// 						"name" : "服务E",
// 						"version" : "版本1.0.1"
// 					  },
// 					  "telemetry" : {
// 						"sdk" : {
// 						  "language" : "go",
// 						  "name" : "opentelemetry",
// 						  "version" : "1.9.0"
// 						}
// 					  }
// 					},
// 					"catagory" : "trace"
// 				  },
// 				  "sort" : [
// 					1662144958114
// 				  ]
// 				}
// 			  ]
// 			},
// 			"aggregations" : {
// 			  "group_by_status" : {
// 				"doc_count_error_upper_bound" : 0,
// 				"sum_other_doc_count" : 0,
// 				"buckets" : [
// 				  {
// 					"key" : "Unset",
// 					"doc_count" : 4
// 				  }
// 				]
// 			  }
// 			}
// 		  }`

// 		notFoundResp := `{
// 			"took" : 522,
// 			"timed_out" : false,
// 			"_shards" : {
// 			  "total" : 1400,
// 			  "successful" : 1400,
// 			  "skipped" : 1313,
// 			  "failed" : 0
// 			},
// 			"hits" : {
// 			  "total" : {
// 				"value" : 0,
// 				"relation" : "eq"
// 			  },
// 			  "max_score" : null,
// 			  "hits" : [ ]
// 			}
// 		  }`

// 		errResp := `{
// 			"status": 400,
// 			"error": {
// 				"root_cause": [{
// 					"type": "illegal_argument_exception",
// 					"reason": "this node does not have the remote_cluster_client role"
// 				}],
// 				"type": "illegal_argument_exception",
// 				"reason": "this node does not have the remote_cluster_client role"
// 			}
// 		}`

// 		dsl := map[string]interface{}{
// 			"query": map[string]interface{}{
// 				"term": map[string]interface{}{
// 					"SpanContext.TraceID.keyword": map[string]string{
// 						"value": "ab4fe1de2d9c951411bbdbc3533d747e",
// 					},
// 				},
// 			},
// 			"size": 2000,
// 			"sort": []map[string]interface{}{
// 				{
// 					"StartTime": map[string]string{
// 						"order": "asc",
// 					},
// 				},
// 			},
// 			"aggs": map[string]interface{}{
// 				"group_by_status": map[string]interface{}{
// 					"terms": map[string]string{
// 						"field": "Status.CodeDesc.keyword",
// 					},
// 				},
// 			},
// 		}
// 		indices := []string{"trace-*"}

// 		Convey("Get trace detail failed, caused by opensearch response err", func() {
// 			osErrResBody := io.NopCloser(strings.NewReader(errResp))
// 			resBytes, _ := ioutil.ReadAll(osErrResBody)

// 			oerr := errors.New(string(resBytes))
// 			osAccessMock.EXPECT().SearchSpans(gomock.Any(), gomock.Any()).AnyTimes().Return(
// 				nil, oerr)

// 			res, err := searchTraceDetail(tsMock, dsl, indices)
// 			So(res, ShouldResemble, interfaces.TraceDetail{})
// 			So(err.Error(), ShouldResemble, errResp)
// 		})

// 		Convey("Get trace detail failed, caused by the trace was not found", func() {
// 			osNotFoundResBody := io.NopCloser(strings.NewReader(notFoundResp))
// 			resBytes, _ := ioutil.ReadAll(osNotFoundResBody)

// 			osAccessMock.EXPECT().SearchSpans(gomock.Any(), gomock.Any()).AnyTimes().Return(
// 				resBytes, nil)

// 			res, err := searchTraceDetail(tsMock, dsl, indices)
// 			So(res, ShouldResemble, interfaces.TraceDetail{})
// 			So(err, ShouldBeEmpty)
// 		})

// 		Convey("Get trace detail failed, caused by wrong field type", func() {
// 			osNonstandardResp := io.NopCloser(strings.NewReader(nonstandardResp))
// 			resBytes, _ := ioutil.ReadAll(osNonstandardResp)
// 			osAccessMock.EXPECT().SearchSpans(gomock.Any(), gomock.Any()).AnyTimes().Return(
// 				resBytes, nil)

// 			expectedErr := errors.New("StartTime is not a json.Number, so it can not be converted to an int64")

// 			res, err := searchTraceDetail(tsMock, dsl, indices)
// 			So(len(res.Spans), ShouldEqual, 0)
// 			So(err, ShouldResemble, expectedErr)
// 		})

// 		Convey("Get trace detail succeeded", func() {
// 			osOkResBody := io.NopCloser(strings.NewReader(okResp))
// 			resBytes, _ := ioutil.ReadAll(osOkResBody)
// 			osAccessMock.EXPECT().SearchSpans(gomock.Any(), gomock.Any()).AnyTimes().Return(
// 				resBytes, nil)

// 			res, err := searchTraceDetail(tsMock, dsl, indices)
// 			So(len(res.Spans), ShouldEqual, 4)
// 			So(res.SpanStats["Unset"], ShouldEqual, 4)
// 			So(res.StartTime, ShouldEqual, int64(1662144778114304000))
// 			So(res.EndTime, ShouldEqual, int64(1662145018114435000))
// 			So(res.Duration, ShouldEqual, int64(240000131000))
// 			So(err, ShouldBeNil)
// 		})
// 	})
// }
