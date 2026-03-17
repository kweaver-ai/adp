package skill

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	logicsskill "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/skill"
)

type SkillHandler interface {
	RegisterSkill(c *gin.Context)
	DeleteSkill(c *gin.Context)
	QuerySkillList(c *gin.Context)
	GetSkillDetail(c *gin.Context)
	GetSkillGuide(c *gin.Context)
	ReadSkillFile(c *gin.Context)
}

type skillHandler struct {
	Logger   interfaces.Logger
	Registry interfaces.SkillRegistry
	Reader   interfaces.SkillReader
}

var (
	once sync.Once
	h    SkillHandler
)

func NewSkillHandler() SkillHandler {
	once.Do(func() {
		conf := config.NewConfigLoader()
		h = &skillHandler{
			Logger:   conf.GetLogger(),
			Registry: logicsskill.NewSkillRegistry(),
			Reader:   logicsskill.NewSkillReader(),
		}
	})
	return h
}
