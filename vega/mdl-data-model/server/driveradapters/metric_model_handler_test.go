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
	testMetricModel = interfaces.MetricModel{
		SimpleMetricModel: interfaces.SimpleMetricModel{
			ModelName:    "16",
			MetricType:   "atomic",
			QueryType:    "promql",
			Tags:         []string{"a", "s", "s", "s", "s"},
			Comment:      "ssss",
			GroupID:      "0",
			GroupName:    "group1",
			Formula:      "avg(rate(node_cpu_seconds_total{name=~\"node-14-35\",mode12=\"system\"}[5m]))by(name)",
			UnitType:     "storeUnit",
			Unit:         "Byte",
			MeasureName:  "__m.a",
			DateField:    interfaces.PROMQL_DATEFIELD,
			MeasureField: interfaces.PROMQL_METRICFIELD,
		},
		DataSource: &interfaces.MetricDataSource{
			Type: interfaces.SOURCE_TYPE_DATA_VIEW,
			ID:   "数据视图1",
		},
	}

	testSimpleMetricModel = interfaces.SimpleMetricModel{
		ModelID:    "1",
		ModelName:  "dataModle",
		MetricType: "atomic",
		QueryType:  "promql",
		Tags:       []string{"tag1", "tag2"},
		UpdateTime: testMetricUpdateTime,
	}

	// testMetricModelWithFilters = interfaces.MetricModelWithFilters{
	// 	MetricModel: testMetricModel,
	// }

	testTask = interfaces.MetricTask{
		TaskID:   "1",
		TaskName: "task1",
		ModelID:  "1",
		Schedule: interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
		// Filters:            []interfaces.Filter{{Name: "labels.mode", Operation: "=", Value: "idle"}},
		TimeWindows:        []string{"5m", "1h"},
		IndexBase:          "base1",
		RetraceDuration:    "1d",
		ScheduleSyncStatus: 1,
		Comment:            "task1-aaa",
		UpdateTime:         testMetricUpdateTime,
		PlanTime:           int64(1699336878575),
	}

	testMetricGroup = interfaces.MetricModelGroup{
		GroupID:   "1",
		GroupName: "111",
	}
)

func MockNewMetricModelRestHandler(appSetting *common.AppSetting,
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

func Test_MetricModelRestHandler_CreateMetricModels(t *testing.T) {
	Convey("Test MetricModelHandler CreateMetricModels\n", t, func() {
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

		handler := MockNewMetricModelRestHandler(appSetting, hydra, mms, dvs, mmgs, mmts)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/metric-models"

		Convey("Success CreateMetricModels \n", func() {
			mms.EXPECT().CheckMetricModelExistByName(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return("", false, nil)
			mms.EXPECT().CreateMetricModels(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return([]string{"1"}, nil)
			mms.EXPECT().CheckMetricModelByMeasureName(gomock.Any(), gomock.Any()).AnyTimes().Return("", false, nil)

			reqParamByte, _ := sonic.Marshal([]interfaces.MetricModel{testMetricModel})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("Success CreateMetricModels with data view id is not null and data source is null \n", func() {
			mms.EXPECT().CheckMetricModelExistByName(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return("", false, nil)
			mms.EXPECT().CreateMetricModels(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return([]string{"1"}, nil)
			mms.EXPECT().CheckMetricModelByMeasureName(gomock.Any(), gomock.Any()).AnyTimes().Return("", false, nil)

			reqParamByte, _ := sonic.Marshal([]interfaces.CreateMetricModel{
				{
					ModelName:    "16",
					MetricType:   "atomic",
					QueryType:    "promql",
					Tags:         []string{"a", "s", "s", "s", "s"},
					Comment:      "ssss",
					GroupName:    "group1",
					Formula:      "avg(rate(node_cpu_seconds_total{name=~\"node-14-35\",mode12=\"system\"}[5m]))by(name)",
					UnitType:     "storeUnit",
					Unit:         "Byte",
					MeasureName:  "__m.a",
					DateField:    interfaces.PROMQL_DATEFIELD,
					MeasureField: interfaces.PROMQL_METRICFIELD,
					DataViewID:   "123",
				},
			})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("Success CreateMetricModels 2 elements \n", func() {
			mms.EXPECT().CheckMetricModelExistByName(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return("", false, nil)
			mms.EXPECT().CreateMetricModels(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return([]string{"1"}, nil)
			mms.EXPECT().CheckMetricModelByMeasureName(gomock.Any(), gomock.Any()).AnyTimes().Return("", false, nil)

			reqParamByte, _ := sonic.Marshal([]interfaces.MetricModel{testMetricModel, {
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:    "17",
					MetricType:   "atomic",
					QueryType:    "promql",
					Tags:         []string{"a", "s", "s", "s", "s"},
					Comment:      "ssss",
					GroupName:    "group1",
					UnitType:     "storeUnit",
					Unit:         "Byte",
					Formula:      "avg(rate(node_cpu_seconds_total{name=~\"node-14-35\",mode12=\"system\"}[5m]))by(name)",
					MeasureName:  "__m.b",
					DateField:    interfaces.PROMQL_DATEFIELD,
					MeasureField: interfaces.PROMQL_METRICFIELD,
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "数据视图1",
				},
			}})

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("Failed CreateMetricModels contentType is not json\n", func() {
			reqParamByte, _ := sonic.Marshal(testMetricModel)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotAcceptable)
		})

		Convey("Failed CreateMetricModels ShouldBind Error\n", func() {
			reqParamByte, _ := sonic.Marshal(testMetricModel)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Model name is null\n", func() {
			reqParamByte, _ := sonic.Marshal([]interfaces.MetricModel{{}})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Model name length exceeded \n", func() {
			reqParamByte, _ := sonic.Marshal([]interfaces.MetricModel{{SimpleMetricModel: interfaces.SimpleMetricModel{
				ModelName: "111111111111111111111111111111111111111111111111",
			}}})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Group name length exceeded \n", func() {
			reqParamByte, _ := sonic.Marshal([]interfaces.MetricModel{
				{
					SimpleMetricModel: interfaces.SimpleMetricModel{
						ModelName: "1111",
						GroupName: "111111111111111111111111111111111111111111111111",
					},
				}})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Model type is null \n", func() {
			reqParamByte, _ := sonic.Marshal([]interfaces.MetricModel{{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName: "11111",
				}}})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Unsupported model type\n", func() {
			reqParamByte, _ := sonic.Marshal([]interfaces.MetricModel{{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "11111",
					MetricType: "atomic1",
				},
			}})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Data view is null\n", func() {
			reqParamByte, _ := sonic.Marshal([]interfaces.MetricModel{{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "11111",
					MetricType: "atomic",
				},
			}})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Query Type is null\n", func() {
			reqParamByte, _ := sonic.Marshal([]interfaces.MetricModel{{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "11111",
					MetricType: "atomic",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "dv",
				},
			}})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Unsupported query Type\n", func() {
			reqParamByte, _ := sonic.Marshal([]interfaces.MetricModel{{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "11111",
					MetricType: "atomic",
					QueryType:  "promql1",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "dv",
				},
			}})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Formula is null\n", func() {
			reqParamByte, _ := sonic.Marshal([]interfaces.MetricModel{{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "11111",
					MetricType: "atomic",
					QueryType:  "promql",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "dv",
				},
			}})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Tags count more than 5\n", func() {
			reqParamByte, _ := sonic.Marshal([]interfaces.MetricModel{{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "11111",
					MetricType: "atomic",
					Tags:       []string{"a", "b", "c", "d", "e", "f"},
					QueryType:  "promql",
					Formula:    "abc",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "dv",
				},
			}})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Tags name is empty string\n", func() {
			reqParamByte, _ := sonic.Marshal([]interfaces.MetricModel{{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "11111",
					MetricType: "atomic",
					QueryType:  "promql",
					Tags:       []string{"a", "", "c", "d", "e", "f"},
					Formula:    "abc",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "dv",
				},
			}})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Tags name length is more than 40\n", func() {
			reqParamByte, _ := sonic.Marshal([]interfaces.MetricModel{{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "11111",
					MetricType: "atomic",
					QueryType:  "promql",
					Tags:       []string{"a", "111111111111111111111111111111111111111111", "c", "d", "e", "f"},
					Formula:    "abc",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "dv",
				},
			}})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Tags name contain special string\n", func() {
			reqParamByte, _ := sonic.Marshal([]interfaces.MetricModel{{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "11111",
					MetricType: "atomic",
					QueryType:  "promql",
					Tags:       []string{"a", "1/", "c", "d", "e", "f"},
					Formula:    "abc",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "dv",
				},
			}})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Invalid comment\n", func() {
			reqParamByte, _ := sonic.Marshal([]interfaces.MetricModel{{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "11111",
					MetricType: "atomic",
					QueryType:  "promql",
					Comment: `ddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd
				ddddddddddddddddsdddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddeeeeeeeedddddd
				ddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddffffffffffffffffffffffffffffffffffffffffffffffff`,
					Formula: "abc",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "dv",
				},
			}})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		// Convey("Combination name exists \n", func() {
		// 	existModelErr := &rest.HTTPError{
		// 		HTTPCode: http.StatusForbidden,
		// 		Language: rest.DefaultLanguage,
		// 		BaseError: rest.BaseError{
		// 			ErrorCode: derrors.DataModel_MetricModel_CombinationNameExisted,
		// 		},
		// 	}
		// 	mms.EXPECT().CheckMetricModelExistByName(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return("aaa", true, existModelErr)

		// 	reqParamByte, _ := sonic.Marshal([]interfaces.MetricModel{testMetricModel})
		// 	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
		// 	req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
		// 	w := httptest.NewRecorder()
		// 	engine.ServeHTTP(w, req)

		// 	So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		// })

		// Convey("Measure name exists \n", func() {
		// 	existModelErr := &rest.HTTPError{
		// 		HTTPCode: http.StatusForbidden,
		// 		Language: "zh-CN",
		// 		BaseError: rest.BaseError{
		// 			ErrorCode: derrors.DataModel_MetricModel_Duplicated_MeasureName,
		// 		},
		// 	}
		// 	mms.EXPECT().CheckMetricModelExistByName(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return("", false, nil)
		// 	mms.EXPECT().CheckMetricModelByMeasureName(gomock.Any(), gomock.Any()).AnyTimes().Return("", true, existModelErr)

		// 	reqParamByte, _ := sonic.Marshal([]interfaces.MetricModel{testMetricModel})
		// 	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
		// 	req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
		// 	w := httptest.NewRecorder()
		// 	engine.ServeHTTP(w, req)

		// 	So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		// })

		// Convey("Duplicate Measure name with db \n", func() {
		// 	mms.EXPECT().CheckMetricModelExistByName(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return("", false, nil)
		// 	mms.EXPECT().CheckMetricModelByMeasureName(gomock.Any(), gomock.Any()).AnyTimes().Return("a", true, nil)

		// 	reqParamByte, _ := sonic.Marshal([]interfaces.MetricModel{testMetricModel})
		// 	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
		// 	req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
		// 	w := httptest.NewRecorder()
		// 	engine.ServeHTTP(w, req)

		// 	So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		// })

		// Convey("Duplicate Measure name in body \n", func() {
		// 	// existModelErr := &rest.HTTPError{
		// 	// 	HTTPCode: http.StatusBadRequest,
		// 	// 	Language: "zh-CN",
		// 	// 	BaseError: rest.BaseError{
		// 	// 		ErrorCode: derrors.DataModel_MetricModel_Duplicated_MeasureName,
		// 	// 	},
		// 	// }
		// 	mms.EXPECT().CheckMetricModelExistByName(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return("", false, nil)
		// 	mms.EXPECT().CheckMetricModelByMeasureName(gomock.Any(), gomock.Any()).AnyTimes().Return("", false, nil)

		// 	model2 := interfaces.MetricModel{
		// 		SimpleMetricModel: interfaces.SimpleMetricModel{
		// 			ModelName:    "17",
		// 			MetricType:   "atomic",
		// 			QueryType:    "promql",
		// 			Tags:         []string{"a", "s", "s", "s", "s"},
		// 			Comment:      "ssss",
		// 			GroupID:      "0",
		// 			GroupName:    "group1",
		// 			UnitType:     "storeUnit",
		// 			Unit:         "Byte",
		// 			Formula:      "avg(rate(node_cpu_seconds_total{name=~\"node-14-35\",mode12=\"system\"}[5m]))by(name)",
		// 			MeasureName:  "__m.a",
		// 			DateField:    interfaces.PROMQL_DATEFIELD,
		// 			MeasureField: interfaces.PROMQL_METRICFIELD,
		// 		},
		// 		DataSource: &interfaces.MetricDataSource{
		// 			Type: interfaces.SOURCE_TYPE_DATA_VIEW,
		// 			ID:   "数据视图1",
		// 		},
		// 	}

		// 	reqParamByte, _ := sonic.Marshal([]interfaces.MetricModel{testMetricModel, model2})
		// 	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
		// 	req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
		// 	w := httptest.NewRecorder()
		// 	engine.ServeHTTP(w, req)

		// 	So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		// })

		// Convey("Duplicate combination name\n", func() {
		// 	mms.EXPECT().CheckMetricModelExistByName(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return("", false, nil)
		// 	mms.EXPECT().CheckFormula(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
		// 	mms.EXPECT().CheckMetricModelByMeasureName(gomock.Any(), gomock.Any()).AnyTimes().Return("", false, nil)

		// 	reqParamByte, _ := sonic.Marshal([]interfaces.MetricModel{testMetricModel, testMetricModel})
		// 	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
		// 	req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
		// 	w := httptest.NewRecorder()
		// 	engine.ServeHTTP(w, req)

		// 	So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		// })

		Convey("Create model failed\n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: derrors.DataModel_MetricModel_InternalError,
				},
			}

			mms.EXPECT().CheckMetricModelExistByName(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return("", false, nil)
			mms.EXPECT().CreateMetricModels(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil, err)
			mms.EXPECT().CheckMetricModelByMeasureName(gomock.Any(), gomock.Any()).AnyTimes().Return("", false, nil)

			reqParamByte, _ := sonic.Marshal([]interfaces.MetricModel{testMetricModel})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_MetricModelRestHandler_UpdateMetricModel(t *testing.T) {
	Convey("Test MetricModelHandler UpdateMetricModel\n", t, func() {
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

		handler := MockNewMetricModelRestHandler(appSetting, hydra, mms, dvs, mmgs, mmts)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/metric-models/1"

		oldMetricModel := interfaces.MetricModel{
			SimpleMetricModel: interfaces.SimpleMetricModel{
				ModelName:    "17",
				MetricType:   "atomic",
				QueryType:    "promql",
				Tags:         []string{"a", "s", "s", "s", "s"},
				Comment:      "ssss",
				GroupID:      "1",
				GroupName:    "group1",
				UnitType:     "storeUnit",
				Unit:         "Byte",
				Formula:      "sum(rate(node_cpu_seconds_total{name=~\"node-14-35\",mode12=\"system\"}[5m]))by(name)",
				DateField:    interfaces.PROMQL_DATEFIELD,
				MeasureField: interfaces.PROMQL_METRICFIELD,
			},
			DataSource: &interfaces.MetricDataSource{
				Type: interfaces.SOURCE_TYPE_DATA_VIEW,
				ID:   "数据视图2",
			},
		}
		// metricModelGroup := interfaces.MetricModelGroup{
		// 	GroupID:   1,
		// 	GroupName: "group1",
		// }
		Convey("Success UpdateMetricModel \n", func() {
			mms.EXPECT().CheckMetricModelExistByName(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return("", false, nil)
			mms.EXPECT().CheckMetricModelByMeasureName(gomock.Any(), gomock.Any()).AnyTimes().Return("", false, nil)
			mms.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(oldMetricModel, nil)
			mms.EXPECT().UpdateMetricModel(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			reqParamByte, _ := sonic.Marshal(testMetricModel)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("Failed contentType is not json\n", func() {
			reqParamByte, _ := sonic.Marshal(interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName: "11111",
				},
			})

			url = "/api/mdl-data-model/v1/metric-models/1"
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotAcceptable)
		})

		Convey("Failed UpdateMetricModel ShouldBind Error\n", func() {

			reqParamByte, _ := sonic.Marshal([]interfaces.MetricModel{testMetricModel})
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Model name is null\n", func() {
			reqParamByte, _ := sonic.Marshal(interfaces.MetricModel{})
			mms.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(oldMetricModel, nil)

			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Model name length exceeded \n", func() {
			reqParamByte, _ := sonic.Marshal(interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName: "111111111111111111111111111111111111111111111111",
				},
			})
			mms.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(oldMetricModel, nil)

			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Model type is null \n", func() {
			reqParamByte, _ := sonic.Marshal(interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName: "11111",
				},
			})
			mms.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(oldMetricModel, nil)

			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Unsupported model type\n", func() {
			reqParamByte, _ := sonic.Marshal(interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "11111",
					MetricType: "atomic1",
				},
			})
			mms.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(oldMetricModel, nil)

			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Data view is null\n", func() {
			reqParamByte, _ := sonic.Marshal(interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "11111",
					MetricType: "atomic",
				},
			})
			mms.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(oldMetricModel, nil)

			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Query Type is null\n", func() {
			reqParamByte, _ := sonic.Marshal(interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "11111",
					MetricType: "atomic",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "dv",
				},
			})
			mms.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(oldMetricModel, nil)

			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Unsupported query Type\n", func() {
			reqParamByte, _ := sonic.Marshal(interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "11111",
					MetricType: "atomic",
					QueryType:  "promql1",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "dv",
				},
			})
			mms.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(oldMetricModel, nil)

			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Formula is null\n", func() {
			reqParamByte, _ := sonic.Marshal(interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "11111",
					MetricType: "atomic",
					QueryType:  "promql",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "dv",
				},
			})
			mms.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(oldMetricModel, nil)

			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Tags count more than 5\n", func() {
			reqParamByte, _ := sonic.Marshal(interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "11111",
					MetricType: "atomic",
					QueryType:  "promql",
					Tags:       []string{"a", "b", "c", "d", "e", "f"},
					Formula:    "abc",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "dv",
				},
			})
			mms.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(oldMetricModel, nil)

			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Tags name is empty string\n", func() {
			reqParamByte, _ := sonic.Marshal(interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "11111",
					MetricType: "atomic",
					QueryType:  "promql",
					Tags:       []string{"a", "", "c", "d", "e", "f"},
					Formula:    "abc",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "dv",
				},
			})
			mms.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(oldMetricModel, nil)

			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Tags name length is more than 40\n", func() {
			reqParamByte, _ := sonic.Marshal(interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "11111",
					MetricType: "atomic",
					QueryType:  "promql",
					Tags:       []string{"a", "111111111111111111111111111111111111111111", "c", "d", "e", "f"},
					Formula:    "abc",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "dv",
				},
			})
			mms.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(oldMetricModel, nil)

			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Tags name contain special string\n", func() {
			reqParamByte, _ := sonic.Marshal(interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "11111",
					MetricType: "atomic",
					QueryType:  "promql",
					Tags:       []string{"a", "1/", "c", "d", "e", "f"},
					Formula:    "abc",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "dv",
				},
			})
			mms.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(oldMetricModel, nil)

			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Invalid comment\n", func() {
			reqParamByte, _ := sonic.Marshal(interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "11111",
					MetricType: "atomic",
					QueryType:  "promql",
					Comment: `ddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd
				ddddddddddddddddsdddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddeeeeeeeedddddd
				ddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddffffffffffffffffffffffffffffffffffffffffffffffff`,
					Formula: "abc",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "dv",
				},
			})
			mms.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(oldMetricModel, nil)

			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		// Convey("RetriveGroupIDByGroupName Failed \n", func() {
		// 	err := &rest.HTTPError{
		// 		HTTPCode: http.StatusInternalServerError,
		// 		Language: rest.DefaultLanguage,
		// 		BaseError: rest.BaseError{
		// 			ErrorCode: derrors.DataModel_MetricModelGroup_InternalError_GetMetricModelGroupIDByNameFailed,
		// 		},
		// 	}
		// 	mms.EXPECT().CheckMetricModelExistByName(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return("", false, nil)
		// 	mms.EXPECT().CheckMetricModelByMeasureName(gomock.Any(), gomock.Any()).AnyTimes().Return(false, nil)
		// 	mms.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(interfaces.MetricModel{}, err)

		// 	// mms.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(oldMetricModel, nil)
		// 	mms.EXPECT().RetriveGroupIDByGroupName(gomock.Any(), gomock.Any()).AnyTimes().Return(testMetricGroup, err)

		// 	reqParamByte, _ := sonic.Marshal(testMetricModel)
		// 	req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
		// 	req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
		// 	w := httptest.NewRecorder()
		// 	engine.ServeHTTP(w, req)
		// 	So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		// })

		Convey("CheckMetricModelExist Failed \n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: derrors.DataModel_MetricModel_InternalError_CheckModelIfExistFailed,
				},
			}
			mms.EXPECT().CheckMetricModelExistByName(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return("", false, err)
			mms.EXPECT().CheckMetricModelByMeasureName(gomock.Any(), gomock.Any()).AnyTimes().Return("", false, nil)
			mms.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(oldMetricModel, nil)

			// mms.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(interfaces.MetricModel{}, nil)
			// mms.EXPECT().CheckMetricModelExist(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(false, err)

			reqParamByte, _ := sonic.Marshal(testMetricModel)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("different model name && new name exists\n", func() {
			// err := &rest.HTTPError{
			// 	HTTPCode: http.StatusBadRequest,
			// 	Language: rest.DefaultLanguage,
			// 	BaseError: rest.BaseError{
			// 		ErrorCode: derrors.MetricModel_ModelNameExisted,
			// 	},
			// }
			mms.EXPECT().CheckMetricModelExistByName(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return("aaa", true, nil)
			mms.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(interfaces.MetricModel{}, nil)
			// mms.EXPECT().CheckMetricModelExist(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)

			reqParamByte, _ := sonic.Marshal(testMetricModel)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("Update Metric Model failed\n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: derrors.DataModel_MetricModel_InternalError,
				},
			}
			mms.EXPECT().CheckMetricModelExistByName(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return("", false, nil)
			mms.EXPECT().CheckMetricModelByMeasureName(gomock.Any(), gomock.Any()).AnyTimes().Return("", false, nil)
			mms.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(oldMetricModel, nil)
			// mms.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(interfaces.MetricModel{}, nil)
			mms.EXPECT().RetriveGroupIDByGroupName(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(testMetricGroup, nil)
			// mms.EXPECT().CheckMetricModelExist(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(false, nil)

			mms.EXPECT().UpdateMetricModel(gomock.Any(), gomock.Any(), gomock.Any()).Return(err)

			reqParamByte, _ := sonic.Marshal(testMetricModel)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_MetricModelRestHandler_DeleteMetricModels(t *testing.T) {
	Convey("Test MetricModelHandler DeleteMetricModels\n", t, func() {
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

		handler := MockNewMetricModelRestHandler(appSetting, hydra, mms, dvs, mmgs, mmts)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/metric-models/1,2"

		Convey("Success DeleteMetricModels \n", func() {
			mms.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).AnyTimes().Return(testMetricModel, nil)
			mms.EXPECT().DeleteMetricModels(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(int64(2), nil)

			reqParamByte, _ := sonic.Marshal(testMetricModel)
			req := httptest.NewRequest(http.MethodDelete, url, bytes.NewReader(reqParamByte))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("Get Metric Model By ID Failed \n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusNotFound,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: derrors.DataModel_MetricModel_MetricModelNotFound,
				},
			}
			mms.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(interfaces.MetricModel{}, err)

			reqParamByte, _ := sonic.Marshal(testMetricModel)
			req := httptest.NewRequest(http.MethodDelete, url, bytes.NewReader(reqParamByte))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("Delete Metric Model failed\n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: derrors.DataModel_MetricModel_InternalError,
				},
			}
			mms.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).AnyTimes().Return(testMetricModel, nil)
			mms.EXPECT().DeleteMetricModels(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(int64(0), err)

			reqParamByte, _ := sonic.Marshal(testMetricModel)
			req := httptest.NewRequest(http.MethodDelete, url, bytes.NewReader(reqParamByte))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_MetricModelRestHandler_ListMetricModels(t *testing.T) {
	Convey("Test MetricModelHandler ListMetricModels\n", t, func() {
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

		handler := MockNewMetricModelRestHandler(appSetting, hydra, mms, dvs, mmgs, mmts)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/metric-models"

		Convey("Success ListSimpleMetricModels \n", func() {
			mms.EXPECT().ListSimpleMetricModels(gomock.Any(), gomock.Any()).AnyTimes().Return([]interfaces.SimpleMetricModel{testSimpleMetricModel}, 1, nil)

			url = url + "?direction=desc&sort=update_time&limit=1000&offset=0&simple_info=true"
			req := httptest.NewRequest(http.MethodGet, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("offest invalid \n", func() {
			url = url + "?direction=desc&sort=update_time&limit=1000&offset=a"
			req := httptest.NewRequest(http.MethodGet, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("name && name_pattren both exists \n", func() {
			url = url + "?direction=desc&sort=update_time&limit=1000&offset=0&name=a&name_pattern=b"
			req := httptest.NewRequest(http.MethodGet, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("ListSimpleMetricModels Failed \n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: derrors.DataModel_MetricModel_InternalError,
				},
			}

			mms.EXPECT().ListSimpleMetricModels(gomock.Any(), gomock.Any()).AnyTimes().Return([]interfaces.SimpleMetricModel{}, 0, err)

			url = url + "?direction=desc&sort=update_time&limit=1000&offset=0"
			req := httptest.NewRequest(http.MethodGet, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_MetricModelRestHandler_GetMetricModels(t *testing.T) {
	Convey("Test MetricModelHandler GetMetricModel\n", t, func() {
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

		handler := MockNewMetricModelRestHandler(appSetting, hydra, mms, dvs, mmgs, mmts)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/metric-models/1"

		Convey("Success GetMetricModel \n", func() {
			mms.EXPECT().GetMetricModels(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().
				Return([]interfaces.MetricModelWithFilters{}, nil)

			req := httptest.NewRequest(http.MethodGet, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("include_view invalid \n", func() {
			reqParamByte, _ := sonic.Marshal(interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName: "11111",
				},
			})

			url = "/api/mdl-data-model/v1/metric-models/1?include_view=a"
			req := httptest.NewRequest(http.MethodGet, url, bytes.NewReader(reqParamByte))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("GetMetricModelContainFilters Failed \n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusNotFound,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: derrors.DataModel_MetricModel_MetricModelNotFound,
				},
			}

			mms.EXPECT().GetMetricModels(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().
				Return([]interfaces.MetricModelWithFilters{}, err)

			req := httptest.NewRequest(http.MethodGet, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})
	})
}

func Test_MetricModelRestHandler_UpdateMetricModels(t *testing.T) {
	Convey("Test MetricModelHandler UpdateMetricModels\n", t, func() {
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

		handler := MockNewMetricModelRestHandler(appSetting, hydra, mms, dvs, mmgs, mmts)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/metric-models/1,2/attributes"

		metricModelGroupName := interfaces.MetricModelGroupName{
			GroupName: "group1",
		}
		modelSimpleInfo := interfaces.SimpleMetricModel{
			ModelID:   "1",
			ModelName: "model1",
			GroupName: "group1",
		}
		modelSimpleInfo2 := interfaces.SimpleMetricModel{
			ModelID:   "2",
			ModelName: "model2",
			GroupName: "group1",
		}
		modelMap := map[string]interfaces.SimpleMetricModel{modelSimpleInfo.ModelID: modelSimpleInfo, modelSimpleInfo2.ModelID: modelSimpleInfo2}
		Convey("Success UpdateMetricModels \n", func() {
			mms.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).AnyTimes().Return(testMetricModel, nil)
			// mms.EXPECT().RetriveGroupIDByGroupName(gomock.Any(), gomock.Any()).AnyTimes().Return(testMetricGroup, nil)
			mms.EXPECT().GetMetricModelsByGroupID(gomock.Any(), gomock.Any()).AnyTimes().Return([]interfaces.MetricModel{}, nil)
			mms.EXPECT().UpdateMetricModelsGroup(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(int64(2), nil)
			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).AnyTimes().Return(modelMap, nil)

			reqParamByte, _ := sonic.Marshal(metricModelGroupName)

			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()

			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("Type Conversion Failed \n", func() {
			metricModelGroupName := interfaces.MetricModelGroupName{
				GroupName: "group1",
			}
			reqParamByte, _ := sonic.Marshal([]interfaces.MetricModelGroupName{metricModelGroupName})
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Get Metric Model By ID Failed \n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusNotFound,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: derrors.DataModel_MetricModel_MetricModelNotFound,
				},
			}
			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, err)

			reqParamByte, _ := sonic.Marshal(metricModelGroupName)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("Invalid Group Name \n", func() {

			metricModelGroupName := interfaces.MetricModelGroupName{
				GroupName: "group1111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111",
			}

			reqParamByte, _ := sonic.Marshal(metricModelGroupName)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Metric Model Existed \n", func() {
			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).AnyTimes().Return(map[string]interfaces.SimpleMetricModel{modelSimpleInfo.ModelID: modelSimpleInfo, "2": modelSimpleInfo}, nil)

			reqParamByte, _ := sonic.Marshal(metricModelGroupName)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Update Metric Model failed\n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: derrors.DataModel_MetricModelGroup_InternalError,
				},
			}
			mms.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).AnyTimes().Return(interfaces.MetricModel{}, nil)
			mms.EXPECT().RetriveGroupIDByGroupName(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(testMetricGroup, nil)
			mms.EXPECT().GetMetricModelsByGroupID(gomock.Any(), gomock.Any()).AnyTimes().Return([]interfaces.MetricModel{}, nil)
			mms.EXPECT().UpdateMetricModelsGroup(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(int64(0), err)
			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).AnyTimes().Return(modelMap, nil)

			reqParamByte, _ := sonic.Marshal(metricModelGroupName)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_MetricModelRestHandler_GetMetricTask(t *testing.T) {
	Convey("Test MetricModelHandler GetMetricTask\n", t, func() {
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

		handler := MockNewMetricModelRestHandler(appSetting, hydra, mms, dvs, mmgs, mmts)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/metric-tasks/1"

		Convey("Success GetMetricTask \n", func() {
			mmts.EXPECT().GetMetricTasksByTaskIDs(gomock.Any(), gomock.Any()).AnyTimes().
				Return([]interfaces.MetricTask{testTask}, nil)

			req := httptest.NewRequest(http.MethodGet, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("GetMetricTasksByIDs Failed \n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusNotFound,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: derrors.DataModel_MetricModel_MetricModelNotFound,
				},
			}

			mmts.EXPECT().GetMetricTasksByTaskIDs(gomock.Any(), gomock.Any()).AnyTimes().
				Return(nil, err)

			req := httptest.NewRequest(http.MethodGet, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("task not found \n", func() {
			mmts.EXPECT().GetMetricTasksByTaskIDs(gomock.Any(), gomock.Any()).AnyTimes().
				Return(nil, nil)

			req := httptest.NewRequest(http.MethodGet, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})
	})
}

func Test_MetricModelRestHandler_UpdateMetricTaskAttributes(t *testing.T) {
	Convey("Test MetricModelHandler UpdateMetricTaskAttributes\n", t, func() {
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

		handler := MockNewMetricModelRestHandler(appSetting, hydra, mms, dvs, mmgs, mmts)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/metric-tasks/1/attr"

		Convey("Success UpdateMetricTaskAttributes \n", func() {
			mmts.EXPECT().UpdateMetricTaskAttributes(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			reqParamByte, _ := sonic.Marshal(testTask)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("Type Conversion Failed \n", func() {
			url = "/api/mdl-data-model/v1/metric-tasks/1s/attr"
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Failed UpdateMetricTaskAttributes ShouldBind Error\n", func() {

			reqParamByte, _ := sonic.Marshal([]interfaces.MetricTask{testTask})
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("UpdateMetricTaskAttributes Failed \n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: derrors.DataModel_MetricModel_InternalError,
				},
			}
			reqParamByte, _ := sonic.Marshal(testTask)
			mmts.EXPECT().UpdateMetricTaskAttributes(gomock.Any(), gomock.Any()).AnyTimes().Return(err)

			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_MetricModelRestHandler_GetMetricModelSourceFields(t *testing.T) {
	Convey("Test MetricModelHandler GetMetricModelSourceFields\n", t, func() {
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

		handler := MockNewMetricModelRestHandler(appSetting, hydra, mms, dvs, mmgs, mmts)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/metric-models/1/fields"

		Convey("Success GetMetricModelSourceFields \n", func() {
			mms.EXPECT().GetMetricModelSourceFields(gomock.Any(), gomock.Any()).AnyTimes().
				Return([]*interfaces.ViewField{}, nil)

			req := httptest.NewRequest(http.MethodGet, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("GetMetricModelContainFilters Failed \n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusNotFound,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: derrors.DataModel_MetricModel_MetricModelNotFound,
				},
			}

			mms.EXPECT().GetMetricModelSourceFields(gomock.Any(), gomock.Any()).AnyTimes().
				Return([]*interfaces.ViewField{}, err)

			req := httptest.NewRequest(http.MethodGet, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})
	})
}
