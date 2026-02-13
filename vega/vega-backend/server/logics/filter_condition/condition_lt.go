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

type LtCond struct {
	mCfg   *interfaces.FilterCondCfg
	mField *interfaces.Property
	mValue any
}

func (c *LtCond) GetOperation() string { return OperationLt }

func (c *LtCond) SupportSubCond() bool       { return false }
func (c *LtCond) NeedName() bool             { return true }
func (c *LtCond) NeedValue() bool            { return true }
func (c *LtCond) NeedConstValue() bool       { return true }
func (c *LtCond) IsSingleValue() bool        { return true }
func (c *LtCond) IsFixedLenArrayValue() bool { return false }
func (c *LtCond) RequiredValueLen() int      { return -1 }

// lt 条件, 判断字段是否小于某个值
func (c *LtCond) New(ctx context.Context, cfg *interfaces.FilterCondCfg,
	fieldsMap map[string]*interfaces.Property) (interfaces.FilterCondition, error) {

	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [lt] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [lt] left field '%s' not found", cfg.Name)
	}

	if cfg.ValueOptCfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [lt] does not support value_from type '%s'", cfg.ValueFrom)
	}
	if IsSlice(cfg.ValueOptCfg.Value) {
		return nil, fmt.Errorf("condition [lt] only supports single value")
	}

	return &LtCond{
		mCfg:   cfg,
		mField: field,
		mValue: cfg.ValueOptCfg.Value,
	}, nil
}
