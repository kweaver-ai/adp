package condition

import (
	"context"
	"fmt"
)

type NullCond struct {
	mCfg             *CondCfg
	mFilterFieldName string
}

func NewNullCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*ViewField) (Condition, error) {
	featureType := FieldFeatureType_Raw
	if IsTextType(fieldsMap[cfg.Name]) {
		featureType = FieldFeatureType_Keyword
	}

	fName, err := GetQueryField(ctx, cfg.Name, fieldsMap, featureType)
	if err != nil {
		return nil, fmt.Errorf("condition [null], %v", err)
	}

	return &NullCond{
		mCfg:             cfg,
		mFilterFieldName: fName,
	}, nil

}

// 检查字段值是否 IS NULL， OpenSearch 默认不会对 null 值进行索引，
// 因此 IS NULL 的逻辑等同于查找“该字段不存在索引值”的文档，查询会匹配以下情况：
// 1. 文档中完全没有这个字段
// 2. 该字段在 JSON 中被显示设为 null
// 3. 该字段是一个空数组
func (cond *NullCond) Convert(ctx context.Context) (string, error) {
	dslStr := fmt.Sprintf(`
	{
		"bool": {
			"must_not": {
				"exists": {
					"field": "%s"
				}
			}
		}
	}`, cond.mFilterFieldName)

	return dslStr, nil
}

func (cond *NullCond) Convert2SQL(ctx context.Context) (string, error) {
	sqlStr := fmt.Sprintf(`"%s" IS NULL`, cond.mFilterFieldName)
	return sqlStr, nil
}
