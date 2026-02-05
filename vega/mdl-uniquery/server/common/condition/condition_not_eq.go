package condition

import (
	"context"
	"fmt"

	"uniquery/common"
	vopt "uniquery/common/value_opt"
)

type NotEqCond struct {
	mCfg             *CondCfg
	mFilterFieldName string
}

func NewNotEqCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*ViewField) (Condition, error) {
	if cfg.ValueOptCfg.ValueFrom != vopt.ValueFrom_Const {
		return nil, fmt.Errorf("condition [not_eq] does not support value_from type '%s'", cfg.ValueFrom)
	}

	if common.IsSlice(cfg.ValueOptCfg.Value) {
		return nil, fmt.Errorf("condition [not_eq] only supports single value")
	}

	featureType := FieldFeatureType_Raw
	if IsTextType(fieldsMap[cfg.Name]) {
		featureType = FieldFeatureType_Keyword
	}

	fName, err := GetQueryField(ctx, cfg.Name, fieldsMap, featureType)
	if err != nil {
		return nil, fmt.Errorf("condition [not_eq], %v", err)
	}

	return &NotEqCond{
		mCfg:             cfg,
		mFilterFieldName: fName,
	}, nil

}

/*
	{
	  "bool": {
	    "must_not": [
	      {
	        "term": {
	          "<field>": {
	            "value": <value>
	          }
	        }
	      }
	    ]
	  }
	}
*/
func (cond *NotEqCond) Convert(ctx context.Context) (string, error) {
	v := cond.mCfg.Value
	vStr, ok := v.(string)
	if ok {
		v = fmt.Sprintf("%q", vStr)
	}

	dslStr := fmt.Sprintf(`
					{
						"bool": {
							"must_not": [
								{
									"term": {
										"%s": {
											"value": %v
										}
									}
								}
							]
						}
					}`, cond.mFilterFieldName, v)

	return dslStr, nil
}

func (cond *NotEqCond) Convert2SQL(ctx context.Context) (string, error) {
	v := cond.mCfg.Value
	vStr, ok := v.(string)
	if ok {
		v = fmt.Sprintf(`'%v'`, vStr)
	}
	sqlStr := fmt.Sprintf(`"%s" <> %v`, cond.mFilterFieldName, v)

	return sqlStr, nil
}
