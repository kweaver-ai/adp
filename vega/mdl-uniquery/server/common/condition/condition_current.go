// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"errors"
	"fmt"
	"time"

	"uniquery/common"
	vopt "uniquery/common/value_opt"
	dtype "uniquery/interfaces/data_type"
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
	mCfg             *CondCfg
	mValue           string
	mFilterFieldName string
}

func NewCurrentCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*ViewField) (Condition, error) {
	if !dtype.DataType_IsDate(cfg.NameField.Type) {
		return nil, fmt.Errorf("condition [current] left field is not a date field: %s:%s", cfg.NameField.Name, cfg.NameField.Type)
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

	fName, err := GetQueryField(ctx, cfg.Name, fieldsMap, FieldFeatureType_Raw)
	if err != nil {
		return nil, fmt.Errorf("condition [current], %v", err)
	}

	return &CurrentCond{
		mCfg:             cfg,
		mValue:           val,
		mFilterFieldName: fName,
	}, nil
}

func (cond *CurrentCond) Convert(ctx context.Context) (string, error) {
	var gte, lt string
	// 利用 ES 的 Date Math 符号进行取整和偏移
	// y: year, M: month, w: week, d: day, h: hour, m: minute
	switch cond.mValue {
	case CurrentYear:
		gte, lt = "now/y", "now/y+1y"
	case CurrentMonth:
		gte, lt = "now/M", "now/M+1M"
	case CurrentWeek:
		gte, lt = "now/w", "now/w+1w"
	case CurrentDay:
		gte, lt = "now/d", "now/d+1d"
	case CurrentHour:
		gte, lt = "now/h", "now/h+1h"
	case CurrentMinute:
		gte, lt = "now/m", "now/m+1m"
	default:
		return "", fmt.Errorf("unsupported current format: %v", cond.mValue)
	}

	return fmt.Sprintf(`{"range":{"%s":{"gte":"%s","lt":"%s"}}}`, cond.mFilterFieldName, gte, lt), nil
}

func (cond *CurrentCond) Convert2SQL(ctx context.Context) (string, error) {
	// 时区从环境变量里取. 当前年、月、周、天等，转成时间字段属于一个范围
	// 注意：这里使用 common.APP_LOCATION 保持与业务逻辑时区一致
	now := time.Now().In(common.APP_LOCATION)
	var start, end time.Time

	switch cond.mValue {
	case CurrentYear:
		start = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, common.APP_LOCATION)
		end = start.AddDate(1, 0, 0)
	case CurrentMonth:
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, common.APP_LOCATION)
		end = start.AddDate(0, 1, 0)
	case CurrentDay:
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, common.APP_LOCATION)
		end = start.AddDate(0, 0, 1)
	case CurrentWeek:
		// 计算本周周一 (ISO 周)
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7 // 将周日从 0 改为 7
		}
		offset := 1 - weekday // 到周一的偏移天数
		start = time.Date(now.Year(), now.Month(), now.Day()+offset, 0, 0, 0, 0, common.APP_LOCATION)
		end = start.AddDate(0, 0, 7)
	case CurrentHour:
		start = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, common.APP_LOCATION)
		end = start.Add(time.Hour)
	case CurrentMinute:
		start = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, common.APP_LOCATION)
		end = start.Add(time.Minute)
	default:
		return "", fmt.Errorf("unsupported current format: %v", cond.mValue)
	}

	// SQL 采用 BETWEEN start AND end (注意：BETWEEN 是包含两端的，通常 end 需要略微减去 1ms 或使用 >= AND <)
	// 这里沿用原逻辑的 from_unixtime。建议使用 >= start AND < end 以保证边界准确。
	sqlStr := fmt.Sprintf(`"%s" >= from_unixtime(%d) AND "%s" < from_unixtime(%d)`,
		cond.mFilterFieldName, start.Unix(), cond.mFilterFieldName, end.Unix())
	return sqlStr, nil
}
