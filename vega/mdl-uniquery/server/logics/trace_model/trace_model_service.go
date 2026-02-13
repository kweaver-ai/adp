// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package trace_model

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"sort"
	"sync"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/codes"
	"golang.org/x/sync/errgroup"

	"uniquery/common"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	"uniquery/logics"
	"uniquery/logics/permission"
	"uniquery/logics/trace_model/data_source"
)

var (
	tmServiceOnce sync.Once
	tmService     interfaces.TraceModelService
)

type traceModelService struct {
	appSetting *common.AppSetting
	tmAccess   interfaces.TraceModelAccess
	dcAccess   interfaces.DataConnectionAccess
	ps         interfaces.PermissionService
}

func NewTraceModelService(appSetting *common.AppSetting) interfaces.TraceModelService {
	tmServiceOnce.Do(func() {
		tmService = &traceModelService{
			appSetting: appSetting,
			tmAccess:   logics.TMAccess,
			dcAccess:   logics.DCAccess,
			ps:         permission.NewPermissionService(appSetting),
		}
	})
	return tmService
}

func (tms *traceModelService) GetSpanList(ctx context.Context, model interfaces.TraceModel, params interfaces.SpanListQueryParams) (entries []interfaces.SpanListEntry, total int64, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 获取span列表")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 决策当前模型id的数据查询权限
	err = tms.ps.CheckPermission(ctx, interfaces.Resource{
		ID:   model.ID,
		Type: interfaces.RESOURCE_TYPE_TRACE_MODEL,
	}, []string{interfaces.OPERATION_TYPE_DATA_QUERY})
	if err != nil {
		return nil, 0, err
	}

	sourceType, err := tms.getUnderlyingDataSouceType(ctx, interfaces.QUERY_CATEGORY_SPAN, model)
	if err != nil {
		return nil, 0, err
	}

	adapter, err := data_source.NewTraceModelAdapter(ctx, sourceType, tms.appSetting)
	if err != nil {
		o11y.Error(ctx, err.Error())
		return nil, 0, err
	}
	return adapter.GetSpanList(ctx, model, params)
}

func (tms *traceModelService) GetTrace(ctx context.Context, model interfaces.TraceModel, params interfaces.TraceQueryParams) (traceDetail interfaces.TraceDetail_, err error) {
	// start1 := time.Now()
	// fmt.Printf("[logic]开始查询Trace详情, 当前时间%v\n", start1)
	// defer func() {
	// 	end1 := time.Now()
	// 	fmt.Printf("[logic]结束查询Trace详情, 当前时间%v, 共耗时%v\n", end1, end1.Sub(start1))
	// }()

	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 获取trace详情")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 决策当前模型id的数据查询权限
	err = tms.ps.CheckPermission(ctx, interfaces.Resource{
		ID:   model.ID,
		Type: interfaces.RESOURCE_TYPE_TRACE_MODEL,
	}, []string{interfaces.OPERATION_TYPE_DATA_QUERY})
	if err != nil {
		return interfaces.TraceDetail_{}, err
	}

	var (
		briefSpanMap       = make(map[string]*interfaces.BriefSpan_)
		detailSpanMap      = make(map[string]interfaces.SpanDetail)
		relatedLogCountMap = make(map[string]int64)
	)

	g, ctx := errgroup.WithContext(ctx)
	// 1. 查询spanMap
	g.Go(
		func() error {
			spanSourceType, err := tms.getUnderlyingDataSouceType(ctx, interfaces.QUERY_CATEGORY_SPAN, model)
			if err != nil {
				return err
			}

			spanAdapter, err := data_source.NewTraceModelAdapter(ctx, spanSourceType, tms.appSetting)
			if err != nil {
				o11y.Error(ctx, err.Error())
				return err
			}

			briefSpanMap, detailSpanMap, err = spanAdapter.GetSpanMap(ctx, model, params)
			if err != nil {
				return err
			}
			return nil
		},
	)

	// 2. 查询relatedLogCountMap
	g.Go(
		func() error {
			if model.EnabledRelatedLog == interfaces.RELATED_LOG_OPEN {
				relatedLogSourceType, err := tms.getUnderlyingDataSouceType(ctx, interfaces.QUERY_CATEGORY_RELATED_LOG, model)
				if err != nil {
					return err
				}

				relatedLogAdapter, err := data_source.NewTraceModelAdapter(ctx, relatedLogSourceType, tms.appSetting)
				if err != nil {
					o11y.Error(ctx, err.Error())
					return err
				}

				relatedLogCountMap, err = relatedLogAdapter.GetRelatedLogCountMap(ctx, model, params)
				if err != nil {
					return err
				}
			}
			return nil
		},
	)

	// 3. 等待所有任务完成，如果有任何任务返回错误，只会捕获第一个错误
	if err := g.Wait(); err != nil {
		return interfaces.TraceDetail_{}, err
	}

	// 4. 判断trace数据是否存在
	if len(briefSpanMap) == 0 {
		errDetails := fmt.Sprintf("The trace whose traceID equals %s was not found!", params.TraceID)
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return interfaces.TraceDetail_{}, rest.NewHTTPError(ctx, http.StatusNotFound, uerrors.Uniquery_TraceModel_TraceNotFound).
			WithErrorDetails(errDetails)
	}

	// 5. 根据relatedLogMap, 添加每个span的关联日志条数
	for k, v := range relatedLogCountMap {
		if briefSpan, ok := briefSpanMap[k]; ok {
			briefSpan.RelatedLogCount = v
		}
	}

	// start3 := time.Now()
	// fmt.Printf("[logic]开始构建trace树结构, 当前时间%v\n", start3)

	// 6. 构建trace树结构, 返回rootSpan
	rootSpan := tms.buildTree(ctx, briefSpanMap)

	// end3 := time.Now()
	// fmt.Printf("[logic]结束构建trace树结构, 当前时间%v, 共耗时%v\n", end3, end3.Sub(start3))

	// start4 := time.Now()
	// fmt.Printf("[logic]开始获取trace统计数据, 当前时间%v\n", start4)

	// 7. 获取trace的统计信息
	sd := tms.getTraceStatisticData(ctx, rootSpan)

	// end4 := time.Now()
	// fmt.Printf("[logic]结束获取trace统计数据, 当前时间%v, 共耗时%v\n", end4, end4.Sub(start4))

	return interfaces.TraceDetail_{
		TraceID:     params.TraceID,
		StartTime:   sd.StartTime,
		EndTime:     sd.EndTime,
		Duration:    sd.EndTime - sd.StartTime,
		StatusStats: sd.StatusStats,
		Depth:       sd.Depth,
		Detail:      rootSpan,
		Spans:       detailSpanMap,
		Services:    sd.Services,
	}, nil
}

func (tms *traceModelService) GetSpan(ctx context.Context, model interfaces.TraceModel, params interfaces.SpanQueryParams) (spanDetail interfaces.SpanDetail, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 获取span详情")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 决策当前模型id的数据查询权限
	err = tms.ps.CheckPermission(ctx, interfaces.Resource{
		ID:   model.ID,
		Type: interfaces.RESOURCE_TYPE_TRACE_MODEL,
	}, []string{interfaces.OPERATION_TYPE_DATA_QUERY})
	if err != nil {
		return interfaces.SpanDetail{}, err
	}

	sourceType, err := tms.getUnderlyingDataSouceType(ctx, interfaces.QUERY_CATEGORY_SPAN, model)
	if err != nil {
		return interfaces.SpanDetail{}, err
	}

	adapter, err := data_source.NewTraceModelAdapter(ctx, sourceType, tms.appSetting)
	if err != nil {
		o11y.Error(ctx, err.Error())
		return interfaces.SpanDetail{}, err
	}
	return adapter.GetSpan(ctx, model, params)
}

func (tms *traceModelService) GetSpanRelatedLogList(ctx context.Context, model interfaces.TraceModel, params interfaces.RelatedLogListQueryParams) (relatedLogList []interfaces.RelatedLogListEntry, total int64, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 获取span关联日志列表")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 决策当前模型id的数据查询权限
	err = tms.ps.CheckPermission(ctx, interfaces.Resource{
		ID:   model.ID,
		Type: interfaces.RESOURCE_TYPE_TRACE_MODEL,
	}, []string{interfaces.OPERATION_TYPE_DATA_QUERY})
	if err != nil {
		return []interfaces.RelatedLogListEntry{}, 0, err
	}

	if model.EnabledRelatedLog == interfaces.RELATED_LOG_CLOSE {
		return []interfaces.RelatedLogListEntry{}, 0, nil
	}

	sourceType, err := tms.getUnderlyingDataSouceType(ctx, interfaces.QUERY_CATEGORY_RELATED_LOG, model)
	if err != nil {
		return []interfaces.RelatedLogListEntry{}, 0, err
	}

	adapter, err := data_source.NewTraceModelAdapter(ctx, sourceType, tms.appSetting)
	if err != nil {
		o11y.Error(ctx, err.Error())
		return []interfaces.RelatedLogListEntry{}, 0, err
	}
	return adapter.GetSpanRelatedLogList(ctx, model, params)
}

func (tms *traceModelService) GetTraceModelByID(ctx context.Context, modelID string) (model interfaces.TraceModel, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 根据ID查询链路模型对象")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	model, isExist, err := tms.tmAccess.GetTraceModelByID(ctx, modelID)
	if err != nil {
		logger.Errorf("Get trace model by id failed, err: %v", err.Error())
		return model, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_TraceModel_InternalError_GetTraceModelByIDFailed).
			WithErrorDetails(err.Error())
	}

	if !isExist {
		errDetails := fmt.Sprintf("The trace model whose id equal to %v was not found", modelID)
		logger.Errorf("Get trace model by id failed, err: %v", errDetails)
		return model, rest.NewHTTPError(ctx, http.StatusNotFound, uerrors.Uniquery_TraceModel_TraceModelNotFound).
			WithErrorDetails(errDetails)
	}

	return model, nil
}

func (tms *traceModelService) SimulateCreateTraceModel(ctx context.Context, model interfaces.TraceModel) (newModel interfaces.TraceModel, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 模拟创建链路模型")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 决策权限。 预览的时候还没有模型id，此时的预览校验用新建或者编辑，
	ops, err := tms.ps.GetResourcesOperations(ctx, interfaces.RESOURCE_TYPE_TRACE_MODEL,
		[]string{interfaces.RESOURCE_ID_ALL})
	if err != nil {
		return model, err
	}
	if len(ops) != 1 {
		// 无权限
		return model, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
			WithErrorDetails("Access denied: insufficient permissions for trace model's create or modify operation.")
	}
	// 从 ops 里找新建或编辑的权限
	for _, op := range ops[0].Operations {
		if op != interfaces.OPERATION_TYPE_CREATE && op != interfaces.OPERATION_TYPE_DELETE {
			return model, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
				WithErrorDetails("Access denied: insufficient permissions for trace model's create or modify operation.")
		}
	}

	simulateModel, err := tms.tmAccess.SimulateCreateTraceModel(ctx, model)
	if err != nil {
		logger.Errorf("Simulate create trace model failed, err: %v", err.Error())
		return model, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_TraceModel_InternalError_SimulateCreateTraceModelFailed).
			WithErrorDetails(err.Error())
	}

	return simulateModel, nil
}

func (tms *traceModelService) SimulateUpdateTraceModel(ctx context.Context, modelID string, model interfaces.TraceModel) (newModel interfaces.TraceModel, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 模拟修改链路模型")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 决策权限。 预览的时候还没有模型id，此时的预览校验用新建或者编辑，
	ops, err := tms.ps.GetResourcesOperations(ctx, interfaces.RESOURCE_TYPE_TRACE_MODEL,
		[]string{interfaces.RESOURCE_ID_ALL})
	if err != nil {
		return model, err
	}
	if len(ops) != 1 {
		// 无权限
		return model, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
			WithErrorDetails("Access denied: insufficient permissions for trace model's create or modify operation.")
	}
	// 从 ops 里找新建或编辑的权限
	for _, op := range ops[0].Operations {
		if op != interfaces.OPERATION_TYPE_CREATE && op != interfaces.OPERATION_TYPE_DELETE {
			return model, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
				WithErrorDetails("Access denied: insufficient permissions for trace model's create or modify operation.")
		}
	}

	simulateModel, err := tms.tmAccess.SimulateUpdateTraceModel(ctx, modelID, model)
	if err != nil {
		logger.Errorf("Simulate update trace model failed, err: %v", err.Error())
		return model, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_TraceModel_InternalError_SimulateUpdateTraceModelFailed).
			WithErrorDetails(err.Error())
	}

	return simulateModel, nil
}

/*
	私有方法
*/

// 获取底层数据源类型
func (tms *traceModelService) getUnderlyingDataSouceType(ctx context.Context, queryCategory string, model interfaces.TraceModel) (sourceType string, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("logic层: 获取%s配置的数据源类型", queryCategory))
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	if queryCategory == interfaces.QUERY_CATEGORY_SPAN && model.SpanSourceType == interfaces.SOURCE_TYPE_DATA_VIEW {
		return interfaces.SOURCE_TYPE_DATA_VIEW, nil
	} else if queryCategory == interfaces.QUERY_CATEGORY_RELATED_LOG && model.RelatedLogSourceType == interfaces.SOURCE_TYPE_DATA_VIEW {
		return interfaces.SOURCE_TYPE_DATA_VIEW, nil
	} else { // queryCategory == interfaces.QUERY_CATEGORY_SPAN && model.SpanSourceType == interfaces.SOURCE_TYPE_DATA_CONNECTION
		spanConfig := model.SpanConfig.(interfaces.SpanConfigWithDataConnection)
		underlyingSourceType, isExist, err := tms.dcAccess.GetDataConnectionTypeByName(ctx, spanConfig.DataConnection.Name)
		if err != nil {
			errDetails := fmt.Sprintf("Get underlying data souce type failed, err: %s", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			return underlyingSourceType, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_TraceModel_InternalError_GetUnderlyingDataSouceTypeFailed).
				WithErrorDetails(err.Error())
		}

		if !isExist {
			errDetails := fmt.Sprintf("Get underlying data souce type failed, the data connection whose name equal to %s was not found", spanConfig.DataConnection.Name)
			logger.Errorf(errDetails)
			o11y.Error(ctx, errDetails)
			return underlyingSourceType, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_TraceModel_InternalError_GetUnderlyingDataSouceTypeFailed).
				WithErrorDetails(errDetails)
		}
		return underlyingSourceType, nil
	}
}

// 构建Trace树结构
func (tms *traceModelService) buildTree(ctx context.Context, spanMap map[string]*interfaces.BriefSpan_) *interfaces.BriefSpan_ {
	_, span := ar_trace.Tracer.Start(ctx, "logic层: 构建trace树结构")
	defer func() {
		span.SetStatus(codes.Ok, "")
		span.End()
	}()

	var rootSpanID string

	for spanID, span := range spanMap {
		parentSpanID := span.ParentSpanID
		if parentSpan, ok := spanMap[parentSpanID]; ok {
			subSpans := parentSpan.Children
			// 使用sort.Search查找插入位置
			idx := sort.Search(len(subSpans), func(i int) bool {
				return subSpans[i].StartTime >= span.StartTime
			})

			// 插入span
			parentSpan.Children = append(subSpans[:idx], append([]*interfaces.BriefSpan_{span}, subSpans[idx:]...)...)
		} else {
			rootSpanID = spanID
		}
	}

	return spanMap[rootSpanID]
}

// 获取Trace统计数据
func (tms *traceModelService) getTraceStatisticData(ctx context.Context, rootSpan *interfaces.BriefSpan_) interfaces.TraceStatisticData {
	_, span := ar_trace.Tracer.Start(ctx, "logic层: 获取trace统计数据")
	defer func() {
		span.SetStatus(codes.Ok, "")
		span.End()
	}()

	sd := interfaces.TraceStatisticData{
		StartTime: math.MaxInt64,
		EndTime:   math.MinInt64,
		StatusStats: map[string]int64{
			interfaces.SPAN_STATUS_UNSET: 0,
			interfaces.SPAN_STATUS_OK:    0,
			interfaces.SPAN_STATUS_ERROR: 0,
		},
	}

	if rootSpan == nil {
		return sd
	}

	serviceMap := make(map[string]struct{})
	queue := []*interfaces.BriefSpan_{rootSpan}
	for len(queue) > 0 {
		levelSize := len(queue)
		for i := 0; i < levelSize; i++ {
			span := queue[0]
			sd.StatusStats[span.Status]++
			sd.StartTime = min(sd.StartTime, span.StartTime)
			sd.EndTime = max(sd.EndTime, span.EndTime)
			serviceMap[span.ServiceName] = struct{}{}

			queue = append(queue[1:], span.Children...)
		}
		sd.Depth++
	}

	for serviceName := range serviceMap {
		sd.Services = append(sd.Services, serviceName)
	}

	return sd
}
