// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"database/sql"
)

// 行列规则结构体
type DataViewRowColumnRule struct {
	RuleID     string      `json:"id"`
	RuleName   string      `json:"name"`
	ViewID     string      `json:"view_id"`
	ViewName   string      `json:"view_name,omitempty"`
	Tags       []string    `json:"tags"`
	Comment    string      `json:"comment"`
	CreateTime int64       `json:"create_time"`
	UpdateTime int64       `json:"update_time"`
	Creator    AccountInfo `json:"creator"`
	Updater    AccountInfo `json:"updater"`
	Fields     []string    `json:"fields"`
	RowFilters *CondCfg    `json:"row_filters"`

	// 操作权限
	Operations []string `json:"operations,omitempty"`
}

type ListRowColumnRuleQueryParams struct {
	Name           string
	NamePattern    string
	ViewID         string
	Tag            string
	IsInnerRequest bool
	PaginationQueryParameters
}

//go:generate mockgen -source ../interfaces/data_view_row_column_rule_access.go -destination ../interfaces/mock/mock_data_view_row_column_rule_access.go
type DataViewRowColumnRuleAccess interface {
	CreateDataViewRowColumnRules(ctx context.Context, rules []*DataViewRowColumnRule) error
	UpdateDataViewRowColumnRule(ctx context.Context, rule *DataViewRowColumnRule) error
	GetDataViewRowColumnRules(ctx context.Context, ruleIDs []string) ([]*DataViewRowColumnRule, error)
	ListDataViewRowColumnRules(ctx context.Context, query *ListRowColumnRuleQueryParams) ([]*DataViewRowColumnRule, error)
	DeleteDataViewRowColumnRules(ctx context.Context, tx *sql.Tx, ruleIDs []string) error

	CheckDataViewRowColumnRuleExistByID(ctx context.Context, ruleID string) (string, bool, error)
	CheckDataViewRowColumnRuleExistByName(ctx context.Context, ruleName, viewID string) (string, bool, error)
	GetSimpleRulesByViewIDs(ctx context.Context, tx *sql.Tx, viewIDs []string) ([]*DataViewRowColumnRule, error)
	GetSimpleRulesByRuleIDs(ctx context.Context, tx *sql.Tx, ruleIDs []string) ([]*DataViewRowColumnRule, error)
}
