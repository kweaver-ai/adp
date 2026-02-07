// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_view

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	libCommon "github.com/kweaver-ai/kweaver-go-lib/common"
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
	DATA_VIEW_TABLE_NAME      = "t_data_view"
	DATA_MODEL_JOB_TABLE_NAME = "t_data_model_job"
)

var (
	dvAccessOnce sync.Once
	dvAccess     interfaces.DataViewAccess
)

type dataViewAccess struct {
	appSetting *common.AppSetting
	db         *sql.DB
}

func NewDataViewAccess(appSetting *common.AppSetting) interfaces.DataViewAccess {
	dvAccessOnce.Do(func() {
		dvAccess = &dataViewAccess{
			appSetting: appSetting,
			db:         libdb.NewDB(&appSetting.DBSetting),
		}
	})

	return dvAccess
}

// 创建数据视图
func (dva *dataViewAccess) CreateDataViews(ctx context.Context, tx *sql.Tx, views []*interfaces.DataView) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Insert data views into DB", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	builder := sq.Insert(DATA_VIEW_TABLE_NAME).
		Columns(
			"f_view_id",
			// "f_uniform_catalog_code",
			"f_view_name",
			"f_technical_name",
			"f_group_id",
			"f_type",
			"f_query_type",
			"f_builtin",
			"f_tags",
			"f_comment",
			"f_data_source_type",
			"f_data_source_id",
			"f_file_name",
			"f_excel_config",
			"f_data_scope",
			"f_fields",
			"f_status",
			"f_metadata_form_id",
			"f_primary_keys",
			"f_sql",
			"f_meta_table_name",
			"f_create_time",
			"f_update_time",
			"f_creator",
			"f_creator_type",
			"f_updater",
			"f_updater_type",
		)

	for _, view := range views {
		tagsStr := libCommon.TagSlice2TagString(view.Tags)
		primaryKeysStr := libCommon.TagSlice2TagString(view.PrimaryKeys)

		// dataSourceBytes, err := sonic.Marshal(view.DataSource)
		// if err != nil {
		// 	errDetails := fmt.Sprintf("Marshal dataSource failed, %s", err.Error())
		// 	logger.Error(errDetails)
		// 	o11y.Error(ctx, errDetails)
		// 	span.SetStatus(codes.Error, "Marshal dataSource failed")

		// 	return err
		// }

		excelConfigBytes, err := sonic.Marshal(view.ExcelConfig)
		if err != nil {
			errDetails := fmt.Sprintf("Marshal excelConfig failed, %s", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			span.SetStatus(codes.Error, "Marshal excelConfig failed")

			return err
		}

		dataScopeBytes, err := sonic.Marshal(view.DataScope)
		if err != nil {
			errDetails := fmt.Sprintf("Marshal dataScope failed, %s", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			span.SetStatus(codes.Error, "Marshal dataScope failed")

			return err
		}

		fieldsBytes, err := sonic.Marshal(view.Fields)
		if err != nil {
			errDetails := fmt.Sprintf("Marshal fields failed, %s", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			span.SetStatus(codes.Error, "Marshal fields failed")

			return err
		}

		// condBytes, err := sonic.Marshal(view.Condition)
		// if err != nil {
		// 	errDetails := fmt.Sprintf("Marshal condition failed, %s", err.Error())
		// 	logger.Error(errDetails)
		// 	o11y.Error(ctx, errDetails)
		// 	span.SetStatus(codes.Error, "Marshal condition failed")

		// 	return err
		// }

		builder = builder.Values(
			view.ViewID,
			// view.UniformCatalogCode,
			view.ViewName,
			view.TechnicalName,
			view.GroupID,
			view.Type,
			view.QueryType,
			view.Builtin,
			tagsStr,
			view.Comment,
			view.DataSourceType,
			view.DataSourceID,
			view.FileName,
			excelConfigBytes,
			dataScopeBytes,
			fieldsBytes,
			view.Status,
			view.MetadataFormID,
			primaryKeysStr,
			view.SQLStr,
			view.MetaTableName,
			view.CreateTime,
			view.UpdateTime,
			view.Creator.ID,
			view.Creator.Type,
			view.Updater.ID,
			view.Updater.Type,
		)
	}

	sqlStr, args, err := builder.ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'create views' sql stmt failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Generate sql stmt failed")

		return err
	}

	sqlStmt := fmt.Sprintf("sql stmt for creating views is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	if tx == nil {
		_, err = dva.db.Exec(sqlStr, args...)
	} else {
		_, err = tx.Exec(sqlStr, args...)
	}
	if err != nil {
		errDetails := fmt.Sprintf("insert data views failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Insert data views failed")

		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 删除数据视图
func (dva *dataViewAccess) DeleteDataViews(ctx context.Context, tx *sql.Tx, viewIDs []string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Delete data views from DB", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("view_ids").String(fmt.Sprintf("%v", viewIDs)),
	)

	if len(viewIDs) == 0 {
		return nil
	}

	sqlStr, args, err := sq.Delete(DATA_VIEW_TABLE_NAME).
		Where(sq.Eq{"f_view_id": viewIDs}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'delete views' sql stmt failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Generate sql stmt failed")

		return err
	}

	sqlStmt := fmt.Sprintf("Sql stmt for deleting views is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	if tx == nil {
		_, err = dva.db.Exec(sqlStr, args...)
	} else {
		_, err = tx.Exec(sqlStr, args...)
	}
	if err != nil {
		errDetails := fmt.Sprintf("Delete data views failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Delete data views failed")

		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 修改数据视图
func (dva *dataViewAccess) UpdateDataView(ctx context.Context, tx *sql.Tx, view *interfaces.DataView) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Update a data view from DB", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("view_id").String(view.ViewID),
	)

	tagsStr := libCommon.TagSlice2TagString(view.Tags)
	primaryKeysStr := libCommon.TagSlice2TagString(view.PrimaryKeys)

	// dataSourceBytes, err := sonic.Marshal(view.DataSource)
	// if err != nil {
	// 	errDetails := fmt.Sprintf("Marshal dataSource failed, %s", err.Error())
	// 	logger.Error(errDetails)
	// 	o11y.Error(ctx, errDetails)
	// 	span.SetStatus(codes.Error, "Marshal dataSource failed")

	// 	return err
	// }

	excelConfigBytes, err := sonic.Marshal(view.ExcelConfig)
	if err != nil {
		errDetails := fmt.Sprintf("Marshal excelConfig failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Marshal excelConfig failed")

		return err
	}

	dataScopeBytes, err := sonic.Marshal(view.DataScope)
	if err != nil {
		errDetails := fmt.Sprintf("Marshal dataScope failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Marshal dataScope failed")

		return err
	}

	fieldsBytes, err := sonic.Marshal(view.Fields)
	if err != nil {
		errDetails := fmt.Sprintf("Marshal fields failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Marshal fields failed")

		return err
	}

	// condBytes, err := sonic.Marshal(view.Condition)
	// if err != nil {
	// 	errDetails := fmt.Sprintf("Marshal condition failed, %s", err.Error())
	// 	logger.Error(errDetails)
	// 	o11y.Error(ctx, errDetails)
	// 	span.SetStatus(codes.Error, "Marshal condition failed")

	// 	return err
	// }

	updateMap := map[string]any{
		"f_view_name":    view.ViewName,
		"f_group_id":     view.GroupID,
		"f_query_type":   view.QueryType,
		"f_tags":         tagsStr,
		"f_comment":      view.Comment,
		"f_file_name":    view.FileName,
		"f_excel_config": excelConfigBytes,
		"f_data_scope":   dataScopeBytes,
		"f_fields":       fieldsBytes,
		"f_primary_keys": primaryKeysStr,
		"f_sql":          view.SQLStr,
		"f_status":       view.Status,
		"f_update_time":  view.UpdateTime,
		"f_updater":      view.Updater.ID,
		"f_updater_type": view.Updater.Type,
	}
	sqlStr, args, err := sq.Update(DATA_VIEW_TABLE_NAME).
		SetMap(updateMap).
		Where(sq.Eq{"f_view_id": view.ViewID}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'update a view' sql stmt failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Generate sql stmt failed")

		return err
	}

	sqlSmt := fmt.Sprintf("Sql stmt for updating a view is '%s'", sqlStr)
	logger.Debug(sqlSmt)
	o11y.Info(ctx, sqlSmt)

	if tx == nil {
		_, err = dva.db.Exec(sqlStr, args...)
	} else {
		_, err = tx.Exec(sqlStr, args...)
	}
	if err != nil {
		errDetails := fmt.Sprintf("Update a data view failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Update data view failed")

		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 按 id 批量获取视图详情
func (dva *dataViewAccess) GetDataViews(ctx context.Context, viewIDs []string) ([]*interfaces.DataView, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Get data views from DB", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("view_ids").String(fmt.Sprintf("%v", viewIDs)),
	)

	sqlStr, args, err := sq.Select(
		"dv.f_view_id",
		"dv.f_view_name",
		"dv.f_technical_name",
		"dv.f_group_id",
		"COALESCE(dvg.f_group_name, '')",
		"dv.f_type",
		"dv.f_query_type",
		"dv.f_builtin",
		"dv.f_tags",
		"dv.f_comment",
		"dv.f_data_source_type",
		"dv.f_data_source_id",
		"dv.f_file_name",
		"dv.f_excel_config",
		"dv.f_data_scope",
		"dv.f_fields",
		"dv.f_status",
		"dv.f_metadata_form_id",
		"dv.f_primary_keys",
		"COALESCE(dv.f_sql, '')",
		"dv.f_meta_table_name",
		"dv.f_create_time",
		"dv.f_update_time",
		"dv.f_delete_time",
		"dv.f_creator",
		"dv.f_creator_type",
		"dv.f_updater",
		"dv.f_updater_type",
	).
		From(fmt.Sprintf("%s as dv", DATA_VIEW_TABLE_NAME)).
		LeftJoin(fmt.Sprintf("%s as dvg on dv.f_group_id = dvg.f_group_id", DATA_VIEW_GROUP_TABLE_NAME)).
		// LeftJoin(fmt.Sprintf("%s as dmj on dv.f_job_id = dmj.f_job_id", DATA_MODEL_JOB_TABLE_NAME)).
		Where(sq.Eq{"dv.f_view_id": viewIDs}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'get views' sql stmt error: %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Generate sql stmt failed")

		return nil, err
	}

	sqlStmt := fmt.Sprintf("Sql stmt for getting views is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	views := []*interfaces.DataView{}
	rows, err := dva.db.Query(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("Query data views failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Query views by IDs failed")

		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tagsStr, primaryKeysStr string
		var excelConfigBytes, dataScopeBytes, fieldsBytes []byte
		view := &interfaces.DataView{
			ModuleType: interfaces.MODULE_TYPE_DATA_VIEW,
		}
		err = rows.Scan(
			&view.ViewID,
			&view.ViewName,
			&view.TechnicalName,
			&view.GroupID,
			&view.GroupName,
			&view.Type,
			&view.QueryType,
			&view.Builtin,
			&tagsStr,
			&view.Comment,
			&view.DataSourceType,
			&view.DataSourceID,
			&view.FileName,
			&excelConfigBytes,
			&dataScopeBytes,
			&fieldsBytes,
			&view.Status,
			&view.MetadataFormID,
			&primaryKeysStr,
			&view.SQLStr,
			&view.MetaTableName,
			&view.CreateTime,
			&view.UpdateTime,
			&view.DeleteTime,
			&view.Creator.ID,
			&view.Creator.Type,
			&view.Updater.ID,
			&view.Updater.Type,
		)
		if err != nil {
			errDetails := fmt.Sprintf("Row scan failed, error: %s", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			span.SetStatus(codes.Error, "Row scan failed")

			return nil, err
		}

		view.Tags = libCommon.TagString2TagSlice(tagsStr)
		view.PrimaryKeys = libCommon.TagString2TagSlice(primaryKeysStr)

		// 反序列化
		// err := sonic.Unmarshal([]byte(dataSourceBytes), &view.DataSource)
		// if err != nil {
		// 	errDetails := fmt.Sprintf("Unmarshal dataSource failed, %s", err.Error())
		// 	logger.Error(errDetails)
		// 	o11y.Error(ctx, errDetails)
		// 	span.SetStatus(codes.Error, "Unmarshal dataSource failed")

		// 	return nil, err
		// }
		if len(excelConfigBytes) != 0 {
			err = sonic.Unmarshal([]byte(excelConfigBytes), &view.ExcelConfig)
			if err != nil {
				errDetails := fmt.Sprintf("Unmarshal excelConfigBytes failed, %s", err.Error())
				logger.Error(errDetails)
				o11y.Error(ctx, errDetails)
				span.SetStatus(codes.Error, "Unmarshal excelConfigBytes failed")

				return nil, err
			}
		}

		if len(dataScopeBytes) != 0 {
			err = sonic.Unmarshal([]byte(dataScopeBytes), &view.DataScope)
			if err != nil {
				errDetails := fmt.Sprintf("Unmarshal dataScopeBytes failed, %s", err.Error())
				logger.Error(errDetails)
				o11y.Error(ctx, errDetails)
				span.SetStatus(codes.Error, "Unmarshal dataScopeBytes failed")

				return nil, err
			}
		}

		err = sonic.Unmarshal([]byte(fieldsBytes), &view.Fields)
		if err != nil {
			errDetails := fmt.Sprintf("Unmarshal fields failed, %s", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			span.SetStatus(codes.Error, "Unmarshal fields failed")

			return nil, err
		}

		// err = sonic.Unmarshal([]byte(condBytes), &view.Condition)
		// if err != nil {
		// 	errDetails := fmt.Sprintf("Unmarshal condition failed, %s", err.Error())
		// 	logger.Error(errDetails)
		// 	o11y.Error(ctx, errDetails)
		// 	span.SetStatus(codes.Error, "Unmarshal condition failed")

		// 	return nil, err
		// }

		views = append(views, view)
	}

	span.SetStatus(codes.Ok, "")
	return views, nil
}

// 查询数据视图列表
func (dva *dataViewAccess) ListDataViews(ctx context.Context,
	query *interfaces.ListViewQueryParams) ([]*interfaces.SimpleDataView, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: List data views from DB",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("type").String(fmt.Sprintf("%v", query.Type)),
		attr.Key("query_type").String(fmt.Sprintf("%v", query.QueryType)),
		attr.Key("keyword").String(query.Keyword),
		attr.Key("name").String(query.Name),
		attr.Key("technical_name").String(query.TechnicalName),
		attr.Key("name_pattern").String(query.NamePattern),
		attr.Key("group_id").String(query.GroupID),
		attr.Key("group_name").String(query.GroupName),
		attr.Key("data_source_type").String(query.DataSourceType),
		attr.Key("data_source_id").String(query.DataSourceID),
		attr.Key("file_name").String(query.FileName),
		attr.Key("builtin").String(fmt.Sprintf("%v", query.Builtin)),
		attr.Key("status").String(fmt.Sprintf("%v", query.Status)),
		attr.Key("create_time_start").String(fmt.Sprintf("%v", query.CreateTimeStart)),
		attr.Key("create_time_end").String(fmt.Sprintf("%v", query.CreateTimeEnd)),
		attr.Key("update_time_start").String(fmt.Sprintf("%v", query.UpdateTimeStart)),
		attr.Key("update_time_end").String(fmt.Sprintf("%v", query.UpdateTimeEnd)),
		attr.Key("offset").String(fmt.Sprintf("%d", query.Offset)),
		attr.Key("limit").String(fmt.Sprintf("%d", query.Limit)),
		attr.Key("sort").String(query.Sort),
		attr.Key("direction").String(query.Direction),
		attr.Key("tag").String(query.Tag),
		// attr.Key("open_streaming").String(fmt.Sprintf("%v", query.OpenStreaming)),
		// attr.Key("field_scope").String(query.FieldScopeStr),
	)

	views := make([]*interfaces.SimpleDataView, 0)

	builder := sq.Select(
		"dv.f_view_id",
		"dv.f_view_name",
		"dv.f_technical_name",
		"dv.f_group_id",
		"COALESCE(dvg.f_group_name, '')",
		"dv.f_type",
		"dv.f_query_type",
		"dv.f_builtin",
		"dv.f_tags",
		"dv.f_comment",
		"dv.f_data_source_type",
		"dv.f_data_source_id",
		"dv.f_file_name",
		"dv.f_status",
		"dv.f_create_time",
		"dv.f_update_time",
		"dv.f_delete_time",
	).
		From(fmt.Sprintf("%s as dv", DATA_VIEW_TABLE_NAME)).
		LeftJoin(fmt.Sprintf("%s as dvg on dv.f_group_id = dvg.f_group_id", DATA_VIEW_GROUP_TABLE_NAME))

	// 过滤
	builder, err := buildViewListQuerySQL(query, builder)
	if err != nil {
		errDetails := fmt.Sprintf("Joint view list query sql failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Joint view list query sql failed")

		return nil, err
	}

	//排序
	if query.Sort == "f_group_name" {
		builder = builder.OrderByClause(fmt.Sprintf("dvg.%s %s,dv.%s %s", query.Sort, query.Direction, "f_view_name", query.Direction))
	} else if query.Sort == "f_view_name" || query.Sort == "f_technical_name" {
		builder = builder.OrderByClause(fmt.Sprintf("dv.%s %s,dvg.%s %s", query.Sort, query.Direction, "f_group_name", query.Direction))
	} else {
		builder = builder.OrderByClause(fmt.Sprintf("dv.%s %s", query.Sort, query.Direction))
	}

	// 接入权限后不在数据库查询时分页，需从数据库中获取所有对象
	//添加分页参数 limit = -1 不分页，可选1-1000
	// if query.Limit != -1 {
	// 	builder = builder.Limit(uint64(query.Limit)).
	// 		Offset(uint64(query.Offset))
	// }

	sqlStr, args, err := builder.ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'list view' sql stmt failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Generate sql stmt failed")

		return nil, err
	}

	sqlStmt := fmt.Sprintf("Sql stmt for listing views is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	rows, err := dva.db.Query(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("List data views failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "List data views failed")

		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tagsStr string
		// var dataSourceBytes []byte
		view := &interfaces.SimpleDataView{}
		err := rows.Scan(
			&view.ViewID,
			&view.ViewName,
			&view.TechnicalName,
			&view.GroupID,
			&view.GroupName,
			&view.Type,
			&view.QueryType,
			&view.Builtin,
			&tagsStr,
			&view.Comment,
			&view.DataSourceType,
			&view.DataSourceID,
			&view.FileName,
			&view.Status,
			&view.CreateTime,
			&view.UpdateTime,
			&view.DeleteTime,
		)
		if err != nil {
			errDetails := fmt.Sprintf("Row scan failed, err: %s", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			span.SetStatus(codes.Error, "Row scan failed")

			return nil, err
		}

		view.Tags = libCommon.TagString2TagSlice(tagsStr)

		// 反序列化
		// err = sonic.Unmarshal(dataSourceBytes, &view.DataSource)
		// if err != nil {
		// 	errDetails := fmt.Sprintf("Unmarshal dataSource failed, %s", err.Error())
		// 	logger.Error(errDetails)
		// 	o11y.Error(ctx, errDetails)
		// 	span.SetStatus(codes.Error, "Unmarshal dataSource failed")

		// 	return nil, err
		// }

		// 只返回数据源类型
		// delete(view.DataSource, "index_base")

		views = append(views, view)
	}

	span.SetStatus(codes.Ok, "")
	return views, nil
}

// 查询数据视图总数
func (dva *dataViewAccess) GetDataViewsTotal(ctx context.Context,
	query *interfaces.ListViewQueryParams) (int, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Get data views total from DB",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("name").String(query.Name),
		attr.Key("name_pattern").String(query.NamePattern),
		attr.Key("group_id").String(query.GroupID),
		attr.Key("tag").String(query.Tag),
		attr.Key("builtin").String(fmt.Sprintf("%v", query.Builtin)),
	)

	builder := sq.Select("COUNT(dv.f_view_id)").
		From(fmt.Sprintf("%s as dv", DATA_VIEW_TABLE_NAME)).
		LeftJoin(fmt.Sprintf("%s as dvg on dv.f_group_id = dvg.f_group_id", DATA_VIEW_GROUP_TABLE_NAME))

	// 过滤
	builder, err := buildViewListQuerySQL(query, builder)
	if err != nil {
		errDetails := fmt.Sprintf("Joint view list query sql failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Joint view list query sql failed")

		return 0, err
	}

	sqlStr, args, err := builder.ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'get views total' sql stmt failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Generate sql stmt failed")

		return 0, err
	}

	sqlStmt := fmt.Sprintf("Sql stmt for getting views total is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	var total int
	row := dva.db.QueryRow(sqlStr, args...)
	err = row.Scan(
		&total,
	)
	if err != nil {
		errDetails := fmt.Sprintf("Scan total failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Scan total failed")

		return 0, err
	}

	span.SetStatus(codes.Ok, "")
	return total, nil
}

// 根据ID获取数据视图
func (dva *dataViewAccess) CheckDataViewExistByID(ctx context.Context, tx *sql.Tx, viewID string) (string, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Check data view exist by ID", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("view_id").String(viewID),
	)

	sqlStr, args, err := sq.Select("f_view_name").
		From(DATA_VIEW_TABLE_NAME).
		Where(sq.Eq{"f_view_id": viewID}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'check view exist by ID' sql stmt failed, %s", err.Error())
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Generate sql stmt failed")

		return "", false, err
	}

	sqlStmt := fmt.Sprintf("sql stmt for checking view exists by ID is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	var name string
	if tx == nil {
		err = dva.db.QueryRow(sqlStr, args...).Scan(&name)
	} else {
		err = tx.QueryRow(sqlStr, args...).Scan(&name)
	}
	if err == sql.ErrNoRows {
		errDetails := fmt.Sprintf("Data view %s not found", viewID)
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Data view not found")

		return "", false, nil
	}

	if err != nil {
		errDetails := fmt.Sprintf("Row scan failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Row scan failed")

		return "", false, err
	}

	span.SetStatus(codes.Ok, "")
	return name, true, nil
}

// 根据视图名称和分组名称获取数据视图
func (dva *dataViewAccess) CheckDataViewExistByName(ctx context.Context, tx *sql.Tx, viewName, groupName string) (string, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Check data view exist by name", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("view_name").String(viewName),
		attr.Key("group_name").String(groupName),
	)

	var sqlStr string
	var args []any
	var err error
	sqlStr, args, err = sq.Select("dv.f_view_id").
		From(fmt.Sprintf("%s as dv", DATA_VIEW_TABLE_NAME)).
		LeftJoin(fmt.Sprintf("%s as dvg on dv.f_group_id = dvg.f_group_id", DATA_VIEW_GROUP_TABLE_NAME)).
		Where(sq.Eq{
			"dv.f_view_name":   viewName,
			"dvg.f_group_name": groupName,
		}).
		ToSql()

	if err != nil {
		errDetails := fmt.Sprintf("Generate 'check view exist by name' sql stmt failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Generate sql stmt failed")

		return "", false, err
	}

	sqlStmt := fmt.Sprintf("Sql stmt for checking view exists by name is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	var viewID string
	if tx == nil {
		err = dva.db.QueryRow(sqlStr, args...).Scan(&viewID)
	} else {
		err = tx.QueryRow(sqlStr, args...).Scan(&viewID)
	}
	if err == sql.ErrNoRows {
		span.SetAttributes(attr.Key("no_rows").Bool(true))
		span.SetStatus(codes.Ok, "")

		return "", false, nil
	}

	if err != nil {
		errDetails := fmt.Sprintf("Row scan failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Row scan failed")

		return "", false, err
	}

	span.SetStatus(codes.Ok, "")
	return viewID, true, nil
}

// 根据视图技术名称和分组名称获取数据视图
func (dva *dataViewAccess) CheckDataViewExistByTechnicalName(ctx context.Context, tx *sql.Tx, viewTechnicalName, groupName string) (string, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Check data view exist by technical name", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("view_technical_name").String(viewTechnicalName),
		attr.Key("group_name").String(groupName),
	)

	var sqlStr string
	var args []any
	var err error
	sqlStr, args, err = sq.Select("dv.f_view_id").
		From(fmt.Sprintf("%s as dv", DATA_VIEW_TABLE_NAME)).
		LeftJoin(fmt.Sprintf("%s as dvg on dv.f_group_id = dvg.f_group_id", DATA_VIEW_GROUP_TABLE_NAME)).
		Where(sq.Eq{
			"dv.f_technical_name": viewTechnicalName,
			"dvg.f_group_name":    groupName,
		}).
		ToSql()

	if err != nil {
		errDetails := fmt.Sprintf("Generate 'check view exist by technical name' sql stmt failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Generate sql stmt failed")

		return "", false, err
	}

	sqlStmt := fmt.Sprintf("Sql stmt for checking view exists by techincal name is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	var viewID string
	if tx == nil {
		err = dva.db.QueryRow(sqlStr, args...).Scan(&viewID)
	} else {
		err = tx.QueryRow(sqlStr, args...).Scan(&viewID)
	}
	if err == sql.ErrNoRows {
		span.SetAttributes(attr.Key("no_rows").Bool(true))
		span.SetStatus(codes.Ok, "")

		return "", false, nil
	}

	if err != nil {
		errDetails := fmt.Sprintf("Row scan failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Row scan failed")

		return "", false, err
	}

	span.SetStatus(codes.Ok, "")
	return viewID, true, nil
}

func (dva *dataViewAccess) UpdateDataViewsAttrs(ctx context.Context, attrs *interfaces.AtomicViewUpdateReq) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Update data view's group", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("view_id").String(attrs.ViewID),
	)
	defer span.End()

	fieldsBytes, err := sonic.Marshal(attrs.Fields)
	if err != nil {
		errDetails := fmt.Sprintf("Marshal fields failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Marshal fields failed")

		return err
	}

	updateMap := map[string]any{
		"f_view_name":    attrs.ViewName,
		"f_fields":       fieldsBytes,
		"f_comment":      attrs.Comment,
		"f_update_time":  attrs.UpdateTime,
		"f_updater":      attrs.Updater.ID,
		"f_updater_type": attrs.Updater.Type,
	}
	sqlStr, args, err := sq.Update(DATA_VIEW_TABLE_NAME).
		SetMap(updateMap).
		Where(sq.Eq{"f_view_id": attrs.ViewID}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'update the group of data views' sql stmt failed, %s", err.Error())
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Generate sql stmt failed")

		return err
	}

	sqlStmt := fmt.Sprintf("sql stmt for update the group of data views is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	_, err = dva.db.Exec(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("Execute sql stmt for updating the group of data views failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Execute sql stmt failed")

		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 批量更新视图的分组
func (dva *dataViewAccess) UpdateDataViewsGroup(ctx context.Context, tx *sql.Tx, viewIDs []string, groupID string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Update data view's group", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("view_ids").String(fmt.Sprintf("%v", viewIDs)),
		attr.Key("group_id").String(groupID),
	)
	defer span.End()

	updateMap := map[string]any{
		"f_group_id":    groupID,
		"f_update_time": time.Now().UnixMilli(),
	}
	sqlStr, args, err := sq.Update(DATA_VIEW_TABLE_NAME).
		SetMap(updateMap).
		Where(sq.Eq{"f_view_id": viewIDs}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'update the group of data views' sql stmt failed, %s", err.Error())
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Generate sql stmt failed")

		return err
	}

	sqlStmt := fmt.Sprintf("sql stmt for update the group of data views is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	if tx == nil {
		_, err = dva.db.Exec(sqlStr, args...)
	} else {
		_, err = tx.Exec(sqlStr, args...)
	}
	if err != nil {
		errDetails := fmt.Sprintf("Execute sql stmt for updating the group of data views failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Execute sql stmt failed")

		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (dva *dataViewAccess) UpdateDataViewRealTimeStreaming(ctx context.Context, tx *sql.Tx, realTimeStreaming *interfaces.RealTimeStreaming) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Update data view real time streaming", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("view_id").String(realTimeStreaming.ViewID),
		attr.Key("job_id").String(realTimeStreaming.JobID),
		attr.Key("open_streaming").String(fmt.Sprintf("%t", realTimeStreaming.OpenStreaming)),
	)

	updateMap := map[string]any{
		"f_open_streaming": realTimeStreaming.OpenStreaming,
		"f_job_id":         realTimeStreaming.JobID,
		"f_update_time":    realTimeStreaming.UpdateTime,
	}

	sqlStr, args, err := sq.Update(DATA_VIEW_TABLE_NAME).
		SetMap(updateMap).
		Where(sq.Eq{"f_view_id": realTimeStreaming.ViewID}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'update data view real-time streaming' sql stmt failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Generate sql stmt failed")

		return err
	}

	sqlStmt := fmt.Sprintf("sql stmt for update data view real-time streaming is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	if tx == nil {
		_, err = dva.db.Exec(sqlStr, args...)
	} else {
		_, err = tx.Exec(sqlStr, args...)
	}
	if err != nil {
		errDetails := fmt.Sprintf("Execute sql stmt for real-time streaming failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Execute sql stmt failed")

		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 内部逻辑，只更新视图状态，不更新更新时间和更新人
func (dva *dataViewAccess) UpdateViewStatus(ctx context.Context, tx *sql.Tx, viewIDs []string, param *interfaces.UpdateViewStatus) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Update data view's status", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("view_ids").String(fmt.Sprintf("%v", viewIDs)),
	)
	defer span.End()

	updateMap := map[string]any{
		"f_status":      param.ViewStatus,
		"f_delete_time": param.DeleteTime,
		// "f_update_time":  param.UpdateTime,
		// "f_updater":      param.Updater.ID,
		// "f_updater_type": param.Updater.Type,
	}
	sqlStr, args, err := sq.Update(DATA_VIEW_TABLE_NAME).
		SetMap(updateMap).
		Where(sq.Eq{"f_view_id": viewIDs}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'update the status of data views' sql stmt failed, %s", err.Error())
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Generate sql stmt failed")

		return err
	}

	sqlStmt := fmt.Sprintf("sql stmt for update the status of data views is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	if tx == nil {
		_, err = dva.db.Exec(sqlStr, args...)
	} else {
		_, err = tx.Exec(sqlStr, args...)
	}
	if err != nil {
		errDetails := fmt.Sprintf("Execute sql stmt for updating the status of data views failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Execute sql stmt failed")

		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 标记删除视图
func (dva *dataViewAccess) MarkDataViewsDeleted(ctx context.Context, tx *sql.Tx, params *interfaces.MarkViewDeletedParams) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Mark data views deleted", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("view_ids").String(fmt.Sprintf("%v", params.ViewIDs)),
	)

	updateMap := map[string]any{
		"f_delete_time": params.DeleteTime,
		"f_status":      params.ViewStatus,
	}
	sqlStr, args, err := sq.Update(DATA_VIEW_TABLE_NAME).
		SetMap(updateMap).
		Where(sq.Eq{"f_view_id": params.ViewIDs}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'update the status of data views' sql stmt failed, %s", err.Error())
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Generate sql stmt failed")

		return err
	}

	sqlStmt := fmt.Sprintf("sql stmt for update the status of data views is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	if tx == nil {
		_, err = dva.db.Exec(sqlStr, args...)
	} else {
		_, err = tx.Exec(sqlStr, args...)
	}
	if err != nil {
		errDetails := fmt.Sprintf("Execute sql stmt for updating the status of data views failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Execute sql stmt failed")

		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 拼接列表查询sql语句
func buildViewListQuerySQL(query *interfaces.ListViewQueryParams, builder sq.SelectBuilder) (sq.SelectBuilder, error) {
	if query.Type != "" {
		builder = builder.Where(sq.Eq{"dv.f_type": query.Type})
	}
	if query.QueryType != "" {
		builder = builder.Where(sq.Eq{"dv.f_query_type": query.QueryType})
	}

	if len(query.Builtin) > 0 {
		builder = builder.Where(sq.Eq{"dv.f_builtin": query.Builtin})
	}

	if len(query.Status) > 0 {
		builder = builder.Where(sq.Eq{"dv.f_status": query.Status})
	}

	if query.CreateTimeStart > 0 {
		builder = builder.Where(sq.GtOrEq{"dv.f_create_time": query.CreateTimeStart})
	}

	if query.CreateTimeEnd > 0 {
		builder = builder.Where(sq.LtOrEq{"dv.f_create_time": query.CreateTimeEnd})
	}

	if query.UpdateTimeStart > 0 {
		builder = builder.Where(sq.GtOrEq{"dv.f_update_time": query.UpdateTimeStart})
	}

	if query.UpdateTimeEnd > 0 {
		builder = builder.Where(sq.LtOrEq{"dv.f_update_time": query.UpdateTimeEnd})
	}

	// if len(query.OpenStreaming) > 0 {
	// 	builder = builder.Where(sq.Eq{"dv.f_open_streaming": query.OpenStreaming})
	// }

	if query.Name != "" {
		builder = builder.Where(sq.Eq{"dv.f_view_name": query.Name})
	} else if query.NamePattern != "" {
		builder = builder.Where(sq.Expr("instr(dv.f_view_name, ?) > 0", query.NamePattern))
	}

	if query.Keyword != "" {
		builder = builder.Where(sq.Expr("(instr(LOWER(dv.f_view_name), LOWER(?)) > 0 OR instr(LOWER(dv.f_technical_name), LOWER(?)) > 0)",
			query.Keyword, query.Keyword))
	}

	if query.TechnicalName != "" {
		builder = builder.Where(sq.Eq{"dv.f_technical_name": query.TechnicalName})
	}

	if query.DataSourceType != "" {
		builder = builder.Where(sq.Eq{"dv.f_data_source_type": query.DataSourceType})
	}

	if query.DataSourceID != "" {
		builder = builder.Where(sq.Eq{"dv.f_data_source_id": query.DataSourceID})
	}

	if query.FileName != "" {
		builder = builder.Where(sq.Eq{"dv.f_file_name": query.FileName})
	}

	// 根据分组 ID 过滤
	if query.GroupID != interfaces.GroupID_All {
		builder = builder.Where(sq.Eq{"dv.f_group_id": query.GroupID})
	}

	// 根据分组名称过滤
	if query.GroupName != interfaces.GroupName_All {
		builder = builder.Where(sq.Eq{"dvg.f_group_name": query.GroupName})
	}

	// 拼接按标签过滤
	if query.Tag != "" {
		// 格式为: %"tagname"%
		builder = builder.Where(sq.Expr("instr(dv.f_tags, ?) > 0", `"`+query.Tag+`"`))
	}

	// 按字段范围过滤
	// if query.FieldScopeStr != "" {
	// 	fieldScope, err := strconv.ParseUint(query.FieldScopeStr, 10, 64)
	// 	if err != nil {
	// 		return builder, err
	// 	}
	// 	builder = builder.Where(sq.Eq{"dv.f_field_scope": uint8(fieldScope)})
	// }

	return builder, nil
}

// 根据数据视图 ID 数组去获取 ID 与数据视图详情的映射关系，只返回视图的重要信息
func (dva *dataViewAccess) GetDetailedDataViewMapByIDs(ctx context.Context, viewIDs []string) (viewMap map[string]*interfaces.DataView, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Get detailed data view map by IDs", trace.WithSpanKind(trace.SpanKindClient))
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("view_ids").String(fmt.Sprintf("%v", viewIDs)),
	)

	// 1. 初始化viewMap
	viewMap = make(map[string]*interfaces.DataView)

	// 2. 生成完整的sql语句和参数列表
	sqlStr, args, err := sq.Select(
		"f_view_id",
		"f_view_name",
		"f_technical_name",
		"f_group_id",
		"f_type",
		"f_query_type",
		"f_data_scope",
		"f_fields",
		"f_builtin",
		"f_data_source_type",
		"f_data_source_id",
		"f_meta_table_name",
	).
		From(DATA_VIEW_TABLE_NAME).
		Where(sq.Eq{"f_view_id": viewIDs}).
		ToSql()

	if err != nil {
		errDetails := fmt.Sprintf("Failed to generate a sql statement using the squirrel sdk before getting detailed data view map by IDs, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return viewMap, err
	}

	// 3. 记录完整的sql语句
	sqlInfo := fmt.Sprintf("The detailed sql statement when getting detailed data view map by IDs is: %v", sqlStr)
	logger.Debug(sqlInfo)
	o11y.Info(ctx, sqlInfo)

	// 4. 执行查询操作
	rows, err := dva.db.Query(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("Failed to get detailed data view map by IDs, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return viewMap, err
	}

	defer rows.Close()

	// 5. Scan查询结果
	for rows.Next() {
		var dataScopeBytes, fieldsBytes []byte
		view := &interfaces.DataView{}
		err = rows.Scan(
			&view.ViewID,
			&view.ViewName,
			&view.TechnicalName,
			&view.GroupID,
			&view.Type,
			&view.QueryType,
			&dataScopeBytes,
			&fieldsBytes,
			&view.Builtin,
			&view.DataSourceType,
			&view.DataSourceID,
			&view.MetaTableName,
		)
		if err != nil {
			errDetails := fmt.Sprintf("Failed to scan row after executing the sql to get detailed data view map by IDs, err: %v", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			return viewMap, err
		}

		if len(dataScopeBytes) != 0 {
			err = sonic.Unmarshal([]byte(dataScopeBytes), &view.DataScope)
			if err != nil {
				errDetails := fmt.Sprintf("Unmarshal dataSource failed, %s", err.Error())
				logger.Error(errDetails)
				o11y.Error(ctx, errDetails)
				span.SetStatus(codes.Error, "Unmarshal dataSource failed")

				return nil, err
			}
		}

		err = sonic.Unmarshal([]byte(fieldsBytes), &view.Fields)
		if err != nil {
			errDetails := fmt.Sprintf("Failed to unmarshal fieldsBytes, err: %v", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			return nil, err
		}

		// 更新viewMap
		viewMap[view.ViewID] = view
	}

	return viewMap, nil
}

// 根据数据视图ID数组去获取ID与数据视图Simple Info的映射关系
func (dva *dataViewAccess) GetSimpleDataViewMapByIDs(ctx context.Context, viewIDs []string) (viewMap map[string]*interfaces.DataView, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven层: 从数据库中获取数据视图simple map(key为数据视图ID, value为数据视图simple对象)", trace.WithSpanKind(trace.SpanKindClient))
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("view_ids").String(fmt.Sprintf("%v", viewIDs)),
	)

	// 1. 初始化viewMap
	viewMap = make(map[string]*interfaces.DataView)

	// 2. 生成完整的sql语句和参数列表
	sqlStr, args, err := sq.Select(
		"f_view_id",
		"f_view_name",
		"f_technical_name",
		"f_query_type",
		"f_group_id",
		"f_builtin",
		"f_data_source_type",
		"f_data_source_id",
	).
		From(DATA_VIEW_TABLE_NAME).
		Where(sq.Eq{"f_view_id": viewIDs}).
		ToSql()

	if err != nil {
		errDetails := fmt.Sprintf("Failed to generate a sql statement using the squirrel sdk before getting simple data view map by ids, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return viewMap, err
	}

	// 3. debug日志级别下, 打印完整的sql语句
	sqlInfo := fmt.Sprintf("The detailed sql statement when getting simple data view map by ids is: %v", sqlStr)
	logger.Debug(sqlInfo)
	o11y.Info(ctx, sqlInfo)

	// 4. 执行查询操作
	rows, err := dva.db.Query(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("Failed to get simple data view map by ids, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return viewMap, err
	}

	defer rows.Close()

	// 5. Scan查询结果
	for rows.Next() {
		view := &interfaces.DataView{}
		err = rows.Scan(
			&view.ViewID,
			&view.ViewName,
			&view.TechnicalName,
			&view.QueryType,
			&view.GroupID,
			&view.Builtin,
			&view.DataSourceType,
			&view.DataSourceID,
		)
		if err != nil {
			errDetails := fmt.Sprintf("Failed to scan row after executing the sql to get simple data view map by ids, err: %v", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			return viewMap, err
		}

		// 更新viewMap
		viewMap[view.ViewID] = view
	}

	return viewMap, nil
}

// 批量根据视图id获取实时订阅任务id，不保证返回的任务id的顺序和传入的视图id的顺序一致
func (dva *dataViewAccess) GetJobsByDataViewIDs(ctx context.Context, viewIDs []string) ([]*interfaces.JobInfo, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Get jobs by data view IDs", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("view_id").String(fmt.Sprintf("%v", viewIDs)),
	)

	sqlStr, args, err := sq.Select("f_job_id").
		From(DATA_VIEW_TABLE_NAME).
		Where(sq.Eq{"f_view_id": viewIDs}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'Get jobs by data view IDs' sql stmt failed, %s", err.Error())
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Generate sql stmt failed")

		return nil, err
	}

	sqlStmt := fmt.Sprintf("sql stmt for Getting job by data view IDs is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	jobs := []*interfaces.JobInfo{}
	rows, err := dva.db.Query(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("Query data views failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Query data views failed")
	}
	defer rows.Close()

	for rows.Next() {
		job := interfaces.JobInfo{}
		err := rows.Scan(&job.JobID)
		if err != nil {
			errDetails := fmt.Sprintf("Scan row failed, %s", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			span.SetStatus(codes.Error, "Scan row failed")

			return nil, err
		}

		jobs = append(jobs, &job)

	}

	span.SetStatus(codes.Ok, "")
	return jobs, nil
}

// 获取分组内所有数据视图的简单信息
func (dva *dataViewAccess) GetSimpleDataViewsByGroupID(ctx context.Context, tx *sql.Tx, groupID string) ([]*interfaces.SimpleDataView, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Select data views by group_id", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("group_id").String(groupID),
	)

	sqlStr, args, err := sq.Select(
		"f_view_id",
		"f_view_name").
		From(DATA_VIEW_TABLE_NAME).
		Where(sq.Eq{
			"f_group_id": groupID,
		}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'get simple data views' sql stmt error: %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Generate sql stmt failed")

		return nil, err
	}

	sqlStmt := fmt.Sprintf("Sql stmt for getting simple data views is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	views := []*interfaces.SimpleDataView{}
	var rows *sql.Rows
	if tx == nil {
		rows, err = dva.db.Query(sqlStr, args...)
	} else {
		rows, err = tx.Query(sqlStr, args...)
	}
	if err != nil {
		errDetails := fmt.Sprintf("Query data views failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Query views by IDs failed")

		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		view := &interfaces.SimpleDataView{}
		err = rows.Scan(
			&view.ViewID,
			&view.ViewName,
		)
		if err != nil {
			errDetails := fmt.Sprintf("Row scan failed, error: %s", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			span.SetStatus(codes.Error, "Row scan failed")

			return nil, err
		}

		views = append(views, view)
	}

	span.SetStatus(codes.Ok, "")
	return views, nil

}

// 导出分组内所有数据视图的详细信息，因为接口可能导出未分组的，因此返回先去掉内置视图
func (dva *dataViewAccess) GetDataViewsByGroupID(ctx context.Context, groupID string) ([]*interfaces.DataView, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Select data views by group_id", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("group_id").String(groupID),
	)

	sqlStr, args, err := sq.Select(
		"dv.f_view_id",
		"dv.f_view_name",
		"dv.f_technical_name",
		"dv.f_group_id",
		"COALESCE(dvg.f_group_name, '')",
		"dv.f_type",
		"dv.f_query_type",
		"dv.f_builtin",
		"dv.f_tags",
		"dv.f_comment",
		"dv.f_data_source_type",
		"dv.f_data_source_id",
		"dv.f_file_name",
		"dv.f_excel_config",
		"dv.f_data_scope",
		"dv.f_fields",
		"dv.f_status",
		"dv.f_metadata_form_id",
		"dv.f_primary_keys",
		"COALESCE(dv.f_sql, '')",
		"dv.f_meta_table_name",
		"dv.f_create_time",
		"dv.f_update_time",
		"dv.f_delete_time",
		"dv.f_creator",
		"dv.f_creator_type",
		"dv.f_updater",
		"dv.f_updater_type",
	).
		From(fmt.Sprintf("%s as dv", DATA_VIEW_TABLE_NAME)).
		LeftJoin(fmt.Sprintf("%s as dvg on dv.f_group_id = dvg.f_group_id", DATA_VIEW_GROUP_TABLE_NAME)).
		// LeftJoin(fmt.Sprintf("%s as dmj on dv.f_job_id = dmj.f_job_id", DATA_MODEL_JOB_TABLE_NAME)).
		Where(sq.Eq{
			"dv.f_group_id": groupID,
			"dv.f_builtin":  interfaces.Non_Builtin,
		}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'get views' sql stmt error: %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Generate sql stmt failed")

		return nil, err
	}

	sqlStmt := fmt.Sprintf("Sql stmt for getting views is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	views := []*interfaces.DataView{}
	rows, err := dva.db.Query(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("Query data views failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Query views by IDs failed")

		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tagsStr, primaryKeysStr string
		var excelConfigBytes, dataScopeBytes, fieldsBytes []byte
		view := &interfaces.DataView{
			ModuleType: interfaces.MODULE_TYPE_DATA_VIEW,
		}
		err = rows.Scan(
			&view.ViewID,
			&view.ViewName,
			&view.TechnicalName,
			&view.GroupID,
			&view.GroupName,
			&view.Type,
			&view.QueryType,
			&view.Builtin,
			&tagsStr,
			&view.Comment,
			&view.DataSourceType,
			&view.DataSourceID,
			&view.FileName,
			&excelConfigBytes,
			&dataScopeBytes,
			&fieldsBytes,
			&view.Status,
			&view.MetadataFormID,
			&primaryKeysStr,
			&view.SQLStr,
			&view.MetaTableName,
			&view.CreateTime,
			&view.UpdateTime,
			&view.DeleteTime,
			&view.Creator.ID,
			&view.Creator.Type,
			&view.Updater.ID,
			&view.Updater.Type,
		)
		if err != nil {
			errDetails := fmt.Sprintf("Row scan failed, error: %s", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			span.SetStatus(codes.Error, "Row scan failed")

			return nil, err
		}

		view.Tags = libCommon.TagString2TagSlice(tagsStr)
		view.PrimaryKeys = libCommon.TagString2TagSlice(primaryKeysStr)

		// 反序列化
		// err := sonic.Unmarshal([]byte(dataSourceBytes), &view.DataSource)
		// if err != nil {
		// 	errDetails := fmt.Sprintf("Unmarshal dataSource failed, %s", err.Error())
		// 	logger.Error(errDetails)
		// 	o11y.Error(ctx, errDetails)
		// 	span.SetStatus(codes.Error, "Unmarshal dataSource failed")

		// 	return nil, err
		// }

		if len(excelConfigBytes) != 0 {
			err = sonic.Unmarshal([]byte(excelConfigBytes), &view.ExcelConfig)
			if err != nil {
				errDetails := fmt.Sprintf("Unmarshal excelConfigBytes failed, %s", err.Error())
				logger.Error(errDetails)
				o11y.Error(ctx, errDetails)
				span.SetStatus(codes.Error, "Unmarshal excelConfigBytes failed")

				return nil, err
			}
		}

		if len(dataScopeBytes) != 0 {
			err = sonic.Unmarshal([]byte(dataScopeBytes), &view.DataScope)
			if err != nil {
				errDetails := fmt.Sprintf("Unmarshal dataScopeBytes failed, %s", err.Error())
				logger.Error(errDetails)
				o11y.Error(ctx, errDetails)
				span.SetStatus(codes.Error, "Unmarshal dataScopeBytes failed")

				return nil, err
			}
		}

		err = sonic.Unmarshal([]byte(fieldsBytes), &view.Fields)
		if err != nil {
			errDetails := fmt.Sprintf("Unmarshal fieldsBytes failed, %s", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			span.SetStatus(codes.Error, "Unmarshal fieldsBytes failed")

			return nil, err
		}

		// err = sonic.Unmarshal([]byte(condBytes), &view.Condition)
		// if err != nil {
		// 	errDetails := fmt.Sprintf("Unmarshal condition failed, %s", err.Error())
		// 	logger.Error(errDetails)
		// 	o11y.Error(ctx, errDetails)
		// 	span.SetStatus(codes.Error, "Unmarshal condition failed")

		// 	return nil, err
		// }

		views = append(views, view)
	}

	span.SetStatus(codes.Ok, "")
	return views, nil
}

func (dva *dataViewAccess) GetDataViewsBySourceID(ctx context.Context, sourceID string) ([]*interfaces.DataView, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Select data views by data_source_id", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("data_source_id").String(sourceID),
	)

	sqlStr, args, err := sq.Select(
		"dv.f_view_id",
		"dv.f_view_name",
		"dv.f_technical_name",
		"dv.f_group_id",
		"COALESCE(dvg.f_group_name, '')",
		"dv.f_type",
		"dv.f_query_type",
		"dv.f_builtin",
		"dv.f_tags",
		"dv.f_comment",
		"dv.f_data_source_type",
		"dv.f_data_source_id",
		"dv.f_file_name",
		"dv.f_excel_config",
		"dv.f_data_scope",
		"dv.f_fields",
		"dv.f_status",
		"dv.f_metadata_form_id",
		"dv.f_primary_keys",
		"COALESCE(dv.f_sql, '')",
		"dv.f_meta_table_name",
		"dv.f_create_time",
		"dv.f_update_time",
		"dv.f_delete_time",
		"dv.f_creator",
		"dv.f_creator_type",
		"dv.f_updater",
		"dv.f_updater_type",
	).
		From(fmt.Sprintf("%s as dv", DATA_VIEW_TABLE_NAME)).
		LeftJoin(fmt.Sprintf("%s as dvg on dv.f_group_id = dvg.f_group_id", DATA_VIEW_GROUP_TABLE_NAME)).
		Where(sq.Eq{"dv.f_data_source_id": sourceID}).ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'get views' sql stmt error: %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Generate sql stmt failed")

		return nil, err
	}

	sqlStmt := fmt.Sprintf("Sql stmt for getting views is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	views := []*interfaces.DataView{}
	rows, err := dva.db.Query(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("Query data views failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Query views by IDs failed")

		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tagsStr, primaryKeysStr string
		var excelConfigBytes, dataScopeBytes, fieldsBytes []byte
		view := &interfaces.DataView{
			ModuleType: interfaces.MODULE_TYPE_DATA_VIEW,
		}
		err = rows.Scan(
			&view.ViewID,
			&view.ViewName,
			&view.TechnicalName,
			&view.GroupID,
			&view.GroupName,
			&view.Type,
			&view.QueryType,
			&view.Builtin,
			&tagsStr,
			&view.Comment,
			&view.DataSourceType,
			&view.DataSourceID,
			&view.FileName,
			&excelConfigBytes,
			&dataScopeBytes,
			&fieldsBytes,
			&view.Status,
			&view.MetadataFormID,
			&primaryKeysStr,
			&view.SQLStr,
			&view.MetaTableName,
			&view.CreateTime,
			&view.UpdateTime,
			&view.DeleteTime,
			&view.Creator.ID,
			&view.Creator.Type,
			&view.Updater.ID,
			&view.Updater.Type,
		)
		if err != nil {
			errDetails := fmt.Sprintf("Row scan failed, error: %s", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			span.SetStatus(codes.Error, "Row scan failed")

			return nil, err
		}

		view.Tags = libCommon.TagString2TagSlice(tagsStr)
		view.PrimaryKeys = libCommon.TagString2TagSlice(primaryKeysStr)

		// 反序列化
		// err := sonic.Unmarshal([]byte(dataSourceBytes), &view.DataSource)
		// if err != nil {
		// 	errDetails := fmt.Sprintf("Unmarshal dataSource failed, %s", err.Error())
		// 	logger.Error(errDetails)
		// 	o11y.Error(ctx, errDetails)
		// 	span.SetStatus(codes.Error, "Unmarshal dataSource failed")

		// 	return nil, err
		// }

		if len(excelConfigBytes) != 0 {
			err := sonic.Unmarshal([]byte(excelConfigBytes), &view.ExcelConfig)
			if err != nil {
				errDetails := fmt.Sprintf("Unmarshal dataSource failed, %s", err.Error())
				logger.Error(errDetails)
				o11y.Error(ctx, errDetails)
				span.SetStatus(codes.Error, "Unmarshal dataSource failed")

				return nil, err
			}
		}

		if len(dataScopeBytes) != 0 {
			err = sonic.Unmarshal([]byte(dataScopeBytes), &view.DataScope)
			if err != nil {
				errDetails := fmt.Sprintf("Unmarshal dataSource failed, %s", err.Error())
				logger.Error(errDetails)
				o11y.Error(ctx, errDetails)
				span.SetStatus(codes.Error, "Unmarshal dataSource failed")

				return nil, err
			}
		}

		if len(fieldsBytes) != 0 {
			err = sonic.Unmarshal([]byte(fieldsBytes), &view.Fields)
			if err != nil {
				errDetails := fmt.Sprintf("Unmarshal fields failed, %s", err.Error())
				logger.Error(errDetails)
				o11y.Error(ctx, errDetails)
				span.SetStatus(codes.Error, "Unmarshal fields failed")

				return nil, err
			}
		}

		// err = sonic.Unmarshal([]byte(condBytes), &view.Condition)
		// if err != nil {
		// 	errDetails := fmt.Sprintf("Unmarshal condition failed, %s", err.Error())
		// 	logger.Error(errDetails)
		// 	o11y.Error(ctx, errDetails)
		// 	span.SetStatus(codes.Error, "Unmarshal condition failed")

		// 	return nil, err
		// }

		views = append(views, view)
	}

	span.SetStatus(codes.Ok, "")
	return views, nil
}
