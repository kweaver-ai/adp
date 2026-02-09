// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	rmock "github.com/kweaver-ai/kweaver-go-lib/rest/mock"
	. "github.com/smartystreets/goconvey/convey"

	"uniquery/common"
	ierrors "uniquery/errors"
	"uniquery/interfaces"
	imock "uniquery/interfaces/mock"
)

func mockNewMEventHandler(appSetting *common.AppSetting,
	hydra rest.Hydra, eService interfaces.EventService) (r *restHandler) {
	r = &restHandler{
		appSetting: appSetting,
		hydra:      hydra,
		eService:   eService,
	}
	r.InitMetric()
	return r
}

func TestQuery(t *testing.T) {
	Convey("Test Query", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydraMock := rmock.NewMockHydra(mockCtrl)
		esMock := imock.NewMockEventService(mockCtrl)
		handler := mockNewMEventHandler(appSetting, hydraMock, esMock)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/events"

		query := interfaces.EventQueryReq{
			Querys: []interfaces.EventQuery{
				{
					QueryType:           "instant_query",
					Id:                  "1",
					EnableMessageFilter: false,
				},
			},
		}
		reqParamByte, _ := json.Marshal(query)

		Convey("Bind json failed  \n", func() {
			reqParamByte1, _ := json.Marshal([]interfaces.EventQueryReq{query})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte1))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Query failed  \n", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				BaseError: rest.BaseError{
					ErrorCode: ierrors.Uniquery_EventModel_InternalError,
				},
			}

			esMock.EXPECT().Query(gomock.Any(), gomock.Any()).Return(0, interfaces.AtomicEvent{}, []interfaces.Records{}, expectedErr)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Success  \n", func() {
			esMock.EXPECT().Query(gomock.Any(), gomock.Any()).Return(1, []interfaces.AtomicEvent{}, []interfaces.Records{}, nil)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func TestQuerySingleEventByEventId(t *testing.T) {
	Convey("Test QuerySingleEventByEventId", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydraMock := rmock.NewMockHydra(mockCtrl)
		esMock := imock.NewMockEventService(mockCtrl)
		handler := mockNewMEventHandler(appSetting, hydraMock, esMock)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/event-models/498386860318393092/events/498406277530001114?start=1706600535000&end=1706603000000"

		Convey("Parse end failed  \n", func() {
			url = "/api/mdl-uniquery/v1/event-models/49/events/4?start=1706600535000&end=1706a03000000"
			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("QuerySingleEventByEventId failed  \n", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				BaseError: rest.BaseError{
					ErrorCode: ierrors.Uniquery_EventModel_InternalError,
				},
			}

			esMock.EXPECT().QuerySingleEventByEventId(gomock.Any(), gomock.Any()).AnyTimes().Return(interfaces.AtomicEvent{}, expectedErr)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Parse start failed  \n", func() {
			url = "/api/mdl-uniquery/v1/event-models/49/events/4?start=1706653a5000&end=1706603000000"
			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Success  \n", func() {
			esMock.EXPECT().QuerySingleEventByEventId(gomock.Any(), gomock.Any()).AnyTimes().Return(interfaces.AtomicEvent{}, nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}
