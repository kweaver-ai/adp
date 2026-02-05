package metric_model

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/sync/semaphore"

	"uniquery/common"
	cond "uniquery/common/condition"
	"uniquery/common/convert"
	"uniquery/common/value_opt"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	dtype "uniquery/interfaces/data_type"
	mock "uniquery/interfaces/mock"
	"uniquery/logics/promql"
	"uniquery/logics/promql/parser"
	"uniquery/logics/promql/static"
	"uniquery/logics/promql/util"
)

var (
	testENCtx  = context.WithValue(context.Background(), rest.XLangKey, rest.AmericanEnglish)
	fiveMin    = "5m"
	fiveMinInt = int64(300000)
)

var (
	dataview = interfaces.DataView{
		ViewID:        "1",
		QueryType:     interfaces.QueryType_DSL,
		Type:          interfaces.ViewType_Atomic,
		TechnicalName: "index1",
		UpdateTime:    time.Now().Add(-1 * time.Hour).UnixMilli(),
		FieldsMap: map[string]*cond.ViewField{
			"a": {Name: "a", Type: dtype.DataType_Text},
		},
	}

	emptyDslResult = map[string]interface{}{
		"_shards": map[string]interface{}{
			"failed":     0,
			"skipped":    0,
			"successful": 0,
			"total":      0,
		},
		"hits": map[string]interface{}{
			"total": map[string]interface{}{
				"value":    0,
				"relation": "eq",
			},
			"max_score": 0,
			"hits":      []string{},
		},
		"aggregations": map[string]interface{}{
			"labels": map[string]interface{}{
				"doc_count_error_upper_bound": 0,
				"sum_other_doc_count":         0,
				"buckets":                     []string{},
			},
		},
		"timed_out": false,
		"took":      0,
	}

	cardinalityDslResult = map[string]interface{}{
		"_shards": map[string]interface{}{
			"failed":     0,
			"skipped":    0,
			"successful": 0,
			"total":      0,
		},
		"aggregations": map[string]interface{}{
			"labels.cpu": map[string]interface{}{
				"value": 30000,
			},
		},
	}

	seriesDslResult = map[string]interface{}{
		"aggregations": map[string]interface{}{
			"cpu": map[string]interface{}{
				"buckets": []map[string]interface{}{
					{
						"key":       "0",
						"doc_count": 11592604,
					},
					{
						"key":       "1",
						"doc_count": 11592604,
					},
					{
						"key":       "2",
						"doc_count": 11592604,
					},
				},
			},
		},
	}

	// queryParam = interfaces.MetricModelQueryParameters{
	// 	IncludeModel: true,
	// }

	// vegaViewFields = interfaces.VegaViewWithFields{
	// 	Fields: []interfaces.VegaViewField{
	// 		{
	// 			Name: "f1",
	// 			Type: "varchar",
	// 		},
	// 		{
	// 			Name: "f2",
	// 			Type: "varchar",
	// 		},
	// 	},
	// 	VegaFieldMap: map[string]interfaces.VegaViewField{
	// 		"f1": {
	// 			Name: "f1",
	// 			Type: "varchar",
	// 		},
	// 		"f2": {
	// 			Name: "f2",
	// 			Type: "varchar",
	// 		},
	// 	},
	// }
)

func MockNewMetricModelService(appSetting *common.AppSetting, mmAccess interfaces.MetricModelAccess,
	ibAccess interfaces.IndexBaseAccess, pService interfaces.PromQLService, dvService interfaces.DataViewService,
	vvs interfaces.VegaService, ps interfaces.PermissionService) *metricModelService {
	return &metricModelService{
		sem:           semaphore.NewWeighted(int64(appSetting.PoolSetting.ExecutePoolSize)),
		appSetting:    appSetting,
		mmAccess:      mmAccess,
		ibAccess:      ibAccess,
		promqlService: pService,
		dvService:     dvService,
		vvs:           vvs,
		ps:            ps,
	}
}

func TestSimulate(t *testing.T) {
	Convey("Test Simulate", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				FullCacheRefreshInterval: 24 * time.Hour,
			},
		}
		loc, _ := time.LoadLocation(os.Getenv("TZ"))
		common.APP_LOCATION = loc

		mmaMock := mock.NewMockMetricModelAccess(mockCtrl)
		ibaMock := mock.NewMockIndexBaseAccess(mockCtrl)
		promQLMock := mock.NewMockPromQLService(mockCtrl)
		dvsMock := mock.NewMockDataViewService(mockCtrl)
		vvsMock := mock.NewMockVegaService(mockCtrl)
		psMock := mock.NewMockPermissionService(mockCtrl)
		ds := MockNewMetricModelService(appSetting, mmaMock, ibaMock, promQLMock, dvsMock, vvsMock, psMock)

		ops := []interfaces.ResourceOps{
			{
				ResourceID: interfaces.RESOURCE_ID_ALL,
				Operations: []string{interfaces.OPERATION_TYPE_CREATE},
			},
		}
		Convey("Simulate failed, metric type or query type", func() {

			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusBadRequest,
				Language: "en-US",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_MetricModel_UnsupportQuery,
					Description:  "Unsupported Query",
					Solution:     "Please check whether the parameter is correct.",
					ErrorLink:    "None",
					ErrorDetails: "Unsupport metric type abc",
				},
			}

			psMock.EXPECT().GetResourcesOperations(gomock.Any(), gomock.Any(), gomock.Any()).Return(ops, nil)
			_, httpErr := ds.Simulate(testENCtx, interfaces.MetricModelQuery{
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "123",
				},
				MetricType: "abc", QueryType: "promql"})
			So(httpErr, ShouldResemble, expectedErr)
		})

		Convey("Simulate failed, because Simulate promql failed", func() {
			dvsMock.EXPECT().GetDataViewByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(
				&interfaces.DataView{}, nil)

			err := errors.New("Simulate promql failed")
			psMock.EXPECT().GetResourcesOperations(gomock.Any(), gomock.Any(), gomock.Any()).Return(ops, nil)
			promQLMock.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(interfaces.PromQLResponse{}, nil, 0, err)
			dvsMock.EXPECT().BuildViewQuery4MetricModel(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(interfaces.ViewQuery4Metric{}, nil)

			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: "en-US",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_MetricModel_InternalError_ExecPromQLFailed,
					Description:  "PromQL Execution Failed",
					Solution:     "Please try this operation again, if the error occurs again, submit the work order or contact technical support engineers.",
					ErrorLink:    "None",
					ErrorDetails: err.Error(),
				},
			}

			_, httpErr := ds.Simulate(testENCtx, interfaces.MetricModelQuery{
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "123",
				},
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &start,
					End:     &end,
					StepStr: &step_1min,
					Step:    &step_60000,
				},
				MetricType: "atomic", QueryType: "promql"})
			So(httpErr, ShouldResemble, expectedErr)
		})

		Convey("Simulate failed, because Simulate GetDataViewByID failed", func() {
			psMock.EXPECT().GetResourcesOperations(gomock.Any(), gomock.Any(), gomock.Any()).Return(ops, nil)
			dvsMock.EXPECT().GetDataViewByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(
				&interfaces.DataView{}, fmt.Errorf("missing ar_dataview parameter or ar_dataview parameter value cannot be empty."))

			expectedErr := fmt.Errorf("missing ar_dataview parameter or ar_dataview parameter value cannot be empty.")

			_, httpErr := ds.Simulate(testENCtx, interfaces.MetricModelQuery{
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "123",
				},
				MetricType: "atomic", QueryType: "promql"})
			So(httpErr, ShouldResemble, expectedErr)
		})

		Convey("Simulate failed, because Simulate data view not found", func() {
			psMock.EXPECT().GetResourcesOperations(gomock.Any(), gomock.Any(), gomock.Any()).Return(ops, nil)
			dvsMock.EXPECT().GetDataViewByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(
				&interfaces.DataView{}, fmt.Errorf("data view not found."))

			_, httpErr := ds.Simulate(testENCtx, interfaces.MetricModelQuery{
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "123",
				},
				MetricType: "atomic", QueryType: "promql"})
			So(httpErr, ShouldNotBeNil)
		})

		// Convey("Simulate failed, because Simulate GetVegaViewFieldsByID failed", func() {
		// 	psMock.EXPECT().GetResourcesOperations(gomock.Any(), gomock.Any(), gomock.Any()).Return(ops, nil)
		// 	vvsMock.EXPECT().GetVegaViewFieldsByID(gomock.Any(), gomock.Any()).Return(interfaces.VegaViewWithFields{}, fmt.Errorf("error"))

		// 	_, httpErr := ds.Simulate(testENCtx, interfaces.MetricModelQuery{
		// 		DataSource: &interfaces.MetricDataSource{
		// 			Type: interfaces.QueryType_SQL,
		// 			ID:   "123",
		// 		},
		// 		MetricType: "atomic", QueryType: "sql"})
		// 	So(httpErr, ShouldNotBeNil)
		// 	So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InternalError_GetVegaViewFieldsByIDFailed)
		// })

		// Convey("Simulate failed, because date field not exists in vega view", func() {
		// 	psMock.EXPECT().GetResourcesOperations(gomock.Any(), gomock.Any(), gomock.Any()).Return(ops, nil)
		// 	vvsMock.EXPECT().GetVegaViewFieldsByID(gomock.Any(), gomock.Any()).Return(vegaViewFields, nil)

		// 	_, httpErr := ds.Simulate(testENCtx, interfaces.MetricModelQuery{
		// 		DataSource: &interfaces.MetricDataSource{
		// 			Type: interfaces.QueryType_SQL,
		// 			ID:   "123",
		// 		},
		// 		MetricType: "atomic",
		// 		QueryType:  "sql",
		// 		DateField:  "f4",
		// 	})
		// 	So(httpErr, ShouldNotBeNil)
		// 	So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_DateFieldNotExisted)
		// })

		// Convey("Simulate failed, because group by field not exists in vega view", func() {
		// 	psMock.EXPECT().GetResourcesOperations(gomock.Any(), gomock.Any(), gomock.Any()).Return(ops, nil)
		// 	vvsMock.EXPECT().GetVegaViewFieldsByID(gomock.Any(), gomock.Any()).Return(vegaViewFields, nil)

		// 	_, httpErr := ds.Simulate(testENCtx, interfaces.MetricModelQuery{
		// 		DataSource: &interfaces.MetricDataSource{
		// 			Type: interfaces.QueryType_SQL,
		// 			ID:   "123",
		// 		},
		// 		MetricType: "atomic",
		// 		QueryType:  "sql",
		// 		DateField:  "f1",
		// 		FormulaConfig: interfaces.SQLConfig{
		// 			GroupByFields: []string{"f4"},
		// 		},
		// 	})
		// 	So(httpErr, ShouldNotBeNil)
		// 	So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_GroupByFieldNotExisted)
		// })

		// Convey("Simulate failed, because aggr field not exists in vega view", func() {
		// 	psMock.EXPECT().GetResourcesOperations(gomock.Any(), gomock.Any(), gomock.Any()).Return(ops, nil)
		// 	vvsMock.EXPECT().GetVegaViewFieldsByID(gomock.Any(), gomock.Any()).Return(vegaViewFields, nil)

		// 	_, httpErr := ds.Simulate(testENCtx, interfaces.MetricModelQuery{
		// 		DataSource: &interfaces.MetricDataSource{
		// 			Type: interfaces.QueryType_SQL,
		// 			ID:   "123",
		// 		},
		// 		MetricType: "atomic",
		// 		QueryType:  "sql",
		// 		DateField:  "f1",
		// 		FormulaConfig: interfaces.SQLConfig{
		// 			AggrExpr: &interfaces.AggrExpr{
		// 				Aggr:  "avg",
		// 				Field: "f4",
		// 			},
		// 		},
		// 	})
		// 	So(httpErr, ShouldNotBeNil)
		// 	So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_AggregationNotExisted)
		// })

		// Convey("Simulate success", func() {
		// 	psMock.EXPECT().GetResourcesOperations(gomock.Any(), gomock.Any(), gomock.Any()).Return(ops, nil)
		// 	dvsMock.EXPECT().GetDataViewQueryFiltersAndFields(gomock.Any(), gomock.Any()).Return(
		// 		&interfaces.DataView{}, true, nil)

		// 	expectResult := interfaces.PromQLResponse{
		// 		Status: "success",
		// 		Data: promql.QueryData{
		// 			Result: static.Matrix{},
		// 		},
		// 	}
		// 	promQLMock.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(expectResult, nil, 0, nil)

		// 	result, httpErr := ds.Simulate(testENCtx, interfaces.MetricModelQuery{
		// 		DataSource: &interfaces.MetricDataSource{
		// 			Type: interfaces.SOURCE_TYPE_DATA_VIEW,
		// 			ID:   "123",
		// 		},
		// 		MetricType: "atomic", QueryType: "promql", MetricModelQueryParameters: queryParam})
		// 	So(result, ShouldResemble, interfaces.MetricModelUniResponse{
		// 		Model: interfaces.MetricModel{},
		// 		Datas: []interfaces.MetricModelData{},
		// 		Step:  "",
		// 	})
		// 	So(httpErr, ShouldBeNil)
		// })

		// Convey("Simulate success with instant query && request from model", func() {
		// 	psMock.EXPECT().GetResourcesOperations(gomock.Any(), gomock.Any(), gomock.Any()).Return(ops, nil)
		// 	dvsMock.EXPECT().GetDataViewQueryFiltersAndFields(gomock.Any(), gomock.Any()).Return(
		// 		&interfaces.DataView{}, true, nil)

		// 	expectResult := interfaces.PromQLResponse{
		// 		Status: "success",
		// 		Data: promql.QueryData{
		// 			Result: static.Matrix{},
		// 		},
		// 	}
		// 	promQLMock.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(expectResult, nil, 0, nil)

		// 	result, httpErr := ds.Simulate(testENCtx,
		// 		interfaces.MetricModelQuery{
		// 			DataSource: &interfaces.MetricDataSource{
		// 				Type: interfaces.SOURCE_TYPE_DATA_VIEW,
		// 				ID:   "123",
		// 			},
		// 			MetricType:                 "atomic",
		// 			QueryType:                  "promql",
		// 			IsModelRequest:             true,
		// 			QueryTimeParams:            interfaces.QueryTimeParams{IsInstantQuery: true},
		// 			MetricModelQueryParameters: queryParam})
		// 	So(result, ShouldResemble, interfaces.MetricModelUniResponse{
		// 		Model: interfaces.MetricModel{},
		// 		Datas: []interfaces.MetricModelData{},
		// 		Step:  "",
		// 	})
		// 	So(httpErr, ShouldBeNil)
		// })
	})
}

func TestEval(t *testing.T) {
	Convey("Test eval", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				FullCacheRefreshInterval: 24 * time.Hour,
			},
		}
		loc, _ := time.LoadLocation(os.Getenv("TZ"))
		common.APP_LOCATION = loc

		mmaMock := mock.NewMockMetricModelAccess(mockCtrl)
		ibaMock := mock.NewMockIndexBaseAccess(mockCtrl)
		promQLMock := mock.NewMockPromQLService(mockCtrl)
		dvsMock := mock.NewMockDataViewService(mockCtrl)
		vvsMock := mock.NewMockVegaService(mockCtrl)
		psMock := mock.NewMockPermissionService(mockCtrl)
		ds := MockNewMetricModelService(appSetting, mmaMock, ibaMock, promQLMock, dvsMock, vvsMock, psMock)

		Convey("eval failed, metric type or query type", func() {

			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusBadRequest,
				Language: "en-US",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_MetricModel_UnsupportQuery,
					Description:  "Unsupported Query",
					Solution:     "Please check whether the parameter is correct.",
					ErrorLink:    "None",
					ErrorDetails: "Unsupport metric type abc",
				},
			}

			_, httpErr := ds.eval(testENCtx, interfaces.MetricModelQuery{
				MetricType: "abc",
				QueryType:  "promql",
				DataSource: &interfaces.MetricDataSource{},
			}, interfaces.MetricModel{}, &interfaces.DataView{})
			So(httpErr, ShouldResemble, expectedErr)
		})

		Convey("eval failed, because exec promql failed", func() {

			err := errors.New("exec promql failed")
			promQLMock.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(interfaces.PromQLResponse{}, nil, 0, err)
			dvsMock.EXPECT().BuildViewQuery4MetricModel(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(interfaces.ViewQuery4Metric{}, nil)

			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: "en-US",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_MetricModel_InternalError_ExecPromQLFailed,
					Description:  "PromQL Execution Failed",
					Solution:     "Please try this operation again, if the error occurs again, submit the work order or contact technical support engineers.",
					ErrorLink:    "None",
					ErrorDetails: err.Error(),
				},
			}

			_, httpErr := ds.eval(testENCtx, interfaces.MetricModelQuery{
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &start,
					End:     &end,
					StepStr: &step_1min,
					Step:    &step_60000,
				},
				MetricType: "atomic", QueryType: "promql",
				DataSource: &interfaces.MetricDataSource{}}, interfaces.MetricModel{}, &interfaces.DataView{})
			So(httpErr, ShouldResemble, expectedErr)
		})

		// Convey("eval success when range query", func() {

		// 	expectResult := interfaces.PromQLResponse{
		// 		Status: "success",
		// 		Data: promql.QueryData{
		// 			Result: static.Matrix{},
		// 		},
		// 	}
		// 	promQLMock.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(expectResult, nil, 0, nil)

		// 	result, httpErr := ds.eval(testENCtx, interfaces.MetricModelQuery{MetricType: "atomic", QueryType: "promql",
		// 		MetricModelQueryParameters: queryParam,
		// 		DataSource:                 &interfaces.MetricDataSource{},
		// 	}, interfaces.MetricModel{})
		// 	So(result, ShouldResemble, interfaces.MetricModelUniResponse{
		// 		Model: interfaces.MetricModel{},
		// 		Datas: []interfaces.MetricModelData{},
		// 		Step:  "",
		// 	})
		// 	So(httpErr, ShouldBeNil)
		// })

		// Convey("eval success when instant query", func() {

		// 	expectResult := interfaces.PromQLResponse{
		// 		Status: "success",
		// 		Data: promql.QueryData{
		// 			Result: static.Matrix{},
		// 		},
		// 	}
		// 	promQLMock.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(expectResult, nil, 0, nil)

		// 	result, httpErr := ds.eval(testENCtx, interfaces.MetricModelQuery{MetricType: "atomic", QueryType: "promql",
		// 		MetricModelQueryParameters: queryParam,
		// 		DataSource:                 &interfaces.MetricDataSource{},
		// 	},
		// 		interfaces.MetricModel{})
		// 	So(result, ShouldResemble, interfaces.MetricModelUniResponse{
		// 		Model: interfaces.MetricModel{},
		// 		Datas: []interfaces.MetricModelData{},
		// 		Step:  "",
		// 	})
		// 	So(httpErr, ShouldBeNil)
		// })

		// Convey("eval success when instant query with filters_mode is error", func() {
		// 	result, httpErr := ds.eval(testENCtx, interfaces.MetricModelQuery{
		// 		MetricType: "atomic",
		// 		QueryType:  "promql",
		// 		Filters: []interfaces.Filter{
		// 			{
		// 				Name:      "fieldNotExist",
		// 				Value:     "b",
		// 				Operation: interfaces.OPERATION_EQ,
		// 			},
		// 		},
		// 		DataSource: &interfaces.MetricDataSource{},
		// 		MetricModelQueryParameters: interfaces.MetricModelQueryParameters{
		// 			IncludeModel: true,
		// 			FilterMode:   interfaces.FILTER_MODE_ERROR,
		// 		},
		// 	},
		// 		interfaces.MetricModel{})
		// 	So(result, ShouldResemble, interfaces.MetricModelUniResponse{
		// 		Model: interfaces.MetricModel{},
		// 		Datas: []interfaces.MetricModelData{},
		// 		Step:  "",
		// 	})
		// 	So(httpErr, ShouldNotBeNil)
		// })

		// Convey("eval success when instant query with filters_mode is ignore", func() {
		// 	expectResult := interfaces.PromQLResponse{
		// 		Status: "success",
		// 		Data: promql.QueryData{
		// 			Result: static.Matrix{},
		// 		},
		// 	}
		// 	promQLMock.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(expectResult, nil, 0, nil)
		// 	result, httpErr := ds.eval(testENCtx, interfaces.MetricModelQuery{
		// 		MetricType: "atomic",
		// 		QueryType:  "promql",
		// 		Filters: []interfaces.Filter{
		// 			{
		// 				Name:      "fieldNotExist",
		// 				Value:     "b",
		// 				Operation: interfaces.OPERATION_EQ,
		// 			},
		// 		},
		// 		DataSource: &interfaces.MetricDataSource{},
		// 		MetricModelQueryParameters: interfaces.MetricModelQueryParameters{
		// 			IncludeModel: true,
		// 			FilterMode:   interfaces.FILTER_MODE_IGNORE,
		// 		},
		// 	},
		// 		interfaces.MetricModel{})
		// 	So(result, ShouldResemble, interfaces.MetricModelUniResponse{
		// 		Model: interfaces.MetricModel{},
		// 		Datas: []interfaces.MetricModelData{},
		// 		Step:  "",
		// 	})
		// 	So(httpErr, ShouldBeNil)
		// })

		// Convey("eval success when instant query with filters_mode is Normal", func() {
		// 	expectResult := interfaces.PromQLResponse{
		// 		Status: "success",
		// 		Data: promql.QueryData{
		// 			Result: static.Matrix{},
		// 		},
		// 	}
		// 	promQLMock.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(expectResult, nil, 0, nil)
		// 	result, httpErr := ds.eval(testENCtx, interfaces.MetricModelQuery{
		// 		MetricType: "atomic",
		// 		QueryType:  "promql",
		// 		Filters: []interfaces.Filter{
		// 			{
		// 				Name:      "fieldNotExist",
		// 				Value:     "b",
		// 				Operation: interfaces.OPERATION_EQ,
		// 			},
		// 		},
		// 		DataSource: &interfaces.MetricDataSource{},
		// 		MetricModelQueryParameters: interfaces.MetricModelQueryParameters{
		// 			IncludeModel: true,
		// 			FilterMode:   interfaces.FILTER_MODE_NORMAL,
		// 		},
		// 	},
		// 		interfaces.MetricModel{})
		// 	So(result, ShouldResemble, interfaces.MetricModelUniResponse{
		// 		Model: interfaces.MetricModel{},
		// 		Datas: []interfaces.MetricModelData{},
		// 		Step:  "",
		// 	})
		// 	So(httpErr, ShouldBeNil)
		// })

		Convey("eval dsl failed, measure field is not a number", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusBadRequest,
				Language: "en-US",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_MetricModel_InvalidParameter_MeasureField,
					Description:  "Invalid Measure Field for DSL Query",
					Solution:     "Please check whether the parameter is correct.",
					ErrorLink:    "None",
					ErrorDetails: "The type of measure field must be number.",
				},
			}

			mdQuery := interfaces.MetricModelQuery{
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &start,
					End:     &end,
					StepStr: &step_1min,
					Step:    &step_60000,
				},
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.DSL,
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "123",
				},
				MeasureField:   "a",
				ContainTopHits: true,
				Formula:        `"a": {{__interval}}`,
			}
			dataView := &interfaces.DataView{
				FieldsMap: map[string]*cond.ViewField{
					"a": {Name: "a", Type: dtype.DataType_String},
				},
			}

			dvsMock.EXPECT().BuildViewQuery4MetricModel(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(interfaces.ViewQuery4Metric{}, nil)

			_, httpErr := ds.eval(testENCtx, mdQuery, interfaces.MetricModel{}, dataView)
			So(httpErr, ShouldResemble, expectedErr)
		})

		Convey("eval failed, dataviewid is not empty and parseDslQuery failed", func() {
			dvsMock.EXPECT().BuildViewQuery4MetricModel(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(interfaces.ViewQuery4Metric{}, nil)

			_, httpErr := ds.eval(testENCtx,
				interfaces.MetricModelQuery{
					QueryTimeParams: interfaces.QueryTimeParams{
						Start:   &start,
						End:     &end,
						StepStr: &step_1min,
						Step:    &step_60000,
					},
					MetricType: interfaces.ATOMIC_METRIC,
					QueryType:  interfaces.DSL,
					DataSource: &interfaces.MetricDataSource{
						Type: interfaces.SOURCE_TYPE_DATA_VIEW,
						ID:   "123",
					}, Formula: `"a": {{__interval}}`},
				interfaces.MetricModel{},
				&interfaces.DataView{})
			So(httpErr, ShouldNotBeNil)
		})

		Convey("eval failed, because of GetDataFromOpenSearch err", func() {
			indices := []string{"a", "b"}
			// dvsMock.EXPECT().GetIndices(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, indices, 200, nil)
			dvsMock.EXPECT().GetDataFromOpenSearch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(nil, http.StatusInternalServerError,
				uerrors.NewOpenSearchError(uerrors.InternalServerError).WithReason("Error getting response from opensearch"))
			dvsMock.EXPECT().BuildViewQuery4MetricModel(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(interfaces.ViewQuery4Metric{
					BaseTypes: []string{"a", "b"},
					Indices:   indices,
				}, nil)

			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: "en-US",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_InternalError,
					Description:  "Internal Error",
					Solution:     "Please try this operation again, if the error occurs again, submit the work order or contact technical support engineers.",
					ErrorLink:    "None",
					ErrorDetails: "{\"status\":500,\"error\":{\"type\":\"UniQuery.InternalServerError\",\"reason\":\"Error getting response from opensearch\"}}",
				},
			}

			_, httpErr := ds.eval(testENCtx, interfaces.MetricModelQuery{
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &start,
					End:     &end,
					StepStr: &step_1min,
					Step:    &step_60000,
				},
				MetricType: interfaces.ATOMIC_METRIC, QueryType: interfaces.DSL,
				Formula: `{"size":0,"aggs": {
					"NAME1": {
					  "terms": {
						"field": "labels.cpu.keyword",
						"size": 10
					  },
					  "aggs": {
						"NAME2": {
						  "date_histogram": {
							"field": "@timestamp",
							"fixed_interval": "1d"
						  },
						  "aggs": {
							"NAME3": {
							  "value_count": {
								"field": "1"
							  }
							}
						  }
						}
					  }
					}
				  }}`,
				MeasureField: "NAME3",
				DataSource:   &interfaces.MetricDataSource{},
			},
				interfaces.MetricModel{}, &dataview)
			So(httpErr, ShouldResemble, expectedErr)
		})

		// Convey("eval success when dsl range", func() {
		// 	indices := []string{"a", "b"}
		// 	dvsMock.EXPECT().GetIndices(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, indices, 200, nil)
		// 	dvsMock.EXPECT().GetDataFromOpenSearch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
		// 		gomock.Any(), gomock.Any()).AnyTimes().Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

		// 	result, httpErr := ds.eval(testENCtx,
		// 		interfaces.MetricModelQuery{MetricType: interfaces.ATOMIC_METRIC, QueryType: interfaces.DSL,
		// 			DataView: dataViewQueryFilters, MetricModelQueryParameters: queryParam, Formula: `{"size":0,"aggs": {
		// 			"NAME1": {
		// 			  "terms": {
		// 				"field": "labels.cpu.keyword",
		// 				"size": 10
		// 			  },
		// 			  "aggs": {
		// 				"NAME2": {
		// 				  "date_histogram": {
		// 					"field": "@timestamp",
		// 					"fixed_interval": "1d"
		// 				  },
		// 				  "aggs": {
		// 					"NAME3": {
		// 					  "value_count": {
		// 						"field": "1"
		// 					  }
		// 					}
		// 				  }
		// 				}
		// 			  }
		// 			}
		// 		  }}`,
		// 			MeasureField: "NAME3",
		// 			DataSource:   &interfaces.MetricDataSource{},
		// 		},
		// 		interfaces.MetricModel{})
		// 	So(result, ShouldResemble, interfaces.MetricModelUniResponse{
		// 		Model: interfaces.MetricModel{},
		// 		Datas: []interfaces.MetricModelData{},
		// 		Step:  "1d",
		// 	})
		// 	So(httpErr, ShouldBeNil)
		// })

		// Convey("eval success when dsl instant && is model request", func() {
		// 	indices := []string{"a", "b"}
		// 	dvsMock.EXPECT().GetIndices(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, indices, 200, nil)
		// 	dvsMock.EXPECT().GetDataFromOpenSearch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
		// 		gomock.Any(), gomock.Any()).AnyTimes().Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

		// 	result, httpErr := ds.eval(testENCtx,
		// 		interfaces.MetricModelQuery{MetricType: interfaces.ATOMIC_METRIC, QueryType: interfaces.DSL,
		// 			DataView: dataViewQueryFilters, IsModelRequest: true, QueryTimeParams: interfaces.QueryTimeParams{IsInstantQuery: true},
		// 			MetricModelQueryParameters: queryParam, Formula: `{"size":0,"aggs": {
		// 			"NAME1": {
		// 			  "terms": {
		// 				"field": "labels.cpu.keyword",
		// 				"size": 10
		// 			  },
		// 			  "aggs": {
		// 				"NAME2": {
		// 				  "date_histogram": {
		// 					"field": "@timestamp",
		// 					"fixed_interval": "1d"
		// 				  },
		// 				  "aggs": {
		// 					"NAME3": {
		// 					  "value_count": {
		// 						"field": "1"
		// 					  }
		// 					}
		// 				  }
		// 				}
		// 			  }
		// 			}
		// 		  }}`,
		// 			MeasureField: "NAME3",
		// 			DataSource:   &interfaces.MetricDataSource{},
		// 		},
		// 		interfaces.MetricModel{})
		// 	So(result, ShouldResemble, interfaces.MetricModelUniResponse{
		// 		Model: interfaces.MetricModel{},
		// 		Datas: []interfaces.MetricModelData{},
		// 		Step:  "",
		// 	})
		// 	So(httpErr, ShouldBeNil)
		// })

		Convey("eval success when dsl range with batch", func() {
			indices := []string{"a", "b"}
			// dvsMock.EXPECT().GetIndices(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, indices, 200, nil)
			dvsMock.EXPECT().GetDataFromOpenSearch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)
			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(convert.MapToByte(cardinalityDslResult), http.StatusOK, nil)
			dvsMock.EXPECT().BuildViewQuery4MetricModel(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(interfaces.ViewQuery4Metric{
					BaseTypes: []string{"a", "b"},
					Indices:   indices,
				}, nil)

			result, httpErr := ds.eval(testENCtx,
				interfaces.MetricModelQuery{
					QueryTimeParams: interfaces.QueryTimeParams{
						Start:   &start,
						End:     &end,
						StepStr: &step_1min,
						Step:    &step_60000,
					},
					MetricType: interfaces.ATOMIC_METRIC, QueryType: interfaces.DSL,
					MetricModelQueryParameters: interfaces.MetricModelQueryParameters{
						IncludeModel:        false,
						IgnoringMemoryCache: true,
					}, Formula: `{"size":0,"aggs": {
					"NAME1": {
					  "terms": {
						"field": "labels.cpu",
						"size": 20000
					  },
					  "aggs": {
						"NAME2": {
						  "date_histogram": {
							"field": "@timestamp",
							"fixed_interval": "1d"
						  },
						  "aggs": {
							"NAME3": {
							  "value_count": {
								"field": "1"
							  }
							}
						  }
						}
					  }
					}
				  }}`,
					MeasureField: "NAME3",
					DataSource:   &interfaces.MetricDataSource{},
				},
				interfaces.MetricModel{}, &dataview)
			So(result, ShouldResemble, interfaces.MetricModelUniResponse{Datas: []interfaces.MetricModelData{}, Step: &step_1min})
			So(httpErr, ShouldBeNil)
		})

		Convey("eval failed, because of sql's query step is invalid", func() {
			dvsMock.EXPECT().BuildViewQuery4MetricModel(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(interfaces.ViewQuery4Metric{
					BaseTypes: []string{"a", "b"},
					Indices:   []string{"a", "b"},
				}, nil)
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusBadRequest,
				Language: "en-US",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_MetricModel_InvalidParameter_Step,
					Description:  "Invalid Step",
					Solution:     "Please check whether the parameter is correct.",
					ErrorLink:    "None",
					ErrorDetails: "sql atomic metric only support calendar steps, expect steps is one of {[minute hour day week month quarter year]}, actaul is 1d",
				},
			}

			d1 := "1d"
			_, httpErr := ds.eval(testENCtx, interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.SQL,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &start,
					End:     &end,
					StepStr: &d1,
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.QueryType_SQL,
					ID:   "1",
				},
				MeasureField: "NAME3"},
				interfaces.MetricModel{}, &interfaces.DataView{})
			So(httpErr, ShouldResemble, expectedErr)
		})

		Convey("eval failed, because of generateSQL failed when sql", func() {
			dvsMock.EXPECT().BuildViewQuery4MetricModel(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(interfaces.ViewQuery4Metric{}, nil)

			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusBadRequest,
				Language: "en-US",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_DataView_InvalidParameter_Filters,
					Description:  "Invalid Filters",
					Solution:     "Please check whether the parameter is correct.",
					ErrorLink:    "None",
					ErrorDetails: "New condition failed, condition config field name '' must in view original fields",
				},
			}

			_, httpErr := ds.eval(testENCtx, interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.SQL,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &start,
					End:     &end,
					StepStr: &day,
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.QueryType_SQL,
					ID:   "1",
				},
				FormulaConfig: interfaces.SQLConfig{
					Condition: &cond.CondCfg{
						Operation: "eq",
						ValueOptCfg: value_opt.ValueOptCfg{
							ValueFrom: "const",
							Value:     "1",
						},
					},
				},
				MeasureField: "NAME3"},
				interfaces.MetricModel{}, &interfaces.DataView{})
			So(httpErr, ShouldResemble, expectedErr)
		})

		Convey("eval failed, because of FetchDatasFromVega failed when sql", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: "en-US",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_MetricModel_InternalError_FetchDatasFromVegaFailed,
					Description:  "Fetch Datas From Vega Failed",
					Solution:     "Please try this operation again, if the error occurs again, submit the work order or contact technical support engineers.",
					ErrorLink:    "None",
					ErrorDetails: "method failed",
				},
			}

			vvsMock.EXPECT().FetchDatasFromVega(gomock.Any(), gomock.Any()).Return(interfaces.VegaFetchData{}, fmt.Errorf("method failed"))
			dvsMock.EXPECT().BuildViewQuery4MetricModel(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(interfaces.ViewQuery4Metric{}, nil)

			_, httpErr := ds.eval(testENCtx, interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.SQL,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &start,
					End:     &end,
					StepStr: &day,
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.QueryType_SQL,
					ID:   "1",
				},
				FormulaConfig: interfaces.SQLConfig{
					AggrExprStr: "avg(f1)",
				},
				DateField:    "f2",
				MeasureField: "NAME3"},
				interfaces.MetricModel{}, &interfaces.DataView{})
			So(httpErr, ShouldResemble, expectedErr)
		})

		Convey("eval success when sql", func() {
			vegaData := interfaces.VegaFetchData{
				Data: [][]any{
					{"a", "2024-02-01", 2.2},
					{"b", "2024-02-01", 3.2},
				},
				Columns: []cond.ViewField{
					{
						Name: "f1",
						Type: "varchar",
					},
					{
						Name: "__time",
						Type: "varchar",
					},
					{
						Name: "__value",
						Type: "float",
					},
				},
			}

			vvsMock.EXPECT().FetchDatasFromVega(gomock.Any(), gomock.Any()).AnyTimes().Return(vegaData, nil)
			dvsMock.EXPECT().BuildViewQuery4MetricModel(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(interfaces.ViewQuery4Metric{}, nil)

			_, httpErr := ds.eval(testENCtx, interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.SQL,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &start,
					End:     &end,
					StepStr: &day,
					Step:    &step_60000,
				},
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.QueryType_SQL,
					ID:   "1",
				},
				FormulaConfig: interfaces.SQLConfig{
					AggrExprStr: "avg(f1)",
				},
				DateField: "f1",
				RequestMetrics: &interfaces.RequestMetrics{
					Type: interfaces.METRICS_SAMEPERIOD,
					SamePeriodCfg: &interfaces.SamePeriodCfg{
						TimeGranularity: "year",
						Offset:          1,
						Method:          []string{interfaces.METRICS_SAMEPERIOD_METHOD_GROWTH_VALUE},
					},
				},
			},
				interfaces.MetricModel{}, &interfaces.DataView{})
			So(httpErr, ShouldBeNil)
		})
	})
}

func TestParsePromqlResult2Uniresponse(t *testing.T) {
	Convey("Test parsePromqlResult2Uniresponse", t, func() {

		promqlMetric := interfaces.MetricModel{
			QueryType: interfaces.PROMQL,
		}
		Convey("parse failed, because promql response Data not QueryData ", func() {
			promqlResp := interfaces.PromQLResponse{
				Status: "success",
				Data:   []string{},
			}

			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusUnprocessableEntity,
				Language: "en-US",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_MetricModel_InternalError_UnSupportPromQLResult,
					Description:  "Unsupported Promql Result Type",
					Solution:     "Please try this operation again, if the error occurs again, submit the work order or contact technical support engineers.",
					ErrorLink:    "None",
					ErrorDetails: "Unsupport Promql Result Type []",
				},
			}

			_, res := parsePromqlResult2Uniresponse(testENCtx, promqlResp,
				interfaces.MetricModelQuery{QueryTimeParams: interfaces.QueryTimeParams{StepStr: &fiveMin}}, promqlMetric)
			So(res, ShouldResemble, expectedErr)
		})

		Convey("parse failed, because promql response Data Result not Matrix ", func() {
			promqlResp := interfaces.PromQLResponse{
				Status: "success",
				Data: promql.QueryData{
					ResultType: parser.ValueTypeString,
					Result:     static.String{},
				},
			}

			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusUnprocessableEntity,
				Language: "en-US",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_MetricModel_InternalError_ParseResultFailed,
					Description:  "Parse Result to UniFormat Failed",
					Solution:     "Please try this operation again, if the error occurs again, submit the work order or contact technical support engineers.",
					ErrorLink:    "None",
					ErrorDetails: "Invalid Promql Result Type \"string\"",
				},
			}

			_, res := parsePromqlResult2Uniresponse(testENCtx, promqlResp,
				interfaces.MetricModelQuery{QueryTimeParams: interfaces.QueryTimeParams{StepStr: &fiveMin}}, promqlMetric)
			So(res, ShouldResemble, expectedErr)
		})

		// Convey("parse success with PageMatrix ", func() {
		// 	series := []static.Series{
		// 		{
		// 			Metric: []*labels.Label{
		// 				{
		// 					Name:  "cpu",
		// 					Value: "1",
		// 				},
		// 			},
		// 			Points: []static.Point{
		// 				{
		// 					T: 1695003480000,
		// 					V: 1.1,
		// 				},
		// 				{
		// 					T: 1695004560000,
		// 					V: 1.1,
		// 				},
		// 			},
		// 		},
		// 	}
		// 	mat := make(static.Matrix, 0)
		// 	mat = append(mat, series...)
		// 	promqlResp := interfaces.PromQLResponse{
		// 		Status: "success",
		// 		Data: promql.QueryData{
		// 			ResultType: parser.ValueTypeMatrix,
		// 			Result:     static.PageMatrix{Matrix: mat, TotalSeries: 1},
		// 		},
		// 		SeriesTotal: 1,
		// 	}

		// 	expectRes := interfaces.MetricModelUniResponse{
		// 		Model: interfaces.MetricModel{},
		// 		Datas: []interfaces.MetricModelData{
		// 			{
		// 				Labels: map[string]string{
		// 					"cpu": "1",
		// 				},
		// 				Times: []interface{}{int64(1695003480000), int64(1695003840000), int64(1695004200000),
		// 					int64(1695004560000), int64(1695004920000), int64(1695005280000)},
		// 				Values: []interface{}{1.1, nil, nil, 1.1, nil, nil},
		// 			},
		// 		},
		// 		Step:          "5m",
		// 		CurrSeriesNum: 1,
		// 		PointTotal:    2,
		// 		SeriesTotal:   1,
		// 	}

		// 	uniRes, res := parsePromqlResult2Uniresponse(testENCtx, promqlResp,
		// 		interfaces.MetricModelQuery{
		// 			QueryTimeParams: interfaces.QueryTimeParams{
		// 				Start:   1695003670521,
		// 				End:     1695005470521,
		// 				StepStr: &fiveMin,
		// 				Step:    360000,
		// 			},
		// 			MetricModelQueryParameters: queryParam,
		// 		})
		// 	So(uniRes, ShouldResemble, expectRes)
		// 	So(res, ShouldBeNil)
		// })

		// Convey("parse success with matrix ", func() {
		// 	series := []static.Series{
		// 		{
		// 			Metric: []*labels.Label{
		// 				{
		// 					Name:  "cpu",
		// 					Value: "1",
		// 				},
		// 			},
		// 			Points: []static.Point{
		// 				{
		// 					T: 1695003480000,
		// 					V: 1.1,
		// 				},
		// 				{
		// 					T: 1695004560000,
		// 					V: 1.1,
		// 				},
		// 			},
		// 		},
		// 	}
		// 	mat := make(static.Matrix, 0)
		// 	mat = append(mat, series...)
		// 	promqlResp := interfaces.PromQLResponse{
		// 		Status: "success",
		// 		Data: promql.QueryData{
		// 			ResultType: parser.ValueTypeMatrix,
		// 			Result:     mat,
		// 		},
		// 	}

		// 	expectRes := interfaces.MetricModelUniResponse{
		// 		Model: interfaces.MetricModel{},
		// 		Datas: []interfaces.MetricModelData{
		// 			{
		// 				Labels: map[string]string{
		// 					"cpu": "1",
		// 				},
		// 				Times: []interface{}{int64(1695003480000), int64(1695003840000), int64(1695004200000),
		// 					int64(1695004560000), int64(1695004920000), int64(1695005280000)},
		// 				Values: []interface{}{1.1, nil, nil, 1.1, nil, nil},
		// 			},
		// 		},
		// 		Step:          "5m",
		// 		CurrSeriesNum: 1,
		// 		PointTotal:    2,
		// 		SeriesTotal:   1,
		// 	}

		// 	uniRes, res := parsePromqlResult2Uniresponse(testENCtx, promqlResp,
		// 		interfaces.MetricModelQuery{
		// 			QueryTimeParams: interfaces.QueryTimeParams{
		// 				Start:   1695003670521,
		// 				End:     1695005470521,
		// 				StepStr: &fiveMin,
		// 				Step:    360000,
		// 			},
		// 			MetricModelQueryParameters: queryParam,
		// 		})
		// 	So(uniRes, ShouldResemble, expectRes)
		// 	So(res, ShouldBeNil)
		// })

		// Convey("parse success with Vector ", func() {
		// 	series := []static.Sample{
		// 		{
		// 			Metric: []*labels.Label{
		// 				{
		// 					Name:  "cpu",
		// 					Value: "1",
		// 				},
		// 			},
		// 			Point: static.Point{
		// 				T: 1652320539120,
		// 				V: 1.1,
		// 			},
		// 		},
		// 	}
		// 	mat := make(static.Vector, 0)
		// 	mat = append(mat, series...)
		// 	promqlResp := interfaces.PromQLResponse{
		// 		Status: "success",
		// 		Data: promql.QueryData{
		// 			ResultType: parser.ValueTypeMatrix,
		// 			Result:     mat,
		// 		},
		// 	}

		// 	expectRes := interfaces.MetricModelUniResponse{
		// 		Model: interfaces.MetricModel{},
		// 		Datas: []interfaces.MetricModelData{
		// 			{
		// 				Labels: map[string]string{
		// 					"cpu": "1",
		// 				},
		// 				Times:  []interface{}{int64(1652320539120)},
		// 				Values: []interface{}{1.1},
		// 			},
		// 		},
		// 		Step:          "5m",
		// 		CurrSeriesNum: 1,
		// 		PointTotal:    1,
		// 		SeriesTotal:   1,
		// 	}

		// 	uniRes, res := parsePromqlResult2Uniresponse(testENCtx, promqlResp, interfaces.MetricModelQuery{QueryTimeParams: interfaces.QueryTimeParams{StepStr: &fiveMin}, MetricModelQueryParameters: queryParam})
		// 	So(uniRes, ShouldResemble, expectRes)
		// 	So(res, ShouldBeNil)
		// })

		// Convey("parse success with Scalar ", func() {
		// 	series := static.Scalar{
		// 		T: 1652320539120,
		// 		V: 1.1,
		// 	}

		// 	promqlResp := interfaces.PromQLResponse{
		// 		Status: "success",
		// 		Data: promql.QueryData{
		// 			ResultType: parser.ValueTypeMatrix,
		// 			Result:     series,
		// 		},
		// 	}

		// 	expectRes := interfaces.MetricModelUniResponse{
		// 		Model: interfaces.MetricModel{},
		// 		Datas: []interfaces.MetricModelData{
		// 			{
		// 				Labels: map[string]string{},
		// 				Times:  []interface{}{int64(1652320539120)},
		// 				Values: []interface{}{1.1},
		// 			},
		// 		},
		// 		Step:          "5m",
		// 		CurrSeriesNum: 1,
		// 	}

		// 	uniRes, res := parsePromqlResult2Uniresponse(testENCtx, promqlResp, interfaces.MetricModelQuery{QueryTimeParams: interfaces.QueryTimeParams{StepStr: &fiveMin}})
		// 	So(uniRes, ShouldResemble, expectRes)
		// 	So(res, ShouldBeNil)
		// })
	})
}

func TestMacthPersist(t *testing.T) {
	Convey("Test macthPersist", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				FullCacheRefreshInterval: 24 * time.Hour,
			},
		}
		mmaMock := mock.NewMockMetricModelAccess(mockCtrl)
		ibaMock := mock.NewMockIndexBaseAccess(mockCtrl)
		promQLMock := mock.NewMockPromQLService(mockCtrl)
		dvsMock := mock.NewMockDataViewService(mockCtrl)
		vvsMock := mock.NewMockVegaService(mockCtrl)
		psMock := mock.NewMockPermissionService(mockCtrl)
		ds := MockNewMetricModelService(appSetting, mmaMock, ibaMock, promQLMock, dvsMock, vvsMock, psMock)

		Convey("return nil with QueryType is dsl and instant query", func() {
			_, err := ds.macthPersist(testENCtx, &interfaces.MetricModelQuery{
				QueryTimeParams: interfaces.QueryTimeParams{IsInstantQuery: true}},
				interfaces.MetricModel{QueryType: interfaces.DSL}, &interfaces.DataView{})
			So(err, ShouldBeNil)
		})

		Convey("return nil with task is null", func() {
			_, err := ds.macthPersist(testENCtx, &interfaces.MetricModelQuery{QueryTimeParams: interfaces.QueryTimeParams{IsInstantQuery: false}},
				interfaces.MetricModel{QueryType: interfaces.DSL}, &interfaces.DataView{})
			So(err, ShouldBeNil)
		})

		Convey("return nil with promql not match", func() {
			_, err := ds.macthPersist(testENCtx, &interfaces.MetricModelQuery{
				QueryTimeParams: interfaces.QueryTimeParams{
					IsInstantQuery: false,
					StepStr:        &fiveMin,
					Step:           &fiveMinInt,
				}},
				interfaces.MetricModel{
					QueryType: interfaces.PROMQL,
					Task: &interfaces.MetricTask{
						TaskID: "1",
						Steps:  []string{"10m"},
					},
				}, &interfaces.DataView{})
			So(err, ShouldBeNil)
		})

		Convey("return err with promql match and GetIndexBasesByTypes err", func() {
			dvsMock.EXPECT().GetDataViewByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("get index base failed"))

			_, err := ds.macthPersist(testENCtx, &interfaces.MetricModelQuery{
				QueryTimeParams: interfaces.QueryTimeParams{
					IsInstantQuery: false,
					StepStr:        &fiveMin,
					Step:           &fiveMinInt,
				},
			},
				interfaces.MetricModel{
					QueryType:   interfaces.PROMQL,
					MeasureName: "__m.a",
					Task: &interfaces.MetricTask{
						TaskID:    "1",
						Steps:     []string{"5m"},
						TaskName:  "task1",
						IndexBase: "base1",
					},
				}, &interfaces.DataView{})
			So(err.Error(), ShouldEqual, "get index base failed")
		})

		Convey("return nil with promql match  ", func() {
			dvsMock.EXPECT().GetDataViewByID(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&dataview, nil)

			_, err := ds.macthPersist(testENCtx, &interfaces.MetricModelQuery{
				QueryTimeParams: interfaces.QueryTimeParams{
					IsInstantQuery: false,
					StepStr:        &fiveMin,
					Step:           &fiveMinInt,
				},
			},
				interfaces.MetricModel{
					QueryType:   interfaces.PROMQL,
					MeasureName: "__m.a",
					Task: &interfaces.MetricTask{
						TaskID:    "1",
						Steps:     []string{"5m"},
						TaskName:  "task1",
						IndexBase: "base1",
					},
				}, &interfaces.DataView{})
			So(err, ShouldBeNil)
		})

		Convey("return nil with dsl parsedsl err", func() {
			_, err := ds.macthPersist(testENCtx, &interfaces.MetricModelQuery{
				QueryTimeParams: interfaces.QueryTimeParams{
					IsInstantQuery: false,
					StepStr:        &fiveMin,
					Step:           &fiveMinInt,
				},
				MetricModelID: "2",
			},
				interfaces.MetricModel{
					QueryType: interfaces.DSL,
					Task: &interfaces.MetricTask{
						TaskID: "1",
						Steps:  []string{"10m"},
					},
				}, &interfaces.DataView{})
			So(err, ShouldNotBeNil)
		})

		Convey("return nil with dsl not match", func() {
			_, err := ds.macthPersist(testENCtx, &interfaces.MetricModelQuery{
				QueryTimeParams: interfaces.QueryTimeParams{
					IsInstantQuery: false,
					Start:          &start,
					End:            &end,
					StepStr:        &fiveMin,
					Step:           &fiveMinInt,
				},
				Formula: `{"size":0,"aggs": {
				"NAME1": {
				  "terms": {
					"field": "labels.cpu.keyword",
					"size": 10
				  },
				  "aggs": {
					"time": {
					  "date_histogram": {
						"field": "@timestamp",
						"fixed_interval": "6m"
					  },
					  "aggs": {
						"NAME": {
						  "top_hits": {
							"size": 3,
							  "sort": [
								{
								  "@timestamp": {
									"order": "desc"
								  }
								}
							  ],
							  "_source": {
								"includes": [ "labels.cpu","metrics.node_cpu_seconds_total" ]
							  }
							}
						}
					  }
					}
				  }
				}
			  }}`},
				interfaces.MetricModel{
					QueryType: interfaces.DSL,
					Task: &interfaces.MetricTask{
						TaskID:      "1",
						Steps:       []string{"5m"},
						TimeWindows: []string{"10m"},
					},
				}, &interfaces.DataView{})
			So(err, ShouldBeNil)
		})

		Convey("return nil with dsl match", func() {
			dvsMock.EXPECT().GetDataViewByID(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&dataview, nil)

			_, err := ds.macthPersist(testENCtx, &interfaces.MetricModelQuery{
				MetricModelID: "1",
				QueryTimeParams: interfaces.QueryTimeParams{
					IsInstantQuery: false,
					Start:          &start,
					End:            &end,
					StepStr:        &fiveMin,
					Step:           &fiveMinInt,
				},
				Formula: `{"size":0,"aggs": {
				"NAME1": {
				  "terms": {
					"field": "labels.cpu.keyword",
					"size": 10
				  },
				  "aggs": {
					"time": {
					  "date_histogram": {
						"field": "@timestamp",
						"fixed_interval": "5m"
					  },
					  "aggs": {
						"NAME": {
						  "top_hits": {
							"size": 3,
							  "sort": [
								{
								  "@timestamp": {
									"order": "desc"
								  }
								}
							  ],
							  "_source": {
								"includes": [ "labels.cpu","metrics.node_cpu_seconds_total" ]
							  }
							}
						}
					  }
					}
				  }
				}
			  }}`},
				interfaces.MetricModel{
					QueryType: interfaces.DSL,
					Task: &interfaces.MetricTask{
						TaskID:      "1",
						Steps:       []string{"5m"},
						TimeWindows: []string{"5m"},
					},
				}, &interfaces.DataView{})
			So(err, ShouldBeNil)
		})
	})
}

func TestExec(t *testing.T) {
	Convey("Test Exec", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		loc, _ := time.LoadLocation(os.Getenv("TZ"))
		common.APP_LOCATION = loc

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				FullCacheRefreshInterval: 24 * time.Hour,
			},
		}
		mmaMock := mock.NewMockMetricModelAccess(mockCtrl)
		ibaMock := mock.NewMockIndexBaseAccess(mockCtrl)
		promQLMock := mock.NewMockPromQLService(mockCtrl)
		dvsMock := mock.NewMockDataViewService(mockCtrl)
		vvsMock := mock.NewMockVegaService(mockCtrl)
		psMock := mock.NewMockPermissionService(mockCtrl)
		ds := MockNewMetricModelService(appSetting, mmaMock, ibaMock, promQLMock, dvsMock, vvsMock, psMock)

		Convey("Exec failed, get metric model failed", func() {

			err := fmt.Errorf("get metric model failed")
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: "en-US",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_MetricModel_InternalError_GetModelByIdFailed,
					Description:  "Get Metric Model by Id Failed",
					Solution:     "Please try this operation again, if the error occurs again, submit the work order or contact technical support engineers.",
					ErrorLink:    "None",
					ErrorDetails: err.Error(),
				},
			}

			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mmaMock.EXPECT().GetMetricModel(gomock.Any(), gomock.Any()).AnyTimes().Return(
				[]interfaces.MetricModel{}, false, err)

			_, _, _, httpErr := ds.Exec(testENCtx, &interfaces.MetricModelQuery{MetricModelID: "1", MetricType: "atomic", QueryType: "promql"})
			So(httpErr, ShouldResemble, expectedErr)
		})

		Convey("Exec failed, metric model not found ", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusNotFound,
				Language: "en-US",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_MetricModel_MetricModelNotFound,
					Description:  "The Metric Model Does Not Exist",
					Solution:     "Please check whether the parameter is correct.",
					ErrorLink:    "None",
					ErrorDetails: "",
				},
			}

			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mmaMock.EXPECT().GetMetricModel(gomock.Any(), gomock.Any()).AnyTimes().Return(
				[]interfaces.MetricModel{}, false, nil)

			_, _, _, httpErr := ds.Exec(testENCtx, &interfaces.MetricModelQuery{MetricModelID: "1", MetricType: "atomic", QueryType: "promql"})
			So(httpErr, ShouldResemble, expectedErr)
		})

		Convey("Exec failed, analysis dimenssion is not belongs to model ", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusBadRequest,
				Language: "en-US",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_MetricModel_AnalysisDimensionNotExisted,
					Description:  "Analysis Dimension Not Existed In Vega Logic View",
					Solution:     "Please check whether the parameter is correct.",
					ErrorLink:    "None",
					ErrorDetails: "analysis dimension[f2] is not belongs to model [vega1], expected analysis dimensions should belongs to [{f1   <nil>}]",
				},
			}
			metricModel := interfaces.MetricModel{
				ModelName:  "txy",
				MetricType: "atomic",
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.QueryType_SQL,
					ID:   "vega1",
				},
				QueryType: "sql",
				FormulaConfig: interfaces.SQLConfig{
					AggrExprStr: "avg(f1)",
				},
				DateField:    "@timestamp",
				AnalysisDims: []interfaces.Field{{Name: "f1"}},
				MeasureField: "value",
				UnitType:     "timeUnit",
				Unit:         "ms",
			}

			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mmaMock.EXPECT().GetMetricModel(gomock.Any(), gomock.Any()).AnyTimes().Return(
				[]interfaces.MetricModel{metricModel}, true, nil)
			dvsMock.EXPECT().GetDataViewByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(
				&dataview, nil)

			_, _, _, httpErr := ds.Exec(testENCtx, &interfaces.MetricModelQuery{MetricModelID: "1", AnalysisDims: []string{"f2"}})
			So(httpErr, ShouldResemble, expectedErr)
		})

		// Convey("Exec failed, GetVegaViewFieldsByID err ", func() {
		// 	expectedErr := &rest.HTTPError{
		// 		HTTPCode: http.StatusInternalServerError,
		// 		Language: "en-US",
		// 		BaseError: rest.BaseError{
		// 			ErrorCode:    uerrors.Uniquery_MetricModel_InternalError_GetVegaViewFieldsByIDFailed,
		// 			Description:  "Get Vega Logic View's Fields By ID Failed",
		// 			Solution:     "Please try this operation again, if the error occurs again, submit the work order or contact technical support engineers.",
		// 			ErrorLink:    "None",
		// 			ErrorDetails: "error",
		// 		},
		// 	}
		// 	metricModel := interfaces.MetricModel{
		// 		ModelName:  "txy",
		// 		MetricType: "atomic",
		// 		DataSource: &interfaces.MetricDataSource{
		// 			Type: interfaces.QueryType_SQL,
		// 			ID:   "vega1",
		// 		},
		// 		QueryType: "sql",
		// 		FormulaConfig: interfaces.SQLConfig{
		// 			AggrExprStr: "avg(f1)",
		// 		},
		// 		DateField:    "@timestamp",
		// 		MeasureField: "value",
		// 		UnitType:     "timeUnit",
		// 		Unit:         "ms",
		// 	}
		// 	metricModelFilters := interfaces.MetricModelWithFilters{
		// 		MetricModel: metricModel,
		// 		DataView:    dataViewQueryFilters,
		// 	}

		// 	psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		// 	mmaMock.EXPECT().GetMetricModel(gomock.Any(), gomock.Any()).AnyTimes().Return(
		// 		[]interfaces.MetricModelWithFilters{metricModelFilters}, true, nil)
		// 	vvsMock.EXPECT().GetVegaViewFieldsByID(gomock.Any(), gomock.Any()).Return(interfaces.VegaViewWithFields{}, fmt.Errorf("error"))

		// 	_, _, _, httpErr := ds.Exec(testENCtx, &interfaces.MetricModelQuery{MetricModelID: "1"})
		// 	So(httpErr, ShouldResemble, expectedErr)
		// })

		Convey("Exec failed, because of GetDataFromOpenSearch err", func() {
			indices := []string{"a", "b"}
			dvsMock.EXPECT().BuildViewQuery4MetricModel(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(interfaces.ViewQuery4Metric{
					BaseTypes: []string{"a", "b"},
					Indices:   indices,
				}, nil)
			dvsMock.EXPECT().GetDataViewByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(
				&dataview, nil)
			dvsMock.EXPECT().GetDataFromOpenSearch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(nil, http.StatusInternalServerError,
				uerrors.NewOpenSearchError(uerrors.InternalServerError).WithReason("Error getting response from opensearch"))

			metricModel := interfaces.MetricModel{
				ModelName:  "txy",
				MetricType: "atomic",
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "主机监控",
				},
				QueryType: "dsl",
				Formula: `{"size":0,"aggs": {
					"NAME1": {
					  "terms": {
						"field": "labels.cpu.keyword",
						"size": 10
					  },
					  "aggs": {
						"NAME2": {
						  "date_histogram": {
							"field": "@timestamp",
							"fixed_interval": "1d"
						  },
						  "aggs": {
							"NAME3": {
							  "value_count": {
								"field": "1"
							  }
							}
						  }
						}
					  }
					}
				  }}`,
				DateField:    "@timestamp",
				MeasureField: "value",
				UnitType:     "timeUnit",
				Unit:         "ms",
			}
			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mmaMock.EXPECT().GetMetricModel(gomock.Any(), gomock.Any()).AnyTimes().Return(
				[]interfaces.MetricModel{metricModel}, true, nil)

			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: "en-US",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_InternalError,
					Description:  "Internal Error",
					Solution:     "Please try this operation again, if the error occurs again, submit the work order or contact technical support engineers.",
					ErrorLink:    "None",
					ErrorDetails: "{\"status\":500,\"error\":{\"type\":\"UniQuery.InternalServerError\",\"reason\":\"Error getting response from opensearch\"}}",
				},
			}

			starti := int64(1678412100123)
			endi := int64(1678412210123)
			stepi := int64(180000)
			_, _, _, httpErr := ds.Exec(testENCtx,
				&interfaces.MetricModelQuery{
					MetricModelID: "1",
					QueryTimeParams: interfaces.QueryTimeParams{
						Start: &starti,
						End:   &endi,
						Step:  &stepi,
					},
				})
			So(httpErr, ShouldResemble, expectedErr)
		})

		// Convey("Exec success ", func() {
		// 	metricModel := interfaces.MetricModel{
		// 		ModelName:  "txy",
		// 		MetricType: "atomic",
		// 		DataSource: &interfaces.MetricDataSource{
		// 			Type: interfaces.SOURCE_TYPE_DATA_VIEW,
		// 			ID:   "主机监控",
		// 		},
		// 		QueryType:    "promql",
		// 		Formula:      "rate(node_network_receive_bytes_total[5m])",
		// 		DateField:    "@timestamp",
		// 		MeasureField: "value",
		// 		UnitType:     "timeUnit",
		// 		Unit:         "ms",
		// 	}
		// 	metricModelFilters := interfaces.MetricModelWithFilters{
		// 		MetricModel: metricModel,
		// 		DataView:    interfaces.DataView{},
		// 	}
		// 	expectResult := interfaces.PromQLResponse{
		// 		Status: "success",
		// 		Data: promql.QueryData{
		// 			Result: static.Matrix{},
		// 		},
		// 	}

		// 	psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		// 	mmaMock.EXPECT().GetMetricModel(gomock.Any(), gomock.Any()).AnyTimes().Return(
		// 		[]interfaces.MetricModelWithFilters{metricModelFilters}, true, nil)

		// 	promQLMock.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(expectResult, nil, 0, nil)

		// 	result, _, _, httpErr := ds.Exec(testENCtx, &interfaces.MetricModelQuery{MetricModelID: "1", QueryTimeParams: interfaces.QueryTimeParams{Start: 1678412100123,
		// 		End: 1678412210123, Step: 180000}, MetricModelQueryParameters: queryParam})
		// 	So(result, ShouldResemble, interfaces.MetricModelUniResponse{
		// 		Model:      metricModel,
		// 		Datas:      []interfaces.MetricModelData{},
		// 		Step:       "",
		// 		StatusCode: 200,
		// 	})
		// 	So(httpErr, ShouldBeNil)
		// })
	})
}

func TestGetMetricModelIDByName(t *testing.T) {
	Convey("Test GetMetricModelIDByName", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				FullCacheRefreshInterval: 24 * time.Hour,
			},
		}
		mmaMock := mock.NewMockMetricModelAccess(mockCtrl)
		ibaMock := mock.NewMockIndexBaseAccess(mockCtrl)
		promQLMock := mock.NewMockPromQLService(mockCtrl)
		dvsMock := mock.NewMockDataViewService(mockCtrl)
		vvsMock := mock.NewMockVegaService(mockCtrl)
		psMock := mock.NewMockPermissionService(mockCtrl)
		ds := MockNewMetricModelService(appSetting, mmaMock, ibaMock, promQLMock, dvsMock, vvsMock, psMock)

		Convey("Get Metric ModelID By Name Failed", func() {
			err := errors.New("Get Metric Model Id  By Name Failed")
			mmaMock.EXPECT().GetMetricModelIDByName(gomock.Any(),
				gomock.Any(), gomock.Any()).Return("0", false, err)
			modelID, httpErr := ds.GetMetricModelIDByName(testENCtx, "group1", "model1")
			So(modelID, ShouldEqual, "")
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_InternalError_GetModelIdByNameFailed)
		})
		Convey("Metric Model Not Found", func() {

			mmaMock.EXPECT().GetMetricModelIDByName(gomock.Any(),
				gomock.Any(), gomock.Any()).Return("0", false, nil)
			modelID, httpErr := ds.GetMetricModelIDByName(testENCtx, "group1", "model1")
			So(modelID, ShouldEqual, "")
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusNotFound)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_MetricModel_MetricModelNotFound)
		})
		Convey("Success", func() {

			mmaMock.EXPECT().GetMetricModelIDByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(
				"1111", true, nil)
			modelID, httpErr := ds.GetMetricModelIDByName(testENCtx, "group1", "model1")
			So(modelID, ShouldEqual, "1111")
			So(httpErr, ShouldBeNil)
		})
	})
}

func TestGetSeries(t *testing.T) {
	Convey("Test getSeries", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				FullCacheRefreshInterval: 24 * time.Hour,
			},
		}
		mmaMock := mock.NewMockMetricModelAccess(mockCtrl)
		ibaMock := mock.NewMockIndexBaseAccess(mockCtrl)
		promQLMock := mock.NewMockPromQLService(mockCtrl)
		dvsMock := mock.NewMockDataViewService(mockCtrl)
		vvsMock := mock.NewMockVegaService(mockCtrl)
		psMock := mock.NewMockPermissionService(mockCtrl)
		ds := MockNewMetricModelService(appSetting, mmaMock, ibaMock, promQLMock, dvsMock, vvsMock, psMock)

		query := interfaces.MetricModelQuery{
			QueryTimeParams: interfaces.QueryTimeParams{
				Start:   &start,
				End:     &end,
				StepStr: &step_1min,
				Step:    &step_60000,
			},
			MetricType: interfaces.ATOMIC_METRIC, QueryType: interfaces.DSL,
			Formula: `{"size":0,"aggs": {
			"cpu": {
			  "terms": {
				"field": "labels.cpu",
				"size": 20000
			  },
			  "aggs": {
				"time": {
				  "date_histogram": {
					"field": "@timestamp",
					"fixed_interval": "1d"
				  },
				  "aggs": {
					"NAME3": {
					  "value_count": {
						"field": "1"
					  }
					}
				  }
				}
			  }
			}
		  }}`,
			MeasureField: "NAME3"}
		dslInfo, _ := parseDsl(testENCtx, &query, &dataview)

		Convey("get series with cache", func() {
			Series_Of_Model_Map.Store(query.MetricModelID, DSLSeries{
				RefreshTime:     time.Now(),
				FullRefreshTime: time.Now(),
				StartTime:       time.UnixMilli(start),
				EndTime:         time.UnixMilli(end),
				Series:          []map[string]string{},
			})
			query.IgnoringMemoryCache = false
			_, httpErr := ds.getSeries(testENCtx, query, &dslInfo, &dataview)

			So(httpErr, ShouldBeNil)
		})

		Convey("get series failed when opensearch with buffer error", func() {
			dvsMock.EXPECT().GetDataFromOpenSearch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).Times(1).Return(nil, http.StatusOK, errors.New("GetDataFromOpenSearchWithBuffer error"))

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(nil, http.StatusOK, errors.New("GetDataFromOpenSearchWithBuffer error"))

			query.IgnoringMemoryCache = true
			_, httpErr := ds.getSeries(testENCtx, query, &dslInfo, &dataview)
			So(httpErr, ShouldNotBeNil)
		})

		Convey("get series failed when opensearch error", func() {
			dvsMock.EXPECT().GetDataFromOpenSearch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).Times(1).Return(convert.MapToByte(cardinalityDslResult), http.StatusOK, nil)
			dvsMock.EXPECT().GetDataFromOpenSearch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(nil, http.StatusOK, errors.New("GetDataFromOpenSearch error"))
			// dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			// 	gomock.Any()).AnyTimes().Return(convert.MapToByte(cardinalityDslResult), http.StatusOK, nil)
			query.IgnoringMemoryCache = true
			_, httpErr := ds.getSeries(testENCtx, query, &dslInfo, &dataview)
			So(httpErr, ShouldNotBeNil)
		})

		Convey("get series success when dsl range with batch", func() {
			dvsMock.EXPECT().GetDataFromOpenSearch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).Times(1).Return(convert.MapToByte(cardinalityDslResult), http.StatusOK, nil)
			dvsMock.EXPECT().GetDataFromOpenSearch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(convert.MapToByte(seriesDslResult), http.StatusOK, nil)

			query.IgnoringMemoryCache = true
			query.MetricModelID = "1"

			result, httpErr := ds.getSeries(testENCtx, query, &dslInfo, &dataview)
			So(len(result), ShouldEqual, 12)
			So(httpErr, ShouldBeNil)
		})

	})
}

func TestGetSeriesData(t *testing.T) {
	Convey("Test getSeriesData", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				FullCacheRefreshInterval: 24 * time.Hour,
			},
		}
		mmaMock := mock.NewMockMetricModelAccess(mockCtrl)
		ibaMock := mock.NewMockIndexBaseAccess(mockCtrl)
		promQLMock := mock.NewMockPromQLService(mockCtrl)
		dvsMock := mock.NewMockDataViewService(mockCtrl)
		vvsMock := mock.NewMockVegaService(mockCtrl)
		psMock := mock.NewMockPermissionService(mockCtrl)
		ds := MockNewMetricModelService(appSetting, mmaMock, ibaMock, promQLMock, dvsMock, vvsMock, psMock)

		util.InitAntsPool(common.PoolSetting{
			MegerPoolSize:       10,
			ExecutePoolSize:     10,
			BatchSubmitPoolSize: 10,
		})

		common.FixedStepsMap = StepsMap

		starti := int64(1722691260000)
		endi := int64(1722730600000)
		stepi := int64(11000)
		query := interfaces.MetricModelQuery{
			MetricType: interfaces.ATOMIC_METRIC,
			QueryType:  interfaces.DSL,
			QueryTimeParams: interfaces.QueryTimeParams{
				Start: &starti,
				End:   &endi, //39,340
				Step:  &stepi,
			},
			QueryTimeNum: 3577,
			Formula: `{"size":0,"aggs": {
			"cpu": {
			  "terms": {
				"field": "labels.cpu",
				"size": 20000
			  },
			  "aggs": {
				"time": {
				  "date_histogram": {
					"field": "@timestamp",
					"fixed_interval": "5m"
				  },
				  "aggs": {
					"NAME3": {
					  "value_count": {
						"field": "1"
					  }
					}
				  }
				}
			  }
			}
		  }}`,
			MeasureField: "NAME3"}
		dslInfo, _ := parseDsl(testENCtx, &query, &dataview)
		series := make([]map[string]string, 0)
		for i := 0; i < 20000; i++ {
			series = append(series, map[string]string{
				"cpu": fmt.Sprintf("%d", i),
			})
		}

		Convey("getSeriesData success with range query ", func() {
			defer util.ExecutePool.Release()
			defer util.InitAntsPool(common.PoolSetting{
				MegerPoolSize:       10,
				ExecutePoolSize:     10,
				BatchSubmitPoolSize: 10,
			})
			dvsMock.EXPECT().GetDataFromOpenSearch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			_, err := ds.getSeriesData(testENCtx, query, dslInfo, interfaces.MetricModel{}, series, []string{"a", "b"}, 10, &dataview)
			So(err, ShouldBeNil)
		})

		Convey("getSeriesData success with instant query ", func() {
			defer util.ExecutePool.Release()
			defer util.InitAntsPool(common.PoolSetting{
				MegerPoolSize:       10,
				ExecutePoolSize:     10,
				BatchSubmitPoolSize: 10,
			})
			dvsMock.EXPECT().GetDataFromOpenSearch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			query.IsInstantQuery = true
			_, err := ds.getSeriesData(testENCtx, query, dslInfo, interfaces.MetricModel{}, series, []string{"a", "b"}, 10, &dataview)
			So(err, ShouldBeNil)
		})

		Convey("getSeriesData success with instant query && mix buckets aggs ", func() {
			starti := int64(1722691260000)
			endi := int64(1722730600000)
			stepi := int64(1800000)
			stepStri := "30m"
			query1 := interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.DSL,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &starti,
					End:     &endi,
					Step:    &stepi,
					StepStr: &stepStri,
				},
				QueryTimeNum: 22,
				Formula: `{
					"size": 0,
					"query": {
					  "bool": {
						"filter": [
						  {
							"exists": {
							  "field": "metrics.node_cpu_seconds_total"
							}
						  }
						],
						"must": []
					  }
					},
					"aggs": {
					  "job": {
						"terms": {
						  "field": "labels.job",
						  "size": 1,
						  "order": {
							"_key": "asc"
						  }
						},
						"aggs": {
						  "mode": {
							"filters": {
							  "filters": {
								"idle": {
								  "match": {
									"labels.mode": "idle"
								  }
								},
								"system": {
								  "match": {
									"labels.mode": "system"
								  }
								}
							  }
							},
							"aggs": {
							  "date_rg": {
								"date_range": {
								  "field": "@timestamp",
								  "format": "yyyy-MM-dd",
								  "ranges": [
									{
									  "from": "now-52d/d",
									  "to": "now-51d/d"
									},
									{
									  "from": "now-53d/d",
									  "to": "now-52d/d"
									}
								  ]
								},
								"aggs": {
								  "instance": {
									"terms": {
									  "field": "labels.instance",
									  "size": 12,
									  "order": {
										"_key": "desc"
									  }
									},
									"aggs": {
									  "cpu": {
										"terms": {
										  "field": "labels.cpu",
										  "size": 48,
										  "order": {
											"_key": "asc"
										  }
										},
										"aggs": {
										  "time": {
											"date_histogram": {
											  "field": "@timestamp",
											  "fixed_interval": "{{__interval}}",
											  "min_doc_count": 1
											},
											"aggs": {
											  "value": {
												"avg": {
												  "field": "metrics.node_cpu_seconds_total"
												}
											  }
											}
										  }
										}
									  }
									}
								  }
								}
							  }
							}
						  }
						}
					  }
					}
				  }`,
				MeasureField: "value"}
			query1.IsInstantQuery = false
			query1.IgnoringMemoryCache = true

			dslInfo1, _ := parseDsl(testENCtx, &query1, &dataview)
			series1 := make([]map[string]string, 0)
			for i := 0; i < 12; i++ {
				for j := 0; j < 48; j++ {
					seriesi := make(map[string]string)
					seriesi["job"] = "prometheus"
					seriesi["instance"] = fmt.Sprintf("10.4.14.%d:9102", i)
					seriesi["cpu"] = fmt.Sprintf("%d", j)
					series1 = append(series1, seriesi)
				}
			}

			defer util.ExecutePool.Release()
			defer util.InitAntsPool(common.PoolSetting{
				MegerPoolSize:       10,
				ExecutePoolSize:     10,
				BatchSubmitPoolSize: 10,
			})
			dvsMock.EXPECT().GetDataFromOpenSearch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			_, errP := processDateHistogram(testENCtx, &query1, dslInfo1, ds.appSetting.PromqlSetting.LookbackDelta)
			So(errP, ShouldBeNil)

			_, err := ds.getSeriesData(testENCtx, query1, dslInfo1, interfaces.MetricModel{}, series1, []string{"a", "b"}, 10, &dataview)
			So(err, ShouldBeNil)
		})

		Convey("getSeriesData success with instant query && 4 terms aggs ", func() {
			starti := int64(1722691260000)
			endi := int64(1722730600000)
			stepi := int64(1800000)
			stepStri := "30m"
			query1 := interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.DSL,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &starti,
					End:     &endi,
					Step:    &stepi,
					StepStr: &stepStri,
				},
				QueryTimeNum: 22,
				Formula: `{
					"size": 0,
					"query": {
					  "bool": {
						"filter": [
						  {
							"exists": {
							  "field": "metrics.node_cpu_seconds_total"
							}
						  }
						],
						"must": []
					  }
					},
					"aggs": {
					  "job": {
						"terms": {
						  "field": "labels.job",
						  "size": 1,
						  "order": {
							"_key": "asc"
						  }
						},
						"aggs": {
						  "mode": {
							"terms": {
							  "field": "labels.mode",
							  "size": 3,
							  "order": {
								"_key": "desc"
							  }
							},
							"aggs": {
							  "instance": {
								"terms": {
								  "field": "labels.instance",
								  "size": 100,
								  "order": {
									"_key": "desc"
								  }
								},
								"aggs": {
								  "cpu": {
									"terms": {
									  "field": "labels.cpu",
									  "size": 25,
									  "order": {
										"_key": "desc"
									  }
									},
									"aggs": {
									  "time": {
										"date_histogram": {
										  "field": "@timestamp",
										  "fixed_interval": "{{__interval}}",
										  "min_doc_count": 1
										},
										"aggs": {
										  "value": {
											"avg": {
											  "field": "metrics.node_cpu_seconds_total"
											}
										  }
										}
									  }
									}
								  }
								}
							  }
							}
						  }
						}
					  }
					}
				  }`,
				MeasureField: "value"}
			query1.IsInstantQuery = false
			query1.IgnoringMemoryCache = true

			dslInfo1, _ := parseDsl(testENCtx, &query1, &dataview)
			series1 := make([]map[string]string, 0)
			modes := []string{"idle", "iowait", "irq"}
			for _, mode := range modes {
				for i := 0; i < 100; i++ {
					for j := 0; j < 25; j++ {
						seriesi := make(map[string]string)
						seriesi["mode"] = mode
						seriesi["job"] = "prometheus"
						seriesi["instance"] = fmt.Sprintf("10.4.14.%d:9102", i)
						seriesi["cpu"] = fmt.Sprintf("%d", j)
						series1 = append(series1, seriesi)
					}
				}
			}

			defer util.ExecutePool.Release()
			defer util.InitAntsPool(common.PoolSetting{
				MegerPoolSize:       10,
				ExecutePoolSize:     10,
				BatchSubmitPoolSize: 10,
			})
			dvsMock.EXPECT().GetDataFromOpenSearch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			_, errP := processDateHistogram(testENCtx, &query1, dslInfo1, ds.appSetting.PromqlSetting.LookbackDelta)
			So(errP, ShouldBeNil)

			_, err := ds.getSeriesData(testENCtx, query1, dslInfo1, interfaces.MetricModel{}, series1, []string{"a", "b"}, 10, &dataview)
			So(err, ShouldBeNil)
		})

		Convey("getSeriesData failed with opensearch error ", func() {
			defer util.ExecutePool.Release()
			defer util.InitAntsPool(common.PoolSetting{
				MegerPoolSize:       10,
				ExecutePoolSize:     10,
				BatchSubmitPoolSize: 10,
			})
			dvsMock.EXPECT().GetDataFromOpenSearch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(nil, http.StatusOK, errors.New("opensearch error"))

			query.IsInstantQuery = true
			_, err := ds.getSeriesData(testENCtx, query, dslInfo, interfaces.MetricModel{}, series, []string{"a", "b"}, 10, &dataview)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestGetMetricModelFields(t *testing.T) {
	Convey("Test GetMetricModelFields", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				FullCacheRefreshInterval: 24 * time.Hour,
			},
		}
		mmaMock := mock.NewMockMetricModelAccess(mockCtrl)
		ibaMock := mock.NewMockIndexBaseAccess(mockCtrl)
		promQLMock := mock.NewMockPromQLService(mockCtrl)
		dvsMock := mock.NewMockDataViewService(mockCtrl)
		vvsMock := mock.NewMockVegaService(mockCtrl)
		psMock := mock.NewMockPermissionService(mockCtrl)
		ds := MockNewMetricModelService(appSetting, mmaMock, ibaMock, promQLMock, dvsMock, vvsMock, psMock)

		Convey("GetMetricModelFields failed with GetMetricModel error", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: "zh-CN",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_MetricModel_InternalError_GetModelByIdFailed,
					Description:  "按ID获取指标模型失败",
					Solution:     "请重试该操作，若再次出现该错误请提交工单或联系技术支持工程师。",
					ErrorLink:    "暂无",
					ErrorDetails: "get metric model failed",
				},
			}

			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mmaMock.EXPECT().GetMetricModel(gomock.Any(), gomock.Any()).Return(
				[]interfaces.MetricModel{}, false, fmt.Errorf("get metric model failed"))

			fields, err := ds.GetMetricModelFields(testCtx, "test-id")

			So(err, ShouldResemble, expectedErr)
			So(fields, ShouldBeNil)
		})

		Convey("GetMetricModelFields failed with metric model not found", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusNotFound,
				Language: "zh-CN",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_MetricModel_MetricModelNotFound,
					Description:  "指标模型不存在",
					Solution:     "请检查参数是否正确。",
					ErrorLink:    "暂无",
					ErrorDetails: "",
				},
			}

			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mmaMock.EXPECT().GetMetricModel(gomock.Any(), gomock.Any()).Return(
				[]interfaces.MetricModel{}, true, nil)

			fields, err := ds.GetMetricModelFields(testCtx, "test-id")

			So(err, ShouldResemble, expectedErr)
			So(fields, ShouldBeNil)
		})

		Convey("GetMetricModelFields success with DSL query type", func() {

			model := interfaces.MetricModel{
				ModelID:    "test-id",
				QueryType:  interfaces.DSL,
				MetricType: interfaces.ATOMIC_METRIC,
				DataSource: &interfaces.MetricDataSource{
					ID: "view1",
				},
				FieldsMap: map[string]interfaces.Field{
					"field1": {Name: "field1", Type: dtype.DataType_Text},
					"field2": {Name: "field2", Type: dtype.DataType_String},
				},
			}
			// dataView := interfaces.DataView{
			// 	FieldsMap: map[string]*cond.ViewField{
			// 		"field1": {Name: "field1", Type: dtype.DataType_Text},
			// 		"field2": {Name: "field2", Type: dtype.DataType_String},
			// 	},
			// }

			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mmaMock.EXPECT().GetMetricModel(gomock.Any(), gomock.Any()).Return(
				[]interfaces.MetricModel{model}, true, nil)
			// dvsMock.EXPECT().GetDataViewByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(
			// 	&dataView, nil)

			fields, err := ds.GetMetricModelFields(testCtx, "test-id")

			So(err, ShouldBeNil)
			So(len(fields), ShouldEqual, 2)
		})

		Convey("GetMetricModelFields failed with PROMQL GetFields error", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: "zh-CN",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_MetricModel_InternalError_GetFieldsFailed,
					Description:  "获取指标原始字段失败",
					Solution:     "请重试该操作，若再次出现该错误请提交工单或联系技术支持工程师。",
					ErrorLink:    "暂无",
					ErrorDetails: "GetFields error",
				},
			}

			model := interfaces.MetricModel{
				ModelID:    "test-id",
				QueryType:  interfaces.PROMQL,
				MetricType: interfaces.ATOMIC_METRIC,
				DataSource: &interfaces.MetricDataSource{
					ID: "view1",
				},
				Formula: "test_formula",
				Task: &interfaces.MetricTask{
					TaskID: "1",
				},
			}

			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mmaMock.EXPECT().GetMetricModel(gomock.Any(), gomock.Any()).Return(
				[]interfaces.MetricModel{model}, true, nil)
			dvsMock.EXPECT().GetDataViewByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(
				&dataview, nil)

			promQLMock.EXPECT().GetFields(gomock.Any(), gomock.Any()).Return(
				nil, http.StatusBadRequest, fmt.Errorf("GetFields error"))

			fields, err := ds.GetMetricModelFields(testCtx, "test-id")

			So(err, ShouldResemble, expectedErr)
			So(fields, ShouldBeNil)
		})

		Convey("GetMetricModelFields success with PROMQL", func() {

			model := interfaces.MetricModel{
				ModelID:    "test-id",
				QueryType:  interfaces.PROMQL,
				MetricType: interfaces.ATOMIC_METRIC,
				DataSource: &interfaces.MetricDataSource{
					ID: "view1",
				},
				Formula: "test_formula",
				FieldsMap: map[string]interfaces.Field{
					"labels.field1": {Name: "labels.field1", Type: dtype.DataType_String},
					"labels.field2": {Name: "labels.field2", Type: dtype.DataType_String},
				},
			}
			dataView := interfaces.DataView{
				FieldsMap: map[string]*cond.ViewField{
					"labels.field1": {Name: "labels.field1", Type: dtype.DataType_String},
					"labels.field2": {Name: "labels.field2", Type: dtype.DataType_String},
				},
			}

			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mmaMock.EXPECT().GetMetricModel(gomock.Any(), gomock.Any()).Return(
				[]interfaces.MetricModel{model}, true, nil)
			dvsMock.EXPECT().GetDataViewByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(
				&dataView, nil)

			promQLMock.EXPECT().GetFields(gomock.Any(), gomock.Any()).Return(
				map[string]bool{
					"field1": true,
					"field2": true,
				}, http.StatusOK, nil)

			fields, err := ds.GetMetricModelFields(testCtx, "test-id")

			So(err, ShouldBeNil)
			So(len(fields), ShouldEqual, 2)
		})

	})
}

func TestGetMetricModelFieldValues(t *testing.T) {
	Convey("Test GetMetricModelFieldValues", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				FullCacheRefreshInterval: 24 * time.Hour,
			},
		}
		mmaMock := mock.NewMockMetricModelAccess(mockCtrl)
		ibaMock := mock.NewMockIndexBaseAccess(mockCtrl)
		promQLMock := mock.NewMockPromQLService(mockCtrl)
		dvsMock := mock.NewMockDataViewService(mockCtrl)
		vvsMock := mock.NewMockVegaService(mockCtrl)
		psMock := mock.NewMockPermissionService(mockCtrl)
		ds := MockNewMetricModelService(appSetting, mmaMock, ibaMock, promQLMock, dvsMock, vvsMock, psMock)

		Convey("GetMetricModelFieldValues when GetMetricModel error", func() {
			expectedErr := rest.NewHTTPError(testENCtx, http.StatusInternalServerError,
				uerrors.Uniquery_MetricModel_InternalError_GetModelByIdFailed).WithErrorDetails("GetMetricModel error")

			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mmaMock.EXPECT().GetMetricModel(gomock.Any(), gomock.Any()).Return(
				nil, false, fmt.Errorf("GetMetricModel error"))

			fieldValues, err := ds.GetMetricModelFieldValues(testENCtx, "test-id", "test-field")

			So(err, ShouldResemble, expectedErr)
			So(fieldValues, ShouldResemble, interfaces.FieldValues{})
		})

		Convey("GetMetricModelFieldValues when model not found", func() {
			expectedErr := rest.NewHTTPError(testENCtx, http.StatusNotFound,
				uerrors.Uniquery_MetricModel_MetricModelNotFound)

			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mmaMock.EXPECT().GetMetricModel(gomock.Any(), gomock.Any()).Return(
				[]interfaces.MetricModel{}, true, nil)

			fieldValues, err := ds.GetMetricModelFieldValues(testENCtx, "test-id", "test-field")

			So(err, ShouldResemble, expectedErr)
			So(fieldValues, ShouldResemble, interfaces.FieldValues{})
		})

		Convey("GetMetricModelFieldValues when model is empty", func() {
			expectedErr := rest.NewHTTPError(testENCtx, http.StatusNotFound,
				uerrors.Uniquery_MetricModel_MetricModelNotFound)

			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mmaMock.EXPECT().GetMetricModel(gomock.Any(), gomock.Any()).Return(
				nil, false, nil)

			fieldValues, err := ds.GetMetricModelFieldValues(testENCtx, "test-id", "test-field")

			So(err, ShouldResemble, expectedErr)
			So(fieldValues, ShouldResemble, interfaces.FieldValues{})
		})

		Convey("GetMetricModelFieldValues when field not found in model fields", func() {
			expectedErr := rest.NewHTTPError(testENCtx, http.StatusBadRequest,
				uerrors.Uniquery_MetricModel_InvalidParameter_FieldName).
				WithErrorDetails("指定的字段[test-field]不属于数据视图[]")

			model := interfaces.MetricModel{
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "",
				},
				MetricType: interfaces.ATOMIC_METRIC,
			}
			dataView := interfaces.DataView{
				FieldsMap: map[string]*cond.ViewField{
					"other-field": {},
				},
			}

			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mmaMock.EXPECT().GetMetricModel(gomock.Any(), gomock.Any()).Return(
				[]interfaces.MetricModel{model}, true, nil)
			dvsMock.EXPECT().GetDataViewByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(
				&dataView, nil)

			fieldValues, err := ds.GetMetricModelFieldValues(testENCtx, "test-id", "test-field")

			So(err, ShouldResemble, expectedErr)
			So(fieldValues, ShouldResemble, interfaces.FieldValues{})
		})

		Convey("GetMetricModelFieldValues when field type is not keyword or text", func() {
			model := interfaces.MetricModel{
				MetricType: interfaces.ATOMIC_METRIC,
				DataSource: &interfaces.MetricDataSource{
					ID: "view1",
				},
			}
			dataView := interfaces.DataView{
				FieldsMap: map[string]*cond.ViewField{
					"test-field": {
						Type: "long",
					},
				},
			}
			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mmaMock.EXPECT().GetMetricModel(gomock.Any(), gomock.Any()).Return(
				[]interfaces.MetricModel{model}, true, nil)
			dvsMock.EXPECT().GetDataViewByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(
				&dataView, nil)

			fieldValues, err := ds.GetMetricModelFieldValues(testENCtx, "test-id", "test-field")

			So(err, ShouldBeNil)
			So(fieldValues, ShouldResemble, interfaces.FieldValues{})
		})

		Convey("GetMetricModelFieldValues when promql get field values failed", func() {
			expectedErr := rest.NewHTTPError(testENCtx, http.StatusInternalServerError,
				uerrors.Uniquery_MetricModel_InternalError_GetFieldValuesFailed).
				WithErrorDetails("promql error")

			model := interfaces.MetricModel{
				ModelID:    "test-id",
				QueryType:  interfaces.PROMQL,
				MetricType: interfaces.ATOMIC_METRIC,
				DataSource: &interfaces.MetricDataSource{
					ID: "view1",
				},
				Formula:    "test-formula",
				UpdateTime: 123,
			}
			dataView := interfaces.DataView{
				FieldsMap: map[string]*cond.ViewField{
					"test-field": {
						Type: dtype.DataType_String,
					},
				},
				UpdateTime: 456,
			}

			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mmaMock.EXPECT().GetMetricModel(gomock.Any(), gomock.Any()).Return(
				[]interfaces.MetricModel{model}, true, nil)
			dvsMock.EXPECT().GetDataViewByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(
				&dataView, nil)

			promQLMock.EXPECT().GetFieldValues(gomock.Any(), gomock.Any(), gomock.Any()).Return(
				nil, 1, errors.New("promql error"))

			fieldValues, err := ds.GetMetricModelFieldValues(testENCtx, "test-id", "test-field")

			So(err, ShouldResemble, expectedErr)
			So(fieldValues, ShouldResemble, interfaces.FieldValues{FieldName: "test-field", Type: dtype.DataType_String})
		})

		Convey("GetMetricModelFieldValues when promql succeeds", func() {
			model := interfaces.MetricModel{
				ModelID:    "test-id",
				QueryType:  interfaces.PROMQL,
				MetricType: interfaces.ATOMIC_METRIC,
				DataSource: &interfaces.MetricDataSource{
					ID: "view1",
				},
				Formula:    "test-formula",
				UpdateTime: 123,
			}
			dataView := interfaces.DataView{
				FieldsMap: map[string]*cond.ViewField{
					"test-field": {
						Type: dtype.DataType_String,
					},
				},
				UpdateTime: 456,
			}

			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mmaMock.EXPECT().GetMetricModel(gomock.Any(), gomock.Any()).Return(
				[]interfaces.MetricModel{model}, true, nil)
			dvsMock.EXPECT().GetDataViewByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(
				&dataView, nil)

			promQLMock.EXPECT().GetFieldValues(gomock.Any(), gomock.Any(), gomock.Any()).Return(
				map[string]bool{"value1": true, "value2": true}, 2, nil)

			fieldValues, err := ds.GetMetricModelFieldValues(testENCtx, "test-id", "test-field")

			So(err, ShouldBeNil)
			So(len(fieldValues.Values), ShouldEqual, 2)
		})

		Convey("GetMetricModelFieldValues success when sdl succeeds", func() {
			model := interfaces.MetricModel{
				ModelID:    "test-id",
				QueryType:  interfaces.DSL,
				MetricType: interfaces.ATOMIC_METRIC,
				DataSource: &interfaces.MetricDataSource{
					ID: "view1",
				},
				Formula: `{"size":0,"aggs": {
					"NAME1": {
					  "terms": {
						"field": "labels.cpu.keyword",
						"size": 20000
					  }
					}
				  }}`,
				UpdateTime: 123,
			}
			dataView := interfaces.DataView{
				FieldsMap: map[string]*cond.ViewField{
					"test-field": {
						Type: dtype.DataType_String,
					},
				},
				UpdateTime: 456,
			}

			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mmaMock.EXPECT().GetMetricModel(gomock.Any(), gomock.Any()).Return(
				[]interfaces.MetricModel{model}, true, nil)
			dvsMock.EXPECT().GetDataViewByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(
				&dataView, nil)

			// Mock first batch query response
			firstResponse := map[string]any{
				"aggregations": map[string]any{
					"fieldV": map[string]any{
						"buckets": []map[string]any{
							{"key": "value1", "doc_count": 1},
							{"key": "value2", "doc_count": 2},
						},
					},
				},
			}

			// Mock second batch query response
			secondResponse := map[string]any{
				"aggregations": map[string]any{
					"fieldV": map[string]any{
						"buckets": []map[string]any{
							{"key": "value3", "doc_count": 3},
							{"key": "value4", "doc_count": 4},
						},
					},
				},
			}

			// Mock third batch query response - empty buckets
			thirdResponse := map[string]any{
				"aggregations": map[string]any{
					"fieldV": map[string]any{
						"buckets": []map[string]any{},
					},
				},
			}

			gomock.InOrder(
				dvsMock.EXPECT().GetDataFromOpenSearch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any()).Return(convert.MapToByte(firstResponse), http.StatusOK, nil),
				dvsMock.EXPECT().GetDataFromOpenSearch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any()).Return(convert.MapToByte(secondResponse), http.StatusOK, nil),
				dvsMock.EXPECT().GetDataFromOpenSearch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any()).Return(convert.MapToByte(thirdResponse), http.StatusOK, nil),
			)

			result, err := ds.GetMetricModelFieldValues(testENCtx, "test-id", "test-field")
			So(err, ShouldBeNil)
			So(len(result.Values), ShouldEqual, 4)
		})
	})
}

func TestProcessDSLFieldValues(t *testing.T) {
	Convey("Test processDSLFieldValues", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				FullCacheRefreshInterval: 24 * time.Hour,
			},
			PromqlSetting: common.PromqlSetting{
				MaxSearchSeriesSize: 10000,
			},
		}
		mmaMock := mock.NewMockMetricModelAccess(mockCtrl)
		ibaMock := mock.NewMockIndexBaseAccess(mockCtrl)
		promQLMock := mock.NewMockPromQLService(mockCtrl)
		dvsMock := mock.NewMockDataViewService(mockCtrl)
		vvsMock := mock.NewMockVegaService(mockCtrl)
		psMock := mock.NewMockPermissionService(mockCtrl)
		ds := MockNewMetricModelService(appSetting, mmaMock, ibaMock, promQLMock, dvsMock, vvsMock, psMock)

		Convey("parseDsl error", func() {
			metricModel := interfaces.MetricModel{
				ModelName:    "test",
				MetricType:   "atomic",
				QueryType:    "dsl",
				Formula:      "invalid formula",
				MeasureField: "test_field",
			}

			fieldValue := interfaces.FieldValues{
				FieldName: "test_field",
				Type:      dtype.DataType_String,
			}

			_, err := ds.processDSLFieldValues(testENCtx, metricModel, &interfaces.DataView{}, fieldValue)
			So(err, ShouldNotBeNil)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Exec failed, because of GetDataFromOpenSearch err", func() {
			dvsMock.EXPECT().GetDataFromOpenSearch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(nil, http.StatusInternalServerError,
				uerrors.NewOpenSearchError(uerrors.InternalServerError).WithReason("Error getting response from opensearch"))

			metricModel := interfaces.MetricModel{
				ModelName:  "txy",
				MetricType: "atomic",
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "主机监控",
				},
				QueryType: "dsl",
				Formula: `{"size":0,"aggs": {
					"NAME1": {
					  "terms": {
						"field": "labels.cpu.keyword",
						"size": 10
					  }
					}
				  }}`,
				DateField:    "@timestamp",
				MeasureField: "value",
				UnitType:     "timeUnit",
				Unit:         "ms",
			}

			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: "en-US",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_InternalError,
					Description:  "Internal Error",
					Solution:     "Please try this operation again, if the error occurs again, submit the work order or contact technical support engineers.",
					ErrorLink:    "None",
					ErrorDetails: "{\"status\":500,\"error\":{\"type\":\"UniQuery.InternalServerError\",\"reason\":\"Error getting response from opensearch\"}}",
				},
			}
			fieldValue := interfaces.FieldValues{}
			_, httpErr := ds.processDSLFieldValues(testENCtx, metricModel, &dataview, fieldValue)
			So(httpErr, ShouldResemble, expectedErr)
		})

		Convey("Exec failed, because of Unmarshal error", func() {
			metricModel := interfaces.MetricModel{
				ModelName:  "txy",
				MetricType: "atomic",
				DataSource: &interfaces.MetricDataSource{
					Type: interfaces.SOURCE_TYPE_DATA_VIEW,
					ID:   "主机监控",
				},
				QueryType:    "dsl",
				Formula:      `invalid json`,
				DateField:    "@timestamp",
				MeasureField: "value",
				UnitType:     "timeUnit",
				Unit:         "ms",
			}

			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: "en-US",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_MetricModel_InternalError_UnmarshalFailed,
					Description:  "Unmarshal Failed",
					Solution:     "Please try this operation again, if the error occurs again, submit the work order or contact technical support engineers.",
					ErrorLink:    "None",
					ErrorDetails: "dsl Unmarshal error: \"Syntax error at index 1: invalid char\\n\\n\\tinvalid json\\n\\t.^..........\\n\"",
				},
			}

			fieldValue := interfaces.FieldValues{}
			_, httpErr := ds.processDSLFieldValues(testENCtx, metricModel, &dataview, fieldValue)
			So(httpErr, ShouldResemble, expectedErr)
		})

		Convey("processDSLFieldValues success with text field and batch query", func() {
			// Mock first batch query response
			firstResponse := map[string]any{
				"aggregations": map[string]any{
					"fieldV": map[string]any{
						"buckets": []map[string]any{
							{"key": "value1", "doc_count": 1},
							{"key": "value2", "doc_count": 2},
						},
					},
				},
			}

			// Mock second batch query response
			secondResponse := map[string]any{
				"aggregations": map[string]any{
					"fieldV": map[string]any{
						"buckets": []map[string]any{
							{"key": "value3", "doc_count": 3},
							{"key": "value4", "doc_count": 4},
						},
					},
				},
			}

			// Mock third batch query response - empty buckets
			thirdResponse := map[string]any{
				"aggregations": map[string]any{
					"fieldV": map[string]any{
						"buckets": []map[string]any{},
					},
				},
			}

			gomock.InOrder(
				dvsMock.EXPECT().GetDataFromOpenSearch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any()).Return(convert.MapToByte(firstResponse), http.StatusOK, nil),
				dvsMock.EXPECT().GetDataFromOpenSearch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any()).Return(convert.MapToByte(secondResponse), http.StatusOK, nil),
				dvsMock.EXPECT().GetDataFromOpenSearch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any()).Return(convert.MapToByte(thirdResponse), http.StatusOK, nil),
			)

			metricModel := interfaces.MetricModel{
				ModelName:  "test",
				MetricType: "atomic",
				QueryType:  "dsl",
				Formula: `{"size":0,"aggs": {
					"NAME1": {
					  "terms": {
						"field": "labels.cpu.keyword",
						"size": 20000
					  }
					}
				  }}`,
				MeasureField: "test_field",
			}
			dataView := interfaces.DataView{
				FieldsMap: map[string]*cond.ViewField{
					"test_field": {
						Name: "test_field",
						Type: "text",
					},
				},
			}

			fieldValue := interfaces.FieldValues{
				FieldName: "test_field",
				Type:      "text",
			}

			result, err := ds.processDSLFieldValues(testENCtx, metricModel, &dataView, fieldValue)
			So(err, ShouldBeNil)
			So(len(result.Values), ShouldEqual, 4)
		})

	})
}

func TestGetMetricModelLabels(t *testing.T) {
	Convey("Test GetMetricModelLabels", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				FullCacheRefreshInterval: 24 * time.Hour,
			},
		}
		mmaMock := mock.NewMockMetricModelAccess(mockCtrl)
		ibaMock := mock.NewMockIndexBaseAccess(mockCtrl)
		promQLMock := mock.NewMockPromQLService(mockCtrl)
		dvsMock := mock.NewMockDataViewService(mockCtrl)
		vvsMock := mock.NewMockVegaService(mockCtrl)
		psMock := mock.NewMockPermissionService(mockCtrl)
		ds := MockNewMetricModelService(appSetting, mmaMock, ibaMock, promQLMock, dvsMock, vvsMock, psMock)

		Convey("GetMetricModelLabels failed with GetMetricModel error", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: "zh-CN",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_MetricModel_InternalError_GetModelByIdFailed,
					Description:  "按ID获取指标模型失败",
					Solution:     "请重试该操作，若再次出现该错误请提交工单或联系技术支持工程师。",
					ErrorLink:    "暂无",
					ErrorDetails: "get metric model failed",
				},
			}

			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mmaMock.EXPECT().GetMetricModel(gomock.Any(), gomock.Any()).Return(
				[]interfaces.MetricModel{}, false, fmt.Errorf("get metric model failed"))

			fields, err := ds.GetMetricModelLabels(testCtx, "test-id")

			So(err, ShouldResemble, expectedErr)
			So(fields, ShouldBeNil)
		})

		Convey("GetMetricModGetMetricModelLabelselFields failed with metric model not found", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusNotFound,
				Language: "zh-CN",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_MetricModel_MetricModelNotFound,
					Description:  "指标模型不存在",
					Solution:     "请检查参数是否正确。",
					ErrorLink:    "暂无",
					ErrorDetails: "",
				},
			}

			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mmaMock.EXPECT().GetMetricModel(gomock.Any(), gomock.Any()).Return(
				[]interfaces.MetricModel{}, true, nil)

			fields, err := ds.GetMetricModelLabels(testCtx, "test-id")

			So(err, ShouldResemble, expectedErr)
			So(fields, ShouldBeNil)
		})

		Convey("GetMetricModelLabels success with DSL query type", func() {

			model := interfaces.MetricModel{
				ModelID:    "test-id",
				QueryType:  interfaces.DSL,
				MetricType: interfaces.ATOMIC_METRIC,
				DataSource: &interfaces.MetricDataSource{
					ID: "view1",
				},
				Formula: `{
						"size": 0,
						"query": {
						  "bool": {
							"filter": [
							  {
								"exists": {
								  "field": "metrics.node_cpu_seconds_total"
								}
							  }
							],
							"must": []
						  }
						},
						"aggs": {
						  "job": {
							"terms": {
							  "field": "labels.job",
							  "size": 1,
							  "order": {
								"_key": "asc"
							  }
							},
							"aggs": {
							  "mode": {
								"filters": {
								  	"errors" :   { "match" : { "body" : "error"   }},
          							"warnings" : { "match" : { "body" : "warning" }}
								},
								"aggs": {
								  "instance": {
									"terms": {
									  "field": "labels.instance",
									  "size": 100,
									  "order": {
										"_key": "desc"
									  }
									},
									"aggs": {
									  "cpu": {
										"terms": {
										  "field": "labels.cpu",
										  "size": 25,
										  "order": {
											"_key": "desc"
										  }
										},
										"aggs": {
										  "time": {
											"date_histogram": {
											  "field": "@timestamp",
											  "fixed_interval": "{{__interval}}",
											  "min_doc_count": 1
											},
											"aggs": {
											  "value": {
												"top_hits": {
													"sort": [
														{
															"date": {
																"order": "desc"
															}
														}
													],
													"_source": {
													"includes": [ "label1", "metrics.node_cpu_seconds_total" ]
													},
													"size": 1 
												}
											  }
											}
										  }
										}
									  }
									}
								  }
								}
							  }
							}
						  }
						}
					  }`,
			}
			dataView := interfaces.DataView{
				FieldsMap: map[string]*cond.ViewField{
					"labels.job":      {Name: "labels.job", Type: dtype.DataType_Text},
					"labels.mode":     {Name: "labels.mode", Type: dtype.DataType_String},
					"labels.instance": {Name: "labels.instance", Type: dtype.DataType_String},
					"labels.cpu":      {Name: "labels.cpu", Type: dtype.DataType_String},
					"label1":          {Name: "label1", Type: dtype.DataType_String},
				},
			}

			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mmaMock.EXPECT().GetMetricModel(gomock.Any(), gomock.Any()).Return(
				[]interfaces.MetricModel{model}, true, nil)
			dvsMock.EXPECT().GetDataViewByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(
				&dataView, nil)

			fields, err := ds.GetMetricModelLabels(testCtx, "test-id")

			So(err, ShouldBeNil)
			So(len(fields), ShouldEqual, 5)
		})

		Convey("GetMetricModelLabels failed with PROMQL GetLabels error", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: "zh-CN",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_MetricModel_InternalError_GetLabelsFailed,
					Description:  "获取指标维度失败",
					Solution:     "请重试该操作，若再次出现该错误请提交工单或联系技术支持工程师。",
					ErrorLink:    "暂无",
					ErrorDetails: "GetLabels error",
				},
			}

			model := interfaces.MetricModel{
				ModelID:    "test-id",
				QueryType:  interfaces.PROMQL,
				MetricType: interfaces.ATOMIC_METRIC,
				DataSource: &interfaces.MetricDataSource{
					ID: "view1",
				},
				Formula: "test_formula",
				Task: &interfaces.MetricTask{
					TaskID: "1",
				},
			}

			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mmaMock.EXPECT().GetMetricModel(gomock.Any(), gomock.Any()).Return(
				[]interfaces.MetricModel{model}, true, nil)
			dvsMock.EXPECT().GetDataViewByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(
				&interfaces.DataView{}, nil)

			promQLMock.EXPECT().GetLabels(gomock.Any(), gomock.Any()).Return(
				nil, http.StatusBadRequest, fmt.Errorf("GetLabels error"))

			fields, err := ds.GetMetricModelLabels(testCtx, "test-id")

			So(err, ShouldResemble, expectedErr)
			So(fields, ShouldBeNil)
		})

		Convey("GetMetricModelLabels success with PROMQL", func() {

			model := interfaces.MetricModel{
				ModelID:    "test-id",
				QueryType:  interfaces.PROMQL,
				MetricType: interfaces.ATOMIC_METRIC,
				DataSource: &interfaces.MetricDataSource{
					ID: "view1",
				},
				Formula: "test_formula",
			}
			dataView := interfaces.DataView{
				FieldsMap: map[string]*cond.ViewField{
					"labels.field1": {Name: "labels.field1", Type: dtype.DataType_String},
					"labels.field2": {Name: "labels.field2", Type: dtype.DataType_String},
				},
			}

			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mmaMock.EXPECT().GetMetricModel(gomock.Any(), gomock.Any()).Return(
				[]interfaces.MetricModel{model}, true, nil)
			dvsMock.EXPECT().GetDataViewByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(
				&dataView, nil)

			promQLMock.EXPECT().GetLabels(gomock.Any(), gomock.Any()).Return(
				map[string]bool{
					"field1": true,
					"field2": true,
				}, http.StatusOK, nil)

			fields, err := ds.GetMetricModelLabels(testCtx, "test-id")

			So(err, ShouldBeNil)
			So(len(fields), ShouldEqual, 2)
		})

	})
}

func TestRewritePromQLFilters(t *testing.T) {
	Convey("Test rewritePromQLFilters", t, func() {

		Convey("should return error when filter_mode=normal with invalid field", func() {
			query := &interfaces.MetricModelQuery{
				MetricModelQueryParameters: interfaces.MetricModelQueryParameters{
					FilterMode: interfaces.FILTER_MODE_NORMAL,
				},
				Filters: []interfaces.Filter{
					{Name: "invalid_field", Value: "test", Operation: interfaces.OPERATION_EQ},
				},
			}
			dataView := &interfaces.DataView{
				FieldsMap: map[string]*cond.ViewField{
					"valid_field": {Name: "valid_field", Type: dtype.DataType_String},
				},
			}

			err := rewritePromQLFilters(testENCtx, query, dataView)
			So(err, ShouldBeNil)
			So(query.Filters, ShouldResemble, []interfaces.Filter{
				{Name: "invalid_field", Value: "test", Operation: interfaces.OPERATION_EQ},
			})
		})

		Convey("should return error when filter_mode=error with invalid field", func() {
			query := &interfaces.MetricModelQuery{
				MetricModelQueryParameters: interfaces.MetricModelQueryParameters{
					FilterMode: interfaces.FILTER_MODE_ERROR,
				},
				Filters: []interfaces.Filter{
					{Name: "invalid_field", Value: "test", Operation: interfaces.OPERATION_EQ},
				},
			}
			dataView := &interfaces.DataView{
				FieldsMap: map[string]*cond.ViewField{
					"valid_field": {Name: "valid_field", Type: dtype.DataType_String},
				},
			}

			err := rewritePromQLFilters(testENCtx, query, dataView)
			So(err, ShouldNotBeNil)
		})

		Convey("should success when filter_mode=ignore with invalid field", func() {
			query := &interfaces.MetricModelQuery{
				MetricModelQueryParameters: interfaces.MetricModelQueryParameters{
					FilterMode: interfaces.FILTER_MODE_IGNORE,
				},
				Filters: []interfaces.Filter{
					{Name: "invalid_field", Value: "test", Operation: interfaces.OPERATION_EQ},
				},
			}
			dataView := &interfaces.DataView{
				FieldsMap: map[string]*cond.ViewField{
					"valid_field": {Name: "valid_field", Type: dtype.DataType_String},
				},
			}

			err := rewritePromQLFilters(testENCtx, query, dataView)
			So(err, ShouldBeNil)
			So(query.Filters, ShouldResemble, []interfaces.Filter{})
		})

		Convey("should success when filter_mode=normal with result field", func() {
			query := &interfaces.MetricModelQuery{
				MetricModelQueryParameters: interfaces.MetricModelQueryParameters{
					FilterMode: interfaces.FILTER_MODE_NORMAL,
				},
				Filters: []interfaces.Filter{
					{Name: "valid_field", Value: "test", Operation: interfaces.OPERATION_EQ, IsResultField: true},
				},
			}
			dataView := &interfaces.DataView{
				FieldsMap: map[string]*cond.ViewField{
					"valid_field": {Name: "valid_field", Type: dtype.DataType_String},
				},
			}

			err := rewritePromQLFilters(testENCtx, query, dataView)
			So(err, ShouldBeNil)
			So(query.Filters, ShouldResemble, []interfaces.Filter{
				{Name: "labels.valid_field", Value: "test", Operation: interfaces.OPERATION_EQ, IsResultField: true},
			})
		})
	})
}

func TestRewriteDSLFilters(t *testing.T) {
	Convey("Test rewriteDSLFilters", t, func() {

		Convey("should return error when filter_mode=normal with invalid field", func() {
			query := &interfaces.MetricModelQuery{
				MetricModelQueryParameters: interfaces.MetricModelQueryParameters{
					FilterMode: interfaces.FILTER_MODE_NORMAL,
				},
				Filters: []interfaces.Filter{
					{Name: "invalid_field", Value: "test", Operation: interfaces.OPERATION_EQ},
				},
			}

			err := rewriteDSLFilters(testENCtx, query, interfaces.DslInfo{}, &interfaces.DataView{})
			So(err, ShouldBeNil)
			So(query.Filters, ShouldResemble, []interfaces.Filter{
				{Name: "invalid_field", Value: "test", Operation: interfaces.OPERATION_EQ},
			})
		})

		Convey("should return error when filter_mode=error with invalid field", func() {
			query := &interfaces.MetricModelQuery{
				MetricModelQueryParameters: interfaces.MetricModelQueryParameters{
					FilterMode: interfaces.FILTER_MODE_ERROR,
				},
				Filters: []interfaces.Filter{
					{Name: "invalid_field", Value: "test", Operation: interfaces.OPERATION_EQ},
				},
			}

			err := rewriteDSLFilters(testENCtx, query, interfaces.DslInfo{}, &interfaces.DataView{})
			So(err, ShouldNotBeNil)
		})

		Convey("should success when filter_mode=normal with result field", func() {
			query := &interfaces.MetricModelQuery{
				MetricModelQueryParameters: interfaces.MetricModelQueryParameters{
					FilterMode: interfaces.FILTER_MODE_NORMAL,
				},
				Filters: []interfaces.Filter{
					{Name: "job", Value: "test", Operation: interfaces.OPERATION_EQ, IsResultField: true},
				},
				Formula: `{
					"size": 0,
					"query": {
					  "bool": {
						"filter": [
						  {
							"exists": {
							  "field": "metrics.node_cpu_seconds_total"
							}
						  }
						],
						"must": []
					  }
					},
					"aggs": {
					  "job": {
						"terms": {
						  "field": "labels.job",
						  "size": 1,
						  "order": {
							"_key": "asc"
						  }
						},
						"aggs": {
						  "mode": {
							"filters": {
								  "errors" :   { "match" : { "body" : "error"   }},
								  "warnings" : { "match" : { "body" : "warning" }}
							},
							"aggs": {
							  "instance": {
								"terms": {
								  "field": "labels.instance",
								  "size": 100,
								  "order": {
									"_key": "desc"
								  }
								},
								"aggs": {
								  "cpu": {
									"terms": {
									  "field": "labels.cpu",
									  "size": 25,
									  "order": {
										"_key": "desc"
									  }
									},
									"aggs": {
									  "time": {
										"date_histogram": {
										  "field": "@timestamp",
										  "fixed_interval": "{{__interval}}",
										  "min_doc_count": 1
										},
										"aggs": {
										  "value": {
											"top_hits": {
												"sort": [
													{
														"date": {
															"order": "desc"
														}
													}
												],
												"_source": {
												"includes": [ "label1", "metrics.node_cpu_seconds_total" ]
												},
												"size": 1 
											}
										  }
										}
									  }
									}
								  }
								}
							  }
							}
						  }
						}
					  }
					}
				  }`,
			}
			dslInfo := interfaces.DslInfo{
				AggInfos: map[int]interfaces.AggInfo{
					0: {AggName: "job", AggType: interfaces.BUCKET_TYPE_TERMS, TermsField: "labels.job"},
					1: {AggName: "mode", AggType: interfaces.BUCKET_TYPE_FILTERS, TermsField: ""},
					2: {AggName: "instance", AggType: interfaces.BUCKET_TYPE_TERMS, TermsField: "labels.instance"},
					3: {AggName: "cpu", AggType: interfaces.BUCKET_TYPE_TERMS, TermsField: "labels.cpu"},
					4: {AggName: "time", AggType: interfaces.BUCKET_TYPE_DATE_HISTOGRAM, TermsField: ""},
					5: {AggName: "value", AggType: interfaces.AGGR_TYPE_TOP_HITS, IncludeFields: []string{"label1", "metrics.node_cpu_seconds_total"}},
				},
			}

			err := rewriteDSLFilters(testENCtx, query, dslInfo, &interfaces.DataView{})
			So(err, ShouldBeNil)
			So(query.Filters, ShouldResemble, []interfaces.Filter{
				{Name: "labels.job", Value: "test", Operation: interfaces.OPERATION_EQ, IsResultField: true},
			})
		})

		Convey("should success when filter_mode=normal with result field is filters", func() {
			query := &interfaces.MetricModelQuery{
				MetricModelQueryParameters: interfaces.MetricModelQueryParameters{
					FilterMode: interfaces.FILTER_MODE_NORMAL,
				},
				Filters: []interfaces.Filter{
					{Name: "mode", Value: "test", Operation: interfaces.OPERATION_EQ, IsResultField: true},
				},
				Formula: `{
					"size": 0,
					"query": {
					  "bool": {
						"filter": [
						  {
							"exists": {
							  "field": "metrics.node_cpu_seconds_total"
							}
						  }
						],
						"must": []
					  }
					},
					"aggs": {
					  "job": {
						"terms": {
						  "field": "labels.job",
						  "size": 1,
						  "order": {
							"_key": "asc"
						  }
						},
						"aggs": {
						  "mode": {
							"filters": {
								  "errors" :   { "match" : { "body" : "error"   }},
								  "warnings" : { "match" : { "body" : "warning" }}
							},
							"aggs": {
							  "instance": {
								"terms": {
								  "field": "labels.instance",
								  "size": 100,
								  "order": {
									"_key": "desc"
								  }
								},
								"aggs": {
								  "cpu": {
									"terms": {
									  "field": "labels.cpu",
									  "size": 25,
									  "order": {
										"_key": "desc"
									  }
									},
									"aggs": {
									  "time": {
										"date_histogram": {
										  "field": "@timestamp",
										  "fixed_interval": "{{__interval}}",
										  "min_doc_count": 1
										},
										"aggs": {
										  "value": {
											"top_hits": {
												"sort": [
													{
														"date": {
															"order": "desc"
														}
													}
												],
												"_source": {
												"includes": [ "label1", "metrics.node_cpu_seconds_total" ]
												},
												"size": 1 
											}
										  }
										}
									  }
									}
								  }
								}
							  }
							}
						  }
						}
					  }
					}
				  }`,
			}
			dslInfo := interfaces.DslInfo{
				AggInfos: map[int]interfaces.AggInfo{
					0: {AggName: "job", AggType: interfaces.BUCKET_TYPE_TERMS, TermsField: "labels.job"},
					1: {AggName: "mode", AggType: interfaces.BUCKET_TYPE_FILTERS, TermsField: ""},
					2: {AggName: "instance", AggType: interfaces.BUCKET_TYPE_TERMS, TermsField: "labels.instance"},
					3: {AggName: "cpu", AggType: interfaces.BUCKET_TYPE_TERMS, TermsField: "labels.cpu"},
					4: {AggName: "time", AggType: interfaces.BUCKET_TYPE_DATE_HISTOGRAM, TermsField: ""},
					5: {AggName: "value", AggType: interfaces.AGGR_TYPE_TOP_HITS, IncludeFields: []string{"label1", "metrics.node_cpu_seconds_total"}},
				},
			}

			err := rewriteDSLFilters(testENCtx, query, dslInfo, &interfaces.DataView{})
			So(err, ShouldNotBeNil)
		})
	})
}
