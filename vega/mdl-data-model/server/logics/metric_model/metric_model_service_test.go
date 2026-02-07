// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package metric_model

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/did"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	"data-model/common"
	derrors "data-model/errors"
	"data-model/interfaces"
	dtype "data-model/interfaces/data_type"
	dmock "data-model/interfaces/mock"
)

var (
	testCtx              = context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage)
	testFormula          = "avg(rate(node_cpu_seconds_total{name=~\"node-14-35\",mode12=\"system\"}[5m]))by(name)"
	testMetricUpdateTime = int64(1735786555379)
	testMetricGroup      = interfaces.MetricModelGroup{
		GroupID:   "1",
		GroupName: "111",
	}
)

func MockNewMetricModelService(appSetting *common.AppSetting,
	dmja interfaces.DataModelJobAccess,
	dvs interfaces.DataViewService,
	mma interfaces.MetricModelAccess,
	mmga interfaces.MetricModelGroupAccess,
	ua interfaces.UniqueryAccess,
	mmts interfaces.MetricModelTaskService,
	iba interfaces.IndexBaseAccess,
	ps interfaces.PermissionService) (*metricModelService, sqlmock.Sqlmock) {
	db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

	return &metricModelService{
		appSetting: appSetting,
		dvs:        dvs,
		dmja:       dmja,
		mma:        mma,
		mmga:       mmga,
		ua:         ua,
		mmts:       mmts,
		iba:        iba,
		db:         db,
		ps:         ps,
	}, smock
}

func Test_MetricModelService_CheckMetricModelExistByName(t *testing.T) {
	Convey("Test CheckMetricModelExistByName", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		mma := dmock.NewMockMetricModelAccess(mockCtrl)
		mmga := dmock.NewMockMetricModelGroupAccess(mockCtrl)
		ua := dmock.NewMockUniqueryAccess(mockCtrl)
		mmts := dmock.NewMockMetricModelTaskService(mockCtrl)
		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		mms, _ := MockNewMetricModelService(appSetting, dmja, dvs, mma, mmga, ua, mmts, iba, ps)

		Convey("Check failed, and return 500 error", func() {
			err := errors.New("Check Metric Model Exist By Name failed")
			mma.EXPECT().GetMetricModelIDByName(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, err)

			_, exists, httpErr := mms.CheckMetricModelExistByName(testCtx, "", "")
			So(exists, ShouldBeFalse)
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InternalError_CheckModelIfExistFailed)
		})

		Convey("Model combination name exists", func() {
			mma.EXPECT().GetMetricModelIDByName(gomock.Any(), gomock.Any(), gomock.Any()).Return("1", true, nil)

			_, exists, httpErr := mms.CheckMetricModelExistByName(testCtx, "", "")
			So(exists, ShouldBeTrue)
			So(httpErr, ShouldBeNil)
		})

		Convey("Check success and model name not exists", func() {
			mma.EXPECT().GetMetricModelIDByName(gomock.Any(), gomock.Any(), gomock.Any()).Return("1", false, nil)

			_, exists, httpErr := mms.CheckMetricModelExistByName(testCtx, "", "")
			So(exists, ShouldBeFalse)
			So(httpErr, ShouldBeNil)
		})
	})
}

// func Test_MetricModelService_CreateMetricModel(t *testing.T) {
// 	Convey("Test CreateMetricModel", t, func() {
// 		mockCtrl := gomock.NewController(t)
// 		defer mockCtrl.Finish()

// 		appSetting := &common.AppSetting{}
// 		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
// 		dvs := dmock.NewMockDataViewService(mockCtrl)
// 		mma := dmock.NewMockMetricModelAccess(mockCtrl)
// 		mmga := dmock.NewMockMetricModelGroupAccess(mockCtrl)
// 		ua := dmock.NewMockUniqueryAccess(mockCtrl)
// 		mmts := dmock.NewMockMetricModelTaskService(mockCtrl)
// 		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
// 		vegaService := dmock.NewMockVegaViewService(mockCtrl)
// 		ps := dmock.NewMockPermissionService(mockCtrl)
// 		mms, smock := MockNewMetricModelService(appSetting, dmja, dvs, mma, mmga, ua, mmts, iba, vegaService, ps)

// 		metricModel := interfaces.MetricModel{
// 			SimpleMetricModel: interfaces.SimpleMetricModel{
// 				ModelName:  "16",
// 				MetricType: "atomic",
// 				QueryType:  "promql",
// 				Tags:       []string{"a", "s", "s", "s", "s"},
// 				Comment:    "ssss",
// 				GroupName:  "sss",
// 				Formula:    testFormula,
// 			},
// 			DataSource: &interfaces.MetricDataSource{
// 				Type: interfaces.SOURCE_TYPE_DATA_VIEW,
// 				ID:   "数据视图1",
// 			},
// 			Task: &task,
// 		}

// 		viewName := "数据视图1"

// 		Convey("GetMetricModelGroupByName failed", func() {
// 			err := errors.New("GetMetricModelGroupByName failed")
// 			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			dvs.EXPECT().CheckDataViewExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(viewName, true, nil)
// 			ua.EXPECT().CheckFormulaByUniquery(gomock.Any(), gomock.Any()).
// 				Return(true, "", nil)
// 			smock.ExpectBegin()
// 			mmga.EXPECT().GetMetricModelGroupByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(testMetricGroup, false, err)
// 			smock.ExpectCommit()
// 			iba.EXPECT().GetSimpleIndexBasesByTypes(gomock.Any(), gomock.Any()).Return([]interfaces.SimpleIndexBase{{BaseType: "1", Name: "1"}}, nil)
// 			dmja.EXPECT().StartJob(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

// 			modelID, httpErr := mms.CreateMetricModel(testCtx, metricModel)
// 			So(modelID, ShouldEqual, "")
// 			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
// 			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModelGroup_InternalError_GetMetricModelGroupIDByNameFailed)
// 		})

// 		Convey("CreateMetricModel failed", func() {
// 			err := errors.New("Create Metric Model failed")
// 			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			dvs.EXPECT().CheckDataViewExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(viewName, true, nil)
// 			ua.EXPECT().CheckFormulaByUniquery(gomock.Any(), gomock.Any()).
// 				Return(true, "", nil)

// 			mmga.EXPECT().CreateMetricModelGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			mmga.EXPECT().GetMetricModelGroupByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(testMetricGroup, false, nil)
// 			iba.EXPECT().GetSimpleIndexBasesByTypes(gomock.Any(), gomock.Any()).Return([]interfaces.SimpleIndexBase{{BaseType: "1", Name: "1"}}, nil)

// 			smock.ExpectBegin()
// 			mma.EXPECT().CreateMetricModel(gomock.Any(), gomock.Any(), gomock.Any()).Return(err)
// 			smock.ExpectCommit()
// 			dmja.EXPECT().StartJob(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

// 			modelID, httpErr := mms.CreateMetricModel(testCtx, metricModel)
// 			So(modelID, ShouldEqual, "")
// 			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
// 			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InternalError)
// 		})

// 		Convey("CreateMetricModel success", func() {
// 			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			dvs.EXPECT().CheckDataViewExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(viewName, true, nil)
// 			ua.EXPECT().CheckFormulaByUniquery(gomock.Any(), gomock.Any()).
// 				Return(true, "", nil)

// 			mmga.EXPECT().GetMetricModelGroupByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(testMetricGroup, true, nil)
// 			iba.EXPECT().GetSimpleIndexBasesByTypes(gomock.Any(), gomock.Any()).Return([]interfaces.SimpleIndexBase{{BaseType: "1", Name: "1"}}, nil)

// 			smock.ExpectBegin()
// 			mma.EXPECT().CreateMetricModel(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			mmts.EXPECT().CreateMetricTask(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			smock.ExpectCommit()
// 			dmja.EXPECT().StartJob(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

// 			_, httpErr := mms.CreateMetricModel(testCtx, metricModel)
// 			So(httpErr, ShouldBeNil)
// 		})

// 		Convey("Data view not exists \n", func() {
// 			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			dvs.EXPECT().CheckDataViewExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(viewName, false, nil)

// 			modelID, httpErr := mms.CreateMetricModel(testCtx, metricModel)
// 			So(modelID, ShouldEqual, "")
// 			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
// 			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_DataView_DataViewNotFound)
// 		})

// 		Convey("Get Data view error \n", func() {
// 			err := &rest.HTTPError{
// 				HTTPCode: http.StatusInternalServerError,
// 				Language: rest.DefaultLanguage,
// 				BaseError: rest.BaseError{
// 					ErrorCode: derrors.DataModel_MetricModel_InternalError_GetDataViewByNameFailed,
// 				},
// 			}
// 			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			dvs.EXPECT().CheckDataViewExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(viewName, false, err)

// 			modelID, httpErr := mms.CreateMetricModel(testCtx, metricModel)
// 			So(modelID, ShouldEqual, "")
// 			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
// 			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_DataView_DataViewNotFound)
// 		})

// 		Convey("Get vega logic view error \n", func() {
// 			model := interfaces.MetricModel{
// 				SimpleMetricModel: interfaces.SimpleMetricModel{
// 					ModelName:  "16",
// 					MetricType: "atomic",
// 					QueryType:  "promql",
// 					Tags:       []string{"a", "s", "s", "s", "s"},
// 					Comment:    "ssss",
// 					GroupName:  "sss",
// 					Formula:    testFormula,
// 				},
// 				DataSource: &interfaces.MetricDataSource{
// 					Type: interfaces.DATA_SOURCE_VEGA_LOGIC_VIEW,
// 					ID:   "视图1",
// 				},
// 			}

// 			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			vegaService.EXPECT().GetVegaViewFieldsByID(gomock.Any(), gomock.Any()).Return(interfaces.VegaViewWithFields{}, fmt.Errorf("error"))

// 			modelID, httpErr := mms.CreateMetricModel(testCtx, model)
// 			So(modelID, ShouldEqual, "")
// 			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
// 			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InternalError_GetVegaViewFieldsByIDFailed)
// 		})

// 		Convey("CheckSqlFormulaConfig error \n", func() {
// 			model := interfaces.MetricModel{
// 				SimpleMetricModel: interfaces.SimpleMetricModel{
// 					ModelName:  "16",
// 					MetricType: "atomic",
// 					QueryType:  "sql",
// 					Tags:       []string{"a", "s", "s", "s", "s"},
// 					Comment:    "ssss",
// 					GroupName:  "sss",
// 					FormulaConfig: interfaces.SQLConfig{
// 						AggrExpr: &interfaces.AggrExpr{
// 							Aggr:  "count",
// 							Field: "f1",
// 						},
// 					},
// 					DateField: "f2",
// 				},
// 				DataSource: &interfaces.MetricDataSource{
// 					Type: interfaces.DATA_SOURCE_VEGA_LOGIC_VIEW,
// 					ID:   "视图1",
// 				},
// 			}
// 			vegaViewFields := interfaces.VegaViewWithFields{
// 				Fields: []interfaces.VegaViewField{
// 					{
// 						Name: "f1",
// 						Type: "varchar",
// 					},
// 					{
// 						Name: "f2",
// 						Type: "varchar",
// 					},
// 				},
// 				FieldMap: map[string]interfaces.VegaViewField{
// 					"f1": {
// 						Name: "f1",
// 						Type: "varchar",
// 					},
// 					"f2": {
// 						Name: "f2",
// 						Type: "varchar",
// 					},
// 				},
// 			}

// 			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			vegaService.EXPECT().GetVegaViewFieldsByID(gomock.Any(), gomock.Any()).Return(vegaViewFields, nil)
// 			ua.EXPECT().CheckFormulaByUniquery(gomock.Any(), gomock.Any()).
// 				Return(false, "invalid", nil)

// 			modelID, httpErr := mms.CreateMetricModel(testCtx, model)
// 			So(modelID, ShouldEqual, "")
// 			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
// 			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_Formula)
// 		})

// 		Convey("DSL measure field type si not a number \n", func() {
// 			dataViewFilters := &interfaces.DataViewQueryFilters{
// 				Name: "数据视图1",
// 				FieldsMap: map[string]*interfaces.ViewField{
// 					"metrics.node_cpu_seconds_total": {
// 						Name: "metrics.node_cpu_seconds_total",
// 						Type: dtype.DataType_Keyword,
// 					},
// 				},
// 			}
// 			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			dvs.EXPECT().CheckDataViewExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(viewName, true, nil)
// 			dvs.EXPECT().GetDataView(gomock.Any(), gomock.Any()).
// 				Return(dataViewFilters, nil)

// 			modelID, httpErr := mms.CreateMetricModel(testCtx, interfaces.MetricModel{
// 				SimpleMetricModel: interfaces.SimpleMetricModel{
// 					ModelName:    "16",
// 					MetricType:   "atomic",
// 					QueryType:    "dsl",
// 					Tags:         []string{"a", "s", "s", "s", "s"},
// 					Comment:      "ssss",
// 					UnitType:     "storeUnit",
// 					Unit:         "bit",
// 					Formula:      "{\"aggs\":{\"cpu2\":{\"aggs\":{\"mode\":{\"aggs\":{\"time2\":{\"aggs\":{\"value\":{\"top_hits\":{\"_source\":{\"includes\":[\"metrics.node_cpu_seconds_total\",\"@timestamp\"]},\"size\":3,\"sort\":[{\"@timestamp\":{\"order\":\"desc\"}}]}}},\"date_histogram\":{\"field\":\"@timestamp\",\"fixed_interval\":\"{{__interval}}\",\"format\":\"yyyy-MM-dd HH:mm:ss.SSS\",\"min_doc_count\":1,\"order\":{\"_key\":\"asc\"}}}},\"terms\":{\"field\":\"labels.mode.keyword\",\"size\":100000}}},\"terms\":{\"field\":\"labels.cpu.keyword\",\"size\":10000}}},\"query\":{\"bool\":{\"filter\":[{\"range\":{\"labels.cpu.keyword\":{\"lte\":\"2\"}}},{\"range\":{\"metrics.node_cpu_seconds_total\":{\"gte\":0}}}]}},\"size\":0,\"sort\":[{\"@timestamp\":{\"order\":\"desc\"}}]}",
// 					DateField:    "time2",
// 					MeasureField: "metrics.node_cpu_seconds_total",
// 				},
// 				DataSource: &interfaces.MetricDataSource{
// 					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
// 					ID:   "数据视图1",
// 				},
// 				IfContainTopHits: true,
// 			})
// 			So(modelID, ShouldEqual, "")
// 			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
// 			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_MeasureField)
// 		})

// 		Convey("DSL GetDataView error \n", func() {
// 			filtersErr := &rest.HTTPError{
// 				HTTPCode: http.StatusInternalServerError,
// 				Language: rest.DefaultLanguage,
// 				BaseError: rest.BaseError{
// 					ErrorCode: derrors.DataModel_MetricModel_InternalError_GetDataViewQueryFiltersFailed,
// 				},
// 			}
// 			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			dvs.EXPECT().CheckDataViewExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(viewName, true, nil)
// 			dvs.EXPECT().GetDataView(gomock.Any(), gomock.Any()).Return(&interfaces.DataViewQueryFilters{}, filtersErr)

// 			modelID, httpErr := mms.CreateMetricModel(testCtx, interfaces.MetricModel{
// 				SimpleMetricModel: interfaces.SimpleMetricModel{
// 					ModelName:    "16",
// 					MetricType:   "atomic",
// 					QueryType:    "dsl",
// 					Tags:         []string{"a", "s", "s", "s", "s"},
// 					Comment:      "ssss",
// 					UnitType:     "storeUnit",
// 					Unit:         "bit",
// 					Formula:      "{\"aggs\":{\"cpu2\":{\"aggs\":{\"mode\":{\"aggs\":{\"time2\":{\"aggs\":{\"value\":{\"top_hits\":{\"_source\":{\"includes\":[\"metrics.node_cpu_seconds_total\",\"@timestamp\"]},\"size\":3,\"sort\":[{\"@timestamp\":{\"order\":\"desc\"}}]}}},\"date_histogram\":{\"field\":\"@timestamp\",\"fixed_interval\":\"{{__interval}}\",\"format\":\"yyyy-MM-dd HH:mm:ss.SSS\",\"min_doc_count\":1,\"order\":{\"_key\":\"asc\"}}}},\"terms\":{\"field\":\"labels.mode.keyword\",\"size\":100000}}},\"terms\":{\"field\":\"labels.cpu.keyword\",\"size\":10000}}},\"query\":{\"bool\":{\"filter\":[{\"range\":{\"labels.cpu.keyword\":{\"lte\":\"2\"}}},{\"range\":{\"metrics.node_cpu_seconds_total\":{\"gte\":0}}}]}},\"size\":0,\"sort\":[{\"@timestamp\":{\"order\":\"desc\"}}]}",
// 					DateField:    "time2",
// 					MeasureField: "metrics.node_cpu_seconds_total",
// 				},
// 				DataSource: &interfaces.MetricDataSource{
// 					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
// 					ID:   "数据视图1",
// 				},
// 				IfContainTopHits: true,
// 			})
// 			So(modelID, ShouldEqual, "")
// 			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
// 			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InternalError_GetDataViewQueryFiltersFailed)
// 		})

// 		Convey("Formula invalid \n", func() {
// 			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			dvs.EXPECT().CheckDataViewExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(viewName, true, nil)
// 			ua.EXPECT().CheckFormulaByUniquery(gomock.Any(), gomock.Any()).
// 				Return(false, "1:75: parse error: unexpected <by> in aggregation", nil)
// 			dmja.EXPECT().StartJob(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

// 			modelID, httpErr := mms.CreateMetricModel(testCtx, metricModel)
// 			So(modelID, ShouldEqual, "")
// 			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
// 			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_Formula)
// 		})

// 		Convey("Task retraceduration invalid \n", func() {
// 			patch := ApplyFuncReturn(did.GenerateDistributedID, 0, nil)
// 			defer patch.Reset()

// 			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			dvs.EXPECT().CheckDataViewExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(viewName, true, nil)
// 			ua.EXPECT().CheckFormulaByUniquery(gomock.Any(), gomock.Any()).
// 				Return(true, "", nil)

// 			iba.EXPECT().GetSimpleIndexBasesByTypes(gomock.Any(), gomock.Any()).Return([]interfaces.SimpleIndexBase{{BaseType: "1", Name: "1"}}, nil)
// 			dmja.EXPECT().StartJob(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

// 			taskErr := interfaces.MetricTask{
// 				TaskID:   "1",
// 				TaskName: "task1",
// 				ModelID:  "1",
// 				Schedule: interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
// 				// Filters:            []interfaces.Filter{{Name: "labels.mode", Operation: "=", Value: "idle"}},
// 				TimeWindows:        []string{"5m", "1h"},
// 				Steps:              []string{"5m", "1h"},
// 				IndexBase:          "base1",
// 				RetraceDuration:    "1a",
// 				ScheduleSyncStatus: 1,
// 				Comment:            "task1-aaa",
// 				UpdateTime:         testMetricUpdateTime,
// 				PlanTime:           int64(1699336878575),
// 			}
// 			metricModelE := interfaces.MetricModel{
// 				SimpleMetricModel: interfaces.SimpleMetricModel{
// 					ModelName:  "16",
// 					MetricType: "atomic",
// 					QueryType:  "promql",
// 					Tags:       []string{"a", "s", "s", "s", "s"},
// 					Comment:    "ssss",
// 					GroupName:  "sss",
// 					Formula:    testFormula,
// 				},
// 				DataSource: &interfaces.MetricDataSource{
// 					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
// 					ID:   "数据视图1",
// 				},
// 				Task: &taskErr,
// 			}

// 			modelID, httpErr := mms.CreateMetricModel(testCtx, metricModelE)
// 			So(modelID, ShouldEqual, "")
// 			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
// 			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_RetraceDuration)
// 		})

// 		Convey("Transaction begin failed \n", func() {
// 			patch := ApplyFuncReturn(did.GenerateDistributedID, 0, nil)
// 			defer patch.Reset()

// 			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			dvs.EXPECT().CheckDataViewExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(viewName, true, nil)
// 			ua.EXPECT().CheckFormulaByUniquery(gomock.Any(), gomock.Any()).
// 				Return(true, "", nil)
// 			iba.EXPECT().GetSimpleIndexBasesByTypes(gomock.Any(), gomock.Any()).Return([]interfaces.SimpleIndexBase{{BaseType: "1", Name: "1"}}, nil)
// 			smock.ExpectBegin().WillReturnError(errors.New("error"))
// 			dmja.EXPECT().StartJob(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

// 			modelID, httpErr := mms.CreateMetricModel(testCtx, metricModel)
// 			So(modelID, ShouldEqual, "")
// 			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
// 			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InternalError_BeginTransactionFailed)
// 		})

// 		Convey("Transaction commit failed \n", func() {
// 			patch := ApplyFuncReturn(did.GenerateDistributedID, 0, nil)
// 			defer patch.Reset()

// 			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			dvs.EXPECT().CheckDataViewExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(viewName, true, nil)
// 			ua.EXPECT().CheckFormulaByUniquery(gomock.Any(), gomock.Any()).
// 				Return(true, "", nil)
// 			mmga.EXPECT().GetMetricModelGroupByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(testMetricGroup, true, nil)
// 			iba.EXPECT().GetSimpleIndexBasesByTypes(gomock.Any(), gomock.Any()).Return([]interfaces.SimpleIndexBase{{BaseType: "1", Name: "1"}}, nil)
// 			dmja.EXPECT().StartJob(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

// 			smock.ExpectBegin()
// 			mma.EXPECT().CreateMetricModel(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			mmts.EXPECT().CreateMetricTask(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
// 			smock.ExpectCommit().WillReturnError(errors.New("error"))

// 			_, err := mms.CreateMetricModel(testCtx, metricModel)
// 			So(err, ShouldResemble, errors.New("error"))
// 		})

// 		Convey("CreateMetricTasks failed \n", func() {
// 			patch := ApplyFuncReturn(did.GenerateDistributedID, 0, nil)
// 			defer patch.Reset()

// 			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			smock.ExpectBegin()
// 			dvs.EXPECT().CheckDataViewExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(viewName, true, nil)
// 			ua.EXPECT().CheckFormulaByUniquery(gomock.Any(), gomock.Any()).
// 				Return(true, "", nil)
// 			mmga.EXPECT().GetMetricModelGroupByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(testMetricGroup, true, nil)
// 			iba.EXPECT().GetSimpleIndexBasesByTypes(gomock.Any(), gomock.Any()).Return([]interfaces.SimpleIndexBase{{BaseType: "1", Name: "1"}}, nil)
// 			mma.EXPECT().CreateMetricModel(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			mmts.EXPECT().CreateMetricTask(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(errors.New("error"))
// 			dmja.EXPECT().StartJob(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

// 			modelID, httpErr := mms.CreateMetricModel(testCtx, metricModel)
// 			So(modelID, ShouldEqual, "")
// 			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
// 			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InternalError)
// 		})

// 	})
// }

func Test_MetricModelService_UpdateMetricModel(t *testing.T) {
	Convey("Test UpdateMetricModel", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		mma := dmock.NewMockMetricModelAccess(mockCtrl)
		mmga := dmock.NewMockMetricModelGroupAccess(mockCtrl)
		ua := dmock.NewMockUniqueryAccess(mockCtrl)
		mmts := dmock.NewMockMetricModelTaskService(mockCtrl)
		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		mms, smock := MockNewMetricModelService(appSetting, dmja, dvs, mma, mmga, ua, mmts, iba, ps)

		metricModel := interfaces.MetricModel{
			SimpleMetricModel: interfaces.SimpleMetricModel{
				ModelName:  "16",
				MetricType: "atomic",
				QueryType:  "promql",
				Tags:       []string{"a", "s", "s", "s", "s"},
				Comment:    "ssss",
				GroupID:    "111",
				Formula:    testFormula,
			},
			DataSource: &interfaces.MetricDataSource{
				Type: interfaces.SOURCE_TYPE_DATA_VIEW,
				ID:   "数据视图1",
			},
			Task: &task,
		}

		viewName := "数据视图1"

		modelTaskMap := make(map[string]interfaces.MetricTask)
		modelTaskMap[metricModel.ModelID] = task

		Convey("UpdateMetricModel failed", func() {
			err := errors.New("Update Metric Model failed")
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin()
			mmga.EXPECT().GetMetricModelGroupByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(testMetricGroup, true, nil)
			mma.EXPECT().UpdateMetricModel(gomock.Any(), gomock.Any(), gomock.Any()).Return(err)
			smock.ExpectCommit()
			dvs.EXPECT().GetDataView(gomock.Any(), gomock.Any()).
				Return(&interfaces.DataView{
					SimpleDataView: interfaces.SimpleDataView{
						ViewName:   viewName,
						QueryType:  interfaces.QueryType_DSL,
						Operations: []string{interfaces.OPERATION_TYPE_DATA_QUERY},
					},
				}, nil)
			ua.EXPECT().CheckFormulaByUniquery(gomock.Any(), gomock.Any()).
				Return(true, "", nil)
			dmja.EXPECT().UpdateJob(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			httpErr := mms.UpdateMetricModel(testCtx, nil, metricModel)
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InternalError)
		})

		Convey("UpdateMetricModel success", func() {

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ua.EXPECT().CheckFormulaByUniquery(gomock.Any(), gomock.Any()).
				Return(true, "", nil)
			// mmts.EXPECT().GetMetricTaskIDsByModelIDs(gomock.Any(), gomock.Any()).Return(nil, nil)
			smock.ExpectBegin()
			dvs.EXPECT().GetDataView(gomock.Any(), gomock.Any()).
				Return(&interfaces.DataView{
					SimpleDataView: interfaces.SimpleDataView{
						ViewName:   viewName,
						QueryType:  interfaces.QueryType_DSL,
						Operations: []string{interfaces.OPERATION_TYPE_DATA_QUERY},
					}}, nil)

			mmga.EXPECT().GetMetricModelGroupByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(testMetricGroup, true, nil)
			mma.EXPECT().UpdateMetricModel(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mmts.EXPECT().GetMetricTasksByModelIDs(gomock.Any(), gomock.Any()).Return(modelTaskMap, nil)
			// mmts.EXPECT().GetMetricTasksByTaskIDs(gomock.Any(), gomock.Any()).Return([]interfaces.MetricTask{task}, nil)
			mmts.EXPECT().UpdateMetricTask(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()
			dmja.EXPECT().UpdateJob(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			httpErr := mms.UpdateMetricModel(testCtx, nil, metricModel)
			So(httpErr, ShouldBeNil)
		})

		// Convey("Data view not exists \n", func() {
		// 	ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		// 	dvs.EXPECT().GetDataView(gomock.Any(), gomock.Any()).
		// 		Return(nil, errors.New("noe exist"))

		// 	httpErr := mms.UpdateMetricModel(testCtx, nil, metricModel)
		// 	So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
		// 	So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_DataView_DataViewNotFound)
		// })

		// Convey("Get Data view error \n", func() {
		// 	err := &rest.HTTPError{
		// 		HTTPCode: http.StatusInternalServerError,
		// 		Language: rest.DefaultLanguage,
		// 		BaseError: rest.BaseError{
		// 			ErrorCode: derrors.DataModel_MetricModel_InternalError_GetDataViewByNameFailed,
		// 		},
		// 	}
		// 	ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		// 	dvs.EXPECT().GetDataView(gomock.Any(), gomock.Any()).
		// 		Return(nil, err)

		// 	httpErr := mms.UpdateMetricModel(testCtx, nil, metricModel)
		// 	So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
		// 	So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_DataView_DataViewNotFound)
		// })

		// Convey("Get vega logic view error \n", func() {
		// 	model := interfaces.MetricModel{
		// 		SimpleMetricModel: interfaces.SimpleMetricModel{
		// 			ModelName:  "16",
		// 			MetricType: "atomic",
		// 			QueryType:  "promql",
		// 			Tags:       []string{"a", "s", "s", "s", "s"},
		// 			Comment:    "ssss",
		// 			GroupName:  "sss",
		// 			Formula:    testFormula,
		// 		},
		// 		DataSource: &interfaces.MetricDataSource{
		// 			Type: interfaces.QueryType_SQL,
		// 			ID:   "视图1",
		// 		},
		// 	}

		// 	ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		// 	dvs.EXPECT().GetDataView(gomock.Any(), gomock.Any()).
		// 		Return(nil, fmt.Errorf("error"))

		// 	httpErr := mms.UpdateMetricModel(testCtx, nil, model)
		// 	So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
		// 	So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InternalError_GetVegaViewFieldsByIDFailed)
		// })

		// Convey("CheckSqlFormulaConfig error \n", func() {
		// 	model := interfaces.MetricModel{
		// 		SimpleMetricModel: interfaces.SimpleMetricModel{
		// 			ModelName:  "16",
		// 			MetricType: "atomic",
		// 			QueryType:  "sql",
		// 			Tags:       []string{"a", "s", "s", "s", "s"},
		// 			Comment:    "ssss",
		// 			GroupName:  "sss",
		// 			FormulaConfig: interfaces.SQLConfig{
		// 				AggrExpr: &interfaces.AggrExpr{
		// 					Aggr:  "count",
		// 					Field: "f1",
		// 				},
		// 			},
		// 			DateField: "f2",
		// 		},
		// 		DataSource: &interfaces.MetricDataSource{
		// 			Type: interfaces.QueryType_SQL,
		// 			ID:   "视图1",
		// 		},
		// 	}
		// 	vegaViewFields := interfaces.VegaViewWithFields{
		// 		Fields: []interfaces.VegaViewField{
		// 			{
		// 				Name: "f1",
		// 				Type: "varchar",
		// 			},
		// 			{
		// 				Name: "f2",
		// 				Type: "varchar",
		// 			},
		// 		},
		// 		FieldMap: map[string]interfaces.VegaViewField{
		// 			"f1": {
		// 				Name: "f1",
		// 				Type: "varchar",
		// 			},
		// 			"f2": {
		// 				Name: "f2",
		// 				Type: "varchar",
		// 			},
		// 		},
		// 	}

		// 	ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		// 	vegaService.EXPECT().GetVegaViewFieldsByID(gomock.Any(), gomock.Any()).Return(vegaViewFields, nil)
		// 	ua.EXPECT().CheckFormulaByUniquery(gomock.Any(), gomock.Any()).
		// 		Return(false, "invalid", nil)

		// 	modelID, httpErr := mms.CreateMetricModel(testCtx, model)
		// 	So(modelID, ShouldEqual, "")
		// 	So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
		// 	So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_Formula)
		// })

		Convey("DSL measure field type si not a number \n", func() {
			dataViewFilters := &interfaces.DataView{
				SimpleDataView: interfaces.SimpleDataView{
					ViewName: "数据视图1",
				},
				FieldsMap: map[string]*interfaces.ViewField{
					"metrics.node_cpu_seconds_total": {
						Name: "metrics.node_cpu_seconds_total",
						Type: dtype.DataType_String,
					},
				},
			}
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dvs.EXPECT().GetDataView(gomock.Any(), gomock.Any()).
				Return(&interfaces.DataView{
					SimpleDataView: interfaces.SimpleDataView{
						ViewName:   viewName,
						QueryType:  interfaces.QueryType_DSL,
						Operations: []string{interfaces.OPERATION_TYPE_DATA_QUERY},
					}}, nil)

			dvs.EXPECT().GetDataView(gomock.Any(), gomock.Any()).
				Return(dataViewFilters, nil)
			dmja.EXPECT().UpdateJob(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			httpErr := mms.UpdateMetricModel(testCtx, nil, interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:    "16",
					MetricType:   "atomic",
					QueryType:    "dsl",
					Tags:         []string{"a", "s", "s", "s", "s"},
					Comment:      "ssss",
					UnitType:     "storeUnit",
					Unit:         "bit",
					Formula:      "{\"aggs\":{\"cpu2\":{\"aggs\":{\"mode\":{\"aggs\":{\"time2\":{\"aggs\":{\"value\":{\"top_hits\":{\"_source\":{\"includes\":[\"metrics.node_cpu_seconds_total\",\"@timestamp\"]},\"size\":3,\"sort\":[{\"@timestamp\":{\"order\":\"desc\"}}]}}},\"date_histogram\":{\"field\":\"@timestamp\",\"fixed_interval\":\"{{__interval}}\",\"format\":\"yyyy-MM-dd HH:mm:ss.SSS\",\"min_doc_count\":1,\"order\":{\"_key\":\"asc\"}}}},\"terms\":{\"field\":\"labels.mode.keyword\",\"size\":100000}}},\"terms\":{\"field\":\"labels.cpu.keyword\",\"size\":10000}}},\"query\":{\"bool\":{\"filter\":[{\"range\":{\"labels.cpu.keyword\":{\"lte\":\"2\"}}},{\"range\":{\"metrics.node_cpu_seconds_total\":{\"gte\":0}}}]}},\"size\":0,\"sort\":[{\"@timestamp\":{\"order\":\"desc\"}}]}",
					DateField:    "time2",
					MeasureField: "metrics.node_cpu_seconds_total",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "数据视图1",
				},
				IfContainTopHits: true,
			})
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_MeasureField)
		})

		Convey("DSL GetDataView error \n", func() {
			filtersErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: derrors.DataModel_MetricModel_InternalError_GetDataViewQueryFiltersFailed,
				},
			}
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dvs.EXPECT().GetDataView(gomock.Any(), gomock.Any()).
				Return(&interfaces.DataView{
					SimpleDataView: interfaces.SimpleDataView{
						ViewName:   viewName,
						QueryType:  interfaces.QueryType_DSL,
						Operations: []string{interfaces.OPERATION_TYPE_DATA_QUERY}}}, nil)
			dvs.EXPECT().GetDataView(gomock.Any(), gomock.Any()).Return(&interfaces.DataView{}, filtersErr)
			dmja.EXPECT().UpdateJob(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			httpErr := mms.UpdateMetricModel(testCtx, nil, interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:    "16",
					MetricType:   "atomic",
					QueryType:    "dsl",
					Tags:         []string{"a", "s", "s", "s", "s"},
					Comment:      "ssss",
					UnitType:     "storeUnit",
					Unit:         "bit",
					Formula:      "{\"aggs\":{\"cpu2\":{\"aggs\":{\"mode\":{\"aggs\":{\"time2\":{\"aggs\":{\"value\":{\"top_hits\":{\"_source\":{\"includes\":[\"metrics.node_cpu_seconds_total\",\"@timestamp\"]},\"size\":3,\"sort\":[{\"@timestamp\":{\"order\":\"desc\"}}]}}},\"date_histogram\":{\"field\":\"@timestamp\",\"fixed_interval\":\"{{__interval}}\",\"format\":\"yyyy-MM-dd HH:mm:ss.SSS\",\"min_doc_count\":1,\"order\":{\"_key\":\"asc\"}}}},\"terms\":{\"field\":\"labels.mode.keyword\",\"size\":100000}}},\"terms\":{\"field\":\"labels.cpu.keyword\",\"size\":10000}}},\"query\":{\"bool\":{\"filter\":[{\"range\":{\"labels.cpu.keyword\":{\"lte\":\"2\"}}},{\"range\":{\"metrics.node_cpu_seconds_total\":{\"gte\":0}}}]}},\"size\":0,\"sort\":[{\"@timestamp\":{\"order\":\"desc\"}}]}",
					DateField:    "time2",
					MeasureField: "metrics.node_cpu_seconds_total",
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "数据视图1",
				},
				IfContainTopHits: true,
			})
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InternalError_GetDataViewQueryFiltersFailed)
		})

		Convey("Formula invalid \n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dvs.EXPECT().GetDataView(gomock.Any(), gomock.Any()).
				Return(&interfaces.DataView{
					SimpleDataView: interfaces.SimpleDataView{
						ViewName:   viewName,
						QueryType:  interfaces.QueryType_DSL,
						Operations: []string{interfaces.OPERATION_TYPE_DATA_QUERY},
					}}, nil)
			ua.EXPECT().CheckFormulaByUniquery(gomock.Any(), gomock.Any()).
				Return(false, "1:75: parse error: unexpected <by> in aggregation", nil)
			dmja.EXPECT().UpdateJob(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			httpErr := mms.UpdateMetricModel(testCtx, nil, metricModel)
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_Formula)
		})

		Convey("Task retraceduration invalid \n", func() {
			patch := ApplyFuncReturn(did.GenerateDistributedID, 0, nil)
			defer patch.Reset()

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dvs.EXPECT().GetDataView(gomock.Any(), gomock.Any()).
				Return(&interfaces.DataView{
					SimpleDataView: interfaces.SimpleDataView{
						ViewName:   viewName,
						QueryType:  interfaces.QueryType_DSL,
						Operations: []string{interfaces.OPERATION_TYPE_DATA_QUERY},
					}}, nil)
			ua.EXPECT().CheckFormulaByUniquery(gomock.Any(), gomock.Any()).
				Return(true, "", nil)
			dmja.EXPECT().UpdateJob(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			smock.ExpectBegin()
			mmga.EXPECT().GetMetricModelGroupByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(testMetricGroup, true, nil)
			mmts.EXPECT().GetMetricTasksByModelIDs(gomock.Any(), gomock.Any()).Return(nil, nil)
			mma.EXPECT().UpdateMetricModel(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()

			taskErr := interfaces.MetricTask{
				TaskName: "task1",
				ModelID:  "1",
				Schedule: interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
				// Filters:         []interfaces.Filter{{Name: "labels.mode", Operation: "=", Value: "idle"}},
				TimeWindows:     []string{"5m", "1h"},
				Steps:           []string{"5m", "1h"},
				IndexBase:       "base1",
				RetraceDuration: "1a",
				Comment:         "task1-aaa",
				UpdateTime:      testMetricUpdateTime,
				PlanTime:        int64(1699336878575),
			}
			metricModelE := interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "16",
					MetricType: "atomic",
					QueryType:  "promql",
					Tags:       []string{"a", "s", "s", "s", "s"},
					Comment:    "ssss",
					Formula:    testFormula,
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "数据视图1",
				},
				Task: &taskErr,
			}

			httpErr := mms.UpdateMetricModel(testCtx, nil, metricModelE)
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_RetraceDuration)
		})

		Convey("Transaction begin failed \n", func() {
			patch := ApplyFuncReturn(did.GenerateDistributedID, 0, nil)
			defer patch.Reset()

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dvs.EXPECT().GetDataView(gomock.Any(), gomock.Any()).
				Return(&interfaces.DataView{
					SimpleDataView: interfaces.SimpleDataView{
						ViewName:   viewName,
						QueryType:  interfaces.QueryType_DSL,
						Operations: []string{interfaces.OPERATION_TYPE_DATA_QUERY},
					}}, nil)
			ua.EXPECT().CheckFormulaByUniquery(gomock.Any(), gomock.Any()).
				Return(true, "", nil)
			smock.ExpectBegin().WillReturnError(errors.New("error"))
			dmja.EXPECT().UpdateJob(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			httpErr := mms.UpdateMetricModel(testCtx, nil, metricModel)
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InternalError_BeginTransactionFailed)
		})

		Convey("Transaction commit failed \n", func() {
			patch := ApplyFuncReturn(did.GenerateDistributedID, 0, nil)
			defer patch.Reset()

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dvs.EXPECT().GetDataView(gomock.Any(), gomock.Any()).
				Return(&interfaces.DataView{
					SimpleDataView: interfaces.SimpleDataView{
						ViewName:   viewName,
						QueryType:  interfaces.QueryType_DSL,
						Operations: []string{interfaces.OPERATION_TYPE_DATA_QUERY},
					}}, nil)
			ua.EXPECT().CheckFormulaByUniquery(gomock.Any(), gomock.Any()).
				Return(true, "", nil)
			// mmts.EXPECT().GetMetricTaskIDsByModelIDs(gomock.Any(), gomock.Any()).Return(nil, nil)
			dmja.EXPECT().UpdateJob(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			smock.ExpectBegin()
			mmga.EXPECT().GetMetricModelGroupByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(testMetricGroup, true, nil)
			mma.EXPECT().UpdateMetricModel(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			// mmts.EXPECT().GetMetricTasksByTaskIDs(gomock.Any(), gomock.Any()).Return([]interfaces.MetricTask{task}, nil)
			mmts.EXPECT().GetMetricTasksByModelIDs(gomock.Any(), gomock.Any()).Return(modelTaskMap, nil)
			mmts.EXPECT().UpdateMetricTask(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit().WillReturnError(errors.New("error"))

			err := mms.UpdateMetricModel(testCtx, nil, metricModel)
			So(err, ShouldResemble, errors.New("error"))
		})

		Convey("UpdateMetricTask failed \n", func() {
			patch := ApplyFuncReturn(did.GenerateDistributedID, 0, nil)
			defer patch.Reset()

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin()
			dvs.EXPECT().GetDataView(gomock.Any(), gomock.Any()).
				Return(&interfaces.DataView{
					SimpleDataView: interfaces.SimpleDataView{
						ViewName:   viewName,
						QueryType:  interfaces.QueryType_DSL,
						Operations: []string{interfaces.OPERATION_TYPE_DATA_QUERY},
					}}, nil)
			ua.EXPECT().CheckFormulaByUniquery(gomock.Any(), gomock.Any()).
				Return(true, "", nil)
			dmja.EXPECT().UpdateJob(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			// mmts.EXPECT().GetMetricTaskIDsByModelIDs(gomock.Any(), gomock.Any()).Return(nil, nil)
			// mmts.EXPECT().GetMetricTasksByTaskIDs(gomock.Any(), gomock.Any()).Return([]interfaces.MetricTask{task}, nil)
			mmga.EXPECT().GetMetricModelGroupByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(testMetricGroup, true, nil)
			mmts.EXPECT().GetMetricTasksByModelIDs(gomock.Any(), gomock.Any()).Return(modelTaskMap, nil)
			mma.EXPECT().UpdateMetricModel(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mmts.EXPECT().UpdateMetricTask(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(errors.New("error"))

			httpErr := mms.UpdateMetricModel(testCtx, nil, metricModel)
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InternalError)
		})

		Convey("Update model CreateMetricTasks failed \n", func() {
			taskErr := interfaces.MetricTask{
				TaskName: "task1",
				ModelID:  "1",
				Schedule: interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
				// Filters:         []interfaces.Filter{{Name: "labels.mode", Operation: "=", Value: "idle"}},
				TimeWindows:     []string{"5m", "1h"},
				Steps:           []string{"5m", "1h"},
				IndexBase:       "base1",
				RetraceDuration: "1h",
				Comment:         "task1-aaa",
				UpdateTime:      testMetricUpdateTime,
				PlanTime:        int64(1699336878575),
			}
			metricModelE := interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "16",
					MetricType: "atomic",
					QueryType:  "promql",
					Tags:       []string{"a", "s", "s", "s", "s"},
					Comment:    "ssss",
					Formula:    testFormula,
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "数据视图1",
				},
				Task: &taskErr,
			}

			patch := ApplyFuncReturn(did.GenerateDistributedID, 0, nil)
			defer patch.Reset()

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin()
			dvs.EXPECT().GetDataView(gomock.Any(), gomock.Any()).
				Return(&interfaces.DataView{
					SimpleDataView: interfaces.SimpleDataView{
						ViewName:   viewName,
						QueryType:  interfaces.QueryType_DSL,
						Operations: []string{interfaces.OPERATION_TYPE_DATA_QUERY},
					}}, nil)
			ua.EXPECT().CheckFormulaByUniquery(gomock.Any(), gomock.Any()).
				Return(true, "", nil)
			dmja.EXPECT().UpdateJob(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			// mmts.EXPECT().GetMetricTaskIDsByModelIDs(gomock.Any(), gomock.Any()).Return(nil, nil)
			mmga.EXPECT().GetMetricModelGroupByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(testMetricGroup, true, nil)
			mmts.EXPECT().GetMetricTasksByModelIDs(gomock.Any(), gomock.Any()).Return(nil, nil)
			// mmts.EXPECT().CheckMetricModelTaskExistByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)
			mma.EXPECT().UpdateMetricModel(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mmts.EXPECT().CreateMetricTask(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(errors.New("error"))

			httpErr := mms.UpdateMetricModel(testCtx, nil, metricModelE)
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InternalError)
		})

		Convey("Update model DeleteMetricTaskByTaskIDs failed \n", func() {
			metricModelE := interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelName:  "16",
					MetricType: "atomic",
					QueryType:  "promql",
					Tags:       []string{"a", "s", "s", "s", "s"},
					Comment:    "ssss",
					Formula:    testFormula,
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "数据视图1",
				},
			}

			patch := ApplyFuncReturn(did.GenerateDistributedID, 0, nil)
			defer patch.Reset()

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin()
			dvs.EXPECT().GetDataView(gomock.Any(), gomock.Any()).
				Return(&interfaces.DataView{SimpleDataView: interfaces.SimpleDataView{ViewName: viewName, QueryType: interfaces.QueryType_DSL,
					Operations: []string{interfaces.OPERATION_TYPE_DATA_QUERY}}}, nil)
			ua.EXPECT().CheckFormulaByUniquery(gomock.Any(), gomock.Any()).
				Return(true, "", nil)

			// mmts.EXPECT().GetMetricTaskIDsByModelIDs(gomock.Any(), gomock.Any()).Return([]string{"1"}, nil)
			// mmts.EXPECT().CheckMetricModelTaskExistByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)
			mmga.EXPECT().GetMetricModelGroupByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(testMetricGroup, true, nil)
			mmts.EXPECT().GetMetricTasksByModelIDs(gomock.Any(), gomock.Any()).Return(modelTaskMap, nil)
			mma.EXPECT().UpdateMetricModel(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mmts.EXPECT().DeleteMetricTaskByTaskIDs(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(errors.New("error"))
			dmja.EXPECT().UpdateJob(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			httpErr := mms.UpdateMetricModel(testCtx, nil, metricModelE)
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InternalError)
		})

	})
}

func Test_MetricModelService_GetMetricModelByID(t *testing.T) {
	Convey("Test GetMetricModelByModelID", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		mma := dmock.NewMockMetricModelAccess(mockCtrl)
		mmga := dmock.NewMockMetricModelGroupAccess(mockCtrl)
		ua := dmock.NewMockUniqueryAccess(mockCtrl)
		mmts := dmock.NewMockMetricModelTaskService(mockCtrl)

		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		mms, _ := MockNewMetricModelService(appSetting, dmja, dvs, mma, mmga, ua, mmts, iba, ps)

		metricModel := interfaces.MetricModel{
			SimpleMetricModel: interfaces.SimpleMetricModel{
				ModelName:  "16",
				MetricType: "atomic",
				QueryType:  "promql",
				Tags:       []string{"a", "s", "s", "s", "s"},
				Comment:    "ssss",
				Formula:    testFormula,
			},
			DataSource: &interfaces.MetricDataSource{
				Type: interfaces.SOURCE_TYPE_DATA_VIEW,
				ID:   "数据视图1",
			},
		}
		var emptyModel interfaces.MetricModel

		Convey("GetMetricModelByModelID failed", func() {
			err := errors.New("Get Metric Model By ID failed")
			mma.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(emptyModel, false, err)

			actualModel, httpErr := mms.GetMetricModelByModelID(testCtx, "0")
			So(actualModel, ShouldResemble, emptyModel)
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InternalError_GetModelByIDFailed)
		})

		Convey("Metric Model Not Found", func() {
			mma.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(emptyModel, false, nil)

			actualModel, httpErr := mms.GetMetricModelByModelID(testCtx, "0")
			So(actualModel, ShouldResemble, emptyModel)
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusNotFound)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_MetricModelNotFound)
		})

		Convey("GetMetricModelByModelID success", func() {

			mma.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(metricModel, true, nil)

			actualModel, httpErr := mms.GetMetricModelByModelID(testCtx, "0")
			So(actualModel, ShouldResemble, metricModel)
			So(httpErr, ShouldBeNil)
		})
	})
}

func Test_MetricModelService_DeleteMetricModels(t *testing.T) {
	Convey("Test DeleteMetricModels", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		mma := dmock.NewMockMetricModelAccess(mockCtrl)
		mmga := dmock.NewMockMetricModelGroupAccess(mockCtrl)
		ua := dmock.NewMockUniqueryAccess(mockCtrl)
		mmts := dmock.NewMockMetricModelTaskService(mockCtrl)
		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		mms, smock := MockNewMetricModelService(appSetting, dmja, dvs, mma, mmga, ua, mmts, iba, ps)
		resrc := map[string]interfaces.ResourceOps{
			"0": {
				ResourceID: "0",
			},
			"1": {
				ResourceID: "1",
			},
		}
		Convey("DeleteMetricModels failed", func() {
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			mmts.EXPECT().GetMetricTaskIDsByModelIDs(gomock.Any(), gomock.Any()).Return([]string{"1"}, nil)
			dmja.EXPECT().StopJobs(gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin()
			err := errors.New("Delete Metric Models failed")
			mma.EXPECT().DeleteMetricModels(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), err)
			smock.ExpectCommit()

			rowsAffect, httpErr := mms.DeleteMetricModels(testCtx, nil, []string{"0", "1"})
			So(rowsAffect, ShouldEqual, 0)
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InternalError)
		})

		Convey("DeleteMetricModels success && effect rows != len(modelIDs)", func() {
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			ps.EXPECT().DeleteResources(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mmts.EXPECT().GetMetricTaskIDsByModelIDs(gomock.Any(), gomock.Any()).Return([]string{"1"}, nil)
			dmja.EXPECT().StopJobs(gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin()
			mma.EXPECT().DeleteMetricModels(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			mmts.EXPECT().DeleteMetricTaskByTaskIDs(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			smock.ExpectCommit()

			rowsAffect, httpErr := mms.DeleteMetricModels(testCtx, nil, []string{"0", "1"})
			So(rowsAffect, ShouldEqual, 1)
			So(httpErr, ShouldBeNil)
		})

		Convey("DeleteMetricModels success && effect rows == len(modelIDs)", func() {
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			ps.EXPECT().DeleteResources(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mmts.EXPECT().GetMetricTaskIDsByModelIDs(gomock.Any(), gomock.Any()).Return([]string{"1"}, nil)
			dmja.EXPECT().StopJobs(gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin()
			mma.EXPECT().DeleteMetricModels(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(2), nil)
			mmts.EXPECT().DeleteMetricTaskByTaskIDs(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			smock.ExpectCommit()

			rowsAffect, httpErr := mms.DeleteMetricModels(testCtx, nil, []string{"0", "1"})
			So(rowsAffect, ShouldEqual, 2)
			So(httpErr, ShouldBeNil)
		})

		Convey("Transaction begin failed \n", func() {
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			mmts.EXPECT().GetMetricTaskIDsByModelIDs(gomock.Any(), gomock.Any()).Return([]string{"1"}, nil)
			dmja.EXPECT().StopJobs(gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin().WillReturnError(errors.New("error"))

			rowsAffect, httpErr := mms.DeleteMetricModels(testCtx, nil, []string{"0", "1"})
			So(rowsAffect, ShouldEqual, 0)
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InternalError_BeginTransactionFailed)
		})

		Convey("Transaction commit failed \n", func() {
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			ps.EXPECT().DeleteResources(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mmts.EXPECT().GetMetricTaskIDsByModelIDs(gomock.Any(), gomock.Any()).Return(nil, nil)

			smock.ExpectBegin()
			mma.EXPECT().DeleteMetricModels(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			mmts.EXPECT().DeleteMetricTaskByTaskIDs(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			smock.ExpectCommit().WillReturnError(errors.New("error"))

			rowsAffect, err := mms.DeleteMetricModels(testCtx, nil, []string{"0", "1"})
			So(rowsAffect, ShouldEqual, 1)
			So(err, ShouldResemble, errors.New("error"))
		})

		Convey("DeleteMetricTaskByTaskIDs failed \n", func() {
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			mmts.EXPECT().GetMetricTaskIDsByModelIDs(gomock.Any(), gomock.Any()).Return(nil, nil)

			smock.ExpectBegin()
			mma.EXPECT().DeleteMetricModels(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			mmts.EXPECT().DeleteMetricTaskByTaskIDs(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(errors.New("error"))
			smock.ExpectCommit()

			rowsAffect, err := mms.DeleteMetricModels(testCtx, nil, []string{"0", "1"})
			So(rowsAffect, ShouldEqual, 0)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InternalError)
		})
	})
}

func Test_MetricModelService_ListSimpleMetricModels(t *testing.T) {
	Convey("Test ListSimpleMetricModels", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		mma := dmock.NewMockMetricModelAccess(mockCtrl)
		mmga := dmock.NewMockMetricModelGroupAccess(mockCtrl)
		ua := dmock.NewMockUniqueryAccess(mockCtrl)
		mmts := dmock.NewMockMetricModelTaskService(mockCtrl)

		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		mms, _ := MockNewMetricModelService(appSetting, dmja, dvs, mma, mmga, ua, mmts, iba, ps)

		simpleMetricModel := interfaces.SimpleMetricModel{
			ModelID:    "1",
			ModelName:  "16",
			MetricType: "atomic",
			QueryType:  "promql",
			Tags:       []string{"a", "s", "s", "s", "s"},
		}
		var emptyModels []interfaces.SimpleMetricModel
		resrc := map[string]interfaces.ResourceOps{
			"1": {
				ResourceID: "1",
			},
		}
		Convey("ListSimpleMetricModels failed", func() {
			err := errors.New("List Simple Metric Models failed")
			mma.EXPECT().ListSimpleMetricModels(gomock.Any(), gomock.Any()).Return(emptyModels, err)

			parameter := interfaces.MetricModelsQueryParams{}
			metricModelArr, cnt, httpErr := mms.ListSimpleMetricModels(testCtx, parameter)
			So(metricModelArr, ShouldBeEmpty)
			So(cnt, ShouldEqual, 0)
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InternalError)
		})

		// Convey("GetMetricModelsTotal failed", func() {
		// 	err := errors.New("GetMetricModelsTotal failed")
		// 	ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		// 		Return(resrc, nil)
		// 	mma.EXPECT().ListSimpleMetricModels(gomock.Any(), gomock.Any()).Return([]interfaces.SimpleMetricModel{simpleMetricModel}, nil)
		// 	// mma.EXPECT().GetMetricModelsTotal(gomock.Any(), gomock.Any()).Return(0, err)

		// 	parameter := interfaces.MetricModelsQueryParams{}
		// 	metricModelArr, cnt, httpErr := mms.ListSimpleMetricModels(testCtx, parameter)
		// 	So(metricModelArr, ShouldBeEmpty)
		// 	So(cnt, ShouldEqual, 0)
		// 	So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
		// 	So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InternalError)
		// })

		Convey("ListSimpleMetricModels success", func() {
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			mma.EXPECT().ListSimpleMetricModels(gomock.Any(), gomock.Any()).Return([]interfaces.SimpleMetricModel{simpleMetricModel}, nil)
			// mma.EXPECT().GetMetricModelsTotal(gomock.Any(), gomock.Any()).Return(1, nil)

			parameter := interfaces.MetricModelsQueryParams{
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit: 10,
				},
			}
			metricModelArr, cnt, httpErr := mms.ListSimpleMetricModels(testCtx, parameter)

			So(metricModelArr, ShouldResemble, []interfaces.SimpleMetricModel{simpleMetricModel})
			So(cnt, ShouldEqual, 1)
			So(httpErr, ShouldBeNil)
		})

	})
}

func Test_MetricModelService_GetMetricModelContainFilters(t *testing.T) {
	Convey("Test GetMetricModelContainFilters", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		mma := dmock.NewMockMetricModelAccess(mockCtrl)
		mmga := dmock.NewMockMetricModelGroupAccess(mockCtrl)
		ua := dmock.NewMockUniqueryAccess(mockCtrl)
		mmts := dmock.NewMockMetricModelTaskService(mockCtrl)
		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		mms, _ := MockNewMetricModelService(appSetting, dmja, dvs, mma, mmga, ua, mmts, iba, ps)

		metricModel := interfaces.MetricModel{
			SimpleMetricModel: interfaces.SimpleMetricModel{
				ModelID:    "0",
				ModelName:  "16",
				MetricType: "atomic",
				QueryType:  "promql",
				Tags:       []string{"a", "s", "s", "s", "s"},
				Comment:    "ssss",
				Formula:    testFormula,
				DataViewID: "数据视图1",
			},
			DataSource: &interfaces.MetricDataSource{
				Type: interfaces.SOURCE_TYPE_DATA_VIEW,
				ID:   "数据视图1",
			},
			// DataViewName: "数据视图1",
			Task: &task,
		}
		modelTaskMap := make(map[string]interfaces.MetricTask)
		modelTaskMap[metricModel.ModelID] = task

		emptyModel := make([]interfaces.MetricModel, 0)
		emptyModelFilters := make([]interfaces.MetricModelWithFilters, 0)
		resrc := map[string]interfaces.ResourceOps{
			"0": {
				ResourceID: "0",
			},
		}
		Convey("GetMetricModelByModelID failed", func() {
			err := errors.New("Get Metric Models By ID failed")
			mma.EXPECT().GetMetricModelsByModelIDs(gomock.Any(), gomock.Any()).Return(emptyModel, err)

			metricModelFilters, httpErr := mms.GetMetricModels(testCtx, []string{"0"}, true)
			So(metricModelFilters, ShouldResemble, emptyModelFilters)
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InternalError_GetModelByIDFailed)
		})

		Convey("GetMetricTasksByModelIDs Failed", func() {
			err := errors.New("GetMetricTasksByModelIDs Failed")
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			mma.EXPECT().GetMetricModelsByModelIDs(gomock.Any(), gomock.Any()).Return([]interfaces.MetricModel{metricModel}, nil)
			mmts.EXPECT().GetMetricTasksByModelIDs(gomock.Any(), gomock.Any()).Return(nil, err)

			metricModelFilters, httpErr := mms.GetMetricModels(testCtx, []string{"0"}, true)
			So(metricModelFilters, ShouldResemble, emptyModelFilters)
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InternalError_GetMetricTasksByModelIDsFailed)
		})

		Convey("GetDataView Failed", func() {
			err := errors.New("GetDataView Failed")
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			mma.EXPECT().GetMetricModelsByModelIDs(gomock.Any(), gomock.Any()).Return([]interfaces.MetricModel{metricModel}, nil)
			mmts.EXPECT().GetMetricTasksByModelIDs(gomock.Any(), gomock.Any()).Return(nil, nil)
			dvs.EXPECT().GetDataView(gomock.Any(), gomock.Any()).Return(&interfaces.DataView{}, err)

			metricModelFilters, httpErr := mms.GetMetricModels(testCtx, []string{"0"}, true)
			So(metricModelFilters, ShouldResemble, emptyModelFilters)
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InternalError_GetDataViewQueryFiltersFailed)
		})

		Convey("GetDataViewNameByID Failed", func() {
			err := errors.New("GetDataViewNameByID Failed")
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			mma.EXPECT().GetMetricModelsByModelIDs(gomock.Any(), gomock.Any()).Return([]interfaces.MetricModel{metricModel}, nil)
			mmts.EXPECT().GetMetricTasksByModelIDs(gomock.Any(), gomock.Any()).Return(nil, nil)
			dvs.EXPECT().CheckDataViewExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, err)

			metricModelFilters, httpErr := mms.GetMetricModels(testCtx, []string{"0"}, false)
			So(metricModelFilters, ShouldResemble, emptyModelFilters)
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InternalError_GetDataViewByIDFailed)
		})

		// Convey("Data View not found", func() {
		// 	ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		// 		Return(resrc, nil)
		// 	mma.EXPECT().GetMetricModelsByModelIDs(gomock.Any(), gomock.Any()).Return([]interfaces.MetricModel{metricModel}, nil)
		// 	mmts.EXPECT().GetMetricTasksByModelIDs(gomock.Any(), gomock.Any()).Return(modelTaskMap, nil)
		// 	dvs.EXPECT().CheckDataViewExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
		// 	iba.EXPECT().GetSimpleIndexBasesByTypes(gomock.Any(), gomock.Any()).Return([]interfaces.SimpleIndexBase{{BaseType: "1", Name: "1"}}, nil)

		// 	expectModel := metricModel
		// 	expectModel.DataSource.Name = ""
		// 	expectModelWithFilters := interfaces.MetricModelWithFilters{
		// 		MetricModel: expectModel,
		// 	}

		// 	metricModelFilters, httpErr := mms.GetMetricModels(testCtx, []string{"0"}, false)
		// 	So(metricModelFilters, ShouldResemble, []interfaces.MetricModelWithFilters{expectModelWithFilters})
		// 	So(httpErr, ShouldBeNil)
		// })

		Convey("GetSimpleIndexBasesByTypes failed", func() {
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			mma.EXPECT().GetMetricModelsByModelIDs(gomock.Any(), gomock.Any()).Return([]interfaces.MetricModel{metricModel}, nil)
			mmts.EXPECT().GetMetricTasksByModelIDs(gomock.Any(), gomock.Any()).Return(modelTaskMap, nil)
			// dvs.EXPECT().GetDataView(gomock.Any(), gomock.Any(), gomock.Any()).
			// 	Return(dataViewFilters, nil)
			iba.EXPECT().GetSimpleIndexBasesByTypes(gomock.Any(), gomock.Any()).Return([]interfaces.SimpleIndexBase{}, errors.New("error"))

			expectModelWithFilters := interfaces.MetricModelWithFilters{
				MetricModel: metricModel,
			}
			expectModelWithFilters.DataSource.ID = metricModel.DataSource.ID

			metricModelFilters, httpErr := mms.GetMetricModels(testCtx, []string{"0"}, true)
			So(metricModelFilters, ShouldResemble, []interfaces.MetricModelWithFilters{})
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InternalError_GetSimpleIndexBasesByTypesFailed)
		})

		Convey("GetMetricModelContainFilters success", func() {
			dataView := &interfaces.DataView{SimpleDataView: interfaces.SimpleDataView{ViewName: "数据视图1"}}
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			mma.EXPECT().GetMetricModelsByModelIDs(gomock.Any(), gomock.Any()).Return([]interfaces.MetricModel{metricModel}, nil)
			iba.EXPECT().GetSimpleIndexBasesByTypes(gomock.Any(), gomock.Any()).Return([]interfaces.SimpleIndexBase{{BaseType: "1", Name: "1"}}, nil)
			mmts.EXPECT().GetMetricTasksByModelIDs(gomock.Any(), gomock.Any()).Return(modelTaskMap, nil)
			dvs.EXPECT().GetDataView(gomock.Any(), gomock.Any()).Return(dataView, nil)

			expectModelWithFilters := interfaces.MetricModelWithFilters{
				MetricModel: metricModel,
			}
			expectModelWithFilters.DataView = dataView
			expectModelWithFilters.DataSource.ID = metricModel.DataSource.ID
			expectModelWithFilters.DataViewName = "数据视图1"

			metricModelFilters, httpErr := mms.GetMetricModels(testCtx, []string{"0"}, true)
			So(metricModelFilters, ShouldResemble, []interfaces.MetricModelWithFilters{expectModelWithFilters})
			So(httpErr, ShouldBeNil)
		})

		// Convey("GetVegaViewFieldsByID Failed", func() {
		// 	model := interfaces.MetricModel{
		// 		SimpleMetricModel: interfaces.SimpleMetricModel{
		// 			ModelID:    "0",
		// 			ModelName:  "16",
		// 			MetricType: "atomic",
		// 			QueryType:  "sql",
		// 			Tags:       []string{"a", "s", "s", "s", "s"},
		// 			Comment:    "ssss",
		// 			GroupName:  "sss",
		// 			FormulaConfig: interfaces.SQLConfig{
		// 				AggrExpr: &interfaces.AggrExpr{
		// 					Aggr:  "count",
		// 					Field: "f1",
		// 				},
		// 			},
		// 			DateField: "f2",
		// 		},
		// 		DataSource: &interfaces.MetricDataSource{
		// 			Type: interfaces.QueryType_SQL,
		// 			ID:   "视图1",
		// 		},
		// 	}

		// 	err := errors.New("GetVegaViewFieldsByID Failed")
		// 	ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		// 		Return(resrc, nil)
		// 	mma.EXPECT().GetMetricModelsByModelIDs(gomock.Any(), gomock.Any()).Return([]interfaces.MetricModel{model}, nil)
		// 	mmts.EXPECT().GetMetricTasksByModelIDs(gomock.Any(), gomock.Any()).Return(nil, nil)
		// 	vegaService.EXPECT().GetVegaViewFieldsByID(gomock.Any(), gomock.Any()).Return(interfaces.VegaViewWithFields{}, err)

		// 	metricModelFilters, httpErr := mms.GetMetricModels(testCtx, []string{"0"}, true)
		// 	So(metricModelFilters, ShouldResemble, emptyModelFilters)
		// 	So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
		// 	So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InternalError_GetVegaViewFieldsByIDFailed)
		// })

	})
}

func Test_MetricModelService_CheckFormula(t *testing.T) {
	Convey("Test CheckFormula", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		mma := dmock.NewMockMetricModelAccess(mockCtrl)
		mmga := dmock.NewMockMetricModelGroupAccess(mockCtrl)
		ua := dmock.NewMockUniqueryAccess(mockCtrl)
		mmts := dmock.NewMockMetricModelTaskService(mockCtrl)

		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		mms, _ := MockNewMetricModelService(appSetting, dmja, dvs, mma, mmga, ua, mmts, iba, ps)

		metricModel := interfaces.MetricModel{
			SimpleMetricModel: interfaces.SimpleMetricModel{
				ModelName:  "16",
				MetricType: "atomic",
				QueryType:  "promql",
				Tags:       []string{"a", "s", "s", "s", "s"},
				Comment:    "ssss",
				Formula:    testFormula,
			},
			DataSource: &interfaces.MetricDataSource{
				Type: interfaces.SOURCE_TYPE_DATA_VIEW,
				ID:   "数据视图1",
			},
		}
		Convey("CheckFormulaByUniquery failed", func() {
			err := errors.New("Check Formula By Uniquery failed")
			ua.EXPECT().CheckFormulaByUniquery(gomock.Any(), gomock.Any()).Return(false, "", err)

			httpErr := mms.CheckFormula(testCtx, metricModel)
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InternalError_CheckFormulaFailed)
		})

		Convey("CheckFormulaByUniquery invalid", func() {
			ua.EXPECT().CheckFormulaByUniquery(gomock.Any(), gomock.Any()).
				Return(false, "1:75: parse error: unexpected <by> in aggregation", nil)

			httpErr := mms.CheckFormula(testCtx, metricModel)
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_Formula)
		})

		Convey("CheckFormulaByUniquery valid", func() {
			ua.EXPECT().CheckFormulaByUniquery(gomock.Any(), gomock.Any()).
				Return(true, "", nil)

			httpErr := mms.CheckFormula(testCtx, metricModel)
			So(httpErr, ShouldBeNil)
		})
	})
}

func Test_MetricModelService_CheckSqlFormulaConfig(t *testing.T) {
	Convey("Test CheckSqlFormulaConfig", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		mma := dmock.NewMockMetricModelAccess(mockCtrl)
		mmga := dmock.NewMockMetricModelGroupAccess(mockCtrl)
		ua := dmock.NewMockUniqueryAccess(mockCtrl)
		mmts := dmock.NewMockMetricModelTaskService(mockCtrl)

		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		mms, _ := MockNewMetricModelService(appSetting, dmja, dvs, mma, mmga, ua, mmts, iba, ps)

		metricModel := interfaces.MetricModel{
			SimpleMetricModel: interfaces.SimpleMetricModel{
				ModelName:  "16",
				MetricType: "atomic",
				QueryType:  "sql",
				Tags:       []string{"a", "s", "s", "s", "s"},
				Comment:    "ssss",
				GroupName:  "sss",
				FormulaConfig: interfaces.SQLConfig{
					AggrExpr: &interfaces.AggrExpr{
						Aggr:  "count",
						Field: "f1",
					},
				},
				DateField: "f2",
			},
			DataSource: &interfaces.MetricDataSource{
				Type: interfaces.QueryType_SQL,
				ID:   "视图1",
			},
		}
		Convey("CheckSqlFormulaConfig failed", func() {
			err := errors.New("Check Formula By Uniquery failed")
			ua.EXPECT().CheckFormulaByUniquery(gomock.Any(), gomock.Any()).Return(false, "", err)

			httpErr := mms.CheckSqlFormulaConfig(testCtx, metricModel)
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InternalError_CheckFormulaFailed)
		})

		Convey("CheckFormulaByUniquery invalid", func() {
			ua.EXPECT().CheckFormulaByUniquery(gomock.Any(), gomock.Any()).
				Return(false, "invalid", nil)

			httpErr := mms.CheckSqlFormulaConfig(testCtx, metricModel)
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_Formula)
		})

		Convey("CheckFormulaByUniquery valid", func() {
			ua.EXPECT().CheckFormulaByUniquery(gomock.Any(), gomock.Any()).
				Return(true, "", nil)

			httpErr := mms.CheckSqlFormulaConfig(testCtx, metricModel)
			So(httpErr, ShouldBeNil)
		})
	})
}

func Test_MetricModelService_GetMetricModelSimpleInfosByIDs(t *testing.T) {
	Convey("Test GetMetricModelSimpleInfosByIDs", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		mma := dmock.NewMockMetricModelAccess(mockCtrl)
		mmga := dmock.NewMockMetricModelGroupAccess(mockCtrl)
		ua := dmock.NewMockUniqueryAccess(mockCtrl)
		mmts := dmock.NewMockMetricModelTaskService(mockCtrl)
		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		mms, _ := MockNewMetricModelService(appSetting, dmja, dvs, mma, mmga, ua, mmts, iba, ps)

		resrc := map[string]interfaces.ResourceOps{
			"1": {
				ResourceID: "1",
			},
		}
		Convey("Get succeed", func() {
			expectedModelMap := map[string]interfaces.SimpleMetricModel{"1": {}}
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			mma.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(expectedModelMap, nil)

			modelIDs := []string{"1"}
			modelMap, err := mms.GetMetricModelSimpleInfosByIDs(testCtx, modelIDs)
			So(modelMap, ShouldResemble, expectedModelMap)
			So(err, ShouldBeNil)
		})
	})
}

func Test_MetricModelService_GetMetricModelsByGroupID(t *testing.T) {
	Convey("Test GetMetricModelsByGroupID", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		mma := dmock.NewMockMetricModelAccess(mockCtrl)
		mmga := dmock.NewMockMetricModelGroupAccess(mockCtrl)
		ua := dmock.NewMockUniqueryAccess(mockCtrl)
		mmts := dmock.NewMockMetricModelTaskService(mockCtrl)

		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		mms, _ := MockNewMetricModelService(appSetting, dmja, dvs, mma, mmga, ua, mmts, iba, ps)

		metricModel := interfaces.MetricModel{
			SimpleMetricModel: interfaces.SimpleMetricModel{
				ModelName:    "17",
				MetricType:   "atomic",
				QueryType:    "promql",
				Tags:         []string{"a", "s", "s", "s", "s"},
				Comment:      "ssss",
				GroupID:      "1",
				GroupName:    "group1",
				UnitType:     "storeUnit",
				Unit:         "bit",
				Formula:      "sum(rate(node_cpu_seconds_total{name=~\"node-14-35\",mode12=\"system\"}[5m]))by(name)",
				DateField:    interfaces.PROMQL_DATEFIELD,
				MeasureField: interfaces.PROMQL_METRICFIELD,
			},
			DataSource: &interfaces.MetricDataSource{
				Type: interfaces.SOURCE_TYPE_DATA_VIEW,
				ID:   "数据视图2",
			},
		}

		Convey("GetMetricModelsByGroupID succeed", func() {
			expectedModels := []interfaces.MetricModel{metricModel, metricModel}
			mma.EXPECT().GetMetricModelsByGroupID(gomock.Any(), gomock.Any()).Return(expectedModels, nil)

			models, err := mms.GetMetricModelsByGroupID(testCtx, "1")
			So(models, ShouldResemble, expectedModels)
			So(err, ShouldBeNil)
		})

		Convey("GetMetricModelsByGroupID failed", func() {
			expectedModels := make([]interfaces.MetricModel, 0)
			err := errors.New("Get metric models by groupID failed")
			mma.EXPECT().GetMetricModelsByGroupID(gomock.Any(), gomock.Any()).Return(expectedModels, err)

			models, httpErr := mms.GetMetricModelsByGroupID(testCtx, "1")

			So(models, ShouldResemble, expectedModels)
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InternalError)
		})
	})
}

func Test_MetricModelService_GetMetricModelIDByName(t *testing.T) {
	Convey("Test GetMetricModelIDByName", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		mma := dmock.NewMockMetricModelAccess(mockCtrl)
		mmga := dmock.NewMockMetricModelGroupAccess(mockCtrl)
		ua := dmock.NewMockUniqueryAccess(mockCtrl)
		mmts := dmock.NewMockMetricModelTaskService(mockCtrl)

		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		mms, _ := MockNewMetricModelService(appSetting, dmja, dvs, mma, mmga, ua, mmts, iba, ps)

		Convey("GetMetricModelIDByName succeed", func() {

			mma.EXPECT().GetMetricModelIDByName(gomock.Any(),
				gomock.Any(), gomock.Any()).Return("111", true, nil)

			modelID, err := mms.GetMetricModelIDByName(testCtx, "group1", "model1")
			So(modelID, ShouldEqual, "111")
			So(err, ShouldBeNil)
		})

		Convey("GetMetricModelIDByName failed", func() {
			err := errors.New("Get metric model id  by group name and model name error")
			mma.EXPECT().GetMetricModelIDByName(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, err)

			modelID, httpErr := mms.GetMetricModelIDByName(testCtx, "group1", "model1")

			So(modelID, ShouldEqual, "")
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InternalError_GetModelIDByNameFailed)
		})

		Convey("Metric Model Not Found ", func() {
			mma.EXPECT().GetMetricModelIDByName(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)

			modelID, httpErr := mms.GetMetricModelIDByName(testCtx, "group1", "model1")

			So(modelID, ShouldEqual, "")
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusNotFound)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_MetricModelNotFound)
		})
	})
}

func Test_MetricModelService_UpdateMetricModelsGroupID(t *testing.T) {
	// Convey("Test UpdateMetricModelsGroupID", t, func() {
	// 	mockCtrl := gomock.NewController(t)
	// 	defer mockCtrl.Finish()

	// 	appSetting := &common.AppSetting{}
	// 	dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
	// 	dvs := dmock.NewMockDataViewService(mockCtrl)
	// 	mma := dmock.NewMockMetricModelAccess(mockCtrl)
	// 	mmga := dmock.NewMockMetricModelGroupAccess(mockCtrl)
	// 	ua := dmock.NewMockUniqueryAccess(mockCtrl)
	// 	mmts := dmock.NewMockMetricModelTaskService(mockCtrl)

	// 	iba := dmock.NewMockIndexBaseAccess(mockCtrl)
	// 	vegaService := dmock.NewMockVegaViewService(mockCtrl)
	// 	ps := dmock.NewMockPermissionService(mockCtrl)
	// 	mms, smock := MockNewMetricModelService(appSetting, dmja, dvs, mma, mmga, ua, mmts, iba, vegaService, ps)

	// 	// group := interfaces.MetricModelGroupName{GroupName: "3"}
	// 	modelsMap := make(map[string]interfaces.SimpleMetricModel)
	// 	modelsMap["1"] = interfaces.SimpleMetricModel{ModelID: "1", ModelName: "12"}
	// 	modelsMap["2"] = interfaces.SimpleMetricModel{ModelID: "2", ModelName: "22"}

	// 	metricModel := interfaces.MetricModel{
	// 		SimpleMetricModel: interfaces.SimpleMetricModel{
	// 			ModelName:    "17",
	// 			MetricType:   "atomic",
	// 			QueryType:    "promql",
	// 			Tags:         []string{"a", "s", "s", "s", "s"},
	// 			Comment:      "ssss",
	// 			GroupID:      "1",
	// 			GroupName:    "group1",
	// 			UnitType:     "storeUnit",
	// 			Unit:         "bit",
	// 			Formula:      "sum(rate(node_cpu_seconds_total{name=~\"node-14-35\",mode12=\"system\"}[5m]))by(name)",
	// 			DateField:    interfaces.PROMQL_DATEFIELD,
	// 			MeasureField: interfaces.PROMQL_METRICFIELD,
	// 		},
	// 		DataSource: &interfaces.MetricDataSource{
	// 			Type: interfaces.SOURCE_TYPE_DATA_VIEW,
	// 			ID:   "数据视图2",
	// 		},
	// 	}
	// 	// expectedModels := []interfaces.MetricModel{metricModel}
	// 	// resrc := map[string]interfaces.ResourceOps{
	// 	// 	"0": {
	// 	// 		ResourceID: "0",
	// 	// 	},
	// 	// }
	// 	// Convey("UpdateMetricModelsGroupID succeed", func() {
	// 	// 	ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
	// 	// 		Return(resrc, nil)
	// 	// 	smock.ExpectBegin()
	// 	// 	mmga.EXPECT().GetMetricModelGroupByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(testMetricGroup, true, nil)
	// 	// 	mma.EXPECT().UpdateMetricModelsGroupID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(2), nil)
	// 	// 	mma.EXPECT().GetMetricModelsByGroupID(gomock.Any(), gomock.Any()).Return(expectedModels, nil)
	// 	// 	smock.ExpectCommit()

	// 	// 	rowsAffect, httpErr := mms.UpdateMetricModelsGroup(testCtx, modelsMap, group)
	// 	// 	So(rowsAffect, ShouldEqual, int64(2))
	// 	// 	So(httpErr, ShouldBeNil)
	// 	// })

	// 	// Convey("UpdateMetricModelsGroupID failed", func() {
	// 	// 	ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
	// 	// 		Return(resrc, nil)
	// 	// 	smock.ExpectBegin()
	// 	// 	err := errors.New("UpdateMetricModels error")
	// 	// 	mmga.EXPECT().GetMetricModelGroupByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(testMetricGroup, true, nil)
	// 	// 	mma.EXPECT().GetMetricModelsByGroupID(gomock.Any(), gomock.Any()).Return(expectedModels, nil)
	// 	// 	mma.EXPECT().UpdateMetricModelsGroupID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), err)
	// 	// 	smock.ExpectCommit()

	// 	// 	rowsAffect, httpErr := mms.UpdateMetricModelsGroup(testCtx, modelsMap, group)

	// 	// 	So(rowsAffect, ShouldEqual, int64(0))
	// 	// 	So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
	// 	// 	So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModelGroup_InternalError)
	// 	// })

	// 	// Convey("Update models number not equal request number", func() {
	// 	// 	ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
	// 	// 		Return(resrc, nil)
	// 	// 	smock.ExpectBegin()
	// 	// 	mmga.EXPECT().GetMetricModelGroupByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(testMetricGroup, true, nil)
	// 	// 	mma.EXPECT().GetMetricModelsByGroupID(gomock.Any(), gomock.Any()).Return(expectedModels, nil)
	// 	// 	mma.EXPECT().UpdateMetricModelsGroupID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
	// 	// 	smock.ExpectCommit()

	// 	// 	rowsAffect, httpErr := mms.UpdateMetricModelsGroup(testCtx, modelsMap, group)

	// 	// 	So(rowsAffect, ShouldEqual, int64(1))
	// 	// 	So(httpErr, ShouldBeNil)

	// 	// })

	// })
}

func Test_MetricModelService_RetriveGroupIDByGroupName(t *testing.T) {
	Convey("Test RetriveGroupIDByGroupName", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		mma := dmock.NewMockMetricModelAccess(mockCtrl)
		mmga := dmock.NewMockMetricModelGroupAccess(mockCtrl)
		ua := dmock.NewMockUniqueryAccess(mockCtrl)
		mmts := dmock.NewMockMetricModelTaskService(mockCtrl)

		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		mms, _ := MockNewMetricModelService(appSetting, dmja, dvs, mma, mmga, ua, mmts, iba, ps)

		Convey("RetriveGroupIDByGroupName succeed", func() {
			mmga.EXPECT().GetMetricModelGroupByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(testMetricGroup, true, nil)

			group, httpErr := mms.RetriveGroupIDByGroupName(testCtx, nil, "group111")
			So(group, ShouldResemble, testMetricGroup)
			So(httpErr, ShouldBeNil)
		})

		Convey("GetMetricModelGroupByName failed", func() {
			err := errors.New("GetMetricModelGroupByName error")
			mmga.EXPECT().GetMetricModelGroupByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(testMetricGroup, false, err)

			_, httpErr := mms.RetriveGroupIDByGroupName(testCtx, nil, "group111")
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModelGroup_InternalError_GetMetricModelGroupIDByNameFailed)
		})

		Convey("CreateMetricModelGroup failed", func() {
			err := errors.New("CreateMetricModelGroup error")
			patch := ApplyFuncReturn(did.GenerateDistributedID, 110, nil)
			defer patch.Reset()

			mmga.EXPECT().GetMetricModelGroupByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(testMetricGroup, false, nil)
			mmga.EXPECT().CreateMetricModelGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(err)

			_, httpErr := mms.RetriveGroupIDByGroupName(testCtx, nil, "group111")
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModelGroup_InternalError)
		})

	})
}

func Test_MetricModelService_CheckDuplicateMeasureName(t *testing.T) {
	Convey("Test CheckMetricModelByMeasureName", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		mma := dmock.NewMockMetricModelAccess(mockCtrl)
		mmga := dmock.NewMockMetricModelGroupAccess(mockCtrl)
		ua := dmock.NewMockUniqueryAccess(mockCtrl)
		mmts := dmock.NewMockMetricModelTaskService(mockCtrl)
		iba := dmock.NewMockIndexBaseAccess(mockCtrl)

		ps := dmock.NewMockPermissionService(mockCtrl)
		mms, _ := MockNewMetricModelService(appSetting, dmja, dvs, mma, mmga, ua, mmts, iba, ps)

		Convey("CheckMetricModelByMeasureName failed", func() {
			mma.EXPECT().CheckMetricModelByMeasureName(gomock.Any(), gomock.Any()).Return("", false, errors.New("error"))

			_, exist, err := mms.CheckMetricModelByMeasureName(testCtx, "__m.a")
			So(exist, ShouldBeFalse)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InternalError_CheckDuplicateMeasureNameFailed)
		})

		Convey("CheckMetricModelByMeasureName success", func() {
			mma.EXPECT().CheckMetricModelByMeasureName(gomock.Any(), gomock.Any()).Return("aa", true, nil)

			_, exist, err := mms.CheckMetricModelByMeasureName(testCtx, "__m.a")
			So(exist, ShouldBeTrue)
			So(err, ShouldBeNil)
		})
	})
}

func Test_MetricModelService_CheckVegaLogicView(t *testing.T) {
	// Convey("Test checkVegaLogicView", t, func() {
	// 	mockCtrl := gomock.NewController(t)
	// 	defer mockCtrl.Finish()

	// 	appSetting := &common.AppSetting{}
	// 	dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
	// 	dvs := dmock.NewMockDataViewService(mockCtrl)
	// 	mma := dmock.NewMockMetricModelAccess(mockCtrl)
	// 	mmga := dmock.NewMockMetricModelGroupAccess(mockCtrl)
	// 	ua := dmock.NewMockUniqueryAccess(mockCtrl)
	// 	mmts := dmock.NewMockMetricModelTaskService(mockCtrl)
	// 	iba := dmock.NewMockIndexBaseAccess(mockCtrl)

	// 	vegaService := dmock.NewMockVegaViewService(mockCtrl)
	// 	ps := dmock.NewMockPermissionService(mockCtrl)
	// 	mms, _ := MockNewMetricModelService(appSetting, dmja, dvs, mma, mmga, ua, mmts, iba, vegaService, ps)

	// 	vegaViewFields := interfaces.VegaViewWithFields{
	// 		Fields: []interfaces.VegaViewField{
	// 			{
	// 				Name: "f1",
	// 				Type: "varchar",
	// 			},
	// 			{
	// 				Name: "f2",
	// 				Type: "varchar",
	// 			},
	// 		},
	// 		FieldMap: map[string]interfaces.VegaViewField{
	// 			"f1": {
	// 				Name: "f1",
	// 				Type: "varchar",
	// 			},
	// 			"f2": {
	// 				Name: "f2",
	// 				Type: "varchar",
	// 			},
	// 		},
	// 	}
	// 	model := interfaces.MetricModel{
	// 		SimpleMetricModel: interfaces.SimpleMetricModel{
	// 			ModelName:  "16",
	// 			MetricType: "atomic",
	// 			QueryType:  "sql",
	// 			Tags:       []string{"a", "s", "s", "s", "s"},
	// 			Comment:    "ssss",
	// 			GroupName:  "sss",
	// 			FormulaConfig: interfaces.SQLConfig{
	// 				AggrExpr: &interfaces.AggrExpr{
	// 					Aggr:  "count",
	// 					Field: "f1",
	// 				},
	// 			},
	// 			DateField: "f2",
	// 		},
	// 		DataSource: &interfaces.MetricDataSource{
	// 			Type: interfaces.QueryType_SQL,
	// 			ID:   "视图1",
	// 		},
	// 	}

	// 	// Convey("GetVegaViewFieldsByID failed", func() {

	// 	// 	vegaService.EXPECT().GetVegaViewFieldsByID(gomock.Any(), gomock.Any()).Return(interfaces.VegaViewWithFields{}, fmt.Errorf("error"))

	// 	// 	err := mms.checkSQLView(testCtx, model, nil)
	// 	// 	So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
	// 	// 	So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InternalError_GetVegaViewFieldsByIDFailed)
	// 	// })

	// 	Convey("date field not exist in vega view failed", func() {
	// 		model1 := interfaces.MetricModel{
	// 			SimpleMetricModel: interfaces.SimpleMetricModel{
	// 				ModelName:  "16",
	// 				MetricType: "atomic",
	// 				QueryType:  "sql",
	// 				Tags:       []string{"a", "s", "s", "s", "s"},
	// 				Comment:    "ssss",
	// 				GroupName:  "sss",
	// 				FormulaConfig: interfaces.SQLConfig{
	// 					AggrExpr: &interfaces.AggrExpr{
	// 						Aggr:  "count",
	// 						Field: "f1",
	// 					},
	// 				},
	// 				DateField: "f3",
	// 			},
	// 			DataSource: &interfaces.MetricDataSource{
	// 				Type: interfaces.QueryType_SQL,
	// 				ID:   "视图1",
	// 			},
	// 		}

	// 		vegaService.EXPECT().GetVegaViewFieldsByID(gomock.Any(), gomock.Any()).Return(vegaViewFields, nil)

	// 		err := mms.checkSQLView(testCtx, model1, nil)
	// 		So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
	// 		So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_DateFieldNotExisted)
	// 	})

	// 	Convey("group by field not exist in vega view failed", func() {
	// 		model1 := interfaces.MetricModel{
	// 			SimpleMetricModel: interfaces.SimpleMetricModel{
	// 				ModelName:  "16",
	// 				MetricType: "atomic",
	// 				QueryType:  "sql",
	// 				Tags:       []string{"a", "s", "s", "s", "s"},
	// 				Comment:    "ssss",
	// 				GroupName:  "sss",
	// 				FormulaConfig: interfaces.SQLConfig{
	// 					AggrExpr: &interfaces.AggrExpr{
	// 						Aggr:  "count",
	// 						Field: "f1",
	// 					},
	// 					GroupByFields: []string{"f1", "f3"},
	// 				},
	// 				DateField: "f2",
	// 			},
	// 			DataSource: &interfaces.MetricDataSource{
	// 				Type: interfaces.QueryType_SQL,
	// 				ID:   "视图1",
	// 			},
	// 		}

	// 		vegaService.EXPECT().GetVegaViewFieldsByID(gomock.Any(), gomock.Any()).Return(vegaViewFields, nil)

	// 		err := mms.checkSQLView(testCtx, model1, nil)
	// 		So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
	// 		So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_GroupByFieldNotExisted)
	// 	})

	// 	Convey("aggr field not exist in vega view failed", func() {
	// 		model1 := interfaces.MetricModel{
	// 			SimpleMetricModel: interfaces.SimpleMetricModel{
	// 				ModelName:  "16",
	// 				MetricType: "atomic",
	// 				QueryType:  "sql",
	// 				Tags:       []string{"a", "s", "s", "s", "s"},
	// 				Comment:    "ssss",
	// 				GroupName:  "sss",
	// 				FormulaConfig: interfaces.SQLConfig{
	// 					AggrExpr: &interfaces.AggrExpr{
	// 						Aggr:  "count",
	// 						Field: "f3",
	// 					},
	// 					GroupByFields: []string{"f1"},
	// 				},
	// 				DateField: "f2",
	// 			},
	// 			DataSource: &interfaces.MetricDataSource{
	// 				Type: interfaces.QueryType_SQL,
	// 				ID:   "视图1",
	// 			},
	// 		}

	// 		vegaService.EXPECT().GetVegaViewFieldsByID(gomock.Any(), gomock.Any()).Return(vegaViewFields, nil)

	// 		err := mms.checkSQLView(testCtx, model1, nil)
	// 		So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
	// 		So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_AggregationNotExisted)
	// 	})

	// 	Convey("analysis dimenssion not exist in vega view failed", func() {
	// 		model1 := interfaces.MetricModel{
	// 			SimpleMetricModel: interfaces.SimpleMetricModel{
	// 				ModelName:  "16",
	// 				MetricType: "atomic",
	// 				QueryType:  "sql",
	// 				Tags:       []string{"a", "s", "s", "s", "s"},
	// 				Comment:    "ssss",
	// 				GroupName:  "sss",
	// 				FormulaConfig: interfaces.SQLConfig{
	// 					AggrExpr: &interfaces.AggrExpr{
	// 						Aggr:  "count",
	// 						Field: "f1",
	// 					},
	// 					GroupByFields: []string{"f1"},
	// 				},
	// 				DateField: "f2",
	// 				AnalysisDims: []interfaces.AnalysisDimension{
	// 					{
	// 						Name: "f3",
	// 						Type: "varchar",
	// 					},
	// 				},
	// 			},
	// 			DataSource: &interfaces.MetricDataSource{
	// 				Type: interfaces.QueryType_SQL,
	// 				ID:   "视图1",
	// 			},
	// 		}

	// 		vegaService.EXPECT().GetVegaViewFieldsByID(gomock.Any(), gomock.Any()).Return(vegaViewFields, nil)

	// 		err := mms.checkSQLView(testCtx, model1, nil)
	// 		So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
	// 		So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_AnalysisDimensionNotExisted)
	// 	})

	// 	Convey("CheckMetricModelByMeasureName success", func() {
	// 		vegaService.EXPECT().GetVegaViewFieldsByID(gomock.Any(), gomock.Any()).Return(vegaViewFields, nil)

	// 		err := mms.checkSQLView(testCtx, model, nil)
	// 		So(err, ShouldBeNil)
	// 	})
	// })
}
