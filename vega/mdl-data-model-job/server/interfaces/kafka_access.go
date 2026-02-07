// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

const (
	// poll 阻塞时间
	POLL_TIMEOUT_MS = 100
	// 生产者生产消息阻塞的时间
	PRODUCE_FLUSH_TIMEOUT_MS = 100

	// 设置为 1, 确保某一时刻只能发送一个请求, 避免因为 retry 导致的消息乱序
	MAX_IN_FLIGHT_REQUESTS_PER_CONNECTION = 1

	// 限制topic最大长度: 249-len("<tenant>.mdl.dr.<clusterID>.customer."),tenant最长10位, kafka本身限制长度为249个字符,clusterID为22个字符
	SRC_TOPIC_MAX_LENGTH = 200

	// 允许的最大消息大小(byte)
	MAX_MESSAGE_BYTES = 20971520

	// kafka消息留存时间, 8小时
	RETENTION_MS = "28800000"
	// kafka消息单个分区的留存大小，100M
	RETENTION_BYTES = "104857600"
)

// 生产者结构体
type KafkaProducer struct {
	Producer     *kafka.Producer
	DeliveryChan chan kafka.Event
}

type TopicMetadata struct {
	TopicName       string
	PartitionsCount int
}

//go:generate mockgen -source ../interfaces/kafka_access.go -destination ../interfaces/mock/mock_kafka_access.go
type KafkaAccess interface {
	NewConsumer(groupID string) (*kafka.Consumer, error)
	NewTransactionalProducer(txId string) (*kafka.Producer, error)
	DoConsume(c *kafka.Consumer) (*kafka.Message, error)
	DoProduce(kp *KafkaProducer, c *kafka.Consumer, msgs []*kafka.Message) error
	DescribeTopics(ctx context.Context, topics []string) ([]TopicMetadata, error)
	CreateTopicOrPartition(ctx context.Context, topic TopicMetadata) error
	DeleteTopic(ctx context.Context, topicNames []string) error
	DeleteConsumerGroups(groupIds []string) error

	NewTrxProducer(uniqId string) (p *kafka.Producer, err error)
	DoProduceMsgToKafka(p *kafka.Producer, messages []*kafka.Message) error
}
