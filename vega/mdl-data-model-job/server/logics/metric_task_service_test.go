// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package logics

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"testing"
	"time"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"data-model-job/common"
	"data-model-job/interfaces"
	dmock "data-model-job/interfaces/mock"
)

var (
	task = interfaces.MetricTask{
		TaskID:     "uint64(1)",
		TaskName:   "task1",
		ModuleType: interfaces.MODULE_TYPE_METRIC_MODEL,
		ModelID:    "1",
		Schedule:   interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
		// Filters:            []interfaces.Filter{{Name: "labels.mode", Operation: "=", Value: "idle"}},
		TimeWindows:        []string{"5m", "1h"},
		Steps:              []string{"5m"},
		IndexBase:          "base1",
		RetraceDuration:    "1d",
		ScheduleSyncStatus: 1,
		Comment:            "task1-aaa",
		UpdateTime:         1699336878575,
		PlanTime:           int64(1699336878575),
	}

	uniqueryData = interfaces.UniResponse{
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
)

func MockNewMetricTaskService(
	appSetting *common.AppSetting, mmAccess interfaces.MetricModelAccess, uAccess interfaces.UniqueryAccess,
	kAccess interfaces.KafkaAccess, iBAccess interfaces.IndexBaseAccess) *metricTaskService {

	return &metricTaskService{
		appSetting: appSetting,
		mmAccess:   mmAccess,
		uAccess:    uAccess,
		kAccess:    kAccess,
		iBAccess:   iBAccess,
	}
}

func Test_MetricTaskService_MetricTaskExecutor(t *testing.T) {
	Convey("Test MetricTaskExecutor", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mmaMock := dmock.NewMockMetricModelAccess(mockCtrl)
		uaMock := dmock.NewMockUniqueryAccess(mockCtrl)
		ibaMock := dmock.NewMockIndexBaseAccess(mockCtrl)
		kaMock := dmock.NewMockKafkaAccess(mockCtrl)
		mtsMock := MockNewMetricTaskService(appSetting, mmaMock, uaMock, kaMock, ibaMock)

		producer := &kafka.Producer{}
		patchF2 := ApplyMethod(reflect.TypeOf(producer), "Close",
			func(*kafka.Producer) {
				// nothing to do
			},
		)
		defer patchF2.Reset()

		Convey("failed, cause by GetIndexBasesByTypes error", func() {
			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{}, fmt.Errorf("error"))

			mmaMock.EXPECT().UpdateTaskAttributesById(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			msg := mtsMock.MetricTaskExecutor(testCtx, task)
			So(msg, ShouldEqual, "error")
		})

		Convey("failed, cause by length of index base != 1", func() {
			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{
					{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "1"}, DataType: "1"},
					{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "1"}, DataType: "2"},
				}, nil)

			mmaMock.EXPECT().UpdateTaskAttributesById(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			msg := mtsMock.MetricTaskExecutor(testCtx, task)
			So(msg, ShouldEqual, "指标模型任务的索引库类型[base1]对应的索引库数量不等于1,为[2]")
		})

		Convey("failed, cause by GetTaskPlanTimeById error", func() {
			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "1"}, DataType: "1"}}, nil)
			mmaMock.EXPECT().GetTaskPlanTimeById(gomock.Any(), gomock.Any()).Return(int64(0), fmt.Errorf("error"))

			mmaMock.EXPECT().UpdateTaskAttributesById(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			msg := mtsMock.MetricTaskExecutor(testCtx, task)
			So(msg, ShouldEqual, "error")
		})

		// Convey("failed, cause by length of plan_times == 0", func() {
		// 	ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
		// 		Return([]interfaces.IndexBase{{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "1"}, DataType: "1"}}, nil)
		// 	mmaMock.EXPECT().GetTaskPlanTimeById(gomock.Any(), gomock.Any()).Return(int64(1699336878575), nil)

		// 	mmaMock.EXPECT().UpdateTaskAttributesById(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		// 	msg := mtsMock.MetricTaskExecutor(testCtx, task)
		// 	So(msg, ShouldEqual, "指标模型任务的计划时间【[1699336878575 1699336888575]】不合法, 期望为计划时间长度为1")
		// })

		Convey("failed, cause by invalid step", func() {
			taskTmp := interfaces.MetricTask{
				TaskID:     "uint64(1)",
				TaskName:   "task1",
				ModuleType: interfaces.MODULE_TYPE_METRIC_MODEL,
				ModelID:    "1",
				Schedule:   interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
				// Filters:            []interfaces.Filter{{Name: "labels.mode", Operation: "=", Value: "idle"}},
				TimeWindows:        []string{"5m", "1h"},
				Steps:              []string{"5M"},
				IndexBase:          "base1",
				RetraceDuration:    "1d",
				ScheduleSyncStatus: 1,
				Comment:            "task1-aaa",
				UpdateTime:         1699336878575,
				PlanTime:           int64(1699336878575),
			}

			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "1"}, DataType: "1"}}, nil)
			mmaMock.EXPECT().GetTaskPlanTimeById(gomock.Any(), gomock.Any()).Return(int64(
				time.Now().Add(time.Duration(time.Minute*10)).UnixNano()/int64(time.Millisecond/time.Nanosecond)), nil)

			mmaMock.EXPECT().UpdateTaskAttributesById(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			msg := mtsMock.MetricTaskExecutor(testCtx, taskTmp)
			So(msg, ShouldEqual, "failed to parse schedule duration, err: not a valid duration string: \"5M\"")
		})

		Convey("success, cause by fixed_now > fixed_plan_time, skip", func() {
			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "1"}, DataType: "1"}}, nil)
			mmaMock.EXPECT().GetTaskPlanTimeById(gomock.Any(), gomock.Any()).Return(int64(
				time.Now().Add(time.Duration(time.Minute*10)).UnixNano()/int64(time.Millisecond/time.Nanosecond)), nil)

			msg := mtsMock.MetricTaskExecutor(testCtx, task)
			So(msg, ShouldNotEqual, "")
		})

		Convey("failed, cause by GetMetricModelData error", func() {
			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "1"}, DataType: "1"}}, nil)
			mmaMock.EXPECT().GetTaskPlanTimeById(gomock.Any(), gomock.Any()).Return(int64(1699336878575), nil)
			uaMock.EXPECT().GetMetricModelData(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().
				Return(interfaces.UniResponse{}, fmt.Errorf("error"))

			mmaMock.EXPECT().UpdateTaskAttributesById(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			msg := mtsMock.MetricTaskExecutor(testCtx, task)
			So(msg, ShouldEqual, "指标模型任务[uint64(1)], 查询参数: time=1699336800000,look_back_delta=5m, 获取指标数据失败[error]. ")
		})

		Convey("failed, cause by marshal error", func() {
			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "1"}, DataType: "1"}}, nil)
			mmaMock.EXPECT().GetTaskPlanTimeById(gomock.Any(), gomock.Any()).Return(int64(1699336878575), nil)
			uaMock.EXPECT().GetMetricModelData(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().
				Return(uniqueryData, nil)

			mmaMock.EXPECT().UpdateTaskAttributesById(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			patch := ApplyFunc(json.Marshal,
				func(v any) ([]byte, error) {
					return []byte{}, errors.New("error")
				},
			)
			defer patch.Reset()

			msg := mtsMock.MetricTaskExecutor(testCtx, task)
			So(msg, ShouldEqual, "error")
		})

		Convey("failed, cause by NewTrxProducer error", func() {
			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "1"}, DataType: "1"}}, nil)
			mmaMock.EXPECT().GetTaskPlanTimeById(gomock.Any(), gomock.Any()).Return(int64(1699336878575), nil)
			uaMock.EXPECT().GetMetricModelData(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().
				Return(uniqueryData, nil)
			kaMock.EXPECT().NewTrxProducer(gomock.Any()).AnyTimes().Return(&kafka.Producer{}, fmt.Errorf("error"))

			mmaMock.EXPECT().UpdateTaskAttributesById(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			msg := mtsMock.MetricTaskExecutor(testCtx, task)
			So(msg, ShouldEqual, "指标模型任务[uint64(1)], 把数据发送到 kafka 失败[error]. ")
		})

		Convey("failed, cause by DoProduceMsgToKafka error", func() {
			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "1"}, DataType: "1"}}, nil)
			mmaMock.EXPECT().GetTaskPlanTimeById(gomock.Any(), gomock.Any()).Return(int64(1699336878575), nil)
			uaMock.EXPECT().GetMetricModelData(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().
				Return(uniqueryData, nil)
			kaMock.EXPECT().NewTrxProducer(gomock.Any()).AnyTimes().Return(producer, nil)
			kaMock.EXPECT().DoProduceMsgToKafka(gomock.Any(), gomock.Any()).AnyTimes().Return(fmt.Errorf("error1"))

			mmaMock.EXPECT().UpdateTaskAttributesById(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			msg := mtsMock.MetricTaskExecutor(testCtx, task)
			So(msg, ShouldEqual, "指标模型任务[uint64(1)], 把数据发送到 kafka 失败[error1]. ")
		})

		Convey("failed, cause by UpdateTaskAttributesById error", func() {
			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "1"}, DataType: "1"}}, nil)
			mmaMock.EXPECT().GetTaskPlanTimeById(gomock.Any(), gomock.Any()).Return(int64(1699336878575), nil)
			uaMock.EXPECT().GetMetricModelData(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().
				Return(uniqueryData, nil)
			kaMock.EXPECT().NewTrxProducer(gomock.Any()).AnyTimes().Return(producer, nil)
			kaMock.EXPECT().DoProduceMsgToKafka(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			mmaMock.EXPECT().UpdateTaskAttributesById(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(fmt.Errorf("error1"))

			msg := mtsMock.MetricTaskExecutor(testCtx, task)
			So(msg, ShouldEqual, "指标模型任务[uint64(1)], 更新任务计划时间失败[error1]. :error1")
		})

		Convey("success", func() {

			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "1"}, DataType: "1"}}, nil)
			mmaMock.EXPECT().GetTaskPlanTimeById(gomock.Any(), gomock.Any()).Return(int64(time.Now().UnixMilli()-600000), nil)
			uaMock.EXPECT().GetMetricModelData(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().
				Return(uniqueryData, nil)
			kaMock.EXPECT().NewTrxProducer(gomock.Any()).AnyTimes().Return(producer, nil)
			kaMock.EXPECT().DoProduceMsgToKafka(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			mmaMock.EXPECT().UpdateTaskAttributesById(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			msg := mtsMock.MetricTaskExecutor(testCtx, task)
			So(msg, ShouldNotEqual, "")
		})

		Convey("success with empty data", func() {
			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "1"}, DataType: "1"}}, nil)
			mmaMock.EXPECT().GetTaskPlanTimeById(gomock.Any(), gomock.Any()).Return(int64(time.Now().UnixMilli()-600000), nil)
			uaMock.EXPECT().GetMetricModelData(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().
				Return(interfaces.UniResponse{}, nil)
			kaMock.EXPECT().NewTrxProducer(gomock.Any()).AnyTimes().Return(producer, nil)
			kaMock.EXPECT().DoProduceMsgToKafka(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			mmaMock.EXPECT().UpdateTaskAttributesById(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			msg := mtsMock.MetricTaskExecutor(testCtx, task)
			So(msg, ShouldEqual, "service: 指标模型任务[uint64(1)]执行完成. 共发送[0]条数据到kafka。")
		})

		Convey("success with empty data promql", func() {
			taskTmp := interfaces.MetricTask{
				TaskID:     "uint64(1)",
				TaskName:   "task1",
				ModuleType: interfaces.MODULE_TYPE_METRIC_MODEL,
				ModelID:    "1",
				Schedule:   interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
				// Filters:            []interfaces.Filter{{Name: "labels.mode", Operation: "=", Value: "idle"}},
				Steps:              []string{"5m"},
				IndexBase:          "base1",
				RetraceDuration:    "1d",
				ScheduleSyncStatus: 1,
				Comment:            "task1-aaa",
				UpdateTime:         1699336878575,
				PlanTime:           int64(1699336878575),
			}

			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "1"}, DataType: "1"}}, nil)
			mmaMock.EXPECT().GetTaskPlanTimeById(gomock.Any(), gomock.Any()).Return(int64(
				time.Now().Add(time.Duration(time.Minute*10)).UnixNano()/int64(time.Millisecond/time.Nanosecond)), nil)
			uaMock.EXPECT().GetMetricModelData(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().
				Return(interfaces.UniResponse{}, nil)
			kaMock.EXPECT().NewTrxProducer(gomock.Any()).AnyTimes().Return(producer, nil)
			kaMock.EXPECT().DoProduceMsgToKafka(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			mmaMock.EXPECT().UpdateTaskAttributesById(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			msg := mtsMock.MetricTaskExecutor(testCtx, taskTmp)
			So(msg, ShouldEqual, "service: 指标模型任务[uint64(1)]执行完成. 共发送[0]条数据到kafka。")
		})

		Convey("failed, cause by objective model task error", func() {
			taskTmp := interfaces.MetricTask{
				TaskID:             "uint64(1)",
				TaskName:           "task1",
				ModuleType:         interfaces.MODULE_TYPE_OBJECTIVE_MODEL,
				ModelID:            "1",
				Schedule:           interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
				Steps:              []string{"5m"},
				IndexBase:          "base1",
				RetraceDuration:    "1d",
				ScheduleSyncStatus: 1,
				Comment:            "task1-aaa",
				UpdateTime:         1699336878575,
				PlanTime:           int64(1699336878575),
			}

			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "1"}, DataType: "1"}}, nil)
			mmaMock.EXPECT().GetTaskPlanTimeById(gomock.Any(), gomock.Any()).Return(int64(1699336878575), nil)
			uaMock.EXPECT().GetObjectiveModelData(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().
				Return(interfaces.ObjectiveModelUniResponse{}, fmt.Errorf("error"))
			mmaMock.EXPECT().UpdateTaskAttributesById(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			msg := mtsMock.MetricTaskExecutor(testCtx, taskTmp)
			So(msg, ShouldEqual, "目标模型任务[uint64(1)], 查询参数: time=1699336800000,look_back_delta=5m, 获取指标数据失败[error]. ")
		})

		Convey("success, objective model task", func() {
			taskTmp := interfaces.MetricTask{
				TaskID:             "uint64(1)",
				TaskName:           "task1",
				ModuleType:         interfaces.MODULE_TYPE_OBJECTIVE_MODEL,
				ModelID:            "1",
				Schedule:           interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
				Steps:              []string{"5m"},
				IndexBase:          "base1",
				RetraceDuration:    "1d",
				ScheduleSyncStatus: 1,
				Comment:            "task1-aaa",
				UpdateTime:         1699336878575,
				PlanTime:           int64(1699336878575),
			}

			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "1"}, DataType: "1"}}, nil)
			mmaMock.EXPECT().GetTaskPlanTimeById(gomock.Any(), gomock.Any()).Return(int64(
				time.Now().Add(-1*time.Duration(time.Minute*5)).UnixNano()/int64(time.Millisecond/time.Nanosecond)), nil)
			uaMock.EXPECT().GetObjectiveModelData(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().
				Return(interfaces.ObjectiveModelUniResponse{
					Model: interfaces.ObjectiveModel{
						ObjectiveType: "slo",
					},
					Datas: []interfaces.SLOObjectiveData{
						{
							Labels:     map[string]string{"label1": "value1"},
							Times:      []any{123456},
							SLI:        []any{1.23},
							Objective:  []float64{1.23},
							Good:       []any{1.23},
							Total:      []any{1.23},
							AchiveRate: []any{1.23},
							Period:     []int64{123456},
						},
					},
				}, nil)
			kaMock.EXPECT().NewTrxProducer(gomock.Any()).AnyTimes().Return(producer, nil)
			kaMock.EXPECT().DoProduceMsgToKafka(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			mmaMock.EXPECT().UpdateTaskAttributesById(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			msg := mtsMock.MetricTaskExecutor(testCtx, taskTmp)
			So(msg, ShouldEqual, "service: 目标模型[uint64(1)]执行完成. 共发送[2]条数据到kafka。")
		})
	})
}

func Test_MetricTaskService_ExecutObjectiveTask(t *testing.T) {
	Convey("Test ExecutObjectiveTask", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mmaMock := dmock.NewMockMetricModelAccess(mockCtrl)
		uaMock := dmock.NewMockUniqueryAccess(mockCtrl)
		ibaMock := dmock.NewMockIndexBaseAccess(mockCtrl)
		kaMock := dmock.NewMockKafkaAccess(mockCtrl)
		mtsMock := MockNewMetricTaskService(appSetting, mmaMock, uaMock, kaMock, ibaMock)

		producer := &kafka.Producer{}
		patchF2 := ApplyMethod(reflect.TypeOf(producer), "Close",
			func(*kafka.Producer) {
				// nothing to do
			},
		)
		defer patchF2.Reset()

		Convey("failed, cause by GetIndexBasesByTypes error", func() {
			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{}, fmt.Errorf("error"))

			msg, err := mtsMock.executObjectiveTask(testCtx, task)
			So(err, ShouldNotBeNil)
			So(msg, ShouldEqual, "")
		})

		Convey("failed, cause by length of index base != 1", func() {
			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{
					{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "1"}, DataType: "1"},
					{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "2"}, DataType: "2"},
				}, nil)

			_, err := mtsMock.executObjectiveTask(testCtx, task)
			So(err, ShouldNotBeNil)
		})

		Convey("failed, cause by GetTaskPlanTimeById error", func() {
			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "1"}, DataType: "1"}}, nil)

			mmaMock.EXPECT().GetTaskPlanTimeById(gomock.Any(), gomock.Any()).Return(int64(0), fmt.Errorf("error"))

			_, err := mtsMock.executObjectiveTask(testCtx, task)
			So(err, ShouldNotBeNil)
		})

		Convey("failed, cause by ParseDuration error", func() {
			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "1"}, DataType: "1"}}, nil)

			mmaMock.EXPECT().GetTaskPlanTimeById(gomock.Any(), gomock.Any()).Return(int64(
				time.Now().Add(time.Duration(time.Minute*10)).UnixNano()/int64(time.Millisecond/time.Nanosecond)), nil)

			patch := ApplyFunc(common.ParseDuration,
				func(duration string, re *regexp.Regexp, allowZero bool) (time.Duration, error) {
					return 0, errors.New("error")
				},
			)
			defer patch.Reset()

			_, err := mtsMock.executObjectiveTask(testCtx, task)
			So(err, ShouldNotBeNil)
		})

		Convey("failed, cause by LoadLocation error", func() {
			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "1"}, DataType: "1"}}, nil)

			mmaMock.EXPECT().GetTaskPlanTimeById(gomock.Any(), gomock.Any()).Return(int64(
				time.Now().Add(time.Duration(time.Minute*10)).UnixNano()/int64(time.Millisecond/time.Nanosecond)), nil)

			patch := ApplyFunc(time.LoadLocation,
				func(name string) (*time.Location, error) {
					return nil, errors.New("error")
				},
			)
			defer patch.Reset()

			_, err := mtsMock.executObjectiveTask(testCtx, task)
			So(err, ShouldNotBeNil)
		})

		Convey("failed, cause by GetObjectiveModelData error", func() {
			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "1"}, DataType: "1"}}, nil)

			mmaMock.EXPECT().GetTaskPlanTimeById(gomock.Any(), gomock.Any()).Return(int64(
				time.Now().Add(-1*time.Duration(time.Minute*10)).UnixNano()/int64(time.Millisecond/time.Nanosecond)), nil)

			uaMock.EXPECT().GetObjectiveModelData(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(interfaces.ObjectiveModelUniResponse{}, fmt.Errorf("error"))

			_, err := mtsMock.executObjectiveTask(testCtx, task)
			So(err, ShouldNotBeNil)
		})

		Convey("failed, cause by objectiveDataTransferToMessage error", func() {
			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "1"}, DataType: "1"}}, nil)

			mmaMock.EXPECT().GetTaskPlanTimeById(gomock.Any(), gomock.Any()).Return(int64(
				time.Now().Add(-1*time.Duration(time.Minute*10)).UnixNano()/int64(time.Millisecond/time.Nanosecond)), nil)

			uaMock.EXPECT().GetObjectiveModelData(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(interfaces.ObjectiveModelUniResponse{Datas: []interfaces.ObjectiveModelUniResponse{{}}}, nil)

			_, err := mtsMock.executObjectiveTask(testCtx, task)
			So(err, ShouldNotBeNil)
		})

		Convey("failed, cause by flushToKafka error", func() {
			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "1"}, DataType: "1"}}, nil)

			mmaMock.EXPECT().GetTaskPlanTimeById(gomock.Any(), gomock.Any()).Return(int64(
				time.Now().Add(-1*time.Duration(time.Minute*10)).UnixNano()/int64(time.Millisecond/time.Nanosecond)), nil)

			uaMock.EXPECT().GetObjectiveModelData(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().
				Return(interfaces.ObjectiveModelUniResponse{
					Model: interfaces.ObjectiveModel{
						ObjectiveType: "slo",
					},
					Datas: []interfaces.SLOObjectiveData{
						{
							Labels:     map[string]string{"label1": "value1"},
							Times:      []any{123456},
							SLI:        []any{1.23},
							Objective:  []float64{1.23},
							Good:       []any{1.23},
							Total:      []any{1.23},
							AchiveRate: []any{1.23},
							Period:     []int64{123456},
						},
					},
				}, nil)

			kaMock.EXPECT().NewTrxProducer(gomock.Any()).Return(nil, fmt.Errorf("error"))

			_, err := mtsMock.executObjectiveTask(testCtx, task)
			So(err, ShouldNotBeNil)
		})
		Convey("failed, cause by UpdateTaskAttributesById error", func() {
			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]interfaces.IndexBase{{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "1"}, DataType: "1"}}, nil)

			mmaMock.EXPECT().GetTaskPlanTimeById(gomock.Any(), gomock.Any()).Return(int64(
				time.Now().Add(-1*time.Duration(time.Minute*10)).UnixNano()/int64(time.Millisecond/time.Nanosecond)), nil)

			uaMock.EXPECT().GetObjectiveModelData(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().
				Return(interfaces.ObjectiveModelUniResponse{
					Model: interfaces.ObjectiveModel{
						ObjectiveType: "kpi",
					},
					Datas: []interfaces.KPIObjectiveData{
						{
							Labels:              map[string]string{"label1": "value1"},
							Times:               []any{123456},
							KPI:                 []any{1.23},
							Objective:           []float64{1.23},
							AchiveRate:          []any{1.23},
							KPIScore:            []any{1.23},
							AssociateMetricNums: []any{1.23},
							Status:              []string{"1"},
							StatusCode:          []int{1},
						},
					},
				}, nil)

			kaMock.EXPECT().NewTrxProducer(gomock.Any()).Return(&kafka.Producer{}, nil)
			kaMock.EXPECT().DoProduceMsgToKafka(gomock.Any(), gomock.Any()).Return(nil)

			mmaMock.EXPECT().UpdateTaskAttributesById(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(fmt.Errorf("error"))

			_, err := mtsMock.executObjectiveTask(testCtx, task)
			So(err, ShouldNotBeNil)
		})

	})
}
