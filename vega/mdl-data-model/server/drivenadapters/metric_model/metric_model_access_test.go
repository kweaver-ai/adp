// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package metric_model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	sq "github.com/Masterminds/squirrel"
	. "github.com/agiledragon/gomonkey/v2"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	"data-model/common"
	"data-model/interfaces"
)

var (
	testUpdateTime = int64(1735786555379)
	testTags       = []string{"a", "b", "c", "d", "e"}

	testCtx         = context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage)
	testFormula     = "avg(rate(node_cpu_seconds_total{name=~\"node-14-35\",mode12=\"system\"}[5m]))by(name)"
	testMetricModel = interfaces.MetricModel{
		SimpleMetricModel: interfaces.SimpleMetricModel{
			ModelID:      "1",
			ModelName:    "16",
			MetricType:   "atomic",
			QueryType:    "promql",
			Tags:         testTags,
			Comment:      "ssss",
			UpdateTime:   testUpdateTime,
			CreateTime:   testUpdateTime,
			GroupID:      "0",
			UnitType:     "storeUnit",
			Unit:         "bit",
			Formula:      testFormula,
			MeasureName:  "__m.a",
			DateField:    interfaces.PROMQL_DATEFIELD,
			MeasureField: interfaces.PROMQL_METRICFIELD,
		},
		DataSource: &interfaces.MetricDataSource{
			Type: interfaces.SOURCE_TYPE_DATA_VIEW,
			ID:   "123",
		},
		ModuleType: interfaces.MODULE_TYPE_METRIC_MODEL,
	}
	emptyMetricModel = interfaces.MetricModel{
		ModuleType: interfaces.MODULE_TYPE_METRIC_MODEL}
)

func MockNewMetricModelAccess(appSetting *common.AppSetting) (*metricModelAccess, sqlmock.Sqlmock) {
	db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	mma := &metricModelAccess{
		appSetting: appSetting,
		db:         db,
	}
	return mma, smock
}

func Test_MetricModelAccess_CreateMetricModel(t *testing.T) {
	Convey("test CreateMetricModel\n", t, func() {
		appSetting := &common.AppSetting{}
		mma, smock := MockNewMetricModelAccess(appSetting)

		sqlStr := fmt.Sprintf("INSERT INTO %s (f_model_id,f_model_name,f_catalog_id,f_catalog_content,"+
			"f_measure_name,f_metric_type,f_data_source,f_query_type,f_formula,f_formula_config,f_order_by_fields,"+
			"f_having_condition,f_analysis_dimessions,f_date_field,f_measure_field,f_unit_type,f_unit,f_tags,f_comment,"+
			"f_group_id,f_builtin,f_calendar_interval,f_creator,f_creator_type,f_create_time,f_update_time) "+
			"VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", METRIC_MODEL_TABLE_NAME)

		Convey("CreateMetricModel Success \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			tx, _ := mma.db.Begin()
			err := mma.CreateMetricModel(testCtx, tx, testMetricModel)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CreateMetricModel prepare error \n", func() {
			expectedErr := errors.New("any error")
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := mma.db.Begin()
			err := mma.CreateMetricModel(testCtx, tx, testMetricModel)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CreateMetricModel  Exec sql error\n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("some error1")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := mma.db.Begin()
			err := mma.CreateMetricModel(testCtx, tx, testMetricModel)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_MetricModelAccess_GetMetricModelByID(t *testing.T) {
	Convey("test GetMetricModelByModelID\n", t, func() {
		appSetting := &common.AppSetting{}
		mma, smock := MockNewMetricModelAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT mm.f_model_id, mm.f_model_name, mm.f_measure_name, "+
			"mm.f_metric_type, mm.f_data_source, mm.f_query_type, mm.f_formula, mm.f_formula_config, "+
			"mm.f_order_by_fields, mm.f_having_condition, mm.f_analysis_dimessions, mm.f_date_field, "+
			"mm.f_measure_field, mm.f_unit_type, mm.f_unit, mm.f_tags, mm.f_comment, mm.f_group_id, "+
			"mm.f_builtin, mm.f_calendar_interval, mm.f_create_time, mm.f_update_time, mmg.f_group_name "+
			"FROM %s AS mmg JOIN %s AS mm on mm.f_group_id = mmg.f_group_id "+
			"WHERE mm.f_model_id = ?", METRIC_MODEL_GROUP_TABLE_NAME, METRIC_MODEL_TABLE_NAME)

		rows := sqlmock.NewRows([]string{"f_model_id", "f_model_name", "f_measure_name",
			"f_metric_type", "f_data_source", "f_query_type", "f_formula", "f_formula_config",
			"f_order_by_fields", "f_having_condition", "f_analysis_dimessions",
			"f_date_field", "f_measure_field", "f_unit_type", "f_unit", "f_tags", "f_comment", "f_group_id",
			"f_builtin", "f_calendar_interval", "f_create_time", "f_update_time", "f_group_name"},
		).AddRow(
			"1", "16", "__m.a", "atomic", `{"type": "data_view", "id": "123"}`, "promql", testFormula, "{}", "[]", "{}", "[]",
			interfaces.PROMQL_DATEFIELD, interfaces.PROMQL_METRICFIELD, "storeUnit",
			"bit", "\"a\"", "ssss", "0", 0, 0, testUpdateTime, testUpdateTime, "默认分组")

		metricModel := interfaces.MetricModel{
			SimpleMetricModel: interfaces.SimpleMetricModel{
				ModelID:         "1",
				ModelName:       "16",
				MetricType:      "atomic",
				QueryType:       "promql",
				Tags:            []string{"a"},
				Comment:         "ssss",
				UpdateTime:      testUpdateTime,
				GroupID:         "0",
				GroupName:       "默认分组",
				CreateTime:      testUpdateTime,
				UnitType:        "storeUnit",
				Unit:            "bit",
				Formula:         testFormula,
				MeasureName:     "__m.a",
				DateField:       interfaces.PROMQL_DATEFIELD,
				MeasureField:    interfaces.PROMQL_METRICFIELD,
				AnalysisDims:    []interfaces.Field{},
				OrderByFields:   []interfaces.OrderField{},
				HavingCondition: &interfaces.CondCfg{},
			},
			DataSource: &interfaces.MetricDataSource{
				Type: interfaces.SOURCE_TYPE_DATA_VIEW,
				ID:   "123",
			},
			ModuleType: interfaces.MODULE_TYPE_METRIC_MODEL,
		}

		modelID := "1"

		Convey("GetMetricModelByModelID Success \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			mModel, exists, err := mma.GetMetricModelByModelID(testCtx, modelID)
			So(err, ShouldBeNil)
			So(mModel, ShouldResemble, metricModel)
			So(exists, ShouldBeTrue)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetMetricModelByModelID Success no row \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(sql.ErrNoRows)

			mModel, exists, err := mma.GetMetricModelByModelID(testCtx, modelID)
			So(mModel, ShouldResemble, emptyMetricModel)
			So(exists, ShouldBeFalse)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetMetricModelByModelID Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			mModel, exists, err := mma.GetMetricModelByModelID(testCtx, modelID)
			So(mModel, ShouldResemble, interfaces.MetricModel{})
			So(exists, ShouldBeFalse)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("TagString2TagSlice \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			mModel, exists, err := mma.GetMetricModelByModelID(testCtx, modelID)
			So(mModel, ShouldResemble, metricModel)
			So(exists, ShouldBeTrue)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_MetricModelAccess_GetMetricModelsByIDs(t *testing.T) {
	Convey("test GetMetricModelsByIDs\n", t, func() {
		appSetting := &common.AppSetting{}
		mma, smock := MockNewMetricModelAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT mm.f_model_id, mm.f_model_name, mm.f_catalog_id, mm.f_catalog_content, mm.f_measure_name, "+
			"mm.f_metric_type, mm.f_data_source, mm.f_query_type, mm.f_formula, mm.f_formula_config, "+
			"mm.f_order_by_fields, mm.f_having_condition, mm.f_analysis_dimessions, mm.f_date_field, "+
			"mm.f_measure_field, mm.f_unit_type, mm.f_unit, mm.f_tags, mm.f_comment, mm.f_group_id, "+
			"mm.f_builtin, mm.f_calendar_interval, mm.f_creator, mm.f_creator_type, mm.f_create_time, mm.f_update_time, mmg.f_group_name "+
			"FROM %s AS mmg JOIN %s AS mm on mm.f_group_id = mmg.f_group_id "+
			"WHERE mm.f_model_id IN (?)", METRIC_MODEL_GROUP_TABLE_NAME, METRIC_MODEL_TABLE_NAME)

		rows := sqlmock.NewRows([]string{"f_model_id", "f_model_name", "f_catalog_id", "f_catalog_content", "f_measure_name",
			"f_metric_type", "f_data_source", "f_query_type", "f_formula", "f_formula_config",
			"f_order_by_fields", "f_having_conditio", "f_analysis_dimessions",
			"f_date_field", "f_measure_field", "f_unit_type", "f_unit", "f_tags", "f_comment", "f_group_id",
			"f_builtin", "f_calendar_interval", "f_creator", "f_creator_type", "f_create_time", "f_update_time", "f_group_name"},
		).AddRow(
			"1", "16", "", "", "__m.a", "atomic", `{"type": "data_view", "id": "123"}`, "promql", testFormula, "{}", "[]", "{}", "[]",
			interfaces.PROMQL_DATEFIELD, interfaces.PROMQL_METRICFIELD, "storeUnit",
			"bit", "\"a\"", "ssss", "0", 0, 0, interfaces.ADMIN_ID, interfaces.ADMIN_TYPE, testUpdateTime, testUpdateTime, "默认分组")

		metricModel := interfaces.MetricModel{
			SimpleMetricModel: interfaces.SimpleMetricModel{
				ModelID:         "1",
				ModelName:       "16",
				GroupID:         "0",
				GroupName:       "默认分组",
				MetricType:      "atomic",
				QueryType:       "promql",
				Tags:            []string{"a"},
				Comment:         "ssss",
				UpdateTime:      testUpdateTime,
				CreateTime:      testUpdateTime,
				UnitType:        "storeUnit",
				Unit:            "bit",
				Formula:         testFormula,
				MeasureName:     "__m.a",
				DateField:       interfaces.PROMQL_DATEFIELD,
				MeasureField:    interfaces.PROMQL_METRICFIELD,
				OrderByFields:   []interfaces.OrderField{},
				HavingCondition: &interfaces.CondCfg{},
				AnalysisDims:    []interfaces.Field{},
				Creator: interfaces.AccountInfo{
					ID:   interfaces.ADMIN_ID,
					Type: interfaces.ADMIN_TYPE,
				},
			},
			DataSource: &interfaces.MetricDataSource{
				Type: interfaces.SOURCE_TYPE_DATA_VIEW,
				ID:   "123",
			},
			ModuleType: interfaces.MODULE_TYPE_METRIC_MODEL,
		}

		modelIDs := []string{"1"}

		Convey("GetMetricModelsByIDs Success \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			mModel, err := mma.GetMetricModelsByModelIDs(testCtx, modelIDs)
			So(mModel, ShouldResemble, []interfaces.MetricModel{metricModel})
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetMetricModelsByIDs Success no row \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(sqlmock.NewRows(nil))

			mModel, err := mma.GetMetricModelsByModelIDs(testCtx, modelIDs)
			So(mModel, ShouldResemble, []interfaces.MetricModel{})
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetMetricModelsByIDs Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			mModel, err := mma.GetMetricModelsByModelIDs(testCtx, modelIDs)
			So(mModel, ShouldResemble, []interfaces.MetricModel{})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("TagString2TagSlice \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			mModel, err := mma.GetMetricModelsByModelIDs(testCtx, modelIDs)
			metricModel.Tags = []string{"a"}
			So(mModel, ShouldResemble, []interfaces.MetricModel{metricModel})
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_MetricModelAccess_UpdateMetricModel(t *testing.T) {
	Convey("Test UpdateMetricModel\n", t, func() {
		appSetting := &common.AppSetting{}
		mma, smock := MockNewMetricModelAccess(appSetting)

		sqlStr := fmt.Sprintf("UPDATE %s SET f_analysis_dimessions = ?, f_builtin = ?, f_calendar_interval = ?, "+
			"f_catalog_content = ?, f_catalog_id = ?, "+
			"f_comment = ?, f_data_source = ?, f_date_field = ?, f_formula = ?, f_formula_config = ?, "+
			"f_group_id = ?, f_having_condition = ?, f_measure_field = ?, f_metric_type = ?, "+
			"f_model_name = ?, f_order_by_fields = ?,  f_query_type = ?, f_tags = ?, "+
			"f_unit = ?, f_unit_type = ?, f_update_time = ? WHERE f_model_id = ?", METRIC_MODEL_TABLE_NAME)

		metricModel := interfaces.MetricModel{
			SimpleMetricModel: interfaces.SimpleMetricModel{
				ModelID:      "1",
				ModelName:    "16",
				MetricType:   "atomic",
				QueryType:    "promql",
				Tags:         []string{"a"},
				Comment:      "ssss",
				UpdateTime:   testUpdateTime,
				GroupID:      "1",
				UnitType:     "storeUnit",
				Unit:         "bit",
				Formula:      testFormula,
				MeasureName:  "__m.a",
				DateField:    interfaces.PROMQL_DATEFIELD,
				MeasureField: interfaces.PROMQL_METRICFIELD,
			},
			DataSource: &interfaces.MetricDataSource{
				Type: interfaces.SOURCE_TYPE_DATA_VIEW,
				ID:   "123",
			},
			ModuleType: interfaces.MODULE_TYPE_METRIC_MODEL,
		}

		Convey("UpdateMetricModel Success \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			tx, _ := mma.db.Begin()
			err := mma.UpdateMetricModel(testCtx, tx, metricModel)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateMetricModel failed prepare \n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("prepare error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := mma.db.Begin()
			err := mma.UpdateMetricModel(testCtx, tx, metricModel)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateMetricModel Failed UpdateSql \n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("sql exec error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := mma.db.Begin()
			err := mma.UpdateMetricModel(testCtx, tx, metricModel)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateMetricModel RowsAffected 2 \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 2))

			tx, _ := mma.db.Begin()
			err := mma.UpdateMetricModel(testCtx, tx, metricModel)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateMetricModel Failed RowsAffected \n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("Get RowsAffected error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewErrorResult(expectedErr))

			tx, _ := mma.db.Begin()
			err := mma.UpdateMetricModel(testCtx, tx, metricModel)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_MetricModelAccess_DeleteMetricModels(t *testing.T) {
	Convey("Test DeleteMetricModels\n", t, func() {
		appSetting := &common.AppSetting{}
		mma, smock := MockNewMetricModelAccess(appSetting)

		sqlStr := fmt.Sprintf("DELETE FROM %s WHERE f_model_id IN (?,?)", METRIC_MODEL_TABLE_NAME)

		modelIDs := []string{"1", "2"}

		Convey("DeleteMetricModels Success \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 2))

			tx, _ := mma.db.Begin()
			rowsAffected, err := mma.DeleteMetricModels(testCtx, tx, modelIDs)
			So(err, ShouldBeNil)
			So(rowsAffected, ShouldEqual, 2)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteMetricModels failed prepare \n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("prepare error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := mma.db.Begin()
			_, err := mma.DeleteMetricModels(testCtx, tx, modelIDs)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteMetricModels RowsAffected 0 \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 0))

			tx, _ := mma.db.Begin()
			rowsAffected, err := mma.DeleteMetricModels(testCtx, tx, modelIDs)
			So(rowsAffected, ShouldEqual, 0)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteMetricModels Failed RowsAffected \n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("Get RowsAffected error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewErrorResult(expectedErr))

			tx, _ := mma.db.Begin()
			_, err := mma.DeleteMetricModels(testCtx, tx, modelIDs)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteMetricModels Failed dbExec \n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("dbExec error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := mma.db.Begin()
			_, err := mma.DeleteMetricModels(testCtx, tx, modelIDs)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteMetricModels null \n", func() {
			smock.ExpectBegin()

			tx, _ := mma.db.Begin()
			rowsAffected, err := mma.DeleteMetricModels(testCtx, tx, []string{})
			So(err, ShouldBeNil)
			So(rowsAffected, ShouldEqual, 0)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

// func Test_MetricModelAccess_ListMetricModels(t *testing.T) {
// 	Convey("Test ListMetricModels\n", t, func() {
// 		appSetting := &common.AppSetting{}
// 		mma, smock := MockNewMetricModelAccess(appSetting)

// 		sqlStr := fmt.Sprintf("SELECT mm.f_model_id, mm.f_model_name, mm.f_measure_name, "+
// 			"mm.f_metric_type, mm.f_data_view_id, mm.f_query_type, mm.f_formula, "+
// 			"mm.f_date_field, mm.f_measure_field, mm.f_unit_type, mm.f_unit, mm.f_tags, "+
// 			"mm.f_comment, mm.f_group_id, mm.f_builtin, mm.f_calendar_interval, "+
// 			"mm.f_create_time, mm.f_update_time, mmg.f_group_name "+
// 			"FROM %s AS mmg JOIN %s AS mm on mm.f_group_id = mmg.f_group_id "+
// 			"WHERE instr(mm.f_model_name, ?) > 0 AND mm.f_group_id = ? "+
// 			"ORDER BY mm.f_update_time desc LIMIT 1000 OFFSET 0",
// 			METRIC_MODEL_GROUP_TABLE_NAME, METRIC_MODEL_TABLE_NAME)

// 		rows := sqlmock.NewRows([]string{"mm.f_model_id", "mm.f_model_name", "mm.f_measure_name",
// 			"mm.f_metric_type", "mm.f_data_view_id", "mm.f_query_type", "mm.f_formula",
// 			"mm.f_date_field", "mm.f_measure_field", "mm.f_unit_type", "mm.f_unit", "mm.f_tags",
// 			"mm.f_comment", "mm.f_group_id", "mm.f_builtin", "mm.f_calendar_interval",
// 			"mm.f_create_time", "mm.f_update_time", "mmg.f_group_name"}).
// 			AddRow("1", "16", "__m.a", "atomic", "123", "promql", testFormula,
// 				interfaces.PROMQL_DATEFIELD, interfaces.PROMQL_METRICFIELD, "storeUnit",
// 				"bit", "\"a\"", "ssss", "1", 0, 0, testUpdateTime, testUpdateTime, "默认分组")

// 		modelQuery := interfaces.MetricModelsQueryParams{
// 			NamePattern: "a",
// 			PaginationQueryParameters: interfaces.PaginationQueryParameters{
// 				Limit:     interfaces.MAX_LIMIT,
// 				Offset:    interfaces.MIN_OFFSET,
// 				Sort:      interfaces.METRIC_MODEL_SORT["update_time"],
// 				Direction: interfaces.DESC_DIRECTION,
// 			},
// 			GroupID: "",
// 		}

// 		Convey("ListMetricModels Success list \n", func() {
// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

// 			_, err := mma.ListMetricModels(testCtx, modelQuery)
// 			So(err, ShouldBeNil)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("ListMetricModels Success list with sort model_name\n", func() {
// 			sqlStr := fmt.Sprintf("SELECT mm.f_model_id, mm.f_model_name, mm.f_measure_name, "+
// 				"mm.f_metric_type, mm.f_data_view_id, mm.f_query_type, mm.f_formula, "+
// 				"mm.f_date_field, mm.f_measure_field, mm.f_unit_type, mm.f_unit, mm.f_tags, "+
// 				"mm.f_comment, mm.f_group_id, mm.f_builtin, mm.f_calendar_interval, "+
// 				"mm.f_create_time, mm.f_update_time, mmg.f_group_name "+
// 				"FROM %s AS mmg JOIN %s AS mm on mm.f_group_id = mmg.f_group_id "+
// 				"WHERE instr(mm.f_model_name, ?) > 0 AND mm.f_group_id = ? "+
// 				"ORDER BY mm.f_model_name desc LIMIT 1000 OFFSET 0",
// 				METRIC_MODEL_GROUP_TABLE_NAME, METRIC_MODEL_TABLE_NAME)

// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

// 			modelQuery.Sort = interfaces.METRIC_MODEL_SORT["model_name"]
// 			_, err := mma.ListMetricModels(testCtx, modelQuery)
// 			So(err, ShouldBeNil)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("ListMetricModels Success list with sort group_name\n", func() {
// 			sqlStr := fmt.Sprintf("SELECT mm.f_model_id, mm.f_model_name, mm.f_measure_name, "+
// 				"mm.f_metric_type, mm.f_data_view_id, mm.f_query_type, mm.f_formula, mm.f_date_field, "+
// 				"mm.f_measure_field, mm.f_unit_type, mm.f_unit, mm.f_tags, mm.f_comment, mm.f_group_id, "+
// 				"mm.f_builtin, mm.f_calendar_interval, mm.f_create_time, mm.f_update_time, mmg.f_group_name "+
// 				"FROM %s AS mmg JOIN %s AS mm on mm.f_group_id = mmg.f_group_id "+
// 				"WHERE instr(mm.f_model_name, ?) > 0 AND mm.f_group_id = ? "+
// 				"ORDER BY mmg.f_group_name desc,mm.f_model_name desc LIMIT 1000 OFFSET 0",
// 				METRIC_MODEL_GROUP_TABLE_NAME, METRIC_MODEL_TABLE_NAME)

// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

// 			modelQuery.Sort = interfaces.METRIC_MODEL_SORT["group_name"]
// 			_, err := mma.ListMetricModels(testCtx, modelQuery)
// 			So(err, ShouldBeNil)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("ListMetricModels Success list no limit\n", func() {
// 			sqlStr := fmt.Sprintf("SELECT mm.f_model_id, mm.f_model_name, mm.f_measure_name, "+
// 				"mm.f_metric_type, mm.f_data_view_id, mm.f_query_type, mm.f_formula, "+
// 				"mm.f_date_field, mm.f_measure_field, mm.f_unit_type, mm.f_unit, mm.f_tags, "+
// 				"mm.f_comment, mm.f_group_id, mm.f_builtin, mm.f_calendar_interval, "+
// 				"mm.f_create_time, mm.f_update_time, mmg.f_group_name "+
// 				"FROM %s AS mmg JOIN %s AS mm on mm.f_group_id = mmg.f_group_id "+
// 				"WHERE instr(mm.f_model_name, ?) > 0 AND mm.f_group_id = ? "+
// 				"ORDER BY mm.f_update_time desc", METRIC_MODEL_GROUP_TABLE_NAME, METRIC_MODEL_TABLE_NAME)

// 			modelQuery.Limit = -1
// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

// 			_, err := mma.ListMetricModels(testCtx, modelQuery)
// 			So(err, ShouldBeNil)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("ListMetricModels Failed dbQuery \n", func() {
// 			expectedErr := errors.New("dbQuery error")
// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

// 			_, err := mma.ListMetricModels(testCtx, modelQuery)
// 			So(err, ShouldResemble, expectedErr)

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})

// 		Convey("ListMetricModels scan error \n", func() {
// 			rows := sqlmock.NewRows([]string{"model_id", "model_name", "metric_type",
// 				"f_data_view_id", "query_type", "formula", "tags", "custom_attribute", "comment"}).
// 				AddRow("1", "16", "atomic", "123", "promql", testFormula, "\"a\"", "", "ssss")

// 			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)
// 			_, err := mma.ListMetricModels(testCtx, modelQuery)
// 			So(err.Error(), ShouldEqual, "sql: expected 9 destination arguments in Scan, not 19")

// 			if err := smock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})
// 	})
// }

func Test_MetricModelAccess_ListSimpleMetricModels(t *testing.T) {
	Convey("Test ListSimpleMetricModels\n", t, func() {
		appSetting := &common.AppSetting{}
		mma, smock := MockNewMetricModelAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT mm.f_model_id, mm.f_model_name, mm.f_measure_name, "+
			"mm.f_group_id, mmg.f_group_name, mm.f_tags, mm.f_comment, mm.f_metric_type, "+
			"mm.f_query_type, mm.f_formula, mm.f_formula_config, mm.f_order_by_fields, mm.f_having_condition, "+
			"mm.f_analysis_dimessions, mm.f_date_field, mm.f_measure_field, "+
			"mm.f_unit_type, mm.f_unit, mm.f_builtin, mm.f_calendar_interval, mm.f_create_time, "+
			"mm.f_update_time FROM %s AS mmg JOIN %s AS mm on mm.f_group_id = mmg.f_group_id "+
			"WHERE instr(mm.f_model_name, ?) > 0 AND mm.f_group_id = ? "+
			"ORDER BY mm.f_update_time desc",
			METRIC_MODEL_GROUP_TABLE_NAME, METRIC_MODEL_TABLE_NAME)

		rows := sqlmock.NewRows([]string{
			"mm.f_model_id", "mm.f_model_name", "mm.f_measure_name", "mm.f_group_id", "mmg.f_group_name",
			"mm.f_tags", "mm.f_comment", "mm.f_metric_type", "mm.f_query_type",
			"mm.f_formula", "mm.f_formula_config", "mm.f_order_by_fields", "mm.f_having_condition", "mm.f_analysis_dimessions",
			"mm.f_date_field", "mm.f_measure_field",
			"mm.f_unit_type", "mm.f_unit", "mm.f_builtin", "mm.f_calendar_interval", "mm.f_create_time", "mm.f_update_time"},
		).AddRow(
			"1", "16", "__m.a", "group_id_1", "group_name_1", "tags", "comments", "atomic",
			"promql", "irate(a)", "{}", "[]", "{}", "[]", "@timestamp", "value", "numUnit", "%", 0, 0, testUpdateTime, testUpdateTime)

		modelQuery := interfaces.MetricModelsQueryParams{
			NamePattern: "a",
			PaginationQueryParameters: interfaces.PaginationQueryParameters{
				Limit:     interfaces.MAX_LIMIT,
				Offset:    interfaces.MIN_OFFSET,
				Sort:      interfaces.METRIC_MODEL_SORT["update_time"],
				Direction: interfaces.DESC_DIRECTION,
			},
			GroupID: "",
		}

		Convey("ListSimpleMetricModels Success list \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := mma.ListSimpleMetricModels(testCtx, modelQuery)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListSimpleMetricModels Success list with sort model_name \n", func() {
			sqlStr := fmt.Sprintf("SELECT mm.f_model_id, mm.f_model_name, mm.f_measure_name, "+
				"mm.f_group_id, mmg.f_group_name, mm.f_tags, mm.f_comment, mm.f_metric_type, "+
				"mm.f_query_type, mm.f_formula, mm.f_formula_config, mm.f_order_by_fields, mm.f_having_condition, "+
				"mm.f_analysis_dimessions, mm.f_date_field, mm.f_measure_field, "+
				"mm.f_unit_type, mm.f_unit, mm.f_builtin, mm.f_calendar_interval, mm.f_create_time, "+
				"mm.f_update_time FROM %s AS mmg JOIN %s AS mm on mm.f_group_id = mmg.f_group_id "+
				"WHERE instr(mm.f_model_name, ?) > 0 AND mm.f_group_id = ? "+
				"ORDER BY mm.f_model_name desc",
				METRIC_MODEL_GROUP_TABLE_NAME, METRIC_MODEL_TABLE_NAME)

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			modelQuery.Sort = interfaces.METRIC_MODEL_SORT["model_name"]
			_, err := mma.ListSimpleMetricModels(testCtx, modelQuery)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListSimpleMetricModels Success list with sort group_name \n", func() {
			sqlStr := fmt.Sprintf("SELECT mm.f_model_id, mm.f_model_name, mm.f_measure_name, "+
				"mm.f_group_id, mmg.f_group_name, mm.f_tags, mm.f_comment, mm.f_metric_type, "+
				"mm.f_query_type, mm.f_formula, mm.f_formula_config, mm.f_order_by_fields, mm.f_having_condition, "+
				"mm.f_analysis_dimessions, mm.f_date_field, mm.f_measure_field, "+
				"mm.f_unit_type, mm.f_unit, mm.f_builtin, mm.f_calendar_interval, mm.f_create_time, "+
				"mm.f_update_time FROM %s AS mmg JOIN %s AS mm on mm.f_group_id = mmg.f_group_id "+
				"WHERE instr(mm.f_model_name, ?) > 0 AND mm.f_group_id = ? "+
				"ORDER BY mmg.f_group_name desc,mm.f_model_name desc",
				METRIC_MODEL_GROUP_TABLE_NAME, METRIC_MODEL_TABLE_NAME)

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			modelQuery.Sort = interfaces.METRIC_MODEL_SORT["group_name"]
			_, err := mma.ListSimpleMetricModels(testCtx, modelQuery)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListSimpleMetricModels Success list no limit\n", func() {
			sqlStr := fmt.Sprintf("SELECT mm.f_model_id, mm.f_model_name, mm.f_measure_name, "+
				"mm.f_group_id, mmg.f_group_name, mm.f_tags, mm.f_comment, mm.f_metric_type, "+
				"mm.f_query_type, mm.f_formula, mm.f_formula_config, mm.f_order_by_fields, mm.f_having_condition, "+
				"mm.f_analysis_dimessions, mm.f_date_field, mm.f_measure_field, "+
				"mm.f_unit_type, mm.f_unit, mm.f_builtin, mm.f_calendar_interval, mm.f_create_time, "+
				"mm.f_update_time FROM %s AS mmg JOIN %s AS mm on mm.f_group_id = mmg.f_group_id "+
				"WHERE instr(mm.f_model_name, ?) > 0 AND mm.f_group_id = ? "+
				"ORDER BY mm.f_update_time desc",
				METRIC_MODEL_GROUP_TABLE_NAME, METRIC_MODEL_TABLE_NAME)

			modelQuery.Limit = -1
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := mma.ListSimpleMetricModels(testCtx, modelQuery)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListSimpleMetricModels Failed dbQuery \n", func() {
			expectedErr := errors.New("dbQuery error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, err := mma.ListSimpleMetricModels(testCtx, modelQuery)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListSimpleMetricModels scan error \n", func() {
			rows := sqlmock.NewRows([]string{"model_id", "model_name", "metric_type", "query_type", "tags"}).
				AddRow("1", "16", "atomic", "promql", "tag1")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := mma.ListSimpleMetricModels(testCtx, modelQuery)
			So(err.Error(), ShouldEqual, "sql: expected 5 destination arguments in Scan, not 22")

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_MetricModelAccess_GetMetricModelsTotal(t *testing.T) {
	Convey("Test GetMetricModelsTotal\n", t, func() {
		appSetting := &common.AppSetting{}
		mma, smock := MockNewMetricModelAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT COUNT(mm.f_model_id) FROM %s AS mm "+
			"WHERE instr(mm.f_model_name, ?) > 0 AND mm.f_group_id = ?", METRIC_MODEL_TABLE_NAME)

		rows := sqlmock.NewRows([]string{"count(mm.model_id)"}).AddRow(1)

		modelQuery := interfaces.MetricModelsQueryParams{
			NamePattern: "a",
			PaginationQueryParameters: interfaces.PaginationQueryParameters{
				Limit:     interfaces.MAX_LIMIT,
				Offset:    interfaces.MIN_OFFSET,
				Sort:      interfaces.METRIC_MODEL_SORT["update_time"],
				Direction: interfaces.DESC_DIRECTION,
			},
			GroupID: "",
		}

		Convey("GetMetricModelsTotal Success\n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			total, err := mma.GetMetricModelsTotal(testCtx, modelQuery)
			So(total, ShouldEqual, 1)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetMetricModelsTotal Failed  Query error\n", func() {
			expectedErr := errors.New("Query error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			total, err := mma.GetMetricModelsTotal(testCtx, modelQuery)
			So(total, ShouldEqual, 0)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetMetricModelsTotal scan error \n", func() {
			rows := sqlmock.NewRows([]string{"count(model_id)"}).AddRow("s")

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)
			_, err := mma.GetMetricModelsTotal(testCtx, modelQuery)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_MetricModelAccess_ProcessQueryCondition(t *testing.T) {
	Convey("test processQueryCondition ", t, func() {
		appSetting := &common.AppSetting{}
		mma, _ := MockNewMetricModelAccess(appSetting)

		pageParams := interfaces.PaginationQueryParameters{
			Limit:     interfaces.MAX_LIMIT,
			Offset:    interfaces.MIN_OFFSET,
			Sort:      interfaces.METRIC_MODEL_SORT["update_time"],
			Direction: interfaces.DESC_DIRECTION,
		}

		sqlBuilder := sq.Select("COUNT(f_model_id)").From(METRIC_MODEL_TABLE_NAME)

		Convey("name query ", func() {
			modelQuery := interfaces.MetricModelsQueryParams{
				Name:                      "name_a",
				GroupID:                   "",
				PaginationQueryParameters: pageParams,
			}

			expectedSqlStr := fmt.Sprintf("SELECT COUNT(f_model_id) FROM %s "+
				"WHERE mm.f_model_name = ? AND mm.f_group_id = ?", METRIC_MODEL_TABLE_NAME)

			sqlBuilder := mma.processQueryCondition(modelQuery, sqlBuilder)
			sqlStr, _, _ := sqlBuilder.ToSql()
			So(sqlStr, ShouldEqual, expectedSqlStr)
		})

		Convey("NamePattern query ", func() {
			modelQuery := interfaces.MetricModelsQueryParams{
				NamePattern:               "name_a",
				GroupID:                   "qq",
				PaginationQueryParameters: pageParams,
			}

			expectedSqlStr := fmt.Sprintf("SELECT COUNT(f_model_id) FROM %s "+
				"WHERE instr(mm.f_model_name, ?) > 0 AND mm.f_group_id = ?", METRIC_MODEL_TABLE_NAME)

			sqlBuilder := mma.processQueryCondition(modelQuery, sqlBuilder)
			sqlStr, _, _ := sqlBuilder.ToSql()
			So(sqlStr, ShouldEqual, expectedSqlStr)
		})

		Convey("MetricType query ", func() {
			modelQuery := interfaces.MetricModelsQueryParams{
				MetricType:                "atomic",
				GroupID:                   "qq",
				PaginationQueryParameters: pageParams,
			}

			expectedSqlStr := fmt.Sprintf("SELECT COUNT(f_model_id) FROM %s "+
				"WHERE mm.f_metric_type = ? AND mm.f_group_id = ?", METRIC_MODEL_TABLE_NAME)

			sqlBuilder := mma.processQueryCondition(modelQuery, sqlBuilder)
			sqlStr, _, _ := sqlBuilder.ToSql()
			So(sqlStr, ShouldEqual, expectedSqlStr)
		})

		Convey("Tag query ", func() {
			modelQuery := interfaces.MetricModelsQueryParams{
				Tag:                       "tag1",
				GroupID:                   interfaces.GroupID_All,
				PaginationQueryParameters: pageParams,
			}

			expectedSqlStr := fmt.Sprintf("SELECT COUNT(f_model_id) FROM %s "+
				"WHERE instr(mm.f_tags, ?) > 0", METRIC_MODEL_TABLE_NAME)

			sqlBuilder := mma.processQueryCondition(modelQuery, sqlBuilder)
			sqlStr, _, _ := sqlBuilder.ToSql()
			So(sqlStr, ShouldEqual, expectedSqlStr)
		})

		Convey("QueryType query ", func() {
			modelQuery := interfaces.MetricModelsQueryParams{
				QueryType:                 "promql",
				GroupID:                   interfaces.GroupID_All,
				PaginationQueryParameters: pageParams,
			}

			expectedSqlStr := fmt.Sprintf("SELECT COUNT(f_model_id) FROM %s "+
				"WHERE mm.f_query_type IN (?)", METRIC_MODEL_TABLE_NAME)

			sqlBuilder := mma.processQueryCondition(modelQuery, sqlBuilder)
			sqlStr, _, _ := sqlBuilder.ToSql()
			So(sqlStr, ShouldEqual, expectedSqlStr)
		})
	})
}

func Test_MetricModelAccess_GetMetricModelSimpleInfosByIDs(t *testing.T) {
	Convey("Test GetMetricModelSimpleInfosByIDs", t, func() {
		appSetting := &common.AppSetting{}
		mma, smock := MockNewMetricModelAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT mm.f_model_id, mm.f_model_name, mmg.f_group_name, "+
			"mm.f_unit_type, mm.f_unit, mm.f_builtin FROM %s AS mm JOIN %s AS mmg "+
			"ON mm.f_group_id = mmg.f_group_id WHERE mm.f_model_id IN (?)",
			METRIC_MODEL_TABLE_NAME, METRIC_MODEL_GROUP_TABLE_NAME)

		modelSimpleInfo := interfaces.SimpleMetricModel{
			ModelID:   "1",
			ModelName: "model1",
			GroupName: "group1",
		}

		rows := sqlmock.NewRows([]string{
			"mm.f_model_id", "mm.f_model_name", "mmg.f_group_name",
			"mm.f_unit_type", "mm.f_unit", "mm.f_builtin"},
		).AddRow(
			modelSimpleInfo.ModelID, modelSimpleInfo.ModelName, modelSimpleInfo.GroupName,
			modelSimpleInfo.UnitType, modelSimpleInfo.Unit, modelSimpleInfo.Builtin)

		modelIDs := []string{modelSimpleInfo.ModelID}

		Convey("Get failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			_, err := mma.GetMetricModelSimpleInfosByIDs(testCtx, modelIDs)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by query error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, err := mma.GetMetricModelSimpleInfosByIDs(testCtx, modelIDs)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Check failed, caused by the scan error", func() {
			expectedErr := errors.New("sql: expected 1 destination arguments in Scan, not 6")
			rows := sqlmock.NewRows([]string{"model_name"}).AddRow(modelSimpleInfo.ModelID)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := mma.GetMetricModelSimpleInfosByIDs(testCtx, modelIDs)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get succeed", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			modelMap, err := mma.GetMetricModelSimpleInfosByIDs(testCtx, modelIDs)
			So(modelMap, ShouldResemble, map[string]interfaces.SimpleMetricModel{
				modelSimpleInfo.ModelID: modelSimpleInfo,
			})
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_MetricModelAccess_GetMetricModelsByGroupID(t *testing.T) {
	Convey("test GetMetricModelsByGroupID\n", t, func() {
		appSetting := &common.AppSetting{}
		mma, smock := MockNewMetricModelAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT mm.f_model_id, mm.f_model_name, mm.f_catalog_id, mm.f_catalog_content, mm.f_metric_type, mm.f_data_source, "+
			"mm.f_query_type, mm.f_formula, mm.f_formula_config, mm.f_order_by_fields, mm.f_having_condition, "+
			"mm.f_analysis_dimessions, mm.f_date_field, mm.f_measure_field, mm.f_unit_type, mm.f_unit, mm.f_tags, "+
			"mm.f_comment, mm.f_group_id, mm.f_update_time, mmg.f_group_name "+
			"FROM %s AS mmg JOIN %s AS mm on mm.f_group_id = mmg.f_group_id "+
			"WHERE mm.f_group_id = ?", METRIC_MODEL_GROUP_TABLE_NAME, METRIC_MODEL_TABLE_NAME)

		rows := sqlmock.NewRows([]string{"mm.f_model_id", "mm.f_model_name", "mm.f_catalog_id", "mm.f_catalog_content", "mm.f_metric_type",
			"mm.f_data_source", "mm.f_query_type", "mm.f_formula", "mm.f_formula_config",
			"mm.f_order_by_fields", "mm.f_having_condition", "mm.f_analysis_dimessions", "mm.f_date_field",
			"mm.f_measure_field", "mm.f_unit_type", "mm.f_unit", "mm.f_tags", "mm.f_comment", "mm.f_group_id", "mm.f_update_time", "mm.f_group_name"},
		).AddRow("1", "16", "", "", "atomic", `{"type": "data_view", "id": "123"}`, "promql", testFormula, "{}", "[]", "{}", "[]",
			interfaces.PROMQL_DATEFIELD, interfaces.PROMQL_METRICFIELD,
			"storeUnit", "bit", "\"a\"", "ssss", "0", testUpdateTime, "")

		metricModel := interfaces.MetricModel{
			SimpleMetricModel: interfaces.SimpleMetricModel{
				ModelID:         "1",
				ModelName:       "16",
				MetricType:      "atomic",
				QueryType:       "promql",
				Tags:            []string{"a"},
				Comment:         "ssss",
				UpdateTime:      testUpdateTime,
				GroupID:         "0",
				UnitType:        "storeUnit",
				Unit:            "bit",
				Formula:         testFormula,
				DateField:       interfaces.PROMQL_DATEFIELD,
				MeasureField:    interfaces.PROMQL_METRICFIELD,
				AnalysisDims:    []interfaces.Field{},
				OrderByFields:   []interfaces.OrderField{},
				HavingCondition: &interfaces.CondCfg{},
			},
			DataSource: &interfaces.MetricDataSource{
				Type: interfaces.SOURCE_TYPE_DATA_VIEW,
				ID:   "123",
			},
			ModuleType: interfaces.MODULE_TYPE_METRIC_MODEL,
		}

		groupID := "1"

		Convey("GetMetricModelsByGroupID Success \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			models, err := mma.GetMetricModelsByGroupID(testCtx, groupID)
			So(models, ShouldResemble, []interfaces.MetricModel{metricModel})
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetMetricModelsByGroupID Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			models, err := mma.GetMetricModelsByGroupID(testCtx, groupID)
			So(models, ShouldResemble, []interfaces.MetricModel{})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_MetricModelAccess_UpdateMetricModelsGroupID(t *testing.T) {
	Convey("Test UpdateMetricModelsGroupID\n", t, func() {
		appSetting := &common.AppSetting{}
		mma, smock := MockNewMetricModelAccess(appSetting)

		sqlStr := fmt.Sprintf("UPDATE %s SET f_group_id = ? "+
			"WHERE f_model_id IN (?,?)", METRIC_MODEL_TABLE_NAME)

		modelIDs := []string{"1", "2"}

		Convey("UpdateMetricModelsGroupID Success \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 2))

			tx, _ := mma.db.Begin()
			rowsAffected, err := mma.UpdateMetricModelsGroupID(testCtx, tx, modelIDs, "3")
			So(err, ShouldBeNil)
			So(rowsAffected, ShouldEqual, 2)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateMetricModelsGroupID failed prepare \n", func() {
			expectedErr := errors.New("prepare error")
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := mma.db.Begin()
			_, err := mma.UpdateMetricModelsGroupID(testCtx, tx, modelIDs, "3")
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateMetricModelsGroupID Failed RowsAffected \n", func() {
			expectedErr := errors.New("Get RowsAffected error")
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewErrorResult(expectedErr))

			tx, _ := mma.db.Begin()
			_, err := mma.UpdateMetricModelsGroupID(testCtx, tx, modelIDs, "2")
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateMetricModelsGroupID Failed dbExec \n", func() {
			expectedErr := errors.New("dbExec error")
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := mma.db.Begin()
			_, err := mma.UpdateMetricModelsGroupID(testCtx, tx, modelIDs, "2")
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateMetricModelsGroupID null \n", func() {
			smock.ExpectBegin()

			tx, _ := mma.db.Begin()
			rowsAffected, err := mma.UpdateMetricModelsGroupID(testCtx, tx, []string{}, "2")
			So(err, ShouldBeNil)
			So(rowsAffected, ShouldEqual, 0)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_MetricModelAccess_GetMetricModelIDByName(t *testing.T) {
	Convey("test GetMetricModelIDByName\n", t, func() {
		appSetting := &common.AppSetting{}
		mma, smock := MockNewMetricModelAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT mm.f_model_id FROM %s AS mmg JOIN %s AS mm "+
			"on mmg.f_group_id = mm.f_group_id WHERE mm.f_model_name = ? "+
			"AND mmg.f_group_name = ?", METRIC_MODEL_GROUP_TABLE_NAME, METRIC_MODEL_TABLE_NAME)

		rows := sqlmock.NewRows([]string{"mm.f_model_id"}).AddRow("1")

		Convey("GetMetricModelIDByName Success \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			modelID, exist, err := mma.GetMetricModelIDByName(testCtx, "group1", "model1")
			So(modelID, ShouldEqual, "1")
			So(exist, ShouldBeTrue)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetMetricModelIDByName Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			modelID, exist, err := mma.GetMetricModelIDByName(testCtx, "group1", "model1")
			So(modelID, ShouldEqual, "")
			So(exist, ShouldBeFalse)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Group Not Found  \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(sql.ErrNoRows)

			modelID, exist, err := mma.GetMetricModelIDByName(testCtx, "group1", "model1")
			So(modelID, ShouldEqual, "")
			So(exist, ShouldBeFalse)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_MetricModelAccess_CheckDuplicateMeasureName(t *testing.T) {
	Convey("test CheckMetricModelByMeasureName\n", t, func() {
		appSetting := &common.AppSetting{}
		mma, smock := MockNewMetricModelAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_model_id FROM %s WHERE f_measure_name = ?", METRIC_MODEL_TABLE_NAME)

		rows := sqlmock.NewRows([]string{"model_id"}).AddRow("1")

		Convey("CheckMetricModelByMeasureName Success \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, exist, err := mma.CheckMetricModelByMeasureName(testCtx, "name1")
			So(err, ShouldBeNil)
			So(exist, ShouldBeTrue)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CheckMetricModelByMeasureName Success no row \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(sql.ErrNoRows)

			_, exist, err := mma.CheckMetricModelByMeasureName(testCtx, "name1")
			So(err, ShouldBeNil)
			So(exist, ShouldBeFalse)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CheckMetricModelByMeasureName Failed \n", func() {
			expectedErr := errors.New("some error7")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, exist, err := mma.CheckMetricModelByMeasureName(testCtx, "name1")
			So(err, ShouldResemble, expectedErr)
			So(exist, ShouldBeFalse)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}
