// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	rmock "github.com/kweaver-ai/kweaver-go-lib/rest/mock"
	. "github.com/smartystreets/goconvey/convey"

	"uniquery/common"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	umock "uniquery/interfaces/mock"
)

var (
	mat    = "[]"
	series = `{
		"status": "success",
		"data": [
			{
				"labels.cpu": "0",
				"labels.instance": "instance_abc",
				"labels.job": "prometheus",
				"labels.mode": "nice"
			}
		]
	}`
	moreThanOneMatchSeries = `{
		"status": "success",
		"data": [
			{
				"labels.cpu": "0",
				"labels.instance": "instance_abc",
				"labels.job": "prometheus",
				"labels.mode": "nice"
			},
			{
				"labels.device": "tmpfs",
				"labels.fstype": "tmpfs",
				"labels.instance": "instance_abc",
				"labels.job": "prometheus",
				"labels.mountpoint": "/run/user/0"
			}
		]
	}`
)

func mockNewPromqlRestHandler(appSetting *common.AppSetting,
	hydra rest.Hydra, promqlService interfaces.PromQLService) (r *restHandler) {
	r = &restHandler{
		appSetting:    appSetting,
		hydra:         hydra,
		promqlService: promqlService,
	}
	r.InitMetric()
	return r
}

func TestPromqlQueryRange(t *testing.T) {
	Convey("Test handler PromqlQueryRange", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydraMock := rmock.NewMockHydra(mockCtrl)
		promqlService := umock.NewMockPromQLService(mockCtrl)
		handler := mockNewPromqlRestHandler(appSetting, hydraMock, promqlService)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/promql/query_range"

		Convey("Success PromqlQueryRange \n", func() {
			expected, _ := sonic.Marshal(mat)
			promqlService.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(interfaces.PromQLResponse{}, expected, http.StatusOK, nil)

			body := []byte(`query=prometheus.metrics.opensearch_indices_shards_docs` +
				`{prometheus.labels.cluster="opensearch",prometheus.labels.index=~"a.*|node.*"}` +
				`&start=1652320500&end=1652320620&step=10s`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), string(expected))
		})

		Convey("Success PromqlQueryRange with macthers value contain \\. \n", func() {
			expected, _ := sonic.Marshal(mat)
			promqlService.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(interfaces.PromQLResponse{}, expected, http.StatusOK, nil)

			body := []byte(`query=prometheus.metrics.opensearch_indices_shards_docs` +
				`{prometheus.labels.ip="a\\.b\\.c",prometheus.labels.cluster="opensearch",prometheus.labels.index=~"a.*|node.*"}` +
				`&start=1652320500&end=1652320620&step=10s`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), string(expected))
		})

		Convey("content-type is not application/x-www-form-urlencoded \n", func() {
			body := []byte(`query=prometheus.metrics.opensearch_indices_shards_docs` +
				`{prometheus.labels.cluster="opensearch",prometheus.labels.index=~"a.*|node.*"}` +
				`&start=1652320500&end=1652320620&step=10s`)

			expected := `{"status":"error",` +
				`"errorType":"status_not_acceptable",` +
				`"error":"Content-Type header [application/json] is not supported, expected is [application/x-www-form-urlencoded]"}`

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotAcceptable)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("start invalid \n", func() {
			body := []byte(`query=prometheus.metrics.opensearch_indices_shards_docs` +
				`{prometheus.labels.cluster="opensearch",prometheus.labels.index=~"a.*|node.*"}` +
				`&start=1652320500s&end=1652320620&step=10s`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("end invalid \n", func() {
			body := []byte(`query=prometheus.metrics.opensearch_indices_shards_docs` +
				`{prometheus.labels.cluster="opensearch",prometheus.labels.index=~"a.*|node.*"}` +
				`&start=1652320500&end=1652320620s&step=10s`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("step invalid \n", func() {
			body := []byte(`query=prometheus.metrics.opensearch_indices_shards_docs` +
				`{prometheus.labels.cluster="opensearch",prometheus.labels.index=~"a.*|node.*"}` +
				`&start=1652320500&end=1652320620&step="10"`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("step lte 0 \n", func() {
			body := []byte(`query=prometheus.metrics.opensearch_indices_shards_docs` +
				`{prometheus.labels.cluster="opensearch",prometheus.labels.index=~"a.*|node.*"}` +
				`&start=1652320500&end=1652320620&step=-1`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("start after now \n", func() {
			now := time.Now()
			hh, err := time.ParseDuration("1h")
			So(err, ShouldBeNil)
			start := now.Add(hh).UnixNano() / 1e9
			h, err := time.ParseDuration("2h")
			So(err, ShouldBeNil)
			end := now.Add(h).UnixNano() / 1e9

			body := []byte(fmt.Sprintf(`query=prometheus.metrics.opensearch_indices_shards_docs`+
				`{prometheus.labels.cluster="opensearch",prometheus.labels.index=~"a.*|node.*"}`+
				`&start=%v&end=%v&step=30s`, start, end))

			expected := fmt.Sprintf(`{"status":"error",`+
				`"errorType":"bad_data",`+
				`"error":"invalid parameter \"start\": start is greater than current time, current time is %v"}`, now.UnixNano()/1e9)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("end after now \n", func() {
			now := time.Now()
			hh, err := time.ParseDuration("1h")
			So(err, ShouldBeNil)
			end := now.Add(hh).UnixNano() / 1e9
			h, err := time.ParseDuration("-1h")
			So(err, ShouldBeNil)
			start := now.Add(h).UnixNano() / 1e9

			expected, _ := sonic.Marshal(mat)
			promqlService.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(interfaces.PromQLResponse{}, expected, http.StatusOK, nil)

			body := []byte(fmt.Sprintf(`query=prometheus.metrics.opensearch_indices_shards_docs`+
				`{prometheus.labels.cluster="opensearch",prometheus.labels.index=~"a.*|node.*"}`+
				`&start=%v&end=%v&step=30s`, start, end))
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), string(expected))
		})

		Convey("end before start \n", func() {
			body := []byte(`query=prometheus.metrics.opensearch_indices_shards_docs` +
				`{prometheus.labels.cluster="opensearch",prometheus.labels.index=~"a.*|node.*"}` +
				`&start=1646671470&end=1646360670&step=3`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("step is not a multiple of 5 minutes when step > 30min \n", func() {
			testStep := "2h30m45s20ms"
			queryStr := fmt.Sprintf("query=%s&start=%d&end=%d&step=%s",
				`query=abc[5m]`, 1646671470, 1646671770, testStep)
			body := []byte(queryStr)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)

			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("step is <= 30min \n", func() {
			expected, _ := sonic.Marshal("[]")
			promqlService.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(interfaces.PromQLResponse{}, expected, http.StatusOK, nil)
			testStep := "20m30s60ms"
			queryStr := fmt.Sprintf("query=%s&start=%d&end=%d&step=%s",
				`query=abc[5m]`, 1646671470, 1646671770, testStep)
			body := []byte(queryStr)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)

			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), string(expected))
		})

		Convey("exceeded maximum resolution of 11,000 points per timeseries \n", func() {
			body := []byte(`query=prometheus.metrics.opensearch_indices_shards_docs` +
				`{prometheus.labels.cluster="opensearch",prometheus.labels.index=~"a.*|node.*"}` +
				`&start=1646360670&end=1647671470&step=3`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("timeout invalid \n", func() {
			body := []byte(`query=prometheus.metrics.opensearch_indices_shards_docs` +
				`{prometheus.labels.cluster="opensearch",prometheus.labels.index=~"a.*|node.*"}` +
				`&start=1646360670&end=1646361670&step=3&timeout=-1s`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		// mock 执行很快，超时最小单位是ms，mock 的 exec 方法在 1ms 内已经返回。测不到。
		Convey("timeout \n", func() {
			resTemp, _ := sonic.Marshal(mat)
			promqlService.EXPECT().Exec(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, q interfaces.Query) (interfaces.PromQLResponse, []byte, int, error) {
				time.Sleep(time.Duration(2) * time.Millisecond)
				return interfaces.PromQLResponse{}, resTemp, http.StatusOK, nil
			})

			body := []byte(`query=prometheus.metrics.opensearch_indices_shards_docs` +
				`{prometheus.labels.cluster="opensearch",prometheus.labels.index=~"a.*|node.*"}` +
				`&start=1646360670&end=1646671470&step=30s&timeout=1ms`)

			expected := `{"status":"error",` +
				`"errorType":"timeout",` +
				`"error":"query timed out in expression evaluation"}`

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusServiceUnavailable)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("expression parse error \n", func() {
			promqlService.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(interfaces.PromQLResponse{}, nil, http.StatusBadRequest, uerrors.PromQLError{
				Typ: uerrors.ErrorBadData,
				Err: fmt.Errorf("expression parse error"),
			})

			body := []byte(`query=prometheus.metrics.opensearch_indices_shards_docs` +
				`{{prometheus.labels.cluster="opensearch",prometheus.labels.index=~"a.*|node.*"}` +
				`&start=1646360670&end=1646671470&step=30s`)

			expected := `{"status":"error",` +
				`"errorType":"bad_data",` +
				`"error":"expression parse error"}`

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("invalid expression type for range query \n", func() {
			promqlService.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(interfaces.PromQLResponse{}, nil, http.StatusUnprocessableEntity, uerrors.PromQLError{
				Typ: uerrors.ErrorExec,
				Err: fmt.Errorf("invalid expression type for range query, must be Scalar or instant Vector"),
			})

			body := []byte(`query=prometheus.metrics.opensearch_indices_shards_docs` +
				`{prometheus.labels.cluster="opensearch",prometheus.labels.index=~"a.*|node.*"}[5m]` +
				`&start=1646360670&end=1646671470&step=30s`)

			expected := `{"status":"error",` +
				`"errorType":"execution",` +
				`"error":"invalid expression type for range query, must be Scalar or instant Vector"}`

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusUnprocessableEntity)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("missing ar_dataview \n", func() {
			promqlService.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(interfaces.PromQLResponse{}, nil, http.StatusUnprocessableEntity, uerrors.PromQLError{
				Typ: uerrors.ErrorBadData,
				Err: fmt.Errorf("missing ar_dataview parameter"),
			})

			body := []byte(`query=prometheus.metrics.opensearch_indices_shards_docs` +
				`{prometheus.labels.cluster="opensearch",prometheus.labels.index=~"a.*|node.*"}` +
				`&start=1646360670&end=1646671470&step=30s`)

			expected := `{"status":"error",` +
				`"errorType":"bad_data",` +
				`"error":"missing ar_dataview parameter"}`

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusUnprocessableEntity)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("ar_dataview parameter is empty \n", func() {
			promqlService.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(interfaces.PromQLResponse{}, nil, http.StatusUnprocessableEntity, uerrors.PromQLError{
				Typ: uerrors.ErrorBadData,
				Err: fmt.Errorf("ar_dataview parameter value cannot be empty"),
			})

			body := []byte(`query=prometheus.metrics.opensearch_indices_shards_docs` +
				`{prometheus.labels.cluster="opensearch",prometheus.labels.index=~"a.*|node.*"}` +
				`&start=1646360670&end=1646671470&step=30s`)

			expected := `{"status":"error",` +
				`"errorType":"bad_data",` +
				`"error":"ar_dataview parameter value cannot be empty"}`

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusUnprocessableEntity)
			compareJsonString(w.Body.String(), expected)
		})
	})
}

func TestPromqlQuery(t *testing.T) {
	Convey("Test handler PromqlQuery", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydraMock := rmock.NewMockHydra(mockCtrl)
		promqlService := umock.NewMockPromQLService(mockCtrl)
		handler := mockNewPromqlRestHandler(appSetting, hydraMock, promqlService)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/promql/query"

		Convey("Success PromqlQuery \n", func() {
			expected, _ := sonic.Marshal(mat)
			promqlService.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(interfaces.PromQLResponse{}, expected, http.StatusOK, nil)

			body := []byte(`query=time()`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			So(w.Body.String(), ShouldEqual, string(expected))
		})

		Convey("content-type is not application/x-www-form-urlencoded \n", func() {
			body := []byte(`query=time()`)
			expected := `{"status":"error",` +
				`"errorType":"status_not_acceptable",` +
				`"error":"Content-Type header [application/json] is not supported, expected is [application/x-www-form-urlencoded]"}`

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotAcceptable)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("Success PromqlQuery parse time \n", func() {
			expected, _ := sonic.Marshal(mat)
			promqlService.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(interfaces.PromQLResponse{}, expected, http.StatusOK, nil)

			body := []byte(`query=time()&time=123`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			So(w.Body.String(), ShouldEqual, string(expected))
		})

		Convey("time invalid \n", func() {
			body := []byte(`query=time()&time=aaa`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("timeout invalid \n", func() {
			body := []byte(`query=time()&timeout=-1s`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		// mock 执行很快，超时最小单位是ms，mock 的 exec 方法在 1ms 内已经返回。测不到。
		Convey("timeout \n", func() {
			resTemp, _ := sonic.Marshal(mat)
			promqlService.EXPECT().Exec(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, q interfaces.Query) (interfaces.PromQLResponse, []byte, int, error) {
				time.Sleep(time.Duration(2) * time.Millisecond)
				return interfaces.PromQLResponse{}, resTemp, http.StatusOK, nil
			})

			body := []byte(`query=time()&timeout=1ms`)

			expected := `{"status":"error",` +
				`"errorType":"timeout",` +
				`"error":"query timed out in expression evaluation"}`

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusServiceUnavailable)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("expression parse error \n", func() {
			promqlService.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(interfaces.PromQLResponse{}, nil, http.StatusBadRequest, uerrors.PromQLError{
				Typ: uerrors.ErrorBadData,
				Err: fmt.Errorf("expression parse error"),
			})

			body := []byte(`query=time()`)

			expected := `{"status":"error",` +
				`"errorType":"bad_data",` +
				`"error":"expression parse error"}`

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("Success PromqlQuery with macthers value contain \\. \n", func() {
			expected, _ := sonic.Marshal(mat)
			promqlService.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(interfaces.PromQLResponse{}, expected, http.StatusOK, nil)

			body := []byte(`query=prometheus.metrics.opensearch_indices_shards_docs` +
				`{prometheus.labels.ip="a\\.b\\.c",prometheus.labels.cluster="opensearch",prometheus.labels.index=~"a.*|node.*"}`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			So(w.Body.String(), ShouldEqual, string(expected))
		})

	})
}

func TestPromqlSeries(t *testing.T) {
	Convey("Test handler PromqlSeries", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydraMock := rmock.NewMockHydra(mockCtrl)
		promqlService := umock.NewMockPromQLService(mockCtrl)
		handler := mockNewPromqlRestHandler(appSetting, hydraMock, promqlService)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/promql/series"

		Convey("Success PromqlSeries with default start end \n", func() {
			expected, _ := sonic.Marshal(series)
			promqlService.EXPECT().Series(gomock.Any()).Times(1).Return(expected, http.StatusOK, nil)

			body := []byte(`match[]=prometheus.metrics.node_cpu_guest_seconds_total{labels.mode="nice"}`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			So(w.Body.String(), ShouldEqual, string(expected))
		})

		Convey("Success PromqlSeries with default start end and matchers value contain \\. \n", func() {
			expected, _ := sonic.Marshal(series)
			promqlService.EXPECT().Series(gomock.Any()).Times(1).Return(expected, http.StatusOK, nil)

			body := []byte(`match[]=prometheus.metrics.node_cpu_guest_seconds_total{labels.instance="a\\.b\\.c"}`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			So(w.Body.String(), ShouldEqual, string(expected))
		})

		Convey("Success PromqlSeries with default start end and label_values \n", func() {

			body := []byte(`match[]=label_values(prometheus.metrics.node_cpu_guest_seconds_total{labels.mode="nice"},labels.instance)`)

			expected := `{"status":"error","errorType":"bad_data","error":"invalid parameter \"match[]\": 1:13: parse error: unexpected \"(\""}`

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("Success PromqlSeries with exact start end \n", func() {
			expected, _ := sonic.Marshal(series)
			promqlService.EXPECT().Series(gomock.Any()).Times(1).Return(expected, http.StatusOK, nil)

			body := []byte(`match[]=prometheus.metrics.node_cpu_guest_seconds_total{labels.mode="nice"}&start=1655346155&end=1655349445`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			So(w.Body.String(), ShouldEqual, string(expected))
		})

		Convey("Success PromqlSeries with exact start end and label_values \n", func() {

			body := []byte(`match[]=label_values(prometheus.metrics.node_cpu_guest_seconds_total{labels.mode="nice"},labels.instance)&start=1655346155&end=1655349445`)

			expected := `{"status":"error","errorType":"bad_data","error":"invalid parameter \"match[]\": 1:13: parse error: unexpected \"(\""}`

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("Success PromqlSeries with more than one match[] with same match[] \n", func() {
			expected, _ := sonic.Marshal(moreThanOneMatchSeries)
			promqlService.EXPECT().Series(gomock.Any()).Times(1).Return(expected, http.StatusOK, nil)

			body := []byte(`match[]=prometheus.metrics.node_cpu_guest_seconds_total{labels.mode="nice"}&match[]=prometheus.metrics.node_cpu_guest_seconds_total{labels.mode="nice"}`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			So(w.Body.String(), ShouldEqual, string(expected))
		})

		Convey("Success PromqlSeries with more than one match[] with diffrent match[] \n", func() {
			expected, _ := sonic.Marshal(moreThanOneMatchSeries)
			promqlService.EXPECT().Series(gomock.Any()).Times(1).Return(expected, http.StatusOK, nil)

			body := []byte(`match[]=prometheus.metrics.node_cpu_guest_seconds_total{labels.mode="nice"}&match[]=prometheus.metrics.node_filesystem_avail_bytes`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			So(w.Body.String(), ShouldEqual, string(expected))
		})

		Convey("Success PromqlSeries with non-metricsName and default start end \n", func() {
			expected, _ := sonic.Marshal(series)
			promqlService.EXPECT().Series(gomock.Any()).Times(1).Return(expected, http.StatusOK, nil)

			body := []byte(`match[]={labels.mode="nice"}`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			So(w.Body.String(), ShouldEqual, string(expected))
		})

		Convey("Success PromqlSeries with filter is not empty  \n", func() {
			expected, _ := sonic.Marshal(mat)
			promqlService.EXPECT().Series(gomock.Any()).Times(1).Return(expected, http.StatusOK, nil)

			body := []byte(`match[]=prometheus.metrics.node_cpu_guest_seconds_total{labels.mode="xxx"}`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			So(w.Body.String(), ShouldEqual, string(expected))
		})

		Convey("Error PromqlSeries with match[] filter is empty \n", func() {

			body := []byte(`match[]={labels.aaa=""}`)

			expected := `{"status":"error","errorType":"bad_data","error":"invalid parameter \"match[]\": match[] must contain at least one non-empty matcher"}`

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("Error PromqlSeries with match[] size == 0 \n", func() {

			body := []byte(``)

			expected := `{"status":"error","errorType":"bad_data","error":"no match[] parameter provided"}`

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("Error PromqlSeries with match[] is empty \n", func() {

			body := []byte(`match[]=`)

			expected := `{"status":"error","errorType":"bad_data","error":"invalid parameter \"match[]\": 1:1: parse error: unexpected end of input"}`

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("Error PromqlSeries with match[] format error \n", func() {

			body := []byte(`match[]=prometheus.metrics.node_exporter_build_info{a="a"}abc`)

			expected := `{"status":"error","errorType":"bad_data","error":"invalid parameter \"match[]\": 1:51: parse error: unexpected identifier \"abc\""}`

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("Error PromqlSeries with match[] format error 2 \n", func() {

			body := []byte(`match[]=prometheus.metrics.node_exporter_build_info{a="a"},abc)`)

			expected := `{"status":"error","errorType":"bad_data","error":"invalid parameter \"match[]\": 1:51: parse error: unexpected \",\""}`

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("Error PromqlSeries with match[] all filter is empty matcher \n", func() {

			body := []byte(`match[]={a=""}`)

			expected := `{"status":"error","errorType":"bad_data","error":"invalid parameter \"match[]\": match[] must contain at least one non-empty matcher"}`

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("Error PromqlSeries with match[] invalid \n", func() {

			body := []byte(`match[]=prometheus.metrics.node_cpu_guest_seconds_total[5m]`)

			expected := `{"status":"error","errorType":"bad_data","error":"invalid parameter \"match[]\": 1:48: parse error: unexpected \"[\""}`

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("Error PromqlSeries with start invalid \n", func() {

			body := []byte(`match[]=prometheus.metrics.node_cpu_guest_seconds_total&start=1655346155a&end=1655349445`)

			expected := `{"status":"error","errorType":"bad_data","error":"invalid parameter \"start\": invalid time value for 'start': cannot parse \"1655346155a\" to a valid timestamp"}`

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("Error PromqlSeries with end invalid \n", func() {

			body := []byte(`match[]=prometheus.metrics.node_cpu_guest_seconds_total&start=1655346155&end=1655349445a`)

			expected := `{"status":"error","errorType":"bad_data","error":"invalid parameter \"end\": invalid time value for 'end': cannot parse \"1655349445a\" to a valid timestamp"}`

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("Error PromqlSeries with end lt start \n", func() {

			body := []byte(`match[]=prometheus.metrics.node_cpu_guest_seconds_total&start=1655349445&end=1655346155`)

			expected := `{"status":"error","errorType":"bad_data","error":"invalid parameter \"end\": end timestamp must not be before start time"}`

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("Error PromqlSeries with start gt now \n", func() {
			now := time.Now()
			hh, err := time.ParseDuration("1h")
			So(err, ShouldBeNil)
			start := now.Add(hh).UnixNano() / 1e9
			h, err := time.ParseDuration("2h")
			So(err, ShouldBeNil)
			end := now.Add(h).UnixNano() / 1e9

			body := []byte(fmt.Sprintf(`match[]=prometheus.metrics.node_cpu_guest_seconds_total`+
				`&start=%v&end=%v`, start, end))

			expected := fmt.Sprintf(`{"status":"error",`+
				`"errorType":"bad_data",`+
				`"error":"invalid parameter \"start\": start is greater than current time, current time is %v"}`, now.UnixNano()/1e9)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("Error PromqlSeries with end gt now \n", func() {
			now := time.Now()
			hh, err := time.ParseDuration("1h")
			So(err, ShouldBeNil)
			end := now.Add(hh).UnixNano() / 1e9
			h, err := time.ParseDuration("-1h")
			So(err, ShouldBeNil)
			start := now.Add(h).UnixNano() / 1e9

			expected, _ := sonic.Marshal(series)
			promqlService.EXPECT().Series(gomock.Any()).Times(1).Return(expected, http.StatusOK, nil)

			body := []byte(fmt.Sprintf(`match[]=prometheus.metrics.node_cpu_guest_seconds_total`+
				`&start=%v&end=%v`, start, end))
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			So(w.Body.String(), ShouldEqual, string(expected))
		})

		Convey("Error PromqlSeries with ar_dataview is missing \n", func() {
			promqlService.EXPECT().Series(gomock.Any()).Times(1).Return(nil, http.StatusBadRequest, uerrors.PromQLError{
				Typ: uerrors.ErrorBadData,
				Err: fmt.Errorf("missing ar_dataview parameter or ar_dataview parameter value cannot be empty."),
			})

			body := []byte(`match[]=prometheus.metrics.node_cpu_guest_seconds_total{labels.mode="nice"}`)

			expected := `{"status":"error",` +
				`"errorType":"bad_data",` +
				`"error":"missing ar_dataview parameter or ar_dataview parameter value cannot be empty."}`

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			compareJsonString(w.Body.String(), expected)

		})
	})
}
