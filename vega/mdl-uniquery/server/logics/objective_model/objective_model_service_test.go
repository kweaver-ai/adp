// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package objective_model

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	"uniquery/common"
	"uniquery/common/convert"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	mock "uniquery/interfaces/mock"
)

var (
	testCtx = context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage)
	start   = time.Now().UnixMilli()
	end     = time.Now().Add(time.Hour).UnixMilli()
)

var (
	metricModel1 = interfaces.MetricModel{
		ModelID:    "12",
		ModelName:  "model1",
		MetricType: "atomic",
		DataSource: &interfaces.MetricDataSource{
			Type: interfaces.SOURCE_TYPE_DATA_VIEW,
			ID:   "主机监控",
		},
		Unit: "%",
	}
	metricModel2 = interfaces.MetricModel{
		ModelID:    "13",
		ModelName:  "model2",
		MetricType: "atomic",
		DataSource: &interfaces.MetricDataSource{
			Type: interfaces.SOURCE_TYPE_DATA_VIEW,
			ID:   "主机监控",
		},
		Unit: "%",
	}
)

func MockNewObjectiveModelService(appSetting *common.AppSetting, mmAccess interfaces.MetricModelAccess,
	mmService interfaces.MetricModelService, omAccess interfaces.ObjectiveModelAccess,
	ps interfaces.PermissionService) *objectiveModelService {

	return &objectiveModelService{
		appSetting: appSetting,
		mmAccess:   mmAccess,
		mmService:  mmService,
		omAccess:   omAccess,
		ps:         ps,
	}
}

func TestSimulate(t *testing.T) {
	Convey("Test Simulate", t, func() {

		InitObjectivePool(common.PoolSetting{
			ObjectivePoolSize: 10,
		})

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				FullCacheRefreshInterval: 24 * time.Hour,
			}}
		mmaMock := mock.NewMockMetricModelAccess(mockCtrl)
		mmServiceMock := mock.NewMockMetricModelService(mockCtrl)
		omaMock := mock.NewMockObjectiveModelAccess(mockCtrl)
		psMock := mock.NewMockPermissionService(mockCtrl)
		ds := MockNewObjectiveModelService(appSetting, mmaMock, mmServiceMock, omaMock, psMock)

		ops := []interfaces.ResourceOps{
			{
				ResourceID: interfaces.RESOURCE_ID_ALL,
				Operations: []string{interfaces.OPERATION_TYPE_CREATE},
			},
		}
		Convey("Simulate failed, because validate failed", func() {
			psMock.EXPECT().GetResourcesOperations(gomock.Any(), gomock.Any(), gomock.Any()).Return(ops, nil)
			query := interfaces.ObjectiveModelQuery{}
			_, err := ds.Simulate(testCtx, query)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_ObjectiveModel_UnsupportQuery)
		})

		Convey("eval kpi success", func() {
			query := interfaces.ObjectiveModelQuery{
				ObjectiveType: interfaces.KPI,
				ObjectiveConfig: interfaces.KPIObjective{
					ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
						{
							Id: "14",
						},
					},
					AdditionalMetricModels: []interfaces.BundleMetricModel{
						{
							Id: "15",
						},
					},
				},
				QueryTimeParams: interfaces.QueryTimeParams{
					Start: &start,
					End:   &end,
				},
			}
			psMock.EXPECT().GetResourcesOperations(gomock.Any(), gomock.Any(), gomock.Any()).Return(ops, nil)
			mmaMock.EXPECT().GetMetricModels(gomock.Any(), gomock.Any()).
				Return([]interfaces.MetricModel{metricModel1, metricModel2}, nil)
			mmServiceMock.EXPECT().Exec(gomock.Any(), gomock.Any()).AnyTimes().Return(interfaces.MetricModelUniResponse{}, 0, 0, nil)

			_, err := ds.Simulate(testCtx, query)
			So(err, ShouldBeNil)
		})

	})
}

func TestExec(t *testing.T) {
	Convey("Test Exec", t, func() {

		InitObjectivePool(common.PoolSetting{
			ObjectivePoolSize: 10,
		})

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				FullCacheRefreshInterval: 24 * time.Hour,
			}}
		mmaMock := mock.NewMockMetricModelAccess(mockCtrl)
		mmServiceMock := mock.NewMockMetricModelService(mockCtrl)
		omaMock := mock.NewMockObjectiveModelAccess(mockCtrl)
		psMock := mock.NewMockPermissionService(mockCtrl)
		ds := MockNewObjectiveModelService(appSetting, mmaMock, mmServiceMock, omaMock, psMock)

		Convey("Exec failed, because omAccess.GetObjectiveModel error", func() {
			query := interfaces.ObjectiveModelQuery{
				ObjectiveModelID: "123",
				QueryTimeParams: interfaces.QueryTimeParams{
					Start: &start,
					End:   &end,
				},
			}

			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			omaMock.EXPECT().GetObjectiveModel(gomock.Any(), gomock.Any()).Return(interfaces.ObjectiveModel{}, false, fmt.Errorf("get objective model error"))

			_, err := ds.Exec(testCtx, query)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_ObjectiveModel_InternalError_GetModelByIdFailed)
		})

		Convey("Exec failed, because objective model not found", func() {
			query := interfaces.ObjectiveModelQuery{
				ObjectiveModelID: "123",
				QueryTimeParams: interfaces.QueryTimeParams{
					Start: &start,
					End:   &end,
				},
			}

			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			omaMock.EXPECT().GetObjectiveModel(gomock.Any(), gomock.Any()).Return(interfaces.ObjectiveModel{}, false, nil)

			_, err := ds.Exec(testCtx, query)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_ObjectiveModel_ObjectiveModelNotFound)
		})

		Convey("Exec failed, because mmService.Exec error", func() {
			query := interfaces.ObjectiveModelQuery{
				ObjectiveType: interfaces.SLO,
				ObjectiveConfig: interfaces.SLOObjective{
					GoodMetricModel: &interfaces.BundleMetricModel{
						Id: "12",
					},
					TotalMetricModel: &interfaces.BundleMetricModel{
						Id: "13",
					},
				},
				QueryTimeParams: interfaces.QueryTimeParams{
					Start: &start,
					End:   &end,
				},
			}
			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			omaMock.EXPECT().GetObjectiveModel(gomock.Any(), gomock.Any()).
				Return(interfaces.ObjectiveModel{ObjectiveType: "invalid type"}, true, nil)
			mmServiceMock.EXPECT().Exec(gomock.Any(), gomock.Any()).AnyTimes().
				Return(interfaces.MetricModelUniResponse{}, 0, 0, fmt.Errorf("exec error"))

			_, err := ds.Exec(testCtx, query)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_ObjectiveModel_UnsupportQuery)
		})

		Convey("success", func() {
			query := interfaces.ObjectiveModelQuery{
				ObjectiveType: interfaces.KPI,
				ObjectiveConfig: interfaces.KPIObjective{
					ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
						{
							Id: "14",
						},
					},
					AdditionalMetricModels: []interfaces.BundleMetricModel{
						{
							Id: "15",
						},
					},
				},
				QueryTimeParams: interfaces.QueryTimeParams{
					Start: &start,
					End:   &end,
				},
			}
			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			omaMock.EXPECT().GetObjectiveModel(gomock.Any(), gomock.Any()).
				Return(interfaces.ObjectiveModel{
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								Id: "14",
							},
						},
						AdditionalMetricModels: []interfaces.BundleMetricModel{
							{
								Id: "15",
							},
						},
					},
				}, true, nil)
			mmaMock.EXPECT().GetMetricModels(gomock.Any(), gomock.Any()).
				Return([]interfaces.MetricModel{metricModel1, metricModel2}, nil)
			mmServiceMock.EXPECT().Exec(gomock.Any(), gomock.Any()).AnyTimes().
				Return(interfaces.MetricModelUniResponse{}, 0, 0, nil)

			_, err := ds.Exec(testCtx, query)
			So(err, ShouldBeNil)
		})

	})
}

func TestEval(t *testing.T) {
	Convey("Test eval", t, func() {

		InitObjectivePool(common.PoolSetting{
			ObjectivePoolSize: 10,
		})

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				FullCacheRefreshInterval: 24 * time.Hour,
			}}
		mmaMock := mock.NewMockMetricModelAccess(mockCtrl)
		mmServiceMock := mock.NewMockMetricModelService(mockCtrl)
		omaMock := mock.NewMockObjectiveModelAccess(mockCtrl)
		psMock := mock.NewMockPermissionService(mockCtrl)
		ds := MockNewObjectiveModelService(appSetting, mmaMock, mmServiceMock, omaMock, psMock)

		Convey("eval failed, because validate failed", func() {
			query := interfaces.ObjectiveModelQuery{
				ObjectiveType: "invalid type",
			}
			_, err := ds.eval(testCtx, query)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_ObjectiveModel_UnsupportQuery)
		})

		Convey("eval slo failed, because metric models have different units", func() {
			query := interfaces.ObjectiveModelQuery{
				ObjectiveType: interfaces.SLO,
				ObjectiveConfig: interfaces.SLOObjective{
					GoodMetricModel: &interfaces.BundleMetricModel{
						Id: "12",
					},
					TotalMetricModel: &interfaces.BundleMetricModel{
						Id: "13",
					},
				},
				QueryTimeParams: interfaces.QueryTimeParams{
					Start: &start,
					End:   &end,
				},
			}

			metricModel1Different := interfaces.MetricModel{
				ModelID:    "12",
				ModelName:  "model1",
				MetricType: "atomic",
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "主机监控",
				},
				Unit: "ms",
			}
			metricModel2Different := interfaces.MetricModel{
				ModelID:    "13",
				ModelName:  "model2",
				MetricType: "atomic",
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "主机监控",
				},
				Unit: "%",
			}

			mmaMock.EXPECT().GetMetricModels(gomock.Any(), gomock.Any()).
				Return([]interfaces.MetricModel{metricModel1Different, metricModel2Different}, nil)

			_, err := ds.eval(testCtx, query)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_ObjectiveModel_AssociateMetricModelsUnit_Different)
		})
		Convey("eval failed, because mmService.Exec error", func() {
			query := interfaces.ObjectiveModelQuery{
				ObjectiveType: interfaces.SLO,
				ObjectiveConfig: interfaces.SLOObjective{
					GoodMetricModel: &interfaces.BundleMetricModel{
						Id: "12",
					},
					TotalMetricModel: &interfaces.BundleMetricModel{
						Id: "13",
					},
				},
				QueryTimeParams: interfaces.QueryTimeParams{
					Start: &start,
					End:   &end,
				},
			}
			mmaMock.EXPECT().GetMetricModels(gomock.Any(), gomock.Any()).
				Return([]interfaces.MetricModel{metricModel1, metricModel2}, nil)
			mmServiceMock.EXPECT().Exec(gomock.Any(), gomock.Any()).AnyTimes().
				Return(interfaces.MetricModelUniResponse{}, 0, 0, fmt.Errorf("exec error"))

			_, err := ds.eval(testCtx, query)
			So(err.Error(), ShouldEqual, "exec error")
		})

		Convey("eval slo success", func() {
			objective := 99.9
			period := int64(90)
			query := interfaces.ObjectiveModelQuery{
				ObjectiveType: interfaces.SLO,
				ObjectiveConfig: interfaces.SLOObjective{
					Objective: &objective,
					Period:    &period,
					GoodMetricModel: &interfaces.BundleMetricModel{
						Id: "12",
					},
					TotalMetricModel: &interfaces.BundleMetricModel{
						Id: "13",
					},
				},
				QueryTimeParams: interfaces.QueryTimeParams{
					Start: &start,
					End:   &end,
				},
			}

			mmaMock.EXPECT().GetMetricModels(gomock.Any(), gomock.Any()).
				Return([]interfaces.MetricModel{metricModel1, metricModel2}, nil)
			mmServiceMock.EXPECT().Exec(gomock.Any(), gomock.Any()).AnyTimes().Return(interfaces.MetricModelUniResponse{}, 0, 0, nil)

			_, err := ds.eval(testCtx, query)
			So(err, ShouldBeNil)
		})

		Convey("eval kpi failed with different units", func() {
			query := interfaces.ObjectiveModelQuery{
				ObjectiveType: interfaces.KPI,
				ObjectiveConfig: interfaces.KPIObjective{
					ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
						{
							Id: "14",
						},
					},
					AdditionalMetricModels: []interfaces.BundleMetricModel{
						{
							Id: "15",
						},
					},
				},
				QueryTimeParams: interfaces.QueryTimeParams{
					Start: &start,
					End:   &end,
				},
			}

			metricModel1 := interfaces.MetricModel{
				ModelID:    "14",
				ModelName:  "model1",
				MetricType: "atomic",
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "主机监控",
				},
				Unit: "%",
			}
			metricModel2 := interfaces.MetricModel{
				ModelID:    "15",
				ModelName:  "model2",
				MetricType: "atomic",
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "主机监控",
				},
				Unit: "ms",
			}

			mmaMock.EXPECT().GetMetricModels(gomock.Any(), gomock.Any()).
				Return([]interfaces.MetricModel{metricModel1, metricModel2}, nil)

			_, err := ds.eval(testCtx, query)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_ObjectiveModel_AssociateMetricModelsUnit_Different)
		})

		Convey("eval kpi success", func() {
			query := interfaces.ObjectiveModelQuery{
				ObjectiveType: interfaces.KPI,
				ObjectiveConfig: interfaces.KPIObjective{
					ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
						{
							Id: "14",
						},
					},
					AdditionalMetricModels: []interfaces.BundleMetricModel{
						{
							Id: "15",
						},
					},
				},
				QueryTimeParams: interfaces.QueryTimeParams{
					Start: &start,
					End:   &end,
				},
			}
			mmaMock.EXPECT().GetMetricModels(gomock.Any(), gomock.Any()).
				Return([]interfaces.MetricModel{metricModel1, metricModel2}, nil)
			mmServiceMock.EXPECT().Exec(gomock.Any(), gomock.Any()).AnyTimes().Return(interfaces.MetricModelUniResponse{}, 0, 0, nil)

			_, err := ds.eval(testCtx, query)
			So(err, ShouldBeNil)
		})
	})
}

func TestProcessSLOObjectiveResponse(t *testing.T) {
	Convey("Test ProcessSLOObjectiveResponse", t, func() {

		Convey("success", func() {
			results := map[string]map[string]interfaces.MetricModelData{
				"12": {
					"series0": {
						Labels: map[string]string{"label1": "value0"},
						Times:  []interface{}{1000, 2000, 3000, 4000},
						Values: []interface{}{100, nil, 300, convert.POS_INF},
					},
					"series1": {
						Labels: map[string]string{"label1": "value1"},
						Times:  []interface{}{1000, 2000, 3000, 4000},
						Values: []interface{}{1000, 2000, 3000, convert.NaN},
					},
				},
				"13": {
					"series1": {
						Labels: map[string]string{"label1": "value1"},
						Times:  []interface{}{1000, 2000, 3000, 4000},
						Values: []interface{}{1000, nil, convert.NEG_INF, 3000},
					},
				},
			}

			objective := 0.99
			period := int64(90)
			from1, to1 := 0.0, 50.0
			from2, to2 := 80.0, 100.0
			from3, to3 := 50.0, 80.0
			sloObjective := interfaces.SLOObjective{
				Objective: &objective,
				Period:    &period,
				GoodMetricModel: &interfaces.BundleMetricModel{
					Id: "12",
				},
				TotalMetricModel: &interfaces.BundleMetricModel{
					Id: "13",
				},
				StatusConfig: &interfaces.ObjectiveStatusConfig{
					Ranges: []interfaces.Range{
						{
							From:   &from1,
							To:     &to1,
							Status: "s1",
						},
						{
							From:   &from2,
							To:     &to2,
							Status: "s2",
						},
						{
							From:   &from3,
							To:     &to3,
							Status: "s3",
						},
					},
					OtherStatus: "other_status",
				},
			}

			resp, err := processSLOObjectiveResponse(testCtx, results, sloObjective)
			So(err, ShouldBeNil)
			So(resp.Datas, ShouldNotBeEmpty)

			data := resp.Datas[0].(interfaces.SLOObjectiveData)
			So(data.Labels, ShouldResemble, map[string]string{"label1": "value1"})
			So(data.Times, ShouldResemble, []interface{}{1000, 2000, 3000, 4000})
		})

		Convey("When metric time points are not equal", func() {
			times1 := []int64{1000, 2000, 3000}
			interfaceTimes1 := make([]interface{}, len(times1))
			for i, t := range times1 {
				interfaceTimes1[i] = t
			}

			times2 := []int64{1000, 2000}
			interfaceTimes2 := make([]interface{}, len(times2))
			for i, t := range times2 {
				interfaceTimes2[i] = t
			}

			v1 := []float64{100, 200, 300}
			interfacev1 := make([]interface{}, len(v1))
			for i, v := range v1 {
				interfacev1[i] = v
			}

			v2 := []float64{1000, 2000}
			interfacev2 := make([]interface{}, len(v2))
			for i, v := range v2 {
				interfacev2[i] = v
			}

			results := map[string]map[string]interfaces.MetricModelData{
				"12": {
					"series1": {
						Labels: map[string]string{"label1": "value1"},
						Times:  interfaceTimes1,
						Values: interfacev1,
					},
				},
				"13": {
					"series1": {
						Labels: map[string]string{"label1": "value1"},
						Times:  interfaceTimes2,
						Values: interfacev2,
					},
				},
			}

			objective := 0.99
			period := int64(90)
			sloObjective := interfaces.SLOObjective{
				Objective: &objective,
				Period:    &period,
				GoodMetricModel: &interfaces.BundleMetricModel{
					Id: "12",
				},
				TotalMetricModel: &interfaces.BundleMetricModel{
					Id: "13",
				},
			}

			_, err := processSLOObjectiveResponse(testCtx, results, sloObjective)
			So(err, ShouldNotBeNil)
		})

		Convey("AssertFloat64 error", func() {
			times1 := []interface{}{1000, 2000, 3000}
			results := map[string]map[string]interfaces.MetricModelData{
				"12": {
					"series1": {
						Labels: map[string]string{"label1": "value1"},
						Times:  times1,
						Values: []interface{}{convert.NEG_INF, "not a number", 300},
					},
				},
				"13": {
					"series1": {
						Labels: map[string]string{"label1": "value1"},
						Times:  times1,
						Values: []interface{}{1000, 2000, 3000},
					},
				},
			}

			objective := 0.99
			period := int64(90)
			sloObjective := interfaces.SLOObjective{
				Objective: &objective,
				Period:    &period,
				GoodMetricModel: &interfaces.BundleMetricModel{
					Id: "12",
				},
				TotalMetricModel: &interfaces.BundleMetricModel{
					Id: "13",
				},
			}

			_, err := processSLOObjectiveResponse(testCtx, results, sloObjective)
			So(err, ShouldNotBeNil)
		})

		Convey("convert.AssertFloat64 error for totals", func() {
			times1 := []interface{}{1000, 2000, 3000}
			results := map[string]map[string]interfaces.MetricModelData{
				"12": {
					"series1": {
						Labels: map[string]string{"label1": "value1"},
						Times:  times1,
						Values: []interface{}{100, 200, 300},
					},
				},
				"13": {
					"series1": {
						Labels: map[string]string{"label1": "value1"},
						Times:  times1,
						Values: []interface{}{nil, convert.POS_INF, "not a number"},
					},
				},
			}

			objective := 0.99
			period := int64(90)
			sloObjective := interfaces.SLOObjective{
				Objective: &objective,
				Period:    &period,
				GoodMetricModel: &interfaces.BundleMetricModel{
					Id: "12",
				},
				TotalMetricModel: &interfaces.BundleMetricModel{
					Id: "13",
				},
			}

			_, err := processSLOObjectiveResponse(testCtx, results, sloObjective)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestProcessKPIObjectiveResponse(t *testing.T) {
	Convey("Test ProcessKPIObjectiveResponse", t, func() {

		Convey("success", func() {
			times := []interface{}{1000, 2000, 3000}
			results := map[string]map[string]interfaces.MetricModelData{
				"14": {
					"series1": {
						Labels: map[string]string{"label1": "value1"},
						Times:  times,
						Values: []interface{}{100, nil, 300, convert.POS_INF},
					},
				},
				"15": {
					"series1": {
						Labels: map[string]string{"label1": "value1"},
						Times:  times,
						Values: []interface{}{1000, 2000, 3000, convert.NaN},
					},
				},
				"16": {
					"series1": {
						Labels: map[string]string{"label1": "value1"},
						Times:  times,
						Values: []interface{}{1000, 2000, 3000, convert.NaN},
					},
				},
			}

			objective := 0.99
			weight := int64(100)
			to1 := 50.0
			from2, to2 := 80.0, 100.0
			from3 := 50.0
			kpiObjective := interfaces.KPIObjective{
				Objective: &objective,
				ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
					{
						Id:     "14",
						Weight: &weight,
					}, {
						Id:     "16",
						Weight: &weight,
					},
				},
				AdditionalMetricModels: []interfaces.BundleMetricModel{
					{
						Id: "15",
					},
				},
				StatusConfig: &interfaces.ObjectiveStatusConfig{
					Ranges: []interfaces.Range{
						{
							To:     &to1,
							Status: "s1",
						},
						{
							From:   &from2,
							To:     &to2,
							Status: "s2",
						},
						{
							From:   &from3,
							Status: "s3",
						},
					},
					OtherStatus: "other_status",
				},
			}

			resp, err := processKPIObjectiveResponse(testCtx, results, kpiObjective)
			So(err, ShouldBeNil)
			So(resp.Datas, ShouldNotBeEmpty)

			data := resp.Datas[0].(interfaces.KPIObjectiveData)
			So(data.Labels, ShouldResemble, map[string]string{"label1": "value1"})
		})

		Convey("When metric time points are not equal", func() {
			times1 := []int64{1000, 2000, 3000}
			interfaceTimes1 := make([]interface{}, len(times1))
			for i, t := range times1 {
				interfaceTimes1[i] = t
			}

			times2 := []int64{1000, 2000}
			interfaceTimes2 := make([]interface{}, len(times2))
			for i, t := range times2 {
				interfaceTimes2[i] = t
			}

			v1 := []float64{100, 200, 300}
			interfacev1 := make([]interface{}, len(v1))
			for i, v := range v1 {
				interfacev1[i] = v
			}

			v2 := []float64{1000, 2000}
			interfacev2 := make([]interface{}, len(v2))
			for i, v := range v2 {
				interfacev2[i] = v
			}

			results := map[string]map[string]interfaces.MetricModelData{
				"14": {
					"series1": {
						Labels: map[string]string{"label1": "value1"},
						Times:  interfaceTimes1,
						Values: interfacev1,
					},
				},
				"15": {
					"series1": {
						Labels: map[string]string{"label1": "value1"},
						Times:  interfaceTimes2,
						Values: interfacev2,
					},
				},
			}

			kpiObjective := interfaces.KPIObjective{
				ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
					{
						Id: "14",
					},
					{
						Id: "15",
					},
				},
			}

			_, err := processKPIObjectiveResponse(testCtx, results, kpiObjective)
			So(err, ShouldNotBeNil)
		})

		Convey("When additional metric time points are not equal", func() {
			times1 := []int64{1000, 2000, 3000}
			interfaceTimes1 := make([]interface{}, len(times1))
			for i, t := range times1 {
				interfaceTimes1[i] = t
			}

			times2 := []int64{1000, 2000}
			interfaceTimes2 := make([]interface{}, len(times2))
			for i, t := range times2 {
				interfaceTimes2[i] = t
			}

			v1 := []float64{100, 200, 300}
			interfacev1 := make([]interface{}, len(v1))
			for i, v := range v1 {
				interfacev1[i] = v
			}

			v2 := []float64{1000, 2000}
			interfacev2 := make([]interface{}, len(v2))
			for i, v := range v2 {
				interfacev2[i] = v
			}

			results := map[string]map[string]interfaces.MetricModelData{
				"14": {
					"series1": {
						Labels: map[string]string{"label1": "value1"},
						Times:  interfaceTimes1,
						Values: interfacev1,
					},
				},
				"15": {
					"series1": {
						Labels: map[string]string{"label1": "value1"},
						Times:  interfaceTimes2,
						Values: interfacev2,
					},
				},
			}

			kpiObjective := interfaces.KPIObjective{
				ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
					{
						Id: "14",
					},
				},
				AdditionalMetricModels: []interfaces.BundleMetricModel{
					{
						Id: "15",
					},
				},
			}

			_, err := processKPIObjectiveResponse(testCtx, results, kpiObjective)
			So(err, ShouldNotBeNil)
		})

		Convey("convert.AssertFloat64 error", func() {
			times := []interface{}{1000, 2000, 3000}
			results := map[string]map[string]interfaces.MetricModelData{
				"14": {
					"series1": {
						Labels: map[string]string{"label1": "value1"},
						Times:  times,
						Values: []interface{}{100, nil, "NOT A NUMBER", convert.POS_INF},
					},
				},
				"15": {
					"series1": {
						Labels: map[string]string{"label1": "value1"},
						Times:  times,
						Values: []interface{}{1000, 2000, 3000, convert.NaN},
					},
				},
				"16": {
					"series1": {
						Labels: map[string]string{"label1": "value1"},
						Times:  times,
						Values: []interface{}{1000, 2000, 3000, convert.NaN},
					},
				},
			}

			objective := 0.99
			weight := int64(100)
			kpiObjective := interfaces.KPIObjective{
				Objective: &objective,
				ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
					{
						Id:     "14",
						Weight: &weight,
					}, {
						Id:     "16",
						Weight: &weight,
					},
				},
				AdditionalMetricModels: []interfaces.BundleMetricModel{
					{
						Id: "15",
					},
				},
			}

			_, err := processKPIObjectiveResponse(testCtx, results, kpiObjective)
			So(err, ShouldNotBeNil)
		})

		Convey("convert.AssertFloat64 error for additional metrics", func() {
			times := []interface{}{1000, 2000, 3000}
			results := map[string]map[string]interfaces.MetricModelData{
				"14": {
					"series1": {
						Labels: map[string]string{"label1": "value1"},
						Times:  times,
						Values: []interface{}{100, 200, 300},
					},
				},
				"15": {
					"series1": {
						Labels: map[string]string{"label1": "value1"},
						Times:  times,
						Values: []interface{}{nil, convert.POS_INF, "not a number"},
					},
				},
				"16": {
					"series1": {
						Labels: map[string]string{"label1": "value1"},
						Times:  times,
						Values: []interface{}{100, 200, 300},
					},
				},
			}

			objective := 0.99
			weight := int64(100)
			kpiObjective := interfaces.KPIObjective{
				Objective: &objective,
				ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
					{
						Id:     "14",
						Weight: &weight,
					}, {
						Id:     "16",
						Weight: &weight,
					},
				},
				AdditionalMetricModels: []interfaces.BundleMetricModel{
					{
						Id: "15",
					},
				},
			}

			_, err := processKPIObjectiveResponse(testCtx, results, kpiObjective)
			So(err, ShouldNotBeNil)
		})

		Convey("convert.AssertFloat64 error for comprehensive metrics", func() {
			times := []interface{}{1000, 2000, 3000}
			results := map[string]map[string]interfaces.MetricModelData{
				"14": {
					"series1": {
						Labels: map[string]string{"label1": "value1"},
						Times:  times,
						Values: []interface{}{100, convert.POS_INF, 200},
					},
				},
				"15": {
					"series1": {
						Labels: map[string]string{"label1": "value1"},
						Times:  times,
						Values: []interface{}{100, 200, 300},
					},
				},
				"16": {
					"series1": {
						Labels: map[string]string{"label1": "value1"},
						Times:  times,
						Values: []interface{}{"not a number", 100, 300},
					},
				},
			}

			objective := 0.99
			weight := int64(100)
			kpiObjective := interfaces.KPIObjective{
				Objective: &objective,
				ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
					{
						Id:     "14",
						Weight: &weight,
					}, {
						Id:     "16",
						Weight: &weight,
					},
				},
				AdditionalMetricModels: []interfaces.BundleMetricModel{
					{
						Id: "15",
					},
				},
			}

			_, err := processKPIObjectiveResponse(testCtx, results, kpiObjective)
			So(err, ShouldNotBeNil)
		})
	})
}
