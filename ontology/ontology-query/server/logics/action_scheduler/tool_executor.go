package action_scheduler

import (
	"context"
	"fmt"

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

	// Build tool execution request
	// Parameters are passed in the body for POST requests
	execRequest := interfaces.ToolExecutionRequest{
		Header:  map[string]any{},
		Body:    params,
		Query:   map[string]any{},
		Path:    map[string]any{},
		Timeout: 300, // 5 minutes timeout
	}

	logger.Debugf("Executing tool: box_id=%s, tool_id=%s", source.BoxID, source.ToolID)

	// Execute through tool-box API
	result, err := aoAccess.ExecuteTool(ctx, source.BoxID, source.ToolID, execRequest)
	if err != nil {
		logger.Errorf("Tool execution failed: %v", err)
		return nil, fmt.Errorf("tool execution failed: %w", err)
	}

	logger.Debugf("Tool execution completed successfully")
	return result, nil
}
