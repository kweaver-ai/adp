// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"errors"
	"fmt"

	"vega-backend/interfaces"
	dtype "vega-backend/interfaces/data_type"
	vopt "vega-backend/interfaces/value_opt"
)

var (
	CurrentYear   = "year"
	CurrentMonth  = "month"
	CurrentWeek   = "week"
	CurrentDay    = "day"
	CurrentHour   = "hour"
	CurrentMinute = "minute"

	CurrentFormatMap = map[string]bool{
		CurrentYear:   true,
		CurrentMonth:  true,
		CurrentWeek:   true,
		CurrentDay:    true,
		CurrentHour:   true,
		CurrentMinute: true,
	}
)

type CurrentCond struct {
	mCfg   *CondCfg
	mField *interfaces.Property
	mValue string
}

// 当前时间 current，判断字段是否为当前时间，时间格式为 "%Y-%m-%d %H:%i"
func NewCurrentCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*interfaces.Property) (Condition, error) {
	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [current] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [current] left field '%s' not found", cfg.Name)
	}
	if !dtype.DataType_IsDate(field.Type) {
		return nil, fmt.Errorf("condition [current] left field is not a date field: %s:%s", field.Name, field.Type)
	}

	if cfg.ValueOptCfg.ValueFrom != vopt.ValueFrom_Const {
		return nil, fmt.Errorf("condition [current] does not support value_from type '%s'", cfg.ValueFrom)
	}
	val, ok := cfg.ValueOptCfg.Value.(string)
	if !ok {
		return nil, fmt.Errorf("condition [current] right value should be string")
	}
	if _, ok := CurrentFormatMap[val]; !ok {
		return nil, errors.New(`condition [current] right value should be 
		one of [` + CurrentYear + `, ` + CurrentMonth + `, ` + CurrentWeek + `, ` + CurrentDay + `, ` + CurrentHour + `, ` + CurrentMinute + `], actual is ` + val)
	}

	return &CurrentCond{
		mCfg:   cfg,
		mField: field,
		mValue: val,
	}, nil
}
