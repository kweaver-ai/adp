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

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	rmock "github.com/kweaver-ai/kweaver-go-lib/rest/mock"
	. "github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"data-model-job/interfaces"
)

var (
	task = interfaces.MetricTask{
		TaskID:   "1",
		TaskName: "task1",
		ModelID:  "1",
		Schedule: interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
		// Filters:            []interfaces.Filter{{Name: "labels.mode", Operation: "=", Value: "idle"}},
		TimeWindows:        []string{"5m", "1h"},
		IndexBase:          "base1",
		RetraceDuration:    "1d",
		ScheduleSyncStatus: 1,
		Comment:            "task1-aaa",
		UpdateTime:         1699336878575,
		PlanTime:           int64(1699336878575),
	}
)

func MockNewMetricModelAccess(mockCtl *gomock.Controller) (*metricModelAccess, *rmock.MockHTTPClient) {
	httpClient := rmock.NewMockHTTPClient(mockCtl)

	mma := &metricModelAccess{
		metricTaskUrl: "http://data-model-anyrobot:13020/api/data-model/v1/metric-tasks",
		httpClient:    httpClient,
	}
	return mma, httpClient
}

func Test_MetricModelAccess_GetTaskPlanTimeById(t *testing.T) {
	Convey("Test GetTaskPlanTimeById", t, func() {
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		mma, httpClient := MockNewMetricModelAccess(mockCtl)

		Convey("failed, caused by http error", func() {
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				AnyTimes().Return(http.StatusInternalServerError, nil, fmt.Errorf("method failed"))

			planTime, err := mma.GetTaskPlanTimeById(testCtx, "1")
			So(planTime, ShouldEqual, 0)
			So(err, ShouldResemble, fmt.Errorf("method failed"))
		})

		Convey("failed, caused by status != 200", func() {
			okResp, _ := json.Marshal(rest.BaseError{ErrorCode: "a", Description: "a", ErrorDetails: "a"})
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(http.StatusAccepted, okResp, nil)

			planTime, err := mma.GetTaskPlanTimeById(testCtx, "1")
			So(planTime, ShouldEqual, 0)
			So(err, ShouldResemble, fmt.Errorf("get metric task 1 return error {\"error_code\":\"a\",\"description\":\"a\",\"solution\":\"\",\"error_link\":\"\",\"error_details\":\"a\"}"))
		})

		Convey("failed, caused by unmarshal base error error", func() {
			okResp, _ := json.Marshal(task)
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(http.StatusAccepted, okResp, nil)

			patch := ApplyFuncReturn(json.Unmarshal, errors.New("error"))
			defer patch.Reset()

			planTime, err := mma.GetTaskPlanTimeById(testCtx, "1")
			So(planTime, ShouldEqual, 0)
			So(err.Error(), ShouldEqual, "error")
		})

		Convey("failed, caused by http result is null", func() {
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(http.StatusOK, nil, nil)

			planTime, err := mma.GetTaskPlanTimeById(testCtx, "1")
			So(planTime, ShouldEqual, 0)
			So(err, ShouldResemble, fmt.Errorf("get metric task 1 return null"))
		})

		Convey("failed, caused by unmarshal MetricTask error", func() {
			okResp, _ := json.Marshal(task)
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(http.StatusOK, okResp, nil)

			patch := ApplyFunc(json.Unmarshal,
				func(data []byte, v any) error {
					if _, ok := v.(*interfaces.MetricTask); ok {
						return errors.New("error")
					}
					return nil
				},
			)
			defer patch.Reset()

			planTime, err := mma.GetTaskPlanTimeById(testCtx, "1")
			So(planTime, ShouldEqual, 0)
			So(err.Error(), ShouldEqual, "error")
		})

		Convey("success", func() {
			okResp, _ := json.Marshal(task)
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(http.StatusOK, okResp, nil)

			planTime, err := mma.GetTaskPlanTimeById(testCtx, "1")
			So(planTime, ShouldResemble, task.PlanTime)
			So(err, ShouldBeNil)
		})
	})
}

func Test_MetricModelAccess_UpdateTaskAttributesById(t *testing.T) {
	Convey("Test UpdateTaskAttributesById", t, func() {
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		mma, httpClient := MockNewMetricModelAccess(mockCtl)

		Convey("failed, caused by http error", func() {
			httpClient.EXPECT().PutNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				AnyTimes().Return(http.StatusInternalServerError, nil, fmt.Errorf("method failed"))

			err := mma.UpdateTaskAttributesById(testCtx, "1", task)
			So(err, ShouldResemble, fmt.Errorf("method failed"))
		})

		Convey("failed, caused by status != 204", func() {
			okResp, _ := json.Marshal(rest.BaseError{ErrorCode: "a", Description: "a", ErrorDetails: "a"})
			httpClient.EXPECT().PutNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(http.StatusAccepted, okResp, nil)

			err := mma.UpdateTaskAttributesById(testCtx, "1", task)
			So(err, ShouldResemble, fmt.Errorf("put metric task 1 return error {\"error_code\":\"a\",\"description\":\"a\",\"solution\":\"\",\"error_link\":\"\",\"error_details\":\"a\"}"))
		})

		Convey("failed, caused by unmarshal error", func() {
			okResp, _ := json.Marshal(rest.BaseError{})
			httpClient.EXPECT().PutNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(http.StatusAccepted, okResp, nil)

			patch := ApplyFuncReturn(json.Unmarshal, errors.New("error"))
			defer patch.Reset()

			err := mma.UpdateTaskAttributesById(testCtx, "1", task)
			So(err.Error(), ShouldEqual, "error")
		})

		Convey("success", func() {
			okResp, _ := json.Marshal(task)
			httpClient.EXPECT().PutNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(http.StatusNoContent, okResp, nil)

			err := mma.UpdateTaskAttributesById(testCtx, "1", task)
			So(err, ShouldBeNil)
		})
	})
}
