// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package trace_model

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

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
	TRACE_MODEL_TABLE_NAME = "t_trace_model"
)

var (
	tmAccessOnce sync.Once
	tmAccess     interfaces.TraceModelAccess
)

type traceModelAccess struct {
	appSetting *common.AppSetting
	db         *sql.DB
}

func NewTraceModelAccess(appSetting *common.AppSetting) interfaces.TraceModelAccess {
	tmAccessOnce.Do(func() {
		tmAccess = &traceModelAccess{
			appSetting: appSetting,
			db:         libdb.NewDB(&appSetting.DBSetting),
		}
	})

	return tmAccess
}

// 批量创建/导入链路模型
// 注: 数据库的批量插入操作具备原子性, 要么都成功, 要么都失败.
func (tma *traceModelAccess) CreateTraceModels(ctx context.Context, models []interfaces.TraceModel) (err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven层: 往数据库中批量插入链路模型", trace.WithSpanKind(trace.SpanKindClient))
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
	)

	// 1. 初始化sqlBuilder
	sqlBuilder := sq.Insert(TRACE_MODEL_TABLE_NAME).
		Columns(
			"f_model_id",
			"f_model_name",
			"f_tags",
			"f_comment",
			"f_creator",
			"f_creator_type",
			"f_create_time",
			"f_update_time",
			"f_span_source_type",
			"f_span_config",
			"f_enabled_related_log",
			"f_related_log_source_type",
			"f_related_log_config",
		)

	// 2. 遍历每一个链路模型, 追加参数
	for _, model := range models {
		// 2.1 存储前的数据处理
		newModel, err := tma.processBeforeStore(ctx, model)
		if err != nil {
			return err
		}

		// 2.2 追加参数
		sqlBuilder = sqlBuilder.Values(
			newModel.ID,
			newModel.Name,
			newModel.TagsStr,
			newModel.Comment,
			newModel.Creator.ID,
			newModel.Creator.Type,
			newModel.CreateTime,
			newModel.UpdateTime,
			newModel.SpanSourceType,
			newModel.SpanConfigBytes,
			newModel.EnabledRelatedLog,
			newModel.RelatedLogSourceType,
			newModel.RelatedLogConfigBytes,
		)
	}

	// 3. 生成完整的sql语句和参数列表
	sqlStr, args, err := sqlBuilder.ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Failed to generate a sql statement using the squirrel sdk before creating trace models in batches, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return err
	}

	// 4. 记录完整的sql语句
	sqlInfo := fmt.Sprintf("The detailed sql statement when creating trace models is: %v", sqlStr)
	logger.Debug(sqlInfo)
	o11y.Info(ctx, sqlInfo)

	// 5. 执行批量插入操作
	_, err = tma.db.Exec(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("Failed to create trace models in batches, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return err
	}

	return nil
}

// 批量删除链路模型
// 注: 数据库的批量删除操作具备原子性, 要么都成功, 要么都失败.
func (tma *traceModelAccess) DeleteTraceModels(ctx context.Context, modelIDs []string) (err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven层: 从数据库中批量删除链路模型", trace.WithSpanKind(trace.SpanKindClient))
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
		attr.Key("model_ids").String(fmt.Sprintf("%v", modelIDs)),
	)

	// 1. 生成完整的sql语句和参数列表
	sqlStr, args, err := sq.Delete(TRACE_MODEL_TABLE_NAME).
		Where(sq.Eq{"f_model_id": modelIDs}).ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Failed to generate a sql statement using the squirrel sdk before deleting trace models in batches, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return err
	}

	// 2. 记录完整的sql语句
	sqlInfo := fmt.Sprintf("The detailed sql statement when deleting trace models is: %v", sqlStr)
	logger.Debug(sqlInfo)
	o11y.Info(ctx, sqlInfo)

	// 3. 执行批量删除操作
	_, err = tma.db.Exec(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("Failed to delete trace models in batches, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return err
	}

	return nil
}

// 修改链路模型
func (tma *traceModelAccess) UpdateTraceModel(ctx context.Context, model interfaces.TraceModel) (err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven层: 从数据库中修改链路模型", trace.WithSpanKind(trace.SpanKindClient))
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
		attr.Key("model_id").String(model.ID),
	)

	// 1. 存储前的数据处理
	newModel, err := tma.processBeforeStore(ctx, model)
	if err != nil {
		return err
	}

	data := map[string]any{
		"f_model_name":              newModel.Name,
		"f_tags":                    newModel.TagsStr,
		"f_comment":                 newModel.Comment,
		"f_update_time":             newModel.UpdateTime,
		"f_span_source_type":        newModel.SpanSourceType,
		"f_span_config":             newModel.SpanConfigBytes,
		"f_enabled_related_log":     newModel.EnabledRelatedLog,
		"f_related_log_source_type": newModel.RelatedLogSourceType,
		"f_related_log_config":      newModel.RelatedLogConfigBytes,
	}

	// 2. 生成完整的sql语句和参数列表
	sqlStr, args, err := sq.Update(TRACE_MODEL_TABLE_NAME).
		SetMap(data).
		Where(sq.Eq{"f_model_id": newModel.ID}).
		ToSql()

	if err != nil {
		errDetails := fmt.Sprintf("Failed to generate a sql statement using the squirrel sdk before updating a trace model, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return err
	}

	// 3. 记录完整的sql语句
	sqlInfo := fmt.Sprintf("The detailed sql statement when updating a trace model is: %v", sqlStr)
	logger.Debug(sqlInfo)
	o11y.Info(ctx, sqlInfo)

	// 4. 执行修改操作
	_, err = tma.db.Exec(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("Failed to update a trace model, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return err
	}

	// todo: 其实这里应该有一个对于sql执行结果的判断, 包括影响行数和匹配函数.
	// 如果影响行数和匹配行数均为0, 说明待更新的对象已被删除;
	// 反之, 如果影响行数为0, 而匹配行数为1, 说明待更新内容与原内容重复;
	// 但碍于现在的sdk只能获取到影响行数, 所以无法进行上述逻辑判断.

	return nil
}

// 查询/导出链路模型
func (tma *traceModelAccess) GetDetailedTraceModelMapByIDs(ctx context.Context, modelIDs []string) (modelMap map[string]interfaces.TraceModel, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven层: 从数据库中获取链路模型map(key为链路模型ID, value为链路模型完整对象)", trace.WithSpanKind(trace.SpanKindClient))
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
		attr.Key("model_ids").String(fmt.Sprintf("%v", modelIDs)),
	)

	// 1. 初始化modelMap
	modelMap = make(map[string]interfaces.TraceModel)

	// 2. 生成完整的sql语句和参数列表
	sqlStr, args, err := sq.Select(
		"f_model_id",
		"f_model_name",
		"f_tags",
		"f_comment",
		"f_creator",
		"f_creator_type",
		"f_create_time",
		"f_update_time",
		"f_span_source_type",
		"f_span_config",
		"f_enabled_related_log",
		"f_related_log_source_type",
		"f_related_log_config",
	).From(TRACE_MODEL_TABLE_NAME).
		Where(sq.Eq{"f_model_id": modelIDs}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Failed to generate a sql statement using the squirrel sdk before getting trace models, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return modelMap, err
	}

	// 3. 记录完整的sql语句
	sqlInfo := fmt.Sprintf("The detailed sql statement when getting trace models is: %v", sqlStr)
	logger.Debug(sqlInfo)
	o11y.Info(ctx, sqlInfo)

	// 4. 执行查询操作
	rows, err := tma.db.Query(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("Failed to get trace models, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return modelMap, err
	}

	defer rows.Close()

	// 5. Scan查询结果
	for rows.Next() {
		model := interfaces.TraceModel{}

		err = rows.Scan(
			&model.ID,
			&model.Name,
			&model.TagsStr,
			&model.Comment,
			&model.Creator.ID,
			&model.Creator.Type,
			&model.CreateTime,
			&model.UpdateTime,
			&model.SpanSourceType,
			&model.SpanConfigBytes,
			&model.EnabledRelatedLog,
			&model.RelatedLogSourceType,
			&model.RelatedLogConfigBytes,
		)

		if err != nil {
			errDetails := fmt.Sprintf("Failed to scan row after executing the sql to get trace models, err: %v", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			return modelMap, err
		}

		newModel, err := tma.processAfterGet(ctx, model)
		if err != nil {
			return modelMap, err
		}

		// 更新modelMap
		modelMap[model.ID] = newModel
	}

	return modelMap, nil
}

// 查询链路模型列表
func (tma *traceModelAccess) ListTraceModels(ctx context.Context,
	queryParams interfaces.TraceModelListQueryParams) (entries []interfaces.TraceModelListEntry, err error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "driven层: 从数据库中获取链路模型列表", trace.WithSpanKind(trace.SpanKindClient))
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
		attr.Key("name").String(fmt.Sprintf("%v", queryParams.Name)),
		attr.Key("name_pattern").String(fmt.Sprintf("%v", queryParams.NamePattern)),
		attr.Key("tag").String(fmt.Sprintf("%v", queryParams.Tag)),
		attr.Key("span_source_types").String(fmt.Sprintf("%v", queryParams.SpanSourceTypes)),
		attr.Key("offset").String(fmt.Sprintf("%v", queryParams.Offset)),
		attr.Key("limit").String(fmt.Sprintf("%v", queryParams.Limit)),
		attr.Key("sort").String(fmt.Sprintf("%v", queryParams.Sort)),
		attr.Key("direction").String(fmt.Sprintf("%v", queryParams.Direction)),
	)

	// 1. 初始化entries
	entries = make([]interfaces.TraceModelListEntry, 0)

	// 2. 拼接列表查询sql
	// 2.1 初始化sql
	sqlBuilder := sq.Select(
		"f_model_id",
		"f_model_name",
		"f_span_source_type",
		"f_tags",
		"f_comment",
		"f_creator",
		"f_creator_type",
		"f_create_time",
		"f_update_time",
	).From(TRACE_MODEL_TABLE_NAME)

	// 2.2 按名称精准/模糊匹配拼接sql语句
	sqlBuilder = tma.extendSQLBuilder(queryParams, sqlBuilder)

	// 2.3 拼接排序sql语句
	sqlBuilder = sqlBuilder.OrderBy(fmt.Sprint(queryParams.Sort, " ", queryParams.Direction))

	// 2.4 拼接分页sql语句
	// 如果limit=-1, 则不分页, 可选范围1-1000
	// if queryParams.Limit != -1 {
	// 	sqlBuilder = sqlBuilder.Offset(uint64(queryParams.Offset)).Limit(uint64(queryParams.Limit))
	// }

	// 2.5 生成完整的sql语句和参数列表
	sqlStr, args, err := sqlBuilder.ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Failed to generate a sql statement using the squirrel sdk before listing trace models, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return entries, err
	}

	// 2.6 debug日志级别下, 打印完整的sql语句
	sqlInfo := fmt.Sprintf("The detailed sql statement when listing trace models is: %v", sqlStr)
	logger.Debug(sqlInfo)
	o11y.Info(ctx, sqlInfo)

	// 3. 执行查询操作
	rows, err := tma.db.Query(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("Failed to list trace models, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return entries, err
	}

	defer rows.Close()

	// 4. Scan查询结果
	for rows.Next() {
		var tagsStr string
		entry := interfaces.TraceModelListEntry{}

		err := rows.Scan(
			&entry.ModelID,
			&entry.ModelName,
			&entry.SpanSourceType,
			&tagsStr,
			&entry.Comment,
			&entry.Creator.ID,
			&entry.Creator.Type,
			&entry.CreateTime,
			&entry.UpdateTime,
		)
		if err != nil {
			errDetails := fmt.Sprintf("Failed to scan row after executing the sql to list trace models, err: %v", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			return entries, err
		}

		// tags由字符串转为字符串数组格式
		entry.Tags = libCommon.TagString2TagSlice(tagsStr)
		entries = append(entries, entry)
	}

	return entries, nil
}

// 查询链路模型总数
func (tma *traceModelAccess) GetTraceModelTotal(ctx context.Context,
	queryParams interfaces.TraceModelListQueryParams) (total int64, err error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "driven层: 从数据库中获取链路模型总数", trace.WithSpanKind(trace.SpanKindClient))
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
		attr.Key("name").String(fmt.Sprintf("%v", queryParams.Name)),
		attr.Key("name_pattern").String(fmt.Sprintf("%v", queryParams.NamePattern)),
		attr.Key("tag").String(fmt.Sprintf("%v", queryParams.Tag)),
		attr.Key("span_source_types").String(fmt.Sprintf("%v", queryParams.SpanSourceTypes)),
	)

	// 1. 生成完整的sql语句和参数列表
	sqlBuilder := sq.Select("COUNT(f_model_id)").From(TRACE_MODEL_TABLE_NAME)
	sqlStr, args, err := tma.extendSQLBuilder(queryParams, sqlBuilder).ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Failed to generate a sql statement using the squirrel sdk before getting trace model total, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return 0, err
	}

	// 2. 记录完整的sql语句
	sqlInfo := fmt.Sprintf("The detailed sql statement when getting trace model total is: %v", sqlStr)
	logger.Debug(sqlInfo)
	o11y.Info(ctx, sqlInfo)

	// 3. 执行查询操作
	row := tma.db.QueryRow(sqlStr, args...)

	// 4. Scan查询结果
	err = row.Scan(
		&total,
	)

	// 因为SQL语句是SELECT COUNT(field_name)...的格式, 所以不会出现err=sql.ErrNoRows的情况
	if err != nil {
		errDetails := fmt.Sprintf("Failed to scan row after executing the sql to get the trace model total, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return total, err
	}

	return total, nil
}

// 根据链路模型ID数组去获取ID与Name的映射关系
func (tma *traceModelAccess) GetSimpleTraceModelMapByIDs(ctx context.Context, modelIDs []string) (modelMap map[string]interfaces.TraceModel, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven层: 从数据库中获取链路模型simple map(key为链路模型ID, value为链路模型simple对象)", trace.WithSpanKind(trace.SpanKindClient))
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
		attr.Key("model_ids").String(fmt.Sprintf("%v", modelIDs)),
	)

	// 1. 初始化map
	modelMap = make(map[string]interfaces.TraceModel)

	// 2. 生成完整的sql语句和参数列表
	sqlStr, args, err := sq.Select(
		"f_model_id",
		"f_model_name",
	).
		From(TRACE_MODEL_TABLE_NAME).
		Where(sq.Eq{"f_model_id": modelIDs}).ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Failed to generate a sql statement using the squirrel sdk before getting trace model simple map by ids, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return modelMap, err
	}

	// 3. 记录完整的sql语句
	sqlInfo := fmt.Sprintf("The detailed sql statement when getting trace model simple map by ids is: %v", sqlStr)
	logger.Debug(sqlInfo)
	o11y.Info(ctx, sqlInfo)

	// 4. 执行查询操作
	rows, err := tma.db.Query(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("Failed to get trace model simple map by ids, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return modelMap, err
	}

	defer rows.Close()

	// 5. Scan查询结果
	for rows.Next() {
		model := interfaces.TraceModel{}

		err := rows.Scan(
			&model.ID,
			&model.Name,
		)

		if err != nil {
			errDetails := fmt.Sprintf("Failed to scan row after executing the sql to get trace model simple map by ids, err: %v", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			return modelMap, err
		}

		// 更新modelMap
		modelMap[model.ID] = model
	}

	return modelMap, nil
}

// 根据链路模型Name数组去获取Name与ID的映射关系
func (tma *traceModelAccess) GetSimpleTraceModelMapByNames(ctx context.Context, modelNames []string) (modelMap map[string]interfaces.TraceModel, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven层: 从数据库中获取链路模型simple map(key为链路模型名称, value为链路模型simple对象)", trace.WithSpanKind(trace.SpanKindClient))
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
		attr.Key("model_names").String(fmt.Sprintf("%v", modelNames)),
	)

	// 1. 初始化modelMap
	modelMap = make(map[string]interfaces.TraceModel)

	// 2. 生成完整的sql语句和参数列表
	sqlStr, args, err := sq.Select(
		"f_model_id",
		"f_model_name",
	).
		From(TRACE_MODEL_TABLE_NAME).
		Where(sq.Eq{"f_model_name": modelNames}).ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Failed to generate a sql statement using the squirrel sdk before getting trace model simple map by names, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return modelMap, err
	}

	// 3. 记录完整的sql语句
	sqlInfo := fmt.Sprintf("The detailed sql statement when getting trace model simple map by names is: %v", sqlStr)
	logger.Debug(sqlInfo)
	o11y.Info(ctx, sqlInfo)

	// 4. 执行查询操作
	rows, err := tma.db.Query(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("Failed to get trace model simple map by names, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return modelMap, err
	}

	defer rows.Close()

	// 5. Scan查询结果
	for rows.Next() {
		model := interfaces.TraceModel{}

		err := rows.Scan(
			&model.ID,
			&model.Name,
		)

		if err != nil {
			errDetails := fmt.Sprintf("Failed to scan row after executing the sql to get trace model simple map by names, err: %v", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			return modelMap, err
		}

		// 更新modelMap
		modelMap[model.Name] = model
	}

	return modelMap, nil
}

/*
	私有方法
*/

// 补充链路模型列表查询的sqlBuilder
func (tma *traceModelAccess) extendSQLBuilder(queryParams interfaces.TraceModelListQueryParams, sqlBuilder sq.SelectBuilder) sq.SelectBuilder {
	if queryParams.Name != "" {
		// 按名称精确查询
		sqlBuilder = sqlBuilder.Where(sq.Eq{"f_model_name": queryParams.Name})
	} else if queryParams.NamePattern != "" {
		// 按名称模糊查询
		sqlBuilder = sqlBuilder.Where(sq.Expr("instr(f_model_name, ?) > 0", queryParams.NamePattern))
	}

	// 标签过滤
	if queryParams.Tag != "" {
		sqlBuilder = sqlBuilder.Where(sq.Expr("instr(f_tags, ?) > 0", `"`+queryParams.Tag+`"`))
	}

	// span数据来源过滤
	if len(queryParams.SpanSourceTypes) != 0 {
		sqlBuilder = sqlBuilder.Where(sq.Eq{"f_span_source_type": queryParams.SpanSourceTypes})
	}

	return sqlBuilder
}

// 数据插入/更新前的预处理
func (tma *traceModelAccess) processBeforeStore(ctx context.Context, model interfaces.TraceModel) (interfaces.TraceModel, error) {
	// 1. 转换tags, 从字符串数组转为字符串存储
	model.TagsStr = libCommon.TagSlice2TagString(model.Tags)

	// 2. 序列化SpanConfig, 数据库不能直接存储json结构
	spanConfigBytes, err := sonic.Marshal(model.SpanConfig)
	if err != nil {
		errDetails := fmt.Sprintf("Failed to marshal span config, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return model, err
	}
	model.SpanConfigBytes = spanConfigBytes

	// 3. 序列化RelatedLogConfig, 数据库不能直接存储json结构
	relatedLogConfigBytes, err := sonic.Marshal(model.RelatedLogConfig)
	if err != nil {
		errDetails := fmt.Sprintf("Failed to marshal span related log config, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return model, err
	}
	model.RelatedLogConfigBytes = relatedLogConfigBytes

	return model, nil
}

// 数据查询后的处理
func (tma *traceModelAccess) processAfterGet(ctx context.Context, model interfaces.TraceModel) (interfaces.TraceModel, error) {
	// 1. tagsStr转为字符串数组
	model.Tags = libCommon.TagString2TagSlice(model.TagsStr)

	// 2. 反序列化spanConfigBytes
	if model.SpanSourceType == interfaces.SOURCE_TYPE_DATA_VIEW {
		spanConf := interfaces.SpanConfigWithDataView{}
		err := sonic.Unmarshal(model.SpanConfigBytes, &spanConf)
		if err != nil {
			errDetails := fmt.Sprintf("Failed to unmarshal spanConfigBytes after getting trace models, err: %v", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			return model, err
		}
		model.SpanConfig = spanConf
	} else {
		spanConf := interfaces.SpanConfigWithDataConnection{}
		err := sonic.Unmarshal(model.SpanConfigBytes, &spanConf)
		if err != nil {
			errDetails := fmt.Sprintf("Failed to unmarshal spanConfigBytes after getting trace models, err: %v", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			return model, err
		}
		model.SpanConfig = spanConf
	}

	// 3. 反序列化relatedConfigBytes
	if model.EnabledRelatedLog == interfaces.RELATED_LOG_OPEN {
		if model.RelatedLogSourceType == interfaces.SOURCE_TYPE_DATA_VIEW {
			relatedLogConf := interfaces.RelatedLogConfigWithDataView{}
			err := sonic.Unmarshal(model.RelatedLogConfigBytes, &relatedLogConf)
			if err != nil {
				errDetails := fmt.Sprintf("Failed to unmarshal relatedLogConfigBytes after getting trace models, err: %v", err.Error())
				logger.Error(errDetails)
				o11y.Error(ctx, errDetails)
				return model, err
			}
			model.RelatedLogConfig = relatedLogConf
		}
	} else {
		model.RelatedLogConfig = interfaces.RelatedLogConfigWithDataView{}
	}

	return model, nil
}
