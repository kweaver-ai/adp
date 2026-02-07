// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package trace

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"sync"

	jsoniter "github.com/json-iterator/go"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/tidwall/gjson"

	"uniquery/common"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	"uniquery/logics"
)

var (
	tServiceOnce sync.Once
	tService     interfaces.TraceService
)

type traceService struct {
	appSetting *common.AppSetting
	osClient   interfaces.OpenSearchAccess
	lgAccess   interfaces.LogGroupAccess
}

func NewTraceService(appSetting *common.AppSetting) interfaces.TraceService {
	tServiceOnce.Do(func() {
		tService = &traceService{
			appSetting: appSetting,
			osClient:   logics.OSAccess,
			lgAccess:   logics.LGAccess,
		}
	})
	return tService
}

// GetSpanList 获取span列表
// func (ts *traceService) GetSpanList(spanListQuery interfaces.SpanListQuery) ([]interfaces.SpanList, int, error) {
// 	// 根据dataViewId获取索引类型列表及must filter
// 	dataView, status, err := ts.dvAccess.GetDataViewQueryFilters(spanListQuery.DataViewId)
// 	if err != nil {
// 		logger.Errorf("GetSpanList error: %s", err.Error())
// 		uerr := rest.NewHTTPError(status, uerrors.Uniquery_InternalError).WithErrorDetails(err.Error())
// 		return []interfaces.SpanList{}, 0, uerr
// 	}

// 	indices := dataView.IndexPattern

// 	// 如果根据dataViewId获取到的索引类型列表为空, 返回400错误
// 	if len(indices) == 0 {
// 		logger.Errorf("GetSpanList failed because the dataView whose dataViewId equals %v did not exist ", spanListQuery.DataViewId)
// 		uerr := rest.NewHTTPError(http.StatusBadRequest, uerrors.Uniquery_NoSuchARDataView).
// 			WithErrorDetails("The dataView whose dataViewId equals " + spanListQuery.DataViewId + " was not found!")
// 		return []interfaces.SpanList{}, 0, uerr
// 	}

// 	// 构造时间范围过滤器
// 	timeRangeFilter := map[string]interface{}{
// 		"range": map[string]interface{}{
// 			"StartTime": map[string]interface{}{
// 				"gte": time.UnixMilli(spanListQuery.StartTime).Format(time.RFC3339Nano),
// 				"lte": time.UnixMilli(spanListQuery.EndTime).Format(time.RFC3339Nano),
// 			},
// 		},
// 	}

// 	mustFilter := dataView.MustFilter.([]interface{})
// 	mustFilter = append(mustFilter, timeRangeFilter)

// 	// 构造span状态过滤器
// 	spanStatusFilter := make([]map[string]interface{}, 0)
// 	for status := range spanListQuery.SpanStatusMap {
// 		filter := map[string]interface{}{
// 			"term": map[string]interface{}{
// 				"Status.Code.keyword": map[string]string{
// 					"value": status,
// 				},
// 			},
// 		}
// 		spanStatusFilter = append(spanStatusFilter, filter)
// 	}

// 	// 构造dsl语句
// 	// 若不使用unmapped_type, 当查询不含span数据的日志分组时, opensearch会报No mapping found for [StartTime] in order to sort on错误
// 	// should下有多个条件时, 必须设置minimum_should_match=1才能实现或操作.
// 	dsl := map[string]interface{}{
// 		"query": map[string]interface{}{
// 			"bool": map[string]interface{}{
// 				"must":                 mustFilter,
// 				"should":               spanStatusFilter,
// 				"minimum_should_match": 1,
// 			},
// 		},
// 		"_source": map[string]interface{}{
// 			"includes": []string{"Name", "StartTime", "SpanContext.TraceID", "SpanContext.SpanID", "Duration", "Status.Code", "Resource.service.name", "SpanKind"},
// 		},
// 		"track_total_hits": true,
// 		"sort": []map[string]interface{}{
// 			{
// 				"StartTime": map[string]string{
// 					"unmapped_type": "date",
// 					"order":         "desc",
// 				},
// 			},
// 		},
// 		"from": spanListQuery.Offset,
// 		"size": spanListQuery.Limit,
// 	}

// 	// 获取Span列表
// 	res, total, err := searchSpanList(ts, dsl, indices)

// 	if err != nil {
// 		logger.Errorf("GetSpanList error: %s", err.Error())
// 		oerr := rest.NewHTTPError(http.StatusInternalServerError, uerrors.Uniquery_InternalError).WithErrorDetails(err.Error())
// 		return res, total, oerr
// 	}
// 	return res, total, nil
// }

// GetTraceDetail 单条trace详情查询
func (ts *traceService) GetTraceDetail(parentCtx context.Context, traceDataViewId, logDataViewId, traceId string) (*interfaces.TraceDetail, error) {
	errCh := make(chan error)
	wg := &sync.WaitGroup{}

	// 基于parentCtx创建一个带有取消信号的context.Context
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	traceDetail := &interfaces.TraceDetail{
		TraceID:   traceId,
		StartTime: math.MaxInt64,
		EndTime:   math.MinInt64,
		Spans:     &interfaces.BriefSpan{},
		Services:  make([]interfaces.Service, 0),
		SpanStats: map[string]int32{
			"Ok":    0,
			"Error": 0,
			"Unset": 0,
		},
		ServiceStats: make(map[interfaces.Service]int32, 0),
	}

	// 分页查询trace详情
	wg.Add(1)
	go scrollSearchTraceDetail(ts, ctx, wg, errCh, traceDataViewId, traceDetail)

	spanRelatedLogStats := interfaces.SpanRelatedLogStats{}

	// 分页查询每条span的关联日志条数
	wg.Add(1)
	go scrollSearchRelatedLogCount(ts, ctx, wg, errCh, logDataViewId, traceId, spanRelatedLogStats)

	successCh := make(chan string)
	go func() {
		wg.Wait()
		successCh <- "finsh"
	}()

	select {
	case err := <-errCh:
		logger.Errorf("GetTraceDetail error: %s", err.Error())
		return traceDetail, err
	case <-successCh:
		// 根据traceDetail.ServiceStats计算该条trace包含多少个service, 追加到traceDetail.Services中
		for svc := range traceDetail.ServiceStats {
			traceDetail.Services = append(traceDetail.Services, svc)
		}

		// 根据这条trace中所有span的状态统计给TraceStatus赋值
		// 只要出现一个Error状态的span, trace的状态就是error, 否则就为ok
		if traceDetail.SpanStats["Error"] > 0 {
			traceDetail.TraceStatus = "error"
		} else {
			traceDetail.TraceStatus = "ok"
		}

		// 整理这条trace的Duration
		traceDetail.Duration = traceDetail.EndTime - traceDetail.StartTime

		// 把每个span的关联日志条数添加到span上
		for k, v := range spanRelatedLogStats {
			if briefSpan, ok := traceDetail.SpanMap[k]; ok {
				briefSpan.RelatedLogCount = v
			}
		}

		// start1 := time.Now()
		// fmt.Printf("开始构建树结构, 当前时间%v\n", start1)

		// 构建trace数据的树结构
		traceDetail.Spans = buildTraceTree(traceDetail.SpanMap)

		// end1 := time.Now()
		// fmt.Printf("树结构构建完毕, 当前时间%v, 共耗时%v\n", end1, end1.Sub(start1))

		// start2 := time.Now()
		// fmt.Printf("开始计算traceTree的深度, 当前时间%v\n", start2)

		// 获取traceTree的层数
		traceDetail.Depth = getTraceTreeDepth(traceDetail.Spans)

		// end2 := time.Now()
		// fmt.Printf("traceTree的深度计算完毕, 当前时间%v, 共耗时%v\n", end2, end2.Sub(start2))

		return traceDetail, nil
	}
}

// func searchSpanList(ts *traceService, query map[string]interface{}, indices []string) ([]interfaces.SpanList, int, error) {
// 	// dsl查询opensearch
// 	resBytes, err := ts.osClient.SearchSpans(query, indices)
// 	if err != nil {
// 		return []interfaces.SpanList{}, 0, err
// 	}

// 	spanLists := make([]interfaces.SpanList, 0)
// 	resJson := string(resBytes)

// 	// 使用gjson对opensearch返回结果进行处理
// 	// 获取total总数
// 	totalStr := gjson.Get(resJson, "hits.total.value").String()
// 	total, _ := strconv.Atoi(totalStr)

// 	spanListArray := gjson.Get(resJson, "hits.hits").Array()
// 	for i := 0; i < len(spanListArray); i++ {
// 		// 获取service name
// 		serviceName := spanListArray[i].Get("_source.Resource.service.name").String()

// 		// 获取span列表项
// 		source := spanListArray[i].Get("_source").String()
// 		span := interfaces.Span{}

// 		err := sonic.Unmarshal([]byte(source), &span)
// 		if err != nil {
// 			return []interfaces.SpanList{}, 0, err
// 		}

// 		// 将span.StartTime由RFC3339转换为Unix时间戳
// 		timeStrs := []string{span.StartTime.(string)}
// 		microTimestamps, err := convert.RFC3339ToMicroTimestamp(timeStrs)
// 		if err != nil {
// 			return []interfaces.SpanList{}, 0, err
// 		}
// 		span.StartTime = microTimestamps[0]

// 		spanList := interfaces.SpanList{
// 			Name:        span.Name,
// 			SpanStatus:  span.Status.Code,
// 			SpanID:      span.SpanContext.SpanID,
// 			TraceID:     span.SpanContext.TraceID,
// 			StartTime:   span.StartTime,
// 			Duration:    span.Duration,
// 			ServiceName: serviceName,
// 		}

// 		// 将span.SpanKind由数字映射成对应的字符串, 如果数字没有对应的字符串则展示原数字
// 		if spanKindStr, ok := interfaces.SPANKIND_MAPPING_TABLE[span.SpanKind]; ok {
// 			spanList.SpanKind = spanKindStr
// 		} else {
// 			spanList.SpanKind = span.SpanKind
// 		}

// 		// span列表项追加到切片中
// 		spanLists = append(spanLists, spanList)
// 	}
// 	return spanLists, total, nil
// }

// scrollSearchTraceDetail 分页获取Trace详情
func scrollSearchTraceDetail(ts *traceService, ctx context.Context, wg *sync.WaitGroup, errCh chan<- error, traceDataViewId string, traceDetail *interfaces.TraceDetail) {
	defer wg.Done()

	select {
	case <-ctx.Done():
		return
	default:
		// oStart := time.Now()
		// fmt.Printf("开始查询Trace详情, 当前时间为%v\n", oStart)

		// 根据traceDataViewId获取索引类型列表及must filter
		traceIndices, traceMustFilter, isExist, err := getIndicesAndMustFilters(ts, traceDataViewId)
		if err != nil {
			uerr := rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_InternalError).
				WithErrorDetails("Get indices and must filters failed by trace_data_view_id: " + err.Error())
			errCh <- uerr
			return
		}

		// 如果isExist为false, 说明traceDataViewId对应的DataView不存在, 但是由于traceDataViewId不是url上的参数, 按照RESTFul规范应返回400错误
		if !isExist {
			uerr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_Trace_TraceDataViewNotFound).
				WithErrorDetails("The trace_data_view whose id equals " + traceDataViewId + " was not found!")
			errCh <- uerr
			return
		}

		// 如果日志分组下没有索引, 没有必须往下执行, 返回TraceNotFound即可
		if len(traceIndices) == 0 {
			uerr := rest.NewHTTPError(ctx, http.StatusNotFound, uerrors.Uniquery_Trace_TraceNotFound).
				WithErrorDetails("The trace whose traceId equals " + traceDetail.TraceID + " was not found!")
			errCh <- uerr
			return
		}

		// 按traceId查找trace数据, 构造termFilter
		termFilter := map[string]interface{}{
			"term": map[string]interface{}{
				"SpanContext.TraceID.keyword": map[string]string{
					"value": traceDetail.TraceID,
				},
			},
		}

		// 把termFilter追加到traceMustFilter中
		traceMustFilter = append(traceMustFilter, termFilter)

		// 根据traceMustFilter构造dsl语句, size为自定义默认值
		traceDsl := map[string]interface{}{
			"query": map[string]interface{}{
				"bool": map[string]interface{}{
					"must": traceMustFilter,
				},
			},
			"docvalue_fields": []string{"SpanContext.SpanID.keyword", "StartTime", "EndTime", "Duration", "Name.keyword", "Parent.SpanID.keyword", "Status.CodeDesc.keyword", "Resource.service.name.keyword", "SpanKindDesc.keyword"},
			"_source":         false,
			"size":            interfaces.MAX_SEARCH_SIZE,
		}

		// resBytes: driven层返回结果
		var resBytes []byte

		// totalCount: span总数
		var totalCount int64

		// searchCount: 分页请求次数
		searchCount := 1

		// scrollId: scroll查询要使用的id
		var scrollId string

		// scrollIds: 记录分页查询中的所有scrollId, 等待分页查询结束时, 统一删除, 避免占用opensearch内存空间
		scrollIds := make([]string, 0)

		// 读写锁
		var mutex sync.RWMutex

		// 调用driven层, 获取trace详情
		for {
			// start := time.Now()
			// fmt.Printf("第%d次开始查询Trace详情, 当前时间为%v\n", searchCount, start)

			if searchCount == 1 {
				// 第一次是调用SearchSubmit接口
				resBytes, _, err = ts.osClient.SearchSubmit(ctx, traceDsl, traceIndices,
					interfaces.DEFAULT_SEARCH_SCROLL_DURATION, interfaces.DEFAULT_PREFERENCE, true)
			} else {
				// 从第二次开始调用Scroll接口
				query := interfaces.Scroll{
					ScrollId: scrollId,
					Scroll:   interfaces.DEFAULT_SEARCH_SCROLL_STR,
				}
				resBytes, _, err = ts.osClient.Scroll(ctx, query)
			}

			// end := time.Now()
			// fmt.Printf("第%d次结束查询Trace详情, 当前时间为%v\n, 耗时%v\n", searchCount, end, end.Sub(start))

			// driven层报错, 返回500错误
			if err != nil {
				uerr := rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_InternalError).WithErrorDetails(err.Error())
				errCh <- uerr
				return
			}

			// start2 := time.Now()
			// fmt.Printf("第%d次开始用GJson处理opensearch返回结果, 当前时间为%v\n", searchCount, start2)

			// 使用gjson对opensearch返回结果进行处理
			resJson := string(resBytes)
			scrollId = gjson.Get(resJson, "_scroll_id").String()
			scrollIds = append(scrollIds, scrollId)

			// 直接定义结构体反序列化span, 得到scrollId, 测试下来性能不如gjson
			// briefResStruct := interfaces.BriefScrollSearchResponse{}
			// decoder := jsoniter.NewDecoder(bytes.NewBuffer(resBytes))
			// decoder.UseNumber()

			// if err := decoder.Decode(&briefResStruct); err != nil {
			// 	uerr := rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_InternalError).WithErrorDetails(err.Error())
			// 	errCh <- uerr
			// 	return
			// }

			// scrollId = briefResStruct.ScrollID
			// scrollIds = append(scrollIds, scrollId)

			// end2 := time.Now()
			// fmt.Printf("第%d次结束用GJson处理opensearch返回结果, 当前时间为%v, 耗时%v\n", searchCount, end2, end2.Sub(start2))

			if searchCount == 1 {
				// 更新totalCount, 因为使用的是scroll_search, 所以返回的total是所有span的统计
				totalCount = gjson.Get(resJson, "hits.total.value").Int()
				// totalCount = int32(briefResStruct.BriefOuterHits.Total.Value)

				if totalCount == 0 {
					// 如果totalCount为0, 说明trace未找到
					uerr := rest.NewHTTPError(ctx, http.StatusNotFound, uerrors.Uniquery_Trace_TraceNotFound).
						WithErrorDetails("The trace whose traceId equals " + traceDetail.TraceID + " was not found!")
					errCh <- uerr
					return
				}

				// 在第一次分页查询获取total后, 再初始化spanMap的容量, 避免后续频繁扩容, 影响性能
				traceDetail.SpanMap = make(map[string]*interfaces.BriefSpan, totalCount)
			}

			wg.Add(1)
			go processSpanArray(resBytes, traceDetail, wg, ctx, errCh, &mutex)

			// 判断是否需要继续分页查询
			if totalCount <= int64(searchCount*interfaces.MAX_SEARCH_SIZE) {
				break
			}
			searchCount++
		}

		// 删除本次分页查询中使用到的所有scroll_id
		go clearScrollIds(context.Background(), ts, scrollIds)

		// oEnd := time.Now()
		// fmt.Printf("结束查询Trace详情, 当前时间为%v\n, 耗时%v\n", oEnd, oEnd.Sub(oStart))
	}
}

// scrollSearchRelatedLogCount 获取单条trace下每个span的关联日志条数
func scrollSearchRelatedLogCount(ts *traceService, ctx context.Context, wg *sync.WaitGroup, errCh chan<- error, logDataViewId string, traceId string, spanRelatedLogStats interfaces.SpanRelatedLogStats) {
	defer wg.Done()

	select {
	case <-ctx.Done():
		return
	default:
		// start := time.Now()
		// fmt.Printf("开始查询Trace关联日志条数, 当前时间为%v\n", start)

		// 根据traceDataViewId与logDataViewId获取索引类型列表及must filter
		logIndices, logMustFilter, isExist, err := getIndicesAndMustFilters(ts, logDataViewId)
		if err != nil {
			uerr := rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_InternalError).
				WithErrorDetails("Get indices and must filters failed by log_data_view_id: " + err.Error())
			errCh <- uerr
			return
		}

		// 如果isExist为false, 说明logDataViewId对应的DataView不存在, 但是由于logDataViewId不是url上的参数, 按照RESTFul规范应返回400错误
		if !isExist {
			uerr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_Trace_LogDataViewNotFound).
				WithErrorDetails("The log_data_view whose id equals " + logDataViewId + " was not found!")
			errCh <- uerr
			return
		}

		// 如果日志分组下没有索引, 直接return
		if len(logIndices) == 0 {
			return
		}

		// 按traceId查找trace数据, 构造termFilter
		termFilter := map[string]interface{}{
			"term": map[string]interface{}{
				"SpanContext.TraceID.keyword": map[string]string{
					// "Link.TraceId.keyword": map[string]string{
					"value": traceId,
				},
			},
		}

		// 把termFilter追加到traceMustFilter中
		logMustFilter = append(logMustFilter, termFilter)

		// 根据traceMustFilter构造dsl语句, size为自定义默认值
		logDsl := map[string]interface{}{
			"query": map[string]interface{}{
				"bool": map[string]interface{}{
					"must": logMustFilter,
				},
			},
			"size": 0,
			"aggs": map[string]interface{}{
				"group_by_SpanID": map[string]interface{}{
					"terms": map[string]interface{}{
						"field": "SpanContext.SpanID.keyword",
						// "field":               "Link.SpanId.keyword",
						"size":                interfaces.MAX_SEARCH_RELATED_LOGS_BUCKET,
						"min_doc_count":       1,
						"shard_min_doc_count": 1,
						"order": map[string]interface{}{
							"_key": interfaces.ASC_DIRECTION,
						},
					},
				},
			},
		}

		// resBytes: driven层返回结果
		var resBytes []byte

		// 调用driven层, 获取关联日志条数
		for {
			// 调用SearchSubmit接口
			resBytes, _, err = ts.osClient.SearchSubmit(ctx, logDsl, logIndices, 0,
				interfaces.DEFAULT_PREFERENCE, interfaces.DEFAULT_TRACK_TOTAL_HITS)

			// driven层报错, 返回500错误
			if err != nil {
				uerr := rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_InternalError).WithErrorDetails(err.Error())
				errCh <- uerr
				return
			}

			// 使用gjson对opensearch返回结果进行处理
			// resJson := string(resBytes)
			// logBuckets := gjson.Get(resJson, "aggregations.group_by_SpanID.buckets").Array()

			// processLogBuckets(logBuckets, spanRelatedLogStats)

			// 用jsoniter批量反序列化bucket
			logAggResponse := interfaces.LogAggResponse{}
			decoder := jsoniter.NewDecoder(bytes.NewBuffer(resBytes))

			if err := decoder.Decode(&logAggResponse); err != nil {
				uerr := rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_InternalError).WithErrorDetails(err.Error())
				errCh <- uerr
				return
			}

			logStatBuckets := logAggResponse.Aggs.GroupBy.Buckets
			length := len(logStatBuckets)
			// 遍历resStruct, 更新spanRelatedLogStats
			for i := 0; i < length; i++ {
				spanId := logStatBuckets[i].Key
				docCount := logStatBuckets[i].DocCount

				spanRelatedLogStats[spanId] = docCount
			}

			// 判断是否需要继续分页查询
			if length < interfaces.MAX_SEARCH_RELATED_LOGS_BUCKET {
				break
			}

			// 找到最后一个spanId
			criticalSpanId := logStatBuckets[length-1].Key

			// 给query重新赋值, 再次查询需要满足SpanContext.SpanID.keyword的range条件
			logDsl["query"] = map[string]interface{}{
				"bool": map[string]interface{}{
					"must": logMustFilter,
					"filter": []map[string]interface{}{
						{
							"range": map[string]interface{}{
								"SpanContext.SpanID.keyword": map[string]interface{}{
									// "Link.SpanId.keyword": map[string]string{
									"gt": criticalSpanId,
								},
							},
						},
					},
				},
			}
		}

		// end := time.Now()
		// fmt.Printf("结束查询Trace关联日志条数, 当前时间为%v\n, 耗时%v\n", end, end.Sub(start))
	}
}

// getIndicesAndMustFilters 可根据日志分组Id获取对应的索引类型列表及must filter
func getIndicesAndMustFilters(ts *traceService, dataViewId string) ([]string, []interface{}, bool, error) {
	traceDataView, isExist, err := ts.lgAccess.GetLogGroupQueryFilters(dataViewId)
	if err != nil || !isExist {
		return []string{}, []interface{}{}, false, err
	}

	indices := traceDataView.IndexPattern
	mustFilter, ok := traceDataView.MustFilter.([]interface{})
	if !ok {
		err := errors.New("the must_filters field cannot be converted to an interface array")
		return []string{}, []interface{}{}, false, err
	}

	return indices, mustFilter, true, nil
}

// processSpanArray 处理gjson格式的span数组
func processSpanArray(resBytes []byte, traceDetail *interfaces.TraceDetail, wg *sync.WaitGroup, ctx context.Context, errCh chan<- error, mutex *sync.RWMutex) {
	// start1 := time.Now()
	// fmt.Printf("开始第%d次反序列化opensearch返回结果, 当前时间为%v\n", searchCount, start1)

	defer wg.Done()

	// 用jsoniter批量反序列化span
	resStruct := interfaces.ScrollSearchResponse{}
	decoder := jsoniter.NewDecoder(bytes.NewBuffer(resBytes))
	decoder.UseNumber()

	if err := decoder.Decode(&resStruct); err != nil {
		uerr := rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_InternalError).WithErrorDetails(err.Error())
		errCh <- uerr
		return
	}

	spanArray := resStruct.OuterHits.InnerHits

	// end1 := time.Now()
	// fmt.Printf("结束第%d次反序列化opensearch返回结果, 当前时间为%v\n, 耗时%v\n", searchCount, end1, end1.Sub(start1))

	// start2 := time.Now()
	// fmt.Printf("开始第%d次遍历处理spanArray, 当前时间为%v\n", searchCount, start2)

	// 遍历spanArray, 更新traceDetail
	for i := 0; i < len(spanArray); i++ {
		span := spanArray[i].Fields
		newSpan := &interfaces.BriefSpan{}

		// 处理SpanContext.SpanID.keyword
		spanIds, ok := span["SpanContext.SpanID.keyword"].([]interface{})
		if !ok {
			continue
		}

		spanId, ok := spanIds[0].(string)
		if !ok {
			continue
		}
		newSpan.SpanContext.SpanID = spanId
		newSpan.Key = spanId

		// 处理Parent.SpanID.keyword
		parentSpanIds, ok := span["Parent.SpanID.keyword"].([]interface{})
		if !ok {
			continue
		}

		parentSpanId, ok := parentSpanIds[0].(string)
		if !ok {
			continue
		}
		newSpan.Parent.SpanID = parentSpanId

		// 处理Resource.service.name.keyword
		serviceNames, ok := span["Resource.service.name.keyword"].([]interface{})
		if ok {
			newSpan.Resource.Service.Name = serviceNames[0]
		}

		// 处理StartTime, 将其转为int64
		startTimes, ok := span["StartTime"].([]interface{})
		if !ok {
			continue
		}

		startTimeStr, ok := startTimes[0].(json.Number)
		if !ok {
			err := fmt.Errorf("an error occurred while converting the field type of the span whose spanId is %s, err: StartTime is not a json.Number, so it can not be converted to an int64", newSpan.SpanContext.SpanID)
			uerr := rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_InternalError).WithErrorDetails(err.Error())
			errCh <- uerr
			return
		}
		startTime, _ := strconv.ParseInt(string(startTimeStr), 10, 64)
		newSpan.StartTime = startTime

		// 处理EndTime, 将其转为int64
		endTimes, ok := span["EndTime"].([]interface{})
		if !ok {
			continue
		}

		endTimeStr, ok := endTimes[0].(json.Number)
		if !ok {
			err := fmt.Errorf("an error occurred while converting the field type of the span whose spanId is %s, err: EndTime is not a json.Number, so it can not be converted to an int64", newSpan.SpanContext.SpanID)
			uerr := rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_InternalError).WithErrorDetails(err.Error())
			errCh <- uerr
			return
		}
		endTime, _ := strconv.ParseInt(string(endTimeStr), 10, 64)
		newSpan.EndTime = endTime

		// 处理Duration
		newSpan.Duration = newSpan.EndTime - newSpan.StartTime

		// 处理Name.keyword
		names, ok := span["Name.keyword"].([]interface{})
		if ok {
			newSpan.Name = names[0]
		}

		// 处理SpanKindDesc.keyword
		spanKindDescs, ok := span["SpanKindDesc.keyword"].([]interface{})
		if ok {
			newSpan.SpanKindDesc = spanKindDescs[0]
		}

		// 处理Status.CodeDesc
		CodeDescs, ok := span["Status.CodeDesc.keyword"].([]interface{})
		if !ok {
			continue
		}

		codeDesc, ok := CodeDescs[0].(string)
		if !ok {
			continue
		}
		newSpan.Status.CodeDesc = codeDesc

		// 添加字段, RelatedLogCount: span关联日志条数; SubSpans: 子span集合
		newSpan.RelatedLogCount = 0
		newSpan.Children = make([]*interfaces.BriefSpan, 0)

		// 加锁, 保证traceDetail并发安全
		mutex.Lock()

		// 更新traceDetail.StartTime和traceDetail.EndTime
		traceDetail.StartTime = minInt64(traceDetail.StartTime, newSpan.StartTime)
		traceDetail.EndTime = maxInt64(traceDetail.EndTime, newSpan.EndTime)

		// 记录svc到ServiceStats中
		traceDetail.ServiceStats[newSpan.Resource.Service] = 1

		// 获取span的状态, 并更新traceDetail.SpanStats
		// status := source.Get("Status.CodeDesc").String()
		status := newSpan.Status.CodeDesc
		if status == "Ok" || status == "Error" || status == "Unset" {
			traceDetail.SpanStats[status] += 1
		}

		// span添加到spanMap中
		traceDetail.SpanMap[newSpan.SpanContext.SpanID] = newSpan

		// 解锁
		mutex.Unlock()
	}

	// end2 := time.Now()
	// fmt.Printf("结束第%d次遍历处理spanArray, 当前时间为%v\n, 耗时%v\n", searchCount, end2, end2.Sub(start2))
}

// processLogBuckets 处理gjson格式的log数组
// func processLogBuckets(logBuckets []gjson.Result, spanRelatedLogStats interfaces.SpanRelatedLogStats) {
// func processLogBuckets(resBytes []byte, ctx context.Context, errCh chan<- error, spanRelatedLogStats interfaces.SpanRelatedLogStats) int64 {
// 	// 用jsoniter批量反序列化bucket
// 	logStatBuckets := []interfaces.LogStatBucket{}
// 	decoder := jsoniter.NewDecoder(bytes.NewBuffer(resBytes))

// 	if err := decoder.Decode(&logStatBuckets); err != nil {
// 		uerr := rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_InternalError).WithErrorDetails(err.Error())
// 		errCh <- uerr
// 		return
// 	}

// 	length := len(logStatBuckets)
// 	// 遍历resStruct, 更新spanRelatedLogStats
// 	for i := 0; i < length; i++ {
// 		spanId := logStatBuckets[i].Key
// 		docCount := logStatBuckets[i].DocCount

// 		spanRelatedLogStats[spanId] = docCount
// 	}

// 	return length, logStatBuckets[length-1].Key
// }

// buildTraceTree 构建Trace树结构
func buildTraceTree(spanMap map[string]*interfaces.BriefSpan) *interfaces.BriefSpan {
	var rootSpanId string

	for spanId, span := range spanMap {
		parentSpanId := span.Parent.SpanID
		if parentSpan, ok := spanMap[parentSpanId]; ok {
			subSpans := parentSpan.Children
			// 使用sort.Search查找插入位置
			idx := sort.Search(len(subSpans), func(i int) bool {
				return subSpans[i].StartTime >= span.StartTime
			})

			// 插入span
			parentSpan.Children = append(subSpans[:idx], append([]*interfaces.BriefSpan{span}, subSpans[idx:]...)...)
		} else {
			rootSpanId = spanId
		}
	}

	return spanMap[rootSpanId]
}

// buildTraceTree2 构建Trace树结构
// func buildTraceTree2(spanMap map[string]*interfaces.BriefSpan) *interfaces.BriefSpan {
// 	var rootSpanId string

// 	for spanId, span := range spanMap {
// 		parentSpanId := span.Parent.SpanID
// 		if parentSpan, ok := spanMap[parentSpanId]; ok {
// 			subSpans := parentSpan.Children
// 			// 使用sort.Search查找插入位置
// 			idx := sort.Search(len(subSpans), func(i int) bool {
// 				return subSpans[i].StartTime >= span.StartTime
// 			})

// 			// 插入span
// 			parentSpan.Children = append(subSpans[:idx], append([]*interfaces.BriefSpan{span}, subSpans[idx:]...)...)
// 		} else {
// 			rootSpanId = spanId
// 		}
// 	}

// 	return spanMap[rootSpanId]
// }

// getTraceTreeDepth 广度优先搜索(BFS), 获取TraceTree的深度
func getTraceTreeDepth(rootSpan *interfaces.BriefSpan) int32 {
	var depth int32

	if rootSpan == nil {
		return depth
	}

	queue := []*interfaces.BriefSpan{rootSpan}
	for len(queue) > 0 {
		levelSize := len(queue)
		for i := 0; i < levelSize; i++ {
			span := queue[0]
			queue = append(queue[1:], span.Children...)
		}
		depth++
	}
	return depth
}

// clearScrollIds 清理一次分页查询中使用到的所有scroll_id
func clearScrollIds(ctx context.Context, ts *traceService, scrollIds []string) {
	para := interfaces.DeleteScroll{
		ScrollId: scrollIds,
	}

	// 必须接返回值, 否则代码检查通不过
	_, _, _ = ts.osClient.DeleteScroll(ctx, para)
}

func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
