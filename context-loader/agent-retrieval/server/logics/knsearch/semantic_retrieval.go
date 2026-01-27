// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package knsearch 实例检索
// file: semantic_retrieval.go
// 实现单关键词和多关键词场景的实例检索逻辑
package knsearch

import (
	"context"
	"sort"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

const (
	// maxKeywordsDefault 默认的最大关键词数量
	maxKeywordsDefault = 5
	// multiKeywordLimitMultiplier 多关键词场景查询限制的倍数
	multiKeywordLimitMultiplier   = 2
	globalScoreFilterRatioDefault = 0.25
)

// SemanticRetrieval 语义实例检索器
type SemanticRetrieval struct {
	logger        interfaces.Logger
	mfModelClient interfaces.DrivenMFModelAPIClient
	ontologyQuery interfaces.DrivenOntologyQuery
	maxKeywords   int // 多关键词最大数量，默认5
}

// NewSemanticRetrieval 创建实例检索器
func NewSemanticRetrieval(
	logger interfaces.Logger,
	mfModelClient interfaces.DrivenMFModelAPIClient,
	ontologyQuery interfaces.DrivenOntologyQuery,
) *SemanticRetrieval {
	return &SemanticRetrieval{
		logger:        logger,
		mfModelClient: mfModelClient,
		ontologyQuery: ontologyQuery,
		maxKeywords:   maxKeywordsDefault,
	}
}

// InstanceResult 实例结果
type InstanceResult struct {
	ObjectTypeID string                 `json:"object_type_id"`
	InstanceID   string                 `json:"instance_id"`
	DisplayName  string                 `json:"display_name"`
	Properties   map[string]interface{} `json:"properties"`
	Score        float64                `json:"score"`
}

// Retrieve 执行实例检索
func (s *SemanticRetrieval) Retrieve(
	ctx context.Context,
	knID string, // Knowledge Network ID
	query string,
	objectTypes []map[string]interface{},
	perTypeLimit int,
) ([]map[string]interface{}, error) {
	s.logger.WithContext(ctx).Debugf("[SemanticRetrieval#Retrieve] Starting instance retrieval, query: %s, objectTypes: %d",
		query, len(objectTypes))

	if len(objectTypes) == 0 {
		return []map[string]interface{}{}, nil
	}

	// 分词
	keywords := s.tokenize(query)
	s.logger.WithContext(ctx).Debugf("[SemanticRetrieval#Retrieve] Keywords: %v", keywords)

	if len(keywords) == 0 {
		keywords = []string{query} // 回退使用完整查询
	}

	// 限制关键词数量
	if len(keywords) > s.maxKeywords {
		keywords = keywords[:s.maxKeywords]
	}

	var allInstances []map[string]interface{}

	if len(keywords) == 1 {
		// 单关键词场景：直接查询
		instances := s.singleKeywordSearch(ctx, knID, keywords[0], objectTypes, perTypeLimit)
		allInstances = instances
	} else {
		// 多关键词场景：并行查询 + 合并 + 重排
		instances := s.multiKeywordSearch(ctx, knID, query, keywords, objectTypes, perTypeLimit)
		allInstances = instances
	}

	s.logger.WithContext(ctx).Infof("[SemanticRetrieval#Retrieve] Retrieved %d instances", len(allInstances))

	return allInstances, nil
}

// singleKeywordSearch 单关键词检索
func (s *SemanticRetrieval) singleKeywordSearch(
	ctx context.Context,
	knID string,
	keyword string,
	objectTypes []map[string]interface{},
	perTypeLimit int,
) []map[string]interface{} {
	var allInstances []map[string]interface{}

	for _, objType := range objectTypes {
		objTypeID, ok := objType["id"].(string)
		if !ok {
			continue
		}

		// 调用OntologyQuery查询实例
		instances, err := s.queryInstances(ctx, knID, objTypeID, keyword, perTypeLimit)
		if err != nil {
			s.logger.WithContext(ctx).Warnf("[SemanticRetrieval#singleKeywordSearch] Query failed for %s: %v", objTypeID, err)
			continue
		}

		allInstances = append(allInstances, instances...)
	}

	return allInstances
}

// multiKeywordSearch 多关键词检索
// 使用 errgroup 并发查询，提高性能
func (s *SemanticRetrieval) multiKeywordSearch(
	ctx context.Context,
	knID string,
	fullQuery string,
	keywords []string,
	objectTypes []map[string]interface{},
	perTypeLimit int,
) []map[string]interface{} {
	// 1. 并发查询每个关键词
	candidateMap := make(map[string]map[string]interface{}) // 去重用
	var mu sync.Mutex                                       // 保护 candidateMap

	g, gCtx := errgroup.WithContext(ctx)

	for _, keyword := range keywords {
		keyword := keyword // 避免闭包陷阱
		for _, objType := range objectTypes {
			objType := objType // 避免闭包陷阱
			objTypeID, ok := objType["id"].(string)
			if !ok {
				continue
			}

			g.Go(func() error {
				instances, err := s.queryInstances(gCtx, knID, objTypeID, keyword, perTypeLimit*multiKeywordLimitMultiplier)
				if err != nil {
					// 单个查询失败不中断整体流程，记录日志
					s.logger.WithContext(gCtx).Warnf("[SemanticRetrieval#multiKeywordSearch] Query failed for %s/%s: %v",
						objTypeID, keyword, err)
					return nil // 不返回 error，允许其他查询继续
				}

				mu.Lock()
				for _, inst := range instances {
					// 使用对象类型ID+实例ID作为key去重
					instID, _ := inst["id"].(string)
					key := objTypeID + ":" + instID
					if _, exists := candidateMap[key]; !exists {
						candidateMap[key] = inst
					}
				}
				mu.Unlock()
				return nil
			})
		}
	}

	// 等待所有并发查询完成
	if err := g.Wait(); err != nil {
		s.logger.WithContext(ctx).Warnf("[SemanticRetrieval#multiKeywordSearch] Concurrent query error: %v", err)
	}

	if len(candidateMap) == 0 {
		return []map[string]interface{}{}
	}

	// 2. 转换为列表
	candidates := make([]map[string]interface{}, 0, len(candidateMap))
	for _, inst := range candidateMap {
		candidates = append(candidates, inst)
	}

	// 3. 使用Vector Rerank对候选进行重排序
	rankedCandidates, err := s.rerankCandidates(ctx, fullQuery, candidates)
	if err != nil {
		s.logger.WithContext(ctx).Warnf("[SemanticRetrieval#multiKeywordSearch] Rerank failed: %v", err)
		// 降级：返回原始候选
		return candidates
	}

	// 4. 全局分数过滤
	filteredCandidates := s.filterByGlobalScoreRatio(rankedCandidates, globalScoreFilterRatioDefault, true)

	// 如果过滤后太少，返回更多结果
	if len(filteredCandidates) < perTypeLimit && len(rankedCandidates) > 0 {
		if len(rankedCandidates) > perTypeLimit*len(objectTypes) {
			filteredCandidates = rankedCandidates[:perTypeLimit*len(objectTypes)]
		} else {
			filteredCandidates = rankedCandidates
		}
	}

	return filteredCandidates
}

// queryInstances 查询实例（调用OntologyQuery）
func (s *SemanticRetrieval) queryInstances(
	ctx context.Context,
	knID string,
	objectTypeID string,
	keyword string,
	limit int,
) ([]map[string]interface{}, error) {
	// 构建检索条件 - 使用match操作进行关键词匹配
	cond := &interfaces.KnCondition{
		Operation: interfaces.KnOperationTypeMatch,
		Value:     keyword,
	}

	// 调用OntologyQuery的QueryObjectInstances接口
	resp, err := s.ontologyQuery.QueryObjectInstances(ctx, &interfaces.QueryObjectInstancesReq{
		KnID:  knID,
		OtID:  objectTypeID,
		Cond:  cond,
		Limit: limit,
	})
	if err != nil {
		return nil, err
	}

	// 转换响应为map格式 - 创建新map避免并发修改原始数据
	var instances []map[string]interface{}
	for _, item := range resp.Data {
		if origMap, ok := item.(map[string]interface{}); ok {
			// 创建新map，避免并发修改原始map
			instMap := make(map[string]interface{}, len(origMap)+1)
			for k, v := range origMap {
				instMap[k] = v
			}
			instMap["object_type_id"] = objectTypeID
			instances = append(instances, instMap)
		}
	}

	return instances, nil
}

// rerankCandidates 使用Vector Rerank对候选进行重排序
func (s *SemanticRetrieval) rerankCandidates(
	ctx context.Context,
	query string,
	candidates []map[string]interface{},
) ([]map[string]interface{}, error) {
	if len(candidates) == 0 {
		return candidates, nil
	}

	// 构建文档列表
	documents := make([]string, len(candidates))
	for i, candidate := range candidates {
		displayName, _ := candidate["display_name"].(string)
		if displayName == "" {
			displayName, _ = candidate["id"].(string)
		}
		documents[i] = displayName
	}

	// 调用Rerank服务
	resp, err := s.mfModelClient.Rerank(ctx, query, documents)
	if err != nil {
		return nil, err
	}

	// 创建索引到分数的映射
	scoreMap := make(map[int]float64)
	for _, result := range resp.Results {
		scoreMap[result.Index] = result.RelevanceScore
	}

	// 添加分数到候选并排序
	type scoredCandidate struct {
		candidate map[string]interface{}
		score     float64
	}
	scored := make([]scoredCandidate, len(candidates))
	for i, candidate := range candidates {
		score := scoreMap[i]
		candidate["_rerank_score"] = score
		scored[i] = scoredCandidate{candidate: candidate, score: score}
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	result := make([]map[string]interface{}, len(scored))
	for i, s := range scored {
		result[i] = s.candidate
	}

	return result, nil
}

// tokenize 分词
// 支持多种分隔符：空格、逗号、分号、斜杠、顿号等
func (s *SemanticRetrieval) tokenize(query string) []string {
	// 替换常见分隔符为空格
	separators := []string{",", ";", "/", "\\", "、", "，", "；", "｜", "|"}
	normalized := query
	for _, sep := range separators {
		normalized = strings.ReplaceAll(normalized, sep, " ")
	}

	// 按空格分词
	parts := strings.Fields(normalized)
	var keywords []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if len(p) > 1 { // 过滤单字符
			keywords = append(keywords, p)
		}
	}
	return keywords
}

// filterByGlobalScoreRatio 全局分数过滤
// 过滤掉分数低于 (最高分 * ratio) 的实例，每类至少保留一个
func (s *SemanticRetrieval) filterByGlobalScoreRatio(
	instances []map[string]interface{},
	ratio float64,
	keepAtLeastOne bool,
) []map[string]interface{} {
	if len(instances) == 0 || ratio <= 0 {
		return instances
	}

	// 找最高分
	maxScore := 0.0
	for _, inst := range instances {
		if score, ok := inst["_rerank_score"].(float64); ok && score > maxScore {
			maxScore = score
		}
	}

	threshold := maxScore * ratio

	// 按对象类型分组过滤
	byType := make(map[string][]map[string]interface{})
	for _, inst := range instances {
		typeID, _ := inst["object_type_id"].(string)
		byType[typeID] = append(byType[typeID], inst)
	}

	var result []map[string]interface{}
	for _, typeInstances := range byType {
		var kept []map[string]interface{}
		var best map[string]interface{}
		bestScore := -1.0

		for _, inst := range typeInstances {
			score, _ := inst["_rerank_score"].(float64)
			if score >= threshold {
				kept = append(kept, inst)
			}
			if score > bestScore {
				bestScore = score
				best = inst
			}
		}

		// 保证每类至少一个
		if keepAtLeastOne && len(kept) == 0 && best != nil {
			kept = append(kept, best)
		}
		result = append(result, kept...)
	}

	return result
}
