// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"
	"time"
)

type RangeCond struct {
	mCfg             *CondCfg
	mValue           []any
	mFilterFieldName string
}

func NewRangeCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*Field) (Condition, error) {
	if cfg.ValueOptCfg.ValueFrom != ValueFrom_Const {
		return nil, fmt.Errorf("condition [range] does not support value from type(%s)", cfg.ValueFrom)
	}

	val, ok := cfg.ValueOptCfg.Value.([]any)
	if !ok {
		return nil, fmt.Errorf("condition [range] right value should be an array of length 2")
	}

	if len(val) != 2 {
		return nil, fmt.Errorf("condition [range] right value should be an array of length 2")
	}

	return &RangeCond{
		mCfg:             cfg,
		mValue:           val,
		mFilterFieldName: getFilterFieldName(cfg.Name, fieldsMap, false),
	}, nil
}

// 范围内 range，右侧值为长度为2的数组，边界为左闭右开，即 [ value[0], value[1] )
func (cond *RangeCond) Pass(ctx context.Context, data *OriginalData) (bool, error) {
	lv, err := data.GetSingleData(ctx, cond.mCfg.NameField)
	if err != nil {
		return false, err
	}
	if lv == nil {
		return false, nil
	}

	rv := cond.mValue
	if len(rv) != 2 {
		return false, fmt.Errorf("condition [range] only support two value range: %v", rv)
	}
	rv0, err := time.Parse(time.RFC3339Nano, rv[0].(string))
	if err != nil {
		return false, err
	}
	rv1, err := time.Parse(time.RFC3339Nano, rv[1].(string))
	if err != nil {
		return false, err
	}

	switch cond.mCfg.NameField.Type {
	case DataType_Byte, DataType_Short, DataType_Integer, DataType_Long:
		return lv.(int64) >= rv[0].(int64) && lv.(int64) < rv[1].(int64), nil

	case DataType_HalfFloat, DataType_Float, DataType_Double:
		return lv.(float64) >= rv[0].(float64) && lv.(float64) < rv[1].(float64), nil
	case DataType_Date:
		return lv.(time.Time).Sub(rv0) >= 0 && lv.(time.Time).Before(rv1), nil

	default:
		return false, nil
	}
}
