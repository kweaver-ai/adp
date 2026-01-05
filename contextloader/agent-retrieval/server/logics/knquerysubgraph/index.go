// Package knquerysubgraph 业务知识网络子图查询业务逻辑
// file: index.go
package knquerysubgraph

import (
	"context"
	"sync"

	"devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/drivenadapters"
	"devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/infra/config"
	"devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/interfaces"
)

// KnQuerySubgraphService 子图查询服务
type KnQuerySubgraphService interface {
	QueryInstanceSubgraph(ctx context.Context, req *interfaces.QueryInstanceSubgraphReq) (resp *interfaces.QueryInstanceSubgraphResp, err error)
}

type knQuerySubgraphService struct {
	Logger        interfaces.Logger
	OntologyQuery interfaces.DrivenOntologyQuery
}

var (
	kqsServiceOnce sync.Once
	kqsService     KnQuerySubgraphService
)

// NewKnQuerySubgraphService 新建 KnQuerySubgraphService
func NewKnQuerySubgraphService() KnQuerySubgraphService {
	kqsServiceOnce.Do(func() {
		conf := config.NewConfigLoader()
		kqsService = &knQuerySubgraphService{
			Logger:        conf.GetLogger(),
			OntologyQuery: drivenadapters.NewOntologyQueryAccess(),
		}
	})
	return kqsService
}

// QueryInstanceSubgraph 查询对象子图
func (s *knQuerySubgraphService) QueryInstanceSubgraph(ctx context.Context, req *interfaces.QueryInstanceSubgraphReq) (resp *interfaces.QueryInstanceSubgraphResp, err error) {
	// 调用 drivenadapters 层查询子图
	resp, err = s.OntologyQuery.QueryInstanceSubgraph(ctx, req)
	return
}
