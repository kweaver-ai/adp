// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package metric_model

// import (
// 	"encoding/json"
// 	"errors"
// 	"testing"

// 	"devops.aishu.cn/AISHUDevOps/DIP/_git/mdl-XxlJobSDK-Go.git/sdk/vo"
// 	. "github.com/agiledragon/gomonkey/v2"
// 	"github.com/golang/mock/gomock"
// 	. "github.com/smartystreets/goconvey/convey"

// 	"data-model/common"
// 	"data-model/interfaces"
// 	dmock "data-model/interfaces/mock"
// )

// func TestSyncStatusFromXXLJob(t *testing.T) {
// 	Convey("Test SyncStatusFromXXLJob", t, func() {
// 		mockCtrl := gomock.NewController(t)
// 		defer mockCtrl.Finish()

// 		appSetting := &common.AppSetting{}
// 		dvs := dmock.NewMockDataViewService(mockCtrl)
// 		mma := dmock.NewMockMetricModelAccess(mockCtrl)
// 		mmga := dmock.NewMockMetricModelGroupAccess(mockCtrl)
// 		ua := dmock.NewMockUniqueryAccess(mockCtrl)
// 		mmts := dmock.NewMockMetricModelTaskService(mockCtrl)
// 		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
// 		mms, _ := MockNewMetricModelService(appSetting, dvs, mma, mmga, ua, mmts, iba)

// 		go InitClient("test")

// 		Convey("failed, becuase of GetMetricModelByModelID failed", func() {
// 			err := errors.New("Check Metric Model Exist By Name failed")
// 			mmaMock.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(interfaces.MetricModel{}, false, err)

// 			err = mmService.SyncStatusFromXXLJob(testCtx, []interfaces.MetricTask{task})
// 			So(err, ShouldNotBeNil)
// 		})

// 		Convey("failed, becuase of Marshal failed", func() {
// 			err := errors.New("Marshal failed")
// 			mmaMock.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(interfaces.MetricModel{}, true, nil)

// 			patch := ApplyFuncReturn(json.Marshal, []byte{}, err)
// 			defer patch.Reset()

// 			err = mmService.SyncStatusFromXXLJob(testCtx, []interfaces.MetricTask{task})
// 			So(err, ShouldNotBeNil)
// 		})

// 		Convey("failed, becuase of AddJob failed", func() {
// 			err := errors.New("AddJob failed")
// 			mmaMock.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(interfaces.MetricModel{}, true, nil)
// 			patch := ApplyFuncReturn(AddJob, vo.JobInfoAddResp{}, err)
// 			defer patch.Reset()

// 			cTask := interfaces.MetricTask{
// 				TaskID:     "uint64(1)",
// 				TaskName:   "task1",
// 				ModuleType: interfaces.MODULE_TYPE_METRIC_MODEL,
// 				ModelID:    "1",
// 				Schedule:   interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
// 				// Filters:            []interfaces.Filter{{Name: "labels.mode", Operation: "=", Value: "idle"}},
// 				TimeWindows:        []string{"5m", "1h"},
// 				IndexBase:          "base1",
// 				RetraceDuration:    "1d",
// 				TaskScheduleID:     0,
// 				ScheduleSyncStatus: 0,
// 				Comment:            "task1-aaa",
// 				UpdateTime:         testMetricUpdateTime,
// 				PlanTimes:          []int64{1699336878575, 1699336878575},
// 			}
// 			err = mmService.SyncStatusFromXXLJob(testCtx, []interfaces.MetricTask{cTask})
// 			So(err, ShouldNotBeNil)
// 		})

// 		Convey("failed, becuase of StartJob failed", func() {
// 			err := errors.New("StartJob failed")

// 			mmaMock.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(interfaces.MetricModel{}, true, nil)

// 			patch1 := ApplyFuncReturn(AddJob, vo.JobInfoAddResp{ID: 1}, nil)
// 			defer patch1.Reset()

// 			patch2 := ApplyFuncReturn(StartJob, err)
// 			defer patch2.Reset()

// 			cTask := interfaces.MetricTask{
// 				TaskID:     "uint64(1)",
// 				TaskName:   "task1",
// 				ModuleType: interfaces.MODULE_TYPE_METRIC_MODEL,
// 				ModelID:    "1",
// 				Schedule:   interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
// 				// Filters:            []interfaces.Filter{{Name: "labels.mode", Operation: "=", Value: "idle"}},
// 				TimeWindows:        []string{"5m", "1h"},
// 				IndexBase:          "base1",
// 				RetraceDuration:    "1d",
// 				TaskScheduleID:     0,
// 				ScheduleSyncStatus: 0,
// 				Comment:            "task1-aaa",
// 				UpdateTime:         testMetricUpdateTime,
// 				PlanTimes:          []int64{1699336878575, 1699336878575},
// 			}
// 			err = mmService.SyncStatusFromXXLJob(testCtx, []interfaces.MetricTask{cTask})
// 			So(err, ShouldNotBeNil)
// 		})

// 		Convey("failed, becuase of UpdateMetricTaskStatusInFinish failed", func() {
// 			err := errors.New("UpdateMetricTaskStatusInFinish failed")
// 			mmaMock.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(interfaces.MetricModel{}, true, nil)
// 			mtsMock.EXPECT().UpdateMetricTaskStatusInFinish(gomock.Any(), gomock.Any()).Return(err)

// 			patch1 := ApplyFuncReturn(AddJob, vo.JobInfoAddResp{ID: 1}, nil)
// 			defer patch1.Reset()

// 			patch2 := ApplyFuncReturn(StartJob, nil)
// 			defer patch2.Reset()

// 			cTask := interfaces.MetricTask{
// 				TaskID:     "uint64(1)",
// 				TaskName:   "task1",
// 				ModuleType: interfaces.MODULE_TYPE_METRIC_MODEL,
// 				ModelID:    "1",
// 				Schedule:   interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
// 				// Filters:            []interfaces.Filter{{Name: "labels.mode", Operation: "=", Value: "idle"}},
// 				TimeWindows:        []string{"5m", "1h"},
// 				IndexBase:          "base1",
// 				RetraceDuration:    "1d",
// 				TaskScheduleID:     0,
// 				ScheduleSyncStatus: 0,
// 				Comment:            "task1-aaa",
// 				UpdateTime:         testMetricUpdateTime,
// 				PlanTimes:          []int64{1699336878575, 1699336878575},
// 			}
// 			err = mmService.SyncStatusFromXXLJob(testCtx, []interfaces.MetricTask{cTask})
// 			So(err, ShouldNotBeNil)
// 		})

// 		Convey("failed, becuase of update scheduleid=0 AddJob failed", func() {
// 			err := errors.New("AddJob failed")
// 			mmaMock.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(interfaces.MetricModel{}, true, nil)
// 			patch := ApplyFuncReturn(AddJob, vo.JobInfoAddResp{}, err)
// 			defer patch.Reset()

// 			cTask := interfaces.MetricTask{
// 				TaskID:     "uint64(1)",
// 				TaskName:   "task1",
// 				ModuleType: interfaces.MODULE_TYPE_METRIC_MODEL,
// 				ModelID:    "1",
// 				Schedule:   interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
// 				// Filters:            []interfaces.Filter{{Name: "labels.mode", Operation: "=", Value: "idle"}},
// 				TimeWindows:        []string{"5m", "1h"},
// 				IndexBase:          "base1",
// 				RetraceDuration:    "1d",
// 				TaskScheduleID:     0,
// 				ScheduleSyncStatus: 1,
// 				Comment:            "task1-aaa",
// 				UpdateTime:         testMetricUpdateTime,
// 				PlanTimes:          []int64{1699336878575, 1699336878575},
// 			}
// 			err = mmService.SyncStatusFromXXLJob(testCtx, []interfaces.MetricTask{cTask})
// 			So(err, ShouldNotBeNil)
// 		})

// 		Convey("failed, becuase of update scheduleid=0 StartJob failed", func() {
// 			err := errors.New("StartJob failed")
// 			mmaMock.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(interfaces.MetricModel{}, true, nil)

// 			patch1 := ApplyFuncReturn(AddJob, vo.JobInfoAddResp{ID: 1}, nil)
// 			defer patch1.Reset()

// 			patch2 := ApplyFuncReturn(StartJob, err)
// 			defer patch2.Reset()

// 			cTask := interfaces.MetricTask{
// 				TaskID:     "uint64(1)",
// 				TaskName:   "task1",
// 				ModuleType: interfaces.MODULE_TYPE_METRIC_MODEL,
// 				ModelID:    "1",
// 				Schedule:   interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
// 				// Filters:            []interfaces.Filter{{Name: "labels.mode", Operation: "=", Value: "idle"}},
// 				TimeWindows:        []string{"5m", "1h"},
// 				IndexBase:          "base1",
// 				RetraceDuration:    "1d",
// 				TaskScheduleID:     0,
// 				ScheduleSyncStatus: 1,
// 				Comment:            "task1-aaa",
// 				UpdateTime:         testMetricUpdateTime,
// 				PlanTimes:          []int64{1699336878575, 1699336878575},
// 			}
// 			err = mmService.SyncStatusFromXXLJob(testCtx, []interfaces.MetricTask{cTask})
// 			So(err, ShouldNotBeNil)
// 		})

// 		Convey("failed, becuase of update UpdateJob failed", func() {
// 			err := errors.New("UpdateJob failed")
// 			mmaMock.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(interfaces.MetricModel{}, true, nil)

// 			patch := ApplyFuncReturn(UpdateJob, err)
// 			defer patch.Reset()

// 			cTask := interfaces.MetricTask{
// 				TaskID:     "uint64(1)",
// 				TaskName:   "task1",
// 				ModuleType: interfaces.MODULE_TYPE_METRIC_MODEL,
// 				ModelID:    "1",
// 				Schedule:   interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
// 				// Filters:            []interfaces.Filter{{Name: "labels.mode", Operation: "=", Value: "idle"}},
// 				TimeWindows:        []string{"5m", "1h"},
// 				IndexBase:          "base1",
// 				RetraceDuration:    "1d",
// 				TaskScheduleID:     2,
// 				ScheduleSyncStatus: 1,
// 				Comment:            "task1-aaa",
// 				UpdateTime:         testMetricUpdateTime,
// 				PlanTimes:          []int64{1699336878575, 1699336878575},
// 			}
// 			err = mmService.SyncStatusFromXXLJob(testCtx, []interfaces.MetricTask{cTask})
// 			So(err, ShouldNotBeNil)
// 		})

// 		Convey("failed, becuase of update UpdateMetricTaskStatusInFinish failed", func() {
// 			err := errors.New("UpdateMetricTaskStatusInFinish failed")
// 			mmaMock.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(interfaces.MetricModel{}, true, nil)
// 			mtsMock.EXPECT().UpdateMetricTaskStatusInFinish(gomock.Any(), gomock.Any()).Return(err)

// 			patch := ApplyFuncReturn(UpdateJob, nil)
// 			defer patch.Reset()

// 			cTask := interfaces.MetricTask{
// 				TaskID:     "uint64(1)",
// 				TaskName:   "task1",
// 				ModuleType: interfaces.MODULE_TYPE_METRIC_MODEL,
// 				ModelID:    "1",
// 				Schedule:   interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
// 				// Filters:            []interfaces.Filter{{Name: "labels.mode", Operation: "=", Value: "idle"}},
// 				TimeWindows:        []string{"5m", "1h"},
// 				IndexBase:          "base1",
// 				RetraceDuration:    "1d",
// 				TaskScheduleID:     2,
// 				ScheduleSyncStatus: 1,
// 				Comment:            "task1-aaa",
// 				UpdateTime:         testMetricUpdateTime,
// 				PlanTimes:          []int64{1699336878575, 1699336878575},
// 			}
// 			err = mmService.SyncStatusFromXXLJob(testCtx, []interfaces.MetricTask{cTask})
// 			So(err, ShouldNotBeNil)
// 		})

// 		Convey("failed, becuase of delete StopJob failed", func() {
// 			err := errors.New("StopJob failed")

// 			patch := ApplyFuncReturn(StopJob, err)
// 			defer patch.Reset()

// 			cTask := interfaces.MetricTask{
// 				TaskID:   "uint64(1)",
// 				TaskName: "task1",
// 				ModelID:  "1",
// 				Schedule: interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
// 				// Filters:            []interfaces.Filter{{Name: "labels.mode", Operation: "=", Value: "idle"}},
// 				TimeWindows:        []string{"5m", "1h"},
// 				IndexBase:          "base1",
// 				RetraceDuration:    "1d",
// 				TaskScheduleID:     2,
// 				ScheduleSyncStatus: 2,
// 				Comment:            "task1-aaa",
// 				UpdateTime:         testMetricUpdateTime,
// 				PlanTimes:          []int64{1699336878575, 1699336878575},
// 			}
// 			err = mmService.SyncStatusFromXXLJob(testCtx, []interfaces.MetricTask{cTask})
// 			So(err, ShouldNotBeNil)
// 		})

// 		Convey("failed, becuase of delete DeleteJob failed", func() {
// 			err := errors.New("DeleteJob failed")

// 			patch1 := ApplyFuncReturn(StopJob, nil)
// 			defer patch1.Reset()

// 			patch2 := ApplyFuncReturn(DeleteJob, err)
// 			defer patch2.Reset()

// 			cTask := interfaces.MetricTask{
// 				TaskID:   "uint64(1)",
// 				TaskName: "task1",
// 				ModelID:  "1",
// 				Schedule: interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
// 				// Filters:            []interfaces.Filter{{Name: "labels.mode", Operation: "=", Value: "idle"}},
// 				TimeWindows:        []string{"5m", "1h"},
// 				IndexBase:          "base1",
// 				RetraceDuration:    "1d",
// 				TaskScheduleID:     2,
// 				ScheduleSyncStatus: 2,
// 				Comment:            "task1-aaa",
// 				UpdateTime:         testMetricUpdateTime,
// 				PlanTimes:          []int64{1699336878575, 1699336878575},
// 			}
// 			err = mmService.SyncStatusFromXXLJob(testCtx, []interfaces.MetricTask{cTask})
// 			So(err, ShouldNotBeNil)
// 		})

// 		Convey("failed, becuase of delete UpdateMetricTaskStatusInFinish failed", func() {
// 			err := errors.New("UpdateMetricTaskStatusInFinish failed")
// 			// mmaMock.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(interfaces.MetricModel{MeasureName: "a"}, true, nil)
// 			mtsMock.EXPECT().DeleteMetricTaskByTaskID(gomock.Any(), gomock.Any()).Return(err)

// 			patch1 := ApplyFuncReturn(StopJob, nil)
// 			defer patch1.Reset()

// 			patch2 := ApplyFuncReturn(DeleteJob, nil)
// 			defer patch2.Reset()

// 			cTask := interfaces.MetricTask{
// 				TaskID:   "uint64(1)",
// 				TaskName: "task1",
// 				ModelID:  "1",
// 				Schedule: interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
// 				// Filters:            []interfaces.Filter{{Name: "labels.mode", Operation: "=", Value: "idle"}},
// 				TimeWindows:        []string{"5m", "1h"},
// 				IndexBase:          "base1",
// 				RetraceDuration:    "1d",
// 				TaskScheduleID:     2,
// 				ScheduleSyncStatus: 2,
// 				Comment:            "task1-aaa",
// 				UpdateTime:         testMetricUpdateTime,
// 				PlanTimes:          []int64{1699336878575, 1699336878575},
// 			}
// 			err = mmService.SyncStatusFromXXLJob(testCtx, []interfaces.MetricTask{cTask})
// 			So(err, ShouldNotBeNil)
// 		})

// 	})
// }
