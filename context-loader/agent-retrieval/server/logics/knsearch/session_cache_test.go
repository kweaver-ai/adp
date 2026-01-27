// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knsearch

import (
	"testing"
	"time"
)

func TestSessionCache_SetAndGet(t *testing.T) {
	cache := NewSessionCache()

	// 设置缓存
	objectTypes := []interface{}{
		map[string]interface{}{"id": "obj1", "name": "对象1"},
		map[string]interface{}{"id": "obj2", "name": "对象2"},
	}
	relationTypes := []interface{}{
		map[string]interface{}{"id": "rel1", "name": "关系1"},
	}

	cache.SetSchema("session-1", "kn-1", objectTypes, relationTypes)

	// 获取缓存
	objs, rels, found := cache.GetSchema("session-1", "kn-1")
	if !found {
		t.Error("Expected to find cached schema")
	}

	if len(objs) != 2 {
		t.Errorf("Expected 2 object types, got %d", len(objs))
	}

	if len(rels) != 1 {
		t.Errorf("Expected 1 relation type, got %d", len(rels))
	}
}

func TestSessionCache_NotFound(t *testing.T) {
	cache := NewSessionCache()

	_, _, found := cache.GetSchema("nonexistent", "kn-1")
	if found {
		t.Error("Expected not to find nonexistent cache")
	}
}

func TestSessionCache_DifferentKeys(t *testing.T) {
	cache := NewSessionCache()

	cache.SetSchema("session-1", "kn-1", []interface{}{"obj1"}, []interface{}{"rel1"})
	cache.SetSchema("session-1", "kn-2", []interface{}{"obj2"}, []interface{}{"rel2"})
	cache.SetSchema("session-2", "kn-1", []interface{}{"obj3"}, []interface{}{"rel3"})

	// 验证不同key的缓存独立
	objs1, _, found1 := cache.GetSchema("session-1", "kn-1")
	objs2, _, found2 := cache.GetSchema("session-1", "kn-2")
	objs3, _, found3 := cache.GetSchema("session-2", "kn-1")

	if !found1 || !found2 || !found3 {
		t.Error("All caches should be found")
	}

	if objs1[0] != "obj1" || objs2[0] != "obj2" || objs3[0] != "obj3" {
		t.Error("Cache values should be independent")
	}
}

func TestSessionCache_Expiry(t *testing.T) {
	cache := &SessionCache{
		schemaCache: make(map[string]*schemaCacheEntry),
		ttl:         100 * time.Millisecond, // 短TTL用于测试
	}

	cache.SetSchema("session-1", "kn-1", []interface{}{"obj"}, []interface{}{"rel"})

	// 立即获取应该成功
	_, _, found := cache.GetSchema("session-1", "kn-1")
	if !found {
		t.Error("Cache should be found immediately")
	}

	// 等待过期
	time.Sleep(150 * time.Millisecond)

	// 过期后应该找不到
	_, _, found = cache.GetSchema("session-1", "kn-1")
	if found {
		t.Error("Cache should have expired")
	}
}

func TestSessionCache_Clear(t *testing.T) {
	cache := &SessionCache{
		schemaCache: make(map[string]*schemaCacheEntry),
		ttl:         100 * time.Millisecond,
	}

	// 添加两个缓存项
	cache.SetSchema("session-1", "kn-1", []interface{}{"obj1"}, []interface{}{})

	// 等待一个过期
	time.Sleep(150 * time.Millisecond)

	// 添加新的缓存项
	cache.SetSchema("session-2", "kn-2", []interface{}{"obj2"}, []interface{}{})

	// 清理过期缓存
	cache.Clear()

	// 过期的应该被清除
	_, _, found1 := cache.GetSchema("session-1", "kn-1")
	if found1 {
		t.Error("Expired cache should be cleared")
	}

	// 未过期的应该还在
	_, _, found2 := cache.GetSchema("session-2", "kn-2")
	if !found2 {
		t.Error("Non-expired cache should remain")
	}
}
