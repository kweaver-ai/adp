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

func MockNewDataViewGroupRestHandler(appSetting *common.AppSetting,
	hydra rest.Hydra,
	dvgs interfaces.DataViewGroupService,
	dvs interfaces.DataViewService) (r *restHandler) {
	r = &restHandler{
		appSetting: appSetting,
		hydra:      hydra,
		dvgs:       dvgs,
		dvs:        dvs,
	}
	return r
}

func Test_DataViewGroupRestHandler_CreateDataViewGroup(t *testing.T) {
	Convey("Test CreateDataViewGroup", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		dvgs := dmock.NewMockDataViewGroupService(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)

		handler := MockNewDataViewGroupRestHandler(appSetting, hydra, dvgs, dvs)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/data-view-groups"

		reqGroup := interfaces.DataViewGroup{
			GroupName: "test1",
		}

		group := &interfaces.DataViewGroup{}

		Convey("Create failed, caused by the error from method ShouldBindJSON", func() {
			expectedErr := errors.New("some errors")
			patch := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", expectedErr)
			defer patch.Reset()

			reqParamByte, _ := sonic.Marshal(reqGroup)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Create failed, caused by the error from func ValidateDataViewGroup", func() {
			reqGroupTest := interfaces.DataViewGroup{
				GroupName: "test1/",
			}

			reqParamByte, _ := sonic.Marshal(reqGroupTest)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Create failed, caused by check group name exist failed", func() {
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_DataViewGroup_InternalError_CheckGroupExistByNameFailed)
			dvgs.EXPECT().CheckDataViewGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(group, false, expectedHttpErr)

			reqParamByte, _ := sonic.Marshal(reqGroup)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Create failed, caused by data view group name is existed", func() {
			dvgs.EXPECT().CheckDataViewGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(group, true, nil)

			reqParamByte, _ := sonic.Marshal(reqGroup)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("Create failed, caused by the error from method CreateDataViewGroup", func() {
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_DataViewGroup_InternalError_CreateGroupFailed)
			dvgs.EXPECT().CheckDataViewGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(group, false, nil)
			dvgs.EXPECT().CreateDataViewGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return("", expectedHttpErr)

			reqParamByte, _ := sonic.Marshal(reqGroup)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Create succeed", func() {
			dvgs.EXPECT().CheckDataViewGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(group, false, nil)
			dvgs.EXPECT().CreateDataViewGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return("1", nil)

			reqParamByte, _ := sonic.Marshal(reqGroup)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})
	})
}

func Test_DataViewGroupRestHandler_DeleteDataViewGroup(t *testing.T) {
	Convey("Test DeleteDataViewGroup", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		dvgs := dmock.NewMockDataViewGroupService(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)

		handler := MockNewDataViewGroupRestHandler(appSetting, hydra, dvgs, dvs)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/data-view-groups/1a"

		Convey("Delete failed, caused by the error from parse bool parameter", func() {
			url += "?delete_views=trueaaa"

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Delete failed, caused by the error from method GetDataViewGroupByID", func() {
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusNotFound, derrors.DataModel_DataViewGroup_GroupNotFound)
			dvgs.EXPECT().GetDataViewGroupByID(gomock.Any(), gomock.Any()).Return(nil, expectedHttpErr)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("Delete failed, caused by the error from method DeleteDataViewGroup", func() {
			group := &interfaces.DataViewGroup{
				GroupID:   "1a",
				GroupName: "x",
			}

			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_DataViewGroup_InternalError_DeleteGroupFailed)
			dvgs.EXPECT().GetDataViewGroupByID(gomock.Any(), gomock.Any()).Return(group, nil)
			dvgs.EXPECT().DeleteDataViewGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, expectedHttpErr)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Delete succeed", func() {
			group := &interfaces.DataViewGroup{
				GroupID:   "1a",
				GroupName: "x",
			}
			dvgs.EXPECT().GetDataViewGroupByID(gomock.Any(), gomock.Any()).Return(group, nil)
			dvgs.EXPECT().DeleteDataViewGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})
	})
}

func Test_DataViewGroupRestHandler_UpdateDataViewGroup(t *testing.T) {
	Convey("Test UpdateDataViewGroup", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		dvgs := dmock.NewMockDataViewGroupService(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)

		handler := MockNewDataViewGroupRestHandler(appSetting, hydra, dvgs, dvs)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/data-view-groups/1a"

		reqGroup := interfaces.DataViewGroup{
			GroupName: "test1",
		}

		oldGroup := &interfaces.DataViewGroup{
			GroupID:   "1a",
			GroupName: "x",
		}

		group := &interfaces.DataViewGroup{}

		Convey("Update failed, caused by the error from method ShouldBindJSON", func() {
			reqParamByte, _ := sonic.Marshal(`{"name": true}`)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Update failed, caused by the error from func ValidateGroupName", func() {
			dvgs.EXPECT().GetDataViewGroupByID(gomock.Any(), gomock.Any()).Return(oldGroup, nil)

			reqGroupTest := interfaces.DataViewGroup{
				GroupName: "test1/a/x",
			}
			reqParamByte, _ := sonic.Marshal(reqGroupTest)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Update failed, caused by the error from method GetDataViewGroupByID", func() {
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusNotFound, derrors.DataModel_DataViewGroup_GroupNotFound)
			dvgs.EXPECT().GetDataViewGroupByID(gomock.Any(), gomock.Any()).Return(nil, expectedHttpErr)

			reqParamByte, _ := sonic.Marshal(reqGroup)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("Update failed, cased by check group name exist failed", func() {
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_DataViewGroup_InternalError_CheckGroupExistByNameFailed)
			dvgs.EXPECT().GetDataViewGroupByID(gomock.Any(), gomock.Any()).Return(oldGroup, nil)
			dvgs.EXPECT().CheckDataViewGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(group, false, expectedHttpErr)

			reqParamByte, _ := sonic.Marshal(reqGroup)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Update failed, caused by the group name already exists", func() {
			dvgs.EXPECT().GetDataViewGroupByID(gomock.Any(), gomock.Any()).Return(oldGroup, nil)
			dvgs.EXPECT().CheckDataViewGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(group, true, nil)

			reqParamByte, _ := sonic.Marshal(reqGroup)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("Update failed, caused by the error from method UpdateDataViewGroup", func() {
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_DataViewGroup_InternalError_UpdateGroupFailed)
			dvgs.EXPECT().GetDataViewGroupByID(gomock.Any(), gomock.Any()).Return(oldGroup, nil)
			dvgs.EXPECT().CheckDataViewGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(group, false, nil)
			dvgs.EXPECT().UpdateDataViewGroup(gomock.Any(), gomock.Any()).Return(expectedHttpErr)

			reqParamByte, _ := sonic.Marshal(reqGroup)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Update succeed", func() {
			dvgs.EXPECT().GetDataViewGroupByID(gomock.Any(), gomock.Any()).Return(oldGroup, nil)
			dvgs.EXPECT().CheckDataViewGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(group, false, nil)
			dvgs.EXPECT().UpdateDataViewGroup(gomock.Any(), gomock.Any()).Return(nil)

			reqParamByte, _ := sonic.Marshal(reqGroup)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})
	})
}

func Test_DataViewGroupRestHandler_ListDataViewGroups(t *testing.T) {
	Convey("Test ListDataViewGroups", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		dvgs := dmock.NewMockDataViewGroupService(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)

		handler := MockNewDataViewGroupRestHandler(appSetting, hydra, dvgs, dvs)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/data-view-groups"

		Convey("List failed, caused by the error from func validatePaginationQueryParameters", func() {
			url = url + "?direction=desc&sort=update_time&limit=10&offset=a"
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("List failed, caused by the error from method ListDataViewGroups", func() {
			url = url + "?direction=desc&sort=update_time&limit=10&offset=0"
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_DataViewGroup_InternalError_ListGroupsFailed)

			dvgs.EXPECT().ListDataViewGroups(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, 0, expectedHttpErr)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("List succeed", func() {
			url = url + "?direction=desc&sort=update_time&limit=10&offset=0"
			expectedEntries := []*interfaces.DataViewGroup{}
			dvgs.EXPECT().ListDataViewGroups(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedEntries, 0, nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func Test_DataViewGroupRestHandler_GetDataViewsInGroup(t *testing.T) {
	Convey("Test GetDataViewsInGroup", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		dvgs := dmock.NewMockDataViewGroupService(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)

		handler := MockNewDataViewGroupRestHandler(appSetting, hydra, dvgs, dvs)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/data-view-groups/1a/data-views"

		Convey("Get failed, caused by the error from method GetDataViewGroupByID", func() {
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusNotFound, derrors.DataModel_DataViewGroup_GroupNotFound)
			dvgs.EXPECT().GetDataViewGroupByID(gomock.Any(), gomock.Any()).Return(nil, expectedHttpErr)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("Get failed, caused by the error from method GetDataViewsInGroup", func() {
			dvgs.EXPECT().GetDataViewGroupByID(gomock.Any(), gomock.Any()).Return(nil, nil)

			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_DataView_InternalError_GetDataViewsFailed)
			dvs.EXPECT().GetDataViewsByGroupID(gomock.Any(), gomock.Any()).Return(nil, expectedHttpErr)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Get succeed", func() {
			group := &interfaces.DataViewGroup{
				GroupID:   "1a",
				GroupName: "x",
			}

			views := []*interfaces.DataView{
				{
					SimpleDataView: interfaces.SimpleDataView{
						ViewID:   "1",
						ViewName: "x",
					},
				},
			}
			dvgs.EXPECT().GetDataViewGroupByID(gomock.Any(), gomock.Any()).Return(group, nil)
			dvs.EXPECT().GetDataViewsByGroupID(gomock.Any(), gomock.Any()).Return(views, nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}
