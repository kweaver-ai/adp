// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package errors

// 目标模型错误码
const (
	// 400
	DataModel_ObjectiveModel_AssociateMetricModelsUnit_Different          = "DataModel.ObjectiveModel.AssociateMetricModelsUnit.Different"
	DataModel_ObjectiveModel_CountExceeded_ComprehensiveMetricModelsTotal = "DataModel.ObjectiveModel.CountExceeded.ComprehensiveMetricModelsTotal"
	DataModel_ObjectiveModel_CountExceeded_AdditionalMetricModelsTotal    = "DataModel.ObjectiveModel.CountExceeded.AdditionalMetricModelsTotal"
	DataModel_ObjectiveModel_CountExceeded_StatusTotal                    = "DataModel.ObjectiveModel.CountExceeded.StatusTotal"
	DataModel_ObjectiveModel_Duplicated_ModelIDInFile                     = "DataModel.ObjectiveModel.Duplicated.ModelIDInFile"
	DataModel_ObjectiveModel_Duplicated_ObjectiveModelName                = "DataModel.ObjectiveModel.Duplicated.ObjectiveModelName"
	DataModel_ObjectiveModel_Duplicated_TaskStep                          = "DataModel.ObjectiveModel.Duplicated.TaskStep"
	DataModel_ObjectiveModel_InvalidParameter                             = "DataModel.ObjectiveModel.InvalidParameter"
	DataModel_ObjectiveModel_InvalidParameter_ComprehensiveWeight         = "DataModel.ObjectiveModel.InvalidParameter.ComprehensiveWeight"
	DataModel_ObjectiveModel_InvalidParameter_Objective                   = "DataModel.ObjectiveModel.InvalidParameter.Objective"
	DataModel_ObjectiveModel_InvalidParameter_ObjectiveConfig             = "DataModel.ObjectiveModel.InvalidParameter.ObjectiveConfig"
	DataModel_ObjectiveModel_InvalidParameter_ObjectiveUnit               = "DataModel.ObjectiveModel.InvalidParameter.ObjectiveUnit"
	DataModel_ObjectiveModel_InvalidParameter_Period                      = "DataModel.ObjectiveModel.InvalidParameter.Period"
	DataModel_ObjectiveModel_InvalidParameter_RetraceDuration             = "DataModel.ObjectiveModel.InvalidParameter.RetraceDuration"
	DataModel_ObjectiveModel_InvalidParameter_ScheduleExpression          = "DataModel.ObjectiveModel.InvalidParameter.ScheduleExpression"
	DataModel_ObjectiveModel_InvalidParameter_ScheduleType                = "DataModel.ObjectiveModel.InvalidParameter.ScheduleType"
	DataModel_ObjectiveModel_InvalidParameter_StatusRanges                = "DataModel.ObjectiveModel.InvalidParameter.StatusRanges"
	DataModel_ObjectiveModel_InvalidParameter_Step                        = "DataModel.ObjectiveModel.InvalidParameter.Step"
	DataModel_ObjectiveModel_LengthExceeded_ModelName                     = "DataModel.ObjectiveModel.LengthExceeded.ModelName"
	DataModel_ObjectiveModel_LengthExceeded_TaskComment                   = "DataModel.ObjectiveModel.LengthExceeded.TaskComment"
	DataModel_ObjectiveModel_ModelIDExisted                               = "DataModel.ObjectiveModel.ModelIDExisted"
	DataModel_ObjectiveModel_ModelNameExisted                             = "DataModel.ObjectiveModel.ModelNameExisted"
	DataModel_ObjectiveModel_NullParameter_AdditionalMetricModelID        = "DataModel.ObjectiveModel.NullParameter.AdditionalMetricModelID"
	DataModel_ObjectiveModel_NullParameter_ComprehensiveMetricModels      = "DataModel.ObjectiveModel.NullParameter.ComprehensiveMetricModels"
	DataModel_ObjectiveModel_NullParameter_ComprehensiveMetricModelID     = "DataModel.ObjectiveModel.NullParameter.ComprehensiveMetricModelID"
	DataModel_ObjectiveModel_NullParameter_ComprehensiveWeight            = "DataModel.ObjectiveModel.NullParameter.ComprehensiveWeight"
	DataModel_ObjectiveModel_NullParameter_GoodMetricModel                = "DataModel.ObjectiveModel.NullParameter.GoodMetricModel"
	DataModel_ObjectiveModel_NullParameter_GoodMetricModelID              = "DataModel.ObjectiveModel.NullParameter.GoodMetricModelID"
	DataModel_ObjectiveModel_NullParameter_IndexBase                      = "DataModel.ObjectiveModel.NullParameter.IndexBase"
	DataModel_ObjectiveModel_NullParameter_ModelName                      = "DataModel.ObjectiveModel.NullParameter.ModelName"
	DataModel_ObjectiveModel_NullParameter_Objective                      = "DataModel.ObjectiveModel.NullParameter.Objective"
	DataModel_ObjectiveModel_NullParameter_ObjectiveConfig                = "DataModel.ObjectiveModel.NullParameter.ObjectiveConfig"
	DataModel_ObjectiveModel_NullParameter_ObjectiveType                  = "DataModel.ObjectiveModel.NullParameter.ObjectiveType"
	DataModel_ObjectiveModel_NullParameter_ObjectiveUnit                  = "DataModel.ObjectiveModel.NullParameter.ObjectiveUnit"
	DataModel_ObjectiveModel_NullParameter_Period                         = "DataModel.ObjectiveModel.NullParameter.Period"
	DataModel_ObjectiveModel_NullParameter_Schedule                       = "DataModel.ObjectiveModel.NullParameter.Schedule"
	DataModel_ObjectiveModel_NullParameter_ScheduleExpression             = "DataModel.ObjectiveModel.NullParameter.ScheduleExpression"
	DataModel_ObjectiveModel_NullParameter_ScheduleType                   = "DataModel.ObjectiveModel.NullParameter.ScheduleType"
	DataModel_ObjectiveModel_NullParameter_Step                           = "DataModel.ObjectiveModel.NullParameter.Step"
	DataModel_ObjectiveModel_NullParameter_Task                           = "DataModel.ObjectiveModel.NullParameter.Task"
	DataModel_ObjectiveModel_NullParameter_TotalMetricModel               = "DataModel.ObjectiveModel.NullParameter.TotalMetricModel"
	DataModel_ObjectiveModel_NullParameter_TotalMetricModelID             = "DataModel.ObjectiveModel.NullParameter.TotalMetricModelID"
	DataModel_ObjectiveModel_UnsupportObjectiveType                       = "DataModel.ObjectiveModel.UnsupportObjectiveType"

	// 404
	DataModel_ObjectiveModel_ObjectiveModelNotFound = "DataModel.ObjectiveModel.ObjectiveModelNotFound"
	DataModel_ObjectiveModel_MetricModelNotFound    = "DataModel.ObjectiveModel.MetricModelNotFound"

	// 500
	DataModel_ObjectiveModel_InternalError_GetSimpleIndexBasesByTypesFailed = "DataModel.ObjectiveModel.InternalError.GetSimpleIndexBasesByTypesFailed"

	DataModel_ObjectiveModel_InternalError                                    = "DataModel.ObjectiveModel.InternalError"
	DataModel_ObjectiveModel_InternalError_BeginTransactionFailed             = "DataModel.ObjectiveModel.InternalError.BeginTransactionFailed"
	DataModel_ObjectiveModel_InternalError_CheckModelIfExistFailed            = "DataModel.ObjectiveModel.InternalError.CheckModelIfExistFailed"
	DataModel_ObjectiveModel_InternalError_GenerateIDFailed                   = "DataModel.ObjectiveModel.InternalError.GenerateIDFailed"
	DataModel_ObjectiveModel_InternalError_GetMetricTasksByModelIDsFailed     = "DataModel.ObjectiveModel.InternalError.GetMetricTasksByModelIDsFailed"
	DataModel_ObjectiveModel_InternalError_GetObjectiveModelsByModelIDsFailed = "DataModel.ObjectiveModel.InternalError.GetObjectiveModelsByModelIDsFailed"
)

var (
	objectiveModelErrCodeList = []string{
		DataModel_ObjectiveModel_NullParameter_ModelName,
		DataModel_ObjectiveModel_NullParameter_ObjectiveType,
		DataModel_ObjectiveModel_NullParameter_ObjectiveConfig,
		DataModel_ObjectiveModel_NullParameter_Task,
		DataModel_ObjectiveModel_NullParameter_Schedule,
		DataModel_ObjectiveModel_NullParameter_ScheduleExpression,
		DataModel_ObjectiveModel_NullParameter_ScheduleType,
		DataModel_ObjectiveModel_NullParameter_Period,
		DataModel_ObjectiveModel_NullParameter_GoodMetricModel,
		DataModel_ObjectiveModel_NullParameter_GoodMetricModelID,
		DataModel_ObjectiveModel_NullParameter_TotalMetricModel,
		DataModel_ObjectiveModel_NullParameter_TotalMetricModelID,
		DataModel_ObjectiveModel_NullParameter_ObjectiveUnit,
		DataModel_ObjectiveModel_NullParameter_ComprehensiveMetricModels,
		DataModel_ObjectiveModel_NullParameter_ComprehensiveMetricModelID,
		DataModel_ObjectiveModel_NullParameter_ComprehensiveWeight,
		DataModel_ObjectiveModel_NullParameter_AdditionalMetricModelID,
		DataModel_ObjectiveModel_NullParameter_Objective,
		DataModel_ObjectiveModel_NullParameter_IndexBase,
		DataModel_ObjectiveModel_NullParameter_Step,
		DataModel_ObjectiveModel_InvalidParameter,
		DataModel_ObjectiveModel_InvalidParameter_RetraceDuration,
		DataModel_ObjectiveModel_InvalidParameter_ScheduleExpression,
		DataModel_ObjectiveModel_InvalidParameter_ScheduleType,
		DataModel_ObjectiveModel_InvalidParameter_Step,
		DataModel_ObjectiveModel_InvalidParameter_ObjectiveConfig,
		DataModel_ObjectiveModel_InvalidParameter_Objective,
		DataModel_ObjectiveModel_InvalidParameter_Period,
		DataModel_ObjectiveModel_InvalidParameter_ObjectiveUnit,
		DataModel_ObjectiveModel_CountExceeded_ComprehensiveMetricModelsTotal,
		DataModel_ObjectiveModel_CountExceeded_AdditionalMetricModelsTotal,
		DataModel_ObjectiveModel_CountExceeded_StatusTotal,
		DataModel_ObjectiveModel_InvalidParameter_ComprehensiveWeight,
		DataModel_ObjectiveModel_InvalidParameter_StatusRanges,
		DataModel_ObjectiveModel_Duplicated_TaskStep,
		DataModel_ObjectiveModel_Duplicated_ObjectiveModelName,
		DataModel_ObjectiveModel_ModelIDExisted,
		DataModel_ObjectiveModel_ModelNameExisted,
		DataModel_ObjectiveModel_LengthExceeded_ModelName,
		DataModel_ObjectiveModel_LengthExceeded_TaskComment,
		DataModel_ObjectiveModel_UnsupportObjectiveType,
		DataModel_ObjectiveModel_AssociateMetricModelsUnit_Different,

		// 404
		DataModel_ObjectiveModel_ObjectiveModelNotFound,
		DataModel_ObjectiveModel_MetricModelNotFound,

		// 500
		DataModel_ObjectiveModel_InternalError_GetSimpleIndexBasesByTypesFailed,

		DataModel_ObjectiveModel_InternalError,
		DataModel_ObjectiveModel_InternalError_BeginTransactionFailed,
		DataModel_ObjectiveModel_InternalError_CheckModelIfExistFailed,
		DataModel_ObjectiveModel_InternalError_GenerateIDFailed,
		DataModel_ObjectiveModel_InternalError_GetMetricTasksByModelIDsFailed,
		DataModel_ObjectiveModel_InternalError_GetObjectiveModelsByModelIDsFailed,
	}
)
