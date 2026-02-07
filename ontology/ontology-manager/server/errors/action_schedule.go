// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package errors

const (
	// 400 Bad Request
	OntologyManager_ActionSchedule_InvalidParameter       = "OntologyManager.ActionSchedule.InvalidParameter"
	OntologyManager_ActionSchedule_InvalidCronExpression  = "OntologyManager.ActionSchedule.InvalidCronExpression"
	OntologyManager_ActionSchedule_InvalidStatus          = "OntologyManager.ActionSchedule.InvalidStatus"
	OntologyManager_ActionSchedule_ActionTypeNotFound     = "OntologyManager.ActionSchedule.ActionTypeNotFound"

	// 404 Not Found
	OntologyManager_ActionSchedule_NotFound = "OntologyManager.ActionSchedule.NotFound"

	// 500 Internal Server Error
	OntologyManager_ActionSchedule_CreateFailed        = "OntologyManager.ActionSchedule.CreateFailed"
	OntologyManager_ActionSchedule_UpdateFailed        = "OntologyManager.ActionSchedule.UpdateFailed"
	OntologyManager_ActionSchedule_DeleteFailed        = "OntologyManager.ActionSchedule.DeleteFailed"
	OntologyManager_ActionSchedule_GetFailed           = "OntologyManager.ActionSchedule.GetFailed"
	OntologyManager_ActionSchedule_GetActionTypeFailed = "OntologyManager.ActionSchedule.GetActionTypeFailed"
)

var (
	actionScheduleErrCodeList = []string{
		OntologyManager_ActionSchedule_InvalidParameter,
		OntologyManager_ActionSchedule_InvalidCronExpression,
		OntologyManager_ActionSchedule_InvalidStatus,
		OntologyManager_ActionSchedule_ActionTypeNotFound,
		OntologyManager_ActionSchedule_NotFound,
		OntologyManager_ActionSchedule_CreateFailed,
		OntologyManager_ActionSchedule_UpdateFailed,
		OntologyManager_ActionSchedule_DeleteFailed,
		OntologyManager_ActionSchedule_GetFailed,
		OntologyManager_ActionSchedule_GetActionTypeFailed,
	}
)
