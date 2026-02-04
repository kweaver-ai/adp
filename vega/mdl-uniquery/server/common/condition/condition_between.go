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
	if !dtype.DataType_IsDate(cfg.NameField.Type) {
		return nil, fmt.Errorf("condition [between] left field is not a date field: %s:%s", cfg.NameField.Name, cfg.NameField.Type)
	}

	if cfg.ValueOptCfg.ValueFrom != vopt.ValueFrom_Const {
		return nil, fmt.Errorf("condition [between] does not support value_from type '%s'", cfg.ValueFrom)
	}

	val, ok := cfg.ValueOptCfg.Value.([]any)
	if !ok {
		return nil, fmt.Errorf("condition [between] right value should be an array of length 2")
	}

	if len(val) != 2 {
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
	return "", nil
}

func (cond *BetweenCond) Convert2SQL(ctx context.Context) (string, error) {
	sqlStr := fmt.Sprintf(`"%s" BETWEEN DATE_TRUNC('minute', CAST('%v' AS TIMESTAMP)) AND DATE_TRUNC('minute', CAST('%v' AS TIMESTAMP))`,
		cond.mFilterFieldName, cond.mValue[0], cond.mValue[1])

	return sqlStr, nil
}
