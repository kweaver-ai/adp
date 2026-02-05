package data_view

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
	DATA_VIEW_GROUP_TABLE_NAME = "t_data_view_group"
)

var (
	dvgAccessOnce sync.Once
	dvgAccess     interfaces.DataViewGroupAccess
)

type dataViewGroupAccess struct {
	appSetting *common.AppSetting
	db         *sql.DB
}

func NewDataViewGroupAccess(appSetting *common.AppSetting) interfaces.DataViewGroupAccess {
	dvgAccessOnce.Do(func() {
		dvgAccess = &dataViewGroupAccess{
			appSetting: appSetting,
			db:         libdb.NewDB(&appSetting.DBSetting),
		}
	})

	return dvgAccess
}

// 创建数据视图分组
func (dvga *dataViewGroupAccess) CreateDataViewGroup(ctx context.Context, tx *sql.Tx, dataViewGroup *interfaces.DataViewGroup) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Insert data view group into DB", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	sqlStr, args, err := sq.Insert(DATA_VIEW_GROUP_TABLE_NAME).
		Columns(
			"f_group_id",
			"f_group_name",
			"f_create_time",
			"f_update_time",
			"f_builtin",
		).
		Values(
			dataViewGroup.GroupID,
			dataViewGroup.GroupName,
			dataViewGroup.CreateTime,
			dataViewGroup.UpdateTime,
			dataViewGroup.Builtin,
		).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'create data view group' sql stmt failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "generate sql stmt failed")

		return err
	}

	sqlStmt := fmt.Sprintf("Sql stmt for creating data view group is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	if tx == nil {
		_, err = dvga.db.Exec(sqlStr, args...)
	} else {
		_, err = tx.Exec(sqlStr, args...)
	}
	if err != nil {
		errDetails := fmt.Sprintf("Insert data view group failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Insert data view group failed")

		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 删除数据视图分组
func (dvga *dataViewGroupAccess) DeleteDataViewGroup(ctx context.Context, tx *sql.Tx, groupID string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Delete data view group from db", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("group_id").String(groupID),
	)

	sqlStr, args, err := sq.Delete(DATA_VIEW_GROUP_TABLE_NAME).
		Where(sq.Eq{"f_group_id": groupID}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'delete data view group' sql stmt failed: %s", err.Error())
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Generate sql stmt failed")

		return err
	}

	sqlStmt := fmt.Sprintf("Sql stmt for deleting data view group is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	_, err = tx.Exec(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("Delete data view group failed: %s", err.Error())
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Delete data view group failed")

		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 更新数据视图分组
func (dvga *dataViewGroupAccess) UpdateDataViewGroup(ctx context.Context, dataViewGroup *interfaces.DataViewGroup) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Update a data view group from DB", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("group_id").String(dataViewGroup.GroupID),
	)

	data := map[string]any{
		"f_group_name":  dataViewGroup.GroupName,
		"f_update_time": dataViewGroup.UpdateTime,
	}

	sqlStr, args, err := sq.Update(DATA_VIEW_GROUP_TABLE_NAME).
		SetMap(data).
		Where(sq.Eq{"f_group_id": dataViewGroup.GroupID}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'update data view group' sql stmt failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "generate sql stmt failed")

		return err
	}

	sqlStmt := fmt.Sprintf("Sql stmt for updating data view group is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	_, err = dvga.db.Exec(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("Update data view group failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Update data view group failed")

		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 查询数据视图分组列表
func (dvga *dataViewGroupAccess) ListDataViewGroups(ctx context.Context, params *interfaces.ListViewGroupQueryParams) ([]*interfaces.DataViewGroup, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Select data view groups", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("offset").String(fmt.Sprintf("%d", params.Offset)),
		attr.Key("limit").String(fmt.Sprintf("%d", params.Limit)),
		attr.Key("sort").String(params.Sort),
		attr.Key("direction").String(params.Direction),
		attr.Key("builtin").String(fmt.Sprintf("%v", params.Builtin)),
	)

	dataViewGroups := make([]*interfaces.DataViewGroup, 0)

	// subquery := sq.Select(
	// 	"f_group_id",
	// 	"count(1) AS cnt").
	// 	From(DATA_VIEW_TABLE_NAME).
	// 	GroupBy("f_group_id")

	builder := sq.Select(
		"f_group_id",
		"f_group_name",
		"f_create_time",
		"f_update_time",
		"f_builtin").
		From(DATA_VIEW_GROUP_TABLE_NAME)
		// JoinClause(subquery.Prefix(" LEFT JOIN (").Suffix(") AS dv ON (dvg.f_group_id = dv.f_group_id)"))

	// 过滤
	builder = buildViewGroupListQuerySQL(params, builder)

	//排序
	builder = builder.OrderBy(fmt.Sprint(params.Sort, " ", params.Direction))

	//添加分页参数 limit = -1 不分页，可选1-1000
	if params.Limit != -1 {
		builder = builder.Offset(uint64(params.Offset)).
			Limit(uint64(params.Limit))
	}

	sqlStr, args, err := builder.ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'list data view groups' sql stmt failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "generate sql stmt failed")

		return nil, err
	}

	sqlStmt := fmt.Sprintf("Sql stmt for listing data view groups is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	rows, err := dvga.db.Query(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("List data view groups failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "List data view groups failed")

		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		dataViewGroup := &interfaces.DataViewGroup{}
		err := rows.Scan(
			&dataViewGroup.GroupID,
			&dataViewGroup.GroupName,
			&dataViewGroup.CreateTime,
			&dataViewGroup.UpdateTime,
			&dataViewGroup.Builtin,
		)
		if err != nil {
			errDetails := fmt.Sprintf("Row scan failed, err: %s", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			span.SetStatus(codes.Error, "Row scan error")

			return nil, err
		}

		dataViewGroups = append(dataViewGroups, dataViewGroup)
	}

	span.SetStatus(codes.Ok, "")
	return dataViewGroups, nil
}

// 查询数据视图分组总数
func (dvga *dataViewGroupAccess) GetDataViewGroupsTotal(ctx context.Context, params *interfaces.ListViewGroupQueryParams) (int, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Get data view groups total from DB", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("builtin").String(fmt.Sprintf("%v", params.Builtin)),
	)

	builder := sq.Select("COUNT(f_group_id)").
		From(DATA_VIEW_GROUP_TABLE_NAME)

	// 过滤
	builder = buildViewGroupListQuerySQL(params, builder)

	sqlStr, args, err := builder.ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'get view groups total' sql stmt failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Generate sql stmt failed")

		return 0, err
	}

	sqlStmt := fmt.Sprintf("Sql stmt for getting view groups total is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	var total int
	row := dvga.db.QueryRow(sqlStr, args...)
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

// 根据ID获取分组信息
func (dvga *dataViewGroupAccess) GetDataViewGroupByID(ctx context.Context,
	groupID string) (*interfaces.DataViewGroup, bool, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Get data view group by id",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("group_id").String(groupID),
	)

	dataViewGroup := &interfaces.DataViewGroup{}
	sqlStr, args, err := sq.Select(
		"f_group_id",
		"f_group_name",
		"f_create_time",
		"f_update_time",
		"f_builtin",
	).From(DATA_VIEW_GROUP_TABLE_NAME).
		Where(sq.Eq{"f_group_id": groupID}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'get data view group by id' sql stmt failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "generate sql stmt failed")

		return dataViewGroup, false, err
	}

	sqlStmt := fmt.Sprintf("Sql stmt for getting data view group by id is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	err = dvga.db.QueryRow(sqlStr, args...).
		Scan(
			&dataViewGroup.GroupID,
			&dataViewGroup.GroupName,
			&dataViewGroup.CreateTime,
			&dataViewGroup.UpdateTime,
			&dataViewGroup.Builtin,
		)

	if err == sql.ErrNoRows {
		errDetails := fmt.Sprintf("Data view group %s not found", groupID)
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Data view group not found")

		return dataViewGroup, false, nil
	}

	if err != nil {
		errDetails := fmt.Sprintf("Row scan failed, error: %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Row scan failed")

		return dataViewGroup, false, err
	}

	span.SetStatus(codes.Ok, "")
	return dataViewGroup, true, nil
}

// 根据分组名称检查分组是否存在，用于创建导入时，更新groupID
func (dvga *dataViewGroupAccess) CheckDataViewGroupExistByName(ctx context.Context, tx *sql.Tx, groupName string, builtin bool) (*interfaces.DataViewGroup, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Check data view group exist by group name", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("group_name").String(groupName),
		attr.Key("builtin").String(fmt.Sprintf("%v", builtin)),
	)

	dataViewGroup := &interfaces.DataViewGroup{}
	sqlStr, args, err := sq.Select(
		"f_group_id",
		"f_builtin").
		From(DATA_VIEW_GROUP_TABLE_NAME).
		Where(sq.Eq{
			"f_group_name": groupName,
			"f_builtin":    builtin,
		}).
		ToSql()

	if err != nil {
		errDetails := fmt.Sprintf("Generate 'get data view group id by group name' sql stmt failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "generate sql stmt failed")

		return dataViewGroup, false, err
	}

	sqlStmt := fmt.Sprintf("Sql stmt for getting data view group id by group name is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	if tx == nil {
		err = dvga.db.QueryRow(sqlStr, args...).Scan(
			&dataViewGroup.GroupID,
			&dataViewGroup.Builtin,
		)
	} else {
		err = tx.QueryRow(sqlStr, args...).Scan(
			&dataViewGroup.GroupID,
			&dataViewGroup.Builtin,
		)
	}

	if err == sql.ErrNoRows {
		span.SetAttributes(attr.Key("no_rows").Bool(true))
		span.SetStatus(codes.Ok, "")
		return dataViewGroup, false, nil
	}

	if err != nil {
		errDetails := fmt.Sprintf("Row scan failed, error: %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Row scan failed")

		return dataViewGroup, false, err
	}

	span.SetStatus(codes.Ok, "")
	return dataViewGroup, true, nil
}

// 标记删除分组
func (dvga *dataViewGroupAccess) MarkDataViewGroupDeleted(ctx context.Context, tx *sql.Tx, params *interfaces.MarkViewGroupDeletedParams) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Mark data view group deleted", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("group_id").String(params.GroupID),
	)

	updateMap := map[string]interface{}{
		"f_delete_time": params.DeleteTime,
	}

	sqlStr, args, err := sq.Update(DATA_VIEW_GROUP_TABLE_NAME).
		SetMap(updateMap).
		Where(sq.Eq{"f_group_id": params.GroupID}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'mark data view group deleted' sql stmt failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "generate sql stmt failed")

		return err
	}

	sqlStmt := fmt.Sprintf("Sql stmt for marking data view group deleted is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	if tx == nil {
		_, err = dvga.db.Exec(sqlStr, args...)
	} else {
		_, err = tx.Exec(sqlStr, args...)
	}
	if err != nil {
		errDetails := fmt.Sprintf("Mark data view group deleted failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "mark data view group deleted failed")

		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 拼接列表查询sql语句
func buildViewGroupListQuerySQL(query *interfaces.ListViewGroupQueryParams, builder sq.SelectBuilder) sq.SelectBuilder {
	if len(query.Builtin) > 0 {
		builder = builder.Where(sq.Eq{"f_builtin": query.Builtin})
	}

	// 过滤掉被逻辑删除的
	if !query.IncludeDeleted {
		builder = builder.Where(sq.Eq{"f_delete_time": 0})
	}

	return builder
}
