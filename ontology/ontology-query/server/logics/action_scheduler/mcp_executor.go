package action_scheduler

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"ontology-query/interfaces"
)

// ExecuteMCP executes an MCP-based action through agent-operator-integration
// API: POST /mcp/execute/tool/{mcp_tool_id}
func ExecuteMCP(ctx context.Context, aoAccess interfaces.AgentOperatorAccess, actionType *interfaces.ActionType, params map[string]any) (any, error) {
	source := actionType.ActionSource

	// Validate MCP configuration - need mcp_id as mcp_tool_id for the API
	if source.McpID == "" {
		return nil, fmt.Errorf("MCP execution requires mcp_id")
	}

	toolName := source.ToolName
	if toolName == "" {
		toolName = source.ToolID
	}

	// Build MCP execution request
	mcpRequest := interfaces.MCPExecutionRequest{
		McpID:      source.McpID,
		ToolName:   toolName,
		Parameters: params,
		Timeout:    60, // Default 60 seconds timeout
	}

	// mcp_tool_id in the API path is the McpID from ActionSource
	mcpToolID := source.McpID

	logger.Debugf("Executing MCP: mcp_tool_id=%s, tool_name=%s", mcpToolID, toolName)

	// Execute through agent-operator-integration MCP endpoint
	result, err := aoAccess.ExecuteMCP(ctx, mcpToolID, toolName, mcpRequest)
	if err != nil {
		logger.Errorf("MCP execution failed: %v", err)
		return nil, fmt.Errorf("MCP execution failed: %w", err)
	}

	logger.Debugf("MCP execution completed successfully")
	return result, nil
}
