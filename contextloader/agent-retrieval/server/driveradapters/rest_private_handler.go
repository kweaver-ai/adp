// Package driveradapters 定义驱动适配器
// @file rest_private_handler.go
// @description: 定义rest私有接口适配器
package driveradapters

import (
	"devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/driveradapters/knactionrecall"
	"devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/driveradapters/knlogicpropertyresolver"
	"devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/driveradapters/knqueryobjectinstance"
	"devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/driveradapters/knquerysubgraph"
	"devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/driveradapters/knretrieval"
	"devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/driveradapters/knsearch"
	"devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/interfaces"
	"github.com/gin-gonic/gin"
)

type restPrivateHandler struct {
	KnRetrievalHandler             knretrieval.KnRetrievalHandler
	KnLogicPropertyResolverHandler knlogicpropertyresolver.KnLogicPropertyResolverHandler
	KnActionRecallHandler          knactionrecall.KnActionRecallHandler
	KnQueryObjectInstanceHandler   knqueryobjectinstance.KnQueryObjectInstanceHandler
	KnQuerySubgraphHandler         knquerysubgraph.KnQuerySubgraphHandler
	KnSearchHandler                knsearch.KnSearchHandler
	Logger                         interfaces.Logger
}

// NewRestPrivateHandler 创建restHandler实例
func NewRestPrivateHandler(logger interfaces.Logger) interfaces.HTTPRouterInterface {
	return &restPrivateHandler{
		KnRetrievalHandler:             knretrieval.NewKnRetrievalHandler(),
		KnLogicPropertyResolverHandler: knlogicpropertyresolver.NewKnLogicPropertyResolverHandler(),
		KnActionRecallHandler:          knactionrecall.NewKnActionRecallHandler(),
		KnQueryObjectInstanceHandler:   knqueryobjectinstance.NewKnQueryObjectInstanceHandler(),
		KnQuerySubgraphHandler:         knquerysubgraph.NewKnQuerySubgraphHandler(),
		KnSearchHandler:                knsearch.NewKnSearchHandler(),
		Logger:                         logger,
	}
}

// RegisterRouter 注册路由
func (r *restPrivateHandler) RegisterRouter(engine *gin.RouterGroup) {
	mws := []gin.HandlerFunc{}
	mws = append(mws, middlewareRequestLog(r.Logger), middlewareTrace, middlewareHeaderAuthContext())
	engine.Use(mws...)

	engine.POST("/kn/semantic-search", r.KnRetrievalHandler.SemanticSearch)
	engine.POST("/kn/logic-property-resolver", r.KnLogicPropertyResolverHandler.ResolveLogicProperties)
	engine.POST("/kn/get_action_info", r.KnActionRecallHandler.GetActionInfo)
	engine.POST("/kn/query_object_instance", r.KnQueryObjectInstanceHandler.QueryObjectInstance)
	engine.POST("/kn/query_instance_subgraph", r.KnQuerySubgraphHandler.QueryInstanceSubgraph)
	engine.POST("/kn/kn_search", r.KnSearchHandler.KnSearch)
}
