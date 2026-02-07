// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package event

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"
	"uniquery/common"
	"uniquery/interfaces"
	imock "uniquery/interfaces/mock"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/golang/mock/gomock"
	did "github.com/kweaver-ai/kweaver-go-lib/did"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	testCtx = context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage)
	testNow = time.Now().UnixMilli()

	eventModel = interfaces.EventModel{
		EventModelID:             "1",
		EventModelName:           "测试中的名称",
		EventModelType:           "atomic",
		UpdateTime:               testNow,
		EventModelTags:           []string{"xx1", "xx2"},
		DataSourceType:           "data_view",
		DataSource:               []string{"1"},
		DataSourceName:           []string{},
		DataSourceGroupName:      []string{},
		DownstreamDependentModel: []string{"1"},
		DetectRule: interfaces.DetectRule{
			DetectRuleId: "",
			Priority:     0,
			Type:         "",
			Formula:      []interfaces.FormulaItem{},
		},
		AggregateRule: interfaces.AggregateRule{
			GroupFields: []string{},
		},
		DefaultTimeWindow: interfaces.TimeInterval{
			Interval: 5,
			Unit:     "m",
		},
		IsActive: 1,
		IsCustom: 1,
		Task:     eventTask,
	}
	eventTask = interfaces.EventTask{
		TaskID:   "1",
		ModelID:  "1",
		Schedule: interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
		StorageConfig: interfaces.StorageConfig{
			IndexBase:    "base1",
			DataViewName: "view1",
			DataViewId:   "1",
		},
		DispatchConfig: interfaces.DispatchConfig{
			TimeOut:        3600,
			RouteStrategy:  "FIRST",
			BlockStrategy:  "",
			FailRetryCount: 3,
		},
		ExecuteParameter:   map[string]any{},
		TaskStatus:         4,
		StatusUpdateTime:   testNow,
		ErrorDetails:       "",
		ScheduleSyncStatus: 1,
		UpdateTime:         testNow,
	}

	formula = []interfaces.FormulaItem{
		{
			Level: 1,
			Filter: interfaces.LogicFilter{
				LogicOperator: "",
				FilterExpress: interfaces.FilterExpress{
					Name:      "value",
					Value:     []interface{}{float64(1), float64(4)},
					Operation: "range",
				},
			},
		},
		{
			Level: 2,
			Filter: interfaces.LogicFilter{
				LogicOperator: "and",
				FilterExpress: interfaces.FilterExpress{
					Name:      "value",
					Value:     []interface{}{float64(1), float64(4)},
					Operation: "range",
				},
				Children: []interfaces.LogicFilter{
					{
						LogicOperator: "",
						FilterExpress: interfaces.FilterExpress{
							Name:      "value",
							Value:     float64(3),
							Operation: "!=",
						},
					},
				},
			},
		},
		{
			Level: 3,
			Filter: interfaces.LogicFilter{
				LogicOperator: "or",
				FilterExpress: interfaces.FilterExpress{
					Name:      "value",
					Value:     []interface{}{float64(1), float64(4)},
					Operation: "range",
				},
				Children: []interfaces.LogicFilter{
					{
						LogicOperator: "",
						FilterExpress: interfaces.FilterExpress{
							Name:      "value",
							Value:     []interface{}{float64(0), float64(1)},
							Operation: "range",
						},
					},
				},
			},
		},
	}
	record = map[string]any{
		"value":          1.33,
		"@timestamp":     1712108595085,
		"labels.host_ip": "localhost",
	}
	sr = interfaces.SourceRecords{
		Records: interfaces.Records{
			record,
		},
	}

	event = interfaces.AtomicEvent{
		BaseEvent: interfaces.BaseEvent{
			Id:                  "1",
			Title:               eventModel.EventModelName + "_" + interfaces.EVENT_MODEL_LEVEL_ZH_CN[1],
			EventModelId:        eventModel.EventModelID,
			EventModelName:      "测试中的名称",
			Level:               1,
			Tags:                eventModel.EventModelTags,
			CreateTime:          testNow,
			EventType:           eventModel.EventModelType,
			DataSource:          eventModel.DataSource,
			DataSourceType:      eventModel.DataSourceType,
			DataSourceName:      eventModel.DataSourceName,
			DataSourceGroupName: eventModel.DataSourceGroupName,
			DefaultTimeWindow:   eventModel.DefaultTimeWindow,
			Schedule:            eventModel.Task.Schedule,
			Labels:              map[string]string{"labels.host_ip": "localhost"},
		},
		Message: "监控对象(localhost)的监控项([])产生了异常(当前值为'1.33')",
	}
)

func MockNewEventService(appSetting *common.AppSetting, engine interfaces.EventEngine,
	emAcces interfaces.EventModelAccess, dvAccess interfaces.DataViewAccess,
	ibAccess interfaces.IndexBaseAccess, ps interfaces.PermissionService) *eventService {

	return &eventService{
		appSetting: appSetting,
		engine:     nil,
		emAccess:   emAcces,
		dvAccess:   dvAccess,
		// uAccess:    uAcess,
		ibAccess: ibAccess,
		ps:       ps,
	}

	// return ess
}

func MockNewEventEngine(
	em interfaces.EventModel,
	dataSource any,
	dataSourceType string,
	dvAccess interfaces.DataViewAccess,
	kAccess interfaces.KafkaAccess,
	// uAccess interfaces.UniqueryAccess,
	appSetting *common.AppSetting,
) *EventEngine {
	return &EventEngine{
		appSetting:     appSetting,
		EventModel:     em,
		DataSource:     dataSource,
		DataSourceType: dataSourceType,
		DVAccess:       dvAccess,
		KAccess:        kAccess,
		// UAcess:         uAccess,
	}
}

func TestComposeEventModelByQueryParam(t *testing.T) {
	Convey("Test ComposeEventModelByQueryParam", t, func() {
		event_query := interfaces.EventQuery{
			QueryType: "instant_query",
			Id:        "1",
			Start:     1,
			End:       2,
		}
		em := ComposeEventModelByQueryParam(event_query)
		So(em, ShouldNotBeNil)
	})

}
func TestApply(t *testing.T) {
	// Convey("TestApply", t, func() {
	// 	mockCtrl := gomock.NewController(t)
	// 	defer mockCtrl.Finish()

	// 	dvaMock := mock.NewMockDataViewAccess(mockCtrl)
	// 	em := interfaces.EventModel{}
	// 	dataSource := []string{"5111313212312122"}
	// 	dataSourceType := "metric_model"

	// 	engine := MockNewEventEngine(em, dataSource, dataSourceType, dvaMock)
	// 	Convey("Success", func() {

	// 	})

	// })

}

func TestCall(t *testing.T) {
	Convey("TestApply", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		// uaMock := imock.NewMockUniqueryAccess(mockCtrl)
		dvaMock := imock.NewMockDataViewAccess(mockCtrl)
		dataSource := []string{"1"}
		dataSourceType := "metric_model"
		kAcess := imock.NewMockKafkaAccess(mockCtrl)

		engine := MockNewEventEngine(eventModel, dataSource, dataSourceType, dvaMock, kAcess, appSetting)
		Convey("Success", func() {

			hit, _, _ := engine.Call(record, eventModel, formula)
			So(hit, ShouldBeTrue)
		})
	})
}

func TestTraversal(t *testing.T) {
	Convey("TestTraversal", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		appSetting := &common.AppSetting{}
		// uaMock := imock.NewMockUniqueryAccess(mockCtrl)
		dvaMock := imock.NewMockDataViewAccess(mockCtrl)
		dataSource := []string{"1"}
		dataSourceType := "metric_model"
		kAcess := imock.NewMockKafkaAccess(mockCtrl)
		engine := MockNewEventEngine(eventModel, dataSource, dataSourceType, dvaMock, kAcess, appSetting)
		Convey("children is nil", func() {

			hit, _ := engine.Traversal(formula[0].Filter, record)
			So(hit, ShouldBeTrue)
		})
		Convey("and", func() {

			hit, _ := engine.Traversal(formula[1].Filter, record)
			So(hit, ShouldBeTrue)
		})
		Convey("or", func() {

			hit, _ := engine.Traversal(formula[2].Filter, record)
			So(hit, ShouldBeFalse)
		})
	})
}

func TestGenerationAtomicEvent(t *testing.T) {
	Convey("TestGenerationAtomicEvent", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		// uaMock := imock.NewMockUniqueryAccess(mockCtrl)
		dvaMock := imock.NewMockDataViewAccess(mockCtrl)
		dataSource := []string{"1"}
		dataSourceType := "metric_model"
		kAcess := imock.NewMockKafkaAccess(mockCtrl)
		engine := MockNewEventEngine(eventModel, dataSource, dataSourceType, dvaMock, kAcess, appSetting)

		Convey("success", func() {
			patch := ApplyFunc(did.GenerateDistributedID,
				func() (id uint64, err error) {
					return 1, nil
				},
			)
			defer patch.Reset()

			expectedEvent := interfaces.AtomicEvent{
				BaseEvent: interfaces.BaseEvent{
					Id:                  "1",
					Title:               eventModel.EventModelName + "_" + interfaces.EVENT_MODEL_LEVEL_ZH_CN[1],
					EventModelId:        eventModel.EventModelID,
					EventModelName:      "测试中的名称",
					Level:               1,
					Tags:                eventModel.EventModelTags,
					CreateTime:          testNow,
					EventType:           eventModel.EventModelType,
					DataSource:          eventModel.DataSource,
					DataSourceType:      eventModel.DataSourceType,
					DataSourceName:      eventModel.DataSourceName,
					DataSourceGroupName: eventModel.DataSourceGroupName,
					DefaultTimeWindow:   eventModel.DefaultTimeWindow,
					Schedule:            eventModel.Task.Schedule,
					Labels:              map[string]string{"labels.host_ip": "localhost"},
				},
				Message: "监控对象(localhost)的监控项(value)产生了异常('value':'1.33')",
			}

			event, err := engine.GenerationAtomicEvent(testCtx, sr, eventModel, formula[0], map[string]any{"value": 1.33}, record)
			So(err, ShouldBeNil)
			So(event.Message, ShouldEqual, expectedEvent.Message)

		})
	})
}
func TestGenerationAggregateEvent(t *testing.T) {
	Convey("Test GenerationAggregateEvent", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		// uaMock := imock.NewMockUniqueryAccess(mockCtrl)
		dvaMock := imock.NewMockDataViewAccess(mockCtrl)
		dataSource := []string{"1"}
		dataSourceType := "metric_model"
		kAcess := imock.NewMockKafkaAccess(mockCtrl)
		engine := MockNewEventEngine(eventModel, dataSource, dataSourceType, dvaMock, kAcess, appSetting)

		Convey("success", func() {
			patch := ApplyFunc(did.GenerateDistributedID,
				func() (id uint64, err error) {
					return 1, nil
				},
			)
			defer patch.Reset()

			expectedEvent := interfaces.AggregateEvent{
				BaseEvent: interfaces.BaseEvent{
					Id:                  "1",
					Title:               eventModel.EventModelName + "_" + interfaces.EVENT_MODEL_LEVEL_ZH_CN[1],
					EventModelId:        eventModel.EventModelID,
					EventModelName:      "测试中的名称",
					Level:               1,
					Tags:                eventModel.EventModelTags,
					CreateTime:          testNow,
					EventType:           eventModel.EventModelType,
					DataSource:          eventModel.DataSource,
					DataSourceType:      eventModel.DataSourceType,
					DataSourceName:      eventModel.DataSourceName,
					DataSourceGroupName: eventModel.DataSourceGroupName,
					DefaultTimeWindow:   eventModel.DefaultTimeWindow,
					Schedule:            eventModel.Task.Schedule,
					Labels:              map[string]string{"labels.host_ip": "localhost"},
				},
				Message: "基于,生成了一个等级为的聚合事件",
			}

			eventModel := interfaces.EventModel{
				EventModelID:        "1",
				EventModelName:      "测试中的名称",
				EventModelType:      "aggregate",
				UpdateTime:          testNow,
				EventModelTags:      []string{"xx1", "xx2"},
				DataSourceType:      "metric_model",
				DataSource:          []string{"1"},
				DataSourceName:      []string{},
				DataSourceGroupName: []string{},
				DetectRule:          interfaces.DetectRule{},
				AggregateRule:       interfaces.AggregateRule{Type: "healthy_compute", GroupFields: []string{}},
				DefaultTimeWindow:   interfaces.TimeInterval{Interval: 5, Unit: "m"},
				IsActive:            1,
				IsCustom:            1,
				Task:                eventTask,
			}

			event, err := engine.GenerationAggregateEvent(testCtx, sr, eventModel, interfaces.EventContext{}, []string{})
			So(err, ShouldBeNil)
			So(event.Message, ShouldEqual, expectedEvent.Message)

		})
	})
}

func Test_eventDataTransferToMessage(t *testing.T) {
	Convey("TestEventDataTransferToMessage", t, func() {

		Convey("success", func() {
			messages := make([]*kafka.Message, 0)
			topic := "default.mdl.view"
			err := eventDataTransferToMessage([]interfaces.IEvent{event}, eventModel.Task, &messages, topic)
			So(err, ShouldBeNil)
			So(len(messages), ShouldEqual, 1)

		})
	})
}

func TestEventEngine_Judge(t *testing.T) {
	Convey("Test Flatten", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		// uaMock := imock.NewMockUniqueryAccess(mockCtrl)
		dvaMock := imock.NewMockDataViewAccess(mockCtrl)
		dataSource := []string{"1"}
		dataSourceType := "metric_model"
		kAcess := imock.NewMockKafkaAccess(mockCtrl)
		engine := MockNewEventEngine(eventModel, dataSource, dataSourceType, dvaMock, kAcess, appSetting)
		sr := interfaces.SourceRecords{
			Records: interfaces.Records{
				{"a": 1, "v": 2},
				{"a": 2, "v": 4}},
		}

		Convey("success, value is {}", func() {
			patch := ApplyFunc(did.GenerateDistributedID,
				func() (id uint64, err error) {
					return 1, nil
				},
			)
			defer patch.Reset()

			// engine.EXPECT().Call(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(interfaces.IEvents{}, 0, nil)
			// engine.EXPECT().GetLastEventLevel(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(1)
			_, _, err := engine.Judge(testCtx, sr, eventModel)
			So(err, ShouldBeNil)
		})
	})
}

func Test_eventService_QuerySingleEventByEventId(t *testing.T) {
	Convey("Test QuerySingleEventByEventId", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			PoolSetting: common.PoolSetting{
				ViewPoolSize: 10,
			}}
		engine := imock.NewMockEventEngine(mockCtrl)
		emAccess := imock.NewMockEventModelAccess(mockCtrl)
		dvAccess := imock.NewMockDataViewAccess(mockCtrl)
		// uAcess := imock.NewMockUniqueryAccess(mockCtrl)
		ibaMock := imock.NewMockIndexBaseAccess(mockCtrl)
		psMock := imock.NewMockPermissionService(mockCtrl)
		es := MockNewEventService(appSetting, engine, emAccess, dvAccess, ibaMock, psMock)

		query := interfaces.EventDetailsQueryReq{
			EventModelID: "1",
			EventID:      "1",
			Start:        1011,
			End:          1013,
		}
		event_query := interfaces.EventQuery{
			QueryType: "instant_query",
			Id:        "1",
			Start:     1,
			End:       2,
		}

		Convey("success", func() {
			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dmq := NewLogDataQuery(appSetting, eventModel.DataSource, eventModel.DataSourceType, eventModel.DefaultTimeWindow, event_query)

			events := interfaces.IEvents{}
			entities := interfaces.Records{}
			sr := interfaces.SourceRecords{
				Records: interfaces.Records{{
					"event_type": "atomic", "default_time_window_str": "{\"interval\":5,\"unit\":\"m\"}",
					"labels_str": "{\"A\":\"V\"}", "message": "", "event_message": "",
					"context_str": "{}", "context": interfaces.EventContext{},
					"trigger_time": time.Now(), "trigger_data_str": "[]", "trigger_data": interfaces.Records{},
					"schedule_str": "{\"type\":\"FIX_RATE\",\"expression\":\"5m\"}", "aggregate_type": "", "detect_type": "", "aggregate_algo": "", "detect_algo": "",
				}},
			}
			var event interfaces.IEvent
			emAccess.EXPECT().GetEventModelById(gomock.Any(), gomock.Any()).AnyTimes().Return([]interfaces.EventModel{eventModel}, nil)
			engine.EXPECT().Apply(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(events, entities, 0)
			engine.EXPECT().QuerySingleEventByEventId(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(event, nil)
			// dvAccess.EXPECT().GetDataViewIDByName(gomock.Any(), gomock.Any()).AnyTimes().Return("1", nil)

			// dmq.EXPECT().FetchSourceRecordsFrom(gomock.Any()).AnyTimes().Return(sr, nil, nil)
			patches := ApplyPrivateMethod(es, "InitEngine",
				func(ctx context.Context, em interfaces.EventModel) error {
					es.engine = engine
					return nil
				})
			defer patches.Reset()

			patch2 := ApplyMethodReturn(dmq, "FetchSourceRecordsFrom", sr, interfaces.Record{}, nil)
			defer patch2.Reset()

			_, err := es.QuerySingleEventByEventId(testCtx, query)
			So(err, ShouldBeNil)
		})

		Convey("failed,caused by GetEventModelById failed \n ", func() {

			var event interfaces.IEvent
			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			emAccess.EXPECT().GetEventModelById(gomock.Any(), gomock.Any()).AnyTimes().Return([]interfaces.EventModel{}, errors.New("some error"))
			engine.EXPECT().QuerySingleEventByEventId(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(event, nil)
			patches := ApplyPrivateMethod(es, "InitEngine",
				func(ctx context.Context, em interfaces.EventModel) error {
					es.engine = engine
					return nil
				})
			defer patches.Reset()
			_, err := es.QuerySingleEventByEventId(testCtx, query)
			So(err, ShouldNotBeNil)
		})
		Convey("failed,caused by len(eventModels)==0 \n ", func() {
			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			emAccess.EXPECT().GetEventModelById(gomock.Any(), gomock.Any()).AnyTimes().Return([]interfaces.EventModel{}, nil)
			engine.EXPECT().QuerySingleEventByEventId(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(event, nil)
			patches := ApplyPrivateMethod(es, "InitEngine",
				func(ctx context.Context, em interfaces.EventModel) error {
					es.engine = engine
					return nil
				})
			defer patches.Reset()
			_, err := es.QuerySingleEventByEventId(testCtx, query)
			So(err, ShouldNotBeNil)
		})

		Convey("failed,caused by len(records) == 0 \n ", func() {
			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			emAccess.EXPECT().GetEventModelById(gomock.Any(), gomock.Any()).AnyTimes().Return([]interfaces.EventModel{eventModel}, nil)
			// dvAccess.EXPECT().GetDataViewIDByName(gomock.Any(), gomock.Any()).Return("1", nil)
			// var event interfaces.IEvent
			// engine.EXPECT().QuerySingleEventByEventId(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(event, nil)
			// patches := ApplyPrivateMethod(es, "InitEngine",
			// 	func(ctx context.Context, em interfaces.EventModel) error {
			// 		es.engine = engine
			// 		return nil
			// 	})
			// defer patches.Reset()
			dmq := NewLogDataQuery(appSetting, eventModel.DataSource, eventModel.DataSourceType, eventModel.DefaultTimeWindow, event_query)
			patch2 := ApplyMethodReturn(dmq, "FetchSourceRecordsFrom",
				interfaces.SourceRecords{}, interfaces.Record{}, nil)
			defer patch2.Reset()

			_, err := es.QuerySingleEventByEventId(testCtx, query)
			So(err, ShouldNotBeNil)
		})

		Convey("failed,caused by FetchSourceRecordsFrom failed \n ", func() {
			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			emAccess.EXPECT().GetEventModelById(gomock.Any(), gomock.Any()).AnyTimes().Return([]interfaces.EventModel{eventModel}, nil)
			// dvAccess.EXPECT().GetDataViewIDByName(gomock.Any(), gomock.Any()).Return("1", nil)
			// var event interfaces.IEvent
			// engine.EXPECT().QuerySingleEventByEventId(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(event, nil)
			// patches := ApplyPrivateMethod(es, "InitEngine",
			// 	func(ctx context.Context, em interfaces.EventModel) error {
			// 		es.engine = engine
			// 		return nil
			// 	})
			// defer patches.Reset()

			dmq := NewLogDataQuery(appSetting, eventModel.DataSource, eventModel.DataSourceType, eventModel.DefaultTimeWindow, event_query)
			patch2 := ApplyMethodReturn(dmq, "FetchSourceRecordsFrom",
				interfaces.SourceRecords{}, interfaces.Record{}, errors.New("some error"))
			defer patch2.Reset()

			_, err := es.QuerySingleEventByEventId(testCtx, query)
			So(err, ShouldNotBeNil)
		})

	})
}

func Test_eventService_Query(t *testing.T) {
	Convey("Test Query", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		// emAccess:    dmock.NewMockEventModelAccess(mockCtrl),
		// kafkaAccess: dmock.NewMockKafkaAccess(mockCtrl),
		// topics:      []string{"default.mdl.view"},
		// engine:      dmock.NewMockEventEngine(mockCtrl),

		engine := imock.NewMockEventEngine(mockCtrl)
		emAccess := imock.NewMockEventModelAccess(mockCtrl)
		dvAccess := imock.NewMockDataViewAccess(mockCtrl)
		// uAcess := imock.NewMockUniqueryAccess(mockCtrl)
		ibaMock := imock.NewMockIndexBaseAccess(mockCtrl)
		psMock := imock.NewMockPermissionService(mockCtrl)
		es := MockNewEventService(&common.AppSetting{}, engine, emAccess, dvAccess, ibaMock, psMock)
		// es.engine = engine
		query := interfaces.EventQueryReq{
			Querys: []interfaces.EventQuery{
				{QueryType: "instant_query", Id: "1"},
			},
		}
		ops := []interfaces.ResourceOps{
			{
				ResourceID: interfaces.RESOURCE_ID_ALL,
				Operations: []string{interfaces.OPERATION_TYPE_CREATE},
			},
		}
		events := interfaces.IEvents{}
		entities := interfaces.Records{}
		psMock.EXPECT().GetResourcesOperations(gomock.Any(), gomock.Any(), gomock.Any()).Return(ops, nil)
		emAccess.EXPECT().GetEventModelById(gomock.Any(), gomock.Any()).AnyTimes().Return([]interfaces.EventModel{eventModel}, nil)
		engine.EXPECT().Apply(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(events, entities, 0)
		patches := ApplyPrivateMethod(es, "InitEngine",
			func(ctx context.Context, em interfaces.EventModel) error {
				es.engine = engine
				return nil
			})

		defer patches.Reset()
		// ess.Subscribe(exitCh)
		_, _, _, err := es.Query(testCtx, query)
		So(err, ShouldBeNil)

	})
}

// func TestEventEngineSendKafka(t *testing.T) {
// 	Convey("Test SendKafka", t, func() {

// 		mockCtrl := gomock.NewController(t)
// 		defer mockCtrl.Finish()

// 		appSetting := &common.AppSetting{}
// 		dvaMock := imock.NewMockDataViewAccess(mockCtrl)
// 		dataSource := []string{"1"}
// 		dataSourceType := "metric_model"
// 		kAcess := imock.NewMockKafkaAccess(mockCtrl)
// 		engine := MockNewEventEngine(eventModel, dataSource, dataSourceType, dvaMock, kAcess, uaMock, appSetting)

// 		// producer, _ := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost"})
// 		producer := &kafka.Producer{}
// 		patchF2 := ApplyMethodReturn(producer, "Close")
// 		defer patchF2.Reset()

// 		records := kafka.Message{
// 			Value: []byte(`{"a": {"b": {"c": "wahaha"}}}`),
// 			Headers: []kafka.Header{
// 				{Key: "__id", Value: []byte(strconv.FormatUint(event.GetBaseEvent().EventModelId, 10))},
// 				{Key: "__type", Value: []byte("event_model")},
// 			},
// 		}
// 		task := interfaces.EventTask{
// 			TaskID: 1,
// 			StorageConfig: interfaces.StorageConfig{
// 				IndexBase: "inner_default",
// 			},
// 		}
// 		Convey("failed, NewTrxProducer failed \n", func() {
// 			kAcess.EXPECT().NewTrxProducer(gomock.Any()).AnyTimes().Return(nil, errors.New("some error"))

// 			err := engine.SendKafka(testCtx,task, []*kafka.Message{&records}, "default.mdl.dip_event_model_data")
// 			So(err, ShouldNotBeNil)
// 		})
// 		Convey("success", func() {
// 			kAcess.EXPECT().NewTrxProducer(gomock.Any()).AnyTimes().Return(producer, nil)
// 			kAcess.EXPECT().DoProduce(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
// 			err := engine.SendKafka(testCtx,task, []*kafka.Message{&records}, "default.mdl.dip_event_model_data")
// 			So(err, ShouldBeNil)
// 		})

// 	})
// }

func TestEventEngine_mit(t *testing.T) {
	Convey("Test mit", t, func() {
		// sr := interfaces.SourceRecords{
		// 	Records: interfaces.Records{
		// 		{"a": 1, "v": 2},
		// 		{"a": 2, "v": 4}},
		// }

		// sr := interfaces.SourceRecords{
		// 	Records: interfaces.Records{{
		// 		"event_type": "atomic", "default_time_window_str": "{\"interval\":5,\"unit\":\"m\"}",
		// 		"labels_str": "{\"A\":\"V\"}", "message": "", "event_message": "",
		// 		"context_str": "{}", "context": interfaces.EventContext{},
		// 		"trigger_time": time.Now(), "trigger_data_str": "[]", "trigger_data": interfaces.Records{},
		// 		"schedule_str": "{\"type\":\"FIX_RATE\",\"expression\":\"5m\"}", "aggregate_type": "", "detect_type": "", "aggregate_algo": "", "detect_algo": "",
		// 	}},
		// }
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		// uaMock := imock.NewMockUniqueryAccess(mockCtrl)
		dvaMock := imock.NewMockDataViewAccess(mockCtrl)
		dataSource := []string{"1"}
		dataSourceType := "metric_model"
		kAcess := imock.NewMockKafkaAccess(mockCtrl)
		engine := MockNewEventEngine(eventModel, dataSource, dataSourceType, dvaMock, kAcess, appSetting)
		events := interfaces.AtomicEvent{
			BaseEvent: interfaces.BaseEvent{
				Id:                  "1",
				Title:               "hello",
				EventModelId:        "1",
				EventModelName:      "测试中的名称",
				EventType:           "atomic",
				Level:               1,
				Tags:                []string{"A", "B", "C"},
				GenerateType:        "batch",
				CreateTime:          testNow,
				DataSourceType:      "data_view",
				DataSource:          []string{"1"},
				DataSourceName:      []string{},
				DataSourceGroupName: []string{},
				DefaultTimeWindow:   interfaces.TimeInterval{Interval: 5, Unit: "m"},
				Schedule:            interfaces.Schedule{Type: "FIX_RATE", Expression: "5m"},
				Labels:              map[string]string{"A": "a", "B": "b"},
				Relations:           map[string]any{},
			},
			TriggerTime: testNow,
			TriggerData: interfaces.Records{},
			Message:     "",
			Context:     interfaces.EventContext{},
			DetectType:  "",
			DetectAlgo:  "",
		}
		Convey("success, =", func() {
			filter := []interfaces.Filter{
				{Name: "level", Operation: "=", Value: 1},
			}
			patch := ApplyFunc(did.GenerateDistributedID,
				func() (id uint64, err error) {
					return 1, nil
				},
			)
			defer patch.Reset()

			// engine.EXPECT().Call(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(interfaces.IEvents{}, 0, nil)
			// engine.EXPECT().GetLastEventLevel(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(1)
			_, err := engine.mit(events, filter)
			So(err, ShouldBeNil)
		})
		Convey("success, contain", func() {
			filter := []interfaces.Filter{
				{Name: "tags", Operation: "contain", Value: []string{"A"}},
			}
			patch := ApplyFunc(did.GenerateDistributedID,
				func() (id uint64, err error) {
					return 1, nil
				},
			)
			defer patch.Reset()

			// engine.EXPECT().Call(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(interfaces.IEvents{}, 0, nil)
			// engine.EXPECT().GetLastEventLevel(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(1)
			_, err := engine.mit(events, filter)
			So(err, ShouldBeNil)
		})
		Convey("success, in", func() {
			filter := []interfaces.Filter{
				{Name: "type", Operation: "in", Value: []string{"atomic"}},
			}
			patch := ApplyFunc(did.GenerateDistributedID,
				func() (id uint64, err error) {
					return 1, nil
				},
			)
			defer patch.Reset()

			// engine.EXPECT().Call(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(interfaces.IEvents{}, 0, nil)
			// engine.EXPECT().GetLastEventLevel(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(1)
			_, err := engine.mit(events, filter)
			So(err, ShouldBeNil)
		})
	})
}

// func TestEventEngine_Filter(t *testing.T) {
// 	Convey("Test Filter", t, func() {
// 		mockCtrl := gomock.NewController(t)
// 		defer mockCtrl.Finish()

// 		dvaMock := mock.NewMockDataViewAccess(mockCtrl)
// 		dataSource := []string{"1"}
// 		dataSourceType := "metric_model"
// 		kAcess := mock.NewMockKafkaAccess(mockCtrl)
// 		engine := MockNewEventEngine(eventModel, dataSource, dataSourceType, dvaMock, kAcess)
// 		events := interfaces.AtomicEvent{
// 			BaseEvent: interfaces.BaseEvent{
// 				Id:                  1,
// 				Title:               "hello",
// 				EventModelId:        "1",
// 				EventModelName:      "测试中的名称",
// 				EventType:           "atomic",
// 				Level:               1,
// 				Tags:                []string{"A", "B", "C"},
// 				GenerateType:        "batch",
// 				CreateTime:          time.Now(),
// 				DataSourceType:      "data_view",
// 				DataSource:          []string{"1"},
// 				DataSourceName:      []string{},
// 				DataSourceGroupName: []string{},
// 				DefaultTimeWindow:   interfaces.TimeInterval{Interval: 5, Unit: "m"},
// 				Schedule:            interfaces.Schedule{Type: "FIX_RATE", Expression: "5m"},
// 				Labels:              map[string]string{"A": "a", "B": "b"},
// 				Relations:           map[string]any{},
// 				LLMApp:              "",
// 				Ekn:                 "",
// 			},
// 			TriggerTime: time.Now(),
// 			TriggerData: interfaces.Records{},
// 			Message:     "",
// 			Context:     interfaces.EventContext{},
// 			DetectType:  "",
// 			DetectAlgo:  "",
// 		}
// 		Convey("success, =", func() {
// 			filter := []interfaces.Filter{
// 				{Name: "level", Operation: "=", Value: 1},
// 			}
// 			patch := ApplyFunc(did.GenerateDistributedID,
// 				func() (id uint64, err error) {
// 					return 1, nil
// 				},
// 			)
// 			defer patch.Reset()

// 			// engine.EXPECT().Call(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(interfaces.IEvents{}, 0, nil)
// 			// engine.EXPECT().GetLastEventLevel(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(1)
// 			_, _, err := engine.Filter(interfaces.IEvents{events}, filter)
// 			So(err, ShouldBeNil)
// 		})
// 		Convey("success, in", func() {
// 			filter := []interfaces.Filter{
// 				{Name: "type", Operation: "in", Value: []string{"atomic"}},
// 			}
// 			patch := ApplyFunc(did.GenerateDistributedID,
// 				func() (id uint64, err error) {
// 					return 1, nil
// 				},
// 			)
// 			defer patch.Reset()

// 			// engine.EXPECT().Call(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(interfaces.IEvents{}, 0, nil)
// 			// engine.EXPECT().GetLastEventLevel(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(1)
// 			_, _, err := engine.Filter(interfaces.IEvents{events}, filter)
// 			So(err, ShouldBeNil)
// 		})

// 		Convey("success, contain", func() {
// 			filter := []interfaces.Filter{
// 				{Name: "tags", Operation: "contain", Value: []string{"A"}},
// 			}
// 			patch := ApplyFunc(did.GenerateDistributedID,
// 				func() (id uint64, err error) {
// 					return 1, nil
// 				},
// 			)
// 			defer patch.Reset()

// 			// engine.EXPECT().Call(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(interfaces.IEvents{}, 0, nil)
// 			// engine.EXPECT().GetLastEventLevel(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(1)
// 			_, _, err := engine.Filter(interfaces.IEvents{events}, filter)
// 			So(err, ShouldBeNil)
// 		})

// 	})
// }

func TestEventEngine_Assemble(t *testing.T) {
	type fields struct {
		EventModel     interfaces.EventModel
		DataSource     any
		DataSourceType string
		DVAccess       interfaces.DataViewAccess
		// UAcess         interfaces.UniqueryAccess
		KAccess interfaces.KafkaAccess
	}
	type args struct {
		Events    interfaces.IEvents
		sortKey   string
		direction string
		limit     int64
		offset    int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    interfaces.IEvents
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rd := &EventEngine{
				EventModel:     tt.fields.EventModel,
				DataSource:     tt.fields.DataSource,
				DataSourceType: tt.fields.DataSourceType,
				DVAccess:       tt.fields.DVAccess,
				// UAcess:         tt.fields.UAcess,
				KAccess: tt.fields.KAccess,
			}
			got, err := rd.Assemble(tt.args.Events, tt.args.sortKey, tt.args.direction, tt.args.limit, tt.args.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("EventEngine.Assemble() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EventEngine.Assemble() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEventEginePersistQuery(t *testing.T) {
	Convey("Test PersistQuery", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		appSetting := &common.AppSetting{
			PoolSetting: common.PoolSetting{
				ViewPoolSize: 10,
			}}
		dvaMock := imock.NewMockDataViewAccess(mockCtrl)
		dataSource := []string{"1"}
		dataSourceType := "metric_model"
		kAcess := imock.NewMockKafkaAccess(mockCtrl)
		engine := MockNewEventEngine(eventModel, dataSource, dataSourceType, dvaMock, kAcess, appSetting)
		// Convey("PersistQuery failed: GetDataViewIDByName failed", func() {
		// 	event_query := interfaces.EventQuery{
		// 		QueryType: "range_query",
		// 		Id:     "1",
		// 		Start:     0,
		// 		End:       0,
		// 	}
		// 	dvaMock.EXPECT().GetDataViewIDByName(gomock.Any(), gomock.Any()).AnyTimes().Return("0", fmt.Errorf("error"))

		// 	_, err := engine.PersistQuery(testCtx,event_query, eventModel, true, true)
		// 	So(err, ShouldNotBeNil)
		// })
		Convey("PersistQuery failed: FetchSourceRecordsFrom failed", func() {
			event_query := interfaces.EventQuery{
				QueryType: "range_query",
				Id:        "1",
				Start:     0,
				End:       0,
			}
			dmq := NewLogDataQuery(appSetting, eventModel.DataSource, eventModel.DataSourceType, eventModel.DefaultTimeWindow, event_query)

			// dvaMock.EXPECT().GetDataViewIDByName(gomock.Any(), gomock.Any()).AnyTimes().Return("0", fmt.Errorf("error"))

			patch := ApplyMethodReturn(dmq, "FetchSourceRecordsFrom",
				interfaces.SourceRecords{}, interfaces.Record{}, fmt.Errorf("error"))
			defer patch.Reset()

			_, err := engine.PersistQuery(testCtx, event_query, eventModel, true, true)
			So(err, ShouldNotBeNil)
		})
		// @timestamp现在是int64了。unmarshal时会成功
		// Convey("failed: unmarshal failed : because @timestamp", func() {
		// 	event_query := interfaces.EventQuery{
		// 		QueryType: "range_query",
		// 		Id:        "1",
		// 		Start:     5,
		// 		End:       5,
		// 	}

		// 	dmq := NewLogDataQuery(appSetting, eventModel.DataSource, eventModel.DataSourceType, eventModel.DefaultTimeWindow, event_query)

		// 	sr := interfaces.SourceRecords{
		// 		Records: interfaces.Records{{
		// 			"event_type":              "atomic",
		// 			"default_time_window_str": "{\"interval\":5,\"unit\":\"m\"}",
		// 			"labels_str":              "{\"A\":\"V\"}",
		// 			"message":                 "",
		// 			"event_message":           "",
		// 			"context_str":             "{}",
		// 			"context":                 interfaces.EventContext{},
		// 			"trigger_time":            time.Now().UnixMilli(),
		// 			"trigger_data_str":        "[]",
		// 			"trigger_data":            interfaces.Records{},
		// 			"schedule_str":            "{\"type\":\"FIX_RATE\",\"expression\":\"5m\"}",
		// 			"aggregate_type":          "",
		// 			"detect_type":             "",
		// 			"aggregate_algo":          "",
		// 			"detect_algo":             "",
		// 			"@timestamp":              1724657861000,
		// 		}},
		// 	}

		// 	// dvaMock.EXPECT().GetDataViewIDByName(gomock.Any(), gomock.Any()).AnyTimes().Return("1", nil)

		// 	patch := ApplyMethodReturn(dmq, "FetchSourceRecordsFrom", sr, interfaces.Record{}, nil)
		// 	defer patch.Reset()

		// 	_, err := engine.PersistQuery(testCtx, event_query, eventModel, true, true)
		// 	So(err, ShouldNotBeNil)
		// })
		Convey("PersistQuery failed: SetDefaultTimeWindow failed", func() {
			event_query := interfaces.EventQuery{
				QueryType: "range_query",
				Id:        "1",
				Start:     0,
				End:       0,
			}

			dmq := NewLogDataQuery(appSetting, eventModel.DataSource, eventModel.DataSourceType, eventModel.DefaultTimeWindow, event_query)

			sr := interfaces.SourceRecords{
				Records: interfaces.Records{{
					"event_type":              "atomic",
					"default_time_window_str": "{\"interval\":,\"unit\":\"m\"}",
				}},
			}

			// dvaMock.EXPECT().GetDataViewIDByName(gomock.Any(), gomock.Any()).AnyTimes().Return("1", nil)

			patch := ApplyMethodReturn(dmq, "FetchSourceRecordsFrom", sr, interfaces.Record{}, nil)
			defer patch.Reset()

			_, err := engine.PersistQuery(testCtx, event_query, eventModel, true, true)
			So(err, ShouldNotBeNil)
		})
		Convey("PersistQuery failed: SetLabels failed", func() {
			event_query := interfaces.EventQuery{
				QueryType: "range_query",
				Id:        "1",
				Start:     0,
				End:       0,
			}

			dmq := NewLogDataQuery(appSetting, eventModel.DataSource, eventModel.DataSourceType, eventModel.DefaultTimeWindow, event_query)

			sr := interfaces.SourceRecords{
				Records: interfaces.Records{{
					"event_type":              "atomic",
					"default_time_window_str": "{\"interval\":5,\"unit\":\"m\"}",
					"labels_str":              "{\"A\"}",
				}},
			}

			// dvaMock.EXPECT().GetDataViewIDByName(gomock.Any(), gomock.Any()).AnyTimes().Return("1", nil)

			patch := ApplyMethodReturn(dmq, "FetchSourceRecordsFrom", sr, interfaces.Record{}, nil)
			defer patch.Reset()

			_, err := engine.PersistQuery(testCtx, event_query, eventModel, true, true)
			So(err, ShouldNotBeNil)
		})
		Convey("PersistQuery failed: SetTriggerData failed", func() {
			event_query := interfaces.EventQuery{
				QueryType: "range_query",
				Id:        "1",
				Start:     0,
				End:       0,
			}

			dmq := NewLogDataQuery(appSetting, eventModel.DataSource, eventModel.DataSourceType, eventModel.DefaultTimeWindow, event_query)

			sr := interfaces.SourceRecords{
				Records: interfaces.Records{{
					"event_type":              "atomic",
					"default_time_window_str": "{\"interval\":5,\"unit\":\"m\"}",
					"labels_str":              "{\"A\":\"V\"}",
					"message":                 "",
					"event_message":           "",
					"trigger_data_str":        "[ddd]",
				}},
			}

			// dvaMock.EXPECT().GetDataViewIDByName(gomock.Any(), gomock.Any()).AnyTimes().Return("1", nil)

			patch := ApplyMethodReturn(dmq, "FetchSourceRecordsFrom", sr, interfaces.Record{}, nil)
			defer patch.Reset()

			_, err := engine.PersistQuery(testCtx, event_query, eventModel, true, true)
			So(err, ShouldNotBeNil)
		})
		Convey("PersistQuery failed: SetSchedule failed", func() {
			event_query := interfaces.EventQuery{
				QueryType: "range_query",
				Id:        "1",
				Start:     0,
				End:       0,
			}

			dmq := NewLogDataQuery(appSetting, eventModel.DataSource, eventModel.DataSourceType, eventModel.DefaultTimeWindow, event_query)

			sr := interfaces.SourceRecords{
				Records: interfaces.Records{{
					"event_type":              "atomic",
					"default_time_window_str": "{\"interval\":5,\"unit\":\"m\"}",
					"labels_str":              "{\"A\":\"V\"}",
					"message":                 "",
					"event_message":           "",
					"trigger_data_str":        "[]",
					"schedule_str":            "{\"ty\"FIX_RATE\",\"expression\":\"5m\"}",
				}},
			}

			// dvaMock.EXPECT().GetDataViewIDByName(gomock.Any(), gomock.Any()).AnyTimes().Return("1", nil)

			patch := ApplyMethodReturn(dmq, "FetchSourceRecordsFrom", sr, interfaces.Record{}, nil)
			defer patch.Reset()

			_, err := engine.PersistQuery(testCtx, event_query, eventModel, true, true)
			So(err, ShouldNotBeNil)
		})

		Convey("PersistQuery failed: SetContext failed", func() {
			event_query := interfaces.EventQuery{
				QueryType: "range_query",
				Id:        "1",
				Start:     0,
				End:       0,
			}

			dmq := NewLogDataQuery(appSetting, eventModel.DataSource, eventModel.DataSourceType, eventModel.DefaultTimeWindow, event_query)

			sr := interfaces.SourceRecords{
				Records: interfaces.Records{{
					"event_type":              "atomic",
					"default_time_window_str": "{\"interval\":5,\"unit\":\"m\"}",
					"labels_str":              "{\"A\":\"V\"}",
					"message":                 "",
					"event_message":           "",
					"trigger_data_str":        "[]",
					"schedule_str":            "{\"ty\"FIX_RATE\",\"expression\":\"5m\"}",
					"context_str":             "{}",
					"context":                 interfaces.EventContext{},
				}},
			}
			// dvaMock.EXPECT().GetDataViewIDByName(gomock.Any(), gomock.Any()).AnyTimes().Return("1", nil)

			patch := ApplyMethodReturn(dmq, "FetchSourceRecordsFrom", sr, interfaces.Record{}, nil)
			defer patch.Reset()

			_, err := engine.PersistQuery(testCtx, event_query, eventModel, true, true)
			So(err, ShouldNotBeNil)
		})
		Convey("success, start,end == -1", func() {
			event_query := interfaces.EventQuery{
				QueryType: "range_query",
				Id:        "1",
				Start:     -1,
				End:       -1,
			}

			dmq := NewLogDataQuery(appSetting, eventModel.DataSource, eventModel.DataSourceType, eventModel.DefaultTimeWindow, event_query)

			sr := interfaces.SourceRecords{
				Records: interfaces.Records{{
					"event_type":              "atomic",
					"default_time_window_str": "{\"interval\":5,\"unit\":\"m\"}",
					"labels_str":              "{\"A\":\"V\"}",
					"message":                 "",
					"event_message":           "",
					"context_str":             "{}",
					"context":                 interfaces.EventContext{},
					"trigger_time":            time.Now().UnixMilli(),
					"trigger_data_str":        "[]",
					"trigger_data":            interfaces.Records{},
					"schedule_str":            "{\"type\":\"FIX_RATE\",\"expression\":\"5m\"}",
					"aggregate_type":          "",
					"detect_type":             "",
					"aggregate_algo":          "",
					"detect_algo":             "",
				}},
			}

			// dvaMock.EXPECT().GetDataViewIDByName(gomock.Any(), gomock.Any()).AnyTimes().Return("1", nil)

			patch := ApplyMethodReturn(dmq, "FetchSourceRecordsFrom", sr, interfaces.Record{}, nil)
			defer patch.Reset()

			_, err := engine.PersistQuery(testCtx, event_query, eventModel, true, true)
			So(err, ShouldBeNil)
		})
		Convey("success", func() {
			event_query := interfaces.EventQuery{
				QueryType: "range_query",
				Id:        "1",
				Start:     0,
				End:       0,
			}

			dmq := NewLogDataQuery(appSetting, eventModel.DataSource, eventModel.DataSourceType, eventModel.DefaultTimeWindow, event_query)

			sr := interfaces.SourceRecords{
				Records: interfaces.Records{{
					"event_type":              "atomic",
					"default_time_window_str": "{\"interval\":5,\"unit\":\"m\"}",
					"labels_str":              "{\"A\":\"V\"}",
					"message":                 "",
					"event_message":           "",
					"context_str":             "{}",
					"context":                 interfaces.EventContext{},
					"trigger_time":            time.Now().UnixMilli(),
					"trigger_data_str":        "[]",
					"trigger_data":            interfaces.Records{},
					"schedule_str":            "{\"type\":\"FIX_RATE\",\"expression\":\"5m\"}",
					"aggregate_type":          "",
					"detect_type":             "",
					"aggregate_algo":          "",
					"detect_algo":             "",
				}},
			}

			// dvaMock.EXPECT().GetDataViewIDByName(gomock.Any(), gomock.Any()).AnyTimes().Return("1", nil)

			patch := ApplyMethodReturn(dmq, "FetchSourceRecordsFrom", sr, interfaces.Record{}, nil)
			defer patch.Reset()

			_, err := engine.PersistQuery(testCtx, event_query, eventModel, true, false)
			So(err, ShouldBeNil)
		})

	})
}

func TestEventEgineGetLastEventLevel(t *testing.T) {
	Convey("Test GetLastEventLevel", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		// uaMock := imock.NewMockUniqueryAccess(mockCtrl)
		dvaMock := imock.NewMockDataViewAccess(mockCtrl)
		dataSource := []string{"1"}
		dataSourceType := "metric_model"
		kAcess := imock.NewMockKafkaAccess(mockCtrl)
		engine := MockNewEventEngine(eventModel, dataSource, dataSourceType, dvaMock, kAcess, appSetting)
		key := "labelkey"
		Convey("success, len(events) == 0", func() {
			event_query := interfaces.EventQuery{
				QueryType: "range_query",
				Id:        "1",
				Start:     0,
				End:       0,
			}

			dmq := NewLogDataQuery(appSetting, eventModel.DataSource, eventModel.DataSourceType, eventModel.DefaultTimeWindow, event_query)
			// dvaMock.EXPECT().GetDataViewIDByName(gomock.Any(), gomock.Any()).AnyTimes().Return("1", nil)

			patch := ApplyMethodReturn(dmq, "FetchSourceRecordsFrom",
				interfaces.SourceRecords{}, interfaces.Record{}, nil)
			defer patch.Reset()

			level := engine.GetLastEventLevel(testCtx, key, eventModel)
			So(level, ShouldEqual, 0)
		})

		Convey("success", func() {
			event_query := interfaces.EventQuery{
				QueryType: "range_query",
				Id:        "1",
				Start:     0,
				End:       0,
			}

			dmq := NewLogDataQuery(appSetting, eventModel.DataSource, eventModel.DataSourceType, eventModel.DefaultTimeWindow, event_query)

			sr := interfaces.SourceRecords{
				Records: interfaces.Records{{
					"event_type": "atomic", "default_time_window_str": "{\"interval\":5,\"unit\":\"m\"}",
					"labels_str": "{\"A\":\"V\"}", "message": "", "event_message": "",
					"context_str": "{}", "context": interfaces.EventContext{},
					"trigger_time": time.Now(), "trigger_data_str": "[]", "trigger_data": interfaces.Records{},
					"schedule_str": "{\"type\":\"FIX_RATE\",\"expression\":\"5m\"}", "aggregate_type": "", "detect_type": "", "aggregate_algo": "", "detect_algo": "",
				}},
			}
			patch := ApplyMethodReturn(dmq, "FetchSourceRecordsFrom", sr, interfaces.Record{}, nil)
			defer patch.Reset()

			_, _, err := engine.Query(testCtx, event_query, eventModel)
			So(err, ShouldBeNil)
		})

	})
}

func TestEventEgineQuery(t *testing.T) {
	Convey("Test Query", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		// uaMock := imock.NewMockUniqueryAccess(mockCtrl)
		dvaMock := imock.NewMockDataViewAccess(mockCtrl)
		dataSource := []string{"1"}
		dataSourceType := "metric_model"
		kAcess := imock.NewMockKafkaAccess(mockCtrl)
		engine := MockNewEventEngine(eventModel, dataSource, dataSourceType, dvaMock, kAcess, appSetting)
		event_query := interfaces.EventQuery{
			QueryType: "range_query",
			Id:        "1",
			Start:     0,
			End:       0,
		}
		Convey("failed, caused by FetchSourceRecordsFrom failed", func() {
			dmq := NewLogDataQuery(appSetting, eventModel.DataSource, eventModel.DataSourceType, eventModel.DefaultTimeWindow, event_query)

			patch := ApplyMethodReturn(dmq, "FetchSourceRecordsFrom",
				interfaces.SourceRecords{}, interfaces.Record{}, fmt.Errorf("some error"))
			defer patch.Reset()

			_, _, err := engine.Query(testCtx, event_query, eventModel)
			So(err, ShouldNotBeNil)
		})
		Convey("success", func() {
			event_query := interfaces.EventQuery{
				QueryType: "range_query",
				Id:        "1",
				Start:     0,
				End:       0,
			}

			dmq := NewLogDataQuery(appSetting, eventModel.DataSource, eventModel.DataSourceType, eventModel.DefaultTimeWindow, event_query)

			sr := interfaces.SourceRecords{
				Records: interfaces.Records{{
					"event_type": "atomic", "default_time_window_str": "{\"interval\":5,\"unit\":\"m\"}",
					"labels_str": "{\"A\":\"V\"}", "message": "", "event_message": "",
					"context_str": "{}", "context": interfaces.EventContext{},
					"trigger_time": time.Now(), "trigger_data_str": "[]", "trigger_data": interfaces.Records{},
					"schedule_str": "{\"type\":\"FIX_RATE\",\"expression\":\"5m\"}", "aggregate_type": "", "detect_type": "", "aggregate_algo": "", "detect_algo": "",
				}},
			}
			patch := ApplyMethodReturn(dmq, "FetchSourceRecordsFrom", sr, interfaces.Record{}, nil)
			defer patch.Reset()

			_, _, err := engine.Query(testCtx, event_query, eventModel)
			So(err, ShouldBeNil)
		})

	})
}
