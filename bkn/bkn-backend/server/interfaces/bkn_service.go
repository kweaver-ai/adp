// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"

	bknsdk "github.com/kweaver-ai/bkn-specification/sdk/golang/bkn"
)

// BKNService BKN 导入导出服务接口
//
//go:generate mockgen -source ../interfaces/bkn_service.go -destination ../interfaces/mock/mock_bkn_service.go
type BKNService interface {
	// ImportFromTar 从 tar 包导入 BKN 定义
	Import(ctx context.Context, bknNetwork *bknsdk.BknNetwork) (string, error)
	// ExportToTar 将知识网络导出为 tar 包
	ExportToTar(ctx context.Context, knID string, branch string) ([]byte, error)
}
