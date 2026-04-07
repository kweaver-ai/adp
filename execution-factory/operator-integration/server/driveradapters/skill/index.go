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
	UpdateSkillStatus(c *gin.Context)
	DownloadSkill(c *gin.Context)
	QuerySkillList(c *gin.Context)
	QuerySkillMarketList(c *gin.Context)
	GetSkillMarketDetail(c *gin.Context)
	GetSkillDetail(c *gin.Context)
	GetSkillContent(c *gin.Context)
	ReadSkillFile(c *gin.Context)
	UpsertSkillRuntimeProfile(c *gin.Context)
	GetSkillRuntimeProfile(c *gin.Context)
	ExecuteSkill(c *gin.Context)
}

type skillHandler struct {
	Logger        interfaces.Logger
	Registry      interfaces.SkillRegistry
	Market        interfaces.SkillMarket
	Reader        interfaces.SkillReader
	RuntimeProfile interfaces.SkillRuntimeProfileService
	Executor      interfaces.SkillExecutionService
}

var (
	once sync.Once
	h    SkillHandler
)

func NewSkillHandler() SkillHandler {
	once.Do(func() {
		conf := config.NewConfigLoader()
		runtimeService := logicsskill.NewSkillRuntimeService()
		executionService := logicsskill.NewSkillExecutionService()
		registry := logicsskill.NewSkillRegistry()
		market, _ := registry.(interfaces.SkillMarket)
		h = &skillHandler{
			Logger:        conf.GetLogger(),
			Registry:      registry,
			Market:        market,
			Reader:        logicsskill.NewSkillReader(),
			RuntimeProfile: runtimeService,
			Executor:      executionService,
		}
	})
	return h
}
