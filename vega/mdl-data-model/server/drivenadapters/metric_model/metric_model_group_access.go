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
	METRIC_MODEL_GROUP_TABLE_NAME = "t_metric_model_group"
)

var (
	mmgAccessOnce sync.Once
	mmgAccess     interfaces.MetricModelGroupAccess
)

type metricModelGroupAccess struct {
	appSetting *common.AppSetting
	db         *sql.DB
}

func NewMetricModelGroupAccess(appSetting *common.AppSetting) interfaces.MetricModelGroupAccess {
	mmgAccessOnce.Do(func() {
		mmgAccess = &metricModelGroupAccess{
			appSetting: appSetting,
			db:         libdb.NewDB(&appSetting.DBSetting),
		}
	})

	return mmgAccess
}

// 创建指标模型分组
func (mmga *metricModelGroupAccess) CreateMetricModelGroup(ctx context.Context, tx *sql.Tx, metricModelGroup interfaces.MetricModelGroup) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Insert into metric model group", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	sqlStr, vals, err := sq.Insert(METRIC_MODEL_GROUP_TABLE_NAME).
		Columns(
			"f_group_id",
			"f_group_name",
			"f_comment",
			"f_create_time",
			"f_update_time",
			"f_builtin",
		).
		Values(
			metricModelGroup.GroupID,
			metricModelGroup.GroupName,
			metricModelGroup.Comment,
			metricModelGroup.CreateTime,
			metricModelGroup.UpdateTime,
			metricModelGroup.Builtin,
		).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of insert model group, error: %s", err.Error())

		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of insert model group, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")

		return err
	}
	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("创建指标模型分组的 sql 语句: %s", sqlStr))
	if tx == nil {
		_, err = mmga.db.Exec(sqlStr, vals...)
	} else {
		_, err = tx.Exec(sqlStr, vals...)
	}
	if err != nil {
		logger.Errorf("insert data error: %v\n", err)
		span.SetStatus(codes.Error, "Insert data error")
		o11y.Error(ctx, fmt.Sprintf("Insert data error: %v ", err))
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 判断分组是否存在
// 根据ID获取分组信息
func (mmga *metricModelGroupAccess) GetMetricModelGroupByID(ctx context.Context,
	groupID string) (interfaces.MetricModelGroup, bool, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("Get metric model group[%s]", groupID), trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	metricModelGroup := interfaces.MetricModelGroup{}
	//查询
	sqlStr, vals, err := sq.Select(
		"f_group_id",
		"f_group_name",
		"f_comment",
		"f_create_time",
		"f_update_time",
		"f_builtin").
		From(METRIC_MODEL_GROUP_TABLE_NAME).
		Where(sq.Eq{"f_group_id": groupID}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of select model group by id, error: %s", err.Error())

		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of select model group by id, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")

		return metricModelGroup, false, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("获取指标模型分组信息的 sql 语句: %s", sqlStr))

	err = mmga.db.QueryRow(sqlStr, vals...).Scan(
		&metricModelGroup.GroupID,
		&metricModelGroup.GroupName,
		&metricModelGroup.Comment,
		&metricModelGroup.CreateTime,
		&metricModelGroup.UpdateTime,
		&metricModelGroup.Builtin,
	)

	if err == sql.ErrNoRows {
		logger.Errorf("query no rows, error: %v \n", err)
		span.SetStatus(codes.Error, fmt.Sprintf("Metric model  group %s not found", groupID))
		o11y.Error(ctx, fmt.Sprintf("Metric model group %s not found, sql err: %v", groupID, err))

		return metricModelGroup, false, nil
	} else if err != nil {
		logger.Errorf("row scan failed, error: %v \n", err)
		span.SetStatus(codes.Error, "Row scan failed")
		o11y.Error(ctx, fmt.Sprintf("Row scan failed, error: %v ", err))

		return interfaces.MetricModelGroup{}, false, err
	}

	span.SetStatus(codes.Ok, "")
	return metricModelGroup, true, nil

}

// 根据分组名称判断分组是否存在（创建重复）
func (mmga *metricModelGroupAccess) GetMetricModelGroupByName(ctx context.Context, tx *sql.Tx, groupName string) (interfaces.MetricModelGroup, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Query metric model group", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	metricModelGroup := interfaces.MetricModelGroup{}
	//查询
	sqlStr, vals, err := sq.Select(
		"f_group_id",
		"f_group_name",
		"f_comment",
		"f_create_time",
		"f_update_time",
		"f_builtin").
		From(METRIC_MODEL_GROUP_TABLE_NAME).
		Where(sq.Eq{"f_group_name": groupName}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of get model group id by group name, error: %s", err.Error())

		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of get model group id by group name, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")

		return metricModelGroup, false, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("根据分组名称判断分组是否存在 sql 语句: %s", sqlStr))

	err = tx.QueryRow(sqlStr, vals...).Scan(
		&metricModelGroup.GroupID,
		&metricModelGroup.GroupName,
		&metricModelGroup.Comment,
		&metricModelGroup.CreateTime,
		&metricModelGroup.UpdateTime,
		&metricModelGroup.Builtin,
	)
	if err == sql.ErrNoRows {
		span.SetAttributes(
			attr.Key("no_rows").Bool(true))
		span.SetStatus(codes.Ok, "")
		return metricModelGroup, false, nil
	} else if err != nil {
		logger.Errorf("row scan failed, err: %v\n", err)

		o11y.Error(ctx, fmt.Sprintf("Row scan failed, err: %v", err))
		span.SetStatus(codes.Error, "Row scan failed ")

		return metricModelGroup, false, err
	}

	span.SetStatus(codes.Ok, "")
	return metricModelGroup, true, nil
}

// 根据分组名称判断分组是否存在（创建重复）
func (mmga *metricModelGroupAccess) CheckMetricModelGroupExist(ctx context.Context, groupName string) (interfaces.MetricModelGroup, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Query metric model group", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	metricModelGroup := interfaces.MetricModelGroup{}
	//查询
	sqlStr, vals, err := sq.Select(
		"f_group_id",
		"f_group_name",
		"f_comment",
		"f_create_time",
		"f_update_time",
		"f_builtin").
		From(METRIC_MODEL_GROUP_TABLE_NAME).
		Where(sq.Eq{"f_group_name": groupName}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of get model group id by group name, error: %s", err.Error())

		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of get model group id by group name, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")

		return metricModelGroup, false, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("根据分组名称判断分组是否存在 sql 语句: %s", sqlStr))

	err = mmga.db.QueryRow(sqlStr, vals...).Scan(
		&metricModelGroup.GroupID,
		&metricModelGroup.GroupName,
		&metricModelGroup.Comment,
		&metricModelGroup.CreateTime,
		&metricModelGroup.UpdateTime,
		&metricModelGroup.Builtin,
	)
	if err == sql.ErrNoRows {
		span.SetAttributes(
			attr.Key("no_rows").Bool(true))
		span.SetStatus(codes.Ok, "")
		return metricModelGroup, false, nil
	} else if err != nil {
		logger.Errorf("row scan failed, err: %v\n", err)

		o11y.Error(ctx, fmt.Sprintf("Row scan failed, err: %v", err))
		span.SetStatus(codes.Error, "Row scan failed ")

		return metricModelGroup, false, err
	}

	span.SetStatus(codes.Ok, "")
	return metricModelGroup, true, nil
}

func (mmga *metricModelGroupAccess) UpdateMetricModelGroup(ctx context.Context, metricModelGroup interfaces.MetricModelGroup) error {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("Update metric model group[%s]", metricModelGroup.GroupID), trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	data := map[string]interface{}{
		"f_group_name":  metricModelGroup.GroupName,
		"f_comment":     metricModelGroup.Comment,
		"f_update_time": metricModelGroup.UpdateTime,
	}

	sqlStr, vals, err := sq.Update(METRIC_MODEL_GROUP_TABLE_NAME).
		SetMap(data).
		Where(sq.Eq{"f_group_id": metricModelGroup.GroupID}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of update model group by group_id, error: %s", err.Error())

		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of update model group by group_id, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return err
	}
	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("修改指标模型分组的 sql 语句: %s", sqlStr))

	ret, err := mmga.db.Exec(sqlStr, vals...)
	if err != nil {
		logger.Errorf("update metric model group error: %v\n", err)
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
		logger.Errorf("UPDATE %s RowsAffected more than 1, RowsAffected is %d, metricModelGroup is %v",
			metricModelGroup.GroupID, RowsAffected, metricModelGroup)

		o11y.Warn(ctx, fmt.Sprintf("Update %s RowsAffected more than 1, RowsAffected is %d, metricModelGroup is %v",
			metricModelGroup.GroupID, RowsAffected, metricModelGroup))
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 查询指标模型分组列表
func (mmga *metricModelGroupAccess) ListMetricModelGroups(ctx context.Context,
	parameter interfaces.ListMetricGroupQueryParams) ([]*interfaces.MetricModelGroup, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "Select metric model groups", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	// subquery := sq.Select("f_group_id", "count(1) AS cnt").
	// 	From(METRIC_MODEL_TABLE_NAME).
	// 	GroupBy("f_group_id")
	subBuilder := sq.Select(
		"f_group_id",
		"f_group_name",
		"f_comment",
		"f_create_time",
		"f_update_time",
		"f_builtin").
		From(METRIC_MODEL_GROUP_TABLE_NAME)
		// JoinClause(subquery.Prefix(" LEFT JOIN (").Suffix(") AS mm ON (mmg.f_group_id = mm.f_group_id)"))

		// 过滤
	subBuilder = buildMetricGroupListQuerySQL(parameter, subBuilder)

	//排序
	subBuilder = subBuilder.OrderBy(fmt.Sprint(parameter.Sort, " ", parameter.Direction))

	//添加分页参数 limit = -1 不分页，可选1-1000
	if parameter.Limit != -1 {
		subBuilder = subBuilder.Offset(uint64(parameter.Offset)).Limit(uint64(parameter.Limit))
	}

	sqlStr, vals, err := subBuilder.ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of select model groups, error: %s", err.Error())

		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of select model groups, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return nil, err
	}
	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("查询指标模型分组列表的 sql 语句: %s; queryParams: %v", sqlStr, parameter))
	rows, err := mmga.db.Query(sqlStr, vals...)
	if err != nil {
		logger.Errorf("list data error: %v\n", err)
		span.SetStatus(codes.Error, "List data error")
		o11y.Error(ctx, fmt.Sprintf("List data error: %v", err))

		return nil, err
	}
	defer rows.Close()

	metricModelGroups := make([]*interfaces.MetricModelGroup, 0)
	for rows.Next() {
		metricModelGroup := &interfaces.MetricModelGroup{}
		err := rows.Scan(
			&metricModelGroup.GroupID,
			&metricModelGroup.GroupName,
			&metricModelGroup.Comment,
			&metricModelGroup.CreateTime,
			&metricModelGroup.UpdateTime,
			&metricModelGroup.Builtin,
		)
		if err != nil {
			logger.Errorf("row scan failed, err: %v \n", err)
			span.SetStatus(codes.Error, "Row scan error")
			o11y.Error(ctx, fmt.Sprintf("Row scan error: %v", err))
			return nil, err
		}
		metricModelGroups = append(metricModelGroups, metricModelGroup)
	}
	span.SetStatus(codes.Ok, "")
	return metricModelGroups, nil
}

func (mmga *metricModelGroupAccess) GetMetricModelGroupsTotal(ctx context.Context, params interfaces.ListMetricGroupQueryParams) (int, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Select metric model groups", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	builder := sq.Select("COUNT(mmg.f_group_id)").
		From(fmt.Sprintf("%s AS mmg", METRIC_MODEL_GROUP_TABLE_NAME))

	// 过滤
	builder = buildMetricGroupListQuerySQL(params, builder)

	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of select model groups total, error: %s", err.Error())

		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of select model groups total, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return 0, err
	}
	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("查询指标模型分组列表的 sql 语句: %s; queryParams: %v", sqlStr, params))

	var total int
	row := mmga.db.QueryRow(sqlStr, vals...)
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

// 删除分组和指标模型
func (mmga *metricModelGroupAccess) DeleteMetricModelGroup(ctx context.Context,
	tx *sql.Tx, groupID string) (int64, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "Delete metric model group from db",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("group_id").String(groupID),
	)

	//删除分组
	sqlStr, vals, err := sq.Delete(METRIC_MODEL_GROUP_TABLE_NAME).
		Where(sq.Eq{"f_group_id": groupID}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of delete model group by group_id, error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of delete model group by group_id, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return 0, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("删除指标模型分组的 sql 语句: %s; 删除的分组id: %s", sqlStr, groupID))

	ret, err := tx.Exec(sqlStr, vals...)
	if err != nil {
		logger.Errorf("delete data error: %v\n", err)
		span.SetStatus(codes.Error, "Delete data error")
		o11y.Error(ctx, fmt.Sprintf("Delete data error: %v ", err))
		return 0, err
	}

	//sql语句影响的行数
	RowsAffected, err := ret.RowsAffected()
	if err != nil {
		logger.Errorf("get RowsAffected error: %v\n", err)
		span.SetStatus(codes.Error, "Get RowsAffected error")
		o11y.Warn(ctx, fmt.Sprintf("Get RowsAffected error: %v ", err))
	}
	logger.Infof("RowsAffected: %d", RowsAffected)

	span.SetStatus(codes.Ok, "")
	return RowsAffected, nil
}

// 拼接列表查询sql语句
func buildMetricGroupListQuerySQL(query interfaces.ListMetricGroupQueryParams, builder sq.SelectBuilder) sq.SelectBuilder {
	if len(query.Builtin) > 0 {
		builder = builder.Where(sq.Eq{"f_builtin": query.Builtin})
	}

	return builder
}
