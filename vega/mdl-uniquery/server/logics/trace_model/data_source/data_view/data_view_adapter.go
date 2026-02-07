// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_source

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic/ast"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	libCommon "github.com/kweaver-ai/kweaver-go-lib/common"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/codes"
	"golang.org/x/sync/errgroup"

	"uniquery/common"
	cond "uniquery/common/condition"
	vopt "uniquery/common/value_opt"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	"uniquery/logics/data_view"
	"uniquery/logics/dsl"
)

var (
	dvaOnce sync.Once
	dva     interfaces.TraceModelAdapter
)

type dataViewAdapter struct {
	dvService  interfaces.DataViewService
	dslService interfaces.DslService
}

func NewDataViewAdapter(appSetting *common.AppSetting) interfaces.TraceModelAdapter {
	dvaOnce.Do(func() {
		dva = &dataViewAdapter{
			dvService:  data_view.NewDataViewService(appSetting),
			dslService: dsl.NewDslService(appSetting),
		}
	})
	return dva
}

func (dvAdapter *dataViewAdapter) GetSpanList(ctx context.Context, model interfaces.TraceModel, params interfaces.SpanListQueryParams) (entries []interfaces.SpanListEntry, total int64, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 通过dataViewAdapter获取span列表")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	spanConf, _ := model.SpanConfig.(interfaces.SpanConfigWithDataView)

	// 1. 生成query, 供dataView查询使用
	if params.TraceID != "_all" {
		params.Condition = &cond.CondCfg{
			Operation: cond.OperationAnd,
			SubConds: []*cond.CondCfg{
				{
					Operation: cond.OperationEq,
					Name:      spanConf.TraceID.FieldName,
					ValueOptCfg: vopt.ValueOptCfg{
						ValueFrom: vopt.ValueFrom_Const,
						Value:     params.TraceID,
					},
				},
				params.Condition,
			},
		}
	}

	query := &interfaces.DataViewQueryV1{
		GlobalFilters: params.Condition,
		SortParamsV1: interfaces.SortParamsV1{
			Sort:      params.Sort,
			Direction: params.Direction,
		},
		ViewQueryCommonParams: interfaces.ViewQueryCommonParams{
			Offset:    params.Offset,
			Limit:     params.Limit,
			NeedTotal: true,
			Format:    interfaces.Format_Original,
		},
	}

	// 2. 基于数据视图查询Span数据
	viewInternalResp, err := dvAdapter.dvService.RetrieveSingleViewData(ctx, spanConf.DataView.ID, query)
	if err != nil {
		return []interfaces.SpanListEntry{}, 0, err
	}

	if len(viewInternalResp.Datas) == 0 {
		return []interfaces.SpanListEntry{}, 0, nil
	}

	// 3. 将viewUniRes[0].Datas[0].Values转换为spanList
	total = viewInternalResp.Total
	astNodes := viewInternalResp.Datas
	spanList := make([]interfaces.SpanListEntry, len(astNodes))

	g, _ := errgroup.WithContext(ctx)
	g.SetLimit(40)
	for i, astNode := range astNodes {
		i, astNode := i, astNode
		g.Go(
			func() error {
				rawSpan, _ := astNode.MapUseNumber()
				abstractSpan := dvAdapter.extractRawSpan(spanConf, astNode, false)
				spanList[i] = dvAdapter.genSpanDetail(rawSpan, abstractSpan)
				return nil
			},
		)
	}

	_ = g.Wait()
	return spanList, total, nil
}

func (dvAdapter *dataViewAdapter) GetSpan(ctx context.Context, model interfaces.TraceModel, params interfaces.SpanQueryParams) (spanDetail interfaces.SpanDetail, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 通过dataViewAdapter获取span详情")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	spanConf, _ := model.SpanConfig.(interfaces.SpanConfigWithDataView)

	// 1. 根据链路模型对象, 以及传入的traceID与spanID
	// 解析其对应的原字段, 并构建filters
	subConditions := make([]*cond.CondCfg, 0)
	if params.TraceID != "_all" {
		subConditions = append(subConditions, &cond.CondCfg{
			Operation: cond.OperationEq,
			Name:      spanConf.TraceID.FieldName,
			ValueOptCfg: vopt.ValueOptCfg{
				ValueFrom: vopt.ValueFrom_Const,
				Value:     params.TraceID,
			},
		})
	}
	vals := strings.Split(params.SpanID, interfaces.DEFAULT_SEPARATOR)
	fieldNames := spanConf.SpanID.FieldNames
	// 如果分割出的字段值数量与字段数量不一致, 说明分隔符选的不好, 应当更换
	if len(vals) != len(fieldNames) {
		errDetails := fmt.Sprintf("The spanID fails to be split by the default separator %v, please pass spanID in the correct format or check whether spanID is configured correctly in the trace model", interfaces.DEFAULT_SEPARATOR)
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return interfaces.SpanDetail{}, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_TraceModel_InternalError_SplitSpanIDFailed).
			WithErrorDetails(errDetails)
	}

	for index, fieldName := range fieldNames {
		subConditions = append(subConditions, &cond.CondCfg{
			Operation: cond.OperationEq,
			Name:      fieldName,
			ValueOptCfg: vopt.ValueOptCfg{
				ValueFrom: vopt.ValueFrom_Const,
				Value:     vals[index],
			},
		})
	}

	// 2. 构建数据视图查询所需的queries
	query := &interfaces.DataViewQueryV1{
		GlobalFilters: &cond.CondCfg{
			Operation: cond.OperationAnd,
			SubConds:  subConditions,
		},
		ViewQueryCommonParams: interfaces.ViewQueryCommonParams{
			Limit:  1,
			Format: interfaces.Format_Original,
		},
	}

	// 3. 基于数据视图查询Span数据
	viewInternalResp, err := dvAdapter.dvService.RetrieveSingleViewData(ctx, spanConf.DataView.ID, query)
	if err != nil {
		return interfaces.SpanDetail{}, err
	}

	if len(viewInternalResp.Datas) == 0 {
		errDetails := fmt.Sprintf("The span whose spanID equals %s was not found!", params.SpanID)
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return interfaces.SpanDetail{}, rest.NewHTTPError(ctx, http.StatusNotFound,
			uerrors.Uniquery_TraceModel_SpanNotFound).WithErrorDetails(errDetails)
	}

	// 4. 将viewUniRes[0].Datas[0].Values转换为spanDetail
	astNode := viewInternalResp.Datas[0]
	rawSpan, _ := astNode.MapUseNumber()
	abstractSpan := dvAdapter.extractRawSpan(spanConf, astNode, false)
	spanDetail = dvAdapter.genSpanDetail(rawSpan, abstractSpan)
	return spanDetail, nil
}

func (dvAdapter *dataViewAdapter) GetSpanRelatedLogList(ctx context.Context, model interfaces.TraceModel, params interfaces.RelatedLogListQueryParams) (relatedLogList []interfaces.RelatedLogListEntry, total int64, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 通过dataViewAdapter获取span关联日志列表")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	relatedLogConf, _ := model.RelatedLogConfig.(interfaces.RelatedLogConfigWithDataView)
	relatedLogList = make([]interfaces.RelatedLogListEntry, 0)

	// 1. 根据链路模型对象, 以及传入的traceID与spanID
	// 解析其对应的原字段, 并构建filters
	subConditions := []*cond.CondCfg{params.Condition}
	if params.TraceID != "_all" {
		subConditions = append(subConditions, &cond.CondCfg{
			Operation: cond.OperationEq,
			Name:      relatedLogConf.TraceID.FieldName,
			ValueOptCfg: vopt.ValueOptCfg{
				ValueFrom: vopt.ValueFrom_Const,
				Value:     params.TraceID,
			},
		})
	}

	if params.SpanID != "_all" {
		vals := strings.Split(params.SpanID, interfaces.DEFAULT_SEPARATOR)
		fieldNames := relatedLogConf.SpanID.FieldNames
		// 如果分割出的字段值数量与字段数量不一致, 说明分隔符选的不好, 应当更换
		if len(vals) != len(fieldNames) {
			errDetails := fmt.Sprintf("The spanID fails to be split by the default separator %v, please pass spanID in the correct format or check whether spanID is configured correctly in the trace model", interfaces.DEFAULT_SEPARATOR)
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			return relatedLogList, total, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_TraceModel_InternalError_SplitSpanIDFailed).
				WithErrorDetails(errDetails)
		}

		for index, fieldName := range fieldNames {
			subConditions = append(subConditions, &cond.CondCfg{
				Operation: cond.OperationEq,
				Name:      fieldName,
				ValueOptCfg: vopt.ValueOptCfg{
					ValueFrom: vopt.ValueFrom_Const,
					Value:     vals[index],
				},
			})
		}
	}

	// 2. 构建数据视图查询所需的queries
	query := &interfaces.DataViewQueryV1{
		GlobalFilters: &cond.CondCfg{
			Operation: cond.OperationAnd,
			SubConds:  subConditions,
		},
		SortParamsV1: interfaces.SortParamsV1{
			Sort:      params.Sort,
			Direction: params.Direction,
		},
		ViewQueryCommonParams: interfaces.ViewQueryCommonParams{
			Offset:    params.Offset,
			Limit:     params.Limit,
			NeedTotal: true,
			Format:    interfaces.Format_Original,
		},
	}

	// 3. 基于数据视图查询Span关联日志数据
	viewInternalResp, err := dvAdapter.dvService.RetrieveSingleViewData(ctx, relatedLogConf.DataView.ID, query)
	if err != nil {
		return relatedLogList, total, err
	}

	if len(viewInternalResp.Datas) == 0 {
		return relatedLogList, total, nil
	}

	// 4. 将viewUniRes.Datas[0].Values转换为relatedLogList
	astNodes := viewInternalResp.Datas
	relatedLogList = make([]interfaces.RelatedLogListEntry, len(astNodes))
	total = viewInternalResp.Total

	g, _ := errgroup.WithContext(ctx)
	g.SetLimit(40)
	for i, astNode := range astNodes {
		i, astNode := i, astNode
		g.Go(
			func() error {
				rawRelatedLog, _ := astNode.MapUseNumber()
				abstractRelatedLog := dvAdapter.extractRawRelatedLog(relatedLogConf, astNode)
				relatedLogEntry := dvAdapter.genRelatedLogDetail(rawRelatedLog, abstractRelatedLog)
				relatedLogList[i] = relatedLogEntry

				return nil
			},
		)
	}

	_ = g.Wait()

	return relatedLogList, total, nil
}

func (dvAdapter *dataViewAdapter) GetSpanMap(ctx context.Context, model interfaces.TraceModel, params interfaces.TraceQueryParams) (briefSpanMap map[string]*interfaces.BriefSpan_, detailSpanMap map[string]interfaces.SpanDetail, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 通过dataViewAdapter获取span map")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	spanConf, _ := model.SpanConfig.(interfaces.SpanConfigWithDataView)

	// 1. 生成queries及viewIDStr, 供dataView查询使用
	query := &interfaces.DataViewQueryV1{
		GlobalFilters: &cond.CondCfg{
			Operation: cond.OperationEq,
			Name:      spanConf.TraceID.FieldName,
			ValueOptCfg: vopt.ValueOptCfg{
				ValueFrom: vopt.ValueFrom_Const,
				Value:     params.TraceID,
			},
		},
		Scroll: interfaces.DEFAULT_SEARCH_SCROLL_STR,
		ViewQueryCommonParams: interfaces.ViewQueryCommonParams{
			Limit:  interfaces.MAX_SEARCH_SIZE,
			Format: interfaces.Format_Original,
		},
	}

	scrollIDs := make([]string, 0)
	defer func() {
		// 删除本次分页查询使用的scrollID
		go dvAdapter.clearScrollIDs(context.Background(), scrollIDs)
	}()

	wg := &sync.WaitGroup{}
	l := &sync.Mutex{}
	spanMap := make(map[string]*interfaces.BriefSpan_, 0)

	// 2. 基于数据视图查询Span数据
	for {
		// start1 := time.Now()
		// fmt.Printf("[logic]开始调用数据视图模块获取span列表, 当前时间%v\n", start1)

		// 2.1 调用dataview的service, 获取rawSpan
		viewInternalResp, err := dvAdapter.dvService.RetrieveSingleViewData(ctx, spanConf.DataView.ID, query)
		if err != nil {
			return spanMap, make(map[string]interfaces.SpanDetail), err
		}

		// end1 := time.Now()
		// fmt.Printf("[logic]结束调用数据视图模块获取span列表, 当前时间%v, 共耗时%v\n", end1, end1.Sub(start1))

		// 2.2 提取scrollID, 记录在scrollIDs中, 最后一并删除掉
		scrollID := viewInternalResp.ScrollId
		scrollIDs = append(scrollIDs, scrollID)

		// 2.3 从astNodes中提取出AbstractSpan, 并存储到map中
		if len(viewInternalResp.Datas) == 0 {
			break
		}
		astNodes := viewInternalResp.Datas

		wg.Add(1)
		go func() {
			defer wg.Done()
			dvAdapter.modifySpanMap(ctx, l, astNodes, spanConf, spanMap)
		}()

		// 2.4 判断是否达到循环终止条件
		if len(astNodes) < interfaces.MAX_SEARCH_SIZE {
			break
		}

		// 2.5 更新queries
		query.ScrollId = scrollID
	}

	wg.Wait()
	return spanMap, make(map[string]interfaces.SpanDetail), nil
}

func (dvAdapter *dataViewAdapter) GetRelatedLogCountMap(ctx context.Context, model interfaces.TraceModel, params interfaces.TraceQueryParams) (countMap map[string]int64, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 通过dataViewAdapter获取关联日志的统计信息")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	relatedLogConf, _ := model.RelatedLogConfig.(interfaces.RelatedLogConfigWithDataView)

	// 1. 生成queries, 供dataView查询使用
	query := &interfaces.DataViewQueryV1{
		GlobalFilters: &cond.CondCfg{
			Operation: cond.OperationEq,
			Name:      relatedLogConf.TraceID.FieldName,
			ValueOptCfg: vopt.ValueOptCfg{
				ValueFrom: vopt.ValueFrom_Const,
				Value:     params.TraceID,
			},
		},
		ViewQueryCommonParams: interfaces.ViewQueryCommonParams{
			Format: interfaces.Format_Original,
		},
	}

	// 2. 遍历traceModel, 组装fields
	fields := append([]string{}, relatedLogConf.SpanID.FieldNames...)

	// 3. 基于数据视图查询Span数据
	stats, err := dvAdapter.dvService.CountMultiFields(ctx, relatedLogConf.DataView.ID, query, fields, interfaces.DEFAULT_SEPARATOR)
	if err != nil {
		return stats, err
	}

	return stats, nil
}

/*
	私有方法
*/

func (dvAdapter *dataViewAdapter) extractRawSpan(spanConfig interfaces.SpanConfigWithDataView, astNode *ast.Node, simpleInfo bool) interfaces.AbstractSpan {
	abstractSpan := interfaces.AbstractSpan{}

	// 1. 提取Name
	nameConf := spanConfig.Name
	paths := common.SplitString2InterfaceArray(nameConf.FieldName, ".")
	if spanName, err := astNode.GetByPath(paths...).String(); err == nil {
		abstractSpan.Name = spanName
	}

	// 2. 提取SpanID
	spanIDConf := spanConfig.SpanID
	vals := make([]string, 0, len(spanIDConf.FieldNames))
	for _, fieldName := range spanIDConf.FieldNames {
		paths := common.SplitString2InterfaceArray(fieldName, ".")
		if val, err := astNode.GetByPath(paths...).InterfaceUseNumber(); err == nil {
			vals = append(vals, common.Any2String(val))
		}
	}
	abstractSpan.SpanID = strings.Join(vals, interfaces.DEFAULT_SEPARATOR)

	// 3. 提取ParentSpanID
	parentSpanIDConfs := spanConfig.ParentSpanID
	abstractSpan.ParentSpanID = dvAdapter.parseParentSpanID(parentSpanIDConfs, astNode)

	// 4. 提取StartTime, EndTime和Duration
	abstractSpan.StartTime, abstractSpan.EndTime, abstractSpan.Duration = dvAdapter.parseTimeAndDuration(spanConfig, astNode)

	// 5. 提取Kind
	kindConf := spanConfig.Kind
	paths = common.SplitString2InterfaceArray(kindConf.FieldName, ".")
	if spanKind, err := astNode.GetByPath(paths...).String(); err == nil {
		abstractSpan.Kind = spanKind
	}

	if val, ok := interfaces.SPAN_KIND_MAP[strings.ToLower(abstractSpan.Kind)]; ok {
		abstractSpan.Kind = val
	} else {
		abstractSpan.Kind = interfaces.SPAN_KIND_UNSPECIFIED
	}

	// 6. 提取Status
	statusConf := spanConfig.Status
	paths = common.SplitString2InterfaceArray(statusConf.FieldName, ".")
	if spanStatus, err := astNode.GetByPath(paths...).String(); err == nil {
		abstractSpan.Status = spanStatus
	}

	if val, ok := interfaces.SPAN_STATUS_MAP[strings.ToLower(abstractSpan.Status)]; ok {
		abstractSpan.Status = val
	} else {
		abstractSpan.Status = interfaces.SPAN_STATUS_UNSET
	}

	// 7. 提取ServiceName
	serviceNameConf := spanConfig.ServiceName
	paths = common.SplitString2InterfaceArray(serviceNameConf.FieldName, ".")
	if serviceName, err := astNode.GetByPath(paths...).String(); err == nil {
		abstractSpan.ServiceName = serviceName
	}

	if !simpleInfo {
		// 8. 提取TraceID
		traceIDConf := spanConfig.TraceID
		paths = common.SplitString2InterfaceArray(traceIDConf.FieldName, ".")
		if traceID, err := astNode.GetByPath(paths...).String(); err == nil {
			abstractSpan.TraceID = traceID
		}
	}

	return abstractSpan
}

func (dvAdapter *dataViewAdapter) parseParentSpanID(confs []interfaces.ParentSpanIDConfig, astNode *ast.Node) string {
	for _, conf := range confs {
		// 判断满足前置条件
		if i, flag := dvAdapter.isSatisfyCondition(conf.Precond, astNode); flag {
			fieldNames := conf.FieldNames
			vals := make([]string, 0, len(fieldNames))

			for _, fieldName := range fieldNames {
				valInfs := &[]any{}
				dvAdapter.getValueFromAstNode(astNode, strings.Split(fieldName, "."), valInfs)
				vals = append(vals, common.Any2String((*valInfs)[i]))
			}
			return strings.Join(vals, interfaces.DEFAULT_SEPARATOR)
		}
	}

	return ""
}

func (dvAdapter *dataViewAdapter) isSatisfyCondition(precond *cond.CondCfg, astNode *ast.Node) (int, bool) {
	if precond == nil {
		return 0, true
	}

	switch precond.Operation {
	case cond.OperationAnd:
		index := 0
		for _, subCond := range precond.SubConds {
			i, flag := dvAdapter.isSatisfyCondition(subCond, astNode)
			if !flag {
				return i, false
			}

			if i > 0 {
				index = i
			}
		}
		return index, true
	case cond.OperationOr:
		index := 0
		subConds := precond.SubConds
		if len(subConds) == 0 {
			return index, true
		}

		for _, subCond := range subConds {
			i, flag := dvAdapter.isSatisfyCondition(subCond, astNode)
			if flag {
				return i, true
			}

			if i > 0 {
				index = i
			}
		}
		return index, false
	default:
		val1s, val2s := &[]any{}, &[]any{}
		dvAdapter.getValueFromAstNode(astNode, strings.Split(precond.Name, "."), val1s)

		// value_from只能有const和field两种类型
		if precond.ValueFrom == vopt.ValueFrom_Field {
			dvAdapter.getValueFromAstNode(astNode, strings.Split(common.Any2String(precond.Value), "."), val2s)
		} else {
			*val2s = append(*val2s, precond.Value)
		}

		if precond.Operation == cond.OperationEq {
			if len(*val1s) == 1 {
				for i := range *val2s {
					if (*val2s)[i] == (*val1s)[0] {
						return i, true
					}
				}
			} else if len(*val2s) == 1 {
				for i := range *val1s {
					if (*val1s)[i] == (*val2s)[0] {
						return i, true
					}
				}
			}
		} else if precond.Operation == cond.OperationNotEq {
			if len(*val1s) == 1 {
				for i := range *val2s {
					if (*val2s)[i] != (*val1s)[0] {
						return i, true
					}
				}
			} else if len(*val2s) == 1 {
				for i := range *val1s {
					if (*val1s)[i] != (*val2s)[0] {
						return i, true
					}
				}
			}
		} else if precond.Operation == cond.OperationRange {
			tStr, _ := (*val1s)[0].(string)
			rangeTimeStrs, _ := (*val2s)[0].([]any)
			ltStr, _ := rangeTimeStrs[0].(string)
			rtStr, _ := rangeTimeStrs[1].(string)

			t, _ := time.Parse(libCommon.RFC3339Milli, tStr)
			lt, _ := time.Parse(libCommon.RFC3339Milli, ltStr)
			rt, _ := time.Parse(libCommon.RFC3339Milli, rtStr)

			if !t.Before(lt) && !t.After(rt) {
				return 0, true
			}
		}
		return 0, false
	}
}

func (dvAdapter *dataViewAdapter) getValueFromAstNode(astNode *ast.Node, paths []string, vals *[]any) {
	if astNode == nil {
		*vals = append(*vals, nil)
		return
	}

	if astNode.TypeSafe() == ast.V_ARRAY {
		astNodeArrs, _ := astNode.ArrayUseNode()
		for i := range astNodeArrs {
			dvAdapter.getValueFromAstNode(astNode.Get(common.Any2String(i)), paths, vals)
		}
	} else if len(paths) == 0 {
		inf, _ := astNode.InterfaceUseNumber()
		*vals = append(*vals, inf)
	} else {
		dvAdapter.getValueFromAstNode(astNode.Get(paths[0]), paths[1:], vals)
	}
}

func (dvAdapter *dataViewAdapter) parseTimeAndDuration(spanConf interfaces.SpanConfigWithDataView, astNode *ast.Node) (int64, int64, int64) {
	// 1. 提取StartTime
	startTimeConf := spanConf.StartTime
	paths := common.SplitString2InterfaceArray(startTimeConf.FieldName, ".")
	valInf, _ := astNode.GetByPath(paths...).Number()
	startTime, err := valInf.Int64()
	if err != nil {
		logger.Errorf("Parse span start time err: %v", err)
	}

	// 根据format调整值
	switch startTimeConf.FieldFormat {
	case interfaces.UNIX_MILLIS:
		startTime *= 1000
	case interfaces.UNIX_NANOS:
		startTime /= 1000
	}

	// 2. 提取EndTime和Duration
	endTimeConf := spanConf.EndTime
	durationConf := spanConf.Duration

	// case1. endTimeConf为空
	if endTimeConf.FieldName == "" {
		// 提取Duration
		// duration, err = rawSpan[durationConf.FieldName].(json.Number).Int64()
		path := common.SplitString2InterfaceArray(durationConf.FieldName, ".")
		valInf, _ := astNode.GetByPath(path...).Number()
		duration, err := valInf.Int64()
		if err != nil {
			logger.Errorf("Parse span duration err: %v", err)
		}

		// 根据unit调整值
		switch durationConf.FieldUnit {
		case interfaces.MS:
			duration *= 1000
		case interfaces.NS:
			duration /= 1000
		}

		// 根据StartTime和Duration计算EndTime
		endTime := startTime + duration
		return startTime, endTime, duration
	} else {
		// case2. endTimeConf不为空
		// 提取EndTime
		paths = common.SplitString2InterfaceArray(endTimeConf.FieldName, ".")
		valInf, _ = astNode.GetByPath(paths...).Number()
		endTime, err := valInf.Int64()
		if err != nil {
			logger.Errorf("Parse span end time err: %v", err)
		}

		// 根据format调整值
		switch endTimeConf.FieldFormat {
		case interfaces.UNIX_MILLIS:
			endTime *= 1000
		case interfaces.UNIX_NANOS:
			endTime /= 1000
		}

		// 根据EndTime和StartTime计算Duration
		duration := endTime - startTime
		return startTime, endTime, duration
	}
}

func (dvAdapter *dataViewAdapter) modifySpanMap(ctx context.Context, l *sync.Mutex, astNodes []*ast.Node, spanConf interfaces.SpanConfigWithDataView, spanMap map[string]*interfaces.BriefSpan_) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 解析data_view模块返回的数据, 得到span map")
	defer func() {
		span.SetStatus(codes.Ok, "")
		span.End()
	}()

	// start1 := time.Now()
	// fmt.Printf("[logic]开始更新TraceStats, 当前时间%v\n", start1)
	// defer func() {
	// 	end1 := time.Now()
	// 	fmt.Printf("[logic]结束更新TraceStats, 当前时间%v, 共耗时%v\n", end1, end1.Sub(start1))
	// }()

	g, _ := errgroup.WithContext(ctx)
	g.SetLimit(40)

	for _, astNode := range astNodes {
		tmpNode := astNode
		g.Go(
			func() error {
				abstractSpan := dvAdapter.extractRawSpan(spanConf, tmpNode, true)

				l.Lock()
				defer l.Unlock()
				spanMap[abstractSpan.SpanID] = &interfaces.BriefSpan_{
					Key:          abstractSpan.SpanID,
					Name:         abstractSpan.Name,
					SpanID:       abstractSpan.SpanID,
					ParentSpanID: abstractSpan.ParentSpanID,
					StartTime:    abstractSpan.StartTime,
					EndTime:      abstractSpan.EndTime,
					Duration:     abstractSpan.Duration,
					Kind:         abstractSpan.Kind,
					Status:       abstractSpan.Status,
					ServiceName:  abstractSpan.ServiceName,
					Children:     make([]*interfaces.BriefSpan_, 0),
				}
				return nil
			},
		)
	}

	_ = g.Wait()
}

// 批量清理scrollID
func (dvAdapter *dataViewAdapter) clearScrollIDs(ctx context.Context, scrollIDs []string) {
	if len(scrollIDs) == 0 {
		return
	}

	para := interfaces.DeleteScroll{
		ScrollId: scrollIDs,
	}

	// 返回值不接, golangci-lint报错
	_, _, _ = dvAdapter.dslService.DeleteScroll(ctx, para)
}

func (dvAdapter *dataViewAdapter) genSpanDetail(rawSpan map[string]any, abstractSpan interfaces.AbstractSpan) map[string]any {
	if rawSpan == nil {
		rawSpan = make(map[string]any)
	}

	rawSpan["__trace_id"] = abstractSpan.TraceID
	rawSpan["__span_id"] = abstractSpan.SpanID
	rawSpan["__parent_span_id"] = abstractSpan.ParentSpanID
	rawSpan["__name"] = abstractSpan.Name
	rawSpan["__start_time"] = abstractSpan.StartTime
	rawSpan["__end_time"] = abstractSpan.EndTime
	rawSpan["__duration"] = abstractSpan.Duration
	rawSpan["__kind"] = abstractSpan.Kind
	rawSpan["__status"] = abstractSpan.Status
	rawSpan["__service_name"] = abstractSpan.ServiceName

	return rawSpan
}

func (dvAdapter *dataViewAdapter) genRelatedLogDetail(rawRelatedLog map[string]any, abstractRelatedLog interfaces.AbstractRelatedLog) map[string]any {
	if rawRelatedLog == nil {
		rawRelatedLog = make(map[string]any)
	}

	rawRelatedLog["__trace_id"] = abstractRelatedLog.TraceID
	rawRelatedLog["__span_id"] = abstractRelatedLog.SpanID

	return rawRelatedLog
}

func (dvAdapter *dataViewAdapter) extractRawRelatedLog(relatedLogConf interfaces.RelatedLogConfigWithDataView, astNode *ast.Node) interfaces.AbstractRelatedLog {
	abstractRelatedLog := interfaces.AbstractRelatedLog{}

	// 1. 生成__trace_id
	traceIDFieldName := relatedLogConf.TraceID.FieldName
	paths := common.SplitString2InterfaceArray(traceIDFieldName, ".")

	if val, err := astNode.GetByPath(paths...).InterfaceUseNumber(); err == nil {
		abstractRelatedLog.TraceID = common.Any2String(val)
	}

	// 2. 生成__span_id
	spanIDFieldNames := relatedLogConf.SpanID.FieldNames
	spanIDVals := make([]string, 0)
	for _, spanIDFieldName := range spanIDFieldNames {
		paths := common.SplitString2InterfaceArray(spanIDFieldName, ".")
		if val, err := astNode.GetByPath(paths...).InterfaceUseNumber(); err == nil {
			spanIDVals = append(spanIDVals, common.Any2String(val))
		}
	}
	abstractRelatedLog.SpanID = strings.Join(spanIDVals, interfaces.DEFAULT_SEPARATOR)

	return abstractRelatedLog
}
