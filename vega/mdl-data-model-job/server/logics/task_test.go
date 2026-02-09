// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package logics

import (
	"context"
	"errors"
	"testing"
	"time"

	libmq "github.com/kweaver-ai/kweaver-go-lib/mq"
	. "github.com/agiledragon/gomonkey/v2"
	"github.com/bytedance/sonic"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"data-model-job/common"
	cond "data-model-job/common/condition"
	"data-model-job/interfaces"
	dmock "data-model-job/interfaces/mock"
)

func NewTestTask(kaMock interfaces.KafkaAccess) *Task {
	taskTest := &Task{
		appSetting: &common.AppSetting{
			MQSetting: libmq.MQSetting{
				Tenant: "default",
			},
		},
		kAccess:        kaMock,
		jobId:          "1a",
		srcTopic:       "test",
		sinkTopic:      "default.mdl.view.1",
		status:         interfaces.TaskStatus_Running,
		RunningChannel: make(chan bool),
		DataView:       &interfaces.DataView{},
	}

	return taskTest
}

func Test_Task_Run(t *testing.T) {
	Convey("Test Task Run", t, func() {
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		kaMock := dmock.NewMockKafkaAccess(mockCtl)
		taskTest := NewTestTask(kaMock)

		Convey("execute failed", func() {
			go func() {
				<-taskTest.RunningChannel
			}()

			patches := ApplyPrivateMethod(taskTest, "execute",
				func(*Task, context.Context) error {
					return errors.New("error")
				})
			defer patches.Reset()

			err := taskTest.Run()
			So(err, ShouldNotBeNil)
		})

		Convey("success", func() {
			go func() {
				<-taskTest.RunningChannel
			}()

			patches := ApplyPrivateMethod(taskTest, "execute",
				func(*Task, context.Context) error {
					return nil
				})
			defer patches.Reset()

			err := taskTest.Run()
			So(err, ShouldBeNil)
		})
	})
}

func Test_Task_String(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	kaMock := dmock.NewMockKafkaAccess(mockCtl)
	taskTest := NewTestTask(kaMock)

	Convey("Test Task Print", t, func() {
		Convey("success", func() {
			res := taskTest.String()
			So(res, ShouldContainSubstring, "task_status")
		})
	})
}

func Test_Task_Stop(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	kaMock := dmock.NewMockKafkaAccess(mockCtl)
	taskTest := NewTestTask(kaMock)

	Convey("Test Task Stop", t, func() {
		Convey("stop a stopped taskTest", func() {
			taskTest.status = interfaces.TaskStatus_Stopped
			err := taskTest.Stop()
			So(err, ShouldNotBeNil)
		})

		Convey("stop a running taskTest", func() {
			taskTest.status = interfaces.TaskStatus_Running
			go func() {
				taskTest.RunningChannel <- false
			}()
			err := taskTest.Stop()
			So(err, ShouldBeNil)
		})
	})
}

func Test_Task_execute(t *testing.T) {
	Convey("Test Task execute", t, func() {
		FailureThreshold = 2
		RetryInterval = 1 * time.Second
		FlushBytes = 10
		FlushInterval = 0

		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		kaMock := dmock.NewMockKafkaAccess(mockCtl)
		taskTest := NewTestTask(kaMock)

		c, _ := kafka.NewConsumer(&kafka.ConfigMap{"bootstrap.servers": "localhost"})
		p, _ := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost"})

		msg := &kafka.Message{
			Timestamp: time.Now(),
			Value:     []byte("{\"__index_base\":\"1234567890abcdefg\"}"),
		}

		patches1 := ApplyMethodReturn(taskTest, "CloseConsumer")
		defer patches1.Reset()

		patches2 := ApplyMethodReturn(taskTest, "CloseProducer")
		defer patches2.Reset()

		Convey("NewConsumer failed", func() {
			kaMock.EXPECT().NewConsumer(gomock.Any()).Return(c, errors.New("error"))

			err := taskTest.execute()
			So(err, ShouldNotBeNil)
		})

		Convey("NewProducer failed", func() {
			kaMock.EXPECT().NewConsumer(gomock.Any()).Return(c, nil)
			kaMock.EXPECT().NewTransactionalProducer(gomock.Any()).Return(p, errors.New("error"))

			err := taskTest.execute()
			So(err, ShouldNotBeNil)
		})

		Convey("Subscribe failed", func() {
			kaMock.EXPECT().NewConsumer(gomock.Any()).Return(c, nil)
			kaMock.EXPECT().NewTransactionalProducer(gomock.Any()).Return(p, nil)

			patches := ApplyMethodReturn(c, "Subscribe", errors.New("error"))
			defer patches.Reset()

			err := taskTest.execute()
			So(err, ShouldNotBeNil)
		})

		Convey("Init transaction failed", func() {
			kaMock.EXPECT().NewConsumer(gomock.Any()).Return(c, nil)
			kaMock.EXPECT().NewTransactionalProducer(gomock.Any()).Return(p, nil)

			patches1 := ApplyMethodReturn(c, "Subscribe", nil)
			defer patches1.Reset()

			patches2 := ApplyMethodReturn(p, "InitTransactions", errors.New("error"))
			defer patches2.Reset()

			err := taskTest.execute()
			So(err, ShouldNotBeNil)
		})

		Convey("DoConsume failed", func() {
			kaMock.EXPECT().NewConsumer(gomock.Any()).Return(c, nil)
			kaMock.EXPECT().NewTransactionalProducer(gomock.Any()).Return(p, nil)

			patches1 := ApplyMethodReturn(c, "Subscribe", nil)
			defer patches1.Reset()

			patches2 := ApplyMethodReturn(p, "InitTransactions", nil)
			defer patches2.Reset()

			kaMock.EXPECT().DoConsume(c).Return(nil, errors.New("error")).AnyTimes()

			err := taskTest.execute()
			So(err, ShouldNotBeNil)
		})

		Convey("packagingMessages failed", func() {
			kaMock.EXPECT().NewConsumer(gomock.Any()).Return(c, nil)
			kaMock.EXPECT().NewTransactionalProducer(gomock.Any()).Return(p, nil)

			patches1 := ApplyMethodReturn(c, "Subscribe", nil)
			defer patches1.Reset()

			patches2 := ApplyMethodReturn(p, "InitTransactions", nil)
			defer patches2.Reset()

			kaMock.EXPECT().DoConsume(c).Return(msg, nil).AnyTimes()

			patches3 := ApplyPrivateMethod(taskTest, "packagingMessages",
				func([]*kafka.Message) ([]*kafka.Message, error) {
					return nil, errors.New("error")
				})
			defer patches3.Reset()

			err := taskTest.execute()
			So(err, ShouldNotBeNil)
		})

		Convey("indexer.BufLen() >= FlushBytes, flushMessage failed", func() {
			kaMock.EXPECT().NewConsumer(gomock.Any()).Return(c, nil)
			kaMock.EXPECT().NewTransactionalProducer(gomock.Any()).Return(p, nil)

			patches1 := ApplyMethodReturn(c, "Subscribe", nil)
			defer patches1.Reset()

			patches2 := ApplyMethodReturn(p, "InitTransactions", nil)
			defer patches2.Reset()

			kaMock.EXPECT().DoConsume(c).Return(msg, nil).AnyTimes()
			patches3 := ApplyPrivateMethod(taskTest, "packagingMessages",
				func([]*kafka.Message) ([]*kafka.Message, error) {
					return nil, nil
				})
			defer patches3.Reset()

			patches4 := ApplyPrivateMethod(taskTest, "flushMessages",
				func([]*kafka.Message, *kafka.Consumer, *interfaces.KafkaProducer) error {
					return errors.New("error")
				})
			defer patches4.Reset()

			err := taskTest.execute()
			So(err, ShouldNotBeNil)
		})

		Convey("time.Since(startTime) > FlushTime, flushMessage failed", func() {
			kaMock.EXPECT().NewConsumer(gomock.Any()).Return(c, nil)
			kaMock.EXPECT().NewTransactionalProducer(gomock.Any()).Return(p, nil)

			patches1 := ApplyMethodReturn(c, "Subscribe", nil)
			defer patches1.Reset()

			patches2 := ApplyMethodReturn(p, "InitTransactions", nil)
			defer patches2.Reset()

			FlushBytes = 100000
			cnt := 0
			kaMock.EXPECT().DoConsume(c).DoAndReturn(
				func(*kafka.Consumer) (*kafka.Message, error) {
					if cnt == 0 {
						cnt++
						return nil, errors.New("error")
					} else {
						return msg, nil
					}
				}).AnyTimes()

			patches4 := ApplyPrivateMethod(taskTest, "flushMessages",
				func([]*kafka.Message, *kafka.Consumer, *interfaces.KafkaProducer) error {
					return errors.New("error")
				})
			defer patches4.Reset()

			err := taskTest.execute()
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_Task_packagingMessages(t *testing.T) {
	Convey("Test Task packagingMessages", t, func() {

		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		kaMock := dmock.NewMockKafkaAccess(mockCtl)
		taskTest := NewTestTask(kaMock)

		Convey("msgs is empty", func() {
			msgs := []*kafka.Message{}
			_, err := taskTest.packagingMessages(msgs)
			So(err, ShouldBeNil)
		})

		Convey("packagingMessage failed", func() {
			msg := &kafka.Message{
				Timestamp: time.Now(),
				Value:     []byte("{\"__index_base\":\"1234567890abcdefg\"}"),
			}
			msgs := []*kafka.Message{msg}

			patches := ApplyPrivateMethod(taskTest, "packagingMessage",
				func(*Task, *kafka.Message) (*kafka.Message, error) {
					return nil, errors.New("error")
				})
			defer patches.Reset()

			_, err := taskTest.packagingMessages(msgs)
			So(err, ShouldNotBeNil)
		})

		Convey("success", func() {
			msg := &kafka.Message{
				Timestamp: time.Now(),
				Value:     []byte("{\"__index_base\":\"1234567890abcdefg\"}"),
			}
			msgs := []*kafka.Message{msg}

			patches := ApplyPrivateMethod(taskTest, "packagingMessage",
				func(*Task, *kafka.Message) (*kafka.Message, error) {
					return nil, nil
				})
			defer patches.Reset()

			_, err := taskTest.packagingMessages(msgs)
			So(err, ShouldBeNil)
		})
	})
}

type testStruct struct{}

func (t *testStruct) Pass(ctx context.Context, data *cond.OriginalData) (bool, error) {
	return true, nil
}

func Test_Task_packagingMessage(t *testing.T) {
	Convey("Test Task packagingMessage", t, func() {

		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		kaMock := dmock.NewMockKafkaAccess(mockCtl)
		taskTest := NewTestTask(kaMock)

		// condIf := &cond.condIf{
		// 	Name:      "field1",
		// 	Operation: cond.OperationEq,
		// 	ValueOptCfg: cond.ValueOptCfg{
		// 		ValueFrom: "const",
		// 		Value:     "test",
		// 	},
		// }

		var condIf cond.Condition = &testStruct{}

		Convey("sonic unmarshal error, but not return error", func() {
			msg := &kafka.Message{
				Timestamp: time.Now(),
				Value:     []byte("{\"__index_base\":\"1234567890abcdefg\""),
			}

			patches := ApplyFuncReturn(sonic.Unmarshal, errors.New("unmarshal error"))
			defer patches.Reset()

			_, err := taskTest.packagingMessage(msg)
			So(err, ShouldBeNil)
		})

		Convey("pass error", func() {
			msg := &kafka.Message{
				Timestamp: time.Now(),
				Value:     []byte("{\"__index_base\":\"1234567890abcdefg\""),
			}

			patches1 := ApplyFuncReturn(sonic.Unmarshal, nil)
			defer patches1.Reset()

			patches2 := ApplyFuncReturn(cond.NewCondition, condIf, nil)
			defer patches2.Reset()

			patches3 := ApplyMethodReturn(condIf, "Pass", false, errors.New("error"))
			defer patches3.Reset()

			_, err := taskTest.packagingMessage(msg)
			So(err, ShouldBeNil)
		})

		Convey("not pass", func() {
			msg := &kafka.Message{
				Timestamp: time.Now(),
				Value:     []byte("{\"__index_base\":\"1234567890abcdefg\""),
			}

			patches1 := ApplyFuncReturn(sonic.Unmarshal, nil)
			defer patches1.Reset()

			patches2 := ApplyFuncReturn(cond.NewCondition, condIf, nil)
			defer patches2.Reset()

			patches3 := ApplyMethodReturn(condIf, "Pass", false, nil)
			defer patches3.Reset()

			_, err := taskTest.packagingMessage(msg)
			So(err, ShouldBeNil)
		})

		Convey("pickdata error", func() {
			msg := &kafka.Message{
				Timestamp: time.Now(),
				Value:     []byte("{\"__index_base\":\"1234567890abcdefg\""),
			}
			taskTest.FieldScope = interfaces.CUSTOM

			patches1 := ApplyFuncReturn(sonic.Unmarshal, nil)
			defer patches1.Reset()

			patches2 := ApplyFuncReturn(cond.NewCondition, condIf, nil)
			defer patches2.Reset()

			patches3 := ApplyMethodReturn(condIf, "Pass", true, nil)
			defer patches3.Reset()

			patches4 := ApplyPrivateMethod(taskTest, "pickData",
				func(*Task, map[string]any, map[string]any) error {
					return errors.New("error")
				})
			defer patches4.Reset()

			_, err := taskTest.packagingMessage(msg)
			So(err, ShouldNotBeNil)
		})

		Convey("sonic marshal error", func() {
			msg := &kafka.Message{
				Timestamp: time.Now(),
				Value:     []byte("{\"__index_base\":\"1234567890abcdefg\""),
			}

			patches1 := ApplyFuncReturn(sonic.Unmarshal, nil)
			defer patches1.Reset()

			patches2 := ApplyFuncReturn(cond.NewCondition, condIf, nil)
			defer patches2.Reset()

			patches3 := ApplyMethodReturn(condIf, "Pass", true, nil)
			defer patches3.Reset()

			patches4 := ApplyPrivateMethod(taskTest, "pickData",
				func(*Task, map[string]any, map[string]any) error {
					return nil
				})
			defer patches4.Reset()

			patches5 := ApplyFuncReturn(sonic.Marshal, nil, errors.New("error"))
			defer patches5.Reset()

			_, err := taskTest.packagingMessage(msg)
			So(err, ShouldNotBeNil)
		})

		Convey("success", func() {
			msg := &kafka.Message{
				Timestamp: time.Now(),
				Value:     []byte("{\"__index_base\":\"1234567890abcdefg\"}"),
			}
			taskTest.FieldScope = interfaces.ALL

			patches1 := ApplyFuncReturn(sonic.Unmarshal, nil)
			defer patches1.Reset()

			patches2 := ApplyFuncReturn(cond.NewCondition, condIf, nil)
			defer patches2.Reset()

			patches3 := ApplyMethodReturn(condIf, "Pass", true, nil)
			defer patches3.Reset()

			patches4 := ApplyPrivateMethod(taskTest, "pickData",
				func(*Task, map[string]any, map[string]any) error {
					return nil
				})
			defer patches4.Reset()

			patches5 := ApplyFuncReturn(sonic.Marshal, nil, nil)
			defer patches5.Reset()

			_, err := taskTest.packagingMessage(msg)
			So(err, ShouldBeNil)
		})
	})
}

func Test_Task_flushMessages(t *testing.T) {
	Convey("Test Task flushMessage", t, func() {
		FailureThreshold = 2

		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		kaMock := dmock.NewMockKafkaAccess(mockCtl)
		taskTest := NewTestTask(kaMock)

		c, _ := kafka.NewConsumer(&kafka.ConfigMap{"bootstrap.servers": "localhost"})
		p, _ := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost"})
		kp := &interfaces.KafkaProducer{
			Producer:     p,
			DeliveryChan: make(chan kafka.Event, 10),
		}

		msg := &kafka.Message{
			Timestamp: time.Now(),
			Value:     []byte("{\"__index_base\":\"1234567890abcdefg\"}"),
		}
		msgs := []*kafka.Message{msg}

		Convey("kAccess.DoProduce failed", func() {
			kaMock.EXPECT().DoProduce(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("error")).AnyTimes()

			err := taskTest.flushMessages(msgs, c, kp)
			So(err, ShouldNotBeNil)
		})

		Convey("success", func() {
			kaMock.EXPECT().DoProduce(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

			err := taskTest.flushMessages(msgs, c, kp)
			So(err, ShouldBeNil)
		})
	})
}

func Test_Task_pickData(t *testing.T) {
	Convey("Test Task pickData", t, func() {

		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		kaMock := dmock.NewMockKafkaAccess(mockCtl)
		taskTest := NewTestTask(kaMock)
		taskTest.Fields = []*cond.Field{
			{
				Name: "name",
				Type: "keyword",
				Path: []string{"name"},
			},
			{
				Name: "age",
				Type: "keyword",
				Path: []string{"age"},
			},
		}
		origin := make(map[string]any)
		pick := make(map[string]any)

		Convey("get data error", func() {
			patches := ApplyFuncReturn(getData, nil, false, errors.New("error"))
			defer patches.Reset()

			err := taskTest.pickData(origin, pick)
			So(err, ShouldNotBeNil)
		})

		Convey("set data error", func() {
			patches1 := ApplyFuncReturn(getData, nil, false, nil)
			defer patches1.Reset()

			patches2 := ApplyFuncReturn(setData, errors.New("error"))
			defer patches2.Reset()

			err := taskTest.pickData(origin, pick)
			So(err, ShouldNotBeNil)
		})

		Convey("success", func() {
			patches1 := ApplyFuncReturn(getData, nil, false, nil)
			defer patches1.Reset()

			patches2 := ApplyFuncReturn(setData, nil)
			defer patches2.Reset()

			err := taskTest.pickData(origin, pick)
			So(err, ShouldBeNil)
		})
	})
}

func Test_Task_getData(t *testing.T) {
	Convey("getData", t, func() {
		field := &cond.Field{
			Name: "age",
			Type: "keyword",
			Path: []string{"age"},
		}
		origin := make(map[string]any)
		Convey("GetDatasByPath error", func() {
			patches := ApplyFuncReturn(GetDatasByPath, nil, false, errors.New("error"))
			defer patches.Reset()

			_, _, err := getData(origin, field)
			So(err, ShouldNotBeNil)
		})

		Convey("datas is null", func() {
			patches := ApplyFuncReturn(GetDatasByPath, []any{}, false, nil)
			defer patches.Reset()

			_, _, err := getData(origin, field)
			So(err, ShouldBeNil)
		})

		Convey("success", func() {
			patches := ApplyFuncReturn(GetDatasByPath, []any{1}, false, nil)
			defer patches.Reset()

			_, _, err := getData(origin, field)
			So(err, ShouldBeNil)
		})
	})
}

func Test_Task_GetDatasByPath(t *testing.T) {
	Convey("GetDatasByPath", t, func() {

		Convey("GetDatasByPath: simple json", func() {
			var jsonStr = `{
				"a":{
					"b":{
						"c": {
							"d": "d"
						}
					}
				}
			}`

			root := map[string]any{}
			_ = sonic.UnmarshalString(jsonStr, &root)
			path := []string{"a", "b", "c", "d"}

			dest, _, err := GetDatasByPath(root, path)
			So(err, ShouldBeNil)
			So(len(dest), ShouldEqual, 1)
		})

		Convey("GetDatasByPath: invalid json", func() {
			var jsonStr = `{
				"a":{
					"b":{
						"c": null
					}
				}
			}`

			root := map[string]any{}
			_ = sonic.UnmarshalString(jsonStr, &root)
			path := []string{"a", "b", "c", "d"}

			dest, _, err := GetDatasByPath(root, path)
			So(err, ShouldBeNil)
			So(len(dest), ShouldEqual, 0)
		})

		Convey("getNodeByPath: none value", func() {
			var jsonStr = `{
				"a":{
					"b":{
						"c": "c"
					}
				}
			}`

			root := map[string]any{}
			_ = sonic.UnmarshalString(jsonStr, &root)
			path := []string{"a", "b", "c", "d"}

			dest, _, err := GetDatasByPath(root, path)
			So(err, ShouldBeNil)
			So(len(dest), ShouldEqual, 0)
		})

		Convey("getNodeByPath: complex json", func() {
			var jsonStr = `{
				"a":[
					{
						"b":[
							{
								"c":[
									{
										"d": "d1"
									},{
										"d": ["d2"]
									}
								]
							},
							{
								"c":[
									{
										"d": ["d3", "d4"]
									}
								]
							}
						]
					}, {
						"b":[
							{
								"c":[
									{
										"d": [["d5"]]
									},{
										"d": [[], ["d6"]]
									}
								]
							},
							{
								"c":[
									{
										"d": ["d7", ["d8"]]
									},
									{
										"d": [["d9"], ["d10", "d11"]]
									}
								]
							}
						]
					}
				]
			}`

			root := map[string]any{}
			_ = sonic.UnmarshalString(jsonStr, &root)
			path := []string{"a", "b", "c", "d"}

			dest, _, err := GetDatasByPath(root, path)
			So(err, ShouldBeNil)
			So(len(dest), ShouldBeGreaterThan, 0)
		})
	})
}

func Test_Task_GetLastDatas(t *testing.T) {
	Convey("Test Task GetLastDatas", t, func() {
		Convey("obj is nil", func() {
			var obj any = nil
			_, _, err := GetLastDatas(obj)
			So(err, ShouldBeNil)
		})

		Convey("obj is slice", func() {
			obj := []any{1, 2}
			_, isSliceValue, err := GetLastDatas(obj)
			So(isSliceValue, ShouldBeTrue)
			So(err, ShouldBeNil)
		})

		Convey("obj is not slice", func() {
			obj := 1
			_, isSliceValue, err := GetLastDatas(obj)
			So(isSliceValue, ShouldBeFalse)
			So(err, ShouldBeNil)
		})
	})
}

func Test_Task_setData(t *testing.T) {
	Convey("Test Task setData", t, func() {
		Convey("data length is 0", func() {
			field := &cond.Field{}
			obj := map[string]any{}
			data := []any{}
			isSliceValue := false
			err := setData(field, obj, data, isSliceValue)
			So(err, ShouldBeNil)
		})

		Convey("set a.b.c", func() {
			field := &cond.Field{
				Name: "a.b.c",
			}
			obj := map[string]any{}
			data := []any{"wahaha"}
			isSliceValue := false
			err := setData(field, obj, data, isSliceValue)
			expectedObj := []byte(`{"a": {"b": {"c": "wahaha"}}}`)
			var expected map[string]any
			if err := sonic.Unmarshal(expectedObj, &expected); err != nil {
				t.Fatal(err.Error())
			}
			So(obj, ShouldResemble, expected)
			So(err, ShouldBeNil)
		})

		Convey("set a.b.c, value is an array", func() {
			field := &cond.Field{
				Name: "a.b.c",
			}
			obj := map[string]any{}
			data := []any{"wahaha"}
			isSliceValue := true
			err := setData(field, obj, data, isSliceValue)
			expectedObj := []byte(`{"a": {"b": {"c": ["wahaha"]}}}`)
			var expected map[string]any
			if err := sonic.Unmarshal(expectedObj, &expected); err != nil {
				t.Fatal(err.Error())
			}
			So(obj, ShouldResemble, expected)
			So(err, ShouldBeNil)
		})
	})
}
