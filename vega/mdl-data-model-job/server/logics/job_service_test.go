// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package logics

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	libmq "github.com/kweaver-ai/kweaver-go-lib/mq"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"data-model-job/common"
	cond "data-model-job/common/condition"
	"data-model-job/interfaces"
	dmock "data-model-job/interfaces/mock"
)

var (
	testCtx = context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage)
)

func MockNewJobService(jaMock interfaces.JobAccess, kaMock interfaces.KafkaAccess, dvsMock interfaces.DataViewService) *jobService {
	job := &jobService{
		appSetting: &common.AppSetting{
			MQSetting: libmq.MQSetting{
				Tenant: "default",
			},
		},
		jAccess:   jaMock,
		kAccess:   kaMock,
		dvService: dvsMock,
		jobMap:    sync.Map{},
		errChan:   make(chan jobError, 1),
		scheduler: NewScheduler(),
	}

	return job
}

func Test_JobService_StartJob(t *testing.T) {
	Convey("Test jobService StartJob", t, func() {
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		jaMock := dmock.NewMockJobAccess(mockCtl)
		kaMock := dmock.NewMockKafkaAccess(mockCtl)
		dvsMock := dmock.NewMockDataViewService(mockCtl)
		jsMock := MockNewJobService(jaMock, kaMock, dvsMock)

		jobInfo := &interfaces.JobInfo{
			JobId: "1a",
		}

		Convey("prepare job failed", func() {
			patches := ApplyPrivateMethod(jsMock, "prepareJob",
				func(*jobService, context.Context, *Job) error {
					return errors.New("error")
				})
			defer patches.Reset()

			jaMock.EXPECT().UpdateJobStatus(gomock.Any()).Return(nil)

			err := jsMock.StartJob(testCtx, jobInfo)
			So(err, ShouldNotBeNil)
		})

		Convey("update job status failed", func() {
			patches := ApplyPrivateMethod(jsMock, "prepareJob",
				func(*jobService, context.Context, *Job) error {
					return errors.New("error")
				})
			defer patches.Reset()

			jaMock.EXPECT().UpdateJobStatus(gomock.Any()).Return(errors.New("error"))

			err := jsMock.StartJob(testCtx, jobInfo)
			So(err, ShouldNotBeNil)
		})

		Convey("success", func() {
			patches1 := ApplyPrivateMethod(jsMock, "prepareJob",
				func(*jobService, context.Context, *Job) error {
					return nil
				})
			defer patches1.Reset()

			err := jsMock.StartJob(testCtx, jobInfo)
			So(err, ShouldBeNil)
		})

		Convey("success with schedule", func() {
			jobInfo := &interfaces.JobInfo{
				JobId:   "1a",
				JobType: interfaces.JOB_TYPE_SCHEDULE,
				Schedule: interfaces.Schedule{
					Type:       interfaces.SCHEDULE_TYPE_CRON,
					Expression: "0 * * * * ?",
				},
			}

			err := jsMock.StartJob(testCtx, jobInfo)
			So(err, ShouldBeNil)
		})
	})
}

func Test_JobService_UpdateJob(t *testing.T) {
	Convey("Test jobService UpdateJob", t, func() {
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		jaMock := dmock.NewMockJobAccess(mockCtl)
		kaMock := dmock.NewMockKafkaAccess(mockCtl)
		dvsMock := dmock.NewMockDataViewService(mockCtl)
		jsMock := MockNewJobService(jaMock, kaMock, dvsMock)

		jobId := "1a"
		jobInfo := &interfaces.JobInfo{
			JobId:     jobId,
			JobStatus: interfaces.JobStatus_Running,
		}
		job := &Job{
			JobInfo: jobInfo,
		}

		Convey("job not in memory", func() {
			err := jsMock.UpdateJob(testCtx, jobInfo)
			So(err, ShouldNotBeNil)
		})

		Convey("prepare job failed", func() {
			jsMock.jobMap.Store(jobId, job)

			patches := ApplyPrivateMethod(jsMock, "prepareJob", func(context.Context, *Job) error {
				return errors.New("error")
			})
			defer patches.Reset()

			jaMock.EXPECT().UpdateJobStatus(gomock.Any()).Return(errors.New("error"))

			err := jsMock.UpdateJob(testCtx, jobInfo)
			So(err, ShouldNotBeNil)
		})

		Convey("update job status failed", func() {
			jsMock.jobMap.Store(jobId, job)

			patches1 := ApplyPrivateMethod(jsMock, "prepareJob", func(context.Context, *Job) error {
				return errors.New("error")
			})
			defer patches1.Reset()

			patches2 := ApplyPrivateMethod(jsMock, "updateJobStatus", func(_ *Job, status string, _ string) error {
				if status == interfaces.JobStatus_Error {
					return nil
				}
				return errors.New("error")
			})
			defer patches2.Reset()

			err := jsMock.UpdateJob(testCtx, jobInfo)
			So(err, ShouldNotBeNil)
		})

		Convey("success", func() {
			jsMock.jobMap.Store(jobId, job)

			patches1 := ApplyPrivateMethod(jsMock, "prepareJob", func(context.Context, *Job) error {
				return nil
			})
			defer patches1.Reset()

			patches2 := ApplyPrivateMethod(jsMock, "updateJobStatus", func(*Job, string, string) error {
				return nil
			})
			defer patches2.Reset()

			err := jsMock.UpdateJob(testCtx, jobInfo)
			So(err, ShouldBeNil)
		})

		Convey("success with schedule job", func() {
			job := &interfaces.JobInfo{
				JobId:   jobId,
				JobType: interfaces.JOB_TYPE_SCHEDULE,
				Schedule: interfaces.Schedule{
					Type:       interfaces.SCHEDULE_TYPE_CRON,
					Expression: "0 * * * * ?",
				},
			}

			jsMock.scheduler.jobs[jobInfo.JobId] = job

			err := jsMock.UpdateJob(testCtx, job)
			So(err, ShouldBeNil)
		})
	})
}

func Test_JobService_StopJob(t *testing.T) {
	Convey("Test jobService StopJob", t, func() {
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		jaMock := dmock.NewMockJobAccess(mockCtl)
		kaMock := dmock.NewMockKafkaAccess(mockCtl)
		dvsMock := dmock.NewMockDataViewService(mockCtl)
		jsMock := MockNewJobService(jaMock, kaMock, dvsMock)

		jobId := "1a"
		jobInfo := &interfaces.JobInfo{
			JobId:     jobId,
			JobStatus: interfaces.JobStatus_Running,
		}
		job := &Job{
			JobInfo: jobInfo,
		}

		Convey("job not in memory", func() {
			err := jsMock.StopJob(testCtx, jobId)
			So(err, ShouldBeNil)
		})

		Convey("delete topic failed", func() {
			jsMock.jobMap.Store(jobId, job)

			kaMock.EXPECT().DeleteTopic(gomock.Any(), gomock.Any()).Return(errors.New("error"))

			err := jsMock.StopJob(testCtx, jobId)
			So(err, ShouldNotBeNil)
		})

		Convey("delete consumer group failed", func() {
			jsMock.jobMap.Store(jobId, job)

			kaMock.EXPECT().DeleteTopic(gomock.Any(), gomock.Any()).Return(nil)
			kaMock.EXPECT().DeleteConsumerGroups(gomock.Any()).Return(errors.New("error"))

			err := jsMock.StopJob(testCtx, jobId)
			So(err, ShouldNotBeNil)
		})

		Convey("success", func() {
			jsMock.jobMap.Store(jobId, job)
			kaMock.EXPECT().DeleteTopic(gomock.Any(), gomock.Any()).Return(nil)
			kaMock.EXPECT().DeleteConsumerGroups(gomock.Any()).Return(nil)

			err := jsMock.StopJob(testCtx, jobId)
			So(err, ShouldBeNil)
		})

		Convey("success with schedule job", func() {
			job := &interfaces.JobInfo{
				JobId:   jobId,
				JobType: interfaces.JOB_TYPE_SCHEDULE,
				Schedule: interfaces.Schedule{
					Type:       interfaces.SCHEDULE_TYPE_CRON,
					Expression: "0 * * * * ?",
				},
			}

			jsMock.scheduler.jobs[jobInfo.JobId] = job

			err := jsMock.StopJob(testCtx, jobId)
			So(err, ShouldBeNil)
		})
	})
}

// func Test_JobService_ListenToErrChan(t *testing.T) {
// 	Convey("Test jobService ListenToErrChan", t, func() {
// 		mockCtl := gomock.NewController(t)
// 		defer mockCtl.Finish()

// 		jaMock := dmock.NewMockJobAccess(mockCtl)
// 		kaMock := dmock.NewMockKafkaAccess(mockCtl)
// 		jsMock := MockNewJobService(jaMock, kaMock)

// 		jobId := uint64(1)
// 		jobInfo := &interfaces.JobInfo{
// 			JobId: jobId,
// 		}
// 		job := &Job{
// 			JobInfo: jobInfo,
// 			status:  interfaces.JobStatus_Running,
// 		}
// 		jsMock.jobMap.Store(jobId, job)
// 		jsMock.errChan <- jobError{
// 			jId:  jobId,
// 			jErr: errors.New("error"),
// 		}

// 		Convey("stop job failed", func() {
// 			patches1 := ApplyPrivateMethod(jsMock, "stopJob", func(context.Context, *Job) error {
// 				return errors.New("error")
// 			})
// 			defer patches1.Reset()

// 			jsMock.ListenToErrChan()
// 		})
// 	})
// }

func Test_JobService_prepareJob(t *testing.T) {
	Convey("Test jobService prepareJob", t, func() {
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		jaMock := dmock.NewMockJobAccess(mockCtl)
		kaMock := dmock.NewMockKafkaAccess(mockCtl)
		dvsMock := dmock.NewMockDataViewService(mockCtl)
		jsMock := MockNewJobService(jaMock, kaMock, dvsMock)

		jobId := "1a"
		jobInfo := &interfaces.JobInfo{
			JobId: jobId,
			DataView: interfaces.DataView{
				FieldScope: interfaces.ALL,
			},
		}
		job := &Job{
			JobInfo: jobInfo,
		}

		baseInfos := []interfaces.IndexBase{
			{Mappings: interfaces.Mappings{
				UserDefinedMappings: []interfaces.IndexBaseField{
					{Field: "x", Type: cond.DataType_Keyword},
					{Field: "y", Type: cond.DataType_Keyword},
				},
			}},
		}

		Convey("get indexbases failed", func() {
			dvsMock.EXPECT().GetIndexBases(gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))

			err := jsMock.prepareJob(testCtx, job)
			So(err, ShouldNotBeNil)
		})

		Convey("new condition failed", func() {
			dvsMock.EXPECT().GetIndexBases(gomock.Any(), gomock.Any()).Return(baseInfos, nil)

			patches2 := ApplyFuncReturn(cond.NewCondition, nil, errors.New("error"))
			defer patches2.Reset()

			err := jsMock.prepareJob(testCtx, job)
			So(err, ShouldNotBeNil)
		})

		Convey("get src topics failed", func() {
			dvsMock.EXPECT().GetIndexBases(gomock.Any(), gomock.Any()).Return(baseInfos, nil)

			patches2 := ApplyFuncReturn(cond.NewCondition, nil, nil)
			defer patches2.Reset()

			patches3 := ApplyPrivateMethod(jsMock, "getDataSourceTopics", func(map[string]any) ([]string, error) {
				return nil, errors.New("error")
			})
			defer patches3.Reset()

			err := jsMock.prepareJob(testCtx, job)
			So(err, ShouldNotBeNil)
		})

		Convey("generate sink topic info failed", func() {
			dvsMock.EXPECT().GetIndexBases(gomock.Any(), gomock.Any()).Return(baseInfos, nil)

			patches2 := ApplyFuncReturn(cond.NewCondition, nil, nil)
			defer patches2.Reset()

			patches3 := ApplyPrivateMethod(jsMock, "getDataSourceTopics", func(map[string]any) ([]string, error) {
				return []string{"test"}, nil
			})
			defer patches3.Reset()

			patches4 := ApplyPrivateMethod(jsMock, "generateSinkTopicInfo",
				func(context.Context, string, []string) (interfaces.TopicMetadata, error) {
					return interfaces.TopicMetadata{}, errors.New("error")
				})
			defer patches4.Reset()

			err := jsMock.prepareJob(testCtx, job)
			So(err, ShouldNotBeNil)
		})

		Convey("create topic failed", func() {
			dvsMock.EXPECT().GetIndexBases(gomock.Any(), gomock.Any()).Return(baseInfos, nil)

			patches2 := ApplyFuncReturn(cond.NewCondition, nil, nil)
			defer patches2.Reset()

			patches3 := ApplyPrivateMethod(jsMock, "getDataSourceTopics", func(map[string]any) ([]string, error) {
				return []string{"test"}, nil
			})
			defer patches3.Reset()

			patches4 := ApplyPrivateMethod(jsMock, "generateSinkTopicInfo",
				func(context.Context, string, []string) (interfaces.TopicMetadata, error) {
					return interfaces.TopicMetadata{}, nil
				})
			defer patches4.Reset()

			kaMock.EXPECT().CreateTopicOrPartition(gomock.Any(), gomock.Any()).Return(errors.New("error"))

			err := jsMock.prepareJob(testCtx, job)
			So(err, ShouldNotBeNil)
		})

		Convey("success", func() {
			dvsMock.EXPECT().GetIndexBases(gomock.Any(), gomock.Any()).Return(baseInfos, nil)

			patches2 := ApplyFuncReturn(cond.NewCondition, nil, nil)
			defer patches2.Reset()

			patches3 := ApplyPrivateMethod(jsMock, "getDataSourceTopics", func(map[string]any) ([]string, error) {
				return []string{"test"}, nil
			})
			defer patches3.Reset()

			patches4 := ApplyPrivateMethod(jsMock, "generateSinkTopicInfo",
				func(context.Context, string, []string) (interfaces.TopicMetadata, error) {
					return interfaces.TopicMetadata{TopicName: "default.mdl.view.345"}, nil
				})
			defer patches4.Reset()

			kaMock.EXPECT().CreateTopicOrPartition(gomock.Any(), gomock.Any()).Return(nil)

			err := jsMock.prepareJob(testCtx, job)
			So(err, ShouldBeNil)
		})
	})
}

func Test_JobService_startTasks(t *testing.T) {
	Convey("Test jobService startTasks", t, func() {
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		jobInfo := &interfaces.JobInfo{
			JobId:     "1a",
			JobStatus: interfaces.JobStatus_Running,
		}
		job := &Job{
			JobInfo:   jobInfo,
			srcTopics: []string{"test1", "test2"},
			sinkTopic: "sinktest",
			// status:    interfaces.JobStatus_Running,
		}

		appSetting := &common.AppSetting{}
		errChan := make(chan jobError, 100)

		Convey("task run failed", func() {
			task := &Task{}

			patches := ApplyMethodReturn(task, "Run", errors.New("error"))
			defer patches.Reset()

			job.startTasks(appSetting, errChan)
			time.Sleep(5 * time.Second)
		})

		Convey("task run success", func() {
			task := &Task{}

			patches := ApplyMethodReturn(task, "Run", nil)
			defer patches.Reset()

			job.startTasks(appSetting, errChan)
			time.Sleep(5 * time.Second)
		})
	})
}

func Test_JobService_stopTasks(t *testing.T) {
	Convey("Test jobService stopTasks", t, func() {
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		jobId := "1a"
		jobInfo := &interfaces.JobInfo{
			JobId:     jobId,
			JobStatus: interfaces.JobStatus_Running,
		}
		task1 := &Task{}
		task2 := &Task{}
		job := &Job{
			JobInfo:   jobInfo,
			srcTopics: []string{"test1", "test2"},
			sinkTopic: "sinktest",
			// status:    interfaces.JobStatus_Running,
			tasks: []*Task{task1, task2},
		}

		Convey("success", func() {
			task := &Task{}
			patches := ApplyMethodReturn(task, "Stop", nil)
			defer patches.Reset()

			job.stopTasks()
		})
	})
}

func Test_JobService_getDataSourceTopics(t *testing.T) {
	Convey("Test jobService getDataSourceTopics", t, func() {
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		jaMock := dmock.NewMockJobAccess(mockCtl)
		kaMock := dmock.NewMockKafkaAccess(mockCtl)
		dvsMock := dmock.NewMockDataViewService(mockCtl)
		jsMock := MockNewJobService(jaMock, kaMock, dvsMock)

		Convey("mapstructure decode failed", func() {
			dataSource := map[string]any{
				"type":       "index_base",
				"index_base": []any{"test1", "test2"},
			}
			topics, err := jsMock.getDataSourceTopics(dataSource)
			So(topics, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("base type is null", func() {
			dataSource := map[string]any{
				"type": "index_base",
				"index_base": []any{
					map[string]any{"base_type": ""},
					map[string]any{"base_type": "test2"},
				},
			}
			topics, err := jsMock.getDataSourceTopics(dataSource)
			So(topics, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("topics is empty", func() {
			dataSource := map[string]any{
				"type":       "index_base",
				"index_base": []any{},
			}
			topics, err := jsMock.getDataSourceTopics(dataSource)
			So(topics, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("success", func() {
			dataSource := map[string]any{
				"type": "index_base",
				"index_base": []any{
					map[string]any{"base_type": "test1"},
					map[string]any{"base_type": "test2"},
				},
			}
			topics, err := jsMock.getDataSourceTopics(dataSource)
			So(topics, ShouldResemble, []string{"default.mdl.process.test1", "default.mdl.process.test2"})
			So(err, ShouldBeNil)
		})
	})
}

func Test_JobService_generateSinkTopicInfo(t *testing.T) {
	Convey("Test jobService generateSinkTopicInfo", t, func() {
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		jaMock := dmock.NewMockJobAccess(mockCtl)
		kaMock := dmock.NewMockKafkaAccess(mockCtl)
		dvsMock := dmock.NewMockDataViewService(mockCtl)
		jsMock := MockNewJobService(jaMock, kaMock, dvsMock)

		viewId := "1"
		srcTopics := []string{"test1", "test2"}

		Convey("get topics info failed", func() {
			kaMock.EXPECT().DescribeTopics(gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))

			_, err := jsMock.generateSinkTopicInfo(testCtx, viewId, srcTopics)
			So(err, ShouldNotBeNil)
		})

		Convey("success", func() {
			mds := []interfaces.TopicMetadata{{TopicName: "test", PartitionsCount: 3}}
			kaMock.EXPECT().DescribeTopics(gomock.Any(), gomock.Any()).Return(mds, nil)

			topics, err := jsMock.generateSinkTopicInfo(testCtx, viewId, srcTopics)
			expected := interfaces.TopicMetadata{
				TopicName:       "default.mdl.view.1",
				PartitionsCount: 3,
			}
			So(topics, ShouldResemble, expected)
			So(err, ShouldBeNil)
		})
	})
}

func Test_JobService_updateJobStatus(t *testing.T) {
	Convey("Test jobService updateJobStatus", t, func() {
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		jaMock := dmock.NewMockJobAccess(mockCtl)
		kaMock := dmock.NewMockKafkaAccess(mockCtl)
		dvsMock := dmock.NewMockDataViewService(mockCtl)
		jsMock := MockNewJobService(jaMock, kaMock, dvsMock)

		jobId := "1a"
		jobInfo := &interfaces.JobInfo{
			JobId: jobId,
		}
		job := &Job{
			JobInfo: jobInfo,
		}

		status := interfaces.JobStatus_Error
		details := "error"

		Convey("update job status failed", func() {
			jaMock.EXPECT().UpdateJobStatus(gomock.Any()).Return(errors.New("error"))
			err := jsMock.updateJobStatus(job, status, details)
			So(err, ShouldNotBeNil)
		})

		Convey("success", func() {
			jaMock.EXPECT().UpdateJobStatus(gomock.Any()).Return(nil)
			err := jsMock.updateJobStatus(job, status, details)
			So(err, ShouldBeNil)
		})
	})
}

func Test_JobService_updateDbJobStatus(t *testing.T) {
	Convey("Test jobService updateDbJobStatus", t, func() {
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		jaMock := dmock.NewMockJobAccess(mockCtl)
		kaMock := dmock.NewMockKafkaAccess(mockCtl)
		dvsMock := dmock.NewMockDataViewService(mockCtl)
		jsMock := MockNewJobService(jaMock, kaMock, dvsMock)

		jobId := "1a"
		status := interfaces.JobStatus_Error
		details := "error"

		Convey("update job status in DB failed", func() {
			jaMock.EXPECT().UpdateJobStatus(gomock.Any()).Return(errors.New("error"))
			err := jsMock.updateDbJobStatus(jobId, status, details)
			So(err, ShouldNotBeNil)
		})

		Convey("success", func() {
			jaMock.EXPECT().UpdateJobStatus(gomock.Any()).Return(nil)
			err := jsMock.updateDbJobStatus(jobId, status, details)
			So(err, ShouldBeNil)
		})
	})
}

func Test_JobService_recoverJobs(t *testing.T) {
	Convey("Test jobService recoverJobs", t, func() {
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		jaMock := dmock.NewMockJobAccess(mockCtl)
		kaMock := dmock.NewMockKafkaAccess(mockCtl)
		dvsMock := dmock.NewMockDataViewService(mockCtl)
		jsMock := MockNewJobService(jaMock, kaMock, dvsMock)

		jobsToCreate := []*interfaces.JobInfo{
			{
				JobId:     "1a",
				JobStatus: interfaces.JobStatus_Error,
			},
		}
		jobsToUpdate := []*interfaces.JobInfo{
			{
				JobId: "3a",
			},
		}
		jobsToDelete := []*Job{
			{
				JobInfo: &interfaces.JobInfo{JobId: "3a"},
			},
		}

		Convey("syncMemoryJobBasedOnDB failed", func() {
			patches := ApplyPrivateMethod(jsMock, "syncMemoryJobBasedOnDB",
				func() ([]*interfaces.JobInfo, []*interfaces.JobInfo, []*Job, error) {
					return nil, nil, nil, errors.New("error")
				})
			defer patches.Reset()

			jsMock.recoverJobs()
		})

		Convey("update job status failed", func() {
			patches1 := ApplyPrivateMethod(jsMock, "syncMemoryJobBasedOnDB",
				func() ([]*interfaces.JobInfo, []*interfaces.JobInfo, []*Job, error) {
					return jobsToCreate, jobsToUpdate, jobsToDelete, nil
				})
			defer patches1.Reset()

			patches2 := ApplyPrivateMethod(jsMock, "updateDbJobStatus",
				func(string, string, string) error {
					return errors.New("error")
				})
			defer patches2.Reset()

			jsMock.recoverJobs()
		})

		Convey("create job failed", func() {
			patches1 := ApplyPrivateMethod(jsMock, "syncMemoryJobBasedOnDB",
				func() ([]*interfaces.JobInfo, []*interfaces.JobInfo, []*Job, error) {
					return jobsToCreate, jobsToUpdate, jobsToDelete, nil
				})
			defer patches1.Reset()

			patches2 := ApplyPrivateMethod(jsMock, "updateDbJobStatus",
				func(string, string, string) error {
					return nil
				})
			defer patches2.Reset()

			patches3 := ApplyPrivateMethod(jsMock, "StartJob",
				func(context.Context, *interfaces.JobInfo) error {
					return errors.New("error")
				})
			defer patches3.Reset()

			jsMock.recoverJobs()
		})

		Convey("update job failed", func() {
			patches1 := ApplyPrivateMethod(jsMock, "syncMemoryJobBasedOnDB",
				func() ([]*interfaces.JobInfo, []*interfaces.JobInfo, []*Job, error) {
					return jobsToCreate, jobsToUpdate, jobsToDelete, nil
				})
			defer patches1.Reset()

			patches2 := ApplyPrivateMethod(jsMock, "updateDbJobStatus",
				func(string, string, string) error {
					return nil
				})
			defer patches2.Reset()

			patches3 := ApplyPrivateMethod(jsMock, "StartJob",
				func(context.Context, *interfaces.JobInfo) error {
					return nil
				})
			defer patches3.Reset()

			patches4 := ApplyPrivateMethod(jsMock, "UpdateJob",
				func(context.Context, *interfaces.JobInfo) error {
					return errors.New("error")
				})
			defer patches4.Reset()

			jsMock.recoverJobs()
		})

		Convey("delete job failed", func() {
			patches1 := ApplyPrivateMethod(jsMock, "syncMemoryJobBasedOnDB",
				func() ([]*interfaces.JobInfo, []*interfaces.JobInfo, []*Job, error) {
					return jobsToCreate, jobsToUpdate, jobsToDelete, nil
				})
			defer patches1.Reset()

			patches2 := ApplyPrivateMethod(jsMock, "updateDbJobStatus",
				func(string, string, string) error {
					return nil
				})
			defer patches2.Reset()

			patches3 := ApplyPrivateMethod(jsMock, "StartJob",
				func(context.Context, *interfaces.JobInfo) error {
					return nil
				})
			defer patches3.Reset()

			patches4 := ApplyPrivateMethod(jsMock, "UpdateJob",
				func(context.Context, *interfaces.JobInfo) error {
					return errors.New("error")
				})
			defer patches4.Reset()

			patches5 := ApplyPrivateMethod(jsMock, "StopJob",
				func(context.Context, *interfaces.JobInfo) error {
					return errors.New("error")
				})
			defer patches5.Reset()

			jsMock.recoverJobs()
		})

		Convey("success", func() {
			patches1 := ApplyPrivateMethod(jsMock, "syncMemoryJobBasedOnDB",
				func() ([]*interfaces.JobInfo, []*interfaces.JobInfo, []*Job, error) {
					return jobsToCreate, jobsToUpdate, jobsToDelete, nil
				})
			defer patches1.Reset()

			patches2 := ApplyPrivateMethod(jsMock, "updateDbJobStatus",
				func(string, string, string) error {
					return nil
				})
			defer patches2.Reset()

			patches3 := ApplyPrivateMethod(jsMock, "StartJob",
				func(context.Context, *interfaces.JobInfo) error {
					return nil
				})
			defer patches3.Reset()

			patches4 := ApplyPrivateMethod(jsMock, "UpdateJob",
				func(context.Context, *interfaces.JobInfo) error {
					return errors.New("error")
				})
			defer patches4.Reset()

			patches5 := ApplyPrivateMethod(jsMock, "StopJob",
				func(context.Context, string) error {
					return nil
				})
			defer patches5.Reset()

			jsMock.recoverJobs()
		})
	})
}

func Test_JobService_syncMemoryJobBasedOnDB(t *testing.T) {
	Convey("Test jobService syncMemoryJobBasedOnDB", t, func() {
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		jaMock := dmock.NewMockJobAccess(mockCtl)
		kaMock := dmock.NewMockKafkaAccess(mockCtl)
		dvsMock := dmock.NewMockDataViewService(mockCtl)
		jsMock := MockNewJobService(jaMock, kaMock, dvsMock)

		jobId := "1a"
		jobInfo := &interfaces.JobInfo{
			JobId:     jobId,
			JobStatus: interfaces.JobStatus_Running,
			DataView: interfaces.DataView{
				DataSource: map[string]any{
					"type": "index_base",
					"index_base": []any{
						map[string]any{
							"base_type": "base1",
						},
					},
				},
			},
		}
		job := &Job{
			JobInfo:   jobInfo,
			srcTopics: []string{"test1", "test2"},
			sinkTopic: "sinktest",
			// status:    interfaces.JobStatus_Running,
		}

		Convey("list view jobs failed", func() {
			jaMock.EXPECT().ListViewJobs().Return(nil, errors.New("error"))
			_, _, _, err := jsMock.syncMemoryJobBasedOnDB()
			So(err, ShouldNotBeNil)
		})

		Convey("job in memory but not in DB", func() {
			jsMock.jobMap.Store(jobId, job)

			jobsInfo := []interfaces.JobInfo{}
			jaMock.EXPECT().ListViewJobs().Return(jobsInfo, nil)
			jobsToCreate, jobsToUpdate, jobsToDelete, err := jsMock.syncMemoryJobBasedOnDB()

			So(len(jobsToCreate), ShouldEqual, 0)
			So(len(jobsToUpdate), ShouldEqual, 0)
			So(len(jobsToDelete), ShouldEqual, 1)
			So(err, ShouldBeNil)
		})

		Convey("job in both memory and db, job failed", func() {
			// job.status = interfaces.JobStatus_Error
			job.JobStatus = interfaces.JobStatus_Error
			jsMock.jobMap.Store(jobId, job)

			jobsInfo := []interfaces.JobInfo{
				{JobId: jobId},
			}
			jaMock.EXPECT().ListViewJobs().Return(jobsInfo, nil)
			jobsToCreate, jobsToUpdate, jobsToDelete, err := jsMock.syncMemoryJobBasedOnDB()

			So(len(jobsToCreate), ShouldEqual, 0)
			So(len(jobsToUpdate), ShouldEqual, 1)
			So(len(jobsToDelete), ShouldEqual, 0)
			So(err, ShouldBeNil)
		})

		Convey("job in both memory and db, job config updated", func() {
			// job.status = interfaces.JobStatus_Running
			job.JobStatus = interfaces.JobStatus_Running
			jsMock.jobMap.Store(jobId, job)

			jobsInfo := []interfaces.JobInfo{
				{
					JobId: jobId,
					DataView: interfaces.DataView{
						DataSource: map[string]any{
							"type": "index_base",
							"index_base": []any{
								map[string]any{
									"base_type": "base2",
								},
							},
						},
					},
				},
			}
			jaMock.EXPECT().ListViewJobs().Return(jobsInfo, nil)
			jobsToCreate, jobsToUpdate, jobsToDelete, err := jsMock.syncMemoryJobBasedOnDB()

			So(len(jobsToCreate), ShouldEqual, 0)
			So(len(jobsToUpdate), ShouldEqual, 1)
			So(len(jobsToDelete), ShouldEqual, 0)
			So(err, ShouldBeNil)
		})

		Convey("job in DB but not in memory", func() {
			jobsInfo := []interfaces.JobInfo{
				{JobId: jobId},
			}
			jaMock.EXPECT().ListViewJobs().Return(jobsInfo, nil)
			jobsToCreate, jobsToUpdate, jobsToDelete, err := jsMock.syncMemoryJobBasedOnDB()

			So(len(jobsToCreate), ShouldEqual, 1)
			So(len(jobsToUpdate), ShouldEqual, 0)
			So(len(jobsToDelete), ShouldEqual, 0)
			So(err, ShouldBeNil)
		})
	})
}

func Test_JobService_WatchJobsTopic(t *testing.T) {
	Convey("Test jobService watchJobsTopic", t, func() {
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		jaMock := dmock.NewMockJobAccess(mockCtl)
		kaMock := dmock.NewMockKafkaAccess(mockCtl)
		dvsMock := dmock.NewMockDataViewService(mockCtl)
		jsMock := MockNewJobService(jaMock, kaMock, dvsMock)

		jobId := "1a"
		jobInfo := &interfaces.JobInfo{
			JobId:     jobId,
			JobStatus: interfaces.JobStatus_Running,
			DataView: interfaces.DataView{
				DataSource: map[string]any{
					"type": "index_base",
					"index_base": []any{
						map[string]any{
							"base_type": "base1",
						},
					},
				},
			},
		}

		job := &Job{
			JobInfo:   jobInfo,
			srcTopics: []string{"test1", "test2"},
			sinkTopic: "sinktest",
		}

		jsMock.jobMap.Store(jobId, job)

		Convey("generateSinkTopicInfo failed", func() {
			patches := ApplyPrivateMethod(jsMock, "generateSinkTopicInfo",
				func(ctx context.Context, viewId string, srcTopics []string) (interfaces.TopicMetadata, error) {
					return interfaces.TopicMetadata{}, errors.New("error")
				})
			defer patches.Reset()

			jsMock.watchJobsTopic()
		})

		Convey("CreateTopicOrPartition failed", func() {
			patches := ApplyPrivateMethod(jsMock, "generateSinkTopicInfo",
				func(ctx context.Context, viewId string, srcTopics []string) (interfaces.TopicMetadata, error) {
					return interfaces.TopicMetadata{}, nil
				})
			defer patches.Reset()

			kaMock.EXPECT().CreateTopicOrPartition(gomock.Any(), gomock.Any()).Return(errors.New("error"))

			jsMock.watchJobsTopic()
		})

	})
}

func TestDeepEqualJobCondition(t *testing.T) {
	Convey("Test deepEqualJobCondition", t, func() {
		Convey("a is nil && b is nil", func() {
			res := deepEqualJobCondition(nil, nil)
			So(res, ShouldBeTrue)
		})

		Convey("a is nil and b is not nil", func() {
			b := &cond.CondCfg{}
			res := deepEqualJobCondition(nil, b)
			So(res, ShouldBeFalse)
		})

		Convey("a is not nil and b is nil", func() {
			a := &cond.CondCfg{
				Name:      "xx",
				Operation: cond.OperationExist,
			}
			res := deepEqualJobCondition(a, nil)
			So(res, ShouldBeFalse)
		})

		Convey("name is not equal", func() {
			a := &cond.CondCfg{
				Name:      "x",
				Operation: cond.OperationExist,
			}

			b := &cond.CondCfg{
				Name:      "y",
				Operation: cond.OperationExist,
			}
			res := deepEqualJobCondition(a, b)
			So(res, ShouldBeFalse)
		})

		Convey("operation is not equal", func() {
			a := &cond.CondCfg{
				Name:      "x",
				Operation: cond.OperationExist,
			}

			b := &cond.CondCfg{
				Name:      "x",
				Operation: cond.OperarionNotIn,
			}
			res := deepEqualJobCondition(a, b)
			So(res, ShouldBeFalse)
		})

		Convey("value from is not equal", func() {
			a := &cond.CondCfg{
				Name:      "x",
				Operation: cond.OperationEq,
				ValueOptCfg: cond.ValueOptCfg{
					ValueFrom: "const",
				},
			}

			b := &cond.CondCfg{
				Name:      "x",
				Operation: cond.OperationEq,
				ValueOptCfg: cond.ValueOptCfg{
					ValueFrom: "field",
				},
			}
			res := deepEqualJobCondition(a, b)
			So(res, ShouldBeFalse)
		})

		Convey("value is not equal", func() {
			a := &cond.CondCfg{
				Name:      "x",
				Operation: cond.OperationEq,
				ValueOptCfg: cond.ValueOptCfg{
					ValueFrom: "const",
					Value:     "melody",
				},
			}

			b := &cond.CondCfg{
				Name:      "x",
				Operation: cond.OperationEq,
				ValueOptCfg: cond.ValueOptCfg{
					ValueFrom: "const",
					Value:     "annie",
				},
			}
			res := deepEqualJobCondition(a, b)
			So(res, ShouldBeFalse)
		})

		Convey("sub conditions length is not equal", func() {
			a := &cond.CondCfg{
				Operation: cond.OperationAnd,
				SubConds:  []*cond.CondCfg{},
			}

			b := &cond.CondCfg{
				Operation: cond.OperationAnd,
				SubConds: []*cond.CondCfg{
					{Name: "x", Operation: cond.OperationExist},
				},
			}
			res := deepEqualJobCondition(a, b)
			So(res, ShouldBeFalse)
		})

		Convey("sub conditions are not equal", func() {
			a := &cond.CondCfg{
				Operation: cond.OperationAnd,
				SubConds: []*cond.CondCfg{
					{Name: "x", Operation: cond.OperationExist},
					{Name: "y", Operation: cond.OperationExist},
					{Name: "z", Operation: cond.OperationExist},
				},
			}

			b := &cond.CondCfg{
				Operation: cond.OperationAnd,
				SubConds: []*cond.CondCfg{
					{Name: "x", Operation: cond.OperationExist},
					{Name: "y", Operation: cond.OperationExist},
					{Name: "z", Operation: cond.OperationNotExist},
				},
			}
			res := deepEqualJobCondition(a, b)
			So(res, ShouldBeFalse)
		})

		Convey("condition struct is not equal", func() {
			a := &cond.CondCfg{
				Name:      "x",
				Operation: cond.OperationEq,
				ValueOptCfg: cond.ValueOptCfg{
					ValueFrom: "const",
					Value:     "melody",
				},
			}

			b := &cond.CondCfg{
				Operation: cond.OperationAnd,
				SubConds: []*cond.CondCfg{
					{Name: "x", Operation: cond.OperationExist},
				},
			}
			res := deepEqualJobCondition(a, b)
			So(res, ShouldBeFalse)
		})
	})
}

func TestRecoverMetricJobs(t *testing.T) {
	Convey("Test jobService recoverMetricJobs", t, func() {
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		jaMock := dmock.NewMockJobAccess(mockCtl)
		kaMock := dmock.NewMockKafkaAccess(mockCtl)
		dvsMock := dmock.NewMockDataViewService(mockCtl)
		jsMock := MockNewJobService(jaMock, kaMock, dvsMock)

		Convey("Failed becuase of ListMetricJobs failed", func() {
			jaMock.EXPECT().ListMetricJobs().Return(nil, errors.New("error"))
			jsMock.recoverMetricJobs()
		})

		Convey("success with metrics", func() {
			createJob := interfaces.JobInfo{
				JobId:      "createJob",
				JobType:    interfaces.JOB_TYPE_SCHEDULE,
				ModuleType: interfaces.MODULE_TYPE_METRIC_MODEL,
				Schedule: interfaces.Schedule{
					Type:       interfaces.SCHEDULE_TYPE_CRON,
					Expression: "0 * * * * ?",
				},
				MetricTask: &task,
			}
			updateJob := interfaces.JobInfo{
				JobId:      "updateJob",
				JobType:    interfaces.JOB_TYPE_SCHEDULE,
				ModuleType: interfaces.MODULE_TYPE_METRIC_MODEL,
				Schedule: interfaces.Schedule{
					Type:       interfaces.SCHEDULE_TYPE_CRON,
					Expression: "0 * * * * ?",
				},
				MetricTask: &task,
			}
			deleteJob := interfaces.JobInfo{
				JobId:      "deleteJob",
				JobType:    interfaces.JOB_TYPE_SCHEDULE,
				ModuleType: interfaces.MODULE_TYPE_METRIC_MODEL,
				Schedule: interfaces.Schedule{
					Type:       interfaces.SCHEDULE_TYPE_CRON,
					Expression: "0 * * * * ?",
				},
				MetricTask: &task,
			}
			jsMock.scheduler.jobs[updateJob.JobId] = &updateJob
			jsMock.scheduler.jobs[deleteJob.JobId] = &deleteJob

			jaMock.EXPECT().ListMetricJobs().Return([]interfaces.JobInfo{createJob, updateJob}, nil)
			jaMock.EXPECT().ListObjectiveJobs().Return(nil, nil)
			jsMock.recoverMetricJobs()
		})

		Convey("success with metrics and objectives", func() {
			createMJob := interfaces.JobInfo{
				JobId:      "createMJob",
				JobType:    interfaces.JOB_TYPE_SCHEDULE,
				ModuleType: interfaces.MODULE_TYPE_METRIC_MODEL,
				Schedule: interfaces.Schedule{
					Type:       interfaces.SCHEDULE_TYPE_CRON,
					Expression: "0 * * * * ?",
				},
				MetricTask: &task,
			}
			updateMJob := interfaces.JobInfo{
				JobId:      "updateMJob",
				JobType:    interfaces.JOB_TYPE_SCHEDULE,
				ModuleType: interfaces.MODULE_TYPE_METRIC_MODEL,
				Schedule: interfaces.Schedule{
					Type:       interfaces.SCHEDULE_TYPE_FIXED,
					Expression: "5m",
				},
				MetricTask: &task,
			}
			deleteMJob := interfaces.JobInfo{
				JobId:      "deleteMJob",
				JobType:    interfaces.JOB_TYPE_SCHEDULE,
				ModuleType: interfaces.MODULE_TYPE_METRIC_MODEL,
				Schedule: interfaces.Schedule{
					Type:       interfaces.SCHEDULE_TYPE_FIXED,
					Expression: "5m",
				},
				MetricTask: &task,
			}
			jsMock.scheduler.jobs[updateMJob.JobId] = &updateMJob
			jsMock.scheduler.jobs[deleteMJob.JobId] = &deleteMJob

			createOJob := interfaces.JobInfo{
				JobId:      "createOJob",
				JobType:    interfaces.JOB_TYPE_SCHEDULE,
				ModuleType: interfaces.MODULE_TYPE_OBJECTIVE_MODEL,
				Schedule: interfaces.Schedule{
					Type:       interfaces.SCHEDULE_TYPE_FIXED,
					Expression: "5m",
				},
				MetricTask: &task,
			}
			updateOJob := interfaces.JobInfo{
				JobId:      "updateOJob",
				JobType:    interfaces.JOB_TYPE_SCHEDULE,
				ModuleType: interfaces.MODULE_TYPE_OBJECTIVE_MODEL,
				Schedule: interfaces.Schedule{
					Type:       interfaces.SCHEDULE_TYPE_FIXED,
					Expression: "5m",
				},
				MetricTask: &task,
			}
			deleteOJob := interfaces.JobInfo{
				JobId:      "deleteOJob",
				JobType:    interfaces.JOB_TYPE_SCHEDULE,
				ModuleType: interfaces.MODULE_TYPE_OBJECTIVE_MODEL,
				Schedule: interfaces.Schedule{
					Type:       interfaces.SCHEDULE_TYPE_FIXED,
					Expression: "5m",
				},
				MetricTask: &task,
			}
			jsMock.scheduler.jobs[updateOJob.JobId] = &updateOJob
			jsMock.scheduler.jobs[deleteOJob.JobId] = &deleteOJob

			jaMock.EXPECT().ListMetricJobs().Return([]interfaces.JobInfo{createMJob, updateMJob}, nil)
			jaMock.EXPECT().ListObjectiveJobs().Return([]interfaces.JobInfo{createOJob, updateOJob}, nil)
			jsMock.recoverMetricJobs()
		})
	})
}

func TestRecoverEventJobs(t *testing.T) {
	Convey("Test jobService recoverEventJobs", t, func() {
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		jaMock := dmock.NewMockJobAccess(mockCtl)
		kaMock := dmock.NewMockKafkaAccess(mockCtl)
		dvsMock := dmock.NewMockDataViewService(mockCtl)
		jsMock := MockNewJobService(jaMock, kaMock, dvsMock)

		Convey("Failed becuase of ListEventJobs failed", func() {
			jaMock.EXPECT().ListEventJobs().Return(nil, errors.New("error"))
			jsMock.recoverEventJobs()
		})

		Convey("success with metrics", func() {
			createJob := interfaces.JobInfo{
				JobId:      "createJob",
				JobType:    interfaces.JOB_TYPE_SCHEDULE,
				ModuleType: interfaces.MODULE_TYPE_EVENT_MODEL,
				Schedule: interfaces.Schedule{
					Type:       interfaces.SCHEDULE_TYPE_CRON,
					Expression: "0 * * * * ?",
				},
				EventTask: &eventTask,
			}
			updateJob := interfaces.JobInfo{
				JobId:      "updateJob",
				JobType:    interfaces.JOB_TYPE_SCHEDULE,
				ModuleType: interfaces.MODULE_TYPE_EVENT_MODEL,
				Schedule: interfaces.Schedule{
					Type:       interfaces.SCHEDULE_TYPE_CRON,
					Expression: "0 * * * * ?",
				},
				EventTask: &eventTask,
			}
			deleteJob := interfaces.JobInfo{
				JobId:      "deleteJob",
				JobType:    interfaces.JOB_TYPE_SCHEDULE,
				ModuleType: interfaces.MODULE_TYPE_EVENT_MODEL,
				Schedule: interfaces.Schedule{
					Type:       interfaces.SCHEDULE_TYPE_CRON,
					Expression: "0 * * * * ?",
				},
				EventTask: &eventTask,
			}
			jsMock.scheduler.jobs[updateJob.JobId] = &updateJob
			jsMock.scheduler.jobs[deleteJob.JobId] = &deleteJob

			jaMock.EXPECT().ListEventJobs().Return([]interfaces.JobInfo{createJob, updateJob}, nil)
			jsMock.recoverEventJobs()
		})
	})
}
