// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package auth

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"data-model/interfaces"
)

// NoopAuthService 空认证服务（认证禁用时使用）
type NoopAuthService struct{}

func NewNoopAuthService() interfaces.AuthService {
	return &NoopAuthService{}
}

func (n *NoopAuthService) VerifyToken(ctx context.Context, c *gin.Context) (rest.Visitor, error) {
	// 从 header 构建模拟的 Visitor，不做任何认证校验
	accountInfo := interfaces.AccountInfo{
		ID:   c.GetHeader(interfaces.HTTP_HEADER_ACCOUNT_ID),
		Type: c.GetHeader(interfaces.HTTP_HEADER_ACCOUNT_TYPE),
	}
	visitor := rest.Visitor{
		ID:         accountInfo.ID,
		Type:       rest.VisitorType(accountInfo.Type),
		TokenID:    "", // 无token
		IP:         c.ClientIP(),
		Mac:        c.GetHeader("X-Request-MAC"),
		UserAgent:  c.GetHeader("User-Agent"),
		ClientType: rest.ClientType_Linux,
	}
	return visitor, nil
}
