// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package errors 服务错误码
package errors

import (
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"uniquery/locale"
)

// 公共错误码
const (
	// 400
	Uniquery_InvalidParameter                   = "Uniquery.InvalidParameter"
	Uniquery_InvalidParameter_OverrideMethod    = "Uniquery.InvalidParameter.OverrideMethod"
	Uniquery_InvalidParameter_RequestBody       = "Uniquery.InvalidParameter.RequestBody"
	Uniquery_InvalidParameter_Start             = "Uniquery.InvalidParameter.Start"
	Uniquery_InvalidParameter_End               = "Uniquery.InvalidParameter.End"
	Uniquery_InvalidParameter_Offset            = "UniQuery.InvalidParameter.Offset"
	Uniquery_InvalidParameter_Limit             = "UniQuery.InvalidParameter.Limit"
	Uniquery_InvalidParameter_Sort              = "UniQuery.InvalidParameter.Sort"
	Uniquery_InvalidParameter_Direction         = "UniQuery.InvalidParameter.Direction"
	Uniquery_InvalidParameter_OffestAndLimitSum = "Uniquery.InvalidParameter.OffestAndLimitSum"
	Uniquery_InvalidParameter_Filter            = "Uniquery.InvalidParameter.Filter"
	Uniquery_InvalidParameter_FilterName        = "Uniquery.InvalidParameter.FilterName"
	Uniquery_InvalidParameter_FilterValue       = "Uniquery.InvalidParameter.FilterValue"
	Uniquery_InvalidParameter_ValueFrom         = "Uniquery.InvalidParameter.ValueFrom"

	Uniquery_NullParameter_OverrideMethod  = "Uniquery.NullParameter.OverrideMethod"
	Uniquery_NullParameter_FieldName       = "Uniquery.NullParameter.FieldName"
	Uniquery_NullParameter_FilterName      = "Uniquery.NullParameter.FilterName"
	Uniquery_NullParameter_FilterOperation = "Uniquery.NullParameter.FilterOperation"
	Uniquery_NullParameter_FilterValue     = "Uniquery.NullParameter.FilterValue"
	Uniquery_UnsupportFilterOperation      = "Uniquery.UnsupportFilterOperation"

	// 401
	Uniquery_InvalidRequestHeader_Authorization = "Uniquery.InvalidRequestHeader.Authorization"

	// 403
	Uniquery_Forbidden_FilterField = "Uniquery.Forbidden.FilterField"

	// 406
	Uniquery_InvalidRequestHeader_ContentType = "Uniquery.InvalidRequestHeader.ContentType"

	// 500
	Uniquery_InternalError                     = "UniQuery.InternalError"
	Uniquery_InternalError_AssertFloat64Failed = "Uniquery.InternalError.AssertFloat64Failed"
	Uniquery_InternalError_CountFailed         = "Uniquery.InternalError.CountFailed"
	Uniquery_InternalError_ScrollFailed        = "Uniquery.InternalError.ScrollFailed"
	Uniquery_InternalError_SearchSubmitFailed  = "Uniquery.InternalError.SearchSubmitFailed"

	// permission
	Uniquery_InternalError_CheckPermissionFailed        = "Uniquery.InternalError.CheckPermissionFailed"
	Uniquery_InternalError_FilterResourcesFailed        = "Uniquery.InternalError.FilterResourcesFailed"
	Uniquery_InternalError_GetResourcesOperationsFailed = "Uniquery.InternalError.GetResourcesOperationsFailed"
)

var (
	errCodeList = []string{
		// 公共错误码
		// 400
		Uniquery_InvalidParameter,
		Uniquery_InvalidParameter_OverrideMethod,
		Uniquery_InvalidParameter_RequestBody,
		Uniquery_InvalidParameter_Start,
		Uniquery_InvalidParameter_End,
		Uniquery_InvalidParameter_Offset,
		Uniquery_InvalidParameter_Limit,
		Uniquery_InvalidParameter_Sort,
		Uniquery_InvalidParameter_Direction,
		Uniquery_InvalidParameter_OffestAndLimitSum,
		Uniquery_InvalidParameter_FilterName,
		Uniquery_InvalidParameter_FilterValue,
		Uniquery_InvalidParameter_ValueFrom,
		Uniquery_InvalidParameter_Filter,
		Uniquery_NullParameter_OverrideMethod,
		Uniquery_NullParameter_FieldName,
		Uniquery_NullParameter_FilterName,
		Uniquery_NullParameter_FilterOperation,
		Uniquery_NullParameter_FilterValue,
		Uniquery_UnsupportFilterOperation,

		// 401
		Uniquery_InvalidRequestHeader_Authorization,

		// 403
		Uniquery_Forbidden_FilterField,

		// 406
		Uniquery_InvalidRequestHeader_ContentType,

		// 500
		Uniquery_InternalError,
		Uniquery_InternalError_AssertFloat64Failed,
		Uniquery_InternalError_CountFailed,
		Uniquery_InternalError_ScrollFailed,
		Uniquery_InternalError_SearchSubmitFailed,

		// permission
		Uniquery_InternalError_CheckPermissionFailed,
		Uniquery_InternalError_FilterResourcesFailed,
		Uniquery_InternalError_GetResourcesOperationsFailed,
	}
)

func init() {
	locale.Register()
	rest.Register(errCodeList)
	rest.Register(dataViewErrCodeList)
	rest.Register(eventModelErrCodeList)
	rest.Register(logGroupErrCodeList)
	rest.Register(metricModelErrCodeList)
	rest.Register(objectiveModelErrCodeList)
	rest.Register(traceErrCodeList)
	rest.Register(traceModelErrCodeList)
}
