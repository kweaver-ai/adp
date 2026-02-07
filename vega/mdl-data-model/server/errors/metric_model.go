// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package errors

// 指标模型错误码
const (
	// 400
	DataModel_MetricModel_AggregationNotExisted                  = "DataModel.MetricModel.AggregationNotExisted"
	DataModel_MetricModel_AnalysisDimensionNotExisted            = "DataModel.MetricModel.AnalysisDimensionNotExisted"
	DataModel_MetricModel_CombinationNameExisted                 = "DataModel.MetricModel.CombinationNameExisted"
	DataModel_MetricModel_OrderByFieldNotExisted                 = "DataModel.MetricModel.OrderByFieldNotExisted"
	DataModel_MetricModel_CountExceeded_TagTotal                 = "DataModel.MetricModel.CountExceeded.TagTotal"
	DataModel_MetricModel_CountExceeded_TaskTotal                = "DataModel.MetricModel.CountExceeded.TaskTotal"
	DataModel_MetricModel_DateFieldNotExisted                    = "DataModel.MetricModel.DateFieldNotExisted"
	DataModel_MetricModel_Duplicated_ModelIDInFile               = "DataModel.MetricModel.Duplicated.ModelIDInFile"
	DataModel_MetricModel_Duplicated_CombinationName             = "DataModel.MetricModel.Duplicated.CombinationName"
	DataModel_MetricModel_Duplicated_IndexBase                   = "DataModel.MetricModel.Duplicated.IndexBase"
	DataModel_MetricModel_Duplicated_MeasureName                 = "DataModel.MetricModel.Duplicated.MeasureName"
	DataModel_MetricModel_Duplicated_ModelName                   = "DataModel.MetricModel.Duplicated.ModelName"
	DataModel_MetricModel_Duplicated_TaskName                    = "DataModel.MetricModel.Duplicated.TaskName"
	DataModel_MetricModel_Duplicated_TaskStep                    = "DataModel.MetricModel.Duplicated.TaskStep"
	DataModel_MetricModel_Duplicated_TimeWindows                 = "DataModel.MetricModel.Duplicated.TimeWindows"
	DataModel_MetricModel_GroupByFieldNotExisted                 = "DataModel.MetricModel.GroupByFieldNotExisted"
	DataModel_MetricModel_IDExisted                              = "DataModel.MetricModel.IDExisted"
	DataModel_MetricModel_InvalidParameter                       = "DataModel.MetricModel.InvalidParameter"
	DataModel_MetricModel_InvalidParameter_DateField             = "DataModel.MetricModel.InvalidParameter.DateField"
	DataModel_MetricModel_InvalidParameter_DateFormat            = "DataModel.MetricModel.InvalidParameter.DateFormat"
	DataModel_MetricModel_InvalidParameter_DataSourceType        = "DataModel.MetricModel.InvalidParameter.DataSourceType"
	DataModel_MetricModel_InvalidParameter_DerivedCondition      = "DataModel.MetricModel.InvalidParameter.DerivedCondition"
	DataModel_MetricModel_InvalidParameter_Formula               = "DataModel.MetricModel.InvalidParameter.Formula"
	DataModel_MetricModel_InvalidParameter_FormulaConfig         = "DataModel.MetricModel.InvalidParameter.FormulaConfig"
	DataModel_MetricModel_InvalidParameter_Group                 = "DataModel.MetricModel.InvalidParameter.Group"
	DataModel_MetricModel_InvalidParameter_HavingConditionName   = "DataModel.MetricModel.InvalidParameter.HavingConditionName"
	DataModel_MetricModel_InvalidParameter_IncludeView           = "DataModel.MetricModel.InvalidParameter.IncludeView"
	DataModel_MetricModel_InvalidParameter_MeasureField          = "DataModel.MetricModel.InvalidParameter.MeasureField"
	DataModel_MetricModel_InvalidParameter_MeasureFieldNotExists = "DataModel.MetricModel.InvalidParameter.MeasureFieldNotExists"
	DataModel_MetricModel_InvalidParameter_MeasureName           = "DataModel.MetricModel.InvalidParameter.MeasureName"
	DataModel_MetricModel_InvalidParameter_ModelID               = "DataModel.MetricModel.InvalidParameter.ModelID"
	DataModel_MetricModel_InvalidParameter_OrderByDirection      = "DataModel.MetricModel.InvalidParameter.OrderByDirection"
	DataModel_MetricModel_InvalidParameter_RetraceDuration       = "DataModel.MetricModel.InvalidParameter.RetraceDuration"
	DataModel_MetricModel_InvalidParameter_ScheduleExpression    = "DataModel.MetricModel.InvalidParameter.ScheduleExpression"
	DataModel_MetricModel_InvalidParameter_ScheduleType          = "DataModel.MetricModel.InvalidParameter.ScheduleType"
	DataModel_MetricModel_InvalidParameter_SQLAggrExpression     = "DataModel.MetricModel.InvalidParameter.SQLAggrExpression"
	DataModel_MetricModel_InvalidParameter_SQLCondition          = "DataModel.MetricModel.InvalidParameter.SQLCondition"
	DataModel_MetricModel_InvalidParameter_Step                  = "DataModel.MetricModel.InvalidParameter.Step"
	DataModel_MetricModel_InvalidParameter_TimeWindows           = "DataModel.MetricModel.InvalidParameter.TimeWindows"
	DataModel_MetricModel_InvalidParameter_UnitType              = "DataModel.MetricModel.InvalidParameter.UnitType"
	DataModel_MetricModel_LengthExceeded_MeasureName             = "DataModel.MetricModel.LengthExceeded.MeasureName"
	DataModel_MetricModel_LengthExceeded_ModelName               = "DataModel.MetricModel.LengthExceeded.ModelName"
	DataModel_MetricModel_LengthExceeded_TaskComment             = "DataModel.MetricModel.LengthExceeded.TaskComment"
	DataModel_MetricModel_LengthExceeded_TaskName                = "DataModel.MetricModel.LengthExceeded.TaskName"
	DataModel_MetricModel_ModelNameExisted                       = "DataModel.MetricModel.ModelNameExisted"
	DataModel_MetricModel_NullParameter_DependMetricModel        = "DataModel.MetricModel.NullParameter.DependMetricModel"
	DataModel_MetricModel_NullParameter_DateField                = "DataModel.MetricModel.NullParameter.DateField"
	DataModel_MetricModel_NullParameter_DataSource               = "DataModel.MetricModel.NullParameter.DataSource"
	DataModel_MetricModel_NullParameter_DataSourceID             = "DataModel.MetricModel.NullParameter.DataSourceID"
	DataModel_MetricModel_NullParameter_DerivedCondition         = "DataModel.MetricModel.NullParameter.DerivedCondition"
	// DataModel_MetricModel_NullParameter_DataView                 = "DataModel.MetricModel.NullParameter.DataView"
	DataModel_MetricModel_NullParameter_Formula                  = "DataModel.MetricModel.NullParameter.Formula"
	DataModel_MetricModel_NullParameter_FormulaConfig            = "DataModel.MetricModel.NullParameter.FormulaConfig"
	DataModel_MetricModel_NullParameter_HavingConditionOperation = "DataModel.MetricModel.NullParameter.HavingConditionOperation"
	DataModel_MetricModel_NullParameter_IndexBase                = "DataModel.MetricModel.NullParameter.IndexBase"
	DataModel_MetricModel_NullParameter_MeasureField             = "DataModel.MetricModel.NullParameter.MeasureField"
	DataModel_MetricModel_NullParameter_MeasureName              = "DataModel.MetricModel.NullParameter.MeasureName"
	DataModel_MetricModel_NullParameter_MetricType               = "DataModel.MetricModel.NullParameter.MetricType"
	DataModel_MetricModel_NullParameter_ModelName                = "DataModel.MetricModel.NullParameter.ModelName"
	DataModel_MetricModel_NullParameter_OrderByDirection         = "DataModel.MetricModel.NullParameter.OrderByDirection"
	DataModel_MetricModel_NullParameter_OrderByName              = "DataModel.MetricModel.NullParameter.OrderByName"
	DataModel_MetricModel_NullParameter_QueryType                = "DataModel.MetricModel.NullParameter.QueryType"
	DataModel_MetricModel_NullParameter_Schedule                 = "DataModel.MetricModel.NullParameter.Schedule"
	DataModel_MetricModel_NullParameter_ScheduleExpression       = "DataModel.MetricModel.NullParameter.ScheduleExpression"
	DataModel_MetricModel_NullParameter_ScheduleType             = "DataModel.MetricModel.NullParameter.ScheduleType"
	DataModel_MetricModel_NullParameter_SQLAggrExpression        = "DataModel.MetricModel.NullParameter.SQLAggrExpression"
	DataModel_MetricModel_NullParameter_Step                     = "DataModel.MetricModel.NullParameter.Step"
	DataModel_MetricModel_NullParameter_TaskName                 = "DataModel.MetricModel.NullParameter.TaskName"
	DataModel_MetricModel_NullParameter_TimeWindows              = "DataModel.MetricModel.NullParameter.TimeWindows"
	DataModel_MetricModel_NullParameter_Unit                     = "DataModel.MetricModel.NullParameter.Unit"
	DataModel_MetricModel_NullParameter_UnitType                 = "DataModel.MetricModel.NullParameter.UnitType"
	DataModel_MetricModel_TaskNameExisted                        = "DataModel.MetricModel.TaskNameExisted"
	DataModel_MetricModel_UnsupportMetricType                    = "DataModel.MetricModel.UnsupportMetricType"
	DataModel_MetricModel_UnsupportQueryType                     = "DataModel.MetricModel.UnsupportQueryType"
	DataModel_MetricModel_UnsupportHavingConditionOperation      = "DataModel.MetricModel.UnsupportHavingConditionOperation"

	// 404
	DataModel_MetricModel_MetricModelNotFound = "DataModel.MetricModel.MetricModelNotFound"
	DataModel_MetricModel_MetricTaskNotFound  = "DataModel.MetricModel.MetricTaskNotFound"

	// 500
	DataModel_MetricModel_InternalError                                  = "DataModel.MetricModel.InternalError"
	DataModel_MetricModel_InternalError_BeginTransactionFailed           = "DataModel.MetricModel.InternalError.BeginTransactionFailed"
	DataModel_MetricModel_InternalError_CheckDuplicateMeasureNameFailed  = "DataModel.MetricModel.InternalError.CheckDuplicateMeasureNameFailed"
	DataModel_MetricModel_InternalError_CheckFormulaFailed               = "DataModel.MetricModel.InternalError.CheckFormulaFailed"
	DataModel_MetricModel_InternalError_CheckMetricModelTaskExistByName  = "DataModel.MetricModel.InternalError.CheckMetricModelTaskExistByName"
	DataModel_MetricModel_InternalError_CheckModelIfExistFailed          = "DataModel.MetricModel.InternalError.CheckModelIfExistFailed"
	DataModel_MetricModel_InternalError_GenerateIDFailed                 = "DataModel.MetricModel.InternalError.GenerateIDFailed"
	DataModel_MetricModel_InternalError_GetDataViewByIDFailed            = "DataModel.MetricModel.InternalError.GetDataViewByIDFailed"
	DataModel_MetricModel_InternalError_GetDataViewByNameFailed          = "DataModel.MetricModel.InternalError.GetDataViewByNameFailed"
	DataModel_MetricModel_InternalError_GetDataViewQueryFiltersFailed    = "DataModel.MetricModel.InternalError.GetDataViewQueryFiltersFailed"
	DataModel_MetricModel_InternalError_GetMetricTaskByIDFailed          = "DataModel.MetricModel.InternalError.GetMetricTaskByIDFailed"
	DataModel_MetricModel_InternalError_GetMetricTasksByModelIDsFailed   = "DataModel.MetricModel.InternalError.GetMetricTasksByModelIDsFailed"
	DataModel_MetricModel_InternalError_GetModelByIDFailed               = "DataModel.MetricModel.InternalError.GetModelByIDFailed"
	DataModel_MetricModel_InternalError_GetModelIDByNameFailed           = "DataModel.MetricModel.InternalError.GetModelIDByNameFailed"
	DataModel_MetricModel_InternalError_GetSimpleIndexBasesByTypesFailed = "DataModel.MetricModel.InternalError.GetSimpleIndexBasesByTypesFailed"
	DataModel_MetricModel_InternalError_MeasureNameRuleCompileFailed     = "DataModel.MetricModel.InternalError.MeasureNameRuleCompileFailed"
)

// 指标模型分组错误码
const (
	// 400
	DataModel_MetricModelGroup_GroupNameExisted         = "DataModel.MetricModelGroup.GroupNameExisted"
	DataModel_MetricModelGroup_InvalidParameter         = "DataModel.MetricModelGroup.InvalidParameter"
	DataModel_MetricModelGroup_InvalidParameter_Builtin = "DataModel.MetricModelGroup.InvalidParameter.Builtin"
	DataModel_MetricModelGroup_InvalidParameter_Force   = "DataModel.MetricModelGroup.InvalidParameter.Force"
	DataModel_MetricModelGroup_InvalidParameter_GroupID = "DataModel.MetricModelGroup.InvalidParameter.GroupID"
	DataModel_MetricModelGroup_LengthExceeded_GroupName = "DataModel.MetricModelGroup.LengthExceeded.GroupName"
	DataModel_MetricModelGroup_NullParameter_GrouplName = "DataModel.MetricModelGroup.NullParameter.GroupName"

	// 403
	DataModel_MetricModelGroup_ForbiddenBuiltinGroupID     = "DataModel.MetricModelGroup.ForbiddenBuiltinGroupID"
	DataModel_MetricModelGroup_ForbiddenDeleteBuiltinGroup = "DataModel.MetricModelGroup.ForbiddenDeleteBuiltinGroup"
	DataModel_MetricModelGroup_GroupNotEmpty               = "DataModel.MetricModelGroup.GroupNotEmpty"

	// 404
	DataModel_MetricModelGroup_GroupNotFound = "DataModel.MetricModelGroup.GroupNotFound"

	// 500
	DataModel_MetricModelGroup_InternalError                                   = "DataModel.MetricModelGroup.InternalError"
	DataModel_MetricModelGroup_InternalError_CheckGroupIfExistFailed           = "DataModel.MetricModelGroup.InternalError.CheckGroupIfExistFailed"
	DataModel_MetricModelGroup_InternalError_GenerateIDFailed                  = "DataModel.MetricModelGroup.InternalError.GenerateIDFailed"
	DataModel_MetricModelGroup_InternalError_GetGroupByIDFailed                = "DataModel.MetricModelGroup.InternalError.GetGroupByIDFailed"
	DataModel_MetricModelGroup_InternalError_GetGroupsTotalFailed              = "DataModel.MetricModelGroup.InternalError.GetGroupsTotalFailed"
	DataModel_MetricModelGroup_InternalError_GetMetricModelGroupIDByNameFailed = "DataModel.MetricModelGroup.InternalError.GetMetricModelGroupIDByNameFailed"
	DataModel_MetricModelGroup_InternalError_ListGroupsFailed                  = "DataModel.MetricModelGroup.InternalError.ListGroupsFailed"
)

var (
	metricModelErrCodeList = []string{
		// 400
		DataModel_MetricModel_AggregationNotExisted,
		DataModel_MetricModel_AnalysisDimensionNotExisted,
		DataModel_MetricModel_IDExisted,
		DataModel_MetricModel_CombinationNameExisted,
		DataModel_MetricModel_OrderByFieldNotExisted,
		DataModel_MetricModel_CountExceeded_TagTotal,
		DataModel_MetricModel_CountExceeded_TaskTotal,
		DataModel_MetricModel_Duplicated_ModelIDInFile,
		DataModel_MetricModel_Duplicated_CombinationName,
		DataModel_MetricModel_Duplicated_IndexBase,
		DataModel_MetricModel_Duplicated_MeasureName,
		DataModel_MetricModel_Duplicated_ModelName,
		DataModel_MetricModel_Duplicated_TaskName,
		DataModel_MetricModel_Duplicated_TaskStep,
		DataModel_MetricModel_Duplicated_TimeWindows,
		DataModel_MetricModel_DateFieldNotExisted,
		DataModel_MetricModel_GroupByFieldNotExisted,
		DataModel_MetricModel_InvalidParameter,
		DataModel_MetricModel_InvalidParameter_DateField,
		DataModel_MetricModel_InvalidParameter_DateFormat,
		DataModel_MetricModel_InvalidParameter_DataSourceType,
		DataModel_MetricModel_InvalidParameter_DerivedCondition,
		DataModel_MetricModel_InvalidParameter_Formula,
		DataModel_MetricModel_InvalidParameter_Group,
		DataModel_MetricModel_InvalidParameter_IncludeView,
		DataModel_MetricModel_InvalidParameter_MeasureField,
		DataModel_MetricModel_InvalidParameter_MeasureFieldNotExists,
		DataModel_MetricModel_InvalidParameter_MeasureName,
		DataModel_MetricModel_InvalidParameter_ModelID,
		DataModel_MetricModel_InvalidParameter_OrderByDirection,
		DataModel_MetricModel_InvalidParameter_RetraceDuration,
		DataModel_MetricModel_InvalidParameter_ScheduleExpression,
		DataModel_MetricModel_InvalidParameter_ScheduleType,
		DataModel_MetricModel_InvalidParameter_FormulaConfig,
		DataModel_MetricModel_InvalidParameter_SQLAggrExpression,
		DataModel_MetricModel_InvalidParameter_SQLCondition,
		DataModel_MetricModel_InvalidParameter_Step,
		DataModel_MetricModel_InvalidParameter_TimeWindows,
		DataModel_MetricModel_InvalidParameter_UnitType,
		DataModel_MetricModel_LengthExceeded_MeasureName,
		DataModel_MetricModel_LengthExceeded_ModelName,
		DataModel_MetricModel_LengthExceeded_TaskComment,
		DataModel_MetricModel_LengthExceeded_TaskName,
		DataModel_MetricModel_ModelNameExisted,
		DataModel_MetricModel_NullParameter_DependMetricModel,
		DataModel_MetricModel_NullParameter_DateField,
		DataModel_MetricModel_NullParameter_DataSource,
		DataModel_MetricModel_NullParameter_DataSourceID,
		DataModel_MetricModel_NullParameter_DerivedCondition,
		// DataModel_MetricModel_NullParameter_DataView,
		DataModel_MetricModel_NullParameter_Formula,
		DataModel_MetricModel_NullParameter_IndexBase,
		DataModel_MetricModel_NullParameter_MeasureField,
		DataModel_MetricModel_NullParameter_MeasureName,
		DataModel_MetricModel_NullParameter_MetricType,
		DataModel_MetricModel_NullParameter_ModelName,
		DataModel_MetricModel_NullParameter_OrderByDirection,
		DataModel_MetricModel_NullParameter_OrderByName,
		DataModel_MetricModel_NullParameter_QueryType,
		DataModel_MetricModel_NullParameter_Schedule,
		DataModel_MetricModel_NullParameter_ScheduleExpression,
		DataModel_MetricModel_NullParameter_ScheduleType,
		DataModel_MetricModel_NullParameter_FormulaConfig,
		DataModel_MetricModel_InvalidParameter_HavingConditionName,
		DataModel_MetricModel_NullParameter_HavingConditionOperation,
		DataModel_MetricModel_NullParameter_SQLAggrExpression,
		DataModel_MetricModel_NullParameter_Step,
		DataModel_MetricModel_NullParameter_TaskName,
		DataModel_MetricModel_NullParameter_TimeWindows,
		DataModel_MetricModel_NullParameter_Unit,
		DataModel_MetricModel_NullParameter_UnitType,
		DataModel_MetricModel_TaskNameExisted,
		DataModel_MetricModel_UnsupportMetricType,
		DataModel_MetricModel_UnsupportQueryType,
		DataModel_MetricModel_UnsupportHavingConditionOperation,

		// 404
		DataModel_MetricModel_MetricModelNotFound,
		DataModel_MetricModel_MetricTaskNotFound,

		// 500
		DataModel_MetricModel_InternalError,
		DataModel_MetricModel_InternalError_BeginTransactionFailed,
		DataModel_MetricModel_InternalError_CheckDuplicateMeasureNameFailed,
		DataModel_MetricModel_InternalError_CheckFormulaFailed,
		DataModel_MetricModel_InternalError_CheckMetricModelTaskExistByName,
		DataModel_MetricModel_InternalError_CheckModelIfExistFailed,
		DataModel_MetricModel_InternalError_GenerateIDFailed,
		DataModel_MetricModel_InternalError_GetDataViewByIDFailed,
		DataModel_MetricModel_InternalError_GetDataViewByNameFailed,
		DataModel_MetricModel_InternalError_GetDataViewQueryFiltersFailed,
		DataModel_MetricModel_InternalError_GetMetricTaskByIDFailed,
		DataModel_MetricModel_InternalError_GetMetricTasksByModelIDsFailed,
		DataModel_MetricModel_InternalError_GetModelByIDFailed,
		DataModel_MetricModel_InternalError_GetModelIDByNameFailed,
		DataModel_MetricModel_InternalError_GetSimpleIndexBasesByTypesFailed,
		DataModel_MetricModel_InternalError_MeasureNameRuleCompileFailed,

		// ---指标模型分组模块---
		// 400
		DataModel_MetricModelGroup_GroupNameExisted,
		DataModel_MetricModelGroup_InvalidParameter,
		DataModel_MetricModelGroup_InvalidParameter_Builtin,
		DataModel_MetricModelGroup_InvalidParameter_Force,
		DataModel_MetricModelGroup_InvalidParameter_GroupID,
		DataModel_MetricModelGroup_LengthExceeded_GroupName,
		DataModel_MetricModelGroup_NullParameter_GrouplName,

		// 404
		DataModel_MetricModelGroup_GroupNotFound,

		//403
		DataModel_MetricModelGroup_ForbiddenBuiltinGroupID,
		DataModel_MetricModelGroup_ForbiddenDeleteBuiltinGroup,
		DataModel_MetricModelGroup_GroupNotEmpty,

		// 500
		DataModel_MetricModelGroup_InternalError,
		DataModel_MetricModelGroup_InternalError_CheckGroupIfExistFailed,
		DataModel_MetricModelGroup_InternalError_GenerateIDFailed,
		DataModel_MetricModelGroup_InternalError_GetGroupByIDFailed,
		DataModel_MetricModelGroup_InternalError_GetGroupsTotalFailed,
		DataModel_MetricModelGroup_InternalError_GetMetricModelGroupIDByNameFailed,
		DataModel_MetricModelGroup_InternalError_ListGroupsFailed,
	}
)
