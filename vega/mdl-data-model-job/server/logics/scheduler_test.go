// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package logics

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"data-model-job/interfaces"
	dmock "data-model-job/interfaces/mock"
)

func Test_JobService_AddScheduleJob(t *testing.T) {
	Convey("Test jobService addScheduleJob", t, func() {
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		jaMock := dmock.NewMockJobAccess(mockCtl)
		kaMock := dmock.NewMockKafkaAccess(mockCtl)
		dvsMock := dmock.NewMockDataViewService(mockCtl)
		jsMock := MockNewJobService(jaMock, kaMock, dvsMock)

		Convey("job exists", func() {
			job := &interfaces.JobInfo{
				JobId:   "job1",
				JobType: interfaces.JOB_TYPE_SCHEDULE,
				Schedule: interfaces.Schedule{
					Type:       interfaces.SCHEDULE_TYPE_CRON,
					Expression: "0 * * * * ?",
				},
			}

			jsMock.scheduler.jobs["job1"] = job

			err := jsMock.addScheduleJob(job)
			So(err.Error(), ShouldEqual, "job with ID job1 already exists")
		})

		Convey("ParseDuration failed", func() {
			job := &interfaces.JobInfo{
				JobId:   "job1",
				JobType: interfaces.JOB_TYPE_SCHEDULE,
				Schedule: interfaces.Schedule{
					Type:       interfaces.SCHEDULE_TYPE_FIXED,
					Expression: "5a",
				},
			}

			err := jsMock.addScheduleJob(job)
			So(err.Error(), ShouldEqual, "not a valid duration string: \"5a\"")
		})

		Convey("unknown schedule type ", func() {
			job := &interfaces.JobInfo{
				JobId:   "job1",
				JobType: interfaces.JOB_TYPE_SCHEDULE,
				Schedule: interfaces.Schedule{
					Type:       "a",
					Expression: "5a",
				},
			}

			err := jsMock.addScheduleJob(job)
			So(err.Error(), ShouldEqual, "unknown schedule type: a")
		})
	})
}

func Test_JobService_DeleteScheduleJob(t *testing.T) {
	Convey("Test jobService deleteScheduleJob", t, func() {
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		jaMock := dmock.NewMockJobAccess(mockCtl)
		kaMock := dmock.NewMockKafkaAccess(mockCtl)
		dvsMock := dmock.NewMockDataViewService(mockCtl)
		jsMock := MockNewJobService(jaMock, kaMock, dvsMock)

		Convey("unknown schedule type ", func() {
			job := &interfaces.JobInfo{
				JobId:   "job1",
				JobType: interfaces.JOB_TYPE_SCHEDULE,
				Schedule: interfaces.Schedule{
					Type:       "a",
					Expression: "5a",
				},
			}
			jsMock.scheduler.jobs[job.JobId] = job

			err := jsMock.deleteScheduleJob(*job)
			So(err.Error(), ShouldEqual, "unknown schedule type: a")
		})

		Convey("delete fixed job ", func() {
			job := &interfaces.JobInfo{
				JobId:   "job1",
				JobType: interfaces.JOB_TYPE_SCHEDULE,
				Schedule: interfaces.Schedule{
					Type:       interfaces.SCHEDULE_TYPE_FIXED,
					Expression: "5m",
				},
				Ticker:   time.NewTicker(time.Duration(time.Minute * 10)),
				StopChan: make(chan struct{}),
			}
			jsMock.scheduler.jobs[job.JobId] = job

			err := jsMock.deleteScheduleJob(*job)
			So(err, ShouldBeNil)
		})

		Convey("delete cron job ", func() {
			job := &interfaces.JobInfo{
				JobId:   "job1",
				JobType: interfaces.JOB_TYPE_SCHEDULE,
				Schedule: interfaces.Schedule{
					Type:       interfaces.SCHEDULE_TYPE_CRON,
					Expression: "0 * * * * ?",
				},
			}
			jsMock.scheduler.jobs[job.JobId] = job

			err := jsMock.deleteScheduleJob(*job)
			So(err, ShouldBeNil)
		})
	})
}

func Test_JobService_UpdateScheduleJob(t *testing.T) {
	Convey("Test jobService updateScheduleJob", t, func() {
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		jaMock := dmock.NewMockJobAccess(mockCtl)
		kaMock := dmock.NewMockKafkaAccess(mockCtl)
		dvsMock := dmock.NewMockDataViewService(mockCtl)
		jsMock := MockNewJobService(jaMock, kaMock, dvsMock)

		Convey("job not found ", func() {
			job := &interfaces.JobInfo{
				JobId:   "job1111",
				JobType: interfaces.JOB_TYPE_SCHEDULE,
				Schedule: interfaces.Schedule{
					Type:       "a",
					Expression: "5a",
				},
			}

			err := jsMock.updateScheduleJob(*job)
			So(err.Error(), ShouldEqual, "job with ID job1111 not found")
		})

		Convey("unknown schedule type ", func() {
			job := &interfaces.JobInfo{
				JobId:   "job1",
				JobType: interfaces.JOB_TYPE_SCHEDULE,
				Schedule: interfaces.Schedule{
					Type:       "a",
					Expression: "5a",
				},
			}
			jsMock.scheduler.jobs[job.JobId] = job

			err := jsMock.updateScheduleJob(*job)
			So(err.Error(), ShouldEqual, "unknown schedule type: a")
		})

		Convey("update fixed job ", func() {
			job := &interfaces.JobInfo{
				JobId:   "job1",
				JobType: interfaces.JOB_TYPE_SCHEDULE,
				Schedule: interfaces.Schedule{
					Type:       interfaces.SCHEDULE_TYPE_FIXED,
					Expression: "5m",
				},
				Ticker:   time.NewTicker(time.Duration(time.Minute * 10)),
				StopChan: make(chan struct{}),
			}
			jsMock.scheduler.jobs[job.JobId] = job

			err := jsMock.updateScheduleJob(*job)
			So(err, ShouldBeNil)
		})

		Convey("update cron job ", func() {
			job := &interfaces.JobInfo{
				JobId:   "job1",
				JobType: interfaces.JOB_TYPE_SCHEDULE,
				Schedule: interfaces.Schedule{
					Type:       interfaces.SCHEDULE_TYPE_CRON,
					Expression: "0 * * * * ?",
				},
			}
			jsMock.scheduler.jobs[job.JobId] = job

			err := jsMock.updateScheduleJob(*job)
			So(err, ShouldBeNil)
		})
	})
}
