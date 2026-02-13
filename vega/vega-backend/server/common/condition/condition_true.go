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

type TrueCond struct {
	mCfg   *CondCfg
	mField *interfaces.Property
}

// bool 类型为真
func NewTrueCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*interfaces.Property) (Condition, error) {
	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [true] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [true] left field '%s' not found", cfg.Name)
	}
	if field.Type != dtype.DataType_Boolean {
		return nil, fmt.Errorf("condition [true] left field is not a boolean field: %s:%s", cfg.Name, field.Type)
	}

	return &TrueCond{
		mCfg:   cfg,
		mField: field,
	}, nil
}
