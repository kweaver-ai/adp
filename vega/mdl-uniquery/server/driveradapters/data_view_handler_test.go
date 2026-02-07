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

	"uniquery/common"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	umock "uniquery/interfaces/mock"
)

func mockNewDataViewRestHandler(appSetting *common.AppSetting,
	hydra rest.Hydra, dvService interfaces.DataViewService) (r *restHandler) {

	r = &restHandler{
		appSetting: appSetting,
		hydra:      hydra,
		dvService:  dvService,
	}
	r.InitMetric()
	return r
}

// func TestDataViewSimulate(t *testing.T) {
// 	Convey("Test handler DataView", t, func() {
// 		test := setGinMode()
// 		defer test()

// 		engine := gin.New()
// 		engine.Use(gin.Recovery())

// 		mockCtrl := gomock.NewController(t)
// 		defer mockCtrl.Finish()

// 		appSetting := &common.AppSetting{}
// 		hydraMock := rmock.NewMockHydra(mockCtrl)
// 		dvService := umock.NewMockDataViewService(mockCtrl)
// 		handler := mockNewDataViewRestHandler(appSetting, hydraMock, dvService)
// 		handler.RegisterPublic(engine)

// 		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

// 		url := "/api/mdl-uniquery/v1/data-views"

// 		query := interfaces.DataViewSimulateQuery{
// 			// DataSource: map[string]any{
// 			// 	"type": "index_base",
// 			// 	"index_base": []any{
// 			// 		interfaces.SimpleIndexBase{
// 			// 			BaseType: "x",
// 			// 		},
// 			// 	},
// 			// },
// 			// FieldScope: 1,
// 			Fields:     []*cond.ViewField{},
// 			ViewQueryCommonParams: interfaces.ViewQueryCommonParams{
// 				Start: 1678412100123,
// 				End:   1678412210123,
// 			},
// 		}
// 		reqParamByte, _ := sonic.Marshal(query)

// 		Convey("Simulate Success \n", func() {
// 			dvService.EXPECT().Simulate(gomock.Any(), gomock.Any()).Return(&interfaces.ViewUniResponseV2{}, nil)

// 			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
// 			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
// 			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			expected, _ := sonic.MarshalString(&interfaces.ViewUniResponseV2{})

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
// 			compareJsonString(w.Body.String(), expected)
// 		})

// 		Convey("Simulate Failed, DataView ShouldBind Error \n", func() {
// 			reqParamByte1, _ := sonic.Marshal([]interfaces.DataViewSimulateQuery{query})
// 			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte1))
// 			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
// 			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
// 		})

// 		Convey("Simulate Failed, Method is null \n", func() {
// 			reqParamByte1, _ := sonic.Marshal(interfaces.DataViewSimulateQuery{})
// 			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte1))
// 			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
// 		})

// 		Convey("Simulate Failed,, Method is not GET \n", func() {
// 			reqParamByte1, _ := sonic.Marshal(interfaces.DataViewSimulateQuery{})
// 			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte1))
// 			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
// 			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodPost)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
// 		})

// 		Convey("GetViewData Failed, Invalid start \n", func() {
// 			query1 := interfaces.DataViewSimulateQuery{
// 				ViewQueryCommonParams: interfaces.ViewQueryCommonParams{
// 					Start: -1,
// 					End:   1678412210123,
// 				},
// 			}
// 			reqParamByte1, _ := sonic.Marshal(query1)

// 			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte1))
// 			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
// 			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
// 		})

// 		Convey("Simulate Failed,, func Simulate error \n", func() {
// 			expectedErr := &rest.HTTPError{
// 				HTTPCode: http.StatusInternalServerError,
// 				BaseError: rest.BaseError{
// 					ErrorCode: uerrors.Uniquery_InternalError_SearchSubmitFailed,
// 				},
// 			}

// 			dvService.EXPECT().Simulate(gomock.Any(), gomock.Any()).Return(&interfaces.ViewUniResponseV2{}, expectedErr)

// 			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
// 			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
// 			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
// 		})
// 	})
// }

// func TestGetViewDataV1(t *testing.T) {
// 	Convey("Test handler GetViewDataV1", t, func() {
// 		test := setGinMode()
// 		defer test()

// 		engine := gin.New()
// 		engine.Use(gin.Recovery())

// 		mockCtrl := gomock.NewController(t)
// 		defer mockCtrl.Finish()

// 		appSetting := &common.AppSetting{}
// 		hydraMock := rmock.NewMockHydra(mockCtrl)
// 		dvService := umock.NewMockDataViewService(mockCtrl)
// 		handler := mockNewDataViewRestHandler(appSetting, hydraMock, dvService)
// 		handler.RegisterPublic(engine)

// 		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

// 		query := interfaces.DataViewQueryV1{
// 			ViewQueryCommonParams: interfaces.ViewQueryCommonParams{
// 				Start: 1678412100123,
// 				End:   1678412210123,
// 			},
// 		}
// 		reqParamByte, _ := sonic.Marshal(query)

// 		totalCount := int64(134)
// 		res := &interfaces.ViewUniResponseV2{
// 			Entries: []map[string]any{
// 				{"__id": "xx"},
// 			},
// 			TotalCount: &totalCount,
// 		}

// 		url := "/api/mdl-uniquery/v1/data-views/1"

// 		Convey("GetViewData Success, one view \n", func() {
// 			dvService.EXPECT().GetSingleViewData(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, nil)

// 			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
// 			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
// 			req.Header.Set(interfaces.X_HTTP_METHOD_OVERRIDE, http.MethodGet)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			expectedRes := &interfaces.ViewUniResponseV1{
// 				Datas: []interfaces.ViewData{{Total: &totalCount, Values: res.Entries}},
// 			}
// 			expected, _ := sonic.MarshalString(expectedRes)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
// 			compareJsonString(w.Body.String(), expected)
// 		})

// 		Convey("GetViewData Success, query with multiple view \n", func() {
// 			totalCount2 := int64(888)
// 			res2 := &interfaces.ViewUniResponseV2{
// 				Entries: []map[string]any{
// 					{"__id": "yy"},
// 				},
// 				TotalCount: &totalCount2,
// 			}
// 			dvService.EXPECT().GetSingleViewData(gomock.Any(), "1", gomock.Any()).Return(res, nil)
// 			dvService.EXPECT().GetSingleViewData(gomock.Any(), "2", gomock.Any()).Return(res2, nil)

// 			url = "/api/mdl-uniquery/v1/data-views/1,2"

// 			reqParamBytes, _ := sonic.Marshal([]interfaces.DataViewQueryV1{query, query})
// 			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamBytes))
// 			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
// 			req.Header.Set(interfaces.X_HTTP_METHOD_OVERRIDE, http.MethodGet)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			expectedRes := []*interfaces.ViewUniResponseV1{
// 				{
// 					Datas: []interfaces.ViewData{
// 						{Total: &totalCount, Values: res.Entries},
// 					},
// 				},
// 				{
// 					Datas: []interfaces.ViewData{
// 						{Total: &totalCount2, Values: res2.Entries},
// 					},
// 				},
// 			}
// 			expected, _ := sonic.MarshalString(expectedRes)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
// 			compareJsonString(w.Body.String(), expected)
// 		})

// 		Convey("GetViewData Failed, Method is null \n", func() {
// 			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
// 			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
// 		})

// 		Convey("GetViewData Failed, Method is not GET \n", func() {
// 			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
// 			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
// 			req.Header.Set(interfaces.X_HTTP_METHOD_OVERRIDE, http.MethodPut)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
// 		})

// 		Convey("GetViewData Failed, DataViewQueryV1 ShouldBind Error \n", func() {
// 			reqParamByte1, _ := sonic.Marshal([]interfaces.DataViewQueryV1{query})
// 			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte1))
// 			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
// 			req.Header.Set(interfaces.X_HTTP_METHOD_OVERRIDE, http.MethodGet)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
// 		})

// 		Convey("GetViewData Failed, DataViewQueryV1 ShouldBind Error with multi view \n", func() {
// 			url = "/api/mdl-uniquery/v1/data-views/1,2"
// 			reqParamByte1, _ := sonic.Marshal(query)
// 			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte1))
// 			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
// 			req.Header.Set(interfaces.X_HTTP_METHOD_OVERRIDE, http.MethodGet)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
// 		})

// 		Convey("GetViewData Failed, Invalid start \n", func() {
// 			query1 := interfaces.DataViewQueryV1{
// 				ViewQueryCommonParams: interfaces.ViewQueryCommonParams{
// 					Start: -1,
// 					End:   1678412210123,
// 				},
// 			}
// 			reqParamByte1, _ := sonic.Marshal(query1)

// 			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte1))
// 			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
// 			req.Header.Set(interfaces.X_HTTP_METHOD_OVERRIDE, http.MethodGet)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
// 		})

// 		Convey("GetViewDatas Failed, Invalid end \n", func() {
// 			query1 := []interfaces.DataViewQueryV1{
// 				{
// 					ViewQueryCommonParams: interfaces.ViewQueryCommonParams{
// 						Start: 1678412210123,
// 						End:   -1,
// 					},
// 				},
// 			}
// 			reqParamByte1, _ := sonic.Marshal(query1)

// 			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte1))
// 			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
// 			req.Header.Set(interfaces.X_HTTP_METHOD_OVERRIDE, http.MethodGet)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
// 		})

// 		Convey("GetViewData Failed, func GetViewData error\n", func() {
// 			expectedErr := &rest.HTTPError{
// 				HTTPCode: http.StatusInternalServerError,
// 				BaseError: rest.BaseError{
// 					ErrorCode: uerrors.Uniquery_InternalError_SearchSubmitFailed,
// 				},
// 			}

// 			dvService.EXPECT().GetSingleViewData(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, expectedErr)

// 			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
// 			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
// 			req.Header.Set(interfaces.X_HTTP_METHOD_OVERRIDE, http.MethodGet)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
// 		})

// 		Convey("GetViewDatas Failed, func GetSingleViewData error\n", func() {
// 			expectedErr := &rest.HTTPError{
// 				HTTPCode: http.StatusInternalServerError,
// 				BaseError: rest.BaseError{
// 					ErrorCode: uerrors.Uniquery_InternalError_SearchSubmitFailed,
// 				},
// 			}

// 			dvService.EXPECT().GetSingleViewData(gomock.Any(), "1", gomock.Any()).Return(res, expectedErr)
// 			dvService.EXPECT().GetSingleViewData(gomock.Any(), "2", gomock.Any()).Return(res, nil)

// 			url := "/api/mdl-uniquery/v1/data-views/1,2"
// 			reqParamByte, _ := sonic.Marshal([]interfaces.DataViewQueryV1{query, query})
// 			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
// 			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
// 			req.Header.Set(interfaces.X_HTTP_METHOD_OVERRIDE, http.MethodGet)
// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			So(w.Result().StatusCode, ShouldEqual, http.StatusMultiStatus)
// 		})
// 	})
// }

func TestGetViewData(t *testing.T) {
	Convey("Test handler GetViewDataV2", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydraMock := rmock.NewMockHydra(mockCtrl)
		dvService := umock.NewMockDataViewService(mockCtrl)
		handler := mockNewDataViewRestHandler(appSetting, hydraMock, dvService)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/data-views/1a"

		query := interfaces.DataViewQueryV2{
			ViewQueryCommonParams: interfaces.ViewQueryCommonParams{
				Start: 1678412100123,
				End:   1678412210123,
			},
		}
		reqParamByte, _ := sonic.Marshal(query)

		totalCount := int64(134)
		res := &interfaces.ViewUniResponseV2{
			Entries: []map[string]any{
				{"__id": "xx"},
			},
			TotalCount: &totalCount,
		}

		Convey("GetViewData Success, one view \n", func() {
			dvService.EXPECT().GetSingleViewData(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, nil)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			expected, _ := sonic.MarshalString(res)
			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("GetViewData Success, query with multiple view \n", func() {
			totalCount2 := int64(888)
			res2 := &interfaces.ViewUniResponseV2{
				Entries: []map[string]any{
					{"__id": "yy"},
				},
				TotalCount: &totalCount2,
			}
			dvService.EXPECT().GetSingleViewData(gomock.Any(), "1a", gomock.Any()).Return(res, nil)
			dvService.EXPECT().GetSingleViewData(gomock.Any(), "2a", gomock.Any()).Return(res2, nil)

			url = "/api/mdl-uniquery/v1/data-views/1a,2a"
			reqParamBytes, _ := sonic.Marshal([]interfaces.DataViewQueryV2{query, query})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamBytes))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			expectedRes := []*interfaces.ViewUniResponseV2{
				res,
				res2,
			}
			expected, _ := sonic.MarshalString(expectedRes)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("GetViewData Failed, Method is null \n", func() {
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("GetViewData Failed, Method is not GET \n", func() {
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodPut)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("GetViewData Failed, DataViewQueryV2 ShouldBind Error \n", func() {
			reqParamByte1, _ := sonic.Marshal([]interfaces.DataViewQueryV2{query})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte1))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("GetViewData Failed, DataViewQueryV1 ShouldBind Error with multi view \n", func() {
			url = "/api/mdl-uniquery/v1/data-views/1a,2a"
			reqParamByte1, _ := sonic.Marshal(query)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte1))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("GetViewData Failed, Invalid include_view \n", func() {
			url = "/api/mdl-uniquery/v1/data-views/1a?include_view=ture34"
			reqParamByte1, _ := sonic.Marshal(query)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte1))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("GetViewData Failed, Invalid allow_non_exist_field \n", func() {
			url = "/api/mdl-uniquery/v1/data-views/1a?allow_non_exist_field=true34"
			reqParamByte1, _ := sonic.Marshal(query)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte1))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("GetViewData Failed, Invalid start \n", func() {
			query1 := interfaces.DataViewQueryV2{
				ViewQueryCommonParams: interfaces.ViewQueryCommonParams{
					Start: -1,
					End:   1678412210123,
				},
			}
			reqParamByte1, _ := sonic.Marshal(query1)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte1))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("GetViewDatas Failed, Invalid end \n", func() {
			query1 := []interfaces.DataViewQueryV2{
				{
					ViewQueryCommonParams: interfaces.ViewQueryCommonParams{
						Start: 1678412210123,
						End:   -1,
					},
				},
			}
			reqParamByte1, _ := sonic.Marshal(query1)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte1))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("GetViewData Failed, func GetSingleViewData error\n", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				BaseError: rest.BaseError{
					ErrorCode: uerrors.Uniquery_InternalError_SearchSubmitFailed,
				},
			}

			dvService.EXPECT().GetSingleViewData(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, expectedErr)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("GetViewDatas Failed, func GetSingleViewData error with multi view \n", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				BaseError: rest.BaseError{
					ErrorCode: uerrors.Uniquery_InternalError_SearchSubmitFailed,
				},
			}

			dvService.EXPECT().GetSingleViewData(gomock.Any(), "1a", gomock.Any()).Return(res, expectedErr)
			dvService.EXPECT().GetSingleViewData(gomock.Any(), "2a", gomock.Any()).Return(res, nil)

			url := "/api/mdl-uniquery/v1/data-views/1a,2a"
			reqParamByte, _ := sonic.Marshal([]interfaces.DataViewQueryV2{query, query})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusMultiStatus)
		})
	})
}

func TestDeleteDataViewPits(t *testing.T) {
	Convey("Test handler DeleteDataViewPits", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydraMock := rmock.NewMockHydra(mockCtrl)
		dvService := umock.NewMockDataViewService(mockCtrl)
		handler := mockNewDataViewRestHandler(appSetting, hydraMock, dvService)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/data-view-pits"

		pits := interfaces.DeletePits{
			PitIDs: []string{interfaces.All_Pits_DataView, "1", "2"},
		}
		reqParamByte, _ := sonic.Marshal(pits)

		res := &interfaces.DeletePitsResp{
			Pits: []struct {
				PitID      string `json:"pit_id"`
				Successful bool   `json:"successful"`
			}{
				{PitID: "1", Successful: true},
				{PitID: "2", Successful: true},
			},
		}

		Convey("Delete failed, x-http-http.MethodPost-override is empty", func() {
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Delete Failed, x-http-http.MethodPost-override is not DELETE", func() {
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodPut)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Delete Failed, ShouldBindJSON error", func() {
			// 修改参数类型，使其匹配不上 Bind error
			reqParamByte1 := []byte(`{"pit_ids": [1, 2]}`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte1))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodDelete)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Delete Failed, DeleteDataViewPits error", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				BaseError: rest.BaseError{
					ErrorCode: uerrors.Uniquery_DataView_InternalError_DeletePointInTimeFailed,
				},
			}
			dvService.EXPECT().DeleteDataViewPits(gomock.Any(), gomock.Any()).Return(res, expectedErr)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodDelete)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Delete succeed", func() {
			dvService.EXPECT().DeleteDataViewPits(gomock.Any(), gomock.Any()).Return(res, nil)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodDelete)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}
