// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package drivenadapters

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	libdb "github.com/kweaver-ai/kweaver-go-lib/db"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	sq "github.com/Masterminds/squirrel"
	"github.com/bytedance/sonic"

	"data-model-job/common"
	"data-model-job/interfaces"
)

const (
	DATA_MODEL_JOB_TABLE_NAME  = "t_data_model_job"
	DATA_VIEW_TABLE_NAME       = "t_data_view"
	METRIC_MODEL_TABLE_NAME    = "t_metric_model"
	METRIC_TASK_TABLE_NAME     = "t_metric_model_task"
	OBJECTIVE_MODEL_TABLE_NAME = "t_objective_model"
	EVENT_MODEL_TABLE_NAME     = "t_event_models"
	EVENT_TASK_TABLE_NAME      = "t_event_model_task"
)

var (
	jAccessOnce sync.Once
	jAccess     interfaces.JobAccess
)

type jobAccess struct {
	appSetting *common.AppSetting
	db         *sql.DB
}

func NewJobAccess(appSetting *common.AppSetting) interfaces.JobAccess {
	jAccessOnce.Do(func() {
		jAccess = &jobAccess{
			appSetting: appSetting,
			db:         libdb.NewDB(&appSetting.DBSetting),
		}
	})

	return jAccess
}

// 联表查询任务的信息
func (ja *jobAccess) ListViewJobs() ([]interfaces.JobInfo, error) {
	jobs := make([]interfaces.JobInfo, 0)
	sqlStr, args, err := sq.Select(
		"j.f_job_id",
		"j.f_job_type",
		"j.f_job_status",
		"j.f_job_status_details",
		"j.f_create_time",
		"j.f_update_time",
		"j.f_creator",
		"j.f_creator_type",
		"COALESCE(v.f_view_id, '')",
		"v.f_data_source",
		"COALESCE(v.f_field_scope, 0)",
		"v.f_fields",
		"v.f_filters").
		From(fmt.Sprintf("%s as j", DATA_MODEL_JOB_TABLE_NAME)).
		LeftJoin(fmt.Sprintf("%s as v on j.f_job_id = v.f_job_id", DATA_VIEW_TABLE_NAME)).
		Where(sq.Eq{"j.f_job_type": interfaces.JOB_TYPE_STREAM}).
		ToSql()

	if err != nil {
		errDetails := fmt.Sprintf("Generate 'list data view jobs' sql stmt failed, %v", err)
		logger.Error(errDetails)
		return nil, err
	}

	rows, err := ja.db.Query(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("List data view jobs failed, %v", err)
		logger.Error(errDetails)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var dataSourceBytes, fieldsBytes, condBytes []byte
		jobInfo := interfaces.JobInfo{}
		err := rows.Scan(
			&jobInfo.JobId,
			&jobInfo.JobType,
			&jobInfo.JobStatus,
			&jobInfo.JobStatusDetails,
			&jobInfo.CreateTime,
			&jobInfo.UpdateTime,
			&jobInfo.Creator.ID,
			&jobInfo.Creator.Type,
			&jobInfo.ViewId,
			// &jobInfo.ViewName,
			&dataSourceBytes,
			&jobInfo.FieldScope,
			&fieldsBytes,
			&condBytes,
		)
		if err != nil {
			errDetails := fmt.Sprintf("Row scan failed, err: %v", err)
			logger.Error(errDetails)
			return nil, err
		}

		// 跳过没有对应的视图对象信息的job
		if jobInfo.ViewId == "" {
			continue
		}

		// 反序列化
		err = sonic.Unmarshal([]byte(dataSourceBytes), &jobInfo.DataSource)
		if err != nil {
			errDetails := fmt.Sprintf("Unmarshal dataSource failed, %v", err)
			logger.Error(errDetails)
			return nil, err
		}

		err = sonic.Unmarshal([]byte(fieldsBytes), &jobInfo.Fields)
		if err != nil {
			errDetails := fmt.Sprintf("Unmarshal fields failed, %v", err)
			logger.Error(errDetails)
			return nil, err
		}

		err = sonic.Unmarshal([]byte(condBytes), &jobInfo.Condition)
		if err != nil {
			errDetails := fmt.Sprintf("Unmarshal condition failed, %v", err)
			logger.Error(errDetails)
			return nil, err
		}

		jobs = append(jobs, jobInfo)
	}

	return jobs, nil
}

func (ja *jobAccess) UpdateJobStatus(job interfaces.JobInfo) error {
	updateMap := map[string]any{
		"f_job_status":         job.JobStatus,
		"f_job_status_details": job.JobStatusDetails,
	}

	sqlStr, args, err := sq.Update(DATA_MODEL_JOB_TABLE_NAME).SetMap(updateMap).Where(sq.Eq{"f_job_id": job.JobId}).ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'update job status' sql stmt failed, %v", err)
		logger.Error(errDetails)
		return err
	}

	_, err = ja.db.Exec(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("Execute sql stmt for 'update job status' failed, %v", err)
		logger.Error(errDetails)
		return err
	}

	return nil
}

// 查询指标模型任务表
func (ja *jobAccess) ListMetricJobs() ([]interfaces.JobInfo, error) {
	jobs := make([]interfaces.JobInfo, 0)
	sqlStr, args, err := sq.Select(
		"task.f_task_id",
		"task.f_task_name",
		"task.f_comment",
		"task.f_module_type",
		"task.f_model_id",
		"model.f_measure_name",
		"task.f_schedule",
		"task.f_time_windows",
		"task.f_steps",
		"task.f_plan_time",
		"task.f_index_base",
		"task.f_retrace_duration",
		"task.f_creator",
		"task.f_creator_type",
	).
		From(fmt.Sprintf("%s as task", METRIC_TASK_TABLE_NAME)).
		Join(fmt.Sprintf("%s as model on task.f_model_id = model.f_model_id", METRIC_MODEL_TABLE_NAME)).
		Where(sq.Eq{"task.f_module_type": interfaces.MODULE_TYPE_METRIC_MODEL}).
		ToSql()

	if err != nil {
		errDetails := fmt.Sprintf("Generate 'list metric tasks' sql stmt failed, %v", err)
		logger.Error(errDetails)
		return nil, err
	}

	rows, err := ja.db.Query(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("List metric tasks failed, %v", err)
		logger.Error(errDetails)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			scheduleBytes []byte
			windowsBytes  []byte
			stepsBytes    []byte
			// planTimeBytes []byte
		)
		task := interfaces.MetricTask{}
		err := rows.Scan(
			&task.TaskID,
			&task.TaskName,
			&task.Comment,
			&task.ModuleType,
			&task.ModelID,
			&task.MeasureName,
			&scheduleBytes,
			&windowsBytes,
			&stepsBytes,
			&task.PlanTime,
			&task.IndexBase,
			&task.RetraceDuration,
			&task.Creator.ID,
			&task.Creator.Type,
		)
		if err != nil {
			errDetails := fmt.Sprintf("Row scan failed, err: %v", err)
			logger.Error(errDetails)
			return nil, err
		}

		// 处理任务信息
		// 2.0 反序列化schedule
		err = json.Unmarshal(scheduleBytes, &task.Schedule)
		if err != nil {
			logger.Errorf("Failed to unmarshal schedule after getting metric model, err: %v", err.Error())
			return jobs, err
		}
		// 2.1 反序列化时间窗口
		err = json.Unmarshal(windowsBytes, &task.TimeWindows)
		if err != nil {
			logger.Errorf("Failed to unmarshal time windows after getting metric model, err: %v", err.Error())
			return jobs, err
		}
		// 2.2 反序列化步长
		err = json.Unmarshal(stepsBytes, &task.Steps)
		if err != nil {
			logger.Errorf("Failed to unmarshal steps after getting metric model, err: %v", err.Error())
			return jobs, err
		}
		// 2.3 反序列化时间窗口对应的计划时间
		// err = json.Unmarshal(planTimeBytes, &task.PlanTime)
		// if err != nil {
		// 	logger.Errorf("Failed to unmarshal time windows after getting metric model, err: %v", err.Error())
		// 	return jobs, err
		// }

		job := interfaces.JobInfo{
			JobId:      task.TaskID,
			JobType:    interfaces.JOB_TYPE_SCHEDULE,
			ModuleType: interfaces.MODULE_TYPE_METRIC_MODEL,
			MetricTask: &task,
			Schedule:   task.Schedule,
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

// 查询目标模型任务表
func (ja *jobAccess) ListObjectiveJobs() ([]interfaces.JobInfo, error) {
	jobs := make([]interfaces.JobInfo, 0)
	sqlStr, args, err := sq.Select(
		"task.f_task_id",
		"task.f_task_name",
		"task.f_comment",
		"task.f_module_type",
		"task.f_model_id",
		"task.f_schedule",
		"task.f_time_windows",
		"task.f_steps",
		"task.f_plan_time",
		"task.f_index_base",
		"task.f_retrace_duration",
		"task.f_creator",
		"task.f_creator_type",
	).
		From(fmt.Sprintf("%s as task", METRIC_TASK_TABLE_NAME)).
		Join(fmt.Sprintf("%s as model on task.f_model_id = model.f_model_id", OBJECTIVE_MODEL_TABLE_NAME)).
		Where(sq.Eq{"task.f_module_type": interfaces.MODULE_TYPE_OBJECTIVE_MODEL}).
		ToSql()

	if err != nil {
		errDetails := fmt.Sprintf("Generate 'list objective tasks' sql stmt failed, %v", err)
		logger.Error(errDetails)
		return nil, err
	}

	rows, err := ja.db.Query(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("List objective tasks failed, %v", err)
		logger.Error(errDetails)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			scheduleBytes []byte
			windowsBytes  []byte
			stepsBytes    []byte
			// planTimeBytes []byte
		)
		task := interfaces.MetricTask{}
		err := rows.Scan(
			&task.TaskID,
			&task.TaskName,
			&task.Comment,
			&task.ModuleType,
			&task.ModelID,
			&scheduleBytes,
			&windowsBytes,
			&stepsBytes,
			&task.PlanTime,
			&task.IndexBase,
			&task.RetraceDuration,
			&task.Creator.ID,
			&task.Creator.Type,
		)
		if err != nil {
			errDetails := fmt.Sprintf("Row scan failed, err: %v", err)
			logger.Error(errDetails)
			return nil, err
		}

		// 处理任务信息
		// 2.0 反序列化schedule
		err = json.Unmarshal(scheduleBytes, &task.Schedule)
		if err != nil {
			logger.Errorf("Failed to unmarshal schedule after getting metric model, err: %v", err.Error())
			return jobs, err
		}
		// 2.1 反序列化时间窗口
		err = json.Unmarshal(windowsBytes, &task.TimeWindows)
		if err != nil {
			logger.Errorf("Failed to unmarshal time windows after getting metric model, err: %v", err.Error())
			return jobs, err
		}
		// 2.2 反序列化步长
		err = json.Unmarshal(stepsBytes, &task.Steps)
		if err != nil {
			logger.Errorf("Failed to unmarshal steps after getting metric model, err: %v", err.Error())
			return jobs, err
		}
		// 2.3 反序列化时间窗口对应的计划时间
		// err = json.Unmarshal(planTimeBytes, &task.PlanTime)
		// if err != nil {
		// 	logger.Errorf("Failed to unmarshal time windows after getting metric model, err: %v", err.Error())
		// 	return jobs, err
		// }

		job := interfaces.JobInfo{
			JobId:      task.TaskID,
			JobType:    interfaces.JOB_TYPE_SCHEDULE,
			ModuleType: interfaces.MODULE_TYPE_OBJECTIVE_MODEL,
			MetricTask: &task,
			Schedule:   task.Schedule,
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

// 查询事件模型任务表
func (ja *jobAccess) ListEventJobs() ([]interfaces.JobInfo, error) {
	jobs := make([]interfaces.JobInfo, 0)
	// 批处理的事件模型的任务都会写在task表中，只有状态是启用的才开启任务
	sqlStr, args, err := sq.Select(
		"task.f_task_id",
		"task.f_model_id",
		"task.f_schedule",
		"task.f_dispatch_config",
		"task.f_execute_parameter",
		"task.f_storage_config",
		"task.f_task_status",
		"task.f_status_update_time",
		"task.f_error_details",
		"task.f_schedule_sync_status",
		"task.f_downstream_dependent_task",
		"task.f_update_time",
		"task.f_creator",
		"task.f_creator_type",
	).
		From(fmt.Sprintf("%s as task", EVENT_TASK_TABLE_NAME)).
		Join(fmt.Sprintf("%s as model on task.f_model_id = model.f_event_model_id", EVENT_MODEL_TABLE_NAME)).
		Where(sq.Eq{"model.f_status": 1}).
		Where(sq.Eq{"model.f_is_active": 1}).
		ToSql()

	if err != nil {
		errDetails := fmt.Sprintf("Generate 'list event tasks' sql stmt failed, %v", err)
		logger.Error(errDetails)
		return nil, err
	}

	rows, err := ja.db.Query(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("List event tasks failed, %v", err)
		logger.Error(errDetails)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			scheduleBytes, storageConfigBytes, dispatchConfigBytes, executeParameterBytes []byte
			DownstreamDependentTask                                                       string
		)

		task := interfaces.EventTask{}
		err := rows.Scan(
			&task.TaskID,
			&task.ModelID,
			&scheduleBytes,
			&dispatchConfigBytes,
			&executeParameterBytes,
			&storageConfigBytes,
			&task.TaskStatus,
			&task.StatusUpdateTime,
			&task.ErrorDetails,
			&task.ScheduleSyncStatus,
			&DownstreamDependentTask,
			&task.UpdateTime,
			&task.Creator.ID,
			&task.Creator.Type,
		)
		if err != nil {
			errDetails := fmt.Sprintf("Row scan failed, err: %v", err)
			logger.Error(errDetails)
			return nil, err
		}

		// 处理任务信息
		// 2.0 反序列化schedule/dispatch_config/execute_parameter
		err = json.Unmarshal(scheduleBytes, &task.Schedule)
		if err != nil {
			logger.Errorf("Failed to unmarshal schedule after getting event model, err: %v", err.Error())
			return jobs, err
		}
		err = json.Unmarshal(storageConfigBytes, &task.StorageConfig)
		if err != nil {
			logger.Errorf("Failed to unmarshal storageConfig after getting event model, err: %v", err.Error())
			return jobs, err
		}
		err = json.Unmarshal(dispatchConfigBytes, &task.DispatchConfig)
		if err != nil {
			logger.Errorf("Failed to unmarshal dispatchConfig after getting event model, err: %v", err.Error())
			return jobs, err
		}
		err = json.Unmarshal(executeParameterBytes, &task.ExecuteParameter)
		if err != nil {
			logger.Errorf("Failed to unmarshal executeParameter after getting event model, err: %v", err.Error())
			return jobs, err
		}
		if DownstreamDependentTask != "" {
			task.DownstreamDependentTask = strings.Split(DownstreamDependentTask, ",")
		} else {
			task.DownstreamDependentTask = []string{}
		}

		job := interfaces.JobInfo{
			JobId:      task.TaskID,
			JobType:    interfaces.JOB_TYPE_SCHEDULE,
			ModuleType: interfaces.MODULE_TYPE_EVENT_MODEL,
			EventTask:  &task,
			Schedule:   task.Schedule,
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}
