// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package static

import (
	"regexp"
	"sort"

	"uniquery/logics/promql/labels"
	"uniquery/logics/promql/parser"
)

const errMsgNoManyToMany = "set operations must only use many-to-many matching"

// EvalSeriesHelper stores extra information about a series.
type EvalSeriesHelper struct {
	// The grouping key used by aggregation.
	GroupingKey uint64
	// Used to map left-hand to right-hand in binary operations.
	Signature string
}

// EvalNodeHelper stores extra information and caches for evaluating a single node across steps.
type EvalNodeHelper struct {
	// Evaluation timestamp.
	Ts int64
	// Vector that can be used for output.
	Out Vector

	// Caches.
	// DropMetricName and label_*.
	Dmn map[uint64]labels.Labels
	// signatureFunc.
	sigf map[string]string
	// funcHistogramQuantile.
	signatureToMetricWithBuckets map[string]*metricWithBuckets
	// label_replace.
	Regex *regexp.Regexp

	lb           *labels.Builder
	lblBuf       []byte
	lblResultBuf []byte

	// For binary vector matching.
	rightSigs    map[string]Sample
	matchedSigs  map[string]map[uint64]struct{}
	resultMetric map[string]labels.Labels
}

func (enh *EvalNodeHelper) signatureFunc(on bool, names ...string) func(labels.Labels) string {
	if enh.sigf == nil {
		enh.sigf = make(map[string]string, len(enh.Out))
	}
	f := signatureFunc(on, enh.lblBuf, names...)

	return func(l labels.Labels) string {
		enh.lblBuf = l.Bytes(enh.lblBuf)
		ret, ok := enh.sigf[string(enh.lblBuf)]
		if ok {
			return ret
		}
		ret = f(l)
		enh.sigf[string(enh.lblBuf)] = ret
		return ret
	}
}

func signatureFunc(on bool, b []byte, names ...string) func(labels.Labels) string {
	sort.Strings(names)
	if on {
		return func(lset labels.Labels) string {
			return string(lset.WithLabels(names...).Bytes(b))
		}
	}

	return func(lset labels.Labels) string {
		return string(lset.WithoutLabels(names...).Bytes(b))
	}
}

func VectorAnd(lhs, rhs Vector, matching *parser.VectorMatching, enh *EvalNodeHelper) Vector {
	if matching.Card != parser.CardManyToMany {
		panic(errMsgNoManyToMany)
	}
	sigf := enh.signatureFunc(matching.On, matching.MatchingLabels...)

	// The set of signatures for the right-hand side Vector.
	rightSigs := map[string]struct{}{}
	// Add all rhs samples to a map so we can easily find matches later.
	for _, rs := range rhs {
		rightSigs[sigf(rs.Metric)] = struct{}{}
	}

	for _, ls := range lhs {
		// If there's a matching entry in the right-hand side Vector, add the sample.
		if _, ok := rightSigs[sigf(ls.Metric)]; ok {
			enh.Out = append(enh.Out, ls)
		}
	}
	return enh.Out
}

func VectorOr(lhs, rhs Vector, matching *parser.VectorMatching, enh *EvalNodeHelper) Vector {
	if matching.Card != parser.CardManyToMany {
		panic(errMsgNoManyToMany)
	}
	sigf := enh.signatureFunc(matching.On, matching.MatchingLabels...)

	leftSigs := map[string]struct{}{}
	// Add everything from the left-hand-side Vector.
	for _, ls := range lhs {
		leftSigs[sigf(ls.Metric)] = struct{}{}
		enh.Out = append(enh.Out, ls)
	}
	// Add all right-hand side elements which have not been added from the left-hand side.
	for _, rs := range rhs {
		if _, ok := leftSigs[sigf(rs.Metric)]; !ok {
			enh.Out = append(enh.Out, rs)
		}
	}
	return enh.Out
}

func VectorUnless(lhs, rhs Vector, matching *parser.VectorMatching, enh *EvalNodeHelper) Vector {
	if matching.Card != parser.CardManyToMany {
		panic(errMsgNoManyToMany)
	}
	sigf := enh.signatureFunc(matching.On, matching.MatchingLabels...)

	rightSigs := map[string]struct{}{}
	for _, rs := range rhs {
		rightSigs[sigf(rs.Metric)] = struct{}{}
	}

	for _, ls := range lhs {
		if _, ok := rightSigs[sigf(ls.Metric)]; !ok {
			enh.Out = append(enh.Out, ls)
		}
	}
	return enh.Out
}
