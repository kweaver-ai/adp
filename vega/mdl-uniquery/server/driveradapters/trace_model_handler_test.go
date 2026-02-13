// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	rmock "github.com/kweaver-ai/kweaver-go-lib/rest/mock"
	. "github.com/smartystreets/goconvey/convey"

	"uniquery/common/convert"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	umock "uniquery/interfaces/mock"
)

func mockNewTraceModelRestHandler(hydra rest.Hydra, tmService interfaces.TraceModelService) (r *restHandler) {
	r = &restHandler{
		hydra:     hydra,
		tmService: tmService,
	}
	r.InitMetric()
	return r
}

func TestPreviewSpanList(t *testing.T) {
	Convey("Test handler PreviewSpanList", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		hydraMock := rmock.NewMockHydra(mockCtrl)
		tmServiceMock := umock.NewMockTraceModelService(mockCtrl)
		handler := mockNewTraceModelRestHandler(hydraMock, tmServiceMock)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/simulate-traces/1/spans"

		Convey("Preview failed, caused by invalid X_HTTP_METHOD_OVERRIDE", func() {
			req := httptest.NewRequest(http.MethodPost, url, nil)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_InvalidParameter_OverrideMethod)
		})

		Convey("Preview failed, caused by the error from method ShouldBindJSON", func() {
			expectedErr := errors.New("some errors")
			queryParams := interfaces.SpanListPreviewParams{}

			patch := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", expectedErr)
			defer patch.Reset()

			reqParamByte, _ := json.Marshal(queryParams)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_InvalidParameter_RequestBody)
		})

		Convey("Preview failed, caused by the error from func validateParamsWhenGetSpanList", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_Offset).
				WithErrorDetails(fmt.Sprintf("The offset is not greater than %d", interfaces.MIN_OFFSET_NUM_OF_LIST))
			queryParams := interfaces.SpanListPreviewParams{}

			patch1 := ApplyMethod(reflect.TypeOf(&gin.Context{}), "ShouldBindJSON",
				func(_ *gin.Context, obj any) error {
					o, _ := obj.(*interfaces.SpanListPreviewParams)
					*o = queryParams
					return nil
				})
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateParamsWhenGetSpanList, interfaces.SpanListQueryParams{}, expectedErr)
			defer patch2.Reset()

			reqParamByte, _ := json.Marshal(queryParams)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_InvalidParameter_Offset)
		})

		Convey("Preview failed, caused by the error from method SimulateCreateTraceModel", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, uerrors.Uniquery_TraceModel_InternalError_SimulateCreateTraceModelFailed)
			queryParams := interfaces.SpanListPreviewParams{}

			patch1 := ApplyMethod(reflect.TypeOf(&gin.Context{}), "ShouldBindJSON",
				func(_ *gin.Context, obj any) error {
					o, _ := obj.(*interfaces.SpanListPreviewParams)
					*o = queryParams
					return nil
				})
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateParamsWhenGetSpanList, interfaces.SpanListQueryParams{}, nil)
			defer patch2.Reset()

			tmServiceMock.EXPECT().SimulateCreateTraceModel(gomock.Any(), gomock.Any()).Return(interfaces.TraceModel{}, expectedErr)

			reqParamByte, _ := json.Marshal(queryParams)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_TraceModel_InternalError_SimulateCreateTraceModelFailed)
		})

		Convey("Preview failed, caused by the error from method SimulateUpdateTraceModel", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, uerrors.Uniquery_TraceModel_InternalError_SimulateUpdateTraceModelFailed)
			queryParams := interfaces.SpanListPreviewParams{
				TraceModel: interfaces.TraceModel{
					ID: "1",
				},
			}

			patch1 := ApplyMethod(reflect.TypeOf(&gin.Context{}), "ShouldBindJSON",
				func(_ *gin.Context, obj any) error {
					o, _ := obj.(*interfaces.SpanListPreviewParams)
					*o = queryParams
					return nil
				})
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateParamsWhenGetSpanList, interfaces.SpanListQueryParams{}, nil)
			defer patch2.Reset()

			tmServiceMock.EXPECT().SimulateUpdateTraceModel(gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.TraceModel{}, expectedErr)

			reqParamByte, _ := json.Marshal(queryParams)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_TraceModel_InternalError_SimulateUpdateTraceModelFailed)
		})

		Convey("Preview failed, caused by the error from method GetSpanList", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, uerrors.Uniquery_TraceModel_InternalError_GetTingYunTraceListFailed)
			queryParams := interfaces.SpanListPreviewParams{
				TraceModel: interfaces.TraceModel{
					ID: "1",
				},
			}

			patch1 := ApplyMethod(reflect.TypeOf(&gin.Context{}), "ShouldBindJSON",
				func(_ *gin.Context, obj any) error {
					o, _ := obj.(*interfaces.SpanListPreviewParams)
					*o = queryParams
					return nil
				})
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateParamsWhenGetSpanList, interfaces.SpanListQueryParams{}, nil)
			defer patch2.Reset()

			tmServiceMock.EXPECT().SimulateUpdateTraceModel(gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.TraceModel{}, nil)
			tmServiceMock.EXPECT().GetSpanList(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, int64(0), expectedErr)

			reqParamByte, _ := json.Marshal(queryParams)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_TraceModel_InternalError_GetTingYunTraceListFailed)
		})

		Convey("Preview succeed", func() {
			queryParams := interfaces.SpanListPreviewParams{
				TraceModel: interfaces.TraceModel{
					ID: "1",
				},
			}

			patch1 := ApplyMethod(reflect.TypeOf(&gin.Context{}), "ShouldBindJSON",
				func(_ *gin.Context, obj any) error {
					o, _ := obj.(*interfaces.SpanListPreviewParams)
					*o = queryParams
					return nil
				})
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateParamsWhenGetSpanList, interfaces.SpanListQueryParams{}, nil)
			defer patch2.Reset()

			tmServiceMock.EXPECT().SimulateUpdateTraceModel(gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.TraceModel{}, nil)
			tmServiceMock.EXPECT().GetSpanList(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, int64(0), nil)

			reqParamByte, _ := json.Marshal(queryParams)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func TestGetSpanList(t *testing.T) {
	Convey("Test handler GetSpanList", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		hydraMock := rmock.NewMockHydra(mockCtrl)
		tmServiceMock := umock.NewMockTraceModelService(mockCtrl)
		handler := mockNewTraceModelRestHandler(hydraMock, tmServiceMock)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/trace-models/1/traces/1/spans"

		Convey("Get failed, caused by invalid X_HTTP_METHOD_OVERRIDE", func() {
			req := httptest.NewRequest(http.MethodPost, url, nil)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_InvalidParameter_OverrideMethod)
		})

		Convey("Get failed, caused by the error from method ShouldBindJSON", func() {
			expectedErr := errors.New("some errors")
			queryParams := interfaces.SpanListQueryParams{}

			patch := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", expectedErr)
			defer patch.Reset()

			reqParamByte, _ := json.Marshal(queryParams)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_InvalidParameter_RequestBody)
		})

		Convey("Get failed, caused by the error from func validateParamsWhenGetSpanList", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_Offset).
				WithErrorDetails(fmt.Sprintf("The offset is not greater than %d", interfaces.MIN_OFFSET_NUM_OF_LIST))
			queryParams := interfaces.SpanListQueryParams{}

			patch1 := ApplyMethod(reflect.TypeOf(&gin.Context{}), "ShouldBindJSON",
				func(_ *gin.Context, obj any) error {
					o, _ := obj.(*interfaces.SpanListQueryParams)
					*o = queryParams
					return nil
				})
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateParamsWhenGetSpanList, interfaces.SpanListQueryParams{}, expectedErr)
			defer patch2.Reset()

			reqParamByte, _ := json.Marshal(queryParams)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_InvalidParameter_Offset)
		})

		Convey("Get failed, caused by the error from method GetTraceModelByID", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, uerrors.Uniquery_TraceModel_InternalError_GetTraceModelByIDFailed)
			queryParams := interfaces.SpanListQueryParams{}

			patch1 := ApplyMethod(reflect.TypeOf(&gin.Context{}), "ShouldBindJSON",
				func(_ *gin.Context, obj any) error {
					o, _ := obj.(*interfaces.SpanListQueryParams)
					*o = queryParams
					return nil
				})
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateParamsWhenGetSpanList, interfaces.SpanListQueryParams{}, nil)
			defer patch2.Reset()

			tmServiceMock.EXPECT().GetTraceModelByID(gomock.Any(), gomock.Any()).Return(interfaces.TraceModel{}, expectedErr)

			reqParamByte, _ := json.Marshal(queryParams)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_TraceModel_InternalError_GetTraceModelByIDFailed)
		})

		Convey("Get failed, caused by the error from method GetSpanList", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, uerrors.Uniquery_TraceModel_InternalError_GetTingYunTraceListFailed)
			queryParams := interfaces.SpanListQueryParams{}

			patch1 := ApplyMethod(reflect.TypeOf(&gin.Context{}), "ShouldBindJSON",
				func(_ *gin.Context, obj any) error {
					o, _ := obj.(*interfaces.SpanListQueryParams)
					*o = queryParams
					return nil
				})
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateParamsWhenGetSpanList, interfaces.SpanListQueryParams{}, nil)
			defer patch2.Reset()

			tmServiceMock.EXPECT().GetTraceModelByID(gomock.Any(), gomock.Any()).Return(interfaces.TraceModel{}, nil)
			tmServiceMock.EXPECT().GetSpanList(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, int64(0), expectedErr)

			reqParamByte, _ := json.Marshal(queryParams)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_TraceModel_InternalError_GetTingYunTraceListFailed)
		})

		Convey("Get succeed", func() {
			queryParams := interfaces.SpanListQueryParams{}

			patch1 := ApplyMethod(reflect.TypeOf(&gin.Context{}), "ShouldBindJSON",
				func(_ *gin.Context, obj any) error {
					o, _ := obj.(*interfaces.SpanListQueryParams)
					*o = queryParams
					return nil
				})
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateParamsWhenGetSpanList, interfaces.SpanListQueryParams{}, nil)
			defer patch2.Reset()

			tmServiceMock.EXPECT().GetTraceModelByID(gomock.Any(), gomock.Any()).Return(interfaces.TraceModel{}, nil)
			tmServiceMock.EXPECT().GetSpanList(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, int64(0), nil)

			reqParamByte, _ := json.Marshal(queryParams)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func TestPreviewTrace(t *testing.T) {
	Convey("Test handler PreviewTrace", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		hydraMock := rmock.NewMockHydra(mockCtrl)
		tmServiceMock := umock.NewMockTraceModelService(mockCtrl)
		handler := mockNewTraceModelRestHandler(hydraMock, tmServiceMock)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/simulate-traces/1"

		Convey("Preview failed, caused by invalid X_HTTP_METHOD_OVERRIDE", func() {
			req := httptest.NewRequest(http.MethodPost, url, nil)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_InvalidParameter_OverrideMethod)
		})

		Convey("Preview failed, caused by the error from method ShouldBindJSON", func() {
			expectedErr := errors.New("some errors")
			queryParams := interfaces.TracePreviewParams{}

			patch := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", expectedErr)
			defer patch.Reset()

			reqParamByte, _ := json.Marshal(queryParams)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_InvalidParameter_RequestBody)
		})

		Convey("Preview failed, caused by the error from method SimulateCreateTraceModel", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, uerrors.Uniquery_TraceModel_InternalError_SimulateCreateTraceModelFailed)
			queryParams := interfaces.TracePreviewParams{}

			patch1 := ApplyMethod(reflect.TypeOf(&gin.Context{}), "ShouldBindJSON",
				func(_ *gin.Context, obj any) error {
					o, _ := obj.(*interfaces.TracePreviewParams)
					*o = queryParams
					return nil
				})
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateParamsWhenGetSpanList, interfaces.SpanListQueryParams{}, nil)
			defer patch2.Reset()

			tmServiceMock.EXPECT().SimulateCreateTraceModel(gomock.Any(), gomock.Any()).Return(interfaces.TraceModel{}, expectedErr)

			reqParamByte, _ := json.Marshal(queryParams)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_TraceModel_InternalError_SimulateCreateTraceModelFailed)
		})

		Convey("Preview failed, caused by the error from method SimulateUpdateTraceModel", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, uerrors.Uniquery_TraceModel_InternalError_SimulateUpdateTraceModelFailed)
			queryParams := interfaces.TracePreviewParams{
				TraceModel: interfaces.TraceModel{
					ID: "1",
				},
			}

			patch := ApplyMethod(reflect.TypeOf(&gin.Context{}), "ShouldBindJSON",
				func(_ *gin.Context, obj any) error {
					o, _ := obj.(*interfaces.TracePreviewParams)
					*o = queryParams
					return nil
				})
			defer patch.Reset()

			tmServiceMock.EXPECT().SimulateUpdateTraceModel(gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.TraceModel{}, expectedErr)

			reqParamByte, _ := json.Marshal(queryParams)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_TraceModel_InternalError_SimulateUpdateTraceModelFailed)
		})

		Convey("Preview failed, caused by the error from method GetTrace", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, uerrors.Uniquery_TraceModel_InternalError_GetTingYunTraceDetailFailed)
			queryParams := interfaces.TracePreviewParams{
				TraceModel: interfaces.TraceModel{
					ID: "1",
				},
			}

			patch := ApplyMethod(reflect.TypeOf(&gin.Context{}), "ShouldBindJSON",
				func(_ *gin.Context, obj any) error {
					o, _ := obj.(*interfaces.TracePreviewParams)
					*o = queryParams
					return nil
				})
			defer patch.Reset()

			tmServiceMock.EXPECT().SimulateUpdateTraceModel(gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.TraceModel{}, nil)
			tmServiceMock.EXPECT().GetTrace(gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.TraceDetail_{}, expectedErr)

			reqParamByte, _ := json.Marshal(queryParams)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_TraceModel_InternalError_GetTingYunTraceDetailFailed)
		})

		Convey("Preview succeed", func() {
			queryParams := interfaces.TracePreviewParams{
				TraceModel: interfaces.TraceModel{
					ID: "1",
				},
			}

			patch1 := ApplyMethod(reflect.TypeOf(&gin.Context{}), "ShouldBindJSON",
				func(_ *gin.Context, obj any) error {
					o, _ := obj.(*interfaces.TracePreviewParams)
					*o = queryParams
					return nil
				})
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateParamsWhenGetSpanList, interfaces.SpanListQueryParams{}, nil)
			defer patch2.Reset()

			tmServiceMock.EXPECT().SimulateUpdateTraceModel(gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.TraceModel{}, nil)
			tmServiceMock.EXPECT().GetTrace(gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.TraceDetail_{}, nil)

			reqParamByte, _ := json.Marshal(queryParams)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func TestGetTrace(t *testing.T) {
	Convey("Test handler GetTrace", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		hydraMock := rmock.NewMockHydra(mockCtrl)
		tmServiceMock := umock.NewMockTraceModelService(mockCtrl)
		handler := mockNewTraceModelRestHandler(hydraMock, tmServiceMock)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/trace-models/1/traces/1"

		Convey("Get failed, caused by invalid X_HTTP_METHOD_OVERRIDE", func() {
			req := httptest.NewRequest(http.MethodPost, url, nil)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_InvalidParameter_OverrideMethod)
		})

		Convey("Get failed, caused by the error from method ShouldBindJSON", func() {
			expectedErr := errors.New("some errors")
			queryParams := interfaces.TracePreviewParams{}

			patch2 := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", expectedErr)
			defer patch2.Reset()

			reqParamByte, _ := json.Marshal(queryParams)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_InvalidParameter_RequestBody)
		})

		Convey("Get failed, caused by the error from method GetTraceModelByID", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusNotFound, uerrors.Uniquery_TraceModel_TraceModelNotFound)
			queryParams := interfaces.TracePreviewParams{
				TraceModel: interfaces.TraceModel{
					ID: "1",
				},
			}

			patch2 := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", nil)
			defer patch2.Reset()

			tmServiceMock.EXPECT().GetTraceModelByID(gomock.Any(), gomock.Any()).Return(interfaces.TraceModel{}, expectedErr)

			reqParamByte, _ := json.Marshal(queryParams)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_TraceModel_TraceModelNotFound)
		})

		Convey("Get failed, caused by the error from method GetTrace", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusNotFound, uerrors.Uniquery_TraceModel_TraceModelNotFound)
			queryParams := interfaces.TracePreviewParams{
				TraceModel: interfaces.TraceModel{
					ID: "1",
				},
			}

			patch2 := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", nil)
			defer patch2.Reset()

			tmServiceMock.EXPECT().GetTraceModelByID(gomock.Any(), gomock.Any()).Return(interfaces.TraceModel{}, nil)
			tmServiceMock.EXPECT().GetTrace(gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.TraceDetail_{}, expectedErr)

			reqParamByte, _ := json.Marshal(queryParams)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_TraceModel_TraceModelNotFound)
		})

		Convey("Get succeed", func() {
			queryParams := interfaces.TracePreviewParams{
				TraceModel: interfaces.TraceModel{
					ID: "1",
				},
			}

			patch2 := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", nil)
			defer patch2.Reset()

			tmServiceMock.EXPECT().GetTraceModelByID(gomock.Any(), gomock.Any()).Return(interfaces.TraceModel{}, nil)
			tmServiceMock.EXPECT().GetTrace(gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.TraceDetail_{}, nil)

			reqParamByte, _ := json.Marshal(queryParams)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func TestPreviewSpan(t *testing.T) {
	Convey("Test handler PreviewSpan", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		hydraMock := rmock.NewMockHydra(mockCtrl)
		tmServiceMock := umock.NewMockTraceModelService(mockCtrl)
		handler := mockNewTraceModelRestHandler(hydraMock, tmServiceMock)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/simulate-traces/1/spans/1"

		Convey("Preview failed, caused by invalid X_HTTP_METHOD_OVERRIDE", func() {
			req := httptest.NewRequest(http.MethodPost, url, nil)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_InvalidParameter_OverrideMethod)
		})

		Convey("Preview failed, caused by the error from method ShouldBindJSON", func() {
			expectedErr := errors.New("some errors")
			queryParams := interfaces.SpanPreviewParams{}

			patch := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", expectedErr)
			defer patch.Reset()

			reqParamByte, _ := json.Marshal(queryParams)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_InvalidParameter_RequestBody)
		})

		Convey("Preview failed, caused by the error from method SimulateCreateTraceModel", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, uerrors.Uniquery_TraceModel_InternalError_SimulateCreateTraceModelFailed)
			queryParams := interfaces.SpanPreviewParams{}

			patch := ApplyMethod(reflect.TypeOf(&gin.Context{}), "ShouldBindJSON",
				func(_ *gin.Context, obj any) error {
					o, _ := obj.(*interfaces.SpanPreviewParams)
					*o = queryParams
					return nil
				})
			defer patch.Reset()

			tmServiceMock.EXPECT().SimulateCreateTraceModel(gomock.Any(), gomock.Any()).Return(interfaces.TraceModel{}, expectedErr)

			reqParamByte, _ := json.Marshal(queryParams)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_TraceModel_InternalError_SimulateCreateTraceModelFailed)
		})

		Convey("Preview failed, caused by the error from method SimulateUpdateTraceModel", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, uerrors.Uniquery_TraceModel_InternalError_SimulateUpdateTraceModelFailed)
			queryParams := interfaces.SpanPreviewParams{
				TraceModel: interfaces.TraceModel{
					ID: "1",
				},
			}

			patch := ApplyMethod(reflect.TypeOf(&gin.Context{}), "ShouldBindJSON",
				func(_ *gin.Context, obj any) error {
					o, _ := obj.(*interfaces.SpanPreviewParams)
					*o = queryParams
					return nil
				})
			defer patch.Reset()

			tmServiceMock.EXPECT().SimulateUpdateTraceModel(gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.TraceModel{}, expectedErr)

			reqParamByte, _ := json.Marshal(queryParams)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_TraceModel_InternalError_SimulateUpdateTraceModelFailed)
		})

		Convey("Preview failed, caused by the error from method GetSpan", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, uerrors.Uniquery_InternalError)
			queryParams := interfaces.SpanPreviewParams{
				TraceModel: interfaces.TraceModel{
					ID: "1",
				},
			}

			patch := ApplyMethod(reflect.TypeOf(&gin.Context{}), "ShouldBindJSON",
				func(_ *gin.Context, obj any) error {
					o, _ := obj.(*interfaces.SpanPreviewParams)
					*o = queryParams
					return nil
				})
			defer patch.Reset()

			tmServiceMock.EXPECT().SimulateUpdateTraceModel(gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.TraceModel{}, nil)
			tmServiceMock.EXPECT().GetSpan(gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.SpanDetail{}, expectedErr)

			reqParamByte, _ := json.Marshal(queryParams)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_InternalError)
		})

		Convey("Preview succeed", func() {
			queryParams := interfaces.SpanPreviewParams{
				TraceModel: interfaces.TraceModel{
					ID: "1",
				},
			}

			patch := ApplyMethod(reflect.TypeOf(&gin.Context{}), "ShouldBindJSON",
				func(_ *gin.Context, obj any) error {
					o, _ := obj.(*interfaces.SpanPreviewParams)
					*o = queryParams
					return nil
				})
			defer patch.Reset()

			tmServiceMock.EXPECT().SimulateUpdateTraceModel(gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.TraceModel{}, nil)
			tmServiceMock.EXPECT().GetSpan(gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.SpanDetail{}, nil)

			reqParamByte, _ := json.Marshal(queryParams)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func TestGetSpan(t *testing.T) {
	Convey("Test handler GetSpan", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		hydraMock := rmock.NewMockHydra(mockCtrl)
		tmServiceMock := umock.NewMockTraceModelService(mockCtrl)
		handler := mockNewTraceModelRestHandler(hydraMock, tmServiceMock)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/trace-models/1/traces/1/spans/1"

		Convey("Get failed, caused by the error from method GetTraceModelByID", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusNotFound, uerrors.Uniquery_TraceModel_TraceModelNotFound)

			tmServiceMock.EXPECT().GetTraceModelByID(gomock.Any(), gomock.Any()).Return(interfaces.TraceModel{}, expectedErr)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_TraceModel_TraceModelNotFound)
		})

		Convey("Get failed, caused by the error from method GetSpan", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, uerrors.Uniquery_InternalError)

			tmServiceMock.EXPECT().GetTraceModelByID(gomock.Any(), gomock.Any()).Return(interfaces.TraceModel{}, nil)
			tmServiceMock.EXPECT().GetSpan(gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.SpanDetail{}, expectedErr)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_InternalError)
		})

		Convey("Get succeed", func() {
			tmServiceMock.EXPECT().GetTraceModelByID(gomock.Any(), gomock.Any()).Return(interfaces.TraceModel{}, nil)
			tmServiceMock.EXPECT().GetSpan(gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.SpanDetail{}, nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func TestPreviewSpanRelatedLogList(t *testing.T) {
	Convey("Test handler PreviewSpanRelatedLogList", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		hydraMock := rmock.NewMockHydra(mockCtrl)
		tmServiceMock := umock.NewMockTraceModelService(mockCtrl)
		handler := mockNewTraceModelRestHandler(hydraMock, tmServiceMock)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/simulate-traces/1/spans/1/related-logs"

		Convey("Preview failed, caused by invalid X_HTTP_METHOD_OVERRIDE", func() {
			req := httptest.NewRequest(http.MethodPost, url, nil)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_InvalidParameter_OverrideMethod)
		})

		Convey("Preview failed, caused by the error from method ShouldBindJSON", func() {
			expectedErr := errors.New("some errors")
			queryParams := interfaces.RelatedLogListPreviewParams{}

			patch := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", expectedErr)
			defer patch.Reset()

			reqParamByte, _ := json.Marshal(queryParams)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_InvalidParameter_RequestBody)
		})

		Convey("Preview failed, caused by the error from func validateParamsWhenGetRelatedLogList", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_Offset).
				WithErrorDetails(fmt.Sprintf("The offset is not greater than %d", interfaces.MIN_OFFSET_NUM_OF_LIST))
			queryParams := interfaces.RelatedLogListPreviewParams{}

			patch1 := ApplyMethod(reflect.TypeOf(&gin.Context{}), "ShouldBindJSON",
				func(_ *gin.Context, obj any) error {
					o, _ := obj.(*interfaces.RelatedLogListPreviewParams)
					*o = queryParams
					return nil
				})
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateParamsWhenGetRelatedLogList, interfaces.RelatedLogListQueryParams{}, expectedErr)
			defer patch2.Reset()

			reqParamByte, _ := json.Marshal(queryParams)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_InvalidParameter_Offset)
		})

		Convey("Preview failed, caused by the error from method SimulateCreateTraceModel", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, uerrors.Uniquery_TraceModel_InternalError_SimulateCreateTraceModelFailed)
			queryParams := interfaces.RelatedLogListPreviewParams{}

			patch1 := ApplyMethod(reflect.TypeOf(&gin.Context{}), "ShouldBindJSON",
				func(_ *gin.Context, obj any) error {
					o, _ := obj.(*interfaces.RelatedLogListPreviewParams)
					*o = queryParams
					return nil
				})
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateParamsWhenGetRelatedLogList, interfaces.RelatedLogListQueryParams{}, nil)
			defer patch2.Reset()

			tmServiceMock.EXPECT().SimulateCreateTraceModel(gomock.Any(), gomock.Any()).Return(interfaces.TraceModel{}, expectedErr)

			reqParamByte, _ := json.Marshal(queryParams)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_TraceModel_InternalError_SimulateCreateTraceModelFailed)
		})

		Convey("Preview failed, caused by the error from method SimulateUpdateTraceModel", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, uerrors.Uniquery_TraceModel_InternalError_SimulateUpdateTraceModelFailed)
			queryParams := interfaces.RelatedLogListPreviewParams{
				TraceModel: interfaces.TraceModel{
					ID: "1",
				},
			}

			patch1 := ApplyMethod(reflect.TypeOf(&gin.Context{}), "ShouldBindJSON",
				func(_ *gin.Context, obj any) error {
					o, _ := obj.(*interfaces.RelatedLogListPreviewParams)
					*o = queryParams
					return nil
				})
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateParamsWhenGetRelatedLogList, interfaces.RelatedLogListQueryParams{}, nil)
			defer patch2.Reset()

			tmServiceMock.EXPECT().SimulateUpdateTraceModel(gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.TraceModel{}, expectedErr)

			reqParamByte, _ := json.Marshal(queryParams)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_TraceModel_InternalError_SimulateUpdateTraceModelFailed)
		})

		Convey("Preview failed, caused by the error from method GetSpanRelatedLogList", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, uerrors.Uniquery_InternalError)
			queryParams := interfaces.RelatedLogListPreviewParams{
				TraceModel: interfaces.TraceModel{
					ID: "1",
				},
			}

			patch1 := ApplyMethod(reflect.TypeOf(&gin.Context{}), "ShouldBindJSON",
				func(_ *gin.Context, obj any) error {
					o, _ := obj.(*interfaces.RelatedLogListPreviewParams)
					*o = queryParams
					return nil
				})
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateParamsWhenGetRelatedLogList, interfaces.RelatedLogListQueryParams{}, nil)
			defer patch2.Reset()

			tmServiceMock.EXPECT().SimulateUpdateTraceModel(gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.TraceModel{}, nil)
			tmServiceMock.EXPECT().GetSpanRelatedLogList(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, int64(0), expectedErr)

			reqParamByte, _ := json.Marshal(queryParams)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_InternalError)
		})

		Convey("Preview succeed", func() {
			queryParams := interfaces.RelatedLogListPreviewParams{
				TraceModel: interfaces.TraceModel{
					ID: "1",
				},
			}

			patch1 := ApplyMethod(reflect.TypeOf(&gin.Context{}), "ShouldBindJSON",
				func(_ *gin.Context, obj any) error {
					o, _ := obj.(*interfaces.RelatedLogListPreviewParams)
					*o = queryParams
					return nil
				})
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateParamsWhenGetRelatedLogList, interfaces.RelatedLogListQueryParams{}, nil)
			defer patch2.Reset()

			tmServiceMock.EXPECT().SimulateUpdateTraceModel(gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.TraceModel{}, nil)
			tmServiceMock.EXPECT().GetSpanRelatedLogList(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, int64(0), nil)

			reqParamByte, _ := json.Marshal(queryParams)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func TestGetSpanRelatedLogList(t *testing.T) {
	Convey("Test handler GetSpanRelatedLogList", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		hydraMock := rmock.NewMockHydra(mockCtrl)
		tmServiceMock := umock.NewMockTraceModelService(mockCtrl)
		handler := mockNewTraceModelRestHandler(hydraMock, tmServiceMock)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/trace-models/1/traces/1/spans/1/related-logs"

		Convey("Preview failed, caused by invalid X_HTTP_METHOD_OVERRIDE", func() {
			req := httptest.NewRequest(http.MethodPost, url, nil)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_InvalidParameter_OverrideMethod)
		})

		Convey("Preview failed, caused by the error from method ShouldBindJSON", func() {
			expectedErr := errors.New("some errors")
			queryParams := interfaces.RelatedLogListQueryParams{}

			patch := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", expectedErr)
			defer patch.Reset()

			reqParamByte, _ := json.Marshal(queryParams)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_InvalidParameter_RequestBody)
		})

		Convey("Preview failed, caused by the error from func validateParamsWhenGetRelatedLogList", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_Offset).
				WithErrorDetails(fmt.Sprintf("The offset is not greater than %d", interfaces.MIN_OFFSET_NUM_OF_LIST))
			queryParams := interfaces.RelatedLogListQueryParams{}

			patch1 := ApplyMethod(reflect.TypeOf(&gin.Context{}), "ShouldBindJSON",
				func(_ *gin.Context, obj any) error {
					o, _ := obj.(*interfaces.RelatedLogListQueryParams)
					*o = queryParams
					return nil
				})
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateParamsWhenGetRelatedLogList, interfaces.RelatedLogListQueryParams{}, expectedErr)
			defer patch2.Reset()

			reqParamByte, _ := json.Marshal(queryParams)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_InvalidParameter_Offset)
		})

		Convey("Preview failed, caused by the error from method GetTraceModelByID", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, uerrors.Uniquery_InternalError)
			queryParams := interfaces.RelatedLogListQueryParams{}

			patch1 := ApplyMethod(reflect.TypeOf(&gin.Context{}), "ShouldBindJSON",
				func(_ *gin.Context, obj any) error {
					o, _ := obj.(*interfaces.RelatedLogListQueryParams)
					*o = queryParams
					return nil
				})
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateParamsWhenGetRelatedLogList, interfaces.RelatedLogListQueryParams{}, nil)
			defer patch2.Reset()

			tmServiceMock.EXPECT().GetTraceModelByID(gomock.Any(), gomock.Any()).Return(interfaces.TraceModel{}, expectedErr)

			reqParamByte, _ := json.Marshal(queryParams)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_InternalError)
		})

		Convey("Preview failed, caused by the error from method GetSpanRelatedLogList", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, uerrors.Uniquery_InternalError)
			queryParams := interfaces.RelatedLogListQueryParams{}

			patch1 := ApplyMethod(reflect.TypeOf(&gin.Context{}), "ShouldBindJSON",
				func(_ *gin.Context, obj any) error {
					o, _ := obj.(*interfaces.RelatedLogListQueryParams)
					*o = queryParams
					return nil
				})
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateParamsWhenGetRelatedLogList, interfaces.RelatedLogListQueryParams{}, nil)
			defer patch2.Reset()

			tmServiceMock.EXPECT().GetTraceModelByID(gomock.Any(), gomock.Any()).Return(interfaces.TraceModel{}, nil)
			tmServiceMock.EXPECT().GetSpanRelatedLogList(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, int64(0), expectedErr)

			reqParamByte, _ := json.Marshal(queryParams)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_InternalError)
		})

		Convey("Get succeed", func() {
			queryParams := interfaces.RelatedLogListQueryParams{}

			patch1 := ApplyMethod(reflect.TypeOf(&gin.Context{}), "ShouldBindJSON",
				func(_ *gin.Context, obj any) error {
					o, _ := obj.(*interfaces.RelatedLogListQueryParams)
					*o = queryParams
					return nil
				})
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateParamsWhenGetRelatedLogList, interfaces.RelatedLogListQueryParams{}, nil)
			defer patch2.Reset()

			tmServiceMock.EXPECT().GetTraceModelByID(gomock.Any(), gomock.Any()).Return(interfaces.TraceModel{}, nil)
			tmServiceMock.EXPECT().GetSpanRelatedLogList(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, int64(0), nil)

			reqParamByte, _ := json.Marshal(queryParams)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}
