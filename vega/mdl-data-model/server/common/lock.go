// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package common

import "sync"

var GLock *GlobalLock[string]

func init() {
	GLock = &GlobalLock[string]{
		mu:      sync.RWMutex{},
		lockMap: make(map[string]*sync.Mutex),
	}
}

type LockMap[K comparable] map[K]*sync.Mutex

type GlobalLock[K comparable] struct {
	mu      sync.RWMutex
	lockMap LockMap[K]
}

func (gLock *GlobalLock[K]) Lock(id K) {
	if _, ok := gLock.lockMap[id]; !ok {
		gLock.mu.Lock()
		gLock.lockMap[id] = &sync.Mutex{}
		gLock.mu.Unlock()
	}

	gLock.mu.RLock()
	defer gLock.mu.RUnlock()

	gLock.lockMap[id].Lock()
}

func (gLock *GlobalLock[K]) Unlock(id K) {
	gLock.mu.RLock()
	defer gLock.mu.RUnlock()

	gLock.lockMap[id].Unlock()
}
