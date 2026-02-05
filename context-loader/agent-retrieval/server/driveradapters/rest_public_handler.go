// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/drivenadapters"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/driveradapters/knretrieval"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/driveradapters/mcp"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

type restPublicHandler struct {
	Hydra              interfaces.Hydra
	KnRetrievalHandler knretrieval.KnRetrievalHandler
	MCPHandler         http.Handler
	Logger             interfaces.Logger
}

// NewRestPublicHandler 创建restHandler实例
func NewRestPublicHandler(logger interfaces.Logger) interfaces.HTTPRouterInterface {
	return &restPublicHandler{
		Hydra:              drivenadapters.NewHydra(),
		KnRetrievalHandler: knretrieval.NewKnRetrievalHandler(),
		MCPHandler:         mcp.NewMCPHandler(),
		Logger:             logger,
	}
}

// RegisterPublic 注册公共路由
func (r *restPublicHandler) RegisterRouter(engine *gin.RouterGroup) {
	mws := []gin.HandlerFunc{}
	mws = append(mws, middlewareRequestLog(r.Logger), middlewareTrace, middlewareIntrospectVerify(r.Hydra))
	engine.Use(mws...)

	engine.POST("/kn/semantic-search", r.KnRetrievalHandler.SemanticSearch)

	// MCP Server (Bearer token auth, supports Cursor/Claude Desktop)
	engine.Any("/mcp/*path", gin.WrapH(r.MCPHandler))
}
