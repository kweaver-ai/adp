// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"reflect"

	vopt "vega-backend/interfaces/value_opt"
)

const (
	AllField = "*"
)

const (
	OperationAnd = "and"
	OperationOr  = "or"

	OperationEq          = "=="
	OperationEq2         = "eq"
	OperationNotEq       = "!="
	OperationNotEq2      = "not_eq"
	OperationGt          = ">"
	OperationGt2         = "gt"
	OperationGte         = ">="
	OperationGte2        = "gte"
	OperationLt          = "<"
	OperationLt2         = "lt"
	OperationLte         = "<="
	OperationLte2        = "lte"
	OperationIn          = "in"
	OperationNotIn       = "not_in"
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
	OperationPrefix      = "prefix"
	OperationNotPrefix   = "not_prefix"
	OperationNull        = "null"
	OperationNotNull     = "not_null"
	OperationTrue        = "true"
	OperationFalse       = "false"
	OperationBefore      = "before"
	OperationCurrent     = "current"
	OperationBetween     = "between"
	OperationKnnVector   = "knn_vector"
	OperationMultiMatch  = "multi_match"
)

type CondCfg struct {
	Name             string     `json:"field,omitempty" mapstructure:"field"` // 传递name
	Operation        string     `json:"operation,omitempty" mapstructure:"operation"`
	SubConds         []*CondCfg `json:"sub_conditions,omitempty" mapstructure:"sub_conditions"`
	vopt.ValueOptCfg `mapstructure:",squash"`

	RemainCfg map[string]any `mapstructure:",remain"`
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
