// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package drivenadapters

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"uniquery/common"
	"uniquery/interfaces"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
)

var (
	kAccessOnce sync.Once
	kAccess     interfaces.KafkaAccess
)

type kafkaAccess struct {
	appSetting *common.AppSetting
}

func NewKafkaAccess(appSetting *common.AppSetting) interfaces.KafkaAccess {
	kAccessOnce.Do(func() {
		kAccess = &kafkaAccess{
			appSetting: appSetting,
		}
	})

	return kAccess
}

// 创建消费者,消费者组ID指定为集群ID
func (ka *kafkaAccess) NewKafkaConsumer() (consumer *kafka.Consumer, err error) {
	var config kafka.ConfigMap
	//TODO: 应该统一用一个消费者组
	groupID := fmt.Sprintf("%s.mdl.event_subscribe", ka.appSetting.MQSetting.Tenant)
	config = kafka.ConfigMap{
		"bootstrap.servers":  fmt.Sprint(ka.appSetting.MQSetting.MQHost) + ":" + fmt.Sprint(ka.appSetting.MQSetting.MQPort),
		"group.id":           groupID,
		"enable.auto.commit": false,
		// "session.timeout.ms":        ka.appSetting.KafkaSetting.SessionTimeoutMs,
		// "socket.timeout.ms":         ka.appSetting.KafkaSetting.SocketTimeoutMs,
		"socket.keepalive.enable": true,
		// "heartbeat.interval.ms":     ka.appSetting.KafkaSetting.HeartbeatIntervalMs,
		"auto.offset.reset":    "latest",
		"enable.partition.eof": true,
		"max.poll.interval.ms": ka.appSetting.KafkaSetting.MaxPollIntervalMs,
		// "max.poll.records":          ka.appSetting.KafkaSetting.MaxPollRecords,
		"max.partition.fetch.bytes": interfaces.MAX_MESSAGE_BYTES,
		"fetch.wait.max.ms":         ka.appSetting.KafkaSetting.FetchWaitMaxMs,

		"security.protocol": "sasl_plaintext",
		"sasl.mechanism":    ka.appSetting.MQSetting.Auth.Mechanism,
		"sasl.username":     ka.appSetting.MQSetting.Auth.Username,
		"sasl.password":     ka.appSetting.MQSetting.Auth.Password,

		"enable.ssl.certificate.verification": false,
	}

	consumer, err = kafka.NewConsumer(&config)
	if err != nil {
		logger.Errorf("Failed to create consumer: %s", err.Error())
		return nil, err
	}
	logger.Infof("create  %v on cluster", consumer)
	return consumer, nil
}

// CreateTransactionalProducer 创建事务生产者
func (ka *kafkaAccess) NewTrxProducer(uniqId string) (p *kafka.Producer, err error) {

	producerConfig := kafka.ConfigMap{
		"bootstrap.servers":                     fmt.Sprint(ka.appSetting.MQSetting.MQHost) + ":" + fmt.Sprint(ka.appSetting.MQSetting.MQPort),
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

		"security.protocol": "sasl_plaintext",
		"sasl.mechanism":    ka.appSetting.MQSetting.Auth.Mechanism,
		"sasl.username":     ka.appSetting.MQSetting.Auth.Username,
		"sasl.password":     ka.appSetting.MQSetting.Auth.Password,

		"enable.ssl.certificate.verification": false,
	}

	p, err = kafka.NewProducer(&producerConfig)
	if err != nil {
		return p, err
	}

	txnOperateTimeout := time.Duration(ka.appSetting.KafkaSetting.TransactionOperationTimeoutMs) * time.Millisecond
	txnCtx, cancel := context.WithTimeout(context.Background(), txnOperateTimeout)
	defer cancel()
	// 初始化事务
	err = p.InitTransactions(txnCtx)
	// err = ka.InitTransactions(txnCtx, p)
	if err != nil {
		logger.Errorf("job %d failed to initTransaction, %s", uniqId, err.Error())
		return p, err
	}

	logger.Debugf("kafkaAccess created producer %s", p)
	return p, nil
}

func (ka *kafkaAccess) CreateTopicIfNotPresent(topics []string) error {

	config := kafka.ConfigMap{
		"bootstrap.servers":       fmt.Sprint(ka.appSetting.MQSetting.MQHost) + ":" + fmt.Sprint(ka.appSetting.MQSetting.MQPort),
		"socket.keepalive.enable": true,

		"security.protocol": "sasl_plaintext",
		"sasl.mechanism":    ka.appSetting.MQSetting.Auth.Mechanism,
		"sasl.username":     ka.appSetting.MQSetting.Auth.Username,
		"sasl.password":     ka.appSetting.MQSetting.Auth.Password,

		"enable.ssl.certificate.verification": false,
	}

	client, err := kafka.NewAdminClient(&config)

	if err != nil {
		fmt.Printf("Failed to create Admin client: %s\n", err)
		return err
	}
	defer client.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Create topics on cluster.
	// Set Admin options to wait for the operation to finish (or at most 60s)
	maxDur, _ := time.ParseDuration("60s")

	for _, topic := range topics {
		describeTopicsResult, err := client.DescribeTopics(ctx,
			kafka.NewTopicCollectionOfTopicNames([]string{topic}))
		if err != nil {
			logger.Errorf("获取topic:%s信息失败,请重试", topic)
			return err
		}

		if len(describeTopicsResult.TopicDescriptions) > 0 {
			if describeTopicsResult.TopicDescriptions[0].Error.Code() == kafka.ErrUnknownTopicOrPart {
				logger.Infof("topic:%s not exist, create topic first", topic)
				_, err = client.CreateTopics(
					ctx,
					// Multiple topics can be created simultaneously
					// by providing more TopicSpecification structs here.
					[]kafka.TopicSpecification{{
						Topic:             topic,
						NumPartitions:     1,
						ReplicationFactor: 1}},
					// Admin options
					kafka.SetAdminOperationTimeout(maxDur))
				if err != nil {
					fmt.Printf("Failed to create topic: %v\n", err)
					return err
				}
				continue
			}
			logger.Infof("topic:%s exist, skip create topic procedure", topic)

		}
	}
	return err
}

// 新建 adminClient
func (ka *kafkaAccess) NewKafkaAdminClient() (*kafka.AdminClient, error) {
	kafkaConfig := kafka.ConfigMap{
		"bootstrap.servers":        fmt.Sprint(ka.appSetting.MQSetting.MQHost) + ":" + fmt.Sprint(ka.appSetting.MQSetting.MQPort),
		"request.timeout.ms":       ka.appSetting.KafkaSetting.TransactionTimeoutMs,
		"retries":                  ka.appSetting.KafkaSetting.Retries,
		"socket.timeout.ms":        ka.appSetting.KafkaSetting.SocketTimeoutMs,
		"allow.auto.create.topics": false,

		"security.protocol": "sasl_plaintext",
		"sasl.mechanism":    ka.appSetting.MQSetting.Auth.Mechanism,
		"sasl.username":     ka.appSetting.MQSetting.Auth.Username,
		"sasl.password":     ka.appSetting.MQSetting.Auth.Password,

		"enable.ssl.certificate.verification": false,
	}

	admin, err := kafka.NewAdminClient(&kafkaConfig)
	if err != nil {
		logger.Errorf("Failed to create Admin client: %v", err)
		return nil, err
	}

	return admin, nil
}

// 批量消费消息
func (ka *kafkaAccess) PollMessages(consumer *kafka.Consumer) (records []*kafka.Message, err error) {
	for i := 0; i < ka.appSetting.KafkaSetting.BatchSize; i++ {
		ev := consumer.Poll(500)
		if ev == nil {
			// partitionEOF 后, 再 poll得到的是 nil
			break
		}
		switch event := ev.(type) {
		case *kafka.Message:
			// logger.Debugf("Message on %s", string(event.Value))
			// if event.Headers != nil {
			// 	logger.Debugf("Headers: %v", event.Headers)
			// }
			records = append(records, event)
		case kafka.Error:
			logger.Errorf("Error: %v: %v", event.Code(), event)
			// return records, kafka.NewError(event.Code(), event.Code().String(), true)
		case kafka.PartitionEOF:
			logger.Debugf("Reached %v PartitionEOF!", event)
			return records, nil
		default:
			logger.Debugf("Ignored %v", event)
		}
	}

	return records, nil
}

// DoProduce 生产一条消息
func (ka *kafkaAccess) DoProduce(producer *kafka.Producer, records []*kafka.Message) error {
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
	deliveryChan := make(chan kafka.Event, ka.appSetting.KafkaSetting.BatchSize)
	defer close(deliveryChan)

	for _, record := range records {
		err = producer.Produce(record, deliveryChan)
		if err != nil {
			logger.Errorf("Failed to send message to topic-partition: %s-%d: %s", topic, par, err.Error())
			return ka.handleTxnError(err, producer)
		}
	}
	logger.Debugf("准备Flush")
	producer.Flush(ka.appSetting.KafkaSetting.ProduceFlushTimeoutMs)

	logger.Debugf("遍历 deliveryChannel")
	for i := 0; i < len(records); i++ {
		event := <-deliveryChan
		message := event.(*kafka.Message)
		if message.TopicPartition.Error != nil {
			logger.Errorf("Delivery failed: %v", message.TopicPartition.Error)
			return message.TopicPartition.Error
		}
	}

	txnOperateTimeout := time.Duration(ka.appSetting.KafkaSetting.TransactionOperationTimeoutMs) * time.Millisecond
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
	txnOperateTimeout := time.Duration(ka.appSetting.KafkaSetting.TransactionOperationTimeoutMs) * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), txnOperateTimeout)
	defer cancel()

	if err.(kafka.Error).TxnRequiresAbort() {
		err = producer.AbortTransaction(ctx)
		logger.Infof("Abort transaction for %v", producer)
		if err != nil {
			if err.(kafka.Error).Code() == kafka.ErrState {
				logger.Debugf("No transaction in progress, ignore the error.")
				err = nil
			} else {
				logger.Errorf("Failed to abort transaction for %v: %v", producer, err)
			}
		}
	}

	return err
}

// 带超时机制的提交位移
func (ka *kafkaAccess) CommitOffset(consumer *kafka.Consumer) error {
	// 设定提交位移的超时时间
	commitOffsetTimeout := interfaces.COMMIT_OFFSET_TIMEOUT_MS * time.Millisecond
	timeoutCtx, cancel := context.WithTimeout(context.Background(), commitOffsetTimeout)
	defer cancel()

	exitChan := make(chan error, 1)

	go func() {
		// CommitOffset接口不阻塞的情况下，完成后关闭 channel
		defer close(exitChan)

		partions, err := consumer.Assignment()
		if err != nil {
			logger.Errorf("KafkaAccess CommitOffset Failed to Assignment: %s", err.Error())
			exitChan <- err
			return
		}
		pos, err := consumer.Position(partions)
		if err != nil {
			logger.Errorf("KafkaAccess CommitOffset Failed to Position: %s", err.Error())
			exitChan <- err
			return
		}
		_, err = consumer.CommitOffsets(pos)
		if err != nil {
			logger.Errorf("KafkaAccess CommitOffset Failed to CommitOffsets: %s", err.Error())
			exitChan <- err
			return
		}
		logger.Debugf("KafkaAccess CommitOffset Position %+v\n", pos)
		logger.Debugf("KafkaAccess withTimeout committed offset successfully.")

	}()

	select {
	case err := <-exitChan:
		if err != nil {
			return err // CommitOffset 返回错误
		}

		return nil // CommitOffset 正常返回
	case <-timeoutCtx.Done():
		logger.Errorf("Timed out commit offset ")
		return errors.New("CommitOffset Timeout Error") // CommitOffsetToSrcCluster timed out
	}
}
