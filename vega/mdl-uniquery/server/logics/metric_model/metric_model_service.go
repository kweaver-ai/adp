package metric_model

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/tidwall/gjson"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"golang.org/x/sync/semaphore"

	"uniquery/common"
	cond "uniquery/common/condition"
	"uniquery/common/convert"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	"uniquery/interfaces/data_type"
	"uniquery/logics"
	"uniquery/logics/data_view"
	"uniquery/logics/permission"
	"uniquery/logics/promql"
	"uniquery/logics/promql/static"
	"uniquery/logics/promql/util"
	"uniquery/logics/vega_view"
)

var (
	mmServiceOnce sync.Once
	mmService     interfaces.MetricModelService
)

type metricModelService struct {
	sem           *semaphore.Weighted // sem 是信号量，控制发送opensearch的并发数
	appSetting    *common.AppSetting
	ibAccess      interfaces.IndexBaseAccess
	mmAccess      interfaces.MetricModelAccess
	osAccess      interfaces.OpenSearchAccess
	staticAccess  interfaces.StaticAccess
	dvService     interfaces.DataViewService
	ps            interfaces.PermissionService
	promqlService interfaces.PromQLService
	vvs           interfaces.VegaService
}

func NewMetricModelService(appSetting *common.AppSetting) interfaces.MetricModelService {
	mmServiceOnce.Do(func() {
		mmService = &metricModelService{
			appSetting:   appSetting,
			dvService:    data_view.NewDataViewService(appSetting),
			ibAccess:     logics.IBAccess,
			mmAccess:     logics.MMAccess,
			osAccess:     logics.OSAccess,
			ps:           permission.NewPermissionService(appSetting),
			sem:          semaphore.NewWeighted(int64(appSetting.PoolSetting.ExecutePoolSize)),
			staticAccess: logics.StAccess,
			vvs:          vega_view.NewVegaService(appSetting),
		}

		// 给promqlService注入mmService的依赖
		promqlService := promql.NewPromQLService(appSetting, mmService)
		mmService.(*metricModelService).promqlService = promqlService
		// 利用mms的dbaccess从数据库中读取静态表的数据，并把静态表数据写入到interfaces的全局变量中。只做一次
		mmService.GetIndexBaseSplitTime()

	})
	return mmService
}

func (mms *metricModelService) GetIndexBaseSplitTime() {
	indexBaseSplitTimes, err := mms.staticAccess.GetIndexBaseSplitTime()
	if err != nil {
		// panic
		logger.Fatalf("GetIndexBaseSplitTime error: %s", err.Error())
		panic(err)
	}
	// 把数据更新到全局变量中
	for _, v := range indexBaseSplitTimes {
		//切到数据视图时，这个值可直接用basetype
		interfaces.INDEX_BASE_SPLIT_TIME[v.BaseType] = v.SplitTime
		interfaces.INDEX_BASE_SPLIT_TIME[v.BaseType] = v.SplitTime
		// promql的query和query_range接口走的是日志分组，用索引模式来判定
		interfaces.INDEX_PATTERN_SPLIT_TIME[v.BaseType+"-*"] = v.SplitTime
		interfaces.INDEX_PATTERN_SPLIT_TIME["mdl-"+v.BaseType+"-*"] = v.SplitTime
	}
}

// 基于指标模型的数据预览或计算公式有效性检查
func (mms *metricModelService) Simulate(ctx context.Context,
	query interfaces.MetricModelQuery) (interfaces.MetricModelUniResponse, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "查询需预览的指标数据")
	defer span.End()

	// 决策权限。 预览的时候还没有模型id，此时的预览校验用新建或者编辑，
	ops, err := mms.ps.GetResourcesOperations(ctx, interfaces.RESOURCE_TYPE_METRIC_MODEL,
		[]string{interfaces.RESOURCE_ID_ALL})
	if err != nil {
		return interfaces.MetricModelUniResponse{}, err
	}
	if len(ops) != 1 {
		// 无权限
		return interfaces.MetricModelUniResponse{}, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
			WithErrorDetails("Access denied: insufficient permissions for metric model's create or modify operation.")
	}
	// 从 ops 里找新建或编辑的权限
	for _, op := range ops[0].Operations {
		if op != interfaces.OPERATION_TYPE_CREATE && op != interfaces.OPERATION_TYPE_MODIFY {
			return interfaces.MetricModelUniResponse{}, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
				WithErrorDetails("Access denied: insufficient permissions for metric model's create or modify operation.")
		}
	}

	// span.SetAttributes(attribute.Key("data_source_type").String(query.DataSource.Type),
	// 	attribute.Key("data_source_id").String(query.DataSource.ID))
	dataView := &interfaces.DataView{}
	ayDims := []interfaces.Field{}
	if query.MetricType == interfaces.ATOMIC_METRIC {
		dataView, err = mms.dvService.GetDataViewByID(ctx, query.DataSource.ID, true)
		if err != nil {
			span.SetStatus(codes.Error, "Get data view by ID failed")
			return interfaces.MetricModelUniResponse{}, err
		}

		if dataView.QueryType == interfaces.QueryType_SQL {
			// 2.预览还需校验请求参数
			// 2. 校验时间字段是否属于逻辑视图
			if _, exist := dataView.FieldsMap[query.DateField]; query.DateField != "" && !exist {
				return interfaces.MetricModelUniResponse{}, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_DateFieldNotExisted).
					WithErrorDetails(fmt.Sprintf("date field[%s] not exists in vega logic view[%s]", query.DateField, query.DataSource.ID))
			}

			// 3. 校验分组字段是否属于逻辑视图
			sqlConfig := query.FormulaConfig.(interfaces.SQLConfig)
			for _, field := range sqlConfig.GroupByFields {
				if _, exist := dataView.FieldsMap[field]; !exist {
					return interfaces.MetricModelUniResponse{}, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_GroupByFieldNotExisted).
						WithErrorDetails(fmt.Sprintf("group by field[%s] not exists in vega logic view[%s]", field, query.DataSource.ID))
				}
			}

			// 4. 校验分析维度是否属于逻辑视图
			for _, field := range query.AnalysisDims {
				if _, exist := dataView.FieldsMap[field]; !exist {
					return interfaces.MetricModelUniResponse{}, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_AnalysisDimensionNotExisted).
						WithErrorDetails(fmt.Sprintf("analysis dimension[%s] not exists in vega logic view[%s]", field, query.DataSource.ID))
				}
				ayDims = append(ayDims, interfaces.Field{
					Name:        field,
					Type:        dataView.FieldsMap[field].Type,
					DisplayName: dataView.FieldsMap[field].DisplayName,
					Comment:     &dataView.FieldsMap[field].Comment,
				})
			}

			// 5. 校验聚合字段是否属于逻辑视图
			if sqlConfig.AggrExpr != nil {
				if _, exist := dataView.FieldsMap[sqlConfig.AggrExpr.Field]; !exist {
					return interfaces.MetricModelUniResponse{}, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_AggregationNotExisted).
						WithErrorDetails(fmt.Sprintf("aggregation field[%s] not exists in vega logic view[%s]", sqlConfig.AggrExpr.Field, query.DataSource.ID))
				}
			}
		}
		query.DataSource.Type = dataView.QueryType
	}

	resp, err := mms.eval(ctx, query, interfaces.MetricModel{}, dataView)
	if err != nil {
		// 添加异常时的 trace 属性
		span.SetStatus(codes.Error, "Eval Metric Model error")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Eval Metric Model error: %v", err))
		return resp, err
	}

	if query.IncludeModel {
		resp.Model = interfaces.MetricModel{
			MetricType:      query.MetricType,
			DataSource:      query.DataSource,
			QueryType:       query.QueryType,
			Formula:         query.Formula,
			FormulaConfig:   query.FormulaConfig,
			AnalysisDims:    ayDims,
			OrderByFields:   query.OrderByFields,
			HavingCondition: query.HavingCondition,
			DateField:       query.DateField,
			MeasureField:    query.MeasureField,
		}
	}

	span.SetStatus(codes.Ok, "")
	return resp, nil
}

func (mms *metricModelService) GetMetricModelIDByName(ctx context.Context, groupName, modelName string) (string, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("根据指标模型分组名称[%s]、模型名称[%s]查询模型ID", groupName, modelName))
	span.SetAttributes(attribute.Key("group_name").String(groupName),
		attribute.Key("model_name").String(modelName))
	defer span.End()

	modelID, exist, err := mms.mmAccess.GetMetricModelIDByName(ctx, groupName, modelName)

	if err != nil {
		logger.Errorf("GetMetricModelIDByName error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("Get metric model id by group name [%s] model name [%s] error: %v", groupName, modelName, err))
		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
			uerrors.Uniquery_MetricModel_InternalError_GetModelIdByNameFailed).WithErrorDetails(err.Error())
	}
	if !exist {
		logger.Debugf("Metric Model %s not found!", modelName)
		span.SetStatus(codes.Error, fmt.Sprintf("Metric model [%s] not found", modelName))
		return "", rest.NewHTTPError(ctx, http.StatusNotFound, uerrors.Uniquery_MetricModel_MetricModelNotFound)
	}
	span.SetStatus(codes.Ok, "")

	return modelID, nil
}

// 基于指标模型的指标数据查询
func (mms *metricModelService) Exec(ctx context.Context, query *interfaces.MetricModelQuery) (interfaces.MetricModelUniResponse, int, int, error) {

	// 获取指标模型信息
	ctx, span := ar_trace.Tracer.Start(ctx, "查询指标模型的指标数据")
	var resps interfaces.MetricModelUniResponse
	seriesTotal := 0
	pointTotal := 0

	// 决策当前模型id的数据查询权限
	err := mms.ps.CheckPermission(ctx, interfaces.Resource{
		ID:   query.MetricModelID,
		Type: interfaces.RESOURCE_TYPE_METRIC_MODEL,
	}, []string{interfaces.OPERATION_TYPE_DATA_QUERY})
	if err != nil {
		return resps, seriesTotal, pointTotal, err
	}

	metricModels, exists, err := mms.mmAccess.GetMetricModel(ctx, query.MetricModelID)
	if err != nil {
		logger.Errorf("Get Metric Model error: %s", err.Error())

		// 添加异常时的 trace 属性
		span.SetAttributes(attribute.Key("model_id").String(query.MetricModelID))
		span.SetStatus(codes.Error, "Get Metric Model error")
		span.End()
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Get Metric Model error: %v", err))

		return resps, seriesTotal, pointTotal, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			uerrors.Uniquery_MetricModel_InternalError_GetModelByIdFailed).WithErrorDetails(err.Error())
	}
	if !exists || len(metricModels) == 0 {
		logger.Debugf("Metric Model %d not found!", query.MetricModelID)

		// 添加异常时的 trace 属性
		span.SetAttributes(attribute.Key("model_id").String(query.MetricModelID))
		span.SetStatus(codes.Error, "Metric Model not found!")
		span.End()
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Metric Model [%s] not found!", query.MetricModelID))

		return resps, seriesTotal, pointTotal, rest.NewHTTPError(ctx, http.StatusNotFound, uerrors.Uniquery_MetricModel_MetricModelNotFound)
	}

	metricModel := metricModels[0]

	query.Formula = metricModel.Formula
	query.FormulaConfig = metricModel.FormulaConfig
	query.MetricType = metricModel.MetricType
	query.QueryType = metricModel.QueryType
	query.MeasureField = metricModel.MeasureField
	query.DataSource = metricModel.DataSource
	query.MetricModelID = metricModel.ModelID
	query.DateField = metricModel.DateField // 按id查询时,需要把模型配置的时间字段赋值给query的时间字段

	dataView := &interfaces.DataView{}
	if query.MetricType == interfaces.ATOMIC_METRIC {
		// query.DataView = metricModel.DataView
		// 原子指标需要再请求视图获取视图详情
		dataView, err = mms.dvService.GetDataViewByID(ctx, metricModel.DataSource.ID, true)
		if err != nil {
			span.SetStatus(codes.Error, "Get data view by ID failed")
			return resps, seriesTotal, pointTotal, err
		}
	}

	// vega数据源，需请求vega视图id的字段列表
	if metricModel.DataSource != nil && metricModel.DataSource.Type == interfaces.QueryType_SQL {
		// 模型配置的分析维度转成map
		modelAnaysisDims := make(map[string]interfaces.Field)
		for _, dim := range metricModel.AnalysisDims {
			modelAnaysisDims[dim.Name] = dim
		}

		// 按id请求时, 检查请求的下钻维度是否属于模型配置的分析维度, 预览时无此要求,预览时只提交一个分析维度,模型上的分析维度不用提交
		// 数据查询的分析维度需要属于模型的分析维度
		for _, field := range query.AnalysisDims {
			if _, exist := modelAnaysisDims[field]; !exist {
				return resps, seriesTotal, pointTotal, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_AnalysisDimensionNotExisted).
					WithErrorDetails(fmt.Sprintf("analysis dimension[%s] is not belongs to model [%s], expected analysis dimensions should belongs to %v",
						field, query.DataSource.ID, metricModel.AnalysisDims))
			}
		}

	}

	// 请求参数匹配持久化任务。匹配任务的时间窗口+过滤条件
	// 从定时任务来的，不走任务匹配。
	if !query.IgnoringStoreCache {
		dataView, err = mms.macthPersist(ctx, query, metricModel, dataView)
		if err != nil {
			return resps, seriesTotal, pointTotal, err
		}
	}

	// 按 指标类型+查询语言 分情况发起查询
	respi, err := mms.eval(ctx, *query, metricModel, dataView)
	if err != nil {
		// 添加异常时的 trace 属性
		span.SetAttributes(attribute.Key("model_id").String(query.MetricModelID))
		span.SetStatus(codes.Error, "Eval Metric Model error")
		span.End()
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Eval Metric Model error: %v", err))

		return resps, seriesTotal, pointTotal, err
	}

	if query.IncludeModel {
		respi.Model = metricModel
	}

	// ok
	span.SetStatus(codes.Ok, "")
	span.End()
	respi.StatusCode = http.StatusOK
	respi.HasMatchPersist = query.HasMatchPersist

	return respi, respi.CurrSeriesNum, respi.PointTotal, nil
}

// 按指标模型的信息查询指标数据
func (mms *metricModelService) eval(ctx context.Context, query interfaces.MetricModelQuery,
	metricModel interfaces.MetricModel, dataView *interfaces.DataView) (interfaces.MetricModelUniResponse, error) {
	resp := interfaces.MetricModelUniResponse{
		Datas:      []interfaces.MetricModelData{},
		Step:       query.StepStr,
		IsVariable: query.IsVariable,
		IsCalendar: query.IsCalendar,
	}

	// 预览时，update_time为空，给当前时间，不会缓存序列，直接查。
	query.ModelUpdateTime = metricModel.UpdateTime
	if query.ModelUpdateTime == 0 {
		query.ModelUpdateTime = time.Now().UnixMilli()
	}

	// 视图更新时间
	if dataView.UpdateTime == 0 {
		dataView.UpdateTime = time.Now().UnixMilli()
	}

	// 按 metric_type 和 promql_type 做相应的请求
	switch query.MetricType {
	case interfaces.ATOMIC_METRIC:
		// 通过视图和时间范围，获取视图查询的索引列表和查询子句。
		viewQuery, err := mms.dvService.BuildViewQuery4MetricModel(ctx, *query.Start, *query.End, dataView)
		if err != nil {
			return resp, err
		}
		// dsl 视图 || 索引库查询 && indices 为空，则返回空结果
		if (dataView.QueryType == interfaces.QueryType_DSL || dataView.QueryType == interfaces.QueryType_IndexBase) && len(viewQuery.Indices) == 0 {
			return resp, nil
		}
		query.ViewQuery4Metric = viewQuery

		switch qt, dsType := query.QueryType, query.DataSource.Type; {
		case qt == interfaces.PROMQL:
			// 执行promql
			return mms.execPromQL(ctx, query, dataView, metricModel)

		case qt == interfaces.DSL_CONFIG:
			// 把 formula_config 转成 dslconfig
			var dslConfig interfaces.MetricModelFormulaConfig
			jsonData, err := json.Marshal(query.FormulaConfig)
			if err != nil {
				return resp, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_FormulaConfig).
					WithErrorDetails(fmt.Sprintf("SQL Config Marshal error: %s", err.Error()))
			}
			err = json.Unmarshal(jsonData, &dslConfig)
			if err != nil {
				return resp, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_FormulaConfig).
					WithErrorDetails(fmt.Sprintf("SQL Config Unmarshal error: %s", err.Error()))
			}
			query.FormulaConfig = dslConfig

			dsl, err := generateDslByConfig(ctx, &query, dataView)
			if err != nil {
				return resp, rest.NewHTTPError(ctx, http.StatusBadRequest,
					uerrors.Uniquery_MetricModel_InvalidParameter_FormulaConfig).WithErrorDetails(err.Error())
			}
			str, _ := sonic.MarshalString(dsl)
			logger.Debugf("dsl config生成的dsl语句: %s", str)

			// 3. 向 opensearch 发起基于数据视图 id 的 dsl 查询。
			submitDslCtx, submitDslSpan := ar_trace.Tracer.Start(ctx, "请求 opensearch 查询数据")
			submitDslSpan.SetAttributes(attribute.Key("dsl_expression").String(fmt.Sprintf("%s", dsl)),
				attribute.Key("index").StringSlice(viewQuery.Indices))

			var scroll time.Duration
			res, _, err := mms.dvService.GetDataFromOpenSearch(submitDslCtx, dsl, viewQuery.Indices, scroll,
				interfaces.DEFAULT_PREFERENCE, query.TraceTotalHits)

			// 记录请求信息和返回信息
			o11y.Info(submitDslCtx, fmt.Sprintf("Opensearch search index: [%s], dsl: [%s]; search response err is [%s]",
				viewQuery.Indices, dsl, err))

			if err != nil {
				// 记录异常日志
				o11y.Error(submitDslCtx, fmt.Sprintf("Opensearch search error: %v", err))

				submitDslSpan.SetStatus(codes.Error, "Opensearch search error")
				submitDslSpan.End()
				return resp, rest.NewHTTPError(submitDslCtx, http.StatusInternalServerError, uerrors.Uniquery_InternalError).
					WithErrorDetails(err.Error())
			}
			submitDslSpan.SetStatus(codes.Ok, "")
			submitDslSpan.End()
			// 4. 把查询结果处理成统一的格式。当前的预览是query_range查询，返回结果为 matrix
			return parseDSLResult2UniresponseForDslConfig(ctx, res, metricModel, query)

		case qt == interfaces.DSL:
			// dsl
			if query.ContainTopHits {
				mType := dataView.FieldsMap[query.MeasureField].Type
				if mType != "long" && mType != "integer" && mType != "short" && mType != "byte" && mType != "double" &&
					mType != "float" && mType != "half_float" && mType != "scaled_float" && mType != "unsigned_long" {

					return resp, rest.NewHTTPError(ctx, http.StatusBadRequest,
						uerrors.Uniquery_MetricModel_InvalidParameter_MeasureField).
						WithErrorDetails("The type of measure field must be number.")
				}
			}

			// 1. 需先解析 dsl 语句，然后把聚合信息留下，用于读取dsl的返回结果，转成统一格式。
			parseDslCtx, parseDslSpan := ar_trace.Tracer.Start(ctx, "解析 dsl 表达式")
			// parseDslSpan.SetAttributes(attribute.Key("dsl_expression").String(fmt.Sprintf("%v", query.Formula)))
			dslInfo, err := parseDsl(parseDslCtx, &query, dataView)
			if err != nil {
				parseDslSpan.SetStatus(codes.Error, "Parse DSL error")
				parseDslSpan.End()
				return resp, err
			}

			// 按请求处理date_histogram的配置信息;当是 range 查询时,改写rangedsl的aggs中的date_histogram的值
			delta, err := processDateHistogram(ctx, &query, dslInfo, mms.appSetting.PromqlSetting.LookbackDelta)
			if err != nil {
				return resp, err
			}
			if query.IsInstantQuery {
				// 即时查询时才需要改变查询的时间范围
				start := *query.End - delta
				query.Start = &start
			}
			// 预估数据点数
			query.QueryTimeNum = evaluateTimeNum(query, dslInfo)

			// 记录解析到 dsl 语句
			// o11y.Info(parseDslCtx, fmt.Sprintf("Parse dsl success, dsl: %s, aggsInfos: %v", dsl, aggInfos))

			parseDslSpan.SetStatus(codes.Ok, "")
			parseDslSpan.End()

			// 如果是模型来的请求、或者查询参数主动设置直接查询时，忽略高基时，就直接查询，触碰到高基就报错。
			if query.IsModelRequest || query.IgnoringHCTS {
				return mms.getDSLDataDirectly(ctx, query, viewQuery.Indices, dslInfo, []interfaces.Filter{}, metricModel, dataView)
			}

			// 如果是来自模型的计算公式有效性检查,则都给size 1 进行直接查询
			if dslInfo.BucketSeriesNum*query.QueryTimeNum > interfaces.DEFAULT_MAX_QUERY_POINTS ||
				dslInfo.BucketSeriesNum > interfaces.DEFAULT_SERIES_NUM {

				// 分批查。一个terms和多个term的分页查不太一样。一个terms无需获取基数，就一直遍历获取即可。
				// 4.拼接查询序列的dsl。获取模型序列
				series, err := mms.getSeries(ctx, query, &dslInfo, dataView)
				if err != nil {
					return resp, err
				}
				if len(series) == 0 {
					return resp, nil
				}

				// 获取查询目标的索引分片数总和
				shardsTotal := 0
				for _, shard := range viewQuery.IndexShards {
					shardsTotal += shard.ShardNum
				}

				// 5.获取序列数据。对series分批获取
				return mms.getSeriesData(ctx, query, dslInfo, metricModel, series, viewQuery.Indices, shardsTotal, dataView)
			}
			// 否则直接查
			return mms.getDSLDataDirectly(ctx, query, viewQuery.Indices, dslInfo, []interfaces.Filter{}, metricModel, dataView)

		case qt == interfaces.SQL && dsType == interfaces.QueryType_SQL:
			// 组装sql,预览时加上limit 1,只查1条数据即可
			// 预览时 model 为nil
			if !query.IsInstantQuery {
				// 如果是sql 范围查询,步长只支持日历步长(拿到模型信息之后才知道模型的类型,查询方式和数据来源类型,所以校验放在这里)
				_, calendarExists := interfaces.CALENDAR_INTERVALS[*query.StepStr]
				// 请求步长不在日历步长集中，不合法
				if !calendarExists {
					// sql指标模型只支持日历步长，不支持固定步长
					return resp, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Step).
						WithErrorDetails(fmt.Sprintf("sql atomic metric only support calendar steps, expect steps is one of {%v}, actaul is %s",
							[]string{"minute", "hour", "day", "week", "month", "quarter", "year"}, *query.StepStr))
				}
			}

			query.IsCalendar = true // sql先只支持日历步长

			sql, err := generateSQL(ctx, query, dataView, viewQuery, metricModel)
			if err != nil {
				return resp, err
			}

			// vega查数, 记录请求的开始结束,返回到接口上
			start := time.Now().UnixMilli()
			datas, err := mms.vvs.FetchDatasFromVega(ctx, sql)
			if err != nil {
				return resp, rest.NewHTTPError(ctx, http.StatusInternalServerError,
					uerrors.Uniquery_MetricModel_InternalError_FetchDatasFromVegaFailed).WithErrorDetails(err.Error())
			}
			// 如果有请求同环比,计算同期对应的时间范围
			samePeriodDatas := interfaces.VegaFetchData{}
			if query.RequestMetrics != nil && query.RequestMetrics.Type == interfaces.METRICS_SAMEPERIOD {
				// chainStart, chainEnd := calcComparisonRanges(query)
				startTime := time.UnixMilli(*query.Start).In(common.APP_LOCATION)
				endTime := time.UnixMilli(*query.End).In(common.APP_LOCATION)
				chainStart := calcComparisonTime(startTime, *query.RequestMetrics.SamePeriodCfg).UnixMilli()
				chainEnd := calcComparisonTime(endTime, *query.RequestMetrics.SamePeriodCfg).UnixMilli()
				newQuery := query
				newQuery.Start = &chainStart
				newQuery.End = &chainEnd
				sql, err := generateSQL(ctx, newQuery, dataView, viewQuery, metricModel)
				if err != nil {
					return resp, err
				}

				samePeriodDatas, err = mms.vvs.FetchDatasFromVega(ctx, sql)
				if err != nil {
					return resp, rest.NewHTTPError(ctx, http.StatusInternalServerError,
						uerrors.Uniquery_MetricModel_InternalError_FetchDatasFromVegaFailed).WithErrorDetails(err.Error())
				}
			}
			vegaFetchDur := time.Now().UnixMilli() - start

			// 把查询结果处理成统一的格式。
			return parseVegaResult2Uniresponse(ctx, datas, samePeriodDatas, query, vegaFetchDur, metricModel)

		default:
			// 异常，不支持的查询类型
			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("Unsupport metric type %s or query type %s", query.MetricType, query.QueryType))

			return resp, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_UnsupportQuery).
				WithErrorDetails(fmt.Sprintf("Unsupport metric type %s or query type %s", query.MetricType, query.QueryType))
		}
	case interfaces.DERIVED_METRIC:
		// 衍生指标：请求参数+衍生指标的过滤条件一起传给原子指标
		derivedConfig := query.FormulaConfig.(interfaces.DerivedConfig)
		newQuery := query
		newQuery.MetricModelID = derivedConfig.DependMetricModel.ID // 查依赖指标
		// 衍生指标的过滤条件传递: 把衍生指标的过滤条件，code和配置化两种模式，不在这里把过滤条件转code。未来dsl promql支持衍生指标时，不好处理
		// 用 and 合并date_condition 和business_condition
		condition := derivedConfig.DateCondition
		if derivedConfig.DateCondition != nil && derivedConfig.BusinessCondition != nil {
			condition = &cond.CondCfg{
				Operation: cond.OperationAnd,
				SubConds:  []*cond.CondCfg{derivedConfig.DateCondition, derivedConfig.BusinessCondition},
			}
		} else if derivedConfig.BusinessCondition != nil {
			condition = derivedConfig.BusinessCondition
		}
		newQuery.Condition = condition
		newQuery.ConditionStr = derivedConfig.ConditionStr

		// 衍生指标的排序字段和having过滤作为请求原子指标的排序字段和having过滤
		newQuery.OrderByFields = append(newQuery.OrderByFields, metricModel.OrderByFields...)
		newQuery.HavingCondition = metricModel.HavingCondition

		// 衍生指标是对原子指标的过滤，其同环比就是原子指标的同环比即可。
		uniResp, _, _, err := mms.Exec(ctx, &newQuery)
		if err != nil {
			return resp, err
		}
		return processOrderHaving(uniResp, query, metricModel), nil
	case interfaces.COMPOSITED_METRIC:
		// 复合指标：转成 PromQl 的原子指标的查询
		// {{id}}+{{id2}} 替换成 metric_model("id")+metric_model("id2")，这样就把复合指标转换成了 promql的原子指标查询
		// todo: 复合指标的同环比尚未支持，复合指标的同环比需单独计算. 找到同期数,同期数就是时间范围不同,
		// 复合指标的查询时不能把同环比往下传，需要单独请求。所以把 query 的同环比置空
		requestMetrics := query.RequestMetrics
		query.RequestMetrics = nil
		query.Formula = convert.TransformExpression(query.Formula)
		query.IsCalendar = true // 复合指标暂时只支持日历步长
		metricDatas, err := mms.execPromQL(ctx, query, dataView, metricModel)
		if err != nil {
			return resp, err
		}
		if requestMetrics == nil {
			return metricDatas, nil
		}

		// 开始计算同环比
		datas := []interfaces.MetricModelData{}
		// 1. 把数据转成序列标识->MetricModelData 的映射,因为MetricModelData中的labels是map,需要按key排序后再拼接.
		currentSeriesMap, err := processMetricDataTimeSeries(metricDatas)
		if err != nil {
			return resp, err
		}
		if requestMetrics.Type == interfaces.METRICS_SAMEPERIOD {
			// chainStart, chainEnd := calcComparisonRanges(query)
			newQuery := query
			if query.IsInstantQuery {
				time := time.UnixMilli(query.Time).In(common.APP_LOCATION)
				chainTime := calcComparisonTime(time, *requestMetrics.SamePeriodCfg).UnixMilli()
				newQuery.Time = chainTime
			} else {
				startTime := time.UnixMilli(*query.Start).In(common.APP_LOCATION)
				endTime := time.UnixMilli(*query.End).In(common.APP_LOCATION)
				chainStart := calcComparisonTime(startTime, *requestMetrics.SamePeriodCfg).UnixMilli()
				chainEnd := calcComparisonTime(endTime, *requestMetrics.SamePeriodCfg).UnixMilli()

				newQuery.Start = &chainStart
				newQuery.End = &chainEnd
			}

			samePeriodDatas, err := mms.execPromQL(ctx, newQuery, dataView, metricModel)
			if err != nil {
				return resp, err
			}

			// 对统一结构计算同环比计算
			// 1. 把数据转成序列标识->MetricModelData 的映射,因为MetricModelData中的labels是map,需要按key排序后再拼接.
			samePeriodSeriesMap, err := processMetricDataTimeSeries(samePeriodDatas)
			if err != nil {
				return resp, err
			}

			datas, err = calcSamePeriodValue(ctx, currentSeriesMap, samePeriodSeriesMap, requestMetrics)
			if err != nil {
				return resp, err
			}
		}
		// 对统一结构计算占比计算
		if requestMetrics.Type == interfaces.METRICS_PROPORTION {
			datas, err = calcProportionValue(ctx, currentSeriesMap)
			if err != nil {
				return resp, err
			}
		}
		resp.Datas = datas
		resp.SeriesTotal = len(datas)
		return processOrderHaving(resp, query, metricModel), nil

	default:
		// 异常，不支持的查询类型
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Unsupport metric type %s", query.MetricType))
		return resp, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_UnsupportQuery).
			WithErrorDetails(fmt.Sprintf("Unsupport metric type %s", query.MetricType))
	}
}

func (mms *metricModelService) execPromQL(ctx context.Context, query interfaces.MetricModelQuery,
	dataView *interfaces.DataView, metricModel interfaces.MetricModel) (interfaces.MetricModelUniResponse, error) {

	resp := interfaces.MetricModelUniResponse{
		Datas:      []interfaces.MetricModelData{},
		Step:       query.StepStr,
		IsVariable: query.IsVariable,
		IsCalendar: query.IsCalendar,
	}

	// 重写filters
	err := rewritePromQLFilters(ctx, &query, dataView)
	if err != nil {
		return resp, err
	}
	// 默认从配置中读取
	seriesSize := mms.appSetting.PromqlSetting.MaxSearchSeriesSize
	if query.IsModelRequest {
		// 保存模型时的计算公式有效性校验，给size给1，避免导入模型时的异常
		seriesSize = 1
	}

	promqlQuery := interfaces.Query{
		QueryStr:            query.Formula,
		Start:               *query.Start,
		End:                 *query.End,
		IsMetricModel:       true,
		DataView:            *dataView,
		ViewQuery4Metric:    query.ViewQuery4Metric,
		Filters:             query.Filters,
		IsInstantQuery:      query.IsInstantQuery,
		IgnoringHCTS:        query.IgnoringHCTS,
		IgnoringMemoryCache: query.IgnoringMemoryCache,
		Limit:               query.Limit,
		Offset:              query.Offset,
		ModelId:             query.MetricModelID,
		ModelUpdateTime:     query.ModelUpdateTime,
		MaxSearchSeriesSize: seriesSize,
		IsModelRequest:      query.IsModelRequest,
		IsCalendar:          query.IsCalendar,
		AnalysisDims:        query.AnalysisDims,
		OrderByFields:       query.OrderByFields,
	}

	if query.IsInstantQuery {
		// promqlQuery.Start = query.Time
		// promqlQuery.End = query.Time
		promqlQuery.Interval = *query.End - *query.Start
		promqlQuery.IntervalStr = fmt.Sprintf("%dms", promqlQuery.Interval)
		// promqlQuery.LookBackDelta = query.LookBackDelta
	} else {
		// 趋势查询才从查询体中读取step
		// 日历步长未转换step，所以step为空，不能直接取值
		if query.Step != nil {
			promqlQuery.Interval = *query.Step
		}
		promqlQuery.IntervalStr = *query.StepStr
	}

	// 请求 promql 的Exec，当前是按 query_range 请求
	res, _, _, err := mms.promqlService.Exec(ctx, promqlQuery)
	if err != nil {
		logger.Errorf("Exec promql error: %s", err.Error())
		return resp, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			uerrors.Uniquery_MetricModel_InternalError_ExecPromQLFailed).WithErrorDetails(err.Error())
	}

	// 把查询结果处理成统一的格式。当前的预览是query_range查询，返回结果为 matrix
	response, err := parsePromqlResult2Uniresponse(ctx, res, query, metricModel)
	if err != nil {
		return response, err
	}
	response.VegaDurationMs = res.VegaDurationMs
	return response, nil
}

// promql 的返回结构转换为统一结构
func parsePromqlResult2Uniresponse(ctx context.Context, promqlRes interfaces.PromQLResponse,
	query interfaces.MetricModelQuery, metricModel interfaces.MetricModel) (interfaces.MetricModelUniResponse, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "Parse promql result to uniresponse")
	defer span.End()

	var resp interfaces.MetricModelUniResponse
	var err error

	seriesTotal := promqlRes.SeriesTotal

	// 接口返回的计算公式实际执行的 step 字段
	var respStep *string
	if !query.IsInstantQuery {
		respStep = query.StepStr
	}

	switch data := promqlRes.Data.(type) {
	case promql.QueryData:
		switch mat := data.Result.(type) {
		case static.PageMatrix:

			resp, err = processMatrix(mat.Matrix, query)
			if err != nil {
				return interfaces.MetricModelUniResponse{}, err
			}
			resp.Step = respStep
			resp.SeriesTotal = seriesTotal
			span.SetStatus(codes.Ok, "")
			// return resp, nil
		case static.Matrix:
			resp, err = processMatrix(mat, query)
			if err != nil {
				return interfaces.MetricModelUniResponse{}, err
			}
			resp.Step = respStep
			resp.SeriesTotal = resp.CurrSeriesNum
			span.SetStatus(codes.Ok, "")
			// return resp, nil
		case static.Vector:
			// mat 是多条序列的集合
			datas := make([]interfaces.MetricModelData, len(mat))
			for i, ss := range mat {
				// 遍历 labels
				labels := make(map[string]string)
				for _, label := range ss.Metric {
					labels[label.Name] = label.Value
				}
				// 遍历数据
				dateValues := make([]any, 1)
				timeStrs := make([]string, 1)
				values := make([]any, 1)
				dateValues[0] = ss.Point.T
				timeStrs[0] = convert.FormatTimeMiliis(ss.Point.T, "")
				values[0] = convert.WrapMetricValue(ss.Point.V)

				datas[i] = interfaces.MetricModelData{
					Labels:   labels,
					Times:    dateValues,
					TimeStrs: timeStrs,
					Values:   values,
				}
			}
			span.SetStatus(codes.Ok, "")
			if seriesTotal == 0 {
				seriesTotal = len(mat)
			}
			resp = interfaces.MetricModelUniResponse{
				Datas:         datas,
				Step:          respStep,
				CurrSeriesNum: len(mat),
				SeriesTotal:   seriesTotal,
				PointTotal:    1,
			}
			// if query.IncludeModel {
			// 	res.Model = metricModel
			// }
			// return res, nil
		case static.Scalar:
			span.SetStatus(codes.Ok, "")
			resp = interfaces.MetricModelUniResponse{
				// Model: metricModel,
				Datas: []interfaces.MetricModelData{
					{
						Labels:   map[string]string{},
						Times:    []any{mat.T},
						TimeStrs: []string{convert.FormatTimeMiliis(mat.T, "")},
						Values:   []any{convert.WrapMetricValue(mat.V)},
					},
				},
				Step:          respStep,
				CurrSeriesNum: 1,
			}
			// if query.IncludeModel {
			// 	res.Model = metricModel
			// }
			// return res, nil
		default:
			span.SetStatus(codes.Error, "Invalid Promql Result Type")
			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("Invalid Promql Result Type %q", data.Result.Type()))

			return resp, rest.NewHTTPError(ctx, http.StatusUnprocessableEntity, uerrors.Uniquery_MetricModel_InternalError_ParseResultFailed).
				WithErrorDetails(fmt.Sprintf("Invalid Promql Result Type %q", data.Result.Type()))
		}

	default:
		// 不支持的查询结果
		span.SetStatus(codes.Error, "Unsupport Promql Result Type")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Unsupport Promql Result Type %q", data))

		return resp, rest.NewHTTPError(ctx, http.StatusUnprocessableEntity, uerrors.Uniquery_MetricModel_InternalError_UnSupportPromQLResult).
			WithErrorDetails(fmt.Sprintf("Unsupport Promql Result Type %q", data))
	}

	// having过滤和排序
	return processOrderHaving(resp, query, metricModel), nil
}

// 处理最后的结果和排序、值过滤
func processOrderHaving(res interfaces.MetricModelUniResponse,
	query interfaces.MetricModelQuery, metricModel interfaces.MetricModel) interfaces.MetricModelUniResponse {

	// 1. 过滤数据（只在即时查询时生效）
	// 复合指标 + 带同环比 + dsl/promql 的指标模型的 having 在 uniquery 中实现
	if metricModel.MetricType == interfaces.COMPOSITED_METRIC ||
		metricModel.QueryType == interfaces.PROMQL ||
		metricModel.QueryType == interfaces.DSL ||
		query.RequestMetrics != nil {

		res = filterData(res, query, metricModel)
	}

	// 2. 排序数据(dsl 和 promql 的排序), sql原子、衍生、复合已经把排序下沉了
	if metricModel.QueryType == interfaces.PROMQL || metricModel.QueryType == interfaces.DSL {
		res = sortData(res, query, metricModel)
	}

	return res
}

// 过滤数据
func filterData(res interfaces.MetricModelUniResponse,
	query interfaces.MetricModelQuery, metricModel interfaces.MetricModel) interfaces.MetricModelUniResponse {

	// 只在即时查询时生效, 趋势查询时，having不生效
	if !query.IsInstantQuery {
		return res
	}

	if query.HavingCondition == nil && metricModel.HavingCondition == nil {
		return res // 无 having 过滤时，不继续过滤，返回原本数据
	}

	var filtered []interfaces.MetricModelData
	for _, data := range res.Datas {
		if len(data.Values) == 0 {
			continue
		}

		// 获取第一个值（即时查询只有一个值）
		value := data.Values[0]
		if matchesHavingCondition(value, query, metricModel) {
			filtered = append(filtered, data)
		}
	}
	res.Datas = filtered

	return res
}

// 检查值是否满足having条件
func matchesHavingCondition(value any, query interfaces.MetricModelQuery, metricModel interfaces.MetricModel) bool {
	// 将any类型转换为float64进行比较
	metricValue, err := convert.AssertFloat64(value)
	if err != nil {
		return false
	}

	// 模型配置的 having 和查询请求的 having 做 and 的合并
	havingConditions := []*cond.CondCfg{query.HavingCondition, metricModel.HavingCondition}
	match := false
	for _, condition := range havingConditions {
		if condition == nil {
			match = true
			continue
		}
		switch condition.Operation {
		case cond.OperationGt:
			v, err := convert.AssertFloat64(condition.Value)
			if err != nil {
				return false
			}
			match = metricValue > v
		case cond.OperationGte:
			v, err := convert.AssertFloat64(condition.Value)
			if err != nil {
				return false
			}
			match = metricValue >= v
		case cond.OperationLt:
			v, err := convert.AssertFloat64(condition.Value)
			if err != nil {
				return false
			}
			match = metricValue < v
		case cond.OperationLte:
			v, err := convert.AssertFloat64(condition.Value)
			if err != nil {
				return false
			}
			match = metricValue <= v
		case cond.OperationEq:
			v, err := convert.AssertFloat64(condition.Value)
			if err != nil {
				return false
			}
			match = metricValue == v
		case cond.OperationNotEq:
			v, err := convert.AssertFloat64(condition.Value)
			if err != nil {
				return false
			}
			match = metricValue != v
		case cond.OperationIn:
			value := condition.Value.([]interface{})
			for i := 0; i < len(value); i++ {
				v, err := convert.AssertFloat64(value[i])
				if err != nil {
					return false
				}
				match = metricValue != v
				if match {
					return true // in 只要匹配到一个就可以
				}
			}
		case cond.OperationNotIn:
			value := condition.Value.([]interface{})
			for i := 0; i < len(value); i++ {
				v, err := convert.AssertFloat64(value[i])
				if err != nil {
					return false
				}
				match = metricValue != v
				if !match {
					return false // not in 需要都不属于，但凡有个为false，则都都false
				}
			}
		case cond.OperationRange:
			value := condition.Value.([]interface{})
			if len(value) != 2 {
				return false
			}
			if !cond.IsSameType(value) {
				return false
			}

			gte, err := convert.AssertFloat64(value[0])
			if err != nil {
				return false
			}
			lte, err := convert.AssertFloat64(value[1])
			if err != nil {
				return false
			}
			match = metricValue >= gte && metricValue < lte // range 左闭右开
		case cond.OperationOutRange:
			value := condition.Value.([]interface{})
			if len(value) != 2 {
				return false
			}
			if !cond.IsSameType(value) {
				return false
			}

			gte, err := convert.AssertFloat64(value[0])
			if err != nil {
				return false
			}
			lte, err := convert.AssertFloat64(value[1])
			if err != nil {
				return false
			}
			match = metricValue < gte && metricValue >= lte // out range 左侧指定字段＜value[0] 或 ≥value[1] 的值
		default:
			return false
		}

		// 存在false，则结束，返回 false， true的话继续下一次遍历
		if !match {
			return false
		}
	}
	return match
}

// 排序数据
func sortData(res interfaces.MetricModelUniResponse,
	query interfaces.MetricModelQuery, metricModel interfaces.MetricModel) interfaces.MetricModelUniResponse {

	orderByFields := []interfaces.OrderField{}
	orderByFields = append(orderByFields, query.OrderByFields...)
	orderByFields = append(orderByFields, metricModel.OrderByFields...)
	// 无排序字段，直接返回
	if len(orderByFields) == 0 {
		return res
	}

	sorted := make([]interfaces.MetricModelData, len(res.Datas))
	copy(sorted, res.Datas)

	sort.Slice(sorted, func(i, j int) bool {
		// 遍历所有排序字段
		for _, orderField := range orderByFields {
			fieldName := orderField.Name

			// 如果是值字段的排序，则从values里取值
			if fieldName == interfaces.VALUE_FIELD {
				// 获取两个数据项在当前字段的值，按第一个时间点的值进行排序
				valI, err := convert.AssertFloat64(sorted[i].Values[0])
				if err != nil {
					return false
				}
				valJ, err := convert.AssertFloat64(sorted[i].Values[0])
				if err != nil {
					return false
				}

				// 如果值不相等，根据排序方向决定顺序
				if valI != valJ {
					if orderField.Direction == interfaces.DESC_DIRECTION {
						// 降序：valI > valJ 时返回 true
						return valI > valJ
					} else {
						// 升序（默认）：valI < valJ 时返回 true
						return valI < valJ
					}
				}
				// 如果当前字段值相等，继续比较下一个字段
			} else {
				// 获取两个数据项在当前字段的值
				valI := sorted[i].Labels[fieldName]
				valJ := sorted[j].Labels[fieldName]

				// 如果值不相等，根据排序方向决定顺序
				if valI != valJ {
					if orderField.Direction == interfaces.DESC_DIRECTION {
						// 降序：valI > valJ 时返回 true
						return valI > valJ
					} else {
						// 升序（默认）：valI < valJ 时返回 true
						return valI < valJ
					}
				}
				// 如果当前字段值相等，继续比较下一个字段
			}
		}

		// 所有排序字段的值都相等，保持原始顺序
		return false
	})

	res.Datas = sorted
	return res
}

func processMatrix(mat static.Matrix, query interfaces.MetricModelQuery) (interfaces.MetricModelUniResponse, error) {
	// mat 是多条序列的集合
	datas := make([]interfaces.MetricModelData, len(mat))
	pointTotal := 0
	// timeZone, err := time.LoadLocation(os.Getenv("TZ"))
	// if err != nil {
	// 	// 记录异常日志
	// 	o11y.Error(ctx, fmt.Sprintf("LoadLocation error: %s", err.Error()))

	// 	return interfaces.MetricModelUniResponse{}, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_MetricModel_InvaliEnvironmentVariable_TZ).
	// 		WithErrorDetails(err.Error())
	// }

	// bucket_key = localToUtc(Math.floor(utcToLocal(value) / interval) * interval))
	// bucket_key计算规则: ((start+offset)/step)*step-offset
	// 时间修正, 同期值和本期值不同,同期值需要把时间置到同期的时间轴上
	fixedStart, fixedEnd := correctingTime(query, common.APP_LOCATION)

	// 生成完整的时间点
	allTimes := make([]int64, 0)
	for currentTime := fixedStart; currentTime <= fixedEnd; {
		allTimes = append(allTimes, currentTime)
		currentTime = getNextPointTime(query, currentTime)
	}

	// _, offset := time.Now().In(common.APP_LOCATION).Zone()
	// fixedStart := int64(math.Floor(float64(query.Start+int64(offset*1000))/float64(query.Step)))*query.Step - int64(offset*1000)
	// fixedEnd := int64(math.Floor(float64(query.End+int64(offset*1000))/float64(query.Step)))*query.Step - int64(offset*1000)

	for i, ss := range mat {
		// 遍历 labels
		labels := make(map[string]string)
		for _, label := range ss.Metric {
			labels[label.Name] = label.Value
		}

		// 遍历数据
		pointTotal += len(ss.Points)
		var dateValues []any
		var values []any
		var timeStrs []string

		// 判断points数组是否为空，如果为空则不进行补点
		if len(ss.Points) > 0 {
			//dateValues和values数组的长度将变化，不再是Points的长度
			// length := (fixedEnd - fixedStart) / query.Step
			dateValues = make([]any, 0, len(allTimes))
			timeStrs = make([]string, 0, len(allTimes))
			values = make([]any, 0, len(allTimes))

			pointIndex := 0
			for time := fixedStart; time <= fixedEnd; {

				// 时间修正可能使得修正的时间大于第一个桶时间
				if time > ss.Points[pointIndex].T {
					time = ss.Points[pointIndex].T
				}

				// 如果当前的时间小于当前点的时间，需要一直补点到两者时间相等 time==ss.Points[pointIndex].T
				if time < ss.Points[pointIndex].T {
					for time < ss.Points[pointIndex].T {
						dateValues = append(dateValues, time)
						timeStrs = append(timeStrs, convert.FormatTimeMiliis(time, *query.StepStr))
						values = append(values, nil)
						time = getNextPointTime(query, time)
					}
				} else if time == ss.Points[pointIndex].T {
					dateValues = append(dateValues, time)
					timeStrs = append(timeStrs, convert.FormatTimeMiliis(time, *query.StepStr))
					values = append(values, convert.WrapMetricValue(ss.Points[pointIndex].V))
					time = getNextPointTime(query, time)
					pointIndex++
				}

				// pointIndex==len(Points) 说明Points已经遍历结束了
				if pointIndex == len(ss.Points) {
					// 如果此时time <= FixedEnd，说明后面缺点
					for time <= fixedEnd {
						dateValues = append(dateValues, time)
						timeStrs = append(timeStrs, convert.FormatTimeMiliis(time, *query.StepStr))
						values = append(values, nil)
						time = getNextPointTime(query, time)
					}
				}

			}
		}

		datas[i] = interfaces.MetricModelData{
			Labels:   labels,
			Times:    dateValues,
			TimeStrs: timeStrs,
			Values:   values,
		}
	}

	res := interfaces.MetricModelUniResponse{
		Datas:         datas,
		CurrSeriesNum: len(mat),
		PointTotal:    pointTotal,
	}
	// if query.IncludeModel {
	// 	res.Model = metricModel
	// }
	return res, nil
}

// 匹配持久化任务
func (mms *metricModelService) macthPersist(ctx context.Context, query *interfaces.MetricModelQuery,
	metricModel interfaces.MetricModel, dataView *interfaces.DataView) (*interfaces.DataView, error) {

	match := false
	var (
		taskName, step, timeWindow, taskBaseType string
		delta                                    int64
	)
	// 先用过滤条件匹配，过滤条件匹配通过之后匹配时间窗口,过滤条件的比较用整个filter的序列化做比较。
	// promql： 范围查询用 step 去匹配持久化任务的 step； 即时查询，用 look_back_delta 去匹配持久化任务的 step。
	// dsl： 即时查询，实时查，不匹配持久化；范围查询：用解析完dsl后的实际应该查询的 interval 去匹配持久化的时间窗口和step。
	// if metricModel.QueryType == interfaces.DSL && query.IsInstantQuery {
	// 	// dsl： 即时查询，实时查，不匹配持久化
	// 	return nil
	// }
	if metricModel.Task != nil {
		switch metricModel.QueryType {
		case interfaces.PROMQL:
			// promql： 范围查询用 step 去匹配持久化任务的 step； 即时查询，用 look_back_delta 去匹配持久化任务的 step。
			var compareStep int64
			if query.IsInstantQuery {
				delta = *query.End - *query.Start
				compareStep = delta
			} else {
				compareStep = *query.Step
			}
			for _, stepi := range metricModel.Task.Steps {
				// 这里的持久化步长匹配应做单位转换后再比较
				stepiT, err := convert.ParseDuration(stepi)
				if err != nil {
					httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Step).
						WithErrorDetails(err.Error())
					return dataView, httpErr
				}
				if compareStep == stepiT.Milliseconds() {
					// 相等，则查
					match = true
					taskName = metricModel.Task.TaskName
					step = stepi
					taskBaseType = metricModel.Task.IndexBase
				}
			}
		case interfaces.DSL:
			// dsl： 即时查询，用look_back_delta和持久化步长、时间窗口匹配即可查持久化数据，
			// 如果多个时间窗口，那就用look_back_delta同时匹配时间窗口
			dslInfo, err := parseDsl(ctx, query, dataView)
			if err != nil {
				return dataView, err
			}
			// 按请求处理date_histogram的配置信息;当是 range 查询时,改写rangedsl的aggs中的date_histogram的值
			delta, err = processDateHistogram(ctx, query, dslInfo, mms.appSetting.PromqlSetting.LookbackDelta)
			if err != nil {
				return dataView, err
			}
			var compareStep int64
			var compareWindow int64
			if query.IsInstantQuery {
				// dsl 持久化配置时，步长决定数据的持久化步长，时间窗口决定数据的查询时间范围。
				// 直接请求的时候，dsl的数据查询时间跟计算公式的interval是常量还是变量有关:
				// 1. 固定步长，则按固定步长interval去查询数据
				// 2. 变量步长，则用请求的look_back_delta去查询数据
				// 所以，即时查询时，用look_back_delta去匹配持久化步长，用计算公式的interval去匹配时间窗口
				compareStep = delta
				compareWindow = *query.End - *query.Start // 查询数据的时间范围被解析到query的start和end中，转成了时间范围
			} else {
				// 范围查询：用解析完dsl后的实际应该查询的 interval 去匹配持久化的时间窗口和step。
				// 范围直接查询时，隐含的是步长和时间窗口是相等的。
				// 直接请求的时候，dsl的数据查询时间跟计算公式的interval是常量还是变量有关:
				// 1. 固定步长，则按固定步长interval作为step去查询数据
				// 2. 变量步长，则用请求的step去查询数据
				// 所以，范围查询时，用step去匹配持久化步长、时间窗口
				// 解析 query.StepStr 为ms时间
				compareStepT, err := convert.ParseDuration(*query.StepStr)
				if err != nil {
					httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Step).
						WithErrorDetails(err.Error())
					return dataView, httpErr
				}
				compareStep = compareStepT.Milliseconds()
				compareWindow = compareStep
			}

			for _, stepi := range metricModel.Task.Steps {
				stepiT, err := convert.ParseDuration(stepi)
				if err != nil {
					httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Step).
						WithErrorDetails(err.Error())
					return dataView, httpErr
				}
				if compareStep == stepiT.Milliseconds() {
					for _, window := range metricModel.Task.TimeWindows {
						windowT, err := convert.ParseDuration(window)
						if err != nil {
							httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Step).
								WithErrorDetails(err.Error())
							return dataView, httpErr
						}
						if compareWindow == windowT.Milliseconds() {
							// 相等，则查
							match = true
							taskName = metricModel.Task.TaskName
							step = stepi
							timeWindow = window
							taskBaseType = metricModel.Task.IndexBase
						}
					}
				}
			}
		}
	}

	if match {
		// 如果匹配，采样查询
		// 获取索引库对应的内置视图，内置视图的id为 __索引库类型
		dataViewPersist, err := mms.dvService.GetDataViewByID(ctx, fmt.Sprintf("__%s", taskBaseType), true)
		if err != nil {
			return dataView, err
		}

		// 过滤条件置空
		// query.Filters = []interfaces.Filter{}
		// 若存在过滤条件的字段不在baseFieldsMap中，则不走持久化查询
		for _, filter := range query.Filters {
			if _, exist := dataView.FieldsMap[interfaces.LABELS_PREFIX+"."+filter.Name]; !exist {
				return dataView, nil
			}
		}

		dataView = dataViewPersist

		switch metricModel.QueryType {
		case interfaces.PROMQL:
			query.Formula = fmt.Sprintf(`%s{task_name="%s",step="%s"}`, metricModel.MeasureName, taskName, step)
		case interfaces.DSL:
			query.Formula = fmt.Sprintf(`%s{task_name="%s",step="%s",time_window="%s"}`, metricModel.MeasureName, taskName, step, timeWindow)
		}
		// 查询语言重置为 promql
		query.QueryType = interfaces.PROMQL

		// 如果是instant query，则需要把time按持久化步长（匹配上了，此时也是回退时长）修正后再查持久化数据
		if query.IsInstantQuery {
			start := int64(math.Floor(float64(*query.Start)/float64(delta))) * delta
			end := int64(math.Floor(float64(*query.End)/float64(delta))) * delta
			query.Start = &start
			query.End = &end
		}
		query.HasMatchPersist = true
	}
	return dataView, nil
}

// 用于缓存当前模型下的目标序列，缓存有效时间10min
var Series_Of_Model_Map sync.Map

// 用于缓存当前模型的配置信息,缓存有效时间10min
var Dsl_Info_Of_Model sync.Map

type DSLInfoCache struct {
	RefreshTime time.Time
	DslInfo     interfaces.DslInfo
}
type DSLSeries struct {
	RefreshTime     time.Time // 增量缓存刷新时间
	StartTime       time.Time // 缓存序列的数据的start
	EndTime         time.Time // 缓存序列的数据的end
	FullRefreshTime time.Time // 全量刷新时间
	Series          []map[string]string
}

// 获取指标模型(DSL)的序列。todo:后续考虑异步来刷新序列缓存。（查询方案比较复杂，后续可以考虑compound_aggs来处理，现在compound性能不好）
func (mms *metricModelService) getSeries(ctx context.Context, query interfaces.MetricModelQuery,
	dslInfo *interfaces.DslInfo, dataView *interfaces.DataView) ([]map[string]string, error) {

	var filter interfaces.Filter
	now := time.Now()
	// 忽略缓存时不走缓存
	dslSeriesCache, start, end, fullFlag, ifFromCache := getSeriesOrTimeFilter(query,
		mms.appSetting.ServerSetting.FullCacheRefreshInterval, dataView)
	if ifFromCache {
		return dslSeriesCache.Series, nil
	}

	filter = interfaces.Filter{
		Name:      "@timestamp",
		Operation: cond.OperationRange,
		Value:     []any{start, end},
	}

	// 1.获取terms的基数
	indexPattern := convert.GetIndexBasePattern(query.ViewQuery4Metric.BaseTypes)

	carAggs := generatCardinalityAggs(dslInfo.TermsInfos)
	carDsl, err := generateDsl(ctx, dslInfo.DSLQuery, carAggs, []interfaces.Filter{filter}, dataView)
	if err != nil {
		return nil, err
	}
	str, _ := sonic.MarshalString(carDsl)
	logger.Debugf("模型[%s]获取terms字段基数生成的dsl语句: %s", query.ModelName, str)

	// // 查全部,用indexPattern来查
	carRes, _, err := mms.dvService.GetDataFromOpenSearch(ctx, carDsl, indexPattern,
		interfaces.DEFAULT_SCROLL, interfaces.DEFAULT_PREFERENCE, interfaces.DEFAULT_TRACK_TOTAL_HITS)
	if err != nil {
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_InternalError).
			WithErrorDetails(err.Error())
	}

	condition := true
	// 2.读取terms字段的基数数据，并算出当前请求获取的size
	aggs := gjson.Get(string(carRes), "aggregations")
	for i, terms := range dslInfo.TermsInfos {
		key := fmt.Sprintf(`%s.value`, strings.Replace(terms.TermsField, ".", "\\.", -1))
		cardi := aggs.Get(key).Int()
		// 如果基数为0，则表示当前字段不存在，返回缓存获取到的series
		if cardi == 0 {
			// 如果全量刷新没有，那就返回空的series，若是增量刷新没有，那就返回缓存中的seies
			if fullFlag {
				return nil, nil
			} else {
				return dslSeriesCache.Series, nil // 无需查询，返回缓存中的series
			}
		}
		dslInfo.TermsInfos[i].EvalSize = min(cardi, terms.ConfigSize) // terms的实际size
	}
	// 3.计算每批请求terms的size。获取序列时，只需要获取terms的即可。
	// (需把filters,range,date_range的大小除去)--去掉,获取序列时无需考虑非terms的大小
	remainSize := interfaces.DEFAULT_GET_SERIES_NUM_BY_BATCH
	fieldBatchNum := make(map[string]int64) // 每个字段查询的批次, 用于应对相同字段的terms的情况.用于计算获取序列集的总遍历次数
	for i := len(dslInfo.TermsInfos) - 1; i >= 0; i-- {
		// 当前字段的计算size大于能查的序列数,则当前字段一批查询的size为能查的序列数
		sizei := min(remainSize, dslInfo.TermsInfos[i].EvalSize)

		// 每个字段最后一个批次的大小
		lastBatchSize := dslInfo.TermsInfos[i].EvalSize % sizei // 最后一批size是余数
		if lastBatchSize == 0 {
			lastBatchSize = sizei // 完整批,那么最后一批的size跟前面分批size一样.
		}
		batch := interfaces.SeriesTerms{
			BatchSize:     sizei,                                                                      // 每批次的查询次数
			BatchNum:      int64(math.Ceil(float64(dslInfo.TermsInfos[i].EvalSize) / float64(sizei))), // 每个terms要遍历的次数
			LastBatchSize: lastBatchSize,                                                              // 最后一个批次的大小
		}
		dslInfo.TermsInfos[i].SeriesTerms = batch

		terms := dslInfo.AggInfos[dslInfo.TermsToAggs[i]]
		terms.SeriesTerms = batch
		dslInfo.AggInfos[dslInfo.TermsToAggs[i]] = terms
		fieldBatchNum[dslInfo.TermsInfos[i].TermsField] = dslInfo.TermsInfos[i].BatchNum

		remainSize /= sizei
	}

	// 总遍历次数
	cnt := int64(1)
	for _, v := range fieldBatchNum {
		cnt *= v
	}

	// 4.改写dsl的aggs的size
	// 生成获取series的aggs
	seriesAggs := generateSeriesAggs(*dslInfo, false)
	series := make([]map[string]string, 0)
	var curRoundFilter []interfaces.Filter
	batchi := int64(0)
	for condition {
		filters := make([]interfaces.Filter, 0)
		filters = append(filters, filter) // 每次查询都应把处理缓存时得到的过滤条件拼接上
		seriesi := make([]map[string]string, 0)
		labels := make(map[string]string, 0)
		ifLastBacth := false
		// 获取当前批的序列的查询条件filter.
		getSeriesPagedFilters(series, len(dslInfo.TermsInfos)-1, *dslInfo,
			interfaces.DEFAULT_GET_SERIES_NUM_BY_BATCH, &filters, &ifLastBacth, &batchi)

		filters = append(filters, curRoundFilter...)
		var seriesDsl map[string]any
		if ifLastBacth {
			// 生成没批的最后一个批次series请求的aggs,最后一个批次请求的size会变
			lastBatchSeriesAggs := generateSeriesAggs(*dslInfo, ifLastBacth)
			// 每个批次的series的请求的过滤条件会变更,query是map,所以每次用新的query对象与当前批次的过滤条件合并
			seriesDsl, err = generateDsl(ctx, dslInfo.DSLQuery, lastBatchSeriesAggs, filters, dataView)
			if err != nil {
				return nil, err
			}
		} else {
			// 每个批次的series的请求的过滤条件会变更,query是map,所以每次用新的query对象与当前批次的过滤条件合并
			seriesDsl, err = generateDsl(ctx, dslInfo.DSLQuery, seriesAggs, filters, dataView)
			if err != nil {
				return nil, err
			}
		}

		str, _ := sonic.MarshalString(seriesDsl)
		logger.Debugf("模型【%s】分批获取序列生成的dsl; %s", query.ModelName, str)

		// 查全部,用indexPattern来查
		seriesRes, _, err := mms.dvService.GetDataFromOpenSearch(ctx, seriesDsl, indexPattern,
			interfaces.DEFAULT_SCROLL, interfaces.DEFAULT_PREFERENCE, interfaces.DEFAULT_TRACK_TOTAL_HITS)
		if err != nil {
			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_InternalError).
				WithErrorDetails(err.Error())
		}
		// 读取数据，并算出当前请求获取的size
		seriesAggsRes := gjson.Get(string(seriesRes), "aggregations")
		IteratorSeriesAggs(seriesAggsRes, labels, &seriesi, 0, dslInfo.TermsInfos)
		series = append(series, seriesi...)

		if len(seriesi) == 0 && batchi+1 < cnt {
			// 如果当前批查不到数据，且batchi小于总的次数，那么跳过当前，往下一层走
			batchi++
			continue
		}

		if len(seriesi) == 0 || batchi+1 >= cnt || query.MaxSearchSeriesSize == 1 {
			condition = false
		}
		batchi++
	}

	if fullFlag {
		// 全量刷。查询到的tsid为全量。直接返回和缓存
		dslSeriesCache = DSLSeries{
			FullRefreshTime: now,
			Series:          series,
		}
	} else {
		// 增量、使用请求时间，需逐个添加到缓存中
		// 把 series 数组中的序列添加到缓存中，需要去重判断.
		// 合并逻辑需要准确。k个字段。
		sortKeys := make([]convert.SortKey, 0)
		for i, v := range dslInfo.TermsInfos {
			ascending := true
			if v.Direction == interfaces.DESC_DIRECTION {
				ascending = false
			}
			sortKeys = append(sortKeys, convert.SortKey{
				Key:       v.AggName,
				Ascending: ascending,
			})

			// 增量查询序列后，增量基数不能代表全量基数的场景，需要更改evalSize为配置的size
			dslInfo.TermsInfos[i].EvalSize = dslInfo.TermsInfos[i].ConfigSize
		}

		dslSeriesCache.Series = convert.MergeAndDeduplicate(dslSeriesCache.Series, series, sortKeys)
	}

	dslSeriesCache.StartTime = time.UnixMilli(start)
	dslSeriesCache.EndTime = time.UnixMilli(end)
	dslSeriesCache.RefreshTime = now

	if query.MetricModelID != "" && len(query.Filters) == 0 && len(series) != 0 {
		// 写入缓存(不为空才写入缓存) dslSeriesCache.RefreshTime = now
		Series_Of_Model_Map.Store(query.MetricModelID, dslSeriesCache)
	}
	return dslSeriesCache.Series, nil
}

// 直接查询opensearch,无需分批
func (mms *metricModelService) getDSLDataDirectly(ctx context.Context, query interfaces.MetricModelQuery, indices []string,
	dslInfo interfaces.DslInfo, filters []interfaces.Filter, metricModel interfaces.MetricModel, dataView *interfaces.DataView) (interfaces.MetricModelUniResponse, error) {

	// 需拼接上时间过滤,在前面parseDSL时,只加上了公式里的过滤条件+接口中的filters+视图的过滤条件
	var resp interfaces.MetricModelUniResponse
	// 添加时间范围的过滤条件
	// filters := make([]interfaces.Filter, 0)
	filters = append(filters, interfaces.Filter{
		Name:      "@timestamp",
		Operation: cond.OperationRange,
		Value:     []any{*query.Start, *query.End},
	})

	if query.IsModelRequest {
		// 如果是来自模型的计算公式有效性检查,构造一个使得数据过滤为空的过滤条件来避开聚合的时间开销
		filters = append(filters, interfaces.Filter{
			Name:      "__invalid_field",
			Operation: cond.OperationEq,
			Value:     "----------------",
		})
	}

	// 如果是range查询,需要对改变date_histogram的interval,所以range和instant的aggs是不一样的
	queryAggs := dslInfo.RangeQueryDSL["aggs"].(map[string]any)
	if query.IsInstantQuery {
		queryAggs = dslInfo.InstantQueryDSL["aggs"].(map[string]any)
	}

	dsl, err := generateDsl(ctx, dslInfo.DSLQuery, queryAggs, filters, dataView)
	if err != nil {
		return resp, err
	}
	str, _ := sonic.MarshalString(dsl)
	logger.Debugf("模型【%s】直接查询数据生成的dsl: %s", metricModel.ModelName, str)

	// 3. 向 opensearch 发起基于数据视图 id 的 dsl 查询。
	submitDslCtx, submitDslSpan := ar_trace.Tracer.Start(ctx, "请求 opensearch 查询数据")
	submitDslSpan.SetAttributes(attribute.Key("dsl_expression").String(fmt.Sprintf("%s", dslInfo.RangeQueryDSL)),
		attribute.Key("index").StringSlice(indices))

	res, _, err := mms.dvService.GetDataFromOpenSearch(submitDslCtx, dsl, indices,
		interfaces.DEFAULT_SCROLL, interfaces.DEFAULT_PREFERENCE, interfaces.DEFAULT_TRACK_TOTAL_HITS)
	// 记录请求信息和返回信息
	o11y.Info(submitDslCtx, fmt.Sprintf("Opensearch search index: [%s], dsl: [%s]; search response err is [%s]",
		indices, dslInfo.RangeQueryDSL, err))

	if err != nil {
		// 记录异常日志
		o11y.Error(submitDslCtx, fmt.Sprintf("Opensearch search error: %v", err))

		submitDslSpan.SetStatus(codes.Error, "Opensearch search error")
		submitDslSpan.End()
		return resp, rest.NewHTTPError(submitDslCtx, http.StatusInternalServerError, uerrors.Uniquery_InternalError).
			WithErrorDetails(err.Error())
	}
	submitDslSpan.SetStatus(codes.Ok, "")
	submitDslSpan.End()
	// 4. 把查询结果处理成统一的格式。当前的预览是query_range查询，返回结果为 matrix
	return parseDSLResult2Uniresponse(ctx, res, dslInfo, query, metricModel)
}

// 获取序列数据,计划对series分批获取
func (mms *metricModelService) getSeriesData(ctx context.Context, query interfaces.MetricModelQuery,
	dslInfo interfaces.DslInfo, metricModel interfaces.MetricModel, series []map[string]string,
	indices []string, shardsTotal int, dataView *interfaces.DataView) (interfaces.MetricModelUniResponse, error) {

	// 时间分桶个数
	// timesNum := max(1, int64(math.Ceil(float64(query.End-query.Start)/float64(query.Step))))
	//  一批能查的terms序列数 = 最大点数/非terms聚合的个数所占的长度/时间点数
	batchSeriesNum := interfaces.DEFAULT_DSL_MAX_QUERY_POINTS / dslInfo.NotTermsSeriesNum / query.QueryTimeNum
	// if batchSeriesNum > int64(len(series)) {
	// 再看最大序列数,大于1000的序列认为是高基,分批查
	batchMax := interfaces.DEFAULT_SERIES_NUM / dslInfo.NotTermsSeriesNum
	if int64(len(series)) > batchMax {
		batchSeriesNum = batchMax
	}
	// }
	// 计算每批次查询的序列范围
	if batchSeriesNum >= int64(len(series)) {
		// 一批能查的terms的序列数 >= terms实际的序列数,那么一次就能查完,无需分批.
		// 例如,terms a,b; filters[1,2],当a和b的基数各位100,而a和b的组合也是100,此时进入分批了,
		// 但是a和b组合不同的关系,导致实际上没有预估的那么多序列,所以存在可以一次查完的情况
		// 此时series里的都满足一次查询出来，那么加上第一个字段的过滤条件。
		batchFilters := processDiractlyDataFilters(dslInfo, series)
		return mms.getDSLDataDirectly(ctx, query, indices, dslInfo, batchFilters, metricModel, dataView)
	} else {
		// 返回结果中增加分批查询的标记.(按_count排序时,触发高基分批查询会改变排序方式,改为按key正序)
		// 分批查.以最后一个为一行,重写每批次的各个terms字段的size+识别过滤条件,再获取数据
		batchFilters := processSeriesDataFilters(query, batchSeriesNum, dslInfo, series)
		batchResultChs := make(chan *BatchResult, len(batchFilters))
		defer close(batchResultChs)
		// 按索引+分片数为单位并发，索引库下有多个索引。
		// 这块只要有一个协程发生异常，则可以中断所有协程返回。现在是等待所有的都执行结束，需优化。
		var wg sync.WaitGroup
		wg.Add(len(batchFilters))
		weigth := int64(min(shardsTotal, mms.appSetting.PoolSetting.ExecutePoolSize))
		for _, filters := range batchFilters {
			// 先从sem中获取weight个值，表示占用weight个并发位置。todo:后续可观测性加上sem的信息
			if err := mms.sem.Acquire(ctx, weigth); err != nil {
				// 处理错误，比如context被取消
				return interfaces.MetricModelUniResponse{}, err
			}

			err := util.ExecutePool.Submit(mms.batchSubmitTaskFuncWapper(ctx, query, filters,
				indices, dslInfo, metricModel, dataView, weigth, batchResultChs, &wg))
			if err != nil {
				// 记录异常日志
				o11y.Error(ctx, fmt.Sprintf("ExecutePool.Submit error: %v", err))
				return interfaces.MetricModelUniResponse{}, err
			}
		}
		// 等待所有执行结束
		wg.Wait()

		var unires interfaces.MetricModelUniResponse
		// 把各批次的数据合在一个集合中
		datas := make([]interfaces.MetricModelData, 0)
		for j := 0; j < len(batchFilters); j++ {
			batchResi := <-batchResultChs
			if batchResi.Err != nil {
				return interfaces.MetricModelUniResponse{}, batchResi.Err
			}
			if j == 0 {
				unires = batchResi.UniResponse
			}
			datas = append(datas, batchResi.UniResponse.Datas...)
		}
		unires.Datas = datas
		unires.IsQueryByBatch = true
		unires.SeriesTotal = len(datas)

		return unires, nil
	}
}

type taskFunc func()
type BatchResult struct {
	UniResponse interfaces.MetricModelUniResponse
	Err         error
}

// 分批次并发获取序列数据
func (mms *metricModelService) batchSubmitTaskFuncWapper(ctx context.Context,
	query interfaces.MetricModelQuery, filters []interfaces.Filter, indices []string,
	dslInfo interfaces.DslInfo, metricModel interfaces.MetricModel, dataView *interfaces.DataView, weigth int64,
	batchResultChs chan<- *BatchResult, wg *sync.WaitGroup) taskFunc {

	return func() {
		defer wg.Done()
		defer mms.sem.Release(weigth)

		// 每个批次获取数据的过滤条件会变,query是map,所以每次用新的query对象与当前批次的过滤条件合并
		queryAggs := dslInfo.RangeQueryDSL["aggs"].(map[string]any)
		if query.IsInstantQuery {
			queryAggs = dslInfo.InstantQueryDSL["aggs"].(map[string]any)
		}

		dsl, err := generateDsl(ctx, dslInfo.DSLQuery, queryAggs, filters, dataView)
		if err != nil {
			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("generateDsl error: %v", err))
			batchResultChs <- &BatchResult{
				Err: err,
			}
			return
		}

		str, _ := sonic.MarshalString(dsl)
		logger.Debugf("模型[%s]分批查询生成的dsl语句: ", metricModel.ModelName, str)

		res, _, err := mms.dvService.GetDataFromOpenSearch(ctx, dsl, indices,
			interfaces.DEFAULT_SCROLL, interfaces.DEFAULT_PREFERENCE, interfaces.DEFAULT_TRACK_TOTAL_HITS)
		if err != nil {
			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("Opensearch search error: %v", err))
			batchResultChs <- &BatchResult{
				Err: rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_InternalError).
					WithErrorDetails(err.Error()),
			}
			return
		}
		respi, err := parseDSLResult2Uniresponse(ctx, res, dslInfo, query, metricModel)
		if err != nil {
			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("parseDSLResult2Uniresponse error: %v", err))
			batchResultChs <- &BatchResult{
				Err: err,
			}
			return
		}

		batchResultChs <- &BatchResult{UniResponse: respi}
	}
}

// 重写promql过滤条件
func rewritePromQLFilters(ctx context.Context, query *interfaces.MetricModelQuery, dataView *interfaces.DataView) error {
	newFilters := make([]interfaces.Filter, 0)
	// 1. 识别过滤条件的原始字段名。原始字段无需转换，结果字段需要找到映射关系
	for i := range query.Filters {
		name := query.Filters[i].Name
		if query.Filters[i].IsResultField {
			// 结果字段找到映射关系：
			// 1. promql的结果字段映射的原始字段是 labels.字段名，暂时忽略labels_join和labels_replace对结果字段的影响
			name = interfaces.LABELS_PREFIX + "." + query.Filters[i].Name

		}
		err := processFiltersAccordingMode(ctx, query, name, query.Filters[i], &newFilters, dataView)
		if err != nil {
			return err
		}
	}
	query.Filters = newFilters
	return nil
}

// 重写promql过滤条件
func rewriteDSLFilters(ctx context.Context, query *interfaces.MetricModelQuery, dslInfo interfaces.DslInfo,
	dataView *interfaces.DataView) error {

	newFilters := make([]interfaces.Filter, 0)
	// 1. 识别过滤条件的原始字段名。原始字段无需转换，结果字段需要找到映射关系
	for i := range query.Filters {
		name := query.Filters[i].Name
		if query.Filters[i].IsResultField {
			// 结果字段找到映射关系：
			// dsl的结果字段是聚合字段，从aggInfo中获取配置信息，找到映射的字段
			for _, aggInfo := range dslInfo.AggInfos {
				if query.Filters[i].Name == aggInfo.AggName {
					switch aggInfo.AggType {
					case interfaces.BUCKET_TYPE_TERMS, interfaces.BUCKET_TYPE_RANGE, interfaces.BUCKET_TYPE_DATE_RANGE:
						name = aggInfo.TermsField
					case interfaces.BUCKET_TYPE_FILTERS:
						return fmt.Errorf("filters 聚合的结果字段不支持过滤，请选择其他字段进行过滤")
					}
				}
			}

		}

		err := processFiltersAccordingMode(ctx, query, name, query.Filters[i], &newFilters, dataView)
		if err != nil {
			return err
		}
	}
	query.Filters = newFilters
	return nil
}

// 根据过滤模型处理filters
func processFiltersAccordingMode(ctx context.Context, query *interfaces.MetricModelQuery, name string, filter interfaces.Filter,
	newFilters *[]interfaces.Filter, dataView *interfaces.DataView) error {

	_, exist := dataView.FieldsMap[name]
	if !exist {
		// 当字段不存在时，根据过滤模式对字段做相应的处理
		switch mode := query.FilterMode; {
		case mode == interfaces.FILTER_MODE_ERROR:
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterName).
				WithErrorDetails(fmt.Sprintf("filter field [%s] donot belong to fields of data source", name))
		case mode == interfaces.FILTER_MODE_IGNORE:
			// do nothing
		case mode == interfaces.FILTER_MODE_NORMAL:
			filter.Name = name
			*newFilters = append(*newFilters, filter)
		}
	} else {
		// 存在时，重写filter
		filter.Name = name
		*newFilters = append(*newFilters, filter)
	}
	return nil
}

// 获取指标模型的原始字段集
func (mms *metricModelService) GetMetricModelFields(ctx context.Context, modelID string) ([]interfaces.Field, error) {
	// 1.获取指标模型的对象信息，include_view为true时包含了视图的字段信息。
	// 获取指标模型信息
	ctx, span := ar_trace.Tracer.Start(ctx, "查询指标模型的字段列表")
	defer span.End()

	// 决策当前模型id的数据查询权限
	err := mms.ps.CheckPermission(ctx, interfaces.Resource{
		ID:   modelID,
		Type: interfaces.RESOURCE_TYPE_METRIC_MODEL,
	}, []string{interfaces.OPERATION_TYPE_DATA_QUERY})
	if err != nil {
		return nil, err
	}

	metricModels, exists, err := mms.mmAccess.GetMetricModel(ctx, modelID)
	if err != nil {
		logger.Errorf("Get Metric Model error: %s", err.Error())

		// 添加异常时的 trace 属性
		span.SetAttributes(attribute.Key("model_id").String(modelID))
		span.SetStatus(codes.Error, "Get Metric Model error")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Get Metric Model error: %v", err))

		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			uerrors.Uniquery_MetricModel_InternalError_GetModelByIdFailed).WithErrorDetails(err.Error())
	}
	if !exists || len(metricModels) == 0 {
		logger.Debugf("Metric Model %s not found!", modelID)

		// 添加异常时的 trace 属性
		span.SetAttributes(attribute.Key("model_id").String(modelID))
		span.SetStatus(codes.Error, "Metric Model not found!")

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Metric Model [%s] not found!", modelID))

		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, uerrors.Uniquery_MetricModel_MetricModelNotFound)
	}
	model := metricModels[0]
	fields := make([]interfaces.Field, 0)
	switch model.MetricType {
	case interfaces.ATOMIC_METRIC:
		switch qt, dsType := model.QueryType, model.DataSource.Type; {
		case qt == interfaces.DSL:
			// 2. dsl时把指标模型身上的字段map转换成数组输出
			for _, field := range model.FieldsMap {
				fields = append(fields, field)
			}
		case qt == interfaces.PROMQL:
			// 3. promql时，解析表达式，在叶子节点从tsid中把字段结构拿到，
			// 输出字段列表到上层，对于promql来说，所有字段都是keyword
			dataView, err := mms.dvService.GetDataViewByID(ctx, model.DataSource.ID, true)
			if err != nil {
				span.SetStatus(codes.Error, "Get data view by ID failed")
				return nil, err
			}
			promqlQuery := interfaces.Query{
				ModelId:             model.ModelID,
				QueryStr:            model.Formula,
				DataView:            *dataView,
				MaxSearchSeriesSize: mms.appSetting.PromqlSetting.MaxSearchSeriesSize,
			}
			// 指标模型更新时间
			promqlQuery.ModelUpdateTime = model.UpdateTime
			if promqlQuery.ModelUpdateTime == 0 {
				promqlQuery.ModelUpdateTime = time.Now().UnixMilli()
			}

			// 视图更新时间
			promqlQuery.DataView.UpdateTime = dataView.UpdateTime
			if promqlQuery.DataView.UpdateTime != 0 {
				promqlQuery.DataView.UpdateTime = time.Now().UnixMilli()
			}

			fieldMap, _, err := mms.promqlService.GetFields(ctx, promqlQuery)
			if err != nil {
				logger.Errorf("Exec promql error: %s", err.Error())
				return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
					uerrors.Uniquery_MetricModel_InternalError_GetFieldsFailed).WithErrorDetails(err.Error())
			}

			// 把fieldMap 映射到模型的原始字段集，用对应的原始字段集输出
			for fieldName := range fieldMap {
				field, exist := model.FieldsMap[interfaces.LABELS_PREFIX+"."+fieldName]
				if exist {
					fields = append(fields, field)
				}
			}

		case qt == interfaces.SQL && dsType == interfaces.QueryType_SQL:
			// 返回指标模型的分组字段和分析维度
			fieldNameMap := map[string]bool{}
			sqlConfig := model.FormulaConfig.(interfaces.SQLConfig)
			for _, field := range sqlConfig.GroupByFields {
				if gField, exist := model.FieldsMap[field]; exist {
					fields = append(fields, gField)
					fieldNameMap[field] = true
				}
				// 不存在就不返回
			}

			for _, field := range model.AnalysisDims {
				// 字段存在且还未被添加过，则添加到返回集中
				if dField, exist := model.FieldsMap[field.Name]; exist && !fieldNameMap[field.Name] {
					fields = append(fields, dField)
					fieldNameMap[field.Name] = true
				}
				// 不存在就不返回
			}

		default:
			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_UnsupportQueryType).
				WithErrorDetails(fmt.Sprintf(`Unsupported query type[%s] in Metric Lables currently`, model.QueryType))
		}
	case interfaces.DERIVED_METRIC, interfaces.COMPOSITED_METRIC:
		// 衍生复合指标的字段列表就是分析维度的字段列表。
		// 需要补上依赖指标的分组字段。
		for _, field := range model.AnalysisDims {
			if f, exist := model.FieldsMap[field.Name]; exist {
				fields = append(fields, f)
			}
		}
		groupFields, err := mms.getMetricGroupFields(ctx, model)
		if err != nil {
			return nil, err
		}

		fields = append(fields, groupFields...)
	}

	// 对 fields 进行排序，排序字段是 Name 字段
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Name < fields[j].Name
	})

	span.SetStatus(codes.Ok, "")
	return fields, nil
}

// 获取指标模型的指定字段的字段值
func (mms *metricModelService) GetMetricModelFieldValues(ctx context.Context, modelID, fieldName string) (interfaces.FieldValues, error) {
	// 1.获取指标模型的对象信息，include_view为true时包含了视图的字段信息。
	// 获取指标模型信息
	ctx, span := ar_trace.Tracer.Start(ctx, "查询指标模型的字段值列表")
	defer span.End()

	fieldValue := interfaces.FieldValues{}

	// 决策当前模型id的数据查询权限
	err := mms.ps.CheckPermission(ctx, interfaces.Resource{
		ID:   modelID,
		Type: interfaces.RESOURCE_TYPE_METRIC_MODEL,
	}, []string{interfaces.OPERATION_TYPE_DATA_QUERY})
	if err != nil {
		return fieldValue, err
	}

	metricModels, exists, err := mms.mmAccess.GetMetricModel(ctx, modelID)
	if err != nil {
		logger.Errorf("Get Metric Model error: %s", err.Error())

		// 添加异常时的 trace 属性
		span.SetAttributes(attribute.Key("model_id").String(modelID))
		span.SetStatus(codes.Error, "Get Metric Model error")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Get Metric Model error: %v", err))

		return fieldValue, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			uerrors.Uniquery_MetricModel_InternalError_GetModelByIdFailed).WithErrorDetails(err.Error())
	}
	if !exists || len(metricModels) == 0 {
		logger.Debugf("Metric Model %s not found!", modelID)

		// 添加异常时的 trace 属性
		span.SetAttributes(attribute.Key("model_id").String(modelID))
		span.SetStatus(codes.Error, "Metric Model not found!")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Metric Model [%s] not found!", modelID))

		return fieldValue, rest.NewHTTPError(ctx, http.StatusNotFound, uerrors.Uniquery_MetricModel_MetricModelNotFound)
	}

	model := metricModels[0]
	if model.MetricType != interfaces.ATOMIC_METRIC {
		return fieldValue, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_UnsupportMetricType).
			WithErrorDetails(fmt.Sprintf("只支持获取原子指标模型的字段值，当前请求的指标模型类型为%s", model.MetricType))
	}
	// 原子指标需获取视图详情。当前这个接口只支持原子指标获取字段值。
	dataView, err := mms.dvService.GetDataViewByID(ctx, model.DataSource.ID, true)
	if err != nil {
		span.SetStatus(codes.Error, "Get data view by ID failed")
		return fieldValue, err
	}
	// 校验字段是否存在于指标模型的数据源的字段集中
	fieldInfo, exist := dataView.FieldsMap[fieldName]
	if !exist {
		return fieldValue, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_FieldName).
			WithErrorDetails(fmt.Sprintf("指定的字段[%s]不属于数据视图[%s]", fieldName, model.DataSource.ID))
	}
	// 校验字段类型是 keyword 或者 text，如果不是直接返回空
	if !(fieldInfo.Type == data_type.DataType_String || fieldInfo.Type == data_type.DataType_Text) {
		return fieldValue, nil
	}
	fieldValue.FieldName = fieldName
	fieldValue.Type = fieldInfo.Type

	switch qt := model.QueryType; {
	case qt == interfaces.DSL:
		// 2. dsl时构造dsl发起查询，dsl 包含模型的dsl，还需要再拼接上如下条件：
		// a. 计算公式中所有的聚合字段都存在
		// b. 时间范围-无，全索引库搜索
		return mms.processDSLFieldValues(ctx, model, dataView, fieldValue)

	case qt == interfaces.PROMQL:
		// 3. promql时，解析表达式，在叶子节点从tsid中把字段结构拿到，
		// 输出字段列表到上层，对于promql来说，所有字段都是keyword
		promqlQuery := interfaces.Query{
			ModelId:             model.ModelID,
			QueryStr:            model.Formula,
			DataView:            *dataView,
			MaxSearchSeriesSize: mms.appSetting.PromqlSetting.MaxSearchSeriesSize,
		}
		// 指标模型更新时间
		promqlQuery.ModelUpdateTime = model.UpdateTime
		if promqlQuery.ModelUpdateTime == 0 {
			promqlQuery.ModelUpdateTime = time.Now().UnixMilli()
		}

		// 视图更新时间
		promqlQuery.DataView.UpdateTime = dataView.UpdateTime
		if promqlQuery.DataView.UpdateTime != 0 {
			promqlQuery.DataView.UpdateTime = time.Now().UnixMilli()
		}

		valuesMap, _, err := mms.promqlService.GetFieldValues(ctx, promqlQuery, fieldName)
		if err != nil {
			logger.Errorf("Exec promql error: %s", err.Error())
			return fieldValue, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				uerrors.Uniquery_MetricModel_InternalError_GetFieldValuesFailed).WithErrorDetails(err.Error())
		}

		// 把fieldMap 映射到模型的原始字段集，用对应的原始字段集输出
		values := make([]string, 0)
		for value := range valuesMap {
			values = append(values, value)
		}
		fieldValue.Values = values
	default:
		// vega-sql 尚未支持
		return fieldValue, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_UnsupportQueryType).
			WithErrorDetails(fmt.Sprintf(`Unsupported query type[%s] in Metric Lables currently`, model.QueryType))
	}

	// 对 fieldValue中的 Values 进行排序
	sort.Strings(fieldValue.Values)
	span.SetStatus(codes.Ok, "")
	return fieldValue, nil
}

func (mms *metricModelService) processDSLFieldValues(ctx context.Context, model interfaces.MetricModel,
	dataView *interfaces.DataView, fieldValue interfaces.FieldValues) (interfaces.FieldValues, error) {
	// 2. dsl时构造dsl发起查询，dsl 包含模型的dsl，还需要再拼接上如下条件：
	// a. 计算公式中所有的聚合字段都存在
	// b. 时间范围-无，全索引库搜索
	// 2.1 解析dsl
	query := &interfaces.MetricModelQuery{
		MetricModelID: model.ModelID,
		ModelName:     model.ModelName,
		Formula:       model.Formula,
		MeasureField:  model.MeasureField,
		// DataView:      dataView,
		QueryType: model.QueryType,
	}
	dslInfo, err := parseDsl(ctx, query, dataView)
	if err != nil {
		return fieldValue, err
	}

	// 2.3 构造 aggs: terms(请求字段)
	field := fieldValue.FieldName
	if fieldValue.Type == data_type.DataType_Text {
		field = fieldValue.FieldName + "." + interfaces.KEYWORD_SUFFIX
	}

	aggName := "fieldV"
	aggs := map[string]any{
		aggName: map[string]any{
			"terms": map[string]any{
				"field": field,
				"size":  mms.appSetting.PromqlSetting.MaxSearchSeriesSize,
				"order": map[string]string{
					"_key": "asc",
				},
			},
		},
	}

	// 2.2 追加过滤条件: 聚合字段都存在
	filtersArr := make([]any, 0)
	for _, aggInfo := range dslInfo.AggInfos {
		if aggInfo.TermsField != "" {
			existQuery := map[string]any{
				"exists": map[string]any{
					"field": aggInfo.TermsField,
				},
			}
			filtersArr = append(filtersArr, existQuery)
		}
	}

	// 时间小于now
	filtersArr = append(filtersArr, map[string]any{
		"range": map[string]any{
			"@timestamp": map[string]any{
				"lte": time.Now().UnixMilli(),
			},
		},
	})

	// 2.4 分批次请求opensearch
	baseTypes, _, err := data_view.GetBaseTypes(dataView)
	if err != nil {
		return fieldValue, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
	}
	indexPattern := convert.GetIndexBasePattern(baseTypes)

	// 每次遍历，都要改变fieldName的起始位置，用前一个请求的最后一个值作为下一个请求的起始值
	values := make([]string, 0)
	beginValue := ""
	condition := true
	for condition {
		// 1. 起始值的fieldName的range过滤条件
		curFilters := make([]any, 0)
		curFilters = append(curFilters, filtersArr...)
		if beginValue != "" {
			fieldRange := map[string]any{
				"range": map[string]any{
					field: map[string]any{
						"gt": beginValue,
					},
				},
			}
			curFilters = append(curFilters, fieldRange)
		}

		var dslQuery map[string]any
		err = sonic.Unmarshal(dslInfo.DSLQuery, &dslQuery)
		if err != nil {
			return fieldValue, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_MetricModel_InternalError_UnmarshalFailed).
				WithErrorDetails(fmt.Sprintf("Dsl Query Unmarshal error: %s", err.Error()))
		}

		dsl := map[string]any{
			"size":  0,
			"query": dslQuery,
			"aggs":  aggs,
		}

		rewriteDslQuery(dsl, curFilters)

		str, _ := sonic.MarshalString(dsl)
		logger.Debugf("模型【%s】分批获取序列生成的dsl; %s", query.ModelName, str)

		// 2. 查全部,用indexPattern来查
		res, _, err := mms.dvService.GetDataFromOpenSearch(ctx, dsl, indexPattern,
			interfaces.DEFAULT_SCROLL, interfaces.DEFAULT_PREFERENCE, interfaces.DEFAULT_TRACK_TOTAL_HITS)
		if err != nil {
			return fieldValue, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_InternalError).
				WithErrorDetails(err.Error())
		}

		// 3. 解析结果，从中读取tsid和labels。只能顺序做，因为这一批的最后一个是下一批的过滤条件
		jsons := string(res)
		buckets := gjson.Get(jsons, fmt.Sprintf("aggregations.%s.buckets", aggName)).Array()
		for _, bucket := range buckets {
			fieldV := bucket.Get("key").String()
			values = append(values, fieldV)
		}
		// 当前批次的最后一个tsid作为下次tsid的起始点(大于 offsetTsid)
		if len(values) > 0 {
			beginValue = values[len(values)-1]
		}

		if len(buckets) == 0 || query.MaxSearchSeriesSize == 1 {
			condition = false
		}
	}
	fieldValue.Values = values

	return fieldValue, nil
}

// 获取指标模型的维度字段集(结果字段)
func (mms *metricModelService) GetMetricModelLabels(ctx context.Context, modelID string) ([]*cond.ViewField, error) {
	// 1.获取指标模型的对象信息，include_view为true时包含了视图的字段信息。
	// 获取指标模型信息
	ctx, span := ar_trace.Tracer.Start(ctx, "查询指标模型的维度字段列表")
	defer span.End()

	// 决策当前模型id的数据查询权限
	err := mms.ps.CheckPermission(ctx, interfaces.Resource{
		ID:   modelID,
		Type: interfaces.RESOURCE_TYPE_METRIC_MODEL,
	}, []string{interfaces.OPERATION_TYPE_DATA_QUERY})
	if err != nil {
		return nil, err
	}

	metricModels, exists, err := mms.mmAccess.GetMetricModel(ctx, modelID)
	if err != nil {
		logger.Errorf("Get Metric Model error: %s", err.Error())

		// 添加异常时的 trace 属性
		span.SetAttributes(attribute.Key("model_id").String(modelID))
		span.SetStatus(codes.Error, "Get Metric Model error")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Get Metric Model error: %v", err))

		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			uerrors.Uniquery_MetricModel_InternalError_GetModelByIdFailed).WithErrorDetails(err.Error())
	}
	if !exists || len(metricModels) == 0 {
		logger.Debugf("Metric Model %s not found!", modelID)

		// 添加异常时的 trace 属性
		span.SetAttributes(attribute.Key("model_id").String(modelID))
		span.SetStatus(codes.Error, "Metric Model not found!")

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Metric Model [%s] not found!", modelID))

		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, uerrors.Uniquery_MetricModel_MetricModelNotFound)
	}
	model := metricModels[0]
	// 原子指标需获取视图详情。当前这个接口只支持原子指标获取字段值。
	dataView, err := mms.dvService.GetDataViewByID(ctx, model.DataSource.ID, true)
	if err != nil {
		span.SetStatus(codes.Error, "Get data view by ID failed")
		return nil, err
	}

	fields := make([]*cond.ViewField, 0)
	switch qt := model.QueryType; {
	case qt == interfaces.DSL:
		// 2. dsl时解析出bucket aggs的聚合名称，作为输出列表，对于top_hits，需把度量字段之外的字段也添加上
		return mms.getDSLModelLabels(ctx, model, *dataView)

	case qt == interfaces.PROMQL:
		// 3. promql时，解析表达式，在叶子节点从tsid中把字段结构拿到，
		// 输出字段列表到上层,上层各自做各自的计算，对于promql来说，所有字段都是keyword

		promqlQuery := interfaces.Query{
			ModelId:             model.ModelID,
			QueryStr:            model.Formula,
			DataView:            *dataView,
			MaxSearchSeriesSize: mms.appSetting.PromqlSetting.MaxSearchSeriesSize,
		}
		// 指标模型更新时间
		promqlQuery.ModelUpdateTime = model.UpdateTime
		if promqlQuery.ModelUpdateTime == 0 {
			promqlQuery.ModelUpdateTime = time.Now().UnixMilli()
		}

		// 视图更新时间
		promqlQuery.DataView.UpdateTime = dataView.UpdateTime
		if promqlQuery.DataView.UpdateTime != 0 {
			promqlQuery.DataView.UpdateTime = time.Now().UnixMilli()
		}

		fieldMap, _, err := mms.promqlService.GetLabels(ctx, promqlQuery)
		if err != nil {
			logger.Errorf("Exec promql error: %s", err.Error())
			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				uerrors.Uniquery_MetricModel_InternalError_GetLabelsFailed).WithErrorDetails(err.Error())
		}

		// 把fieldMap 映射到模型的原始字段集，用对应的原始字段集输出
		for fieldName := range fieldMap {
			fields = append(fields, &cond.ViewField{
				Name: fieldName,
				Type: data_type.DataType_String,
			})
		}
	default:
		// vega-sql 尚未支持
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_UnsupportQueryType).
			WithErrorDetails(fmt.Sprintf(`Unsupported query type[%s] in Metric Lables currently`, model.QueryType))
	}
	// 对 fields 进行排序，排序字段是 Name 字段
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Name < fields[j].Name
	})

	span.SetStatus(codes.Ok, "")
	return fields, nil
}

func (mms *metricModelService) getDSLModelLabels(ctx context.Context, model interfaces.MetricModel,
	dataView interfaces.DataView) ([]*cond.ViewField, error) {

	fields := make([]*cond.ViewField, 0)
	query := &interfaces.MetricModelQuery{
		MetricModelID: model.ModelID,
		ModelName:     model.ModelName,
		Formula:       model.Formula,
		MeasureField:  model.MeasureField,
		// DataView:      dataView,
		QueryType: model.QueryType,
	}
	dslInfo, err := parseDsl(ctx, query, &dataView)
	if err != nil {
		return fields, err
	}
	// 维度字段的类型是聚合使用的字段的类型
	for _, aggInfo := range dslInfo.AggInfos {
		switch aggInfo.AggType {
		case interfaces.BUCKET_TYPE_TERMS, interfaces.BUCKET_TYPE_RANGE, interfaces.BUCKET_TYPE_DATE_RANGE:
			originalField, exist := dataView.FieldsMap[aggInfo.TermsField]
			if exist {
				// 不存在就不添加到维度字段列表中
				fields = append(fields, &cond.ViewField{
					Name: aggInfo.AggName,
					Type: originalField.Type,
				})
			}
		case interfaces.BUCKET_TYPE_FILTERS:
			fields = append(fields, &cond.ViewField{
				Name: aggInfo.AggName,
				Type: data_type.DataType_String,
			})
		case interfaces.AGGR_TYPE_TOP_HITS:
			for _, field := range aggInfo.IncludeFields {
				originalField, exist := dataView.FieldsMap[field]
				if exist {
					// 不存在就不添加到维度字段列表中
					fields = append(fields, &cond.ViewField{
						Name: field,
						Type: originalField.Type,
					})
				}
			}
		}
	}
	return fields, nil
}

// 按序列构建指标数据的映射
func processMetricDataTimeSeries(metricData interfaces.MetricModelUniResponse) (map[string]interfaces.MetricModelData, error) {

	// 每个序列的维度不尽相同,所以每次构建序列标识的时候,对labels的map的key进行排序再拼接.

	// 处理成序列之和,对空缺的步长点补null
	// 时间修正, 同期值和本期值不同,同期值需要把时间置到同期的时间轴上
	// fixedStart, fixedEnd := correctingTime(query, common.APP_LOCATION)
	// // 如果是同期数据,需要把start和end对应到同期的start end
	// if isSamePeriod {
	// 	fixedStart = calcComparisonTime(time.UnixMilli(fixedStart).In(common.APP_LOCATION),
	// 		*query.RequestMetrics.SamePeriodCfg).UnixMilli()
	// 	fixedEnd = calcComparisonTime(time.UnixMilli(fixedEnd).In(common.APP_LOCATION),
	// 		*query.RequestMetrics.SamePeriodCfg).UnixMilli()
	// }
	// // 生成完整的时间点
	// allTimes := make([]any, 0)
	// allTimeStrs := make([]string, 0)
	// for currentTime := fixedStart; currentTime <= fixedEnd; {
	// 	allTimes = append(allTimes, currentTime)
	// 	allTimeStrs = append(allTimeStrs, convert.FormatTimeMiliis(currentTime, query.StepStr))
	// 	// 格式化时间
	// 	currentTime = getNextPointTime(query, currentTime)
	// }

	// 使用map来按维度分组数据
	seriesMap := make(map[string]interfaces.MetricModelData)
	for _, seriesi := range metricData.Datas {
		// 构建labels的key
		var key string
		labelNames := []string{}
		for ln := range seriesi.Labels {
			labelNames = append(labelNames, ln)
		}
		for _, ln := range labelNames {
			labelValue := fmt.Sprintf("%s=%s", ln, seriesi.Labels[ln])
			key += labelValue + ","
		}

		// 获取或创建TimeSeries
		_, exists := seriesMap[key]
		if !exists {
			seriesMap[key] = seriesi
		}
	}

	return seriesMap, nil
}

// 获取指标模型的映射视图的原始字段的字段集
func (mms *metricModelService) getMetricGroupFields(ctx context.Context,
	model interfaces.MetricModel) ([]interfaces.Field, error) {

	switch model.MetricType {
	case interfaces.ATOMIC_METRIC:
		return mms.getAtomicGroupFields(model)
	case interfaces.DERIVED_METRIC:
		return mms.getDerivedGroupFields(ctx, model)
	case interfaces.COMPOSITED_METRIC:
		return mms.getCompositeGroupFields(ctx, model)
	default:
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest,
			uerrors.Uniquery_MetricModel_UnsupportMetricType)
	}
}

// 原子指标：视图的字段集
func (mms *metricModelService) getAtomicGroupFields(
	model interfaces.MetricModel) ([]interfaces.Field, error) {

	// sql-原子：分组字段
	groupFields := []interfaces.Field{}
	if model.DataSource.Type == interfaces.QueryType_SQL {
		// 分组字段
		for _, groupBy := range model.FormulaConfig.(interfaces.SQLConfig).GroupByFieldsDetail {
			if f, exist := model.FieldsMap[groupBy.Name]; exist {
				groupFields = append(groupFields, f)
			}
		}
	}

	// 原子指标的分组字段
	return groupFields, nil
}

// 衍生指标：其原子对应的视图的字段集
func (mms *metricModelService) getDerivedGroupFields(ctx context.Context,
	model interfaces.MetricModel) ([]interfaces.Field, error) {

	derivedModel := model.FormulaConfig.(interfaces.DerivedConfig)
	// 先获取依赖的原子指标模型
	dependMetricModels, exists, err := mms.mmAccess.GetMetricModel(ctx, derivedModel.DependMetricModel.ID)
	if err != nil {
		logger.Errorf("Get Metric Model error: %s", err.Error())

		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			uerrors.Uniquery_MetricModel_InternalError_GetModelByIdFailed).WithErrorDetails(err.Error())
	}
	if !exists || len(dependMetricModels) == 0 {
		logger.Errorf("Depend Metric Model %s not found!", derivedModel.DependMetricModel.ID)
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, uerrors.Uniquery_MetricModel_MetricModelNotFound)
	}

	// 衍生指标的原始字段集是其依赖的原子指标的字段集
	return mms.getMetricGroupFields(ctx, dependMetricModels[0])
}

// 复合指标：其依赖的指标模型的视图字段集的
func (mms *metricModelService) getCompositeGroupFields(ctx context.Context,
	model interfaces.MetricModel) ([]interfaces.Field, error) {

	groupFields := []interfaces.Field{}
	groupMap := map[string]interfaces.Field{}

	modelIDs := convert.ExtractModelIDs(model.Formula)
	// 获取所有依赖指标的原始字段集
	for _, dep := range modelIDs {
		dependModel, exist, err := mms.mmAccess.GetMetricModel(ctx, dep)
		if err != nil {
			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				uerrors.Uniquery_MetricModel_InternalError_GetModelByIdFailed).WithErrorDetails(err.Error())

		}
		if !exist {
			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_MetricModelNotFound).
				WithErrorDetails(fmt.Sprintf("The Depend metric model[%s] of composite metric model with id [%s] not exists!", dep, model.ModelID))
		}
		// 各个依赖指标的分组字段
		fieldsi, err := mms.getMetricGroupFields(ctx, dependModel[0])
		if err != nil {
			return nil, err
		}
		for _, fi := range fieldsi {
			groupMap[fi.Name] = fi
		}
	}

	for _, v := range groupMap {
		groupFields = append(groupFields, v)
	}
	return groupFields, nil
}
