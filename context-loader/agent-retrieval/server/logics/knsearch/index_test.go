// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knsearch

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

// TestKnSearch_Success 测试 KnSearch 成功场景
func TestKnSearch_Success(t *testing.T) {
	Convey("TestKnSearch_Success", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockDataRetrieval := mocks.NewMockDataRetrieval(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()

		service := &knSearchService{
			Logger:        mockLogger,
			DataRetrieval: mockDataRetrieval,
		}

		ctx := context.Background()
		req := &interfaces.KnSearchReq{
			Query: "测试查询",
			KnID:  "kn-001",
		}

		// Mock DataRetrieval 成功响应
		mockDataRetrieval.EXPECT().KnSearch(gomock.Any(), gomock.Any()).
			Return(&interfaces.KnSearchResp{
				ObjectTypes: []interface{}{},
				Nodes:       []interface{}{},
			}, nil)

		resp, err := service.KnSearch(ctx, req)
		So(err, ShouldBeNil)
		So(resp, ShouldNotBeNil)
	})
}

// TestKnSearch_Error 测试 KnSearch 错误场景
func TestKnSearch_Error(t *testing.T) {
	Convey("TestKnSearch_Error", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockDataRetrieval := mocks.NewMockDataRetrieval(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()

		service := &knSearchService{
			Logger:        mockLogger,
			DataRetrieval: mockDataRetrieval,
		}

		ctx := context.Background()
		req := &interfaces.KnSearchReq{
			Query: "测试查询",
			KnID:  "kn-001",
		}

		// Mock DataRetrieval 错误
		mockDataRetrieval.EXPECT().KnSearch(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("data retrieval error"))

		_, err := service.KnSearch(ctx, req)
		So(err, ShouldNotBeNil)
	})
}

// TestKnSearch_KnIDConversion 测试 KnID 转换逻辑
func TestKnSearch_KnIDConversion(t *testing.T) {
	Convey("TestKnSearch_KnIDConversion", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockDataRetrieval := mocks.NewMockDataRetrieval(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()

		service := &knSearchService{
			Logger:        mockLogger,
			DataRetrieval: mockDataRetrieval,
		}

		ctx := context.Background()
		req := &interfaces.KnSearchReq{
			Query: "测试查询",
			KnID:  "kn-001",
		}

		// 验证 KnID 被正确转换为 knIDs 数组
		mockDataRetrieval.EXPECT().KnSearch(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, r *interfaces.KnSearchReq) (*interfaces.KnSearchResp, error) {
				// 检查 knIDs 被正确设置
				knIDs := r.GetKnIDs()
				So(len(knIDs), ShouldEqual, 1)
				So(knIDs[0].KnowledgeNetworkID, ShouldEqual, "kn-001")
				return &interfaces.KnSearchResp{}, nil
			})

		_, err := service.KnSearch(ctx, req)
		So(err, ShouldBeNil)
	})
}
