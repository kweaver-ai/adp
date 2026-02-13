// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package errors 服务错误码
package errors

// 链路模型错误码
const (
	// 400
	DataModel_TraceModel_DependentDataConnectionNotFound          = "DataModel.TraceModel.DependentDataConnectionNotFound"
	DataModel_TraceModel_DependentDataViewNotFound                = "DataModel.TraceModel.DependentDataViewNotFound"
	DataModel_TraceModel_InvalidBasicAttributeConfig              = "DataModel.TraceModel.InvalidBasicAttributeConfig"
	DataModel_TraceModel_InvalidBasicAttributeConfig_Duration     = "DataModel.TraceModel.InvalidBasicAttributeConfig.Duration"
	DataModel_TraceModel_InvalidBasicAttributeConfig_EndTime      = "DataModel.TraceModel.InvalidBasicAttributeConfig.EndTime"
	DataModel_TraceModel_InvalidBasicAttributeConfig_Kind         = "DataModel.TraceModel.InvalidBasicAttributeConfig.Kind"
	DataModel_TraceModel_InvalidBasicAttributeConfig_Name         = "DataModel.TraceModel.InvalidBasicAttributeConfig.Name"
	DataModel_TraceModel_InvalidBasicAttributeConfig_ParentSpanID = "DataModel.TraceModel.InvalidBasicAttributeConfig.ParentSpanID"
	DataModel_TraceModel_InvalidBasicAttributeConfig_ServiceName  = "DataModel.TraceModel.InvalidBasicAttributeConfig.ServiceName"
	DataModel_TraceModel_InvalidBasicAttributeConfig_SpanID       = "DataModel.TraceModel.InvalidBasicAttributeConfig.SpanID"
	DataModel_TraceModel_InvalidBasicAttributeConfig_StartTime    = "DataModel.TraceModel.InvalidBasicAttributeConfig.StartTime"
	DataModel_TraceModel_InvalidBasicAttributeConfig_Status       = "DataModel.TraceModel.InvalidBasicAttributeConfig.Status"
	DataModel_TraceModel_InvalidBasicAttributeConfig_TraceID      = "DataModel.TraceModel.InvalidBasicAttributeConfig.TraceID"
	DataModel_TraceModel_InvalidParameter_EnabledRelatedLog       = "DataModel.TraceModel.InvalidParameter.EnabledRelatedLog"
	DataModel_TraceModel_InvalidParameter_ModelIDs                = "DataModel.TraceModel.InvalidParameter.ModelIDs"
	DataModel_TraceModel_InvalidParameter_RelatedLogSourceType    = "DataModel.TraceModel.InvalidParameter.RelatedLogSourceType"
	DataModel_TraceModel_InvalidParameter_SpanSourceType          = "DataModel.TraceModel.InvalidParameter.SpanSourceType"
	DataModel_TraceModel_InvalidRelatedLogConfig_SpanID           = "DataModel.TraceModel.InvalidRelatedLogConfig.SpanID"
	DataModel_TraceModel_InvalidRelatedLogConfig_TraceID          = "DataModel.TraceModel.InvalidRelatedLogConfig.TraceID"
	DataModel_TraceModel_LengthExceeded_ModelName                 = "DataModel.TraceModel.LengthExceeded.ModelName"
	DataModel_TraceModel_ModelNameExisted                         = "DataModel.TraceModel.ModelNameExisted"
	DataModel_TraceModel_NotUniqueInBatch_ModelName               = "DataModel.TraceModel.NotUniqueInBatch.ModelName"
	DataModel_TraceModel_NullParameter_ModelName                  = "DataModel.TraceModel.NullParameter.ModelName"

	// 404
	DataModel_TraceModel_TraceModelNotFound = "DataModel.TraceModel.TraceModelNotFound"

	// 500
	DataModel_TraceModel_InternalError_CreateTraceModelsFailed             = "DataModel.TraceModel.InternalError.CreateTraceModelsFailed"
	DataModel_TraceModel_InternalError_DeleteTraceModelsFailed             = "DataModel.TraceModel.InternalError.DeleteTraceModelsFailed"
	DataModel_TraceModel_InternalError_GetDetailedTraceModelMapByIDsFailed = "DataModel.TraceModel.InternalError.GetDetailedTraceModelMapByIDsFailed"
	DataModel_TraceModel_InternalError_GetSimpleTraceModelMapByIDsFailed   = "DataModel.TraceModel.InternalError.GetSimpleTraceModelMapByIDsFailed"
	DataModel_TraceModel_InternalError_GetSimpleTraceModelMapByNamesFailed = "DataModel.TraceModel.InternalError.GetSimpleTraceModelMapByNamesFailed"
	DataModel_TraceModel_InternalError_GetTraceModelsFailed                = "DataModel.TraceModel.InternalError.GetTraceModelsFailed"
	DataModel_TraceModel_InternalError_GetTraceModelTotalFailed            = "DataModel.TraceModel.InternalError.GetTraceModelTotalFailed"
	DataModel_TraceModel_InternalError_GetUnderlyingDataSourceTypeFailed   = "DataModel.TraceModel.InternalError.GetUnderlyingDataSourceTypeFailed"
	DataModel_TraceModel_InternalError_InitTraceModelProcessor             = "DataModel.TraceModel.InternalError.InitTraceModelProcessor"
	DataModel_TraceModel_InternalError_ListTraceModelsFailed               = "DataModel.TraceModel.InternalError.ListTraceModelsFailed"
	DataModel_TraceModel_InternalError_UpdateTraceModelFailed              = "DataModel.TraceModel.InternalError.UpdateTraceModelFailed"
)

var (
	traceModelErrCodeList = []string{
		// 400
		DataModel_TraceModel_DependentDataConnectionNotFound,
		DataModel_TraceModel_DependentDataViewNotFound,
		DataModel_TraceModel_InvalidBasicAttributeConfig,
		DataModel_TraceModel_InvalidBasicAttributeConfig_Duration,
		DataModel_TraceModel_InvalidBasicAttributeConfig_EndTime,
		DataModel_TraceModel_InvalidBasicAttributeConfig_Kind,
		DataModel_TraceModel_InvalidBasicAttributeConfig_Name,
		DataModel_TraceModel_InvalidBasicAttributeConfig_ParentSpanID,
		DataModel_TraceModel_InvalidBasicAttributeConfig_ServiceName,
		DataModel_TraceModel_InvalidBasicAttributeConfig_SpanID,
		DataModel_TraceModel_InvalidBasicAttributeConfig_StartTime,
		DataModel_TraceModel_InvalidBasicAttributeConfig_Status,
		DataModel_TraceModel_InvalidBasicAttributeConfig_TraceID,
		DataModel_TraceModel_InvalidParameter_EnabledRelatedLog,
		DataModel_TraceModel_InvalidParameter_ModelIDs,
		DataModel_TraceModel_InvalidParameter_RelatedLogSourceType,
		DataModel_TraceModel_InvalidParameter_SpanSourceType,
		DataModel_TraceModel_InvalidRelatedLogConfig_SpanID,
		DataModel_TraceModel_InvalidRelatedLogConfig_TraceID,
		DataModel_TraceModel_LengthExceeded_ModelName,
		DataModel_TraceModel_ModelNameExisted,
		DataModel_TraceModel_NotUniqueInBatch_ModelName,
		DataModel_TraceModel_NullParameter_ModelName,

		// 404
		DataModel_TraceModel_TraceModelNotFound,

		// 500
		DataModel_TraceModel_InternalError_CreateTraceModelsFailed,
		DataModel_TraceModel_InternalError_DeleteTraceModelsFailed,
		DataModel_TraceModel_InternalError_GetDetailedTraceModelMapByIDsFailed,
		DataModel_TraceModel_InternalError_GetSimpleTraceModelMapByIDsFailed,
		DataModel_TraceModel_InternalError_GetSimpleTraceModelMapByNamesFailed,
		DataModel_TraceModel_InternalError_GetTraceModelsFailed,
		DataModel_TraceModel_InternalError_GetTraceModelTotalFailed,
		DataModel_TraceModel_InternalError_GetUnderlyingDataSourceTypeFailed,
		DataModel_TraceModel_InternalError_InitTraceModelProcessor,
		DataModel_TraceModel_InternalError_ListTraceModelsFailed,
		DataModel_TraceModel_InternalError_UpdateTraceModelFailed,
	}
)
