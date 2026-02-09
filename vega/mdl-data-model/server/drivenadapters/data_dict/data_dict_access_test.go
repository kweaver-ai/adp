// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_dict

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
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

func MockNewDataDictAccess(appSetting *common.AppSetting) (*dataDictAccess, sqlmock.Sqlmock) {
	db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	dda := &dataDictAccess{
		appSetting: appSetting,
		db:         db,
	}
	return dda, smock
}

func Test_DataDictAccess_ListDataDicts(t *testing.T) {
	Convey("Test DictAccess ListDataDicts\n", t, func() {
		appSetting := &common.AppSetting{}
		dda, smock := MockNewDataDictAccess(appSetting)

		rows := sqlmock.NewRows([]string{"f_dict_id", "f_dict_name", "f_dict_type",
			"f_unique_key", "f_dimension", "f_tags", "f_comment", "f_create_time", "f_update_time"}).
			AddRow("1", "1", "kv_dict", true,
				"{\"keys\":[{\"id\":\"item_key\",\"name\":\"key\"}],\"values\":[{\"id\":\"item_value\",\"name\":\"value\"}]}",
				"\"a\",\"b\",\"c\"", "comment222", 1638953462000, 1638953462000)

		dictQuery := interfaces.DataDictQueryParams{
			NamePattern: "",
			PaginationQueryParameters: interfaces.PaginationQueryParameters{
				Limit:     interfaces.MAX_LIMIT,
				Offset:    interfaces.MIN_OFFSET,
				Sort:      interfaces.DATA_DICT_SORT["update_time"],
				Direction: interfaces.DESC_DIRECTION,
			},
		}

		sqlStr := fmt.Sprintf("SELECT f_dict_id, f_dict_name, f_dict_type, f_unique_key, "+
			"f_dimension, f_tags, f_comment, f_create_time, f_update_time "+
			"FROM %s ORDER BY f_update_time desc", DATA_DICT_TABLE_NAME)

		Convey("ListDataDicts Success list \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := dda.ListDataDicts(testCtx, dictQuery)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListDataDicts Success list no limit\n", func() {
			dictQuery.Limit = -1

			sqlStr := fmt.Sprintf("SELECT f_dict_id, f_dict_name, f_dict_type, f_unique_key, "+
				"f_dimension, f_tags, f_comment, f_create_time, f_update_time "+
				"FROM %s ORDER BY f_update_time desc", DATA_DICT_TABLE_NAME)

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := dda.ListDataDicts(testCtx, dictQuery)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListDataDicts Failed dbQuery \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, err := dda.ListDataDicts(testCtx, dictQuery)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListDataDicts Failed Unmarshal \n", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFuncReturn(sonic.Unmarshal, expectedErr)
			defer patch.Reset()

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := dda.ListDataDicts(testCtx, dictQuery)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListDataDicts Failed ToSql \n", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			_, err := dda.ListDataDicts(testCtx, dictQuery)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataDictAccess_GetDictTotal(t *testing.T) {
	Convey("Test DictAccess GetDictTotal\n", t, func() {
		appSetting := &common.AppSetting{}
		dda, smock := MockNewDataDictAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT COUNT(f_dict_id) FROM %s", DATA_DICT_TABLE_NAME)
		rows := sqlmock.NewRows([]string{"count(f_dict_id)"}).AddRow(1)

		dictQuery := interfaces.DataDictQueryParams{
			NamePattern: "",
			PaginationQueryParameters: interfaces.PaginationQueryParameters{
				Limit:     interfaces.MAX_LIMIT,
				Offset:    interfaces.MIN_LIMIT,
				Sort:      interfaces.DATA_DICT_SORT["update_time"],
				Direction: interfaces.DESC_DIRECTION,
			},
		}

		Convey("GetDictTotal Success\n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			total, err := dda.GetDictTotal(testCtx, dictQuery)
			So(total, ShouldEqual, 1)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetDictTotal Failed  Query error\n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			total, err := dda.GetDictTotal(testCtx, dictQuery)
			So(total, ShouldEqual, 0)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataDictAccess_GetDataDict(t *testing.T) {
	Convey("Test DictAccess GetDataDict\n", t, func() {
		appSetting := &common.AppSetting{}
		dda, smock := MockNewDataDictAccess(appSetting)

		dictID := "1"

		sqlStr := fmt.Sprintf("SELECT f_dict_id, f_dict_name, f_dict_type, f_unique_key, f_dimension, "+
			"f_dict_store, f_tags, f_comment, f_create_time, f_update_time FROM %s "+
			"WHERE f_dict_id = ?", DATA_DICT_TABLE_NAME)

		rows := sqlmock.NewRows([]string{"f_dict_id", "f_dict_name", "f_dict_type", "f_unique_key",
			"f_dimension", "f_dict_store", "f_tags", "f_comment", "f_create_time", "f_update_time"}).
			AddRow("1", "qqww", "kv_dict", true,
				"{\"keys\":[{\"id\":\"item_key\",\"name\":\"key\"}],\"values\":[{\"id\":\"item_value\",\"name\":\"value\"}]}",
				"t_data_dict_item", "\"a\",\"b\",\"c\"", "comment222", 1638953462000, 1638953462000)

		Convey("GetDataDict Success \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := dda.GetDataDictByID(testCtx, dictID)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetDataDict Success no row \n", func() {
			expectedErr := sql.ErrNoRows
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, err := dda.GetDataDictByID(testCtx, dictID)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetDataDict Failed Unmarshal \n", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFuncReturn(sonic.Unmarshal, expectedErr)
			defer patch.Reset()

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := dda.GetDataDictByID(testCtx, dictID)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetDataDict Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, err := dda.GetDataDictByID(testCtx, dictID)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataDictAccess_CreateDataDict(t *testing.T) {
	Convey("Test DictAccess CreateDataDict", t, func() {
		appSetting := &common.AppSetting{}
		dda, smock := MockNewDataDictAccess(appSetting)

		sqlStr := fmt.Sprintf("INSERT INTO %s (f_dict_id,f_dict_name,f_dict_type,"+
			"f_unique_key,f_dimension,f_dict_store,f_tags,f_comment,f_create_time,f_update_time) "+
			"VALUES (?,?,?,?,?,?,?,?,?,?)", DATA_DICT_TABLE_NAME)

		dictInfo := interfaces.DataDict{
			DictID:     "1",
			DictName:   "test",
			DictType:   "kv_dict",
			UniqueKey:  true,
			Dimension:  interfaces.DATA_DICT_KV_DIMENSION,
			DictStore:  "t_data_dict_item",
			Tags:       []string{"a", "b", "c"},
			Comment:    "",
			CreateTime: testNow,
			UpdateTime: testNow,
		}

		Convey("CreateDataDict Success \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			tx, _ := dda.db.Begin()
			err := dda.CreateDataDict(testCtx, tx, dictInfo)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CreateDataDict marshal error\n", func() {
			expectedErr := errors.New("some error1")
			patch := ApplyFuncReturn(sonic.Marshal, []byte{}, expectedErr)
			defer patch.Reset()

			tx, _ := dda.db.Begin()
			err := dda.CreateDataDict(testCtx, tx, dictInfo)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CreateDataDict Exec error \n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("any error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := dda.db.Begin()
			err := dda.CreateDataDict(testCtx, tx, dictInfo)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataDictAccess_UpdateDataDict(t *testing.T) {
	Convey("Test DictAccess UpdateDataDict\n", t, func() {
		appSetting := &common.AppSetting{}
		dda, smock := MockNewDataDictAccess(appSetting)

		sqlStr := fmt.Sprintf("UPDATE %s SET f_comment = ?, f_dict_name = ?, f_dimension = ?, "+
			"f_tags = ?, f_update_time = ? WHERE f_dict_id = ?", DATA_DICT_TABLE_NAME)

		dict := interfaces.DataDict{
			DictID:     "1",
			DictName:   "test",
			DictType:   "kv_dict",
			UniqueKey:  true,
			Dimension:  interfaces.DATA_DICT_KV_DIMENSION,
			DictStore:  "t_data_dict_item",
			Tags:       []string{"a", "b", "c"},
			Comment:    "",
			UpdateTime: testNow,
		}

		Convey("UpdateDataDict Success \n", func() {
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			err := dda.UpdateDataDict(testCtx, dict)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateDataDict Failed Marshal \n", func() {
			expectedErr := errors.New("some error3")
			patch := ApplyFuncReturn(sonic.Marshal, []byte{}, expectedErr)
			defer patch.Reset()

			err := dda.UpdateDataDict(testCtx, dict)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateDataDict Failed UpdateSql \n", func() {
			expectedErr := errors.New("some error3")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			err := dda.UpdateDataDict(testCtx, dict)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateDataDict Failed RowsAffected \n", func() {
			expectedErr := errors.New("Get RowsAffected error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			err := dda.UpdateDataDict(testCtx, dict)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateDataDict Failed RowsAffected > 1 \n", func() {
			expectedErr := errors.New("UpdateDataDict update Dict RowsAffected more than 1")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 5))

			err := dda.UpdateDataDict(testCtx, dict)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataDictAccess_DeleteDataDict(t *testing.T) {
	Convey("Test DictAccess DeleteDataDict\n", t, func() {
		appSetting := &common.AppSetting{}
		dda, smock := MockNewDataDictAccess(appSetting)

		sqlStr := fmt.Sprintf("DELETE FROM %s WHERE f_dict_id = ?", DATA_DICT_TABLE_NAME)

		dictID := "1"

		Convey("DeleteDataDict Success \n", func() {
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			rowsAffected, err := dda.DeleteDataDict(testCtx, dictID)
			So(rowsAffected, ShouldEqual, 1)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteDataDict RowsAffected 0 \n", func() {
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 0))

			rowsAffected, err := dda.DeleteDataDict(testCtx, dictID)
			So(rowsAffected, ShouldEqual, 0)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteDataDict Failed dbExec \n", func() {
			expectedErr := errors.New("some error2")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, err := dda.DeleteDataDict(testCtx, dictID)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataDictAccess_CreateDimensionDictStore(t *testing.T) {
	Convey("Test DictAccess CreateDimensionDictStore\n", t, func() {
		appSetting := &common.AppSetting{}
		dda, smock := MockNewDataDictAccess(appSetting)

		sqlStr := "CREATE TABLE (f_item_id varchar(40) NOT NULL COMMENT '数据字典项id', " +
			"f_key_2132131321 TEXT NOT NULL COMMENT 'key111', " +
			"f_value_2132131321 TEXT NOT NULL COMMENT 'value111', " +
			"f_comment varchar(255) DEFAULT NULL COMMENT '多维度数据字典项说明')"

		dimension := interfaces.Dimension{
			Keys: []interfaces.DimensionItem{
				{
					ID:   "f_key_2132131321",
					Name: "key111",
				},
			},
			Values: []interfaces.DimensionItem{
				{
					ID:   "f_value_2132131321",
					Name: "value111",
				},
			},
			Comment: "",
		}

		Convey("CreateDimensionDictStore Success \n", func() {
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			err := dda.CreateDimensionDictStore(testCtx, "", dimension)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CreateDimensionDictStore Failed Exec \n", func() {
			expectedErr := errors.New("some error2")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			err := dda.CreateDimensionDictStore(testCtx, "", dimension)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataDictAccess_AddDimensionIndex(t *testing.T) {
	Convey("Test DictAccess AddDimensionIndex\n", t, func() {
		appSetting := &common.AppSetting{}
		dda, smock := MockNewDataDictAccess(appSetting)

		sqlStr := "ALTER TABLE ADD UNIQUE INDEX _uk_keys(f_key_2132131321)"

		dimension := interfaces.Dimension{
			Keys: []interfaces.DimensionItem{
				{
					ID:   "f_key_2132131321",
					Name: "key111",
				},
			},
			Values: []interfaces.DimensionItem{
				{
					ID:   "f_value_2132131321",
					Name: "value111",
				},
			},
			Comment: "",
		}

		os.Setenv("DB_TYPE", "MYSQL")

		Convey("AddDimensionIndex Success DM8\n", func() {
			os.Setenv("DB_TYPE", "DM8")

			sqlStr = "CREATE UNIQUE INDEX _uk_keys ON (f_key_2132131321)"
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			err := dda.AddDimensionIndex(testCtx, "", dimension.Keys)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("AddDimensionIndex Success MYSQL\n", func() {
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			err := dda.AddDimensionIndex(testCtx, "", dimension.Keys)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("AddDimensionIndex Failed ORACLE\n", func() {
			os.Setenv("DB_TYPE", "ORACLE")
			expectedErr := errors.New("unsupported database type")

			err := dda.AddDimensionIndex(testCtx, "", dimension.Keys)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("AddDimensionIndex Failed Exec \n", func() {
			expectedErr := errors.New("some error2")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			err := dda.AddDimensionIndex(testCtx, "", dimension.Keys)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataDictAccess_DropDimensionIndex(t *testing.T) {
	Convey("Test DictAccess DropDimensionIndex\n", t, func() {
		appSetting := &common.AppSetting{}
		dda, smock := MockNewDataDictAccess(appSetting)

		sqlStr := "ALTER TABLE DROP INDEX _uk_keys"

		os.Setenv("DB_TYPE", "MYSQL")

		Convey("DropDimensionIndex Success DM8\n", func() {
			os.Setenv("DB_TYPE", "DM8")

			sqlStr = "DROP INDEX _uk_keys"
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			err := dda.DropDimensionIndex(testCtx, "")
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DropDimensionIndex Success MYSQL\n", func() {
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			err := dda.DropDimensionIndex(testCtx, "")
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DropDimensionIndex Failed ORACLO\n", func() {
			os.Setenv("DB_TYPE", "ORACLO")

			expectedErr := errors.New("unsupported database type")
			err := dda.DropDimensionIndex(testCtx, "")
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DropDimensionIndex Failed Exec \n", func() {
			expectedErr := errors.New("some error2")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			err := dda.DropDimensionIndex(testCtx, "")
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataDictAccess_UpdateDictUpdateTime(t *testing.T) {
	Convey("Test DictAccess UpdateDictUpdateTime\n", t, func() {
		appSetting := &common.AppSetting{}
		dda, smock := MockNewDataDictAccess(appSetting)

		sqlStr := fmt.Sprintf("UPDATE %s SET f_update_time = ? WHERE f_dict_id = ?", DATA_DICT_TABLE_NAME)

		dict := interfaces.DataDict{
			DictID:     "1",
			UpdateTime: testNow, //设置更新时间
		}

		Convey("UpdateDictUpdateTime Success \n", func() {
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			err := dda.UpdateDictUpdateTime(testCtx, dict.DictID, testNow)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateDictUpdateTime Failed UpdateSql \n", func() {
			expectedErr := errors.New("some error3")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			err := dda.UpdateDictUpdateTime(testCtx, dict.DictID, testNow)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}
func Test_DataDictAccess_CheckDictExistByName(t *testing.T) {
	Convey("Test DictAccess CheckDictExistByName\n", t, func() {
		appSetting := &common.AppSetting{}
		dda, smock := MockNewDataDictAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_dict_id, f_dict_name FROM %s WHERE f_dict_name = ?", DATA_DICT_TABLE_NAME)

		dictName := "1"

		rows := sqlmock.NewRows([]string{"f_dict_id", "f_dict_name"}).AddRow("1", dictName)

		Convey("CheckDictExistByName Success \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			exist, err := dda.CheckDictExistByName(testCtx, dictName)
			So(err, ShouldBeNil)
			So(exist, ShouldBeTrue)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CheckDictExistByName Success no row \n", func() {
			expectedErr := sql.ErrNoRows
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			exist, err := dda.CheckDictExistByName(testCtx, dictName)
			So(err, ShouldBeNil)
			So(exist, ShouldBeFalse)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CheckDictExistByName Failed \n", func() {
			expectedErr := errors.New("some error7")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			exist, err := dda.CheckDictExistByName(testCtx, dictName)
			So(err, ShouldResemble, expectedErr)
			So(exist, ShouldBeFalse)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}
