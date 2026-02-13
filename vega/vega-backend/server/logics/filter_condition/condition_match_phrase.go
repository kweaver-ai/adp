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

type MatchPhraseCond struct {
	mCfg    *interfaces.FilterCondCfg
	mFields []*interfaces.Property
}

func (c *MatchPhraseCond) GetOperation() string { return OperationMatchPhrase }

func (c *MatchPhraseCond) SupportSubCond() bool       { return false }
func (c *MatchPhraseCond) NeedName() bool             { return true }
func (c *MatchPhraseCond) NeedValue() bool            { return true }
func (c *MatchPhraseCond) NeedConstValue() bool       { return true }
func (c *MatchPhraseCond) IsSingleValue() bool        { return true }
func (c *MatchPhraseCond) IsFixedLenArrayValue() bool { return false }
func (c *MatchPhraseCond) RequiredValueLen() int      { return -1 }

// match_phrase 条件, 判断字段是否匹配某个短语
// 支持全部字段 *
func (c *MatchPhraseCond) New(ctx context.Context, cfg *interfaces.FilterCondCfg,
	fieldsMap map[string]*interfaces.Property) (interfaces.FilterCondition, error) {

	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [match_phrase] left field is empty")
	}
	mFields := make([]*interfaces.Property, 0)
	if cfg.Name == interfaces.AllField {
		for fieldName := range fieldsMap {
			mFields = append(mFields, fieldsMap[fieldName])
		}
	} else {
		field, ok := fieldsMap[cfg.Name]
		if !ok {
			return nil, fmt.Errorf("condition [match_phrase] left field '%s' not found", cfg.Name)
		}
		mFields = append(mFields, field)
	}

	if cfg.ValueOptCfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [match_phrase] does not support value_from type '%s'", cfg.ValueFrom)
	}

	return &MatchPhraseCond{
		mCfg:    cfg,
		mFields: mFields,
	}, nil
}
