// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"

	"vega-backend/interfaces"
	vopt "vega-backend/interfaces/value_opt"
)

type MatchCond struct {
	mCfg    *CondCfg
	mFields []*interfaces.Property
}

// match 条件, 判断字段是否匹配某个字符串
// 支持全部字段 *
func NewMatchCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*interfaces.Property) (Condition, error) {
	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [match] left field is empty")
	}
	mFields := make([]*interfaces.Property, 0)
	if cfg.Name == AllField {
		for fieldName := range fieldsMap {
			mFields = append(mFields, fieldsMap[fieldName])
		}
	} else {
		field, ok := fieldsMap[cfg.Name]
		if !ok {
			return nil, fmt.Errorf("condition [match] left field '%s' not found", cfg.Name)
		}
		mFields = append(mFields, field)
	}

	if cfg.ValueOptCfg.ValueFrom != vopt.ValueFrom_Const {
		return nil, fmt.Errorf("condition [match] does not support value_from type '%s'", cfg.ValueFrom)
	}

	return &MatchCond{
		mCfg:    cfg,
		mFields: mFields,
	}, nil
}
