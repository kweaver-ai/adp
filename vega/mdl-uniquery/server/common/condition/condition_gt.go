package condition

import (
	"context"
	"fmt"

	"uniquery/common"
	vopt "uniquery/common/value_opt"
)

type GtCond struct {
	mCfg             *CondCfg
	mFilterFieldName string
}

func NewGtCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*ViewField) (Condition, error) {
	if cfg.ValueOptCfg.ValueFrom != vopt.ValueFrom_Const {
		return nil, fmt.Errorf("condition [gt] does not support value_from type '%s'", cfg.ValueFrom)
	}

	if common.IsSlice(cfg.ValueOptCfg.Value) {
		return nil, fmt.Errorf("condition [gt] only supports single value")
	}

	fName, err := GetQueryField(ctx, cfg.Name, fieldsMap, FieldFeatureType_Raw)
	if err != nil {
		return nil, fmt.Errorf("condition [gt], %v", err)
	}

	return &GtCond{
		mCfg:             cfg,
		mFilterFieldName: fName,
	}, nil

}

func (cond *GtCond) Convert(ctx context.Context) (string, error) {
	v := cond.mCfg.Value
	vStr, ok := v.(string)
	if ok {
		v = fmt.Sprintf("%q", vStr)
	}
	dslStr := fmt.Sprintf(`
					{
						"range": {
							"%s": {
								"gt": %v
							}
						}
					}`, cond.mFilterFieldName, v)

	return dslStr, nil

}

func (cond *GtCond) Convert2SQL(ctx context.Context) (string, error) {
	v := cond.mCfg.Value
	vStr, ok := v.(string)
	if ok {
		v = fmt.Sprintf(`'%v'`, vStr)
	}
	sqlStr := fmt.Sprintf(`"%s" > %v`, cond.mFilterFieldName, v)

	return sqlStr, nil
}
