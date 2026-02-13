// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"net/http"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/gin-gonic/gin"

	"data-model-job/common"
	"data-model-job/interfaces"
	"data-model-job/logics"
	"data-model-job/version"
)

type RestHandler interface {
	RegisterPublic(engine *gin.Engine)
}

type restHandler struct {
	appSetting *common.AppSetting
	jobService interfaces.JobService
}

func NewRestHandler(appSetting *common.AppSetting) RestHandler {
	return &restHandler{
		appSetting: appSetting,
		jobService: logics.NewJobService(appSetting),
	}
}

func (r *restHandler) RegisterPublic(c *gin.Engine) {

	c.GET("/health", r.HealthCheck)

	apiV1 := c.Group("/api/mdl-data-model-job/v1")
	{
		apiV1.POST("/jobs", r.StartJob)
		apiV1.PUT("/jobs/:id", r.UpdateJob)
		apiV1.DELETE("/jobs/:ids", r.StopJobs)

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

func GenerateVisitor(c *gin.Context) rest.Visitor {
	accountInfo := interfaces.AccountInfo{
		ID:   c.GetHeader(interfaces.HTTP_HEADER_ACCOUNT_ID),
		Type: c.GetHeader(interfaces.HTTP_HEADER_ACCOUNT_TYPE),
	}
	visitor := rest.Visitor{
		ID:         accountInfo.ID,
		Type:       rest.VisitorType(accountInfo.Type),
		TokenID:    "", // 无token
		IP:         c.ClientIP(),
		Mac:        c.GetHeader("X-Request-MAC"),
		UserAgent:  c.GetHeader("User-Agent"),
		ClientType: rest.ClientType_Linux,
	}
	return visitor
}
