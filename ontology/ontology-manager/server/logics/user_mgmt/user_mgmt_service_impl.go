// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package user_mgmt

import (
	"context"

	"ontology-manager/common"
	umAccess "ontology-manager/drivenadapters/user_mgmt"
	"ontology-manager/interfaces"
)

type UserMgmtServiceImpl struct {
	uma interfaces.UserMgmtAccess
}

func NewUserMgmtServiceImpl(appSetting *common.AppSetting) interfaces.UserMgmtService {
	return &UserMgmtServiceImpl{
		uma: umAccess.NewUserMgmtAccess(appSetting),
	}
}

func (s *UserMgmtServiceImpl) GetAccountNames(ctx context.Context, accountInfos []*interfaces.AccountInfo) error {
	return s.uma.GetAccountNames(ctx, accountInfos)
}
