// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package static

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/require"

	"uniquery/interfaces"
	"uniquery/logics/promql/labels"
	"uniquery/logics/promql/parser"
)

var (
	testCtx = context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage)
)

var lhs, rhs float64 = 1, 2

func TestScalarBinop(t *testing.T) {
	Convey("test ScalarBinop", t, func() {

		Convey("scalar op scalar is ADD ", func() {
			result, err := ScalarBinop(parser.ADD, lhs, rhs)
			So(result, ShouldEqual, 3)
			So(err, ShouldBeNil)

		})

		Convey("scalar op scalar is SUB ", func() {
			result, err := ScalarBinop(parser.SUB, lhs, rhs)
			So(result, ShouldEqual, -1)
			So(err, ShouldBeNil)

		})

		Convey("scalar op scalar is MUL ", func() {
			result, err := ScalarBinop(parser.MUL, lhs, rhs)
			So(result, ShouldEqual, 2)
			So(err, ShouldBeNil)

		})

		Convey("scalar op scalar is DIV ", func() {
			result, err := ScalarBinop(parser.DIV, lhs, rhs)
			So(result, ShouldEqual, 0.5)
			So(err, ShouldBeNil)

		})

		Convey("scalar op scalar is POW ", func() {
			result, err := ScalarBinop(parser.POW, lhs, rhs)
			So(result, ShouldEqual, 1)
			So(err, ShouldBeNil)

		})

		Convey("scalar op scalar is MOD ", func() {
			result, err := ScalarBinop(parser.MOD, lhs, rhs)
			So(result, ShouldEqual, 1)
			So(err, ShouldBeNil)

		})

		Convey("scalar op scalar is EQLC ", func() {
			result, err := ScalarBinop(parser.EQLC, lhs, rhs)
			So(result, ShouldEqual, 0)
			So(err, ShouldBeNil)
		})

		Convey("scalar op scalar is NEQ ", func() {
			result, err := ScalarBinop(parser.NEQ, lhs, rhs)
			So(result, ShouldEqual, 1)
			So(err, ShouldBeNil)
		})

		Convey("scalar op scalar is GTR ", func() {
			result, err := ScalarBinop(parser.GTR, lhs, rhs)
			So(result, ShouldEqual, 0)
			So(err, ShouldBeNil)
		})

		Convey("scalar op scalar is LSS ", func() {
			result, err := ScalarBinop(parser.LSS, lhs, rhs)
			So(result, ShouldEqual, 1)
			So(err, ShouldBeNil)
		})

		Convey("scalar op scalar is GTE ", func() {
			result, err := ScalarBinop(parser.GTE, lhs, rhs)
			So(result, ShouldEqual, 0)
			So(err, ShouldBeNil)
		})

		Convey("scalar op scalar is LTE ", func() {
			result, err := ScalarBinop(parser.LTE, lhs, rhs)
			So(result, ShouldEqual, 1)
			So(err, ShouldBeNil)
		})
	})
}

func TestVectorElemBinop(t *testing.T) {
	Convey("test VectorElemBinop", t, func() {

		Convey("VectorElem op VectorElem is ADD ", func() {
			result, b, err := vectorElemBinop(parser.ADD, lhs, rhs)
			So(result, ShouldEqual, 3)
			So(b, ShouldBeTrue)
			So(err, ShouldBeNil)

		})

		Convey("VectorElem op VectorElem is SUB ", func() {
			result, b, err := vectorElemBinop(parser.SUB, lhs, rhs)
			So(result, ShouldEqual, -1)
			So(b, ShouldBeTrue)
			So(err, ShouldBeNil)

		})

		Convey("VectorElem op VectorElem is MUL ", func() {
			result, b, err := vectorElemBinop(parser.MUL, lhs, rhs)
			So(result, ShouldEqual, 2)
			So(b, ShouldBeTrue)
			So(err, ShouldBeNil)

		})

		Convey("VectorElem op VectorElem is DIV ", func() {
			result, b, err := vectorElemBinop(parser.DIV, lhs, rhs)
			So(result, ShouldEqual, 0.5)
			So(b, ShouldBeTrue)
			So(err, ShouldBeNil)

		})

		Convey("VectorElem op VectorElem is POW ", func() {
			result, b, err := vectorElemBinop(parser.POW, lhs, rhs)
			So(result, ShouldEqual, 1)
			So(b, ShouldBeTrue)
			So(err, ShouldBeNil)

		})

		Convey("VectorElem op VectorElem is MOD ", func() {
			result, b, err := vectorElemBinop(parser.MOD, lhs, rhs)
			So(result, ShouldEqual, 1)
			So(b, ShouldBeTrue)
			So(err, ShouldBeNil)

		})

		Convey("VectorElem op VectorElem is EQLC ", func() {
			result, b, err := vectorElemBinop(parser.EQLC, lhs, rhs)
			So(result, ShouldEqual, 1)
			So(b, ShouldBeFalse)
			So(err, ShouldBeNil)
		})

		Convey("VectorElem op VectorElem is NEQ ", func() {
			result, b, err := vectorElemBinop(parser.NEQ, lhs, rhs)
			So(result, ShouldEqual, 1)
			So(b, ShouldBeTrue)
			So(err, ShouldBeNil)
		})

		Convey("VectorElem op VectorElem is GTR ", func() {
			result, b, err := vectorElemBinop(parser.GTR, lhs, rhs)
			So(result, ShouldEqual, 1)
			So(b, ShouldBeFalse)
			So(err, ShouldBeNil)
		})

		Convey("VectorElem op VectorElem is LSS ", func() {
			result, b, err := vectorElemBinop(parser.LSS, lhs, rhs)
			So(result, ShouldEqual, 1)
			So(b, ShouldBeTrue)
			So(err, ShouldBeNil)
		})

		Convey("VectorElem op VectorElem is GTE ", func() {
			result, b, err := vectorElemBinop(parser.GTE, lhs, rhs)
			So(result, ShouldEqual, 1)
			So(b, ShouldBeFalse)
			So(err, ShouldBeNil)
		})

		Convey("VectorElem op VectorElem is LTE ", func() {
			result, b, err := vectorElemBinop(parser.LTE, lhs, rhs)
			So(result, ShouldEqual, 1)
			So(b, ShouldBeTrue)
			So(err, ShouldBeNil)
		})
	})
}

var (
	lhsVector = []Sample{
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
	}

	rhsVector = []Sample{
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
	rhsScalar = Scalar{T: 0, V: 2}

	matcher   = parser.VectorMatching{Card: parser.CardOneToOne, MatchingLabels: []string{}, On: false, Include: []string{}}
	onMatcher = parser.VectorMatching{Card: parser.CardOneToOne, MatchingLabels: []string{"cluster"}, On: true, Include: []string{}}
)

func TestVectorscalarBinop(t *testing.T) {
	Convey("test VectorscalarBinop", t, func() {

		Convey("vector op scalar is ADD ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 1652320539000,
						V: 3,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorscalarBinop(parser.ADD, lhsVector, rhsScalar, false, false, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("scalar op vector is ADD ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 1652320539000,
						V: 3,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh2 := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result2, err2 := VectorscalarBinop(parser.ADD, lhsVector, rhsScalar, true, false, enh2)
			So(result2, ShouldResemble, expectVector)
			So(err2, ShouldBeNil)
		})

		Convey("vector op scalar is SUB ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 1652320539000,
						V: -1,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorscalarBinop(parser.SUB, lhsVector, rhsScalar, false, false, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)

		})

		Convey("scalar op vector is SUB ", func() {
			expected := []Sample{
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
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh2 := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result2, err2 := VectorscalarBinop(parser.SUB, lhsVector, rhsScalar, true, false, enh2)
			So(result2, ShouldResemble, expectVector)
			So(err2, ShouldBeNil)
		})

		Convey("vector op scalar is MUL ", func() {
			expected := []Sample{
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
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorscalarBinop(parser.MUL, lhsVector, rhsScalar, false, false, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)

		})

		Convey("scalar op vector is MUL ", func() {
			expected := []Sample{
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
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh2 := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result2, err2 := VectorscalarBinop(parser.MUL, lhsVector, rhsScalar, true, false, enh2)
			So(result2, ShouldResemble, expectVector)
			So(err2, ShouldBeNil)
		})

		Convey("vector op scalar is DIV ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 1652320539000,
						V: 0.5,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorscalarBinop(parser.DIV, lhsVector, rhsScalar, false, false, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("scalar op vector is DIV ", func() {
			expected := []Sample{
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
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh2 := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result2, err2 := VectorscalarBinop(parser.DIV, lhsVector, rhsScalar, true, false, enh2)
			So(result2, ShouldResemble, expectVector)
			So(err2, ShouldBeNil)
		})

		Convey("vector op scalar is POW ", func() {
			expected := []Sample{
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
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorscalarBinop(parser.POW, lhsVector, rhsScalar, false, false, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("scalar op vector is POW ", func() {
			expected := []Sample{
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
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh2 := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result2, err2 := VectorscalarBinop(parser.POW, lhsVector, rhsScalar, true, false, enh2)
			So(result2, ShouldResemble, expectVector)
			So(err2, ShouldBeNil)
		})

		Convey("vector op scalar is MOD ", func() {
			expected := []Sample{
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
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorscalarBinop(parser.MOD, lhsVector, rhsScalar, false, false, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("scalar op vector is MOD ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 1652320539000,
						V: 0,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh2 := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result2, err2 := VectorscalarBinop(parser.MOD, lhsVector, rhsScalar, true, false, enh2)
			So(result2, ShouldResemble, expectVector)
			So(err2, ShouldBeNil)
		})

		Convey("vector op scalar is EQLC ", func() {
			expectVector := make(Vector, 0)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorscalarBinop(parser.EQLC, lhsVector, rhsScalar, false, false, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("scalar op vector is EQLC ", func() {
			expectVector := make(Vector, 0)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorscalarBinop(parser.EQLC, lhsVector, rhsScalar, true, false, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("vector op scalar is NEQ ", func() {
			expected := []Sample{
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
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorscalarBinop(parser.NEQ, lhsVector, rhsScalar, false, false, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("scalar op vector is NEQ ", func() {
			expected := []Sample{
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
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorscalarBinop(parser.NEQ, lhsVector, rhsScalar, true, false, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("vector op scalar is GTR ", func() {
			expectVector := make(Vector, 0)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorscalarBinop(parser.GTR, lhsVector, rhsScalar, false, false, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("scalar op vector is GTR ", func() {
			expected := []Sample{
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
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorscalarBinop(parser.GTR, lhsVector, rhsScalar, true, false, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("vector op scalar is LSS ", func() {
			expected := []Sample{
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
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorscalarBinop(parser.LSS, lhsVector, rhsScalar, false, false, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("scalar op vector is LSS ", func() {
			expectVector := make(Vector, 0)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorscalarBinop(parser.LSS, lhsVector, rhsScalar, true, false, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("vector op scalar is GTE ", func() {
			expectVector := make(Vector, 0)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorscalarBinop(parser.GTE, lhsVector, rhsScalar, false, false, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("scalar op vector is GTE ", func() {
			expected := []Sample{
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
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorscalarBinop(parser.GTE, lhsVector, rhsScalar, true, false, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("vector op scalar is LTE ", func() {
			expected := []Sample{
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
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorscalarBinop(parser.LTE, lhsVector, rhsScalar, false, false, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("scalar op vector is LTE ", func() {
			expectVector := make(Vector, 0)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorscalarBinop(parser.LTE, lhsVector, rhsScalar, true, false, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("vector op scalar is EQLC BOOL ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 1652320539000,
						V: 0,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorscalarBinop(parser.EQLC, lhsVector, rhsScalar, false, true, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("scalar op vector is EQLC BOOL ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 1652320539000,
						V: 0,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorscalarBinop(parser.EQLC, lhsVector, rhsScalar, true, true, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("vector op scalar is NEQ BOOL ", func() {
			expected := []Sample{
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
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorscalarBinop(parser.NEQ, lhsVector, rhsScalar, false, true, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("scalar op vector is NEQ BOOL ", func() {
			expected := []Sample{
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
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorscalarBinop(parser.NEQ, lhsVector, rhsScalar, true, true, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("vector op scalar is GTR BOOL ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 1652320539000,
						V: 0,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorscalarBinop(parser.GTR, lhsVector, rhsScalar, false, true, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("scalar op vector is GTR BOOL ", func() {
			expected := []Sample{
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
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorscalarBinop(parser.GTR, lhsVector, rhsScalar, true, true, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("vector op scalar is LSS BOOL ", func() {
			expected := []Sample{
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
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorscalarBinop(parser.LSS, lhsVector, rhsScalar, false, true, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("scalar op vector is LSS BOOL ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 1652320539000,
						V: 0,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorscalarBinop(parser.LSS, lhsVector, rhsScalar, true, true, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("vector op scalar is GTE BOOL ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 1652320539000,
						V: 0,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorscalarBinop(parser.GTE, lhsVector, rhsScalar, false, true, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("scalar op vector is GTE BOOL ", func() {
			expected := []Sample{
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
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorscalarBinop(parser.GTE, lhsVector, rhsScalar, true, true, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("vector op scalar is LTE BOOL ", func() {
			expected := []Sample{
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
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorscalarBinop(parser.LTE, lhsVector, rhsScalar, false, true, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("scalar op vector is LTE BOOL ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 1652320539000,
						V: 0,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorscalarBinop(parser.LTE, lhsVector, rhsScalar, true, true, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})
	})
}

func TestVectorBinop(t *testing.T) {
	Convey("test VectorBinop", t, func() {
		leftVector := []Sample{
			{
				Point: Point{
					T: 1652320539000,
					V: 1,
				},
				Metric: []*labels.Label{
					{Name: "cluster", Value: "opensearch"},
					{Name: "node", Value: "node-1"},
					{Name: "cpu", Value: "0"},
					{Name: "k8s_app", Value: "node-exporter"},
				},
			},
		}

		rightVector := []Sample{
			{
				Point: Point{
					T: 1652320539000,
					V: 2,
				},
				Metric: []*labels.Label{
					{Name: "cluster", Value: "opensearch"},
					{Name: "node", Value: "node-1"},
					{Name: "job", Value: "kubernetes-services-endpoints"},
					{Name: "namespace", Value: "kube-syste"},
				},
			},
		}
		Convey("Vector op Vector is ADD ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 0,
						V: 3,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorBinop(parser.ADD, lhsVector, rhsVector, &matcher, false, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("Vector op Vector is SUB ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 0,
						V: -1,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorBinop(parser.SUB, lhsVector, rhsVector, &matcher, false, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)

		})

		Convey("Vector op Vector is MUL ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 0,
						V: 2,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorBinop(parser.MUL, lhsVector, rhsVector, &matcher, false, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)

		})

		Convey("Vector op Vector is DIV ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 0,
						V: 0.5,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorBinop(parser.DIV, lhsVector, rhsVector, &matcher, false, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("Vector op Vector is POW ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 0,
						V: 1,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorBinop(parser.POW, lhsVector, rhsVector, &matcher, false, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("Vector op Vector is MOD ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 0,
						V: 1,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorBinop(parser.MOD, lhsVector, rhsVector, &matcher, false, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("ManyToMany error ", func() {
			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			manyToManyMatcher := parser.VectorMatching{Card: parser.CardManyToMany, MatchingLabels: []string{}, On: false, Include: []string{}}
			result, err := VectorBinop(parser.MOD, lhsVector, rhsVector, &manyToManyMatcher, false, enh)

			So(result, ShouldBeNil)
			So(err.Error(), ShouldEqual, "many-to-many only allowed for set operators")
		})

		Convey("ManyToOne error ", func() {
			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			manyToManyMatcher := parser.VectorMatching{Card: parser.CardManyToOne, MatchingLabels: []string{}, On: false, Include: []string{}}
			result, err := VectorBinop(parser.MOD, lhsVector, rhsVector, &manyToManyMatcher, false, enh)

			So(result, ShouldBeNil)
			So(err.Error(), ShouldEqual, "many-to-one or one-to-many operators is not supported")
		})

		Convey("leftVector op left_join(label list) group_left(label list) rightVector , rightVector is nil", func() {
			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			leftjoinMatcher := parser.VectorMatching{
				Card:           parser.LeftJoinManyToMany,
				MatchingLabels: []string{},
				LeftJoin:       true,
				ExistGroupLeft: true,
				Include:        []string{},
			}
			result, err := VectorBinop(parser.ADD, leftVector, []Sample{}, &leftjoinMatcher, false, enh)
			var expectedVector Vector = Vector{
				{
					Point: Point{
						T: 1652320539000,
						V: 1,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
						{Name: "cpu", Value: "0"},
						{Name: "k8s_app", Value: "node-exporter"},
					},
				},
			}
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectedVector)
		})

		Convey("leftVector op left_join(label list) group_left(label list) rightVector successful,no matches found situations", func() {
			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			leftjoinMatcher := parser.VectorMatching{
				Card:           parser.LeftJoinManyToMany,
				MatchingLabels: []string{"cluster", "cpu"},
				LeftJoin:       true,
				ExistGroupLeft: true,
				Include:        []string{},
			}

			expectedVector := Vector{
				{
					Point: Point{
						T: 1652320539000,
						V: 1,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
						{Name: "cpu", Value: "0"},
						{Name: "k8s_app", Value: "node-exporter"},
					},
				},
			}
			result, err := VectorBinop(parser.ADD, leftVector, rightVector, &leftjoinMatcher, false, enh)

			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectedVector)
		})

		Convey("leftVector op left_join(label list) group_left(label list) rightVector successful,one to one situations", func() {
			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			leftjoinMatcher := parser.VectorMatching{
				Card:           parser.LeftJoinManyToMany,
				MatchingLabels: []string{"cluster"},
				LeftJoin:       true,
				ExistGroupLeft: true,
				Include:        []string{},
			}

			expectedVector := Vector{
				{
					Point: Point{
						T: 1652320539000,
						V: 3,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "cpu", Value: "0"},
						{Name: "k8s_app", Value: "node-exporter"},
						{Name: "node", Value: "node-1"},
					},
				},
			}
			result, err := VectorBinop(parser.ADD, leftVector, rightVector, &leftjoinMatcher, false, enh)

			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectedVector)
		})

		Convey("leftVector op left_join(label list) group_left(label list) rightVector,one to many situations", func() {
			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			manyToOneMatcher := parser.VectorMatching{
				Card:           parser.LeftJoinManyToMany,
				MatchingLabels: []string{"cluster"},
				LeftJoin:       true,
				ExistGroupLeft: true,
				Include:        []string{"namespace", "job"},
			}
			rightVector = []Sample{
				{
					Point: Point{
						T: 1652320539000,
						V: 2,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
						{Name: "job", Value: "kubernetes-services-endpoints"},
						{Name: "namespace", Value: "kube-syste"},
					},
				},
				{
					Point: Point{
						T: 1652320539000,
						V: 3,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-2"},
						{Name: "job", Value: "kubernetes-services-endpoints2"},
						{Name: "namespace", Value: "kube-syste2"},
					},
				},
			}

			expectedVector := Vector{
				{
					Point: Point{
						T: 1652320539000,
						V: 3,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "cpu", Value: "0"},
						{Name: "job", Value: "kubernetes-services-endpoints"},
						{Name: "k8s_app", Value: "node-exporter"},
						{Name: "namespace", Value: "kube-syste"},
						{Name: "node", Value: "node-1"},
					},
				},
				{
					Point: Point{
						T: 1652320539000,
						V: 4,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "cpu", Value: "0"},
						{Name: "job", Value: "kubernetes-services-endpoints2"},
						{Name: "k8s_app", Value: "node-exporter"},
						{Name: "namespace", Value: "kube-syste2"},
						{Name: "node", Value: "node-1"},
					},
				},
			}
			result, err := VectorBinop(parser.ADD, leftVector, rightVector, &manyToOneMatcher, false, enh)

			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectedVector)
		})

		Convey("leftVector op left_join(label list) rightVector,one to many situations", func() {
			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			manyToOneMatcher := parser.VectorMatching{
				Card:           parser.LeftJoinManyToMany,
				MatchingLabels: []string{"cluster"},
				LeftJoin:       true,
				ExistGroupLeft: false,
				Include:        []string{},
			}
			rightVector = []Sample{
				{
					Point: Point{
						T: 1652320539000,
						V: 2,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
						{Name: "job", Value: "kubernetes-services-endpoints"},
						{Name: "namespace", Value: "kube-syste"},
					},
				},
				{
					Point: Point{
						T: 1652320539000,
						V: 3,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-2"},
						{Name: "job", Value: "kubernetes-services-endpoints2"},
						{Name: "namespace", Value: "kube-syste2"},
					},
				},
			}

			result, err := VectorBinop(parser.ADD, leftVector, rightVector, &manyToOneMatcher, false, enh)

			So(result, ShouldBeNil)
			So(err.Error(), ShouldEqual, "one-to-many need group_left in the formula")
		})

		Convey("many to many error ", func() {
			lhsVectorMutil := []Sample{
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
						T: 1652320542000,
						V: 1,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
			}

			rhsVectorMutil := []Sample{
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
				{
					Point: Point{
						T: 1652320542000,
						V: 2,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
			}

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorBinop(parser.MOD, lhsVectorMutil, rhsVectorMutil, &matcher, false, enh)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldEqual, "found duplicate series for the match group, many-to-many only allowed for set operators")
		})

		Convey("mutil times op ", func() {
			lhsVector2 := []Sample{
				{
					Point: Point{
						T: 1652320542000,
						V: 1,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
			}

			rhsVector2 := []Sample{
				{
					Point: Point{
						T: 1652320542000,
						V: 2,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
			}

			expected := []Sample{
				{
					Point: Point{
						T: 0,
						V: 2,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			expected2 := []Sample{
				{
					Point: Point{
						T: 0,
						V: 2,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
				{
					Point: Point{
						T: 0,
						V: 2,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
			}
			expectVector2 := make(Vector, 0)
			expectVector2 = append(expectVector2, expected2...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorBinop(parser.MUL, lhsVector, rhsVector, &matcher, false, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)

			result2, err := VectorBinop(parser.MUL, lhsVector2, rhsVector2, &matcher, false, enh)
			So(result2, ShouldResemble, expectVector2)
			So(err, ShouldBeNil)

		})

		Convey("labels not match ", func() {
			lhsVector2 := []Sample{
				{
					Point: Point{
						T: 1652320542000,
						V: 1,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
			}

			rhsVector2 := []Sample{
				{
					Point: Point{
						T: 1652320542000,
						V: 2,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-2"},
					},
				},
			}
			expectVector2 := make(Vector, 0)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result2, err := VectorBinop(parser.MUL, lhsVector2, rhsVector2, &matcher, false, enh)
			So(result2, ShouldResemble, expectVector2)
			So(err, ShouldBeNil)

		})

		Convey("multiple matches for labels ", func() {
			lhsVector2 := []Sample{
				{
					Point: Point{
						T: 1652320542000,
						V: 1,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
				{
					Point: Point{
						T: 1652320542000,
						V: 1,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
			}

			rhsVector2 := []Sample{
				{
					Point: Point{
						T: 1652320542000,
						V: 2,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
			}

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}

			result2, err := VectorBinop(parser.MUL, lhsVector2, rhsVector2, &matcher, false, enh)
			So(result2, ShouldBeNil)
			So(err.Error(), ShouldEqual, "multiple matches for labels: many-to-one matching must be explicit (group_left/group_right)")

		})

		Convey("Vector op on() Vector is ADD ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 0,
						V: 3,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						// {Name: "node", Value: "node-1"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorBinop(parser.ADD, lhsVector, rhsVector, &onMatcher, false, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("Vector op Vector is EQLC ", func() {
			expectVector := make(Vector, 0)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorBinop(parser.EQLC, lhsVector, rhsVector, &matcher, false, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("Vector op Vector is NEQ ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 0,
						V: 1,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorBinop(parser.NEQ, lhsVector, rhsVector, &matcher, false, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("Vector op Vector is GTR ", func() {
			expectVector := make(Vector, 0)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorBinop(parser.GTR, lhsVector, rhsVector, &matcher, false, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("Vector op Vector is LSS ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 0,
						V: 1,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorBinop(parser.LSS, lhsVector, rhsVector, &matcher, false, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("Vector op Vector is GTE ", func() {
			expectVector := make(Vector, 0)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorBinop(parser.GTE, lhsVector, rhsVector, &matcher, false, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("Vector op Vector is LTE ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 0,
						V: 1,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorBinop(parser.LTE, lhsVector, rhsVector, &matcher, false, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("Vector op Vector is EQLC BOOL ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 0,
						V: 0,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorBinop(parser.EQLC, lhsVector, rhsVector, &matcher, true, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("Vector op Vector is NEQ BOOL ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 0,
						V: 1,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorBinop(parser.NEQ, lhsVector, rhsVector, &matcher, true, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("Vector op Vector is GTR BOOL ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 0,
						V: 0,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorBinop(parser.GTR, lhsVector, rhsVector, &matcher, true, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("Vector op Vector is LSS BOOL ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 0,
						V: 1,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorBinop(parser.LSS, lhsVector, rhsVector, &matcher, true, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("Vector op Vector is GTE BOOL ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 0,
						V: 0,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorBinop(parser.GTE, lhsVector, rhsVector, &matcher, true, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

		Convey("Vector op Vector is LTE BOOL ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 0,
						V: 1,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := VectorBinop(parser.LTE, lhsVector, rhsVector, &matcher, true, enh)
			So(result, ShouldResemble, expectVector)
			So(err, ShouldBeNil)
		})

	})
}

func TestProcessOutJoin(t *testing.T) {
	Convey("test processOutJoin", t, func() {
		leftVector := []Sample{
			{
				Point: Point{
					T: 1652320539000,
					V: 1,
				},
				Metric: []*labels.Label{
					{Name: "node", Value: "node-1"},
					{Name: "cpu", Value: "0"},
					{Name: "mode", Value: "system"},
				},
			},
		}

		rightVector := []Sample{
			{
				Point: Point{
					T: 1652320539000,
					V: 2,
				},
				Metric: []*labels.Label{
					{Name: "node", Value: "node-1"},
					{Name: "cpu", Value: "0"},
					{Name: "mode", Value: "system"},
				},
			},
			{
				Point: Point{
					T: 1652320539000,
					V: 2,
				},
				Metric: []*labels.Label{
					{Name: "node", Value: "node-1"},
					{Name: "cpu", Value: "1"},
					{Name: "mode", Value: "system"},
				},
			},
		}

		Convey("one to many without select left or select right ", func() {
			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			outJoinMatcher := parser.VectorMatching{Card: parser.OutJoinManyToMany,
				MatchingLabels: []string{"node"}, OutJoin: true}
			result, err := processOutJoin(parser.ADD, leftVector, rightVector, &outJoinMatcher, enh)
			So(err.Error(), ShouldEqual, "out_join match result cannot contain metrics with the same labelset")
			So(result, ShouldBeNil)
		})

		Convey("one to many with select right, result is deplicate series ", func() {
			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			outJoinMatcher := parser.VectorMatching{Card: parser.OutJoinManyToMany,
				MatchingLabels: []string{"node"}, OutJoin: true, IncludeRight: []string{"mode"}}
			result, err := processOutJoin(parser.ADD, leftVector, rightVector, &outJoinMatcher, enh)
			So(err.Error(), ShouldEqual, "vector cannot contain metrics with the same labelset")
			So(result, ShouldBeNil)
		})

		Convey("one to many with select right, result is unique series ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 1652320539000,
						V: 3,
					},
					Metric: []*labels.Label{
						{Name: "cpu", Value: "0"},
						{Name: "node", Value: "node-1"},
					},
				},
				{
					Point: Point{
						T: 1652320539000,
						V: 3,
					},
					Metric: []*labels.Label{
						{Name: "cpu", Value: "1"},
						{Name: "node", Value: "node-1"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			outJoinMatcher := parser.VectorMatching{Card: parser.OutJoinManyToMany,
				MatchingLabels: []string{"node"}, OutJoin: true, IncludeRight: []string{"cpu"}}
			result, err := processOutJoin(parser.ADD, leftVector, rightVector, &outJoinMatcher, enh)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectVector)
		})

		Convey("one to many with select left, result is deplicate series ", func() {
			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			outJoinMatcher := parser.VectorMatching{Card: parser.OutJoinManyToMany,
				MatchingLabels: []string{"node"}, OutJoin: true, IncludeLeft: []string{"mode"}}
			result, err := processOutJoin(parser.ADD, rightVector, leftVector, &outJoinMatcher, enh)
			So(err.Error(), ShouldEqual, "vector cannot contain metrics with the same labelset")
			So(result, ShouldBeNil)
		})

		Convey("one to many with select left, result is unique series ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 1652320539000,
						V: 3,
					},
					Metric: []*labels.Label{
						{Name: "cpu", Value: "0"},
						{Name: "node", Value: "node-1"},
					},
				},
				{
					Point: Point{
						T: 1652320539000,
						V: 3,
					},
					Metric: []*labels.Label{
						{Name: "cpu", Value: "1"},
						{Name: "node", Value: "node-1"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			outJoinMatcher := parser.VectorMatching{Card: parser.OutJoinManyToMany,
				MatchingLabels: []string{"node"}, OutJoin: true, IncludeLeft: []string{"cpu"}}
			result, err := processOutJoin(parser.ADD, rightVector, leftVector, &outJoinMatcher, enh)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectVector)
		})

		Convey("right is null", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 1652320539000,
						V: 1,
					},
					Metric: []*labels.Label{
						{Name: "cpu", Value: "0"},
						{Name: "node", Value: "node-1"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			outJoinMatcher := parser.VectorMatching{Card: parser.OutJoinManyToMany,
				MatchingLabels: []string{"node"}, OutJoin: true, IncludeLeft: []string{"cpu"}}
			result, err := processOutJoin(parser.ADD, leftVector, nil, &outJoinMatcher, enh)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectVector)
		})

		Convey("left is null", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 1652320539000,
						V: 1,
					},
					Metric: []*labels.Label{
						{Name: "node", Value: "node-1"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			outJoinMatcher := parser.VectorMatching{Card: parser.OutJoinManyToMany,
				MatchingLabels: []string{"node"}, OutJoin: true, IncludeLeft: []string{"cpu"}}
			result, err := processOutJoin(parser.ADD, nil, leftVector, &outJoinMatcher, enh)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectVector)
		})

		Convey("one to many with select left, select right result is unique series ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 1652320539000,
						V: 3,
					},
					Metric: []*labels.Label{
						{Name: "cpu", Value: "0"},
						{Name: "mode", Value: "system"},
						{Name: "node", Value: "node-1"},
					},
				},
				{
					Point: Point{
						T: 1652320539000,
						V: 3,
					},
					Metric: []*labels.Label{
						{Name: "cpu", Value: "1"},
						{Name: "mode", Value: "system"},
						{Name: "node", Value: "node-1"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			outJoinMatcher := parser.VectorMatching{Card: parser.OutJoinManyToMany,
				MatchingLabels: []string{"node"}, OutJoin: true, IncludeLeft: []string{"mode"}, IncludeRight: []string{"cpu"}}
			result, err := processOutJoin(parser.ADD, leftVector, rightVector, &outJoinMatcher, enh)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectVector)
		})

		Convey("left not match right, remain left ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 1652320539000,
						V: 3,
					},
					Metric: []*labels.Label{
						{Name: "cpu", Value: "0"},
						{Name: "mode", Value: "system"},
						{Name: "node", Value: "node-1"},
					},
				},
				{
					Point: Point{
						T: 1652320539000,
						V: 2,
					},
					Metric: []*labels.Label{
						{Name: "cpu", Value: "1"},
						{Name: "mode", Value: "system"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			outJoinMatcher := parser.VectorMatching{Card: parser.OutJoinManyToMany,
				MatchingLabels: []string{"cpu"}, OutJoin: true, IncludeLeft: []string{"node"}, IncludeRight: []string{"mode"}}
			result, err := processOutJoin(parser.ADD, leftVector, rightVector, &outJoinMatcher, enh)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectVector)
		})

	})
}

func TestPreprocessAndWrapWithStepInvariantExpr(t *testing.T) {

	startTime := time.Unix(1000, 0)
	endTime := time.Unix(9999, 0)
	var testCases = []struct {
		input    string      // The input to be parsed.
		expected parser.Expr // The expected expression AST.
	}{
		{
			input: "123.4567",
			expected: &parser.StepInvariantExpr{
				Expr: &parser.NumberLiteral{
					Val:      123.4567,
					PosRange: parser.PositionRange{Start: 0, End: 8},
				},
			},
		}, {
			input: "foo * bar",
			expected: &parser.BinaryExpr{
				Op: parser.MUL,
				LHS: &parser.VectorSelector{
					Name: "foo",
					LabelMatchers: []*labels.Matcher{
						parser.MustLabelMatcher(labels.MatchEqual, "__name__", "foo"),
					},
					PosRange: parser.PositionRange{
						Start: 0,
						End:   3,
					},
				},
				RHS: &parser.VectorSelector{
					Name: "bar",
					LabelMatchers: []*labels.Matcher{
						parser.MustLabelMatcher(labels.MatchEqual, "__name__", "bar"),
					},
					PosRange: parser.PositionRange{
						Start: 6,
						End:   9,
					},
				},
				VectorMatching: &parser.VectorMatching{Card: parser.CardOneToOne},
			},
		}, {
			input: "2 * foo",
			expected: &parser.BinaryExpr{
				Op: parser.MUL,
				LHS: &parser.StepInvariantExpr{
					Expr: &parser.NumberLiteral{
						Val:      2,
						PosRange: parser.PositionRange{Start: 0, End: 1},
					},
				},
				RHS: &parser.VectorSelector{
					Name: "foo",
					LabelMatchers: []*labels.Matcher{
						parser.MustLabelMatcher(labels.MatchEqual, "__name__", "foo"),
					},
					PosRange: parser.PositionRange{
						Start: 4,
						End:   7,
					},
				},
			},
		}, {
			input: "(2 + foo)",
			expected: &parser.ParenExpr{
				Expr: &parser.BinaryExpr{
					Op: parser.ADD,
					LHS: &parser.StepInvariantExpr{
						Expr: &parser.NumberLiteral{
							Val:      2,
							PosRange: parser.PositionRange{Start: 1, End: 2},
						},
					},
					RHS: &parser.VectorSelector{
						Name: "foo",
						LabelMatchers: []*labels.Matcher{
							parser.MustLabelMatcher(labels.MatchEqual, "__name__", "foo"),
						},
						PosRange: parser.PositionRange{
							Start: 5,
							End:   8,
						},
					},
				},
				PosRange: parser.PositionRange{
					Start: 0,
					End:   9,
				},
			},
		}, {
			input: "(-1)",
			expected: &parser.StepInvariantExpr{
				Expr: &parser.NumberLiteral{
					Val:      -1,
					PosRange: parser.PositionRange{Start: 1, End: 3},
				},
			},
		}, {
			input: "-foo",
			expected: &parser.UnaryExpr{
				Op: parser.SUB,
				Expr: &parser.VectorSelector{
					Name: "foo",
					LabelMatchers: []*labels.Matcher{
						parser.MustLabelMatcher(labels.MatchEqual, "__name__", "foo"),
					},
					PosRange: parser.PositionRange{
						Start: 1,
						End:   4,
					},
				},
			},
		}, {
			input: "foo * bar @ 10",
			expected: &parser.BinaryExpr{
				Op: parser.MUL,
				LHS: &parser.VectorSelector{
					Name: "foo",
					LabelMatchers: []*labels.Matcher{
						parser.MustLabelMatcher(labels.MatchEqual, "__name__", "foo"),
					},
					PosRange: parser.PositionRange{
						Start: 0,
						End:   3,
					},
				},
				RHS: &parser.StepInvariantExpr{
					Expr: &parser.VectorSelector{
						Name: "bar",
						LabelMatchers: []*labels.Matcher{
							parser.MustLabelMatcher(labels.MatchEqual, "__name__", "bar"),
						},
						PosRange: parser.PositionRange{
							Start: 6,
							End:   14,
						},
						Timestamp: makeInt64Pointer(10000),
					},
				},
				VectorMatching: &parser.VectorMatching{Card: parser.CardOneToOne},
			},
		}, {
			input: "foo @ 20 * bar @ 10",
			expected: &parser.StepInvariantExpr{
				Expr: &parser.BinaryExpr{
					Op: parser.MUL,
					LHS: &parser.VectorSelector{
						Name: "foo",
						LabelMatchers: []*labels.Matcher{
							parser.MustLabelMatcher(labels.MatchEqual, "__name__", "foo"),
						},
						PosRange: parser.PositionRange{
							Start: 0,
							End:   8,
						},
						Timestamp: makeInt64Pointer(20000),
					},
					RHS: &parser.VectorSelector{
						Name: "bar",
						LabelMatchers: []*labels.Matcher{
							parser.MustLabelMatcher(labels.MatchEqual, "__name__", "bar"),
						},
						PosRange: parser.PositionRange{
							Start: 11,
							End:   19,
						},
						Timestamp: makeInt64Pointer(10000),
					},
					VectorMatching: &parser.VectorMatching{Card: parser.CardOneToOne},
				},
			},
		}, {
			input: `foo @ start()`,
			expected: &parser.StepInvariantExpr{
				Expr: &parser.VectorSelector{
					Name: "foo",
					LabelMatchers: []*labels.Matcher{
						parser.MustLabelMatcher(labels.MatchEqual, "__name__", "foo"),
					},
					PosRange: parser.PositionRange{
						Start: 0,
						End:   13,
					},
					Timestamp:  makeInt64Pointer(startTime.UnixMilli()),
					StartOrEnd: parser.START,
				},
			},
		}, {
			input: `foo @ end()`,
			expected: &parser.StepInvariantExpr{
				Expr: &parser.VectorSelector{
					Name: "foo",
					LabelMatchers: []*labels.Matcher{
						parser.MustLabelMatcher(labels.MatchEqual, "__name__", "foo"),
					},
					PosRange: parser.PositionRange{
						Start: 0,
						End:   11,
					},
					Timestamp:  makeInt64Pointer(endTime.UnixMilli()),
					StartOrEnd: parser.END,
				},
			},
		}, {
			input: `time()`,
			expected: &parser.Call{
				Func: &parser.Function{
					Name:       "time",
					ArgTypes:   []parser.ValueType{},
					ReturnType: parser.ValueTypeScalar,
				},
				Args: parser.Expressions{},
				PosRange: parser.PositionRange{
					Start: 0,
					End:   6,
				},
			},
		}, {
			input: "sum by (foo)(some_metric)",
			expected: &parser.AggregateExpr{
				Op: parser.SUM,
				Expr: &parser.VectorSelector{
					Name: "some_metric",
					LabelMatchers: []*labels.Matcher{
						parser.MustLabelMatcher(labels.MatchEqual, "__name__", "some_metric"),
					},
					PosRange: parser.PositionRange{
						Start: 13,
						End:   24,
					},
				},
				Grouping: []string{"foo"},
				PosRange: parser.PositionRange{
					Start: 0,
					End:   25,
				},
			},
		}, {
			input: "sum by (foo)(some_metric @ 10)",
			expected: &parser.StepInvariantExpr{
				Expr: &parser.AggregateExpr{
					Op: parser.SUM,
					Expr: &parser.VectorSelector{
						Name: "some_metric",
						LabelMatchers: []*labels.Matcher{
							parser.MustLabelMatcher(labels.MatchEqual, "__name__", "some_metric"),
						},
						PosRange: parser.PositionRange{
							Start: 13,
							End:   29,
						},
						Timestamp: makeInt64Pointer(10000),
					},
					Grouping: []string{"foo"},
					PosRange: parser.PositionRange{
						Start: 0,
						End:   30,
					},
				},
			},
		}, {
			input: "sum(some_metric1 @ 10) + sum(some_metric2 @ 20)",
			expected: &parser.StepInvariantExpr{
				Expr: &parser.BinaryExpr{
					Op:             parser.ADD,
					VectorMatching: &parser.VectorMatching{},
					LHS: &parser.AggregateExpr{
						Op: parser.SUM,
						Expr: &parser.VectorSelector{
							Name: "some_metric1",
							LabelMatchers: []*labels.Matcher{
								parser.MustLabelMatcher(labels.MatchEqual, "__name__", "some_metric1"),
							},
							PosRange: parser.PositionRange{
								Start: 4,
								End:   21,
							},
							Timestamp: makeInt64Pointer(10000),
						},
						PosRange: parser.PositionRange{
							Start: 0,
							End:   22,
						},
					},
					RHS: &parser.AggregateExpr{
						Op: parser.SUM,
						Expr: &parser.VectorSelector{
							Name: "some_metric2",
							LabelMatchers: []*labels.Matcher{
								parser.MustLabelMatcher(labels.MatchEqual, "__name__", "some_metric2"),
							},
							PosRange: parser.PositionRange{
								Start: 29,
								End:   46,
							},
							Timestamp: makeInt64Pointer(20000),
						},
						PosRange: parser.PositionRange{
							Start: 25,
							End:   47,
						},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.input, func(t *testing.T) {
			expr, err := parser.ParseExpr(testCtx, test.input)
			require.NoError(t, err)
			expr, _ = PreprocessExpr(expr, startTime.UnixMilli(), endTime.UnixMilli())
			require.Equal(t, test.expected, expr, "error on input '%s'", test.input)
		})
	}
}

func TestGenerateGroupingKey(t *testing.T) {
	Convey("Test GenerateGroupingKey", t, func() {
		Convey("without is true", func() {
			metrics := []*labels.Label{{Name: "prometheus.labels.cluster", Value: "opensearch"}}
			hash, buf := GenerateGroupingKey(metrics, nil, true, nil)
			So(hash, ShouldEqual, uint64(10122032242877022144))
			So(buf, ShouldResemble, []byte{112, 114, 111, 109, 101, 116, 104, 101, 117, 115, 46, 108, 97, 98, 101, 108, 115, 46, 99, 108, 117, 115, 116, 101, 114, 255, 111, 112, 101, 110, 115, 101, 97, 114, 99, 104, 255})
		})

		Convey("without is false, and grouping is nil", func() {
			metrics := []*labels.Label{{Name: "prometheus.labels.cluster", Value: "opensearch"}}
			hash, buf := GenerateGroupingKey(metrics, nil, false, nil)
			So(hash, ShouldEqual, uint64(0))
			So(buf, ShouldBeNil)
		})

		Convey("without is false, and grouping is not nil", func() {
			metrics := []*labels.Label{{Name: "prometheus.labels.cluster", Value: "opensearch"}}
			hash, buf := GenerateGroupingKey(metrics, []string{"prometheus.labels.cluster"}, false, nil)
			So(hash, ShouldEqual, uint64(10122032242877022144))
			So(buf, ShouldResemble, []byte{112, 114, 111, 109, 101, 116, 104, 101, 117, 115, 46, 108, 97, 98, 101, 108, 115, 46, 99, 108, 117, 115, 116, 101, 114, 255, 111, 112, 101, 110, 115, 101, 97, 114, 99, 104, 255})
		})
	})
}

func TestParseMatchersParam(t *testing.T) {
	Convey("test ParseMatchersParam ", t, func() {

		Convey("match[] missing {}", func() {
			expectMatcher := [][]*labels.Matcher{
				{
					{
						Type:  labels.MatchEqual,
						Name:  "__name__",
						Value: "aa",
					},
				},
			}
			matcherSets, err := ParseMatchersParam([]string{"aa"})
			So(err, ShouldBeNil)
			So(matcherSets, ShouldResemble, expectMatcher)
		})

		Convey("match[] is aa{a='a'}abc ", func() {

			matcherSets, err := ParseMatchersParam([]string{"aa{a='a'}abc"})
			So(err.Error(), ShouldEqual, `1:10: parse error: unexpected identifier "abc"`)
			So(matcherSets, ShouldBeNil)
		})

		Convey("one matchers ", func() {
			expectMatcher := [][]*labels.Matcher{
				{
					{
						Type:  labels.MatchEqual,
						Name:  "__name__",
						Value: "aa",
					},
				},
			}
			matcherSets, err := ParseMatchersParam([]string{"aa{}"})
			So(err, ShouldBeNil)
			So(matcherSets, ShouldResemble, expectMatcher)
		})

		Convey("one matchers with . ", func() {
			expectMatcher := [][]*labels.Matcher{
				{

					{
						Type:  labels.MatchEqual,
						Name:  "labels.instance",
						Value: "a.b.c",
					},
					{
						Type:  labels.MatchEqual,
						Name:  "__name__",
						Value: "aa",
					},
				},
			}
			matcherSets, err := ParseMatchersParam([]string{`aa{labels.instance="a.b.c"}`})
			So(err, ShouldBeNil)
			So(matcherSets, ShouldResemble, expectMatcher)
		})

		Convey("one matchers with \\. ", func() {
			expectMatcher := [][]*labels.Matcher{
				{

					{
						Type:  labels.MatchEqual,
						Name:  "labels.instance",
						Value: "a\\.b\\.c",
					},
					{
						Type:  labels.MatchEqual,
						Name:  "__name__",
						Value: "aa",
					},
				},
			}
			matcherSets, err := ParseMatchersParam([]string{`aa{labels.instance="a\\.b\\.c"}`})
			So(err, ShouldBeNil)
			So(matcherSets, ShouldResemble, expectMatcher)
		})

		Convey("more than one matchers ", func() {
			expectMatcher := [][]*labels.Matcher{
				{
					{
						Type:  labels.MatchEqual,
						Name:  "__name__",
						Value: "aa",
					},
				},
				{
					{
						Type:  labels.MatchEqual,
						Name:  "labels.mode",
						Value: "nice",
					},
					{
						Type:  labels.MatchEqual,
						Name:  "__name__",
						Value: "prometheus.metrics.node_cpu_guest_seconds_total",
					},
				},
			}
			matcherSets, err := ParseMatchersParam([]string{"aa{}",
				`prometheus.metrics.node_cpu_guest_seconds_total{labels.mode="nice"}`})
			So(err, ShouldBeNil)
			So(matcherSets, ShouldResemble, expectMatcher)
		})

		Convey("matchers is empty ", func() {
			matcherSets, err := ParseMatchersParam([]string{""})
			So(err.Error(), ShouldEqual, `1:1: parse error: unexpected end of input`)
			So(matcherSets, ShouldBeNil)
		})

		Convey("matchers is {} ", func() {
			matcherSets, err := ParseMatchersParam([]string{"{}"})
			So(err.Error(), ShouldEqual, `match[] must contain at least one non-empty matcher`)
			So(matcherSets, ShouldBeNil)
		})

		Convey("matchers is {a=\"\"} ", func() {
			matcherSets, err := ParseMatchersParam([]string{`{a=""}`})
			So(err.Error(), ShouldEqual, `match[] must contain at least one non-empty matcher`)
			So(matcherSets, ShouldBeNil)
		})

	})
}

var (
	aggVec = []Sample{
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
				{Name: "cluster", Value: "opensearch"},
				{Name: "node", Value: "node-2"},
			},
		},
	}

	groupingKeyWithNoGroup1, _ = GenerateGroupingKey([]*labels.Label{
		{Name: "cluster", Value: "opensearch"},
		{Name: "node", Value: "node-1"},
	}, []string{}, false, make([]byte, 0, 1024))

	groupingKeyWithNoGroup2, _ = GenerateGroupingKey([]*labels.Label{
		{Name: "cluster", Value: "opensearch"},
		{Name: "node", Value: "node-2"},
	}, []string{}, false, make([]byte, 0, 1024))

	aggSeriesHelper = []EvalSeriesHelper{
		{
			GroupingKey: groupingKeyWithNoGroup1,
			Signature:   "",
		},
		{
			GroupingKey: groupingKeyWithNoGroup2,
			Signature:   "",
		},
	}

	groupingKeyNode1, _ = GenerateGroupingKey([]*labels.Label{
		{Name: "cluster", Value: "opensearch"},
		{Name: "node", Value: "node-1"},
	}, []string{"node"}, false, make([]byte, 0, 1024))

	groupingKeyNode2, _ = GenerateGroupingKey([]*labels.Label{
		{Name: "cluster", Value: "opensearch"},
		{Name: "node", Value: "node-2"},
	}, []string{"node"}, false, make([]byte, 0, 1024))

	aggSeriesHelperWithGroupByNode = []EvalSeriesHelper{
		{
			GroupingKey: groupingKeyNode1,
			Signature:   "",
		},
		{
			GroupingKey: groupingKeyNode2,
			Signature:   "",
		},
	}

	groupingKeyCluster1, _ = GenerateGroupingKey([]*labels.Label{
		{Name: "cluster", Value: "opensearch"},
		{Name: "node", Value: "node-1"},
	}, []string{"cluster"}, false, make([]byte, 0, 1024))

	groupingKeyCluster2, _ = GenerateGroupingKey([]*labels.Label{
		{Name: "cluster", Value: "opensearch"},
		{Name: "node", Value: "node-2"},
	}, []string{"cluster"}, false, make([]byte, 0, 1024))

	aggSeriesHelperWithGroupByCluster = []EvalSeriesHelper{
		{
			GroupingKey: groupingKeyCluster1,
			Signature:   "",
		},
		{
			GroupingKey: groupingKeyCluster2,
			Signature:   "",
		},
	}
)

func TestAggregation(t *testing.T) {
	Convey("test Aggregation", t, func() {

		Convey("Aggregation topk param gt 9223372036854774784", func() {

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := Aggregation(parser.TOPK, []string{}, false, float64(9223372036854774785000), aggVec, aggSeriesHelper, enh)
			So(err.Error(), ShouldEqual, "Scalar value 9.223372036854775e+21 overflows int64")
			So(result, ShouldBeNil)
		})

		Convey("Aggregation topk param lt 1", func() {
			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := Aggregation(parser.TOPK, []string{}, false, float64(0.5), aggVec, aggSeriesHelper, enh)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, Vector{})
		})

		Convey("Aggregation topk 1 ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 0,
						V: 2,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-2"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := Aggregation(parser.TOPK, []string{}, false, float64(1), aggVec, aggSeriesHelper, enh)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectVector)
		})

		Convey("Aggregation topk 10 more than 2 ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 0,
						V: 2,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-2"},
					},
				},
				{
					Point: Point{
						T: 0,
						V: 1,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := Aggregation(parser.TOPK, []string{}, false, float64(10), aggVec, aggSeriesHelper, enh)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectVector)
		})

		Convey("Aggregation topk 1 by cluster ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 0,
						V: 2,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-2"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}

			result, err := Aggregation(parser.TOPK, []string{"cluster"}, false, float64(1), aggVec, aggSeriesHelperWithGroupByCluster, enh)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectVector)
		})

		Convey("Aggregation topk 1 by node ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 0,
						V: 1,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
				{
					Point: Point{
						T: 0,
						V: 2,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-2"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}

			result, err := Aggregation(parser.TOPK, []string{"node"}, false, float64(1), aggVec, aggSeriesHelperWithGroupByNode, enh)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectVector)
		})

		Convey("Aggregation bottomk param gt 9223372036854774784", func() {

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := Aggregation(parser.BOTTOMK, []string{}, false, float64(9223372036854774785000), aggVec, aggSeriesHelper, enh)
			So(err.Error(), ShouldEqual, "Scalar value 9.223372036854775e+21 overflows int64")
			So(result, ShouldBeNil)
		})

		Convey("Aggregation bottomk param lt 1", func() {
			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := Aggregation(parser.BOTTOMK, []string{}, false, float64(0.5), aggVec, aggSeriesHelper, enh)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, Vector{})
		})

		Convey("Aggregation bottomk 1 ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 0,
						V: 1,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-2"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			aggVecBottomK := []Sample{
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
				{
					Point: Point{
						T: 1652320539000,
						V: 1,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-2"},
					},
				},
			}

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			result, err := Aggregation(parser.BOTTOMK, []string{}, false, float64(1), aggVecBottomK, aggSeriesHelper, enh)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectVector)
		})

		Convey("Aggregation bottomk 1 by cluster ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 0,
						V: 1,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}

			result, err := Aggregation(parser.BOTTOMK, []string{"cluster"}, false, float64(1), aggVec, aggSeriesHelperWithGroupByCluster, enh)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectVector)
		})

		Convey("Aggregation bottomk 1 by node ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 0,
						V: 1,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
				{
					Point: Point{
						T: 0,
						V: 2,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-2"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}

			result, err := Aggregation(parser.BOTTOMK, []string{"node"}, false, float64(1), aggVec, aggSeriesHelperWithGroupByNode, enh)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectVector)
		})

		Convey("Aggregation sum no by ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 0,
						V: 3,
					},
					Metric: []*labels.Label{},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}

			result, err := Aggregation(parser.SUM, []string{}, false, float64(1), aggVec, aggSeriesHelper, enh)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectVector)
		})

		Convey("Aggregation sum by cluster ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 0,
						V: 3,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}

			result, err := Aggregation(parser.SUM, []string{"cluster"}, false, float64(1), aggVec, aggSeriesHelperWithGroupByCluster, enh)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectVector)
		})

		Convey("Aggregation sum by node ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 0,
						V: 1,
					},
					Metric: []*labels.Label{
						{Name: "node", Value: "node-1"},
					},
				},
				{
					Point: Point{
						T: 0,
						V: 2,
					},
					Metric: []*labels.Label{
						{Name: "node", Value: "node-2"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}

			result, err := Aggregation(parser.SUM, []string{"node"}, false, float64(1), aggVec, aggSeriesHelperWithGroupByNode, enh)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectVector)
		})

		Convey("Aggregation avg no by ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 0,
						V: 1.5,
					},
					Metric: []*labels.Label{},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}

			result, err := Aggregation(parser.AVG, []string{}, false, float64(1), aggVec, aggSeriesHelper, enh)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectVector)
		})

		Convey("Aggregation avg by cluster ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 0,
						V: 1.5,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}

			result, err := Aggregation(parser.AVG, []string{"cluster"}, false, float64(1), aggVec, aggSeriesHelperWithGroupByCluster, enh)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectVector)
		})

		Convey("Aggregation avg by node ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 0,
						V: 1,
					},
					Metric: []*labels.Label{
						{Name: "node", Value: "node-1"},
					},
				},
				{
					Point: Point{
						T: 0,
						V: 2,
					},
					Metric: []*labels.Label{
						{Name: "node", Value: "node-2"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}

			result, err := Aggregation(parser.AVG, []string{"node"}, false, float64(1), aggVec, aggSeriesHelperWithGroupByNode, enh)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectVector)
		})

		Convey("Aggregation avg no by and value all is inf ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 0,
						V: math.Inf(1),
					},
					Metric: []*labels.Label{},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			aggVec = []Sample{
				{
					Point: Point{
						T: 1652320539000,
						V: math.Inf(1),
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
				{
					Point: Point{
						T: 1652320539000,
						V: math.Inf(1),
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-2"},
					},
				},
			}
			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}

			result, err := Aggregation(parser.AVG, []string{}, false, float64(1), aggVec, aggSeriesHelper, enh)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectVector)
		})

		Convey("Aggregation avg no by and value contain inf ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 0,
						V: math.Inf(1),
					},
					Metric: []*labels.Label{},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			aggVec = []Sample{
				{
					Point: Point{
						T: 1652320539000,
						V: math.Inf(1),
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-1"},
					},
				},
				{
					Point: Point{
						T: 1652320539000,
						V: 1,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
						{Name: "node", Value: "node-2"},
					},
				},
			}
			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}

			result, err := Aggregation(parser.AVG, []string{}, false, float64(1), aggVec, aggSeriesHelper, enh)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectVector)
		})

		Convey("Aggregation count no by ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 0,
						V: 2,
					},
					Metric: []*labels.Label{},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}

			result, err := Aggregation(parser.COUNT, []string{}, false, float64(1), aggVec, aggSeriesHelper, enh)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectVector)
		})

		Convey("Aggregation count by cluster ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 0,
						V: 2,
					},
					Metric: []*labels.Label{
						{Name: "cluster", Value: "opensearch"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}

			result, err := Aggregation(parser.COUNT, []string{"cluster"}, false, float64(1), aggVec, aggSeriesHelperWithGroupByCluster, enh)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectVector)
		})

		Convey("Aggregation count by node ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 0,
						V: 1,
					},
					Metric: []*labels.Label{
						{Name: "node", Value: "node-1"},
					},
				},
				{
					Point: Point{
						T: 0,
						V: 1,
					},
					Metric: []*labels.Label{
						{Name: "node", Value: "node-2"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}

			result, err := Aggregation(parser.COUNT, []string{"node"}, false, float64(1), aggVec, aggSeriesHelperWithGroupByNode, enh)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectVector)
		})

		Convey("Aggregation count without node ", func() {
			expected := []Sample{
				{
					Point: Point{
						T: 0,
						V: 1,
					},
					Metric: []*labels.Label{
						{Name: "node", Value: "node-1"},
					},
				},
				{
					Point: Point{
						T: 0,
						V: 1,
					},
					Metric: []*labels.Label{
						{Name: "node", Value: "node-2"},
					},
				},
			}
			expectVector := make(Vector, 0)
			expectVector = append(expectVector, expected...)

			enh := &EvalNodeHelper{Out: make(Vector, 0, 2)}
			groupingKeyCluster1, _ = GenerateGroupingKey([]*labels.Label{
				{Name: "cluster", Value: "opensearch"},
				{Name: "node", Value: "node-1"},
			}, []string{"cluster"}, true, make([]byte, 0, 1024))

			groupingKeyCluster2, _ = GenerateGroupingKey([]*labels.Label{
				{Name: "cluster", Value: "opensearch"},
				{Name: "node", Value: "node-2"},
			}, []string{"cluster"}, true, make([]byte, 0, 1024))

			aggSeriesHelperWithGroupByClusterTmp := []EvalSeriesHelper{
				{
					GroupingKey: groupingKeyCluster1,
					Signature:   "",
				},
				{
					GroupingKey: groupingKeyCluster2,
					Signature:   "",
				},
			}

			result, err := Aggregation(parser.COUNT, []string{"cluster"}, true, nil, aggVec, aggSeriesHelperWithGroupByClusterTmp, enh)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectVector)
		})
	})
}

func TestBtos(t *testing.T) {
	Convey("test btos", t, func() {

		Convey("btos para is true ", func() {
			result := btos(true)
			So(result, ShouldEqual, 1)
		})

		Convey("btos para is false ", func() {
			result := btos(false)
			So(result, ShouldEqual, 0)
		})
	})
}

func TestMatchLable(t *testing.T) {

	Convey("Test MatchLable", t, func() {

		leftMetricMap := map[string]string{
			"cluster":   "opensearch",
			"node":      "node-1",
			"pod":       "node-exporter-tdhtm",
			"namespace": "kube-system",
		}

		rightMetricMap := map[string]string{
			"cluster":   "opensearch",
			"job":       "kubernetes-services-endpoints",
			"cpu":       "0",
			"namespace": "kube-system",
			"mode":      "iqr",
		}

		Convey("matching successful", func() {
			matching := parser.VectorMatching{
				Card:           parser.CardManyToOne,
				MatchingLabels: []string{"cluster"},
				On:             true,
				Include:        []string{"job"},
			}

			match := matchLable(leftMetricMap, rightMetricMap, &matching)
			So(match, ShouldBeTrue)

		})

		Convey("matching fail", func() {
			matching := parser.VectorMatching{
				Card:           parser.CardManyToOne,
				MatchingLabels: []string{"cluster", "node", "aaa"},
				On:             true,
				Include:        []string{"job"},
			}
			match := matchLable(leftMetricMap, rightMetricMap, &matching)
			So(match, ShouldBeFalse)

		})

		Convey("MatchingLabels is empty", func() {
			matching := parser.VectorMatching{
				Card:           parser.CardManyToOne,
				MatchingLabels: []string{},
				On:             true,
				Include:        []string{"job"},
			}
			match := matchLable(leftMetricMap, rightMetricMap, &matching)
			So(match, ShouldBeFalse)

		})
	})
}

func TestJoin(t *testing.T) {

	Convey("Test union", t, func() {
		leftSample := Sample{

			Point: Point{
				T: 1652320539000,
				V: 1,
			},
			Metric: []*labels.Label{
				{Name: "cluster", Value: "opensearch"},
				{Name: "node", Value: "node-1"},
			},
		}
		leftMetricMap := map[string]string{
			"cluster": "opensearch",
			"node":    "node-1",
		}

		rightSample := Sample{

			Point: Point{
				T: 1652320539000,
				V: 2,
			},
			Metric: []*labels.Label{
				{Name: "cluster", Value: "opensearch"},
				{Name: "job", Value: "kubernetes-services-endpoints"},
				{Name: "mode", Value: "nice"},
			},
		}

		rightMetricMap := map[string]string{
			"cluster": "opensearch",
			"job":     "kubernetes-services-endpoints",
			"mode":    "nice",
		}

		var op parser.ItemType = parser.ADD
		Convey("the parameter for group_left is empty", func() {
			matching := parser.VectorMatching{
				Card:           parser.CardManyToOne,
				MatchingLabels: []string{"cluster"},
				On:             true,
				Include:        []string{},
			}
			expectedSample := Sample{
				Point: Point{
					T: 1652320539000,
					V: 3,
				},
				Metric: []*labels.Label{
					{Name: "cluster", Value: "opensearch"},
					{Name: "node", Value: "node-1"},
				},
			}
			result := join(leftSample, rightSample, leftMetricMap, rightMetricMap, op, &matching)
			So(result, ShouldResemble, expectedSample)
		})

		Convey("the parameter of group_left is not empty", func() {
			matching := parser.VectorMatching{
				Card:           parser.CardManyToOne,
				MatchingLabels: []string{"cluster"},
				On:             true,
				Include:        []string{"job"},
			}

			expectedSample := Sample{
				Point: Point{
					T: 1652320539000,
					V: 3,
				},
				Metric: []*labels.Label{
					{Name: "cluster", Value: "opensearch"},
					{Name: "job", Value: "kubernetes-services-endpoints"},
					{Name: "node", Value: "node-1"},
				},
			}
			result := join(leftSample, rightSample, leftMetricMap, rightMetricMap, op, &matching)
			So(result, ShouldResemble, expectedSample)
		})
	})
}

func TestFillMissingPoint(t *testing.T) {

	Convey("Test FillMissingPoint", t, func() {
		inputSeries := []Series{
			{
				Metric: labels.Labels{&labels.Label{Name: "index", Value: "nas_statis-2022.05-0"}},
				Points: []Point{
					{T: 1652320200000, V: 20}, {T: 1652320260000, V: 20}, {T: 1652320320000, V: 20}, {T: 1652320380000, V: 20}, {T: 1652320440000, V: 20}, {T: 1652320500000, V: 20}, {T: 1652320560000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652321400000, V: 0},
				},
			},
			{
				Metric: labels.Labels{&labels.Label{Name: "index", Value: "node_statis-2022.05-0"}},
				Points: []Point{
					{T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 1}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 1}, {T: 1652320800000, V: 1}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652321400000, V: 0},
				},
			},
			{
				Metric: labels.Labels{&labels.Label{Name: "index", Value: "noee_statis-2022.05-0"}},
				Points: []Point{
					{T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652321400000, V: 0},
				},
			},
		}
		var inputMat Matrix
		inputMat = append(inputMat, inputSeries...)

		query := interfaces.Query{
			QueryStr:       "continuous_k_minute_downtime(5, -1, 0, a)",
			Start:          1652319900000,
			End:            1652321400000,
			Interval:       60000,
			FixedStart:     1652319900000,
			FixedEnd:       1652321400000,
			IsInstantQuery: false,
			IsMetricModel:  true,
			// LookBackDelta:  1200000,
			ModelId: "123456",
		}

		Convey("filling with precedingMissingPolicy=-1, middleMissingPolicy=0", func() {
			expectSeries := []Series{
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "nas_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 20}, {T: 1652320260000, V: 20}, {T: 1652320320000, V: 20}, {T: 1652320380000, V: 20}, {T: 1652320440000, V: 20}, {T: 1652320500000, V: 20}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "node_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 1}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 1}, {T: 1652320800000, V: 1}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "noee_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 0}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0},
					},
				},
			}
			var expectMat Matrix
			expectMat = append(expectMat, expectSeries...)

			result := FillMissingPoint(query, []Matrix{inputMat}, -1, 0)
			So(result, ShouldResemble, []Matrix{expectMat})
		})

		Convey("filling with precedingMissingPolicy=-1, middleMissingPolicy=-1", func() {
			expectSeries := []Series{
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "nas_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 20}, {T: 1652320260000, V: 20}, {T: 1652320320000, V: 20}, {T: 1652320380000, V: 20}, {T: 1652320440000, V: 20}, {T: 1652320500000, V: 20}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "node_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 1}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 1}, {T: 1652320800000, V: 1}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "noee_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 0}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0},
					},
				},
			}
			var expectMat Matrix
			expectMat = append(expectMat, expectSeries...)

			result := FillMissingPoint(query, []Matrix{inputMat}, -1, -1)
			So(result, ShouldResemble, []Matrix{expectMat})
		})

		Convey("filling with precedingMissingPolicy=-1, middleMissingPolicy=1", func() {
			expectSeries := []Series{
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "nas_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 20}, {T: 1652320260000, V: 20}, {T: 1652320320000, V: 20}, {T: 1652320380000, V: 20}, {T: 1652320440000, V: 20}, {T: 1652320500000, V: 20}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 1}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 1}, {T: 1652321040000, V: 1}, {T: 1652321100000, V: 1}, {T: 1652321160000, V: 1}, {T: 1652321220000, V: 1}, {T: 1652321280000, V: 1}, {T: 1652321340000, V: 1},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "node_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 1}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 1}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 1}, {T: 1652320800000, V: 1}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 1}, {T: 1652321040000, V: 1}, {T: 1652321100000, V: 1}, {T: 1652321160000, V: 1}, {T: 1652321220000, V: 1}, {T: 1652321280000, V: 1}, {T: 1652321340000, V: 1},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "noee_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 0}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 1}, {T: 1652321040000, V: 1}, {T: 1652321100000, V: 1}, {T: 1652321160000, V: 1}, {T: 1652321220000, V: 1}, {T: 1652321280000, V: 1}, {T: 1652321340000, V: 1},
					},
				},
			}
			var expectMat Matrix
			expectMat = append(expectMat, expectSeries...)

			result := FillMissingPoint(query, []Matrix{inputMat}, -1, 1)
			So(result, ShouldResemble, []Matrix{expectMat})
		})

		Convey("filling with precedingMissingPolicy=1, middleMissingPolicy=0", func() {
			expectSeries := []Series{
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "nas_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 1}, {T: 1652319960000, V: 1}, {T: 1652320020000, V: 1}, {T: 1652320080000, V: 1}, {T: 1652320140000, V: 1}, {T: 1652320200000, V: 20}, {T: 1652320260000, V: 20}, {T: 1652320320000, V: 20}, {T: 1652320380000, V: 20}, {T: 1652320440000, V: 20}, {T: 1652320500000, V: 20}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "node_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 1}, {T: 1652319960000, V: 1}, {T: 1652320020000, V: 1}, {T: 1652320080000, V: 1}, {T: 1652320140000, V: 1}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 1}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 1}, {T: 1652320800000, V: 1}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "noee_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 1}, {T: 1652319960000, V: 1}, {T: 1652320020000, V: 1}, {T: 1652320080000, V: 1}, {T: 1652320140000, V: 1}, {T: 1652320200000, V: 1}, {T: 1652320260000, V: 1}, {T: 1652320320000, V: 1}, {T: 1652320380000, V: 1}, {T: 1652320440000, V: 1}, {T: 1652320500000, V: 1}, {T: 1652320560000, V: 1}, {T: 1652320620000, V: 1}, {T: 1652320680000, V: 1}, {T: 1652320740000, V: 1}, {T: 1652320800000, V: 1}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0},
					},
				},
			}
			var expectMat Matrix
			expectMat = append(expectMat, expectSeries...)

			result := FillMissingPoint(query, []Matrix{inputMat}, 1, 0)
			So(result, ShouldResemble, []Matrix{expectMat})
		})

		Convey("filling with precedingMissingPolicy=1, middleMissingPolicy=-1", func() {
			expectSeries := []Series{
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "nas_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 1}, {T: 1652319960000, V: 1}, {T: 1652320020000, V: 1}, {T: 1652320080000, V: 1}, {T: 1652320140000, V: 1}, {T: 1652320200000, V: 20}, {T: 1652320260000, V: 20}, {T: 1652320320000, V: 20}, {T: 1652320380000, V: 20}, {T: 1652320440000, V: 20}, {T: 1652320500000, V: 20}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "node_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 1}, {T: 1652319960000, V: 1}, {T: 1652320020000, V: 1}, {T: 1652320080000, V: 1}, {T: 1652320140000, V: 1}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 1}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 1}, {T: 1652320800000, V: 1}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "noee_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 1}, {T: 1652319960000, V: 1}, {T: 1652320020000, V: 1}, {T: 1652320080000, V: 1}, {T: 1652320140000, V: 1}, {T: 1652320200000, V: 1}, {T: 1652320260000, V: 1}, {T: 1652320320000, V: 1}, {T: 1652320380000, V: 1}, {T: 1652320440000, V: 1}, {T: 1652320500000, V: 1}, {T: 1652320560000, V: 1}, {T: 1652320620000, V: 1}, {T: 1652320680000, V: 1}, {T: 1652320740000, V: 1}, {T: 1652320800000, V: 1}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0},
					},
				},
			}
			var expectMat Matrix
			expectMat = append(expectMat, expectSeries...)

			result := FillMissingPoint(query, []Matrix{inputMat}, 1, -1)
			So(result, ShouldResemble, []Matrix{expectMat})
		})

		Convey("filling with precedingMissingPolicy=1, middleMissingPolicy=1", func() {
			expectSeries := []Series{
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "nas_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 1}, {T: 1652319960000, V: 1}, {T: 1652320020000, V: 1}, {T: 1652320080000, V: 1}, {T: 1652320140000, V: 1}, {T: 1652320200000, V: 20}, {T: 1652320260000, V: 20}, {T: 1652320320000, V: 20}, {T: 1652320380000, V: 20}, {T: 1652320440000, V: 20}, {T: 1652320500000, V: 20}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "node_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 1}, {T: 1652319960000, V: 1}, {T: 1652320020000, V: 1}, {T: 1652320080000, V: 1}, {T: 1652320140000, V: 1}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 1}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 1}, {T: 1652320800000, V: 1}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "noee_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 1}, {T: 1652319960000, V: 1}, {T: 1652320020000, V: 1}, {T: 1652320080000, V: 1}, {T: 1652320140000, V: 1}, {T: 1652320200000, V: 1}, {T: 1652320260000, V: 1}, {T: 1652320320000, V: 1}, {T: 1652320380000, V: 1}, {T: 1652320440000, V: 1}, {T: 1652320500000, V: 1}, {T: 1652320560000, V: 1}, {T: 1652320620000, V: 1}, {T: 1652320680000, V: 1}, {T: 1652320740000, V: 1}, {T: 1652320800000, V: 1}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0},
					},
				},
			}
			var expectMat Matrix
			expectMat = append(expectMat, expectSeries...)

			result := FillMissingPoint(query, []Matrix{inputMat}, 1, 0)
			So(result, ShouldResemble, []Matrix{expectMat})
		})
	})
}

func TestCombineEvalUsability(t *testing.T) {

	Convey("Test CombineEvalUsability", t, func() {
		inputSeries := []Series{
			{
				Metric: labels.Labels{&labels.Label{Name: "index", Value: "nas_statis-2022.05-0"}},
				Points: []Point{
					{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 20}, {T: 1652320260000, V: 20}, {T: 1652320320000, V: 20}, {T: 1652320380000, V: 20}, {T: 1652320440000, V: 20}, {T: 1652320500000, V: 20}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
				},
			},
			{
				Metric: labels.Labels{&labels.Label{Name: "index", Value: "node_statis-2022.05-0"}},
				Points: []Point{
					{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 1}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 1}, {T: 1652320800000, V: 1}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
				},
			},
			{
				Metric: labels.Labels{&labels.Label{Name: "index", Value: "noee2_statis-2022.05-0"}},
				Points: []Point{
					{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 0}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
				},
			},
		}
		inputSeries2 := []Series{
			{
				Metric: labels.Labels{&labels.Label{Name: "index", Value: "nas1_statis-2022.05-0"}},
				Points: []Point{
					{T: 1652319900000, V: 1}, {T: 1652319960000, V: 1}, {T: 1652320020000, V: 1}, {T: 1652320080000, V: 1}, {T: 1652320140000, V: 1}, {T: 1652320200000, V: 20}, {T: 1652320260000, V: 20}, {T: 1652320320000, V: 20}, {T: 1652320380000, V: 20}, {T: 1652320440000, V: 20}, {T: 1652320500000, V: 20}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
				},
			},
			{
				Metric: labels.Labels{&labels.Label{Name: "index", Value: "node_statis-2022.05-0"}},
				Points: []Point{
					{T: 1652319900000, V: 1}, {T: 1652319960000, V: 1}, {T: 1652320020000, V: 1}, {T: 1652320080000, V: 1}, {T: 1652320140000, V: 1}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 1}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 1}, {T: 1652320800000, V: 1}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
				},
			},
			{
				Metric: labels.Labels{&labels.Label{Name: "index", Value: "noee_statis-2022.05-0"}},
				Points: []Point{
					{T: 1652319900000, V: 1}, {T: 1652319960000, V: 1}, {T: 1652320020000, V: 1}, {T: 1652320080000, V: 1}, {T: 1652320140000, V: 1}, {T: 1652320200000, V: 1}, {T: 1652320260000, V: 1}, {T: 1652320320000, V: 1}, {T: 1652320380000, V: 1}, {T: 1652320440000, V: 1}, {T: 1652320500000, V: 1}, {T: 1652320560000, V: 1}, {T: 1652320620000, V: 1}, {T: 1652320680000, V: 1}, {T: 1652320740000, V: 1}, {T: 1652320800000, V: 1}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
				},
			},
		}
		var inputMat Matrix
		inputMat = append(inputMat, inputSeries...)
		var inputMat2 Matrix
		inputMat2 = append(inputMat2, inputSeries2...)

		Convey("filling with precedingMissingPolicy=-1, middleMissingPolicy=0", func() {
			expectSeries := []Series{
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "nas_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 0}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 0}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "node_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 1}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 1}, {T: 1652320800000, V: 1}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "noee2_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 0}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 0}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "nas1_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 0}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 0}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "noee_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 0}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 0}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
					},
				},
			}
			var expectMat Matrix
			expectMat = append(expectMat, expectSeries...)

			result := CombineEvalUsability([]Matrix{inputMat, inputMat2}, -1, 0)
			So(result, ShouldResemble, expectMat)
		})

		Convey("filling with precedingMissingPolicy=-1, middleMissingPolicy=-1", func() {
			expectSeries := []Series{
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "nas_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 0}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 0}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "node_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 1}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 1}, {T: 1652320800000, V: 1}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "noee2_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 0}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 0}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "nas1_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 0}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 0}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "noee_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 0}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 0}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
					},
				},
			}
			var expectMat Matrix
			expectMat = append(expectMat, expectSeries...)

			result := CombineEvalUsability([]Matrix{inputMat, inputMat2}, -1, -1)
			So(result, ShouldResemble, expectMat)
		})

		Convey("filling with precedingMissingPolicy=-1, middleMissingPolicy=1", func() {
			expectSeries := []Series{
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "nas_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 0}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 0}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "node_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 1}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 1}, {T: 1652320800000, V: 1}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "noee2_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 0}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 0}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "nas1_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 0}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 0}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "noee_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 0}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 0}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
					},
				},
			}
			var expectMat Matrix
			expectMat = append(expectMat, expectSeries...)

			result := CombineEvalUsability([]Matrix{inputMat, inputMat2}, -1, 1)
			So(result, ShouldResemble, expectMat)
		})

		Convey("filling with precedingMissingPolicy=1, middleMissingPolicy=0", func() {
			expectSeries := []Series{
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "nas_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 1}, {T: 1652320260000, V: 1}, {T: 1652320320000, V: 1}, {T: 1652320380000, V: 1}, {T: 1652320440000, V: 1}, {T: 1652320500000, V: 1}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "node_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 1}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 1}, {T: 1652320800000, V: 1}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "noee2_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 0}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "nas1_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 1}, {T: 1652319960000, V: 1}, {T: 1652320020000, V: 1}, {T: 1652320080000, V: 1}, {T: 1652320140000, V: 1}, {T: 1652320200000, V: 1}, {T: 1652320260000, V: 1}, {T: 1652320320000, V: 1}, {T: 1652320380000, V: 1}, {T: 1652320440000, V: 1}, {T: 1652320500000, V: 1}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "noee_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 1}, {T: 1652319960000, V: 1}, {T: 1652320020000, V: 1}, {T: 1652320080000, V: 1}, {T: 1652320140000, V: 1}, {T: 1652320200000, V: 1}, {T: 1652320260000, V: 1}, {T: 1652320320000, V: 1}, {T: 1652320380000, V: 1}, {T: 1652320440000, V: 1}, {T: 1652320500000, V: 1}, {T: 1652320560000, V: 1}, {T: 1652320620000, V: 1}, {T: 1652320680000, V: 1}, {T: 1652320740000, V: 1}, {T: 1652320800000, V: 1}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
					},
				},
			}
			var expectMat Matrix
			expectMat = append(expectMat, expectSeries...)

			result := CombineEvalUsability([]Matrix{inputMat, inputMat2}, 1, 0)
			So(result, ShouldResemble, expectMat)
		})

		Convey("filling with precedingMissingPolicy=1, middleMissingPolicy=-1", func() {
			expectSeries := []Series{
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "nas_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 1}, {T: 1652320260000, V: 1}, {T: 1652320320000, V: 1}, {T: 1652320380000, V: 1}, {T: 1652320440000, V: 1}, {T: 1652320500000, V: 1}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "node_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 1}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 1}, {T: 1652320800000, V: 1}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "noee2_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 0}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "nas1_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 1}, {T: 1652319960000, V: 1}, {T: 1652320020000, V: 1}, {T: 1652320080000, V: 1}, {T: 1652320140000, V: 1}, {T: 1652320200000, V: 1}, {T: 1652320260000, V: 1}, {T: 1652320320000, V: 1}, {T: 1652320380000, V: 1}, {T: 1652320440000, V: 1}, {T: 1652320500000, V: 1}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "noee_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 1}, {T: 1652319960000, V: 1}, {T: 1652320020000, V: 1}, {T: 1652320080000, V: 1}, {T: 1652320140000, V: 1}, {T: 1652320200000, V: 1}, {T: 1652320260000, V: 1}, {T: 1652320320000, V: 1}, {T: 1652320380000, V: 1}, {T: 1652320440000, V: 1}, {T: 1652320500000, V: 1}, {T: 1652320560000, V: 1}, {T: 1652320620000, V: 1}, {T: 1652320680000, V: 1}, {T: 1652320740000, V: 1}, {T: 1652320800000, V: 1}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
					},
				},
			}
			var expectMat Matrix
			expectMat = append(expectMat, expectSeries...)

			result := CombineEvalUsability([]Matrix{inputMat, inputMat2}, 1, -1)
			So(result, ShouldResemble, expectMat)
		})

		Convey("filling with precedingMissingPolicy=1, middleMissingPolicy=1", func() {
			expectSeries := []Series{
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "nas_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 1}, {T: 1652320260000, V: 1}, {T: 1652320320000, V: 1}, {T: 1652320380000, V: 1}, {T: 1652320440000, V: 1}, {T: 1652320500000, V: 1}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "node_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 1}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 1}, {T: 1652320800000, V: 1}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "noee2_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 0}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "nas1_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 1}, {T: 1652319960000, V: 1}, {T: 1652320020000, V: 1}, {T: 1652320080000, V: 1}, {T: 1652320140000, V: 1}, {T: 1652320200000, V: 1}, {T: 1652320260000, V: 1}, {T: 1652320320000, V: 1}, {T: 1652320380000, V: 1}, {T: 1652320440000, V: 1}, {T: 1652320500000, V: 1}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "noee_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652319900000, V: 1}, {T: 1652319960000, V: 1}, {T: 1652320020000, V: 1}, {T: 1652320080000, V: 1}, {T: 1652320140000, V: 1}, {T: 1652320200000, V: 1}, {T: 1652320260000, V: 1}, {T: 1652320320000, V: 1}, {T: 1652320380000, V: 1}, {T: 1652320440000, V: 1}, {T: 1652320500000, V: 1}, {T: 1652320560000, V: 1}, {T: 1652320620000, V: 1}, {T: 1652320680000, V: 1}, {T: 1652320740000, V: 1}, {T: 1652320800000, V: 1}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
					},
				},
			}
			var expectMat Matrix
			expectMat = append(expectMat, expectSeries...)

			result := CombineEvalUsability([]Matrix{inputMat, inputMat2}, 1, 1)
			So(result, ShouldResemble, expectMat)
		})

	})
}

func TestCalculateUnavailableTime(t *testing.T) {

	Convey("Test CalculateUnavailableTime", t, func() {
		inputSeries := []Series{
			{
				Metric: labels.Labels{&labels.Label{Name: "index", Value: "nas_statis-2022.05-0"}},
				Points: []Point{
					{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 0}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 0}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
				},
			},
			{
				Metric: labels.Labels{&labels.Label{Name: "index", Value: "node_statis-2022.05-0"}},
				Points: []Point{
					{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 1}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 1}, {T: 1652320800000, V: 1}, {T: 1652320860000, V: 1}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
				},
			},
			{
				Metric: labels.Labels{&labels.Label{Name: "index", Value: "noee2_statis-2022.05-0"}},
				Points: []Point{
					{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 0}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 0}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
				},
			},
			{
				Metric: labels.Labels{&labels.Label{Name: "index", Value: "nas1_statis-2022.05-0"}},
				Points: []Point{
					{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 0}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 0}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
				},
			},
			{
				Metric: labels.Labels{&labels.Label{Name: "index", Value: "noee_statis-2022.05-0"}},
				Points: []Point{
					{T: 1652319900000, V: 0}, {T: 1652319960000, V: 0}, {T: 1652320020000, V: 0}, {T: 1652320080000, V: 0}, {T: 1652320140000, V: 0}, {T: 1652320200000, V: 0}, {T: 1652320260000, V: 0}, {T: 1652320320000, V: 0}, {T: 1652320380000, V: 0}, {T: 1652320440000, V: 0}, {T: 1652320500000, V: 0}, {T: 1652320560000, V: 0}, {T: 1652320620000, V: 0}, {T: 1652320680000, V: 0}, {T: 1652320740000, V: 0}, {T: 1652320800000, V: 0}, {T: 1652320860000, V: 0}, {T: 1652320920000, V: 0}, {T: 1652320980000, V: 0}, {T: 1652321040000, V: 0}, {T: 1652321100000, V: 0}, {T: 1652321160000, V: 0}, {T: 1652321220000, V: 0}, {T: 1652321280000, V: 0}, {T: 1652321340000, V: 0}, {T: 1652321400000, V: 0},
				},
			},
		}
		var inputMat Matrix
		inputMat = append(inputMat, inputSeries...)
		rangeQuery := interfaces.Query{
			QueryStr:       "continuous_k_minute_downtime(5, -1, 0, a)",
			Start:          1652320200000,
			End:            1652321400000,
			Interval:       5 * 60000,
			FixedStart:     1652320200000,
			FixedEnd:       1652321400000,
			IsInstantQuery: false,
			IsMetricModel:  true,
			// LookBackDelta:  1200000,
			ModelId: "123456",
		}
		Convey("CalculateUnavailableTime when range query", func() {
			expectSeries := []Series{
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "nas_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652320200000, V: 5}, {T: 1652320500000, V: 5}, {T: 1652320800000, V: 5}, {T: 1652321100000, V: 5}, {T: 1652321400000, V: 5},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "node_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652320200000, V: 5}, {T: 1652320500000, V: 3}, {T: 1652320800000, V: 3}, {T: 1652321100000, V: 4}, {T: 1652321400000, V: 5},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "noee2_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652320200000, V: 5}, {T: 1652320500000, V: 5}, {T: 1652320800000, V: 5}, {T: 1652321100000, V: 5}, {T: 1652321400000, V: 5},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "nas1_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652320200000, V: 5}, {T: 1652320500000, V: 5}, {T: 1652320800000, V: 5}, {T: 1652321100000, V: 5}, {T: 1652321400000, V: 5},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "noee_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652320200000, V: 5}, {T: 1652320500000, V: 5}, {T: 1652320800000, V: 5}, {T: 1652321100000, V: 5}, {T: 1652321400000, V: 5},
					},
				},
			}
			var expectMat Matrix
			expectMat = append(expectMat, expectSeries...)

			result := CalculateUnavailableTime(inputMat, rangeQuery, 3)
			So(result, ShouldResemble, expectMat)
		})

		Convey("CalculateUnavailableTime when instant query", func() {
			expectSeries := []Series{
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "nas_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652321400000, V: 21},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "node_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652321400000, V: 17},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "noee2_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652321400000, V: 21},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "nas1_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652321400000, V: 21},
					},
				},
				{
					Metric: labels.Labels{&labels.Label{Name: "index", Value: "noee_statis-2022.05-0"}},
					Points: []Point{
						{T: 1652321400000, V: 21},
					},
				},
			}
			var expectMat Matrix
			expectMat = append(expectMat, expectSeries...)

			instantQuery := interfaces.Query{
				QueryStr:       "continuous_k_minute_downtime(5, -1, 0, a)",
				Start:          1652321400000,
				End:            1652321400000,
				Interval:       5 * 60000,
				FixedStart:     1652320200000,
				FixedEnd:       1652321400000,
				IsInstantQuery: true,
				IsMetricModel:  true,
				// LookBackDelta:  1200000,
				ModelId: "123456",
			}

			result := CalculateUnavailableTime(inputMat, instantQuery, 3)
			So(result, ShouldResemble, expectMat)
		})
	})
}
