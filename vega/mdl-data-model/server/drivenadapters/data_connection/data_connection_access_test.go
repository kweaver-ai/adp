// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_connection

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	sq "github.com/Masterminds/squirrel"
	. "github.com/agiledragon/gomonkey/v2"
	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	"data-model/common"
	"data-model/interfaces"
)

var (
	testCtx = context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage)
	testNow = time.Now().UnixMilli()
)

func MockNewDataConnectionAccess(appSetting *common.AppSetting) (*dataConnectionAccess, sqlmock.Sqlmock) {
	db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	dca := &dataConnectionAccess{
		appSetting: appSetting,
		db:         db,
	}

	return dca, smock
}

func Test_DataConnectonAccess_CreateDataConnecton(t *testing.T) {
	Convey("Test CreateDataConnecton", t, func() {
		appSetting := &common.AppSetting{}
		dca, smock := MockNewDataConnectionAccess(appSetting)

		sqlStr := fmt.Sprintf("INSERT INTO %s (f_connection_id,f_connection_name,f_tags,"+
			"f_comment,f_create_time,f_update_time,f_data_source_type,f_config,f_config_md5) "+
			"VALUES (?,?,?,?,?,?,?,?,?)", DATA_CONNECTION_TABLE_NAME)

		conn := interfaces.DataConnection{}

		Convey("Create failed, caused by the error from func ProcessBeforeStore", func() {
			expectedErr := errors.New("some error")
			patch := ApplyPrivateMethod(dca, "processBeforeStore",
				func(t *dataConnectionAccess, ctx context.Context, conn *interfaces.DataConnection) (string, []byte, error) {
					return "", []byte("{}"), expectedErr
				},
			)
			defer patch.Reset()

			smock.ExpectBegin()

			tx, _ := dca.db.Begin()
			err := dca.CreateDataConnection(testCtx, tx, &conn)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Create failed, caused by the error from squirrel func ToSql", func() {
			patch1 := ApplyPrivateMethod(dca, "processBeforeStore",
				func(t *dataConnectionAccess, ctx context.Context, conn *interfaces.DataConnection) (string, []byte, error) {
					return "", []byte("{}"), nil
				},
			)
			defer patch1.Reset()

			expectedErr := errors.New("some error")
			patch2 := ApplyMethodReturn(sq.InsertBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch2.Reset()

			smock.ExpectBegin()

			tx, _ := dca.db.Begin()
			err := dca.CreateDataConnection(testCtx, tx, &conn)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Create failed, caused by exec sql error", func() {
			patch := ApplyPrivateMethod(dca, "processBeforeStore",
				func(t *dataConnectionAccess, ctx context.Context, conn *interfaces.DataConnection) (string, []byte, error) {
					return "", []byte("{}"), nil
				},
			)
			defer patch.Reset()

			expectedErr := errors.New("some error")
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := dca.db.Begin()
			err := dca.CreateDataConnection(testCtx, tx, &conn)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Create succeed", func() {
			patch := ApplyPrivateMethod(dca, "processBeforeStore",
				func(t *dataConnectionAccess, ctx context.Context, conn *interfaces.DataConnection) (string, []byte, error) {
					return "", []byte("{}"), nil
				},
			)
			defer patch.Reset()

			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			tx, _ := dca.db.Begin()
			err := dca.CreateDataConnection(testCtx, tx, &conn)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataConnectonAccess_DeleteDataConnections(t *testing.T) {
	Convey("Test DeleteDataConnections", t, func() {
		appSetting := &common.AppSetting{}
		dca, smock := MockNewDataConnectionAccess(appSetting)

		sqlStr := fmt.Sprintf("DELETE FROM %s WHERE f_connection_id IN (?)", DATA_CONNECTION_TABLE_NAME)

		connIDs := []string{"1"}

		Convey("Delete failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.DeleteBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			smock.ExpectBegin()

			tx, _ := dca.db.Begin()
			err := dca.DeleteDataConnections(testCtx, tx, connIDs)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Delete failed, caused by exec sql error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := dca.db.Begin()
			err := dca.DeleteDataConnections(testCtx, tx, connIDs)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Delete succeed", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 2))

			tx, _ := dca.db.Begin()
			err := dca.DeleteDataConnections(testCtx, tx, connIDs)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataConnectonAccess_UpdateDataConnection(t *testing.T) {
	Convey("Test UpdateDataConnection", t, func() {
		appSetting := &common.AppSetting{}
		dca, smock := MockNewDataConnectionAccess(appSetting)

		sqlStr := fmt.Sprintf("UPDATE %s SET f_comment = ?, f_config = ?, "+
			"f_config_md5 = ?, f_connection_name = ?, f_tags = ?, f_update_time = ? "+
			"WHERE f_connection_id = ?", DATA_CONNECTION_TABLE_NAME)

		conn := interfaces.DataConnection{}

		Convey("Update failed, caused by marshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyPrivateMethod(dca, "processBeforeStore",
				func(t *dataConnectionAccess, ctx context.Context, conn *interfaces.DataConnection) (string, []byte, error) {
					return "", []byte("{}"), expectedErr
				},
			)
			defer patch.Reset()

			smock.ExpectBegin()

			tx, _ := dca.db.Begin()
			err := dca.UpdateDataConnection(testCtx, tx, &conn)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Update failed, caused by the error from squirrel func ToSql", func() {
			patch1 := ApplyPrivateMethod(dca, "processBeforeStore",
				func(t *dataConnectionAccess, ctx context.Context, conn *interfaces.DataConnection) (string, []byte, error) {
					return "", []byte("{}"), nil
				},
			)
			defer patch1.Reset()

			expectedErr := errors.New("some error")
			patch2 := ApplyMethodReturn(sq.UpdateBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch2.Reset()

			smock.ExpectBegin()

			tx, _ := dca.db.Begin()
			err := dca.UpdateDataConnection(testCtx, tx, &conn)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Update failed, caused by exec sql error", func() {

			patch := ApplyPrivateMethod(dca, "processBeforeStore",
				func(t *dataConnectionAccess, ctx context.Context, conn *interfaces.DataConnection) (string, []byte, error) {
					return "", []byte("{}"), nil
				},
			)
			defer patch.Reset()

			expectedErr := errors.New("some error")
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := dca.db.Begin()
			err := dca.UpdateDataConnection(testCtx, tx, &conn)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Update succeed", func() {
			patch := ApplyPrivateMethod(dca, "processBeforeStore",
				func(t *dataConnectionAccess, ctx context.Context, conn *interfaces.DataConnection) (string, []byte, error) {
					return "", []byte("{}"), nil
				},
			)
			defer patch.Reset()

			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			tx, _ := dca.db.Begin()
			err := dca.UpdateDataConnection(testCtx, tx, &conn)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataConnectonAccess_GetDataConnection(t *testing.T) {
	Convey("Test GetDataConnection", t, func() {
		appSetting := &common.AppSetting{}
		dca, smock := MockNewDataConnectionAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT dc.f_connection_id, dc.f_connection_name, dc.f_tags, "+
			"dc.f_comment, dc.f_create_time, dc.f_update_time, dc.f_data_source_type, dc.f_config, "+
			"dc.f_config_md5, dcs.f_status, dcs.f_detection_time FROM %s AS dc JOIN %s As dcs "+
			"on dc.f_connection_id = dcs.f_connection_id WHERE dc.f_connection_id = ?",
			DATA_CONNECTION_TABLE_NAME, DATA_CONNECTION_STATUS_TABLE_NAME)

		connID := "1"

		Convey("Get failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			_, _, err := dca.GetDataConnection(testCtx, connID)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by scan error", func() {
			expectedErr := errors.New("sql: expected 1 destination arguments in Scan, not 11")
			rows := sqlmock.NewRows([]string{"dc.f_connection_id"}).AddRow(connID)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, _, err := dca.GetDataConnection(testCtx, connID)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get succeed, and return correct rows", func() {
			rows := sqlmock.NewRows([]string{
				"dc.f_connection_id", "dc.f_connection_name", "dc.f_tags", "dc.f_comment",
				"dc.f_create_time", "dc.f_update_time", "dc.f_data_source_type", "dc.f_config",
				"dc.f_config_md5", "dcs.f_status", "dcs.f_detection_time",
			}).AddRow(
				"1", "conn1", "", "",
				testNow, testNow, "tingyun", []byte("{}"),
				"md5", "ok", testNow,
			)

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, _, err := dca.GetDataConnection(testCtx, connID)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataConnectonAccess_ListDataConnections(t *testing.T) {
	Convey("Test ListDataConnections", t, func() {
		appSetting := &common.AppSetting{}
		dca, smock := MockNewDataConnectionAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_connection_id, f_connection_name, f_tags, f_comment, "+
			"f_create_time, f_update_time, f_data_source_type FROM %s WHERE instr(f_connection_name, ?) > 0 "+
			"ORDER BY f_update_time desc LIMIT 1000 OFFSET 0", DATA_CONNECTION_TABLE_NAME)

		listQueryPara := interfaces.DataConnectionListQueryParams{
			CommonListQueryParams: interfaces.CommonListQueryParams{
				NamePattern: "a",
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:     interfaces.MAX_LIMIT,
					Offset:    interfaces.MIN_OFFSET,
					Sort:      interfaces.DATA_CONNECTION_SORT["update_time"],
					Direction: interfaces.DESC_DIRECTION,
				},
			},
		}

		Convey("List failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			_, err := dca.ListDataConnections(testCtx, listQueryPara)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("List failed, caused by query error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, err := dca.ListDataConnections(testCtx, listQueryPara)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("List failed, caused by row scan error", func() {
			expectedErr := errors.New("sql: expected 2 destination arguments in Scan, not 7")
			rows := sqlmock.NewRows(
				[]string{"f_connection_id", "f_connection_name"}).
				AddRow(1, "conn1")

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)
			_, err := dca.ListDataConnections(testCtx, listQueryPara)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("List succeed", func() {
			rows := sqlmock.NewRows([]string{"f_connection_id", "f_connection_name",
				"f_tags", "f_comment", "f_create_time", "f_update_time", "f_data_source_type"}).
				AddRow(1, "conn1", "", "", testNow, testNow, "tingyun")

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			entries, err := dca.ListDataConnections(testCtx, listQueryPara)
			So(err, ShouldBeNil)
			So(len(entries), ShouldEqual, 1)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataConnectonAccess_GetDataConnectionTotal(t *testing.T) {
	Convey("Test GetTraceModelTotal", t, func() {
		appSetting := &common.AppSetting{}
		dca, smock := MockNewDataConnectionAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT COUNT(f_connection_id) FROM %s "+
			"WHERE f_connection_name = ?", DATA_CONNECTION_TABLE_NAME)

		listQueryPara := interfaces.DataConnectionListQueryParams{
			CommonListQueryParams: interfaces.CommonListQueryParams{
				Name: "a",
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:     interfaces.MAX_LIMIT,
					Offset:    interfaces.MIN_OFFSET,
					Sort:      interfaces.DATA_CONNECTION_SORT["update_time"],
					Direction: interfaces.DESC_DIRECTION,
				},
			},
		}

		Convey("Get failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			total, err := dca.GetDataConnectionTotal(testCtx, listQueryPara)
			So(total, ShouldEqual, 0)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by the scan error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			total, err := dca.GetDataConnectionTotal(testCtx, listQueryPara)
			So(total, ShouldEqual, 0)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get succeed", func() {
			rows := sqlmock.NewRows([]string{"COUNT(f_connection_id)"}).AddRow(1)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			total, err := dca.GetDataConnectionTotal(testCtx, listQueryPara)
			So(total, ShouldEqual, 1)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataConnectonAccess_GetMapAboutName2ID(t *testing.T) {
	Convey("Test GetMapAboutName2ID", t, func() {
		appSetting := &common.AppSetting{}
		dca, smock := MockNewDataConnectionAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_connection_id, f_connection_name "+
			"FROM %s WHERE f_connection_name IN (?)", DATA_CONNECTION_TABLE_NAME)

		connNames := []string{"1"}

		Convey("Get failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			_, err := dca.GetMapAboutName2ID(testCtx, connNames)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by query error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, err := dca.GetMapAboutName2ID(testCtx, connNames)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by scan error", func() {
			expectedErr := errors.New("sql: expected 1 destination arguments in Scan, not 2")
			rows := sqlmock.NewRows([]string{"f_connection_name"}).AddRow(connNames[0])
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := dca.GetMapAboutName2ID(testCtx, connNames)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get succeed, and return correct rows", func() {
			rows := sqlmock.NewRows([]string{"f_connection_id", "f_connection_name"}).AddRow("1", "1")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := dca.GetMapAboutName2ID(testCtx, connNames)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataConnectonAccess_GetMapAboutID2Name(t *testing.T) {
	Convey("Test GetMapAboutID2Name", t, func() {
		appSetting := &common.AppSetting{}
		dca, smock := MockNewDataConnectionAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_connection_id, f_connection_name "+
			"FROM %s WHERE f_connection_id IN (?)", DATA_CONNECTION_TABLE_NAME)

		connIDs := []string{"1"}

		Convey("Get failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			_, err := dca.GetMapAboutID2Name(testCtx, connIDs)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by query error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, err := dca.GetMapAboutID2Name(testCtx, connIDs)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by scan error", func() {
			expectedErr := errors.New("sql: expected 1 destination arguments in Scan, not 2")
			rows := sqlmock.NewRows([]string{"f_connection_id"}).AddRow(connIDs[0])
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := dca.GetMapAboutID2Name(testCtx, connIDs)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get succeed, and return correct rows", func() {
			rows := sqlmock.NewRows([]string{"f_connection_id", "f_connection_name"}).AddRow("1", "1")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := dca.GetMapAboutID2Name(testCtx, connIDs)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataConnectonAccess_GetDataConnectionsByConfigMD5(t *testing.T) {
	Convey("Test GetDataConnectionsByConfigMD5", t, func() {
		appSetting := &common.AppSetting{}
		dca, smock := MockNewDataConnectionAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_connection_id, f_connection_name, f_data_source_type, "+
			"f_config_md5 FROM %s WHERE f_config_md5 = ?", DATA_CONNECTION_TABLE_NAME)

		configMD5 := "tingyun"

		Convey("Get failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			_, err := dca.GetDataConnectionsByConfigMD5(testCtx, configMD5)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by query error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, err := dca.GetDataConnectionsByConfigMD5(testCtx, configMD5)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by scan error", func() {
			expectedErr := errors.New("sql: expected 1 destination arguments in Scan, not 4")
			rows := sqlmock.NewRows([]string{"f_data_source_type"}).AddRow(configMD5)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := dca.GetDataConnectionsByConfigMD5(testCtx, configMD5)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get succeed, and return correct rows", func() {
			rows := sqlmock.NewRows([]string{"f_connection_id", "f_connection_name",
				"f_data_source_type", "f_config_md5"}).AddRow("1", "conn1", "tingyun", "tingyun")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := dca.GetDataConnectionsByConfigMD5(testCtx, configMD5)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataConnectonAccess_GetDataConnectionSourceType(t *testing.T) {
	Convey("Test GetDataConnectionSourceType", t, func() {
		appSetting := &common.AppSetting{}
		dca, smock := MockNewDataConnectionAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_data_source_type FROM %s "+
			"WHERE f_connection_id = ?", DATA_CONNECTION_TABLE_NAME)

		connID := "1"

		Convey("Get failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			_, _, err := dca.GetDataConnectionSourceType(testCtx, connID)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by scan error", func() {
			expectedErr := errors.New("sql: expected 2 destination arguments in Scan, not 1")
			rows := sqlmock.NewRows([]string{"f_data_source_type", "f_connection_id"}).AddRow("1", "1")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, _, err := dca.GetDataConnectionSourceType(testCtx, connID)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get succeed, and return correct rows", func() {
			rows := sqlmock.NewRows([]string{"dc.f_data_source_type"}).AddRow("tingyun")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, _, err := dca.GetDataConnectionSourceType(testCtx, connID)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataConnectonAccess_DataConnectionExtendSQLBuilder(t *testing.T) {
	Convey("Test extendSQLBuilder", t, func() {
		appSetting := &common.AppSetting{}
		dca, _ := MockNewDataConnectionAccess(appSetting)

		pageParams := interfaces.PaginationQueryParameters{
			Limit:     interfaces.MAX_LIMIT,
			Offset:    interfaces.MIN_OFFSET,
			Sort:      interfaces.DATA_CONNECTION_SORT["update_time"],
			Direction: interfaces.DESC_DIRECTION,
		}

		sqlBuilder := sq.Select(
			"f_connection_id",
			"f_connection_name",
			"f_tags",
			"f_comment",
			"f_create_time",
			"f_update_time",
			"f_data_source_type",
		).From(DATA_CONNECTION_TABLE_NAME)

		Convey("Name is not an empty string, and tag is an empty string", func() {
			expectedArgs := []interface{}{"test"}
			expectedStr := "SELECT f_connection_id, f_connection_name, f_tags, f_comment," +
				" f_create_time, f_update_time, f_data_source_type FROM " + DATA_CONNECTION_TABLE_NAME +
				" WHERE f_connection_name = ?"

			listQueryParams := interfaces.DataConnectionListQueryParams{
				CommonListQueryParams: interfaces.CommonListQueryParams{
					Name:                      "test",
					PaginationQueryParameters: pageParams,
				},
			}

			sqlStr, args, err := dca.extendSQLBuilder(listQueryParams, sqlBuilder).ToSql()
			So(sqlStr, ShouldEqual, expectedStr)
			So(args, ShouldResemble, expectedArgs)
			So(err, ShouldBeNil)
		})

		Convey("NamePattern is not an empty string, and tag is an empty string", func() {
			expectedArgs := []any{"test_1"}
			expectedStr := "SELECT f_connection_id, f_connection_name, f_tags, f_comment," +
				" f_create_time, f_update_time, f_data_source_type FROM " + DATA_CONNECTION_TABLE_NAME +
				" WHERE instr(f_connection_name, ?) > 0"

			listQueryParams := interfaces.DataConnectionListQueryParams{
				CommonListQueryParams: interfaces.CommonListQueryParams{
					NamePattern:               "test_1",
					PaginationQueryParameters: pageParams,
				},
			}
			sqlStr, args, err := dca.extendSQLBuilder(listQueryParams, sqlBuilder).ToSql()
			So(sqlStr, ShouldEqual, expectedStr)
			So(args, ShouldResemble, expectedArgs)
			So(err, ShouldBeNil)
		})

		Convey("Both name and tag are not empty strings", func() {
			expectedArgs := []any{"test", "\"tag_1\""}
			expectedStr := "SELECT f_connection_id, f_connection_name, f_tags, f_comment," +
				" f_create_time, f_update_time, f_data_source_type FROM " + DATA_CONNECTION_TABLE_NAME +
				" WHERE f_connection_name = ? AND instr(f_tags, ?) > 0"

			listQueryParams := interfaces.DataConnectionListQueryParams{
				CommonListQueryParams: interfaces.CommonListQueryParams{
					Name:                      "test",
					Tag:                       "tag_1",
					PaginationQueryParameters: pageParams,
				},
			}

			sqlStr, args, err := dca.extendSQLBuilder(listQueryParams, sqlBuilder).ToSql()
			So(sqlStr, ShouldEqual, expectedStr)
			So(args, ShouldResemble, expectedArgs)
			So(err, ShouldBeNil)
		})

		Convey("application_scope is not empty array", func() {
			expectedArgs := []any{"test", "\"tag_1\"", "tingyun"}
			expectedStr := "SELECT f_connection_id, f_connection_name, f_tags, f_comment," +
				" f_create_time, f_update_time, f_data_source_type FROM " + DATA_CONNECTION_TABLE_NAME +
				" WHERE f_connection_name = ? AND instr(f_tags, ?) > 0 AND f_data_source_type IN (?)"

			listQueryParams := interfaces.DataConnectionListQueryParams{
				ApplicationScope: []string{"trace_model"},
				CommonListQueryParams: interfaces.CommonListQueryParams{
					Name:                      "test",
					Tag:                       "tag_1",
					PaginationQueryParameters: pageParams,
				},
			}

			sqlStr, args, err := dca.extendSQLBuilder(listQueryParams, sqlBuilder).ToSql()
			So(sqlStr, ShouldEqual, expectedStr)
			So(args, ShouldResemble, expectedArgs)
			So(err, ShouldBeNil)
		})

		Convey("Both name and NamePattern are empty strings", func() {
			expectedStr := "SELECT f_connection_id, f_connection_name, f_tags, f_comment," +
				" f_create_time, f_update_time, f_data_source_type FROM " + DATA_CONNECTION_TABLE_NAME

			listQueryParams := interfaces.DataConnectionListQueryParams{
				CommonListQueryParams: interfaces.CommonListQueryParams{
					PaginationQueryParameters: pageParams,
				},
			}

			sqlStr, _, err := dca.extendSQLBuilder(listQueryParams, sqlBuilder).ToSql()
			So(sqlStr, ShouldEqual, expectedStr)
			So(err, ShouldBeNil)
		})
	})
}

func Test_DataConnectonAccess_DataConnectionProcessBeforeStore(t *testing.T) {
	Convey("Test processBeforeStore", t, func() {
		appSetting := &common.AppSetting{}
		dca, _ := MockNewDataConnectionAccess(appSetting)

		conn := interfaces.DataConnection{
			DataSourceConfig: map[string]string{},
		}

		Convey("Process failed, caused by detailed_config marshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFuncReturn(sonic.Marshal, []byte{}, expectedErr)
			defer patch.Reset()

			_, _, err := dca.processBeforeStore(testCtx, &conn)
			So(err, ShouldEqual, expectedErr)
		})

		Convey("Process success", func() {
			_, _, err := dca.processBeforeStore(testCtx, &conn)
			So(err, ShouldBeNil)
		})
	})
}

func Test_DataConnectonAccess_CreateDataConnectonStatus(t *testing.T) {
	Convey("Test CreateDataConnectonStatus", t, func() {
		appSetting := &common.AppSetting{}
		dca, smock := MockNewDataConnectionAccess(appSetting)

		sqlStr := fmt.Sprintf("INSERT INTO %s (f_connection_id,f_status,f_detection_time) "+
			"VALUES (?,?,?)", DATA_CONNECTION_STATUS_TABLE_NAME)

		status := interfaces.DataConnectionStatus{}

		Convey("Create failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.InsertBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			smock.ExpectBegin()

			tx, _ := dca.db.Begin()
			err := dca.CreateDataConnectionStatus(testCtx, tx, status)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Create failed, caused by exec sql error", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("some error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := dca.db.Begin()
			err := dca.CreateDataConnectionStatus(testCtx, tx, status)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Create succeed", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			tx, _ := dca.db.Begin()
			err := dca.CreateDataConnectionStatus(testCtx, tx, status)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataConnectonAccess_DeleteDataConnectionStatuses(t *testing.T) {
	Convey("Test DeleteDataConnectionStatuses", t, func() {
		appSetting := &common.AppSetting{}
		dca, smock := MockNewDataConnectionAccess(appSetting)

		sqlStr := fmt.Sprintf("DELETE FROM %s WHERE f_connection_id IN (?)", DATA_CONNECTION_STATUS_TABLE_NAME)

		connIDs := []string{"1"}

		Convey("Delete failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.DeleteBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			smock.ExpectBegin()

			tx, _ := dca.db.Begin()
			err := dca.DeleteDataConnectionStatuses(testCtx, tx, connIDs)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Delete failed, caused by exec sql error", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("some error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := dca.db.Begin()
			err := dca.DeleteDataConnectionStatuses(testCtx, tx, connIDs)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Delete succeed", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 2))

			tx, _ := dca.db.Begin()
			err := dca.DeleteDataConnectionStatuses(testCtx, tx, connIDs)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataConnectonAccess_UpdateDataConnectionStatus(t *testing.T) {
	Convey("Test UpdateDataConnectionStatus", t, func() {
		appSetting := &common.AppSetting{}
		dca, smock := MockNewDataConnectionAccess(appSetting)

		sqlStr := fmt.Sprintf("UPDATE %s SET f_detection_time = ?, f_status = ? "+
			"WHERE f_connection_id = ?", DATA_CONNECTION_STATUS_TABLE_NAME)

		status := interfaces.DataConnectionStatus{}

		Convey("Update failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.UpdateBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			smock.ExpectBegin()

			tx, _ := dca.db.Begin()
			err := dca.UpdateDataConnectionStatus(testCtx, tx, status)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Update failed, caused by exec sql error", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("some error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := dca.db.Begin()
			err := dca.UpdateDataConnectionStatus(testCtx, tx, status)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Update succeed", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			tx, _ := dca.db.Begin()
			err := dca.UpdateDataConnectionStatus(testCtx, tx, status)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}
