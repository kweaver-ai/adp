// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_dict

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
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
	DATA_DICT_TABLE_NAME = "t_data_dict"
	// key维度复合索引名称 前缀表名称保证唯一性
	DICT_INDEX_NAME = "uk_keys"
)

var (
	ddAccessOnce sync.Once
	ddAccess     interfaces.DataDictAccess
)

type dataDictAccess struct {
	appSetting *common.AppSetting
	db         *sql.DB
}

// 配置db的客户端参数
func NewDataDictAccess(appSetting *common.AppSetting) interfaces.DataDictAccess {
	ddAccessOnce.Do(func() {
		ddAccess = &dataDictAccess{
			appSetting: appSetting,
			db:         libdb.NewDB(&appSetting.DBSetting),
		}
	})
	return ddAccess
}

// 分页查询数据字典
func (dda *dataDictAccess) ListDataDicts(ctx context.Context,
	dictQuery interfaces.DataDictQueryParams) ([]interfaces.DataDict, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "Select data dicts", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	dataDicts := make([]interfaces.DataDict, 0)

	builder := sq.Select(
		"f_dict_id",
		"f_dict_name",
		"f_dict_type",
		"f_unique_key",
		"f_dimension",
		"f_tags",
		"f_comment",
		"f_create_time",
		"f_update_time").
		From(DATA_DICT_TABLE_NAME)

	if dictQuery.Name != "" {
		// 名称精确查询
		builder = builder.Where(sq.Eq{"f_dict_name": dictQuery.Name})
	} else if dictQuery.NamePattern != "" {
		// 模糊查询
		builder = builder.Where(sq.Expr("instr(f_dict_name, ?) > 0", dictQuery.NamePattern))
	}

	// tag过滤
	if dictQuery.Tag != "" {
		builder = builder.Where(sq.Expr("instr(f_tags, ?) > 0", `"`+dictQuery.Tag+`"`))
	}

	// type过滤
	if dictQuery.Type != "" {
		builder = builder.Where(sq.Eq{"f_dict_type": dictQuery.Type})
	}

	// 排序
	builder = builder.OrderBy(dictQuery.Sort + " " + dictQuery.Direction)
	// 添加分页参数 limit = -1 不分页
	// if dictQuery.Limit != -1 {
	// 	builder = builder.Offset(uint64(dictQuery.Offset)).
	// 		Limit(uint64(dictQuery.Limit))
	// }
	// 生成sql
	sqlStr, params, err := builder.ToSql()
	if err != nil {
		logger.Errorf("ListDataDicts builder to sql error: %v\n", err)

		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of select dicts, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")

		return dataDicts, err
	}
	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("查询数据字典列表的 sql 语句: %s; queryParams: %v", sqlStr, dictQuery))

	rows, err := dda.db.Query(sqlStr, params...)
	if err != nil {
		logger.Errorf("ListDataDicts list data error: %v\n", err)

		span.SetStatus(codes.Error, "List data error")
		o11y.Error(ctx, fmt.Sprintf("List data error: %v", err))

		return dataDicts, err
	}
	defer rows.Close()

	for rows.Next() {
		tagsStr := ""
		dimensionString := ""
		dict := interfaces.DataDict{}
		err := rows.Scan(
			&dict.DictID,
			&dict.DictName,
			&dict.DictType,
			&dict.UniqueKey,
			&dimensionString,
			&tagsStr,
			&dict.Comment,
			&dict.CreateTime,
			&dict.UpdateTime,
		)
		if err != nil {
			logger.Errorf("ListDataDicts row scan failed, err: %v \n", err)
			span.SetStatus(codes.Error, "Row scan error")
			o11y.Error(ctx, fmt.Sprintf("Row scan error: %v", err))
			return dataDicts, err
		}
		// tags string 转成数组的格式
		dict.Tags = libCommon.TagString2TagSlice(tagsStr)
		// dimension json字符串转结构体
		dimension := interfaces.Dimension{}
		err = sonic.Unmarshal([]byte(dimensionString), &dimension)
		if err != nil {
			logger.Errorf("ListDataDicts Failed to unmarshal dimension, err: %v", err.Error())
			span.SetStatus(codes.Error, "Dimension unmarshal error")
			o11y.Error(ctx, fmt.Sprintf("Dimension string [%s] unmarshal failed, error: %v ", dimensionString, err))
			return dataDicts, err
		}
		dict.Dimension = dimension

		dataDicts = append(dataDicts, dict)
	}

	span.SetStatus(codes.Ok, "")
	return dataDicts, nil
}

// 获取数据字典总数
func (dda *dataDictAccess) GetDictTotal(ctx context.Context,
	dictQuery interfaces.DataDictQueryParams) (int64, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "Select data dicts total number", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	builder := sq.Select("COUNT(f_dict_id)").From(DATA_DICT_TABLE_NAME)

	if dictQuery.Name != "" {
		// 名称精确查询
		builder = builder.Where(sq.Eq{"f_dict_name": dictQuery.Name})
	} else if dictQuery.NamePattern != "" {
		// 模糊查询
		builder = builder.Where(sq.Expr("instr(f_dict_name, ?) > 0", dictQuery.NamePattern))
	}

	// tag过滤
	if dictQuery.Tag != "" {
		builder = builder.Where(sq.Expr("instr(f_tags, ?) > 0", `"`+dictQuery.Tag+`"`))
	}

	// type过滤
	if dictQuery.Type != "" {
		builder = builder.Where(sq.Eq{"f_dict_type": dictQuery.Type})
	}

	sqlStr, params, err := builder.ToSql()
	if err != nil {
		logger.Errorf("GetDictTotal builder to sql error: %v\n", err)

		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of select data dicts total, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")

		return 0, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("查询数据字典总数的 sql 语句: %s; queryParams: %v", sqlStr, dictQuery))

	total := int64(0)
	err = dda.db.QueryRow(sqlStr, params...).Scan(&total)
	if err != nil {
		logger.Errorf("GetDictTotal Scan dict total error: %v\n", err)
		span.SetStatus(codes.Error, "Get data dict total error")
		o11y.Error(ctx, fmt.Sprintf("Get data dict total error: %v", err))
		return 0, err
	}

	span.SetStatus(codes.Ok, "")
	return total, nil
}

// 根据数据字典id获取数据字典信息
func (dda *dataDictAccess) GetDataDictByID(ctx context.Context, dictID string) (interfaces.DataDict, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("Get data dict[%s]", dictID), trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	builder := sq.Select(
		"f_dict_id",
		"f_dict_name",
		"f_dict_type",
		"f_unique_key",
		"f_dimension",
		"f_dict_store",
		"f_tags",
		"f_comment",
		"f_create_time",
		"f_update_time").
		From(DATA_DICT_TABLE_NAME).
		Where(sq.Eq{"f_dict_id": dictID})
	sqlStr, params, err := builder.ToSql()
	if err != nil {
		logger.Errorf("builder to sql failed, error: %v \n", err)
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of select dict by id, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return interfaces.DataDict{}, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("获取数据字典信息的 sql 语句: %s", sqlStr))

	dict := interfaces.DataDict{}
	tagsStr := ""
	dimensionString := ""
	row := dda.db.QueryRow(sqlStr, params...)
	err = row.Scan(
		&dict.DictID,
		&dict.DictName,
		&dict.DictType,
		&dict.UniqueKey,
		&dimensionString,
		&dict.DictStore,
		&tagsStr,
		&dict.Comment,
		&dict.CreateTime,
		&dict.UpdateTime,
	)

	if err == sql.ErrNoRows {
		logger.Errorf("query no rows, error: %v \n", err)
		span.SetStatus(codes.Error, fmt.Sprintf("Data dict %s not found", dictID))
		o11y.Error(ctx, fmt.Sprintf("Data dict %s not found, sql err: %v", dictID, err))
		return interfaces.DataDict{}, err
	}
	if err != nil {
		logger.Errorf("row scan failed, error: %v \n", err)
		span.SetStatus(codes.Error, "Row scan failed")
		o11y.Error(ctx, fmt.Sprintf("Row scan failed, error: %v ", err))
		return interfaces.DataDict{}, err
	}

	// tags string 转成数组的格式
	dict.Tags = libCommon.TagString2TagSlice(tagsStr)

	// dimension json字符串转结构体
	dimension := interfaces.Dimension{}
	err = sonic.Unmarshal([]byte(dimensionString), &dimension)
	if err != nil {
		logger.Errorf("Failed to unmarshal dimension, err: %v", err.Error())
		span.SetStatus(codes.Error, "Dimension string transfers to struct error")
		o11y.Error(ctx, fmt.Sprintf("Dimension string [%s] to struct failed, error: %v ", dimensionString, err))
		return interfaces.DataDict{}, err
	}
	dict.Dimension = dimension

	span.SetStatus(codes.Ok, "")
	return dict, nil
}

// 根据名称查询字典是否存在
func (dda *dataDictAccess) CheckDictExistByName(ctx context.Context, dictName string) (bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("Check data dict[%s]", dictName), trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	builder := sq.Select(
		"f_dict_id",
		"f_dict_name").
		From(DATA_DICT_TABLE_NAME).
		Where(sq.Eq{"f_dict_name": dictName})

	sqlStr, params, err := builder.ToSql()
	if err != nil {
		logger.Errorf("builder to sql failed, error: %v \n", err)
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of select dict by name, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return false, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("获取数据字典信息的 sql 语句: %s", sqlStr))

	dictInfo := interfaces.DataDict{}
	row := dda.db.QueryRow(sqlStr, params...)
	err = row.Scan(
		&dictInfo.DictID,
		&dictInfo.DictName,
	)

	if err == sql.ErrNoRows {
		span.SetStatus(codes.Error, fmt.Sprintf("Data dict %s not found", dictName))
		o11y.Error(ctx, fmt.Sprintf("Data dict %s not found, sql err: %v", dictName, err))
		return false, nil
	}
	if err != nil {
		logger.Errorf("row scan failed, err: %v\n", err)
		span.SetStatus(codes.Error, "Row scan failed")
		o11y.Error(ctx, fmt.Sprintf("Row scan failed, error: %v ", err))
		return false, err
	}

	span.SetStatus(codes.Ok, "")
	return true, nil
}

// 创建数据字典
func (dda *dataDictAccess) CreateDataDict(ctx context.Context, tx *sql.Tx, dict interfaces.DataDict) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Insert into data dict", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	// 维度结构体
	dimen, err := sonic.Marshal(dict.Dimension)
	if err != nil {
		logger.Errorf("CreateDataDict Failed to marshal Dimension, err: %v", err.Error())
		span.SetStatus(codes.Error, "Failed to marshal Dimension! ")
		o11y.Error(ctx, fmt.Sprintf("CreateDataDict Failed to marshal Dimension, error: %v ", err))
		return err
	}

	// tags 转成 string 的格式
	tagsStr := libCommon.TagSlice2TagString(dict.Tags)

	// 拼接sql
	builder := sq.Insert(DATA_DICT_TABLE_NAME).
		Columns(
			"f_dict_id",
			"f_dict_name",
			"f_dict_type",
			"f_unique_key",
			"f_dimension",
			"f_dict_store",
			"f_tags",
			"f_comment",
			"f_create_time",
			"f_update_time").
		Values(
			dict.DictID,
			dict.DictName,
			dict.DictType,
			dict.UniqueKey,
			dimen,
			dict.DictStore,
			tagsStr,
			dict.Comment,
			dict.CreateTime,
			dict.UpdateTime)

	sqlStr, params, err := builder.ToSql()
	if err != nil {
		logger.Errorf("CreateDataDict builder to sql failed, error: %v \n", err)
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of insert data dict, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("创建数据字典的 sql 语句: %s", sqlStr))

	_, err = tx.Exec(sqlStr, params...)
	if err != nil {
		logger.Errorf("CreateDataDict insert data error: %v\n", err)
		span.SetStatus(codes.Error, "Insert data error")
		o11y.Error(ctx, fmt.Sprintf("Insert data error: %v ", err))
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 修改数据字典
func (dda *dataDictAccess) UpdateDataDict(ctx context.Context, dict interfaces.DataDict) error {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("Update data dict[%s]", dict.DictID), trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	// 维度结构体转换
	dimen, err := sonic.Marshal(dict.Dimension)
	if err != nil {
		logger.Errorf("UpdateDataDict Failed to marshal Dimension, err: %v", err.Error())
		span.SetStatus(codes.Error, "Data dict marshal error")
		o11y.Error(ctx, fmt.Sprintf("The struct of data dict marshal error: %v ", err))
		return err
	}

	// tags 转成 string 的格式
	tagsStr := libCommon.TagSlice2TagString(dict.Tags)

	// 拼接sql
	dataMap := map[string]interface{}{
		"f_dict_name":   dict.DictName,
		"f_tags":        tagsStr,
		"f_comment":     dict.Comment,
		"f_dimension":   dimen,
		"f_update_time": dict.UpdateTime,
	}

	builder := sq.Update(DATA_DICT_TABLE_NAME).
		SetMap(dataMap).
		Where(sq.Eq{"f_dict_id": dict.DictID})

	sqlStr, params, err := builder.ToSql()
	if err != nil {
		logger.Errorf("UpdateDataDict builder to sql failed, error: %v \n", err)
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of update data dict, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("修改数据字典的 sql 语句: %s", sqlStr))

	ret, err := dda.db.Exec(sqlStr, params...)
	if err != nil {
		logger.Errorf("UpdateDataDict update dict error: %v\n", err)
		span.SetStatus(codes.Error, "Update data error")
		o11y.Error(ctx, fmt.Sprintf("Update data error: %v ", err))
		return err
	}

	//sql语句影响的行数
	RowsAffected, err := ret.RowsAffected()
	if err != nil {
		logger.Errorf("UpdateDataDict Get RowsAffected error: %v\n", err)
		o11y.Error(ctx, fmt.Sprintf("Get RowsAffected error: %v ", err))
		return err
	}
	if RowsAffected > 1 {
		logger.Errorf("UPDATE %d RowsAffected more than 1, RowsAffected is %d, dict is %v", dict.DictID, RowsAffected, dict)
		err = errors.New("UpdateDataDict update Dict RowsAffected more than 1")
		o11y.Warn(ctx, fmt.Sprintf("Update %s RowsAffected more than 1, RowsAffected is %d, dict is %v",
			dict.DictID, RowsAffected, dict))
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 删除数据字典
func (dda *dataDictAccess) DeleteDataDict(ctx context.Context, dictID string) (int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete data dict from db", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("dict_id").String(fmt.Sprintf("%v", dictID)))
	defer span.End()

	builder := sq.Delete(DATA_DICT_TABLE_NAME).
		Where(sq.Eq{"f_dict_id": dictID})

	sqlStr, params, err := builder.ToSql()
	if err != nil {
		logger.Errorf("DeleteDataDict builder to sql failed, error: %v \n", err)
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of delete data dict by id, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return 0, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("删除数据字典的 sql 语句: %s", sqlStr))

	ret, err := dda.db.Exec(sqlStr, params...)
	if err != nil {
		logger.Errorf("DeleteDataDict delete data error: %v\n", err)
		span.SetStatus(codes.Error, "Delete data error")
		o11y.Error(ctx, fmt.Sprintf("Delete data error: %v ", err))
		return 0, err
	}

	//sql语句影响的行数
	RowsAffected, _ := ret.RowsAffected()
	logger.Infof("DeleteDataDict RowsAffected: %d", RowsAffected)

	span.SetStatus(codes.Ok, "")
	return RowsAffected, nil
}

// 创建dimension字典的数据库表
func (dda *dataDictAccess) CreateDimensionDictStore(ctx context.Context, dictStore string, dimension interfaces.Dimension) error {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("Create Dimension Dict Store[%s]", dictStore), trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	sqlStr := fmt.Sprintf("CREATE TABLE %s (f_item_id varchar(40) NOT NULL COMMENT '数据字典项id', ", dictStore)
	for _, dk := range dimension.Keys {
		sqlStr = sqlStr + fmt.Sprintf("%s TEXT NOT NULL COMMENT '%s', ", dk.ID, dk.Name)
	}
	for _, dv := range dimension.Values {
		sqlStr = sqlStr + fmt.Sprintf("%s TEXT NOT NULL COMMENT '%s', ", dv.ID, dv.Name)
	}
	sqlStr = sqlStr + fmt.Sprintf("f_comment varchar(%d) DEFAULT NULL COMMENT '多维度数据字典项说明')", interfaces.COMMENT_MAX_LENGTH)

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("创建维度字典存储表的 sql 语句: %s", sqlStr))

	_, err := dda.db.Exec(sqlStr)
	if err != nil {
		logger.Errorf("exec create dict sql error: %v\n", err)
		span.SetStatus(codes.Error, "Create dimension dict store error")
		o11y.Error(ctx, fmt.Sprintf("Create dimension dict store error: %v ", err))
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 创建dimension字典的索引
// 需sql适配
func (dda *dataDictAccess) AddDimensionIndex(ctx context.Context, dictStore string, keys []interfaces.DimensionItem) error {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("Add Index to Store[%s]", dictStore), trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	sqlStr := ""
	indexName := dictStore + "_" + DICT_INDEX_NAME
	dbType := strings.ToUpper(libdb.GetDBType())
	switch dbType {
	case "DM8":
		sqlStr = fmt.Sprintf("CREATE UNIQUE INDEX %s ON %s(", indexName, dictStore)
		for _, k := range keys {
			sqlStr = sqlStr + fmt.Sprintf("%s,", k.ID)
		}
	case "MYSQL", "MARIADB", "TIDB":
		sqlStr = fmt.Sprintf("ALTER TABLE %s ADD UNIQUE INDEX %s(", dictStore, indexName)
		for _, k := range keys {
			sqlStr = sqlStr + fmt.Sprintf("%s,", k.ID)
		}
	default:
		logger.Errorf("Unsupported DataBase Type")
		err := errors.New("unsupported database type")
		span.SetStatus(codes.Error, "database type error")
		o11y.Error(ctx, fmt.Sprintf("database type error: %v ", err))
		return err
	}
	sqlStr = sqlStr[:len(sqlStr)-1] + ")"

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("添加索引的 sql 语句: %s", sqlStr))

	_, err := dda.db.Exec(sqlStr)
	if err != nil {
		logger.Errorf("exec add index sql error: %v\n", err)
		span.SetStatus(codes.Error, "Add index error")
		o11y.Error(ctx, fmt.Sprintf("Add index error: %v ", err))
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 删除dimension字典的索引
// 需sql适配
func (dda *dataDictAccess) DropDimensionIndex(ctx context.Context, dictStore string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("Drop Index to Store[%s]", dictStore), trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	sqlStr := ""
	indexName := dictStore + "_" + DICT_INDEX_NAME
	dbType := strings.ToUpper(libdb.GetDBType())
	switch dbType {
	case "DM8":
		sqlStr = fmt.Sprintf("DROP INDEX %s", indexName)
	case "MYSQL", "MARIADB", "TIDB":
		sqlStr = fmt.Sprintf("ALTER TABLE %s DROP INDEX %s", dictStore, indexName)
	default:
		logger.Errorf("Unsupported DataBase Type")
		err := errors.New("unsupported database type")
		span.SetStatus(codes.Error, "database type error")
		o11y.Error(ctx, fmt.Sprintf("database type error: %v ", err))
		return err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("删除索引的 sql 语句: %s", sqlStr))

	_, err := dda.db.Exec(sqlStr)
	if err != nil {
		logger.Errorf("exec drop index sql error: %v\n", err)
		span.SetStatus(codes.Error, "Drop index error")
		o11y.Error(ctx, fmt.Sprintf("Drop index error: %v ", err))
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 设置数据字典表更新时间
func (dda *dataDictAccess) UpdateDictUpdateTime(ctx context.Context, dictID string, updateTime int64) error {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("Update Dict UpdateTime[%s]", dictID), trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	builder := sq.Update(DATA_DICT_TABLE_NAME).
		Set("f_update_time", updateTime).
		Where(sq.Eq{"f_dict_id": dictID})

	sqlStr, params, err := builder.ToSql()
	if err != nil {
		logger.Errorf("UpdateDictUpdateTime builder to sql failed, error: %v \n", err)
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of UpdateDictUpdateTime, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed")
		return err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("修改数据字典时间的 sql 语句: %s", sqlStr))

	_, err = dda.db.Exec(sqlStr, params...)
	if err != nil {
		logger.Errorf("updatetime error: %v\n", err)
		span.SetStatus(codes.Error, "updatetime error")
		o11y.Error(ctx, fmt.Sprintf("updatetime error: %v ", err))
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}
