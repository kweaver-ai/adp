// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"

	cond "uniquery/common/condition"
)

// 行列规则结构体
type DataViewRowColumnRule struct {
	RuleID     string        `json:"id"`
	RuleName   string        `json:"name"`
	ViewID     string        `json:"view_id"`
	Tags       []string      `json:"tags"`
	Comment    string        `json:"comment"`
	Fields     []string      `json:"fields"`
	RowFilters *cond.CondCfg `json:"row_filters"`
	// CreateTime int64         `json:"create_time"`
	// UpdateTime int64         `json:"update_time"`
	// Creator    string        `json:"creator"`
	// Updater    string        `json:"updater"`

	// 操作权限
	Operations []string `json:"operations"`
}

type ListRowColumnRulesResult struct {
	Entries    []*DataViewRowColumnRule `json:"entries"`
	TotalCount int                      `json:"total_count"`
}

//go:generate mockgen -source ../interfaces/data_view_row_column_rules_access.go -destination ../interfaces/mock/mock_data_view_row_column_rules_access.go
type DataViewRowColumnRuleAccess interface {
	GetRulesByViewID(ctx context.Context, viewID string) ([]*DataViewRowColumnRule, error)
}
