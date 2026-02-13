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

type NotExistCond struct {
	mCfg   *CondCfg
	mfield *interfaces.Property
}

func NewNotExistCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*interfaces.Property) (Condition, error) {
	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [not_exist] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [not_exist] left field '%s' not found", cfg.Name)
	}

	return &NotExistCond{
		mCfg:   cfg,
		mfield: field,
	}, nil
}
