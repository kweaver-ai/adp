// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_dict

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

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
	DATA_DICT_ITEM_TABLE_NAME = "t_data_dict_item"
)

var (
	ddiAccessOnce sync.Once
	ddiAccess     interfaces.DataDictItemAccess
)

type dataDictItemAccess struct {
	appSetting *common.AppSetting
	db         *sql.DB
}

// 配置db的客户端参数
func NewDictItemAccess(appSetting *common.AppSetting) interfaces.DataDictItemAccess {
	ddiAccessOnce.Do(func() {
		ddiAccess = &dataDictItemAccess{
			appSetting: appSetting,
			db:         libdb.NewDB(&appSetting.DBSetting),
		}
	})
	return ddiAccess
}

// 获取KV字典的所有数据字典项 内部调用
func (ddia *dataDictItemAccess) GetKVDictItems(ctx context.Context, dictID string) ([]map[string]string, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Get KV Dict Items", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	items := []map[string]string{}
	builder := sq.Select(
		"f_item_id as \"id\"",
		"f_item_key as \"key\"",
		"f_item_value as \"value\"",
		"f_comment as \"comment\"").
		From(DATA_DICT_ITEM_TABLE_NAME).
		Where(sq.Eq{"f_dict_id": dictID})

	sqlStr, params, err := builder.ToSql()
	if err != nil {
		logger.Errorf("GetKVDictItems builder to sql failed, error: %v \n", err)
		o11y.Error(ctx, fmt.Sprintf("GetKVDictItems builder to sql failed , error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return items, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("获取KV字典项的 sql 语句: %s;", sqlStr))

	rows, err := ddia.db.Query(sqlStr, params...)
	if err != nil {
		logger.Errorf("GetKVDictItems list data error: %v\n", err)
		span.SetStatus(codes.Error, "Get KV Dict Items error")
		o11y.Error(ctx, fmt.Sprintf("Get KV Dict Items error: %v ", err))
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		dictItem := interfaces.KvDictItem{}
		err := rows.Scan(
			&dictItem.ItemID,
			&dictItem.Key,
			&dictItem.Value,
			&dictItem.Comment,
		)
		if err != nil {
			logger.Errorf("GetKVDictItemsrow scan failed, err: %v \n", err)
			span.SetStatus(codes.Error, "Get KV Dict Items error")
			o11y.Error(ctx, fmt.Sprintf("Get KV Dict Items error: %v ", err))
			return items, err
		}

		// 序列化
		str, err := sonic.Marshal(dictItem)
		if err != nil {
			logger.Errorf("GetKVDictItems kv struct marshal failed, err: %v \n", err)
			span.SetStatus(codes.Error, "Get KV Dict Items error")
			o11y.Error(ctx, fmt.Sprintf("Get KV Dict Items error: %v ", err))
			return items, err
		}
		// 反序列化成map
		m := make(map[string]string)
		err = sonic.Unmarshal([]byte(str), &m)
		if err != nil {
			logger.Errorf("GetKVDictItems Unmarshal failed, err: %v \n", err)
			span.SetStatus(codes.Error, "Get KV Dict Items error")
			o11y.Error(ctx, fmt.Sprintf("Get KV Dict Items error: %v ", err))
			return items, err
		}
		items = append(items, m)
	}

	span.SetStatus(codes.Ok, "")
	return items, nil
}

// 获取dimension字典的所有字典项 内部调用
func (ddia *dataDictItemAccess) GetDimensionDictItems(ctx context.Context, dictStore string, dimension interfaces.Dimension) ([]map[string]string, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Get Dimension Dict Items", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	items := []map[string]string{}
	cols := make([]string, 0)
	for i := 0; i < len(dimension.Keys); i++ {
		str := ""
		str = fmt.Sprintf("%s as \"%s\"", dimension.Keys[i].ID, dimension.Keys[i].Name)
		cols = append(cols, str)
	}
	for i := 0; i < len(dimension.Values); i++ {
		str := ""
		str = fmt.Sprintf("%s as \"%s\"", dimension.Values[i].ID, dimension.Values[i].Name)
		cols = append(cols, str)
	}
	cols = append(cols, "f_comment as \"comment\"")

	builder := sq.Select(cols...).From(dictStore)
	sqlStr, params, err := builder.ToSql()
	if err != nil {
		logger.Errorf("GetDimensionDictItems builder to sql failed, error: %v \n", err)
		o11y.Error(ctx, fmt.Sprintf("GetDimensionDictItems builder to sql failed , error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return items, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("获取维度字典项的 sql 语句: %s;", sqlStr))

	// 处理结果
	rows, err := ddia.db.Query(sqlStr, params...)
	if err != nil {
		logger.Errorf("GetDimensionDictItems list data error: %v\n", err)
		span.SetStatus(codes.Error, "Get Dimension Dict Items error")
		o11y.Error(ctx, fmt.Sprintf("Get Dimension Dict Items error: %v ", err))
		return items, err
	}
	defer rows.Close()

	columns, _ := rows.Columns()
	col := len(columns)
	// 临时存储每行数据
	vals := make([][]byte, col)
	cache := make([]interface{}, col)
	// 为每一列初始化一个指针
	for index := range vals {
		cache[index] = &vals[index]
	}
	for rows.Next() {
		err = rows.Scan(cache...)
		if err != nil {
			logger.Errorf("GetDimensionDictItems row scan failed, err: %v \n", err)
			span.SetStatus(codes.Error, "Get Dimension Dict Items error")
			o11y.Error(ctx, fmt.Sprintf("Get Dimension Dict Items error: %v ", err))
			return items, err
		}
		item := make(map[string]string)
		for i, data := range vals {
			item[columns[i]] = string(data)
		}
		items = append(items, item)
	}

	span.SetStatus(codes.Ok, "")
	return items, nil
}

// 添加维度列
func (ddia *dataDictItemAccess) AddDimensionColumn(ctx context.Context, dictStore string, new interfaces.DimensionItem) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Add Dimension Column from table", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	sqlStr := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s varchar(%d) NOT NULL", dictStore, new.ID, interfaces.CREATE_DATA_DICT_ITEM_SIZE)

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("新增维度存储表列的 sql 语句: %s;", sqlStr))

	_, err := ddia.db.Exec(sqlStr)
	if err != nil {
		logger.Errorf("add column %s error: %v\n", new.Name, err)
		span.SetStatus(codes.Error, "Add column error")
		o11y.Error(ctx, fmt.Sprintf("Add column error: %v ", err))
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 删除维度列
func (ddia *dataDictItemAccess) DropDimensionColumn(ctx context.Context, dictStore string, new interfaces.DimensionItem) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Drop Dimension Column from table", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	sqlStr := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", dictStore, new.ID)

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("删除维度存储表列的 sql 语句: %s;", sqlStr))
	_, err := ddia.db.Exec(sqlStr)
	if err != nil {
		logger.Errorf("drop column %s error: %v\n", new.Name, err)
		span.SetStatus(codes.Error, "Drop column error")
		o11y.Error(ctx, fmt.Sprintf("Drop column error: %v ", err))
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 清空某个KV字典的数据字典项 内部调用
func (ddia *dataDictItemAccess) DeleteDataDictItems(ctx context.Context, dictID string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete data dict items from db", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	builder := sq.Delete(DATA_DICT_ITEM_TABLE_NAME).
		Where(sq.Eq{"f_dict_id": dictID})

	sqlStr, params, err := builder.ToSql()
	if err != nil {
		logger.Errorf("DeleteDataDictItems builder to sql failed, error: %v \n", err)
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of clear items , error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("清空键值字典项的 sql 语句: %s;", sqlStr))

	_, err = ddia.db.Exec(sqlStr, params...)
	if err != nil {
		logger.Errorf("DeleteDataDictItems clear dict items error: %v\n", err)
		span.SetStatus(codes.Error, "Delete data error")
		o11y.Error(ctx, fmt.Sprintf("Delete data error: %v ", err))
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 删除某个维度字典的表 内部调用
func (ddia *dataDictItemAccess) DeleteDimensionTable(ctx context.Context, dictStore string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete data dict items from db", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	sqlStr := fmt.Sprintf("DROP TABLE %s", dictStore)

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("清空维度字典项的 sql 语句: %s;", sqlStr))

	_, err := ddia.db.Exec(sqlStr)
	if err != nil {
		logger.Errorf("drop table %s error: %v\n", dictStore, err)
		span.SetStatus(codes.Error, "Delete data error")
		o11y.Error(ctx, fmt.Sprintf("Delete data error: %v ", err))
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 查看数据字典项列表
func (ddia *dataDictItemAccess) ListDataDictItems(ctx context.Context, dict interfaces.DataDict, dictItemsQuery interfaces.DataDictItemQueryParams) ([]map[string]string, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Select data dict items", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	var builder sq.SelectBuilder
	// kv字典
	if dict.DictType == interfaces.DATA_DICT_TYPE_KV {
		builder = sq.Select(
			"f_item_id as \"id\"",
			"f_item_key as \"key\"",
			"f_item_value as \"value\"",
			"f_comment as \"comment\"").
			From(DATA_DICT_ITEM_TABLE_NAME).
			Where(sq.Eq{"f_dict_id": dict.DictID})
		for _, v := range dictItemsQuery.Patterns {
			if interfaces.DATA_DICT_ITEM_SORT[v.QueryField] != "" && v.QueryPattern != "" {
				builder = builder.Where(sq.Expr("instr("+interfaces.DATA_DICT_ITEM_SORT[v.QueryField]+", ?) > 0", v.QueryPattern))
			}
		}
	} else {
		cols := make([]string, 0)
		cols = append(cols, "f_item_id as \"id\"")
		for i := 0; i < len(dict.Dimension.Keys); i++ {
			cols = append(cols, fmt.Sprintf("%s as \"%s\"", dict.Dimension.Keys[i].ID, dict.Dimension.Keys[i].Name))
		}
		for i := 0; i < len(dict.Dimension.Values); i++ {
			cols = append(cols, fmt.Sprintf("%s as \"%s\"", dict.Dimension.Values[i].ID, dict.Dimension.Values[i].Name))
		}
		cols = append(cols, "f_comment as \"comment\"")
		builder = sq.Select(cols...).From(dict.DictStore)
		// 拼接多个维度查询条件
		for _, pattern := range dictItemsQuery.Patterns {
			for _, k := range dict.Dimension.Keys {
				if k.Name == pattern.QueryField && pattern.QueryPattern != "" {
					builder = builder.Where(sq.Expr("instr("+k.ID+", ?) > 0", pattern.QueryPattern))
				}
			}
			for _, v := range dict.Dimension.Values {
				if v.Name == pattern.QueryField && pattern.QueryPattern != "" {
					builder = builder.Where(sq.Expr("instr("+v.ID+", ?) > 0", pattern.QueryPattern))
				}
			}
		}
	}

	//添加排序参数
	builder = builder.OrderBy(dictItemsQuery.Sort + " " + dictItemsQuery.Direction)
	//添加分页参数 limit = -1 不分页
	if dictItemsQuery.Limit != -1 {
		builder = builder.Offset(uint64(dictItemsQuery.Offset)).
			Limit(uint64(dictItemsQuery.Limit))
	}

	sqlStr, params, err := builder.ToSql()
	if err != nil {
		logger.Errorf("ListDataDictItems builder to sql failed, error: %v \n", err)
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of select dict items, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return nil, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("查询数据字典项列表的 sql 语句: %s; queryParams: %v", sqlStr, dictItemsQuery))

	dataDictItems := make([]map[string]string, 0)
	rows, err := ddia.db.Query(sqlStr, params...)
	if err != nil {
		logger.Errorf("ListDataDictItems list data error: %v\n", err)
		span.SetStatus(codes.Error, "List data error")
		o11y.Error(ctx, fmt.Sprintf("List data error: %v", err))
		return dataDictItems, err
	}
	defer rows.Close()

	columns, _ := rows.Columns()
	col := len(columns)
	// 临时存储每行数据
	vals := make([][]byte, col)
	cache := make([]interface{}, col)
	// 为每一列初始化一个指针
	for index := range vals {
		cache[index] = &vals[index]
	}
	for rows.Next() {
		err = rows.Scan(cache...)
		if err != nil {
			logger.Errorf("ListDataDictItems row scan failed, err: %v \n", err)
			span.SetStatus(codes.Error, "Row scan error")
			o11y.Error(ctx, fmt.Sprintf("Row scan error: %v", err))
			return dataDictItems, err
		}
		item := make(map[string]string)
		for i, data := range vals {
			item[columns[i]] = string(data)
		}
		dataDictItems = append(dataDictItems, item)
	}
	span.SetStatus(codes.Ok, "")
	return dataDictItems, nil
}

// 获取数据字典总数
func (ddia *dataDictItemAccess) GetDictItemTotal(ctx context.Context,
	dict interfaces.DataDict, dictItemsQuery interfaces.DataDictItemQueryParams) (int, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "Select data dict items total number", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	var builder sq.SelectBuilder

	if dict.DictType == interfaces.DATA_DICT_TYPE_KV {
		builder = sq.Select("COUNT(f_item_id)").
			From(DATA_DICT_ITEM_TABLE_NAME).
			Where(sq.Eq{"f_dict_id": dict.DictID})

		for _, v := range dictItemsQuery.Patterns {
			if interfaces.DATA_DICT_ITEM_SORT[v.QueryField] != "" && v.QueryPattern != "" {
				builder = builder.Where(sq.Expr("instr("+interfaces.DATA_DICT_ITEM_SORT[v.QueryField]+", ?) > 0", v.QueryPattern))
			}
		}
	} else {
		builder = sq.Select("COUNT(f_item_id)").From(dict.DictStore)
		// 拼接多个维度查询条件
		for _, pattern := range dictItemsQuery.Patterns {
			for _, k := range dict.Dimension.Keys {
				if k.Name == pattern.QueryField && pattern.QueryPattern != "" {
					builder = builder.Where(sq.Expr("instr("+k.ID+", ?) > 0", pattern.QueryPattern))
				}
			}
			for _, v := range dict.Dimension.Values {
				if v.Name == pattern.QueryField && pattern.QueryPattern != "" {
					builder = builder.Where(sq.Expr("instr("+v.ID+", ?) > 0", pattern.QueryPattern))
				}
			}
		}
	}

	sqlStr, params, err := builder.ToSql()
	if err != nil {
		logger.Errorf("ListDataDictItems builder to sql failed, error: %v \n", err)
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of select data dict items total, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return 0, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("查询数据字典项总数的 sql 语句: %s; queryParams: %v", sqlStr, dictItemsQuery.Patterns))

	rows, err := ddia.db.Query(sqlStr, params...)
	if err != nil {
		logger.Errorf("get dict item total count error: %v\n", err)
		span.SetStatus(codes.Error, "Get data dict item total error")
		o11y.Error(ctx, fmt.Sprintf("Get data dict item total error: %v", err))
		return 0, err
	}
	defer rows.Close()

	total := 0
	for rows.Next() {
		err := rows.Scan(
			&total,
		)
		if err != nil {
			logger.Errorf("scan total failed, err: %v\n", err)
			span.SetStatus(codes.Error, "Get data dict item total error")
			o11y.Error(ctx, fmt.Sprintf("Get data dict item total error: %v", err))
			return 0, err
		}
	}

	span.SetStatus(codes.Ok, "")
	return total, nil
}

// 查询同字典下 key是否存在
func (ddia *dataDictItemAccess) CountDictItemByKey(ctx context.Context,
	dictID string, dictStore string, keys []interfaces.DimensionItem) (int, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("Check DictItem By Key dict[%s]", dictID), trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	var builder sq.SelectBuilder
	// kv字典
	if dictStore == interfaces.DATA_DICT_STORE_DEFAULT {
		builder = sq.Select("count(f_item_id)").From(DATA_DICT_ITEM_TABLE_NAME).
			Where(sq.Eq{"f_dict_id": dictID}).
			Where(sq.Eq{"f_item_key": keys[0].Value})
	} else {
		builder = sq.Select("count(f_item_id)").From(dictStore)
		for _, di := range keys {
			builder = builder.Where(sq.Eq{di.ID: di.Value})
		}
	}

	sqlStr, params, err := builder.ToSql()
	if err != nil {
		logger.Errorf("CountDictItemByKey builder to sql failed, error: %v \n", err)
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of check dict item by key, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return 0, err
	}
	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("获取数据字典项信息的 sql 语句: %s", sqlStr))

	rows, err := ddia.db.Query(sqlStr, params...)
	if err != nil {
		logger.Errorf("check dict item by key error: %v\n", err)
		span.SetStatus(codes.Error, "check dict item by key error")
		o11y.Error(ctx, fmt.Sprintf("check dict item by key error: %v", err))
		return 0, err
	}
	defer rows.Close()

	cnt := 0
	for rows.Next() {
		err = rows.Scan(
			&cnt,
		)
		if err != nil {
			logger.Errorf("row scan failed, err: %v\n", err)
			span.SetStatus(codes.Error, "Row scan failed")
			o11y.Error(ctx, fmt.Sprintf("Row scan failed, error: %v ", err))
			return 0, err
		}
	}

	span.SetStatus(codes.Ok, "")
	return cnt, nil
}

// 查询同字典下某个 key 对应的字典项 ID
func (ddia *dataDictItemAccess) GetDictItemIDByKey(ctx context.Context, dictID string, dictStore string,
	keys []interfaces.DimensionItem) ([]string, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("Get DictItemID By Key dict[%s]", dictID),
		trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)
	defer span.End()

	var builder sq.SelectBuilder
	// kv字典
	if dictStore == interfaces.DATA_DICT_STORE_DEFAULT {
		builder = sq.Select("f_item_id").From(DATA_DICT_ITEM_TABLE_NAME).
			Where(sq.Eq{"f_dict_id": dictID}).
			Where(sq.Eq{"f_item_key": keys[0].Value})
	} else {
		builder = sq.Select("f_item_id").From(dictStore)
		for _, di := range keys {
			builder = builder.Where(sq.Eq{di.ID: di.Value})
		}
	}

	sqlStr, params, err := builder.ToSql()
	if err != nil {
		logger.Errorf("GetDictItemIDByKey builder to sql failed, error: %v \n", err)
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of get dict item by key, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed")
		return nil, err
	}
	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("获取数据字典项信息的 sql 语句: %s", sqlStr))

	rows, err := ddia.db.Query(sqlStr, params...)
	if err != nil {
		logger.Errorf("get dict item by key error: %v\n", err)
		span.SetStatus(codes.Error, "get dict item by key error")
		o11y.Error(ctx, fmt.Sprintf("get dict item by key error: %v", err))
		return nil, err
	}
	defer rows.Close()

	itemIDs := []string{}
	for rows.Next() {
		var itemID string
		err = rows.Scan(
			&itemID,
		)
		if err != nil {
			logger.Errorf("row scan failed, err: %v\n", err)
			span.SetStatus(codes.Error, "Row scan failed")
			o11y.Error(ctx, fmt.Sprintf("Row scan failed, error: %v ", err))
			return nil, err
		}

		itemIDs = append(itemIDs, itemID)
	}

	span.SetStatus(codes.Ok, "")
	return itemIDs, nil
}

// 创建单个数据字典项
func (ddia *dataDictItemAccess) CreateDataDictItem(ctx context.Context,
	dictID string, itemID string, dictStore string, dimension interfaces.Dimension) error {

	ctx, span := ar_trace.Tracer.Start(ctx, "Insert into data dict item", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	var builder sq.InsertBuilder
	// kv字典
	if dictStore == interfaces.DATA_DICT_STORE_DEFAULT {
		builder = sq.Insert(DATA_DICT_ITEM_TABLE_NAME).
			Columns(
				"f_dict_id",
				"f_item_id",
				"f_item_key",
				"f_item_value",
				"f_comment",
			).
			Values(
				dictID,
				itemID,
				dimension.Keys[0].Value,
				dimension.Values[0].Value,
				dimension.Comment,
			)
	} else {
		var vals []interface{}
		builder = sq.Insert(dictStore).
			Columns("f_item_id")
		vals = append(vals, itemID)
		for _, di := range dimension.Keys {
			builder = builder.Columns(di.ID)
			vals = append(vals, di.Value)
		}
		for _, di := range dimension.Values {
			builder = builder.Columns(di.ID)
			vals = append(vals, di.Value)
		}
		builder = builder.Columns("f_comment")
		vals = append(vals, dimension.Comment)
		builder = builder.Values(vals...)
	}

	sqlStr, params, err := builder.ToSql()
	if err != nil {
		logger.Errorf("CreateDataDictItem builder to sql failed, error: %v \n", err)
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of insert data dict item, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("创建数据字典项的 sql 语句: %s", sqlStr))

	_, err = ddia.db.Exec(sqlStr, params...)
	if err != nil {
		logger.Errorf("insert data error: %v\n", err)
		span.SetStatus(codes.Error, "Insert data error")
		o11y.Error(ctx, fmt.Sprintf("Insert data error: %v ", err))
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 根据itemID 获取item
func (ddia *dataDictItemAccess) GetDictItemByItemID(ctx context.Context,
	dictStore string, itemID string) (map[string]string, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("Get data dict item[%s]", itemID), trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	builder := sq.Select("*").
		From(dictStore).
		Where(sq.Eq{"f_item_id": itemID})

	sqlStr, params, err := builder.ToSql()
	if err != nil {
		logger.Errorf("GetDictItemByItemID builder to sql failed, error: %v \n", err)
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of select dict item by id, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return map[string]string{}, err
	}

	rows, err := ddia.db.Query(sqlStr, params...)
	if err == sql.ErrNoRows {
		logger.Errorf("query no rows, error: %v \n", err)
		span.SetStatus(codes.Error, fmt.Sprintf("Data dict item %s not found", itemID))
		o11y.Error(ctx, fmt.Sprintf("Data dict item %s not found, sql err: %v", itemID, err))
		return map[string]string{}, err
	}
	if err != nil {
		logger.Errorf("row scan failed, error: %v \n", err)
		span.SetStatus(codes.Error, "Row scan failed")
		o11y.Error(ctx, fmt.Sprintf("Row scan failed, error: %v ", err))
		return map[string]string{}, err
	}
	defer rows.Close()

	columns, _ := rows.Columns()
	col := len(columns)
	// 临时存储每行数据
	vals := make([][]byte, col)
	cache := make([]interface{}, col)
	// 为每一列初始化一个指针
	for index := range vals {
		cache[index] = &vals[index]
	}

	dataDictItem := make(map[string]string)
	for rows.Next() {
		_ = rows.Scan(cache...)
		item := make(map[string]string)
		for i, data := range vals {
			item[columns[i]] = string(data)
		}
		dataDictItem = item
	}

	span.SetStatus(codes.Ok, "")
	return dataDictItem, nil
}

// 更新单个数据字典项
func (ddia *dataDictItemAccess) UpdateDataDictItem(ctx context.Context,
	dictID string, itemID string, dictStore string, dimension interfaces.Dimension) error {

	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("Update data dict[%s]", dictID), trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	var builder sq.UpdateBuilder
	// kv字典
	if dictStore == interfaces.DATA_DICT_STORE_DEFAULT {
		// 拼接sql
		dataMap := map[string]interface{}{
			"f_item_key":   dimension.Keys[0].Value,
			"f_item_value": dimension.Values[0].Value,
			"f_comment":    dimension.Comment,
		}
		builder = sq.Update(DATA_DICT_ITEM_TABLE_NAME).
			SetMap(dataMap).
			Where(sq.Eq{"f_dict_id": dictID}).
			Where(sq.Eq{"f_item_id": itemID})
	} else {
		builder = sq.Update(dictStore)
		for _, di := range dimension.Keys {
			builder = builder.Set(di.ID, di.Value)
		}
		for _, di := range dimension.Values {
			builder = builder.Set(di.ID, di.Value)
		}
		builder = builder.Set("f_comment", dimension.Comment)
		builder = builder.Where(sq.Eq{"f_item_id": itemID})
	}

	sqlStr, params, err := builder.ToSql()
	if err != nil {
		logger.Errorf("UpdateDataDictItem builder to sql failed, error: %v \n", err)
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of update data dict, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("修改数据字典项的 sql 语句: %s", sqlStr))

	_, err = ddia.db.Exec(sqlStr, params...)
	if err != nil {
		logger.Errorf("update data error: %v\n", err)
		span.SetStatus(codes.Error, "Update data error")
		o11y.Error(ctx, fmt.Sprintf("Update data error: %v ", err))
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 删除单个数据字典项
func (ddia *dataDictItemAccess) DeleteDataDictItem(ctx context.Context,
	dictID string, itemID string, dictStore string) (int64, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "Delete data dict item from db",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("dict_id").String(fmt.Sprintf("%v", dictID)))

	var builder sq.DeleteBuilder

	if dictStore == interfaces.DATA_DICT_STORE_DEFAULT {
		builder = sq.Delete(dictStore).
			Where(sq.Eq{"f_dict_id": dictID}).
			Where(sq.Eq{"f_item_id": itemID})
	} else {
		builder = sq.Delete(dictStore).
			Where(sq.Eq{"f_item_id": itemID})
	}

	sqlStr, params, err := builder.ToSql()
	if err != nil {
		logger.Errorf("DeleteDataDictItem builder to sql failed, error: %v \n", err)
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of delete data dict item by id, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return 0, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("删除数据字典项的 sql 语句: %s", sqlStr))

	ret, err := ddia.db.Exec(sqlStr, params...)
	if err != nil {
		logger.Errorf("delete dict item error: %v\n", err)
		span.SetStatus(codes.Error, "Delete data error")
		o11y.Error(ctx, fmt.Sprintf("Delete data error: %v ", err))
		return 0, err
	}

	//sql语句影响的行数
	RowsAffected, _ := ret.RowsAffected()
	span.SetStatus(codes.Ok, "")

	return RowsAffected, nil
}

// 事务内批量删除某个字典的字典项
func (ddia *dataDictItemAccess) DeleteDataDictItemsByItemIDs(ctx context.Context,
	tx *sql.Tx, dictID string, itemIDs []string, dictStore string) error {

	ctx, span := ar_trace.Tracer.Start(ctx, "Transaction delete data dict items from db", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("dict_id").String(fmt.Sprintf("%v", dictID)),
		attr.Key("item_ids").String(fmt.Sprintf("%v", itemIDs)),
	)

	var builder sq.DeleteBuilder

	if dictStore == interfaces.DATA_DICT_STORE_DEFAULT {
		builder = sq.Delete(dictStore).
			Where(sq.Eq{"f_dict_id": dictID}).
			Where(sq.Eq{"f_item_id": itemIDs})
	} else {
		builder = sq.Delete(dictStore).
			Where(sq.Eq{"f_item_id": itemIDs})
	}

	sqlStr, params, err := builder.ToSql()
	if err != nil {
		logger.Errorf("DeleteDataDictItemsByItemIDs builder to sql failed, %v", err)
		o11y.Error(ctx, fmt.Sprintf("Build the sql of delete data dict items by item ids failed, %v", err))
		span.SetStatus(codes.Error, "Build sql failed")
		return err
	}

	// 记录处理的 sql 字符串
	sqlStmt := fmt.Sprintf("Sql stmt for deleting dict items is %s", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	_, err = tx.Exec(sqlStr, params...)
	if err != nil {
		logger.Errorf("Delete dict items failed, %v", err)
		span.SetStatus(codes.Error, "Delete data items failed")
		o11y.Error(ctx, fmt.Sprintf("Delete data items failed, %v", err))
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (ddia *dataDictItemAccess) CreateKVDictItems(ctx context.Context,
	tx *sql.Tx, dictID string, items []interfaces.KvDictItem) error {

	ctx, span := ar_trace.Tracer.Start(ctx, "Insert into KV data dict items", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	builder := sq.Insert(DATA_DICT_ITEM_TABLE_NAME).
		Columns(
			"f_dict_id",
			"f_item_id",
			"f_item_key",
			"f_item_value",
			"f_comment",
		)
	for i := 0; i < len(items); i++ {
		builder = builder.Values(
			items[i].DictID,
			items[i].ItemID,
			items[i].Key,
			items[i].Value,
			items[i].Comment)
	}

	sqlStr, params, err := builder.ToSql()
	if err != nil {
		logger.Errorf("CreateKVDictItems builder to sql failed, error: %v \n", err)
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of insert KV data dict items, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("批量创建键值字典项的 sql 语句: %s", sqlStr))

	_, err = tx.Exec(sqlStr, params...)
	if err != nil {
		logger.Errorf("CreateKVDictItems insert data error: %v\n", err)
		span.SetStatus(codes.Error, "Insert data error")
		o11y.Error(ctx, fmt.Sprintf("Insert data error: %v ", err))
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (ddia *dataDictItemAccess) CreateDimensionDictItems(ctx context.Context,
	tx *sql.Tx, dictID string, dictStore string, dimensions []interfaces.Dimension) error {

	ctx, span := ar_trace.Tracer.Start(ctx, "Insert into dimension data dict items", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	builder := sq.Insert(dictStore).Columns("f_item_id")
	// 拼接sql列名
	for _, di := range dimensions[0].Keys {
		builder = builder.Columns(di.ID)
	}
	for _, di := range dimensions[0].Values {
		builder = builder.Columns(di.ID)
	}
	builder = builder.Columns("f_comment")

	// 添加参数值
	for i := 0; i < len(dimensions); i++ {
		var vals []interface{}
		vals = append(vals, dimensions[i].ItemID)
		for _, di := range dimensions[i].Keys {
			vals = append(vals, di.Value)
		}
		for _, di := range dimensions[i].Values {
			vals = append(vals, di.Value)
		}
		vals = append(vals, dimensions[i].Comment)

		builder = builder.Values(vals...)
	}

	sqlStr, params, err := builder.ToSql()
	if err != nil {
		logger.Errorf("CreateDimensionDictItems builder to sql failed, error: %v \n", err)
		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of insert dimension data dict items, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("批量创建维度字典项的 sql 语句: %s", sqlStr))

	_, err = tx.Exec(sqlStr, params...)
	if err != nil {
		logger.Errorf("CreateDimensionDictItems insert data error: %v\n", err)
		span.SetStatus(codes.Error, "Insert data error")
		o11y.Error(ctx, fmt.Sprintf("Insert data error: %v ", err))
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}
