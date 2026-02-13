// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

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

var (
	testMetricUpdateTime = int64(1735786555379)
)

func MockNewMetricModelGroupRestHandler(appSetting *common.AppSetting,
	hydra rest.Hydra,
	mms interfaces.MetricModelService,
	dvs interfaces.DataViewService,
	mmgs interfaces.MetricModelGroupService,
	mmts interfaces.MetricModelTaskService) (r *restHandler) {

	r = &restHandler{
		appSetting: appSetting,
		hydra:      hydra,
		mms:        mms,
		dvs:        dvs,
		mmgs:       mmgs,
		mmts:       mmts,
	}
	return r
}

func Test_MetricModelGroupRestHandler_CreateMetricModelGroup(t *testing.T) {
	Convey("Test MetricModelHandler CreateMetricModelGroup\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		mmgs := dmock.NewMockMetricModelGroupService(mockCtrl)
		mmts := dmock.NewMockMetricModelTaskService(mockCtrl)

		handler := MockNewMetricModelGroupRestHandler(appSetting, hydra, mms, dvs, mmgs, mmts)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/metric-model-groups"

		metricModelGroup := interfaces.MetricModelGroup{
			GroupName: "group1",
			Comment:   "111",
		}

		Convey("Success CreateMetricModelGroup \n", func() {
			mmgs.EXPECT().CheckMetricModelGroupExist(gomock.Any(), gomock.Any()).Return(false, nil)
			mmgs.EXPECT().CreateMetricModelGroup(gomock.Any(), gomock.Any()).Return("1", nil)

			reqParamByte, _ := sonic.Marshal(metricModelGroup)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("Failed CreateMetricModels ShouldBind Error\n", func() {
			reqParamByte, _ := sonic.Marshal([]interfaces.MetricModelGroup{metricModelGroup})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Group name is null\n", func() {
			reqParamByte, _ := sonic.Marshal(interfaces.MetricModelGroup{})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Group name length exceeded \n", func() {
			reqParamByte, _ := sonic.Marshal(interfaces.MetricModelGroup{
				GroupName: "111111111111111111111111111111111111111111111111",
			})

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})
		Convey("Invalid comment\n", func() {
			reqParamByte, _ := sonic.Marshal(interfaces.MetricModelGroup{
				GroupName: "group1",
				Comment: `ddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd
				ddddddddddddddddsdddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddeeeeeeeedddddd
				ddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddffffffffffffffffffffffffffffffffffffffffffffffff`,
			})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Group name exists \n", func() {
			existGroupErr := &rest.HTTPError{
				HTTPCode: http.StatusBadRequest,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: derrors.DataModel_MetricModelGroup_GroupNameExisted,
				},
			}
			reqParamByte, _ := sonic.Marshal(metricModelGroup)
			mmgs.EXPECT().CheckMetricModelGroupExist(gomock.Any(), gomock.Any()).AnyTimes().Return(true, existGroupErr)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("CheckMetricModelGroupExistByName failed \n", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: derrors.DataModel_MetricModelGroup_InternalError_CheckGroupIfExistFailed,
				},
			}
			reqParamByte, _ := sonic.Marshal(metricModelGroup)
			mmgs.EXPECT().CheckMetricModelGroupExist(gomock.Any(), gomock.Any()).AnyTimes().Return(false, expectedErr)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Create model group failed\n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: derrors.DataModel_MetricModelGroup_InternalError,
				},
			}

			mmgs.EXPECT().CheckMetricModelGroupExist(gomock.Any(), gomock.Any()).AnyTimes().Return(false, nil)
			mmgs.EXPECT().CreateMetricModelGroup(gomock.Any(), gomock.Any()).AnyTimes().Return("", err)

			reqParamByte, _ := sonic.Marshal(metricModelGroup)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_MetricModelGroupRestHandler_UpdateMetricModelGroup(t *testing.T) {
	Convey("Test MetricModelHandler UpdateMetricModelGroup\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		mmgs := dmock.NewMockMetricModelGroupService(mockCtrl)
		mmts := dmock.NewMockMetricModelTaskService(mockCtrl)

		handler := MockNewMetricModelGroupRestHandler(appSetting, hydra, mms, dvs, mmgs, mmts)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/metric-model-groups/1"

		metricModelGroup := interfaces.MetricModelGroup{
			GroupName: "group1",
			Comment:   "111",
		}
		oldMetricModelGroup := interfaces.MetricModelGroup{
			GroupID:    "1",
			GroupName:  "groupOld",
			Comment:    "111",
			UpdateTime: testMetricUpdateTime,
		}

		Convey("Success UpdateMetricModelGroup \n", func() {
			mmgs.EXPECT().GetMetricModelGroupByID(gomock.Any(), gomock.Any()).Return(oldMetricModelGroup, nil)
			mmgs.EXPECT().CheckMetricModelGroupExist(gomock.Any(), gomock.Any()).Return(false, nil)
			mmgs.EXPECT().UpdateMetricModelGroup(gomock.Any(), gomock.Any()).Return(nil)

			reqParamByte, _ := sonic.Marshal(metricModelGroup)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("Failed UpdateMetricModelGroup ShouldBind Error\n", func() {
			reqParamByte, _ := sonic.Marshal([]interfaces.MetricModelGroup{metricModelGroup})
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("GetMetricModelGroupByID failed \n", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusNotFound,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: derrors.DataModel_MetricModelGroup_GroupNotFound,
				},
			}
			reqParamByte, _ := sonic.Marshal(metricModelGroup)
			mmgs.EXPECT().GetMetricModelGroupByID(gomock.Any(), gomock.Any()).AnyTimes().Return(interfaces.MetricModelGroup{}, expectedErr)

			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("Group name length exceeded \n", func() {
			reqParamByte, _ := sonic.Marshal(interfaces.MetricModelGroup{
				GroupName: "111111111111111111111111111111111111111111111111",
			})
			mmgs.EXPECT().GetMetricModelGroupByID(gomock.Any(), gomock.Any()).AnyTimes().Return(oldMetricModelGroup, nil)

			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Invalid comment\n", func() {
			reqParamByte, _ := sonic.Marshal(interfaces.MetricModelGroup{
				GroupName: "group1",
				Comment: `ddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd
				ddddddddddddddddsdddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddeeeeeeeedddddd
				ddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddffffffffffffffffffffffffffffffffffffffffffffffff`,
			})
			mmgs.EXPECT().GetMetricModelGroupByID(gomock.Any(), gomock.Any()).AnyTimes().Return(oldMetricModelGroup, nil)

			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("CheckMetricModelGroupExistByName failed \n", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: derrors.DataModel_MetricModelGroup_InternalError_CheckGroupIfExistFailed,
				},
			}

			reqParamByte, _ := sonic.Marshal(metricModelGroup)
			mmgs.EXPECT().GetMetricModelGroupByID(gomock.Any(), gomock.Any()).Return(oldMetricModelGroup, nil)
			mmgs.EXPECT().CheckMetricModelGroupExist(gomock.Any(), gomock.Any()).AnyTimes().Return(false, expectedErr)

			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Group name exists \n", func() {
			existGroupErr := &rest.HTTPError{
				HTTPCode: http.StatusBadRequest,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: derrors.DataModel_MetricModelGroup_GroupNameExisted,
				},
			}
			reqParamByte, _ := sonic.Marshal(metricModelGroup)
			mmgs.EXPECT().GetMetricModelGroupByID(gomock.Any(), gomock.Any()).Return(oldMetricModelGroup, nil)
			mmgs.EXPECT().CheckMetricModelGroupExist(gomock.Any(), gomock.Any()).AnyTimes().Return(true, existGroupErr)

			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Update model group failed\n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: derrors.DataModel_MetricModelGroup_InternalError,
				},
			}
			mmgs.EXPECT().GetMetricModelGroupByID(gomock.Any(), gomock.Any()).Return(oldMetricModelGroup, nil)
			mmgs.EXPECT().CheckMetricModelGroupExist(gomock.Any(), gomock.Any()).AnyTimes().Return(false, nil)
			mmgs.EXPECT().UpdateMetricModelGroup(gomock.Any(), gomock.Any()).AnyTimes().Return(err)

			reqParamByte, _ := sonic.Marshal(metricModelGroup)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_MetricModelGroupRestHandler_ListMetricModelGroups(t *testing.T) {
	Convey("Test MetricModelHandler ListMetricModelGroups\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		mmgs := dmock.NewMockMetricModelGroupService(mockCtrl)
		mmts := dmock.NewMockMetricModelTaskService(mockCtrl)

		handler := MockNewMetricModelGroupRestHandler(appSetting, hydra, mms, dvs, mmgs, mmts)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/metric-model-groups"

		metricModelGroup := interfaces.MetricModelGroup{
			GroupID:          "1",
			GroupName:        "group1",
			Comment:          "111",
			UpdateTime:       testMetricUpdateTime,
			MetricModelCount: 10,
		}

		Convey("Success ListMetricModelGroups \n", func() {
			mmgs.EXPECT().ListMetricModelGroups(gomock.Any(), gomock.Any()).Return([]*interfaces.MetricModelGroup{&metricModelGroup}, 1, nil)

			req := httptest.NewRequest(http.MethodGet, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("Invalid PaginationQueryParameters \n", func() {
			url = url + "?limit=-1&offset=aa"
			req := httptest.NewRequest(http.MethodGet, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("ListMetricModelGroups failed \n", func() {
			httpErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: derrors.DataModel_MetricModelGroup_InvalidParameter,
				},
			}
			mmgs.EXPECT().ListMetricModelGroups(gomock.Any(), gomock.Any()).Return([]*interfaces.MetricModelGroup{}, 0, httpErr)

			req := httptest.NewRequest(http.MethodGet, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_MetricModelGroupRestHandler_DeleteMetricModelGroup(t *testing.T) {
	Convey("Test MetricModelHandler DeleteMetricModelGroup\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		mmgs := dmock.NewMockMetricModelGroupService(mockCtrl)
		mmts := dmock.NewMockMetricModelTaskService(mockCtrl)

		handler := MockNewMetricModelGroupRestHandler(appSetting, hydra, mms, dvs, mmgs, mmts)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/metric-model-groups/1"

		metricModelGroup := interfaces.MetricModelGroup{
			GroupID:    "1",
			GroupName:  "groupOld",
			Comment:    "111",
			UpdateTime: testMetricUpdateTime,
		}

		Convey("Success DeleteMetricModelGroup \n", func() {
			mmgs.EXPECT().GetMetricModelGroupByID(gomock.Any(), gomock.Any()).Return(metricModelGroup, nil)
			mmgs.EXPECT().DeleteMetricModelGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), []interfaces.MetricModel{}, nil)

			url = url + "?force=false"
			req := httptest.NewRequest(http.MethodDelete, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("Invalid Parameter Force \n", func() {
			url = "/api/mdl-data-model/v1/metric-model-groups/1?force=11"
			req := httptest.NewRequest(http.MethodDelete, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("GetMetricModelGroupByID failed \n", func() {
			httpErr := &rest.HTTPError{
				HTTPCode: http.StatusNotFound,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: derrors.DataModel_MetricModelGroup_GroupNotFound,
				},
			}
			mmgs.EXPECT().GetMetricModelGroupByID(gomock.Any(), gomock.Any()).Return(interfaces.MetricModelGroup{}, httpErr)
			req := httptest.NewRequest(http.MethodDelete, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("DeleteMetricModelGroupAndModels failed  and force is true \n", func() {
			httpErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: derrors.DataModel_MetricModelGroup_InternalError,
				},
			}
			mmgs.EXPECT().GetMetricModelGroupByID(gomock.Any(), gomock.Any()).Return(metricModelGroup, nil)
			mmgs.EXPECT().DeleteMetricModelGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), []interfaces.MetricModel{}, httpErr)

			url = url + "?force=true"
			req := httptest.NewRequest(http.MethodDelete, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}
