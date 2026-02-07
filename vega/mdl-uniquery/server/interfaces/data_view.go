// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"time"
	cond "uniquery/common/condition"
)

// 泛型接口
type ViewQueryInterface interface {
	GetCommonParams() *ViewQueryCommonParams
	GetGlobalFilters() *cond.CondCfg
	GetSortParams() []*SortParamsV2
	GetScroll() *Scroll
	GetQueryParams() map[string]any
	GetNeedTotal() bool
	GetSearchAfterParams() *SearchAfterParams
	GetVegaDuration() int64
	GetRowColumnRules() []*DataViewRowColumnRule

	SetFormat(format string)
	SetNeedTotal(needTotal bool)
	SetLimit(limit int)
	SetVegaDuration(dur int64)
	SetRowColumnRules(rowColumnRules []*DataViewRowColumnRule)
}

func (dvsq *DataViewSimulateQuery) GetCommonParams() *ViewQueryCommonParams {
	return &dvsq.ViewQueryCommonParams
}

func (dvq *DataViewQueryV1) GetCommonParams() *ViewQueryCommonParams {
	return &dvq.ViewQueryCommonParams
}

func (dvqV2 *DataViewQueryV2) GetCommonParams() *ViewQueryCommonParams {
	return &dvqV2.ViewQueryCommonParams
}

func (dvsq *DataViewSimulateQuery) GetGlobalFilters() *cond.CondCfg {
	return nil
}

func (dvq *DataViewQueryV1) GetGlobalFilters() *cond.CondCfg {
	return dvq.GlobalFilters
}

func (dvqV2 *DataViewQueryV2) GetGlobalFilters() *cond.CondCfg {
	return dvqV2.ActualCondition
}

func (dvsq *DataViewSimulateQuery) GetSortParams() []*SortParamsV2 {
	return dvsq.Sort
}

func (dvq *DataViewQueryV1) GetSortParams() []*SortParamsV2 {
	if dvq.Sort == "" {
		return []*SortParamsV2{}
	}

	return []*SortParamsV2{
		{
			Field:     dvq.Sort,
			Direction: dvq.Direction,
		},
	}
}

func (dvqV2 *DataViewQueryV2) GetSortParams() []*SortParamsV2 {
	return dvqV2.Sort
}

func (dvsq *DataViewSimulateQuery) GetRowColumnRules() []*DataViewRowColumnRule {
	return dvsq.RowColumnRules
}

func (dvq *DataViewQueryV1) GetRowColumnRules() []*DataViewRowColumnRule {
	return dvq.RowColumnRules
}

func (dvqV2 *DataViewQueryV2) GetRowColumnRules() []*DataViewRowColumnRule {
	return dvqV2.RowColumnRules
}

func (dvsq *DataViewSimulateQuery) SetRowColumnRules(rowColumnRules []*DataViewRowColumnRule) {
	dvsq.RowColumnRules = rowColumnRules
}

func (dvq *DataViewQueryV1) SetRowColumnRules(rowColumnRules []*DataViewRowColumnRule) {
	dvq.RowColumnRules = rowColumnRules
}

func (dvqV2 *DataViewQueryV2) SetRowColumnRules(rowColumnRules []*DataViewRowColumnRule) {
	dvqV2.RowColumnRules = rowColumnRules
}

func (dvsq *DataViewSimulateQuery) SetFormat(format string) {
	dvsq.Format = format
}

func (dvq *DataViewQueryV1) SetFormat(format string) {
	dvq.Format = format
}

func (dvqV2 *DataViewQueryV2) SetFormat(format string) {
	dvqV2.Format = format
}

func (dvsq *DataViewSimulateQuery) SetNeedTotal(needTotal bool) {
	dvsq.NeedTotal = needTotal
}

func (dvq *DataViewQueryV1) SetNeedTotal(needTotal bool) {
	dvq.NeedTotal = needTotal
}

func (dvqV2 *DataViewQueryV2) SetNeedTotal(needTotal bool) {
	dvqV2.NeedTotal = needTotal
}

func (dvsq *DataViewSimulateQuery) SetLimit(limit int) {
	dvsq.Limit = limit
}

func (dvq *DataViewQueryV1) SetLimit(limit int) {
	dvq.Limit = limit
}

func (dvqV2 *DataViewQueryV2) SetLimit(limit int) {
	dvqV2.Limit = limit
}

func (dvsq *DataViewSimulateQuery) SetVegaDuration(dur int64) {
	dvsq.VegaDurationMs = dur
}

func (dvq *DataViewQueryV1) SetVegaDuration(dur int64) {
	dvq.VegaDurationMs = dur
}

func (dvqV2 *DataViewQueryV2) SetVegaDuration(dur int64) {
	dvqV2.VegaDurationMs = dur
}

func (dvsq *DataViewSimulateQuery) GetVegaDuration() int64 {
	return dvsq.VegaDurationMs
}

func (dvq *DataViewQueryV1) GetVegaDuration() int64 {
	return dvq.VegaDurationMs
}

func (dvqV2 *DataViewQueryV2) GetVegaDuration() int64 {
	return dvqV2.VegaDurationMs
}

func (dvsq *DataViewSimulateQuery) GetQueryParams() map[string]any {
	return map[string]any{
		QueryParam_AllowNonExistField: false,
		QueryParam_IncludeView:        true,
		QueryParam_Timeout:            time.Duration(0),
	}
}

func (dvq *DataViewQueryV1) GetQueryParams() map[string]any {
	return map[string]any{
		QueryParam_AllowNonExistField: dvq.AllowNonExistField,
		QueryParam_IncludeView:        true,
		QueryParam_Timeout:            time.Duration(0),
	}
}

func (dvqV2 *DataViewQueryV2) GetQueryParams() map[string]any {
	return map[string]any{
		QueryParam_AllowNonExistField: dvqV2.AllowNonExistField,
		QueryParam_IncludeView:        dvqV2.IncludeView,
		QueryParam_Timeout:            dvqV2.Timeout,
	}
}

func (dvsq *DataViewSimulateQuery) GetScroll() *Scroll {
	return &Scroll{}
}

func (dvq *DataViewQueryV1) GetScroll() *Scroll {
	return &Scroll{
		Scroll:   dvq.Scroll,
		ScrollId: dvq.ScrollId,
	}
}

func (dvqV2 *DataViewQueryV2) GetScroll() *Scroll {
	return &Scroll{}
}

func (dvsq *DataViewSimulateQuery) GetNeedTotal() bool {
	return true
}

func (dvq *DataViewQueryV1) GetNeedTotal() bool {
	return dvq.NeedTotal
}

func (dvqV2 *DataViewQueryV2) GetNeedTotal() bool {
	return dvqV2.NeedTotal
}

func (dvsq *DataViewSimulateQuery) GetSearchAfterParams() *SearchAfterParams {
	return nil
}

func (dvq *DataViewQueryV1) GetSearchAfterParams() *SearchAfterParams {
	return nil
}

func (dvqV2 *DataViewQueryV2) GetSearchAfterParams() *SearchAfterParams {
	return &dvqV2.SearchAfterParams
}

type SortParamsV1 struct {
	Sort      string `json:"sort"`
	Direction string `json:"direction"`
}

type SortParamsV2 struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

// 视图查询公共参数
type ViewQueryCommonParams struct {
	Start          int64                    `json:"start"`
	End            int64                    `json:"end"`
	DateField      string                   `json:"date_field"`
	Offset         int                      `json:"offset"`
	Limit          int                      `json:"limit"`
	Format         string                   `json:"format"`
	NeedTotal      bool                     `json:"need_total"`
	UseSearchAfter bool                     `json:"use_search_after"`
	SqlStr         string                   `json:"sql"`
	RowColumnRules []*DataViewRowColumnRule `json:"row_column_rules"`
	OutputFields   []string                 `json:"output_fields"` // 指定输出的字段列表
}

// 预览查询结构体
type DataViewSimulateQuery struct {
	ViewQueryCommonParams
	Sort           []*SortParamsV2   `json:"sort"`
	Type           string            `json:"type" binding:"required,oneof=atomic custom"`
	QueryType      string            `json:"query_type" binding:"required,oneof=SQL DSL IndexBase"`
	TechnicalName  string            `json:"technical_name"`
	DataSourceType string            `json:"data_source_type"`
	DataSourceID   string            `json:"data_source_id"`
	FileName       string            `json:"file_name"`
	ExcelConfig    *ExcelConfig      `json:"excel_config"`
	DataScope      []*DataScopeNode  `json:"data_scope"`
	Fields         []*cond.ViewField `json:"fields"`
	VegaDurationMs int64             `json:"-"`
}

// 视图数据查询请求体
type DataViewQueryV1 struct {
	SortParamsV1
	ViewQueryCommonParams
	GlobalFilters      *cond.CondCfg `json:"filters"` // 仪表盘使用的全局过滤条件
	AllowNonExistField bool          `json:"-"`       // 控制参数，如果过滤条件的字段不在视图字段里，是否抛出异常，true则返回空数据，false则抛出异常

	Scroll   string `json:"scroll"`
	ScrollId string `json:"scroll_id"`

	VegaDurationMs int64 `json:"-"`
}

// 视图数据查询请求体v2
type DataViewQueryV2 struct {
	AllowNonExistField bool `json:"-"`
	IncludeView        bool `json:"-"` // 控制是否返回视图对象，查询参数
	// GlobalFilters      *cond.CondCfg   `json:"filters"`
	GlobalFilters map[string]any  `json:"filters"`
	Sort          []*SortParamsV2 `json:"sort"`
	Timeout       time.Duration   `json:"-"` // 超时时间，查询参数
	ViewQueryCommonParams
	SearchAfterParams

	ActualCondition *cond.CondCfg `json:"-"`
	VegaDurationMs  int64         `json:"-"`
}

type SearchAfterParams struct {
	SearchAfter  []any  `json:"search_after"`
	PitID        string `json:"pit_id"`
	PitKeepAlive string `json:"pit_keep_alive"`
}
