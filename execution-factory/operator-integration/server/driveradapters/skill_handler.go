package driveradapters

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/driveradapters/skill"
)

type SkillRestHandler interface {
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

func (r *skillRestHandler) RegisterPublic(engine *gin.RouterGroup) {
	engine.POST("/skills", middlewareBusinessDomain(true, false), r.SkillHandler.RegisterSkill)
	engine.GET("/skills", middlewareBusinessDomain(true, false), r.SkillHandler.QuerySkillList)
	engine.GET("/skills/:skill_id", middlewareBusinessDomain(true, false), r.SkillHandler.GetSkillDetail)
	engine.GET("/skills/:skill_id/guide", middlewareBusinessDomain(true, false), r.SkillHandler.GetSkillGuide)
	engine.POST("/skills/:skill_id/files/read", middlewareBusinessDomain(true, false), r.SkillHandler.ReadSkillFile)
	engine.DELETE("/skills/:skill_id", middlewareBusinessDomain(true, false), r.SkillHandler.DeleteSkill)
}
