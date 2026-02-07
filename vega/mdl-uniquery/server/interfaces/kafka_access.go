// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "github.com/confluentinc/confluent-kafka-go/v2/kafka"

const (
	// 设置为 1, 确保某一时刻只能发送一个请求, 避免因为 retry 导致的消息乱序
	MAX_IN_FLIGHT_REQUESTS_PER_CONNECTION = 1
	// poll 阻塞时间
	POLL_TIMEOUT_MS = 500
	// 允许的最大消息大小(byte)
	MAX_MESSAGE_BYTES = 20971520
	// 向源集群提交位移的超时时间
	COMMIT_OFFSET_TIMEOUT_MS = 30000
)

//go:generate mockgen -source ../interfaces/kafka_access.go -destination ../interfaces/mock/mock_kafka_access.go
type KafkaAccess interface {
	NewKafkaConsumer() (consumer *kafka.Consumer, err error)
	NewTrxProducer(uniqId string) (p *kafka.Producer, err error)
	CreateTopicIfNotPresent(topics []string) error
	NewKafkaAdminClient() (*kafka.AdminClient, error)
	PollMessages(consumer *kafka.Consumer) (record []*kafka.Message, err error)
	DoProduce(p *kafka.Producer, messages []*kafka.Message) error
	CommitOffset(consumer *kafka.Consumer) (err error)
}
