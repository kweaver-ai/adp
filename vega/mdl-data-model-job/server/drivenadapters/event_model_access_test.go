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
	eventModel = interfaces.EventModel{
		EventModelID:        "1",
		EventModelName:      "测试中的名称",
		EventModelType:      "atomic",
		UpdateTime:          1699336878575,
		EventModelTags:      []string{"xx1", "xx2"},
		DataSourceType:      "metric_model",
		DataSource:          []string{"1"},
		DataSourceName:      []string{},
		DataSourceGroupName: []string{},
		DetectRule:          interfaces.DetectRule{DetectRuleId: "0", Priority: 0, Type: "", Formula: []interfaces.FormulaItem{}},
		AggregateRule:       interfaces.AggregateRule{GroupFields: []string{}},
		DefaultTimeWindow:   interfaces.TimeInterval{Interval: 5, Unit: "m"},
		EventModelComment:   "comment",
		IsActive:            1,
		IsCustom:            1,
		Task:                eventTask,
	}
	eventTask = interfaces.EventTask{
		TaskID:   "1",
		ModelID:  "1",
		Schedule: interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
		StorageConfig: interfaces.StorageConfig{
			IndexBase:    "event_eventTask",
			DataViewName: "dip-event_eventTask-data-view",
		},
		DispatchConfig: interfaces.DispatchConfig{
			TimeOut:        3600,
			FailRetryCount: 3,
			RouteStrategy:  "FIRST",
			BlockStrategy:  "SERIAL_EXECUTION",
		},
		ExecuteParameter: map[string]any{
			"key": 1,
		},
		ScheduleSyncStatus: 1,
		StatusUpdateTime:   1699336878575,
		TaskStatus:         4,
		ErrorDetails:       "",
		UpdateTime:         1699336878575,
	}
)

func MockNewEventModelAccess(mockCtl *gomock.Controller) (*eventModelAccess, *rmock.MockHTTPClient) {
	httpClient := rmock.NewMockHTTPClient(mockCtl)

	ema := &eventModelAccess{
		eventTaskUrl: "http://data-model-anyrobot:13020/api/data-model/v1/event-models",
		httpClient:   httpClient,
	}
	return ema, httpClient
}

func Test_EventModelAccess_UpdateEventTaskAttributesById(t *testing.T) {
	Convey("Test UpdateEventTaskAttributesById", t, func() {
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		ema, httpClient := MockNewEventModelAccess(mockCtl)

		Convey("failed, caused by http error", func() {
			httpClient.EXPECT().PutNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				AnyTimes().Return(http.StatusInternalServerError, nil, fmt.Errorf("method failed"))

			err := ema.UpdateEventTaskAttributesById(testCtx, eventTask)
			So(err, ShouldResemble, fmt.Errorf("method failed"))
		})

		Convey("failed, caused by status != 204", func() {
			okResp, _ := json.Marshal(rest.BaseError{ErrorCode: "a", Description: "a", ErrorDetails: "a"})
			httpClient.EXPECT().PutNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(http.StatusAccepted, okResp, nil)

			err := ema.UpdateEventTaskAttributesById(testCtx, eventTask)
			So(err, ShouldResemble, fmt.Errorf("put event task 1 return error {\"error_code\":\"a\",\"description\":\"a\",\"solution\":\"\",\"error_link\":\"\",\"error_details\":\"a\"}"))
		})

		Convey("failed, caused by unmarshal error", func() {
			okResp, _ := json.Marshal(rest.BaseError{})
			httpClient.EXPECT().PutNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(http.StatusAccepted, okResp, nil)

			patch := ApplyFuncReturn(json.Unmarshal, errors.New("error"))
			defer patch.Reset()

			err := ema.UpdateEventTaskAttributesById(testCtx, eventTask)
			So(err.Error(), ShouldEqual, "error")
		})

		Convey("success", func() {
			okResp, _ := json.Marshal(eventTask)
			httpClient.EXPECT().PutNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(http.StatusNoContent, okResp, nil)

			err := ema.UpdateEventTaskAttributesById(testCtx, eventTask)
			So(err, ShouldBeNil)
		})
	})
}

func Test_EventModelAccess_GetEventModel(t *testing.T) {
	Convey("Test GetEventModel", t, func() {
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		ema, httpClient := MockNewEventModelAccess(mockCtl)

		Convey("failed, caused by http error", func() {
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				AnyTimes().Return(http.StatusInternalServerError, nil, fmt.Errorf("method failed"))

			expectEventModel := interfaces.EventModel{}
			eventModel, err := ema.GetEventModel(testCtx, eventTask.ModelID)
			So(err, ShouldResemble, fmt.Errorf("method failed"))
			So(eventModel, ShouldResemble, expectEventModel)
		})

		Convey("failed, caused by status != 200", func() {
			okResp, _ := json.Marshal(rest.BaseError{ErrorCode: "a", Description: "a", ErrorDetails: "a"})
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(http.StatusAccepted, okResp, nil)

			expectEventModel := interfaces.EventModel{}
			eventModel, err := ema.GetEventModel(testCtx, eventTask.ModelID)
			So(eventModel, ShouldResemble, expectEventModel)
			So(err, ShouldResemble, fmt.Errorf("get eventmodel by modelID 1 return error {\"error_code\":\"a\",\"description\":\"a\",\"solution\":\"\",\"error_link\":\"\",\"error_details\":\"a\"}"))
		})
		Convey("failed, caused unmarshal error", func() {
			okResp, _ := json.Marshal(rest.BaseError{ErrorCode: "a", Description: "a", ErrorDetails: "a"})
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(http.StatusAccepted, okResp, nil)

			patch := ApplyFuncReturn(json.Unmarshal, errors.New("error"))
			defer patch.Reset()

			expectEventModel := interfaces.EventModel{}
			eventModel, err := ema.GetEventModel(testCtx, eventTask.ModelID)
			So(eventModel, ShouldResemble, expectEventModel)
			So(err, ShouldResemble, errors.New("error"))
		})

		Convey("failed, caused by unmarshal httpErr error", func() {
			okResp, _ := json.Marshal(rest.BaseError{})
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(http.StatusAccepted, okResp, nil)

			patch := ApplyFunc(json.Unmarshal,
				func(data []byte, v any) error {
					if _, ok := v.(rest.BaseError); ok {
						return errors.New("error")
					}
					return nil
				},
			)
			defer patch.Reset()

			expectEventModel := interfaces.EventModel{}
			eventModel, err := ema.GetEventModel(testCtx, eventTask.ModelID)
			So(eventModel, ShouldResemble, expectEventModel)
			So(err.Error(), ShouldEqual, "error")
		})

		Convey("failed, caused by unmarshal eventModel error", func() {
			okResp, _ := json.Marshal([]interfaces.EventModel{eventModel})
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(http.StatusOK, okResp, nil)

			patch := ApplyFuncReturn(json.Unmarshal, errors.New("error"))
			defer patch.Reset()

			expectEventModel := interfaces.EventModel{}
			eventModel, err := ema.GetEventModel(testCtx, eventTask.ModelID)
			So(eventModel, ShouldResemble, expectEventModel)
			So(err.Error(), ShouldEqual, "error")
		})

		Convey("success", func() {
			okResp, _ := json.Marshal([]interfaces.EventModel{eventModel})
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(http.StatusOK, okResp, nil)

			_, err := ema.GetEventModel(testCtx, eventTask.ModelID)
			So(err, ShouldBeNil)
		})

		Convey("success with empty data", func() {
			okResp, _ := json.Marshal([]interfaces.EventModel{})
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(http.StatusOK, okResp, nil)

			_, err := ema.GetEventModel(testCtx, eventTask.ModelID)
			So(err, ShouldResemble, fmt.Errorf("EventModel 1 not found"))
		})
	})
}
