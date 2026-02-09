// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package event

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/panjf2000/ants/v2"
	"github.com/patrickmn/go-cache"

	"uniquery/common"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	"uniquery/logics"
)

var (
	PackagePoolSize int = 5
	PackagePool, _      = ants.NewPool(PackagePoolSize, ants.WithPreAlloc(true), ants.WithNonblocking(false))
)

var (
	eventSubOnce sync.Once
	ess          interfaces.EventSubService
)

type eventSubService struct {
	appSetting  *common.AppSetting
	engine      interfaces.EventEngine
	emAccess    interfaces.EventModelAccess
	kafkaAccess interfaces.KafkaAccess
	topics      []string
}

func NewEventSubService(appSetting *common.AppSetting) interfaces.EventSubService {
	engine = &EventEngine{
		appSetting:     appSetting,
		EventModel:     interfaces.EventModel{},
		DataSource:     interfaces.DataSource{},
		DataSourceType: "",
		DVAccess:       logics.DVAccess,
		KAccess:        logics.KAccess,
	}
	eventSubOnce.Do(func() {
		ess = &eventSubService{
			appSetting:  appSetting,
			engine:      engine,
			emAccess:    logics.EMAccess,
			kafkaAccess: logics.KAccess,
			topics:      []string{},
		}

	})
	return ess
}

func (ess *eventSubService) Subscribe(exitCh chan bool) error {
	consumer, err := ess.kafkaAccess.NewKafkaConsumer()

	if err != nil {
		logger.Errorf("NewKafkaConsumer failed:%v", err)
		return err
	}
	uniqueID := fmt.Sprintf("%s.sdp.mdl-model-persistence.input_%s", ess.appSetting.MQSetting.Tenant, common.RandStringRunes(5))

	// uniqueID := fmt.Sprintf("%d-%s-%s", task.TaskID, topic,common.RandStringRunes(5))
	producer, err1 := ess.kafkaAccess.NewTrxProducer(uniqueID)
	if err1 != nil {
		return err1
	}
	defer producer.Close()
	err = ess.DoSubscribe(consumer, producer)
	if err != nil {
		return err
	}
	return nil
}

func (ess *eventSubService) DoSubscribe(consumer *kafka.Consumer, producer *kafka.Producer) error {
	var topicChan = make(chan []string)
	var stopSignal = make(chan bool, 1)

	// initTopics, err := ess.kafkaAccess.GetInitTopics()
	topics := []string{
		fmt.Sprintf(interfaces.DEFAULT_SUBSCRIBE_TOPIC, ess.appSetting.MQSetting.Tenant),
		fmt.Sprintf(interfaces.ATOMIC_EVENT_DATA_TOPIC, ess.appSetting.MQSetting.Tenant),
	}
	err := ess.kafkaAccess.CreateTopicIfNotPresent(topics)
	if err != nil {
		logger.Errorf("create default topics list failed: %s\n", err)
		return err
	}
	initTopics, err := ess.getInitTopics()
	ess.topics = initTopics

	// initTopics = append(initTopics, defaultTopic)
	if err != nil {
		logger.Errorf("fetch topics list failed: %s\n", err)
		// initTopics = append(initTopics, topics...)
	}

	err = consumer.SubscribeTopics(initTopics, nil)
	if err != nil {
		logger.Errorf("subscribe default topics list failed: %s\n", err)
		return err
	}

	// 定期执行，然后发送到管道

	go func() {
		ess.refreshTopics(stopSignal, topicChan)
	}()

	//NOTE: 关闭资源
	defer func() {
		logger.Debugf("Consumer close success : %v", err)
		stopSignal <- true

	}()
	_cache := cache.New(interfaces.EXPIRATION_TIME, interfaces.DELETE_TIME)
	for {
		select {
		case topics := <-topicChan:
			// logger.Debugf("topic list: %+v", topics)
			// mutex.Lock()
			err := consumer.SubscribeTopics(topics, nil)
			// mutex.Unlock()
			if err != nil {
				logger.Errorf("SubscribeTopics failed:%v", err)
				return err
			}
			ess.topics = topics
			_ = ess.DoProcess(consumer, producer, _cache)

		default:
			_ = ess.DoProcess(consumer, producer, _cache)

		}
	}
}

func (ess *eventSubService) getInitTopics() ([]string, error) {

	// var topic = "default.mdl.view.*"
	// _consumer, err := ess.kafkaAccess.NewKafkaConsumer()
	kaAdmin, err := ess.kafkaAccess.NewKafkaAdminClient()
	if err != nil {
		logger.Errorf("Failed to create Admin client: %v", err)
		return []string{}, err
	}
	defer kaAdmin.Close()
	// 刷新元数据缓存
	metaData, err := kaAdmin.GetMetadata(nil, true, 5000)

	// metaData, err := _consumer.GetMetadata(&topic, true, 5000)

	if err != nil {
		logger.Errorf("Error refreshing topics metadata: %v", err)
		return []string{}, err
	}

	topicPrefix := fmt.Sprintf("%s.mdl.view.", ess.appSetting.MQSetting.Tenant)
	filteredTopics := make([]string, 0)
	for _, v := range metaData.Topics {
		if strings.HasPrefix(v.Topic, topicPrefix) {
			filteredTopics = append(filteredTopics, v.Topic)
		}
	}
	topics := []string{
		fmt.Sprintf(interfaces.DEFAULT_SUBSCRIBE_TOPIC, ess.appSetting.MQSetting.Tenant),
		fmt.Sprintf(interfaces.ATOMIC_EVENT_DATA_TOPIC, ess.appSetting.MQSetting.Tenant),
	}
	filteredTopics = append(filteredTopics, topics...)

	logger.Debugf("filteredTopics: %v\n", filteredTopics)
	return filteredTopics, nil
}

func (ess *eventSubService) refreshTopics(stopSignal chan bool, topicChan chan []string) {
	ticker := time.NewTicker(2 * time.Minute) // 定时器，每隔一分钟获取一次主题列表

	for {
		select {
		case <-ticker.C:
			// filteredTopics, err := ess.kafkaAccess.GetInitTopics()
			filteredTopics, err := ess.getInitTopics()
			if err != nil {
				logger.Errorf("fetch topics list failed: %s\n", err)
				break
			}
			//默认初始存在的监听topic有2个，如果刷新后的topic个数大于2，则说明有新topic加入,需要刷新监听列表
			if len(filteredTopics) > 2 && !reflect.DeepEqual(ess.topics, filteredTopics) {
				logger.Debugf("refresh topic list success")
				topicChan <- filteredTopics
			}

		case <-stopSignal:
			ticker.Stop()
			logger.Errorf("topic 列表监听程序退出！！")

			return
		}
	}
}

func (ess *eventSubService) DoProcess(consumer *kafka.Consumer, producer *kafka.Producer, _cache *cache.Cache) error {
	start_time := time.Now()
	// 消费订阅的所有topic的数据
	records, err := ess.kafkaAccess.PollMessages(consumer)
	//NOTE:超时事件5min，清理时间10min
	if err != nil {
		logger.Errorf("poll messages failed:%v", err)
		return err
	}
	if len(records) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	mu := sync.Mutex{}
	messages := make([]*kafka.Message, 0)
	entitiesMessages := make([]*kafka.Message, 0)

	//NOTE 按照事件模型进行分组处理。
	MapRecord := ess.divideRecordBySourceId(records)

	for dataSource, source_records := range MapRecord {

		wg.Add(1)
		_ = PackagePool.Submit(func() {
			_events, _entities, err1 := ess.Invoke(engine, dataSource.DataSourceId, dataSource.DataSourceType, source_records, _cache)
			mu.Lock()
			// append 通过过滤条件的item
			if len(_events) > 0 {
				messages = append(messages, _events...)
			}
			if len(_entities) > 0 {
				entitiesMessages = append(entitiesMessages, _entities...)
			}
			if err1 != nil {
				err = err1
			}
			mu.Unlock()
			wg.Done()
		})
	}
	wg.Wait()

	if len(records) > 0 {
		if len(messages) > 0 {
			//NOTE 发送到索引库存储
			err := ess.kafkaAccess.DoProduce(producer, messages)
			if err != nil {
				return err
			} else {
				logger.Infof(" 基于%d 条原始记录，生成事件, 把事件发送到 索引库 成功.发送了%d条数据 ", len(records), len(messages))
			}

			//NOTE 如果是原子事件数据，则发送至原子事件的专属topic，去触发原子事件模型的依赖任务执行。
			cnt, err := ess.FlushAtomicEvent(messages)
			if err != nil {
				logger.Errorf("把原子事件数据发送到topic:%s 失败[%s]. ", fmt.Sprintf(interfaces.ATOMIC_EVENT_DATA_TOPIC, ess.appSetting.MQSetting.Tenant), err.Error())
				return err
			} else {
				logger.Infof(" 基于%d条原始记录生成%d条原子事件,把事件发送到topic:%s成功 ", len(records), cnt, fmt.Sprintf(interfaces.ATOMIC_EVENT_DATA_TOPIC, ess.appSetting.MQSetting.Tenant))
			}
		} else {
			logger.Infof("消费%d条原始记录,未产生事件!", len(records))
		}

		err = ess.kafkaAccess.CommitOffset(consumer)
		if err != nil {
			logger.Errorf("提交失败，此次处理的数据记录条数为: %d", len(records))
		}
		runtime := time.Since(start_time)
		logger.Debugf("process messages success, count: %d ,time cost,%v", len(records), runtime)
	}
	return nil
}

// 从kafka消费到的数据的Headers里读取数据源类型和ID，并按照数据源类型+ID进行分组处理
func (ess *eventSubService) divideRecordBySourceId(records []*kafka.Message) map[interfaces.DataSource][]*kafka.Message {
	// var dataSource string
	var dataSourceType string
	var dataSourceId string
	var MapRecords = make(map[interfaces.DataSource][]*kafka.Message)
	for _, record := range records {
		if len(record.Headers) > 0 {
			dataSourceId = string(record.Headers[0].Value)
			dataSourceType = string(record.Headers[1].Value)
			//NOTE 暂时默认为data_view
			// dataSourceType = "data_view"
		}
		if dataSourceId == "" {
			source_sr := make(map[string]any)
			_ = json.Unmarshal(record.Value, &source_sr)
			logger.Errorf("Found one record without data source id,the content of record is %v", source_sr)
			continue
		}
		dataSource := interfaces.DataSource{
			DataSourceId:   dataSourceId,
			DataSourceType: dataSourceType,
		}
		MapRecords[dataSource] = append(MapRecords[dataSource], record)
	}
	return MapRecords
}

func (ess *eventSubService) Invoke(engine interfaces.EventEngine, dataSource string, dataSourceType string, records []*kafka.Message, _cache *cache.Cache) ([]*kafka.Message, []*kafka.Message, error) {

	var events interfaces.IEvents
	var err error
	//NOTE flattern sr to top-level
	source_sr := make(map[string]any)
	ParsedRecords := interfaces.Records{}
	ctx := context.Background()
	for _, record := range records {
		sr := make(map[string]any)
		err := json.Unmarshal(record.Value, &source_sr)
		if err != nil {
			return []*kafka.Message{}, []*kafka.Message{}, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_InternalError)
		}
		err = flatten(true, "", source_sr, sr)
		if err != nil {
			return []*kafka.Message{}, []*kafka.Message{}, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_InternalError)
		}
		ParsedRecords = append(ParsedRecords, sr)
	}
	var ems []interfaces.EventModel
	//NOTE 如果是指标数据或日志数据
	if dataSourceType == "data_view" || dataSourceType == "metric_data" {
		ems, _ = ess.QueryEventModelBySourceId(dataSource, dataSourceType, _cache)
	} else if dataSourceType == "event_model" { //NOTE 如果是原子事件数据，则寻找此事件模型属性中的依赖任务进行触发。
		var dependentModelType = dataSourceType
		ems, _ = ess.QueryEventModelByDownstreamDependent(dataSource, dependentModelType, _cache)
	} else {
		ems = []interfaces.EventModel{}
	}
	collectMessages := make([]*kafka.Message, 0)
	collectEntities := []*kafka.Message{}
	if len(ems) == 0 {
		return collectMessages, collectEntities, nil
	}

	sourceRecords := interfaces.SourceRecords{
		Records: ParsedRecords,
	}

	// logger.Debugf(" dataSource %s, data source type %s , prepare %d event_model to judge", dataSource, dataSourceType, len(ems))
	for _, em := range ems {
		logger.Debugf(" dataSource %s, event_model %s :prepare to judge", dataSource, em.EventModelName)
		if em.DetectRule.Type == "agi_detect" || em.AggregateRule.Type == "agi_aggregation" {
			// do nothing
		} else {
			events, _, err = engine.Judge(ctx, sourceRecords, em)
			if err != nil {
				logger.Errorf("judge source record failed:%v", err)
				continue
			}
			if len(events) == 0 {
				continue
			}
		}

		messages := make([]*kafka.Message, 0)
		kaEntities := make([]*kafka.Message, 0)
		topic := fmt.Sprintf(interfaces.MODEL_PERSIST_INPUT, ess.appSetting.MQSetting.Tenant)

		//NOTE 将事件数据转化为kafka  message
		err = eventDataTransferToMessage(events, em.Task, &messages, topic)
		if err != nil {
			logger.Errorf("事件模型[%s]把数据转换为 kafka massage 失败[%s]. ", em.EventModelName, err.Error())
			continue
		}
		// logger.Infof("事件模型[%s], 把事件发送到 告警 成功. ", em.EventModelName)
		collectMessages = append(collectMessages, messages...)
		collectEntities = append(collectMessages, kaEntities...)

		logger.Debugf(" source_data_id %s ,event_model %s :generate %d event by judge with %d source records %d", dataSource, em.EventModelName, len(sourceRecords.Records))
	}
	logger.Debugf(" source_data_id %s, total %d event_model generate %d event by judge with  source records %d", dataSource, len(ems), len(events), len(sourceRecords.Records))
	return collectMessages, collectEntities, nil
}

// 拼接新的字段名: a.b.c
func joinKey(top bool, prefix, subkey string) string {
	if subkey == "" {
		return prefix
	}

	key := prefix

	if top {
		key += subkey
	} else {
		key += "." + subkey
	}

	return key
}

func flatten(top bool, prefix string, src any, dest map[string]any) error {
	assign := func(newKey string, val any) error {
		switch value := val.(type) {
		case map[string]any:
			// 保留值为{}的字段
			if len(value) == 0 {
				dest[newKey] = value
				break
			}

			if err := flatten(false, newKey, value, dest); err != nil {
				return err
			}

		case []any:
			// 保留值为[]的字段
			if len(value) == 0 {
				dest[newKey] = value
				break
			}

			switch value[0].(type) {
			// 如果数组的元素是map或数组，继续向下展开
			case map[string]any, []any:
				if err := flatten(false, newKey, value, dest); err != nil {
					return err
				}
			default:

				dest[newKey] = value

			}

		default:
			if existVal, ok := dest[newKey]; !ok {
				dest[newKey] = value

			} else {
				// 如果展开后的字段已存在，则将值存成数组
				vals := make([]any, 0)

				switch existedValue := existVal.(type) {
				case []any:
					existedValue = append(existedValue, value)

					dest[newKey] = existedValue

				default:
					vals = append(vals, existedValue, value)

					dest[newKey] = vals

				}

			}
		}

		return nil

	}

	switch nested := src.(type) {
	case map[string]any:
		for key, val := range nested {
			newKey := joinKey(top, prefix, key)
			err := assign(newKey, val)
			if err != nil {
				return err
			}
		}
	case []any:
		for _, val := range nested {
			newKey := joinKey(top, prefix, "")
			err := assign(newKey, val)
			if err != nil {
				return err
			}

		}
	default:
		return errors.New("not a valid input: map or slice")
	}

	return nil
}

func (ess *eventSubService) FlushAtomicEvent(messages []*kafka.Message) (int, error) {
	topic := fmt.Sprintf(interfaces.ATOMIC_EVENT_DATA_TOPIC, ess.appSetting.MQSetting.Tenant)

	producer, err := ess.kafkaAccess.NewTrxProducer(topic)
	if err != nil {
		return 0, err
	}
	defer producer.Close()
	//修改topic
	FixedMessages := []*kafka.Message{}
	for _, message := range messages {

		dataSourceType := message.Headers[1].Value
		//NOTE 只发送原子事件数据去触发相对应的依赖任务运行
		if string(dataSourceType) != "event_model" {
			continue
		}
		message.TopicPartition.Topic = &topic
		FixedMessages = append(FixedMessages, message)
	}
	if len(FixedMessages) == 0 {
		return 0, nil
	}
	// 4.2 消费数据
	err = ess.kafkaAccess.DoProduce(producer, FixedMessages)
	if err != nil {
		return 0, err
	}
	return len(FixedMessages), nil
}

func (ess *eventSubService) QueryEventModelBySourceId(dataSource string, dataSourceType string, _cache *cache.Cache) ([]interfaces.EventModel, error) {
	// em := interfaces.EventModel{}
	var ems []interfaces.EventModel
	if cacheEventModels, ok := _cache.Get(dataSource); !ok {
		// sourceId, _ := strconv.ParseUint(dataSource, 10, 64)
		if dataSourceType == "data_view" {
			// ems = ess.GetEventModelBySourceId(dataSource)
			ems, _ = ess.emAccess.GetEventModelBySourceId(context.Background(), dataSource)
		}
		_cache.Set(dataSource, ems, interfaces.EXPIRATION_TIME)
	} else {
		ems = cacheEventModels.([]interfaces.EventModel)
	}

	if len(ems) == 0 {
		return []interfaces.EventModel{}, nil
	}
	return ems, nil

}

func (ess *eventSubService) QueryEventModelByDownstreamDependent(dependentModelId string, dependentModelType string, _cache *cache.Cache) ([]interfaces.EventModel, error) {
	var ems []interfaces.EventModel
	u := dependentModelId + "_depend" //NOTE 为了区分数据源引用或依赖引用，此处用_depend来标记依赖引用。
	if cacheEventModels, ok := _cache.Get(u); !ok {
		// sourceId, _ := strconv.ParseUint(dataSource, 10, 64)
		ems = ess.GetEventModelByDownstreamDependent(dependentModelId, dependentModelType)
		if len(ems) == 0 {
			return ems, nil
		}
		_cache.Set(u, ems, interfaces.EXPIRATION_TIME)
	} else {
		ems = cacheEventModels.([]interfaces.EventModel)
	}

	if len(ems) == 0 {
		return []interfaces.EventModel{}, nil
	}
	return ems, nil
}

func (ess *eventSubService) GetEventModelByDownstreamDependent(dependentModelId string, dependentModelType string) []interfaces.EventModel {
	ems, _ := ess.emAccess.GetEventModelById(context.Background(), dependentModelId)
	var dependentModels = []interfaces.EventModel{}
	if len(ems) > 0 {
		DownstreamDependentModelIds := ems[0].DownstreamDependentModel
		for _, id := range DownstreamDependentModelIds {
			ems, _ := ess.emAccess.GetEventModelById(context.Background(), id)
			if len(ems) > 0 {
				dependentModels = append(dependentModels, ems[0])
			}
		}
	}
	return dependentModels
}

// 数据转换为 event 数据格式. category, 任务名称，时间窗口
func eventDataTransferToMessage(events interfaces.IEvents, task interfaces.EventTask, messages *[]*kafka.Message, topic string) error {
	for _, event := range events {
		message := make(map[string]interface{})
		// 补齐元字段： type: indexbase; __index_base: indexbase; __data_type: 索引库的data_type
		message["category"] = "event"
		message["__data_type"] = task.StorageConfig.IndexBase
		message["__index_base"] = task.StorageConfig.IndexBase
		message["type"] = task.StorageConfig.IndexBase

		//事件信息
		message["event_message"] = event.GenerateMessage()
		bytes, err := json.Marshal(event.GetContext())
		if err != nil {
			return err
		}
		message["context_str"] = string(bytes)
		message["context"] = event.GetContext()
		//NOTE 将id 用string 发送给kafka,不会丢失精度
		message["id"] = event.GetBaseEvent().Id
		message["title"] = event.GetBaseEvent().Title
		message["event_model_id"] = event.GetBaseEvent().EventModelId
		message["event_type"] = event.GetBaseEvent().EventType
		if event.GetBaseEvent().EventType == "aggregate" {
			message["aggregate_algo"] = event.(interfaces.AggregateEvent).AggregateAlgo
			message["aggregate_type"] = event.(interfaces.AggregateEvent).AggregateType
		} else {
			message["detect_algo"] = event.(interfaces.AtomicEvent).DetectAlgo
			message["detect_type"] = event.(interfaces.AtomicEvent).DetectType
		}
		message["level"] = event.GetLevel()
		message["level_name"] = event.GetBaseEvent().LevelName
		message["generate_type"] = event.GetBaseEvent().GenerateType
		message["@timestamp"] = event.GetBaseEvent().CreateTime
		message["trigger_time"] = event.GetTriggerTime()
		bytes, err = json.Marshal(event.GetTriggerData())
		if err != nil {
			return err
		}
		message["trigger_data_str"] = string(bytes)

		message["tags"] = event.GetTag()
		message["event_model_name"] = event.GetBaseEvent().EventModelName
		message["data_source"] = event.GetBaseEvent().DataSource
		message["data_source_name"] = event.GetBaseEvent().DataSourceName
		message["data_source_type"] = event.GetBaseEvent().DataSourceType
		bytes, err = json.Marshal(event.GetBaseEvent().Labels)
		if err != nil {
			return err
		}
		message["labels_str"] = string(bytes)

		bytes, err = json.Marshal(event.GetBaseEvent().Relations)
		if err != nil {
			return err
		}
		message["relations_str"] = string(bytes)
		message["relations"] = event.GetBaseEvent().Relations

		bytes, err = json.Marshal(map[string]any{
			"interval": event.GetBaseEvent().DefaultTimeWindow.Interval,
			"unit":     event.GetBaseEvent().DefaultTimeWindow.Unit,
		})
		if err != nil {
			return err
		}
		message["default_time_window_str"] = string(bytes)

		bytes, err = json.Marshal(event.GetBaseEvent().Schedule)
		if err != nil {
			return err
		}
		message["schedule_str"] = string(bytes)

		bytes, err = json.Marshal(message)
		if err != nil {
			return err
		}

		headers := []kafka.Header{
			{Key: "__id", Value: []byte(event.GetBaseEvent().EventModelId)},
			{Key: "__type", Value: []byte("event_model")},
		}
		kafkaMessage := &kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: kafka.PartitionAny,
			},
			Value:   bytes,
			Headers: headers,
		}
		*messages = append(*messages, kafkaMessage)

	}
	return nil
}
