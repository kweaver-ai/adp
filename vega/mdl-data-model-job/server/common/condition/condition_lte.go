// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"
)

type LteCond struct {
	mCfg             *CondCfg
	mFilterFieldName string
}

func NewLteCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*Field) (Condition, error) {
	if cfg.ValueOptCfg.ValueFrom != ValueFrom_Const {
		return nil, fmt.Errorf("condition [lte] does nor support value from type(%s)", cfg.ValueFrom)
	}

	if IsSlice(cfg.ValueOptCfg.Value) {
		return nil, fmt.Errorf("condition [lte] only supports single value")
	}

	return &LteCond{
		mCfg:             cfg,
		mFilterFieldName: getFilterFieldName(cfg.Name, fieldsMap, false),
	}, nil

}

func (cond *LteCond) Pass(ctx context.Context, data *OriginalData) (bool, error) {
	return compare(ctx, data, cond.mCfg)
}
