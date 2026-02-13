// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package leafnodes

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/tidwall/gjson"
	"go.opentelemetry.io/otel/codes"

	uerrors "uniquery/errors"
	"uniquery/interfaces"
	"uniquery/logics/promql/labels"
	"uniquery/logics/promql/parser"
	"uniquery/logics/promql/static"
	"uniquery/logics/promql/util"
)

// 根据查询语句和请求参数从 opensearch 中获取结果，返回结果集 Matrix 和错误信息
func (leafNodes *LeafNodes) EvalVectorSelector(ctx context.Context, expr *parser.VectorSelector, groupBy []string, aggregationType string,
	query *interfaces.Query) (parser.Value, int, error) {
	// 构造 time range.， 如果是 query接口，即 start == end, interval也应该用成倒推的时间区间。对于fixedStart 和fixedEnd也需重新计算
	// 对于 rate，irate等算子，有时间区间的，start倒推evalRange；对于 evalRange == 0 时，取配置的LookBackDelta
	// query 是指针，瞬时查询修改了 start，一个语句中涉及到多个指标的查询时，会把 start 时间会被修改多次。所以在实际查询的地方，用新的
	ctx, span := ar_trace.Tracer.Start(ctx, "Eval 叶子节点 VectorSelector")
	defer span.End()

	// start := query.Start
	if query.IsInstantQuery {
		// lookBackDelta := convert.GetLookBackDelta(query.LookBackDelta, leafNodes.appSetting.PromqlSetting.LookbackDelta)

		// start = query.Start - lookBackDelta
		// query.Interval = lookBackDelta
		query.Interval = query.End - query.Start
	}

	newQuery := *query
	// newQuery.Start = start
	newQuery.IsPersistMetric = interfaces.IsPersistMetric(expr.Name)

	// 通用处理： 获取日志分组的索引信息 -> 构造 dsl -> 获取索引库下的所有索引以及对应的分片数 -> 执行 dsl
	result, status, err := leafNodes.commonProcess(ctx, expr, &newQuery, aggregationType)
	if err != nil {
		// 记录异常的日志
		o11y.Error(ctx, fmt.Sprintf("Common Process Error: %v", err))

		return nil, status, err
	}
	matrixResult, ok := result.(static.Matrix)
	if ok {
		return matrixResult, status, err
	}

	// 合并结果
	mergeCtx, mergeSpan := ar_trace.Tracer.Start(ctx, "Sampling merge")
	defer mergeSpan.End()
	mat, err := samplingMerge(result.(MapResult), &newQuery)
	if err != nil {
		// 记录异常的日志
		o11y.Error(mergeCtx, fmt.Sprintf("Sampling merge Error: %v", err))
		span.SetStatus(codes.Error, "Sampling merge Error")

		return nil, http.StatusUnprocessableEntity, uerrors.PromQLError{
			Typ: uerrors.ErrorExec,
			Err: errors.New(err.Error()),
		}
	}
	mergeSpan.SetStatus(codes.Ok, "")

	if query.IfNeedAllSeries {
		return mat, http.StatusOK, nil
	}
	span.SetStatus(codes.Ok, "")
	return static.PageMatrix{Matrix: mat, TotalSeries: result.(MapResult).TotalSeries}, http.StatusOK, nil
}

// 采样聚合时的合并
func samplingMerge(mapResult MapResult, query *interfaces.Query) (static.Matrix, error) {
	// 对 map 的 key 排序，使得同一个请求输出的结果不会乱序
	keys := []string{}
	for key := range mapResult.LabelsMap {
		keys = append(keys, key)
	}
	// 对 keys 排序，从小到大
	sort.Strings(keys)

	// 如果是instant query,不用协程并发处理.串行处理。
	// 对于instant query， 采样聚合只需要最后一个样本点值，那么直接取最后一个point的值，并且把样本点的时间替换成query.end
	if query.IsInstantQuery {
		return mergeSamples4InstantQuery(keys, mapResult, query)
	}

	// 对于query_range, 用协程池提交各个分片数据的合并
	return mergeSamples4RangeQuery(mapResult, keys, query)
}

func mergeSamples4RangeQuery(mapResult MapResult, keys []string, query *interfaces.Query) (static.Matrix, error) {
	mat := make(static.Matrix, 0)

	chs := make(chan map[string]static.Series, len(mapResult.TsValueMap))
	defer close(chs)
	var wg sync.WaitGroup
	wg.Add(len(mapResult.TsValueMap))
	for _, k := range keys {
		tsArr := mapResult.TsValueMap[k]
		// 用协程池提交执行的操作
		err := util.MegerPool.Submit(samplingMergeTaskFuncWapper(chs, k, tsArr, mapResult.LabelsMap, query, &wg))
		if err != nil {
			return nil, err
		}
	}
	wg.Wait()

	seriesMap := make(map[string]static.Series)
	for j := 0; j < len(mapResult.TsValueMap); j++ {
		series := <-chs
		for k, v := range series {
			seriesMap[k] = v
		}
	}

	for _, k := range keys {
		mat = append(mat, seriesMap[k])
	}

	return mat, nil
}

func mergeSamples4InstantQuery(keys []string, mapResult MapResult, query *interfaces.Query) (static.Matrix, error) {
	mat := make(static.Matrix, 0)
	for _, k := range keys {
		tsArr := mapResult.TsValueMap[k]
		_, end, totalNum := getPointsStatsInfo(tsArr)

		// 2. 遍历 tsArr，合并每个数组的值,即把每个点的数据都放在目标map中，如果相同，则合并
		pointMap := make(map[int64]*static.Point, totalNum)
		duplicatePoint(pointMap, tsArr)
		// 对于instant query， 采样聚合只需要最后一个样本点值，那么直接取最后一个point的值，并且把样本点的时间替换成query.end
		mat = append(mat, static.Series{
			Metric: parseLabelsStr(k, mapResult.LabelsMap),
			Points: []static.Point{
				{
					T: query.End,
					V: pointMap[end].V,
				},
			},
		})
	}
	return mat, nil
}

// 通过拿到实际结果的最大最小样本点，结合step，可以按step的步长构造结果集，从k个数组中取值放入返回数组中。
func samplingMergeTaskFuncWapper(chs chan<- map[string]static.Series, k string, tsArr [][]gjson.Result, labelsMap map[string][]*labels.Label,
	query *interfaces.Query, wg *sync.WaitGroup) taskFunc {

	return func() {
		defer wg.Done()
		// 1. 从k个数组中选取时间最小的作为开始时间, 选取时间最大的作为结束时间
		start, end, totalNum := getPointsStatsInfo(tsArr)

		// 2. 遍历 tsArr，合并每个数组的值,即把每个点的数据都放在目标map中，如果相同，则合并
		pointMap := make(map[int64]*static.Point, totalNum)
		duplicatePoint(pointMap, tsArr)
		// 3. 从start到end遍历，补充缺失的数据点
		var resPoints = make([]static.Point, 0, (end-start)/query.Interval+1)
		seriesMap := make(map[string]static.Series)
		// for ts := start; ts <= end; ts += query.Interval {
		for ts := start; ts <= end; ts = static.GetNextPointTime(*query, ts) {
			if point, ok := pointMap[ts]; ok {
				// append 到结果数组中，并把point的时间换成时间窗的时间
				resPoints = append(resPoints, static.Point{
					T: ts,
					V: point.V,
				})
			} else {
				// 利用step来看是否存在断了的时间窗，有就用上一个时间窗的值
				// 如果是查询持久化后的数据，都无需补点。若是查询频率大于等于5m，都不补
				if query.Interval < 300000 && !query.IsPersistMetric && !query.NotNeedFilling {
					resPoints = append(resPoints, static.Point{
						T: ts,
						V: resPoints[len(resPoints)-1].V,
					})
				}
			}
		}

		seriesMap[k] = static.Series{
			Metric: parseLabelsStr(k, labelsMap),
			Points: resPoints,
		}

		chs <- seriesMap
	}
}

// 获取指标的维度字段
func (leafNodes *LeafNodes) EvalVectorSelectorFields(ctx context.Context, expr *parser.VectorSelector,
	query *interfaces.Query, fieldName string) (map[string]bool, int, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "Eval 叶子节点 VectorSelector 的字段")
	defer span.End()

	// 通用处理： 获取日志分组的索引信息 -> 构造 dsl -> 获取索引库下的所有索引以及对应的分片数 -> 执行 dsl
	tsid, status, err := leafNodes.getTsidOfMetric(ctx, expr, query)
	if err != nil {
		// 记录异常的日志
		o11y.Error(ctx, fmt.Sprintf("Common Process Error: %v", err))

		return nil, status, err
	}

	// 从tsid中获取字段集
	// 遍历每个序列的字段集，对字段名去重得到指标的字段结构
	fieldMap := make(map[string]bool, 0)
	if fieldName == "" {
		// 字段名为空时，表示拿的是指标的字段
		for _, labels := range tsid.TsidsMap {
			for _, labelName := range labels {
				fieldMap[labelName.Name] = true
			}
		}
	} else {
		// 字段不为空，表示拿的是指定字段的值列表
		name := strings.Replace(fieldName, interfaces.LABELS_PREFIX+".", "", -1)
		for _, labels := range tsid.TsidsMap {
			for _, labelName := range labels {
				if labelName.Name == name {
					fieldMap[labelName.Value] = true
				}
			}
		}
	}

	span.SetStatus(codes.Ok, "")
	return fieldMap, http.StatusOK, nil
}
