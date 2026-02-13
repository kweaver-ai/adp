// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package errors

// 指标模型
const (
	// 400
	Uniquery_EventModel_InvalidParameter = "Uniquery.EventModel.InvalidParameter"

	// 404
	Uniquery_EventModel_EventModelNotFound = "Uniquery.EventModel.EventModelNotFound"
	Uniquery_EventModel_EventNotFound      = "Uniquery.EventModel.EventNotFound"

	// 500
	Uniquery_EventModel_InternalError                         = "Uniquery.EventModel.InternalError"
	Uniquery_EventModel_InternalError_GetEventModelByIdFailed = "Uniquery.EventModel.InternalError.GetEventModelByIdFailed"
)

var (
	eventModelErrCodeList = []string{
		// 400
		Uniquery_EventModel_InvalidParameter,

		// 404
		Uniquery_EventModel_EventModelNotFound,
		Uniquery_EventModel_EventNotFound,

		// 500
		Uniquery_EventModel_InternalError,
		Uniquery_EventModel_InternalError_GetEventModelByIdFailed,
	}
)
