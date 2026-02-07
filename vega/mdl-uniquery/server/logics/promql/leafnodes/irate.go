// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package leafnodes
// @Description: irate叶子节点处理逻辑
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

// IrateEval
//
//	@Description: irate处理主函数
//	@receiver leafNodes
//	@param expr
//	@param groupBy
//	@param aggregationType
//	@param query
//	@param selRange
//	@return parser.Value
//	@return int
//	@return error
func (leafNodes *LeafNodes) IrateEval(ctx context.Context, expr *parser.MatrixSelector, groupBy []string, aggregationType string,
	query *interfaces.Query) (parser.Value, int, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Eval 叶子节点 Irate")
	defer span.End()

	start := query.Start
	selRange := expr.Range.Milliseconds()

	if query.IsInstantQuery {
		// if range auto is true, range 等于请求参数中的delta，
		if expr.Auto && expr.Range == 0 {
			// selRange = convert.GetLookBackDelta(query.LookBackDelta, leafNodes.appSetting.PromqlSetting.LookbackDelta)
			selRange = query.End - query.Start
		}
		// 结束时间 - range
		start = query.End - selRange
		query.Interval = selRange
	}
	match, ok := expr.VectorSelector.(*parser.VectorSelector)
	if !ok {
		return nil, http.StatusUnprocessableEntity, uerrors.PromQLError{
			Typ: uerrors.ErrorExec,
			Err: fmt.Errorf("promql.Exec: invalid expression type %q", expr.VectorSelector.Type()),
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
	result, status, err := leafNodes.commonProcess(ctx, match, &newQuery, aggregationType)
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
	mergeCtx, mergeSpan := ar_trace.Tracer.Start(ctx, "Irate merge")
	defer mergeSpan.End()
	mat, err := irateMerge(result.(MapResult), &newQuery, expr.Range)
	if err != nil {
		// 记录异常的日志
		o11y.Error(mergeCtx, fmt.Sprintf("Irate merge Error: %v", err))
		span.SetStatus(codes.Error, "Irate merge Error")

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

// gcd
// @Description: 获取最大公约数
// @param a
// @param b
// @return int
func getCommonDivisor(a, b int64) int64 {
	for b != 0 {
		tmp := b
		b = a % b
		a = tmp
	}
	return a
}

// IrateMerge
//
//	@Description: 合并数据方法
//	@receiver leafNodes
//	@param mapResult
//	@param query
//	@param selRange
//	@return static.IrateMatrix
//	@return error
func irateMerge(mapResult MapResult, query *interfaces.Query, selRange time.Duration) (static.Matrix, error) {
	// 对 map 的 key 排序，使得同一个请求输出的结果不会乱序
	keys := []string{}
	for key := range mapResult.LabelsMap {
		keys = append(keys, key)
	}
	// 对 keys 排序，从小到大
	sort.Strings(keys)

	if query.IsInstantQuery {
		return mergeInstant(mapResult, query, keys)
	}
	return mergeRange(mapResult, query, selRange, keys)
}

func mergeInstant(mapResult MapResult, query *interfaces.Query, keys []string) (static.Matrix, error) {
	mat := make(static.Matrix, 0)
	for _, k := range keys {
		tsArr := mapResult.TsValueMap[k]
		var start, end int64
		start, end = math.MaxInt64, math.MinInt64
		totalNum := 0
		for i := 0; i < len(tsArr); i++ {
			totalNum += len(tsArr[i])
			minTs := tsArr[i][0].Get("key").Int()
			if minTs < start {
				start = minTs
			}
			maxTs := tsArr[i][len(tsArr[i])-1].Get("key").Int()
			if maxTs > end {
				end = maxTs
			}
		}
		// 2. 遍历 tsArr，合并每个数组的值,即把每个点的数据都放在目标map中，如果相同，则合并
		pointMap := make(map[int64]static.IratePoint, totalNum)
		iratePoint(&pointMap, tsArr)
		var tempPoints = make(map[int64]static.IratePoint, 1)
		var finalPoints = make([]static.Point, 0)
		// 对于instant query， 采样聚合只需要最后一个样本点值，那么直接取最后一个point的值，并且把样本点的时间替换成query.end
		previousV := pointMap[end].PreviousV
		previousT := pointMap[end].PreviousT
		if pointMap[end].PreviousV == 0 && len(pointMap) > 1 {
			// todo: 日历步长时，此处是取上一个时间点的数据，用 end-query.Interval不对，需修正
			previousV = pointMap[end-query.Interval].LastV
			previousT = pointMap[end-query.Interval].LastT
		}
		tempPoints = map[int64]static.IratePoint{
			query.End: {
				PreviousT: previousT,
				PreviousV: previousV,
				LastT:     pointMap[end].LastT,
				LastV:     pointMap[end].LastV,
			},
		}
		irateFunction(&tempPoints, query.End, &finalPoints)
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

func mergeRange(mapResult MapResult, query *interfaces.Query, selRange time.Duration, keys []string) (static.Matrix, error) {
	chs := make(chan map[string]static.Series, len(mapResult.TsValueMap))
	defer close(chs)

	mat := make(static.Matrix, 0)
	var wg sync.WaitGroup
	wg.Add(len(mapResult.TsValueMap))
	for _, k := range keys {
		tsArr := mapResult.TsValueMap[k]
		// 用协程池提交执行的操作
		err := util.MegerPool.Submit(irateMergeTaskFuncWapper(chs, k, tsArr, mapResult.LabelsMap, query, &wg, selRange))
		if err != nil {
			return nil, err
		}
	}
	wg.Wait()

	seriesMap := make(map[string]static.Series)
	for j := 0; j < len(mapResult.TsValueMap); j++ {
		series := <-chs
		// mat = append(mat, series)
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

// irateMergeTaskFuncWapper
// @Description: irate函数时间直方图和irateSampling合并方法
// @param chs
// @param k
// @param tsArr
// @param labelsMap
// @param query
// @param wg
// @param selRange
// @return taskFunc
func irateMergeTaskFuncWapper(chs chan<- map[string]static.Series, k string, tsArr [][]gjson.Result,
	labelsMap map[string][]*labels.Label,
	query *interfaces.Query, wg *sync.WaitGroup, selRange time.Duration) taskFunc {
	return func() {
		defer wg.Done()
		// 1. 从k个数组中选取时间最小的作为开始时间, 选取时间最大的作为结束时间
		var start, end int64
		start = query.FixedStart
		end = query.FixedEnd
		totalNum := 0
		for i := 0; i < len(tsArr); i++ {
			totalNum += len(tsArr[i])
		}
		// 2. 遍历 tsArr，合并每个数组的值,即把每个点的数据都放在目标map中，如果相同，则合并
		pointMap := make(map[int64]static.IratePoint, totalNum)
		iratePoint(&pointMap, tsArr)
		// 3. 从start到end遍历，补充缺失的数据点
		var tempPoints = make(map[int64]static.IratePoint, 1)
		var finalPoints = make([]static.Point, 0)
		seriesMap := make(map[string]static.Series)
		selRangeTime := selRange.Milliseconds()
		step := selRangeTime / query.SubIntervalWith30min
		// for ts := start; ts <= end; ts += query.Interval {
		for ts := start; ts <= end; ts = static.GetNextPointTime(*query, ts) {
			//根据step，对每个时间点做range处理
			for s := step; s >= 0; s-- {
				lastTs := ts + query.SubIntervalWith30min*s
				if _, exist := pointMap[lastTs]; !exist {
					continue
				}
				previousT := pointMap[lastTs].PreviousT
				previousV := pointMap[lastTs].PreviousV
				lastT := pointMap[lastTs].LastT
				lastV := pointMap[lastTs].LastV
				comparePoints(&tempPoints, lastT, lastV, ts, previousT, previousV)
				if previousT != 0 {
					break
				}
			}
			irateFunction(&tempPoints, ts, &finalPoints)
			delete(tempPoints, ts)
		}

		seriesMap[k] = static.Series{
			Metric: parseLabelsStr(k, labelsMap),
			Points: finalPoints,
		}

		chs <- seriesMap
	}
}

// irateFunction
// @Description: irate计算
// @param tempPoints 计算的样本点
// @param ts 当前时间戳
// @param finalPoints 存储irate值
// @return *[]static.Point
func irateFunction(tempPoints *map[int64]static.IratePoint, ts int64, finalPoints *[]static.Point) *[]static.Point {
	value := (*tempPoints)[ts]
	if value.PreviousT == 0 {
		return finalPoints
	}
	var resultValue float64
	if value.LastV < value.PreviousV {
		// Counter reset.
		resultValue = value.LastV
	} else {
		resultValue = value.LastV - value.PreviousV
	}

	sampledInterval := value.LastT - value.PreviousT
	if sampledInterval == 0 {
		// Avoid dividing by 0.
		return finalPoints
	}
	// Convert to per-second.
	resultValue /= float64(sampledInterval) / 1000

	*finalPoints = append(*finalPoints, static.Point{T: ts, V: resultValue})
	return finalPoints
}

// iratePoint
// @Description: 合并分片的数据
// @param pointMap: 合并存储的样本map
// @param tsArr: 分片数据
func iratePoint(pointMap *map[int64]static.IratePoint, tsArr [][]gjson.Result) {
	for _, tsArri := range tsArr {
		for _, pointij := range tsArri {
			currentT := pointij.Get("key").Int()
			previousT := pointij.Get("value.previousTimestamp").Int()
			previousValue := pointij.Get("value.previousValue").Float()
			lastT := pointij.Get("value.lastTimestamp").Int()
			lastValue := pointij.Get("value.lastValue").Float()
			comparePoints(pointMap, lastT, lastValue, currentT, previousT, previousValue)
		}
	}
}

// comparePoints
// @Description: 对两个point进行比较，最终输出最大的样本点和第二大样本点
// @param pointMap: 样本map,也是最终输出
// @param lastT: lastSample的时间戳
// @param lastV: lastSample的值
// @param ts: step步长的时间窗
// @param previousT: previousSample的时间戳
// @param previousV: previousSample的值
func comparePoints(pointMap *map[int64]static.IratePoint, lastT int64, lastV float64, ts int64, previousT int64, previousV float64) {

	if point, ok := (*pointMap)[ts]; ok {
		if lastT > point.LastT || (lastT == point.LastT && lastV > point.LastV) {
			(*pointMap)[ts] = static.IratePoint{
				PreviousT: point.LastT,
				PreviousV: point.LastV,
				LastT:     lastT,
				LastV:     lastV,
			}
		} else if lastT > point.PreviousT || (lastT == point.PreviousT && lastV > point.PreviousV) {
			//走到这里，值已确定
			(*pointMap)[ts] = static.IratePoint{
				PreviousT: lastT,
				PreviousV: lastV,
				LastT:     point.LastT,
				LastV:     point.LastV,
			}
			return
		}
		if previousT > point.LastT || (previousT == point.LastT && previousV > point.LastV) {
			(*pointMap)[ts] = static.IratePoint{
				PreviousT: previousT,
				PreviousV: previousV,
				LastT:     lastT,
				LastV:     lastV,
			}
		}
	} else {
		(*pointMap)[ts] = static.IratePoint{
			PreviousT: previousT,
			PreviousV: previousV,
			LastT:     lastT,
			LastV:     lastV,
		}
	}
}
