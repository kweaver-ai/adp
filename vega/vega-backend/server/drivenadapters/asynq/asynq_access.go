// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package drivenadapters provides Asynq task queue implementation.
package drivenadapters

import (
	"context"
	"fmt"
	"sync"

	"github.com/hibiken/asynq"
	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"vega-backend/common"
	"vega-backend/interfaces"
)

var (
	aqAccessOnce sync.Once
	aqAccess     interfaces.AsynqAccess
)

type asynqAccess struct {
	appSetting *common.AppSetting
}

// NewAsynqAccess creates or returns the singleton AsynqAccess implementation.
func NewAsynqAccess(appSetting *common.AppSetting) interfaces.AsynqAccess {
	aqAccessOnce.Do(func() {
		aqAccess = &asynqAccess{
			appSetting: appSetting,
		}
	})

	return aqAccess
}

// CreateClient creates and returns the Asynq client for enqueueing tasks.
func (aqa *asynqAccess) CreateClient(ctx context.Context) *asynq.Client {
	redisOpt := asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%d", aqa.appSetting.RedisSetting.Host, aqa.appSetting.RedisSetting.Port),
		Username: aqa.appSetting.RedisSetting.Username,
		Password: aqa.appSetting.RedisSetting.Password,
	}
	return asynq.NewClient(redisOpt)
}

// CreateServer creates and returns the Asynq server for processing tasks.
func (aqa *asynqAccess) CreateServer(ctx context.Context) *asynq.Server {
	redisOpt := asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%d", aqa.appSetting.RedisSetting.Host, aqa.appSetting.RedisSetting.Port),
		Username: aqa.appSetting.RedisSetting.Username,
		Password: aqa.appSetting.RedisSetting.Password,
	}
	return asynq.NewServer(redisOpt, asynq.Config{
		Concurrency: 10,
		Queues: map[string]int{
			"high":    6,
			"default": 3,
			"low":     1,
		},
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			logger.Errorf("Task %s failed: %v", task.Type(), err)
		}),
	})
}
