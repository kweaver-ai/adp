// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"
	"vega-backend/interfaces"
)

type OrCond struct {
	mCfg      *CondCfg
	mSubConds []Condition
}

func newOrCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*interfaces.Property) (cond Condition, err error) {
	subConds := []Condition{}

	if len(cfg.SubConds) == 0 {
		return nil, fmt.Errorf("sub condition size is 0")
	}

	if len(cfg.SubConds) > MaxSubCondition {
		return nil, fmt.Errorf("sub condition size limit %d but %d", MaxSubCondition, len(cfg.SubConds))
	}

	for _, subCond := range cfg.SubConds {
		cond, err = NewCondition(ctx, subCond, fieldsMap)
		if err != nil {
			return nil, err
		}

		subConds = append(subConds, cond)
	}

	return &OrCond{
		mCfg:      cfg,
		mSubConds: subConds,
	}, nil
}
