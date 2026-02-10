// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package driveradapters provides HTTP handlers (primary adapters).
package driveradapters

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/trace"

	oerrors "vega-backend/errors"
	"vega-backend/interfaces"
)

// ListResources handles GET /api/vega-backend/v1/resources
func (r *restHandler) ListResources(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"ListResources", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := r.generateAccountInfo(c)
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	params := interfaces.ResourcesQueryParams{
		PaginationParams: interfaces.PaginationParams{
			Offset: getIntQuery(c, "offset", interfaces.DefaultOffset),
			Limit:  getIntQuery(c, "limit", interfaces.DefaultLimit),
		},
		CatalogID: c.Query("catalog_id"),
		Category:  c.Query("category"),
		Status:    c.Query("status"),
		Database:  c.Query("database"),
	}

	entries, total, err := r.rs.List(ctx, params)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	result := map[string]any{
		"entries":     entries,
		"total_count": total,
	}

	logger.Debug("Handler ListResources Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// CreateResource handles POST /api/vega-backend/v1/resources
func (r *restHandler) CreateResource(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"CreateResource", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := r.generateAccountInfo(c)
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置 trace 的相关 api 的属性
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	var req interfaces.ResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest,
			oerrors.VegaManager_InvalidParameter_RequestBody).WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if err := ValidateResourceRequest(ctx, &req); err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
	}

	// Check if name exists
	exists, err := r.rs.CheckExistByName(ctx, req.CatalogID, req.Name)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.VegaManager_Resource_InternalError).WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if exists {
		httpErr := rest.NewHTTPError(ctx, http.StatusConflict, oerrors.VegaManager_Resource_NameExists)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	id, err := r.rs.Create(ctx, &req)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	result := map[string]any{"id": id}

	logger.Debug("Handler CreateResource Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusCreated, result)
}

// GetResource handles GET /api/vega-backend/v1/resources/:ids
func (r *restHandler) GetResources(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"GetResources", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := r.generateAccountInfo(c)
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	ids := strings.Split(c.Param("ids"), ",")

	resources, err := r.rs.GetByIDs(ctx, ids)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if len(resources) != len(ids) {
		for _, id := range ids {
			found := false
			for _, resource := range resources {
				if resource.ID == id {
					found = true
					break
				}
			}
			if !found {
				httpErr := rest.NewHTTPError(ctx, http.StatusNotFound,
					oerrors.VegaManager_Resource_NotFound).WithErrorDetails(fmt.Sprintf("id %s not found", id))
				o11y.AddHttpAttrs4HttpError(span, httpErr)
				rest.ReplyError(c, httpErr)
				return
			}
		}
	}
	result := map[string]any{"entries": resources}

	logger.Debug("Handler GetResource Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// UpdateResource handles PUT /api/vega-backend/v1/resources/:id
func (r *restHandler) UpdateResource(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"UpdateResource", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := r.generateAccountInfo(c)
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	id := c.Param("id")

	var req interfaces.ResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest,
			oerrors.VegaManager_InvalidParameter_RequestBody).WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if err := ValidateResourceRequest(ctx, &req); err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
	}

	// Check if id exists
	resource, err := r.rs.GetByID(ctx, id)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	req.OriginResource = resource

	if err := r.rs.Update(ctx, id, &req); err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler UpdateResource Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// DeleteResource handles DELETE /api/vega-backend/v1/resources/:ids
func (r *restHandler) DeleteResources(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"DeleteResources", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := r.generateAccountInfo(c)
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	ids := strings.Split(c.Param("ids"), ",")

	// Check if ids exists
	for _, id := range ids {
		exists, err := r.rs.CheckExistByID(ctx, id)
		if err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError,
				oerrors.VegaManager_Resource_InternalError).WithErrorDetails(err.Error())
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		if !exists {
			httpErr := rest.NewHTTPError(ctx, http.StatusNotFound,
				oerrors.VegaManager_Resource_NotFound).WithErrorDetails(fmt.Sprintf("id %s not found", id))
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}

	if err := r.rs.DeleteByIDs(ctx, ids); err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler DeleteResource Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// ListDatasetDocuments handles GET /api/vega-backend/v1/resources/dataset/:id
func (r *restHandler) ListDatasetDocuments(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"ListDatasetDocuments", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := r.generateAccountInfo(c)
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	datasetID := c.Param("id")

	// 解析请求体
	var params interfaces.QueryParams
	if err := c.ShouldBindJSON(&params); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.VegaManager_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 调用 dataset 服务列出文档
	documents, total, err := r.ds.ListDocuments(ctx, datasetID, &params)
	if err != nil {
		if httpErr, ok := err.(*rest.HTTPError); ok {
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
		} else {
			// 处理其他类型的错误
			httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_InternalError_MarshalDataFailed).
				WithErrorDetails(err.Error())
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
		}
		return
	}

	resultData := map[string]any{
		"entries":     documents,
		"total_count": total,
	}

	logger.Debug("Handler ListDatasetDocuments Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, resultData)
}

// CreateDatasetDocuments handles POST /api/vega-backend/v1/resources/dataset/:id/docs
func (r *restHandler) CreateDatasetDocuments(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"CreateDatasetDocuments", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := r.generateAccountInfo(c)
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	datasetID := c.Param("id")

	// 获取请求体
	var documents []map[string]any
	if err := c.ShouldBindJSON(&documents); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.VegaManager_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 调用 dataset 服务批量创建文档
	docIDs, err := r.ds.CreateDocuments(ctx, datasetID, documents)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	resultData := map[string]any{"ids": docIDs}

	logger.Debug("Handler CreateDatasetDocuments Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusCreated)
	rest.ReplyOK(c, http.StatusCreated, resultData)
}

// GetDatasetDocument handles GET /api/vega-backend/v1/resources/dataset/:id/:docid
func (r *restHandler) GetDatasetDocument(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"GetDatasetDocument", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := r.generateAccountInfo(c)
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	datasetID := c.Param("id")
	docID := c.Param("docid")

	// 调用 dataset 服务获取文档
	document, err := r.ds.GetDocument(ctx, datasetID, docID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	resultData := map[string]any{"document": document}

	logger.Debug("Handler GetDatasetDocument Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, resultData)
}

// UpdateDatasetDocument handles PUT /api/vega-backend/v1/resources/dataset/:id/:docid
func (r *restHandler) UpdateDatasetDocument(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"UpdateDatasetDocument", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := r.generateAccountInfo(c)
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	datasetID := c.Param("id")
	docID := c.Param("docid")

	// 获取请求体
	var document map[string]any
	if err := c.ShouldBindJSON(&document); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.VegaManager_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 调用 dataset 服务更新文档
	if err := r.ds.UpdateDocument(ctx, datasetID, docID, document); err != nil {
		println(err.Error())
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler UpdateDatasetDocument Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}


// DeleteDatasetDocument handles DELETE /api/vega-backend/v1/resources/dataset/:id/:docid
func (r *restHandler) DeleteDatasetDocument(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"DeleteDatasetDocument", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := r.generateAccountInfo(c)
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	datasetID := c.Param("id")
	docID := c.Param("docid")

	// 调用 dataset 服务删除文档
	if err := r.ds.DeleteDocument(ctx, datasetID, docID); err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler DeleteDatasetDocument Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}


