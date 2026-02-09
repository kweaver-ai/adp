// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_view

import (
	"fmt"
	"strings"

	"uniquery/logics/data_view/parsing" // 导入生成的解析器

	"github.com/antlr4-go/antlr/v4"
	"github.com/bytedance/sonic"
)

// FieldInfo 表示SQL查询输出的字段信息
type FieldInfo struct {
	Name      string `json:"name"`       // 字段名或表达式
	Alias     string `json:"alias"`      // 字段别名（如果没有别名为空）
	IsStar    bool   `json:"is_star"`    // 是否是*通配符
	IsComplex bool   `json:"is_complex"` // 是否是复杂表达式（函数、CASE等）
}

// QueryAnalysis 表示SQL查询的分析结果
type QueryAnalysis struct {
	Fields       []FieldInfo `json:"fields"`
	HasStar      bool        `json:"has_star"`
	HasUnion     bool        `json:"has_union"`
	HasJoin      bool        `json:"has_join"`
	HasAggregate bool        `json:"has_aggregate"`
	HasSubquery  bool        `json:"has_subquery"`
	HasCase      bool        `json:"has_case"`
	Error        error       `json:"error,omitempty"`
}

// String 返回分析结果的字符串表示
func (q *QueryAnalysis) String() string {
	if q.Error != nil {
		return fmt.Sprintf("分析错误: %v", q.Error)
	}

	result := fmt.Sprintf("查询字段 (%d 个):\n", len(q.Fields))
	for i, field := range q.Fields {
		fieldDesc := field.Name
		if field.Alias != "" {
			fieldDesc = fmt.Sprintf("%s AS %s", field.Name, field.Alias)
		}
		if field.IsStar {
			fieldDesc = "*"
		}
		result += fmt.Sprintf("  %d. %s\n", i+1, fieldDesc)
	}

	result += "\n查询特征:\n"
	result += fmt.Sprintf("  - 包含UNION: %t\n", q.HasUnion)
	result += fmt.Sprintf("  - 包含JOIN: %t\n", q.HasJoin)
	result += fmt.Sprintf("  - 包含聚合函数: %t\n", q.HasAggregate)
	result += fmt.Sprintf("  - 包含子查询: %t\n", q.HasSubquery)
	result += fmt.Sprintf("  - 包含CASE表达式: %t\n", q.HasCase)

	return result
}

// SQLFieldParser SQL字段解析器
type SQLFieldParser struct {
	listener *sqlFieldListener
}

// NewSQLFieldParser 创建新的SQL字段解析器
func NewSQLFieldParser() *SQLFieldParser {
	return &SQLFieldParser{
		listener: newSqlFieldListener(),
	}
}

// Parse 解析SQL语句并返回字段分析结果
func (p *SQLFieldParser) Parse(sql string) *QueryAnalysis {
	// 创建输入流
	input := antlr.NewInputStream(sql)

	// 创建词法分析器
	lexer := parsing.NewSqlBaseLexer(input)

	// 创建令牌流
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	// 创建语法分析器
	parser := parsing.NewSqlBaseParser(stream)

	// 添加错误监听器
	parser.RemoveErrorListeners()
	errorListener := newErrorListener()
	parser.AddErrorListener(errorListener)

	// 构建解析树 - 根据实际语法规则调整起始规则
	tree := parser.Query() // 或者可能是 parser.Query() 或 parser.Statement()

	// 遍历解析树
	antlr.ParseTreeWalkerDefault.Walk(p.listener, tree)

	analysis := p.listener.getAnalysis()
	if errorListener.hasErrors() {
		analysis.Error = fmt.Errorf("SQL语法错误: %s", strings.Join(errorListener.getErrors(), "; "))
	}

	return analysis
}

// sqlFieldListener 自定义字段解析监听器
type sqlFieldListener struct {
	*parsing.BaseSqlBaseListener
	analysis          *QueryAnalysis
	currentQueryLevel int
	inSelectClause    bool
	currentField      *FieldInfo
}

// newSqlFieldListener 创建新的字段监听器
func newSqlFieldListener() *sqlFieldListener {
	return &sqlFieldListener{
		analysis: &QueryAnalysis{
			Fields: make([]FieldInfo, 0),
		},
		currentQueryLevel: 0,
		inSelectClause:    false,
	}
}

// getAnalysis 获取分析结果
func (l *sqlFieldListener) getAnalysis() *QueryAnalysis {
	return l.analysis
}

// EnterQuery 进入查询
func (l *sqlFieldListener) EnterQuery(ctx *parsing.QueryContext) {
	l.currentQueryLevel++
}

// ExitQuery 退出查询
func (l *sqlFieldListener) ExitQuery(ctx *parsing.QueryContext) {
	l.currentQueryLevel--
}

// EnterQuerySpecification 进入查询规范（SELECT语句）
func (l *sqlFieldListener) EnterQuerySpecification(ctx *parsing.QuerySpecificationContext) {
	l.inSelectClause = true
}

// ExitQuerySpecification 退出查询规范（SELECT语句）
func (l *sqlFieldListener) ExitQuerySpecification(ctx *parsing.QuerySpecificationContext) {
	l.inSelectClause = false
}

// EnterSelectSingle 处理SELECT单个字段
func (l *sqlFieldListener) EnterSelectSingle(ctx *parsing.SelectSingleContext) {
	if !l.inSelectClause || l.currentQueryLevel > 1 {
		// 只处理最外层的SELECT字段
		return
	}

	// 获取选择项的文本
	itemText := l.getText(ctx)

	// 创建字段信息
	l.currentField = &FieldInfo{
		Name:      itemText,
		IsStar:    l.isStarExpression(itemText),
		IsComplex: l.isComplexExpression(itemText),
	}

	if l.isStarExpression(itemText) {
		l.analysis.HasStar = true
	}

	// 检查是否有别名
	if alias := l.extractAlias(ctx); alias != "" {
		l.currentField.Alias = alias
		l.currentField.Name = strings.TrimSuffix(l.currentField.Name, " AS "+alias)
		l.currentField.Name = strings.TrimSuffix(l.currentField.Name, " "+alias)
	}

	// 添加到分析结果
	l.analysis.Fields = append(l.analysis.Fields, *l.currentField)
	l.currentField = nil
}

// isStarExpression 检查是否是*通配符
func (l *sqlFieldListener) isStarExpression(expr string) bool {
	return strings.TrimSpace(expr) == "*"
}

// isComplexExpression 检查是否是复杂表达式
func (l *sqlFieldListener) isComplexExpression(expr string) bool {
	trimmed := strings.ToUpper(strings.TrimSpace(expr))

	// 检查函数调用
	if strings.Contains(trimmed, "(") && strings.Contains(trimmed, ")") {
		return true
	}

	// 检查CASE表达式
	if strings.HasPrefix(trimmed, "CASE") {
		l.analysis.HasCase = true
		return true
	}

	// 检查聚合函数
	if l.isAggregateFunction(trimmed) {
		l.analysis.HasAggregate = true
		return true
	}

	// 检查算术表达式
	if strings.ContainsAny(trimmed, "+-*/%") {
		return true
	}

	return false
}

// isAggregateFunction 检查是否是聚合函数
func (l *sqlFieldListener) isAggregateFunction(expr string) bool {
	upperExpr := strings.ToUpper(expr)
	aggregateFuncs := []string{
		"COUNT(", "SUM(", "AVG(", "MIN(", "MAX(",
		"GROUP_CONCAT(", "ARRAY_AGG(", "STRING_AGG(",
	}

	for _, funcName := range aggregateFuncs {
		if strings.Contains(upperExpr, funcName) {
			return true
		}
	}
	return false
}

// extractAlias 提取字段别名
func (l *sqlFieldListener) extractAlias(ctx antlr.ParserRuleContext) string {
	// 根据实际的语法规则提取别名
	// 这里是一个通用实现，您可能需要根据您的g4语法调整

	children := ctx.GetChildren()
	for _, child := range children {
		if terminal, ok := child.(antlr.TerminalNode); ok {
			text := terminal.GetText()
			upperText := strings.ToUpper(text)
			if upperText == "AS" {
				// 找到AS关键字，下一个兄弟节点应该是别名
				return l.getNextSiblingText(child)
			}
		}
	}

	// 如果没有AS关键字，检查最后一个子节点是否可能是别名
	// 这需要根据具体语法规则调整
	return ""
}

// getNextSiblingText 获取下一个兄弟节点的文本
func (l *sqlFieldListener) getNextSiblingText(node antlr.Tree) string {
	parent := getParent(node)
	if parent == nil {
		return ""
	}

	children := parent.GetChildren()
	found := false
	for _, child := range children {
		if found {
			return l.getText(child)
		}
		if child == node {
			found = true
		}
	}
	return ""
}

// getParent 获取节点的父节点
func getParent(node antlr.Tree) antlr.Tree {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case antlr.ParserRuleContext:
		return n.GetParent()
	case antlr.TerminalNode:
		return n.GetParent()
	default:
		return nil
	}
}

// getText 安全获取节点的文本内容
func (l *sqlFieldListener) getText(node antlr.Tree) string {
	if node == nil {
		return ""
	}

	switch ctx := node.(type) {
	case antlr.ParserRuleContext:
		return ctx.GetText()
	case antlr.TerminalNode:
		return ctx.GetText()
	default:
		return fmt.Sprintf("%v", node)
	}
}

// EnterSetOperation 处理UNION操作
func (l *sqlFieldListener) EnterSetOperation(ctx *parsing.SetOperationContext) {
	l.analysis.HasUnion = true
}

// EnterJoinRelation 处理JOIN关系
func (l *sqlFieldListener) EnterJoinRelation(ctx *parsing.JoinRelationContext) {
	l.analysis.HasJoin = true
}

// EnterSubquery 处理子查询
func (l *sqlFieldListener) EnterSubquery(ctx *parsing.SubqueryContext) {
	if l.currentQueryLevel > 0 {
		l.analysis.HasSubquery = true
	}
}

// errorListener 自定义错误监听器
type errorListener struct {
	*antlr.DefaultErrorListener
	errors []string
}

// newErrorListener 创建错误监听器
func newErrorListener() *errorListener {
	return &errorListener{
		errors: make([]string, 0),
	}
}

// SyntaxError 处理语法错误
func (l *errorListener) SyntaxError(
	recognizer antlr.Recognizer,
	offendingSymbol interface{},
	line, column int,
	msg string,
	e antlr.RecognitionException,
) {
	errorMsg := fmt.Sprintf("第%d行第%d列: %s", line, column, msg)
	l.errors = append(l.errors, errorMsg)
}

// hasErrors 检查是否有错误
func (l *errorListener) hasErrors() bool {
	return len(l.errors) > 0
}

// getErrors 获取所有错误
func (l *errorListener) getErrors() []string {
	return l.errors
}

// GetFieldNames 获取所有字段名称（优先使用别名，没有别名使用字段名）
func (q *QueryAnalysis) GetFieldNames() []string {
	names := make([]string, len(q.Fields))
	for i, field := range q.Fields {
		if field.Alias != "" {
			names[i] = field.Alias
		} else if field.IsStar {
			names[i] = "*"
		} else {
			names[i] = field.Name
		}
	}
	return names
}

// HasComplexFields 检查是否包含复杂字段
func (q *QueryAnalysis) HasComplexFields() bool {
	for _, field := range q.Fields {
		if field.IsComplex {
			return true
		}
	}
	return false
}

// GetSimpleFieldNames 获取简单字段名称（排除复杂表达式和*）
func (q *QueryAnalysis) GetSimpleFieldNames() []string {
	var names []string
	for _, field := range q.Fields {
		if !field.IsComplex && !field.IsStar {
			if field.Alias != "" {
				names = append(names, field.Alias)
			} else {
				names = append(names, field.Name)
			}
		}
	}
	return names
}

// FormatAsJSON 将解析结果格式化为 JSON
func (info *QueryAnalysis) FormatAsJSON() string {
	jsonData, err := sonic.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "JSON 格式化失败: %v"}`, err)
	}
	return string(jsonData)
}
