package catalog

import (
	"fmt"
	"strings"
	"time"

	"vega-backend/tests/at/fixtures"
)

// catalogReadOnlyFields Catalog更新时需要移除的只读字段
var catalogReadOnlyFields = []string{
	"id", "type", "create_time", "update_time",
	"creator", "updater",
	"health_check_status", "health_check_time", "health_check_message",
}

// ========== 通用Payload生成函数 ==========

// BuildMinimalPayload 构建最小字段的catalog payload（仅必填字段）
func BuildMinimalPayload() map[string]any {
	return map[string]any{
		"name": GenerateUniqueName("minimal-catalog"),
	}
}

// BuildLogicalCatalogPayload 构建逻辑catalog的payload
func BuildLogicalCatalogPayload() map[string]any {
	return map[string]any{
		"name":        GenerateUniqueName("logical-catalog"),
		"description": "逻辑目录测试",
		"tags":        []string{"logical", "test"},
	}
}

// ========== Name相关Payload ==========

// BuildPayloadWithLongName 构建指定长度name的payload
func BuildPayloadWithLongName(length int) map[string]any {
	return map[string]any{
		"name": BuildUniqueStringWithLength(length),
	}
}

// BuildPayloadWithMinName 构建最小长度name的payload（单字符）
func BuildPayloadWithMinName() map[string]any {
	chars := "abcdefghijklmnopqrstuvwxyz"
	idx := time.Now().UnixNano() % int64(len(chars))
	return map[string]any{
		"name": string(chars[idx]),
	}
}

// BuildPayloadWithEmptyName 构建空name字段的payload
func BuildPayloadWithEmptyName() map[string]any {
	return map[string]any{
		"name": "",
	}
}

// BuildPayloadWithMissingName 构建缺少name字段的payload
func BuildPayloadWithMissingName() map[string]any {
	return map[string]any{
		"connector_type": "mysql",
	}
}

// BuildPayloadWithWhitespaceName 构建只有空格name的payload
func BuildPayloadWithWhitespaceName() map[string]any {
	return map[string]any{
		"name": "   ",
	}
}

// BuildPayloadWithSpecialCharsName 构建包含特殊字符名称的payload
func BuildPayloadWithSpecialCharsName(specialName string) map[string]any {
	return map[string]any{
		"name": specialName,
	}
}

// ========== Description相关Payload ==========

// BuildPayloadWithLongDescription 构建超长description的payload
func BuildPayloadWithLongDescription(descriptionLength int) map[string]any {
	longDescription := strings.Repeat("这是一个很长的注释。", descriptionLength/10+1)
	return map[string]any{
		"name":        GenerateUniqueName("long-description-catalog"),
		"description": longDescription,
	}
}

// BuildPayloadWithExactDescription 构建指定长度description的payload
func BuildPayloadWithExactDescription(descriptionLength int) map[string]any {
	description := strings.Repeat("c", descriptionLength)
	return map[string]any{
		"name":        GenerateUniqueName("exact-description-catalog"),
		"description": description,
	}
}

// ========== Tags相关Payload ==========

// BuildPayloadWithManyTags 构建包含大量tags的payload
func BuildPayloadWithManyTags(tagCount int) map[string]any {
	tags := make([]string, tagCount)
	for i := range tags {
		tags[i] = fmt.Sprintf("tag-%d", i)
	}
	return map[string]any{
		"name": GenerateUniqueName("many-tags-catalog"),
		"tags": tags,
	}
}

// BuildPayloadWithEmptyTags 构建空tags数组的payload
func BuildPayloadWithEmptyTags() map[string]any {
	return map[string]any{
		"name": GenerateUniqueName("empty-tags-catalog"),
		"tags": []string{},
	}
}

// BuildPayloadWithEmptyTagInArray 构建包含空字符串tag的payload
func BuildPayloadWithEmptyTagInArray() map[string]any {
	return map[string]any{
		"name": GenerateUniqueName("empty-tag-catalog"),
		"tags": []string{"valid-tag", ""},
	}
}

// BuildPayloadWithInvalidCharTag 构建包含非法字符tag的payload
func BuildPayloadWithInvalidCharTag() map[string]any {
	return map[string]any{
		"name": GenerateUniqueName("invalid-char-tag-catalog"),
		"tags": []string{"tag/name", "valid"},
	}
}

// BuildPayloadWithLongTag 构建指定长度tag的payload
func BuildPayloadWithLongTag(tagLength int) map[string]any {
	longTag := strings.Repeat("t", tagLength)
	return map[string]any{
		"name": GenerateUniqueName("long-tag-catalog"),
		"tags": []string{longTag},
	}
}

// ========== Connector相关Payload ==========

// BuildInvalidPayload 构建无效的payload（用于负向测试）
func BuildInvalidPayload() map[string]any {
	return map[string]any{
		"name": "",
	}
}

// BuildPayloadWithInvalidConnectorType 构建无效connector_type的payload
func BuildPayloadWithInvalidConnectorType() map[string]any {
	return map[string]any{
		"name":           GenerateUniqueName("invalid-connector-type"),
		"connector_type": "invalid_type_xyz",
	}
}

// BuildPayloadWithMissingConnectorFields 构建缺少connector_config必要字段的payload
func BuildPayloadWithMissingConnectorFields() map[string]any {
	return map[string]any{
		"name":           GenerateUniqueName("missing-connector-fields"),
		"connector_type": "mysql",
		"connector_config": map[string]any{
			"host": "localhost",
			// 缺少port, username, password（database为可选字段）
		},
	}
}

// ========== 安全测试Payload ==========

// BuildPayloadWithSQLInjection 构建SQL注入测试payload
func BuildPayloadWithSQLInjection() map[string]any {
	return map[string]any{
		"name": fmt.Sprintf("test'; DROP TABLE t_catalog; --%d", time.Now().UnixNano()),
	}
}

// BuildPayloadWithXSS 构建XSS测试payload
func BuildPayloadWithXSS() map[string]any {
	return map[string]any{
		"name": fmt.Sprintf("<script>alert('xss')</script>%d", time.Now().UnixNano()),
	}
}

// BuildPayloadWithPathTraversal 构建路径遍历测试payload
func BuildPayloadWithPathTraversal() map[string]any {
	return map[string]any{
		"name": fmt.Sprintf("../../../etc/passwd%d", time.Now().UnixNano()),
	}
}

// ========== 辅助函数（委托到fixtures包） ==========

// GenerateUniqueName 生成唯一名称（带时间戳）
func GenerateUniqueName(prefix string) string {
	return fixtures.GenerateUniqueName(prefix)
}

// GenerateUniqueNameWithSuffix 生成唯一名称（带自定义后缀）
func GenerateUniqueNameWithSuffix(prefix, suffix string) string {
	return fixtures.GenerateUniqueNameWithSuffix(prefix, suffix)
}

// BuildStringWithLength 构建指定长度的字符串
func BuildStringWithLength(char string, length int) string {
	return fixtures.BuildStringWithLength(char, length)
}

// BuildUniqueStringWithLength 构建指定长度的唯一字符串（包含时间戳）
func BuildUniqueStringWithLength(length int) string {
	return fixtures.BuildUniqueStringWithLength(length)
}

// ========== Update辅助函数 ==========

// BuildUpdatePayload 基于现有catalog数据构建更新payload
// 接收原始catalog数据和需要修改的字段
func BuildUpdatePayload(original map[string]any, updates map[string]any) map[string]any {
	return fixtures.BuildUpdatePayload(original, updates, catalogReadOnlyFields)
}

// InjectEncryptedPassword 向payload的connector_config中注入加密后的password
// 因为GET响应不再返回敏感字段，update时需要重新注入
func InjectEncryptedPassword(payload map[string]any, encryptedPassword string) {
	if connCfg, ok := payload["connector_config"].(map[string]any); ok {
		connCfg["password"] = encryptedPassword
	}
}

// DeepCopyMap 深拷贝map
func DeepCopyMap(original map[string]any) map[string]any {
	return fixtures.DeepCopyMap(original)
}

// DeepCopySlice 深拷贝slice
func DeepCopySlice(original []any) []any {
	return fixtures.DeepCopySlice(original)
}
