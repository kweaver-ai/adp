// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
)

type MetricModelQuery struct {
	MetricType      string            `json:"metric_type"`
	DataSource      *MetricDataSource `json:"data_source"` // 计算公式有效性检查时，用id就可以少请求一次
	QueryType       string            `json:"query_type"`
	Formula         string            `json:"formula"`
	FormulaConfig   any               `json:"formula_config,omitempty"`
	AnalysisDims    []string          `json:"analysis_dimensions,omitempty"`
	OrderByFields   []OrderField      `json:"order_by_fields,omitempty"`
	HavingCondition *CondCfg          `json:"having_condition,omitempty"`
	DateField       string            `json:"date_field"`
	MeasureField    string            `json:"measure_field"`
	IsInstantQuery  bool              `json:"instant"`
	// Start          int64             `json:"start"`
	// End            int64             `json:"end"`
	Step           string `json:"step"`
	LookBackDelta  string `json:"look_back_delta"`
	IsModelRequest bool   `json:"is_model_request"` // 是否是来自data-model的请求.data-model的请求是做计算公式有效性检查的, 其不需要再次校验dsl公式中的规则的请求，此时无需校验 dsl 公式是否符合规则，且获取序列时只获取一个序列,且用深度优先的方式收据,预览时需要校验。
}

type DataViewSql struct {
	SqlStr string `json:"sql_str"`
}

//go:generate mockgen -source ../interfaces/uniquery_access.go -destination ../interfaces/mock/mock_uniquery_access.go
type UniqueryAccess interface {
	// 指标模型
	CheckFormulaByUniquery(ctx context.Context, query MetricModelQuery) (bool, string, error)

	// 数据视图
	BuildDataViewSql(ctx context.Context, view *DataView) (string, error)
}
