// Package driveradapters 定义驱动适配器
// @file rest_public_handler.go
// @description: 定义rest公共适配器
package driveradapters

import (
	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/drivenadapters"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/driveradapters/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
)

type restPublicHandler struct {
	Hydra               interfaces.Hydra
	OperatorRestHandler OperatorRestHandler
	ToolBoxRestHandler  ToolBoxRestHandler
	MCPRestHandler      MCPRestHandler
	ImpexHandler        common.ImpexHandler
	UnifiedProxyHandler common.UnifiedProxyHandler
	TemplateHandler     common.TemplateHandler
	AIGenerationHandler common.AIGenerationHandler
	Logger              interfaces.Logger
}

// NewRestPublicHandler 创建restHandler实例
func NewRestPublicHandler() interfaces.HTTPRouterInterface {
	return &restPublicHandler{
		Hydra:               drivenadapters.NewHydra(),
		OperatorRestHandler: NewOperatorRestHandler(),
		ToolBoxRestHandler:  NewToolBoxRestHandler(),
		MCPRestHandler:      NewMCPRestHandler(),
		ImpexHandler:        common.NewImpexHandler(),
		UnifiedProxyHandler: common.NewUnifiedProxyHandler(),
		TemplateHandler:     common.NewTemplateHandler(),
		AIGenerationHandler: common.NewAIGenerationHandler(),
		Logger:              config.NewConfigLoader().GetLogger(),
	}
}

// RegisterPublic 注册公共路由
func (r *restPublicHandler) RegisterRouter(engine *gin.RouterGroup) {
	mws := []gin.HandlerFunc{}
	mws = append(mws, middlewareRequestLog(r.Logger), middlewareTrace, middlewareIntrospectVerify(r.Hydra))
	engine.Use(mws...)
	// 算子注册相关接口
	r.OperatorRestHandler.RegisterPublic(engine)
	// 工具箱相关接口
	r.ToolBoxRestHandler.RegisterPublic(engine)
	// MCP 相关接口
	r.MCPRestHandler.RegisterPublic(engine)
	// 导入导出
	engine.GET("/impex/export/:type/:id", r.ImpexHandler.Export)
	engine.POST("/impex/import/:type", middlewareBusinessDomain(true, false), r.ImpexHandler.Import)
	// 函数执行
	engine.POST("/function/execute", middlewareBusinessDomain(true, false), r.UnifiedProxyHandler.FunctionExecute)
	// 获取Python模板
	engine.GET("/template/:template_type", middlewareBusinessDomain(true, false), r.TemplateHandler.GetTemplate)
	// AI辅助生成
	engine.POST("/ai_generate/function/:type", middlewareBusinessDomain(true, false), r.AIGenerationHandler.FunctionAIGeneration)
	// 获取提示词模板
	engine.GET("/ai_generate/prompt/:type", middlewareBusinessDomain(true, false), r.AIGenerationHandler.GetPromptTemplate)
}
