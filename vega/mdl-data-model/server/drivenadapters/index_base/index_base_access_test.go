// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package index_base

import (
	"context"
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

	"data-model/common"
	"data-model/interfaces"
)

var (
	testCtx = context.WithValue(context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage),
		interfaces.ACCOUNT_INFO_KEY, interfaces.AccountInfo{
			ID:   interfaces.ADMIN_ID,
			Type: interfaces.ADMIN_TYPE,
		})
)

func MockNewIndexBaseAccess(appSetting *common.AppSetting,
	httpClient rest.HTTPClient) interfaces.IndexBaseAccess {

	iba := &indexBaseAccess{
		appSetting: appSetting,
		httpClient: httpClient,
	}
	return iba
}

func Test_IndexBaseAccess_GetIndexBasesByTypes(t *testing.T) {
	Convey("Test GetIndexBasesByTypes", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			IndexBaseUrl: "http://localhost:13012/api/mdl-index-base/v1",
		}
		httpClient := rmock.NewMockHTTPClient(mockCtrl)
		iba := MockNewIndexBaseAccess(appSetting, httpClient)

		baseResp := []interfaces.IndexBase{
			{
				SimpleIndexBase: interfaces.SimpleIndexBase{
					BaseType: "a",
				},
			},
			{
				SimpleIndexBase: interfaces.SimpleIndexBase{
					BaseType: "b",
				},
			},
		}
		okResp, _ := sonic.Marshal(baseResp)

		Convey("failed, caused by http error", func() {
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				http.StatusInternalServerError, nil, fmt.Errorf("method failed"))

			indexbase, err := iba.GetIndexBasesByTypes(testCtx, []string{"a", "b"})

			So(indexbase, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("failed, caused by status != 200", func() {
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				http.StatusInternalServerError, okResp, nil)

			indexbase, err := iba.GetIndexBasesByTypes(testCtx, []string{"a", "b"})

			So(indexbase, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("failed, caused by unmarshal error", func() {
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				http.StatusOK, okResp, nil)

			patch := ApplyFuncReturn(sonic.Unmarshal, errors.New("error"))
			defer patch.Reset()

			indexbase, err := iba.GetIndexBasesByTypes(testCtx, []string{"a", "b"})

			So(indexbase, ShouldBeNil)
			So(err.Error(), ShouldEqual, "error")
		})

		Convey("failed, caused by the length of Data is less than requested", func() {
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				http.StatusOK, okResp, nil)

			indexbase, err := iba.GetIndexBasesByTypes(testCtx, []string{"a", "b", "c"})

			So(indexbase, ShouldNotBeNil)
			So(err, ShouldBeNil)
		})

		Convey("success", func() {
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				http.StatusOK, okResp, nil)

			indexbase, err := iba.GetIndexBasesByTypes(testCtx, []string{"a", "b"})
			So(indexbase, ShouldNotBeEmpty)
			So(err, ShouldBeNil)
		})
	})
}

func Test_IndexBaseAccess_GetSimpleIndexBasesByTypes(t *testing.T) {
	Convey("Test GetSimpleIndexBasesByTypes", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			IndexBaseUrl: "http://localhost:13012/api/mdl-index-base/v1",
		}
		httpClient := rmock.NewMockHTTPClient(mockCtrl)
		iba := MockNewIndexBaseAccess(appSetting, httpClient)

		Convey("failed, caused by http error", func() {
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(http.StatusInternalServerError, nil, fmt.Errorf("method failed"))

			indeBase, err := iba.GetSimpleIndexBasesByTypes(testCtx, []string{"a"})
			So(indeBase, ShouldBeEmpty)
			So(err, ShouldResemble, fmt.Errorf("method failed"))
		})

		Convey("failed, caused by status != 200", func() {
			okResp, _ := sonic.Marshal(rest.BaseError{ErrorCode: "a", Description: "a", ErrorDetails: "a"})
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(http.StatusAccepted, okResp, nil)

			indeBase, err := iba.GetSimpleIndexBasesByTypes(testCtx, []string{"a"})
			So(indeBase, ShouldBeEmpty)
			So(err, ShouldResemble, fmt.Errorf("get index base [a] return error {\"error_code\":\"a\",\"description\":\"a\",\"solution\":\"\",\"error_link\":\"\",\"error_details\":\"a\"}"))
		})

		Convey("failed, caused by http result is null", func() {
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(http.StatusOK, nil, nil)

			indeBase, err := iba.GetSimpleIndexBasesByTypes(testCtx, []string{"a"})
			So(indeBase, ShouldBeEmpty)
			So(err, ShouldResemble, fmt.Errorf("get index base [a] return null"))
		})

		Convey("failed, caused by unmarshal error", func() {
			okResp, _ := sonic.Marshal([]interfaces.SimpleIndexBase{{BaseType: "base1", Name: "base1"}})
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(http.StatusOK, okResp, nil)

			patch := ApplyFuncReturn(sonic.Unmarshal, errors.New("error"))
			defer patch.Reset()

			indeBase, err := iba.GetSimpleIndexBasesByTypes(testCtx, []string{"a"})
			So(indeBase, ShouldBeEmpty)
			So(err.Error(), ShouldEqual, "error")
		})

		Convey("failed, caused by length not equal", func() {
			okResp, _ := sonic.Marshal([]interfaces.SimpleIndexBase{{BaseType: "a", Name: "a", Comment: "a"}})
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(http.StatusOK, okResp, nil)

			indeBase, err := iba.GetSimpleIndexBasesByTypes(testCtx, []string{"a", "b"})
			So(indeBase, ShouldBeNil)
			So(err.Error(), ShouldEqual, "have any IndexBase[[a b]] doesn't exist, expect number is 2, actual number is 1")
		})

		Convey("success", func() {
			okResp, _ := sonic.Marshal([]interfaces.SimpleIndexBase{{BaseType: "a", Name: "a", Comment: "a"}})
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(http.StatusOK, okResp, nil)

			indeBase, err := iba.GetSimpleIndexBasesByTypes(testCtx, []string{"a"})
			So(indeBase, ShouldResemble, []interfaces.SimpleIndexBase{{BaseType: "a", Name: "a", Comment: "a"}})
			So(err, ShouldBeNil)
		})
	})
}
