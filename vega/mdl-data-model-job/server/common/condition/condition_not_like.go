// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"
	"strings"
)

type NotLikeCond struct {
	mCfg             *CondCfg
	mValue           string
	mFilterFieldName string
}

func NewNotLikeCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*Field) (Condition, error) {
	if !DataType_IsString(cfg.NameField.Type) {
		return nil, fmt.Errorf("condition [not_like] left field is not a string field: %s:%s", cfg.NameField.Name, cfg.NameField.Type)
	}

	if cfg.ValueOptCfg.ValueFrom != ValueFrom_Const {
		return nil, fmt.Errorf("condition [not_like] does not support value from type(%s)", cfg.ValueFrom)
	}

	val, ok := cfg.ValueOptCfg.Value.(string)
	if !ok {
		return nil, fmt.Errorf("condition [not_like] right value is not a string value: %v", cfg.Value)
	}

	return &NotLikeCond{
		mCfg:             cfg,
		mValue:           val,
		mFilterFieldName: getFilterFieldName(cfg.Name, fieldsMap, false),
	}, nil
}

func (cond *NotLikeCond) Pass(ctx context.Context, data *OriginalData) (bool, error) {
	lv, err := data.GetSingleData(ctx, cond.mCfg.NameField)
	if err != nil {
		return false, err
	}
	if lv == nil {
		return false, nil
	}

	rv := cond.mValue
	if rv == "" {
		return false, fmt.Errorf("condition [not_like] does not support empty pattern")
	}

	return !strings.Contains(lv.(string), rv), nil
}
