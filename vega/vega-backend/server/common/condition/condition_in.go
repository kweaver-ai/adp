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

type InCond struct {
	mCfg   *CondCfg
	mField *interfaces.Property
	mValue []any
}

// in 条件, 判断字段是否在某个数组中
func NewInCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*interfaces.Property) (Condition, error) {
	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [in] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [in] left field '%s' not found", cfg.Name)
	}

	if cfg.ValueOptCfg.ValueFrom != vopt.ValueFrom_Const {
		return nil, fmt.Errorf("condition [in] does not support value_from type '%s'", cfg.ValueFrom)
	}
	val, ok := cfg.ValueOptCfg.Value.([]any)
	if !ok {
		return nil, fmt.Errorf("condition [in] right value should be an array")
	}
	if len(val) == 0 {
		return nil, fmt.Errorf("condition [in] right value should be an array of length >= 1")
	}

	return &InCond{
		mCfg:   cfg,
		mField: field,
		mValue: val,
	}, nil
}
