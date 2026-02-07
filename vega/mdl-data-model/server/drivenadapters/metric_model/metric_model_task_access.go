// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package metric_model

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	libdb "github.com/kweaver-ai/kweaver-go-lib/db"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"data-model/common"
	"data-model/interfaces"
)

const (
	METRIC_MODEL_TASK_TABLE_NAME = "t_metric_model_task"
)

var (
	mmtAccessOnce sync.Once
	mmtAccess     interfaces.MetricModelTaskAccess
)

type metricModelTaskAccess struct {
	appSetting *common.AppSetting
	db         *sql.DB
}

func NewMetricModelTaskAccess(appSetting *common.AppSetting) interfaces.MetricModelTaskAccess {
	mmtAccessOnce.Do(func() {
		mmtAccess = &metricModelTaskAccess{
			appSetting: appSetting,
			db:         libdb.NewDB(&appSetting.DBSetting),
		}
	})

	return mmtAccess
}

func (mmta *metricModelTaskAccess) CreateMetricTask(ctx context.Context, tx *sql.Tx, task interfaces.MetricTask) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Insert into metric model task", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()
	// 1 初始化sql语句
	sqlBuilder := sq.Insert(METRIC_MODEL_TASK_TABLE_NAME).
		Columns(
			"f_task_id",
			"f_task_name",
			"f_module_type",
			"f_model_id",
			"f_schedule",
			"f_time_windows",
			"f_steps",
			"f_index_base",
			"f_retrace_duration",
			"f_schedule_sync_status",
			"f_comment",
			"f_update_time",
			"f_plan_time",
			"f_creator",
			"f_creator_type",
		)

	// 2.0 反序列化schedule
	scheduleBytes, err := sonic.Marshal(task.Schedule)
	if err != nil {
		logger.Errorf("Failed to marshal schedule, err: %v", err.Error())
		return err
	}

	// 2.1 反序列化时间窗口
	windowsBytes, err := sonic.Marshal(task.TimeWindows)
	if err != nil {
		logger.Errorf("Failed to marshal time windows, err: %v", err.Error())
		return err
	}

	// 2.2 反序列化步长
	stepsBytes, err := sonic.Marshal(task.Steps)
	if err != nil {
		logger.Errorf("Failed to marshal steps, err: %v", err.Error())
		return err
	}

	// 2.3 计算计划执行时间，now 或者根据追溯时长做个减法计算而得
	// planTimeBytes, err := sonic.Marshal(task.PlanTime)
	// if err != nil {
	// 	logger.Errorf("Failed to marshal time windows, err: %v", err.Error())
	// 	return err
	// }

	// 2.3 追加参数
	sqlBuilder = sqlBuilder.Values(
		task.TaskID,
		task.TaskName,
		task.ModuleType,
		task.ModelID,
		scheduleBytes,
		windowsBytes,
		stepsBytes,
		task.IndexBase,
		task.RetraceDuration,
		task.ScheduleSyncStatus,
		task.Comment,
		task.UpdateTime,
		task.PlanTime,
		task.Creator.ID,
		task.Creator.Type,
	)

	// 3. 生成完整的sql语句和参数列表
	sqlStr, args, err := sqlBuilder.ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of insert metric task, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of insert metric task, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")

		return err
	}
	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("创建指标模型的 sql 语句: %s", sqlStr))

	// 执行批量insert
	_, err = tx.Exec(sqlStr, args...)
	if err != nil {
		logger.Errorf("insert data error: %v\n", err)

		span.SetStatus(codes.Error, "Insert data error")
		o11y.Error(ctx, fmt.Sprintf("Insert data error: %v ", err))
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 获取指标模型下的任务id列表
func (mmta *metricModelTaskAccess) GetMetricTaskIDsByModelIDs(ctx context.Context, modelIDs []string) ([]string, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Select metric model tasks", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	taskIDs := make([]string, 0)

	sqlStr, vals, err := sq.Select("f_task_id").
		From(METRIC_MODEL_TASK_TABLE_NAME).
		Where(sq.Eq{"f_model_id": modelIDs}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of select tasks, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of select tasks, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return taskIDs, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("查询指标模型任务id列表的 sql 语句: %s; modelID: %v", sqlStr, modelIDs))
	rows, err := mmta.db.Query(sqlStr, vals...)
	if err != nil {
		logger.Errorf("list data error: %v\n", err)
		span.SetStatus(codes.Error, "List data error")
		o11y.Error(ctx, fmt.Sprintf("List data error: %v", err))

		return taskIDs, err
	}
	defer rows.Close()

	for rows.Next() {
		var taskID string
		err := rows.Scan(&taskID)
		if err != nil {
			logger.Errorf("row scan failed, err: %v \n", err)
			span.SetStatus(codes.Error, "Row scan error")
			o11y.Error(ctx, fmt.Sprintf("Row scan error: %v", err))
			return taskIDs, err
		}

		taskIDs = append(taskIDs, taskID)
	}

	span.SetStatus(codes.Ok, "")
	return taskIDs, nil
}

// 更新任务
func (mmta *metricModelTaskAccess) UpdateMetricTask(ctx context.Context, tx *sql.Tx, task interfaces.MetricTask) error {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("Update metric model task[%s]", task.TaskID), trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	// 2.0 反序列化schedule
	scheduleBytes, err := sonic.Marshal(task.Schedule)
	if err != nil {
		logger.Errorf("Failed to marshal schedule, err: %v", err.Error())
		return err
	}
	// 2.1 反序列化时间窗口
	windowsBytes, err := sonic.Marshal(task.TimeWindows)
	if err != nil {
		logger.Errorf("Failed to marshal time windows, err: %v", err.Error())
		return err
	}

	// 2.2 反序列化步长
	stepsBytes, err := sonic.Marshal(task.Steps)
	if err != nil {
		logger.Errorf("Failed to marshal steps, err: %v", err.Error())
		return err
	}

	data := map[string]interface{}{
		"f_task_name":    task.TaskName,
		"f_module_type":  task.ModuleType,
		"f_model_id":     task.ModelID,
		"f_schedule":     scheduleBytes,
		"f_time_windows": windowsBytes,
		"f_steps":        stepsBytes,
		"f_index_base":   task.IndexBase,
		// "f_retrace_duration": task.RetraceDuration, // 追溯时长不修改
		"f_schedule_sync_status": task.ScheduleSyncStatus,
		"f_comment":              task.Comment,
		"f_update_time":          task.UpdateTime,
	}
	sqlStr, vals, err := sq.Update(METRIC_MODEL_TASK_TABLE_NAME).
		SetMap(data).
		Where(sq.Eq{"f_task_id": task.TaskID}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of update metric model task by task_id, error: %s", err.Error())

		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of update metric model task by task_id, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return err
	}
	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("修改指标模型任务的 sql 语句: %s", sqlStr))

	ret, err := tx.Exec(sqlStr, vals...)
	if err != nil {
		logger.Errorf("update metric model task error: %v\n", err)
		span.SetStatus(codes.Error, "Update data error")
		o11y.Error(ctx, fmt.Sprintf("Update data error: %v ", err))
		return err
	}

	//sql语句影响的行数
	RowsAffected, err := ret.RowsAffected()
	if err != nil {
		logger.Errorf("Get RowsAffected error: %v\n", err)
		o11y.Warn(ctx, fmt.Sprintf("Get RowsAffected error: %v ", err))
	}

	if RowsAffected > 1 {
		// 影响行数不等于1不报错，更新操作已经发生
		logger.Errorf("UPDATE %s RowsAffected more than 1, RowsAffected is %d, metricModel is %v",
			task.TaskID, RowsAffected, task)

		o11y.Warn(ctx, fmt.Sprintf("Update %s RowsAffected more than 1, RowsAffected is %d, metricModel is %v",
			task.TaskID, RowsAffected, task))
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 按任务id批量获取任务信息
func (mmta *metricModelTaskAccess) GetMetricTasksByTaskIDs(ctx context.Context, taskIDs []string) ([]interfaces.MetricTask, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Select metric model tasks", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	tasks := make([]interfaces.MetricTask, 0)

	sqlStr, vals, err := sq.Select(
		"f_task_id",
		"f_task_name",
		"f_module_type",
		"f_model_id",
		"f_schedule",
		"f_time_windows",
		"f_steps",
		"f_index_base",
		"f_retrace_duration",
		"f_schedule_sync_status",
		"f_comment",
		"f_update_time",
		"f_plan_time").
		From(METRIC_MODEL_TASK_TABLE_NAME).
		Where(sq.Eq{"f_task_id": taskIDs}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of select tasks, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of select tasks, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return tasks, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("查询指标模型任务id列表的 sql 语句: %s; taskIDs: %v", sqlStr, taskIDs))
	rows, err := mmta.db.Query(sqlStr, vals...)
	if err != nil {
		logger.Errorf("list data error: %v\n", err)
		span.SetStatus(codes.Error, "List data error")
		o11y.Error(ctx, fmt.Sprintf("List data error: %v", err))

		return tasks, err
	}
	defer rows.Close()

	for rows.Next() {
		task := interfaces.MetricTask{}
		var (
			scheduleBytes []byte
			windowsBytes  []byte
			stepsBytes    []byte
			// planTimeBytes []byte
		)
		err := rows.Scan(
			&task.TaskID,
			&task.TaskName,
			&task.ModuleType,
			&task.ModelID,
			&scheduleBytes,
			&windowsBytes,
			&stepsBytes,
			&task.IndexBase,
			&task.RetraceDuration,
			&task.ScheduleSyncStatus,
			&task.Comment,
			&task.UpdateTime,
			&task.PlanTime,
		)
		if err != nil {
			logger.Errorf("row scan failed, err: %v \n", err)
			span.SetStatus(codes.Error, "Row scan error")
			o11y.Error(ctx, fmt.Sprintf("Row scan error: %v", err))
			return tasks, err
		}

		// 处理任务信息
		// 2.0 反序列化schedule
		err = sonic.Unmarshal(scheduleBytes, &task.Schedule)
		if err != nil {
			logger.Errorf("Failed to unmarshal schedule after getting metric model, err: %v", err.Error())
			return tasks, err
		}
		// 2.1 反序列化时间窗口
		err = sonic.Unmarshal(windowsBytes, &task.TimeWindows)
		if err != nil {
			logger.Errorf("Failed to unmarshal time windows after getting metric model, err: %v", err.Error())
			return tasks, err
		}
		// 2.2 反序列化步长
		err = sonic.Unmarshal(stepsBytes, &task.Steps)
		if err != nil {
			logger.Errorf("Failed to unmarshal steps after getting metric model, err: %v", err.Error())
			return tasks, err
		}
		// 2.3 反序列化时间窗口对应的计划时间
		// err = sonic.Unmarshal(planTimeBytes, &task.PlanTime)
		// if err != nil {
		// 	logger.Errorf("Failed to unmarshal time windows after getting metric model, err: %v", err.Error())
		// 	return tasks, err
		// }

		tasks = append(tasks, task)
	}

	span.SetStatus(codes.Ok, "")
	return tasks, nil
}

// 按模型id批量获取任务信息
func (mmta *metricModelTaskAccess) GetMetricTasksByModelIDs(ctx context.Context, modelIDs []string) (map[string]interfaces.MetricTask, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Select metric model tasks", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	modelTaskMap := make(map[string]interfaces.MetricTask, 0)

	sqlStr, vals, err := sq.Select(
		"f_task_id",
		"f_task_name",
		"f_module_type",
		"f_model_id",
		"f_schedule",
		"f_time_windows",
		"f_steps",
		"f_index_base",
		"f_retrace_duration",
		"f_schedule_sync_status",
		"f_execute_status",
		"f_comment",
		"f_creator",
		"f_creator_type",
		"f_update_time",
		"f_plan_time").
		From(METRIC_MODEL_TASK_TABLE_NAME).
		Where(sq.Eq{"f_model_id": modelIDs}).
		// Where(sq.NotEq{"f_schedule_sync_status": interfaces.SCHEDULE_SYNC_STATUS_DELETE}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of select tasks, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of select tasks, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return modelTaskMap, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("查询指标模型任务id列表的 sql 语句: %s; modelIDs: %v", sqlStr, modelIDs))
	rows, err := mmta.db.Query(sqlStr, vals...)
	if err != nil {
		logger.Errorf("list data error: %v\n", err)
		span.SetStatus(codes.Error, "List data error")
		o11y.Error(ctx, fmt.Sprintf("List data error: %v", err))

		return modelTaskMap, err
	}
	defer rows.Close()

	for rows.Next() {
		task := interfaces.MetricTask{}
		var (
			scheduleBytes []byte
			windowsBytes  []byte
			stepsBytes    []byte
			// planTimeBytes []byte
		)
		err := rows.Scan(
			&task.TaskID,
			&task.TaskName,
			&task.ModuleType,
			&task.ModelID,
			&scheduleBytes,
			&windowsBytes,
			&stepsBytes,
			&task.IndexBase,
			&task.RetraceDuration,
			&task.ScheduleSyncStatus,
			&task.ExeccuteStatus,
			&task.Comment,
			&task.Creator.ID,
			&task.Creator.Type,
			&task.UpdateTime,
			&task.PlanTime,
		)
		if err != nil {
			logger.Errorf("row scan failed, err: %v \n", err)
			span.SetStatus(codes.Error, "Row scan error")
			o11y.Error(ctx, fmt.Sprintf("Row scan error: %v", err))
			return modelTaskMap, err
		}

		// 处理任务信息
		// 2.0 反序列化schedule
		err = sonic.Unmarshal(scheduleBytes, &task.Schedule)
		if err != nil {
			logger.Errorf("Failed to unmarshal schedule after getting metric model, err: %v", err.Error())
			return modelTaskMap, err
		}
		// 2.1 反序列化时间窗口
		err = sonic.Unmarshal(windowsBytes, &task.TimeWindows)
		if err != nil {
			logger.Errorf("Failed to unmarshal time windows after getting metric model, err: %v", err.Error())
			return modelTaskMap, err
		}
		// 2.2 反序列化步长
		err = sonic.Unmarshal(stepsBytes, &task.Steps)
		if err != nil {
			logger.Errorf("Failed to unmarshal steps after getting metric model, err: %v", err.Error())
			return modelTaskMap, err
		}
		// 2.3 反序列化时间窗口对应的计划时间
		// err = sonic.Unmarshal(planTimeBytes, &task.PlanTime)
		// if err != nil {
		// 	logger.Errorf("Failed to unmarshal time windows after getting metric model, err: %v", err.Error())
		// 	return modelTaskMap, err
		// }

		modelTaskMap[task.ModelID] = task
	}

	span.SetStatus(codes.Ok, "")
	return modelTaskMap, nil
}

// 更新任务状态为完成，更新调度id
func (mmta *metricModelTaskAccess) UpdateMetricTaskStatusInFinish(ctx context.Context, task interfaces.MetricTask) error {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("Update metric model task[%s]", task.TaskID), trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	data := map[string]interface{}{
		"f_schedule_sync_status": interfaces.SCHEDULE_SYNC_STATUS_FINISH,
		"f_update_time":          time.Now().UnixMilli(),
	}
	sqlStr, vals, err := sq.Update(METRIC_MODEL_TASK_TABLE_NAME).
		SetMap(data).
		Where(sq.Eq{"f_task_id": task.TaskID}).
		Where(sq.LtOrEq{"f_update_time": task.UpdateTime}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of update metric model task by task_id, error: %s", err.Error())

		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of update metric model task by task_id, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return err
	}
	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("修改指标模型任务的 sql 语句: %s", sqlStr))

	ret, err := mmta.db.Exec(sqlStr, vals...)
	if err != nil {
		logger.Errorf("update metric model task error: %v\n", err)
		span.SetStatus(codes.Error, "Update data error")
		o11y.Error(ctx, fmt.Sprintf("Update data error: %v ", err))
		return err
	}

	//sql语句影响的行数
	RowsAffected, err := ret.RowsAffected()
	if err != nil {
		logger.Errorf("Get RowsAffected error: %v\n", err)
		o11y.Warn(ctx, fmt.Sprintf("Get RowsAffected error: %v ", err))
	}

	if RowsAffected > 1 {
		// 影响行数不等于1不报错，更新操作已经发生
		logger.Errorf("UPDATE %s RowsAffected more than 1, RowsAffected is %d, metricModel is %v",
			task.TaskID, RowsAffected, task)

		o11y.Warn(ctx, fmt.Sprintf("Update %s RowsAffected more than 1, RowsAffected is %d, metricModel is %v",
			task.TaskID, RowsAffected, task))
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 更新任务的计划时间
func (mmta *metricModelTaskAccess) UpdateMetricTaskAttributes(ctx context.Context, task interfaces.MetricTask) error {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("Update metric model task[%s]", task.TaskID), trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()
	data := make(map[string]interface{})
	// 计划时间
	if task.PlanTime != 0 {
		// 2.2 反序列化时间窗口
		// planTimeBytes, err := sonic.Marshal(task.PlanTime)
		// if err != nil {
		// 	logger.Errorf("Failed to marshal time windows, err: %v", err.Error())
		// 	return err
		// }
		data["f_plan_time"] = task.PlanTime
	}

	// 任务状态
	if task.ExeccuteStatus != 0 {
		data["f_execute_status"] = task.ExeccuteStatus
	}

	sqlStr, vals, err := sq.Update(METRIC_MODEL_TASK_TABLE_NAME).
		SetMap(data).
		Where(sq.Eq{"f_task_id": task.TaskID}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of update metric model task by task_id, error: %s", err.Error())

		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of update metric model task by task_id, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return err
	}
	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("修改指标模型任务的 sql 语句: %s", sqlStr))

	ret, err := mmta.db.Exec(sqlStr, vals...)
	if err != nil {
		logger.Errorf("update metric model task error: %v\n", err)
		span.SetStatus(codes.Error, "Update data error")
		o11y.Error(ctx, fmt.Sprintf("Update data error: %v ", err))
		return err
	}

	//sql语句影响的行数
	RowsAffected, err := ret.RowsAffected()
	if err != nil {
		logger.Errorf("Get RowsAffected error: %v\n", err)
		o11y.Warn(ctx, fmt.Sprintf("Get RowsAffected error: %v ", err))
	}

	if RowsAffected > 1 {
		// 影响行数不等于1不报错，更新操作已经发生
		logger.Errorf("UPDATE %s RowsAffected more than 1, RowsAffected is %d, metricModel is %v",
			task.TaskID, RowsAffected, task)

		o11y.Warn(ctx, fmt.Sprintf("Update %s RowsAffected more than 1, RowsAffected is %d, metricModel is %v",
			task.TaskID, RowsAffected, task))
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 物理删除任务
func (mmta *metricModelTaskAccess) DeleteMetricTaskByTaskIDs(ctx context.Context, tx *sql.Tx, taskIDs []string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete metric model task from db", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("task_ids").String(fmt.Sprintf("%v", taskIDs)))
	defer span.End()

	sqlStr, vals, err := sq.Delete(METRIC_MODEL_TASK_TABLE_NAME).
		Where(sq.Eq{"f_task_id": taskIDs}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of delete model task by f_task_id, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of delete model by f_task_id, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("删除指标模型的持久化任务的 sql 语句: %s; 删除的任务id: %v", sqlStr, taskIDs))

	ret, err := tx.Exec(sqlStr, vals...)
	if err != nil {
		logger.Errorf("delete data error: %v\n", err)
		span.SetStatus(codes.Error, "Delete data error")
		o11y.Error(ctx, fmt.Sprintf("Delete data error: %v ", err))
		return err
	}

	//sql语句影响的行数
	RowsAffected, err := ret.RowsAffected()
	if err != nil {
		logger.Errorf("Get RowsAffected error: %v\n", err)
		span.SetStatus(codes.Error, "Get RowsAffected error")
		o11y.Warn(ctx, fmt.Sprintf("Get RowsAffected error: %v ", err))
	}
	logger.Infof("RowsAffected: %d", RowsAffected)

	span.SetStatus(codes.Ok, "")
	return nil
}

func (mmta *metricModelTaskAccess) CheckMetricModelTaskExistByName(ctx context.Context,
	task interfaces.MetricTask, deleteTaskIDs []string) (bool, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "Query metric model task", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()
	//查询
	sqlBuilder := sq.Select(
		"f_task_id").
		From(METRIC_MODEL_TASK_TABLE_NAME).
		Where(sq.Eq{"f_model_id": task.ModelID}).
		Where(sq.Eq{"f_task_name": task.TaskName})

	if len(deleteTaskIDs) > 0 {
		sqlBuilder = sqlBuilder.Where(sq.NotEq{"f_task_id": deleteTaskIDs})
	}

	sqlStr, vals, err := sqlBuilder.ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of get model id by name, error: %s", err.Error())

		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of get model id by name, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")

		return false, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("获取指标模型任务信息的 sql 语句: %s", sqlStr))

	taskInfo := interfaces.MetricTask{}
	err = mmta.db.QueryRow(sqlStr, vals...).Scan(
		&taskInfo.TaskID,
	)
	if err == sql.ErrNoRows {
		span.SetAttributes(attr.Key("no_rows").Bool(true))
		span.SetStatus(codes.Ok, "")

		return false, nil
	} else if err != nil {
		logger.Errorf("row scan failed, err: %v\n", err)

		o11y.Error(ctx, fmt.Sprintf("Row scan failed, err: %v", err))
		span.SetStatus(codes.Error, "Row scan failed ")

		return false, err
	}

	span.SetStatus(codes.Ok, "")
	return true, nil
}
