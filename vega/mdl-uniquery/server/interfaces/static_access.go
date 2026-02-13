// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "time"

type IndexBaseSplitTime struct {
	BaseType  string    `json:"base_type"`
	SplitTime time.Time `json:"split_time"`
}

//go:generate mockgen -source ../interfaces/static_access.go -destination ../interfaces/mock/mock_static_access.go
type StaticAccess interface {
	// 根据名称获取到ID ，可以用于根据名称获取到指标模型对象信息
	GetIndexBaseSplitTime() ([]IndexBaseSplitTime, error)
}
