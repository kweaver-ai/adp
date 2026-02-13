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

type GtCond struct {
	mCfg   *interfaces.FilterCondCfg
	mField *interfaces.Property
	mValue any
}

func (c *GtCond) GetOperation() string { return OperationGt }

func (c *GtCond) SupportSubCond() bool       { return false }
func (c *GtCond) NeedName() bool             { return true }
func (c *GtCond) NeedValue() bool            { return true }
func (c *GtCond) NeedConstValue() bool       { return true }
func (c *GtCond) IsSingleValue() bool        { return true }
func (c *GtCond) IsFixedLenArrayValue() bool { return false }
func (c *GtCond) RequiredValueLen() int      { return -1 }

// gt 条件, 判断字段是否大于某个值
func (c *GtCond) New(ctx context.Context, cfg *interfaces.FilterCondCfg,
	fieldsMap map[string]*interfaces.Property) (interfaces.FilterCondition, error) {

	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [gt] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [gt] left field '%s' not found", cfg.Name)
	}

	if cfg.ValueOptCfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [gt] does not support value_from type '%s'", cfg.ValueFrom)
	}
	if IsSlice(cfg.ValueOptCfg.Value) {
		return nil, fmt.Errorf("condition [gt] only supports single value")
	}

	return &GtCond{
		mCfg:   cfg,
		mField: field,
		mValue: cfg.ValueOptCfg.Value,
	}, nil
}
