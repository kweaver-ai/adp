// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_model_job

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	sq "github.com/Masterminds/squirrel"
	. "github.com/agiledragon/gomonkey/v2"
	"github.com/bytedance/sonic"
	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	rmock "github.com/kweaver-ai/kweaver-go-lib/rest/mock"
	. "github.com/smartystreets/goconvey/convey"

	"data-model/common"
	"data-model/interfaces"
)

var (
	testCtx = context.WithValue(context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage),
		interfaces.ACCOUNT_INFO_KEY, interfaces.AccountInfo{
			ID:   interfaces.ADMIN_ID,
			Type: interfaces.ADMIN_TYPE,
		})
)

func MockNewDataModelJobAccess(appSetting *common.AppSetting,
	httpClient rest.HTTPClient) (*dataModelJobAccess, sqlmock.Sqlmock) {
	db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	dmja := &dataModelJobAccess{
		db:         db,
		appSetting: appSetting,
		httpClient: httpClient,
	}
	return dmja, smock
}

func Test_DataModelJobAccess_CreateDataModelJob(t *testing.T) {
	Convey("Test CreateDataModelJob", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			IndexBaseUrl: "http://localhost:13012/api/data-model-job/v1",
		}
		httpClient := rmock.NewMockHTTPClient(mockCtrl)
		dmja, smock := MockNewDataModelJobAccess(appSetting, httpClient)

		sqlStr := fmt.Sprintf("INSERT INTO %s (f_job_id,f_job_type,f_job_config,"+
			"f_job_status,f_job_status_details,f_create_time,f_update_time,f_creator,f_creator_type) "+
			"VALUES (?,?,?,?,?,?,?,?,?)", DATA_MODEL_JOB_TABLE_NAME)

		job := &interfaces.JobInfo{
			JobID:     "1a",
			JobConfig: map[string]any{},
		}

		Convey("Create failed, caused by jobConfig marshal error", func() {
			expectedErr := errors.New("some error")
			patch := ApplyFuncReturn(sonic.Marshal, []byte{}, expectedErr)
			defer patch.Reset()

			smock.ExpectBegin()

			tx, _ := dmja.db.Begin()
			err := dmja.CreateDataModelJob(testCtx, tx, job)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Create failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.InsertBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			smock.ExpectBegin()

			tx, _ := dmja.db.Begin()
			err := dmja.CreateDataModelJob(testCtx, tx, job)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Create failed, caused by exec sql error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := dmja.db.Begin()
			err := dmja.CreateDataModelJob(testCtx, tx, job)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Create succeed", func() {
			// WithArgs()中不传参数, 应该代表任意参数下都会被mock.
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			tx, _ := dmja.db.Begin()
			err := dmja.CreateDataModelJob(testCtx, tx, job)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataModelJobAccess_DeleteDataModelJobs(t *testing.T) {
	Convey("Test DeleteDataModelJobs", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			IndexBaseUrl: "http://localhost:13012/api/data-model-job/v1",
		}
		httpClient := rmock.NewMockHTTPClient(mockCtrl)
		dmja, smock := MockNewDataModelJobAccess(appSetting, httpClient)

		sqlStr := fmt.Sprintf("DELETE FROM %s WHERE f_job_id IN (?,?)", DATA_MODEL_JOB_TABLE_NAME)

		Convey("Delete failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.DeleteBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			smock.ExpectBegin()

			tx, _ := dmja.db.Begin()
			err := dmja.DeleteDataModelJobs(testCtx, tx, []string{"1a", "2b"})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Delete failed, caused by exec sql error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := dmja.db.Begin()
			err := dmja.DeleteDataModelJobs(testCtx, tx, []string{"1a", "2b"})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Delete succeed", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 2))

			tx, _ := dmja.db.Begin()
			err := dmja.DeleteDataModelJobs(testCtx, tx, []string{"1a", "2b"})
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DataModelJobAccess_StartJob(t *testing.T) {
	Convey("Test StartJob", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			IndexBaseUrl: "http://localhost:13012/api/data-model-job/v1",
		}
		httpClient := rmock.NewMockHTTPClient(mockCtrl)
		dmja, _ := MockNewDataModelJobAccess(appSetting, httpClient)

		resp := map[string]string{"id": "1"}
		okResp, _ := sonic.Marshal(resp)

		job := &interfaces.DataModelJobCfg{
			JobID: "1a",
		}

		Convey("failed, caused by http error", func() {
			httpClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				http.StatusInternalServerError, nil, fmt.Errorf("method failed"))

			err := dmja.StartJob(testCtx, job)
			So(err, ShouldNotBeNil)
		})

		Convey("failed, caused by status != 201", func() {
			httpClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				http.StatusBadRequest, okResp, nil)

			err := dmja.StartJob(testCtx, job)
			So(err, ShouldNotBeNil)
		})

		Convey("failed, caused by unmarshal error", func() {
			httpClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				http.StatusBadRequest, okResp, nil)

			patch := ApplyFuncReturn(sonic.Unmarshal, errors.New("error"))
			defer patch.Reset()

			err := dmja.StartJob(testCtx, job)
			So(err.Error(), ShouldEqual, "error")
		})

		Convey("success", func() {
			httpClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				http.StatusCreated, okResp, nil)

			err := dmja.StartJob(testCtx, job)
			So(err, ShouldBeNil)
		})
	})
}

func Test_DataModelJobAccess_StopJob(t *testing.T) {
	Convey("Test StopJob", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		appSetting := &common.AppSetting{
			IndexBaseUrl: "http://localhost:13012/api/data-model-job/v1",
		}
		httpClient := rmock.NewMockHTTPClient(mockCtrl)
		dmja, _ := MockNewDataModelJobAccess(appSetting, httpClient)

		jobIDs := []string{"1a", "2b"}

		Convey("failed, caused by http error", func() {
			httpClient.EXPECT().DeleteNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any()).Return(
				http.StatusInternalServerError, nil, fmt.Errorf("method failed"))

			err := dmja.StopJobs(testCtx, jobIDs)
			So(err, ShouldNotBeNil)
		})

		Convey("failed, caused by status != 204", func() {
			httpClient.EXPECT().DeleteNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any()).Return(
				http.StatusBadRequest, nil, nil)

			err := dmja.StopJobs(testCtx, jobIDs)
			So(err, ShouldNotBeNil)
		})

		Convey("failed, caused by unmarshal error", func() {
			httpClient.EXPECT().DeleteNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any()).Return(
				http.StatusBadRequest, nil, nil)

			patch := ApplyFuncReturn(sonic.Unmarshal, errors.New("error"))
			defer patch.Reset()

			err := dmja.StopJobs(testCtx, jobIDs)
			So(err.Error(), ShouldEqual, "error")
		})

		Convey("success", func() {
			httpClient.EXPECT().DeleteNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any()).Return(
				http.StatusNoContent, nil, nil)

			err := dmja.StopJobs(testCtx, jobIDs)
			So(err, ShouldBeNil)
		})
	})
}

func Test_DataModelJobAccess_UpdateJob(t *testing.T) {
	Convey("Test UpdateJob", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		appSetting := &common.AppSetting{
			IndexBaseUrl: "http://localhost:13012/api/data-model-job/v1",
		}
		httpClient := rmock.NewMockHTTPClient(mockCtrl)
		dmja, _ := MockNewDataModelJobAccess(appSetting, httpClient)

		job := &interfaces.DataModelJobCfg{
			JobID: "1a",
		}

		Convey("failed, caused by http error", func() {
			httpClient.EXPECT().PutNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				http.StatusInternalServerError, nil, fmt.Errorf("method failed"))

			err := dmja.UpdateJob(testCtx, job)
			So(err, ShouldNotBeNil)
		})

		Convey("failed, caused by status != 204", func() {
			httpClient.EXPECT().PutNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				http.StatusBadRequest, nil, nil)

			err := dmja.UpdateJob(testCtx, job)

			So(err, ShouldNotBeNil)
		})

		Convey("failed, caused by unmarshal error", func() {
			httpClient.EXPECT().PutNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				http.StatusBadRequest, nil, nil)

			patch := ApplyFuncReturn(sonic.Unmarshal, errors.New("error"))
			defer patch.Reset()

			err := dmja.UpdateJob(testCtx, job)
			So(err.Error(), ShouldEqual, "error")
		})

		Convey("success", func() {
			httpClient.EXPECT().PutNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				http.StatusNoContent, nil, nil)

			err := dmja.UpdateJob(testCtx, job)
			So(err, ShouldBeNil)
		})
	})
}
