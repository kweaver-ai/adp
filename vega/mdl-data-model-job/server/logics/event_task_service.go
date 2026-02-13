// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package logics

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"

	"data-model-job/common"
	"data-model-job/interfaces"
)

var (
	etsOnce sync.Once
	ets     interfaces.EventTaskService
)

type eventTaskService struct {
	appSetting *common.AppSetting
	emAccess   interfaces.EventModelAccess
	uAccess    interfaces.UniqueryAccess
	kAccess    interfaces.KafkaAccess
	iBAccess   interfaces.IndexBaseAccess
}

func NewEventTaskService(appSetting *common.AppSetting) interfaces.EventTaskService {
	etsOnce.Do(func() {
		ets = &eventTaskService{
			appSetting: appSetting,
			emAccess:   EMAccess,
			uAccess:    UAccess,
			kAccess:    KAccess,
			iBAccess:   IBAccess,
		}

	})
	return ets
}

func (etService *eventTaskService) EventTaskExecutor(ctx context.Context, task interfaces.EventTask) string {
	// accountInfo 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, task.Creator)

	//NOTE: 根据modelId获取task信息
	eventModel, err := etService.emAccess.GetEventModel(ctx, task.ModelID)
	if err != nil {
		logger.Errorf("EventTaskExecutor failed, GetEventModel failed:%v!", err)
		// 更新任务执行状态为失败
		updateErr := etService.emAccess.UpdateEventTaskAttributesById(ctx,
			interfaces.EventTask{
				TaskID:           task.TaskID,
				TaskStatus:       interfaces.SCHEDULE_RUNNING_STATUS_FAILED,
				ErrorDetails:     err.Error(),
				StatusUpdateTime: time.Now().UnixMilli(),
			})

		if updateErr != nil {
			logger.Errorf("EventTaskExecutor failed, UpdateEventTaskAttributesById failed:%v!", err)
			return fmt.Sprintf("%s:%s", err.Error(), updateErr.Error())
		}
		return err.Error()
	}
	if !(task.ExecuteParameter == nil) {
		eventModel.Task.ExecuteParameter = task.ExecuteParameter
	}
	msg, err := etService.executTask(ctx, task)
	if err != nil {
		logger.Errorf("EventTaskExecutor failed, 事件模型[%s] executTask failed:%v!", eventModel.EventModelName, err)
		// 更新任务执行状态为失败
		updateErr := etService.emAccess.UpdateEventTaskAttributesById(ctx,
			interfaces.EventTask{
				TaskID:           task.TaskID,
				TaskStatus:       interfaces.SCHEDULE_RUNNING_STATUS_FAILED,
				ErrorDetails:     err.Error(),
				StatusUpdateTime: time.Now().UnixMilli(),
			})
		if updateErr != nil {
			logger.Errorf("EventTaskExecutor failed, 事件模型[%s] UpdateEventTaskAttributesById failed:%v!", eventModel.EventModelName, err)
			return fmt.Sprintf("%s:%s", err.Error(), updateErr.Error())
		}
		return err.Error()
	} else {
		logger.Debugf("事件模型[%s] executTask success!", eventModel.EventModelName)
		return msg
	}
}

func (etService *eventTaskService) executTask(ctx context.Context, task interfaces.EventTask) (string, error) {
	// 根据 task 中配置的信息，查询数据，查询结果写到 kafka
	// 请求 index mgnt 获取索引库的data type
	indexBases, err := etService.iBAccess.GetIndexBasesByTypes(ctx, []string{task.StorageConfig.IndexBase})
	if err != nil {
		logger.Errorf("ExecutTask failed, GetIndexBasesByTypes failed:%v!", err)
		return "", err
	}
	if len(indexBases) != 1 {
		logger.Errorf("ExecutTask failed, len(indexBases) is not equal to 1!")
		// 索引库数量不等1，return
		return "", fmt.Errorf("索引库类型[%s]对应的索引库数量不等于1,为[%d]", task.StorageConfig.IndexBase, len(indexBases))
	}
	eventData, msg, err := etService.GetCompleteEventData(ctx, task)
	if err != nil {
		logger.Errorf("ExecutTask failed, GetCompleteEventData failed:%v!", err)
		return msg, err
	}
	// 查询得到数据才发送
	if eventData.TotalCount > 0 {
		logger.Debugf("当前任务获取事件个数为%d", len(eventData.Entries))
		messages := make([]*kafka.Message, 0)

		// 3.2.数据转换为 event 数据格式. category, 任务名称，时间窗口
		err := etService.eventDataTransferToMessage(eventData.Entries, task, indexBases[0], &messages)
		if err != nil {
			logger.Errorf("ExecutTask failed, eventDataTransferToMessage failed:%v!", err)
			msg += fmt.Sprintf("任务[%s],把数据转换为 kafka massage 失败[%s]. ", task.TaskID, err.Error())
			return msg, err
		}

		// 4 把查询结果写kafka
		err = etService.sendKafka(task, messages)
		if err != nil {
			logger.Errorf("ExecutTask failed, sendKafka failed:%v!", err)
			msg += fmt.Sprintf("任务[%s], 把数据发送到kafka失败[%s]. ", task.TaskID, err.Error())
			return msg, err
		} else {
			msg += fmt.Sprintf("任务[%s], 发送[%d]条数据到kafka", task.TaskID, len(messages))
		}

	} else {
		logger.Debugf("ExecutTask, 获取到的数据为空!")
		msg += fmt.Sprintf("任务[%s], 获取到的数据为空. ", task.TaskID)
	}
	//更新执行状态为成功
	err = etService.emAccess.UpdateEventTaskAttributesById(ctx,
		interfaces.EventTask{
			TaskID:           task.TaskID,
			TaskStatus:       interfaces.SCHEDULE_RUNNING_STATUS_SUCCESS,
			ErrorDetails:     "",
			StatusUpdateTime: time.Now().UnixMilli(),
		})
	if err != nil {
		logger.Errorf("ExecutTask failed, UpdateEventTaskAttributesById failed:%v!", err)
		return "", fmt.Errorf("任务[%s], 更新任务执行状态失败[%s]. ", task.TaskID, err.Error())
	}
	logger.Debugf("service: 事件模型任务[%v]执行完成. %s", task.TaskID, msg)
	return msg, nil
}

// 发送当前窗口的数据到kafka
func (etService *eventTaskService) sendKafka(task interfaces.EventTask,
	messages []*kafka.Message) (err error) {

	// 4 把查询结果写kafka topic 需要注意
	// 4.1 创建生产者,传一个uniqueId
	topic := fmt.Sprintf(interfaces.MODEL_PERSIST_INPUT, etService.appSetting.MQSetting.Tenant)

	uniqueId := fmt.Sprintf("%s-%s", task.TaskID, topic)
	producer, err := etService.kAccess.NewTrxProducer(uniqueId)
	if err != nil {
		return err
	}
	defer producer.Close()
	// 4.2 消费数据
	err = etService.kAccess.DoProduceMsgToKafka(producer, messages)
	if err != nil {
		return err
	}
	return nil
}

// 数据转换为 event 数据格式. category, 任务名称，时间窗口
func (etService *eventTaskService) eventDataTransferToMessage(events []interfaces.EventModelData, task interfaces.EventTask,
	indexbase interfaces.IndexBase, messages *[]*kafka.Message) error {

	topic := fmt.Sprintf(interfaces.MODEL_PERSIST_INPUT, etService.appSetting.MQSetting.Tenant)

	for _, event := range events {
		if event.GenerateType == "" {
			event.GenerateType = interfaces.GENERATE_TYPE_FOR_BATCH
		}
		message := make(map[string]interface{})
		// 补齐元字段： type: indexbase; __index_base: indexbase; __data_type: 索引库的data_type
		message["category"] = "event"
		message["__data_type"] = indexbase.DataType
		message["__index_base"] = indexbase.BaseType
		message["type"] = indexbase.BaseType

		//事件信息
		message["event_message"] = event.Message
		bytes, err := json.Marshal(event.Context)
		if err != nil {
			return err
		}
		message["context_str"] = string(bytes)
		message["context"] = event.Context
		//NOTE 将id 用string 发送给kafka,不会丢失精度
		message["id"] = event.Id
		message["title"] = event.Title
		message["event_model_id"] = event.EventModelId
		message["event_type"] = event.EventType
		if event.EventType == "aggregate" {
			message["aggregate_algo"] = event.AggregateAlgo
			message["aggregate_type"] = event.AggregateType
		} else {
			message["detect_algo"] = event.DetectAlgo
			message["detect_type"] = event.DetectType
		}
		message["level"] = event.Level
		message["generate_type"] = event.GenerateType
		message["level_name"] = event.LevelName
		message["@timestamp"] = event.CreateTime
		message["trigger_time"] = event.TriggerTime
		bytes, err = json.Marshal(event.TriggerData)
		if err != nil {
			return err
		}
		message["trigger_data_str"] = string(bytes)
		message["tags"] = event.Tags
		message["event_model_name"] = event.EventModelName
		message["data_source"] = event.DataSource
		message["data_source_name"] = event.DataSourceName
		message["data_source_type"] = event.DataSourceType
		bytes, err = json.Marshal(event.Labels)
		if err != nil {
			return err
		}
		message["labels_str"] = string(bytes)

		bytes, err = json.Marshal(event.Relations)
		if err != nil {
			return err
		}
		message["relations_str"] = string(bytes)
		message["relations"] = event.Relations
		// message["llm_app"] = event.LLMApp
		// message["ekn"] = event.Ekn

		bytes, err = json.Marshal(map[string]any{
			"interval": event.DefaultTimeWindow.Interval,
			"unit":     event.DefaultTimeWindow.Unit,
		})
		if err != nil {
			return err
		}
		message["default_time_window_str"] = string(bytes)

		bytes, err = json.Marshal(event.Schedule)
		if err != nil {
			return err
		}
		message["schedule_str"] = string(bytes)

		bytes, err = json.Marshal(message)
		if err != nil {
			return err
		}
		kafkaMessage := &kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: kafka.PartitionAny,
			},
			Value: bytes,
		}
		*messages = append(*messages, kafkaMessage)

	}
	return nil
}

// NOTE获取补全事件数据（+2min）
func (etService *eventTaskService) GetCompleteEventData(ctx context.Context, task interfaces.EventTask) (interfaces.EventModelResponse, string, error) {
	msg := ""
	eventModel, err := etService.emAccess.GetEventModel(ctx, task.ModelID)
	if err != nil {
		logger.Errorf("GetCompleteEventData failed,GetEventModel failed:%v", err)
		return interfaces.EventModelResponse{}, "", err
	}
	timeWindowStr := strconv.Itoa(eventModel.DefaultTimeWindow.Interval) + eventModel.DefaultTimeWindow.Unit
	timeWindowT, _ := common.ParseDuration(timeWindowStr, common.DurationDayHourMinuteRE, true)
	timeWindow := common.DurationMilliseconds(timeWindowT)
	step := common.DurationMilliseconds(time.Minute * time.Duration(2))

	curTime := common.GetCurrentTimestamp()
	start := curTime - 2*timeWindow
	end := curTime + step - timeWindow

	logger.Debugf("事件模型[%s]: 调度时间LogDateTime为%d", eventModel.EventModelName, curTime)
	logger.Debugf("事件模型[%s]: 持久化请求时间start:%d,end:%d", eventModel.EventModelName, start, end)
	// 3.1. 请求uniquery，获取最新的一条数据记录。为了获取最近数据的时间
	event, err := etService.uAccess.GetEventModelData(ctx, interfaces.EventModelQueryRequest{
		Limit:  -1,
		Offset: 0,
		Querys: []interfaces.EventQuery{
			{
				//end - 2*timewindow
				Start: start,
				//end+ 2min - timewindow
				End:       end, //当前时间戳(ms)
				QueryType: "range_query",
				Id:        task.ModelID,
				Limit:     1,
				SortKey:   "trigger_time",
				Direction: "desc",
			},
		},
		SortKey:   "trigger_time",
		Direction: "desc",
	})
	if err != nil {
		msg += fmt.Sprintf("任务[%s], 事件模型[%s]: 获取事件数据失败[%s] ", task.TaskID, eventModel.EventModelName, err.Error())
		logger.Errorf("事件模型[%s]: GetEventModelData failed:[%v]", eventModel.EventModelName, err)
		return interfaces.EventModelResponse{}, msg, err
	}

	end = curTime / 1000 * 1000
	start = end - timeWindow - step
	if event.TotalCount > 0 {
		//NOTE： 取触发时间而不是事件检测时间
		for _, event := range event.Entries {
			logger.Debugf("最近一次事件的trrigerTime:%v", event.TriggerTime)
			if event.TriggerTime >= start {
				//NOTE: 去除重复的事件
				start = event.TriggerTime + 1
			}
		}
	}

	//NOTE 时间戳带参数执行
	if endTimeStamp, ok := task.ExecuteParameter["end"]; ok {
		logger.Debugf("%+v", task.ExecuteParameter["end"])
		logger.Debugf("%T", task.ExecuteParameter["end"])
		if _, ok = endTimeStamp.(float64); ok {
			end = int64(endTimeStamp.(float64))
			if startTimeStamp, ok := task.ExecuteParameter["start"]; ok {
				if _, ok = startTimeStamp.(float64); ok {
					start = int64(startTimeStamp.(float64))
				} else {
					logger.Debugf("事件模型[%s]: 执行参数中的start:%v,转换为int64失败", eventModel.EventModelName, startTimeStamp)
					msg += fmt.Sprintf("事件模型[%s]: 执行参数中的start:%v,转换为int64失败", eventModel.EventModelName, startTimeStamp)
					return interfaces.EventModelResponse{}, msg, fmt.Errorf("事件模型[%s]: 执行参数中的start:%v,转换为int64失败", eventModel.EventModelName, startTimeStamp)
				}

			} else {
				start = end - timeWindow
			}
		} else {
			logger.Debugf("事件模型[%s]: 执行参数中的end:%v,转换为int64失败", eventModel.EventModelName, endTimeStamp)
			msg += fmt.Sprintf("事件模型[%s]: 执行参数中的end:%v,转换为int64失败", eventModel.EventModelName, endTimeStamp)
			return interfaces.EventModelResponse{}, msg, fmt.Errorf("事件模型[%s]: 执行参数中的end:%v,转换为int64失败", eventModel.EventModelName, endTimeStamp)
		}
	}
	//NOTE 提取自定义参数,用于数据源过滤,暂时只支持多条与
	var extraction []interfaces.Filter
	if source_extract, ok := task.ExecuteParameter["extraction"]; ok {
		if v, ok := source_extract.(map[string]any); ok {
			for _k, _v := range v {
				_filter := interfaces.Filter{Name: _k, Operation: "=", Value: _v.(string)}
				extraction = append(extraction, _filter)
			}
		}
	}
	logger.Debugf("事件模型[%s]: 任务执行参数：%v", eventModel.EventModelName, extraction)
	logger.Debugf("事件模型[%s]: 即时查询请求时间start:%d,end:%d", eventModel.EventModelName, start, end)
	eventData, err := etService.uAccess.GetEventModelData(ctx, interfaces.EventModelQueryRequest{
		Limit:  -1,
		Offset: 0,
		Querys: []interfaces.EventQuery{
			{
				Start:      start,
				End:        end, //调度时间(ms)
				QueryType:  "instant_query",
				Id:         task.ModelID,
				Limit:      -1,
				Extraction: extraction,
			},
		},
	})
	if err != nil {
		msg += fmt.Sprintf("任务[%s], 事件模型[%s]: 获取事件数据失败[%s] ", task.TaskID, eventModel.EventModelName, err.Error())
		logger.Errorf("任务[%s],事件模型[%s]: GetEventModelData failed:[%v]", task.TaskID, eventModel.EventModelName, err)
		return interfaces.EventModelResponse{}, msg, err
	}
	// if eventModel.AggregateRule.Type == "agi_aggregation" || eventModel.DetectRule.Type == "agi_detect" {
	// 	msg = fmt.Sprintf("任务[%d],事件模型[%s]: 智能算法应用. ", task.TaskID, eventModel.EventModelName)
	// 	return interfaces.EventModelResponse{}, msg, nil
	// }
	return eventData, msg, nil
}
