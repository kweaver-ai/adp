// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"
)

type AndCond struct {
	mCfg      *CondCfg
	mSubConds []Condition
}

func newAndCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*Field) (Condition, error) {
	subConds := []Condition{}

	if len(cfg.SubConds) == 0 {
		return nil, fmt.Errorf("sub condition size is 0")
	}

	if len(cfg.SubConds) > MaxSubCondition {
		return nil, fmt.Errorf("sub condition size limit %d but %d", MaxSubCondition, len(cfg.SubConds))
	}

	for _, subCond := range cfg.SubConds {
		cond, err := NewCondition(ctx, subCond, fieldsMap)
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

func (cond *AndCond) Pass(ctx context.Context, data *OriginalData) (bool, error) {
	for _, subCond := range cond.mSubConds {
		rslt, err := subCond.Pass(ctx, data)
		if err != nil {
			return false, err
		}
		if !rslt {
			return false, nil
		}
	}
	return true, nil
}
