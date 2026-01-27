// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package knsearch provides business logic for knowledge network search operations.
package knsearch

import (
	"context"
	"sync"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/drivenadapters"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/config"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

// useLocalSearch Feature Flag: 是否使用本地KnSearch
// 迁移验证通过后改为true，最终删除此开关和远程调用代码
const useLocalSearch = false

// KnSearchService kn_search service
type KnSearchService interface {
	KnSearch(ctx context.Context, req *interfaces.KnSearchReq) (resp *interfaces.KnSearchResp, err error)
}

type knSearchService struct {
	Logger         interfaces.Logger
	DataRetrieval  interfaces.DataRetrieval
	LocalKnSearch  *LocalKnSearch
	UseLocalSearch bool
}

var (
	ksServiceOnce sync.Once
	ksService     KnSearchService
)

// NewKnSearchService creates new KnSearchService
func NewKnSearchService() KnSearchService {
	ksServiceOnce.Do(func() {
		conf := config.NewConfigLoader()
		logger := conf.GetLogger()

		// 创建统一的mf-model-api客户端（同时提供LLM和Rerank能力）
		mfModelClient := drivenadapters.NewMFModelAPIClient()
		ontologyQuery := drivenadapters.NewOntologyQueryAccess()
		ontologyManager := drivenadapters.NewOntologyManagerAccess()

		ksService = &knSearchService{
			Logger:         logger,
			DataRetrieval:  drivenadapters.NewDataRetrievalClient(),
			LocalKnSearch:  NewLocalKnSearch(logger, mfModelClient, ontologyQuery, ontologyManager),
			UseLocalSearch: useLocalSearch,
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

	if s.UseLocalSearch {
		// 使用本地逻辑
		s.Logger.WithContext(ctx).Info("[KnSearch] Using local KnSearch logic")
		resp, err = s.LocalKnSearch.Search(ctx, req)
		if err != nil {
			s.Logger.WithContext(ctx).Errorf("[KnSearch] Local search failed: %v", err)
			return
		}
		return
	}

	// 使用远程调用
	resp, err = s.DataRetrieval.KnSearch(ctx, req)
	return
}
