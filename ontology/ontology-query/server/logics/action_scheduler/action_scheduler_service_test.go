package action_scheduler

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"ontology-query/interfaces"
)

func Test_buildExecutionParams(t *testing.T) {
	Convey("Test buildExecutionParams", t, func() {
		s := &actionSchedulerService{}

		Convey("should get value from object property (VALUE_FROM_PROP)", func() {
			actionType := &interfaces.ActionType{
				Parameters: []interfaces.Parameter{
					{
						Name:      "target_ip",
						ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_PROP,
						Value:     "pod_ip",
					},
				},
			}
			identity := map[string]any{
				"pod_ip":   "192.168.1.1",
				"pod_name": "test-pod",
			}

			params, err := s.buildExecutionParams(actionType, identity, nil)

			So(err, ShouldBeNil)
			So(params["target_ip"], ShouldEqual, "192.168.1.1")
		})

		Convey("should get value from constant (VALUE_FROM_CONST)", func() {
			actionType := &interfaces.ActionType{
				Parameters: []interfaces.Parameter{
					{
						Name:      "timeout",
						ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_CONST,
						Value:     60,
					},
				},
			}
			identity := map[string]any{}

			params, err := s.buildExecutionParams(actionType, identity, nil)

			So(err, ShouldBeNil)
			So(params["timeout"], ShouldEqual, 60)
		})

		Convey("should get value from dynamic params (VALUE_FROM_INPUT)", func() {
			actionType := &interfaces.ActionType{
				Parameters: []interfaces.Parameter{
					{
						Name:      "Authorization",
						ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_INPUT,
					},
				},
			}
			identity := map[string]any{}
			dynamicParams := map[string]any{
				"Authorization": "Bearer token123",
			}

			params, err := s.buildExecutionParams(actionType, identity, dynamicParams)

			So(err, ShouldBeNil)
			So(params["Authorization"], ShouldEqual, "Bearer token123")
		})

		Convey("should handle mixed parameter sources", func() {
			actionType := &interfaces.ActionType{
				Parameters: []interfaces.Parameter{
					{
						Name:      "target_ip",
						ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_PROP,
						Value:     "pod_ip",
					},
					{
						Name:      "timeout",
						ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_CONST,
						Value:     30,
					},
					{
						Name:      "token",
						ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_INPUT,
					},
				},
			}
			identity := map[string]any{
				"pod_ip": "10.0.0.1",
			}
			dynamicParams := map[string]any{
				"token": "abc123",
			}

			params, err := s.buildExecutionParams(actionType, identity, dynamicParams)

			So(err, ShouldBeNil)
			So(params["target_ip"], ShouldEqual, "10.0.0.1")
			So(params["timeout"], ShouldEqual, 30)
			So(params["token"], ShouldEqual, "abc123")
		})

		Convey("should handle missing property in identity", func() {
			actionType := &interfaces.ActionType{
				Parameters: []interfaces.Parameter{
					{
						Name:      "target_ip",
						ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_PROP,
						Value:     "pod_ip",
					},
				},
			}
			identity := map[string]any{
				"pod_name": "test-pod", // pod_ip is missing
			}

			params, err := s.buildExecutionParams(actionType, identity, nil)

			So(err, ShouldBeNil)
			_, exists := params["target_ip"]
			So(exists, ShouldBeFalse) // Parameter should not be set if property is missing
		})

		Convey("should handle missing dynamic param", func() {
			actionType := &interfaces.ActionType{
				Parameters: []interfaces.Parameter{
					{
						Name:      "token",
						ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_INPUT,
					},
				},
			}
			identity := map[string]any{}

			params, err := s.buildExecutionParams(actionType, identity, nil)

			So(err, ShouldBeNil)
			_, exists := params["token"]
			So(exists, ShouldBeFalse) // Parameter should not be set if dynamic param is missing
		})

		Convey("should handle empty parameters", func() {
			actionType := &interfaces.ActionType{
				Parameters: []interfaces.Parameter{},
			}
			identity := map[string]any{}

			params, err := s.buildExecutionParams(actionType, identity, nil)

			So(err, ShouldBeNil)
			So(len(params), ShouldEqual, 0)
		})
	})
}

func Test_ActionExecutionRequest_Validation(t *testing.T) {
	Convey("Test ActionExecutionRequest", t, func() {
		Convey("should have required fields", func() {
			req := &interfaces.ActionExecutionRequest{
				KNID:         "kn_001",
				ActionTypeID: "at_001",
				UniqueIdentities: []map[string]any{
					{"pod_ip": "192.168.1.1"},
				},
			}

			So(req.KNID, ShouldEqual, "kn_001")
			So(req.ActionTypeID, ShouldEqual, "at_001")
			So(len(req.UniqueIdentities), ShouldEqual, 1)
		})

		Convey("should support branch field", func() {
			req := &interfaces.ActionExecutionRequest{
				KNID:         "kn_001",
				Branch:       "feature/test",
				ActionTypeID: "at_001",
				UniqueIdentities: []map[string]any{
					{"pod_ip": "192.168.1.1"},
				},
			}

			So(req.KNID, ShouldEqual, "kn_001")
			So(req.Branch, ShouldEqual, "feature/test")
			So(req.ActionTypeID, ShouldEqual, "at_001")
		})

		Convey("should default branch to empty string when not set", func() {
			req := &interfaces.ActionExecutionRequest{
				KNID:         "kn_001",
				ActionTypeID: "at_001",
			}

			So(req.Branch, ShouldEqual, "")
		})

		Convey("should handle multiple objects", func() {
			req := &interfaces.ActionExecutionRequest{
				UniqueIdentities: []map[string]any{
					{"pod_ip": "192.168.1.1", "id": 1},
					{"pod_ip": "192.168.1.2", "id": 2},
					{"pod_ip": "192.168.1.3", "id": 3},
				},
			}

			So(len(req.UniqueIdentities), ShouldEqual, 3)
		})

		Convey("should handle dynamic params", func() {
			req := &interfaces.ActionExecutionRequest{
				DynamicParams: map[string]any{
					"Authorization": "Bearer xxx",
					"Timeout":       60,
				},
			}

			So(req.DynamicParams["Authorization"], ShouldEqual, "Bearer xxx")
			So(req.DynamicParams["Timeout"], ShouldEqual, 60)
		})
	})
}

func Test_ActionExecutionResponse(t *testing.T) {
	Convey("Test ActionExecutionResponse", t, func() {
		Convey("should have correct structure", func() {
			resp := &interfaces.ActionExecutionResponse{
				ExecutionID: "exec_123",
				Status:      interfaces.ExecutionStatusPending,
				Message:     "Action execution started",
				CreatedAt:   1704067200000,
			}

			So(resp.ExecutionID, ShouldEqual, "exec_123")
			So(resp.Status, ShouldEqual, "pending")
			So(resp.Message, ShouldEqual, "Action execution started")
			So(resp.CreatedAt, ShouldEqual, int64(1704067200000))
		})
	})
}

func Test_ObjectExecutionResult(t *testing.T) {
	Convey("Test ObjectExecutionResult", t, func() {
		Convey("should represent success result", func() {
			result := interfaces.ObjectExecutionResult{
				UniqueIdentity: map[string]any{"pod_ip": "192.168.1.1"},
				Status:         interfaces.ObjectStatusSuccess,
				Parameters:     map[string]any{"target_ip": "192.168.1.1"},
				Result:         map[string]any{"message": "OK"},
				DurationMs:     1200,
			}

			So(result.Status, ShouldEqual, "success")
			So(result.ErrorMessage, ShouldEqual, "")
			So(result.DurationMs, ShouldEqual, int64(1200))
		})

		Convey("should represent failed result", func() {
			result := interfaces.ObjectExecutionResult{
				UniqueIdentity: map[string]any{"pod_ip": "192.168.1.2"},
				Status:         interfaces.ObjectStatusFailed,
				Parameters:     map[string]any{"target_ip": "192.168.1.2"},
				ErrorMessage:   "Connection timeout",
				DurationMs:     5000,
			}

			So(result.Status, ShouldEqual, "failed")
			So(result.ErrorMessage, ShouldEqual, "Connection timeout")
			So(result.Result, ShouldBeNil)
		})
	})
}

func Test_ActionSource_Types(t *testing.T) {
	Convey("Test ActionSource types", t, func() {
		Convey("should handle Tool source", func() {
			source := interfaces.ActionSource{
				Type:   interfaces.ActionSourceTypeTool,
				BoxID:  "box_001",
				ToolID: "tool_001",
			}

			So(source.Type, ShouldEqual, "tool")
			So(source.BoxID, ShouldEqual, "box_001")
			So(source.ToolID, ShouldEqual, "tool_001")
		})

		Convey("should handle MCP source", func() {
			source := interfaces.ActionSource{
				Type:     interfaces.ActionSourceTypeMCP,
				McpID:    "mcp_001",
				ToolName: "restart_service",
			}

			So(source.Type, ShouldEqual, "mcp")
			So(source.McpID, ShouldEqual, "mcp_001")
			So(source.ToolName, ShouldEqual, "restart_service")
		})
	})
}
