package condition

import (
	"context"
	"fmt"

	"uniquery/common"
	vopt "uniquery/common/value_opt"
)

type EqCond struct {
	mCfg             *CondCfg
	mFilterFieldName string
}

func NewEqCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*ViewField) (Condition, error) {
	if cfg.ValueOptCfg.ValueFrom != vopt.ValueFrom_Const {
		return nil, fmt.Errorf("condition [eq] does not support value_from type '%s'", cfg.ValueFrom)
	}

	if common.IsSlice(cfg.ValueOptCfg.Value) {
		return nil, fmt.Errorf("condition [eq] only supports single value")
	}

	featureType := FieldFeatureType_Raw
	if IsTextType(fieldsMap[cfg.Name]) {
		featureType = FieldFeatureType_Keyword
	}

	fName, err := GetQueryField(ctx, cfg.Name, fieldsMap, featureType)
	if err != nil {
		return nil, fmt.Errorf("condition [eq], %v", err)
	}

	return &EqCond{
		mCfg:             cfg,
		mFilterFieldName: fName,
	}, nil

}

// 注：由于term是包含操作，而不是等值比较，如果field的值是数组的话，无法做到精确相等。
func (cond *EqCond) Convert(ctx context.Context) (string, error) {
	v := cond.mCfg.Value
	vStr, ok := v.(string)
	if ok {
		v = fmt.Sprintf("%q", vStr)
	}

	dslStr := fmt.Sprintf(`
					{
						"term": {
							"%s": {
								"value": %v
							}
						}
					}`, cond.mFilterFieldName, v)

	return dslStr, nil
}

func (cond *EqCond) Convert2SQL(ctx context.Context) (string, error) {
	v := cond.mCfg.Value
	vStr, ok := v.(string)
	if ok {
		v = fmt.Sprintf(`'%v'`, vStr)
	}
	sqlStr := fmt.Sprintf(`"%s" = %v`, cond.mFilterFieldName, v)

	return sqlStr, nil
}
