// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"bytes"
	"errors"
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

func MockNewDataViewRestHandler(appSetting *common.AppSetting,
	hydra rest.Hydra, dvs interfaces.DataViewService) (r *restHandler) {

	r = &restHandler{
		appSetting: appSetting,
		hydra:      hydra,
		dvs:        dvs,
	}
	return r
}

func Test_DataViewRestHandler_HandlePostOverride(t *testing.T) {
	Convey("Test HandlePostOverride", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)

		handler := MockNewDataViewRestHandler(appSetting, hydra, dvs)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/data-views"

		Convey("HandlePostOverride failed, caused by the invalid overrideMethod", func() {
			reqParamByte, _ := sonic.Marshal([]interfaces.DataView{})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodPut)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})
	})
}

// func Test_DataViewRestHandler_CreateDataViews(t *testing.T) {
// 	Convey("Test CreateDataViews", t, func() {
// 		test := setGinMode()
// 		defer test()

// 		engine := gin.New()
// 		engine.Use(gin.Recovery())

// 		mockCtrl := gomock.NewController(t)
// 		defer mockCtrl.Finish()

// 		appSetting := &common.AppSetting{}
// 		hydra := rmock.NewMockHydra(mockCtrl)
// 		dvs := dmock.NewMockDataViewService(mockCtrl)

// 		handler := MockNewDataViewRestHandler(appSetting, hydra, dvs)
// 		handler.RegisterPublic(engine)

// 		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

// 		url := "/api/mdl-data-model/v1/data-views"

// 		reqViews1 := []interfaces.DataView{
// 			{
// 				SimpleDataView: interfaces.SimpleDataView{
// 					ViewID:   "xyz",
// 					ViewName: "test1",
// 					GroupID:  "abcd",
// 				},
// 				ModuleType: interfaces.MODULE_TYPE_DATA_VIEW,
// 			},
// 		}

// 		reqViews2 := []interfaces.DataView{
// 			{
// 				SimpleDataView: interfaces.SimpleDataView{
// 					ViewID:   "test1",
// 					ViewName: "test1",
// 					GroupID:  "abcd",
// 				},
// 				ModuleType: interfaces.MODULE_TYPE_DATA_VIEW,
// 			},
// 			{
// 				SimpleDataView: interfaces.SimpleDataView{
// 					ViewID:   "test2",
// 					ViewName: "test2",
// 					GroupID:  "abcd",
// 				},
// 				ModuleType: interfaces.MODULE_TYPE_DATA_VIEW,
// 			},
// 		}

// 		Convey("Create failed, caused by import_mode is not supported", func() {
// 			reqParamByte, _ := sonic.Marshal([]interfaces.DataView{
// 				{SimpleDataView: interfaces.SimpleDataView{ViewID: "a"}, ModuleType: interfaces.MODULE_TYPE_DATA_VIEW},
// 			})
// 			createUrl := url + "?import_mode=some_value"
// 			req := httptest.NewRequest(http.MethodPost, createUrl, bytes.NewReader(reqParamByte))
// 			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodPost)
// 			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
// 		})

// 		Convey("Create failed, caused by the error from method ShouldBindJSON", func() {
// 			expectedErr := errors.New("some errors")
// 			patch := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", expectedErr)
// 			defer patch.Reset()

// 			reqParamByte, _ := sonic.Marshal(reqViews1)
// 			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
// 			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodPost)
// 			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
// 		})

// 		Convey("Create failed, caused by no data view was passed in", func() {
// 			reqParamByte, _ := sonic.Marshal([]interfaces.DataView{})
// 			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
// 			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodPost)
// 			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
// 		})

// 		Convey("Create failed, caused by import data view module name is not 'data_view'", func() {
// 			reqParamByte, _ := sonic.Marshal([]interfaces.DataView{
// 				{SimpleDataView: interfaces.SimpleDataView{ViewName: "a"},
// 					ModuleType: "someone",
// 				},
// 			})
// 			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
// 			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodPost)
// 			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
// 		})

// 		Convey("Create failed, caused by data view id is duplicated", func() {
// 			reqParamByte, _ := sonic.Marshal([]interfaces.DataView{
// 				{SimpleDataView: interfaces.SimpleDataView{ViewID: "a"}},
// 				{SimpleDataView: interfaces.SimpleDataView{ViewID: "a"}},
// 			})

// 			patch := ApplyFuncReturn(ValidateDataView, nil)
// 			defer patch.Reset()

// 			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
// 			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodPost)
// 			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
// 		})

// 		Convey("Create failed, caused by data view name duplicated in the same group", func() {
// 			reqParamByte, _ := sonic.Marshal([]interfaces.DataView{
// 				{SimpleDataView: interfaces.SimpleDataView{ViewID: "a", ViewName: "a", GroupID: "aaa"}},
// 				{SimpleDataView: interfaces.SimpleDataView{ViewID: "b", ViewName: "a", GroupID: "aaa"}},
// 			})

// 			patch := ApplyFuncReturn(ValidateDataView, nil)
// 			defer patch.Reset()

// 			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
// 			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodPost)
// 			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
// 		})

// 		Convey("Create failed, caused by the error from func validateDataViews", func() {
// 			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_DataView_NullParameter_ViewName)

// 			patch := ApplyFuncReturn(ApplyFuncReturn, expectedHttpErr)
// 			defer patch.Reset()

// 			reqParamByte, _ := sonic.Marshal(reqViews1)
// 			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
// 			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodPost)
// 			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
// 		})

// 		Convey("Create failed, caused by the error from method CreateDataViews", func() {
// 			patch := ApplyFuncReturn(ValidateDataView, nil)
// 			defer patch.Reset()

// 			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_DataView_InternalError_CreateDataViewsFailed)
// 			expectedViewIDs := []string{}
// 			dvs.EXPECT().CreateDataViews(gomock.Any(), gomock.Not(gomock.Len(0)), gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedViewIDs, expectedHttpErr)

// 			reqParamByte, _ := sonic.Marshal(reqViews1)
// 			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
// 			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodPost)
// 			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
// 		})

// 		Convey("Create succeed", func() {
// 			patch := ApplyFuncReturn(ValidateDataView, nil)
// 			defer patch.Reset()
// 			dvs.EXPECT().CreateDataViews(gomock.Any(), gomock.Not(gomock.Len(0)), gomock.Any(),
// 				gomock.Any(), gomock.Any()).Return([]string{"1", "2"}, nil)

// 			reqParamByte, _ := sonic.Marshal(reqViews2)
// 			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
// 			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodPost)
// 			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
// 		})
// 	})
// }

func Test_DataViewRestHandler_DeleteDataViews(t *testing.T) {
	Convey("Test DeleteDataViews", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)

		handler := MockNewDataViewRestHandler(appSetting, hydra, dvs)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/data-views"

		viewIDsReq := interfaces.ViewIDsReq{
			IDs: []string{"1", "2"},
		}

		reqParamByte, _ := sonic.Marshal(viewIDsReq)

		Convey("Get failed, caused by the error from method ShouldBindJSON", func() {
			expectedErr := errors.New("some errors")
			patch := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", expectedErr)
			defer patch.Reset()

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodDelete)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Delete failed, caused by the error from method CheckDataViewExistByID", func() {

			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError,
				derrors.DataModel_DataView_InternalError_CheckViewIfExistFailed).
				WithErrorDetails("Get data view name by id failed, err: some errors")
			dvs.EXPECT().CheckDataViewExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, expectedHttpErr).AnyTimes()

			url = "/api/mdl-data-model/v1/data-views/1,2"
			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Delete failed, caused by the error from method DeleteDataViews", func() {

			dvs.EXPECT().CheckDataViewExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return("1", true, nil).AnyTimes()

			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError,
				derrors.DataModel_DataView_InternalError_DeleteDataViewsFailed).
				WithErrorDetails("some errors")
			dvs.EXPECT().DeleteDataViews(gomock.Any(), gomock.Any()).Return(expectedHttpErr)

			url = "/api/mdl-data-model/v1/data-views/1,2"
			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Delete succeed", func() {

			dvs.EXPECT().CheckDataViewExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return("1", true, nil).AnyTimes()
			dvs.EXPECT().DeleteDataViews(gomock.Any(), gomock.Any()).Return(nil)

			url = "/api/mdl-data-model/v1/data-views/1,2"
			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})
	})
}

// func Test_DataViewRestHandler_UpdateDataView(t *testing.T) {
// 	Convey("Test UpdateDataView", t, func() {
// 		test := setGinMode()
// 		defer test()

// 		engine := gin.New()
// 		engine.Use(gin.Recovery())

// 		mockCtrl := gomock.NewController(t)
// 		defer mockCtrl.Finish()

// 		appSetting := &common.AppSetting{}
// 		hydra := rmock.NewMockHydra(mockCtrl)
// 		dvs := dmock.NewMockDataViewService(mockCtrl)

// 		handler := MockNewDataViewRestHandler(appSetting, hydra, dvs)
// 		handler.RegisterPublic(engine)

// 		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

// 		url := "/api/mdl-data-model/v1/data-views/1"

// 		reqView := interfaces.DataView{
// 			SimpleDataView: interfaces.SimpleDataView{
// 				ViewName: "test1",
// 			},
// 		}

// 		Convey("Update failed, caused by the error from method ShouldBindJSON", func() {
// 			expectedErr := errors.New("some errors")
// 			patch := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", expectedErr)
// 			defer patch.Reset()

// 			reqParamByte, _ := sonic.Marshal(reqView)
// 			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
// 			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
// 		})

// 		Convey("Update failed, caused by the error from func ValidateDataView", func() {

// 			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_DataView_NullParameter_ViewName)
// 			patch2 := ApplyFuncReturn(ValidateDataView, expectedHttpErr)
// 			defer patch2.Reset()

// 			reqParamByte, _ := sonic.Marshal(reqView)
// 			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
// 			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
// 		})

// 		Convey("Update failed, caused by the error from method UpdateDataView", func() {
// 			patch := ApplyFuncReturn(ValidateDataView, nil)
// 			defer patch.Reset()

// 			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_DataView_InternalError_UpdateDataViewFailed).
// 				WithErrorDetails("some errors")
// 			dvs.EXPECT().UpdateDataView(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedHttpErr)

// 			reqParamByte, _ := sonic.Marshal(reqView)
// 			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
// 			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
// 		})

// 		Convey("Update succeed", func() {
// 			patch := ApplyFuncReturn(ValidateDataView, nil)
// 			defer patch.Reset()

// 			dvs.EXPECT().UpdateDataView(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

// 			reqParamByte, _ := sonic.Marshal(reqView)
// 			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
// 			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
// 		})
// 	})
// }

func Test_DataViewRestHandler_GetDataViews(t *testing.T) {
	Convey("Test GetDataViews", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)

		handler := MockNewDataViewRestHandler(appSetting, hydra, dvs)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/data-views"

		viewIDsReq := interfaces.ViewIDsReq{
			IDs: []string{"1", "2"},
		}

		reqParamByte, _ := sonic.Marshal(viewIDsReq)

		Convey("Get failed, caused by the error from method ShouldBindJSON", func() {
			expectedErr := errors.New("some errors")
			patch := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", expectedErr)
			defer patch.Reset()

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Get failed, caused by the error from method GetDataViews", func() {

			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_DataView_InternalError_GetDataViewsFailed)
			dvs.EXPECT().GetDataViews(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, expectedHttpErr)

			url = "/api/mdl-data-model/v1/data-views/1,2"
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Get succeed", func() {

			dvs.EXPECT().GetDataViews(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*interfaces.DataView{}, nil)

			url = "/api/mdl-data-model/v1/data-views/1,2"
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

// func Test_DataViewRestHandler_ListDataViews(t *testing.T) {
// 	Convey("Test ListDataViews", t, func() {
// 		test := setGinMode()
// 		defer test()

// 		engine := gin.New()
// 		engine.Use(gin.Recovery())

// 		mockCtrl := gomock.NewController(t)
// 		defer mockCtrl.Finish()

// 		appSetting := &common.AppSetting{}
// 		hydra := rmock.NewMockHydra(mockCtrl)
// 		dvs := dmock.NewMockDataViewService(mockCtrl)

// 		handler := MockNewDataViewRestHandler(appSetting, hydra, dvs)
// 		handler.RegisterPublic(engine)

// 		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

// 		url := "/api/mdl-data-model/v1/data-views"

// 		Convey("List failed, caused by the error from builtin", func() {
// 			url = url + "?direction=desc&sort=update_time&limit=1000&offset=0&builtin=foo"
// 			req := httptest.NewRequest(http.MethodGet, url, nil)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
// 		})

// 		Convey("List failed, caused by the error from open_streaming", func() {
// 			url = url + "?direction=desc&sort=update_time&limit=1000&offset=0&open_streaming=foo"
// 			req := httptest.NewRequest(http.MethodGet, url, nil)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
// 		})

// 		Convey("List failed, caused by the error from func name and namePattern", func() {
// 			url = url + "?direction=desc&sort=update_time&limit=1000&offset=0&name=a&name_pattern=b"
// 			req := httptest.NewRequest(http.MethodGet, url, nil)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
// 		})

// 		Convey("List failed, caused by the error from func validatePaginationQueryParameters", func() {
// 			url = url + "?direction=desc&sort=update_time&limit=1000&offset=a"
// 			req := httptest.NewRequest(http.MethodGet, url, nil)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
// 		})

// 		Convey("List failed, caused by the error from method ListDataViews", func() {
// 			url = url + "?direction=desc&sort=update_time&limit=1000&offset=0"
// 			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_DataView_InternalError_ListDataViewsFailed)

// 			dvs.EXPECT().ListDataViews(gomock.Any(), gomock.Any()).Return(nil, 0, expectedHttpErr)

// 			req := httptest.NewRequest(http.MethodGet, url, nil)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
// 		})

// 		Convey("List succeed", func() {
// 			url = url + "?direction=desc&sort=update_time&limit=1000&offset=0"
// 			expectedEntries := []*interfaces.SimpleDataView{}
// 			dvs.EXPECT().ListDataViews(gomock.Any(), gomock.Any()).Return(expectedEntries, 0, nil)

// 			req := httptest.NewRequest(http.MethodGet, url, nil)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
// 		})
// 	})
// }
