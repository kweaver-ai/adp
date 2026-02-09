// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package logics

import (
	"context"
	"data-model-job/common"
	"data-model-job/interfaces"
	"fmt"
	"sync"
	"time"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/robfig/cron/v3"
)

// Scheduler 任务调度器
type Scheduler struct {
	jobs map[string]*interfaces.JobInfo // 存储所有任务
	cron *cron.Cron                     // Cron 调度器
	mu   sync.Mutex                     // 保证线程安全
}

// NewScheduler 创建一个新的调度器
func NewScheduler() *Scheduler {
	return &Scheduler{
		jobs: make(map[string]*interfaces.JobInfo),
		cron: cron.New(cron.WithSeconds()), // 支持秒级精度
	}
}

// 添加任务
func (jService *jobService) addScheduleJob(jobInfo *interfaces.JobInfo) error {
	jService.scheduler.mu.Lock()
	defer jService.scheduler.mu.Unlock()

	if _, exists := jService.scheduler.jobs[jobInfo.JobId]; exists {
		logger.Errorf("job with ID %s already exists", jobInfo.JobId)
		return fmt.Errorf("job with ID %s already exists", jobInfo.JobId)
	}

	switch jobInfo.Schedule.Type {
	case interfaces.SCHEDULE_TYPE_FIXED:
		jobInfo.StopChan = make(chan struct{})
		step, err := common.ParseDuration(jobInfo.Schedule.Expression, common.DurationDayHourMinuteRE, true)
		if err != nil {
			logger.Infof("schedule job[%s]'s fixed expression %s is invalid.", jobInfo.JobId, jobInfo.Schedule.Expression)
			return err
		}

		ticker := time.NewTicker(step)
		jobInfo.Ticker = ticker
		go jService.runFixedJob(jobInfo)
	case interfaces.SCHEDULE_TYPE_CRON:
		id, err := jService.scheduler.cron.AddFunc(jobInfo.Schedule.Expression, func() {
			jService.executeJob(jobInfo)
		})
		if err != nil {
			logger.Errorf("failed to add cron job: %v", err)
			return fmt.Errorf("failed to add cron job: %v", err)
		}
		jobInfo.CronID = id
		jService.scheduler.cron.Start()
	default:
		return fmt.Errorf("unknown schedule type: %s", jobInfo.Schedule.Type)
	}

	jService.scheduler.jobs[jobInfo.JobId] = jobInfo
	logger.Infof("Job %s added: %s", jobInfo.ModuleType, jobInfo.JobId)
	return nil
}

// DeleteTask 删除任务
func (jService *jobService) deleteScheduleJob(jobInfo interfaces.JobInfo) error {
	jService.scheduler.mu.Lock()
	defer jService.scheduler.mu.Unlock()

	switch jobInfo.Schedule.Type {
	case interfaces.SCHEDULE_TYPE_FIXED:
		// 固定频率任务通过channel取消
		if jobInfo.StopChan != nil {
			// 先停止ticker，防止新的事件产生
			jobInfo.Ticker.Stop()
			close(jobInfo.StopChan)
			jobInfo.StopChan = nil
		}
	case interfaces.SCHEDULE_TYPE_CRON:
		jService.scheduler.cron.Remove(jobInfo.CronID)
	default:
		return fmt.Errorf("unknown schedule type: %s", jobInfo.Schedule.Type)
	}

	delete(jService.scheduler.jobs, jobInfo.JobId)
	logger.Infof("Job [%s] deleted: %s", jobInfo.ModuleType, jobInfo.JobId)
	return nil
}

// UpdateTask 更新任务
func (jService *jobService) updateScheduleJob(newJob interfaces.JobInfo) error {
	jService.scheduler.mu.Lock()
	defer jService.scheduler.mu.Unlock()

	job, exists := jService.scheduler.jobs[newJob.JobId]
	if !exists {
		return fmt.Errorf("job with ID %s not found", newJob.JobId)
	}

	// 先停止旧任务
	switch job.Schedule.Type {
	case interfaces.SCHEDULE_TYPE_FIXED:
		// 固定频率任务可以通过channel取消
		if job.StopChan != nil {
			// 先停止ticker，防止新的事件产生
			job.Ticker.Stop()
			close(job.StopChan)
			job.StopChan = nil
		}
	case interfaces.SCHEDULE_TYPE_CRON:
		jService.scheduler.cron.Remove(job.CronID)
	default:
		return fmt.Errorf("unknown schedule type: %s", job.Schedule.Type)
	}

	// 更新任务属性
	if job.MetricTask != nil {
		// 把原任务的creator留下. 因为目标模型的update task是以提交内容为准，不查数据库
		mCreator := job.MetricTask.Creator
		newJob.MetricTask.Creator = mCreator
	}
	job.JobType = newJob.JobType
	job.ModuleType = newJob.ModuleType
	job.JobConfig = newJob.JobConfig
	job.Schedule = newJob.Schedule
	job.DataView = newJob.DataView
	job.MetricTask = newJob.MetricTask
	job.EventTask = newJob.EventTask
	job.Schedule = newJob.Schedule

	// 重新启动任务
	switch job.Schedule.Type {
	case interfaces.SCHEDULE_TYPE_FIXED:
		// fixed： 启动一个新任务
		newJob.StopChan = make(chan struct{})
		step, err := common.ParseDuration(newJob.Schedule.Expression, common.DurationDayHourMinuteRE, true)
		if err != nil {
			logger.Infof("schedule job[%s]'s fixed expression %s is invalid.", newJob.JobId, newJob.Schedule.Expression)
			return err
		}
		ticker := time.NewTicker(step)
		newJob.Ticker = ticker

		go jService.runFixedJob(&newJob)
		jService.scheduler.jobs[newJob.JobId] = &newJob
	case interfaces.SCHEDULE_TYPE_CRON:
		id, err := jService.scheduler.cron.AddFunc(job.Schedule.Expression, func() {
			jService.executeJob(job)
		})
		if err != nil {
			return fmt.Errorf("failed to add cron job: %v", err)
		}
		job.CronID = id
		jService.scheduler.cron.Start()
	default:
		return fmt.Errorf("unknown schedule type: %s", job.Schedule.Type)
	}

	logger.Infof("Job [%s] updated: %s", job.ModuleType, job.JobId)
	return nil
}

// runFixedTask 运行固定频率任务
func (jService *jobService) runFixedJob(jobInfo *interfaces.JobInfo) {
	logger.Infof("Fixed job [%s] %s started.", jobInfo.ModuleType, jobInfo.JobId)

	defer jobInfo.Ticker.Stop()
	for {
		select {
		case <-jobInfo.Ticker.C:
			jService.executeJob(jobInfo)
		case <-jobInfo.StopChan:
			logger.Infof("Fixed job [%s] %s stopped\n", jobInfo.ModuleType, jobInfo.JobId)
			logger.Infof("Remaining jobs num: %d", len(jService.scheduler.jobs))
			return
		}
	}
}

// 执行metric任务
func (sjService *jobService) executeJob(jobInfo *interfaces.JobInfo) {
	logger.Infof("Executing [%s] job: %s", jobInfo.ModuleType, jobInfo.JobId)
	// 指标类任务调用MetricTaskExecutor
	switch jobInfo.ModuleType {
	case interfaces.MODULE_TYPE_METRIC_MODEL, interfaces.MODULE_TYPE_OBJECTIVE_MODEL:
		if jobInfo.MetricTask != nil {
			sjService.mtService.MetricTaskExecutor(context.Background(), *jobInfo.MetricTask)
		} else {
			logger.Errorf("Executing metric task : %s failed, because of metric task is empty.", jobInfo.JobId)
		}

	case interfaces.MODULE_TYPE_EVENT_MODEL:
		// 事件类任务调用EventTaskExecutor
		sjService.executeEventJob(jobInfo)
	}
}

// 执行event任务
func (sjService *jobService) executeEventJob(jobInfo *interfaces.JobInfo) {
	if jobInfo.EventTask != nil {
		task := *jobInfo.EventTask
		sjService.etService.EventTaskExecutor(context.Background(), task)
	} else {
		logger.Errorf("Executing event task : %s failed, because of event task is empty.", jobInfo.JobId)
	}

	// 当前任务执行完成后，触发依赖任务
	for _, depID := range jobInfo.EventTask.DownstreamDependentTask {
		depJob, exists := sjService.scheduler.jobs[depID]
		if !exists {
			logger.Infof("Dependency task %s not found", depID)
			continue
		}
		logger.Infof("Triggering dependency task: %s", depJob.JobId)
		sjService.executeJob(depJob)
	}
}
