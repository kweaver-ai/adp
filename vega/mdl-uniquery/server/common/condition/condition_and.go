package condition

import (
	"context"
	"fmt"
)

type AndCond struct {
	mCfg      *CondCfg
	mSubConds []Condition
	needScore bool
}

func newAndCond(ctx context.Context, cfg *CondCfg, vType string, fieldsMap map[string]*ViewField) (cond Condition, needScore bool, err error) {
	subConds := []Condition{}

	if len(cfg.SubConds) == 0 {
		return nil, false, fmt.Errorf("sub condition size is 0")
	}

	if len(cfg.SubConds) > MaxSubCondition {
		return nil, false, fmt.Errorf("sub condition size limit %d but %d", MaxSubCondition, len(cfg.SubConds))
	}

	for _, subCond := range cfg.SubConds {
		cond, needScore, err = NewCondition(ctx, subCond, vType, fieldsMap)
		if err != nil {
			return nil, needScore, err
		}

		if cond != nil {
			subConds = append(subConds, cond)
		}

	}

	return &AndCond{
		mCfg:      cfg,
		mSubConds: subConds,
		needScore: needScore,
	}, needScore, nil

}

func (cond *AndCond) Convert(ctx context.Context) (string, error) {
	var res string
	if cond.needScore {
		res = `
	{
		"bool": {
			"must": [
				%s
			]
		}
	}
	`
	} else {
		res = `
	{
		"bool": {
			"filter": [
				%s
			]
		}
	}
	`
	}

	dslStr := ""
	for i, subCond := range cond.mSubConds {
		dsl, err := subCond.Convert(ctx)
		if err != nil {
			return "", err
		}

		if i != len(cond.mSubConds)-1 {
			dsl += ","
		}

		dslStr += dsl

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
