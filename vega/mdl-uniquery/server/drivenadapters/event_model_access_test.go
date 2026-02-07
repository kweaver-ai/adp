// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package drivenadapters

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	rmock "github.com/kweaver-ai/kweaver-go-lib/rest/mock"
	. "github.com/smartystreets/goconvey/convey"

	"uniquery/common"
	"uniquery/interfaces"
)

func MockNewEventModelAccess(mockCtrl *gomock.Controller) (interfaces.EventModelAccess, *rmock.MockHTTPClient) {
	mockHttpClient := rmock.NewMockHTTPClient(mockCtrl)

	emaMock := &eventModelAccess{
		appSetting: &common.AppSetting{
			EventModelUrl: "http://localhost:13011/api/data-model/v1",
		},
		httpClient: mockHttpClient,
	}

	return emaMock, mockHttpClient
}

func TestGetEventModelByID(t *testing.T) {
	Convey("Test GetEventModelsByID", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		emaMock, mockHttpClient := MockNewEventModelAccess(mockCtrl)

		Convey("GetEventModelsByID failed,  func GetNoUnmarshal error", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, fmt.Errorf("method failed"))

			models, err := emaMock.GetEventModelById(testCtx, "1")

			So(len(models), ShouldEqual, 0)
			So(err, ShouldNotBeNil)
		})

		Convey("GetEventModelsByID failed, event model not found", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusNotFound, nil, nil)

			models, err := emaMock.GetEventModelById(testCtx, "1")

			So(len(models), ShouldEqual, 0)
			So(err, ShouldNotBeNil)
		})

		Convey("get event model failed because 400", func() {
			bytes, _ := json.Marshal(rest.BaseError{ErrorCode: "dddd",
				Description:  "400",
				ErrorLink:    "",
				Solution:     "",
				ErrorDetails: "failed",
			})
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusBadRequest, bytes, nil)

			models, err := emaMock.GetEventModelById(testCtx, "1")

			So(len(models), ShouldEqual, 0)
			So(err, ShouldNotBeNil)
		})

		Convey("get event model failed because 500", func() {
			bytes, _ := json.Marshal(rest.BaseError{ErrorCode: "as",
				Description:  "500",
				ErrorLink:    "",
				Solution:     "",
				ErrorDetails: "failed",
			})
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusInternalServerError, bytes, nil)

			models, err := emaMock.GetEventModelById(testCtx, "1")

			So(len(models), ShouldEqual, 0)
			So(err, ShouldNotBeNil)
		})

		Convey("get event model failed because 500 && unmarshal to baseError error", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusInternalServerError, []uint8{}, nil)

			models, err := emaMock.GetEventModelById(testCtx, "1")

			So(len(models), ShouldEqual, 0)
			So(err, ShouldNotBeNil)
		})

		Convey("response unmarshal to models failed", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, []uint8{}, nil)

			models, err := emaMock.GetEventModelById(testCtx, "1")

			So(len(models), ShouldEqual, 0)
			So(err, ShouldNotBeNil)
		})

		Convey("success ", func() {
			models := []interfaces.EventModel{
				{
					EventModelID: "1",
				},
			}

			bytes, _ := json.Marshal(models)
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, bytes, nil)

			models, err := emaMock.GetEventModelById(testCtx, "1")

			So(len(models), ShouldEqual, 1)
			So(err, ShouldBeNil)
		})

	})
}

func TestGetEventModelByViewId(t *testing.T) {
	Convey("Test GetEventModelByViewId", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		emaMock, mockHttpClient := MockNewEventModelAccess(mockCtrl)

		Convey("GetEventModelByViewId failed,  func GetNoUnmarshal error", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, fmt.Errorf("method failed"))

			models, err := emaMock.GetEventModelBySourceId(testCtx, "1")

			So(len(models), ShouldEqual, 0)
			So(err, ShouldNotBeNil)
		})

		Convey("GetEventModelByViewId failed, event model not found", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusNotFound, nil, nil)

			models, err := emaMock.GetEventModelBySourceId(testCtx, "1")

			So(len(models), ShouldEqual, 0)
			So(err, ShouldNotBeNil)
		})

		Convey("get event model failed because 400", func() {
			bytes, _ := json.Marshal(rest.BaseError{ErrorCode: "dddd",
				Description:  "400",
				ErrorLink:    "",
				Solution:     "",
				ErrorDetails: "failed",
			})
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusBadRequest, bytes, nil)

			models, err := emaMock.GetEventModelBySourceId(testCtx, "1")

			So(len(models), ShouldEqual, 0)
			So(err, ShouldNotBeNil)
		})

		Convey("get event model failed because 500", func() {
			bytes, _ := json.Marshal(rest.BaseError{ErrorCode: "as",
				Description:  "500",
				ErrorLink:    "",
				Solution:     "",
				ErrorDetails: "failed",
			})
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusInternalServerError, bytes, nil)

			models, err := emaMock.GetEventModelBySourceId(testCtx, "1")

			So(len(models), ShouldEqual, 0)
			So(err, ShouldNotBeNil)
		})

		Convey("get event model failed because 500 && unmarshal to baseError error", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusInternalServerError, []uint8{}, nil)

			models, err := emaMock.GetEventModelBySourceId(testCtx, "1")

			So(len(models), ShouldEqual, 0)
			So(err, ShouldNotBeNil)
		})

		Convey("response unmarshal to models failed", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, []uint8{}, nil)

			models, err := emaMock.GetEventModelBySourceId(testCtx, "1")

			So(len(models), ShouldEqual, 0)
			So(err, ShouldNotBeNil)
		})

		Convey("success ", func() {
			models := EventModelRecords{
				Entries: []interfaces.EventModel{
					{
						EventModelID: "1",
						DataSource:   []string{"1"},
					},
				},
				Total: 1,
			}

			bytes, _ := json.Marshal(models)
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, bytes, nil)

			result, err := emaMock.GetEventModelBySourceId(testCtx, "1")

			So(len(result), ShouldEqual, 1)
			So(err, ShouldBeNil)
		})

	})
}
