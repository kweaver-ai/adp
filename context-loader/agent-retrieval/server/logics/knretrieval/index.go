// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package knretrieval 基于业务知识网络实现统一检索
// file: index.go
package knretrieval

import (
	"sync"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/drivenadapters"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/config"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

type knRetrievalServiceImpl struct {
	logger                interfaces.Logger
	agentClient           interfaces.AgentApp
	ontologyQueryAccess   interfaces.DrivenOntologyQuery
	ontologyManagerAccess interfaces.OntologyManagerAccess
	dataRetrieval         interfaces.DataRetrieval
}

var (
	krOnce             sync.Once
	knRetrievalService interfaces.IKnRetrievalService
)

func NewKnRetrievalService() interfaces.IKnRetrievalService {
	krOnce.Do(func() {
		knRetrievalService = &knRetrievalServiceImpl{
			logger:                config.NewConfigLoader().GetLogger(),
			agentClient:           drivenadapters.NewAgentAppClient(),
			ontologyQueryAccess:   drivenadapters.NewOntologyQueryAccess(),
			ontologyManagerAccess: drivenadapters.NewOntologyManagerAccess(),
			dataRetrieval:         drivenadapters.NewDataRetrievalClient(),
		}
	})
	return knRetrievalService
}
