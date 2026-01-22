// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knretrieval

import (
	"testing"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
	. "github.com/smartystreets/goconvey/convey"
)

// TestDeduplicateConcepts 测试 deduplicateConcepts 函数
func TestDeduplicateConcepts(t *testing.T) {
	Convey("TestDeduplicateConcepts", t, func() {
		service := &knRetrievalServiceImpl{}

		Convey("去重重复概念", func() {
			concepts := []*interfaces.ConceptResult{
				{ConceptType: interfaces.KnConceptTypeObject, ConceptID: "obj-001"},
				{ConceptType: interfaces.KnConceptTypeObject, ConceptID: "obj-001"}, // 重复
				{ConceptType: interfaces.KnConceptTypeObject, ConceptID: "obj-002"},
				{ConceptType: interfaces.KnConceptTypeRelation, ConceptID: "rel-001"},
				{ConceptType: interfaces.KnConceptTypeRelation, ConceptID: "rel-001"}, // 重复
			}

			result := service.deduplicateConcepts(concepts)
			So(len(result), ShouldEqual, 3)
		})

		Convey("相同 ID 不同类型不去重", func() {
			concepts := []*interfaces.ConceptResult{
				{ConceptType: interfaces.KnConceptTypeObject, ConceptID: "id-001"},
				{ConceptType: interfaces.KnConceptTypeAction, ConceptID: "id-001"},
			}

			result := service.deduplicateConcepts(concepts)
			So(len(result), ShouldEqual, 2)
		})

		Convey("空数组", func() {
			result := service.deduplicateConcepts([]*interfaces.ConceptResult{})
			So(len(result), ShouldEqual, 0)
		})

		Convey("无重复", func() {
			concepts := []*interfaces.ConceptResult{
				{ConceptType: interfaces.KnConceptTypeObject, ConceptID: "obj-001"},
				{ConceptType: interfaces.KnConceptTypeObject, ConceptID: "obj-002"},
			}

			result := service.deduplicateConcepts(concepts)
			So(len(result), ShouldEqual, 2)
		})
	})
}

// TestFilterQueryStrategysBySearchScope 测试 filterQueryStrategysBySearchScope 函数
func TestFilterQueryStrategysBySearchScope(t *testing.T) {
	Convey("TestFilterQueryStrategysBySearchScope", t, func() {
		service := &knRetrievalServiceImpl{}

		Convey("包含所有概念类型", func() {
			includeAll := true
			searchScope := &interfaces.SearchScopeConfig{
				IncludeObjectTypes:   &includeAll,
				IncludeRelationTypes: &includeAll,
				IncludeActionTypes:   &includeAll,
			}

			strategies := []*interfaces.SemanticQueryStrategy{
				{Filter: &interfaces.QueryStrategyFilter{ConceptType: interfaces.KnConceptTypeObject}},
				{Filter: &interfaces.QueryStrategyFilter{ConceptType: interfaces.KnConceptTypeRelation}},
				{Filter: &interfaces.QueryStrategyFilter{ConceptType: interfaces.KnConceptTypeAction}},
			}

			result := service.filterQueryStrategysBySearchScope(strategies, searchScope)
			So(len(result), ShouldEqual, 3)
		})

		Convey("排除对象类型", func() {
			includeTrue := true
			includeFalse := false
			searchScope := &interfaces.SearchScopeConfig{
				IncludeObjectTypes:   &includeFalse,
				IncludeRelationTypes: &includeTrue,
				IncludeActionTypes:   &includeTrue,
			}

			strategies := []*interfaces.SemanticQueryStrategy{
				{Filter: &interfaces.QueryStrategyFilter{ConceptType: interfaces.KnConceptTypeObject}},
				{Filter: &interfaces.QueryStrategyFilter{ConceptType: interfaces.KnConceptTypeRelation}},
				{Filter: &interfaces.QueryStrategyFilter{ConceptType: interfaces.KnConceptTypeAction}},
			}

			result := service.filterQueryStrategysBySearchScope(strategies, searchScope)
			So(len(result), ShouldEqual, 2)
			So(result[0].Filter.ConceptType, ShouldEqual, interfaces.KnConceptTypeRelation)
		})

		Convey("仅包含行动类型", func() {
			includeTrue := true
			includeFalse := false
			searchScope := &interfaces.SearchScopeConfig{
				IncludeObjectTypes:   &includeFalse,
				IncludeRelationTypes: &includeFalse,
				IncludeActionTypes:   &includeTrue,
			}

			strategies := []*interfaces.SemanticQueryStrategy{
				{Filter: &interfaces.QueryStrategyFilter{ConceptType: interfaces.KnConceptTypeObject}},
				{Filter: &interfaces.QueryStrategyFilter{ConceptType: interfaces.KnConceptTypeRelation}},
				{Filter: &interfaces.QueryStrategyFilter{ConceptType: interfaces.KnConceptTypeAction}},
			}

			result := service.filterQueryStrategysBySearchScope(strategies, searchScope)
			So(len(result), ShouldEqual, 1)
			So(result[0].Filter.ConceptType, ShouldEqual, interfaces.KnConceptTypeAction)
		})

		Convey("策略无 Filter", func() {
			includeTrue := true
			searchScope := &interfaces.SearchScopeConfig{
				IncludeObjectTypes:   &includeTrue,
				IncludeRelationTypes: &includeTrue,
				IncludeActionTypes:   &includeTrue,
			}

			strategies := []*interfaces.SemanticQueryStrategy{
				{Filter: nil}, // 无 Filter 的策略应该被保留
			}

			result := service.filterQueryStrategysBySearchScope(strategies, searchScope)
			So(len(result), ShouldEqual, 1)
		})

		Convey("空策略数组", func() {
			includeTrue := true
			searchScope := &interfaces.SearchScopeConfig{
				IncludeObjectTypes:   &includeTrue,
				IncludeRelationTypes: &includeTrue,
				IncludeActionTypes:   &includeTrue,
			}

			result := service.filterQueryStrategysBySearchScope([]*interfaces.SemanticQueryStrategy{}, searchScope)
			So(len(result), ShouldEqual, 0)
		})
	})
}
