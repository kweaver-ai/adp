// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package drivenadapters

import (
	"errors"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	libmq "github.com/kweaver-ai/kweaver-go-lib/mq"
	. "github.com/smartystreets/goconvey/convey"

	"uniquery/common"
)

func Test_kafkaAccess_NewKafkaConsumer(t *testing.T) {
	Convey("Test KafkaAccess NewConsumer", t, func() {
		kaTest := &kafkaAccess{}
		kaTest.appSetting = &common.AppSetting{
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
		}

		Convey("NewConsumer failed", func() {
			patches := ApplyFuncReturn(kafka.NewConsumer, nil, errors.New("error"))
			defer patches.Reset()

			_, err := kaTest.NewKafkaConsumer()
			So(err, ShouldNotBeNil)
		})

		Convey("success", func() {
			c := &kafka.Consumer{}
			patches := ApplyFuncReturn(kafka.NewConsumer, c, nil)
			defer patches.Reset()

			_, err := kaTest.NewKafkaConsumer()
			So(err, ShouldBeNil)
		})
	})
}

func TestNewTrxProducer(t *testing.T) {
	Convey("Test KafkaAccess NewTransactionalProducer", t, func() {
		kaTest := &kafkaAccess{}
		kaTest.appSetting = &common.AppSetting{
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
		}
		txId := "test"

		Convey("NewTransactionalProducer failed", func() {
			patches := ApplyFuncReturn(kafka.NewProducer, nil, errors.New("error"))
			defer patches.Reset()
			_, err := kaTest.NewTrxProducer(txId)
			So(err, ShouldNotBeNil)
		})

		Convey("success", func() {
			p := &kafka.Producer{}
			patches := ApplyFuncReturn(kafka.NewProducer, p, nil)
			defer patches.Reset()
			patch2 := ApplyMethodReturn(p, "InitTransactions", nil)
			defer patch2.Reset()

			_, err := kaTest.NewTrxProducer(txId)
			So(err, ShouldBeNil)
		})
	})
}

func Test_kafkaAccess_PollMessages(t *testing.T) {
	Convey("Test KafkaAccess PollMessages", t, func() {
		kaTest := &kafkaAccess{}
		kaTest.appSetting = &common.AppSetting{
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
		}
		c := &kafka.Consumer{}
		Convey("PollMessages success", func() {
			patches := ApplyMethodReturn(c, "Poll", &kafka.Message{})
			defer patches.Reset()
			_, err := kaTest.PollMessages(c)
			So(err, ShouldBeNil)
		})
	})
}

func Test_kafkaAccess_CreateTopicIfNotPresent(t *testing.T) {
	Convey("Test KafkaAccess CreateTopicIfNotPresent", t, func() {
		ka := &kafkaAccess{
			appSetting: &common.AppSetting{
				MQSetting: libmq.MQSetting{
					MQType: "kafka",
					MQHost: "kafka",
					MQPort: 9092,
					Tenant: "default",
					Auth: libmq.MQAuthSetting{
						Mechanism: "PLAIN",
						Username:  "test",
						Password:  "testpwd",
					},
				},
			},
		}

		// client := &kafka.AdminClient{}
		// resultsCreate := []kafka.TopicResult{
		// 	{
		// 		Topic: "default.mdl.view",
		// 		Error: kafka.Error{},
		// 	},
		// }
		// resultDescribeResult := kafka.DescribeTopicsResult{
		// 	TopicDescriptions: []kafka.TopicDescription{
		// 		{Name: "default.mdl.view"},
		// 	},
		// }
		Convey("CreateTopicIfNotPresent success", func() {
			// patch1 := ApplyFunc(kafka.NewAdminClient,
			// 	func(_ *kafka.ConfigMap) (*kafka.AdminClient, error) {
			// 		return client, nil
			// 	},
			// )
			// defer patch1.Reset()
			// patch := ApplyMethod(client, "DescribeTopics",
			// 	func(*kafka.AdminClient, context.Context, kafka.TopicCollection,
			// 		...kafka.DescribeTopicsAdminOption) (kafka.DescribeTopicsResult, error) {
			// 		return resultDescribeResult, nil
			// 	},
			// )
			// defer patch.Reset()
			// patch3 := ApplyMethod(client, "CreateTopics",
			// 	func(*kafka.AdminClient, context.Context, []kafka.TopicSpecification, ...kafka.CreateTopicsAdminOption) ([]kafka.TopicResult, error) {
			// 		return resultsCreate, nil
			// 	},
			// )
			// defer patch3.Reset()
			err := ka.CreateTopicIfNotPresent([]string{})
			So(err, ShouldBeNil)
		})
	})
}
