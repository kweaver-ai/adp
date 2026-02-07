// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_connection

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
	DATA_CONNECTION_TABLE_NAME        = "t_data_connection"
	DATA_CONNECTION_STATUS_TABLE_NAME = "t_data_connection_status"
)

var (
	dcAccessOnce sync.Once
	dcAccess     interfaces.DataConnectionAccess
)

type dataConnectionAccess struct {
	appSetting *common.AppSetting
	db         *sql.DB
}

func NewDataConnectionAccess(appSetting *common.AppSetting) interfaces.DataConnectionAccess {
	dcAccessOnce.Do(func() {
		dcAccess = &dataConnectionAccess{
			appSetting: appSetting,
			db:         libdb.NewDB(&appSetting.DBSetting),
		}
	})

	return dcAccess
}

// 创建数据连接
func (dca *dataConnectionAccess) CreateDataConnection(ctx context.Context,
	tx *sql.Tx, conn *interfaces.DataConnection) (err error) {

	ctx, span := ar_trace.Tracer.Start(ctx,
		"driven层: 往数据库中插入数据连接", trace.WithSpanKind(trace.SpanKindClient))
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

	// 1. 存储前的数据处理
	tagsStr, detailedConfigBytes, err := dca.processBeforeStore(ctx, conn)
	if err != nil {
		return err
	}

	// 2. 生成完整的sql语句和参数列表
	sqlStr, args, err := sq.Insert(DATA_CONNECTION_TABLE_NAME).
		Columns(
			"f_connection_id",
			"f_connection_name",
			"f_tags",
			"f_comment",
			"f_create_time",
			"f_update_time",
			"f_data_source_type",
			"f_config",
			"f_config_md5").
		Values(
			conn.ID,
			conn.Name,
			tagsStr,
			conn.Comment,
			conn.CreateTime,
			conn.UpdateTime,
			conn.DataSourceType,
			detailedConfigBytes,
			conn.DataSourceConfigMD5).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Failed to generate a sql statement using the squirrel sdk before creating a data connection, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return err
	}

	// 3. 记录完整的sql语句
	sqlInfo := fmt.Sprintf("The detailed sql statement when creating a data connection is: %v", sqlStr)
	logger.Debug(sqlInfo)
	o11y.Info(ctx, sqlInfo)

	// 4. 执行插入操作
	_, err = tx.Exec(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("Failed to create a data connection, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return err
	}

	return nil
}

// 批量删除数据连接
// 注: 数据库的批量删除操作具备原子性, 要么都成功, 要么都失败.
func (dca *dataConnectionAccess) DeleteDataConnections(ctx context.Context, tx *sql.Tx, connIDs []string) (err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven层: 从数据库中批量删除数据连接配置", trace.WithSpanKind(trace.SpanKindClient))
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
		attr.Key("conn_ids").String(fmt.Sprintf("%v", connIDs)),
	)

	// 1. 生成完整的sql语句和参数列表
	sqlStr, args, err := sq.Delete(DATA_CONNECTION_TABLE_NAME).
		Where(sq.Eq{"f_connection_id": connIDs}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Failed to generate a sql statement using the squirrel sdk before deleting data connections in batches, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return err
	}

	// 2. 记录完整的sql语句
	sqlInfo := fmt.Sprintf("The detailed sql statement when deleting data connections is: %v", sqlStr)
	logger.Debug(sqlInfo)
	o11y.Info(ctx, sqlInfo)

	// 3. 执行批量删除操作
	_, err = tx.Exec(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("Failed to delete data connections in batches, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return err
	}

	return nil
}

// 修改数据连接
func (dca *dataConnectionAccess) UpdateDataConnection(ctx context.Context, tx *sql.Tx, conn *interfaces.DataConnection) (err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven层: 从数据库中修改数据连接配置", trace.WithSpanKind(trace.SpanKindClient))
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
		attr.Key("conn_id").String(conn.ID),
	)

	// 1. 存储前的数据处理
	tagsStr, detailedConfigBytes, err := dca.processBeforeStore(ctx, conn)
	if err != nil {
		return err
	}

	data := map[string]any{
		"f_connection_name": conn.Name,
		"f_tags":            tagsStr,
		"f_comment":         conn.Comment,
		"f_update_time":     conn.UpdateTime,
		"f_config":          detailedConfigBytes,
		"f_config_md5":      conn.DataSourceConfigMD5,
	}

	// 2. 生成完整的sql语句和参数列表
	sqlStr, args, err := sq.Update(DATA_CONNECTION_TABLE_NAME).
		SetMap(data).
		Where(sq.Eq{"f_connection_id": conn.ID}).
		ToSql()

	if err != nil {
		errDetails := fmt.Sprintf("Failed to generate a sql statement using the squirrel sdk before updating a data connection, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return err
	}

	// 3. 记录完整的sql语句
	sqlInfo := fmt.Sprintf("The detailed sql statement when updating a data connection is: %v", sqlStr)
	logger.Debug(sqlInfo)
	o11y.Info(ctx, sqlInfo)

	// 4. 执行修改操作
	_, err = tx.Exec(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("Failed to update a data connection, err: %v", err.Error())
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

// 查询数据连接
func (dca *dataConnectionAccess) GetDataConnection(ctx context.Context,
	connID string) (*interfaces.DataConnection, bool, error) {

	var err error
	ctx, span := ar_trace.Tracer.Start(ctx, "driven层: 从数据库中查询数据连接详情", trace.WithSpanKind(trace.SpanKindClient))
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
		attr.Key("conn_id").String(connID),
	)

	// 1. 生成完整的sql语句和参数列表
	sqlStr, args, err := sq.Select(
		"dc.f_connection_id",
		"dc.f_connection_name",
		"dc.f_tags",
		"dc.f_comment",
		"dc.f_create_time",
		"dc.f_update_time",
		"dc.f_data_source_type",
		"dc.f_config",
		"dc.f_config_md5",
		"dcs.f_status",
		"dcs.f_detection_time",
	).From(DATA_CONNECTION_TABLE_NAME + " " + "AS dc").
		Join(DATA_CONNECTION_STATUS_TABLE_NAME + " " + "As dcs on dc.f_connection_id = dcs.f_connection_id").
		Where(sq.Eq{"dc.f_connection_id": connID}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Failed to generate a sql statement using the squirrel sdk before getting a data connection, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return &interfaces.DataConnection{}, false, err
	}

	// 2. 记录完整的sql语句
	sqlInfo := fmt.Sprintf("The detailed sql statement when getting a data connection is: %v", sqlStr)
	logger.Debug(sqlInfo)
	o11y.Info(ctx, sqlInfo)

	// 3. 执行查询操作
	row := dca.db.QueryRow(sqlStr, args...)

	// 4. Scan查询结果
	tagsStr := ""
	detailedConfigBytes := []byte(nil)
	conn := interfaces.DataConnection{}
	err = row.Scan(
		&conn.ID,
		&conn.Name,
		&tagsStr,
		&conn.Comment,
		&conn.CreateTime,
		&conn.UpdateTime,
		&conn.DataSourceType,
		&detailedConfigBytes,
		&conn.DataSourceConfigMD5,
		&conn.DataConnectionStatus.Status,
		&conn.DataConnectionStatus.DetectionTime,
	)

	// 5. 根据err的值, 返回不同的结果
	switch err {
	case nil:
		// 返回结果处理
		conn.Tags = libCommon.TagString2TagSlice(tagsStr)
		conn.DataSourceConfig = detailedConfigBytes
		conn.ApplicationScope = interfaces.DataSourceType2ApplicationScope[conn.DataSourceType]
		return &conn, true, nil
	case sql.ErrNoRows:
		// 处理err=sql.ErrNoRows的情况
		span.SetAttributes(attr.Key("no_rows").Bool(true))
		return &interfaces.DataConnection{}, false, nil
	default:
		errDetails := fmt.Sprintf("Failed to scan row after executing the sql to get a data connection, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return &interfaces.DataConnection{}, false, err
	}
}

// 查询数据连接列表
func (dca *dataConnectionAccess) ListDataConnections(ctx context.Context,
	queryParams interfaces.DataConnectionListQueryParams) (entries []*interfaces.DataConnectionListEntry, err error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "driven层: 从数据库中获取数据连接列表", trace.WithSpanKind(trace.SpanKindClient))
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
		attr.Key("name").String(queryParams.Name),
		attr.Key("name_pattern").String(queryParams.NamePattern),
		attr.Key("tag").String(queryParams.Tag),
		attr.Key("offset").String(fmt.Sprintf("%v", queryParams.Offset)),
		attr.Key("limit").String(fmt.Sprintf("%v", queryParams.Limit)),
		attr.Key("sort").String(queryParams.Sort),
		attr.Key("direction").String(queryParams.Direction),
		attr.Key("application_scope").StringSlice(queryParams.ApplicationScope),
	)

	// 1. 拼接列表查询sql
	// 1.1 初始化sql
	sqlBuilder := sq.Select(
		"f_connection_id",
		"f_connection_name",
		"f_tags",
		"f_comment",
		"f_create_time",
		"f_update_time",
		"f_data_source_type",
	).From(DATA_CONNECTION_TABLE_NAME)

	// 1.2 按名称精准/模糊匹配拼接sql语句
	sqlBuilder = dca.extendSQLBuilder(queryParams, sqlBuilder)

	// 1.3 拼接排序sql语句
	sqlBuilder = sqlBuilder.OrderBy(fmt.Sprint(queryParams.Sort, " ", queryParams.Direction))

	// 1.4 拼接分页sql语句
	// 如果limit=-1, 则不分页, 可选范围1-1000
	if queryParams.Limit != -1 {
		sqlBuilder = sqlBuilder.Offset(uint64(queryParams.Offset)).
			Limit(uint64(queryParams.Limit))
	}

	// 1.5 生成完整的sql语句和参数列表
	sqlStr, args, err := sqlBuilder.ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Failed to generate a sql statement using the squirrel sdk before listing data connections, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return nil, err
	}

	// 1.6 debug日志级别下, 打印完整的sql语句
	sqlInfo := fmt.Sprintf("The detailed sql statement when listing data connections is: %v", sqlStr)
	logger.Debug(sqlInfo)
	o11y.Info(ctx, sqlInfo)

	// 2. 执行查询操作
	rows, err := dca.db.Query(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("Failed to list data connections, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return nil, err
	}

	defer rows.Close()

	// 3. Scan查询结果
	entries = make([]*interfaces.DataConnectionListEntry, 0)
	for rows.Next() {
		var tagsStr string
		entry := interfaces.DataConnectionListEntry{}

		err := rows.Scan(
			&entry.ID,
			&entry.Name,
			&tagsStr,
			&entry.Comment,
			&entry.CreateTime,
			&entry.UpdateTime,
			&entry.DataSourceType,
		)
		if err != nil {
			errDetails := fmt.Sprintf("Failed to scan row after executing the sql to list data connections, err: %v", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			return entries, err
		}

		// tags由字符串转为字符串数组格式
		entry.Tags = libCommon.TagString2TagSlice(tagsStr)
		// 补充应用范围
		entry.ApplicationScope = interfaces.DataSourceType2ApplicationScope[entry.DataSourceType]
		entries = append(entries, &entry)
	}

	return entries, nil
}

// 查询数据连接总数
func (dca *dataConnectionAccess) GetDataConnectionTotal(ctx context.Context,
	queryParams interfaces.DataConnectionListQueryParams) (total int, err error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "driven层: 从数据库中获取数据连接总数", trace.WithSpanKind(trace.SpanKindClient))
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
		attr.Key("name").String(queryParams.Name),
		attr.Key("name_pattern").String(queryParams.NamePattern),
		attr.Key("tag").String(queryParams.Tag),
		attr.Key("application_scope").StringSlice(queryParams.ApplicationScope),
	)

	// 1. 生成完整的sql语句和参数列表
	sqlBuilder := sq.Select("COUNT(f_connection_id)").From(DATA_CONNECTION_TABLE_NAME)
	sqlStr, args, err := dca.extendSQLBuilder(queryParams, sqlBuilder).ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Failed to generate a sql statement using the squirrel sdk before getting data connection total, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return 0, err
	}

	// 2. 记录完整的sql语句
	sqlInfo := fmt.Sprintf("The detailed sql statement when getting data connection total is: %v", sqlStr)
	logger.Debug(sqlInfo)
	o11y.Info(ctx, sqlInfo)

	// 3. 执行查询操作
	row := dca.db.QueryRow(sqlStr, args...)

	// 4. Scan查询结果
	err = row.Scan(
		&total,
	)

	// 因为SQL语句是SELECT COUNT(field_name)...的格式, 所以不会出现err=sql.ErrNoRows的情况
	if err != nil {
		errDetails := fmt.Sprintf("Failed to scan row after executing the sql to get data connection total, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return total, err
	}

	return total, nil
}

// 查询数据连接名称和ID的映射关系
func (dca *dataConnectionAccess) GetMapAboutName2ID(ctx context.Context,
	connNames []string) (name2ID map[string]string, err error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "driven层: 从数据库中查询数据连接名称与ID的映射关系",
		trace.WithSpanKind(trace.SpanKindClient))

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
		attr.Key("connection_names").StringSlice(connNames),
	)

	// 1. 初始化map
	name2ID = make(map[string]string)

	// 2. 生成完整的sql语句和参数列表
	sqlStr, args, err := sq.Select(
		"f_connection_id",
		"f_connection_name",
	).
		From(DATA_CONNECTION_TABLE_NAME).
		Where(sq.Eq{"f_connection_name": connNames}).ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Failed to generate a sql statement using the squirrel sdk before getting data connection map about name to id, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return name2ID, err
	}

	// 3. 记录完整的sql语句
	sqlInfo := fmt.Sprintf("The detailed sql statement when getting data connection map about name to id is: %v", sqlStr)
	logger.Debug(sqlInfo)
	o11y.Info(ctx, sqlInfo)

	// 4. 执行查询操作
	rows, err := dca.db.Query(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("Failed to get data connection map about name to id, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return name2ID, err
	}

	defer rows.Close()

	// 5. Scan查询结果
	for rows.Next() {
		var (
			id   string
			name string
		)
		err := rows.Scan(
			&id,
			&name,
		)

		if err != nil {
			errDetails := fmt.Sprintf("Failed to scan row after executing the sql to get data connection map about name to id, err: %v", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			return name2ID, err
		}

		// 更新name2ID
		name2ID[name] = id
	}

	return name2ID, nil
}

// 查询数据连接ID和名称的映射关系
func (dca *dataConnectionAccess) GetMapAboutID2Name(ctx context.Context,
	connIDs []string) (id2Name map[string]string, err error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "driven层: 从数据库中查询数据连接ID与名称的映射关系",
		trace.WithSpanKind(trace.SpanKindClient))
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
		attr.Key("conn_ids").String(fmt.Sprintf("%v", connIDs)),
	)

	// 1. 初始化map
	id2Name = make(map[string]string)

	// 2. 生成完整的sql语句和参数列表
	sqlStr, args, err := sq.Select(
		"f_connection_id",
		"f_connection_name",
	).
		From(DATA_CONNECTION_TABLE_NAME).
		Where(sq.Eq{"f_connection_id": connIDs}).ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Failed to generate a sql statement using the squirrel sdk before getting data connection map about id to name, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return id2Name, err
	}

	// 3. 记录完整的sql语句
	sqlInfo := fmt.Sprintf("The detailed sql statement when getting data connection map about id to name is: %v", sqlStr)
	logger.Debug(sqlInfo)
	o11y.Info(ctx, sqlInfo)

	// 4. 执行查询操作
	rows, err := dca.db.Query(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("Failed to get data connection map about id to name, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return id2Name, err
	}

	defer rows.Close()

	// 5. Scan查询结果
	for rows.Next() {
		var (
			id   string
			name string
		)
		err := rows.Scan(
			&id,
			&name,
		)

		if err != nil {
			errDetails := fmt.Sprintf("Failed to scan row after executing the sql to get data connection map about id to name, err: %v", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			return id2Name, err
		}

		// 更新id2Name
		id2Name[id] = name
	}

	return id2Name, nil
}

// 根据config_md5查询数据连接
func (dca *dataConnectionAccess) GetDataConnectionsByConfigMD5(ctx context.Context,
	configMD5 string) (connMap map[string]*interfaces.DataConnection, err error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "driven层: 根据详细配置的md5从数据库中查询数据连接", trace.WithSpanKind(trace.SpanKindClient))
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
		attr.Key("config_md5").String(configMD5),
	)

	// 1. 初始化connMap
	connMap = make(map[string]*interfaces.DataConnection, 0)

	// 2. 生成完整的sql语句和参数列表
	sqlStr, args, err := sq.Select(
		"f_connection_id",
		"f_connection_name",
		"f_data_source_type",
		"f_config_md5",
	).
		From(DATA_CONNECTION_TABLE_NAME).
		Where(sq.Eq{"f_config_md5": configMD5}).ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Failed to generate a sql statement using the squirrel sdk before getting data connection map by config_md5, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return connMap, err
	}

	// 3. 记录完整的sql语句
	sqlInfo := fmt.Sprintf("The detailed sql statement when getting data connections by config_md5 is: %v", sqlStr)
	logger.Debug(sqlInfo)
	o11y.Info(ctx, sqlInfo)

	// 4. 执行查询操作
	rows, err := dca.db.Query(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("Failed to get data connections by config_md5, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return connMap, err
	}

	defer rows.Close()

	// 5. Scan查询结果
	for rows.Next() {
		conn := interfaces.DataConnection{}
		detailedConfigBytes := []byte(nil)
		err := rows.Scan(
			&conn.ID,
			&conn.Name,
			&conn.DataSourceType,
			&conn.DataSourceConfigMD5,
		)

		if err != nil {
			errDetails := fmt.Sprintf("Failed to scan row after executing the sql to get data connections by data_source_type, err: %v", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			return connMap, err
		}

		// 更新conns
		conn.DataSourceConfig = detailedConfigBytes
		connMap[conn.ID] = &conn
	}

	return connMap, nil
}

// 根据数据连接ID查询数据来源类型
func (dca *dataConnectionAccess) GetDataConnectionSourceType(ctx context.Context, connID string) (sourceType string, isExist bool, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven层: 根据数据连接ID查询数据来源类型", trace.WithSpanKind(trace.SpanKindClient))
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
		attr.Key("conn_id").String(connID),
	)

	// 1. 生成完整的sql语句和参数列表
	sqlStr, args, err := sq.Select("f_data_source_type").
		From(DATA_CONNECTION_TABLE_NAME).
		Where(sq.Eq{"f_connection_id": connID}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Failed to generate a sql statement using the squirrel sdk before getting data connection data_source_type, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return sourceType, false, err
	}

	// 2. 记录完整的sql语句
	sqlInfo := fmt.Sprintf("The detailed sql statement when getting data connection data_source_type is: %v", sqlStr)
	logger.Debug(sqlInfo)
	o11y.Info(ctx, sqlInfo)

	// 3. 执行查询操作
	row := dca.db.QueryRow(sqlStr, args...)

	// 4. Scan查询结果
	err = row.Scan(
		&sourceType,
	)

	// 处理err=sql.ErrNoRows的情况
	if err == sql.ErrNoRows {
		span.SetAttributes(attr.Key("no_rows").Bool(true))
		return sourceType, false, nil
	}

	if err != nil {
		errDetails := fmt.Sprintf("Failed to scan row after executing the sql to get data connection data_source_type, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return sourceType, false, err
	}

	return sourceType, true, nil
}

/*
	私有方法
*/

// 补充数据连接列表查询的sqlBuilder
func (dca *dataConnectionAccess) extendSQLBuilder(queryParams interfaces.DataConnectionListQueryParams, sqlBuilder sq.SelectBuilder) sq.SelectBuilder {
	if queryParams.Name != "" {
		// 按名称精确查询
		sqlBuilder = sqlBuilder.Where(sq.Eq{"f_connection_name": queryParams.Name})
	} else if queryParams.NamePattern != "" {
		// 按名称模糊查询
		sqlBuilder = sqlBuilder.Where(sq.Expr("instr(f_connection_name, ?) > 0", queryParams.NamePattern))
	}

	// 标签过滤
	if queryParams.Tag != "" {
		sqlBuilder = sqlBuilder.Where(sq.Expr("instr(f_tags, ?) > 0", `"`+queryParams.Tag+`"`))
	}

	// 应用对象过滤
	if len(queryParams.ApplicationScope) != 0 {
		dataSourceTypes := make([]string, 0)
		for _, applicationObject := range queryParams.ApplicationScope {
			dataSourceTypes = append(dataSourceTypes,
				interfaces.ApplicationObject2DataSourceTypes[applicationObject]...)
		}
		sqlBuilder = sqlBuilder.Where(sq.Eq{"f_data_source_type": dataSourceTypes})
	}

	return sqlBuilder
}

// 数据插入/更新前的预处理
func (dca *dataConnectionAccess) processBeforeStore(ctx context.Context, conn *interfaces.DataConnection) (string, []byte, error) {
	// 1. 转换tags, 从字符串数组转为字符串存储
	tagsStr := libCommon.TagSlice2TagString(conn.Tags)

	// 2. 序列化detailedConfig, 数据库不能直接存储json结构
	if configBytes, ok := conn.DataSourceConfig.([]byte); !ok {
		detailedConfigBytes, err := sonic.Marshal(conn.DataSourceConfig)
		if err != nil {
			errDetails := fmt.Sprintf("Failed to marshal config, err: %v", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			return tagsStr, detailedConfigBytes, err
		}

		return tagsStr, detailedConfigBytes, nil
	} else {
		return tagsStr, configBytes, nil
	}
}

func (dca *dataConnectionAccess) CreateDataConnectionStatus(ctx context.Context,
	tx *sql.Tx, status interfaces.DataConnectionStatus) (err error) {

	ctx, span := ar_trace.Tracer.Start(ctx,
		"driven层: 往数据库中插入数据连接状态", trace.WithSpanKind(trace.SpanKindClient))
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

	// 1. 生成完整的sql语句和参数列表
	sqlStr, args, err := sq.Insert(DATA_CONNECTION_STATUS_TABLE_NAME).Columns(
		"f_connection_id",
		"f_status",
		"f_detection_time",
	).Values(
		status.ID,
		status.Status,
		status.DetectionTime,
	).ToSql()

	if err != nil {
		errDetails := fmt.Sprintf("Failed to generate a sql statement using the squirrel sdk before creating a data connection status, err: %v", err.Error())
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		return err
	}

	// 2. debug日志级别下, 打印完整的sql语句
	sqlInfo := fmt.Sprintf("The detailed sql statement when creating a data connection status is: %v", sqlStr)
	logger.Debugf(sqlInfo)
	o11y.Info(ctx, sqlInfo)

	// 3. 执行插入操作
	_, err = tx.Exec(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("Failed to create a data connection status, err: %v", err.Error())
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		return err
	}

	return nil
}

func (dca *dataConnectionAccess) DeleteDataConnectionStatuses(ctx context.Context, tx *sql.Tx, connIDs []string) (err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven层: 从数据库中批量删除数据连接状态", trace.WithSpanKind(trace.SpanKindClient))
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
		attr.Key("conn_ids").String(fmt.Sprintf("%v", connIDs)),
	)

	// 1. 生成完整的sql语句和参数列表
	sqlStr, args, err := sq.Delete(DATA_CONNECTION_STATUS_TABLE_NAME).
		Where(sq.Eq{"f_connection_id": connIDs}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Failed to generate a sql statement using the squirrel sdk before deleting data connection statuses in batches, err: %v", err.Error())
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		return err
	}

	// 2. debug日志级别下, 打印完整的sql语句
	sqlInfo := fmt.Sprintf("The detailed sql statement when deleting data connection statuses is: %v", sqlStr)
	logger.Debugf(sqlInfo)
	o11y.Info(ctx, sqlInfo)

	// 3. 执行批量删除操作
	_, err = tx.Exec(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("Failed to delete data connection statuses in batches, err: %v", err.Error())
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		return err
	}

	return nil
}

func (dca *dataConnectionAccess) UpdateDataConnectionStatus(ctx context.Context, tx *sql.Tx, status interfaces.DataConnectionStatus) (err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven层: 从数据库中修改数据连接状态", trace.WithSpanKind(trace.SpanKindClient))
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
		attr.Key("conn_id").String(status.ID),
	)

	// 1. 生成完整的sql语句和参数列表
	sqlStr, args, err := sq.Update(DATA_CONNECTION_STATUS_TABLE_NAME).SetMap(
		map[string]any{
			"f_status":         status.Status,
			"f_detection_time": status.DetectionTime,
		},
	).
		Where(sq.Eq{"f_connection_id": status.ID}).ToSql()

	if err != nil {
		errDetails := fmt.Sprintf("Failed to generate a sql statement using the squirrel sdk before updating a data connection status, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return err
	}

	// 2. 记录完整的sql语句
	sqlInfo := fmt.Sprintf("The detailed sql statement when updating a data connection status is: %v", sqlStr)
	logger.Debug(sqlInfo)
	o11y.Info(ctx, sqlInfo)

	// 3. 执行修改操作
	_, err = tx.Exec(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("Failed to update a data connection status, err: %v", err.Error())
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
