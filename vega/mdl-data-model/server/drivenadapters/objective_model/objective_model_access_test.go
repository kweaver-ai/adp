// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package objective_model

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
	testCtx = context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage)
)

func MockNewObjectiveModelAccess(appSetting *common.AppSetting) (*objectiveModelAccess, sqlmock.Sqlmock) {
	db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	oma := &objectiveModelAccess{
		appSetting: appSetting,
		db:         db,
	}
	return oma, smock
}

func Test_ObjectiveModelAccess_CheckObjectiveModelExistByID(t *testing.T) {
	Convey("Test CheckObjectiveModelExistByID", t, func() {
		appSetting := &common.AppSetting{}
		oma, smock := MockNewObjectiveModelAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_model_name FROM %s "+
			"WHERE f_model_id = ?", OBJECTIVE_MODEL_TABLE_NAME)

		Convey("When query succeeds and model exists", func() {
			rows := sqlmock.NewRows([]string{"f_model_name"}).AddRow("1")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, exists, err := oma.CheckObjectiveModelExistByID(testCtx, "test-id")
			So(err, ShouldBeNil)
			So(exists, ShouldBeTrue)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("When query succeeds and model does not exist", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(sql.ErrNoRows)

			_, exists, err := oma.CheckObjectiveModelExistByID(testCtx, "test-id")
			So(err, ShouldBeNil)
			So(exists, ShouldBeFalse)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("When query fails", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(sqlmock.ErrCancelled)

			_, exists, err := oma.CheckObjectiveModelExistByID(testCtx, "test-id")
			So(err, ShouldNotBeNil)
			So(exists, ShouldBeFalse)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_ObjectiveModelAccess_CheckObjectiveModelExistByName(t *testing.T) {
	Convey("Test CheckObjectiveModelExistByName", t, func() {
		appSetting := &common.AppSetting{}
		oma, smock := MockNewObjectiveModelAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_model_id FROM %s "+
			"WHERE f_model_name = ?", OBJECTIVE_MODEL_TABLE_NAME)

		Convey("When query succeeds and model exists", func() {
			rows := sqlmock.NewRows([]string{"f_model_id"}).AddRow("1")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, exists, err := oma.CheckObjectiveModelExistByName(testCtx, "test-name")
			So(err, ShouldBeNil)
			So(exists, ShouldBeTrue)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("When query succeeds and model does not exist", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(sql.ErrNoRows)

			_, exists, err := oma.CheckObjectiveModelExistByName(testCtx, "test-name")
			So(err, ShouldBeNil)
			So(exists, ShouldBeFalse)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("When query fails", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(sqlmock.ErrCancelled)

			_, exists, err := oma.CheckObjectiveModelExistByName(testCtx, "test-name")
			So(err, ShouldNotBeNil)
			So(exists, ShouldBeFalse)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_ObjectiveModelAccess_CreateObjectiveModel(t *testing.T) {
	Convey("Test CreateObjectiveModel", t, func() {
		appSetting := &common.AppSetting{}
		oma, smock := MockNewObjectiveModelAccess(appSetting)

		sqlStr := fmt.Sprintf("INSERT INTO %s (f_model_id,f_model_name,f_objective_type,"+
			"f_objective_config,f_tags,f_comment,f_creator,f_creator_type,f_create_time,f_update_time) "+
			"VALUES (?,?,?,?,?,?,?,?,?,?)", OBJECTIVE_MODEL_TABLE_NAME)

		model := interfaces.ObjectiveModel{
			ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
				ModelID:       "test-id",
				ModelName:     "test-model",
				ObjectiveType: interfaces.KPI,
				ObjectiveConfig: interfaces.KPIObjective{
					Objective: &[]float64{99}[0],
					Unit:      "ms",
					ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
						{
							ID:     "123",
							Weight: &[]int64{100}[0],
						},
					},
				},
			},
			Task: &interfaces.MetricTask{
				IndexBase: "test-index",
			},
		}

		Convey("When insert succeeds", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			tx, _ := oma.db.Begin()
			err := oma.CreateObjectiveModel(testCtx, tx, model)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("When insert fails", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(sqlmock.ErrCancelled)

			tx, _ := oma.db.Begin()
			err := oma.CreateObjectiveModel(testCtx, tx, model)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("When objective config marshal fails", func() {
			model.ObjectiveConfig = make(chan int) // Invalid JSON type

			smock.ExpectBegin()

			tx, _ := oma.db.Begin()
			err := oma.CreateObjectiveModel(testCtx, tx, model)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_ObjectiveModelAccess_ListObjectiveModels(t *testing.T) {
	Convey("Test ListObjectiveModels", t, func() {
		appSetting := &common.AppSetting{}
		oma, smock := MockNewObjectiveModelAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_model_id, f_model_name, f_objective_type, f_objective_config, "+
			"f_tags, f_comment, f_create_time, f_update_time FROM %s WHERE instr(f_model_name, ?) > 0 "+
			"ORDER BY f_update_time desc", OBJECTIVE_MODEL_TABLE_NAME)

		model := interfaces.ObjectiveModel{
			ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
				ModelID:       "test-id",
				ModelName:     "test-model",
				ObjectiveType: interfaces.KPI,
				ObjectiveConfig: interfaces.KPIObjective{
					Objective: &[]float64{99}[0],
					Unit:      "ms",
				},
				Tags:       []string{"tag1", "tag2"},
				Comment:    "test comment",
				UpdateTime: 1732945349330,
			},
		}

		rows := sqlmock.NewRows([]string{
			"f_model_id",
			"f_model_name",
			"f_objective_type",
			"f_objective_config",
			"f_tags",
			"f_comment",
			"f_create_time",
			"f_update_time",
		}).AddRow(
			model.ModelID,
			model.ModelName,
			model.ObjectiveType,
			`{"objective":99,"unit":"ms"}`,
			`["tag1","tag2"]`,
			model.Comment,
			model.CreateTime,
			model.UpdateTime,
		)

		modelQuery := interfaces.ObjectiveModelsQueryParams{
			NamePattern: "a",
			PaginationQueryParameters: interfaces.PaginationQueryParameters{
				Limit:     interfaces.MAX_LIMIT,
				Offset:    interfaces.MIN_OFFSET,
				Sort:      interfaces.OBJECTIVE_MODEL_SORT["update_time"],
				Direction: interfaces.DESC_DIRECTION,
			},
		}

		Convey("When query fails due to ToSql error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			_, err := oma.ListObjectiveModels(testCtx, modelQuery)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("When query fails", func() {
			expectedErr := errors.New("query error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, err := oma.ListObjectiveModels(testCtx, modelQuery)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("When scan fails", func() {
			rows := sqlmock.NewRows([]string{"f_model_id"}).AddRow("test-id")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := oma.ListObjectiveModels(testCtx, modelQuery)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("When query succeeds", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			models, err := oma.ListObjectiveModels(testCtx, modelQuery)
			So(err, ShouldBeNil)
			So(len(models), ShouldEqual, 1)
			So(models[0].ModelID, ShouldEqual, model.ModelID)
			So(models[0].ModelName, ShouldEqual, model.ModelName)
			So(models[0].ObjectiveType, ShouldEqual, model.ObjectiveType)
			So(models[0].Comment, ShouldEqual, model.Comment)
			So(models[0].UpdateTime, ShouldEqual, model.UpdateTime)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_ObjectiveModelAccess_GetObjectiveModelsTotal(t *testing.T) {
	Convey("Given an ObjectiveModelAccess", t, func() {
		appSetting := &common.AppSetting{}
		oma, smock := MockNewObjectiveModelAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT COUNT(f_model_id) FROM %s "+
			"WHERE instr(f_model_name, ?) > 0", OBJECTIVE_MODEL_TABLE_NAME)

		modelQuery := interfaces.ObjectiveModelsQueryParams{
			NamePattern: "a",
			PaginationQueryParameters: interfaces.PaginationQueryParameters{
				Limit:     interfaces.MAX_LIMIT,
				Offset:    interfaces.MIN_OFFSET,
				Sort:      interfaces.OBJECTIVE_MODEL_SORT["update_time"],
				Direction: interfaces.DESC_DIRECTION,
			},
		}

		Convey("When query fails due to ToSql error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			_, err := oma.GetObjectiveModelsTotal(testCtx, modelQuery)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("When query fails", func() {
			expectedErr := errors.New("query error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, err := oma.GetObjectiveModelsTotal(testCtx, modelQuery)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("When scan fails", func() {
			badRows := sqlmock.NewRows([]string{"count"}).AddRow("invalid")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(badRows)

			_, err := oma.GetObjectiveModelsTotal(testCtx, modelQuery)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("When query succeeds", func() {
			expectedTotal := int64(42)
			rows := sqlmock.NewRows([]string{"count"}).AddRow(expectedTotal)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			total, err := oma.GetObjectiveModelsTotal(testCtx, modelQuery)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, expectedTotal)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_ObjectiveModelAccess_GetObjectiveModelsByModelIDs(t *testing.T) {
	Convey("Test GetObjectiveModelsByModelIDs", t, func() {
		appSetting := &common.AppSetting{}
		oma, smock := MockNewObjectiveModelAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_model_id, f_model_name, f_objective_type, "+
			"f_objective_config, f_tags, f_comment, f_creator, f_creator_type, f_create_time, f_update_time "+
			"FROM %s WHERE f_model_id IN (?,?)", OBJECTIVE_MODEL_TABLE_NAME)

		modelIDs := []string{"test-id-1", "test-id-2"}

		Convey("When query fails due to ToSql error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			_, err := oma.GetObjectiveModelsByModelIDs(testCtx, modelIDs)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("When query fails", func() {
			expectedErr := errors.New("query error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, err := oma.GetObjectiveModelsByModelIDs(testCtx, modelIDs)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("When scan fails", func() {
			rows := sqlmock.NewRows([]string{"f_model_id", "f_model_name", "f_description",
				"f_creator", "f_creator_type", "f_create_time", "f_update_time"}).
				AddRow("test-id-1", "test-name-1", "desc1",
					interfaces.ADMIN_ID, interfaces.ADMIN_TYPE, "invalid", "invalid")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := oma.GetObjectiveModelsByModelIDs(testCtx, modelIDs)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("When query succeeds", func() {
			rows := sqlmock.NewRows([]string{"f_model_id",
				"f_model_name",
				"f_objective_type",
				"f_objective_config",
				"f_tags",
				"f_comment",
				"f_creator",
				"f_creator_type",
				"f_create_time",
				"f_update_time"}).
				AddRow("test-id-1", "test-name-1", "kpi", "{}", "", "", interfaces.ADMIN_ID, interfaces.ADMIN_TYPE, 1732945349330, 1732945349330).
				AddRow("test-id-2", "test-name-2", "slo", "{}", "", "", interfaces.ADMIN_ID, interfaces.ADMIN_TYPE, 1732945349330, 1732945349330)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			models, err := oma.GetObjectiveModelsByModelIDs(testCtx, modelIDs)
			So(err, ShouldBeNil)
			So(len(models), ShouldEqual, 2)
			So(models[0].ModelID, ShouldEqual, "test-id-1")
			So(models[1].ModelID, ShouldEqual, "test-id-2")

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_ObjectiveModelAccess_UpdateObjectiveModel(t *testing.T) {
	Convey("Test UpdateObjectiveModel", t, func() {
		appSetting := &common.AppSetting{}
		oma, smock := MockNewObjectiveModelAccess(appSetting)

		sqlStr := fmt.Sprintf("UPDATE %s SET f_comment = ?, f_model_name = ?, "+
			"f_objective_config = ?, f_objective_type = ?, f_tags = ?, f_update_time = ? "+
			"WHERE f_model_id = ?", OBJECTIVE_MODEL_TABLE_NAME)

		model := interfaces.ObjectiveModel{
			ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
				ModelID:         "test-id",
				ModelName:       "test-name",
				ObjectiveType:   "kpi",
				ObjectiveConfig: "{}",
				Tags:            []string{},
				Comment:         "",
			},
		}

		Convey("When ToSql fails", func() {
			expectedErr := errors.New("to sql error")
			patch := ApplyMethodReturn(sq.UpdateBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			smock.ExpectBegin()

			tx, _ := oma.db.Begin()
			err := oma.UpdateObjectiveModel(testCtx, tx, model)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("When exec fails", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("exec error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := oma.db.Begin()
			err := oma.UpdateObjectiveModel(testCtx, tx, model)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("When update succeeds", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			tx, _ := oma.db.Begin()
			err := oma.UpdateObjectiveModel(testCtx, tx, model)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_ObjectiveModelAccess_DeleteObjectiveModels(t *testing.T) {
	Convey("Test DeleteObjectiveModels", t, func() {
		appSetting := &common.AppSetting{}
		oma, smock := MockNewObjectiveModelAccess(appSetting)

		sqlStr := fmt.Sprintf("DELETE FROM %s WHERE f_model_id IN (?,?)", OBJECTIVE_MODEL_TABLE_NAME)

		modelIDs := []string{"test-id-1", "test-id-2"}

		Convey("When ToSql fails", func() {
			expectedErr := errors.New("to sql error")
			patch := ApplyMethodReturn(sq.DeleteBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			smock.ExpectBegin()

			tx, _ := oma.db.Begin()
			_, err := oma.DeleteObjectiveModels(testCtx, tx, modelIDs)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("When exec fails", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("exec error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := oma.db.Begin()
			_, err := oma.DeleteObjectiveModels(testCtx, tx, modelIDs)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("When delete succeeds", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 2))

			tx, _ := oma.db.Begin()
			_, err := oma.DeleteObjectiveModels(testCtx, tx, modelIDs)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}
