// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_view

import (
	"context"
	// 	"database/sql"
	// 	"errors"
	// 	"fmt"
	// 	"testing"
	"time"

	// 	// libCommon "github.com/kweaver-ai/kweaver-go-lib/common"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	// 	"github.com/DATA-DOG/go-sqlmock"
	// 	sq "github.com/Masterminds/squirrel"
	// 	. "github.com/agiledragon/gomonkey/v2"
	// 	"github.com/bytedance/sonic"
	// 	. "github.com/smartystreets/goconvey/convey"
	// "data-model/common"
	"data-model/interfaces"
)

var (
	testCtx = context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage)
	testNow = time.Now().UnixMilli()
)

var (
	// 	// dataSource = map[string]any{
	// 	// 	"type": "index_base",
	// 	// 	"index_base": []interfaces.SimpleIndexBase{
	// 	// 		{BaseType: "x"},
	// 	// 		{BaseType: "y"},
	// 	// 		{BaseType: "z"},
	// 	// 	},
	// 	// }

	// 	TestExcel = &interfaces.ExcelConfig{
	// 		Sheet:     "sheet1",
	// 		StartCell: "A1",
	// 		EndCell:   "B2",
	// 	}

	// 	TestFields = []interfaces.ViewField{
	// 		{Name: "错误ID", Type: "text", Comment: ""},
	// 		{Name: "语言", Type: "text", Comment: ""},
	// 	}

	// 	// testDataScope = &interfaces.DataScope{}

	// 	// condCfg = &interfaces.CondCfg{
	// 	// 	Name:      "签名",
	// 	// 	Operation: "==",
	// 	// 	ValueOptCfg: interfaces.ValueOptCfg{
	// 	// 		Value:     "kk",
	// 	// 		ValueFrom: dcond.ValueFrom_Const,
	// 	// 	},
	// 	// }

	// 	// excelBytes, _     = sonic.Marshal(TestExcel)
	// 	// fieldsBytes, _    = sonic.Marshal(TestFields)
	// 	// dataScopeBytes, _ = sonic.Marshal(testDataScope)
	viewUpdateTime = int64(1729649493334)
	// 	// viewTags          = []string{"a", "b", "c", "d", "e"}
	// 	// viewTagsStr       = libCommon.TagSlice2TagString(viewTags)

	views = []*interfaces.DataView{
		{
			SimpleDataView: interfaces.SimpleDataView{
				ViewID:     "1",
				ViewName:   "slslsl",
				GroupID:    "linux",
				GroupName:  "linux",
				Comment:    "a comment",
				UpdateTime: viewUpdateTime,
			},
		},
		{
			SimpleDataView: interfaces.SimpleDataView{
				ViewID:     "1",
				ViewName:   "slslsl",
				GroupID:    "",
				GroupName:  "",
				Comment:    "a comment",
				UpdateTime: viewUpdateTime,
			},
		},
	}
)

// func MockNewDataViewAccess(appSetting *common.AppSetting) (*dataViewAccess, sqlmock.Sqlmock) {
// 	db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
// 	dva := &dataViewAccess{
// 		appSetting: appSetting,
// 		db:         db,
// 	}

// 	return dva, smock
// }

// func Test_DataViewAccess_CreateDataViews(t *testing.T) {
// 	Convey("Test CreateDataViews", t, func() {
// 		appSetting := &common.AppSetting{}
// 		dva, smock := MockNewDataViewAccess(appSetting)

// 		sqlStr := fmt.Sprintf("INSERT INTO %s (f_view_id,f_view_name,f_technical_name,f_group_id,f_type,f_query_type,f_builtin,f_tags,"+
// 			"f_comment,f_data_source_type,f_data_source_id,f_file_name,f_excel_config,f_data_scope,f_fields,f_status,f_metadata_form_id,f_primary_keys,f_sql,"+
// 			"f_create_time,f_update_time,f_creator,f_updater) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", DATA_VIEW_TABLE_NAME)

// 		Convey("Create failed, caused by datasource marshal error", func() {
// 			expectedErr := errors.New("some error")
// 			patch := ApplyFuncReturn(sonic.Marshal, []byte{}, expectedErr)
// 			defer patch.Reset()

// 			smock.ExpectBegin()

// 			tx, _ := dva.db.Begin()
// 			err := dva.CreateDataViews(testCtx, tx, views)
// 			So(err, ShouldResemble, expectedErr)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Create failed, caused by fields marshal error", func() {
// 			expectedErr := errors.New("some error")
// 			patch := ApplyFunc(sonic.Marshal,
// 				func(v any) ([]byte, error) {
// 					if _, ok := v.([]*interfaces.ViewField); ok {
// 						return []byte{}, expectedErr
// 					}
// 					return []byte{}, nil
// 				},
// 			)
// 			defer patch.Reset()

// 			smock.ExpectBegin()

// 			tx, _ := dva.db.Begin()
// 			err := dva.CreateDataViews(testCtx, tx, views)
// 			So(err, ShouldResemble, expectedErr)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Create failed, caused by condition marshal error", func() {
// 			expectedErr := errors.New("some error")
// 			patch := ApplyFunc(sonic.Marshal,
// 				func(v any) ([]byte, error) {
// 					if _, ok := v.(map[string]any); ok {
// 						return []byte{}, nil
// 					}
// 					if _, ok := v.([]*interfaces.ViewField); ok {
// 						return []byte{}, nil
// 					}

// 					return []byte{}, expectedErr
// 				},
// 			)
// 			defer patch.Reset()

// 			smock.ExpectBegin()

// 			tx, _ := dva.db.Begin()
// 			err := dva.CreateDataViews(testCtx, tx, views)
// 			So(err, ShouldResemble, expectedErr)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Create failed, caused by the error from squirrel func ToSql", func() {
// 			expectedErr := errors.New("some error")
// 			patch := ApplyMethodReturn(sq.InsertBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
// 			defer patch.Reset()

// 			smock.ExpectBegin()

// 			tx, _ := dva.db.Begin()
// 			err := dva.CreateDataViews(testCtx, tx, views)
// 			So(err, ShouldResemble, expectedErr)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Create failed, caused by exec sql error", func() {
// 			expectedErr := errors.New("some error")
// 			smock.ExpectBegin()
// 			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

// 			tx, _ := dva.db.Begin()
// 			err := dva.CreateDataViews(testCtx, tx, views)
// 			So(err, ShouldResemble, expectedErr)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Create succeed", func() {
// 			// WithArgs()中不传参数, 应该代表任意参数下都会被mock.
// 			smock.ExpectBegin()
// 			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

// 			tx, _ := dva.db.Begin()
// 			err := dva.CreateDataViews(testCtx, tx, views)
// 			So(err, ShouldBeNil)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})
// 	})
// }

// func Test_DataViewAccess_DeleteDataViews(t *testing.T) {
// 	Convey("Test DeleteDataViews", t, func() {
// 		appSetting := &common.AppSetting{}
// 		dva, smock := MockNewDataViewAccess(appSetting)

// 		sqlStr := fmt.Sprintf("DELETE FROM %s WHERE f_view_id IN (?,?)", DATA_VIEW_TABLE_NAME)

// 		Convey("Delete failed, caused by the error from squirrel func ToSql", func() {
// 			expectedErr := errors.New("some error")
// 			patch := ApplyMethodReturn(sq.DeleteBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
// 			defer patch.Reset()

// 			smock.ExpectBegin()

// 			tx, _ := dva.db.Begin()
// 			err := dva.DeleteDataViews(testCtx, tx, []string{"1", "2"})
// 			So(err, ShouldResemble, expectedErr)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Delete failed, caused by exec sql error", func() {
// 			expectedErr := errors.New("some error")
// 			smock.ExpectBegin()
// 			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

// 			tx, _ := dva.db.Begin()
// 			err := dva.DeleteDataViews(testCtx, tx, []string{"1", "2"})
// 			So(err, ShouldResemble, expectedErr)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Delete succeed", func() {
// 			smock.ExpectBegin()
// 			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 2))

// 			tx, _ := dva.db.Begin()
// 			err := dva.DeleteDataViews(testCtx, tx, []string{"1", "2"})
// 			So(err, ShouldBeNil)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})
// 	})
// }

// func Test_DataViewAccess_UpdateDataView(t *testing.T) {
// 	Convey("Test UpdateDataView", t, func() {
// 		appSetting := &common.AppSetting{}
// 		dva, smock := MockNewDataViewAccess(appSetting)

// 		view := views[0]

// 		sqlStr := fmt.Sprintf("UPDATE %s SET f_comment = ?, f_data_source = ?, f_field_scope = ?, "+
// 			"f_fields = ?, f_filters = ?, f_group_id = ?, f_loggroup_filters = ?, f_tags = ?, "+
// 			"f_update_time = ?, f_view_name = ? WHERE f_view_id = ?", DATA_VIEW_TABLE_NAME)

// 		Convey("Update failed, caused by datasource marshal error", func() {
// 			expectedErr := errors.New("some error")
// 			patch := ApplyFuncReturn(sonic.Marshal, []byte{}, expectedErr)
// 			defer patch.Reset()

// 			smock.ExpectBegin()

// 			tx, _ := dva.db.Begin()
// 			err := dva.UpdateDataView(testCtx, tx, view)
// 			So(err, ShouldResemble, expectedErr)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Update failed, caused by fields marshal error", func() {
// 			expectedErr := errors.New("some error")
// 			patch := ApplyFunc(sonic.Marshal,
// 				func(v any) ([]byte, error) {
// 					if _, ok := v.([]*interfaces.ViewField); ok {
// 						return []byte{}, expectedErr
// 					}
// 					return []byte{}, nil
// 				},
// 			)
// 			defer patch.Reset()

// 			smock.ExpectBegin()

// 			tx, _ := dva.db.Begin()
// 			err := dva.UpdateDataView(testCtx, tx, view)
// 			So(err, ShouldResemble, expectedErr)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Update failed, caused by condition marshal error", func() {
// 			expectedErr := errors.New("some error")
// 			patch := ApplyFunc(sonic.Marshal,
// 				func(v any) ([]byte, error) {
// 					if _, ok := v.(map[string]any); ok {
// 						return []byte{}, nil
// 					}
// 					if _, ok := v.([]*interfaces.ViewField); ok {
// 						return []byte{}, nil
// 					}
// 					return []byte{}, expectedErr
// 				},
// 			)
// 			defer patch.Reset()

// 			smock.ExpectBegin()

// 			tx, _ := dva.db.Begin()
// 			err := dva.UpdateDataView(testCtx, tx, view)
// 			So(err, ShouldResemble, expectedErr)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Update failed, caused by the error from squirrel func ToSql", func() {
// 			expectedErr := errors.New("some error")
// 			patch := ApplyMethodReturn(sq.UpdateBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
// 			defer patch.Reset()

// 			smock.ExpectBegin()

// 			tx, _ := dva.db.Begin()
// 			err := dva.UpdateDataView(testCtx, tx, view)
// 			So(err, ShouldResemble, expectedErr)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Update failed, caused by exec sql error", func() {
// 			expectedErr := errors.New("some error")
// 			smock.ExpectBegin()
// 			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

// 			tx, _ := dva.db.Begin()
// 			err := dva.UpdateDataView(testCtx, tx, view)
// 			So(err, ShouldResemble, expectedErr)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Update succeed", func() {
// 			smock.ExpectBegin()
// 			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

// 			tx, _ := dva.db.Begin()
// 			err := dva.UpdateDataView(testCtx, tx, view)
// 			So(err, ShouldBeNil)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})
// 	})
// }

// func Test_DataViewAccess_GetDataViews(t *testing.T) {
// 	Convey("Test GetDataViews", t, func() {
// 		appSetting := &common.AppSetting{}
// 		dva, smock := MockNewDataViewAccess(appSetting)

// 		view := views[0]
// 		var tagsStr, primaryKeysStr string
// 		var excelConfigBytes, dataScopeBytes, fieldsBytes []byte

// 		rows := sqlmock.NewRows([]string{
// 			"dv.f_view_id",
// 			// "dv.f_uniform_catalog_code",
// 			"dv.f_view_name",
// 			"dv.f_technical_name",
// 			"dv.f_group_id",
// 			"COALESCE(dvg.f_group_name, '')",
// 			"dv.f_type",
// 			"dv.f_query_type",
// 			"dv.f_builtin",
// 			"dv.f_tags",
// 			"dv.f_comment",
// 			"dv.f_data_source_type",
// 			"dv.f_data_source_id",
// 			"dv.f_file_name",
// 			"dv.f_excel_config",
// 			"dv.f_data_scope",
// 			"dv.f_fields",
// 			"dv.f_status",
// 			"dv.f_metadata_form_id",
// 			"dv.f_primary_keys",
// 			"COALESCE(dv.f_sql, '')",
// 			"dv.f_create_time",
// 			"dv.f_update_time",
// 			"dv.f_creator",
// 			"dv.f_updater"},
// 		).AddRow(
// 			&view.ViewName,
// 			&view.TechnicalName,
// 			&view.GroupID,
// 			&view.GroupName,
// 			&view.Type,
// 			&view.QueryType,
// 			&view.Builtin,
// 			&tagsStr,
// 			&view.Comment,
// 			&view.DataSourceType,
// 			&view.DataSourceID,
// 			&view.FileName,
// 			&excelConfigBytes,
// 			&dataScopeBytes,
// 			&fieldsBytes,
// 			&view.Status,
// 			&view.MetadataFormID,
// 			&primaryKeysStr,
// 			&view.SQLStr,
// 			&view.CreateTime,
// 			&view.UpdateTime,
// 			&view.Creator,
// 			&view.Updater)

// 		sqlStr := fmt.Sprintf("SELECT dv.f_view_id, dv.f_view_name, dv.f_group_id, "+
// 			"COALESCE(dvg.f_group_name, ''), dv.f_tags, dv.f_data_source, dv.f_field_scope, "+
// 			"dv.f_fields, dv.f_filters, dv.f_comment, dv.f_creator, dv.f_create_time, dv.f_update_time, "+
// 			"dv.f_builtin, COALESCE(dv.f_loggroup_filters, ''), dv.f_open_streaming, dv.f_job_id, "+
// 			"COALESCE(dmj.f_job_status, ''), COALESCE(dmj.f_job_status_details, '') FROM %s as dv "+
// 			"LEFT JOIN %s as dvg on dv.f_group_id = dvg.f_group_id "+
// 			"LEFT JOIN %s as dmj on dv.f_job_id = dmj.f_job_id WHERE dv.f_view_id IN (?)",
// 			DATA_VIEW_TABLE_NAME, DATA_VIEW_GROUP_TABLE_NAME, DATA_MODEL_JOB_TABLE_NAME)

// 		Convey("Get failed, caused by the error from squirrel func ToSql", func() {
// 			expectedErr := errors.New("some error")
// 			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
// 			defer patch.Reset()

// 			viewIDs := []string{views[0].ViewID}
// 			_, err := dva.GetDataViews(testCtx, viewIDs)
// 			So(err, ShouldResemble, expectedErr)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Get failed, caused by query error", func() {
// 			expectedErr := errors.New("some error")
// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

// 			viewIDs := []string{views[0].ViewID}
// 			_, err := dva.GetDataViews(testCtx, viewIDs)
// 			So(err, ShouldResemble, expectedErr)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Get failed, caused by scan error", func() {
// 			rows := sqlmock.NewRows([]string{"view_id", "view_name"}).
// 				AddRow(views[0].ViewID, views[0].ViewName)
// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

// 			viewIDs := []string{views[0].ViewID}
// 			_, err := dva.GetDataViews(testCtx, viewIDs)
// 			So(err, ShouldNotBeNil)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Get failed, caused by dataSource unmarshal error", func() {
// 			expectedErr := errors.New("some error")
// 			patch := ApplyFuncReturn(sonic.Unmarshal, expectedErr)
// 			defer patch.Reset()

// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

// 			viewIDs := []string{views[0].ViewID}
// 			_, err := dva.GetDataViews(testCtx, viewIDs)
// 			So(err, ShouldResemble, expectedErr)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Get failed, caused by fields unmarshal error", func() {
// 			expectedErr := errors.New("some error")
// 			patch := ApplyFunc(sonic.Unmarshal,
// 				func(data []byte, v any) error {
// 					if _, ok := v.(*[]*interfaces.ViewField); ok {
// 						return expectedErr
// 					}
// 					return nil
// 				},
// 			)
// 			defer patch.Reset()

// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

// 			viewIDs := []string{views[0].ViewID}
// 			_, err := dva.GetDataViews(testCtx, viewIDs)
// 			So(err, ShouldResemble, expectedErr)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Get failed, caused by condition unmarshal error", func() {
// 			expectedErr := errors.New("some error")
// 			patch := ApplyFunc(sonic.Unmarshal,
// 				func(data []byte, v any) error {
// 					if _, ok := v.(**interfaces.CondCfg); ok {
// 						return expectedErr
// 					}
// 					if _, ok := v.(*[]*interfaces.ViewField); ok {
// 						return nil
// 					}
// 					return expectedErr
// 				},
// 			)
// 			defer patch.Reset()

// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

// 			viewIDs := []string{views[0].ViewID}
// 			_, err := dva.GetDataViews(testCtx, viewIDs)
// 			So(err, ShouldResemble, expectedErr)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Get succeed, and return correct rows", func() {
// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

// 			viewIDs := []string{views[0].ViewID}
// 			_, err := dva.GetDataViews(testCtx, viewIDs)
// 			So(err, ShouldBeNil)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})
// 	})
// }

// func Test_DataViewAccess_ListDataViews(t *testing.T) {
// 	Convey("Test ListDataViews", t, func() {
// 		appSetting := &common.AppSetting{}
// 		dva, smock := MockNewDataViewAccess(appSetting)

// 		sqlStr := fmt.Sprintf("SELECT dv.f_view_id, dv.f_view_name, dv.f_group_id, "+
// 			"COALESCE(dvg.f_group_name, ''), dv.f_tags, dv.f_data_source, dv.f_comment, dv.f_creator, dv.f_create_time, "+
// 			"dv.f_update_time, dv.f_builtin, dv.f_open_streaming FROM %s as dv LEFT JOIN %s as dvg "+
// 			"on dv.f_group_id = dvg.f_group_id WHERE dv.f_builtin IN (?) AND instr(dv.f_view_name, ?) > 0 "+
// 			"AND dv.f_group_id = ? AND dvg.f_group_name = ? AND instr(dv.f_tags, ?) > 0 "+
// 			"ORDER BY dv.f_update_time desc", DATA_VIEW_TABLE_NAME, DATA_VIEW_GROUP_TABLE_NAME)

// 		var view = views[0]
// 		var tagsStr string

// 		rows1 := sqlmock.NewRows([]string{
// 			"dv.f_view_id",
// 			"dv.f_view_name",
// 			"dv.f_technical_name",
// 			"dv.f_group_id",
// 			"COALESCE(dvg.f_group_name, '')",
// 			"dv.f_type",
// 			"dv.f_query_type",
// 			"dv.f_builtin",
// 			"dv.f_tags",
// 			"dv.f_comment",
// 			"dv.f_data_source_type",
// 			"dv.f_data_source_id",
// 			"dv.f_file_name",
// 			"dv.f_status",
// 			"dv.f_create_time",
// 			"dv.f_update_time"},
// 		).AddRow(
// 			&view.ViewID,
// 			// &view.UniformCatalogCode,
// 			&view.ViewName,
// 			&view.TechnicalName,
// 			&view.GroupID,
// 			&view.GroupName,
// 			&view.Type,
// 			&view.QueryType,
// 			&view.Builtin,
// 			&tagsStr,
// 			&view.Comment,
// 			&view.DataSourceType,
// 			&view.DataSourceID,
// 			&view.FileName,
// 			&view.Status,
// 			&view.CreateTime,
// 			&view.UpdateTime)

// 		rows2 := sqlmock.NewRows([]string{
// 			"dv.f_view_id",
// 			"dv.f_view_name",
// 			"dv.f_technical_name",
// 			"dv.f_group_id",
// 			"COALESCE(dvg.f_group_name, '')",
// 			"dv.f_type",
// 			"dv.f_query_type",
// 			"dv.f_builtin",
// 			"dv.f_tags",
// 			"dv.f_comment",
// 			"dv.f_data_source_type",
// 			"dv.f_data_source_id",
// 			"dv.f_file_name",
// 			"dv.f_status",
// 			"dv.f_create_time",
// 			"dv.f_update_time"},
// 		).AddRow(
// 			&view.ViewID,
// 			// &view.UniformCatalogCode,
// 			&view.ViewName,
// 			&view.TechnicalName,
// 			&view.GroupID,
// 			&view.GroupName,
// 			&view.Type,
// 			&view.QueryType,
// 			&view.Builtin,
// 			&tagsStr,
// 			&view.Comment,
// 			&view.DataSourceType,
// 			&view.DataSourceID,
// 			&view.FileName,
// 			&view.Status,
// 			&view.CreateTime,
// 			&view.UpdateTime).AddRow(
// 			&view.ViewID,
// 			// &view.UniformCatalogCode,
// 			&view.ViewName,
// 			&view.TechnicalName,
// 			&view.GroupID,
// 			&view.GroupName,
// 			&view.Type,
// 			&view.QueryType,
// 			&view.Builtin,
// 			&tagsStr,
// 			&view.Comment,
// 			&view.DataSourceType,
// 			&view.DataSourceID,
// 			&view.FileName,
// 			&view.Status,
// 			&view.CreateTime,
// 			&view.UpdateTime).AddRow(
// 			&view.ViewID,
// 			// &view.UniformCatalogCode,
// 			&view.ViewName,
// 			&view.TechnicalName,
// 			&view.GroupID,
// 			&view.GroupName,
// 			&view.Type,
// 			&view.QueryType,
// 			&view.Builtin,
// 			&tagsStr,
// 			&view.Comment,
// 			&view.DataSourceType,
// 			&view.DataSourceID,
// 			&view.FileName,
// 			&view.Status,
// 			&view.CreateTime,
// 			&view.UpdateTime)

// 		listQuery := &interfaces.ListViewQueryParams{
// 			NamePattern: "a",
// 			Tag:         "b",
// 			Builtin:     []bool{false},
// 			PaginationQueryParameters: interfaces.PaginationQueryParameters{
// 				Limit:     interfaces.MAX_LIMIT,
// 				Offset:    interfaces.MIN_OFFSET,
// 				Sort:      interfaces.DATA_VIEW_SORT["update_time"],
// 				Direction: interfaces.DESC_DIRECTION,
// 			},
// 		}

// 		Convey("List failed, caused by the error from squirrel func ToSql", func() {
// 			expectedErr := errors.New("some error")
// 			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
// 			defer patch.Reset()

// 			_, err := dva.ListDataViews(testCtx, listQuery)
// 			So(err, ShouldResemble, expectedErr)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("List failed, caused by query error", func() {
// 			expectedErr := errors.New("some error")
// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

// 			_, err := dva.ListDataViews(testCtx, listQuery)
// 			So(err, ShouldResemble, expectedErr)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("List failed, caused by row scan error", func() {
// 			rows := sqlmock.NewRows([]string{"f_view_id", "f_view_name"}).
// 				AddRow(views[0].ViewID, views[0].ViewName)

// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)
// 			_, err := dva.ListDataViews(testCtx, listQuery)
// 			So(err, ShouldNotBeNil)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("List succeed with limit is equal to 1", func() {
// 			sqlStr := fmt.Sprintf("SELECT dv.f_view_id, dv.f_view_name, dv.f_group_id, "+
// 				"COALESCE(dvg.f_group_name, ''), dv.f_tags, dv.f_data_source, dv.f_comment, dv.f_creator, dv.f_create_time, "+
// 				"dv.f_update_time, dv.f_builtin, dv.f_open_streaming FROM %s as dv LEFT JOIN %s as dvg "+
// 				"on dv.f_group_id = dvg.f_group_id WHERE dv.f_builtin IN (?) AND instr(dv.f_view_name, ?) > 0 "+
// 				"AND dv.f_group_id = ? AND dvg.f_group_name = ? AND instr(dv.f_tags, ?) > 0 "+
// 				"ORDER BY dv.f_update_time desc", DATA_VIEW_TABLE_NAME, DATA_VIEW_GROUP_TABLE_NAME)

// 			listQuery.Limit = 1
// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows1)

// 			entries, err := dva.ListDataViews(testCtx, listQuery)
// 			So(err, ShouldBeNil)
// 			So(len(entries), ShouldEqual, 1)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("List succeed with no limit", func() {
// 			sqlStr := fmt.Sprintf("SELECT dv.f_view_id, dv.f_view_name, dv.f_group_id, "+
// 				"COALESCE(dvg.f_group_name, ''), dv.f_tags, dv.f_data_source, dv.f_comment, dv.f_creator, dv.f_create_time, "+
// 				"dv.f_update_time, dv.f_builtin, dv.f_open_streaming FROM %s as dv LEFT JOIN %s as dvg "+
// 				"on dv.f_group_id = dvg.f_group_id WHERE dv.f_builtin IN (?) AND instr(dv.f_view_name, ?) > 0 "+
// 				"AND dv.f_group_id = ? AND dvg.f_group_name = ? AND instr(dv.f_tags, ?) > 0 "+
// 				"ORDER BY dv.f_update_time desc", DATA_VIEW_TABLE_NAME, DATA_VIEW_GROUP_TABLE_NAME)

// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows2)

// 			listQuery.Limit = -1
// 			entries, err := dva.ListDataViews(testCtx, listQuery)
// 			So(err, ShouldBeNil)
// 			So(len(entries), ShouldEqual, 2)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("List failed, caused by dataSource unmarshal error", func() {
// 			expectedErr := errors.New("some error")
// 			patch := ApplyFuncReturn(sonic.Unmarshal, expectedErr)
// 			defer patch.Reset()

// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows2)

// 			_, err := dva.ListDataViews(testCtx, listQuery)
// 			So(err, ShouldResemble, expectedErr)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})
// 	})
// }

// func Test_DataViewAccess_GetDataViewsTotal(t *testing.T) {
// 	Convey("Test GetDataViewsTotal", t, func() {
// 		appSetting := &common.AppSetting{}
// 		dva, smock := MockNewDataViewAccess(appSetting)

// 		sqlStr := fmt.Sprintf("SELECT COUNT(dv.f_view_id) FROM %s as dv LEFT JOIN %s as dvg "+
// 			"on dv.f_group_id = dvg.f_group_id WHERE instr(dv.f_view_name, ?) > 0 "+
// 			"AND dv.f_group_id = ? AND dvg.f_group_name = ?", DATA_VIEW_TABLE_NAME, DATA_VIEW_GROUP_TABLE_NAME)

// 		rows := sqlmock.NewRows([]string{"COUNT(f_view_id)"}).AddRow(1)

// 		listQuery := &interfaces.ListViewQueryParams{
// 			NamePattern: "a",
// 			PaginationQueryParameters: interfaces.PaginationQueryParameters{
// 				Limit:     interfaces.MAX_LIMIT,
// 				Offset:    interfaces.MIN_OFFSET,
// 				Sort:      interfaces.DATA_VIEW_SORT["update_time"],
// 				Direction: interfaces.DESC_DIRECTION,
// 			},
// 		}

// 		Convey("Get failed, caused by the error from squirrel func ToSql", func() {
// 			expectedErr := errors.New("some error")
// 			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
// 			defer patch.Reset()

// 			total, err := dva.GetDataViewsTotal(testCtx, listQuery)
// 			So(total, ShouldEqual, 0)
// 			So(err, ShouldResemble, expectedErr)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Get failed, caused by the scan error", func() {
// 			expectedErr := errors.New("some error")
// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

// 			total, err := dva.GetDataViewsTotal(testCtx, listQuery)
// 			So(total, ShouldEqual, 0)
// 			So(err, ShouldResemble, expectedErr)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Get succeed", func() {
// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

// 			total, err := dva.GetDataViewsTotal(testCtx, listQuery)
// 			So(total, ShouldEqual, 1)
// 			So(err, ShouldBeNil)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})
// 	})
// }

// func Test_DataViewAccess_CheckDataViewExistByName(t *testing.T) {
// 	Convey("Test CheckDataViewExistByName", t, func() {
// 		appSetting := &common.AppSetting{}
// 		dva, smock := MockNewDataViewAccess(appSetting)

// 		viewName := views[0].ViewName
// 		groupName := views[0].GroupName
// 		rows := sqlmock.NewRows([]string{"f_view_id"}).AddRow(views[0].ViewID)
// 		sqlStr := fmt.Sprintf("SELECT dv.f_view_id FROM %s as dv LEFT JOIN %s as dvg "+
// 			"on dv.f_group_id = dvg.f_group_id WHERE dv.f_view_name = ? "+
// 			"AND dvg.f_group_name = ?", DATA_VIEW_TABLE_NAME, DATA_VIEW_GROUP_TABLE_NAME)

// 		Convey("Check failed, caused by the error from squirrel func ToSql", func() {
// 			expectedErr := errors.New("some error")
// 			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
// 			defer patch.Reset()

// 			smock.ExpectBegin()

// 			tx, _ := dva.db.Begin()
// 			_, _, err := dva.CheckDataViewExistByName(testCtx, tx, viewName, groupName)
// 			So(err, ShouldResemble, expectedErr)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Get failed, caused by query error", func() {
// 			expectedErr := errors.New("some error")
// 			smock.ExpectBegin()
// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

// 			tx, _ := dva.db.Begin()
// 			_, _, err := dva.CheckDataViewExistByName(testCtx, tx, viewName, groupName)
// 			So(err, ShouldResemble, expectedErr)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Check failed, caused by the scan error", func() {
// 			expectedErr := errors.New("sql: expected 2 destination arguments in Scan, not 1")
// 			rows := sqlmock.NewRows([]string{"f_view_id", "f_view_name"}).
// 				AddRow(views[0].ViewID, views[0].ViewName)
// 			smock.ExpectBegin()
// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

// 			tx, _ := dva.db.Begin()
// 			_, _, err := dva.CheckDataViewExistByName(testCtx, tx, viewName, groupName)
// 			So(err, ShouldResemble, expectedErr)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Get failed, caused by ErrNoRows", func() {
// 			expectedErr := sql.ErrNoRows
// 			smock.ExpectBegin()
// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

// 			tx, _ := dva.db.Begin()
// 			_, exist, err := dva.CheckDataViewExistByName(testCtx, tx, viewName, groupName)
// 			So(exist, ShouldBeFalse)
// 			So(err, ShouldBeNil)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Check succeed", func() {
// 			smock.ExpectBegin()
// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

// 			tx, _ := dva.db.Begin()
// 			_, exist, err := dva.CheckDataViewExistByName(testCtx, tx, viewName, groupName)

// 			So(exist, ShouldBeTrue)
// 			So(err, ShouldBeNil)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})
// 	})
// }

// func Test_DataViewAccess_CheckDataViewExistByID(t *testing.T) {
// 	Convey("Test CheckDataViewExistByID", t, func() {
// 		appSetting := &common.AppSetting{}
// 		dva, smock := MockNewDataViewAccess(appSetting)

// 		viewID := views[0].ViewID

// 		rows := sqlmock.NewRows([]string{"f_view_name"}).AddRow(views[0].ViewName)

// 		sqlStr := fmt.Sprintf("SELECT f_view_name FROM %s WHERE f_view_id = ?", DATA_VIEW_TABLE_NAME)

// 		Convey("Get failed, caused by the error from squirrel func ToSql", func() {
// 			expectedErr := errors.New("some error")
// 			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
// 			defer patch.Reset()

// 			tx, _ := dva.db.Begin()
// 			_, _, err := dva.CheckDataViewExistByID(testCtx, tx, viewID)
// 			So(err, ShouldResemble, expectedErr)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Get failed, caused by query error", func() {
// 			expectedErr := errors.New("some error")
// 			smock.ExpectBegin()
// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

// 			tx, _ := dva.db.Begin()
// 			_, _, err := dva.CheckDataViewExistByID(testCtx, tx, viewID)
// 			So(err, ShouldResemble, expectedErr)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Get failed, caused by ErrNoRows", func() {
// 			expectedErr := sql.ErrNoRows
// 			smock.ExpectBegin()
// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

// 			tx, _ := dva.db.Begin()
// 			name, exist, err := dva.CheckDataViewExistByID(testCtx, tx, viewID)
// 			So(name, ShouldEqual, "")
// 			So(exist, ShouldBeFalse)
// 			So(err, ShouldBeNil)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Get failed, caused by the scan error", func() {
// 			expectedErr := errors.New("sql: expected 0 destination arguments in Scan, not 1")
// 			rows := sqlmock.NewRows([]string{}).AddRow()
// 			smock.ExpectBegin()
// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

// 			tx, _ := dva.db.Begin()
// 			_, _, err := dva.CheckDataViewExistByID(testCtx, tx, viewID)
// 			So(err, ShouldResemble, expectedErr)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Get succeed", func() {
// 			smock.ExpectBegin()
// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

// 			tx, _ := dva.db.Begin()
// 			name, exist, err := dva.CheckDataViewExistByID(testCtx, tx, viewID)

// 			So(name, ShouldEqual, views[0].ViewName)
// 			So(exist, ShouldBeTrue)
// 			So(err, ShouldBeNil)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})
// 	})
// }

// func Test_DataViewAccess_JointViewListQuerySQL(t *testing.T) {
// 	Convey("Test JointViewListQuerySQL", t, func() {
// 		pageParams := interfaces.PaginationQueryParameters{}
// 		pageParams.Limit = interfaces.MAX_LIMIT
// 		pageParams.Offset = interfaces.MIN_OFFSET
// 		pageParams.Sort = interfaces.DATA_VIEW_SORT["update_time"]
// 		pageParams.Direction = interfaces.DESC_DIRECTION

// 		sql := sq.Select("f_view_id", "f_view_name", "f_tags", "f_update_time").
// 			From(DATA_VIEW_TABLE_NAME)

// 		Convey("Not include builtin view", func() {
// 			expectedArgs := []interface{}{false, "test"}
// 			expectedStr := `SELECT f_view_id, f_view_name, f_tags, f_update_time FROM t_data_view WHERE dv.f_builtin IN (?) AND dv.f_view_name = ?`

// 			listQueryParams := &interfaces.ListViewQueryParams{
// 				Name:                      "test",
// 				GroupID:                   interfaces.GroupID_All,
// 				Builtin:                   []bool{false},
// 				GroupName:                 interfaces.GroupName_All,
// 				PaginationQueryParameters: pageParams,
// 			}

// 			builder, _ := buildViewListQuerySQL(listQueryParams, sql)
// 			sqlStr, args, err := builder.ToSql()
// 			So(sqlStr, ShouldEqual, expectedStr)
// 			So(args, ShouldResemble, expectedArgs)
// 			So(err, ShouldBeNil)
// 		})

// 		Convey("Open streaming is not empty", func() {
// 			expectedArgs := []any{true}
// 			expectedStr := `SELECT f_view_id, f_view_name, f_tags, f_update_time FROM t_data_view WHERE dv.f_open_streaming IN (?)`

// 			listQueryParams := &interfaces.ListViewQueryParams{
// 				GroupID:   interfaces.GroupID_All,
// 				GroupName: interfaces.GroupName_All,
// 			}

// 			builder, _ := buildViewListQuerySQL(listQueryParams, sql)
// 			sqlStr, args, err := builder.ToSql()
// 			So(sqlStr, ShouldEqual, expectedStr)
// 			So(args, ShouldResemble, expectedArgs)
// 			So(err, ShouldBeNil)
// 		})

// 		Convey("GroupID is not all", func() {
// 			expectedArgs := []interface{}{""}
// 			expectedStr := "SELECT f_view_id, f_view_name, f_tags, f_update_time FROM t_data_view WHERE dv.f_group_id = ?"

// 			listQueryParams := &interfaces.ListViewQueryParams{
// 				GroupID:   "",
// 				GroupName: interfaces.GroupName_All,
// 			}

// 			builder, _ := buildViewListQuerySQL(listQueryParams, sql)
// 			sqlStr, args, err := builder.ToSql()
// 			So(sqlStr, ShouldEqual, expectedStr)
// 			So(args, ShouldResemble, expectedArgs)
// 			So(err, ShouldBeNil)
// 		})

// 		Convey("GroupName is ungrouped", func() {
// 			expectedArgs := []interface{}{""}
// 			expectedStr := "SELECT f_view_id, f_view_name, f_tags, f_update_time FROM t_data_view WHERE dvg.f_group_name = ?"

// 			listQueryParams := &interfaces.ListViewQueryParams{
// 				GroupID:   interfaces.GroupID_All,
// 				GroupName: "",
// 			}

// 			builder, _ := buildViewListQuerySQL(listQueryParams, sql)
// 			sqlStr, args, err := builder.ToSql()
// 			So(sqlStr, ShouldEqual, expectedStr)
// 			So(args, ShouldResemble, expectedArgs)
// 			So(err, ShouldBeNil)
// 		})

// 		Convey("GroupName is not all", func() {
// 			expectedArgs := []interface{}{"test"}
// 			expectedStr := "SELECT f_view_id, f_view_name, f_tags, f_update_time FROM t_data_view WHERE dvg.f_group_name = ?"

// 			listQueryParams := &interfaces.ListViewQueryParams{
// 				GroupID:   interfaces.GroupID_All,
// 				GroupName: "test",
// 			}

// 			builder, _ := buildViewListQuerySQL(listQueryParams, sql)
// 			sqlStr, args, err := builder.ToSql()
// 			So(sqlStr, ShouldEqual, expectedStr)
// 			So(args, ShouldResemble, expectedArgs)
// 			So(err, ShouldBeNil)
// 		})

// 		Convey("Name is not an empty string", func() {
// 			expectedArgs := []interface{}{"test"}
// 			expectedStr := "SELECT f_view_id, f_view_name, f_tags, f_update_time FROM t_data_view WHERE dv.f_view_name = ?"

// 			listQueryParams := &interfaces.ListViewQueryParams{
// 				Name:                      "test",
// 				GroupID:                   interfaces.GroupID_All,
// 				GroupName:                 interfaces.GroupName_All,
// 				PaginationQueryParameters: pageParams,
// 			}

// 			builder, _ := buildViewListQuerySQL(listQueryParams, sql)
// 			sqlStr, args, err := builder.ToSql()
// 			So(sqlStr, ShouldEqual, expectedStr)
// 			So(args, ShouldResemble, expectedArgs)
// 			So(err, ShouldBeNil)
// 		})

// 		Convey("NamePattern is not an empty string", func() {
// 			expectedStr := "SELECT f_view_id, f_view_name, f_tags, f_update_time FROM t_data_view WHERE instr(dv.f_view_name, ?) > 0"
// 			expectedArgs := []interface{}{"test_1"}
// 			listQueryParams := &interfaces.ListViewQueryParams{
// 				NamePattern:               "test_1",
// 				GroupID:                   interfaces.GroupID_All,
// 				GroupName:                 interfaces.GroupName_All,
// 				PaginationQueryParameters: pageParams,
// 			}
// 			builder, _ := buildViewListQuerySQL(listQueryParams, sql)
// 			sqlStr, args, err := builder.ToSql()
// 			So(sqlStr, ShouldEqual, expectedStr)
// 			So(args, ShouldResemble, expectedArgs)
// 			So(err, ShouldBeNil)
// 		})

// 		Convey("Tag is not an empty string", func() {
// 			expectedStr := "SELECT f_view_id, f_view_name, f_tags, f_update_time FROM t_data_view WHERE instr(dv.f_tags, ?) > 0"
// 			expectedArgs := []interface{}{`"a_%"`}

// 			listQueryParams := &interfaces.ListViewQueryParams{
// 				Tag:                       "a_%",
// 				GroupID:                   interfaces.GroupID_All,
// 				GroupName:                 interfaces.GroupName_All,
// 				PaginationQueryParameters: pageParams,
// 			}
// 			builder, _ := buildViewListQuerySQL(listQueryParams, sql)
// 			sqlStr, args, err := builder.ToSql()
// 			So(sqlStr, ShouldEqual, expectedStr)
// 			So(args, ShouldResemble, expectedArgs)
// 			So(err, ShouldBeNil)
// 		})

// 		Convey("Field scope is not empty", func() {
// 			expectedStr := "SELECT f_view_id, f_view_name, f_tags, f_update_time FROM t_data_view WHERE dv.f_field_scope = ?"
// 			expectedArgs := []any{interfaces.FieldScope_Custom}

// 			listQueryParams := &interfaces.ListViewQueryParams{
// 				GroupID:   interfaces.GroupID_All,
// 				GroupName: interfaces.GroupName_All,
// 			}

// 			builder, _ := buildViewListQuerySQL(listQueryParams, sql)
// 			sqlStr, args, err := builder.ToSql()
// 			So(sqlStr, ShouldEqual, expectedStr)
// 			So(args, ShouldResemble, expectedArgs)
// 			So(err, ShouldBeNil)
// 		})
// 	})
// }

// func Test_DataViewAccess_GetDetailedDataViewMapByIDs(t *testing.T) {
// 	Convey("Test GetDetailedDataViewMapByIDs", t, func() {
// 		appSetting := &common.AppSetting{}
// 		dva, smock := MockNewDataViewAccess(appSetting)

// 		sqlStr := fmt.Sprintf("SELECT f_view_id, f_view_name, f_group_id, f_data_source, "+
// 			"f_field_scope, f_fields, f_filters, COALESCE(f_loggroup_filters, '') "+
// 			"FROM %s WHERE f_view_id IN (?)", DATA_VIEW_TABLE_NAME)

// 		viewIDs := []string{"1"}

// 		Convey("Get failed, caused by the error from squirrel func ToSql", func() {
// 			expectedErr := errors.New("some error")
// 			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
// 			defer patch.Reset()

// 			_, err := dva.GetDetailedDataViewMapByIDs(testCtx, viewIDs)
// 			So(err, ShouldNotBeNil)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Get failed, caused by query error", func() {
// 			expectedErr := errors.New("some error")
// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

// 			_, err := dva.GetDetailedDataViewMapByIDs(testCtx, viewIDs)
// 			So(err, ShouldNotBeNil)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Get failed, caused by scan error", func() {
// 			rows := sqlmock.NewRows([]string{"f_view_id", "f_view_name"}).AddRow("1", "view1")
// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

// 			_, err := dva.GetDetailedDataViewMapByIDs(testCtx, viewIDs)
// 			So(err, ShouldNotBeNil)
// 		})

// 		Convey("Get failed, caused by dataSource unmarshal error", func() {
// 			expectedErr := errors.New("some error")
// 			rows := sqlmock.NewRows([]string{
// 				"f_view_id", "f_view_name", "f_tags", "f_data_source", "f_field_scope",
// 				"f_fields", "f_filters", "f_comment", "f_update_time", "f_builtin"}).
// 				AddRow("1", "view1", "", "", 0, []byte{}, []byte{}, "", "", 0)

// 			patch := ApplyFunc(sonic.Unmarshal,
// 				func(data []byte, v any) error {
// 					if _, ok := v.(*map[string]any); ok {
// 						return expectedErr
// 					}
// 					return nil
// 				},
// 			)
// 			defer patch.Reset()

// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

// 			_, err := dva.GetDetailedDataViewMapByIDs(testCtx, viewIDs)
// 			So(err, ShouldNotBeNil)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Get failed, caused by fields unmarshal error", func() {
// 			expectedErr := errors.New("some error")

// 			rows := sqlmock.NewRows([]string{
// 				"f_view_id", "f_view_name", "f_group_id", "f_data_source",
// 				"f_field_scope", "f_fields", "f_filters", "COALESCE(f_loggroup_filters, '')",
// 			}).AddRow("1", "view1", "", []byte{}, 0, []byte{}, []byte{}, "")

// 			patch := ApplyFunc(sonic.Unmarshal,
// 				func(data []byte, v any) error {
// 					if _, ok := v.(*[]*interfaces.ViewField); ok {
// 						return expectedErr
// 					}
// 					return nil
// 				},
// 			)
// 			defer patch.Reset()

// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

// 			_, err := dva.GetDetailedDataViewMapByIDs(testCtx, viewIDs)
// 			So(err, ShouldNotBeNil)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Get failed, caused by condition unmarshal error", func() {
// 			expectedErr := errors.New("some error")

// 			rows := sqlmock.NewRows([]string{
// 				"f_view_id", "f_view_name", "f_group_id", "f_data_source",
// 				"f_field_scope", "f_fields", "f_filters", "COALESCE(f_loggroup_filters, '')",
// 			}).AddRow("1", "view1", "", []byte{}, 0, []byte{}, []byte{}, "")

// 			patch := ApplyFunc(sonic.Unmarshal,
// 				func(data []byte, v any) error {
// 					if _, ok := v.(**interfaces.CondCfg); ok {
// 						return expectedErr
// 					}
// 					if _, ok := v.(*[]*interfaces.ViewField); ok {
// 						return nil
// 					}
// 					return expectedErr
// 				},
// 			)
// 			defer patch.Reset()

// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

// 			_, err := dva.GetDetailedDataViewMapByIDs(testCtx, viewIDs)
// 			So(err, ShouldNotBeNil)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Get succeed, and return correct rows", func() {
// 			rows := sqlmock.NewRows([]string{
// 				"f_view_id", "f_view_name", "f_group_id", "f_data_source",
// 				"f_field_scope", "f_fields", "f_filters", "COALESCE(f_loggroup_filters, '')",
// 			}).AddRow("1", "view1", "", []byte{}, 0, []byte{}, []byte{}, "")

// 			patch := ApplyFuncReturn(sonic.Unmarshal, nil)
// 			defer patch.Reset()

// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

// 			_, err := dva.GetDetailedDataViewMapByIDs(testCtx, viewIDs)
// 			So(err, ShouldBeNil)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})
// 	})
// }

// func Test_DataViewAccess_GetSimpleDataViewMapByIDs(t *testing.T) {
// 	Convey("Test GetSimpleDataViewMapByIDs", t, func() {
// 		appSetting := &common.AppSetting{}
// 		dva, smock := MockNewDataViewAccess(appSetting)

// 		sqlStr := fmt.Sprintf("SELECT f_view_id, f_view_name, f_group_id, "+
// 			"f_builtin FROM %s WHERE f_view_id IN (?)", DATA_VIEW_TABLE_NAME)

// 		viewIDs := []string{"1"}

// 		Convey("Get failed, caused by the error from squirrel func ToSql", func() {
// 			expectedErr := errors.New("some error")
// 			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
// 			defer patch.Reset()

// 			_, err := dva.GetSimpleDataViewMapByIDs(testCtx, viewIDs)
// 			So(err, ShouldNotBeNil)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Get failed, caused by query error", func() {
// 			expectedErr := errors.New("some error")
// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

// 			_, err := dva.GetSimpleDataViewMapByIDs(testCtx, viewIDs)
// 			So(err, ShouldNotBeNil)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Get failed, caused by scan error", func() {
// 			rows := sqlmock.NewRows([]string{"f_view_id"}).AddRow("1")
// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

// 			_, err := dva.GetSimpleDataViewMapByIDs(testCtx, viewIDs)
// 			So(err, ShouldNotBeNil)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Get succeed, and return correct rows", func() {
// 			rows := sqlmock.NewRows([]string{"f_view_id", "f_view_name", "f_group_id", "f_builtin"}).AddRow(1, "view1", "", false)
// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

// 			_, err := dva.GetSimpleDataViewMapByIDs(testCtx, viewIDs)
// 			So(err, ShouldBeNil)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})
// 	})
// }

// func Test_DataViewAccess_GetDataViewsByGroupID(t *testing.T) {
// 	Convey("Test GetDataViewsByGroupID", t, func() {
// 		appSetting := &common.AppSetting{}
// 		dva, smock := MockNewDataViewAccess(appSetting)

// 		view := views[0]
// 		var tagsStr, primaryKeysStr string
// 		var excelConfigBytes, dataScopeBytes, fieldsBytes []byte

// 		rows := sqlmock.NewRows([]string{
// 			"dv.f_view_id",
// 			// "dv.f_uniform_catalog_code",
// 			"dv.f_view_name",
// 			"dv.f_technical_name",
// 			"dv.f_group_id",
// 			"COALESCE(dvg.f_group_name, '')",
// 			"dv.f_type",
// 			"dv.f_query_type",
// 			"dv.f_builtin",
// 			"dv.f_tags",
// 			"dv.f_comment",
// 			"dv.f_data_source_type",
// 			"dv.f_data_source_id",
// 			"dv.f_file_name",
// 			"dv.f_excel_config",
// 			"dv.f_data_scope",
// 			"dv.f_fields",
// 			"dv.f_status",
// 			"dv.f_metadata_form_id",
// 			"dv.f_primary_keys",
// 			"COALESCE(dv.f_sql, '')",
// 			"dv.f_create_time",
// 			"dv.f_update_time",
// 			"dv.f_creator",
// 			"dv.f_updater"},
// 		).AddRow(
// 			&view.ViewName,
// 			&view.TechnicalName,
// 			&view.GroupID,
// 			&view.GroupName,
// 			&view.Type,
// 			&view.QueryType,
// 			&view.Builtin,
// 			&tagsStr,
// 			&view.Comment,
// 			&view.DataSourceType,
// 			&view.DataSourceID,
// 			&view.FileName,
// 			&excelConfigBytes,
// 			&dataScopeBytes,
// 			&fieldsBytes,
// 			&view.Status,
// 			&view.MetadataFormID,
// 			&primaryKeysStr,
// 			&view.SQLStr,
// 			&view.CreateTime,
// 			&view.UpdateTime,
// 			&view.Creator,
// 			&view.Updater)

// 		sqlStr := fmt.Sprintf("SELECT dv.f_view_id, dv.f_view_name, dv.f_group_id, "+
// 			"COALESCE(dvg.f_group_name, ''), dv.f_tags, dv.f_data_source, dv.f_field_scope, "+
// 			"dv.f_fields, dv.f_filters, dv.f_comment, dv.f_creator, dv.f_create_time, dv.f_update_time, "+
// 			"dv.f_builtin, COALESCE(dv.f_loggroup_filters, ''), dv.f_open_streaming, "+
// 			"dv.f_job_id, COALESCE(dmj.f_job_status, ''), COALESCE(dmj.f_job_status_details, '') "+
// 			"FROM %s as dv LEFT JOIN %s as dvg on dv.f_group_id = dvg.f_group_id "+
// 			"LEFT JOIN %s as dmj on dv.f_job_id = dmj.f_job_id WHERE dv.f_builtin = ? "+
// 			"AND dv.f_group_id = ?",
// 			DATA_VIEW_TABLE_NAME, DATA_VIEW_GROUP_TABLE_NAME, DATA_MODEL_JOB_TABLE_NAME)

// 		groupID := "x"

// 		Convey("Get failed, caused by the error from squirrel func ToSql", func() {
// 			expectedErr := errors.New("some error")
// 			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
// 			defer patch.Reset()

// 			_, err := dva.GetDataViewsByGroupID(testCtx, groupID)
// 			So(err, ShouldResemble, expectedErr)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Get failed, caused by query error", func() {
// 			expectedErr := errors.New("some error")
// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

// 			_, err := dva.GetDataViewsByGroupID(testCtx, groupID)
// 			So(err, ShouldResemble, expectedErr)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Get failed, caused by scan error", func() {
// 			rows := sqlmock.NewRows([]string{"view_id", "view_name"}).
// 				AddRow(views[0].ViewID, views[0].ViewName)
// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

// 			_, err := dva.GetDataViewsByGroupID(testCtx, groupID)
// 			So(err, ShouldNotBeNil)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("Get succeed, and return correct rows", func() {
// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

// 			_, err := dva.GetDataViewsByGroupID(testCtx, groupID)
// 			So(err, ShouldBeNil)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})
// 	})
// }
