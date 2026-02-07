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

	"uniquery/common/condition"
	"uniquery/interfaces"
)

func TestGetVegaViewFieldsByID(t *testing.T) {
	Convey("Test GetVegaViewFieldsByID", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockHttpClient := rmock.NewMockHTTPClient(mockCtrl)
		vvAccess := &vegaViewAccess{httpClient: mockHttpClient}

		var expect interfaces.VegaViewWithFields
		Convey("get request method failed", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, fmt.Errorf("method failed"))

			vegaView, err := vvAccess.GetVegaViewFieldsByID(testCtx, "123")

			So(vegaView, ShouldResemble, expect)
			So(err, ShouldResemble, fmt.Errorf("get Vega View Fields request failed: method failed"))
		})

		Convey("get metric model failed because unmarshal", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusInternalServerError, nil, nil)

			vegaView, err := vvAccess.GetVegaViewFieldsByID(testCtx, "123")

			So(vegaView, ShouldResemble, expect)
			So(err, ShouldNotBeNil)
		})

		Convey("get metric model failed because 500", func() {
			bytes, _ := sonic.Marshal(rest.BaseError{ErrorCode: "as",
				Description:  "500",
				ErrorLink:    "",
				Solution:     "",
				ErrorDetails: "failed",
			})
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusInternalServerError, bytes, nil)

			vegaView, err := vvAccess.GetVegaViewFieldsByID(testCtx, "123")

			So(vegaView, ShouldResemble, expect)
			So(err.Error(), ShouldEqual, "get Vega View Error: {\"error_code\":\"\",\"description\":\"500\",\"solution\":\"\",\"error_link\":\"\",\"error_details\":null}")
		})

		Convey("response nil ", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, nil)

			vegaView, err := vvAccess.GetVegaViewFieldsByID(testCtx, "123")
			So(vegaView, ShouldResemble, expect)
			So(err, ShouldBeNil)
		})

		Convey("response unmarshal failed ", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, []uint8{}, nil)

			vegaView, err := vvAccess.GetVegaViewFieldsByID(testCtx, "123")

			So(vegaView, ShouldResemble, expect)
			So(err, ShouldNotBeNil)
		})

		Convey("success ", func() {
			vegaViewE := interfaces.VegaViewWithFields{
				Catalog: "a",
				Fields: []interfaces.VegaViewField{
					{
						Name: "f1",
						Type: "varchar",
					},
					{
						Name: "f2",
						Type: "varchar",
					},
				},
			}

			bytes, _ := sonic.Marshal(vegaViewE)
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, bytes, nil)

			vegaView, err := vvAccess.GetVegaViewFieldsByID(testCtx, "123")

			So(vegaView, ShouldResemble, vegaViewE)
			So(err, ShouldBeNil)
		})
	})
}

func TestFetchDatasFromVega(t *testing.T) {
	Convey("Test FetchDatasFromVega", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockHttpClient := rmock.NewMockHTTPClient(mockCtrl)
		vvAccess := &vegaViewAccess{httpClient: mockHttpClient}

		var expect interfaces.VegaFetchData
		Convey("get request method failed", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, fmt.Errorf("method failed"))

			vegaView, err := vvAccess.FetchDatasFromVega(testCtx, "a", "123")

			So(vegaView, ShouldResemble, expect)
			So(err, ShouldResemble, fmt.Errorf("fetch data from vega gateway failed: method failed"))
		})

		Convey("get metric model failed because unmarshal", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusInternalServerError, nil, nil)

			vegaView, err := vvAccess.FetchDatasFromVega(testCtx, "a", "123")

			So(vegaView, ShouldResemble, expect)
			So(err, ShouldNotBeNil)
		})

		Convey("get metric model failed because 500", func() {
			bytes, _ := sonic.Marshal(rest.BaseError{ErrorCode: "as",
				Description:  "500",
				ErrorLink:    "",
				Solution:     "",
				ErrorDetails: "failed",
			})
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusInternalServerError, bytes, nil)

			vegaView, err := vvAccess.FetchDatasFromVega(testCtx, "a", "123")

			So(vegaView, ShouldResemble, expect)
			So(err.Error(), ShouldEqual, "fetch data from vega gateway Error: {\"error_code\":\"\",\"description\":\"500\",\"solution\":\"\",\"error_link\":\"\",\"error_details\":null}")
		})

		Convey("response nil ", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, nil)

			vegaView, err := vvAccess.FetchDatasFromVega(testCtx, "a", "123")
			So(vegaView, ShouldResemble, expect)
			So(err, ShouldBeNil)
		})

		Convey("response unmarshal failed ", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, []uint8{}, nil)

			vegaView, err := vvAccess.FetchDatasFromVega(testCtx, "a", "123")

			So(vegaView, ShouldResemble, expect)
			So(err, ShouldNotBeNil)
		})

		Convey("success ", func() {
			vegaData := interfaces.VegaFetchData{
				Data: [][]any{
					{
						[]any{"a", 2.2},
					},
					{
						[]any{"b", 3.2},
					},
				},
				Columns: []condition.ViewField{
					{
						Name: "f1",
						Type: "varchar",
					},
					{
						Name: "f2",
						Type: "varchar",
					},
				},
			}

			bytes, _ := sonic.Marshal(vegaData)
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, bytes, nil)

			vegaView, err := vvAccess.FetchDatasFromVega(testCtx, "a", "123")

			So(vegaView, ShouldResemble, vegaData)
			So(err, ShouldBeNil)
		})

		Convey("success with uri is not empty ", func() {
			vegaData := interfaces.VegaFetchData{
				Data: [][]any{
					{
						[]any{"a", 2.2},
					},
					{
						[]any{"b", 3.2},
					},
				},
				Columns: []condition.ViewField{
					{
						Name: "f1",
						Type: "varchar",
					},
					{
						Name: "f2",
						Type: "varchar",
					},
				},
			}

			bytes, _ := sonic.Marshal(vegaData)
			mockHttpClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, bytes, nil)

			vegaView, err := vvAccess.FetchDatasFromVega(testCtx, "", "123")

			So(vegaView, ShouldResemble, vegaData)
			So(err, ShouldBeNil)
		})
	})
}
