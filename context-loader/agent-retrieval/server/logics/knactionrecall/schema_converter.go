// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knactionrecall

import (
	"context"
	"fmt"
	"strings"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

const (
	// MaxSchemaDepth 最大 $ref 引用递归深度，用于防止过深嵌套和循环引用导致的无限递归
	// 作用：
	// 1. 限制非循环的深层嵌套引用（如 A -> B -> C -> D）
	// 2. 作为循环引用的第二道防线（循环引用检测优先触发）
	// 建议值：2-3 层
	// - 2 层：适合简单场景（如树形结构）
	// - 3 层：适合复杂场景（多层嵌套）
	MaxSchemaDepth = 3
)

// convertSchemaToFunctionCall 将 OpenAPI Schema 转换为 OpenAI Function Call Schema
// 改进：保持分层结构（header/path/query/body），而不是扁平化
func (s *knActionRecallServiceImpl) convertSchemaToFunctionCall(ctx context.Context, apiSpec map[string]interface{}) (map[string]interface{}, error) {
	// 使用分层结构：header/path/query/body
	properties := map[string]interface{}{
		"header": map[string]interface{}{
			"type":        "object",
			"description": "HTTP Header 参数",
			"properties":  make(map[string]interface{}),
		},
		"path": map[string]interface{}{
			"type":        "object",
			"description": "URL Path 参数",
			"properties":  make(map[string]interface{}),
		},
		"query": map[string]interface{}{
			"type":        "object",
			"description": "URL Query 参数",
			"properties":  make(map[string]interface{}),
		},
		"body": map[string]interface{}{
			"type":        "object",
			"description": "Request Body 参数",
			"properties":  make(map[string]interface{}),
		},
	}

	// 各位置的必填参数
	requiredByLocation := map[string][]string{
		"header": {},
		"path":   {},
		"query":  {},
		"body":   {},
	}

	// 用于循环引用检测的访问记录
	visitedRefs := make(map[string]bool)

	// 1. 处理 parameters (path/query/header)
	if params, ok := apiSpec["parameters"].([]interface{}); ok {
		for _, paramItem := range params {
			param, ok := paramItem.(map[string]interface{})
			if !ok {
				continue
			}

			paramName, _ := param["name"].(string)
			if paramName == "" {
				continue
			}

			paramLocation, _ := param["in"].(string) // path/query/header
			if paramLocation == "" {
				continue
			}

			// 解析参数 schema（支持 $ref，支持深度控制）
			paramSchema, err := s.resolveSchema(ctx, param["schema"], apiSpec, visitedRefs, 0)
			if err != nil {
				s.logger.WithContext(ctx).Warnf("[KnActionRecall#convertSchema] Failed to resolve param schema for %s: %v", paramName, err)
				continue
			}

			// 构建参数定义
			propDef := s.buildPropertyDefinition(paramSchema, param["description"])

			// 根据位置放入对应的 properties
			if locationProps, ok := properties[paramLocation].(map[string]interface{}); ok {
				if props, ok := locationProps["properties"].(map[string]interface{}); ok {
					props[paramName] = propDef
				}
			}

			// 收集必填参数
			if isRequired, ok := param["required"].(bool); ok && isRequired {
				requiredByLocation[paramLocation] = append(requiredByLocation[paramLocation], paramName)
			}
		}
	}

	// 2. 处理 request_body (body 参数)
	if requestBody, ok := apiSpec["request_body"].(map[string]interface{}); ok {
		if content, ok := requestBody["content"].(map[string]interface{}); ok {
			if appJSON, ok := content["application/json"].(map[string]interface{}); ok {
				if schema, ok := appJSON["schema"].(map[string]interface{}); ok {
					// 解析 body schema（支持 $ref，支持深度控制）
					bodySchema, err := s.resolveSchema(ctx, schema, apiSpec, visitedRefs, 0)
					if err != nil {
						s.logger.WithContext(ctx).Warnf("[KnActionRecall#convertSchema] Failed to resolve body schema: %v", err)
						// 添加一个通用的 body 参数作为兜底
						if bodyProps, ok := properties["body"].(map[string]interface{}); ok {
							if props, ok := bodyProps["properties"].(map[string]interface{}); ok {
								props["request_body"] = map[string]interface{}{
									"type":        "object",
									"description": "请求体参数，详见 original_schema",
								}
							}
						}
					} else {
						// 展开 body schema 的 properties
						if bodyProps, ok := properties["body"].(map[string]interface{}); ok {
							if props, ok := bodyProps["properties"].(map[string]interface{}); ok {
								s.mergeSchemaProperties(ctx, props, bodySchema, apiSpec, visitedRefs, 0)
							}
							// 合并必填项
							if bodyRequired, ok := bodySchema["required"].([]interface{}); ok {
								for _, req := range bodyRequired {
									if reqStr, ok := req.(string); ok {
										requiredByLocation["body"] = append(requiredByLocation["body"], reqStr)
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// 3. 设置各位置的 required 字段
	for location, required := range requiredByLocation {
		if len(required) > 0 {
			if locationProps, ok := properties[location].(map[string]interface{}); ok {
				locationProps["required"] = required
			}
		}
	}

	// 4. 清理空的 location（如果没有参数，移除该位置）
	result := map[string]interface{}{
		"type":       "object",
		"properties": make(map[string]interface{}),
	}

	resultProps := result["properties"].(map[string]interface{})
	for location, locationProps := range properties {
		if props, ok := locationProps.(map[string]interface{})["properties"].(map[string]interface{}); ok {
			if len(props) > 0 {
				resultProps[location] = locationProps
			}
		}
	}

	// 如果没有任何参数，至少返回一个空结构
	if len(resultProps) == 0 {
		resultProps["body"] = map[string]interface{}{
			"type":        "object",
			"description": "Request Body 参数",
			"properties":  make(map[string]interface{}),
		}
	}

	return result, nil
}

// resolveSchema 解析 schema，支持 $ref 引用、循环引用检测和深度控制
// 采用深度限制剪枝策略：
// - 每次解析 $ref 时，深度 +1
// - 解析 properties 中的属性时，深度不变（同一层级）
// - 达到最大深度时，执行剪枝（保留类型和原始描述，移除 properties）
// currentDepth: 当前递归深度，用于控制循环引用的展开深度
func (s *knActionRecallServiceImpl) resolveSchema(ctx context.Context, schema interface{}, apiSpec map[string]interface{}, visitedRefs map[string]bool, currentDepth int) (map[string]interface{}, error) {
	if schema == nil {
		return map[string]interface{}{"type": "string"}, nil
	}

	schemaMap, ok := schema.(map[string]interface{})
	if !ok {
		return map[string]interface{}{"type": "string"}, nil
	}

	// 如果直接有 type 且没有 $ref，直接返回
	if _, hasType := schemaMap["type"]; hasType && schemaMap["$ref"] == nil {
		// 如果有 properties，需要递归处理（深度不变）
		if props, ok := schemaMap["properties"].(map[string]interface{}); ok {
			resolvedProps := make(map[string]interface{})
			for propName, propDef := range props {
				resolvedProp, err := s.resolveSchema(ctx, propDef, apiSpec, visitedRefs, currentDepth)
				if err != nil {
					s.logger.WithContext(ctx).Warnf("[KnActionRecall#resolveSchema] Failed to resolve property %s: %v", propName, err)
					continue
				}
				resolvedProps[propName] = resolvedProp
			}
			schemaMap["properties"] = resolvedProps
		}
		// 处理 array.items（深度不变）
		if schemaMap["type"] == "array" {
			if items, ok := schemaMap["items"].(map[string]interface{}); ok {
				resolvedItems, err := s.resolveSchema(ctx, items, apiSpec, visitedRefs, currentDepth)
				if err != nil {
					s.logger.WithContext(ctx).Warnf("[KnActionRecall#resolveSchema] Failed to resolve array items: %v", err)
				} else {
					schemaMap["items"] = resolvedItems
				}
			}
		}
		return schemaMap, nil
	}

	// 处理 $ref 引用
	if refPath, ok := schemaMap["$ref"].(string); ok {
		// 检查循环引用（必须在深度检查之前，避免无限递归）
		if visitedRefs[refPath] {
			// 检测到循环引用，执行剪枝
			s.logger.WithContext(ctx).Debugf("[KnActionRecall#resolveSchema] Circular reference detected for %s (depth: %d), pruning", refPath, currentDepth)
			// 获取被引用的 schema 基本信息，然后剪枝
			referencedSchema, err := s.getReferencedSchema(refPath, apiSpec)
			if err != nil {
				s.logger.WithContext(ctx).Warnf("[KnActionRecall#resolveSchema] Failed to get referenced schema for pruning: %v", err)
				return map[string]interface{}{"type": "object"}, nil
			}
			return s.pruneSchema(referencedSchema), nil
		}

		// 检查是否达到最大深度
		if currentDepth >= MaxSchemaDepth {
			s.logger.WithContext(ctx).Debugf("[KnActionRecall#resolveSchema] Max depth reached for %s (depth: %d), pruning", refPath, currentDepth)
			// 获取被引用的 schema 基本信息，然后剪枝
			referencedSchema, err := s.getReferencedSchema(refPath, apiSpec)
			if err != nil {
				s.logger.WithContext(ctx).Warnf("[KnActionRecall#resolveSchema] Failed to get referenced schema for pruning: %v", err)
				return map[string]interface{}{"type": "object"}, nil
			}
			return s.pruneSchema(referencedSchema), nil
		}

		// 标记为已访问
		wasVisited := visitedRefs[refPath]
		visitedRefs[refPath] = true
		defer func() {
			// 递归返回时，如果这是第一次访问，清理标记
			if !wasVisited {
				delete(visitedRefs, refPath)
			}
		}()

		// 解析 $ref 路径（深度 +1）
		resolvedSchema, err := s.resolveDollarRef(ctx, refPath, apiSpec, visitedRefs, currentDepth+1)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve $ref %s: %w", refPath, err)
		}

		return resolvedSchema, nil
	}

	// 如果有 properties，递归处理（深度不变，同一层级）
	if props, ok := schemaMap["properties"].(map[string]interface{}); ok {
		resolvedProps := make(map[string]interface{})
		for propName, propDef := range props {
			resolvedProp, err := s.resolveSchema(ctx, propDef, apiSpec, visitedRefs, currentDepth)
			if err != nil {
				s.logger.WithContext(ctx).Warnf("[KnActionRecall#resolveSchema] Failed to resolve property %s: %v", propName, err)
				continue
			}
			resolvedProps[propName] = resolvedProp
		}
		schemaMap["properties"] = resolvedProps
	}

	// 处理 array.items（深度不变）
	if schemaMap["type"] == "array" {
		if items, ok := schemaMap["items"].(map[string]interface{}); ok {
			resolvedItems, err := s.resolveSchema(ctx, items, apiSpec, visitedRefs, currentDepth)
			if err != nil {
				s.logger.WithContext(ctx).Warnf("[KnActionRecall#resolveSchema] Failed to resolve array items: %v", err)
			} else {
				schemaMap["items"] = resolvedItems
			}
		}
	}

	return schemaMap, nil
}

// getReferencedSchema 获取被引用的 schema 定义（不解析，只获取基本信息）
func (s *knActionRecallServiceImpl) getReferencedSchema(refPath string, apiSpec map[string]interface{}) (map[string]interface{}, error) {
	// 解析 $ref 路径格式：#/components/schemas/SchemaName
	if !strings.HasPrefix(refPath, "#/components/schemas/") {
		return nil, fmt.Errorf("unsupported $ref path format: %s (only #/components/schemas/* is supported)", refPath)
	}

	schemaName := strings.TrimPrefix(refPath, "#/components/schemas/")
	if schemaName == "" {
		return nil, fmt.Errorf("empty schema name in $ref: %s", refPath)
	}

	// 从 components.schemas 中查找
	components, ok := apiSpec["components"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("components not found in api_spec")
	}

	schemas, ok := components["schemas"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("components.schemas not found in api_spec")
	}

	schema, ok := schemas[schemaName].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("schema %s not found in components.schemas", schemaName)
	}

	return schema, nil
}

// pruneSchema 剪枝函数：当达到最大深度时，保留类型和原始描述，移除 properties
// 核心策略：不添加循环引用说明，节省 token
func (s *knActionRecallServiceImpl) pruneSchema(schema map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// 保留类型信息
	if schemaType, ok := schema["type"].(string); ok && schemaType != "" {
		result["type"] = schemaType
	} else {
		result["type"] = "object" // 默认类型
	}

	// 保留原始 description（如果存在，不修改，不添加循环引用说明）
	if desc, ok := schema["description"].(string); ok && desc != "" {
		result["description"] = desc
	}

	// 如果是 array，保留 items 结构但不展开 properties
	if result["type"] == "array" {
		if items, ok := schema["items"].(map[string]interface{}); ok {
			// 递归剪枝 items
			result["items"] = s.pruneSchema(items)
		}
	}

	// 不包含 properties（避免继续递归）
	// 不添加循环引用说明（节省 token）

	return result
}

// resolveDollarRef 解析 $ref 引用（完整实现，支持循环引用检测和深度控制）
func (s *knActionRecallServiceImpl) resolveDollarRef(ctx context.Context, refPath string, apiSpec map[string]interface{}, visitedRefs map[string]bool, currentDepth int) (map[string]interface{}, error) {
	// 获取被引用的 schema
	schema, err := s.getReferencedSchema(refPath, apiSpec)
	if err != nil {
		return nil, err
	}

	// 递归解析（可能包含嵌套的 $ref，传递深度信息）
	return s.resolveSchema(ctx, schema, apiSpec, visitedRefs, currentDepth)
}

// buildPropertyDefinition 构建属性定义
func (s *knActionRecallServiceImpl) buildPropertyDefinition(schema map[string]interface{}, description interface{}) map[string]interface{} {
	propDef := make(map[string]interface{})

	// 类型
	if propType, ok := schema["type"].(string); ok && propType != "" {
		propDef["type"] = propType
	} else {
		propDef["type"] = "string" // 默认类型
	}

	// 描述（优先使用参数级别的 description，其次使用 schema 中的 description）
	if desc, ok := description.(string); ok && desc != "" {
		propDef["description"] = desc
	} else if desc, ok := schema["description"].(string); ok && desc != "" {
		propDef["description"] = desc
	}

	// 枚举
	if enum, ok := schema["enum"].([]interface{}); ok {
		propDef["enum"] = enum
	}

	// 如果 schema 有 properties，保留嵌套结构
	if props, ok := schema["properties"].(map[string]interface{}); ok {
		propDef["properties"] = props
		propDef["type"] = "object"
	}

	// 如果 schema 是 array，保留 items 结构
	if schema["type"] == "array" {
		if items, ok := schema["items"].(map[string]interface{}); ok {
			propDef["items"] = items
		}
	}

	return propDef
}

// mergeSchemaProperties 合并 schema 的 properties 到目标 properties
func (s *knActionRecallServiceImpl) mergeSchemaProperties(ctx context.Context, targetProps map[string]interface{}, schema map[string]interface{}, apiSpec map[string]interface{}, visitedRefs map[string]bool, currentDepth int) {
	if props, ok := schema["properties"].(map[string]interface{}); ok {
		for propName, propDef := range props {
			resolvedProp, err := s.resolveSchema(ctx, propDef, apiSpec, visitedRefs, currentDepth)
			if err != nil {
				s.logger.WithContext(ctx).Warnf("[KnActionRecall#mergeSchemaProperties] Failed to resolve property %s: %v", propName, err)
				continue
			}
			targetProps[propName] = s.buildPropertyDefinition(resolvedProp, nil)
		}
	}
}

// mapFixedParams 映射固定参数到 header/path/query/body
func (s *knActionRecallServiceImpl) mapFixedParams(ctx context.Context, parameters map[string]interface{}, apiSpec map[string]interface{}) interfaces.KnFixedParams {
	fixedParams := interfaces.KnFixedParams{
		Header: make(map[string]interface{}),
		Path:   make(map[string]interface{}),
		Query:  make(map[string]interface{}),
		Body:   make(map[string]interface{}),
	}

	// 建立参数名到位置的映射表
	paramLocationMap := make(map[string]string)
	if params, ok := apiSpec["parameters"].([]interface{}); ok {
		for _, paramItem := range params {
			if param, ok := paramItem.(map[string]interface{}); ok {
				if name, ok := param["name"].(string); ok {
					if in, ok := param["in"].(string); ok {
						paramLocationMap[name] = in
					}
				}
			}
		}
	}

	// 根据映射表分类参数
	for key, value := range parameters {
		location := paramLocationMap[key]
		switch location {
		case "header":
			fixedParams.Header[key] = value
		case "path":
			fixedParams.Path[key] = value
		case "query":
			fixedParams.Query[key] = value
		case "body":
			fixedParams.Body[key] = value
		default:
			// 未找到映射，使用命名规则判断
			if isHeaderParam(key) {
				fixedParams.Header[key] = value
			} else {
				// 默认放入 body
				fixedParams.Body[key] = value
			}
		}
	}

	return fixedParams
}

// isHeaderParam 判断是否为 header 参数（基于命名规则）
func isHeaderParam(key string) bool {
	// 常见的 header 参数名称模式
	headerPatterns := []string{
		"x-", "X-",
		"authorization", "Authorization",
		"content-type", "Content-Type",
	}

	for _, pattern := range headerPatterns {
		if len(key) >= len(pattern) && key[:len(pattern)] == pattern {
			return true
		}
	}

	return false
}
