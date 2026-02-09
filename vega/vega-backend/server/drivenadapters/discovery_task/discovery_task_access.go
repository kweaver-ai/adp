// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package discovery_task provides DiscoveryTask data access operations.
package discovery_task

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

	"vega-backend/common"
	"vega-backend/interfaces"
)

const (
	DISCOVERY_TASK_TABLE_NAME = "t_discovery_task"
)

var (
	dtAccessOnce sync.Once
	dtAccess     interfaces.DiscoveryTaskAccess
)

type discoveryTaskAccess struct {
	appSetting *common.AppSetting
	db         *sql.DB
}

// NewDiscoveryTaskAccess creates a new DiscoveryTaskAccess.
func NewDiscoveryTaskAccess(appSetting *common.AppSetting) interfaces.DiscoveryTaskAccess {
	dtAccessOnce.Do(func() {
		dtAccess = &discoveryTaskAccess{
			appSetting: appSetting,
			db:         libdb.NewDB(&appSetting.DBSetting),
		}
	})
	return dtAccess
}

// Create creates a new DiscoveryTask.
func (da *discoveryTaskAccess) Create(ctx context.Context, task *interfaces.DiscoveryTask) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Insert into discovery_task",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))

	sqlStr, vals, err := sq.Insert(DISCOVERY_TASK_TABLE_NAME).
		Columns(
			"f_id",
			"f_catalog_id",
			"f_trigger_type",
			"f_status",
			"f_progress",
			"f_message",
			"f_start_time",
			"f_finish_time",
			"f_result",
			"f_creator",
			"f_creator_type",
			"f_create_time",
		).
		Values(
			task.ID,
			task.CatalogID,
			task.TriggerType,
			task.Status,
			task.Progress,
			task.Message,
			task.StartTime,
			task.FinishTime,
			"", // result initially empty
			task.Creator.ID,
			task.Creator.Type,
			task.CreateTime,
		).ToSql()
	if err != nil {
		logger.Errorf("Failed to build insert discovery_task sql: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Failed to build insert discovery_task sql: %v", err))
		span.SetStatus(codes.Error, "Build sql failed")
		return err
	}

	o11y.Info(ctx, fmt.Sprintf("Insert discovery_task SQL: %s", sqlStr))

	_, err = da.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		logger.Errorf("Insert discovery_task failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Insert discovery_task failed: %v", err))
		span.SetStatus(codes.Error, "Insert failed")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// GetByID retrieves a DiscoveryTask by ID.
func (da *discoveryTaskAccess) GetByID(ctx context.Context, id string) (*interfaces.DiscoveryTask, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Query discovery_task by ID",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(attr.Key("task_id").String(id))

	sqlStr, vals, err := sq.Select(
		"f_id",
		"f_catalog_id",
		"f_trigger_type",
		"f_status",
		"f_progress",
		"f_message",
		"f_start_time",
		"f_finish_time",
		"f_result",
		"f_creator",
		"f_creator_type",
		"f_create_time",
	).From(DISCOVERY_TASK_TABLE_NAME).
		Where(sq.Eq{"f_id": id}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build select discovery_task sql: %v", err)
		span.SetStatus(codes.Error, "Build sql failed")
		return nil, err
	}

	task := &interfaces.DiscoveryTask{}
	var resultStr sql.NullString

	row := da.db.QueryRowContext(ctx, sqlStr, vals...)
	err = row.Scan(
		&task.ID,
		&task.CatalogID,
		&task.TriggerType,
		&task.Status,
		&task.Progress,
		&task.Message,
		&task.StartTime,
		&task.FinishTime,
		&resultStr,
		&task.Creator.ID,
		&task.Creator.Type,
		&task.CreateTime,
	)
	if err == sql.ErrNoRows {
		span.SetStatus(codes.Ok, "")
		return nil, nil
	}
	if err != nil {
		logger.Errorf("Scan discovery_task failed: %v", err)
		span.SetStatus(codes.Error, "Scan failed")
		return nil, err
	}

	// Deserialize result
	if resultStr.Valid && resultStr.String != "" {
		task.Result = &interfaces.DiscoveryResult{}
		_ = sonic.UnmarshalString(resultStr.String, task.Result)
	}

	span.SetStatus(codes.Ok, "")
	return task, nil
}

// List lists DiscoveryTasks with filters.
func (da *discoveryTaskAccess) List(ctx context.Context, params interfaces.DiscoveryTaskQueryParams) ([]*interfaces.DiscoveryTask, int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "List discovery_tasks",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	builder := sq.Select(
		"f_id",
		"f_catalog_id",
		"f_trigger_type",
		"f_status",
		"f_progress",
		"f_message",
		"f_start_time",
		"f_finish_time",
		"f_result",
		"f_creator",
		"f_creator_type",
		"f_create_time",
	).From(DISCOVERY_TASK_TABLE_NAME)

	countBuilder := sq.Select("COUNT(*)").From(DISCOVERY_TASK_TABLE_NAME)

	if params.CatalogID != "" {
		builder = builder.Where(sq.Eq{"f_catalog_id": params.CatalogID})
		countBuilder = countBuilder.Where(sq.Eq{"f_catalog_id": params.CatalogID})
	}
	if params.Status != "" {
		builder = builder.Where(sq.Eq{"f_status": params.Status})
		countBuilder = countBuilder.Where(sq.Eq{"f_status": params.Status})
	}
	if params.TriggerType != "" {
		builder = builder.Where(sq.Eq{"f_trigger_type": params.TriggerType})
		countBuilder = countBuilder.Where(sq.Eq{"f_trigger_type": params.TriggerType})
	}

	countSql, countVals, _ := countBuilder.ToSql()
	var total int64
	err := da.db.QueryRowContext(ctx, countSql, countVals...).Scan(&total)
	if err != nil {
		logger.Errorf("Failed to count discovery_tasks: %v", err)
		span.SetStatus(codes.Error, "Count failed")
		return nil, 0, err
	}

	// Pagination
	if params.Limit > 0 {
		builder = builder.Limit(uint64(params.Limit)).Offset(uint64(params.Offset))
	}
	builder = builder.OrderBy("f_create_time DESC")

	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		span.SetStatus(codes.Error, "Build sql failed")
		return nil, 0, err
	}

	rows, err := da.db.QueryContext(ctx, sqlStr, vals...)
	if err != nil {
		span.SetStatus(codes.Error, "Query failed")
		return nil, 0, err
	}
	defer rows.Close()

	tasks := make([]*interfaces.DiscoveryTask, 0)
	for rows.Next() {
		task := &interfaces.DiscoveryTask{}
		var resultStr sql.NullString

		err := rows.Scan(
			&task.ID,
			&task.CatalogID,
			&task.TriggerType,
			&task.Status,
			&task.Progress,
			&task.Message,
			&task.StartTime,
			&task.FinishTime,
			&resultStr,
			&task.Creator.ID,
			&task.Creator.Type,
			&task.CreateTime,
		)
		if err != nil {
			span.SetStatus(codes.Error, "Scan row failed")
			return nil, 0, err
		}

		if resultStr.Valid && resultStr.String != "" {
			task.Result = &interfaces.DiscoveryResult{}
			_ = sonic.UnmarshalString(resultStr.String, task.Result)
		}

		tasks = append(tasks, task)
	}

	span.SetStatus(codes.Ok, "")
	return tasks, total, nil
}

// UpdateStatus updates a DiscoveryTask's status and message.
func (da *discoveryTaskAccess) UpdateStatus(ctx context.Context, id, status, message string, stime int64) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Update discovery_task status",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("task_id").String(id),
		attr.Key("status").String(status),
	)

	data := map[string]any{
		"f_status":  status,
		"f_message": message,
	}
	if status == interfaces.DiscoveryTaskStatusRunning {
		data["f_start_time"] = stime
	}
	sqlStr, vals, err := sq.Update(DISCOVERY_TASK_TABLE_NAME).
		SetMap(data).
		Where(sq.Eq{"f_id": id}).
		ToSql()
	if err != nil {
		span.SetStatus(codes.Error, "Build sql failed")
		return err
	}

	_, err = da.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		span.SetStatus(codes.Error, "Update failed")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// UpdateProgress updates a DiscoveryTask's progress.
func (da *discoveryTaskAccess) UpdateProgress(ctx context.Context, id string, progress int) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Update discovery_task progress",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	sqlStr, vals, err := sq.Update(DISCOVERY_TASK_TABLE_NAME).
		Set("f_progress", progress).
		Set("f_update_time", time.Now().UnixMilli()).
		Where(sq.Eq{"f_id": id}).
		ToSql()
	if err != nil {
		span.SetStatus(codes.Error, "Build sql failed")
		return err
	}

	_, err = da.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		span.SetStatus(codes.Error, "Update failed")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// UpdateResult updates a DiscoveryTask's result and sets status to completed.
func (da *discoveryTaskAccess) UpdateResult(ctx context.Context, id string, result *interfaces.DiscoveryResult, stime int64) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Update discovery_task result",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	resultBytes, _ := sonic.MarshalString(result)

	sqlStr, vals, err := sq.Update(DISCOVERY_TASK_TABLE_NAME).
		Set("f_status", interfaces.DiscoveryTaskStatusCompleted).
		Set("f_result", resultBytes).
		Set("f_progress", 100).
		Set("f_finish_time", stime).
		Where(sq.Eq{"f_id": id}).
		ToSql()
	if err != nil {
		span.SetStatus(codes.Error, "Build sql failed")
		return err
	}

	_, err = da.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		span.SetStatus(codes.Error, "Update failed")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}
