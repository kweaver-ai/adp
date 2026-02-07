// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"encoding/json"
	"fmt"
)

type KnnCond struct {
	mCfg             *CondCfg
	mFilterFieldName string
	mSubConds        []Condition
}

func NewKnnCond(ctx context.Context, cfg *CondCfg, fieldScope uint8, fieldsMap map[string]*ViewField) (Condition, error) {
	if cfg.ValueOptCfg.ValueFrom != ValueFrom_Const {
		return nil, fmt.Errorf("condition [knn] does not support value_from type '%s'", cfg.ValueFrom)
	}

	name := getFilterFieldName(cfg.Name, fieldsMap, true)
	var field string
	// 如果指定*查询，则把 * 换成 _vector
	if name == AllField {
		field = "_vector"
	} else {
		field = name
	}

	subConds := []Condition{}
	for _, subCond := range cfg.SubConds {
		cond, err := NewCondition(ctx, subCond, fieldScope, fieldsMap)
		if err != nil {
			return nil, err
		}

		if cond != nil {
			subConds = append(subConds, cond)
		}

	}

	return &KnnCond{
		mCfg:             cfg,
		mFilterFieldName: field,
		mSubConds:        subConds,
	}, nil
}

func (cond *KnnCond) Convert(ctx context.Context, vectorizer func(ctx context.Context, words []string) ([]*VectorResp, error)) (string, error) {
	v := fmt.Sprintf("%v", cond.mCfg.Value)

	vector, err := vectorizer(ctx, []string{v})
	if err != nil {
		// 如果错误是因为 DefaultSmallModelEnabled 为 false，则忽略此 knn 条件，返回空字符串
		if err.Error() == DEFAULT_SMALL_MODEL_ENABLED_FALSE_ERROR {
			return "", nil
		}
		return "", fmt.Errorf("condition [knn]: vectorizer [%s] failed, error: %s", v, err.Error())
	}
	res, err := json.Marshal(vector[0].Vector)
	if err != nil {
		return "", fmt.Errorf("condition [in] json marshal right value failed, %s", err.Error())
	}

	// sub condition
	subDSL := ""
	if len(cond.mSubConds) > 0 {
		subDSL = `
		,
		"filter": {
			"bool": {
				"must": [
					%s
				]
			}
		}
		`

		subCondStr := ""
		validSubDSLs := []string{}
		for _, subCond := range cond.mSubConds {
			dsl, err := subCond.Convert(ctx, vectorizer)
			if err != nil {
				return "", err
			}

			// 过滤掉空字符串（被忽略的条件）
			if dsl != "" && dsl != "{}" {
				validSubDSLs = append(validSubDSLs, dsl)
			}
		}

		// 如果有有效的子条件，才添加 filter
		if len(validSubDSLs) > 0 {
			for i, dsl := range validSubDSLs {
				if i != len(validSubDSLs)-1 {
					subCondStr += dsl + ","
				} else {
					subCondStr += dsl
				}
			}
			subDSL = fmt.Sprintf(subDSL, subCondStr)
		} else {
			// 所有子条件都被忽略，不添加 filter
			subDSL = ""
		}
	}

	// limit_key 和 limit_value 未给时，填入默认值
	key := cond.mCfg.RemainCfg["limit_key"]
	value := cond.mCfg.RemainCfg["limit_value"]
	if key == nil || key == "" {
		key = KNN_LIMIT_KEY_DEFAULT
	}
	if value == nil {
		value = KNN_LIMIT_VALUE_DEFAULT
	}

	dslStr := fmt.Sprintf(`
					{
						"knn": {
							"%s":{
								"%s": %v,
								"vector": %v
								%s
							}
						}
					}`, cond.mFilterFieldName, key, value,
		string(res), subDSL)

	return dslStr, nil
}

func (cond *KnnCond) Convert2SQL(ctx context.Context) (string, error) {
	return "", nil
}
