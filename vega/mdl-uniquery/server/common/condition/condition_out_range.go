package condition

import (
	"context"
	"fmt"

	vopt "uniquery/common/value_opt"
	dtype "uniquery/interfaces/data_type"
)

type OutRangeCond struct {
	mCfg             *CondCfg
	mValue           []any
	mFilterFieldName string
}

func NewOutRangeCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*ViewField) (Condition, error) {
	if cfg.ValueOptCfg.ValueFrom != vopt.ValueFrom_Const {
		return nil, fmt.Errorf("condition [out_range] does not support value_from type '%s'", cfg.ValueFrom)
	}

	val, ok := cfg.ValueOptCfg.Value.([]any)
	if !ok {
		return nil, fmt.Errorf("condition [out_range] right value should be an array of length 2")
	}

	if len(val) != 2 {
		return nil, fmt.Errorf("condition [out_range] right value should be an array of length 2")
	}

	if !IsSameType(val) {
		return nil, fmt.Errorf("condition [out_range] right value should be of the same type")
	}

	fName, err := GetQueryField(ctx, cfg.Name, fieldsMap, FieldFeatureType_Raw)
	if err != nil {
		return nil, fmt.Errorf("condition [out_range], %v", err)
	}

	return &OutRangeCond{
		mCfg:             cfg,
		mValue:           val,
		mFilterFieldName: fName,
	}, nil
}

// out_range  (-inf, value[0]) || [value[1], +inf)
func (cond *OutRangeCond) Convert(ctx context.Context) (string, error) {
	lt := cond.mValue[0]
	gte := cond.mValue[1]

	if _, ok := lt.(string); ok {
		lt = fmt.Sprintf("%q", lt)
		gte = fmt.Sprintf("%q", gte)
	}

	var dslStr string
	if cond.mCfg.NameField.Type == dtype.DataType_Datetime {
		format := ""
		switch lt.(type) {
		case string:
			format = "strict_date_optional_time"
		case float64:
			format = "epoch_millis"
			lt = int64(lt.(float64))
			gte = int64(gte.(float64))
		}

		dslStr = fmt.Sprintf(`
					{
						"bool": {
							"should": [
								{
									"range": {
										"%s": {
											"lt": %v,
											"format": "%s"
										}
									}
								},
								{
									"range": {
										"%s": {
											"gte":  %v,
											"format": "%s"
										}
									}
								}
							]
						}
					}`, cond.mFilterFieldName, lt, format, cond.mFilterFieldName, gte, format)

	} else {

		dslStr = fmt.Sprintf(`
		{
			"bool": {
				"should": [
					{
						"range": {
							"%s": {
								"lt": %v
							}
						}
					},
					{
						"range": {
							"%s": {
								"gte":  %v
							}
						}
					}
				]
			}
		}`, cond.mFilterFieldName, lt, cond.mFilterFieldName, gte)

	}

	return dslStr, nil
}

func (cond *OutRangeCond) Convert2SQL(ctx context.Context) (string, error) {
	// out_range表示 (-inf, value[0]) || [value[1], +inf)
	lt := cond.mValue[0]
	gte := cond.mValue[1]

	// 处理字符串类型的值，需要用单引号包裹
	ltStr, ok := lt.(string)
	if ok {
		ltStr = Special.Replace(fmt.Sprintf("%q", ltStr))
	} else {
		ltStr = fmt.Sprintf("%v", lt)
	}

	gteStr, ok := gte.(string)
	if ok {
		gteStr = Special.Replace(fmt.Sprintf("%q", gteStr))
	} else {
		gteStr = fmt.Sprintf("%v", gte)
	}

	// 构建SQL条件：字段名 < 左边界 OR 字段名 >= 右边界
	sqlStr := fmt.Sprintf("(\"%s\" < %s OR \"%s\" >= %s)", cond.mFilterFieldName, ltStr, cond.mFilterFieldName, gteStr)
	return sqlStr, nil
}
