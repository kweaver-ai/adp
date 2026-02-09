// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package auth

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"data-model/common"
	"data-model/interfaces"
)

type hydraAuthService struct {
	hydra rest.Hydra
}

func NewHydraAuthService(appSetting *common.AppSetting) interfaces.AuthService {
	return &hydraAuthService{
		hydra: rest.NewHydra(appSetting.HydraAdminSetting),
	}
}

func (s *hydraAuthService) VerifyToken(ctx context.Context, c *gin.Context) (rest.Visitor, error) {
	return s.hydra.VerifyToken(ctx, c)
}
