// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package objective_model

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	"data-model/common"
	derrors "data-model/errors"
	"data-model/interfaces"
	dmock "data-model/interfaces/mock"
)

var (
	testCtx = context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage)
)

func MockNewObjectiveModelService(appSetting *common.AppSetting,
	dmja interfaces.DataModelJobAccess,
	mms interfaces.MetricModelService,
	oma interfaces.ObjectiveModelAccess,
	mmts interfaces.MetricModelTaskService,
	iba interfaces.IndexBaseAccess,
	ps interfaces.PermissionService) (*objectiveModelService, sqlmock.Sqlmock) {

	db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	oms := &objectiveModelService{
		appSetting: appSetting,
		db:         db,
		dmja:       dmja,
		iba:        iba,
		mms:        mms,
		mmts:       mmts,
		oma:        oma,
		ps:         ps,
	}
	return oms, smock
}

func Test_ObjectiveModelService_CheckObjectiveModelExistByID(t *testing.T) {
	Convey("Test CheckObjectiveModelExistByID", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
		oma := dmock.NewMockObjectiveModelAccess(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)
		mmts := dmock.NewMockMetricModelTaskService(mockCtrl)
		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		oms, _ := MockNewObjectiveModelService(appSetting, dmja, mms, oma, mmts, iba, ps)

		Convey("When objective model exists", func() {
			oma.EXPECT().CheckObjectiveModelExistByID(gomock.Any(), gomock.Any()).Return("model-1", true, nil)

			name, exists, err := oms.CheckObjectiveModelExistByID(testCtx, "test-id")
			So(err, ShouldBeNil)
			So(exists, ShouldBeTrue)
			So(name, ShouldEqual, "model-1")
		})

		Convey("When objective model does not exist", func() {

			oma.EXPECT().CheckObjectiveModelExistByID(gomock.Any(), gomock.Any()).Return("", false, nil)

			name, exists, err := oms.CheckObjectiveModelExistByID(testCtx, "test-id")
			So(err, ShouldBeNil)
			So(exists, ShouldBeFalse)
			So(name, ShouldBeEmpty)
		})

		Convey("When error occurs", func() {
			expectedErr := errors.New("database error")
			oma.EXPECT().CheckObjectiveModelExistByID(gomock.Any(), gomock.Any()).Return("", false, expectedErr)

			name, exists, err := oms.CheckObjectiveModelExistByID(testCtx, "test-id")
			So(err, ShouldNotBeNil)
			So(exists, ShouldBeFalse)
			So(name, ShouldBeEmpty)
		})
	})
}

func Test_ObjectiveModelService_CheckObjectiveModelExistByName(t *testing.T) {
	Convey("Test CheckObjectiveModelExistByName", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
		oma := dmock.NewMockObjectiveModelAccess(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)
		mmts := dmock.NewMockMetricModelTaskService(mockCtrl)
		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		oms, _ := MockNewObjectiveModelService(appSetting, dmja, mms, oma, mmts, iba, ps)

		Convey("When objective model exists", func() {
			oma.EXPECT().CheckObjectiveModelExistByName(gomock.Any(), gomock.Any()).Return("sss", true, nil)

			id, exist, err := oms.CheckObjectiveModelExistByName(testCtx, "test-model")
			So(id, ShouldEqual, "sss")
			So(exist, ShouldBeTrue)
			So(err, ShouldBeNil)
		})

		Convey("When objective model does not exist", func() {
			oma.EXPECT().CheckObjectiveModelExistByName(gomock.Any(), gomock.Any()).Return("", false, nil)

			_, _, err := oms.CheckObjectiveModelExistByName(testCtx, "test-model")
			So(err, ShouldBeNil)
		})

		Convey("When error occurs", func() {
			expectedErr := errors.New("database error")
			oma.EXPECT().CheckObjectiveModelExistByName(gomock.Any(), gomock.Any()).Return("", false, expectedErr)

			_, _, err := oms.CheckObjectiveModelExistByName(testCtx, "test-model")
			So(err, ShouldNotBeNil)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InternalError_CheckModelIfExistFailed)
		})
	})
}

func Test_ObjectiveModelService_CreateObjectiveModel(t *testing.T) {
	Convey("Test CreateObjectiveModel", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
		oma := dmock.NewMockObjectiveModelAccess(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)
		mmts := dmock.NewMockMetricModelTaskService(mockCtrl)
		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		oms, smock := MockNewObjectiveModelService(appSetting, dmja, mms, oma, mmts, iba, ps)

		Convey("When check metric model exists fails", func() {
			objective := float64(99)
			period := int64(100)
			model := interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.SLO,
					ObjectiveConfig: interfaces.SLOObjective{
						Objective: &objective,
						Period:    &period,
						GoodMetricModel: &interfaces.BundleMetricModel{
							ID: "123",
						},
						TotalMetricModel: &interfaces.BundleMetricModel{
							ID: "456",
						},
					},
				},
				Task: &interfaces.MetricTask{
					Schedule: interfaces.Schedule{
						Type:       interfaces.SCHEDULE_TYPE_CRON,
						Expression: "* * * * * *",
					},
					Steps:           []string{"1h"},
					RetraceDuration: "24h",
					IndexBase:       "test-index",
				},
			}

			expectedErr := errors.New("check metric model error")
			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.SimpleMetricModel{}, expectedErr)
			dmja.EXPECT().StartJob(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			_, err := oms.CreateObjectiveModel(testCtx, model)
			So(err, ShouldNotBeNil)
		})

		Convey("When check metric model exists fails with KPI type", func() {
			objective := float64(99)
			weight := int64(100)
			model := interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "ms",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "123",
								Weight: &weight,
							},
						},
					},
				},
				Task: &interfaces.MetricTask{
					Schedule: interfaces.Schedule{
						Type:       interfaces.SCHEDULE_TYPE_CRON,
						Expression: "* * * * * *",
					},
					Steps:           []string{"1h"},
					RetraceDuration: "24h",
					IndexBase:       "test-index",
				},
			}

			expectedErr := errors.New("check metric model error")
			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.SimpleMetricModel{}, expectedErr)
			dmja.EXPECT().StartJob(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			_, err := oms.CreateObjectiveModel(testCtx, model)
			So(err, ShouldNotBeNil)
		})

		Convey("When check index base fails", func() {
			objective := float64(99)
			weight := int64(100)
			model := interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "ms",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "123",
								Weight: &weight,
							},
						},
					},
				},
				Task: &interfaces.MetricTask{
					Schedule: interfaces.Schedule{
						Type:       interfaces.SCHEDULE_TYPE_CRON,
						Expression: "* * * * * *",
					},
					Steps:           []string{"1h"},
					RetraceDuration: "24h",
					IndexBase:       "test-index",
				},
			}

			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.SimpleMetricModel{
				"123": {ModelName: "test-metric-model"},
			}, nil)
			expectedErr := errors.New("check index base error")
			iba.EXPECT().GetSimpleIndexBasesByTypes(gomock.Any(), gomock.Any()).Return(nil, expectedErr)
			dmja.EXPECT().StartJob(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			_, err := oms.CreateObjectiveModel(testCtx, model)
			So(err, ShouldNotBeNil)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InternalError_GetSimpleIndexBasesByTypesFailed)
		})

		Convey("When parse retrace duration fails", func() {
			objective := float64(99)
			weight := int64(100)
			model := interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "ms",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "123",
								Weight: &weight,
							},
						},
					},
				},
				Task: &interfaces.MetricTask{
					Schedule: interfaces.Schedule{
						Type:       interfaces.SCHEDULE_TYPE_CRON,
						Expression: "* * * * * *",
					},
					Steps:           []string{"1h"},
					RetraceDuration: "invalid",
					IndexBase:       "test-index",
				},
			}

			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.SimpleMetricModel{
				"123": {ModelName: "test-metric-model"},
			}, nil)
			iba.EXPECT().GetSimpleIndexBasesByTypes(gomock.Any(), gomock.Any()).Return([]interfaces.SimpleIndexBase{
				{Name: "test-index"},
			}, nil)
			dmja.EXPECT().StartJob(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			_, err := oms.CreateObjectiveModel(testCtx, model)
			So(err, ShouldNotBeNil)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InvalidParameter_RetraceDuration)
		})

		Convey("When begin transaction fails", func() {
			objective := float64(99)
			weight := int64(100)
			model := interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "ms",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "123",
								Weight: &weight,
							},
						},
					},
				},
				Task: &interfaces.MetricTask{
					Schedule: interfaces.Schedule{
						Type:       interfaces.SCHEDULE_TYPE_CRON,
						Expression: "* * * * * *",
					},
					Steps:           []string{"1h"},
					RetraceDuration: "1h",
					IndexBase:       "test-index",
				},
			}

			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.SimpleMetricModel{
				"123": {ModelName: "test-metric-model"},
			}, nil)
			iba.EXPECT().GetSimpleIndexBasesByTypes(gomock.Any(), gomock.Any()).Return([]interfaces.SimpleIndexBase{
				{Name: "test-index"},
			}, nil)
			dmja.EXPECT().StartJob(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			smock.ExpectBegin().WillReturnError(errors.New("begin transaction failed"))

			_, err := oms.CreateObjectiveModel(testCtx, model)
			So(err, ShouldNotBeNil)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InternalError_BeginTransactionFailed)
		})

		Convey("When commit transaction fails", func() {
			objective := float64(99)
			weight := int64(100)
			model := interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "ms",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "123",
								Weight: &weight,
							},
						},
					},
				},
				Task: &interfaces.MetricTask{
					Schedule: interfaces.Schedule{
						Type:       interfaces.SCHEDULE_TYPE_CRON,
						Expression: "* * * * * *",
					},
					Steps:           []string{"1h"},
					RetraceDuration: "1h",
					IndexBase:       "test-index",
				},
			}

			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.SimpleMetricModel{
				"123": {ModelName: "test-metric-model"},
			}, nil)
			iba.EXPECT().GetSimpleIndexBasesByTypes(gomock.Any(), gomock.Any()).Return([]interfaces.SimpleIndexBase{
				{Name: "test-index"},
			}, nil)
			smock.ExpectBegin()
			oma.EXPECT().CreateObjectiveModel(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mmts.EXPECT().CreateMetricTask(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			smock.ExpectCommit().WillReturnError(errors.New("commit transaction failed"))
			dmja.EXPECT().StartJob(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			_, err := oms.CreateObjectiveModel(testCtx, model)
			So(err, ShouldResemble, errors.New("commit transaction failed"))
		})

		Convey("When create objective model fails", func() {
			objective := float64(99)
			weight := int64(100)
			model := interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "ms",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "123",
								Weight: &weight,
							},
						},
					},
				},
				Task: &interfaces.MetricTask{
					Schedule: interfaces.Schedule{
						Type:       interfaces.SCHEDULE_TYPE_CRON,
						Expression: "* * * * * *",
					},
					Steps:           []string{"1h"},
					RetraceDuration: "1h",
					IndexBase:       "test-index",
				},
			}

			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.SimpleMetricModel{
				"123": {ModelName: "test-metric-model"},
			}, nil)
			iba.EXPECT().GetSimpleIndexBasesByTypes(gomock.Any(), gomock.Any()).Return([]interfaces.SimpleIndexBase{
				{Name: "test-index"},
			}, nil)
			smock.ExpectBegin()
			oma.EXPECT().CreateObjectiveModel(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("create objective model failed"))
			smock.ExpectRollback()
			dmja.EXPECT().StartJob(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			_, err := oms.CreateObjectiveModel(testCtx, model)
			So(err, ShouldNotBeNil)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InternalError)
		})

		Convey("When create metric task fails", func() {
			objective := float64(99)
			weight := int64(100)
			model := interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "ms",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "123",
								Weight: &weight,
							},
						},
					},
				},
				Task: &interfaces.MetricTask{
					Schedule: interfaces.Schedule{
						Type:       interfaces.SCHEDULE_TYPE_CRON,
						Expression: "* * * * * *",
					},
					Steps:           []string{"1h"},
					RetraceDuration: "1h",
					IndexBase:       "test-index",
				},
			}

			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.SimpleMetricModel{
				"123": {ModelName: "test-metric-model"},
			}, nil)
			iba.EXPECT().GetSimpleIndexBasesByTypes(gomock.Any(), gomock.Any()).Return([]interfaces.SimpleIndexBase{
				{Name: "test-index"},
			}, nil)
			smock.ExpectBegin()
			oma.EXPECT().CreateObjectiveModel(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mmts.EXPECT().CreateMetricTask(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("create metric task failed"))
			smock.ExpectRollback()
			dmja.EXPECT().StartJob(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			_, err := oms.CreateObjectiveModel(testCtx, model)
			So(err, ShouldNotBeNil)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InternalError)
		})
	})
}

func Test_ObjectiveModelService_ListObjectiveModels(t *testing.T) {
	Convey("Test ListObjectiveModels", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
		oma := dmock.NewMockObjectiveModelAccess(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)
		mmts := dmock.NewMockMetricModelTaskService(mockCtrl)
		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		oms, _ := MockNewObjectiveModelService(appSetting, dmja, mms, oma, mmts, iba, ps)

		resrc := map[string]interfaces.ResourceOps{
			"test-id": {
				ResourceID: "test-id",
			},
		}
		Convey("When list objective models fails", func() {
			params := interfaces.ObjectiveModelsQueryParams{}
			expectedErr := errors.New("list objective models error")
			oma.EXPECT().ListObjectiveModels(gomock.Any(), gomock.Any()).Return(nil, expectedErr)

			_, _, err := oms.ListObjectiveModels(testCtx, params)
			So(err, ShouldNotBeNil)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InternalError)
		})

		Convey("When get metric tasks fails", func() {
			params := interfaces.ObjectiveModelsQueryParams{}
			models := []interfaces.ObjectiveModel{
				{
					ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
						ModelID:   "test-id",
						ModelName: "test-model",
					},
				},
			}
			oma.EXPECT().ListObjectiveModels(gomock.Any(), gomock.Any()).Return(models, nil)
			mmts.EXPECT().GetMetricTasksByModelIDs(gomock.Any(), gomock.Any()).Return(nil, errors.New("get metric tasks error"))
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)

			_, _, err := oms.ListObjectiveModels(testCtx, params)
			So(err, ShouldNotBeNil)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InternalError_GetMetricTasksByModelIDsFailed)
		})

		Convey("When GetMetricModelSimpleInfosByIDs fails for SLO type", func() {
			params := interfaces.ObjectiveModelsQueryParams{}
			models := []interfaces.ObjectiveModel{
				{
					ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
						ModelID:       "test-id",
						ModelName:     "test-model",
						ObjectiveType: interfaces.SLO,
						ObjectiveConfig: interfaces.SLOObjective{
							GoodMetricModel: &interfaces.BundleMetricModel{
								ID: "123",
							},
							TotalMetricModel: &interfaces.BundleMetricModel{
								ID: "456",
							},
						},
					},
					Task: &interfaces.MetricTask{
						IndexBase: "test-index",
					},
				},
			}
			task := interfaces.MetricTask{
				ModelID:   "test-id",
				IndexBase: "test-index",
			}
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			oma.EXPECT().ListObjectiveModels(gomock.Any(), gomock.Any()).Return(models, nil)
			mmts.EXPECT().GetMetricTasksByModelIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.MetricTask{"test-id": task}, nil)
			iba.EXPECT().GetSimpleIndexBasesByTypes(gomock.Any(), gomock.Any()).Return([]interfaces.SimpleIndexBase{{Name: "test-index"}}, nil)
			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(nil, errors.New("get metric model error"))

			_, _, err := oms.ListObjectiveModels(testCtx, params)
			So(err, ShouldNotBeNil)
		})

		Convey("When GetMetricModelSimpleInfosByIDs fails for KPI type", func() {
			params := interfaces.ObjectiveModelsQueryParams{}
			weight := int64(100)
			models := []interfaces.ObjectiveModel{
				{
					ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
						ModelID:       "test-id",
						ModelName:     "test-model",
						ObjectiveType: interfaces.KPI,
						ObjectiveConfig: interfaces.KPIObjective{
							ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
								{
									ID:     "123",
									Weight: &weight,
								},
							},
						},
					},
					Task: &interfaces.MetricTask{
						IndexBase: "test-index",
					},
				},
			}
			task := interfaces.MetricTask{
				ModelID:   "test-id",
				IndexBase: "test-index",
			}
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			oma.EXPECT().ListObjectiveModels(gomock.Any(), gomock.Any()).Return(models, nil)
			mmts.EXPECT().GetMetricTasksByModelIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.MetricTask{"test-id": task}, nil)
			iba.EXPECT().GetSimpleIndexBasesByTypes(gomock.Any(), gomock.Any()).Return([]interfaces.SimpleIndexBase{{Name: "test-index"}}, nil)
			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(nil, errors.New("get metric model error"))

			_, _, err := oms.ListObjectiveModels(testCtx, params)
			So(err, ShouldNotBeNil)
		})

		Convey("When list objective models succeeds with SLO type", func() {
			params := interfaces.ObjectiveModelsQueryParams{
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit: 10,
				},
			}
			objective := float64(99)
			period := int64(100)
			weight := int64(100)
			models := []interfaces.ObjectiveModel{
				{
					ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
						ModelID:       "test-id",
						ModelName:     "test-model",
						ObjectiveType: interfaces.SLO,
						ObjectiveConfig: interfaces.SLOObjective{
							Objective: &objective,
							Period:    &period,
							GoodMetricModel: &interfaces.BundleMetricModel{
								ID: "123",
							},
							TotalMetricModel: &interfaces.BundleMetricModel{
								ID: "456",
							},
						},
					},
					Task: &interfaces.MetricTask{
						IndexBase: "test-index",
					},
				},
				{
					ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
						ModelID:       "test-id2",
						ModelName:     "test-model",
						ObjectiveType: interfaces.KPI,
						ObjectiveConfig: interfaces.KPIObjective{
							Objective: &objective,
							Unit:      "ms",
							ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
								{
									ID:     "123",
									Weight: &weight,
								},
							},
							AdditionalMetricModels: []interfaces.BundleMetricModel{
								{
									ID: "456",
								},
							},
						},
					},
					Task: &interfaces.MetricTask{
						IndexBase: "test-index",
					},
				},
			}
			task := interfaces.MetricTask{
				ModelID:   "test-id",
				IndexBase: "test-index",
			}
			resrctmp := map[string]interfaces.ResourceOps{
				"test-id": {
					ResourceID: "test-id",
				},
				"test-id2": {
					ResourceID: "test-id2",
				},
			}
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrctmp, nil)
			oma.EXPECT().ListObjectiveModels(gomock.Any(), gomock.Any()).Return(models, nil)
			mmts.EXPECT().GetMetricTasksByModelIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.MetricTask{"test-id": task}, nil)
			iba.EXPECT().GetSimpleIndexBasesByTypes(gomock.Any(), gomock.Any()).AnyTimes().Return([]interfaces.SimpleIndexBase{{Name: "test-index"}}, nil)
			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).AnyTimes().Return(map[string]interfaces.SimpleMetricModel{
				"123": {ModelName: "good-metric"},
				"456": {ModelName: "total-metric"},
			}, nil)

			result, total, err := oms.ListObjectiveModels(testCtx, params)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 2)
			So(len(result), ShouldEqual, 2)
			So(result[0].ModelID, ShouldEqual, "test-id")
			So(result[0].ObjectiveType, ShouldEqual, interfaces.SLO)
		})
	})
}

func Test_ObjectiveModelService_GetObjectiveModels(t *testing.T) {
	Convey("Test GetObjectiveModels", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
		oma := dmock.NewMockObjectiveModelAccess(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)
		mmts := dmock.NewMockMetricModelTaskService(mockCtrl)
		iba := dmock.NewMockIndexBaseAccess(mockCtrl)

		ps := dmock.NewMockPermissionService(mockCtrl)
		oms, _ := MockNewObjectiveModelService(appSetting, dmja, mms, oma, mmts, iba, ps)

		resrc := map[string]interfaces.ResourceOps{
			"test-id": {
				ResourceID: "test-id",
			},
		}
		Convey("When get objective models fails", func() {
			modelIDs := []string{"test-id"}
			expectedErr := errors.New("get objective models error")
			oma.EXPECT().GetObjectiveModelsByModelIDs(gomock.Any(), gomock.Any()).Return(nil, expectedErr)

			_, err := oms.GetObjectiveModels(testCtx, modelIDs)
			So(err, ShouldNotBeNil)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InternalError_GetObjectiveModelsByModelIDsFailed)
		})

		Convey("When some objective models not found", func() {
			modelIDs := []string{"test-id-1", "test-id-2"}
			models := []interfaces.ObjectiveModel{
				{
					ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
						ModelID: "test-id-1",
					},
				},
			}
			oma.EXPECT().GetObjectiveModelsByModelIDs(gomock.Any(), gomock.Any()).Return(models, nil)

			_, err := oms.GetObjectiveModels(testCtx, modelIDs)
			So(err, ShouldNotBeNil)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusNotFound)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_ObjectiveModelNotFound)
		})

		Convey("When get metric tasks fails", func() {
			modelIDs := []string{"test-id"}
			models := []interfaces.ObjectiveModel{
				{
					ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
						ModelID: "test-id",
					},
				},
			}
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			oma.EXPECT().GetObjectiveModelsByModelIDs(gomock.Any(), gomock.Any()).Return(models, nil)
			mmts.EXPECT().GetMetricTasksByModelIDs(gomock.Any(), gomock.Any()).Return(nil, errors.New("get metric tasks error"))

			_, err := oms.GetObjectiveModels(testCtx, modelIDs)
			So(err, ShouldNotBeNil)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InternalError_GetMetricTasksByModelIDsFailed)
		})

		Convey("When get simple index bases fails", func() {
			modelIDs := []string{"test-id"}
			models := []interfaces.ObjectiveModel{
				{
					ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
						ModelID: "test-id",
					},
					Task: &interfaces.MetricTask{
						IndexBase: "test-index",
					},
				},
			}
			task := interfaces.MetricTask{
				ModelID:   "test-id",
				IndexBase: "test-index",
			}

			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			oma.EXPECT().GetObjectiveModelsByModelIDs(gomock.Any(), gomock.Any()).Return(models, nil)
			mmts.EXPECT().GetMetricTasksByModelIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.MetricTask{"test-id": task}, nil)
			iba.EXPECT().GetSimpleIndexBasesByTypes(gomock.Any(), gomock.Any()).Return(nil, errors.New("get simple index bases error"))

			_, err := oms.GetObjectiveModels(testCtx, modelIDs)
			So(err, ShouldNotBeNil)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InternalError_GetSimpleIndexBasesByTypesFailed)
		})

		Convey("When GetMetricModelSimpleInfosByIDs fails for SLOObjective", func() {
			modelIDs := []string{"test-id"}
			objective := float64(99)
			period := int64(100)
			models := []interfaces.ObjectiveModel{
				{
					ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
						ModelID:       "test-id",
						ModelName:     "test-model",
						ObjectiveType: interfaces.SLO,
						ObjectiveConfig: interfaces.SLOObjective{
							Objective: &objective,
							Period:    &period,
							GoodMetricModel: &interfaces.BundleMetricModel{
								ID: "123",
							},
							TotalMetricModel: &interfaces.BundleMetricModel{
								ID: "456",
							},
						},
					},
					Task: &interfaces.MetricTask{
						IndexBase: "test-index",
					},
				},
			}
			task := interfaces.MetricTask{
				ModelID:   "test-id",
				IndexBase: "test-index",
			}

			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			oma.EXPECT().GetObjectiveModelsByModelIDs(gomock.Any(), gomock.Any()).Return(models, nil)
			mmts.EXPECT().GetMetricTasksByModelIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.MetricTask{"test-id": task}, nil)
			iba.EXPECT().GetSimpleIndexBasesByTypes(gomock.Any(), gomock.Any()).Return([]interfaces.SimpleIndexBase{{Name: "test-index"}}, nil)
			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(nil, errors.New("get metric model simple infos error"))

			_, err := oms.GetObjectiveModels(testCtx, modelIDs)
			So(err, ShouldNotBeNil)
		})

		Convey("When GetMetricModelSimpleInfosByIDs fails for KPIObjective", func() {
			modelIDs := []string{"test-id"}
			objective := float64(99)
			weight := int64(1)
			models := []interfaces.ObjectiveModel{
				{
					ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
						ModelID:       "test-id",
						ModelName:     "test-model",
						ObjectiveType: interfaces.KPI,
						ObjectiveConfig: interfaces.KPIObjective{
							Objective: &objective,
							Unit:      "ms",
							ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
								{
									ID:     "123",
									Weight: &weight,
								},
							},
						},
					},
					Task: &interfaces.MetricTask{
						IndexBase: "test-index",
					},
				},
			}
			task := interfaces.MetricTask{
				ModelID:   "test-id",
				IndexBase: "test-index",
			}

			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			oma.EXPECT().GetObjectiveModelsByModelIDs(gomock.Any(), gomock.Any()).Return(models, nil)
			mmts.EXPECT().GetMetricTasksByModelIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.MetricTask{"test-id": task}, nil)
			iba.EXPECT().GetSimpleIndexBasesByTypes(gomock.Any(), gomock.Any()).Return([]interfaces.SimpleIndexBase{{Name: "test-index"}}, nil)
			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(nil, errors.New("get metric model simple infos error"))

			_, err := oms.GetObjectiveModels(testCtx, modelIDs)
			So(err, ShouldNotBeNil)
		})

		Convey("When GetObjectiveModels succeeds with SLO", func() {
			modelIDs := []string{"test-id", "test-id2"}
			objective := float64(99)
			weight := int64(1)
			models := []interfaces.ObjectiveModel{
				{
					ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
						ModelID:       "test-id",
						ModelName:     "test-model",
						ObjectiveType: interfaces.SLO,
						ObjectiveConfig: interfaces.SLOObjective{
							Objective: &objective,
							GoodMetricModel: &interfaces.BundleMetricModel{
								ID: "123",
							},
							TotalMetricModel: &interfaces.BundleMetricModel{
								ID: "456",
							},
						},
					},
					Task: &interfaces.MetricTask{
						IndexBase: "test-index",
					},
				},
				{
					ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
						ModelID:       "test-id2",
						ModelName:     "test-model2",
						ObjectiveType: interfaces.KPI,
						ObjectiveConfig: interfaces.KPIObjective{
							Objective: &objective,
							Unit:      "ms",
							ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
								{
									ID:     "123",
									Weight: &weight,
								},
							},
						},
					},
					Task: &interfaces.MetricTask{
						IndexBase: "test-index",
					},
				},
			}
			task := interfaces.MetricTask{
				ModelID:   "test-id",
				IndexBase: "test-index",
			}
			resrctmp := map[string]interfaces.ResourceOps{
				"test-id": {
					ResourceID: "test-id",
				},
				"test-id2": {
					ResourceID: "test-id2",
				},
			}
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrctmp, nil)
			oma.EXPECT().GetObjectiveModelsByModelIDs(gomock.Any(), gomock.Any()).Return(models, nil)
			mmts.EXPECT().GetMetricTasksByModelIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.MetricTask{"test-id": task}, nil)
			iba.EXPECT().GetSimpleIndexBasesByTypes(gomock.Any(), gomock.Any()).AnyTimes().Return([]interfaces.SimpleIndexBase{{Name: "test-index"}}, nil)
			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).AnyTimes().Return(map[string]interfaces.SimpleMetricModel{
				"123": {ModelName: "good-metric"},
				"456": {ModelName: "total-metric"},
			}, nil)

			result, err := oms.GetObjectiveModels(testCtx, modelIDs)
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 2)
			So(result[0].ModelID, ShouldEqual, "test-id")
		})

	})
}

func Test_ObjectiveModelService_UpdateObjectiveModel(t *testing.T) {
	Convey("Test UpdateObjectiveModel", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
		oma := dmock.NewMockObjectiveModelAccess(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)
		mmts := dmock.NewMockMetricModelTaskService(mockCtrl)
		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		oms, smock := MockNewObjectiveModelService(appSetting, dmja, mms, oma, mmts, iba, ps)

		Convey("When GetMetricModelSimpleInfosByIDs fails for SLOObjective", func() {
			objective := float64(99)
			period := int64(100)
			model := interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelID:       "test-id",
					ModelName:     "test-model",
					ObjectiveType: interfaces.SLO,
					ObjectiveConfig: interfaces.SLOObjective{
						Objective: &objective,
						Period:    &period,
						GoodMetricModel: &interfaces.BundleMetricModel{
							ID: "123",
						},
						TotalMetricModel: &interfaces.BundleMetricModel{
							ID: "456",
						},
					},
				},
				Task: &interfaces.MetricTask{
					IndexBase: "test-index",
				},
			}
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(nil, errors.New("get metric model simple infos error"))

			err := oms.UpdateObjectiveModel(testCtx, nil, model)
			So(err, ShouldNotBeNil)
		})

		Convey("When GetMetricModelSimpleInfosByIDs fails for KPIObjective", func() {
			objective := float64(99)
			weight := int64(100)
			model := interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelID:       "test-id",
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "ms",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "123",
								Weight: &weight,
							},
						},
					},
				},
				Task: &interfaces.MetricTask{
					IndexBase: "test-index",
				},
			}
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(nil, errors.New("get metric model simple infos error"))

			err := oms.UpdateObjectiveModel(testCtx, nil, model)
			So(err, ShouldNotBeNil)
		})

		Convey("When GetSimpleIndexBasesByTypes fails", func() {
			objective := float64(99)
			weight := int64(100)
			model := interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelID:       "test-id",
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "ms",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "123",
								Weight: &weight,
							},
						},
					},
				},
				Task: &interfaces.MetricTask{
					IndexBase: "test-index",
				},
			}
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.SimpleMetricModel{
				"123": {ModelName: "test-metric-model"},
			}, nil)
			iba.EXPECT().GetSimpleIndexBasesByTypes(gomock.Any(), gomock.Any()).Return(nil, errors.New("get simple index bases error"))

			err := oms.UpdateObjectiveModel(testCtx, nil, model)
			So(err, ShouldNotBeNil)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InternalError_GetSimpleIndexBasesByTypesFailed)
		})

		Convey("When begin transaction fails", func() {
			objective := float64(99)
			weight := int64(100)
			model := interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelID:       "test-id",
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "ms",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "123",
								Weight: &weight,
							},
						},
					},
				},
				Task: &interfaces.MetricTask{
					IndexBase: "test-index",
				},
			}
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.SimpleMetricModel{
				"123": {ModelName: "test-metric-model"},
			}, nil)
			iba.EXPECT().GetSimpleIndexBasesByTypes(gomock.Any(), gomock.Any()).Return([]interfaces.SimpleIndexBase{
				{Name: "test-index"},
			}, nil)
			smock.ExpectBegin().WillReturnError(errors.New("begin transaction failed"))

			err := oms.UpdateObjectiveModel(testCtx, nil, model)
			So(err, ShouldNotBeNil)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InternalError_BeginTransactionFailed)
		})

		Convey("When update objective model fails", func() {
			objective := float64(99)
			weight := int64(100)
			model := interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelID:       "test-id",
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "ms",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "123",
								Weight: &weight,
							},
						},
					},
				},
				Task: &interfaces.MetricTask{
					IndexBase: "test-index",
				},
			}
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.SimpleMetricModel{
				"123": {ModelName: "test-metric-model"},
			}, nil)
			iba.EXPECT().GetSimpleIndexBasesByTypes(gomock.Any(), gomock.Any()).Return([]interfaces.SimpleIndexBase{
				{Name: "test-index"},
			}, nil)
			smock.ExpectBegin()
			oma.EXPECT().UpdateObjectiveModel(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("update objective model failed"))
			smock.ExpectRollback()

			err := oms.UpdateObjectiveModel(testCtx, nil, model)
			So(err, ShouldNotBeNil)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InternalError)
		})

		Convey("When update metric task fails", func() {
			objective := float64(99)
			weight := int64(100)
			model := interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelID:       "test-id",
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "ms",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "123",
								Weight: &weight,
							},
						},
					},
				},
				Task: &interfaces.MetricTask{
					IndexBase: "test-index",
				},
			}
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.SimpleMetricModel{
				"123": {ModelName: "test-metric-model"},
			}, nil)
			iba.EXPECT().GetSimpleIndexBasesByTypes(gomock.Any(), gomock.Any()).Return([]interfaces.SimpleIndexBase{
				{Name: "test-index"},
			}, nil)
			smock.ExpectBegin()
			oma.EXPECT().UpdateObjectiveModel(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mmts.EXPECT().UpdateMetricTask(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("update metric task failed"))
			smock.ExpectRollback()

			err := oms.UpdateObjectiveModel(testCtx, nil, model)
			So(err, ShouldNotBeNil)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InternalError)
		})

		Convey("When update objective model succeeds", func() {
			objective := float64(99)
			weight := int64(100)
			model := interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelID:       "test-id",
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "ms",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "123",
								Weight: &weight,
							},
						},
					},
				},
				Task: &interfaces.MetricTask{
					IndexBase: "test-index",
				},
			}
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.SimpleMetricModel{
				"123": {ModelName: "test-metric-model"},
			}, nil)
			iba.EXPECT().GetSimpleIndexBasesByTypes(gomock.Any(), gomock.Any()).Return([]interfaces.SimpleIndexBase{
				{Name: "test-index"},
			}, nil)
			smock.ExpectBegin()
			oma.EXPECT().UpdateObjectiveModel(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mmts.EXPECT().UpdateMetricTask(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()
			dmja.EXPECT().UpdateJob(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			err := oms.UpdateObjectiveModel(testCtx, nil, model)
			So(err, ShouldBeNil)
		})
	})
}

func Test_ObjectiveModelService_DeleteObjectiveModels(t *testing.T) {
	Convey("Test DeleteObjectiveModels", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
		oma := dmock.NewMockObjectiveModelAccess(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)
		mmts := dmock.NewMockMetricModelTaskService(mockCtrl)
		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		oms, smock := MockNewObjectiveModelService(appSetting, dmja, mms, oma, mmts, iba, ps)
		resrc := map[string]interfaces.ResourceOps{
			"test-id": {
				ResourceID: "test-id",
			},
		}
		Convey("When begin transaction fails", func() {
			modelIDs := []string{"test-id"}
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			mmts.EXPECT().GetMetricTaskIDsByModelIDs(gomock.Any(), gomock.Any()).Return(nil, nil)
			smock.ExpectBegin().WillReturnError(errors.New("begin transaction error"))

			_, err := oms.DeleteObjectiveModels(testCtx, modelIDs)
			So(err, ShouldNotBeNil)
		})

		Convey("When delete objective models fails", func() {
			modelIDs := []string{"test-id"}
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			mmts.EXPECT().GetMetricTaskIDsByModelIDs(gomock.Any(), gomock.Any()).Return(nil, nil)
			smock.ExpectBegin()
			oma.EXPECT().DeleteObjectiveModels(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), errors.New("delete objective models error"))
			smock.ExpectRollback()

			_, err := oms.DeleteObjectiveModels(testCtx, modelIDs)
			So(err, ShouldNotBeNil)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InternalError)
		})

		Convey("When GetMetricTaskIDsByModelIDs fails", func() {
			modelIDs := []string{"test-id"}

			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			smock.ExpectBegin()
			mmts.EXPECT().GetMetricTaskIDsByModelIDs(gomock.Any(), gomock.Any()).Return(nil, errors.New("get metric task ids error"))
			smock.ExpectRollback()

			_, err := oms.DeleteObjectiveModels(testCtx, modelIDs)
			So(err, ShouldNotBeNil)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InternalError)
		})

		Convey("When DeleteMetricTaskByTaskIDs fails", func() {
			modelIDs := []string{"test-id"}
			taskIDs := []string{"1"}

			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			smock.ExpectBegin()
			oma.EXPECT().DeleteObjectiveModels(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), nil)
			mmts.EXPECT().GetMetricTaskIDsByModelIDs(gomock.Any(), gomock.Any()).Return(taskIDs, nil)
			dmja.EXPECT().StopJobs(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			mmts.EXPECT().DeleteMetricTaskByTaskIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("set task sync status error"))
			smock.ExpectRollback()

			_, err := oms.DeleteObjectiveModels(testCtx, modelIDs)
			So(err, ShouldNotBeNil)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InternalError)
		})

		Convey("When delete objective models succeeds", func() {
			modelIDs := []string{"test-id"}
			taskIDs := []string{"1"}

			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			ps.EXPECT().DeleteResources(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin()
			oma.EXPECT().DeleteObjectiveModels(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			mmts.EXPECT().GetMetricTaskIDsByModelIDs(gomock.Any(), gomock.Any()).Return(taskIDs, nil)
			dmja.EXPECT().StopJobs(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			mmts.EXPECT().DeleteMetricTaskByTaskIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()

			count, err := oms.DeleteObjectiveModels(testCtx, modelIDs)
			So(err, ShouldBeNil)
			So(count, ShouldEqual, 1)
		})
	})
}
