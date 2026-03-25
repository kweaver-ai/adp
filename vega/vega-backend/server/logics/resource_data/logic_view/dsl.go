// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package logic_view

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/mitchellh/mapstructure"

	"vega-backend/interfaces"
)

// 三种情况需要拼接 dsl
// 1. 没有pit，有search_after
// 2. 有pit，有search_after
// 3. 有pit，没有search_after
func getSearchAfterDSL(searchAfterParams *interfaces.SearchAfterParams) (interfaces.DSLCfg, error) {
	var dsl interfaces.DSLCfg

	if searchAfterParams == nil {
		return dsl, nil
	}

	if len(searchAfterParams.SearchAfter) > 0 {
		dsl.SearchAfter = searchAfterParams.SearchAfter
	}

	// 设置pit
	if searchAfterParams.PitID != "" {
		dsl.Pit = &struct {
			ID        string `json:"id,omitempty"`
			KeepAlive string `json:"keep_alive,omitempty"`
		}{}
		dsl.Pit.ID = searchAfterParams.PitID
		if searchAfterParams.PitKeepAlive != "" {
			dsl.Pit.KeepAlive = searchAfterParams.PitKeepAlive
		}
	}

	return dsl, nil

}

func marshalDSL(dsl interfaces.DSLCfg) (bytes.Buffer, error) {
	// 序列化为JSON
	dslBytes, err := sonic.Marshal(dsl)
	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("data view marshal interfaces.DSLCfg error: %s", err.Error())
	}

	var queryBuffer bytes.Buffer
	queryBuffer.Write(dslBytes)

	// fmt.Println(queryBuffer.String())
	return queryBuffer, nil
}

// DSL生成器
func buildDSL(ctx context.Context, query interfaces.ResourceDataQueryParams, view *interfaces.LogicView,
	viewIndicesMap map[string][]string) (interfaces.DSLCfg, error) {
	sortParams := completeDSLSortParams(query.Sort, query.UseSearchAfter)

	var dsl interfaces.DSLCfg
	// 设置分页参数和track_total_hits
	dsl.From = query.Offset
	dsl.Size = query.Limit
	if query.NeedTotal {
		dsl.TrackTotalHits = true
	}

	if len(sortParams) > 0 {
		sort := []map[string]any{}
		for _, sp := range sortParams {
			if sp.Field == "" || sp.Direction == "" {
				return dsl, rest.NewHTTPError(ctx, http.StatusBadRequest,
					rest.PublicError_BadRequest).
					WithErrorDetails("The sort field and direction cannot be empty")
			}

			sortFieldName := sp.Field
			sortField, ok := view.FieldsMap[sp.Field]
			// 不校验排序字段是否在视图字段列表里，为_score字段排序开绿灯
			// if !ok {
			// 	return bytes.Buffer{}, rest.NewHTTPError(ctx, http.StatusForbidden,
			// 		uerrors.VegaBackend_LogicView_InvalidFieldPermission_Sort).
			// 		WithErrorDetails(fmt.Sprintf("The sort field '%s' is not in the view fields list", sp.Field))
			// }

			if ok {
				if sortField.Type == interfaces.DataType_Binary {
					return dsl, rest.NewHTTPError(ctx, http.StatusBadRequest,
						rest.PublicError_BadRequest).
						WithErrorDetails(fmt.Sprintf("The sort field '%s' is binary type, do not support sorting", sp.Field))
				}

				// text类型的字段需要看其下有没有配置keyword索引，配了就用 xxx.keyword 进行排序。否则不纳入排序
				// string类型的字段直接支持排序，若其有全文索引，则在字段的 keyword 下有 text
				if IsTextType(sortField) {
					if HasFeature(sortField, interfaces.PropertyFeatureType_Keyword) {
						sortFieldName = sortFieldName + ".keyword"
					} else {
						continue
					}
				}
			}

			// 需要将视图字段__score转为opensearch内置字段_score, 暂时不修改，兼容处理
			if sortFieldName == "__score" {
				sortFieldName = "_score"
			}

			sort = append(sort, map[string]any{
				sortFieldName: sp.Direction,
			})
		}

		dsl.Sort = sort
	}

	// 获取searchAfter参数
	searchAfterDSL, err := getSearchAfterDSL(nil)
	if err != nil {
		return dsl, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).
			WithErrorDetails(fmt.Sprintf("failed to get search after dsl, %s", err.Error()))
	}

	// 合并searchAfterDSL到主DSL结构体
	dsl.SearchAfter = searchAfterDSL.SearchAfter
	dsl.Pit = searchAfterDSL.Pit

	// 构建查询条件
	queryDSL, err := buildDSLQuery(ctx, view, viewIndicesMap)
	if err != nil {
		return dsl, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).
			WithErrorDetails(fmt.Sprintf("failed to build query dsl, %s", err.Error()))
	}

	// 合并查询条件到主DSL结构体
	dsl.Query = queryDSL.Query

	// 添加全局过滤条件，全局过滤条件的字段应该在视图字段列表里
	dsl, err = addGlobalFiltersToDSL(ctx, dsl, query.FilterCondCfg, view.FieldsMap)
	if err != nil {
		return dsl, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).
			WithErrorDetails(fmt.Sprintf("failed to add global filters to dsl, %s", err.Error()))
	}

	logger.Infof("view_indices_map is %v", viewIndicesMap)

	return dsl, nil
}

// 生成Resource节点的查询条件, 返回查询条件DSL, 是否需要计算分数, 错误信息
func buildResourceQuery(ctx context.Context, node *interfaces.LogicDefinitionNode, viewIndicesMap map[string][]string) (any, bool, error) {
	var cfg interfaces.ResourceNodeCfg
	err := mapstructure.Decode(node.Config, &cfg)
	if err != nil {
		return "", false, fmt.Errorf("failed to decode view node config, %s", err.Error())
	}

	if cfg.ResourceID == "" {
		return "", false, fmt.Errorf("resource id is empty")
	}

	indices, exists := viewIndicesMap[cfg.ResourceID]
	if !exists {
		return "", false, fmt.Errorf("no indices found for resource ID: %s", cfg.ResourceID)
	}

	indexConditions := map[string]any{
		"terms": map[string]any{
			"_index": indices,
		},
	}

	var filterCondition map[string]any
	// 使用原子视图的fieldsMap，包含索引库的全部字段
	// filterConditionStr, needScore, err := buildDSLCondition(ctx, cfg.Filters, cfg.Resource.FieldsMap)
	filterConditionStr, needScore, err := buildDSLCondition(ctx, cfg.Filters, nil)

	if err != nil {
		return "", false, err
	}

	if filterConditionStr != "" {
		if err := sonic.Unmarshal([]byte(filterConditionStr), &filterCondition); err != nil {
			return "", false, fmt.Errorf("failed to unmarshal filter condition, %s", err.Error())
		}
	}

	if filterCondition == nil {
		return indexConditions, false, nil
	}

	if needScore {
		return map[string]any{
			"bool": map[string]any{
				"must": []any{indexConditions, filterCondition},
			},
		}, true, nil
	}

	return map[string]any{
		"bool": map[string]any{
			"filter": []any{indexConditions, filterCondition},
		},
	}, false, nil
}

// 添加全局过滤条件到DSL
func addGlobalFiltersToDSL(ctx context.Context, dsl interfaces.DSLCfg, filters *interfaces.FilterCondCfg,
	fieldsMap map[string]*interfaces.ViewProperty) (interfaces.DSLCfg, error) {
	// condStr, needScore, err := buildDSLCondition(ctx, filters, fieldsMap)
	// if err != nil {
	// 	return dsl, err
	// }

	// if condStr != "" {
	// 	var filterCondition map[string]any
	// 	if err := sonic.Unmarshal([]byte(condStr), &filterCondition); err != nil {
	// 		return dsl, fmt.Errorf("failed to unmarshal filter condition, %s", err.Error())
	// 	}

	// 	// 如果需要打分，使用must查询
	// 	if needScore {
	// 		dsl.TrackScores = true
	// 		dsl.Query.Bool.Must = append(dsl.Query.Bool.Must, filterCondition)
	// 	} else {
	// 		dsl.Query.Bool.Filter = append(dsl.Query.Bool.Filter, filterCondition)
	// 	}
	// }

	// return dsl, nil
	return dsl, nil
}

func buildDSLQuery(ctx context.Context, view *interfaces.LogicView, viewIndicesMap map[string][]string) (interfaces.DSLCfg, error) {
	// 自定义视图logic definition不能为null
	if view.LogicDefinition == nil {
		return interfaces.DSLCfg{}, fmt.Errorf("logic definition is nil")
	}

	// 提取所有视图节点
	var viewNodes []*interfaces.LogicDefinitionNode
	for _, node := range view.LogicDefinition {
		switch node.Type {
		case interfaces.LogicDefinitionNodeType_Resource:
			viewNodes = append(viewNodes, node)
		case interfaces.LogicDefinitionNodeType_Union:
			var unionCfg *interfaces.UnionNodeCfg
			err := mapstructure.Decode(node.Config, &unionCfg)
			if err != nil {
				return interfaces.DSLCfg{}, fmt.Errorf("failed to decode union node config, %s", err.Error())
			}

			// interfaces.DSLCfg 类视图只允许配置 all
			if unionCfg.UnionType != interfaces.UnionType_All {
				return interfaces.DSLCfg{}, fmt.Errorf("unsupported union type: %s", unionCfg.UnionType)
			}
		case interfaces.LogicDefinitionNodeType_Output:
		default:
			return interfaces.DSLCfg{}, fmt.Errorf("unsupported node type: %s", node.Type)
		}
	}

	var dsl interfaces.DSLCfg
	// 根据视图节点数量决定查询结构
	if len(viewNodes) == 1 {
		// 单视图节点，直接使用filter，不用should
		query, trackScores, err := buildResourceQuery(ctx, viewNodes[0], viewIndicesMap)
		if err != nil {
			return interfaces.DSLCfg{}, err
		}
		dsl.Query.Bool.Filter = []any{query}
		dsl.TrackScores = trackScores

	} else {
		// 多视图节点，使用should,
		// track_scores逻辑：只要有一个视图节点需要track_scores，就设置为true
		trackScores := false
		shouldQueries := make([]any, 0, len(viewNodes))
		for _, node := range viewNodes {
			query, tScore, err := buildResourceQuery(ctx, node, viewIndicesMap)
			if err != nil {
				return interfaces.DSLCfg{}, err
			}
			shouldQueries = append(shouldQueries, query)

			if tScore {
				trackScores = true
			}
		}

		dsl.Query.Bool.Should = shouldQueries
		// 设置min_should_match为1，确保至少匹配一个should条件
		dsl.Query.Bool.MinShouldMatch = 1
		dsl.TrackScores = trackScores
	}

	return dsl, nil
}

// 构造过滤条件
func buildDSLCondition(ctx context.Context, cfg *interfaces.FilterCondCfg, fieldsMap map[string]*interfaces.ViewProperty) (string, bool, error) {
	// var dslStr string
	// // 将过滤条件拼接到 dsl 的 query 中
	// // 创建一个包含查询类型的上下文
	// ctx = context.WithValue(ctx, cond.CtxKey_QueryType, interfaces.QueryType_DSL)
	// CondCfg, needScore, err := cond.NewCondition(ctx, cfg, fieldsMap)
	// if err != nil {
	// 	return "", needScore, fmt.Errorf("failed to new condition, %s", err.Error())
	// }

	// if CondCfg != nil {
	// 	dslStr, err = CondCfg.Convert(ctx)
	// 	if err != nil {
	// 		return "", needScore, fmt.Errorf("failed to convert condition to dsl, %s", err.Error())
	// 	}
	// }

	// return dslStr, needScore, nil
	return "", false, nil
}

// 获取原子视图和索引列表的映射
func getViewIndicesMap(indices []string, baseTypeViewMap map[string]string) (map[string][]string, error) {
	// 创建视图ID到索引列表的映射结果
	viewIndicesMap := make(map[string][]string)

	// 遍历所有索引
	for _, index := range indices {
		// 按连字符拆分索引名，获取索引库（第二部分）
		parts := strings.Split(index, "-")
		if len(parts) < 2 {
			continue
		}

		baseType := parts[1]

		// 查找哪些视图ID关联了这个索引库
		if viewID, ok := baseTypeViewMap[baseType]; ok {
			// 初始化视图ID的索引列表（如果不存在）
			if _, exists := viewIndicesMap[viewID]; !exists {
				viewIndicesMap[viewID] = make([]string, 0)
			}
			viewIndicesMap[viewID] = append(viewIndicesMap[viewID], index)
		} else {
			return nil, fmt.Errorf("base type %s does not have a associated view", baseType)
		}
	}

	return viewIndicesMap, nil
}

// 补充 sort 字段
func completeDSLSortParams(sort []*interfaces.SortField, useSearchAfter bool) []*interfaces.SortField {
	defaultSort := []*interfaces.SortField{}
	if useSearchAfter {
		defaultSort = []*interfaces.SortField{
			{Field: "_id", Direction: interfaces.DESC_DIRECTION},
		}
	}

	sort = append(sort, defaultSort...)
	newSort := []*interfaces.SortField{}
	// 去重
	sortFieldSet := map[string]struct{}{}
	for _, sortParam := range sort {
		if _, ok := sortFieldSet[sortParam.Field]; !ok {
			newSort = append(newSort, sortParam)
			sortFieldSet[sortParam.Field] = struct{}{}
		}
	}

	return newSort
}

// 检查字段是否为 text 类型
func IsTextType(fieldInfo *interfaces.ViewProperty) bool {
	return fieldInfo != nil && fieldInfo.Type == interfaces.DataType_Text
}

// 检查字段特征是否包含指定特征
func HasFeature(fieldInfo *interfaces.ViewProperty, feature string) bool {
	for _, f := range fieldInfo.Features {
		if f.FeatureType == feature {
			return true
		}
	}
	return false
}
