// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package permission

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"uniquery/common"
	derrors "uniquery/errors"
	"uniquery/interfaces"
	"uniquery/logics"
)

var (
	pServiceOnce sync.Once
	pService     interfaces.PermissionService
)

type permissionService struct {
	appSetting *common.AppSetting
	pa         interfaces.PermissionAccess
}

func NewPermissionService(appSetting *common.AppSetting) interfaces.PermissionService {
	pServiceOnce.Do(func() {
		pService = &permissionService{
			appSetting: appSetting,
			pa:         logics.PA,
		}
	})
	return pService
}

func (ps *permissionService) CheckPermission(ctx context.Context, resource interfaces.Resource, ops []string) error {
	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	if accountInfo.ID == "" || accountInfo.Type == "" {
		return rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
			WithErrorDetails("Access denied: missing account ID or type")
	}

	ok, err := ps.pa.CheckPermission(ctx, interfaces.PermissionCheck{
		Accessor: interfaces.Accessor{
			ID:   accountInfo.ID,
			Type: accountInfo.Type,
		},
		Resource:   resource,
		Operations: ops,
	})
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.Uniquery_InternalError_CheckPermissionFailed).WithErrorDetails(err)
	}
	if !ok {
		return rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
			WithErrorDetails(fmt.Sprintf("Access denied: insufficient permissions for[%v]", ops))
	}
	return nil
}

// CheckPermissionWithResult 返回权限检查的结果（bool值）
func (ps *permissionService) CheckPermissionWithResult(ctx context.Context, resource interfaces.Resource, ops []string) (bool, error) {
	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	if accountInfo.ID == "" || accountInfo.Type == "" {
		return false, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
			WithErrorDetails("Access denied: missing account ID or type")
	}

	result, err := ps.pa.CheckPermission(ctx, interfaces.PermissionCheck{
		Accessor: interfaces.Accessor{
			ID:   accountInfo.ID,
			Type: accountInfo.Type,
		},
		Resource:   resource,
		Operations: ops,
	})

	if err != nil {
		return false, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.Uniquery_InternalError_CheckPermissionFailed).WithErrorDetails(err)
	}

	return result, nil
}

// 过滤资源列表
func (ps *permissionService) FilterResources(ctx context.Context, resourceType string, ids []string,
	ops []string, allowOperation bool) (map[string]interfaces.ResourceOps, error) {

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	if accountInfo.ID == "" || accountInfo.Type == "" {
		return nil, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
			WithErrorDetails("Access denied: missing account ID or type")
	}

	resources := []interfaces.Resource{}
	for _, id := range ids {
		resources = append(resources, interfaces.Resource{
			ID:   id,
			Type: resourceType,
		})
	}

	matchResouces, err := ps.pa.FilterResources(ctx, interfaces.ResourcesFilter{
		Accessor: interfaces.Accessor{
			ID:   accountInfo.ID,
			Type: accountInfo.Type,
		},
		Resources:      resources,
		Operations:     ops,
		AllowOperation: allowOperation,
	})
	if err != nil {
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.Uniquery_InternalError_FilterResourcesFailed).WithErrorDetails(err)
	}

	idMap := map[string]interfaces.ResourceOps{}
	for _, resourceOps := range matchResouces {
		idMap[resourceOps.ResourceID] = resourceOps
	}

	return idMap, nil
}

// 获取资源操作
func (ps *permissionService) GetResourcesOperations(ctx context.Context,
	resourceType string, ids []string) ([]interfaces.ResourceOps, error) {

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	if accountInfo.ID == "" || accountInfo.Type == "" {
		return nil, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
			WithErrorDetails("Access denied: missing account ID or type")
	}

	resources := []interfaces.Resource{}
	for _, id := range ids {
		resources = append(resources, interfaces.Resource{
			ID:   id,
			Type: resourceType,
		})
	}

	matchResouces, err := ps.pa.GetResourcesOperations(ctx, interfaces.ResourcesFilter{
		Accessor: interfaces.Accessor{
			ID:   accountInfo.ID,
			Type: accountInfo.Type,
		},
		Resources: resources,
	})
	if err != nil {
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.Uniquery_InternalError_GetResourcesOperationsFailed).WithErrorDetails(err)
	}

	return matchResouces, nil
}
