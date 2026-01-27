// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package knsearch 知识网络检索逻辑
// file: knsearch_logic.go
// 本文件实现本地的kn_search逻辑，不依赖Python data-retrieval服务
package knsearch

import (
	"context"
	"encoding/json"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/logics/knrerank"
)

// defaultTopK 默认的TopK参数
const defaultTopK = 10

// LocalKnSearch 本地KnSearch逻辑实现
// 对应Python的5步检索流程
type LocalKnSearch struct {
	logger            interfaces.Logger
	mfModelClient     interfaces.DrivenMFModelAPIClient
	ontologyQuery     interfaces.DrivenOntologyQuery
	ontologyManager   interfaces.OntologyManagerAccess // 用于获取知识网络Schema
	sessionCache      *SessionCache
	conceptRetrieval  *ConceptRetrieval
	semanticRetrieval *SemanticRetrieval
	knowledgeReranker *knrerank.KnowledgeReranker // 复用的 reranker 实例
}

// NewLocalKnSearch 创建本地KnSearch实例
func NewLocalKnSearch(
	logger interfaces.Logger,
	mfModelClient interfaces.DrivenMFModelAPIClient,
	ontologyQuery interfaces.DrivenOntologyQuery,
	ontologyManager interfaces.OntologyManagerAccess,
) *LocalKnSearch {
	return &LocalKnSearch{
		logger:            logger,
		mfModelClient:     mfModelClient,
		ontologyQuery:     ontologyQuery,
		ontologyManager:   ontologyManager,
		sessionCache:      NewSessionCache(),
		conceptRetrieval:  NewConceptRetrieval(logger, mfModelClient),
		semanticRetrieval: NewSemanticRetrieval(logger, mfModelClient, ontologyQuery),
		knowledgeReranker: knrerank.NewKnowledgeReranker(mfModelClient, logger), // 单例
	}
}

// Search 执行本地5步检索
func (l *LocalKnSearch) Search(ctx context.Context, req *interfaces.KnSearchReq) (*interfaces.KnSearchResp, error) {
	l.logger.WithContext(ctx).Infof("[LocalKnSearch#Search] Starting local search for kn_id=%s, query=%s",
		req.KnID, req.Query)

	// 获取会话ID（用于缓存）
	sessionID := ""
	if req.SessionID != nil {
		sessionID = *req.SessionID
	}

	// Step 1: 尝试从缓存获取Schema
	cachedObjs, cachedRels, found := l.sessionCache.GetSchema(sessionID, req.KnID)
	var objectTypes []map[string]interface{}
	var relationTypes []map[string]interface{}

	if found && sessionID != "" {
		l.logger.WithContext(ctx).Debug("[LocalKnSearch#Search] Using cached schema")
		objectTypes = l.interfaceSliceToMapSlice(cachedObjs)
		relationTypes = l.interfaceSliceToMapSlice(cachedRels)
	} else {
		// Step 2: 获取知识网络详情
		networkDetails := l.getNetworkDetails(ctx, req.KnID)

		// Step 3: Schema Retrieval (概念召回)
		schemaResult, err := l.conceptRetrieval.Retrieve(
			ctx,
			req.Query,
			networkDetails,
			defaultTopK,
			req.XAccountID,
			req.XAccountType,
		)
		if err != nil {
			l.logger.WithContext(ctx).Warnf("[LocalKnSearch#Search] Schema retrieval failed: %v", err)
			objectTypes = []map[string]interface{}{}
			relationTypes = []map[string]interface{}{}
		} else {
			objectTypes = schemaResult.ObjectTypes
			relationTypes = schemaResult.RelationTypes
		}

		// Step 3.5: Concept Rerank (可选)
		enableRerank := true
		if req.EnableRerank != nil {
			enableRerank = *req.EnableRerank
		}
		if enableRerank && len(objectTypes) > 0 {
			objectTypes, relationTypes = l.conceptRerank(ctx, req, objectTypes, relationTypes)
		}

		// 缓存Schema
		if sessionID != "" {
			l.sessionCache.SetSchema(sessionID, req.KnID,
				l.mapSliceToInterfaceSlice(objectTypes),
				l.mapSliceToInterfaceSlice(relationTypes))
		}
	}

	// Step 4: 检查是否只需要Schema
	onlySchema := false
	if req.OnlySchema != nil && *req.OnlySchema {
		onlySchema = true
	}

	if onlySchema {
		l.logger.WithContext(ctx).Debug("[LocalKnSearch#Search] Returning schema only")
		return &interfaces.KnSearchResp{
			ObjectTypes:   l.mapSliceToInterfaceSlice(objectTypes),
			RelationTypes: l.mapSliceToInterfaceSlice(relationTypes),
		}, nil
	}

	// 解析检索配置
	retConfig := l.parseRetrievalConfig(req.RetrievalConfig)
	perTypeLimit := defaultTopK
	if retConfig != nil && retConfig.SemanticInstanceRetrieval != nil {
		if retConfig.SemanticInstanceRetrieval.PerTypeInstanceLimit > 0 {
			perTypeLimit = retConfig.SemanticInstanceRetrieval.PerTypeInstanceLimit
		}
	}

	// Step 5: Instance Retrieval (实例检索)
	nodes, err := l.semanticRetrieval.Retrieve(ctx, req.KnID, req.Query, objectTypes, perTypeLimit)
	if err != nil {
		l.logger.WithContext(ctx).Warnf("[LocalKnSearch#Search] Instance retrieval failed: %v", err)
		nodes = []map[string]interface{}{}
	}

	l.logger.WithContext(ctx).Infof("[LocalKnSearch#Search] Completed: %d object types, %d relation types, %d nodes",
		len(objectTypes), len(relationTypes), len(nodes))

	// Step 6: 组装返回结果
	return &interfaces.KnSearchResp{
		ObjectTypes:   l.mapSliceToInterfaceSlice(objectTypes),
		RelationTypes: l.mapSliceToInterfaceSlice(relationTypes),
		Nodes:         l.mapSliceToInterfaceSlice(nodes),
	}, nil
}

// getNetworkDetails 获取知识网络详情
// 调用 OntologyManager API 获取对象类型和关系类型的完整 Schema
func (l *LocalKnSearch) getNetworkDetails(ctx context.Context, knID string) map[string]interface{} {
	l.logger.WithContext(ctx).Debugf("[LocalKnSearch#getNetworkDetails] Getting network details for kn_id=%s", knID)

	// 初始化结果
	networkDetails := map[string]interface{}{
		"id":             knID,
		"name":           "",
		"comment":        "",
		"relation_types": []interface{}{},
		"object_types":   []interface{}{},
	}

	// 构造查询请求（不使用关键词，获取所有概念）
	queryReq := &interfaces.QueryConceptsReq{
		KnID: knID,
	}

	// 获取对象类型列表
	objectTypesResp, err := l.ontologyManager.SearchObjectTypes(ctx, queryReq)
	if err != nil {
		l.logger.WithContext(ctx).Warnf("[LocalKnSearch#getNetworkDetails] Failed to get object types: %v", err)
	} else if objectTypesResp != nil && len(objectTypesResp.Entries) > 0 {
		objectTypes := make([]interface{}, 0, len(objectTypesResp.Entries))
		for _, ot := range objectTypesResp.Entries {
			objectTypes = append(objectTypes, l.objectTypeToMap(ot))
		}
		networkDetails["object_types"] = objectTypes
		l.logger.WithContext(ctx).Debugf("[LocalKnSearch#getNetworkDetails] Got %d object types", len(objectTypes))
	}

	// 获取关系类型列表
	relationTypesResp, err := l.ontologyManager.SearchRelationTypes(ctx, queryReq)
	if err != nil {
		l.logger.WithContext(ctx).Warnf("[LocalKnSearch#getNetworkDetails] Failed to get relation types: %v", err)
	} else if relationTypesResp != nil && len(relationTypesResp.Entries) > 0 {
		relationTypes := make([]interface{}, 0, len(relationTypesResp.Entries))
		for _, rt := range relationTypesResp.Entries {
			relationTypes = append(relationTypes, l.relationTypeToMap(rt))
		}
		networkDetails["relation_types"] = relationTypes
		l.logger.WithContext(ctx).Debugf("[LocalKnSearch#getNetworkDetails] Got %d relation types", len(relationTypes))
	}

	return networkDetails
}

// objectTypeToMap 将 ObjectType 结构转换为 map[string]interface{}
func (l *LocalKnSearch) objectTypeToMap(ot *interfaces.ObjectType) map[string]interface{} {
	result := map[string]interface{}{
		"id":      ot.ID,
		"name":    ot.Name,
		"comment": ot.Comment,
	}

	// 转换 data_properties
	if ot.DataProperties != nil {
		dataProps := make([]interface{}, 0, len(ot.DataProperties))
		for _, dp := range ot.DataProperties {
			propMap := map[string]interface{}{
				"name":         dp.Name,
				"display_name": dp.DisplayName,
				"data_type":    dp.Type,
				"comment":      dp.Comment,
			}
			if dp.ConditionOperations != nil {
				propMap["condition_operations"] = dp.ConditionOperations
			}
			dataProps = append(dataProps, propMap)
		}
		result["data_properties"] = dataProps
	}

	// 转换 logic_properties
	if ot.LogicProperties != nil {
		logicProps := make([]interface{}, 0, len(ot.LogicProperties))
		for _, lp := range ot.LogicProperties {
			propMap := map[string]interface{}{
				"name":         lp.Name,
				"display_name": lp.DisplayName,
				"data_type":    string(lp.Type),
				"comment":      lp.Comment,
			}
			logicProps = append(logicProps, propMap)
		}
		result["logic_properties"] = logicProps
	}

	// 添加其他重要字段
	if ot.PrimaryKeys != nil {
		result["primary_keys"] = ot.PrimaryKeys
	}

	return result
}

// relationTypeToMap 将 RelationType 结构转换为 map[string]interface{}
func (l *LocalKnSearch) relationTypeToMap(rt *interfaces.RelationType) map[string]interface{} {
	return map[string]interface{}{
		"id":                    rt.ID,
		"name":                  rt.Name,
		"comment":               rt.Comment,
		"source_object_type_id": rt.SourceObjectTypeID,
		"target_object_type_id": rt.TargetObjectTypeID,
	}
}

// 辅助函数：map切片转interface切片
func (l *LocalKnSearch) mapSliceToInterfaceSlice(maps []map[string]interface{}) []interface{} {
	result := make([]interface{}, len(maps))
	for i, m := range maps {
		result[i] = m
	}
	return result
}

// 辅助函数：interface切片转map切片
func (l *LocalKnSearch) interfaceSliceToMapSlice(ifaces []interface{}) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(ifaces))
	for _, iface := range ifaces {
		if m, ok := iface.(map[string]interface{}); ok {
			result = append(result, m)
		}
	}
	return result
}

// conceptRerank 对概念进行重排序
// 对应 Python 版本 Step 4: Concept Rerank
func (l *LocalKnSearch) conceptRerank(
	ctx context.Context,
	req *interfaces.KnSearchReq,
	objectTypes []map[string]interface{},
	relationTypes []map[string]interface{},
) (filteredObjects, filteredRelations []map[string]interface{}) {
	// 构建 ConceptResult 列表
	concepts := l.buildConceptsFromSchema(objectTypes, relationTypes)
	if len(concepts) == 0 {
		return objectTypes, relationTypes
	}

	// 构建重排请求（复用已初始化的 reranker）
	rerankReq := &interfaces.KnowledgeRerankReq{
		QueryUnderstanding: &interfaces.QueryUnderstanding{
			OriginQuery: req.Query,
		},
		KnowledgeConcepts: concepts,
		Action:            interfaces.KnowledgeRerankActionLLM,
	}

	// 调用 Reranker
	rerankedConcepts, err := l.knowledgeReranker.Rerank(ctx, rerankReq)
	if err != nil {
		l.logger.WithContext(ctx).Warnf("[LocalKnSearch#conceptRerank] Rerank failed: %v", err)
		return objectTypes, relationTypes
	}

	// 根据重排结果过滤
	return l.filterByRerankScore(rerankedConcepts, objectTypes, relationTypes)
}

// buildConceptsFromSchema 从 Schema 构建 ConceptResult 列表
func (l *LocalKnSearch) buildConceptsFromSchema(
	objectTypes []map[string]interface{},
	relationTypes []map[string]interface{},
) []*interfaces.ConceptResult {
	var concepts []*interfaces.ConceptResult

	// 添加对象类型
	for _, ot := range objectTypes {
		conceptID, _ := ot["id"].(string)
		conceptName, _ := ot["name"].(string)
		concepts = append(concepts, &interfaces.ConceptResult{
			ConceptType:   interfaces.KnConceptTypeObject,
			ConceptID:     conceptID,
			ConceptName:   conceptName,
			ConceptDetail: ot,
		})
	}

	// 添加关系类型
	for _, rt := range relationTypes {
		conceptID, _ := rt["id"].(string)
		conceptName, _ := rt["name"].(string)
		concepts = append(concepts, &interfaces.ConceptResult{
			ConceptType:   interfaces.KnConceptTypeRelation,
			ConceptID:     conceptID,
			ConceptName:   conceptName,
			ConceptDetail: rt,
		})
	}

	return concepts
}

// filterByRerankScore 根据重排分数过滤概念
// 保留 RerankScore > 0 的概念
func (l *LocalKnSearch) filterByRerankScore(
	rerankedConcepts []*interfaces.ConceptResult,
	objectTypes []map[string]interface{},
	relationTypes []map[string]interface{},
) (filteredObjects, filteredRelations []map[string]interface{}) {
	// 收集分数 > 0 的概念ID
	relevantObjIDs := make(map[string]bool)
	relevantRelIDs := make(map[string]bool)

	for _, concept := range rerankedConcepts {
		if concept.RerankScore > 0 {
			if concept.ConceptType == interfaces.KnConceptTypeObject {
				relevantObjIDs[concept.ConceptID] = true
			} else if concept.ConceptType == interfaces.KnConceptTypeRelation {
				relevantRelIDs[concept.ConceptID] = true
			}
		}
	}

	// 如果都没有选中，返回原列表
	if len(relevantObjIDs) == 0 && len(relevantRelIDs) == 0 {
		return objectTypes, relationTypes
	}

	// 过滤对象类型
	for _, ot := range objectTypes {
		if id, ok := ot["id"].(string); ok && relevantObjIDs[id] {
			filteredObjects = append(filteredObjects, ot)
		}
	}

	// 过滤关系类型
	for _, rt := range relationTypes {
		if id, ok := rt["id"].(string); ok && relevantRelIDs[id] {
			filteredRelations = append(filteredRelations, rt)
		}
	}

	return filteredObjects, filteredRelations
}

// parseRetrievalConfig 解析检索配置
// 将 any 类型的配置转换为结构化的 RetrievalConfig
func (l *LocalKnSearch) parseRetrievalConfig(config any) *interfaces.RetrievalConfig {
	if config == nil {
		return nil
	}

	// 尝试类型断言
	if rc, ok := config.(*interfaces.RetrievalConfig); ok {
		return rc
	}

	// 尝试 JSON 转换（处理 map 类型）
	if configMap, ok := config.(map[string]interface{}); ok {
		bytes, err := json.Marshal(configMap)
		if err != nil {
			l.logger.Warnf("[LocalKnSearch#parseRetrievalConfig] Failed to marshal config: %v", err)
			return nil
		}
		var rc interfaces.RetrievalConfig
		if err := json.Unmarshal(bytes, &rc); err != nil {
			l.logger.Warnf("[LocalKnSearch#parseRetrievalConfig] Failed to unmarshal config: %v", err)
			return nil
		}
		return &rc
	}

	return nil
}
