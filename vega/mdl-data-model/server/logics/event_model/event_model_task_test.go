// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package event_model

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
// 	"data-model/logics/metric_model"
// )

// var (
// 	eventModel = interfaces.EventModel{
// 		EventModelName: "测试中的名称",
// 		EventModelType: "atomic",
// 		EventModelTags: []string{"xx1", "xx2"},
// 		DataSourceType: "metric_model",
// 		DataSource:     []string{"1"},
// 		DetectRule:     detectRule,
// 		AggregateRule:  interfaces.AggregateRule{},
// 		DefaultTimeWindow: interfaces.TimeInterval{
// 			Interval: 5,
// 			Unit:     "m",
// 		},
// 		EventModelComment: "",
// 		IsActive:          1,
// 		IsCustom:          1,
// 		Task:              eventTask,
// 	}
// )

// func TestEventSyncStatusFromXXLJob(t *testing.T) {
// 	Convey("Test SyncStatusFromXXLJob", t, func() {
// 		mockCtrl := gomock.NewController(t)
// 		defer mockCtrl.Finish()

// 		appSetting := &common.AppSetting{}
// 		ema := dmock.NewMockEventModelAccess(mockCtrl)
// 		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
// 		mms := dmock.NewMockMetricModelService(mockCtrl)
// 		dvs := dmock.NewMockDataViewService(mockCtrl)

// 		ems, _ := MockNewEventModelService(appSetting, ema, mms, iba, dvs)

// 		go metric_model.InitClient("http://test")

// 		Convey("failed, becuase of Marshal failed", func() {
// 			err := errors.New("Marshal failed")

// 			patch := ApplyFuncReturn(json.Marshal, []byte{}, err)
// 			defer patch.Reset()

// 			err = emsMock.SyncStatusFromXXLJob(testCtx, []interfaces.EventTask{eventTask})
// 			So(err, ShouldNotBeNil)
// 		})
// 		Convey("failed, becuase of GetEventModelByID failed", func() {
// 			err := errors.New("GetEventModelByID failed")
// 			emaMock.EXPECT().GetEventModelByID(gomock.Any()).Return(interfaces.EventModel{}, err)
// 			eventTask = interfaces.EventTask{
// 				TaskID:   uint64(1),
// 				ModelID:  uint64(1),
// 				Schedule: interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
// 				StorageConfig: interfaces.StorageConfig{
// 					IndexBase:    "base1",
// 					DataViewName: "view1",
// 					DataViewID:   "__base1",
// 				},
// 				DispatchConfig: interfaces.DispatchConfig{
// 					TimeOut:        3600,
// 					RouteStrategy:  "FIRST",
// 					BlockStrategy:  "",
// 					FailRetryCount: 3,
// 				},
// 				ExecuteParameter:   map[string]any{},
// 				TaskStatus:         0,
// 				StatusUpdateTime:   "2023-02-22 15:29:11",
// 				ErrorDetails:       "",
// 				TaskScheduleID:     1,
// 				ScheduleSyncStatus: interfaces.SCHEDULE_SYNC_STATUS_CREATE,
// 				UpdateTime:         "2023-02-22 15:29:11",
// 			}
// 			err = emsMock.SyncStatusFromXXLJob(testCtx, []interfaces.EventTask{eventTask})
// 			So(err, ShouldNotBeNil)
// 		})

// 		Convey("failed, becuase of addJob failed", func() {
// 			err := errors.New("addJob failed")
// 			patch := ApplyFuncReturn(addEventJob, vo.JobInfoAddResp{}, err)
// 			defer patch.Reset()

// 			eventTask = interfaces.EventTask{
// 				TaskID:   uint64(1),
// 				ModelID:  uint64(1),
// 				Schedule: interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
// 				StorageConfig: interfaces.StorageConfig{
// 					IndexBase:    "base1",
// 					DataViewName: "view1",
// 					DataViewID:   "__base1",
// 				},
// 				DispatchConfig: interfaces.DispatchConfig{
// 					TimeOut:        3600,
// 					RouteStrategy:  "FIRST",
// 					BlockStrategy:  "",
// 					FailRetryCount: 3,
// 				},
// 				ExecuteParameter:   map[string]any{},
// 				TaskStatus:         4,
// 				StatusUpdateTime:   "2023-02-22 15:29:11",
// 				ErrorDetails:       "",
// 				TaskScheduleID:     1,
// 				ScheduleSyncStatus: interfaces.SCHEDULE_SYNC_STATUS_CREATE,
// 				UpdateTime:         "2023-02-22 15:29:11",
// 			}
// 			emaMock.EXPECT().GetEventModelByID(gomock.Any()).Return(eventModel, nil)
// 			mmsMock.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(nil, nil)
// 			emaMock.EXPECT().GetEventTaskByModelID(gomock.Any(), gomock.Any()).Return(eventTask, true, nil)

// 			err = emsMock.SyncStatusFromXXLJob(testCtx, []interfaces.EventTask{eventTask})
// 			So(err, ShouldNotBeNil)
// 		})

// 		Convey("failed, becuase of StartJob failed", func() {
// 			err := errors.New("StartJob failed")

// 			emaMock.EXPECT().UpdateEventTaskStatusInFinish(gomock.Any(), gomock.Any()).Return(err)
// 			emaMock.EXPECT().GetEventModelByID(gomock.Any()).Return(eventModel, nil)
// 			mmsMock.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(nil, nil)
// 			emaMock.EXPECT().GetEventTaskByModelID(gomock.Any(), gomock.Any()).Return(eventTask, true, nil)

// 			patch1 := ApplyFuncReturn(addEventJob, vo.JobInfoAddResp{ID: 1}, nil)
// 			defer patch1.Reset()

// 			patch2 := ApplyFuncReturn(metric_model.StartJob, err)
// 			defer patch2.Reset()

// 			err = emsMock.SyncStatusFromXXLJob(testCtx, []interfaces.EventTask{eventTask})
// 			So(err, ShouldNotBeNil)
// 		})
// 		Convey("failed, becuase of UpdateEventTaskStatusInFinish failed", func() {
// 			err := errors.New("UpdateEventTaskStatusInFinish failed")

// 			eventTask = interfaces.EventTask{
// 				TaskID:   uint64(1),
// 				ModelID:  uint64(1),
// 				Schedule: interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
// 				StorageConfig: interfaces.StorageConfig{
// 					IndexBase:    "base1",
// 					DataViewName: "view1",
// 					DataViewID:   "__base1",
// 				},
// 				DispatchConfig: interfaces.DispatchConfig{
// 					TimeOut:        3600,
// 					RouteStrategy:  "FIRST",
// 					BlockStrategy:  "",
// 					FailRetryCount: 3,
// 				},
// 				ExecuteParameter:   map[string]any{},
// 				TaskStatus:         4,
// 				StatusUpdateTime:   "2023-02-22 15:29:11",
// 				ErrorDetails:       "",
// 				TaskScheduleID:     1,
// 				ScheduleSyncStatus: interfaces.SCHEDULE_SYNC_STATUS_CREATE,
// 				UpdateTime:         "2023-02-22 15:29:11",
// 			}

// 			emaMock.EXPECT().GetEventModelByID(gomock.Any()).Return(eventModel, nil)
// 			mmsMock.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(nil, nil)
// 			emaMock.EXPECT().GetEventTaskByModelID(gomock.Any(), gomock.Any()).Return(eventTask, true, nil)

// 			patch1 := ApplyFuncReturn(addEventJob, vo.JobInfoAddResp{ID: 1}, nil)
// 			defer patch1.Reset()

// 			patch2 := ApplyFuncReturn(metric_model.StartJob, nil)
// 			defer patch2.Reset()

// 			emaMock.EXPECT().UpdateEventTaskStatusInFinish(gomock.Any(), gomock.Any()).Return(err)
// 			err = emsMock.SyncStatusFromXXLJob(testCtx, []interfaces.EventTask{eventTask})
// 			So(err, ShouldNotBeNil)
// 		})

// 		Convey("failed, becuase of update scheduleid=0 addEventJob failed", func() {
// 			err := errors.New("addEventJob failed")
// 			patch := ApplyFuncReturn(addEventJob, vo.JobInfoAddResp{}, err)
// 			defer patch.Reset()

// 			eventTask = interfaces.EventTask{
// 				TaskID:   uint64(1),
// 				ModelID:  uint64(1),
// 				Schedule: interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
// 				StorageConfig: interfaces.StorageConfig{
// 					IndexBase:    "base1",
// 					DataViewName: "view1",
// 					DataViewID:   "__base1",
// 				},
// 				DispatchConfig: interfaces.DispatchConfig{
// 					TimeOut:        3600,
// 					RouteStrategy:  "FIRST",
// 					BlockStrategy:  "",
// 					FailRetryCount: 3,
// 				},
// 				ExecuteParameter:   map[string]any{},
// 				TaskStatus:         4,
// 				StatusUpdateTime:   "2023-02-22 15:29:11",
// 				ErrorDetails:       "",
// 				TaskScheduleID:     0,
// 				ScheduleSyncStatus: interfaces.SCHEDULE_SYNC_STATUS_UPDATE,
// 				UpdateTime:         "2023-02-22 15:29:11",
// 			}
// 			// emaMock.EXPECT().UpdateEventTaskStatusInFinish(gomock.Any(), gomock.Any()).Return(err)
// 			err = emsMock.SyncStatusFromXXLJob(testCtx, []interfaces.EventTask{eventTask})
// 			So(err, ShouldNotBeNil)
// 		})

// 		Convey("failed, becuase of update updateEventJob failed", func() {
// 			err := errors.New("updateEventJob failed")

// 			patch := ApplyFuncReturn(updateEventJob, err)
// 			defer patch.Reset()

// 			eventTask = interfaces.EventTask{
// 				TaskID:   uint64(1),
// 				ModelID:  uint64(1),
// 				Schedule: interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
// 				StorageConfig: interfaces.StorageConfig{
// 					IndexBase:    "base1",
// 					DataViewName: "view1",
// 					DataViewID:   "__base1",
// 				},
// 				DispatchConfig: interfaces.DispatchConfig{
// 					TimeOut:        3600,
// 					RouteStrategy:  "FIRST",
// 					BlockStrategy:  "",
// 					FailRetryCount: 3,
// 				},
// 				ExecuteParameter:   map[string]any{},
// 				TaskStatus:         4,
// 				StatusUpdateTime:   "2023-02-22 15:29:11",
// 				ErrorDetails:       "",
// 				TaskScheduleID:     1,
// 				ScheduleSyncStatus: interfaces.SCHEDULE_SYNC_STATUS_UPDATE,
// 				UpdateTime:         "2023-02-22 15:29:11",
// 			}
// 			// emaMock.EXPECT().UpdateEventTaskStatusInFinish(gomock.Any(), gomock.Any()).Return(err)

// 			err = emsMock.SyncStatusFromXXLJob(testCtx, []interfaces.EventTask{eventTask})
// 			So(err, ShouldNotBeNil)
// 		})

// 		Convey("failed, becuase of update UpdateEventTaskStatusInFinish failed", func() {
// 			err := errors.New("UpdateEventTaskStatusInFinish failed")
// 			emaMock.EXPECT().UpdateEventTaskStatusInFinish(gomock.Any(), gomock.Any()).Return(err)

// 			patch := ApplyFuncReturn(updateEventJob, nil)
// 			defer patch.Reset()

// 			eventTask = interfaces.EventTask{
// 				TaskID:   uint64(1),
// 				ModelID:  uint64(1),
// 				Schedule: interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
// 				StorageConfig: interfaces.StorageConfig{
// 					IndexBase:    "base1",
// 					DataViewName: "view1",
// 					DataViewID:   "__base1",
// 				},
// 				DispatchConfig: interfaces.DispatchConfig{
// 					TimeOut:        3600,
// 					RouteStrategy:  "FIRST",
// 					BlockStrategy:  "",
// 					FailRetryCount: 3,
// 				},
// 				ExecuteParameter:   map[string]any{},
// 				TaskStatus:         4,
// 				StatusUpdateTime:   "2023-02-22 15:29:11",
// 				ErrorDetails:       "",
// 				TaskScheduleID:     1,
// 				ScheduleSyncStatus: interfaces.SCHEDULE_SYNC_STATUS_UPDATE,
// 				UpdateTime:         "2023-02-22 15:29:11",
// 			}
// 			// emaMock.EXPECT().UpdateEventTaskStatusInFinish(gomock.Any(), gomock.Any()).Return(err)
// 			err = emsMock.SyncStatusFromXXLJob(testCtx, []interfaces.EventTask{eventTask})
// 			So(err, ShouldNotBeNil)
// 		})

// 		Convey("failed, becuase of delete StopJob failed", func() {
// 			err := errors.New("StopJob failed")

// 			patch := ApplyFuncReturn(metric_model.StopJob, err)
// 			defer patch.Reset()

// 			eventTask = interfaces.EventTask{
// 				TaskID:   uint64(1),
// 				ModelID:  uint64(1),
// 				Schedule: interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
// 				StorageConfig: interfaces.StorageConfig{
// 					IndexBase:    "base1",
// 					DataViewName: "view1",
// 					DataViewID:   "__base1",
// 				},
// 				DispatchConfig: interfaces.DispatchConfig{
// 					TimeOut:        3600,
// 					RouteStrategy:  "FIRST",
// 					BlockStrategy:  "",
// 					FailRetryCount: 3,
// 				},
// 				ExecuteParameter:   map[string]any{},
// 				TaskStatus:         4,
// 				StatusUpdateTime:   "2023-02-22 15:29:11",
// 				ErrorDetails:       "",
// 				TaskScheduleID:     1,
// 				ScheduleSyncStatus: interfaces.SCHEDULE_SYNC_STATUS_DELETE,
// 				UpdateTime:         "2023-02-22 15:29:11",
// 			}
// 			// emaMock.EXPECT().UpdateEventTaskStatusInFinish(gomock.Any(), gomock.Any()).Return(err)

// 			err = emsMock.SyncStatusFromXXLJob(testCtx, []interfaces.EventTask{eventTask})
// 			So(err, ShouldNotBeNil)
// 		})

// 		Convey("failed, becuase of delete deleteJob failed", func() {
// 			err := errors.New("deleteJob failed")

// 			patch1 := ApplyFuncReturn(metric_model.StopJob, nil)
// 			defer patch1.Reset()

// 			patch2 := ApplyFuncReturn(metric_model.DeleteJob, err)
// 			defer patch2.Reset()

// 			eventTask = interfaces.EventTask{
// 				TaskID:   uint64(1),
// 				ModelID:  uint64(1),
// 				Schedule: interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
// 				StorageConfig: interfaces.StorageConfig{
// 					IndexBase:    "base1",
// 					DataViewName: "view1",
// 					DataViewID:   "__base1",
// 				},
// 				DispatchConfig: interfaces.DispatchConfig{
// 					TimeOut:        3600,
// 					RouteStrategy:  "FIRST",
// 					BlockStrategy:  "",
// 					FailRetryCount: 3,
// 				},
// 				ExecuteParameter:   map[string]any{},
// 				TaskStatus:         4,
// 				StatusUpdateTime:   "2023-02-22 15:29:11",
// 				ErrorDetails:       "",
// 				TaskScheduleID:     1,
// 				ScheduleSyncStatus: interfaces.SCHEDULE_SYNC_STATUS_DELETE,
// 				UpdateTime:         "2023-02-22 15:29:11",
// 			}
// 			// emaMock.EXPECT().UpdateEventTaskStatusInFinish(gomock.Any(), gomock.Any()).Return(err)

// 			err = emsMock.SyncStatusFromXXLJob(testCtx, []interfaces.EventTask{eventTask})
// 			So(err, ShouldNotBeNil)
// 		})

// 		Convey("failed, becuase of delete DeleteEventTaskByTaskID failed", func() {
// 			err := errors.New("DeleteEventTaskByTaskID failed")
// 			emaMock.EXPECT().DeleteEventTaskByTaskID(gomock.Any(), gomock.Any()).Return(err)

// 			patch1 := ApplyFuncReturn(metric_model.StopJob, nil)
// 			defer patch1.Reset()

// 			patch2 := ApplyFuncReturn(metric_model.DeleteJob, nil)
// 			defer patch2.Reset()

// 			eventTask = interfaces.EventTask{
// 				TaskID:   uint64(1),
// 				ModelID:  uint64(1),
// 				Schedule: interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
// 				StorageConfig: interfaces.StorageConfig{
// 					IndexBase:    "base1",
// 					DataViewName: "view1",
// 					DataViewID:   "__base1",
// 				},
// 				DispatchConfig: interfaces.DispatchConfig{
// 					TimeOut:        3600,
// 					RouteStrategy:  "FIRST",
// 					BlockStrategy:  "",
// 					FailRetryCount: 3,
// 				},
// 				ExecuteParameter:   map[string]any{},
// 				TaskStatus:         4,
// 				StatusUpdateTime:   "2023-02-22 15:29:11",
// 				ErrorDetails:       "",
// 				TaskScheduleID:     1,
// 				ScheduleSyncStatus: interfaces.SCHEDULE_SYNC_STATUS_DELETE,
// 				UpdateTime:         "2023-02-22 15:29:11",
// 			}
// 			// emaMock.EXPECT().UpdateEventTaskStatusInFinish(gomock.Any(), gomock.Any()).Return(err)

// 			err = emsMock.SyncStatusFromXXLJob(testCtx, []interfaces.EventTask{eventTask})
// 			So(err, ShouldNotBeNil)
// 		})

// 	})
// }
