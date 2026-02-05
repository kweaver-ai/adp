package condition

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"

	vopt "uniquery/common/value_opt"
)

type MatchCond struct {
	mCfg              *CondCfg
	mFilterFieldNames []string
}

func NewMatchCond(ctx context.Context, cfg *CondCfg, vType string, fieldsMap map[string]*ViewField) (Condition, error) {
	if cfg.ValueOptCfg.ValueFrom != vopt.ValueFrom_Const {
		return nil, fmt.Errorf("condition [match] does not support value_from type '%s'", cfg.ValueFrom)
	}

	// name := getFilterFieldName(ctx, cfg.Name, fieldsMap, true)
	name, err := GetQueryField(ctx, cfg.Name, fieldsMap, FieldFeatureType_Fulltext)
	if err != nil {
		return nil, fmt.Errorf("condition [match], %v", err)
	}
	var fields []string
	// 如果指定*查询，并且字段列表为自己选的字段，那么将查询的字段替换成视图的字段列表
	if name == AllField && vType == vType_Custom {
		fields = make([]string, 0, len(fieldsMap))
		for fieldName := range fieldsMap {
			fields = append(fields, fieldName)
		}
	} else {
		fields = append(fields, name)
	}

	return &MatchCond{
		mCfg:              cfg,
		mFilterFieldNames: fields,
	}, nil
}

func (cond *MatchCond) Convert(ctx context.Context) (string, error) {
	v := cond.mCfg.Value
	vStr, ok := v.(string)
	if ok {
		v = fmt.Sprintf("%q", vStr)
	}

	fields, err := sonic.Marshal(cond.mFilterFieldNames)
	if err != nil {
		return "", fmt.Errorf("condition [match] marshal fields error: %s", err.Error())
	}

	dslStr := fmt.Sprintf(`
					{
						"multi_match": {
							"query": %v,
							"type": "best_fields",
							"fields": %v
						}
					}`, v, string(fields))

	return dslStr, nil
}

// SQL 类不支持全文检索
func (cond *MatchCond) Convert2SQL(ctx context.Context) (string, error) {
	return "", fmt.Errorf("condition [match] does not support convert to SQL")
}
