// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"

	"vega-backend/interfaces"
	dtype "vega-backend/interfaces/data_type"
	vopt "vega-backend/interfaces/value_opt"
)

type BetweenCond struct {
	mCfg   *CondCfg
	mField *interfaces.Property
	mValue []any
}

// between 条件，判断字段是否在某个区间内, 区间包含左右边界
func NewBetweenCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*interfaces.Property) (Condition, error) {
	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [between] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [between] left field '%s' not found", cfg.Name)
	}
	if !dtype.DataType_IsDate(field.Type) && !dtype.DataType_IsNumber(field.Type) {
		return nil, fmt.Errorf("condition [between] left field is not a date or number field: %s:%s", cfg.Name, field.Type)
	}

	if cfg.ValueOptCfg.ValueFrom != vopt.ValueFrom_Const {
		return nil, fmt.Errorf("condition [between] does not support value_from type '%s'", cfg.ValueFrom)
	}
	val, ok := cfg.ValueOptCfg.Value.([]any)
	if !ok {
		return nil, fmt.Errorf("condition [between] right value should be an array")
	}
	if len(val) != 2 {
		return nil, fmt.Errorf("condition [between] right value should be an array of length 2")
	}

	return &BetweenCond{
		mCfg:   cfg,
		mField: field,
		mValue: val,
	}, nil
}
