package data_view

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/mitchellh/mapstructure"

	"uniquery/common"
	cond "uniquery/common/condition"
	"uniquery/interfaces"
)

func buildCountSql(fromTableStr string) string {
	return fmt.Sprintf(`SELECT count(*) FROM (%s)t`, fromTableStr)
}

// 构造视图的sql
// func buildViewSql(view *interfaces.DataView, previewDataScopeNodeID string) (string, error) {
func buildViewSql(ctx context.Context, view *interfaces.DataView) (string, error) {

	if view.Type == interfaces.ViewType_Atomic {
		return buildAtomicViewSql(view), nil

	} else {
		if len(view.DataScope) == 0 {
			return "", fmt.Errorf("custom view %s data scope nodes is empty", view.ViewID)
		}

		generator := NewSQLGenerator(view.DataScope)
		// 找出配置里输出节点，未必是最后一个节点
		var outputNode *interfaces.DataScopeNode
		for _, node := range view.DataScope {
			if node.Type == interfaces.DataScopeNodeType_Output {
				outputNode = node
				break
			}

			// // 目标节点存在时，以目标节点为输出节点
			// if node.ID == previewDataScopeNodeID {
			// 	outputNode = node
			// 	break
			// }
		}
		if outputNode == nil {
			return "", fmt.Errorf("custom view '%s' data scope nodes is empty", view.ViewName)
		}

		sql, err := generator.buildNodeSQL(ctx, outputNode.ID)
		if err != nil {
			return "", fmt.Errorf("build custom view '%s' sql failed: %w", view.ViewName, err)
		}

		return sql, nil
	}
}

// SQLGenerator 用于生成SQL
type SQLGenerator struct {
	nodes map[string]*interfaces.DataScopeNode
	sqls  map[string]string
}

// NewSQLGenerator 创建SQL生成器
func NewSQLGenerator(nodes []*interfaces.DataScopeNode) *SQLGenerator {
	nodeMap := make(map[string]*interfaces.DataScopeNode)
	for i := range nodes {
		nodeMap[nodes[i].ID] = nodes[i]
	}
	return &SQLGenerator{
		nodes: nodeMap,
		sqls:  make(map[string]string),
	}
}

// buildSQL 生成指定节点的SQL
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
	case interfaces.DataScopeNodeType_View:
		sql, err = g.buildViewNodeSQL(ctx, node)
	case interfaces.DataScopeNodeType_Join:
		sql, err = g.buildJoinNodeSQL(ctx, node)
	case interfaces.DataScopeNodeType_Union:
		sql, err = g.buildUnionNodeSQL(ctx, node)
	case interfaces.DataScopeNodeType_Sql:
		sql, err = g.buildSqlNodeSQL(ctx, node)
	case interfaces.DataScopeNodeType_Output:
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
func (g *SQLGenerator) GetNodeFieldsMap(nodeID string) (map[string]*cond.ViewField, error) {
	node, ok := g.nodes[nodeID]
	if !ok {
		return nil, fmt.Errorf("node %s not found", nodeID)
	}
	return node.OutputFieldsMap, nil
}

// buildViewSQL 生成view节点的SQL
// SELECT [DISTINCT] fields FROM view_id WHERE conditions
func (g *SQLGenerator) buildViewNodeSQL(ctx context.Context, node *interfaces.DataScopeNode) (string, error) {
	var cfg interfaces.ViewNodeCfg
	err := mapstructure.Decode(node.Config, &cfg)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal view config for node %s: %v", node.ID, err)
	}

	if cfg.View == nil {
		return "", fmt.Errorf("view is nil")
	}

	fields := make([]string, 0, len(node.OutputFields))
	outputFieldsMap := make(map[string]*cond.ViewField)
	for _, of := range node.OutputFields {
		// if of.Alias != "" {
		// 	fields[i] = fmt.Sprintf("%s AS %s", of.OriginalName, of.Alias)
		// } else {
		fields = append(fields, common.QuotationMark(of.OriginalName))
		// 字段加双引号，兼容国产数据库
		// fields = append(fields, fmt.Sprintf("\"%s\"", of.OriginalName))
		// }

		// 构造输出视图的fieldsMap, 一个表内 name 不可能重复
		outputFieldsMap[of.Name] = of
	}
	// 维护每个节点的output fields map
	node.OutputFieldsMap = outputFieldsMap

	fieldsStr := strings.Join(fields, ", ")

	fieldsClause := fieldsStr
	// 去重字段要在output_fields列表里
	if cfg.Distinct.Enable {
		if len(cfg.Distinct.Fields) > 0 {
			// 名称映射，将 去重的字段name 映射为视图的原始字段名original_name
			distinctFields := make([]string, 0, len(cfg.Distinct.Fields))
			for _, df := range cfg.Distinct.Fields {
				if of, ok := outputFieldsMap[df]; ok {
					distinctFields = append(distinctFields, common.QuotationMark(of.OriginalName))
				}
			}
			fieldsClause = "DISTINCT " + strings.Join(distinctFields, ", ")

		} else {
			fieldsClause = "DISTINCT " + fieldsStr
		}
	}

	whereClause := ""
	if cfg.Filters != nil {
		// 过滤的字段未必在输出字段列表里，如果要将name映射成original_name，需要拿原始表的所有字段
		condition, err := buildSQLCondition(ctx, cfg.Filters, interfaces.ViewType_Custom, cfg.View.FieldsMap)
		if err != nil {
			return "", err
		}
		if condition != "" {
			whereClause = "WHERE " + condition
		}
	}

	sql := fmt.Sprintf("SELECT %s FROM %s %s", fieldsClause, cfg.View.MetaTableName, whereClause)
	return sql, nil
}

// buildJoinSQL 生成join节点的SQL
func (g *SQLGenerator) buildJoinNodeSQL(ctx context.Context, node *interfaces.DataScopeNode) (string, error) {
	var cfg interfaces.JoinNodeCfg
	err := mapstructure.Decode(node.Config, &cfg)
	if err != nil {
		return "", err
	}

	if len(node.InputNodes) != 2 {
		return "", fmt.Errorf("join node %s requires two input nodes", node.ID)
	}

	leftNodeID := node.InputNodes[0]
	rightNodeID := node.InputNodes[1]
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
			fmt.Sprintf("lft.%s %s rgt.%s", common.QuotationMark(leftField.OriginalName), onCond.Operator, common.QuotationMark(rightField.OriginalName)))
	}
	onClause := strings.Join(onConditionsStr, " AND ")

	// 构建输出字段
	// fields := make([]string, 0, len(node.OutputFields))
	outputFieldsMap := make(map[string]*cond.ViewField)
	for _, of := range node.OutputFields {
		// var tableAlias string
		// // 要判断 src_node 是否在inout_nodes里，又多判断了一次
		// if of.SrcNodeID == leftNodeID {
		// 	tableAlias = "lft"
		// } else if of.SrcNodeID == rightNodeID {
		// 	tableAlias = "rgt"
		// } else {
		// 	return "", fmt.Errorf("output field src_node %s not in input nodes for node %s", of.SrcNodeID, node.ID)
		// }

		// fieldExpr := fmt.Sprintf("%s.%s", tableAlias, of.OriginalName)
		// fields = append(fields, fieldExpr)

		// 构造输出视图的fieldsMap, name 和 字段的映射
		outputFieldsMap[of.Name] = of
	}
	// 维护每个节点的output fields map
	node.OutputFieldsMap = outputFieldsMap
	// fieldsStr := strings.Join(fields, ", ")
	// 简化 sql 生成，join 或者 union 这里 select 的时候直接 select *
	// 实际业务使用时建议在view节点配置好字段，view节点select的时候会select具体的字段
	fieldsStr := "*"

	fieldsClause := fieldsStr
	if cfg.Distinct.Enable {
		if len(cfg.Distinct.Fields) > 0 {
			// 名称映射，将 去重的字段name 映射为视图的原始字段名original_name
			distinctFields := make([]string, 0, len(cfg.Distinct.Fields))
			for _, df := range cfg.Distinct.Fields {
				if of, ok := outputFieldsMap[df]; ok {
					distinctFields = append(distinctFields, common.QuotationMark(of.OriginalName))
				}
			}

			fieldsClause = "DISTINCT " + strings.Join(distinctFields, ", ")
		} else {
			fieldsClause = "DISTINCT " + fieldsStr
		}
	}

	whereClause := ""
	if cfg.Filters != nil {
		// join后过滤，字段应该使用 outputFieldsMap 中的字段
		condition, err := buildSQLCondition(ctx, cfg.Filters, interfaces.ViewType_Custom, outputFieldsMap)
		if err != nil {
			return "", err
		}
		if condition != "" {
			whereClause = "WHERE " + condition
		}
	}

	sql := fmt.Sprintf("SELECT %s FROM (%s) AS lft %s JOIN (%s) AS rgt ON %s %s",
		fieldsClause, leftSQL, strings.ToUpper(cfg.JoinType), rightSQL, onClause, whereClause)
	return sql, nil
}

// buildUnionNodeSQL 生成 union节点的SQL
func (g *SQLGenerator) buildUnionNodeSQL(ctx context.Context, node *interfaces.DataScopeNode) (string, error) {
	if len(node.InputNodes) < 2 {
		return "", fmt.Errorf("union node %s requires at least two input nodes", node.ID)
	}

	var cfg interfaces.UnionNodeCfg
	err := mapstructure.Decode(node.Config, &cfg)
	if err != nil {
		return "", fmt.Errorf("failed to decode union config for node %s: %v", node.ID, err)
	}

	// 检查unionFields是否和input_nodes数量一致
	if len(cfg.UnionFields) != len(node.InputNodes) {
		return "", fmt.Errorf("union node %s requires union fields for each input node", node.ID)
	}

	// 校验每个节点的合并字段数量和输出字段数量一致
	outputFieldCount := len(node.OutputFields)
	for i, uf := range cfg.UnionFields {
		if len(uf) != outputFieldCount {
			return "", fmt.Errorf("node %s has %d fields, but output requires %d fields",
				node.InputNodes[i], len(uf), outputFieldCount)
		}
	}

	// 生成所有输入节点的SQL
	inputSQLs := make([]string, len(node.InputNodes))
	for i, inputNodeID := range node.InputNodes {
		sql, err := g.buildNodeSQL(ctx, inputNodeID)
		if err != nil {
			return "", err
		}
		inputSQLs[i] = sql
	}

	// 构建字段映射, union_fields 和 input_nodes是一一对应的
	fieldMappings := make(map[string][]interfaces.UnionField)
	for i, inputNodeID := range node.InputNodes {
		fieldMappings[inputNodeID] = cfg.UnionFields[i]
	}

	// 构建SELECT子句
	selectClauses := make([]string, len(node.InputNodes))
	for i, inputNodeID := range node.InputNodes {
		nodeFields := fieldMappings[inputNodeID]
		if nodeFields == nil {
			return "", fmt.Errorf("no field mapping found for node %s in union node %s", inputNodeID, node.ID)
		}

		selectFields := make([]string, len(nodeFields))
		for j, field := range nodeFields {
			outputField := node.OutputFields[j].Name

			if field.ValueFrom == "const" {
				selectFields[j] = fmt.Sprintf("'%s' AS %s", field.Field, common.QuotationMark(outputField))
			} else {
				// field 是字段名name，需映射成original_name
				inputNodeFieldsMap, err := g.GetNodeFieldsMap(inputNodeID)
				if err != nil {
					return "", fmt.Errorf("failed to get node fields map for node %s in union node %s: %v", inputNodeID, node.ID, err)
				}
				if of, ok := inputNodeFieldsMap[field.Field]; ok {
					selectFields[j] = fmt.Sprintf("%s AS %s", common.QuotationMark(of.OriginalName), common.QuotationMark(outputField))
				} else {
					selectFields[j] = fmt.Sprintf("%s AS %s", common.QuotationMark(field.Field), common.QuotationMark(outputField))
				}
			}
		}
		selectClauses[i] = strings.Join(selectFields, ", ")
	}

	// 构建UNION SQL
	var unionType string
	if cfg.UnionType == interfaces.UnionType_All {
		unionType = "UNION ALL"
	} else if cfg.UnionType == interfaces.UnionType_Distinct {
		unionType = "UNION"
	} else {
		return "", fmt.Errorf("invalid union type %s for node %s", cfg.UnionType, node.ID)
	}

	// 构建完整的UNION查询
	unionParts := make([]string, len(node.InputNodes))
	for i := range node.InputNodes {
		unionParts[i] = fmt.Sprintf("SELECT %s FROM (%s) AS t%d", selectClauses[i], inputSQLs[i], i+1)
	}

	sql := strings.Join(unionParts, " "+unionType+" ")

	// 构建输出字段map, union的输出字段应该和第一个select的字段保持一致，outputFieldsMap 的字段key是name
	outputFieldsMap := make(map[string]*cond.ViewField)
	for _, field := range node.OutputFields {
		outputFieldsMap[field.Name] = field
	}
	// 维护每个节点的output fields map
	node.OutputFieldsMap = outputFieldsMap

	// 处理UNION后的过滤条件
	if cfg.Filters != nil {
		// union后过滤，过滤字段应该在输出字段里
		condition, err := buildSQLCondition(ctx, cfg.Filters, interfaces.ViewType_Custom, outputFieldsMap)

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
func (g *SQLGenerator) buildSqlNodeSQL(ctx context.Context, node *interfaces.DataScopeNode) (string, error) {
	// 检查input_nodes是否为空
	if len(node.InputNodes) == 0 {
		return "", fmt.Errorf("sql node %s requires at least one input node", node.ID)
	}

	// 构建输出字段map, union的输出字段应该和第一个select的字段保持一致，outputFieldsMap 的字段key是name
	outputFieldsMap := make(map[string]*cond.ViewField)
	for _, field := range node.OutputFields {
		outputFieldsMap[field.Name] = field
	}
	// 维护每个节点的output fields map
	node.OutputFieldsMap = outputFieldsMap

	var cfg interfaces.SQLNodeCfg
	err := mapstructure.Decode(node.Config, &cfg)
	if err != nil {
		return "", fmt.Errorf("failed to decode sql config for node %s: %v", node.ID, err)
	}

	// select a from {{.node1}}
	// 创建节点SQL映射上下文
	nodeSQLs := make(map[string]string)
	for _, inputNodeID := range node.InputNodes {
		sql, err := g.buildNodeSQL(ctx, inputNodeID)
		if err != nil {
			return "", fmt.Errorf("failed to build SQL for input node %s in SQL node %s: %v", inputNodeID, node.ID, err)
		}
		nodeSQLs[inputNodeID] = fmt.Sprintf("(%s)", sql)
	}

	// select a from {{table "users"}}
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

	// 解析模板
	tmpl, err := template.New("sql").Funcs(funcMap).Parse(cfg.SQLExpression)
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

	// re := regexp.MustCompile(`{{(\w+)}}`)
	// matches := re.FindAllStringSubmatch(cfg.SQLExpression, -1)
	// for _, match := range matches {
	// 	nodeID := match[1]
	// 	nodeSQL, err := g.buildNodeSQL(nodeID)
	// 	if err != nil {
	// 		return "", err
	// 	}
	// 	cfg.SQLExpression = strings.ReplaceAll(cfg.SQLExpression, match[0], nodeSQL)
	// }

	// return cfg.SQLExpression, nil
}

// buildOutputNodeSQL 生成output节点的SQL
func (g *SQLGenerator) buildOutputNodeSQL(ctx context.Context, node *interfaces.DataScopeNode) (string, error) {
	if len(node.InputNodes) != 1 {
		return "", fmt.Errorf("output node %s requires exactly one input node", node.ID)
	}

	// 构建输出字段map, union的输出字段应该和第一个select的字段保持一致，outputFieldsMap 的字段key是name
	outputFieldsMap := make(map[string]*cond.ViewField)
	for _, field := range node.OutputFields {
		outputFieldsMap[field.Name] = field
	}
	// 维护每个节点的output fields map
	node.OutputFieldsMap = outputFieldsMap

	inputNodeID := node.InputNodes[0]
	inputSQL, err := g.buildNodeSQL(ctx, inputNodeID)
	if err != nil {
		return "", err
	}

	// fields := make([]string, 0, len(node.OutputFields))
	// for _, of := range node.OutputFields {
	// 	fields = append(fields, of.Name)
	// }
	// fieldsStr := strings.Join(fields, ", ")

	// sql := fmt.Sprintf("SELECT %s FROM (%s) AS output", fieldsStr, inputSQL)

	sql := inputSQL
	return sql, nil
}

// 构造原子视图的sql
func buildAtomicViewSql(view *interfaces.DataView) string {
	return fmt.Sprintf(`SELECT * FROM %s`, view.MetaTableName)
}

// 构建时间过滤Sql
func buildTimeFilterSql(dateField string, start int64, end int64) string {
	if dateField == "" {
		return ""
	}

	return fmt.Sprintf(`%s BETWEEN from_unixtime(%d) AND from_unixtime(%d)`, dateField, start/1000, end/1000)
}

// buildCondition 构建过滤条件, fieldsMap 为这个引用视图的字段map
func buildSQLCondition(ctx context.Context, filter *cond.CondCfg, vType string, fieldsMap map[string]*cond.ViewField) (string, error) {
	var condStr string
	if filter != nil {
		// sql 查询不需要打分
		// 创建一个包含查询类型的上下文
		ctx = context.WithValue(ctx, cond.CtxKey_QueryType, interfaces.QueryType_SQL)
		condCfg, _, err := cond.NewCondition(ctx, filter, vType, fieldsMap)
		if err != nil {
			return "", fmt.Errorf("new condition failed, %v", err)
		}

		// 3. 生成sql
		if condCfg != nil {
			condStr, err = condCfg.Convert2SQL(ctx)
			if err != nil {
				return "", fmt.Errorf("convert condition to SQL failed, %v", err)
			}
		}
	}

	return condStr, nil
}

// buildRowColumnRulesSQL 构建行列规则过滤到SQL
func buildRowColumnRulesSQL(ctx context.Context, rules []*interfaces.DataViewRowColumnRule,
	view *interfaces.DataView) (string, []*cond.ViewField, map[string]*cond.ViewField, error) {

	// 行列规则长度为0， 可能查的是全量数据
	if len(rules) == 0 {
		return "", view.Fields, view.FieldsMap, nil
	}

	mergedFields := make([]*cond.ViewField, 0)
	mergedFieldsMap := map[string]*cond.ViewField{}
	for _, rule := range rules {
		// 判断列是否在视图字段列表里
		for _, field := range rule.Fields {
			if _, exists := view.FieldsMap[field]; !exists {
				return "", mergedFields, mergedFieldsMap, fmt.Errorf("field %s not found in view fields map", field)
			}

			vf := view.FieldsMap[field]
			mergedFields = append(mergedFields, vf)
			mergedFieldsMap[field] = vf
		}
	}

	mergedRowFilters := make([]*cond.CondCfg, 0)
	for _, rule := range rules {
		// 合并行规则
		if isValidFilters(rule.RowFilters) {
			mergedRowFilters = append(mergedRowFilters, rule.RowFilters)
		}
	}

	var finalCond *cond.CondCfg
	if len(mergedRowFilters) == 0 {
		finalCond = nil
	} else if len(mergedRowFilters) == 1 {
		finalCond = mergedRowFilters[0]
	} else {
		// 行列规则之间是 or 关系
		finalCond = &cond.CondCfg{
			Operation: cond.OperationOr,
			SubConds:  mergedRowFilters,
		}
	}

	// 构建行规则的SQL
	condStr, err := buildSQLCondition(ctx, finalCond, view.Type, view.FieldsMap)
	if err != nil {
		return "", mergedFields, mergedFieldsMap, err
	}

	return condStr, mergedFields, mergedFieldsMap, nil
}

func isValidFilters(cfg *cond.CondCfg) bool {
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
func buildSQLSortParams(sort []*interfaces.SortParamsV2) string {
	if len(sort) == 0 {
		return ""
	}

	var sortSql strings.Builder
	for i, sortParam := range sort {
		if i > 0 {
			sortSql.WriteString(", ")
		}
		sortSql.WriteString(fmt.Sprintf("%s %s", common.QuotationMark(sortParam.Field), sortParam.Direction))
	}

	return sortSql.String()
}

// 补充 sort 字段
func prepareSQLSortParams(sort []*interfaces.SortParamsV2, fieldsMap map[string]*cond.ViewField) []*interfaces.SortParamsV2 {

	newSort := []*interfaces.SortParamsV2{}
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
