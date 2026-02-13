// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package scan_record

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	sq "github.com/Masterminds/squirrel"
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
	SCAN_RECORD_TABLE_NAME = "t_scan_record"
)

var (
	tmAccessOnce sync.Once
	tmAccess     interfaces.ScanRecordAccess
)

type scanRecordAccess struct {
	appSetting *common.AppSetting
	db         *sql.DB
}

func NewScanRecordAccess(appSetting *common.AppSetting) interfaces.ScanRecordAccess {
	tmAccessOnce.Do(func() {
		tmAccess = &scanRecordAccess{
			appSetting: appSetting,
			db:         libdb.NewDB(&appSetting.DBSetting),
		}
	})

	return tmAccess
}

func (sra *scanRecordAccess) ListScanRecords(ctx context.Context, params *interfaces.PaginationQueryParameters) ([]*interfaces.ScanRecord, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Select scan records", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("offset").String(fmt.Sprintf("%d", params.Offset)),
		attr.Key("limit").String(fmt.Sprintf("%d", params.Limit)),
	)

	scanRecords := make([]*interfaces.ScanRecord, 0)

	builder := sq.Select(
		"f_record_id",
		"f_data_source_id",
		"f_scanner",
		"f_scan_time",
		"f_data_source_status",
		"f_metadata_task_id").
		From(SCAN_RECORD_TABLE_NAME)

	//添加分页参数 limit = -1 不分页，可选1-1000
	if params.Limit != -1 {
		builder = builder.Offset(uint64(params.Offset)).
			Limit(uint64(params.Limit))
	}

	sqlStr, args, err := builder.ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'list scan records' sql stmt failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "generate sql stmt failed")

		return nil, err
	}

	sqlStmt := fmt.Sprintf("Sql stmt for listing scan records is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	rows, err := sra.db.Query(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("List scan records failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "List scan records failed")

		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		scanRecord := &interfaces.ScanRecord{}
		err := rows.Scan(
			&scanRecord.RecordID,
			&scanRecord.DataSourceID,
			&scanRecord.Scanner,
			&scanRecord.ScanTime,
			&scanRecord.DataSourceStatus,
			&scanRecord.MetadataTaskID,
		)
		if err != nil {
			errDetails := fmt.Sprintf("Row scan failed, err: %s", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			span.SetStatus(codes.Error, "Row scan error")

			return nil, err
		}

		scanRecords = append(scanRecords, scanRecord)
	}

	span.SetStatus(codes.Ok, "")
	return scanRecords, nil
}

func (sra *scanRecordAccess) CreateScanRecord(ctx context.Context, scanRecord *interfaces.ScanRecord) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Insert scan record into DB", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	builder := sq.Insert(SCAN_RECORD_TABLE_NAME).
		Columns(
			"f_record_id",
			"f_data_source_id",
			"f_scanner",
			"f_scan_time",
			"f_data_source_status",
			"f_metadata_task_id",
		)

	builder = builder.Values(
		scanRecord.RecordID,
		scanRecord.DataSourceID,
		scanRecord.Scanner,
		scanRecord.ScanTime,
		scanRecord.DataSourceStatus,
		scanRecord.MetadataTaskID,
	)

	sqlStr, args, err := builder.ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'create scan record' sql stmt failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Generate sql stmt failed")

		return err
	}

	sqlStmt := fmt.Sprintf("sql stmt for creating scan record is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	_, err = sra.db.Exec(sqlStr, args...)

	if err != nil {
		errDetails := fmt.Sprintf("insert scan record failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Insert scan record failed")

		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}
func (sra *scanRecordAccess) UpdateScanRecord(ctx context.Context, scanRecord *interfaces.ScanRecord) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Update a scan record from DB", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("record_id").String(scanRecord.RecordID),
	)

	data := map[string]any{
		"f_data_source_id":     scanRecord.DataSourceID,
		"f_scanner":            scanRecord.Scanner,
		"f_scan_time":          scanRecord.ScanTime,
		"f_data_source_status": scanRecord.DataSourceStatus,
		"f_metadata_task_id":   scanRecord.MetadataTaskID,
	}

	sqlStr, args, err := sq.Update(SCAN_RECORD_TABLE_NAME).
		SetMap(data).
		Where(sq.Eq{"f_record_id": scanRecord.RecordID}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'update scan record' sql stmt failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "generate sql stmt failed")

		return err
	}

	sqlStmt := fmt.Sprintf("Sql stmt for updating scan record is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	_, err = sra.db.Exec(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("Update scan record failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Update scan record failed")

		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (sra *scanRecordAccess) UpdateScanRecordStatus(ctx context.Context, status *interfaces.ScanRecordStatus) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Update a scan record from DB", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("record_id").String(status.ID),
	)

	data := map[string]any{
		"f_data_source_status": status.Status,
	}

	sqlStr, args, err := sq.Update(SCAN_RECORD_TABLE_NAME).
		SetMap(data).
		Where(sq.Eq{"f_record_id": status.ID}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'update scan record status' sql stmt failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "generate sql stmt failed")

		return err
	}

	sqlStmt := fmt.Sprintf("Sql stmt for updating scan record status is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	_, err = sra.db.Exec(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("Update scan record status failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Update scan record status failed")

		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (sra *scanRecordAccess) DeleteByDataSourceId(ctx context.Context, datasourceId string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Delete scan record from DB", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("data_source_id").String(datasourceId),
	)

	sqlStr, args, err := sq.Delete(SCAN_RECORD_TABLE_NAME).
		Where(sq.Eq{"f_data_source_id": datasourceId}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'delete scan record' sql stmt failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Generate sql stmt failed")

		return err
	}

	sqlStmt := fmt.Sprintf("Sql stmt for deleting scan record is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	_, err = sra.db.Exec(sqlStr, args...)

	if err != nil {
		errDetails := fmt.Sprintf("Delete scan record failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Delete scan record failed")

		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (sra *scanRecordAccess) GetByDataSourceId(ctx context.Context, dataSourceId string) (*interfaces.ScanRecord, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Get scan records by data source id",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("data_source_id").String(dataSourceId),
	)

	sqlStr, args, err := sq.Select(
		"f_record_id",
		"f_data_source_id",
		"f_scanner",
		"f_scan_time",
		"f_data_source_status",
		"f_metadata_task_id",
	).From(SCAN_RECORD_TABLE_NAME).
		Where(sq.Eq{"f_data_source_id": dataSourceId}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'get scan records by data source id' sql stmt failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "generate sql stmt failed")

		return nil, false, err
	}

	sqlStmt := fmt.Sprintf("Sql stmt for getting scan records by data source id is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	rows := sra.db.QueryRow(sqlStr, args...)

	record := &interfaces.ScanRecord{}
	err = rows.Scan(
		&record.RecordID,
		&record.DataSourceID,
		&record.Scanner,
		&record.ScanTime,
		&record.DataSourceStatus,
		&record.MetadataTaskID,
	)
	if err == sql.ErrNoRows {
		span.SetAttributes(attr.Key("no_rows").Bool(true))
		span.SetStatus(codes.Ok, "")

		return nil, false, nil
	}

	if err != nil {
		errDetails := fmt.Sprintf("Row scan failed, error: %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Row scan failed")

		return nil, false, err
	}

	span.SetStatus(codes.Ok, "")
	return record, true, nil
}

func (sra *scanRecordAccess) GetByTaskIds(ctx context.Context, taskIds []string) ([]*interfaces.ScanRecord, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Get scan records by task ids",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("task_ids").String(fmt.Sprintf("%v", taskIds)),
	)

	sqlStr, args, err := sq.Select(
		"f_record_id",
		"f_data_source_id",
		"f_scanner",
		"f_scan_time",
		"f_data_source_status",
		"f_metadata_task_id",
	).From(SCAN_RECORD_TABLE_NAME).
		Where(sq.Eq{"f_scanner": taskIds}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'get scan records by task ids' sql stmt failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "generate sql stmt failed")

		return nil, err
	}

	sqlStmt := fmt.Sprintf("Sql stmt for getting scan records by task ids is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	scanRecords := []*interfaces.ScanRecord{}
	rows, err := sra.db.Query(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("Query scan records by task ids failed, error: %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Query scan records by task ids failed")

		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		record := &interfaces.ScanRecord{}
		err = rows.Scan(
			&record.RecordID,
			&record.DataSourceID,
			&record.Scanner,
			&record.ScanTime,
			&record.DataSourceStatus,
			&record.MetadataTaskID,
		)
		if err != nil {
			errDetails := fmt.Sprintf("Row scan failed, error: %s", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			span.SetStatus(codes.Error, "Row scan failed")

			return nil, err
		}

		scanRecords = append(scanRecords, record)
	}

	span.SetStatus(codes.Ok, "")
	return scanRecords, nil
}

func (sra *scanRecordAccess) GetByDataSourceIdAndScanner(ctx context.Context, dataSourceId string, taskId string) ([]*interfaces.ScanRecord, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Get scan records by data source id and task id",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("data_source_id").String(dataSourceId),
		attr.Key("task_id").String(taskId),
	)

	if taskId == "" {
		taskId = interfaces.ManagementScanner
	}

	sqlStr, args, err := sq.Select(
		"f_record_id",
		"f_data_source_id",
		"f_scanner",
		"f_scan_time",
		"f_data_source_status",
		"f_metadata_task_id",
	).From(SCAN_RECORD_TABLE_NAME).
		Where(sq.Eq{
			"f_data_source_id": dataSourceId,
			"f_scanner":        taskId,
		}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'get scan records by data source id and task id' sql stmt failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "generate sql stmt failed")

		return nil, err
	}

	sqlStmt := fmt.Sprintf("Sql stmt for getting scan records  by data source id and task id is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	scanRecords := []*interfaces.ScanRecord{}
	rows, err := sra.db.Query(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("Query scan records by  by data source id and task id failed, error: %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Query scan records by  by data source id and task id failed")

		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		record := &interfaces.ScanRecord{}
		err = rows.Scan(
			&record.RecordID,
			&record.DataSourceID,
			&record.Scanner,
			&record.ScanTime,
			&record.DataSourceStatus,
			&record.MetadataTaskID,
		)
		if err != nil {
			errDetails := fmt.Sprintf("Row scan failed, error: %s", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			span.SetStatus(codes.Error, "Row scan failed")

			return nil, err
		}

		scanRecords = append(scanRecords, record)
	}

	span.SetStatus(codes.Ok, "")
	return scanRecords, nil
}
