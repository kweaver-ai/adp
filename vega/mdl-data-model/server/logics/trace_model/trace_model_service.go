// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package trace_model

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/rs/xid"
	"go.opentelemetry.io/otel/codes"

	"data-model/common"
	derrors "data-model/errors"
	"data-model/interfaces"
	dcond "data-model/interfaces/condition"
	dtype "data-model/interfaces/data_type"
	"data-model/logics"
	"data-model/logics/data_connection"
	"data-model/logics/data_view"
	"data-model/logics/permission"
	"data-model/logics/trace_model/data_source"
)

var (
	tmServiceOnce sync.Once
	tmService     interfaces.TraceModelService
)

type traceModelService struct {
	appSetting *common.AppSetting
	dvs        interfaces.DataViewService
	ps         interfaces.PermissionService
	dcs        interfaces.DataConnectionService
	tma        interfaces.TraceModelAccess
}

func NewTraceModelService(appSetting *common.AppSetting) interfaces.TraceModelService {
	tmServiceOnce.Do(func() {
		tmService = &traceModelService{
			appSetting: appSetting,
			dvs:        data_view.NewDataViewService(appSetting),
			ps:         permission.NewPermissionService(appSetting),
			dcs:        data_connection.NewDataConnectionService(appSetting),
			tma:        logics.TMA,
		}
	})
	return tmService
}

// 批量创建/导入链路模型
func (tms *traceModelService) CreateTraceModels(ctx context.Context, reqModels []interfaces.TraceModel) (modelIDs []string, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 批量创建链路模型")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 判断userid是否有创建链路模型的权限
	err = tms.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_TRACE_MODEL,
		ID:   interfaces.RESOURCE_ID_ALL,
	}, []string{interfaces.OPERATION_TYPE_CREATE})
	if err != nil {
		return nil, err
	}

	// 1. 根据数据视图 ID 获取对应的viewMap
	viewIDs := tms.getDependentViewIDs(reqModels)
	viewMap, err := tms.getDetailedViewMapByIDs(ctx, viewIDs)
	if err != nil {
		return nil, err
	}

	// 2. 根据数据连接名称获取对应的connMap
	connNames := tms.getDependentConnectionNames(reqModels)
	connMap, err := tms.getConnectionMapByNames(ctx, connNames)
	if err != nil {
		return nil, err
	}

	// 3. 校验reqModels所依赖的数据视图字段
	err = tms.validateReqTraceModels(ctx, reqModels, viewMap)
	if err != nil {
		return nil, err
	}

	// 4. 修改reqModels, 如更新依赖的数据连接ID, 生成update_time等
	err = tms.modifyReqModels(ctx, false, connMap, reqModels)
	if err != nil {
		return nil, err
	}

	// 5. 调用driven层, 批量创建链路模型
	err = tms.tma.CreateTraceModels(ctx, reqModels)
	if err != nil {
		logger.Errorf("Create trace models failed, err: %v", err.Error())
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_TraceModel_InternalError_CreateTraceModelsFailed).
			WithErrorDetails(err.Error())
	}

	resrcs := make([]interfaces.Resource, 0)
	// 6. 初始化链路模型ID数组
	modelIDs = make([]string, len(reqModels))
	for i := range reqModels {
		modelIDs[i] = reqModels[i].ID

		resrcs = append(resrcs, interfaces.Resource{
			ID:   reqModels[i].ID,
			Type: interfaces.RESOURCE_TYPE_TRACE_MODEL,
			Name: reqModels[i].Name,
		})
	}

	// 注册资源策略
	err = tms.ps.CreateResources(ctx, resrcs, interfaces.COMMON_OPERATIONS)
	if err != nil {
		return nil, err
	}

	return modelIDs, nil
}

// 模拟创建链路模型
func (tms *traceModelService) SimulateCreateTraceModel(ctx context.Context, reqModel interfaces.TraceModel) (model interfaces.TraceModel, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 模拟创建链路模型")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 判断userid是否有创建链路模型的权限
	err = tms.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_TRACE_MODEL,
		ID:   interfaces.RESOURCE_ID_ALL,
	}, []string{interfaces.OPERATION_TYPE_CREATE})
	if err != nil {
		return interfaces.TraceModel{}, err
	}

	// 1. 根据数据视图 ID 获取对应的viewMap
	reqModels := []interfaces.TraceModel{reqModel}
	viewIDs := tms.getDependentViewIDs(reqModels)
	viewMap, err := tms.getDetailedViewMapByIDs(ctx, viewIDs)
	if err != nil {
		return reqModels[0], err
	}

	// 2. 根据数据连接名称获取对应的connMap
	connNames := tms.getDependentConnectionNames(reqModels)
	connMap, err := tms.getConnectionMapByNames(ctx, connNames)
	if err != nil {
		return reqModels[0], err
	}

	// 3. 校验reqModels所依赖的数据视图字段
	err = tms.validateReqTraceModels(ctx, reqModels, viewMap)
	if err != nil {
		return reqModels[0], err
	}

	// 4. 修改reqModels, 如更新依赖的数据连接ID, 生成update_time等
	err = tms.modifyReqModels(ctx, false, connMap, reqModels)
	if err != nil {
		return reqModels[0], err
	}

	return reqModels[0], nil
}

// 批量删除链路模型
func (tms *traceModelService) DeleteTraceModels(ctx context.Context, modelIDs []string) (err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 批量删除链路模型")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	matchResouces, err := tms.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_TRACE_MODEL, modelIDs,
		[]string{interfaces.OPERATION_TYPE_DELETE}, false)
	if err != nil {
		return err
	}
	// 资源过滤后的数量跟请求的数量不等，说明有部分模型没有权限，不能删除
	if len(matchResouces) != len(modelIDs) {
		return rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
			WithErrorDetails("Access denied: insufficient permissions for trace model's creation operation.")
	}

	err = tms.tma.DeleteTraceModels(ctx, modelIDs)
	if err != nil {
		logger.Errorf("Delete trace models failed, err: %v", err.Error())
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_TraceModel_InternalError_DeleteTraceModelsFailed).
			WithErrorDetails(err.Error())
	}

	//  清除资源策略
	err = tms.ps.DeleteResources(ctx, interfaces.RESOURCE_TYPE_TRACE_MODEL, modelIDs)
	if err != nil {
		return err
	}

	return nil
}

// 修改链路模型
func (tms *traceModelService) UpdateTraceModel(ctx context.Context, reqModel interfaces.TraceModel) (err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 修改链路模型")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	err = tms.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_TRACE_MODEL,
		ID:   reqModel.ID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}

	// 1. 根据数据视图 ID 获取对应的viewMap
	reqModels := []interfaces.TraceModel{reqModel}
	viewIDs := tms.getDependentViewIDs(reqModels)
	viewMap, err := tms.getDetailedViewMapByIDs(ctx, viewIDs)
	if err != nil {
		return err
	}

	// 2. 根据数据连接名称获取对应的connMap
	connNames := tms.getDependentConnectionNames(reqModels)
	connMap, err := tms.getConnectionMapByNames(ctx, connNames)
	if err != nil {
		return err
	}

	// 3. 校验reqModels所依赖的数据视图字段
	err = tms.validateReqTraceModels(ctx, reqModels, viewMap)
	if err != nil {
		return err
	}

	// 4. 修改reqModels, 如更新依赖的数据连接ID, 生成update_time等
	err = tms.modifyReqModels(ctx, true, connMap, reqModels)
	if err != nil {
		return err
	}

	// 5. 调用driven层, 修改链路模型
	err = tms.tma.UpdateTraceModel(ctx, reqModels[0])
	if err != nil {
		logger.Errorf("Update trace model failed, err: %v", err.Error())
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_TraceModel_InternalError_UpdateTraceModelFailed).
			WithErrorDetails(err.Error())
	}

	// 请求更新资源名称的接口，更新资源的名称
	err = tms.ps.UpdateResource(ctx, interfaces.Resource{
		ID:   reqModels[0].ID,
		Type: interfaces.RESOURCE_TYPE_TRACE_MODEL,
		Name: reqModels[0].Name,
	})
	if err != nil {
		return err
	}

	return nil
}

// 模拟修改链路模型
func (tms *traceModelService) SimulateUpdateTraceModel(ctx context.Context, reqModel interfaces.TraceModel) (model interfaces.TraceModel, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 模拟修改链路模型")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	err = tms.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_TRACE_MODEL,
		ID:   reqModel.ID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return interfaces.TraceModel{}, err
	}

	// 1. 根据数据视图 ID 获取对应的viewMap
	reqModels := []interfaces.TraceModel{reqModel}
	viewIDs := tms.getDependentViewIDs(reqModels)
	viewMap, err := tms.getDetailedViewMapByIDs(ctx, viewIDs)
	if err != nil {
		return reqModels[0], err
	}

	// 2. 根据数据连接名称获取对应的connMap
	connNames := tms.getDependentConnectionNames(reqModels)
	connMap, err := tms.getConnectionMapByNames(ctx, connNames)
	if err != nil {
		return reqModels[0], err
	}

	// 3. 校验reqModels所依赖的数据视图字段
	err = tms.validateReqTraceModels(ctx, reqModels, viewMap)
	if err != nil {
		return reqModels[0], err
	}

	// 4. 修改reqModels, 如更新依赖的数据连接ID, 生成update_time等
	err = tms.modifyReqModels(ctx, true, connMap, reqModels)
	if err != nil {
		return reqModels[0], err
	}

	return reqModels[0], nil
}

// 批量查询/导出链路模型
func (tms *traceModelService) GetTraceModels(ctx context.Context, modelIDs []string) (models []interfaces.TraceModel, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 批量查询链路模型")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 先获取资源序列
	matchResouces, err := tms.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_TRACE_MODEL, modelIDs,
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, true)
	if err != nil {
		return nil, err
	}

	// 资源过滤后的数量跟请求的数量不等，说明有部分模型没有权限，不能查看？还是返回有权限的
	if len(matchResouces) != len(modelIDs) {
		return nil, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
			WithErrorDetails("Access denied: insufficient permissions for trace model's view_detail operation.")
	}

	resModels := make([]interfaces.TraceModel, 0, len(modelIDs))

	// 1. 获取所有的链路模型
	modelMap, err := tms.tma.GetDetailedTraceModelMapByIDs(ctx, modelIDs)
	if err != nil {
		logger.Errorf("Get detailed trace model map by ids failed, err: %v", err.Error())
		return resModels, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_TraceModel_InternalError_GetDetailedTraceModelMapByIDsFailed).WithErrorDetails(err.Error())
	}

	// 2. 校验是否有模型不存在, 并整理resModels
	for _, modelID := range modelIDs {
		model, ok := modelMap[modelID]
		if !ok {
			errDetails := fmt.Sprintf("The trace model whose id equal to %v was not found", modelID)
			logger.Errorf(errDetails)
			o11y.Error(ctx, errDetails)
			return resModels, rest.NewHTTPError(ctx, http.StatusNotFound, derrors.DataModel_TraceModel_TraceModelNotFound).
				WithErrorDetails(errDetails)
		}

		// model 上附加 opreration 字段
		model.Operations = matchResouces[model.ID].Operations
		resModels = append(resModels, model)
	}

	// 3. 根据数据视图ID获取对应的viewMap
	viewIDs := tms.getDependentViewIDs(resModels)
	viewMap, err := tms.getSimpleViewMapByIDs(ctx, viewIDs)
	if err != nil {
		return resModels, err
	}

	// 4. 根据数据连接ID获取对应的connMap
	connIDs := tms.getDependentConnectionIDs(resModels)
	connMap, err := tms.getConnectionMapByIDs(ctx, connIDs)
	if err != nil {
		return resModels, err
	}

	// 5. 修改resModels, 如更新依赖的数据视图名称等
	tms.modifyResModels(ctx, viewMap, connMap, resModels)

	return resModels, nil
}

// 查询链路模型列表
func (tms *traceModelService) ListTraceModels(ctx context.Context,
	queryParams interfaces.TraceModelListQueryParams) (entries []interfaces.TraceModelListEntry,
	total int, err error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 查询链路模型列表与总数")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 1. 获取链路模型列表
	entries, err = tms.tma.ListTraceModels(ctx, queryParams)
	if err != nil {
		logger.Errorf("List trace models failed, err: %v", err.Error())
		return entries, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_TraceModel_InternalError_ListTraceModelsFailed).WithErrorDetails(err.Error())
	}

	if len(entries) == 0 {
		return entries, 0, nil
	}

	// 根据权限过滤有查看权限的对象，过滤后的数组的总长度就是总数，无需再请求总数
	// 处理资源id
	resMids := make([]string, 0)
	for _, m := range entries {
		resMids = append(resMids, m.ModelID)
	}
	matchResoucesMap, err := tms.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_TRACE_MODEL, resMids,
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, true)
	if err != nil {
		return entries, 0, err
	}

	// 遍历对象
	results := make([]interfaces.TraceModelListEntry, 0)
	for _, model := range entries {
		if resrc, exist := matchResoucesMap[model.ModelID]; exist {
			model.Operations = resrc.Operations // 用户当前有权限的操作
			results = append(results, model)
		}
	}

	// limit = -1,则返回所有
	if queryParams.Limit == -1 {
		return results, len(results), nil
	}

	// 分页
	// 检查起始位置是否越界
	if queryParams.Offset < 0 || queryParams.Offset >= len(results) {
		return nil, 0, nil
	}
	// 计算结束位置
	end := queryParams.Offset + queryParams.Limit
	if end > len(results) {
		end = len(results)
	}

	span.SetStatus(codes.Ok, "")
	return results[queryParams.Offset:end], len(results), nil

	// 2. 获取链路模型总数
	// total, err = tms.tma.GetTraceModelTotal(ctx, queryParams)
	// if err != nil {
	// 	logger.Errorf("Get trace model total failed, err: %v", err.Error())
	// 	return entries, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
	// 		derrors.DataModel_TraceModel_InternalError_GetTraceModelTotalFailed).WithErrorDetails(err.Error())
	// }

}

// 查询链路模型字段信息
func (tms *traceModelService) GetTraceModelFieldInfo(ctx context.Context, modelID string) (tmFieldInfo interfaces.TraceModelFieldInfo, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 查询链路模型字段信息")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 先获取资源序列
	matchResouces, err := tms.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_TRACE_MODEL, []string{modelID},
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, true)
	if err != nil {
		return interfaces.TraceModelFieldInfo{}, err
	}

	// 资源过滤后的数量跟请求的数量不等，说明有部分模型没有权限，不能查看？还是返回有权限的
	if len(matchResouces) != 1 {
		return interfaces.TraceModelFieldInfo{}, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
			WithErrorDetails("Access denied: insufficient permissions for trace model's view_detail operation.")
	}

	tmFieldInfo = interfaces.TraceModelFieldInfo{
		Span:       make([]interfaces.TraceModelField, 0),
		RelatedLog: make([]interfaces.TraceModelField, 0),
	}

	// 1. 获取所有的链路模型
	modelMap, err := tms.tma.GetDetailedTraceModelMapByIDs(ctx, []string{modelID})
	if err != nil {
		logger.Errorf("Get detailed trace model map by ids failed, err: %v", err.Error())
		return tmFieldInfo, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_TraceModel_InternalError_GetDetailedTraceModelMapByIDsFailed).WithErrorDetails(err.Error())
	}

	// 2. 校验模型是否存在
	model, ok := modelMap[modelID]
	if !ok {
		errDetails := fmt.Sprintf("The trace model whose id equal to %v was not found", modelID)
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		return tmFieldInfo, rest.NewHTTPError(ctx, http.StatusNotFound, derrors.DataModel_TraceModel_TraceModelNotFound).
			WithErrorDetails(errDetails)
	}

	sourceType, err := tms.getUnderlyingDataSourceType(ctx, interfaces.QUERY_CATEGORY_SPAN, model)
	if err != nil {
		return tmFieldInfo, err
	}

	// 3. 处理span的字段信息
	processor, err := data_source.NewTraceModelProcessor(ctx, tms.appSetting, sourceType)
	if err != nil {
		o11y.Error(ctx, err.Error())
		return tmFieldInfo, err
	}

	spanFieldInfos, err := processor.GetSpanFieldInfo(ctx, model)
	if err != nil {
		return tmFieldInfo, err
	}
	tmFieldInfo.Span = append(tmFieldInfo.Span, spanFieldInfos...)

	// 4. 处理related_log的字段信息
	if model.EnabledRelatedLog == interfaces.RELATED_LOG_OPEN {
		sourceType, err := tms.getUnderlyingDataSourceType(ctx, interfaces.QUERY_CATEGORY_RELATED_LOG, model)
		if err != nil {
			return tmFieldInfo, err
		}

		processor, err := data_source.NewTraceModelProcessor(ctx, tms.appSetting, sourceType)
		if err != nil {
			o11y.Error(ctx, err.Error())
			return tmFieldInfo, err
		}

		relatedLogFieldInfos, err := processor.GetRelatedLogFieldInfo(ctx, model)
		if err != nil {
			return tmFieldInfo, err
		}
		tmFieldInfo.RelatedLog = append(tmFieldInfo.RelatedLog, relatedLogFieldInfos...)
	}

	return tmFieldInfo, nil
}

func (tms *traceModelService) ListTraceModelSrcs(ctx context.Context, param interfaces.TraceModelListQueryParams) ([]interfaces.Resource, int, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "查询链路模型实例列表")
	span.End()

	models, err := tms.tma.ListTraceModels(ctx, param)
	emptyResources := []interfaces.Resource{}
	if err != nil {
		logger.Errorf("ListTraceModels error: %s", err.Error())
		span.SetStatus(codes.Error, "List simple trace models error")
		span.End()
		return emptyResources, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_TraceModel_InternalError_ListTraceModelsFailed).WithErrorDetails(err.Error())
	}
	if len(models) == 0 {
		return emptyResources, 0, nil
	}

	// 根据权限过滤有查看权限的对象，过滤后的数组的总长度就是总数，无需再请求总数
	resMids := make([]string, 0)
	for _, m := range models {
		resMids = append(resMids, m.ModelID)
	}
	// 校验权限管理的操作权限
	matchResoucesMap, err := tms.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_TRACE_MODEL, resMids,
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, false)
	if err != nil {
		return emptyResources, 0, err
	}

	// 遍历对象
	results := make([]interfaces.Resource, 0)
	for _, model := range models {
		if _, exist := matchResoucesMap[model.ModelID]; exist {
			results = append(results, interfaces.Resource{
				ID:   model.ModelID,
				Type: interfaces.RESOURCE_TYPE_TRACE_MODEL,
				Name: model.ModelName,
			})
		}
	}

	// limit = -1,则返回所有
	if param.Limit == -1 {
		return results, len(results), nil
	}

	// 分页
	// 检查起始位置是否越界
	if param.Offset < 0 || param.Offset >= len(results) {
		return nil, 0, nil
	}
	// 计算结束位置
	end := param.Offset + param.Limit
	if end > len(results) {
		end = len(results)
	}

	span.SetStatus(codes.Ok, "")
	return results[param.Offset:end], len(results), nil
}

// 根据链路模型ID数组去获取ID与简单对象的映射关系
func (tms *traceModelService) GetSimpleTraceModelMapByIDs(ctx context.Context, modelIDs []string) (modelMap map[string]interfaces.TraceModel, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 根据链路模型ID数组获取对应的simple map(key为链路模型ID, value为链路模型simple对象)")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	modelMap, err = tms.tma.GetSimpleTraceModelMapByIDs(ctx, modelIDs)
	if err != nil {
		logger.Errorf("Get simple trace model map by ids failed, err: %v", err.Error())
		return modelMap, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_TraceModel_InternalError_GetSimpleTraceModelMapByIDsFailed).
			WithErrorDetails(err.Error())
	}

	return modelMap, nil
}

// 根据链路模型Name数组去获取Name与对象的映射关系
func (tms *traceModelService) GetSimpleTraceModelMapByNames(ctx context.Context, modelNames []string) (modelMap map[string]interfaces.TraceModel, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 根据链路模型名称数组获取对应的simple map(key为链路模型名称, value为链路模型simple对象)")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 调用driven层, 获取Name与对象的映射关系
	modelMap, err = tms.tma.GetSimpleTraceModelMapByNames(ctx, modelNames)
	if err != nil {
		logger.Errorf("Get simple trace model map by names failed, err: %v", err.Error())
		return modelMap, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_TraceModel_InternalError_GetSimpleTraceModelMapByNamesFailed).
			WithErrorDetails(err.Error())
	}

	return modelMap, nil
}

/*
	私有方法
*/

// 获取reqModels中所有依赖的数据视图名称
// func (tms *traceModelService) getDependentViewNames(reqModels []interfaces.TraceModel) []string {
// 	m := make(map[string]struct{})
// 	for _, reqModel := range reqModels {
// 		if reqModel.SpanSourceType == interfaces.SOURCE_TYPE_DATA_VIEW {
// 			spanConf := reqModel.SpanConfig.(interfaces.SpanConfigWithDataView)
// 			// span所在数据视图
// 			m[spanConf.DataView.Name] = struct{}{}
// 		}

// 		if reqModel.EnabledRelatedLog == interfaces.RELATED_LOG_OPEN {
// 			if reqModel.RelatedLogSourceType == interfaces.SOURCE_TYPE_DATA_VIEW {
// 				relatedLogConf := reqModel.RelatedLogConfig.(interfaces.RelatedLogConfigWithDataView)
// 				// related log所在数据视图
// 				m[relatedLogConf.DataView.Name] = struct{}{}
// 			}
// 		}
// 	}

// 	viewNames := make([]string, 0)
// 	for name := range m {
// 		viewNames = append(viewNames, name)
// 	}

// 	return viewNames
// }

// 根据数据视图 ID 数组, 获取对应的viewMap(value为详细的view对象)
func (tms *traceModelService) getDetailedViewMapByIDs(ctx context.Context, viewIDs []string) (viewMap map[string]*interfaces.DataView, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 获取链路模型依赖的数据视图map(key为数据视图ID, value为数据视图完整对象)")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 1. 临界情况判断
	if len(viewIDs) == 0 {
		return nil, nil
	}

	// 2. 根据数据视图 ID 数组获取数据视图详情, 并校验其存在性
	viewMap, err = tms.dvs.GetDetailedDataViewMapByIDs(ctx, viewIDs)
	if err != nil {
		return nil, err
	}

	for _, ID := range viewIDs {
		if _, ok := viewMap[ID]; !ok {
			errDetails := fmt.Sprintf("The dependent data view ID %v does not exist in the database!", ID)
			logger.Errorf(errDetails)
			o11y.Error(ctx, errDetails)
			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_DependentDataViewNotFound).
				WithErrorDetails(errDetails)
		}
	}

	// 3. 根据数据视图的Fields得到FieldsMap
	for _, view := range viewMap {
		fieldsMap := make(map[string]string)

		// 前端传递的是name, 这里需要维护name和type的映射
		for _, field := range view.Fields {
			fieldsMap[field.Name] = field.Type
		}

		view.FieldTypeMap = fieldsMap
		viewMap[view.ViewID] = view
	}

	return viewMap, nil
}

// 获取reqModels中所有依赖的数据连接名称
func (tms *traceModelService) getDependentConnectionNames(reqModels []interfaces.TraceModel) []string {
	m := make(map[string]struct{})
	for _, reqModel := range reqModels {
		if reqModel.SpanSourceType == interfaces.SOURCE_TYPE_DATA_CONNECTION {
			spanConf := reqModel.SpanConfig.(interfaces.SpanConfigWithDataConnection)
			// span所在数据视图
			m[spanConf.DataConnection.Name] = struct{}{}
		}
	}

	connNames := make([]string, 0)
	for name := range m {
		connNames = append(connNames, name)
	}

	return connNames
}

// 根据数据视图名称数组, 获取对应的viewMap(value为详细的view对象)
func (tms *traceModelService) getConnectionMapByNames(ctx context.Context, connNames []string) (connMap map[string]string, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 获取链路模型依赖的数据连接map(key为数据连接名称, value为数据连接简单对象)")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 1. 临界情况判断
	if len(connNames) == 0 {
		return nil, nil
	}

	// 2. 根据数据视图Name数组获取数据视图详情
	connMap, err = tms.dcs.GetMapAboutName2ID(ctx, connNames)
	if err != nil {
		return nil, err
	}

	// 3. 校验存在性
	for _, name := range connNames {
		if _, ok := connMap[name]; !ok {
			errDetails := fmt.Sprintf("The dependent data connection named %v does not exist in the database!", name)
			logger.Errorf(errDetails)
			o11y.Error(ctx, errDetails)
			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_DependentDataConnectionNotFound).
				WithErrorDetails(errDetails)
		}
	}

	return connMap, nil
}

// 校验reqModels所依赖的数据视图字段
func (tms *traceModelService) validateReqTraceModels(ctx context.Context, reqModels []interfaces.TraceModel, viewMap map[string]*interfaces.DataView) (err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 校验所有链路模型所需的数据视图字段是否存在")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	for _, model := range reqModels {
		if model.SpanSourceType == interfaces.SOURCE_TYPE_DATA_VIEW {
			// 校验Span的基础属性配置
			spanConf := model.SpanConfig.(interfaces.SpanConfigWithDataView)
			viewID := spanConf.DataView.ID
			err := tms.validateSpanBasicAttrs(ctx, spanConf, viewMap[viewID].FieldTypeMap)
			if err != nil {
				o11y.Error(ctx, err.Error())
				return err
			}
		}

		if model.EnabledRelatedLog == interfaces.RELATED_LOG_OPEN {
			if model.RelatedLogSourceType == interfaces.SOURCE_TYPE_DATA_VIEW {
				// 校验Span的关联日志配置
				relatedLogConf := model.RelatedLogConfig.(interfaces.RelatedLogConfigWithDataView)
				viewID := relatedLogConf.DataView.ID
				err := tms.validateRelatedLogWithDataView(ctx, relatedLogConf, viewMap[viewID].FieldTypeMap)
				if err != nil {
					o11y.Error(ctx, err.Error())
					return err
				}
			}
		}
	}

	return nil
}

// 校验Span的基础属性配置
func (tms *traceModelService) validateSpanBasicAttrs(ctx context.Context, attrs interfaces.SpanConfigWithDataView, fieldMap map[string]string) error {
	/*
		part1. 必填字段校验
		包括: TraceID, SpanID, ParentSpanID, Name, StartTime, ServiceName
	*/

	// 1.1 TraceID校验
	fieldName := attrs.TraceID.FieldName
	fieldType, ok := fieldMap[fieldName]
	if !ok {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_TraceID).
			WithErrorDetails(fmt.Sprintf("Field %v does not exist in the selected data view", fieldName))
	}

	if !tms.isValidFieldType(fieldType, interfaces.VALID_FIELD_TYPES_FOR_TRACE_ID) {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_TraceID).
			WithErrorDetails(fmt.Sprintf("The type of field %v is invalid, valid type is in %v", fieldName, interfaces.VALID_FIELD_TYPES_FOR_TRACE_ID))
	}

	// 1.2 SpanID校验
	for _, fieldName := range attrs.SpanID.FieldNames {
		fieldType, ok := fieldMap[fieldName]
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_SpanID).
				WithErrorDetails(fmt.Sprintf("Field %v does not exist in the selected data view", fieldName))
		}

		if !tms.isValidFieldType(fieldType, interfaces.VALID_FIELD_TYPES_FOR_SPAN_ID) {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_SpanID).
				WithErrorDetails(fmt.Sprintf("The type of field %v is invalid, valid type is in %v", fieldName, interfaces.VALID_FIELD_TYPES_FOR_SPAN_ID))
		}
	}

	// 1.3 ParentSpanID校验
	for _, config := range attrs.ParentSpanID {
		// 1.3.1 前置条件校验
		if err := tms.validPrecond(config.Precond, fieldMap); err != nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_ParentSpanID).
				WithErrorDetails(err)
		}

		// 1.3.2 FieldNames校验
		for _, fieldName := range config.FieldNames {
			fieldType, ok := fieldMap[fieldName]
			if !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_ParentSpanID).
					WithErrorDetails(fmt.Sprintf("Field %v does not exist in the selected data view", fieldName))
			}

			if !tms.isValidFieldType(fieldType, interfaces.VALID_FIELD_TYPES_FOR_PARENT_SPAN_ID) {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_ParentSpanID).
					WithErrorDetails(fmt.Sprintf("The type of field %v is invalid, valid type is in %v", fieldName, interfaces.VALID_FIELD_TYPES_FOR_PARENT_SPAN_ID))
			}
		}
	}

	// 1.4 Name校验
	fieldName = attrs.Name.FieldName
	fieldType, ok = fieldMap[fieldName]
	if !ok {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_Name).
			WithErrorDetails(fmt.Sprintf("Field %v does not exist in the selected data view", fieldName))
	}

	if !tms.isValidFieldType(fieldType, interfaces.VALID_FIELD_TYPES_FOR_NAME) {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_Name).
			WithErrorDetails(fmt.Sprintf("The type of field %v is invalid, valid type is in %v", fieldName, interfaces.VALID_FIELD_TYPES_FOR_NAME))
	}

	// 1.5 StartTime校验
	fieldName = attrs.StartTime.FieldName
	fieldType, ok = fieldMap[fieldName]
	if !ok {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_StartTime).
			WithErrorDetails(fmt.Sprintf("Field %v does not exist in the selected data view", fieldName))
	}

	if !tms.isValidFieldType(fieldType, interfaces.VALID_FIELD_TYPES_FOR_START_TIME) {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_StartTime).
			WithErrorDetails(fmt.Sprintf("The type of field %v is invalid, valid type is in %v", fieldName, interfaces.VALID_FIELD_TYPES_FOR_START_TIME))
	}

	// 1.6 ServiceName校验
	fieldName = attrs.ServiceName.FieldName
	fieldType, ok = fieldMap[fieldName]
	if !ok {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_ServiceName).
			WithErrorDetails(fmt.Sprintf("Field %v does not exist in the selected data view", fieldName))
	}

	if !tms.isValidFieldType(fieldType, interfaces.VALID_FIELD_TYPES_FOR_SERVICE_NAME) {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_ServiceName).
			WithErrorDetails(fmt.Sprintf("The type of field %v is invalid, valid type is in %v", fieldName, interfaces.VALID_FIELD_TYPES_FOR_SERVICE_NAME))
	}

	/*
		part2. 非必填字段校验
		包括: EndTime, Duration, Kind, Status
	*/

	// 2.1 EndTime校验
	if fieldName := attrs.EndTime.FieldName; fieldName != "" {
		fieldType, ok = fieldMap[fieldName]
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_EndTime).
				WithErrorDetails(fmt.Sprintf("Field %v does not exist in the selected data view", fieldName))
		}

		if !tms.isValidFieldType(fieldType, interfaces.VALID_FIELD_TYPES_FOR_END_TIME) {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_EndTime).
				WithErrorDetails(fmt.Sprintf("The type of field %v is invalid, valid type is in %v", fieldName, interfaces.VALID_FIELD_TYPES_FOR_END_TIME))
		}
	}

	// 2.2 Duration校验
	if fieldName := attrs.Duration.FieldName; fieldName != "" {
		fieldType, ok = fieldMap[fieldName]
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_Duration).
				WithErrorDetails(fmt.Sprintf("Field %v does not exist in the selected data view", fieldName))
		}

		if !tms.isValidFieldType(fieldType, interfaces.VALID_FIELD_TYPES_FOR_DURATION) {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_Duration).
				WithErrorDetails(fmt.Sprintf("The type of field %v is invalid, valid type is in %v", fieldName, interfaces.VALID_FIELD_TYPES_FOR_DURATION))
		}
	}

	// 2.3 Kind校验
	if fieldName := attrs.Kind.FieldName; fieldName != "" {
		fieldType, ok = fieldMap[fieldName]
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_Kind).
				WithErrorDetails(fmt.Sprintf("Field %v does not exist in the selected data view", fieldName))
		}

		if !tms.isValidFieldType(fieldType, interfaces.VALID_FIELD_TYPES_FOR_KIND) {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_Kind).
				WithErrorDetails(fmt.Sprintf("The type of field %v is invalid, valid type is in %v", fieldName, interfaces.VALID_FIELD_TYPES_FOR_KIND))
		}
	}

	// 2.4 Status校验
	if fieldName := attrs.Status.FieldName; fieldName != "" {
		fieldType, ok = fieldMap[fieldName]
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_Status).
				WithErrorDetails(fmt.Sprintf("Field %v does not exist in the selected data view", fieldName))
		}

		if !tms.isValidFieldType(fieldType, interfaces.VALID_FIELD_TYPES_FOR_STATUS) {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_Status).
				WithErrorDetails(fmt.Sprintf("The type of field %v is invalid, valid type is in %v", fieldName, interfaces.VALID_FIELD_TYPES_FOR_STATUS))
		}
	}

	return nil
}

// 校验是否为合法的字段类型
func (tms *traceModelService) isValidFieldType(fieldType string, validDataSet []string) bool {
	for _, validType := range validDataSet {
		if fieldType == validType {
			return true
		}
	}

	return false
}

// 校验前置条件里的字段类型是否合法
func (tms *traceModelService) validPrecond(precond *interfaces.CondCfg, fieldMap map[string]string) error {
	if precond == nil {
		return nil
	}

	switch precond.Operation {
	case "":
		return nil
	case dcond.OperationAnd, dcond.OperationOr:
		for _, subCond := range precond.SubConds {
			if err := tms.validPrecond(subCond, fieldMap); err != nil {
				return err
			}
		}
		return nil
	case dcond.Operation_RANGE:
		t, ok := fieldMap[precond.Name]
		if !ok {
			return fmt.Errorf("field %v does not exist in the selected data view", precond.Name)
		}

		if t != dtype.DataType_Datetime {
			return fmt.Errorf("field %v is not date type, can not use operation range", precond.Name)
		}
		return nil
	default:
		if _, ok := fieldMap[precond.Name]; !ok {
			return fmt.Errorf("field %v does not exist in the selected data view", precond.Name)
		}

		if precond.ValueFrom == dcond.ValueFrom_Field {
			valueStr, _ := precond.Value.(string) // driver层已经校验过, 所以类型转换一定不报错
			if _, ok := fieldMap[valueStr]; !ok {
				return fmt.Errorf("field %v does not exist in the selected data view", valueStr)
			}
		}
		return nil
	}
}

// 校验Span的关联日志配置
func (tms *traceModelService) validateRelatedLogWithDataView(ctx context.Context, relatedLog interfaces.RelatedLogConfigWithDataView, fieldMap map[string]string) error {
	// 1. TraceID校验
	fieldName := relatedLog.TraceID.FieldName
	fieldType, ok := fieldMap[fieldName]
	if !ok {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidRelatedLogConfig_TraceID).
			WithErrorDetails(fmt.Sprintf("Field %v does not exist in the selected data view", fieldName))
	}

	if !tms.isValidFieldType(fieldType, interfaces.VALID_FIELD_TYPES_FOR_TRACE_ID) {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidRelatedLogConfig_TraceID).
			WithErrorDetails(fmt.Sprintf("The type of field %v is invalid, valid type is in %v", fieldName, interfaces.VALID_FIELD_TYPES_FOR_TRACE_ID))
	}

	// 2. SpanID校验
	for _, fieldName := range relatedLog.SpanID.FieldNames {
		fieldType, ok := fieldMap[fieldName]
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidRelatedLogConfig_SpanID).
				WithErrorDetails(fmt.Sprintf("Field %v does not exist in the selected data view", fieldName))
		}

		if !tms.isValidFieldType(fieldType, interfaces.VALID_FIELD_TYPES_FOR_SPAN_ID) {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidRelatedLogConfig_SpanID).
				WithErrorDetails(fmt.Sprintf("The type of field %v is invalid, valid type is in %v", fieldName, interfaces.VALID_FIELD_TYPES_FOR_SPAN_ID))
		}
	}

	return nil
}

// 修改reqModels
func (tms *traceModelService) modifyReqModels(ctx context.Context, isUpdate bool, connMap map[string]string, reqModels []interfaces.TraceModel) (err error) {
	_, span := ar_trace.Tracer.Start(ctx, "logic层: 修改所有传入的链路模型(包括更新数据视图/数据连接ID等操作)")

	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	for i := range reqModels {
		if reqModels[i].SpanSourceType == interfaces.SOURCE_TYPE_DATA_VIEW {
			spanConf := reqModels[i].SpanConfig.(interfaces.SpanConfigWithDataView)
			reqModels[i].SpanConfig = spanConf
		} else {
			// 1. 更新Span所在数据连接的ID
			spanConf := reqModels[i].SpanConfig.(interfaces.SpanConfigWithDataConnection)
			connName := spanConf.DataConnection.Name
			spanConf.DataConnection.ID = connMap[connName]
			reqModels[i].SpanConfig = spanConf
		}

		// 2. 更新关联日志所在数据视图的配置
		if reqModels[i].EnabledRelatedLog == interfaces.RELATED_LOG_OPEN {
			if reqModels[i].RelatedLogSourceType == interfaces.SOURCE_TYPE_DATA_VIEW {
				relatedLogConf := reqModels[i].RelatedLogConfig.(interfaces.RelatedLogConfigWithDataView)
				reqModels[i].RelatedLogConfig = relatedLogConf
			}
		}

		// 3. 如果是创建链路模型, 则要为链路模型生成相应的分布式ID
		if !isUpdate {
			reqModels[i].ID = xid.New().String()
		}

		// 4. 生成更新时间
		reqModels[i].CreateTime = time.Now().UnixMilli()
		reqModels[i].UpdateTime = reqModels[i].CreateTime
	}

	return nil
}

// 获取resModels中所有依赖的数据视图ID
func (tms *traceModelService) getDependentViewIDs(resModels []interfaces.TraceModel) []string {
	m := make(map[string]struct{})
	for _, resModel := range resModels {
		if resModel.SpanSourceType == interfaces.SOURCE_TYPE_DATA_VIEW {
			spanConf := resModel.SpanConfig.(interfaces.SpanConfigWithDataView)
			// span所在数据视图
			m[spanConf.DataView.ID] = struct{}{}
		}

		if resModel.EnabledRelatedLog == interfaces.RELATED_LOG_OPEN {
			if resModel.RelatedLogSourceType == interfaces.SOURCE_TYPE_DATA_VIEW {
				relatedLogConf := resModel.RelatedLogConfig.(interfaces.RelatedLogConfigWithDataView)
				// related log所在数据视图
				m[relatedLogConf.DataView.ID] = struct{}{}
			}
		}
	}

	viewIDs := make([]string, 0)
	for id := range m {
		viewIDs = append(viewIDs, id)
	}

	return viewIDs
}

// 根据数据视图名称ID, 获取对应的viewMap(value为简单的view对象)
func (tms *traceModelService) getSimpleViewMapByIDs(ctx context.Context, viewIDs []string) (viewMap map[string]*interfaces.DataView, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 获取链路模型依赖的数据视图simple map(key为数据视图ID, value为数据视图simple对象)")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	viewMap, err = tms.dvs.GetSimpleDataViewsByIDs(ctx, viewIDs, true)
	if err != nil {
		return nil, err
	}

	for _, id := range viewIDs {
		if _, ok := viewMap[id]; !ok {
			errDetails := fmt.Sprintf("The dependent data view whose id equal to %v does not exist in the database!", id)
			logger.Warnf(errDetails)
			o11y.Warn(ctx, errDetails)
		}
	}

	return viewMap, nil
}

// 获取resModels中所有依赖的数据连接ID
func (tms *traceModelService) getDependentConnectionIDs(resModels []interfaces.TraceModel) []string {
	m := make(map[string]struct{})
	for _, resModel := range resModels {
		if resModel.SpanSourceType == interfaces.SOURCE_TYPE_DATA_CONNECTION {
			spanConf := resModel.SpanConfig.(interfaces.SpanConfigWithDataConnection)
			// span所在数据视图
			m[spanConf.DataConnection.ID] = struct{}{}
		}
	}

	viewIDs := make([]string, 0)
	for id := range m {
		viewIDs = append(viewIDs, id)
	}

	return viewIDs
}

// 根据数据视图名称数组, 获取对应的viewMap(value为详细的view对象)
func (tms *traceModelService) getConnectionMapByIDs(ctx context.Context, connIDs []string) (connMap map[string]string, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 获取链路模型依赖的数据连接map(key为数据连接ID, value为数据连接名称)")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 1. 临界情况判断
	if len(connIDs) == 0 {
		return nil, nil
	}

	// 2. 根据数据视图ID数组获取数据连接map
	connMap, err = tms.dcs.GetMapAboutID2Name(ctx, connIDs)
	if err != nil {
		return nil, err
	}

	// 3. 校验存在性
	for _, id := range connIDs {
		if _, ok := connMap[id]; !ok {
			errDetails := fmt.Sprintf("The dependent data connection whose id equal to %v does not exist in the database!", id)
			logger.Warnf(errDetails)
			o11y.Warn(ctx, errDetails)
		}
	}

	return connMap, nil
}

// 修改resModels
func (tms *traceModelService) modifyResModels(ctx context.Context, viewMap map[string]*interfaces.DataView, connMap map[string]string, resModels []interfaces.TraceModel) {
	_, span := ar_trace.Tracer.Start(ctx, "logic层: 修改链路模型列表, 更新依赖的数据视图/数据连接名称")
	defer func() {
		span.SetStatus(codes.Ok, "")
		span.End()
	}()

	for i := range resModels {
		// 1. 更新Span所在数据视图/数据连接的Name
		if resModels[i].SpanSourceType == interfaces.SOURCE_TYPE_DATA_VIEW {
			spanConf := resModels[i].SpanConfig.(interfaces.SpanConfigWithDataView)
			viewID := spanConf.DataView.ID
			if view, ok := viewMap[viewID]; ok {
				spanConf.DataView.Name = view.ViewName
				resModels[i].SpanConfig = spanConf
			}
		} else {
			spanConf := resModels[i].SpanConfig.(interfaces.SpanConfigWithDataConnection)
			connID := spanConf.DataConnection.ID
			if connName, ok := connMap[connID]; ok {
				spanConf.DataConnection.Name = connName
				resModels[i].SpanConfig = spanConf
			}
		}

		// 2. 更新关联日志所在数据视图的ID
		if resModels[i].EnabledRelatedLog == interfaces.RELATED_LOG_OPEN {
			if resModels[i].RelatedLogSourceType == interfaces.SOURCE_TYPE_DATA_VIEW {
				relatedLogConf := resModels[i].RelatedLogConfig.(interfaces.RelatedLogConfigWithDataView)
				viewID := relatedLogConf.DataView.ID
				relatedLogConf.DataView.Name = viewMap[viewID].ViewName
				resModels[i].RelatedLogConfig = relatedLogConf
			}
		}
	}
}

// 获取底层数据源类型
func (tms *traceModelService) getUnderlyingDataSourceType(ctx context.Context, queryCategory string, model interfaces.TraceModel) (sourceType string, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 查询数据来源类型")
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
		underlyingSourceType, isExist, err := tms.dcs.GetDataConnectionSourceType(ctx, spanConfig.DataConnection.ID)
		if err != nil {
			errDetails := fmt.Sprintf("Get underlying data souce type failed, err: %v", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			return underlyingSourceType, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_TraceModel_InternalError_GetUnderlyingDataSourceTypeFailed).
				WithErrorDetails(err.Error())
		}

		if !isExist {
			errDetails := fmt.Sprintf("Get underlying data souce type failed, the data connection whose id equal to %v was not found", spanConfig.DataConnection.ID)
			logger.Errorf(errDetails)
			o11y.Error(ctx, errDetails)
			return underlyingSourceType, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_TraceModel_InternalError_GetUnderlyingDataSourceTypeFailed).
				WithErrorDetails(errDetails)
		}
		return underlyingSourceType, nil
	}
}
