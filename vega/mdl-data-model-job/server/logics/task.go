// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package logics

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/bytedance/sonic"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"

	"data-model-job/common"
	cond "data-model-job/common/condition"
	"data-model-job/interfaces"
)

type Task struct {
	appSetting     *common.AppSetting
	kAccess        interfaces.KafkaAccess
	jobId          string
	srcTopic       string
	sinkTopic      string
	viewCond       cond.Condition
	status         string
	RunningChannel chan bool
	*interfaces.DataView
}

func NewTask(appSetting *common.AppSetting, view *interfaces.DataView, jobId string,
	srcTopic, sinkTopic string, viewCond cond.Condition) *Task {
	return &Task{
		appSetting:     appSetting,
		kAccess:        KAccess,
		jobId:          jobId,
		srcTopic:       srcTopic,
		sinkTopic:      sinkTopic,
		viewCond:       viewCond,
		status:         interfaces.TaskStatus_Running,
		RunningChannel: make(chan bool),
		DataView:       view,
	}
}

// task 信息打印
func (t *Task) String() string {
	return fmt.Sprintf("{task_job_id = %s, task_src_topic = %s, task_sink_topic = %s, task_status = %s}",
		t.jobId, t.srcTopic, t.sinkTopic, t.status)
}

func (t *Task) Run() error {
	logger.Debugf("Start task, source topic is: %s", t.srcTopic)
	err := t.execute()
	if err != nil {
		t.status = interfaces.TaskStatus_Error
		logger.Errorf("Failed to execute task %s, error: %s", t, err.Error())
	}
	return err
}

// 停止 task
func (t *Task) Stop() error {
	switch tStatus := t.status; tStatus {
	case interfaces.TaskStatus_Stopped:
		logger.Infof("Stop a stopped task %s", t)
		return fmt.Errorf("Stop a stopped task")
	default:
		logger.Debugf("Task %s will be set to stopping status", t)
		t.status = interfaces.TaskStatus_Stopping
	}

	// 等待 executeTask 的协程结束
	<-t.RunningChannel

	t.status = interfaces.TaskStatus_Stopped
	logger.Infof("Task %s has stopped", t)
	return nil
}

func (t *Task) CloseConsumer(c *kafka.Consumer) {
	// 异步关闭consumer
	// 30s后kafka报：Connection setup timed out in state APIVERSION_QUERY (after 30000ms in state APIVERSION_QUERY)
	// go c.Close()
	// 改成同步关闭，避免消费组删除不成功
	c.Close()
}

func (t *Task) CloseProducer(kp *interfaces.KafkaProducer) {
	// flush in-flight msg within queue.
	i := kp.Producer.Flush(3000)
	if i > 0 {
		logger.Warnf("There are still %d un-flushed outstanding events on task %s", i, t)
	}

	close(kp.DeliveryChan)
	logger.Debugf("Delivery channel for producer %v closed on task %s", kp.Producer, t)

	kp.Producer.Close()
	logger.Debugf("Producer %v closed on task %s", kp.Producer, t)
}

func (t *Task) execute() error {
	// 创建消费者
	groupId := ComsumerGroupID(t.appSetting.MQSetting.Tenant, t.ViewId)
	c, err := t.kAccess.NewConsumer(groupId)
	if err != nil {
		logger.Errorf("Task Failed to NewConsumer: %v", err)
		return err
	}
	defer func() {
		t.CloseConsumer(c)
	}()

	// 生产者发送消息使用的 delivery channel
	deliveryChan := make(chan kafka.Event, t.appSetting.ServerSetting.FlushItems)

	// 创建生产者
	txId := fmt.Sprintf("%s_%s", t.ViewId, t.srcTopic)
	p, err := t.kAccess.NewTransactionalProducer(txId)
	if err != nil {
		return err
	}

	kp := &interfaces.KafkaProducer{
		Producer:     p,
		DeliveryChan: deliveryChan,
	}
	// 函数结束前关闭生产者和 delivery channel
	defer t.CloseProducer(kp)

	// 订阅topic
	err = c.Subscribe(t.srcTopic, nil)
	if err != nil {
		logger.Errorf("Task failed to Subscribe: %v", err)
		return err
	}

	// 初始化事务
	txnOperateTimeout := time.Duration(t.appSetting.KafkaSetting.TransactionTimeoutMs) * time.Millisecond
	txnCtx, cancel := context.WithTimeout(context.Background(), txnOperateTimeout)
	defer cancel()
	err = p.InitTransactions(txnCtx)
	if err != nil {
		logger.Errorf("Task failed to initTransaction on, %v", err)
		return err
	}

	lastFlushTime := time.Now()
	currentMsgs := make([]*kafka.Message, 0, FlushItems+1)
	currentBufLen := 0

	cnt := 0
	for t.status == interfaces.TaskStatus_Running {
		msg, err := t.kAccess.DoConsume(c)
		if err != nil {
			cnt++
			if cnt >= FailureThreshold {
				logger.Errorf("Task failed to DoConsume, need to restart, retry[%d]: %s", cnt, err.Error())
				return err
			}

			logger.Errorf("Task failed to DoConsume: wait to retry[%d]: %s", cnt, err.Error())
			time.Sleep(RetryInterval)
			continue
		}

		if msg != nil {
			currentMsgs = append(currentMsgs, msg)
			currentBufLen += len(msg.Value)
			// 满足批量大小或批量条数
			if currentBufLen >= FlushBytes || len(currentMsgs) >= FlushItems {
				items, err := t.packagingMessages(currentMsgs)
				if err != nil {
					return err
				}
				currentMsgs = make([]*kafka.Message, 0, FlushItems+1)
				currentBufLen = 0

				// 写入
				err = t.flushMessages(items, c, kp)
				if err != nil {
					return err
				}

				lastFlushTime = time.Now()
			}
		}

		// 满足时间间隔
		if time.Since(lastFlushTime) > FlushInterval {
			if len(currentMsgs) > 0 {
				items, err := t.packagingMessages(currentMsgs)
				if err != nil {
					return err
				}

				currentMsgs = make([]*kafka.Message, 0, FlushItems+1)
				currentBufLen = 0

				err = t.flushMessages(items, c, kp)
				if err != nil {
					return err
				}
			}

			lastFlushTime = time.Now()
		}
	}

	return nil
}

func (t *Task) packagingMessages(msgs []*kafka.Message) ([]*kafka.Message, error) {
	if len(msgs) == 0 {
		return nil, nil
	}

	items := make([]*kafka.Message, 0, len(msgs))
	mu := sync.Mutex{}

	var err error
	var wg sync.WaitGroup
	for _, msg := range msgs {
		wg.Add(1)
		msg := msg
		_ = PackagePool.Submit(func() {
			item, err1 := t.packagingMessage(msg)

			mu.Lock()
			// append 通过过滤条件的item
			if item != nil {
				items = append(items, item)
			}
			if err1 != nil {
				err = err1
			}
			mu.Unlock()

			wg.Done()
		})
	}
	wg.Wait()

	if err != nil {
		return nil, err
	}

	return items, nil
}

func (t *Task) flushMessages(msgs []*kafka.Message, c *kafka.Consumer, kp *interfaces.KafkaProducer) error {
	cnt := 0
	for {
		err := t.kAccess.DoProduce(kp, c, msgs)
		if err == nil {
			break
		}

		cnt++
		if cnt >= FailureThreshold {
			logger.Errorf("DoProduce failed, need to restart, %v", err)
			return err
		}

		logger.Errorf("DoProduce failed, wait to retry[%d]: %v", cnt, err)
		time.Sleep(RetryInterval)
	}

	return nil
}

// 过滤条件过滤数据、字段筛选
func (t *Task) packagingMessage(msg *kafka.Message) (*kafka.Message, error) {
	// 将视图id添加到kafka消息headers里
	item := msg
	headers := []kafka.Header{{Key: "__id", Value: []byte(t.ViewId)},
		{Key: "__type", Value: []byte("data_view")}}
	item.Headers = headers

	// 如果没有过滤条件并且选择全部字段, 是identity场景，则无须反序列化和序列化，直接返回
	if t.viewCond == nil && t.FieldScope == interfaces.ALL {
		item.TopicPartition.Topic = &t.sinkTopic
		item.Value = msg.Value

		return item, nil
	}

	var origin map[string]any
	err := sonic.Unmarshal(msg.Value, &origin)
	if err != nil {
		logger.Errorf("Data is not in JSON format, unmarshal msg failed: %v", err)
		return nil, nil
	}

	// 根据过滤条件过滤数据
	if t.viewCond != nil {
		data := &cond.OriginalData{
			Origin: origin,
			Output: map[string][]any{},
		}
		isPass, err := t.viewCond.Pass(context.Background(), data)
		if err != nil {
			logger.Errorf("Condition pass failed: %v", err)
			// 如果当前这条数据过滤时出错，返回nil，不影响下一条数据继续过滤
			return nil, nil
		}

		// 没通过条件的数据，返回 nil
		if !isPass {
			return nil, nil
		}
	}

	// 没有过滤条件+部分字段、有过滤条件+全部字段、有过滤条件+部分字段这三种
	// 场景 pick 为视图最终输出数据，实时订阅只支持一种输出format: original
	pick := make(map[string]any)
	if t.FieldScope == interfaces.ALL {
		pick = origin
	} else {
		err = t.pickData(origin, pick)
		if err != nil {
			return nil, err
		}
	}

	val, err := sonic.Marshal(pick)
	if err != nil {
		logger.Errorf("Marshal msg failed: %v", err)
		return nil, err
	}

	item.TopicPartition.Topic = &t.sinkTopic
	item.Value = val

	return item, nil
}

// 对数据做过滤，只包含视图字段
func (t *Task) pickData(origin, pick map[string]any) error {
	// 过滤字段， 转成pick
	for _, field := range t.Fields {
		field := field
		value, isSliceValue, err := getData(origin, field)
		if err != nil {
			return err
		}

		err = setData(field, pick, value, isSliceValue)
		if err != nil {
			return err
		}
	}

	return nil
}

func getData(origin map[string]any, field *cond.Field) ([]any, bool, error) {
	field.InitFieldPath()
	oDatas, isSliceValue, err := GetDatasByPath(origin, field.Path)
	if err != nil {
		return nil, isSliceValue, err
	}
	if len(oDatas) == 0 {
		return nil, isSliceValue, nil
	}

	return oDatas, isSliceValue, nil
}

// array 里面相同类型 可以获取内部数据，如果非相同类型，则nil
func GetDatasByPath(obj any, path []string) ([]any, bool, error) {
	if reflect.TypeOf(obj) == nil {
		return []any{}, false, nil
	}

	current := obj
	for idx := 0; idx < len(path); idx++ {
		switch reflect.TypeOf(current).Kind() {
		case reflect.Map:
			value, ok := current.(map[string]any)[path[idx]]
			if !ok || value == nil {
				return []any{}, false, nil
			}
			// found
			current = value

		case reflect.Slice:
			res := []any{}
			for _, sub := range current.([]any) {
				subRes, isSliceValue, err := GetDatasByPath(sub, path[idx:])
				if err != nil {
					return []any{}, isSliceValue, err
				}
				res = append(res, subRes...)
			}
			return res, true, nil

		default:
			// invalid path
			return []any{}, false, nil
		}
	}

	// path is empty
	return GetLastDatas(current)
}

func GetLastDatas(obj any) (res []any, isSliceValue bool, err error) {
	if obj == nil {
		return []any{}, isSliceValue, nil
	}
	switch reflect.TypeOf(obj).Kind() {
	case reflect.Slice:
		isSliceValue = true
		for _, sub := range obj.([]any) {
			subRes, isSliceValue, err := GetLastDatas(sub)
			if err != nil {
				return []any{}, isSliceValue, err
			}
			res = append(res, subRes...)
		}
	default:
		res = append(res, obj)
		return res, isSliceValue, nil
	}

	return res, isSliceValue, nil
}

// 写入k v
// 子节点不存在则构建树结构往根树并入
// 比如 {"a":{"b":"c"}} 写入 ["a", "d"] value 为 123
// 则最终为 {"a":{"b":"c","d":123}}
func setData(field *cond.Field, obj map[string]any, data []any, isSliceValue bool) error {
	if len(data) == 0 {
		return nil
	}

	current := obj
	field.InitFieldPath()
	for idx := 0; idx < len(field.Path)-1; idx++ {
		if value, ok := current[field.Path[idx]]; ok {
			if reflect.TypeOf(value).Kind() == reflect.Map {
				current = value.(map[string]any)
			} else {
				return fmt.Errorf("current path is not a map: %s", field.Path[idx])
			}
		} else {
			tmp := make(map[string]interface{})
			current[field.Path[idx]] = tmp
			current = tmp
		}
	}
	if len(data) == 1 && !isSliceValue {
		current[field.Path[len(field.Path)-1]] = data[0]
	} else {
		current[field.Path[len(field.Path)-1]] = data
	}

	return nil
}
