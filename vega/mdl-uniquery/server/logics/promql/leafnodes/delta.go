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
	"sync"
	"time"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/tidwall/gjson"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	uerrors "uniquery/errors"
	"uniquery/interfaces"
	"uniquery/logics/promql/labels"
	"uniquery/logics/promql/parser"
	"uniquery/logics/promql/static"
	"uniquery/logics/promql/util"
)

// delta 算子：构造dsl，指定分片查询，合并结果
func (ln *LeafNodes) DeltaAggs(ctx context.Context, expr *parser.MatrixSelector, query *interfaces.Query, call FunctionCall) (parser.Value, int, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Eval 叶子节点 Delta")
	defer span.End()

	start := query.Start
	selRange := expr.Range.Milliseconds()
	if query.IsInstantQuery {
		// if range auto is true,range 等于请求参数中的 lookBackDelta，
		if expr.Auto && expr.Range == 0 {
			// selRange = convert.GetLookBackDelta(query.LookBackDelta, ln.appSetting.PromqlSetting.LookbackDelta)
			selRange = query.End - query.Start
		}
		// 结束时间 - range
		start = query.End - selRange
		query.Interval = selRange
	}
	vs, ok := expr.VectorSelector.(*parser.VectorSelector)
	if !ok {
		return nil, http.StatusUnprocessableEntity, uerrors.PromQLError{
			Typ: uerrors.ErrorExec,
			Err: fmt.Errorf("leafnodes.Delta: invalid expression type %q", expr.VectorSelector.Type()),
		}
	}
	// fix range: if range auto is true, range = step
	if !query.IsInstantQuery && expr.Auto && expr.Range == 0 {
		selRange = query.Interval
	}
	query.SubIntervalWith30min = getCommonDivisor(selRange, query.Interval)
	query.SubIntervalWith2h = query.SubIntervalWith30min

	newQuery := *query
	newQuery.Start = start
	expr.Range = time.Duration(selRange) * time.Millisecond

	span.SetAttributes(attribute.Key("start").Int64(start),
		attribute.Key("subInterval30min").Int64(query.SubIntervalWith30min),
		attribute.Key("subInterval2h").Int64(query.SubIntervalWith2h),
	)

	// 通用处理： 获取日志分组的索引信息 -> 构造 dsl -> 获取索引库下的所有索引以及对应的分片数 -> 执行 dsl
	result, status, err := ln.commonProcess(ctx, vs, &newQuery, interfaces.DELTA_AGG)
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
	mergeCtx, mergeSpan := ar_trace.Tracer.Start(ctx, "Delta merge")
	defer mergeSpan.End()

	mat, err := deltaMerge(result.(MapResult), expr, &newQuery, call)
	if err != nil {
		// 记录异常的日志
		o11y.Error(mergeCtx, fmt.Sprintf("Delta merge Error: %v", err))
		span.SetStatus(codes.Error, "Delta merge Error")

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

// instant query 串行合并, range query 起协程同时合并
func deltaMerge(mapResult MapResult, expr *parser.MatrixSelector, query *interfaces.Query, call FunctionCall) (static.Matrix, error) {
	keys := make([]string, 0, len(mapResult.LabelsMap))
	for key := range mapResult.LabelsMap {
		keys = append(keys, key)
	}
	// 对 keys排序，从小到大
	sort.Strings(keys)

	if query.IsInstantQuery {
		return deltaMerge4InstantQuery(keys, mapResult, query, expr, call)
	}
	return deltaMerge4RangeQuery(keys, mapResult, query, expr, call)
}

// instant query 串行合并
func deltaMerge4InstantQuery(keys []string, mapResult MapResult, query *interfaces.Query, e *parser.MatrixSelector,
	call FunctionCall) (static.Matrix, error) {

	mat := make(static.Matrix, 0, len(keys))
	for _, k := range keys {
		tsArr := mapResult.TsValueMap[k]
		start, end, totalNum := getPointsStatsInfo(tsArr)
		pointMap := make(map[int64]*static.DeltaPoint, totalNum)

		// 遍历tsArr，合并每个分片数组的值，即把每个点的数据都放在目标map中，如果时间 key 相同，则合并
		deltaMergePointWithSameKey(pointMap, tsArr)

		if len(pointMap) == 0 {
			return mat, nil
		}

		// instant 合并多个段为一个段
		// 对于instant query， delta 聚合只需要取第一个和最后一个样本点值，并且把样本点的时间替换成query.end
		tempPoints := map[int64]*static.DeltaPoint{
			end: {
				FirstTimestamp: pointMap[start].FirstTimestamp,
				FirstValue:     pointMap[start].FirstValue,
				LastTimestamp:  pointMap[end].LastTimestamp,
				LastValue:      pointMap[end].LastValue,
				PointsCount:    pointMap[start].PointsCount + pointMap[end].PointsCount,
			},
		}

		finalPoints := make([]static.Point, 0)
		deltaFunction(&tempPoints, &finalPoints, end, query, e, call)
		if len(finalPoints) == 0 {
			continue
		}

		mat = append(mat, static.Series{
			Metric: parseLabelsStr(k, mapResult.LabelsMap),
			Points: finalPoints,
		})

	}

	return mat, nil

}

// 区间查询, 每个__labels_str下的聚合结果各自用一个goroutine去合并
func deltaMerge4RangeQuery(keys []string, mapResult MapResult, query *interfaces.Query, expr *parser.MatrixSelector, call FunctionCall) (static.Matrix, error) {
	mat := make(static.Matrix, 0)
	chs := make(chan map[string]static.Series, len(mapResult.TsValueMap))
	defer close(chs)
	var wg sync.WaitGroup
	wg.Add(len(mapResult.TsValueMap))
	// 按序列并发
	for _, k := range keys {
		tsArr := mapResult.TsValueMap[k]
		err := util.MegerPool.Submit(deltaMergeTaskFuncWrapper(chs, k, tsArr, mapResult.LabelsMap, query, &wg, expr, call))
		if err != nil {
			return nil, err
		}
	}
	wg.Wait()

	seriesMap := make(map[string]static.Series)
	for j := 0; j < len(mapResult.TsValueMap); j++ {
		series := <-chs

		for k, v := range series {
			if len(v.Points) == 0 {
				continue
			}
			seriesMap[k] = v
		}
	}

	if len(seriesMap) == 0 {
		return mat, nil
	}

	for _, k := range keys {
		if _, exist := seriesMap[k]; exist {
			mat = append(mat, seriesMap[k])
		}
	}

	return mat, nil
}

// 各个labels下的合并逻辑
func deltaMergeTaskFuncWrapper(chs chan<- map[string]static.Series, k string, tsArr [][]gjson.Result, labelsMap map[string][]*labels.Label,
	query *interfaces.Query, wg *sync.WaitGroup, expr *parser.MatrixSelector, call FunctionCall) taskFunc {

	return func() {
		defer wg.Done()

		// 先遍历tsArr去重
		pointMap := make(map[int64]*static.DeltaPoint, 0)
		deltaMergePointWithSameKey(pointMap, tsArr)

		finalPoints := make([]static.Point, 0)
		selRangeTime := expr.Range.Milliseconds()
		tempPoints := make(map[int64]*static.DeltaPoint, 1)
		seriesMap := make(map[string]static.Series)
		stepNum := selRangeTime / query.SubIntervalWith30min

		// delta 的查询涉及到 substep 的拆分，所以在合并时，开始时间和结束时间按 fixedStart 和 fixedEnd 为准
		// for ts := query.FixedStart; ts <= query.FixedEnd; ts += query.Interval {
		for ts := query.FixedStart; ts <= query.FixedEnd; ts = static.GetNextPointTime(*query, ts) {
			// 根据step映射的点的个数，对每个时间点做range处理，向右数 stepNum 个点。
			// delta: 取 range 内的第一个点和最后一个点，然后再合并这两个点为一个点（first 为第一个点的 first，last 为最后一个点的 last），再计算 delta 值

			// 找 range 内的第一个点和最后一个点，以及累计 docCount, 用于 delta 值的系数的计算。
			for s := int64(0); s < stepNum; s++ {
				lastTs := ts + query.SubIntervalWith30min*s

				if _, exist := pointMap[lastTs]; !exist {
					continue
				}
				compareDeltaPoints(tempPoints, ts, pointMap[lastTs])
			}

			// 计算 delta 值
			deltaFunction(&tempPoints, &finalPoints, ts, query, expr, call)
			delete(tempPoints, ts)
		}

		seriesMap[k] = static.Series{
			Metric: parseLabelsStr(k, labelsMap),
			Points: finalPoints,
		}

		chs <- seriesMap
	}
}

// 按时间窗去重k个数组（即k个分片）中的样本点, key 为 timestamp, 多个分片原始查询结果按 firstPoint 和 lastPoint 的逻辑合并在一起
func deltaMergePointWithSameKey(pointMap map[int64]*static.DeltaPoint, tsArr [][]gjson.Result) {
	// tsArr 数组是各个分片上的时间分桶和值的列表
	for _, tsArri := range tsArr {
		// 分片 i 的时间桶 key 和 value
		for _, pointij := range tsArri {
			currentT := pointij.Get("key").Int()

			currentPoint := &static.DeltaPoint{
				FirstTimestamp: pointij.Get("value.firstTimestamp").Int(),
				FirstValue:     pointij.Get("value.firstValue").Float(),
				LastTimestamp:  pointij.Get("value.lastTimestamp").Int(),
				LastValue:      pointij.Get("value.lastValue").Float(),
				PointsCount:    pointij.Get("doc_count").Int(),
			}
			compareDeltaPoints(pointMap, currentT, currentPoint)
		}
	}
}

// 对两个point进行比较，最终输出最大的样本点和最小的样本点
func compareDeltaPoints(pointMap map[int64]*static.DeltaPoint, ts int64, currentPoint *static.DeltaPoint) {
	if existPoint, ok := pointMap[ts]; ok {

		// 计算时间区间，即 firstT 和 lastT
		if existPoint.FirstTimestamp > currentPoint.FirstTimestamp ||
			(existPoint.FirstTimestamp == currentPoint.FirstTimestamp && existPoint.FirstValue < currentPoint.FirstValue) {
			// 如果 firstT 不一样，取其中时间最小的那个样本；如果时间戳一样，取值最大的那个样本。
			existPoint.FirstTimestamp = currentPoint.FirstTimestamp
			existPoint.FirstValue = currentPoint.FirstValue
		}

		if existPoint.LastTimestamp < currentPoint.LastTimestamp ||
			(existPoint.LastTimestamp == currentPoint.LastTimestamp && existPoint.LastValue < currentPoint.LastValue) {
			// 如果 lastT 不一样，取其中时间最大的那个样本；如果时间戳一样，取值最大的那个样本。
			existPoint.LastTimestamp = currentPoint.LastTimestamp
			existPoint.LastValue = currentPoint.LastValue
		}
		existPoint.PointsCount += currentPoint.PointsCount

	} else {
		pointMap[ts] = &static.DeltaPoint{
			FirstTimestamp: currentPoint.FirstTimestamp,
			FirstValue:     currentPoint.FirstValue,
			LastTimestamp:  currentPoint.LastTimestamp,
			LastValue:      currentPoint.LastValue,
			PointsCount:    currentPoint.PointsCount,
		}
	}
}

// 基于合并完之后的数据计算相应的值。
func deltaFunction(tempPoints *map[int64]*static.DeltaPoint, finalPoints *[]static.Point, ts int64, query *interfaces.Query,
	e *parser.MatrixSelector, call FunctionCall) *[]static.Point {

	inArgs := make([]parser.Value, 1)

	// 如果当前时间点对应的 range 没有数据点，跳过计算 delta 值
	if _, ok := (*tempPoints)[ts]; !ok {
		return finalPoints
	}

	win := (*tempPoints)[ts]
	inArgs[0] = static.DeltaPoint{
		FirstTimestamp: win.FirstTimestamp,
		FirstValue:     win.FirstValue,
		LastTimestamp:  win.LastTimestamp,
		LastValue:      win.LastValue,
		PointsCount:    win.PointsCount,
	}

	enh := &static.EvalNodeHelper{
		Ts:  ts,
		Out: make(static.Vector, 0, 1),
	}

	args := parser.Expressions{e}
	vec := call.Call(inArgs, args, enh)
	if len(vec) > 0 {
		if query.IsInstantQuery {
			ts = query.End
		}

		*finalPoints = append(*finalPoints, static.Point{T: ts, V: vec[0].V})
	}

	return finalPoints
}
