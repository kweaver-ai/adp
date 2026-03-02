package dagmodel

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/mod"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Converter MongoDB BSON查询条件转SQL转换器
type Converter struct {
	tableName   string
	fieldMap    map[string]string // 自定义字段映射（优先级最高）
	autoConvert bool              // 是否自动转换字段名
}

type Option func(*Converter)

// WithFieldMap 自定义字段映射，优先级最高
func WithFieldMap(m map[string]string) Option {
	return func(c *Converter) {
		c.fieldMap = m
	}
}

// WithAutoConvert 开启自动字段名转换（默认开启）
func WithAutoConvert(on bool) Option {
	return func(c *Converter) {
		c.autoConvert = on
	}
}

func NewConverter(tableName string, opts ...Option) *Converter {
	c := &Converter{
		tableName:   tableName,
		fieldMap:    make(map[string]string),
		autoConvert: true,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// ==================== 字段名转换核心 ====================

// convertField 将 MongoDB 字段名转为 SQL 字段名
//
// 规则：
//  1. 优先使用自定义映射 fieldMap
//  2. 自动转换：驼峰 → f_ + 下划线分隔
//
// 示例：
//
//	_id        → f_id
//	name       → f_name
//	userId     → f_user_id
//	createdAt  → f_created_at
//	HTMLParser → f_html_parser
func (c *Converter) convertField(mongoField string) string {
	// 逻辑操作符不转换
	if strings.HasPrefix(mongoField, "$") {
		return mongoField
	}

	// 1. 优先自定义映射
	if mapped, ok := c.fieldMap[mongoField]; ok {
		return mapped
	}

	// 2. 不开启自动转换，直接返回
	if !c.autoConvert {
		return mongoField
	}

	// 3. 自动转换：camelCase / PascalCase → f_snake_case
	return camelToFSnake(mongoField)
}

// camelToFSnake 驼峰/任意命名 → f_snake_case
//
//	_id          → f_id
//	name         → f_name
//	userId       → f_user_id
//	createdAt    → f_created_at
//	myHTTPServer → f_my_http_server
//	HTMLParser   → f_html_parser
//	user_name    → f_user_name  (已经是下划线的保持不变)
//	already_f_xx → f_already_f_xx (不会重复加 f_)
func camelToFSnake(s string) string {
	if strings.HasPrefix(s, "f_") {
		return s
	}

	// 去除开头的 _ (如 _id → id，后面统一加 f_)
	s = strings.TrimLeft(s, "_")
	if s == "" {
		return "f_"
	}

	var result strings.Builder
	runes := []rune(s)

	for i, r := range runes {
		if unicode.IsUpper(r) {
			// 需要插入下划线的情况：
			// 1. 不是第一个字符
			// 2. 前一个字符不是大写（普通驼峰：userId → user_id）
			//    或者后一个字符是小写（连续大写末尾：HTMLParser → html_parser）
			if i > 0 {
				prevIsUpper := unicode.IsUpper(runes[i-1])
				nextIsLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])

				if !prevIsUpper || nextIsLower {
					result.WriteRune('_')
				}
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}

	snaked := result.String()

	// 如果已经有下划线连接的，去除多余的
	snaked = strings.ReplaceAll(snaked, "__", "_")

	return "f_" + snaked
}

// ==================== 转换主入口 ====================

type Result struct {
	SQL    string
	Conds  string
	Params []interface{}
}

func (c *Converter) Convert(query interface{}) (*Result, error) {
	m, err := toMap(query)
	if err != nil {
		return nil, fmt.Errorf("invalid query type: %w", err)
	}

	where, params, err := c.parseCondition(m)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf("SELECT * FROM %s", c.tableName)
	if where != "" {
		sql += " WHERE " + where
	}

	return &Result{SQL: sql, Params: params}, nil
}

func (c *Converter) ConvertConds(query interface{}) (*Result, error) {
	m, err := toMap(query)
	if err != nil {
		return nil, fmt.Errorf("invalid query type: %w", err)
	}

	where, params, err := c.parseCondition(m)
	if err != nil {
		return nil, err
	}

	return &Result{Conds: where, Params: params}, nil
}

// ==================== 类型统一转换 ====================

func toMap(v interface{}) (map[string]interface{}, error) {
	switch val := v.(type) {
	case bson.M:
		result := make(map[string]interface{}, len(val))
		for k, v := range val {
			result[k] = v
		}
		return result, nil
	case bson.D:
		result := make(map[string]interface{}, len(val))
		for _, elem := range val {
			result[elem.Key] = elem.Value
		}
		return result, nil
	case map[string]interface{}:
		return val, nil
	default:
		return nil, fmt.Errorf("unsupported type: %T", v)
	}
}

func toSlice(v interface{}) ([]interface{}, error) {
	switch val := v.(type) {
	case bson.A:
		return []interface{}(val), nil
	case []interface{}:
		return val, nil
	case []bson.M:
		result := make([]interface{}, len(val))
		for i, item := range val {
			result[i] = item
		}
		return result, nil
	case []bson.D:
		result := make([]interface{}, len(val))
		for i, item := range val {
			result[i] = item
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unsupported slice type: %T", v)
	}
}

func normalizeValue(v interface{}) interface{} {
	switch val := v.(type) {
	case primitive.ObjectID:
		return val.Hex()
	case primitive.DateTime:
		return val.Time().Format("2006-01-02 15:04:05")
	case time.Time:
		return val.Format("2006-01-02 15:04:05")
	case primitive.Timestamp:
		return time.Unix(int64(val.T), 0).Format("2006-01-02 15:04:05")
	case primitive.Decimal128:
		return val.String()
	case primitive.Regex:
		return val.Pattern
	case int32:
		return int64(val)
	case float32:
		return float64(val)
	default:
		return v
	}
}

// ==================== 核心解析 ====================

func (c *Converter) parseCondition(query map[string]interface{}) (string, []interface{}, error) {
	var conditions []string
	var params []interface{}

	keys := make([]string, 0, len(query))
	for k := range query {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := query[key]

		switch key {
		case "$and":
			sql, p, err := c.parseLogicalOp(value, "AND")
			if err != nil {
				return "", nil, err
			}
			conditions = append(conditions, sql)
			params = append(params, p...)

		case "$or":
			sql, p, err := c.parseLogicalOp(value, "OR")
			if err != nil {
				return "", nil, err
			}
			conditions = append(conditions, "("+sql+")")
			params = append(params, p...)

		case "$nor":
			sql, p, err := c.parseLogicalOp(value, "OR")
			if err != nil {
				return "", nil, err
			}
			conditions = append(conditions, "NOT ("+sql+")")
			params = append(params, p...)

		case "$not":
			subMap, err := toMap(value)
			if err != nil {
				return "", nil, fmt.Errorf("$not must be object: %w", err)
			}
			sql, p, err := c.parseCondition(subMap)
			if err != nil {
				return "", nil, err
			}
			conditions = append(conditions, "NOT ("+sql+")")
			params = append(params, p...)

		default:
			// 这里对字段名做转换
			sqlField := c.convertField(key)
			sqls, p, err := c.parseFieldCondition(sqlField, value)
			if err != nil {
				return "", nil, err
			}
			conditions = append(conditions, sqls...)
			params = append(params, p...)
		}
	}

	return strings.Join(conditions, " AND "), params, nil
}

func (c *Converter) parseFieldCondition(field string, value interface{}) ([]string, []interface{}, error) {
	var conditions []string
	var params []interface{}

	if subMap, err := toMap(value); err == nil {
		hasOperator := false
		for k := range subMap {
			if strings.HasPrefix(k, "$") {
				hasOperator = true
				break
			}
		}
		if hasOperator {
			ops := make([]string, 0, len(subMap))
			for op := range subMap {
				ops = append(ops, op)
			}
			sort.Strings(ops)
			for _, op := range ops {
				sql, p, err := c.parseOperator(field, op, subMap[op])
				if err != nil {
					return nil, nil, err
				}
				conditions = append(conditions, sql)
				params = append(params, p...)
			}
			return conditions, params, nil
		}
	}

	if regex, ok := value.(primitive.Regex); ok {
		likePattern := regexToLike(regex.Pattern)
		if strings.Contains(regex.Options, "i") {
			conditions = append(conditions, fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", field))
		} else {
			conditions = append(conditions, fmt.Sprintf("%s LIKE ?", field))
		}
		params = append(params, likePattern)
		return conditions, params, nil
	}

	if value == nil {
		conditions = append(conditions, fmt.Sprintf("%s IS NULL", field))
	} else {
		conditions = append(conditions, fmt.Sprintf("%s = ?", field))
		params = append(params, normalizeValue(value))
	}

	return conditions, params, nil
}

func (c *Converter) parseOperator(field, op string, value interface{}) (string, []interface{}, error) {
	switch op {
	case "$eq":
		if value == nil {
			return fmt.Sprintf("%s IS NULL", field), nil, nil
		}
		return fmt.Sprintf("%s = ?", field), []interface{}{normalizeValue(value)}, nil
	case "$ne":
		if value == nil {
			return fmt.Sprintf("%s IS NOT NULL", field), nil, nil
		}
		return fmt.Sprintf("%s != ?", field), []interface{}{normalizeValue(value)}, nil
	case "$gt":
		return fmt.Sprintf("%s > ?", field), []interface{}{normalizeValue(value)}, nil
	case "$gte":
		return fmt.Sprintf("%s >= ?", field), []interface{}{normalizeValue(value)}, nil
	case "$lt":
		return fmt.Sprintf("%s < ?", field), []interface{}{normalizeValue(value)}, nil
	case "$lte":
		return fmt.Sprintf("%s <= ?", field), []interface{}{normalizeValue(value)}, nil
	case "$in":
		return c.parseInOp(field, value, false)
	case "$nin":
		return c.parseInOp(field, value, true)
	case "$exists":
		exists, ok := value.(bool)
		if !ok {
			return "", nil, fmt.Errorf("$exists must be bool")
		}
		if exists {
			return fmt.Sprintf("%s IS NOT NULL", field), nil, nil
		}
		return fmt.Sprintf("%s IS NULL", field), nil, nil
	case "$regex":
		pattern := ""
		switch v := value.(type) {
		case string:
			pattern = v
		case primitive.Regex:
			pattern = v.Pattern
		default:
			return "", nil, fmt.Errorf("$regex must be string or Regex")
		}
		return fmt.Sprintf("%s LIKE ?", field), []interface{}{regexToLike(pattern)}, nil
	case "$mod":
		arr, err := toSlice(value)
		if err != nil || len(arr) != 2 {
			return "", nil, fmt.Errorf("$mod must be [divisor, remainder]")
		}
		return fmt.Sprintf("%s %% ? = ?", field),
			[]interface{}{normalizeValue(arr[0]), normalizeValue(arr[1])}, nil
	case "$not":
		subMap, err := toMap(value)
		if err != nil {
			if regex, ok := value.(primitive.Regex); ok {
				return fmt.Sprintf("%s NOT LIKE ?", field), []interface{}{regexToLike(regex.Pattern)}, nil
			}
			return "", nil, fmt.Errorf("$not must be object or regex")
		}
		var subConds []string
		var subParams []interface{}
		for subOp, subVal := range subMap {
			sql, p, err := c.parseOperator(field, subOp, subVal)
			if err != nil {
				return "", nil, err
			}
			subConds = append(subConds, sql)
			subParams = append(subParams, p...)
		}
		return "NOT (" + strings.Join(subConds, " AND ") + ")", subParams, nil
	case "$size":
		return fmt.Sprintf("JSON_LENGTH(%s) = ?", field), []interface{}{normalizeValue(value)}, nil
	default:
		return "", nil, fmt.Errorf("unsupported operator: %s", op)
	}
}

func (c *Converter) parseLogicalOp(value interface{}, logicOp string) (string, []interface{}, error) {
	arr, err := toSlice(value)
	if err != nil {
		return "", nil, fmt.Errorf("$%s must be array: %w", strings.ToLower(logicOp), err)
	}
	var conditions []string
	var params []interface{}
	for _, item := range arr {
		subMap, err := toMap(item)
		if err != nil {
			return "", nil, fmt.Errorf("$%s elements must be objects: %w", strings.ToLower(logicOp), err)
		}
		sql, p, err := c.parseCondition(subMap)
		if err != nil {
			return "", nil, err
		}
		if sql != "" {
			conditions = append(conditions, sql)
			params = append(params, p...)
		}
	}
	return strings.Join(conditions, fmt.Sprintf(" %s ", logicOp)), params, nil
}

func (c *Converter) parseInOp(field string, value interface{}, negate bool) (string, []interface{}, error) {
	arr, err := toSlice(value)
	if err != nil {
		return "", nil, fmt.Errorf("$in/$nin must be array: %w", err)
	}
	if len(arr) == 0 {
		if negate {
			return "1=1", nil, nil
		}
		return "1=0", nil, nil
	}
	var hasNull bool
	var vals []interface{}
	for _, v := range arr {
		if v == nil {
			hasNull = true
		} else {
			vals = append(vals, normalizeValue(v))
		}
	}
	ph := make([]string, len(vals))
	for i := range vals {
		ph[i] = "?"
	}
	var parts []string
	var params []interface{}
	if len(vals) > 0 {
		op := "IN"
		if negate {
			op = "NOT IN"
		}
		parts = append(parts, fmt.Sprintf("%s %s (%s)", field, op, strings.Join(ph, ", ")))
		params = append(params, vals...)
	}
	if hasNull {
		if negate {
			parts = append(parts, fmt.Sprintf("%s IS NOT NULL", field))
		} else {
			parts = append(parts, fmt.Sprintf("%s IS NULL", field))
		}
	}
	joiner := " OR "
	if negate {
		joiner = " AND "
	}
	result := strings.Join(parts, joiner)
	if len(parts) > 1 {
		result = "(" + result + ")"
	}
	return result, params, nil
}

func regexToLike(pattern string) string {
	pattern = strings.TrimPrefix(pattern, "/")
	if idx := strings.LastIndex(pattern, "/"); idx > 0 {
		pattern = pattern[:idx]
	}
	result := pattern
	hasPrefix := strings.HasPrefix(result, "^")
	hasSuffix := strings.HasSuffix(result, "$")
	result = strings.TrimPrefix(result, "^")
	result = strings.TrimSuffix(result, "$")
	result = strings.ReplaceAll(result, "\\.", "{{DOT}}")
	result = strings.ReplaceAll(result, "%", "\\%")
	result = strings.ReplaceAll(result, "_", "\\_")
	result = strings.ReplaceAll(result, ".*", "%")
	result = strings.ReplaceAll(result, ".+", "_%")
	result = strings.ReplaceAll(result, "{{DOT}}", ".")
	if !hasPrefix {
		result = "%" + result
	}
	if !hasSuffix {
		result = result + "%"
	}
	return result
}

const ObjTypeDag uint8 = 1
const ObjTypeDagInstance uint8 = 2

func BuildListDagInstanceQuery(input *mod.ListDagInstanceInput, cnt bool) (string, []interface{}) {
	var conds []string
	var args []interface{}
	var dynKeys []string
	var dynArgs []interface{}
	var dagIDs []uint64
	var status []string

	for _, v := range input.DagIDs {
		id, _ := strconv.ParseUint(v, 10, 64)
		dagIDs = append(dagIDs, id)
	}

	for _, v := range input.Status {
		status = append(status, string(v))
	}

	if input.HasCmd {
		dynKeys = append(dynKeys, "(f_key = ?)")
		dynArgs = append(dynArgs, "cmd")
	}

	if input.TimeRange != nil {
		dynKeys = append(dynKeys, "(f_key = ? AND f_value >= ? AND f_value <= ?)")
		dynArgs = append(dynArgs, input.TimeRange.Field, input.TimeRange.Begin, input.TimeRange.End)
	}

	if input.MatchQuery != nil {
		dynKeys = append(dynKeys, "(f_key = ? AND f_value = ?)")
		dynArgs = append(dynArgs, input.MatchQuery.Field, input.MatchQuery.Value)
	}

	needJoin := len(dynKeys) > 0
	p := ""
	if needJoin {
		p = "di."
	}

	// 主表条件
	addSlice := func(col string, vals []interface{}) {
		if len(vals) > 0 {
			conds = append(conds, fmt.Sprintf("%s%s IN (%s)", p, camelToFSnake(col), strings.TrimSuffix(strings.Repeat("?,", len(vals)), ",")))
			args = append(args, vals...)
		}
	}
	addU64Slice := func(col string, ids []uint64) {
		v := make([]interface{}, len(ids))
		for i, id := range ids {
			v[i] = id
		}
		addSlice(col, v)
	}
	addStrSlice := func(col string, ss []string) {
		v := make([]interface{}, len(ss))
		for i, s := range ss {
			v[i] = s
		}
		addSlice(col, v)
	}

	addU64Slice("f_dag_id", dagIDs)
	addStrSlice("f_status", status)
	addStrSlice("f_user_id", input.UserIDs)
	addSlice("f_priority", input.Priority)
	if input.Worker != "" {
		conds = append(conds, p+"f_worker = ?")
		args = append(args, input.Worker)
	}
	if input.ExcludeModeVM {
		conds = append(conds, p+"f_mode < 1")
	}
	if input.UpdatedEnd > 0 {
		conds = append(conds, p+"f_updated_at <= ?")
		args = append(args, input.UpdatedEnd)
	}

	// 构建 SQL
	var sql string
	if needJoin {
		// 子查询: GROUP BY + HAVING
		sub, subArgs := BuildMetadataSearchQuery(dynKeys, dynArgs, ObjTypeDagInstance)

		if cnt {
			sql = "SELECT COUNT(*) FROM dag_instance di INNER JOIN (" + sub + ") ms ON di.f_id = ms.f_obj_id"
		} else {
			sql = "SELECT di.* FROM t_dag_instance di INNER JOIN (" + sub + ") ms ON di.f_id = ms.f_obj_id"
		}
		args = append(subArgs, args...)
	} else {
		if cnt {
			sql = "SELECT COUNT(*) FROM t_dag_instance"
		} else {
			sql = "SELECT * FROM t_dag_instance"
		}
	}

	if len(conds) > 0 {
		sql += " WHERE " + strings.Join(conds, " AND ")
	}

	if !cnt {
		if input.SortBy != "" {
			dir := utils.IfNot(input.Order == 0, "DESC", "ASC")
			sql += fmt.Sprintf(" ORDER BY %s%s %s", p, input.SortBy, dir)
		}
		if input.Limit > 0 {
			sql += " LIMIT ? OFFSET ?"
			args = append(args, input.Limit, input.Limit*input.Offset)
		}
	}

	return sql, args
}

func BuildMetadataSearchQuery(dynKeys []string, dynArgs []interface{}, objType uint8) (string, []interface{}) {
	sub := fmt.Sprintf(
		"SELECT f_obj_id FROM t_metadata_search WHERE f_type = ? AND (%s) GROUP BY f_obj_id HAVING COUNT(DISTINCT f_key) = ?",
		strings.Join(dynKeys, " OR "),
	)
	subArgs := append([]interface{}{objType}, dynArgs...)
	subArgs = append(subArgs, len(dynKeys))
	return sub, subArgs
}

func BuildDagVars(dag *entity.Dag) []*DagVar {
	var vars []*DagVar
	for k, v := range dag.Vars {
		id, _ := utils.GetUniqueID()
		dagID, _ := strconv.ParseUint(dag.ID, 10, 64)
		vars = append(vars, &DagVar{
			ID:           id,
			DagID:        dagID,
			VarName:      k,
			DefaultValue: v.DefaultValue,
			VarType:      "string",
			Description:  v.Desc,
		})
	}

	return vars
}

func BuildDagSearch(dag *entity.Dag) []*MetadataSearch {
	datas := []*MetadataSearch{}
	dagID, _ := strconv.ParseUint(dag.ID, 10, 64)

	buildSearch := func(key, value, vtype string) *MetadataSearch {
		id, _ := utils.GetUniqueID()
		return &MetadataSearch{
			ID:        id,
			ObjID:     dagID,
			Key:       key,
			Value:     value,
			ValueType: vtype,
			Type:      "dag",
		}
	}

	for _, accessor := range dag.Accessors {
		datas = append(datas, buildSearch("accessor.id", accessor.ID, "string"))
	}

	for _, task := range dag.Tasks {
		if docid, ok := task.Params["docid"]; ok {
			datas = append(datas, buildSearch("steps.*.parameters.docid", fmt.Sprintf("%v", docid), "string"))
		}
		if docids, ok := task.Params["docids"]; ok {
			for _, docid := range docids.([]interface{}) {
				datas = append(datas, buildSearch("steps.*.parameters.docid", fmt.Sprintf("%v", docid), "string"))
			}
		}
		datas = append(datas, buildSearch("steps.*.operator", task.ActionName, "string"))
	}

	return datas
}
