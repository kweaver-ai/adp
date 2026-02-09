// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	rmock "github.com/kweaver-ai/kweaver-go-lib/rest/mock"
	. "github.com/smartystreets/goconvey/convey"

	"data-model/common"
	derrors "data-model/errors"
	"data-model/interfaces"
	dmock "data-model/interfaces/mock"
)

func MockNewTraceModelRestHandler(appSetting *common.AppSetting,
	hydra rest.Hydra, tms interfaces.TraceModelService) (r *restHandler) {
	r = &restHandler{
		appSetting: appSetting,
		hydra:      hydra,
		tms:        tms,
	}
	return r
}

func Test_TraceModelRestHandler_CreateTraceModels(t *testing.T) {
	Convey("Test CreateTraceModels", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tms := dmock.NewMockTraceModelService(mockCtrl)
		hydra := rmock.NewMockHydra(mockCtrl)

		handler := MockNewTraceModelRestHandler(appSetting, hydra, tms)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/trace-models"

		Convey("Create failed, caused by the error from method ShouldBindJSON", func() {
			expectedErr := errors.New("some errors")
			models := []interfaces.TraceModel{}

			patch := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", expectedErr)
			defer patch.Reset()

			reqParamByte, _ := sonic.Marshal(models)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodPost)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Create failed, because no model was passed in", func() {
			models := []interfaces.TraceModel{}

			patch := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", nil)
			defer patch.Reset()

			reqParamByte, _ := sonic.Marshal(models)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodPost)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Create failed, caused by the error from func convertTraceModels", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidParameter_SpanSourceType).
				WithErrorDetails("some errors")
			models := make([]interfaces.TraceModel, 1)

			patch := ApplyFuncReturn(convertTraceModels, expectedErr)
			defer patch.Reset()

			reqParamByte, _ := sonic.Marshal(models)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodPost)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Create failed, caused by the error from func validateTraceModelsWhenCreate", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusForbidden, derrors.DataModel_TraceModel_ModelNameExisted).
				WithErrorDetails("some errors")
			models := make([]interfaces.TraceModel, 1)

			patch1 := ApplyFuncReturn(convertTraceModels, nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateTraceModelsWhenCreate, expectedErr)
			defer patch2.Reset()

			reqParamByte, _ := sonic.Marshal(models)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodPost)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("Create failed, caused by the error from method CreateTraceModels", func() {
			models := make([]interfaces.TraceModel, 1)
			expectedErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_DataView_InternalError_GetDetailedDataViewMapByIDsFailed).
				WithErrorDetails("some errors")

			patch1 := ApplyFuncReturn(convertTraceModels, nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateTraceModelsWhenCreate, nil)
			defer patch2.Reset()

			tms.EXPECT().CreateTraceModels(gomock.Any(), gomock.Any()).Return(nil, expectedErr)

			reqParamByte, _ := sonic.Marshal(models)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodPost)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Create succeed", func() {
			models := make([]interfaces.TraceModel, 1)
			expectedModelIDs := []string{"1", "2"}

			patch1 := ApplyFuncReturn(convertTraceModels, nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateTraceModelsWhenCreate, nil)
			defer patch2.Reset()

			tms.EXPECT().CreateTraceModels(gomock.Any(), gomock.Any()).Return(expectedModelIDs, nil)

			reqParamByte, _ := sonic.Marshal(models)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodPost)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})
	})
}

func Test_TraceModelRestHandler_SimulateCreateTraceModel(t *testing.T) {
	Convey("Test SimulateCreateTraceModel", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tms := dmock.NewMockTraceModelService(mockCtrl)
		hydra := rmock.NewMockHydra(mockCtrl)

		handler := MockNewTraceModelRestHandler(appSetting, hydra, tms)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/simulate-trace-models"

		Convey("Simulate create failed, caused by the error from method ShouldBindJSON", func() {
			expectedErr := errors.New("some errors")
			model := interfaces.TraceModel{}

			patch := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", expectedErr)
			defer patch.Reset()

			reqParamByte, _ := sonic.Marshal(model)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodPost)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Simulate create failed, caused by the error from func convertTraceModels", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidParameter_SpanSourceType).
				WithErrorDetails("some errors")
			model := interfaces.TraceModel{}

			patch1 := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(convertTraceModels, expectedErr)
			defer patch2.Reset()

			reqParamByte, _ := sonic.Marshal(model)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodPost)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Simulate create failed, caused by the error from func validateTraceModelsWhenCreate", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusForbidden, derrors.DataModel_TraceModel_ModelNameExisted).
				WithErrorDetails("some errors")
			model := interfaces.TraceModel{}

			patch1 := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(convertTraceModels, nil)
			defer patch2.Reset()

			patch3 := ApplyFuncReturn(validateTraceModelsWhenCreate, expectedErr)
			defer patch3.Reset()

			reqParamByte, _ := sonic.Marshal(model)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodPost)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("Simulate create failed, caused by the error from method SimulateCreateTraceModel", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_DataView_InternalError_GetDetailedDataViewMapByIDsFailed).
				WithErrorDetails("some errors")
			model := interfaces.TraceModel{}

			patch1 := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(convertTraceModels, nil)
			defer patch2.Reset()

			patch3 := ApplyFuncReturn(validateTraceModelsWhenCreate, nil)
			defer patch3.Reset()

			tms.EXPECT().SimulateCreateTraceModel(gomock.Any(), gomock.Any()).Return(model, expectedErr)

			reqParamByte, _ := sonic.Marshal(model)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodPost)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Simulate create succeed", func() {
			model := interfaces.TraceModel{}

			patch1 := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(convertTraceModels, nil)
			defer patch2.Reset()

			patch3 := ApplyFuncReturn(validateTraceModelsWhenCreate, nil)
			defer patch3.Reset()

			tms.EXPECT().SimulateCreateTraceModel(gomock.Any(), gomock.Any()).Return(model, nil)

			reqParamByte, _ := sonic.Marshal(model)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodPost)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func Test_TraceModelRestHandler_DeleteTraceModels(t *testing.T) {
	Convey("Test DeleteTraceModels", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tms := dmock.NewMockTraceModelService(mockCtrl)
		hydra := rmock.NewMockHydra(mockCtrl)

		handler := MockNewTraceModelRestHandler(appSetting, hydra, tms)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/trace-models/1,2"

		Convey("Delete failed, caused by the error from method GetSimpleTraceModelMapByIDs", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_TraceModel_InternalError_GetSimpleTraceModelMapByIDsFailed).
				WithErrorDetails("some errors")

			tms.EXPECT().GetSimpleTraceModelMapByIDs(gomock.Any(), gomock.Any()).Return(nil, expectedErr)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Delete failed, because some models do not exist", func() {

			tms.EXPECT().GetSimpleTraceModelMapByIDs(gomock.Any(), gomock.Any()).Return(nil, nil)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("Delete failed, caused by the error from method DeleteTraceModels", func() {
			expectedmodelMap := map[string]interfaces.TraceModel{
				"1": {},
				"2": {},
			}
			expectedErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_TraceModel_InternalError_DeleteTraceModelsFailed).
				WithErrorDetails("some errors")

			tms.EXPECT().GetSimpleTraceModelMapByIDs(gomock.Any(), gomock.Any()).Return(expectedmodelMap, nil)
			tms.EXPECT().DeleteTraceModels(gomock.Any(), gomock.Any()).Return(expectedErr)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Delete succeed", func() {
			expectedmodelMap := map[string]interfaces.TraceModel{
				"1": {},
				"2": {},
			}

			tms.EXPECT().GetSimpleTraceModelMapByIDs(gomock.Any(), gomock.Any()).Return(expectedmodelMap, nil)
			tms.EXPECT().DeleteTraceModels(gomock.Any(), gomock.Any()).Return(nil)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})
	})
}

func Test_TraceModelRestHandler_UpdateTraceModel(t *testing.T) {
	Convey("Test UpdateTraceModel", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tms := dmock.NewMockTraceModelService(mockCtrl)
		hydra := rmock.NewMockHydra(mockCtrl)

		handler := MockNewTraceModelRestHandler(appSetting, hydra, tms)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/trace-models/1"

		reqModel := interfaces.TraceModel{
			Name: "test1",
		}

		Convey("Update failed, caused by the error from method GetSimpleTraceModelMapByIDs", func() {

			expectedErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_DataView_InternalError_GetSimpleDataViewMapByNamesFailed).
				WithErrorDetails("some errors")
			tms.EXPECT().GetSimpleTraceModelMapByIDs(gomock.Any(), gomock.Any()).Return(nil, expectedErr)

			reqParamByte, _ := sonic.Marshal(reqModel)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Update failed, caused by the simple traceModelMap is nil", func() {

			tms.EXPECT().GetSimpleTraceModelMapByIDs(gomock.Any(), gomock.Any()).Return(nil, nil)

			reqParamByte, _ := sonic.Marshal(reqModel)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("Update failed, caused by the error from method ShouldBindJSON", func() {
			expectedSimpleMap := map[string]interfaces.TraceModel{"1": {}}
			expectedErr := errors.New("some errors")

			tms.EXPECT().GetSimpleTraceModelMapByIDs(gomock.Any(), gomock.Any()).Return(expectedSimpleMap, nil)

			patch2 := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", expectedErr)
			defer patch2.Reset()

			reqParamByte, _ := sonic.Marshal(reqModel)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Update failed, caused by the error from func convertTraceModels", func() {
			expectedSimpleMap := map[string]interfaces.TraceModel{"1": {}}
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidParameter_SpanSourceType).
				WithErrorDetails("some errors")

			tms.EXPECT().GetSimpleTraceModelMapByIDs(gomock.Any(), gomock.Any()).Return(expectedSimpleMap, nil)

			patch2 := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", nil)
			defer patch2.Reset()

			patch3 := ApplyFuncReturn(convertTraceModels, expectedErr)
			defer patch3.Reset()

			reqParamByte, _ := sonic.Marshal(reqModel)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Update failed, caused by the error from func validateTraceModelWhenUpdate", func() {
			expectedSimpleMap := map[string]interfaces.TraceModel{"1": {}}
			expectedErr := rest.NewHTTPError(testCtx, http.StatusForbidden, derrors.DataModel_TraceModel_ModelNameExisted).
				WithErrorDetails("some errors")

			tms.EXPECT().GetSimpleTraceModelMapByIDs(gomock.Any(), gomock.Any()).Return(expectedSimpleMap, nil)

			patch2 := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", nil)
			defer patch2.Reset()

			patch3 := ApplyFuncReturn(convertTraceModels, nil)
			defer patch3.Reset()

			patch4 := ApplyFuncReturn(validateTraceModelWhenUpdate, expectedErr)
			defer patch4.Reset()

			reqParamByte, _ := sonic.Marshal(reqModel)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("Update failed, caused by the error from method UpdateTraceModel", func() {
			expectedSimpleMap := map[string]interfaces.TraceModel{"1": {}}
			expectedErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_TraceModel_InternalError_UpdateTraceModelFailed).
				WithErrorDetails("some errors")

			tms.EXPECT().GetSimpleTraceModelMapByIDs(gomock.Any(), gomock.Any()).Return(expectedSimpleMap, nil)

			patch2 := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", nil)
			defer patch2.Reset()

			patch3 := ApplyFuncReturn(convertTraceModels, nil)
			defer patch3.Reset()

			patch4 := ApplyFuncReturn(validateTraceModelWhenUpdate, nil)
			defer patch4.Reset()

			tms.EXPECT().UpdateTraceModel(gomock.Any(), gomock.Any()).Return(expectedErr)

			reqParamByte, _ := sonic.Marshal(reqModel)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Update succeed", func() {
			expectedSimpleMap := map[string]interfaces.TraceModel{"1": {}}

			tms.EXPECT().GetSimpleTraceModelMapByIDs(gomock.Any(), gomock.Any()).Return(expectedSimpleMap, nil)

			patch2 := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", nil)
			defer patch2.Reset()

			patch3 := ApplyFuncReturn(convertTraceModels, nil)
			defer patch3.Reset()

			patch4 := ApplyFuncReturn(validateTraceModelWhenUpdate, nil)
			defer patch4.Reset()

			tms.EXPECT().UpdateTraceModel(gomock.Any(), gomock.Any()).Return(nil)

			reqParamByte, _ := sonic.Marshal(reqModel)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})
	})
}

func Test_TraceModelRestHandler_SimulateUpdateTraceModel(t *testing.T) {
	Convey("Test SimulateUpdateTraceModel", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tms := dmock.NewMockTraceModelService(mockCtrl)
		hydra := rmock.NewMockHydra(mockCtrl)

		handler := MockNewTraceModelRestHandler(appSetting, hydra, tms)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/simulate-trace-models/1"

		reqModel := interfaces.TraceModel{
			Name: "test1",
		}

		Convey("Simulate update failed, caused by the error from method GetSimpleTraceModelMapByIDs", func() {

			expectedErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_DataView_InternalError_GetSimpleDataViewMapByNamesFailed).
				WithErrorDetails("some errors")
			tms.EXPECT().GetSimpleTraceModelMapByIDs(gomock.Any(), gomock.Any()).Return(nil, expectedErr)

			reqParamByte, _ := sonic.Marshal(reqModel)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Simulate update failed, caused by the simple traceModelMap is nil", func() {

			tms.EXPECT().GetSimpleTraceModelMapByIDs(gomock.Any(), gomock.Any()).Return(nil, nil)

			reqParamByte, _ := sonic.Marshal(reqModel)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("Simulate update failed, caused by the error from method ShouldBindJSON", func() {
			expectedSimpleMap := map[string]interfaces.TraceModel{"1": {}}
			expectedErr := errors.New("some errors")

			tms.EXPECT().GetSimpleTraceModelMapByIDs(gomock.Any(), gomock.Any()).Return(expectedSimpleMap, nil)

			patch2 := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", expectedErr)
			defer patch2.Reset()

			reqParamByte, _ := sonic.Marshal(reqModel)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Simulate update failed, caused by the error from func convertTraceModels", func() {
			expectedSimpleMap := map[string]interfaces.TraceModel{"1": {}}
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidParameter_SpanSourceType).
				WithErrorDetails("some errors")

			tms.EXPECT().GetSimpleTraceModelMapByIDs(gomock.Any(), gomock.Any()).Return(expectedSimpleMap, nil)

			patch2 := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", nil)
			defer patch2.Reset()

			patch3 := ApplyFuncReturn(convertTraceModels, expectedErr)
			defer patch3.Reset()

			reqParamByte, _ := sonic.Marshal(reqModel)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Simulate update failed, caused by the error from func validateTraceModelWhenUpdate", func() {
			expectedSimpleMap := map[string]interfaces.TraceModel{"1": {}}
			expectedErr := rest.NewHTTPError(testCtx, http.StatusForbidden, derrors.DataModel_TraceModel_ModelNameExisted).
				WithErrorDetails("some errors")

			tms.EXPECT().GetSimpleTraceModelMapByIDs(gomock.Any(), gomock.Any()).Return(expectedSimpleMap, nil)

			patch2 := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", nil)
			defer patch2.Reset()

			patch3 := ApplyFuncReturn(convertTraceModels, nil)
			defer patch3.Reset()

			patch4 := ApplyFuncReturn(validateTraceModelWhenUpdate, expectedErr)
			defer patch4.Reset()

			reqParamByte, _ := sonic.Marshal(reqModel)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("Simulate update failed, caused by the error from method SimulateUpdateTraceModel", func() {
			expectedSimpleMap := map[string]interfaces.TraceModel{"1": {}}
			expectedErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_TraceModel_InternalError_UpdateTraceModelFailed).
				WithErrorDetails("some errors")

			tms.EXPECT().GetSimpleTraceModelMapByIDs(gomock.Any(), gomock.Any()).Return(expectedSimpleMap, nil)

			patch2 := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", nil)
			defer patch2.Reset()

			patch3 := ApplyFuncReturn(convertTraceModels, nil)
			defer patch3.Reset()

			patch4 := ApplyFuncReturn(validateTraceModelWhenUpdate, nil)
			defer patch4.Reset()

			tms.EXPECT().SimulateUpdateTraceModel(gomock.Any(), gomock.Any()).Return(reqModel, expectedErr)

			reqParamByte, _ := sonic.Marshal(reqModel)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Simulate update succeed", func() {
			expectedSimpleMap := map[string]interfaces.TraceModel{"1": {}}

			tms.EXPECT().GetSimpleTraceModelMapByIDs(gomock.Any(), gomock.Any()).Return(expectedSimpleMap, nil)

			patch2 := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", nil)
			defer patch2.Reset()

			patch3 := ApplyFuncReturn(convertTraceModels, nil)
			defer patch3.Reset()

			patch4 := ApplyFuncReturn(validateTraceModelWhenUpdate, nil)
			defer patch4.Reset()

			tms.EXPECT().SimulateUpdateTraceModel(gomock.Any(), gomock.Any()).Return(reqModel, nil)

			reqParamByte, _ := sonic.Marshal(reqModel)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func Test_TraceModelRestHandler_GetTraceModels(t *testing.T) {
	Convey("Test GetTraceModels", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tms := dmock.NewMockTraceModelService(mockCtrl)
		hydra := rmock.NewMockHydra(mockCtrl)

		handler := MockNewTraceModelRestHandler(appSetting, hydra, tms)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/trace-models/1,2"

		Convey("Get failed, caused by the error from method GetTraceModels", func() {

			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError,
				derrors.DataModel_TraceModel_InternalError_GetDetailedTraceModelMapByIDsFailed)
			tms.EXPECT().GetTraceModels(gomock.Any(), gomock.Any()).Return(nil, expectedHttpErr)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Get succeed", func() {

			tms.EXPECT().GetTraceModels(gomock.Any(), gomock.Any()).Return(nil, nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func Test_TraceModelRestHandler_ListTraceModels(t *testing.T) {
	Convey("Test ListTraceModels", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tms := dmock.NewMockTraceModelService(mockCtrl)
		hydra := rmock.NewMockHydra(mockCtrl)

		handler := MockNewTraceModelRestHandler(appSetting, hydra, tms)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/trace-models?span_source_type=1"

		Convey("List failed, caused by the error from func validateNameandNamePattern", func() {
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest,
				derrors.DataModel_ConflictParameter_NameAndNamePatternCoexist).
				WithErrorDetails("Parameters name_pattern and name are passed in at the same time")
			patch := ApplyFuncReturn(validateNameandNamePattern, expectedHttpErr)
			defer patch.Reset()

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("List failed, caused by the error from func validateSpanSourceType", func() {
			patch1 := ApplyFuncReturn(validateNameandNamePattern, nil)
			defer patch1.Reset()

			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest,
				derrors.DataModel_TraceModel_InvalidParameter_SpanSourceType).
				WithErrorDetails("span_source_type is invalid, valid span_source_type is data_view or data_connection")
			patch2 := ApplyFuncReturn(validateSpanSourceType, expectedHttpErr)
			defer patch2.Reset()

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("List failed, caused by the error from func validatePaginationQueryParameters", func() {
			patch1 := ApplyFuncReturn(validateNameandNamePattern, nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateSpanSourceType, nil)
			defer patch2.Reset()

			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest,
				derrors.DataModel_InvalidParameter_Offset).
				WithErrorDetails("some errors")
			patch3 := ApplyFuncReturn(validatePaginationQueryParameters,
				interfaces.PaginationQueryParameters{}, expectedHttpErr)
			defer patch3.Reset()

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("List failed, caused by the error from method ListTraceModels", func() {
			patch1 := ApplyFuncReturn(validateNameandNamePattern, nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateSpanSourceType, nil)
			defer patch2.Reset()

			patch3 := ApplyFuncReturn(validatePaginationQueryParameters,
				interfaces.PaginationQueryParameters{}, nil)
			defer patch3.Reset()

			expectedEntries := []interfaces.TraceModelListEntry{}
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError,
				derrors.DataModel_TraceModel_InternalError_ListTraceModelsFailed).WithErrorDetails("some error")

			tms.EXPECT().ListTraceModels(gomock.Any(), gomock.Any()).
				Return(expectedEntries, 0, expectedHttpErr)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("List succeed", func() {
			patch1 := ApplyFuncReturn(validateNameandNamePattern, nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateSpanSourceType, nil)
			defer patch2.Reset()

			patch3 := ApplyFuncReturn(validatePaginationQueryParameters,
				interfaces.PaginationQueryParameters{}, nil)
			defer patch3.Reset()

			expectedEntries := []interfaces.TraceModelListEntry{}
			tms.EXPECT().ListTraceModels(gomock.Any(), gomock.Any()).
				Return(expectedEntries, 0, nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func Test_TraceModelRestHandler_GetTraceModelFieldInfo(t *testing.T) {
	Convey("Test GetTraceModelFieldInfo", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tms := dmock.NewMockTraceModelService(mockCtrl)
		hydra := rmock.NewMockHydra(mockCtrl)

		handler := MockNewTraceModelRestHandler(appSetting, hydra, tms)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/trace-models/1/field-info"

		Convey("Get failed, caused by the error from method 'GetTraceModelFieldInfo'", func() {

			expectedErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_DataView_InternalError_GetSimpleDataViewMapByNamesFailed).
				WithErrorDetails("some errors")
			tms.EXPECT().GetTraceModelFieldInfo(gomock.Any(), gomock.Any()).Return(interfaces.TraceModelFieldInfo{}, expectedErr)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Get succeed", func() {

			tms.EXPECT().GetTraceModelFieldInfo(gomock.Any(), gomock.Any()).Return(interfaces.TraceModelFieldInfo{}, nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func Test_TraceModelRestHandler_ConvertTraceModels(t *testing.T) {
	Convey("Test convertTraceModels", t, func() {
		reqModels := []interfaces.TraceModel{}

		Convey("Convert failed, caused by the error from func `convertSpanConfig`", func() {
			expectedErr := errors.New("some errors")

			patch := ApplyFuncReturn(convertSpanConfig, expectedErr)
			defer patch.Reset()

			err := convertTraceModels(testCtx, reqModels)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Convert failed, caused by the error from func `convertRelatedLogConfig`", func() {
			expectedErr := errors.New("some errors")

			patch1 := ApplyFuncReturn(convertSpanConfig, nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(convertRelatedLogConfig, expectedErr)
			defer patch2.Reset()

			err := convertTraceModels(testCtx, reqModels)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Convert succeed", func() {
			patch1 := ApplyFuncReturn(convertSpanConfig, nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(convertRelatedLogConfig, nil)
			defer patch2.Reset()

			err := convertTraceModels(testCtx, reqModels)
			So(err, ShouldBeNil)
		})
	})
}

func Test_TraceModelRestHandler_ConvertSpanConfig(t *testing.T) {
	Convey("Test convertSpanConfig", t, func() {

		Convey("Convert failed, caused by the error from func `sonic.Marshal`", func() {
			reqModels := []interfaces.TraceModel{
				{
					SpanSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
					SpanConfig:     interfaces.SpanConfigWithDataView{},
				},
			}
			expectedErr := errors.New("some errors")
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError,
				derrors.DataModel_InternalError_MarshalDataFailed).WithErrorDetails(fmt.Sprintf("Marshal filed span_config failed, err: %v", expectedErr.Error()))

			patch := ApplyFuncReturn(sonic.Marshal, []byte(nil), expectedErr)
			defer patch.Reset()

			err := convertSpanConfig(testCtx, reqModels)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Convert failed when span_source_type is data_view, caused by the error from func `sonic.Unmarshal`", func() {
			reqModels := []interfaces.TraceModel{
				{
					SpanSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
					SpanConfig:     interfaces.SpanConfigWithDataView{},
				},
			}
			expectedErr := errors.New("some errors")

			patch1 := ApplyFuncReturn(sonic.Marshal, []byte(nil), nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(sonic.Unmarshal, expectedErr)
			defer patch2.Reset()

			err := convertSpanConfig(testCtx, reqModels)
			So(err, ShouldNotBeNil)
		})

		Convey("Convert failed when span_source_type is data_connection, caused by the error from func `sonic.Unmarshal`", func() {
			reqModels := []interfaces.TraceModel{
				{
					SpanSourceType: interfaces.SOURCE_TYPE_DATA_CONNECTION,
					SpanConfig:     interfaces.SpanConfigWithDataConnection{},
				},
			}
			expectedErr := errors.New("some errors")
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError,
				derrors.DataModel_InternalError_UnmarshalDataFailed).WithErrorDetails(fmt.Sprintf("Field span_config cannot be unmarshaled to SpanConfigWithDataConnection, err: %v", expectedErr.Error()))

			patch1 := ApplyFuncReturn(sonic.Marshal, []byte(nil), nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(sonic.Unmarshal, expectedErr)
			defer patch2.Reset()

			err := convertSpanConfig(testCtx, reqModels)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Convert failed, caused by the invalid span_source_type", func() {
			reqModels := []interfaces.TraceModel{
				{
					SpanSourceType: "",
					SpanConfig:     interfaces.SpanConfigWithDataConnection{},
				},
			}
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidParameter_SpanSourceType).
				WithErrorDetails("span_source_type is invalid, valid span_source_type is " + interfaces.SOURCE_TYPE_DATA_VIEW +
					" or " + interfaces.SOURCE_TYPE_DATA_CONNECTION)

			patch := ApplyFuncReturn(sonic.Marshal, []byte(nil), nil)
			defer patch.Reset()

			err := convertSpanConfig(testCtx, reqModels)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Convert succeed", func() {
			reqModels := []interfaces.TraceModel{
				{
					SpanSourceType: interfaces.SOURCE_TYPE_DATA_CONNECTION,
					SpanConfig:     interfaces.SpanConfigWithDataConnection{},
				},
			}

			patch1 := ApplyFuncReturn(sonic.Marshal, []byte(nil), nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(sonic.Unmarshal, nil)
			defer patch2.Reset()

			err := convertSpanConfig(testCtx, reqModels)
			So(err, ShouldBeNil)
		})
	})
}

func Test_TraceModelRestHandler_ConvertRelatedLogConfig(t *testing.T) {
	Convey("Test convertRelatedLogConfig", t, func() {

		Convey("Convert failed, caused by the error from func `sonic.Marshal`", func() {
			reqModels := []interfaces.TraceModel{
				{
					EnabledRelatedLog:    interfaces.RELATED_LOG_OPEN,
					RelatedLogSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
					RelatedLogConfig:     interfaces.RelatedLogConfigWithDataView{},
				},
			}
			expectedErr := errors.New("some errors")
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError,
				derrors.DataModel_InternalError_MarshalDataFailed).WithErrorDetails(fmt.Sprintf("Marshal filed related_log_config failed, err: %v", expectedErr.Error()))

			patch := ApplyFuncReturn(sonic.Marshal, []byte(nil), expectedErr)
			defer patch.Reset()

			err := convertRelatedLogConfig(testCtx, reqModels)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Convert failed, caused by the error from func `sonic.Unmarshal`", func() {
			reqModels := []interfaces.TraceModel{
				{
					EnabledRelatedLog:    interfaces.RELATED_LOG_OPEN,
					RelatedLogSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
					RelatedLogConfig:     interfaces.RelatedLogConfigWithDataView{},
				},
			}
			expectedErr := errors.New("some errors")

			patch1 := ApplyFuncReturn(sonic.Marshal, []byte(nil), nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(sonic.Unmarshal, expectedErr)
			defer patch2.Reset()

			err := convertRelatedLogConfig(testCtx, reqModels)
			So(err, ShouldNotBeNil)
		})

		Convey("Convert failed, caused by the invalid related_log_source_type", func() {
			reqModels := []interfaces.TraceModel{
				{
					EnabledRelatedLog:    interfaces.RELATED_LOG_OPEN,
					RelatedLogSourceType: "",
					RelatedLogConfig:     interfaces.RelatedLogConfigWithDataView{},
				},
			}

			patch := ApplyFuncReturn(sonic.Marshal, []byte(nil), nil)
			defer patch.Reset()

			err := convertRelatedLogConfig(testCtx, reqModels)
			So(err, ShouldNotBeNil)
		})

		Convey("Convert succeed", func() {
			reqModels := []interfaces.TraceModel{
				{
					EnabledRelatedLog:    interfaces.RELATED_LOG_OPEN,
					RelatedLogSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
					RelatedLogConfig:     interfaces.RelatedLogConfigWithDataView{},
				},
			}

			patch1 := ApplyFuncReturn(sonic.Marshal, []byte(nil), nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(sonic.Unmarshal, nil)
			defer patch2.Reset()

			err := convertRelatedLogConfig(testCtx, reqModels)
			So(err, ShouldBeNil)
		})
	})
}
