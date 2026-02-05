// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package drivenadapters

import (
	"context"
	"errors"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/mocks"
)

// TestGetAppInfo_Success 测试 GetAppInfo 成功场景
func TestGetAppInfo_Success(t *testing.T) {
	convey.Convey("TestGetAppInfo_Success", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()

		client := &userManagementClient{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/user-management",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()

		// Mock HTTP 成功响应
		mockHTTPClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(200, map[string]interface{}{
				"id":   "app-001",
				"name": "测试应用",
			}, nil)

		resp, err := client.GetAppInfo(ctx, "app-001")
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp, convey.ShouldNotBeNil)
	})
}

// TestGetAppInfo_HTTPError 测试 GetAppInfo HTTP 错误
func TestGetAppInfo_HTTPError(t *testing.T) {
	convey.Convey("TestGetAppInfo_HTTPError", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

		client := &userManagementClient{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/user-management",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()

		// Mock HTTP 错误
		mockHTTPClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(0, nil, errors.New("connection refused"))

		_, err := client.GetAppInfo(ctx, "app-001")
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// TestGetUsersInfo_Success 测试 GetUsersInfo 成功场景
func TestGetUsersInfo_Success(t *testing.T) {
	convey.Convey("TestGetUsersInfo_Success", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()

		client := &userManagementClient{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/user-management",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()

		// Mock HTTP 成功响应
		mockHTTPClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(200, []interface{}{
				map[string]interface{}{
					"id":   "user-001",
					"name": "用户1",
				},
				map[string]interface{}{
					"id":   "user-002",
					"name": "用户2",
				},
			}, nil)

		resp, err := client.GetUsersInfo(ctx, []string{"user-001", "user-002"}, []string{"name"})
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp, convey.ShouldNotBeNil)
	})
}

// TestRemoveUserIDs 测试 removeUserIDs 函数
func TestRemoveUserIDs(t *testing.T) {
	convey.Convey("TestRemoveUserIDs", t, func() {
		client := &userManagementClient{}

		convey.Convey("移除部分元素", func() {
			source := []string{"a", "b", "c", "d", "e"}
			toRemove := []string{"b", "d"}
			result := client.removeUserIDs(source, toRemove)
			convey.So(len(result), convey.ShouldEqual, 3)
			convey.So(result, convey.ShouldContain, "a")
			convey.So(result, convey.ShouldContain, "c")
			convey.So(result, convey.ShouldContain, "e")
			convey.So(result, convey.ShouldNotContain, "b")
			convey.So(result, convey.ShouldNotContain, "d")
		})

		convey.Convey("移除全部元素", func() {
			source := []string{"a", "b"}
			toRemove := []string{"a", "b"}
			result := client.removeUserIDs(source, toRemove)
			convey.So(len(result), convey.ShouldEqual, 0)
		})

		convey.Convey("移除不存在的元素", func() {
			source := []string{"a", "b", "c"}
			toRemove := []string{"x", "y"}
			result := client.removeUserIDs(source, toRemove)
			convey.So(len(result), convey.ShouldEqual, 3)
		})

		convey.Convey("空数组", func() {
			source := []string{}
			toRemove := []string{"a"}
			result := client.removeUserIDs(source, toRemove)
			convey.So(len(result), convey.ShouldEqual, 0)
		})
	})
}

// TestGetUserInfo_Success 测试 GetUserInfo 成功场景
func TestGetUserInfo_Success(t *testing.T) {
	convey.Convey("TestGetUserInfo_Success", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()

		client := &userManagementClient{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/user-management",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()

		// Mock HTTP 成功响应
		mockHTTPClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(200, []interface{}{
				map[string]interface{}{
					"id":   "user-001",
					"name": "用户1",
				},
			}, nil)

		resp, err := client.GetUserInfo(ctx, "user-001")
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp, convey.ShouldNotBeNil)
	})
}

// TestGetUserInfo_NotFound 测试 GetUserInfo 用户不存在
func TestGetUserInfo_NotFound(t *testing.T) {
	convey.Convey("TestGetUserInfo_NotFound", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

		client := &userManagementClient{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/user-management",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()

		// Mock HTTP 成功响应但返回空数组
		mockHTTPClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(200, []interface{}{}, nil)

		_, err := client.GetUserInfo(ctx, "nonexistent-user")
		convey.So(err, convey.ShouldNotBeNil)
	})
}
