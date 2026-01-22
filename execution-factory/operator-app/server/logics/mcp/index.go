package mcp

import (
	"sync"

	"github.com/kweaver-ai/adp/execution-factory/operator-app/server/dbaccess"
	"github.com/kweaver-ai/adp/execution-factory/operator-app/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-app/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-app/server/interfaces/model"
)

type mcpInstanceServiceImpl struct {
	Logger           interfaces.Logger
	DBResourceDeploy model.DBResourceDeploy
	DBTx             model.DBTx
}

var (
	mcpOnce            sync.Once
	mcpInstanceService *mcpInstanceServiceImpl
)

func NewMCPInstanceService() interfaces.IMCPInstanceService {
	configLoader := config.NewConfigLoader()
	mcpOnce.Do(func() {
		mcpInstanceService = &mcpInstanceServiceImpl{
			Logger:           configLoader.GetLogger(),
			DBResourceDeploy: dbaccess.NewResourceDeployDB(),
			DBTx:             dbaccess.NewBaseTx(),
		}
	})
	return mcpInstanceService
}
