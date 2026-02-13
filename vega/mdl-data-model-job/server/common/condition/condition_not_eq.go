// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"
)

type NotEqCond struct {
	mCfg             *CondCfg
	mFilterFieldName string
}

func NewNotEqCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*Field) (Condition, error) {
	if cfg.ValueOptCfg.ValueFrom != ValueFrom_Const {
		return nil, fmt.Errorf("condition [not_eq] does not support value from type(%s)", cfg.ValueFrom)
	}

	if IsSlice(cfg.ValueOptCfg.Value) {
		return nil, fmt.Errorf("condition [not_eq] only supports single value")
	}

	return &NotEqCond{
		mCfg:             cfg,
		mFilterFieldName: getFilterFieldName(cfg.Name, fieldsMap, false),
	}, nil

}

// 不等于 ne（!=），右侧值为单个值
func (cond *NotEqCond) Pass(ctx context.Context, data *OriginalData) (bool, error) {
	return compare(ctx, data, cond.mCfg)
}
