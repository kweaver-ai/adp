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

type ExistCond struct {
	mCfg   *CondCfg
	mfield *interfaces.Property
}

// 存在 exist，判断字段是否存在
func NewExistCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*interfaces.Property) (Condition, error) {
	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [eq] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [eq] left field '%s' not found", cfg.Name)
	}

	return &ExistCond{
		mCfg:   cfg,
		mfield: field,
	}, nil
}
