// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knsearch

import (
	"context"
	"sync"
	"testing"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

// mockMFModelClient 测试用的统一MF-Model-API客户端
type mockMFModelClient struct {
	chatResponse string
	chatError    error
	rerankResp   *interfaces.RerankResp
	rerankError  error
}

func (m *mockMFModelClient) Chat(ctx context.Context, req *interfaces.LLMChatReq) (string, error) {
	return m.chatResponse, m.chatError
}

func (m *mockMFModelClient) Rerank(ctx context.Context, query string, documents []string) (*interfaces.RerankResp, error) {
	return m.rerankResp, m.rerankError
}

// mockOntologyQuery 测试用的mock OntologyQuery
type mockOntologyQuery struct {
	response *interfaces.QueryObjectInstancesResp
	err      error
}

func (m *mockOntologyQuery) QueryObjectInstances(ctx context.Context, req *interfaces.QueryObjectInstancesReq) (*interfaces.QueryObjectInstancesResp, error) {
	return m.response, m.err
}

func (m *mockOntologyQuery) QueryLogicProperties(ctx context.Context, req *interfaces.QueryLogicPropertiesReq) (*interfaces.QueryLogicPropertiesResp, error) {
	return nil, nil
}

func (m *mockOntologyQuery) QueryActions(ctx context.Context, req *interfaces.QueryActionsRequest) (*interfaces.QueryActionsResponse, error) {
	return nil, nil
}

func (m *mockOntologyQuery) QueryInstanceSubgraph(ctx context.Context, req *interfaces.QueryInstanceSubgraphReq) (*interfaces.QueryInstanceSubgraphResp, error) {
	return nil, nil
}

func TestSemanticRetrieval_Tokenize(t *testing.T) {
	sr := &SemanticRetrieval{logger: &mockLogger{}}

	tests := []struct {
		name          string
		query         string
		expectedCount int
	}{
		{
			name:          "SEM-TOK-001: 正常分词",
			query:         "北京 餐饮 连锁",
			expectedCount: 3,
		},
		{
			name:          "SEM-TOK-002: 过滤单字符",
			query:         "北京 a 餐饮",
			expectedCount: 2, // "a"被过滤
		},
		{
			name:          "SEM-TOK-003: 空查询",
			query:         "",
			expectedCount: 0,
		},
		{
			name:          "SEM-TOK-004: 多空格处理",
			query:         "北京   餐饮",
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sr.tokenize(tt.query)
			if len(result) != tt.expectedCount {
				t.Errorf("Expected %d keywords, got %d: %v", tt.expectedCount, len(result), result)
			}
		})
	}
}

func TestSemanticRetrieval_MaxKeywords(t *testing.T) {
	sr := &SemanticRetrieval{
		logger:      &mockLogger{},
		maxKeywords: 5,
	}

	tests := []struct {
		name          string
		query         string
		expectedCount int
	}{
		{
			name:          "SEM-MAX-001: 不超限",
			query:         "a1 b2 c3",
			expectedCount: 3,
		},
		{
			name:          "SEM-MAX-002: 刚好5个",
			query:         "a1 b2 c3 d4 e5",
			expectedCount: 5,
		},
		{
			name:          "SEM-MAX-003: 超过5个取前5",
			query:         "a1 b2 c3 d4 e5 f6 g7",
			expectedCount: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keywords := sr.tokenize(tt.query)
			if len(keywords) > sr.maxKeywords {
				keywords = keywords[:sr.maxKeywords]
			}
			if len(keywords) != tt.expectedCount {
				t.Errorf("Expected %d keywords after limit, got %d", tt.expectedCount, len(keywords))
			}
		})
	}
}

func TestSemanticRetrieval_Retrieve_EmptyObjectTypes(t *testing.T) {
	sr := &SemanticRetrieval{
		logger:        &mockLogger{},
		mfModelClient: &mockMFModelClient{},
		ontologyQuery: &mockOntologyQuery{},
		maxKeywords:   5,
	}

	result, err := sr.Retrieve(context.Background(), "test-kn-1", "测试查询", []map[string]interface{}{}, 10)
	if err != nil {
		t.Fatalf("Retrieve failed: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 results for empty object types, got %d", len(result))
	}
}

func TestSemanticRetrieval_SingleKeyword(t *testing.T) {
	mockQuery := &mockOntologyQuery{
		response: &interfaces.QueryObjectInstancesResp{
			Data: []interface{}{
				map[string]interface{}{"id": "inst1", "display_name": "实例1"},
				map[string]interface{}{"id": "inst2", "display_name": "实例2"},
			},
		},
	}

	sr := &SemanticRetrieval{
		logger:        &mockLogger{},
		mfModelClient: &mockMFModelClient{},
		ontologyQuery: mockQuery,
		maxKeywords:   5,
	}

	objectTypes := []map[string]interface{}{
		{"id": "obj1"},
	}

	result, err := sr.Retrieve(context.Background(), "test-kn-1", "北京", objectTypes, 10)
	if err != nil {
		t.Fatalf("Retrieve failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 results, got %d", len(result))
	}
}

func TestSemanticRetrieval_MultiKeyword_Dedup(t *testing.T) {
	callCount := 0
	mockQuery := &mockOntologyQuery{
		response: &interfaces.QueryObjectInstancesResp{
			Data: []interface{}{
				// 每次返回相同的实例，模拟多关键词命中同一实例
				map[string]interface{}{"id": "inst1", "display_name": "实例1"},
			},
		},
	}

	// 创建一个自定义的mock，增加调用计数
	customQuery := &countingMockOntologyQuery{
		response:  mockQuery.response,
		callCount: &callCount,
	}

	mockRerank := &mockMFModelClient{
		rerankResp: &interfaces.RerankResp{
			Results: []interfaces.RerankResult{
				{Index: 0, RelevanceScore: 0.9},
			},
		},
	}

	sr := &SemanticRetrieval{
		logger:        &mockLogger{},
		mfModelClient: mockRerank,
		ontologyQuery: customQuery,
		maxKeywords:   5,
	}

	objectTypes := []map[string]interface{}{
		{"id": "obj1"},
	}

	result, err := sr.Retrieve(context.Background(), "test-kn-1", "北京 餐饮 连锁", objectTypes, 10)
	if err != nil {
		t.Fatalf("Retrieve failed: %v", err)
	}

	// 验证去重：虽然3个关键词，但都命中同一实例，最终只有1个
	if len(result) != 1 {
		t.Errorf("Expected 1 result after dedup, got %d", len(result))
	}
}

// countingMockOntologyQuery 用于计数调用次数的mock
type countingMockOntologyQuery struct {
	response  *interfaces.QueryObjectInstancesResp
	callCount *int
	mu        sync.Mutex // 保护 callCount
}

func (m *countingMockOntologyQuery) QueryObjectInstances(ctx context.Context, req *interfaces.QueryObjectInstancesReq) (*interfaces.QueryObjectInstancesResp, error) {
	m.mu.Lock()
	*m.callCount++
	m.mu.Unlock()
	// 返回深拷贝的响应，避免并发修改
	resp := &interfaces.QueryObjectInstancesResp{
		Data: make([]interface{}, len(m.response.Data)),
	}
	for i, item := range m.response.Data {
		if origMap, ok := item.(map[string]interface{}); ok {
			newMap := make(map[string]interface{})
			for k, v := range origMap {
				newMap[k] = v
			}
			resp.Data[i] = newMap
		}
	}
	return resp, nil
}

func (m *countingMockOntologyQuery) QueryLogicProperties(ctx context.Context, req *interfaces.QueryLogicPropertiesReq) (*interfaces.QueryLogicPropertiesResp, error) {
	return nil, nil
}

func (m *countingMockOntologyQuery) QueryActions(ctx context.Context, req *interfaces.QueryActionsRequest) (*interfaces.QueryActionsResponse, error) {
	return nil, nil
}

func (m *countingMockOntologyQuery) QueryInstanceSubgraph(ctx context.Context, req *interfaces.QueryInstanceSubgraphReq) (*interfaces.QueryInstanceSubgraphResp, error) {
	return nil, nil
}
