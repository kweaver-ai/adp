package errors

// 概念分组错误码
const (
	// 400
	OntologyManager_ConceptGroup_Duplicated_Name                    = "OntologyManager.ConceptGroup.Duplicated.Name"
	OntologyManager_ConceptGroup_InvalidParameter                   = "OntologyManager.ConceptGroup.InvalidParameter"
	OntologyManager_ConceptGroup_InvalidParameter_ConceptCondition  = "OntologyManager.ConceptGroup.InvalidParameter.ConceptCondition"
	OntologyManager_ConceptGroup_InvalidParameter_Direction         = "OntologyManager.ConceptGroup.InvalidParameter.Direction"
	OntologyManager_ConceptGroup_InvalidParameter_IncludeStatistics = "OntologyManager.ConceptGroup.InvalidParameter.IncludeStatistics"
	OntologyManager_ConceptGroup_InvalidParameter_IncludeTypeInfo   = "OntologyManager.ConceptGroup.InvalidParameter.IncludeTypeInfo"
	OntologyManager_ConceptGroup_InvalidParameter_PathLength        = "OntologyManager.ConceptGroup.InvalidParameter.PathLength"
	OntologyManager_ConceptGroup_ConceptGroupIDExisted              = "OntologyManager.ConceptGroup.ConceptGroupIDExisted"
	OntologyManager_ConceptGroup_ConceptGroupNameExisted            = "OntologyManager.ConceptGroup.ConceptGroupNameExisted"
	OntologyManager_ConceptGroup_ConceptGroupRelationExisted        = "OntologyManager.ConceptGroup.ConceptGroupRelationExisted"
	OntologyManager_ConceptGroup_ConceptGroupRelationNotExisted     = "OntologyManager.ConceptGroup.ConceptGroupRelationNotExisted"
	OntologyManager_ConceptGroup_LengthExceeded_Name                = "OntologyManager.ConceptGroup.LengthExceeded.Name"
	OntologyManager_ConceptGroup_NullParameter_Direction            = "OntologyManager.ConceptGroup.NullParameter.Direction"
	OntologyManager_ConceptGroup_NullParameter_Name                 = "OntologyManager.ConceptGroup.NullParameter.Name"
	OntologyManager_ConceptGroup_NullParameter_SourceObjectTypeId   = "OntologyManager.ConceptGroup.NullParameter.SourceObjectTypeId"

	// 404
	OntologyManager_ConceptGroup_ConceptGroupNotFound = "OntologyManager.ConceptGroup.ConceptGroupNotFound"
	OntologyManager_ConceptGroup_ObjectTypeNotFound   = "OntologyManager.ConceptGroup.ObjectTypeNotFound"

	// 500
	OntologyManager_ConceptGroup_InternalError                                      = "OntologyManager.ConceptGroup.InternalError"
	OntologyManager_ConceptGroup_InternalError_AddObjectTypesToConceptGroupFailed   = "OntologyManager.ConceptGroup.InternalError.AddObjectTypesToConceptGroupFailed"
	OntologyManager_ConceptGroup_InternalError_MissingTransaction                   = "OntologyManager.ConceptGroup.InternalError.MissingTransaction"
	OntologyManager_ConceptGroup_InternalError_BeginTransactionFailed               = "OntologyManager.ConceptGroup.InternalError.BeginTransactionFailed"
	OntologyManager_ConceptGroup_InternalError_BindBusinessDomainFailed             = "OntologyManager.ConceptGroup.InternalError.BindBusinessDomainFailed"
	OntologyManager_ConceptGroup_InternalError_UnbindBusinessDomainFailed           = "OntologyManager.ConceptGroup.InternalError.UnbindBusinessDomainFailed"
	OntologyManager_ConceptGroup_InternalError_CheckConceptGroupIfExistFailed       = "OntologyManager.ConceptGroup.InternalError.CheckConceptGroupIfExistFailed"
	OntologyManager_ConceptGroup_InternalError_GetConceptGroupByIDFailed            = "OntologyManager.ConceptGroup.InternalError.GetConceptGroupByIDFailed"
	OntologyManager_ConceptGroup_InternalError_UpdateConceptGroupFailed             = "OntologyManager.ConceptGroup.InternalError.UpdateConceptGroupFailed"
	OntologyManager_ConceptGroup_InternalError_CreateConceptGroupFailed             = "OntologyManager.ConceptGroup.InternalError.CreateConceptGroupFailed"
	OntologyManager_ConceptGroup_InternalError_CreateConceptGroupRelationFailed     = "OntologyManager.ConceptGroup.InternalError.CreateConceptGroupRelationFailed"
	OntologyManager_ConceptGroup_InternalError_GetActionTypesTotalFailed            = "OntologyManager.ConceptGroup.InternalError.GetActionTypesTotalFailed"
	OntologyManager_ConceptGroup_InternalError_GetConceptIDsByConceptGroupIDsFailed = "OntologyManager.ConceptGroup.InternalError.GetConceptIDsByConceptGroupIDsFailed"
	OntologyManager_ConceptGroup_InternalError_GetRelationTypesTotalFailed          = "OntologyManager.ConceptGroup.InternalError.GetRelationTypesTotalFailed"
	OntologyManager_ConceptGroup_InternalError_GetVectorFailed                      = "OntologyManager.ConceptGroup.InternalError.GetVectorFailed"
	OntologyManager_ConceptGroup_InternalError_InsertOpenSearchDataFailed           = "OntologyManager.ConceptGroup.InternalError.InsertOpenSearchDataFailed"
	OntologyManager_ConceptGroup_InternalError_CreateObjectTypesFailed              = "OntologyManager.ConceptGroup.InternalError.CreateObjectTypesFailed"
	OntologyManager_ConceptGroup_InternalError_CreateRelationTypesFailed            = "OntologyManager.ConceptGroup.InternalError.CreateRelationTypesFailed"
	OntologyManager_ConceptGroup_InternalError_CreateActionTypesFailed              = "OntologyManager.ConceptGroup.InternalError.CreateActionTypesFailed"
)

var (
	ConceptGroupErrCodeList = []string{
		// 400
		OntologyManager_ConceptGroup_Duplicated_Name,
		OntologyManager_ConceptGroup_InvalidParameter,
		OntologyManager_ConceptGroup_InvalidParameter_ConceptCondition,
		OntologyManager_ConceptGroup_InvalidParameter_Direction,
		OntologyManager_ConceptGroup_InvalidParameter_IncludeStatistics,
		OntologyManager_ConceptGroup_InvalidParameter_IncludeTypeInfo,
		OntologyManager_ConceptGroup_InvalidParameter_PathLength,
		OntologyManager_ConceptGroup_ConceptGroupIDExisted,
		OntologyManager_ConceptGroup_ConceptGroupNameExisted,
		OntologyManager_ConceptGroup_ConceptGroupRelationExisted,
		OntologyManager_ConceptGroup_ConceptGroupRelationNotExisted,
		OntologyManager_ConceptGroup_LengthExceeded_Name,
		OntologyManager_ConceptGroup_NullParameter_Direction,
		OntologyManager_ConceptGroup_NullParameter_Name,
		OntologyManager_ConceptGroup_NullParameter_SourceObjectTypeId,

		// 404
		OntologyManager_ConceptGroup_ConceptGroupNotFound,
		OntologyManager_ConceptGroup_ObjectTypeNotFound,

		// 500
		OntologyManager_ConceptGroup_InternalError,
		OntologyManager_ConceptGroup_InternalError_AddObjectTypesToConceptGroupFailed,
		OntologyManager_ConceptGroup_InternalError_MissingTransaction,
		OntologyManager_ConceptGroup_InternalError_CheckConceptGroupIfExistFailed,
		OntologyManager_ConceptGroup_InternalError_BeginTransactionFailed,
		OntologyManager_ConceptGroup_InternalError_BindBusinessDomainFailed,
		OntologyManager_ConceptGroup_InternalError_UnbindBusinessDomainFailed,
		OntologyManager_ConceptGroup_InternalError_GetConceptGroupByIDFailed,
		OntologyManager_ConceptGroup_InternalError_UpdateConceptGroupFailed,
		OntologyManager_ConceptGroup_InternalError_CreateConceptGroupFailed,
		OntologyManager_ConceptGroup_InternalError_CreateConceptGroupRelationFailed,
		OntologyManager_ConceptGroup_InternalError_GetActionTypesTotalFailed,
		OntologyManager_ConceptGroup_InternalError_GetConceptIDsByConceptGroupIDsFailed,
		OntologyManager_ConceptGroup_InternalError_GetRelationTypesTotalFailed,
		OntologyManager_ConceptGroup_InternalError_GetVectorFailed,
		OntologyManager_ConceptGroup_InternalError_InsertOpenSearchDataFailed,
		OntologyManager_ConceptGroup_InternalError_CreateObjectTypesFailed,
		OntologyManager_ConceptGroup_InternalError_CreateRelationTypesFailed,
		OntologyManager_ConceptGroup_InternalError_CreateActionTypesFailed,
	}
)
