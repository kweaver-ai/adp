package condition

import (
	"context"
	"fmt"

	dtype "uniquery/interfaces/data_type"
)

type TrueCond struct {
	mCfg             *CondCfg
	mFilterFieldName string
}

// bool 类型为真
func NewTrueCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*ViewField) (Condition, error) {
	if cfg.NameField.Type != dtype.DataType_Boolean {
		return nil, fmt.Errorf("condition [true] left field is not a boolean field: %s:%s", cfg.NameField.Name, cfg.NameField.Type)
	}

	featureType := FieldFeatureType_Raw
	if IsTextType(fieldsMap[cfg.Name]) {
		featureType = FieldFeatureType_Keyword
	}

	fName, err := GetQueryField(ctx, cfg.Name, fieldsMap, featureType)
	if err != nil {
		return nil, fmt.Errorf("condition [true], %v", err)
	}

	return &TrueCond{
		mCfg:             cfg,
		mFilterFieldName: fName,
	}, nil
}

func (cond *TrueCond) Convert(ctx context.Context) (string, error) {
	dslStr := fmt.Sprintf(`
					{
						"term": {
							"%s": true
						}
					}`, cond.mFilterFieldName)

	return dslStr, nil
}

func (cond *TrueCond) Convert2SQL(ctx context.Context) (string, error) {
	sqlStr := fmt.Sprintf(`"%s" = true`, cond.mFilterFieldName)
	return sqlStr, nil
}
