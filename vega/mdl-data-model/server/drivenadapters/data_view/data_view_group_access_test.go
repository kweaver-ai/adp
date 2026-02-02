package data_view

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	sq "github.com/Masterminds/squirrel"
	. "github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"

	"data-model/common"
	"data-model/interfaces"
)

func MockNewDataViewGroupAccess(appSetting *common.AppSetting) (*dataViewGroupAccess, sqlmock.Sqlmock) {
	db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	dvga := &dataViewGroupAccess{
		appSetting: appSetting,
		db:         db,
	}
	return dvga, smock
}

func Test_DataViewGroupAccess_CreateDataViewGroup(t *testing.T) {
	Convey("Test CreateDataViewGroup", t, func() {
		appSetting := &common.AppSetting{}
		dvga, smock := MockNewDataViewGroupAccess(appSetting)

		sqlStr := fmt.Sprintf("INSERT INTO %s (f_group_id,f_group_name,f_create_time,"+
			"f_update_time,f_builtin) VALUES (?,?,?,?,?)", DATA_VIEW_GROUP_TABLE_NAME)

		group := &interfaces.DataViewGroup{
			GroupID:    "x",
			GroupName:  "x",
			CreateTime: testNow,
			UpdateTime: testNow,
		}

		Convey("Create failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.InsertBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			smock.ExpectBegin()

			tx, _ := dvga.db.Begin()
			err := dvga.CreateDataViewGroup(testCtx, tx, group)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Create failed, caused by exec sql error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := dvga.db.Begin()
			err := dvga.CreateDataViewGroup(testCtx, tx, group)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Create succeed", func() {
			// WithArgs()中不传参数, 应该代表任意参数下都会被 smock
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			tx, _ := dvga.db.Begin()
			err := dvga.CreateDataViewGroup(testCtx, tx, group)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataViewGroupAccess_DeleteDataViewsGroup(t *testing.T) {
	Convey("Test DeleteDataViewsGroup", t, func() {
		appSetting := &common.AppSetting{}
		dvga, smock := MockNewDataViewGroupAccess(appSetting)

		sqlStr := fmt.Sprintf("DELETE FROM %s WHERE f_group_id = ?", DATA_VIEW_GROUP_TABLE_NAME)

		Convey("Delete failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.DeleteBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			tx, _ := dvga.db.Begin()
			err := dvga.DeleteDataViewGroup(testCtx, tx, "1")
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Delete failed, caused by exec sql error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := dvga.db.Begin()
			err := dvga.DeleteDataViewGroup(testCtx, tx, "1")
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Delete succeed", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 2))

			tx, _ := dvga.db.Begin()
			err := dvga.DeleteDataViewGroup(testCtx, tx, "1")
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataViewGroupAccess_UpdateDataViewGroup(t *testing.T) {
	Convey("Test UpdateDataViewGroup", t, func() {
		appSetting := &common.AppSetting{}
		dvga, smock := MockNewDataViewGroupAccess(appSetting)

		sqlStr := fmt.Sprintf("UPDATE %s SET f_group_name = ?, "+
			"f_update_time = ? WHERE f_group_id = ? ", DATA_VIEW_GROUP_TABLE_NAME)

		group := &interfaces.DataViewGroup{
			GroupID:    "x",
			GroupName:  "x",
			CreateTime: testNow,
			UpdateTime: testNow,
		}

		Convey("Update failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.UpdateBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			err := dvga.UpdateDataViewGroup(testCtx, group)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Update failed, caused by exec sql error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			err := dvga.UpdateDataViewGroup(testCtx, group)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Update succeed", func() {
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			err := dvga.UpdateDataViewGroup(testCtx, group)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataViewGroupAccess_ListDataViewsGroup(t *testing.T) {
	Convey("Test ListDataViewsGroup", t, func() {
		appSetting := &common.AppSetting{}
		dvga, smock := MockNewDataViewGroupAccess(appSetting)

		group := &interfaces.DataViewGroup{
			GroupID:    "x",
			GroupName:  "x",
			CreateTime: testNow,
			UpdateTime: testNow,
		}

		rows1 := sqlmock.NewRows([]string{
			"f_group_id",
			"f_group_name",
			"f_create_time",
			"f_update_time",
			"f_builtin",
		}).AddRow(
			group.GroupID,
			group.GroupName,
			group.CreateTime,
			group.UpdateTime,
			false,
		)
		rows2 := sqlmock.NewRows([]string{
			"f_group_id",
			"f_group_name",
			"f_create_time",
			"f_update_time",
			"f_builtin",
		}).AddRow(
			group.GroupID,
			group.GroupName,
			group.CreateTime,
			group.UpdateTime,
			true,
		).AddRow(
			group.GroupID,
			group.GroupName,
			group.CreateTime,
			group.UpdateTime,
			false,
		)

		sqlStr := fmt.Sprintf("SELECT f_group_id, f_group_name, f_create_time, "+
			"f_update_time, f_builtin FROM %s WHERE f_builtin IN (?) AND f_delete_time = ? "+
			"ORDER BY f_update_time desc LIMIT 1000 OFFSET 0", DATA_VIEW_GROUP_TABLE_NAME)

		listQuery := &interfaces.ListViewGroupQueryParams{
			Builtin: []bool{false},
			PaginationQueryParameters: interfaces.PaginationQueryParameters{
				Limit:     interfaces.MAX_LIMIT,
				Offset:    interfaces.MIN_OFFSET,
				Sort:      interfaces.DATA_VIEW_GROUP_SORT["update_time"],
				Direction: interfaces.DESC_DIRECTION,
			},
		}

		Convey("List failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			_, err := dvga.ListDataViewGroups(testCtx, listQuery)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("List failed, caused by query error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, err := dvga.ListDataViewGroups(testCtx, listQuery)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("List failed, caused by row scan error", func() {
			rows := sqlmock.NewRows([]string{"f_group_id", "f_group_name", "f_create_time", "f_update_time"}).
				AddRow(group.GroupID, group.GroupName, group.CreateTime, group.UpdateTime)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := dvga.ListDataViewGroups(testCtx, listQuery)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("List succeed with limit is equal to 1", func() {
			sqlStr := fmt.Sprintf("SELECT f_group_id, f_group_name, f_create_time, "+
				"f_update_time, f_builtin FROM %s WHERE f_builtin IN (?) AND f_delete_time = ? "+
				"ORDER BY f_update_time desc LIMIT 1 OFFSET 0", DATA_VIEW_GROUP_TABLE_NAME)

			listQuery.Limit = 1
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows1)

			entries, err := dvga.ListDataViewGroups(testCtx, listQuery)
			So(err, ShouldBeNil)
			So(len(entries), ShouldEqual, 1)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("List succeed with no limit", func() {
			sqlStr := fmt.Sprintf("SELECT f_group_id, f_group_name, f_create_time, "+
				"f_update_time, f_builtin FROM %s WHERE f_builtin IN (?) AND f_delete_time = ? "+
				"ORDER BY f_update_time desc", DATA_VIEW_GROUP_TABLE_NAME)

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows2)

			listQuery.Limit = -1
			entries, err := dvga.ListDataViewGroups(testCtx, listQuery)
			So(err, ShouldBeNil)
			So(len(entries), ShouldEqual, 2)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataViewGroupAccess_GetDataViewGroupsTotal(t *testing.T) {
	Convey("Test GetDataViewGroupsTotal", t, func() {
		appSetting := &common.AppSetting{}
		dvga, smock := MockNewDataViewGroupAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT COUNT(f_group_id) FROM %s "+
			"WHERE f_builtin IN (?)", DATA_VIEW_GROUP_TABLE_NAME)

		rows := sqlmock.NewRows([]string{"COUNT(f_group_id)"}).AddRow(1)

		listQuery := &interfaces.ListViewGroupQueryParams{
			IncludeDeleted: true,
			Builtin:        []bool{true},
			PaginationQueryParameters: interfaces.PaginationQueryParameters{
				Limit:     interfaces.MAX_LIMIT,
				Offset:    interfaces.MIN_OFFSET,
				Sort:      interfaces.DATA_VIEW_GROUP_SORT["name"],
				Direction: interfaces.DESC_DIRECTION,
			},
		}

		Convey("Get failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			total, err := dvga.GetDataViewGroupsTotal(testCtx, listQuery)
			So(total, ShouldEqual, 0)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by the scan error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			total, err := dvga.GetDataViewGroupsTotal(testCtx, listQuery)
			So(total, ShouldEqual, 0)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get succeed", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)
			total, err := dvga.GetDataViewGroupsTotal(testCtx, listQuery)

			So(total, ShouldEqual, 1)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataViewGroupAccess_GetDataViewGroupByID(t *testing.T) {
	Convey("Test GetDataViewGroupByID", t, func() {
		appSetting := &common.AppSetting{}
		dvga, smock := MockNewDataViewGroupAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_group_id, f_group_name, f_create_time, "+
			"f_update_time, f_builtin FROM %s WHERE f_group_id = ?", DATA_VIEW_GROUP_TABLE_NAME)

		group := &interfaces.DataViewGroup{
			GroupID:    "x",
			GroupName:  "x",
			CreateTime: testNow,
			UpdateTime: testNow,
		}

		Convey("Get failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			_, _, err := dvga.GetDataViewGroupByID(testCtx, group.GroupID)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by query error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, _, err := dvga.GetDataViewGroupByID(testCtx, group.GroupID)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by ErrNoRows", func() {
			expectedErr := sql.ErrNoRows
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, exist, err := dvga.GetDataViewGroupByID(testCtx, group.GroupID)
			So(exist, ShouldBeFalse)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get succeed", func() {
			rows := sqlmock.NewRows([]string{"f_group_id", "f_group_name", "f_create_time", "f_update_time", "f_builtin"}).
				AddRow(group.GroupID, group.GroupName, group.CreateTime, group.UpdateTime, false)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)
			_, exist, err := dvga.GetDataViewGroupByID(testCtx, group.GroupID)

			So(exist, ShouldBeTrue)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataViewGroupAccess_CheckDataViewGroupExistByName(t *testing.T) {
	Convey("Test CheckDataViewGroupExistByName", t, func() {
		appSetting := &common.AppSetting{}
		dvga, smock := MockNewDataViewGroupAccess(appSetting)

		rows := sqlmock.NewRows([]string{"f_group_id", "f_builtin"}).AddRow(views[0].GroupID, views[0].Builtin)

		sqlStr := fmt.Sprintf("SELECT f_group_id, f_builtin FROM %s WHERE f_builtin = ? AND f_group_name = ?", DATA_VIEW_GROUP_TABLE_NAME)

		group := &interfaces.DataViewGroup{
			GroupID:    "x",
			GroupName:  "x",
			Builtin:    false,
			CreateTime: testNow,
			UpdateTime: testNow,
		}

		Convey("Check failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			_, _, err := dvga.CheckDataViewGroupExistByName(testCtx, nil, group.GroupName, group.Builtin)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by query error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, _, err := dvga.CheckDataViewGroupExistByName(testCtx, nil, group.GroupName, group.Builtin)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Check failed, caused by the scan error", func() {
			rows = sqlmock.NewRows([]string{"f_group_id", "f_builtin", "f_group_name"}).
				AddRow(group.GroupID, group.Builtin, group.GroupName)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, _, err := dvga.CheckDataViewGroupExistByName(testCtx, nil, group.GroupName, group.Builtin)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by ErrNoRows", func() {
			expectedErr := sql.ErrNoRows
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, exist, err := dvga.CheckDataViewGroupExistByName(testCtx, nil, group.GroupName, group.Builtin)
			So(exist, ShouldBeFalse)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Check succeed", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, exist, err := dvga.CheckDataViewGroupExistByName(testCtx, nil, group.GroupName, group.Builtin)
			So(exist, ShouldBeTrue)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}
