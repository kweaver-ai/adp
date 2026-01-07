// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knsearch

import (
	"context"
	"sync"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/drivenadapters"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/config"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

// KnSearchService kn_search service
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

// NewKnSearchService creates new KnSearchService
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

// KnSearch Knowledge network retrieval
func (s *knSearchService) KnSearch(ctx context.Context, req *interfaces.KnSearchReq) (resp *interfaces.KnSearchResp, err error) {
	// Convert kn_id to kn_ids array (internal use, not exposed)
	knIDs := []*interfaces.KnDataSourceConfig{
		{
			KnowledgeNetworkID: req.KnID,
		},
	}
	req.SetKnIDs(knIDs)
	resp, err = s.DataRetrieval.KnSearch(ctx, req)
	return
}
