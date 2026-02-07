// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package drivenadapters

import (
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
	uerrors "uniquery/errors"
	"uniquery/interfaces"
)

func MockNewIndexBaseAccess(mockCtrl *gomock.Controller) (interfaces.IndexBaseAccess, *rmock.MockHTTPClient) {
	mockHttpClient := rmock.NewMockHTTPClient(mockCtrl)

	ibaMock := &indexBaseAccess{
		appSetting: &common.AppSetting{
			IndexBaseUrl: "http://localhost:13012/api/mdl-index-base/v1",
		},
		httpClient: mockHttpClient,
	}

	return ibaMock, mockHttpClient
}

func TestGetIndexBasesByTypes(t *testing.T) {
	Convey("Test GetIndexBasesByTypes", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		ibaMock, mockHttpClient := MockNewIndexBaseAccess(mockCtrl)

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
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				http.StatusInternalServerError, nil, fmt.Errorf("method failed"))

			indexbase, err := ibaMock.GetIndexBasesByTypes(testCtx, []string{"a", "b"})

			So(indexbase, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("failed, caused by status != 200", func() {
			expectedErr := rest.BaseError{
				ErrorCode: uerrors.Uniquery_DataView_InternalError_GetIndexBaseByTypeFailed,
			}
			errResp, _ := sonic.Marshal(expectedErr)
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				http.StatusInternalServerError, errResp, nil)

			indexbase, err := ibaMock.GetIndexBasesByTypes(testCtx, []string{"a", "b"})

			So(indexbase, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("failed, caused by unmarshal baseError failed", func() {
			expectedErr := rest.BaseError{
				ErrorCode: uerrors.Uniquery_DataView_InternalError_GetIndexBaseByTypeFailed,
			}

			errResp, _ := sonic.Marshal(expectedErr)
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				http.StatusInternalServerError, errResp, nil)

			patch := ApplyFunc(sonic.Unmarshal,
				func(data []byte, v any) error {
					return errors.New("error")
				},
			)
			defer patch.Reset()

			indexbase, err := ibaMock.GetIndexBasesByTypes(testCtx, []string{"a", "b"})

			So(indexbase, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("failed, caused by unmarshal error", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				http.StatusOK, okResp, nil)

			patch := ApplyFunc(sonic.Unmarshal,
				func(data []byte, v any) error {
					return errors.New("error")
				},
			)
			defer patch.Reset()

			indexbase, err := ibaMock.GetIndexBasesByTypes(testCtx, []string{"a", "b"})

			So(indexbase, ShouldBeNil)
			So(err.Error(), ShouldEqual, "error")
		})

		Convey("failed, caused by the length of Data is less than requested", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				http.StatusOK, okResp, nil)

			indexbase, err := ibaMock.GetIndexBasesByTypes(testCtx, []string{"a", "b", "c"})

			So(indexbase, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("success", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				http.StatusOK, okResp, nil)

			indexbase, err := ibaMock.GetIndexBasesByTypes(testCtx, []string{"a", "b"})
			So(indexbase, ShouldNotBeEmpty)
			So(err, ShouldBeNil)
		})
	})
}

func TestGetIndices(t *testing.T) {
	Convey("Test GetIndices", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		ibaMock, mockHttpClient := MockNewIndexBaseAccess(mockCtrl)

		indicesResp := map[string]map[string]interfaces.Indice{
			"indices": {
				"ar_my_library2-2023.07.10-000000": {
					IndexName: "ar_my_library2-2023.07.10-000000",
					StartTime: 123,
					EndTime:   456,
					ShardNum:  3,
				},
				"ar_test_rotation_rw10-2023.08.01-000006": {
					IndexName: "ar_test_rotation_rw10-2023.08.01-000006",
					StartTime: 123,
					EndTime:   456,
					ShardNum:  3,
				},
			},
		}
		okResp, _ := sonic.Marshal(indicesResp)

		Convey("failed, caused by http error", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				http.StatusInternalServerError, nil, fmt.Errorf("method failed"))

			indices, respCode, err := ibaMock.GetIndices(testCtx, []string{"a", "b"}, 1676208000000, 1676380800000)

			So(indices, ShouldBeNil)
			So(respCode, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
		})

		Convey("failed, caused by status != 200", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				http.StatusInternalServerError, okResp, nil)

			indices, respCode, err := ibaMock.GetIndices(testCtx, []string{"a", "b"}, 1676208000000, 1676380800000)

			So(indices, ShouldBeNil)
			So(respCode, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
		})

		Convey("failed, caused by unmarshal baseError failed", func() {
			expectedErr := rest.BaseError{
				ErrorCode: uerrors.Uniquery_DataView_InternalError_GetIndexBaseByTypeFailed,
			}
			errResp, _ := sonic.Marshal(expectedErr)

			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				http.StatusInternalServerError, errResp, nil)

			patch := ApplyFunc(sonic.Unmarshal,
				func(data []byte, v any) error {
					return errors.New("error")
				},
			)
			defer patch.Reset()

			indices, respCode, err := ibaMock.GetIndices(testCtx, []string{"a", "b"}, 1676208000000, 1676380800000)

			So(indices, ShouldBeNil)
			So(respCode, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
		})

		Convey("failed, caused by unmarshal error", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				http.StatusOK, okResp, nil)

			patch := ApplyFunc(sonic.Unmarshal,
				func(data []byte, v any) error {
					return errors.New("error")
				},
			)
			defer patch.Reset()

			indices, respCode, err := ibaMock.GetIndices(testCtx, []string{"a", "b"}, 1676208000000, 1676380800000)

			So(indices, ShouldBeNil)
			So(respCode, ShouldEqual, http.StatusOK)
			So(err.Error(), ShouldEqual, "error")
		})

		Convey("success", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				http.StatusOK, okResp, nil)

			indices, respCode, err := ibaMock.GetIndices(testCtx, []string{"a", "b"}, 1676208000000, 1676380800000)
			So(indices, ShouldNotBeEmpty)
			So(respCode, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})
	})
}
