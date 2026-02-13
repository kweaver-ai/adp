// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package errors 服务错误码
package errors

import (
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"data-model/locale"
)

// 公共错误码, 服务内所有模块均可使用
const (
	// 400
	DataModel_ConflictParameter_NameAndNamePatternCoexist = "DataModel.ConflictParameter.NameAndNamePatternCoexist"
	DataModel_CountExceeded_Tags                          = "DataModel.CountExceeded.Tags"
	DataModel_InvalidParameter_Direction                  = "DataModel.InvalidParameter.Direction"
	DataModel_InvalidParameter_Filters                    = "DataModel.InvalidParameter.Filters"
	DataModel_InvalidParameter_FilterValue                = "DataModel.InvalidParameter.FilterValue"
	DataModel_InvalidParameter_ID                         = "DataModel.InvalidParameter.ID"
	DataModel_InvalidParameter_Limit                      = "DataModel.InvalidParameter.Limit"
	DataModel_InvalidParameter_ModuleType                 = "DataModel.InvalidParameter.ModuleType"
	DataModel_InvalidParameter_Offset                     = "DataModel.InvalidParameter.Offset"
	DataModel_InvalidParameter_OverrideMethod             = "DataModel.InvalidParameter.OverrideMethod"
	DataModel_InvalidParameter_RequestBody                = "DataModel.InvalidParameter.RequestBody"
	DataModel_InvalidParameter_SimpleInfo                 = "DataModel.InvalidParameter.SimpleInfo"
	DataModel_InvalidParameter_Sort                       = "DataModel.InvalidParameter.Sort"
	DataModel_InvalidParameter_Tag                        = "DataModel.InvalidParameter.Tag"
	DataModel_InvalidParameter_Tags                       = "DataModel.InvalidParameter.Tags"
	DataModel_InvalidParameter_ValueFrom                  = "DataModel.InvalidParameter.ValueFrom"
	DataModel_LengthExceeded_Comment                      = "DataModel.LengthExceeded.Comment"
	DataModel_LengthExceeded_Tag                          = "DataModel.LengthExceeded.Tag"
	DataModel_NullParameter_FilterName                    = "DataModel.NullParameter.FilterName"
	DataModel_NullParameter_FilterOperation               = "DataModel.NullParameter.FilterOperation"
	DataModel_NullParameter_FilterValue                   = "DataModel.NullParameter.FilterValue"
	DataModel_NullParameter_Tag                           = "DataModel.NullParameter.Tag"
	DataModel_UnsupportFilterOperation                    = "DataModel.UnsupportFilterOperation"

	// 403
	DataModel_Forbidden_FilterField = "DataModel.Forbidden.FilterField"

	// 406
	DataModel_InvalidRequestHeader_ContentType = "DataModel.InvalidRequestHeader.ContentType"

	// 500
	DataModel_InternalError                         = "DataModel.InternalError"
	DataModel_InternalError_BeginTransactionFailed  = "DataModel.InternalError.BeginTransactionFailed"
	DataModel_InternalError_CommitTransactionFailed = "DataModel.InternalError.CommitTransactionFailed"
	DataModel_InternalError_DataBaseError           = "DataModel.InternalError.DataBaseError"
	DataModel_InternalError_MarshalDataFailed       = "DataModel.InternalError.MarshalDataFailed"
	DataModel_InternalError_UnmarshalDataFailed     = "DataModel.InternalError.UnmarshalDataFailed"
	DataModel_InternalError_StartJobFailed          = "DataModel.InternalError.StartJobFailed"
	DataModel_InternalError_StopJobFailed           = "DataModel.InternalError.StopJobFailed"

	// Permission
	DataModel_InternalError_CheckPermissionFailed = "DataModel.InternalError.CheckPermissionFailed"
	DataModel_InternalError_CreateResourcesFailed = "DataModel.InternalError.CreateResourcesFailed"
	DataModel_InternalError_DeleteResourcesFailed = "DataModel.InternalError.DeleteResourcesFailed"
	DataModel_InternalError_FilterResourcesFailed = "DataModel.InternalError.FilterResourcesFailed"
	DataModel_InternalError_UpdateResourceFailed  = "DataModel.InternalError.UpdateResourceFailed"
	DataModel_InternalError_MQPublishMsgFailed    = "DataModel.InternalError.MQPublishMsgFailed"
)

// 数据标签模块错误码
const (
	//400
	DataModel_InvalidParameter_DataTagName = "DataModel.InvalidParameter.DataTagName"

	//404
	DataModel_DataTagNotFound = "DataModel.DataTagNotFound"

	//500
	DataModel_InternalError_FailedToGetAllDataTagUserInfo = "DataModel.InternalError.FailedToGetAllDataTagUserInfo"
)

const (
	// 400
	EventModel_InvalidParameter                              = "EventModel.InvalidParameter"
	EventModel_DependentModelIllegal                         = "EventModel.DependentModelIllegal"
	EventModel_RunModeIllegal                                = "EventModel.RunModeIllegal"
	EventModel_TaskSyncCreateFailed                          = "EventModel.TaskSyncCreateFailed"
	EventModel_DataSourceIllegal                             = "EventModel.DataSourceIllegal"
	EventModel_ModelNameExisted                              = "EventModel.ModelNameExisted"
	EventModel_RefByOther                                    = "EventModel.RefByOther"
	DataModel_EventModel_NullParameter_Schedule              = "DataModel.EventModel.NullParameter.Schedule"
	DataModel_EventModel_NullParameter_ScheduleType          = "DataModel.EventModel.NullParameter.ScheduleType"
	DataModel_EventModel_InvalidParameter_ScheduleType       = "DataModel.EventModel.InvalidParameter.ScheduleType"
	DataModel_EventModel_NullParameter_ScheduleExpression    = "DataModel.EventModel.NullParameter.ScheduleExpression"
	DataModel_EventModel_InvalidParameter_ScheduleExpression = "DataModel.EventModel.InvalidParameter.ScheduleExpression"
	DataModel_EventModel_NullParameter_IndexBase             = "DataModel.EventModel.NullParameter.IndexBase"
	DataModel_EventModel_NullParameter_DataViewID            = "DataModel.EventModel.NullParameter.DataViewID"
	DataModel_EventModel_InvalidParameter_BlockStrategy      = "DataModel.EventModel.InvalidParameter.BlockStrategy"
	DataModel_EventModel_InvalidParameter_RouteStrategy      = "DataModel.EventModel.InvalidParameter.RouteStrategy"
	DataModel_EventModel_InvalidParameter_TimeOut            = "DataModel.EventModel.InvalidParameter.TimeOut"
	DataModel_EventModel_InvalidParameter_FailRetryCount     = "DataModel.EventModel.InvalidParameter.FailRetryCount"

	// 404
	EventModel_EventModelNotFound = "EventModel.EventModelNotFound"

	// 500
	EventModel_InternalError                                  = "EventModel.InternalError"
	EventModel_InternalError_GenerateIDFailed                 = "EventModel.InternalError.GenerateIDFailed"
	EventModel_InternalError_GetModelByIDFailed               = "EventModel.InternalError.GetModelByIDFailed"
	DataModel_EventModel_InternalError_BeginTransactionFailed = "DataModel.EventModel.InternalError.BeginTransactionFailed"
	DataModel_EventModel_InternalError_TaskBeingModified      = "DataModel.EventModel.InternalError.TaskBeingModified"
)

var (
	errCodeList = []string{
		// ---公共错误码---
		// 400
		DataModel_ConflictParameter_NameAndNamePatternCoexist,
		DataModel_CountExceeded_Tags,
		DataModel_InvalidParameter_Direction,
		DataModel_InvalidParameter_Filters,
		DataModel_InvalidParameter_FilterValue,
		DataModel_InvalidParameter_Limit,
		DataModel_InvalidParameter_ModuleType,
		DataModel_InvalidParameter_Offset,
		DataModel_InvalidParameter_OverrideMethod,
		DataModel_InvalidParameter_RequestBody,
		DataModel_InvalidParameter_SimpleInfo,
		DataModel_InvalidParameter_Sort,
		DataModel_InvalidParameter_Tag,
		DataModel_InvalidParameter_Tags,
		DataModel_InvalidParameter_ValueFrom,
		DataModel_LengthExceeded_Comment,
		DataModel_LengthExceeded_Tag,
		DataModel_NullParameter_FilterName,
		DataModel_NullParameter_FilterOperation,
		DataModel_NullParameter_FilterValue,
		DataModel_NullParameter_Tag,
		DataModel_UnsupportFilterOperation,
		DataModel_InvalidParameter_ID,

		// 403
		DataModel_Forbidden_FilterField,

		// 406
		DataModel_InvalidRequestHeader_ContentType,

		// 500
		DataModel_InternalError,
		DataModel_InternalError_BeginTransactionFailed,
		DataModel_InternalError_CommitTransactionFailed,
		DataModel_InternalError_DataBaseError,
		DataModel_InternalError_MarshalDataFailed,
		DataModel_InternalError_UnmarshalDataFailed,
		DataModel_InternalError_StartJobFailed,
		DataModel_InternalError_StopJobFailed,

		// permission
		DataModel_InternalError_CheckPermissionFailed,
		DataModel_InternalError_CreateResourcesFailed,
		DataModel_InternalError_DeleteResourcesFailed,
		DataModel_InternalError_FilterResourcesFailed,
		DataModel_InternalError_UpdateResourceFailed,
		DataModel_InternalError_MQPublishMsgFailed,

		// ---数据标签模块---
		// 400
		DataModel_InvalidParameter_DataTagName,

		// 404
		DataModel_DataTagNotFound,

		// 500
		DataModel_InternalError_FailedToGetAllDataTagUserInfo,

		// eventmodel
		//400
		EventModel_InvalidParameter,
		EventModel_DependentModelIllegal,
		EventModel_RunModeIllegal,
		EventModel_TaskSyncCreateFailed,
		EventModel_DataSourceIllegal,
		EventModel_ModelNameExisted,
		EventModel_RefByOther,
		DataModel_EventModel_NullParameter_Schedule,
		DataModel_EventModel_NullParameter_ScheduleType,
		DataModel_EventModel_InvalidParameter_ScheduleType,
		DataModel_EventModel_NullParameter_ScheduleExpression,
		DataModel_EventModel_InvalidParameter_ScheduleExpression,
		DataModel_EventModel_NullParameter_IndexBase,
		DataModel_EventModel_NullParameter_DataViewID,
		DataModel_EventModel_InvalidParameter_BlockStrategy,
		DataModel_EventModel_InvalidParameter_RouteStrategy,
		DataModel_EventModel_InvalidParameter_TimeOut,
		DataModel_EventModel_InvalidParameter_FailRetryCount,

		// 404
		EventModel_EventModelNotFound,

		// 500
		EventModel_InternalError,
		EventModel_InternalError_GenerateIDFailed,
		EventModel_InternalError_GetModelByIDFailed,
		DataModel_EventModel_InternalError_BeginTransactionFailed,
		DataModel_EventModel_InternalError_TaskBeingModified,
	}
)

func init() {
	locale.Register()
	rest.Register(errCodeList)
	rest.Register(dataConnectionErrCodeList)
	rest.Register(dataDictErrCodeList)
	rest.Register(dataViewErrCodeList)
	rest.Register(metricModelErrCodeList)
	rest.Register(objectiveModelErrCodeList)
	rest.Register(traceModelErrCodeList)
}
