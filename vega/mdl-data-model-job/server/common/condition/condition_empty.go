// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"
)

type EmptyCond struct {
	mCfg             *CondCfg
	mFilterFieldName string
}

// 空值 empty
func NewEmptyCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*Field) (Condition, error) {
	// 只允许字符串类型
	if !DataType_IsString(cfg.NameField.Type) {
		return nil, fmt.Errorf("condition [empty] left field %s is not of string type, but %s", cfg.Name, cfg.NameField.Type)
	}

	return &EmptyCond{
		mCfg:             cfg,
		mFilterFieldName: getFilterFieldName(cfg.Name, fieldsMap, false),
	}, nil

}

// 检查字段值是否为空字符串
func (cond *EmptyCond) Pass(ctx context.Context, data *OriginalData) (bool, error) {
	lv, err := data.GetSingleData(ctx, cond.mCfg.NameField)
	if err != nil {
		return false, err
	}
	if lv == nil {
		return false, nil
	}

	lvValue, ok := lv.(string)
	if !ok {
		return false, fmt.Errorf("condition [empty] left field %s value %v is not of string type", cond.mCfg.Name, lv)
	}

	return lvValue == "", nil
}
