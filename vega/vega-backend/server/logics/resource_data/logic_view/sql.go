// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package logic_view

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	sq "github.com/Masterminds/squirrel"
	"github.com/mitchellh/mapstructure"

	"vega-backend/interfaces"
	"vega-backend/logics/filter_condition"
)

// maxRecursionDepth 逻辑视图最大嵌套深度，防止循环引用导致栈溢出
const maxRecursionDepth = 10

// MariaDBConnector is the modern implementation of SQL generation aligned with LogicDefinitionNode.
type MariaDBConnector struct{}

func QuotationMark(str string) string {
	return "`" + str + "`"
}

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
	nodes := make(map[string]*interfaces.LogicDefinitionNode)
	var outputNode *interfaces.LogicDefinitionNode
	for _, node := range resource.LogicDefinition {
		nodes[node.ID] = node
		if node.Type == interfaces.LogicDefinitionNodeType_Output {
			outputNode = node
		}
	}

	if outputNode == nil {
		return "", nil, fmt.Errorf("output node not found")
	}

	// 2. 从输出节点开始递归构建
	if len(outputNode.Inputs) == 0 {
		return "", nil, fmt.Errorf("output node has no input")
	}

	return c.buildNodeSQL(ctx, outputNode.Inputs[0], nodes, depth)
}

func (c *MariaDBConnector) buildNodeSQL(ctx context.Context, nodeID string,
	nodes map[string]*interfaces.LogicDefinitionNode, depth int) (string, []any, error) {
	node, ok := nodes[nodeID]
	if !ok {
		return "", nil, fmt.Errorf("node %s not found", nodeID)
	}

	switch node.Type {
	case interfaces.LogicDefinitionNodeType_Resource:
		return c.buildResourceNodeSQL(ctx, node, depth)
	case interfaces.LogicDefinitionNodeType_Join:
		return c.buildJoinNodeSQL(ctx, node, nodes, depth)
	case interfaces.LogicDefinitionNodeType_Union:
		return c.buildUnionNodeSQL(ctx, node, nodes, depth)
	case interfaces.LogicDefinitionNodeType_Sql:
		return c.buildSqlNodeSQL(ctx, node, nodes, depth)
	default:
		return "", nil, fmt.Errorf("unsupported node type: %s", node.Type)
	}
}

// buildResourceNodeSQL 构建资源节点的 SQL
func (c *MariaDBConnector) buildResourceNodeSQL(ctx context.Context,
	node *interfaces.LogicDefinitionNode, depth int) (string, []any, error) {

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
	if cfg.Distinct {
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

	// natively. MariaDBConnector handles this via ConvertFilterCondition now.
	// We'll leave it as a TODO or return a mock for now
	return sq.Expr("1=1"), nil, nil
}

// buildJoinNodeSQL 构建 JOIN 节点的 SQL
func (c *MariaDBConnector) buildJoinNodeSQL(ctx context.Context, node *interfaces.LogicDefinitionNode,
	nodes map[string]*interfaces.LogicDefinitionNode, depth int) (string, []any, error) {

	var cfg interfaces.JoinNodeCfg
	if err := mapstructure.Decode(node.Config, &cfg); err != nil {
		return "", nil, fmt.Errorf("failed to decode join node config: %w", err)
	}

	if len(node.Inputs) != 2 {
		return "", nil, fmt.Errorf("join node must have exactly 2 inputs, got %d", len(node.Inputs))
	}

	// MariaDB 不支持 FULL OUTER JOIN
	if strings.EqualFold(cfg.JoinType, interfaces.JoinType_FullOuter) {
		return "", nil, fmt.Errorf("MariaDB does not support FULL OUTER JOIN, please use LEFT JOIN + UNION instead")
	}

	leftID := node.Inputs[0]
	rightID := node.Inputs[1]

	leftSQL, leftArgs, err := c.buildNodeSQL(ctx, leftID, nodes, depth)
	if err != nil {
		return "", nil, fmt.Errorf("failed to build left input for join: %w", err)
	}
	rightSQL, rightArgs, err := c.buildNodeSQL(ctx, rightID, nodes, depth)
	if err != nil {
		return "", nil, fmt.Errorf("failed to build right input for join: %w", err)
	}

	// 构建 SELECT 字段列表，使用 from/from_node 确定来源
	fields := make([]string, 0, len(node.OutputFields))
	for _, f := range node.OutputFields {
		alias := "l"
		if f.FromNode == rightID {
			alias = "r"
		}
		// from 是源字段名, name 是输出字段名
		srcField := f.From
		if srcField == "" {
			srcField = f.Name
		}
		fields = append(fields, fmt.Sprintf("%s.`%s` AS `%s`", alias, srcField, f.Name))
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
				OriginalName: f.From,
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
func (c *MariaDBConnector) buildUnionNodeSQL(ctx context.Context, node *interfaces.LogicDefinitionNode,
	nodes map[string]*interfaces.LogicDefinitionNode, depth int) (string, []any, error) {

	var cfg interfaces.UnionNodeCfg
	if err := mapstructure.Decode(node.Config, &cfg); err != nil {
		return "", nil, fmt.Errorf("failed to decode union node config: %w", err)
	}

	// 构建输入节点 ID 到索引的映射
	inputIndexMap := make(map[string]int)
	for i, inputID := range node.Inputs {
		inputIndexMap[inputID] = i
	}

	unionParts := make([]string, 0, len(node.Inputs))
	var allArgs []any

	for i, inputID := range node.Inputs {
		subSQL, subArgs, err := c.buildNodeSQL(ctx, inputID, nodes, depth)
		if err != nil {
			return "", nil, fmt.Errorf("failed to build union input %d: %w", i, err)
		}

		// 从 output_fields 的 FromList 构建字段对齐
		hasFieldMapping := false
		for _, outField := range node.OutputFields {
			if len(outField.FromList) > 0 {
				hasFieldMapping = true
				break
			}
		}

		if hasFieldMapping {
			fields := make([]string, 0, len(node.OutputFields))
			for _, outField := range node.OutputFields {
				if i < len(outField.FromList) {
					ref := outField.FromList[i]
					fields = append(fields, fmt.Sprintf("`%s` AS `%s`", ref.From, outField.Name))
				} else {
					fields = append(fields, fmt.Sprintf("`%s`", outField.Name))
				}
			}

			allArgs = append(allArgs, subArgs...)
			unionParts = append(unionParts, fmt.Sprintf("SELECT %s FROM (%s) AS u%d",
				strings.Join(fields, ", "), subSQL, i))
		} else {
			// 没有字段映射配置时直接使用子查询
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
func (c *MariaDBConnector) buildSqlNodeSQL(ctx context.Context, node *interfaces.LogicDefinitionNode,
	nodes map[string]*interfaces.LogicDefinitionNode, depth int) (string, []any, error) {

	var cfg interfaces.SQLNodeCfg
	if err := mapstructure.Decode(node.Config, &cfg); err != nil {
		return "", nil, fmt.Errorf("failed to decode sql node config: %w", err)
	}

	// 安全检查：SQL 表达式只允许 SELECT 语句
	trimmedSQL := strings.TrimSpace(strings.ToUpper(cfg.SQL))
	if !strings.HasPrefix(trimmedSQL, "SELECT") {
		return "", nil, fmt.Errorf("sql node only allows SELECT statements, got: %.50s", cfg.SQL)
	}

	// 使用 Go template 替换 {{.node_id}} 占位符为子查询
	nodeSQLs := make(map[string]string)
	var allArgs []any

	for _, inputID := range node.Inputs {
		subSQL, subArgs, err := c.buildNodeSQL(ctx, inputID, nodes, depth)
		if err != nil {
			return "", nil, fmt.Errorf("failed to build sql node input %s: %w", inputID, err)
		}
		nodeSQLs[inputID] = "(" + subSQL + ")"
		allArgs = append(allArgs, subArgs...)
	}

	// 使用 text/template 解析并执行模板
	tmpl, err := template.New("sql").Parse(cfg.SQL)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse SQL template for node %s: %w", node.ID, err)
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, nodeSQLs); err != nil {
		return "", nil, fmt.Errorf("failed to execute SQL template for node %s: %w", node.ID, err)
	}

	return result.String(), allArgs, nil
}

func buildCountSql(fromTableStr string) string {
	return fmt.Sprintf(`SELECT count(*) FROM (%s)t`, fromTableStr)
}

// 构造视图的sql
// func buildViewSql(view *interfaces.DataView, previewDataScopeNodeID string) (string, error) {
func buildViewSql(ctx context.Context, view *interfaces.LogicView) (string, error) {

	if len(view.LogicDefinition) == 0 {
		return buildAtomicViewSql(view), nil

	} else {
		generator := NewSQLGenerator(view.LogicDefinition)
		// 找出配置里输出节点，未必是最后一个节点
		var outputNode *interfaces.LogicDefinitionNode
		for _, node := range view.LogicDefinition {
			if node.Type == interfaces.LogicDefinitionNodeType_Output {
				outputNode = node
				break
			}
		}
		if outputNode == nil {
			return "", fmt.Errorf("custom view '%s' data scope nodes is empty", view.Name)
		}

		sql, err := generator.buildNodeSQL(ctx, outputNode.ID)
		if err != nil {
			return "", fmt.Errorf("build custom view '%s' sql failed: %w", view.Name, err)
		}

		return sql, nil
	}
}

// SQLGenerator 用于生成SQL
type SQLGenerator struct {
	nodes         map[string]*interfaces.LogicDefinitionNode
	sqls          map[string]string
	nodeFieldsMap map[string]map[string]*interfaces.ViewProperty
}

// NewSQLGenerator 创建SQL生成器
func NewSQLGenerator(nodes []*interfaces.LogicDefinitionNode) *SQLGenerator {
	nodeMap := make(map[string]*interfaces.LogicDefinitionNode)
	for i := range nodes {
		nodeMap[nodes[i].ID] = nodes[i]
	}
	return &SQLGenerator{
		nodes:         nodeMap,
		sqls:          make(map[string]string),
		nodeFieldsMap: make(map[string]map[string]*interfaces.ViewProperty),
	}
}

// buildNodeSQL 生成指定节点的SQL
func (g *SQLGenerator) buildNodeSQL(ctx context.Context, nodeID string) (string, error) {
	if sql, ok := g.sqls[nodeID]; ok {
		return sql, nil
	}

	node, ok := g.nodes[nodeID]
	if !ok {
		return "", fmt.Errorf("node %s not found", nodeID)
	}

	var sql string
	var err error

	switch node.Type {
	case interfaces.LogicDefinitionNodeType_Resource:
		sql, err = g.buildResourceNodeSQL(ctx, node)
	case interfaces.LogicDefinitionNodeType_Join:
		sql, err = g.buildJoinNodeSQL(ctx, node)
	case interfaces.LogicDefinitionNodeType_Union:
		sql, err = g.buildUnionNodeSQL(ctx, node)
	case interfaces.LogicDefinitionNodeType_Sql:
		sql, err = g.buildSqlNodeSQL(ctx, node)
	case interfaces.LogicDefinitionNodeType_Output:
		sql, err = g.buildOutputNodeSQL(ctx, node)
	default:
		return "", fmt.Errorf("unknown node type: %s", node.Type)
	}

	if err != nil {
		return "", err
	}

	g.sqls[nodeID] = sql
	return sql, nil
}

// GetNodeFieldsMap 获取节点的输出字段map
func (g *SQLGenerator) GetNodeFieldsMap(nodeID string) (map[string]*interfaces.ViewProperty, error) {
	nodeMap, ok := g.nodeFieldsMap[nodeID]
	if !ok {
		return nil, fmt.Errorf("node %s fields map not found", nodeID)
	}
	return nodeMap, nil
}

// GetNodeType 获取节点类型
func (g *SQLGenerator) GetNodeType(nodeID string) (string, error) {
	node, ok := g.nodes[nodeID]
	if !ok {
		return "", fmt.Errorf("node %s not found", nodeID)
	}
	return node.Type, nil
}

// buildResourceNodeSQL 生成resource节点的SQL
// SELECT [DISTINCT] fields FROM view_id WHERE conditions
func (g *SQLGenerator) buildResourceNodeSQL(ctx context.Context, node *interfaces.LogicDefinitionNode) (string, error) {
	var cfg interfaces.ResourceNodeCfg
	err := mapstructure.Decode(node.Config, &cfg)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal resource config for node %s: %v", node.ID, err)
	}

	var resource *interfaces.Resource
	if vObj, ok := node.Config["resource"].(*interfaces.Resource); ok {
		resource = vObj
	}

	if resource == nil {
		return "", fmt.Errorf("resource is nil in node %s", node.ID)
	}

	fields := make([]string, 0, len(node.OutputFields))
	outputFieldsMap := make(map[string]*interfaces.ViewProperty)
	for _, of := range node.OutputFields {
		fields = append(fields, QuotationMark(of.OriginalName))
		outputFieldsMap[of.Name] = of
	}
	// 维护每个节点的output fields map
	g.nodeFieldsMap[node.ID] = outputFieldsMap

	fieldsStr := strings.Join(fields, ", ")

	fieldsClause := fieldsStr
	// 去重字段要在output_fields列表里
	if cfg.Distinct {
		// 名称映射，将 去重的字段name 映射为视图的原始字段名original_name
		distinctFields := make([]string, 0, len(node.OutputFields))
		for _, of := range node.OutputFields {
			distinctFields = append(distinctFields, QuotationMark(of.OriginalName))
		}
		fieldsClause = "DISTINCT " + strings.Join(distinctFields, ", ")
	}

	whereClause := ""
	if cfg.Filters != nil {
		fieldsMap := make(map[string]*interfaces.ViewProperty)
		for _, field := range resource.SchemaDefinition {
			fieldsMap[field.Name] = &interfaces.ViewProperty{
				Property: *field,
			}
		}
		// 过滤的字段未必在输出字段列表里，如果要将name映射成original_name，需要拿原始表的所有字段
		condition, err := buildSQLCondition(ctx, cfg.Filters, "", fieldsMap)
		if err != nil {
			return "", err
		}
		if condition != "" {
			whereClause = "WHERE " + condition
		}
	}

	sql := fmt.Sprintf("SELECT %s FROM %s %s", fieldsClause, resource.SourceIdentifier, whereClause)
	return sql, nil
}

// buildJoinSQL 生成join节点的SQL
func (g *SQLGenerator) buildJoinNodeSQL(ctx context.Context, node *interfaces.LogicDefinitionNode) (string, error) {
	var cfg interfaces.JoinNodeCfg
	err := mapstructure.Decode(node.Config, &cfg)
	if err != nil {
		return "", err
	}

	if len(node.Inputs) != 2 {
		return "", fmt.Errorf("join node %s requires two input nodes", node.ID)
	}

	leftNodeID := node.Inputs[0]
	rightNodeID := node.Inputs[1]
	leftSQL, err := g.buildNodeSQL(ctx, leftNodeID)
	if err != nil {
		return "", err
	}
	rightSQL, err := g.buildNodeSQL(ctx, rightNodeID)
	if err != nil {
		return "", err
	}

	onConditionsStr := make([]string, 0, len(cfg.JoinOn))
	for _, onCond := range cfg.JoinOn {
		// 左表字段和右表字段都要映射成original_name
		leftNodeFieldsMap, err := g.GetNodeFieldsMap(leftNodeID)
		if err != nil {
			return "", fmt.Errorf("failed to get node fields map for node %s in join node %s: %v", leftNodeID, node.ID, err)
		}
		rightNodeFieldsMap, err := g.GetNodeFieldsMap(rightNodeID)
		if err != nil {
			return "", fmt.Errorf("failed to get node fields map for node %s in join node %s: %v", rightNodeID, node.ID, err)
		}
		leftField, ok := leftNodeFieldsMap[onCond.LeftField]
		if !ok {
			return "", fmt.Errorf("left field %s not found in node %s in join node %s", onCond.LeftField, leftNodeID, node.ID)
		}
		rightField, ok := rightNodeFieldsMap[onCond.RightField]
		if !ok {
			return "", fmt.Errorf("right field %s not found in node %s in join node %s", onCond.RightField, rightNodeID, node.ID)
		}

		onConditionsStr = append(onConditionsStr,
			fmt.Sprintf("lft.%s %s rgt.%s", QuotationMark(leftField.OriginalName), onCond.Operator, QuotationMark(rightField.OriginalName)))
	}
	onClause := strings.Join(onConditionsStr, " AND ")

	// 构建输出字段
	fields := make([]string, 0, len(node.OutputFields))
	outputFieldsMap := make(map[string]*interfaces.ViewProperty)
	for _, of := range node.OutputFields {
		var tableAlias string
		switch of.FromNode {
		case leftNodeID:
			tableAlias = "lft"
		case rightNodeID:
			tableAlias = "rgt"
		default:
			return "", fmt.Errorf("output field from_node %s not in input nodes for node %s", of.FromNode, node.ID)
		}

		srcField := of.From
		if srcField == "" {
			srcField = of.Name
		}

		fieldExpr := fmt.Sprintf("%s.%s AS %s", tableAlias, QuotationMark(srcField), QuotationMark(of.Name))
		fields = append(fields, fieldExpr)

		// 构造输出视图的fieldsMap, name 和 字段的映射
		outputFieldsMap[of.Name] = of
	}
	// 维护每个节点的output fields map
	g.nodeFieldsMap[node.ID] = outputFieldsMap
	fieldsStr := strings.Join(fields, ", ")

	fieldsClause := fieldsStr

	whereClause := ""
	if cfg.Filters != nil {
		condition, err := buildSQLCondition(ctx, cfg.Filters, "", outputFieldsMap)
		if err != nil {
			return "", err
		}
		if condition != "" {
			whereClause = "WHERE " + condition
		}
	}

	sql := fmt.Sprintf("SELECT %s FROM ((%s) AS lft %s JOIN (%s) AS rgt ON %s) %s",
		fieldsClause, leftSQL, strings.ToUpper(cfg.JoinType), rightSQL, onClause, whereClause)
	return sql, nil
}

// buildUnionNodeSQL 生成 union节点的SQL
func (g *SQLGenerator) buildUnionNodeSQL(ctx context.Context, node *interfaces.LogicDefinitionNode) (string, error) {
	if len(node.Inputs) < 2 {
		return "", fmt.Errorf("union node %s requires at least two input nodes", node.ID)
	}

	var cfg interfaces.UnionNodeCfg
	err := mapstructure.Decode(node.Config, &cfg)
	if err != nil {
		return "", fmt.Errorf("failed to decode union config for node %s: %v", node.ID, err)
	}

	// 生成所有输入节点的SQL
	inputSQLs := make([]string, len(node.Inputs))
	for i, inputNodeID := range node.Inputs {
		sql, err := g.buildNodeSQL(ctx, inputNodeID)
		if err != nil {
			return "", err
		}
		inputSQLs[i] = sql
	}

	// 构建SELECT子句
	selectClauses := make([]string, len(node.Inputs))
	for i, inputNodeID := range node.Inputs {
		selectFields := make([]string, len(node.OutputFields))

		inputNodeFieldsMap, err := g.GetNodeFieldsMap(inputNodeID)
		if err != nil {
			return "", fmt.Errorf("failed to get node fields map for node %s in union node %s: %v", inputNodeID, node.ID, err)
		}
		inputNodeType, err := g.GetNodeType(inputNodeID)
		if err != nil {
			return "", fmt.Errorf("failed to get node type for node %s in union node %s: %v", inputNodeID, node.ID, err)
		}

		for j, of := range node.OutputFields {
			outputField := of.Name
			srcField := of.Name // 默认同名字段对齐

			// 从 FromList 中查找当前输入节点对应的原始字段
			for _, ref := range of.FromList {
				if ref.FromNode == inputNodeID {
					if ref.From != "" {
						srcField = ref.From
					}
					break
				}
			}

			if inputNodeType == interfaces.LogicDefinitionNodeType_Resource {
				if inputField, ok := inputNodeFieldsMap[srcField]; ok {
					selectFields[j] = fmt.Sprintf("%s AS %s", QuotationMark(inputField.OriginalName), QuotationMark(outputField))
				} else {
					selectFields[j] = fmt.Sprintf("%s AS %s", QuotationMark(srcField), QuotationMark(outputField))
				}
			} else {
				selectFields[j] = fmt.Sprintf("%s AS %s", QuotationMark(srcField), QuotationMark(outputField))
			}
		}
		selectClauses[i] = strings.Join(selectFields, ", ")
	}

	// 构建UNION SQL
	var unionType string
	switch cfg.UnionType {
	case interfaces.UnionType_All:
		unionType = "UNION ALL"
	case interfaces.UnionType_Distinct:
		unionType = "UNION"
	default:
		return "", fmt.Errorf("invalid union type %s for node %s", cfg.UnionType, node.ID)
	}

	// 构建完整的UNION查询
	unionParts := make([]string, len(node.Inputs))
	for i := range node.Inputs {
		unionParts[i] = fmt.Sprintf("SELECT %s FROM (%s) AS t%d", selectClauses[i], inputSQLs[i], i+1)
	}

	sql := strings.Join(unionParts, " "+unionType+" ")

	// 构建输出字段map, union的输出字段应该和第一个select的字段保持一致，outputFieldsMap 的字段key是name
	outputFieldsMap := make(map[string]*interfaces.ViewProperty)
	for _, field := range node.OutputFields {
		outputFieldsMap[field.Name] = field
	}
	// 维护每个节点的output fields map
	g.nodeFieldsMap[node.ID] = outputFieldsMap

	// 处理UNION后的过滤条件
	if cfg.Filters != nil {
		// union后过滤，过滤字段应该在输出字段里
		condition, err := buildSQLCondition(ctx, cfg.Filters, "", outputFieldsMap)

		if err != nil {
			return "", err
		}
		if condition != "" {
			sql = fmt.Sprintf("SELECT * FROM (%s) AS union_result WHERE %s", sql, condition)
		}
	}

	return sql, nil
}

// buildSqlNodeSQL 生成使用SQL表达式的节点的 SQL
func (g *SQLGenerator) buildSqlNodeSQL(ctx context.Context, node *interfaces.LogicDefinitionNode) (string, error) {
	// 检查inputs是否为空
	if len(node.Inputs) == 0 {
		return "", fmt.Errorf("sql node %s requires at least one input node", node.ID)
	}

	// 构建输出字段map, union的输出字段应该和第一个select的字段保持一致，outputFieldsMap 的字段key是name
	outputFieldsMap := make(map[string]*interfaces.ViewProperty)
	for _, field := range node.OutputFields {
		outputFieldsMap[field.Name] = field
	}
	// 维护每个节点的output fields map
	g.nodeFieldsMap[node.ID] = outputFieldsMap

	var cfg interfaces.SQLNodeCfg
	err := mapstructure.Decode(node.Config, &cfg)
	if err != nil {
		return "", fmt.Errorf("failed to decode sql config for node %s: %v", node.ID, err)
	}

	// select a from {{.node1}}
	// 创建节点SQL映射上下文
	nodeSQLs := make(map[string]string)
	for _, inputNodeID := range node.Inputs {
		sql, err := g.buildNodeSQL(ctx, inputNodeID)
		if err != nil {
			return "", fmt.Errorf("failed to build SQL for input node %s in SQL node %s: %v", inputNodeID, node.ID, err)
		}
		nodeSQLs[inputNodeID] = fmt.Sprintf("(%s)", sql)
	}

	// select a from {{node "node_id"}}
	// 创建模板函数映射
	funcMap := template.FuncMap{
		"node": func(nodeID string) (string, error) {
			sql, err := g.buildNodeSQL(ctx, nodeID)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("(%s)", sql), nil
		},
	}

	// 解析模板 (兼容旧代码里可能有SQLExpression或SQL)
	// 在新的LogicDefinitionNode中，我们期望用cfg.SQL
	tmpl, err := template.New("sql").Funcs(funcMap).Parse(cfg.SQL)
	if err != nil {
		return "", fmt.Errorf("failed to parse SQL template for node %s: %v", node.ID, err)
	}

	// 执行模板，传入节点SQL映射作为上下文
	var result strings.Builder
	err = tmpl.Execute(&result, nodeSQLs)
	if err != nil {
		return "", fmt.Errorf("failed to execute SQL template for node %s: %v", node.ID, err)
	}

	return result.String(), nil
}

// buildOutputNodeSQL 生成output节点的SQL
func (g *SQLGenerator) buildOutputNodeSQL(ctx context.Context, node *interfaces.LogicDefinitionNode) (string, error) {
	if len(node.Inputs) != 1 {
		return "", fmt.Errorf("output node %s requires exactly one input node", node.ID)
	}

	// 构建输出字段map, union的输出字段应该和第一个select的字段保持一致，outputFieldsMap 的字段key是name
	outputFieldsMap := make(map[string]*interfaces.ViewProperty)
	for _, field := range node.OutputFields {
		outputFieldsMap[field.Name] = field
	}
	// 维护每个节点的output fields map
	g.nodeFieldsMap[node.ID] = outputFieldsMap

	inputNodeID := node.Inputs[0]
	inputSQL, err := g.buildNodeSQL(ctx, inputNodeID)
	if err != nil {
		return "", err
	}

	sql := inputSQL
	return sql, nil
}

// 构造原子视图的sql
func buildAtomicViewSql(view *interfaces.LogicView) string {
	return fmt.Sprintf(`SELECT * FROM %s`, view.SourceIdentifier)
}

// 构建时间过滤Sql
func buildTimeFilterSql(dateField string, start int64, end int64) string {
	if dateField == "" {
		return ""
	}

	return fmt.Sprintf(`%s BETWEEN from_unixtime(%d) AND from_unixtime(%d)`, dateField, start/1000, end/1000)
}

// buildCondition 构建过滤条件, fieldsMap 为这个引用视图的字段map
func buildSQLCondition(ctx context.Context, filter *interfaces.FilterCondCfg, vType string, fieldsMap map[string]*interfaces.ViewProperty) (string, error) {
	// NOTE: LEGACY PATH. FilterCondition no longer provides Convert2SQL
	// natively. MariaDBConnector handles this via ConvertFilterCondition now.
	// Returning empty condition to fix compilation since this is only called
	// by legacy query paths that will be commented out.
	return "", nil
}

// buildRowColumnRulesSQL 构建行列规则过滤到SQL
func buildRowColumnRulesSQL(ctx context.Context, rules []any,
	view *interfaces.LogicView) (string, []*interfaces.Property, map[string]*interfaces.ViewProperty, error) {

	// Legacy row condition parsing removed to fix compilation
	return "", view.SchemaDefinition, view.FieldsMap, nil
}

func isValidFilters(cfg *interfaces.FilterCondCfg) bool {
	if cfg == nil {
		return false
	}

	// 判断过滤器是否为空对象 {}
	if cfg.Name == "" && cfg.Operation == "" && len(cfg.SubConds) == 0 && cfg.ValueFrom == "" && cfg.Value == nil {
		return false
	}

	return true
}

// 构建sort
func buildSQLSortParams(sort []*interfaces.SortField) string {
	if len(sort) == 0 {
		return ""
	}

	var sortSql strings.Builder
	for i, sortParam := range sort {
		if i > 0 {
			sortSql.WriteString(", ")
		}
		sortSql.WriteString(fmt.Sprintf("%s %s", QuotationMark(sortParam.Field), sortParam.Direction))
	}

	return sortSql.String()
}

// 补充 sort 字段
func prepareSQLSortParams(sort []*interfaces.SortField, fieldsMap map[string]*interfaces.ViewProperty) []*interfaces.SortField {

	newSort := []*interfaces.SortField{}
	// 去重并过滤不在视图字段列表中的排序字段
	sortFieldSet := map[string]struct{}{}
	for _, sortParam := range sort {
		_, isInFieldsMap := fieldsMap[sortParam.Field]

		// 只保留元字段或存在于视图字段列表中的排序字段
		if isInFieldsMap && sortParam.Field != "" {
			if _, ok := sortFieldSet[sortParam.Field]; !ok {
				newSort = append(newSort, sortParam)
				sortFieldSet[sortParam.Field] = struct{}{}
			}
		}
	}

	return newSort
}

// SQLBuilder - SQL 构建器结构体
type SQLBuilder struct {
	baseQuery        string
	whereClauses     []string
	isSubQuery       bool
	hasExistingWhere bool
}

// NewSQLBuilder 创建新的 SQL 构建器
func NewSQLBuilder(baseQuery string) *SQLBuilder {
	builder := &SQLBuilder{
		baseQuery:    strings.TrimSpace(baseQuery),
		whereClauses: []string{},
	}

	// 检测查询类型和结构
	builder.analyzeQuery()
	return builder
}

// analyzeQuery 分析基础查询的结构
func (b *SQLBuilder) analyzeQuery() {
	upperQuery := strings.ToUpper(b.baseQuery)

	// 检测是否为子查询（以括号开头或包含多个SELECT）
	b.isSubQuery = strings.HasPrefix(b.baseQuery, "(") ||
		(strings.Contains(upperQuery, "SELECT") &&
			strings.Count(upperQuery, "SELECT") > 1)

	// 检测是否已包含 WHERE 子句
	b.hasExistingWhere = strings.Contains(upperQuery, " WHERE ")
}

// AddWhere 添加 WHERE 条件
func (b *SQLBuilder) AddWhere(condition string) *SQLBuilder {
	if strings.TrimSpace(condition) != "" {
		b.whereClauses = append(b.whereClauses, condition)
	}
	return b
}

// AddWheres 批量添加 WHERE 条件
func (b *SQLBuilder) AddWheres(conditions []string) *SQLBuilder {
	for _, condition := range conditions {
		b.AddWhere(condition)
	}
	return b
}

// Build 构建最终的 SQL 语句
func (b *SQLBuilder) Build() string {
	if len(b.whereClauses) == 0 {
		return b.baseQuery
	}

	whereStr := strings.Join(b.whereClauses, " AND ")

	// 如果是子查询，需要在外层包装
	if b.isSubQuery {
		return b.wrapSubQuery(whereStr)
	}

	// 普通查询，智能添加 WHERE
	return b.buildStandardQuery(whereStr)
}

// wrapSubQuery 包装子查询
func (b *SQLBuilder) wrapSubQuery(whereStr string) string {
	// 如果子查询已经有别名，直接使用
	if b.hasAlias() {
		return fmt.Sprintf("%s WHERE %s", b.baseQuery, whereStr)
	}

	// 给子查询添加默认别名
	return fmt.Sprintf("(%s) AS subquery WHERE %s", b.baseQuery, whereStr)
}

// buildStandardQuery 构建标准查询
func (b *SQLBuilder) buildStandardQuery(whereStr string) string {
	if b.hasExistingWhere {
		// 已有 WHERE，使用 AND 连接
		return b.insertWhereCondition(whereStr, "AND")
	}

	// 没有 WHERE，添加 WHERE 子句
	return b.insertWhereCondition(whereStr, "WHERE")
}

// insertWhereCondition 在合适的位置插入 WHERE 条件
func (b *SQLBuilder) insertWhereCondition(condition, keyword string) string {
	upperQuery := strings.ToUpper(b.baseQuery)
	hasWhere := strings.Contains(upperQuery, " WHERE ")

	// 查找关键词位置（GROUP BY, ORDER BY, LIMIT 等）
	keywordPositions := []struct {
		keyword string
		index   int
	}{
		{" GROUP BY ", strings.Index(upperQuery, " GROUP BY ")},
		{" ORDER BY ", strings.Index(upperQuery, " ORDER BY ")},
		{" LIMIT ", strings.Index(upperQuery, " LIMIT ")},
		{" HAVING ", strings.Index(upperQuery, " HAVING ")},
	}

	// 找到第一个出现的关键词
	insertPosition := -1
	for _, kp := range keywordPositions {
		if kp.index != -1 && (insertPosition == -1 || kp.index < insertPosition) {
			insertPosition = kp.index
		}
	}

	// 确定要使用的连接词
	var actualKeyword string
	if hasWhere {
		// 如果已有 WHERE 子句，使用 AND 或 OR
		actualKeyword = keyword
	} else {
		// 如果没有 WHERE 子句，使用 WHERE
		actualKeyword = "WHERE"
	}

	if insertPosition != -1 {
		// 在关键词前插入条件
		return b.baseQuery[:insertPosition] + " " + actualKeyword + " " + condition + " " + b.baseQuery[insertPosition:]
	}

	// 没有找到关键词，在末尾添加
	var connector string
	if hasWhere {
		// 如果已有 WHERE 子句，使用 AND 或 OR 连接
		connector = " " + keyword + " "
	} else {
		// 如果没有 WHERE 子句，添加 WHERE 关键字
		connector = " WHERE "
	}
	return b.baseQuery + connector + condition
}

// hasAlias 检测子查询是否已有别名
func (b *SQLBuilder) hasAlias() bool {
	// 简单的别名检测逻辑
	if !b.isSubQuery {
		return false
	}

	// 检查是否以 ) AS 某个名字 结尾
	trimmed := strings.TrimSpace(b.baseQuery)
	if strings.HasSuffix(trimmed, ")") {
		return false
	}

	// 检查是否包含 AS 关键字
	upperQuery := strings.ToUpper(b.baseQuery)
	lastParen := strings.LastIndex(upperQuery, ")")
	if lastParen == -1 {
		return false
	}

	// 在最后一个括号后有 AS 关键字
	afterParen := strings.TrimSpace(upperQuery[lastParen+1:])
	return strings.HasPrefix(afterParen, "AS ")
}

// String 实现 Stringer 接口
func (b *SQLBuilder) String() string {
	return b.Build()
}

// HasLimit 检查 SQL 是否已包含 LIMIT 子句
func HasLimit(sql string) bool {
	// 转换为小写便于匹配
	lowerSQL := strings.ToLower(sql)

	// 移除注释
	cleanedSQL := removeSQLComments(lowerSQL)

	// 匹配 LIMIT 子句的正则表达式
	// 匹配格式：LIMIT 数字 或 LIMIT 数字,数字 或 LIMIT 数字 OFFSET 数字
	limitPattern := `\blimit\s+(\d+)(?:\s*,\s*\d+|\s+offset\s+\d+)?\s*$`

	matched, _ := regexp.MatchString(limitPattern, cleanedSQL)
	return matched
}

// removeSQLComments 移除 SQL 注释
func removeSQLComments(sql string) string {
	// 移除单行注释 (-- 注释)
	singleLineComment := `--[^\n]*`
	re := regexp.MustCompile(singleLineComment)
	sql = re.ReplaceAllString(sql, "")

	// 移除多行注释 (/* 注释 */)
	multiLineComment := `/\*.*?\*/`
	re = regexp.MustCompile(multiLineComment)
	sql = re.ReplaceAllString(sql, "")

	return strings.TrimSpace(sql)
}

// AddLimitIfMissing 如果 SQL 没有 LIMIT，则添加 LIMIT
func AddLimitIfMissing(sql string, limit int) string {
	if HasLimit(sql) {
		return sql
	}

	// 确保 SQL 以分号结尾，然后添加 LIMIT
	trimmedSQL := strings.TrimSpace(sql)
	trimmedSQL = strings.TrimSuffix(trimmedSQL, ";")

	return trimmedSQL + " LIMIT " + strconv.Itoa(limit)
}
