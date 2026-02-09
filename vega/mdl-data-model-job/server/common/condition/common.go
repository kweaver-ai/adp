// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"
)

const (
	KEYWORD_SUFFIX           = "keyword"
	DESENSITIZE_FIELD_SUFFIX = "_desensitize"

	DefaultField = "*"
)

func getFilterFieldName(name string, fieldsMap map[string]*Field, isFullTextQuery bool) string {
	// 全文检索字段为 "*"
	if name == DefaultField {
		return name
	}

	// 如果是脱敏字段，字段添加后缀 _desensitize
	desensitizeFieldName := name + DESENSITIZE_FIELD_SUFFIX

	fieldType, ok1 := fieldsMap[name]
	_, ok2 := fieldsMap[desensitizeFieldName]
	if ok1 && ok2 {
		// 脱敏字段
		name = desensitizeFieldName
	}

	// 全文检索情况下，text 类型的字段不需要添加 keyword 后缀
	// 精确查询情况下，text 类型的字段给字段名加上后缀 .keyword
	if !isFullTextQuery && fieldType.Type == DataType_Text {
		name = wrapKeyWordFieldName(name)
	}

	return name
}

// 转换成 keyword
func wrapKeyWordFieldName(fields ...string) string {
	for _, field := range fields {
		if field == "" {
			return ""
		}
	}

	return strings.Join(fields, ".") + "." + KEYWORD_SUFFIX
}

const (
	DataType_Keyword = "keyword"
	DataType_Text    = "text"
	DataType_Binary  = "binary"

	DataType_Byte      = "byte"
	DataType_Short     = "short"
	DataType_Integer   = "integer"
	DataType_Long      = "long"
	DataType_HalfFloat = "half_float"
	DataType_Float     = "float"
	DataType_Double    = "double"

	DataType_Boolean = "boolean"

	DataType_Date = "date"

	DataType_Ip       = "ip"
	DataType_GeoPoint = "geo_point"
	DataType_GeoShape = "geo_shape"
)

func DataType_IsString(t string) bool {
	return (t == DataType_Keyword || t == DataType_Text)
}

const (
	OperationAnd = "and"
	OperationOr  = "or"

	OperationEq          = "=="
	OperationNotEq       = "!="
	OperationGt          = ">"
	OperationGte         = ">="
	OperationLt          = "<"
	OperationLte         = "<="
	OperationIn          = "in"
	OperarionNotIn       = "not_in"
	OperationLike        = "like"
	OperationNotLike     = "not_like"
	OperationContain     = "contain"
	OperationNotContain  = "not_contain"
	OperationRange       = "range"
	OperationOutRange    = "out_range"
	OperationExist       = "exist"
	OperationNotExist    = "not_exist"
	OperationEmpty       = "empty"
	OperationNotEmpty    = "not_empty"
	OperationRegex       = "regex"
	OperationMatch       = "match"
	OperationMatchPhrase = "match_phrase"
)

const (
	ValueFrom_Const = "const"
)

var (
	OperationMap = map[string]struct{}{
		OperationAnd:         {},
		OperationOr:          {},
		OperationEq:          {},
		OperationNotEq:       {},
		OperationGt:          {},
		OperationGte:         {},
		OperationLt:          {},
		OperationLte:         {},
		OperationIn:          {},
		OperarionNotIn:       {},
		OperationLike:        {},
		OperationNotLike:     {},
		OperationContain:     {},
		OperationNotContain:  {},
		OperationRange:       {},
		OperationOutRange:    {},
		OperationExist:       {},
		OperationNotExist:    {},
		OperationEmpty:       {},
		OperationNotEmpty:    {},
		OperationRegex:       {},
		OperationMatch:       {},
		OperationMatchPhrase: {},
	}
)

type Filter struct {
	Name      string `json:"name"`
	Operation string `json:"operation"`
	Value     any    `json:"value"`
}

type CondCfg struct {
	Name        string     `json:"field,omitempty" mapstructure:"field"`
	Operation   string     `json:"operation,omitempty" mapstructure:"operation"`
	SubConds    []*CondCfg `json:"sub_conditions,omitempty" mapstructure:"sub_conditions"`
	ValueOptCfg `mapstructure:",squash"`

	NameField *Field `json:"-" mapstructure:"-"`
}

func (cfg *CondCfg) String() string {
	return fmt.Sprintf("{field = %s, operation = %s, sub_conditions = %v, value_from = %s, value = %v}",
		cfg.Name, cfg.Operation, cfg.SubConds, cfg.ValueFrom, cfg.Value)
}

type ValueOptCfg struct {
	ValueFrom string `json:"value_from,omitempty" mapstructure:"value_from"`
	Value     any    `json:"value,omitempty" mapstructure:"value"`
}

type Field struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Comment string `json:"comment"`

	Path []string `json:"-"`
}

func (field *Field) String() string {
	return fmt.Sprintf("{name = %s, type = %s, comment = %s}",
		field.Name, field.Type, field.Comment)
}

func (field *Field) InitFieldPath() {
	if len(field.Path) == 0 {
		field.Path = strings.Split(field.Name, ".")
	}
}

func IsSlice(i any) bool {
	kind := reflect.ValueOf(i).Kind()
	return kind == reflect.Slice || kind == reflect.Array
}

func IsSameType(arr []any) bool {
	if len(arr) == 0 {
		return true
	}

	firstType := reflect.TypeOf(arr[0])
	for _, v := range arr {
		if reflect.TypeOf(v) != firstType {
			return false
		}
	}

	return true
}

type Int interface {
	int | int8 | int16 | int32 | int64
}

type Uint interface {
	uint | uint8 | uint16 | uint32 | uint64
}

type Float interface {
	float32 | float64
}

type Bool interface {
	bool
}

func compareBool(left bool, right bool, op string) bool {
	switch op {
	case OperationEq:
		return left == right
	case OperationNotEq:
		return left != right
	}
	return false
}

func compareValue[T Int | Uint | Float | string](left, right T, op string) bool {
	switch op {
	case OperationEq:
		return left == right
	case OperationNotEq:
		return left != right
	}
	return false
}

func compareTime(left time.Time, right time.Time, op string) bool {
	switch op {
	case OperationEq:
		return left.Equal(right)
	case OperationNotEq:
		return !left.Equal(right)
	}
	return false
}

func compare(ctx context.Context, data *OriginalData, cfg *CondCfg) (bool, error) {
	lv, err := data.GetSingleData(ctx, cfg.NameField)
	if err != nil {
		return false, err
	}
	if lv == nil {
		return false, nil
	}

	rv := cfg.Value
	if rv == nil {
		return false, nil
	}

	switch cfg.NameField.Type {
	case DataType_Boolean:
		return compareBool(lv.(bool), rv.(bool), cfg.Operation), nil

	case DataType_Short, DataType_Integer, DataType_Long:
		return compareValue(lv.(int64), rv.(int64), cfg.Operation), nil

	case DataType_Float, DataType_Double:
		return compareValue(lv.(float64), rv.(float64), cfg.Operation), nil

	case DataType_Keyword, DataType_Text:
		return compareValue(lv.(string), rv.(string), cfg.Operation), nil

	case DataType_Date:
		return compareTime(lv.(time.Time), rv.(time.Time), cfg.Operation), nil
	}

	return false, nil
}

type OriginalData struct {
	Origin map[string]any
	Output map[string][]any
}

func (data *OriginalData) GetSingleData(ctx context.Context, field *Field) (any, error) {
	v, err := data.GetData(ctx, field)
	if err != nil {
		return nil, err
	}

	if len(v) == 0 {
		return nil, nil
	} else if len(v) > 1 {
		return nil, fmt.Errorf("only support single data: %v", v)
	}

	return v[0], nil
}

func (pd *OriginalData) GetData(ctx context.Context, field *Field) ([]any, error) {
	if vData, ok := pd.Output[field.Name]; ok {
		return vData, nil
	}

	field.InitFieldPath()
	oDatas, err := GetDatasByPath(ctx, pd.Origin, field.Path)
	if err != nil {
		return nil, err
	}

	if len(oDatas) == 0 {
		// logger.Debugf("this field %s does not exist in the data", strings.Join(field.Path, "."))
		return nil, nil
	}

	vData := []any{}
	for _, oData := range oDatas {
		v, err := ConvertValueByType(ctx, field.Type, oData)
		if err != nil {
			return nil, err
		}
		vData = append(vData, v)
	}

	pd.Output[field.Name] = vData
	return vData, nil
}

// array 里面相同类型 可以获取内部数据，如果非相同类型，则nil
func GetDatasByPath(ctx context.Context, obj any, path []string) ([]any, error) {
	if reflect.TypeOf(obj) == nil {
		return []any{}, nil
	}

	current := obj
	for idx := 0; idx < len(path); idx++ {
		switch reflect.TypeOf(current).Kind() {
		case reflect.Map:
			value, ok := current.(map[string]any)[path[idx]]
			if !ok || value == nil {
				return []any{}, nil
			}
			// found
			current = value

		case reflect.Slice:
			res := []any{}
			for _, sub := range current.([]any) {
				subRes, err := GetDatasByPath(ctx, sub, path[idx:])
				if err != nil {
					return []any{}, err
				}
				res = append(res, subRes...)
			}
			return res, nil

		default:
			// invalid path
			return []any{}, nil
		}
	}

	// path is empty
	return GetLastDatas(ctx, current)
}

func GetLastDatas(ctx context.Context, obj any) (res []any, err error) {
	if obj == nil {
		return []any{}, nil
	}
	switch reflect.TypeOf(obj).Kind() {
	case reflect.Slice:
		for _, sub := range obj.([]any) {
			subRes, err := GetLastDatas(ctx, sub)
			if err != nil {
				return []any{}, err
			}
			res = append(res, subRes...)
		}
	default:
		res = append(res, obj)
		return res, nil
	}

	return res, nil
}

// result coule be single value or slice value
// result type should only be string, time.Time, int64, float64, bool
func ConvertValueByType(ctx context.Context, vType string, vData any) (any, error) {
	if vData == nil {
		return nil, fmt.Errorf("vData is nil")
	}
	switch v := vData.(type) {
	case string:
		switch vType {
		case DataType_Date:
			v, err := time.Parse(time.RFC3339Nano, vData.(string))
			if err != nil {
				return nil, err
			}
			return v, nil
		case DataType_Keyword, DataType_Text:
			return v, nil
		default:
			return nil, fmt.Errorf("invalid value")
		}

	case float64:
		switch vType {
		case DataType_Short:
			v := int64(int16(vData.(float64)))
			return v, nil
		case DataType_Integer:
			v := int64(int32(vData.(float64)))
			return v, nil
		case DataType_Long:
			v := int64(vData.(float64))
			return v, nil
		case DataType_Float:
			v := float64(float32(vData.(float64)))
			return v, nil
		case DataType_Double:
			return v, nil
		default:
			return nil, fmt.Errorf("invalid value")
		}

	case bool:
		switch vType {
		case DataType_Boolean:
			return v, nil
		default:
			return nil, fmt.Errorf("invalid value")
		}

	default:
		return nil, fmt.Errorf("invalid value")
	}
}
