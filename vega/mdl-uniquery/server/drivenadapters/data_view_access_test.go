// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package drivenadapters

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	rmock "github.com/kweaver-ai/kweaver-go-lib/rest/mock"
	. "github.com/smartystreets/goconvey/convey"

	"uniquery/common"
	"uniquery/interfaces"
)

func MockNewDataViewAccess(mockCtrl *gomock.Controller) (interfaces.DataViewAccess, *rmock.MockHTTPClient) {
	mockHttpClient := rmock.NewMockHTTPClient(mockCtrl)

	dvaMock := &dataViewAccess{
		appSetting: &common.AppSetting{
			DataViewUrl: "http://localhost:13011/api/mdl-data-model/v1",
		},
		httpClient: mockHttpClient,
	}

	return dvaMock, mockHttpClient
}

func TestGetDataViewsByIDs(t *testing.T) {
	Convey("Test GetDataViewsByIDs", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		dvaMock, mockHttpClient := MockNewDataViewAccess(mockCtrl)

		Convey("GetDataViewsByIDs failed,  func GetNoUnmarshal error", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, fmt.Errorf("method failed"))

			views, err := dvaMock.GetDataViewsByIDs(testCtx, "1,2", false)

			So(len(views), ShouldEqual, 0)
			So(err, ShouldNotBeNil)
		})

		Convey("GetDataViewsByIDs failed, data view not found", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusNotFound, nil, nil)

			views, err := dvaMock.GetDataViewsByIDs(testCtx, "1,2", false)

			So(len(views), ShouldEqual, 0)
			So(err, ShouldNotBeNil)
		})

		Convey("get data view failed because 400", func() {
			bytes, _ := sonic.Marshal(rest.BaseError{ErrorCode: "dddd",
				Description:  "400",
				ErrorLink:    "",
				Solution:     "",
				ErrorDetails: "failed",
			})
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusBadRequest, bytes, nil)

			views, err := dvaMock.GetDataViewsByIDs(testCtx, "1,2", false)

			So(len(views), ShouldEqual, 0)
			So(err, ShouldNotBeNil)
		})

		Convey("get data view failed because 500", func() {
			bytes, _ := sonic.Marshal(rest.BaseError{ErrorCode: "as",
				Description:  "500",
				ErrorLink:    "",
				Solution:     "",
				ErrorDetails: "failed",
			})
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusInternalServerError, bytes, nil)

			views, err := dvaMock.GetDataViewsByIDs(testCtx, "1,2", false)

			So(len(views), ShouldEqual, 0)
			So(err, ShouldNotBeNil)
		})

		Convey("get data view failed because 500 && unmarshal to baseError error", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusInternalServerError, []uint8{}, nil)

			views, err := dvaMock.GetDataViewsByIDs(testCtx, "1,2", false)

			So(len(views), ShouldEqual, 0)
			So(err, ShouldNotBeNil)
		})

		Convey("response unmarshal to views failed", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, []uint8{}, nil)

			views, err := dvaMock.GetDataViewsByIDs(testCtx, "1,2", false)

			So(len(views), ShouldEqual, 0)
			So(err, ShouldNotBeNil)
		})

		Convey("success ", func() {
			views := []*interfaces.DataView{
				{
					ViewID: "1",
				},
				{
					ViewID: "2",
				},
			}

			bytes, _ := sonic.Marshal(views)
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, bytes, nil)

			views, err := dvaMock.GetDataViewsByIDs(testCtx, "1,2", false)

			So(len(views), ShouldEqual, 2)
			So(err, ShouldBeNil)
		})

	})
}

func TestGetDataViewIDByName(t *testing.T) {
	Convey("Test GetDataViewIDByName", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		dvaMock, mockHttpClient := MockNewDataViewAccess(mockCtrl)

		name := "dip-event-task-data-view"

		Convey("GetDataViewIDByName failed,  func GetNoUnmarshal error", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, fmt.Errorf("method failed"))

			viewID, err := dvaMock.GetDataViewIDByName(testCtx, name)

			So(viewID, ShouldEqual, "")
			So(err, ShouldNotBeNil)
		})

		Convey("GetDataViewIDByName failed, data view not found", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusNotFound, nil, nil)

			viewID, err := dvaMock.GetDataViewIDByName(testCtx, name)

			So(viewID, ShouldEqual, "")
			So(err, ShouldNotBeNil)
		})

		Convey("get data view failed because 400", func() {
			bytes, _ := sonic.Marshal(rest.BaseError{ErrorCode: "dddd",
				Description:  "400",
				ErrorLink:    "",
				Solution:     "",
				ErrorDetails: "failed",
			})
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusBadRequest, bytes, nil)

			viewID, err := dvaMock.GetDataViewIDByName(testCtx, name)

			So(viewID, ShouldEqual, "")
			So(err, ShouldNotBeNil)
		})

		Convey("get data view failed because 500", func() {
			bytes, _ := sonic.Marshal(rest.BaseError{ErrorCode: "as",
				Description:  "500",
				ErrorLink:    "",
				Solution:     "",
				ErrorDetails: "failed",
			})
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusInternalServerError, bytes, nil)

			viewID, err := dvaMock.GetDataViewIDByName(testCtx, name)

			So(viewID, ShouldEqual, "")
			So(err, ShouldNotBeNil)
		})

		Convey("get data view failed because 500 && unmarshal to baseError error", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusInternalServerError, []uint8{}, nil)

			viewID, err := dvaMock.GetDataViewIDByName(testCtx, name)

			So(viewID, ShouldEqual, "")
			So(err, ShouldNotBeNil)
		})

		Convey("response unmarshal  failed", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, []uint8{}, nil)

			viewID, err := dvaMock.GetDataViewIDByName(testCtx, name)

			So(viewID, ShouldEqual, "")
			So(err, ShouldNotBeNil)
		})

		Convey("result is empty ", func() {
			result := struct {
				Entries []interfaces.DataView `json:"entries"`
				Total   int                   `json:"total_count"`
			}{
				[]interfaces.DataView{},
				0,
			}

			bytes, _ := sonic.Marshal(result)
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, bytes, nil)

			viewID, err := dvaMock.GetDataViewIDByName(testCtx, name)

			So(viewID, ShouldEqual, "")
			So(err, ShouldNotBeNil)
		})

		Convey("success ", func() {
			result := struct {
				Entries []interfaces.DataView `json:"entries"`
				Total   int                   `json:"total_count"`
			}{
				[]interfaces.DataView{
					{
						ViewID: "1",
					},
				},
				1,
			}

			bytes, _ := sonic.Marshal(result)
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, bytes, nil)

			viewID, err := dvaMock.GetDataViewIDByName(testCtx, name)

			So(viewID, ShouldEqual, "1")
			So(err, ShouldBeNil)
		})

	})
}
