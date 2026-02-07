// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package event

import (
	"encoding/json"
	"errors"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/golang/mock/gomock"
	libmq "github.com/kweaver-ai/kweaver-go-lib/mq"
	"github.com/patrickmn/go-cache"
	. "github.com/smartystreets/goconvey/convey"

	"uniquery/common"
	"uniquery/interfaces"
	imock "uniquery/interfaces/mock"
)

type ViewField struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Comment string `json:"comment"`

	Path []string `json:"-"`
}

func MockNewEventSubService(appSetting *common.AppSetting,
	engine interfaces.EventEngine,
	emAcces interfaces.EventModelAccess,
	kafkaAcess interfaces.KafkaAccess, topics []string) *eventSubService {

	return &eventSubService{
		appSetting:  appSetting,
		engine:      nil,
		emAccess:    emAcces,
		kafkaAccess: kafkaAcess,
		topics:      topics,
	}

	// return ess
}

func TestFlatten(t *testing.T) {
	Convey("Test Flatten", t, func() {
		Convey("success, value is {}", func() {
			top := true
			prefix := ""
			srcStr := `{"a": {"b": 1, "c": 2, "d": {}}}`
			var src map[string]any
			err := json.Unmarshal([]byte(srcStr), &src)
			if err != nil {
				t.Fatal(err)
			}
			dest := make(map[string]any)

			var expectRes map[string]any
			expectResStr := `{"a.b": 1, "a.c": 2, "a.d": {}}`
			err = json.Unmarshal([]byte(expectResStr), &expectRes)
			if err != nil {
				t.Fatal(err)
			}

			err = flatten(top, prefix, src, dest)
			So(err, ShouldBeNil)

			So(dest, ShouldResemble, expectRes)
		})

		Convey("success, value is []", func() {
			top := true
			prefix := ""
			srcStr := `{"a": {"b": 1, "c": 2, "d": []}}`
			var src map[string]any
			err := json.Unmarshal([]byte(srcStr), &src)
			if err != nil {
				t.Fatal(err)
			}
			dest := make(map[string]any)

			var expectRes map[string]any
			expectResStr := `{"a.b": 1, "a.c": 2, "a.d": []}`
			err = json.Unmarshal([]byte(expectResStr), &expectRes)
			if err != nil {
				t.Fatal(err)
			}

			err = flatten(top, prefix, src, dest)
			So(err, ShouldBeNil)

			So(dest, ShouldResemble, expectRes)
		})

		Convey("success, value is [object1, object2]", func() {
			top := true
			prefix := ""
			srcStr := `{"a": {"b": 1, "c": 2, "d": [{"e": 3}, {"f": 4, "e": 5}]}}`
			var src map[string]any
			err := json.Unmarshal([]byte(srcStr), &src)
			if err != nil {
				t.Fatal(err)
			}
			dest := make(map[string]any)

			var expectRes map[string]any
			expectResStr := `{"a.b": 1, "a.c": 2, "a.d.e": [3, 5], "a.d.f": 4}`
			err = json.Unmarshal([]byte(expectResStr), &expectRes)
			if err != nil {
				t.Fatal(err)
			}

			err = flatten(top, prefix, src, dest)
			So(err, ShouldBeNil)

			So(dest, ShouldResemble, expectRes)
		})
	})
}

func Test_eventSubService_Invoke(t *testing.T) {
	Convey("Test Flatten", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		engine := imock.NewMockEventEngine(mockCtrl)
		topics := []string{"default.mdl.view"}
		emAccess := imock.NewMockEventModelAccess(mockCtrl)
		kafkaAccess := imock.NewMockKafkaAccess(mockCtrl)
		_cache := cache.New(interfaces.EXPIRATION_TIME, interfaces.DELETE_TIME)
		records := kafka.Message{
			Value:   []byte(`{"a": {"b": {"c": "wahaha"}}}`),
			Headers: []kafka.Header{{Key: "__view_id", Value: []byte("1")}},
		}

		ess := MockNewEventSubService(&common.AppSetting{}, engine, emAccess, kafkaAccess, topics)
		Convey("success, value is {}", func() {
			emAccess.EXPECT().GetEventModelBySourceId(gomock.Any(), gomock.Any()).AnyTimes().Return([]interfaces.EventModel{eventModel}, nil)
			engine.EXPECT().Judge(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(interfaces.IEvents{}, 0, nil)
			_, _, err := ess.Invoke(engine, eventModel.DataSource[0], "data_view", []*kafka.Message{&records}, _cache)
			So(err, ShouldBeNil)
		})
	})

}

// func Test_eventSubService_Subscribe(t *testing.T) {
// 	type fields struct {
// 		appSetting  *common.AppSetting
// 		engine      interfaces.EventEngine
// 		emAccess    interfaces.EventModelAccess
// 		kafkaAccess interfaces.KafkaAccess
// 		topics      []string
// 	}
// 	type args struct {
// 		view_id string
// 	}
// 	mockCtrl := gomock.NewController(t)
// 	defer mockCtrl.Finish()

// 	field := fields{
// 		appSetting:  &common.AppSetting{},
// 		emAccess:    imock.NewMockEventModelAccess(mockCtrl),
// 		kafkaAccess: imock.NewMockKafkaAccess(mockCtrl),
// 		topics:      []string{"default.mdl.view"},
// 		engine:      imock.NewMockEventEngine(mockCtrl),

// 		// eventModelService:  imock.NewMockEventModelService(mockCtrl),
// 	}

// 	tests := []struct {
// 		name   string
// 		fields fields
// 		args   args
// 		want   []interfaces.EventModel
// 	}{
// 		// TODO: Add test cases.
// 		{"success", field, args{view_id: "1"}, nil},
// 	}

// 	records := kafka.Message{
// 		Value:   []byte(`{"a": {"b": {"c": "wahaha"}}}`),
// 		Headers: []kafka.Header{{Key: "__view_id", Value: []byte("1")}},
// 	}
// 	// field.metricModelService.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, nil)
// 	emAccess := imock.NewMockEventModelAccess(mockCtrl)
// 	kafkaAccess := imock.NewMockKafkaAccess(mockCtrl)
// 	engine := imock.NewMockEventEngine(mockCtrl)
// 	topics := []string{"default.mdl.view"}
// 	// groupID := "default.mdl.event_subscribe"
// 	// kafkaProtocol := "sasl_ssl"
// 	// config := kafka.ConfigMap{
// 	// 	"bootstrap.servers":  "10.4.110.255:31000",
// 	// 	"group.id":           groupID,
// 	// 	"enable.auto.commit": false,
// 	// 	// "session.timeout.ms":        ka.appSetting.KafkaSetting.SessionTimeoutMs,
// 	// 	// "socket.timeout.ms":         ka.appSetting.KafkaSetting.SocketTimeoutMs,
// 	// 	"socket.keepalive.enable": true,
// 	// 	// "heartbeat.interval.ms":     ka.appSetting.KafkaSetting.HeartbeatIntervalMs,
// 	// 	// "security.protocol":    kafkaProtocol,
// 	// 	"auto.offset.reset":    "latest",
// 	// 	"enable.partition.eof": true,
// 	// 	// "max.poll.interval.ms": ka.appSetting.KafkaSetting.MaxPollIntervalMs,
// 	// 	// "max.poll.records":          ka.appSetting.KafkaSetting.MaxPollRecords,
// 	// 	"max.partition.fetch.bytes": interfaces.MAX_MESSAGE_BYTES,
// 	// 	// "fetch.wait.max.ms":         ka.appSetting.KafkaSetting.FetchWaitMaxMs,
// 	// }
// 	// consumer, err := kafka.NewConsumer(&config)
// 	// fmt.Printf("err: %v\n", err)
// 	// producer, err := kafka.NewProducer(&config)
// 	// fmt.Printf("err: %v\n", err)

// 	consumer, _ := kafka.NewConsumer(&kafka.ConfigMap{"bootstrap.servers": "localhost"})
// 	producer, _ := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost"})

// 	ess := MockNewEventSubService(&common.AppSetting{}, engine, emAccess, kafkaAccess, topics)
// 	exitCh := make(chan bool)
// 	md := &kafka.Metadata{
// 		Brokers: []kafka.BrokerMetadata{{
// 			ID:   1,
// 			Host: "",
// 			Port: 8080,
// 		}},

// 		Topics: map[string]kafka.TopicMetadata{
// 			"default.mdl.view.1": {Topic: "default.mdl.view.1", Partitions: []kafka.PartitionMetadata{}},
// 			"default.mdl.view":   {Topic: "default.mdl.view", Partitions: []kafka.PartitionMetadata{}},
// 		}}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			emAccess.EXPECT().GetEventModelByViewId(gomock.Any(), gomock.Any()).AnyTimes().Return([]interfaces.EventModel{eventModel}, nil)
// 			kafkaAccess.EXPECT().NewKafkaConsumer().AnyTimes().Return(consumer, nil)
// 			kafkaAccess.EXPECT().NewTrxProducer("default.mdl.raw.dip_event_model_data").AnyTimes().Return(producer, nil)
// 			kafkaAccess.EXPECT().NewTrxProducer("default.mdl.event_alert").AnyTimes().Return(producer, nil)
// 			kafkaAccess.EXPECT().CreateTopicIfNotPresent(interfaces.DEFAULT_SUBSCRIBE_TOPIC).Times(1).Return(nil)
// 			kafkaAccess.EXPECT().PollMessages(gomock.Any()).AnyTimes().Return([]*kafka.Message{&records}, nil)
// 			kafkaAccess.EXPECT().DoProduce(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
// 			kafkaAccess.EXPECT().CommitOffset(gomock.Any()).AnyTimes().Return(nil)

// 			patches := ApplyMethodReturn(consumer, "GetMetadata", md, nil)
// 			defer patches.Reset()
// 			// SubscribeTopics
// 			patch1 := ApplyMethodReturn(consumer, "SubscribeTopics", nil)
// 			defer patch1.Reset()
// 			ess.Subscribe(exitCh)
// 			if got := ess.Subscribe(exitCh); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("eventSubService.GetEventModelByViewID() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func Test_eventSubService_getInitTopics(t *testing.T) {
	Convey("Test getInitTopics", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		// field.metricModelService.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, nil)
		emAccess := imock.NewMockEventModelAccess(mockCtrl)
		kafkaAccess := imock.NewMockKafkaAccess(mockCtrl)
		engine := imock.NewMockEventEngine(mockCtrl)
		topics := []string{"default.mdl.view"}
		// consumer, _ := kafka.NewConsumer(&kafka.ConfigMap{"bootstrap.servers": "localhost"})
		admin, _ := kafka.NewAdminClient(&kafka.ConfigMap{"bootstrap.servers": "localhost"})
		producer, _ := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost"})
		patches1 := ApplyMethodReturn(admin, "Close")
		defer patches1.Reset()

		// patches2 := ApplyMethodReturn(producer, "Close", nil)
		// defer patches2.Reset()
		ess := MockNewEventSubService(&common.AppSetting{
			MQSetting: libmq.MQSetting{
				MQType: "kafka",
				MQHost: "localhost",
				MQPort: 9092,
				Tenant: "default",
				Auth: libmq.MQAuthSetting{
					Username: "test",
					Password: "testpwd",
				},
			},
		}, engine, emAccess, kafkaAccess, topics)
		// exitCh := make(chan bool)
		md := &kafka.Metadata{
			Brokers: []kafka.BrokerMetadata{{
				ID:   1,
				Host: "",
				Port: 8080,
			}},

			Topics: map[string]kafka.TopicMetadata{
				"default.mdl.view.1": {Topic: "default.mdl.view.1", Partitions: []kafka.PartitionMetadata{}},
				"default.mdl.view":   {Topic: "default.mdl.view", Partitions: []kafka.PartitionMetadata{}},
			}}

		emAccess.EXPECT().GetEventModelBySourceId(gomock.Any(), gomock.Any()).AnyTimes().Return([]interfaces.EventModel{eventModel}, nil)
		// kafkaAccess.EXPECT().NewKafkaConsumer().AnyTimes().Return(consumer, nil)
		kafkaAccess.EXPECT().NewKafkaAdminClient().AnyTimes().Return(admin, nil)
		kafkaAccess.EXPECT().NewTrxProducer("default.mdl.raw.dip_event_model_data").AnyTimes().Return(producer, nil)
		kafkaAccess.EXPECT().CreateTopicIfNotPresent("default.mdl.view").AnyTimes().Return(nil)

		patches := ApplyMethodReturn(admin, "GetMetadata", md, nil)
		defer patches.Reset()

		// ess.Subscribe(exitCh)
		_, err := ess.getInitTopics()
		So(err, ShouldBeNil)

	})

}

func Test_joinKey(t *testing.T) {
	type args struct {
		top    bool
		prefix string
		subkey string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{"success", args{top: true, prefix: "a", subkey: "b"}, "ab"},
		{"success", args{top: false, prefix: "a", subkey: "b"}, "a.b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := joinKey(tt.args.top, tt.args.prefix, tt.args.subkey); got != tt.want {
				t.Errorf("joinKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEventSubServiceFlushAtomicEvent(t *testing.T) {
	Convey("Test FlushAtomicEvent", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		emAccess := imock.NewMockEventModelAccess(mockCtrl)
		kafkaAccess := imock.NewMockKafkaAccess(mockCtrl)
		engine := imock.NewMockEventEngine(mockCtrl)
		topics := []string{"default.mdl.view"}
		consumer, _ := kafka.NewConsumer(&kafka.ConfigMap{"bootstrap.servers": "localhost"})
		producer, _ := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost"})
		patches1 := ApplyMethodReturn(consumer, "Close", nil)
		defer patches1.Reset()

		records := kafka.Message{
			Value: []byte(`{"a": {"b": {"c": "wahaha"}}}`),
			Headers: []kafka.Header{
				{Key: "__view_id", Value: []byte("1")},
				{Key: "dataSourceType", Value: []byte("event_model")},
			},
		}

		Convey("failed,caused by new producer error", func() {
			ess := MockNewEventSubService(&common.AppSetting{}, engine, emAccess, kafkaAccess, topics)

			kafkaAccess.EXPECT().NewTrxProducer(gomock.Any()).AnyTimes().Return(nil, errors.New("some error"))

			_, err := ess.FlushAtomicEvent([]*kafka.Message{&records})
			So(err, ShouldNotBeNil)
		})
		Convey("failed,caused by do produce error", func() {
			ess := MockNewEventSubService(&common.AppSetting{}, engine, emAccess, kafkaAccess, topics)

			kafkaAccess.EXPECT().NewTrxProducer(gomock.Any()).AnyTimes().Return(producer, nil)
			kafkaAccess.EXPECT().DoProduce(gomock.Any(), gomock.Any()).AnyTimes().Return(errors.New("some error"))
			_, err := ess.FlushAtomicEvent([]*kafka.Message{&records})
			So(err, ShouldNotBeNil)
		})
		Convey("success", func() {
			ess := MockNewEventSubService(&common.AppSetting{}, engine, emAccess, kafkaAccess, topics)

			kafkaAccess.EXPECT().NewTrxProducer(gomock.Any()).AnyTimes().Return(producer, nil)
			kafkaAccess.EXPECT().DoProduce(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			_, err := ess.FlushAtomicEvent([]*kafka.Message{&records})
			So(err, ShouldBeNil)
		})
		Convey("success length 0", func() {
			ess := MockNewEventSubService(&common.AppSetting{}, engine, emAccess, kafkaAccess, topics)

			kafkaAccess.EXPECT().NewTrxProducer(gomock.Any()).AnyTimes().Return(producer, nil)
			kafkaAccess.EXPECT().DoProduce(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			_, err := ess.FlushAtomicEvent([]*kafka.Message{})
			So(err, ShouldBeNil)
		})
	})
}

func TestEventSubServiceDivideRecordBySourceId(t *testing.T) {
	Convey("Test DivideRecordBySourceId", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		emAccess := imock.NewMockEventModelAccess(mockCtrl)
		kafkaAccess := imock.NewMockKafkaAccess(mockCtrl)
		engine := imock.NewMockEventEngine(mockCtrl)
		topics := []string{"default.mdl.view"}
		essMock := MockNewEventSubService(&common.AppSetting{}, engine, emAccess, kafkaAccess, topics)

		records := []*kafka.Message{
			{
				Value: []byte(`{"a": {"b": {"c": "wahaha"}}}`),
				Headers: []kafka.Header{
					{Key: "__view_id", Value: []byte("1")},
					{Key: "dataSourceType", Value: []byte("event_model")},
				},
			},
		}
		expected := map[interfaces.DataSource][]*kafka.Message{
			{DataSourceId: "1", DataSourceType: "event_model"}: {records[0]},
		}

		Convey("success,no dataSourceId", func() {
			records := []*kafka.Message{
				{
					Value: []byte(`{"a": {"b": {"c": "wahaha"}}}`),
					Headers: []kafka.Header{
						{Key: "__view_id", Value: []byte("1")},
						{Key: "dataSourceType", Value: []byte("event_model")},
					},
				},
				{
					Value: []byte(`{"a": {"b": {"c": "wahaha"}}}`),
					Headers: []kafka.Header{
						{Key: "__view_id", Value: []byte("")},
						{Key: "dataSourceType", Value: []byte("event_model")},
					},
				},
			}
			result := essMock.divideRecordBySourceId(records)
			So(result, ShouldEqual, expected)
		})
		Convey("success", func() {
			result := essMock.divideRecordBySourceId(records)
			So(result, ShouldEqual, expected)
		})

	})
}

// func Test_eventSubService_DoProcess(t *testing.T) {
// 	Convey("Test DoProcess", t, func() {

// 		mockCtrl := gomock.NewController(t)
// 		defer mockCtrl.Finish()

// 		// field.metricModelService.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, nil)
// 		emAccess := dmock.NewMockEventModelAccess(mockCtrl)
// 		kafkaAccess := dmock.NewMockKafkaAccess(mockCtrl)
// 		engine := dmock.NewMockEventEngine(mockCtrl)
// 		topics := []string{"default.mdl.view"}
// 		consumer, _ := kafka.NewConsumer(&kafka.ConfigMap{"bootstrap.servers": "localhost"})
// 		producer, _ := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost"})
// 		patches1 := ApplyMethodReturn(consumer, "Close", nil)
// 		defer patches1.Reset()
// 		records := kafka.Message{
// 			Value:   []byte(`{"a": {"b": {"c": "wahaha"}}}`),
// 			Headers: []kafka.Header{{Key: "__view_id", Value: []byte("1")}},
// 		}

// 		// patches2 := ApplyMethodReturn(producer, "Close", nil)
// 		// defer patches2.Reset()
// 		ess := MockNewEventSubService(&common.AppSetting{}, engine, emAccess, kafkaAccess, topics)
// 		// exitCh := make(chan bool)

// 		// emAccess.EXPECT().GetEventModelByViewId(gomock.Any(), gomock.Any()).AnyTimes().Return([]interfaces.EventModel{eventModel}, nil)
// 		// kafkaAccess.EXPECT().NewKafkaConsumer().AnyTimes().Return(consumer, nil)
// 		kafkaAccess.EXPECT().NewTrxProducer(gomock.Any()).AnyTimes().Return(producer, nil)
// 		kafkaAccess.EXPECT().DoProduce(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
// 		kafkaAccess.EXPECT().PollMessages(gomock.Any()).AnyTimes().Return([]*kafka.Message{&records}, nil)
// 		kafkaAccess.EXPECT().CommitOffset(gomock.Any()).AnyTimes().Return(nil)
// 		patches := ApplyPrivateMethod(ess, "InitEngine",
// 			func(ctx context.Context, em interfaces.EventModel) (interfaces.EventEngine, error) {
// 				return engine, nil
// 			})
// 		defer patches.Reset()

// 		patches2 := ApplyPrivateMethod(ess, "Invoke",
// 			func(engine interfaces.EventEngine, record *kafka.Message, _cache *cache.Cache) ([]*kafka.Message, error) {
// 				return []*kafka.Message{&records}, nil
// 			})

// 		defer patches2.Reset()
// 		// ess.Subscribe(exitCh)
// 		_cache := cache.New(interfaces.EXPIRATION_TIME, interfaces.DELETE_TIME)
// 		err := ess.DoProcess(consumer, producer, _cache)
// 		So(err, ShouldBeNil)

// 	})
// }

func TestEventSubServiceGetEventModelByDownstreamDependent(t *testing.T) {
	Convey("Test GetEventModelByDownstreamDependent", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		emAccess := imock.NewMockEventModelAccess(mockCtrl)
		kafkaAccess := imock.NewMockKafkaAccess(mockCtrl)
		engine := imock.NewMockEventEngine(mockCtrl)
		topics := []string{"default.mdl.view"}

		essMock := MockNewEventSubService(&common.AppSetting{}, engine, emAccess, kafkaAccess, topics)
		Convey("failed \n", func() {
			emAccess.EXPECT().GetEventModelById(gomock.Any(), gomock.Any()).Return([]interfaces.EventModel{}, nil)
			models := essMock.GetEventModelByDownstreamDependent("1", "event_model")

			So(len(models), ShouldEqual, 0)
		})
		Convey("Success \n", func() {
			emAccess.EXPECT().GetEventModelById(gomock.Any(), gomock.Any()).AnyTimes().Return([]interfaces.EventModel{eventModel, eventModel}, nil)
			models := essMock.GetEventModelByDownstreamDependent("1", "event_model")

			So(len(models), ShouldEqual, 1)

		})
	})
}

func TestEventSubServiceQueryEventModelByDownstreamDependent(t *testing.T) {
	Convey("Test QueryEventModelByDownstreamDependent", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		emAccess := imock.NewMockEventModelAccess(mockCtrl)
		kafkaAccess := imock.NewMockKafkaAccess(mockCtrl)
		engine := imock.NewMockEventEngine(mockCtrl)
		topics := []string{"default.mdl.view"}

		essMock := MockNewEventSubService(&common.AppSetting{}, engine, emAccess, kafkaAccess, topics)
		Convey("Success, no ems \n", func() {
			emAccess.EXPECT().GetEventModelById(gomock.Any(), gomock.Any()).Return([]interfaces.EventModel{}, nil)

			models, err := essMock.QueryEventModelByDownstreamDependent("1", "event_model", cache.New(interfaces.EXPIRATION_TIME, interfaces.DELETE_TIME))
			So(len(models), ShouldEqual, 0)
			So(err, ShouldBeNil)
		})
		Convey("Success \n", func() {
			emAccess.EXPECT().GetEventModelById(gomock.Any(), gomock.Any()).AnyTimes().Return([]interfaces.EventModel{eventModel, eventModel}, nil)
			_, err := essMock.QueryEventModelByDownstreamDependent("1", "event_model", cache.New(interfaces.EXPIRATION_TIME, interfaces.DELETE_TIME))
			So(err, ShouldBeNil)
		})

	})
}
