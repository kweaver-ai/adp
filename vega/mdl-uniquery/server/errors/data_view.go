package errors

// 数据视图
const (
	// 400
	Uniquery_DataView_BinaryFieldSortNotSupported         = "Uniquery.DataView.BinaryFieldSortNotSupported"
	Uniquery_DataView_CountExceeded_Filters               = "Uniquery.DataView.CountExceeded.Filters"
	Uniquery_DataView_FieldTypeConflict                   = "Uniquery.DataView.FieldTypeConflict"
	Uniquery_DataView_InvalidFilterField_FieldNotInView   = "Uniquery.DataView.InvalidFilterField.FieldNotInView"
	Uniquery_DataView_InvalidParameter_AllowNonExistField = "Uniquery.DataView.InvalidParameter.AllowNonExistField"
	Uniquery_DataView_InvalidParameter_DataScope          = "Uniquery.DataView.InvalidParameter.DataScope"
	Uniquery_DataView_InvalidParameter_DataSource         = "Uniquery.DataView.InvalidParameter.DataSource"
	Uniquery_DataView_InvalidParameter_Direction          = "Uniquery.DataView.InvalidParameter.Direction"
	Uniquery_DataView_InvalidParameter_FieldScope         = "Uniquery.DataView.InvalidParameter.FieldScope"
	Uniquery_DataView_InvalidParameter_Filters            = "Uniquery.DataView.InvalidParameter.Filters"
	Uniquery_DataView_InvalidParameter_Format             = "Uniquery.DataView.InvalidParameter.Format"
	Uniquery_DataView_InvalidParameter_IncludeView        = "Uniquery.DataView.InvalidParameter.IncludeView"
	Uniquery_DataView_InvalidParameter_PitKeepAlive       = "Uniquery.DataView.InvalidParameter.PitKeepAlive"
	Uniquery_DataView_InvalidParameter_QueryType          = "Uniquery.DataView.InvalidParameter.QueryType"
	Uniquery_DataView_InvalidParameter_Scroll             = "Uniquery.DataView.InvalidParameter.Scroll"
	Uniquery_DataView_InvalidParameter_Sort               = "Uniquery.DataView.InvalidParameter.Sort"
	Uniquery_DataView_InvalidParameter_ViewIDs            = "Uniquery.DataView.InvalidParameter.ViewIDs"
	Uniquery_DataView_MissingRequiredField                = "Uniquery.DataView.MissingRequiredField"
	Uniquery_DataView_NullParameter_Fields                = "Uniquery.DataView.NullParameter.Fields"
	Uniquery_DataView_NullParameter_IndexBaseNames        = "Uniquery.DataView.NullParameter.IndexBaseNames"
	Uniquery_DataView_NullParameter_Scroll                = "Uniquery.DataView.NullParameter.Scroll"
	Uniquery_DataView_OffsetNotAllowedWithScroll          = "Uniquery.DataView.OffsetNotAllowedWithScroll"
	Uniquery_DataView_UnsupportDataSourceType             = "Uniquery.DataView.UnsupportDataSourceType"

	// 403
	Uniquery_DataView_InvalidFieldPermission_Sort = "Uniquery.DataView.InvalidFieldPermission.Sort"

	// 404
	Uniquery_DataView_DataViewNotFound                 = "Uniquery.DataView.DataViewNotFound"
	Uniquery_DataView_PointInTimeSearchContextNotFound = "Uniquery.DataView.PointInTimeSearchContextNotFound"

	// 500
	Uniquery_DataView_InternalError_ConvertSearchAfterToDSLFailed  = "Uniquery.DataView.InternalError.ConvertSearchAfterToDSLFailed"
	Uniquery_DataView_InternalError_ConvertToDSLFailed             = "Uniquery.DataView.InternalError.ConvertToDSLFailed"
	Uniquery_DataView_InternalError_ConvertToViewUniResponseFailed = "Uniquery.DataView.InternalError.ConvertToViewUniResponseFailed"
	Uniquery_DataView_InternalError_CreatePointInTimeFailed        = "Uniquery.DataView.InternalError.CreatePointInTimeFailed"
	Uniquery_DataView_InternalError_DeletePointInTimeFailed        = "Uniquery.DataView.InternalError.DeletePointInTimeFailed"
	Uniquery_DataView_InternalError_FetchDataFromVegaFailed        = "Uniquery.DataView.InternalError.FetchDataFromVegaFailed"
	Uniquery_DataView_InternalError_GetDataViewByIDFailed          = "Uniquery.DataView.InternalError.GetDataViewByIDFailed"
	Uniquery_DataView_InternalError_GetDocumentsFailed             = "Uniquery.DataView.InternalError.GetDocumentsFailed"
	Uniquery_DataView_InternalError_GetIndexBaseByTypeFailed       = "Uniquery.DataView.InternalError.GetIndexBaseByTypeFailed"
	Uniquery_DataView_InternalError_GetIndicesFailed               = "Uniquery.DataView.InternalError.GetIndicesFailed"
	Uniquery_DataView_InternalError_GetPitIdFailed                 = "Uniquery.DataView.InternalError.GetPitIdFailed"
	Uniquery_DataView_InternalError_GetSearchAfterValueFailed      = "Uniquery.DataView.InternalError.GetSearchAfterValueFailed"
	Uniquery_DataView_InternalError_GetScrollIdFailed              = "Uniquery.DataView.InternalError.GetScrollIdFailed"
	Uniquery_DataView_InternalError_GetTotalFailed                 = "Uniquery.DataView.InternalError.GetTotalFailed"
	Uniquery_DataView_InternalError_InvalidReferenceView           = "Uniquery.DataView.InternalError.InvalidReferenceView"
	Uniquery_DataView_InternalError_LoadIndexShardsFailed          = "Uniquery.DataView.InternalError.LoadIndexShardsFailed"
	Uniquery_DataView_InternalError_MarshalFailed                  = "Uniquery.DataView.InternalError.MarshalFailed"
	Uniquery_DataView_InternalError_ProcessDocFailed               = "Uniquery.DataView.InternalError.ProcessDocFailed"
	Uniquery_DataView_InternalError_SetDocIdFailed                 = "Uniquery.DataView.InternalError.SetDocIdFailed"
	Uniquery_DataView_InternalError_SetDocIndexFailed              = "Uniquery.DataView.InternalError.SetDocIndexFailed"
	Uniquery_DataView_InternalError_SubmitTaskFailed               = "Uniquery.DataView.InternalError.SubmitTaskFailed"
	Uniquery_DataView_InternalError_UnmarshalFailed                = "Uniquery.DataView.InternalError.UnmarshalFailed"
)

var (
	dataViewErrCodeList = []string{
		// 400
		Uniquery_DataView_BinaryFieldSortNotSupported,
		Uniquery_DataView_CountExceeded_Filters,
		Uniquery_DataView_FieldTypeConflict,
		Uniquery_DataView_InvalidFilterField_FieldNotInView,
		Uniquery_DataView_InvalidParameter_AllowNonExistField,
		Uniquery_DataView_InvalidParameter_DataScope,
		Uniquery_DataView_InvalidParameter_DataSource,
		Uniquery_DataView_InvalidParameter_Direction,
		Uniquery_DataView_InvalidParameter_FieldScope,
		Uniquery_DataView_InvalidParameter_Filters,
		Uniquery_DataView_InvalidParameter_Format,
		Uniquery_DataView_InvalidParameter_IncludeView,
		Uniquery_DataView_InvalidParameter_PitKeepAlive,
		Uniquery_DataView_InvalidParameter_QueryType,
		Uniquery_DataView_InvalidParameter_Scroll,
		Uniquery_DataView_InvalidParameter_Sort,
		Uniquery_DataView_InvalidParameter_ViewIDs,
		Uniquery_DataView_MissingRequiredField,
		Uniquery_DataView_NullParameter_Fields,
		Uniquery_DataView_NullParameter_IndexBaseNames,
		Uniquery_DataView_NullParameter_Scroll,
		Uniquery_DataView_OffsetNotAllowedWithScroll,
		Uniquery_DataView_UnsupportDataSourceType,

		// 403
		Uniquery_DataView_InvalidFieldPermission_Sort,

		// 404
		Uniquery_DataView_DataViewNotFound,
		Uniquery_DataView_PointInTimeSearchContextNotFound,
		// 500
		Uniquery_DataView_InternalError_ConvertSearchAfterToDSLFailed,
		Uniquery_DataView_InternalError_ConvertToDSLFailed,
		Uniquery_DataView_InternalError_ConvertToViewUniResponseFailed,
		Uniquery_DataView_InternalError_CreatePointInTimeFailed,
		Uniquery_DataView_InternalError_DeletePointInTimeFailed,
		Uniquery_DataView_InternalError_FetchDataFromVegaFailed,
		Uniquery_DataView_InternalError_GetDataViewByIDFailed,
		Uniquery_DataView_InternalError_GetDocumentsFailed,
		Uniquery_DataView_InternalError_GetIndexBaseByTypeFailed,
		Uniquery_DataView_InternalError_GetIndicesFailed,
		Uniquery_DataView_InternalError_GetPitIdFailed,
		Uniquery_DataView_InternalError_GetSearchAfterValueFailed,
		Uniquery_DataView_InternalError_GetScrollIdFailed,
		Uniquery_DataView_InternalError_GetTotalFailed,
		Uniquery_DataView_InternalError_InvalidReferenceView,
		Uniquery_DataView_InternalError_LoadIndexShardsFailed,
		Uniquery_DataView_InternalError_MarshalFailed,
		Uniquery_DataView_InternalError_ProcessDocFailed,
		Uniquery_DataView_InternalError_SetDocIdFailed,
		Uniquery_DataView_InternalError_SetDocIndexFailed,
		Uniquery_DataView_InternalError_SubmitTaskFailed,
		Uniquery_DataView_InternalError_UnmarshalFailed,
	}
)
