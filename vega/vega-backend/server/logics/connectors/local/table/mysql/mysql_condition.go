// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package mysql

import (
	"context"
	"fmt"

	"vega-backend/interfaces"
	"vega-backend/logics/filter_condition"

	sq "github.com/Masterminds/squirrel"
)

func ConvertFilterCondition(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	switch condition.GetOperation() {
	case filter_condition.OperationAnd:
		return ConvertFilterConditionAnd(ctx, condition, fieldsMap)

	case filter_condition.OperationOr:
		return ConvertFilterConditionOr(ctx, condition, fieldsMap)

	default:
		return ConvertFilterConditionWithOpr(ctx, condition, fieldsMap)
	}
}

func ConvertFilterConditionAnd(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	condAnd, ok := condition.(*filter_condition.AndCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.AndCond")
	}

	convertedConds := sq.And{}
	for _, subCond := range condAnd.SubConds {
		convertedCond, err := ConvertFilterConditionWithOpr(ctx, subCond, fieldsMap)
		if err != nil {
			return nil, err
		}
		convertedConds = append(convertedConds, convertedCond)
	}

	return convertedConds, nil
}

func ConvertFilterConditionOr(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	condOr, ok := condition.(*filter_condition.OrCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.OrCond")
	}

	convertedConds := sq.Or{}
	for _, subCond := range condOr.SubConds {
		convertedCond, err := ConvertFilterConditionWithOpr(ctx, subCond, fieldsMap)
		if err != nil {
			return nil, err
		}
		convertedConds = append(convertedConds, convertedCond)
	}

	return convertedConds, nil
}

func ConvertFilterConditionWithOpr(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	switch condition.GetOperation() {
	case filter_condition.OperationEqual:
		return ConvertFilterConditionEqual(ctx, condition, fieldsMap)
	case filter_condition.OperationNotEqual:
		return ConvertFilterConditionNotEqual(ctx, condition, fieldsMap)

	default:
		return nil, fmt.Errorf("operation %s is not supported", condition.GetOperation())
	}
}

func ConvertFilterConditionEqual(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.EqualCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.EqualCond")
	}

	switch cond.Cfg.ValueFrom {
	case interfaces.ValueFrom_Const:
		return sq.Eq{fmt.Sprintf("`%s`", cond.Field.OriginalName): cond.Value}, nil
	case interfaces.ValueFrom_Field:
		return sq.Expr(fmt.Sprintf("`%s` = `%s`", cond.Field.OriginalName, cond.Value)), nil
	default:
		return nil, fmt.Errorf("value_from %s is not supported", cond.Cfg.ValueFrom)
	}
}

func ConvertFilterConditionNotEqual(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.NotEqualCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NotEqualCond")
	}

	switch cond.Cfg.ValueFrom {
	case interfaces.ValueFrom_Const:
		return sq.NotEq{fmt.Sprintf("`%s`", cond.Field.OriginalName): cond.Value}, nil
	case interfaces.ValueFrom_Field:
		return sq.Expr(fmt.Sprintf("`%s` <> `%s`", cond.Field.OriginalName, cond.Value)), nil
	default:
		return nil, fmt.Errorf("value_from %s is not supported", cond.Cfg.ValueFrom)
	}
}
