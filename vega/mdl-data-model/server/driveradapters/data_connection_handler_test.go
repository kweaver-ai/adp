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

func MockNewDataConnectionRestHandler(appSetting *common.AppSetting,
	hydra rest.Hydra,
	dcs interfaces.DataConnectionService) (r *restHandler) {

	r = &restHandler{
		appSetting: appSetting,
		hydra:      hydra,
		dcs:        dcs,
	}
	return r
}

func Test_DataConnectionRestHandler_CreateDataConnection(t *testing.T) {
	Convey("Test CreateDataConnection", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		dcs := dmock.NewMockDataConnectionService(mockCtrl)

		handler := MockNewDataConnectionRestHandler(appSetting, hydra, dcs)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/data-connections"

		Convey("Create failed, caused by the error from method ShouldBindJSON", func() {
			expectedErr := errors.New("some errors")
			conn := interfaces.DataConnection{}

			patch := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", expectedErr)
			defer patch.Reset()

			reqParamByte, _ := sonic.Marshal(conn)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Create failed, caused by the error from func validateDataConnectionWhenCreate", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusForbidden,
				derrors.DataModel_TraceModel_ModelNameExisted).WithErrorDetails("some errors")
			conn := interfaces.DataConnection{}

			patch1 := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateDataConnectionWhenCreate, expectedErr)
			defer patch2.Reset()

			reqParamByte, _ := sonic.Marshal(conn)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("Create failed, caused by the error from method CreateDataConnection", func() {
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError,
				derrors.DataModel_DataConnection_InternalError_CreateDataConnectionFailed)
			conn := interfaces.DataConnection{}

			patch1 := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateDataConnectionWhenCreate, nil)
			defer patch2.Reset()

			dcs.EXPECT().CreateDataConnection(gomock.Any(), gomock.Any()).Return("1", expectedHttpErr)

			reqParamByte, _ := sonic.Marshal(conn)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Create succeed", func() {
			conn := interfaces.DataConnection{}
			expectedConnID := "1"

			patch1 := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateDataConnectionWhenCreate, nil)
			defer patch2.Reset()

			dcs.EXPECT().CreateDataConnection(gomock.Any(), gomock.Any()).Return(expectedConnID, nil)

			reqParamByte, _ := sonic.Marshal(conn)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})
	})
}

func Test_DataConnectionRestHandler_DeleteDataConnections(t *testing.T) {
	Convey("Test DeleteDataConnections", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		dcs := dmock.NewMockDataConnectionService(mockCtrl)

		handler := MockNewDataConnectionRestHandler(appSetting, hydra, dcs)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/data-connections/1,2"

		Convey("Delete failed, because no invalid conn id was passed in", func() {
			url = "/api/mdl-data-model/v1/data-connections/,"
			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Delete failed, caused by the error from method GetMapAboutID2Name", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError,
				derrors.DataModel_DataConnection_InternalError_GetMapAboutID2NameFailed)

			dcs.EXPECT().GetMapAboutID2Name(gomock.Any(), gomock.Any()).Return(nil, expectedErr)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Delete failed, because some conns do not exist", func() {

			dcs.EXPECT().GetMapAboutID2Name(gomock.Any(), gomock.Any()).Return(nil, nil)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("Delete failed, caused by the error from method DeleteDataConnections", func() {
			expectedConnMap := map[string]string{
				"1": "conn1",
				"2": "conn2",
			}
			expectedErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_TraceModel_InternalError_DeleteTraceModelsFailed).
				WithErrorDetails("some errors")

			dcs.EXPECT().GetMapAboutID2Name(gomock.Any(), gomock.Any()).Return(expectedConnMap, nil)
			dcs.EXPECT().DeleteDataConnections(gomock.Any(), gomock.Any()).Return(expectedErr)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Delete succeed", func() {
			expectedConnMap := map[string]string{
				"1": "conn1",
				"2": "conn2",
			}

			dcs.EXPECT().GetMapAboutID2Name(gomock.Any(), gomock.Any()).Return(expectedConnMap, nil)
			dcs.EXPECT().DeleteDataConnections(gomock.Any(), gomock.Any()).Return(nil)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})
	})
}

func Test_DataConnectionRestHandler_UpdateDataConnection(t *testing.T) {
	Convey("Test UpdateDataConnection", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		dcs := dmock.NewMockDataConnectionService(mockCtrl)

		handler := MockNewDataConnectionRestHandler(appSetting, hydra, dcs)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/data-connections/1"

		reqConn := interfaces.DataConnection{
			Name: "test1",
		}

		Convey("Update failed, caused by the error from method 'GetDataConnection'", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_DataView_InternalError_GetSimpleDataViewMapByNamesFailed).
				WithErrorDetails("some errors")
			dcs.EXPECT().GetDataConnection(gomock.Any(), gomock.Any(), gomock.Any()).Return(&interfaces.DataConnection{}, false, expectedErr)

			reqParamByte, _ := sonic.Marshal(reqConn)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Update failed, caused by the conn is not exist", func() {
			dcs.EXPECT().GetDataConnection(gomock.Any(), gomock.Any(), gomock.Any()).Return(&interfaces.DataConnection{}, false, nil)

			reqParamByte, _ := sonic.Marshal(reqConn)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("Update failed, caused by the error from method ShouldBindJSON", func() {
			expectedConn := interfaces.DataConnection{}
			expectedErr := errors.New("some errors")

			dcs.EXPECT().GetDataConnection(gomock.Any(), gomock.Any(), gomock.Any()).Return(&expectedConn, true, nil)

			patch2 := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", expectedErr)
			defer patch2.Reset()

			reqParamByte, _ := sonic.Marshal(reqConn)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Update failed, caused by the error from func validateDataConnectionWhenUpdate", func() {
			expectedConn := interfaces.DataConnection{}
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_DataConnection_ForbiddenUpdateParameter_DataSourceType)

			dcs.EXPECT().GetDataConnection(gomock.Any(), gomock.Any(), gomock.Any()).Return(&expectedConn, true, nil)

			patch2 := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", nil)
			defer patch2.Reset()

			patch3 := ApplyFuncReturn(validateDataConnectionWhenUpdate, expectedErr)
			defer patch3.Reset()

			reqParamByte, _ := sonic.Marshal(reqConn)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Update failed, caused by the error from method UpdateDataConnection", func() {
			expectedConn := interfaces.DataConnection{}
			expectedErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_DataConnection_InternalError_UpdateDataConnectionFailed).
				WithErrorDetails("some errors")

			dcs.EXPECT().GetDataConnection(gomock.Any(), gomock.Any(), gomock.Any()).Return(&expectedConn, true, nil)

			patch2 := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", nil)
			defer patch2.Reset()

			patch3 := ApplyFuncReturn(validateDataConnectionWhenUpdate, nil)
			defer patch3.Reset()

			dcs.EXPECT().UpdateDataConnection(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedErr)

			reqParamByte, _ := sonic.Marshal(reqConn)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Update succeed", func() {
			expectedConn := interfaces.DataConnection{}

			dcs.EXPECT().GetDataConnection(gomock.Any(), gomock.Any(), gomock.Any()).Return(&expectedConn, true, nil)

			patch2 := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", nil)
			defer patch2.Reset()

			patch3 := ApplyFuncReturn(validateDataConnectionWhenUpdate, nil)
			defer patch3.Reset()

			dcs.EXPECT().UpdateDataConnection(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			reqParamByte, _ := sonic.Marshal(reqConn)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})
	})
}

func Test_DataConnectionRestHandler_GetDataConnection(t *testing.T) {
	Convey("Test GetDataConnection", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		dcs := dmock.NewMockDataConnectionService(mockCtrl)

		handler := MockNewDataConnectionRestHandler(appSetting, hydra, dcs)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/data-connections/1"

		Convey("Get failed, caused by the error from method GetDataConnection", func() {

			expectedErr := errors.New("some errors")
			dcs.EXPECT().GetDataConnection(gomock.Any(), gomock.Any(), gomock.Any()).Return(&interfaces.DataConnection{}, false, expectedErr)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Get failed, becaused the conn do not exist", func() {

			dcs.EXPECT().GetDataConnection(gomock.Any(), gomock.Any(), gomock.Any()).Return(&interfaces.DataConnection{}, false, nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("Get succeed", func() {
			dcs.EXPECT().GetDataConnection(gomock.Any(), gomock.Any(), gomock.Any()).Return(&interfaces.DataConnection{}, true, nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func Test_DataConnectionRestHandler_ListDataConnections(t *testing.T) {
	Convey("Test ListDataConnections", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		dcs := dmock.NewMockDataConnectionService(mockCtrl)

		handler := MockNewDataConnectionRestHandler(appSetting, hydra, dcs)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/data-connections"

		Convey("List failed, caused by the error from func validateNameandNamePattern", func() {
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_ConflictParameter_NameAndNamePatternCoexist).
				WithErrorDetails("Parameters name_pattern and name are passed in at the same time")
			patch := ApplyFuncReturn(validateNameandNamePattern, expectedHttpErr)
			defer patch.Reset()

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("List failed, caused by the error from func validatePaginationQueryParameters", func() {
			patch1 := ApplyFuncReturn(validateNameandNamePattern, nil)
			defer patch1.Reset()

			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_Offset).
				WithErrorDetails("some errors")
			patch2 := ApplyFuncReturn(validatePaginationQueryParameters, interfaces.PaginationQueryParameters{}, expectedHttpErr)
			defer patch2.Reset()

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("List failed, caused by the error from method ListDataConnections", func() {
			patch1 := ApplyFuncReturn(validateNameandNamePattern, nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validatePaginationQueryParameters, interfaces.PaginationQueryParameters{}, nil)
			defer patch2.Reset()

			expectedEntries := []*interfaces.DataConnectionListEntry{}
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError,
				derrors.DataModel_DataConnection_InternalError_ListDataConnectionsFailed).WithErrorDetails("some error")

			dcs.EXPECT().ListDataConnections(gomock.Any(), gomock.Any()).Return(expectedEntries, 0, expectedHttpErr)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("List succeed", func() {
			patch1 := ApplyFuncReturn(validateNameandNamePattern, nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validatePaginationQueryParameters, interfaces.PaginationQueryParameters{}, nil)
			defer patch2.Reset()

			expectedEntries := []*interfaces.DataConnectionListEntry{}
			dcs.EXPECT().ListDataConnections(gomock.Any(), gomock.Any()).Return(expectedEntries, 0, nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}
