package condition

import (
	"context"
	"fmt"
)

type NotNullCond struct {
	mCfg             *CondCfg
	mFilterFieldName string
}

func NewNotNullCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*ViewField) (Condition, error) {
	featureType := FieldFeatureType_Raw
	if IsTextType(fieldsMap[cfg.Name]) {
		featureType = FieldFeatureType_Keyword
	}

	fName, err := GetQueryField(ctx, cfg.Name, fieldsMap, featureType)
	if err != nil {
		return nil, fmt.Errorf("condition [not_null], %v", err)
	}

	return &NotNullCond{
		mCfg:             cfg,
		mFilterFieldName: fName,
	}, nil

}

// 检查字段值是否为空字符串
func (cond *NotNullCond) Convert(ctx context.Context) (string, error) {
	dslStr := fmt.Sprintf(`
	{
		"exists": {
			"field": "%s"
		}
	}`, cond.mFilterFieldName)

	return dslStr, nil
}

func (cond *NotNullCond) Convert2SQL(ctx context.Context) (string, error) {
	sqlStr := fmt.Sprintf(`"%s" IS NOT NULL`, cond.mFilterFieldName)
	return sqlStr, nil
}
