package driveradapters

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/driveradapters/skill"
)

type SkillRestHandler interface {
	// RegisterPrivate 注册内部API
	RegisterPrivate(engine *gin.RouterGroup)
	// RegisterPublic 注册公开API
	RegisterPublic(engine *gin.RouterGroup)
}

type skillRestHandler struct {
	SkillHandler skill.SkillHandler
}

var (
	sOnce    sync.Once
	sHandler SkillRestHandler
)

func NewSkillRestHandler() SkillRestHandler {
	sOnce.Do(func() {
		sHandler = &skillRestHandler{
			SkillHandler: skill.NewSkillHandler(),
		}
	})
	return sHandler
}
func (r *skillRestHandler) RegisterPrivate(engine *gin.RouterGroup) {
	/*市场接口*/
	// 查询技能市场列表
	engine.GET("/skills/market", middlewareBusinessDomain(true, false), r.SkillHandler.QuerySkillMarketList)
	// 查询技能市场详情
	engine.GET("/skills/market/:skill_id", middlewareBusinessDomain(true, false), r.SkillHandler.GetSkillMarketDetail)
	/*读取接口*/
	// 查询技能内容
	engine.GET("/skills/:skill_id/content", middlewareBusinessDomain(true, false), r.SkillHandler.GetSkillContent)
	// 读取技能文件
	engine.POST("/skills/:skill_id/files/read", middlewareBusinessDomain(true, false), r.SkillHandler.ReadSkillFile)
}

func (r *skillRestHandler) RegisterPublic(engine *gin.RouterGroup) {
	/*管理接口*/
	// 注册技能
	engine.POST("/skills", middlewareBusinessDomain(true, false), r.SkillHandler.RegisterSkill)
	// 查询技能列表
	engine.GET("/skills", middlewareBusinessDomain(true, false), r.SkillHandler.QuerySkillList)
	// 查询技能详情
	engine.GET("/skills/:skill_id", middlewareBusinessDomain(true, false), r.SkillHandler.GetSkillDetail)
	// 下载技能
	engine.GET("/skills/:skill_id/download", middlewareBusinessDomain(true, false), r.SkillHandler.DownloadSkill)
	// 删除技能
	engine.DELETE("/skills/:skill_id", middlewareBusinessDomain(true, false), r.SkillHandler.DeleteSkill)
	// 更新状态
	engine.PUT("/skills/:skill_id/status", middlewareBusinessDomain(true, false), r.SkillHandler.UpdateSkillStatus)
	/*市场接口*/
	// 查询技能市场列表
	engine.GET("/skills/market", middlewareBusinessDomain(true, false), r.SkillHandler.QuerySkillMarketList)
	// 查询技能市场详情
	engine.GET("/skills/market/:skill_id", middlewareBusinessDomain(true, false), r.SkillHandler.GetSkillMarketDetail)
	/*读取接口*/
	// 查询技能内容
	engine.GET("/skills/:skill_id/content", middlewareBusinessDomain(true, false), r.SkillHandler.GetSkillContent)
	// 读取技能文件
	engine.POST("/skills/:skill_id/files/read", middlewareBusinessDomain(true, false), r.SkillHandler.ReadSkillFile)
}
