// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package errors

// 指标模型
const (
	// 400
	Uniquery_MetricModel_AggregationNotExisted                      = "Uniquery.MetricModel.AggregationNotExisted"
	Uniquery_MetricModel_AnalysisDimensionNotExisted                = "Uniquery.MetricModel.AnalysisDimensionNotExisted"
	Uniquery_MetricModel_DateFieldNotExisted                        = "Uniquery.MetricModel.DateFieldNotExisted"
	Uniquery_MetricModel_GroupByFieldNotExisted                     = "Uniquery.MetricModel.GroupByFieldNotExisted"
	Uniquery_MetricModel_InvalidParameter                           = "Uniquery.MetricModel.InvalidParameter"
	Uniquery_MetricModel_InvalidParameter_DataSourceType            = "Uniquery.MetricModel.InvalidParameter.DataSourceType"
	Uniquery_MetricModel_InvalidParameter_DerivedCondition          = "Uniquery.MetricModel.InvalidParameter.DerivedCondition"
	Uniquery_MetricModel_InvalidParameter_FieldName                 = "Uniquery.MetricModel.InvalidParameter.FieldName"
	Uniquery_MetricModel_InvalidParameter_FillNull                  = "Uniquery.MetricModel.InvalidParameter.FillNull"
	Uniquery_MetricModel_InvalidParameter_Formula                   = "Uniquery.MetricModel.InvalidParameter.Formula"
	Uniquery_MetricModel_InvalidParameter_FormulaConfig             = "Uniquery.MetricModel.InvalidParameter.FormulaConfig"
	Uniquery_MetricModel_InvalidParameter_HavingConditionName       = "Uniquery.MetricModel.InvalidParameter.HavingConditionName"
	Uniquery_MetricModel_InvalidParameter_IgnoringHCTS              = "Uniquery.MetricModel.InvalidParameter.IgnoringHCTS"
	Uniquery_MetricModel_InvalidParameter_IgnoringMemoryCache       = "Uniquery.MetricModel.InvalidParameter.IgnoringMemoryCache"
	Uniquery_MetricModel_InvalidParameter_IgnoringStoreCache        = "Uniquery.MetricModel.InvalidParameter.IgnoringStoreCache"
	Uniquery_MetricModel_InvalidParameter_IncludeModel              = "Uniquery.MetricModel.InvalidParameter.IncludeModel"
	Uniquery_MetricModel_InvalidParameter_Limit                     = "Uniquery.MetricModel.InvalidParameter.Limit"
	Uniquery_MetricModel_InvalidParameter_LookBackDelta             = "Uniquery.MetricModel.InvalidParameter.LookBackDelta"
	Uniquery_MetricModel_InvalidParameter_MeasureField              = "Uniquery.MetricModel.InvalidParameter.MeasureField"
	Uniquery_MetricModel_InvalidParameter_ModelID                   = "Uniquery.MetricModel.InvalidParameter.ModelID"
	Uniquery_MetricModel_InvalidParameter_Offset                    = "Uniquery.MetricModel.InvalidParameter.Offset"
	Uniquery_MetricModel_InvalidParameter_OrderByDirection          = "Uniquery.MetricModel.InvalidParameter.OrderByDirection"
	Uniquery_MetricModel_InvalidParameter_RequestMetricsType        = "Uniquery.MetricModel.InvalidParameter.RequestMetricsType"
	Uniquery_MetricModel_InvalidParameter_SamePeriodMethod          = "Uniquery.MetricModel.InvalidParameter.SamePeriodMethod"
	Uniquery_MetricModel_InvalidParameter_SamePeriodOffset          = "Uniquery.MetricModel.InvalidParameter.SamePeriodOffset"
	Uniquery_MetricModel_InvalidParameter_SamePeriodTimeGranularity = "Uniquery.MetricModel.InvalidParameter.SamePeriodTimeGranularity"
	Uniquery_MetricModel_InvalidParameter_SQLAggrExpression         = "Uniquery.MetricModel.InvalidParameter.SQLAggrExpression"
	Uniquery_MetricModel_InvalidParameter_SQLCondition              = "Uniquery.MetricModel.InvalidParameter.SQLCondition"
	Uniquery_MetricModel_InvalidParameter_SqlConfig                 = "Uniquery.MetricModel.InvalidParameter.SqlConfig"
	Uniquery_MetricModel_InvalidParameter_Step                      = "Uniquery.MetricModel.InvalidParameter.Step"
	Uniquery_MetricModel_InvalidParameter_Time                      = "Uniquery.MetricModel.InvalidParameter.Time"
	Uniquery_MetricModel_NullParameter_DateField                    = "Uniquery.MetricModel.NullParameter.DateField"
	Uniquery_MetricModel_NullParameter_DataSource                   = "Uniquery.MetricModel.NullParameter.DataSource"
	Uniquery_MetricModel_NullParameter_DataSourceID                 = "Uniquery.MetricModel.NullParameter.DataSourceID"
	Uniquery_MetricModel_NullParameter_DependMetricModel            = "Uniquery.MetricModel.NullParameter.DependMetricModel"
	Uniquery_MetricModel_NullParameter_DerivedCondition             = "Uniquery.MetricModel.NullParameter.DerivedCondition"
	Uniquery_MetricModel_NullParameter_Formula                      = "Uniquery.MetricModel.NullParameter.Formula"
	Uniquery_MetricModel_NullParameter_FormulaConfig                = "Uniquery.MetricModel.NullParameter.FormulaConfig"
	Uniquery_MetricModel_NullParameter_MeasureField                 = "Uniquery.MetricModel.NullParameter.MeasureField"
	Uniquery_MetricModel_NullParameter_MetricType                   = "Uniquery.MetricModel.NullParameter.MetricType"
	Uniquery_MetricModel_NullParameter_HavingConditionOperation     = "Uniquery.MetricModel.NullParameter.HavingConditionOperation"
	Uniquery_MetricModel_NullParameter_OrderByDirection             = "Uniquery.MetricModel.NullParameter.OrderByDirection"
	Uniquery_MetricModel_NullParameter_OrderByName                  = "Uniquery.MetricModel.NullParameter.OrderByName"
	Uniquery_MetricModel_NullParameter_QueryType                    = "Uniquery.MetricModel.NullParameter.QueryType"
	Uniquery_MetricModel_NullParameter_RequestMetricsType           = "Uniquery.MetricModel.NullParameter.RequestMetricsType"
	Uniquery_MetricModel_NullParameter_SamePeriodCfg                = "Uniquery.MetricModel.NullParameter.SamePeriodCfg"
	Uniquery_MetricModel_NullParameter_SamePeriodMethod             = "Uniquery.MetricModel.NullParameter.SamePeriodMethod"
	Uniquery_MetricModel_NullParameter_SQLAggrExpression            = "Uniquery.MetricModel.NullParameter.SQLAggrExpression"
	Uniquery_MetricModel_NullParameter_Step                         = "Uniquery.MetricModel.NullParameter.Step"
	Uniquery_MetricModel_UnsupportDataSourceType                    = "Uniquery.MetricModel.UnsupportDataSourceType"
	Uniquery_MetricModel_UnsupportMetricType                        = "Uniquery.MetricModel.UnsupportMetricType"
	Uniquery_MetricModel_UnsupportQuery                             = "Uniquery.MetricModel.UnsupportQuery"
	Uniquery_MetricModel_UnsupportQueryType                         = "Uniquery.MetricModel.UnsupportQueryType"
	Uniquery_MetricModel_UnsupportHavingConditionOperation          = "Uniquery.MetricModel.UnsupportHavingConditionOperation"

	//404
	Uniquery_MetricModel_DataViewNotFound    = "Uniquery.MetricModel.DataViewNotFound"
	Uniquery_MetricModel_MetricModelNotFound = "Uniquery.MetricModel.MetricModelNotFound"

	// 500
	Uniquery_MetricModel_InternalError                               = "Uniquery.MetricModel.InternalError"
	Uniquery_MetricModel_InternalError_ExecPromQLFailed              = "Uniquery.MetricModel.InternalError.ExecPromQLFailed"
	Uniquery_MetricModel_InternalError_FetchDatasFromVegaFailed      = "Uniquery.MetricModel.InternalError.FetchDatasFromVegaFailed"
	Uniquery_MetricModel_InternalError_GetFieldsFailed               = "Uniquery.MetricModel.InternalError.GetFieldsFailed"
	Uniquery_MetricModel_InternalError_GetFieldValuesFailed          = "Uniquery.MetricModel.InternalError.GetFieldValuesFailed"
	Uniquery_MetricModel_InternalError_GetLabelsFailed               = "Uniquery.MetricModel.InternalError.GetLabelsFailed"
	Uniquery_MetricModel_InternalError_GetDataViewQueryFiltersFailed = "Uniquery.MetricModel.InternalError.GetDataViewQueryFiltersFailed"
	Uniquery_MetricModel_InternalError_GetModelByIdFailed            = "Uniquery.MetricModel.InternalError.GetMetricModelByIdFailed"
	Uniquery_MetricModel_InternalError_GetModelIdByNameFailed        = "Uniquery.MetricModel.InternalError.GetMetricModelIdByNameFailed"
	Uniquery_MetricModel_InternalError_GetVegaViewFieldsByIDFailed   = "Uniquery.MetricModel.InternalError.GetVegaViewFieldsByIDFailed"
	Uniquery_MetricModel_InternalError_MarshalFailed                 = "Uniquery.MetricModel.InternalError.MarshalFailed"
	Uniquery_MetricModel_InternalError_ParseResultFailed             = "Uniquery.MetricModel.InternalError.ParseResultFailed"
	Uniquery_MetricModel_InternalError_ParseUpdateTimeFailed         = "Uniquery.MetricModel.InternalError.ParseUpdateTimeFailed"
	Uniquery_MetricModel_InternalError_UnmarshalFailed               = "Uniquery.MetricModel.InternalError.UnmarshalFailed"
	Uniquery_MetricModel_InternalError_UnSupportPromQLResult         = "Uniquery.MetricModel.InternalError.UnSupportPromQLResult"
	Uniquery_MetricModel_InvaliEnvironmentVariable_TZ                = "Uniquery.MetricModel.InvaliEnvironmentVariable.Timezone"
)

var (
	metricModelErrCodeList = []string{
		// 400
		Uniquery_MetricModel_AggregationNotExisted,
		Uniquery_MetricModel_AnalysisDimensionNotExisted,
		Uniquery_MetricModel_DateFieldNotExisted,
		Uniquery_MetricModel_GroupByFieldNotExisted,
		Uniquery_MetricModel_InvalidParameter,
		Uniquery_MetricModel_InvalidParameter_DataSourceType,
		Uniquery_MetricModel_InvalidParameter_DerivedCondition,
		Uniquery_MetricModel_InvalidParameter_FieldName,
		Uniquery_MetricModel_InvalidParameter_FillNull,
		Uniquery_MetricModel_InvalidParameter_Formula,
		Uniquery_MetricModel_InvalidParameter_FormulaConfig,
		Uniquery_MetricModel_InvalidParameter_IgnoringHCTS,
		Uniquery_MetricModel_InvalidParameter_IgnoringStoreCache,
		Uniquery_MetricModel_InvalidParameter_IgnoringMemoryCache,
		Uniquery_MetricModel_InvalidParameter_IncludeModel,
		Uniquery_MetricModel_InvalidParameter_Limit,
		Uniquery_MetricModel_InvalidParameter_LookBackDelta,
		Uniquery_MetricModel_InvalidParameter_MeasureField,
		Uniquery_MetricModel_InvalidParameter_ModelID,
		Uniquery_MetricModel_InvalidParameter_Offset,
		Uniquery_MetricModel_InvalidParameter_OrderByDirection,
		Uniquery_MetricModel_InvalidParameter_RequestMetricsType,
		Uniquery_MetricModel_InvalidParameter_SamePeriodMethod,
		Uniquery_MetricModel_InvalidParameter_SamePeriodOffset,
		Uniquery_MetricModel_InvalidParameter_SamePeriodTimeGranularity,
		Uniquery_MetricModel_InvalidParameter_SQLAggrExpression,
		Uniquery_MetricModel_InvalidParameter_SQLCondition,
		Uniquery_MetricModel_InvalidParameter_SqlConfig,
		Uniquery_MetricModel_InvalidParameter_Step,
		Uniquery_MetricModel_InvalidParameter_Time,
		Uniquery_MetricModel_NullParameter_DateField,
		Uniquery_MetricModel_NullParameter_DataSource,
		Uniquery_MetricModel_NullParameter_DataSourceID,
		Uniquery_MetricModel_NullParameter_DependMetricModel,
		Uniquery_MetricModel_NullParameter_DerivedCondition,
		Uniquery_MetricModel_NullParameter_Formula,
		Uniquery_MetricModel_NullParameter_FormulaConfig,
		Uniquery_MetricModel_NullParameter_MeasureField,
		Uniquery_MetricModel_NullParameter_MetricType,
		Uniquery_MetricModel_InvalidParameter_HavingConditionName,
		Uniquery_MetricModel_NullParameter_HavingConditionOperation,
		Uniquery_MetricModel_NullParameter_OrderByDirection,
		Uniquery_MetricModel_NullParameter_OrderByName,
		Uniquery_MetricModel_NullParameter_QueryType,
		Uniquery_MetricModel_NullParameter_RequestMetricsType,
		Uniquery_MetricModel_NullParameter_SamePeriodCfg,
		Uniquery_MetricModel_NullParameter_SamePeriodMethod,
		Uniquery_MetricModel_NullParameter_SQLAggrExpression,
		Uniquery_MetricModel_NullParameter_Step,
		Uniquery_MetricModel_UnsupportDataSourceType,
		Uniquery_MetricModel_UnsupportMetricType,
		Uniquery_MetricModel_UnsupportQuery,
		Uniquery_MetricModel_UnsupportQueryType,
		Uniquery_MetricModel_UnsupportHavingConditionOperation,

		// 404
		Uniquery_MetricModel_DataViewNotFound,
		Uniquery_MetricModel_MetricModelNotFound,

		// 500
		Uniquery_MetricModel_InternalError,
		Uniquery_MetricModel_InternalError_ExecPromQLFailed,
		Uniquery_MetricModel_InternalError_FetchDatasFromVegaFailed,
		Uniquery_MetricModel_InternalError_GetFieldsFailed,
		Uniquery_MetricModel_InternalError_GetFieldValuesFailed,
		Uniquery_MetricModel_InternalError_GetLabelsFailed,
		Uniquery_MetricModel_InternalError_GetDataViewQueryFiltersFailed,
		Uniquery_MetricModel_InternalError_GetModelByIdFailed,
		Uniquery_MetricModel_InternalError_GetModelIdByNameFailed,
		Uniquery_MetricModel_InternalError_GetVegaViewFieldsByIDFailed,
		Uniquery_MetricModel_InternalError_MarshalFailed,
		Uniquery_MetricModel_InternalError_ParseResultFailed,
		Uniquery_MetricModel_InternalError_ParseUpdateTimeFailed,
		Uniquery_MetricModel_InternalError_UnmarshalFailed,
		Uniquery_MetricModel_InternalError_UnSupportPromQLResult,
		Uniquery_MetricModel_InvaliEnvironmentVariable_TZ,
	}
)
