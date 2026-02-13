// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"
)

type NotContainCond struct {
	mCfg             *CondCfg
	IsSliceValue     bool
	mValue           any
	mSliceValue      []any
	mFilterFieldName string
}

func NewNotContainCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*Field) (Condition, error) {
	if cfg.ValueFrom != ValueFrom_Const {
		return nil, fmt.Errorf("condition [not_contain] does not support value from type(%s)", cfg.ValueFrom)
	}

	notContainCond := &NotContainCond{
		mCfg:             cfg,
		mFilterFieldName: getFilterFieldName(cfg.Name, fieldsMap, false),
	}

	if IsSlice(cfg.Value) {
		if len(cfg.Value.([]any)) == 0 {
			return nil, fmt.Errorf("condition [not_contain] right value is an empty array")
		}

		notContainCond.IsSliceValue = true
		notContainCond.mSliceValue = cfg.Value.([]any)

	} else {
		notContainCond.IsSliceValue = false
		notContainCond.mValue = cfg.Value
	}

	return notContainCond, nil

}

// 不包含 not_contain，左侧属性值为数组，右侧值为单个值或数组，如果为数组，意味着数组内的值都应在属性值外
func (cond *NotContainCond) Pass(ctx context.Context, data *OriginalData) (bool, error) {
	lv, err := data.GetData(ctx, cond.mCfg.NameField)
	if err != nil {
		return false, err
	}

	// 原始数据中没有拿到这个字段的值，返回true
	if len(lv) == 0 {
		return true, nil
	}

	var rv []any
	if cond.IsSliceValue {
		rv = cond.mSliceValue
	} else {
		rv = []any{cond.mValue}
	}

	for _, rrv := range rv {
		rslt := cond.CheckNotContain(lv, rrv)
		if !rslt {
			return false, nil
		}
	}
	return true, nil

}

func (cond *NotContainCond) CheckNotContain(lv []any, rv any) bool {
	for _, rlv := range lv {
		if rlv == rv {
			return false
		}
	}
	return true
}
