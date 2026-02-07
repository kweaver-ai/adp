// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package mysql

import (
	"fmt"
	"time"

	"vega-backend-tests/at/catalog/helpers"
	"vega-backend-tests/at/setup"
)

// MySQLPayloadBuilder MySQL catalog payload构建器
type MySQLPayloadBuilder struct {
	config     setup.MySQLConfig
	testConfig *setup.TestConfig
}

// NewMySQLPayloadBuilder 创建MySQL payload构建器
func NewMySQLPayloadBuilder(config setup.MySQLConfig) *MySQLPayloadBuilder {
	return &MySQLPayloadBuilder{config: config}
}

// SetTestConfig 设置测试配置（包含加密器）
func (b *MySQLPayloadBuilder) SetTestConfig(tc *setup.TestConfig) {
	b.testConfig = tc
}

// encryptPassword 加密密码
func (b *MySQLPayloadBuilder) encryptPassword(password string) string {
	if b.testConfig != nil {
		return b.testConfig.EncryptString(password)
	}
	return password
}

// GetEncryptedPassword 返回加密后的正确密码
func (b *MySQLPayloadBuilder) GetEncryptedPassword() string {
	return b.encryptPassword(b.config.Password)
}

// GetConnectorType 返回connector类型
func (b *MySQLPayloadBuilder) GetConnectorType() string {
	return "mysql"
}

// BuildCreatePayload 构建基本的MySQL catalog创建payload
func (b *MySQLPayloadBuilder) BuildCreatePayload() map[string]any {
	return map[string]any{
		"name":           helpers.GenerateUniqueName("test-mysql-catalog"),
		"connector_type": "mysql",
		"connector_config": map[string]any{
			"host":      b.config.Host,
			"port":      b.config.Port,
			"databases": []string{b.config.Database},
			"username":  b.config.Username,
			"password":  b.encryptPassword(b.config.Password),
		},
	}
}

// BuildFullCreatePayload 构建完整的MySQL catalog创建payload（包含所有可选字段）
func (b *MySQLPayloadBuilder) BuildFullCreatePayload() map[string]any {
	payload := b.BuildCreatePayload()
	payload["description"] = "完整的测试catalog，包含所有可选字段"
	payload["tags"] = []string{"test", "mysql", "at", "full-fields"}

	// 添加MySQL options
	connectorConfig := payload["connector_config"].(map[string]any)
	connectorConfig["options"] = map[string]any{
		"charset":   "utf8mb4",
		"parseTime": "true",
		"loc":       "Local",
	}

	return payload
}

// BuildCreatePayloadWithOptions 构建包含options的MySQL catalog payload
func (b *MySQLPayloadBuilder) BuildCreatePayloadWithOptions(options map[string]any) map[string]any {
	payload := b.BuildCreatePayload()
	connectorConfig := payload["connector_config"].(map[string]any)
	connectorConfig["options"] = options
	return payload
}

// BuildCreatePayloadWithWrongCredentials 构建错误凭证的MySQL catalog payload
func (b *MySQLPayloadBuilder) BuildCreatePayloadWithWrongCredentials() map[string]any {
	payload := b.BuildCreatePayload()
	connectorConfig := payload["connector_config"].(map[string]any)
	connectorConfig["password"] = b.encryptPassword("wrong_password_123")
	return payload
}

// BuildCreatePayloadWithInvalidConfig 构建无效配置的MySQL catalog payload（不存在的数据库）
func (b *MySQLPayloadBuilder) BuildCreatePayloadWithInvalidConfig() map[string]any {
	payload := b.BuildCreatePayload()
	connectorConfig := payload["connector_config"].(map[string]any)
	connectorConfig["databases"] = []string{"nonexistent_db_" + fmt.Sprintf("%d", time.Now().UnixNano())}
	return payload
}

// SupportsTestConnection MySQL支持连接测试
func (b *MySQLPayloadBuilder) SupportsTestConnection() bool {
	return true
}

// GetRequiredConfigFields 返回MySQL connector_config必需的字段
// database 为可选字段，不指定时为实例级连接
func (b *MySQLPayloadBuilder) GetRequiredConfigFields() []string {
	return []string{"host", "port", "username", "password"}
}

// BuildCreatePayloadWithoutDatabase 构建不含database的MySQL catalog payload（实例级连接）
func (b *MySQLPayloadBuilder) BuildCreatePayloadWithoutDatabase() map[string]any {
	return map[string]any{
		"name":           helpers.GenerateUniqueName("test-mysql-instance-catalog"),
		"connector_type": "mysql",
		"connector_config": map[string]any{
			"host":     b.config.Host,
			"port":     b.config.Port,
			"username": b.config.Username,
			"password": b.encryptPassword(b.config.Password),
		},
	}
}

// ========== MySQL特定Payload生成函数 ==========

// BuildCreatePayloadWithInvalidPort 构建无效port的MySQL payload
func (b *MySQLPayloadBuilder) BuildCreatePayloadWithInvalidPort() map[string]any {
	return map[string]any{
		"name":           helpers.GenerateUniqueName("invalid-port-catalog"),
		"connector_type": "mysql",
		"connector_config": map[string]any{
			"host":      b.config.Host,
			"port":      "not_a_number",
			"databases": []string{b.config.Database},
			"username":  b.config.Username,
			"password":  b.encryptPassword(b.config.Password),
		},
	}
}

// BuildCreatePayloadWithNonExistentDB 构建不存在数据库的MySQL payload
func (b *MySQLPayloadBuilder) BuildCreatePayloadWithNonExistentDB() map[string]any {
	return b.BuildCreatePayloadWithInvalidConfig()
}

// GetConfig 返回MySQL配置（供测试中直接使用）
func (b *MySQLPayloadBuilder) GetConfig() setup.MySQLConfig {
	return b.config
}
