// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"

	"vega-backend/interfaces"
)

type NullCond struct {
	mCfg   *CondCfg
	mField *interfaces.Property
}

// null 条件, 判断字段是否为空
func NewNullCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*interfaces.Property) (Condition, error) {
	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [null] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [null] left field '%s' not found", cfg.Name)
	}

	return &NullCond{
		mCfg:   cfg,
		mField: field,
	}, nil

}
