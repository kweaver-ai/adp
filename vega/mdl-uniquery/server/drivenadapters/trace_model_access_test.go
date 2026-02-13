// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package drivenadapters

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/bytedance/sonic"
	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	rmock "github.com/kweaver-ai/kweaver-go-lib/rest/mock"
	. "github.com/smartystreets/goconvey/convey"

	"uniquery/common"
	"uniquery/interfaces"
)

func TestGetTraceModelByID(t *testing.T) {
	Convey("Test GetTraceModelByID", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		testCtx = context.WithValue(context.WithValue(context.Background(),
			rest.XLangKey, rest.DefaultLanguage),
			interfaces.ACCOUNT_INFO_KEY, interfaces.AccountInfo{
				ID:   interfaces.ADMIN_ID,
				Type: interfaces.ADMIN_TYPE,
			})

		mockHttpClient := rmock.NewMockHTTPClient(mockCtrl)
		tmAccess := &traceModelAccess{
			appSetting: &common.AppSetting{},
			httpClient: mockHttpClient,
		}

		Convey("Get failed, caused by the error from method 'GetNoUnmarshal'", func() {
			expectedErr := fmt.Errorf("some errors")

			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, expectedErr)

			_, _, err := tmAccess.GetTraceModelByID(testCtx, "1")
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by the trace model was not found", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusNotFound, nil, nil)

			_, isExist, err := tmAccess.GetTraceModelByID(testCtx, "1")
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
			expectedErr := fmt.Errorf("failed to get trace model by http client: %v", string(expectedRespData))

			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusInternalServerError, expectedRespData, nil)

			_, _, err := tmAccess.GetTraceModelByID(testCtx, "1")
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by the error from func 'sonic.Unmarshal'", func() {
			expectedErr := errors.New("some errors")

			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, nil)

			patch := ApplyFuncReturn(sonic.Unmarshal, expectedErr)
			defer patch.Reset()

			_, _, err := tmAccess.GetTraceModelByID(testCtx, "1")
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by the error from method 'commonProcessTraceModel'", func() {
			expectedErr := errors.New("some errors")

			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, nil)

			patch1 := ApplyFuncReturn(sonic.Unmarshal, nil)
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(reflect.TypeOf(&traceModelAccess{}), "commonProcessTraceModel",
				func(t *traceModelAccess, testCtx context.Context, model interfaces.TraceModel) (interfaces.TraceModel, error) {
					return interfaces.TraceModel{}, expectedErr
				},
			)
			defer patch2.Reset()

			_, _, err := tmAccess.GetTraceModelByID(testCtx, "1")
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get succeed", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, nil)

			patch1 := ApplyFuncReturn(sonic.Unmarshal, nil)
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(reflect.TypeOf(&traceModelAccess{}), "commonProcessTraceModel",
				func(t *traceModelAccess, testCtx context.Context, model interfaces.TraceModel) (interfaces.TraceModel, error) {
					return interfaces.TraceModel{}, nil
				},
			)
			defer patch2.Reset()

			_, isExist, err := tmAccess.GetTraceModelByID(testCtx, "1")
			So(err, ShouldResemble, nil)
			So(isExist, ShouldBeTrue)
		})
	})
}

func TestSimulateCreateTraceModel(t *testing.T) {
	Convey("Test SimulateCreateTraceModel", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		testCtx = context.WithValue(context.WithValue(context.Background(),
			rest.XLangKey, rest.DefaultLanguage),
			interfaces.ACCOUNT_INFO_KEY, interfaces.AccountInfo{
				ID:   interfaces.ADMIN_ID,
				Type: interfaces.ADMIN_TYPE,
			})

		mockHttpClient := rmock.NewMockHTTPClient(mockCtrl)
		tmAccess := &traceModelAccess{
			appSetting: &common.AppSetting{},
			httpClient: mockHttpClient,
		}

		Convey("Simulate create failed, caused by the error from method 'PostNoUnmarshal'", func() {
			expectedErr := fmt.Errorf("some errors")
			model := interfaces.TraceModel{}

			mockHttpClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, expectedErr)

			_, err := tmAccess.SimulateCreateTraceModel(testCtx, model)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Simulate create failed, caused by the respCode from method 'PostNoUnmarshal'", func() {
			expectedRespData, _ := json.Marshal(rest.BaseError{ErrorCode: "as",
				Description:  "500",
				ErrorLink:    "",
				Solution:     "",
				ErrorDetails: "failed",
			})
			expectedErr := fmt.Errorf("Failed to simulate create trace model by http client: %v", string(expectedRespData))
			model := interfaces.TraceModel{}

			mockHttpClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusInternalServerError, expectedRespData, nil)

			_, err := tmAccess.SimulateCreateTraceModel(testCtx, model)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Simulate create failed, caused by the error from func 'sonic.Unmarshal'", func() {
			expectedErr := errors.New("some errors")
			model := interfaces.TraceModel{}

			mockHttpClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, nil)

			patch := ApplyFuncReturn(sonic.Unmarshal, expectedErr)
			defer patch.Reset()

			_, err := tmAccess.SimulateCreateTraceModel(testCtx, model)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Simulate create failed, caused by the error from method 'commonProcessTraceModel'", func() {
			expectedErr := errors.New("some errors")
			model := interfaces.TraceModel{}

			mockHttpClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, nil)

			patch1 := ApplyFuncReturn(sonic.Unmarshal, nil)
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(reflect.TypeOf(&traceModelAccess{}), "commonProcessTraceModel",
				func(t *traceModelAccess, testCtx context.Context, model interfaces.TraceModel) (interfaces.TraceModel, error) {
					return interfaces.TraceModel{}, expectedErr
				},
			)
			defer patch2.Reset()

			_, err := tmAccess.SimulateCreateTraceModel(testCtx, model)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Simulate create succeed", func() {
			model := interfaces.TraceModel{}

			mockHttpClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, nil)

			patch1 := ApplyFuncReturn(sonic.Unmarshal, nil)
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(reflect.TypeOf(&traceModelAccess{}), "commonProcessTraceModel",
				func(t *traceModelAccess, testCtx context.Context, model interfaces.TraceModel) (interfaces.TraceModel, error) {
					return interfaces.TraceModel{}, nil
				},
			)
			defer patch2.Reset()

			_, err := tmAccess.SimulateCreateTraceModel(testCtx, model)
			So(err, ShouldResemble, nil)
		})
	})
}

func TestSimulateUpdateTraceModel(t *testing.T) {
	Convey("Test SimulateUpdateTraceModel", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		testCtx = context.WithValue(context.WithValue(context.Background(),
			rest.XLangKey, rest.DefaultLanguage),
			interfaces.ACCOUNT_INFO_KEY, interfaces.AccountInfo{
				ID:   interfaces.ADMIN_ID,
				Type: interfaces.ADMIN_TYPE,
			})

		mockHttpClient := rmock.NewMockHTTPClient(mockCtrl)
		tmAccess := &traceModelAccess{
			appSetting: &common.AppSetting{},
			httpClient: mockHttpClient,
		}

		Convey("Simulate update failed, caused by the error from method 'PutNoUnmarshal'", func() {
			expectedErr := fmt.Errorf("some errors")
			model := interfaces.TraceModel{}

			mockHttpClient.EXPECT().PutNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, expectedErr)

			_, err := tmAccess.SimulateUpdateTraceModel(testCtx, model.ID, model)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Simulate update failed, caused by the respCode from method 'PutNoUnmarshal'", func() {
			expectedRespData, _ := json.Marshal(rest.BaseError{ErrorCode: "as",
				Description:  "500",
				ErrorLink:    "",
				Solution:     "",
				ErrorDetails: "failed",
			})
			expectedErr := fmt.Errorf("Failed to simulate update trace model by http client: %v", string(expectedRespData))
			model := interfaces.TraceModel{}

			mockHttpClient.EXPECT().PutNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusInternalServerError, expectedRespData, nil)

			_, err := tmAccess.SimulateUpdateTraceModel(testCtx, model.ID, model)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Simulate update failed, caused by the error from func 'sonic.Unmarshal'", func() {
			expectedErr := errors.New("some errors")
			model := interfaces.TraceModel{}

			mockHttpClient.EXPECT().PutNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, nil)

			patch := ApplyFuncReturn(sonic.Unmarshal, expectedErr)
			defer patch.Reset()

			_, err := tmAccess.SimulateUpdateTraceModel(testCtx, model.ID, model)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Simulate update failed, caused by the error from method 'commonProcessTraceModel'", func() {
			expectedErr := errors.New("some errors")
			model := interfaces.TraceModel{}

			mockHttpClient.EXPECT().PutNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, nil)

			patch1 := ApplyFuncReturn(sonic.Unmarshal, nil)
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(reflect.TypeOf(&traceModelAccess{}), "commonProcessTraceModel",
				func(t *traceModelAccess, testCtx context.Context, model interfaces.TraceModel) (interfaces.TraceModel, error) {
					return interfaces.TraceModel{}, expectedErr
				},
			)
			defer patch2.Reset()

			_, err := tmAccess.SimulateUpdateTraceModel(testCtx, model.ID, model)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Simulate update succeed", func() {
			model := interfaces.TraceModel{}

			mockHttpClient.EXPECT().PutNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, nil)

			patch1 := ApplyFuncReturn(sonic.Unmarshal, nil)
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(reflect.TypeOf(&traceModelAccess{}), "commonProcessTraceModel",
				func(t *traceModelAccess, testCtx context.Context, model interfaces.TraceModel) (interfaces.TraceModel, error) {
					return interfaces.TraceModel{}, nil
				},
			)
			defer patch2.Reset()

			_, err := tmAccess.SimulateUpdateTraceModel(testCtx, model.ID, model)
			So(err, ShouldResemble, nil)
		})
	})
}

func TestCommonProcessTraceModel(t *testing.T) {
	Convey("Test commonProcessTraceModel", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		testCtx := context.Background()
		mockHttpClient := rmock.NewMockHTTPClient(mockCtrl)
		tmAccess := &traceModelAccess{
			appSetting: &common.AppSetting{},
			httpClient: mockHttpClient,
		}

		Convey("Get failed, caused by the error from func 'sonic.Marshal'", func() {
			expectedErr := errors.New("some errors")

			patch := ApplyFuncReturn(sonic.Marshal, []byte(nil), expectedErr)
			defer patch.Reset()

			_, err := tmAccess.commonProcessTraceModel(testCtx, interfaces.TraceModel{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by the error from func 'sonic.Unmarshal' when span_source_type is data_view", func() {
			expectedErr := errors.New("some errors")
			model := interfaces.TraceModel{
				SpanSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
			}

			patch := ApplyFuncReturn(sonic.Unmarshal, expectedErr)
			defer patch.Reset()

			_, err := tmAccess.commonProcessTraceModel(testCtx, model)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by the error from func 'sonic.Marshal' when parsing related log config", func() {
			expectedErr := errors.New("some errors")
			model := interfaces.TraceModel{
				SpanSourceType:    interfaces.SOURCE_TYPE_DATA_CONNECTION,
				SpanConfig:        "",
				EnabledRelatedLog: interfaces.RELATED_LOG_OPEN,
			}

			patch1 := ApplyFuncSeq(sonic.Marshal, []OutputCell{
				{Values: Params{[]byte(nil), nil}, Times: 1},
				{Values: Params{nil, expectedErr}, Times: 1},
			})
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(sonic.Unmarshal, nil)
			defer patch2.Reset()

			_, err := tmAccess.commonProcessTraceModel(testCtx, model)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by the error from func 'sonic.Unmarshal' when parsing related log config", func() {
			expectedErr := errors.New("some errors")
			model := interfaces.TraceModel{
				SpanSourceType:       interfaces.SOURCE_TYPE_DATA_CONNECTION,
				SpanConfig:           "",
				EnabledRelatedLog:    interfaces.RELATED_LOG_OPEN,
				RelatedLogSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
			}

			patch1 := ApplyFuncReturn(sonic.Marshal, []byte(nil), nil)
			defer patch1.Reset()

			patch2 := ApplyFuncSeq(sonic.Unmarshal, []OutputCell{
				{Values: Params{nil}, Times: 1},
				{Values: Params{expectedErr}, Times: 1},
			})
			defer patch2.Reset()

			_, err := tmAccess.commonProcessTraceModel(testCtx, model)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get succeed", func() {
			model := interfaces.TraceModel{
				SpanSourceType:       interfaces.SOURCE_TYPE_DATA_CONNECTION,
				SpanConfig:           "",
				EnabledRelatedLog:    interfaces.RELATED_LOG_OPEN,
				RelatedLogSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
			}

			patch1 := ApplyFuncReturn(sonic.Marshal, []byte(nil), nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(sonic.Unmarshal, nil)
			defer patch2.Reset()

			_, err := tmAccess.commonProcessTraceModel(testCtx, model)
			So(err, ShouldBeNil)
		})
	})
}
