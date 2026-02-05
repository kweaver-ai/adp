package condition

import (
	"context"
	"fmt"

	vopt "uniquery/common/value_opt"

	"github.com/bytedance/sonic"
)

type KnnVectorCond struct {
	mCfg             *CondCfg
	mFilterFieldName string
	mSubConds        []Condition
}

func NewKnnVectorCond(ctx context.Context, cfg *CondCfg, vType string, fieldsMap map[string]*ViewField) (Condition, error) {
	if cfg.ValueOptCfg.ValueFrom != vopt.ValueFrom_Const {
		return nil, fmt.Errorf("condition [knn_vector] does not support value_from type '%s'", cfg.ValueFrom)
	}

	// knnByte, err := json.Marshal(cfg.ValueOptCfg.Value)
	// if err != nil {
	// 	return nil, err
	// }
	// var knnParam KnnParams
	// err = json.Unmarshal(knnByte, &knnParam)
	// if err != nil {
	// 	return nil, err
	// }

	// val, ok := cfg.ValueOptCfg.Value.([]any)
	// if !ok {
	// 	return nil, fmt.Errorf("condition [knn] right value should be an array of length 2")
	// }

	// if len(val) != 2 {
	// 	return nil, fmt.Errorf("condition [knn] right value should be an array of length 2")
	// }

	// _, ok = val[1].(float64)
	// if !ok {
	// 	return nil, fmt.Errorf("condition [knn]'s interval value should be a integer")
	// }

	// name := getFilterFieldName(ctx, cfg.Name, fieldsMap, true)
	name, err := GetQueryField(ctx, cfg.Name, fieldsMap, FieldFeatureType_Vector)
	if err != nil {
		return nil, fmt.Errorf("condition [knn_vector], %v", err)
	}
	// 不支持*查询
	// var field string
	// // 如果指定*查询，则把 * 换成 _vector
	// if name == AllField {
	// 	field = "_vector"
	// } else {
	// 	field = name
	// }

	subConds := []Condition{}
	for _, subCond := range cfg.SubConds {
		cond, _, err := NewCondition(ctx, subCond, vType, fieldsMap)
		if err != nil {
			return nil, err
		}

		if cond != nil {
			subConds = append(subConds, cond)
		}

	}

	return &KnnVectorCond{
		mCfg:             cfg,
		mFilterFieldName: name,
		mSubConds:        subConds,
	}, nil
}

func (cond *KnnVectorCond) Convert(ctx context.Context) (string, error) {
	// vector := fmt.Sprintf("%v", cond.mCfg.Value)

	// vector, err := vectorizer(ctx, []string{v})
	// if err != nil {
	// 	return "", fmt.Errorf("condition [knn]: vectorizer [%s] failed, error: %s", v, err.Error())
	// }
	// value是这样的格式： [2, 3, 5, 7]
	res, err := sonic.Marshal(cond.mCfg.Value)
	if err != nil {
		return "", fmt.Errorf("condition [knn_vector] json marshal right value failed, %s", err.Error())
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
		for i, subCond := range cond.mSubConds {
			dsl, err := subCond.Convert(ctx)
			if err != nil {
				return "", err
			}

			if i != len(cond.mSubConds)-1 {
				dsl += ","
			}

			subCondStr += dsl

		}
		subDSL = fmt.Sprintf(subDSL, subCondStr)
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
					}`, cond.mFilterFieldName, cond.mCfg.RemainCfg["limit_key"], cond.mCfg.RemainCfg["limit_value"],
		string(res), subDSL)

	return dslStr, nil
}

func (cond *KnnVectorCond) Convert2SQL(ctx context.Context) (string, error) {
	return "", nil
}
