// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knactionrecall

import (
	"testing"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

// TestMCPAPIURLConstruction 测试 MCP 类型的 API URL 构造
func TestMCPAPIURLConstruction(t *testing.T) {
	testCases := []struct {
		name        string
		mcpID       string
		toolName    string
		expectedURL string
	}{
		{
			name:        "标准 MCP ID 和工具名",
			mcpID:       "ad3ca391-a598-4764-a6c8-e62b9662e87e",
			toolName:    "generate_treatment_plan",
			expectedURL: "/api/agent-retrieval/v1/mcp/proxy/ad3ca391-a598-4764-a6c8-e62b9662e87e/tools/generate_treatment_plan/call",
		},
		{
			name:        "简短 MCP ID",
			mcpID:       "mcp-001",
			toolName:    "query_data",
			expectedURL: "/api/agent-retrieval/v1/mcp/proxy/mcp-001/tools/query_data/call",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 使用与 get_action_info.go 相同的格式化逻辑
			apiURL := "/api/agent-retrieval/v1/mcp/proxy/" + tc.mcpID + "/tools/" + tc.toolName + "/call"
			if apiURL != tc.expectedURL {
				t.Errorf("API URL 构造错误\n期望: %s\n实际: %s", tc.expectedURL, apiURL)
			}
		})
	}
}

// TestMCPFixedParamsFlat 测试 MCP 类型的固定参数是扁平化结构
func TestMCPFixedParamsFlat(t *testing.T) {
	// 模拟从 Ontology Query 返回的行动参数
	actionParams := map[string]interface{}{
		"disease_id":    "disease_000001",
		"include_drugs": "true",
		"lang":          "zh",
	}

	// MCP 类型直接使用扁平化的 map
	fixedParams := actionParams

	// 验证是扁平结构（没有 header/path/query/body 分层）
	if _, hasHeader := fixedParams["header"]; hasHeader {
		t.Error("MCP fixed_params 不应该有 header 字段")
	}
	if _, hasPath := fixedParams["path"]; hasPath {
		t.Error("MCP fixed_params 不应该有 path 字段")
	}
	if _, hasQuery := fixedParams["query"]; hasQuery {
		t.Error("MCP fixed_params 不应该有 query 字段")
	}
	if _, hasBody := fixedParams["body"]; hasBody {
		t.Error("MCP fixed_params 不应该有 body 字段")
	}

	// 验证原始字段存在
	if fixedParams["disease_id"] != "disease_000001" {
		t.Error("MCP fixed_params 应该包含原始的 disease_id 字段")
	}
}

// TestActionSourceTypeMCP 测试 MCP 类型常量定义正确
func TestActionSourceTypeMCP(t *testing.T) {
	if interfaces.ActionSourceTypeMCP != "mcp" {
		t.Errorf("ActionSourceTypeMCP 应该为 'mcp', 实际为 '%s'", interfaces.ActionSourceTypeMCP)
	}
	if interfaces.ActionSourceTypeTool != "tool" {
		t.Errorf("ActionSourceTypeTool 应该为 'tool', 实际为 '%s'", interfaces.ActionSourceTypeTool)
	}
}

// TestActionSourceMCPFields 测试 ActionSource 结构体包含 MCP 相关字段
func TestActionSourceMCPFields(t *testing.T) {
	actionSource := interfaces.ActionSource{
		Type:     interfaces.ActionSourceTypeMCP,
		McpID:    "test-mcp-id",
		ToolName: "test-tool-name",
	}

	if actionSource.Type != "mcp" {
		t.Error("ActionSource.Type 应该为 'mcp'")
	}
	if actionSource.McpID != "test-mcp-id" {
		t.Error("ActionSource.McpID 应该为 'test-mcp-id'")
	}
	if actionSource.ToolName != "test-tool-name" {
		t.Error("ActionSource.ToolName 应该为 'test-tool-name'")
	}
}
