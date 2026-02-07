// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package metric_model

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/smartystreets/goconvey/convey"

	"data-model/common"
	"data-model/interfaces"
)

var (
	testMetricModelGroup = interfaces.MetricModelGroup{
		GroupID:    "1",
		GroupName:  "group1",
		Comment:    "this is group 1",
		CreateTime: testUpdateTime,
		UpdateTime: testUpdateTime,
	}
)

func MockNewMetricModelGroupAccess(appSetting *common.AppSetting) (*metricModelGroupAccess, sqlmock.Sqlmock) {
	db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	mmga := &metricModelGroupAccess{
		appSetting: appSetting,
		db:         db,
	}
	return mmga, smock
}

func Test_MetricModelGroupAccess_CreateMetricModelGroup(t *testing.T) {
	Convey("test CreateMetricModelGroup \n", t, func() {
		appSetting := &common.AppSetting{}
		mmga, smock := MockNewMetricModelGroupAccess(appSetting)

		sqlStr := fmt.Sprintf("INSERT INTO %s (f_group_id,"+
			"f_group_name,f_comment,f_create_time,f_update_time,f_builtin) "+
			"VALUES (?,?,?,?,?,?)", METRIC_MODEL_GROUP_TABLE_NAME)

		Convey("CreateMetricModelGroup Success \n", func() {
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			err := mmga.CreateMetricModelGroup(testCtx, nil, testMetricModelGroup)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CreateMetricModelGroup  Exec sql error\n", func() {
			expectedErr := errors.New("some error1")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			err := mmga.CreateMetricModelGroup(testCtx, nil, testMetricModelGroup)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_MetricModelGroupAccess_GetMetricModelGroupByID(t *testing.T) {
	Convey("test GetMetricModelGroupByID\n", t, func() {
		appSetting := &common.AppSetting{}
		mmga, smock := MockNewMetricModelGroupAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_group_id, f_group_name, "+
			"f_comment, f_create_time, f_update_time, f_builtin "+
			"FROM %s WHERE f_group_id = ?", METRIC_MODEL_GROUP_TABLE_NAME)

		rows := sqlmock.NewRows([]string{"f_group_id", "f_group_name",
			"f_comment", "f_create_time", "f_update_time", "f_builtin"},
		).AddRow("1", "group1", "this is group 1", testUpdateTime, testUpdateTime, false)

		Convey("GetMetricModelGroupByID Success \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			modelGroup, exists, err := mmga.GetMetricModelGroupByID(testCtx, "1")
			So(modelGroup, ShouldResemble, testMetricModelGroup)
			So(exists, ShouldBeTrue)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetMetricModelGroupByID Success no row \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(sql.ErrNoRows)

			modelGroup, exists, err := mmga.GetMetricModelGroupByID(testCtx, "1")
			So(modelGroup, ShouldResemble, interfaces.MetricModelGroup{})
			So(exists, ShouldBeFalse)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetMetricModelGroupByID Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			modelGroup, exists, err := mmga.GetMetricModelGroupByID(testCtx, "1")
			So(modelGroup, ShouldResemble, interfaces.MetricModelGroup{})
			So(exists, ShouldBeFalse)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_MetricModelGroupAccess_GetMetricModelGroupByName(t *testing.T) {
	Convey("test GetMetricModelGroupByName\n", t, func() {
		appSetting := &common.AppSetting{}
		mmga, smock := MockNewMetricModelGroupAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_group_id, f_group_name, "+
			"f_comment, f_create_time, f_update_time, f_builtin FROM %s "+
			"WHERE f_group_name = ?", METRIC_MODEL_GROUP_TABLE_NAME)

		rows := sqlmock.NewRows([]string{"f_group_id", "f_group_name",
			"f_comment", "f_create_time", "f_update_time", "f_builtin"},
		).AddRow("1", "group1", "this is group 1", testUpdateTime, testUpdateTime, false)

		Convey("GetMetricModelGroupByName Success \n", func() {
			smock.ExpectBegin()
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)
			tx, _ := mmga.db.Begin()
			_, exists, err := mmga.GetMetricModelGroupByName(testCtx, tx, "group1")
			So(exists, ShouldBeTrue)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetMetricModelGroupByName Success no row \n", func() {
			smock.ExpectBegin()
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(sql.ErrNoRows)
			tx, _ := mmga.db.Begin()

			_, exists, err := mmga.GetMetricModelGroupByName(testCtx, tx, "group1")
			So(exists, ShouldBeFalse)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetMetricModelGroupByName Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectBegin()
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)
			tx, _ := mmga.db.Begin()

			_, exists, err := mmga.GetMetricModelGroupByName(testCtx, tx, "group1")
			So(exists, ShouldBeFalse)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_MetricModelGroupAccess_UpdateMetricModelGroup(t *testing.T) {
	Convey("Test UpdateMetricModelGroup\n", t, func() {
		appSetting := &common.AppSetting{}
		mmga, smock := MockNewMetricModelGroupAccess(appSetting)

		sqlStr := fmt.Sprintf("UPDATE %s SET f_comment = ?, f_group_name = ?, "+
			"f_update_time = ? WHERE f_group_id = ?", METRIC_MODEL_GROUP_TABLE_NAME)

		Convey("UpdateMetricModelGroup Success \n", func() {
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			err := mmga.UpdateMetricModelGroup(testCtx, testMetricModelGroup)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateMetricModelGroup failed prepare \n", func() {
			expectedErr := errors.New("prepare error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			err := mmga.UpdateMetricModelGroup(testCtx, testMetricModelGroup)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateMetricModelGroup Failed UpdateSql \n", func() {
			expectedErr := errors.New("sql exec error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			err := mmga.UpdateMetricModelGroup(testCtx, testMetricModelGroup)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateMetricModelGroup RowsAffected 2 \n", func() {
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 2))

			err := mmga.UpdateMetricModelGroup(testCtx, testMetricModelGroup)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateMetricModelGroup Failed RowsAffected \n", func() {
			expectedErr := errors.New("Get RowsAffected error")
			smock.ExpectExec(sqlStr).WithArgs().
				WillReturnResult(sqlmock.NewErrorResult(expectedErr))

			err := mmga.UpdateMetricModelGroup(testCtx, testMetricModelGroup)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_MetricModelGroupAccess_ListMetricModelGroups(t *testing.T) {
	Convey("Test ListMetricModelGroups\n", t, func() {
		appSetting := &common.AppSetting{}
		mmga, smock := MockNewMetricModelGroupAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT "+
			"f_group_id, f_group_name, f_comment, f_create_time, "+
			"f_update_time, f_builtin FROM %s "+
			"WHERE f_builtin IN (?) ORDER BY f_group_name asc LIMIT 1000 OFFSET 0",
			METRIC_MODEL_GROUP_TABLE_NAME)

		rows := sqlmock.NewRows([]string{"f_group_id",
			"f_group_name", "f_comment", "f_create_time", "f_update_time", "f_builtin"},
		).AddRow("1", "group1", "", testUpdateTime, testUpdateTime, 0)

		groupQuery := interfaces.ListMetricGroupQueryParams{
			Builtin: []bool{false},
			PaginationQueryParameters: interfaces.PaginationQueryParameters{
				Limit:     interfaces.MAX_LIMIT,
				Offset:    interfaces.MIN_OFFSET,
				Sort:      interfaces.METRIC_MODEL_GROUP_SORT["group_name"],
				Direction: interfaces.ASC_DIRECTION,
			},
		}

		Convey("ListMetricModelGroups Success list \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := mmga.ListMetricModelGroups(testCtx, groupQuery)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListMetricModelGroups Success list no limit\n", func() {
			sqlStr := fmt.Sprintf("SELECT "+
				"f_group_id, f_group_name, f_comment, f_create_time, "+
				"f_update_time, f_builtin FROM %s "+
				"WHERE f_builtin IN (?) ORDER BY f_group_name asc",
				METRIC_MODEL_GROUP_TABLE_NAME)

			groupQuery.Limit = -1
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := mmga.ListMetricModelGroups(testCtx, groupQuery)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListMetricModelGroups Failed dbQuery \n", func() {
			expectedErr := errors.New("dbQuery error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, err := mmga.ListMetricModelGroups(testCtx, groupQuery)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListMetricModelGroups scan error \n", func() {
			rows := sqlmock.NewRows([]string{"metric_model_count", "f_group_id", "f_group_name"}).
				AddRow(20, "1", "group1")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := mmga.ListMetricModelGroups(testCtx, groupQuery)
			So(err.Error(), ShouldEqual, "sql: expected 3 destination arguments in Scan, not 6")

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_MetricModelGroupAccess_DeleteMetricModelGroup(t *testing.T) {
	Convey("Test DeleteMetricModelGroup\n", t, func() {
		appSetting := &common.AppSetting{}
		mmga, smock := MockNewMetricModelGroupAccess(appSetting)

		sqlStr := fmt.Sprintf("DELETE FROM %s WHERE f_group_id = ?", METRIC_MODEL_GROUP_TABLE_NAME)

		Convey("DeleteMetricModelGroup Success \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 2))

			tx, _ := mmga.db.Begin()
			rowsAffected, err := mmga.DeleteMetricModelGroup(testCtx, tx, "1")
			So(err, ShouldBeNil)
			So(rowsAffected, ShouldEqual, 2)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteMetricModelGroup failed prepare \n", func() {
			expectedErr := errors.New("prepare error")
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := mmga.db.Begin()
			_, err := mmga.DeleteMetricModelGroup(testCtx, tx, "1")
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteMetricModelGroup Failed dbExec \n", func() {
			expectedErr := errors.New("dbExec error")
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := mmga.db.Begin()
			_, err := mmga.DeleteMetricModelGroup(testCtx, tx, "1")
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}
