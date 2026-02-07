// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"net/http"
	"net/http/httptest"
	"testing"

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

func mockNewTraceRestHandler(appSetting *common.AppSetting,
	hydra rest.Hydra, tService interfaces.TraceService) (r *restHandler) {
	r = &restHandler{
		appSetting: appSetting,
		hydra:      hydra,
		tService:   tService,
	}
	r.InitMetric()
	return r
}

// func TestGetSpanList(t *testing.T) {
// 	Convey("Test handler GetSpanList", t, func() {
// 		test := setGinMode()
// 		defer test()

// 		engine := gin.New()
// 		engine.Use(gin.Recovery())

// 		mockCtrl := gomock.NewController(t)
// 		defer mockCtrl.Finish()

// 		tService := umock.NewMockTraceService(mockCtrl)
// 		handler := mockNewTraceRestHandler(tService)
// 		handler.RegisterPublic(engine)

// 		Convey("Get Span List failed, caused by empty ar_dataview", func() {
// 			url := "/api/mdl-uniquery/v1/spans"

// 			req := httptest.NewRequest(http.MethodGet, url, nil)
// 			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
// 			res, _ := convert.JsonToMap(w.Body.String())
// 			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_MissingParameter_ARDataView)
// 		})

// 		Convey("Get Span List failed, caused by empty start_time", func() {
// 			baseURL := "/api/mdl-uniquery/v1/spans"
// 			url := baseURL + "?ar_dataview=69bbd3de-ac32-11ed-89e8-ca9b4576213469bbd3de-ac32-11ed-89e8-ca9b45762134"

// 			req := httptest.NewRequest(http.MethodGet, url, nil)
// 			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
// 			res, _ := convert.JsonToMap(w.Body.String())
// 			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_MissingParameter_StartTime)
// 		})

// 		Convey("Get Span List failed, caused by empty end_time", func() {
// 			baseURL := "/api/mdl-uniquery/v1/spans"
// 			url := baseURL + "?ar_dataview=69bbd3de-ac32-11ed-89e8-ca9b4576213469bbd3de-ac32-11ed-89e8-ca9b45762134&start_time=1614136611000"

// 			req := httptest.NewRequest(http.MethodGet, url, nil)
// 			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
// 			res, _ := convert.JsonToMap(w.Body.String())
// 			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_MissingParameter_EndTime)
// 		})

// 		Convey("Get Span List failed, caused by error from ValidateOffsetAndLimit function", func() {
// 			expectedErr := &rest.HTTPError{
// 				HTTPCode: http.StatusBadRequest,
// 				BaseError: rest.BaseError{
// 					ErrorCode:    uerrors.Uniquery_InvalidParameter_Limit,
// 					Description:  "指定的每页数量限制无效",
// 					Solution:     "",
// 					ErrorLink:    "",
// 					ErrorDetails: fmt.Sprintf("The number per page is not in the range of [%d,%d]", interfaces.MIN_LIMIT, interfaces.MAX_LIMIT),
// 				},
// 			}
// 			patch := ApplyFunc(ValidateOffsetAndLimit,
// 				func(offsetStr, limitStr string) (int, int, error) {
// 					return 0, 0, expectedErr
// 				},
// 			)
// 			defer patch.Reset()

// 			baseURL := "/api/mdl-uniquery/v1/spans"
// 			url := baseURL + "?ar_dataview=69bbd3de-ac32-11ed-89e8-ca9b4576213469bbd3de-ac32-11ed-89e8-ca9b45762134&start_time=1614136611000&end_time=1614136621000&limit=1001"
// 			req := httptest.NewRequest(http.MethodGet, url, nil)
// 			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
// 			res, _ := convert.JsonToMap(w.Body.String())
// 			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_InvalidParameter_Limit)
// 		})

// 		Convey("Get Span List failed, caused by error from ValidateSpanQueryTime function", func() {
// 			expectedErr := &rest.HTTPError{
// 				HTTPCode: http.StatusBadRequest,
// 				BaseError: rest.BaseError{
// 					ErrorCode:    uerrors.Uniquery_InvalidParameter_StartTime,
// 					Description:  "指定的开始时间无效",
// 					Solution:     "",
// 					ErrorLink:    "",
// 					ErrorDetails: "The start_time is greater than current time, current timestamp is 1677208611000.",
// 				},
// 			}
// 			patch := ApplyFunc(ValidateSpanQueryTime,
// 				func(startTime, endTime string) (int64, int64, error) {
// 					return 0, 0, expectedErr
// 				},
// 			)
// 			defer patch.Reset()

// 			baseURL := "/api/mdl-uniquery/v1/spans"
// 			url := baseURL + "?ar_dataview=69bbd3de-ac32-11ed-89e8-ca9b4576213469bbd3de-ac32-11ed-89e8-ca9b45762134&start_time=1614136611000&end_time=1614136621000"
// 			req := httptest.NewRequest(http.MethodGet, url, nil)
// 			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
// 			res, _ := convert.JsonToMap(w.Body.String())
// 			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_InvalidParameter_StartTime)
// 		})

// 		Convey("Get Span List failed, caused by error from ValidateSpanStatuses function", func() {
// 			expectedErr := &rest.HTTPError{
// 				HTTPCode: http.StatusBadRequest,
// 				BaseError: rest.BaseError{
// 					ErrorCode:    uerrors.Uniquery_InvalidParameter_SpanStatuses,
// 					Description:  "指定的跨度状态无效",
// 					Solution:     "",
// 					ErrorLink:    "",
// 					ErrorDetails: fmt.Sprintf("The span_statuses is not in the set of [%s]", interfaces.DEFAULT_SPAN_STATUSES),
// 				},
// 			}

// 			patch1 := ApplyFunc(ValidateSpanQueryTime,
// 				func(startTime, endTime string) (int64, int64, error) {
// 					return 1614136611000, 1614136621000, nil
// 				},
// 			)
// 			defer patch1.Reset()

// 			patch2 := ApplyFunc(ValidateSpanStatuses,
// 				func(spanStatuses string) (map[string]int, error) {
// 					return map[string]int{}, expectedErr
// 				},
// 			)
// 			defer patch2.Reset()

// 			baseURL := "/api/mdl-uniquery/v1/spans"
// 			url := baseURL + "?ar_dataview=69bbd3de-ac32-11ed-89e8-ca9b4576213469bbd3de-ac32-11ed-89e8-ca9b45762134&start_time=1614136611000&end_time=1614136621000&span_statuses=abc"
// 			req := httptest.NewRequest(http.MethodGet, url, nil)
// 			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
// 			res, _ := convert.JsonToMap(w.Body.String())
// 			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_InvalidParameter_SpanStatuses)
// 		})

// 		Convey("Get Span List failed, caused by error from logical layer", func() {
// 			expectedErr := &rest.HTTPError{
// 				HTTPCode: http.StatusInternalServerError,
// 				BaseError: rest.BaseError{
// 					ErrorCode:    uerrors.Uniquery_InternalError,
// 					Description:  "服务器内部错误",
// 					Solution:     "",
// 					ErrorLink:    "",
// 					ErrorDetails: "opensearch error",
// 				},
// 			}
// 			tService.EXPECT().GetSpanList(gomock.Any()).AnyTimes().Return(
// 				[]interfaces.SpanList{}, 0, expectedErr)

// 			patch := ApplyFunc(ValidateSpanQueryTime,
// 				func(startTime, endTime string) (int64, int64, error) {
// 					return 1614136611000, 1614136621000, nil
// 				},
// 			)
// 			defer patch.Reset()

// 			baseURL := "/api/mdl-uniquery/v1/spans"
// 			url := baseURL + "?ar_dataview=69bbd3de-ac32-11ed-89e8-ca9b4576213469bbd3de-ac32-11ed-89e8-ca9b45762134&start_time=1614136611000&end_time=1614136621000&span_statuses=Ok,Error"

// 			req := httptest.NewRequest(http.MethodGet, url, nil)
// 			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
// 			res, _ := convert.JsonToMap(w.Body.String())
// 			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_InternalError)
// 		})

// 		Convey("Get Span List successed", func() {
// 			expectedSpanList := []interfaces.SpanList{}
// 			tService.EXPECT().GetSpanList(gomock.Any()).AnyTimes().Return(
// 				expectedSpanList, 0, nil)

// 			patch := ApplyFunc(ValidateSpanQueryTime,
// 				func(startTime, endTime string) (int64, int64, error) {
// 					return 1614136611000, 1614136621000, nil
// 				},
// 			)
// 			defer patch.Reset()

// 			baseURL := "/api/mdl-uniquery/v1/spans"
// 			url := baseURL + "?ar_dataview=69bbd3de-ac32-11ed-89e8-ca9b4576213469bbd3de-ac32-11ed-89e8-ca9b45762134&start_time=1614136611000&end_time=1614136621000&span_statuses=Ok"

// 			req := httptest.NewRequest(http.MethodGet, url, nil)
// 			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
// 		})
// 	})
// }

func TestGetTraceDetail(t *testing.T) {
	Convey("Test handler GetTraceDetail", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydraMock := rmock.NewMockHydra(mockCtrl)
		tService := umock.NewMockTraceService(mockCtrl)
		handler := mockNewTraceRestHandler(appSetting, hydraMock, tService)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		// spanJson1 := `{
		// 	"Name" : "Span2-1",
		// 	"SpanContext" : {
		// 	  "TraceID" : "ab4fe1de2d9c951411bbdbc3533d747e",
		// 	  "SpanID" : "44c39ef193f3e08b",
		// 	  "TraceFlags" : "01",
		// 	  "TraceState" : {
		// 		"rojo" : "00f067aa0ba902b7",
		// 		"congo" : "t61rcWkgMzE"
		// 	  }
		// 	},
		// 	"Parent" : {
		// 	  "TraceID" : "00000000000000000000000000000000",
		// 	  "SpanID" : "0000000000000000",
		// 	  "TraceFlags" : "00",
		// 	  "TraceState" : null
		// 	},
		// 	"SpanKind" : 1,
		// 	"StartTime" : 1662144778114304000,
		// 	"EndTime" : 1662144838114435000,
		// 	"@timestamp" : "2022-09-02T18:52:58.114Z",
		// 	"Duration" : 131000,
		// 	"Attributes" : {
		// 	  "num" : "45"
		// 	},
		// 	"Events" : [
		// 	  {
		// 		"Name" : "Nice operation!",
		// 		"Attributes" : {
		// 		  "bogons" : 100
		// 		},
		// 		"Time" : 1662058438114304000
		// 	  }
		// 	],
		// 	"Links" : [
		// 	  {
		// 		"TraceID" : "ab4fe1de2d9c951411bbdbc3533d747f",
		// 		"SpanID" : "44c39ef193f3e08G",
		// 		"TraceState" : {
		// 		  "rojo" : "00f067aa0ba902b7",
		// 		  "congo" : "t61rcWkgMzE"
		// 		},
		// 		"Attributes" : {
		// 		  "num1" : 100
		// 		}
		// 	  }
		// 	],
		// 	"Status" : {
		// 	  "Code": 0,
		// 	  "CodeDesc" : "Unset",
		// 	  "Description" : ""
		// 	},
		// 	"Resource" : {
		// 	  "job_id" : "13bf135ce8b1481e9329a5e3b62171ae",
		// 	  "service" : {
		// 		"name" : "服务A",
		// 		"version" : "版本1.0.1"
		// 	  },
		// 	  "telemetry" : {
		// 		"sdk" : {
		// 		  "language" : "go",
		// 		  "name" : "opentelemetry",
		// 		  "version" : "1.9.0"
		// 		}
		// 	  }
		// 	},
		// 	"catagory" : "trace"
		//   }`

		// spanJson2 := `{
		// 	"Name" : "Span2-2",
		// 	"SpanContext" : {
		// 	  "TraceID" : "ab4fe1de2d9c951411bbdbc3533d747e",
		// 	  "SpanID" : "44c39ef193f3e07b",
		// 	  "TraceFlags" : "01",
		// 	  "TraceState" : {
		// 		"rojo" : "00f067aa0ba902b7",
		// 		"congo" : "t61rcWkgMzE"
		// 	  }
		// 	},
		// 	"Parent" : {
		// 	  "TraceID" : "ab4fe1de2d9c951411bbdbc3533d747e",
		// 	  "SpanID" : "44c39ef193f3e08b",
		// 	  "TraceFlags" : "01",
		// 	  "TraceState" : {
		// 		"rojo" : "00f067aa0ba902b7",
		// 		"congo" : "t61rcWkgMzE"
		// 	  }
		// 	},
		// 	"SpanKind" : 1,
		// 	"StartTime" : 1662144838114304000,
		// 	"EndTime" : 1662144898114435000,
		// 	"@timestamp" : "2022-09-02T18:53:58.114Z",
		// 	"Duration" : 131000,
		// 	"Attributes" : {
		// 	  "num" : "45"
		// 	},
		// 	"Events" : [
		// 	  {
		// 		"Name" : "Nice operation!",
		// 		"Attributes" : {
		// 		  "bogons" : 100
		// 		},
		// 		"Time" : 1662058438114304000
		// 	  }
		// 	],
		// 	"Links" : [
		// 	  {
		// 		"TraceID" : "ab4fe1de2d9c951411bbdbc3533d747f",
		// 		"SpanID" : "44c39ef193f3e08G",
		// 		"TraceState" : {
		// 		  "rojo" : "00f067aa0ba902b7",
		// 		  "congo" : "t61rcWkgMzE"
		// 		},
		// 		"Attributes" : {
		// 		  "num1" : 100
		// 		}
		// 	  }
		// 	],
		// 	"Status" : {
		// 	  "Code": 0,
		// 	  "CodeDesc" : "Unset",
		// 	  "Description" : ""
		// 	},
		// 	"Resource" : {
		// 	  "job_id" : "13bf135ce8b1481e9329a5e3b62171ae",
		// 	  "service" : {
		// 		"name" : "服务B",
		// 		"version" : "版本1.0.1"
		// 	  },
		// 	  "telemetry" : {
		// 		"sdk" : {
		// 		  "language" : "go",
		// 		  "name" : "opentelemetry",
		// 		  "version" : "1.9.0"
		// 		}
		// 	  }
		// 	},
		// 	"catagory" : "trace"
		//   }`

		// spanJson3 := `{
		// 	"Name" : "Span2-3",
		// 	"SpanContext" : {
		// 	  "TraceID" : "ab4fe1de2d9c951411bbdbc3533d747e",
		// 	  "SpanID" : "44c39ef193f3e06b",
		// 	  "TraceFlags" : "01",
		// 	  "TraceState" : {
		// 		"rojo" : "00f067aa0ba902b7",
		// 		"congo" : "t61rcWkgMzE"
		// 	  }
		// 	},
		// 	"Parent" : {
		// 	  "TraceID" : "ab4fe1de2d9c951411bbdbc3533d747e",
		// 	  "SpanID" : "44c39ef193f3e07b",
		// 	  "TraceFlags" : "01",
		// 	  "TraceState" : {
		// 		"rojo" : "00f067aa0ba902b7",
		// 		"congo" : "t61rcWkgMzE"
		// 	  }
		// 	},
		// 	"SpanKind" : 1,
		// 	"StartTime" : 1662144898114304000,
		// 	"EndTime" : 1662144958114435000,
		// 	"@timestamp" : "2022-09-02T18:54:58.114Z",
		// 	"Duration" : 131000,
		// 	"Attributes" : {
		// 	  "num" : "45"
		// 	},
		// 	"Events" : [
		// 	  {
		// 		"name" : "Nice operation!",
		// 		"attributes" : {
		// 		  "bogons" : 100
		// 		},
		// 		"time" : 1662058438114304000
		// 	  }
		// 	],
		// 	"Links" : [
		// 	  {
		// 		"TraceID" : "ab4fe1de2d9c951411bbdbc3533d747f",
		// 		"SpanID" : "44c39ef193f3e08G",
		// 		"TraceState" : {
		// 		  "rojo" : "00f067aa0ba902b7",
		// 		  "congo" : "t61rcWkgMzE"
		// 		},
		// 		"Attributes" : {
		// 		  "num1" : 100
		// 		}
		// 	  }
		// 	],
		// 	"Status" : {
		// 	  "Code": 0,
		// 	  "CodeDesc" : "Unset",
		// 	  "Description" : ""
		// 	},
		// 	"Resource" : {
		// 	  "job_id" : "13bf135ce8b1481e9329a5e3b62171ae",
		// 	  "service" : {
		// 		"name" : "服务D",
		// 		"version" : "版本1.0.1"
		// 	  },
		// 	  "telemetry" : {
		// 		"sdk" : {
		// 		  "language" : "go",
		// 		  "name" : "opentelemetry",
		// 		  "version" : "1.9.0"
		// 		}
		// 	  }
		// 	},
		// 	"catagory" : "trace"
		//   }`
		// traceId := "ab4fe1de2d9c951411bbdbc3533d747e"
		// spanStats := map[string]int32{
		// 	"ok":    0,
		// 	"error": 0,
		// 	"unset": 3,
		// }
		// traceStatus := "ok"
		// services := []interfaces.Service{
		// 	{
		// 		Name: "服务D",
		// 	},
		// 	{
		// 		Name: "服务B",
		// 	},
		// 	{
		// 		Name: "服务A",
		// 	},
		// }

		// span1 := make(map[string]interface{}, 0)
		// err := sonic.Unmarshal([]byte(spanJson1), &span1)
		// So(err, ShouldBeNil)

		// span2 := make(map[string]interface{}, 0)
		// err = sonic.Unmarshal([]byte(spanJson2), &span2)
		// So(err, ShouldBeNil)

		// span3 := make(map[string]interface{}, 0)
		// err = sonic.Unmarshal([]byte(spanJson3), &span3)
		// So(err, ShouldBeNil)

		Convey("Get Trace Detail failed, caused by trace_data_view_id is not passed in", func() {
			url := "/api/mdl-uniquery/v1/traces/ab4fe1de2d9c951411bbdbc3533d747e"

			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_Trace_MissingParameter_TraceDataViewID)
		})

		Convey("Get Trace Detail failed, caused by trace_data_view_id is empty string", func() {
			url := "/api/mdl-uniquery/v1/traces/ab4fe1de2d9c951411bbdbc3533d747e?trace_data_view_id="

			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_Trace_InvalidParameter_TraceDataViewID)
		})

		Convey("Get Trace Detail failed, caused by log_data_view_id is not passed in", func() {
			url := "/api/mdl-uniquery/v1/traces/ab4fe1de2d9c951411bbdbc3533d747e?trace_data_view_id=1"

			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_Trace_MissingParameter_LogDataViewID)
		})

		Convey("Get Trace Detail failed, caused by log_data_view_id is empty string", func() {
			url := "/api/mdl-uniquery/v1/traces/ab4fe1de2d9c951411bbdbc3533d747e?trace_data_view_id=1&log_data_view_id="

			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_Trace_InvalidParameter_LogDataViewID)
		})

		Convey("Get Trace Detail failed, caused by error from logical layer", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_InternalError,
					Description:  "服务器内部错误",
					Solution:     "",
					ErrorLink:    "",
					ErrorDetails: "opensearch error",
				},
			}
			tService.EXPECT().GetTraceDetail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				&interfaces.TraceDetail{}, expectedErr)

			url := "/api/mdl-uniquery/v1/traces/ab4fe1de2d9c951411bbdbc3533d747e?trace_data_view_id=1&log_data_view_id=2"

			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
			res, _ := convert.JsonToMap(w.Body.String())
			So(res["error_code"].(string), ShouldEqual, uerrors.Uniquery_InternalError)
		})

		Convey("Get Trace Detail successed", func() {
			traceId := "ab4fe1de2d9c951411bbdbc3533d747e"
			expectedTraceDetail := &interfaces.TraceDetail{
				TraceID:     traceId,
				TraceStatus: "ok",
			}
			tService.EXPECT().GetTraceDetail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				expectedTraceDetail, nil)

			url := "/api/mdl-uniquery/v1/traces/ab4fe1de2d9c951411bbdbc3533d747e?trace_data_view_id=1&log_data_view_id=2"

			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}
