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

type GteCond struct {
	mCfg   *CondCfg
	mField *interfaces.Property
	mValue any
}

// gte 条件, 判断字段是否大于等于某个值
func NewGteCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*interfaces.Property) (Condition, error) {
	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [gte] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [gte] left field '%s' not found", cfg.Name)
	}

	if cfg.ValueOptCfg.ValueFrom != vopt.ValueFrom_Const {
		return nil, fmt.Errorf("condition [gte] does not support value_from type '%s'", cfg.ValueFrom)
	}
	if IsSlice(cfg.ValueOptCfg.Value) {
		return nil, fmt.Errorf("condition [gte] only supports single value")
	}

	return &GteCond{
		mCfg:   cfg,
		mField: field,
		mValue: cfg.ValueOptCfg.Value,
	}, nil
}
