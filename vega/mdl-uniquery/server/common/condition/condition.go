package condition

import (
	"context"
	"fmt"
	"strings"

	"github.com/kweaver-ai/kweaver-go-lib/logger"

	dtype "uniquery/interfaces/data_type"
)

const MaxSubCondition = 100

// sql的字符串转义
var Special = strings.NewReplacer(`\`, `\\\\`, `'`, `\'`, `%`, `\%`, `_`, `\_`)

//go:generate mockgen -source ../condition/condition.go -destination ../condition/mock/mock_condition.go
type Condition interface {
	Convert(ctx context.Context) (string, error)
	Convert2SQL(ctx context.Context) (string, error) // 把condition转成sql的where条件
}

// 将过滤条件拼接到 dsl 请求的 query 部分
func NewCondition(ctx context.Context, cfg *CondCfg, vType string, fieldsMap map[string]*ViewField) (cond Condition, needScore bool, err error) {
	if cfg == nil {
		return nil, false, nil
	}

	// 判断过滤器是否为空对象 {}
	if cfg.Name == "" && cfg.Operation == "" && len(cfg.SubConds) == 0 && cfg.ValueFrom == "" && cfg.Value == nil {
		return nil, false, nil
	}

	switch cfg.Operation {
	case OperationAnd:
		cond, needScore, err = newAndCond(ctx, cfg, vType, fieldsMap)
	case OperationOr:
		cond, needScore, err = newOrCond(ctx, cfg, vType, fieldsMap)
	default:
		cond, needScore, err = NewCondWithOpr(ctx, cfg, vType, fieldsMap)
	}
	if err != nil {
		return nil, needScore, err
	}

	return cond, needScore, nil
}

func NewCondWithOpr(ctx context.Context, cfg *CondCfg, vType string, fieldsMap map[string]*ViewField) (cond Condition, needScore bool, err error) {
	// multi_match之外的才校验
	if cfg.Operation != OperationMultiMatch {
		// 判断除 * 之外的字段
		if cfg.Name != AllField {
			field, ok := fieldsMap[cfg.Name]
			if !ok {
				return nil, needScore, fmt.Errorf("condition config field name '%s' must in view original fields", cfg.Name)
			}

			// 字段类型为空的字段不支持过滤查询
			if field.Type == "" {
				return nil, needScore, fmt.Errorf("condition config field '%s' is empty type, do not support filtering", cfg.Name)
			}

			// 如果字段类型是 binary 类型，则不支持过滤
			if field.Type == dtype.DataType_Binary {
				return nil, needScore, fmt.Errorf("condition config field '%s' is binary type, do not support filtering", cfg.Name)
			}

			cfg.NameField = field
		}
	}

	switch cfg.Operation {
	case OperationEq:
		cond, err = NewEqCond(ctx, cfg, fieldsMap)
	case OperationNotEq:
		cond, err = NewNotEqCond(ctx, cfg, fieldsMap)
	case OperationGt:
		cond, err = NewGtCond(ctx, cfg, fieldsMap)
	case OperationGte:
		cond, err = NewGteCond(ctx, cfg, fieldsMap)
	case OperationLt:
		cond, err = NewLtCond(ctx, cfg, fieldsMap)
	case OperationLte:
		cond, err = NewLteCond(ctx, cfg, fieldsMap)
	case OperationIn:
		cond, err = NewInCond(ctx, cfg, fieldsMap)
	case OperationNotIn:
		cond, err = NewNotInCond(ctx, cfg, fieldsMap)
	case OperationLike:
		cond, err = NewLikeCond(ctx, cfg, fieldsMap)
	case OperationNotLike:
		cond, err = NewNotLikeCond(ctx, cfg, fieldsMap)
	case OperationContain:
		cond, err = NewContainCond(ctx, cfg, fieldsMap)
	case OperationNotContain:
		cond, err = NewNotContainCond(ctx, cfg, fieldsMap)
	case OperationRange:
		cond, err = NewRangeCond(ctx, cfg, fieldsMap)
	case OperationOutRange:
		cond, err = NewOutRangeCond(ctx, cfg, fieldsMap)
	case OperationExist:
		cond, err = NewExistCond(ctx, cfg)
	case OperationNotExist:
		cond, err = NewNotExistCond(ctx, cfg)
	case OperationEmpty:
		cond, err = NewEmptyCond(ctx, cfg, fieldsMap)
	case OperationNotEmpty:
		cond, err = NewNotEmptyCond(ctx, cfg, fieldsMap)
	case OperationRegex:
		cond, err = NewRegexCond(ctx, cfg, fieldsMap)
	case OperationMatch:
		cond, err = NewMatchCond(ctx, cfg, vType, fieldsMap)
		// 如果有全文检索，则需要打分
		needScore = true
	case OperationMatchPhrase:
		cond, err = NewMatchPhraseCond(ctx, cfg, vType, fieldsMap)
		// 如果有全文检索，则需要打分
		needScore = true
	case OperationPrefix:
		cond, err = NewPrefixCond(ctx, cfg, fieldsMap)
	case OperationNotPrefix:
		cond, err = NewNotPrefixCond(ctx, cfg, fieldsMap)
	case OperationNull:
		cond, err = NewNullCond(ctx, cfg, fieldsMap)
	case OperationNotNull:
		cond, err = NewNotNullCond(ctx, cfg, fieldsMap)
	case OperationTrue:
		cond, err = NewTrueCond(ctx, cfg, fieldsMap)
	case OperationFalse:
		cond, err = NewFalseCond(ctx, cfg, fieldsMap)
	case OperationBefore:
		cond, err = NewBeforeCond(ctx, cfg, fieldsMap)
	case OperationCurrent:
		cond, err = NewCurrentCond(ctx, cfg, fieldsMap)
	case OperationBetween:
		cond, err = NewBetweenCond(ctx, cfg, fieldsMap)
	case OperationKnnVector:
		cond, err = NewKnnVectorCond(ctx, cfg, vType, fieldsMap)
		// 如果有knn_vector检索，则需要打分
		needScore = true
	case OperationMultiMatch:
		cond, err = NewMultiMatchCond(ctx, cfg, vType, fieldsMap)
		needScore = true

	default:
		return nil, needScore, fmt.Errorf("not support condition's operation: %s", cfg.Operation)
	}
	if err != nil {
		return nil, needScore, err
	}

	return cond, needScore, nil
}

// 获取过滤条件的字段，name是过滤条件配置的字段，需要将 name 转为 original_name
func GetQueryField(ctx context.Context, name string, fieldsMap map[string]*ViewField, featureType FieldFeatureType) (string, error) {
	// 全文检索允许字段为 "*"
	if name == AllField {
		return name, nil
	}

	fieldInfo, ok1 := fieldsMap[name]
	if ok1 {
		// 无特征，直接用原字段
		if featureType == FieldFeatureType_Raw {
			return fieldInfo.OriginalName, nil
		}

		refName, err := fieldInfo.getTargetField(ctx, featureType)
		if err != nil {
			return "", err
		}
		refFieldInfo, ok2 := fieldsMap[refName]
		if ok2 {
			return refFieldInfo.OriginalName, nil
		} else {
			// 原生特征，使用的是original_name.sub_field_name 比如 original_name.keyword,所以这里返回refName
			return refName, nil
		}
	} else {
		return name, nil
	}
}

// GetTargetField 获取特定查询意图下的实际字段名
func (v *ViewField) getTargetField(ctx context.Context, intent FieldFeatureType) (string, error) {
	// 1. 优先级：优先寻找该意图下已启用的(Enabled)特征
	for _, f := range v.Features {
		if f.FeatureType == intent && f.IsDefault {
			return f.RefField, nil
		}
	}

	// 2. 兜底：若未配置特征但逻辑类型匹配，则直接使用字段名（处理存量或简单数据）
	if TypeToFeatureType(v.Type) == intent {
		return v.Name, nil
	}

	// 全文检索情况下，text 类型的字段不需要添加 keyword 后缀
	// 精确查询情况下，text 类型的字段并且有 keyword 特征，给字段名加上后缀 .keyword
	// 只有查询类型为 DSL 或 IndexBase 才能加 keyword 后缀
	// if IsTermLevelQuery(ctx, intent) &&
	// 	IsTextType(v) && HasFeature(v, FieldFeatureType_Keyword) {
	// 	return wrapKeyWordFieldName(v.Name)
	// }
	// 3. 再次兜底，避免查询报错
	var queryType string
	if ctx.Value(CtxKey_QueryType) != nil {
		queryType = ctx.Value(CtxKey_QueryType).(string)
	}

	switch queryType {
	case QueryType_DSL, QueryType_IndexBase:
		if IsTermLevelQuery(queryType, intent) && IsTextType(v) {
			return wrapKeyWordFieldName(v.Name), nil
		}
	case QueryType_SQL:
		return v.Name, nil
	default:
		return "", fmt.Errorf("field '%s' with type '%s' does not have feature type '%s'", v.Name, v.Type, intent)
	}

	return "", fmt.Errorf("field '%s' with type '%s' does not have feature type '%s'", v.Name, v.Type, intent)
}

// 判断是否是term-level查询
func IsTermLevelQuery(queryType string, intent FieldFeatureType) bool {
	return (queryType == QueryType_DSL || queryType == QueryType_IndexBase) && intent == FieldFeatureType_Keyword
}

// 将类型转换为对应的特征
func TypeToFeatureType(dataType string) FieldFeatureType {
	switch dataType {
	case dtype.DataType_String:
		return FieldFeatureType_Keyword
	case dtype.DataType_Text:
		return FieldFeatureType_Fulltext
	case dtype.DataType_Vector:
		return FieldFeatureType_Vector
	default:
		return ""
	}
}

// 获取过滤条件的字段，name是过滤条件配置的字段，对应name，需要将 name 转为 original_name
// fieldsMap 的key是name
// func getFilterFieldName(ctx context.Context, name string, fieldsMap map[string]*ViewField, isFullTextQuery bool) string {
// 	// 全文检索允许字段为 "*"
// 	if name == AllField {
// 		return name
// 	}

// 	// 如果字段为 __id, 转化为 open search内置字段 _id
// 	// 2025.9.1更新，经过管道的数据里都是包含__id 字段的
// 	// if name == MetaField_ID {
// 	// 	return OS_MetaField_ID
// 	// }

// 	// 如果是脱敏字段，字段添加后缀 _desensitize
// 	desensitizeFieldName := name + DESENSITIZE_FIELD_SUFFIX

// 	fieldInfo, ok1 := fieldsMap[name]
// 	desensitizeFieldInfo, ok2 := fieldsMap[desensitizeFieldName]
// 	if ok1 && ok2 {
// 		// 脱敏字段
// 		// name = desensitizeFieldName
// 		name = desensitizeFieldInfo.OriginalName
// 	} else if ok1 {
// 		// 非脱敏字段
// 		name = fieldInfo.OriginalName
// 	}

// 	// 从 ctx 获取查询类型
// 	var queryType string
// 	if ctx.Value(CtxKey_QueryType) != nil {
// 		queryType = ctx.Value(CtxKey_QueryType).(string)
// 	}

// 	// 全文检索情况下，text 类型的字段不需要添加 keyword 后缀
// 	// 精确查询情况下，text 类型的字段并且有 keyword 特征，给字段名加上后缀 .keyword
// 	// 只有查询类型为 DSL 或 IndexBase 才能加 keyword 后缀
// 	if queryType == QueryType_DSL || queryType == QueryType_IndexBase {
// 		if !isFullTextQuery && IsTextType(fieldInfo) && HasFeature(fieldInfo, FieldFeatureType_Keyword) {
// 			name = wrapKeyWordFieldName(name)
// 		}
// 	}

// 	return name
// }

// 检查字段是否为 text 类型
func IsTextType(fieldInfo *ViewField) bool {
	return fieldInfo != nil && fieldInfo.Type == dtype.DataType_Text
}

// 检查字段特征是否包含指定特征
func HasFeature(fieldInfo *ViewField, feature FieldFeatureType) bool {
	for _, f := range fieldInfo.Features {
		if f.FeatureType == feature {
			return true
		}
	}
	return false
}

// 转换成 keyword
func wrapKeyWordFieldName(fields ...string) string {
	for _, field := range fields {
		if field == "" {
			logger.Warn("missing metric name")
			return ""
		}
	}

	return strings.Join(fields, ".") + "." + dtype.KEYWORD_SUFFIX
}
