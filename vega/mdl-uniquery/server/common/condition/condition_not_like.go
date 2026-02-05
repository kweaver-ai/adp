package condition

import (
	"context"
	"fmt"

	vopt "uniquery/common/value_opt"
	dtype "uniquery/interfaces/data_type"
)

type NotLikeCond struct {
	mCfg             *CondCfg
	mValue           string
	mRealValue       string
	mFilterFieldName string
}

func NewNotLikeCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*ViewField) (Condition, error) {
	if !dtype.DataType_IsString(cfg.NameField.Type) &&
		dtype.SimpleTypeMapping[cfg.NameField.Type] != dtype.DataType_String {
		return nil, fmt.Errorf("condition [not_like] left field is not a string field: %s:%s", cfg.NameField.Name, cfg.NameField.Type)
	}

	if cfg.ValueOptCfg.ValueFrom != vopt.ValueFrom_Const {
		return nil, fmt.Errorf("condition [not_like] does not support value_from type '%s'", cfg.ValueFrom)
	}

	// 如果有 real_value 则跳过 value 的校验
	var val, realVal string
	if cfg.ValueOptCfg.RealValue != nil {
		var ok bool
		realVal, ok = cfg.ValueOptCfg.RealValue.(string)
		if !ok {
			return nil, fmt.Errorf("condition [not_like] right real value is not a string value: %v", cfg.RealValue)
		}
	} else {
		var ok bool
		val, ok = cfg.ValueOptCfg.Value.(string)
		if !ok {
			return nil, fmt.Errorf("condition [not_like] right value is not a string value: %v", cfg.Value)
		}
	}

	featureType := FieldFeatureType_Raw
	if IsTextType(fieldsMap[cfg.Name]) {
		featureType = FieldFeatureType_Keyword
	}

	fName, err := GetQueryField(ctx, cfg.Name, fieldsMap, featureType)
	if err != nil {
		return nil, fmt.Errorf("condition [not_like], %v", err)
	}

	return &NotLikeCond{
		mCfg:             cfg,
		mValue:           val,
		mRealValue:       realVal,
		mFilterFieldName: fName,
	}, nil
}

func (cond *NotLikeCond) Convert(ctx context.Context) (string, error) {
	valPattern := fmt.Sprintf(".*%s.*", cond.mCfg.Value)
	v := fmt.Sprintf("%q", valPattern)

	dslStr := fmt.Sprintf(`
					{
						"bool": {
							"must_not": [
								{
									"regexp": {
										"%s": %v
									}
								}
							]
						}
					}`, cond.mFilterFieldName, v)

	return dslStr, nil
}

func (cond *NotLikeCond) Convert2SQL(ctx context.Context) (string, error) {
	// real_value: 内部接口调用，值已拼接好 %，支持自定义前缀/后缀匹配
	// value: 前端传入，不带 %，后端自动转义特殊字符并添加 %value% 通配符
	var vStr string
	if cond.mRealValue != "" {
		vStr = cond.mRealValue
	} else {
		vStr = "%" + Special.Replace(cond.mValue) + "%"
	}

	sqlStr := fmt.Sprintf(`"%s" NOT LIKE '%s'`, cond.mFilterFieldName, vStr)
	return sqlStr, nil
}
