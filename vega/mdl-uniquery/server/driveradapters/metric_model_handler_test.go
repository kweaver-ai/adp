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
	"os"
	"testing"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	rmock "github.com/kweaver-ai/kweaver-go-lib/rest/mock"
	. "github.com/smartystreets/goconvey/convey"

	"uniquery/common"
	cond "uniquery/common/condition"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	"uniquery/interfaces/data_type"
	umock "uniquery/interfaces/mock"
)

var (
	mmStart1 = int64(1678412100123)
	mmEnd1   = int64(1678412210123)

	StepsMap = map[string]string{
		"1m":  "1m",
		"5m":  "5m",
		"10m": "10m",
		"15m": "15m",
		"20m": "20m",
		"30m": "30m",
		"1h":  "1h",
		"2h":  "2h",
		"3h":  "3h",
		"6h":  "6h",
		"12h": "12h",
		"1d":  "1d",
	}
)

func mockNewMetricModelRestHandler(appSetting *common.AppSetting,
	hydra rest.Hydra, mmService interfaces.MetricModelService) (r *restHandler) {

	appSetting.ServerSetting = common.ServerSetting{
		IgnoringHcts: false,
	}
	r = &restHandler{
		appSetting: appSetting,
		hydra:      hydra,
		mmService:  mmService,
	}
	r.InitMetric()
	return r
}

func TestMetricModel(t *testing.T) {
	Convey("Test handler MetricModel", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		loc, _ := time.LoadLocation(os.Getenv("TZ"))
		common.APP_LOCATION = loc

		appSetting := &common.AppSetting{}
		hydraMock := rmock.NewMockHydra(mockCtrl)
		mmService := umock.NewMockMetricModelService(mockCtrl)
		handler := mockNewMetricModelRestHandler(appSetting, hydraMock, mmService)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/metric-model"

		// now := time.Now()
		metricModel := interfaces.MetricModelQuery{
			MetricType: "atomic",
			DataSource: &interfaces.MetricDataSource{
				Type: interfaces.SOURCE_TYPE_DATA_VIEW,
				ID:   "123",
			},
			QueryType: "promql",
			Formula:   "avg(rate(node_cpu_seconds_total{name=~\"node-14-35\",mode12=\"system\"}[5m]))by(name)",
			QueryTimeParams: interfaces.QueryTimeParams{
				Start:   &start1,
				End:     &end1,
				StepStr: &step_5m,
			},
		}
		reqParamByte, _ := sonic.Marshal(metricModel)

		common.FixedStepsMap = StepsMap

		Convey("Success MetricModel \n", func() {
			mmService.EXPECT().Simulate(gomock.Any(), gomock.Any()).Return(interfaces.MetricModelUniResponse{}, nil)

			// expected := `{"datas":null,"is_variable":false,"is_calendar":false,"overall_ms":2,"status_code":0}`

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			// compareJsonString(w.Body.String(), expected)
		})

		Convey("Failed MetricModel ShouldBind Error \n", func() {
			reqParamByte1, _ := sonic.Marshal([]interfaces.MetricModelQuery{metricModel})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte1))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Method is null\n", func() {

			reqParamByte1, _ := sonic.Marshal(interfaces.MetricModelQuery{})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte1))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			// req.Header.Set(interfaces.X_HTTP_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Method is not GET \n", func() {

			reqParamByte1, _ := sonic.Marshal(interfaces.MetricModelQuery{})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte1))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodPost)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MetricType is null \n", func() {

			reqParamByte1, _ := sonic.Marshal(interfaces.MetricModelQuery{})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte1))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Exec Failed\n", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				BaseError: rest.BaseError{
					ErrorCode: uerrors.Uniquery_MetricModel_InternalError_ExecPromQLFailed,
				},
			}

			mmService.EXPECT().Simulate(gomock.Any(), gomock.Any()).Return(interfaces.MetricModelUniResponse{}, expectedErr)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

	})
}

func TestGetMetricModelData(t *testing.T) {
	Convey("Test handler GetMetricModelData", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydraMock := rmock.NewMockHydra(mockCtrl)
		mmService := umock.NewMockMetricModelService(mockCtrl)
		handler := mockNewMetricModelRestHandler(appSetting, hydraMock, mmService)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/metric-models/1"

		metricModel := interfaces.MetricModelQuery{
			QueryTimeParams: interfaces.QueryTimeParams{
				Start:   &mmStart1,
				End:     &mmEnd1,
				StepStr: &step_5m,
			},
		}
		reqParamByte, _ := sonic.Marshal(metricModel)
		common.FixedStepsMap = StepsMap

		Convey("Success GetMetricModelData \n", func() {
			mmService.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(interfaces.MetricModelUniResponse{}, 0, 0, nil)

			// expected := `{"datas":null,"is_variable":false,"is_calendar":false,"status_code":0}`

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			// compareJsonString(w.Body.String(), expected)
		})

		Convey("Success MetricModel with multi model \n", func() {
			// expected := `[{"datas":null,"is_variable":false,"is_calendar":false,"status_code":0},{"datas":null,"is_variable":false,"is_calendar":false,"status_code":0}]`

			mmService.EXPECT().Exec(gomock.Any(), gomock.Any()).AnyTimes().Return(interfaces.MetricModelUniResponse{}, 0, 0, nil)
			reqParamBytes, _ := sonic.Marshal([]interfaces.MetricModelQuery{metricModel, metricModel})

			url = "/api/mdl-uniquery/v1/metric-models/1,2"
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamBytes))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			// compareJsonString(w.Body.String(), expected)
		})

		Convey("Method is null \n", func() {
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Method is not GET \n", func() {
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodPost)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Failed MetricModelQuery ShouldBind Error \n", func() {
			reqParamByte1, _ := sonic.Marshal([]interfaces.MetricModelQuery{metricModel})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte1))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Failed MetricModelQuery ShouldBind Error with multi model \n", func() {
			url = "/api/mdl-uniquery/v1/metric-models/1,2"
			reqParamByte1, _ := sonic.Marshal(metricModel)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte1))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Invalid start \n", func() {
			starti := int64(-1)
			stepi := "100ms"
			metricModel1 := interfaces.MetricModelQuery{
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &starti,
					End:     &mmEnd1,
					StepStr: &stepi,
				},
			}
			reqParamByte1, _ := sonic.Marshal(metricModel1)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte1))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Exec Failed\n", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				BaseError: rest.BaseError{
					ErrorCode: uerrors.Uniquery_MetricModel_InternalError_ExecPromQLFailed,
				},
			}

			mmService.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(interfaces.MetricModelUniResponse{}, 0, 0, expectedErr)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("ignoring_store_cache is invalid \n", func() {
			url = "/api/mdl-uniquery/v1/metric-models/1?ignoring_store_cache=a"
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("include_model is invalid \n", func() {
			url = "/api/mdl-uniquery/v1/metric-models/1?include_model=a"
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Failed MetricModelQuery ignoring_store_cache Error with multi model \n", func() {
			url = "/api/mdl-uniquery/v1/metric-models/1,2?ignoring_store_cache=a"
			reqParamByte1, _ := sonic.Marshal([]interfaces.MetricModelQuery{metricModel, metricModel})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte1))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Failed MetricModelQuery include_model Error with multi model \n", func() {
			url = "/api/mdl-uniquery/v1/metric-models/1,2?include_model=a"
			reqParamByte1, _ := sonic.Marshal([]interfaces.MetricModelQuery{metricModel, metricModel})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte1))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})
	})
}

func TestGetMetricModelFields(t *testing.T) {
	Convey("Test handler GetMetricModelFields", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydraMock := rmock.NewMockHydra(mockCtrl)
		mmService := umock.NewMockMetricModelService(mockCtrl)
		handler := mockNewMetricModelRestHandler(appSetting, hydraMock, mmService)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/metric-models/modelid1/fields"

		Convey("GetMetricModelFields success", func() {
			expectedFields := []interfaces.Field{
				{Name: "field1", Type: data_type.DataType_String},
				{Name: "field2", Type: data_type.DataType_String},
			}

			mmService.EXPECT().GetMetricModelFields(gomock.Any(), "modelid1").Return(expectedFields, nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("GetMetricModelFields internal error", func() {
			mmService.EXPECT().GetMetricModelFields(gomock.Any(), "modelid1").Return(
				nil, errors.New("internal error"))

			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("GetMetricModelFields empty model_id", func() {
			url = "/api/mdl-uniquery/v1/metric-models//fields"
			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})
	})
}

func TestGetMetricModelFieldValues(t *testing.T) {
	Convey("Test handler GetMetricModelFieldValues", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydraMock := rmock.NewMockHydra(mockCtrl)
		mmService := umock.NewMockMetricModelService(mockCtrl)
		handler := mockNewMetricModelRestHandler(appSetting, hydraMock, mmService)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/metric-models/modelid1/field_values/field1"

		Convey("GetMetricModelFieldValues success", func() {
			expectedValues := interfaces.FieldValues{
				Values: []string{"value1", "value2"},
			}

			mmService.EXPECT().GetMetricModelFieldValues(gomock.Any(), "modelid1", "field1").Return(expectedValues, nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("GetMetricModelFieldValues internal error", func() {
			mmService.EXPECT().GetMetricModelFieldValues(gomock.Any(), "modelid1", "field1").Return(
				interfaces.FieldValues{}, errors.New("internal error"))

			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("GetMetricModelFieldValues empty model_id", func() {
			url = "/api/mdl-uniquery/v1/metric-models//field_values/field1"
			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("GetMetricModelFieldValues empty field_name", func() {
			url = "/api/mdl-uniquery/v1/metric-models/modelid1/field_values/"
			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})
	})
}

func TestGetMetricModelLabels(t *testing.T) {
	Convey("Test handler GetMetricModelLabels", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydraMock := rmock.NewMockHydra(mockCtrl)
		mmService := umock.NewMockMetricModelService(mockCtrl)
		handler := mockNewMetricModelRestHandler(appSetting, hydraMock, mmService)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/metric-models/modelid1/labels"

		Convey("GetMetricModelLabels success", func() {
			expectedFields := []*cond.ViewField{
				{Name: "field1", Type: data_type.DataType_String},
				{Name: "field2", Type: data_type.DataType_String},
			}

			mmService.EXPECT().GetMetricModelLabels(gomock.Any(), "modelid1").Return(expectedFields, nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("GetMetricModelLabels internal error", func() {
			mmService.EXPECT().GetMetricModelLabels(gomock.Any(), "modelid1").Return(
				nil, errors.New("internal error"))

			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("GetMetricModelLabels empty model_id", func() {
			url = "/api/mdl-uniquery/v1/metric-models//labels"
			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})
	})
}
