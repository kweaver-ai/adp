package condition

import (
	"context"
	"fmt"

	dtype "uniquery/interfaces/data_type"
)

type FalseCond struct {
	mCfg             *CondCfg
	mFilterFieldName string
}

func NewFalseCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*ViewField) (Condition, error) {
	if cfg.NameField.Type != dtype.DataType_Boolean {
		return nil, fmt.Errorf("condition [false] left field is not a boolean field: %s:%s", cfg.NameField.Name, cfg.NameField.Type)
	}

	featureType := FieldFeatureType_Raw
	if IsTextType(fieldsMap[cfg.Name]) {
		featureType = FieldFeatureType_Keyword
	}

	fName, err := GetQueryField(ctx, cfg.Name, fieldsMap, featureType)
	if err != nil {
		return nil, fmt.Errorf("condition [false], %v", err)
	}

	return &FalseCond{
		mCfg:             cfg,
		mFilterFieldName: fName,
	}, nil
}

// term 查询逻辑等于 字段存在 + 相等
func (cond *FalseCond) Convert(ctx context.Context) (string, error) {
	dslStr := fmt.Sprintf(`
					{
						"term": {
							"%s": false	
						}
					}`, cond.mFilterFieldName)

	return dslStr, nil
}

func (cond *FalseCond) Convert2SQL(ctx context.Context) (string, error) {
	sqlStr := fmt.Sprintf(`"%s" = false`, cond.mFilterFieldName)
	return sqlStr, nil
}
