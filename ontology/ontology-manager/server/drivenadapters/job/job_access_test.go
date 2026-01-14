package job

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	"ontology-manager/common"
	"ontology-manager/interfaces"
)

var (
	testUpdateTime = int64(1735786555379)

	testCtx = context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage)

	testJobInfo = &interfaces.JobInfo{
		ID:      "job1",
		Name:    "Test Job",
		KNID:    "kn1",
		Branch:  "main",
		JobType: interfaces.JobTypeFull,
		Creator: interfaces.AccountInfo{
			ID:   "admin",
			Type: "admin",
		},
		CreateTime: testUpdateTime,
	}
)

func MockNewJobAccess(appSetting *common.AppSetting) (*jobAccess, sqlmock.Sqlmock) {
	db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	ja := &jobAccess{
		appSetting: appSetting,
		db:         db,
	}
	return ja, smock
}

func Test_jobAccess_CreateJob(t *testing.T) {
	Convey("test CreateJob\n", t, func() {
		appSetting := &common.AppSetting{}
		ja, smock := MockNewJobAccess(appSetting)

		sqlStr := fmt.Sprintf("INSERT INTO %s (f_id,f_name,f_kn_id,f_branch,f_job_type,f_job_concept_config,"+
			"f_state,f_state_detail,f_creator,f_creator_type,f_create_time) VALUES (?,?,?,?,?,?,?,?,?,?,?)", JOB_TABLE_NAME)

		Convey("CreateJob Success \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			tx, _ := ja.db.Begin()
			err := ja.CreateJob(testCtx, tx, testJobInfo)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CreateJob Exec sql error\n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("some error1")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := ja.db.Begin()
			err := ja.CreateJob(testCtx, tx, testJobInfo)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_jobAccess_GetJob(t *testing.T) {
	Convey("test GetJob\n", t, func() {
		appSetting := &common.AppSetting{}
		ja, smock := MockNewJobAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_id, f_name, f_kn_id, f_branch, f_job_type, f_job_concept_config, "+
			"f_state, f_state_detail, f_creator, f_creator_type, f_create_time, f_finish_time, f_time_cost "+
			"FROM %s WHERE f_id = ?", JOB_TABLE_NAME)

		jobID := "job1"
		jobConceptConfigStr, _ := sonic.MarshalString([]interfaces.ConceptConfig{})

		Convey("GetJob Success \n", func() {
			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_kn_id", "f_branch", "f_job_type",
				"f_job_concept_config", "f_state", "f_state_detail",
				"f_creator", "f_creator_type", "f_create_time",
				"f_finish_time", "f_time_cost",
			}).AddRow(
				jobID, "Test Job", "kn1", "main", interfaces.JobTypeFull,
				jobConceptConfigStr, interfaces.JobStateRunning, "",
				"admin", "admin", testUpdateTime,
				int64(2000), int64(1000),
			)

			smock.ExpectQuery(sqlStr).WithArgs(jobID).WillReturnRows(rows)

			job, err := ja.GetJob(testCtx, jobID)
			So(err, ShouldBeNil)
			So(job, ShouldNotBeNil)
			So(job.ID, ShouldEqual, jobID)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetJob Success no row \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs(jobID).WillReturnError(sql.ErrNoRows)

			job, err := ja.GetJob(testCtx, jobID)
			So(job, ShouldBeNil)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetJob Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs(jobID).WillReturnError(expectedErr)

			job, err := ja.GetJob(testCtx, jobID)
			So(job, ShouldBeNil)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_jobAccess_DeleteJobs(t *testing.T) {
	Convey("test DeleteJobs\n", t, func() {
		appSetting := &common.AppSetting{}
		ja, smock := MockNewJobAccess(appSetting)

		sqlStr := fmt.Sprintf("DELETE FROM %s WHERE f_id IN (?,?)", JOB_TABLE_NAME)

		jobIDs := []string{"job1", "job2"}

		Convey("DeleteJobs Success \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(0, 2))

			tx, _ := ja.db.Begin()
			err := ja.DeleteJobs(testCtx, tx, jobIDs)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteJobs null \n", func() {
			smock.ExpectBegin()

			tx, _ := ja.db.Begin()
			err := ja.DeleteJobs(testCtx, tx, []string{})
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteJobs Failed dbExec \n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("dbExec error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := ja.db.Begin()
			err := ja.DeleteJobs(testCtx, tx, jobIDs)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_jobAccess_DeleteTasks(t *testing.T) {
	Convey("test DeleteTasks\n", t, func() {
		appSetting := &common.AppSetting{}
		ja, smock := MockNewJobAccess(appSetting)

		sqlStr := fmt.Sprintf("DELETE FROM %s WHERE f_job_id IN (?,?)", TASK_TABLE_NAME)

		jobIDs := []string{"job1", "job2"}

		Convey("DeleteTasks Success \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(0, 5))

			tx, _ := ja.db.Begin()
			err := ja.DeleteTasks(testCtx, tx, jobIDs)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteTasks null \n", func() {
			smock.ExpectBegin()

			tx, _ := ja.db.Begin()
			err := ja.DeleteTasks(testCtx, tx, []string{})
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteTasks Failed dbExec \n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("dbExec error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := ja.db.Begin()
			err := ja.DeleteTasks(testCtx, tx, jobIDs)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_jobAccess_UpdateJobState(t *testing.T) {
	Convey("test UpdateJobState\n", t, func() {
		appSetting := &common.AppSetting{}
		ja, smock := MockNewJobAccess(appSetting)

		sqlStr := fmt.Sprintf("UPDATE %s SET f_state = ?, f_state_detail = ? WHERE f_id = ?", JOB_TABLE_NAME)

		jobID := "job1"
		stateInfo := interfaces.JobStateInfo{
			State:       interfaces.JobStateRunning,
			StateDetail: "Running",
		}

		Convey("UpdateJobState Success \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(0, 1))

			tx, _ := ja.db.Begin()
			err := ja.UpdateJobState(testCtx, tx, jobID, stateInfo)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateJobState Success with finish time \n", func() {
			stateInfo.FinishTime = 2000
			stateInfo.TimeCost = 1000
			sqlStrWithFinish := fmt.Sprintf("UPDATE %s SET f_state = ?, f_state_detail = ?, f_finish_time = ?, f_time_cost = ? WHERE f_id = ?", JOB_TABLE_NAME)

			smock.ExpectBegin()
			smock.ExpectExec(sqlStrWithFinish).WithArgs().WillReturnResult(sqlmock.NewResult(0, 1))

			tx, _ := ja.db.Begin()
			err := ja.UpdateJobState(testCtx, tx, jobID, stateInfo)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateJobState Failed \n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("some error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := ja.db.Begin()
			err := ja.UpdateJobState(testCtx, tx, jobID, stateInfo)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_jobAccess_GetJobs(t *testing.T) {
	Convey("test GetJobs\n", t, func() {
		appSetting := &common.AppSetting{}
		ja, smock := MockNewJobAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_id, f_name, f_kn_id, f_branch, f_job_type, f_job_concept_config, "+
			"f_state, f_state_detail, f_creator, f_creator_type, f_create_time, f_finish_time, f_time_cost "+
			"FROM %s WHERE f_id IN (?,?)", JOB_TABLE_NAME)

		jobIDs := []string{"job1", "job2"}
		jobConceptConfigStr, _ := sonic.MarshalString([]interfaces.ConceptConfig{})

		Convey("GetJobs Success \n", func() {
			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_kn_id", "f_branch", "f_job_type",
				"f_job_concept_config", "f_state", "f_state_detail",
				"f_creator", "f_creator_type", "f_create_time",
				"f_finish_time", "f_time_cost",
			}).AddRow(
				"job1", "Test Job 1", "kn1", "main", interfaces.JobTypeFull,
				jobConceptConfigStr, interfaces.JobStateRunning, "",
				"admin", "admin", testUpdateTime,
				int64(2000), int64(1000),
			).AddRow(
				"job2", "Test Job 2", "kn1", "main", interfaces.JobTypeFull,
				jobConceptConfigStr, interfaces.JobStateCompleted, "",
				"admin", "admin", testUpdateTime,
				int64(2000), int64(1000),
			)

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			jobs, err := ja.GetJobs(testCtx, jobIDs)
			So(err, ShouldBeNil)
			So(jobs, ShouldNotBeNil)
			So(len(jobs), ShouldEqual, 2)
			So(jobs["job1"], ShouldNotBeNil)
			So(jobs["job2"], ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetJobs Success no row \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(sqlmock.NewRows(nil))

			jobs, err := ja.GetJobs(testCtx, jobIDs)
			So(jobs, ShouldResemble, map[string]*interfaces.JobInfo{})
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetJobs null \n", func() {
			jobs, err := ja.GetJobs(testCtx, []string{})
			So(jobs, ShouldResemble, map[string]*interfaces.JobInfo{})
			So(err, ShouldBeNil)
		})

		Convey("GetJobs Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			jobs, err := ja.GetJobs(testCtx, jobIDs)
			So(jobs, ShouldBeNil)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_jobAccess_UpdateTaskState(t *testing.T) {
	Convey("test UpdateTaskState\n", t, func() {
		appSetting := &common.AppSetting{}
		ja, smock := MockNewJobAccess(appSetting)

		sqlStr := fmt.Sprintf("UPDATE %s SET f_state = ?, f_state_detail = ? WHERE f_id = ?", TASK_TABLE_NAME)

		taskID := "task1"
		stateInfo := interfaces.TaskStateInfo{
			State:       interfaces.TaskStateRunning,
			StateDetail: "Running",
		}

		Convey("UpdateTaskState Success \n", func() {
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(0, 1))

			err := ja.UpdateTaskState(testCtx, taskID, stateInfo)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateTaskState Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			err := ja.UpdateTaskState(testCtx, taskID, stateInfo)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_jobAccess_ListJobs(t *testing.T) {
	Convey("test ListJobs\n", t, func() {
		appSetting := &common.AppSetting{}
		ja, smock := MockNewJobAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_id, f_name, f_kn_id, f_branch, f_job_type, f_job_concept_config, "+
			"f_state, f_state_detail, f_creator, f_creator_type, f_create_time, f_finish_time, f_time_cost "+
			"FROM %s WHERE f_kn_id = ? ORDER BY f_update_time DESC", JOB_TABLE_NAME)

		queryParams := interfaces.JobsQueryParams{
			KNID:   "kn1",
			Branch: "main",
		}
		queryParams.Sort = "f_update_time"
		queryParams.Direction = "DESC"
		jobConceptConfigStr, _ := sonic.MarshalString([]interfaces.ConceptConfig{})

		Convey("ListJobs Success \n", func() {
			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_kn_id", "f_branch", "f_job_type",
				"f_job_concept_config", "f_state", "f_state_detail",
				"f_creator", "f_creator_type", "f_create_time",
				"f_finish_time", "f_time_cost",
			}).AddRow(
				"job1", "Test Job 1", "kn1", "main", interfaces.JobTypeFull,
				jobConceptConfigStr, interfaces.JobStateRunning, "",
				"admin", "admin", testUpdateTime,
				int64(2000), int64(1000),
			)

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			jobs, err := ja.ListJobs(testCtx, queryParams)
			So(err, ShouldBeNil)
			So(jobs, ShouldNotBeNil)
			So(len(jobs), ShouldEqual, 1)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListJobs Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			jobs, err := ja.ListJobs(testCtx, queryParams)
			So(jobs, ShouldBeNil)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_jobAccess_GetJobsTotal(t *testing.T) {
	Convey("test GetJobsTotal\n", t, func() {
		appSetting := &common.AppSetting{}
		ja, smock := MockNewJobAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE f_kn_id = ?", JOB_TABLE_NAME)

		queryParams := interfaces.JobsQueryParams{
			KNID:   "kn1",
			Branch: "main",
		}

		Convey("GetJobsTotal Success\n", func() {
			rows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(10)

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			total, err := ja.GetJobsTotal(testCtx, queryParams)
			So(total, ShouldEqual, 10)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetJobsTotal Failed  Query error\n", func() {
			expectedErr := errors.New("Query error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			total, err := ja.GetJobsTotal(testCtx, queryParams)
			So(total, ShouldEqual, 0)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_jobAccess_CreateTasks(t *testing.T) {
	Convey("test CreateTasks\n", t, func() {
		appSetting := &common.AppSetting{}
		ja, smock := MockNewJobAccess(appSetting)

		sqlStr := fmt.Sprintf("INSERT INTO %s (f_id,f_name,f_job_id,f_concept_type,f_concept_id,f_state,f_state_detail) VALUES (?,?,?,?,?,?,?)", TASK_TABLE_NAME)

		taskInfos := map[string]*interfaces.TaskInfo{
			"task1": {
				ID:          "task1",
				Name:        "Task 1",
				JobID:       "job1",
				ConceptType: "object_type",
				ConceptID:   "ot1",
				TaskStateInfo: interfaces.TaskStateInfo{
					State: interfaces.TaskStatePending,
				},
			},
		}

		Convey("CreateTasks Success \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			tx, _ := ja.db.Begin()
			err := ja.CreateTasks(testCtx, tx, taskInfos)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CreateTasks null \n", func() {
			smock.ExpectBegin()

			tx, _ := ja.db.Begin()
			err := ja.CreateTasks(testCtx, tx, map[string]*interfaces.TaskInfo{})
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CreateTasks Exec sql error\n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("some error1")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := ja.db.Begin()
			err := ja.CreateTasks(testCtx, tx, taskInfos)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_jobAccess_ListTasks(t *testing.T) {
	Convey("test ListTasks\n", t, func() {
		appSetting := &common.AppSetting{}
		ja, smock := MockNewJobAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_id, f_name, f_job_id, f_concept_type, f_concept_id, f_index, "+
			"f_doc_count, f_state, f_state_detail, f_start_time, f_finish_time, f_time_cost "+
			"FROM %s WHERE f_job_id = ? ORDER BY f_update_time DESC", TASK_TABLE_NAME)

		queryParams := interfaces.TasksQueryParams{
			JobID: "job1",
		}
		queryParams.Sort = "f_update_time"
		queryParams.Direction = "DESC"

		Convey("ListTasks Success \n", func() {
			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_job_id", "f_concept_type", "f_concept_id",
				"f_index", "f_doc_count", "f_state", "f_state_detail",
				"f_start_time", "f_finish_time", "f_time_cost",
			}).AddRow(
				"task1", "Task 1", "job1", "object_type", "ot1",
				"", int64(0), interfaces.TaskStateRunning, "",
				int64(1000), int64(2000), int64(1000),
			)

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			tasks, err := ja.ListTasks(testCtx, queryParams)
			So(err, ShouldBeNil)
			So(tasks, ShouldNotBeNil)
			So(len(tasks), ShouldEqual, 1)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListTasks Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			tasks, err := ja.ListTasks(testCtx, queryParams)
			So(tasks, ShouldBeNil)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_jobAccess_GetTasksTotal(t *testing.T) {
	Convey("test GetTasksTotal\n", t, func() {
		appSetting := &common.AppSetting{}
		ja, smock := MockNewJobAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT count(*) FROM %s WHERE f_job_id = ?", TASK_TABLE_NAME)

		queryParams := interfaces.TasksQueryParams{
			JobID: "job1",
		}

		Convey("GetTasksTotal Success\n", func() {
			rows := sqlmock.NewRows([]string{"count(*)"}).AddRow(5)

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			total, err := ja.GetTasksTotal(testCtx, queryParams)
			So(total, ShouldEqual, 5)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetTasksTotal Failed  Query error\n", func() {
			expectedErr := errors.New("Query error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			total, err := ja.GetTasksTotal(testCtx, queryParams)
			So(total, ShouldEqual, 0)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}
