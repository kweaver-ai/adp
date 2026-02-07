// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package static

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bytedance/sonic"

	"uniquery/logics/promql/labels"
	"uniquery/logics/promql/parser"
)

func (Matrix) Type() parser.ValueType       { return parser.ValueTypeMatrix }
func (Vector) Type() parser.ValueType       { return parser.ValueTypeVector }
func (Scalar) Type() parser.ValueType       { return parser.ValueTypeScalar }
func (String) Type() parser.ValueType       { return parser.ValueTypeString }
func (IrateMatrix) Type() parser.ValueType  { return parser.ValueTypeMatrix }
func (IrateSeries) Type() parser.ValueType  { return parser.ValueTypeMatrix }
func (RatePoint) Type() parser.ValueType    { return parser.ValueTypeMatrix }
func (ChangesPoint) Type() parser.ValueType { return parser.ValueTypeMatrix }
func (DeltaPoint) Type() parser.ValueType   { return parser.ValueTypeMatrix }
func (PageMatrix) Type() parser.ValueType   { return parser.ValueTypeMatrix }

// String represents a string value.
type String struct {
	T int64
	V string
}

func (s String) String() string {
	return s.V
}

// Scalar is a data point that's explicitly not associated with a metric.
type Scalar struct {
	T int64
	V float64
}

func (s Scalar) String() string {
	v := strconv.FormatFloat(s.V, 'f', -1, 64)
	return fmt.Sprintf("scalar: %v @[%v]", v, s.T)
}

func (s Scalar) MarshalJSON() ([]byte, error) {
	v := strconv.FormatFloat(s.V, 'f', -1, 64)
	return sonic.Marshal([...]interface{}{float64(s.T) / 1000, v})
}

// Series is a stream of data points belonging to a metric.
type Series struct {
	Metric labels.Labels `json:"metric"`
	Points []Point       `json:"values"`
}

func (s Series) String() string {
	vals := make([]string, len(s.Points))
	for i, v := range s.Points {
		vals[i] = v.String()
	}
	return fmt.Sprintf("%s =>\n%s", s.Metric.String(), strings.Join(vals, "\n"))
}

// Point represents a single data point for a given timestamp.
type Point struct {
	T int64
	V float64
}

func (p Point) String() string {
	v := strconv.FormatFloat(p.V, 'f', -1, 64)
	return fmt.Sprintf("%v @[%v]", v, p.T)
}

// MarshalJSON implements sonic.Marshaler.
func (p Point) MarshalJSON() ([]byte, error) {
	v := strconv.FormatFloat(p.V, 'f', -1, 64)
	return sonic.Marshal([...]interface{}{float64(p.T) / 1000, v})
}

// Sample is a single sample belonging to a metric.
type Sample struct {
	Point
	Metric labels.Labels
}

func (s Sample) MarshalJSON() ([]byte, error) {
	v := struct {
		M labels.Labels `json:"metric"`
		V Point         `json:"value"`
	}{
		M: s.Metric,
		V: s.Point,
	}
	return sonic.Marshal(v)
}

// Vector is basically only an alias for model.Samples, but the
// contract is that in a Vector, all Samples have the same timestamp.
type Vector []Sample

func (vec Vector) String() string {
	entries := make([]string, len(vec))
	for i, s := range vec {
		entries[i] = s.String()
	}
	return strings.Join(entries, "\n")
}

func (vec Vector) ContainsSameLabelset() bool {
	l := make(map[uint64]struct{}, len(vec))
	for _, s := range vec {
		hash := s.Metric.Hash()
		if _, ok := l[hash]; ok {
			return true
		}
		l[hash] = struct{}{}
	}
	return false
}

// Matrix is a slice of Series that implements sort.Interface and
// has a String method.
type Matrix []Series
type PageMatrix struct {
	Matrix
	TotalSeries int
}

func (m Matrix) String() string {
	// (fabxc): sort, or can we rely on order from the querier?
	strs := make([]string, len(m))

	for i, ss := range m {
		strs[i] = ss.String()
	}

	return strings.Join(strs, "\n")
}

type RatePoint struct {
	FirstTimestamp    int64
	FirstValue        float64
	LastTimestamp     int64
	LastValue         float64
	CounterCorrection float64
	PointsCount       int64
}

func (p RatePoint) String() string {
	v := strconv.FormatFloat(p.FirstValue, 'f', -1, 64)
	return fmt.Sprintf("%v @[%v]", v, p.FirstTimestamp)
}

// ContainsSameLabelset checks if a matrix has samples with the same labelset.
// Such a behavior is semantically undefined.
// https://github.com/prometheus/prometheus/issues/4562
func (m Matrix) ContainsSameLabelset() bool {
	l := make(map[uint64]struct{}, len(m))
	for _, ss := range m {
		hash := ss.Metric.Hash()
		if _, ok := l[hash]; ok {
			return true
		}
		l[hash] = struct{}{}
	}
	return false
}

type IrateMatrix []IrateSeries

type IratePoint struct {
	PreviousT int64
	PreviousV float64
	LastT     int64
	LastV     float64
}

type IrateSeries struct {
	Metric labels.Labels        `json:"metric"`
	Points map[int64]IratePoint `json:"values"`
}

func (p IratePoint) String() string {

	previousV := strconv.FormatFloat(p.PreviousV, 'f', -1, 64)
	lastV := strconv.FormatFloat(p.LastV, 'f', -1, 64)
	return fmt.Sprintf("lastValue:%v,previousValue:%v,lastTimestamp:%v,previousTimestamp:%v", lastV, previousV, p.LastT, p.PreviousT)
}

func (i IrateSeries) String() string {
	vals := make([]string, len(i.Points))
	for i, v := range i.Points {
		vals[i] = v.String()
	}
	return fmt.Sprintf("%s =>\n%s", i.Metric, strings.Join(vals, "\n"))
}
func (irate IrateMatrix) String() string {
	// (fabxc): sort, or can we rely on order from the querier?
	strs := make([]string, len(irate))

	for i, ss := range irate {
		strs[i] = ss.String()
	}

	return strings.Join(strs, "\n")
}

type ChangesPoint struct {
	FirstTimestamp int64
	FirstValue     float64
	LastTimestamp  int64
	LastValue      float64
	Changes        int64
}

func (p ChangesPoint) String() string {
	firstV := strconv.FormatFloat(p.FirstValue, 'f', -1, 64)
	lastV := strconv.FormatFloat(p.LastValue, 'f', -1, 64)
	return fmt.Sprintf("FirstTimestamp:%v,FirstValue:%v,LastTimestamp:%v,LastValue:%v,Changes:%d", p.FirstTimestamp, firstV, p.LastTimestamp, lastV, p.Changes)
}

type AGGPoint struct {
	Value float64
	Count int64
}

func (p AGGPoint) String() string {
	lastV := strconv.FormatFloat(p.Value, 'f', -1, 64)
	return fmt.Sprintf("Value:%v,Count:%d", lastV, p.Count)
}

type DeltaPoint struct {
	FirstTimestamp int64
	FirstValue     float64
	LastTimestamp  int64
	LastValue      float64
	PointsCount    int64
}

func (p DeltaPoint) String() string {
	firstV := strconv.FormatFloat(p.FirstValue, 'f', -1, 64)
	lastV := strconv.FormatFloat(p.LastValue, 'f', -1, 64)
	return fmt.Sprintf("FirstTimestamp:%v,FirstValue:%v,LastTimestamp:%v,LastValue:%v,PointsCount:%v", p.FirstTimestamp,
		firstV, p.LastTimestamp, lastV, p.PointsCount)
}

func (mMatrix PageMatrix) String() string {
	// (fabxc): sort, or can we rely on order from the querier?
	strs := make([]string, len(mMatrix.Matrix))

	for i, ss := range mMatrix.Matrix {
		strs[i] = ss.String()
	}

	return strings.Join(strs, "\n")
}
