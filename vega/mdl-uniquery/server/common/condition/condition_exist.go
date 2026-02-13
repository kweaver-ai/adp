// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"
)

type ExistCond struct {
	mCfg       *CondCfg
	mfieldName string
}

func NewExistCond(ctx context.Context, cfg *CondCfg) (Condition, error) {
	return &ExistCond{
		mCfg:       cfg,
		mfieldName: cfg.Name,
	}, nil
}

func (cond *ExistCond) Convert(ctx context.Context) (string, error) {
	dslStr := `
	{
		"exists": {
			"field": "%s"
		}
	}
	`

	return fmt.Sprintf(dslStr, cond.mfieldName), nil
}

// sql中没有字段存在的过滤条件,暂时用非空表达
func (cond *ExistCond) Convert2SQL(ctx context.Context) (string, error) {
	return fmt.Sprintf(`"%s" IS NOT NULL`, cond.mfieldName), nil
}
