// Package driveradapters provides HTTP handlers (primary adapters).
package driveradapters

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	libCommon "github.com/kweaver-ai/kweaver-go-lib/common"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/middleware"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"vega-backend/common"
	"vega-backend/interfaces"
	"vega-backend/logics/catalog"
	connectortype "vega-backend/logics/connector_type"
	discoverytask "vega-backend/logics/discovery_task"
	"vega-backend/logics/resource"
	"vega-backend/version"
)

// RestHandler interface
type RestHandler interface {
	RegisterPublic(engine *gin.Engine)
}

type restHandler struct {
	appSetting *common.AppSetting
	cs         interfaces.CatalogService
	rs         interfaces.ResourceService
	cts        interfaces.ConnectorTypeService
	dts        interfaces.DiscoveryTaskService // 任务服务
}

// NewRestHandler creates a new RestHandler.
func NewRestHandler(appSetting *common.AppSetting) RestHandler {
	cs := catalog.NewCatalogService(appSetting)
	rs := resource.NewResourceService(appSetting)
	return &restHandler{
		appSetting: appSetting,
		cs:         cs,
		rs:         rs,
		cts:        connectortype.NewConnectorTypeService(appSetting),
		dts:        discoverytask.NewDiscoveryTaskService(appSetting),
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

		// DiscoveryTask APIs
		discoveryTasks := apiV1.Group("/discovery-tasks")
		{
			discoveryTasks.GET("", r.ListDiscoveryTasks)
			discoveryTasks.GET("/:id", r.GetDiscoveryTask)
		}
	}

	logger.Info("RestHandler RegisterPublic")
}

// HealthCheck handles GET /health
func (r *restHandler) HealthCheck(c *gin.Context) {
	serverInfo := struct {
		ServerName    string `json:"server_name"`
		ServerVersion string `json:"server_version"`
		Language      string `json:"language"`
		GoVersion     string `json:"go_version"`
		GoArch        string `json:"go_arch"`
		Status        string `json:"status"`
	}{
		ServerName:    version.ServerName,
		ServerVersion: version.ServerVersion,
		Language:      version.LanguageGo,
		GoVersion:     version.GoVersion,
		GoArch:        version.GoArch,
		Status:        "healthy",
	}
	rest.ReplyOK(c, http.StatusOK, serverInfo)
}

// verifyJsonContentType middleware
func (r *restHandler) verifyJsonContentType() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.ContentType() != interfaces.CONTENT_TYPE_JSON {
			httpErr := rest.NewHTTPError(c, http.StatusNotAcceptable, "VegaManager.InvalidRequestHeader.ContentType")
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

// generateAccountInfo extracts account info from request headers
func (r *restHandler) generateAccountInfo(c *gin.Context) interfaces.AccountInfo {
	return interfaces.AccountInfo{
		ID:   c.GetHeader(interfaces.HTTP_HEADER_ACCOUNT_ID),
		Type: c.GetHeader(interfaces.HTTP_HEADER_ACCOUNT_TYPE),
	}
}

// Health handles GET /health
func (h *restHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "vega-backend",
	})
}
