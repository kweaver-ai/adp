// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"
	"strings"

	"vega-backend/interfaces"
)

const MaxSubCondition = 100

// sql的字符串转义
var Special = strings.NewReplacer(`\`, `\\\\`, `'`, `\'`, `%`, `\%`, `_`, `\_`)

//go:generate mockgen -source ../condition/condition.go -destination ../condition/mock/mock_condition.go
type Condition interface {
}

// 将过滤条件拼接到 dsl 请求的 query 部分
func NewCondition(ctx context.Context, cfg *CondCfg,
	fieldsMap map[string]*interfaces.Property) (cond Condition, err error) {

	if cfg == nil {
		return nil, nil
	}

	// 判断过滤器是否为空对象 {}
	if cfg.Name == "" && cfg.Operation == "" && len(cfg.SubConds) == 0 && cfg.ValueFrom == "" && cfg.Value == nil {
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

func NewCondWithOpr(ctx context.Context, cfg *CondCfg,
	fieldsMap map[string]*interfaces.Property) (cond Condition, err error) {

	switch cfg.Operation {
	case OperationEq, OperationEq2:
		cond, err = NewEqCond(ctx, cfg, fieldsMap)
	case OperationNotEq, OperationNotEq2:
		cond, err = NewNotEqCond(ctx, cfg, fieldsMap)
	case OperationGt, OperationGt2:
		cond, err = NewGtCond(ctx, cfg, fieldsMap)
	case OperationGte, OperationGte2:
		cond, err = NewGteCond(ctx, cfg, fieldsMap)
	case OperationLt, OperationLt2:
		cond, err = NewLtCond(ctx, cfg, fieldsMap)
	case OperationLte, OperationLte2:
		cond, err = NewLteCond(ctx, cfg, fieldsMap)
	case OperationIn:
		cond, err = NewInCond(ctx, cfg, fieldsMap)
	case OperationNotIn:
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
		cond, err = NewExistCond(ctx, cfg, fieldsMap)
	case OperationNotExist:
		cond, err = NewNotExistCond(ctx, cfg, fieldsMap)
	case OperationEmpty:
		cond, err = NewEmptyCond(ctx, cfg, fieldsMap)
	case OperationNotEmpty:
		cond, err = NewNotEmptyCond(ctx, cfg, fieldsMap)
	case OperationRegex:
		cond, err = NewRegexCond(ctx, cfg, fieldsMap)
	case OperationMatch:
		cond, err = NewMatchCond(ctx, cfg, fieldsMap)
	case OperationMatchPhrase:
		cond, err = NewMatchPhraseCond(ctx, cfg, fieldsMap)
	case OperationPrefix:
		cond, err = NewPrefixCond(ctx, cfg, fieldsMap)
	case OperationNotPrefix:
		cond, err = NewNotPrefixCond(ctx, cfg, fieldsMap)
	case OperationNull:
		cond, err = NewNullCond(ctx, cfg, fieldsMap)
	case OperationNotNull:
		cond, err = NewNotNullCond(ctx, cfg, fieldsMap)
	case OperationTrue:
		cond, err = NewTrueCond(ctx, cfg, fieldsMap)
	case OperationFalse:
		cond, err = NewFalseCond(ctx, cfg, fieldsMap)
	case OperationBefore:
		cond, err = NewBeforeCond(ctx, cfg, fieldsMap)
	case OperationCurrent:
		cond, err = NewCurrentCond(ctx, cfg, fieldsMap)
	case OperationBetween:
		cond, err = NewBetweenCond(ctx, cfg, fieldsMap)
	case OperationKnnVector:
		cond, err = NewKnnVectorCond(ctx, cfg, fieldsMap)
	case OperationMultiMatch:
		cond, err = NewMultiMatchCond(ctx, cfg, fieldsMap)

	default:
		return nil, fmt.Errorf("not support condition's operation: %s", cfg.Operation)
	}
	if err != nil {
		return nil, err
	}

	return cond, nil
}
