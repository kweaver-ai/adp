// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package static

import (
	"bytes"
	"container/heap"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"uniquery/common/convert"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	"uniquery/logics/promql/labels"
	"uniquery/logics/promql/parser"
)

type groupedAggregation struct {
	labels      labels.Labels
	value       float64
	mean        float64
	groupCount  int
	heap        vectorByValueHeap
	reverseHeap vectorByReverseValueHeap
}

// btos returns 1 if b is true, 0 otherwise.
func btos(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

// scalarBinop evaluates a binary operation between two Scalars.
func ScalarBinop(op parser.ItemType, lhs, rhs float64) (float64, error) {
	switch op {
	case parser.ADD:
		return lhs + rhs, nil
	case parser.SUB:
		return lhs - rhs, nil
	case parser.MUL:
		return lhs * rhs, nil
	case parser.DIV:
		return lhs / rhs, nil
	case parser.POW:
		return math.Pow(lhs, rhs), nil
	case parser.MOD:
		return math.Mod(lhs, rhs), nil
	case parser.EQLC:
		return btos(lhs == rhs), nil
	case parser.NEQ:
		return btos(lhs != rhs), nil
	case parser.GTR:
		return btos(lhs > rhs), nil
	case parser.LSS:
		return btos(lhs < rhs), nil
	case parser.GTE:
		return btos(lhs >= rhs), nil
	case parser.LTE:
		return btos(lhs <= rhs), nil
	}
	return 0, uerrors.PromQLError{
		Typ: uerrors.ErrorExec,
		Err: fmt.Errorf("operator %q not allowed for Scalar operations", op),
	}
}

// vectorElemBinop evaluates a binary operation between two Vector elements.
func vectorElemBinop(op parser.ItemType, lhs, rhs float64) (float64, bool, error) {
	switch op {
	case parser.ADD:
		return lhs + rhs, true, nil
	case parser.SUB:
		return lhs - rhs, true, nil
	case parser.MUL:
		return lhs * rhs, true, nil
	case parser.DIV:
		return lhs / rhs, true, nil
	case parser.POW:
		return math.Pow(lhs, rhs), true, nil
	case parser.MOD:
		return math.Mod(lhs, rhs), true, nil
	case parser.EQLC:
		return lhs, lhs == rhs, nil
	case parser.NEQ:
		return lhs, lhs != rhs, nil
	case parser.GTR:
		return lhs, lhs > rhs, nil
	case parser.LSS:
		return lhs, lhs < rhs, nil
	case parser.GTE:
		return lhs, lhs >= rhs, nil
	case parser.LTE:
		return lhs, lhs <= rhs, nil
	}

	return 0, false, uerrors.PromQLError{
		Typ: uerrors.ErrorExec,
		Err: fmt.Errorf("operator %q not allowed for operations between Vectors", op),
	}
}

// VectorscalarBinop evaluates a binary operation between a Vector and a Scalar.
func VectorscalarBinop(op parser.ItemType, lhs Vector, rhs Scalar, swap, returnBool bool, enh *EvalNodeHelper) (Vector, error) {
	for _, lhsSample := range lhs {
		lv, rv := lhsSample.V, rhs.V
		// lhs always contains the Vector. If the original position was different
		// swap for calculating the value.
		if swap {
			lv, rv = rv, lv
		}
		value, keep, err := vectorElemBinop(op, lv, rv)
		if err != nil {
			return nil, err
		}

		// Catch cases where the scalar is the LHS in a scalar-vector comparison operation.
		// We want to always keep the vector element value as the output value, even if it's on the RHS.
		if op.IsComparisonOperator() && swap {
			value = rv
		}
		if returnBool {
			if keep {
				value = 1.0
			} else {
				value = 0.0
			}
			keep = true
		}

		if keep {
			lhsSample.V = value
			enh.Out = append(enh.Out, lhsSample)
		}
	}
	return enh.Out, nil
}

// VectorBinop evaluates a binary operation between two Vectors, excluding set operators.
func VectorBinop(op parser.ItemType, lhs, rhs Vector, matching *parser.VectorMatching, returnBool bool, enh *EvalNodeHelper) (Vector, error) {
	// 四则运算的匹配类型是OneToOne
	if matching.Card == parser.CardManyToMany {
		return nil, uerrors.PromQLError{
			Typ: uerrors.ErrorExec,
			Err: fmt.Errorf("many-to-many only allowed for set operators"),
		}
	}
	// 判断是多对一的话，抛尚未支持的异常. 一对多也不支持，
	if matching.Card == parser.CardManyToOne || matching.Card == parser.CardOneToMany {
		return nil, uerrors.PromQLError{
			Typ: uerrors.ErrorExec,
			Err: fmt.Errorf("many-to-one or one-to-many operators is not supported"),
		}
	}
	// 如果是left_join的多对多，进行以下join操作
	if matching.Card == parser.LeftJoinManyToMany {

		// 如果右向量为空、直接返回左侧向量
		if len(rhs) == 0 {
			enh.Out = lhs
			return enh.Out, nil
		}

		res := make(Vector, 0, 64)
		for _, leftSample := range lhs {
			// 把leftMetric转化成map，方便快速根据label找到值
			leftMetricMap := make(map[string]string)
			for _, label := range leftSample.Metric {
				leftMetricMap[label.Name] = label.Value
			}

			matchedMetrics := make([]Sample, 0)

			// 遍历右侧序列，寻找匹配的序列
			for _, rightSample := range rhs {
				// 把rightMetric转化成map，方便快速根据label找到值
				rightMetricMap := make(map[string]string)
				for _, label := range rightSample.Metric {
					rightMetricMap[label.Name] = label.Value
				}

				// 判断是否匹配
				match := matchLable(leftMetricMap, rightMetricMap, matching)
				if match {
					matchedMetrics = append(matchedMetrics, rightSample)
				}
			}

			if len(matchedMetrics) != 0 {
				// 匹配到多条序列，根据group_left里的label对原序列进行扩充
				// 如果公式中没有写group_left，在一对多的情况下报错
				if !matching.ExistGroupLeft && len(matchedMetrics) > 1 {
					return nil, uerrors.PromQLError{
						Typ: uerrors.ErrorExec,
						Err: fmt.Errorf("one-to-many need group_left in the formula"),
					}
				}

				// 如果匹配到了多条序列，判断扩充的序列是否会重复
				if len(matchedMetrics) > 1 {
					if len(matching.Include) == 0 {
						return nil, uerrors.PromQLError{
							Typ: uerrors.ErrorExec,
							Err: fmt.Errorf("vector cannot contain metrics with the same labelset"),
						}
					}

					// 把匹配的序列的标签全部转化成map[string]string类型的数组
					matchedMetricArr := make([]map[string]string, 0)
					for _, matchedSample := range matchedMetrics {
						matchedMetricMap := make(map[string]string, 0)
						for _, label := range matchedSample.Metric {
							matchedMetricMap[label.Name] = label.Value
						}
						matchedMetricArr = append(matchedMetricArr, matchedMetricMap)
					}

					// 把标签的值拼接判断是否重复
					uniqueMap := make(map[string]bool)
					for _, m := range matchedMetricArr {
						combinedValues := ""
						for _, label := range matching.Include {
							combinedValues += m[label]
						}

						if uniqueMap[combinedValues] {
							return nil, uerrors.PromQLError{
								Typ: uerrors.ErrorExec,
								Err: fmt.Errorf("out_join match result vector cannot contain metrics with the same labelset"),
							}
						}

						uniqueMap[combinedValues] = true
					}

				}

				// 进行扩充
				for _, matchedSample := range matchedMetrics {
					matchedMetricMap := make(map[string]string)
					for _, label := range matchedSample.Metric {
						matchedMetricMap[label.Name] = label.Value
					}

					joinMetric := join(leftSample, matchedSample, leftMetricMap, matchedMetricMap, op, matching)
					res = append(res, joinMetric)
				}

			} else {
				// 没有匹配到序列
				res = append(res, leftSample)
			}

		}

		enh.Out = res
		return enh.Out, nil
	}

	// 如果是 out_join 的多对多，进行以下 join 操作
	if matching.Card == parser.OutJoinManyToMany {
		// todo: out_join的实现中，使用了hashmap来取labels，为了保证顺序，又使用了一个数组。可考虑使用红黑树来实现。
		return processOutJoin(op, lhs, rhs, matching, enh)
	}

	// 根据关联操作类型和匹配的字段列表组装成的字符串来生成标识符
	sigf := enh.signatureFunc(matching.On, matching.MatchingLabels...)

	// All samples from the rhs hashed by the matching label/values.
	if enh.rightSigs == nil {
		enh.rightSigs = make(map[string]Sample, len(enh.Out))
	} else {
		for k := range enh.rightSigs {
			delete(enh.rightSigs, k)
		}
	}
	rightSigs := enh.rightSigs

	// Add all rhs samples to a map so we can easily find matches later.
	for _, rs := range rhs {
		// 先把右边的 vector 的 labels 组成一个字符串，然后放在一个map中。
		sig := sigf(rs.Metric)
		// The rhs is guaranteed to be the 'one' side. Having multiple samples
		// with the same signature means that the matching is many-to-many.
		if _, found := rightSigs[sig]; found {
			return nil, uerrors.PromQLError{
				Typ: uerrors.ErrorExec,
				Err: fmt.Errorf("found duplicate series for the match group, many-to-many only allowed for set operators"),
			}
		}
		rightSigs[sig] = rs
	}

	// Tracks the match-signature. For one-to-one operations the value is nil. For many-to-one
	// the value is a set of signatures to detect duplicated result elements.
	if enh.matchedSigs == nil {
		enh.matchedSigs = make(map[string]map[uint64]struct{}, len(rightSigs))
	} else {
		for k := range enh.matchedSigs {
			delete(enh.matchedSigs, k)
		}
	}
	matchedSigs := enh.matchedSigs

	// For all lhs samples find a respective rhs sample and perform
	// the binary operation.
	for _, ls := range lhs {
		sig := sigf(ls.Metric)

		_, exists := matchedSigs[sig]
		if matching.Card == parser.CardOneToOne {
			if exists {
				// 一对一的请求中出现vector数据是多对一，抛异常
				return nil, uerrors.PromQLError{
					Typ: uerrors.ErrorExec,
					Err: fmt.Errorf("multiple matches for labels: many-to-one matching must be explicit (group_left/group_right)"),
				}
			}
			matchedSigs[sig] = nil // Set existence to true.
		}

		rs, found := rightSigs[sig] // Look for a match in the rhs Vector.
		if !found {
			continue
		}

		// Account for potentially swapped sidedness.
		vl, vr := ls.V, rs.V
		value, keep, err := vectorElemBinop(op, vl, vr)
		if err != nil {
			return nil, err
		}
		if returnBool {
			if keep {
				value = 1.0
			} else {
				value = 0.0
			}
		} else if !keep {
			continue
		}

		metric := resultMetric(ls.Metric, rs.Metric, matching, enh)

		enh.Out = append(enh.Out, Sample{
			Metric: metric,
			Point:  Point{V: value},
		})
	}
	return enh.Out, nil
}

func processOutJoin(op parser.ItemType, lhs, rhs Vector, matching *parser.VectorMatching, enh *EvalNodeHelper) (Vector, error) {
	// 给序列打标记，所有label作为一个序列描述，所以基于所有label给序列打标记
	sigf := enh.signatureFunc(false, "")
	lSampleSig, rSampleSig := make(map[string]bool), make(map[string]bool)
	lSamples, rSamples := make(map[string]Sample), make(map[string]Sample)
	lSamplesArr, rSamplesArr := make([]string, 0), make([]string, 0)
	for _, ls := range lhs {
		sig := sigf(ls.Metric)
		lSampleSig[sig], lSamples[sig] = false, ls
		lSamplesArr = append(lSamplesArr, sig)
	}
	for _, rs := range rhs {
		sig := sigf(rs.Metric)
		rSampleSig[sig], rSamples[sig] = false, rs
		rSamplesArr = append(rSamplesArr, sig)
	}

	res := make(Vector, 0)
	// 先以左边为准，进行关联取值
	for _, lsig := range lSamplesArr {
		leftSample := lSamples[lsig]
		// 把leftMetric转化成map，方便快速根据label找到值
		// 组装来自于匹配字段和 select_left
		// 如果 sample 的metric只有一个 __tsid，则matching.MatchingLabels设置为__tsid
		if len(leftSample.Metric) == 1 && leftSample.Metric[0].Name == interfaces.TSID {
			matching.MatchingLabels = []string{interfaces.TSID}
		}
		labelArr, labelMap, leftMetricMap := getJoinedLabels(matching.MatchingLabels, matching.IncludeLeft, leftSample.Metric)

		matchedMetrics := make([]Sample, 0)
		// 遍历右侧序列，寻找匹配的序列
		for _, rsig := range rSamplesArr {
			rightSample := rSamples[rsig]
			// 把rightMetric转化成map，方便快速根据label找到值
			rightMetricMap := make(map[string]string)
			for _, label := range rightSample.Metric {
				rightMetricMap[label.Name] = label.Value
			}

			// 判断是否匹配
			match := matchLable(leftMetricMap, rightMetricMap, matching)
			if match {
				matchedMetrics = append(matchedMetrics, rightSample)
				lSampleSig[lsig] = true
				rSampleSig[rsig] = true
			}
		}

		if len(matchedMetrics) > 1 && len(matching.IncludeLeft) == 0 && len(matching.IncludeRight) == 0 {
			// 按关联字段匹配到了多条，若此时select_left和select_right都不指定，就只有关联字段是维度，必然存在重复
			return nil, uerrors.PromQLError{
				Typ: uerrors.ErrorExec,
				Err: fmt.Errorf("out_join match result cannot contain metrics with the same labelset"),
			}
		}

		if len(matchedMetrics) > 0 {
			// 匹配到多条序列，根据select_lefthe select_right里的label对结果序列进行扩充
			// 组装结果序列
			for _, matchedSample := range matchedMetrics {
				matchedMetricMap := make(map[string]string)
				for _, label := range matchedSample.Metric {
					matchedMetricMap[label.Name] = label.Value
				}

				// 2. 来自于 select_right 的维度。用新的数组来接
				labelArri := make([]string, 0)
				labelArri = append(labelArri, labelArr...)
				for _, label := range matching.IncludeRight {
					if _, exists := matchedMetricMap[label]; exists {
						labelArri = append(labelArri, label)
						labelMap[label] = matchedMetricMap[label]
					}
				}

				// 计算结果样本点
				sample := getJoinedSample(op, leftSample.T, leftSample.V, matchedSample.V, labelArri, labelMap)
				res = append(res, sample)
			}

		} else {
			// 没有匹配到序列。维度描述为 out_join 和 select_left中的字段
			sample := getJoinedSample(op, leftSample.T, leftSample.V, 0, labelArr, labelMap)
			res = append(res, sample)
			lSampleSig[lsig] = true
		}
	}

	// 补充右边的
	for k, rightSample := range rSamples {
		if !rSampleSig[k] {
			// false为未与左边匹配的，单独处理append到结果序列中
			// 组装维度，来自于匹配字段和 select_right
			labelArr, labelMap, _ := getJoinedLabels(matching.MatchingLabels, matching.IncludeRight, rightSample.Metric)

			sample := getJoinedSample(op, rightSample.T, 0, rightSample.V, labelArr, labelMap)
			res = append(res, sample)
			lSampleSig[k] = true
		}
	}

	// 是否包含相同的labels
	if res.ContainsSameLabelset() {
		return nil, uerrors.PromQLError{
			Typ: uerrors.ErrorExec,
			Err: fmt.Errorf("vector cannot contain metrics with the same labelset"),
		}
	}
	enh.Out = res
	return enh.Out, nil
}

// out join的结果已经准备好，组装成sample，labels和数据点
func getJoinedSample(op parser.ItemType, pointT int64, leftV, rightV float64, labelArr []string,
	labelMap map[string]string) Sample {

	// 计算指标值
	floatValue, _, _ := vectorElemBinop(op, leftV, rightV)

	var sample Sample
	sample.Point.V = floatValue
	sample.Point.T = pointT

	sort.Strings(labelArr)
	// 维度append到sample中
	for _, label := range labelArr {
		sample.Metric = append(sample.Metric, &labels.Label{
			Name:  label,
			Value: labelMap[label],
		})
	}
	return sample
}

// 根据匹配字段和提取字段获取关联结果中所需要字段信息列表
func getJoinedLabels(matchingLabels, includeLabels []string, labels labels.Labels) ([]string, map[string]string, map[string]string) {
	metricMap := make(map[string]string)
	for _, label := range labels {
		metricMap[label.Name] = label.Value
	}

	labelArr := make([]string, 0)
	labelMap := make(map[string]string)
	for _, label := range matchingLabels {
		if _, exists := metricMap[label]; exists {
			labelArr = append(labelArr, label)
			labelMap[label] = metricMap[label]
		}
	}
	for _, label := range includeLabels {
		if _, exists := metricMap[label]; exists {
			labelArr = append(labelArr, label)
			labelMap[label] = metricMap[label]
		}
	}
	return labelArr, labelMap, metricMap
}

// PreprocessExpr wraps all possible step invariant parts of the given expression with
// StepInvariantExpr. It also resolves the preprocessors.
func PreprocessExpr(expr parser.Expr, start, end int64) (parser.Expr, error) {
	isStepInvariant, err := preprocessExprHelper(expr, start, end)
	if err != nil {
		return nil, err
	}
	if isStepInvariant {
		return newStepInvariantExpr(expr), nil
	}
	return expr, nil
}

// preprocessExprHelper wraps the child nodes of the expression
// with a StepInvariantExpr wherever it's step invariant. The returned boolean is true if the
// passed expression qualifies to be wrapped by StepInvariantExpr.
// It also resolves the preprocessors.
func preprocessExprHelper(expr parser.Expr, start, end int64) (bool, error) {
	switch n := expr.(type) {
	case *parser.VectorSelector:
		if n.StartOrEnd == parser.START {
			n.Timestamp = makeInt64Pointer(start)
		} else if n.StartOrEnd == parser.END {
			n.Timestamp = makeInt64Pointer(end)
		}
		return n.Timestamp != nil, nil

	case *parser.AggregateExpr:
		//逐层preprocessExprHelper
		return preprocessExprHelper(n.Expr, start, end)

	case *parser.BinaryExpr:
		isInvariant1, err1 := preprocessExprHelper(n.LHS, start, end)
		isInvariant2, err2 := preprocessExprHelper(n.RHS, start, end)

		if err1 != nil || err2 != nil {
			return false, uerrors.PromQLError{
				Typ: uerrors.ErrorExec,
				Err: fmt.Errorf("%s; %s", err1, err2),
			}
		}
		if isInvariant1 && isInvariant2 {
			return true, nil
		}

		if isInvariant1 {
			n.LHS = newStepInvariantExpr(n.LHS)
		}
		if isInvariant2 {
			n.RHS = newStepInvariantExpr(n.RHS)
		}

		return false, nil

	case *parser.ParenExpr:
		return preprocessExprHelper(n.Expr, start, end)

	case *parser.UnaryExpr:
		return preprocessExprHelper(n.Expr, start, end)

	case *parser.StringLiteral, *parser.NumberLiteral:
		return true, nil

	case *parser.Call:
		//_, ok := AtModifierUnsafeFunctions[n.Func.Name]
		//isStepInvariant := !ok
		//isStepInvariantSlice := make([]bool, len(n.Args))
		//var err error
		//
		//for i := range n.Args {
		//	isStepInvariantSlice[i], err = preprocessExprHelper(n.Args[i], start, end)
		//	if err != nil {
		//		return false, err
		//	}
		//	isStepInvariant = isStepInvariant && isStepInvariantSlice[i]
		//}
		//
		//if isStepInvariant {
		//	// The function and all arguments are step invariant.
		//	return true, nil
		//}
		//
		//for i, isi := range isStepInvariantSlice {
		//	if isi {
		//		n.Args[i] = newStepInvariantExpr(n.Args[i])
		//	}
		//}
		return false, nil

	default:
		return false, uerrors.PromQLError{
			Typ: uerrors.ErrorExec,
			Err: fmt.Errorf("found unexpected node %#v", expr),
		}
	}
}

func makeInt64Pointer(val int64) *int64 {
	valp := new(int64)
	*valp = val
	return valp
}

func newStepInvariantExpr(expr parser.Expr) parser.Expr {
	if e, ok := expr.(*parser.ParenExpr); ok {
		// Wrapping the inside of () makes it easy to unwrap the paren later.
		// But this effectively unwraps the paren.
		return newStepInvariantExpr(e.Expr)
	}
	return &parser.StepInvariantExpr{Expr: expr}
}

// resultMetric returns the metric for the given sample(s) based on the Vector
// binary operation and the matching options.
func resultMetric(lhs, rhs labels.Labels, matching *parser.VectorMatching, enh *EvalNodeHelper) labels.Labels {
	if enh.resultMetric == nil {
		enh.resultMetric = make(map[string]labels.Labels, len(enh.Out))
	}

	if enh.lb == nil {
		enh.lb = labels.NewBuilder(lhs)
	} else {
		enh.lb.Reset(lhs)
	}

	buf := bytes.NewBuffer(enh.lblResultBuf[:0])
	enh.lblBuf = lhs.Bytes(enh.lblBuf)
	buf.Write(enh.lblBuf)
	enh.lblBuf = rhs.Bytes(enh.lblBuf)
	buf.Write(enh.lblBuf)
	enh.lblResultBuf = buf.Bytes()

	if ret, ok := enh.resultMetric[string(enh.lblResultBuf)]; ok {
		return ret
	}
	str := string(enh.lblResultBuf)

	if matching.Card == parser.CardOneToOne {
		if matching.On {
		Outer:
			for _, l := range lhs {
				for _, n := range matching.MatchingLabels {
					if l.Name == n {
						continue Outer
					}
				}
				enh.lb.Del(l.Name)
			}
		} else {
			enh.lb.Del(matching.MatchingLabels...)
		}
	}

	ret := enh.lb.Labels()
	enh.resultMetric[str] = ret
	return ret
}

var pointPool = sync.Pool{}

func GetPointSlice(sz int) []Point {
	p := pointPool.Get()
	if p != nil {
		return p.([]Point)
	}
	return make([]Point, 0, sz)
}

func PutPointSlice(p []Point) {
	//nolint:staticcheck // Ignore SA6002 relax staticcheck verification.
	pointPool.Put(p[:0])
}

// AtModifierUnsafeFunctions are the functions whose result
// can vary if evaluation time is changed when the arguments are
// step invariant. It also includes functions that use the timestamps
// of the passed instant vector argument to calculate a result since
// that can also change with change in eval time.
// var AtModifierUnsafeFunctions = map[string]struct{}{
// 	// Step invariant functions.
// 	"days_in_month": {}, "day_of_month": {}, "day_of_week": {},
// 	"hour": {}, "minute": {}, "month": {}, "year": {},
// 	"predict_linear": {}, "time": {},
// 	// Uses timestamp of the argument for the result,
// 	// hence unsafe to use with @ modifier.
// 	"timestamp": {},
// }

// unwrapParenExpr does the AST equivalent of removing parentheses around a expression.
func UnwrapParenExpr(e *parser.Expr) {
	for {
		if p, ok := (*e).(*parser.ParenExpr); ok {
			*e = p.Expr
		} else {
			break
		}
	}
}

func UnwrapStepInvariantExpr(e parser.Expr) parser.Expr {
	if p, ok := e.(*parser.StepInvariantExpr); ok {
		return p.Expr
	}
	return e
}

// groupingKey builds and returns the grouping key for the given metric and
// grouping labels.
func GenerateGroupingKey(metric labels.Labels, grouping []string, without bool, buf []byte) (uint64, []byte) {
	if without {
		return metric.HashWithoutLabels(buf, grouping...)
	}

	if len(grouping) == 0 {
		// No need to generate any hash if there are no grouping labels.
		return 0, buf
	}

	return metric.HashForLabels(buf, grouping...)
}

// aggregation evaluates an aggregation operation on a Vector. The provided grouping labels
// must be sorted.
func Aggregation(op parser.ItemType, grouping []string, without bool, param interface{}, vec Vector, seriesHelper []EvalSeriesHelper,
	enh *EvalNodeHelper) (Vector, error) {
	result := map[uint64]*groupedAggregation{}
	orderedResult := []*groupedAggregation{}

	var k int64
	if op == parser.TOPK || op == parser.BOTTOMK {
		f := param.(float64)
		if !convert.ConvertibleToInt64(f) {
			return nil, uerrors.PromQLError{
				Typ: uerrors.ErrorExec,
				Err: fmt.Errorf("Scalar value %v overflows int64", f),
			}
		}
		k = int64(f)
		if k < 1 {
			return Vector{}, nil
		}
	}
	// recomputeGroupingKey用于 count_values 聚合,暂时先去掉
	// var recomputeGroupingKey bool
	lb := labels.NewBuilder(nil)
	// var buf []byte
	for si, s := range vec {
		metric := s.Metric

		// count_values 聚合时才需要 recomputed grouping key, 先注掉
		// We can use the pre-computed grouping key unless grouping labels have changed.
		// var groupingKey uint64
		// if !recomputeGroupingKey {
		groupingKey := seriesHelper[si].GroupingKey
		// } else {
		// 	groupingKey, buf = GenerateGroupingKey(metric, grouping, without, buf)
		// }

		group, ok := result[groupingKey]
		// Add a new group if it doesn't exist.
		if !ok {
			lb.Reset(metric)
			if without {
				lb.Del(grouping...)
				lb.Del(labels.MetricName)
			} else {
				lb.Keep(grouping...)
			}
			m := lb.Labels()
			newAgg := &groupedAggregation{
				labels:     m,
				value:      s.V,
				mean:       s.V,
				groupCount: 1,
			}

			result[groupingKey] = newAgg
			orderedResult = append(orderedResult, newAgg)

			inputVecLen := int64(len(vec))
			resultSize := k
			if k > inputVecLen {
				resultSize = inputVecLen
			}
			switch op {
			case parser.TOPK:
				result[groupingKey].heap = make(vectorByValueHeap, 0, resultSize)
				heap.Push(&result[groupingKey].heap, &Sample{
					Point:  Point{V: s.V},
					Metric: s.Metric,
				})
			case parser.BOTTOMK:
				result[groupingKey].reverseHeap = make(vectorByReverseValueHeap, 0, resultSize)
				heap.Push(&result[groupingKey].reverseHeap, &Sample{
					Point:  Point{V: s.V},
					Metric: s.Metric,
				})
			}
			continue
		}

		switch op {
		case parser.SUM:
			group.value += s.V

		case parser.AVG:
			group.groupCount++
			if math.IsInf(group.mean, 0) {
				if math.IsInf(s.V, 0) && (group.mean > 0) == (s.V > 0) {
					// The `mean` and `s.V` values are `Inf` of the same sign.  They
					// can't be subtracted, but the value of `mean` is correct
					// already.
					break
				}
				if !math.IsInf(s.V, 0) && !math.IsNaN(s.V) {
					// At this stage, the mean is an infinite. If the added
					// value is neither an Inf or a Nan, we can keep that mean
					// value.
					// This is required because our calculation below removes
					// the mean value, which would look like Inf += x - Inf and
					// end up as a NaN.
					break
				}
			}
			// Divide each side of the `-` by `group.groupCount` to avoid float64 overflows.
			group.mean += s.V/float64(group.groupCount) - group.mean/float64(group.groupCount)

		case parser.COUNT:
			group.groupCount++
		case parser.MAX:
			if group.value < s.V || math.IsNaN(group.value) {
				group.value = s.V
			}
		case parser.MIN:
			if group.value > s.V || math.IsNaN(group.value) {
				group.value = s.V
			}

		case parser.TOPK:
			if int64(len(group.heap)) < k || group.heap[0].V < s.V || math.IsNaN(group.heap[0].V) {
				if int64(len(group.heap)) == k {
					heap.Pop(&group.heap)
				}
				heap.Push(&group.heap, &Sample{
					Point:  Point{V: s.V},
					Metric: s.Metric,
				})
			}

		case parser.BOTTOMK:
			if int64(len(group.reverseHeap)) < k || group.reverseHeap[0].V > s.V || math.IsNaN(group.reverseHeap[0].V) {
				if int64(len(group.reverseHeap)) == k {
					heap.Pop(&group.reverseHeap)
				}
				heap.Push(&group.reverseHeap, &Sample{
					Point:  Point{V: s.V},
					Metric: s.Metric,
				})
			}

		default:
			return nil, uerrors.PromQLError{
				Typ: uerrors.ErrorExec,
				Err: fmt.Errorf("expected aggregation operator but got %q", op),
			}
		}
	}

	// Construct the result Vector from the aggregated groups.
	for _, aggr := range orderedResult {
		switch op {
		case parser.AVG:
			aggr.value = aggr.mean

		case parser.COUNT:
			aggr.value = float64(aggr.groupCount)

		case parser.TOPK:
			// The heap keeps the lowest value on top, so reverse it.
			sort.Sort(sort.Reverse(aggr.heap))
			for _, v := range aggr.heap {
				enh.Out = append(enh.Out, Sample{
					Metric: v.Metric,
					Point:  Point{V: v.V},
				})
			}
			continue // Bypass default append.

		case parser.BOTTOMK:
			// The heap keeps the highest value on top, so reverse it.
			sort.Sort(sort.Reverse(aggr.reverseHeap))
			for _, v := range aggr.reverseHeap {
				enh.Out = append(enh.Out, Sample{
					Metric: v.Metric,
					Point:  Point{V: v.V},
				})
			}
			continue // Bypass default append.

		default:
			// For other aggregations, we already have the right value.
		}

		enh.Out = append(enh.Out, Sample{
			Metric: aggr.labels,
			Point:  Point{V: aggr.value},
		})
	}
	return enh.Out, nil
}

// 解析 match[] 参数
func ParseMatchersParam(matchers []string) ([][]*labels.Matcher, error) {
	var matcherSets [][]*labels.Matcher

	for _, s := range matchers {
		matchers, err := parser.ParseMetricSelector(s)
		if err != nil {
			return nil, err
		}
		matcherSets = append(matcherSets, matchers)
	}

	// 逐个 match[] 判断其 matchers 匹配的 value 是否都为空.
OUTER:
	for _, ms := range matcherSets {
		for _, lm := range ms {
			if lm != nil && !lm.Matches("") {
				// 如果遍历当前 match[] 的 matchers 遇到了匹配的 value 不是 "", 那么就退出当前 match[] 的遍历，继续遍历下一个 match[]
				continue OUTER
			}
		}
		return nil, fmt.Errorf("match[] must contain at least one non-empty matcher")
	}
	return matcherSets, nil
}

// matchLable 判断连接条件
func matchLable(leftMetricMap, rightMetricMap map[string]string, matching *parser.VectorMatching) bool {
	if len(matching.MatchingLabels) == 0 {
		return false
	}
	// 左右两边的维度都是空集时，表示可以关联成功，返回true
	if len(leftMetricMap) == 0 && len(rightMetricMap) == 0 {
		return true
	}

	// 先对left_join中的label和leftMetric的label做交集
	intersection := make([]string, 0)
	// 如果左右两边是 __tsid，则交集是__tsid
	_, lExist := leftMetricMap[interfaces.TSID]
	_, rExist := rightMetricMap[interfaces.TSID]
	if lExist && rExist {
		intersection = append(intersection, interfaces.TSID)
	} else {
		for _, label := range matching.MatchingLabels {
			if _, exists := leftMetricMap[label]; exists {
				intersection = append(intersection, label)
			}
		}
	}
	if len(intersection) == 0 {
		return false
	}

	// 匹配过程是先把left_join中的label和leftMetric的label做交集，用交集去和rightMetric去做匹配
	for _, label := range intersection {
		if _, exists := rightMetricMap[label]; !exists {
			return false
		}

		if leftMetricMap[label] != rightMetricMap[label] {
			return false
		}
	}

	return true
}

// join 整理join后的标签以及指标值
func join(leftSample, rightSample Sample, leftMetricMap, rightMetricMap map[string]string, op parser.ItemType,
	matching *parser.VectorMatching) Sample {
	var res Sample

	// 计算指标值
	floatValue, _, _ := vectorElemBinop(op, leftSample.Point.V, rightSample.Point.V)
	res.Point.V = floatValue
	res.Point.T = leftSample.T

	// 整合join后的label列表，如果扩充的标签已经在存在，用右侧的值覆盖左侧的值
	for _, label := range matching.Include {
		if _, exist := rightMetricMap[label]; exist {
			leftMetricMap[label] = rightMetricMap[label]
		}
	}

	labelArr := make([]string, 0, len(leftMetricMap))
	for key := range leftMetricMap {
		labelArr = append(labelArr, key)
	}
	sort.Strings(labelArr)

	for _, label := range labelArr {
		res.Metric = append(res.Metric, &labels.Label{
			Name:  label,
			Value: leftMetricMap[label],
		})
	}

	return res
}

// 按填充策略对每个指标结果集填充
func FillMissingPoint(query interfaces.Query, matrixes []Matrix, precedingMissingPolicy, middleMissingPolicy int) []Matrix {
	// 填充策略与填充值转换
	precedingFilling := float64(precedingMissingPolicy)
	if precedingMissingPolicy == -1 {
		precedingFilling = 0
	}
	middleFiliing := float64(middleMissingPolicy)
	if middleMissingPolicy == -1 {
		middleFiliing = 0
	}
	for i, mat := range matrixes {
		for j, ss := range mat {
			// 创建一个 map 来快速查找已有的时间点数据
			pointMap := make(map[int64]float64)
			for _, point := range ss.Points {
				pointMap[point.T] = point.V
			}
			// 创建一个切片来存储结果，包括填充的点
			pointsI := make([]Point, 0)
			currentTime := query.FixedStart

			// 处理开始时间到第一个数据点之间的数据缺失。若是没有数据，points为空，则按前序策略从fixedStart填充到fixedEnd
			firstPointT := query.FixedEnd
			if len(ss.Points) > 0 {
				firstPointT = ss.Points[0].T
			}
			for currentTime < firstPointT {
				pointsI = append(pointsI, Point{T: currentTime, V: precedingFilling})
				// currentTime += query.Interval
				currentTime = GetNextPointTime(query, currentTime)
			}

			// 处理已有的数据点和它们之间的数据缺失
			// 如果第一个点不存在，则使用填充值作为初始值（这里其实不会用到，因为上面已经处理了开始到第一个点）
			lastValue := float64(precedingMissingPolicy)
			for currentTime < query.End { // end是小于的关系，不能等于end。否则会多计算一个数据点
				if value, exists := pointMap[currentTime]; exists {
					pointsI = append(pointsI, Point{T: currentTime, V: value})
					lastValue = value // 更新最后一个已知值
				} else {
					// 根据填充策略处理中间缺失的点
					fillValue := middleFiliing
					if middleMissingPolicy == 0 { // 沿用前一分钟的状态
						fillValue = lastValue
					}
					pointsI = append(pointsI, Point{T: currentTime, V: fillValue})
				}
				// currentTime += query.Interval
				currentTime = GetNextPointTime(query, currentTime)
			}
			mat[j].Points = pointsI
		}
		matrixes[i] = mat
	}
	return matrixes
}

// 联合多个指标来计算整体的可用性.都非0才可用
func CombineEvalUsability(matrixes []Matrix, precedingMissingPolicy, middleMissingPolicy int) Matrix {
	enh := &EvalNodeHelper{Out: make(Vector, 0)}
	sigf := enh.signatureFunc(false, "")
	// 因为每个表达式的序列都补点了，时间上是对齐的，所以表达式下的序列在每个时间点上都有值，
	// 则表达式下的第一个时间点的序列可以代表当下指标的序列集。
	// 1.先计算每个表达式下的序列集
	mergedMap := make(map[string]Series)
	// 初始化一个映射来记录每个键在所有对象中出现的次数
	keyCounts := make(map[string]int)
	// 记录series的顺序
	seriesOrd := make([]string, 0)
	for _, mat := range matrixes {
		for _, series := range mat {
			sig := sigf(series.Metric)
			keyCounts[sig]++

			// 如果key不存在于mergedMap中，则初始化
			if _, exists := mergedMap[sig]; !exists {
				// ps := make([]Point, len(series.Points))
				mergedMap[sig] = series
				seriesOrd = append(seriesOrd, sig)
			}

			// 遍历points并更新mergeSeries中的值
			for i, point := range series.Points {
				// 检查当前t上的v值是否都非0
				allNonZero := true
				if mergedMap[sig].Points[i].V == 0 || point.V == 0 {
					allNonZero = false
				}
				// 更新mergedMap中对应t的v值的“合并状态”（这里用V字段来存储合并后的结果，0或1）
				if mergedMap[sig].Points[i].V != 0 && allNonZero {
					// 如果之前已经是1且当前也满足都非0，则保持为1
					// 或者这是第一个点，直接设置合并状态
					mergedMap[sig].Points[i].T = point.T
					mergedMap[sig].Points[i].V = 1
				} else {
					// 否则设置为0
					mergedMap[sig].Points[i].T = point.T
					mergedMap[sig].Points[i].V = 0
				}
			}
		}
	}
	// 2. 按策略填充不完整的序列
	// 填充策略与填充值转换
	precedingFilling := float64(precedingMissingPolicy)
	if precedingMissingPolicy == -1 {
		precedingFilling = 0
	}
	for sig, count := range keyCounts {
		if count != len(matrixes) {
			// 按preceding_missing_policy策略把mergedMap中的point逐个处理
			for i, point := range mergedMap[sig].Points {
				if point.V == 1 && precedingFilling == 1 {
					mergedMap[sig].Points[i].V = 1
				} else {
					mergedMap[sig].Points[i].V = 0
				}
			}
		}
	}

	// map转数组
	var mat Matrix
	for _, sig := range seriesOrd {
		mat = append(mat, mergedMap[sig])
	}
	return mat
}

// 计算连续超过5分钟不可用的时长总和
func CalculateUnavailableTime(mat Matrix, query interfaces.Query, kMinute int64) Matrix {
	// 分即时查询和范围查询。
	// 范围查询计算的是每个步长点在每个步长区间内的不可用时长；
	// 即时查询计算的是查询时间范围内的不可用时长。

	kmin := float64(kMinute)
	if query.IsInstantQuery {
		return calculateUnavailableTime4InstantQuery(mat, query, kmin)
	}
	return calculateUnavailableTime4RangeQuery(query, mat, kMinute, kmin)
}

// 计算范围查询的不可用时长
func calculateUnavailableTime4RangeQuery(query interfaces.Query, mat Matrix, kMinute int64, kmin float64) Matrix {
	var matrix Matrix
	// 遍历序列，对每个step做累加不可用时长的计算
	// 表示每个步长时间点上计算不可用时间的数据的时间区间为1个步长，趋势图时不考虑步长交叉时的连续不可用时长
	// 计算每个点的不可用时长
	stepNum := query.Interval / interfaces.KMINUTE_DOWNTTIME_STEP

	// 计算遍历开始时间
	firstPointT := query.FixedStart - (stepNum-1)*60*1000
	for _, series := range mat {
		finalPoints := make([]Point, 0)
		var lastIndex int
		// for ts := query.FixedStart; ts <= query.FixedEnd; ts += query.Interval {
		for ts := query.FixedStart; ts <= query.FixedEnd; ts = GetNextPointTime(query, ts) {

			if query.Interval < kMinute*60*1000 {
				// 请求步长小于k-minute，那么每个step上的不可用时长为0
				finalPoints = append(finalPoints, Point{T: ts, V: 0})
				continue
			}

			// 大于等于k-minute，计算每个step上的累计不可用时长。
			// 每个步长点的不可用时长为当前点往前追k-1个数据点来计算
			var unavailableDuration float64 // 累积不可用时长
			var unavailableMinutes float64  // 记录连续不可用时间段的持续分钟数，遇到一个+1
			isUnavailable := false          // 标记当前是否处于不可用状态
			// var startUnavailabel float64    // 前序不可用时间长度，且大于等于5分钟
			for i := lastIndex; i < len(series.Points); i++ {
				lastIndex = i
				// point遍历到当前ts时，当前step的遍历就该结束了，继续下一个step的计算
				if series.Points[i].T > ts {
					break
				}
				if series.Points[i].T < firstPointT {
					// 前序的点不考虑，从firstPointT开始计算不可用时长
					continue
				}

				if isUnavailable {
					// 如果当前仍然处于不可用状态
					if series.Points[i].V == 0 {
						// 继续检查下一个点
						unavailableMinutes++
						// 检查是否是数组中的最后一个点，并且它是不可用的
						// 如果是，我们需要计算到当前时间（假设当前时间作为查询的结束时间）的持续时间
						if series.Points[i].T == ts {
							if unavailableMinutes >= kmin {
								unavailableDuration += unavailableMinutes
							}
						}
					} else {
						// 当前点变为可用，计算不可用的持续时间
						if unavailableMinutes >= kmin {
							unavailableDuration += unavailableMinutes
						}
						// 重置状态
						isUnavailable = false
						unavailableMinutes = 0
					}
				} else {
					// 如果当前处于可用状态
					if series.Points[i].V == 0 {
						// 当前点变为不可用，记录开始时间
						unavailableMinutes++
						isUnavailable = true
						// 检查是否是数组中的最后一个点，并且它是不可用的
						// 如果是，我们需要计算到当前时间（假设当前时间作为查询的结束时间）的持续时间
						if series.Points[i].T == ts {
							if unavailableMinutes >= kmin {
								unavailableDuration += unavailableMinutes
							}
						}
					}
					// 如果当前点仍然可用，则不改变状态
				}
			}
			finalPoints = append(finalPoints, Point{T: ts, V: unavailableDuration})
		}
		matrix = append(matrix, Series{
			Metric: series.Metric,
			Points: finalPoints,
		})
	}
	return matrix
}

// 计算即时查询的不可用时长
func calculateUnavailableTime4InstantQuery(mat Matrix, query interfaces.Query, kmin float64) Matrix {
	// 遍历序列，进行累加不可用时长的计算
	// query的fixedStart是计算的时间起始点。
	// 计算交叉点的不可用时长
	var matrix Matrix
	for _, series := range mat {
		var unavailableDuration float64 // 累积不可用时长
		var unavailableMinutes float64  // 记录连续不可用时间段的持续分钟数，遇到一个+1
		isUnavailable := false          // 标记当前是否处于不可用状态
		var startUnavailabel float64    // 前序不可用时间长度，且大于等于5分钟
		for i, point := range series.Points {
			if isUnavailable {
				// 如果当前仍然处于不可用状态
				if point.T == query.FixedStart {
					// 分割点，fixedSart之前的不可用时间长度是参考值
					// 若分割点之前的时间的不可用时长大于0，小于5，则加上这个时长，与后续的数据进行比较看是否能形成连续时间
					// 若等于5，因为这5分钟已经被统计在前一个时间窗口上了，在加上的不可用时间段时长度时减去5分钟
					if unavailableMinutes >= kmin {
						startUnavailabel = unavailableMinutes
					}
				}

				if point.V == 0 {
					// 继续检查下一个点
					unavailableMinutes++

					// 检查是否是数组中的最后一个点，并且它是不可用的
					// 如果是，我们需要计算到当前时间（假设当前时间作为查询的结束时间）的持续时间
					if i == len(series.Points)-1 {
						if unavailableMinutes >= kmin {
							if startUnavailabel >= kmin {
								unavailableMinutes -= startUnavailabel
								startUnavailabel = 0
							}
							unavailableDuration += unavailableMinutes
						}
					}
				} else {
					// 当前点变为可用，计算不可用的持续时间
					if unavailableMinutes >= kmin {
						if startUnavailabel >= kmin {
							unavailableMinutes -= startUnavailabel
							startUnavailabel = 0
						}
						unavailableDuration += unavailableMinutes
					}
					// 重置状态
					isUnavailable = false
					unavailableMinutes = 0
				}
			} else {
				// 如果当前处于可用状态
				if point.V == 0 {
					// 当前点变为不可用，记录开始时间
					unavailableMinutes++
					isUnavailable = true
					// 检查是否是数组中的最后一个点，并且它是不可用的
					// 如果是，我们需要计算到当前时间（假设当前时间作为查询的结束时间）的持续时间
					if i == len(series.Points)-1 {
						if unavailableMinutes >= kmin {
							if startUnavailabel >= kmin {
								unavailableMinutes -= startUnavailabel
								startUnavailabel = 0
							}
							unavailableDuration += unavailableMinutes
						}
					}
				}
				// 如果当前点仍然可用，则不改变状态
			}
		}
		matrix = append(matrix, Series{
			Metric: series.Metric,
			Points: []Point{{T: query.End, V: unavailableDuration}},
		})
	}
	return matrix
}

// correctingTime 修正开始时间和结束时间，符合opensearch的分桶区间
func CorrectingTime(query interfaces.Query, zoneLocation *time.Location) (int64, int64) {

	// 将时间戳转换为时间
	startTime := time.UnixMilli(query.Start)
	endTime := time.UnixMilli(query.End)

	// 如果是日历间隔，按照日历间隔进行修正时间
	if query.IsCalendar {
		switch query.IntervalStr {
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
		// stepT, _ := convert.ParseDuration(query.IntervalStr)
		// step := stepT.Milliseconds()

		// 先计算桶，然后再按照时区偏移
		_, offset := startTime.In(zoneLocation).Zone()
		fixedStart := int64(math.Floor(float64(query.Start+int64(offset*1000))/float64(query.Interval)))*query.Interval - int64(offset*1000)
		fixedEnd := int64(math.Floor(float64(query.End+int64(offset*1000))/float64(query.Interval)))*query.Interval - int64(offset*1000)

		return fixedStart, fixedEnd
	}

	return 0, 0
}

// getNextPointTime 获取下一个时间点
func GetNextPointTime(query interfaces.Query, currentTime int64) int64 {
	if query.IsCalendar {
		// 将时间戳转换为时间对象
		switch query.IntervalStr {
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
		return currentTime + query.Interval
	}

	return 0
}
