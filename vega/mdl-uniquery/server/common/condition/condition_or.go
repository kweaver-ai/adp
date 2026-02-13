// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"
)

type OrCond struct {
	mCfg      *CondCfg
	mSubConds []Condition
}

func newOrCond(ctx context.Context, cfg *CondCfg, vType string, fieldsMap map[string]*ViewField) (cond Condition, needScore bool, err error) {
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

		subConds = append(subConds, cond)
	}

	return &OrCond{
		mCfg:      cfg,
		mSubConds: subConds,
	}, needScore, nil

}

func (cond *OrCond) Convert(ctx context.Context) (string, error) {
	res := `
	{
		"bool": {
			"should": [
				%s
			]
		}
	}
	`

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

func (cond *OrCond) Convert2SQL(ctx context.Context) (string, error) {
	sql := ""
	for i, subCond := range cond.mSubConds {
		where, err := subCond.Convert2SQL(ctx)
		if err != nil {
			return "", err
		}
		where = fmt.Sprintf("(%s)", where)
		if i != len(cond.mSubConds)-1 {
			where += " OR "
		}

		sql += where

	}
	sql = fmt.Sprintf("(%s)", sql)
	return sql, nil
}
