// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"

	vopt "uniquery/common/value_opt"
	dtype "uniquery/interfaces/data_type"
)

type BetweenCond struct {
	mCfg             *CondCfg
	mValue           []any
	mFilterFieldName string
}

func NewBetweenCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*ViewField) (Condition, error) {
	// 1. 修改校验逻辑：支持时间类型 OR 数值类型
	isDate := dtype.DataType_IsDate(cfg.NameField.Type)
	isNumber := dtype.DataType_IsNumber(cfg.NameField.Type)
	if !isDate && !isNumber {
		return nil, fmt.Errorf("condition [between] left field is neither a date nor a numeric field: %s:%s", cfg.NameField.Name, cfg.NameField.Type)
	}

	if cfg.ValueOptCfg.ValueFrom != vopt.ValueFrom_Const {
		return nil, fmt.Errorf("condition [between] does not support value_from type '%s'", cfg.ValueFrom)
	}

	val, ok := cfg.ValueOptCfg.Value.([]any)
	if !ok || len(val) != 2 {
		return nil, fmt.Errorf("condition [between] right value should be an array of length 2")
	}

	fName, err := GetQueryField(ctx, cfg.Name, fieldsMap, FieldFeatureType_Raw)
	if err != nil {
		return nil, fmt.Errorf("condition [between], %v", err)
	}

	return &BetweenCond{
		mCfg:             cfg,
		mValue:           val,
		mFilterFieldName: fName,
	}, nil
}

func (cond *BetweenCond) Convert(ctx context.Context) (string, error) {
	gte := cond.mValue[0]
	lte := cond.mValue[1]

	// 如果是时间类型，参考 RangeCond 处理格式
	if dtype.DataType_IsDate(cond.mCfg.NameField.Type) {
		var format string
		switch gte.(type) {
		case string:
			format = "yyyy-MM-dd HH:mm:ss.SSS"
			gte = fmt.Sprintf("%q", gte)
			lte = fmt.Sprintf("%q", lte)
		case float64:
			format = "epoch_millis"
			gte = int64(gte.(float64))
			lte = int64(lte.(float64))
		}
		return fmt.Sprintf(`{"range":{"%s":{"gte":%v,"lte":%v,"format":"%s"}}}`, 
			cond.mFilterFieldName, gte, lte, format), nil
	}

	// 数值类型处理
	return fmt.Sprintf(`{"range":{"%s":{"gte":%v,"lte":%v}}}`, cond.mFilterFieldName, gte, lte), nil
}

func (cond *BetweenCond) Convert2SQL(ctx context.Context) (string, error) {
	// 1. 如果是时间类型，保留原有的 TRUNC 处理
	if dtype.DataType_IsDate(cond.mCfg.NameField.Type) {
		return fmt.Sprintf(`"%s" BETWEEN DATE_TRUNC('minute', CAST('%v' AS TIMESTAMP)) AND DATE_TRUNC('minute', CAST('%v' AS TIMESTAMP))`,
			cond.mFilterFieldName, cond.mValue[0], cond.mValue[1]), nil
	}

	// 2. 如果是数值类型，直接生成简单的 BETWEEN 子句
	return fmt.Sprintf(`"%s" BETWEEN %v AND %v`, cond.mFilterFieldName, cond.mValue[0], cond.mValue[1]), nil
}
