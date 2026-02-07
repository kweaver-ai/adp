// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package errors 服务错误码
package errors

// Task 相关错误码
const (
	// 404 Not Found
	VegaManager_Task_NotFound = "VegaManager.Task.NotFound"
)

var (
	TaskErrCodeList = []string{
		VegaManager_Task_NotFound,
	}
)
