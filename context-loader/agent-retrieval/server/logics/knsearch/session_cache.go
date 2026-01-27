// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package knsearch 会话缓存
// file: session_cache.go
// 实现Schema和实例数据的会话级缓存，减少重复查询
package knsearch

import (
	"sync"
	"time"
)

// SessionCache 会话缓存管理器
// TTL: 5分钟（数据一致性优先）
// 多关键词场景下禁用实例缓存
type SessionCache struct {
	mu          sync.RWMutex
	schemaCache map[string]*schemaCacheEntry
	ttl         time.Duration
}

type schemaCacheEntry struct {
	objectTypes   []interface{}
	relationTypes []interface{}
	expireAt      time.Time
}

const defaultCacheTTL = 5 * time.Minute

// NewSessionCache 创建会话缓存实例
func NewSessionCache() *SessionCache {
	return &SessionCache{
		schemaCache: make(map[string]*schemaCacheEntry),
		ttl:         defaultCacheTTL,
	}
}

// GetSchema 获取缓存的Schema
func (c *SessionCache) GetSchema(sessionID, knID string) (objectTypes, relationTypes []interface{}, found bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := c.makeKey(sessionID, knID)
	entry, exists := c.schemaCache[key]
	if !exists {
		return nil, nil, false
	}

	// 检查是否过期
	if time.Now().After(entry.expireAt) {
		return nil, nil, false
	}

	return entry.objectTypes, entry.relationTypes, true
}

// SetSchema 设置Schema缓存
func (c *SessionCache) SetSchema(sessionID, knID string, objectTypes, relationTypes []interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.makeKey(sessionID, knID)
	c.schemaCache[key] = &schemaCacheEntry{
		objectTypes:   objectTypes,
		relationTypes: relationTypes,
		expireAt:      time.Now().Add(c.ttl),
	}
}

// Clear 清除过期缓存
func (c *SessionCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.schemaCache {
		if now.After(entry.expireAt) {
			delete(c.schemaCache, key)
		}
	}
}

// makeKey 生成缓存key
// 格式: schema:{session_id}:{kn_id}
func (c *SessionCache) makeKey(sessionID, knID string) string {
	return "schema:" + sessionID + ":" + knID
}
