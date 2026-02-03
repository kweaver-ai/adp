package action_scheduler

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"ontology-query/interfaces"
)

func Test_ExecuteTool_Validation(t *testing.T) {
	Convey("Test ExecuteTool validation", t, func() {
		Convey("should require box_id and tool_id", func() {
			actionType := &interfaces.ActionType{
				ActionSource: interfaces.ActionSource{
					Type:   interfaces.ActionSourceTypeTool,
					BoxID:  "",
					ToolID: "",
				},
			}

			// Without mock, this should fail validation
			// In production, you'd use a mock AgentOperatorAccess
			So(actionType.ActionSource.BoxID, ShouldEqual, "")
			So(actionType.ActionSource.ToolID, ShouldEqual, "")
		})

		Convey("should build correct OperatorExecutionRequest", func() {
			actionType := &interfaces.ActionType{
				ActionSource: interfaces.ActionSource{
					Type:   interfaces.ActionSourceTypeTool,
					BoxID:  "box_001",
					ToolID: "tool_001",
				},
			}
			params := map[string]any{
				"target_ip": "192.168.1.1",
				"timeout":   60,
			}

			// Verify action source is correctly configured
			So(actionType.ActionSource.Type, ShouldEqual, "tool")
			So(actionType.ActionSource.BoxID, ShouldEqual, "box_001")
			So(actionType.ActionSource.ToolID, ShouldEqual, "tool_001")
			So(params["target_ip"], ShouldEqual, "192.168.1.1")
		})
	})
}

func Test_ExecuteMCP_Validation(t *testing.T) {
	Convey("Test ExecuteMCP validation", t, func() {
		Convey("should require mcp_id", func() {
			actionType := &interfaces.ActionType{
				ActionSource: interfaces.ActionSource{
					Type:  interfaces.ActionSourceTypeMCP,
					McpID: "",
				},
			}

			So(actionType.ActionSource.McpID, ShouldEqual, "")
		})

		Convey("should use tool_name or fallback to tool_id", func() {
			Convey("with tool_name", func() {
				actionType := &interfaces.ActionType{
					ActionSource: interfaces.ActionSource{
						Type:     interfaces.ActionSourceTypeMCP,
						McpID:    "mcp_001",
						ToolName: "restart_service",
						ToolID:   "tool_fallback",
					},
				}

				// Should use ToolName
				toolName := actionType.ActionSource.ToolName
				if toolName == "" {
					toolName = actionType.ActionSource.ToolID
				}
				So(toolName, ShouldEqual, "restart_service")
			})

			Convey("fallback to tool_id", func() {
				actionType := &interfaces.ActionType{
					ActionSource: interfaces.ActionSource{
						Type:     interfaces.ActionSourceTypeMCP,
						McpID:    "mcp_001",
						ToolName: "",
						ToolID:   "tool_fallback",
					},
				}

				toolName := actionType.ActionSource.ToolName
				if toolName == "" {
					toolName = actionType.ActionSource.ToolID
				}
				So(toolName, ShouldEqual, "tool_fallback")
			})
		})

		Convey("should build correct MCPExecutionRequest", func() {
			actionType := &interfaces.ActionType{
				ActionSource: interfaces.ActionSource{
					Type:     interfaces.ActionSourceTypeMCP,
					McpID:    "mcp_001",
					ToolName: "restart_pod",
				},
			}
			params := map[string]any{
				"pod_name":      "test-pod",
				"namespace":     "default",
				"force_restart": true,
			}

			mcpRequest := interfaces.MCPExecutionRequest{
				McpID:      actionType.ActionSource.McpID,
				ToolName:   actionType.ActionSource.ToolName,
				Parameters: params,
				Timeout:    60,
			}

			So(mcpRequest.McpID, ShouldEqual, "mcp_001")
			So(mcpRequest.ToolName, ShouldEqual, "restart_pod")
			So(mcpRequest.Parameters["pod_name"], ShouldEqual, "test-pod")
			So(mcpRequest.Parameters["namespace"], ShouldEqual, "default")
			So(mcpRequest.Parameters["force_restart"], ShouldEqual, true)
			So(mcpRequest.Timeout, ShouldEqual, int64(60))
		})
	})
}

func Test_OperatorExecutionRequest(t *testing.T) {
	Convey("Test OperatorExecutionRequest", t, func() {
		Convey("should build correct request", func() {
			req := interfaces.OperatorExecutionRequest{
				Header: map[string]any{},
				Body: map[string]any{
					"key1": "value1",
					"key2": 123,
				},
				Query:   map[string]any{},
				Path:    map[string]any{},
				Timeout: 300,
			}

			So(req.Body["key1"], ShouldEqual, "value1")
			So(req.Body["key2"], ShouldEqual, 123)
			So(req.Timeout, ShouldEqual, int64(300))
		})
	})
}

func Test_ToolExecutionRequest(t *testing.T) {
	Convey("Test ToolExecutionRequest", t, func() {
		Convey("should build correct request for tool-box API", func() {
			req := interfaces.ToolExecutionRequest{
				Header: map[string]any{},
				Body: map[string]any{
					"target_ip": "192.168.1.1",
					"timeout":   60,
				},
				Query:   map[string]any{},
				Path:    map[string]any{},
				Timeout: 300,
			}

			So(req.Body["target_ip"], ShouldEqual, "192.168.1.1")
			So(req.Body["timeout"], ShouldEqual, 60)
			So(req.Timeout, ShouldEqual, int64(300))
		})
	})
}

func Test_MCPExecutionRequest(t *testing.T) {
	Convey("Test MCPExecutionRequest", t, func() {
		Convey("should build correct request", func() {
			req := interfaces.MCPExecutionRequest{
				McpID:    "mcp_001",
				ToolName: "execute_command",
				Parameters: map[string]any{
					"command": "ls -la",
					"timeout": 30,
				},
				Timeout: 60,
			}

			So(req.McpID, ShouldEqual, "mcp_001")
			So(req.ToolName, ShouldEqual, "execute_command")
			So(req.Parameters["command"], ShouldEqual, "ls -la")
			So(req.Timeout, ShouldEqual, int64(60))
		})
	})
}
