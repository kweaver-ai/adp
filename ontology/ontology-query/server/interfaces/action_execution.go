package interfaces

// Action execution status constants
const (
	ExecutionStatusPending   = "pending"
	ExecutionStatusRunning   = "running"
	ExecutionStatusCompleted = "completed"
	ExecutionStatusFailed    = "failed"
)

// Object execution status constants
const (
	ObjectStatusPending = "pending"
	ObjectStatusSuccess = "success"
	ObjectStatusFailed  = "failed"
)

// Trigger type constants
const (
	TriggerTypeManual    = "manual"
	TriggerTypeScheduled = "scheduled"
)

// Action source type constants
const (
	ActionSourceTypeTool = "tool"
	ActionSourceTypeMCP  = "mcp"
)

// ActionExecutionRequest represents the request to execute an action
type ActionExecutionRequest struct {
	KNID             string           `json:"-"`
	Branch           string           `json:"-"`
	ActionTypeID     string           `json:"-"`
	TriggerType      string           `json:"trigger_type,omitempty"` // "manual" or "scheduled", defaults to "manual"
	UniqueIdentities []map[string]any `json:"unique_identities"`
	DynamicParams    map[string]any   `json:"dynamic_params,omitempty"`
}

// ActionExecutionResponse represents the immediate response after submitting execution
type ActionExecutionResponse struct {
	ExecutionID string `json:"execution_id"`
	Status      string `json:"status"`
	Message     string `json:"message"`
	CreatedAt   int64  `json:"created_at"`
}

// ActionExecution represents a single execution request (may contain multiple objects)
type ActionExecution struct {
	ID               string                  `json:"id"` // execution_id
	KNID             string                  `json:"kn_id"`
	ActionTypeID     string                  `json:"action_type_id"`
	ActionTypeName   string                  `json:"action_type_name"`
	ActionSourceType string                  `json:"action_source_type"` // "tool" | "mcp"
	ActionSource     ActionSource            `json:"action_source"`
	ObjectTypeID     string                  `json:"object_type_id"`
	TriggerType      string                  `json:"trigger_type"` // "manual" | "scheduled"
	Status           string                  `json:"status"`       // "pending" | "running" | "completed" | "failed"
	TotalCount       int                     `json:"total_count"`
	SuccessCount     int                     `json:"success_count"`
	FailedCount      int                     `json:"failed_count"`
	Results          []ObjectExecutionResult `json:"results"`
	DynamicParams    map[string]any          `json:"dynamic_params,omitempty"`
	ExecutorID       string                  `json:"executor_id"` // user who triggered
	StartTime        int64                   `json:"start_time"`
	EndTime          int64                   `json:"end_time,omitempty"`
	DurationMs       int64                   `json:"duration_ms,omitempty"`
}

// ObjectExecutionResult represents execution result for a single object
type ObjectExecutionResult struct {
	UniqueIdentity map[string]any `json:"unique_identity"`
	Status         string         `json:"status"` // "pending" | "success" | "failed"
	Parameters     map[string]any `json:"parameters,omitempty"`
	Result         any            `json:"result,omitempty"`
	ErrorMessage   string         `json:"error_message,omitempty"`
	DurationMs     int64          `json:"duration_ms,omitempty"`
}

// ActionLogQuery represents query parameters for execution logs
type ActionLogQuery struct {
	KNID           string  `json:"-"`
	ActionTypeID   string  `json:"action_type_id,omitempty"`
	Status         string  `json:"status,omitempty"`
	TriggerType    string  `json:"trigger_type,omitempty"`
	StartTimeRange []int64 `json:"start_time_range,omitempty"` // [start, end]
	Limit          int     `json:"limit,omitempty"`
	NeedTotal      bool    `json:"need_total,omitempty"`
	SearchAfter    []any   `json:"search_after,omitempty"`
}

// ActionExecutionList represents a list of action executions with pagination
type ActionExecutionList struct {
	Entries     []ActionExecution `json:"entries"`
	TotalCount  int               `json:"total_count,omitempty"`
	SearchAfter []any             `json:"search_after,omitempty"`
}

// MCPExecutionRequest represents the request to execute an MCP action
type MCPExecutionRequest struct {
	McpID      string         `json:"mcp_id"`
	ToolName   string         `json:"tool_name"`
	Parameters map[string]any `json:"parameters"`
	Timeout    int64          `json:"timeout"` // timeout in seconds
}
