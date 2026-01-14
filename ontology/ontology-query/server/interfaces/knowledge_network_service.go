package interfaces

import "context"

//go:generate mockgen -source ../interfaces/knowledge_network_service.go -destination ../interfaces/mock/mock_knowledge_network_service.go
type KnowledgeNetworkService interface {
	SearchSubgraph(ctx context.Context, query *SubGraphQueryBaseOnSource) (ObjectSubGraph, error)
	SearchSubgraphByTypePath(ctx context.Context, query *SubGraphQueryBaseOnTypePath) (PathsEntries, error)
}
