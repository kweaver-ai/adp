// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package filter_condition

import (
	"context"
	"fmt"

	"vega-backend/interfaces"
)

type GteCond struct {
	mCfg   *interfaces.FilterCondCfg
	mField *interfaces.Property
	mValue any
}

func (c *GteCond) GetOperation() string { return OperationGte }

func (c *GteCond) SupportSubCond() bool       { return false }
func (c *GteCond) NeedName() bool             { return true }
func (c *GteCond) NeedValue() bool            { return true }
func (c *GteCond) NeedConstValue() bool       { return true }
func (c *GteCond) IsSingleValue() bool        { return true }
func (c *GteCond) IsFixedLenArrayValue() bool { return false }
func (c *GteCond) RequiredValueLen() int      { return -1 }

// gte 条件, 判断字段是否大于等于某个值
func (c *GteCond) New(ctx context.Context, cfg *interfaces.FilterCondCfg,
	fieldsMap map[string]*interfaces.Property) (interfaces.FilterCondition, error) {

	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [gte] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [gte] left field '%s' not found", cfg.Name)
	}

	if cfg.ValueOptCfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [gte] does not support value_from type '%s'", cfg.ValueFrom)
	}
	if IsSlice(cfg.ValueOptCfg.Value) {
		return nil, fmt.Errorf("condition [gte] only supports single value")
	}

	return &GteCond{
		mCfg:   cfg,
		mField: field,
		mValue: cfg.ValueOptCfg.Value,
	}, nil
}
