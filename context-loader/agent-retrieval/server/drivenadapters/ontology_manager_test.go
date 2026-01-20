// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package drivenadapters

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

// TestSearchObjectTypes_Success 测试 SearchObjectTypes 成功场景
func TestSearchObjectTypes_Success(t *testing.T) {
	Convey("TestSearchObjectTypes_Success", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()

		client := &ontologyManagerAccess{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/ontology-manager",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()
		req := &interfaces.QueryConceptsReq{
			KnID: "kn-001",
		}

		// Mock HTTP 成功响应
		mockHTTPClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(200, []byte(`{"object_types": []}`), nil)

		resp, err := client.SearchObjectTypes(ctx, req)
		So(err, ShouldBeNil)
		So(resp, ShouldNotBeNil)
	})
}

// TestSearchObjectTypes_HTTPError 测试 SearchObjectTypes HTTP 错误
func TestSearchObjectTypes_HTTPError(t *testing.T) {
	Convey("TestSearchObjectTypes_HTTPError", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

		client := &ontologyManagerAccess{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/ontology-manager",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()
		req := &interfaces.QueryConceptsReq{
			KnID: "kn-001",
		}

		// Mock HTTP 错误
		mockHTTPClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(0, nil, errors.New("connection refused"))

		_, err := client.SearchObjectTypes(ctx, req)
		So(err, ShouldNotBeNil)
	})
}

// TestSearchObjectTypes_NotFound 测试 SearchObjectTypes 404 错误
func TestSearchObjectTypes_NotFound(t *testing.T) {
	Convey("TestSearchObjectTypes_NotFound", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()

		client := &ontologyManagerAccess{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/ontology-manager",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()
		req := &interfaces.QueryConceptsReq{
			KnID: "kn-001",
		}

		// Mock 404 响应
		mockHTTPClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(404, nil, nil)

		_, err := client.SearchObjectTypes(ctx, req)
		So(err, ShouldNotBeNil)
	})
}

// TestGetObjectTypeDetail_Success 测试 GetObjectTypeDetail 成功场景
func TestGetObjectTypeDetail_Success(t *testing.T) {
	Convey("TestGetObjectTypeDetail_Success", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()

		client := &ontologyManagerAccess{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/ontology-manager",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()

		// Mock HTTP 成功响应
		mockHTTPClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(200, []byte(`{"entries": [{"id": "ot-001", "name": "测试对象类"}]}`), nil)

		resp, err := client.GetObjectTypeDetail(ctx, "kn-001", []string{"ot-001"}, true)
		So(err, ShouldBeNil)
		So(resp, ShouldNotBeNil)
		So(len(resp), ShouldEqual, 1)
	})
}

// TestGetObjectTypeDetail_HTTPError 测试 GetObjectTypeDetail HTTP 错误
func TestGetObjectTypeDetail_HTTPError(t *testing.T) {
	Convey("TestGetObjectTypeDetail_HTTPError", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

		client := &ontologyManagerAccess{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/ontology-manager",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()

		// Mock HTTP 错误
		mockHTTPClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(0, nil, errors.New("connection refused"))

		_, err := client.GetObjectTypeDetail(ctx, "kn-001", []string{"ot-001"}, true)
		So(err, ShouldNotBeNil)
	})
}

// TestSearchRelationTypes_Success 测试 SearchRelationTypes 成功场景
func TestSearchRelationTypes_Success(t *testing.T) {
	Convey("TestSearchRelationTypes_Success", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()

		client := &ontologyManagerAccess{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/ontology-manager",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()
		req := &interfaces.QueryConceptsReq{
			KnID: "kn-001",
		}

		// Mock HTTP 成功响应
		mockHTTPClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(200, []byte(`{"relation_types": []}`), nil)

		resp, err := client.SearchRelationTypes(ctx, req)
		So(err, ShouldBeNil)
		So(resp, ShouldNotBeNil)
	})
}

// TestSearchRelationTypes_HTTPError 测试 SearchRelationTypes HTTP 错误
func TestSearchRelationTypes_HTTPError(t *testing.T) {
	Convey("TestSearchRelationTypes_HTTPError", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

		client := &ontologyManagerAccess{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/ontology-manager",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()
		req := &interfaces.QueryConceptsReq{
			KnID: "kn-001",
		}

		// Mock HTTP 错误
		mockHTTPClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(0, nil, errors.New("connection refused"))

		_, err := client.SearchRelationTypes(ctx, req)
		So(err, ShouldNotBeNil)
	})
}

// TestSearchActionTypes_Success 测试 SearchActionTypes 成功场景
func TestSearchActionTypes_Success(t *testing.T) {
	Convey("TestSearchActionTypes_Success", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()

		client := &ontologyManagerAccess{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/ontology-manager",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()
		req := &interfaces.QueryConceptsReq{
			KnID: "kn-001",
		}

		// Mock HTTP 成功响应
		mockHTTPClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(200, []byte(`{"action_types": []}`), nil)

		resp, err := client.SearchActionTypes(ctx, req)
		So(err, ShouldBeNil)
		So(resp, ShouldNotBeNil)
	})
}

// TestGetActionTypeDetail_Success 测试 GetActionTypeDetail 成功场景
func TestGetActionTypeDetail_Success(t *testing.T) {
	Convey("TestGetActionTypeDetail_Success", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()

		client := &ontologyManagerAccess{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/ontology-manager",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()

		// Mock HTTP 成功响应
		mockHTTPClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(200, []byte(`[{"id": "at-001", "name": "测试行动类"}]`), nil)

		resp, err := client.GetActionTypeDetail(ctx, "kn-001", []string{"at-001"}, true)
		So(err, ShouldBeNil)
		So(resp, ShouldNotBeNil)
		So(len(resp), ShouldEqual, 1)
	})
}

// TestGetActionTypeDetail_HTTPError 测试 GetActionTypeDetail HTTP 错误
func TestGetActionTypeDetail_HTTPError(t *testing.T) {
	Convey("TestGetActionTypeDetail_HTTPError", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

		client := &ontologyManagerAccess{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/ontology-manager",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()

		// Mock HTTP 错误
		mockHTTPClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(0, nil, errors.New("connection refused"))

		_, err := client.GetActionTypeDetail(ctx, "kn-001", []string{"at-001"}, true)
		So(err, ShouldNotBeNil)
	})
}
