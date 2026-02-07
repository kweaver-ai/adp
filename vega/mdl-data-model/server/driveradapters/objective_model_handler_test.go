// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	objective99        float64 = 99
	period90           int64   = 90
	testObjectiveModel         = interfaces.ObjectiveModel{
		ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
			ModelID:       "1",
			ModelName:     "16",
			ObjectiveType: interfaces.SLO,
			ObjectiveConfig: interfaces.SLOObjective{
				Objective: &objective99,
				Period:    &period90,
				GoodMetricModel: &interfaces.BundleMetricModel{
					ID: "1",
				},
				TotalMetricModel: &interfaces.BundleMetricModel{
					ID: "2",
				},
			},
			Tags:    []string{"a", "s", "s", "s", "s"},
			Comment: "ssss",
		},
		Task: &taskO,
	}

	taskO = interfaces.MetricTask{
		TaskID:          "1",
		TaskName:        "task1",
		ModuleType:      interfaces.MODULE_TYPE_OBJECTIVE_MODEL,
		ModelID:         "1",
		Steps:           []string{"1d"},
		Schedule:        interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
		IndexBase:       "base1",
		RetraceDuration: "1d",
	}
)

func MockNewObjectiveModelRestHandler(appSetting *common.AppSetting,
	hydra rest.Hydra,
	mms interfaces.MetricModelService,
	oms interfaces.ObjectiveModelService,
	mmts interfaces.MetricModelTaskService) (r *restHandler) {

	r = &restHandler{
		appSetting: appSetting,
		hydra:      hydra,
		mms:        mms,
		oms:        oms,
		mmts:       mmts,
	}
	return r
}

func Test_ObjectiveModelRestHandler_CreateObjectiveModels(t *testing.T) {
	Convey("Test ObjectiveModelHandler CreateObjectiveModels\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)
		oms := dmock.NewMockObjectiveModelService(mockCtrl)
		mmts := dmock.NewMockMetricModelTaskService(mockCtrl)

		handler := MockNewObjectiveModelRestHandler(appSetting, hydra, mms, oms, mmts)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		common.PersistStepsMap = StepsMap

		url := "/api/mdl-data-model/v1/objective-models"

		Convey("When creating objective model successfully", func() {
			oms.EXPECT().CheckObjectiveModelExistByID(gomock.Any(), gomock.Any()).AnyTimes().Return("123", false, nil)
			oms.EXPECT().CheckObjectiveModelExistByName(gomock.Any(), gomock.Any()).AnyTimes().Return("123", false, nil)
			oms.EXPECT().CreateObjectiveModels(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return([]string{"id1"}, nil)

			reqParamByte, _ := sonic.Marshal([]interfaces.ObjectiveModel{testObjectiveModel})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("When binding parameter failed", func() {
			reqParamByte := []byte(`[{"invalid": json}]`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("When objective model ID already exists in request body", func() {
			duplicateModel2 := interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelID:   "1", // Duplicate ID
					ModelName: "Model 2",
				},
			}

			oms.EXPECT().CheckObjectiveModelExistByID(gomock.Any(), gomock.Any()).AnyTimes().Return("123", false, nil)
			oms.EXPECT().CheckObjectiveModelExistByName(gomock.Any(), gomock.Any()).AnyTimes().Return("123", false, nil)

			reqParamByte, _ := sonic.Marshal([]interfaces.ObjectiveModel{testObjectiveModel, duplicateModel2})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("When validating objective model failed", func() {
			invalidModel := interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelID:   "1",
					ModelName: "", // Invalid empty name
				},
			}

			reqParamByte, _ := sonic.Marshal([]interfaces.ObjectiveModel{invalidModel})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("When objective model name already exists in request body", func() {
			duplicateModel := interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelID:   "2",
					ModelName: "16", // Same name as testObjectiveModel
				},
			}

			oms.EXPECT().CheckObjectiveModelExistByID(gomock.Any(), gomock.Any()).AnyTimes().Return("", false, nil)
			oms.EXPECT().CheckObjectiveModelExistByName(gomock.Any(), gomock.Any()).AnyTimes().Return("", false, nil)

			reqParamByte, _ := sonic.Marshal([]interfaces.ObjectiveModel{testObjectiveModel, duplicateModel})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		// Convey("When checking objective model existence by ID fails", func() {
		// 	err := rest.NewHTTPError(context.Background(), http.StatusInternalServerError,
		// 		derrors.DataModel_ObjectiveModel_InternalError_CheckModelIfExistFailed).WithErrorDetails(fmt.Errorf("database error"))

		// 	oms.EXPECT().CheckObjectiveModelExistByID(gomock.Any(), gomock.Any()).Return("", false, err)

		// 	reqParamByte, _ := sonic.Marshal([]interfaces.ObjectiveModel{testObjectiveModel})
		// 	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
		// 	req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
		// 	w := httptest.NewRecorder()
		// 	engine.ServeHTTP(w, req)

		// 	So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		// })

		// Convey("When objective model ID already exists in database", func() {
		// 	oms.EXPECT().CheckObjectiveModelExistByID(gomock.Any(), gomock.Any()).Return("existing-model", true, nil)
		// 	oms.EXPECT().CheckObjectiveModelExistByName(gomock.Any(), gomock.Any()).Return("", false, nil)

		// 	reqParamByte, _ := sonic.Marshal([]interfaces.ObjectiveModel{testObjectiveModel})
		// 	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
		// 	req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
		// 	w := httptest.NewRecorder()
		// 	engine.ServeHTTP(w, req)

		// 	So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		// })

		// Convey("When checking objective model existence by name fails", func() {
		// 	oms.EXPECT().CheckObjectiveModelExistByID(gomock.Any(), gomock.Any()).Return("", false, nil)
		// 	oms.EXPECT().CheckObjectiveModelExistByName(gomock.Any(), gomock.Any()).Return("", false, fmt.Errorf("database error"))

		// 	reqParamByte, _ := sonic.Marshal([]interfaces.ObjectiveModel{testObjectiveModel})
		// 	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
		// 	req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
		// 	w := httptest.NewRecorder()
		// 	engine.ServeHTTP(w, req)

		// 	So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		// })

		Convey("When creating objective model fails", func() {
			// oms.EXPECT().CheckObjectiveModelExistByID(gomock.Any(), gomock.Any()).Return("", false, nil)
			// oms.EXPECT().CheckObjectiveModelExistByName(gomock.Any(), gomock.Any()).Return("", false, nil)
			oms.EXPECT().CreateObjectiveModels(gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{}, fmt.Errorf("database error"))

			reqParamByte, _ := sonic.Marshal([]interfaces.ObjectiveModel{testObjectiveModel})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_ObjectiveModelRestHandler_ListObjectiveModels(t *testing.T) {
	Convey("Test ListObjectiveModels", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)
		oms := dmock.NewMockObjectiveModelService(mockCtrl)
		mmts := dmock.NewMockMetricModelTaskService(mockCtrl)

		handler := MockNewObjectiveModelRestHandler(appSetting, hydra, mms, oms, mmts)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		common.PersistStepsMap = StepsMap

		url := "/api/mdl-data-model/v1/objective-models"

		Convey("When listing objective models succeeds", func() {
			expectedModels := []interfaces.ObjectiveModel{
				{
					ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
						ModelID:   "test-id-1",
						ModelName: "test-model-1",
					},
				},
				{
					ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
						ModelID:   "test-id-2",
						ModelName: "test-model-2",
					},
				},
			}

			oms.EXPECT().ListObjectiveModels(gomock.Any(), gomock.Any()).Return(expectedModels, 2, nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("When pagination query validation fails", func() {
			url = url + "?offset=invalid"
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("When name_pattern and name parameters conflict", func() {
			url = url + "?name_pattern=test&name=test"
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)

			var response map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&response)

			So(err, ShouldBeNil)
			So(response["error_details"], ShouldEqual, "name_pattern and name cannot exists at the same time")
		})

		Convey("When listing objective models fails", func() {
			expectedError := rest.NewHTTPError(context.Background(), http.StatusInternalServerError, derrors.DataModel_ObjectiveModel_InternalError)
			oms.EXPECT().ListObjectiveModels(gomock.Any(), gomock.Any()).Return(nil, 0, expectedError)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_ObjectiveModelRestHandler_GetObjectiveModels(t *testing.T) {
	Convey("Test ObjectiveModelHandler GetObjectiveModels\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)
		oms := dmock.NewMockObjectiveModelService(mockCtrl)
		mmts := dmock.NewMockMetricModelTaskService(mockCtrl)

		handler := MockNewObjectiveModelRestHandler(appSetting, hydra, mms, oms, mmts)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/objective-models/test-id"

		Convey("When getting objective model succeeds", func() {
			expectedModel := interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelID:   "test-id",
					ModelName: "test-model",
				},
			}

			oms.EXPECT().GetObjectiveModels(gomock.Any(), []string{"test-id"}).Return([]interfaces.ObjectiveModel{expectedModel}, nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)

			var response []interfaces.ObjectiveModel
			err := json.NewDecoder(w.Body).Decode(&response)

			So(err, ShouldBeNil)
			So(response[0].ModelID, ShouldEqual, expectedModel.ModelID)
			So(response[0].ModelName, ShouldEqual, expectedModel.ModelName)
		})

		Convey("When getting objective model fails", func() {
			expectedError := rest.NewHTTPError(context.Background(), http.StatusInternalServerError, derrors.DataModel_ObjectiveModel_InternalError_GetMetricTasksByModelIDsFailed)
			oms.EXPECT().GetObjectiveModels(gomock.Any(), []string{"test-id"}).Return(nil, expectedError)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_ObjectiveModelRestHandler_UpdateObjectiveModel(t *testing.T) {
	Convey("Test UpdateObjectiveModel", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)
		oms := dmock.NewMockObjectiveModelService(mockCtrl)
		mmts := dmock.NewMockMetricModelTaskService(mockCtrl)

		handler := MockNewObjectiveModelRestHandler(appSetting, hydra, mms, oms, mmts)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		common.PersistStepsMap = StepsMap

		url := "/api/mdl-data-model/v1/objective-models/1"

		Convey("When updating objective model succeeds", func() {
			oms.EXPECT().CheckObjectiveModelExistByID(gomock.Any(), gomock.Any()).AnyTimes().Return("16", true, nil)

			oms.EXPECT().UpdateObjectiveModel(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			body, _ := sonic.Marshal(testObjectiveModel)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("When updating objective model fails due to invalid request body", func() {
			invalidJSON := []byte(`{"invalid": json}`)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(invalidJSON))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("When updating objective model fails due to validation error", func() {
			oms.EXPECT().CheckObjectiveModelExistByID(gomock.Any(), gomock.Any()).AnyTimes().Return("16", true, nil)

			// Create an invalid model that will fail validation
			invalidModel := testObjectiveModel
			invalidModel.ModelName = "" // Empty name will fail validation

			body, _ := sonic.Marshal(invalidModel)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("When updating objective model fails due to error getting model by ID", func() {
			oms.EXPECT().CheckObjectiveModelExistByID(gomock.Any(), gomock.Any()).Return("", false,
				rest.NewHTTPError(context.Background(), http.StatusInternalServerError, derrors.DataModel_ObjectiveModel_InternalError_CheckModelIfExistFailed))

			body, _ := sonic.Marshal(testObjectiveModel)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("When updating objective model fails because model does not exist", func() {
			oms.EXPECT().CheckObjectiveModelExistByID(gomock.Any(), gomock.Any()).Return("", false, nil)

			body, _ := sonic.Marshal(testObjectiveModel)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("When updating objective model fails due to error checking model name existence", func() {
			oms.EXPECT().CheckObjectiveModelExistByID(gomock.Any(), gomock.Any()).Return("old_name", true, nil)
			oms.EXPECT().CheckObjectiveModelExistByName(gomock.Any(), gomock.Any()).Return("", false,
				rest.NewHTTPError(context.Background(), http.StatusInternalServerError, derrors.DataModel_ObjectiveModel_InternalError_CheckModelIfExistFailed))

			// testObjectiveModel.ModelName = "new_name"
			body, _ := sonic.Marshal(testObjectiveModel)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("When updating objective model fails due to update error", func() {
			oms.EXPECT().CheckObjectiveModelExistByID(gomock.Any(), gomock.Any()).Return("old_name", true, nil)
			oms.EXPECT().CheckObjectiveModelExistByName(gomock.Any(), gomock.Any()).Return("", false, nil)
			oms.EXPECT().UpdateObjectiveModel(gomock.Any(), gomock.Any(), gomock.Any()).Return(
				rest.NewHTTPError(context.Background(), http.StatusInternalServerError, derrors.DataModel_ObjectiveModel_InternalError))

			body, _ := sonic.Marshal(testObjectiveModel)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_ObjectiveModelRestHandler_DeleteObjectiveModels(t *testing.T) {
	Convey("Test DeleteObjectiveModels", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)
		oms := dmock.NewMockObjectiveModelService(mockCtrl)
		mmts := dmock.NewMockMetricModelTaskService(mockCtrl)

		handler := MockNewObjectiveModelRestHandler(appSetting, hydra, mms, oms, mmts)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/objective-models/1,2"

		Convey("When deleting objective models succeeds", func() {
			oms.EXPECT().CheckObjectiveModelExistByID(gomock.Any(), gomock.Any()).AnyTimes().Return("model1", true, nil)
			oms.EXPECT().DeleteObjectiveModels(gomock.Any(), gomock.Any()).Return(int64(2), nil)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("When deleting objective models fails due to model check error", func() {
			oms.EXPECT().CheckObjectiveModelExistByID(gomock.Any(), gomock.Any()).Return("", false,
				rest.NewHTTPError(context.Background(), http.StatusInternalServerError, derrors.DataModel_ObjectiveModel_InternalError))

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("When deleting objective models fails because model does not exist", func() {
			oms.EXPECT().CheckObjectiveModelExistByID(gomock.Any(), gomock.Any()).Return("", false, nil)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("When deleting objective models fails due to delete error", func() {
			oms.EXPECT().CheckObjectiveModelExistByID(gomock.Any(), gomock.Any()).AnyTimes().Return("model1", true, nil)
			oms.EXPECT().DeleteObjectiveModels(gomock.Any(), gomock.Any()).Return(int64(0),
				rest.NewHTTPError(context.Background(), http.StatusInternalServerError, derrors.DataModel_ObjectiveModel_InternalError))

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}
