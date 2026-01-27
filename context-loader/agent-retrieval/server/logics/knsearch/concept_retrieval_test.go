// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knsearch

import (
	"context"
	"testing"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

// mockLogger 测试用的mock logger
type mockLogger struct{}

func (m *mockLogger) Debug(args ...interface{})                                  {}
func (m *mockLogger) Debugf(format string, args ...interface{})                  {}
func (m *mockLogger) Info(args ...interface{})                                   {}
func (m *mockLogger) Infof(format string, args ...interface{})                   {}
func (m *mockLogger) Warn(args ...interface{})                                   {}
func (m *mockLogger) Warnf(format string, args ...interface{})                   {}
func (m *mockLogger) Error(args ...interface{})                                  {}
func (m *mockLogger) Errorf(format string, args ...interface{})                  {}
func (m *mockLogger) WithContext(ctx context.Context) interfaces.Logger          { return m }
func (m *mockLogger) WithField(key string, value interface{}) interfaces.Logger  { return m }
func (m *mockLogger) WithFields(fields map[string]interface{}) interfaces.Logger { return m }

func TestConceptRetrieval_ParseIndices(t *testing.T) {
	cr := &ConceptRetrieval{logger: &mockLogger{}}

	tests := []struct {
		name     string
		content  string
		expected []int
	}{
		{
			name:     "SCH-IDX-001: JSON数组格式",
			content:  "[1, 3, 5]",
			expected: []int{0, 2, 4},
		},
		{
			name:     "SCH-IDX-002: 带前置文本",
			content:  "相关的关系类型编号是：[2, 4]",
			expected: []int{1, 3},
		},
		{
			name:     "SCH-IDX-003: 空数组",
			content:  "[]",
			expected: []int{},
		},
		{
			name:     "SCH-IDX-004: 无效索引(0)",
			content:  "[0, 2, 3]",
			expected: []int{1, 2}, // 0无效（必须>0），2和3转为0-based后为1和2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			indices, err := cr.parseIndices(tt.content)
			if err != nil {
				t.Fatalf("parseIndices failed: %v", err)
			}

			if len(indices) != len(tt.expected) {
				t.Errorf("Expected %d indices, got %d", len(tt.expected), len(indices))
				return
			}

			for i, idx := range indices {
				if idx != tt.expected[i] {
					t.Errorf("Index %d: expected %d, got %d", i, tt.expected[i], idx)
				}
			}
		})
	}
}

func TestConceptRetrieval_ExtractRelationTypes(t *testing.T) {
	cr := &ConceptRetrieval{logger: &mockLogger{}}

	tests := []struct {
		name           string
		networkDetails map[string]interface{}
		expectedCount  int
	}{
		{
			name: "SCH-EXT-001: 正常提取",
			networkDetails: map[string]interface{}{
				"relation_types": []interface{}{
					map[string]interface{}{"id": "rel1", "name": "关系1"},
					map[string]interface{}{"id": "rel2", "name": "关系2"},
				},
			},
			expectedCount: 2,
		},
		{
			name:           "SCH-EXT-002: 空详情",
			networkDetails: map[string]interface{}{},
			expectedCount:  0,
		},
		{
			name: "SCH-EXT-003: 空关系类型",
			networkDetails: map[string]interface{}{
				"relation_types": []interface{}{},
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cr.extractRelationTypes(tt.networkDetails)
			if len(result) != tt.expectedCount {
				t.Errorf("Expected %d relation types, got %d", tt.expectedCount, len(result))
			}
		})
	}
}

func TestConceptRetrieval_FilterByIndices(t *testing.T) {
	cr := &ConceptRetrieval{logger: &mockLogger{}}

	items := []map[string]interface{}{
		{"id": "item0"},
		{"id": "item1"},
		{"id": "item2"},
		{"id": "item3"},
		{"id": "item4"},
	}

	tests := []struct {
		name        string
		indices     []int
		topK        int
		expectedIDs []string
	}{
		{
			name:        "SCH-FLT-001: 正常过滤",
			indices:     []int{0, 2, 4},
			topK:        10,
			expectedIDs: []string{"item0", "item2", "item4"},
		},
		{
			name:        "SCH-FLT-002: 空索引返回前topK",
			indices:     []int{},
			topK:        2,
			expectedIDs: []string{"item0", "item1"},
		},
		{
			name:        "SCH-FLT-003: topK限制",
			indices:     []int{0, 1, 2, 3, 4},
			topK:        2,
			expectedIDs: []string{"item0", "item1"},
		},
		{
			name:        "SCH-FLT-004: 索引越界忽略",
			indices:     []int{0, 100, 2},
			topK:        10,
			expectedIDs: []string{"item0", "item2"},
		},
		{
			name:        "SCH-FLT-005: 重复索引去重",
			indices:     []int{0, 0, 1, 1},
			topK:        10,
			expectedIDs: []string{"item0", "item1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cr.filterByIndices(items, tt.indices, tt.topK)
			if len(result) != len(tt.expectedIDs) {
				t.Errorf("Expected %d items, got %d", len(tt.expectedIDs), len(result))
				return
			}
			for i, item := range result {
				if item["id"] != tt.expectedIDs[i] {
					t.Errorf("Item %d: expected %s, got %s", i, tt.expectedIDs[i], item["id"])
				}
			}
		})
	}
}

func TestConceptRetrieval_FilterObjectsByRelations(t *testing.T) {
	cr := &ConceptRetrieval{logger: &mockLogger{}}

	objectTypes := []map[string]interface{}{
		{"id": "obj1"},
		{"id": "obj2"},
		{"id": "obj3"},
	}

	tests := []struct {
		name          string
		relationTypes []map[string]interface{}
		expectedCount int
	}{
		{
			name: "SCH-OBJ-001: 正常关联",
			relationTypes: []map[string]interface{}{
				{"source_object_type_id": "obj1", "target_object_type_id": "obj2"},
			},
			expectedCount: 2, // obj1和obj2
		},
		{
			name:          "SCH-OBJ-002: 无关系返回全部",
			relationTypes: []map[string]interface{}{},
			expectedCount: 3,
		},
		{
			name: "SCH-OBJ-003: 不存在的对象ID",
			relationTypes: []map[string]interface{}{
				{"source_object_type_id": "unknown", "target_object_type_id": "obj1"},
			},
			expectedCount: 1, // 只有obj1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cr.filterObjectsByRelations(objectTypes, tt.relationTypes)
			if len(result) != tt.expectedCount {
				t.Errorf("Expected %d objects, got %d", tt.expectedCount, len(result))
			}
		})
	}
}
