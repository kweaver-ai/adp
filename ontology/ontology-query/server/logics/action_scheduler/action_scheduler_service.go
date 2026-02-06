package action_scheduler

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/rs/xid"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"ontology-query/common"
	cond "ontology-query/common/condition"
	oerrors "ontology-query/errors"
	"ontology-query/interfaces"
	"ontology-query/logics"
	"ontology-query/logics/action_logs"
	"ontology-query/logics/object_type"
)

// Environment variable for max execution objects limit
const (
	envMaxExecutionObjects     = "ACTION_EXECUTION_MAX_OBJECTS"
	defaultMaxExecutionObjects = 10000
)

// maxExecutionObjects is the maximum number of objects allowed in a single execution
var maxExecutionObjects = defaultMaxExecutionObjects

func init() {
	if val := os.Getenv(envMaxExecutionObjects); val != "" {
		if n, err := strconv.Atoi(val); err == nil && n > 0 {
			maxExecutionObjects = n
			logger.Infof("Action execution max objects limit set to %d", maxExecutionObjects)
		}
	}
}

var (
	assOnce    sync.Once
	assService interfaces.ActionSchedulerService
)

type actionSchedulerService struct {
	appSetting  *common.AppSetting
	omAccess    interfaces.OntologyManagerAccess
	aoAccess    interfaces.AgentOperatorAccess
	logsService interfaces.ActionLogsService
	ots         interfaces.ObjectTypeService

	// Reserved hooks for future extension
	duplicateCheckHook  interfaces.DuplicateCheckHook
	permissionCheckHook interfaces.PermissionCheckHook
}

// NewActionSchedulerService creates a singleton instance of ActionSchedulerService
func NewActionSchedulerService(appSetting *common.AppSetting) interfaces.ActionSchedulerService {
	assOnce.Do(func() {
		assService = &actionSchedulerService{
			appSetting:  appSetting,
			omAccess:    logics.OMA,
			aoAccess:    logics.AOA,
			logsService: action_logs.NewActionLogsService(appSetting),
			ots:         object_type.NewObjectTypeService(appSetting),
		}
	})
	return assService
}

// ExecuteAction starts async action execution and returns execution_id immediately
func (s *actionSchedulerService) ExecuteAction(ctx context.Context, req *interfaces.ActionExecutionRequest) (*interfaces.ActionExecutionResponse, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "ExecuteAction", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	span.SetAttributes(
		attr.Key("kn_id").String(req.KNID),
		attr.Key("action_type_id").String(req.ActionTypeID),
	)

	// Get action type from ontology-manager first (needed for both scan mode and normal mode)
	actionType, actionTypeSnapshot, exists, err := s.omAccess.GetActionType(ctx, req.KNID, req.Branch, req.ActionTypeID)
	if err != nil {
		logger.Errorf("Failed to get action type: %v", err)
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_ActionExecution_GetActionTypeFailed).
			WithErrorDetails(err.Error())
	}
	if !exists {
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, oerrors.OntologyQuery_ActionExecution_ActionTypeNotFound).
			WithErrorDetails(fmt.Sprintf("Action type not found: %s", req.ActionTypeID))
	}

	var condition *cond.CondCfg
	// When _instance_identities is empty, scan all matching instances based on action type condition
	if len(req.InstanceIdentities) == 0 {
		logger.Infof("No _instance_identities provided, scanning all matching instances for action type %s", req.ActionTypeID)
		span.SetAttributes(attr.Key("scan_mode").Bool(true))

		condition = actionType.Condition
	} else {
		logger.Infof("_instance_identities provided, scanning only matching instances for action type %s", req.ActionTypeID)
		span.SetAttributes(attr.Key("scan_mode").Bool(false))

		if actionType.Condition != nil {
			instance_condition := logics.BuildInstanceIdentitiesCondition(req.InstanceIdentities)
			condition = &cond.CondCfg{
				Operation: "and",
				SubConds:  []*cond.CondCfg{instance_condition, actionType.Condition},
			}
		} else {
			condition = logics.BuildInstanceIdentitiesCondition(req.InstanceIdentities)
		}
	}

	// Query objects matching the action type condition
	objectQuery := &interfaces.ObjectQueryBaseOnObjectType{
		ActualCondition: condition, // Use action type's condition directly
		PageQuery: interfaces.PageQuery{
			Limit:     interfaces.MAX_LIMIT,
			NeedTotal: true,
		},
		KNID:         req.KNID,
		Branch:       req.Branch,
		ObjectTypeID: actionType.ObjectTypeID,
		CommonQueryParameters: interfaces.CommonQueryParameters{
			IncludeTypeInfo: false,
		},
	}

	objects, err := s.ots.GetObjectsByObjectTypeID(ctx, objectQuery)
	if err != nil {
		logger.Errorf("Failed to scan matching instances: %v", err)
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_ActionExecution_GetActionTypeFailed).
			WithErrorDetails(fmt.Sprintf("Failed to scan matching instances: %v", err))
	}

	// Extract unique identities from objects using primary keys
	for _, objData := range objects.Datas {
		instanceInfo := interfaces.ObjectSystemInfo{
			InstanceIdentity: map[string]any{},
		}
		if instance_id, ok := objData[interfaces.SYSTEM_PROPERTY_INSTANCE_ID]; ok {
			instanceInfo.InstanceID = instance_id
		}
		if identity, ok := objData[interfaces.SYSTEM_PROPERTY_INSTANCE_IDENTITY]; ok {
			if identityMap, ok := identity.(map[string]any); ok {
				instanceInfo.InstanceIdentity = identityMap
			}
		}
		if display, ok := objData[interfaces.SYSTEM_PROPERTY_DISPLAY]; ok {
			instanceInfo.Display = display
		}
		req.Instances = append(req.Instances, instanceInfo)
		req.ObjDatas = append(req.ObjDatas, objData)
	}

	// If no matching instances found after scanning, return appropriate response
	if len(req.Instances) == 0 {
		logger.Infof("No matching instances found for action type %s after scanning", req.ActionTypeID)
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ActionExecution_InvalidParameter).
			WithErrorDetails("No matching instances found for the action type condition")
	}

	logger.Infof("Scanned and found %d matching instances for action type %s", len(req.Instances), req.ActionTypeID)
	span.SetAttributes(attr.Key("scan_mode").Bool(true))
	span.SetAttributes(attr.Key("scanned_count").Int(len(req.Instances)))

	// Check execution objects limit
	if len(req.Instances) > maxExecutionObjects {
		logger.Warnf("Execution objects count %d exceeds limit %d", len(req.Instances), maxExecutionObjects)
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ActionExecution_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("Number of objects (%d) exceeds the maximum limit (%d). Please reduce the scope or adjust the ACTION_EXECUTION_MAX_OBJECTS environment variable.",
				len(req.Instances), maxExecutionObjects))
	}

	// Get executor info from context
	executor := interfaces.AccountInfo{}
	if accountInfo := ctx.Value(interfaces.ACCOUNT_INFO_KEY); accountInfo != nil {
		executor = accountInfo.(interfaces.AccountInfo)
	}

	// Reserved: Permission check hook
	if s.permissionCheckHook != nil {
		if err := s.permissionCheckHook(ctx, executor.ID, &actionType); err != nil {
			return nil, err
		}
	}

	// Reserved: Duplicate check hook
	if s.duplicateCheckHook != nil {
		proceed, err := s.duplicateCheckHook(ctx, req)
		if err != nil {
			return nil, err
		}
		if !proceed {
			return nil, rest.NewHTTPError(ctx, http.StatusConflict, oerrors.OntologyQuery_ActionExecution_DuplicateExecution).
				WithErrorDetails("Duplicate execution detected")
		}
	}

	// Generate execution ID
	executionID := xid.New().String()
	now := time.Now().UnixMilli()

	// Determine trigger type (default to manual if not specified)
	triggerType := req.TriggerType
	if triggerType == "" {
		triggerType = interfaces.TriggerTypeManual
	}

	// Create execution record with metadata only (no Results to save space)
	// Results will be stored incrementally during execution
	execution := &interfaces.ActionExecution{
		ID:                 executionID,
		KNID:               req.KNID,
		ActionTypeID:       actionType.ATID,
		ActionTypeName:     actionType.ATName,
		ActionSourceType:   actionType.ActionSource.Type,
		ActionSource:       actionType.ActionSource,
		ObjectTypeID:       actionType.ObjectTypeID,
		TriggerType:        triggerType,
		Status:             interfaces.ExecutionStatusPending,
		TotalCount:         len(req.Instances),
		SuccessCount:       0,
		FailedCount:        0,
		Results:            []interfaces.ObjectExecutionResult{}, // Empty initially to save space
		DynamicParams:      req.DynamicParams,
		ExecutorID:         executor.ID, // deprecated, kept for backward compatibility
		Executor:           executor,    // full executor info
		StartTime:          now,
		ActionTypeSnapshot: actionTypeSnapshot, // 保存执行时的行动类配置快照
	}

	// Save initial execution record (metadata only)
	if err := s.logsService.CreateExecution(ctx, execution); err != nil {
		logger.Errorf("Failed to create execution record: %v", err)
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_ActionExecution_CreateExecutionFailed).
			WithErrorDetails(err.Error())
	}

	// Start async execution in goroutine
	go s.executeAsync(execution, &actionType, req)

	// Return immediate response
	return &interfaces.ActionExecutionResponse{
		ExecutionID: executionID,
		Status:      interfaces.ExecutionStatusPending,
		Message:     "Action execution started",
		CreatedAt:   now,
	}, nil
}

// GetExecution retrieves execution status and results
func (s *actionSchedulerService) GetExecution(ctx context.Context, knID, executionID string) (*interfaces.ActionExecution, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "GetExecution", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	span.SetAttributes(
		attr.Key("kn_id").String(knID),
		attr.Key("execution_id").String(executionID),
	)

	query := &interfaces.ActionLogDetailQuery{
		KNID:         knID,
		LogID:        executionID,
		ResultsLimit: 10000, // Get all results for internal use
	}
	exec, err := s.logsService.GetExecution(ctx, query)
	if err != nil {
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, oerrors.OntologyQuery_ActionExecution_ExecutionNotFound).
			WithErrorDetails(err.Error())
	}

	return exec, nil
}

// Batch size for incremental result storage
const batchSize = 100

// executeAsync executes the action asynchronously with batch storage and cancellation support
func (s *actionSchedulerService) executeAsync(execution *interfaces.ActionExecution,
	actionType *interfaces.ActionType, req *interfaces.ActionExecutionRequest) {

	// Create a new context for async execution
	ctx := context.Background()
	// Restore account info from execution record for downstream API calls (user_id header)
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, execution.Executor)

	logger.Infof("Starting async execution: %s, total objects: %d", execution.ID, len(req.Instances))

	// Update status to running
	if err := s.logsService.UpdateExecution(ctx, execution.KNID, execution.ID, map[string]any{
		"status": interfaces.ExecutionStatusRunning,
	}); err != nil {
		logger.Warnf("Failed to update execution status to running: %v", err)
	}

	// Execute objects in batches
	successCount := 0
	failedCount := 0
	cancelledCount := 0
	allResults := []interfaces.ObjectExecutionResult{}
	cancelled := false

	for i, objData := range req.ObjDatas {
		// Check cancellation status at the start of each batch
		if i%batchSize == 0 && i > 0 {
			// Check if execution has been cancelled
			if s.isExecutionCancelled(ctx, execution.KNID, execution.ID) {
				logger.Infof("Execution %s cancelled, stopping at object %d/%d", execution.ID, i, len(req.ObjDatas))
				cancelled = true
				// Mark remaining objects as cancelled
				for j := i; j < len(req.Instances); j++ {
					allResults = append(allResults, interfaces.ObjectExecutionResult{
						ObjectSystemInfo: req.Instances[j],
						Status:           interfaces.ObjectStatusCancelled,
						ErrorMessage:     "execution cancelled",
					})
					cancelledCount++
				}
				break
			}

			// Batch update: save current progress
			s.updateExecutionProgress(ctx, execution, successCount, failedCount, allResults)
			logger.Debugf("Execution %s progress: %d/%d completed", execution.ID, i, len(req.ObjDatas))
		}

		startTime := time.Now().UnixMilli()

		// Build parameters for this object
		params, err := s.buildExecutionParams(actionType, objData, req.DynamicParams)
		if err != nil {
			endTime := time.Now().UnixMilli()
			allResults = append(allResults, interfaces.ObjectExecutionResult{
				ObjectSystemInfo: req.Instances[i],
				Status:           interfaces.ObjectStatusFailed,
				ErrorMessage:     fmt.Sprintf("Failed to build parameters: %v", err),
				StartTime:        startTime,
				EndTime:          endTime,
				DurationMs:       endTime - startTime,
			})
			failedCount++
			continue
		}

		// Execute based on action source type
		var result any
		var execErr error

		switch actionType.ActionSource.Type {
		case interfaces.ActionSourceTypeTool:
			result, execErr = ExecuteTool(ctx, s.aoAccess, actionType, params)
		case interfaces.ActionSourceTypeMCP:
			result, execErr = ExecuteMCP(ctx, s.aoAccess, actionType, params)
		default:
			execErr = fmt.Errorf("unsupported action source type: %s", actionType.ActionSource.Type)
		}

		endTime := time.Now().UnixMilli()
		if execErr != nil {
			allResults = append(allResults, interfaces.ObjectExecutionResult{
				ObjectSystemInfo: req.Instances[i],
				Status:           interfaces.ObjectStatusFailed,
				Parameters:       params,
				ErrorMessage:     execErr.Error(),
				StartTime:        startTime,
				EndTime:          endTime,
				DurationMs:       endTime - startTime,
			})
			failedCount++
		} else {
			allResults = append(allResults, interfaces.ObjectExecutionResult{
				ObjectSystemInfo: req.Instances[i],
				Status:           interfaces.ObjectStatusSuccess,
				Parameters:       params,
				Result:           result,
				StartTime:        startTime,
				EndTime:          endTime,
				DurationMs:       endTime - startTime,
			})
			successCount++
		}
	}

	// Determine final status
	var finalStatus string
	if cancelled {
		finalStatus = interfaces.ExecutionStatusCancelled
	} else if failedCount == len(req.Instances) {
		finalStatus = interfaces.ExecutionStatusFailed
	} else {
		finalStatus = interfaces.ExecutionStatusCompleted
	}

	endTime := time.Now().UnixMilli()

	// Update final execution record
	updates := map[string]any{
		"status":        finalStatus,
		"success_count": successCount,
		"failed_count":  failedCount,
		"results":       allResults,
		"end_time":      endTime,
		"duration_ms":   endTime - execution.StartTime,
	}

	if err := s.logsService.UpdateExecution(ctx, execution.KNID, execution.ID, updates); err != nil {
		logger.Errorf("Failed to update execution record: %v", err)
	}

	logger.Infof("Completed async execution: %s, success: %d, failed: %d, cancelled: %d",
		execution.ID, successCount, failedCount, cancelledCount)
}

// isExecutionCancelled checks if the execution has been cancelled
func (s *actionSchedulerService) isExecutionCancelled(ctx context.Context, knID, execID string) bool {
	query := &interfaces.ActionLogDetailQuery{
		KNID:         knID,
		LogID:        execID,
		ResultsLimit: 0, // Only need metadata, not results
	}
	exec, err := s.logsService.GetExecution(ctx, query)
	if err != nil {
		logger.Warnf("Failed to check execution status: %v", err)
		return false
	}
	return exec.Status == interfaces.ExecutionStatusCancelled
}

// updateExecutionProgress updates the execution progress (batch update)
func (s *actionSchedulerService) updateExecutionProgress(ctx context.Context, execution *interfaces.ActionExecution, successCount, failedCount int, results []interfaces.ObjectExecutionResult) {
	updates := map[string]any{
		"success_count": successCount,
		"failed_count":  failedCount,
		"results":       results,
	}
	if err := s.logsService.UpdateExecution(ctx, execution.KNID, execution.ID, updates); err != nil {
		logger.Warnf("Failed to update execution progress: %v", err)
	}
}

// buildExecutionParams builds the execution parameters from action type parameters and object data
func (s *actionSchedulerService) buildExecutionParams(actionType *interfaces.ActionType,
	instance map[string]any, dynamicParams map[string]any) (map[string]any, error) {

	params := make(map[string]any)

	for _, param := range actionType.Parameters {
		switch param.ValueFrom {
		case interfaces.LOGIC_PARAMS_VALUE_FROM_PROP:
			// Get value from object property
			if propName, ok := param.Value.(string); ok {
				if val, exists := instance[propName]; exists {
					params[param.Name] = val
				}
			}
		case interfaces.LOGIC_PARAMS_VALUE_FROM_CONST:
			// Use constant value
			params[param.Name] = param.Value
		case interfaces.LOGIC_PARAMS_VALUE_FROM_INPUT:
			// Get value from dynamic params
			if dynamicParams != nil {
				if val, exists := dynamicParams[param.Name]; exists {
					params[param.Name] = val
				}
			}
		}
	}

	return params, nil
}
