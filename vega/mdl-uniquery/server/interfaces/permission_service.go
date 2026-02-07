// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

//go:generate mockgen -source ../interfaces/permission_service.go -destination ../interfaces/mock/mock_permission_service.go
type PermissionService interface {
	CheckPermission(ctx context.Context, resource Resource, ops []string) error
	CheckPermissionWithResult(ctx context.Context, resource Resource, ops []string) (bool, error)
	FilterResources(ctx context.Context, resourceType string, ids []string,
		ops []string, allowOperation bool) (map[string]ResourceOps, error)
	GetResourcesOperations(ctx context.Context, resourceType string, ids []string) ([]ResourceOps, error)
}
