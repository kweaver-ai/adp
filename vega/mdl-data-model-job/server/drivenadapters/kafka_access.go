// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package drivenadapters

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"

	"data-model-job/common"
	"data-model-job/interfaces"
)

var (
	kAccessOnce sync.Once
	kAccess     interfaces.KafkaAccess
)

type kafkaAccess struct {
	appSetting *common.AppSetting
	kaClient   *kafka.AdminClient
}

func NewKafkaAccess(appSetting *common.AppSetting) interfaces.KafkaAccess {
	kAccessOnce.Do(func() {
		client, err := newKafkaAdminClient(appSetting)
		if err != nil {
			panic(err)
		}

		kAccess = &kafkaAccess{
			appSetting: appSetting,
			kaClient:   client,
		}
	})

	return kAccess
}

// 新建 adminClient
func newKafkaAdminClient(appSetting *common.AppSetting) (*kafka.AdminClient, error) {
	kafkaConfig := kafka.ConfigMap{
		"bootstrap.servers":                   fmt.Sprintf("%s:%d", appSetting.MQSetting.MQHost, appSetting.MQSetting.MQPort),
		"retries":                             appSetting.KafkaSetting.Retries,
		"request.timeout.ms":                  appSetting.KafkaSetting.TransactionTimeoutMs,
		"socket.timeout.ms":                   appSetting.KafkaSetting.SocketTimeoutMs,
		"allow.auto.create.topics":            false,
		"enable.ssl.certificate.verification": false,
		"security.protocol":                   "sasl_plaintext",
		"sasl.mechanism":                      appSetting.MQSetting.Auth.Mechanism,
		"sasl.username":                       appSetting.MQSetting.Auth.Username,
		"sasl.password":                       appSetting.MQSetting.Auth.Password,
	}

	admin, err := kafka.NewAdminClient(&kafkaConfig)
	if err != nil {
		logger.Errorf("Failed to create Admin client: %v", err)
		return nil, err
	}

	return admin, nil
}

// 创建消费者
func (ka *kafkaAccess) NewConsumer(groupID string) (*kafka.Consumer, error) {
	consumerConfig := kafka.ConfigMap{
		"bootstrap.servers":                   fmt.Sprintf("%s:%d", ka.appSetting.MQSetting.MQHost, ka.appSetting.MQSetting.MQPort),
		"group.id":                            groupID,
		"enable.auto.commit":                  false,
		"auto.offset.reset":                   ka.appSetting.KafkaSetting.AutoOffsetReset,
		"isolation.level":                     "read_committed",
		"session.timeout.ms":                  ka.appSetting.KafkaSetting.SessionTimeoutMs,
		"socket.timeout.ms":                   ka.appSetting.KafkaSetting.SocketTimeoutMs,
		"socket.keepalive.enable":             true,
		"heartbeat.interval.ms":               ka.appSetting.KafkaSetting.HeartbeatIntervalMs,
		"max.poll.interval.ms":                ka.appSetting.KafkaSetting.MaxPollIntervalMs,
		"allow.auto.create.topics":            false,
		"enable.ssl.certificate.verification": false,
		"security.protocol":                   "sasl_plaintext",
		"sasl.mechanism":                      ka.appSetting.MQSetting.Auth.Mechanism,
		"sasl.username":                       ka.appSetting.MQSetting.Auth.Username,
		"sasl.password":                       ka.appSetting.MQSetting.Auth.Password,
	}

	c, err := kafka.NewConsumer(&consumerConfig)
	if err != nil {
		logger.Errorf("Failed to create consumer: %v", err)
		return nil, err
	}

	logger.Debugf("Create %s consumer %v on cluster %s", groupID, c, ka.appSetting.MQSetting.MQHost)

	return c, nil
}

// 创建事务生产者
func (ka *kafkaAccess) NewTransactionalProducer(txId string) (*kafka.Producer, error) {
	producerConfig := kafka.ConfigMap{
		"client.id":                             txId,
		"bootstrap.servers":                     fmt.Sprintf("%s:%d", ka.appSetting.MQSetting.MQHost, ka.appSetting.MQSetting.MQPort),
		"acks":                                  "all",
		"transactional.id":                      txId,
		"enable.idempotence":                    true,
		"retries":                               ka.appSetting.KafkaSetting.Retries,
		"max.in.flight.requests.per.connection": interfaces.MAX_IN_FLIGHT_REQUESTS_PER_CONNECTION,
		"message.max.bytes":                     interfaces.MAX_MESSAGE_BYTES,
		"retry.backoff.ms":                      ka.appSetting.KafkaSetting.RetryBackoffMs,
		"transaction.timeout.ms":                ka.appSetting.KafkaSetting.TransactionTimeoutMs,
		"socket.keepalive.enable":               true,
		"allow.auto.create.topics":              false,
		"enable.ssl.certificate.verification":   false,
		"security.protocol":                     "sasl_plaintext",
		"sasl.mechanism":                        ka.appSetting.MQSetting.Auth.Mechanism,
		"sasl.username":                         ka.appSetting.MQSetting.Auth.Username,
		"sasl.password":                         ka.appSetting.MQSetting.Auth.Password,
	}

	p, err := kafka.NewProducer(&producerConfig)
	if err != nil {
		logger.Errorf("Create transactional producer failed, %v", err)
		return nil, err
	}

	// Listen to all the client instance-level errors.
	// It's important to read these errors too otherwise the events channel will eventually fill up
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case kafka.Error:
				// Generic client instance-level errors, such as broker connection failures, authentication issues, etc.
				// These errors should generally be considered informational as the underlying client will automatically try to
				// recover from any errors encountered, the application does not need to take action on them.
				logger.Errorf("Kafka error, error msg: %v", ev)
			default:
				// default do nothing, won't record log
			}
		}
	}()

	logger.Debugf("Create %s producer %v on cluster %s", txId, p, ka.appSetting.MQSetting.MQHost)
	return p, nil
}

// 执行消费 返回 message 或者 err
func (ka *kafkaAccess) DoConsume(c *kafka.Consumer) (*kafka.Message, error) {
	ev := c.Poll(interfaces.POLL_TIMEOUT_MS)
	if ev == nil {
		return nil, nil
	}

	switch e := ev.(type) {
	case *kafka.Message:
		return e, nil

	case kafka.Error:
		err := errors.New(e.String())
		return nil, err

	default:
		logger.Debugf("Ignored %v", e)
	}
	return nil, nil
}

func (ka *kafkaAccess) DoProduce(kp *interfaces.KafkaProducer, c *kafka.Consumer, msgs []*kafka.Message) error {
	err := kp.Producer.BeginTransaction()
	if err != nil {
		logger.Errorf("Begin transaction failed, %v", err)
		return err
	}

	logger.Debugf("Preparing to produce %d messages", len(msgs))

	// // 生产者发送消息使用的 delivery channel
	// deliveryChan := make(chan kafka.Event, ka.appSetting.ServerSetting.FlushItems)
	// defer close(deliveryChan)

	for _, record := range msgs {
		err = kp.Producer.Produce(record, kp.DeliveryChan)
		if err != nil {
			logger.Errorf("Send message failed, %v", err)
			abortErr := ka.abortTxn(err, kp.Producer)
			if abortErr != nil {
				return abortErr
			}
			return err
		}
	}

	for i := 0; i < len(msgs); i++ {
		event, ok := <-kp.DeliveryChan
		if !ok {
			logger.Error("kafka produce message to topic fail because of delivery chan is closed")
			return fmt.Errorf("delivery chan of kafka producer is closed")
		}

		message := event.(*kafka.Message)

		if message.TopicPartition.Error != nil {
			logger.Errorf("Delivery failed: %v", message.TopicPartition.Error)
			return message.TopicPartition.Error
		}
	}

	kp.Producer.Flush(interfaces.PRODUCE_FLUSH_TIMEOUT_MS)

	// 获取当前消费者消费的分区
	partitions, err := c.Assignment()
	if err != nil {
		logger.Errorf("Get conusmer assignment failed, %v", err)
		return err
	}
	// 获取给定分区的要提交的位移
	position, err := ka.getConsumerPosition(c, partitions)
	if err != nil {
		logger.Errorf("Get conusmer position failed, %v", err)
		return err
	}
	consumerMetadata, err := c.GetConsumerGroupMetadata()
	if err != nil {
		logger.Errorf("Get consumer group metadata failed: %v", err)
		return err
	}

	txnOperateTimeout := time.Duration(ka.appSetting.KafkaSetting.TransactionTimeoutMs) * time.Millisecond
	sendOffsetsCtx, cancel := context.WithTimeout(context.Background(), txnOperateTimeout)
	defer cancel()

	err = kp.Producer.SendOffsetsToTransaction(sendOffsetsCtx, position, consumerMetadata)
	if err != nil {
		logger.Errorf("Send offsets to transaction failed, %v", err)
		abortErr := ka.abortTxn(err, kp.Producer)
		if abortErr != nil {
			return abortErr
		}
		return err
	}

	commitCtx, cancel := context.WithTimeout(context.Background(), txnOperateTimeout)
	defer cancel()

	err = kp.Producer.CommitTransaction(commitCtx)
	if err != nil {
		logger.Errorf("Committed transaction failed, %s", err)
		abortErr := ka.abortTxn(err, kp.Producer)
		if abortErr != nil {
			return abortErr
		}

		return err
	}

	return nil
}

// 批量获取topic的元信息
func (ka *kafkaAccess) DescribeTopics(ctx context.Context, topics []string) ([]interfaces.TopicMetadata, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()
	describeTopicsResult, err := ka.kaClient.DescribeTopics(ctx, kafka.NewTopicCollectionOfTopicNames(topics))
	if err != nil {
		logger.Errorf("Failed to describe topics: %v", err)
		return nil, err
	}

	metadatas := make([]interfaces.TopicMetadata, 0, len(topics))
	for _, topic := range describeTopicsResult.TopicDescriptions {
		if topic.Error.Code() != 0 {
			logger.Errorf("Describe topic %s error, %s", topic.Name, topic.Error)
			return nil, topic.Error
		}

		md := interfaces.TopicMetadata{
			TopicName:       topic.Name,
			PartitionsCount: len(topic.Partitions),
		}

		metadatas = append(metadatas, md)
	}

	return metadatas, nil

}

// 手动创建 topic和新增分区, 如果 topic 不存在，则创建; 若存在的 topic 的分区增加，则新建增加的分区
// 设置目标topic的留存周期 8 小时，留存大小500M
func (ka *kafkaAccess) CreateTopicOrPartition(ctx context.Context, topic interfaces.TopicMetadata) error {
	// 获取集群元数据, 对于topic原信息，拿的是所有topic
	metadata, err := ka.kaClient.GetMetadata(nil, true, ka.appSetting.KafkaSetting.AdminClientRequestTimeoutMs)
	if err != nil {
		logger.Errorf("Failed to get metadata, %s.", err.Error())
		return err
	}

	// 计算目标topic的副本因子
	brokersCount := len(metadata.Brokers)
	replicationFactor, err := ka.calculateReplicationFactor(brokersCount)
	if err != nil {
		return err
	}

	allTopics := metadata.Topics
	topicSpecifications := make([]kafka.TopicSpecification, 0, 1)
	partitionsSpecifications := make([]kafka.PartitionsSpecification, 0)

	if _, ok := allTopics[topic.TopicName]; !ok {
		topicSpecifications = append(topicSpecifications, kafka.TopicSpecification{
			Topic:             topic.TopicName,
			NumPartitions:     topic.PartitionsCount,
			ReplicationFactor: replicationFactor,
			Config: map[string]string{
				"retention.ms":    strconv.Itoa(ka.appSetting.KafkaSetting.RetentionMs),
				"retention.bytes": strconv.Itoa(ka.appSetting.KafkaSetting.RetentionBytes),
			},
		})
	} else if topic.PartitionsCount > len(allTopics[topic.TopicName].Partitions) {
		partitionsSpecifications = append(partitionsSpecifications, kafka.PartitionsSpecification{
			Topic:      topic.TopicName,
			IncreaseTo: topic.PartitionsCount,
		})
	}

	if len(topicSpecifications) > 0 {
		adminOperateTimeout := time.Duration(ka.appSetting.KafkaSetting.AdminClientOperationTimeoutMs) * time.Millisecond
		results, err := ka.kaClient.CreateTopics(ctx, topicSpecifications, kafka.SetAdminOperationTimeout(adminOperateTimeout))
		if err != nil {
			logger.Errorf("Failed to create topics %v: %v", topicSpecifications, err)
			return err
		}

		for _, result := range results {
			logger.Debugf("Topic %s createdResult: %v.", result.Topic, result)
		}
	}

	if len(partitionsSpecifications) > 0 {
		logger.Debugf("Create partitions %v", partitionsSpecifications)

		adminOperateTimeout := time.Duration(ka.appSetting.KafkaSetting.AdminClientOperationTimeoutMs) * time.Millisecond
		results, err := ka.kaClient.CreatePartitions(ctx, partitionsSpecifications, kafka.SetAdminOperationTimeout(adminOperateTimeout))
		if err != nil {
			logger.Errorf("Failed to create partitions %v: %v", partitionsSpecifications, err)
			return err
		}

		for _, result := range results {
			logger.Debugf("Successfully create partitions for Topic %s createdResult: %v.", result.Topic, result)
		}
	}

	return nil
}

// 删除topic
func (ka *kafkaAccess) DeleteTopic(ctx context.Context, topicNames []string) error {
	dur := time.Duration(ka.appSetting.KafkaSetting.AdminClientRequestTimeoutMs) * time.Millisecond
	results, err := ka.kaClient.DeleteTopics(ctx, topicNames, kafka.SetAdminOperationTimeout(dur))
	if err != nil {
		detail := fmt.Sprintf("Failed to delete topic %s: %s", strings.Join(topicNames, ","), err.Error())
		logger.Errorf(detail)
		return err
	}

	logger.Infof("Delete topic's result, %v", results)
	return nil
}

// 删除消费者组
func (ka *kafkaAccess) DeleteConsumerGroups(groups []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	dur := time.Duration(ka.appSetting.KafkaSetting.AdminClientRequestTimeoutMs) * time.Millisecond

	result, err := ka.kaClient.DeleteConsumerGroups(ctx, groups, kafka.SetAdminRequestTimeout(dur))
	if err != nil {
		logger.Errorf("Failed to delete groups: %s", err)
		return err
	}

	logger.Infof("Delete consumer group's result, %v", result)
	return nil
}

// 获取消费者需要提交的位移
func (ka *kafkaAccess) getConsumerPosition(c *kafka.Consumer,
	partitions []kafka.TopicPartition) ([]kafka.TopicPartition, error) {

	position, err := c.Position(partitions)
	if err != nil {
		logger.Errorf("Failed to get consumer offsets, %v", err)
		return nil, err
	}

	return position, nil
}

// 中止事务
func (ka *kafkaAccess) abortTxn(err error, producer *kafka.Producer) error {
	txnOperateTimeout := time.Duration(ka.appSetting.KafkaSetting.TransactionTimeoutMs) * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), txnOperateTimeout)
	defer cancel()

	logger.Infof("Current 'TxnRequiresAbort' result is %v", err.(kafka.Error).TxnRequiresAbort())

	if err.(kafka.Error).TxnRequiresAbort() {
		err = producer.AbortTransaction(ctx)
		logger.Infof("Abort transaction for %v", producer)
		if err != nil {
			if err.(kafka.Error).Code() == kafka.ErrState {
				logger.Infof("No transaction in progress, ignore the error")
				err = nil
			} else {
				logger.Errorf("Abort transaction for %v failed: %v", producer, err)
			}
		}
	}

	return err
}

// 确定创建的 topic 的副本数, 副本因子数不能小于 1
// 设置规则:
//
//	 kafka 应用集群数节点数建议为奇数个
//	如果节点数为 1-2, 副本数设置为 1
//	如果节点数为 3-4, 副本数设置为 2
//	如果节点数 >= 5, 副本数设置为 3
func (ka *kafkaAccess) calculateReplicationFactor(brokersCount int) (replicationFactor int, err error) {
	if brokersCount >= 5 {
		return 3, nil
	} else if brokersCount >= 3 {
		return 2, nil
	} else if brokersCount >= 1 {
		return 1, nil
	} else {
		logger.Error("The num of kafka brokers must greater than 0")
		return 0, errors.New("the num of kafka brokers must greater than 0")
	}
}

// CreateTransactionalProducer 创建事务生产者
func (ka *kafkaAccess) NewTrxProducer(uniqId string) (p *kafka.Producer, err error) {
	producerConfig := kafka.ConfigMap{
		"bootstrap.servers":                     fmt.Sprintf("%s:%d", ka.appSetting.MQSetting.MQHost, ka.appSetting.MQSetting.MQPort),
		"client.id":                             uniqId,
		"transactional.id":                      uniqId,
		"acks":                                  "all",
		"enable.idempotence":                    true,
		"retries":                               ka.appSetting.KafkaSetting.Retries,
		"max.in.flight.requests.per.connection": interfaces.MAX_IN_FLIGHT_REQUESTS_PER_CONNECTION,
		"message.max.bytes":                     interfaces.MAX_MESSAGE_BYTES,
		"retry.backoff.ms":                      ka.appSetting.KafkaSetting.RetryBackoffMs,
		"transaction.timeout.ms":                ka.appSetting.KafkaSetting.TransactionTimeoutMs,
		"socket.keepalive.enable":               true,
		"enable.ssl.certificate.verification":   false,
		"security.protocol":                     "sasl_plaintext",
		"sasl.mechanism":                        ka.appSetting.MQSetting.Auth.Mechanism,
		"sasl.username":                         ka.appSetting.MQSetting.Auth.Username,
		"sasl.password":                         ka.appSetting.MQSetting.Auth.Password,
	}

	p, err = kafka.NewProducer(&producerConfig)
	if err != nil {
		return p, err
	}

	txnOperateTimeout := time.Duration(ka.appSetting.KafkaSetting.TransactionTimeoutMs) * time.Millisecond
	txnCtx, cancel := context.WithTimeout(context.Background(), txnOperateTimeout)
	defer cancel()
	// 初始化事务
	err = p.InitTransactions(txnCtx)
	if err != nil {
		logger.Errorf("job %d failed to initTransaction, %s", uniqId, err.Error())
		return p, err
	}

	logger.Debugf("kafkaAccess created producer %s", p)
	return p, nil
}

// DoProduce 把msg推到kafka
func (ka *kafkaAccess) DoProduceMsgToKafka(producer *kafka.Producer, records []*kafka.Message) error {
	if len(records) == 0 {
		return nil
	}
	topic := records[0].TopicPartition.Topic
	par := records[0].TopicPartition.Partition

	err := producer.BeginTransaction()
	if err != nil {
		logger.Errorf("Failed to begin transaction on topic-partition: %s-%d, %v", topic, par, err)
		return ka.handleTxnError(err, producer)
	}
	logger.Debugf("Begin Transaction on topic-partition: %s-%d", topic, par)

	logger.Debugf("Preparing to produce %d messages on topic-partition: %s-%d", len(records), topic, par)
	// 生产者发送消息使用的 delivery channel
	deliveryChan := make(chan kafka.Event, len(records))
	defer close(deliveryChan)

	for _, record := range records {
		err = producer.Produce(record, deliveryChan)
		if err != nil {
			logger.Errorf("Failed to send message to topic-partition: %s-%d: %s", topic, par, err.Error())
			return ka.handleTxnError(err, producer)
		}
	}

	logger.Debugf("遍历 deliveryChannel")
	for i := 0; i < len(records); i++ {
		event := <-deliveryChan
		message := event.(*kafka.Message)
		if message.TopicPartition.Error != nil {
			logger.Errorf("Delivery failed: %v", message.TopicPartition.Error)
			return message.TopicPartition.Error
		}
	}

	logger.Debugf("准备Flush")
	producer.Flush(interfaces.PRODUCE_FLUSH_TIMEOUT_MS)

	txnOperateTimeout := time.Duration(ka.appSetting.KafkaSetting.TransactionTimeoutMs) * time.Millisecond
	commitCtx, cancel := context.WithTimeout(context.Background(), txnOperateTimeout)
	defer cancel()

	err = producer.CommitTransaction(commitCtx)
	if err != nil {
		logger.Errorf("kafkaAccess CommitTransaction Error [inputTopic: %v] [partition: %v], Error: %+v", topic, par, err)
		return ka.handleTxnError(err, producer)
	}
	logger.Debugf("Committed transaction success,%v", producer)

	return nil
}

// handleTxnError 处理事务错误
func (ka *kafkaAccess) handleTxnError(err error, producer *kafka.Producer) error {
	txnOperateTimeout := time.Duration(ka.appSetting.KafkaSetting.TransactionTimeoutMs) * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), txnOperateTimeout)
	defer cancel()

	if err.(kafka.Error).TxnRequiresAbort() {
		err = producer.AbortTransaction(ctx)
		logger.Infof("Abort transaction for %v", producer)
		if err != nil {
			if err.(kafka.Error).Code() == kafka.ErrState {
				logger.Infof("No transaction in progress, ignore the error.")
				err = nil
			} else {
				logger.Errorf("Failed to abort transaction for %v: %v", producer, err)
			}
		}
	}

	return err
}
