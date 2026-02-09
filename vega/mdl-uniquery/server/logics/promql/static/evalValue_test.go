// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package static

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"uniquery/logics/promql/labels"
	"uniquery/logics/promql/parser"
)

var (
	lhsVectors = []Sample{
		{
			Point: Point{
				T: 1652320539000,
				V: 1,
			},
			Metric: []*labels.Label{
				{Name: "cluster", Value: "opensearch"},
				{Name: "node", Value: "node-1"},
			},
		},
		{
			Point: Point{
				T: 1652320539000,
				V: 2,
			},
			Metric: []*labels.Label{
				{Name: "cluster", Value: "opensearch2"},
				{Name: "node", Value: "node-2"},
			},
		},
	}

	rhsVector1 = []Sample{
		{
			Point: Point{
				T: 1652320539000,
				V: 2,
			},
			Metric: []*labels.Label{
				{Name: "cluster", Value: "opensearch"},
				{Name: "node", Value: "node-1"},
			},
		},
	}

	rhsVector3 = []Sample{
		{
			Point: Point{
				T: 1652320539000,
				V: 3,
			},
			Metric: []*labels.Label{
				{Name: "cluster", Value: "opensearch3"},
				{Name: "node", Value: "node-3"},
			},
		},
	}

	matchers = parser.VectorMatching{Card: parser.CardManyToMany, MatchingLabels: []string{}, On: false, Include: []string{}}
)

func TestVectorAnd(t *testing.T) {
	Convey("test VectorAnd", t, func() {

		Convey("vectorAnd sameLabel", func() {
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, lhsVectors[0])

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result := VectorAnd(lhsVectors, rhsVector1, &matchers, enh)
			So(result, ShouldResemble, expectVector)
		})

		Convey("vectorAnd diffLabel", func() {
			expectVector := make(Vector, 0)

			enh2 := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result2 := VectorAnd(lhsVectors, rhsVector3, &matchers, enh2)

			So(result2, ShouldResemble, expectVector)
		})

	})
}

func TestVectorOr(t *testing.T) {
	Convey("test VectorOr", t, func() {

		Convey("vectorOr sameLabel", func() {
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, lhsVectors...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result := VectorOr(lhsVectors, rhsVector1, &matchers, enh)
			So(result, ShouldResemble, expectVector)
		})

		Convey("vectorOr diffLabel", func() {
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, lhsVectors...)
			expectVector = append(expectVector, rhsVector3...)

			enh2 := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result2 := VectorOr(lhsVectors, rhsVector3, &matchers, enh2)

			So(result2, ShouldResemble, expectVector)
		})

	})
}

func TestVectorUnless(t *testing.T) {
	Convey("test VectorUnless", t, func() {

		Convey("vectorUnless sameLabel", func() {
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, lhsVectors[1])

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result := VectorUnless(lhsVectors, rhsVector1, &matchers, enh)
			So(result, ShouldResemble, expectVector)
		})

		Convey("vectorUnless diffLabel", func() {
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, lhsVectors...)

			enh2 := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result2 := VectorUnless(lhsVectors, rhsVector3, &matchers, enh2)

			So(result2, ShouldResemble, expectVector)
		})

	})
}
