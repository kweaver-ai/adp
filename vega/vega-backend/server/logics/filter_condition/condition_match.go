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

type MatchCond struct {
	mCfg    *interfaces.FilterCondCfg
	mFields []*interfaces.Property
}

func (c *MatchCond) GetOperation() string { return OperationMatch }

func (c *MatchCond) SupportSubCond() bool       { return false }
func (c *MatchCond) NeedName() bool             { return true }
func (c *MatchCond) NeedValue() bool            { return true }
func (c *MatchCond) NeedConstValue() bool       { return true }
func (c *MatchCond) IsSingleValue() bool        { return true }
func (c *MatchCond) IsFixedLenArrayValue() bool { return false }
func (c *MatchCond) RequiredValueLen() int      { return -1 }

// match 条件, 判断字段是否匹配某个字符串
// 支持全部字段 *
func (c *MatchCond) New(ctx context.Context, cfg *interfaces.FilterCondCfg,
	fieldsMap map[string]*interfaces.Property) (interfaces.FilterCondition, error) {

	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [match] left field is empty")
	}
	mFields := make([]*interfaces.Property, 0)
	if cfg.Name == interfaces.AllField {
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

	if cfg.ValueOptCfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [match] does not support value_from type '%s'", cfg.ValueFrom)
	}

	return &MatchCond{
		mCfg:    cfg,
		mFields: mFields,
	}, nil
}
