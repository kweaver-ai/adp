// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"
	"os"

	dtype "ontology-query/interfaces/data_type"
)

type BeforeCond struct {
	mCfg             *CondCfg
	mValue           []any
	mFilterFieldName string
}

func NewBeforeCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*DataProperty) (Condition, error) {
	// 检查是否为日期/时间类型
	simpleType := dtype.SimpleTypeMapping[cfg.NameField.Type]
	if simpleType != dtype.SimpleDate && simpleType != dtype.SimpleDatetime && simpleType != dtype.SimpleTime {
		return nil, fmt.Errorf("condition [before] left field is not a date/time field: %s:%s", cfg.NameField.Name, cfg.NameField.Type)
	}

	if cfg.ValueOptCfg.ValueFrom != ValueFrom_Const {
		return nil, fmt.Errorf("condition [before] does not support value_from type '%s'", cfg.ValueOptCfg.ValueFrom)
	}

	val, ok := cfg.ValueOptCfg.Value.([]any)
	if !ok {
		return nil, fmt.Errorf("condition [before] right value should be an array of length 2")
	}

	if len(val) != 2 {
		return nil, fmt.Errorf("condition [before] right value should be an array of length 2")
	}

	// 第一个值应该是数字
	if _, ok := val[0].(float64); !ok {
		if _, ok := val[0].(int); !ok {
			if _, ok := val[0].(int64); !ok {
				return nil, fmt.Errorf("condition [before]'s interval value should be a number")
			}
		}
	}

	// 第二个值应该是字符串（unit）
	_, ok = val[1].(string)
	if !ok {
		return nil, fmt.Errorf("condition [before]'s unit value should be a string")
	}

	return &BeforeCond{
		mCfg:             cfg,
		mValue:           val,
		mFilterFieldName: getFilterFieldName(cfg.Name, fieldsMap, false),
	}, nil
}

func (cond *BeforeCond) Convert(ctx context.Context, vectorizer func(ctx context.Context, property *DataProperty, word string) ([]VectorResp, error)) (string, error) {
	// before 操作符主要用于 SQL，OpenSearch DSL 暂不实现
	return "", nil
}

func (cond *BeforeCond) Convert2SQL(ctx context.Context) (string, error) {
	unit := cond.mValue[1].(string)

	// 获取时区，默认为 UTC
	tz := os.Getenv("TZ")
	if tz == "" {
		tz = "UTC"
	}

	// 获取数值
	var numValue interface{}
	switch v := cond.mValue[0].(type) {
	case float64:
		numValue = v
	case int:
		numValue = float64(v)
	case int64:
		numValue = float64(v)
	default:
		return "", fmt.Errorf("condition [before] invalid number value: %v", cond.mValue[0])
	}

	sqlStr := fmt.Sprintf(`"%s" >= DATE_add('%s', -%v, CURRENT_TIMESTAMP AT TIME ZONE 'UTC' AT TIME ZONE '%s') 
		AND "%s" <= CURRENT_TIMESTAMP AT TIME ZONE 'UTC' AT TIME ZONE '%s'`,
		cond.mFilterFieldName, unit, numValue, tz, cond.mFilterFieldName, tz)
	return sqlStr, nil
}

func rewriteBeforeCond(cfg *CondCfg) (*CondCfg, error) {
	// 过滤条件中的属性字段换成映射的视图字段
	if cfg.NameField.Name == "" {
		return nil, fmt.Errorf("过去[before]操作符使用的过滤字段[%s]在对象类的属性中不存在", cfg.Name)
	}
	return &CondCfg{
		Name:        cfg.NameField.MappedField.Name,
		Operation:   cfg.Operation,
		ValueOptCfg: cfg.ValueOptCfg,
	}, nil
}
