// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package trace_model

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

func MockNewTraceModelAccess(appSetting *common.AppSetting) (*traceModelAccess, sqlmock.Sqlmock) {
	db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	tma := &traceModelAccess{
		appSetting: appSetting,
		db:         db,
	}

	return tma, smock
}

func Test_TraceModelAccess_CreateTraceModels(t *testing.T) {
	Convey("Test CreateTraceModels", t, func() {
		appSetting := &common.AppSetting{}
		tma, smock := MockNewTraceModelAccess(appSetting)

		sqlStr := fmt.Sprintf("INSERT INTO %s (f_model_id,f_model_name,f_tags,"+
			"f_comment,f_creator,f_creator_type,f_create_time,f_update_time,f_span_source_type,f_span_config,"+
			"f_enabled_related_log,f_related_log_source_type,f_related_log_config) "+
			"VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)", TRACE_MODEL_TABLE_NAME)

		models := []interfaces.TraceModel{
			{
				ID: "1",
			},
		}

		Convey("Create failed, caused by the error from func ProcessBeforeStore", func() {
			expectedErr := errors.New("some error")
			patch := ApplyPrivateMethod(&traceModelAccess{}, "processBeforeStore",
				func(t *traceModelAccess, ctx context.Context, model interfaces.TraceModel) (interfaces.TraceModel, error) {
					return interfaces.TraceModel{}, expectedErr
				},
			)
			defer patch.Reset()

			err := tma.CreateTraceModels(testCtx, models)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Create failed, caused by the error from squirrel func ToSql", func() {
			patch1 := ApplyPrivateMethod(&traceModelAccess{}, "processBeforeStore",
				func(t *traceModelAccess, ctx context.Context, model interfaces.TraceModel) (interfaces.TraceModel, error) {
					return interfaces.TraceModel{}, nil
				},
			)
			defer patch1.Reset()

			expectedErr := errors.New("some error")
			patch2 := ApplyMethodReturn(sq.InsertBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch2.Reset()

			err := tma.CreateTraceModels(testCtx, models)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Create failed, caused by exec sql error", func() {
			patch1 := ApplyPrivateMethod(&traceModelAccess{}, "processBeforeStore",
				func(t *traceModelAccess, ctx context.Context, model interfaces.TraceModel) (interfaces.TraceModel, error) {
					return interfaces.TraceModel{}, nil
				},
			)
			defer patch1.Reset()

			expectedErr := errors.New("some error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			err := tma.CreateTraceModels(testCtx, models)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Create succeed", func() {
			patch1 := ApplyPrivateMethod(&traceModelAccess{}, "processBeforeStore",
				func(t *traceModelAccess, ctx context.Context, model interfaces.TraceModel) (interfaces.TraceModel, error) {
					return interfaces.TraceModel{}, nil
				},
			)
			defer patch1.Reset()

			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			err := tma.CreateTraceModels(testCtx, models)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_TraceModelAccess_DeleteTraceModels(t *testing.T) {
	Convey("Test DeleteTraceModels", t, func() {
		appSetting := &common.AppSetting{}
		tma, smock := MockNewTraceModelAccess(appSetting)

		sqlStr := fmt.Sprintf("DELETE FROM %s WHERE f_model_id IN (?)", TRACE_MODEL_TABLE_NAME)

		modelIDs := []string{"1"}

		Convey("Delete failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.DeleteBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			err := tma.DeleteTraceModels(testCtx, modelIDs)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Delete failed, caused by exec sql error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			err := tma.DeleteTraceModels(testCtx, modelIDs)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Delete succeed", func() {
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 2))

			err := tma.DeleteTraceModels(testCtx, modelIDs)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_TraceModelAccess_UpdateTraceModel(t *testing.T) {
	Convey("Test UpdateTraceModel", t, func() {
		appSetting := &common.AppSetting{}
		tma, smock := MockNewTraceModelAccess(appSetting)

		sqlStr := fmt.Sprintf("UPDATE %s SET f_comment = ?, f_enabled_related_log = ?, "+
			"f_model_name = ?, f_related_log_config = ?, f_related_log_source_type = ?, "+
			"f_span_config = ?, f_span_source_type = ?, f_tags = ?, f_update_time = ? "+
			"WHERE f_model_id = ? ", TRACE_MODEL_TABLE_NAME)

		model := interfaces.TraceModel{
			ID: "1",
		}

		Convey("Update failed, caused by vertices marshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyPrivateMethod(&traceModelAccess{}, "processBeforeStore",
				func(t *traceModelAccess, ctx context.Context, model interfaces.TraceModel) (interfaces.TraceModel, error) {
					return interfaces.TraceModel{}, expectedErr
				},
			)
			defer patch.Reset()

			err := tma.UpdateTraceModel(testCtx, model)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Update failed, caused by the error from squirrel func ToSql", func() {
			patch1 := ApplyPrivateMethod(&traceModelAccess{}, "processBeforeStore",
				func(t *traceModelAccess, ctx context.Context, model interfaces.TraceModel) (interfaces.TraceModel, error) {
					return interfaces.TraceModel{}, nil
				},
			)
			defer patch1.Reset()

			expectedErr := errors.New("some error")
			patch2 := ApplyMethodReturn(sq.UpdateBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch2.Reset()

			err := tma.UpdateTraceModel(testCtx, model)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Update failed, caused by exec sql error", func() {
			patch1 := ApplyPrivateMethod(&traceModelAccess{}, "processBeforeStore",
				func(t *traceModelAccess, ctx context.Context, model interfaces.TraceModel) (interfaces.TraceModel, error) {
					return interfaces.TraceModel{}, nil
				},
			)
			defer patch1.Reset()

			expectedErr := errors.New("some error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			err := tma.UpdateTraceModel(testCtx, model)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Update succeed", func() {
			patch1 := ApplyPrivateMethod(&traceModelAccess{}, "processBeforeStore",
				func(t *traceModelAccess, ctx context.Context, model interfaces.TraceModel) (interfaces.TraceModel, error) {
					return interfaces.TraceModel{}, nil
				},
			)
			defer patch1.Reset()

			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			err := tma.UpdateTraceModel(testCtx, model)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_TraceModelAccess_GetDetailedTraceModelMapByIDs(t *testing.T) {
	Convey("Test GetDetailedTraceModelMapByIDs", t, func() {
		appSetting := &common.AppSetting{}
		tma, smock := MockNewTraceModelAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_model_id, f_model_name, f_tags, f_comment, "+
			"f_creator, f_creator_type, f_create_time, f_update_time, f_span_source_type, f_span_config, "+
			"f_enabled_related_log, f_related_log_source_type, f_related_log_config "+
			"FROM %s WHERE f_model_id IN (?,?)", TRACE_MODEL_TABLE_NAME)

		modelIDs := []string{"1", "2"}

		Convey("Get failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			_, err := tma.GetDetailedTraceModelMapByIDs(testCtx, modelIDs)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by query error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, err := tma.GetDetailedTraceModelMapByIDs(testCtx, modelIDs)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by scan error", func() {
			rows := sqlmock.NewRows([]string{"f_model_id"}).AddRow("1")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := tma.GetDetailedTraceModelMapByIDs(testCtx, modelIDs)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get succeed, and return correct rows", func() {
			expectedRows := sqlmock.NewRows([]string{
				"f_model_id", "f_model_name", "f_tags", "f_comment", "f_creator", "f_creator_type",
				"f_create_time", "f_update_time", "f_span_source_type", "f_span_config",
				"f_enabled_related_log", "f_related_log_source_type", "f_related_log_config",
			}).AddRow(
				1, "model1", "", "", interfaces.ADMIN_ID, interfaces.ADMIN_TYPE,
				testNow, testNow, interfaces.SOURCE_TYPE_DATA_VIEW, []byte("{}"),
				0, "", []byte("{}"),
			)

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(expectedRows)

			_, err := tma.GetDetailedTraceModelMapByIDs(testCtx, modelIDs)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_TraceModelAccess_ListTraceModels(t *testing.T) {
	Convey("Test ListTraceModels", t, func() {
		appSetting := &common.AppSetting{}
		tma, smock := MockNewTraceModelAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_model_id, f_model_name, f_span_source_type, f_tags, "+
			"f_comment, f_creator, f_creator_type, f_create_time, f_update_time "+
			"FROM %s WHERE instr(f_model_name, ?) > 0 "+
			"ORDER BY f_update_time desc", TRACE_MODEL_TABLE_NAME)

		listQueryPara := interfaces.TraceModelListQueryParams{
			CommonListQueryParams: interfaces.CommonListQueryParams{
				NamePattern: "a",
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:     interfaces.MAX_LIMIT,
					Offset:    interfaces.MIN_OFFSET,
					Sort:      interfaces.TRACE_MODEL_SORT["update_time"],
					Direction: interfaces.DESC_DIRECTION,
				},
			},
		}

		Convey("List failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			_, err := tma.ListTraceModels(testCtx, listQueryPara)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("List failed, caused by query error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, err := tma.ListTraceModels(testCtx, listQueryPara)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("List failed, caused by row scan error", func() {
			expectedRows := sqlmock.NewRows([]string{"f_model_id", "f_model_name"}).AddRow("1", "model1")

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(expectedRows)
			_, err := tma.ListTraceModels(testCtx, listQueryPara)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("List succeed", func() {
			expectedRows := sqlmock.NewRows([]string{"f_model_id", "f_model_name", "f_span_source_type",
				"f_tags", "f_comment", "f_creator", "f_creator_type", "f_create_time", "f_update_time"}).
				AddRow("1", "model1", "data_view", "", "",
					interfaces.ADMIN_ID, interfaces.ADMIN_TYPE, testNow, testNow)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(expectedRows)

			_, err := tma.ListTraceModels(testCtx, listQueryPara)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_TraceModelAccess_GetTraceModelTotal(t *testing.T) {
	Convey("Test GetTraceModelTotal", t, func() {
		appSetting := &common.AppSetting{}
		tma, smock := MockNewTraceModelAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT COUNT(f_model_id) FROM %s "+
			"WHERE f_model_name = ?", TRACE_MODEL_TABLE_NAME)

		listQueryPara := interfaces.TraceModelListQueryParams{
			CommonListQueryParams: interfaces.CommonListQueryParams{
				Name: "a",
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:     interfaces.MAX_LIMIT,
					Offset:    interfaces.MIN_OFFSET,
					Sort:      interfaces.TRACE_MODEL_SORT["update_time"],
					Direction: interfaces.DESC_DIRECTION,
				},
			},
		}

		Convey("Get failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			total, err := tma.GetTraceModelTotal(testCtx, listQueryPara)
			So(total, ShouldEqual, 0)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by the scan error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			total, err := tma.GetTraceModelTotal(testCtx, listQueryPara)
			So(total, ShouldEqual, 0)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get succeed", func() {
			expectedRows := sqlmock.NewRows([]string{"COUNT(f_model_id)"}).AddRow(1)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(expectedRows)

			total, err := tma.GetTraceModelTotal(testCtx, listQueryPara)
			So(total, ShouldEqual, 1)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_TraceModelAccess_GetSimpleTraceModelMapByIDs(t *testing.T) {
	Convey("Test GetSimpleTraceModelMapByIDs", t, func() {
		appSetting := &common.AppSetting{}
		tma, smock := MockNewTraceModelAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_model_id, f_model_name FROM %s "+
			"WHERE f_model_id IN (?)", TRACE_MODEL_TABLE_NAME)

		modelIDs := []string{"1"}

		Convey("Get failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			_, err := tma.GetSimpleTraceModelMapByIDs(testCtx, modelIDs)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by query error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, err := tma.GetSimpleTraceModelMapByIDs(testCtx, modelIDs)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by scan error", func() {
			expectedErr := errors.New("sql: expected 1 destination arguments in Scan, not 2")
			rows := sqlmock.NewRows([]string{"f_model_id"}).AddRow(1)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := tma.GetSimpleTraceModelMapByIDs(testCtx, modelIDs)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get succeed, and return correct rows", func() {
			expectedRows := sqlmock.NewRows([]string{"f_model_id", "f_model_name"}).AddRow(1, "model1")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(expectedRows)

			_, err := tma.GetSimpleTraceModelMapByIDs(testCtx, modelIDs)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_TraceModelAccess_GetSimpleTraceModelMapByNames(t *testing.T) {
	Convey("Test GetSimpleTraceModelMapByNames", t, func() {
		appSetting := &common.AppSetting{}
		tma, smock := MockNewTraceModelAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_model_id, f_model_name "+
			"FROM %s WHERE f_model_name IN (?)", TRACE_MODEL_TABLE_NAME)

		modelNames := []string{"1"}

		Convey("Get failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			_, err := tma.GetSimpleTraceModelMapByNames(testCtx, modelNames)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by query error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, err := tma.GetSimpleTraceModelMapByNames(testCtx, modelNames)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by scan error", func() {
			expectedErr := errors.New("sql: expected 1 destination arguments in Scan, not 2")
			rowsErr := sqlmock.NewRows([]string{"f_model_id"}).AddRow("1")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rowsErr)

			_, err := tma.GetSimpleTraceModelMapByNames(testCtx, modelNames)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get succeed, and return correct rows", func() {
			expectedRows := sqlmock.NewRows([]string{"f_model_id", "f_model_name"}).AddRow("1", "model1")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(expectedRows)

			_, err := tma.GetSimpleTraceModelMapByNames(testCtx, modelNames)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_TraceModelAccess_ExtendSQLBuilder(t *testing.T) {
	Convey("Test extendSQLBuilder", t, func() {
		appSetting := &common.AppSetting{}
		tma, _ := MockNewTraceModelAccess(appSetting)

		pageParams := interfaces.PaginationQueryParameters{
			Limit:     interfaces.MAX_LIMIT,
			Offset:    interfaces.MIN_OFFSET,
			Sort:      interfaces.TRACE_MODEL_SORT["update_time"],
			Direction: interfaces.DESC_DIRECTION,
		}

		sqlBuilder := sq.Select(
			"f_model_id",
			"f_model_name",
			"f_span_source_type",
			"f_tags",
			"f_comment",
			"f_create_time",
			"f_update_time",
		).From(TRACE_MODEL_TABLE_NAME)

		Convey("Name is not an empty string, and tag is an empty string", func() {
			expectedArgs := []interface{}{"test"}
			expectedStr := "SELECT f_model_id, f_model_name, f_span_source_type, f_tags, f_comment," +
				" f_create_time, f_update_time FROM " + TRACE_MODEL_TABLE_NAME +
				" WHERE f_model_name = ?"

			listQueryParams := interfaces.TraceModelListQueryParams{
				CommonListQueryParams: interfaces.CommonListQueryParams{
					Name:                      "test",
					PaginationQueryParameters: pageParams,
				},
			}

			sqlStr, args, err := tma.extendSQLBuilder(listQueryParams, sqlBuilder).ToSql()
			So(sqlStr, ShouldEqual, expectedStr)
			So(args, ShouldResemble, expectedArgs)
			So(err, ShouldBeNil)
		})

		Convey("NamePattern is not an empty string, and tag is an empty string", func() {
			expectedArgs := []any{"test_1"}
			expectedStr := "SELECT f_model_id, f_model_name, f_span_source_type, f_tags, f_comment," +
				" f_create_time, f_update_time FROM " + TRACE_MODEL_TABLE_NAME +
				" WHERE instr(f_model_name, ?) > 0"

			listQueryParams := interfaces.TraceModelListQueryParams{
				CommonListQueryParams: interfaces.CommonListQueryParams{
					NamePattern:               "test_1",
					PaginationQueryParameters: pageParams,
				},
			}
			sqlStr, args, err := tma.extendSQLBuilder(listQueryParams, sqlBuilder).ToSql()
			So(sqlStr, ShouldEqual, expectedStr)
			So(args, ShouldResemble, expectedArgs)
			So(err, ShouldBeNil)
		})

		Convey("Both name and tag are not empty strings", func() {
			expectedArgs := []any{"test", "\"tag_1\""}
			expectedStr := "SELECT f_model_id, f_model_name, f_span_source_type, f_tags, f_comment," +
				" f_create_time, f_update_time FROM " + TRACE_MODEL_TABLE_NAME +
				" WHERE f_model_name = ? AND instr(f_tags, ?) > 0"

			listQueryParams := interfaces.TraceModelListQueryParams{
				CommonListQueryParams: interfaces.CommonListQueryParams{
					Name:                      "test",
					Tag:                       "tag_1",
					PaginationQueryParameters: pageParams,
				},
			}

			sqlStr, args, err := tma.extendSQLBuilder(listQueryParams, sqlBuilder).ToSql()
			So(sqlStr, ShouldEqual, expectedStr)
			So(args, ShouldResemble, expectedArgs)
			So(err, ShouldBeNil)
		})

		Convey("span_source_type is not empty strings", func() {
			expectedArgs := []any{"test", "\"tag_1\"", "data_view"}
			expectedStr := "SELECT f_model_id, f_model_name, f_span_source_type, f_tags, f_comment," +
				" f_create_time, f_update_time FROM " + TRACE_MODEL_TABLE_NAME +
				" WHERE f_model_name = ? AND instr(f_tags, ?) > 0 AND f_span_source_type IN (?)"

			listQueryParams := interfaces.TraceModelListQueryParams{
				SpanSourceTypes: []string{"data_view"},
				CommonListQueryParams: interfaces.CommonListQueryParams{
					Name:                      "test",
					Tag:                       "tag_1",
					PaginationQueryParameters: pageParams,
				},
			}

			sqlStr, args, err := tma.extendSQLBuilder(listQueryParams, sqlBuilder).ToSql()
			So(sqlStr, ShouldEqual, expectedStr)
			So(args, ShouldResemble, expectedArgs)
			So(err, ShouldBeNil)
		})

		Convey("Both name and NamePattern are empty strings", func() {
			expectedStr := "SELECT f_model_id, f_model_name, f_span_source_type, f_tags, f_comment," +
				" f_create_time, f_update_time FROM " + TRACE_MODEL_TABLE_NAME

			listQueryParams := interfaces.TraceModelListQueryParams{
				CommonListQueryParams: interfaces.CommonListQueryParams{
					PaginationQueryParameters: pageParams,
				},
			}

			sqlStr, _, err := tma.extendSQLBuilder(listQueryParams, sqlBuilder).ToSql()
			So(sqlStr, ShouldEqual, expectedStr)
			So(err, ShouldBeNil)
		})
	})
}

func Test_TraceModelAccess_ProcessBeforeStore(t *testing.T) {
	Convey("Test processBeforeStore", t, func() {
		appSetting := &common.AppSetting{}
		tma, _ := MockNewTraceModelAccess(appSetting)

		Convey("Process failed, caused by span_config marshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFuncReturn(sonic.Marshal, []byte{}, expectedErr)
			defer patch.Reset()

			model := interfaces.TraceModel{}
			_, err := tma.processBeforeStore(testCtx, model)
			So(err, ShouldEqual, expectedErr)
		})

		Convey("Process failed, caused by related_log_config marshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFunc(sonic.Marshal,
				func(v any) ([]byte, error) {
					if _, ok := v.(interfaces.SpanConfigWithDataView); ok {
						return []byte{}, nil
					}
					return []byte{}, expectedErr
				},
			)
			defer patch.Reset()

			model := interfaces.TraceModel{
				SpanConfig: interfaces.SpanConfigWithDataView{},
			}
			_, err := tma.processBeforeStore(testCtx, model)
			So(err, ShouldEqual, expectedErr)
		})

		Convey("Process success", func() {
			patch := ApplyFuncReturn(sonic.Marshal, []byte{}, nil)
			defer patch.Reset()

			model := interfaces.TraceModel{}
			_, err := tma.processBeforeStore(testCtx, model)
			So(err, ShouldBeNil)
		})
	})
}

func Test_TraceModelAccess_ProcessAfterGet(t *testing.T) {
	Convey("Test processAfterGet", t, func() {
		appSetting := &common.AppSetting{}
		tma, _ := MockNewTraceModelAccess(appSetting)

		Convey("Process failed, caused by span_config of data_view unmarshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFuncReturn(sonic.Unmarshal, expectedErr)
			defer patch.Reset()

			model := interfaces.TraceModel{
				SpanSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
			}
			_, err := tma.processAfterGet(testCtx, model)
			So(err, ShouldEqual, expectedErr)
		})

		Convey("Process failed, caused by span_config of data_connection unmarshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFuncReturn(sonic.Unmarshal, expectedErr)
			defer patch.Reset()

			model := interfaces.TraceModel{
				SpanSourceType: interfaces.SOURCE_TYPE_DATA_CONNECTION,
			}
			_, err := tma.processAfterGet(testCtx, model)
			So(err, ShouldEqual, expectedErr)
		})

		Convey("Process failed, caused by related_log_config of data_view unmarshal error", func() {
			expectedErr := errors.New("some error")
			count := 0
			patch := ApplyFunc(sonic.Unmarshal, func(data []byte, v any) error {
				if count >= 0 {
					return expectedErr
				}
				count++
				return nil
			})
			defer patch.Reset()

			model := interfaces.TraceModel{
				SpanSourceType:       interfaces.SOURCE_TYPE_DATA_CONNECTION,
				EnabledRelatedLog:    interfaces.RELATED_LOG_OPEN,
				RelatedLogSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
			}

			_, err := tma.processAfterGet(testCtx, model)
			So(err, ShouldEqual, expectedErr)
		})

		Convey("Process succeed", func() {
			patch := ApplyFuncReturn(sonic.Unmarshal, nil)
			defer patch.Reset()

			model := interfaces.TraceModel{
				SpanSourceType:       interfaces.SOURCE_TYPE_DATA_CONNECTION,
				EnabledRelatedLog:    interfaces.RELATED_LOG_OPEN,
				RelatedLogSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
			}

			_, err := tma.processAfterGet(testCtx, model)
			So(err, ShouldBeNil)
		})
	})
}
