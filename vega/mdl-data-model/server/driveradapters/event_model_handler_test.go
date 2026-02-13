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
	testUpdateTime = int64(1735786555379)
	detectRule     = interfaces.DetectRule{
		DetectRuleID: "1000",
		Priority:     99,
		Type:         "range_detect",
		Formula: []interfaces.FormulaItem{
			{
				Level: 1,
				Filter: interfaces.LogicFilter{
					LogicOperator: "",
					FilterExpress: interfaces.FilterExpress{Name: "values", Value: []float64{0.9, 1.0}, Operation: "range"},
					Children:      []interfaces.LogicFilter{},
				},
			},
			{
				Level: 2,
				Filter: interfaces.LogicFilter{
					LogicOperator: "",
					FilterExpress: interfaces.FilterExpress{Name: "values", Value: []float64{0.9, 1.0}, Operation: "range"},
					Children:      []interfaces.LogicFilter{},
				},
			},
		},
	}

	eventModel = interfaces.EventModelCreateRequest{
		EventModelRequest: interfaces.EventModelRequest{
			EventModelName: "测试中的名称",
			EventModelType: "atomic",
			EventModelTags: []string{"xx1", "xx2"},
			DataSourceType: "metric_model",
			DataSource:     []string{"1"},
			DetectRule:     detectRule,
			AggregateRule:  interfaces.AggregateRule{},
			DefaultTimeWindow: interfaces.TimeInterval{
				Interval: 5,
				Unit:     "m",
			},
			EventModelComment: ""},
		IsActive: 1,
		IsCustom: 1,
	}

	oldEventModel = interfaces.EventModel{
		EventModelID:      "1",
		EventModelName:    "测试中的名称",
		EventModelType:    "atomic",
		UpdateTime:        testUpdateTime,
		EventModelTags:    []string{"xx1", "xx2"},
		DataSourceType:    "metric_model",
		DataSource:        []string{"1"},
		DetectRule:        detectRule,
		DefaultTimeWindow: interfaces.TimeInterval{Interval: 5, Unit: "m"},
		EventModelComment: "",
		IsActive:          -1,
		IsCustom:          -1,
	}
	newEventModel = interfaces.EventModelUpateRequest{
		EventModelRequest: interfaces.EventModelRequest{
			EventModelName: "测试中的名称_修改",
			EventModelType: "atomic",
			// UpdateTime:        "2022-12-13 01:01:01",
			EventModelTags: []string{"xx1", "xx2", "xx3"},
			DataSourceType: "metric_model",
			DataSource:     []string{"2"},
			DetectRule:     detectRule,
			DefaultTimeWindow: interfaces.TimeInterval{
				Interval: 5,
				Unit:     "m",
			},
			EventModelComment: "modify_comment",
		},
		EventModelID: "1",
		IsActive:     1,
	}
)

func MockNewEventModelRestHandler(appSetting *common.AppSetting,
	hydra rest.Hydra,
	ems interfaces.EventModelService,
	mms interfaces.MetricModelService) (r *restHandler) {
	r = &restHandler{
		appSetting: appSetting,
		hydra:      hydra,
		ems:        ems,
		mms:        mms,
	}
	return r
}

func Test_EventModelRestHandler_CreateEventModels(t *testing.T) {
	Convey("Test EventModelHandler CreateEventModels\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		ems := dmock.NewMockEventModelService(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)

		handler := MockNewEventModelRestHandler(appSetting, hydra, ems, mms)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/event-models"

		Convey("Success CreateEventModels \n", func() {
			//dmock 不需要实际请求的方法
			ems.EXPECT().EventModelCreateValidate(gomock.Any(), gomock.Any()).AnyTimes().Return(oldEventModel, nil)
			ems.EXPECT().CreateEventModels(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, nil)
			// eventModel.
			reqParamByte, _ := sonic.Marshal([]interfaces.EventModelCreateRequest{eventModel})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("Failed CreateEventModels contentType is not json\n", func() {
			reqParamByte, _ := sonic.Marshal([]interfaces.EventModelCreateRequest{eventModel})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotAcceptable)
		})

		Convey("Failed CreateEventModels Name contail illegal char\n", func() {
			ems.EXPECT().EventModelCreateValidate(gomock.Any(), gomock.Any()).AnyTimes().Return(oldEventModel, nil)
			ems.EXPECT().CreateEventModels(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, nil)
			eventModelTT := eventModel
			eventModelTT.EventModelName = "{}"

			reqParamByte, _ := sonic.Marshal([]interfaces.EventModelCreateRequest{eventModelTT})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("unsupport event type\n", func() {
			// ems.EXPECT().EventModelCreateValidate(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			// ems.EXPECT().CreateEventModel(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			eventModelT := eventModel
			eventModelT.EventModelType = "atomic_agg"

			reqParamByte, _ := sonic.Marshal([]interfaces.EventModelCreateRequest{eventModelT})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("tags size exceed 5\n", func() {
			eventModel3 := eventModel
			// eventModelT.EventModelType = "atomic_agg"
			eventModel3.EventModelTags = []string{"1", "2", "3", "4", "5", "6"}

			reqParamByte, _ := sonic.Marshal([]interfaces.EventModelCreateRequest{eventModel3})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("name length exceed 40\n", func() {
			eventModel4 := eventModel
			// eventModelT.EventModelType = "atomic_agg"
			eventModel4.EventModelName = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

			reqParamByte, _ := sonic.Marshal([]interfaces.EventModelCreateRequest{eventModel4})
			// reqParamByte, _ := sonic.Marshal(eventModel4)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("tag contain illegal charater\n", func() {
			eventModel5 := eventModel
			// eventModelT.EventModelType = "atomic_agg"
			eventModel5.EventModelTags = []string{"{}", ",#[]"}

			reqParamByte, _ := sonic.Marshal([]interfaces.EventModelCreateRequest{eventModel5})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		// Convey("Failed CreateValidate failed\n", func() {
		// 	ems.EXPECT().EventModelCreateValidate(gomock.Any(), gomock.Any()).AnyTimes().Return(interfaces.EventModel{}, rest.NewHTTPError(context.TODO(), http.StatusInternalServerError,
		// 		derrors.EventModel_InternalError))
		// 	// ems.EXPECT().CreateEventModel(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
		// 	// eventModel.EventModelName = "{}"

		// 	reqParamByte, _ := sonic.Marshal([]interfaces.EventModelCreateRequest{eventModel})
		// 	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
		// 	req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
		// 	w := httptest.NewRecorder()
		// 	engine.ServeHTTP(w, req)

		// 	So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		// })
	})
}

func Test_EventModelRestHandler_UpdateEventModels(t *testing.T) {
	Convey("Test EventModelHandler UpdateEventModels\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		ems := dmock.NewMockEventModelService(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)

		handler := MockNewEventModelRestHandler(appSetting, hydra, ems, mms)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/event-models/1"

		Convey("Success UpdateEventModels \n", func() {
			//dmock 不需要实际请求的方法
			ems.EXPECT().EventModelUpdateValidate(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			ems.EXPECT().UpdateEventModel(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			// ems.EXPECT().GetEventModelByID(gomock.Any(), gomock.Any()).AnyTimes().Return(oldEventModel, nil)

			reqParamByte, _ := sonic.Marshal(newEventModel)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("Failed UpdateEventModels with bad name \n", func() {
			//dmock 不需要实际请求的方法
			ems.EXPECT().EventModelUpdateValidate(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			ems.EXPECT().UpdateEventModel(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			// ems.EXPECT().GetEventModelByID(gomock.Any(), gomock.Any()).AnyTimes().Return(oldEventModel, nil)

			newEventModel1 := newEventModel
			newEventModel1.EventModelName = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
			reqParamByte, _ := sonic.Marshal(newEventModel1)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Failed UpdateEventModels with bad tags \n", func() {
			//dmock 不需要实际请求的方法
			ems.EXPECT().EventModelUpdateValidate(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			ems.EXPECT().UpdateEventModel(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			// ems.EXPECT().GetEventModelByID(gomock.Any(), gomock.Any()).AnyTimes().Return(oldEventModel, nil)

			newEventModel1 := newEventModel
			newEventModel1.EventModelTags = []string{"hello", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
			reqParamByte, _ := sonic.Marshal(newEventModel1)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Failed UpdateEventModels with illegal tags \n", func() {
			//dmock 不需要实际请求的方法
			ems.EXPECT().EventModelUpdateValidate(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			ems.EXPECT().UpdateEventModel(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			// ems.EXPECT().GetEventModelByID(gomock.Any(), gomock.Any()).AnyTimes().Return(oldEventModel, nil)

			newEventModel4 := newEventModel
			newEventModel4.EventModelTags = []string{"hello", "{}[]/#"}
			reqParamByte, _ := sonic.Marshal(newEventModel4)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Failed UpdateEventModels with bad data source type \n", func() {
			//dmock 不需要实际请求的方法
			ems.EXPECT().EventModelUpdateValidate(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			ems.EXPECT().UpdateEventModel(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			// ems.EXPECT().GetEventModelByID(gomock.Any(), gomock.Any()).AnyTimes().Return(oldEventModel, nil)
			newEventModel2 := newEventModel
			newEventModel2.DataSourceType = "hello"

			reqParamByte, _ := sonic.Marshal(newEventModel2)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Failed UpdateEventModels validate error \n", func() {
			//dmock 不需要实际请求的方法
			ems.EXPECT().EventModelUpdateValidate(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.NewHTTPError(context.TODO(), http.StatusInternalServerError,
				derrors.EventModel_InternalError))
			ems.EXPECT().UpdateEventModel(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.NewHTTPError(context.TODO(), http.StatusInternalServerError,
				derrors.EventModel_InternalError))
			// ems.EXPECT().GetEventModelByID(gomock.Any(), gomock.Any()).AnyTimes().Return(oldEventModel, nil)

			reqParamByte, _ := sonic.Marshal(newEventModel)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Failed UpdateEventModels inter error \n", func() {
			//dmock 不需要实际请求的方法
			ems.EXPECT().EventModelUpdateValidate(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			ems.EXPECT().UpdateEventModel(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.NewHTTPError(context.TODO(), http.StatusInternalServerError,
				derrors.EventModel_InternalError))
			// ems.EXPECT().GetEventModelByID(gomock.Any(), gomock.Any()).AnyTimes().Return(oldEventModel, nil)

			reqParamByte, _ := sonic.Marshal(newEventModel)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Failed UpdateEventModels Not Found \n", func() {
			//dmock 不需要实际请求的方法
			ems.EXPECT().EventModelUpdateValidate(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			ems.EXPECT().UpdateEventModel(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.NewHTTPError(context.TODO(), http.StatusNotFound,
				derrors.EventModel_EventModelNotFound))
			// ems.EXPECT().GetEventModelByID(gomock.Any(), gomock.Any()).AnyTimes().Return(oldEventModel, nil)

			reqParamByte, _ := sonic.Marshal(newEventModel)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})
	})
}

func Test_EventModelRestHandler_DeleteEventModels(t *testing.T) {
	Convey("Test EventModelHandler DeleteEventModels\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		ems := dmock.NewMockEventModelService(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)

		handler := MockNewEventModelRestHandler(appSetting, hydra, ems, mms)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/event-models/1,2"

		var models []interfaces.EventModel
		models = append(models, oldEventModel)

		Convey("Success Delete EventModels \n", func() {
			//dmock 不需要实际请求的方法
			ems.EXPECT().DeleteEventModels(gomock.Any(), gomock.Any()).AnyTimes().Return(models, nil)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("Failed DeleteEventModels error \n", func() {
			//dmock 不需要实际请求的方法
			ems.EXPECT().DeleteEventModels(gomock.Any(), gomock.Any()).AnyTimes().Return(models, rest.NewHTTPError(context.TODO(), http.StatusInternalServerError,
				derrors.EventModel_InternalError))
			// ems.EXPECT().GetEventModelByID(gomock.Any(), gomock.Any()).AnyTimes().Return(oldEventModel, nil)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Failed DeleteEventModels not found \n", func() {
			//dmock 不需要实际请求的方法
			ems.EXPECT().DeleteEventModels(gomock.Any(), gomock.Any()).AnyTimes().Return(models, rest.NewHTTPError(context.TODO(), http.StatusNotFound,
				derrors.EventModel_EventModelNotFound))
			// ems.EXPECT().GetEventModelByID(gomock.Any(), gomock.Any()).AnyTimes().Return(oldEventModel, nil)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})
	})
}

func Test_EventModelRestHandler_QueryEventModels(t *testing.T) {
	Convey("Test EventModelHandler QueryEventModels\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		ems := dmock.NewMockEventModelService(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)

		handler := MockNewEventModelRestHandler(appSetting, hydra, ems, mms)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/event-models"

		var models []interfaces.EventModel
		models = append(models, oldEventModel)

		Convey("Success Query EventModels \n", func() {
			//dmock 不需要实际请求的方法
			ems.EXPECT().QueryEventModels(gomock.Any(), gomock.Any()).AnyTimes().Return(models, 1, nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("Failed QueryEventModels error \n", func() {
			//dmock 不需要实际请求的方法
			ems.EXPECT().QueryEventModels(gomock.Any(), gomock.Any()).AnyTimes().Return(models, 0, rest.NewHTTPError(context.TODO(), http.StatusInternalServerError,
				derrors.EventModel_InternalError))

			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Failed QueryEventModels param error \n", func() {
			//dmock 不需要实际请求的方法
			ems.EXPECT().QueryEventModels(gomock.Any(), gomock.Any()).AnyTimes().Return(models, 0, nil)

			url = url + "?direction=desc&sort=update_time&limit=1000&offset=0"
			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})
	})
}

func Test_EventModelRestHandler_QueryEventModelByID(t *testing.T) {
	Convey("Test EventModelHandler QueryEventModelByID\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		ems := dmock.NewMockEventModelService(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)

		handler := MockNewEventModelRestHandler(appSetting, hydra, ems, mms)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/event-models/1"

		Convey("Success Query EventModels \n", func() {
			//dmock 不需要实际请求的方法
			ems.EXPECT().GetEventModelByID(gomock.Any(), gomock.Any()).AnyTimes().Return(oldEventModel, nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("Failed QueryEventModels param error \n", func() {
			//dmock 不需要实际请求的方法
			ems.EXPECT().GetEventModelByID(gomock.Any(), gomock.Any()).AnyTimes().Return(oldEventModel, rest.NewHTTPError(context.TODO(), http.StatusNotFound,
				derrors.EventModel_EventModelNotFound))

			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})
	})
}

func Test_EventModelRestHandler_QueryEventLevel(t *testing.T) {
	Convey("Test EventModelHandler QueryEventLevel\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		ems := dmock.NewMockEventModelService(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)

		handler := MockNewEventModelRestHandler(appSetting, hydra, ems, mms)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/event-level"

		Convey("Success Query EventModels zh-CN \n", func() {
			//dmock 不需要实际请求的方法
			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set("Accept-Language", "zh-CN")

			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("Success Query EventModels en-US \n", func() {
			//dmock 不需要实际请求的方法
			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set("Accept-Language", "en-US")
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)

			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			fmt.Println(w.Result().Body)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func Test_EventModelRestHandler_UpdateEventTaskStatus(t *testing.T) {
	Convey("Test EventModelHandler UpdateEventTaskStatus\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		ems := dmock.NewMockEventModelService(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)

		handler := MockNewEventModelRestHandler(appSetting, hydra, ems, mms)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/event-task/1/attr"

		reqParam := interfaces.EventTask{
			TaskStatus:       4,
			StatusUpdateTime: testUpdateTime,
		}

		Convey("UpdateEventTaskStatus failed,caused by binding json failed \n", func() {
			reqParamByte, _ := sonic.Marshal([]interfaces.EventTask{reqParam})
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set("Accept-Language", "en-US")
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("UpdateEventTaskStatus failed,caused by UpdateEventTaskAttributes failed \n", func() {
			reqParamByte, _ := sonic.Marshal(reqParam)
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: derrors.EventModel_InternalError,
				},
			}

			ems.EXPECT().UpdateEventTaskAttributes(gomock.Any(), gomock.Any()).Return(expectedErr)

			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set("Accept-Language", "en-US")
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("UpdateEventTaskStatus successed \n", func() {
			reqParamByte, _ := sonic.Marshal(reqParam)

			ems.EXPECT().UpdateEventTaskAttributes(gomock.Any(), gomock.Any()).Return(nil)

			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set("Accept-Language", "en-US")
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})
	})
}
