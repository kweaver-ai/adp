// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package drivenadapters

import (
	"errors"
	"math/rand"
	"testing"

	libmq "github.com/kweaver-ai/kweaver-go-lib/mq"
	. "github.com/agiledragon/gomonkey/v2"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	. "github.com/smartystreets/goconvey/convey"

	"data-model-job/common"
	"data-model-job/interfaces"
)

func MockNewKafkaAccess() *kafkaAccess {
	appSetting := &common.AppSetting{
		MQSetting: libmq.MQSetting{
			Auth: libmq.MQAuthSetting{
				Mechanism: "PLAIN",
				Username:  "test",
				Password:  "testpwd",
			},
		},
		KafkaSetting: common.KafkaSetting{
			SessionTimeoutMs:     45000,
			FetchWaitMaxMs:       299000,
			SocketTimeoutMs:      300000,
			Retries:              5,
			RetryBackoffMs:       500,
			TransactionTimeoutMs: 900000,
			MaxPollIntervalMs:    300000,
			HeartbeatIntervalMs:  10000,
		},
	}

	ka := &kafkaAccess{
		appSetting: appSetting,
	}
	return ka
}

func Test_KafkaAccess_NewKafkaAdminClient(t *testing.T) {
	Convey("create new kafkaAdminClient", t, func() {
		ka := MockNewKafkaAccess()

		Convey("1. Failed to create new adminClient", func() {
			patch := ApplyFuncReturn(kafka.NewAdminClient,
				nil, kafka.NewError(0, "Failed to create new adminClient", true))
			defer patch.Reset()

			a, err := newKafkaAdminClient(ka.appSetting)
			So(a, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("2. Successfully create adminClient", func() {
			patch := ApplyFuncReturn(kafka.NewAdminClient, &kafka.AdminClient{}, nil)
			defer patch.Reset()

			a, err := newKafkaAdminClient(ka.appSetting)
			So(a, ShouldNotBeNil)
			So(err, ShouldBeNil)
		})
	})
}

func Test_KafkaAccess_NewConsumer(t *testing.T) {
	Convey("Test KafkaAccess NewConsumer", t, func() {
		ka := MockNewKafkaAccess()
		groupID := "test"

		Convey("NewConsumer failed", func() {
			patches := ApplyFuncReturn(kafka.NewConsumer, nil, errors.New("error"))
			defer patches.Reset()

			_, err := ka.NewConsumer(groupID)
			So(err, ShouldNotBeNil)
		})

		Convey("success", func() {
			c := &kafka.Consumer{}
			patches := ApplyFuncReturn(kafka.NewConsumer, c, nil)
			defer patches.Reset()

			_, err := ka.NewConsumer(groupID)
			So(err, ShouldBeNil)
		})
	})
}

func Test_KafkaAccess_NewTransactionalProducer(t *testing.T) {
	Convey("Test KafkaAccess NewTransactionalProducer", t, func() {
		ka := MockNewKafkaAccess()
		txId := "test"

		Convey("NewTransactionalProducer failed", func() {
			patches := ApplyFuncReturn(kafka.NewProducer, nil, errors.New("error"))
			defer patches.Reset()

			_, err := ka.NewTransactionalProducer(txId)
			So(err, ShouldNotBeNil)
		})

		Convey("success", func() {
			p := &kafka.Producer{}
			patches := ApplyFuncReturn(kafka.NewProducer, p, nil)
			defer patches.Reset()

			_, err := ka.NewTransactionalProducer(txId)
			So(err, ShouldBeNil)
		})
	})
}

func Test_KafkaAccess_DoConsume(t *testing.T) {
	Convey("Test KafkaAccess DoConsume", t, func() {
		ka := MockNewKafkaAccess()
		c := &kafka.Consumer{}

		Convey("DoConsume timeout", func() {
			patches := ApplyMethodReturn(c, "Poll", nil)
			defer patches.Reset()

			msg, err := ka.DoConsume(c)
			So(msg, ShouldBeNil)
			So(err, ShouldBeNil)
		})

		Convey("event kafka.Message", func() {
			patches := ApplyMethodReturn(c, "Poll", &kafka.Message{})
			defer patches.Reset()

			msg, err := ka.DoConsume(c)
			So(msg, ShouldNotBeNil)
			So(err, ShouldBeNil)
		})

		Convey("event kafka.Error", func() {
			patches := ApplyMethodReturn(c, "Poll", kafka.Error{})
			defer patches.Reset()

			msg, err := ka.DoConsume(c)
			So(msg, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("event error type", func() {
			patches := ApplyMethodReturn(c, "Poll", c)
			defer patches.Reset()

			msg, err := ka.DoConsume(c)
			So(msg, ShouldBeNil)
			So(err, ShouldBeNil)
		})
	})
}

func Test_KafkaAccess_DoProduce(t *testing.T) {
	Convey("batch produce", t, func() {
		ka := MockNewKafkaAccess()

		p := &kafka.Producer{}
		kp := &interfaces.KafkaProducer{
			Producer:     p,
			DeliveryChan: make(chan kafka.Event, 10),
		}
		c := &kafka.Consumer{}
		sinkTopic, sinkPar := ".mdl.dr.clusterId_x", int32(0)

		records := []*kafka.Message{
			{Value: []byte("hello")},
		}

		message := &kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &sinkTopic,
				Partition: sinkPar,
				Offset:    kafka.Offset(123456),
			},
			Value: []byte("hello"),
		}

		Convey("failed to begin transaction", func() {
			patch1 := ApplyMethodReturn(p, "BeginTransaction", kafka.NewError(0, "Failed to begin transaction", true))
			defer patch1.Reset()

			err := ka.DoProduce(kp, c, records)
			So(err, ShouldNotBeNil)
		})

		Convey("failed to send message", func() {
			patch1 := ApplyMethodReturn(p, "BeginTransaction", nil)
			defer patch1.Reset()

			patch2 := ApplyMethodReturn(p, "Produce", kafka.NewError(0, "Failed to send message", true))
			defer patch2.Reset()

			err := ka.DoProduce(kp, c, records)
			So(err, ShouldNotBeNil)
		})

		Convey("failed to get assignment", func() {
			patch1 := ApplyMethodReturn(p, "BeginTransaction", nil)
			defer patch1.Reset()

			patch2 := ApplyMethod(p, "Produce",
				func(_ *kafka.Producer, _ *kafka.Message, deliveryChan chan kafka.Event) error {
					go func() {
						deliveryChan <- message
					}()

					return nil
				},
			)
			defer patch2.Reset()

			patch3 := ApplyMethodReturn(p, "Flush", 0)
			defer patch3.Reset()

			patch4 := ApplyMethodReturn(c, "Assignment", nil, kafka.NewError(0, "Failed to get consumer group metadata", true))
			defer patch4.Reset()

			err := ka.DoProduce(kp, c, records)

			So(err, ShouldNotBeNil)
		})

		Convey("failed to get consumer position", func() {
			patch1 := ApplyMethodReturn(p, "BeginTransaction", nil)
			defer patch1.Reset()

			patch2 := ApplyMethod(p, "Produce",
				func(_ *kafka.Producer, _ *kafka.Message, deliveryChan chan kafka.Event) error {
					go func() {
						deliveryChan <- message
					}()

					return nil
				},
			)
			defer patch2.Reset()

			patch3 := ApplyMethodReturn(p, "Flush", 0)
			defer patch3.Reset()

			patch4 := ApplyMethodReturn(c, "Assignment", nil, nil)
			defer patch4.Reset()

			patch5 := ApplyMethodReturn(c, "Position", nil, kafka.NewError(0, "Failed to get consumer group metadata", true))
			defer patch5.Reset()

			err := ka.DoProduce(kp, c, records)

			So(err, ShouldNotBeNil)
		})

		Convey("failed to get consumer group metadata", func() {
			patch1 := ApplyMethodReturn(p, "BeginTransaction", nil)
			defer patch1.Reset()

			patch2 := ApplyMethod(p, "Produce",
				func(_ *kafka.Producer, _ *kafka.Message, deliveryChan chan kafka.Event) error {
					go func() {
						deliveryChan <- message
					}()

					return nil
				},
			)

			defer patch2.Reset()

			patch3 := ApplyMethodReturn(p, "Flush", 0)
			defer patch3.Reset()

			patch4 := ApplyMethodReturn(c, "Assignment", nil, nil)
			defer patch4.Reset()

			patch5 := ApplyMethodReturn(c, "Position", nil, nil)
			defer patch5.Reset()

			patch6 := ApplyMethodReturn(c, "GetConsumerGroupMetadata",
				&kafka.ConsumerGroupMetadata{}, kafka.NewError(0, "Failed to get consumer group metadata", true))
			defer patch6.Reset()

			err := ka.DoProduce(kp, c, records)

			So(err, ShouldNotBeNil)
		})

		Convey("failed to send offsets to transaction", func() {
			patch1 := ApplyMethodReturn(p, "BeginTransaction", nil)
			defer patch1.Reset()

			patch2 := ApplyMethod(p, "Produce",
				func(_ *kafka.Producer, _ *kafka.Message, deliveryChan chan kafka.Event) error {
					go func() {
						deliveryChan <- message
					}()

					return nil
				},
			)
			defer patch2.Reset()

			patch3 := ApplyMethodReturn(p, "Flush", 0)
			defer patch3.Reset()

			patch4 := ApplyMethodReturn(c, "Assignment", nil, nil)
			defer patch4.Reset()

			patch5 := ApplyMethodReturn(c, "Position", nil, nil)
			defer patch5.Reset()

			patch6 := ApplyMethodReturn(c, "GetConsumerGroupMetadata", &kafka.ConsumerGroupMetadata{}, nil)
			defer patch6.Reset()

			patch7 := ApplyMethodReturn(p, "SendOffsetsToTransaction",
				kafka.NewError(0, "Failed to send offsets to transaction", true))
			defer patch7.Reset()

			err := ka.DoProduce(kp, c, records)

			So(err, ShouldNotBeNil)
		})

		Convey("failed to commit transaction", func() {
			patch1 := ApplyMethodReturn(p, "BeginTransaction", nil)
			defer patch1.Reset()

			patch2 := ApplyMethod(p, "Produce",
				func(_ *kafka.Producer, _ *kafka.Message, deliveryChan chan kafka.Event) error {
					go func() {
						deliveryChan <- message
					}()

					return nil
				},
			)
			defer patch2.Reset()

			patch3 := ApplyMethodReturn(p, "Flush", 0)
			defer patch3.Reset()

			patch4 := ApplyMethodReturn(c, "Assignment", nil, nil)
			defer patch4.Reset()

			patch5 := ApplyMethodReturn(c, "Position", nil, nil)
			defer patch5.Reset()

			patch6 := ApplyMethodReturn(c, "GetConsumerGroupMetadata", &kafka.ConsumerGroupMetadata{}, nil)
			defer patch6.Reset()

			patch7 := ApplyMethodReturn(p, "SendOffsetsToTransaction", nil)
			defer patch7.Reset()

			patch8 := ApplyMethodReturn(p, "CommitTransaction", kafka.NewError(0, "Failed to commit transaction", true))
			defer patch8.Reset()

			err := ka.DoProduce(kp, c, records)

			So(err, ShouldNotBeNil)
		})

		Convey("succeed to batch produce", func() {
			patch1 := ApplyMethodReturn(p, "BeginTransaction", nil)
			defer patch1.Reset()

			patch2 := ApplyMethod(p, "Produce",
				func(_ *kafka.Producer, _ *kafka.Message, deliveryChan chan kafka.Event) error {
					go func() {
						deliveryChan <- message
					}()

					return nil
				},
			)
			defer patch2.Reset()

			patch3 := ApplyMethodReturn(p, "Flush", 0)
			defer patch3.Reset()

			patch4 := ApplyMethodReturn(c, "Assignment", nil, nil)
			defer patch4.Reset()

			patch5 := ApplyMethodReturn(c, "Position", nil, nil)
			defer patch5.Reset()

			patch6 := ApplyMethodReturn(c, "GetConsumerGroupMetadata", &kafka.ConsumerGroupMetadata{}, nil)
			defer patch6.Reset()

			patch7 := ApplyMethodReturn(p, "SendOffsetsToTransaction", nil)
			defer patch7.Reset()

			patch8 := ApplyMethodReturn(p, "CommitTransaction", nil)
			defer patch8.Reset()

			err := ka.DoProduce(kp, c, records)

			So(err, ShouldBeNil)
		})
	})
}

func Test_KafkaAccess_DescribeTopics(t *testing.T) {
	Convey("describe Topics", t, func() {
		ka := MockNewKafkaAccess()
		client := &kafka.AdminClient{}
		topic := []string{"test"}

		Convey("failed to describe topics", func() {
			patch := ApplyMethodReturn(client, "DescribeTopics",
				kafka.DescribeTopicsResult{}, errors.New("some error"))
			defer patch.Reset()

			_, err := ka.DescribeTopics(testCtx, topic)
			So(err, ShouldNotBeNil)
		})

		Convey("Successfully to describe topics", func() {
			results := kafka.DescribeTopicsResult{
				TopicDescriptions: []kafka.TopicDescription{
					{Name: "topic1"},
					{Name: "topic2"},
				},
			}
			patch := ApplyMethodReturn(client, "DescribeTopics", results, nil)
			defer patch.Reset()

			_, err := ka.DescribeTopics(testCtx, topic)
			So(err, ShouldBeNil)
		})
	})
}

func Test_KafkaAccess_CreateTopic(t *testing.T) {
	Convey("create Topics", t, func() {
		ka := MockNewKafkaAccess()
		client := &kafka.AdminClient{}

		Convey("Failed to create kafka adminClient", func() {
			patch1 := ApplyFuncReturn(kafka.ResourceTypeFromString,
				kafka.ResourceType(kafka.ResourceBroker), kafka.NewError(0, "Failed to create topic", true))
			defer patch1.Reset()

			patch2 := ApplyMethodReturn(client, "GetMetadata",
				&kafka.Metadata{}, kafka.NewError(0, "Failed to get metadata", true))
			defer patch2.Reset()

			patch3 := ApplyMethodReturn(client, "CreateTopics", nil, kafka.NewError(0, "Failed to create topic", true))
			defer patch3.Reset()

			topic := interfaces.TopicMetadata{
				TopicName:       "bXk",
				PartitionsCount: 1,
			}

			err := ka.CreateTopicOrPartition(testCtx, topic)
			So(err, ShouldNotBeNil)
		})

		Convey("Successfully to create topics", func() {
			const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
			longTopic := make([]byte, 202)
			for i := range longTopic {
				longTopic[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
			}
			srcPar := int32(0)
			errTopic := "mdl.error.bXk"
			processTopic := "mdl.process.bXk"
			md := &kafka.Metadata{
				Brokers: []kafka.BrokerMetadata{{
					ID:   1,
					Host: "",
					Port: 8080,
				}},
				Topics: map[string]kafka.TopicMetadata{
					errTopic:     {Topic: errTopic, Partitions: []kafka.PartitionMetadata{}},
					processTopic: {Topic: processTopic, Partitions: []kafka.PartitionMetadata{{ID: srcPar}}},
				}}

			patch1 := ApplyFuncReturn(kafka.NewAdminClient, client, nil)
			defer patch1.Reset()

			patch2 := ApplyMethodReturn(client, "GetMetadata", md, nil)
			defer patch2.Reset()

			results := kafka.ConfigResourceResult{
				Type:  kafka.ResourceType(kafka.ResourceBroker),
				Name:  "1001",
				Error: kafka.Error{},
				Config: map[string]kafka.ConfigEntryResult{
					"default.replication.factor": {
						Name:       "default.replication.factor",
						Value:      "1",
						IsReadOnly: true,
						IsDefault:  false,
					},
					"num.partitions": {
						Name:       "num.partitions",
						Value:      "1",
						IsReadOnly: true,
						IsDefault:  false,
					},
				},
			}
			patch3 := ApplyMethodReturn(client, "DescribeConfigs", []kafka.ConfigResourceResult{results}, nil)
			defer patch3.Reset()

			resultsCreate := []kafka.TopicResult{
				{
					Topic: processTopic,
					Error: kafka.Error{},
				},
				{
					Topic: errTopic,
					Error: kafka.Error{},
				},
			}
			patch4 := ApplyMethodReturn(client, "CreateTopics", resultsCreate, nil)
			defer patch4.Reset()

			topic := interfaces.TopicMetadata{
				TopicName:       "bXk",
				PartitionsCount: 1,
			}

			err := ka.CreateTopicOrPartition(testCtx, topic)
			So(err, ShouldBeNil)
		})
	})
}

func Test_KafkaAccess_DeleteTopic(t *testing.T) {
	Convey("delete Topics", t, func() {
		ka := MockNewKafkaAccess()
		client := &kafka.AdminClient{}

		errTopic := "mdl.error.bXk"
		processTopic := "mdl.process.bXk"

		Convey("failed to delete topics", func() {
			topic := []string{errTopic, processTopic}
			patch := ApplyMethodReturn(client, "DeleteTopics", nil, errors.New("some error"))
			defer patch.Reset()

			err := ka.DeleteTopic(testCtx, topic)
			So(err, ShouldNotBeNil)
		})

		Convey("Successfully to delete topics", func() {
			resultsDelete := []kafka.TopicResult{
				{
					Topic: processTopic,
					Error: kafka.Error{},
				},
				{
					Topic: errTopic,
					Error: kafka.Error{},
				},
			}

			patch := ApplyMethodReturn(client, "DeleteTopics", resultsDelete, nil)
			defer patch.Reset()

			topic := []string{errTopic, processTopic}
			err := ka.DeleteTopic(testCtx, topic)
			So(err, ShouldBeNil)
		})
	})
}

func Test_KafkaAccess_DeleteConsumerGroups(t *testing.T) {
	Convey("delete consumerGroups", t, func() {
		ka := MockNewKafkaAccess()
		client := &kafka.AdminClient{}
		groups := []string{"test"}

		Convey("failed to delete consumerGroups", func() {
			patch := ApplyMethodReturn(client, "DeleteConsumerGroups",
				kafka.DeleteConsumerGroupsResult{}, errors.New("some error"))
			defer patch.Reset()

			err := ka.DeleteConsumerGroups(groups)
			So(err, ShouldNotBeNil)
		})

		Convey("Successfully to delete consumerGroups", func() {
			patch := ApplyMethodReturn(client, "DeleteConsumerGroups", kafka.DeleteConsumerGroupsResult{}, nil)
			defer patch.Reset()

			err := ka.DeleteConsumerGroups(groups)
			So(err, ShouldBeNil)
		})
	})
}

func Test_KafkaAccess_GetConsumerPosition(t *testing.T) {
	Convey("get getConsumerPosition", t, func() {
		ka := MockNewKafkaAccess()

		c := &kafka.Consumer{}
		pars := []kafka.TopicPartition{
			{Partition: 0},
		}

		Convey("failed to get consumer position", func() {
			patch := ApplyMethodReturn(c, "Position", nil, errors.New("some error"))
			defer patch.Reset()

			_, err := ka.getConsumerPosition(c, pars)
			So(err, ShouldNotBeNil)
		})

		Convey("Successfully to get consumer position", func() {
			patch := ApplyMethodReturn(c, "Position", nil, nil)
			defer patch.Reset()

			_, err := ka.getConsumerPosition(c, pars)
			So(err, ShouldBeNil)
		})
	})
}

func Test_KafkaAccess_CalculateReplicationFactor(t *testing.T) {
	Convey("set replicationFactor", t, func() {
		ka := MockNewKafkaAccess()

		Convey("Failed, the num of kafka brokers must greater than 0", func() {
			brokersCount := 0
			replicationFactor, err := ka.calculateReplicationFactor(brokersCount)
			So(replicationFactor, ShouldEqual, 0)
			So(err, ShouldNotBeNil)
		})

		Convey("Success, Single-node cluster", func() {
			brokersCount := 1
			replicationFactor, err := ka.calculateReplicationFactor(brokersCount)
			So(replicationFactor, ShouldEqual, 1)
			So(err, ShouldBeNil)
		})

		Convey("Success, Two-node cluster", func() {
			brokersCount := 2
			replicationFactor, err := ka.calculateReplicationFactor(brokersCount)
			So(replicationFactor, ShouldEqual, 1)
			So(err, ShouldBeNil)
		})

		Convey("Success, Three-node cluster", func() {
			brokersCount := 3
			replicationFactor, err := ka.calculateReplicationFactor(brokersCount)
			So(replicationFactor, ShouldEqual, 2)
			So(err, ShouldBeNil)
		})

		Convey("Success, Four-node cluster", func() {
			brokersCount := 4
			replicationFactor, err := ka.calculateReplicationFactor(brokersCount)
			So(replicationFactor, ShouldEqual, 2)
			So(err, ShouldBeNil)
		})

		Convey("Success, Five-node cluster", func() {
			brokersCount := 5
			replicationFactor, err := ka.calculateReplicationFactor(brokersCount)
			So(replicationFactor, ShouldEqual, 3)
			So(err, ShouldBeNil)
		})

		Convey("Success, Six-node cluster", func() {
			brokersCount := 6
			replicationFactor, err := ka.calculateReplicationFactor(brokersCount)
			So(replicationFactor, ShouldEqual, 3)
			So(err, ShouldBeNil)
		})
	})
}

func Test_KafkaAccess_NewTrxProducer(t *testing.T) {
	Convey("create transactional producer", t, func() {
		ka := MockNewKafkaAccess()
		producer := &kafka.Producer{}

		Convey("1. failed to create txn producer", func() {
			patchF1 := ApplyFuncReturn(kafka.NewProducer, nil,
				kafka.NewError(0, "Failed to create txn producer", true))
			defer patchF1.Reset()

			p, err := ka.NewTrxProducer("123")
			So(p, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("2. failed to init transactions", func() {
			patchF1 := ApplyFuncReturn(kafka.NewProducer, &kafka.Producer{}, nil)
			defer patchF1.Reset()

			patchF4 := ApplyMethodReturn(producer, "InitTransactions",
				kafka.NewError(0, "Failed to init transactions", true))
			defer patchF4.Reset()

			p, err := ka.NewTrxProducer("1234")
			So(p, ShouldNotBeNil)
			So(err, ShouldResemble, kafka.NewError(0, "Failed to init transactions", true))
		})

		Convey("3. success", func() {
			patchF1 := ApplyFuncReturn(kafka.NewProducer, &kafka.Producer{}, nil)
			defer patchF1.Reset()

			patchF4 := ApplyMethodReturn(producer, "InitTransactions", nil)
			defer patchF4.Reset()

			p, err := ka.NewTrxProducer("1234")
			So(p, ShouldNotBeNil)
			So(err, ShouldBeNil)
		})
	})
}

func Test_KafkaAccess_DoProduceMsgToKafka(t *testing.T) {
	Convey("Test DoProduceMsgToKafka", t, func() {
		ka := MockNewKafkaAccess()
		producer := &kafka.Producer{}

		topic := "topic"
		msg := &kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: kafka.PartitionAny,
			},
			Value: []byte{},
		}

		Convey("1. no records", func() {
			err := ka.DoProduceMsgToKafka(producer, []*kafka.Message{})
			So(err, ShouldBeNil)
		})

		Convey("2. failed to begin transaction", func() {
			patchF1 := ApplyMethodReturn(producer, "BeginTransaction",
				kafka.NewError(0, "Failed to begin transaction", true))
			defer patchF1.Reset()

			err := ka.DoProduceMsgToKafka(producer, []*kafka.Message{msg})
			So(err, ShouldNotBeNil)
		})

		Convey("3. failed to send message", func() {
			patchF1 := ApplyMethodReturn(producer, "BeginTransaction", nil)
			defer patchF1.Reset()

			patchF2 := ApplyMethodReturn(producer, "Produce",
				kafka.NewError(0, "Failed to send message", true))
			defer patchF2.Reset()

			err := ka.DoProduceMsgToKafka(producer, []*kafka.Message{msg})
			So(err, ShouldNotBeNil)
		})

		Convey("4. failed to deliveryChan error", func() {
			patchF1 := ApplyMethodReturn(producer, "BeginTransaction", nil)
			defer patchF1.Reset()

			patchF2 := ApplyMethod(producer, "Produce",
				func(_ *kafka.Producer, _ *kafka.Message, deliveryChan chan kafka.Event) error {
					go func() {
						deliveryChan <- &kafka.Message{
							TopicPartition: kafka.TopicPartition{
								Topic:     &topic,
								Partition: kafka.PartitionAny,
								Error:     kafka.NewError(0, "Delivery failed", true),
							},
							Value: []byte{},
						}
					}()

					return nil
				},
			)
			defer patchF2.Reset()

			patchF3 := ApplyMethodReturn(producer, "Flush", 0)
			defer patchF3.Reset()

			err := ka.DoProduceMsgToKafka(producer, []*kafka.Message{msg})
			So(err, ShouldNotBeNil)
		})

		Convey("5. failed to commit transaction", func() {
			patchF1 := ApplyMethodReturn(producer, "BeginTransaction", nil)
			defer patchF1.Reset()

			patchF2 := ApplyMethod(producer, "Produce",
				func(_ *kafka.Producer, _ *kafka.Message, deliveryChan chan kafka.Event) error {
					go func() {
						deliveryChan <- msg
					}()

					return nil
				},
			)
			defer patchF2.Reset()

			patchF3 := ApplyMethodReturn(producer, "Flush", 0)
			defer patchF3.Reset()

			patchF6 := ApplyMethodReturn(producer, "CommitTransaction",
				kafka.NewError(0, "Failed to commit transaction", true))
			defer patchF6.Reset()

			err := ka.DoProduceMsgToKafka(producer, []*kafka.Message{msg})
			So(err, ShouldNotBeNil)
		})

		Convey("6. failed to commit transaction TxnRequiresAbort", func() {
			patchF1 := ApplyMethodReturn(producer, "BeginTransaction", nil)
			defer patchF1.Reset()

			patchF2 := ApplyMethod(producer, "Produce",
				func(_ *kafka.Producer, _ *kafka.Message, deliveryChan chan kafka.Event) error {
					go func() {
						deliveryChan <- msg
					}()

					return nil
				},
			)
			defer patchF2.Reset()

			patchF3 := ApplyMethodReturn(producer, "Flush", 0)
			defer patchF3.Reset()

			patchF6 := ApplyMethodReturn(producer, "CommitTransaction",
				kafka.NewError(0, "Failed to commit transaction", true))
			defer patchF6.Reset()

			err := ka.DoProduceMsgToKafka(producer, []*kafka.Message{msg})
			So(err, ShouldNotBeNil)
		})

		Convey("6. success", func() {
			patchF1 := ApplyMethodReturn(producer, "BeginTransaction", nil)
			defer patchF1.Reset()

			patchF2 := ApplyMethod(producer, "Produce",
				func(_ *kafka.Producer, _ *kafka.Message, deliveryChan chan kafka.Event) error {
					go func() {
						deliveryChan <- msg
					}()

					return nil
				},
			)
			defer patchF2.Reset()

			patchF3 := ApplyMethodReturn(producer, "Flush", 0)
			defer patchF3.Reset()

			patchF6 := ApplyMethodReturn(producer, "CommitTransaction", nil)
			defer patchF6.Reset()

			err := ka.DoProduceMsgToKafka(producer, []*kafka.Message{msg})
			So(err, ShouldBeNil)
		})

	})
}

func Test_KafkaAccess_HandleTxnError(t *testing.T) {
	Convey("handleTxnError", t, func() {
		ka := MockNewKafkaAccess()
		producer := &kafka.Producer{}
		err := kafka.NewError(0, "", true)

		Convey("1. failed to abort transaction", func() {
			patchF1 := ApplyMethodReturn(err, "TxnRequiresAbort", true)
			defer patchF1.Reset()

			patchF2 := ApplyMethodReturn(producer, "AbortTransaction",
				kafka.NewError(0, " Failed to abort transaction", true))
			defer patchF2.Reset()

			err := ka.handleTxnError(err, producer)
			So(err, ShouldNotBeNil)
		})

		Convey("2. no transaction in progress, failed to abort", func() {
			patchF1 := ApplyMethodReturn(err, "TxnRequiresAbort", true)
			defer patchF1.Reset()

			patchF2 := ApplyMethodReturn(producer, "AbortTransaction",
				kafka.NewError(kafka.ErrState, "No transaction in progress", false))
			defer patchF2.Reset()

			err := ka.handleTxnError(err, producer)
			So(err, ShouldBeNil)
		})

		Convey("3. err.TxnRequiresAbort is false", func() {
			patchF1 := ApplyMethodReturn(err, "TxnRequiresAbort", false)
			defer patchF1.Reset()

			err := ka.handleTxnError(err, producer)
			So(err, ShouldNotBeNil)
		})
	})
}
