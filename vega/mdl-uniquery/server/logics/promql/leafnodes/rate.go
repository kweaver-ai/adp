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

type FunctionCall = static.FunctionCall

// rate 算子：构造dsl，指定分片查询，合并结果
func (ln *LeafNodes) RateAggs(ctx context.Context, expr *parser.MatrixSelector, query *interfaces.Query, call FunctionCall) (parser.Value, int, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Eval 叶子节点 Rate")
	defer span.End()

	// rate有时间区间, 对于instant query, start倒推evalRange
	v, newQuery, status, err := processParam(expr, query)
	if err != nil {
		return nil, status, err
	}

	span.SetAttributes(attribute.Key("start").Int64(query.Start),
		attribute.Key("subInterval30min").Int64(newQuery.SubIntervalWith30min),
		attribute.Key("subInterval2h").Int64(newQuery.SubIntervalWith2h),
	)

	// 通用处理： 获取日志分组的索引信息 -> 构造 dsl -> 获取索引库下的所有索引以及对应的分片数 -> 执行 dsl
	result, status, err := ln.commonProcess(ctx, v, newQuery, interfaces.RATE_AGG)
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
	mergeCtx, mergeSpan := ar_trace.Tracer.Start(ctx, "Rate merge")
	defer mergeSpan.End()
	mat, err := rateMerge(result.(MapResult), expr, newQuery, call)
	if err != nil {
		// 记录异常的日志
		o11y.Error(mergeCtx, fmt.Sprintf("Rate merge Error: %v", err))
		span.SetStatus(codes.Error, "Rate merge Error")

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

// rate需要计算counterCorrection, 如果 step > 路由，那么对于同一个索引的数据，
// opensearch按step划窗的数据, 会落在超过2个分片上，会造成counterCorrection计算不准确
// 因此需要拆分substep,可参考文档里 increase 算子解释
// https://confluence.aishu.cn/pages/viewpage.action?pageId=122325570
func calcSubInterval(interval, stepRange time.Duration, routing time.Duration) time.Duration {
	var subInterval time.Duration
	// 最多循环120次
	for subInterval = routing; subInterval > 0; subInterval -= 1 * time.Minute {
		if interval%subInterval == 0 && stepRange%subInterval == 0 {
			break
		}
	}

	return subInterval
}

// instant query 串行合并, range query 起协程同时合并
func rateMerge(mapResult MapResult, e *parser.MatrixSelector, query *interfaces.Query, call FunctionCall) (static.Matrix, error) {
	chs := make(chan map[string]static.Series, len(mapResult.TsValueMap))
	defer close(chs)

	keys := make([]string, 0, len(mapResult.LabelsMap))
	for key := range mapResult.LabelsMap {
		keys = append(keys, key)
	}
	// 对 keys排序，从小到大
	sort.Strings(keys)

	if query.IsInstantQuery {
		return instantQueryMerge(keys, mapResult, query, e, call)
	}

	return rangeQueryMerge(chs, keys, mapResult, query, e, call)

}

// instant query 串行合并
func instantQueryMerge(keys []string, mapResult MapResult, query *interfaces.Query, e *parser.MatrixSelector, call FunctionCall) (static.Matrix, error) {
	mat := make(static.Matrix, 0, len(keys))
	for _, k := range keys {
		tsArr := mapResult.TsValueMap[k]
		start, end, totalNum := getPointsStatsInfo(tsArr)
		pointMap := make(map[int64]*static.RatePoint, totalNum)

		// 遍历tsArr，合并每个数组的值，即把每个点的数据都放在目标map中，如果相同，则合并
		mergePointWithSameKeyTS(&pointMap, tsArr)

		if len(pointMap) == 0 {
			return mat, nil
		}

		var (
			docCount          int64
			counterCorrection float64
		)

		for ts := start; ts <= end; ts += query.SubIntervalWith30min {
			if point, ok := pointMap[ts]; ok {
				if ts > start {
					for lastTs := ts - query.SubIntervalWith30min; lastTs >= start; lastTs -= query.SubIntervalWith30min {
						// 所以当前一个不存在时，需继续循环再向前取一个窗口，知道取到为止。
						if _, exist := pointMap[lastTs]; !exist {
							continue
						}
						// 取到了比较，然后退出循环
						if point.FirstValue < pointMap[lastTs].LastValue {
							counterCorrection += pointMap[lastTs].LastValue
						}
						break
					}
				}
				counterCorrection += point.CounterCorrection
				docCount += point.PointsCount
			}
		}

		tempPoints := map[int64]*static.RatePoint{
			end: {
				FirstTimestamp:    pointMap[start].FirstTimestamp,
				FirstValue:        pointMap[start].FirstValue,
				LastTimestamp:     pointMap[end].LastTimestamp,
				LastValue:         pointMap[end].LastValue,
				CounterCorrection: counterCorrection,
				PointsCount:       docCount,
			},
		}

		finalPoints := make([]static.Point, 0)
		rateFunction(&tempPoints, &finalPoints, end, query, e, call)
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
func rangeQueryMerge(chs chan map[string]static.Series, keys []string, mapResult MapResult, query *interfaces.Query, e *parser.MatrixSelector, call FunctionCall) (static.Matrix, error) {
	mat := make(static.Matrix, 0)

	var wg sync.WaitGroup
	wg.Add(len(mapResult.TsValueMap))
	for _, k := range keys {
		tsArr := mapResult.TsValueMap[k]
		err := util.MegerPool.Submit(rateMergeTaskFuncWrapper(chs, k, tsArr, mapResult.LabelsMap, query, &wg, e, call))
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
func rateMergeTaskFuncWrapper(chs chan<- map[string]static.Series, k string, tsArr [][]gjson.Result,
	labelsMap map[string][]*labels.Label, query *interfaces.Query, wg *sync.WaitGroup, e *parser.MatrixSelector, call FunctionCall) taskFunc {
	return func() {
		defer wg.Done()

		// 遍历tsArr去重
		pointMap := make(map[int64]*static.RatePoint, 0)
		mergePointWithSameKeyTS(&pointMap, tsArr)

		finalPoints := make([]static.Point, 0)
		selRangeTime := e.Range.Milliseconds()
		tempPoints := make(map[int64]*static.RatePoint, 1)
		seriesMap := make(map[string]static.Series)
		stepNum := selRangeTime / query.SubIntervalWith30min

		// 计算每个点的 rate 值
		// for ts := query.FixedStart; ts <= query.FixedEnd; ts += query.Interval {
		for ts := query.FixedStart; ts <= query.FixedEnd; ts = static.GetNextPointTime(*query, ts) {
			// 根据step映射的点的个数，对每个时间点做range处理，向右数 stepNum 个点。
			for s := int64(0); s < stepNum; s++ {
				lastTs := ts + query.SubIntervalWith30min*s
				if _, exist := pointMap[lastTs]; !exist {
					continue
				}

				compareRatePoints(&tempPoints, ts, pointMap[lastTs])
			}

			rateFunction(&tempPoints, &finalPoints, ts, query, e, call)
			delete(tempPoints, ts)
		}

		seriesMap[k] = static.Series{
			Metric: parseLabelsStr(k, labelsMap),
			Points: finalPoints,
		}

		chs <- seriesMap
	}
}

// 合并相同时间戳的 point
func mergePointWithSameKeyTS(pointMap *map[int64]*static.RatePoint, tsArr [][]gjson.Result) {
	for _, tsArri := range tsArr {
		for _, pointij := range tsArri {
			currentT := pointij.Get("key").Int()

			currentPoint := &static.RatePoint{
				FirstTimestamp:    pointij.Get("value.firstTimestamp").Int(),
				FirstValue:        pointij.Get("value.firstValue").Float(),
				LastTimestamp:     pointij.Get("value.lastTimestamp").Int(),
				LastValue:         pointij.Get("value.lastValue").Float(),
				CounterCorrection: pointij.Get("value.counterCorrection").Float(),
				PointsCount:       pointij.Get("doc_count").Int(),
			}
			compareRatePoints(pointMap, currentT, currentPoint)
		}
	}
}

// 对两个point进行比较，最终输出最大的样本点和最小的样本点
func compareRatePoints(pointMap *map[int64]*static.RatePoint, ts int64, currentPoint *static.RatePoint) {
	if existPoint, ok := (*pointMap)[ts]; ok {
		// 1. 先计算 counterCorrection
		if existPoint.FirstTimestamp > currentPoint.LastTimestamp {
			// 时间上没有交集，exist 在后，current 在前。 counterCorrection 相加，再比较两者之间的修正量
			existPoint.CounterCorrection += currentPoint.CounterCorrection
			if existPoint.FirstValue < currentPoint.LastValue {
				existPoint.CounterCorrection += currentPoint.LastValue
			}
		} else if existPoint.LastTimestamp < currentPoint.FirstTimestamp {
			// 时间上没有交集，exist 在前，current 在后。 counterCorrection 相加，再比较两者之间的修正量
			existPoint.CounterCorrection += currentPoint.CounterCorrection
			if existPoint.LastValue > currentPoint.FirstValue {
				existPoint.CounterCorrection += existPoint.LastValue
			}
		} else {
			// 时间上有交集，取 counterCorrection 最大的
			if existPoint.CounterCorrection < currentPoint.CounterCorrection {
				existPoint.CounterCorrection = currentPoint.CounterCorrection
			}
		}

		// 2. 再计算时间区间，即 firstT 和 lastT
		if existPoint.FirstTimestamp > currentPoint.FirstTimestamp || (existPoint.FirstTimestamp == currentPoint.FirstTimestamp && existPoint.FirstValue < currentPoint.FirstValue) {
			// 如果 firstT 不一样，取其中时间最小的那个样本；如果时间戳一样，取值最大的那个样本。
			existPoint.FirstTimestamp = currentPoint.FirstTimestamp
			existPoint.FirstValue = currentPoint.FirstValue
		}

		if existPoint.LastTimestamp < currentPoint.LastTimestamp || (existPoint.LastTimestamp == currentPoint.LastTimestamp && existPoint.LastValue < currentPoint.LastValue) {
			// 如果 lastT 不一样，取其中时间最大的那个样本；如果时间戳一样，取值最大的那个样本。
			existPoint.LastTimestamp = currentPoint.LastTimestamp
			existPoint.LastValue = currentPoint.LastValue
		}

		existPoint.PointsCount += currentPoint.PointsCount

	} else {
		(*pointMap)[ts] = &static.RatePoint{
			FirstTimestamp:    currentPoint.FirstTimestamp,
			FirstValue:        currentPoint.FirstValue,
			LastTimestamp:     currentPoint.LastTimestamp,
			LastValue:         currentPoint.LastValue,
			CounterCorrection: currentPoint.CounterCorrection,
			PointsCount:       currentPoint.PointsCount,
		}
	}
}

func rateFunction(tempPoints *map[int64]*static.RatePoint, finalPoints *[]static.Point, ts int64, query *interfaces.Query, e *parser.MatrixSelector, call FunctionCall) *[]static.Point {
	inArgs := make([]parser.Value, 1)

	// 如果当前时间点对应的 range 没有数据点，跳过计算rate值
	var win *static.RatePoint
	if ratePoint, ok := (*tempPoints)[ts]; ok {
		win = ratePoint
	} else {
		return finalPoints
	}

	inArgs[0] = static.RatePoint{
		FirstTimestamp:    win.FirstTimestamp,
		FirstValue:        win.FirstValue,
		LastTimestamp:     win.LastTimestamp,
		LastValue:         win.LastValue,
		CounterCorrection: win.CounterCorrection,
		PointsCount:       win.PointsCount,
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
