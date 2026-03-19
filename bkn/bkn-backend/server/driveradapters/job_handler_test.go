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
	"go.uber.org/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	"bkn-backend/common"
	berrors "bkn-backend/errors"
	"bkn-backend/interfaces"
	bmock "bkn-backend/interfaces/mock"
)

func MockNewJobRestHandler(appSetting *common.AppSetting,
	hydra hydra.Hydra,
	js interfaces.JobService,
	kns interfaces.KNService) (r *restHandler) {

	r = &restHandler{
		appSetting: appSetting,
		hydra:      hydra,
		js:         js,
		kns:        kns,
	}
	return r
}

func Test_JobRestHandler_CreateJob(t *testing.T) {
	Convey("Test JobHandler CreateJob\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydraMock := bmock.NewMockHydra(mockCtrl)
		js := bmock.NewMockJobService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewJobRestHandler(appSetting, hydraMock, js, kns)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/jobs"

		jobInfo := interfaces.JobInfo{
			Name:    "job1",
			Branch:  interfaces.MAIN_BRANCH,
			JobType: interfaces.JobTypeFull,
		}

		Convey("Success CreateJob \n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			js.EXPECT().CreateJob(gomock.Any(), gomock.Any()).Return("job1", nil)

			reqParamByte, _ := sonic.Marshal(jobInfo)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("Failed CreateJob ShouldBind Error\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)

			reqParamByte, _ := sonic.Marshal([]interfaces.JobInfo{jobInfo})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Job name is null\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)

			reqParamByte, _ := sonic.Marshal(interfaces.JobInfo{})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("KN not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, nil)

			reqParamByte, _ := sonic.Marshal(jobInfo)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("CheckKNExistByID failed\n", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_Job_InternalError,
				},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, expectedErr)

			reqParamByte, _ := sonic.Marshal(jobInfo)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("CreateJob failed\n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_Job_InternalError,
				},
			}

			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			js.EXPECT().CreateJob(gomock.Any(), gomock.Any()).Return("", err)

			reqParamByte, _ := sonic.Marshal(jobInfo)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_JobRestHandler_DeleteJobs(t *testing.T) {
	Convey("Test JobHandler DeleteJobs\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydraMock := bmock.NewMockHydra(mockCtrl)
		js := bmock.NewMockJobService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewJobRestHandler(appSetting, hydraMock, js, kns)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		jobIDs := "job1,job2"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/jobs/" + jobIDs

		Convey("Success DeleteJobs\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			js.EXPECT().GetJobsByIDs(gomock.Any(), gomock.Any()).Return(map[string]*interfaces.JobInfo{
				"job1": {ID: "job1", Name: "job1", KNID: knID, Branch: interfaces.MAIN_BRANCH},
				"job2": {ID: "job2", Name: "job2", KNID: knID, Branch: interfaces.MAIN_BRANCH},
			}, nil)
			js.EXPECT().DeleteJobsByIDs(gomock.Any(), knID, gomock.Any(), gomock.Any()).Return(nil)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("KN not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, nil)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("Job not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			js.EXPECT().GetJobsByIDs(gomock.Any(), gomock.Any()).Return(map[string]*interfaces.JobInfo{
				"job1": {ID: "job1", Name: "job1", KNID: knID, Branch: interfaces.MAIN_BRANCH},
			}, nil)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})
	})
}

func Test_JobRestHandler_ListJobs(t *testing.T) {
	Convey("Test JobHandler ListJobs\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydraMock := bmock.NewMockHydra(mockCtrl)
		js := bmock.NewMockJobService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewJobRestHandler(appSetting, hydraMock, js, kns)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/jobs"

		Convey("Success ListJobs\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			js.EXPECT().ListJobs(gomock.Any(), gomock.Any()).Return([]*interfaces.JobInfo{}, int64(0), nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("KN not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("Invalid pagination parameters\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)

			req := httptest.NewRequest(http.MethodGet, url+"?offset=invalid", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Invalid job_type\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)

			req := httptest.NewRequest(http.MethodGet, url+"?job_type=invalid", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Invalid state\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)

			req := httptest.NewRequest(http.MethodGet, url+"?state=invalid", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("ListJobs failed\n", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_Job_InternalError,
				},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			js.EXPECT().ListJobs(gomock.Any(), gomock.Any()).Return(nil, int64(0), expectedErr)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_JobRestHandler_ListTasks(t *testing.T) {
	Convey("Test JobHandler ListTasks\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydraMock := bmock.NewMockHydra(mockCtrl)
		js := bmock.NewMockJobService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewJobRestHandler(appSetting, hydraMock, js, kns)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		jobID := "job1"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/jobs/" + jobID + "/tasks"

		Convey("Success ListTasks\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			js.EXPECT().GetJobByID(gomock.Any(), jobID).Return(&interfaces.JobInfo{
				ID:     jobID,
				KNID:   knID,
				Branch: interfaces.MAIN_BRANCH,
			}, nil)
			js.EXPECT().ListTasks(gomock.Any(), gomock.Any()).Return([]*interfaces.TaskInfo{}, int64(0), nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("KN not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("Job not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			js.EXPECT().GetJobByID(gomock.Any(), jobID).Return(nil, nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("GetJobByID failed\n", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_Job_InternalError,
				},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			js.EXPECT().GetJobByID(gomock.Any(), jobID).Return(nil, expectedErr)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Invalid pagination parameters\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			js.EXPECT().GetJobByID(gomock.Any(), jobID).Return(&interfaces.JobInfo{
				ID:     jobID,
				KNID:   knID,
				Branch: interfaces.MAIN_BRANCH,
			}, nil)

			req := httptest.NewRequest(http.MethodGet, url+"?offset=invalid", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Invalid concept_type\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			js.EXPECT().GetJobByID(gomock.Any(), jobID).Return(&interfaces.JobInfo{
				ID:     jobID,
				KNID:   knID,
				Branch: interfaces.MAIN_BRANCH,
			}, nil)

			req := httptest.NewRequest(http.MethodGet, url+"?concept_type=invalid", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Invalid state\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			js.EXPECT().GetJobByID(gomock.Any(), jobID).Return(&interfaces.JobInfo{
				ID:     jobID,
				KNID:   knID,
				Branch: interfaces.MAIN_BRANCH,
			}, nil)

			req := httptest.NewRequest(http.MethodGet, url+"?state=invalid", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("ListTasks failed\n", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_Job_InternalError,
				},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			js.EXPECT().GetJobByID(gomock.Any(), jobID).Return(&interfaces.JobInfo{
				ID:     jobID,
				KNID:   knID,
				Branch: interfaces.MAIN_BRANCH,
			}, nil)
			js.EXPECT().ListTasks(gomock.Any(), gomock.Any()).Return(nil, int64(0), expectedErr)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}
