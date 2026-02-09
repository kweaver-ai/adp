// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package metric_model

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"data-model/common"
	"data-model/interfaces"
	dmock "data-model/interfaces/mock"
)

var (
	task = interfaces.MetricTask{
		TaskID:     "1",
		TaskName:   "task1",
		ModuleType: interfaces.MODULE_TYPE_METRIC_MODEL,
		ModelID:    "1",
		Schedule:   interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
		// Filters:            []interfaces.Filter{{Name: "labels.mode", Operation: "=", Value: "idle"}},
		TimeWindows:        []string{"5m", "1h"},
		Steps:              []string{"5m", "1h"},
		IndexBase:          "base1",
		IndexBaseName:      "1",
		RetraceDuration:    "1d",
		ScheduleSyncStatus: 1,
		Comment:            "task1-aaa",
		UpdateTime:         testMetricUpdateTime,
		PlanTime:           int64(1699336878575),
	}
)

func MockNewMetricModelTaskService(appSetting *common.AppSetting,
	mmta interfaces.MetricModelTaskAccess) *metricModelTaskService {

	return &metricModelTaskService{
		appSetting: appSetting,
		mmta:       mmta,
	}
}

func Test_MetricTaskService_CreateMetricTasks(t *testing.T) {
	Convey("Test CreateMetricTasks", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mmta := dmock.NewMockMetricModelTaskAccess(mockCtrl)
		mmts := MockNewMetricModelTaskService(appSetting, mmta)
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		Convey("CreateMetricModel success", func() {
			smock.ExpectBegin()
			mmta.EXPECT().CreateMetricTask(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			tx, _ := db.Begin()
			httpErr := mmts.CreateMetricTask(testCtx, tx, task)
			So(httpErr, ShouldBeNil)
		})
	})
}

func Test_MetricTaskService_GetMetricTaskIDsByModelIDs(t *testing.T) {
	Convey("Test GetMetricTaskIDsByModelIDs", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mmta := dmock.NewMockMetricModelTaskAccess(mockCtrl)
		mmts := MockNewMetricModelTaskService(appSetting, mmta)

		Convey("GetMetricTaskIDsByModelIDs success", func() {
			mmta.EXPECT().GetMetricTaskIDsByModelIDs(gomock.Any(), gomock.Any()).Return([]string{"1", "2"}, nil)

			taskIDs, err := mmts.GetMetricTaskIDsByModelIDs(testCtx, []string{"1"})
			So(taskIDs, ShouldResemble, []string{"1", "2"})
			So(err, ShouldBeNil)
		})
	})
}

func Test_MetricTaskService_UpdateMetricTask(t *testing.T) {
	Convey("Test UpdateMetricTask", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mmta := dmock.NewMockMetricModelTaskAccess(mockCtrl)
		mmts := MockNewMetricModelTaskService(appSetting, mmta)
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		Convey("UpdateMetricTask success", func() {
			smock.ExpectBegin()
			mmta.EXPECT().UpdateMetricTask(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			tx, _ := db.Begin()
			httpErr := mmts.UpdateMetricTask(testCtx, tx, task)
			So(httpErr, ShouldBeNil)
		})
	})
}

func Test_MetricTaskService_GetMetricTasksByTaskIDs(t *testing.T) {
	Convey("Test GetMetricTasksByTaskIDs", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mmta := dmock.NewMockMetricModelTaskAccess(mockCtrl)
		mmts := MockNewMetricModelTaskService(appSetting, mmta)

		Convey("GetMetricTasksByTaskIDs success", func() {
			mmta.EXPECT().GetMetricTasksByTaskIDs(gomock.Any(), gomock.Any()).Return([]interfaces.MetricTask{task}, nil)

			actualTask, httpErr := mmts.GetMetricTasksByTaskIDs(testCtx, []string{"1"})
			So(actualTask, ShouldResemble, []interfaces.MetricTask{task})
			So(httpErr, ShouldBeNil)
		})
	})
}

func Test_MetricTaskService_UpdateMetricTaskStatusInFinish(t *testing.T) {
	Convey("Test UpdateMetricTaskStatusInFinish", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mmta := dmock.NewMockMetricModelTaskAccess(mockCtrl)
		mmts := MockNewMetricModelTaskService(appSetting, mmta)

		Convey("UpdateMetricTaskStatusInFinish success", func() {
			mmta.EXPECT().UpdateMetricTaskStatusInFinish(gomock.Any(), gomock.Any()).Return(nil)

			httpErr := mmts.UpdateMetricTaskStatusInFinish(testCtx, task)
			So(httpErr, ShouldBeNil)
		})
	})
}

func Test_MetricTaskService_UpdateMetricTaskAttributes(t *testing.T) {
	Convey("Test UpdateMetricTaskAttributes", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mmta := dmock.NewMockMetricModelTaskAccess(mockCtrl)
		mmts := MockNewMetricModelTaskService(appSetting, mmta)

		Convey("UpdateMetricTaskAttributes success", func() {
			mmta.EXPECT().UpdateMetricTaskAttributes(gomock.Any(), gomock.Any()).Return(nil)

			httpErr := mmts.UpdateMetricTaskAttributes(testCtx, task)
			So(httpErr, ShouldBeNil)
		})
	})
}

func Test_MetricTaskService_DeleteMetricTaskByTaskID(t *testing.T) {
	Convey("Test DeleteMetricTaskByTaskID", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mmta := dmock.NewMockMetricModelTaskAccess(mockCtrl)
		mmts := MockNewMetricModelTaskService(appSetting, mmta)
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		Convey("DeleteMetricTaskByTaskID success", func() {
			smock.ExpectBegin()
			mmta.EXPECT().DeleteMetricTaskByTaskIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			tx, _ := db.Begin()
			httpErr := mmts.DeleteMetricTaskByTaskIDs(testCtx, tx, []string{"1"})
			So(httpErr, ShouldBeNil)
		})
	})
}
