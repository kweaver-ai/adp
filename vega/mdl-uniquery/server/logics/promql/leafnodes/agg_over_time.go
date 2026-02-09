// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package leafnodes

import (
	"context"
	"errors"
	"fmt"
	"math"
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

// agg_over_time 算子：构造dsl，指定分片查询，合并结果
func (ln *LeafNodes) AggOverTime(ctx context.Context, expr *parser.MatrixSelector, query *interfaces.Query, aggName string) (parser.Value, int, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Eval 叶子节点 AggOverTime")
	defer span.End()

	start := query.Start
	selRange := expr.Range.Milliseconds()
	if query.IsInstantQuery {
		// if range auto is true,range 等于请求参数中的delta，
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
			Err: fmt.Errorf("leafnodes.AggOverTime: invalid expression type %q", expr.VectorSelector.Type()),
		}
	}
	// fix range: if range auto is true, range = step
	if !query.IsInstantQuery && expr.Auto && expr.Range == 0 {
		selRange = query.Interval
	}
	// 获取range和interval的最大公约数。对于agg_over_time来说，30分钟路由和2h路由的subInterval是已用的，即两个值一样，
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
	result, status, err := ln.commonProcess(ctx, vs, &newQuery, aggName)
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
	mergeCtx, mergeSpan := ar_trace.Tracer.Start(ctx, "AggOverTime merge")
	defer mergeSpan.End()
	mat, err := aggOverTimeMerge(result.(MapResult), expr, &newQuery, aggName)
	if err != nil {
		// 记录异常的日志
		o11y.Error(mergeCtx, fmt.Sprintf("AggOverTime merge Error: %v", err))
		span.SetStatus(codes.Error, "AggOverTime merge Error")

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
func aggOverTimeMerge(mapResult MapResult, expr *parser.MatrixSelector, query *interfaces.Query, aggName string) (static.Matrix, error) {
	keys := make([]string, 0, len(mapResult.LabelsMap))
	for key := range mapResult.LabelsMap {
		keys = append(keys, key)
	}
	// 对 keys排序，从小到大
	sort.Strings(keys)

	if query.IsInstantQuery {
		return aggOverTimeMerge4InstantQuery(keys, mapResult, query, aggName)
	}
	return aggOverTimeMerge4RangeQuery(keys, mapResult, query, expr, aggName)
}

// instant query 串行合并
func aggOverTimeMerge4InstantQuery(keys []string, mapResult MapResult, query *interfaces.Query, aggName string) (static.Matrix, error) {
	mat := make(static.Matrix, 0, len(keys))
	for _, k := range keys {
		tsArr := mapResult.TsValueMap[k]
		start, end, totalNum := getPointsStatsInfo(tsArr)
		pointMap := make(map[int64]*static.AGGPoint, totalNum)

		// 遍历tsArr，合并每个数组的值，即把每个点的数据都放在目标map中，如果相同，则合并
		aggOverTimeMergePointWithSameKey(aggName, pointMap, tsArr)

		if len(pointMap) == 0 {
			return mat, nil
		}

		var finalPoint *static.AGGPoint
		// 先合并多个段为一个段
		for ts := start; ts <= end; ts += query.SubIntervalWith30min {
			if currentPoint, ok := pointMap[ts]; ok {
				if finalPoint != nil {
					mergeTwoPointWithSameTimeKey(aggName, currentPoint, finalPoint)
				} else {
					finalPoint = &static.AGGPoint{
						Count: currentPoint.Count,
						Value: currentPoint.Value,
					}
				}
			}
		}

		mat = append(mat, static.Series{
			Metric: parseLabelsStr(k, mapResult.LabelsMap),
			Points: []static.Point{
				{T: query.End, V: aggOverTimeFunction(aggName, finalPoint)},
			},
		})

	}

	return mat, nil

}

// 区间查询, 每个__labels_str下的聚合结果各自用一个goroutine去合并
func aggOverTimeMerge4RangeQuery(keys []string, mapResult MapResult, query *interfaces.Query, expr *parser.MatrixSelector, aggName string) (static.Matrix, error) {
	mat := make(static.Matrix, 0)
	chs := make(chan map[string]static.Series, len(mapResult.TsValueMap))
	defer close(chs)
	var wg sync.WaitGroup
	wg.Add(len(mapResult.TsValueMap))
	for _, k := range keys {
		tsArr := mapResult.TsValueMap[k]
		err := util.MegerPool.Submit(aggOverTimeMergeTaskFuncWrapper(chs, k, tsArr, mapResult.LabelsMap, query, &wg, expr, aggName))
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
		if _, exist := seriesMap[k]; exist {
			mat = append(mat, seriesMap[k])
		}
	}

	return mat, nil
}

// 各个labels下的合并逻辑
func aggOverTimeMergeTaskFuncWrapper(chs chan<- map[string]static.Series, k string, tsArr [][]gjson.Result,
	labelsMap map[string][]*labels.Label, query *interfaces.Query, wg *sync.WaitGroup, expr *parser.MatrixSelector, aggName string) taskFunc {

	return func() {
		defer wg.Done()

		// 先遍历tsArr去重
		pointMap := make(map[int64]*static.AGGPoint, 0)
		aggOverTimeMergePointWithSameKey(aggName, pointMap, tsArr)

		finalPoints := make([]static.Point, 0)
		selRangeTime := expr.Range.Milliseconds()
		tempPoints := make(map[int64]*static.AGGPoint, 1)
		seriesMap := make(map[string]static.Series)
		stepNum := selRangeTime / query.SubIntervalWith30min

		// aggOverTime 的查询涉及到 substep 的拆分，所以在合并时，开始时间和结束时间按 fixedStart 和 fixedEnd 为准
		// for ts := query.FixedStart; ts <= query.FixedEnd; ts += query.Interval {
		for ts := query.FixedStart; ts <= query.FixedEnd; ts = static.GetNextPointTime(*query, ts) {
			// 根据step映射的点的个数，对每个时间点做range处理，向右数 stepNum 个点。把他们的值相加
			for s := int64(0); s < stepNum; s++ {
				lastTs := ts + query.SubIntervalWith30min*s

				if _, exist := pointMap[lastTs]; !exist {
					continue
				}

				// 得到当前点的值，需把时间桶key相同的合并在一起
				currentPoint := pointMap[lastTs]
				if existPoint, ok := tempPoints[ts]; ok {
					mergeTwoPointWithSameTimeKey(aggName, currentPoint, existPoint)
				} else {
					tempPoints[ts] = &static.AGGPoint{
						Count: currentPoint.Count,
						Value: currentPoint.Value,
					}
				}
			}

			// 如果当前时间点对应的 range 没有数据点，继续下一个。
			if _, ok := tempPoints[ts]; !ok {
				continue
			}

			finalPoints = append(finalPoints, static.Point{
				T: ts,
				V: aggOverTimeFunction(aggName, tempPoints[ts]),
			})
			delete(tempPoints, ts)
		}

		seriesMap[k] = static.Series{
			Metric: parseLabelsStr(k, labelsMap),
			Points: finalPoints,
		}

		chs <- seriesMap
	}
}

// 把时间桶的 key 相同的点合并成一个点
func mergeTwoPointWithSameTimeKey(aggName string, cuurentPoint *static.AGGPoint, existPoint *static.AGGPoint) {
	// sum avg count 需要相加。max 和 min 需要比较取较大或较小
	existPoint.Count += cuurentPoint.Count
	switch aggName {
	case interfaces.MAX_OVER_TIME:
		// 取大
		if cuurentPoint.Value > existPoint.Value || math.IsNaN(existPoint.Value) {
			existPoint.Value = cuurentPoint.Value
		}
	case interfaces.MIN_OVER_TIME:
		// 取小
		if cuurentPoint.Value < existPoint.Value || math.IsNaN(existPoint.Value) {
			existPoint.Value = cuurentPoint.Value
		}
	case interfaces.AVG_OVER_TIME, interfaces.SUM_OVER_TIME, interfaces.COUNT_OVER_TIME:
		// 求和
		existPoint.Value += cuurentPoint.Value
	}
}

// 基于合并完之后的数据计算相应的聚合值。avg_over_time = sum(sum) / sum(count)，其余的聚合函数直接取值。
func aggOverTimeFunction(aggName string, tempPoint *static.AGGPoint) float64 {
	var value float64
	switch aggName {
	case interfaces.AVG_OVER_TIME:
		// avg = 总和/总计数
		value = float64(tempPoint.Value / float64(tempPoint.Count))
	case interfaces.SUM_OVER_TIME, interfaces.MAX_OVER_TIME, interfaces.MIN_OVER_TIME, interfaces.COUNT_OVER_TIME:
		value = tempPoint.Value
	}
	return value
}

// 按时间窗去重k个数组（即k个分片）中的样本点
func aggOverTimeMergePointWithSameKey(aggName string, pointMap map[int64]*static.AGGPoint, tsArr [][]gjson.Result) {
	// tsArr 数组是各个分片上的时间分桶和值的列表
	for _, tsArri := range tsArr {
		for _, pointij := range tsArri {
			currentT := pointij.Get("key").Int()
			currentCount := pointij.Get("doc_count").Int()
			currentValue := pointij.Get("value.value").Float()

			// 得到当前点的值，需把时间桶key相同的合并在一起
			if existPoint, ok := pointMap[currentT]; ok {
				mergeTwoPointWithSameTimeKey(aggName,
					&static.AGGPoint{
						Count: currentCount,
						Value: currentValue,
					}, existPoint)
			} else {
				pointMap[currentT] = &static.AGGPoint{
					Count: currentCount,
					Value: currentValue,
				}
			}
		}
	}
}
