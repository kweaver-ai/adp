// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package errors 服务错误码
package errors

// 数据连接错误码
const (
	// 400
	DataModel_DataConnection_ConnectionNameExisted                   = "DataModel.DataConnection.ConnectionNameExisted"
	DataModel_DataConnection_DuplicatedParameter_Config              = "DataModel.DataConnection.DuplicatedParameter.Config"
	DataModel_DataConnection_ForbiddenUpdateParameter_DataSourceType = "DataModel.DataConnection.ForbiddenUpdateParameter.DataSourceType"
	DataModel_DataConnection_InvalidParameter_ConnectionIDs          = "DataModel.DataConnection.InvalidParameter.ConnectionIDs"
	DataModel_DataConnection_InvalidParameter_DataSourceType         = "DataModel.DataConnection.InvalidParameter.DataSourceType"
	DataModel_DataConnection_InvalidParameter_Protocol               = "DataModel.DataConnection.InvalidParameter.Protocol"
	DataModel_DataConnection_LengthExceeded_ConnectionName           = "DataModel.DataConnection.LengthExceeded.ConnectionName"
	DataModel_DataConnection_NullParameter_Address                   = "DataModel.DataConnection.NullParameter.Address"
	DataModel_DataConnection_NullParameter_ApiKey                    = "DataModel.DataConnection.NullParameter.ApiKey"
	DataModel_DataConnection_NullParameter_ConnectionName            = "DataModel.DataConnection.NullParameter.ConnectionName"
	DataModel_DataConnection_NullParameter_SecretKey                 = "DataModel.DataConnection.NullParameter.SecretKey"

	// 404
	DataModel_DataConnection_DataConnectionNotFound = "DataModel.DataConnection.DataConnectionNotFound"

	// 500
	DataModel_DataConnection_InternalError_CreateDataConnectionFailed         = "DataModel.DataConnection.InternalError.CreateDataConnectionFailed"
	DataModel_DataConnection_InternalError_CreateDataConnectionStatusFailed   = "DataModel.DataConnection.InternalError.CreateDataConnectionStatusFailed"
	DataModel_DataConnection_InternalError_DeleteDataConnectionsFailed        = "DataModel.DataConnection.InternalError.DeleteDataConnectionsFailed"
	DataModel_DataConnection_InternalError_DeleteDataConnectionStatusesFailed = "DataModel.DataConnection.InternalError.DeleteDataConnectionStatusesFailed"
	DataModel_DataConnection_InternalError_GetAccessTokenFailed               = "DataModel.DataConnection.InternalError.GetAccessTokenFailed"
	DataModel_DataConnection_InternalError_GetDataConnectionsFailed           = "DataModel.DataConnection.InternalError.GetDataConnectionsFailed"
	DataModel_DataConnection_InternalError_GetDataConnectionSourceTypeFailed  = "DataModel.DataConnection.InternalError.GetGetDataConnectionSourceTypeFailed"
	DataModel_DataConnection_InternalError_GetDataConnectionTotalFailed       = "DataModel.DataConnection.InternalError.GetDataConnectionTotalFailed"
	DataModel_DataConnection_InternalError_GetMapAboutID2NameFailed           = "DataModel.DataConnection.InternalError.GetMapAboutID2NameFailed"
	DataModel_DataConnection_InternalError_GetMapAboutName2IDFailed           = "DataModel.DataConnection.InternalError.GetMapAboutName2IDFailed"
	DataModel_DataConnection_InternalError_InitDataConnectionProcessor        = "DataModel.DataConnection.InternalError.InitDataConnectionProcessor"
	DataModel_DataConnection_InternalError_ListDataConnectionsFailed          = "DataModel.DataConnection.InternalError.ListDataConnectionsFailed"
	DataModel_DataConnection_InternalError_UpdateDataConnectionFailed         = "DataModel.DataConnection.InternalError.UpdateDataConnectionFailed"
	DataModel_DataConnection_InternalError_UpdateDataConnectionStatusFailed   = "DataModel.DataConnection.InternalError.UpdateDataConnectionStatusFailed"
	DataModel_DataConnection_InternalError_VerifyConnectivityFailed           = "DataModel.DataConnection.InternalError.VerifyConnectivityFailed"
)

var (
	dataConnectionErrCodeList = []string{
		// 400
		DataModel_DataConnection_ConnectionNameExisted,
		DataModel_DataConnection_DuplicatedParameter_Config,
		DataModel_DataConnection_ForbiddenUpdateParameter_DataSourceType,
		DataModel_DataConnection_InvalidParameter_ConnectionIDs,
		DataModel_DataConnection_InvalidParameter_DataSourceType,
		DataModel_DataConnection_InvalidParameter_Protocol,
		DataModel_DataConnection_LengthExceeded_ConnectionName,
		DataModel_DataConnection_NullParameter_Address,
		DataModel_DataConnection_NullParameter_ApiKey,
		DataModel_DataConnection_NullParameter_ConnectionName,
		DataModel_DataConnection_NullParameter_SecretKey,

		// 404
		DataModel_DataConnection_DataConnectionNotFound,

		// 500
		DataModel_DataConnection_InternalError_CreateDataConnectionFailed,
		DataModel_DataConnection_InternalError_CreateDataConnectionStatusFailed,
		DataModel_DataConnection_InternalError_DeleteDataConnectionsFailed,
		DataModel_DataConnection_InternalError_DeleteDataConnectionStatusesFailed,
		DataModel_DataConnection_InternalError_GetAccessTokenFailed,
		DataModel_DataConnection_InternalError_GetDataConnectionsFailed,
		DataModel_DataConnection_InternalError_GetDataConnectionSourceTypeFailed,
		DataModel_DataConnection_InternalError_GetDataConnectionTotalFailed,
		DataModel_DataConnection_InternalError_GetMapAboutID2NameFailed,
		DataModel_DataConnection_InternalError_GetMapAboutName2IDFailed,
		DataModel_DataConnection_InternalError_InitDataConnectionProcessor,
		DataModel_DataConnection_InternalError_ListDataConnectionsFailed,
		DataModel_DataConnection_InternalError_UpdateDataConnectionFailed,
		DataModel_DataConnection_InternalError_UpdateDataConnectionStatusFailed,
		DataModel_DataConnection_InternalError_VerifyConnectivityFailed,
	}
)
