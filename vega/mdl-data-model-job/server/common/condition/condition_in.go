// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"
)

type InCond struct {
	mCfg             *CondCfg
	mValue           []any
	mFilterFieldName string
}

// 属于 in，右侧值为一个或多个相同类型值组成的数组
func NewInCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*Field) (Condition, error) {
	if cfg.ValueOptCfg.ValueFrom != ValueFrom_Const {
		return nil, fmt.Errorf("condition [in] does not support value from type(%s)", cfg.ValueFrom)
	}

	if !IsSlice(cfg.ValueOptCfg.Value) {
		return nil, fmt.Errorf("condition [in] right value should be an array")
	}

	if !IsSameType(cfg.ValueOptCfg.Value.([]any)) {
		return nil, fmt.Errorf("condition [in] right value should be an array composed of elements of same type")
	}

	return &InCond{
		mCfg:             cfg,
		mValue:           cfg.ValueOptCfg.Value.([]any),
		mFilterFieldName: getFilterFieldName(cfg.Name, fieldsMap, false),
	}, nil
}

func (cond *InCond) Pass(ctx context.Context, data *OriginalData) (bool, error) {
	lv, err := data.GetSingleData(ctx, cond.mCfg.NameField)
	if err != nil {
		return false, err
	}
	if lv == nil {
		return false, nil
	}

	rv := cond.mValue
	if len(rv) == 0 {
		return false, nil
	}

	for _, rrv := range rv {
		if lv == rrv {
			return true, nil
		}
	}

	return false, nil
}
