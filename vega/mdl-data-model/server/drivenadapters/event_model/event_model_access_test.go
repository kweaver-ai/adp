// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package event_model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	sq "github.com/Masterminds/squirrel"
	. "github.com/agiledragon/gomonkey/v2"
	"github.com/bytedance/sonic"
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

	detectRule = interfaces.DetectRule{
		DetectRuleID: "1",
		Priority:     99,
		Type:         "range_detect",
		Formula: []interfaces.FormulaItem{
			{
				Level: 1,
				Filter: interfaces.LogicFilter{
					LogicOperator: "",
					FilterExpress: interfaces.FilterExpress{Name: "cpu利用率", Value: []interface{}{0.9, 0.97}, Operation: "range"},
					Children:      []interfaces.LogicFilter{},
				},
			},
			{
				Level: 2,
				Filter: interfaces.LogicFilter{
					LogicOperator: "",
					FilterExpress: interfaces.FilterExpress{Name: "cpu利用率", Value: []interface{}{0.85, 0.9}, Operation: "range"},
					Children:      []interfaces.LogicFilter{},
				},
			},
		},
		DetectAlgo:   "",
		AnalysisAlgo: map[string]string{},
		CreateTime:   testUpdateTime,
		UpdateTime:   testUpdateTime,
	}

	aggregateRule = interfaces.AggregateRule{
		AggregateRuleID: "1",
		Priority:        99,
		Type:            "healthy_compute",
		AggregateAlgo:   "MaxLevelMap",
		AnalysisAlgo:    map[string]string{},
		GroupFields:     []string{},
		CreateTime:      testUpdateTime,
		UpdateTime:      testUpdateTime,
	}

	eventModelQuery = interfaces.EventModelQueryRequest{
		EventModelName:        "测试中的名称",
		EventModelNamePattern: "测试中的名称",
		EventModelType:        "atomic",
		EventModelTag:         "xx1,xx2",
		SortKey:               "update_time",
		Direction:             "asc",
		Offset:                0,
		Limit:                 10,
		IsActive:              "1",
		IsCustom:              1,
		ScheduleSyncStatus:    "1",
		TaskStatus:            "4",
	}

	oldEventModel = interfaces.EventModel{
		EventModelID:      "1",
		EventModelName:    "测试中的名称",
		EventModelType:    "atomic",
		CreateTime:        testUpdateTime,
		UpdateTime:        testUpdateTime,
		EventModelTags:    []string{"xx1", "xx2"},
		DataSourceType:    "metric_model",
		DataSource:        []string{"1"},
		DetectRule:        detectRule,
		AggregateRule:     aggregateRule,
		DefaultTimeWindow: interfaces.TimeInterval{Interval: 5, Unit: "m"},
		EventModelComment: "comment",
		IsActive:          1,
		IsCustom:          1,
		EnableSubscribe:   0,
		Status:            1,
	}

	oldEventModelForDetectRule = interfaces.EventModel{
		EventModelID:             "1",
		EventModelName:           "测试中的名称",
		EventModelType:           "atomic",
		CreateTime:               testUpdateTime,
		UpdateTime:               testUpdateTime,
		EventModelTags:           []string{"xx1", "xx2"},
		DataSourceType:           "metric_model",
		DataSource:               []string{"1"},
		DetectRule:               detectRule,
		AggregateRule:            interfaces.AggregateRule{},
		DefaultTimeWindow:        interfaces.TimeInterval{Interval: 5, Unit: "m"},
		EventModelComment:        "comment",
		IsActive:                 1,
		IsCustom:                 1,
		EnableSubscribe:          0,
		Status:                   1,
		DownstreamDependentModel: []string{},
	}

	// oldEventModelForAggregateRule = interfaces.EventModel{
	// 	EventModelID:      "1",
	// 	EventModelName:    "测试中的名称",
	// 	EventModelType:    "aggregate",
	// 	UpdateTime:        testUpdateTime,
	// 	EventModelTags:    []string{"xx1", "xx2"},
	// 	DataSourceType:    "metric_model",
	// 	DataSource:        []string{"1"},
	// 	DetectRule:        interfaces.DetectRule{},
	// 	AggregateRule:     aggregateRule,
	// 	DefaultTimeWindow: interfaces.TimeInterval{Interval: 5, Unit: "m"},
	// 	EventModelComment: "comment",
	// 	IsActive:          1,
	// 	IsCustom:          1,
	// }

	eventModelRecord = interfaces.EventModel{
		EventModelID:             "1",
		EventModelName:           "测试中的名称",
		EventModelType:           "atomic",
		CreateTime:               testUpdateTime,
		UpdateTime:               testUpdateTime,
		EventModelTags:           []string{"xx1", "xx2"},
		DataSourceType:           "metric_model",
		DataSource:               []string{"1"},
		DataSourceName:           []string{},
		DataSourceGroupName:      []string{},
		DetectRule:               interfaces.DetectRule{DetectRuleID: "", Priority: 0, Type: "", Formula: []interfaces.FormulaItem{}, AnalysisAlgo: map[string]string{}},
		AggregateRule:            interfaces.AggregateRule{GroupFields: []string{}, AnalysisAlgo: map[string]string{}},
		DefaultTimeWindow:        interfaces.TimeInterval{Interval: 5, Unit: "m"},
		EventModelComment:        "comment",
		IsActive:                 1,
		IsCustom:                 1,
		EnableSubscribe:          0,
		Status:                   1,
		DownstreamDependentModel: []string{},
		Task:                     eventTask,
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
			RouteStrategy:  "FIRST",
			BlockStrategy:  "",
			FailRetryCount: 3,
		},
		ExecuteParameter:        map[string]any{},
		TaskStatus:              4,
		StatusUpdateTime:        testUpdateTime,
		ErrorDetails:            "",
		ScheduleSyncStatus:      1,
		CreateTime:              testUpdateTime,
		UpdateTime:              testUpdateTime,
		DownstreamDependentTask: []string{},
	}
	eventScheduleBytes, _    = sonic.Marshal(eventTask.Schedule)
	storageConfigBytes, _    = sonic.Marshal(eventTask.StorageConfig)
	dispatchConfigBytes, _   = sonic.Marshal(eventTask.DispatchConfig)
	executeParameterBytes, _ = sonic.Marshal(eventTask.ExecuteParameter)
	downstreamDependentTask  = strings.Join(eventTask.DownstreamDependentTask, ",")
)

func MockNewEventModelAccess(appSetting *common.AppSetting,
	mma interfaces.MetricModelAccess) (*eventModelAccess, sqlmock.Sqlmock) {
	db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	ema := &eventModelAccess{
		appSetting: appSetting,
		db:         db,
		mma:        mma,
	}

	return ema, smock
}

func Test_EventModelAccess_GetEventModelByID(t *testing.T) {
	Convey("test GetEventModelByID\n", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mma := dmock.NewMockMetricModelAccess(mockCtrl)
		ema, smock := MockNewEventModelAccess(appSetting, mma)

		event_rows_with_detect := sqlmock.NewRows([]string{
			"f_event_model_id", "f_event_model_name", "f_event_model_type", "f_event_model_tags",
			"f_event_model_comment", "f_data_source_type", "f_data_source", "f_detect_rule_id",
			"f_aggregate_rule_id", "COALESCE(f_downstream_dependent_model,'')", "f_default_time_window",
			"f_is_active", "f_is_custom", "f_enable_subscribe", "f_status", "f_create_time", "f_update_time",
		}).AddRow(
			"1", "测试中的名称", "atomic", "xx1,xx2",
			"comment", "metric_model", "[\"1\"]", "1",
			"", "", "{\"interval\":5,\"unit\":\"m\"}",
			1, 1, 0, 1, testUpdateTime, testUpdateTime)

		rule_rows := sqlmock.NewRows([]string{
			"f_detect_rule_id", "f_detect_rule_type", "f_formula", "f_detect_algo",
			"f_detect_analysis_algo", "f_rule_priority", "f_create_time", "f_update_time",
		}).AddRow(
			"1", "range_detect", `[{"level":1,"filter":{"logic_operator":"","filter_express":{"name":"cpu利用率","value":[0.9,0.97],"operation":"range"},"children":[]}},
		{"level":2,"filter":{"logic_operator":"","filter_express":{"name":"cpu利用率","value":[0.85,0.9],"operation":"range"},"children":[]}}]`, "",
			"{}", 99, testUpdateTime, testUpdateTime)

		model_sql := fmt.Sprintf("SELECT f_event_model_id, f_event_model_name, f_event_model_type, "+
			"f_event_model_tags, f_event_model_comment, f_data_source_type, f_data_source, f_detect_rule_id, "+
			"f_aggregate_rule_id, COALESCE(f_downstream_dependent_model,''), f_default_time_window, "+
			"f_is_active, f_is_custom, f_enable_subscribe, f_status, f_create_time, f_update_time "+
			"FROM %s WHERE f_event_model_id = ?", EVENT_MODEL_TABLE_NAME)
		//NOTE: 注意空格不对也会报错
		rule_sql := fmt.Sprintf("SELECT f_detect_rule_id, f_detect_rule_type, f_formula, "+
			"COALESCE(f_detect_algo,''), COALESCE(f_detect_analysis_algo,'{}'), f_rule_priority, "+
			"f_create_time, f_update_time FROM %s WHERE f_detect_rule_id = ?", DETECT_RULE_TABLE_NAME)

		modelID := "1"

		Convey("GetEventModelByID Success \n", func() {
			smock.ExpectQuery(model_sql).WithArgs().WillReturnRows(event_rows_with_detect)
			smock.ExpectQuery(rule_sql).WithArgs().WillReturnRows(rule_rows)

			em, err := ema.GetEventModelByID(modelID)
			fmt.Printf("%v", em)
			So(em, ShouldResemble, oldEventModelForDetectRule)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetEventModelByID Failed  not found \n", func() {
			smock.ExpectQuery(model_sql).WithArgs().WillReturnError(sql.ErrNoRows)

			em, err := ema.GetEventModelByID(modelID)
			fmt.Printf("%v", em)
			So(err, ShouldBeError)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetEventModelByID Failed inter error \n", func() {
			smock.ExpectQuery(model_sql).WithArgs().WillReturnError(errors.New(derrors.EventModel_InternalError))

			em, err := ema.GetEventModelByID(modelID)
			fmt.Printf("%v", em)
			So(err.Error(), ShouldEqual, derrors.EventModel_InternalError)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetEventModelByID Failed  inter error with rule find error \n", func() {
			smock.ExpectQuery(model_sql).WithArgs().WillReturnRows(event_rows_with_detect)
			smock.ExpectQuery(rule_sql).WithArgs().WillReturnError(errors.New(derrors.EventModel_InternalError))

			em, err := ema.GetEventModelByID(modelID)
			fmt.Printf("%v", em)
			So(err.Error(), ShouldEqual, derrors.EventModel_InternalError)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetEventModelByID Failed  inter error with rule not found \n", func() {
			smock.ExpectQuery(model_sql).WithArgs().WillReturnRows(event_rows_with_detect)
			smock.ExpectQuery(rule_sql).WithArgs().WillReturnError(sql.ErrNoRows)

			em, err := ema.GetEventModelByID(modelID)
			fmt.Printf("%v", em)
			So(err, ShouldBeError)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_EventModelAccess_CreateEventModel(t *testing.T) {
	Convey("test CreateModel\n", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mma := dmock.NewMockMetricModelAccess(mockCtrl)
		ema, smock := MockNewEventModelAccess(appSetting, mma)

		model_sql := fmt.Sprintf("INSERT INTO %s (f_event_model_id,f_event_model_name,"+
			"f_event_model_group_name,f_event_model_type,f_event_model_comment,f_event_model_tags,"+
			"f_data_source_type,f_data_source,f_detect_rule_id,f_aggregate_rule_id,f_default_time_window,"+
			"f_is_active,f_is_custom,f_enable_subscribe,f_status,f_downstream_dependent_model,"+
			"f_creator,f_creator_type,f_create_time,f_update_time) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", EVENT_MODEL_TABLE_NAME)
		rule_sql := fmt.Sprintf("INSERT INTO %s (f_detect_rule_id,f_detect_rule_type,f_rule_priority,"+
			"f_formula,f_detect_algo,f_detect_analysis_algo,f_create_time,f_update_time) "+
			"VALUES (?,?,?,?,?,?,?,?)", DETECT_RULE_TABLE_NAME)

		Convey("CreateModel Success \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(rule_sql).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
			smock.ExpectExec(model_sql).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			tx, _ := ema.db.Begin()
			_, err := ema.CreateEventModels(tx, []interfaces.EventModel{oldEventModelForDetectRule})
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CreateModel Failed \n", func() {
			expectedErr := errors.New(derrors.EventModel_InternalError)
			smock.ExpectBegin()
			smock.ExpectExec(rule_sql).WithArgs().WillReturnError(expectedErr)

			tx, _ := ema.db.Begin()
			_, err := ema.CreateEventModels(tx, []interfaces.EventModel{oldEventModelForDetectRule})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CreateModel Failed2 \n", func() {
			expectedErr := errors.New(derrors.EventModel_InternalError)
			smock.ExpectBegin()
			smock.ExpectExec(rule_sql).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
			smock.ExpectExec(model_sql).WithArgs().WillReturnError(expectedErr)

			tx, _ := ema.db.Begin()
			_, err := ema.CreateEventModels(tx, []interfaces.EventModel{oldEventModelForDetectRule})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_EventModelAccess_UpdateEventModel(t *testing.T) {
	Convey("test UpdateModel\n", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mma := dmock.NewMockMetricModelAccess(mockCtrl)
		ema, smock := MockNewEventModelAccess(appSetting, mma)

		model_sql := fmt.Sprintf(`UPDATE %s SET f_event_model_name = ?, f_event_model_tags = ?, f_event_model_comment = ?,
			f_data_source = ?,	f_data_source_type = ?, f_default_time_window = ?, f_is_active = ?, f_enable_subscribe = ?,
			f_status = ?, f_update_time = ?, f_downstream_dependent_model = ? WHERE f_event_model_id = ?`, EVENT_MODEL_TABLE_NAME)

		rule_sql := fmt.Sprintf(`UPDATE %s SET f_detect_rule_type = ?, f_formula = ?, f_detect_algo = ?, 
			f_detect_analysis_algo = ?, f_update_time = ? WHERE f_detect_rule_id = ?`, DETECT_RULE_TABLE_NAME)

		Convey("update Success \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(model_sql).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
			smock.ExpectExec(rule_sql).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			tx, _ := ema.db.Begin()
			err := ema.UpdateEventModel(tx, oldEventModel)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("update failed with model  \n", func() {
			expectedErr := errors.New(derrors.EventModel_InternalError)
			smock.ExpectBegin()
			smock.ExpectExec(model_sql).WithArgs().WillReturnError(expectedErr)

			tx, _ := ema.db.Begin()
			err := ema.UpdateEventModel(tx, oldEventModel)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("update failed with rule\n", func() {
			expectedErr := errors.New(derrors.EventModel_InternalError)
			smock.ExpectBegin()
			smock.ExpectExec(model_sql).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
			smock.ExpectExec(rule_sql).WithArgs().WillReturnError(expectedErr)

			tx, _ := ema.db.Begin()
			err := ema.UpdateEventModel(tx, oldEventModel)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_EventModelAccess_DeleteEventModels(t *testing.T) {
	Convey("test DeleteModel\n", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mma := dmock.NewMockMetricModelAccess(mockCtrl)
		ema, smock := MockNewEventModelAccess(appSetting, mma)

		model_sql := fmt.Sprintf(`DELETE FROM %s WHERE f_event_model_id = ?`, EVENT_MODEL_TABLE_NAME)
		rule_sql := fmt.Sprintf(`DELETE FROM %s WHERE f_detect_rule_id = ?`, DETECT_RULE_TABLE_NAME)
		arule_sql := fmt.Sprintf(`DELETE FROM %s WHERE f_aggregate_rule_id = ?`, AGGREGATE_RULE_TABLE_NAME)

		Convey("delete Success \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(rule_sql).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
			smock.ExpectExec(arule_sql).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
			smock.ExpectExec(model_sql).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			tx, _ := ema.db.Begin()
			err := ema.DeleteEventModels(tx, []interfaces.EventModel{oldEventModel})
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_EventModelAccess_QueryEventModels(t *testing.T) {
	Convey("test QueryModel\n", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mma := dmock.NewMockMetricModelAccess(mockCtrl)
		ema, smock := MockNewEventModelAccess(appSetting, mma)

		event_rows := sqlmock.NewRows([]string{
			"f_event_model_id", "f_event_model_name", "f_event_model_type", "f_event_model_tags",
			"f_event_model_comment", "f_data_source_type", "f_data_source", "f_detect_rule_id",
			"f_aggregate_rule_id", "COALESCE(f_downstream_dependent_model,'')",
			"COALESCE(f_detect_rule_type,'')", "COALESCE(f_detect_algo,'')", "COALESCE(f_formula,'')",
			"COALESCE(f_aggregate_rule_type, '')", "COALESCE(f_aggregate_algo,'')",
			"COALESCE(f_group_fields,'')", "f_default_time_window", "f_is_active", "f_is_custom",
			"f_enable_subscribe", "COALESCE(f_detect_analysis_algo,'{}')",
			"COALESCE(f_aggregate_analysis_algo,'{}')", "f_status", "event_models.f_create_time",
			"event_models.f_update_time", "event_model_detect_rules.f_create_time",
			"event_model_detect_rules.f_update_time", "event_model_aggregate_rules.f_create_time",
			"event_model_aggregate_rules.f_update_time", "f_task_id", "f_model_id", "f_storage_config",
			"f_schedule", "f_dispatch_config", "f_execute_parameter", "f_task_status", "f_error_details",
			"COALESCE(f_downstream_dependent_task,'')", "f_status_update_time", "f_schedule_sync_status",
			"t_event_model_task.f_create_time", "t_event_model_task.f_update_time"},
		).AddRow(
			"1", "测试中的名称", "atomic", "xx1,xx2", "comment", "metric_model", "[\"1\"]", "", "", "", "", "", "",
			"", "", "", "{\"interval\":5,\"unit\":\"m\"}", 1, 1, 0, "{}", "{}", 1, testUpdateTime, testUpdateTime,
			0, 0, 0, 0, eventTask.TaskID, eventTask.ModelID, storageConfigBytes, eventScheduleBytes,
			dispatchConfigBytes, executeParameterBytes, eventTask.TaskStatus, eventTask.ErrorDetails, "",
			eventTask.StatusUpdateTime, eventTask.ScheduleSyncStatus, eventTask.CreateTime, eventTask.UpdateTime)

		sqlStr := fmt.Sprintf("SELECT f_event_model_id, f_event_model_name, f_event_model_type, "+
			"f_event_model_tags, f_event_model_comment, f_data_source_type, f_data_source, "+
			"f_detect_rule_id, f_aggregate_rule_id, COALESCE(f_downstream_dependent_model,''), "+
			"COALESCE(f_detect_rule_type,''), COALESCE(f_detect_algo,''), COALESCE(f_formula,''), "+
			"COALESCE(f_aggregate_rule_type, ''), COALESCE(f_aggregate_algo,''), "+
			"COALESCE(f_group_fields,''), f_default_time_window, f_is_active, f_is_custom, "+
			"f_enable_subscribe, COALESCE(f_detect_analysis_algo,'{}'), "+
			"COALESCE(f_aggregate_analysis_algo,'{}'), f_status, t_event_models.f_create_time, "+
			"t_event_models.f_update_time, COALESCE(t_event_model_detect_rules.f_create_time,0), "+
			"COALESCE(t_event_model_detect_rules.f_update_time,0), "+
			"COALESCE(t_event_model_aggregate_rules.f_create_time,0), "+
			"COALESCE(t_event_model_aggregate_rules.f_update_time,0), COALESCE(f_task_id,0), "+
			"COALESCE(f_model_id,0), COALESCE(f_storage_config,'{}'), COALESCE(f_schedule,'{}'), "+
			"COALESCE(f_dispatch_config,'{}'), COALESCE(f_execute_parameter,'{}'), "+
			"COALESCE(f_task_status,0), COALESCE(f_error_details,''), "+
			"COALESCE(f_downstream_dependent_task,''), COALESCE(f_status_update_time,''), "+
			"COALESCE(f_schedule_sync_status,0), COALESCE(t_event_model_task.f_create_time,0), "+
			"COALESCE(t_event_model_task.f_update_time,0) FROM %s LEFT JOIN %s USING (f_detect_rule_id) "+
			"LEFT JOIN %s USING (f_aggregate_rule_id) LEFT JOIN %s "+
			"ON t_event_models.f_event_model_id = t_event_model_task.f_model_id "+
			"WHERE f_event_model_name LIKE ? AND f_event_model_name = ? AND f_event_model_type IN (?) "+
			"AND instr(f_event_model_tags, ?) > 0 AND f_is_active IN (?) AND f_is_custom = ? "+
			"AND f_task_status IN (?) AND f_schedule_sync_status IN (?) "+
			"ORDER BY t_event_models.f_update_time asc",
			EVENT_MODEL_TABLE_NAME, DETECT_RULE_TABLE_NAME, AGGREGATE_RULE_TABLE_NAME, EVENT_TASK_TABLE_NAME)

		Convey("query Success \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(event_rows)
			mma.EXPECT().GetMetricModelSimpleInfosByIDs(gomock.Any(), gomock.Any()).Return(
				map[string]interfaces.SimpleMetricModel{}, nil)

			ems, err := ema.QueryEventModels(testCtx, eventModelQuery)
			So(ems[0], ShouldResemble, eventModelRecord)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_EventModelAccess_QueryTotalNumberEventModels(t *testing.T) {
	Convey("test QueryModel\n", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mma := dmock.NewMockMetricModelAccess(mockCtrl)
		ema, smock := MockNewEventModelAccess(appSetting, mma)

		total_rows := sqlmock.NewRows([]string{"total"}).AddRow(int(0))

		sqlStr := fmt.Sprintf("SELECT count(f_event_model_id) FROM %s "+
			"LEFT JOIN %s USING (f_detect_rule_id) LEFT JOIN %s USING (f_aggregate_rule_id) "+
			"LEFT JOIN %s ON t_event_models.f_event_model_id = t_event_model_task.f_model_id "+
			"WHERE f_event_model_name LIKE ? AND f_event_model_name = ? "+
			"AND f_event_model_type IN (?) AND instr(f_event_model_tags, ?) > 0 "+
			"AND f_is_active IN (?) AND f_is_custom = ? AND f_task_status IN (?) "+
			"AND f_schedule_sync_status IN (?)", EVENT_MODEL_TABLE_NAME, DETECT_RULE_TABLE_NAME,
			AGGREGATE_RULE_TABLE_NAME, EVENT_TASK_TABLE_NAME)

		Convey("query total Success \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(total_rows)

			total, err := ema.QueryTotalNumberEventModels(eventModelQuery)
			So(total, ShouldEqual, int(0))
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("query total fail \n", func() {
			expectedErr := errors.New(derrors.EventModel_EventModelNotFound)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, err := ema.QueryTotalNumberEventModels(eventModelQuery)
			So(err.Error(), ShouldEqual, derrors.EventModel_EventModelNotFound)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_EventModelAccess_GetEventModelMapByNames(t *testing.T) {
	Convey("Test GetEventModelMapByNames", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mma := dmock.NewMockMetricModelAccess(mockCtrl)
		ema, smock := MockNewEventModelAccess(appSetting, mma)

		rows := sqlmock.NewRows([]string{"f_event_model_name", "f_event_model_id"}).AddRow(
			eventModelRecord.EventModelName, eventModelRecord.EventModelID)

		sqlStr := fmt.Sprintf("SELECT f_event_model_name, f_event_model_id "+
			"FROM %s WHERE f_event_model_name IN (?)", EVENT_MODEL_TABLE_NAME)

		Convey("Get failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			modelNames := []string{eventModelRecord.EventModelName}
			_, err := ema.GetEventModelMapByNames(modelNames)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by query error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			modelNames := []string{eventModelRecord.EventModelName}
			_, err := ema.GetEventModelMapByNames(modelNames)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Check failed, caused by the scan error", func() {
			expectedErr := errors.New("sql: expected 1 destination arguments in Scan, not 2")
			rows := sqlmock.NewRows([]string{"event_model_name"}).
				AddRow(eventModelRecord.EventModelID)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			modelNames := []string{eventModelRecord.EventModelName}
			_, err := ema.GetEventModelMapByNames(modelNames)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get succeed", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			modelNames := []string{eventModelRecord.EventModelName}
			modelMap, err := ema.GetEventModelMapByNames(modelNames)
			So(modelMap, ShouldResemble, map[string]string{
				eventModelRecord.EventModelName: eventModelRecord.EventModelID,
			})
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_EventModelAccess_GetEventModelMapByIDs(t *testing.T) {
	Convey("Test GetEventModelMapByIDs", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mma := dmock.NewMockMetricModelAccess(mockCtrl)
		ema, smock := MockNewEventModelAccess(appSetting, mma)

		rows := sqlmock.NewRows([]string{
			"f_event_model_id",
			"f_event_model_name"},
		).AddRow(
			eventModelRecord.EventModelID,
			eventModelRecord.EventModelName)

		sqlStr := fmt.Sprintf("SELECT f_event_model_id, f_event_model_name FROM %s "+
			"WHERE f_event_model_id IN (?)", EVENT_MODEL_TABLE_NAME)

		Convey("Get failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			modelIDs := []string{eventModelRecord.EventModelID}
			_, err := ema.GetEventModelMapByIDs(modelIDs)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by query error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			modelIDs := []string{eventModelRecord.EventModelID}
			_, err := ema.GetEventModelMapByIDs(modelIDs)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Check failed, caused by the scan error", func() {
			expectedErr := errors.New("sql: expected 1 destination arguments in Scan, not 2")
			rows := sqlmock.NewRows([]string{"f_model_name"}).
				AddRow(eventModelRecord.EventModelID)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			modelIDs := []string{eventModelRecord.EventModelID}
			_, err := ema.GetEventModelMapByIDs(modelIDs)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get succeed", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			modelIDs := []string{eventModelRecord.EventModelID}
			modelMap, err := ema.GetEventModelMapByIDs(modelIDs)
			So(modelMap, ShouldResemble, map[string]string{
				eventModelRecord.EventModelID: eventModelRecord.EventModelName,
			})
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_EventModelAccess_CreateEventTask(t *testing.T) {
	Convey("Test CreateEventTask\n", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mma := dmock.NewMockMetricModelAccess(mockCtrl)
		ema, smock := MockNewEventModelAccess(appSetting, mma)

		sqlStr := fmt.Sprintf("INSERT INTO %s (f_task_id,f_model_id,f_schedule,f_dispatch_config,"+
			"f_execute_parameter,f_storage_config,f_task_status,f_status_update_time,f_error_details,"+
			"f_schedule_sync_status,f_downstream_dependent_task,f_creator,f_creator_type,f_create_time,f_update_time) "+
			"VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", EVENT_TASK_TABLE_NAME)

		Convey("CreateEventTask Success \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			tx, _ := ema.db.Begin()
			err := ema.CreateEventTask(testCtx, tx, eventTask)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CreateEventTask prepare error \n", func() {
			expectedErr := errors.New("any error")
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := ema.db.Begin()
			err := ema.CreateEventTask(testCtx, tx, eventTask)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Create failed, caused by Schedule marshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFuncReturn(sonic.Marshal, []byte{}, expectedErr)
			defer patch.Reset()

			smock.ExpectBegin()

			tx, _ := ema.db.Begin()
			err := ema.CreateEventTask(testCtx, tx, eventTask)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Create failed, caused by storageConfig marshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFunc(sonic.Marshal,
				func(v any) ([]byte, error) {
					if _, ok := v.(interfaces.StorageConfig); ok {
						return []byte{}, expectedErr
					}
					return []byte{}, nil
				},
			)
			defer patch.Reset()

			smock.ExpectBegin()

			tx, _ := ema.db.Begin()
			err := ema.CreateEventTask(testCtx, tx, eventTask)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Create failed, caused by executeParameter marshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFunc(sonic.Marshal,
				func(v any) ([]byte, error) {
					if _, ok := v.(map[string]any); ok {
						return []byte{}, expectedErr
					}
					return []byte{}, nil
				},
			)
			defer patch.Reset()

			smock.ExpectBegin()

			tx, _ := ema.db.Begin()
			err := ema.CreateEventTask(testCtx, tx, eventTask)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Create failed, caused by dispatchConfigBytes marshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFunc(sonic.Marshal,
				func(v any) ([]byte, error) {
					if _, ok := v.(interfaces.DispatchConfig); ok {
						return []byte{}, expectedErr
					}
					return []byte{}, nil
				},
			)
			defer patch.Reset()

			smock.ExpectBegin()

			tx, _ := ema.db.Begin()
			err := ema.CreateEventTask(testCtx, tx, eventTask)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Create failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.InsertBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			smock.ExpectBegin()

			tx, _ := ema.db.Begin()
			err := ema.CreateEventTask(testCtx, tx, eventTask)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CreateEventTasks  Exec sql error\n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := ema.db.Begin()
			err := ema.CreateEventTask(testCtx, tx, eventTask)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_EventModelAccess_GetEventTaskIDByModelIDs(t *testing.T) {
	Convey(" Test GetEventTaskIDByModelIDs\n", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mma := dmock.NewMockMetricModelAccess(mockCtrl)
		ema, smock := MockNewEventModelAccess(appSetting, mma)

		rows := sqlmock.NewRows([]string{"f_task_id"}).AddRow("1")

		sqlStr := fmt.Sprintf("SELECT f_task_id FROM %s WHERE f_model_id IN (?)", EVENT_TASK_TABLE_NAME)

		Convey("GetEventTaskIDByModelIDs Success \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			taskID, err := ema.GetEventTaskIDByModelIDs(testCtx, []string{"1"})
			So(err, ShouldBeNil)
			So(taskID, ShouldEqual, []string{"1"})

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			taskID, err := ema.GetEventTaskIDByModelIDs(testCtx, []string{"1"})
			So(taskID, ShouldBeNil)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by query error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			taskID, err := ema.GetEventTaskIDByModelIDs(testCtx, []string{"1"})
			So(taskID, ShouldBeNil)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Check failed, caused by the scan error", func() {
			expectedErr := errors.New("sql: expected 2 destination arguments in Scan, not 1")
			rows := sqlmock.NewRows([]string{"f_task_id", "f_task_id"}).
				AddRow("1", string("2"))
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			taskID, err := ema.GetEventTaskIDByModelIDs(testCtx, []string{"1"})
			So(taskID, ShouldEqual, []string{})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_EventModelAccess_UpdateEventTask(t *testing.T) {
	Convey("Test UpdateEventTask\n", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mma := dmock.NewMockMetricModelAccess(mockCtrl)
		ema, smock := MockNewEventModelAccess(appSetting, mma)

		sqlStr := fmt.Sprintf("UPDATE %s SET f_dispatch_config = ?, f_downstream_dependent_task = ?, "+
			"f_execute_parameter = ?, f_schedule = ?, f_schedule_sync_status = ?, f_storage_config = ?, "+
			"f_update_time = ? WHERE f_task_id = ? ", EVENT_TASK_TABLE_NAME)

		Convey("UpdateEventTask Success \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			tx, _ := ema.db.Begin()
			err := ema.UpdateEventTask(testCtx, tx, eventTask)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateEventTask failed, caused by Schedule marshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFuncReturn(sonic.Marshal, []byte{}, expectedErr)
			defer patch.Reset()

			smock.ExpectBegin()

			tx, _ := ema.db.Begin()
			err := ema.UpdateEventTask(testCtx, tx, eventTask)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateEventTask failed, caused by storageConfig marshal error", func() {
			expectedErr := errors.New("some error")

			patch := ApplyFunc(sonic.Marshal,
				func(v any) ([]byte, error) {
					if _, ok := v.(interfaces.StorageConfig); ok {
						return []byte{}, expectedErr
					}
					return []byte{}, nil
				},
			)
			defer patch.Reset()

			smock.ExpectBegin()

			tx, _ := ema.db.Begin()
			err := ema.UpdateEventTask(testCtx, tx, eventTask)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateEventTask failed, caused by dispatchConfig marshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFunc(sonic.Marshal,
				func(v any) ([]byte, error) {
					if _, ok := v.(interfaces.DispatchConfig); ok {
						return []byte{}, expectedErr
					}
					return []byte{}, nil
				},
			)
			defer patch.Reset()

			smock.ExpectBegin()

			tx, _ := ema.db.Begin()
			err := ema.UpdateEventTask(testCtx, tx, eventTask)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateEventTask Exec sql error\n", func() {
			expectedErr := errors.New("some error1")
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := ema.db.Begin()
			err := ema.UpdateEventTask(testCtx, tx, eventTask)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateEventTask failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.UpdateBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			smock.ExpectBegin()

			tx, _ := ema.db.Begin()
			err := ema.UpdateEventTask(testCtx, tx, eventTask)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_EventModelAccess_GetEventTaskByTaskID(t *testing.T) {
	Convey("Test GetEventTaskByTaskID\n", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mma := dmock.NewMockMetricModelAccess(mockCtrl)
		ema, smock := MockNewEventModelAccess(appSetting, mma)

		rows := sqlmock.NewRows([]string{
			"f_task_id",
			"f_model_id",
			"f_schedule",
			"f_dispatch_config",
			"f_execute_parameter",
			"f_storage_config",
			"f_task_status",
			"f_status_update_time",
			"f_error_details",
			"f_schedule_sync_status",
			"f_downstream_dependent_task",
			"f_creator",
			"f_creator_type",
			"f_create_time",
			"f_update_time",
		}).AddRow(
			eventTask.TaskID,
			eventTask.ModelID,
			eventScheduleBytes,
			dispatchConfigBytes,
			executeParameterBytes,
			storageConfigBytes,
			eventTask.TaskStatus,
			eventTask.StatusUpdateTime,
			eventTask.ErrorDetails,
			eventTask.ScheduleSyncStatus,
			downstreamDependentTask,
			eventTask.Creator.ID,
			eventTask.Creator.Type,
			eventTask.CreateTime,
			eventTask.UpdateTime,
		)

		sqlStr := fmt.Sprintf("SELECT f_task_id, f_model_id, f_schedule, f_dispatch_config, "+
			"f_execute_parameter, f_storage_config,	f_task_status, f_status_update_time, f_error_details, "+
			"f_schedule_sync_status, f_downstream_dependent_task, f_creator, f_creator_type, f_create_time, f_update_time "+
			"FROM %s WHERE f_task_id = ?", EVENT_TASK_TABLE_NAME)

		Convey("GetEventTaskByTaskID Success \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			task, err := ema.GetEventTaskByTaskID(testCtx, "1")
			task.DownstreamDependentTask = []string{}
			So(err, ShouldBeNil)
			So(task, ShouldResemble, eventTask)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			task, err := ema.GetEventTaskByTaskID(testCtx, "1")
			So(task, ShouldResemble, interfaces.EventTask{})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by query error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, err := ema.GetEventTaskByTaskID(testCtx, "1")
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by Schedule unmarshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFuncReturn(sonic.Unmarshal, expectedErr)
			defer patch.Reset()

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			task, err := ema.GetEventTaskByTaskID(testCtx, "1")
			So(task, ShouldResemble, interfaces.EventTask{})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by storageConfig unmarshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFunc(sonic.Unmarshal,
				func(data []byte, v any) error {
					if _, ok := v.(*interfaces.StorageConfig); ok {
						return expectedErr
					}
					return nil
				},
			)
			defer patch.Reset()

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)
			task, err := ema.GetEventTaskByTaskID(testCtx, "1")
			So(task, ShouldResemble, interfaces.EventTask{})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by dispatchConfig unmarshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFunc(sonic.Unmarshal,
				func(data []byte, v any) error {
					if _, ok := v.(*interfaces.DispatchConfig); ok {
						return expectedErr
					}
					return nil
				},
			)
			defer patch.Reset()

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			task, err := ema.GetEventTaskByTaskID(testCtx, "1")
			So(task, ShouldResemble, interfaces.EventTask{})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by executeParameter unmarshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFunc(sonic.Unmarshal,
				func(data []byte, v any) error {
					if _, ok := v.(*map[string]any); ok {
						return expectedErr
					}
					return nil
				},
			)
			defer patch.Reset()

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)
			task, err := ema.GetEventTaskByTaskID(testCtx, "1")
			So(task, ShouldResemble, interfaces.EventTask{})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_EventModelAccess_GetEventTaskByModelID(t *testing.T) {
	Convey("Test GetEventTaskByModelID\n", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mma := dmock.NewMockMetricModelAccess(mockCtrl)
		ema, smock := MockNewEventModelAccess(appSetting, mma)

		rows := sqlmock.NewRows([]string{
			"f_task_id",
			"f_model_id",
			"f_schedule",
			"f_dispatch_config",
			"f_execute_parameter",
			"f_storage_config",
			"f_task_status",
			"f_status_update_time",
			"f_error_details",
			"f_schedule_sync_status",
			"f_downstream_dependent_task",
			"f_creator",
			"f_creator_type",
			"f_create_time",
			"f_update_time",
		}).AddRow(
			eventTask.TaskID,
			eventTask.ModelID,
			eventScheduleBytes,
			dispatchConfigBytes,
			executeParameterBytes,
			storageConfigBytes,
			eventTask.TaskStatus,
			eventTask.StatusUpdateTime,
			eventTask.ErrorDetails,
			eventTask.ScheduleSyncStatus,
			downstreamDependentTask,
			eventTask.Creator.ID,
			eventTask.Creator.Type,
			eventTask.CreateTime,
			eventTask.UpdateTime)

		sqlStr := fmt.Sprintf("SELECT f_task_id, f_model_id, f_schedule, f_dispatch_config, "+
			"f_execute_parameter, f_storage_config, f_task_status, f_status_update_time, f_error_details, "+
			"f_schedule_sync_status, COALESCE(f_downstream_dependent_task, ''), f_creator, f_creator_type, "+
			"f_create_time, f_update_time FROM %s WHERE f_model_id = ?", EVENT_TASK_TABLE_NAME)

		Convey("GetEventTaskByModelID Success \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			task, exist, err := ema.GetEventTaskByModelID(testCtx, "1")
			So(err, ShouldBeNil)
			So(exist, ShouldBeTrue)
			So(task, ShouldResemble, eventTask)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			task, exist, err := ema.GetEventTaskByModelID(testCtx, "1")
			So(task, ShouldResemble, interfaces.EventTask{})
			So(exist, ShouldBeFalse)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by query error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			task, exist, err := ema.GetEventTaskByModelID(testCtx, "1")
			So(task, ShouldResemble, interfaces.EventTask{})
			So(exist, ShouldBeFalse)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Check failed, caused by the scan error", func() {
			expectedErr := errors.New("sql: expected 2 destination arguments in Scan, not 15")
			rows := sqlmock.NewRows([]string{"f_task_id", "f_task_id"}).AddRow("1", "")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			task, exist, err := ema.GetEventTaskByModelID(testCtx, "1")
			So(task, ShouldResemble, interfaces.EventTask{})
			So(exist, ShouldBeFalse)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by Schedule unmarshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFuncReturn(sonic.Unmarshal, expectedErr)
			defer patch.Reset()

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			task, exist, err := ema.GetEventTaskByModelID(testCtx, "1")
			So(task, ShouldResemble, interfaces.EventTask{})
			So(exist, ShouldBeFalse)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by storageConfig unmarshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFunc(sonic.Unmarshal,
				func(data []byte, v any) error {
					if _, ok := v.(*interfaces.StorageConfig); ok {
						return expectedErr
					}
					return nil
				},
			)
			defer patch.Reset()

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			task, exist, err := ema.GetEventTaskByModelID(testCtx, "1")
			So(task, ShouldResemble, interfaces.EventTask{})
			So(exist, ShouldBeFalse)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by dispatchConfig unmarshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFunc(sonic.Unmarshal,
				func(data []byte, v any) error {
					if _, ok := v.(*interfaces.DispatchConfig); ok {
						return expectedErr
					}
					return nil
				},
			)
			defer patch.Reset()

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			task, exist, err := ema.GetEventTaskByModelID(testCtx, "1")
			So(task, ShouldResemble, interfaces.EventTask{})
			So(exist, ShouldBeFalse)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by executeParameter unmarshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFunc(sonic.Unmarshal,
				func(data []byte, v any) error {
					if _, ok := v.(*map[string]any); ok {
						return expectedErr
					}
					return nil
				},
			)
			defer patch.Reset()

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			task, exist, err := ema.GetEventTaskByModelID(testCtx, "1")
			So(task, ShouldResemble, interfaces.EventTask{})
			So(exist, ShouldBeFalse)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_EventModelAccess_UpdateEventTaskStatusInFinish(t *testing.T) {
	Convey("test UpdateEventTaskStatusInFinish\n", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mma := dmock.NewMockMetricModelAccess(mockCtrl)
		ema, smock := MockNewEventModelAccess(appSetting, mma)

		sqlStr := fmt.Sprintf("UPDATE %s SET f_schedule_sync_status = ?, f_update_time = ? "+
			"WHERE f_task_id = ? AND f_update_time <= ? ", EVENT_TASK_TABLE_NAME)

		Convey("UpdateEventTaskStatusInFinish Success \n", func() {
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			err := ema.UpdateEventTaskStatusInFinish(testCtx, eventTask)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateEventTaskStatusInFinish Exec sql error\n", func() {
			expectedErr := errors.New("some error1")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			err := ema.UpdateEventTaskStatusInFinish(testCtx, eventTask)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateEventTaskStatusInFinish affected > 1 \n", func() {
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 2))

			err := ema.UpdateEventTaskStatusInFinish(testCtx, eventTask)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateEventTaskStatusInFinish failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.UpdateBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			err := ema.UpdateEventTaskStatusInFinish(testCtx, eventTask)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_EventModelAccess_UpdateEventTaskAttributes(t *testing.T) {
	Convey("test UpdateEventTaskAttributes\n", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mma := dmock.NewMockMetricModelAccess(mockCtrl)
		ema, smock := MockNewEventModelAccess(appSetting, mma)

		sqlStr := fmt.Sprintf("UPDATE %s SET f_error_details = ?, f_status_update_time = ?, "+
			"f_task_status = ? WHERE f_task_id = ? ", EVENT_TASK_TABLE_NAME)

		Convey("UpdateEventTaskAttributes Success \n", func() {
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			err := ema.UpdateEventTaskAttributes(testCtx, eventTask)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateEventTaskAttributes Exec sql error\n", func() {
			expectedErr := errors.New("some error1")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			err := ema.UpdateEventTaskAttributes(testCtx, eventTask)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateEventTaskAttributes affected > 1 \n", func() {
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 2))

			err := ema.UpdateEventTaskAttributes(testCtx, eventTask)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateEventTaskAttributes failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.UpdateBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			err := ema.UpdateEventTaskAttributes(testCtx, eventTask)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_EventModelAccess_DeleteEventTaskByTaskIDs(t *testing.T) {
	Convey("Test DeleteEventTaskByTaskIDs", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mma := dmock.NewMockMetricModelAccess(mockCtrl)
		ema, smock := MockNewEventModelAccess(appSetting, mma)

		sqlStr := fmt.Sprintf("DELETE FROM %s WHERE f_task_id IN (?)", EVENT_TASK_TABLE_NAME)

		Convey("Delete failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.DeleteBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			smock.ExpectBegin()

			tx, _ := ema.db.Begin()
			err := ema.DeleteEventTaskByTaskIDs(testCtx, tx, []string{"1"})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Delete failed, caused by exec sql error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := ema.db.Begin()
			err := ema.DeleteEventTaskByTaskIDs(testCtx, tx, []string{"1"})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Delete succeed", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 2))

			tx, _ := ema.db.Begin()
			err := ema.DeleteEventTaskByTaskIDs(testCtx, tx, []string{"1"})
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}
