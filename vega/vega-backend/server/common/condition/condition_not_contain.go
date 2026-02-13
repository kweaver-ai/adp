// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"

	"vega-backend/interfaces"
	vopt "vega-backend/interfaces/value_opt"
)

type NotContainCond struct {
	mCfg   *CondCfg
	mField *interfaces.Property
	mValue []any
}

// 不包含 not_contain，左侧属性值为数组，右侧值为数组，组内的值都应在属性值外
func NewNotContainCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*interfaces.Property) (Condition, error) {
	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [not_contain] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [not_contain] left field '%s' not found", cfg.Name)
	}

	if cfg.ValueOptCfg.ValueFrom != vopt.ValueFrom_Const {
		return nil, fmt.Errorf("condition [not_contain] does not support value_from type '%s'", cfg.ValueFrom)
	}
	val, ok := cfg.ValueOptCfg.Value.([]any)
	if !ok {
		return nil, fmt.Errorf("condition [not_contain] right value should be an array")
	}
	if len(val) == 0 {
		return nil, fmt.Errorf("condition [not_contain] right value should be an array of length >= 1")
	}

	return &NotContainCond{
		mCfg:   cfg,
		mField: field,
		mValue: val,
	}, nil
}
