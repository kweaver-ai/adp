// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package logics

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/mitchellh/mapstructure"
	"github.com/panjf2000/ants/v2"

	"data-model-job/common"
	cond "data-model-job/common/condition"
	"data-model-job/interfaces"
)

var (
	jServiceOnce sync.Once
	jService     interfaces.JobService
)

var (
	FlushBytes    int = 5 * 1024 * 1024
	FlushItems    int = 10000
	FlushInterval     = 5 * time.Second

	RetryInterval        = 3000 * time.Millisecond
	FailureThreshold     = 5
	PackagePoolSize  int = 500
	PackagePool, _       = ants.NewPool(PackagePoolSize, ants.WithPreAlloc(true), ants.WithNonblocking(false))
)

type jobError struct {
	jId  string
	jErr error
}

type jobService struct {
	appSetting *common.AppSetting
	dvService  interfaces.DataViewService
	etService  interfaces.EventTaskService
	jAccess    interfaces.JobAccess
	kAccess    interfaces.KafkaAccess
	mtService  interfaces.MetricTaskService
	jobMap     sync.Map
	scheduler  *Scheduler
	// job发生错误后将err写入errChan
	errChan chan jobError
}

func NewJobService(appSetting *common.AppSetting) interfaces.JobService {
	jServiceOnce.Do(func() {
		FlushBytes = appSetting.ServerSetting.FlushMiB * 1024 * 1024
		FlushItems = appSetting.ServerSetting.FlushItems
		FlushInterval = time.Duration(appSetting.ServerSetting.FlushIntervalSec) * time.Second
		RetryInterval = time.Duration(appSetting.ServerSetting.RetryIntervalMs) * time.Millisecond
		FailureThreshold = appSetting.ServerSetting.FailureThreshold
		PackagePoolSize = appSetting.ServerSetting.PackagePoolSize
		PackagePool, _ = ants.NewPool(PackagePoolSize, ants.WithPreAlloc(true), ants.WithNonblocking(false))

		jService = &jobService{
			appSetting: appSetting,
			dvService:  NewDataViewService(appSetting),
			mtService:  NewMetricTaskService(appSetting),
			etService:  NewEventTaskService(appSetting),
			jAccess:    JAccess,
			kAccess:    KAccess,
			errChan:    make(chan jobError, 100),
			scheduler:  NewScheduler(),
		}

		// 任务自动恢复协程
		go jService.WatchJob(appSetting.ServerSetting.WatchJobsIntervalMin * time.Minute)

		// 开启一个协程监听job的errChan，如果监听到错误，则停止运行此job，并更新状态为失败状态
		go jService.ListenToErrChan()
	})

	return jService
}

type Job struct {
	*interfaces.JobInfo
	srcTopics []string
	sinkTopic string
	viewCond  cond.Condition
	tasks     []*Task
}

// job 信息打印
func (j *Job) String() string {
	return fmt.Sprintf("{job_id = %s, job_type = %s, job_status = %s, job_status_details = %s, job_view_id = %s, "+
		"source_topics = %v, sink_topic = %s}",
		j.JobId, j.JobType, j.JobStatus, j.JobStatusDetails, j.ViewId, j.srcTopics, j.sinkTopic)
}

// 新建内存中的job对象
func newJob(jobInfo *interfaces.JobInfo) *Job {
	dmJob := &Job{
		JobInfo: jobInfo,
		tasks:   make([]*Task, 0),
	}

	dmJob.JobStatus = interfaces.JobStatus_Running

	return dmJob
}

// 启动job
func (jService *jobService) StartJob(ctx context.Context, jobInfo *interfaces.JobInfo) error {
	// 以下是创建的是视图的job，视图的job有很多字段是指标、目标不用的，所以要根据 type 来进行不同的操作，构建不同的job类型
	// 指标模型提交的jobInfo中包含了持久化task表中的内容和信息，所以直接把JobConfig反序列化为MetricTask.
	if jobInfo.JobType == interfaces.JOB_TYPE_SCHEDULE {
		// 指标类任务,按调度策略，生成不同的调度
		return jService.addScheduleJob(jobInfo)
	}

	if _, ok := jService.jobMap.Load(jobInfo.JobId); ok {
		logger.Errorf("Job %s already exists in memory!", jobInfo.JobId)
		return fmt.Errorf("job already exists in memory")
	}

	// 创建job对象
	dmJob := newJob(jobInfo)

	// 将job添加到内存的全局jobMap里
	jService.addJob(dmJob.JobId, dmJob)

	err := jService.prepareJob(ctx, dmJob)
	if err != nil {
		logger.Errorf("Prepare job failed, %v", err)
		updateJobStatusErr := jService.updateJobStatus(dmJob, interfaces.JobStatus_Error, err.Error())
		if updateJobStatusErr != nil {
			return updateJobStatusErr
		}

		return err
	}

	// 启动tasks
	dmJob.startTasks(jService.appSetting, jService.errChan)

	return nil
}

// 更新job
func (jService *jobService) UpdateJob(ctx context.Context, jobInfo *interfaces.JobInfo) error {
	// 指标模型提交的jobInfo中包含了持久化task表中的内容和信息，所以直接把JobConfig反序列化为MetricTask.
	if jService.scheduler != nil {
		task, exists := jService.scheduler.jobs[jobInfo.JobId]
		if exists {
			if task.JobType == interfaces.JOB_TYPE_SCHEDULE {
				return jService.updateScheduleJob(*jobInfo)
			}
		}
	}

	job, ok := jService.jobMap.Load(jobInfo.JobId)
	if !ok {
		logger.Errorf("Update a job %d that does not exist in memory", jobInfo.JobId)
		return fmt.Errorf("update a job that does not exist in memory")
	}

	dmJob := job.(*Job)

	// 停止运行tasks
	dmJob.stopTasks()

	// 创建job对象
	dmJob = newJob(jobInfo)

	// 更新内存里的job对象
	jService.updateJob(dmJob.JobId, dmJob)

	err := jService.prepareJob(ctx, dmJob)
	if err != nil {
		logger.Errorf("Prepare job failed, %v", err)
		updateJobStatusErr := jService.updateJobStatus(dmJob, interfaces.JobStatus_Error, err.Error())
		if updateJobStatusErr != nil {
			return updateJobStatusErr
		}

		return err
	}

	// 如果更新之前任务为失败状态，需要更新数据库的任务状态为运行中
	err = jService.updateJobStatus(dmJob, interfaces.JobStatus_Running, "")
	if err != nil {
		logger.Errorf("Failed to update job %d status, %s", dmJob.JobId, err.Error())
		return err
	}

	// 启动 tasks
	dmJob.startTasks(jService.appSetting, jService.errChan)

	return nil
}

// 停止job
func (jService *jobService) StopJob(ctx context.Context, jobId string) error {
	// 指标模型提交的jobInfo中包含了持久化task表中的内容和信息，所以直接把JobConfig反序列化为MetricTask.
	if jService.scheduler != nil {
		task, exists := jService.scheduler.jobs[jobId]
		if exists {
			if task.JobType == interfaces.JOB_TYPE_SCHEDULE {
				return jService.deleteScheduleJob(*task)
			}
		}
	}

	// 以下是创建的是视图的job，视图的job有很多字段是指标、目标不用的，所以要根据 type 来进行不同的操作，构建不同的job类型
	job, ok := jService.jobMap.Load(jobId)
	if !ok {
		errDetails := fmt.Sprintf("Stop a job %s that does not exist in memory.", jobId)
		logger.Warn(errDetails)
		return nil
	}

	dmJob := job.(*Job)

	// 停止tasks
	dmJob.stopTasks()

	// 删掉目标topic
	err := jService.kAccess.DeleteTopic(ctx, []string{dmJob.sinkTopic})
	if err != nil {
		return err
	}

	// 删掉消费组
	groupId := ComsumerGroupID(jService.appSetting.MQSetting.Tenant, dmJob.ViewId)
	err = jService.kAccess.DeleteConsumerGroups([]string{groupId})
	if err != nil {
		return err
	}

	// 从内存中移除job
	jService.removeJob(jobId)

	logger.Debugf("Stopped job %d", dmJob.JobId)
	return nil
}

// 服务重启时自动恢复正常和失败的任务
// 每隔10分钟自动恢复失败的任务
func (jService *jobService) WatchJob(interval time.Duration) {
	logger.Infof("Auto recover jobs, interval: %s", interval)

	for {

		// 恢复或同步视图的任务
		jService.recoverJobs()
		jService.watchJobsTopic()

		// 恢复或同步指标类（指标、目标）的任务
		jService.recoverMetricJobs()

		// 恢复或同步事件的定时任务
		jService.recoverEventJobs()

		time.Sleep(interval)
	}
}

// 运行任务的准备工作
func (jService *jobService) prepareJob(ctx context.Context, dmJob *Job) error {
	// userId 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, interfaces.AccountInfo{
		ID:   dmJob.Creator.ID,
		Type: dmJob.Creator.Type,
	})

	// 如果字段范围为全部，去索引库拿全部字段
	if dmJob.FieldScope == interfaces.ALL {
		baseInfos, err := jService.dvService.GetIndexBases(ctx, &dmJob.DataView)
		if err != nil {
			return err
		}

		viewFields := []*cond.Field{}

		for _, base := range baseInfos {
			allBaseFields := mergeIndexBaseFields(base.Mappings)

			for _, field := range allBaseFields {
				viewFields = append(viewFields, &cond.Field{
					Name: field.Field,
					Type: field.Type,
				})
			}
		}

		dmJob.Fields = viewFields
	}

	// 构造 fieldsMap
	fieldsMap := make(map[string]*cond.Field)
	for _, fi := range dmJob.Fields {
		field := fi
		fieldsMap[fi.Name] = field
	}
	// 将fieldsMap传给dmJob
	dmJob.FieldsMap = fieldsMap

	// 创建condition接口
	viewCond, err := cond.NewCondition(ctx, dmJob.Condition, dmJob.FieldsMap)
	if err != nil {
		logger.Errorf("Job %s new condition failed: %v", dmJob, err)
		return err
	}
	// 将viewCond接口对象传给dmJob
	dmJob.viewCond = viewCond

	// 获取待消费的topics列表
	srcTopics, err := jService.getDataSourceTopics(dmJob.DataSource)
	if err != nil {
		return fmt.Errorf("get source topics failed, %v", err)
	}

	// 将源topic传给dmJob
	dmJob.srcTopics = srcTopics

	// 生成目标topic信息
	topicMetadata, err := jService.generateSinkTopicInfo(ctx, dmJob.ViewId, srcTopics)
	if err != nil {
		return fmt.Errorf("generate sink topic info failed, %v", err)
	}

	// 将目标topic名称传给dmJob
	dmJob.sinkTopic = topicMetadata.TopicName

	// 创建目标topic
	err = jService.kAccess.CreateTopicOrPartition(ctx, topicMetadata)
	if err != nil {
		return fmt.Errorf("create sink topic info failed, %v", err)
	}

	logger.Debugf("prepare job %s finished", dmJob)
	return nil
}

// 监听job的errChan
func (jService *jobService) ListenToErrChan() {
	for jobErr := range jService.errChan {
		logger.Debugf("Received error of job %d from channel", jobErr.jId)
		go func(jobErr jobError) {
			val, ok := jService.jobMap.Load(jobErr.jId)
			if ok {
				if dmJob, ok := val.(*Job); ok {
					// 停止job，任务失败的时候不删除topic和消费者组
					dmJob.stopTasks()

					// 更新任务状态为失败
					if err := jService.updateJobStatus(dmJob, interfaces.JobStatus_Error, jobErr.jErr.Error()); err != nil {
						logger.Errorf("Update job status to 'error' failed, %v", err)
						return
					}
				}
			}
		}(jobErr)
	}

	// 正常情况下errChan不应该被关闭
	logger.Errorf("Error channel closed, exiting loop")
}

// 将job添加到jobsMap
func (jService *jobService) addJob(id string, job *Job) {
	jService.jobMap.Store(id, job)
}

// 更新jobMap的值
func (jService *jobService) updateJob(id string, job *Job) {
	jService.jobMap.Store(id, job)
}

// 将job从jobsMap里移除
func (jService *jobService) removeJob(jobId string) {
	jService.jobMap.Delete(jobId)
}

// 启动tasks
func (dmJob *Job) startTasks(appSetting *common.AppSetting, errChan chan jobError) {
	var mu sync.Mutex
	// 一个task消费 一个topic
	for _, topic := range dmJob.srcTopics {
		go func(topic string) {
			task := NewTask(appSetting, &dmJob.DataView, dmJob.JobId, topic, dmJob.sinkTopic, dmJob.viewCond)
			mu.Lock()
			dmJob.tasks = append(dmJob.tasks, task)
			mu.Unlock()

			defer func() {
				task.RunningChannel <- false
			}()

			logger.Infof("Start run task %s", task)
			err := task.Run()
			if err != nil {
				errChan <- jobError{jId: dmJob.JobId, jErr: err}
				logger.Info("An error has been sent to the channel")
			}
		}(topic)
	}
}

// 停止tasks
func (dmJob *Job) stopTasks() {
	var wg sync.WaitGroup
	for _, task := range dmJob.tasks {
		wg.Add(1)
		go func(task *Task) {
			defer wg.Done()
			err := task.Stop()
			if err != nil {
				return
			}
		}(task)
	}
	wg.Wait()

	// 删除所有 tasks,长度为0, 容量保持不变
	dmJob.tasks = dmJob.tasks[:0]
}

// 一个视图一个消费组
func ComsumerGroupID(tenant string, viewId string) string {
	return fmt.Sprintf("%s.mdl.view.%s", tenant, viewId)
}

// 根据索引库确定对应的topics
func (jService *jobService) getDataSourceTopics(dataSource map[string]any) ([]string, error) {
	var bases []interfaces.SimpleIndexBase
	err := mapstructure.Decode(dataSource[interfaces.Index_Base], &bases)
	if err != nil {
		return nil, fmt.Errorf("mapstructure decode dataSource failed, %v", err)
	}

	topics := make([]string, 0, len(bases))

	for _, base := range bases {
		// 索引库类型不能为空
		if base.BaseType == "" {
			return nil, fmt.Errorf("the index base type is null")
		}

		topic := fmt.Sprintf(interfaces.Process_Topic, jService.appSetting.MQSetting.Tenant, base.BaseType)
		topics = append(topics, topic)
	}

	if len(topics) == 0 {
		errDetails := "no topics to subscribe, check the data source config of the data view is empty"
		logger.Error(errDetails)
		return nil, errors.New(errDetails)
	}

	return topics, nil
}

// 生成目标 topic 信息
func (jService *jobService) generateSinkTopicInfo(ctx context.Context, viewId string, srcTopics []string) (interfaces.TopicMetadata, error) {
	// 获取来源topics的分区信息
	topicsMetadata, err := jService.kAccess.DescribeTopics(ctx, srcTopics)
	if err != nil {
		return interfaces.TopicMetadata{}, err
	}

	if len(topicsMetadata) == 0 {
		logger.Errorf("View %s topics %v don't exist", viewId, srcTopics)
		return interfaces.TopicMetadata{}, fmt.Errorf("view %s data source's topics %v don't exist", viewId, srcTopics)
	}

	// 目标topic的分区数量是来源topics的分区最大值
	partitionCount := 0
	for _, metadata := range topicsMetadata {
		partitionCount = max(partitionCount, metadata.PartitionsCount)
	}

	topicName := fmt.Sprintf(interfaces.Sink_Topic, jService.appSetting.MQSetting.Tenant, viewId)

	return interfaces.TopicMetadata{
		TopicName:       topicName,
		PartitionsCount: partitionCount,
	}, nil
}

// 先更新数据库，再更新内存
func (jService *jobService) updateJobStatus(dmJob *Job, status, details string) error {
	jobInfo := interfaces.JobInfo{
		JobId:            dmJob.JobId,
		JobStatus:        status,
		JobStatusDetails: details,
	}
	err := jService.jAccess.UpdateJobStatus(jobInfo)
	if err != nil {
		logger.Errorf("Update job %d's status failed", dmJob.JobId)

		dmJob.JobStatus = interfaces.JobStatus_Error
		return fmt.Errorf("failed to update the job status and job details in DB, %s", err.Error())
	}

	dmJob.JobStatus = status
	return nil
}

// 更新数据库中的job状态
func (jService *jobService) updateDbJobStatus(jobId string, status, details string) error {
	jobInfo := interfaces.JobInfo{
		JobId:            jobId,
		JobStatus:        status,
		JobStatusDetails: details,
	}
	err := jService.jAccess.UpdateJobStatus(jobInfo)
	if err != nil {
		logger.Errorf("Update job %d's status failed, %v", jobId, err)
		return err
	}

	return nil
}

// 恢复任务
func (jService *jobService) recoverJobs() {
	ctx := context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage)
	jobsToCreate, jobsToUpdate, jobsToDelete, err := jService.syncMemoryJobBasedOnDB()
	if err != nil {
		return
	}

	logger.Debugf("Recover: %d jobs need to create", len(jobsToCreate))
	logger.Debugf("Recover: %d jobs need to update", len(jobsToUpdate))
	logger.Debugf("Recover: %d jobs need to delete", len(jobsToDelete))

	logger.Debugf("Recover: jobsToCreate list is %v", jobsToCreate)
	logger.Debugf("Recover: jobsToUpdate list is %v", jobsToUpdate)
	logger.Debugf("Recover: jobsToDelete list is %v", jobsToDelete)

	for _, jobInfo := range jobsToCreate {
		logger.Debugf("Recover: create job %d", jobInfo.JobId)

		// 更新数据库job状态为运行中
		if jobInfo.JobStatus == interfaces.JobStatus_Error {
			err = jService.updateDbJobStatus(jobInfo.JobId, interfaces.JobStatus_Running, "")
			if err != nil {
				logger.Errorf("Recover: update job %d status failed, %s", jobInfo.JobId, err.Error())
				// 如果此次更新数据库不成功，等待下次自动恢复流程恢复这个任务
				continue
			}
		}

		if err := jService.StartJob(ctx, jobInfo); err != nil {
			// 失败了继续循环的下一个
			logger.Errorf("Recover: create job %d failed, %s", jobInfo.JobId, err.Error())

		}

	}

	// 处理待更新的job列表
	for _, jobInfo := range jobsToUpdate {
		logger.Debugf("Recover: update job %d", jobInfo.JobId)

		if err := jService.UpdateJob(ctx, jobInfo); err != nil {
			logger.Errorf("Recover: update job %d failed, %s", jobInfo.JobId, err.Error())
		}
	}

	// 处理待删除的job列表
	for _, dmJob := range jobsToDelete {
		logger.Debugf("Recover: delete job %d", dmJob.JobId)

		if err := jService.StopJob(ctx, dmJob.JobId); err != nil {
			logger.Errorf("Recover: delete job %d failed", dmJob.JobId, err.Error())
		}
	}
}

// 同步内存和数据库的Job信息，以数据库为准
// 1. job内存中有, 数据库中没有, 将job加入待删除的列表
// 2. job在内存和数据库都有, 将失败的job加入待创建的列表，判断运行状态的job任务配置是否更新，如果有更新则将job加入待更新的列表
// 3. job在数据库有，内存中没有, 将失败的和运行的job都加入待创建的列表
func (jService *jobService) syncMemoryJobBasedOnDB() (jobsToCreate, jobsToUpdate []*interfaces.JobInfo,
	jobsToDelete []*Job, err error) {
	// 查询数据库中所有的数据视图的job
	jobs, err := jService.jAccess.ListViewJobs()
	if err != nil {
		logger.Errorf("Recover: list jobs failed, %v", err)
		return nil, nil, nil, err
	}
	logger.Debugf("Recover: there are %d data view jobs in db", len(jobs))

	jobsMap := make(map[string]interfaces.JobInfo)
	for _, v := range jobs {
		val := v
		jobsMap[v.JobId] = val
	}

	jobsToCreate = make([]*interfaces.JobInfo, 0)
	jobsToUpdate = make([]*interfaces.JobInfo, 0)
	jobsToDelete = make([]*Job, 0)

	jService.jobMap.Range(func(jobId, job interface{}) bool {
		dmJobId := jobId.(string)
		dmJob := job.(*Job)
		logger.Debugf("Recover: job %d in memory, jobStatus %s", dmJob.JobId, dmJob.JobStatus)

		// 1. job在内存里, 但是没在数据库中, 添加job到jobsToDelete列表里
		if jobInfo, ok := jobsMap[dmJobId]; !ok {
			jobsToDelete = append(jobsToDelete, dmJob)
		} else { // 2. job在内存和数据库里都有, 将失败的job添加到jobsToUpdate列表里
			if dmJob.JobStatus == interfaces.JobStatus_Error {
				logger.Infof("Job %d is in both memory and DB, add failed job to the jobsToUpdate list", dmJob.JobId)
				jobsToUpdate = append(jobsToUpdate, &jobInfo)
			} else {
				// 比较job配置是否更新
				dbViewJobConfig := &interfaces.ViewJobCfg{
					DataSource: jobInfo.DataSource,
					FieldScope: jobInfo.FieldScope,
					Fields:     jobInfo.Fields,
					Condition:  jobInfo.Condition,
					JobStatus:  jobInfo.JobStatus, // 内存和数据库job状态可能不一致
				}

				memoryViewJobConfig := &interfaces.ViewJobCfg{
					DataSource: dmJob.DataSource,
					FieldScope: dmJob.FieldScope,
					Fields:     dmJob.Fields,
					Condition:  dmJob.Condition,
					JobStatus:  dmJob.JobStatus,
				}

				equal, err := compareJobConfig(dbViewJobConfig, memoryViewJobConfig)
				if err != nil {
					logger.Errorf("compare job config failed, %v", err)
					return true
				}

				// job配置更新了, 将job添加到jobsToUpdate列表里
				if !equal {
					logger.Infof("Job %d is in both memory and DB, add job with configuration changes to the jobsToUpdate list", dmJob.JobId)
					logger.Infof("Job config in DB is %s", dbViewJobConfig)
					logger.Infof("Job config in memory is %s", memoryViewJobConfig)

					jobsToUpdate = append(jobsToUpdate, &jobInfo)
				}
			}

		}

		return true
	})

	for _, jobInfo := range jobsMap {
		jobInfo := jobInfo
		// 3. job在数据库里, 但是没在内存里, 将失败的和允许中的job添加到jobsToCreate列表里
		if _, ok := jService.jobMap.Load(jobInfo.JobId); !ok {
			jobsToCreate = append(jobsToCreate, &jobInfo)

		}
	}

	return jobsToCreate, jobsToUpdate, jobsToDelete, nil
}

// 轮询 topic 的状态
// 每隔10分钟，检查运行中的任务源端topic是否新增分区，如果新增，目标topic的分区也相应增加
func (jService *jobService) watchJobsTopic() {
	ctx := context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage)
	jService.jobMap.Range(func(jobID, job interface{}) bool {
		dmJob := job.(*Job)

		// 生成目标topic信息
		topicMetadata, err := jService.generateSinkTopicInfo(ctx, dmJob.ViewId, dmJob.srcTopics)
		if err != nil {
			logger.Errorf("Watch job topic: job %s generate sink topic info failed, %v", dmJob, err)
			return true
		}

		// 创建目标topic或者topic分区增加了创建分区
		err = jService.kAccess.CreateTopicOrPartition(ctx, topicMetadata)
		if err != nil {
			logger.Errorf("Watch job topic: job %s create sink topic or partition failed, %v", dmJob, err)
			return true
		}

		return true
	})
}

// 比较两个任务的配置是否相等
func compareJobConfig(job1, job2 *interfaces.ViewJobCfg) (bool, error) {
	// 如果内存和数据库job状态不一致，返回false，然后触发任务重启
	if job1.JobStatus != job2.JobStatus {
		logger.Info("Compare job config, job status in DB and memory are inconsistent")
		return false, nil
	}

	var bases1, bases2 []interfaces.SimpleIndexBase
	err := mapstructure.Decode(job1.DataSource[interfaces.INDEX_BASE], &bases1)
	if err != nil {
		logger.Errorf("Compare job config, mapstructure decode job's dataSource failed, %v", err)
		return false, err
	}

	baseTypes1 := make(map[string]struct{})
	for _, base := range bases1 {
		baseTypes1[base.BaseType] = struct{}{}
	}

	err = mapstructure.Decode(job2.DataSource[interfaces.INDEX_BASE], &bases2)
	if err != nil {
		logger.Errorf("Compare job config, mapstructure decode job's dataSource failed, %v", err)
		return false, err
	}

	baseTypes2 := make(map[string]struct{})
	for _, base := range bases2 {
		baseTypes2[base.BaseType] = struct{}{}
	}

	if !reflect.DeepEqual(baseTypes1, baseTypes2) {
		logger.Info("Compare job config, job's base_types in DB and memory are inconsistent")
		return false, nil
	}

	if job1.FieldScope != job2.FieldScope {
		logger.Info("Compare job config, job's fieldScope in DB and memory are inconsistent")
		return false, nil
	}

	// 如果是全部字段，字段存储为 []，无需检查fields
	// 如果是部分字段，只需检查字段的名字和类型是否发生变化，备注变化和其他内部字段变化也无需检查
	if job1.FieldScope == interfaces.CUSTOM {
		job1FieldsMap := make(map[string]string)
		for i := range job1.Fields {
			job1FieldsMap[job1.Fields[i].Name] = job1.Fields[i].Type
		}

		job2FieldsMap := make(map[string]string)
		for j := range job2.Fields {
			job2FieldsMap[job2.Fields[j].Name] = job2.Fields[j].Type
		}

		if !reflect.DeepEqual(job1FieldsMap, job2FieldsMap) {
			logger.Info("Compare job config, job's fields in DB and memory are inconsistent")
			return false, nil
		}
	}

	if !deepEqualJobCondition(job1.Condition, job2.Condition) {
		logger.Info("Compare job config, job's condition in DB and memory are inconsistent")
		return false, nil
	}

	return true, nil
}

func deepEqualJobCondition(a, b *cond.CondCfg) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		logger.Info("Compare job config, one is nil, and one is not nil")
		return false
	}

	if a.Name != b.Name {
		logger.Info("Compare job config, the two condition field names are different")
		return false
	}

	if a.Operation != b.Operation {
		logger.Info("Compare job config, the two operations are different")
		return false
	}

	if a.ValueFrom != b.ValueFrom {
		logger.Info("Compare job config, the two value_from are different")
		return false
	}

	if !reflect.DeepEqual(a.Value, b.Value) {
		logger.Info("Compare job config, the two values are different")
		return false
	}

	if len(a.SubConds) != len(b.SubConds) {
		logger.Info("Compare job config, the two sub_conditions length are different")
		return false
	}

	for i := range a.SubConds {
		if !deepEqualJobCondition(a.SubConds[i], b.SubConds[i]) {
			logger.Info("Compare job config, the two sub_conditions are different")
			return false
		}
	}

	return true
}

// 恢复metric任务
func (jService *jobService) recoverMetricJobs() {
	ctx := context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage)

	// 恢复指标、目标的任务
	jobsToCreate, jobsToUpdate, jobsToDelete, err := jService.syncMemoryMetricJobBasedOnDB()
	if err != nil {
		logger.Errorf("Metric Recover: syncMemoryMetricJobBasedOnDB failed, %s", err.Error())
		return
	}

	logger.Debugf("Metric Recover: %d jobs need to create", len(jobsToCreate))
	logger.Debugf("Metric Recover: %d jobs need to update", len(jobsToUpdate))
	logger.Debugf("Metric Recover: %d jobs need to delete", len(jobsToDelete))

	logger.Debugf("Metric Recover: jobsToCreate list is %v", jobsToCreate)
	logger.Debugf("Metric Recover: jobsToUpdate list is %v", jobsToUpdate)
	logger.Debugf("Metric Recover: jobsToDelete list is %v", jobsToDelete)

	for _, jobInfo := range jobsToCreate {
		logger.Debugf("Metric Recover: create job %d", jobInfo.JobId)

		if err := jService.StartJob(ctx, jobInfo); err != nil {
			// 失败了继续循环的下一个
			logger.Errorf("Metric Recover: create job %d failed, %s", jobInfo.JobId, err.Error())
		}

	}

	// 处理待更新的job列表
	for _, jobInfo := range jobsToUpdate {
		logger.Debugf("Metric Recover: update job %d", jobInfo.JobId)

		if err := jService.UpdateJob(ctx, jobInfo); err != nil {
			logger.Errorf("Metric Recover: update job %d failed, %s", jobInfo.JobId, err.Error())
		}
	}

	// 处理待删除的job列表
	for _, dmJob := range jobsToDelete {
		logger.Debugf("Metric Recover: delete job %d", dmJob.JobId)

		if err := jService.StopJob(ctx, dmJob.JobId); err != nil {
			logger.Errorf("Metric Recover: delete job %d failed", dmJob.JobId, err.Error())
		}
	}
}

// 同步内存和数据库的Job信息，以数据库为准
// 1. job内存中有, 数据库中没有, 将job加入待删除的列表
// 2. job在内存和数据库都有, 将job加入待更新的列表
// 3. job在数据库有，内存中没有, 将job加入待创建的列表
func (jService *jobService) syncMemoryMetricJobBasedOnDB() (jobsToCreate, jobsToUpdate, jobsToDelete []*interfaces.JobInfo, err error) {
	// 查询数据库中所有的指标任务
	jobs, err := jService.jAccess.ListMetricJobs()
	if err != nil {
		logger.Errorf("Recover: list jobs failed, %v", err)
		return nil, nil, nil, err
	}
	logger.Debugf("Recover: there are %d metric model jobs in db", len(jobs))

	// 查询数据库中所有的指标任务
	objectiveJobs, err := jService.jAccess.ListObjectiveJobs()
	if err != nil {
		logger.Errorf("Recover: list jobs failed, %v", err)
		return nil, nil, nil, err
	}
	logger.Debugf("Recover: there are %d objective model jobs in db", len(objectiveJobs))

	jobsMap := make(map[string]interfaces.JobInfo)
	for _, v := range jobs {
		val := v
		jobsMap[v.JobId] = val
	}

	for _, v := range objectiveJobs {
		val := v
		jobsMap[v.JobId] = val
	}

	jobsToCreate = make([]*interfaces.JobInfo, 0)
	jobsToUpdate = make([]*interfaces.JobInfo, 0)
	jobsToDelete = make([]*interfaces.JobInfo, 0)

	jService.scheduler.mu.Lock()
	defer jService.scheduler.mu.Unlock()

	for jobId, jobInfoInMem := range jService.scheduler.jobs {
		logger.Debugf("Recover: job %d in memory, jobStatus %s", jobInfoInMem.JobId, jobInfoInMem.JobStatus)

		jobInfoInDB, ok := jobsMap[jobId]
		if jobInfoInDB.ModuleType == interfaces.MODULE_TYPE_METRIC_MODEL ||
			jobInfoInDB.ModuleType == interfaces.MODULE_TYPE_OBJECTIVE_MODEL {
			if !ok {
				// 1. job在内存里, 但是没在数据库中, 添加job到jobsToDelete列表里
				jobsToDelete = append(jobsToDelete, jobInfoInMem)
			} else {
				// 2. job在内存和数据库里都有, 将失败的job添加到jobsToUpdate列表里
				// 以数据库为准，将job加入待更新的列表
				// 比较内存和数据库中的对象是否相同，如果相同就不更新
				equal, err := compareMetricJobConfig(jobInfoInMem.MetricTask, jobInfoInDB.MetricTask)
				if err != nil {
					logger.Errorf("compare metric job config failed, %v", err)
					return nil, nil, nil, err
				}

				// job配置更新了, 将job添加到jobsToUpdate列表里
				if !equal {
					logger.Infof("Metric job %d is in both memory and DB, add job with configuration changes to the jobsToUpdate list", jobInfoInDB.JobId)
					logger.Infof("Metric job config in DB is %s", jobInfoInDB.MetricTask)
					logger.Infof("Metric job config in memory is %s", jobInfoInMem.MetricTask)

					jobsToUpdate = append(jobsToUpdate, &jobInfoInDB)
				}
			}
		}
	}

	for _, jobInfo := range jobsMap {
		jobInfo := jobInfo
		// 3. job在数据库里, 但是没在内存里, 将失败的和允许中的job添加到jobsToCreate列表里
		if _, ok := jService.scheduler.jobs[jobInfo.JobId]; !ok {
			jobsToCreate = append(jobsToCreate, &jobInfo)
		}
	}

	return jobsToCreate, jobsToUpdate, jobsToDelete, nil
}

// 恢复metric任务
func (jService *jobService) recoverEventJobs() {
	ctx := context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage)

	// 恢复指标、目标的任务
	jobsToCreate, jobsToUpdate, jobsToDelete, err := jService.syncMemoryEventJobBasedOnDB()
	if err != nil {
		return
	}

	logger.Debugf("Event Recover: %d jobs need to create", len(jobsToCreate))
	logger.Debugf("Event Recover: %d jobs need to update", len(jobsToUpdate))
	logger.Debugf("Event Recover: %d jobs need to delete", len(jobsToDelete))

	logger.Debugf("Event Recover: jobsToCreate list is %v", jobsToCreate)
	logger.Debugf("Event Recover: jobsToUpdate list is %v", jobsToUpdate)
	logger.Debugf("Event Recover: jobsToDelete list is %v", jobsToDelete)

	for _, jobInfo := range jobsToCreate {
		logger.Debugf("Event Recover: create job %d", jobInfo.JobId)

		if err := jService.StartJob(ctx, jobInfo); err != nil {
			// 失败了继续循环的下一个
			logger.Errorf("Event Recover: create job %d failed, %s", jobInfo.JobId, err.Error())
		}

	}

	// 处理待更新的job列表
	for _, jobInfo := range jobsToUpdate {
		logger.Debugf("Event Recover: update job %d", jobInfo.JobId)

		if err := jService.UpdateJob(ctx, jobInfo); err != nil {
			logger.Errorf("Event Recover: update job %d failed, %s", jobInfo.JobId, err.Error())
		}
	}

	// 处理待删除的job列表
	for _, dmJob := range jobsToDelete {
		logger.Debugf("Event Recover: delete job %d", dmJob.JobId)

		if err := jService.StopJob(ctx, dmJob.JobId); err != nil {
			logger.Errorf("Event Recover: delete job %d failed", dmJob.JobId, err.Error())
		}
	}
}

// 同步内存和数据库的Job信息，以数据库为准
// 1. job内存中有, 数据库中没有, 将job加入待删除的列表
// 2. job在内存和数据库都有, 将job加入待更新的列表
// 3. job在数据库有，内存中没有, 将job加入待创建的列表
func (jService *jobService) syncMemoryEventJobBasedOnDB() (jobsToCreate, jobsToUpdate, jobsToDelete []*interfaces.JobInfo, err error) {
	// 查询数据库中所有的指标任务
	jobs, err := jService.jAccess.ListEventJobs()
	if err != nil {
		logger.Errorf("Recover: list event jobs failed, %v", err)
		return nil, nil, nil, err
	}
	logger.Debugf("Recover: there are %d event model jobs in db", len(jobs))

	jobsMap := make(map[string]interfaces.JobInfo)
	for _, v := range jobs {
		val := v
		jobsMap[v.JobId] = val
	}

	jobsToCreate = make([]*interfaces.JobInfo, 0)
	jobsToUpdate = make([]*interfaces.JobInfo, 0)
	jobsToDelete = make([]*interfaces.JobInfo, 0)

	jService.scheduler.mu.Lock()
	defer jService.scheduler.mu.Unlock()

	for jobId, jobInfoInMem := range jService.scheduler.jobs {
		logger.Debugf("Recover: Event job %d in memory, jobStatus %s", jobInfoInMem.JobId, jobInfoInMem.JobStatus)

		jobInfoInDB, ok := jobsMap[jobId]
		if jobInfoInDB.ModuleType == interfaces.MODULE_TYPE_EVENT_MODEL {
			if !ok {
				// 1. job在内存里, 但是没在数据库中, 添加job到jobsToDelete列表里
				jobsToDelete = append(jobsToDelete, jobInfoInMem)
			} else {
				// 2. job在内存和数据库里都有, 将失败的job添加到jobsToUpdate列表里
				// 以数据库为准，将job加入待更新的列表
				// 比较内存和数据库中的对象是否相同，如果相同就不更新
				equal, err := compareEventJobConfig(jobInfoInMem.EventTask, jobInfoInDB.EventTask)
				if err != nil {
					logger.Errorf("compare event job config failed, %v", err)
					return nil, nil, nil, err
				}

				// job配置更新了, 将job添加到jobsToUpdate列表里
				if !equal {
					logger.Infof("Event job %d is in both memory and DB, add job with configuration changes to the jobsToUpdate list", jobInfoInDB.JobId)
					logger.Infof("Event job config in DB is %s", jobInfoInDB.EventTask)
					logger.Infof("Event job config in memory is %s", jobInfoInMem.EventTask)

					jobsToUpdate = append(jobsToUpdate, &jobInfoInDB)
				}
			}
		}
	}

	for _, jobInfo := range jobsMap {
		jobInfo := jobInfo
		// 3. job在数据库里, 但是没在内存里, 将失败的和允许中的job添加到jobsToCreate列表里
		if _, ok := jService.scheduler.jobs[jobInfo.JobId]; !ok {
			jobsToCreate = append(jobsToCreate, &jobInfo)
		}
	}

	return jobsToCreate, jobsToUpdate, jobsToDelete, nil
}

// 比较两个指标类的任务的配置是否相等
func compareMetricJobConfig(job1, job2 *interfaces.MetricTask) (bool, error) {
	// 比较 taskName
	if job1.TaskName != job2.TaskName {
		logger.Info("Compare job config, metric task name in DB and memory are inconsistent")
		return false, nil
	}
	// 比较 module_type
	if job1.ModuleType != job2.ModuleType {
		logger.Info("Compare job config, metric task module_type in DB and memory are inconsistent")
		return false, nil
	}
	// 比较 model_id
	if job1.ModelID != job2.ModelID {
		logger.Info("Compare job config, metric task model_id in DB and memory are inconsistent")
		return false, nil
	}
	// 比较 schedule
	if job1.Schedule.Type != job2.Schedule.Type {
		logger.Info("Compare job config, metric task schedule type in DB and memory are inconsistent")
		return false, nil
	}
	if job1.Schedule.Expression != job2.Schedule.Expression {
		logger.Info("Compare job config, metric task schedule expression in DB and memory are inconsistent")
		return false, nil
	}
	// 比较 time_windows
	if !reflect.DeepEqual(job1.TimeWindows, job1.TimeWindows) {
		logger.Info("Compare job config, metric task time windows in DB and memory are inconsistent")
		return false, nil
	}
	// 比较 steps
	if !reflect.DeepEqual(job1.Steps, job1.Steps) {
		logger.Info("Compare job config, metric task steps in DB and memory are inconsistent")
		return false, nil
	}
	// 比较 index_base
	if job1.IndexBase != job2.IndexBase {
		logger.Info("Compare job config, metric task index base in DB and memory are inconsistent")
		return false, nil
	}
	// plan_times是根据api拿最新的，不用比较
	// retrace_duration是创建的时候给的，不允许修改，所以也不用比较

	return true, nil
}

// 比较两个指标类的任务的配置是否相等
func compareEventJobConfig(job1, job2 *interfaces.EventTask) (bool, error) {
	// 比较 model_id
	if job1.ModelID != job2.ModelID {
		logger.Info("Compare job config, event task model_id in DB and memory are inconsistent")
		return false, nil
	}
	// 比较 storage_config
	if job1.StorageConfig.DataViewId != job2.StorageConfig.DataViewId {
		logger.Info("Compare job config, event task storage_config's data_view_id in DB and memory are inconsistent")
		return false, nil
	}
	if job1.StorageConfig.DataViewName != job2.StorageConfig.DataViewName {
		logger.Info("Compare job config, event task storage_config's data_view_name in DB and memory are inconsistent")
		return false, nil
	}
	if job1.StorageConfig.IndexBase != job2.StorageConfig.IndexBase {
		logger.Info("Compare job config, event task storage_config's index_base in DB and memory are inconsistent")
		return false, nil
	}
	// 比较 schedule
	if job1.Schedule.Type != job2.Schedule.Type {
		logger.Info("Compare job config, event task schedule type in DB and memory are inconsistent")
		return false, nil
	}
	if job1.Schedule.Expression != job2.Schedule.Expression {
		logger.Info("Compare job config, event task schedule expression in DB and memory are inconsistent")
		return false, nil
	}
	// 比较 dispatch_config
	if job1.DispatchConfig.TimeOut != job2.DispatchConfig.TimeOut {
		logger.Info("Compare job config, event task dispatch_config's timeout in DB and memory are inconsistent")
		return false, nil
	}
	if job1.DispatchConfig.RouteStrategy != job2.DispatchConfig.RouteStrategy {
		logger.Info("Compare job config, event task dispatch_config's route_strategy in DB and memory are inconsistent")
		return false, nil
	}
	if job1.DispatchConfig.BlockStrategy != job2.DispatchConfig.BlockStrategy {
		logger.Info("Compare job config, event task dispatch_config's block_strategy in DB and memory are inconsistent")
		return false, nil
	}
	if job1.DispatchConfig.FailRetryCount != job2.DispatchConfig.FailRetryCount {
		logger.Info("Compare job config, event task dispatch_config's fail_retry_count in DB and memory are inconsistent")
		return false, nil
	}
	// 比较 execute_params
	if !reflect.DeepEqual(job1.ExecuteParameter, job1.ExecuteParameter) {
		logger.Info("Compare job config, event task execute_parameter in DB and memory are inconsistent")
		return false, nil
	}
	// 比较 downstream_dependent_task
	if !reflect.DeepEqual(job1.DownstreamDependentTask, job1.DownstreamDependentTask) {
		logger.Info("Compare job config, event task downstream_dependent_task in DB and memory are inconsistent")
		return false, nil
	}

	return true, nil
}
