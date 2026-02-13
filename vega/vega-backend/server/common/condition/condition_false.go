// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"

	"vega-backend/interfaces"
	dtype "vega-backend/interfaces/data_type"
)

type FalseCond struct {
	mCfg   *CondCfg
	mField *interfaces.Property
}

// false 条件，判断字段是否为 false
func NewFalseCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*interfaces.Property) (Condition, error) {
	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [false] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [false] left field '%s' not found", cfg.Name)
	}
	if field.Type != dtype.DataType_Boolean {
		return nil, fmt.Errorf("condition [false] left field is not a boolean field: %s:%s", cfg.Name, field.Type)
	}

	return &FalseCond{
		mCfg:   cfg,
		mField: field,
	}, nil
}
