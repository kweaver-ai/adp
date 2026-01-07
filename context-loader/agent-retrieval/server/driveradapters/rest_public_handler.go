// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/drivenadapters"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/driveradapters/knretrieval"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
	"github.com/gin-gonic/gin"
)

type restPublicHandler struct {
	Hydra              interfaces.Hydra
	KnRetrievalHandler knretrieval.KnRetrievalHandler
	Logger             interfaces.Logger
}

// NewRestPublicHandler 创建restHandler实例
func NewRestPublicHandler(logger interfaces.Logger) interfaces.HTTPRouterInterface {
	return &restPublicHandler{
		Hydra:              drivenadapters.NewHydra(),
		KnRetrievalHandler: knretrieval.NewKnRetrievalHandler(),
		Logger:             logger,
	}
}

// RegisterPublic 注册公共路由
func (r *restPublicHandler) RegisterRouter(engine *gin.RouterGroup) {
	mws := []gin.HandlerFunc{}
	mws = append(mws, middlewareRequestLog(r.Logger), middlewareTrace, middlewareIntrospectVerify(r.Hydra))
	engine.Use(mws...)

	engine.POST("/kn/semantic-search", r.KnRetrievalHandler.SemanticSearch)
}
