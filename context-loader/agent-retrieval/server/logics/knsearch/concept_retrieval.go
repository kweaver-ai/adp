// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package knsearch Schema召回
// file: concept_retrieval.go
// 使用LLM筛选与用户查询相关的关系类型和对象类型
package knsearch

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/config"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

// minMatchLength 索引匹配的最小返回长度
const minMatchLength = 2

// ConceptRetrieval Schema召回器
type ConceptRetrieval struct {
	logger        interfaces.Logger
	mfModelClient interfaces.DrivenMFModelAPIClient
	config        *config.RerankLLMConfig
}

// NewConceptRetrieval 创建Schema召回器
func NewConceptRetrieval(logger interfaces.Logger, mfModelClient interfaces.DrivenMFModelAPIClient) *ConceptRetrieval {
	conf := config.NewConfigLoader()
	return &ConceptRetrieval{
		logger:        logger,
		mfModelClient: mfModelClient,
		config:        &conf.RerankLLM,
	}
}

// SchemaRetrievalResult Schema召回结果
type SchemaRetrievalResult struct {
	ObjectTypes   []map[string]interface{} // 过滤后的对象类型
	RelationTypes []map[string]interface{} // 过滤后的关系类型
}

// Retrieve 执行Schema召回
func (c *ConceptRetrieval) Retrieve(
	ctx context.Context,
	query string,
	networkDetails map[string]interface{},
	topK int,
	accountID, accountType string,
) (*SchemaRetrievalResult, error) {
	c.logger.WithContext(ctx).Debugf("[ConceptRetrieval#Retrieve] Starting schema retrieval for query: %s", query)

	// 提取关系类型和对象类型
	relationTypes := c.extractRelationTypes(networkDetails)
	objectTypes := c.extractObjectTypes(networkDetails)

	if len(relationTypes) == 0 {
		// 无关系类型场景：直接返回对象类型
		c.logger.WithContext(ctx).Debug("[ConceptRetrieval#Retrieve] No relation types, returning object types only")
		return &SchemaRetrievalResult{
			ObjectTypes:   objectTypes,
			RelationTypes: []map[string]interface{}{},
		}, nil
	}

	// 构建Prompt
	prompt := c.buildPrompt(query, relationTypes, objectTypes, networkDetails)

	// 调用LLM
	selectedIndices, err := c.callLLM(ctx, prompt, accountID, accountType)
	if err != nil {
		c.logger.WithContext(ctx).Warnf("[ConceptRetrieval#Retrieve] LLM call failed: %v", err)
		// 降级：返回前topK个关系类型
		if len(relationTypes) > topK {
			relationTypes = relationTypes[:topK]
		}
		return &SchemaRetrievalResult{
			ObjectTypes:   c.filterObjectsByRelations(objectTypes, relationTypes),
			RelationTypes: relationTypes,
		}, nil
	}

	// 根据LLM返回的索引过滤关系类型
	filteredRelations := c.filterByIndices(relationTypes, selectedIndices, topK)

	// 根据过滤后的关系类型筛选对象类型
	filteredObjects := c.filterObjectsByRelations(objectTypes, filteredRelations)

	c.logger.WithContext(ctx).Infof("[ConceptRetrieval#Retrieve] Filtered %d relations, %d objects",
		len(filteredRelations), len(filteredObjects))

	return &SchemaRetrievalResult{
		ObjectTypes:   filteredObjects,
		RelationTypes: filteredRelations,
	}, nil
}

// extractRelationTypes 提取关系类型
func (c *ConceptRetrieval) extractRelationTypes(networkDetails map[string]interface{}) []map[string]interface{} {
	var result []map[string]interface{}
	if relations, ok := networkDetails["relation_types"].([]interface{}); ok {
		for _, r := range relations {
			if relMap, ok := r.(map[string]interface{}); ok {
				result = append(result, relMap)
			}
		}
	}
	return result
}

// extractObjectTypes 提取对象类型
func (c *ConceptRetrieval) extractObjectTypes(networkDetails map[string]interface{}) []map[string]interface{} {
	var result []map[string]interface{}
	if objects, ok := networkDetails["object_types"].([]interface{}); ok {
		for _, o := range objects {
			if objMap, ok := o.(map[string]interface{}); ok {
				result = append(result, objMap)
			}
		}
	}
	return result
}

// buildPrompt 构建Schema召回Prompt
func (c *ConceptRetrieval) buildPrompt(
	query string,
	relationTypes []map[string]interface{},
	objectTypes []map[string]interface{},
	networkDetails map[string]interface{},
) string {
	var prompt strings.Builder

	// 网络信息
	knID, _ := networkDetails["id"].(string)
	knName, _ := networkDetails["name"].(string)
	knComment, _ := networkDetails["comment"].(string)

	prompt.WriteString("当前知识网络信息:\n")
	prompt.WriteString(fmt.Sprintf("- ID: %s, 名称: %s, 描述: %s\n\n", knID, knName, knComment))

	// 构建对象ID到名称的映射
	objIDToName := make(map[string]string)
	for _, obj := range objectTypes {
		if id, ok := obj["id"].(string); ok {
			name, _ := obj["name"].(string)
			objIDToName[id] = name
		}
	}

	// 关系类型编号
	prompt.WriteString("关系类型编号:\n")
	for i, rel := range relationTypes {
		srcObjID, _ := rel["source_object_type_id"].(string)
		tgtObjID, _ := rel["target_object_type_id"].(string)
		relName, _ := rel["name"].(string)
		relComment, _ := rel["comment"].(string)

		srcName := objIDToName[srcObjID]
		if srcName == "" {
			srcName = "未知对象"
		}
		tgtName := objIDToName[tgtObjID]
		if tgtName == "" {
			tgtName = "未知对象"
		}

		if relComment != "" {
			prompt.WriteString(fmt.Sprintf("%d. %s -> %s（%s） -> %s\n", i+1, srcName, relName, relComment, tgtName))
		} else {
			prompt.WriteString(fmt.Sprintf("%d. %s -> %s -> %s\n", i+1, srcName, relName, tgtName))
		}
	}

	// 对象类型详细信息（增强部分）
	prompt.WriteString("\n对象类型详细信息:\n")
	for _, obj := range objectTypes {
		objID, _ := obj["id"].(string)
		objName, _ := obj["name"].(string)
		objComment, _ := obj["comment"].(string)

		if objComment != "" {
			prompt.WriteString(fmt.Sprintf("%s (id=%s, %s)\n", objName, objID, objComment))
		} else {
			prompt.WriteString(fmt.Sprintf("%s (id=%s)\n", objName, objID))
		}

		// 添加属性信息（最多5个）
		if dataProps, ok := obj["data_properties"].([]interface{}); ok && len(dataProps) > 0 {
			prompt.WriteString("   属性:\n")
			maxProps := 5
			if len(dataProps) < maxProps {
				maxProps = len(dataProps)
			}
			for j := 0; j < maxProps; j++ {
				if propMap, ok := dataProps[j].(map[string]interface{}); ok {
					propName, _ := propMap["name"].(string)
					propComment, _ := propMap["comment"].(string)
					if propComment != "" {
						prompt.WriteString(fmt.Sprintf("     - %s (%s)\n", propName, propComment))
					} else {
						prompt.WriteString(fmt.Sprintf("     - %s\n", propName))
					}
				}
			}
		}
	}

	// 指令
	prompt.WriteString("\n请根据以下问题，并结合对象类型信息和关系类型编号，判断问题涉及的子图信息，")
	prompt.WriteString("并只返回子图涉及的关系类型编号，不要解释，以列表形式返回，例如：[2, 1, 3]，")
	prompt.WriteString("都不相关则返回[]，如果不确定或只要一点点相关，就返回，不要遗漏任何关系类型\n\n")
	prompt.WriteString(fmt.Sprintf("问题: %s", query))

	return prompt.String()
}

// callLLM 调用LLM并解析返回的索引
func (c *ConceptRetrieval) callLLM(ctx context.Context, prompt, accountID, accountType string) ([]int, error) {
	messages := []interfaces.LLMMessage{
		{
			Role:    "system",
			Content: "你是一个智能关系类型筛选助手。你能够根据用户提出的问题，从关系类型列表中筛选出相关的类型，返回相关关系类型的编号，不确定的关系务必返回，要求很高的召回率，遗漏后果很严重。",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	req := &interfaces.LLMChatReq{
		Model:            c.config.Model,
		Messages:         messages,
		Temperature:      c.config.Temperature,
		TopK:             c.config.TopK,
		TopP:             c.config.TopP,
		FrequencyPenalty: c.config.FrequencyPenalty,
		PresencePenalty:  c.config.PresencePenalty,
		MaxTokens:        c.config.MaxTokens,
		Stream:           true,
		AccountID:        accountID,
		AccountType:      accountType,
	}

	content, err := c.mfModelClient.Chat(ctx, req)
	if err != nil {
		return nil, err
	}

	c.logger.WithContext(ctx).Debugf("[ConceptRetrieval#callLLM] LLM response: %s", content)

	return c.parseIndices(content)
}

// parseIndices 从LLM响应解析索引
func (c *ConceptRetrieval) parseIndices(content string) ([]int, error) {
	arrayRegex := regexp.MustCompile(`\[([0-9,\s]+)\]`)
	match := arrayRegex.FindStringSubmatch(content)
	if len(match) < minMatchLength {
		return []int{}, nil
	}

	indicesStr := match[1]
	if strings.TrimSpace(indicesStr) == "" {
		return []int{}, nil
	}

	var indices []int
	parts := strings.Split(indicesStr, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err == nil && n > 0 {
			indices = append(indices, n-1) // 转为0-based
		}
	}

	return indices, nil
}

// filterByIndices 根据索引过滤
func (c *ConceptRetrieval) filterByIndices(items []map[string]interface{}, indices []int, topK int) []map[string]interface{} {
	if len(indices) == 0 {
		// 如果LLM未选中任何，返回前topK个
		if len(items) > topK {
			return items[:topK]
		}
		return items
	}

	var result []map[string]interface{}
	seen := make(map[int]bool)
	for _, idx := range indices {
		if idx >= 0 && idx < len(items) && !seen[idx] {
			result = append(result, items[idx])
			seen[idx] = true
		}
	}

	if len(result) > topK {
		result = result[:topK]
	}

	return result
}

// filterObjectsByRelations 根据关系类型过滤对象类型
func (c *ConceptRetrieval) filterObjectsByRelations(objectTypes, relationTypes []map[string]interface{}) []map[string]interface{} {
	// 收集相关的对象类型ID
	relevantObjIDs := make(map[string]bool)
	for _, rel := range relationTypes {
		if srcID, ok := rel["source_object_type_id"].(string); ok {
			relevantObjIDs[srcID] = true
		}
		if tgtID, ok := rel["target_object_type_id"].(string); ok {
			relevantObjIDs[tgtID] = true
		}
	}

	// 如果没有相关ID，返回所有对象类型
	if len(relevantObjIDs) == 0 {
		return objectTypes
	}

	// 过滤
	var result []map[string]interface{}
	for _, obj := range objectTypes {
		if id, ok := obj["id"].(string); ok && relevantObjIDs[id] {
			result = append(result, obj)
		}
	}

	return result
}
