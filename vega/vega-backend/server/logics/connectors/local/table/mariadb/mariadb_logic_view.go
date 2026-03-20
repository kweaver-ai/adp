// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package mariadb

import (
	"context"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/mitchellh/mapstructure"

	"vega-backend/interfaces"
	"vega-backend/logics/filter_condition"
)

// maxRecursionDepth 逻辑视图最大嵌套深度，防止循环引用导致栈溢出
const maxRecursionDepth = 10

// BuildLogicViewSQL 构建逻辑视图的 SQL
func (c *MariaDBConnector) BuildLogicViewSQL(ctx context.Context, resource *interfaces.Resource) (string, []any, error) {
	return c.buildLogicViewSQLWithDepth(ctx, resource, maxRecursionDepth)
}

func (c *MariaDBConnector) buildLogicViewSQLWithDepth(ctx context.Context, resource *interfaces.Resource, depth int) (string, []any, error) {
	if depth <= 0 {
		return "", nil, fmt.Errorf("max recursion depth (%d) exceeded, possible circular reference in logic view", maxRecursionDepth)
	}

	if resource.LogicDefinition == nil {
		return "", nil, fmt.Errorf("logic definition is empty")
	}

	// 1. 将节点索引化
	nodes := make(map[string]*interfaces.DataScopeNode)
	var outputNode *interfaces.DataScopeNode
	for _, node := range resource.LogicDefinition {
		nodes[node.ID] = node
		if node.Type == interfaces.DataScopeNodeType_Output {
			outputNode = node
		}
	}

	if outputNode == nil {
		return "", nil, fmt.Errorf("output node not found")
	}

	// 2. 从输出节点开始递归构建
	if len(outputNode.InputNodes) == 0 {
		return "", nil, fmt.Errorf("output node has no input")
	}

	return c.buildNodeSQL(ctx, outputNode.InputNodes[0], nodes, depth)
}

func (c *MariaDBConnector) buildNodeSQL(ctx context.Context, nodeID string,
	nodes map[string]*interfaces.DataScopeNode, depth int) (string, []any, error) {
	node, ok := nodes[nodeID]
	if !ok {
		return "", nil, fmt.Errorf("node %s not found", nodeID)
	}

	switch node.Type {
	case interfaces.DataScopeNodeType_Resource:
		return c.buildResourceNodeSQL(ctx, node, depth)
	case interfaces.DataScopeNodeType_Join:
		return c.buildJoinNodeSQL(ctx, node, nodes, depth)
	case interfaces.DataScopeNodeType_Union:
		return c.buildUnionNodeSQL(ctx, node, nodes, depth)
	case interfaces.DataScopeNodeType_Sql:
		return c.buildSqlNodeSQL(ctx, node, nodes, depth)
	default:
		return "", nil, fmt.Errorf("unsupported node type: %s", node.Type)
	}
}

// buildResourceNodeSQL 构建资源节点的 SQL
func (c *MariaDBConnector) buildResourceNodeSQL(ctx context.Context,
	node *interfaces.DataScopeNode, depth int) (string, []any, error) {

	var cfg interfaces.ResourceNodeCfg
	if err := mapstructure.Decode(node.Config, &cfg); err != nil {
		return "", nil, fmt.Errorf("failed to decode resource node config: %w", err)
	}

	res := cfg.Resource
	if res == nil {
		return "", nil, fmt.Errorf("resource not found in node %s", node.ID)
	}

	// 如果资源本身也是逻辑视图，递归构建（消耗一层深度）
	if res.Category == interfaces.ResourceCategoryLogicView {
		return c.buildLogicViewSQLWithDepth(ctx, res, depth-1)
	}

	// 构建字段映射
	fieldMap := make(map[string]*interfaces.Property)
	for _, prop := range res.SchemaDefinition {
		fieldMap[prop.Name] = prop
	}

	// 构建 SELECT 字段列表
	var fields []string
	if len(node.OutputFields) > 0 {
		fields = make([]string, 0, len(node.OutputFields))
		for _, f := range node.OutputFields {
			sourceProp, ok := fieldMap[f.Name]
			if !ok {
				fields = append(fields, fmt.Sprintf("`%s`", f.Name))
			} else {
				if sourceProp.OriginalName != "" && sourceProp.OriginalName != f.Name {
					fields = append(fields, fmt.Sprintf("`%s` AS `%s`", sourceProp.OriginalName, f.Name))
				} else {
					fields = append(fields, fmt.Sprintf("`%s`", f.Name))
				}
			}
		}
	} else {
		fields = []string{"*"}
	}

	// 构建表源
	builder := sq.Select(fields...).From(res.SourceIdentifier)

	// 处理去重
	if cfg.Distinct.Enable {
		builder = builder.Distinct()
	}

	// 处理过滤条件
	filterCond, filterArgs, err := c.buildFilterSQL(ctx, cfg.Filters, fieldMap)
	if err != nil {
		return "", nil, fmt.Errorf("failed to build resource node filter: %w", err)
	}
	if filterCond != nil {
		builder = builder.Where(filterCond)
	}

	sqlStr, args, err := builder.ToSql()
	if err != nil {
		return "", nil, err
	}
	args = append(args, filterArgs...)
	return sqlStr, args, nil
}

// buildFilterSQL 将 FilterCondCfg 转换为 squirrel 条件
func (c *MariaDBConnector) buildFilterSQL(ctx context.Context, filters *interfaces.FilterCondCfg,
	fieldMap map[string]*interfaces.Property) (sq.Sqlizer, []any, error) {

	if filters == nil {
		return nil, nil, nil
	}

	filterCond, err := filter_condition.NewFilterCondition(ctx, filters, fieldMap)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create filter condition: %w", err)
	}
	if filterCond == nil {
		return nil, nil, nil
	}

	sqlCond, err := c.ConvertFilterCondition(ctx, filterCond, fieldMap)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert filter condition to SQL: %w", err)
	}

	return sqlCond, nil, nil
}

// buildJoinNodeSQL 构建 JOIN 节点的 SQL
func (c *MariaDBConnector) buildJoinNodeSQL(ctx context.Context, node *interfaces.DataScopeNode,
	nodes map[string]*interfaces.DataScopeNode, depth int) (string, []any, error) {

	var cfg interfaces.JoinNodeCfg
	if err := mapstructure.Decode(node.Config, &cfg); err != nil {
		return "", nil, fmt.Errorf("failed to decode join node config: %w", err)
	}

	if len(node.InputNodes) != 2 {
		return "", nil, fmt.Errorf("join node must have exactly 2 inputs, got %d", len(node.InputNodes))
	}

	// MariaDB 不支持 FULL OUTER JOIN
	if strings.EqualFold(cfg.JoinType, interfaces.JoinType_FullOuter) {
		return "", nil, fmt.Errorf("MariaDB does not support FULL OUTER JOIN, please use LEFT JOIN + UNION instead")
	}

	leftID := node.InputNodes[0]
	rightID := node.InputNodes[1]

	leftSQL, leftArgs, err := c.buildNodeSQL(ctx, leftID, nodes, depth)
	if err != nil {
		return "", nil, fmt.Errorf("failed to build left input for join: %w", err)
	}
	rightSQL, rightArgs, err := c.buildNodeSQL(ctx, rightID, nodes, depth)
	if err != nil {
		return "", nil, fmt.Errorf("failed to build right input for join: %w", err)
	}

	// 构建 SELECT 字段列表
	fields := make([]string, 0, len(node.OutputFields))
	for _, f := range node.OutputFields {
		// SrcNodeID 指向来源的输入节点
		alias := "l"
		if f.SrcNodeID == rightID {
			alias = "r"
		}
		fields = append(fields, fmt.Sprintf("%s.`%s` AS `%s`", alias, f.OriginalName, f.Name))
	}

	// 构建 JOIN ON 条件
	joinOnParts := make([]string, 0, len(cfg.JoinOn))
	for _, on := range cfg.JoinOn {
		joinOnParts = append(joinOnParts, fmt.Sprintf("l.`%s` = r.`%s`", on.LeftField, on.RightField))
	}
	joinOn := strings.Join(joinOnParts, " AND ")

	joinType := strings.ToUpper(cfg.JoinType)
	if joinType == "" {
		joinType = "INNER"
	}

	// 合并参数：注意不能直接 append 到 leftArgs 上，避免污染
	allArgs := make([]any, 0, len(leftArgs)+len(rightArgs))
	allArgs = append(allArgs, leftArgs...)
	allArgs = append(allArgs, rightArgs...)

	sqlStr := fmt.Sprintf("SELECT %s FROM (%s) AS l %s JOIN (%s) AS r ON %s",
		strings.Join(fields, ", "), leftSQL, joinType, rightSQL, joinOn)

	// 处理 Join 节点自身的过滤条件
	if cfg.Filters != nil {
		// Join 后的字段需要构建一个临时的 fieldMap
		joinFieldMap := make(map[string]*interfaces.Property)
		for _, f := range node.OutputFields {
			joinFieldMap[f.Name] = &interfaces.Property{
				Name:         f.Name,
				Type:         f.Type,
				OriginalName: f.OriginalName,
			}
		}

		filterCond, _, err := c.buildFilterSQL(ctx, cfg.Filters, joinFieldMap)
		if err != nil {
			return "", nil, fmt.Errorf("failed to build join node filter: %w", err)
		}
		if filterCond != nil {
			whereSql, whereArgs, err := filterCond.ToSql()
			if err != nil {
				return "", nil, fmt.Errorf("failed to convert join filter to SQL: %w", err)
			}
			sqlStr = fmt.Sprintf("SELECT * FROM (%s) AS j WHERE %s", sqlStr, whereSql)
			allArgs = append(allArgs, whereArgs...)
		}
	}

	return sqlStr, allArgs, nil
}

// buildUnionNodeSQL 构建 UNION 节点的 SQL
func (c *MariaDBConnector) buildUnionNodeSQL(ctx context.Context, node *interfaces.DataScopeNode,
	nodes map[string]*interfaces.DataScopeNode, depth int) (string, []any, error) {

	var cfg interfaces.UnionNodeCfg
	if err := mapstructure.Decode(node.Config, &cfg); err != nil {
		return "", nil, fmt.Errorf("failed to decode union node config: %w", err)
	}

	unionParts := make([]string, 0, len(node.InputNodes))
	var allArgs []any

	for i, inputID := range node.InputNodes {
		subSQL, subArgs, err := c.buildNodeSQL(ctx, inputID, nodes, depth)
		if err != nil {
			return "", nil, fmt.Errorf("failed to build union input %d: %w", i, err)
		}

		// Union 节点需要对齐字段
		if len(cfg.UnionFields) > i && len(cfg.UnionFields[i]) > 0 {
			// 按照 UnionFields 配置选择和重命名字段
			var selectArgs []any
			fields := make([]string, 0, len(node.OutputFields))
			for j, outField := range node.OutputFields {
				if j >= len(cfg.UnionFields[i]) {
					break
				}
				uField := cfg.UnionFields[i][j]
				if uField.ValueFrom == "field" {
					fields = append(fields, fmt.Sprintf("`%s` AS `%s`", uField.Field, outField.Name))
				} else {
					// 常量值
					fields = append(fields, fmt.Sprintf("? AS `%s`", outField.Name))
					selectArgs = append(selectArgs, uField.Field)
				}
			}

			// 参数顺序：SELECT 子句中的 ? 在 FROM 子查询的 ? 之前
			allArgs = append(allArgs, selectArgs...)
			allArgs = append(allArgs, subArgs...)
			unionParts = append(unionParts, fmt.Sprintf("SELECT %s FROM (%s) AS u%d",
				strings.Join(fields, ", "), subSQL, i))
		} else {
			// 没有 UnionFields 配置时直接使用子查询
			allArgs = append(allArgs, subArgs...)
			unionParts = append(unionParts, fmt.Sprintf("(%s)", subSQL))
		}
	}

	unionOp := "UNION ALL"
	if cfg.UnionType == interfaces.UnionType_Distinct {
		unionOp = "UNION"
	}

	sql := strings.Join(unionParts, " "+unionOp+" ")
	return "SELECT * FROM (" + sql + ") AS union_result", allArgs, nil
}

// buildSqlNodeSQL 构建自定义 SQL 节点
func (c *MariaDBConnector) buildSqlNodeSQL(ctx context.Context, node *interfaces.DataScopeNode,
	nodes map[string]*interfaces.DataScopeNode, depth int) (string, []any, error) {

	var cfg interfaces.SQLNodeCfg
	if err := mapstructure.Decode(node.Config, &cfg); err != nil {
		return "", nil, fmt.Errorf("failed to decode sql node config: %w", err)
	}

	// 安全检查：SQL 表达式只允许 SELECT 语句
	trimmedSQL := strings.TrimSpace(strings.ToUpper(cfg.SQLExpression))
	if !strings.HasPrefix(trimmedSQL, "SELECT") {
		return "", nil, fmt.Errorf("sql node only allows SELECT statements, got: %.50s", cfg.SQLExpression)
	}

	// 将 SQL 中的 ${node_id} 占位符替换为子查询
	finalSQL := cfg.SQLExpression
	var allArgs []any

	for _, inputID := range node.InputNodes {
		subSQL, subArgs, err := c.buildNodeSQL(ctx, inputID, nodes, depth)
		if err != nil {
			return "", nil, fmt.Errorf("failed to build sql node input %s: %w", inputID, err)
		}
		placeholder := fmt.Sprintf("${%s}", inputID)
		finalSQL = strings.ReplaceAll(finalSQL, placeholder, "("+subSQL+")")
		allArgs = append(allArgs, subArgs...)
	}

	return finalSQL, allArgs, nil
}
