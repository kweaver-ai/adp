// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package event_model

// import (
// 	"context"
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"strings"
// 	"time"

// 	xxlsdk "devops.aishu.cn/AISHUDevOps/DIP/_git/mdl-XxlJobSDK-Go.git/sdk"
// 	"devops.aishu.cn/AISHUDevOps/DIP/_git/mdl-XxlJobSDK-Go.git/sdk/constant"
// 	"devops.aishu.cn/AISHUDevOps/DIP/_git/mdl-XxlJobSDK-Go.git/sdk/vo"
// 	"github.com/kweaver-ai/kweaver-go-lib/logger"

// 	"data-model/common"
// 	"data-model/interfaces"
// 	"data-model/logics/metric_model"
// )

// const (
// 	EVENT_APP_NAME       = "event-persist-jobs"
// 	EventExecutorHandler = "eventTaskhandler"
// )

// var (
// 	// 非阻塞通道
// 	sdk       *xxlsdk.Client
// 	username                = "test"
// 	password                = "testpwd"
// 	jobGroups []vo.JobGroup = []vo.JobGroup{
// 		{
// 			AppName:     EVENT_APP_NAME,
// 			Title:       "eventJobs",
// 			AddressType: constant.JOB_GROUP_ADDR_TYPE_AUTO,
// 		},
// 	}

// 	eventTaskCh = make(chan uint64, 100)
// )

// // 初始化xxljob的client。
// func InitClient(baseUrl string) {
// 	for {
// 		var err error
// 		sdk, err = xxlsdk.NewClient(baseUrl, username, password)
// 		if err != nil {
// 			logger.Error(err)
// 			time.Sleep(time.Second * 5)
// 		}
// 		err = sdk.Init(jobGroups)
// 		if err != nil {
// 			logger.Error(err)
// 			time.Sleep(time.Second * 5)
// 		} else {
// 			break
// 		}
// 		logger.Error("init sdk success")
// 	}
// }

// func (emService *eventModelService) SyncSchedule() {
// 	ctx := context.Background()
// 	// 异步初始化
// 	go InitClient(emService.appSetting.XxlJobAdminUrl)

// 	tick := time.NewTicker(time.Minute * time.Duration(1))
// 	for {
// 		select {
// 		case taskid := <-eventTaskCh:
// 			logger.Debugf("从通道中获取到任务,id为 %v", taskid)
// 			// 1. 按任务id从数据库中查询任务信息
// 			task, err := emService.GetEventTaskByTaskID(ctx, taskid)
// 			if err != nil {
// 				// 记录日志，结束
// 				logger.Error(err)
// 				continue
// 			}
// 			// 2. 根据任务实际状态做相应的操作
// 			err = emService.SyncStatusFromXXLJob(ctx, []interfaces.EventTask{task})
// 			if err != nil {
// 				logger.Error(err)
// 				continue
// 			}
// 		case <-tick.C:
// 			logger.Debug("1分钟超时通道触发")
// 			// 超时，重新查询数据库处理进行中的任务
// 			// 1. 查询处理中的任务
// 			tasks, err := emService.GetProcessingEventTasks(ctx)
// 			if err != nil {
// 				// 记录可观测性日志，结束
// 				logger.Error(err)
// 				continue
// 			}
// 			logger.Debugf("同步任务个数：%d", len(tasks))
// 			// 2. 根据任务实际状态做相应的操作
// 			err = emService.SyncStatusFromXXLJob(ctx, tasks)
// 			if err != nil {
// 				// 记录可观测性日志，结束
// 				logger.Error(err)
// 				continue
// 			}
// 		}
// 	}
// }

// // 若是修改中的任务，调度id为空，那么发起一个创建任务的请求。若调度id不为空，则发起一个修改任务的请求
// // 若是删除中的任务，调度id为空，那么直接删除任务数据
// // 调度xxljob等待其返回结果，成功就更新数据库，失败就不更新数据库
// func (emService *eventModelService) SyncStatusFromXXLJob(ctx context.Context, tasks []interfaces.EventTask) error {
// 	// 遍历，同步任务
// 	for _, task := range tasks {
// 		taskBytes, err := json.Marshal(struct {
// 			ModelID uint64 `json:"model_id,string"`
// 		}{task.ModelID})
// 		if err != nil {
// 			logger.Error(err)
// 			return err
// 		}
// 		// 创建中
// 		switch task.ScheduleSyncStatus {
// 		case interfaces.SCHEDULE_SYNC_STATUS_CREATE:
// 			// 1.创建任务
// 			model, httpErr := emService.GetEventModelByID(ctx, task.ModelID)
// 			if httpErr != nil {
// 				logger.Errorf("GetEventModelByID failed when addEventJob,error: %v", httpErr)
// 				return httpErr
// 			}
// 			resp, err := addEventJob(task, string(taskBytes), model.EventModelName)
// 			if err != nil {
// 				logger.Errorf("addEventJob [%d] failed, error: %v", task.TaskID, err)
// 				return err
// 			}
// 			if model.IsActive == 1 && model.Status == 1 {
// 				// 2. 启动任务
// 				err = metric_model.StartJob(resp.ID)
// 				if err != nil {
// 					logger.Errorf("startJob [%d] failed, error: %v", task.TaskID, err)
// 					return err
// 				}
// 			}
// 			// 如果任务的更新时间小于等于当前内存中的对象的更新时间，那么就更新状态未完成，否则，任务被修改过，不能把任务的状态更新为完成。
// 			// 3. 创建成功后更新任务状态。update set TaskScheduleID = resp.ID, status=完成 where taskID = ? and update_time <= task.update_time
// 			task.TaskScheduleID = resp.ID
// 			task.UpdateTime = common.GenerateUpdateTime()
// 			err = emService.UpdateEventTaskStatusInFinish(ctx, task)
// 			if err != nil {
// 				logger.Errorf("addEventJob success, but UpdateEventTaskStatusInFinish [%d] failed, error: %v", task.TaskID, err)
// 				return err
// 			}
// 		case interfaces.SCHEDULE_SYNC_STATUS_UPDATE:
// 			// 修改中的任务，需要判断调度id是否为空
// 			if task.TaskScheduleID == 0 {
// 				// 调度id为空，则创建任务
// 				resp, err := addEventJob(task, string(taskBytes), fmt.Sprintf("%d", task.TaskID))
// 				if err != nil {
// 					logger.Errorf("task status is [%d], but task schedule id is 0, need to addJob [%d] failed, error: %v",
// 						task.ScheduleSyncStatus, task.TaskID, err)
// 					return err
// 				}

// 				// // 2. 启动任务
// 				// err = startJob(resp.ID)
// 				// if err != nil {
// 				// 	logger.Errorf("task status is [%d], but task schedule id is 0, need to startJob [%d] failed, error: %v",
// 				// 		task.ScheduleSyncStatus, task.TaskID, err)
// 				// 	return err
// 				// }
// 				task.TaskScheduleID = resp.ID
// 			} else {
// 				// 修改任务
// 				err := updateEventJob(task, string(taskBytes))
// 				if err != nil {
// 					logger.Errorf("updateEventJob [%d] failed, error: %v", task.TaskID, err)
// 					return err
// 				}
// 			}
// 			// 操作成功之后，更新任务状态， update status where taskID = ? and update_time <= task.update_time
// 			task.UpdateTime = common.GenerateUpdateTime()
// 			err = emService.UpdateEventTaskStatusInFinish(ctx, task)
// 			if err != nil {
// 				logger.Errorf("addEventJob or updateEventJob success, but UpdateEventTaskStatusInFinish [%d] failed, error: %v", task.TaskID, err)
// 				return err
// 			}
// 		case interfaces.SCHEDULE_SYNC_STATUS_DELETE:
// 			if task.TaskScheduleID != 0 {
// 				// 调度id不为空，则停止删除任务
// 				err := metric_model.StopJob(task.TaskScheduleID)
// 				if err != nil {
// 					logger.Errorf("stopJob [%d] failed, error: %v", task.TaskID, err)
// 					return err
// 				}
// 				err = metric_model.DeleteJob(task.TaskScheduleID)
// 				if err != nil {
// 					logger.Errorf("deleteJob [%d] failed, error: %v", task.TaskID, err)
// 					return err
// 				}
// 			}
// 			// 物理删除任务
// 			err := emService.DeleteEventTaskByTaskID(ctx, task.TaskID)
// 			if err != nil {
// 				logger.Errorf("deleteJob success, but DeleteEventTaskByTaskID [%d] failed, error: %v", task.TaskID, err)
// 				return err
// 			}
// 		default:
// 			// do nothing
// 		}
// 	}
// 	return nil
// }

// // 添加任务
// func addEventJob(task interfaces.EventTask, taskStr, modelName string) (vo.JobInfoAddResp, error) {
// 	expr := task.Schedule.Expression
// 	if task.Schedule.Type == interfaces.SCHEDULE_TYPE_FIXED {
// 		durationV, err := common.ParseDuration(task.Schedule.Expression, common.DurationDayHourMinuteRE, true)
// 		if err != nil {
// 			return vo.JobInfoAddResp{}, err
// 		}
// 		if durationV <= 0 {
// 			return vo.JobInfoAddResp{}, fmt.Errorf("zero or negative schedule expression is not accepted. Try a positive integer")
// 		}
// 		logger.Debugf("执行参数长度：%d; ", len(taskStr))
// 		logger.Debugf("执行参数：%v; ", taskStr)
// 		// 毫秒转秒
// 		stepV := int64(durationV/(time.Millisecond/time.Nanosecond)) / 1000
// 		expr = fmt.Sprintf("%d", stepV)
// 	}

// 	req := vo.JobInfoAddReq{
// 		AppName:                EVENT_APP_NAME,
// 		JobDesc:                modelName,
// 		Author:                 "admin",
// 		AlertEmail:             "",
// 		ScheduleType:           task.Schedule.Type,
// 		ScheduleConf:           expr,
// 		GlueType:               constant.JOB_TYPE_BEAN,
// 		ExecutorHandler:        EventExecutorHandler,
// 		ExecutorParam:          taskStr,
// 		ChildJobID:             strings.Join(task.DownstreamDependentTask, ","),
// 		MisfireStrategy:        constant.JOB_MISFIRE_STRATEGY_DO_NOTHING,
// 		ExecutorRouteStrategy:  task.DispatchConfig.RouteStrategy,
// 		ExecutorBlockStrategy:  task.DispatchConfig.BlockStrategy,
// 		ExecutorTimeout:        task.DispatchConfig.TimeOut,
// 		ExecutorFailRetryCount: task.DispatchConfig.FailRetryCount,
// 	}
// 	return sdk.JobInfoSvc.Add(req)
// }

// // 更新
// func updateEventJob(task interfaces.EventTask, taskStr string) error {
// 	expr := task.Schedule.Expression
// 	if task.Schedule.Type == interfaces.SCHEDULE_TYPE_FIXED {
// 		durationV, err := common.ParseDuration(task.Schedule.Expression, common.DurationDayHourMinuteRE, true)
// 		if err != nil {
// 			return err
// 		}
// 		if durationV <= 0 {
// 			return fmt.Errorf("zero or negative schedule expression is not accepted. Try a positive integer")
// 		}
// 		// 毫秒转秒
// 		stepV := int64(durationV/(time.Millisecond/time.Nanosecond)) / 1000
// 		expr = fmt.Sprintf("%d", stepV)
// 	}

// 	req := vo.JobInfoUpdateReq{
// 		ID:                     task.TaskScheduleID,
// 		AppName:                EVENT_APP_NAME,
// 		JobDesc:                fmt.Sprintf("%d", task.TaskID),
// 		Author:                 "admin",
// 		AlertEmail:             "",
// 		ScheduleType:           task.Schedule.Type,
// 		ScheduleConf:           expr,
// 		ExecutorHandler:        EventExecutorHandler,
// 		ExecutorParam:          taskStr,
// 		ChildJobID:             strings.Join(task.DownstreamDependentTask, ","),
// 		MisfireStrategy:        constant.JOB_MISFIRE_STRATEGY_DO_NOTHING,
// 		ExecutorRouteStrategy:  task.DispatchConfig.RouteStrategy,
// 		ExecutorBlockStrategy:  task.DispatchConfig.BlockStrategy,
// 		ExecutorTimeout:        task.DispatchConfig.TimeOut,
// 		ExecutorFailRetryCount: task.DispatchConfig.FailRetryCount,
// 	}
// 	return sdk.JobInfoSvc.Update(req)
// }
// func TrrigerEventJob(id int, taskStr string) error {

// 	req := vo.JobInfoTriggerReq{
// 		ID:            id,
// 		ExecutorParam: taskStr,
// 	}
// 	return sdk.JobInfoSvc.Trigger(req)
// }

// func (eventModelService *eventModelService) ValidateExecuteParam(ctx context.Context, executeParam map[string]any) (bool, error) {
// 	start, ok := executeParam["start"]
// 	var start_timestamp, end_timestamp float64
// 	if ok {
// 		start_timestamp = start.(float64)
// 	}
// 	end, ok := executeParam["end"]
// 	if ok {
// 		end_timestamp = end.(float64)
// 	}
// 	if start_timestamp >= end_timestamp && start_timestamp > 0 {
// 		return false, errors.New("执行参数不合理，填入的开始时间不能大于等于结束时间")
// 	}
// 	return true, nil
// }
