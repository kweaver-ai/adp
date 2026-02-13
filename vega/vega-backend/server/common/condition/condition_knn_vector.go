// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"

	"vega-backend/interfaces"
	dtype "vega-backend/interfaces/data_type"
	vopt "vega-backend/interfaces/value_opt"
)

type KnnVectorCond struct {
	mCfg             *CondCfg
	mFilterFieldName string
	mSubConds        []Condition
}

// knn_vector 条件, 判断字段是否匹配某个向量
func NewKnnVectorCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*interfaces.Property) (Condition, error) {
	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [in] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [in] left field '%s' not found", cfg.Name)
	}
	if field.Type != dtype.DataType_Vector {
		return nil, fmt.Errorf("condition [knn_vector] left field '%s' type must be vector", cfg.Name)
	}

	if cfg.ValueOptCfg.ValueFrom != vopt.ValueFrom_Const {
		return nil, fmt.Errorf("condition [knn_vector] does not support value_from type '%s'", cfg.ValueFrom)
	}

	subConds := []Condition{}
	for _, subCond := range cfg.SubConds {
		cond, err := NewCondition(ctx, subCond, fieldsMap)
		if err != nil {
			return nil, err
		}

		if cond != nil {
			subConds = append(subConds, cond)
		}

	}

	return &KnnVectorCond{
		mCfg:      cfg,
		mSubConds: subConds,
	}, nil
}
