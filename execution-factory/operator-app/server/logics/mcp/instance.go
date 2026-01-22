package mcp

import (
	"context"
	"net/http"

	infraerrors "github.com/kweaver-ai/adp/execution-factory/operator-app/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-app/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-app/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-app/server/logics/mcp/core"
	"github.com/kweaver-ai/adp/execution-factory/operator-app/server/logics/mcp/deployer"
	"github.com/kweaver-ai/adp/execution-factory/operator-app/server/logics/mcp/storage"
	"github.com/kweaver-ai/adp/execution-factory/operator-app/server/utils"
)

func (s *mcpInstanceServiceImpl) InitOnStartup(ctx context.Context) (err error) {
	s.Logger.Info("init mcp instance on startup")
	defer func() {
		if err != nil {
			s.Logger.Error("init mcp instance on startup failed", "error", err)
		} else {
			s.Logger.Info("init mcp instance on startup done")
		}
	}()

	// 1. 从数据库查询所有mcp配置
	resourceDeploys, err := s.DBResourceDeploy.SelectList(ctx, nil, &model.ResourceDeployDB{
		Type: interfaces.ResourceDeployTypeMCP.String(),
	})
	if err != nil {
		s.Logger.Errorf("[InitOnStartup] select mcp resource deploy failed, error: %+v", err)
		return
	}

	// 2. 初始化instanceManager和存储
	mcpStorage := storage.NewMemoryStore()
	instanceManager := core.NewInstanceManager(mcpStorage, deployer.NewHTTPDeployer(), deployer.NewSSEDeployer())

	// 3. 初始化mcp实例
	for _, resourceDeploy := range resourceDeploys {
		var mcpConfig *interfaces.MCPConfig
		mcpConfig, err = utils.JSONToObjectWithError[*interfaces.MCPConfig](resourceDeploy.Config)
		if err != nil {
			s.Logger.Errorf("[InitOnStartup] create mcp instance failed, mcpID: %s, mcpVersion: %d, error: %+v", mcpConfig.MCPID, mcpConfig.Version, err)
			return
		}
		_, err = instanceManager.Create(ctx, mcpConfig)
		if err != nil {
			s.Logger.Errorf("[InitOnStartup] create mcp instance failed, mcpID: %s, mcpVersion: %d, error: %+v", mcpConfig.MCPID, mcpConfig.Version, err)
			return
		}
	}
	return
}

func (s *mcpInstanceServiceImpl) CreateMCPInstance(ctx context.Context, req *interfaces.MCPDeployCreateRequest) (*interfaces.MCPDeployCreateResponse, error) {
	// 检查实例是否存在
	mcpStorage := storage.NewMemoryStore()
	if mcpStorage.Exists(req.MCPID, req.Version) {
		s.Logger.Warnf("[CreateMCPInstance] mcp instance already exists, mcpID: %s, mcpVersion: %d", req.MCPID, req.Version)
		return nil, infraerrors.NewHTTPError(context.Background(), http.StatusBadRequest, infraerrors.ErrExtMCPInstanceAlreadyExists, nil)
	}

	tx, err := s.DBTx.GetTx(ctx)
	if err != nil {
		s.Logger.Errorf("[CreateMCPInstance] get tx failed, error: %+v", err)
		return nil, err
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				s.Logger.Errorf("[CreateMCPInstance] rollback failed: %v", rbErr)
			}
		} else {
			if cmErr := tx.Commit(); cmErr != nil {
				s.Logger.Errorf("[CreateMCPInstance] commit failed: %v", cmErr)
			}
		}
	}()

	mcpConfig := &interfaces.MCPConfig{
		MCPID:        req.MCPID,
		Version:      req.Version,
		Name:         req.Name,
		Instructions: req.Instructions,
		Tools:        req.Tools,
	}

	// 保存实例配置
	resourceDeploy := &model.ResourceDeployDB{
		ResourceID:  req.MCPID,
		Type:        interfaces.ResourceDeployTypeMCP.String(),
		Version:     req.Version,
		Name:        req.Name,
		Description: req.Instructions,
		Config:      utils.ObjectToJSON(mcpConfig),
	}

	_, err = s.DBResourceDeploy.Insert(ctx, tx, resourceDeploy)
	if err != nil {
		s.Logger.Errorf("[CreateMCPInstance] insert resource deploy failed, error: %+v", err)
		return nil, err
	}

	// 创建mcp实例
	instanceManager := core.NewInstanceManager(mcpStorage, deployer.NewHTTPDeployer(), deployer.NewSSEDeployer())
	instance, err := instanceManager.Create(ctx, mcpConfig)
	if err != nil {
		s.Logger.Errorf("[CreateMCPInstance] create mcp instance failed, mcpID: %s, mcpVersion: %d, error: %+v", req.MCPID, req.Version, err)
		return nil, err
	}
	return &interfaces.MCPDeployCreateResponse{
		MCPID:      req.MCPID,
		MCPVersion: req.Version,
		StreamURL:  instance.StreamRoutePath,
		SSEURL:     instance.SSERoutePath,
	}, nil
}

func (s *mcpInstanceServiceImpl) DeleteMCPInstance(ctx context.Context, mcpID string, mcpVersion int) error {
	// 删除数据实例配置
	tx, err := s.DBTx.GetTx(ctx)
	if err != nil {
		s.Logger.Errorf("[DeleteMCPInstance] get tx failed, error: %+v", err)
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	err = s.DBResourceDeploy.Delete(ctx, tx, mcpVersion, interfaces.ResourceDeployTypeMCP.String(), mcpID)
	if err != nil {
		s.Logger.Errorf("[DeleteMCPInstance] delete mcp instance failed, mcpID: %s, mcpVersion: %d, error: %+v", mcpID, mcpVersion, err)
		return err
	}

	mcpStorage := storage.NewMemoryStore()
	instanceManager := core.NewInstanceManager(mcpStorage, deployer.NewHTTPDeployer(), deployer.NewSSEDeployer())
	err = instanceManager.Delete(ctx, mcpID, mcpVersion)
	if err != nil {
		s.Logger.Errorf("[DeleteMCPInstance] delete mcp instance failed, mcpID: %s, mcpVersion: %d, error: %+v", mcpID, mcpVersion, err)
		return err
	}
	return nil
}

func (s *mcpInstanceServiceImpl) DeleteByMCPID(ctx context.Context, mcpID string) error {
	// 获取所有实例版本
	resourceDeploys, err := s.DBResourceDeploy.SelectListByResourceID(ctx, mcpID)
	if err != nil {
		s.Logger.Errorf("[DeleteByMCPID] get mcp instance config failed, mcpID: %s, error: %+v", mcpID, err)
		return err
	}
	if len(resourceDeploys) == 0 {
		s.Logger.Warnf("[DeleteByMCPID] mcp instance not found, mcpID: %s", mcpID)
		return nil
	}
	for _, resourceDeploy := range resourceDeploys {
		err = s.DeleteMCPInstance(ctx, mcpID, resourceDeploy.Version)
		if err != nil {
			s.Logger.Errorf("[DeleteByMCPID] delete mcp instance failed, mcpID: %s, mcpVersion: %d, error: %+v", mcpID, resourceDeploy.Version, err)
			return err
		}
	}
	return nil
}

func (s *mcpInstanceServiceImpl) UpdateMCPInstance(ctx context.Context, req *interfaces.MCPDeployUpdateRequest) (*interfaces.MCPDeployUpdateResponse, error) {
	// 更新实例配置
	tx, err := s.DBTx.GetTx(ctx)
	if err != nil {
		s.Logger.Errorf("[UpdateMCPInstance] get tx failed, error: %+v", err)
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	mcpConfig := &interfaces.MCPConfig{
		MCPID:        req.MCPID,
		Version:      req.Version,
		Name:         req.Name,
		Instructions: req.Instructions,
		Tools:        req.Tools,
	}

	err = s.DBResourceDeploy.Update(ctx, tx, &model.ResourceDeployDB{
		ResourceID:  req.MCPID,
		Type:        interfaces.ResourceDeployTypeMCP.String(),
		Version:     req.Version,
		Name:        req.Name,
		Description: req.Instructions,
		Config:      utils.ObjectToJSON(mcpConfig),
	})
	if err != nil {
		s.Logger.Errorf("[UpdateMCPInstance] update mcp instance config failed, mcpID: %s, mcpVersion: %d, error: %+v", req.MCPID, req.Version, err)
		return nil, err
	}

	// 移除旧的实例
	mcpStorage := storage.NewMemoryStore()
	instanceManager := core.NewInstanceManager(mcpStorage, deployer.NewHTTPDeployer(), deployer.NewSSEDeployer())
	err = instanceManager.Delete(ctx, req.MCPID, req.Version)
	if err != nil {
		s.Logger.Errorf("[UpdateMCPInstance] delete old mcp instance failed, mcpID: %s, mcpVersion: %d, error: %+v", req.MCPID, req.Version, err)
		return nil, err
	}
	// 创建新的实例
	instance, err := instanceManager.Create(ctx, mcpConfig)
	if err != nil {
		s.Logger.Errorf("[UpdateMCPInstance] create new mcp instance failed, mcpID: %s, mcpVersion: %d, error: %+v", req.MCPID, req.Version, err)
		return nil, err
	}
	// 返回新的实例
	return &interfaces.MCPDeployUpdateResponse{
		MCPID:      req.MCPID,
		MCPVersion: req.Version,
		StreamURL:  instance.StreamRoutePath,
		SSEURL:     instance.SSERoutePath,
	}, nil
}

func (s *mcpInstanceServiceImpl) GetMCPInstance(ctx context.Context, mcpID string, mcpVersion int) (*interfaces.MCPServerInstance, error) {
	mcpStorage := storage.NewMemoryStore()
	instance, err := mcpStorage.Get(mcpID, mcpVersion)
	if err != nil {
		s.Logger.Warnf("[GetMCPInstance] get mcp instance failed, mcpID: %s, mcpVersion: %d, error: %v", mcpID, mcpVersion, err)
		return nil, err
	}
	return instance, nil
}

// UpgradeMCPInstance 升级mcp实例
func (s *mcpInstanceServiceImpl) UpgradeMCPInstance(ctx context.Context, req *interfaces.MCPDeployCreateRequest) (*interfaces.MCPDeployCreateResponse, error) {
	// 检查实例是否存在
	mcpStorage := storage.NewMemoryStore()
	if mcpStorage.Exists(req.MCPID, req.Version) {
		// 实例存在则更新
		resp, err := s.UpdateMCPInstance(ctx, &interfaces.MCPDeployUpdateRequest{
			MCPID:        req.MCPID,
			Version:      req.Version,
			Name:         req.Name,
			Instructions: req.Instructions,
			Tools:        req.Tools,
		})
		if err != nil {
			s.Logger.Errorf("[UpgradeMCPInstance] upgrade mcp instance failed, mcpID: %s, mcpVersion: %d, error: %+v", req.MCPID, req.Version, err)
			return nil, err
		}
		return &interfaces.MCPDeployCreateResponse{
			MCPID:      req.MCPID,
			MCPVersion: req.Version,
			StreamURL:  resp.StreamURL,
			SSEURL:     resp.SSEURL,
		}, nil
	}
	return s.CreateMCPInstance(ctx, req)
}
