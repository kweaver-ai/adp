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
	"github.com/kweaver-ai/kweaver-go-lib/audit"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	verrors "vega-backend/errors"
	"vega-backend/interfaces"
)

// ListCatalogs handles GET /api/vega-backend/v1/catalogs
func (r *restHandler) ListCatalogs(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"ListCatalogs", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 获取查询参数
	tag := strings.TrimSpace(c.Query("tag"))
	typ := c.Query("type")
	healthCheckStatus := c.Query("health_check_status")
	offset := c.DefaultQuery("offset", interfaces.DEFAULT_OFFSET)
	limit := c.DefaultQuery("limit", interfaces.DEFAULT_LIMIT)
	sort := c.DefaultQuery("sort", "update_time")
	direction := c.DefaultQuery("direction", interfaces.DESC_DIRECTION)

	// 校验分页查询参数
	pageParam, err := validatePaginationQueryParams(ctx,
		offset, limit, sort, direction, interfaces.CATALOG_SORT)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	params := interfaces.CatalogsQueryParams{
		PaginationQueryParams: pageParam,
		Tag:                   tag,
		Type:                  typ,
		HealthCheckStatus:     healthCheckStatus,
	}

	entries, total, err := r.cs.List(ctx, params)
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

	logger.Debug("Handler ListCatalogs Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// CreateCatalog handles POST /api/vega-backend/v1/catalogs
func (r *restHandler) CreateCatalog(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"CreateCatalog", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置 trace 的相关 api 的属性
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	var req interfaces.CatalogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if err := ValidateCatalogRequest(ctx, &req); err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Check if name exists
	exists, err := r.cs.CheckExistByName(ctx, req.Name)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if exists {
		httpErr := rest.NewHTTPError(ctx, http.StatusConflict, verrors.VegaBackend_Catalog_NameExists)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	id, err := r.cs.Create(ctx, &req)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 成功创建记录审计日志
	audit.NewInfoLog(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
		interfaces.GenerateCatalogAuditObject(id, req.Name), "")

	result := map[string]any{"id": id}

	logger.Debug("Handler CreateCatalog Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusCreated, result)
}

// GetCatalogs handles GET /api/vega-backend/v1/catalogs/:ids
func (r *restHandler) GetCatalogs(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"GetCatalogs", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	ids := strings.Split(c.Param("ids"), ",")

	catalogs, err := r.cs.GetByIDs(ctx, ids)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if len(catalogs) != len(ids) {
		for _, id := range ids {
			found := false
			for _, catalog := range catalogs {
				if catalog.ID == id {
					found = true
					break
				}
			}
			if !found {
				httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Catalog_NotFound).
					WithErrorDetails(fmt.Sprintf("id %s not found", id))
				o11y.AddHttpAttrs4HttpError(span, httpErr)
				rest.ReplyError(c, httpErr)
				return
			}
		}
	}

	result := map[string]any{"entries": catalogs}

	logger.Debug("Handler GetCatalogs Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// UpdateCatalog handles PUT /api/vega-backend/v1/catalogs/:id
func (r *restHandler) UpdateCatalog(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"UpdateCatalog", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	id := c.Param("id")

	var req interfaces.CatalogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if err := ValidateCatalogRequest(ctx, &req); err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Check if id exists
	catalog, err := r.cs.GetByID(ctx, id, false)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	req.OriginCatalog = catalog

	// Apply updates
	if req.Name != catalog.Name {
		exists, err := r.cs.CheckExistByName(ctx, req.Name)
		if err != nil {
			httpErr := err.(*rest.HTTPError)
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
		}
		if exists {
			span.SetStatus(codes.Error, "Catalog name exists")
			httpErr := rest.NewHTTPError(ctx, http.StatusConflict, verrors.VegaBackend_Catalog_NameExists)
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		req.IfNameModify = true
	}

	if err := r.cs.Update(ctx, id, &req); err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	audit.NewInfoLog(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
		interfaces.GenerateCatalogAuditObject(id, req.Name), "")

	logger.Debug("Handler UpdateCatalog Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// DeleteCatalog handles DELETE /api/vega-backend/v1/catalogs/:ids
func (r *restHandler) DeleteCatalogs(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"DeleteCatalogs", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	ids := strings.Split(c.Param("ids"), ",")

	// Check if ids exists
	for _, id := range ids {
		exists, err := r.cs.CheckExistByID(ctx, id)
		if err != nil {
			httpErr := err.(*rest.HTTPError)
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		if !exists {
			httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Catalog_NotFound).
				WithErrorDetails(fmt.Sprintf("id %s not found", id))
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}

	if err := r.cs.DeleteByIDs(ctx, ids); err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	for _, id := range ids {
		audit.NewWarnLog(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			interfaces.GenerateCatalogAuditObject(id, ""), audit.SUCCESS, "")
	}

	logger.Debug("Handler DeleteCatalog Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// GetCatalogHealthStatus handles GET /api/vega-backend/v1/catalogs/:ids/health_status
func (r *restHandler) GetCatalogHealthStatus(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"GetCatalogHealthStatus", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	id := c.Param("ids")

	catalog, err := r.cs.GetByID(ctx, id, false)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	result := map[string]any{
		"id":                  catalog.ID,
		"health_check_status": catalog.HealthCheckStatus,
		"last_check_time":     catalog.LastCheckTime,
		"health_check_result": catalog.HealthCheckResult,
	}

	logger.Debug("Handler GetCatalogsHealthStatus Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// TestConnection handles POST /api/vega-backend/v1/catalogs/:id/test-connection
func (r *restHandler) TestConnection(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"TestConnection", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	id := c.Param("id")

	// Check if id exists
	catalog, err := r.cs.GetByID(ctx, id, false)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	result, err := r.cs.TestConnection(ctx, catalog)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler TestConnection Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// DiscoverCatalogResources handles POST /api/vega-backend/v1/catalogs/:id/discover
// 触发异步扫描任务，返回任务信息
func (r *restHandler) DiscoverCatalogResources(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"DiscoverCatalogResources", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	id := c.Param("id")

	// Get catalog to verify it exists
	catalog, err := r.cs.GetByID(ctx, id, false)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if catalog == nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Catalog_NotFound)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Create discover task (async)
	taskID, err := r.dts.Create(ctx, catalog.ID)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	result := map[string]any{
		"id": taskID,
	}

	logger.Debug("Handler DiscoverCatalogResources Success - Task Created")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// ListCatalogResources handles GET /api/vega-backend/v1/catalogs/:id/resources
func (r *restHandler) ListCatalogResources(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"ListCatalogResources", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	id := c.Param("ids")

	// Check if id exists
	exists, err := r.cs.CheckExistByID(ctx, id)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if !exists {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Catalog_NotFound)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 获取查询参数
	category := c.Query("category")
	status := c.Query("status")
	database := c.Query("database")
	offset := c.DefaultQuery("offset", interfaces.DEFAULT_OFFSET)
	limit := c.DefaultQuery("limit", interfaces.DEFAULT_LIMIT)
	sort := c.DefaultQuery("sort", "update_time")
	direction := c.DefaultQuery("direction", interfaces.DESC_DIRECTION)

	// 校验分页查询参数
	pageParam, err := validatePaginationQueryParams(ctx,
		offset, limit, sort, direction, interfaces.CATALOG_SORT)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	params := interfaces.ResourcesQueryParams{
		PaginationQueryParams: pageParam,
		CatalogID:             id,
		Category:              category,
		Status:                status,
		Database:              database,
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

	logger.Debug("Handler ListCatalogResources Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}
