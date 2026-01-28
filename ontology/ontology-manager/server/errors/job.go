package errors

const (
	// 400
	OntologyManager_Job_InvalidParameter                  = "OntologyManager.Job.InvalidParameter"
	OntologyManager_Job_InvalidParameter_JobType          = "OntologyManager.Job.InvalidParameter.JobType"
	OntologyManager_Job_InvalidParameter_JobState         = "OntologyManager.Job.InvalidParameter.JobState"
	OntologyManager_Job_InvalidParameter_JobConceptConfig = "OntologyManager.Job.InvalidParameter.JobConceptConfig"
	OntologyManager_Job_InvalidParameter_TaskState        = "OntologyManager.Job.InvalidParameter.TaskState"
	OntologyManager_Job_InvalidParameter_ConceptType      = "OntologyManager.Job.InvalidParameter.ConceptType"
	OntologyManager_Job_NullParameter_Name                = "OntologyManager.Job.NullParameter.Name"
	OntologyManager_Job_LengthExceeded_Name               = "OntologyManager.Job.LengthExceeded.Name"
	OntologyManager_Job_NoneConceptType                   = "OntologyManager.Job.NoneConceptType"
	OntologyManager_Job_InvalidObjectType                 = "OntologyManager.Job.InvalidObjectType"

	// 403
	OntologyManager_Job_CreateConflict = "OntologyManager.Job.CreateConflict"
	OntologyManager_Job_JobRunning     = "OntologyManager.Job.JobRunning"

	// 404
	OntologyManager_Job_JobNotFound = "OntologyManager.Job.JobNotFound"

	// 500
	OntologyManager_Job_InternalError                         = "OntologyManager.Job.InternalError"
	OntologyManager_Job_InternalError_BeginTransactionFailed  = "OntologyManager.Job.InternalError.BeginTransactionFailed"
	OntologyManager_Job_InternalError_CommitTransactionFailed = "OntologyManager.Job.InternalError.CommitTransactionFailed"
	OntologyManager_Job_InternalError_MissingTransaction      = "OntologyManager.Job.InternalError.MissingTransaction"
)

var (
	JobErrCodeList = []string{
		OntologyManager_Job_InvalidParameter,
		OntologyManager_Job_InvalidParameter_JobType,
		OntologyManager_Job_InvalidParameter_JobState,
		OntologyManager_Job_InvalidParameter_JobConceptConfig,
		OntologyManager_Job_InvalidParameter_TaskState,
		OntologyManager_Job_InvalidParameter_ConceptType,
		OntologyManager_Job_NullParameter_Name,
		OntologyManager_Job_LengthExceeded_Name,
		OntologyManager_Job_NoneConceptType,
		OntologyManager_Job_InvalidObjectType,

		OntologyManager_Job_CreateConflict,
		OntologyManager_Job_JobRunning,

		OntologyManager_Job_JobNotFound,

		OntologyManager_Job_InternalError,
		OntologyManager_Job_InternalError_BeginTransactionFailed,
		OntologyManager_Job_InternalError_CommitTransactionFailed,
		OntologyManager_Job_InternalError_MissingTransaction,
	}
)
