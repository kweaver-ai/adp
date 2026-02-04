package condition

import (
	"context"
	"fmt"

	"github.com/dlclark/regexp2"

	vopt "uniquery/common/value_opt"
	dtype "uniquery/interfaces/data_type"
)

type RegexCond struct {
	mCfg             *CondCfg
	mValue           string
	mRegexp          *regexp2.Regexp
	mFilterFieldName string
}

func NewRegexCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*ViewField) (Condition, error) {
	if !dtype.DataType_IsString(cfg.NameField.Type) &&
		dtype.SimpleTypeMapping[cfg.NameField.Type] != dtype.DataType_String {
		return nil, fmt.Errorf("condition [regex] left field is not a string field: %s:%s", cfg.NameField.Name, cfg.NameField.Type)
	}

	if cfg.ValueOptCfg.ValueFrom != vopt.ValueFrom_Const {
		return nil, fmt.Errorf("condition [regex] does not support value_from type '%s'", cfg.ValueFrom)
	}

	val, ok := cfg.ValueOptCfg.Value.(string)
	if !ok {
		return nil, fmt.Errorf("condition [regex] right value is not a string value: %v", cfg.Value)
	}

	regexp, err := regexp2.Compile(val, regexp2.RE2)
	if err != nil {
		return nil, fmt.Errorf("condition [regex] regular expression error: %s", err.Error())
	}

	featureType := FieldFeatureType_Raw
	if IsTextType(fieldsMap[cfg.Name]) {
		featureType = FieldFeatureType_Keyword
	}

	fName, err := GetQueryField(ctx, cfg.Name, fieldsMap, featureType)
	if err != nil {
		return nil, fmt.Errorf("condition [regex], %v", err)
	}

	return &RegexCond{
		mCfg:             cfg,
		mValue:           val,
		mRegexp:          regexp,
		mFilterFieldName: fName,
	}, nil
}

func (cond *RegexCond) Convert(ctx context.Context) (string, error) {
	dslStr := fmt.Sprintf(`
					{
						"regexp": {
							"%s": "%v"
						}
					}`, cond.mFilterFieldName, cond.mValue)

	return dslStr, nil
}

// REGEXP_LIKE(column, 'pattern')
func (cond *RegexCond) Convert2SQL(ctx context.Context) (string, error) {
	sqlStr := fmt.Sprintf(`REGEXP_LIKE("%s", '%s')`, cond.mFilterFieldName, cond.mValue)

	return sqlStr, nil
}
