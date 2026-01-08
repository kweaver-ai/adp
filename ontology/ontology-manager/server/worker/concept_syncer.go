package worker

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/mitchellh/mapstructure"

	"ontology-manager/common"
	"ontology-manager/interfaces"
	"ontology-manager/logics"
)

var (
	cSyncerOnce sync.Once
	cSyncer     *ConceptSyncer
)

type ConceptSyncer struct {
	appSetting *common.AppSetting
	ata        interfaces.ActionTypeAccess
	cga        interfaces.ConceptGroupAccess
	mfa        interfaces.ModelFactoryAccess
	kna        interfaces.KNAccess
	osa        interfaces.OpenSearchAccess
	ota        interfaces.ObjectTypeAccess
	rta        interfaces.RelationTypeAccess
}

func NewConceptSyncer(appSetting *common.AppSetting) *ConceptSyncer {
	cSyncerOnce.Do(func() {
		cSyncer = &ConceptSyncer{
			appSetting: appSetting,
			ata:        logics.ATA,
			mfa:        logics.MFA,
			kna:        logics.KNA,
			cga:        logics.CGA,
			osa:        logics.OSA,
			ota:        logics.OTA,
			rta:        logics.RTA,
		}
	})
	return cSyncer
}

// KNDetailInfo 知识网络详情信息结构
type KNDetailInfo struct {
	NetworkInfo   map[string]any `json:"network_info"`
	ObjectTypes   []SimpleItem   `json:"object_types"`
	RelationTypes []SimpleItem   `json:"relation_types"`
	ActionTypes   []SimpleItem   `json:"action_types"`
	ConceptGroups []SimpleItem   `json:"concept_groups"`
}

// SimpleItem 简化项结构，仅保留id、name、tag、comment字段
type SimpleItem struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Tags    []string `json:"tags"`
	Comment string   `json:"comment"`

	// for relation types
	SourceObjectTypeName string `json:"source_object_type_name,omitempty"`
	TargetObjectTypeName string `json:"target_object_type_name,omitempty"`

	// for action types
	ObjectTypeName string `json:"object_type_name,omitempty"`
}

// GeneratorTicker 生成业务知识网络详情定时任务
func (cs *ConceptSyncer) Start() {
	for {
		err := cs.handleKNs()
		if err != nil {
			logger.Errorf("[handleKNs] Failed: %v", err)
		}
		time.Sleep(5 * time.Minute)
	}
}

// handleKNs 处理业务知识网络详情 todo：补充 对象类、关系类、行动类的detail，并且要更新概念索引
func (cs *ConceptSyncer) handleKNs() error {
	defer func() {
		if rerr := recover(); rerr != nil {
			logger.Errorf("[handleKNs] Failed: %v", rerr)
			debug.PrintStack()
			return
		}
	}()

	logger.Debug("[handleKNs] Start")

	ctx := context.Background()

	knsInDB, err := cs.kna.GetAllKNs(ctx)
	if err != nil {
		logger.Errorf("Failed to list knowledge networks: %v", err)
		return err
	}

	knsInOS, err := cs.getAllKNsFromOpenSearch(ctx)
	if err != nil {
		logger.Errorf("Failed to list knowledge networks in OpenSearch: %v", err)
		return err
	}

	for _, knInDB := range knsInDB {
		need_update := false
		knInOS, exist := knsInOS[knInDB.KNID]
		if !exist {
			need_update = true
		} else if knInDB.UpdateTime != knInOS.UpdateTime {
			need_update = true
		}

		err := cs.handleKnowledgeNetwork(ctx, knInDB, need_update)
		if err != nil {
			logger.Errorf("Failed to handle knowledge network %s (%s %s): %v", knInDB.KNName, knInDB.KNID, knInDB.Branch, err)
			continue
		}
	}

	logger.Info("handle KNs completed")
	return nil
}

// handleKnowledgeNetwork 处理单个知识网络
func (cs *ConceptSyncer) handleKnowledgeNetwork(ctx context.Context, kn *interfaces.KN, need_update bool) error {
	logger.Debugf("Handle knowledge network: %s (%s %s), %s", kn.KNName, kn.KNID, kn.Branch)

	// 获取对象类型列表
	objectTypes, ot_need_update, err := cs.handleObjectTypes(ctx, kn.KNID, kn.Branch)
	if err != nil {
		logger.Errorf("Failed to handle object types %s %s: %v", kn.KNID, kn.Branch, err)
		return err
	}
	objectTypesMap := map[string]string{}
	for _, objectType := range objectTypes {
		objectTypesMap[objectType.ID] = objectType.Name
	}

	// 获取关系类型列表
	relationTypes, rt_need_update, err := cs.handleRelationTypes(ctx, kn.KNID, kn.Branch, objectTypesMap)
	if err != nil {
		logger.Errorf("Failed to handle relation types %s %s: %v", kn.KNID, kn.Branch, err)
		return err
	}

	// 获取行动类型列表
	actionTypes, at_need_update, err := cs.handleActionTypes(ctx, kn.KNID, kn.Branch, objectTypesMap)
	if err != nil {
		logger.Errorf("Failed to handle action types %s %s: %v", kn.KNID, kn.Branch, err)
		return err
	}

	conceptGroups, cg_need_update, err := cs.handleConceptGroups(ctx, kn.KNID, kn.Branch)
	if err != nil {
		logger.Errorf("Failed to handle concept groups %s %s: %v", kn.KNID, kn.Branch, err)
		return err
	}

	if !need_update && !ot_need_update && !rt_need_update && !at_need_update && !cg_need_update {
		logger.Debugf("Knowledge network %s (%s %s) does not need update", kn.KNName, kn.KNID, kn.Branch)
		return nil
	}

	// 创建知识网络详情信息
	knDetail := KNDetailInfo{
		NetworkInfo: map[string]any{
			"id":                   kn.KNID,
			"name":                 kn.KNName,
			"tags":                 kn.Tags,
			"comment":              kn.Comment,
			"object_types_count":   int64(len(objectTypes)),
			"relation_types_count": int64(len(relationTypes)),
			"action_types_count":   int64(len(actionTypes)),
			"concept_groups_count": int64(len(conceptGroups)),
		},
		ObjectTypes:   objectTypes,
		RelationTypes: relationTypes,
		ActionTypes:   actionTypes,
		ConceptGroups: conceptGroups,
	}

	// 转换为JSON字符串
	jsonData, err := sonic.MarshalString(knDetail)
	if err != nil {
		logger.Errorf("Failed to marshal KN detail to JSON: %v", err)
		return err
	}

	// 更新知识网络详情
	kn.Detail = jsonData
	err = cs.kna.UpdateKNDetail(ctx, kn.KNID, kn.Branch, jsonData)
	if err != nil {
		logger.Errorf("Failed to update KN detail for %s (%s %s): %v", kn.KNName, kn.KNID, kn.Branch, err)
		return err
	}

	err = cs.insertOpenSearchDataForKN(ctx, kn)
	if err != nil {
		logger.Errorf("Failed to insert open search data for KN %s (%s %s): %v", kn.KNName, kn.KNID, kn.Branch, err)
		return err
	}

	logger.Debugf("Generated KN detail for %s (%s %s): %s", kn.KNName, kn.KNID, kn.Branch, string(jsonData))
	return nil
}

// handleObjectTypes 获取知识网络的对象类型
func (cs *ConceptSyncer) handleObjectTypes(ctx context.Context, knID string, branch string) ([]SimpleItem, bool, error) {
	logger.Debugf("Handle object types for knowledge network %s %s", knID, branch)
	objectTypesInDB, err := cs.ota.GetAllObjectTypesByKnID(ctx, knID, branch)
	if err != nil {
		return []SimpleItem{}, false, err
	}

	objectTypesInOS, err := cs.getAllObjectTypesFromOpenSearchByKnID(ctx, knID, branch)
	if err != nil {
		return []SimpleItem{}, false, err
	}

	need_update := false
	add_list := []*interfaces.ObjectType{}
	for _, otInDB := range objectTypesInDB {
		otInOS, exist := objectTypesInOS[otInDB.OTID]
		if !exist {
			add_list = append(add_list, otInDB)
		} else if otInDB.UpdateTime != otInOS.UpdateTime {
			add_list = append(add_list, otInDB)
		}
	}
	if len(add_list) > 0 {
		logger.Debugf("Need add (%d) object types to OpenSearch", len(add_list))
		need_update = true
	}
	// TODO 获取opensearch 中 list
	// 对比list，判断是否需要更新

	err = cs.insertOpenSearchDataForObjectTypes(ctx, add_list)
	if err != nil {
		return []SimpleItem{}, false, err
	}

	// 简化为仅保留id、name、tag、comment字段
	simpleObjectTypes := make([]SimpleItem, 0, len(objectTypesInDB))
	for _, otInDB := range objectTypesInDB {
		simpleItem := SimpleItem{
			ID:      otInDB.OTID,
			Name:    otInDB.OTName,
			Tags:    otInDB.Tags,
			Comment: otInDB.Comment,
		}
		simpleObjectTypes = append(simpleObjectTypes, simpleItem)
	}

	logger.Debugf("Handle object types for knowledge network %s %s done", knID, branch)
	return simpleObjectTypes, need_update, nil
}

// handleRelationTypes 获取知识网络的关系类型
func (cs *ConceptSyncer) handleRelationTypes(ctx context.Context, knID string,
	branch string, objectTypesMap map[string]string) ([]SimpleItem, bool, error) {
	logger.Debugf("Handle relation types for knowledge network %s %s", knID, branch)
	relationTypesInDB, err := cs.rta.GetAllRelationTypesByKnID(ctx, knID, branch)
	if err != nil {
		return []SimpleItem{}, false, err
	}

	relationTypesInOS, err := cs.getAllRelationTypesFromOpenSearchByKnID(ctx, knID, branch)
	if err != nil {
		return []SimpleItem{}, false, err
	}

	need_update := false
	add_list := []*interfaces.RelationType{}
	for _, rtInDB := range relationTypesInDB {
		rtInOS, exist := relationTypesInOS[rtInDB.RTID]
		if !exist {
			add_list = append(add_list, rtInDB)
		} else if rtInDB.UpdateTime != rtInOS.UpdateTime {
			add_list = append(add_list, rtInDB)
		}
	}
	if len(add_list) > 0 {
		logger.Debugf("Need add (%d) relation types to OpenSearch", len(add_list))
		need_update = true
	}

	err = cs.insertOpenSearchDataForRelationTypes(ctx, add_list)
	if err != nil {
		return []SimpleItem{}, false, err
	}

	// 简化为仅保留id、name、tag、comment字段
	simpleRelationTypes := make([]SimpleItem, 0, len(relationTypesInDB))
	for _, rtInDB := range relationTypesInDB {
		sourceObjectTypeName := objectTypesMap[rtInDB.SourceObjectTypeID]
		targetObjectTypeName := objectTypesMap[rtInDB.TargetObjectTypeID]
		simpleItem := SimpleItem{
			ID:                   rtInDB.RTID,
			Name:                 rtInDB.RTName,
			Tags:                 rtInDB.Tags,
			Comment:              rtInDB.Comment,
			SourceObjectTypeName: sourceObjectTypeName,
			TargetObjectTypeName: targetObjectTypeName,
		}
		simpleRelationTypes = append(simpleRelationTypes, simpleItem)
	}

	logger.Debugf("Handle relation types for knowledge network %s %s done", knID, branch)
	return simpleRelationTypes, need_update, nil
}

// handleActionTypes 获取知识网络的行动类型
func (cs *ConceptSyncer) handleActionTypes(ctx context.Context, knID string,
	branch string, objectTypesMap map[string]string) ([]SimpleItem, bool, error) {
	logger.Debugf("Handle action types for knowledge network %s %s", knID, branch)
	actionTypesInDB, err := cs.ata.GetAllActionTypesByKnID(ctx, knID, branch)
	if err != nil {
		return []SimpleItem{}, false, err
	}

	actionTypesInOS, err := cs.getAllActionTypesFromOpenSearchByKnID(ctx, knID, branch)
	if err != nil {
		return []SimpleItem{}, false, err
	}

	need_update := false
	add_list := []*interfaces.ActionType{}
	for _, atInDB := range actionTypesInDB {
		atInOS, exist := actionTypesInOS[atInDB.ATID]
		if !exist {
			add_list = append(add_list, atInDB)
		} else if atInDB.UpdateTime != atInOS.UpdateTime {
			add_list = append(add_list, atInDB)
		}
	}
	if len(add_list) > 0 {
		logger.Debugf("Need add (%d) action types to OpenSearch", len(add_list))
		need_update = true
	}

	err = cs.insertOpenSearchDataForActionTypes(ctx, add_list)
	if err != nil {
		return []SimpleItem{}, false, err
	}

	// 简化为仅保留id、name、tag、comment字段
	simpleActionTypes := make([]SimpleItem, 0, len(actionTypesInDB))
	for _, atInDB := range actionTypesInDB {
		objectTypeName := objectTypesMap[atInDB.ObjectTypeID]
		simpleItem := SimpleItem{
			ID:             atInDB.ATID,
			Name:           atInDB.ATName,
			Tags:           atInDB.Tags,
			Comment:        atInDB.Comment,
			ObjectTypeName: objectTypeName,
		}
		simpleActionTypes = append(simpleActionTypes, simpleItem)
	}

	logger.Debugf("Handle action types for knowledge network %s %s done", knID, branch)
	return simpleActionTypes, need_update, nil
}

// handleConceptGroups 获取知识网络的概念组
func (cs *ConceptSyncer) handleConceptGroups(ctx context.Context, knID string, branch string) ([]SimpleItem, bool, error) {
	logger.Debugf("Handle concept groups for knowledge network %s %s", knID, branch)
	conceptGroupsInDB, err := cs.cga.GetAllConceptGroupsByKnID(ctx, knID, branch)
	if err != nil {
		return []SimpleItem{}, false, err
	}

	conceptGroupsInOS, err := cs.getAllConceptGroupsFromOpenSearchByKnID(ctx, knID, branch)
	if err != nil {
		return []SimpleItem{}, false, err
	}

	need_update := false
	add_list := []*interfaces.ConceptGroup{}
	for _, cgInDB := range conceptGroupsInDB {
		cgInOS, exist := conceptGroupsInOS[cgInDB.CGID]
		if !exist {
			add_list = append(add_list, cgInDB)
		} else if cgInDB.UpdateTime != cgInOS.UpdateTime {
			add_list = append(add_list, cgInDB)
		}
	}
	if len(add_list) > 0 {
		logger.Debugf("Need add (%d) concept groups to OpenSearch", len(add_list))
		need_update = true
	}

	// TODO 获取opensearch 中 list
	// 对比list，判断是否需要更新

	err = cs.insertOpenSearchDataForConceptGroups(ctx, add_list)
	if err != nil {
		return []SimpleItem{}, false, err
	}

	// 简化为仅保留id、name、tag、comment字段
	simpleConceptGroups := make([]SimpleItem, 0, len(conceptGroupsInDB))
	for _, cgInDB := range conceptGroupsInDB {
		simpleItem := SimpleItem{
			ID:      cgInDB.CGID,
			Name:    cgInDB.CGName,
			Tags:    cgInDB.Tags,
			Comment: cgInDB.Comment,
		}
		simpleConceptGroups = append(simpleConceptGroups, simpleItem)
	}

	logger.Debugf("Handle concept groups for knowledge network %s %s done", knID, branch)
	return simpleConceptGroups, need_update, nil
}

func (cs *ConceptSyncer) insertOpenSearchDataForKN(ctx context.Context, kn *interfaces.KN) error {
	if cs.appSetting.ServerSetting.DefaultSmallModelEnabled {
		words := []string{kn.KNName}
		words = append(words, kn.Tags...)
		words = append(words, kn.Comment, kn.Detail)
		word := strings.Join(words, "\n")

		defaultModel, err := cs.mfa.GetDefaultModel(ctx)
		if err != nil {
			logger.Errorf("GetDefaultModel error: %s", err.Error())
			return err
		}
		vectors, err := cs.mfa.GetVector(ctx, defaultModel, []string{word})
		if err != nil {
			logger.Errorf("GetVector error: %s", err.Error())
			return err
		}

		kn.Vector = vectors[0].Vector
	}

	docid := interfaces.GenerateConceptDocuemtnID(kn.KNID, interfaces.MODULE_TYPE_KN, kn.KNID, kn.Branch)
	kn.ModuleType = interfaces.MODULE_TYPE_KN

	err := cs.osa.InsertData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, docid, kn)
	if err != nil {
		logger.Errorf("InsertData error: %s", err.Error())
		return err
	}

	return nil
}

func (cs *ConceptSyncer) insertOpenSearchDataForObjectTypes(ctx context.Context, objectTypes []*interfaces.ObjectType) error {
	if len(objectTypes) == 0 {
		return nil
	}

	if cs.appSetting.ServerSetting.DefaultSmallModelEnabled {
		words := []string{}
		for _, objectType := range objectTypes {
			arr := []string{objectType.OTName}
			arr = append(arr, objectType.Tags...)
			arr = append(arr, objectType.Comment, objectType.Detail)
			word := strings.Join(arr, "\n")
			words = append(words, word)
		}

		dftModel, err := cs.mfa.GetDefaultModel(ctx)
		if err != nil {
			logger.Errorf("GetDefaultModel error: %s", err.Error())
			return err
		}
		vectors, err := cs.mfa.GetVector(ctx, dftModel, words)
		if err != nil {
			logger.Errorf("GetVector error: %s", err.Error())
			return err
		}

		if len(vectors) != len(objectTypes) {
			logger.Errorf("GetVector error: expect vectors num is [%d], actual vectors num is [%d]", len(objectTypes), len(vectors))
			return fmt.Errorf("GetVector error: expect vectors num is [%d], actual vectors num is [%d]", len(objectTypes), len(vectors))
		}

		for i, objectType := range objectTypes {
			objectType.Vector = vectors[i].Vector
		}
	}

	for _, objectType := range objectTypes {
		docid := interfaces.GenerateConceptDocuemtnID(objectType.KNID, interfaces.MODULE_TYPE_OBJECT_TYPE,
			objectType.OTID, objectType.Branch)
		objectType.ModuleType = interfaces.MODULE_TYPE_OBJECT_TYPE

		err := cs.osa.InsertData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, docid, objectType)
		if err != nil {
			logger.Errorf("InsertData error: %s", err.Error())
			return err
		}
	}
	return nil
}

func (cs *ConceptSyncer) insertOpenSearchDataForActionTypes(ctx context.Context, actionTypes []*interfaces.ActionType) error {
	if len(actionTypes) == 0 {
		return nil
	}

	if cs.appSetting.ServerSetting.DefaultSmallModelEnabled {
		words := []string{}
		for _, actionType := range actionTypes {
			arr := []string{actionType.ATName}
			arr = append(arr, actionType.Tags...)
			arr = append(arr, actionType.Comment, actionType.Detail)
			word := strings.Join(arr, "\n")
			words = append(words, word)
		}

		dftModel, err := cs.mfa.GetDefaultModel(ctx)
		if err != nil {
			logger.Errorf("GetDefaultModel error: %s", err.Error())
			return err
		}
		vectors, err := cs.mfa.GetVector(ctx, dftModel, words)
		if err != nil {
			logger.Errorf("GetVector error: %s", err.Error())
			return err
		}

		if len(vectors) != len(actionTypes) {
			logger.Errorf("GetVector error: expect vectors num is [%d], actual vectors num is [%d]", len(actionTypes), len(vectors))
			return fmt.Errorf("GetVector error: expect vectors num is [%d], actual vectors num is [%d]", len(actionTypes), len(vectors))
		}

		for i, actionType := range actionTypes {
			actionType.Vector = vectors[i].Vector
		}
	}

	for _, actionType := range actionTypes {
		docid := interfaces.GenerateConceptDocuemtnID(actionType.KNID, interfaces.MODULE_TYPE_ACTION_TYPE,
			actionType.ATID, actionType.Branch)
		actionType.ModuleType = interfaces.MODULE_TYPE_ACTION_TYPE

		err := cs.osa.InsertData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, docid, actionType)
		if err != nil {
			logger.Errorf("InsertData error: %s", err.Error())
			return err
		}
	}
	return nil
}

func (cs *ConceptSyncer) insertOpenSearchDataForRelationTypes(ctx context.Context, relationTypes []*interfaces.RelationType) error {
	if len(relationTypes) == 0 {
		return nil
	}

	if cs.appSetting.ServerSetting.DefaultSmallModelEnabled {
		words := []string{}
		for _, relationType := range relationTypes {
			arr := []string{relationType.RTName}
			arr = append(arr, relationType.Tags...)
			arr = append(arr, relationType.Comment, relationType.Detail)
			word := strings.Join(arr, "\n")
			words = append(words, word)
		}

		dftModel, err := cs.mfa.GetDefaultModel(ctx)
		if err != nil {
			logger.Errorf("GetDefaultModel error: %s", err.Error())
			return err
		}
		vectors, err := cs.mfa.GetVector(ctx, dftModel, words)
		if err != nil {
			logger.Errorf("GetVector error: %s", err.Error())
			return err
		}

		if len(vectors) != len(relationTypes) {
			logger.Errorf("GetVector error: expect vectors num is [%d], actual vectors num is [%d]", len(relationTypes), len(vectors))
			return fmt.Errorf("GetVector error: expect vectors num is [%d], actual vectors num is [%d]", len(relationTypes), len(vectors))
		}

		for i, relationType := range relationTypes {
			relationType.Vector = vectors[i].Vector
		}
	}

	for _, relationType := range relationTypes {
		docid := interfaces.GenerateConceptDocuemtnID(relationType.KNID, interfaces.MODULE_TYPE_RELATION_TYPE,
			relationType.RTID, relationType.Branch)
		relationType.ModuleType = interfaces.MODULE_TYPE_RELATION_TYPE

		err := cs.osa.InsertData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, docid, relationType)
		if err != nil {
			logger.Errorf("InsertData error: %s", err.Error())
			return err
		}
	}
	return nil
}

func (cs *ConceptSyncer) insertOpenSearchDataForConceptGroups(ctx context.Context, conceptGroups []*interfaces.ConceptGroup) error {
	if len(conceptGroups) == 0 {
		return nil
	}

	if cs.appSetting.ServerSetting.DefaultSmallModelEnabled {
		words := []string{}
		for _, conceptGroup := range conceptGroups {
			arr := []string{conceptGroup.CGName}
			arr = append(arr, conceptGroup.Tags...)
			arr = append(arr, conceptGroup.Comment, conceptGroup.Detail)
			word := strings.Join(arr, "\n")
			words = append(words, word)
		}

		dftModel, err := cs.mfa.GetDefaultModel(ctx)
		if err != nil {
			logger.Errorf("GetDefaultModel error: %s", err.Error())
			return err
		}
		vectors, err := cs.mfa.GetVector(ctx, dftModel, words)
		if err != nil {
			logger.Errorf("GetVector error: %s", err.Error())
			return err
		}

		if len(vectors) != len(conceptGroups) {
			logger.Errorf("GetVector error: expect vectors num is [%d], actual vectors num is [%d]", len(conceptGroups), len(vectors))
			return fmt.Errorf("GetVector error: expect vectors num is [%d], actual vectors num is [%d]", len(conceptGroups), len(vectors))
		}

		for i, conceptGroup := range conceptGroups {
			conceptGroup.Vector = vectors[i].Vector
		}
	}

	for _, conceptGroup := range conceptGroups {
		docid := interfaces.GenerateConceptDocuemtnID(conceptGroup.KNID, interfaces.MODULE_TYPE_CONCEPT_GROUP,
			conceptGroup.CGID, conceptGroup.Branch)
		conceptGroup.ModuleType = interfaces.MODULE_TYPE_CONCEPT_GROUP

		err := cs.osa.InsertData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, docid, conceptGroup)
		if err != nil {
			logger.Errorf("InsertData error: %s", err.Error())
			return err
		}
	}
	return nil
}

func (cs *ConceptSyncer) getAllKNsFromOpenSearch(ctx context.Context) (map[string]*interfaces.KN, error) {
	query := map[string]any{
		"size": 10000,
		"query": map[string]any{
			"bool": map[string]any{
				"filter": []any{
					map[string]any{
						"term": map[string]any{
							"module_type": interfaces.MODULE_TYPE_KN,
						},
					},
				},
			},
		},
	}

	hits, err := cs.osa.SearchData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, query)
	if err != nil {
		return map[string]*interfaces.KN{}, err
	}

	kns := map[string]*interfaces.KN{}
	for _, hit := range hits {

		kn := interfaces.KN{}
		err := mapstructure.Decode(hit.Source, &kn)
		if err != nil {
			return map[string]*interfaces.KN{}, err
		}

		kns[kn.KNID] = &kn
	}

	return kns, nil
}

func (cs *ConceptSyncer) getAllObjectTypesFromOpenSearchByKnID(ctx context.Context,
	knID string, branch string) (map[string]*interfaces.ObjectType, error) {

	query := map[string]any{
		"size": 10000,
		"query": map[string]any{
			"bool": map[string]any{
				"filter": []any{
					map[string]any{
						"term": map[string]any{
							"kn_id": knID,
						},
					},
					map[string]any{
						"term": map[string]any{
							"branch": branch,
						},
					},
					map[string]any{
						"term": map[string]any{
							"module_type": interfaces.MODULE_TYPE_OBJECT_TYPE,
						},
					},
				},
			},
		},
	}

	hits, err := cs.osa.SearchData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, query)
	if err != nil {
		return map[string]*interfaces.ObjectType{}, err
	}

	objectTypes := map[string]*interfaces.ObjectType{}
	for _, hit := range hits {

		objectType := interfaces.ObjectType{}
		err := mapstructure.Decode(hit.Source, &objectType)
		if err != nil {
			return map[string]*interfaces.ObjectType{}, err
		}

		objectTypes[objectType.OTID] = &objectType
	}

	return objectTypes, nil
}

func (cs *ConceptSyncer) getAllRelationTypesFromOpenSearchByKnID(ctx context.Context,
	knID string, branch string) (map[string]*interfaces.RelationType, error) {

	query := map[string]any{
		"size": 10000,
		"query": map[string]any{
			"bool": map[string]any{
				"filter": []any{
					map[string]any{
						"term": map[string]any{
							"kn_id": knID,
						},
					},
					map[string]any{
						"term": map[string]any{
							"branch": branch,
						},
					},
					map[string]any{
						"term": map[string]any{
							"module_type": interfaces.MODULE_TYPE_RELATION_TYPE,
						},
					},
				},
			},
		},
	}

	hits, err := cs.osa.SearchData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, query)
	if err != nil {
		return map[string]*interfaces.RelationType{}, err
	}

	relationTypes := map[string]*interfaces.RelationType{}
	for _, hit := range hits {

		relationType := interfaces.RelationType{}
		err = mapstructure.Decode(hit.Source, &relationType)
		if err != nil {
			return map[string]*interfaces.RelationType{}, err
		}

		relationTypes[relationType.RTID] = &relationType
	}

	return relationTypes, nil
}

func (cs *ConceptSyncer) getAllActionTypesFromOpenSearchByKnID(ctx context.Context,
	knID string, branch string) (map[string]*interfaces.ActionType, error) {

	query := map[string]any{
		"size": 10000,
		"query": map[string]any{
			"bool": map[string]any{
				"filter": []any{
					map[string]any{
						"term": map[string]any{
							"kn_id": knID,
						},
					},
					map[string]any{
						"term": map[string]any{
							"branch": branch,
						},
					},
					map[string]any{
						"term": map[string]any{
							"module_type": interfaces.MODULE_TYPE_ACTION_TYPE,
						},
					},
				},
			},
		},
	}

	hits, err := cs.osa.SearchData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, query)
	if err != nil {
		return map[string]*interfaces.ActionType{}, err
	}

	actionTypes := map[string]*interfaces.ActionType{}
	for _, hit := range hits {

		actionType := interfaces.ActionType{}
		err = mapstructure.Decode(hit.Source, &actionType)
		if err != nil {
			return map[string]*interfaces.ActionType{}, err
		}

		actionTypes[actionType.ATID] = &actionType
	}

	return actionTypes, nil
}

func (cs *ConceptSyncer) getAllConceptGroupsFromOpenSearchByKnID(ctx context.Context,
	knID string, branch string) (map[string]*interfaces.ConceptGroup, error) {

	query := map[string]any{
		"size": 10000,
		"query": map[string]any{
			"bool": map[string]any{
				"filter": []any{
					map[string]any{
						"term": map[string]any{
							"kn_id": knID,
						},
					},
					map[string]any{
						"term": map[string]any{
							"branch": branch,
						},
					},
					map[string]any{
						"term": map[string]any{
							"module_type": interfaces.MODULE_TYPE_CONCEPT_GROUP,
						},
					},
				},
			},
		},
	}

	hits, err := cs.osa.SearchData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, query)
	if err != nil {
		return map[string]*interfaces.ConceptGroup{}, err
	}

	conceptGroups := map[string]*interfaces.ConceptGroup{}
	for _, hit := range hits {

		conceptGroup := interfaces.ConceptGroup{}
		err := mapstructure.Decode(hit.Source, &conceptGroup)
		if err != nil {
			return map[string]*interfaces.ConceptGroup{}, err
		}

		conceptGroups[conceptGroup.CGID] = &conceptGroup
	}

	return conceptGroups, nil
}
