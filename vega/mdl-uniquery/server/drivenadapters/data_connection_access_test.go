// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package drivenadapters

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/bytedance/sonic"
	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	rmock "github.com/kweaver-ai/kweaver-go-lib/rest/mock"
	. "github.com/smartystreets/goconvey/convey"

	"uniquery/common"
)

func TestGetDataConnectionByID(t *testing.T) {
	Convey("Test GetDataConnectionByID", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockHttpClient := rmock.NewMockHTTPClient(mockCtrl)
		dcAccess := &dataConnectionAccess{
			appSetting: &common.AppSetting{},
			httpClient: mockHttpClient,
		}

		Convey("Get failed, caused by the error from method 'GetNoUnmarshal'", func() {
			expectedErr := fmt.Errorf("some errors")

			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, expectedErr)

			_, _, err := dcAccess.GetDataConnectionByID(testCtx, "1")
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by the data connection is not found", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusNotFound, nil, nil)

			_, isExist, err := dcAccess.GetDataConnectionByID(testCtx, "1")
			So(err, ShouldBeNil)
			So(isExist, ShouldBeFalse)
		})

		Convey("Get failed, caused by the respCode from method 'GetNoUnmarshal'", func() {
			expectedRespData, _ := json.Marshal(rest.BaseError{ErrorCode: "as",
				Description:  "500",
				ErrorLink:    "",
				Solution:     "",
				ErrorDetails: "failed",
			})
			expectedErr := fmt.Errorf("Failed to get data connection by http client: %v", string(expectedRespData))

			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusInternalServerError, expectedRespData, nil)

			_, _, err := dcAccess.GetDataConnectionByID(testCtx, "1")
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by the error from func 'sonic.Unmarshal'", func() {
			expectedErr := errors.New("some errors")

			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, nil)

			patch := ApplyFuncReturn(sonic.Unmarshal, expectedErr)
			defer patch.Reset()

			_, _, err := dcAccess.GetDataConnectionByID(testCtx, "1")
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get succeed", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, nil)

			patch := ApplyFuncReturn(sonic.Unmarshal, nil)
			defer patch.Reset()

			_, isExist, err := dcAccess.GetDataConnectionByID(testCtx, "1")
			So(err, ShouldResemble, nil)
			So(isExist, ShouldBeTrue)
		})
	})
}

func TestGetDataConnectionTypeByName(t *testing.T) {
	Convey("Test GetDataConnectionTypeByName", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockHttpClient := rmock.NewMockHTTPClient(mockCtrl)
		dcAccess := &dataConnectionAccess{
			appSetting: &common.AppSetting{},
			httpClient: mockHttpClient,
		}

		Convey("Get failed, caused by the error from method 'GetNoUnmarshal'", func() {
			expectedErr := fmt.Errorf("some errors")

			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, expectedErr)

			_, _, err := dcAccess.GetDataConnectionTypeByName(testCtx, "1")
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by the respCode from method 'GetNoUnmarshal'", func() {
			expectedRespData, _ := json.Marshal(rest.BaseError{ErrorCode: "as",
				Description:  "500",
				ErrorLink:    "",
				Solution:     "",
				ErrorDetails: "failed",
			})
			expectedErr := fmt.Errorf("Failed to get data connection type by http client: %v", string(expectedRespData))

			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusInternalServerError, expectedRespData, nil)

			_, _, err := dcAccess.GetDataConnectionTypeByName(testCtx, "1")
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by the error from func 'sonic.Unmarshal'", func() {
			expectedErr := errors.New("some errors")

			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, nil)

			patch := ApplyFuncReturn(sonic.Unmarshal, expectedErr)
			defer patch.Reset()

			_, _, err := dcAccess.GetDataConnectionTypeByName(testCtx, "1")
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by no results were found", func() {
			expectedRespData := []byte("{}")

			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, expectedRespData, nil)

			_, isExist, err := dcAccess.GetDataConnectionTypeByName(testCtx, "1")
			So(err, ShouldResemble, nil)
			So(isExist, ShouldBeFalse)
		})

		Convey("Get succeed", func() {
			connLists := struct {
				Entries []struct {
					DataSourceType string `json:"data_source_type"`
				} `json:"entries"`
			}{
				Entries: []struct {
					DataSourceType string `json:"data_source_type"`
				}{
					{
						DataSourceType: "type1",
					},
				},
			}
			expectedRespData, _ := sonic.Marshal(connLists)

			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, expectedRespData, nil)

			_, isExist, err := dcAccess.GetDataConnectionTypeByName(testCtx, "1")
			So(err, ShouldResemble, nil)
			So(isExist, ShouldBeTrue)
		})
	})
}
