package interfaces

//go:generate mockgen -source=interface.go -destination=../mocks/interface.go -package=mocks
import "context"

// App 应用接口
type App interface {
	Start() error
	Stop(context.Context)
}

type ResourceDeployType string

func (r ResourceDeployType) String() string {
	return string(r)
}

// IKnQuerySubgraphService 子图查询服务接口
type IKnQuerySubgraphService interface {
	// QueryInstanceSubgraph 查询对象子图
	QueryInstanceSubgraph(ctx context.Context, req *QueryInstanceSubgraphReq) (resp *QueryInstanceSubgraphResp, err error)
}

// IKnSearchService kn_search 服务接口
type IKnSearchService interface {
	// KnSearch 知识网络检索
	KnSearch(ctx context.Context, req *KnSearchReq) (resp *KnSearchResp, err error)
}
