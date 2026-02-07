// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package metric_model

// import (
// 	"context"
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"time"

// 	xxlsdk "devops.aishu.cn/AISHUDevOps/DIP/_git/mdl-XxlJobSDK-Go.git/sdk"
// 	"devops.aishu.cn/AISHUDevOps/DIP/_git/mdl-XxlJobSDK-Go.git/sdk/constant"
// 	"devops.aishu.cn/AISHUDevOps/DIP/_git/mdl-XxlJobSDK-Go.git/sdk/vo"
// 	"github.com/kweaver-ai/kweaver-go-lib/logger"

// 	"data-model/common"
// 	"data-model/interfaces"
// )

// const (
// 	METRIC_APP_NAME = "metric-persist-jobs"
// 	ExecutorHandler = "metricTaskhandler"
// )

// var (
// 	// 非阻塞通道
// 	TaskCh = make(chan []string, 100)

// 	sdk       *xxlsdk.Client
// 	username                = "test"
// 	password                = "testpwd"
// 	jobGroups []vo.JobGroup = []vo.JobGroup{
// 		{
// 			AppName:     METRIC_APP_NAME,
// 			Title:       "metricJobs",
// 			AddressType: constant.JOB_GROUP_ADDR_TYPE_AUTO,
// 		},
// 	}
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

// func (mmService *metricModelService) SyncSchedule() {
// 	ctx := context.Background()
// 	// 异步初始化
// 	go InitClient(mmService.appSetting.XxlJobAdminUrl)

// 	tick := time.NewTicker(time.Second * time.Duration(30))
// 	for {
// 		select {
// 		case taskids := <-TaskCh:
// 			logger.Debugf("从通道中获取到任务,id为 %v", taskids)
// 			// 1. 按任务id从数据库中查询任务信息
// 			tasks, err := mmService.mmtService.GetMetricTasksByTaskIDs(ctx, taskids)
// 			if err != nil {
// 				// 记录日志，结束
// 				logger.Error(err)
// 				continue
// 			}
// 			logger.Debugf("同步任务个数：%d", len(tasks))
// 			// 2. 根据任务实际状态做相应的操作
// 			err = mmService.SyncStatusFromXXLJob(ctx, tasks)
// 			if err != nil {
// 				// 记录可观测性日志，结束
// 				logger.Error(err)
// 				continue
// 			}
// 		case <-tick.C:
// 			logger.Debug("30秒超时通道触发")
// 			// 超时，重新查询数据库处理进行中的任务
// 			// 1. 查询处理中的任务
// 			tasks, err := mmService.mmtService.GetProcessingMetricTasks(ctx)
// 			if err != nil {
// 				// 记录可观测性日志，结束
// 				logger.Error(err)
// 				continue
// 			}
// 			logger.Debugf("同步任务个数：%d", len(tasks))
// 			// 2. 根据任务实际状态做相应的操作
// 			err = mmService.SyncStatusFromXXLJob(ctx, tasks)
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
// func (mmService *metricModelService) SyncStatusFromXXLJob(ctx context.Context, tasks []interfaces.MetricTask) error {
// 	// 遍历，同步任务
// 	var errs error
// 	for _, task := range tasks {
// 		if task.ModuleType == interfaces.MODULE_TYPE_METRIC_MODEL {
// 			// 为新建 更新的任务获取 measure name，如果是删除的任务，此时会报模型不存在的错误
// 			if task.ScheduleSyncStatus == interfaces.SCHEDULE_SYNC_STATUS_CREATE || task.ScheduleSyncStatus == interfaces.SCHEDULE_SYNC_STATUS_UPDATE {
// 				var model interfaces.MetricModel
// 				model, err := mmService.GetMetricModelByModelID(ctx, task.ModelID)
// 				if err != nil {
// 					logger.Error(err)
// 					errs = errors.Join(errs, err)
// 					continue
// 				}
// 				task.MeasureName = model.MeasureName
// 			}
// 		}

// 		var taskBytes []byte
// 		taskBytes, err := json.Marshal(task)
// 		if err != nil {
// 			logger.Error(err)
// 			errs = errors.Join(errs, err)
// 			continue
// 		}
// 		// 创建中
// 		switch task.ScheduleSyncStatus {
// 		case interfaces.SCHEDULE_SYNC_STATUS_CREATE:
// 			// 1.创建任务
// 			resp, err := AddJob(task, string(taskBytes))
// 			if err != nil {
// 				logger.Errorf("AddJob [%s] failed, error: %v", task.TaskName, err)
// 				errs = errors.Join(errs, err)
// 				continue
// 			}
// 			// 2. 启动任务
// 			err = StartJob(resp.ID)
// 			if err != nil {
// 				logger.Errorf("StartJob [%s] failed, error: %v", task.TaskName, err)
// 				errs = errors.Join(errs, err)
// 				continue
// 			}

// 			// 如果任务的更新时间小于等于当前内存中的对象的更新时间，那么就更新状态未完成，否则，任务被修改过，不能把任务的状态更新为完成。
// 			// 3. 创建成功后更新任务状态。update set TaskScheduleID = resp.ID, status=完成 where taskID = ? and update_time <= task.update_time
// 			task.TaskScheduleID = resp.ID
// 			// task.UpdateTime = common.GenerateUpdateTime()
// 			err = mmService.mmtService.UpdateMetricTaskStatusInFinish(ctx, task)
// 			if err != nil {
// 				logger.Errorf("AddJob success, but UpdateMetricTaskStatusInFinish [%s] failed, error: %v", task.TaskName, err)
// 				errs = errors.Join(errs, err)
// 				continue
// 			}
// 		case interfaces.SCHEDULE_SYNC_STATUS_UPDATE:
// 			// 修改中的任务，需要判断调度id是否为空
// 			if task.TaskScheduleID == 0 {
// 				// 调度id为空，则创建任务
// 				resp, err := AddJob(task, string(taskBytes))
// 				if err != nil {
// 					logger.Errorf("task status is [%d], but task schedule id is 0, need to AddJob [%s] failed, error: %v",
// 						task.ScheduleSyncStatus, task.TaskName, err)
// 					errs = errors.Join(errs, err)
// 					continue
// 				}
// 				// 2. 启动任务
// 				err = StartJob(resp.ID)
// 				if err != nil {
// 					logger.Errorf("task status is [%d], but task schedule id is 0, need to StartJob [%s] failed, error: %v",
// 						task.ScheduleSyncStatus, task.TaskName, err)
// 					errs = errors.Join(errs, err)
// 					continue
// 				}
// 				task.TaskScheduleID = resp.ID
// 			} else {
// 				// 修改任务
// 				err = UpdateJob(task, string(taskBytes))
// 				if err != nil {
// 					logger.Errorf("UpdateJob [%s] failed, error: %v", task.TaskName, err)
// 					errs = errors.Join(errs, err)
// 					continue
// 				}
// 			}
// 			// 操作成功之后，更新任务状态， update status where taskID = ? and update_time <= task.update_time
// 			// task.UpdateTime = common.GenerateUpdateTime()
// 			err = mmService.mmtService.UpdateMetricTaskStatusInFinish(ctx, task)
// 			if err != nil {
// 				logger.Errorf("AddJob or UpdateJob success, but UpdateMetricTaskStatusInFinish [%s] failed, error: %v", task.TaskName, err)
// 				errs = errors.Join(errs, err)
// 				continue
// 			}
// 		case interfaces.SCHEDULE_SYNC_STATUS_DELETE:
// 			if task.TaskScheduleID != 0 {
// 				// 调度id不为空，则停止删除任务
// 				err = StopJob(task.TaskScheduleID)
// 				if err != nil {
// 					logger.Errorf("StopJob [%s] failed, error: %v", task.TaskName, err)
// 					errs = errors.Join(errs, err)
// 					continue
// 				}
// 				err = DeleteJob(task.TaskScheduleID)
// 				if err != nil {
// 					logger.Errorf("DeleteJob [%s] failed, error: %v", task.TaskName, err)
// 					errs = errors.Join(errs, err)
// 					continue
// 				}
// 			}
// 			// 物理删除任务
// 			err = mmService.mmtService.DeleteMetricTaskByTaskID(ctx, task.TaskID)
// 			if err != nil {
// 				logger.Errorf("DeleteJob success, but DeleteMetricTaskByTaskID [%s] failed, error: %v", task.TaskName, err)
// 				errs = errors.Join(errs, err)
// 				continue
// 			}
// 		default:
// 			// do nothing
// 		}
// 	}
// 	return errs
// }

// // 添加任务
// func AddJob(task interfaces.MetricTask, taskStr string) (vo.JobInfoAddResp, error) {
// 	expr := task.Schedule.Expression
// 	if task.Schedule.Type == interfaces.SCHEDULE_TYPE_FIXED {
// 		durationV, err := common.ParseDuration(task.Schedule.Expression, common.DurationDayHourMinuteRE, true)
// 		if err != nil {
// 			return vo.JobInfoAddResp{}, err
// 		}
// 		if durationV <= 0 {
// 			return vo.JobInfoAddResp{}, fmt.Errorf("zero or negative schedule expression is not accepted. Try a positive integer")
// 		}
// 		// 毫秒转秒
// 		stepV := int64(durationV/(time.Millisecond/time.Nanosecond)) / 1000
// 		expr = fmt.Sprintf("%d", stepV)
// 	}
// 	req := vo.JobInfoAddReq{
// 		AppName:                METRIC_APP_NAME,
// 		JobDesc:                fmt.Sprintf("%s:%s", task.TaskName, task.Comment),
// 		Author:                 "admin",
// 		AlertEmail:             "",
// 		ScheduleType:           task.Schedule.Type,
// 		ScheduleConf:           expr,
// 		GlueType:               constant.JOB_TYPE_BEAN,
// 		ExecutorHandler:        ExecutorHandler,
// 		ExecutorParam:          taskStr,
// 		ExecutorRouteStrategy:  constant.JOB_EXEC_ROUTE_STRATEGY_FIRST,
// 		ChildJobID:             "",
// 		MisfireStrategy:        constant.JOB_MISFIRE_STRATEGY_DO_NOTHING,
// 		ExecutorBlockStrategy:  constant.JOB_EXEC_BLOCK_STRATEGY_DISCARD_LATER,
// 		ExecutorTimeout:        interfaces.XXL_JOB_TASK_EXEC_TIMEOUT,
// 		ExecutorFailRetryCount: 3,
// 	}
// 	return sdk.JobInfoSvc.Add(req)
// }

// // 启动任务
// func StartJob(jobID int) error {
// 	req := vo.JobInfoStartReq{
// 		ID: jobID,
// 	}
// 	return sdk.JobInfoSvc.Start(req)
// }

// // 停止任务
// func StopJob(jobID int) error {
// 	req := vo.JobInfoStopReq{
// 		ID: jobID,
// 	}
// 	return sdk.JobInfoSvc.Stop(req)
// }

// // 删除任务
// func DeleteJob(jobID int) error {
// 	req := vo.JobInfoDelReq{
// 		ID: jobID,
// 	}
// 	return sdk.JobInfoSvc.Delete(req)
// }

// // 更新
// func UpdateJob(task interfaces.MetricTask, taskStr string) error {
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
// 		AppName:                METRIC_APP_NAME,
// 		JobDesc:                fmt.Sprintf("%s:%s", task.TaskName, task.Comment),
// 		Author:                 "admin",
// 		AlertEmail:             "",
// 		ScheduleType:           task.Schedule.Type,
// 		ScheduleConf:           expr,
// 		ExecutorHandler:        ExecutorHandler,
// 		ExecutorParam:          taskStr,
// 		ExecutorRouteStrategy:  constant.JOB_EXEC_ROUTE_STRATEGY_FIRST,
// 		ChildJobID:             "",
// 		MisfireStrategy:        constant.JOB_MISFIRE_STRATEGY_DO_NOTHING,
// 		ExecutorBlockStrategy:  constant.JOB_EXEC_BLOCK_STRATEGY_DISCARD_LATER,
// 		ExecutorTimeout:        interfaces.XXL_JOB_TASK_EXEC_TIMEOUT,
// 		ExecutorFailRetryCount: 3,
// 	}
// 	return sdk.JobInfoSvc.Update(req)
// }
