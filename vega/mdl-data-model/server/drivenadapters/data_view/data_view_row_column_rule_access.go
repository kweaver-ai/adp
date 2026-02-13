// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_view

import (
	"context"
	"database/sql"
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
	DATA_VIEW_ROW_COLUMN_RULE_TABLE_NAME = "t_data_view_row_column_rule"
)

var (
	dvrcrAccessOnce sync.Once
	dvrcrAccess     interfaces.DataViewRowColumnRuleAccess
)

type dataViewRowColumnRuleAccess struct {
	appSetting *common.AppSetting
	db         *sql.DB
}

func NewDataViewRowColumnRuleAccess(appSetting *common.AppSetting) interfaces.DataViewRowColumnRuleAccess {
	dvrcrAccessOnce.Do(func() {
		dvrcrAccess = &dataViewRowColumnRuleAccess{
			appSetting: appSetting,
			db:         libdb.NewDB(&appSetting.DBSetting),
		}
	})

	return dvrcrAccess
}

// 创建数据视图行列规则
func (dva *dataViewRowColumnRuleAccess) CreateDataViewRowColumnRules(ctx context.Context, rules []*interfaces.DataViewRowColumnRule) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Insert data view row column rules into DB", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	builder := sq.Insert(DATA_VIEW_ROW_COLUMN_RULE_TABLE_NAME).
		Columns(
			"f_rule_id",
			"f_rule_name",
			"f_view_id",
			"f_tags",
			"f_comment",
			"f_fields",
			"f_row_filters",
			"f_create_time",
			"f_update_time",
			"f_creator",
			"f_creator_type",
			"f_updater",
			"f_updater_type",
		)

	for _, rule := range rules {
		tagsStr := libCommon.TagSlice2TagString(rule.Tags)

		fieldsBytes, err := sonic.Marshal(rule.Fields)
		if err != nil {
			errDetails := fmt.Sprintf("Marshal fields failed, %s", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			span.SetStatus(codes.Error, "Marshal fields failed")

			return err
		}

		rowFiltersBytes, err := sonic.Marshal(rule.RowFilters)
		if err != nil {
			errDetails := fmt.Sprintf("Marshal row filters failed, %s", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			span.SetStatus(codes.Error, "Marshal row filters failed")

			return err
		}

		builder = builder.Values(
			rule.RuleID,
			rule.RuleName,
			rule.ViewID,
			tagsStr,
			rule.Comment,
			fieldsBytes,
			rowFiltersBytes,
			rule.CreateTime,
			rule.UpdateTime,
			rule.Creator.ID,
			rule.Creator.Type,
			rule.Updater.ID,
			rule.Updater.Type,
		)
	}

	sqlStr, args, err := builder.ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'create row column rules' sql stmt failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Generate sql stmt failed")

		return err
	}

	sqlStmt := fmt.Sprintf("sql stmt for creating row column rules is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	_, err = dva.db.Exec(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("insert row column rules failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Insert row column rules failed")

		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 删除数据视图行列规则
func (dva *dataViewRowColumnRuleAccess) DeleteDataViewRowColumnRules(ctx context.Context, tx *sql.Tx, ruleIDs []string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Delete data view row column rules from DB", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("rule_ids").String(fmt.Sprintf("%v", ruleIDs)),
	)

	sqlStr, args, err := sq.Delete(DATA_VIEW_ROW_COLUMN_RULE_TABLE_NAME).
		Where(sq.Eq{"f_rule_id": ruleIDs}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'delete row column rules' sql stmt failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Generate sql stmt failed")

		return err
	}

	sqlStmt := fmt.Sprintf("Sql stmt for deleting row column rules is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	if tx == nil {
		_, err = dva.db.Exec(sqlStr, args...)
	} else {
		_, err = tx.Exec(sqlStr, args...)
	}
	if err != nil {
		errDetails := fmt.Sprintf("Delete row column rules failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Delete row column rules failed")

		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 修改数据视图行列规则
func (dva *dataViewRowColumnRuleAccess) UpdateDataViewRowColumnRule(ctx context.Context, rule *interfaces.DataViewRowColumnRule) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Update a data view row column rule from DB", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("rule_id").String(rule.RuleID),
	)

	tagsStr := libCommon.TagSlice2TagString(rule.Tags)
	fieldsBytes, err := sonic.Marshal(rule.Fields)
	if err != nil {
		errDetails := fmt.Sprintf("Marshal fields failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Marshal fields failed")

		return err
	}

	rowFiltersBytes, err := sonic.Marshal(rule.RowFilters)
	if err != nil {
		errDetails := fmt.Sprintf("Marshal row filters failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Marshal row filters failed")

		return err
	}

	updateMap := map[string]any{
		"f_rule_name": rule.RuleName,
		// "f_view_id":      rule.ViewID,
		"f_tags":         tagsStr,
		"f_comment":      rule.Comment,
		"f_fields":       fieldsBytes,
		"f_row_filters":  rowFiltersBytes,
		"f_update_time":  rule.UpdateTime,
		"f_updater":      rule.Updater.ID,
		"f_updater_type": rule.Updater.Type,
	}
	sqlStr, args, err := sq.Update(DATA_VIEW_ROW_COLUMN_RULE_TABLE_NAME).
		SetMap(updateMap).
		Where(sq.Eq{"f_rule_id": rule.RuleID}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'update a row column rule' sql stmt failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Generate sql stmt failed")

		return err
	}

	sqlSmt := fmt.Sprintf("Sql stmt for updating a row column rule is '%s'", sqlStr)
	logger.Debug(sqlSmt)
	o11y.Info(ctx, sqlSmt)

	_, err = dva.db.Exec(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("Update a row column rule failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Update row column rule failed")

		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 按 id 批量获取行列规则详情
func (dva *dataViewRowColumnRuleAccess) GetDataViewRowColumnRules(ctx context.Context, ruleIDs []string) ([]*interfaces.DataViewRowColumnRule, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Get data view row column rules from DB", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("rule_ids").String(fmt.Sprintf("%v", ruleIDs)),
	)

	sqlStr, args, err := sq.Select(
		"f_rule_id",
		"f_rule_name",
		"f_view_id",
		"f_tags",
		"f_comment",
		"f_fields",
		"f_row_filters",
		"f_create_time",
		"f_update_time",
		"f_creator",
		"f_creator_type",
		"f_updater",
		"f_updater_type",
	).
		From(DATA_VIEW_ROW_COLUMN_RULE_TABLE_NAME).
		Where(sq.Eq{"f_rule_id": ruleIDs}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'get row column rules' sql stmt error: %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Generate sql stmt failed")

		return nil, err
	}

	sqlStmt := fmt.Sprintf("Sql stmt for getting row column rules is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	rows, err := dva.db.Query(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("Query data views failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Query views by IDs failed")

		return nil, err
	}
	defer rows.Close()

	rules := make([]*interfaces.DataViewRowColumnRule, 0)
	for rows.Next() {
		var tagsStr string
		var fieldsBytes, rowFiltersBytes []byte
		rule := &interfaces.DataViewRowColumnRule{}
		err = rows.Scan(
			&rule.RuleID,
			&rule.RuleName,
			&rule.ViewID,
			&tagsStr,
			&rule.Comment,
			&fieldsBytes,
			&rowFiltersBytes,
			&rule.CreateTime,
			&rule.UpdateTime,
			&rule.Creator.ID,
			&rule.Creator.Type,
			&rule.Updater.ID,
			&rule.Updater.Type,
		)
		if err != nil {
			errDetails := fmt.Sprintf("Row scan failed, error: %s", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			span.SetStatus(codes.Error, "Row scan failed")

			return nil, err
		}

		rule.Tags = libCommon.TagString2TagSlice(tagsStr)

		err = sonic.Unmarshal(fieldsBytes, &rule.Fields)
		if err != nil {
			errDetails := fmt.Sprintf("Unmarshal fields failed, %s", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			span.SetStatus(codes.Error, "Unmarshal fields failed")

			return nil, err
		}

		err = sonic.Unmarshal(rowFiltersBytes, &rule.RowFilters)
		if err != nil {
			errDetails := fmt.Sprintf("Unmarshal row filters failed, %s", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			span.SetStatus(codes.Error, "Unmarshal row filters failed")

			return nil, err
		}

		rules = append(rules, rule)
	}

	span.SetStatus(codes.Ok, "")
	return rules, nil
}

// 查询数据视图行列规则列表
func (dva *dataViewRowColumnRuleAccess) ListDataViewRowColumnRules(ctx context.Context,
	query *interfaces.ListRowColumnRuleQueryParams) ([]*interfaces.DataViewRowColumnRule, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: List data view row column rules from DB",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("name").String(query.Name),
		attr.Key("name_pattern").String(query.NamePattern),
		attr.Key("view_id").String(query.ViewID),
		attr.Key("offset").String(fmt.Sprintf("%d", query.Offset)),
		attr.Key("limit").String(fmt.Sprintf("%d", query.Limit)),
		attr.Key("sort").String(query.Sort),
		attr.Key("direction").String(query.Direction),
		attr.Key("tag").String(query.Tag),
	)

	rules := make([]*interfaces.DataViewRowColumnRule, 0)

	builder := sq.Select(
		"f_rule_id",
		"f_rule_name",
		"f_view_id",
		"f_tags",
		"f_comment",
		"f_fields",
		"f_row_filters",
		"f_create_time",
		"f_update_time",
		"f_creator",
		"f_creator_type",
		"f_updater",
		"f_updater_type",
	).
		From(DATA_VIEW_ROW_COLUMN_RULE_TABLE_NAME)

	// 过滤
	builder, err := buildRowColumnRuleListQuerySQL(query, builder)
	if err != nil {
		errDetails := fmt.Sprintf("Joint row column rule list query sql failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Joint row column rule list query sql failed")

		return nil, err
	}

	//排序
	builder = builder.OrderBy(fmt.Sprintf("%s %s", query.Sort, query.Direction))

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
		var fieldsBytes, rowFiltersBytes []byte
		rule := &interfaces.DataViewRowColumnRule{}
		err := rows.Scan(
			&rule.RuleID,
			&rule.RuleName,
			&rule.ViewID,
			&tagsStr,
			&rule.Comment,
			&fieldsBytes,
			&rowFiltersBytes,
			&rule.CreateTime,
			&rule.UpdateTime,
			&rule.Creator.ID,
			&rule.Creator.Type,
			&rule.Updater.ID,
			&rule.Updater.Type,
		)
		if err != nil {
			errDetails := fmt.Sprintf("Row scan failed, err: %s", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			span.SetStatus(codes.Error, "Row scan failed")

			return nil, err
		}

		rule.Tags = libCommon.TagString2TagSlice(tagsStr)

		err = sonic.Unmarshal(fieldsBytes, &rule.Fields)
		if err != nil {
			errDetails := fmt.Sprintf("Unmarshal fields failed, %s", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			span.SetStatus(codes.Error, "Unmarshal fields failed")

			return nil, err
		}

		err = sonic.Unmarshal(rowFiltersBytes, &rule.RowFilters)
		if err != nil {
			errDetails := fmt.Sprintf("Unmarshal row filters failed, %s", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			span.SetStatus(codes.Error, "Unmarshal row filters failed")

			return nil, err
		}

		rules = append(rules, rule)
	}

	span.SetStatus(codes.Ok, "")
	return rules, nil
}

// 查询某个数据视图下的所有行列规则,只返回简单信息
func (dva *dataViewRowColumnRuleAccess) GetSimpleRulesByViewIDs(ctx context.Context, tx *sql.Tx, viewIDs []string) ([]*interfaces.DataViewRowColumnRule, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Get data view row column rules by view IDs", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("view_ids").String(strings.Join(viewIDs, ",")),
	)

	sqlStr, args, err := sq.Select(
		"f_rule_id",
		"f_rule_name",
		"f_view_id",
	).
		From(DATA_VIEW_ROW_COLUMN_RULE_TABLE_NAME).
		Where(sq.Eq{"f_view_id": viewIDs}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'get row column rules' sql stmt error: %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Generate sql stmt failed")

		return nil, err
	}

	sqlStmt := fmt.Sprintf("Sql stmt for getting row column rules is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

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

	rules := make([]*interfaces.DataViewRowColumnRule, 0)
	for rows.Next() {
		rule := &interfaces.DataViewRowColumnRule{}
		err = rows.Scan(
			&rule.RuleID,
			&rule.RuleName,
			&rule.ViewID,
		)
		if err != nil {
			errDetails := fmt.Sprintf("Row scan failed, error: %s", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			span.SetStatus(codes.Error, "Row scan failed")

			return nil, err
		}

		rules = append(rules, rule)
	}

	span.SetStatus(codes.Ok, "")
	return rules, nil
}

// 查询某个数据视图下的所有行列规则,只返回简单信息
func (dva *dataViewRowColumnRuleAccess) GetSimpleRulesByRuleIDs(ctx context.Context, tx *sql.Tx, ruleIDs []string) ([]*interfaces.DataViewRowColumnRule, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Get data view row column rules by rule IDs", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("rule_ids").String(strings.Join(ruleIDs, ",")),
	)

	sqlStr, args, err := sq.Select(
		"f_rule_id",
		"f_rule_name",
		"f_view_id",
	).
		From(DATA_VIEW_ROW_COLUMN_RULE_TABLE_NAME).
		Where(sq.Eq{"f_rule_id": ruleIDs}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'get row column rules' sql stmt error: %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Generate sql stmt failed")

		return nil, err
	}

	sqlStmt := fmt.Sprintf("Sql stmt for getting row column rules is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

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

	rules := make([]*interfaces.DataViewRowColumnRule, 0)
	for rows.Next() {
		rule := &interfaces.DataViewRowColumnRule{}
		err = rows.Scan(
			&rule.RuleID,
			&rule.RuleName,
			&rule.ViewID,
		)
		if err != nil {
			errDetails := fmt.Sprintf("Row scan failed, error: %s", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			span.SetStatus(codes.Error, "Row scan failed")

			return nil, err
		}

		rules = append(rules, rule)
	}

	span.SetStatus(codes.Ok, "")
	return rules, nil
}

// 根据ID获取数据视图行列规则
func (dva *dataViewRowColumnRuleAccess) CheckDataViewRowColumnRuleExistByID(ctx context.Context, ruleID string) (string, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Check data view row column rule exist by ID", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("rule_id").String(ruleID),
	)

	sqlStr, args, err := sq.Select("f_rule_name").
		From(DATA_VIEW_ROW_COLUMN_RULE_TABLE_NAME).
		Where(sq.Eq{"f_rule_id": ruleID}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'check row column rule exist by ID' sql stmt failed, %s", err.Error())
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Generate sql stmt failed")

		return "", false, err
	}

	sqlStmt := fmt.Sprintf("sql stmt for checking row column rule exists by ID is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	var name string

	err = dva.db.QueryRow(sqlStr, args...).Scan(&name)
	if err == sql.ErrNoRows {
		errDetails := fmt.Sprintf("Data view row column rule %s not found", ruleID)
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Data view row column rule not found")

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
func (dva *dataViewRowColumnRuleAccess) CheckDataViewRowColumnRuleExistByName(ctx context.Context, ruleName, viewID string) (string, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Check data view row column rule exist by name", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("rule_name").String(ruleName),
		attr.Key("view_id").String(viewID),
	)

	var sqlStr string
	var args []any
	var err error
	sqlStr, args, err = sq.Select("f_view_id").
		From(DATA_VIEW_ROW_COLUMN_RULE_TABLE_NAME).
		Where(sq.Eq{
			"f_rule_name": ruleName,
			"f_view_id":   viewID,
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

	var ruleID string
	err = dva.db.QueryRow(sqlStr, args...).Scan(&ruleID)
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

// 拼接列表查询sql语句
func buildRowColumnRuleListQuerySQL(query *interfaces.ListRowColumnRuleQueryParams, builder sq.SelectBuilder) (sq.SelectBuilder, error) {
	if query.Name != "" {
		builder = builder.Where(sq.Eq{"f_rule_name": query.Name})
	} else if query.NamePattern != "" {
		builder = builder.Where(sq.Expr("instr(f_rule_name, ?) > 0", query.NamePattern))
	}

	// 根据视图 ID 过滤
	if query.ViewID != "" {
		builder = builder.Where(sq.Eq{"f_view_id": query.ViewID})
	}

	// 拼接按标签过滤
	if query.Tag != "" {
		// 格式为: %"tagname"%
		builder = builder.Where(sq.Expr("instr(f_tags, ?) > 0", `"`+query.Tag+`"`))
	}

	return builder, nil
}
