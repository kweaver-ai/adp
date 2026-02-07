// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"database/sql"
)

//go:generate mockgen -source ../interfaces/data_view_row_column_rule_service.go -destination ../interfaces/mock/mock_data_view_row_column_rule_service.go
type DataViewRowColumnRuleService interface {
	CreateDataViewRowColumnRules(ctx context.Context, rules []*DataViewRowColumnRule) ([]string, error)
	UpdateDataViewRowColumnRule(ctx context.Context, rule *DataViewRowColumnRule) error
	GetDataViewRowColumnRules(ctx context.Context, ruleIDs []string) ([]*DataViewRowColumnRule, error)
	ListDataViewRowColumnRules(ctx context.Context, query *ListRowColumnRuleQueryParams) ([]*DataViewRowColumnRule, int, error)
	// GetDataViewRowColumnRulesTotal(ctx context.Context, query *ListRowColumnRuleQueryParams) (int, error)
	DeleteDataViewRowColumnRules(ctx context.Context, ruleIDs []string) error

	// 校验数据视图行列规则是否存在
	CheckDataViewRowColumnRuleExistByID(ctx context.Context, ruleID string) (string, error)

	// 获取行列规则的资源实例列表
	ListDataViewRowColumnRuleSrcs(ctx context.Context, params *ListRowColumnRuleQueryParams) ([]*Resource, int, error)
	DeleteRowColumnRulesByViewIDs(ctx context.Context, tx *sql.Tx, viewIDs []string) error
}
