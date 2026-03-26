// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
)

// AgentOperatorIntegrationAccess 访问 agent-operator-integration 的接口
//
//go:generate mockgen -source=agent_operator_integration_access.go -destination=mock/mock_agent_operator_integration_access.go -package=mock
type AgentOperatorIntegrationAccess interface {
	// RegisterInternalTool 注册/更新内置工具，body 为请求体 JSON
	RegisterInternalTool(ctx context.Context, body []byte) error
	// GetToolBoxToolByID sends a minimal GET to tool-box to verify the tool exists (HTTP layer / gateway).
	GetToolBoxToolByID(ctx context.Context, boxID string, toolID string) error
}
