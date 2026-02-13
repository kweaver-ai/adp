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

var (
	// match_type
	MatchTypeMap = map[string]bool{
		"best_fields":   true,
		"most_fields":   true,
		"cross_fields":  true,
		"phrase":        true,
		"phrase_prefix": true,
		"bool_prefix":   true,
	}
)

type MultiMatchCond struct {
	mCfg       *CondCfg
	mFields    []*interfaces.Property
	mMatchType string
}

// multi_match 条件, 判断多个字段是否匹配某个规则
// 支持全部字段 *
func NewMultiMatchCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*interfaces.Property) (Condition, error) {

	// 从cfg的 ReaminCfg 中获取 fields，这是属于 multi_match的fields字段，是个字符串数组，
	// 如果想要全部字段匹配，可不填或者填 ["*"], 不支持填字符串 *， 需要一个数组
	var mFields []*interfaces.Property
	cfgFields, ok := cfg.RemainCfg["fields"].([]any)
	if !ok {
		return nil, fmt.Errorf("condition [multi_match] 'fields' value should be an array")
	}

	if len(cfgFields) == 1 && cfgFields[0].(string) == AllField {
		mFields = make([]*interfaces.Property, 0, len(fieldsMap))
		for _, field := range fieldsMap {
			mFields = append(mFields, field)
		}
	} else {
		// 字段数组里的需要是个字符串数组
		for _, cfgField := range cfgFields {
			fieldName, ok := cfgField.(string)
			if !ok {
				return nil, fmt.Errorf("condition [multi_match] 'fields' value should be a field name array, contain non string value[%v]", cfgField)
			}
			field, ok := fieldsMap[fieldName]
			if !ok {
				return nil, fmt.Errorf("condition [multi_match] 'fields' exists any field not exists in resource [%s]", fieldName)
			}
			mFields = append(mFields, field)
		}
	}

	// 校验match_type的有效性, match_type可以为空
	matchType, ok := cfg.RemainCfg["match_type"].(string)
	if !ok {
		return nil, fmt.Errorf("condition [multi_match] 'match_type' value should be a string")
	}
	if !MatchTypeMap[matchType] {
		return nil, fmt.Errorf("condition [multi_match] 'match_type' value should be one of [%v], actual is[%v]", MatchTypeMap, matchType)
	}

	return &MultiMatchCond{
		mCfg:       cfg,
		mFields:    mFields,
		mMatchType: matchType,
	}, nil
}
