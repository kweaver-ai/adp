// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
)

const (
	SOURCE_TYPE_ANYROBOT = "anyrobot"
	SOURCE_TYPE_TINGYUN  = "tingyun"

	APPLICATION_OBJECT_LOG_GROUP   = "log_group"
	APPLICATION_OBJECT_TRACE_MODEL = "trace_model"
)

var (
	DataSourceType2ApplicationScope = map[string][]string{
		SOURCE_TYPE_ANYROBOT: {APPLICATION_OBJECT_LOG_GROUP},
		SOURCE_TYPE_TINGYUN:  {APPLICATION_OBJECT_TRACE_MODEL},
	}
	ApplicationObject2DataSourceTypes = map[string][]string{
		APPLICATION_OBJECT_LOG_GROUP:   {SOURCE_TYPE_ANYROBOT},
		APPLICATION_OBJECT_TRACE_MODEL: {SOURCE_TYPE_TINGYUN},
	}
)

//go:generate mockgen -source ../interfaces/data_connection_processor.go -destination ../interfaces/mock/mock_data_connection_processor.go
type DataConnectionProcessor interface {
	// 创建时的校验函数
	ValidateWhenCreate(ctx context.Context, conn *DataConnection) error
	// 修改时的校验函数
	ValidateWhenUpdate(ctx context.Context, conn *DataConnection, preConn *DataConnection) error
	// 计算详细配置的md5
	ComputeConfigMD5(ctx context.Context, conn *DataConnection) (string, error)
	// 生成auth_info和连接状态
	GenerateAuthInfoAndStatus(ctx context.Context, conn *DataConnection) error
	// 更新auth_info和连接状态
	UpdateAuthInfoAndStatus(ctx context.Context, conn *DataConnection) (bool, error)
	// 隐藏auth_info, 不让查询
	HideAuthInfo(ctx context.Context, conn *DataConnection) error
}
