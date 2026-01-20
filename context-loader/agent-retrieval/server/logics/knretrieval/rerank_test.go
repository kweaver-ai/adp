// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knretrieval

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

// TestFilterRerankScoreZero 测试 filterRerankScoreZero 函数
func TestFilterRerankScoreZero(t *testing.T) {
	Convey("TestFilterRerankScoreZero", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)

		service := &knRetrievalServiceImpl{
			logger: mockLogger,
		}

		Convey("过滤掉 RerankScore 为 0 的结果", func() {
			concepts := []*interfaces.ConceptResult{
				{ConceptID: "1", ConceptName: "Concept1", RerankScore: 0.8},
				{ConceptID: "2", ConceptName: "Concept2", RerankScore: 0},
				{ConceptID: "3", ConceptName: "Concept3", RerankScore: 0.5},
				{ConceptID: "4", ConceptName: "Concept4", RerankScore: 0},
			}

			result := service.filterRerankScoreZero(concepts)
			So(len(result), ShouldEqual, 2)
			So(result[0].ConceptID, ShouldEqual, "1")
			So(result[1].ConceptID, ShouldEqual, "3")
		})

		Convey("全部为 0 时返回空", func() {
			concepts := []*interfaces.ConceptResult{
				{ConceptID: "1", RerankScore: 0},
				{ConceptID: "2", RerankScore: 0},
			}

			result := service.filterRerankScoreZero(concepts)
			So(len(result), ShouldEqual, 0)
		})

		Convey("空输入返回空", func() {
			result := service.filterRerankScoreZero(nil)
			So(result, ShouldBeNil)
		})
	})
}

// TestRerankByDataRetrieval_DefaultAction 测试 rerankByDataRetrieval default action 场景
func TestRerankByDataRetrieval_DefaultAction(t *testing.T) {
	Convey("TestRerankByDataRetrieval_DefaultAction", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockDataRetrieval := mocks.NewMockDataRetrieval(ctrl)

		service := &knRetrievalServiceImpl{
			logger:        mockLogger,
			dataRetrieval: mockDataRetrieval,
		}

		ctx := context.Background()
		queryUnderstanding := &interfaces.QueryUnderstanding{
			OriginQuery: "测试查询",
		}

		concepts := []*interfaces.ConceptResult{
			{ConceptID: "1", ConceptName: "Concept1", RerankScore: 0.8},
			{ConceptID: "2", ConceptName: "Concept2", RerankScore: 0.5},
		}

		// default action 不调用 KnowledgeRerank
		result, err := service.rerankByDataRetrieval(ctx, queryUnderstanding, concepts, interfaces.KnowledgeRerankActionDefault, 10)
		So(err, ShouldBeNil)
		So(len(result), ShouldEqual, 2)
	})
}

// TestRerankByDataRetrieval_VectorAction 测试 rerankByDataRetrieval vector action 场景
func TestRerankByDataRetrieval_VectorAction(t *testing.T) {
	Convey("TestRerankByDataRetrieval_VectorAction", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockDataRetrieval := mocks.NewMockDataRetrieval(ctrl)

		service := &knRetrievalServiceImpl{
			logger:        mockLogger,
			dataRetrieval: mockDataRetrieval,
		}

		ctx := context.Background()
		queryUnderstanding := &interfaces.QueryUnderstanding{
			OriginQuery: "测试查询",
		}

		concepts := []*interfaces.ConceptResult{
			{ConceptID: "1", ConceptName: "Concept1", RerankScore: 0.8},
		}

		// Mock KnowledgeRerank 调用
		mockDataRetrieval.EXPECT().KnowledgeRerank(gomock.Any(), gomock.Any()).
			Return([]*interfaces.ConceptResult{
				{ConceptID: "1", ConceptName: "Concept1", RerankScore: 0.9},
			}, nil)

		result, err := service.rerankByDataRetrieval(ctx, queryUnderstanding, concepts, interfaces.KnowledgeRerankActionVector, 10)
		So(err, ShouldBeNil)
		So(len(result), ShouldEqual, 1)
	})
}

// TestRerankByDataRetrieval_Error 测试 rerankByDataRetrieval 错误场景
func TestRerankByDataRetrieval_Error(t *testing.T) {
	Convey("TestRerankByDataRetrieval_Error", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockDataRetrieval := mocks.NewMockDataRetrieval(ctrl)

		service := &knRetrievalServiceImpl{
			logger:        mockLogger,
			dataRetrieval: mockDataRetrieval,
		}

		ctx := context.Background()
		queryUnderstanding := &interfaces.QueryUnderstanding{
			OriginQuery: "测试查询",
		}

		concepts := []*interfaces.ConceptResult{
			{ConceptID: "1", ConceptName: "Concept1"},
		}

		// Mock KnowledgeRerank 错误
		mockDataRetrieval.EXPECT().KnowledgeRerank(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("rerank failed"))

		_, err := service.rerankByDataRetrieval(ctx, queryUnderstanding, concepts, interfaces.KnowledgeRerankActionVector, 10)
		So(err, ShouldNotBeNil)
	})
}

// TestRerankByDataRetrieval_WithLimit 测试 rerankByDataRetrieval 分页限制
func TestRerankByDataRetrieval_WithLimit(t *testing.T) {
	Convey("TestRerankByDataRetrieval_WithLimit", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockDataRetrieval := mocks.NewMockDataRetrieval(ctrl)

		service := &knRetrievalServiceImpl{
			logger:        mockLogger,
			dataRetrieval: mockDataRetrieval,
		}

		ctx := context.Background()
		queryUnderstanding := &interfaces.QueryUnderstanding{
			OriginQuery: "测试查询",
		}

		concepts := []*interfaces.ConceptResult{
			{ConceptID: "1", RerankScore: 0.9},
			{ConceptID: "2", RerankScore: 0.8},
			{ConceptID: "3", RerankScore: 0.7},
			{ConceptID: "4", RerankScore: 0.6},
			{ConceptID: "5", RerankScore: 0.5},
		}

		// limit=2 只返回前 2 个
		result, err := service.rerankByDataRetrieval(ctx, queryUnderstanding, concepts, interfaces.KnowledgeRerankActionDefault, 2)
		So(err, ShouldBeNil)
		So(len(result), ShouldEqual, 2)
	})
}
