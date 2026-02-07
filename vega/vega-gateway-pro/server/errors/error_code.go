// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package errors 服务错误码
package errors

import (
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"vega-gateway-pro/locale"
)

// 公共错误码
const (
	// 400
	InvalidParameter                   = "vega-gateway-pro.InvalidParameter"
	InvalidParameter_OverrideMethod    = "vega-gateway-pro.InvalidParameter.OverrideMethod"
	InvalidParameter_RequestBody       = "vega-gateway-pro.InvalidParameter.RequestBody"
	InvalidParameter_Start             = "vega-gateway-pro.InvalidParameter.Start"
	InvalidParameter_End               = "vega-gateway-pro.InvalidParameter.End"
	InvalidParameter_Offset            = "vega-gateway-pro.InvalidParameter.Offset"
	InvalidParameter_Limit             = "vega-gateway-pro.InvalidParameter.Limit"
	InvalidParameter_Sort              = "vega-gateway-pro.InvalidParameter.Sort"
	InvalidParameter_Direction         = "vega-gateway-pro.InvalidParameter.Direction"
	InvalidParameter_OffestAndLimitSum = "vega-gateway-pro.InvalidParameter.OffestAndLimitSum"
	InvalidParameter_Filter            = "vega-gateway-pro.InvalidParameter.Filter"
	InvalidParameter_FilterName        = "vega-gateway-pro.InvalidParameter.FilterName"
	InvalidParameter_FilterValue       = "vega-gateway-pro.InvalidParameter.FilterValue"
	InvalidParameter_ValueFrom         = "vega-gateway-pro.InvalidParameter.ValueFrom"

	NullParameter_OverrideMethod  = "vega-gateway-pro.NullParameter.OverrideMethod"
	NullParameter_FieldName       = "vega-gateway-pro.NullParameter.FieldName"
	NullParameter_FilterName      = "vega-gateway-pro.NullParameter.FilterName"
	NullParameter_FilterOperation = "vega-gateway-pro.NullParameter.FilterOperation"
	NullParameter_FilterValue     = "vega-gateway-pro.NullParameter.FilterValue"
	UnsupportFilterOperation      = "vega-gateway-pro.UnsupportFilterOperation"

	// 401
	InvalidRequestHeader_Authorization = "vega-gateway-pro.InvalidRequestHeader.Authorization"

	// 403
	Forbidden_FilterField = "vega-gateway-pro.Forbidden.FilterField"

	// 406
	InvalidRequestHeader_ContentType = "vega-gateway-pro.InvalidRequestHeader.ContentType"

	// 500
	InternalError                    = "vega-gateway-pro.InternalError"
	InternalError_CountFailed        = "vega-gateway-pro.InternalError.CountFailed"
	InternalError_ScrollFailed       = "vega-gateway-pro.InternalError.ScrollFailed"
	InternalError_SearchSubmitFailed = "vega-gateway-pro.InternalError.SearchSubmitFailed"

	// permission
	InternalError_CheckPermissionFailed        = "vega-gateway-pro.InternalError.CheckPermissionFailed"
	InternalError_FilterResourcesFailed        = "vega-gateway-pro.InternalError.FilterResourcesFailed"
	InternalError_GetResourcesOperationsFailed = "vega-gateway-pro.InternalError.GetResourcesOperationsFailed"
)

var (
	errCodeList = []string{
		// 公共错误码
		// 400
		InvalidParameter,
		InvalidParameter_OverrideMethod,
		InvalidParameter_RequestBody,
		InvalidParameter_Start,
		InvalidParameter_End,
		InvalidParameter_Offset,
		InvalidParameter_Limit,
		InvalidParameter_Sort,
		InvalidParameter_Direction,
		InvalidParameter_OffestAndLimitSum,
		InvalidParameter_FilterName,
		InvalidParameter_FilterValue,
		InvalidParameter_ValueFrom,
		InvalidParameter_Filter,
		NullParameter_OverrideMethod,
		NullParameter_FieldName,
		NullParameter_FilterName,
		NullParameter_FilterOperation,
		NullParameter_FilterValue,
		UnsupportFilterOperation,

		// 401
		InvalidRequestHeader_Authorization,

		// 403
		Forbidden_FilterField,

		// 406
		InvalidRequestHeader_ContentType,

		// 500
		InternalError,
		InternalError_CountFailed,
		InternalError_ScrollFailed,
		InternalError_SearchSubmitFailed,

		// permission
		InternalError_CheckPermissionFailed,
		InternalError_FilterResourcesFailed,
		InternalError_GetResourcesOperationsFailed,
	}
)

func init() {
	locale.Register()
	rest.Register(errCodeList)
}
