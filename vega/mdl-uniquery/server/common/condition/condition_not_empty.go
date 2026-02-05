package condition

import (
	"context"
	"fmt"

	dtype "uniquery/interfaces/data_type"
)

type NotEmptyCond struct {
	mCfg             *CondCfg
	mFilterFieldName string
}

func NewNotEmptyCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*ViewField) (Condition, error) {
	// 只允许字符串类型
	if !dtype.DataType_IsString(cfg.NameField.Type) {
		return nil, fmt.Errorf("condition [not_empty] left field %s is not of string type, but %s", cfg.Name, cfg.NameField.Type)
	}

	featureType := FieldFeatureType_Raw
	if IsTextType(fieldsMap[cfg.Name]) {
		featureType = FieldFeatureType_Keyword
	}

	fName, err := GetQueryField(ctx, cfg.Name, fieldsMap, featureType)
	if err != nil {
		return nil, fmt.Errorf("condition [not_empty], %v", err)
	}

	return &NotEmptyCond{
		mCfg:             cfg,
		mFilterFieldName: fName,
	}, nil

}

// 字段存在且不能为空字符串
func (cond *NotEmptyCond) Convert(ctx context.Context) (string, error) {
	dslStr := fmt.Sprintf(`
	{
		"bool": {
			"must": {
				"exists": {
					"field": "%s"
				}
			},
			"must_not": {
				"term": {
					"%s": ""
				}
			}
		}
	}`, cond.mFilterFieldName, cond.mFilterFieldName)

	return dslStr, nil
}

func (cond *NotEmptyCond) Convert2SQL(ctx context.Context) (string, error) {
	sqlStr := fmt.Sprintf(`"%s" IS NOT NULL AND "%s" <> ''`, cond.mFilterFieldName, cond.mFilterFieldName)
	return sqlStr, nil
}
