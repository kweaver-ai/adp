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

type NotLikeCond struct {
	mCfg   *CondCfg
	mField *interfaces.Property
	mValue string
}

func NewNotLikeCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*interfaces.Property) (Condition, error) {
	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [not_like] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [not_like] left field '%s' not found", cfg.Name)
	}
	if dtype.DataType_IsNumber(field.Type) {
		return nil, fmt.Errorf("condition [not_like] left field is not a string/text field: %s:%s", cfg.Name, field.Type)
	}

	if cfg.ValueOptCfg.ValueFrom != vopt.ValueFrom_Const {
		return nil, fmt.Errorf("condition [not_like] does not support value_from type '%s'", cfg.ValueFrom)
	}
	val, ok := cfg.ValueOptCfg.Value.(string)
	if !ok {
		return nil, fmt.Errorf("condition [not_like] right value is not a string value: %v", cfg.Value)
	}

	return &NotLikeCond{
		mCfg:   cfg,
		mField: field,
		mValue: val,
	}, nil
}
