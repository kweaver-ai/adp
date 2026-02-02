// Package errors ConnectorType 模块错误码
package errors

// ConnectorType 错误码
const (
	// 400 Bad Request
	VegaManager_ConnectorType_InvalidParameter          = "VegaManager.ConnectorType.InvalidParameter"
	VegaManager_ConnectorType_InvalidParameter_Type     = "VegaManager.ConnectorType.InvalidParameter.Type"
	VegaManager_ConnectorType_InvalidParameter_Name     = "VegaManager.ConnectorType.InvalidParameter.Name"
	VegaManager_ConnectorType_InvalidParameter_Mode     = "VegaManager.ConnectorType.InvalidParameter.Mode"
	VegaManager_ConnectorType_InvalidParameter_Category = "VegaManager.ConnectorType.InvalidParameter.Category"
	VegaManager_ConnectorType_InvalidParameter_Endpoint = "VegaManager.ConnectorType.InvalidParameter.Endpoint"
	VegaManager_ConnectorType_BadRequest                = "VegaManager.ConnectorType.BadRequest"

	// 404 Not Found / 409 Conflict
	VegaManager_ConnectorType_NotFound   = "VegaManager.ConnectorType.NotFound"
	VegaManager_ConnectorType_TypeExists = "VegaManager.ConnectorType.TypeExists"
	VegaManager_ConnectorType_NameExists = "VegaManager.ConnectorType.NameExists"

	// 500 Internal Server Error
	VegaManager_ConnectorType_InternalError                = "VegaManager.ConnectorType.InternalError"
	VegaManager_ConnectorType_InternalError_RegisterFailed = "VegaManager.ConnectorType.InternalError.RegisterFailed"
	VegaManager_ConnectorType_InternalError_GetFailed      = "VegaManager.ConnectorType.InternalError.GetFailed"
	VegaManager_ConnectorType_InternalError_UpdateFailed   = "VegaManager.ConnectorType.InternalError.UpdateFailed"
	VegaManager_ConnectorType_InternalError_DeleteFailed   = "VegaManager.ConnectorType.InternalError.DeleteFailed"
)

var ConnectorTypeErrCodeList = []string{
	VegaManager_ConnectorType_InvalidParameter,
	VegaManager_ConnectorType_InvalidParameter_Type,
	VegaManager_ConnectorType_InvalidParameter_Name,
	VegaManager_ConnectorType_InvalidParameter_Mode,
	VegaManager_ConnectorType_InvalidParameter_Category,
	VegaManager_ConnectorType_InvalidParameter_Endpoint,
	VegaManager_ConnectorType_BadRequest,
	VegaManager_ConnectorType_NotFound,
	VegaManager_ConnectorType_NameExists,
	VegaManager_ConnectorType_TypeExists,
	VegaManager_ConnectorType_InternalError,
	VegaManager_ConnectorType_InternalError_RegisterFailed,
	VegaManager_ConnectorType_InternalError_GetFailed,
	VegaManager_ConnectorType_InternalError_UpdateFailed,
	VegaManager_ConnectorType_InternalError_DeleteFailed,
}
