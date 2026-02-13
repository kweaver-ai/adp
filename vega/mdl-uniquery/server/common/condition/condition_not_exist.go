// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"
)

type NotExistCond struct {
	mCfg       *CondCfg
	mfieldName string
}

func NewNotExistCond(ctx context.Context, cfg *CondCfg) (Condition, error) {
	return &NotExistCond{
		mCfg:       cfg,
		mfieldName: cfg.Name,
	}, nil
}

func (cond *NotExistCond) Convert(ctx context.Context) (string, error) {
	dslStr := `
	{
		"bool": {
			"must_not": [
				{
					"exists": {
						"field": "%s"
					}
				}
			]
		}
	}
	`

	return fmt.Sprintf(dslStr, cond.mfieldName), nil
}

func (cond *NotExistCond) Convert2SQL(ctx context.Context) (string, error) {
	sqlStr := fmt.Sprintf(`"%s" IS NULL`, cond.mfieldName)
	return sqlStr, nil
}
