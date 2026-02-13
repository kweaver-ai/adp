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

type BeforeCond struct {
	mCfg   *CondCfg
	mField *interfaces.Property
	mValue []any
}

// before 条件，判断字段是否在某个时间之前
func NewBeforeCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*interfaces.Property) (Condition, error) {
	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [before] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [before] left field '%s' not found", cfg.Name)
	}
	if !dtype.DataType_IsDate(field.Type) {
		return nil, fmt.Errorf("condition [before] left field is not a date field: %s:%s", cfg.Name, field.Type)
	}

	if cfg.ValueOptCfg.ValueFrom != vopt.ValueFrom_Const {
		return nil, fmt.Errorf("condition [before] does not support value_from type '%s'", cfg.ValueFrom)
	}
	val, ok := cfg.ValueOptCfg.Value.([]any)
	if !ok {
		return nil, fmt.Errorf("condition [before] right value should be an array")
	}
	if len(val) != 2 {
		return nil, fmt.Errorf("condition [before] right value should be an array of length 2")
	}
	if _, ok := val[0].(string); ok {
		return nil, fmt.Errorf("condition [before]'s interval value should be an number")
	}
	if _, ok = val[1].(string); !ok {
		return nil, fmt.Errorf("condition [before]'s interval value should be a string")
	}

	return &BeforeCond{
		mCfg:   cfg,
		mField: field,
		mValue: val,
	}, nil
}
