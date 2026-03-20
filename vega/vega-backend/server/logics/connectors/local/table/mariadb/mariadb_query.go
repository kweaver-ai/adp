// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package mariadb provides MariaDB database connector implementation.
package mariadb

import (
	"context"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/go-sql-driver/mysql"
	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"vega-backend/interfaces"
)

// convertValue converts []byte to string for MariaDB driver compatibility
func convertValue(v any) any {
	if b, ok := v.([]byte); ok {
		return string(b)
	}
	return v
}

func (c *MariaDBConnector) ExecuteQuery(ctx context.Context, resource *interfaces.Resource,
	params *interfaces.ResourceDataQueryParams) (*interfaces.QueryResult, error) {

	if err := c.Connect(ctx); err != nil {
		return nil, err
	}

	fieldMap := map[string]*interfaces.Property{}
	for _, prop := range resource.SchemaDefinition {
		fieldMap[prop.Name] = prop
	}

	var condition sq.Sqlizer
	var err error
	if params.ActualFilterCond != nil {
		condition, err = c.ConvertFilterCondition(ctx, params.ActualFilterCond, fieldMap)
		if err != nil {
			return nil, err
		}
	}

	result := &interfaces.QueryResult{
		Rows: make([]map[string]any, 0),
	}

	var lvSql string
	var lvArgs []any
	if resource.Category == interfaces.ResourceCategoryLogicView {
		var args []any
		lvSql, args, err = c.BuildLogicViewSQL(ctx, resource)
		if err != nil {
			return nil, fmt.Errorf("failed to build logic view sql: %w", err)
		}
		lvArgs = args
	}
	logger.Infof("1 lvSql: %s", lvSql)
	logger.Infof("2 lvArgs: %v", lvArgs)

	if params.NeedTotal {
		var countBuilder sq.SelectBuilder
		if resource.Category == interfaces.ResourceCategoryLogicView {
			countBuilder = sq.Select("COUNT(1)").From(fmt.Sprintf("(%s) AS t", lvSql))
		} else {
			countBuilder = sq.Select("COUNT(1)").From(resource.SourceIdentifier)
		}

		if condition != nil {
			countBuilder = countBuilder.Where(condition)
		}

		query, args, err := countBuilder.ToSql()
		if err != nil {
			return nil, fmt.Errorf("failed to build query: %w", err)
		}

		// 合并参数（显式拷贝，避免污染 lvArgs）
		finalCountArgs := make([]any, 0, len(lvArgs)+len(args))
		finalCountArgs = append(finalCountArgs, lvArgs...)
		finalCountArgs = append(finalCountArgs, args...)

		logger.Debugf("count query: %s, args: %v", query, finalCountArgs)

		var total int64
		row := c.db.QueryRowContext(ctx, query, finalCountArgs...)
		if err := row.Scan(&total); err != nil {
			return nil, fmt.Errorf("failed to scan total: %w", err)
		}

		result.Total = total
	}

	fields := []string{"*"}
	if len(params.OutputFields) > 0 {
		fields = params.OutputFields
	}

	logger.Infof("lvSql: %s", lvSql)
	logger.Infof("lvArgs: %v", lvArgs)

	var builder sq.SelectBuilder
	if resource.Category == interfaces.ResourceCategoryLogicView {
		builder = sq.Select(fields...).From(fmt.Sprintf("(%s) AS t", lvSql))
	} else {
		builder = sq.Select(fields...).From(resource.SourceIdentifier)
	}

	// 1. 处理过滤条件
	if condition != nil {
		builder = builder.Where(condition)
	}

	// 2. 处理 search_after 过滤 (Keyset Pagination)
	if params.UseSearchAfter && len(params.SearchAfter) > 0 && len(params.Sort) > 0 {
		cursorCond, cursorArgs, err := c.buildSearchAfterCondition(params.Sort, params.SearchAfter)
		if err != nil {
			return nil, fmt.Errorf("failed to build search_after condition: %w", err)
		}
		if cursorCond != "" {
			builder = builder.Where(cursorCond, cursorArgs...)
		}
	}

	// 3. 处理排序
	if len(params.Sort) > 0 {
		for _, sortField := range params.Sort {
			direction := "ASC"
			if sortField.Direction == "desc" {
				direction = "DESC"
			}
			builder = builder.OrderBy(fmt.Sprintf("`%s` %s", sortField.Field, direction))
		}
	}

	// 4. 处理分页 (search_after 模式忽略 offset)
	if params.Limit > 0 {
		builder = builder.Limit(uint64(params.Limit))
	}
	if !params.UseSearchAfter && params.Offset > 0 {
		builder = builder.Offset(uint64(params.Offset))
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	// 合并参数：逻辑视图的参数在前（显式拷贝）
	finalArgs := make([]any, 0, len(lvArgs)+len(args))
	finalArgs = append(finalArgs, lvArgs...)
	finalArgs = append(finalArgs, args...)

	logger.Debugf("query: %s, args: %v", query, finalArgs)

	rows, err := c.db.QueryContext(ctx, query, finalArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	result.Columns = columns

	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := make(map[string]any)
		for i, col := range columns {
			row[col] = convertValue(values[i])
		}
		result.Rows = append(result.Rows, row)
	}

	// 5. 提取 search_after (下一页的游标)
	if params.UseSearchAfter && len(result.Rows) > 0 && len(params.Sort) > 0 {
		lastRow := result.Rows[len(result.Rows)-1]
		searchAfter := make([]any, 0, len(params.Sort))
		for _, sortField := range params.Sort {
			val, ok := lastRow[sortField.Field]
			if !ok {
				return nil, fmt.Errorf("sort field %s not found in result for search_after", sortField.Field)
			}
			searchAfter = append(searchAfter, val)
		}
		result.SearchAfter = searchAfter
	}

	return result, nil
}

// buildSearchAfterCondition 构建 Keysets Pagination 的 WHERE 子句
// 逻辑：(f1, f2, f3) > (v1, v2, v3) 展开为:
// (f1 > v1) OR (f1 = v1 AND f2 > v2) OR (f1 = v1 AND f2 = v2 AND f3 > v3)
// 如果是 DESC 则 > 改为 <
func (c *MariaDBConnector) buildSearchAfterCondition(sorts []*interfaces.SortField,
	after []any) (string, []any, error) {

	if len(sorts) == 0 || len(after) == 0 {
		return "", nil, nil
	}

	n := len(sorts)
	if len(after) < n {
		n = len(after)
	}

	var orParts []string
	var allArgs []any

	for i := 0; i < n; i++ {
		var andParts []string
		var andArgs []any

		// 前面相等的列
		for j := 0; j < i; j++ {
			andParts = append(andParts, fmt.Sprintf("`%s` = ?", sorts[j].Field))
			andArgs = append(andArgs, after[j])
		}

		// 当前不相等的列
		op := ">"
		if sorts[i].Direction == "desc" {
			op = "<"
		}
		andParts = append(andParts, fmt.Sprintf("`%s` %s ?", sorts[i].Field, op))
		andArgs = append(andArgs, after[i])

		orParts = append(orParts, "("+strings.Join(andParts, " AND ")+")")
		allArgs = append(allArgs, andArgs...)
	}

	return "(" + strings.Join(orParts, " OR ") + ")", allArgs, nil
}
