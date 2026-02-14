// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package driveradapters provides HTTP handlers (primary adapters).
package driveradapters

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	libCommon "github.com/kweaver-ai/kweaver-go-lib/common"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/middleware"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"vega-backend/common"
	"vega-backend/interfaces"
	"vega-backend/logics/auth"
	"vega-backend/logics/catalog"
	"vega-backend/logics/connector_type"
	"vega-backend/logics/dataset"
	"vega-backend/logics/discover_task"
	"vega-backend/logics/resource"
	"vega-backend/logics/resource_data"
	"vega-backend/version"
)

// RestHandler interface
type RestHandler interface {
	RegisterPublic(engine *gin.Engine)
}

type restHandler struct {
	appSetting *common.AppSetting
	as         interfaces.AuthService
	cs         interfaces.CatalogService
	rs         interfaces.ResourceService
	ds         interfaces.DatasetService
	cts        interfaces.ConnectorTypeService
	dts        interfaces.DiscoverTaskService
	rds        interfaces.ResourceDataService
}

// NewRestHandler creates a new RestHandler.
func NewRestHandler(appSetting *common.AppSetting) RestHandler {
	cs := catalog.NewCatalogService(appSetting)
	rs := resource.NewResourceService(appSetting)
	ds := dataset.NewDatasetService(appSetting)
	return &restHandler{
		appSetting: appSetting,
		as:         auth.NewAuthService(appSetting),
		cs:         cs,
		rs:         rs,
		ds:         ds,
		cts:        connector_type.NewConnectorTypeService(appSetting),
		dts:        discover_task.NewDiscoverTaskService(appSetting),
		rds:        resource_data.NewResourceDataService(appSetting),
	}
}

// RegisterPublic registers public API routes.
func (r *restHandler) RegisterPublic(engine *gin.Engine) {
	engine.Use(r.accessLog())
	engine.Use(middleware.TracingMiddleware())

	engine.GET("/health", r.HealthCheck)

	apiV1 := engine.Group("/api/vega-backend/v1")
	{
		// Catalog APIs
		catalogs := apiV1.Group("/catalogs")
		{
			catalogs.GET("", r.ListCatalogs)
			catalogs.POST("", r.verifyJsonContentType(), r.CreateCatalog)
			catalogs.GET("/:ids", r.GetCatalogs)
			catalogs.PUT("/:id", r.verifyJsonContentType(), r.UpdateCatalog)
			catalogs.DELETE("/:ids", r.DeleteCatalogs)
			catalogs.GET("/:ids/health-status", r.GetCatalogHealthStatus)
			catalogs.POST("/:id/test-connection", r.TestConnection)
			catalogs.POST("/:id/discover", r.DiscoverCatalogResources)
			catalogs.GET("/:ids/resources", r.ListCatalogResources)
		}

		// Resource APIs
		resources := apiV1.Group("/resources")
		{
			resources.GET("", r.ListResources)
			resources.POST("", r.verifyJsonContentType(), r.CreateResource)
			resources.GET("/:ids", r.GetResources)
			resources.PUT("/:id", r.verifyJsonContentType(), r.UpdateResource)
			resources.DELETE("/:ids", r.DeleteResources)

			resources.POST("/:id/data", r.verifyJsonContentType(), r.QueryResourceData) // method override GET list and get

			resources.POST("/dataset/:id/docs", r.verifyJsonContentType(), r.CreateDatasetDocuments)
			resources.PUT("/dataset/:id/docs", r.verifyJsonContentType(), r.UpdateDatasetDocuments)
			resources.DELETE("/dataset/:id/docs/:ids", r.DeleteDatasetDocuments)
		}

		// ConnectorType APIs
		connectorTypes := apiV1.Group("/connector-types")
		{
			connectorTypes.GET("", r.ListConnectorTypes)
			connectorTypes.POST("", r.verifyJsonContentType(), r.RegisterConnectorType)
			connectorTypes.GET("/:type", r.GetConnectorType)
			connectorTypes.PUT("/:type", r.verifyJsonContentType(), r.UpdateConnectorType)
			connectorTypes.DELETE("/:type", r.DeleteConnectorType)
			connectorTypes.POST("/:type/enabled", r.SetConnectorTypeEnabled)
		}

		// DiscoverTask APIs
		discoverTasks := apiV1.Group("/discover-tasks")
		{
			discoverTasks.GET("", r.ListDiscoverTasks)
			discoverTasks.GET("/:id", r.GetDiscoverTask)
		}
	}

	logger.Info("RestHandler RegisterPublic")
}

// HealthCheck 健康检查
func (r *restHandler) HealthCheck(c *gin.Context) {
	// 返回服务信息
	serverInfo := o11y.ServerInfo{
		ServerName:    version.ServerName,
		ServerVersion: version.ServerVersion,
		Language:      version.LanguageGo,
		GoVersion:     version.GoVersion,
		GoArch:        version.GoArch,
	}
	rest.ReplyOK(c, http.StatusOK, serverInfo)
}

// verifyJsonContentType middleware
func (r *restHandler) verifyJsonContentType() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.ContentType() != interfaces.CONTENT_TYPE_JSON {
			httpErr := rest.NewHTTPError(c, http.StatusNotAcceptable, "VegaBackend.InvalidRequestHeader.ContentType")
			rest.ReplyError(c, httpErr)
			c.Abort()
			return
		}
		c.Next()
	}
}

// accessLog middleware
func (r *restHandler) accessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		beginTime := time.Now()
		c.Next()
		endTime := time.Now()
		durTime := endTime.Sub(beginTime).Seconds()

		logger.Debugf("access log: url: %s, method: %s, begin_time: %s, end_time: %s, subTime: %f",
			c.Request.URL.Path,
			c.Request.Method,
			beginTime.Format(libCommon.RFC3339Milli),
			endTime.Format(libCommon.RFC3339Milli),
			durTime,
		)
	}
}

// verifyOAuth verifies OAuth token
func (r *restHandler) verifyOAuth(ctx context.Context, c *gin.Context) (hydra.Visitor, error) {
	visitor, err := r.as.VerifyToken(ctx, c)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusUnauthorized, rest.PublicError_Unauthorized).
			WithErrorDetails(err.Error())
		rest.ReplyError(c, httpErr)
		return visitor, err
	}

	return visitor, nil
}
