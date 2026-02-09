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

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	rmock "github.com/kweaver-ai/kweaver-go-lib/rest/mock"
	. "github.com/smartystreets/goconvey/convey"

	"uniquery/common"
	"uniquery/interfaces"
	umock "uniquery/interfaces/mock"
)

var (
	objective99 float64 = 99
	period90    int64   = 90
)

func mockNewObjectiveModelRestHandler(hydra rest.Hydra, omService interfaces.ObjectiveModelService) (r *restHandler) {
	r = &restHandler{
		hydra:     hydra,
		omService: omService,
	}
	r.InitMetric()
	return r
}

func TestObjectiveSimulate(t *testing.T) {
	Convey("Test ObjectiveSimulate", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		hydraMock := rmock.NewMockHydra(mockCtl)
		mockOMService := umock.NewMockObjectiveModelService(mockCtl)
		handler := mockNewObjectiveModelRestHandler(hydraMock, mockOMService)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/objective-models"

		// now := time.Now()
		objectiveModel := interfaces.ObjectiveModelQuery{
			ObjectiveType: interfaces.SLO,
			ObjectiveConfig: interfaces.SLOObjective{
				Objective: &objective99,
				Period:    &period90,
				GoodMetricModel: &interfaces.BundleMetricModel{
					Id: "1",
				},
				TotalMetricModel: &interfaces.BundleMetricModel{
					Id: "2",
				},
			},
			QueryTimeParams: interfaces.QueryTimeParams{
				Start:   &start1,
				End:     &end1,
				StepStr: &step_5m,
			},
		}
		reqParamByte, _ := sonic.Marshal(objectiveModel)

		common.FixedStepsMap = StepsMap

		Convey("When simulation succeeds", func() {
			mockOMService.EXPECT().Simulate(gomock.Any(), gomock.Any()).Return(interfaces.ObjectiveModelUniResponse{}, nil)

			expected := `{"datas":null}`

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("When simulation fails due to invalid parameter binding", func() {
			reqParamByte1, _ := sonic.Marshal([]interfaces.ObjectiveModelQuery{objectiveModel})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte1))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("When simulation fails due to invalid method override", func() {
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, "INVALID")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("When simulation fails due to invalid objective model validation", func() {
			invalidModel := interfaces.ObjectiveModelQuery{
				ObjectiveType: "", // Invalid: empty objective type
				ObjectiveConfig: map[string]interface{}{
					"objective": 0, // Invalid: objective <= 0
				},
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &start1,
					End:     &end1,
					StepStr: &step_5m,
				},
			}
			reqParamByte, _ := sonic.Marshal(invalidModel)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("When simulation fails due to objective model service error", func() {
			mockOMService.EXPECT().Simulate(gomock.Any(), gomock.Any()).Return(interfaces.ObjectiveModelUniResponse{}, errors.New("simulation failed"))

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func TestGetObjectiveModelData(t *testing.T) {
	Convey("Test GetObjectiveModelData", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		hydraMock := rmock.NewMockHydra(mockCtl)
		mockOMService := umock.NewMockObjectiveModelService(mockCtl)
		handler := mockNewObjectiveModelRestHandler(hydraMock, mockOMService)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/objective-models/id1"

		// now := time.Now()
		objectiveModel := interfaces.ObjectiveModelQuery{
			QueryTimeParams: interfaces.QueryTimeParams{
				Start:   &start1,
				End:     &end1,
				StepStr: &step_5m,
			},
		}
		reqParamByte, _ := sonic.Marshal(objectiveModel)

		common.FixedStepsMap = StepsMap

		Convey("When get objective model data succeeds", func() {
			mockOMService.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(interfaces.ObjectiveModelUniResponse{}, nil)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("When get objective model data fails due to invalid request body", func() {
			invalidJson := []byte(`{invalid json}`)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(invalidJson))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("When get objective model data fails due to invalid method override header", func() {
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, "INVALID_METHOD")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("When get objective model data fails due to include_model invalid", func() {
			url = "/api/mdl-uniquery/v1/objective-models/id1?include_model=q"
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("When get objective model data fails due to invalid step", func() {
			stepi := "invalid_step"
			objectiveModel := interfaces.ObjectiveModelQuery{
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &start1,
					End:     &end1,
					StepStr: &stepi,
				},
			}
			reqParamByte, _ := sonic.Marshal(objectiveModel)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("When get objective model data fails due to service exec error", func() {
			mockOMService.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(interfaces.ObjectiveModelUniResponse{}, fmt.Errorf("exec error"))

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}
