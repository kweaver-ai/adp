// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/agiledragon/gomonkey/v2"
	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"data-model-job/common"
	derrors "data-model-job/errors"
	"data-model-job/interfaces"
	dmock "data-model-job/interfaces/mock"
)

var (
	testCtx = context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage)
)

func mockNewJobRestHandler(appSetting *common.AppSetting,
	jService interfaces.JobService) (r *restHandler) {

	r = &restHandler{
		appSetting: appSetting,
		jobService: jService,
	}
	return r
}

func setGinMode() func() {
	old := gin.Mode()
	gin.SetMode(gin.TestMode)
	return func() {
		gin.SetMode(old)
	}
}

func Test_JobRestHandler_StartJob(t *testing.T) {
	Convey("Test StartJob", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		appSetting := &common.AppSetting{}
		dvService := dmock.NewMockJobService(mockCtl)
		handler := mockNewJobRestHandler(appSetting, dvService)
		handler.RegisterPublic(engine)

		url := "/api/mdl-data-model-job/v1/jobs"

		Convey("Create failed, caused by the error from method ShouldBindJSON", func() {
			var c *gin.Context
			expectedErr := errors.New("some errors")
			patch := ApplyMethodReturn(c, "ShouldBindJSON", expectedErr)
			defer patch.Reset()

			body := interfaces.JobInfo{
				JobId: "1a",
			}
			reqParamByte, _ := sonic.Marshal(body)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Create failed, caused by the error from method StartJob", func() {
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModelJob_InternalError_StartJobFailed)
			dvService.EXPECT().StartJob(gomock.Any(), gomock.Any()).Return(expectedHttpErr)

			body := interfaces.JobInfo{
				JobId: "1a",
			}
			reqParamByte, _ := sonic.Marshal(body)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Create succeed", func() {
			dvService.EXPECT().StartJob(gomock.Any(), gomock.Any()).Return(nil)
			body := interfaces.JobInfo{
				JobId:   "1a",
				JobType: interfaces.DATA_VIEW,
			}
			reqParamByte, _ := sonic.Marshal(body)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})
	})
}

func Test_JobRestHandler_UpdateJob(t *testing.T) {
	Convey("Test UpdateJob", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		appSetting := &common.AppSetting{}
		dvService := dmock.NewMockJobService(mockCtl)
		handler := mockNewJobRestHandler(appSetting, dvService)
		handler.RegisterPublic(engine)

		url := "/api/mdl-data-model-job/v1/jobs/1"

		Convey("Update failed, caused by the error from method ShouldBindJSON", func() {
			var c *gin.Context
			expectedErr := errors.New("some errors")
			patch := ApplyMethodReturn(c, "ShouldBindJSON", expectedErr)
			defer patch.Reset()

			body := interfaces.JobInfo{
				JobId: "1a",
			}
			reqParamByte, _ := sonic.Marshal(body)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Update failed, caused by the error from method UpdateJob", func() {
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModelJob_InternalError_UpdateJobFailed)
			dvService.EXPECT().UpdateJob(gomock.Any(), gomock.Any()).Return(expectedHttpErr)

			body := interfaces.JobInfo{
				JobId: "1a",
			}
			reqParamByte, _ := sonic.Marshal(body)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Update succeed", func() {
			dvService.EXPECT().UpdateJob(gomock.Any(), gomock.Any()).Return(nil)
			body := interfaces.JobInfo{
				JobId:   "1a",
				JobType: interfaces.DATA_VIEW,
			}
			reqParamByte, _ := sonic.Marshal(body)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})
	})
}

func Test_JobRestHandler_StopJob(t *testing.T) {
	Convey("Test StopJob", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		appSetting := &common.AppSetting{}
		dvService := dmock.NewMockJobService(mockCtl)
		handler := mockNewJobRestHandler(appSetting, dvService)
		handler.RegisterPublic(engine)

		url := "/api/mdl-data-model-job/v1/jobs/1a"

		Convey("Delete failed, caused by the error from method StopJob", func() {
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModelJob_InternalError_StopJobFailed)
			dvService.EXPECT().StopJob(gomock.Any(), gomock.Any()).Return(expectedHttpErr)

			body := interfaces.JobInfo{
				JobId: "1a",
			}
			reqParamByte, _ := sonic.Marshal(body)
			req := httptest.NewRequest(http.MethodDelete, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Delete succeed", func() {
			dvService.EXPECT().StopJob(gomock.Any(), gomock.Any()).Return(nil)
			body := interfaces.JobInfo{
				JobId:   "1a",
				JobType: interfaces.DATA_VIEW,
			}
			reqParamByte, _ := sonic.Marshal(body)
			req := httptest.NewRequest(http.MethodDelete, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})
	})
}
