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

// TestConvertMCPSchemaToFunctionCall_BodyDefaultDescription 测试 MCP Schema 转换时 body 参数默认描述逻辑
// 规则：当第一层存在 body 参数但缺少 description 时，自动添加 "Request Body参数"
func TestConvertMCPSchemaToFunctionCall_BodyDefaultDescription(t *testing.T) {
	service := &knActionRecallServiceImpl{
		logger: &mockLogger{},
	}

	ctx := context.Background()

	// Case 1: body 存在但通过 $ref 引用，引用的 schema 没有 description
	// 期望：自动添加默认描述 "Request Body参数"
	t.Run("body_without_description_via_ref", func(t *testing.T) {
		inputJSON := `{
			"$defs": {
				"UpdateEventStatusRequest": {
					"type": "object",
					"properties": {
						"status": {"type": "string"}
					}
				}
			},
			"type": "object",
			"properties": {
				"body": {"$ref": "#/$defs/UpdateEventStatusRequest"},
				"path": {
					"type": "object",
					"description": "URL 路径参数",
					"properties": {
						"event_id": {"type": "string"}
					}
				}
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

		props := result["properties"].(map[string]interface{})
		body := props["body"].(map[string]interface{})

		// 验证 body 添加了默认描述
		if desc, ok := body["description"].(string); !ok || desc != "Request Body参数" {
			t.Errorf("Expected body description 'Request Body参数', got %v", body["description"])
		}

		// 验证 path 保持原有描述
		path := props["path"].(map[string]interface{})
		if desc, ok := path["description"].(string); !ok || desc != "URL 路径参数" {
			t.Errorf("Expected path description 'URL 路径参数', got %v", path["description"])
		}
	})

	// Case 2: body 存在且已有 description
	// 期望：保留原有描述，不覆盖
	t.Run("body_with_existing_description", func(t *testing.T) {
		inputJSON := `{
			"type": "object",
			"properties": {
				"body": {
					"type": "object",
					"description": "自定义请求体描述",
					"properties": {
						"name": {"type": "string"}
					}
				}
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

		props := result["properties"].(map[string]interface{})
		body := props["body"].(map[string]interface{})

		// 验证保留原有描述
		if desc, ok := body["description"].(string); !ok || desc != "自定义请求体描述" {
			t.Errorf("Expected body description '自定义请求体描述', got %v", body["description"])
		}
	})

	// Case 3: 没有 body 参数
	// 期望：不做任何处理，不报错
	t.Run("no_body_property", func(t *testing.T) {
		inputJSON := `{
			"type": "object",
			"properties": {
				"query": {
					"type": "object",
					"properties": {
						"limit": {"type": "integer"}
					}
				}
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

		props := result["properties"].(map[string]interface{})

		// 验证 body 不存在
		if _, ok := props["body"]; ok {
			t.Error("Expected no body property, but found one")
		}

		// 验证 query 存在
		if _, ok := props["query"]; !ok {
			t.Error("Expected query property to exist")
		}
	})

	// Case 4: body 直接定义（非 $ref）且无 description
	// 期望：自动添加默认描述 "Request Body参数"
	t.Run("body_direct_without_description", func(t *testing.T) {
		inputJSON := `{
			"type": "object",
			"properties": {
				"body": {
					"type": "object",
					"properties": {
						"name": {"type": "string"}
					}
				}
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

		props := result["properties"].(map[string]interface{})
		body := props["body"].(map[string]interface{})

		// 验证 body 添加了默认描述
		if desc, ok := body["description"].(string); !ok || desc != "Request Body参数" {
			t.Errorf("Expected body description 'Request Body参数', got %v", body["description"])
		}
	})
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

// TestConvertSchemaToFunctionCall_WithParameters 测试带参数的 OpenAPI Schema 转换
func TestConvertSchemaToFunctionCall_WithParameters(t *testing.T) {
	service := &knActionRecallServiceImpl{
		logger: &mockLogger{},
	}

	ctx := context.Background()

	apiSpec := map[string]interface{}{
		"parameters": []interface{}{
			map[string]interface{}{
				"name":        "id",
				"in":          "path",
				"required":    true,
				"description": "资源ID",
				"schema":      map[string]interface{}{"type": "string"},
			},
			map[string]interface{}{
				"name":        "limit",
				"in":          "query",
				"required":    false,
				"description": "返回数量限制",
				"schema":      map[string]interface{}{"type": "integer"},
			},
			map[string]interface{}{
				"name":        "X-Request-ID",
				"in":          "header",
				"required":    true,
				"description": "请求ID",
				"schema":      map[string]interface{}{"type": "string"},
			},
		},
	}

	result, err := service.convertSchemaToFunctionCall(ctx, apiSpec)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result["type"] != "object" {
		t.Errorf("Expected type object, got %v", result["type"])
	}

	props := result["properties"].(map[string]interface{})

	// 检查 path 参数
	if pathProps, ok := props["path"].(map[string]interface{}); ok {
		pathParams := pathProps["properties"].(map[string]interface{})
		if _, ok := pathParams["id"]; !ok {
			t.Error("Expected path to have id parameter")
		}
	} else {
		t.Error("Expected path to exist in properties")
	}

	// 检查 query 参数
	if queryProps, ok := props["query"].(map[string]interface{}); ok {
		queryParams := queryProps["properties"].(map[string]interface{})
		if _, ok := queryParams["limit"]; !ok {
			t.Error("Expected query to have limit parameter")
		}
	} else {
		t.Error("Expected query to exist in properties")
	}

	// 检查 header 参数
	if headerProps, ok := props["header"].(map[string]interface{}); ok {
		headerParams := headerProps["properties"].(map[string]interface{})
		if _, ok := headerParams["X-Request-ID"]; !ok {
			t.Error("Expected header to have X-Request-ID parameter")
		}
	} else {
		t.Error("Expected header to exist in properties")
	}
}

// TestConvertSchemaToFunctionCall_WithRequestBody 测试带 request_body 的 Schema 转换
func TestConvertSchemaToFunctionCall_WithRequestBody(t *testing.T) {
	service := &knActionRecallServiceImpl{
		logger: &mockLogger{},
	}

	ctx := context.Background()

	apiSpec := map[string]interface{}{
		"request_body": map[string]interface{}{
			"content": map[string]interface{}{
				"application/json": map[string]interface{}{
					"schema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"name": map[string]interface{}{
								"type":        "string",
								"description": "名称",
							},
							"age": map[string]interface{}{
								"type":        "integer",
								"description": "年龄",
							},
						},
						"required": []interface{}{"name"},
					},
				},
			},
		},
	}

	result, err := service.convertSchemaToFunctionCall(ctx, apiSpec)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	props := result["properties"].(map[string]interface{})

	// 检查 body 参数
	if bodyProps, ok := props["body"].(map[string]interface{}); ok {
		bodyParams := bodyProps["properties"].(map[string]interface{})
		if _, ok := bodyParams["name"]; !ok {
			t.Error("Expected body to have name parameter")
		}
		if _, ok := bodyParams["age"]; !ok {
			t.Error("Expected body to have age parameter")
		}
	} else {
		t.Error("Expected body to exist in properties")
	}
}

// TestConvertSchemaToFunctionCall_Empty 测试空 Schema
func TestConvertSchemaToFunctionCall_Empty(t *testing.T) {
	service := &knActionRecallServiceImpl{
		logger: &mockLogger{},
	}

	ctx := context.Background()

	apiSpec := map[string]interface{}{}

	result, err := service.convertSchemaToFunctionCall(ctx, apiSpec)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result["type"] != "object" {
		t.Errorf("Expected type object, got %v", result["type"])
	}

	props := result["properties"].(map[string]interface{})
	// 空 Schema 应该至少有一个 body 字段
	if _, ok := props["body"]; !ok {
		t.Error("Expected body to exist in properties for empty schema")
	}
}

// TestMapFixedParams_AllLocations 测试固定参数映射到所有位置
func TestMapFixedParams_AllLocations(t *testing.T) {
	service := &knActionRecallServiceImpl{
		logger: &mockLogger{},
	}

	ctx := context.Background()

	parameters := map[string]interface{}{
		"id":           "123",
		"limit":        10,
		"X-Request-ID": "req-001",
		"name":         "test",
	}

	apiSpec := map[string]interface{}{
		"parameters": []interface{}{
			map[string]interface{}{"name": "id", "in": "path"},
			map[string]interface{}{"name": "limit", "in": "query"},
			map[string]interface{}{"name": "X-Request-ID", "in": "header"},
		},
	}

	result := service.mapFixedParams(ctx, parameters, apiSpec)

	// 检查 path 参数
	if result.Path["id"] != "123" {
		t.Errorf("Expected path[id] = '123', got %v", result.Path["id"])
	}

	// 检查 query 参数
	if result.Query["limit"] != 10 {
		t.Errorf("Expected query[limit] = 10, got %v", result.Query["limit"])
	}

	// 检查 header 参数
	if result.Header["X-Request-ID"] != "req-001" {
		t.Errorf("Expected header[X-Request-ID] = 'req-001', got %v", result.Header["X-Request-ID"])
	}

	// 检查未映射的参数进入 body
	if result.Body["name"] != "test" {
		t.Errorf("Expected body[name] = 'test', got %v", result.Body["name"])
	}
}

// TestMapFixedParams_HeaderByNaming 测试通过命名规则判断 header 参数
func TestMapFixedParams_HeaderByNaming(t *testing.T) {
	service := &knActionRecallServiceImpl{
		logger: &mockLogger{},
	}

	ctx := context.Background()

	parameters := map[string]interface{}{
		"x-custom-header": "value1",
		"Authorization":   "Bearer token",
		"normal-param":    "value2",
	}

	apiSpec := map[string]interface{}{} // 没有参数定义

	result := service.mapFixedParams(ctx, parameters, apiSpec)

	// x- 开头的参数应该进入 header
	if result.Header["x-custom-header"] != "value1" {
		t.Errorf("Expected header[x-custom-header] = 'value1', got %v", result.Header["x-custom-header"])
	}

	// Authorization 应该进入 header
	if result.Header["Authorization"] != "Bearer token" {
		t.Errorf("Expected header[Authorization] = 'Bearer token', got %v", result.Header["Authorization"])
	}

	// 普通参数应该进入 body
	if result.Body["normal-param"] != "value2" {
		t.Errorf("Expected body[normal-param] = 'value2', got %v", result.Body["normal-param"])
	}
}

// TestIsHeaderParam 测试 header 参数判断
func TestIsHeaderParam(t *testing.T) {
	testCases := []struct {
		key      string
		expected bool
	}{
		{"x-custom-header", true},
		{"X-Request-ID", true},
		{"authorization", true},
		{"Authorization", true},
		{"content-type", true},
		{"Content-Type", true},
		{"normal-param", false},
		{"id", false},
		{"name", false},
	}

	for _, tc := range testCases {
		t.Run(tc.key, func(t *testing.T) {
			result := isHeaderParam(tc.key)
			if result != tc.expected {
				t.Errorf("isHeaderParam(%s) = %v, expected %v", tc.key, result, tc.expected)
			}
		})
	}
}

// TestBuildPropertyDefinition 测试属性定义构建
func TestBuildPropertyDefinition(t *testing.T) {
	service := &knActionRecallServiceImpl{
		logger: &mockLogger{},
	}

	// 测试基本类型
	schema := map[string]interface{}{
		"type":        "string",
		"description": "测试描述",
	}
	result := service.buildPropertyDefinition(schema, nil)
	if result["type"] != "string" {
		t.Errorf("Expected type string, got %v", result["type"])
	}
	if result["description"] != "测试描述" {
		t.Errorf("Expected description '测试描述', got %v", result["description"])
	}

	// 测试带枚举
	schema = map[string]interface{}{
		"type": "string",
		"enum": []interface{}{"a", "b", "c"},
	}
	result = service.buildPropertyDefinition(schema, nil)
	if result["enum"] == nil {
		t.Error("Expected enum to be preserved")
	}

	// 测试带 properties 的对象
	schema = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{"type": "string"},
		},
	}
	result = service.buildPropertyDefinition(schema, nil)
	if result["type"] != "object" {
		t.Errorf("Expected type object, got %v", result["type"])
	}
	if result["properties"] == nil {
		t.Error("Expected properties to be preserved")
	}

	// 测试数组类型
	schema = map[string]interface{}{
		"type": "array",
		"items": map[string]interface{}{
			"type": "string",
		},
	}
	result = service.buildPropertyDefinition(schema, nil)
	if result["items"] == nil {
		t.Error("Expected items to be preserved for array type")
	}

	// 测试参数级别描述覆盖 schema 描述
	schema = map[string]interface{}{
		"type":        "string",
		"description": "schema描述",
	}
	result = service.buildPropertyDefinition(schema, "参数描述")
	if result["description"] != "参数描述" {
		t.Errorf("Expected param description to override schema description, got %v", result["description"])
	}
}

// TestPruneSchema 测试 schema 剪枝
func TestPruneSchema(t *testing.T) {
	service := &knActionRecallServiceImpl{
		logger: &mockLogger{},
	}

	// 测试基本剪枝
	schema := map[string]interface{}{
		"type":        "object",
		"description": "测试对象",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{"type": "string"},
		},
	}
	result := service.pruneSchema(schema)
	if result["type"] != "object" {
		t.Errorf("Expected type object, got %v", result["type"])
	}
	if result["description"] != "测试对象" {
		t.Errorf("Expected description '测试对象', got %v", result["description"])
	}
	if _, hasProps := result["properties"]; hasProps {
		t.Error("Expected properties to be removed after pruning")
	}

	// 测试数组类型剪枝
	schema = map[string]interface{}{
		"type": "array",
		"items": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{"type": "string"},
			},
		},
	}
	result = service.pruneSchema(schema)
	if result["type"] != "array" {
		t.Errorf("Expected type array, got %v", result["type"])
	}
	if result["items"] == nil {
		t.Error("Expected items to be preserved for array")
	}
	items := result["items"].(map[string]interface{})
	if _, hasProps := items["properties"]; hasProps {
		t.Error("Expected items properties to be removed after pruning")
	}

	// 测试无类型 schema
	schema = map[string]interface{}{
		"description": "无类型",
	}
	result = service.pruneSchema(schema)
	if result["type"] != "object" {
		t.Errorf("Expected default type object, got %v", result["type"])
	}
}
