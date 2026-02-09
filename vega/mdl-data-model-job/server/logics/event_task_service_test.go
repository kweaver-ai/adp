// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package logics

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"data-model-job/common"
	"data-model-job/interfaces"
	dmock "data-model-job/interfaces/mock"
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
			IndexBase:    "event_task",
			DataViewName: "dip-event_task-data-view",
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
		Entities: []interfaces.Records{
			{
				map[string]any{
					"aaa": "xxx",
				},
			},
			{},
			nil,
		},
		TotalCount: 1,
	}
)

func MockNewEventTaskService(
	appSetting *common.AppSetting, emAccess interfaces.EventModelAccess, uAccess interfaces.UniqueryAccess,
	kAccess interfaces.KafkaAccess, iBAccess interfaces.IndexBaseAccess) *eventTaskService {

	return &eventTaskService{
		appSetting: appSetting,
		emAccess:   emAccess,
		uAccess:    uAccess,
		kAccess:    kAccess,
		iBAccess:   iBAccess,
	}
}

func Test_EventTaskService_EventTaskExecutor(t *testing.T) {
	Convey("Test EventTaskExecutor", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		emaMock := dmock.NewMockEventModelAccess(mockCtrl)
		uaMock := dmock.NewMockUniqueryAccess(mockCtrl)
		ibaMock := dmock.NewMockIndexBaseAccess(mockCtrl)
		kaMock := dmock.NewMockKafkaAccess(mockCtrl)
		etsMock := MockNewEventTaskService(appSetting, emaMock, uaMock, kaMock, ibaMock)

		producer := &kafka.Producer{}
		patchF2 := ApplyMethodReturn(producer, "Close")
		defer patchF2.Reset()

		Convey("failed, cause by GetEventModel error", func() {
			emaMock.EXPECT().GetEventModel(gomock.Any(), gomock.Any()).Return(interfaces.EventModel{}, errors.New("error"))
			emaMock.EXPECT().UpdateEventTaskAttributesById(gomock.Any(), gomock.Any()).Return(nil)

			msg := etsMock.EventTaskExecutor(testCtx, eventTask)
			So(msg, ShouldEqual, "error")
		})

		Convey("failed, cause by GetIndexBasesByTypes error", func() {
			emaMock.EXPECT().GetEventModel(gomock.Any(), gomock.Any()).Return(eventModel, nil)
			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{}, fmt.Errorf("error"))
			emaMock.EXPECT().UpdateEventTaskAttributesById(gomock.Any(), gomock.Any()).Return(nil)

			msg := etsMock.EventTaskExecutor(testCtx, eventTask)
			So(msg, ShouldEqual, "error")
		})

		Convey("failed, cause by length of index base != 1", func() {
			emaMock.EXPECT().GetEventModel(gomock.Any(), gomock.Any()).Return(eventModel, nil)
			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{
					{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "1"}, DataType: "1"},
					{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "2"}, DataType: "2"},
				}, nil)
			emaMock.EXPECT().UpdateEventTaskAttributesById(gomock.Any(), gomock.Any()).Return(nil)

			msg := etsMock.EventTaskExecutor(testCtx, eventTask)
			So(msg, ShouldEqual, "索引库类型[event_task]对应的索引库数量不等于1,为[2]")
		})

		Convey("failed, cause by GetEventModel error 2", func() {
			emaMock.EXPECT().GetEventModel(gomock.Any(), gomock.Any()).Return(eventModel, nil)
			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "1"}, DataType: "1"}}, nil)
			emaMock.EXPECT().GetEventModel(gomock.Any(), gomock.Any()).Return(interfaces.EventModel{}, fmt.Errorf("error"))
			emaMock.EXPECT().UpdateEventTaskAttributesById(gomock.Any(), gomock.Any()).Return(nil)

			msg := etsMock.EventTaskExecutor(testCtx, eventTask)
			So(msg, ShouldEqual, "error")
		})

		Convey("failed, cause by range query GetEventModelData error", func() {
			emaMock.EXPECT().GetEventModel(gomock.Any(), gomock.Any()).Return(eventModel, nil)
			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "1"}, DataType: "1"}}, nil)
			emaMock.EXPECT().GetEventModel(gomock.Any(), gomock.Any()).Return(eventModel, nil)
			uaMock.EXPECT().GetEventModelData(gomock.Any(), gomock.Any()).Return(interfaces.EventModelResponse{}, fmt.Errorf("error"))
			emaMock.EXPECT().UpdateEventTaskAttributesById(gomock.Any(), gomock.Any()).Return(nil)

			msg := etsMock.EventTaskExecutor(testCtx, eventTask)
			So(msg, ShouldEqual, "error")
		})

		Convey("failed, cause by instant query  GetEventModelData error", func() {
			emaMock.EXPECT().GetEventModel(gomock.Any(), gomock.Any()).Return(eventModel, nil)
			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "1"}, DataType: "1"}}, nil)
			emaMock.EXPECT().GetEventModel(gomock.Any(), gomock.Any()).Return(eventModel, nil)
			uaMock.EXPECT().GetEventModelData(gomock.Any(), gomock.Any()).Return(eventResponse, nil)
			uaMock.EXPECT().GetEventModelData(gomock.Any(), gomock.Any()).Return(interfaces.EventModelResponse{}, fmt.Errorf("error"))
			emaMock.EXPECT().UpdateEventTaskAttributesById(gomock.Any(), gomock.Any()).Return(nil)

			msg := etsMock.EventTaskExecutor(testCtx, eventTask)
			So(msg, ShouldEqual, "error")
		})

		Convey("failed, cause by marshal error", func() {
			emaMock.EXPECT().GetEventModel(gomock.Any(), gomock.Any()).Return(eventModel, nil)
			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "1"}, DataType: "1"}}, nil)
			emaMock.EXPECT().GetEventModel(gomock.Any(), gomock.Any()).Return(eventModel, nil)
			uaMock.EXPECT().GetEventModelData(gomock.Any(), gomock.Any()).AnyTimes().
				Return(eventResponse, nil)
			emaMock.EXPECT().UpdateEventTaskAttributesById(gomock.Any(), gomock.Any()).Return(nil)

			patch := ApplyFuncReturn(json.Marshal, []byte{}, errors.New("error"))
			defer patch.Reset()

			msg := etsMock.EventTaskExecutor(testCtx, eventTask)
			So(msg, ShouldEqual, "任务[0], 把数据发送到 kafka 失败[error]. ")
		})

		Convey("failed, cause by sendKafka NewTrxProducer error", func() {
			emaMock.EXPECT().GetEventModel(gomock.Any(), gomock.Any()).Return(eventModel, nil)
			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "1"}, DataType: "1"}}, nil)
			emaMock.EXPECT().GetEventModel(gomock.Any(), gomock.Any()).Return(eventModel, nil)
			uaMock.EXPECT().GetEventModelData(gomock.Any(), gomock.Any()).AnyTimes().
				Return(eventResponse, nil)
			kaMock.EXPECT().NewTrxProducer(gomock.Any()).AnyTimes().Return(&kafka.Producer{}, fmt.Errorf("error"))
			emaMock.EXPECT().UpdateEventTaskAttributesById(gomock.Any(), gomock.Any()).Return(nil)

			msg := etsMock.EventTaskExecutor(testCtx, eventTask)
			So(msg, ShouldEqual, "error")
		})

		Convey("failed, cause by  sendKafka DoProduceMsgToKafka error", func() {
			emaMock.EXPECT().GetEventModel(gomock.Any(), gomock.Any()).Return(eventModel, nil)
			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "1"}, DataType: "1"}}, nil)
			emaMock.EXPECT().GetEventModel(gomock.Any(), gomock.Any()).Return(eventModel, nil)
			uaMock.EXPECT().GetEventModelData(gomock.Any(), gomock.Any()).AnyTimes().
				Return(eventResponse, nil)
			kaMock.EXPECT().NewTrxProducer(gomock.Any()).AnyTimes().Return(producer, nil)
			kaMock.EXPECT().DoProduceMsgToKafka(gomock.Any(), gomock.Any()).AnyTimes().Return(fmt.Errorf("error1"))
			emaMock.EXPECT().UpdateEventTaskAttributesById(gomock.Any(), gomock.Any()).Return(nil)

			msg := etsMock.EventTaskExecutor(testCtx, eventTask)
			So(msg, ShouldEqual, "error1")
		})

		Convey("failed, cause by UpdateEventTaskAttributesById error", func() {
			emaMock.EXPECT().GetEventModel(gomock.Any(), gomock.Any()).Return(eventModel, nil)
			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "1"}, DataType: "1"}}, nil)
			emaMock.EXPECT().GetEventModel(gomock.Any(), gomock.Any()).Return(eventModel, nil)
			uaMock.EXPECT().GetEventModelData(gomock.Any(), gomock.Any()).AnyTimes().
				Return(eventResponse, nil)
			kaMock.EXPECT().NewTrxProducer(gomock.Any()).AnyTimes().Return(producer, nil)
			kaMock.EXPECT().DoProduceMsgToKafka(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			emaMock.EXPECT().UpdateEventTaskAttributesById(gomock.Any(), gomock.Any()).AnyTimes().Return(fmt.Errorf("error1"))

			msg := etsMock.EventTaskExecutor(testCtx, eventTask)
			So(msg, ShouldEqual, "任务[1], 更新任务执行状态失败[error1]. :error1")
		})

		Convey("failed ,start is not int64", func() {
			eventTask := interfaces.EventTask{
				TaskID:   "1",
				ModelID:  "1",
				Schedule: interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
				StorageConfig: interfaces.StorageConfig{
					IndexBase:    "event_task",
					DataViewName: "dip-event_task-data-view",
				},
				DispatchConfig: interfaces.DispatchConfig{
					TimeOut:        3600,
					FailRetryCount: 3,
					RouteStrategy:  "FIRST",
					BlockStrategy:  "SERIAL_EXECUTION",
				},
				ExecuteParameter: map[string]any{
					"key":   1,
					"start": "65165asdasd",
					"end":   "1729239251000",
				},
				ScheduleSyncStatus: 1,
				StatusUpdateTime:   1699336878575,
				TaskStatus:         4,
				ErrorDetails:       "",
				UpdateTime:         1699336878575,
			}

			emaMock.EXPECT().GetEventModel(gomock.Any(), gomock.Any()).Return(eventModel, nil)
			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "1"}, DataType: "1"}}, nil)
			emaMock.EXPECT().GetEventModel(gomock.Any(), gomock.Any()).Return(eventModel, nil)
			uaMock.EXPECT().GetEventModelData(gomock.Any(), gomock.Any()).AnyTimes().
				Return(eventResponse, nil)
			emaMock.EXPECT().UpdateEventTaskAttributesById(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			msg := etsMock.EventTaskExecutor(testCtx, eventTask)
			So(msg, ShouldNotEqual, "")
		})

		Convey("success", func() {
			emaMock.EXPECT().GetEventModel(gomock.Any(), gomock.Any()).Return(eventModel, nil)
			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "1"}, DataType: "1"}}, nil)
			emaMock.EXPECT().GetEventModel(gomock.Any(), gomock.Any()).Return(eventModel, nil)
			uaMock.EXPECT().GetEventModelData(gomock.Any(), gomock.Any()).AnyTimes().
				Return(eventResponse, nil)
			kaMock.EXPECT().NewTrxProducer(gomock.Any()).AnyTimes().Return(producer, nil)
			kaMock.EXPECT().DoProduceMsgToKafka(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			emaMock.EXPECT().UpdateEventTaskAttributesById(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			msg := etsMock.EventTaskExecutor(testCtx, eventTask)
			So(msg, ShouldNotEqual, "")
		})

		Convey("success and ExecuteParameter has start and end", func() {
			eventTask := interfaces.EventTask{
				TaskID:   "1",
				ModelID:  "1",
				Schedule: interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
				StorageConfig: interfaces.StorageConfig{
					IndexBase:    "event_task",
					DataViewName: "dip-event_task-data-view",
				},
				DispatchConfig: interfaces.DispatchConfig{
					TimeOut:        3600,
					FailRetryCount: 3,
					RouteStrategy:  "FIRST",
					BlockStrategy:  "SERIAL_EXECUTION",
				},
				ExecuteParameter: map[string]any{
					"key":   1,
					"start": "1729239238000",
					"end":   "1729239251000",
				},
				ScheduleSyncStatus: 1,
				StatusUpdateTime:   1699336878575,
				TaskStatus:         4,
				ErrorDetails:       "",
				UpdateTime:         1699336878575,
			}

			emaMock.EXPECT().GetEventModel(gomock.Any(), gomock.Any()).Return(eventModel, nil)
			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "1"}, DataType: "1"}}, nil)
			emaMock.EXPECT().GetEventModel(gomock.Any(), gomock.Any()).Return(eventModel, nil)
			uaMock.EXPECT().GetEventModelData(gomock.Any(), gomock.Any()).AnyTimes().
				Return(eventResponse, nil)
			kaMock.EXPECT().NewTrxProducer(gomock.Any()).AnyTimes().Return(producer, nil)
			kaMock.EXPECT().DoProduceMsgToKafka(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			emaMock.EXPECT().UpdateEventTaskAttributesById(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			msg := etsMock.EventTaskExecutor(testCtx, eventTask)
			So(msg, ShouldNotEqual, "")
		})

		Convey("success with empty data", func() {
			emaMock.EXPECT().GetEventModel(gomock.Any(), gomock.Any()).Return(eventModel, nil)
			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "1"}, DataType: "1"}}, nil)
			emaMock.EXPECT().GetEventModel(gomock.Any(), gomock.Any()).Return(eventModel, nil)
			uaMock.EXPECT().GetEventModelData(gomock.Any(), gomock.Any()).AnyTimes().
				Return(interfaces.EventModelResponse{}, nil)
			kaMock.EXPECT().NewTrxProducer(gomock.Any()).AnyTimes().Return(producer, nil)
			kaMock.EXPECT().DoProduceMsgToKafka(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			emaMock.EXPECT().UpdateEventTaskAttributesById(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			msg := etsMock.EventTaskExecutor(testCtx, eventTask)
			So(msg, ShouldEqual, "任务[1], 获取到的数据为空. ")
		})

		Convey("success with agi ", func() {
			eventModel := interfaces.EventModel{
				EventModelID:        "1",
				EventModelName:      "测试中的名称",
				EventModelType:      "atomic",
				UpdateTime:          1699336878575,
				EventModelTags:      []string{"xx1", "xx2"},
				DataSourceType:      "metric_model",
				DataSource:          []string{"1"},
				DataSourceName:      []string{},
				DataSourceGroupName: []string{},
				DetectRule:          interfaces.DetectRule{DetectRuleId: "0", Priority: 0, Type: "agi_detect", Formula: []interfaces.FormulaItem{}},
				AggregateRule:       interfaces.AggregateRule{GroupFields: []string{}},
				DefaultTimeWindow:   interfaces.TimeInterval{Interval: 5, Unit: "m"},
				EventModelComment:   "comment",
				IsActive:            1,
				IsCustom:            1,
				Task:                eventTask,
			}
			emaMock.EXPECT().GetEventModel(gomock.Any(), gomock.Any()).Return(eventModel, nil)
			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "1"}, DataType: "1"}}, nil)
			emaMock.EXPECT().GetEventModel(gomock.Any(), gomock.Any()).Return(eventModel, nil)
			uaMock.EXPECT().GetEventModelData(gomock.Any(), gomock.Any()).AnyTimes().
				Return(interfaces.EventModelResponse{}, nil)
			emaMock.EXPECT().UpdateEventTaskAttributesById(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			msg := etsMock.EventTaskExecutor(testCtx, eventTask)
			So(msg, ShouldEqual, "任务[1], 获取到的数据为空. ")
		})
	})
}
