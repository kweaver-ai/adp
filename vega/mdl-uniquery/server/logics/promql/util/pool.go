// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package util

import (
	"uniquery/common"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/panjf2000/ants/v2"
)

var (
	MegerPool       *ants.Pool // promql用于合并各个分片的数据的线程池
	ExecutePool     *ants.Pool // promql提交给opensearch按索引+分片请求数据的线程池
	BatchSubmitPool *ants.Pool // 高基场景下，需要分批次提交的线程池
)

// 还需再比较下其他的协程池
func InitAntsPool(poolSetting common.PoolSetting) {
	// 初始化协程池
	mergePool, err := ants.NewPool(poolSetting.MegerPoolSize, ants.WithPreAlloc(true), ants.WithNonblocking(false))
	if err != nil {
		logger.Fatalf("common.InitAntPool err:%v", err)
		panic(err)
	}
	MegerPool = mergePool

	executePool, err := ants.NewPool(poolSetting.ExecutePoolSize, ants.WithPreAlloc(true), ants.WithNonblocking(false))
	if err != nil {
		logger.Fatalf("common.InitAntPool err:%v", err)
		panic(err)
	}
	ExecutePool = executePool

	batchSubmitPool, err := ants.NewPool(poolSetting.BatchSubmitPoolSize, ants.WithPreAlloc(true), ants.WithNonblocking(false))
	if err != nil {
		logger.Fatalf("common.InitAntPool err:%v", err)
		panic(err)
	}
	BatchSubmitPool = batchSubmitPool
}
