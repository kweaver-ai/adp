// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"
)

type NotInCond struct {
	mCfg             *CondCfg
	mValue           []any
	mFilterFieldName string
}

func NewNotInCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*Field) (Condition, error) {
	if cfg.ValueOptCfg.ValueFrom != ValueFrom_Const {
		return nil, fmt.Errorf("condition [not_in] does not support value from type(%s)", cfg.ValueFrom)
	}

	if !IsSlice(cfg.ValueOptCfg.Value) {
		return nil, fmt.Errorf("condition [not_in] right value should be an array")
	}

	if !IsSameType(cfg.ValueOptCfg.Value.([]any)) {
		return nil, fmt.Errorf("condition [not_in] right value should be an array composed of elements of same type")
	}

	return &NotInCond{
		mCfg:             cfg,
		mValue:           cfg.ValueOptCfg.Value.([]any),
		mFilterFieldName: getFilterFieldName(cfg.Name, fieldsMap, false),
	}, nil
}

func (cond *NotInCond) Pass(ctx context.Context, data *OriginalData) (bool, error) {
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
			return false, nil
		}
	}

	return true, nil
}
