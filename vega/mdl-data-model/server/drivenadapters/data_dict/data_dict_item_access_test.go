// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_dict

import (
	"database/sql"
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

func MockNewDataDictItemAccess(appSetting *common.AppSetting) (*dataDictItemAccess, sqlmock.Sqlmock) {
	db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	ddia := &dataDictItemAccess{
		appSetting: appSetting,
		db:         db,
	}
	return ddia, smock
}

func Test_DataDictItemAccess_ListDataDictItems(t *testing.T) {
	Convey("Test DictItemAccess ListDataDictItems\n", t, func() {
		appSetting := &common.AppSetting{}
		ddia, smock := MockNewDataDictItemAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_item_id as \"id\", f_item_key as \"key\", "+
			"f_item_value as \"value\", f_comment as \"comment\" FROM %s WHERE f_dict_id = ? "+
			"ORDER BY f_item_id desc LIMIT 1000 OFFSET 0", DATA_DICT_ITEM_TABLE_NAME)

		rows := sqlmock.NewRows([]string{"f_dict_id", "f_item_id", "f_item_key", "f_item_value",
			"f_comment"}).AddRow("1", "1", "item_key", "item_value", "comment222")

		dictItemQuery := interfaces.DataDictItemQueryParams{
			Patterns: []interfaces.DataDictItemQueryPattern{
				{
					QueryField:   "key",
					QueryPattern: "",
				},
				{
					QueryField:   "value",
					QueryPattern: "",
				},
			},
			PaginationQueryParameters: interfaces.PaginationQueryParameters{
				Limit:     interfaces.MAX_LIMIT,
				Offset:    interfaces.MIN_OFFSET,
				Sort:      interfaces.DATA_DICT_ITEM_SORT["id"],
				Direction: interfaces.DESC_DIRECTION,
			},
		}
		dictInfo := interfaces.DataDict{
			DictID:    "1",
			DictName:  "test",
			DictType:  "kv_dict",
			UniqueKey: true,
			Tags:      []string{"a", "b", "c"},
			Dimension: interfaces.DATA_DICT_KV_DIMENSION,
			DictStore: "t_data_dict_item",
			Comment:   "",
		}

		Convey("ListDataDictItems Success list \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := ddia.ListDataDictItems(testCtx, dictInfo, dictItemQuery)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListDataDictItems Success dimesion list \n", func() {
			dictInfo.DictStore = "t_dim4864146446464"

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := ddia.ListDataDictItems(testCtx, dictInfo, dictItemQuery)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListDataDictItems Success list no limit\n", func() {
			dictItemQuery.Limit = -1

			sqlStr := fmt.Sprintf("SELECT f_item_id as \"id\", f_item_key as \"key\", "+
				"f_item_value as \"value\", f_comment as \"comment\" FROM %s "+
				"WHERE f_dict_id = ? ORDER BY f_item_id desc", DATA_DICT_ITEM_TABLE_NAME)

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := ddia.ListDataDictItems(testCtx, dictInfo, dictItemQuery)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListDataDictItems Failed ToSql\n", func() {
			expectedErr := errors.New("some ToSql")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			_, err := ddia.ListDataDictItems(testCtx, dictInfo, dictItemQuery)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListDataDictItems Failed dbQuery \n", func() {
			expectedErr := errors.New("some error5")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, err := ddia.ListDataDictItems(testCtx, dictInfo, dictItemQuery)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}
func Test_DataDictItemAccess_GetDictItemTotal(t *testing.T) {
	Convey("Test DictItemAccess GetDictItemTotal\n", t, func() {
		appSetting := &common.AppSetting{}
		ddia, smock := MockNewDataDictItemAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT COUNT(f_item_id) FROM %s WHERE f_dict_id = ?", DATA_DICT_ITEM_TABLE_NAME)

		rows := sqlmock.NewRows([]string{"count(f_item_id)"}).AddRow(1)

		dictItemQuery := interfaces.DataDictItemQueryParams{
			Patterns: []interfaces.DataDictItemQueryPattern{
				{
					QueryField:   "key",
					QueryPattern: "",
				},
				{
					QueryField:   "value",
					QueryPattern: "",
				},
			},
			PaginationQueryParameters: interfaces.PaginationQueryParameters{
				Limit:     interfaces.MAX_LIMIT,
				Offset:    interfaces.MIN_OFFSET,
				Sort:      interfaces.DATA_DICT_ITEM_SORT["id"],
				Direction: interfaces.DESC_DIRECTION,
			},
		}
		dictInfo := interfaces.DataDict{
			DictID:    "1",
			DictName:  "test",
			DictType:  "kv_dict",
			UniqueKey: true,
			Tags:      []string{"a", "b", "c"},
			Dimension: interfaces.DATA_DICT_KV_DIMENSION,
			DictStore: "t_data_dict_item",
			Comment:   "",
		}

		Convey("GetDictItemTotal Success\n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			total, err := ddia.GetDictItemTotal(testCtx, dictInfo, dictItemQuery)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 1)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetDictItemTotal dimension Success\n", func() {
			dictInfo.DictStore = "t_dim4864146446464"

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			total, err := ddia.GetDictItemTotal(testCtx, dictInfo, dictItemQuery)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 1)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetDictItemTotal Failed ToSql\n", func() {
			expectedErr := errors.New("some ToSql")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			_, err := ddia.GetDictItemTotal(testCtx, dictInfo, dictItemQuery)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetDictItemTotal Failed  Query error\n", func() {
			expectedErr := errors.New("some error9")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			total, err := ddia.GetDictItemTotal(testCtx, dictInfo, dictItemQuery)
			So(err, ShouldResemble, expectedErr)
			So(total, ShouldEqual, 0)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataDictItemAccess_GetKVDictItems(t *testing.T) {
	Convey("Test ddia GetKVDictItems\n", t, func() {
		appSetting := &common.AppSetting{}
		ddia, smock := MockNewDataDictItemAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_item_id as \"id\", f_item_key as \"key\", f_item_value as \"value\", "+
			"f_comment as \"comment\" FROM %s WHERE f_dict_id = ?", DATA_DICT_ITEM_TABLE_NAME)

		dictID := "1"

		rows := sqlmock.NewRows([]string{"f_item_id", "f_item_key", "f_item_value", "f_comment"}).
			AddRow("id", "key", "value", "comment").
			AddRow("id0", "key0", "value0", "comment0")

		Convey("GetKVDictItems Success \n", func() {
			items := []map[string]string{
				{"comment": "comment", "key": "key", "value": "value", "item_id": "id"},
				{"comment": "comment0", "key": "key0", "value": "value0", "item_id": "id0"},
			}
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			getItems, err := ddia.GetKVDictItems(testCtx, dictID)
			So(err, ShouldBeNil)
			So(getItems, ShouldResemble, items)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetKVDictItems Tosql \n", func() {
			expectedErr := errors.New("some ToSql")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			_, err := ddia.GetKVDictItems(testCtx, dictID)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetKVDictItems Failed \n", func() {
			expectedErr := errors.New("some error6")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, err := ddia.GetKVDictItems(testCtx, dictID)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetKVDictItems Marshal Failed \n", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFuncReturn(sonic.Marshal, []byte{}, expectedErr)
			defer patch.Reset()

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := ddia.GetKVDictItems(testCtx, dictID)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataDictItemAccess_GetDimensionDictItems(t *testing.T) {
	Convey("Test ddia GetDimensionDictItems\n", t, func() {
		appSetting := &common.AppSetting{}
		ddia, smock := MockNewDataDictItemAccess(appSetting)

		dictStore := "t_dim254243243"
		sqlStr := fmt.Sprintf("SELECT f_key_2132131321 as \"key111\", "+
			"f_value_2132131321 as \"value111\", f_comment as \"comment\" FROM %s", dictStore)

		rows := sqlmock.NewRows([]string{"f_item_key", "f_item_value", "f_comment"}).
			AddRow("key", "value", "comment").AddRow("key0", "value0", "comment0")

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

		Convey("GetDimensionDictItems Success \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := ddia.GetDimensionDictItems(testCtx, dictStore, dimension)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetDimensionDictItems Tosql \n", func() {
			expectedErr := errors.New("some ToSql")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			_, err := ddia.GetDimensionDictItems(testCtx, dictStore, dimension)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetDimensionDictItems Failed \n", func() {
			expectedErr := errors.New("some error6")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, err := ddia.GetDimensionDictItems(testCtx, dictStore, dimension)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataDictItemAccess_AddDimensionColumn(t *testing.T) {
	Convey("Test ddia AddDimensionColumn\n", t, func() {
		appSetting := &common.AppSetting{}
		ddia, smock := MockNewDataDictItemAccess(appSetting)

		dictStore := "t_dim254243243"
		sqlStr := fmt.Sprintf("ALTER TABLE %s ADD COLUMN f_2332 varchar(1000) NOT NULL", dictStore)

		new := interfaces.DimensionItem{
			ID:   "f_2332",
			Name: "key34432",
		}

		Convey("AddDimensionColumn Success \n", func() {
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			err := ddia.AddDimensionColumn(testCtx, dictStore, new)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("AddDimensionColumn Failed \n", func() {
			expectedErr := errors.New("some error6")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			err := ddia.AddDimensionColumn(testCtx, dictStore, new)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataDictItemAccess_DropDimensionColumn(t *testing.T) {
	Convey("Test ddia DropDimensionColumn\n", t, func() {
		appSetting := &common.AppSetting{}
		ddia, smock := MockNewDataDictItemAccess(appSetting)

		dictStore := "t_dim254243243"
		sqlStr := fmt.Sprintf("ALTER TABLE %s DROP COLUMN f_2332", dictStore)

		new := interfaces.DimensionItem{
			ID:   "f_2332",
			Name: "key34432",
		}

		Convey("DropDimensionColumn Success \n", func() {
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			err := ddia.DropDimensionColumn(testCtx, dictStore, new)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DropDimensionColumn Failed \n", func() {
			expectedErr := errors.New("some error6")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			err := ddia.DropDimensionColumn(testCtx, dictStore, new)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataDictItemAccess_DeleteDimensionTable(t *testing.T) {
	Convey("Test ddia DeleteDimensionTable\n", t, func() {
		appSetting := &common.AppSetting{}
		ddia, smock := MockNewDataDictItemAccess(appSetting)

		dictStore := "t_dim254243243"
		sqlStr := fmt.Sprintf("DROP TABLE %s", dictStore)

		Convey("DeleteDimensionTable Success \n", func() {
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			err := ddia.DeleteDimensionTable(testCtx, dictStore)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteDimensionTable Failed \n", func() {
			expectedErr := errors.New("some error6")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			err := ddia.DeleteDimensionTable(testCtx, dictStore)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataDictItemAccess_DeleteDataDictItems(t *testing.T) {
	Convey("Test DictAccess DeleteDataDictItems\n", t, func() {
		appSetting := &common.AppSetting{}
		ddia, smock := MockNewDataDictItemAccess(appSetting)

		sqlStr := fmt.Sprintf("DELETE FROM %s WHERE f_dict_id = ?", DATA_DICT_ITEM_TABLE_NAME)

		dictID := "1"

		Convey("DeleteDataDictItems Success \n", func() {
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			err := ddia.DeleteDataDictItems(testCtx, dictID)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteDataDictItems Tosql \n", func() {
			expectedErr := errors.New("some ToSql")
			patch := ApplyMethodReturn(sq.DeleteBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			err := ddia.DeleteDataDictItems(testCtx, dictID)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteDataDictItems Failed dbExec \n", func() {
			expectedErr := errors.New("some error2")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			err := ddia.DeleteDataDictItems(testCtx, dictID)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataDictItemAccess_CheckDictItemByKey(t *testing.T) {
	Convey("Test ddia CountDictItemByKey\n", t, func() {
		appSetting := &common.AppSetting{}
		ddia, smock := MockNewDataDictItemAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT count(f_item_id) FROM %s "+
			"WHERE f_dict_id = ? AND f_item_key = ?", DATA_DICT_ITEM_TABLE_NAME)

		dictID := "1"

		rows := sqlmock.NewRows([]string{"item_id"}).AddRow("1")

		keys := []interfaces.DimensionItem{
			{
				ID:    "f_key_432432422",
				Name:  "key00",
				Value: "testkey",
			},
		}

		Convey("CountDictItemByKey Success \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			exist, err := ddia.CountDictItemByKey(testCtx, dictID, interfaces.DATA_DICT_STORE_DEFAULT, keys)
			So(err, ShouldBeNil)
			So(exist, ShouldEqual, 1)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CountDictItemByKey Failed Tosql \n", func() {
			expectedErr := errors.New("some ToSql")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			exist, err := ddia.CountDictItemByKey(testCtx, dictID, interfaces.DATA_DICT_STORE_DEFAULT, keys)
			So(err, ShouldResemble, expectedErr)
			So(exist, ShouldEqual, 0)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CountDictItemByKey Failed \n", func() {
			expectedErr := errors.New("some error7")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, err := ddia.CountDictItemByKey(testCtx, dictID, interfaces.DATA_DICT_STORE_DEFAULT, keys)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataDictItemAccess_CreateDataDictItem(t *testing.T) {
	Convey("Test ddia CreateDataDictItem", t, func() {
		appSetting := &common.AppSetting{}
		ddia, smock := MockNewDataDictItemAccess(appSetting)

		sqlStr := fmt.Sprintf("INSERT INTO %s (f_dict_id,f_item_id,f_item_key,"+
			"f_item_value,f_comment) VALUES (?,?,?,?,?)", DATA_DICT_ITEM_TABLE_NAME)

		dictID := "1"
		itemID := "1"

		dimension := interfaces.Dimension{
			Keys: []interfaces.DimensionItem{
				{
					ID:    "f_key_2132131321",
					Name:  "key111",
					Value: "testkey",
				},
			},
			Values: []interfaces.DimensionItem{
				{
					ID:    "f_value_2132131321",
					Name:  "value111",
					Value: "testvalue",
				},
			},
			Comment: "testcomment",
		}
		dictStore := "t_dim254243243"

		Convey("CreateDataDictItem Success \n", func() {
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			err := ddia.CreateDataDictItem(testCtx, dictID, itemID, interfaces.DATA_DICT_STORE_DEFAULT, dimension)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CreateDataDictItem Tosql error \n", func() {
			expectedErr := errors.New("some ToSql")
			patch := ApplyMethodReturn(sq.InsertBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			err := ddia.CreateDataDictItem(testCtx, dictID, itemID, dictStore, dimension)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CreateDataDictItem  Exec sql error\n", func() {
			expectedErr := errors.New("some error1")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			err := ddia.CreateDataDictItem(testCtx, dictID, itemID, interfaces.DATA_DICT_STORE_DEFAULT, dimension)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataDictItemAccess_GetDictItemByItemID(t *testing.T) {
	Convey("Test DictItemAccess GetDictItemByItemID\n", t, func() {
		appSetting := &common.AppSetting{}
		ddia, smock := MockNewDataDictItemAccess(appSetting)

		dictStore := "t_dim254243243"
		sqlStr := fmt.Sprintf("SELECT * FROM %s WHERE f_item_id = ?", dictStore)

		itemID := "1"

		rows := sqlmock.NewRows([]string{"f_dict_id", "f_item_id", "f_item_key",
			"f_item_value", "f_comment"}).AddRow("1", itemID, "test key", "test value", "")

		Convey("GetDictItemByItemID Success \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := ddia.GetDictItemByItemID(testCtx, dictStore, itemID)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetDictItemByItemID Failed Tosql \n", func() {
			expectedErr := errors.New("some ToSql")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			_, err := ddia.GetDictItemByItemID(testCtx, interfaces.DATA_DICT_STORE_DEFAULT, itemID)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetDictItemByItemID Success no row \n", func() {
			expectedErr := sql.ErrNoRows
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, err := ddia.GetDictItemByItemID(testCtx, dictStore, itemID)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetDictItemByItemID Failed \n", func() {
			expectedErr := errors.New("some error7")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, err := ddia.GetDictItemByItemID(testCtx, dictStore, itemID)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataDictItemAccess_UpdateDataDictItem(t *testing.T) {
	Convey("Test ddia UpdateDataDictItem\n", t, func() {
		appSetting := &common.AppSetting{}
		ddia, smock := MockNewDataDictItemAccess(appSetting)

		dictStore := "t_dim254243243"
		sqlStr := fmt.Sprintf("UPDATE %s SET f_key2132131321 = ?, "+
			"f_value2132131321 = ?, f_comment = ? WHERE f_item_id = ?", dictStore)

		dictID := "1"
		itemID := "1"

		dimension := interfaces.Dimension{
			Keys: []interfaces.DimensionItem{
				{
					ID:    "f_key2132131321",
					Name:  "key111",
					Value: "testkey",
				},
			},
			Values: []interfaces.DimensionItem{
				{
					ID:    "f_value2132131321",
					Name:  "value111",
					Value: "testvalue",
				},
			},
			Comment: "testcomment",
		}

		Convey("UpdateDataDictItem Success \n", func() {
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
			err := ddia.UpdateDataDictItem(testCtx, dictID, itemID, dictStore, dimension)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateDataDictItem failed ToSql \n", func() {
			expectedErr := errors.New("some ToSql")
			patch := ApplyMethodReturn(sq.UpdateBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			err := ddia.UpdateDataDictItem(testCtx, dictID, itemID, interfaces.DATA_DICT_STORE_DEFAULT, dimension)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateDataDictItem Failed UpdateSql \n", func() {
			expectedErr := errors.New("some error3")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			err := ddia.UpdateDataDictItem(testCtx, dictID, itemID, dictStore, dimension)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataDictItemAccess_DeleteDataDictItem(t *testing.T) {
	Convey("Test ddia DeleteDataDictItem\n", t, func() {
		appSetting := &common.AppSetting{}
		ddia, smock := MockNewDataDictItemAccess(appSetting)

		dictStore := "t_dim254243243"
		sqlStr := fmt.Sprintf("DELETE FROM %s WHERE f_item_id = ?", dictStore)

		dictID := "1"
		itemID := "1"

		Convey("DeleteDataDictItem Success \n", func() {
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			rowsAffected, err := ddia.DeleteDataDictItem(testCtx, dictID, itemID, dictStore)
			So(err, ShouldBeNil)
			So(rowsAffected, ShouldEqual, 1)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteDataDictItem failed prepare \n", func() {
			expectedErr := errors.New("some ToSql")
			patch := ApplyMethodReturn(sq.DeleteBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			_, err := ddia.DeleteDataDictItem(testCtx, dictID, itemID, dictStore)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteDataDictItem RowsAffected 0 \n", func() {
			sqlStr = fmt.Sprintf("DELETE FROM %s WHERE f_dict_id = ? "+
				"AND f_item_id = ?", interfaces.DATA_DICT_STORE_DEFAULT)

			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 0))
			rowsAffected, err := ddia.DeleteDataDictItem(testCtx, dictID, itemID, interfaces.DATA_DICT_STORE_DEFAULT)
			So(rowsAffected, ShouldEqual, 0)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteDataDictItem Failed dbExec \n", func() {
			sqlStr = fmt.Sprintf("DELETE FROM %s WHERE f_dict_id = ? "+
				"AND f_item_id = ?", interfaces.DATA_DICT_STORE_DEFAULT)

			expectedErr := errors.New("some error2")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)
			_, err := ddia.DeleteDataDictItem(testCtx, dictID, itemID, interfaces.DATA_DICT_STORE_DEFAULT)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataDictItemAccess_CreateKVDictItems(t *testing.T) {
	Convey("Test ddia CreateKVDictItems", t, func() {
		appSetting := &common.AppSetting{}
		ddia, smock := MockNewDataDictItemAccess(appSetting)

		sqlStr := fmt.Sprintf("INSERT INTO %s (f_dict_id,f_item_id,f_item_key,f_item_value,f_comment) "+
			"VALUES (?,?,?,?,?)", DATA_DICT_ITEM_TABLE_NAME)

		dictID := "1"

		dictItemInfo := []interfaces.KvDictItem{
			{
				DictID:  "1",
				ItemID:  "1",
				Key:     "testkey",
				Value:   "testvalue",
				Comment: "",
			}}

		Convey("CreateDataDictItems Success \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 2))

			tx, _ := ddia.db.Begin()
			err := ddia.CreateKVDictItems(testCtx, tx, dictID, dictItemInfo)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CreateDataDictItems Tosql error \n", func() {
			expectedErr := errors.New("some ToSql")
			patch := ApplyMethodReturn(sq.InsertBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			smock.ExpectBegin()

			tx, _ := ddia.db.Begin()
			err := ddia.CreateKVDictItems(testCtx, tx, dictID, dictItemInfo)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CreateDataDictItems  Exec sql error\n", func() {
			expectedErr := errors.New("some error1")
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := ddia.db.Begin()
			err := ddia.CreateKVDictItems(testCtx, tx, dictID, dictItemInfo)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataDictItemAccess_CreateDimensionDictItems(t *testing.T) {
	Convey("Test ddia CreateDimensionDictItems", t, func() {
		appSetting := &common.AppSetting{}
		ddia, smock := MockNewDataDictItemAccess(appSetting)

		dictStore := "t_dim254243243"
		sqlStr := fmt.Sprintf("INSERT INTO %s (f_item_id,f_key2132131321,f_value2132131321,f_comment) "+
			"VALUES (?,?,?,?)", dictStore)

		dictID := "1"

		dimensions := []interfaces.Dimension{
			{
				ItemID: "1",
				Keys: []interfaces.DimensionItem{
					{
						ID:    "f_key2132131321",
						Name:  "key111",
						Value: "testkey",
					},
				},
				Values: []interfaces.DimensionItem{
					{
						ID:    "f_value2132131321",
						Name:  "value111",
						Value: "testvalue",
					},
				},
				Comment: "testcomment"},
		}

		Convey("CreateDimensionDictItems Success \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 2))

			tx, _ := ddia.db.Begin()
			err := ddia.CreateDimensionDictItems(testCtx, tx, dictID, dictStore, dimensions)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CreateDimensionDictItems Tosql error \n", func() {
			expectedErr := errors.New("some ToSql")
			patch := ApplyMethodReturn(sq.InsertBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			smock.ExpectBegin()

			tx, _ := ddia.db.Begin()
			err := ddia.CreateDimensionDictItems(testCtx, tx, dictID, dictStore, dimensions)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CreateDimensionDictItems  Exec sql error\n", func() {
			expectedErr := errors.New("some error1")
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := ddia.db.Begin()
			err := ddia.CreateDimensionDictItems(testCtx, tx, dictID, dictStore, dimensions)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}
