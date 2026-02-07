// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package objective_model

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"sync"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/panjf2000/ants/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"uniquery/common"
	"uniquery/common/convert"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	"uniquery/logics"
	"uniquery/logics/metric_model"
	"uniquery/logics/permission"
)

var (
	omServiceOnce sync.Once
	omService     interfaces.ObjectiveModelService
	objectivePool *ants.Pool
)

type objectiveModelService struct {
	appSetting *common.AppSetting
	mmAccess   interfaces.MetricModelAccess
	mmService  interfaces.MetricModelService
	omAccess   interfaces.ObjectiveModelAccess
	ps         interfaces.PermissionService
}

func NewobjectiveModelService(appSetting *common.AppSetting) interfaces.ObjectiveModelService {
	omServiceOnce.Do(func() {
		omService = &objectiveModelService{
			appSetting: appSetting,
			mmAccess:   logics.MMAccess,
			mmService:  metric_model.NewMetricModelService(appSetting),
			omAccess:   logics.OMAccess,
			ps:         permission.NewPermissionService(appSetting),
		}

		InitObjectivePool(appSetting.PoolSetting)
	})
	return omService
}

// 初始化协程池
func InitObjectivePool(poolSetting common.PoolSetting) {
	pool, err := ants.NewPool(poolSetting.ObjectivePoolSize, ants.WithPreAlloc(true), ants.WithNonblocking(false))
	if err != nil {
		logger.Fatalf("Init objective get metric pool failed, %s", err.Error())
		panic(err)
	}

	objectivePool = pool
}

// 目标模型的数据预览
func (oms *objectiveModelService) Simulate(ctx context.Context,
	query interfaces.ObjectiveModelQuery) (interfaces.ObjectiveModelUniResponse, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "查询需预览的目标模型数据")
	defer span.End()

	// 决策权限。 预览的时候还没有模型id，此时的预览校验用新建或者编辑，
	ops, err := oms.ps.GetResourcesOperations(ctx, interfaces.RESOURCE_TYPE_OBJECTIVE_MODEL,
		[]string{interfaces.RESOURCE_ID_ALL})
	if err != nil {
		return interfaces.ObjectiveModelUniResponse{}, err
	}
	if len(ops) != 1 {
		// 无权限
		return interfaces.ObjectiveModelUniResponse{}, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
			WithErrorDetails("Access denied: insufficient permissions for objective model's create or modify operation.")
	}
	// 从 ops 里找新建或编辑的权限
	for _, op := range ops[0].Operations {
		if op != interfaces.OPERATION_TYPE_CREATE && op != interfaces.OPERATION_TYPE_DELETE {
			return interfaces.ObjectiveModelUniResponse{}, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
				WithErrorDetails("Access denied: insufficient permissions for objective model's create or modify operation.")
		}
	}

	resp, err := oms.eval(ctx, query)
	if err != nil {
		// 添加异常时的 trace 属性
		span.SetStatus(codes.Error, "Eval Objective Model error")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Eval Objective Model error: %v", err))
		return resp, err
	}

	span.SetStatus(codes.Ok, "")
	return resp, nil
}

// 基于目标模型的指标数据查询
func (ds *objectiveModelService) Exec(ctx context.Context, query interfaces.ObjectiveModelQuery) (interfaces.ObjectiveModelUniResponse, error) {
	var resps interfaces.ObjectiveModelUniResponse

	// 获取目标模型信息
	ctx, span := ar_trace.Tracer.Start(ctx, "查询目标模型的指标数据")
	defer span.End()

	// 决策当前模型id的数据查询权限
	err := ds.ps.CheckPermission(ctx, interfaces.Resource{
		ID:   query.ObjectiveModelID,
		Type: interfaces.RESOURCE_TYPE_OBJECTIVE_MODEL,
	}, []string{interfaces.OPERATION_TYPE_DATA_QUERY})
	if err != nil {
		return resps, err
	}

	objectiveModel, exists, err := ds.omAccess.GetObjectiveModel(ctx, query.ObjectiveModelID)
	if err != nil {
		logger.Errorf("Get Objective Model error: %s", err.Error())

		// 添加异常时的 trace 属性
		span.SetAttributes(attribute.Key("model_id").String(query.ObjectiveModelID))
		span.SetStatus(codes.Error, "Get Objective Model error")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Get Objective Model error: %v", err))

		return resps, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			uerrors.Uniquery_ObjectiveModel_InternalError_GetModelByIdFailed).WithErrorDetails(err.Error())
	}
	if !exists {
		logger.Debugf("Objective Model %s not found!", query.ObjectiveModelID)

		// 添加异常时的 trace 属性
		span.SetAttributes(attribute.Key("model_id").String(query.ObjectiveModelID))
		span.SetStatus(codes.Error, "Objective Model not found!")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Objective Model [%s] not found!", query.ObjectiveModelID))

		return resps, rest.NewHTTPError(ctx, http.StatusNotFound, uerrors.Uniquery_ObjectiveModel_ObjectiveModelNotFound)
	}

	query.ObjectiveConfig = objectiveModel.ObjectiveConfig
	query.ObjectiveType = objectiveModel.ObjectiveType

	// 发起查询
	respi, err := ds.eval(ctx, query)
	if err != nil {
		// 添加异常时的 trace 属性
		span.SetAttributes(attribute.Key("model_id").String(query.ObjectiveModelID))
		span.SetStatus(codes.Error, "Eval Objective Model error")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Eval Objective Model error: %v", err))

		return resps, err
	}

	if query.IncludeModel {
		respi.Model = objectiveModel
	}

	// ok
	span.SetStatus(codes.Ok, "")

	return respi, nil
}

// 按目标模型的信息查询指标数据
func (ds *objectiveModelService) eval(ctx context.Context, query interfaces.ObjectiveModelQuery) (interfaces.ObjectiveModelUniResponse, error) {

	var resp interfaces.ObjectiveModelUniResponse

	// 按 objective_type 做相应的计算
	switch query.ObjectiveType {
	case interfaces.SLO:
		sloObjective := query.ObjectiveConfig.(interfaces.SLOObjective)

		// 获取关联指标模型的基本信息，然后判断模型是否存在，模型的单位是否相同。
		err := ds.checkMetricModelsUnit(ctx, []string{sloObjective.GoodMetricModel.Id, sloObjective.TotalMetricModel.Id})
		if err != nil {
			return resp, err
		}

		results := make([]metricResult, 0)
		errCh := make(chan error, 2)
		defer close(errCh)

		taskCtx := &taskContext{
			ctx:     ctx,
			results: results,
			wg:      sync.WaitGroup{},
			errCh:   errCh,
		}

		// 并发查询
		for _, id := range []string{sloObjective.GoodMetricModel.Id, sloObjective.TotalMetricModel.Id} {
			taskCtx.wg.Add(1)

			err := objectivePool.Submit(ds.getMetricModelData(taskCtx, query, id))
			if err != nil {
				detail := fmt.Sprintf("submit task of get metric data failed, %s", err.Error())
				return resp, rest.NewHTTPError(ctx, http.StatusInternalServerError,
					uerrors.Uniquery_ObjectiveModel_InternalError_SubmitGetMetricDataTaskFailed).WithErrorDetails(detail)
			}
		}

		// 读取数据
		taskCtx.wg.Wait()

		if len(taskCtx.errCh) > 0 {
			err := <-taskCtx.errCh
			return resp, err
		}

		// 数组转map
		metricMap := make(map[string]map[string]interfaces.MetricModelData)
		for _, res := range taskCtx.results {
			metricMap[res.metricModelID] = res.seriesMap
		}

		// 计算其他指标
		// 把查询结果处理成统一的格式。当前的预览是query_range查询，返回结果为 matrix
		return processSLOObjectiveResponse(ctx, metricMap, sloObjective)

	case interfaces.KPI:
		kpiObjective := query.ObjectiveConfig.(interfaces.KPIObjective)
		metricModelIds := make([]string, 0)
		for _, mm := range kpiObjective.ComprehensiveMetricModels {
			metricModelIds = append(metricModelIds, mm.Id)
		}
		for _, mm := range kpiObjective.AdditionalMetricModels {
			metricModelIds = append(metricModelIds, mm.Id)
		}

		// 获取关联指标模型的基本信息，然后判断模型是否存在，模型的单位是否相同。
		err := ds.checkMetricModelsUnit(ctx, metricModelIds)
		if err != nil {
			return resp, err
		}

		metricNum := len(kpiObjective.ComprehensiveMetricModels) + len(kpiObjective.AdditionalMetricModels)
		// results := make(map[string]map[string]interfaces.MetricModelData)
		results := make([]metricResult, 0)
		errCh := make(chan error, metricNum)
		defer close(errCh)

		taskCtx := &taskContext{
			ctx:     ctx,
			results: results,
			wg:      sync.WaitGroup{},
			errCh:   errCh,
		}

		// 并发查询
		for _, id := range metricModelIds {
			taskCtx.wg.Add(1)
			err := objectivePool.Submit(ds.getMetricModelData(taskCtx, query, id))
			if err != nil {
				detail := fmt.Sprintf("submit task of processing a document failed, %s", err.Error())
				return resp, rest.NewHTTPError(ctx, http.StatusInternalServerError,
					uerrors.Uniquery_ObjectiveModel_InternalError_SubmitGetMetricDataTaskFailed).WithErrorDetails(detail)
			}
		}

		// 读取数据
		taskCtx.wg.Wait()

		if len(taskCtx.errCh) > 0 {
			err := <-taskCtx.errCh
			return resp, err
		}

		metricMap := make(map[string]map[string]interfaces.MetricModelData)
		for _, res := range taskCtx.results {
			metricMap[res.metricModelID] = res.seriesMap
		}

		// 计算其他指标
		// 把查询结果处理成统一的格式。当前的预览是query_range查询，返回结果为 matrix
		return processKPIObjectiveResponse(ctx, metricMap, kpiObjective)
	default:
		// 异常，不支持的查询类型
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Unsupport objective type %s", query.ObjectiveType))

		return resp, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_ObjectiveModel_UnsupportQuery).
			WithErrorDetails(fmt.Sprintf("Unsupport objective type %s", query.ObjectiveType))
	}
}

func (ds *objectiveModelService) checkMetricModelsUnit(ctx context.Context, metricModelIds []string) error {

	metricModels, err := ds.mmAccess.GetMetricModels(ctx, metricModelIds)
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			uerrors.Uniquery_ObjectiveModel_InternalError_GetMetricModelsFailed).WithErrorDetails(err)
	}
	unit := ""
	for i, mmodel := range metricModels {
		if i == 0 {
			unit = mmodel.Unit
		}
		if unit != mmodel.Unit {
			errDetails := fmt.Sprintf("Exist some metric model's unit[%s] with whose id is [%s] and name is [%s] was not equal others[%s]!",
				mmodel.Unit, mmodel.ModelID, mmodel.ModelName, unit)
			logger.Error(errDetails)
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_ObjectiveModel_AssociateMetricModelsUnit_Different).
				WithErrorDetails(errDetails)
		}
	}
	return nil
}

type taskContext struct {
	// results map[string]map[string]interfaces.MetricModelData
	results []metricResult
	ctx     context.Context
	wg      sync.WaitGroup
	errCh   chan error
}

type metricResult struct {
	metricModelID string
	seriesMap     map[string]interfaces.MetricModelData
}

// 获取指标模型数据
func (ds *objectiveModelService) getMetricModelData(taskCtx *taskContext, query interfaces.ObjectiveModelQuery, id string) func() {
	return func() {
		defer taskCtx.wg.Done()

		ctx, span := ar_trace.Tracer.Start(taskCtx.ctx, "查询指标模型的指标数据")
		span.SetAttributes(attribute.Key("objective_model_id").String(query.ObjectiveModelID),
			attribute.Key("metric_model_id").String(id))
		span.End()

		// 2. 请求metric model 获取指标数据
		query.MetricModelQueryParameters.Limit = interfaces.DEFAULT_SERIES_LIMIT_INT
		uniResponse, _, _, err := ds.mmService.Exec(ctx, &interfaces.MetricModelQuery{
			QueryTimeParams:            query.QueryTimeParams,
			MetricModelQueryParameters: query.MetricModelQueryParameters,
			MetricModelID:              id,
		})
		if err != nil {
			span.SetStatus(codes.Error, fmt.Sprintf("Get metric data by id[%s] error", id))

			logger.Errorf("Get metric data error: %s", err.Error())
			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("Get metric data error: %v", err))

			taskCtx.errCh <- err
			return
		}
		seriesMap := transferMetricData2SeriesMap(uniResponse.Datas)

		taskCtx.results = append(taskCtx.results, metricResult{metricModelID: id, seriesMap: seriesMap})
		span.SetStatus(codes.Ok, "")
	}
}

// 计算SLO的指标
func processSLOObjectiveResponse(ctx context.Context, results map[string]map[string]interfaces.MetricModelData,
	sloObjective interfaces.SLOObjective) (interfaces.ObjectiveModelUniResponse, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "Parse promql result to uniresponse")
	defer span.End()

	resp := interfaces.ObjectiveModelUniResponse{Datas: []any{}}

	// 对ranges排序，这样的话对于状态值转换就顺序从1开始
	rgs := make([]interfaces.Range, 0)
	if sloObjective.StatusConfig != nil {
		rgs = sortRanges(*sloObjective.StatusConfig)
	}
	goodSeries := results[sloObjective.GoodMetricModel.Id]
	totalSeries := results[sloObjective.TotalMetricModel.Id]

	// good 和total的序列需相同，分别对good 和 total的序列转成map，完全匹配才能参与计算，不匹配的被过滤掉。

	// lables是个map，需要字段名(key)排序，然后用key和value组成字符串作为序列标识。用这个序列标识去作为good和total的匹配。
	// var sloDatas []interfaces.SLOObjectiveData
	timePoints := 0
	for seriesKey, goods := range goodSeries {
		// 拿到序列，去total中获取，挨个序列在每个时间点上做计算
		sloData := interfaces.SLOObjectiveData{
			Labels: goods.Labels,
			Times:  goods.Times,
		}
		totals, ok := totalSeries[seriesKey]
		if !ok {
			// 若不存在，则当前的序列不添加到最终结果中，遍历下一个
			continue
		}

		// 需要指标模型补点后的结果，如果时间数组不相同，报错
		if len(goods.Times) != len(totals.Times) {
			// 报错
			errorDetail := fmt.Sprintf("The number of time points [%d] for good metric does not equal the number of time points [%d] for total metric.",
				len(goods.Times), len(totals.Times))
			// 记录异常日志
			o11y.Error(ctx, errorDetail)
			span.SetStatus(codes.Error, errorDetail)

			return resp, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				uerrors.Uniquery_ObjectiveModel_InternalError_MetricTimePointsNotEqual).WithErrorDetails(errorDetail)
		}
		// 计算
		for i, good := range goods.Values {
			sloData.Objective = append(sloData.Objective, *sloObjective.Objective)
			sloData.Period = append(sloData.Period, *sloObjective.Period)
			// 特殊值
			var sli, achiveRate, totalErrorBudget, leftErrorBudget, burnRate any
			var status string

			// ranges排过序，这样的话对于状态值转换就顺序从1开始
			var statusCode int
			if good == nil || totals.Values[i] == nil {
				sli, achiveRate, totalErrorBudget, leftErrorBudget, burnRate = nil, nil, nil, nil, nil
				if sloObjective.StatusConfig == nil || sloObjective.StatusConfig.OtherStatus == "" {
					status = interfaces.STATUS_DEFAULT
					statusCode = interfaces.STATUS_CODE_DEFAULT
				} else {
					status = sloObjective.StatusConfig.OtherStatus
					statusCode = interfaces.STATUS_CODE_OTHER
				}
			} else if good == convert.POS_INF || good == convert.NEG_INF || good == convert.NaN {
				sli, achiveRate, totalErrorBudget, leftErrorBudget, burnRate = good, good, good, good, good
				if sloObjective.StatusConfig == nil || sloObjective.StatusConfig.OtherStatus == "" {
					status = interfaces.STATUS_DEFAULT
					statusCode = interfaces.STATUS_CODE_DEFAULT
				} else {
					status = sloObjective.StatusConfig.OtherStatus
					statusCode = interfaces.STATUS_CODE_OTHER
				}
			} else if totals.Values[i] == convert.POS_INF || totals.Values[i] == convert.NEG_INF || totals.Values[i] == convert.NaN {
				sli, achiveRate, totalErrorBudget, leftErrorBudget, burnRate = totals.Values[i], totals.Values[i], totals.Values[i], totals.Values[i], totals.Values[i]
				if sloObjective.StatusConfig == nil || sloObjective.StatusConfig.OtherStatus == "" {
					status = interfaces.STATUS_DEFAULT
					statusCode = interfaces.STATUS_CODE_DEFAULT
				} else {
					status = sloObjective.StatusConfig.OtherStatus
					statusCode = interfaces.STATUS_CODE_OTHER
				}
			} else {
				goodV, err := convert.AssertFloat64(good)
				if err != nil {
					return resp, rest.NewHTTPError(ctx, http.StatusInternalServerError,
						uerrors.Uniquery_InternalError_AssertFloat64Failed).
						WithErrorDetails(fmt.Sprintf("err: %v, good metric value is %v", err, good))
				}
				totalV, err := convert.AssertFloat64(totals.Values[i])
				if err != nil {
					return resp, rest.NewHTTPError(ctx, http.StatusInternalServerError,
						uerrors.Uniquery_InternalError_AssertFloat64Failed).
						WithErrorDetails(fmt.Sprintf("err: %v, total metric value is %v", err, totals.Values[i]))
				}
				sliV := goodV / totalV * 100
				sli = convert.WrapMetricValue(sliV)
				achiveRate = convert.WrapMetricValue(min(100, 100*sliV/(*sloObjective.Objective)))
				totalErrorBudget = convert.WrapMetricValue((100 - *sloObjective.Objective) / 100 * totalV)
				leftErrorBudget = convert.WrapMetricValue((sliV - (*sloObjective.Objective)) / 100 * totalV)
				if *sloObjective.Objective == interfaces.SLO_MAX_OBJECTIVE {
					// 当slo的目标定位100时，sli 小于100，燃烧率位为 100%； 当 sli 等于100的时候，燃烧率为0.
					if sliV == *sloObjective.Objective {
						burnRate = 100
					} else {
						burnRate = 0
					}
				} else {
					burnRate = convert.WrapMetricValue((100 - sliV) / (100 - *sloObjective.Objective) * 100)
				}

				// 状态转换
				if sloObjective.StatusConfig == nil {
					status = interfaces.STATUS_DEFAULT
					statusCode = interfaces.STATUS_CODE_DEFAULT
				} else {
					index, rg := convert.FindRange(sliV, rgs)
					if rg == nil {
						status = interfaces.STATUS_DEFAULT
						statusCode = interfaces.STATUS_CODE_DEFAULT
					} else {
						status = rg.Status
						statusCode = index + 1
					}
				}
			}

			sloData.SLI = append(sloData.SLI, sli)
			sloData.AchiveRate = append(sloData.AchiveRate, achiveRate)
			sloData.TotalErrorBudget = append(sloData.TotalErrorBudget, totalErrorBudget)
			sloData.LeftErrorBudget = append(sloData.LeftErrorBudget, leftErrorBudget)
			sloData.BurnRate = append(sloData.BurnRate, burnRate)
			sloData.Status = append(sloData.Status, status)
			sloData.StatusCode = append(sloData.StatusCode, statusCode)
			sloData.Good = goods.Values
			sloData.Total = totals.Values
		}
		timePoints = len(sloData.Times)
		resp.Datas = append(resp.Datas, sloData)
	}

	resp.SeriesTotal = len(resp.Datas)
	if len(resp.Datas) == 0 {
		resp.PointTotalPerSeries = 0
	} else {
		resp.PointTotalPerSeries = timePoints
	}

	span.SetStatus(codes.Ok, "")
	return resp, nil
}

// 计算kpi的指标
func processKPIObjectiveResponse(ctx context.Context, results map[string]map[string]interfaces.MetricModelData,
	kpiObjective interfaces.KPIObjective) (interfaces.ObjectiveModelUniResponse, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "Parse promql result to uniresponse")
	defer span.End()

	resp := interfaces.ObjectiveModelUniResponse{Datas: []any{}}

	// 对ranges排序，这样的话对于状态值转换就顺序从1开始
	rgs := make([]interfaces.Range, 0)
	if kpiObjective.StatusConfig != nil {
		rgs = sortRanges(*kpiObjective.StatusConfig)
	}

	firstMetricSeries := results[kpiObjective.ComprehensiveMetricModels[0].Id]
	// good 和total的序列需相同，分别对good 和 total的序列转成map，完全匹配才能参与计算，不匹配的被过滤掉。
	// lables是个map，需要字段名(key)排序，然后用key和value组成字符串作为序列标识。用这个序列标识去作为good和total的匹配。
	// var kpiDatas []interfaces.KPIObjectiveData
	timePoints := 0
	for seriesKey, first := range firstMetricSeries {
		// 拿到序列，去综合计算指标集中获取，挨个序列在每个时间点上做加权求和
		remainComMetrics := make([]interfaces.MetricModelData, 0)
		timePointsNum := len(first.Times)
		for j := 1; j < len(kpiObjective.ComprehensiveMetricModels); j++ {
			metrici, ok := results[kpiObjective.ComprehensiveMetricModels[j].Id]
			if !ok {
				// 若某个关联模型的数据无，直接返回空
				return resp, nil
			}
			mData, ok := metrici[seriesKey]
			if !ok {
				// 序列在其他模型中不存在，不参与计算，结束当前序列的遍历，遍历下一个序列
				break
			}
			if timePointsNum != len(mData.Times) {
				// 报错
				errorDetail := fmt.Sprintf("The number of time points [%d] for first comprehensive metric does not equal the number of time points [%d] for comprehensive metrics[%d].",
					len(first.Times), len(mData.Times), j)
				// 记录异常日志
				o11y.Error(ctx, errorDetail)
				span.SetStatus(codes.Error, errorDetail)

				return resp, rest.NewHTTPError(ctx, http.StatusInternalServerError,
					uerrors.Uniquery_ObjectiveModel_InternalError_MetricTimePointsNotEqual).WithErrorDetails(errorDetail)
			}

			remainComMetrics = append(remainComMetrics, mData)
		}

		additionalMetrics := make([]interfaces.MetricModelData, 0)
		for j, additional := range kpiObjective.AdditionalMetricModels {
			metrici, ok := results[additional.Id]
			if !ok {
				// 若某个关联模型的数据无，直接返回空
				return resp, nil
			}
			mData, ok := metrici[seriesKey]
			if !ok {
				// 序列在其他模型中不存在，不参与计算，结束当前序列的遍历，遍历下一个序列
				break
			}
			if timePointsNum != len(mData.Times) {
				// 报错
				errorDetail := fmt.Sprintf("The number of time points [%d] for first comprehensive metric does not equal the number of time points [%d] for additional metrics[%d].",
					len(first.Times), len(mData.Times), j)
				// 记录异常日志
				o11y.Error(ctx, errorDetail)
				span.SetStatus(codes.Error, errorDetail)

				return resp, rest.NewHTTPError(ctx, http.StatusInternalServerError,
					uerrors.Uniquery_ObjectiveModel_InternalError_MetricTimePointsNotEqual).WithErrorDetails(errorDetail)
			}

			additionalMetrics = append(additionalMetrics, mData)
		}

		if len(remainComMetrics) != len(kpiObjective.ComprehensiveMetricModels)-1 ||
			len(additionalMetrics) != len(kpiObjective.AdditionalMetricModels) {
			// 当前序列在剩余的指标中不是都能找到序列，则遍历下一个序列
			continue
		}

		kpiData := interfaces.KPIObjectiveData{
			Labels: first.Labels,
			Times:  first.Times,
		}
		// 需要指标模型补点后的结果，如果时间数组不相同，报错

		// 计算
		for i, firstMetricV := range first.Values {
			kpiData.Objective = append(kpiData.Objective, *kpiObjective.Objective)
			// 特殊值
			var kpi, achiveRate, kpiScore, associateMetricNum any
			var status string
			// ranges排过序，这样的话对于状态值转换就顺序从1开始
			var statusCode int

			// 判断其他指标模型在当前时间点是不是空
			othersIsNull := false
			for _, metric := range remainComMetrics {
				if metric.Values[i] == nil || metric.Values[i] == convert.POS_INF ||
					metric.Values[i] == convert.NEG_INF || metric.Values[i] == convert.NaN {
					othersIsNull = true
				}
			}
			for _, metric := range additionalMetrics {
				if metric.Values[i] == nil || metric.Values[i] == convert.POS_INF ||
					metric.Values[i] == convert.NEG_INF || metric.Values[i] == convert.NaN {
					othersIsNull = true
				}
			}

			if firstMetricV == nil || othersIsNull {
				kpi, achiveRate, kpiScore, associateMetricNum = nil, nil, nil, nil
				if kpiObjective.StatusConfig == nil || kpiObjective.StatusConfig.OtherStatus == "" {
					status = interfaces.STATUS_DEFAULT
					statusCode = interfaces.STATUS_CODE_DEFAULT
				} else {
					status = kpiObjective.StatusConfig.OtherStatus
					statusCode = interfaces.STATUS_CODE_OTHER
				}
			} else if firstMetricV == convert.POS_INF || firstMetricV == convert.NEG_INF || firstMetricV == convert.NaN {
				kpi, achiveRate, kpiScore, associateMetricNum = firstMetricV, firstMetricV, firstMetricV, firstMetricV
				if kpiObjective.StatusConfig == nil || kpiObjective.StatusConfig.OtherStatus == "" {
					status = interfaces.STATUS_DEFAULT
					statusCode = interfaces.STATUS_CODE_DEFAULT
				} else {
					status = kpiObjective.StatusConfig.OtherStatus
					statusCode = interfaces.STATUS_CODE_OTHER
				}
			} else {
				// 计算。读取后续综合指标的值来做加权求和
				firstV, err := convert.AssertFloat64(firstMetricV)
				if err != nil {
					return resp, rest.NewHTTPError(ctx, http.StatusInternalServerError,
						uerrors.Uniquery_InternalError_AssertFloat64Failed).
						WithErrorDetails(fmt.Sprintf("err: %v, metric value is %v", err, firstMetricV))
				}
				kpiI := firstV * float64((*kpiObjective.ComprehensiveMetricModels[0].Weight))
				for m, comMetric := range remainComMetrics {
					comMetricV, err := convert.AssertFloat64(comMetric.Values[i])
					if err != nil {
						return resp, rest.NewHTTPError(ctx, http.StatusInternalServerError,
							uerrors.Uniquery_InternalError_AssertFloat64Failed).
							WithErrorDetails(fmt.Sprintf("err: %v, metric value is %v", err, comMetric.Values[i]))
					}
					kpiI += comMetricV * float64((*kpiObjective.ComprehensiveMetricModels[m+1].Weight))
				}
				kpiI /= 100
				// 附加计算
				for _, additionMetric := range additionalMetrics {
					additionMetricV, err := convert.AssertFloat64(additionMetric.Values[i])
					if err != nil {
						return resp, rest.NewHTTPError(ctx, http.StatusInternalServerError,
							uerrors.Uniquery_InternalError_AssertFloat64Failed).
							WithErrorDetails(fmt.Sprintf("err: %v, metric value is %v", err, additionMetric.Values[i]))
					}
					kpiI += additionMetricV
				}

				kpi = convert.WrapMetricValue(kpiI)
				achiveRateV := kpiI / (*kpiObjective.Objective) * 100
				achiveRate = convert.WrapMetricValue(achiveRateV)
				kpiScoreV := achiveRateV
				if kpiObjective.ScoreMax != nil {
					kpiScoreV = min(kpiScoreV, *kpiObjective.ScoreMax)
				}
				if kpiObjective.ScoreMin != nil {
					kpiScoreV = max(kpiScoreV, *kpiObjective.ScoreMin)
				}
				kpiScore = convert.WrapMetricValue(kpiScoreV)

				associateMetricNum = len(kpiObjective.ComprehensiveMetricModels) + len(kpiObjective.AdditionalMetricModels)
				// 状态转换
				if kpiObjective.StatusConfig == nil {
					status = interfaces.STATUS_DEFAULT
					statusCode = interfaces.STATUS_CODE_DEFAULT
				} else {
					index, rg := convert.FindRange(kpiI, rgs)
					if rg == nil {
						status = interfaces.STATUS_DEFAULT
						statusCode = interfaces.STATUS_CODE_DEFAULT
					} else {
						status = rg.Status
						statusCode = index + 1
					}
				}
			}
			kpiData.KPI = append(kpiData.KPI, kpi)
			kpiData.AchiveRate = append(kpiData.AchiveRate, achiveRate)
			kpiData.KPIScore = append(kpiData.KPIScore, kpiScore)
			kpiData.AssociateMetricNums = append(kpiData.AssociateMetricNums, associateMetricNum)
			kpiData.Status = append(kpiData.Status, status)
			kpiData.StatusCode = append(kpiData.StatusCode, statusCode)
		}
		timePoints = len(kpiData.Times)
		resp.Datas = append(resp.Datas, kpiData)
	}

	// resp.Datas = kpiDatas
	resp.SeriesTotal = len(resp.Datas)
	if len(resp.Datas) == 0 {
		resp.PointTotalPerSeries = 0
	} else {
		resp.PointTotalPerSeries = timePoints
	}
	span.SetStatus(codes.Ok, "")
	return resp, nil

}

func sortRanges(statusConfig interfaces.ObjectiveStatusConfig) []interfaces.Range {
	rgs := make([]interfaces.Range, 0)
	for _, rg := range statusConfig.Ranges {
		if rg.From == nil {
			rg.From = &interfaces.NEG_INF
		}
		if rg.To == nil {
			rg.To = &interfaces.POS_INF
		}
		rgs = append(rgs, rg)
	}

	sort.Slice(rgs, func(i, j int) bool {
		if rgs[i].From == rgs[j].From {
			return *rgs[i].To < *rgs[j].To
		}
		return *rgs[i].From < *rgs[j].From
	})

	return rgs
}

func transferMetricData2SeriesMap(datas []interfaces.MetricModelData) map[string]interfaces.MetricModelData {
	seriesMap := make(map[string]interfaces.MetricModelData)
	for _, series := range datas {
		lNames := make([]string, 0)
		for lName := range series.Labels {
			// k是维度字段名，v是维度字段值
			lNames = append(lNames, lName)
		}
		sort.Strings(lNames)

		labelsSig := ""
		for _, lName := range lNames {
			labelsSig += lName + "=" + series.Labels[lName] + ","
		}
		seriesMap[labelsSig] = series
	}
	return seriesMap
}
