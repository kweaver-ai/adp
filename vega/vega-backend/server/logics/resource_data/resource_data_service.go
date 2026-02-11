// Package dataset provides Dataset management business logic.
package dataset

import (
	"context"
	"net/http"
	"sync"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/codes"

	"vega-backend/common"
	verrors "vega-backend/errors"
	"vega-backend/interfaces"
	"vega-backend/logics/dataset"
)

var (
	rdServiceOnce sync.Once
	rdService     interfaces.ResourceDataService
)

type resourceDataService struct {
	appSetting *common.AppSetting
	ds         interfaces.DatasetService
}

// NewResourceDataService creates a new ResourceDataService.
func NewResourceDataService(appSetting *common.AppSetting) interfaces.ResourceDataService {
	rdServiceOnce.Do(func() {
		rdService = &resourceDataService{
			appSetting: appSetting,
			ds:         dataset.NewDatasetService(appSetting),
		}
	})
	return rdService
}

// Query 列出 resource 中的文档
func (rds *resourceDataService) Query(ctx context.Context, resource *interfaces.Resource, params *interfaces.ResourceDataQueryParams) ([]map[string]any, int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "List resource documents")
	defer span.End()

	logger.Debugf("Query, resourceID: %s, params: %v", resource.ID, params)

	switch resource.Category {
	case interfaces.ResourceCategoryDataset:
		// 调用 dataset access 列出文档
		documents, total, err := rds.ds.ListDocuments(ctx, resource.ID, params)
		if err != nil {
			span.SetStatus(codes.Error, "List dataset documents failed")
			return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
				WithErrorDetails(err.Error())
		}
		return documents, total, nil

	default:
		span.SetStatus(codes.Error, "Unsupported resource category")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Resource_InternalError_InvalidCategory).
			WithErrorDetails(resource.Category)
	}
}
