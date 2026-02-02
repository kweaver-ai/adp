// Package driveradapters provides HTTP handlers.
package driveradapters

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/trace"

	oerrors "vega-manager/errors"
	"vega-manager/interfaces"
)

// GetDiscoveryTask handles GET /api/vega-manager/v1/catalogs/:id/discover/:taskId
func (r *restHandler) GetDiscoveryTask(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"GetDiscoveryTask", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := r.generateAccountInfo(c)
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	catalogID := c.Param("id")
	taskID := c.Param("taskId")

	// Verify catalog exists
	catalog, err := r.cs.GetByID(ctx, catalogID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if catalog == nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, oerrors.VegaManager_Catalog_NotFound)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Get task
	task, err := r.dts.GetByID(ctx, taskID)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_Catalog_InternalError).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if task == nil || task.CatalogID != catalogID {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, oerrors.VegaManager_Task_NotFound)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler GetDiscoveryTask Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, task)
}

// ListDiscoveryTasks handles GET /api/vega-manager/v1/catalogs/:id/discover/tasks
func (r *restHandler) ListDiscoveryTasks(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"ListDiscoveryTasks", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := r.generateAccountInfo(c)
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	catalogID := c.Param("id")

	// Verify catalog exists
	catalog, err := r.cs.GetByID(ctx, catalogID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if catalog == nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, oerrors.VegaManager_Catalog_NotFound)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Parse pagination params
	params := interfaces.PaginationParams{}
	if err := c.ShouldBindQuery(&params); err == nil {
		if params.Limit == 0 {
			params.Limit = 10
		}
	}

	// List tasks
	tasks, total, err := r.dts.List(ctx, catalogID, params)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.VegaManager_Catalog_InternalError).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler ListDiscoveryTasks Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, gin.H{
		"items": tasks,
		"total": total,
	})
}
