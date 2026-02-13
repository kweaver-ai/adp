// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"
)

type NotEmptyCond struct {
	mCfg             *CondCfg
	mFilterFieldName string
}

// 非空 not_empty
func NewNotEmptyCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*Field) (Condition, error) {
	// 只允许字符串类型
	if !DataType_IsString(cfg.NameField.Type) {
		return nil, fmt.Errorf("condition [not_empty] left field %s is not of string type, but %s", cfg.Name, cfg.NameField.Type)
	}

	return &NotEmptyCond{
		mCfg:             cfg,
		mFilterFieldName: getFilterFieldName(cfg.Name, fieldsMap, false),
	}, nil

}

// 字段存在且不能为空字符串
func (cond *NotEmptyCond) Pass(ctx context.Context, data *OriginalData) (bool, error) {
	lv, err := data.GetSingleData(ctx, cond.mCfg.NameField)
	if err != nil {
		return false, err
	}
	if lv == nil {
		return false, nil
	}

	lvValue, ok := lv.(string)
	if !ok {
		return false, fmt.Errorf("condition [not_empty] left field %s value %v is not of string type", cond.mCfg.Name, lv)
	}

	return lvValue != "", nil
}
