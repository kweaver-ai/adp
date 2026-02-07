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

// changes 算子：构造dsl，指定分片查询，合并结果
func (ln *LeafNodes) ChangesAggs(ctx context.Context, expr *parser.MatrixSelector, query *interfaces.Query) (parser.Value, int, error) {
	// changes 有时间区间, 对于instant query, start倒推evalRange
	ctx, span := ar_trace.Tracer.Start(ctx, "Eval 叶子节点 Changes")
	defer span.End()

	vs, newQuery, status, err := processParam(expr, query)
	if err != nil {
		return nil, status, err
	}

	span.SetAttributes(attribute.Key("start").Int64(query.Start),
		attribute.Key("subInterval30min").Int64(newQuery.SubIntervalWith30min),
		attribute.Key("subInterval2h").Int64(newQuery.SubIntervalWith2h),
	)

	// 通用处理： 获取日志分组的索引信息 -> 构造 dsl -> 获取索引库下的所有索引以及对应的分片数 -> 执行 dsl
	result, status, err := ln.commonProcess(ctx, vs, newQuery, interfaces.CHANGES_AGG)
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
	mergeCtx, mergeSpan := ar_trace.Tracer.Start(ctx, "Changes merge")
	defer mergeSpan.End()
	mat, err := changesMerge(result.(MapResult), expr, newQuery)
	if err != nil {
		// 记录异常的日志
		o11y.Error(mergeCtx, fmt.Sprintf("Changes merge Error: %v", err))
		span.SetStatus(codes.Error, "Changes merge Error")

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
func changesMerge(mapResult MapResult, expr *parser.MatrixSelector, query *interfaces.Query) (static.Matrix, error) {
	keys := make([]string, 0, len(mapResult.LabelsMap))
	for key := range mapResult.LabelsMap {
		keys = append(keys, key)
	}
	// 对 keys排序，从小到大
	sort.Strings(keys)

	if query.IsInstantQuery {
		return changesMerge4InstantQuery(keys, mapResult, query)
	}
	return changesMerge4RangeQuery(keys, mapResult, query, expr)
}

// instant query 串行合并
func changesMerge4InstantQuery(keys []string, mapResult MapResult, query *interfaces.Query) (static.Matrix, error) {
	mat := make(static.Matrix, 0, len(keys))
	for _, k := range keys {
		tsArr := mapResult.TsValueMap[k]
		start, end, totalNum := getPointsStatsInfo(tsArr)
		pointMap := make(map[int64]*static.ChangesPoint, totalNum)

		// 遍历tsArr，合并每个数组的值，即把每个点的数据都放在目标map中，如果相同，则合并
		changesMergePointWithSameKey(pointMap, tsArr)

		if len(pointMap) == 0 {
			return mat, nil
		}

		var changes int64
		for ts := start; ts <= end; ts += query.SubIntervalWith30min {
			if point, ok := pointMap[ts]; ok {
				if ts > start {
					for lastTs := ts - query.SubIntervalWith30min; lastTs >= start; lastTs -= query.SubIntervalWith30min {
						// 所以当前一个不存在时，需继续循环再向前取一个窗口，知道取到为止。
						if _, exist := pointMap[lastTs]; !exist {
							continue
						}
						// 取到了比较，然后退出循环
						if ts > start && point.FirstValue != pointMap[lastTs].LastValue &&
							!(math.IsNaN(point.FirstValue) && math.IsNaN(pointMap[lastTs].LastValue)) {
							changes++
						}
						break
					}
				}
				changes += point.Changes
			}
		}

		mat = append(mat, static.Series{
			Metric: parseLabelsStr(k, mapResult.LabelsMap),
			Points: []static.Point{
				{T: query.End, V: float64(changes)},
			},
		})
	}

	return mat, nil

}

// 区间查询, 每个__labels_str下的聚合结果各自用一个goroutine去合并
func changesMerge4RangeQuery(keys []string, mapResult MapResult, query *interfaces.Query, expr *parser.MatrixSelector) (static.Matrix, error) {
	mat := make(static.Matrix, 0)
	chs := make(chan map[string]static.Series, len(mapResult.TsValueMap))
	defer close(chs)
	var wg sync.WaitGroup
	wg.Add(len(mapResult.TsValueMap))
	for _, k := range keys {
		tsArr := mapResult.TsValueMap[k]
		err := util.MegerPool.Submit(changesMergeTaskFuncWrapper(chs, k, tsArr, mapResult.LabelsMap, query, &wg, expr))
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
func changesMergeTaskFuncWrapper(chs chan<- map[string]static.Series, k string, tsArr [][]gjson.Result,
	labelsMap map[string][]*labels.Label, query *interfaces.Query, wg *sync.WaitGroup, expr *parser.MatrixSelector) taskFunc {

	return func() {
		defer wg.Done()

		// 先遍历tsArr去重
		pointMap := make(map[int64]*static.ChangesPoint, 0)
		changesMergePointWithSameKey(pointMap, tsArr)

		finalPoints := make([]static.Point, 0)
		selRangeTime := expr.Range.Milliseconds()
		tempPoints := make(map[int64]*static.ChangesPoint, 1)
		seriesMap := make(map[string]static.Series)
		stepNum := selRangeTime / query.SubIntervalWith30min

		// changes 的查询涉及到 substep 的拆分，所以在合并时，开始时间和结束时间按 fixedStart 和 fixedEnd 为准
		// for ts := query.FixedStart; ts <= query.FixedEnd; ts += query.Interval {
		for ts := query.FixedStart; ts <= query.FixedEnd; ts = static.GetNextPointTime(*query, ts) {
			// 根据step映射的点的个数，对每个时间点做range处理，向右数 stepNum 个点。
			for s := int64(0); s < stepNum; s++ {
				lastTs := ts + query.SubIntervalWith30min*s
				if _, exist := pointMap[lastTs]; !exist {
					continue
				}

				compareChangesPoints(tempPoints, ts, pointMap[lastTs])
			}

			// 如果当前时间点对应的 range 没有数据点，继续下一个。
			if _, ok := tempPoints[ts]; !ok {
				continue
			}

			finalPoints = append(finalPoints, static.Point{T: ts, V: float64(tempPoints[ts].Changes)})
			delete(tempPoints, ts)
		}

		seriesMap[k] = static.Series{
			Metric: parseLabelsStr(k, labelsMap),
			Points: finalPoints,
		}

		chs <- seriesMap
	}
}

// 按时间窗去重k个数组（即k个分片）中的样本点
func changesMergePointWithSameKey(pointMap map[int64]*static.ChangesPoint, tsArr [][]gjson.Result) {
	// tsArr 数组是各个分片上的时间分桶和值的列表
	for _, tsArri := range tsArr {
		for _, pointij := range tsArri {
			currentT := pointij.Get("key").Int()

			currentPoint := &static.ChangesPoint{
				FirstTimestamp: pointij.Get("value.firstTimestamp").Int(),
				FirstValue:     pointij.Get("value.firstValue").Float(),
				LastTimestamp:  pointij.Get("value.lastTimestamp").Int(),
				LastValue:      pointij.Get("value.lastValue").Float(),
				Changes:        pointij.Get("value.changes").Int(),
			}
			compareChangesPoints(pointMap, currentT, currentPoint)
		}
	}
}

// 对两个point进行比较，最终输出最大的样本点和最小的样本点
func compareChangesPoints(pointMap map[int64]*static.ChangesPoint, ts int64, currentPoint *static.ChangesPoint) {
	if existPoint, ok := pointMap[ts]; ok {
		// 1. 先计算 changes
		if existPoint.FirstTimestamp > currentPoint.LastTimestamp {
			// 时间上没有交集，exist 在后，current 在前。 changes 相加，再比较两者之间的变化次数
			existPoint.Changes += currentPoint.Changes
			if existPoint.FirstValue != currentPoint.LastValue && !(math.IsNaN(existPoint.FirstValue) && math.IsNaN(currentPoint.LastValue)) {
				existPoint.Changes++
			}
		} else if existPoint.LastTimestamp < currentPoint.FirstTimestamp {
			// 时间上没有交集，exist 在前，current 在后。 changes 相加，再比较两者之间的变化次数
			existPoint.Changes += currentPoint.Changes
			if existPoint.LastValue != currentPoint.FirstValue && !(math.IsNaN(existPoint.LastValue) && math.IsNaN(currentPoint.FirstValue)) {
				existPoint.Changes++
			}
		} else {
			// 时间上有交集，取 changes 最大的
			if existPoint.Changes < currentPoint.Changes {
				existPoint.Changes = currentPoint.Changes
			}
		}

		// 2. 再计算时间区间，即 firstT 和 lastT
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
	} else {
		pointMap[ts] = &static.ChangesPoint{
			FirstTimestamp: currentPoint.FirstTimestamp,
			FirstValue:     currentPoint.FirstValue,
			LastTimestamp:  currentPoint.LastTimestamp,
			LastValue:      currentPoint.LastValue,
			Changes:        currentPoint.Changes,
		}
	}
}
