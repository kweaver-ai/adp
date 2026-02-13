// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"
)

type ContainCond struct {
	mCfg             *CondCfg
	IsSliceValue     bool
	mValue           any
	mSliceValue      []any
	mFilterFieldName string
}

func NewContainCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*Field) (Condition, error) {
	if cfg.ValueFrom != ValueFrom_Const {
		return nil, fmt.Errorf("condition [contain] does not support value from type(%s)", cfg.ValueFrom)
	}

	containCond := &ContainCond{
		mCfg:             cfg,
		mFilterFieldName: getFilterFieldName(cfg.Name, fieldsMap, false),
	}

	if IsSlice(cfg.Value) {
		if len(cfg.Value.([]any)) == 0 {
			return nil, fmt.Errorf("condition [contain] right value is an empty array")
		}

		containCond.IsSliceValue = true
		containCond.mSliceValue = cfg.Value.([]any)

	} else {
		containCond.IsSliceValue = false
		containCond.mValue = cfg.Value
	}

	return containCond, nil

}

// 包含 contain，左侧属性值为数组，右侧值为单个值或数组，如果为数组，意味着数组内的值都应在属性值内
func (cond *ContainCond) Pass(ctx context.Context, data *OriginalData) (bool, error) {
	leftValues, err := data.GetData(ctx, cond.mCfg.NameField)
	if err != nil {
		return false, err
	}

	// 原始数据中没有拿到这个字段的值，返回false
	if len(leftValues) == 0 {
		return false, nil
	}

	var rightValues []any
	if cond.IsSliceValue {
		rightValues = cond.mSliceValue
	} else {
		rightValues = []any{cond.mValue}
	}

	for _, rv := range rightValues {
		rslt := cond.CheckContain(leftValues, rv)
		if !rslt {
			return false, nil
		}
	}
	return true, nil

}

func (cond *ContainCond) CheckContain(leftValues []any, rv any) bool {
	for _, lv := range leftValues {
		if rv == lv {
			return true
		}
	}
	return false
}
