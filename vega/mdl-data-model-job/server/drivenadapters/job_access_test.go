// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package drivenadapters

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	sq "github.com/Masterminds/squirrel"
	. "github.com/agiledragon/gomonkey/v2"
	"github.com/bytedance/sonic"
	. "github.com/smartystreets/goconvey/convey"

	"data-model-job/common"
	cond "data-model-job/common/condition"
	"data-model-job/interfaces"
)

var (
	testCtx = context.WithValue(context.WithValue(context.Background(),
		rest.XLangKey, rest.DefaultLanguage),
		interfaces.ACCOUNT_INFO_KEY, interfaces.AccountInfo{
			ID:   interfaces.ADMIN_ID,
			Type: interfaces.ADMIN_TYPE,
		})
)

var (
	dataSource = map[string]any{
		"type": "index_base",
		"index_base": []interfaces.SimpleIndexBase{
			{BaseType: "x"},
			{BaseType: "y"},
			{BaseType: "z"},
		},
	}

	fields = []*cond.Field{
		{Name: "错误ID", Type: "text", Comment: ""},
		{Name: "语言", Type: "text", Comment: ""},
	}

	condCfg = &cond.CondCfg{
		Name:      "签名",
		Operation: "==",
		ValueOptCfg: cond.ValueOptCfg{
			Value:     "kk",
			ValueFrom: cond.ValueFrom_Const,
		},
	}

	dataSourceBytes, _ = sonic.Marshal(dataSource)
	fieldsBytes, _     = sonic.Marshal(fields)
	condBytes, _       = sonic.Marshal(condCfg)

	view = &interfaces.DataView{
		ViewId:     "1",
		DataSource: dataSource,
		FieldScope: interfaces.ALL,
		Fields:     fields,
		Condition:  condCfg,
	}

	job = interfaces.JobInfo{
		JobId:            "2a",
		JobType:          interfaces.JOB_TYPE_STREAM,
		JobConfig:        map[string]any{},
		JobStatus:        interfaces.JobStatus_Running,
		JobStatusDetails: "",
		UpdateTime:       common.GetCurrentTimestamp(),
		DataView:         *view,
	}
)

func MockNewJobAccess() (interfaces.JobAccess, sqlmock.Sqlmock) {
	db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	ja := &jobAccess{
		db: db,
	}
	return ja, smock
}

func Test_JobAccess_ListViewJobs(t *testing.T) {
	Convey("Test ListViewJobs", t, func() {
		row := sqlmock.NewRows([]string{
			"f_job_id", "f_job_type", "f_job_status", "f_job_status_details", "f_create_time", "f_update_time",
			"f_creator", "f_creator_type", "f_view_id", "f_data_source", "f_field_scope", "f_fields", "f_filters"},
		).AddRow(
			job.JobId, job.JobType, job.JobStatus, job.JobStatusDetails, job.CreateTime, job.UpdateTime,
			job.Creator.ID, job.Creator.Type, job.ViewId, dataSourceBytes, job.FieldScope, fieldsBytes, condBytes)

		sqlStr := fmt.Sprintf("SELECT j.f_job_id, j.f_job_type, j.f_job_status, j.f_job_status_details, "+
			"j.f_create_time, j.f_update_time, j.f_creator, j.f_creator_type, COALESCE(v.f_view_id, ''), v.f_data_source, "+
			"COALESCE(v.f_field_scope, 0), v.f_fields, v.f_filters FROM %s as j LEFT JOIN %s as v "+
			"on j.f_job_id = v.f_job_id WHERE j.f_job_type = ?", DATA_MODEL_JOB_TABLE_NAME, DATA_VIEW_TABLE_NAME)

		ja, smock := MockNewJobAccess()

		Convey("List failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			_, err := ja.ListViewJobs()
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by query error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, err := ja.ListViewJobs()
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by scan error", func() {
			rowsErr := sqlmock.NewRows([]string{"f_job_id"}).AddRow(job.ViewId)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rowsErr)

			_, err := ja.ListViewJobs()
			So(err, ShouldNotBeNil)
		})

		Convey("Get failed, caused by dataSource unmarshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFuncReturn(sonic.Unmarshal, expectedErr)
			defer patch.Reset()

			smock.ExpectQuery(sqlStr).WithArgs(interfaces.JOB_TYPE_STREAM).WillReturnRows(row)

			_, err := ja.ListViewJobs()
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by fields unmarshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFunc(sonic.Unmarshal,
				func(data []byte, v any) error {
					if _, ok := v.(*[]*cond.Field); ok {
						return expectedErr
					}
					return nil
				},
			)
			defer patch.Reset()

			smock.ExpectQuery(sqlStr).WithArgs(interfaces.JOB_TYPE_STREAM).WillReturnRows(row)

			_, err := ja.ListViewJobs()
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by condition unmarshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFunc(sonic.Unmarshal,
				func(data []byte, v any) error {
					if _, ok := v.(**cond.CondCfg); ok {
						return expectedErr
					}
					return nil
				},
			)
			defer patch.Reset()

			smock.ExpectQuery(sqlStr).WithArgs(interfaces.JOB_TYPE_STREAM).WillReturnRows(row)

			_, err := ja.ListViewJobs()
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get succeed, and return correct rows", func() {
			smock.ExpectQuery(sqlStr).WithArgs(interfaces.JOB_TYPE_STREAM).WillReturnRows(row)

			_, err := ja.ListViewJobs()
			So(err, ShouldBeNil)
		})
	})
}

func Test_JobAccess_UpdateJobStatus(t *testing.T) {
	Convey("Test UpdateDataView", t, func() {
		ja, smock := MockNewJobAccess()

		sqlStr := fmt.Sprintf("UPDATE %s SET f_job_status = ?, f_job_status_details = ? "+
			"WHERE f_job_id = ?", DATA_MODEL_JOB_TABLE_NAME)

		Convey("Update failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.UpdateBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			err := ja.UpdateJobStatus(job)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Update failed, caused by exec sql error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			err := ja.UpdateJobStatus(job)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Update succeed", func() {
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			err := ja.UpdateJobStatus(job)
			So(err, ShouldBeNil)
		})
	})
}

func Test_JobAccess_ListMetricJobs(t *testing.T) {
	Convey("test ListMetricJobs\n", t, func() {
		ja, smock := MockNewJobAccess()

		task1 := interfaces.MetricTask{
			TaskID:          "1",
			TaskName:        "task1",
			Comment:         "task1-aaa",
			ModuleType:      interfaces.MODULE_TYPE_METRIC_MODEL,
			ModelID:         "1",
			MeasureName:     "__m.ddd",
			Schedule:        interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
			TimeWindows:     []string{"5m", "1h"},
			Steps:           []string{"5m"},
			PlanTime:        int64(1699336878575),
			IndexBase:       "base1",
			RetraceDuration: "1d",
		}

		rows := sqlmock.NewRows([]string{
			"task.f_task_id", "task.f_task_name", "task.f_comment", "task.f_module_type",
			"task.f_model_id", "model.f_measure_name", "task.f_schedule", "task.f_time_windows",
			"task.f_steps", "task.f_plan_time", "task.f_index_base", "task.f_retrace_duration",
			"task.f_creator", "task.f_creator_type"},
		).AddRow(
			task1.TaskID, task1.TaskName, task1.Comment, task1.ModuleType,
			task1.ModelID, task1.MeasureName, `{"type":"FIX_RATE","expression":"1m"}`, `["5m","1h"]`,
			`["5m"]`, task1.PlanTime, task1.IndexBase, task1.RetraceDuration, task1.Creator.ID, task1.Creator.Type)

		sqlStr := fmt.Sprintf("SELECT task.f_task_id, task.f_task_name, task.f_comment, "+
			"task.f_module_type, task.f_model_id, model.f_measure_name, task.f_schedule, task.f_time_windows, "+
			"task.f_steps, task.f_plan_time, task.f_index_base, task.f_retrace_duration, task.f_creator, task.f_creator_type "+
			"FROM %s as task JOIN %s as model on task.f_model_id = model.f_model_id "+
			"WHERE task.f_module_type = ?", METRIC_TASK_TABLE_NAME, METRIC_MODEL_TABLE_NAME)

		Convey("ListMetricJobs Success \n", func() {
			expect := []interfaces.JobInfo{
				{
					JobId:      task1.TaskID,
					JobType:    interfaces.JOB_TYPE_SCHEDULE,
					ModuleType: interfaces.MODULE_TYPE_METRIC_MODEL,
					MetricTask: &task1,
					Schedule:   task1.Schedule,
				},
			}

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			tasks, err := ja.ListMetricJobs()
			So(err, ShouldBeNil)
			So(tasks, ShouldResemble, expect)
		})

		Convey("Get failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			tasks, err := ja.ListMetricJobs()
			So(tasks, ShouldBeNil)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by query error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			tasks, err := ja.ListMetricJobs()
			So(tasks, ShouldBeNil)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Check failed, caused by the scan error", func() {
			expectedErr := errors.New("sql: expected 2 destination arguments in Scan, not 14")
			rowsErr := sqlmock.NewRows([]string{"f_task_id", "f_task_id"}).
				AddRow(task.TaskName, task.TaskName)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rowsErr)

			tasks, err := ja.ListMetricJobs()
			So(tasks, ShouldBeNil)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by Schedule unmarshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFuncReturn(json.Unmarshal, expectedErr)
			defer patch.Reset()

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			tasks, err := ja.ListMetricJobs()
			So(tasks, ShouldResemble, []interfaces.JobInfo{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by Filters unmarshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFunc(json.Unmarshal,
				func(data []byte, v any) error {
					if _, ok := v.(*[]interfaces.Filter); ok {
						return expectedErr
					}
					return nil
				},
			)
			defer patch.Reset()

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			tasks, err := ja.ListMetricJobs()
			So(tasks, ShouldBeNil)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by TimeWindows unmarshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFunc(json.Unmarshal,
				func(data []byte, v any) error {
					if _, ok := v.(*[]string); ok {
						return expectedErr
					}
					return nil
				},
			)
			defer patch.Reset()

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			tasks, err := ja.ListMetricJobs()
			So(tasks, ShouldBeNil)
			So(err, ShouldResemble, expectedErr)
		})
	})
}

func Test_JobAccess_ListObjectiveJobs(t *testing.T) {
	Convey("test ListObjectiveJobs\n", t, func() {
		ja, smock := MockNewJobAccess()

		task1 := interfaces.MetricTask{
			TaskID:          "1",
			TaskName:        "task1",
			Comment:         "task1-aaa",
			ModuleType:      interfaces.MODULE_TYPE_OBJECTIVE_MODEL,
			ModelID:         "1",
			Schedule:        interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
			TimeWindows:     []string{"5m", "1h"},
			Steps:           []string{"5m"},
			PlanTime:        int64(1699336878575),
			IndexBase:       "base1",
			RetraceDuration: "1d",
		}

		rows := sqlmock.NewRows([]string{
			"task.f_task_id", "task.f_task_name", "task.f_comment", "task.f_module_type",
			"task.f_model_id", "task.f_schedule", "task.f_time_windows", "task.f_steps",
			"task.f_plan_time", "task.f_index_base", "task.f_retrace_duration", "task.f_creator", "task.f_creator_type",
		}).AddRow(
			task1.TaskID, task1.TaskName, task1.Comment, task1.ModuleType,
			task1.ModelID, `{"type":"FIX_RATE","expression":"1m"}`, `["5m","1h"]`, `["5m"]`,
			task1.PlanTime, task1.IndexBase, task1.RetraceDuration, task1.Creator.ID, task1.Creator.Type)

		sqlStr := fmt.Sprintf("SELECT task.f_task_id, task.f_task_name, task.f_comment, "+
			"task.f_module_type, task.f_model_id, task.f_schedule, task.f_time_windows, task.f_steps, "+
			"task.f_plan_time, task.f_index_base, task.f_retrace_duration, task.f_creator, task.f_creator_type "+
			"FROM %s as task "+
			"JOIN %s as model on task.f_model_id = model.f_model_id WHERE task.f_module_type = ?",
			METRIC_TASK_TABLE_NAME, OBJECTIVE_MODEL_TABLE_NAME)

		Convey("ListMetricJobs Success \n", func() {
			expect := []interfaces.JobInfo{
				{
					JobId:      task1.TaskID,
					JobType:    interfaces.JOB_TYPE_SCHEDULE,
					ModuleType: interfaces.MODULE_TYPE_OBJECTIVE_MODEL,
					MetricTask: &task1,
					Schedule:   task1.Schedule,
				},
			}

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			tasks, err := ja.ListObjectiveJobs()
			So(err, ShouldBeNil)
			So(tasks, ShouldResemble, expect)
		})

		Convey("Get failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			tasks, err := ja.ListObjectiveJobs()
			So(tasks, ShouldBeNil)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by query error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			tasks, err := ja.ListObjectiveJobs()
			So(tasks, ShouldBeNil)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Check failed, caused by the scan error", func() {
			expectedErr := errors.New("sql: expected 2 destination arguments in Scan, not 13")
			rowsErr := sqlmock.NewRows([]string{"f_task_id", "f_task_id"}).
				AddRow(task.TaskName, task.TaskName)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rowsErr)

			tasks, err := ja.ListObjectiveJobs()
			So(tasks, ShouldBeNil)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by Schedule unmarshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFunc(json.Unmarshal,
				func(data []byte, v any) error {
					return expectedErr
				},
			)
			defer patch.Reset()

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			tasks, err := ja.ListObjectiveJobs()
			So(tasks, ShouldResemble, []interfaces.JobInfo{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by Filters unmarshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFunc(json.Unmarshal,
				func(data []byte, v any) error {
					if _, ok := v.(*[]interfaces.Filter); ok {
						return expectedErr
					}
					return nil
				},
			)
			defer patch.Reset()

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			tasks, err := ja.ListObjectiveJobs()
			So(tasks, ShouldBeNil)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by TimeWindows unmarshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFunc(json.Unmarshal,
				func(data []byte, v any) error {
					if _, ok := v.(*[]string); ok {
						return expectedErr
					}
					return nil
				},
			)
			defer patch.Reset()

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			tasks, err := ja.ListObjectiveJobs()
			So(tasks, ShouldBeNil)
			So(err, ShouldResemble, expectedErr)
		})
	})
}

func Test_JobAccess_ListEventJobs(t *testing.T) {
	Convey("test ListEventJobs\n", t, func() {
		ja, smock := MockNewJobAccess()

		task1 := interfaces.EventTask{
			TaskID:   "1",
			ModelID:  "1",
			Schedule: interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
			DispatchConfig: interfaces.DispatchConfig{
				TimeOut:        3600,
				RouteStrategy:  "FIRST",
				BlockStrategy:  "",
				FailRetryCount: 3,
			},
			ExecuteParameter: map[string]any{},
			StorageConfig: interfaces.StorageConfig{
				IndexBase:    "base1",
				DataViewName: "view1",
			},
			TaskStatus:              4,
			StatusUpdateTime:        1699336878575,
			ErrorDetails:            "",
			ScheduleSyncStatus:      1,
			DownstreamDependentTask: []string{},
			UpdateTime:              1699336878575,
		}
		eventScheduleBytes, _ := json.Marshal(task1.Schedule)
		storageConfigBytes, _ := json.Marshal(task1.StorageConfig)
		dispatchConfigBytes, _ := json.Marshal(task1.DispatchConfig)
		executeParameterBytes, _ := json.Marshal(task1.ExecuteParameter)
		downstreamDependentTask := strings.Join(task1.DownstreamDependentTask, ",")

		rows := sqlmock.NewRows([]string{
			"task.f_task_id", "task.f_model_id", "task.f_schedule", "task.f_dispatch_config",
			"task.f_execute_parameter", "task.f_storage_config", "task.f_task_status",
			"task.f_status_update_time", "task.f_error_details", "task.f_schedule_sync_status",
			"task.f_downstream_dependent_task", "task.f_update_time", "task.f_creator", "task.f_creator_type"}).
			AddRow(
				task1.TaskID, task1.ModelID, eventScheduleBytes, dispatchConfigBytes, executeParameterBytes,
				storageConfigBytes, task1.TaskStatus, task1.StatusUpdateTime, task1.ErrorDetails,
				task1.ScheduleSyncStatus, downstreamDependentTask, task1.UpdateTime, task1.Creator.ID, task1.Creator.Type)

		sqlStr := fmt.Sprintf("SELECT task.f_task_id, task.f_model_id, task.f_schedule, task.f_dispatch_config, "+
			"task.f_execute_parameter, task.f_storage_config, task.f_task_status, task.f_status_update_time, "+
			"task.f_error_details, task.f_schedule_sync_status, task.f_downstream_dependent_task, "+
			"task.f_update_time, task.f_creator, task.f_creator_type "+
			"FROM %s as task JOIN %s as model on task.f_model_id = model.f_event_model_id "+
			"WHERE model.f_status = ? AND model.f_is_active = ?", EVENT_TASK_TABLE_NAME, EVENT_MODEL_TABLE_NAME)

		Convey("ListMetricJobs Success \n", func() {
			expect := []interfaces.JobInfo{
				{
					JobId:      task1.TaskID,
					JobType:    interfaces.JOB_TYPE_SCHEDULE,
					ModuleType: interfaces.MODULE_TYPE_EVENT_MODEL,
					EventTask:  &task1,
					Schedule:   task1.Schedule,
				},
			}
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			tasks, err := ja.ListEventJobs()
			So(err, ShouldBeNil)
			So(tasks, ShouldResemble, expect)
		})

		Convey("Get failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			tasks, err := ja.ListEventJobs()
			So(tasks, ShouldBeNil)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by query error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			tasks, err := ja.ListEventJobs()
			So(tasks, ShouldBeNil)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Check failed, caused by the scan error", func() {
			expectedErr := errors.New("sql: expected 2 destination arguments in Scan, not 14")
			rowsErr := sqlmock.NewRows([]string{"f_task_id", "f_task_id"}).
				AddRow(task.TaskName, task.TaskName)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rowsErr)

			tasks, err := ja.ListEventJobs()
			So(tasks, ShouldBeNil)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by Schedule unmarshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFuncReturn(json.Unmarshal, expectedErr)
			defer patch.Reset()

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			tasks, err := ja.ListEventJobs()
			So(tasks, ShouldResemble, []interfaces.JobInfo{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by Filters unmarshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFunc(json.Unmarshal,
				func(data []byte, v any) error {
					if _, ok := v.(*[]interfaces.Filter); ok {
						return expectedErr
					}
					return nil
				},
			)
			defer patch.Reset()

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			tasks, err := ja.ListEventJobs()
			So(tasks, ShouldBeNil)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by TimeWindows unmarshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFunc(json.Unmarshal,
				func(data []byte, v any) error {
					if _, ok := v.(*[]string); ok {
						return expectedErr
					}
					return nil
				},
			)
			defer patch.Reset()

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			tasks, err := ja.ListEventJobs()
			So(tasks, ShouldBeNil)
			So(err, ShouldResemble, expectedErr)
		})
	})
}
