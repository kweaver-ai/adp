// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package logics

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"

	"data-model-job/common"
	"data-model-job/interfaces"
)

var (
	mtsOnce sync.Once
	mts     interfaces.MetricTaskService
)

type metricTaskService struct {
	appSetting *common.AppSetting
	mmAccess   interfaces.MetricModelAccess
	uAccess    interfaces.UniqueryAccess
	kAccess    interfaces.KafkaAccess
	iBAccess   interfaces.IndexBaseAccess
}

func NewMetricTaskService(appSetting *common.AppSetting) interfaces.MetricTaskService {
	mtsOnce.Do(func() {
		// 初始化一个调度器，全局一个
		mts = &metricTaskService{
			appSetting: appSetting,
			mmAccess:   MMAccess,
			uAccess:    UAccess,
			kAccess:    KAccess,
			iBAccess:   IBAccess,
		}

	})
	return mts
}

func (mtService *metricTaskService) MetricTaskExecutor(ctx context.Context, metricTask interfaces.MetricTask) string {
	// accountInfo 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, metricTask.Creator)

	// 按 module 来调用不同的函数
	switch metricTask.ModuleType {
	case interfaces.MODULE_TYPE_METRIC_MODEL:
		msg, err := mtService.executTask(ctx, metricTask)
		if err != nil {
			logger.Errorf(err.Error())

			// 更新任务的状态为失败
			updateErr := mtService.mmAccess.UpdateTaskAttributesById(ctx, metricTask.TaskID,
				interfaces.MetricTask{ExeccuteStatus: interfaces.SCHEDULE_SYNC_STATUS_FAILED})

			if updateErr != nil {
				return fmt.Sprintf("%s:%s", err.Error(), updateErr.Error())
			}

			return err.Error()
		} else {
			return msg
		}
	case interfaces.MODULE_TYPE_OBJECTIVE_MODEL:
		msg, err := mtService.executObjectiveTask(ctx, metricTask)
		if err != nil {
			logger.Errorf(err.Error())

			// 更新任务的状态为失败
			updateErr := mtService.mmAccess.UpdateTaskAttributesById(ctx, metricTask.TaskID,
				interfaces.MetricTask{ExeccuteStatus: interfaces.SCHEDULE_SYNC_STATUS_FAILED})

			if updateErr != nil {
				return fmt.Sprintf("%s:%s", err.Error(), updateErr.Error())
			}

			return err.Error()
		} else {
			return msg
		}
	default:
		return fmt.Sprintf("Unsurpport module type: %s", metricTask.ModuleType)
	}
}

func (mtService *metricTaskService) executTask(ctx context.Context, task interfaces.MetricTask) (string, error) {
	// 2. 请求 index mgnt 获取索引库的data type
	indexBases, err := mtService.iBAccess.GetIndexBasesByTypes(ctx, []string{task.IndexBase})
	if err != nil {
		return "", err
	}
	if len(indexBases) != 1 {
		// 索引库数量不等1，return
		return "", fmt.Errorf("指标模型任务的索引库类型[%s]对应的索引库数量不等于1,为[%d]", task.IndexBase, len(indexBases))
	}

	// 3. 根据 taskid 获取当前计划执行时间 api调用
	// todo：plan_time 只维护一个，任务的数据都写成功了才算成功，一起写，不分部分写，以一个任务为单位，都成功了才更新plan_time,
	// 在data-model初始化plan_time时，需要修改
	taskPlanTime, err := mtService.mmAccess.GetTaskPlanTimeById(ctx, task.TaskID)
	if err != nil {
		return "", err
	}
	if taskPlanTime == 0 {
		return "", fmt.Errorf("指标模型任务的计划时间【%v】不合法, 期望为计划时间长度为1", taskPlanTime)
	}

	steps := make([]int64, 0)
	for _, stepStr := range task.Steps {
		// 解析步长为具体毫秒
		step, err := common.ParseDuration(stepStr, common.DurationDayHourMinuteRE, true)
		if err != nil {
			logger.Errorf("Failed to parse schedule duration, err: %s", err.Error())
			return "", fmt.Errorf("failed to parse schedule duration, err: %s", err.Error())
		}
		steps = append(steps, int64(step/(time.Millisecond/time.Nanosecond)))
	}

	ts := time.Now().UnixNano() / int64(time.Millisecond/time.Nanosecond)
	// 所以对plan_time按任务的steps的5m修正，对now按步长修正
	fixedPlanTime := int64(math.Floor(float64(taskPlanTime)/float64(interfaces.MIN_STEP))) * interfaces.MIN_STEP
	timeZone := os.Getenv("TZ")
	if timeZone == "" {
		timeZone = interfaces.DEFAULT_TIME_ZONE
	}
	location, err := time.LoadLocation(timeZone)
	if err != nil {
		// 记录异常日志
		logger.Errorf("LoadLocation error: %s", err.Error())
		return "", fmt.Errorf("LoadLocation error: %s", err.Error())
	}
	_, offset := time.Now().In(location).Zone()

	msgTotal := 0
	var planTime int64
	for planTime = fixedPlanTime; planTime <= ts; {
		messages := make([]*kafka.Message, 0)
		for i, stepStr := range task.Steps {
			// 判断当前步长点是否能被step整除。 用plan_time偏移到utc时间的除。
			p := planTime + int64(offset*1000)
			if p%steps[i] != 0 {
				// 跳过，遍历下一个step
				logger.Debugf("跳过：指标模型任务[%s]当前计算时间点plantime:%d. 当前时间now:%d, 执行步长为%s", task.TaskID, planTime, ts, stepStr)
				continue
			}

			// 确定 look_back_delta。因为在管理端，语言是 promql 时，校验了时间窗口应为空
			lookBackDeltas := make([]string, 0)
			if len(task.TimeWindows) == 0 {
				// promql 时，用任务的 step 作为即时查询的 look_back_delta
				lookBackDeltas = append(lookBackDeltas, stepStr)
			} else {
				// dsl 时，用时间窗口 timw_window 作为即时查询的 look_back_delta
				lookBackDeltas = append(lookBackDeltas, task.TimeWindows...)
			}

			// 遍历 look_back_delta，查询数据并组装kafka message
			for _, lookBackDelta := range lookBackDeltas {
				// 请求uniquery，即时查询，

				metricData, err := mtService.uAccess.GetMetricModelData(ctx, task.ModelID, interfaces.MetricModelQuery{
					IsInstantQuery: true,
					Time:           planTime,
					LookBackDelta:  lookBackDelta,
				})
				logger.Debugf("指标模型任务[%s], 当前step【%s】查询的步长点为【%d】", task.TaskID, stepStr, planTime)
				if err != nil {
					return "", fmt.Errorf("指标模型任务[%s], 查询参数: time=%d,look_back_delta=%s, 获取指标数据失败[%s]. ", task.TaskID, planTime, lookBackDelta, err.Error())
				}

				//  组装kafka messages
				if len(metricData.Datas) > 0 {
					// 3.2.数据转换为 metric 数据格式. category, 任务名称，时间窗口
					err = mtService.metricDataTransferToMessage(planTime-steps[i], metricData, lookBackDelta, task, stepStr, indexBases[0], &messages)
					if err != nil {
						return "", fmt.Errorf("指标模型任务[%s], 时间窗口[%s]把数据转换为 kafka massage 失败[%s]. ", task.TaskID, lookBackDelta, err.Error())
					}
				}
			}
		}
		planTime += interfaces.MIN_STEP

		// 每次个时间点上提交一次kafka，不等批次，一个步长点提交一次数据，并更新planTime
		if len(messages) > 0 {
			err = mtService.flushToKafka(task, messages)
			if err != nil {
				return "", fmt.Errorf("指标模型任务[%s], 把数据发送到 kafka 失败[%s]. ", task.TaskID, err.Error())
			}

			// 更新数据库中任务的计划执行时间plan_time 按5m修正后按任务的steps的最大公约数递增，任务的状态为任务执行中执行成功
			taskPlanTime = planTime
			err = mtService.mmAccess.UpdateTaskAttributesById(ctx, task.TaskID,
				interfaces.MetricTask{PlanTime: taskPlanTime, ExeccuteStatus: interfaces.SCHEDULE_SYNC_STATUS_SUCCESS})
			if err != nil {
				return "", fmt.Errorf("指标模型任务[%s], 更新任务计划时间失败[%s]. ", task.TaskID, err.Error())
			}

			msgTotal += len(messages)
			logger.Debugf("service: 指标模型任务[%s]执行完成. 共发送[%d]条数据到kafka。", task.TaskID, len(messages))
		}
	}

	logger.Debugf("service:指标模型任务 [%s]执行完成. 共发送[%d]条数据到kafka。", task.TaskID, msgTotal)
	return fmt.Sprintf("service: 指标模型任务[%s]执行完成. 共发送[%d]条数据到kafka。", task.TaskID, msgTotal), nil
}

func (mtService *metricTaskService) executObjectiveTask(ctx context.Context, task interfaces.MetricTask) (string, error) {
	// 2. 请求 index mgnt 获取索引库的data type
	indexBases, err := mtService.iBAccess.GetIndexBasesByTypes(ctx, []string{task.IndexBase})
	if err != nil {
		return "", err
	}
	if len(indexBases) != 1 {
		// 索引库数量不等1，return
		return "", fmt.Errorf("目标模型任务的索引库类型[%s]对应的索引库数量不等于1,为[%d]", task.IndexBase, len(indexBases))
	}

	// 3. 根据 taskid 获取当前计划执行时间 api调用
	// todo：plan_time 只维护一个，任务的数据都写成功了才算成功，一起写，不分部分写，以一个任务为单位，都成功了才更新plan_time,
	// 在data-model初始化plan_time时，需要修改
	taskPlanTime, err := mtService.mmAccess.GetTaskPlanTimeById(ctx, task.TaskID)
	if err != nil {
		return "", err
	}
	if taskPlanTime == 0 {
		return "", fmt.Errorf("目标模型任务[%s]的计划时间【%v】不合法, 期望为计划时间长度为1", task.TaskID, taskPlanTime)
	}

	steps := make([]int64, 0)
	for _, stepStr := range task.Steps {
		// 解析步长为具体毫秒
		step, err := common.ParseDuration(stepStr, common.DurationDayHourMinuteRE, true)
		if err != nil {
			logger.Errorf("Failed to parse schedule duration, err: %s", err.Error())
			return "", fmt.Errorf("failed to parse schedule duration, err: %s", err.Error())
		}
		steps = append(steps, int64(step/(time.Millisecond/time.Nanosecond)))
	}

	ts := time.Now().UnixNano() / int64(time.Millisecond/time.Nanosecond)
	// 所以对plan_time按任务的steps的5m修正，对now按步长修正
	fixedPlanTime := int64(math.Floor(float64(taskPlanTime)/float64(interfaces.MIN_STEP))) * interfaces.MIN_STEP
	timeZone := os.Getenv("TZ")
	if timeZone == "" {
		timeZone = interfaces.DEFAULT_TIME_ZONE
	}
	location, err := time.LoadLocation(timeZone)
	if err != nil {
		// 记录异常日志
		logger.Errorf("LoadLocation error: %s", err.Error())
		return "", fmt.Errorf("LoadLocation error: %s", err.Error())
	}
	_, offset := time.Now().In(location).Zone()

	msgTotal := 0
	var planTime int64
	for planTime = fixedPlanTime; planTime <= ts; {
		messages := make([]*kafka.Message, 0)
		for i, stepStr := range task.Steps {
			// 判断当前步长点是否能被step整除。 用plan_time偏移到utc时间的除。
			p := planTime + int64(offset*1000)
			if p%steps[i] != 0 {
				// 跳过，遍历下一个step
				logger.Debugf("跳过：目标模型任务[%s]当前计算时间点plantime:%d. 当前时间now:%d, 执行步长为%s", task.TaskID, planTime, ts, stepStr)
				continue
			}

			// 确定 look_back_delta。因为在管理端，语言是 promql 时，校验了时间窗口应为空
			lookBackDelta := stepStr
			if len(task.TimeWindows) == 1 {
				// slo 时，用时间窗口 timw_window（在创建SLO目标模型时把周期赋值到任务的time_window中了）作为即时查询的 look_back_delta
				lookBackDelta = task.TimeWindows[0]
			}

			// 用任务的 step 作为即时查询的 look_back_delta
			// 查询数据并组装kafka message
			// 请求uniquery，即时查询，
			objectiveData, err := mtService.uAccess.GetObjectiveModelData(ctx, task.ModelID, interfaces.MetricModelQuery{
				IsInstantQuery: true,
				Time:           planTime,
				LookBackDelta:  lookBackDelta, // 对于slo来说，此处是目标模型配置的周期。kpi就用自己的step。
			})
			logger.Debugf("查询的目标模型为【%s】,当前step【%s】查询的步长点为【%d】,look_back_delta=%s", task.ModelID, stepStr, planTime, lookBackDelta)
			if err != nil {
				return "", fmt.Errorf("目标模型任务[%s], 查询参数: time=%d,look_back_delta=%s, 获取指标数据失败[%s]. ", task.TaskID, planTime, lookBackDelta, err.Error())
			}

			//  组装kafka messages
			if objectiveData.Datas != nil {
				// 3.2.数据转换为 metric 数据格式. category, 任务名称，时间窗口
				err = mtService.objectiveDataTransferToMessage(planTime-steps[i], objectiveData, task, stepStr, indexBases[0], &messages)
				if err != nil {
					return "", fmt.Errorf("目标模型任务[%s], 时间窗口[%s]把数据转换为 kafka massage 失败[%s]. ", task.TaskID, stepStr, err.Error())
				}
			}

		}
		planTime += interfaces.MIN_STEP

		// 每次个时间点上提交一次kafka，不等批次，一个步长点提交一次数据，并更新planTime
		if len(messages) > 0 {
			err = mtService.flushToKafka(task, messages)
			if err != nil {
				return "", fmt.Errorf("目标模型任务[%s], 把数据发送到 kafka 失败[%s]. ", task.TaskID, err.Error())
			}

			// 更新数据库中任务的计划执行时间plan_time 按5m修正后按任务的steps的最大公约数递增，任务的状态为任务执行中执行成功
			taskPlanTime = planTime
			err = mtService.mmAccess.UpdateTaskAttributesById(ctx, task.TaskID,
				interfaces.MetricTask{PlanTime: taskPlanTime, ExeccuteStatus: interfaces.SCHEDULE_SYNC_STATUS_SUCCESS})
			if err != nil {
				return "", fmt.Errorf("目标模型任务[%s], 更新任务计划时间失败[%s]. ", task.TaskID, err.Error())
			}

			msgTotal += len(messages)
			logger.Debugf("service: 目标模型[%s]执行完成. 共发送[%d]条数据到kafka。", task.TaskID, len(messages))
		}
	}

	logger.Debugf("service: 目标模型[%s]执行完成. 共发送[%d]条数据到kafka。", task.TaskID, msgTotal)
	return fmt.Sprintf("service: 目标模型[%s]执行完成. 共发送[%d]条数据到kafka。", task.TaskID, msgTotal), nil
}

// 发送当前窗口的数据到kafka
func (mtService *metricTaskService) flushToKafka(task interfaces.MetricTask,
	messages []*kafka.Message) (err error) {

	// 4 把查询结果写kafka topic 需要注意
	// 4.1 创建生产者,穿一个uniqId
	topic := fmt.Sprintf(interfaces.MODEL_PERSIST_INPUT, mtService.appSetting.MQSetting.Tenant)
	uniqueId := fmt.Sprintf("%s-%s", task.TaskID, topic)
	producer, err := mtService.kAccess.NewTrxProducer(uniqueId)
	if err != nil {
		return err
	}
	defer producer.Close()

	// 4.2 消费数据
	err = mtService.kAccess.DoProduceMsgToKafka(producer, messages)
	if err != nil {
		return err
	}
	return nil
}

// 根据时间窗口和计划执行时间、当前时间来计算查询数据的start，end和step
// func getStartEndTime(window string, planTime int64, endT time.Time) (int64, int64, string, error) {
// 	// planTime / window * window, now / window * window
// 	// startTime := time.UnixMilli(planTime)

// 	var stepV int64
// 	switch window {
// 	// case interfaces.PREVIOUS_HOUR:
// 	// 	// 保留年月日时，分秒置0
// 	// 	start := startTime.Truncate(time.Hour)
// 	// 	end := endT.Truncate(time.Hour)
// 	// 	if start == end {
// 	// 		return 0, 0, "", fmt.Errorf("开始结束相等，不处理")
// 	// 	}
// 	// 	return start.UnixMilli(), end.UnixMilli(), "1h", nil
// 	// case interfaces.PREVIOUS_DAY:
// 	// 	// 将小时、分钟和秒部分设置为零
// 	// 	year, month, day := startTime.Date()
// 	// 	start := time.Date(year, month, day, 0, 0, 0, 0, time.Local)

// 	// 	year, month, day = endT.Date()
// 	// 	end := time.Date(year, month, day, 0, 0, 0, 0, time.Local)

// 	// 	if start == end {
// 	// 		return 0, 0, "", fmt.Errorf("开始结束相等，不处理")
// 	// 	}
// 	// 	return start.UnixMilli(), end.UnixMilli(), "1d", nil
// 	// case interfaces.PREVIOUS_WEEK:
// 	// year, month, day := startTime.Date()
// 	// start := time.Date(year, month, day, 0, 0, 0, 0, time.Local)

// 	// year, month, day = endT.Date()
// 	// end := time.Date(year, month, day, 0, 0, 0, 0, time.Local)

// 	// startDay := int(start.Weekday())
// 	// endDay := int(end.Weekday())
// 	// // 减去天数，得到星期一的日期
// 	// start = start.AddDate(0, 0, -(7+startDay-1)%7)
// 	// end = end.AddDate(0, 0, -(7+endDay-1)%7)
// 	// if start == end {
// 	// 	return 0, 0, "", fmt.Errorf("开始结束相等，不处理")
// 	// }
// 	// return start.UnixMilli(), end.UnixMilli(), "7d", nil
// 	// case interfaces.PREVIOUS_MONTH:
// 	// 	// 将天、小时、分钟和秒部分设置为零
// 	// 	start := time.Date(startTime.Year(), startTime.Month(), 1, 0, 0, 0, 0, time.Local)
// 	// 	end := time.Date(endTime.Year(), endTime.Month(), 1, 0, 0, 0, 0, time.Local)
// 	// 	return start.UnixMilli(), end.UnixMilli(), nil
// 	default:
// 		step, err := common.ParseDuration(window, common.DurationDayHourMinuteRE, true)
// 		if err != nil {
// 			logger.Errorf("Failed to parse schedule duration, err: %v", err.Error())
// 			return 0, 0, "", err
// 		}
// 		stepV = int64(step / (time.Millisecond / time.Nanosecond))

// 		// 向下取整
// 		start := int64(math.Floor(float64(planTime)/float64(stepV))) * stepV
// 		end := int64(math.Floor(float64(endT.UnixNano()/int64(time.Millisecond/time.Nanosecond))/float64(stepV))) * stepV

// 		return start, end, window, nil
// 	}
// }

// 数据转换为 metric 数据格式. category, 任务名称，时间窗口
func (mtService *metricTaskService) metricDataTransferToMessage(stepi int64, metricData interfaces.UniResponse, window string, task interfaces.MetricTask,
	step string, indexbase interfaces.IndexBase, messages *[]*kafka.Message) error {
	// 数据转成json
	// messages := make([]*kafka.Message, 0)
	topic := fmt.Sprintf(interfaces.MODEL_PERSIST_INPUT, mtService.appSetting.MQSetting.Tenant)
	// metricname := fmt.Sprintf("metrics.%s", task.MeasureName)

	for _, data := range metricData.Datas {
		message := make(map[string]interface{})
		// 补齐元字段： type: indexbase; __index_base: indexbase; __data_type: 索引库的data_type
		message["category"] = "metric"
		message["__data_type"] = indexbase.DataType
		message["__index_base"] = indexbase.BaseType
		message["type"] = indexbase.BaseType

		data.Labels["task_name"] = task.TaskName
		data.Labels["step"] = step
		// promql 时不能配置timewindows,dsl时才有time_window
		if len(task.TimeWindows) != 0 {
			data.Labels["time_window"] = window
		}

		message["labels"] = data.Labels

		labelNames := make([]string, 0)
		labelMap := make(map[string]string)
		for k, v := range message["labels"].(map[string]string) {
			labelNames = append(labelNames, k)
			labelMap[k] = v
		}
		sort.Strings(labelNames)
		var builder strings.Builder
		builder.WriteString(task.MeasureName)
		for _, ln := range labelNames {
			builder.WriteString(",")
			builder.WriteString(ln)
			builder.WriteString("=")
			builder.WriteString(labelMap[ln])
		}
		md5Hasher := md5.New()
		md5Hasher.Write([]byte(builder.String()))
		hashed := md5Hasher.Sum(nil)
		docId := hex.EncodeToString(hashed)
		message["__id"] = fmt.Sprintf("%d-%s", stepi, docId) // id生成规则变更为 时间戳-tsid(tsid包含了度量名称+labels_str)

		if task.MeasureName == "" {
			logger.Errorf("metric task [%s]'s measure name is empty", task.ModelID)
		}

		// 时间和值
		for _, v := range data.Values {
			if v == nil || v == "+Inf" || v == "-Inf" || v == "NaN" {
				continue
			}
			// t := time.UnixMilli(stepi).Format(time.RFC3339)
			// logger.Debugf("timestamp[%v]转换为RFC3339的后为[%s]。", stepi, t)
			message["@timestamp"] = stepi
			message["metrics"] = map[string]map[string]float64{
				"__m": {
					strings.Replace(task.MeasureName, "__m.", "", -1): v.(float64),
				},
			}

			bytes, err := json.Marshal(message)
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
	}
	return nil
}

// 数据转换为 metric 数据格式. category, 任务名称，时间窗口
func (mtService *metricTaskService) objectiveDataTransferToMessage(stepi int64, objectiveData interfaces.ObjectiveModelUniResponse, task interfaces.MetricTask,
	step string, indexbase interfaces.IndexBase, messages *[]*kafka.Message) error {

	// 数据转成json
	switch objectiveData.Model.ObjectiveType {
	case interfaces.SLO:
		// 把 datas 转成 SLOObjectiveData
		var sloDatas []interfaces.SLOObjectiveData
		jsonData, err := json.Marshal(objectiveData.Datas)
		if err != nil {
			return err
		}
		err = json.Unmarshal(jsonData, &sloDatas)
		if err != nil {
			return err
		}
		objectiveData.Datas = sloDatas
		err = mtService.sloDataTransferToMessage(stepi, sloDatas, objectiveData.Model, task, step, indexbase, messages)
		if err != nil {
			return err
		}
	case interfaces.KPI:
		// 把 datas 转成 SLOObjectiveData
		var kpiDatas []interfaces.KPIObjectiveData
		jsonData, err := json.Marshal(objectiveData.Datas)
		if err != nil {
			return err
		}
		err = json.Unmarshal(jsonData, &kpiDatas)
		if err != nil {
			return err
		}
		objectiveData.Datas = kpiDatas
		err = mtService.kpiDataTransferToMessage(stepi, kpiDatas, objectiveData.Model, task, step, indexbase, messages)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported objective type %s", objectiveData.Model.ObjectiveType)
	}

	return nil
}

// 数据转换为 metric 数据格式. category, 任务名称，时间窗口
func (mtService *metricTaskService) sloDataTransferToMessage(stepi int64, objectiveData []interfaces.SLOObjectiveData, model interfaces.ObjectiveModel,
	task interfaces.MetricTask, step string, indexbase interfaces.IndexBase, messages *[]*kafka.Message) error {

	// 数据转成json
	topic := fmt.Sprintf(interfaces.MODEL_PERSIST_INPUT, mtService.appSetting.MQSetting.Tenant)

	for _, data := range objectiveData {
		message := make(map[string]interface{})
		// 补齐元字段： type: indexbase; __index_base: indexbase; __data_type: 索引库的data_type
		message["category"] = "metric"
		message["__data_type"] = indexbase.DataType
		message["__index_base"] = indexbase.BaseType
		message["type"] = indexbase.BaseType

		data.Labels["objective_model_id"] = task.ModelID
		data.Labels["objective_model_name"] = model.ModelName
		data.Labels["step"] = step

		message["labels"] = data.Labels

		labelNames := make([]string, 0)
		labelMap := make(map[string]string)
		for k, v := range message["labels"].(map[string]string) {
			labelNames = append(labelNames, k)
			labelMap[k] = v
		}
		sort.Strings(labelNames)
		var builder strings.Builder
		builder.WriteString(task.MeasureName)
		for _, ln := range labelNames {
			builder.WriteString(",")
			builder.WriteString(ln)
			builder.WriteString("=")
			builder.WriteString(labelMap[ln])
		}
		md5Hasher := md5.New()
		md5Hasher.Write([]byte(builder.String()))
		hashed := md5Hasher.Sum(nil)
		docId := hex.EncodeToString(hashed)
		message["__id"] = fmt.Sprintf("%d-%s", stepi, docId) // id生成规则变更为 时间戳-tsid(tsid包含了度量名称+labels_str)

		// 时间和值。值是数组，但是是size为1的，因为是即时查询
		for i := range data.SLI {
			if data.SLI != nil && data.SLI[i] == nil || data.SLI[i] == "+Inf" || data.SLI[i] == "-Inf" || data.SLI[i] == "NaN" {
				continue
			}

			t := time.UnixMilli(stepi).Format(time.RFC3339)
			logger.Debugf("timestamp[%v]转换为RFC3339的后为[%s]。", stepi, t)
			message["@timestamp"] = t

			// 把所有的数据
			valueMap := make(map[string]any)
			valueMap[interfaces.SLO_SLI] = data.SLI[i]
			valueMap[interfaces.SLO_OBJECTIVE] = data.Objective[i]

			if data.Good != nil && data.Good[i] != nil && data.Good[i] != "+Inf" && data.Good[i] != "-Inf" && data.Good[i] != "NaN" {
				valueMap[interfaces.SLO_GOOD] = data.Good[i]
			}
			if data.Total != nil && data.Total[i] != nil && data.Total[i] != "+Inf" && data.Total[i] != "-Inf" && data.Total[i] != "NaN" {
				valueMap[interfaces.SLO_TOTAL] = data.Total[i]
			}
			valueMap[interfaces.SLO_ACHIEVEMENT_RATE] = data.AchiveRate[i]
			if data.TotalErrorBudget != nil && data.TotalErrorBudget[i] != nil && data.TotalErrorBudget[i] != "+Inf" && data.TotalErrorBudget[i] != "-Inf" && data.TotalErrorBudget[i] != "NaN" {
				valueMap[interfaces.SLO_TOTAL_ERROR_BUDGET] = data.TotalErrorBudget[i]
			}
			if data.LeftErrorBudget != nil && data.LeftErrorBudget[i] != nil && data.LeftErrorBudget[i] != "+Inf" && data.LeftErrorBudget[i] != "-Inf" && data.LeftErrorBudget[i] != "NaN" {
				valueMap[interfaces.SLO_LEFT_ERROR_BUDGET] = data.LeftErrorBudget[i]
			}
			if data.BurnRate != nil && data.BurnRate[i] != nil && data.BurnRate[i] != "+Inf" && data.BurnRate[i] != "-Inf" && data.BurnRate[i] != "NaN" {
				valueMap[interfaces.SLO_BURN_RATE] = data.BurnRate[i]
			}
			if data.Status != nil && data.Status[i] != "" {
				valueMap[interfaces.SLO_STATUS] = data.Status[i]
				valueMap[interfaces.SLO_STATUS_CODE] = data.StatusCode[i]
			}
			valueMap[interfaces.SLO_PERIOD] = data.Period[i]

			message["metrics"] = map[string]map[string]any{
				"__m": valueMap,
			}

			bytes, err := json.Marshal(message)
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
	}
	return nil
}

// 数据转换为 metric 数据格式. category, 任务名称，时间窗口
func (mtService *metricTaskService) kpiDataTransferToMessage(stepi int64, objectiveData []interfaces.KPIObjectiveData, model interfaces.ObjectiveModel,
	task interfaces.MetricTask, step string, indexbase interfaces.IndexBase, messages *[]*kafka.Message) error {

	// 数据转成json
	topic := fmt.Sprintf(interfaces.MODEL_PERSIST_INPUT, mtService.appSetting.MQSetting.Tenant)

	for _, data := range objectiveData {
		message := make(map[string]interface{})
		// 补齐元字段： type: indexbase; __index_base: indexbase; __data_type: 索引库的data_type
		message["category"] = "metric"
		message["__data_type"] = indexbase.DataType
		message["__index_base"] = indexbase.BaseType
		message["type"] = indexbase.BaseType

		data.Labels["objective_model_id"] = task.ModelID
		data.Labels["objective_model_name"] = model.ModelName
		data.Labels["step"] = step

		message["labels"] = data.Labels

		labelNames := make([]string, 0)
		labelMap := make(map[string]string)
		for k, v := range message["labels"].(map[string]string) {
			labelNames = append(labelNames, k)
			labelMap[k] = v
		}
		sort.Strings(labelNames)
		var builder strings.Builder
		builder.WriteString(task.MeasureName)
		for _, ln := range labelNames {
			builder.WriteString(",")
			builder.WriteString(ln)
			builder.WriteString("=")
			builder.WriteString(labelMap[ln])
		}
		md5Hasher := md5.New()
		md5Hasher.Write([]byte(builder.String()))
		hashed := md5Hasher.Sum(nil)
		docId := hex.EncodeToString(hashed)
		message["__id"] = fmt.Sprintf("%d-%s", stepi, docId) // id生成规则变更为 时间戳-tsid(tsid包含了度量名称+labels_str)

		// 时间和值。值是数组，但是是size为1的，因为是即时查询
		for i := range data.KPI {
			if data.KPI != nil && data.KPI[i] == nil || data.KPI[i] == "+Inf" || data.KPI[i] == "-Inf" || data.KPI[i] == "NaN" {
				continue
			}

			t := time.UnixMilli(stepi).Format(time.RFC3339)
			logger.Debugf("timestamp[%v]转换为RFC3339的后为[%s]。", stepi, t)
			message["@timestamp"] = t

			// 把所有的数据
			valueMap := make(map[string]any)
			valueMap[interfaces.KPI_KPI] = data.KPI[i]
			valueMap[interfaces.KPI_OBJECTIVE] = data.Objective[i]
			valueMap[interfaces.KPI_ACHIEVEMENT_RATE] = data.AchiveRate[i]
			valueMap[interfaces.KPI_SCORE] = data.KPIScore[i]
			valueMap[interfaces.KPI_ASSOCIATE_METRIC_NUM] = data.AssociateMetricNums[i]
			if data.Status[i] != "" {
				valueMap[interfaces.KPI_STATUS] = data.Status[i]
				valueMap[interfaces.KPI_STATUS_CODE] = data.StatusCode[i]
			}

			message["metrics"] = map[string]map[string]any{
				"__m": valueMap,
			}

			bytes, err := json.Marshal(message)
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
	}
	return nil
}
