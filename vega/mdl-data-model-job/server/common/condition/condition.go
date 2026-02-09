// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"
)

const MaxSubCondition = 10

type Condition interface {
	// 判断数据是否通过过滤条件
	Pass(ctx context.Context, data *OriginalData) (bool, error)
}

func NewCondition(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*Field) (cond Condition, err error) {
	if cfg == nil {
		return nil, nil
	}
	switch cfg.Operation {
	case OperationAnd:
		cond, err = newAndCond(ctx, cfg, fieldsMap)
	case OperationOr:
		cond, err = newOrCond(ctx, cfg, fieldsMap)
	default:
		cond, err = NewCondWithOpr(ctx, cfg, fieldsMap)
	}
	if err != nil {
		return nil, err
	}

	return cond, nil
}

func NewCondWithOpr(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*Field) (cond Condition, err error) {
	// 如果全局过滤条件的字段不在视图字段列表里，数据返回空
	field, ok := fieldsMap[cfg.Name]
	if cfg.Name != DefaultField && !ok {
		return nil, fmt.Errorf("condition config field name must in view original fields")
	}

	cfg.NameField = field

	switch cfg.Operation {
	case OperationEq:
		cond, err = NewEqCond(ctx, cfg, fieldsMap)
	case OperationNotEq:
		cond, err = NewNotEqCond(ctx, cfg, fieldsMap)
	case OperationGt:
		cond, err = NewGtCond(ctx, cfg, fieldsMap)
	case OperationGte:
		cond, err = NewGteCond(ctx, cfg, fieldsMap)
	case OperationLt:
		cond, err = NewLtCond(ctx, cfg, fieldsMap)
	case OperationLte:
		cond, err = NewLteCond(ctx, cfg, fieldsMap)
	case OperationIn:
		cond, err = NewInCond(ctx, cfg, fieldsMap)
	case OperarionNotIn:
		cond, err = NewNotInCond(ctx, cfg, fieldsMap)
	case OperationLike:
		cond, err = NewLikeCond(ctx, cfg, fieldsMap)
	case OperationNotLike:
		cond, err = NewNotLikeCond(ctx, cfg, fieldsMap)
	case OperationContain:
		cond, err = NewContainCond(ctx, cfg, fieldsMap)
	case OperationNotContain:
		cond, err = NewNotContainCond(ctx, cfg, fieldsMap)
	case OperationRange:
		cond, err = NewRangeCond(ctx, cfg, fieldsMap)
	case OperationOutRange:
		cond, err = NewOutRangeCond(ctx, cfg, fieldsMap)
	case OperationExist:
		cond, err = NewExistCond(ctx, cfg)
	case OperationNotExist:
		cond, err = NewNotExistCond(ctx, cfg)
	case OperationEmpty:
		cond, err = NewEmptyCond(ctx, cfg, fieldsMap)
	case OperationNotEmpty:
		cond, err = NewNotEmptyCond(ctx, cfg, fieldsMap)
	case OperationRegex:
		cond, err = NewRegexCond(ctx, cfg, fieldsMap)
	default:
		return nil, fmt.Errorf("not support condition's operation: %s", cfg.Operation)
	}
	if err != nil {
		return nil, err
	}

	return cond, nil
}
