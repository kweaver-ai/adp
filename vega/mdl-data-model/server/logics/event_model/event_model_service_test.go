// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package event_model

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bmizerany/assert"
	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	"data-model/common"
	derrors "data-model/errors"
	"data-model/interfaces"
	dmock "data-model/interfaces/mock"
)

var (
	testCtx        = context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage)
	testUpdateTime = int64(1735786555379)
	testFormula    = "avg(rate(node_cpu_seconds_total{name=~\"node-14-35\",mode12=\"system\"}[5m]))by(name)"

	detectRule = interfaces.DetectRule{
		DetectRuleID: "1000",
		Priority:     99,
		Type:         "range_detect",
		Formula: []interfaces.FormulaItem{
			{
				Level: 1,
				Filter: interfaces.LogicFilter{
					LogicOperator: "",
					FilterExpress: interfaces.FilterExpress{Name: "values", Value: []float64{0.9, 1.0}, Operation: "range"},
					Children:      []interfaces.LogicFilter{},
				},
			},
			{
				Level: 2,
				Filter: interfaces.LogicFilter{
					LogicOperator: "",
					FilterExpress: interfaces.FilterExpress{Name: "values", Value: []float64{0.9, 1.0}, Operation: "range"},
					Children:      []interfaces.LogicFilter{},
				},
			},
		},
	}
	illeageDetectRule = interfaces.DetectRule{
		DetectRuleID: "1000",
		Priority:     99,
		Type:         "range_detect",
		Formula: []interfaces.FormulaItem{
			{
				Level: 1,
				Filter: interfaces.LogicFilter{
					LogicOperator: "",
					FilterExpress: interfaces.FilterExpress{Name: "values", Value: 1.0, Operation: "range"},
					Children:      []interfaces.LogicFilter{},
				},
			},
			{
				Level: 2,
				Filter: interfaces.LogicFilter{
					LogicOperator: "",
					FilterExpress: interfaces.FilterExpress{Name: "values", Value: []float64{0.9, 1.0}, Operation: "range"},
					Children:      []interfaces.LogicFilter{},
				},
			},
		},
		UpdateTime: testUpdateTime,
	}
	// eventModel = interfaces.EventModelCreateRequest{
	// 	EventModelRequest: interfaces.EventModelRequest{
	// 		EventModelName:    "测试中的名称",
	// 		EventModelType:    "atomic",
	// 		// UpdateTime:        testUpdateTime,
	// 		EventModelTags:    []string{"xx1", "xx2"},
	// 		DataSourceType:    "metric_model",
	// 		DataSource:        "1",
	// 		DetectRule:        detectRule,
	// 		DefaultTimeWindow: "5m",
	// 		EventModelComment: ""},
	// 	IsActive: -1,
	// 	IsCustom: -1,
	// }

	oldEventModel = interfaces.EventModel{

		EventModelName:      "测试中的名称",
		EventModelType:      "atomic",
		UpdateTime:          testUpdateTime,
		EventModelTags:      []string{"xx1", "xx2"},
		DataSourceType:      "metric_model",
		DataSource:          []string{"1"},
		DataSourceName:      []string{"1"},
		DataSourceGroupName: []string{"1"},
		DetectRule:          detectRule,
		DefaultTimeWindow:   interfaces.TimeInterval{Interval: 5, Unit: "m"},
		EventModelComment:   "",
		IsActive:            1,
		IsCustom:            -1,
		Task: interfaces.EventTask{
			Schedule: interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
			StorageConfig: interfaces.StorageConfig{
				IndexBase:    "base1",
				DataViewName: "view1",
				DataViewID:   "__base1",
			},
			DispatchConfig: interfaces.DispatchConfig{
				TimeOut:        3600,
				RouteStrategy:  "FIRST",
				BlockStrategy:  "",
				FailRetryCount: 3,
			},
			ExecuteParameter: map[string]any{},
		},
	}

	illeagalOldEventModel = interfaces.EventModel{

		EventModelName:    "测试中的名称",
		EventModelType:    "atomic",
		UpdateTime:        testUpdateTime,
		EventModelTags:    []string{"xx1", "xx2"},
		DataSourceType:    "metric_model",
		DataSource:        []string{"1"},
		DetectRule:        detectRule,
		DefaultTimeWindow: interfaces.TimeInterval{Interval: -20, Unit: "d"},
		EventModelComment: "",
		IsActive:          -1,
		IsCustom:          -1,
	}

	newEventModel = interfaces.EventModelUpateRequest{
		EventModelRequest: interfaces.EventModelRequest{
			EventModelName: "测试中的名称_修改",
			EventModelType: "atomic",
			// UpdateTime:        "2022-12-13 01:01:01",
			EventModelTags: []string{"xx1", "xx2", "xx3"},
			DataSourceType: "metric_model",
			DataSource:     []string{"2"},
			DetectRule:     detectRule,
			AggregateRule:  interfaces.AggregateRule{},
			DefaultTimeWindow: interfaces.TimeInterval{
				Interval: 5,
				Unit:     "m",
			},
			EventModelComment: "modify_comment",
			EventTaskRequest: interfaces.EventTaskRequest{
				Schedule: interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
				StorageConfig: interfaces.StorageConfig{
					IndexBase:    "base1",
					DataViewName: "view1",
				},
				DispatchConfig: interfaces.DispatchConfig{
					TimeOut:        3600,
					RouteStrategy:  "FIRST",
					BlockStrategy:  "",
					FailRetryCount: 3,
				},
				ExecuteParameter: map[string]any{},
			},
		},
		EventModelID: "1",
		IsActive:     1,
	}

	metricModelT = interfaces.MetricModel{
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

	eventModelQueryReq = interfaces.EventModelQueryRequest{
		EventModelName: "测试中的名称",
		EventModelType: "atomic",
		EventModelTag:  "xx1",
		IsActive:       "",
		IsCustom:       -1,
		Direction:      "asc",
		SortKey:        "update_time",
		Limit:          10,
		Offset:         0,
	}

	eventTask = interfaces.EventTask{
		TaskID:   "1",
		ModelID:  "1",
		Schedule: interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
		StorageConfig: interfaces.StorageConfig{
			IndexBase:    "base1",
			DataViewName: "view1",
		},
		DispatchConfig: interfaces.DispatchConfig{
			TimeOut:        3600,
			RouteStrategy:  "ROUND",
			BlockStrategy:  "SERIAL_EXECUTION",
			FailRetryCount: 1,
		},
		ExecuteParameter:   map[string]any{},
		TaskStatus:         4,
		StatusUpdateTime:   testUpdateTime,
		ErrorDetails:       "",
		ScheduleSyncStatus: 3,
		UpdateTime:         testUpdateTime,
	}
	// eventScheduleBytes, _    = json.Marshal(eventTask.Schedule)
	// storageConfigBytes, _    = json.Marshal(eventTask.StorageConfig)
	// dispatchConfigBytes, _   = json.Marshal(eventTask.DispatchConfig)
	// executeParameterBytes, _ = json.Marshal(eventTask.ExecuteParameter)
)

func MockNewEventModelService(appSetting *common.AppSetting,
	dmja interfaces.DataModelJobAccess,
	ema interfaces.EventModelAccess,
	mms interfaces.MetricModelService,
	iba interfaces.IndexBaseAccess,
	dvs interfaces.DataViewService,
	ps interfaces.PermissionService) (*eventModelService, sqlmock.Sqlmock) {

	db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	ems := &eventModelService{
		appSetting: appSetting,
		dmja:       dmja,
		ema:        ema,
		iba:        iba,
		mms:        mms,
		dvs:        dvs,
		db:         db,
		ps:         ps,
	}
	return ems, smock
}

func Test_EventModelService_CheckEventModelExistByName(t *testing.T) {
	type fields struct {
		appSetting *common.AppSetting
		dmja       *dmock.MockDataModelJobAccess
		ema        *dmock.MockEventModelAccess
		iba        *dmock.MockIndexBaseAccess
		mms        *dmock.MockMetricModelService
		dvs        *dmock.MockDataViewService
		ps         *dmock.MockPermissionService
	}
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	field := fields{
		appSetting: &common.AppSetting{},
		ema:        dmock.NewMockEventModelAccess(mockCtrl),
	}

	type args struct {
		ctx       context.Context
		modelName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
		{"exist", field, args{context.TODO(), "测试中的名称"}, true, false},
		{"not exist", field, args{context.TODO(), "测试中的名称_修改"}, false, false},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			// ems := &eventModelService{
			// 	appSetting:       tt.fields.appSetting,
			// 	ema: tt.fields.ema,
			// }
			ems, _ := MockNewEventModelService(tt.fields.appSetting, tt.fields.dmja, tt.fields.ema, tt.fields.mms,
				tt.fields.iba, tt.fields.dvs, tt.fields.ps)
			if tt.name == "not exist" {
				field.ema.EXPECT().QueryEventModels(gomock.Any(), gomock.Any()).Return([]interfaces.EventModel{}, nil)

			} else {
				field.ema.EXPECT().QueryEventModels(gomock.Any(), gomock.Any()).Return([]interfaces.EventModel{oldEventModel}, nil)
			}

			isExist, err := ems.CheckEventModelExistByName(tt.args.ctx, tt.args.modelName)

			if (err != nil) != tt.wantErr {
				t.Errorf("eventModelService.CheckEventModelExistByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if isExist != tt.want {
				t.Errorf("eventModelService.CheckEventModelExistByName() = %v, want %v", isExist, tt.want)
			}
		})
	}
}

func Test_EventModelService_ValidateEventModelDataSource(t *testing.T) {
	type fields struct {
		appSetting *common.AppSetting
		dmja       *dmock.MockDataModelJobAccess
		ema        *dmock.MockEventModelAccess
		iba        *dmock.MockIndexBaseAccess
		mms        *dmock.MockMetricModelService
		dvs        *dmock.MockDataViewService
		ps         *dmock.MockPermissionService
	}
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	field := fields{
		appSetting: &common.AppSetting{},
		ema:        dmock.NewMockEventModelAccess(mockCtrl),
		mms:        dmock.NewMockMetricModelService(mockCtrl),
	}
	type args struct {
		ctx                 context.Context
		dataSource          string
		dataSourceType      string
		dataSourceName      string
		dataSourceGroupName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"exist", field, args{context.TODO(), "1", "metric_model", "", ""}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ems, _ := MockNewEventModelService(tt.fields.appSetting, tt.fields.dmja, tt.fields.ema, tt.fields.mms,
				tt.fields.iba, tt.fields.dvs, tt.fields.ps)
			field.mms.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(metricModelT, nil)
			if _, err := ems.ValidateEventModelDataSource(tt.args.ctx, tt.args.dataSource, tt.args.dataSourceType, tt.args.dataSourceName, tt.args.dataSourceGroupName); (err != nil) != tt.wantErr {
				t.Errorf("eventModelService.ValidateEventModelDataSource() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_EventModelService_ValidateEventModelDetectRule(t *testing.T) {

	type fields struct {
		appSetting *common.AppSetting
		dmja       *dmock.MockDataModelJobAccess
		ema        *dmock.MockEventModelAccess
		iba        *dmock.MockIndexBaseAccess
		mms        *dmock.MockMetricModelService
		dvs        *dmock.MockDataViewService
		ps         *dmock.MockPermissionService
	}
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	field := fields{
		appSetting: &common.AppSetting{},
		ema:        dmock.NewMockEventModelAccess(mockCtrl),
	}

	type args struct {
		ctx            context.Context
		detectRule     interfaces.DetectRule
		detectRuleType string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"legal", field, args{context.TODO(), detectRule, "range_detect"}, true},
		{"illegal", field, args{context.TODO(), illeageDetectRule, "range_detect"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ems, _ := MockNewEventModelService(tt.fields.appSetting, tt.fields.dmja, tt.fields.ema, tt.fields.mms,
				tt.fields.iba, tt.fields.dvs, tt.fields.ps)
			err := ems.ValidateEventModelDetectRule(tt.args.ctx, tt.args.detectRule, tt.args.detectRuleType)
			if tt.name == "legal" {
				assert.Equal(t, err, (*rest.HTTPError)(nil))
			} else {
				assert.Equal(t, err.BaseError.ErrorCode, derrors.EventModel_InvalidParameter)
			}
		})
	}
}

func Test_EventModelService_EventModelCreateValidate(t *testing.T) {
	type fields struct {
		appSetting *common.AppSetting
		dmja       *dmock.MockDataModelJobAccess
		ema        *dmock.MockEventModelAccess
		iba        *dmock.MockIndexBaseAccess
		mms        *dmock.MockMetricModelService
		dvs        *dmock.MockDataViewService
		ps         *dmock.MockPermissionService
	}
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	field := fields{
		appSetting: &common.AppSetting{},
		dmja:       dmock.NewMockDataModelJobAccess(mockCtrl),
		ema:        dmock.NewMockEventModelAccess(mockCtrl),
		iba:        dmock.NewMockIndexBaseAccess(mockCtrl),
		mms:        dmock.NewMockMetricModelService(mockCtrl),
		dvs:        dmock.NewMockDataViewService(mockCtrl),

		// ems:  dmock.NewMockEventModelService(mockCtrl),
	}

	type args struct {
		ctx        context.Context
		eventModel interfaces.EventModel
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{"success", field, args{context.TODO(), oldEventModel}},
		{"failed", field, args{context.TODO(), illeagalOldEventModel}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ems, _ := MockNewEventModelService(tt.fields.appSetting, tt.fields.dmja, tt.fields.ema, tt.fields.mms,
				tt.fields.iba, tt.fields.dvs, tt.fields.ps)
			field.mms.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).AnyTimes().Return(metricModelT, nil)
			if tt.name == "success" {
				field.ema.EXPECT().QueryEventModels(gomock.Any(), gomock.Any()).AnyTimes().Return([]interfaces.EventModel{}, nil)
				field.iba.EXPECT().GetSimpleIndexBasesByTypes(gomock.Any(), gomock.Any()).AnyTimes().Return([]interfaces.SimpleIndexBase{
					{
						BaseType: "a",
					},
				}, nil)
				expectedErr := errors.New("some error")
				expectedHttpErr := rest.NewHTTPError(tt.args.ctx, http.StatusInternalServerError, derrors.DataModel_DataView_Existed_ViewName).
					WithErrorDetails(expectedErr.Error())

				field.dvs.EXPECT().CheckDataViewExistByID(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return("", true, expectedHttpErr)
			} else {
				field.ema.EXPECT().QueryEventModels(gomock.Any(), gomock.Any()).AnyTimes().Return([]interfaces.EventModel{}, nil)
			}
			_, httpErr := ems.EventModelCreateValidate(tt.args.ctx, tt.args.eventModel)

			if tt.name == "success" {
				assert.Equal(t, httpErr.BaseError.ErrorCode, derrors.EventModel_DataSourceIllegal)
			} else {
				assert.Equal(t, httpErr.BaseError.ErrorCode, derrors.EventModel_InvalidParameter)
			}

		})
	}
}

func Test_EventModelService_EventModelUpdateValidate(t *testing.T) {
	Convey("Test EventModelUpdateValidate", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
		ema := dmock.NewMockEventModelAccess(mockCtrl)
		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		ems, _ := MockNewEventModelService(appSetting, dmja, ema, mms, iba, dvs, ps)

		err := errors.New("error")

		Convey("EventModelUpdateValidate failed,caused by GetEventModelByID failed ", func() {
			ema.EXPECT().GetEventModelByID(gomock.Any()).Return(oldEventModel, err)
			httpErr := ems.EventModelUpdateValidate(testCtx, newEventModel)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, derrors.EventModel_InternalError)
		})
		Convey("EventModelUpdateValidate failed,caused by GetEventTaskByModelID failed ", func() {
			ema.EXPECT().GetEventModelByID(gomock.Any()).Return(oldEventModel, nil)
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).AnyTimes().Return(map[string]interfaces.SimpleMetricModel{
				"1": {
					ModelName: "1",
					GroupName: "1",
				},
			}, nil)
			ema.EXPECT().GetEventTaskByModelID(gomock.Any(), gomock.Any()).Return(interfaces.EventTask{}, false, err)
			httpErr := ems.EventModelUpdateValidate(testCtx, newEventModel)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, derrors.EventModel_InternalError)
		})

		Convey("EventModelUpdateValidate failed,caused by CheckEventModelExistByName failed ", func() {
			ema.EXPECT().GetEventModelByID(gomock.Any()).Return(oldEventModel, nil)
			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).AnyTimes().Return(map[string]interfaces.SimpleMetricModel{
				"1": {
					ModelName: "1",
					GroupName: "1",
				},
			}, nil)
			ema.EXPECT().GetEventTaskByModelID(gomock.Any(), gomock.Any()).Return(eventTask, true, nil)
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			ema.EXPECT().QueryEventModels(gomock.Any(), gomock.Any()).Return([]interfaces.EventModel{}, err)
			httpErr := ems.EventModelUpdateValidate(testCtx, newEventModel)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, derrors.EventModel_InvalidParameter)
		})
		Convey("EventModelUpdateValidate failed,caused by BatchValidateDataSources failed ", func() {
			ema.EXPECT().GetEventModelByID(gomock.Any()).Return(oldEventModel, nil)
			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).AnyTimes().Return(map[string]interfaces.SimpleMetricModel{
				"1": {
					ModelName: "1",
					GroupName: "1",
				},
			}, nil)
			ema.EXPECT().GetEventTaskByModelID(gomock.Any(), gomock.Any()).Return(eventTask, true, nil)

			ema.EXPECT().QueryEventModels(gomock.Any(), gomock.Any()).Return([]interfaces.EventModel{}, nil)
			mms.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(interfaces.MetricModel{}, err)
			mms.EXPECT().GetMetricModelIDByName(gomock.Any(), gomock.Any(), gomock.Any()).Return("0", err)
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			httpErr := ems.EventModelUpdateValidate(testCtx, newEventModel)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, derrors.EventModel_DataSourceIllegal)
		})
		Convey("EventModelUpdateValidate failed,caused by GetSimpleIndexBasesByTypes failed ", func() {
			ema.EXPECT().GetEventModelByID(gomock.Any()).Return(oldEventModel, nil)
			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).AnyTimes().Return(map[string]interfaces.SimpleMetricModel{
				"1": {
					ModelName: "1",
					GroupName: "1",
				},
			}, nil)
			ema.EXPECT().GetEventTaskByModelID(gomock.Any(), gomock.Any()).Return(eventTask, true, nil)
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			ema.EXPECT().QueryEventModels(gomock.Any(), gomock.Any()).Return([]interfaces.EventModel{}, nil)
			mms.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelID: "1",
				},
			}, nil)

			iba.EXPECT().GetSimpleIndexBasesByTypes(gomock.Any(), gomock.Any()).Return([]interfaces.SimpleIndexBase{}, err)
			httpErr := ems.EventModelUpdateValidate(testCtx, newEventModel)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, derrors.EventModel_TaskSyncCreateFailed)
		})
		Convey("EventModelUpdateValidate failed,caused by CheckDataViewExistByID failed ", func() {
			ema.EXPECT().GetEventModelByID(gomock.Any()).Return(oldEventModel, nil)
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).AnyTimes().Return(map[string]interfaces.SimpleMetricModel{
				"1": {
					ModelName: "1",
					GroupName: "1",
				},
			}, nil)
			ema.EXPECT().GetEventTaskByModelID(gomock.Any(), gomock.Any()).Return(eventTask, true, nil)

			ema.EXPECT().QueryEventModels(gomock.Any(), gomock.Any()).Return([]interfaces.EventModel{}, nil)
			mms.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelID: "1",
				},
			}, nil)

			iba.EXPECT().GetSimpleIndexBasesByTypes(gomock.Any(), gomock.Any()).Return([]interfaces.SimpleIndexBase{
				{
					BaseType: "a",
				},
			}, nil)
			expectedErr := errors.New("some error")
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.EventModel_InternalError).
				WithErrorDetails(expectedErr.Error())
			dvs.EXPECT().CheckDataViewExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return("", true, expectedHttpErr)

			httpErr := ems.EventModelUpdateValidate(testCtx, newEventModel)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, derrors.DataModel_DataView_InternalError_CheckViewIfExistFailed)
		})
		Convey("EventModelUpdateValidate success", func() {
			ema.EXPECT().GetEventModelByID(gomock.Any()).Return(oldEventModel, nil)
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).AnyTimes().Return(map[string]interfaces.SimpleMetricModel{
				"1": {
					ModelName: "1",
					GroupName: "1",
				},
			}, nil)
			ema.EXPECT().GetEventTaskByModelID(gomock.Any(), gomock.Any()).Return(eventTask, true, nil)

			ema.EXPECT().QueryEventModels(gomock.Any(), gomock.Any()).Return([]interfaces.EventModel{}, nil)
			mms.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(interfaces.MetricModel{
				SimpleMetricModel: interfaces.SimpleMetricModel{
					ModelID: "1",
				},
			}, nil)

			iba.EXPECT().GetSimpleIndexBasesByTypes(gomock.Any(), gomock.Any()).Return([]interfaces.SimpleIndexBase{
				{
					BaseType: "a",
				},
			}, nil)
			dvs.EXPECT().CheckDataViewExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return("", true, nil)

			httpErr := ems.EventModelUpdateValidate(testCtx, newEventModel)
			So(httpErr, ShouldBeNil)
		})
	})
}

func Test_EventModelService_CreateEventModels(t *testing.T) {
	Convey("Test CreateEventModels", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
		ema := dmock.NewMockEventModelAccess(mockCtrl)
		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		ems, smock := MockNewEventModelService(appSetting, dmja, ema, mms, iba, dvs, ps)

		// metricModel := interfaces.MetricModel{
		// 	SimpleMetricModel: interfaces.SimpleMetricModel{
		// 		ModelID: "id1",
		// 	},
		// }
		// indexbase := []interfaces.SimpleIndexBase{
		// 	{
		// 		BaseType: "index",
		// 	},
		// }
		Convey("CreateEventModels succeed", func() {
			models := []interfaces.EventModel{oldEventModel}
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin()
			ema.EXPECT().CreateEventModels(gomock.Any(), gomock.Any()).Return([]map[string]any{}, nil)
			ema.EXPECT().CreateEventTask(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			// ema.EXPECT().QueryEventModels(gomock.Any(), gomock.Any()).Return(nil, nil)
			// mms.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(metricModel, nil)
			// iba.EXPECT().GetSimpleIndexBasesByTypes(gomock.Any(), gomock.Any()).Return(indexbase, nil)
			// dvs.EXPECT().CheckDataViewExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return("", true, nil)
			ps.EXPECT().CreateResources(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()
			_, httpErr := ems.CreateEventModels(testCtx, models)
			So(httpErr, ShouldBeNil)
		})

		Convey("CreateEventModels failed ,caused by transaction begin failed", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin().WillReturnError(errors.New("error"))
			models := []interfaces.EventModel{oldEventModel}
			_, err := ems.CreateEventModels(testCtx, models)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, derrors.DataModel_EventModel_InternalError_BeginTransactionFailed)
		})

		Convey("CreateEventModels failed ,caused by createEventModels failed", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin()
			// ema.EXPECT().QueryEventModels(gomock.Any(), gomock.Any()).Return(nil, nil)
			// mms.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(metricModel, nil)
			// iba.EXPECT().GetSimpleIndexBasesByTypes(gomock.Any(), gomock.Any()).Return(indexbase, nil)
			// dvs.EXPECT().CheckDataViewExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return("", true, nil)
			ema.EXPECT().CreateEventModels(gomock.Any(), gomock.Any()).Return([]map[string]any{}, errors.New("error"))
			models := []interfaces.EventModel{oldEventModel}
			_, err := ems.CreateEventModels(testCtx, models)

			httpErr := err.(*rest.HTTPError)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, derrors.EventModel_InternalError)
		})

		Convey("CreateEventModels failed ,caused by createEventTask failed", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin()
			// ema.EXPECT().QueryEventModels(gomock.Any(), gomock.Any()).Return(nil, nil)
			// mms.EXPECT().GetMetricModelByModelID(gomock.Any(), gomock.Any()).Return(metricModel, nil)
			// iba.EXPECT().GetSimpleIndexBasesByTypes(gomock.Any(), gomock.Any()).Return(indexbase, nil)
			// dvs.EXPECT().CheckDataViewExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return("", true, nil)
			ema.EXPECT().CreateEventModels(gomock.Any(), gomock.Any()).Return([]map[string]any{}, nil)
			ema.EXPECT().CreateEventTask(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(errors.New("error"))
			models := []interfaces.EventModel{oldEventModel}
			_, err := ems.CreateEventModels(testCtx, models)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, derrors.EventModel_InternalError)
		})
	})
}

func Test_EventModelService_UpdateEventModel(t *testing.T) {
	Convey("Test UpdateEventModel", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
		ema := dmock.NewMockEventModelAccess(mockCtrl)
		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		ems, smock := MockNewEventModelService(appSetting, dmja, ema, mms, iba, dvs, ps)

		Convey("UpdateEventModel failed, caused by GetEventModelByID failed ", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ema.EXPECT().GetEventModelByID(gomock.Any()).Return(oldEventModel, errors.New("error"))
			err := ems.UpdateEventModel(testCtx, newEventModel)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, derrors.EventModel_InternalError)
		})
		Convey("UpdateEventModel failed, caused by GetEventTaskByModelID failed ", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			ema.EXPECT().GetEventModelByID(gomock.Any()).Return(oldEventModel, nil)
			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.SimpleMetricModel{
				"1": {
					ModelName: "1",
					GroupName: "1",
				},
			}, nil)
			ema.EXPECT().GetEventTaskByModelID(gomock.Any(), gomock.Any()).Return(eventTask, false, errors.New("error"))
			err := ems.UpdateEventModel(testCtx, newEventModel)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, derrors.EventModel_InternalError)
		})

		Convey("UpdateEventModel failed, caused by transaction begin failed failed ", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			ema.EXPECT().GetEventModelByID(gomock.Any()).Return(oldEventModel, nil)
			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.SimpleMetricModel{
				"1": {
					ModelName: "1",
					GroupName: "1",
				},
			}, nil)
			ema.EXPECT().GetEventTaskByModelID(gomock.Any(), gomock.Any()).Return(eventTask, true, nil)
			smock.ExpectBegin().WillReturnError(errors.New("error"))
			err := ems.UpdateEventModel(testCtx, newEventModel)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, derrors.DataModel_EventModel_InternalError_BeginTransactionFailed)
		})
		Convey("UpdateEventModel failed, caused by UpdateEventModel failed ", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			ema.EXPECT().GetEventModelByID(gomock.Any()).Return(oldEventModel, nil)
			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.SimpleMetricModel{
				"1": {
					ModelName: "1",
					GroupName: "1",
				},
			}, nil)
			ema.EXPECT().GetEventTaskByModelID(gomock.Any(), gomock.Any()).Return(eventTask, true, nil)
			smock.ExpectBegin()
			ema.EXPECT().UpdateEventModel(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			err := ems.UpdateEventModel(testCtx, newEventModel)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, derrors.EventModel_InternalError)
		})
		Convey("UpdateEventModel failed, caused by UpdateEventTask failed ", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			ema.EXPECT().GetEventModelByID(gomock.Any()).Return(oldEventModel, nil)
			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.SimpleMetricModel{
				"1": {
					ModelName: "1",
					GroupName: "1",
				},
			}, nil)
			ema.EXPECT().GetEventTaskByModelID(gomock.Any(), gomock.Any()).Return(eventTask, true, nil)

			smock.ExpectBegin()
			ema.EXPECT().UpdateEventModel(gomock.Any(), gomock.Any()).Return(nil)
			ema.EXPECT().GetEventTaskByModelID(gomock.Any(), gomock.Any()).Return(eventTask, true, nil)
			ema.EXPECT().UpdateEventTask(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("error"))
			err := ems.UpdateEventModel(testCtx, newEventModel)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, derrors.EventModel_InternalError)
		})
		Convey("UpdateEventModel succeed", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			ps.EXPECT().UpdateResource(gomock.Any(), gomock.Any()).Return(nil)
			ema.EXPECT().GetEventModelByID(gomock.Any()).Return(oldEventModel, nil)
			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.SimpleMetricModel{
				"1": {
					ModelName: "1",
					GroupName: "1",
				},
			}, nil)
			ema.EXPECT().GetEventTaskByModelID(gomock.Any(), gomock.Any()).Return(eventTask, true, nil)

			smock.ExpectBegin()
			ema.EXPECT().UpdateEventModel(gomock.Any(), gomock.Any()).Return(nil)
			ema.EXPECT().GetEventTaskByModelID(gomock.Any(), gomock.Any()).Return(eventTask, true, nil)
			ema.EXPECT().UpdateEventTask(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()

			httpErr := ems.UpdateEventModel(testCtx, newEventModel)
			So(httpErr, ShouldBeNil)
		})

		Convey("UpdateEventModel succeed, and create task", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			ema.EXPECT().GetEventModelByID(gomock.Any()).Return(oldEventModel, nil)
			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.SimpleMetricModel{
				"1": {
					ModelName: "1",
					GroupName: "1",
				},
			}, nil)
			ema.EXPECT().GetEventTaskByModelID(gomock.Any(), gomock.Any()).Return(eventTask, true, nil)

			smock.ExpectBegin()
			ema.EXPECT().UpdateEventModel(gomock.Any(), gomock.Any()).Return(nil)
			ema.EXPECT().GetEventTaskByModelID(gomock.Any(), gomock.Any()).Return(interfaces.EventTask{}, false, nil)

			ema.EXPECT().CreateEventTask(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()

			httpErr := ems.UpdateEventModel(testCtx, newEventModel)
			So(httpErr, ShouldBeNil)
		})

		// Convey("UpdateEventModel succeed, and delete task", func() {
		// 	ema.EXPECT().GetEventModelByID(gomock.Any()).Return(oldEventModel, nil)
		// 	mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.SimpleMetricModel{
		// 		1: {
		// 			ModelName: "1",
		// 			GroupName: "1",
		// 		},
		// 	}, nil)
		// 	ema.EXPECT().GetEventTaskByModelID(gomock.Any(), gomock.Any()).Return(eventTask, true, nil)

		// 	smock.ExpectBegin()
		// 	ema.EXPECT().UpdateEventModel(gomock.Any(), gomock.Any()).Return(nil)
		// 	ema.EXPECT().GetEventTaskByModelID(gomock.Any(), gomock.Any()).Return(eventTask, true, nil)

		// 	ema.EXPECT().SetTaskSyncStatusByTaskID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		// 	smock.ExpectCommit()

		// 	request := interfaces.EventModelUpateRequest{
		// 		EventModelRequest: interfaces.EventModelRequest{
		// 			EventModelName: "测试中的名称_修改",
		// 			EventModelType: "atomic",
		// 			// UpdateTime:        "2022-12-13 01:01:01",
		// 			EventModelTags: []string{"xx1", "xx2", "xx3"},
		// 			DataSourceType: "metric_model",
		// 			DataSource:     []string{"2"},
		// 			DetectRule:     detectRule,
		// 			AggregateRule:  interfaces.AggregateRule{},
		// 			DefaultTimeWindow: interfaces.TimeInterval{
		// 				Interval: 5,
		// 				Unit:     "m",
		// 			},
		// 			EventModelComment: "modify_comment",
		// 			EventTaskRequest:  interfaces.EventTaskRequest{},
		// 		},
		// 		EventModelID: "1",
		// 		IsActive:     1,
		// 	}
		// 	httpErr := ems.UpdateEventModel(testCtx, request)
		// 	So(httpErr, ShouldBeNil)
		// })
		// Convey("UpdateEventModel succeed, and stop job", func() {
		// 	ema.EXPECT().GetEventModelByID(gomock.Any()).Return(oldEventModel, nil)
		// 	mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.SimpleMetricModel{
		// 		"1": {
		// 			ModelName: "1",
		// 			GroupName: "1",
		// 		},
		// 	}, nil)
		// 	ema.EXPECT().GetEventTaskByModelID(gomock.Any(), gomock.Any()).AnyTimes().Return(eventTask, true, nil)

		// 	patch := ApplyFuncReturn(metric_model.StopJob, nil)
		// 	defer patch.Reset()

		// 	smock.ExpectBegin()
		// 	ema.EXPECT().UpdateEventModel(gomock.Any(), gomock.Any()).Return(nil)
		// 	smock.ExpectCommit()

		// 	request := interfaces.EventModelUpateRequest{
		// 		EventModelRequest: interfaces.EventModelRequest{
		// 			EventModelName: "测试中的名称_修改",
		// 			EventModelType: "atomic",
		// 			// UpdateTime:        "2022-12-13 01:01:01",
		// 			EventModelTags: []string{"xx1", "xx2", "xx3"},
		// 			DataSourceType: "metric_model",
		// 			DataSource:     []string{"2"},
		// 			DetectRule:     detectRule,
		// 			AggregateRule:  interfaces.AggregateRule{},
		// 			DefaultTimeWindow: interfaces.TimeInterval{
		// 				Interval: 5,
		// 				Unit:     "m",
		// 			},
		// 			EventModelComment: "modify_comment",
		// 			EventTaskRequest:  interfaces.EventTaskRequest{},
		// 		},
		// 		EventModelID: "1",
		// 		IsActive:     0,
		// 	}
		// 	httpErr := ems.UpdateEventModel(testCtx, request)
		// 	So(httpErr, ShouldBeNil)
		// })
	})
}

func Test_EventModelService_GetEventModelByID(t *testing.T) {
	type fields struct {
		appSetting *common.AppSetting
		dmja       *dmock.MockDataModelJobAccess
		ema        *dmock.MockEventModelAccess
		iba        *dmock.MockIndexBaseAccess
		mms        *dmock.MockMetricModelService
		dvs        *dmock.MockDataViewService
		ems        *dmock.MockEventModelService
		ps         *dmock.MockPermissionService
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	field := fields{
		appSetting: &common.AppSetting{},
		dmja:       dmock.NewMockDataModelJobAccess(mockCtrl),
		ema:        dmock.NewMockEventModelAccess(mockCtrl),
		mms:        dmock.NewMockMetricModelService(mockCtrl),
		ems:        dmock.NewMockEventModelService(mockCtrl),
	}

	type args struct {
		ctx     context.Context
		modelID string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{"event model not found", field, args{context.TODO(), "1"}},
		//{"success", field, args{context.TODO(), string(2)}},
		//{"intererror", field, args{context.TODO(), string(3)}},
		//{"event task not exist", field, args{context.TODO(), string(4)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ems, _ := MockNewEventModelService(tt.fields.appSetting, tt.fields.dmja, tt.fields.ema, tt.fields.mms,
				tt.fields.iba, tt.fields.dvs, tt.fields.ps)
			field.mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, nil)
			field.ema.EXPECT().GetEventModelByID(gomock.Any()).AnyTimes().DoAndReturn(func(any) (interfaces.EventModel, error) {
				if tt.name == "success" {
					return oldEventModel, nil
				} else if tt.name == "intererror" {
					return oldEventModel, errors.New(derrors.EventModel_InternalError)
				} else if tt.name == "event model not found" {
					return interfaces.EventModel{}, errors.New(derrors.EventModel_EventModelNotFound)
				} else {
					return oldEventModel, nil
				}
			})
			field.ema.EXPECT().GetEventTaskByModelID(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(any, any) (interfaces.EventTask, bool, error) {
				if tt.name == "success" {
					return eventTask, true, nil
				} else if tt.name == "intererror" {
					return eventTask, false, errors.New(derrors.EventModel_InternalError)
				} else if tt.name == "event model not found" {
					return eventTask, true, nil
				} else {
					return eventTask, false, nil
				}
			})
			_, httpErr := ems.GetEventModelByID(tt.args.ctx, tt.args.modelID)
			switch tt.name {
			case "success":
				assert.Equal(t, httpErr, (*rest.HTTPError)(nil))
			case "not found":
				assert.Equal(t, httpErr.BaseError.ErrorCode, derrors.EventModel_EventModelNotFound)
			case "intererror":
				assert.Equal(t, httpErr.BaseError.ErrorCode, derrors.EventModel_InternalError)
			case "event task not exist":
				assert.Equal(t, httpErr, (*rest.HTTPError)(nil))
			}
		})
	}
}

func Test_EventModelService_DeleteEventModels(t *testing.T) {
	Convey("Test DeleteEventModels", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}

		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
		ema := dmock.NewMockEventModelAccess(mockCtrl)
		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		ems, smock := MockNewEventModelService(appSetting, dmja, ema, mms, iba, dvs, ps)
		resrc := map[string]interfaces.ResourceOps{
			"1": {
				ResourceID: "1",
			},
		}
		Convey("DeleteEventModels failed,caused by GetEventModelByID failed ", func() {
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			ema.EXPECT().GetEventTaskIDByModelIDs(gomock.Any(), gomock.Any()).Return(nil, nil)
			ema.EXPECT().GetEventModelByID(gomock.Any()).Return(oldEventModel, errors.New("error"))
			_, err := ems.DeleteEventModels(testCtx, []string{"1"})
			httpErr := err.(*rest.HTTPError)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, derrors.EventModel_InternalError)
		})
		Convey("DeleteEventModels failed,caused by GetEventTaskByModelID failed ", func() {
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ema.EXPECT().GetEventTaskIDByModelIDs(gomock.Any(), gomock.Any()).Return(nil, nil)
			ema.EXPECT().GetEventModelByID(gomock.Any()).Return(oldEventModel, nil)
			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.SimpleMetricModel{
				"1": {
					ModelName: "1",
					GroupName: "1",
				},
			}, nil)
			ema.EXPECT().GetEventTaskByModelID(gomock.Any(), gomock.Any()).Return(interfaces.EventTask{}, false, errors.New("error"))
			_, err := ems.DeleteEventModels(testCtx, []string{"1"})
			httpErr := err.(*rest.HTTPError)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, derrors.EventModel_InternalError)
		})
		Convey("DeleteEventModels failed,caused by GetEventModelRefsByID failed ", func() {
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ema.EXPECT().GetEventTaskIDByModelIDs(gomock.Any(), gomock.Any()).Return(nil, nil)
			ema.EXPECT().GetEventModelByID(gomock.Any()).Return(oldEventModel, nil)
			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.SimpleMetricModel{
				"1": {
					ModelName: "1",
					GroupName: "1",
				},
			}, nil)
			ema.EXPECT().GetEventTaskByModelID(gomock.Any(), gomock.Any()).Return(eventTask, true, nil)
			ema.EXPECT().GetEventModelRefsByID(gomock.Any()).Return(0, errors.New("error"))
			// ema.EXPECT().GetEventModelDependenceByID(gomock.Any()).Return(0, nil)

			_, err := ems.DeleteEventModels(testCtx, []string{"1"})
			httpErr := err.(*rest.HTTPError)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, derrors.EventModel_InternalError)
		})
		Convey("DeleteEventModels failed,caused by transaction begin failed ", func() {
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ema.EXPECT().GetEventTaskIDByModelIDs(gomock.Any(), gomock.Any()).Return(nil, nil)
			ema.EXPECT().GetEventModelByID(gomock.Any()).Return(oldEventModel, nil)
			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.SimpleMetricModel{
				"1": {
					ModelName: "1",
					GroupName: "1",
				},
			}, nil)
			ema.EXPECT().GetEventTaskByModelID(gomock.Any(), gomock.Any()).Return(eventTask, true, nil)
			ema.EXPECT().GetEventModelRefsByID(gomock.Any()).Return(0, nil)
			ema.EXPECT().GetEventModelDependenceByID(gomock.Any()).Return(0, nil)

			smock.ExpectBegin().WillReturnError(errors.New("error"))
			_, err := ems.DeleteEventModels(testCtx, []string{"1"})
			httpErr := err.(*rest.HTTPError)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, derrors.DataModel_EventModel_InternalError_BeginTransactionFailed)
		})
		Convey("DeleteEventModels failed,caused by DeleteEventModels failed ", func() {
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ema.EXPECT().GetEventTaskIDByModelIDs(gomock.Any(), gomock.Any()).Return(nil, nil)
			ema.EXPECT().GetEventModelByID(gomock.Any()).Return(oldEventModel, nil)
			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.SimpleMetricModel{
				"1": {
					ModelName: "1",
					GroupName: "1",
				},
			}, nil)
			ema.EXPECT().GetEventTaskByModelID(gomock.Any(), gomock.Any()).Return(eventTask, true, nil)
			ema.EXPECT().GetEventModelRefsByID(gomock.Any()).Return(0, nil)
			ema.EXPECT().GetEventModelDependenceByID(gomock.Any()).Return(0, nil)

			smock.ExpectBegin()
			ema.EXPECT().DeleteEventModels(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			_, err := ems.DeleteEventModels(testCtx, []string{"1"})
			httpErr := err.(*rest.HTTPError)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, derrors.EventModel_InternalError)
		})

		Convey("DeleteEventModels failed,caused by GetEventTaskIDByModelIDs failed ", func() {
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			ema.EXPECT().GetEventTaskIDByModelIDs(gomock.Any(), gomock.Any()).AnyTimes().Return([]string{"1"}, errors.New("error"))

			_, err := ems.DeleteEventModels(testCtx, []string{"1"})
			httpErr := err.(*rest.HTTPError)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, derrors.EventModel_InternalError)
		})
		Convey("DeleteEventModels failed,caused by DeleteEventTaskByTaskIDs failed ", func() {
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ema.EXPECT().GetEventModelByID(gomock.Any()).Return(oldEventModel, nil)
			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.SimpleMetricModel{
				"1": {
					ModelName: "1",
					GroupName: "1",
				},
			}, nil)
			ema.EXPECT().GetEventTaskByModelID(gomock.Any(), gomock.Any()).Return(eventTask, true, nil)
			ema.EXPECT().GetEventModelRefsByID(gomock.Any()).Return(0, nil)
			ema.EXPECT().GetEventModelDependenceByID(gomock.Any()).Return(0, nil)

			smock.ExpectBegin()
			ema.EXPECT().DeleteEventModels(gomock.Any(), gomock.Any()).Return(nil)
			ema.EXPECT().GetEventTaskIDByModelIDs(gomock.Any(), gomock.Any()).AnyTimes().Return([]string{"1"}, nil)
			ema.EXPECT().DeleteEventTaskByTaskIDs(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(errors.New("error"))
			dmja.EXPECT().StopJobs(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			_, err := ems.DeleteEventModels(testCtx, []string{"1"})
			httpErr := err.(*rest.HTTPError)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, derrors.EventModel_InternalError)
		})
		Convey("DeleteEventModels succeed", func() {
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ps.EXPECT().DeleteResources(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ema.EXPECT().GetEventModelByID(gomock.Any()).Return(oldEventModel, nil)
			mms.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.SimpleMetricModel{
				"1": {
					ModelName: "1",
					GroupName: "1",
				},
			}, nil)
			ema.EXPECT().GetEventTaskByModelID(gomock.Any(), gomock.Any()).Return(eventTask, true, nil)
			ema.EXPECT().GetEventModelRefsByID(gomock.Any()).Return(0, nil)
			ema.EXPECT().GetEventModelDependenceByID(gomock.Any()).Return(0, nil)

			smock.ExpectBegin()
			ema.EXPECT().DeleteEventModels(gomock.Any(), gomock.Any()).Return(nil)
			ema.EXPECT().GetEventTaskIDByModelIDs(gomock.Any(), gomock.Any()).AnyTimes().Return([]string{"1"}, nil)
			ema.EXPECT().DeleteEventTaskByTaskIDs(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			dmja.EXPECT().StopJobs(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			smock.ExpectCommit()

			_, httpErr := ems.DeleteEventModels(testCtx, []string{"1"})
			So(httpErr, ShouldBeNil)
		})

	})
}

func Test_EventModelService_QueryEventModels(t *testing.T) {
	type fields struct {
		appSetting *common.AppSetting
		dmja       *dmock.MockDataModelJobAccess
		ema        *dmock.MockEventModelAccess
		iba        *dmock.MockIndexBaseAccess
		mms        *dmock.MockMetricModelService
		dvs        *dmock.MockDataViewService
		ems        *dmock.MockEventModelService
		ps         *dmock.MockPermissionService
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	field := fields{
		appSetting: &common.AppSetting{},
		dmja:       dmock.NewMockDataModelJobAccess(mockCtrl),
		ema:        dmock.NewMockEventModelAccess(mockCtrl),
		mms:        dmock.NewMockMetricModelService(mockCtrl),
		ems:        dmock.NewMockEventModelService(mockCtrl),
		ps:         dmock.NewMockPermissionService(mockCtrl),
	}

	type args struct {
		ctx    context.Context
		params interfaces.EventModelQueryRequest
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{"not found", field, args{context.TODO(), eventModelQueryReq}},
		{"success", field, args{context.TODO(), eventModelQueryReq}},
		//{"intererror", field, args{context.TODO(), eventModelQueryReq}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ems, _ := MockNewEventModelService(tt.fields.appSetting, tt.fields.dmja, tt.fields.ema, tt.fields.mms,
				tt.fields.iba, tt.fields.dvs, tt.fields.ps)
			field.ema.EXPECT().QueryTotalNumberEventModels(gomock.Any()).AnyTimes().DoAndReturn(func(any) (int, error) {
				if tt.name == "success" {
					return 2, nil
				} else if tt.name == "not found" {
					return 0, nil
				} else {
					return 0, errors.New(derrors.EventModel_InternalError)
				}
			})
			field.ema.EXPECT().QueryEventModels(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(any, any) ([]interfaces.EventModel, error) {
				if tt.name == "success" {
					return []interfaces.EventModel{oldEventModel}, nil
				} else if tt.name == "not found" {
					return nil, nil
				} else {
					return nil, errors.New(derrors.EventModel_InternalError)
				}
			})
			field.ema.EXPECT().GetEventTaskByModelID(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(any, any) (interfaces.EventTask, bool, error) {
				if tt.name == "success" {
					return eventTask, true, nil
				} else if tt.name == "not found" {
					return eventTask, false, nil
				} else {
					return eventTask, true, nil
				}
			})
			field.ems.EXPECT().GetEventTaskByModelID(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(any, any) (interfaces.EventTask, bool, error) {
				if tt.name == "success" {
					return eventTask, true, nil
				} else if tt.name == "not found" {
					return eventTask, false, nil
				} else {
					return eventTask, true, nil
				}
			})

			_, total, err := ems.QueryEventModels(tt.args.ctx, tt.args.params)

			switch tt.name {
			case "success":
				assert.Equal(t, err, nil)
			case "not found":
				assert.Equal(t, total, 0)
			case "intererror":
				httpErr := err.(*rest.HTTPError)
				assert.Equal(t, httpErr.BaseError.ErrorCode, derrors.EventModel_InternalError)
			}

		})
	}
}

func Test_EventModelService_GetEventModelMapByNames(t *testing.T) {
	Convey("Test GetEventModelMapByNames", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
		ema := dmock.NewMockEventModelAccess(mockCtrl)
		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		ems, _ := MockNewEventModelService(appSetting, dmja, ema, mms, iba, dvs, ps)

		Convey("Get succeed", func() {
			expectedModelMap := make(map[string]string)
			ema.EXPECT().GetEventModelMapByNames(gomock.Any()).Return(expectedModelMap, nil)

			modelNames := []string{"test"}
			modelMap, err := ems.GetEventModelMapByNames(modelNames)
			So(modelMap, ShouldResemble, expectedModelMap)
			So(err, ShouldBeNil)
		})
	})
}

func Test_EventModelService_GetEventModelMapByIDs(t *testing.T) {
	Convey("Test GetEventModelMapByIDs", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		appSetting := &common.AppSetting{}
		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
		ema := dmock.NewMockEventModelAccess(mockCtrl)
		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		ems, _ := MockNewEventModelService(appSetting, dmja, ema, mms, iba, dvs, ps)

		Convey("Get succeed", func() {
			expectedModelMap := make(map[string]string)
			ema.EXPECT().GetEventModelMapByIDs(gomock.Any()).Return(expectedModelMap, nil)

			modelIDs := []string{"1"}
			modelMap, err := ems.GetEventModelMapByIDs(modelIDs)
			So(modelMap, ShouldResemble, expectedModelMap)
			So(err, ShouldBeNil)
		})
	})
}

func Test_EventModelService_CreateEventTask(t *testing.T) {
	Convey("Test CreateEventTask", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
		ema := dmock.NewMockEventModelAccess(mockCtrl)
		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		ems, _ := MockNewEventModelService(appSetting, dmja, ema, mms, iba, dvs, ps)

		Convey("CreateEventTask success", func() {
			ema.EXPECT().CreateEventTask(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			db, _, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			tx, _ := db.Begin()
			httpErr := ems.CreateEventTask(testCtx, tx, eventTask)
			So(httpErr, ShouldBeNil)
		})
	})
}

func Test_EventModelService_GetEventTaskIDByModelID(t *testing.T) {
	Convey("Test GetEventTaskIDByModelID", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
		ema := dmock.NewMockEventModelAccess(mockCtrl)
		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		ems, _ := MockNewEventModelService(appSetting, dmja, ema, mms, iba, dvs, ps)

		Convey("GetEventTaskIDByModelID success", func() {

			ema.EXPECT().GetEventTaskIDByModelIDs(gomock.Any(), gomock.Any()).Return([]string{"2"}, nil)

			taskID, err := ems.GetEventTaskIDByModelIDs(testCtx, []string{"1"})
			So(taskID, ShouldEqual, []string{"2"})
			So(err, ShouldBeNil)
		})
	})
}

func Test_EventModelService_UpdateEventTask(t *testing.T) {
	Convey("Test UpdateEventTask", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
		ema := dmock.NewMockEventModelAccess(mockCtrl)
		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		ems, _ := MockNewEventModelService(appSetting, dmja, ema, mms, iba, dvs, ps)

		Convey("UpdateEventTask success", func() {
			ema.EXPECT().UpdateEventTask(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			db, _, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			tx, _ := db.Begin()
			httpErr := ems.UpdateEventTask(testCtx, tx, eventTask)
			So(httpErr, ShouldBeNil)
		})
	})
}

func Test_EventModelService_GetEventTaskByTaskID(t *testing.T) {
	Convey("Test GetEventTaskByTaskID", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
		ema := dmock.NewMockEventModelAccess(mockCtrl)
		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		ems, _ := MockNewEventModelService(appSetting, dmja, ema, mms, iba, dvs, ps)

		Convey("GetEventTaskByTaskID success", func() {
			ema.EXPECT().GetEventTaskByTaskID(gomock.Any(), gomock.Any()).Return(eventTask, nil)

			task, httpErr := ems.GetEventTaskByTaskID(testCtx, "1")
			So(httpErr, ShouldBeNil)
			So(task, ShouldResemble, eventTask)
		})
	})
}

func Test_EventModelService_GetEventTaskByModelID(t *testing.T) {
	Convey("Test GetEventTaskByModelID", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
		ema := dmock.NewMockEventModelAccess(mockCtrl)
		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		ems, _ := MockNewEventModelService(appSetting, dmja, ema, mms, iba, dvs, ps)

		Convey("GetEventTaskByModelID success", func() {
			ema.EXPECT().GetEventTaskByModelID(gomock.Any(), gomock.Any()).Return(eventTask, true, nil)

			task, exist, httpErr := ems.GetEventTaskByModelID(testCtx, "1")
			So(httpErr, ShouldBeNil)
			So(exist, ShouldEqual, true)
			So(task, ShouldResemble, eventTask)
		})
	})
}

func Test_EventModelService_DeleteEventTaskByTaskIDs(t *testing.T) {
	Convey("Test DeleteEventTaskByTaskIDs", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
		ema := dmock.NewMockEventModelAccess(mockCtrl)
		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		ems, _ := MockNewEventModelService(appSetting, dmja, ema, mms, iba, dvs, ps)

		Convey("DeleteEventTaskByTaskID success", func() {
			ema.EXPECT().DeleteEventTaskByTaskIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			db, _, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			tx, _ := db.Begin()
			httpErr := ems.DeleteEventTaskByTaskIDs(testCtx, tx, []string{"1"})
			So(httpErr, ShouldBeNil)
		})
	})
}

func Test_EventModelService_UpdateEventTaskStatusInFinish(t *testing.T) {
	Convey("Test UpdateEventTaskStatusInFinish", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
		ema := dmock.NewMockEventModelAccess(mockCtrl)
		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		ems, _ := MockNewEventModelService(appSetting, dmja, ema, mms, iba, dvs, ps)

		Convey("UpdateEventTaskStatusInFinish success", func() {
			ema.EXPECT().UpdateEventTaskStatusInFinish(gomock.Any(), gomock.Any()).Return(nil)
			httpErr := ems.UpdateEventTaskStatusInFinish(testCtx, eventTask)
			So(httpErr, ShouldBeNil)
		})
	})
}

func Test_EventModelService_InitDispatchConfig(t *testing.T) {
	Convey("Test InitDispatchConfig", t, func() {

		eventTask1 := interfaces.EventTask{
			TaskID:   "1",
			ModelID:  "1",
			Schedule: interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
			StorageConfig: interfaces.StorageConfig{
				IndexBase:    "base1",
				DataViewName: "view1",
			},
			DispatchConfig: interfaces.DispatchConfig{
				TimeOut:        0,
				RouteStrategy:  "",
				BlockStrategy:  "",
				FailRetryCount: 0,
			},
			ExecuteParameter:   map[string]any{},
			TaskStatus:         4,
			StatusUpdateTime:   testUpdateTime,
			ErrorDetails:       "",
			ScheduleSyncStatus: 3,
			UpdateTime:         testUpdateTime,
		}
		Convey("InitDispatchConfig success", func() {
			task := InitDispatchConfig(eventTask1)
			So(task, ShouldResemble, eventTask)
		})
	})
}
