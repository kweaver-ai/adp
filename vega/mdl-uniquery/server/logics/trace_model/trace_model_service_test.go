// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package trace_model

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"reflect"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	rest "github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	"uniquery/common"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	umock "uniquery/interfaces/mock"
	"uniquery/logics/trace_model/data_source"
)

var (
	testCtx = context.WithValue(context.WithValue(context.Background(),
		rest.XLangKey, rest.DefaultLanguage),
		interfaces.ACCOUNT_INFO_KEY, interfaces.AccountInfo{
			ID:   interfaces.ADMIN_ID,
			Type: interfaces.ADMIN_TYPE,
		})
)

func MockNewTraceModelService(appSetting *common.AppSetting, tmAccess interfaces.TraceModelAccess,
	dcAccess interfaces.DataConnectionAccess, ps interfaces.PermissionService) (ts *traceModelService) {
	ts = &traceModelService{
		appSetting: appSetting,
		tmAccess:   tmAccess,
		dcAccess:   dcAccess,
		ps:         ps,
	}
	return ts
}

func TestGetSpanList(t *testing.T) {
	Convey("Test GetSpanList", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mockTMAccess := umock.NewMockTraceModelAccess(mockCtrl)
		mockDCAccess := umock.NewMockDataConnectionAccess(mockCtrl)
		mockTMAdapter := umock.NewMockTraceModelAdapter(mockCtrl)
		psMock := umock.NewMockPermissionService(mockCtrl)

		mockTMService := MockNewTraceModelService(appSetting, mockTMAccess, mockDCAccess, psMock)

		Convey("Get failed, caused by the error from method 'getUnderlyingDataSouceType' while searching span", func() {
			expectedErr := errors.New("some errors")
			patch := ApplyPrivateMethod(reflect.TypeOf(&traceModelService{}), "getUnderlyingDataSouceType",
				func(tmService *traceModelService, ctx context.Context, queryCategory string, model interfaces.TraceModel) (string, error) {
					return "", expectedErr
				})
			defer patch.Reset()
			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			reqModel := interfaces.TraceModel{
				EnabledRelatedLog: interfaces.RELATED_LOG_OPEN,
			}
			_, _, err := mockTMService.GetSpanList(testCtx, reqModel, interfaces.SpanListQueryParams{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by the error from method 'NewTraceModelAdapter' while searching span", func() {
			expectedErr := errors.New("some errors")
			patch1 := ApplyPrivateMethod(reflect.TypeOf(&traceModelService{}), "getUnderlyingDataSouceType",
				func(tmService *traceModelService, ctx context.Context, queryCategory string, model interfaces.TraceModel) (string, error) {
					return "", nil
				})
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(data_source.NewTraceModelAdapter, nil, expectedErr)
			defer patch2.Reset()
			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			reqModel := interfaces.TraceModel{
				EnabledRelatedLog: interfaces.RELATED_LOG_OPEN,
			}
			_, _, err := mockTMService.GetSpanList(testCtx, reqModel, interfaces.SpanListQueryParams{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get succeed", func() {
			patch1 := ApplyPrivateMethod(reflect.TypeOf(&traceModelService{}), "getUnderlyingDataSouceType",
				func(tmService *traceModelService, ctx context.Context, queryCategory string, model interfaces.TraceModel) (string, error) {
					return "", nil
				})
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(data_source.NewTraceModelAdapter, mockTMAdapter, nil)
			defer patch2.Reset()

			mockTMAdapter.EXPECT().GetSpanList(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, int64(0), nil)
			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			reqModel := interfaces.TraceModel{
				EnabledRelatedLog: interfaces.RELATED_LOG_OPEN,
			}
			_, _, err := mockTMService.GetSpanList(testCtx, reqModel, interfaces.SpanListQueryParams{})
			So(err, ShouldBeNil)
		})
	})
}

// func TestGetTrace(t *testing.T) {
// 	Convey("Test GetTrace", t, func() {
// 		mockCtrl := gomock.NewController(t)
// 		defer mockCtrl.Finish()

// 		appSetting := &common.AppSetting{}
// 		mockTMAccess := umock.NewMockTraceModelAccess(mockCtrl)
// 		mockDCAccess := umock.NewMockDataConnectionAccess(mockCtrl)
// 		mockTMAdapter := dsMock.NewMockTraceModelAdapter(mockCtrl)

// 		mockTMService := MockNewTraceModelService(appSetting, mockTMAccess, mockDCAccess)
//

// 		Convey("Get failed, caused by the error from method 'getUnderlyingDataSouceType' while searching span", func() {
// 			expectedErr := errors.New("some errors")
// 			patch := ApplyPrivateMethod(reflect.TypeOf(&traceModelService{}), "getUnderlyingDataSouceType",
// 				func(tmService *traceModelService, ctx context.Context, queryCategory string, model interfaces.TraceModel) (string, error) {
// 					return "", expectedErr
// 				})
// 			defer patch.Reset()

// 			reqModel := interfaces.TraceModel{
// 				EnabledRelatedLog: interfaces.RELATED_LOG_OPEN,
// 			}
// 			_, err := mockTMService.GetTrace(testCtx,reqModel, interfaces.TraceQueryParams{})
// 			So(err, ShouldResemble, expectedErr)
// 		})

// 		Convey("Get failed, caused by the error from method 'NewTraceModelAdapter' while searching span", func() {
// 			expectedErr := errors.New("some errors")
// 			patch1 := ApplyPrivateMethod(reflect.TypeOf(&traceModelService{}), "getUnderlyingDataSouceType",
// 				func(tmService *traceModelService, ctx context.Context, queryCategory string, model interfaces.TraceModel) (string, error) {
// 					return "", nil
// 				})
// 			defer patch1.Reset()

// 			patch2 := ApplyFuncReturn(data_source.NewTraceModelAdapter, nil, expectedErr)
// 			defer patch2.Reset()

// 			reqModel := interfaces.TraceModel{
// 				EnabledRelatedLog: interfaces.RELATED_LOG_OPEN,
// 			}
// 			_, err := mockTMService.GetTrace(testCtx,reqModel, interfaces.TraceQueryParams{})
// 			So(err, ShouldResemble, expectedErr)
// 		})

// 		Convey("Get failed, caused by the error from method 'getUnderlyingDataSouceType' while searching related_log", func() {
// 			expectedErr := errors.New("some errors")
// 			patch1 := ApplyPrivateMethod(reflect.TypeOf(&traceModelService{}), "getUnderlyingDataSouceType",
// 				func(tmService *traceModelService, ctx context.Context, queryCategory string, model interfaces.TraceModel) (string, error) {
// 					if queryCategory == interfaces.QUERY_CATEGORY_SPAN {
// 						return "", nil
// 					} else {
// 						return "", expectedErr
// 					}
// 				})
// 			defer patch1.Reset()

// 			patch2 := ApplyFuncReturn(data_source.NewTraceModelAdapter, mockTMAdapter, nil)
// 			defer patch2.Reset()

// 			mockTMAdapter.EXPECT().GetSpanMap(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().
// 				Return(map[string]*interfaces.BriefSpan_{}, map[string]interfaces.SpanDetail{}, nil)

// 			reqModel := interfaces.TraceModel{
// 				EnabledRelatedLog: interfaces.RELATED_LOG_OPEN,
// 			}
// 			_, err := mockTMService.GetTrace(testCtx,reqModel, interfaces.TraceQueryParams{})
// 			So(err, ShouldResemble, expectedErr)
// 		})

// 		Convey("Get failed, caused by the error from method 'NewTraceModelAdapter' while searching related_log", func() {
// 			expectedErr := errors.New("some errors")
// 			patch1 := ApplyPrivateMethod(reflect.TypeOf(&traceModelService{}), "getUnderlyingDataSouceType",
// 				func(tmService *traceModelService, ctx context.Context, queryCategory string, model interfaces.TraceModel) (string, error) {
// 					return "", nil
// 				})
// 			defer patch1.Reset()

// 			patch2 := ApplyFuncSeq(data_source.NewTraceModelAdapter, []OutputCell{
// 				{Values: Params{mockTMAdapter, nil}, Times: 1},
// 				{Values: Params{mockTMAdapter, expectedErr}, Times: 1},
// 			})
// 			defer patch2.Reset()

// 			mockTMAdapter.EXPECT().GetSpanMap(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().
// 				Return(map[string]*interfaces.BriefSpan_{}, map[string]interfaces.SpanDetail{}, nil)

// 			reqModel := interfaces.TraceModel{
// 				EnabledRelatedLog: interfaces.RELATED_LOG_OPEN,
// 			}
// 			_, err := mockTMService.GetTrace(testCtx,reqModel, interfaces.TraceQueryParams{})
// 			So(err, ShouldResemble, expectedErr)
// 		})

// 		Convey("Get failed, caused by the error from method 'GetRelatedLogCountMap' while searching related_log", func() {
// 			expectedErr := errors.New("some errors")
// 			patch1 := ApplyPrivateMethod(reflect.TypeOf(&traceModelService{}), "getUnderlyingDataSouceType",
// 				func(tmService *traceModelService, ctx context.Context, queryCategory string, model interfaces.TraceModel) (string, error) {
// 					return "", nil
// 				})
// 			defer patch1.Reset()

// 			patch2 := ApplyFuncReturn(data_source.NewTraceModelAdapter, mockTMAdapter, nil)
// 			defer patch2.Reset()

// 			mockTMAdapter.EXPECT().GetSpanMap(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().
// 				Return(map[string]*interfaces.BriefSpan_{}, map[string]interfaces.SpanDetail{}, nil)
// 			mockTMAdapter.EXPECT().GetRelatedLogCountMap(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().
// 				Return(map[string]int64{}, expectedErr)

// 			reqModel := interfaces.TraceModel{
// 				EnabledRelatedLog: interfaces.RELATED_LOG_OPEN,
// 			}
// 			_, err := mockTMService.GetTrace(testCtx,reqModel, interfaces.TraceQueryParams{})
// 			So(err, ShouldResemble, expectedErr)
// 		})

// 		Convey("Get failed, caused by the trace was not found", func() {
// 			patch1 := ApplyPrivateMethod(reflect.TypeOf(&traceModelService{}), "getUnderlyingDataSouceType",
// 				func(tmService *traceModelService, ctx context.Context, queryCategory string, model interfaces.TraceModel) (string, error) {
// 					return "", nil
// 				})
// 			defer patch1.Reset()

// 			patch2 := ApplyFuncReturn(data_source.NewTraceModelAdapter, mockTMAdapter, nil)
// 			defer patch2.Reset()

// 			mockTMAdapter.EXPECT().GetSpanMap(gomock.Any(), gomock.Any(), gomock.Any()).
// 				Return(map[string]*interfaces.BriefSpan_{}, map[string]interfaces.SpanDetail{}, nil)
// 			mockTMAdapter.EXPECT().GetRelatedLogCountMap(gomock.Any(), gomock.Any(), gomock.Any()).
// 				Return(map[string]int64{}, nil)

// 			reqModel := interfaces.TraceModel{
// 				EnabledRelatedLog: interfaces.RELATED_LOG_OPEN,
// 			}
// 			params := interfaces.TraceQueryParams{
// 				TraceID: "1",
// 			}
// 			_, err := mockTMService.GetTrace(testCtx,reqModel, params)
// 			So(err, ShouldResemble, rest.NewHTTPError(testCtx,http.StatusNotFound, uerrors.Uniquery_TraceModel_TraceNotFound).
// 				WithErrorDetails("The trace whose traceID equals "+params.TraceID+" was not found!"))
// 		})

// 		Convey("Get succeed", func() {
// 			patch1 := ApplyPrivateMethod(reflect.TypeOf(&traceModelService{}), "getUnderlyingDataSouceType",
// 				func(tmService *traceModelService, ctx context.Context, queryCategory string, model interfaces.TraceModel) (string, error) {
// 					return "", nil
// 				})
// 			defer patch1.Reset()

// 			patch2 := ApplyFuncReturn(data_source.NewTraceModelAdapter, mockTMAdapter, nil)
// 			defer patch2.Reset()

// 			mockTMAdapter.EXPECT().GetSpanMap(gomock.Any(), gomock.Any(), gomock.Any()).
// 				Return(map[string]*interfaces.BriefSpan_{"key1": &interfaces.BriefSpan_{}}, map[string]interfaces.SpanDetail{}, nil)
// 			mockTMAdapter.EXPECT().GetRelatedLogCountMap(gomock.Any(), gomock.Any(), gomock.Any()).
// 				Return(map[string]int64{}, nil)

// 			reqModel := interfaces.TraceModel{
// 				EnabledRelatedLog: interfaces.RELATED_LOG_OPEN,
// 			}
// 			params := interfaces.TraceQueryParams{
// 				TraceID: "1",
// 			}
// 			_, err := mockTMService.GetTrace(testCtx,reqModel, params)
// 			So(err, ShouldBeNil)
// 		})
// 	})
// }

func TestGetSpan(t *testing.T) {
	Convey("Test GetSpan", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mockTMAccess := umock.NewMockTraceModelAccess(mockCtrl)
		mockDCAccess := umock.NewMockDataConnectionAccess(mockCtrl)
		mockTMAdapter := umock.NewMockTraceModelAdapter(mockCtrl)
		psMock := umock.NewMockPermissionService(mockCtrl)

		mockTMService := MockNewTraceModelService(appSetting, mockTMAccess, mockDCAccess, psMock)

		Convey("Get failed, caused by the error from method 'getUnderlyingDataSouceType'", func() {
			expectedErr := errors.New("some errors")
			patch := ApplyPrivateMethod(reflect.TypeOf(&traceModelService{}), "getUnderlyingDataSouceType",
				func(tmService *traceModelService, ctx context.Context, queryCategory string, model interfaces.TraceModel) (string, error) {
					return "", expectedErr
				})
			defer patch.Reset()
			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			reqModel := interfaces.TraceModel{
				EnabledRelatedLog: interfaces.RELATED_LOG_OPEN,
			}
			params := interfaces.SpanQueryParams{}
			_, err := mockTMService.GetSpan(testCtx, reqModel, params)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by the error from method 'NewTraceModelAdapter'", func() {
			expectedErr := errors.New("some errors")
			patch1 := ApplyPrivateMethod(reflect.TypeOf(&traceModelService{}), "getUnderlyingDataSouceType",
				func(tmService *traceModelService, ctx context.Context, queryCategory string, model interfaces.TraceModel) (string, error) {
					return "", nil
				})
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(data_source.NewTraceModelAdapter, nil, expectedErr)
			defer patch2.Reset()
			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			reqModel := interfaces.TraceModel{
				EnabledRelatedLog: interfaces.RELATED_LOG_OPEN,
			}
			params := interfaces.SpanQueryParams{}
			_, err := mockTMService.GetSpan(testCtx, reqModel, params)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get succeed", func() {
			patch1 := ApplyPrivateMethod(reflect.TypeOf(&traceModelService{}), "getUnderlyingDataSouceType",
				func(tmService *traceModelService, ctx context.Context, queryCategory string, model interfaces.TraceModel) (string, error) {
					return "", nil
				})
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(data_source.NewTraceModelAdapter, mockTMAdapter, nil)
			defer patch2.Reset()

			mockTMAdapter.EXPECT().GetSpan(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(interfaces.SpanDetail{}, nil)
			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			reqModel := interfaces.TraceModel{
				EnabledRelatedLog: interfaces.RELATED_LOG_OPEN,
			}
			params := interfaces.SpanQueryParams{}
			_, err := mockTMService.GetSpan(testCtx, reqModel, params)
			So(err, ShouldBeNil)
		})
	})
}

func TestGetSpanRelatedLogList(t *testing.T) {
	Convey("Test GetSpanRelatedLogList", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mockTMAccess := umock.NewMockTraceModelAccess(mockCtrl)
		mockDCAccess := umock.NewMockDataConnectionAccess(mockCtrl)
		mockTMAdapter := umock.NewMockTraceModelAdapter(mockCtrl)
		psMock := umock.NewMockPermissionService(mockCtrl)

		mockTMService := MockNewTraceModelService(appSetting, mockTMAccess, mockDCAccess, psMock)

		Convey("Get succeed with empty result, caused by the enabled_related_log is 0", func() {
			reqModel := interfaces.TraceModel{
				EnabledRelatedLog: interfaces.RELATED_LOG_CLOSE,
			}
			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			_, total, err := mockTMService.GetSpanRelatedLogList(testCtx, reqModel, interfaces.RelatedLogListQueryParams{})
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 0)
		})

		Convey("Get failed, caused by the error from method 'getUnderlyingDataSouceType'", func() {
			expectedErr := errors.New("some errors")
			patch := ApplyPrivateMethod(reflect.TypeOf(&traceModelService{}), "getUnderlyingDataSouceType",
				func(tmService *traceModelService, ctx context.Context, queryCategory string, model interfaces.TraceModel) (string, error) {
					return "", expectedErr
				})
			defer patch.Reset()
			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			reqModel := interfaces.TraceModel{
				EnabledRelatedLog: interfaces.RELATED_LOG_OPEN,
			}
			_, _, err := mockTMService.GetSpanRelatedLogList(testCtx, reqModel, interfaces.RelatedLogListQueryParams{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by the error from method 'NewTraceModelAdapter'", func() {
			expectedErr := errors.New("some errors")
			patch1 := ApplyPrivateMethod(reflect.TypeOf(&traceModelService{}), "getUnderlyingDataSouceType",
				func(tmService *traceModelService, ctx context.Context, queryCategory string, model interfaces.TraceModel) (string, error) {
					return "", nil
				})
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(data_source.NewTraceModelAdapter, nil, expectedErr)
			defer patch2.Reset()
			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			reqModel := interfaces.TraceModel{
				EnabledRelatedLog: interfaces.RELATED_LOG_OPEN,
			}
			_, _, err := mockTMService.GetSpanRelatedLogList(testCtx, reqModel, interfaces.RelatedLogListQueryParams{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get succeed", func() {
			patch1 := ApplyPrivateMethod(reflect.TypeOf(&traceModelService{}), "getUnderlyingDataSouceType",
				func(tmService *traceModelService, ctx context.Context, queryCategory string, model interfaces.TraceModel) (string, error) {
					return "", nil
				})
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(data_source.NewTraceModelAdapter, mockTMAdapter, nil)
			defer patch2.Reset()

			mockTMAdapter.EXPECT().GetSpanRelatedLogList(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, int64(0), nil)
			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			reqModel := interfaces.TraceModel{
				EnabledRelatedLog: interfaces.RELATED_LOG_OPEN,
			}
			_, _, err := mockTMService.GetSpanRelatedLogList(testCtx, reqModel, interfaces.RelatedLogListQueryParams{})
			So(err, ShouldBeNil)
		})
	})
}

func TestGetTraceModelByID(t *testing.T) {
	Convey("Test GetTraceModelByID", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mockTMAccess := umock.NewMockTraceModelAccess(mockCtrl)
		mockDCAccess := umock.NewMockDataConnectionAccess(mockCtrl)
		psMock := umock.NewMockPermissionService(mockCtrl)

		mockTMService := MockNewTraceModelService(appSetting, mockTMAccess, mockDCAccess, psMock)

		Convey("Get failed, caused by the error from method 'GetTraceModelByID'", func() {
			expectedErr := errors.New("some errors")
			mockTMAccess.EXPECT().GetTraceModelByID(gomock.Any(), gomock.Any()).Return(interfaces.TraceModel{}, false, expectedErr)

			_, err := mockTMService.GetTraceModelByID(testCtx, "1")
			So(err, ShouldResemble, rest.NewHTTPError(testCtx, http.StatusInternalServerError, uerrors.Uniquery_TraceModel_InternalError_GetTraceModelByIDFailed).
				WithErrorDetails(expectedErr.Error()))
		})

		Convey("Get failed, caused by the trace model was not found", func() {
			reqModelID := "1"
			expectedErrorDetails := fmt.Sprintf("The trace model whose id equal to %v was not found", reqModelID)

			mockTMAccess.EXPECT().GetTraceModelByID(gomock.Any(), gomock.Any()).Return(interfaces.TraceModel{}, false, nil)

			_, err := mockTMService.GetTraceModelByID(testCtx, reqModelID)
			So(err, ShouldResemble, rest.NewHTTPError(testCtx, http.StatusNotFound, uerrors.Uniquery_TraceModel_TraceModelNotFound).
				WithErrorDetails(expectedErrorDetails))
		})

		Convey("Get succeed", func() {
			reqModelID := "1"
			mockTMAccess.EXPECT().GetTraceModelByID(gomock.Any(), gomock.Any()).Return(interfaces.TraceModel{}, true, nil)

			_, err := mockTMService.GetTraceModelByID(testCtx, reqModelID)
			So(err, ShouldBeNil)
		})
	})
}

func TestSimulateCreateTraceModel(t *testing.T) {
	Convey("Test SimulateCreateTraceModel", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		ops := []interfaces.ResourceOps{
			{
				ResourceID: interfaces.RESOURCE_ID_ALL,
				Operations: []string{interfaces.OPERATION_TYPE_CREATE},
			},
		}

		appSetting := &common.AppSetting{}
		mockTMAccess := umock.NewMockTraceModelAccess(mockCtrl)
		mockDCAccess := umock.NewMockDataConnectionAccess(mockCtrl)
		psMock := umock.NewMockPermissionService(mockCtrl)

		mockTMService := MockNewTraceModelService(appSetting, mockTMAccess, mockDCAccess, psMock)

		Convey("Simulate failed, caused by the error from method 'SimulateCreateTraceModel'", func() {
			expectedErr := errors.New("some errors")
			mockTMAccess.EXPECT().SimulateCreateTraceModel(gomock.Any(), gomock.Any()).Return(interfaces.TraceModel{}, expectedErr)
			psMock.EXPECT().GetResourcesOperations(gomock.Any(), gomock.Any(), gomock.Any()).Return(ops, nil)

			_, err := mockTMService.SimulateCreateTraceModel(testCtx, interfaces.TraceModel{})
			So(err, ShouldResemble, rest.NewHTTPError(testCtx, http.StatusInternalServerError, uerrors.Uniquery_TraceModel_InternalError_SimulateCreateTraceModelFailed).
				WithErrorDetails(expectedErr.Error()))
		})

		Convey("Simulate succeed", func() {
			mockTMAccess.EXPECT().SimulateCreateTraceModel(gomock.Any(), gomock.Any()).Return(interfaces.TraceModel{}, nil)
			psMock.EXPECT().GetResourcesOperations(gomock.Any(), gomock.Any(), gomock.Any()).Return(ops, nil)

			_, err := mockTMService.SimulateCreateTraceModel(testCtx, interfaces.TraceModel{})
			So(err, ShouldBeNil)
		})
	})
}

func TestSimulateUpdateTraceModel(t *testing.T) {
	Convey("Test SimulateUpdateTraceModel", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		ops := []interfaces.ResourceOps{
			{
				ResourceID: interfaces.RESOURCE_ID_ALL,
				Operations: []string{interfaces.OPERATION_TYPE_CREATE},
			},
		}

		appSetting := &common.AppSetting{}
		mockTMAccess := umock.NewMockTraceModelAccess(mockCtrl)
		mockDCAccess := umock.NewMockDataConnectionAccess(mockCtrl)
		psMock := umock.NewMockPermissionService(mockCtrl)

		mockTMService := MockNewTraceModelService(appSetting, mockTMAccess, mockDCAccess, psMock)

		Convey("Simulate failed, caused by the error from method 'SimulateUpdateTraceModel'", func() {
			expectedErr := errors.New("some errors")
			mockTMAccess.EXPECT().SimulateUpdateTraceModel(gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.TraceModel{}, expectedErr)
			psMock.EXPECT().GetResourcesOperations(gomock.Any(), gomock.Any(), gomock.Any()).Return(ops, nil)

			_, err := mockTMService.SimulateUpdateTraceModel(testCtx, "1", interfaces.TraceModel{})
			So(err, ShouldResemble, rest.NewHTTPError(testCtx, http.StatusInternalServerError, uerrors.Uniquery_TraceModel_InternalError_SimulateUpdateTraceModelFailed).
				WithErrorDetails(expectedErr.Error()))
		})

		Convey("Simulate succeed", func() {
			mockTMAccess.EXPECT().SimulateUpdateTraceModel(gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.TraceModel{}, nil)
			psMock.EXPECT().GetResourcesOperations(gomock.Any(), gomock.Any(), gomock.Any()).Return(ops, nil)

			_, err := mockTMService.SimulateUpdateTraceModel(testCtx, "1", interfaces.TraceModel{})
			So(err, ShouldBeNil)
		})
	})
}

/*
	私有方法
*/

func TestGetUnderlyingDataSouceType(t *testing.T) {
	Convey("Test getUnderlyingDataSouceType", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mockTMAccess := umock.NewMockTraceModelAccess(mockCtrl)
		mockDCAccess := umock.NewMockDataConnectionAccess(mockCtrl)
		psMock := umock.NewMockPermissionService(mockCtrl)

		mockTMService := MockNewTraceModelService(appSetting, mockTMAccess, mockDCAccess, psMock)

		Convey("queryCategory is interfaces.QUERY_CATEGORY_SPAN and model.SpanSourceType is interfaces.SOURCE_TYPE_DATA_VIEW", func() {
			sourceType, err := mockTMService.getUnderlyingDataSouceType(testCtx, interfaces.QUERY_CATEGORY_SPAN, interfaces.TraceModel{
				SpanSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
			})
			So(err, ShouldBeNil)
			So(sourceType, ShouldEqual, interfaces.SOURCE_TYPE_DATA_VIEW)
		})

		Convey("queryCategory is interfaces.QUERY_CATEGORY_RELATED_LOG and model.RelatedLogSourceType is interfaces.SOURCE_TYPE_DATA_VIEW", func() {
			sourceType, err := mockTMService.getUnderlyingDataSouceType(testCtx, interfaces.QUERY_CATEGORY_RELATED_LOG, interfaces.TraceModel{
				RelatedLogSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
			})
			So(err, ShouldBeNil)
			So(sourceType, ShouldEqual, interfaces.SOURCE_TYPE_DATA_VIEW)
		})

		Convey("other condition, and method 'GetDataConnectionTypeByName' return error", func() {
			expectedErr := errors.New("some errors")
			mockDCAccess.EXPECT().GetDataConnectionTypeByName(gomock.Any(), gomock.Any()).Return("", false, expectedErr)

			_, err := mockTMService.getUnderlyingDataSouceType(testCtx, interfaces.QUERY_CATEGORY_SPAN, interfaces.TraceModel{
				RelatedLogSourceType: interfaces.SOURCE_TYPE_DATA_CONNECTION,
				SpanConfig:           interfaces.SpanConfigWithDataConnection{},
			})
			So(err, ShouldResemble, rest.NewHTTPError(testCtx, http.StatusInternalServerError, uerrors.Uniquery_TraceModel_InternalError_GetUnderlyingDataSouceTypeFailed).
				WithErrorDetails(expectedErr.Error()))
		})

		Convey("other condition, and data connection was not found", func() {
			expectedErr := fmt.Errorf("Get underlying data souce type failed, the data connection whose name equal to %s was not found", "conn1")
			mockDCAccess.EXPECT().GetDataConnectionTypeByName(gomock.Any(), gomock.Any()).Return("", false, nil)

			_, err := mockTMService.getUnderlyingDataSouceType(testCtx, interfaces.QUERY_CATEGORY_SPAN, interfaces.TraceModel{
				RelatedLogSourceType: interfaces.SOURCE_TYPE_DATA_CONNECTION,
				SpanConfig: interfaces.SpanConfigWithDataConnection{
					DataConnection: interfaces.DataConnectionConfig{
						Name: "conn1",
					},
				},
			})
			So(err, ShouldResemble, rest.NewHTTPError(testCtx, http.StatusInternalServerError, uerrors.Uniquery_TraceModel_InternalError_GetUnderlyingDataSouceTypeFailed).
				WithErrorDetails(expectedErr.Error()))
		})

		Convey("other condition, and return correct result", func() {
			mockDCAccess.EXPECT().GetDataConnectionTypeByName(gomock.Any(), gomock.Any()).Return(interfaces.SOURCE_TYPE_TINGYUN, true, nil)

			sourceType, err := mockTMService.getUnderlyingDataSouceType(testCtx, interfaces.QUERY_CATEGORY_SPAN, interfaces.TraceModel{
				RelatedLogSourceType: interfaces.SOURCE_TYPE_DATA_CONNECTION,
				SpanConfig: interfaces.SpanConfigWithDataConnection{
					DataConnection: interfaces.DataConnectionConfig{
						Name: "conn1",
					},
				},
			})
			So(err, ShouldBeNil)
			So(sourceType, ShouldEqual, interfaces.SOURCE_TYPE_TINGYUN)
		})
	})
}

func TestBuildTree(t *testing.T) {
	Convey("Test buildTree", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mockTMAccess := umock.NewMockTraceModelAccess(mockCtrl)
		mockDCAccess := umock.NewMockDataConnectionAccess(mockCtrl)
		psMock := umock.NewMockPermissionService(mockCtrl)

		mockTMService := MockNewTraceModelService(appSetting, mockTMAccess, mockDCAccess, psMock)

		Convey("build succeed", func() {
			spanMap := map[string]*interfaces.BriefSpan_{
				"key_1": {
					SpanID:       "key_1",
					ParentSpanID: "key_0",
				},
				"key_0": {
					SpanID:       "key_0",
					ParentSpanID: "key_-1",
				},
			}

			rootSpan := mockTMService.buildTree(testCtx, spanMap)
			So(rootSpan, ShouldResemble, &interfaces.BriefSpan_{
				SpanID:       "key_0",
				ParentSpanID: "key_-1",
				Children: []*interfaces.BriefSpan_{
					{
						SpanID:       "key_1",
						ParentSpanID: "key_0",
					},
				},
			})
		})
	})
}

func TestGetTraceStatisticData(t *testing.T) {
	Convey("Test getTraceStatisticData", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mockTMAccess := umock.NewMockTraceModelAccess(mockCtrl)
		mockDCAccess := umock.NewMockDataConnectionAccess(mockCtrl)
		psMock := umock.NewMockPermissionService(mockCtrl)

		mockTMService := MockNewTraceModelService(appSetting, mockTMAccess, mockDCAccess, psMock)

		Convey("rootSpan is null", func() {
			res := mockTMService.getTraceStatisticData(testCtx, nil)
			So(res, ShouldResemble, interfaces.TraceStatisticData{
				StartTime: math.MaxInt64,
				EndTime:   math.MinInt64,
				StatusStats: map[string]int64{
					interfaces.SPAN_STATUS_UNSET: 0,
					interfaces.SPAN_STATUS_OK:    0,
					interfaces.SPAN_STATUS_ERROR: 0,
				},
			})
		})

		Convey("normal rootSpan", func() {
			rootSpan := &interfaces.BriefSpan_{
				Status: interfaces.SPAN_STATUS_UNSET,
			}

			res := mockTMService.getTraceStatisticData(testCtx, rootSpan)
			So(res, ShouldResemble, interfaces.TraceStatisticData{
				StartTime: int64(0),
				EndTime:   int64(0),
				StatusStats: map[string]int64{
					interfaces.SPAN_STATUS_UNSET: int64(1),
					interfaces.SPAN_STATUS_OK:    int64(0),
					interfaces.SPAN_STATUS_ERROR: int64(0),
				},
				Depth:    int64(1),
				Services: []string{""},
			})
		})
	})
}
