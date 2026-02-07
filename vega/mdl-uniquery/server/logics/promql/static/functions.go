// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package static

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"uniquery/interfaces"
	"uniquery/logics/data_dict"
	"uniquery/logics/promql/labels"
	"uniquery/logics/promql/parser"
)

// "uniquery/logics/promql/static"

type FunctionCall interface {
	New(args parser.Expressions) FunctionCall
	Call(vals []parser.Value, args parser.Expressions, enh *EvalNodeHelper) Vector
}

// === time() float64 ===
type funcTime struct {
}

func (f funcTime) New(args parser.Expressions) FunctionCall {
	return &funcTime{}
}
func (f funcTime) Call(vals []parser.Value, args parser.Expressions, enh *EvalNodeHelper) Vector {
	return Vector{Sample{Point: Point{
		V: float64(enh.Ts) / 1000,
	}}}
}

func simpleFunc(vals []parser.Value, enh *EvalNodeHelper, f func(float64) float64) Vector {
	for _, el := range vals[0].(Vector) {
		enh.Out = append(enh.Out, Sample{
			Metric: el.Metric,
			Point:  Point{V: f(el.V)},
		})
	}
	return enh.Out
}

// === abs(Vector parser.ValueTypeVector) Vector ===
type funcAbs struct{}

func (f funcAbs) New(args parser.Expressions) FunctionCall {
	return &funcAbs{}
}
func (f funcAbs) Call(vals []parser.Value, args parser.Expressions, enh *EvalNodeHelper) Vector {
	return simpleFunc(vals, enh, math.Abs)
}

// === ceil(Vector parser.ValueTypeVector) Vector ===
type funcCeil struct{}

func (f funcCeil) New(args parser.Expressions) FunctionCall {
	return &funcCeil{}
}
func (f funcCeil) Call(vals []parser.Value, args parser.Expressions, enh *EvalNodeHelper) Vector {
	return simpleFunc(vals, enh, math.Ceil)
}

// === floor(Vector parser.ValueTypeVector) Vector ===
type funcFloor struct{}

func (f funcFloor) New(args parser.Expressions) FunctionCall {
	return &funcFloor{}
}
func (f funcFloor) Call(vals []parser.Value, args parser.Expressions, enh *EvalNodeHelper) Vector {
	return simpleFunc(vals, enh, math.Floor)
}

// === exp(Vector parser.ValueTypeVector) Vector ===
type funcExp struct{}

func (f funcExp) New(args parser.Expressions) FunctionCall {
	return &funcExp{}
}
func (f funcExp) Call(vals []parser.Value, args parser.Expressions, enh *EvalNodeHelper) Vector {
	return simpleFunc(vals, enh, math.Exp)
}

// === sqrt(Vector VectorNode) Vector ===
type funcSqrt struct{}

func (f funcSqrt) New(args parser.Expressions) FunctionCall {
	return &funcSqrt{}
}
func (f funcSqrt) Call(vals []parser.Value, args parser.Expressions, enh *EvalNodeHelper) Vector {
	return simpleFunc(vals, enh, math.Sqrt)
}

// === ln(Vector parser.ValueTypeVector) Vector ===
type funcLn struct{}

func (f funcLn) New(args parser.Expressions) FunctionCall {
	return &funcLn{}
}
func (f funcLn) Call(vals []parser.Value, args parser.Expressions, enh *EvalNodeHelper) Vector {
	return simpleFunc(vals, enh, math.Log)
}

// === log2(Vector parser.ValueTypeVector) Vector ===
type funcLog2 struct{}

func (f funcLog2) New(args parser.Expressions) FunctionCall {
	return &funcLog2{}
}
func (funcLog2) Call(vals []parser.Value, args parser.Expressions, enh *EvalNodeHelper) Vector {
	return simpleFunc(vals, enh, math.Log2)
}

// === log10(Vector parser.ValueTypeVector) Vector ===
type funcLog10 struct{}

func (f funcLog10) New(args parser.Expressions) FunctionCall {
	return &funcLog10{}
}
func (f funcLog10) Call(vals []parser.Value, args parser.Expressions, enh *EvalNodeHelper) Vector {
	return simpleFunc(vals, enh, math.Log10)
}

// funcIrate
// @Description: irate function回调函数
// @param vals
// @param args
// @param enh
// @return static.Vector
type funcIrate struct{}

func (f funcIrate) New(args parser.Expressions) FunctionCall {
	return &funcIrate{}
}
func (f funcIrate) Call(vals []parser.Value, args parser.Expressions, enh *EvalNodeHelper) Vector {
	return instantValue(vals, enh.Out, true)
}

// instantValue
// @Description: 计算irate
// @param vals
// @param out
// @param isRate
// @return static.Vector
func instantValue(vals []parser.Value, out Vector, isRate bool) Vector {
	samples := vals[0].(IrateSeries)
	// No sense in trying to compute a rate without at least two points. Drop
	// this Vector element.
	var keys = make([]int, 0)
	for k := range samples.Points {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)

	for _, k := range keys {
		ts := int64(k)
		value := samples.Points[ts]
		if value.PreviousT == 0 {
			out = append(out, Sample{
				Point: Point{T: ts, V: 0},
			})
			continue
		}
		var resultValue float64
		if isRate && value.LastV < value.PreviousV {
			// Counter reset.
			resultValue = value.LastV
		} else {
			resultValue = value.LastV - value.PreviousV
		}

		sampledInterval := value.LastT - value.PreviousT
		if sampledInterval == 0 {
			// Avoid dividing by 0.
			out = append(out, Sample{
				Point: Point{T: ts, V: 0},
			})
			continue
		}

		if isRate {
			// Convert to per-second.
			resultValue /= float64(sampledInterval) / 1000
		}
		out = append(out, Sample{
			Metric: samples.Metric,
			Point:  Point{T: ts, V: resultValue},
		})
	}
	return out
}

// === sort(node parser.ValueTypeVector) Vector ===
type funcSort struct{}

func (f funcSort) New(args parser.Expressions) FunctionCall {
	return &funcSort{}
}
func (f funcSort) Call(vals []parser.Value, args parser.Expressions, enh *EvalNodeHelper) Vector {
	// NaN should sort to the bottom, so take descending sort with NaN first and
	// reverse it.
	byValueSorter := vectorByReverseValueHeap(vals[0].(Vector))
	sort.Sort(sort.Reverse(byValueSorter))
	return Vector(byValueSorter)
}

// === sortDesc(node parser.ValueTypeVector) Vector ===
type funcSortDesc struct{}

func (f funcSortDesc) New(args parser.Expressions) FunctionCall {
	return &funcSortDesc{}
}
func (f funcSortDesc) Call(vals []parser.Value, args parser.Expressions, enh *EvalNodeHelper) Vector {
	// NaN should sort to the bottom, so take ascending sort with NaN first and
	// reverse it.
	byValueSorter := vectorByValueHeap(vals[0].(Vector))
	sort.Sort(sort.Reverse(byValueSorter))
	return Vector(byValueSorter)
}

// FunctionCalls 定义回调函数
// === label_replace(Vector parser.ValueTypeVector, dst_label, replacement, src_labelname, regex parser.ValueTypeString) Vector ===
// TODO后续把Call方法中需要的参数提到结构体中
type funcLabelReplace struct{}

func (f funcLabelReplace) New(args parser.Expressions) FunctionCall {
	return &funcLabelReplace{}
}
func (f funcLabelReplace) Call(vals []parser.Value, args parser.Expressions, enh *EvalNodeHelper) Vector {
	var (
		vector   = vals[0].(Vector)
		dst      = StringFromArg(args[1])
		repl     = StringFromArg(args[2])
		src      = StringFromArg(args[3])
		regexStr = StringFromArg(args[4])
	)

	if enh.Regex == nil {
		var err error
		enh.Regex, err = regexp.Compile("^(?:" + regexStr + ")$")
		if err != nil {
			panic(fmt.Errorf("invalid regular expression in label_replace(): %s", regexStr))
		}
		if !labels.LabelName(dst).IsValid() {
			panic(fmt.Errorf("invalid destination label name in label_replace(): %s", dst))
		}
		enh.Dmn = make(map[uint64]labels.Labels, len(enh.Out))
	}

	for _, el := range vector {
		h := el.Metric.Hash()
		var outMetric labels.Labels
		if l, ok := enh.Dmn[h]; ok {
			outMetric = l
		} else {
			srcVal := el.Metric.Get(src)
			indexes := enh.Regex.FindStringSubmatchIndex(srcVal)
			if indexes == nil {
				// If there is no match, no replacement should take place.
				outMetric = el.Metric
				enh.Dmn[h] = outMetric
			} else {
				res := enh.Regex.ExpandString([]byte{}, repl, srcVal, indexes)

				lb := labels.NewBuilder(el.Metric).Del(dst)
				if len(res) > 0 {
					lb.Set(dst, string(res))
				}
				outMetric = lb.Labels()
				enh.Dmn[h] = outMetric
			}
		}

		enh.Out = append(enh.Out, Sample{
			Metric: outMetric,
			Point:  Point{V: el.Point.V},
		})
	}
	return enh.Out
}

// === label_join(vector model.ValVector, dest_labelname, separator, src_labelname...) Vector ===
// TODO后续把Call方法中需要的参数提到结构体中
type funcLabelJoin struct{}

func (f funcLabelJoin) New(args parser.Expressions) FunctionCall {
	return &funcLabelJoin{}
}
func (f funcLabelJoin) Call(vals []parser.Value, args parser.Expressions, enh *EvalNodeHelper) Vector {
	var (
		vector    = vals[0].(Vector)
		dst       = StringFromArg(args[1])
		sep       = StringFromArg(args[2])
		srcLabels = make([]string, len(args)-3)
	)

	if enh.Dmn == nil {
		enh.Dmn = make(map[uint64]labels.Labels, len(enh.Out))
	}

	for i := 3; i < len(args); i++ {
		src := StringFromArg(args[i])
		if !labels.LabelName(src).IsValid() {
			panic(fmt.Errorf("invalid source label name in label_join(): %s", src))
		}
		srcLabels[i-3] = src
	}

	if !labels.LabelName(dst).IsValid() {
		panic(fmt.Errorf("invalid destination label name in label_join(): %s", dst))
	}

	srcVals := make([]string, len(srcLabels))
	for _, el := range vector {
		h := el.Metric.Hash()
		var outMetric labels.Labels
		if l, ok := enh.Dmn[h]; ok {
			outMetric = l
		} else {

			for i, src := range srcLabels {
				srcVals[i] = el.Metric.Get(src)
			}

			lb := labels.NewBuilder(el.Metric)

			strval := strings.Join(srcVals, sep)
			if strval == "" {
				lb.Del(dst)
			} else {
				lb.Set(dst, strval)
			}

			outMetric = lb.Labels()
			enh.Dmn[h] = outMetric
		}

		enh.Out = append(enh.Out, Sample{
			Metric: outMetric,
			Point:  Point{V: el.Point.V},
		})
	}
	return enh.Out
}

func StringFromArg(e parser.Expr) string {
	tmp := UnwrapStepInvariantExpr(e) // Unwrap StepInvariant
	UnwrapParenExpr(&tmp)             // Optionally unwrap ParenExpr
	return tmp.(*parser.StringLiteral).Val
}

// === rate(node parser.ValueTypeMatrix) Vector ===  vector 是个样本数组
type funcRate struct{}

func (f funcRate) New(args parser.Expressions) FunctionCall {
	return &funcRate{}
}
func (f funcRate) Call(vals []parser.Value, args parser.Expressions, enh *EvalNodeHelper) Vector {
	return extrapolatedRate(vals, args, enh, true, true)
}

// === increase(node parser.ValueTypeMatrix) Vector ===
type funcIncrease struct{}

func (f funcIncrease) New(args parser.Expressions) FunctionCall {
	return &funcIncrease{}
}
func (f funcIncrease) Call(vals []parser.Value, args parser.Expressions, enh *EvalNodeHelper) Vector {
	return extrapolatedRate(vals, args, enh, true, false)
}

// === delta(node parser.ValueTypeMatrix) Vector ===
// TODO后续把Call方法中需要的参数提到结构体中
type funcDelta struct{}

func (f funcDelta) New(args parser.Expressions) FunctionCall {
	return &funcDelta{}
}
func (f funcDelta) Call(vals []parser.Value, args parser.Expressions, enh *EvalNodeHelper) Vector {
	ms := args[0].(*parser.MatrixSelector)

	var (
		deltaPoint = vals[0].(DeltaPoint)
		firstPoint = Point{T: deltaPoint.FirstTimestamp, V: deltaPoint.FirstValue}
		lastPoint  = Point{T: deltaPoint.LastTimestamp, V: deltaPoint.LastValue}
		rangeStart = enh.Ts
		rangeEnd   = enh.Ts + ms.Range.Milliseconds()
	)

	// No sense in trying to compute a rate without at least two points. Drop
	// this Vector element.
	if deltaPoint.PointsCount < 2 || (firstPoint.T == lastPoint.T && firstPoint.V == lastPoint.V) {
		return enh.Out
	}

	resultValue := lastPoint.V - firstPoint.V

	durationToStart := float64(firstPoint.T-rangeStart) / 1000
	durationToEnd := float64(rangeEnd-lastPoint.T) / 1000

	sampledInterval := float64(lastPoint.T-firstPoint.T) / 1000
	averageDurationBetweenSamples := sampledInterval / float64(deltaPoint.PointsCount-1)

	// If the first/last samples are close to the boundaries of the range,
	// extrapolate the result. This is as we expect that another sample
	// will exist given the spacing between samples we've seen thus far,
	// with an allowance for noise.
	extrapolationThreshold := averageDurationBetweenSamples * 1.1
	extrapolateToInterval := sampledInterval

	if durationToStart < extrapolationThreshold {
		extrapolateToInterval += durationToStart
	} else {
		extrapolateToInterval += averageDurationBetweenSamples / 2
	}

	if durationToEnd < extrapolationThreshold {
		extrapolateToInterval += durationToEnd
	} else {
		extrapolateToInterval += averageDurationBetweenSamples / 2
	}
	resultValue = resultValue * (extrapolateToInterval / sampledInterval)

	return append(enh.Out, Sample{
		Point: Point{V: resultValue},
	})
}

// extrapolatedRate is a utility function for rate/increase/delta.
// It calculates the rate (allowing for counter resets if isCounter is true),
// extrapolates if the first/last sample is close to the boundary, and returns
// the result as either per-second (if isRate is true) or overall.
func extrapolatedRate(vals []parser.Value, args parser.Expressions, enh *EvalNodeHelper, isCounter, isRate bool) Vector {
	ms := args[0].(*parser.MatrixSelector)

	var (
		ratePoint  = vals[0].(RatePoint)
		firstPoint = Point{T: ratePoint.FirstTimestamp, V: ratePoint.FirstValue}
		lastPoint  = Point{T: ratePoint.LastTimestamp, V: ratePoint.LastValue}
		rangeStart = enh.Ts
		rangeEnd   = enh.Ts + ms.Range.Milliseconds()
	)

	// No sense in trying to compute a rate without at least two points. Drop
	// this Vector element.
	if ratePoint.PointsCount < 2 || (firstPoint.T == lastPoint.T && firstPoint.V == lastPoint.V) {
		return enh.Out
	}

	resultValue := lastPoint.V - firstPoint.V

	if isCounter {
		resultValue += ratePoint.CounterCorrection
	}

	durationToStart := float64(firstPoint.T-rangeStart) / 1000
	durationToEnd := float64(rangeEnd-lastPoint.T) / 1000

	sampledInterval := float64(lastPoint.T-firstPoint.T) / 1000
	averageDurationBetweenSamples := sampledInterval / float64(ratePoint.PointsCount-1)

	if isCounter && resultValue > 0 && firstPoint.V >= 0 {
		// Counters cannot be negative. If we have any slope at
		// all (i.e. resultValue went up), we can extrapolate
		// the zero point of the counter. If the duration to the
		// zero point is shorter than the durationToStart, we
		// take the zero point as the start of the series,
		// thereby avoiding extrapolation to negative counter
		// values.
		durationToZero := sampledInterval * (firstPoint.V / resultValue)
		if durationToZero < durationToStart {
			durationToStart = durationToZero
		}
	}

	// If the first/last samples are close to the boundaries of the range,
	// extrapolate the result. This is as we expect that another sample
	// will exist given the spacing between samples we've seen thus far,
	// with an allowance for noise.
	extrapolationThreshold := averageDurationBetweenSamples * 1.1
	extrapolateToInterval := sampledInterval

	if durationToStart < extrapolationThreshold {
		extrapolateToInterval += durationToStart
	} else {
		extrapolateToInterval += averageDurationBetweenSamples / 2
	}

	if durationToEnd < extrapolationThreshold {
		extrapolateToInterval += durationToEnd
	} else {
		extrapolateToInterval += averageDurationBetweenSamples / 2
	}
	resultValue = resultValue * (extrapolateToInterval / sampledInterval)
	if isRate {
		resultValue = resultValue / ms.Range.Seconds()
	}

	return append(enh.Out, Sample{
		Point: Point{V: resultValue},
	})
}

// === histogram_quantile(k parser.ValueTypeScalar, Vector parser.ValueTypeVector) Vector ===
// TODO后续把Call方法中需要的参数提到结构体中
type funcHistogramQuantile struct{}

func (f funcHistogramQuantile) New(args parser.Expressions) FunctionCall {
	return &funcHistogramQuantile{}
}
func (f funcHistogramQuantile) Call(vals []parser.Value, args parser.Expressions, enh *EvalNodeHelper) Vector {
	q := vals[0].(Vector)[0].V
	inVec := vals[1].(Vector)

	sigf := signatureFunc(false, enh.lblBuf, excludedLabels...)

	if enh.signatureToMetricWithBuckets == nil {
		enh.signatureToMetricWithBuckets = map[string]*metricWithBuckets{}
	} else {
		for _, v := range enh.signatureToMetricWithBuckets {
			v.buckets = v.buckets[:0]
		}
	}
	for _, el := range inVec {
		upperBound, err := strconv.ParseFloat(
			el.Metric.Get(labels.BucketLabel), 64,
		)
		if err != nil {
			// Oops, no bucket label or malformed label value. Skip.
			continue
		}
		l := sigf(el.Metric)

		mb, ok := enh.signatureToMetricWithBuckets[l]
		if !ok {
			el.Metric = labels.NewBuilder(el.Metric).
				Del(labels.BucketLabel, labels.MetricName).
				Labels()

			mb = &metricWithBuckets{el.Metric, nil}
			enh.signatureToMetricWithBuckets[l] = mb
		}
		mb.buckets = append(mb.buckets, bucket{upperBound, el.V})
	}

	for _, mb := range enh.signatureToMetricWithBuckets {
		if len(mb.buckets) > 0 {
			enh.Out = append(enh.Out, Sample{
				Metric: mb.metric,
				Point:  Point{V: bucketQuantile(q, mb.buckets)},
			})
		}
	}

	return enh.Out
}

// === changes() float64 ===
type funcChange struct{}

func (f funcChange) New(args parser.Expressions) FunctionCall {
	return &funcChange{}
}
func (f funcChange) Call(vals []parser.Value, args parser.Expressions, enh *EvalNodeHelper) Vector {
	// 返回空，changes不回调此函数，计算 changes 在遍历 range 内的数据点有哪些时就已经完成计算，无需再回调。
	return Vector{}
}

// === agg_over_time() float64 ===
type funcAggOverTime struct{}

func (f funcAggOverTime) New(args parser.Expressions) FunctionCall {
	return &funcAggOverTime{}
}
func (f funcAggOverTime) Call(vals []parser.Value, args parser.Expressions, enh *EvalNodeHelper) Vector {
	// 返回空，agg_over_time 不回调此函数，计算 agg_over_time 在遍历 range 内的数据点有哪些时就已经完成计算，无需再回调。
	return Vector{}
}

// TODO后续把Call方法中需要的参数提到结构体中
type funcPercentRank struct{}

func (f funcPercentRank) New(args parser.Expressions) FunctionCall {
	return &funcPercentRank{}
}
func (f funcPercentRank) Call(vals []parser.Value, args parser.Expressions, enh *EvalNodeHelper) Vector {
	//取出序列
	vector := vals[0].(Vector)
	precision := int(args[1].(*parser.NumberLiteral).Val)

	if precision < 1 {
		panic(fmt.Errorf("invalid  precision in percent_rank(): %d", precision))
	}
	if precision > 16 {
		precision = 16
	}
	//小于当前指标值的个数
	counts := make(map[float64]int, 0)
	//指标值对应的排位值
	percentRank := make(map[float64]float64, 0)
	//指标值副本
	vSlice := make([]float64, 0)

	for _, v := range vector {
		vSlice = append(vSlice, v.V)
	}
	//升序排列
	sort.Slice(vSlice, func(i, j int) bool {
		return vSlice[i] < vSlice[j]
	})

	n := len(vector)
	if n == 1 {
		//对于NaN 、±INF 不排位
		if math.IsNaN(vector[0].V) || math.IsInf(vector[0].V, 1) || math.IsInf(vector[0].V, -1) {
			enh.Out = append(enh.Out, Sample{
				Metric: vector[0].Metric,
				Point:  Point{T: vector[0].T, V: vector[0].V},
			})
		} else {
			enh.Out = append(enh.Out, Sample{
				Metric: vector[0].Metric,
				Point:  Point{T: vector[0].T, V: truncateFloat(100, precision)},
			})
		}
		return enh.Out
	}

	for i, v := range vSlice {
		//计算小于当前指标值的个数
		if _, ok := counts[v]; !ok {
			counts[v] = i
		}
		//计算当前指标值的百分比排位
		percentRank[v] = truncateFloat(float64(counts[v])/float64(n-1)*100, precision-2)
	}

	for _, v := range vector {
		//对于NaN 、±INF 不排位
		if math.IsNaN(v.V) || math.IsInf(v.V, 1) || math.IsInf(v.V, -1) {
			enh.Out = append(enh.Out, Sample{
				Metric: v.Metric,
				Point:  Point{T: v.T, V: v.V},
			})
		} else {
			enh.Out = append(enh.Out, Sample{
				Metric: v.Metric,
				Point:  Point{T: v.T, V: percentRank[v.V]},
			})
		}
	}

	return enh.Out
}

// 截断浮点数
func truncateFloat(num float64, decimalPlaces int) float64 {
	multiplier := math.Pow(10, float64(decimalPlaces))
	truncated := math.Trunc(num*multiplier) / multiplier
	return truncated
}

// TODO后续把Call方法中需要的参数提到结构体中
type funcRank struct{}

func (f funcRank) New(args parser.Expressions) FunctionCall {
	return &funcRank{}
}
func (f funcRank) Call(vals []parser.Value, args parser.Expressions, enh *EvalNodeHelper) Vector {
	// 取出序列
	vector := vals[0].(Vector)
	order := int(args[1].(*parser.NumberLiteral).Val)

	//指标值副本
	vSlice := make([]float64, 0)

	for _, v := range vector {
		vSlice = append(vSlice, v.V)
	}
	if order == 0 {
		// 降序排列
		sort.Slice(vSlice, func(i, j int) bool {
			return vSlice[i] > vSlice[j]
		})
	} else {
		// 升序排列
		sort.Slice(vSlice, func(i, j int) bool {
			return vSlice[i] < vSlice[j]
		})
	}

	//指标值对应对排名
	Rank := make(map[float64]int, 0)

	rank := 1

	// 遍历排序后的切片，并计算指标值排名
	for i, v := range vSlice {
		if i > 0 && vSlice[i-1] != v {
			rank = i + 1
		}
		Rank[v] = rank
	}

	for _, v := range vector {
		if math.IsNaN(v.V) || math.IsInf(v.V, 1) || math.IsInf(v.V, -1) {
			enh.Out = append(enh.Out, Sample{
				Metric: v.Metric,
				Point:  Point{T: v.T, V: v.V},
			})
		} else {
			enh.Out = append(enh.Out, Sample{
				Metric: v.Metric,
				Point:  Point{T: v.T, V: float64(Rank[v.V])},
			})
		}
	}
	return enh.Out
}

type funcDictLabels struct {
	dict         interfaces.DataDict
	matchParams  map[string]string
	expandParams map[string]string
	expandLabels []string
}

func (f *funcDictLabels) New(args parser.Expressions) FunctionCall {

	// 读取缓存中的数据
	dictName := string(args[1].(*parser.StringLiteral).Val)
	dict, ok := data_dict.GetDictByName(dictName)
	if !ok {
		panic(fmt.Errorf("failed find dict: %s ", dictName))
	}

	keyMap := make(map[string]bool, 0)
	valueMap := make(map[string]bool, 0)
	for _, v := range dict.Dimension.Keys {
		keyMap[v.Name] = true
	}
	for _, v := range dict.Dimension.Values {
		valueMap[v.Name] = true
	}

	matchParams := make(map[string]string, 0)
	expandParams := make(map[string]string, 0)
	expandLabels := make([]string, 0)
	// 参数解析，两个参数为一组，分辨是连接条件还是扩充字段
	for i := 2; i < len(args); i = i + 2 {
		dict_label := string(args[i].(*parser.StringLiteral).Val)
		vector_label := string(args[i+1].(*parser.StringLiteral).Val)
		if keyMap[dict_label] {
			matchParams[dict_label] = vector_label
		} else if valueMap[dict_label] {
			// 如果是扩充字段，判断参数是否符合label标准
			if !labels.LabelName(vector_label).IsValid() {
				panic(fmt.Errorf("vector_label: %s is invalid", vector_label))
			}
			expandParams[dict_label] = vector_label
			expandLabels = append(expandLabels, vector_label)
		} else {
			panic(errors.New("invalid parameter in dict_labels()"))
		}
	}

	// 需要全部key才能查询到相应的数据
	if len(matchParams) != len(keyMap) {
		panic(errors.New("dict_labels() requires all the keys of the dictionary"))
	}

	// 给要扩充的label排序，以保证Metric里的label顺序是一致的
	sort.Strings(expandLabels)

	return &funcDictLabels{
		dict:         dict,
		expandParams: expandParams,
		matchParams:  matchParams,
		expandLabels: expandLabels,
	}
}

func (f *funcDictLabels) Call(vals []parser.Value, args parser.Expressions, enh *EvalNodeHelper) Vector {
	// 如果向量是否为空，返回空向量
	vector := vals[0].(Vector)
	if len(vector) == 0 {
		return enh.Out
	}

	// 如果字典数据为空，返回原序列
	if len(f.dict.DictRecords) == 0 {
		enh.Out = append(enh.Out, vector...)
		return enh.Out
	}

	// 遍历序列
	for _, v := range vals[0].(Vector) {
		// 将一条时间序列的label转化成map
		labelMap := make(map[string]string, 0)
		for _, label := range v.Metric {
			labelMap[label.Name] = label.Value
		}

		// 判断当前时间序列的连接条件是否成立，如果不成立，那这条时间序列不做扩充
		isConnection := true
		for _, value := range f.matchParams {
			if _, exit := labelMap[value]; !exit {
				isConnection = false
				break
			}
		}
		if !isConnection {
			enh.Out = append(enh.Out, Sample{
				Metric: v.Metric,
				Point:  v.Point,
			})
			continue
		}

		// 按照字典的key给值进行拼接
		keyArr := make([]string, 0)
		for _, key := range f.dict.Dimension.Keys {
			vectorLabel := f.matchParams[key.Name]
			keyArr = append(keyArr, labelMap[vectorLabel])
		}

		records, ok := data_dict.GetRecordsByKey(f.dict, keyArr)

		// 获取匹配的字典数据为空
		if !ok {
			enh.Out = append(enh.Out, Sample{
				Metric: v.Metric,
				Point:  v.Point,
			})
			continue
		}

		for _, record := range records {
			var metric labels.Labels
			metric = append(metric, v.Metric...)

			expendMap := make(map[string]string, 0)
			for key, value := range f.expandParams {
				expendMap[value] = record[key]
			}

			// 通过遍历排过序的label数组，保证扩充的标签加入的顺序一致
			for _, label := range f.expandLabels {
				if _, exits := labelMap[label]; exits {
					metric = append(metric, &labels.Label{
						Name:  "__m." + label,
						Value: expendMap[label],
					})
				} else {
					metric = append(metric, &labels.Label{
						Name:  label,
						Value: expendMap[label],
					})
				}
			}
			enh.Out = append(enh.Out, Sample{
				Metric: metric.Sort(),
				Point:  v.Point,
			})
		}
	}

	return enh.Out
}

type funcDictValues struct {
	dict         interfaces.DataDict
	measureField string
	sortLabels   []string
	joinMap      map[string]string
}

func (f *funcDictValues) New(args parser.Expressions) FunctionCall {
	// 读取缓存中的数据
	dictName := string(args[0].(*parser.StringLiteral).Val)
	dict, ok := data_dict.GetDictByName(dictName)
	if !ok {
		panic(fmt.Errorf("failed find dict: %s ", dictName))
	}
	if !dict.UniqueKey {
		panic(errors.New("dictionary must be uniqueKey"))
	}

	// 给字典的所有label进行排序，以保证插入到Metric里的label顺序是一致的
	sortLabels := make([]string, 0)

	keyMap := make(map[string]bool, 0)
	valueMap := make(map[string]bool, 0)
	for _, v := range dict.Dimension.Keys {
		keyMap[v.Name] = true
		sortLabels = append(sortLabels, v.Name)
	}
	for _, v := range dict.Dimension.Values {
		valueMap[v.Name] = true
		sortLabels = append(sortLabels, v.Name)
	}

	sort.Strings(sortLabels)

	// joinMap的键是字典的label，值是转化为向量的label
	joinMap := make(map[string]string, 0)
	joinKeyMap := make(map[string]string, 0)

	// 参数解析，两个参数为一组，分辨是连接条件还是扩充字段
	measureField := string(args[1].(*parser.StringLiteral).Val)
	if !valueMap[measureField] {
		panic(fmt.Errorf("invalid parameter measure_field: %s in dict_values()", measureField))
	}
	for i := 2; i < len(args); i = i + 2 {
		dict_label := string(args[i].(*parser.StringLiteral).Val)
		vector_label := string(args[i+1].(*parser.StringLiteral).Val)

		// 校验vector_label
		if !labels.LabelName(vector_label).IsValid() {
			panic(fmt.Errorf("vector_label: %s is invalid", vector_label))
		}

		if keyMap[dict_label] {
			joinKeyMap[dict_label] = vector_label
		} else if valueMap[dict_label] {
			joinMap[dict_label] = vector_label
		} else {
			panic(fmt.Errorf("invalid parameter in dict_values(),dict key:%s dose not exists", dict_label))
		}
	}

	for key := range keyMap {
		// 如果key在joinMap里不存在，不含特殊字符的话也需要扩充
		if _, exits := joinKeyMap[key]; !exits {
			if labels.LabelName(key).IsValid() {
				joinMap[key] = key
			} else {
				panic(errors.New("the dict has chinese key need to trealletr"))
			}
		}
	}

	for k, v := range joinKeyMap {
		joinMap[k] = v
	}

	return &funcDictValues{
		dict:         dict,
		sortLabels:   sortLabels,
		joinMap:      joinMap,
		measureField: measureField,
	}
}

func (f funcDictValues) Call(vals []parser.Value, args parser.Expressions, enh *EvalNodeHelper) Vector {

	// 如果字典数据为空，返回空
	if len(f.dict.DictRecords) == 0 {
		return enh.Out
	}
	// 字典转化指标
	for _, lineData := range f.dict.DictRecords {

		if len(lineData) == 0 {
			panic(errors.New("get dict data failed"))
		}

		if lineData[0][f.measureField] == "" {
			panic(errors.New("cannot convert an empty string into a metric value"))
		}

		metricValue, err := strconv.ParseFloat(lineData[0][f.measureField], 64)
		if err != nil {
			panic(fmt.Errorf("conversion metric failed: %v", err))
		}
		var sample Sample
		sample.Point.V = metricValue

		for _, label := range f.sortLabels {
			if _, exist := f.joinMap[label]; exist {
				sample.Metric = append(sample.Metric, &labels.Label{
					Name:  f.joinMap[label],
					Value: lineData[0][label],
				})
			}
		}
		sample.Metric.Sort()
		sample.Point.V = metricValue
		enh.Out = append(enh.Out, sample)
	}

	return enh.Out
}

// === cumulative_sum(Vector parser.ValueTypeVector) Vector ===
type funcCumulativeSum struct{}

func (f funcCumulativeSum) New(args parser.Expressions) FunctionCall {
	return &funcCumulativeSum{}
}

// 通用函数都是在每个step上对v做计算，即对指标在同一个时间点上的各个序列做计算。
// 而累加和是对每个序列在时间轴上的累计求和的过程。所以，累加和不能用通用function的模式来进行计算，特例处理。
func (f funcCumulativeSum) Call(vals []parser.Value, args parser.Expressions, enh *EvalNodeHelper) Vector {
	// do nothing
	return enh.Out
}

// === greatest(v1 instant-vector, v2 instant-vector, ...) Vector ===
type funcGreatest struct{}

func (f funcGreatest) New(args parser.Expressions) FunctionCall {
	return &funcGreatest{}
}

// 通用函数都是在每个step上对v做计算，即对指标在同一个时间点上的各个序列做计算。
// 而累加和是对每个序列在时间轴上的累计求和的过程。所以，累加和不能用通用function的模式来进行计算，特例处理。
func (f funcGreatest) Call(vals []parser.Value, args parser.Expressions, enh *EvalNodeHelper) Vector {
	return extrapolatedMaxOrMin(vals, enh, true)
}

// === least(v1 instant-vector ,v2 instant-vector, ...) Vector ===
type funcLeast struct{}

func (f funcLeast) New(args parser.Expressions) FunctionCall {
	return &funcLeast{}
}

// 通用函数都是在每个step上对v做计算，即对指标在同一个时间点上的各个序列做计算。
// 而累加和是对每个序列在时间轴上的累计求和的过程。所以，累加和不能用通用function的模式来进行计算，特例处理。
func (f funcLeast) Call(vals []parser.Value, args parser.Expressions, enh *EvalNodeHelper) Vector {
	return extrapolatedMaxOrMin(vals, enh, false)
}

// 比较vector，取样本值最大的作为最后的输出结果
func extrapolatedMaxOrMin(vals []parser.Value, enh *EvalNodeHelper, isMax bool) Vector {
	// 比较最大值。关联匹配，比较最大值
	if len(vals) == 0 {
		return enh.Out
	}
	if len(vals) == 1 {
		enh.Out = vals[0].(Vector)
		return enh.Out
	}
	// 根据关联操作类型和匹配的字段列表组装成的字符串来生成标识符,以第一个vector的labels为准
	sigf := enh.signatureFunc(false, []string{}...)

	// All samples from the rhs hashed by the matching label/values.
	if enh.rightSigs == nil {
		enh.rightSigs = make(map[string]Sample, len(enh.Out))
	} else {
		for k := range enh.rightSigs {
			delete(enh.rightSigs, k)
		}
	}
	firstSigs := enh.rightSigs

	// 第一个
	// Add all rhs samples to a map so we can easily find matches later.
	for _, rs := range vals[0].(Vector) {
		// 先把右边的 vector 的 labels 组成一个字符串，然后放在一个map中。
		sig := sigf(rs.Metric)
		// The rhs is guaranteed to be the 'one' side. Having multiple samples
		// with the same signature means that the matching is many-to-many.
		if _, found := firstSigs[sig]; found {
			panic(errors.New("found duplicate series for the match group, many-to-many only allowed for set operators"))
		}
		firstSigs[sig] = rs
	}

	// 第2个及之后的向量
	matchedSigArr := make([]map[string]Sample, 0)
	for j := 1; j < len(vals); j++ {
		matchedSigJ := make(map[string]Sample)
		for _, vjSample := range vals[j].(Vector) {
			sigj := sigf(vjSample.Metric)
			if _, found := matchedSigJ[sigj]; found {
				// 一对一的请求中出现vector数据是多对一，抛异常
				panic(errors.New("multiple matches for labels: many-to-one matching must be explicit (group_left/group_right)"))
			}
			matchedSigJ[sigj] = vjSample
		}
		matchedSigArr = append(matchedSigArr, matchedSigJ)
	}

	// Tracks the match-signature. For one-to-one operations the value is nil. For many-to-one
	// the value is a set of signatures to detect duplicated result elements.
	if enh.matchedSigs == nil {
		enh.matchedSigs = make(map[string]map[uint64]struct{}, len(firstSigs))
	} else {
		for k := range enh.matchedSigs {
			delete(enh.matchedSigs, k)
		}
	}
	matchedSigs := enh.matchedSigs

	// 以第一个vector为准，去其他vector中找匹配的序列sig
	for sig, firstSample := range firstSigs {
		_, exists := matchedSigs[sig]
		if exists {
			// 一对一的请求中出现vector数据是多对一，抛异常
			panic(errors.New("multiple matches for labels: many-to-one matching must be explicit (group_left/group_right)"))
		}
		matchedSigs[sig] = nil // Set existence to true.

		// Look for a match in the other Vector.
		found := false
		value := firstSample.Point.V
		for _, vj := range matchedSigArr {
			sigj, exist := vj[sig]
			if !exist {
				found = false
				break
			}
			found = true
			if isMax {
				value = max(sigj.V, value)
			} else {
				value = min(sigj.V, value)
			}
		}
		if !found {
			continue
		}

		// Account for potentially swapped sidedness.
		metric := resultMetric(firstSample.Metric, firstSample.Metric, &parser.VectorMatching{Card: parser.CardOneToOne, MatchingLabels: []string{}}, enh)

		enh.Out = append(enh.Out, Sample{
			Metric: metric,
			Point:  Point{V: value},
		})
	}
	return enh.Out
}

type funcClamp struct{}

func (f funcClamp) New(args parser.Expressions) FunctionCall {
	return &funcClamp{}
}
func (f funcClamp) Call(vals []parser.Value, args parser.Expressions, enh *EvalNodeHelper) Vector {
	minV := args[1].(*parser.NumberLiteral).Val
	maxV := args[2].(*parser.NumberLiteral).Val
	if math.IsNaN(minV) || math.IsNaN(maxV) {
		for _, el := range vals[0].(Vector) {
			enh.Out = append(enh.Out, Sample{
				Metric: el.Metric,
				Point:  Point{V: minV},
			})
		}
	} else if minV > maxV {
		enh.Out = []Sample{}
	} else {
		for _, el := range vals[0].(Vector) {
			v := min(max(el.V, minV), maxV)
			enh.Out = append(enh.Out, Sample{
				Metric: el.Metric,
				Point:  Point{V: v},
			})
		}
	}
	return enh.Out
}

type funcClampMax struct{}

func (f funcClampMax) New(args parser.Expressions) FunctionCall {
	return &funcClampMax{}
}
func (f funcClampMax) Call(vals []parser.Value, args parser.Expressions, enh *EvalNodeHelper) Vector {
	maxV := args[1].(*parser.NumberLiteral).Val
	if math.IsNaN(maxV) {
		for _, el := range vals[0].(Vector) {
			enh.Out = append(enh.Out, Sample{
				Metric: el.Metric,
				Point:  Point{V: maxV},
			})
		}
	} else {
		for _, el := range vals[0].(Vector) {
			v := min(el.V, maxV)
			enh.Out = append(enh.Out, Sample{
				Metric: el.Metric,
				Point:  Point{V: v},
			})
		}
	}
	return enh.Out
}

type funcClampMin struct{}

func (f funcClampMin) New(args parser.Expressions) FunctionCall {
	return &funcClampMin{}
}
func (f funcClampMin) Call(vals []parser.Value, args parser.Expressions, enh *EvalNodeHelper) Vector {
	minV := args[1].(*parser.NumberLiteral).Val
	if math.IsNaN(minV) {
		for _, el := range vals[0].(Vector) {
			enh.Out = append(enh.Out, Sample{
				Metric: el.Metric,
				Point:  Point{V: minV},
			})
		}
	} else {
		for _, el := range vals[0].(Vector) {
			v := max(el.V, minV)
			enh.Out = append(enh.Out, Sample{
				Metric: el.Metric,
				Point:  Point{V: v},
			})
		}
	}
	return enh.Out
}

// === continuous_k_minute_downtime(Vector parser.ValueTypeVector, Scalar parser.ValueTypeScalar) Vector ===
type funcKMinuteDowntime struct{}

func (f funcKMinuteDowntime) New(args parser.Expressions) FunctionCall {
	return &funcKMinuteDowntime{}
}
func (f funcKMinuteDowntime) Call(vals []parser.Value, args parser.Expressions, enh *EvalNodeHelper) Vector {
	// do nathing. 因为 k分钟连续不可用时长的计算，要参考的是序列的时间线上的数据，无法单点计算。所以不在通用function的模式中计算，特例处理
	return enh.Out
}

// === dsl(string parser.ValueTypeString) Vector ===
type funcMetricModel struct{}

func (f funcMetricModel) New(args parser.Expressions) FunctionCall {
	return &funcMetricModel{}
}
func (f funcMetricModel) Call(vals []parser.Value, args parser.Expressions, enh *EvalNodeHelper) Vector {
	// do nothing. 因为 dsl 是叶子节点，在叶子节点处理其逻辑
	return enh.Out
}

// FunctionCalls is a list of all functions supported by PromQL, including their types.
var FunctionCalls = map[string]FunctionCall{
	"abs":                          &funcAbs{},
	"ceil":                         &funcCeil{},
	"exp":                          &funcExp{},
	"floor":                        &funcFloor{},
	"time":                         &funcTime{},
	"irate":                        &funcIrate{},
	"sort":                         &funcSort{},
	"sort_desc":                    &funcSortDesc{},
	"label_replace":                &funcLabelReplace{},
	"label_join":                   &funcLabelJoin{},
	"ln":                           &funcLn{},
	"log10":                        &funcLog10{},
	"log2":                         &funcLog2{},
	"rate":                         &funcRate{},
	"sqrt":                         &funcSqrt{},
	"histogram_quantile":           &funcHistogramQuantile{},
	"increase":                     &funcIncrease{},
	"changes":                      &funcChange{},
	"avg_over_time":                &funcAggOverTime{},
	"sum_over_time":                &funcAggOverTime{},
	"count_over_time":              &funcAggOverTime{},
	"max_over_time":                &funcAggOverTime{},
	"min_over_time":                &funcAggOverTime{},
	"delta":                        &funcDelta{},
	"percent_rank":                 &funcPercentRank{},
	"rank":                         &funcRank{},
	"dict_labels":                  &funcDictLabels{},
	"dict_values":                  &funcDictValues{},
	"cumulative_sum":               &funcCumulativeSum{},
	"greatest":                     &funcGreatest{},
	"least":                        &funcLeast{},
	"clamp":                        &funcClamp{},
	"clamp_max":                    &funcClampMax{},
	"clamp_min":                    &funcClampMin{},
	"continuous_k_minute_downtime": &funcKMinuteDowntime{},
	"metric_model":                 &funcMetricModel{},
}
