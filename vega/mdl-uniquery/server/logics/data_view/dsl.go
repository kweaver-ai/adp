package data_view

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

	cond "uniquery/common/condition"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	dtype "uniquery/interfaces/data_type"
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
func buildDSL(ctx context.Context, query interfaces.ViewQueryInterface, view *interfaces.DataView,
	viewIndicesMap map[string][]string) (interfaces.DSLCfg, error) {
	queryParams := query.GetCommonParams()
	sortParams := query.GetSortParams()
	sortParams = completeDSLSortParams(sortParams, queryParams.UseSearchAfter, view.QueryType)

	var dsl interfaces.DSLCfg
	// 设置分页参数和track_total_hits
	dsl.From = queryParams.Offset
	dsl.Size = queryParams.Limit
	if view.QueryType == interfaces.QueryType_DSL && queryParams.NeedTotal {
		dsl.TrackTotalHits = true
	}

	if len(sortParams) > 0 {
		sort := []map[string]any{}
		for _, sp := range sortParams {
			if sp.Field == "" || sp.Direction == "" {
				return dsl, rest.NewHTTPError(ctx, http.StatusBadRequest,
					uerrors.Uniquery_DataView_InvalidParameter_Sort).
					WithErrorDetails("The sort field and direction cannot be empty")
			}

			sortFieldName := sp.Field
			sortField, ok := view.FieldsMap[sp.Field]
			// 不校验排序字段是否在视图字段列表里，为_score字段排序开绿灯
			// if !ok {
			// 	return bytes.Buffer{}, rest.NewHTTPError(ctx, http.StatusForbidden,
			// 		uerrors.Uniquery_DataView_InvalidFieldPermission_Sort).
			// 		WithErrorDetails(fmt.Sprintf("The sort field '%s' is not in the view fields list", sp.Field))
			// }

			if ok {
				if sortField.Type == dtype.DataType_Binary {
					return dsl, rest.NewHTTPError(ctx, http.StatusBadRequest,
						uerrors.Uniquery_DataView_BinaryFieldSortNotSupported).
						WithErrorDetails(fmt.Sprintf("The sort field '%s' is binary type, do not support sorting", sp.Field))
				}

				// text类型的字段需要看其下有没有配置keyword索引，配了就用 xxx.keyword 进行排序。否则不纳入排序
				// string类型的字段直接支持排序，若其有全文索引，则在字段的 keyword 下有 text
				if cond.IsTextType(sortField) {
					if cond.HasFeature(sortField, cond.FieldFeatureType_Keyword) {
						sortFieldName = sortFieldName + "." + dtype.KEYWORD_SUFFIX
					} else {
						continue
					}
				}
			}

			// 不需要将视图字段__id转为opensearch内置字段_id, 因为新的管道数据里已经存了 __id
			// if sortFieldName == "__id" {
			// 	sortFieldName = "_id"
			// }

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
	searchAfterDSL, err := getSearchAfterDSL(query.GetSearchAfterParams())
	if err != nil {
		return dsl, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			uerrors.Uniquery_DataView_InternalError_ConvertSearchAfterToDSLFailed).
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

	// 添加时间范围过滤条件
	timeRangeFilter := map[string]any{
		"range": map[string]any{
			"@timestamp": map[string]int64{},
		},
	}

	timeRangeMap := timeRangeFilter["range"].(map[string]any)["@timestamp"].(map[string]int64)
	if queryParams.Start != 0 {
		timeRangeMap["gte"] = queryParams.Start
	}
	if queryParams.End != 0 {
		timeRangeMap["lte"] = queryParams.End
	}

	// 只有当有时间范围条件时才添加
	if len(timeRangeMap) > 0 {
		dsl.Query.Bool.Filter = append(dsl.Query.Bool.Filter, timeRangeFilter)
	}

	// 添加全局过滤条件，全局过滤条件的字段应该在视图字段列表里
	dsl, err = addGlobalFiltersToDSL(ctx, dsl, query.GetGlobalFilters(), view.FieldsMap, view.Type)
	if err != nil {
		return dsl, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).
			WithErrorDetails(fmt.Sprintf("failed to add global filters to dsl, %s", err.Error()))
	}

	// 添加行列规则，多个行列规则的应用之间是or的关系
	// 在这里添加，不会应用在指标模型上，因为指标模型只需用到DSL Query 部分
	dsl, newFields, newFieldsMap, err := addRowColumnRulesToDSL(ctx, dsl, query.GetRowColumnRules(), view)
	if err != nil {
		return dsl, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).
			WithErrorDetails(fmt.Sprintf("failed to add row column rules to dsl, %s", err.Error()))
	}

	// 更新视图字段列表为行列规则限制后的字段列表
	defer func() {
		view.Fields = newFields
		view.FieldsMap = newFieldsMap
	}()

	logger.Infof("view_indices_map is %v", viewIndicesMap)

	return dsl, nil
}

// 生成视图节点的查询条件, 返回查询条件DSL, 是否需要计算分数, 错误信息
func buildViewQuery(ctx context.Context, node *interfaces.DataScopeNode, viewIndicesMap map[string][]string) (any, bool, error) {
	var cfg interfaces.ViewNodeCfg
	err := mapstructure.Decode(node.Config, &cfg)
	if err != nil {
		return "", false, fmt.Errorf("failed to decode view node config, %s", err.Error())
	}

	if cfg.View == nil {
		return "", false, fmt.Errorf("view is nil")
	}

	indices, exists := viewIndicesMap[cfg.ViewID]
	if !exists {
		return "", false, fmt.Errorf("no indices found for view ID: %s", cfg.ViewID)
	}

	indexConditions := map[string]any{
		"terms": map[string]any{
			"_index": indices,
		},
	}

	var filterCondition map[string]any
	// 使用原子视图的fieldsMap，包含索引库的全部字段
	filterConditionStr, needScore, err := buildDSLCondition(ctx, cfg.Filters, interfaces.ViewType_Custom, cfg.View.FieldsMap)
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
func addGlobalFiltersToDSL(ctx context.Context, dsl interfaces.DSLCfg, filters *cond.CondCfg, fieldsMap map[string]*cond.ViewField, viewType string) (interfaces.DSLCfg, error) {
	condStr, needScore, err := buildDSLCondition(ctx, filters, viewType, fieldsMap)
	if err != nil {
		return dsl, err
	}

	if condStr != "" {
		var filterCondition map[string]any
		if err := sonic.Unmarshal([]byte(condStr), &filterCondition); err != nil {
			return dsl, fmt.Errorf("failed to unmarshal filter condition, %s", err.Error())
		}

		// 如果需要打分，使用must查询
		if needScore {
			dsl.TrackScores = true
			dsl.Query.Bool.Must = append(dsl.Query.Bool.Must, filterCondition)
		} else {
			dsl.Query.Bool.Filter = append(dsl.Query.Bool.Filter, filterCondition)
		}
	}

	return dsl, nil
}

func addRowColumnRulesToDSL(ctx context.Context, dsl interfaces.DSLCfg, rules []*interfaces.DataViewRowColumnRule,
	view *interfaces.DataView) (interfaces.DSLCfg, []*cond.ViewField, map[string]*cond.ViewField, error) {

	// 行列规则长度为0， 可能查的是全量数据
	if len(rules) == 0 {
		return dsl, view.Fields, view.FieldsMap, nil
	}

	mergedFields := make([]*cond.ViewField, 0)
	mergedFieldsMap := map[string]*cond.ViewField{}
	for _, rule := range rules {
		// 判断列是否在视图字段列表里
		for _, field := range rule.Fields {
			if _, exists := view.FieldsMap[field]; !exists {
				return dsl, mergedFields, mergedFieldsMap, fmt.Errorf("field %s not found in view fields map", field)
			}

			vf := view.FieldsMap[field]
			mergedFields = append(mergedFields, vf)
			mergedFieldsMap[field] = vf
		}
	}

	mergedRowFilters := make([]*cond.CondCfg, 0)
	for _, rule := range rules {
		// 合并行规则
		if isValidFilters(rule.RowFilters) {
			mergedRowFilters = append(mergedRowFilters, rule.RowFilters)
		}
	}

	var finalCond *cond.CondCfg
	if len(mergedRowFilters) == 0 {
		finalCond = nil
	} else if len(mergedRowFilters) == 1 {
		finalCond = mergedRowFilters[0]
	} else {
		// 行列规则之间是 or 关系
		finalCond = &cond.CondCfg{
			Operation: cond.OperationOr,
			SubConds:  mergedRowFilters,
		}
	}

	// 构建行规则的DSL条件
	condStr, needScore, err := buildDSLCondition(ctx, finalCond, view.Type, view.FieldsMap)
	if err != nil {
		return dsl, mergedFields, mergedFieldsMap, err
	}

	if condStr != "" {
		var rowFilterCondition map[string]any
		if err := sonic.Unmarshal([]byte(condStr), &rowFilterCondition); err != nil {
			return dsl, mergedFields, mergedFieldsMap, fmt.Errorf("failed to unmarshal filter condition, %s", err.Error())
		}

		// 如果需要打分，使用must查询
		if needScore {
			dsl.TrackScores = true
			dsl.Query.Bool.Must = append(dsl.Query.Bool.Must, rowFilterCondition)
		} else {
			dsl.Query.Bool.Filter = append(dsl.Query.Bool.Filter, rowFilterCondition)
		}
	}

	return dsl, mergedFields, mergedFieldsMap, nil
}

func buildDSLQuery(ctx context.Context, view *interfaces.DataView, viewIndicesMap map[string][]string) (interfaces.DSLCfg, error) {
	// globalFilters := query.GetGlobalFilters()

	// 原子视图的时候查询只有全局过滤条件
	if view.Type == interfaces.ViewType_Atomic {
		return interfaces.DSLCfg{}, nil
		// return buildAtomicViewQuery(view, globalFilters)
	}

	// 自定义视图data scope不能为null
	if view.DataScope == nil {
		return interfaces.DSLCfg{}, fmt.Errorf("data scope is nil")
	}

	// 提取所有视图节点
	var viewNodes []*interfaces.DataScopeNode
	for _, node := range view.DataScope {
		switch node.Type {
		case interfaces.DataScopeNodeType_View:
			viewNodes = append(viewNodes, node)
		case interfaces.DataScopeNodeType_Union:
			var unionCfg *interfaces.UnionNodeCfg
			err := mapstructure.Decode(node.Config, &unionCfg)
			if err != nil {
				return interfaces.DSLCfg{}, fmt.Errorf("failed to decode union node config, %s", err.Error())
			}

			// interfaces.DSLCfg 类视图只允许配置 all
			if unionCfg.UnionType != interfaces.UnionType_All {
				return interfaces.DSLCfg{}, fmt.Errorf("unsupported union type: %s", unionCfg.UnionType)
			}
		case interfaces.DataScopeNodeType_Output:
		default:
			return interfaces.DSLCfg{}, fmt.Errorf("unsupported node type: %s", node.Type)
		}
	}

	var dsl interfaces.DSLCfg
	// 根据视图节点数量决定查询结构
	if len(viewNodes) == 1 {
		// 单视图节点，直接使用filter，不用should
		query, trackScores, err := buildViewQuery(ctx, viewNodes[0], viewIndicesMap)
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
			query, tScore, err := buildViewQuery(ctx, node, viewIndicesMap)
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

// // 构造原子视图查询
// func buildAtomicViewQuery(view *interfaces.DataView, globalFilters *cond.CondCfg) (interfaces.DSLCfg, error) {
// 	var dsl interfaces.DSLCfg
// 	if globalFilters != nil {
// 		cfg := globalFilters

// 		// 将过滤条件拼接到 dsl 的 query 中, 原子视图对应的是所属索引库的全部字段
// 		CondCfg, needScore, err := cond.NewCondition(cfg, interfaces.ViewType_Atomic, view.FieldsMap)
// 		if err != nil {
// 			return dsl, fmt.Errorf("failed to new condition, %s", err.Error())
// 		}

// 		if CondCfg != nil {
// 			condStr, err := CondCfg.Convert()
// 			if err != nil {
// 				return dsl, fmt.Errorf("failed to convert condition to dsl, %s", err.Error())
// 			}

// 			// 将条件字符串解析为map
// 			var conditionMap map[string]any
// 			if err := sonic.Unmarshal([]byte(condStr), &conditionMap); err != nil {
// 				return dsl, fmt.Errorf("failed to unmarshal condition: %s", err.Error())
// 			}

// 			// 如果需要打分，使用must查询
// 			if needScore {
// 				dsl.Query.Bool.Must = append(dsl.Query.Bool.Must, conditionMap)
// 			} else {
// 				dsl.Query.Bool.Filter = append(dsl.Query.Bool.Filter, conditionMap)
// 			}

// 		}
// 	}

// 	return dsl, nil

// }

// 构造过滤条件
func buildDSLCondition(ctx context.Context, cfg *cond.CondCfg, vType string, fieldsMap map[string]*cond.ViewField) (string, bool, error) {
	var dslStr string
	// 将过滤条件拼接到 dsl 的 query 中
	// 创建一个包含查询类型的上下文
	ctx = context.WithValue(ctx, cond.CtxKey_QueryType, interfaces.QueryType_DSL)
	CondCfg, needScore, err := cond.NewCondition(ctx, cfg, vType, fieldsMap)
	if err != nil {
		return "", needScore, fmt.Errorf("failed to new condition, %s", err.Error())
	}

	if CondCfg != nil {
		dslStr, err = CondCfg.Convert(ctx)
		if err != nil {
			return "", needScore, fmt.Errorf("failed to convert condition to dsl, %s", err.Error())
		}
	}

	return dslStr, needScore, nil
}

// 多个视图
// // 构建最终的查询
// finalQuery := map[string]any{
// 	"bool": map[string]any{
// 		"should":               shouldClauses,
// 		"minimum_should_match": 1,
// 	},
// }

// // 将map序列化为JSON字符串
// resultJSON, err := sonic.Marshal(finalQuery)
// if err != nil {
// 	return "", fmt.Errorf("failed to marshal query to JSON: %w", err)
// }

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
func completeDSLSortParams(sort []*interfaces.SortParamsV2, useSearchAfter bool, queryType string) []*interfaces.SortParamsV2 {
	var defaultSort []*interfaces.SortParamsV2
	switch queryType {
	case interfaces.QueryType_IndexBase:
		// 如果使用 search_after 分页, 补全 tiebreaker 字段排序
		if useSearchAfter {
			defaultSort = []*interfaces.SortParamsV2{
				{Field: interfaces.MetaField_Timestamp, Direction: interfaces.DESC_DIRECTION},
				{Field: interfaces.MetaField_ID, Direction: interfaces.DESC_DIRECTION},
			}
		} else {
			defaultSort = []*interfaces.SortParamsV2{
				{Field: interfaces.MetaField_Timestamp, Direction: interfaces.DESC_DIRECTION},
			}
		}
	case interfaces.QueryType_DSL:
		if useSearchAfter {
			defaultSort = []*interfaces.SortParamsV2{
				{Field: "_id", Direction: interfaces.DESC_DIRECTION},
			}
		} else {
			defaultSort = []*interfaces.SortParamsV2{}
		}
	default:
		defaultSort = []*interfaces.SortParamsV2{}
	}

	sort = append(sort, defaultSort...)
	newSort := []*interfaces.SortParamsV2{}
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

// // 将查询条件转成dsl，过滤条件、时间范围、分页、排序
// func toDSL(ctx context.Context, query interfaces.ViewQueryInterface, view *interfaces.DataView, indices []string) (bytes.Buffer, error) {
// 	queryParams := query.GetCommonParams()
// 	globalFilters := query.GetGlobalFilters()
// 	sortParams := query.GetSortParams()

// 	searchAfterDSL, err := getSearchAfterDSL(query.GetSearchAfterParams())
// 	if err != nil {
// 		return bytes.Buffer{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
// 			uerrors.Uniquery_DataView_InternalError_ConvertSearchAfterToDSLFailed).
// 			WithErrorDetails(fmt.Sprintf("failed to get search after dsl, %s", err.Error()))
// 	}

// 	var queryBuffer bytes.Buffer
// 	var queryStr string

// 	if len(sortParams) > 0 {
// 		sort := []map[string]any{}
// 		for _, sp := range sortParams {
// 			if sp.Field == "" || sp.Direction == "" {
// 				return bytes.Buffer{}, rest.NewHTTPError(ctx, http.StatusBadRequest,
// 					uerrors.Uniquery_DataView_InvalidParameter_Sort).
// 					WithErrorDetails("The sort field and direction cannot be empty")
// 			}

// 			sortField, ok := view.FieldsMap[sp.Field]
// 			if !ok {
// 				return bytes.Buffer{}, rest.NewHTTPError(ctx, http.StatusForbidden,
// 					uerrors.Uniquery_DataView_InvalidFieldPermission_Sort).
// 					WithErrorDetails(fmt.Sprintf("The sort field '%s' is not in the view fields list", sp.Field))
// 			}

// 			if sortField.Type == dtype.DATATYPE_BINARY {
// 				return bytes.Buffer{}, rest.NewHTTPError(ctx, http.StatusBadRequest,
// 					uerrors.Uniquery_DataView_BinaryFieldSortNotSupported).
// 					WithErrorDetails(fmt.Sprintf("The sort field '%s' is binary type, do not support sorting", sp.Field))
// 			}

// 			sortFieldName := sp.Field
// 			// 不需要将视图字段__id转为opensearch内置字段_id, 因为新的管道数据里已经存了  __id
// 			// if sortFieldName == "__id" {
// 			// 	sortFieldName = "_id"
// 			// }
// 			if sortField.Type == dtype.DATATYPE_TEXT {
// 				sortFieldName = sortFieldName + "." + dtype.KEYWORD_SUFFIX
// 			}

// 			sort = append(sort, map[string]any{
// 				sortFieldName: sp.Direction,
// 			})
// 		}

// 		sortStr, err := sonic.Marshal(sort)
// 		if err != nil {
// 			return bytes.Buffer{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
// 				uerrors.Uniquery_DataView_InternalError_MarshalFailed).
// 				WithErrorDetails(fmt.Errorf("data view marshal sort error: %s", err.Error()))
// 		}

// 		queryStr = fmt.Sprintf(`{%s
// 		"from": %d,
// 		"size": %d,
// 		"sort": %s,
// 		"query": {
// 			"bool": {
// 				"filter": [`, searchAfterDSL, queryParams.Offset, queryParams.Limit, string(sortStr))
// 	} else {
// 		queryStr = fmt.Sprintf(`{%s
// 			"from": %d,
// 			"size": %d,
// 			"query": {
// 				"bool": {
// 					"filter": [`, searchAfterDSL, queryParams.Offset, queryParams.Limit)
// 	}
// 	queryBuffer.WriteString(queryStr)

// 	// 视图数据查询接口视图的过滤条件和全局过滤条件合一起
// 	var dslStr string
// 	// if globalFilters != nil || view.Condition != nil {
// 	if globalFilters != nil {
// 		// var cfg *cond.CondCfg
// 		// if globalFilters == nil {
// 		// 	cfg = view.Condition
// 		// } else if view.Condition == nil {
// 		// 	cfg = globalFilters
// 		// } else {
// 		// 	cfg = &cond.CondCfg{
// 		// 		Operation: cond.OperationAnd,
// 		// 		SubConds: []*cond.CondCfg{
// 		// 			globalFilters,
// 		// 			view.Condition,
// 		// 		},
// 		// 	}
// 		// }

// 		cfg := globalFilters

// 		// 将过滤条件拼接到 dsl 的 query 中
// 		CondCfg, err := cond.NewCondition(cfg, view.Type, view.FieldsMap)
// 		if err != nil {
// 			return bytes.Buffer{}, rest.NewHTTPError(ctx, http.StatusBadRequest,
// 				uerrors.Uniquery_DataView_InvalidParameter_Filters).
// 				WithErrorDetails(fmt.Sprintf("failed to new condition, %s", err.Error()))
// 		}

// 		if CondCfg != nil {
// 			dslStr, err = CondCfg.Convert()
// 			if err != nil {
// 				return bytes.Buffer{}, rest.NewHTTPError(ctx, http.StatusBadRequest,
// 					uerrors.Uniquery_DataView_InvalidParameter_Filters).
// 					WithErrorDetails(fmt.Sprintf("failed to convert condition to dsl, %s", err.Error()))
// 			}
// 		}
// 	}

// 	mustFilters := []map[string]map[string]any{}
// 	// if view.LogGroupFilters != "" {
// 	// 	mustFilters = []map[string]map[string]any{
// 	// 		{
// 	// 			"query_string": {
// 	// 				"query":            view.LogGroupFilters,
// 	// 				"analyze_wildcard": true,
// 	// 			},
// 	// 		},
// 	// 	}
// 	// }

// 	mustStr, err := sonic.Marshal(mustFilters)
// 	if err != nil {
// 		return bytes.Buffer{}, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_DataView_InternalError_MarshalFailed).
// 			WithErrorDetails(fmt.Errorf("data view marshal must_filters error: %s", err.Error()))
// 	}

// 	if dslStr != "" {
// 		dslStr += ","
// 	}

// 	timeRangeStr := fmt.Sprintf(`%s
// 					{
// 						"range": {
// 							"@timestamp": {
// 								%s
// 							}
// 						}
// 					}
// 				],
// 				"must": %s
// 			}
// 		}
// 	}`, dslStr, func() string {
// 		switch {
// 		case queryParams.Start == 0 && queryParams.End == 0:
// 			return ""
// 		case queryParams.Start == 0:
// 			return fmt.Sprintf(`"lte": %d`, queryParams.End)
// 		case queryParams.End == 0:
// 			return fmt.Sprintf(`"gte": %d`, queryParams.Start)
// 		default:
// 			return fmt.Sprintf(`"gte": %d, "lte": %d`, queryParams.Start, queryParams.End)
// 		}
// 	}(), string(mustStr))

// 	queryBuffer.WriteString(timeRangeStr)

// 	logger.Debug(queryBuffer.String())
// 	// fmt.Println(queryBuffer.String())

// 	return queryBuffer, nil
// }
