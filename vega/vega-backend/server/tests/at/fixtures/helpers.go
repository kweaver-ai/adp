package fixtures

import (
	"net/http"
	"testing"
	"time"

	"vega-backend/tests/testutil"
)

// CleanupCatalogs 清理所有现有的catalog，确保测试环境干净
func CleanupCatalogs(client *testutil.HTTPClient, t *testing.T) {
	// 获取所有catalog（使用offset+limit分页）
	listResp := client.GET("/api/vega-backend/v1/catalogs?offset=0&limit=1000")
	if listResp.StatusCode != http.StatusOK {
		t.Logf("⚠ 获取catalog列表失败，状态码: %d", listResp.StatusCode)
		return
	}

	if listResp.Body == nil {
		t.Logf("✓ catalog列表为空，无需清理")
		return
	}

	entries, ok := listResp.Body["entries"].([]any)
	if !ok || len(entries) == 0 {
		t.Logf("✓ catalog列表为空，无需清理")
		return
	}

	t.Logf("⏳ 开始清理 %d 个现有catalog...", len(entries))

	deletedCount := 0
	for _, entry := range entries {
		catalogMap, ok := entry.(map[string]any)
		if !ok {
			continue
		}

		catalogID, ok := catalogMap["id"].(string)
		if !ok {
			continue
		}

		deleteResp := client.DELETE("/api/vega-backend/v1/catalogs/" + catalogID)
		if deleteResp.StatusCode == http.StatusOK || deleteResp.StatusCode == http.StatusNoContent {
			deletedCount++
		} else {
			t.Logf("⚠ 删除catalog %s 失败，状态码: %d", catalogID, deleteResp.StatusCode)
		}
	}

	t.Logf("✓ 清理完成，成功删除 %d 个catalog", deletedCount)
}

// CleanupResources 清理所有现有的resource，确保测试环境干净
// 将来resource测试使用
func CleanupResources(client *testutil.HTTPClient, t *testing.T) {
	// 获取所有resource
	listResp := client.GET("/api/vega-backend/v1/resources?offset=0&limit=1000")
	if listResp.StatusCode != http.StatusOK {
		t.Logf("⚠ 获取resource列表失败，状态码: %d", listResp.StatusCode)
		return
	}

	if listResp.Body == nil {
		t.Logf("✓ resource列表为空，无需清理")
		return
	}

	entries, ok := listResp.Body["entries"].([]any)
	if !ok || len(entries) == 0 {
		t.Logf("✓ resource列表为空，无需清理")
		return
	}

	t.Logf("⏳ 开始清理 %d 个现有resource...", len(entries))

	deletedCount := 0
	for _, entry := range entries {
		resourceMap, ok := entry.(map[string]any)
		if !ok {
			continue
		}

		resourceID, ok := resourceMap["id"].(string)
		if !ok {
			continue
		}

		deleteResp := client.DELETE("/api/vega-backend/v1/resources/" + resourceID)
		if deleteResp.StatusCode == http.StatusOK || deleteResp.StatusCode == http.StatusNoContent {
			deletedCount++
		} else {
			t.Logf("⚠ 删除resource %s 失败，状态码: %d", resourceID, deleteResp.StatusCode)
		}
	}

	t.Logf("✓ 清理完成，成功删除 %d 个resource", deletedCount)
}

// ExtractFromEntriesResponse 从entries格式响应中提取第一个对象
// GET /xxx/:id 返回 {"entries": [...]} 格式
func ExtractFromEntriesResponse(resp testutil.HTTPResponse) map[string]any {
	if resp.Body == nil {
		return nil
	}

	// 尝试从entries数组中提取
	if entries, ok := resp.Body["entries"].([]any); ok && len(entries) > 0 {
		if item, ok := entries[0].(map[string]any); ok {
			return item
		}
	}

	// 如果响应直接就是对象（向后兼容）
	if _, hasID := resp.Body["id"]; hasID {
		return resp.Body
	}

	return nil
}

// ExtractID 从创建响应中提取ID
func ExtractID(resp testutil.HTTPResponse) string {
	if resp.Body == nil {
		return ""
	}

	if id, ok := resp.Body["id"].(string); ok {
		return id
	}

	return ""
}

// WaitForCondition 等待条件满足（用于异步操作）
func WaitForCondition(condition func() bool, timeout time.Duration, interval time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(interval)
	}
	return false
}
