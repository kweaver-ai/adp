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

type NotEmptyCond struct {
	mCfg   *CondCfg
	mField *interfaces.Property
}

// not_empty 条件，判断字段是否不为空字符串
func NewNotEmptyCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*interfaces.Property) (Condition, error) {
	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [not_empty] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [not_empty] left field '%s' not found", cfg.Name)
	}

	// 只允许字符串类型
	if !dtype.DataType_IsString(field.Type) {
		return nil, fmt.Errorf("condition [not_empty] left field %s is not of string type, but %s", cfg.Name, field.Type)
	}

	return &NotEmptyCond{
		mCfg:   cfg,
		mField: field,
	}, nil

}
