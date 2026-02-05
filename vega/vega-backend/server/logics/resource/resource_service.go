// Package resource provides Resource management business logic.
package resource

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/rs/xid"
	"go.opentelemetry.io/otel/codes"

	"vega-backend/common"
	resourceAccess "vega-backend/drivenadapters/resource"
	oerrors "vega-backend/errors"
	"vega-backend/interfaces"
)

var (
	rServiceOnce sync.Once
	rService     interfaces.ResourceService
)

type resourceService struct {
	appSetting *common.AppSetting
	ra         interfaces.ResourceAccess
}

// NewResourceService creates a new ResourceService.
func NewResourceService(appSetting *common.AppSetting) interfaces.ResourceService {
	rServiceOnce.Do(func() {
		rService = &resourceService{
			appSetting: appSetting,
			ra:         resourceAccess.NewResourceAccess(appSetting),
		}
	})
	return rService
}

// Create creates a new Resource.
func (rs *resourceService) Create(ctx context.Context, req *interfaces.ResourceRequest) (string, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Create resource")
	defer span.End()

	// Get account info from context
	accountInfo := interfaces.AccountInfo{}
	if v := ctx.Value(interfaces.ACCOUNT_INFO_KEY); v != nil {
		accountInfo = v.(interfaces.AccountInfo)
	}

	now := time.Now().UnixMilli()
	resource := &interfaces.Resource{
		ID:               xid.New().String(),
		CatalogID:        req.CatalogID,
		Name:             req.Name,
		Tags:             req.Tags,
		Description:      req.Description,
		Category:         req.Category,
		Status:           req.Status,
		Database:         req.Database,
		SourceIdentifier: req.SourceIdentifier,
		Creator:          accountInfo,
		CreateTime:       now,
		Updater:          accountInfo,
		UpdateTime:       now,
	}

	if err := rs.ra.Create(ctx, resource); err != nil {
		logger.Errorf("Create resource failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Create resource failed: %v", err))
		span.SetStatus(codes.Error, "Create resource failed")
		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_Resource_InternalError_CreateFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return resource.ID, nil
}

// Get retrieves a Resource by ID.
func (rs *resourceService) GetByID(ctx context.Context, id string) (*interfaces.Resource, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Get resource")
	defer span.End()

	resource, err := rs.ra.GetByID(ctx, id)
	if err != nil {
		span.SetStatus(codes.Error, "Get resource failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_Resource_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}
	if resource == nil {
		span.SetStatus(codes.Error, "Resource not found")
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, oerrors.VegaManager_Resource_NotFound)
	}

	span.SetStatus(codes.Ok, "")
	return resource, nil
}

// GetByIDs retrieves Resources by IDs.
func (rs *resourceService) GetByIDs(ctx context.Context, ids []string) ([]*interfaces.Resource, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Get resources by IDs")
	defer span.End()

	resources, err := rs.ra.GetByIDs(ctx, ids)
	if err != nil {
		span.SetStatus(codes.Error, "Get resources failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_Resource_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return resources, nil
}

// GetByCatalogID retrieves all Resources under a Catalog.
func (rs *resourceService) GetByCatalogID(ctx context.Context, catalogID string) ([]*interfaces.Resource, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Get resources by catalog ID")
	defer span.End()

	resources, err := rs.ra.GetByCatalogID(ctx, catalogID)
	if err != nil {
		span.SetStatus(codes.Error, "Get resources failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_Resource_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return resources, nil
}

// GetByName retrieves a Resource by catalog and name.
func (rs *resourceService) GetByName(ctx context.Context, catalogID string, name string) (*interfaces.Resource, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Get resource by name")
	defer span.End()

	resource, err := rs.ra.GetByName(ctx, catalogID, name)
	if err != nil {
		span.SetStatus(codes.Error, "Get resource failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_Resource_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}
	if resource == nil {
		span.SetStatus(codes.Error, "Resource not found")
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, oerrors.VegaManager_Resource_NotFound)
	}

	span.SetStatus(codes.Ok, "")
	return resource, nil
}

// List lists Resources with filters.
func (rs *resourceService) List(ctx context.Context, params interfaces.ResourcesQueryParams) ([]*interfaces.Resource, int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "List resources")
	defer span.End()

	resources, total, err := rs.ra.List(ctx, params)
	if err != nil {
		span.SetStatus(codes.Error, "List resources failed")
		return []*interfaces.Resource{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_Resource_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return resources, total, nil
}

// Update updates a Resource.
func (rs *resourceService) Update(ctx context.Context, id string, req *interfaces.ResourceRequest) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Update resource")
	defer span.End()

	resource, err := rs.ra.GetByID(ctx, id)
	if err != nil {
		span.SetStatus(codes.Error, "Get resource failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_Resource_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}
	if resource == nil {
		span.SetStatus(codes.Error, "Resource not found")
		return rest.NewHTTPError(ctx, http.StatusNotFound, oerrors.VegaManager_Resource_NotFound)
	}

	// Apply updates
	if req.Name != resource.Name {
		exists, err := rs.CheckExistByName(ctx, resource.CatalogID, req.Name)
		if err != nil {
			return err
		}
		if exists {
			span.SetStatus(codes.Error, "Resource name exists")
			return rest.NewHTTPError(ctx, http.StatusConflict, oerrors.VegaManager_Resource_NameExists)
		}
		resource.Name = req.Name
	}
	resource.Tags = req.Tags
	resource.Description = req.Description

	// Get account info
	accountInfo := interfaces.AccountInfo{}
	if v := ctx.Value(interfaces.ACCOUNT_INFO_KEY); v != nil {
		accountInfo = v.(interfaces.AccountInfo)
	}

	now := time.Now().UnixMilli()
	resource.Updater = accountInfo
	resource.UpdateTime = now

	if err := rs.ra.Update(ctx, resource); err != nil {
		span.SetStatus(codes.Error, "Update resource failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_Resource_InternalError_UpdateFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// UpdateStatus updates a Resource's status.
func (rs *resourceService) UpdateStatus(ctx context.Context, id string, status string, statusMessage string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Update resource status")
	defer span.End()

	if err := rs.ra.UpdateStatus(ctx, id, status, statusMessage); err != nil {
		span.SetStatus(codes.Error, "Update resource status failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_Resource_InternalError_UpdateFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// DeleteByIDs deletes Resources by IDs.
func (rs *resourceService) DeleteByIDs(ctx context.Context, ids []string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete resources")
	defer span.End()

	if len(ids) == 0 {
		span.SetStatus(codes.Ok, "")
		return nil
	}

	if err := rs.ra.DeleteByIDs(ctx, ids); err != nil {
		span.SetStatus(codes.Error, "Delete resources failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_Resource_InternalError_DeleteFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// CheckExistByID checks if a resource exists by ID.
func (rs *resourceService) CheckExistByID(ctx context.Context, id string) (bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Check resource exist by ID")
	defer span.End()

	resource, err := rs.ra.GetByID(ctx, id)
	if err != nil {
		span.SetStatus(codes.Error, "GetByID failed")
		return false, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_Resource_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return resource != nil, nil
}

// CheckExistByName checks if a Resource exists by name.
func (rs *resourceService) CheckExistByName(ctx context.Context, catalogID string, name string) (bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Check resource exist by name")
	defer span.End()

	resource, err := rs.ra.GetByName(ctx, catalogID, name)
	if err != nil {
		span.SetStatus(codes.Error, "GetByName failed")
		return false, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_Resource_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return resource != nil, nil
}

// UpdateResource updates a Resource directly.
func (rs *resourceService) UpdateResource(ctx context.Context, resource *interfaces.Resource) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Update resource")
	defer span.End()

	if err := rs.ra.Update(ctx, resource); err != nil {
		span.SetStatus(codes.Error, "Update resource failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_Resource_InternalError_UpdateFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}
