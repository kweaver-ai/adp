package action_scheduler

import (
	"context"
	"fmt"
	"strings"

	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"ontology-query/interfaces"
)

// ExecuteTool executes a tool-based action through tool-box API
// API: POST /tool-box/{box_id}/proxy/{tool_id}
func ExecuteTool(ctx context.Context, aoAccess interfaces.AgentOperatorAccess, actionType *interfaces.ActionType, params map[string]any) (any, error) {
	source := actionType.ActionSource

	// Validate tool configuration
	if source.BoxID == "" || source.ToolID == "" {
		return nil, fmt.Errorf("tool execution requires box_id and tool_id")
	}

	// Build tool execution request using ActionType.Parameters configuration
	execRequest := buildToolExecutionRequest(actionType.Parameters, params)
	execRequest.Timeout = 300 // 5 minutes timeout

	logger.Debugf("Executing tool: box_id=%s, tool_id=%s, request=%+v", source.BoxID, source.ToolID, execRequest)

	// Execute through tool-box API
	result, err := aoAccess.ExecuteTool(ctx, source.BoxID, source.ToolID, execRequest)
	if err != nil {
		logger.Errorf("Tool execution failed: %v", err)
		return nil, fmt.Errorf("tool execution failed: %w", err)
	}

	logger.Debugf("Tool execution completed successfully")
	return result, nil
}

// buildToolExecutionRequest builds ToolExecutionRequest based on ActionType.Parameters configuration
// Note: params already contains processed values from buildExecutionParams
func buildToolExecutionRequest(configParams []interfaces.Parameter, params map[string]any) interfaces.ToolExecutionRequest {
	request := interfaces.ToolExecutionRequest{
		Header: map[string]any{},
		Query:  map[string]any{},
		Body:   map[string]any{},
		Path:   map[string]any{},
	}

	// If no parameters configured, put all params in body (backward compatible)
	if len(configParams) == 0 {
		request.Body = params
		return request
	}

	// Process each configured parameter - get value from params (already processed by buildExecutionParams)
	// and assign to the appropriate location based on Source
	for _, param := range configParams {
		// Get value from params (params keys are param.Name from buildExecutionParams)
		value := params[param.Name]

		// Skip if value is nil
		if value == nil {
			continue
		}

		// Assign to appropriate location based on source
		switch strings.ToLower(param.Source) {
		case "header":
			request.Header[param.Name] = value
		case "query":
			request.Query[param.Name] = value
		case "body":
			request.Body[param.Name] = value
		case "path":
			request.Path[param.Name] = value
		default:
			// Default to body
			request.Body[param.Name] = value
		}
	}

	return request
}
