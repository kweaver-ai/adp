// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package auth

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"ontology-manager/common"
	authAccess "ontology-manager/drivenadapters/auth"
	"ontology-manager/interfaces"
)

type hydraAuthService struct {
	aa interfaces.AuthAccess
}

func NewHydraAuthService(appSetting *common.AppSetting) interfaces.AuthService {
	return &hydraAuthService{
		aa: authAccess.NewHydraAuthAccess(appSetting),
	}
}

func (s *hydraAuthService) VerifyToken(ctx context.Context, c *gin.Context) (rest.Visitor, error) {
	return s.aa.VerifyToken(ctx, c)
}
