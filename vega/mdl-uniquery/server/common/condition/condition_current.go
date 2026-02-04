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

	if val != "%Y" &&
		val != "%Y-%m" &&
		val != "%Y-%m-%d" &&
		val != "%Y-%m-%d %H" &&
		val != "%Y-%m-%d %H:%i" &&
		val != "%x-%v" {
		return nil, errors.New(`condition [current] right value should be 
		one of ["%Y", "%Y-%m", "%Y-%m-%d", "%Y-%m-%d %H", "%Y-%m-%d %H:%i", "%x-%v"], actual is ` + val)
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
	return "", nil
}

func (cond *CurrentCond) Convert2SQL(ctx context.Context) (string, error) {
	// 时区从环境变量里取. 当前年、月、周、天等，转成时间字段属于一个范围
	now := time.Now()
	var start, end time.Time
	switch cond.mCfg.Value.(string) {
	case `%Y`:
		start = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, common.APP_LOCATION)
		end = start.AddDate(1, 0, 0)
	case `%Y-%m`:
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, common.APP_LOCATION)
		end = start.AddDate(0, 1, 0)
	case `%Y-%m-%d`:
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, common.APP_LOCATION)
		end = start.AddDate(0, 0, 1)
	case `%x-%v`:
		// 计算本周的周一
		weekday := now.Weekday()
		offset := int(time.Monday - weekday)
		if offset > 0 {
			offset -= 7 // 如果今天是周日，需要减去7天
		}
		start = time.Date(now.Year(), now.Month(), now.Day()+offset, 0, 0, 0, 0, common.APP_LOCATION)
		end = start.AddDate(0, 0, 7)
	case `%Y-%m-%d %H`:
		start = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, common.APP_LOCATION)
		end = start.Add(time.Hour)
	case `%Y-%m-%d %H:%i`:
		start = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, common.APP_LOCATION)
		end = start.Add(time.Minute)
	default:
		return "", fmt.Errorf("unsupported time format: %v", cond.mCfg.Value)
	}

	sqlStr := fmt.Sprintf(`"%s" BETWEEN from_unixtime(%d) AND from_unixtime(%d)`,
		cond.mFilterFieldName, start.UnixMilli()/1000, end.UnixMilli()/1000)
	return sqlStr, nil
}
