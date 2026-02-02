package fixtures

import (
	"fmt"
	"strings"
	"time"
)

// ========== 名称生成辅助函数 ==========

// GenerateUniqueName 生成唯一名称（带时间戳）
func GenerateUniqueName(prefix string) string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s-%d", prefix, timestamp)
}

// GenerateUniqueNameWithSuffix 生成唯一名称（带自定义后缀）
func GenerateUniqueNameWithSuffix(prefix, suffix string) string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s-%d-%s", prefix, timestamp, suffix)
}

// ========== 字符串构建辅助函数 ==========

// BuildStringWithLength 构建指定长度的字符串
func BuildStringWithLength(char string, length int) string {
	return strings.Repeat(char, length)
}

// BuildUniqueStringWithLength 构建指定长度的唯一字符串（包含时间戳）
func BuildUniqueStringWithLength(length int) string {
	timestamp := fmt.Sprintf("%d", time.Now().UnixNano())
	if length <= len(timestamp) {
		return strings.Repeat("a", length)
	}
	padding := strings.Repeat("a", length-len(timestamp))
	return padding + timestamp
}

// ========== 深拷贝辅助函数 ==========

// DeepCopyMap 深拷贝map
func DeepCopyMap(original map[string]any) map[string]any {
	if original == nil {
		return nil
	}

	result := make(map[string]any)
	for key, value := range original {
		switch v := value.(type) {
		case map[string]any:
			result[key] = DeepCopyMap(v)
		case []any:
			result[key] = DeepCopySlice(v)
		default:
			result[key] = value
		}
	}
	return result
}

// DeepCopySlice 深拷贝slice
func DeepCopySlice(original []any) []any {
	if original == nil {
		return nil
	}

	result := make([]any, len(original))
	for i, value := range original {
		switch v := value.(type) {
		case map[string]any:
			result[i] = DeepCopyMap(v)
		case []any:
			result[i] = DeepCopySlice(v)
		default:
			result[i] = value
		}
	}
	return result
}

// ========== Update辅助函数 ==========

// BuildUpdatePayload 基于现有数据构建更新payload
// 接收原始数据、需要修改的字段和只读字段列表
func BuildUpdatePayload(original map[string]any, updates map[string]any, readOnlyFields []string) map[string]any {
	// 深拷贝原始数据
	payload := DeepCopyMap(original)

	// 移除只读字段
	for _, field := range readOnlyFields {
		delete(payload, field)
	}

	// 应用更新
	for key, value := range updates {
		if value == nil {
			delete(payload, key)
		} else {
			payload[key] = value
		}
	}

	return payload
}
