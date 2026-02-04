package condition

import (
	"context"
	"fmt"
	"os"

	vopt "uniquery/common/value_opt"
	dtype "uniquery/interfaces/data_type"
)

type BeforeCond struct {
	mCfg             *CondCfg
	mValue           []any
	mFilterFieldName string
}

func NewBeforeCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*ViewField) (Condition, error) {
	if !dtype.DataType_IsDate(cfg.NameField.Type) {
		return nil, fmt.Errorf("condition [before] left field is not a date field: %s:%s", cfg.NameField.Name, cfg.NameField.Type)
	}

	if cfg.ValueOptCfg.ValueFrom != vopt.ValueFrom_Const {
		return nil, fmt.Errorf("condition [before] does not support value_from type '%s'", cfg.ValueFrom)
	}

	val, ok := cfg.ValueOptCfg.Value.([]any)
	if !ok {
		return nil, fmt.Errorf("condition [before] right value should be an array of length 2")
	}

	if len(val) != 2 {
		return nil, fmt.Errorf("condition [before] right value should be an array of length 2")
	}

	if _, ok := val[0].(string); ok {
		return nil, fmt.Errorf("condition [before]'s interval value should be an number")
	}
	_, ok = val[1].(string)
	if !ok {
		return nil, fmt.Errorf("condition [before]'s interval value should be a string")
	}

	fName, err := GetQueryField(ctx, cfg.Name, fieldsMap, FieldFeatureType_Raw)
	if err != nil {
		return nil, fmt.Errorf("condition [before], %v", err)
	}

	return &BeforeCond{
		mCfg:             cfg,
		mValue:           val,
		mFilterFieldName: fName,
	}, nil
}

func (cond *BeforeCond) Convert(ctx context.Context) (string, error) {
	return "", nil
}

func (cond *BeforeCond) Convert2SQL(ctx context.Context) (string, error) {
	unit := cond.mValue[1].(string)
	sqlStr := fmt.Sprintf(`"%s" >= DATE_add('%s', -%v, CURRENT_TIMESTAMP AT TIME ZONE 'UTC' AT TIME ZONE '%s') 
								AND %s <= CURRENT_TIMESTAMP AT TIME ZONE 'UTC' AT TIME ZONE '%s'`,
		cond.mFilterFieldName, unit, cond.mValue[0], os.Getenv("TZ"), cond.mFilterFieldName, os.Getenv("TZ"))
	return sqlStr, nil
}
