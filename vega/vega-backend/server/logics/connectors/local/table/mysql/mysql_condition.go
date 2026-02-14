// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package mysql

import (
	"context"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"

	"vega-backend/interfaces"
	"vega-backend/logics/filter_condition"
)

var Special = strings.NewReplacer(`\`, `\\\\`, `'`, `\'`, `%`, `\%`, `_`, `\_`)

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
	case filter_condition.OperationEqual, filter_condition.OperationEqual2:
		return ConvertFilterConditionEqual(ctx, condition, fieldsMap)
	case filter_condition.OperationNotEqual, filter_condition.OperationNotEqual2:
		return ConvertFilterConditionNotEqual(ctx, condition, fieldsMap)
	case filter_condition.OperationGt, filter_condition.OperationGt2:
		return ConvertFilterConditionGt(ctx, condition, fieldsMap)
	case filter_condition.OperationGte, filter_condition.OperationGte2:
		return ConvertFilterConditionGte(ctx, condition, fieldsMap)
	case filter_condition.OperationLt, filter_condition.OperationLt2:
		return ConvertFilterConditionLt(ctx, condition, fieldsMap)
	case filter_condition.OperationLte, filter_condition.OperationLte2:
		return ConvertFilterConditionLte(ctx, condition, fieldsMap)
	case filter_condition.OperationIn:
		return ConvertFilterConditionIn(ctx, condition, fieldsMap)
	case filter_condition.OperationNotIn:
		return ConvertFilterConditionNotIn(ctx, condition, fieldsMap)
	case filter_condition.OperationLike:
		return ConvertFilterConditionLike(ctx, condition, fieldsMap)
	case filter_condition.OperationNotLike:
		return ConvertFilterConditionNotLike(ctx, condition, fieldsMap)
	case filter_condition.OperationContain:
		return ConvertFilterConditionContain(ctx, condition, fieldsMap)
	case filter_condition.OperationNotContain:
		return ConvertFilterConditionNotContain(ctx, condition, fieldsMap)
	case filter_condition.OperationRange:
		return ConvertFilterConditionRange(ctx, condition, fieldsMap)
	case filter_condition.OperationOutRange:
		return ConvertFilterConditionOutRange(ctx, condition, fieldsMap)
	case filter_condition.OperationNull:
		return ConvertFilterConditionNull(ctx, condition, fieldsMap)
	case filter_condition.OperationNotNull:
		return ConvertFilterConditionNotNull(ctx, condition, fieldsMap)
	case filter_condition.OperationEmpty:
		return ConvertFilterConditionEmpty(ctx, condition, fieldsMap)
	case filter_condition.OperationNotEmpty:
		return ConvertFilterConditionNotEmpty(ctx, condition, fieldsMap)
	case filter_condition.OperationPrefix:
		return ConvertFilterConditionPrefix(ctx, condition, fieldsMap)
	case filter_condition.OperationNotPrefix:
		return ConvertFilterConditionNotPrefix(ctx, condition, fieldsMap)
	case filter_condition.OperationBetween:
		return ConvertFilterConditionBetween(ctx, condition, fieldsMap)
	case filter_condition.OperationExist:
		return ConvertFilterConditionExist(ctx, condition, fieldsMap)
	case filter_condition.OperationNotExist:
		return ConvertFilterConditionNotExist(ctx, condition, fieldsMap)
	case filter_condition.OperationRegex:
		return ConvertFilterConditionRegex(ctx, condition, fieldsMap)
	case filter_condition.OperationTrue:
		return ConvertFilterConditionTrue(ctx, condition, fieldsMap)
	case filter_condition.OperationFalse:
		return ConvertFilterConditionFalse(ctx, condition, fieldsMap)
	case filter_condition.OperationBefore:
		return ConvertFilterConditionBefore(ctx, condition, fieldsMap)
	case filter_condition.OperationCurrent:
		return ConvertFilterConditionCurrent(ctx, condition, fieldsMap)
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
		return sq.Eq{fmt.Sprintf("`%s`", cond.Lfield.OriginalName): cond.Value}, nil
	case interfaces.ValueFrom_Field:
		return sq.Expr(fmt.Sprintf("`%s` = `%s`", cond.Lfield.OriginalName, cond.Rfield.OriginalName)), nil
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
		return sq.NotEq{fmt.Sprintf("`%s`", cond.Lfield.OriginalName): cond.Value}, nil
	case interfaces.ValueFrom_Field:
		return sq.Expr(fmt.Sprintf("`%s` <> `%s`", cond.Lfield.OriginalName, cond.Rfield.OriginalName)), nil
	default:
		return nil, fmt.Errorf("value_from %s is not supported", cond.Cfg.ValueFrom)
	}
}

func ConvertFilterConditionGt(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.GtCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.GtCond")
	}

	switch cond.Cfg.ValueFrom {
	case interfaces.ValueFrom_Const:
		return sq.Gt{fmt.Sprintf("`%s`", cond.Lfield.OriginalName): cond.Value}, nil
	case interfaces.ValueFrom_Field:
		return sq.Expr(fmt.Sprintf("`%s` > `%s`", cond.Lfield.OriginalName, cond.Rfield.OriginalName)), nil
	default:
		return nil, fmt.Errorf("value_from %s is not supported", cond.Cfg.ValueFrom)
	}
}

func ConvertFilterConditionGte(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.GteCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.GteCond")
	}

	switch cond.Cfg.ValueFrom {
	case interfaces.ValueFrom_Const:
		return sq.GtOrEq{fmt.Sprintf("`%s`", cond.Lfield.OriginalName): cond.Value}, nil
	case interfaces.ValueFrom_Field:
		return sq.Expr(fmt.Sprintf("`%s` >= `%s`", cond.Lfield.OriginalName, cond.Rfield.OriginalName)), nil
	default:
		return nil, fmt.Errorf("value_from %s is not supported", cond.Cfg.ValueFrom)
	}
}

func ConvertFilterConditionLt(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.LtCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.LtCond")
	}

	switch cond.Cfg.ValueFrom {
	case interfaces.ValueFrom_Const:
		return sq.Lt{fmt.Sprintf("`%s`", cond.Lfield.OriginalName): cond.Value}, nil
	case interfaces.ValueFrom_Field:
		return sq.Expr(fmt.Sprintf("`%s` < `%s`", cond.Lfield.OriginalName, cond.Rfield.OriginalName)), nil
	default:
		return nil, fmt.Errorf("value_from %s is not supported", cond.Cfg.ValueFrom)
	}
}

func ConvertFilterConditionLte(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.LteCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.LteCond")
	}

	switch cond.Cfg.ValueFrom {
	case interfaces.ValueFrom_Const:
		return sq.LtOrEq{fmt.Sprintf("`%s`", cond.Lfield.OriginalName): cond.Value}, nil
	case interfaces.ValueFrom_Field:
		return sq.Expr(fmt.Sprintf("`%s` <= `%s`", cond.Lfield.OriginalName, cond.Rfield.OriginalName)), nil
	default:
		return nil, fmt.Errorf("value_from %s is not supported", cond.Cfg.ValueFrom)
	}
}

func ConvertFilterConditionIn(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.InCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.InCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [in] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	return sq.Eq{fmt.Sprintf("`%s`", cond.Lfield.OriginalName): cond.Value}, nil
}

func ConvertFilterConditionNotIn(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.NotInCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NotInCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [not_in] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	return sq.NotEq{fmt.Sprintf("`%s`", cond.Lfield.OriginalName): cond.Value}, nil
}

func ConvertFilterConditionLike(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.LikeCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.LikeCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [like] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	vStr := "%" + Special.Replace(cond.Value) + "%"
	return sq.Like{fmt.Sprintf("`%s`", cond.Lfield.OriginalName): vStr}, nil
}

func ConvertFilterConditionNotLike(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.NotLikeCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NotLikeCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [not_like] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	vStr := "%" + Special.Replace(cond.Value) + "%"
	return sq.NotLike{fmt.Sprintf("`%s`", cond.Lfield.OriginalName): vStr}, nil
}

func ConvertFilterConditionContain(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.ContainCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.ContainCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [contain] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	values := cond.Value
	exprs := make(sq.And, len(values))
	for i, v := range values {
		exprs[i] = sq.Expr(fmt.Sprintf("FIND_IN_SET(?, `%s`) > 0", cond.Lfield.OriginalName), v)
	}
	return exprs, nil
}

func ConvertFilterConditionNotContain(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.NotContainCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NotContainCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [not_contain] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	values := cond.Value
	exprs := make(sq.Or, len(values))
	for i, v := range values {
		exprs[i] = sq.Expr(fmt.Sprintf("FIND_IN_SET(?, `%s`) = 0", cond.Lfield.OriginalName), v)
	}
	return exprs, nil
}

func ConvertFilterConditionRange(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.RangeCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.RangeCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [range] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	values := cond.Value
	if len(values) != 2 {
		return nil, fmt.Errorf("range condition requires exactly 2 values")
	}

	return sq.And{
		sq.GtOrEq{fmt.Sprintf("`%s`", cond.Lfield.OriginalName): values[0]},
		sq.LtOrEq{fmt.Sprintf("`%s`", cond.Lfield.OriginalName): values[1]},
	}, nil
}

func ConvertFilterConditionOutRange(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.OutRangeCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.OutRangeCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [out_range] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	values := cond.Value
	if len(values) != 2 {
		return nil, fmt.Errorf("out_range condition requires exactly 2 values")
	}

	return sq.Or{
		sq.Lt{fmt.Sprintf("`%s`", cond.Lfield.OriginalName): values[0]},
		sq.Gt{fmt.Sprintf("`%s`", cond.Lfield.OriginalName): values[1]},
	}, nil
}

func ConvertFilterConditionNull(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.NullCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NullCond")
	}

	return sq.Eq{fmt.Sprintf("`%s`", cond.Lfield.OriginalName): nil}, nil
}

func ConvertFilterConditionNotNull(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.NotNullCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NotNullCond")
	}

	return sq.NotEq{fmt.Sprintf("`%s`", cond.Lfield.OriginalName): nil}, nil
}

func ConvertFilterConditionEmpty(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.EmptyCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.EmptyCond")
	}

	return sq.Eq{fmt.Sprintf("`%s`", cond.Lfield.OriginalName): ""}, nil
}

func ConvertFilterConditionNotEmpty(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.NotEmptyCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NotEmptyCond")
	}

	return sq.NotEq{fmt.Sprintf("`%s`", cond.Lfield.OriginalName): ""}, nil
}

func ConvertFilterConditionPrefix(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.PrefixCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.PrefixCond")
	}

	vStr := Special.Replace(cond.Value) + "%"
	return sq.Like{fmt.Sprintf("`%s`", cond.Lfield.OriginalName): vStr}, nil
}

func ConvertFilterConditionNotPrefix(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.NotPrefixCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NotPrefixCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [not_prefix] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	vStr := Special.Replace(cond.Value) + "%"
	return sq.NotLike{fmt.Sprintf("`%s`", cond.Lfield.OriginalName): vStr}, nil
}

func ConvertFilterConditionBetween(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.BetweenCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.BetweenCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [between] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	values := cond.Value
	if len(values) != 2 {
		return nil, fmt.Errorf("between condition requires exactly 2 values")
	}

	return sq.And{
		sq.GtOrEq{fmt.Sprintf("`%s`", cond.Lfield.OriginalName): values[0]},
		sq.LtOrEq{fmt.Sprintf("`%s`", cond.Lfield.OriginalName): values[1]},
	}, nil
}

func ConvertFilterConditionExist(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.ExistCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.ExistCond")
	}

	return sq.NotEq{fmt.Sprintf("`%s`", cond.Lfield.OriginalName): nil}, nil
}

func ConvertFilterConditionNotExist(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.NotExistCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NotExistCond")
	}

	return sq.Eq{fmt.Sprintf("`%s`", cond.Lfield.OriginalName): nil}, nil
}

func ConvertFilterConditionRegex(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.RegexCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.RegexCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [regex] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	return sq.Expr(fmt.Sprintf("`%s` REGEXP ?", cond.Lfield.OriginalName), cond.Value), nil
}

func ConvertFilterConditionTrue(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.TrueCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.TrueCond")
	}

	return sq.Eq{fmt.Sprintf("`%s`", cond.Lfield.OriginalName): true}, nil
}

func ConvertFilterConditionFalse(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.FalseCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.FalseCond")
	}

	return sq.Eq{fmt.Sprintf("`%s`", cond.Lfield.OriginalName): false}, nil
}

func ConvertFilterConditionBefore(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.BeforeCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.BeforeCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [before] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	values := cond.Value
	if len(values) != 2 {
		return nil, fmt.Errorf("before condition requires exactly 2 values")
	}

	interval, ok := values[0].(float64)
	if !ok {
		return nil, fmt.Errorf("condition [before] interval value should be a number")
	}
	unit, ok := values[1].(string)
	if !ok {
		return nil, fmt.Errorf("condition [before] unit value should be a string")
	}

	return sq.Expr(fmt.Sprintf("`%s` < DATE_SUB(NOW(), INTERVAL ? %s)", cond.Lfield.OriginalName, unit), int(interval)), nil
}

func ConvertFilterConditionCurrent(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (sq.Sqlizer, error) {

	cond, ok := condition.(*filter_condition.CurrentCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.CurrentCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [current] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	var dateFormat string
	switch cond.Value {
	case filter_condition.CurrentYear:
		dateFormat = "%Y"
	case filter_condition.CurrentMonth:
		dateFormat = "%Y-%m"
	case filter_condition.CurrentWeek:
		dateFormat = "%Y-%u"
	case filter_condition.CurrentDay:
		dateFormat = "%Y-%m-%d"
	case filter_condition.CurrentHour:
		dateFormat = "%Y-%m-%d %H"
	case filter_condition.CurrentMinute:
		dateFormat = "%Y-%m-%d %H:%i"
	default:
		return nil, fmt.Errorf("condition [current] unsupported format: %s", cond.Value)
	}

	return sq.Expr(fmt.Sprintf("DATE_FORMAT(`%s`, '%s') = DATE_FORMAT(NOW(), '%s')", cond.Lfield.OriginalName, dateFormat, dateFormat)), nil
}
