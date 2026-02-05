package condition

import (
	"context"
	"fmt"

	vopt "uniquery/common/value_opt"
	dtype "uniquery/interfaces/data_type"
)

type RangeCond struct {
	mCfg             *CondCfg
	mValue           []any
	mFilterFieldName string
}

func NewRangeCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*ViewField) (Condition, error) {
	if cfg.ValueOptCfg.ValueFrom != vopt.ValueFrom_Const {
		return nil, fmt.Errorf("condition [range] does not support value_from type '%s'", cfg.ValueFrom)
	}

	val, ok := cfg.ValueOptCfg.Value.([]any)
	if !ok {
		return nil, fmt.Errorf("condition [range] right value should be an array of length 2")
	}

	if len(val) != 2 {
		return nil, fmt.Errorf("condition [range] right value should be an array of length 2")
	}

	if !IsSameType(val) {
		return nil, fmt.Errorf("condition [range] right value should be of the same type")
	}

	fName, err := GetQueryField(ctx, cfg.Name, fieldsMap, FieldFeatureType_Raw)
	if err != nil {
		return nil, fmt.Errorf("condition [range], %v", err)
	}

	return &RangeCond{
		mCfg:             cfg,
		mValue:           val,
		mFilterFieldName: fName,
	}, nil
}

// range 左闭右开区间
func (cond *RangeCond) Convert(ctx context.Context) (string, error) {
	gte := cond.mValue[0]
	lt := cond.mValue[1]

	if _, ok := gte.(string); ok {
		gte = fmt.Sprintf("%q", gte)
		lt = fmt.Sprintf("%q", lt)
	}

	var dslStr string
	if cond.mCfg.NameField.Type == dtype.DataType_Datetime {
		var format string
		switch gte.(type) {
		case string:
			format = "strict_date_optional_time"
		case float64:
			format = "epoch_millis"
			gte = int64(gte.(float64))
			lt = int64(lt.(float64))
		}

		dslStr = fmt.Sprintf(`
		{
			"range": {
				"%s": {
					"gte": %v,
					"lt": %v,
					"format": "%s"
				}
			}
		}`, cond.mFilterFieldName, gte, lt, format)

	} else {
		dslStr = fmt.Sprintf(`
			{
				"range": {
					"%s": {
						"gte": %v,
						"lt": %v
					}
				}
			}`, cond.mFilterFieldName, gte, lt)

	}

	return dslStr, nil
}

func (cond *RangeCond) Convert2SQL(ctx context.Context) (string, error) {
	// range表示左闭右开区间 [gte, lt)
	gte := cond.mValue[0]
	lt := cond.mValue[1]

	// 处理字符串类型的值，需要用单引号包裹
	gteStr, ok := gte.(string)
	if ok {
		gteStr = Special.Replace(fmt.Sprintf("%q", gteStr))
	} else {
		gteStr = fmt.Sprintf("%v", gte)
	}

	ltStr, ok := lt.(string)
	if ok {
		ltStr = Special.Replace(fmt.Sprintf("%q", ltStr))
	} else {
		ltStr = fmt.Sprintf("%v", lt)
	}

	// 构建SQL条件：字段名 >= 左边界 AND 字段名 < 右边界
	sqlStr := fmt.Sprintf("\"%s\" >= %s AND \"%s\" < %s", cond.mFilterFieldName, gteStr, cond.mFilterFieldName, ltStr)
	return sqlStr, nil
}
