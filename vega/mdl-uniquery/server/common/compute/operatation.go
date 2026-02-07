// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package compute

import (
	"reflect"
	"strings"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
)

// 小于
func Lt(value interface{}, element interface{}) bool {

	t := reflect.TypeOf(element)

	if t.Kind() == reflect.Float64 {
		return value.(float64) < element.(float64)
	}
	if t.Kind() == reflect.Float32 {
		return value.(float32) < element.(float32)
	}
	if t.Kind() == reflect.Int32 {
		return value.(int32) < element.(int32)
	}
	if t.Kind() == reflect.String {
		ret := value.(string) < element.(string)
		return ret
	}

	return false

}

func Op_in(value interface{}, elements []interface{}) bool {
	for _, elem := range elements {
		if value == elem {
			return true
		}
	}
	return false
}

func Op_contain(values []interface{}, elements []interface{}) bool {
	for _, elem := range elements {
		if Op_in(elem, values) {
			return true
		}
	}
	return false
}

func Equal(value any, element any) bool {
	return value == element
}

func Like(value interface{}, element interface{}) bool {
	val, ok := value.(string)
	ele, okk := element.(string)
	if ok && okk {
		return strings.Contains(val, ele)
	} else {
		return false
	}

}

func processSingleValue(operator string, value any, target interface{}) (hit bool) {
	switch operator {
	case "like":
		hit = Like(value, target)
	case "not_like", "notlike", "NotLike":
		hit = !Like(value, target)
	case "==", "=":
		hit = Equal(value, target)
	case "!=":
		hit = !Equal(value, target)
	case ">":
		hit = !(Lt(value, target) || Equal(value, target))
	case ">=":
		hit = !Lt(value, target)
	case "<":
		hit = Lt(value, target)
	case "<=":
		hit = Lt(value, target) || Equal(value, target)
	}
	return hit

}

func processMultiple(operator string, value any, elements []interface{}) (hit bool) {
	switch operator {
	case "range":
		hit = (Lt(elements[0], value) || Equal(value, elements[0])) && Lt(value, elements[len(elements)-1])
	case "not_range", "notrange", "NotRange":
		hit = Lt(value, elements[0]) || (Lt(elements[len(elements)-1], value) || Equal(value, elements[len(elements)-1]))
	case "in":
		hit = Op_in(value, elements)

	case "contain":
		if reflect.ValueOf(value).Kind() == reflect.Slice {
			var values []any
			values = append(values, value.([]any)...)
			hit = Op_contain(values, elements)
		}

	}
	return hit
}

// NOTE: 执行操作符，得到结果
func Exec(operator string, value any, target interface{}) (hit bool) {

	logger.Debugf("detect judge record: value %#v\n target %#v,operator: %#v", value, target, operator)
	if value == nil {
		return false
	}
	switch val := target.(type) {

	case float64, string:
		hit = processSingleValue(operator, value, target)
	case []interface{}:
		hit = processMultiple(operator, value, val)
	}

	return hit
}

type FieldValueType interface {
	int | float64 | string
}

func In[T FieldValueType](fieldValue T, targetValues []T) (bool, error) {
	var flag = false
	for _, targetValue := range targetValues {
		if targetValue == fieldValue {
			flag = true
		} else {
			continue
		}

	}
	return flag, nil
}

func Contain[T FieldValueType](fieldValues []T, targetValues []T) (bool, error) {
	var flag = false
	for _, targetValue := range targetValues {
		ok, _ := In(targetValue, fieldValues)
		if ok {
			return true, nil
		}
	}

	return flag, nil
}

func EqualTo[T FieldValueType](fieldValue T, targetValue T) (bool, error) {
	return fieldValue == targetValue, nil
}
