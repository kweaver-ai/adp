// Package driveradapters provides HTTP handlers (primary adapters).
package driveradapters

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/trace"

	oerrors "vega-backend/errors"
	"vega-backend/interfaces"
)

// ListConnectorTypes handles GET /api/vega-backend/v1/connector-types
func (r *restHandler) ListConnectorTypes(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"ListConnectorTypes", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := r.generateAccountInfo(c)
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	var enabled *bool
	if enabledStr := c.Query("enabled"); enabledStr != "" {
		b, err := strconv.ParseBool(enabledStr)
		if err == nil {
			enabled = &b
		}
	}

	params := interfaces.ConnectorTypesQueryParams{
		PaginationParams: interfaces.PaginationParams{
			Offset: getIntQuery(c, "offset", interfaces.DefaultOffset),
			Limit:  getIntQuery(c, "limit", interfaces.DefaultLimit),
		},
		Mode:     c.Query("mode"),
		Category: c.Query("category"),
		Enabled:  enabled,
	}

	entries, total, err := r.cts.List(ctx, params)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	result := map[string]any{
		"entries": entries,
		"total":   total,
	}

	logger.Debug("Handler ListConnectorTypes Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// CreateConnectorType handles POST /api/vega-backend/v1/connector-types
func (r *restHandler) RegisterConnectorType(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"RegisterConnectorType", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := r.generateAccountInfo(c)
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置 trace 的相关 api 的属性
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	var req interfaces.ConnectorTypeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.VegaManager_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())
		rest.ReplyError(c, httpErr)
		return
	}

	// Check if type exists
	exists, err := r.cts.CheckExistByType(ctx, req.Type)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_ConnectorType_InternalError).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if exists {
		httpErr := rest.NewHTTPError(ctx, http.StatusConflict, oerrors.VegaManager_ConnectorType_TypeExists)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if err := ValidateConnectorTypeReq(ctx, &req); err != nil {
		rest.ReplyError(c, err)
		return
	}

	if err := r.cts.Register(ctx, &req); err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	result := map[string]any{"type": req.Type}

	logger.Debug("Handler CreateCatalog Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusCreated, result)
}

// GetConnectorType handles GET /api/vega-backend/v1/connector-types/:id
func (r *restHandler) GetConnectorType(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"GetConnectorType", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := r.generateAccountInfo(c)
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	id := c.Param("id")

	connectorType, err := r.cts.GetByType(ctx, id)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler GetConnectorType Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, connectorType)
}

// UpdateConnectorType handles PUT /api/vega-backend/v1/connector-types/:type
func (r *restHandler) UpdateConnectorType(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"UpdateConnectorType", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := r.generateAccountInfo(c)
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	tp := c.Param("type")

	var req interfaces.ConnectorTypeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.VegaManager_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())
		rest.ReplyError(c, httpErr)
		return
	}
	req.Type = tp

	if err := ValidateConnectorTypeReq(ctx, &req); err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if _, err := r.cts.GetByType(ctx, tp); err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if err := r.cts.Update(ctx, &req); err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler UpdateConnectorType Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// DeleteConnectorType handles DELETE /api/vega-backend/v1/connector-types/:type
func (r *restHandler) DeleteConnectorType(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"DeleteConnectorType", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := r.generateAccountInfo(c)
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	tp := c.Param("type")

	ct, err := r.cts.GetByType(ctx, tp)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_ConnectorType_InternalError).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if ct.Mode == interfaces.ConnectorModeLocal {
		httpErr := rest.NewHTTPError(ctx, http.StatusForbidden, oerrors.VegaManager_ConnectorType_BadRequest).
			WithErrorDetails("can not delete local connector type")
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if err := r.cts.DeleteByType(ctx, tp); err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler DeleteConnectorType Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// SetConnectorTypeEnabled handles POST /api/vega-backend/v1/connector-types/:type/enable
func (r *restHandler) SetConnectorTypeEnabled(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"SetConnectorTypeEnabled", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := r.generateAccountInfo(c)
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	tp := c.Param("type")

	exists, err := r.cts.CheckExistByType(ctx, tp)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_ConnectorType_InternalError).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if !exists {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, oerrors.VegaManager_ConnectorType_NotFound)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	var req struct {
		Value bool `json:"value"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.VegaManager_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())
		rest.ReplyError(c, httpErr)
		return
	}

	if err := r.cts.SetEnabled(ctx, tp, req.Value); err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, err)
		return
	}

	logger.Debug("Handler SetConnectorTypeEnabled Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}
