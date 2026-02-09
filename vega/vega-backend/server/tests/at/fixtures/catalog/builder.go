// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package catalog

import (
	"vega-backend/tests/at/setup"
)

// PayloadBuilder Catalog payload构建器接口
// 定义不同类型catalog（MySQL、OpenSearch等）需要实现的payload生成方法
type PayloadBuilder interface {
	// GetConnectorType 返回connector类型标识符（如 "mysql", "opensearch"）
	GetConnectorType() string

	// BuildCreatePayload 构建基本的catalog创建payload（必填字段）
	BuildCreatePayload() map[string]any

	// BuildFullCreatePayload 构建完整的catalog创建payload（包含所有可选字段）
	BuildFullCreatePayload() map[string]any

	// BuildCreatePayloadWithOptions 构建包含自定义options的payload
	BuildCreatePayloadWithOptions(options map[string]any) map[string]any

	// BuildCreatePayloadWithWrongCredentials 构建错误凭证的payload（用于测试连接失败场景）
	BuildCreatePayloadWithWrongCredentials() map[string]any

	// BuildCreatePayloadWithInvalidConfig 构建无效配置的payload（如不存在的host/database）
	BuildCreatePayloadWithInvalidConfig() map[string]any

	// SupportsTestConnection 是否支持连接测试
	SupportsTestConnection() bool

	// GetRequiredConfigFields 返回connector_config必需的字段名列表
	GetRequiredConfigFields() []string

	// GetEncryptedPassword 返回加密后的正确密码（用于update时回填）
	GetEncryptedPassword() string
}

// NewPayloadBuilder 根据connector类型创建对应的PayloadBuilder
// 工厂方法，用于统一创建不同类型的Builder
func NewPayloadBuilder(connectorType string, config *setup.TestConfig) PayloadBuilder {
	switch connectorType {
	case "mysql":
		b := NewMySQLPayloadBuilder(config.TargetMySQL)
		b.SetTestConfig(config)
		return b
	case "opensearch":
		b := NewOpenSearchPayloadBuilder(config.TargetOpenSearch)
		b.SetTestConfig(config)
		return b
	default:
		return nil
	}
}

// SupportedConnectorTypes 返回支持的所有connector类型
func SupportedConnectorTypes() []string {
	return []string{"mysql", "opensearch"}
}
