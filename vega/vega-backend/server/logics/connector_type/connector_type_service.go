// Package connectortype provides ConnectorType management business logic.
package connectortype

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/codes"

	"vega-backend/common"
	connectorTypeAccess "vega-backend/drivenadapters/connector_type"
	oerrors "vega-backend/errors"
	"vega-backend/interfaces"
	"vega-backend/logics/connectors/factory"
)

var (
	ctServiceOnce sync.Once
	ctService     interfaces.ConnectorTypeService
)

type connectorTypeService struct {
	appSetting *common.AppSetting
	cta        interfaces.ConnectorTypeAccess
	cf         *factory.ConnectorFactory
}

// NewConnectorTypeService creates a new ConnectorTypeService.
func NewConnectorTypeService(appSetting *common.AppSetting) interfaces.ConnectorTypeService {
	ctServiceOnce.Do(func() {
		ctService = &connectorTypeService{
			appSetting: appSetting,
			cta:        connectorTypeAccess.NewConnectorTypeAccess(appSetting),
			cf:         factory.GetFactory(),
		}
	})
	return ctService
}

// Register register a new ConnectorType.
func (cts *connectorTypeService) Register(ctx context.Context, req *interfaces.ConnectorTypeReq) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Register connector type")
	defer span.End()

	ct := &interfaces.ConnectorType{
		Type:        req.Type,
		Name:        req.Name,
		Description: req.Description,
		Mode:        req.Mode,
		Category:    req.Category,
		Endpoint:    req.Endpoint,
		FieldConfig: req.FieldConfig,
		Enabled:     req.Enabled,
	}

	if err := cts.cta.Create(ctx, ct); err != nil {
		logger.Errorf("Register connector type failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Register connector type failed: %v", err))
		span.SetStatus(codes.Error, "Register connector type failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_ConnectorType_InternalError_RegisterFailed).
			WithErrorDetails(err.Error())
	}

	if err := cts.cf.RegisterConnector(ctx, ct.Type, ct); err != nil {
		logger.Errorf("Register connector type failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Register connector type failed: %v", err))
		span.SetStatus(codes.Error, "Register connector type failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_ConnectorType_InternalError_RegisterFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// GetByType retrieves a ConnectorType by Type.
func (cts *connectorTypeService) GetByType(ctx context.Context, tp string) (*interfaces.ConnectorType, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Get connector type")
	defer span.End()

	ct, err := cts.cta.GetByType(ctx, tp)
	if err != nil {
		span.SetStatus(codes.Error, "Get connector type failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_ConnectorType_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}
	if ct == nil {
		span.SetStatus(codes.Error, "Connector type not found")
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, oerrors.VegaManager_ConnectorType_NotFound)
	}

	span.SetStatus(codes.Ok, "")
	return ct, nil
}

// List lists ConnectorTypes with filters.
func (cts *connectorTypeService) List(ctx context.Context, params interfaces.ConnectorTypesQueryParams) ([]*interfaces.ConnectorType, int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "List connector types")
	defer span.End()

	connectorTypes, total, err := cts.cta.List(ctx, params)
	if err != nil {
		span.SetStatus(codes.Error, "List connector types failed")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_ConnectorType_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return connectorTypes, total, nil
}

// Update updates a ConnectorType.
func (cts *connectorTypeService) Update(ctx context.Context, req *interfaces.ConnectorTypeReq) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Update connector type")
	defer span.End()

	ct := req.OriginConnectorType
	if ct == nil {
		span.SetStatus(codes.Error, "Connector type not found")
		return rest.NewHTTPError(ctx, http.StatusNotFound, oerrors.VegaManager_ConnectorType_NotFound)
	}

	// Apply updates
	if req.Name != ct.Name {
		exists, err := cts.CheckExistByName(ctx, req.Name)
		if err != nil {
			return err
		}
		if exists {
			span.SetStatus(codes.Error, "connector type name exists")
			return rest.NewHTTPError(ctx, http.StatusConflict, oerrors.VegaManager_ConnectorType_NameExists)
		}
		ct.Name = req.Name
	}

	if req.Type != ct.Type {
		span.SetStatus(codes.Error, "can not change connector type")
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.VegaManager_ConnectorType_InvalidParameter_Type)
	}
	if req.Mode != ct.Mode {
		span.SetStatus(codes.Error, "can not change connector mode")
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.VegaManager_ConnectorType_InvalidParameter_Mode)
	}
	if req.Category != ct.Category {
		span.SetStatus(codes.Error, "can not change connector category")
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.VegaManager_ConnectorType_InvalidParameter_Category)
	}

	ct.Type = req.Type
	ct.Description = req.Description
	ct.Mode = req.Mode
	ct.Category = req.Category
	ct.Endpoint = req.Endpoint
	ct.FieldConfig = req.FieldConfig
	ct.Enabled = req.Enabled

	if err := cts.cta.Update(ctx, ct); err != nil {
		span.SetStatus(codes.Error, "Update connector type failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_ConnectorType_InternalError_UpdateFailed).
			WithErrorDetails(err.Error())
	}

	if err := cts.cf.RegisterConnector(ctx, ct.Type, ct); err != nil {
		logger.Errorf("Register connector type failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Register connector type failed: %v", err))
		span.SetStatus(codes.Error, "Register connector type failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_ConnectorType_InternalError_RegisterFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// Delete deletes a ConnectorType.
func (cts *connectorTypeService) DeleteByType(ctx context.Context, tp string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete connector type")
	defer span.End()

	if err := cts.cta.DeleteByType(ctx, tp); err != nil {
		span.SetStatus(codes.Error, "Delete connector type failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_ConnectorType_InternalError_DeleteFailed).
			WithErrorDetails(err.Error())
	}

	if err := cts.cf.DeleteConnector(ctx, tp); err != nil {
		logger.Errorf("Delete connector type failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Delete connector type failed: %v", err))
		span.SetStatus(codes.Error, "Delete connector type failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_ConnectorType_InternalError_DeleteFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// SetEnabled sets the enabled status of a ConnectorType.
func (cts *connectorTypeService) SetEnabled(ctx context.Context, tp string, enabled bool) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Set enabled connector type")
	defer span.End()

	if err := cts.cta.SetEnabled(ctx, tp, enabled); err != nil {
		span.SetStatus(codes.Error, "Set enabled connector type failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_ConnectorType_InternalError_UpdateFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// CheckExistByType checks if a ConnectorType exists by Type.
func (cts *connectorTypeService) CheckExistByType(ctx context.Context, tp string) (bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Check connector type exist by Type")
	defer span.End()

	ct, err := cts.cta.GetByType(ctx, tp)
	if err != nil {
		span.SetStatus(codes.Error, "GetByType failed")
		return false, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_Catalog_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return ct != nil, nil
}

// CheckExistByName checks if a ConnectorType exists by Name.
func (cts *connectorTypeService) CheckExistByName(ctx context.Context, name string) (bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Check connector type exist by Name")
	defer span.End()

	ct, err := cts.cta.GetByName(ctx, name)
	if err != nil {
		span.SetStatus(codes.Error, "GetByName failed")
		return false, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_ConnectorType_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return ct != nil, nil
}
