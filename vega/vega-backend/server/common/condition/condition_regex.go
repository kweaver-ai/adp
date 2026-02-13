// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"

	"github.com/dlclark/regexp2"

	"vega-backend/interfaces"
	dtype "vega-backend/interfaces/data_type"
	vopt "vega-backend/interfaces/value_opt"
)

type RegexCond struct {
	mCfg    *CondCfg
	mField  *interfaces.Property
	mValue  string
	mRegexp *regexp2.Regexp
}

// regex 条件, 判断字段是否匹配某个正则表达式
func NewRegexCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*interfaces.Property) (Condition, error) {
	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [regex] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [regex] left field '%s' not found", cfg.Name)
	}
	if !dtype.DataType_IsString(field.Type) {
		return nil, fmt.Errorf("condition [regex] left field '%s' type must be string", cfg.Name)
	}

	if cfg.ValueOptCfg.ValueFrom != vopt.ValueFrom_Const {
		return nil, fmt.Errorf("condition [regex] does not support value_from type '%s'", cfg.ValueFrom)
	}
	val, ok := cfg.ValueOptCfg.Value.(string)
	if !ok {
		return nil, fmt.Errorf("condition [regex] right value is not a string value: %v", cfg.Value)
	}
	regexp, err := regexp2.Compile(val, regexp2.RE2)
	if err != nil {
		return nil, fmt.Errorf("condition [regex] regular expression error: %s", err.Error())
	}

	return &RegexCond{
		mCfg:    cfg,
		mField:  field,
		mValue:  val,
		mRegexp: regexp,
	}, nil
}
