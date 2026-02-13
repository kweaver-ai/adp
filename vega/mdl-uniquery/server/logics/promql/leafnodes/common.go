// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package leafnodes

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/tidwall/gjson"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"uniquery/common"
	cond "uniquery/common/condition"
	"uniquery/common/convert"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	"uniquery/logics/promql/labels"
	"uniquery/logics/promql/parser"
	"uniquery/logics/promql/util"
)

type LeafNodes struct {
	appSetting *common.AppSetting
	osAccess   interfaces.OpenSearchAccess
	lgAccess   interfaces.LogGroupAccess
	dvService  interfaces.DataViewService
}

func NewLeafNodes(appSetting *common.AppSetting, osAccess interfaces.OpenSearchAccess,
	lgAccess interfaces.LogGroupAccess, dvService interfaces.DataViewService) *LeafNodes {
	return &LeafNodes{
		appSetting: appSetting,
		osAccess:   osAccess,
		lgAccess:   lgAccess,
		dvService:  dvService,
	}
}

var (
	// 用于缓存索引库下的所有索引及分片数
	Number_Of_Shards_Map sync.Map

	// ExitChan 每30分钟刷新一次缓存：number_of_shards_map
	ExitChan = make(chan bool)

	// 用于缓存指标在当前模型下的目标序列Tsid，缓存有效时间10min
	Tsids_Of_Model_Metric_Map sync.Map
)

type MapResult struct {
	LabelsMap   map[string][]*labels.Label
	TsValueMap  map[string][][]gjson.Result
	TotalSeries int
	Err         uerrors.PromQLError
	Status      int
}

type TsidData struct {
	RefreshTime     time.Time // 增量缓存刷新时间
	StartTime       time.Time // 缓存序列的数据的start
	EndTime         time.Time // 缓存序列的数据的end
	FullRefreshTime time.Time // 全量刷新时间
	Tsids           []string
	TsidsMap        map[string]labels.Labels
}

type taskFunc func()

func (leafNodes *LeafNodes) RefreshShards() {
	tick := time.NewTicker(time.Minute * time.Duration(30))
	defer tick.Stop()
	for {
		select {
		case <-ExitChan:
			logger.Info("exit RefreshShards")
			return
		case <-tick.C:
			Number_Of_Shards_Map.Range(leafNodes.Refresh)
		}
	}
}

// 返回索引库下的所有索引以及对应的分片数，错误状态码，错误信息。
func (leafNodes *LeafNodes) GetIndicesNumberOfShards(ctx context.Context, indices []string) ([]*interfaces.IndexShards, int, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "获取索引分片信息")
	span.SetAttributes(attribute.Key("indexes").StringSlice(indices))
	defer span.End()

	var indexShardsArr = make([]*interfaces.IndexShards, 0)
	for _, indexBase := range indices {
		// 缓存存的是 indexbase-> []interfaces.IndexShards
		v, ok := Number_Of_Shards_Map.Load(indexBase)
		// 如果缓存中的索引库的索引分片信息为空，那么再次从opensearch中获取
		if ok && len(v.([]*interfaces.IndexShards)) > 0 {
			indexShardsArr = append(indexShardsArr, v.([]*interfaces.IndexShards)...)
		} else {
			// 从 opensearch 中查询 GET _cat/indices/metricbeat*?v&h=index,pri&format=json
			// 通过视图提供的接口获取
			resBytes, status, err := leafNodes.dvService.LoadIndexShards(ctx, indexBase)
			if err != nil {
				span.SetStatus(codes.Error, fmt.Sprintf("GET _cat/indices/%s error", indexBase))
				// 记录异常日志
				o11y.Error(ctx, fmt.Sprintf("GET _cat/indices/%s?v&h=index,pri&format=json error: %v", indexBase, err))

				return nil, status, err
			}

			var indexShardsArri []*interfaces.IndexShards
			err = sonic.Unmarshal(resBytes, &indexShardsArri)
			if err != nil {
				span.SetStatus(codes.Error, "Unmarshal index shards error")
				// 记录异常日志
				o11y.Error(ctx, fmt.Sprintf("Unmarshal index shards error: %v", err))

				return nil, http.StatusInternalServerError, errors.New("common.GetIndicesNumberOfShards Unmarshal index shards error: " + err.Error())
			}

			// 放入返回结果集中
			indexShardsArr = append(indexShardsArr, indexShardsArri...)
			// 写入缓存
			Number_Of_Shards_Map.Store(indexBase, indexShardsArri)
		}
	}

	span.SetStatus(codes.Ok, "")
	return indexShardsArr, http.StatusOK, nil
}

// 遍历到每个元素的处理函数：每个元素都重新按索引库重新加载
func (leafNodes *LeafNodes) Refresh(key, value interface{}) bool {
	resBytes, _, err := leafNodes.dvService.LoadIndexShards(context.Background(), key.(string))
	if err != nil {
		return false
	}

	var indexShardsArri []interfaces.IndexShards
	err = sonic.Unmarshal(resBytes, &indexShardsArri)
	if err != nil {
		return false
	}

	// 放入缓存 map 中
	var indexShardsArr = make([]interfaces.IndexShards, 0)
	indexShardsArr = append(indexShardsArr, indexShardsArri...)

	Number_Of_Shards_Map.Store(key, indexShardsArr)
	return true
}

// 返回指标在当前模型下的目标序列数据
func (ln *LeafNodes) GetTsidsOfModelMetric(ctx context.Context, expr *parser.VectorSelector,
	query *interfaces.Query, baseTypes []string, mustFilter interface{}) (TsidData, int, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "获取目标序列")
	span.SetAttributes(attribute.Key("indexes").StringSlice(baseTypes))
	defer span.End()

	exprJson, err := sonic.Marshal(expr)
	if err != nil {
		return TsidData{}, http.StatusInternalServerError, err
	}
	md5Hasher := md5.New()
	md5Hasher.Write(exprJson)
	hashed := md5Hasher.Sum(nil)
	exprId := hex.EncodeToString(hashed)
	cacheKey := fmt.Sprintf("%s+%s", query.ModelId, exprId)

	now := time.Now()
	tsidDataCache, start, end, fullFlag, ifFromCache := getTsidOrTimeFilter(query, cacheKey, ln.appSetting.ServerSetting.FullCacheRefreshInterval)
	if ifFromCache {
		return tsidDataCache, http.StatusOK, nil
	}
	// 需要查询，组装时间范围过滤条件
	filter := interfaces.Filter{
		Name:      "@timestamp",
		Operation: cond.OperationRange,
		Value:     []any{start, end},
	}

	// newQuery 用于构造缓存刷新的时间过滤条件
	newQuery := *query
	newQuery.Filters = append(newQuery.Filters, filter)

	// 从缓存中获取不到tsidData，或者是缓存刷新时间小于模型更新时间，或者缓存时间距离当前大于10min，那么从opensearch中获取
	// 在获取tsid时，需要按basetype获取所有的tsid，不是用索引优化的内容。
	indexPattern := convert.GetIndexBasePattern(baseTypes)

	tsids := make([]string, 0)
	tsidsMap := make(map[string]labels.Labels)
	tsidOffset := "0"
	condition := true
	for condition {
		// 1. 构造dsl： query(exists metricName && filters), terms(tsid, order by tsid,size 1w),top_hits(size:1)
		dsl, status, err := makeGetTsidDSL(*expr, &newQuery, tsidOffset, mustFilter)
		if err != nil {
			return TsidData{}, status, err
		}
		// 2. 发起查询
		res, status, err := ln.dvService.GetDataFromOpenSearchWithBuffer(ctx, *dsl, indexPattern,
			0, interfaces.DEFAULT_PREFERENCE)
		if err != nil {
			return TsidData{}, status, err
		}

		// 3. 解析结果，从中读取tsid和labels。只能顺序做，因为这一批的最后一个是下一批的过滤条件
		jsons := string(res)
		buckets := gjson.Get(jsons, fmt.Sprintf("aggregations.%s.buckets", interfaces.TSID)).Array()
		for _, bucket := range buckets {
			tsid := bucket.Get("key").String()
			// labelArr := make([]*labels.Label, 0)
			var labelsArr labels.Labels
			for labelN, labelV := range bucket.Get("labels.hits.hits.0._source.labels").Map() {
				labelsArr = append(labelsArr, &labels.Label{
					Name:  labelN,
					Value: labelV.String(),
				})
			}
			tsids = append(tsids, tsid)
			tsidsMap[tsid] = labelsArr.Sort()
		}
		// 当前批次的最后一个tsid作为下次tsid的起始点(大于 offsetTsid)
		if len(tsids) > 0 {
			tsidOffset = tsids[len(tsids)-1]
		}

		if len(buckets) == 0 || query.MaxSearchSeriesSize == 1 {
			condition = false
		}
	}

	if fullFlag {
		// 全量刷。查询到的tsid为全量。直接返回和缓存
		tsidDataCache = TsidData{
			Tsids:           tsids,
			TsidsMap:        tsidsMap,
			FullRefreshTime: now,
		}
	} else {
		// 增量、使用请求时间，需逐个添加到缓存中
		for k, v := range tsidsMap {
			if _, exist := tsidDataCache.TsidsMap[k]; !exist {
				// tsid在缓存中不存在，追加到缓存对象中。存在的保持缓存的对象，不改变
				// 合并时还需保证有序。
				tsidDataCache.TsidsMap[k] = v
				tsidDataCache.Tsids = append(tsidDataCache.Tsids, k)
				sort.Strings(tsidDataCache.Tsids)
			}
		}
	}
	tsidDataCache.StartTime = time.UnixMilli(start)
	tsidDataCache.EndTime = time.UnixMilli(end)
	tsidDataCache.RefreshTime = now

	// 预览时不缓存，若 filters 不为空，则实时查询，结果不缓存。否则刷新缓存
	if query.ModelId != "" && len(query.Filters) == 0 {
		// 写入缓存
		Tsids_Of_Model_Metric_Map.Store(cacheKey, tsidDataCache)
	}

	span.SetStatus(codes.Ok, "")
	return tsidDataCache, http.StatusOK, nil
}

// 通用处理函数： 获取日志分组的索引信息 -> 构造 dsl -> 获取索引库下的所有索引以及对应的分片数 -> 执行 dsl
func (ln *LeafNodes) commonProcess(ctx context.Context, expr *parser.VectorSelector, query *interfaces.Query,
	aggregationType string) (interface{}, int, error) {

	// 如果是指标模型的接口，走commonProcessForMetricModel，使用的是数据视图
	if query.IsMetricModel {
		return ln.commonProcessForMetricModel(ctx, expr, query, aggregationType)
	}

	// 请求 data-manager 获取日志分组的索引信息
	var (
		logGroup = query.LogGroup
		status   int
		err      error
	)

	// 是指标模型查询 && dataviewid 为空，则不查，否则查
	if !(query.IsMetricModel && query.LogGroupId == "") {
		logGroup, status, err = ln.GetLogGroupQueryFilters(ctx, query.LogGroupId)
		if err != nil {
			return nil, status, err
		}
		query.LogGroup = logGroup
	}

	// 区分__labels_str和tsid的索引模式列表
	tsidIndexPs := make([]string, 0)
	labelsStrIndexPs := make([]string, 0)
	// 当前是索引模式，当切换到视图时，用的是索引库类型，届时INDEX_PATTERN_SPLIT_TIME的key就是用索引库类型，无需拼接模式
	// 日志分组拿到的是pattern，需要改为type，不拼接匹配符，或者是保留两个
	for _, indexP := range logGroup.IndexPattern {
		splitTime, ok := interfaces.INDEX_PATTERN_SPLIT_TIME[indexP]
		if !ok {
			// 不存在静态表中的索引库，直接扔给tsid查
			tsidIndexPs = append(tsidIndexPs, indexP)
		} else if splitTime.Before(time.UnixMilli(query.Start)) {
			// 分割点在查询开始时间之前，查tsid
			tsidIndexPs = append(tsidIndexPs, indexP)
		} else if splitTime.After(time.UnixMilli(query.End)) {
			labelsStrIndexPs = append(labelsStrIndexPs, indexP)
		} else {
			tsidIndexPs = append(tsidIndexPs, indexP)
			labelsStrIndexPs = append(labelsStrIndexPs, indexP)
		}
	}

	// 获取索引库下的所有索引以及对应的分片数
	// 根据indexParttern得到的索引分片信息中无法得知索引所属的类型，所以__labels_str和tsid的索引库模型分开处理。
	labelsStrShardsArr, status, err := ln.GetIndicesNumberOfShards(ctx, labelsStrIndexPs)
	if err != nil {
		return nil, status, uerrors.PromQLError{
			Typ: uerrors.ErrorInternal,
			Err: errors.New(err.Error()),
		}
	}

	tsidShardsArr, status, err := ln.GetIndicesNumberOfShards(ctx, tsidIndexPs)
	if err != nil {
		return nil, status, uerrors.PromQLError{
			Typ: uerrors.ErrorInternal,
			Err: errors.New(err.Error()),
		}
	}

	if len(labelsStrShardsArr) == 0 && len(tsidShardsArr) == 0 {
		return MapResult{}, http.StatusOK, nil
	}

	mapResult := MapResult{
		LabelsMap:  make(map[string][]*labels.Label),
		TsValueMap: make(map[string][][]gjson.Result),
	}
	if len(tsidShardsArr) > 0 {
		tsidResult, status, err := ln.getMetricResultWithTsid(ctx, expr, query, logGroup.IndexPattern, logGroup.MustFilter, aggregationType, tsidShardsArr)
		if err != nil {
			return nil, status, err
		}
		mapResult = tsidResult
	}

	if len(labelsStrShardsArr) > 0 {
		status, err = ln.getMetricResultWithLabelsStr(ctx, expr, query, aggregationType, logGroup.MustFilter, labelsStrShardsArr,
			&mapResult)
		if err != nil {
			return nil, status, err
		}
	}
	return mapResult, http.StatusOK, nil
}

// 指标模型调用视图查询
// 通用处理函数： 获取数据视图的过滤条件信息 -> 构造 dsl -> 获取索引库下的所有索引以及对应的分片数 -> 执行 dsl
func (ln *LeafNodes) commonProcessForMetricModel(ctx context.Context, expr *parser.VectorSelector, query *interfaces.Query,
	aggregationType string) (interface{}, int, error) {

	var (
		// dataView = query.DataView
		status int
		err    error
	)

	mustFilters := []map[string]map[string]interface{}{}
	// if dataView.LogGroupFilters != "" {
	// 	mustFilters = []map[string]map[string]interface{}{
	// 		{
	// 			"query_string": {
	// 				"query":            dataView.LogGroupFilters,
	// 				"analyze_wildcard": true,
	// 			},
	// 		},
	// 	}
	// }

	// 区分__labels_str和tsid的索引模式列表
	tsidIndexPs := make([]string, 0)
	labelsStrIndexPs := make([]string, 0)
	// 当前是索引模式，当切换到视图时，用的是索引库类型，届时INDEX_BASE_SPLIT_TIME的key就是用索引库类型，无需拼接模式
	for _, indexP := range query.ViewQuery4Metric.BaseTypes {
		splitTime, ok := interfaces.INDEX_BASE_SPLIT_TIME[indexP]
		if !ok {
			// 不存在静态表中的索引库，直接扔给tsid查
			tsidIndexPs = append(tsidIndexPs, indexP)
		} else if splitTime.Before(time.UnixMilli(query.Start)) {
			// 分割点在查询开始时间之前，查tsid
			tsidIndexPs = append(tsidIndexPs, indexP)
		} else if splitTime.After(time.UnixMilli(query.End)) {
			labelsStrIndexPs = append(labelsStrIndexPs, indexP)
		} else {
			tsidIndexPs = append(tsidIndexPs, indexP)
			labelsStrIndexPs = append(labelsStrIndexPs, indexP)
		}
	}

	// 请求索引优化获取索引库下的所有索引以及对应的分片数
	// 获取索引库下的所有索引以及对应的分片数
	// 根据indexParttern得到的索引分片信息中无法得知索引所属的类型，所以__labels_str和tsid的索引库模型分开处理。

	labelsStrShardsArr, _, status, err := ln.dvService.GetIndices(ctx, labelsStrIndexPs, query.Start, query.End)
	if err != nil {
		return nil, status, uerrors.PromQLError{
			Typ: uerrors.ErrorInternal,
			Err: errors.New(err.Error()),
		}
	}

	tsidShardsArr, _, status, err := ln.dvService.GetIndices(ctx, tsidIndexPs, query.Start, query.End)
	if err != nil {
		return nil, status, uerrors.PromQLError{
			Typ: uerrors.ErrorInternal,
			Err: errors.New(err.Error()),
		}
	}

	if len(labelsStrShardsArr) == 0 && len(tsidShardsArr) == 0 {
		return MapResult{}, http.StatusOK, nil
	}

	mapResult := MapResult{
		LabelsMap:  make(map[string][]*labels.Label),
		TsValueMap: make(map[string][][]gjson.Result),
	}
	if len(tsidShardsArr) > 0 {
		// 从视图中获取到视图的过滤条件
		tsidResult, status, err := ln.getMetricResultWithTsid(ctx, expr, query, query.ViewQuery4Metric.BaseTypes, mustFilters, aggregationType, tsidShardsArr)
		if err != nil {
			return nil, status, err
		}
		mapResult = tsidResult
	}

	if len(labelsStrShardsArr) > 0 {
		status, err = ln.getMetricResultWithLabelsStr(ctx, expr, query, aggregationType, mustFilters, labelsStrShardsArr,
			&mapResult)
		if err != nil {
			return nil, status, err
		}
	}
	return mapResult, http.StatusOK, nil
}

// 通过labels_str获取数据，当符合分页条件时，labels_str也需分页。但是__labels_str不做分批查询，若是桶超限异常，则直接报错。
func (ln *LeafNodes) getMetricResultWithLabelsStr(ctx context.Context, expr *parser.VectorSelector,
	query *interfaces.Query, aggregationType string, mustFilter interface{},
	labelsStrShardsArr []*interfaces.IndexShards, mapResult *MapResult) (int, error) {
	// 构造 dsl
	groupBy := []string{interfaces.LABELS_STR}
	dsl, status, err := makeDSL(*expr, query, groupBy, aggregationType, mustFilter, false)
	if err != nil {
		o11y.Error(ctx, fmt.Sprintf("According to query[%v] to make dsl error: %v", query, err))

		return status, err
	}

	// 执行 dsl，并把各个shard的数据__labels_str聚合
	labelsStrResult, err := ln.ExecuteDslAndProcess(ctx, *dsl, labelsStrShardsArr, groupBy)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	if !query.IfNeedAllSeries && query.Limit != -1 && len(mapResult.LabelsMap) == 0 {
		// 分页，无tsid，按 __labels_str分页
		// 1. 先拿到所有的__labels_str对应的tsid。按tsid进行排序，按limit和offset取值append
		tsids := make([]string, 0)
		tmpMapResult := MapResult{
			LabelsMap:  make(map[string][]*labels.Label),
			TsValueMap: make(map[string][][]gjson.Result),
		}
		for k := range labelsStrResult.LabelsMap {
			lbs := parseLabelsStr(k, labelsStrResult.LabelsMap)
			// 把labels_str转成tsid
			md5Hasher := md5.New()
			md5Hasher.Write([]byte(k))
			hashed := md5Hasher.Sum(nil)
			tsid := hex.EncodeToString(hashed)

			tsids = append(tsids, tsid)
			tmpMapResult.LabelsMap[tsid] = lbs
			tmpMapResult.TsValueMap[tsid] = labelsStrResult.TsValueMap[k]
		}
		sort.Strings(tsids)
		length := len(tsids)
		mapResult.TotalSeries = length

		end := query.Offset + query.Limit
		if end > int64(length) {
			end = int64(len(tsids))
		}
		for i := query.Offset; i < end; i++ {
			mapResult.LabelsMap[tsids[i]] = tmpMapResult.LabelsMap[tsids[i]]
			mapResult.TsValueMap[tsids[i]] = tmpMapResult.TsValueMap[tsids[i]]
		}
	} else {
		for k := range labelsStrResult.LabelsMap {
			lbs := parseLabelsStr(k, labelsStrResult.LabelsMap)
			// 把labels_str转成tsid
			md5Hasher := md5.New()
			md5Hasher.Write([]byte(k))
			hashed := md5Hasher.Sum(nil)
			tsid := hex.EncodeToString(hashed)

			_, ok := mapResult.LabelsMap[tsid]
			if !query.IfNeedAllSeries && query.Limit != -1 {
				// 分页有tsid，存在的append，不存在的忽略
				if ok {
					mapResult.TsValueMap[tsid] = append(mapResult.TsValueMap[tsid], labelsStrResult.TsValueMap[k]...)
				}
			} else {
				// 非分页时，全append
				if !ok {
					// tsid存在了，就不需要append，不存在，则塞一个新的
					mapResult.LabelsMap[tsid] = lbs
					mapResult.TsValueMap[tsid] = labelsStrResult.TsValueMap[k]
				} else {
					// 存在，则把value，append进去
					mapResult.TsValueMap[tsid] = append(mapResult.TsValueMap[tsid], labelsStrResult.TsValueMap[k]...)
				}
			}
		}
		if mapResult.TotalSeries == 0 {
			mapResult.TotalSeries = len(mapResult.LabelsMap)
		}
	}
	return http.StatusOK, nil
}

// 通过tsid获取时序数据，返回的是tsid的各分片的数据
func (ln *LeafNodes) getMetricResultWithTsid(ctx context.Context, expr *parser.VectorSelector,
	query *interfaces.Query, baseTypes []string, mustFilter interface{}, aggregationType string,
	tsidShardsArr []*interfaces.IndexShards) (MapResult, int, error) {

	mapResults := MapResult{
		LabelsMap:  make(map[string][]*labels.Label),
		TsValueMap: make(map[string][][]gjson.Result),
	}

	// 如果是模型保存的请求(请求的数据时间范围为最近30m，步长为5m，在service处理时把取的序列size设置为了1)，直接查询，返回。
	// 参数主动设置直接查询时，直接查询，触碰到高基就报错。
	if query.IsModelRequest || query.IgnoringHCTS {
		return ln.getTsidResultsDirectly(ctx, expr, query, aggregationType, baseTypes, mustFilter, tsidShardsArr, TsidData{})
	}

	// 1. 获取指标在当前模型下的序列
	tsidData, status, err := ln.GetTsidsOfModelMetric(ctx, expr, query, baseTypes, mustFilter)
	if err != nil {
		return mapResults, status, err
	}

	// 3. 单批可请求的序列数 >= 目标序列数 && 不分页查询，则直接查询metric数据
	// 如果subInterval大于0，则用subInterval来计算桶数
	interval := query.Interval
	if query.SubIntervalWith2h > 0 {
		interval = query.SubIntervalWith2h
	}
	buckets := int64(math.Ceil(float64(query.End-query.Start) / float64(interval)))
	// 每批查询的序列数，最大查1w个桶
	batchSeriesSize := min(interfaces.PROMQL_BATCH_MAX_SERIES_SIZE, interfaces.DEFAULT_MAX_QUERY_POINTS/buckets)

	// 2. 计算offset和limit
	targetSeries := int64(len(tsidData.Tsids))
	if !query.IfNeedAllSeries && query.Limit != -1 {
		// 单指标操作 && 分页查询
		limit := query.Limit
		if targetSeries > limit {
			targetSeries = limit
		}
	}

	if targetSeries <= batchSeriesSize && query.Limit == -1 {
		// 不分页 && 目标序列数少，直接从metrics中过滤tsid以及查询指标数据
		return ln.getTsidResultsDirectly(ctx, expr, query, aggregationType, baseTypes, mustFilter, tsidShardsArr, tsidData)
	} else {
		// 按tsids分批次查询
		// 分批用tsid作为query条件来请求时序数据
		getMetricResultCtx, getMetricResultSpan := ar_trace.Tracer.Start(ctx, "Get Metric Data by tsid")
		defer getMetricResultSpan.End()
		// 如果查询分批大于等于2次，就并发查询
		allTsidsLen := int64(len(tsidData.Tsids))
		if query.Offset+targetSeries >= allTsidsLen {
			// 最后一个一页的请求数量需调整targetSeries
			targetSeries = allTsidsLen - query.Offset
			if targetSeries <= 0 {
				// 此分页无可查询的序列
				return mapResults, http.StatusOK, nil
			}
		}
		batch := int(math.Ceil(float64(targetSeries) / float64(batchSeriesSize)))
		if batch >= 2 {
			tsidChs := make(chan *MapResult, batch)
			defer close(tsidChs)
			// 按索引+分片数为单位并发，索引库下有多个索引。
			// 这块只要有一个协程发生异常，则可以中断所有协程返回。现在是等待所有的都执行结束，需优化。
			var wg sync.WaitGroup
			wg.Add(batch)
			for i := int64(0); i < int64(batch); i++ {
				size := batchSeriesSize
				if (i+1)*batchSeriesSize > targetSeries {
					size = targetSeries - i*batchSeriesSize
				}
				tsidi := tsidData.Tsids[query.Offset+i*batchSeriesSize : query.Offset+i*batchSeriesSize+size]
				// Submit tasks one by one.
				err := util.BatchSubmitPool.Submit(ln.batchSubmitTaskFuncWapper(getMetricResultCtx, expr, query,
					aggregationType, tsidShardsArr, baseTypes, mustFilter, TsidData{Tsids: tsidi, TsidsMap: tsidData.TsidsMap}, tsidChs, &wg))
				if err != nil {
					// 记录异常日志
					o11y.Error(ctx, fmt.Sprintf("ExecutePool.Submit error: %v", err))
					return mapResults, status, err
				}
			}
			// 等待所有执行结束
			wg.Wait()

			// 按 key 把合并各个 shard 的 map 为一个map
			labelsMap := make(map[string][]*labels.Label)
			tsValueMap := make(map[string][][]gjson.Result)

			for j := 0; j < batch; j++ {
				mapResulti := <-tsidChs
				if mapResulti.Err.Err != nil {
					return MapResult{}, 500, mapResulti.Err
				}
				for k, v := range mapResulti.LabelsMap {
					labelsMap[k] = v
					tsValueMap[k] = append(tsValueMap[k], mapResulti.TsValueMap[k]...) // 拿到时间序列，先放着待合并
				}
			}
			return MapResult{LabelsMap: labelsMap, TsValueMap: tsValueMap, TotalSeries: len(labelsMap)}, 200, nil
		} else {
			mapRes, status, err := ln.getTsidResultsByBatch(getMetricResultCtx, expr, query, aggregationType,
				tsidShardsArr, baseTypes, mustFilter,
				TsidData{Tsids: tsidData.Tsids[query.Offset : query.Offset+targetSeries],
					TsidsMap: tsidData.TsidsMap})
			mapRes.TotalSeries = len(tsidData.Tsids)
			return mapRes, status, err
		}
	}
}

// 从metrics中过滤并获取tsid的结果
func (ln *LeafNodes) getTsidResultsDirectly(ctx context.Context, expr *parser.VectorSelector,
	query *interfaces.Query, aggregationType string, baseTypes []string, mustFilter interface{},
	indexShardsArr []*interfaces.IndexShards, tsidData TsidData) (MapResult, int, error) {

	mapResults := MapResult{
		LabelsMap:  make(map[string][]*labels.Label),
		TsValueMap: make(map[string][][]gjson.Result),
	}

	// 构造 dsl
	groupBy := []string{interfaces.TSID}
	dsl, status, err := makeDSL(*expr, query, groupBy, aggregationType, mustFilter, true)
	if err != nil {
		o11y.Error(ctx, fmt.Sprintf("According to query[%v] to make dsl error: %v", query, err))
		return mapResults, status, err
	}

	// 执行 dsl，并把各个shard的数据 tsid 聚合
	mapResult, err := ln.ExecuteDslAndProcess(ctx, *dsl, indexShardsArr, groupBy)
	if err != nil {
		return mapResults, http.StatusInternalServerError, err
	}
	// 把tsid处理成labels。每批查询的序列是不同的，所以可以各自转
	status, err = ln.appendMapResult(ctx, mapResult, tsidData, expr, query, baseTypes, mustFilter, &mapResults)
	if err != nil {
		return mapResults, status, err
	}

	return mapResults, http.StatusOK, nil
}

// 分批次并发获取时序数据
func (ln *LeafNodes) batchSubmitTaskFuncWapper(ctx context.Context, expr *parser.VectorSelector,
	query *interfaces.Query, aggregationType string, indexShardsArr []*interfaces.IndexShards,
	indices []string, mustFilter interface{}, tsidData TsidData,
	mapResultChs chan<- *MapResult, wg *sync.WaitGroup) taskFunc {

	return func() {
		defer wg.Done()

		tsResult, _, err := ln.getTsidResultsByBatch(ctx, expr, query, aggregationType,
			indexShardsArr, indices, mustFilter, tsidData)
		// 错误处理
		if err != nil {
			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("getTsidResults error: %v", err))
			mapResultChs <- &MapResult{
				Err: uerrors.PromQLError{
					Typ: uerrors.ErrorInternal,
					Err: errors.New(err.Error()),
				},
			}
			return
		}

		mapResultChs <- &tsResult
	}
}

// 按tsid批量获取时序数据
func (ln *LeafNodes) getTsidResultsByBatch(ctx context.Context, expr *parser.VectorSelector, query *interfaces.Query,
	aggregationType string, indexShardsArr []*interfaces.IndexShards, baseTypes []string, mustFilter interface{}, tsidData TsidData) (MapResult, int, error) {

	mapResults := MapResult{
		LabelsMap:  make(map[string][]*labels.Label),
		TsValueMap: make(map[string][][]gjson.Result),
	}

	// 构造 dsl
	groupBy := []string{interfaces.TSID}
	dsl, status, err := makeBatchTsidDSL(*expr, query, groupBy, aggregationType, tsidData.Tsids)
	if err != nil {
		o11y.Error(ctx, fmt.Sprintf("According to query[%v] to make dsl error: %v", query, err))
		return mapResults, status, err
	}

	// 执行 dsl，并把各个shard的数据__labels_str聚合
	mapResult, err := ln.ExecuteDslAndProcess(ctx, *dsl, indexShardsArr, groupBy)
	if err != nil {
		return mapResults, http.StatusInternalServerError, err
	}
	// 把tsid处理成labels。每批查询的序列是不同的，所以可以各自转
	status, err = ln.appendMapResult(ctx, mapResult, tsidData, expr, query, baseTypes, mustFilter, &mapResults)
	if err != nil {
		return mapResults, status, err
	}

	return mapResults, http.StatusOK, nil
}

func (ln *LeafNodes) appendMapResult(ctx context.Context, mapResult MapResult, tsidData TsidData,
	expr *parser.VectorSelector, query *interfaces.Query, baseTypes []string, mustFilter interface{}, finalMapResults *MapResult) (int, error) {
	// 把tsid处理成labels。每批查询的序列是不同的，所以可以各自转
	hasGetTsidData := false
	for tsidi, labelsArr := range mapResult.LabelsMap {
		// 如果查询的tsid不在缓存中，翻译不出来，那么刷新metric对应的缓存。
		var lbs labels.Labels
		// 若是指标模型的请求，就无需获取序列把tsid翻译成labels
		if !query.IsModelRequest {
			if _, ok := tsidData.TsidsMap[tsidi]; !ok {
				// tsidi不存在于map中，则需要获取一次tsidData,无需获取多次。获取之后当前批次使用这个tsidData进行
				if !hasGetTsidData {
					// 如果找不到，那就刷请求时间范围内的tsid。默认用全量的时间范围，当这种情况刷新缓存时，用请求的时间范围
					tsidDataTmp, status, err := ln.GetTsidsOfModelMetric(ctx, expr, query, baseTypes, mustFilter)
					if err != nil {
						return status, err
					}
					tsidData = tsidDataTmp
					hasGetTsidData = true
				}
				lbs = tsidData.TsidsMap[tsidi]
			} else {
				lbs = tsidData.TsidsMap[tsidi]
			}
		} else {
			// 指标模型保存时的计算公式有效性检查，直接使用__tsid为维度来校验
			lbs = labelsArr
		}
		finalMapResults.LabelsMap[tsidi] = lbs
		finalMapResults.TsValueMap[tsidi] = append(finalMapResults.TsValueMap[tsidi],
			mapResult.TsValueMap[tsidi]...)
	}
	finalMapResults.TotalSeries = len(finalMapResults.LabelsMap)
	return http.StatusOK, nil
}

// 执行 dsl 语句，从 opensearch 中查询数据,并做初步处理。返回两个map, 分别按 keyStr 存储 labels 数组和 日期直方图下的 gjson.Result 数组
func (leafNodes *LeafNodes) ExecuteDslAndProcess(ctx context.Context, query bytes.Buffer,
	indexShardsArr []*interfaces.IndexShards, groupBy []string) (MapResult, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "并发请求 opensearch 查询数据")
	defer span.End()

	o11y.Info(ctx, fmt.Sprintf("Submit search dsl: %s", query.String()))

	// 计算所有索引的分片数之和
	var totalNum int
	for i := 0; i < len(indexShardsArr); i++ {
		pri, err := strconv.Atoi(indexShardsArr[i].Pri)
		if err != nil {
			span.SetStatus(codes.Error, "Index shard number convert error")
			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("Index shard number convert error: %v", err))

			// 如果异常 抛错
			return MapResult{}, uerrors.PromQLError{
				Typ: uerrors.ErrorInternal,
				Err: errors.New(err.Error()),
			}
		}
		totalNum += pri
	}
	span.SetAttributes(attribute.Key("index_number").Int(len(indexShardsArr)),
		attribute.Key("current_request_shard_total").Int(totalNum))

	// errchs := make(chan *uerrors.PromQLError, totalNum)
	mapResultChs := make(chan *MapResult, totalNum)
	// defer close(errchs)
	defer close(mapResultChs)
	// 按索引+分片数为单位并发，索引库下有多个索引。
	// 这块只要有一个协程发生异常，则可以中断所有协程返回。现在是等待所有的都执行结束，需优化。
	var wg sync.WaitGroup
	wg.Add(totalNum)
	for i := 0; i < len(indexShardsArr); i++ {
		pri, _ := strconv.Atoi(indexShardsArr[i].Pri)
		// Submit tasks one by one.
		for j := 0; j < pri; j++ {
			err := util.ExecutePool.Submit(leafNodes.searchSubmitTaskFuncWapper(ctx, j, []string{indexShardsArr[i].IndexName}, query, groupBy, mapResultChs, &wg))
			if err != nil {
				span.SetStatus(codes.Error, "ExecutePool.Submit error")
				// 记录异常日志
				o11y.Error(ctx, fmt.Sprintf("ExecutePool.Submit error: %v", err))

				return MapResult{}, uerrors.PromQLError{
					Typ: uerrors.ErrorExec,
					Err: errors.New("ExecutePool.Submit error: " + err.Error()),
				}
			}
		}
	}
	// 等待所有执行结束
	wg.Wait()

	// 按 key 把合并各个 shard 的 map 为一个map
	labelsMap := make(map[string][]*labels.Label)
	tsValueMap := make(map[string][][]gjson.Result)

	for j := 0; j < totalNum; j++ {
		mapResulti := <-mapResultChs
		if mapResulti.Err.Err != nil {
			return MapResult{}, mapResulti.Err
		}
		for k, v := range mapResulti.LabelsMap {
			labelsMap[k] = v
			tsValueMap[k] = append(tsValueMap[k], mapResulti.TsValueMap[k]...) // 拿到时间序列，先放着待合并
		}
	}

	span.SetStatus(codes.Ok, "")

	return MapResult{
		LabelsMap:  labelsMap,
		TsValueMap: tsValueMap,
	}, nil
}

// 并发请求opensearch获取时序数据
func (leafNodes *LeafNodes) searchSubmitTaskFuncWapper(ctx context.Context, shardId int, index []string, dsl bytes.Buffer,
	groupBy []string, mapResultChs chan<- *MapResult, wg *sync.WaitGroup) taskFunc {

	return func() {
		defer wg.Done()
		// todo: 记录opensearch返回的数据量
		ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("请求索引 %v 的 [%d] 分片的数据", index, shardId))
		span.SetAttributes(attribute.Key("index").String(index[0]),
			attribute.Key("shard_id").Int(shardId))
		defer span.End()

		res, _, err := leafNodes.dvService.GetDataFromOpenSearchWithBuffer(ctx, dsl, index,
			0, "_shards:"+strconv.Itoa(shardId))
		// res, _, err := leafNodes.osAccess.SearchSubmitWithBuffer(ctx, dsl, index, 0, "_shards:"+strconv.Itoa(shardId))

		// 错误处理
		if err != nil {
			span.SetStatus(codes.Error, "Opensearch search error")

			// 记录异常日志
			// o11y.Error(ctx, fmt.Sprintf("Opensearch search [%s] error: %v", dsl, err))
			mapResultChs <- &MapResult{
				Err: uerrors.PromQLError{
					Typ: uerrors.ErrorInternal,
					Err: errors.New(err.Error()),
				},
			}
			return
		}
		// 拿到之后先解析组装成map，然后再把各个shard上的map按key合并数组。
		// TsValueMap. key：shard的group by字段值拼接的字符串，value：keys下各个shard的ts values的数组， 用于后面的合并
		mapResult := &MapResult{
			LabelsMap:  make(map[string][]*labels.Label),
			TsValueMap: make(map[string][][]gjson.Result),
		}
		json := string(res)
		aggregations := gjson.Get(json, "aggregations")

		iteratorTermsAgg(aggregations, "", make([]*labels.Label, 0), 0,
			groupBy, mapResult)

		span.SetStatus(codes.Ok, "")

		mapResultChs <- mapResult
	}
}

// 返回指定日志分组id的日志库类型列表和过滤条件，错误状态码，错误信息。不支持多个日志分组id的查询
func (leafNodes *LeafNodes) GetLogGroupQueryFilters(ctx context.Context, logGroupId string) (interfaces.LogGroup, int, error) {
	if logGroupId == "" {
		o11y.Error(ctx, "missing ar_dataview parameter or ar_dataview parameter value cannot be empty")

		return interfaces.LogGroup{}, http.StatusBadRequest, uerrors.PromQLError{
			Typ: uerrors.ErrorBadData,
			Err: errors.New("missing ar_dataview parameter or ar_dataview parameter value cannot be empty"),
		}
	}

	loggroup, _, err := leafNodes.lgAccess.GetLogGroupQueryFilters(logGroupId)
	if err != nil {
		o11y.Error(ctx, fmt.Sprintf("Get dataview[%s] query filters error: %v", logGroupId, err))

		return loggroup, http.StatusInternalServerError, uerrors.PromQLError{
			Typ: uerrors.ErrorInternal,
			Err: err,
		}
	}

	return loggroup, http.StatusOK, nil
}

// 获取指标下的tsid
func (ln *LeafNodes) getTsidOfMetric(ctx context.Context, expr *parser.VectorSelector,
	query *interfaces.Query) (TsidData, int, error) {

	var (
		// dataView = query.DataView
		status int
		err    error
	)

	mustFilters := []map[string]map[string]interface{}{}
	// if dataView.LogGroupFilters != "" {
	// 	mustFilters = []map[string]map[string]interface{}{
	// 		{
	// 			"query_string": {
	// 				"query":            dataView.LogGroupFilters,
	// 				"analyze_wildcard": true,
	// 			},
	// 		},
	// 	}
	// }

	// 1. 获取指标在当前模型下的序列
	tsidData, status, err := ln.GetTsidsOfModelMetric(ctx, expr, query, query.ViewQuery4Metric.BaseTypes, mustFilters)
	if err != nil {
		return TsidData{}, status, err
	}

	return tsidData, http.StatusOK, nil
}

// 参考请求的开始结束时间来获取tsid
func getTsidOrTimeFilter(query *interfaces.Query, cacheKey string, fullCacheRefreshTime time.Duration) (TsidData, int64, int64, bool, bool) {
	start, end := query.Start, query.End
	fullFlag := true
	var tsidDataCache TsidData
	// 忽略缓存时不走缓存
	if !query.IgnoringMemoryCache {
		// 缓存存的是 modelId + exprId -> TsidData
		v, ok := Tsids_Of_Model_Metric_Map.Load(cacheKey)
		if ok {
			tsidDataCache = v.(TsidData)

			canUseCache := false
			if query.Start < tsidDataCache.StartTime.UnixMilli() &&
				query.End > tsidDataCache.EndTime.UnixMilli() {
				// 缓存缺失请求的前后两段，全量查
				start, end = query.Start, query.End
			} else if query.Start < tsidDataCache.StartTime.UnixMilli() &&
				query.End <= tsidDataCache.EndTime.UnixMilli() {
				// 缓存缺失开始段的时间，增量查
				fullFlag = false
				start, end = query.Start, tsidDataCache.StartTime.UnixMilli()
			} else if query.Start > tsidDataCache.StartTime.UnixMilli() &&
				query.End > tsidDataCache.EndTime.UnixMilli() {
				// 缓存缺失end段的时间。缺失的end到now的时长不超过10min，也可以尝试读缓存，
				// 但是这10min，数据查到，序列没找到，但是序列不会更新那么频繁，是小概率事件.所以可以走缓存，避免当前时间短暂变化引起的查询慢
				if time.Since(tsidDataCache.EndTime) > 10*time.Minute {
					// 增量查
					fullFlag = false
					start, end = tsidDataCache.EndTime.UnixMilli(), query.End
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
				fullDiff := time.Since(tsidDataCache.FullRefreshTime)
				if tsidDataCache.RefreshTime.UnixMilli() >= query.ModelUpdateTime &&
					tsidDataCache.RefreshTime.UnixMilli() >= query.DataView.UpdateTime &&
					len(query.Filters) == 0 {
					// 视图且模型的在缓存刷新之后未更新 且 请求没有过滤条件
					if fullCacheRefreshTime.Milliseconds() == 0 {
						fullCacheRefreshTime = 24 * time.Hour
					}
					if fullDiff <= fullCacheRefreshTime { // 24h改为全局配置
						// 全量缓存未到期，读取缓存
						return tsidDataCache, start, end, fullFlag, true
					} else {
						// 全量缓存到期，按缓存开始结束刷新缓存
						fullFlag = true
						start, end = tsidDataCache.StartTime.UnixMilli(), tsidDataCache.EndTime.UnixMilli()
					}
				}
			}
		}
	}

	return tsidDataCache, start, end, fullFlag, false
}
