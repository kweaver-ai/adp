// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package errors Catalog 模块错误码
package errors

// Catalog 错误码
const (
	// 400 Bad Request
	VegaManager_Catalog_InvalidParameter                 = "VegaManager.Catalog.InvalidParameter"
	VegaManager_Catalog_InvalidParameter_Type            = "VegaManager.Catalog.InvalidParameter.Type"
	VegaManager_Catalog_InvalidParameter_Name            = "VegaManager.Catalog.InvalidParameter.Name"
	VegaManager_Catalog_InvalidParameter_ConnectorType   = "VegaManager.Catalog.InvalidParameter.ConnectorType"
	VegaManager_Catalog_InvalidParameter_ConnectorConfig = "VegaManager.Catalog.InvalidParameter.ConnectorConfig"
	VegaManager_Catalog_LengthExceeded_Name              = "VegaManager.Catalog.LengthExceeded.Name"
	VegaManager_Catalog_LengthExceeded_Description       = "VegaManager.Catalog.LengthExceeded.Description"

	// 403 Forbidden
	VegaManager_Catalog_NotFound   = "VegaManager.Catalog.NotFound"
	VegaManager_Catalog_NameExists = "VegaManager.Catalog.NameExists"
	VegaManager_Catalog_HasAssets  = "VegaManager.Catalog.HasAssets"
	VegaManager_Catalog_IsDisabled = "VegaManager.Catalog.IsDisabled"

	// 500 Internal Server Error
	VegaManager_Catalog_InternalError                      = "VegaManager.Catalog.InternalError"
	VegaManager_Catalog_InternalError_CreateFailed         = "VegaManager.Catalog.InternalError.CreateFailed"
	VegaManager_Catalog_InternalError_GetFailed            = "VegaManager.Catalog.InternalError.GetFailed"
	VegaManager_Catalog_InternalError_UpdateFailed         = "VegaManager.Catalog.InternalError.UpdateFailed"
	VegaManager_Catalog_InternalError_DeleteFailed         = "VegaManager.Catalog.InternalError.DeleteFailed"
	VegaManager_Catalog_InternalError_TestConnectionFailed = "VegaManager.Catalog.InternalError.TestConnectionFailed"
	VegaManager_Catalog_InternalError_EncryptFailed        = "VegaManager.Catalog.InternalError.EncryptFailed"

	VegaManager_Catalog_InvalidParameter_SensitiveFieldNotEncrypted = "VegaManager.Catalog.InvalidParameter.SensitiveFieldNotEncrypted"
)

var CatalogErrCodeList = []string{
	VegaManager_Catalog_InvalidParameter,
	VegaManager_Catalog_InvalidParameter_Type,
	VegaManager_Catalog_InvalidParameter_Name,
	VegaManager_Catalog_InvalidParameter_ConnectorType,
	VegaManager_Catalog_InvalidParameter_ConnectorConfig,
	VegaManager_Catalog_LengthExceeded_Name,
	VegaManager_Catalog_LengthExceeded_Description,
	VegaManager_Catalog_NotFound,
	VegaManager_Catalog_NameExists,
	VegaManager_Catalog_HasAssets,
	VegaManager_Catalog_IsDisabled,
	VegaManager_Catalog_InternalError,
	VegaManager_Catalog_InternalError_CreateFailed,
	VegaManager_Catalog_InternalError_GetFailed,
	VegaManager_Catalog_InternalError_UpdateFailed,
	VegaManager_Catalog_InternalError_DeleteFailed,
	VegaManager_Catalog_InternalError_TestConnectionFailed,
	VegaManager_Catalog_InternalError_EncryptFailed,
	VegaManager_Catalog_InvalidParameter_SensitiveFieldNotEncrypted,
}
