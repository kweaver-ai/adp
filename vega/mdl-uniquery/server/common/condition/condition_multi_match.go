package condition

import (
	"context"
	"fmt"
	"uniquery/common"

	"github.com/bytedance/sonic"
)

type MultiMatchCond struct {
	mCfg              *CondCfg
	mFilterFieldNames []string
}

func NewMultiMatchCond(ctx context.Context, cfg *CondCfg, vType string, fieldsMap map[string]*ViewField) (Condition, error) {

	// 从cfg的 ReaminCfg 中获取 fields，这是属于 multi_match的fields字段，是个字符串数组，
	// 如果想要全部字段匹配，可不填或者填 ["*"], 不支持填字符串 *， 需要一个数组
	var fields []string
	cfgFields, exist := cfg.RemainCfg["fields"]
	if exist {
		// 存在 fields 时需要是一个数组
		if !common.IsSlice(cfgFields) {
			return nil, fmt.Errorf("condition [multi_match] 'fields' value should be an array")
		}
		// 字段数组里的需要是个字符串数组
		for _, cfgField := range cfgFields.([]any) {
			field, ok := cfgField.(string)
			if !ok {
				return nil, fmt.Errorf("condition [multi_match] 'fields' value should be a string array, contain non string value[%v]", cfgField)
			}

			// 字段数组里的每个元素都需要是字符串
			// name := getFilterFieldName(ctx, field, fieldsMap, true)
			name, err := GetQueryField(ctx, field, fieldsMap, FieldFeatureType_Fulltext)
			if err != nil {
				return nil, fmt.Errorf("condition [multi_match], %v", err)
			}
			if name == AllField && vType == vType_Custom {
				fields = make([]string, 0, len(fieldsMap))
				for fieldName := range fieldsMap {
					fields = append(fields, fieldName)
				}
			} else {
				if _, ok := fieldsMap[field]; !ok {
					return nil, fmt.Errorf("condition [multi_match] 'fields' exists any field not exists in data_view [%s]", field)
				}
				fields = append(fields, name)
			}
		}
	}

	// 校验match_type的有效性, match_type可以为空
	matchType, exist := cfg.RemainCfg["match_type"]
	if exist && matchType != "" {
		mtype, ok := matchType.(string)
		if !ok {
			return nil, fmt.Errorf("condition [multi_match] 'match_type' value should be a string, actual is[%v]", matchType)
		}
		if !MatchTypeMap[mtype] {
			return nil, fmt.Errorf("condition [multi_match] 'match_type' value should be one of [%v], actual is[%v]", MatchTypeMap, mtype)
		}
	}

	return &MultiMatchCond{
		mCfg:              cfg,
		mFilterFieldNames: fields,
	}, nil
}

func (cond *MultiMatchCond) Convert(ctx context.Context) (string, error) {
	v := cond.mCfg.Value
	vStr, ok := v.(string)
	if ok {
		v = fmt.Sprintf("%q", vStr)
	}

	fields, err := sonic.Marshal(cond.mFilterFieldNames)
	if err != nil {
		return "", fmt.Errorf("condition [multi_match] marshal fields error: %s", err.Error())
	}

	// 默认是 best_fields
	matchType := "best_fields"
	if mt, exist := cond.mCfg.RemainCfg["match_type"]; exist {
		if mtStr, ok := mt.(string); exist && ok {
			matchType = mtStr
		} else {
			return "", fmt.Errorf("condition [multi_match] match_type[%v] should be a string", mt)
		}
	}

	dslStr := fmt.Sprintf(`
					{
						"multi_match": {
							"query": %v,
							"type": "%s"`, v, matchType)

	// 如果不指定 fields，则用 index.query.default_field 配置的字段查询，默认是*
	if len(cond.mFilterFieldNames) > 0 {
		dslStr = fmt.Sprintf(`%s,
							"fields": %v
						}
					}`, dslStr, string(fields))
	} else {
		dslStr = fmt.Sprintf(`%s
						}
					}`, dslStr)
	}

	return dslStr, nil
}

func (cond *MultiMatchCond) Convert2SQL(ctx context.Context) (string, error) {
	return "", nil
}
