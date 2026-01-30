package errors

// 数据视图错误码
const (
	// 400
	DataModel_DataView_CountExceeded_Filters              = "DataModel.DataView.CountExceeded.Filters"
	DataModel_DataView_Duplicated_FieldDisplayName        = "DataModel.DataView.Duplicated.FieldDisplayName"
	DataModel_DataView_Duplicated_FieldFeatureName        = "DataModel.DataView.Duplicated.FieldFeatureName"
	DataModel_DataView_Duplicated_FieldName               = "DataModel.DataView.Duplicated.FieldName"
	DataModel_DataView_Duplicated_ViewID                  = "DataModel.DataView.Duplicated.ViewID"
	DataModel_DataView_Duplicated_ViewName                = "DataModel.DataView.Duplicated.ViewName"
	DataModel_DataView_FieldTypeConflict                  = "DataModel.DataView.FieldTypeConflict"
	DataModel_DataView_FilterFieldTypeMisMatchOperation   = "DataModel.DataView.FilterFieldTypeMisMatchOperation"
	DataModel_DataView_IndexbaseNotFound                  = "DataModel.DataView.IndexbaseNotFound"
	DataModel_DataView_InvalidParameter_AttributeFields   = "DataModel.DataView.InvalidParameter.AttributeFields"
	DataModel_DataView_InvalidParameter_Builtin           = "DataModel.DataView.InvalidParameter.Builtin"
	DataModel_DataView_InvalidParameter_DataScope         = "DataModel.DataView.InvalidParameter.DataScope"
	DataModel_DataView_InvalidParameter_DataSource        = "DataModel.DataView.InvalidParameter.DataSource"
	DataModel_DataView_InvalidParameter_FieldFeatureName  = "DataModel.DataView.InvalidParameter.FieldFeatureName"
	DataModel_DataView_InvalidParameter_FieldName         = "DataModel.DataView.InvalidParameter.FieldName"
	DataModel_DataView_InvalidParameter_FieldScope        = "DataModel.DataView.InvalidParameter.FieldScope"
	DataModel_DataView_InvalidParameter_ImportMode        = "DataModel.DataView.InvalidParameter.ImportMode"
	DataModel_DataView_InvalidParameter_IncludeBuiltin    = "DataModel.DataView.InvalidParameter.IncludeBuiltin"
	DataModel_DataView_InvalidParameter_LogGroupFilters   = "DataModel.DataView.InvalidParameter.LogGroupFilters"
	DataModel_DataView_InvalidParameter_OpenStreaming     = "DataModel.DataView.InvalidParameter.OpenStreaming"
	DataModel_DataView_InvalidParameter_TechnicalName     = "DataModel.DataView.InvalidParameter.TechnicalName"
	DataModel_DataView_InvalidParameter_ViewID            = "DataModel.DataView.InvalidParameter.ViewID"
	DataModel_DataView_LengthExceeded_Comment             = "DataModel.DataView.LengthExceeded.Comment"
	DataModel_DataView_LengthExceeded_FieldComment        = "DataModel.DataView.LengthExceeded.FieldComment"
	DataModel_DataView_LengthExceeded_FieldDisplayName    = "DataModel.DataView.LengthExceeded.FieldDisplayName"
	DataModel_DataView_LengthExceeded_FieldFeatureComment = "DataModel.DataView.LengthExceeded.FieldFeatureComment"
	DataModel_DataView_LengthExceeded_FieldFeatureName    = "DataModel.DataView.LengthExceeded.FieldFeatureName"
	DataModel_DataView_LengthExceeded_FieldName           = "DataModel.DataView.LengthExceeded.FieldName"
	DataModel_DataView_LengthExceeded_ViewGroupName       = "DataModel.DataView.LengthExceeded.ViewGroupName"
	DataModel_DataView_LengthExceeded_ViewName            = "DataModel.DataView.LengthExceeded.ViewName"
	DataModel_DataView_MissingRequiredField               = "DataModel.DataView.MissingRequiredField"
	DataModel_DataView_NullParameter_AttributeFields      = "DataModel.DataView.NullParameter.AttributeFields"
	DataModel_DataView_NullParameter_Fields               = "DataModel.DataView.NullParameter.Fields"
	DataModel_DataView_NullParameter_IndexBaseType        = "DataModel.DataView.NullParameter.IndexBaseType"
	DataModel_DataView_NullParameter_ViewName             = "DataModel.DataView.NullParameter.ViewName"
	DataModel_DataView_UnsupportDataSourceType            = "DataModel.DataView.UnsupportDataSourceType"

	// 403
	DataModel_DataView_Duplicated_ViewIDInFile            = "DataModel.DataView.Duplicated.ViewIDInFile"
	DataModel_DataView_Duplicated_ViewNameInFile          = "DataModel.DataView.Duplicated.ViewNameInFile"
	DataModel_DataView_Duplicated_ViewTechnicalNameInFile = "DataModel.DataView.Duplicated.ViewTechnicalNameInFile"
	DataModel_DataView_Existed_ViewID                     = "DataModel.DataView.Existed.ViewID"
	DataModel_DataView_Existed_ViewName                   = "DataModel.DataView.Existed.ViewName"
	DataModel_DataView_Existed_TechnicalName              = "DataModel.DataView.Existed.TechnicalName"
	DataModel_DataView_FilterBinaryFieldsForbidden        = "DataModel.DataView.FilterBinaryFieldsForbidden"
	DataModel_DataView_InvalidBuiltinGroupMatch           = "DataModel.DataView.InvalidBuiltinGroupMatch"
	DataModel_DataView_InvalidFieldPermission_Filters     = "DataModel.DataView.InvalidFieldPermission.Filters"
	DataModel_DataView_RealTimeStreamingForbidden         = "DataModel.DataView.RealTimeStreamingForbidden"

	// 404
	DataModel_DataView_DataViewNotFound = "DataModel.DataView.DataViewNotFound"

	// 406
	DataModel_DataView_Unsupport_ContextType = "DataModel.DataView.Unsupport.ContextType"

	// 500
	DataModel_DataView_InternalError_BeginDbTransactionFailed              = "DataModel.DataView.InternalError.BeginDbTransactionFailed"
	DataModel_DataView_InternalError_CheckJobConfigChanged                 = "DataModel.DataView.InternalError.CheckJobConfigChanged"
	DataModel_DataView_InternalError_CheckViewIfExistFailed                = "DataModel.DataView.InternalError.CheckViewIfExistFailed"
	DataModel_DataView_InternalError_ConvertBaseFieldsToViewFieldsFailed   = "DataModel.DataView.InternalError.ConvertBaseFieldsToViewFieldsFailed"
	DataModel_DataView_InternalError_CreateDataModelJobFailed              = "DataModel.DataView.InternalError.CreateDataModelJobFailed"
	DataModel_DataView_InternalError_CreateDataViewsFailed                 = "DataModel.DataView.InternalError.CreateDataViewsFailed"
	DataModel_DataView_InternalError_DeleteDataModelJobFailed              = "DataModel.DataView.InternalError.DeleteDataModelJobFailed"
	DataModel_DataView_InternalError_DeleteDataViewsFailed                 = "DataModel.DataView.InternalError.DeleteDataViewsFailed"
	DataModel_DataView_InternalError_GetDataViewIDsByNamesFailed           = "DataModel.DataView.InternalError.GetDataViewIDsByNamesFailed"
	DataModel_DataView_InternalError_GetDataViewsByGroupIDFailed           = "DataModel.DataView.InternalError.GetDataViewsByGroupIDFailed"
	DataModel_DataView_InternalError_GetDataViewsFailed                    = "DataModel.DataView.InternalError.GetDataViewsFailed"
	DataModel_DataView_InternalError_GetDataViewsTotalFailed               = "DataModel.DataView.InternalError.GetDataViewsTotalFailed"
	DataModel_DataView_InternalError_GetDetailedDataViewMapByIDsFailed     = "DataModel.DataView.InternalError.GetDetailedDataViewMapByIDsFailed"
	DataModel_DataView_InternalError_GetIndexBaseByTypeFailed              = "DataModel.DataView.InternalError.GetIndexBaseByTypeFailed"
	DataModel_DataView_InternalError_GetJobsByDataViewIDsFailed            = "DataModel.DataView.InternalError.GetJobsByDataViewIDsFailed"
	DataModel_DataView_InternalError_GetSimpleDataViewMapByIDsFailed       = "DataModel.DataView.InternalError.GetSimpleDataViewMapByIDsFailed"
	DataModel_DataView_InternalError_GetSimpleDataViewMapByNamesFailed     = "DataModel.DataView.InternalError.GetSimpleDataViewMapByNamesFailed"
	DataModel_DataView_InternalError_InvalidReferenceView                  = "DataModel.DataView.InternalError.InvalidReferenceView"
	DataModel_DataView_InternalError_ListDataViewsFailed                   = "DataModel.DataView.InternalError.ListDataViewsFailed"
	DataModel_DataView_InternalError_MarshalViewAttrFailed                 = "DataModel.DataView.InternalError.MarshalViewAttrFailed"
	DataModel_DataView_InternalError_UnMarshalViewAttrFailed               = "DataModel.DataView.InternalError.UnMarshalViewAttrFailed"
	DataModel_DataView_InternalError_UpdateDataViewFailed                  = "DataModel.DataView.InternalError.UpdateDataViewFailed"
	DataModel_DataView_InternalError_UpdateDataViewRealTimeStreamingFailed = "DataModel.DataView.InternalError.UpdateDataViewRealTimeStreamingFailed"
	DataModel_DataView_InternalError_UpdateDataViewsGroupFailed            = "DataModel.DataView.InternalError.UpdateDataViewsGroupFailed"
)

// 数据视图分组错误码
const (
	// 400
	DataModel_DataViewGroup_Existed_GroupName            = "DataModel.DataViewGroup.Existed.GroupName"
	DataModel_DataViewGroup_InvalidParameter_Builtin     = "DataModel.DataViewGroup.InvalidParameter.Builtin"
	DataModel_DataViewGroup_InvalidParameter_DeleteViews = "DataModel.DataViewGroup.InvalidParameter.DeleteViews"
	DataModel_DataViewGroup_InvalidParameter_GroupName   = "DataModel.DataViewGroup.InvalidParameter.GroupName"
	DataModel_DataViewGroup_InvalidParameter_RequestBody = "DataModel.DataViewGroup.InvalidParameter.RequestBody"
	DataModel_DataViewGroup_LengthExceeded_GroupName     = "DataModel.DataViewGroup.LengthExceeded.GroupName"
	DataModel_DataViewGroup_NullParameter_GroupID        = "DataModel.DataViewGroup.NullParameter.GroupID"
	DataModel_DataViewGroup_NullParameter_GroupName      = "DataModel.DataViewGroup.NullParameter.GroupName"

	// 403
	DataModel_DataViewGroup_ForbiddenBuiltinGroupID   = "DataModel.DataViewGroup.ForbiddenBuiltinGroupID"
	DataModel_DataViewGroup_ForbiddenBuiltinGroupName = "DataModel.DataViewGroup.ForbiddenBuiltinGroupName"
	DataModel_DataViewGroup_GroupNotEmpty             = "DataModel.DataViewGroup.GroupNotEmpty"

	// 404
	DataModel_DataViewGroup_GroupNotFound = "DataModel.DataViewGroup.GroupNotFound"

	// 500
	DataModel_DataViewGroup_InternalError_BeginDBTransactionFailed     = "DataModel.DataViewGroup.InternalError.BeginDBTransactionFailed"
	DataModel_DataViewGroup_InternalError_CheckGroupExistByNameFailed  = "DataModel.DataViewGroup.InternalError.CheckGroupExistByNameFailed"
	DataModel_DataViewGroup_InternalError_CreateGroupFailed            = "DataModel.DataViewGroup.InternalError.CreateGroupFailed"
	DataModel_DataViewGroup_InternalError_DeleteDataViewsInGroupFailed = "DataModel.DataViewGroup.InternalError.DeleteDataViewsInGroupFailed"
	DataModel_DataViewGroup_InternalError_DeleteGroupFailed            = "DataModel.DataViewGroup.InternalError.DeleteGroupFailed"
	DataModel_DataViewGroup_InternalError_GetGroupByIDFailed           = "DataModel.DataViewGroup.InternalError.GetGroupByIDFailed"
	DataModel_DataViewGroup_InternalError_GetGroupsTotalFailed         = "DataModel.DataViewGroup.InternalError.GetGroupsTotalFailed"
	DataModel_DataViewGroup_InternalError_GetViewsByGroupIDFailed      = "DataModel.DataViewGroup.InternalError.GetViewsByGroupIDFailed"
	DataModel_DataViewGroup_InternalError_ListGroupsFailed             = "DataModel.DataViewGroup.InternalError.ListGroupsFailed"
	DataModel_DataViewGroup_InternalError_UpdateGroupFailed            = "DataModel.DataViewGroup.InternalError.UpdateGroupFailed"
)

// 数据视图行列规则错误码
const (
	// 400
	DataModel_DataViewRowColumnRule_ExistByName             = "DataModel.DataViewRowColumnRule.ExistByName"
	DataModel_DataViewRowColumnRule_LengthExceeded_RuleName = "DataModel.DataViewRowColumnRule.LengthExceeded.RuleName"
	DataModel_DataViewRowColumnRule_NullParameter_RuleID    = "DataModel.DataViewRowColumnRule.NullParameter.RuleID"
	DataModel_DataViewRowColumnRule_NullParameter_RuleName  = "DataModel.DataViewRowColumnRule.NullParameter.RuleName"
	DataModel_DataViewRowColumnRule_NullParameter_ViewID    = "DataModel.DataViewRowColumnRule.NullParameter.ViewID"
)

var (
	dataViewErrCodeList = []string{
		// ---数据视图模块---
		// 400
		DataModel_DataView_CountExceeded_Filters,
		DataModel_DataView_Duplicated_FieldDisplayName,
		DataModel_DataView_Duplicated_FieldFeatureName,
		DataModel_DataView_Duplicated_FieldName,
		DataModel_DataView_Duplicated_ViewID,
		DataModel_DataView_Duplicated_ViewName,
		DataModel_DataView_FieldTypeConflict,
		DataModel_DataView_FilterFieldTypeMisMatchOperation,
		DataModel_DataView_IndexbaseNotFound,
		DataModel_DataView_InvalidParameter_AttributeFields,
		DataModel_DataView_InvalidParameter_Builtin,
		DataModel_DataView_InvalidParameter_DataScope,
		DataModel_DataView_InvalidParameter_DataSource,
		DataModel_DataView_InvalidParameter_FieldFeatureName,
		DataModel_DataView_InvalidParameter_FieldName,
		DataModel_DataView_InvalidParameter_FieldScope,
		DataModel_DataView_InvalidParameter_ImportMode,
		DataModel_DataView_InvalidParameter_IncludeBuiltin,
		DataModel_DataView_InvalidParameter_LogGroupFilters,
		DataModel_DataView_InvalidParameter_OpenStreaming,
		DataModel_DataView_InvalidParameter_TechnicalName,
		DataModel_DataView_InvalidParameter_ViewID,
		DataModel_DataView_LengthExceeded_Comment,
		DataModel_DataView_LengthExceeded_FieldComment,
		DataModel_DataView_LengthExceeded_FieldDisplayName,
		DataModel_DataView_LengthExceeded_FieldFeatureComment,
		DataModel_DataView_LengthExceeded_FieldFeatureName,
		DataModel_DataView_LengthExceeded_FieldName,
		DataModel_DataView_LengthExceeded_ViewGroupName,
		DataModel_DataView_LengthExceeded_ViewName,
		DataModel_DataView_MissingRequiredField,
		DataModel_DataView_NullParameter_AttributeFields,
		DataModel_DataView_NullParameter_Fields,
		DataModel_DataView_NullParameter_IndexBaseType,
		DataModel_DataView_NullParameter_ViewName,
		DataModel_DataView_UnsupportDataSourceType,

		// 403
		DataModel_DataView_Duplicated_ViewIDInFile,
		DataModel_DataView_Duplicated_ViewNameInFile,
		DataModel_DataView_Duplicated_ViewTechnicalNameInFile,
		DataModel_DataView_Existed_ViewID,
		DataModel_DataView_Existed_ViewName,
		DataModel_DataView_Existed_TechnicalName,
		DataModel_DataView_FilterBinaryFieldsForbidden,
		DataModel_DataView_InvalidBuiltinGroupMatch,
		DataModel_DataView_InvalidFieldPermission_Filters,
		DataModel_DataView_RealTimeStreamingForbidden,

		// 404
		DataModel_DataView_DataViewNotFound,

		// 406
		DataModel_DataView_Unsupport_ContextType,

		//500
		DataModel_DataView_InternalError_BeginDbTransactionFailed,
		DataModel_DataView_InternalError_CheckJobConfigChanged,
		DataModel_DataView_InternalError_CheckViewIfExistFailed,
		DataModel_DataView_InternalError_ConvertBaseFieldsToViewFieldsFailed,
		DataModel_DataView_InternalError_CreateDataModelJobFailed,
		DataModel_DataView_InternalError_CreateDataViewsFailed,
		DataModel_DataView_InternalError_DeleteDataModelJobFailed,
		DataModel_DataView_InternalError_DeleteDataViewsFailed,
		DataModel_DataView_InternalError_GetDataViewIDsByNamesFailed,
		DataModel_DataView_InternalError_GetDataViewsByGroupIDFailed,
		DataModel_DataView_InternalError_GetDataViewsFailed,
		DataModel_DataView_InternalError_GetDataViewsTotalFailed,
		DataModel_DataView_InternalError_GetDetailedDataViewMapByIDsFailed,
		DataModel_DataView_InternalError_GetIndexBaseByTypeFailed,
		DataModel_DataView_InternalError_GetJobsByDataViewIDsFailed,
		DataModel_DataView_InternalError_GetSimpleDataViewMapByIDsFailed,
		DataModel_DataView_InternalError_GetSimpleDataViewMapByNamesFailed,
		DataModel_DataView_InternalError_InvalidReferenceView,
		DataModel_DataView_InternalError_ListDataViewsFailed,
		DataModel_DataView_InternalError_MarshalViewAttrFailed,
		DataModel_DataView_InternalError_UnMarshalViewAttrFailed,
		DataModel_DataView_InternalError_UpdateDataViewFailed,
		DataModel_DataView_InternalError_UpdateDataViewRealTimeStreamingFailed,
		DataModel_DataView_InternalError_UpdateDataViewsGroupFailed,

		// ---数据视图分组模块---
		// 400
		DataModel_DataViewGroup_Existed_GroupName,
		DataModel_DataViewGroup_InvalidParameter_Builtin,
		DataModel_DataViewGroup_InvalidParameter_DeleteViews,
		DataModel_DataViewGroup_InvalidParameter_GroupName,
		DataModel_DataViewGroup_InvalidParameter_RequestBody,
		DataModel_DataViewGroup_LengthExceeded_GroupName,
		DataModel_DataViewGroup_NullParameter_GroupID,
		DataModel_DataViewGroup_NullParameter_GroupName,

		// 403
		DataModel_DataViewGroup_ForbiddenBuiltinGroupID,
		DataModel_DataViewGroup_ForbiddenBuiltinGroupName,
		DataModel_DataViewGroup_GroupNotEmpty,

		// 404
		DataModel_DataViewGroup_GroupNotFound,

		// 500
		DataModel_DataViewGroup_InternalError_BeginDBTransactionFailed,
		DataModel_DataViewGroup_InternalError_CheckGroupExistByNameFailed,
		DataModel_DataViewGroup_InternalError_CreateGroupFailed,
		DataModel_DataViewGroup_InternalError_DeleteDataViewsInGroupFailed,
		DataModel_DataViewGroup_InternalError_DeleteGroupFailed,
		DataModel_DataViewGroup_InternalError_GetGroupByIDFailed,
		DataModel_DataViewGroup_InternalError_GetGroupsTotalFailed,
		DataModel_DataViewGroup_InternalError_GetViewsByGroupIDFailed,
		DataModel_DataViewGroup_InternalError_ListGroupsFailed,
		DataModel_DataViewGroup_InternalError_UpdateGroupFailed,

		// ---数据视图行列规则模块---
		// 400
		DataModel_DataViewRowColumnRule_ExistByName,
		DataModel_DataViewRowColumnRule_LengthExceeded_RuleName,
		DataModel_DataViewRowColumnRule_NullParameter_RuleID,
		DataModel_DataViewRowColumnRule_NullParameter_RuleName,
		DataModel_DataViewRowColumnRule_NullParameter_ViewID,
	}
)
