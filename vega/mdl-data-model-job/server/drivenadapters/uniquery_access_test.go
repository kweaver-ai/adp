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
	"time"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	rmock "github.com/kweaver-ai/kweaver-go-lib/rest/mock"
	. "github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"data-model-job/interfaces"
)

var (
	emptyData = interfaces.UniResponse{}

	expectData = interfaces.UniResponse{
		Datas: []interfaces.Data{
			{
				Labels: map[string]string{
					"cpu": "1",
				},
				Times: []interface{}{float64(1695003480000), float64(1695003840000), float64(1695004200000),
					float64(1695004560000), float64(1695004920000), float64(1695005280000)},
				Values: []interface{}{1.1, nil, nil, 1.1, nil, nil},
			},
		},
		Step: "5m",
	}

	eventResponse = interfaces.EventModelResponse{
		Entries: []interfaces.EventModelData{

			{
				Id:           "1",
				Title:        "xjx_紧急",
				EventModelId: "1",
				EventType:    "atomic",
				Level:        1,
				Message:      "指标模型11的值超过90",
				// TriggerTime:    "2023-02-22 15:29:11",
				TriggerData:    interfaces.Records{},
				Tags:           []string{},
				EventModelName: "111",
				// CreateTime:     "2023-02-22 15:29:11",
				DataSource:     []string{"11111111"},
				DataSourceName: []string{},
				DataSourceType: "metric_model",
				DefaultTimeWindow: interfaces.TimeInterval{
					Interval: 5,
					Unit:     "m",
				},
			},
		},
		TotalCount: 1,
	}
	emptyEventResponse = interfaces.EventModelResponse{}
)

func MockNewUniQueryAccess(mockCtl *gomock.Controller) (interfaces.UniqueryAccess, *rmock.MockHTTPClient) {
	httpClient := rmock.NewMockHTTPClient(mockCtl)

	ua := &uniqueryAccess{
		uniqueryUrl: "http://uniquery-anyrobot:13011/api/uniquery/v1/metric-model",
		httpClient:  httpClient,
	}
	return ua, httpClient
}

func Test_UniQueryAccess_GetMetricModelData(t *testing.T) {
	Convey("Test GetMetricModelData", t, func() {
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		ua, httpClient := MockNewUniQueryAccess(mockCtl)

		query := interfaces.MetricModelQuery{
			IsInstantQuery: true,
			Time:           1,
			LookBackDelta:  "1m",
		}

		Convey("failed, caused by http error", func() {
			httpClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				AnyTimes().Return(http.StatusInternalServerError, nil, fmt.Errorf("method failed"))

			data, err := ua.GetMetricModelData(testCtx, "1", query)
			So(data, ShouldResemble, emptyData)
			So(err, ShouldResemble, fmt.Errorf("get request method failed: method failed"))
		})

		Convey("failed, caused by status != 200", func() {
			okResp, _ := json.Marshal(rest.BaseError{ErrorCode: "a", Description: "a", ErrorDetails: "a"})
			httpClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(http.StatusAccepted, okResp, nil)

			data, err := ua.GetMetricModelData(testCtx, "1", query)
			So(data, ShouldResemble, emptyData)
			So(err, ShouldResemble, fmt.Errorf(`get metric data 1 return error {"error_code":"a","description":"a","solution":"","error_link":"","error_details":"a"}`))
		})

		Convey("failed, caused by http result is null", func() {
			httpClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(http.StatusOK, nil, nil)

			data, err := ua.GetMetricModelData(testCtx, "1", query)
			So(data, ShouldResemble, emptyData)
			So(err, ShouldResemble, fmt.Errorf("get metric data 1 return null"))
		})

		Convey("failed, caused by unmarshal base error error", func() {
			okResp, _ := json.Marshal(rest.BaseError{ErrorCode: "a", Description: "a", ErrorDetails: "a"})
			httpClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(http.StatusAccepted, okResp, nil)

			patch := ApplyFuncReturn(json.Unmarshal, errors.New("error"))
			defer patch.Reset()

			data, err := ua.GetMetricModelData(testCtx, "1", query)
			So(data, ShouldResemble, emptyData)
			So(err, ShouldResemble, fmt.Errorf("error"))
		})

		Convey("failed, caused by Unmarshal metric data error", func() {
			okResp, _ := json.Marshal(expectData)
			httpClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(http.StatusOK, okResp, nil)

			expectedErr := errors.New("some error")
			patch := ApplyFunc(json.Unmarshal,
				func(data []byte, v any) error {
					if _, ok := v.(*interfaces.UniResponse); ok {
						return expectedErr
					}
					return nil
				},
			)
			defer patch.Reset()

			data, err := ua.GetMetricModelData(testCtx, "1", query)
			So(data, ShouldResemble, emptyData)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("success", func() {
			okResp, _ := json.Marshal(expectData)
			httpClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(http.StatusOK, okResp, nil)

			data, err := ua.GetMetricModelData(testCtx, "1", query)
			So(data, ShouldResemble, expectData)
			So(err, ShouldBeNil)
		})
	})
}

func Test_UniQueryAccess_GetEventModelData(t *testing.T) {
	Convey("Test GetEventModelData", t, func() {
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		ua, httpClient := MockNewUniQueryAccess(mockCtl)

		query := interfaces.EventModelQueryRequest{
			Querys: []interfaces.EventQuery{
				{
					End:       time.Now().Unix() * 1000, //当前时间戳(ms)
					QueryType: "instant_query",
					Id:        eventTask.ModelID,
				},
			},
		}

		Convey("failed, caused by http error", func() {
			httpClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				AnyTimes().Return(http.StatusInternalServerError, nil, fmt.Errorf("method failed"))

			data, err := ua.GetEventModelData(testCtx, query)
			So(data, ShouldResemble, emptyEventResponse)
			So(err, ShouldNotBeNil)
		})

		Convey("failed, caused by status != 200", func() {
			okResp, _ := json.Marshal(rest.BaseError{ErrorCode: "a", Description: "a", ErrorDetails: "a"})
			patches := ApplyMethodReturn(httpClient, "PostNoUnmarshal", http.StatusAccepted, okResp, nil)
			defer patches.Reset()

			data, err := ua.GetEventModelData(testCtx, query)
			So(data, ShouldResemble, emptyEventResponse)
			So(err, ShouldNotBeNil)
		})

		Convey("failed, caused by unmarshal error", func() {
			okResp, _ := json.Marshal(rest.BaseError{ErrorCode: "a", Description: "a", ErrorDetails: "a"})
			patches := ApplyMethodReturn(httpClient, "PostNoUnmarshal", http.StatusAccepted, okResp, nil)
			defer patches.Reset()

			patch := ApplyFuncReturn(json.Unmarshal, errors.New("error"))
			defer patch.Reset()

			data, err := ua.GetEventModelData(testCtx, query)
			So(data, ShouldResemble, emptyEventResponse)
			So(err, ShouldNotBeNil)
		})

		Convey("failed, caused by http result is null", func() {
			patches := ApplyMethodReturn(httpClient, "PostNoUnmarshal", http.StatusOK, nil, nil)
			defer patches.Reset()

			data, err := ua.GetEventModelData(testCtx, query)
			So(data, ShouldResemble, emptyEventResponse)
			So(err, ShouldNotBeNil)

		})

		Convey("failed, caused by Unmarshal event data error", func() {
			okResp, _ := json.Marshal(eventResponse)
			patches := ApplyMethodReturn(httpClient, "PostNoUnmarshal", http.StatusOK, okResp, nil)
			defer patches.Reset()

			expectedErr := errors.New("some error")
			patch := ApplyFunc(json.Unmarshal,
				func(data []byte, v any) error {
					if _, ok := v.(*interfaces.EventModelResponse); ok {
						return expectedErr
					}
					return nil
				},
			)
			defer patch.Reset()

			data, err := ua.GetEventModelData(testCtx, query)
			So(data, ShouldResemble, emptyEventResponse)
			So(err, ShouldResemble, expectedErr)
		})

	})
}

func Test_UniQueryAccess_GetObjectiveModelData(t *testing.T) {
	Convey("TestGetObjectiveModelData", t, func() {
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		ua, httpClient := MockNewUniQueryAccess(mockCtl)

		query := interfaces.MetricModelQuery{
			IsInstantQuery: true,
			Time:           123456,
			LookBackDelta:  "5m",
		}
		modelId := "test-model"

		emptyObjectiveResponse := interfaces.ObjectiveModelUniResponse{}

		objectiveResponse := interfaces.ObjectiveModelUniResponse{
			Datas: []interfaces.SLOObjectiveData{
				{
					Labels: map[string]string{
						"label1": "value1",
					},
					Times:            []any{123456},
					SLI:              []any{1.23},
					Objective:        []float64{1.23},
					Good:             []any{1.23},
					Total:            []any{1.23},
					AchiveRate:       []any{1.23},
					TotalErrorBudget: []any{1.23},
					LeftErrorBudget:  []any{1.23},
					BurnRate:         []any{1.23},
					Period:           []int64{123456},
					Status:           []string{"status1"},
					StatusCode:       []int{1},
				},
			},
		}

		Convey("success", func() {
			okResp, _ := json.Marshal(objectiveResponse)
			patches := ApplyMethodReturn(httpClient, "PostNoUnmarshal", http.StatusOK, okResp, nil)
			defer patches.Reset()

			data, err := ua.GetObjectiveModelData(testCtx, modelId, query)
			So(err, ShouldBeNil)
			So(len(data.Datas.([]interface{})), ShouldEqual, 1)
		})

		Convey("failed, caused by http client error", func() {
			expectedErr := errors.New("http client error")
			patches := ApplyMethodReturn(httpClient, "PostNoUnmarshal", http.StatusOK, nil, expectedErr)
			defer patches.Reset()

			data, err := ua.GetObjectiveModelData(testCtx, modelId, query)
			So(data, ShouldResemble, emptyObjectiveResponse)
			So(err, ShouldNotBeNil)
		})

		Convey("failed, caused by unmarshal error", func() {
			okResp, _ := json.Marshal(rest.BaseError{ErrorCode: "a", Description: "a", ErrorDetails: "a"})
			patches := ApplyMethodReturn(httpClient, "PostNoUnmarshal", http.StatusAccepted, okResp, nil)
			defer patches.Reset()

			patch := ApplyFuncReturn(json.Unmarshal, errors.New("error"))
			defer patch.Reset()

			data, err := ua.GetObjectiveModelData(testCtx, modelId, query)
			So(data, ShouldResemble, emptyObjectiveResponse)
			So(err, ShouldNotBeNil)
		})

		Convey("failed, caused by http result is null", func() {
			patches := ApplyMethodReturn(httpClient, "PostNoUnmarshal", http.StatusOK, nil, nil)
			defer patches.Reset()

			data, err := ua.GetObjectiveModelData(testCtx, modelId, query)
			So(data, ShouldResemble, emptyObjectiveResponse)
			So(err, ShouldNotBeNil)
		})

		Convey("failed, caused by Unmarshal objective data error", func() {
			okResp, _ := json.Marshal(objectiveResponse)
			patches := ApplyMethodReturn(httpClient, "PostNoUnmarshal", http.StatusOK, okResp, nil)
			defer patches.Reset()

			expectedErr := errors.New("some error")
			patch := ApplyFunc(json.Unmarshal,
				func(data []byte, v any) error {
					if _, ok := v.(*interfaces.ObjectiveModelUniResponse); ok {
						return expectedErr
					}
					return nil
				},
			)
			defer patch.Reset()

			data, err := ua.GetObjectiveModelData(testCtx, modelId, query)
			So(data, ShouldResemble, emptyObjectiveResponse)
			So(err, ShouldResemble, expectedErr)
		})
	})
}
