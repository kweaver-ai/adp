// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package errors

// Trace
const (
	// 400
	Uniquery_Trace_InvalidParameter_EndTime         = "UniQuery.Trace.InvalidParameter.EndTime"
	Uniquery_Trace_InvalidParameter_LogDataViewID   = "UniQuery.Trace.InvalidParameter.LogDataViewID"
	Uniquery_Trace_InvalidParameter_SpanStatuses    = "UniQuery.Trace.InvalidParameter.SpanStatuses"
	Uniquery_Trace_InvalidParameter_StartTime       = "UniQuery.Trace.InvalidParameter.StartTime"
	Uniquery_Trace_InvalidParameter_TraceDataViewID = "UniQuery.Trace.InvalidParameter.TraceDataViewID"
	Uniquery_Trace_MissingParameter_LogDataViewID   = "UniQuery.Trace.MissingParameter.LogDataViewID"
	Uniquery_Trace_MissingParameter_TraceDataViewID = "UniQuery.Trace.MissingParameter.TraceDataViewID"

	// 404
	Uniquery_Trace_LogDataViewNotFound   = "UniQuery.Trace.LogDataViewNotFound"
	Uniquery_Trace_TraceDataViewNotFound = "UniQuery.Trace.TraceDataViewNotFound"
	Uniquery_Trace_TraceNotFound         = "UniQuery.Trace.TraceNotFound"
)

var (
	traceErrCodeList = []string{
		// 400
		Uniquery_Trace_InvalidParameter_EndTime,
		Uniquery_Trace_InvalidParameter_LogDataViewID,
		Uniquery_Trace_InvalidParameter_SpanStatuses,
		Uniquery_Trace_InvalidParameter_StartTime,
		Uniquery_Trace_InvalidParameter_TraceDataViewID,
		Uniquery_Trace_MissingParameter_LogDataViewID,
		Uniquery_Trace_MissingParameter_TraceDataViewID,

		// 404
		Uniquery_Trace_LogDataViewNotFound,
		Uniquery_Trace_TraceDataViewNotFound,
		Uniquery_Trace_TraceNotFound,
	}
)
