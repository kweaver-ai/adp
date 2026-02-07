// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package common

import (
	"testing"

	"vega-backend/tests/at/setup"
	"vega-backend/tests/testutil"
)

// BaseTestSuite 基础测试套件
// 提供通用的测试环境初始化和HTTP客户端
type BaseTestSuite struct {
	Config *setup.TestConfig
	Client *testutil.HTTPClient
	T      *testing.T
}

// NewBaseTestSuite 创建基础测试套件
func NewBaseTestSuite(t *testing.T) (*BaseTestSuite, error) {
	// 加载测试配置
	config, err := setup.LoadTestConfig()
	if err != nil {
		return nil, err
	}

	// 创建HTTP客户端
	client := testutil.NewHTTPClient(config.VegaManager.BaseURL)

	return &BaseTestSuite{
		Config: config,
		Client: client,
		T:      t,
	}, nil
}

// Setup 初始化测试环境
func (s *BaseTestSuite) Setup() error {
	// 验证服务可用性
	if err := s.Client.CheckHealth(); err != nil {
		return err
	}

	s.T.Logf("✓ AT测试环境就绪，VEGA Manager: %s", s.Config.VegaManager.BaseURL)
	return nil
}

// APIBasePath 返回API基础路径
func APIBasePath() string {
	return "/api/vega-backend/v1"
}
