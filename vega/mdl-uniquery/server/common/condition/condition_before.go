// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"
	"os"

	vopt "uniquery/common/value_opt"
	dtype "uniquery/interfaces/data_type"
)

type BeforeCond struct {
	mCfg             *CondCfg
	mValue           any
	mUnit            string
	mFilterFieldName string
}

func NewBeforeCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*ViewField) (Condition, error) {
	if !dtype.DataType_IsDate(cfg.NameField.Type) {
		return nil, fmt.Errorf("condition [before] left field is not a date field: %s:%s", cfg.NameField.Name, cfg.NameField.Type)
	}

	if cfg.ValueOptCfg.ValueFrom != vopt.ValueFrom_Const {
		return nil, fmt.Errorf("condition [before] does not support value_from type '%s'", cfg.ValueFrom)
	}

	unit, exist := cfg.RemainCfg["unit"].(string)
	if !exist {
		return nil, fmt.Errorf("condition [before] unit is not specified")
	}

	fName, err := GetQueryField(ctx, cfg.Name, fieldsMap, FieldFeatureType_Raw)
	if err != nil {
		return nil, fmt.Errorf("condition [before], %v", err)
	}

	return &BeforeCond{
		mCfg:             cfg,
		mValue:           cfg.ValueOptCfg.Value,
		mUnit:            unit,
		mFilterFieldName: fName,
	}, nil
}

func (cond *BeforeCond) Convert(ctx context.Context) (string, error) {
	unitMap := map[string]string{
		"year":   "y",
		"month":  "M",
		"week":   "w",
		"day":    "d",
		"hour":   "h",
		"minute": "m",
		"second": "s",
	}

	unit, ok := unitMap[cond.mUnit]
	if !ok {
		unit = cond.mUnit // 如果已经缩写过则直接用
	}

	// 统一处理数值类型
	var val any = cond.mValue
	if f, ok := val.(float64); ok {
		val = int64(f)
	}

	return fmt.Sprintf(`{"range":{"%s":{"gte":"now-%v%s","lte":"now"}}}`,
		cond.mFilterFieldName, val, unit), nil
}

func (cond *BeforeCond) Convert2SQL(ctx context.Context) (string, error) {
	sqlStr := fmt.Sprintf(`"%s" >= DATE_add('%s', -%v, CURRENT_TIMESTAMP AT TIME ZONE 'UTC' AT TIME ZONE '%s') 
								AND %s <= CURRENT_TIMESTAMP AT TIME ZONE 'UTC' AT TIME ZONE '%s'`,
		cond.mFilterFieldName, cond.mUnit, cond.mValue, os.Getenv("TZ"), cond.mFilterFieldName, os.Getenv("TZ"))
	return sqlStr, nil
}
