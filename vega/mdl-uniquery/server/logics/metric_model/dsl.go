// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package metric_model

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/tidwall/gjson"
	"go.opentelemetry.io/otel/codes"

	"uniquery/common"
	cond "uniquery/common/condition"
	"uniquery/common/convert"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	"uniquery/logics/promql/leafnodes"
)

// 解析 dsl 查询语句
func parseDsl(ctx context.Context, query *interfaces.MetricModelQuery, dataView *interfaces.DataView) (interfaces.DslInfo, error) {

	// 忽略缓存时不走缓存
	// if !query.IgnoringMemoryCache {
	// 	v, ok := Dsl_Info_Of_Model.Load(query.MetricModelID)
	// 	if ok {
	// 		dslInfoCache := v.(DSLInfoCache)
	// 		// 加数据视图的更新时间
	// 		// 若配置信息的缓存刷新时间大于模型更新时间 && 缓存刷新时间大于数据视图的更新时间 && 缓存时间距离当前小于10min
	// 		// && api的过滤条件为空，则返回缓存。否则实时查并刷新缓存
	// 		diff := time.Since(dslInfoCache.RefreshTime)
	// 		if dslInfoCache.RefreshTime.After(query.ModelUpdateTime) &&
	// 			dslInfoCache.RefreshTime.After(query.DataView.ViewUpdateTime) &&
	// 			diff <= 10*time.Minute && len(query.Filters) == 0 {
	// 			return dslInfoCache.DslInfo, nil
	// 		}
	// 	}
	// }

	dslInfo := interfaces.DslInfo{}
	// 1. 把 dsl 语句转成 json，结构为 map[string]interfaces{}
	var dsl map[string]any
	err := sonic.Unmarshal([]byte(query.Formula), &dsl)
	if err != nil {
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("dsl Unmarshal error: %s", err.Error()))
		return dslInfo, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_MetricModel_InternalError_UnmarshalFailed).
			WithErrorDetails(fmt.Sprintf("dsl Unmarshal error: %s", err.Error()))
	}

	// 2. 递归读取 json
	if _, exist := dsl["aggs"]; !exist {
		// 记录异常日志
		o11y.Error(ctx, "dsl missing aggregation.")
		return dslInfo, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Formula).
			WithErrorDetails("dsl missing aggregation.")
	}
	aggs, ok := dsl["aggs"].(map[string]interface{})
	if !ok {
		// 记录异常日志
		o11y.Error(ctx, "The aggregation of dsl is not a map")
		return dslInfo, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Formula).
			WithErrorDetails("The aggregation of dsl is not a map")
	}
	// 3. 解析 aggs,分两份,instant query 和 range query
	// 解析instant query 的 aggs
	aggBytes, err := sonic.Marshal(aggs)
	if err != nil {
		return dslInfo, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_MetricModel_InternalError_MarshalFailed).
			WithErrorDetails(fmt.Sprintf("DSL query Marshal error: %s", err.Error()))
	}
	var instantAggs map[string]any
	err = sonic.Unmarshal(aggBytes, &instantAggs)
	if err != nil {
		return dslInfo, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_MetricModel_InternalError_UnmarshalFailed).
			WithErrorDetails(fmt.Sprintf("Dsl Query Unmarshal error: %s", err.Error()))
	}
	aggInfos2 := make(map[int]interfaces.AggInfo, 0)
	err = parseAggs(ctx, instantAggs, aggInfos2, query, true, 0)
	if err != nil {
		return dslInfo, err
	}
	// 解析range query 的aggs
	aggInfos := make(map[int]interfaces.AggInfo, 0)
	err = parseAggs(ctx, aggs, aggInfos, query, false, 0)
	if err != nil {
		return dslInfo, err
	}
	dslInfo.AggInfos = aggInfos

	// 4. 读取terms聚合,从 aggInfo中把bucket aggs 抽出来
	termsInfos := make([]interfaces.AggInfo, 0)
	termsToAggs := make([]int, 0) // 记录terms在原有dsl aggs中的位置
	targetSeriesNum := int64(1)   // 记录分桶聚合的配置的size
	notTermsSeriesNum := int64(1) // 记录非terms分桶的大小
	var dateHistogram interfaces.AggInfo
	for i := 1; i < len(aggInfos)+1; i++ {
		switch aggInfos[i].AggType {
		case interfaces.BUCKET_TYPE_TERMS:
			targetSeriesNum = targetSeriesNum * aggInfos[i].ConfigSize
			termsInfos = append(termsInfos, aggInfos[i])
			termsToAggs = append(termsToAggs, i)
		case interfaces.BUCKET_TYPE_FILTERS, interfaces.BUCKET_TYPE_RANGE, interfaces.BUCKET_TYPE_DATE_RANGE:
			targetSeriesNum = targetSeriesNum * aggInfos[i].ConfigSize
			notTermsSeriesNum = notTermsSeriesNum * aggInfos[i].ConfigSize
		case interfaces.BUCKET_TYPE_DATE_HISTOGRAM:
			dateHistogram = aggInfos[i]
		}
	}
	dslInfo.TermsInfos = termsInfos
	dslInfo.TermsToAggs = termsToAggs
	dslInfo.BucketSeriesNum = targetSeriesNum
	dslInfo.NotTermsSeriesNum = notTermsSeriesNum
	dslInfo.DateHistogram = dateHistogram

	// 重写filters
	err = rewriteDSLFilters(ctx, query, dslInfo, dataView)
	if err != nil {
		return dslInfo, err
	}

	// 6. 拼接数据视图的过滤条件+外部请求的过滤条件.不包含时间的过滤条件
	err = ParseQuery(ctx, query, dsl, dataView)
	if err != nil {
		return dslInfo, err
	}
	dslInfo.RangeQueryDSL = dsl
	dslInfo.InstantQueryDSL = map[string]any{
		"size":  0,
		"query": dsl["query"],
		"aggs":  instantAggs,
	}

	queryBytes, err := sonic.Marshal(dslInfo.RangeQueryDSL["query"])
	if err != nil {
		return dslInfo, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_MetricModel_InternalError_MarshalFailed).
			WithErrorDetails(fmt.Sprintf("DSL query Marshal error: %s", err.Error()))
	}
	dslInfo.DSLQuery = queryBytes

	// 写入缓存
	// 预览时不缓存，若 filters 不为空，则实时查询，结果不缓存。否则刷新缓存
	// if query.MetricModelID != 0 && len(query.Filters) == 0 {
	// 	Dsl_Info_Of_Model.Store(query.MetricModelID, DSLInfoCache{RefreshTime: time.Now(), DslInfo: dslInfo})
	// }
	return dslInfo, nil
}

// parseAggs
// @Description: 逐层解析子聚合（前面已经对时间字段做了校验，解析时无需再校验，直接解析读取即可）
// @param ctx
// @param aggs: dsl 中 "aggs" 对应的值内容
// @param aggInfos: 表示各层聚合的聚合信息
// @param query: 基于指标模型查询的请求参数
// @param configDelta: 系统配置的即时查询的查询的回退时间，与 promql 复用配置
// @return error
func parseAggs(ctx context.Context, aggs map[string]any, aggInfos map[int]interfaces.AggInfo,
	query *interfaces.MetricModelQuery, isInstant bool, aggsLayers int) error {

	if len(aggs) != 1 {
		// 并行聚合不支持
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Multiple aggregation is not supported, aggs is %v", aggs))

		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Formula).
			WithErrorDetails("Multiple aggregation is not supported")
	}
	// aggsDetail 的 key 是聚合名称，value 是个定义了聚合的 map
	for aggName, value := range aggs {
		// key 是聚合名称
		// value 是个map，map 中有两个元素，一个 key 是 aggType，一个 key 是 aggs 或者 aggregations
		aggValues, ok := value.(map[string]any)
		if !ok {
			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("The aggregation of dsl is not a map, aggs is %v", value))

			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Formula).
				WithErrorDetails("The aggregation of dsl is not a map")
		}

		aggsLayers++
		for aggType, aggValue := range aggValues {
			switch aggType {
			case interfaces.AGGS:
				// 递归遍历子聚合
				aggValueMap, ok := aggValue.(map[string]any)
				if !ok {
					// 记录异常日志
					o11y.Error(ctx, fmt.Sprintf("The sub-aggregation of dsl is not a map, aggs is %v", aggValue))

					return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Formula).
						WithErrorDetails("The sub-aggregation of dsl is not a map")
				}
				err := parseAggs(ctx, aggValueMap, aggInfos, query, isInstant, aggsLayers)
				if err != nil {
					return err
				}
			case interfaces.BUCKET_TYPE_TERMS:
				// 分桶聚合类型：terms（词条），记录排序方式和size。
				aggValueMap, _ := aggValue.(map[string]any)
				var termSize int64
				// termSize := int64(-1) // 默认查全部序列,给int32的最大值(这也是opensearch支持的最大值)
				if size, exist := aggValueMap["size"]; exist {
					termSize = int64(size.(float64))
				} else {
					termSize = math.MaxInt32
					aggValueMap["size"] = math.MaxInt32
				}

				// 默认按_key 升序排序.这些信息是为了分批查询序列时使用
				// 按子聚合排序时,当前指标模型不支持这样的查询,暂不考虑.
				sort := interfaces.DSL_TERMS_ORDER_BY_KEY
				direaction := interfaces.ASC_DIRECTION
				if order, exist := aggValueMap["order"]; exist {
					for k, v := range order.(map[string]any) {
						sort = k
						direaction = v.(string)
						// __count的排序方式则修改为按key正序.若未触发高基,则按key正序获取数据,若未触发高基,则还是按原方式进行.
						if k == interfaces.DSL_TERMS_ORDER_BY_COUNT {
							sort = interfaces.DSL_TERMS_ORDER_BY_KEY
							direaction = interfaces.ASC_DIRECTION
						}
					}
				}
				// 把terms除了aggs的字段外其他都留下.便于后续获取terms的序列
				termsInfo := interfaces.AggInfo{
					AggName:    aggName,
					AggType:    aggType,
					TermsField: aggValueMap["field"].(string),
					ConfigSize: termSize,
					EvalSize:   termSize, // 默认是termSize,若是配置的size大于基数,那么就用基数作为size
					Sort:       sort,
					Direction:  direaction,
				}
				aggInfos[aggsLayers] = termsInfo
			case interfaces.BUCKET_TYPE_FILTERS:
				// 分桶聚合类型： date_histogram（日期直方图）、filters（过滤）、range（范围）、date_range（日期范围）、multi_terms（多词条）
				//  还需记录分桶的个数
				filtersMap, _ := aggValue.(map[string]any)["filters"].(map[string]any)
				aggInfos[aggsLayers] = interfaces.AggInfo{
					AggName:    aggName,
					AggType:    aggType,
					ConfigSize: int64(len(filtersMap)),
					EvalSize:   int64(len(filtersMap)),
				}
			case interfaces.BUCKET_TYPE_RANGE, interfaces.BUCKET_TYPE_DATE_RANGE:
				// 分桶聚合类型： date_histogram（日期直方图）、filters（过滤）、range（范围）、date_range（日期范围）、multi_terms（多词条）
				//   还需记录分桶的个数
				ranges, _ := aggValue.(map[string]any)["ranges"].([]any)
				aggInfos[aggsLayers] = interfaces.AggInfo{
					AggName:    aggName,
					AggType:    aggType,
					TermsField: aggValue.(map[string]any)["field"].(string),
					ConfigSize: int64(len(ranges)),
					EvalSize:   int64(len(ranges)),
				}
			// case interfaces.MULTI_TERMS:
			// 	termsFields, ok := aggValue.(map[string]any)["terms"].([]map[string]string)
			// 	if !ok {
			// 		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Formula).
			// 			WithErrorDetails("The terms of multi_terms aggregation should be an array")
			// 	}
			// 	fields := make([]string, 0)
			// 	for _, term := range termsFields {
			// 		fields = append(fields, term["field"])
			// 	}
			// 	aggInfos[aggsLayers] = interfaces.AggInfo{
			// 		AggName:     aggName,
			// 		AggType:     aggType,
			// 		TermsFields: fields,
			// 	}
			case interfaces.BUCKET_TYPE_DATE_HISTOGRAM:
				dateHisto := aggValue.(map[string]any)
				intervalType := ""
				intervalValue := ""
				if value, exists := dateHisto["fixed_interval"]; exists {
					intervalType = interfaces.INTERVAL_TYPE_FIXED
					intervalValue = value.(string)
				} else if value, exists := dateHisto["calendar_interval"]; exists {
					intervalType = interfaces.INTERVAL_TYPE_CALENDAR
					intervalValue = value.(string)
				}

				// 解析时区
				timeZone := ""
				if zone, exists := dateHisto["time_zone"]; exists {
					timeZone = zone.(string)
				} else {
					timeZone = os.Getenv("TZ")
					dateHisto["time_zone"] = timeZone
				}
				queryZone, err := time.LoadLocation(timeZone)
				if err != nil {
					// 记录异常日志
					o11y.Error(ctx, fmt.Sprintf("LoadLocation error: %s", err.Error()))

					return rest.NewHTTPError(ctx, http.StatusInternalServerError,
						uerrors.Uniquery_MetricModel_InvaliEnvironmentVariable_TZ).WithErrorDetails(err.Error())
				}

				aggInfos[aggsLayers] = interfaces.AggInfo{
					AggName:       aggName,
					AggType:       aggType,
					IsDateField:   true,
					IntervalType:  intervalType,
					IntervalValue: intervalValue,
					ZoneLocation:  queryZone,
				}
				if isInstant {
					// 瞬时查询时，应该是去掉日期直方图的聚合
					tmp := aggValues[interfaces.AGGS].(map[string]any)
					delete(aggs, aggInfos[aggsLayers].AggName)
					// 把 tmp 复制到 aggs 中
					for k, v := range tmp {
						aggs[k] = v
					}
					// 记录日志
					o11y.Debug(ctx, fmt.Sprintf("Instant Query, delete date_histogram, final aggs is: %v", aggs))
				}

			case interfaces.AGGR_TYPE_VALUE_COUNT, interfaces.AGGR_TYPE_CARDINALITY,
				interfaces.AGGR_TYPE_SUM, interfaces.AGGR_TYPE_AVG,
				interfaces.AGGR_TYPE_MAX, interfaces.AGGR_TYPE_MIN:

				// 值聚合解析
				aggInfos[aggsLayers] = interfaces.AggInfo{
					AggName: aggName,
					AggType: aggType,
				}
			case interfaces.AGGR_TYPE_TOP_HITS:
				// top_hits 解析
				var topHit interfaces.TopHits
				topAgg, err := sonic.Marshal(aggValue)
				if err != nil {
					// 记录异常日志
					o11y.Error(ctx, fmt.Sprintf("TopHits marshal error: %s", err))

					return rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_MetricModel_InternalError_MarshalFailed).
						WithErrorDetails(fmt.Sprintf("TopHits marshal error: %s", err.Error()))
				}

				err = sonic.Unmarshal(topAgg, &topHit)
				if err != nil {
					// 记录异常日志
					o11y.Error(ctx, fmt.Sprintf("TopHits Unmarshal error: %s", err))

					return rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_MetricModel_InternalError_UnmarshalFailed).
						WithErrorDetails(fmt.Sprintf("TopHits Unmarshal error: %s", err.Error()))
				}

				includes := make([]string, 0)
				for _, field := range topHit.Source.Includes {
					if field != query.MeasureField {
						includes = append(includes, field)
					}
				}

				aggInfos[aggsLayers] = interfaces.AggInfo{
					AggName:       aggName,
					AggType:       aggType,
					IncludeFields: includes,
				}
			}
		}
	}

	return nil
}

// 拼接数据视图的过滤条件+外部请求的过滤条件+计算公式的过滤条件.不包含时间的过滤条件
func ParseQuery(ctx context.Context, query *interfaces.MetricModelQuery, dsl map[string]any, dataView *interfaces.DataView) error {
	// 如果是基于指标模型的查询，则需把过滤器 filters 的过滤条件拼接到 dsl 的 query 中
	// 接口请求的filter
	// filters的最后一个字符考虑不添加
	filterStr, _, err := leafnodes.AppendFilters(interfaces.Query{
		IsMetricModel: true,
		DataView:      *dataView,
		Filters:       query.Filters,
	})
	if err != nil {
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Append filters {%v} error: %v", query.Filters, err))

		return rest.NewHTTPError(ctx, http.StatusBadRequest,
			uerrors.Uniquery_UnsupportFilterOperation).WithErrorDetails(err.Error())
	}
	// 去掉末尾的逗号
	if len(filterStr) > 0 {
		filterStr = filterStr[:len(filterStr)-1]
	}

	// 将视图的过滤条件转成dsl, 暂时注释，原子视图没有过滤条件，自定义视图时再考虑
	// var dslStr string
	// if query.DataView.Condition != nil {
	// 	cfg := query.DataView.Condition

	// 	// 将过滤条件拼接到 dsl 的 query 中
	// 	CondCfg, err := cond.NewCondition(ctx, cfg, query.DataView.FieldScope, query.DataView.FieldsMap)
	// 	if err != nil {
	// 		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_Filters).
	// 			WithErrorDetails(fmt.Sprintf("New condition failed, %v", err))
	// 	}

	// 	if CondCfg != nil {
	// 		dslStr, err = CondCfg.Convert(ctx)
	// 		if err != nil {
	// 			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_Filters).
	// 				WithErrorDetails(fmt.Sprintf("Convert condition to dsl failed, %v", err))
	// 		}
	// 	}
	// }

	// 构造日志分组的must_filter
	// var must string
	// if query.DataView.LogGroupFilters != "" {
	// 	logf, err := sonic.Marshal(query.DataView.LogGroupFilters)
	// 	if err != nil {
	// 		return rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_MetricModel_InternalError_MarshalFailed).
	// 			WithErrorDetails(fmt.Sprintf("query.DataView.LogGroupFilters Marshal error: %s", err.Error()))

	// 	}
	// 	must = fmt.Sprintf(`
	// 	{
	// 		"query_string": {
	// 			"query": %s,
	// 			"analyze_wildcard": true
	// 		}
	// 	}`, logf)
	// }

	dslStr := query.ViewQuery4Metric.QueryStr
	// filters不为空， dsl或must不为空，需保留逗号，其余情况雀鲷filters末尾的逗号
	if len(filterStr) > 0 && dslStr != "" {
		filterStr += ","
	}

	filtersStr := fmt.Sprintf(`[%s
		            %s
				]`, filterStr, dslStr)

	// filtersStr := fmt.Sprintf(`[%s]`, filterStr)

	// 构造 time range
	filtersArr := make([]any, 0)
	err = sonic.Unmarshal([]byte(filtersStr), &filtersArr)
	if err != nil {
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Filters of Config(contain metric model filters, data_view's filters)  Unmarshal error: %s", err.Error()))

		return rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_MetricModel_InternalError_UnmarshalFailed).
			WithErrorDetails(fmt.Sprintf("Filters of Config(contain metric model filters, data_view's filters) Unmarshal error: %s", err.Error()))
	}

	if query.QueryType == interfaces.DSL_CONFIG && query.FormulaConfig != nil {
		if query.FormulaConfig.(interfaces.MetricModelFormulaConfig).QueryString != "" {
			must := map[string]any{
				"query_string": map[string]any{
					"query":            query.FormulaConfig.(interfaces.MetricModelFormulaConfig).QueryString,
					"analyze_wildcard": true,
				},
			}
			filtersArr = append(filtersArr, must)
		}

		dateField := query.DateField
		if dateField == "" {
			dateField = interfaces.DEFAULT_DATE_FIELD
		}
		// 还需拼接时间过滤
		// 构造 time range
		time := map[string]any{
			"range": map[string]any{
				dateField: map[string]any{
					"gte": query.Start,
					"lt":  query.End,
				},
			},
		}
		filtersArr = append(filtersArr, time)
	}

	rewriteDslQuery(dsl, filtersArr)
	return nil
}

// 重写 dsl 的 query 部分，把过滤条件和时间过滤拼到 dsl 语句中
func rewriteDslQuery(dsl map[string]any, filtersArr []any) {
	dslQuery, exists := dsl["query"].(map[string]any)
	if !exists {
		// 如果没有 query，则把当前的直接拼接上
		dsl["query"] = map[string]any{
			"bool": map[string]any{
				"filter": filtersArr,
			},
		}
	} else {
		// 如果存在，则判断bool是否存在，如果不存在，则说明是单个过滤条件的情况。此时应把其和mustfilters合并到bool中
		boolStr, exists := dslQuery["bool"].(map[string]any)
		if !exists {
			// 把 mustfilters 和 dsl 的请求一起拼到 must 数组中
			// 清空query，然后重新组装
			delete(dsl, "query")
			dsl["query"] = map[string]any{
				"bool": map[string]any{
					"filter": filtersArr,
				},
			}
		} else {
			// 如果 bool 存在，判断是否有filters
			dslFilter, exists := boolStr["filter"]
			if !exists {
				// 如果不存在，塞一个新的filter进去
				boolStr["filter"] = filtersArr
			} else {
				// must 是数组的话，就往数组里append；如果不是数组，则需要把filter转成数组
				dslFilterArr, ok := dslFilter.([]any)
				if ok {
					filtersArr = append(filtersArr, dslFilterArr...)
					boolStr["filter"] = filtersArr
				} else {
					filtersArr = append(filtersArr, dslFilter)
					// 清空query，然后重新组装
					boolStr["filter"] = filtersArr
				}
			}
		}
	}
}

// DSL 的返回结构转换为统一结构
func parseDSLResult2Uniresponse(ctx context.Context, dslRes []byte, dslInfo interfaces.DslInfo,
	query interfaces.MetricModelQuery, model interfaces.MetricModel) (interfaces.MetricModelUniResponse, error) {

	_, span := ar_trace.Tracer.Start(ctx, "Parse dsl result to uniresponse")
	defer span.End()

	// var resp *interfaces.MetricModelUniResponse
	// 递归获取 terms 层次的 agg, 组装 labels map 和 tsValueMap
	json := string(dslRes)
	aggregations := gjson.Get(json, "aggregations")
	labels := make(map[string]string)
	datas := make([]interfaces.MetricModelData, 0)

	// 时间修正
	fixedStart, fixedEnd := correctingTime(query, dslInfo.DateHistogram.ZoneLocation)

	// return mapResult
	IteratorAggs(aggregations, labels, 1, dslInfo.AggInfos, &datas, query, fixedStart, fixedEnd)

	span.SetStatus(codes.Ok, "")
	res := interfaces.MetricModelUniResponse{
		Datas:       datas,
		Step:        query.StepStr,
		IsVariable:  query.IsVariable,
		IsCalendar:  query.IsCalendar,
		SeriesTotal: len(datas),
	}

	return processOrderHaving(res, query, model), nil
}

// 递归各层聚合结果，把 dsl 的查询结果组装成统一格式
func IteratorAggs(aggregations gjson.Result, labels map[string]string, groupByi int,
	aggInfos map[int]interfaces.AggInfo, datas *[]interfaces.MetricModelData, query interfaces.MetricModelQuery,
	fixedStart int64, fixedEnd int64) {

	// 取桶
	switch aggInfos[groupByi].AggType {
	case interfaces.BUCKET_TYPE_TERMS, interfaces.BUCKET_TYPE_RANGE, interfaces.BUCKET_TYPE_DATE_RANGE:
		bucketsi := aggregations.Get(fmt.Sprintf("%s.buckets", aggInfos[groupByi].AggName)).Array()
		for _, bucket := range bucketsi {
			bukectKey := bucket.Get("key").String()
			labels[aggInfos[groupByi].AggName] = bukectKey
			// terms 下得有 date_histogram（因为当前 dsl 作为时序分析，校验时判断了必须含有 date_histogram）
			IteratorAggs(bucket, labels, groupByi+1, aggInfos, datas, query, fixedStart, fixedEnd)
		}
	case interfaces.BUCKET_TYPE_FILTERS:
		bucketsi := aggregations.Get(fmt.Sprintf("%s.buckets", aggInfos[groupByi].AggName)).Map()
		for bukectKey, bucket := range bucketsi {
			labels[aggInfos[groupByi].AggName] = bukectKey
			IteratorAggs(bucket, labels, groupByi+1, aggInfos, datas, query, fixedStart, fixedEnd)
		}
	// case interfaces.MULTI_TERMS:
	//do nothing
	case interfaces.BUCKET_TYPE_DATE_HISTOGRAM:
		// 时间字段是时序
		dateValues := make([]any, 0)
		values := make([]any, 0)

		// top_hits 可能产生多个值，平铺，桶内labels相同的，取第一个。
		valuesTop := make(map[string][]any, 0)
		labelsTopMap := make(map[string]map[string]string, 0)
		dateValuesTop := make(map[string][]any, 0)

		if query.IsInstantQuery {
			// 如果是即时查询，时间桶没有。直接遍历值聚合
			switch aggInfos[groupByi+1].AggType {
			case interfaces.AGGR_TYPE_VALUE_COUNT, interfaces.AGGR_TYPE_CARDINALITY,
				interfaces.AGGR_TYPE_SUM, interfaces.AGGR_TYPE_AVG,
				interfaces.AGGR_TYPE_MAX, interfaces.AGGR_TYPE_MIN:

				dateValues = append(dateValues, query.Time)
				values = append(values, convert.WrapMetricValue(aggregations.Get(fmt.Sprintf("%s.value", aggInfos[groupByi+1].AggName)).Float()))
			case interfaces.AGGR_TYPE_TOP_HITS:
				readTopHits(aggregations, aggInfos[groupByi+1], query, query.Time, labelsTopMap, valuesTop, dateValuesTop, &dateValues, &values)
			}
		} else {
			bucketsi := aggregations.Get(fmt.Sprintf("%s.buckets", aggInfos[groupByi].AggName)).Array()

			bucketsiIndex := 0

			if len(bucketsi) != 0 {
				for currentTime := fixedStart; currentTime <= fixedEnd; {
					for _, bucket := range bucketsi {
						// date_histogram 下是值聚合，前面已经校验。各值聚合读取会有些许不同，需要case by case 读取值
						// 读取值
						dateV := bucket.Get("key").Int()

						// 时间修正可能使得修正的时间大于第一个桶时间
						if currentTime > dateV {
							currentTime = dateV
						}

						if currentTime < dateV {
							for currentTime < dateV {
								dateValues = append(dateValues, currentTime)
								values = append(values, nil)
								currentTime = getNextPointTime(query, currentTime)
							}
						}
						//如果当前时间等于当前点的时间，直接赋值
						if currentTime == dateV {
							switch aggInfos[groupByi+1].AggType {
							case interfaces.AGGR_TYPE_VALUE_COUNT, interfaces.AGGR_TYPE_CARDINALITY,
								interfaces.AGGR_TYPE_SUM, interfaces.AGGR_TYPE_AVG,
								interfaces.AGGR_TYPE_MAX, interfaces.AGGR_TYPE_MIN:

								dateValues = append(dateValues, dateV)
								values = append(values, convert.WrapMetricValue(bucket.Get(fmt.Sprintf("%s.value", aggInfos[groupByi+1].AggName)).Float()))
							case interfaces.AGGR_TYPE_TOP_HITS:
								readTopHits(bucket, aggInfos[groupByi+1], query, dateV, labelsTopMap, valuesTop, dateValuesTop, &dateValues, &values)
							}
							bucketsiIndex++
							currentTime = getNextPointTime(query, currentTime)
						}
						if bucketsiIndex == len(bucketsi) {
							//如果currentTime<fixedEnd，说明后面还需要继续补点
							for currentTime <= fixedEnd {
								dateValues = append(dateValues, currentTime)
								values = append(values, nil)
								currentTime = getNextPointTime(query, currentTime)
							}
						}
					}
				}
			}
		}

		// tophits 因为可能含有维度描述的字段，需要重新组装
		for k, v := range valuesTop {
			labelsi := make(map[string]string, 0)
			for field, value := range labels {
				labelsi[field] = value
			}
			for field, value := range labelsTopMap[k] {
				labelsi[field] = value
			}
			*datas = append(*datas, interfaces.MetricModelData{
				Labels: labelsi,
				Times:  dateValuesTop[k],
				Values: v,
			})
		}
		// 构造 datas，labels是一个map，map 是引用，构造得时候需要把map重新赋值
		if len(values) > 0 {
			labelsi := make(map[string]string, 0)
			for field, value := range labels {
				labelsi[field] = value
			}
			*datas = append(*datas, interfaces.MetricModelData{
				Labels: labelsi,
				Times:  dateValues,
				Values: values,
			})
		}
	}
}

// 读取 top_hits 的聚合结果到统一格式中
func readTopHits(aggregations gjson.Result, aggInfo interfaces.AggInfo, query interfaces.MetricModelQuery, time int64,
	labelsTopMap map[string]map[string]string, valuesTop map[string][]any, dateValuesTop map[string][]any,
	dateValues *[]any, values *[]any) {

	// 一个时间桶内相同的label合成一条，取第一个
	labelsTopMapTi := make(map[string]string, 0)
	hits := aggregations.Get(fmt.Sprintf("%s.hits.hits", aggInfo.AggName)).Array()
	for _, hit := range hits {
		source := hit.Get("_source")

		if len(aggInfo.IncludeFields) > 0 {
			labelsTopKey := ""
			labelsTop := make(map[string]string, 0)
			for _, field := range aggInfo.IncludeFields {
				labelTopi := source.Get(field).String()
				labelsTopKey = labelsTopKey + "|" + labelTopi
				// 构造 top_hits 的labels
				labelsTop[field] = labelTopi
			}
			// 如果 top_hits 的labelKey存在，那么就继续下一个。
			_, exist := labelsTopMapTi[labelsTopKey]
			if exist {
				continue
			}
			// labelsKey 不存在
			labelsTopMapTi[labelsTopKey] = labelsTopKey
			// 记下hits里的相关信息，后续合并用
			labelsTopMap[labelsTopKey] = labelsTop
			// 度量值
			valuesTop[labelsTopKey] = append(valuesTop[labelsTopKey], convert.WrapMetricValue(source.Get(query.MeasureField).Float()))
			dateValuesTop[labelsTopKey] = append(dateValuesTop[labelsTopKey], time)
		} else {
			// 如果包含字段中只有度量字段，则直接拼接第一个到values，不涉及合并序列
			// 度量值
			*dateValues = append(*dateValues, time)
			*values = append(*values, convert.WrapMetricValue(source.Get(query.MeasureField).Float()))
			break
		}
	}
}

// correctingTime 修正开始时间和结束时间，符合opensearch的分桶区间
func correctingTime(query interfaces.MetricModelQuery, zoneLocation *time.Location) (int64, int64) {

	// 将时间戳转换为时间
	startTime := time.UnixMilli(*query.Start)
	endTime := time.UnixMilli(*query.End)

	// 如果是日历间隔，按照日历间隔进行修正时间
	if query.IsCalendar {
		switch *query.StepStr {
		case "minute", "1m":
			// 将秒部分设置为零
			fixStart := startTime.Truncate(time.Minute)
			fixEnd := endTime.Truncate(time.Minute)
			return fixStart.UnixMilli(), fixEnd.UnixMilli()

		case "hour", "1h":
			// 将分钟部分设置为零
			fixStart := startTime.Truncate(time.Hour)
			fixEnd := endTime.Truncate(time.Hour)
			return fixStart.UnixMilli(), fixEnd.UnixMilli()

		case "day", "1d":
			// 将小时、分钟和秒部分设置为零
			year, month, day := startTime.Date()
			fixStart := time.Date(year, month, day, 0, 0, 0, 0, zoneLocation)

			year, month, day = endTime.Date()
			fixEnd := time.Date(year, month, day, 0, 0, 0, 0, zoneLocation)

			return fixStart.UnixMilli(), fixEnd.UnixMilli()

		// 向前找到本周的周一
		case "week", "1w":
			year, month, day := startTime.Date()
			fixStart := time.Date(year, month, day, 0, 0, 0, 0, zoneLocation)

			year, month, day = endTime.Date()
			fixEnd := time.Date(year, month, day, 0, 0, 0, 0, zoneLocation)

			startDay := int(fixStart.Weekday())
			endDay := int(fixEnd.Weekday())
			// 减去天数，得到星期一的日期
			fixStart = fixStart.AddDate(0, 0, -(7+startDay-1)%7)
			fixEnd = fixEnd.AddDate(0, 0, -(7+endDay-1)%7)

			return fixStart.UnixMilli(), fixEnd.UnixMilli()

		case "month", "1M":
			// 将天、小时、分钟和秒部分设置为零
			fixStart := time.Date(startTime.Year(), startTime.Month(), 1, 0, 0, 0, 0, zoneLocation)
			fixEnd := time.Date(endTime.Year(), endTime.Month(), 1, 0, 0, 0, 0, zoneLocation)
			return fixStart.UnixMilli(), fixEnd.UnixMilli()

		// 向前找到奔季度第一天
		case "quarter", "1q":
			// 计算季度（0 表示第一季度，1 表示第二季度...）
			startQuarter := (int(startTime.Month()) - 1) / 3
			endQuarter := (int(endTime.Month()) - 1) / 3
			// 计算季度的第一个月
			startMonth := time.Month(startQuarter*3 + 1)
			endMonth := time.Month(endQuarter*3 + 1)
			// 构建季度的第一天
			startTime = time.Date(startTime.Year(), startMonth, 1, 0, 0, 0, 0, zoneLocation)
			endTime = time.Date(endTime.Year(), endMonth, 1, 0, 0, 0, 0, zoneLocation)

			// 然后将天、小时、分钟和置位领零
			fixStart := time.Date(startTime.Year(), startTime.Month(), 1, 0, 0, 0, 0, zoneLocation)
			fixEnd := time.Date(endTime.Year(), endTime.Month(), 1, 0, 0, 0, 0, zoneLocation)
			return fixStart.UnixMilli(), fixEnd.UnixMilli()

		case "year", "1y":
			// 将月、天、小时、分钟和秒部分设置为零
			fixStart := time.Date(startTime.Year(), time.January, 1, 0, 0, 0, 0, zoneLocation)
			fixEnd := time.Date(endTime.Year(), time.January, 1, 0, 0, 0, 0, zoneLocation)
			return fixStart.UnixMilli(), fixEnd.UnixMilli()

		}
	} else {
		// 把step转化成毫秒，使用stepStr，变量的情况下实际执行的step和请求的step不同
		stepStr := ""
		if query.StepStr != nil {
			stepStr = *query.StepStr
		}
		stepT, _ := convert.ParseDuration(stepStr)
		step := stepT.Milliseconds()

		// 先计算桶，然后再按照时区偏移
		_, offset := startTime.In(zoneLocation).Zone()
		fixedStart := int64(math.Floor(float64(*query.Start+int64(offset*1000))/float64(step)))*step - int64(offset*1000)
		fixedEnd := int64(math.Floor(float64(*query.End+int64(offset*1000))/float64(step)))*step - int64(offset*1000)

		return fixedStart, fixedEnd
	}

	return 0, 0
}

// getNextPointTime 获取下一个时间点
func getNextPointTime(query interfaces.MetricModelQuery, currentTime int64) int64 {
	if query.IsCalendar {
		// 将时间戳转换为时间对象
		switch *query.StepStr {
		case "minute", "1m":
			return currentTime + time.Minute.Milliseconds()
		case "hour", "1h":
			return currentTime + time.Hour.Milliseconds()
		case "day", "1d":
			return currentTime + (time.Hour * 24).Milliseconds()
		case "week", "1w":
			return currentTime + (time.Hour * 24 * 7).Milliseconds()
		case "month", "1M":
			t := time.UnixMilli(currentTime)
			return t.AddDate(0, 1, 0).UnixMilli()
		case "quarter", "1q":
			t := time.UnixMilli(currentTime)
			return t.AddDate(0, 3, 0).UnixMilli()
		case "year", "1y":
			t := time.UnixMilli(currentTime)
			return t.AddDate(1, 0, 0).UnixMilli()
		}
	} else {
		return currentTime + *query.Step
	}

	return 0
}

// 根据模型配置,生成获取terms字段的基数的dsl
func generatCardinalityAggs(termsInfos []interfaces.AggInfo) map[string]any {
	aggs := make(map[string]any)
	termFields := make(map[string]int)
	for _, terms := range termsInfos {
		if _, exists := termFields[terms.TermsField]; exists {
			continue
		}
		termFields[terms.TermsField] = 1
		aggs[terms.TermsField] = map[string]any{
			"cardinality": map[string]any{
				"field":               terms.TermsField,
				"precision_threshold": interfaces.DEFAULT_CARDINALITY_PRECISION_THRESHOLD,
			},
		}
	}
	return aggs
}

// 生成查询opensearch的dsl.
func generateDsl(ctx context.Context, queryByte []byte, aggs map[string]any,
	filters []interfaces.Filter, dataView *interfaces.DataView) (map[string]any, error) {

	// 如果是基于指标模型的查询，则需把过滤器 filters 的过滤条件拼接到 dsl 的 query 中
	filterStr, _, err := leafnodes.AppendFilters(interfaces.Query{
		IsMetricModel: true,
		DataView:      *dataView,
		Filters:       filters,
	})
	if err != nil {
		// 记录异常日志
		// o11y.Error(ctx, fmt.Sprintf("Append filters {%v} error: %v", query.Filters, err))
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_Filter).
			WithErrorDetails(err.Error())
	}
	// 删除最后一个逗号
	if len(filterStr) > 0 {
		filterStr = filterStr[:len(filterStr)-1]
	}

	filtersArr := make([]any, 0)
	err = sonic.Unmarshal([]byte(fmt.Sprintf(`[%s]`, filterStr)), &filtersArr)
	if err != nil {
		// 记录异常日志
		// o11y.Error(ctx, fmt.Sprintf("Filters Unmarshal error: %s", err.Error()))
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_MetricModel_InternalError_UnmarshalFailed).
			WithErrorDetails(fmt.Sprintf("Filters Unmarshal error: %s", err.Error()))
	}

	// 每个批次的series的请求的过滤条件会变更,query是map,所以每次用新的query对象与当前批次的过滤条件合并
	var dslQuery map[string]any
	err = sonic.Unmarshal(queryByte, &dslQuery)
	if err != nil {
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_MetricModel_InternalError_UnmarshalFailed).
			WithErrorDetails(fmt.Sprintf("Dsl Query Unmarshal error: %s", err.Error()))
	}

	dsl := map[string]any{
		"size":  0,
		"query": dslQuery,
		"aggs":  aggs,
	}

	rewriteDslQuery(dsl, filtersArr)
	return dsl, nil
}

// 根据配置信息和分批信息,生成获取序列的dsl的aggs部分
func generateSeriesAggs(dslInfo interfaces.DslInfo, ifLastBacth bool) map[string]any {
	subAggs := make(map[string]any)
	for i := len(dslInfo.TermsInfos) - 1; i >= 0; i-- {
		var size int64
		if ifLastBacth {
			size = dslInfo.TermsInfos[i].LastBatchSize
		} else {
			size = dslInfo.TermsInfos[i].BatchSize
		}

		subAggs = map[string]any{
			dslInfo.TermsInfos[i].AggName: map[string]any{
				"terms": map[string]any{
					"field": dslInfo.TermsInfos[i].TermsField,
					"size":  size,
					"order": map[string]string{
						dslInfo.TermsInfos[i].Sort: dslInfo.TermsInfos[i].Direction,
					},
				},
				"aggs": subAggs,
			},
		}
	}
	return subAggs
}

// 获取当前批的序列的查询条件filter
func getSeriesPagedFilters(series []map[string]string, index int, dslInfo interfaces.DslInfo,
	remainSize int64, filters *[]interfaces.Filter, ifLastBacth *bool, batchNum *int64) {

	if index < 0 || len(series) == 0 {
		return
	}

	if dslInfo.TermsInfos[index].EvalSize > remainSize {
		if *batchNum != 0 && *batchNum%dslInfo.TermsInfos[index].BatchNum == 0 {
			// 当前轮(游标前进1时)的第一批,此时需把1的字段的过滤条件留下,再对下一个字段分批查询
			batchI := *batchNum / dslInfo.TermsInfos[index].BatchNum
			getSeriesPagedFilters(series, index-1, dslInfo,
				remainSize/dslInfo.TermsInfos[index].EvalSize, filters, ifLastBacth, &batchI)
		} else {
			if (*batchNum+1)%dslInfo.TermsInfos[index].BatchNum == 0 {
				// 当前轮的最后一批,需改变聚合的size.
				*ifLastBacth = true
			}

			// 按排序方式拼过滤条件
			if dslInfo.TermsInfos[index].Direction == interfaces.ASC_DIRECTION {
				*filters = append(*filters, interfaces.Filter{
					Name:      dslInfo.TermsInfos[index].TermsField,
					Operation: cond.OperationGt,
					Value:     series[len(series)-1][dslInfo.TermsInfos[index].AggName],
				})
			} else {
				*filters = append(*filters, interfaces.Filter{
					Name:      dslInfo.TermsInfos[index].TermsField,
					Operation: cond.OperationLt,
					Value:     series[len(series)-1][dslInfo.TermsInfos[index].AggName],
				})
			}

			// 遍历前面的字段的等于的条件
			for pre := 0; pre < index; pre++ {
				*filters = append(*filters, interfaces.Filter{
					Name:      dslInfo.TermsInfos[pre].TermsField,
					Operation: interfaces.OPERATION_EQ,
					Value:     series[len(series)-1][dslInfo.TermsInfos[pre].AggName],
				})
			}

		}
	} else {
		// 当前字段的size小于还剩余可查的序列大小,就遍历前一个,按前一个字段作为分割点.
		getSeriesPagedFilters(series, index-1, dslInfo,
			remainSize/dslInfo.TermsInfos[index].EvalSize, filters, ifLastBacth, batchNum)
	}
}

// 解析获取序列的返回结果的aggs
func IteratorSeriesAggs(aggregations gjson.Result, labels map[string]string, series *[]map[string]string, groupByi int,
	aggInfos []interfaces.AggInfo) {

	if groupByi >= len(aggInfos) {
		labelsi := make(map[string]string, 0)
		for field, value := range labels {
			labelsi[field] = value
		}
		*series = append(*series, labelsi)
		// labels = make(map[string]string)
		return
	}

	// 取桶
	switch aggInfos[groupByi].AggType {
	case interfaces.BUCKET_TYPE_TERMS, interfaces.BUCKET_TYPE_RANGE, interfaces.BUCKET_TYPE_DATE_RANGE:
		bucketsi := aggregations.Get(fmt.Sprintf("%s.buckets", aggInfos[groupByi].AggName)).Array()
		for _, bucket := range bucketsi {
			bukectKey := bucket.Get("key").String()
			labels[aggInfos[groupByi].AggName] = bukectKey
			IteratorSeriesAggs(bucket, labels, series, groupByi+1, aggInfos)
		}
	case interfaces.BUCKET_TYPE_FILTERS:
		bucketsi := aggregations.Get(fmt.Sprintf("%s.buckets", aggInfos[groupByi].AggName)).Map()
		for bukectKey, bucket := range bucketsi {
			labels[aggInfos[groupByi].AggName] = bukectKey
			IteratorSeriesAggs(bucket, labels, series, groupByi+1, aggInfos)
		}
	}

}

// 处理各个批次的获取序列数据的过滤条件
func processSeriesDataFilters(query interfaces.MetricModelQuery, batchSeriesNum int64,
	dslInfo interfaces.DslInfo, series []map[string]string) [][]interfaces.Filter {

	splitSize := batchSeriesNum // 最后一个分批字段的分批大小
	splitIndex := 0             // 用于分批的字段位置
	for splitIndex = len(dslInfo.TermsInfos) - 1; splitIndex >= 0; splitIndex-- {
		sizei := dslInfo.TermsInfos[splitIndex].EvalSize
		if sizei > splitSize {
			// i作为分批查询的分割字段
			break
		}
		splitSize /= sizei
	}
	// 前面的
	if splitIndex == -1 {
		splitIndex = 0
	}
	// 以i作为分批查询的点,前i-1个字段并发.因为查询范围不同,并发分批获取数据的分割字段可能不同,所以每次都实时处理
	// 需把前i-1个字段从序列中去重得到
	// 获取不同的 i-1 个字段
	batchFilters := make(map[string][]string)
	seriesK := make(map[string]map[string]int) // 每个前i-1个元素的不同的key的i的值的排序
	seriesKArr := make(map[string][]string)
	for _, seriesi := range series {
		fieldValues := make([]string, 0)
		key := ""
		for i := 0; i < splitIndex; i++ {
			fieldValues = append(fieldValues, seriesi[dslInfo.TermsInfos[i].AggName])
			key += seriesi[dslInfo.TermsInfos[i].AggName]
		}
		batchFilters[key] = fieldValues
		if _, exist := seriesK[key][seriesi[dslInfo.TermsInfos[splitIndex].AggName]]; !exist {
			value := make(map[string]int)
			value[seriesi[dslInfo.TermsInfos[splitIndex].AggName]] = 1
			seriesK[key] = value
			seriesKArr[key] = append(seriesKArr[key], seriesi[dslInfo.TermsInfos[splitIndex].AggName])
		}
	}

	// 构造每批并发的过滤条件
	filtersArr := make([][]interfaces.Filter, 0)
	for key, batchi := range batchFilters {
		// 第i批对应的过滤条件
		// 前i-1个字段是等于
		filters := make([]interfaces.Filter, 0)
		for i, fieldV := range batchi {
			filters = append(filters, interfaces.Filter{
				Name:      dslInfo.TermsInfos[i].TermsField,
				Operation: interfaces.OPERATION_EQ,
				Value:     fieldV,
			})
		}
		// 第i个字段分批范围
		targetSeries := int64(len(seriesKArr[key])) // 长度以第i个字段的实际子序列数为准
		splitBatch := int64(math.Ceil(float64(targetSeries) / float64(splitSize)))
		for i := int64(0); i < splitBatch; i++ {
			size := splitSize
			if (i+1)*splitSize > targetSeries {
				size = targetSeries - i*splitSize
			}

			filtersi := filters
			// 按排序方式拼过滤条件
			if dslInfo.TermsInfos[splitIndex].Direction == interfaces.ASC_DIRECTION {
				filtersi = append(filtersi, interfaces.Filter{
					Name:      dslInfo.TermsInfos[splitIndex].TermsField,
					Operation: cond.OperationGte,
					Value:     seriesKArr[key][i*splitSize],
				})
				filtersi = append(filtersi, interfaces.Filter{
					Name:      dslInfo.TermsInfos[splitIndex].TermsField,
					Operation: cond.OperationLte,
					Value:     seriesKArr[key][i*splitSize+size-1],
				})
			} else {
				filtersi = append(filtersi, interfaces.Filter{
					Name:      dslInfo.TermsInfos[splitIndex].TermsField,
					Operation: cond.OperationLte,
					Value:     seriesKArr[key][i*splitSize],
				})
				filtersi = append(filtersi, interfaces.Filter{
					Name:      dslInfo.TermsInfos[splitIndex].TermsField,
					Operation: cond.OperationGte,
					Value:     seriesKArr[key][i*splitSize+size-1],
				})
			}

			// 添加时间范围的过滤条件
			filtersi = append(filtersi, interfaces.Filter{
				Name:      interfaces.DEFAULT_DATE_FIELD,
				Operation: cond.OperationRange,
				Value:     []any{*query.Start, *query.End},
			})

			filtersArr = append(filtersArr, filtersi)
		}
	}
	return filtersArr
}

// 处理已有序列数据后的直接查询的过滤条件
func processDiractlyDataFilters(dslInfo interfaces.DslInfo, series []map[string]string) []interfaces.Filter {
	// 全部序列都查，用聚合的第一个字段在series序列中的第一个和最后一个值作为范围
	filters := make([]interfaces.Filter, 0)

	// 按排序方式拼过滤条件
	if dslInfo.TermsInfos[0].Direction == interfaces.ASC_DIRECTION {
		filters = append(filters, interfaces.Filter{
			Name:      dslInfo.TermsInfos[0].TermsField,
			Operation: cond.OperationGte,
			Value:     series[0][dslInfo.TermsInfos[0].AggName], //seriesKArr[dslInfo.TermsInfos[0].AggName][0],
		})
		filters = append(filters, interfaces.Filter{
			Name:      dslInfo.TermsInfos[0].TermsField,
			Operation: cond.OperationLte,
			Value:     series[len(series)-1][dslInfo.TermsInfos[0].AggName], // seriesKArr[dslInfo.TermsInfos[0].AggName][len(seriesKArr)-1],
		})
	} else {
		filters = append(filters, interfaces.Filter{
			Name:      dslInfo.TermsInfos[0].TermsField,
			Operation: cond.OperationLte,
			Value:     series[0][dslInfo.TermsInfos[0].AggName], //seriesKArr[dslInfo.TermsInfos[0].AggName][0],
		})
		filters = append(filters, interfaces.Filter{
			Name:      dslInfo.TermsInfos[0].TermsField,
			Operation: cond.OperationGte,
			Value:     series[len(series)-1][dslInfo.TermsInfos[0].AggName], // seriesKArr[dslInfo.TermsInfos[0].AggName][len(seriesKArr)-1],
		})
	}

	return filters
}

// 按请求处理date_histogram.
// 即时查询: 提取查询的开始时间和结束时间; 范围查询:提取实际请求的step,并构造一个实际查询的date_histogram
func processDateHistogram(ctx context.Context, query *interfaces.MetricModelQuery,
	dslInfo interfaces.DslInfo, configDelta time.Duration) (int64, error) {

	var lookBackDelta int64
	if query.IsInstantQuery {
		switch dslInfo.DateHistogram.IntervalType {
		case interfaces.INTERVAL_TYPE_FIXED:
			if dslInfo.DateHistogram.IntervalValue == interfaces.VARIABLE_INTERVAL {
				// 如果fixed_interval是变量
				lookBackDelta = convert.GetLookBackDelta(*query.End-*query.Start, configDelta)
			} else {
				// 如果fixed_interval是常量，look_back_delta会被替换成计算公式中的常量
				LookBackDeltaT, err := convert.ParseDuration(dslInfo.DateHistogram.IntervalValue)
				if err != nil {
					httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Formula).
						WithErrorDetails(err.Error())
					return 0, httpErr
				}
				if LookBackDeltaT <= 0 {
					httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Formula).
						WithErrorDetails("the value of fixed_interval is invalid")
					return 0, httpErr
				}
				lookBackDelta = LookBackDeltaT.Milliseconds()
			}
		case interfaces.INTERVAL_TYPE_CALENDAR:
			if dslInfo.DateHistogram.IntervalValue == interfaces.VARIABLE_INTERVAL {
				// 如果 calendar_interval 是变量
				lookBackDelta = convert.GetLookBackDelta(*query.End-*query.Start, configDelta)
			} else {
				// 如果 calendar_interval 是常量
				switch dslInfo.DateHistogram.IntervalValue {
				case "minute", "1m":
					lookBackDelta = time.Minute.Milliseconds()
				case "hour", "1h":
					lookBackDelta = time.Hour.Milliseconds()
				case "day", "1d":
					currentTime := time.Unix(query.Time, 0)
					oneDayAgo := currentTime.AddDate(0, 0, -1)
					lookBackDelta = currentTime.Sub(oneDayAgo).Milliseconds()
				case "week", "1w":
					currentTime := time.Unix(query.Time, 0)
					oneWeekAgo := currentTime.AddDate(0, 0, -7)
					lookBackDelta = currentTime.Sub(oneWeekAgo).Milliseconds()
				case "month", "1M":
					currentTime := time.Unix(query.Time, 0)
					oneMonthAgo := currentTime.AddDate(0, -1, 0)
					lookBackDelta = currentTime.Sub(oneMonthAgo).Milliseconds()
				case "quarter", "1q":
					currentTime := time.Unix(query.Time, 0)
					threeMonthAgo := currentTime.AddDate(0, -3, 0)
					lookBackDelta = currentTime.Sub(threeMonthAgo).Milliseconds()
				case "year", "1y":
					currentTime := time.Unix(query.Time, 0)
					oneyearAgo := currentTime.AddDate(-1, 0, 0)
					lookBackDelta = currentTime.Sub(oneyearAgo).Milliseconds()
				default:
					httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Formula).
						WithErrorDetails("the value of calendar_interval is invalid")
					return 0, httpErr
				}
			}
		default:
			return lookBackDelta, fmt.Errorf("invalid interval_type: %s", dslInfo.DateHistogram.IntervalType)
		}
		// dsl 实际上应该使用的 start end
		// query.Start = query.Time - lookBackDelta
		// query.End = query.Time
	} else {
		// 范围查询
		// 改写rangedsl的aggs中的date_histogram的值,并重新赋值查询的step
		// 因为判断处理是以解析出来的date_histogram为准,所以可以再此处改写是变量的interval,用实参替换

		queryStep := query.StepStr
		switch dslInfo.DateHistogram.IntervalType {
		case interfaces.INTERVAL_TYPE_FIXED:
			if dslInfo.DateHistogram.IntervalValue == interfaces.VARIABLE_INTERVAL {
				// fixed_interval 的变量支持固定间隔  15s, 30s, 1m, 2m, 5m, 10m, 15m, 20m, 30m, 1h, 2h, 3h, 6h, 12h, 1d, 1y。不支持日历间隔
				if _, exists := common.FixedStepsMap[*queryStep]; !exists {
					return 0, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Step).
						WithErrorDetails(fmt.Sprintf("Unsupport fixed_interval, expected one of %v, actual is %s", common.FixedSteps, *queryStep))
				}
				// dhAggInfo.IntervalValue = queryStep
				query.IsVariable = true
				// 改写dslInfo中的rangeDSL的aggs部分
				modifyVaribleInterval(dslInfo, *queryStep, "fixed_interval")

				// 记录日志
				o11y.Debug(ctx, fmt.Sprintf("The fixed_interval of date_histogram is using variable, replace variable [%s] to [%s]",
					dslInfo.DateHistogram.IntervalValue, *queryStep))
			} else {
				query.StepStr = &dslInfo.DateHistogram.IntervalValue
				// 固定步长为常量时，需校验点数是否超过11000
				stepT, err := convert.ParseDuration(*query.StepStr)
				if err != nil {
					httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Step).
						WithErrorDetails(err.Error())
					return 0, httpErr
				}

				if time.UnixMilli(*query.End).Sub(time.UnixMilli(*query.Start))/stepT > 11000 {
					httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Step).
						WithErrorDetails("exceeded maximum resolution of 11,000 points per timeseries. Try decreasing the query resolution (?step=XX)")
					return 0, httpErr
				}
			}
		case interfaces.INTERVAL_TYPE_CALENDAR:
			query.IsCalendar = true
			if dslInfo.DateHistogram.IntervalValue == interfaces.VARIABLE_INTERVAL {
				// calendar_interval的变量支持  day, week, month, quarter, year。
				if _, exists := interfaces.CALENDAR_INTERVALS[*queryStep]; !exists {
					return 0, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Step).
						WithErrorDetails(fmt.Sprintf("Unsupport calendar_interval,expected one of [minute, hour, day, week, month, quarter, year], actual is %s", *queryStep))
				}
				// dhAggInfo.IntervalValue = queryStep
				// 改写dslInfo中的rangeDSL的aggs部分
				modifyVaribleInterval(dslInfo, *queryStep, "calendar_interval")
				query.IsVariable = true
				// 记录日志
				o11y.Debug(ctx, fmt.Sprintf("The calendar_interval of date_histogram is using variable, replace variable [%s] to [%s]",
					dslInfo.DateHistogram.IntervalValue, *queryStep))
			} else {
				query.StepStr = &dslInfo.DateHistogram.IntervalValue
			}
		}
	}

	return lookBackDelta, nil
}

// 修改rangeDSL中的aggs的date_histogram的变量{{__interval}}为实参
func modifyVaribleInterval(dslInfo interfaces.DslInfo, actualInterval, intervalType string) {
	path := make([]string, 0)
	for i := 1; i < len(dslInfo.AggInfos); i++ {
		path = append(path, dslInfo.AggInfos[i].AggName)
		if dslInfo.AggInfos[i].AggName == dslInfo.DateHistogram.AggName {
			// 遇到 date_histogram就退出
			break
		}
	}

	currentMap := dslInfo.RangeQueryDSL["aggs"].(map[string]any)
	for _, key := range path[:len(path)-1] { // 遍历到最后一个key的前一个
		nextMap, ok := currentMap[key].(map[string]any)
		if !ok {
			// 如果当前项不是map类型，则直接返回，不进行修改
			return
		}
		currentMap = nextMap["aggs"].(map[string]any)
	}
	// 到达最后一个key，进行值修改
	lastKey := path[len(path)-1]

	currentMap[lastKey].(map[string]any)[interfaces.BUCKET_TYPE_DATE_HISTOGRAM].(map[string]any)[intervalType] = actualInterval
}

func evaluateTimeNum(query interfaces.MetricModelQuery, dslInfo interfaces.DslInfo) int64 {
	if query.IsInstantQuery {
		return 1
	}

	if query.IsCalendar {
		// 日历间隔
		start, end := correctingTime(query, dslInfo.DateHistogram.ZoneLocation)
		startTime := time.UnixMilli(start)
		endTime := time.UnixMilli(end)
		num := int64(1)
		switch *query.StepStr {
		case "minute", "1m":
			num = int64(endTime.Sub(startTime).Minutes())
		case "hour", "1h":
			num = int64(endTime.Sub(startTime).Hours())
		case "day", "1d":
			num = int64(endTime.Sub(startTime).Hours() / 24)
		case "week", "1w":
			num = int64(endTime.Sub(startTime).Hours() / (24 * 7))
		case "month", "1M":
			num = int64(endTime.Sub(startTime).Hours() / (24 * 30))
		case "quarter", "1q":
			num = int64(endTime.Sub(startTime).Hours() / (24 * 90))
		case "year", "1y":
			num = int64(endTime.Sub(startTime).Hours() / (24 * 365))
		}
		return max(1, num)
	}
	return max(1, int64(math.Ceil(float64(*query.End-*query.Start)/float64(*query.Step))))
}

// 从序列缓存中获取数据或者是获取序列的时间过滤条件
func getSeriesOrTimeFilter(query interfaces.MetricModelQuery, fullCacheRefreshTime time.Duration, dataView *interfaces.DataView) (DSLSeries, int64, int64, bool, bool) {
	start, end := *query.Start, *query.End
	fullFlag := true
	var dslSeriesCache DSLSeries
	// 忽略缓存时不走缓存
	if !query.IgnoringMemoryCache {
		v, ok := Series_Of_Model_Map.Load(query.MetricModelID)
		if ok {
			dslSeriesCache = v.(DSLSeries)

			canUseCache := false
			if *query.Start < dslSeriesCache.StartTime.UnixMilli() &&
				*query.End > dslSeriesCache.EndTime.UnixMilli() {
				// 缓存缺失请求的前后两段，全量查
				start, end = *query.Start, *query.End
			} else if *query.Start < dslSeriesCache.StartTime.UnixMilli() &&
				*query.End <= dslSeriesCache.EndTime.UnixMilli() {
				// 缓存缺失开始段的时间，增量查
				fullFlag = false
				start, end = *query.Start, dslSeriesCache.StartTime.UnixMilli()
			} else if *query.Start > dslSeriesCache.StartTime.UnixMilli() &&
				*query.End > dslSeriesCache.EndTime.UnixMilli() {
				// 缓存缺失end段的时间。缺失的end到now的时长不超过10min，也可以尝试读缓存，
				// 但是这10min，数据查到，序列没找到，但是序列不会更新那么频繁，是小概率事件.所以可以走缓存，避免当前时间短暂变化引起的查询慢
				if time.Since(dslSeriesCache.EndTime) > 10*time.Minute {
					// 增量查
					fullFlag = false
					start, end = dslSeriesCache.EndTime.UnixMilli(), *query.End
				} else {
					// end到now不超过10min，不刷新，从缓存中读
					canUseCache = true
				}
			} else {
				// 缓存能覆盖start,end，从缓存中读取
				canUseCache = true
			}

			// 加数据视图的更新时间
			// 若 tsidData 的缓存刷新时间大于模型更新时间 && 缓存时间距离当前小于10min && api的过滤条件为空，则返回缓存。否则实时查并刷新缓存
			if canUseCache {
				fullDiff := time.Since(dslSeriesCache.FullRefreshTime)
				if dslSeriesCache.RefreshTime.UnixMilli() >= query.ModelUpdateTime &&
					dslSeriesCache.RefreshTime.UnixMilli() >= dataView.UpdateTime &&
					len(query.Filters) == 0 {
					// 视图且模型的在缓存刷新之后未更新 且 请求没有过滤条件
					if fullCacheRefreshTime.Milliseconds() == 0 {
						fullCacheRefreshTime = 24 * time.Hour
					}
					if fullDiff <= fullCacheRefreshTime { // 24h改为全局配置
						// 全量缓存未到期，读取缓存
						return dslSeriesCache, start, end, fullFlag, true
					} else {
						// 全量缓存到期，按缓存开始结束刷新缓存
						start, end = dslSeriesCache.StartTime.UnixMilli(), dslSeriesCache.EndTime.UnixMilli()
					}
				}
			}
		}
	}
	return dslSeriesCache, start, end, fullFlag, false
}
