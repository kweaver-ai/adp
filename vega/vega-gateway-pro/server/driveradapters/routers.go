package driveradapters

import (
	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/middleware"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"net/http"
	"vega-gateway-pro/common"
	"vega-gateway-pro/interfaces"
	"vega-gateway-pro/logics/fetch"
	"vega-gateway-pro/version"
)

type RestHandler interface {
	RegisterPublic(engine *gin.Engine)
}

type restHandler struct {
	fetchService interfaces.FetchService
	hydra        rest.Hydra
}

func NewRestHandler(appSetting *common.AppSetting) RestHandler {
	r := &restHandler{
		fetchService: fetch.NewFetchService(appSetting),
		hydra:        rest.NewHydra(appSetting.HydraAdminSetting),
	}
	return r
}

func (r *restHandler) RegisterPublic(c *gin.Engine) {
	c.Use(middleware.TracingMiddleware())

	c.GET("/health", r.HealthCheck)

	exApiV1 := c.Group("/api/vega-gateway/v2", r.verifyOAuthMiddleWare())
	{
		// 数据查询接口
		exApiV1.POST("/fetch", r.FetchQuery)
		// 批次查询接口
		exApiV1.GET("/fetch/:query_id/:slug/:token", r.NextQuery)

	}

	inApiV1 := c.Group("/api/internal/vega-gateway/v2")
	{
		// 数据查询接口
		inApiV1.POST("/fetch", r.FetchQuery)
		// 批次查询接口
		inApiV1.GET("/fetch/:query_id/:slug/:token", r.NextQuery)

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

// gin中间件 校验oauth
func (r *restHandler) verifyOAuthMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := rest.GetLanguageCtx(c)
		_, err := r.hydra.VerifyToken(ctx, c)
		if err != nil {
			httpError := rest.NewHTTPError(ctx, http.StatusUnauthorized, rest.PublicError_Unauthorized).
				WithErrorDetails(err.Error())
			rest.ReplyError(c, httpError)
			c.Abort()
			return
		}

		//执行后续操作
		c.Next()
	}
}
