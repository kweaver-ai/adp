package action_scheduler

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/rs/xid"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"ontology-query/common"
	oerrors "ontology-query/errors"
	"ontology-query/interfaces"
	"ontology-query/logics"
	"ontology-query/logics/action_logs"
)

var (
	assOnce    sync.Once
	assService interfaces.ActionSchedulerService
)

type actionSchedulerService struct {
	appSetting  *common.AppSetting
	omAccess    interfaces.OntologyManagerAccess
	aoAccess    interfaces.AgentOperatorAccess
	logsService interfaces.ActionLogsService

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

	// Validate request
	if len(req.UniqueIdentities) == 0 {
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ActionExecution_InvalidParameter).
			WithErrorDetails("unique_identities is required and cannot be empty")
	}

	// Get action type from ontology-manager
	actionType, exists, err := s.omAccess.GetActionType(ctx, req.KNID, req.Branch, req.ActionTypeID)
	if err != nil {
		logger.Errorf("Failed to get action type: %v", err)
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_ActionExecution_GetActionTypeFailed).
			WithErrorDetails(err.Error())
	}
	if !exists {
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, oerrors.OntologyQuery_ActionExecution_ActionTypeNotFound).
			WithErrorDetails(fmt.Sprintf("Action type not found: %s", req.ActionTypeID))
	}

	// Get executor ID from context
	executorID := ""
	if accountInfo := ctx.Value(interfaces.ACCOUNT_INFO_KEY); accountInfo != nil {
		executorID = accountInfo.(interfaces.AccountInfo).ID
	}

	// Reserved: Permission check hook
	if s.permissionCheckHook != nil {
		if err := s.permissionCheckHook(ctx, executorID, &actionType); err != nil {
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

	// Build initial object results
	objectResults := make([]interfaces.ObjectExecutionResult, len(req.UniqueIdentities))
	for i, identity := range req.UniqueIdentities {
		objectResults[i] = interfaces.ObjectExecutionResult{
			UniqueIdentity: identity,
			Status:         interfaces.ObjectStatusPending,
		}
	}

	// Determine trigger type (default to manual if not specified)
	triggerType := req.TriggerType
	if triggerType == "" {
		triggerType = interfaces.TriggerTypeManual
	}

	// Create execution record
	execution := &interfaces.ActionExecution{
		ID:               executionID,
		KNID:             req.KNID,
		ActionTypeID:     actionType.ATID,
		ActionTypeName:   actionType.ATName,
		ActionSourceType: actionType.ActionSource.Type,
		ActionSource:     actionType.ActionSource,
		ObjectTypeID:     actionType.ObjectTypeID,
		TriggerType:      triggerType,
		Status:           interfaces.ExecutionStatusPending,
		TotalCount:       len(req.UniqueIdentities),
		SuccessCount:     0,
		FailedCount:      0,
		Results:          objectResults,
		DynamicParams:    req.DynamicParams,
		ExecutorID:       executorID,
		StartTime:        now,
	}

	// Save initial execution record
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

	exec, err := s.logsService.GetExecution(ctx, knID, executionID)
	if err != nil {
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, oerrors.OntologyQuery_ActionExecution_ExecutionNotFound).
			WithErrorDetails(err.Error())
	}

	return exec, nil
}

// executeAsync executes the action asynchronously
func (s *actionSchedulerService) executeAsync(execution *interfaces.ActionExecution, actionType *interfaces.ActionType, req *interfaces.ActionExecutionRequest) {
	// Create a new context for async execution
	ctx := context.Background()

	logger.Infof("Starting async execution: %s", execution.ID)

	// Update status to running
	if err := s.logsService.UpdateExecution(ctx, execution.KNID, execution.ID, map[string]any{
		"status": interfaces.ExecutionStatusRunning,
	}); err != nil {
		// Log error but continue execution - the record exists in pending status
		// and will be updated again at completion
		logger.Warnf("Failed to update execution status to running: %v", err)
	}

	// Execute each object
	successCount := 0
	failedCount := 0
	results := make([]interfaces.ObjectExecutionResult, len(req.UniqueIdentities))

	for i, identity := range req.UniqueIdentities {
		objectStart := time.Now()

		// Build parameters for this object
		params, err := s.buildExecutionParams(actionType, identity, req.DynamicParams)
		if err != nil {
			results[i] = interfaces.ObjectExecutionResult{
				UniqueIdentity: identity,
				Status:         interfaces.ObjectStatusFailed,
				ErrorMessage:   fmt.Sprintf("Failed to build parameters: %v", err),
				DurationMs:     time.Since(objectStart).Milliseconds(),
			}
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

		if execErr != nil {
			results[i] = interfaces.ObjectExecutionResult{
				UniqueIdentity: identity,
				Status:         interfaces.ObjectStatusFailed,
				Parameters:     params,
				ErrorMessage:   execErr.Error(),
				DurationMs:     time.Since(objectStart).Milliseconds(),
			}
			failedCount++
		} else {
			results[i] = interfaces.ObjectExecutionResult{
				UniqueIdentity: identity,
				Status:         interfaces.ObjectStatusSuccess,
				Parameters:     params,
				Result:         result,
				DurationMs:     time.Since(objectStart).Milliseconds(),
			}
			successCount++
		}
	}

	// Determine final status
	finalStatus := interfaces.ExecutionStatusCompleted
	if failedCount == len(req.UniqueIdentities) {
		finalStatus = interfaces.ExecutionStatusFailed
	}

	endTime := time.Now().UnixMilli()

	// Update final execution record
	updates := map[string]any{
		"status":        finalStatus,
		"success_count": successCount,
		"failed_count":  failedCount,
		"results":       results,
		"end_time":      endTime,
		"duration_ms":   endTime - execution.StartTime,
	}

	if err := s.logsService.UpdateExecution(ctx, execution.KNID, execution.ID, updates); err != nil {
		logger.Errorf("Failed to update execution record: %v", err)
	}

	logger.Infof("Completed async execution: %s, success: %d, failed: %d", execution.ID, successCount, failedCount)
}

// buildExecutionParams builds the execution parameters from action type parameters and object data
func (s *actionSchedulerService) buildExecutionParams(actionType *interfaces.ActionType, identity map[string]any, dynamicParams map[string]any) (map[string]any, error) {
	params := make(map[string]any)

	for _, param := range actionType.Parameters {
		switch param.ValueFrom {
		case interfaces.LOGIC_PARAMS_VALUE_FROM_PROP:
			// Get value from object property
			if propName, ok := param.Value.(string); ok {
				if val, exists := identity[propName]; exists {
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
