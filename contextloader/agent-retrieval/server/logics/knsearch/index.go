package knsearch

import (
	"context"
	"sync"

	"devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/drivenadapters"
	"devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/infra/config"
	"devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/interfaces"
)

// KnSearchService kn_search 服务
type KnSearchService interface {
	KnSearch(ctx context.Context, req *interfaces.KnSearchReq) (resp *interfaces.KnSearchResp, err error)
}

type knSearchService struct {
	Logger        interfaces.Logger
	DataRetrieval interfaces.DataRetrieval
}

var (
	ksServiceOnce sync.Once
	ksService     KnSearchService
)

// NewKnSearchService 新建 KnSearchService
func NewKnSearchService() KnSearchService {
	ksServiceOnce.Do(func() {
		conf := config.NewConfigLoader()
		ksService = &knSearchService{
			Logger:        conf.GetLogger(),
			DataRetrieval: drivenadapters.NewDataRetrievalClient(),
		}
	})
	return ksService
}

// KnSearch 知识网络检索
func (s *knSearchService) KnSearch(ctx context.Context, req *interfaces.KnSearchReq) (resp *interfaces.KnSearchResp, err error) {
	// 将 kn_id 转换为 kn_ids 数组（内部使用，不对外暴露）
	knIDs := []*interfaces.KnDataSourceConfig{
		{
			KnowledgeNetworkID: req.KnID,
		},
	}
	req.SetKnIDs(knIDs)
	resp, err = s.DataRetrieval.KnSearch(ctx, req)
	return
}
