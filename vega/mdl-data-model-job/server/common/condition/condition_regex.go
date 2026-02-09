// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"

	"github.com/dlclark/regexp2"
)

type RegexCond struct {
	mCfg             *CondCfg
	mValue           string
	mRegexp          *regexp2.Regexp
	mFilterFieldName string
}

func NewRegexCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*Field) (Condition, error) {
	if !DataType_IsString(cfg.NameField.Type) {
		return nil, fmt.Errorf("condition [regex] left field is not a string field: %s:%s", cfg.NameField.Name, cfg.NameField.Type)
	}

	if cfg.ValueOptCfg.ValueFrom != ValueFrom_Const {
		return nil, fmt.Errorf("condition [regex] does not support value from type(%s)", cfg.ValueFrom)
	}

	val, ok := cfg.ValueOptCfg.Value.(string)
	if !ok {
		return nil, fmt.Errorf("condition [regex] right value is not a string value: %v", cfg.Value)
	}

	regexp, err := regexp2.Compile(val, regexp2.RE2)
	if err != nil {
		return nil, fmt.Errorf("regular expression error: %v", err)
	}

	return &RegexCond{
		mCfg:             cfg,
		mValue:           val,
		mRegexp:          regexp,
		mFilterFieldName: getFilterFieldName(cfg.Name, fieldsMap, false),
	}, nil
}

func (cond *RegexCond) Pass(ctx context.Context, data *OriginalData) (bool, error) {
	lv, err := data.GetData(ctx, cond.mCfg.NameField)
	if err != nil {
		return false, err
	}

	// 原始数据中没有拿到这个字段的值，返回false
	if len(lv) == 0 {
		return false, nil
	}

	match, err := cond.mRegexp.MatchString(lv[0].(string))
	if err != nil {
		return false, err
	}
	return match, nil
}
