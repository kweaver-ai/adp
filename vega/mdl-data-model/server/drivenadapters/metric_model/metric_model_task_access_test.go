// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package metric_model

import (
	"errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	sq "github.com/Masterminds/squirrel"
	. "github.com/agiledragon/gomonkey/v2"
	"github.com/bytedance/sonic"
	. "github.com/smartystreets/goconvey/convey"

	"data-model/common"
	"data-model/interfaces"
)

var (
	testMetricTask = interfaces.MetricTask{
		TaskID:     "1",
		TaskName:   "task1",
		ModuleType: interfaces.MODULE_TYPE_OBJECTIVE_MODEL,
		ModelID:    "1",
		Schedule:   interfaces.Schedule{Type: "FIX_RATE", Expression: "1m"},
		// Filters:            []interfaces.Filter{{Name: "labels.mode", Operation: "=", Value: "idle"}},
		TimeWindows:        []string{"5m", "1h"},
		Steps:              []string{"5m"},
		IndexBase:          "base1",
		RetraceDuration:    "1d",
		ScheduleSyncStatus: 1,
		Comment:            "task1-aaa",
		UpdateTime:         testUpdateTime,
		PlanTime:           int64(1699336878575),
	}
)

func MockNewMetricModelTaskAccess(appSetting *common.AppSetting) (*metricModelTaskAccess, sqlmock.Sqlmock) {
	db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	mmta := &metricModelTaskAccess{
		appSetting: appSetting,
		db:         db,
	}
	return mmta, smock
}

func Test_MetricModelTaskAccess_CreateMetricTasks(t *testing.T) {
	Convey("test CreateMetricTasks\n", t, func() {
		appSetting := &common.AppSetting{}
		mmta, smock := MockNewMetricModelTaskAccess(appSetting)

		sqlStr := fmt.Sprintf("INSERT INTO %s (f_task_id,f_task_name,f_module_type,"+
			"f_model_id,f_schedule,f_time_windows,f_steps,f_index_base,f_retrace_duration,"+
			"f_schedule_sync_status,f_comment,f_update_time,f_plan_time,f_creator,f_creator_type) "+
			"VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", METRIC_MODEL_TASK_TABLE_NAME)

		Convey("CreateMetricTasks Success \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			tx, _ := mmta.db.Begin()
			err := mmta.CreateMetricTask(testCtx, tx, testMetricTask)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CreateMetricTasks prepare error \n", func() {
			expectedErr := errors.New("any error")
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := mmta.db.Begin()
			err := mmta.CreateMetricTask(testCtx, tx, testMetricTask)
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

			tx, _ := mmta.db.Begin()
			err := mmta.CreateMetricTask(testCtx, tx, testMetricTask)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Create failed, caused by TimeWindows marshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFunc(sonic.Marshal,
				func(v any) ([]byte, error) {
					if _, ok := v.([]string); ok {
						return []byte{}, expectedErr
					}
					return []byte{}, nil
				},
			)
			defer patch.Reset()

			smock.ExpectBegin()

			tx, _ := mmta.db.Begin()
			err := mmta.CreateMetricTask(testCtx, tx, testMetricTask)
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

			tx, _ := mmta.db.Begin()
			err := mmta.CreateMetricTask(testCtx, tx, testMetricTask)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CreateMetricTasks  Exec sql error\n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("some error1")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := mmta.db.Begin()
			err := mmta.CreateMetricTask(testCtx, tx, testMetricTask)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_MetricModelTaskAccess_GetMetricTaskIDsByModelIDs(t *testing.T) {
	Convey("test GetMetricTaskIDsByModelIDs\n", t, func() {
		appSetting := &common.AppSetting{}
		mmta, smock := MockNewMetricModelTaskAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_task_id FROM %s WHERE f_model_id IN (?)", METRIC_MODEL_TASK_TABLE_NAME)

		rows := sqlmock.NewRows([]string{"f_task_id"}).AddRow("1").AddRow("2")

		Convey("GetMetricTaskIDsByModelIDs Success \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			taskIDs, err := mmta.GetMetricTaskIDsByModelIDs(testCtx, []string{"1"})
			So(err, ShouldBeNil)
			So(taskIDs, ShouldResemble, []string{"1", "2"})

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			taskIDs, err := mmta.GetMetricTaskIDsByModelIDs(testCtx, []string{"1"})
			So(taskIDs, ShouldResemble, []string{})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by query error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			taskIDs, err := mmta.GetMetricTaskIDsByModelIDs(testCtx, []string{"1"})
			So(taskIDs, ShouldResemble, []string{})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Check failed, caused by the scan error", func() {
			expectedErr := errors.New("sql: expected 2 destination arguments in Scan, not 1")
			rows := sqlmock.NewRows([]string{"f_task_id", "f_task_id"}).
				AddRow(testMetricTask.TaskName, testMetricTask.TaskName)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			taskIDs, err := mmta.GetMetricTaskIDsByModelIDs(testCtx, []string{"1"})
			So(taskIDs, ShouldResemble, []string{})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_MetricModelTaskAccess_UpdateMetricTask(t *testing.T) {
	Convey("test UpdateMetricTask\n", t, func() {
		appSetting := &common.AppSetting{}
		mmta, smock := MockNewMetricModelTaskAccess(appSetting)

		sqlStr := fmt.Sprintf("UPDATE %s SET f_comment = ?, f_index_base = ?, "+
			"f_model_id = ?, f_module_type = ?, f_schedule = ?, f_schedule_sync_status = ?, "+
			"f_steps = ?, f_task_name = ?, f_time_windows = ?, f_update_time = ? "+
			"WHERE f_task_id = ?", METRIC_MODEL_TASK_TABLE_NAME)

		Convey("UpdateMetricTask Success \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			tx, _ := mmta.db.Begin()
			err := mmta.UpdateMetricTask(testCtx, tx, testMetricTask)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateMetricTask failed, caused by Schedule marshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFuncReturn(sonic.Marshal, []byte{}, expectedErr)
			defer patch.Reset()

			smock.ExpectBegin()

			tx, _ := mmta.db.Begin()
			err := mmta.UpdateMetricTask(testCtx, tx, testMetricTask)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateMetricTask failed, caused by TimeWindows marshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFunc(sonic.Marshal,
				func(v any) ([]byte, error) {
					if _, ok := v.([]string); ok {
						return []byte{}, expectedErr
					}
					return []byte{}, nil
				},
			)
			defer patch.Reset()

			smock.ExpectBegin()

			tx, _ := mmta.db.Begin()
			err := mmta.UpdateMetricTask(testCtx, tx, testMetricTask)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateMetricTask Exec sql error\n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("some error1")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := mmta.db.Begin()
			err := mmta.UpdateMetricTask(testCtx, tx, testMetricTask)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteMetricTasksInLogicByIDs failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.UpdateBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			smock.ExpectBegin()

			tx, _ := mmta.db.Begin()
			err := mmta.UpdateMetricTask(testCtx, tx, testMetricTask)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_MetricModelTaskAccess_GetMetricTasksByIDs(t *testing.T) {
	Convey("test GetMetricTasksByIDs\n", t, func() {
		appSetting := &common.AppSetting{}
		mmta, smock := MockNewMetricModelTaskAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_task_id, f_task_name, f_module_type, f_model_id, "+
			"f_schedule, f_time_windows, f_steps, f_index_base, f_retrace_duration, "+
			"f_schedule_sync_status, f_comment, f_update_time, f_plan_time FROM %s "+
			"WHERE f_task_id IN (?)", METRIC_MODEL_TASK_TABLE_NAME)

		rows := sqlmock.NewRows([]string{
			"f_task_id", "f_task_name", "f_module_type", "f_model_id", "f_schedule",
			"f_time_windows", "f_steps", "f_index_base", "f_retrace_duration",
			"f_schedule_sync_status", "f_comment", "f_update_time", "f_plan_time"},
		).AddRow(
			testMetricTask.TaskID, testMetricTask.TaskName, interfaces.MODULE_TYPE_OBJECTIVE_MODEL,
			testMetricTask.ModelID, `{"type":"FIX_RATE","expression":"1m"}`, `["5m","1h"]`, `["5m"]`,
			testMetricTask.IndexBase, testMetricTask.RetraceDuration, testMetricTask.ScheduleSyncStatus,
			testMetricTask.Comment, testMetricTask.UpdateTime, testMetricTask.PlanTime)

		Convey("GetMetricTasksByIDs Success \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			tasks, err := mmta.GetMetricTasksByTaskIDs(testCtx, []string{"1"})
			So(err, ShouldBeNil)
			So(tasks, ShouldResemble, []interfaces.MetricTask{testMetricTask})

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			tasks, err := mmta.GetMetricTasksByTaskIDs(testCtx, []string{"1"})
			So(tasks, ShouldResemble, []interfaces.MetricTask{})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by query error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			tasks, err := mmta.GetMetricTasksByTaskIDs(testCtx, []string{"1"})
			So(tasks, ShouldResemble, []interfaces.MetricTask{})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Check failed, caused by the scan error", func() {
			expectedErr := errors.New("sql: expected 2 destination arguments in Scan, not 13")
			rows := sqlmock.NewRows([]string{"f_task_id", "f_task_id"}).
				AddRow(testMetricTask.TaskName, testMetricTask.TaskName)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			tasks, err := mmta.GetMetricTasksByTaskIDs(testCtx, []string{"1"})
			So(tasks, ShouldResemble, []interfaces.MetricTask{})
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

			tasks, err := mmta.GetMetricTasksByTaskIDs(testCtx, []string{"1"})
			So(tasks, ShouldResemble, []interfaces.MetricTask{})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by TimeWindows unmarshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFunc(sonic.Unmarshal,
				func(data []byte, v any) error {
					if _, ok := v.(*[]string); ok {
						return expectedErr
					}
					return nil
				},
			)
			defer patch.Reset()

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			tasks, err := mmta.GetMetricTasksByTaskIDs(testCtx, []string{"1"})
			So(tasks, ShouldResemble, []interfaces.MetricTask{})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_MetricModelTaskAccess_GetMetricTasksByModelIDs(t *testing.T) {
	Convey("test GetMetricTasksByModelIDs\n", t, func() {
		appSetting := &common.AppSetting{}
		mmta, smock := MockNewMetricModelTaskAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_task_id, f_task_name, f_module_type, f_model_id, "+
			"f_schedule, f_time_windows, f_steps, f_index_base, f_retrace_duration, "+
			"f_schedule_sync_status, f_execute_status, f_comment, f_creator, f_creator_type, f_update_time, f_plan_time "+
			"FROM %s WHERE f_model_id IN (?)", METRIC_MODEL_TASK_TABLE_NAME)

		rows := sqlmock.NewRows([]string{
			"f_task_id", "f_task_name", "f_module_type", "f_model_id", "f_schedule",
			"f_time_windows", "f_steps", "f_index_base", "f_retrace_duration",
			"f_schedule_sync_status", "f_execute_status", "f_comment", "f_creator", "f_creator_type", "f_update_time", "f_plan_time"},
		).AddRow(
			"1", "16", "type", "id2", `{"type":"CRON","expression":"0 10 0 * * ?"}`, `["1d"]`, `["1d"]`,
			"base1", "365d", 3, 4, "ssss", interfaces.ADMIN_ID, interfaces.ADMIN_TYPE, testUpdateTime, 1732032300000)

		Convey("GetMetricTasksByIDs Success \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := mmta.GetMetricTasksByModelIDs(testCtx, []string{"1"})
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			tasks, err := mmta.GetMetricTasksByModelIDs(testCtx, []string{"1"})
			So(tasks, ShouldResemble, map[string]interfaces.MetricTask{})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by query error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, err := mmta.GetMetricTasksByModelIDs(testCtx, []string{"1"})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Check failed, caused by the scan error", func() {
			expectedErr := errors.New("sql: expected 2 destination arguments in Scan, not 16")
			rows := sqlmock.NewRows([]string{"f_task_id", "f_task_id"}).
				AddRow(testMetricTask.TaskName, testMetricTask.TaskName)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := mmta.GetMetricTasksByModelIDs(testCtx, []string{"1"})
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

			_, err := mmta.GetMetricTasksByModelIDs(testCtx, []string{"1"})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by TimeWindows unmarshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFunc(sonic.Unmarshal,
				func(data []byte, v any) error {
					if _, ok := v.(*[]string); ok {
						return expectedErr
					}
					return nil
				},
			)
			defer patch.Reset()

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			tasks, err := mmta.GetMetricTasksByModelIDs(testCtx, []string{"1"})
			So(tasks, ShouldResemble, map[string]interfaces.MetricTask{})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_MetricModelTaskAccess_UpdateMetricTaskStatusInFinish(t *testing.T) {
	Convey("test UpdateMetricTaskStatusInFinish\n", t, func() {
		appSetting := &common.AppSetting{}
		mmta, smock := MockNewMetricModelTaskAccess(appSetting)

		sqlStr := fmt.Sprintf("UPDATE %s SET f_schedule_sync_status = ?, f_update_time = ? "+
			"WHERE f_task_id = ? AND f_update_time <= ?", METRIC_MODEL_TASK_TABLE_NAME)

		Convey("UpdateMetricTaskStatusInFinish Success \n", func() {
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			err := mmta.UpdateMetricTaskStatusInFinish(testCtx, testMetricTask)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateMetricTaskStatusInFinish Exec sql error\n", func() {
			expectedErr := errors.New("some error1")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			err := mmta.UpdateMetricTaskStatusInFinish(testCtx, testMetricTask)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateMetricTaskStatusInFinish affected > 1 \n", func() {
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 2))

			err := mmta.UpdateMetricTaskStatusInFinish(testCtx, testMetricTask)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateMetricTaskStatusInFinish failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.UpdateBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			err := mmta.UpdateMetricTaskStatusInFinish(testCtx, testMetricTask)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_MetricModelTaskAccess_UpdateMetricTaskAttributes(t *testing.T) {
	Convey("test UpdateMetricTaskAttributes\n", t, func() {
		appSetting := &common.AppSetting{}
		mmta, smock := MockNewMetricModelTaskAccess(appSetting)

		sqlStr := fmt.Sprintf("UPDATE %s SET f_plan_time = ? "+
			"WHERE f_task_id = ?", METRIC_MODEL_TASK_TABLE_NAME)

		Convey("UpdateMetricTaskAttributes Success \n", func() {
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			err := mmta.UpdateMetricTaskAttributes(testCtx, testMetricTask)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateMetricTaskAttributes Exec sql error\n", func() {
			expectedErr := errors.New("some error1")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			err := mmta.UpdateMetricTaskAttributes(testCtx, testMetricTask)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateMetricTaskAttributes affected > 1 \n", func() {
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 2))

			err := mmta.UpdateMetricTaskAttributes(testCtx, testMetricTask)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateMetricTaskAttributes failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.UpdateBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			err := mmta.UpdateMetricTaskAttributes(testCtx, testMetricTask)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_MetricModelTaskAccess_DeleteMetricTaskByIDs(t *testing.T) {
	Convey("Test DeleteMetricTaskByTaskIDs", t, func() {
		appSetting := &common.AppSetting{}
		mmta, smock := MockNewMetricModelTaskAccess(appSetting)

		sqlStr := fmt.Sprintf("DELETE FROM %s WHERE f_task_id IN (?)", METRIC_MODEL_TASK_TABLE_NAME)

		Convey("Delete failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.DeleteBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			smock.ExpectBegin()

			tx, _ := mmta.db.Begin()

			err := mmta.DeleteMetricTaskByTaskIDs(testCtx, tx, []string{"1"})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Delete failed, caused by exec sql error", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("some error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := mmta.db.Begin()

			err := mmta.DeleteMetricTaskByTaskIDs(testCtx, tx, []string{"1"})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Delete succeed", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 2))

			tx, _ := mmta.db.Begin()

			err := mmta.DeleteMetricTaskByTaskIDs(testCtx, tx, []string{"1"})
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}
