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

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	rmock "github.com/kweaver-ai/kweaver-go-lib/rest/mock"
	. "github.com/smartystreets/goconvey/convey"

	"uniquery/common"
	"uniquery/common/convert"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	umock "uniquery/interfaces/mock"
)

var (
	dslResult = map[string]interface{}{
		"_shards": map[string]interface{}{
			"failed":     0,
			"skipped":    0,
			"successful": 0,
			"total":      0,
		},
		"hits": map[string]interface{}{
			"total": map[string]interface{}{
				"value":    0,
				"relation": "eq",
			},
			"max_score": 0,
			"hits":      []string{},
		},
		"timed_out": false,
		"took":      0,
	}

	dslCount = map[string]interface{}{
		"count": 0,
		"_shards": map[string]interface{}{
			"total":      0,
			"successful": 0,
			"skipped":    0,
			"failed":     0,
		},
	}

	dslDeleteAllScrollResult = map[string]interface{}{
		"succeeded": true,
		"num_freed": 1,
	}
)

func mockNewDSLRestHandler(appSetting *common.AppSetting,
	hydra rest.Hydra, dslService interfaces.DslService) (r *restHandler) {
	r = &restHandler{
		appSetting: appSetting,
		hydra:      hydra,
		dslService: dslService,
	}
	r.InitMetric()
	return r
}

func TestDslGetResult(t *testing.T) {
	Convey("Test handler get_result", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydraMock := rmock.NewMockHydra(mockCtrl)
		dslService := umock.NewMockDslService(mockCtrl)
		handler := mockNewDSLRestHandler(appSetting, hydraMock, dslService)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		Convey("illegal scroll", func() {
			searchBody := map[string]interface{}{}
			url := "/api/mdl-uniquery/v1/dsl/kc,123/_search?scroll=1"

			reqParamByte, _ := sonic.Marshal(searchBody)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("body invalid", func() {
			url := "/api/mdl-uniquery/v1/dsl/_search"

			reqParamByte := []byte("12345")
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error"].(map[string]interface{})["type"], ShouldEqual, uerrors.IllegalArgumentException)
		})

		Convey("library, param not exist", func() {
			resTemp, _ := sonic.Marshal(dslResult)
			oerr := uerrors.NewOpenSearchError(uerrors.IllegalArgumentException).WithReason("x_library must be array")
			dslService.EXPECT().Search(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(resTemp, http.StatusBadRequest, oerr)

			searchBody := map[string]interface{}{
				"size": 10,
			}
			url := "/api/mdl-uniquery/v1/dsl/_search"

			reqParamByte, _ := sonic.Marshal(searchBody)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error"].(map[string]interface{})["type"], ShouldEqual, uerrors.IllegalArgumentException)
		})

		Convey("scroll exist", func() {
			resTemp, _ := sonic.Marshal(dslResult)
			dslService.EXPECT().Search(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(resTemp, http.StatusOK, nil)

			searchBody := map[string]interface{}{
				"size": 10,
			}
			url := "/api/mdl-uniquery/v1/dsl/kc,123/_search?scroll=1m"

			reqParamByte, _ := sonic.Marshal(searchBody)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), string(resTemp))
		})
	})
}

func TestDslScroll(t *testing.T) {
	Convey("Test scroll handler", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydraMock := rmock.NewMockHydra(mockCtrl)
		dslService := umock.NewMockDslService(mockCtrl)
		handler := mockNewDSLRestHandler(appSetting, hydraMock, dslService)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/dsl/_search/scroll"

		Convey("not json body", func() {
			searchBody := "12345"

			reqParamByte, _ := sonic.Marshal(searchBody)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("empty body", func() {
			req := httptest.NewRequest(http.MethodPost, url, nil)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("illegal json body", func() {
			searchBody := map[string]interface{}{
				"x_library": "123",
				"size":      10,
			}

			resTemp := `{"status":400,"error":{` +
				`"type":"UniQuery.IllegalArgumentException",` +
				`"reason":"Key: 'Scroll.ScrollId' Error:Field validation for 'ScrollId' failed on the 'required' tag"}}`

			reqParamByte, _ := sonic.Marshal(searchBody)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			compareJsonString(w.Body.String(), string(resTemp))
		})

		Convey("ScrollSearch failed", func() {
			dslService.EXPECT().ScrollSearch(gomock.Any(),
				gomock.Any()).Return(nil, http.StatusBadRequest, uerrors.NewOpenSearchError(uerrors.IllegalArgumentException))

			searchBody := map[string]interface{}{
				"scroll_id": "12232",
				"scroll":    "1h",
			}

			resTemp := `{"status":400,` +
				`"error":{"type":"UniQuery.IllegalArgumentException",` +
				`"reason":"illegal_argument_exception"}}`

			reqParamByte, _ := sonic.Marshal(searchBody)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			compareJsonString(w.Body.String(), string(resTemp))
		})

		Convey("ScrollSearch success", func() {
			resTemp, _ := sonic.Marshal(dslResult)
			dslService.EXPECT().ScrollSearch(gomock.Any(),
				gomock.Any()).Return(resTemp, http.StatusOK, nil)

			searchBody := map[string]interface{}{
				"scroll_id": "12232",
				"scroll":    "1h",
			}

			reqParamByte, _ := sonic.Marshal(searchBody)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), string(resTemp))
		})
	})
}

func TestDslGetCount(t *testing.T) {
	Convey("Test handler get_count", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydraMock := rmock.NewMockHydra(mockCtrl)
		dslService := umock.NewMockDslService(mockCtrl)
		handler := mockNewDSLRestHandler(appSetting, hydraMock, dslService)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		Convey("body invalid", func() {
			url := "/api/mdl-uniquery/v1/dsl/_count"

			reqParamByte := []byte("12345")
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error"].(map[string]interface{})["type"], ShouldEqual, uerrors.IllegalArgumentException)
		})

		Convey("library, param not exist", func() {
			resTemp, _ := sonic.Marshal(dslCount)
			oerr := uerrors.NewOpenSearchError(uerrors.IllegalArgumentException).WithReason("x_library must be array")
			dslService.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(resTemp, http.StatusBadRequest, oerr)

			queryBody := map[string]interface{}{
				"x_library": "111",
				"query": map[string]interface{}{
					"match_all": map[string]interface{}{},
				},
			}
			url := "/api/mdl-uniquery/v1/dsl/_count"

			reqParamByte, _ := sonic.Marshal(queryBody)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error"].(map[string]interface{})["type"], ShouldEqual, uerrors.IllegalArgumentException)
		})

		Convey("query is normal", func() {
			resTemp, _ := sonic.Marshal(dslCount)
			dslService.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(resTemp, http.StatusOK, nil)

			queryBody := map[string]interface{}{
				"x_library": "kafka",
				"query": map[string]interface{}{
					"match_all": map[string]interface{}{},
				},
			}
			url := "/api/mdl-uniquery/v1/dsl/_count"

			reqParamByte, _ := sonic.Marshal(queryBody)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), string(resTemp))
		})

		Convey("illegal contentType", func() {
			// resTemp, _ := sonic.Marshal(dslCount)
			// dslService.EXPECT().Count(gomock.Any(), gomock.Any()).AnyTimes().Return(resTemp, http.StatusOK, nil)

			queryBody := map[string]interface{}{
				"x_library": "kafka",
				"query": map[string]interface{}{
					"match_all": map[string]interface{}{},
				},
			}
			url := "/api/mdl-uniquery/v1/dsl/_count"

			resTemp := `{"status":"error",` +
				`"errorType":"status_not_acceptable",` +
				`"error":"Content-Type header [application/x-www-form-urlencoded] is not supported, expected is [application/json]"}`

			reqParamByte, _ := sonic.Marshal(queryBody)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotAcceptable)
			compareJsonString(w.Body.String(), string(resTemp))
		})
	})
}

func TestDslDeleteScroll(t *testing.T) {
	Convey("Test handler delete scroll", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydraMock := rmock.NewMockHydra(mockCtrl)
		dslService := umock.NewMockDslService(mockCtrl)
		handler := mockNewDSLRestHandler(appSetting, hydraMock, dslService)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		Convey("delete scroll success", func() {
			resTemp, _ := sonic.Marshal(dslDeleteAllScrollResult)
			dslService.EXPECT().DeleteScroll(gomock.Any(), gomock.Any()).AnyTimes().Return(resTemp, http.StatusOK, nil)

			queryBody := map[string]interface{}{
				"scroll_id": []string{"_all"},
			}
			url := "/api/mdl-uniquery/v1/dsl/_search/scroll"

			reqParamByte, _ := sonic.Marshal(queryBody)
			req := httptest.NewRequest(http.MethodDelete, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), string(resTemp))
		})

		Convey("body without scroll_id", func() {
			url := "/api/mdl-uniquery/v1/dsl/_search/scroll"

			reqParamByte := []byte("12345")
			req := httptest.NewRequest(http.MethodDelete, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error"].(map[string]interface{})["type"], ShouldEqual, uerrors.IllegalArgumentException)
		})

		Convey("scroll_id invalid, not array", func() {
			queryBody := map[string]interface{}{
				"scroll_id": "_all",
			}
			url := "/api/mdl-uniquery/v1/dsl/_search/scroll"

			reqParamByte, _ := sonic.Marshal(queryBody)
			req := httptest.NewRequest(http.MethodDelete, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error"].(map[string]interface{})["type"], ShouldEqual, uerrors.IllegalArgumentException)
		})

		Convey("scroll_id valid, but expired", func() {
			err := errors.New("{\"succeeded\":true,\"num_freed\":0}")
			dslService.EXPECT().DeleteScroll(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, http.StatusNotFound, err)

			queryBody := map[string]interface{}{
				"scroll_id": []string{"FGluY2x1ZGVfY29udGV4dF91dWlkDXF1ZXJ5QW5kRmV0Y2gBFmhFRWEtSS1wUXZpc0xfQ3Q5NUxzNEEAAAAAAAAABBZCQXl2QnB3X1N5S09BQ3J0TE5vQVJn"},
			}
			url := "/api/mdl-uniquery/v1/dsl/_search/scroll"

			reqParamByte, _ := sonic.Marshal(queryBody)
			req := httptest.NewRequest(http.MethodDelete, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)

			res, _ := convert.JsonToMap(w.Body.String())
			So(res["num_freed"], ShouldEqual, 0)
		})

		Convey("scroll_id is array, but invalid", func() {
			err := errors.New("{\"error\":{\"root_cause\":[{\"type\":\"illegal_argument_exception\",\"reason\":\"Cannot parse scroll id\"}],\"type\":\"illegal_argument_exception\",\"reason\":\"Cannot parse scroll id\",\"caused_by\":{\"type\":\"array_index_out_of_bounds_exception\",\"reason\":\"arraycopy: last source index 3657 out of bounds for byte[2]\"}},\"status\":400}")
			dslService.EXPECT().DeleteScroll(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, http.StatusBadRequest, err)

			queryBody := map[string]interface{}{
				"scroll_id": []string{"xxx"},
			}
			url := "/api/mdl-uniquery/v1/dsl/_search/scroll"

			reqParamByte, _ := sonic.Marshal(queryBody)
			req := httptest.NewRequest(http.MethodDelete, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error"].(map[string]interface{})["type"], ShouldEqual, "illegal_argument_exception")
		})

		Convey("delete scroll with InternalServerError", func() {
			oerr := uerrors.NewOpenSearchError(uerrors.InternalServerError).WithReason("error connection")
			dslService.EXPECT().DeleteScroll(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, http.StatusInternalServerError, oerr)

			queryBody := map[string]interface{}{
				"scroll_id": []string{"_all"},
			}
			url := "/api/mdl-uniquery/v1/dsl/_search/scroll"

			reqParamByte, _ := sonic.Marshal(queryBody)
			req := httptest.NewRequest(http.MethodDelete, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error"].(map[string]interface{})["type"], ShouldEqual, uerrors.InternalServerError)
		})
	})
}

func TestDslDeleteAllScroll(t *testing.T) {
	Convey("Test handler delete all scroll", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydraMock := rmock.NewMockHydra(mockCtrl)
		dslService := umock.NewMockDslService(mockCtrl)
		handler := mockNewDSLRestHandler(appSetting, hydraMock, dslService)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		Convey("delete all scroll success", func() {
			resTemp, _ := sonic.Marshal(dslDeleteAllScrollResult)
			dslService.EXPECT().DeleteScroll(gomock.Any(), gomock.Any()).AnyTimes().Return(resTemp, http.StatusOK, nil)

			url := "/api/mdl-uniquery/v1/dsl/_search/scroll/_all"

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), string(resTemp))
		})

		Convey("delete all scroll failed", func() {
			oerr := uerrors.NewOpenSearchError(uerrors.InternalServerError).WithReason("error connection")
			dslService.EXPECT().DeleteScroll(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, http.StatusInternalServerError, oerr)

			url := "/api/mdl-uniquery/v1/dsl/_search/scroll/_all"

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error"].(map[string]interface{})["type"], ShouldEqual, uerrors.InternalServerError)
		})
	})
}
