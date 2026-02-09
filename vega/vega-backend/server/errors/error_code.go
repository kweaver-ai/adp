// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package errors 服务错误码
package errors

import (
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"vega-backend/locale"
)

// 公共错误码
const (
	// 400 Bad Request
	VegaManager_InvalidParameter_Name        = "VegaManager.InvalidParameter.Name"
	VegaManager_InvalidParameter_Description = "VegaManager.InvalidParameter.Description"
	VegaManager_InvalidParameter_Tag         = "VegaManager.InvalidParameter.Tag"
	VegaManager_InvalidParameter_RequestBody = "VegaManager.InvalidParameter.RequestBody"
	VegaManager_InvalidParameter_Limit       = "VegaManager.InvalidParameter.Limit"
	VegaManager_InvalidParameter_Offset      = "VegaManager.InvalidParameter.Offset"
	VegaManager_InvalidParameter_Sort        = "VegaManager.InvalidParameter.Sort"
	VegaManager_InvalidParameter_Direction   = "VegaManager.InvalidParameter.Direction"
	VegaManager_InvalidParameter_ID          = "VegaManager.InvalidParameter.ID"

	// 406 Not Acceptable
	VegaManager_InvalidRequestHeader_ContentType = "VegaManager.InvalidRequestHeader.ContentType"

	// 500 Internal Server Error
	VegaManager_InternalError_BeginTransactionFailed  = "VegaManager.InternalError.BeginTransactionFailed"
	VegaManager_InternalError_CommitTransactionFailed = "VegaManager.InternalError.CommitTransactionFailed"
	VegaManager_InternalError_MarshalDataFailed       = "VegaManager.InternalError.MarshalDataFailed"
	VegaManager_InternalError_UnMarshalDataFailed     = "VegaManager.InternalError.UnMarshalDataFailed"
)

var (
	commonErrCodeList = []string{
		VegaManager_InvalidParameter_Name,
		VegaManager_InvalidParameter_Description,
		VegaManager_InvalidParameter_Tag,

		VegaManager_InvalidParameter_RequestBody,
		VegaManager_InvalidParameter_Limit,
		VegaManager_InvalidParameter_Offset,
		VegaManager_InvalidParameter_Sort,
		VegaManager_InvalidParameter_Direction,
		VegaManager_InvalidParameter_ID,
		VegaManager_InvalidRequestHeader_ContentType,

		VegaManager_InternalError_BeginTransactionFailed,
		VegaManager_InternalError_CommitTransactionFailed,
		VegaManager_InternalError_MarshalDataFailed,
		VegaManager_InternalError_UnMarshalDataFailed,
	}
)

func init() {
	locale.Register()
	rest.Register(commonErrCodeList)
	rest.Register(CatalogErrCodeList)
	rest.Register(ResourceErrCodeList)
	rest.Register(ConnectorTypeErrCodeList)
	rest.Register(TaskErrCodeList)
}
