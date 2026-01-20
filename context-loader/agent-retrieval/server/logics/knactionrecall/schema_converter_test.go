package knactionrecall

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

type mockLogger struct{}

func (m *mockLogger) WithContext(ctx context.Context) interfaces.Logger { return m }
func (m *mockLogger) Debug(v ...interface{})                            {}
func (m *mockLogger) Info(v ...interface{})                             {}
func (m *mockLogger) Warn(v ...interface{})                             {}
func (m *mockLogger) Error(v ...interface{})                            {}
func (m *mockLogger) Debugf(format string, v ...interface{})            {}
func (m *mockLogger) Infof(format string, v ...interface{})             {}
func (m *mockLogger) Warnf(format string, v ...interface{})             {}
func (m *mockLogger) Errorf(format string, v ...interface{})            {}

func TestConvertMCPSchemaToFunctionCall(t *testing.T) {
	service := &knActionRecallServiceImpl{
		logger: &mockLogger{},
	}

	ctx := context.Background()

	// Case 1: Simple Schema
	inputJSON := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"}
		}
	}`
	var inputMap map[string]interface{}
	if err := json.Unmarshal([]byte(inputJSON), &inputMap); err != nil {
		t.Fatalf("Failed to unmarshal test JSON: %v", err)
	}

	result, err := service.convertMCPSchemaToFunctionCall(ctx, inputMap)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result["type"] != "object" {
		t.Errorf("Expected type object, got %v", result["type"])
	}

	// Case 2: With $defs
	inputJSON = `{
		"$defs": {
			"Person": {
				"type": "object",
				"properties": {
					"name": {"type": "string"}
				}
			}
		},
		"properties": {
			"owner": {"$ref": "#/$defs/Person"}
		}
	}`
	if err := json.Unmarshal([]byte(inputJSON), &inputMap); err != nil {
		t.Fatalf("Failed to unmarshal test JSON: %v", err)
	}
	result, err = service.convertMCPSchemaToFunctionCall(ctx, inputMap)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	props := result["properties"].(map[string]interface{})
	owner := props["owner"].(map[string]interface{})
	if owner["type"] != "object" {
		t.Errorf("Expected owner type object, got %v", owner["type"])
	}
	ownerProps := owner["properties"].(map[string]interface{})
	if _, ok := ownerProps["name"]; !ok {
		t.Errorf("Expected owner to have name property")
	}

	// Check $defs is removed
	if _, ok := result["$defs"]; ok {
		t.Errorf("Expected $defs to be removed")
	}
}

func TestResolveMCPSchemaCircular(t *testing.T) {
	service := &knActionRecallServiceImpl{
		logger: &mockLogger{},
	}

	ctx := context.Background()

	// Case 3: Circular Reference
	inputJSON := `{
		"$defs": {
			"Node": {
				"type": "object",
				"properties": {
					"child": {"$ref": "#/$defs/Node"}
				}
			}
		},
		"properties": {
			"root": {"$ref": "#/$defs/Node"}
		}
	}`
	var inputMap map[string]interface{}
	if err := json.Unmarshal([]byte(inputJSON), &inputMap); err != nil {
		t.Fatalf("Failed to unmarshal test JSON: %v", err)
	}

	result, err := service.convertMCPSchemaToFunctionCall(ctx, inputMap)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should not crash and should prune
	props := result["properties"].(map[string]interface{})
	root := props["root"].(map[string]interface{})
	rootProps := root["properties"].(map[string]interface{})
	child := rootProps["child"].(map[string]interface{})

	// Child should be pruned (no properties) or recursively resolved up to depth limit
	// Since circular detection is immediate for same path in visitedRefs
	// Root visits Node. Node visits Child (Node).
	// If depth limit is 3, it might expand a bit.
	// But visitedRefs checks path.
	// resolveMCPSchema calls resolveMCPSchema for ref.
	// visitedRefs is passed.
	// root -> Node (visited["#/$defs/Node"] = true)
	// Node.properties.child -> ref "#/$defs/Node"
	// check visited -> true -> prune.
	// So child should be pruned.

	if _, ok := child["properties"]; ok {
		// If it's pruned, it shouldn't have properties
		t.Errorf("Expected circular reference to be pruned, but found properties")
	}
}
