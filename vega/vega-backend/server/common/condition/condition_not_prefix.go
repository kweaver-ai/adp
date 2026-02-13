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

type NotPrefixCond struct {
	mCfg   *CondCfg
	mField *interfaces.Property
	mValue string
}

// not_prefix 条件, 判断字段是否不以某个前缀开头
func NewNotPrefixCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*interfaces.Property) (Condition, error) {
	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [not_prefix] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [not_prefix] left field '%s' not found", cfg.Name)
	}
	if !dtype.DataType_IsString(field.Type) {
		return nil, fmt.Errorf("condition [not_prefix] left field '%s' is not a string/text field", cfg.Name)
	}

	if cfg.ValueOptCfg.ValueFrom != vopt.ValueFrom_Const {
		return nil, fmt.Errorf("condition [not_prefix] does not support value_from type '%s'", cfg.ValueFrom)
	}
	val, ok := cfg.ValueOptCfg.Value.(string)
	if !ok {
		return nil, fmt.Errorf("condition [not_prefix] right value is not a string value: %v", cfg.Value)
	}

	return &NotPrefixCond{
		mCfg:   cfg,
		mField: field,
		mValue: val,
	}, nil
}
