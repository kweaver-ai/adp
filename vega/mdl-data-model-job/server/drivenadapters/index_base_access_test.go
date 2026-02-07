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

	rmock "github.com/kweaver-ai/kweaver-go-lib/rest/mock"
	. "github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"data-model-job/common"
	"data-model-job/interfaces"
)

func MockNewIndexBaseAccess(appSetting *common.AppSetting,
	mockCtl *gomock.Controller) (interfaces.IndexBaseAccess, *rmock.MockHTTPClient) {

	httpClient := rmock.NewMockHTTPClient(mockCtl)

	iba := &indexBaseAccess{
		appSetting: appSetting,
		httpClient: httpClient,
	}

	return iba, httpClient
}

func Test_IndexBaseAccess_GetIndexBasesByTypes(t *testing.T) {
	Convey("Test GetIndexBasesByTypes", t, func() {
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		appSetting := &common.AppSetting{
			IndexBaseUrl: "http://localhost:13012/api/index-mgmt/v1",
		}
		iba, httpClient := MockNewIndexBaseAccess(appSetting, mockCtl)

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
		okResp, _ := json.Marshal(baseResp)

		Convey("failed, caused by http error", func() {
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).Return(http.StatusInternalServerError, nil, fmt.Errorf("method failed"))

			indexbase, err := iba.GetIndexBasesByTypes(testCtx, []string{"a", "b"})
			So(indexbase, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("failed, caused by status != 200", func() {
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).Return(http.StatusInternalServerError, okResp, nil)

			indexbase, err := iba.GetIndexBasesByTypes(testCtx, []string{"a", "b"})
			So(indexbase, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("failed, caused by unmarshal error", func() {
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).Return(http.StatusOK, okResp, nil)

			patch := ApplyFuncReturn(json.Unmarshal, errors.New("error"))
			defer patch.Reset()

			indexbase, err := iba.GetIndexBasesByTypes(testCtx, []string{"a", "b"})

			So(indexbase, ShouldBeNil)
			So(err.Error(), ShouldEqual, "error")
		})

		Convey("failed, caused by the length of Data is less than requested", func() {
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).Return(http.StatusOK, okResp, nil)

			indexbase, err := iba.GetIndexBasesByTypes(testCtx, []string{"a", "b", "c"})

			So(indexbase, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("success", func() {
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).Return(http.StatusOK, okResp, nil)

			indexbase, err := iba.GetIndexBasesByTypes(testCtx, []string{"a", "b"})
			So(indexbase, ShouldNotBeEmpty)
			So(err, ShouldBeNil)
		})
	})
}
