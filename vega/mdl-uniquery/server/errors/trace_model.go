// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package errors 服务错误码
package errors

// 链路模型
const (
	// 400
	Uniquery_TraceModel_InvalidParameter_ModelID = "Uniquery.TraceModel.InvalidParameter.ModelID"

	// 404
	Uniquery_TraceModel_TraceModelNotFound = "Uniquery.TraceModel.TraceModelNotFound"
	Uniquery_TraceModel_TraceNotFound      = "Uniquery.TraceModel.TraceNotFound"
	Uniquery_TraceModel_SpanNotFound       = "Uniquery.TraceModel.SpanNotFound"

	// 500
	Uniquery_TraceModel_InternalError_GetTraceModelByIDFailed          = "Uniquery.TraceModel.InternalError.GetTraceModelByIDFailed"
	Uniquery_TraceModel_InternalError_SimulateCreateTraceModelFailed   = "Uniquery.TraceModel.InternalError.SimulateCreateTraceModelFailed"
	Uniquery_TraceModel_InternalError_SimulateUpdateTraceModelFailed   = "Uniquery.TraceModel.InternalError.SimulateUpdateTraceModelFailed"
	Uniquery_TraceModel_InternalError_SplitSpanIDFailed                = "Uniquery.TraceModel.InternalError.SplitSpanIDFailed"
	Uniquery_TraceModel_InternalError_GetUnderlyingDataSouceTypeFailed = "Uniquery.TraceModel.InternalError.GetUnderlyingDataSouceTypeFailed"
	Uniquery_TraceModel_InternalError_GetDataConnectionByIDFailed      = "Uniquery.TraceModel.InternalError.GetDataConnectionByIDFailed"
	Uniquery_TraceModel_InternalError_ProcessDataConnectionFailed      = "Uniquery.TraceModel.InternalError.ProcessDataConnectionFailed"
	Uniquery_TraceModel_InternalError_GetTingYunTraceListFailed        = "Uniquery.TraceModel.InternalError.GetTingYunTraceListFailed"
	Uniquery_TraceModel_InternalError_GetTingYunTraceDetailFailed      = "Uniquery.TraceModel.InternalError.GetTingYunTraceDetailFailed"
)

var (
	traceModelErrCodeList = []string{
		// 400
		Uniquery_TraceModel_InvalidParameter_ModelID,

		// 404
		Uniquery_TraceModel_SpanNotFound,
		Uniquery_TraceModel_TraceModelNotFound,
		Uniquery_TraceModel_TraceNotFound,

		// 500
		Uniquery_TraceModel_InternalError_GetDataConnectionByIDFailed,
		Uniquery_TraceModel_InternalError_GetTingYunTraceDetailFailed,
		Uniquery_TraceModel_InternalError_GetTingYunTraceListFailed,
		Uniquery_TraceModel_InternalError_GetTraceModelByIDFailed,
		Uniquery_TraceModel_InternalError_GetUnderlyingDataSouceTypeFailed,
		Uniquery_TraceModel_InternalError_ProcessDataConnectionFailed,
		Uniquery_TraceModel_InternalError_SimulateCreateTraceModelFailed,
		Uniquery_TraceModel_InternalError_SimulateUpdateTraceModelFailed,
		Uniquery_TraceModel_InternalError_SplitSpanIDFailed,
	}
)
