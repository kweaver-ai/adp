package condition

import (
	"context"
	"fmt"
)

type AndCond struct {
	mCfg      *CondCfg
	mSubConds []Condition
}

func newAndCond(ctx context.Context, cfg *CondCfg, fieldScope uint8, fieldsMap map[string]*ViewField) (Condition, error) {
	subConds := []Condition{}

	if len(cfg.SubConds) == 0 {
		return nil, fmt.Errorf("sub condition size is 0")
	}

	if len(cfg.SubConds) > MaxSubCondition {
		return nil, fmt.Errorf("sub condition size limit %d but %d", MaxSubCondition, len(cfg.SubConds))
	}

	for _, subCond := range cfg.SubConds {
		cond, err := NewCondition(ctx, subCond, fieldScope, fieldsMap)
		if err != nil {
			return nil, err
		}

		if cond != nil {
			subConds = append(subConds, cond)
		}

	}

	return &AndCond{
		mCfg:      cfg,
		mSubConds: subConds,
	}, nil

}

func (cond *AndCond) Convert(ctx context.Context, vectorizer func(ctx context.Context, words []string) ([]*VectorResp, error)) (string, error) {
	res := `
	{
		"bool": {
			"must": [
				%s
			]
		}
	}
	`

	dslStr := ""
	validDSLs := []string{}
	for _, subCond := range cond.mSubConds {
		dsl, err := subCond.Convert(ctx, vectorizer)
		if err != nil {
			return "", err
		}

		// 过滤掉空字符串（被忽略的条件）
		if dsl != "" && dsl != "{}" {
			validDSLs = append(validDSLs, dsl)
		}
	}

	// 如果所有子条件都被过滤掉，返回空对象
	if len(validDSLs) == 0 {
		return "{}", nil
	}

	// 如果只有一个有效子条件，直接返回该子条件的 DSL，不需要包装在 bool.must 中
	if len(validDSLs) == 1 {
		return validDSLs[0], nil
	}

	// 多个有效子条件，用逗号连接
	for i, dsl := range validDSLs {
		if i != len(validDSLs)-1 {
			dslStr += dsl + ","
		} else {
			dslStr += dsl
		}
	}

	res = fmt.Sprintf(res, dslStr)
	return res, nil

}

func (cond *AndCond) Convert2SQL(ctx context.Context) (string, error) {
	sql := ""
	for i, subCond := range cond.mSubConds {
		where, err := subCond.Convert2SQL(ctx)
		if err != nil {
			return "", err
		}

		if i != len(cond.mSubConds)-1 {
			where += " AND "
		}

		sql += where

	}
	return sql, nil
}
