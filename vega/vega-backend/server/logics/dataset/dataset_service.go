// Package dataset provides Dataset management business logic.
package dataset

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
	datasetAccess "vega-backend/drivenadapters/dataset"
	verrors "vega-backend/errors"
	"vega-backend/interfaces"
)

var (
	dsServiceOnce sync.Once
	dsService     interfaces.DatasetService
)

type datasetService struct {
	appSetting *common.AppSetting
	da         interfaces.DatasetAccess
}

// NewDatasetService creates a new DatasetService.
func NewDatasetService(appSetting *common.AppSetting) interfaces.DatasetService {
	dsServiceOnce.Do(func() {
		dsService = &datasetService{
			appSetting: appSetting,
			da:         datasetAccess.NewDatasetAccess(appSetting),
		}
	})
	return dsService
}

// Create a new Dataset.
func (ds *datasetService) Create(ctx context.Context, res *interfaces.Resource) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Create dataset")
	defer span.End()

	// 调用 dataset access 创建 dataset 索引，索引名称为 <res.source_identifier>-<catalog_id>
	err := ds.da.Create(ctx, res.ID, res.SchemaDefinition)
	if err != nil {
		logger.Errorf("Create dataset index failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Create dataset index failed: %v", err))
		span.SetStatus(codes.Error, "Create dataset index failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_CreateFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// Update a Dataset.
func (ds *datasetService) Update(ctx context.Context, res *interfaces.Resource) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Update dataset")
	defer span.End()

	// 调用 dataset access 更新 dataset 索引，索引名称为 <res.source_identifier>-<id>
	if err := ds.da.Update(ctx, fmt.Sprintf("%s-%s", res.SourceIdentifier, res.ID), res.SchemaDefinition); err != nil {
		span.SetStatus(codes.Error, "Update dataset failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_UpdateFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// Delete a Dataset.
func (ds *datasetService) Delete(ctx context.Context, res *interfaces.Resource) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete dataset")
	defer span.End()

	// Check dataset exist first
	exist, err := ds.da.CheckExist(ctx, res.ID)
	if err != nil {
		span.SetStatus(codes.Error, "Check dataset exist failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
			WithErrorDetails(err.Error())
	}
	if exist {
		// Delete from storage
		if err := ds.da.Delete(ctx, res.ID); err != nil {
			span.SetStatus(codes.Error, "Delete dataset failed")
			return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_DeleteFailed).
				WithErrorDetails(err.Error())
		}
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// ListDocuments 列出 dataset 中的文档
func (ds *datasetService) ListDocuments(ctx context.Context, res *interfaces.Resource, params *interfaces.ResourceDataQueryParams) ([]map[string]any, int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "List dataset documents")
	defer span.End()

	// 调用 dataset access 列出文档
	documents, total, err := ds.da.ListDocuments(ctx, res.ID, params)
	if err != nil {
		span.SetStatus(codes.Error, "List dataset documents failed")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return documents, total, nil
}

// CreateDocuments 批量创建 dataset 文档
func (ds *datasetService) CreateDocuments(ctx context.Context, id string, documents []map[string]any) ([]string, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Create dataset documents")
	defer span.End()

	// 调用 dataset access 批量创建文档
	docIDs, err := ds.da.CreateDocuments(ctx, id, documents)
	if err != nil {
		span.SetStatus(codes.Error, "Create dataset documents failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_CreateFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return docIDs, nil
}

// GetDocument 获取 dataset 文档
func (ds *datasetService) GetDocument(ctx context.Context, id string, docID string) (map[string]any, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Get dataset document")
	defer span.End()

	// 调用 dataset access 获取文档
	document, err := ds.da.GetDocument(ctx, id, docID)
	if err != nil {
		span.SetStatus(codes.Error, "Get dataset document failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return document, nil
}

// UpdateDocument 更新 dataset 文档
func (ds *datasetService) UpdateDocument(ctx context.Context, id string, docID string, document map[string]any) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Update dataset document")
	defer span.End()

	// 调用 dataset access 更新文档
	if err := ds.da.UpdateDocument(ctx, id, docID, document); err != nil {
		span.SetStatus(codes.Error, "Update dataset document failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_UpdateFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// DeleteDocument 删除 dataset 文档
func (ds *datasetService) DeleteDocument(ctx context.Context, id string, docID string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete dataset document")
	defer span.End()

	// 调用 dataset access 删除文档
	if err := ds.da.DeleteDocument(ctx, id, docID); err != nil {
		span.SetStatus(codes.Error, "Delete dataset document failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_DeleteFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// UpdateDocuments 批量更新 dataset 文档
func (ds *datasetService) UpdateDocuments(ctx context.Context, id string, updateRequests []map[string]any) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Update dataset documents")
	defer span.End()

	// 调用 dataset access 批量更新文档
	if err := ds.da.UpdateDocuments(ctx, id, updateRequests); err != nil {
		span.SetStatus(codes.Error, "Update dataset documents failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_UpdateFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// DeleteDocuments 批量删除 dataset 文档
func (ds *datasetService) DeleteDocuments(ctx context.Context, id string, docIDs string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete dataset documents")
	defer span.End()

	// 调用 dataset access 批量删除文档
	if err := ds.da.DeleteDocuments(ctx, id, docIDs); err != nil {
		span.SetStatus(codes.Error, "Delete dataset documents failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_DeleteFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}
