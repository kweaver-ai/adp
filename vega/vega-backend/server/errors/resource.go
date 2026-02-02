// Package errors Resource 模块错误码
package errors

// Resource 错误码
const (
	// 400 Bad Request
	VegaManager_Resource_InvalidParameter           = "VegaManager.Resource.InvalidParameter"
	VegaManager_Resource_InvalidParameter_Type      = "VegaManager.Resource.InvalidParameter.Type"
	VegaManager_Resource_InvalidParameter_Name      = "VegaManager.Resource.InvalidParameter.Name"
	VegaManager_Resource_InvalidParameter_CatalogID = "VegaManager.Resource.InvalidParameter.CatalogID"
	VegaManager_Resource_LengthExceeded_Name        = "VegaManager.Resource.LengthExceeded.Name"
	VegaManager_Resource_LengthExceeded_Description = "VegaManager.Resource.LengthExceeded.Description"

	// 403 Forbidden
	VegaManager_Resource_NotFound        = "VegaManager.Resource.NotFound"
	VegaManager_Resource_NameExists      = "VegaManager.Resource.NameExists"
	VegaManager_Resource_CatalogNotFound = "VegaManager.Resource.CatalogNotFound"
	VegaManager_Resource_IsDisabled      = "VegaManager.Resource.IsDisabled"
	VegaManager_Resource_AlreadyEnabled  = "VegaManager.Resource.AlreadyEnabled"
	VegaManager_Resource_AlreadyDisabled = "VegaManager.Resource.AlreadyDisabled"

	// 500 Internal Server Error
	VegaManager_Resource_InternalError              = "VegaManager.Resource.InternalError"
	VegaManager_Resource_InternalError_CreateFailed = "VegaManager.Resource.InternalError.CreateFailed"
	VegaManager_Resource_InternalError_GetFailed    = "VegaManager.Resource.InternalError.GetFailed"
	VegaManager_Resource_InternalError_UpdateFailed = "VegaManager.Resource.InternalError.UpdateFailed"
	VegaManager_Resource_InternalError_DeleteFailed = "VegaManager.Resource.InternalError.DeleteFailed"
	VegaManager_Resource_InternalError_SyncFailed   = "VegaManager.Resource.InternalError.SyncFailed"
)

var ResourceErrCodeList = []string{
	VegaManager_Resource_InvalidParameter,
	VegaManager_Resource_InvalidParameter_Type,
	VegaManager_Resource_InvalidParameter_Name,
	VegaManager_Resource_InvalidParameter_CatalogID,
	VegaManager_Resource_LengthExceeded_Name,
	VegaManager_Resource_LengthExceeded_Description,
	VegaManager_Resource_NotFound,
	VegaManager_Resource_NameExists,
	VegaManager_Resource_CatalogNotFound,
	VegaManager_Resource_IsDisabled,
	VegaManager_Resource_AlreadyEnabled,
	VegaManager_Resource_AlreadyDisabled,
	VegaManager_Resource_InternalError,
	VegaManager_Resource_InternalError_CreateFailed,
	VegaManager_Resource_InternalError_GetFailed,
	VegaManager_Resource_InternalError_UpdateFailed,
	VegaManager_Resource_InternalError_DeleteFailed,
	VegaManager_Resource_InternalError_SyncFailed,
}
