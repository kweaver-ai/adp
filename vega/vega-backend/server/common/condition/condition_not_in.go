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

type NotInCond struct {
	mCfg   *CondCfg
	mField *interfaces.Property
	mValue []any
}

// not_in 条件, 判断字段是否不在某个值数组中
func NewNotInCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*interfaces.Property) (Condition, error) {
	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [not_in] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [not_in] left field '%s' not found", cfg.Name)
	}

	if cfg.ValueOptCfg.ValueFrom != vopt.ValueFrom_Const {
		return nil, fmt.Errorf("condition [not_in] does not support value_from type '%s'", cfg.ValueFrom)
	}
	val, ok := cfg.ValueOptCfg.Value.([]any)
	if !ok {
		return nil, fmt.Errorf("condition [not_in] right value should be an array")
	}
	if len(val) == 0 {
		return nil, fmt.Errorf("condition [not_in] right value should be an array of length >= 1")
	}

	return &NotInCond{
		mCfg:   cfg,
		mField: field,
		mValue: val,
	}, nil
}
