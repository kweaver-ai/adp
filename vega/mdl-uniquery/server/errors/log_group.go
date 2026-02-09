// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package errors

// 日志分组
const (
	// 400
	Uniquery_LogGroup_InvalidParameter_DataConnectionID = "Uniquery.LogGroup.InvalidParameter.DataConnectionID"
	Uniquery_LogGroup_InvalidParameter_LogGroupID       = "Uniquery.LogGroup.InvalidParameter.LogGroupID"
	Uniquery_LogGroup_InvalidParameter_UserID           = "Uniquery.LogGroup.InvalidParameter.UserID"
	Uniquery_LogGroup_InvalidParameter_JobID            = "Uniquery.LogGroup.InvalidParameter.JobID"

	// 404
	Uniquery_LogGroup_DataConnectionNotFound = "Uniquery.LogGroup.DataConnectionNotFound"

	// 500
	Uniquery_LogGroup_InternalError = "Uniquery.LogGroup.InternalError"
)

var (
	logGroupErrCodeList = []string{
		// 400
		Uniquery_LogGroup_InvalidParameter_DataConnectionID,
		Uniquery_LogGroup_InvalidParameter_LogGroupID,
		Uniquery_LogGroup_InvalidParameter_UserID,
		Uniquery_LogGroup_InvalidParameter_JobID,

		// 404
		Uniquery_LogGroup_DataConnectionNotFound,

		// 500
		Uniquery_LogGroup_InternalError,
	}
)
