package condition

import (
	"context"
	"fmt"

	"uniquery/common"
	vopt "uniquery/common/value_opt"
)

type LteCond struct {
	mCfg             *CondCfg
	mFilterFieldName string
}

func NewLteCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*ViewField) (Condition, error) {
	if cfg.ValueOptCfg.ValueFrom != vopt.ValueFrom_Const {
		return nil, fmt.Errorf("condition [lte] does not support value_from type '%s'", cfg.ValueFrom)
	}

	if common.IsSlice(cfg.ValueOptCfg.Value) {
		return nil, fmt.Errorf("condition [lte] only supports single value")
	}

	fName, err := GetQueryField(ctx, cfg.Name, fieldsMap, FieldFeatureType_Raw)
	if err != nil {
		return nil, fmt.Errorf("condition [lte], %v", err)
	}

	return &LteCond{
		mCfg:             cfg,
		mFilterFieldName: fName,
	}, nil

}

func (cond *LteCond) Convert(ctx context.Context) (string, error) {
	v := cond.mCfg.Value
	vStr, ok := v.(string)
	if ok {
		v = fmt.Sprintf("%q", vStr)
	}
	dslStr := fmt.Sprintf(`
					{
						"range": {
							"%s": {
								"lte": %v
							}
						}
					}`, cond.mFilterFieldName, v)

	return dslStr, nil
}

func (cond *LteCond) Convert2SQL(ctx context.Context) (string, error) {
	v := cond.mCfg.Value
	vStr, ok := v.(string)
	if ok {
		v = fmt.Sprintf(`'%v'`, vStr)
	}
	sqlStr := fmt.Sprintf(`"%s" <= %v`, cond.mFilterFieldName, v)

	return sqlStr, nil
}
