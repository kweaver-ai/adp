// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package drivenadapters

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	rmock "github.com/kweaver-ai/kweaver-go-lib/rest/mock"
	. "github.com/smartystreets/goconvey/convey"

	"uniquery/interfaces"
)

var (
	testCtx = context.WithValue(context.WithValue(context.Background(),
		rest.XLangKey, rest.DefaultLanguage),
		interfaces.ACCOUNT_INFO_KEY, interfaces.AccountInfo{
			ID:   interfaces.ADMIN_ID,
			Type: interfaces.ADMIN_TYPE,
		})
)

func TestGetMetricModel(t *testing.T) {
	Convey("Test GetMetricModel", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockHttpClient := rmock.NewMockHTTPClient(mockCtrl)
		dmAccess := &metricModelAccess{httpClient: mockHttpClient}

		var expect []interfaces.MetricModel
		Convey("get request method failed", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, fmt.Errorf("method failed"))

			metricModel, exists, err := dmAccess.GetMetricModel(testCtx, "123")

			So(metricModel, ShouldResemble, expect)
			So(exists, ShouldBeFalse)
			So(err, ShouldResemble, fmt.Errorf("get request method failed: method failed"))
		})

		Convey("metric model not found", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusNotFound, nil, nil)

			metricModel, exists, err := dmAccess.GetMetricModel(testCtx, "123")

			So(metricModel, ShouldResemble, expect)
			So(exists, ShouldBeFalse)
			So(err, ShouldBeNil)
		})

		Convey("get metric model failed because unmarshal", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusInternalServerError, nil, nil)

			metricModel, exists, err := dmAccess.GetMetricModel(testCtx, "123")

			So(metricModel, ShouldResemble, expect)
			So(exists, ShouldBeFalse)
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

			metricModel, exists, err := dmAccess.GetMetricModel(testCtx, "123")

			So(metricModel, ShouldResemble, expect)
			So(exists, ShouldBeFalse)
			So(err.Error(), ShouldEqual, "get metric model failed: failed")
		})

		Convey("response nil ", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, nil)

			metricModel, exists, err := dmAccess.GetMetricModel(testCtx, "123")
			So(metricModel, ShouldResemble, expect)
			So(exists, ShouldBeFalse)
			So(err, ShouldBeNil)
		})

		Convey("response unmarshal failed ", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, []uint8{}, nil)

			metricModel, exists, err := dmAccess.GetMetricModel(testCtx, "123")

			So(metricModel, ShouldResemble, expect)
			So(exists, ShouldBeFalse)
			So(err, ShouldNotBeNil)
		})

		Convey("success ", func() {
			metricModelE := interfaces.MetricModel{
				ModelName:  "txy",
				MetricType: "atomic",
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "主机监控",
				},
				QueryType:    "promql",
				Formula:      "rate(node_network_receive_bytes_total[5m])",
				DateField:    "@timestamp",
				MeasureField: "value",
				UnitType:     "timeUnit",
				Unit:         "ms",
			}

			bytes, _ := sonic.Marshal([]interfaces.MetricModel{metricModelE})
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, bytes, nil)

			metricModel, exists, err := dmAccess.GetMetricModel(testCtx, "123")

			So(metricModel, ShouldResemble, []interfaces.MetricModel{metricModelE})
			So(exists, ShouldBeTrue)
			So(err, ShouldBeNil)
		})

		Convey("failed when model is for sql", func() {
			metricModelE := interfaces.MetricModel{
				ModelName:  "txy",
				MetricType: "atomic",
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "主机监控",
				},
				QueryType:     "sql",
				FormulaConfig: "{}",
				DateField:     "@timestamp",
				MeasureField:  "value",
				UnitType:      "timeUnit",
				Unit:          "ms",
			}

			bytes, _ := sonic.Marshal([]interfaces.MetricModel{metricModelE})
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, bytes, nil)

			_, exists, err := dmAccess.GetMetricModel(testCtx, "123")

			So(exists, ShouldBeFalse)
			So(err, ShouldNotBeNil)
		})

		Convey("success when model is for sql", func() {
			metricModelE := interfaces.MetricModel{
				ModelName:  "txy",
				MetricType: "atomic",
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "主机监控",
				},
				QueryType: "sql",
				FormulaConfig: interfaces.SQLConfig{
					AggrExprStr: "avg(f1)",
				},
				DateField:    "@timestamp",
				MeasureField: "value",
				UnitType:     "timeUnit",
				Unit:         "ms",
			}

			bytes, _ := sonic.Marshal([]interfaces.MetricModel{metricModelE})
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, bytes, nil)

			metricModel, exists, err := dmAccess.GetMetricModel(testCtx, "123")

			So(metricModel, ShouldResemble, []interfaces.MetricModel{metricModelE})
			So(exists, ShouldBeTrue)
			So(err, ShouldBeNil)
		})
	})
}

func TestGetMetricModelIDByName(t *testing.T) {
	Convey("Test GetMetricModel", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockHttpClient := rmock.NewMockHTTPClient(mockCtrl)
		mmAccess := &metricModelAccess{httpClient: mockHttpClient}

		Convey("get request method failed", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, fmt.Errorf("method failed"))

			modelID, exists, err := mmAccess.GetMetricModelIDByName(testCtx, "group1", "model1")

			So(modelID, ShouldEqual, "")
			So(exists, ShouldBeFalse)
			So(err, ShouldResemble, fmt.Errorf("get request method failed: method failed"))
		})

		Convey("metric model not found", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusNotFound, nil, nil)
			modelID, exists, err := mmAccess.GetMetricModelIDByName(testCtx, "group1", "model1")

			So(modelID, ShouldEqual, "")
			So(exists, ShouldBeFalse)
			So(err, ShouldBeNil)
		})

		Convey("get metric model failed because unmarshal", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusInternalServerError, nil, nil)

			modelID, exists, err := mmAccess.GetMetricModelIDByName(testCtx, "group1", "model1")

			So(modelID, ShouldEqual, "")
			So(exists, ShouldBeFalse)
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

			modelID, exists, err := mmAccess.GetMetricModelIDByName(testCtx, "group1", "model1")

			So(modelID, ShouldEqual, "")
			So(exists, ShouldBeFalse)
			So(err.Error(), ShouldEqual, "get metric model id failed: failed")
		})

		Convey("response nil ", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, nil)

			modelID, exists, err := mmAccess.GetMetricModelIDByName(testCtx, "group1", "model1")

			So(modelID, ShouldEqual, "")
			So(exists, ShouldBeFalse)
			So(err, ShouldBeNil)
		})

		Convey("response unmarshal failed ", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, []uint8{}, nil)
			modelID, exists, err := mmAccess.GetMetricModelIDByName(testCtx, "group1", "model1")

			So(modelID, ShouldEqual, "")
			So(exists, ShouldBeFalse)
			So(err, ShouldNotBeNil)
		})

		Convey("success ", func() {
			metricModelE := interfaces.MetricModel{
				ModelID:    "123",
				ModelName:  "model1",
				MetricType: "atomic",
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "主机监控",
				},
				QueryType:    "promql",
				Formula:      "rate(node_network_receive_bytes_total[5m])",
				DateField:    "@timestamp",
				MeasureField: "value",
				UnitType:     "timeUnit",
				Unit:         "ms",
				GroupID:      "123",
				GroupName:    "group1",
			}

			bytes, _ := sonic.Marshal(metricModelE)
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, bytes, nil)

			modelID, exists, err := mmAccess.GetMetricModelIDByName(testCtx, "group1", "model1")

			So(modelID, ShouldEqual, "123")
			So(exists, ShouldBeTrue)
			So(err, ShouldBeNil)
		})
	})
}

func TestGetMetricModels(t *testing.T) {
	Convey("Test GetMetricModels", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockHttpClient := rmock.NewMockHTTPClient(mockCtrl)
		mmAccess := &metricModelAccess{httpClient: mockHttpClient}

		Convey("get request method failed", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, fmt.Errorf("method failed"))

			metricModels, err := mmAccess.GetMetricModels(testCtx, []string{"123"})

			So(metricModels, ShouldBeEmpty)
			So(err, ShouldResemble, fmt.Errorf("get request method failed: method failed"))
		})

		Convey("response unmarshal failed", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, []uint8{}, nil)

			metricModels, err := mmAccess.GetMetricModels(testCtx, []string{"123"})

			So(metricModels, ShouldBeEmpty)
			So(err, ShouldNotBeNil)
		})

		Convey("success", func() {
			metricModelE := interfaces.MetricModel{
				ModelID:    "123",
				ModelName:  "model1",
				MetricType: "atomic",
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "主机监控",
				},
				QueryType:    "promql",
				Formula:      "rate(node_network_receive_bytes_total[5m])",
				DateField:    "@timestamp",
				MeasureField: "value",
				UnitType:     "timeUnit",
				Unit:         "ms",
				GroupID:      "123",
				GroupName:    "group1",
			}

			bytes, _ := sonic.Marshal([]interfaces.MetricModel{metricModelE})
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, bytes, nil)

			metricModels, err := mmAccess.GetMetricModels(testCtx, []string{"123"})

			So(len(metricModels), ShouldEqual, 1)
			So(metricModels[0], ShouldResemble, metricModelE)
			So(err, ShouldBeNil)
		})
	})
}
