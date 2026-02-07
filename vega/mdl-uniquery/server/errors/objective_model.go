// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package errors

// 目标模型错误码
const (
	// 400
	Uniquery_ObjectiveModel_AssociateMetricModelsUnit_Different          = "Uniquery.ObjectiveModel.AssociateMetricModelsUnit.Different"
	Uniquery_ObjectiveModel_CountExceeded_ComprehensiveMetricModelsTotal = "Uniquery.ObjectiveModel.CountExceeded.ComprehensiveMetricModelsTotal"
	Uniquery_ObjectiveModel_CountExceeded_AdditionalMetricModelsTotal    = "Uniquery.ObjectiveModel.CountExceeded.AdditionalMetricModelsTotal"
	Uniquery_ObjectiveModel_CountExceeded_StatusTotal                    = "Uniquery.ObjectiveModel.CountExceeded.StatusTotal"
	Uniquery_ObjectiveModel_InvalidParameter                             = "Uniquery.ObjectiveModel.InvalidParameter"
	Uniquery_ObjectiveModel_InvalidParameter_ComprehensiveWeight         = "Uniquery.ObjectiveModel.InvalidParameter.ComprehensiveWeight"
	Uniquery_ObjectiveModel_InvalidParameter_IgnoringMemoryCache         = "Uniquery.ObjectiveModel.InvalidParameter.IgnoringMemoryCache"
	Uniquery_ObjectiveModel_InvalidParameter_IgnoringStoreCache          = "Uniquery.ObjectiveModel.InvalidParameter.IgnoringStoreCache"
	Uniquery_ObjectiveModel_InvalidParameter_IncludeModel                = "Uniquery.ObjectiveModel.InvalidParameter.IncludeModel"
	Uniquery_ObjectiveModel_InvalidParameter_MetricModelID               = "Uniquery.ObjectiveModel.InvalidParameter.MetricModelID"
	Uniquery_ObjectiveModel_InvalidParameter_Objective                   = "Uniquery.ObjectiveModel.InvalidParameter.Objective"
	Uniquery_ObjectiveModel_InvalidParameter_ObjectiveUnit               = "Uniquery.ObjectiveModel.InvalidParameter.ObjectiveUnit"
	Uniquery_ObjectiveModel_InvalidParameter_ObjectiveConfig             = "Uniquery.ObjectiveModel.InvalidParameter.ObjectiveConfig"
	Uniquery_ObjectiveModel_InvalidParameter_StatusRanges                = "Uniquery.ObjectiveModel.InvalidParameter.StatusRanges"
	Uniquery_ObjectiveModel_MetricModelNotFound                          = "Uniquery.ObjectiveModel.MetricModelNotFound"
	Uniquery_ObjectiveModel_NullParameter_ComprehensiveMetricModels      = "Uniquery.ObjectiveModel.NullParameter.ComprehensiveMetricModels"
	Uniquery_ObjectiveModel_NullParameter_ComprehensiveMetricModelID     = "Uniquery.ObjectiveModel.NullParameter.ComprehensiveMetricModelID"
	Uniquery_ObjectiveModel_NullParameter_ComprehensiveWeight            = "Uniquery.ObjectiveModel.NullParameter.ComprehensiveWeight"
	Uniquery_ObjectiveModel_NullParameter_GoodMetricModel                = "Uniquery.ObjectiveModel.NullParameter.GoodMetricModel"
	Uniquery_ObjectiveModel_NullParameter_GoodMetricModelID              = "Uniquery.ObjectiveModel.NullParameter.GoodMetricModelID"
	Uniquery_ObjectiveModel_NullParameter_Objective                      = "Uniquery.ObjectiveModel.NullParameter.Objective"
	Uniquery_ObjectiveModel_NullParameter_ObjectiveConfig                = "Uniquery.ObjectiveModel.NullParameter.ObjectiveConfig"
	Uniquery_ObjectiveModel_NullParameter_ObjectiveType                  = "Uniquery.ObjectiveModel.NullParameter.ObjectiveType"
	Uniquery_ObjectiveModel_NullParameter_ObjectiveUnit                  = "Uniquery.ObjectiveModel.NullParameter.ObjectiveUnit"
	Uniquery_ObjectiveModel_NullParameter_Period                         = "Uniquery.ObjectiveModel.NullParameter.Period"
	Uniquery_ObjectiveModel_NullParameter_TotalMetricModel               = "Uniquery.ObjectiveModel.NullParameter.TotalMetricModel"
	Uniquery_ObjectiveModel_NullParameter_TotalMetricModelID             = "Uniquery.ObjectiveModel.NullParameter.TotalMetricModelID"
	Uniquery_ObjectiveModel_NullParameter_AdditionalMetricModelID        = "Uniquery.ObjectiveModel.NullParameter.AdditionalMetricModelID"
	Uniquery_ObjectiveModel_UnsupportQuery                               = "Uniquery.ObjectiveModel.UnsupportQuery"
	Uniquery_ObjectiveModel_UnsupportObjectiveType                       = "Uniquery.ObjectiveModel.UnsupportObjectiveType"

	//404
	Uniquery_ObjectiveModel_ObjectiveModelNotFound = "Uniquery.ObjectiveModel.ObjectiveModelNotFound"

	// 500
	Uniquery_ObjectiveModel_InternalError                               = "Uniquery.ObjectiveModel.InternalError"
	Uniquery_ObjectiveModel_InternalError_GetMetricModelsFailed         = "Uniquery.ObjectiveModel.InternalError.GetMetricModelsFailed"
	Uniquery_ObjectiveModel_InternalError_GetModelByIdFailed            = "Uniquery.ObjectiveModel.InternalError.GetObjectiveModelByIdFailed"
	Uniquery_ObjectiveModel_InternalError_MetricTimePointsNotEqual      = "Uniquery.ObjectiveModel.InternalError.MetricTimePointsNotEqual"
	Uniquery_ObjectiveModel_InternalError_SubmitGetMetricDataTaskFailed = "Uniquery.ObjectiveModel.InternalError.SubmitGetMetricDataTaskFailed"
)

var (
	objectiveModelErrCodeList = []string{
		// 400
		Uniquery_ObjectiveModel_InvalidParameter,
		Uniquery_ObjectiveModel_InvalidParameter_IgnoringMemoryCache,
		Uniquery_ObjectiveModel_InvalidParameter_IgnoringStoreCache,
		Uniquery_ObjectiveModel_InvalidParameter_IncludeModel,
		Uniquery_ObjectiveModel_UnsupportObjectiveType,
		Uniquery_ObjectiveModel_NullParameter_ObjectiveType,
		Uniquery_ObjectiveModel_NullParameter_ObjectiveConfig,
		Uniquery_ObjectiveModel_NullParameter_Objective,
		Uniquery_ObjectiveModel_InvalidParameter_Objective,
		Uniquery_ObjectiveModel_NullParameter_Period,
		Uniquery_ObjectiveModel_NullParameter_GoodMetricModel,
		Uniquery_ObjectiveModel_NullParameter_GoodMetricModelID,
		Uniquery_ObjectiveModel_NullParameter_TotalMetricModel,
		Uniquery_ObjectiveModel_NullParameter_TotalMetricModelID,
		Uniquery_ObjectiveModel_NullParameter_ObjectiveUnit,
		Uniquery_ObjectiveModel_NullParameter_ComprehensiveMetricModels,
		Uniquery_ObjectiveModel_NullParameter_ComprehensiveMetricModelID,
		Uniquery_ObjectiveModel_NullParameter_ComprehensiveWeight,
		Uniquery_ObjectiveModel_InvalidParameter_ComprehensiveWeight,
		Uniquery_ObjectiveModel_NullParameter_AdditionalMetricModelID,
		Uniquery_ObjectiveModel_InvalidParameter_ObjectiveConfig,
		Uniquery_ObjectiveModel_InvalidParameter_MetricModelID,
		Uniquery_ObjectiveModel_InvalidParameter_StatusRanges,
		Uniquery_ObjectiveModel_InvalidParameter_ObjectiveUnit,
		Uniquery_ObjectiveModel_UnsupportQuery,
		Uniquery_ObjectiveModel_CountExceeded_ComprehensiveMetricModelsTotal,
		Uniquery_ObjectiveModel_CountExceeded_AdditionalMetricModelsTotal,
		Uniquery_ObjectiveModel_CountExceeded_StatusTotal,
		Uniquery_ObjectiveModel_MetricModelNotFound,
		Uniquery_ObjectiveModel_AssociateMetricModelsUnit_Different,

		//404
		Uniquery_ObjectiveModel_ObjectiveModelNotFound,

		// 500
		Uniquery_ObjectiveModel_InternalError,
		Uniquery_ObjectiveModel_InternalError_GetMetricModelsFailed,
		Uniquery_ObjectiveModel_InternalError_SubmitGetMetricDataTaskFailed,
		Uniquery_ObjectiveModel_InternalError_MetricTimePointsNotEqual,

		Uniquery_ObjectiveModel_InternalError_GetModelByIdFailed,
	}
)
