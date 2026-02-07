// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package static

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"testing"
	"time"

	. "github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"

	"uniquery/interfaces"
	"uniquery/logics/data_dict"
	"uniquery/logics/promql/labels"
	"uniquery/logics/promql/parser"
)

var (
	keyNode1 = "node-1"
)

func TestFuncMath(t *testing.T) {
	vals := []parser.Value{Vector{
		{
			Point: Point{T: 1650883566000, V: 2.79},
			Metric: labels.Labels{
				&labels.Label{Name: "cluster", Value: "txy"},
				&labels.Label{Name: "name", Value: keyNode1},
			},
		},
	}}
	args := parser.Expressions{&parser.MatrixSelector{}}

	Convey("test func matb", t, func() {
		Convey("invoke floor", func() {
			var f funcFloor = funcFloor{}
			enh := &EvalNodeHelper{Out: Vector{}}
			vec := f.Call(vals, args, enh)
			So(vec[0].Point.V, ShouldEqual, 2)
		})

		Convey("invoke ceil", func() {
			var f funcCeil = funcCeil{}
			enh := &EvalNodeHelper{Out: Vector{}}
			vec := f.Call(vals, args, enh)
			So(vec[0].Point.V, ShouldEqual, 3)
		})

		Convey("invoke abs", func() {
			var f funcAbs = funcAbs{}
			enh := &EvalNodeHelper{Out: Vector{}}
			vec := f.Call(vals, args, enh)
			So(vec[0].Point.V, ShouldEqual, 2.79)
		})

		Convey("invoke exp", func() {
			var f funcExp = funcExp{}
			enh := &EvalNodeHelper{Out: Vector{}}
			vec := f.Call(vals, args, enh)
			So(vec[0].Point.V, ShouldEqual, 16.28101980178843)
		})

		Convey("invoke sqrt", func() {
			var f funcSqrt = funcSqrt{}
			enh := &EvalNodeHelper{Out: Vector{}}
			vec := f.Call(vals, args, enh)
			So(vec[0].Point.V, ShouldEqual, 1.6703293088490065)
		})

		Convey("invoke ln", func() {
			var f funcLn = funcLn{}
			enh := &EvalNodeHelper{Out: Vector{}}
			vec := f.Call(vals, args, enh)
			So(vec[0].Point.V, ShouldEqual, 1.0260415958332743)
		})

		Convey("invoke log2", func() {
			var f funcLog2 = funcLog2{}
			enh := &EvalNodeHelper{Out: Vector{}}
			vec := f.Call(vals, args, enh)
			So(vec[0].Point.V, ShouldEqual, 1.4802651220544627)
		})

		Convey("invoke log10", func() {
			var f funcLog10 = funcLog10{}
			enh := &EvalNodeHelper{Out: Vector{}}
			vec := f.Call(vals, args, enh)
			So(vec[0].Point.V, ShouldEqual, 0.44560420327359757)
		})

	})
}

func TestFuncTime(t *testing.T) {
	enh := EvalNodeHelper{Ts: 111000}

	Convey("test function time", t, func() {
		var f funcTime = funcTime{}
		res := f.Call(nil, nil, &enh)

		So(res, ShouldResemble, Vector{Sample{Point: Point{
			V: 111,
		}}})
	})

}

func TestSort(t *testing.T) {
	Convey("test sort function ", t, func() {
		var f funcSort = funcSort{}
		Convey("each of these points has a value of normal number ", func() {
			data := []parser.Value{
				Vector{
					Sample{Metric: nil, Point: Point{T: 1655346170, V: 2}},
					Sample{Metric: nil, Point: Point{T: 1655346170, V: 1}},
					Sample{Metric: nil, Point: Point{T: 1655346170, V: 4}},
					Sample{Metric: nil, Point: Point{T: 1655346170, V: 3}},
				},
			}
			res := f.Call(data, nil, nil)
			So(res[0].Point.V, ShouldEqual, 1)
			So(res[1].Point.V, ShouldEqual, 2)
			So(res[2].Point.V, ShouldEqual, 3)
			So(res[3].Point.V, ShouldEqual, 4)
		})
		Convey("a point has a value of NaN ", func() {
			data := []parser.Value{
				Vector{
					Sample{Metric: nil, Point: Point{T: 1655346170, V: 2}},
					Sample{Metric: nil, Point: Point{T: 1655346170, V: 1}},
					Sample{Metric: nil, Point: Point{T: 1655346170, V: math.NaN()}},
					Sample{Metric: nil, Point: Point{T: 1655346170, V: 3}},
				},
			}
			res := f.Call(data, nil, nil)
			So(res[0].Point.V, ShouldEqual, 1)
			So(res[1].Point.V, ShouldEqual, 2)
			So(res[2].Point.V, ShouldEqual, 3)
		})
	})
}

func TestSortDesc(t *testing.T) {
	Convey("test sort_desc function", t, func() {
		var f funcSortDesc = funcSortDesc{}
		Convey("each of these points has a value of normal number ", func() {
			data := []parser.Value{
				Vector{
					Sample{Metric: nil, Point: Point{T: 1655346170, V: 2}},
					Sample{Metric: nil, Point: Point{T: 1655346170, V: 1}},
					Sample{Metric: nil, Point: Point{T: 1655346170, V: 4}},
					Sample{Metric: nil, Point: Point{T: 1655346170, V: 3}},
				},
			}
			res := f.Call(data, nil, nil)
			So(res[0].Point.V, ShouldEqual, 4)
			So(res[1].Point.V, ShouldEqual, 3)
			So(res[2].Point.V, ShouldEqual, 2)
			So(res[3].Point.V, ShouldEqual, 1)
		})
		Convey("a point has a value of NaN ", func() {
			data := []parser.Value{
				Vector{
					Sample{Metric: nil, Point: Point{T: 1655346170, V: 2}},
					Sample{Metric: nil, Point: Point{T: 1655346170, V: 1}},
					Sample{Metric: nil, Point: Point{T: 1655346170, V: math.NaN()}},
					Sample{Metric: nil, Point: Point{T: 1655346170, V: 3}},
				},
			}
			res := f.Call(data, nil, nil)
			So(res[0].Point.V, ShouldEqual, 3)
			So(res[1].Point.V, ShouldEqual, 2)
			So(res[2].Point.V, ShouldEqual, 1)
		})
	})
}

func TestFuncLabelReplace(t *testing.T) {
	vals := []parser.Value{
		Vector{
			Sample{
				Metric: labels.Labels{
					{
						Name:  "cluster",
						Value: "txy",
					},
					{
						Name:  "name",
						Value: keyNode1,
					},
				},
				Point: Point{
					T: 1650883566000,
					V: 361,
				},
			},
		},
	}
	var f funcLabelReplace = funcLabelReplace{}
	Convey("test function label_replace ok", t, func() {

		args := parser.Expressions{
			&parser.VectorSelector{
				Name:           "prometheus.metrics.elasticsearch_os_cpu_percent2",
				StartOrEnd:     0,
				OriginalOffset: 0,
				Offset:         0,
				LabelMatchers: []*labels.Matcher{
					{
						Name:  "labels.cluster",
						Value: "txy",
					},
					{
						Name:  "__name__",
						Value: "prometheus.metrics.elasticsearch_os_cpu_percent2",
					},
				},
				PosRange: parser.PositionRange{Start: 14, End: 139},
			},
			&parser.StringLiteral{
				Val:      "new",
				PosRange: parser.PositionRange{Start: 140, End: 145},
			},
			&parser.StringLiteral{
				Val:      "$1",
				PosRange: parser.PositionRange{Start: 146, End: 150},
			},
			&parser.StringLiteral{
				Val:      "name",
				PosRange: parser.PositionRange{Start: 151, End: 157},
			},
			&parser.StringLiteral{
				Val:      "^([a-z]+)-[0-9]+$",
				PosRange: parser.PositionRange{Start: 158, End: 177},
			},
		}
		enh := EvalNodeHelper{}

		var eres = Vector{
			Sample{
				Metric: labels.Labels{
					{
						Name:  "cluster",
						Value: "txy",
					},
					{
						Name:  "name",
						Value: keyNode1,
					},
					{
						Name:  "new",
						Value: "node",
					},
				},
				Point: Point{
					T: 0,
					V: 361,
				},
			},
		}
		fmt.Print(eres[0])
		res := f.Call(vals, args, &enh)

		So(res, ShouldResemble, eres)
	})

	Convey("test function label_replace regex no match", t, func() {
		args_1 := parser.Expressions{
			&parser.VectorSelector{
				Name:           "prometheus.metrics.elasticsearch_os_cpu_percent2",
				StartOrEnd:     0,
				OriginalOffset: 0,
				Offset:         0,
				LabelMatchers: []*labels.Matcher{
					{
						Name:  "labels.cluster",
						Value: "txy",
					},
					{
						Name:  "__name__",
						Value: "prometheus.metrics.elasticsearch_os_cpu_percent2",
					},
				},
				PosRange: parser.PositionRange{Start: 14, End: 139},
			},
			&parser.StringLiteral{
				Val:      "new",
				PosRange: parser.PositionRange{Start: 140, End: 145},
			},
			&parser.StringLiteral{
				Val:      "$0",
				PosRange: parser.PositionRange{Start: 146, End: 150},
			},
			&parser.StringLiteral{
				Val:      "cluster",
				PosRange: parser.PositionRange{Start: 151, End: 160},
			},
			&parser.StringLiteral{
				Val:      "[0-9]",
				PosRange: parser.PositionRange{Start: 161, End: 168},
			},
		}
		enh := EvalNodeHelper{}

		var eres = Vector{
			Sample{
				Metric: labels.Labels{
					{
						Name:  "cluster",
						Value: "txy",
					},
					{
						Name:  "name",
						Value: keyNode1,
					},
				},
				Point: Point{
					T: 0,
					V: 361,
				},
			},
		}

		res := f.Call(vals, args_1, &enh)

		So(res, ShouldResemble, eres)
	})

}

func TestFuncLabelJoin(t *testing.T) {
	vals := []parser.Value{
		Vector{
			Sample{
				Metric: labels.Labels{
					{
						Name:  "cluster",
						Value: "txy",
					},
					{
						Name:  "name",
						Value: keyNode1,
					},
				},
				Point: Point{
					T: 1650883566000,
					V: 361,
				},
			},
		},
	}
	args := parser.Expressions{
		&parser.VectorSelector{
			Name:           "prometheus.metrics.elasticsearch_os_cpu_percent2",
			StartOrEnd:     0,
			OriginalOffset: 0,
			Offset:         0,
			LabelMatchers: []*labels.Matcher{
				{
					Name:  "labels.cluster",
					Value: "txy",
				},
				{
					Name:  "__name__",
					Value: "prometheus.metrics.elasticsearch_os_cpu_percent2",
				},
			},
			PosRange: parser.PositionRange{Start: 14, End: 139},
		},
		&parser.StringLiteral{
			Val:      "new",
			PosRange: parser.PositionRange{Start: 140, End: 145},
		},
		&parser.StringLiteral{
			Val:      "+",
			PosRange: parser.PositionRange{Start: 146, End: 149},
		},
		&parser.StringLiteral{
			Val:      "cluster",
			PosRange: parser.PositionRange{Start: 150, End: 159},
		},
		&parser.StringLiteral{
			Val:      "name",
			PosRange: parser.PositionRange{Start: 160, End: 166},
		},
	}
	enh := EvalNodeHelper{}
	var f funcLabelJoin = funcLabelJoin{}
	Convey("test function label_join ok", t, func() {

		var eres = Vector{
			Sample{
				Metric: labels.Labels{
					{
						Name:  "cluster",
						Value: "txy",
					},
					{
						Name:  "name",
						Value: keyNode1,
					},
					{
						Name:  "new",
						Value: "txy+node-1",
					},
				},
				Point: Point{
					T: 0,
					V: 361,
				},
			},
		}

		res := f.Call(vals, args, &enh)

		So(res, ShouldResemble, eres)
	})
}

func TestFuncRate(t *testing.T) {
	Convey("test func rate", t, func() {
		Convey("invoke extrapolatedRate", func() {
			vals := []parser.Value{RatePoint{PointsCount: 1}}
			args := parser.Expressions{&parser.MatrixSelector{}}
			enh := &EvalNodeHelper{Out: Vector{}}
			var f funcRate = funcRate{}
			vec := f.Call(vals, args, enh)

			So(len(vec), ShouldEqual, 0)
		})
	})
}

func TestFuncHistogramQuantile(t *testing.T) {
	Convey("test func HistogramQuantile", t, func() {
		var f funcHistogramQuantile = funcHistogramQuantile{}
		Convey("invoke HistogramQuantile", func() {
			vals := []parser.Value{
				Vector{
					{Point{T: 123, V: 0.9}, labels.Labels{}},
				},
				Vector{
					{Point{T: 123, V: 125}, labels.Labels{{Name: "le", Value: "0.3"}}},
					{Point{T: 123, V: 125}, labels.Labels{{Name: "le", Value: "0.3"}}},
					{Point{T: 123, V: 125}, labels.Labels{{Name: "le", Value: "+Inf"}}},
				}}
			args := parser.Expressions{&parser.MatrixSelector{}}
			mb := metricWithBuckets{
				metric:  labels.Labels{{Name: "le", Value: "0.3"}},
				buckets: buckets{{upperBound: 0.1, count: 100}},
			}
			enh := &EvalNodeHelper{Out: Vector{}, signatureToMetricWithBuckets: map[string]*metricWithBuckets{"abc": &mb}}

			vec := f.Call(vals, args, enh)
			So(len(vec), ShouldEqual, 1)
		})

		Convey("cover error data which is Non-increasing", func() {
			vals := []parser.Value{
				Vector{
					{Point{T: 123, V: 0.9}, labels.Labels{}},
				},
				Vector{
					{Point{T: 123, V: 123}, labels.Labels{{Name: "le", Value: "0.1"}}},
					{Point{T: 123, V: 113}, labels.Labels{{Name: "le", Value: "0.3"}}},
					{Point{T: 123, V: 125}, labels.Labels{{Name: "le", Value: "+Inf"}}},
				}}
			args := parser.Expressions{&parser.MatrixSelector{}}
			enh := &EvalNodeHelper{Out: Vector{}}

			vec := f.Call(vals, args, enh)
			So(len(vec), ShouldEqual, 1)
		})

		Convey("cover error data that q < 0 ", func() {
			vals := []parser.Value{
				Vector{
					{Point{T: 123, V: -0.9}, labels.Labels{}},
				},
				Vector{
					{Point{T: 123, V: 123}, labels.Labels{{Name: "le", Value: "0.1"}}},
					{Point{T: 123, V: 113}, labels.Labels{{Name: "le", Value: "0.3"}}},
					{Point{T: 123, V: 125}, labels.Labels{{Name: "le", Value: "+Inf"}}},
				}}
			args := parser.Expressions{&parser.MatrixSelector{}}
			enh := &EvalNodeHelper{Out: Vector{}}
			defer func() {
				r := recover()
				if err, ok := r.(error); ok {
					So(err.Error(), ShouldEqual, "invalid parameter of histogram_quantile: -0.900000, it should be bigger than 0 and smaller than 1")
				}
			}()
			f.Call(vals, args, enh)
		})

		Convey("cover error data that q > 1 ", func() {
			vals := []parser.Value{
				Vector{
					{Point{T: 123, V: 1.9}, labels.Labels{}},
				},
				Vector{
					{Point{T: 123, V: 123}, labels.Labels{{Name: "le", Value: "0.1"}}},
					{Point{T: 123, V: 113}, labels.Labels{{Name: "le", Value: "0.3"}}},
					{Point{T: 123, V: 125}, labels.Labels{{Name: "le", Value: "+Inf"}}},
				}}
			args := parser.Expressions{&parser.MatrixSelector{}}
			enh := &EvalNodeHelper{Out: Vector{}}
			defer func() {
				r := recover()
				if err, ok := r.(error); ok {
					So(err.Error(), ShouldEqual, "invalid parameter of histogram_quantile: 1.900000, it should be bigger than 0 and smaller than 1")
				}
			}()
			f.Call(vals, args, enh)
		})

		Convey("upperBound is not +Inf", func() {
			vals := []parser.Value{
				Vector{
					{Point{T: 123, V: 0.9}, labels.Labels{}},
				},
				Vector{
					{Point{T: 123, V: 123}, labels.Labels{{Name: "le", Value: "0.1"}}},
					{Point{T: 123, V: 125}, labels.Labels{{Name: "le", Value: "0.3"}}},
					{Point{T: 123, V: 125}, labels.Labels{{Name: "le", Value: "9.1"}}},
				}}
			args := parser.Expressions{&parser.MatrixSelector{}}
			enh := &EvalNodeHelper{Out: Vector{}}

			vec := f.Call(vals, args, enh)
			So(len(vec), ShouldEqual, 1)
		})

		Convey("the number of buckets is less than 2", func() {
			vals := []parser.Value{
				Vector{
					{Point{T: 123, V: 0.9}, labels.Labels{}},
				},
				Vector{
					{Point{T: 123, V: 125}, labels.Labels{{Name: "le", Value: "+Inf"}}},
				}}
			args := parser.Expressions{&parser.MatrixSelector{}}
			enh := &EvalNodeHelper{Out: Vector{}}

			vec := f.Call(vals, args, enh)
			So(len(vec), ShouldEqual, 1)
		})

		Convey("the value of last bucket is 0", func() {
			vals := []parser.Value{
				Vector{
					{Point{T: 123, V: 0.9}, labels.Labels{}},
				},
				Vector{
					{Point{T: 123, V: 0}, labels.Labels{{Name: "le", Value: "0.1"}}},
					{Point{T: 123, V: 0}, labels.Labels{{Name: "le", Value: "+Inf"}}},
				}}
			args := parser.Expressions{&parser.MatrixSelector{}}
			enh := &EvalNodeHelper{Out: Vector{}}

			vec := f.Call(vals, args, enh)
			So(len(vec), ShouldEqual, 1)
			So(strconv.FormatFloat(vec[0].Point.V, 'f', -1, 32), ShouldEqual, "NaN")
		})

		Convey("the result is in the last bucket", func() {
			vals := []parser.Value{
				Vector{
					{Point{T: 123, V: 0.9}, labels.Labels{}},
				},
				Vector{
					{Point{T: 123, V: 80}, labels.Labels{{Name: "le", Value: "0.1"}}},
					{Point{T: 123, V: 200}, labels.Labels{{Name: "le", Value: "+Inf"}}},
				}}
			args := parser.Expressions{&parser.MatrixSelector{}}
			enh := &EvalNodeHelper{Out: Vector{}}

			vec := f.Call(vals, args, enh)

			So(len(vec), ShouldEqual, 1)
			So(vec[0].Point.V, ShouldEqual, 0.1)
		})

		Convey("the value of first bucket is less than 0 and the result is in the first bucket", func() {
			vals := []parser.Value{
				Vector{
					{Point{T: 123, V: 0.1}, labels.Labels{}},
				},
				Vector{
					{Point{T: 123, V: 12}, labels.Labels{{Name: "le", Value: "-0.1"}}},
					{Point{T: 123, V: 23}, labels.Labels{{Name: "le", Value: "0.1"}}},
					{Point{T: 123, V: 100}, labels.Labels{{Name: "le", Value: "+Inf"}}},
				}}
			args := parser.Expressions{&parser.MatrixSelector{}}
			enh := &EvalNodeHelper{Out: Vector{}}

			vec := f.Call(vals, args, enh)
			So(len(vec), ShouldEqual, 1)
			So(vec[0].Point.V, ShouldEqual, -0.1)

			fmt.Printf("the result: %+v, type:%+v, vector:%f, \n", vec, vec.Type(), vec[0].Point.V)
		})

		Convey("correct the value of first bucket is less than 0 and the result is in the first bucket", func() {
			vals := []parser.Value{
				Vector{
					{Point{T: 123, V: 0.9}, labels.Labels{}},
				},
				Vector{
					{Point{T: 123, V: 85}, labels.Labels{{Name: "le", Value: "0.1"}}},
					{Point{T: 123, V: 99}, labels.Labels{{Name: "le", Value: "0.3"}}},
					{Point{T: 123, V: 100}, labels.Labels{{Name: "le", Value: "+Inf"}}},
				}}
			args := parser.Expressions{&parser.MatrixSelector{}}
			enh := &EvalNodeHelper{Out: Vector{}}

			vec := f.Call(vals, args, enh)
			So(len(vec), ShouldEqual, 1)
			//So(vec[0].Point.V, ShouldEqual, -0.1)

			fmt.Printf("the result: %+v, type:%+v, vector:%f, \n", vec, vec.Type(), vec[0].Point.V)
		})

	})
}

func TestFuncIncrease(t *testing.T) {
	Convey("test func increase", t, func() {
		Convey("invoke extrapolatedRate", func() {
			vals := []parser.Value{RatePoint{PointsCount: 1}}
			args := parser.Expressions{&parser.MatrixSelector{}}
			enh := &EvalNodeHelper{Out: Vector{}}
			var f funcIncrease = funcIncrease{}
			vec := f.Call(vals, args, enh)

			So(len(vec), ShouldEqual, 0)
		})
	})
}

func TestExtrapolatedRate(t *testing.T) {
	Convey("test extrapolated rate", t, func() {

		Convey("1. there is one point in range", func() {
			vals := []parser.Value{RatePoint{PointsCount: 1}}
			args := parser.Expressions{&parser.MatrixSelector{}}
			enh := &EvalNodeHelper{Out: Vector{}}

			vec := extrapolatedRate(vals, args, enh, true, true)

			So(len(vec), ShouldEqual, 0)
		})

		Convey("2. there are multiple points in range", func() {
			vals := []parser.Value{RatePoint{
				FirstTimestamp:    1658371910000,
				FirstValue:        100,
				LastTimestamp:     1658371950000,
				LastValue:         555,
				CounterCorrection: 0.0,
				PointsCount:       5,
			}}
			args := parser.Expressions{&parser.MatrixSelector{
				Range: 1 * time.Minute,
			}}
			enh := &EvalNodeHelper{Ts: 1658371900000, Out: make(Vector, 0, 1)}

			vec := extrapolatedRate(vals, args, enh, true, true)

			So(len(vec), ShouldBeGreaterThan, 0)
			So(vec[0].Point.V, ShouldEqual, 11.145833333333334)
		})

		Convey("3. durationToZero < durationToStart, correct the left boundary of the range", func() {
			vals := []parser.Value{RatePoint{
				FirstTimestamp:    1658371940000,
				FirstValue:        100,
				LastTimestamp:     1658371950000,
				LastValue:         155,
				CounterCorrection: 0.0,
				PointsCount:       5,
			}}
			args := parser.Expressions{&parser.MatrixSelector{
				Range: 1 * time.Minute,
			}}
			enh := &EvalNodeHelper{Ts: 1658371900000, Out: make(Vector, 0, 1)}

			vec := extrapolatedRate(vals, args, enh, true, true)

			So(len(vec), ShouldBeGreaterThan, 0)
			So(vec[0].Point.V, ShouldEqual, 1.1458333333333333)
		})

	})
}

func TestFuncDelta(t *testing.T) {
	Convey("test funcDelta", t, func() {
		var f funcDelta = funcDelta{}
		Convey("1. there is one point in range", func() {
			vals := []parser.Value{DeltaPoint{PointsCount: 1}}
			args := parser.Expressions{&parser.MatrixSelector{}}
			enh := &EvalNodeHelper{Out: Vector{}}

			vec := f.Call(vals, args, enh)

			So(len(vec), ShouldEqual, 0)
		})

		Convey("2. there are multiple points in range", func() {
			vals := []parser.Value{DeltaPoint{
				FirstTimestamp: 1658371910000,
				FirstValue:     100,
				LastTimestamp:  1658371950000,
				LastValue:      555,
				PointsCount:    5,
			}}
			args := parser.Expressions{&parser.MatrixSelector{
				Range: 1 * time.Minute,
			}}
			enh := &EvalNodeHelper{Ts: 1658371900000, Out: make(Vector, 0, 1)}

			vec := f.Call(vals, args, enh)

			So(len(vec), ShouldBeGreaterThan, 0)
			So(vec[0].Point.V, ShouldEqual, 682.5)
		})

		Convey("3. durationToZero < durationToStart, correct the left boundary of the range", func() {
			vals := []parser.Value{DeltaPoint{
				FirstTimestamp: 1658371940000,
				FirstValue:     100,
				LastTimestamp:  1658371950000,
				LastValue:      155,
				PointsCount:    5,
			}}
			args := parser.Expressions{&parser.MatrixSelector{
				Range: 1 * time.Minute,
			}}
			enh := &EvalNodeHelper{Ts: 1658371900000, Out: make(Vector, 0, 1)}

			vec := f.Call(vals, args, enh)

			So(len(vec), ShouldBeGreaterThan, 0)
			So(vec[0].Point.V, ShouldEqual, 68.75)
		})

	})
}

func TestFuncPercentRank(t *testing.T) {
	Convey("test funcPercentRank", t, func() {
		var f funcPercentRank = funcPercentRank{}
		vals := []parser.Value{
			Vector{
				Sample{Metric: labels.Labels{{Name: "cluster", Value: "1"}}, Point: Point{T: 1655346170, V: 110}},
				Sample{Metric: labels.Labels{{Name: "cluster", Value: "2"}}, Point: Point{T: 1655346170, V: 11}},
				Sample{Metric: labels.Labels{{Name: "cluster", Value: "3"}}, Point: Point{T: 1655346170, V: 51}},
				Sample{Metric: labels.Labels{{Name: "cluster", Value: "4"}}, Point: Point{T: 1655346170, V: 3}},
			},
		}
		args := parser.Expressions{
			&parser.VectorSelector{},
			&parser.NumberLiteral{
				Val: 3,
			},
		}
		enh := EvalNodeHelper{}
		Convey("1. The given precision exceeds 16", func() {
			args := parser.Expressions{
				&parser.VectorSelector{},
				&parser.NumberLiteral{
					Val: 18,
				},
			}
			res := f.Call(vals, args, &enh)
			So(res[0].V, ShouldEqual, 100)
			So(res[1].V, ShouldEqual, 33.33333333333333)
			So(res[2].V, ShouldEqual, 66.66666666666666)
			So(res[3].V, ShouldEqual, 0)

		})
		Convey("2. The given precision less than 3", func() {
			args := parser.Expressions{
				&parser.VectorSelector{},
				&parser.NumberLiteral{
					Val: 1,
				},
			}
			res := f.Call(vals, args, &enh)
			So(res[0].V, ShouldEqual, 100)
			So(res[1].V, ShouldEqual, 30)
			So(res[2].V, ShouldEqual, 60)
			So(res[3].V, ShouldEqual, 0)

		})
		Convey("2. The data contains NaN ,INF ,-INF", func() {
			vals := []parser.Value{
				Vector{
					Sample{Metric: labels.Labels{{Name: "cluster", Value: "1"}}, Point: Point{T: 1655346170, V: math.NaN()}},
					Sample{Metric: labels.Labels{{Name: "cluster", Value: "2"}}, Point: Point{T: 1655346170, V: 0}},
					Sample{Metric: labels.Labels{{Name: "cluster", Value: "3"}}, Point: Point{T: 1655346170, V: math.Inf(1)}},
					Sample{Metric: labels.Labels{{Name: "cluster", Value: "4"}}, Point: Point{T: 1655346170, V: math.Inf(-1)}},
				},
			}
			res := f.Call(vals, args, &enh)
			So(strconv.FormatFloat(res[0].V, 'f', -1, 32), ShouldEqual, "NaN")
			So(res[1].V, ShouldEqual, 66.6)
			So(res[2].V, ShouldResemble, math.Inf(1))
			So(res[3].V, ShouldResemble, math.Inf(-1))
		})
		Convey("3. The length of vector equals 1 ", func() {
			vals := []parser.Value{
				Vector{
					Sample{Metric: labels.Labels{{Name: "cluster", Value: "2"}}, Point: Point{T: 1655346170, V: 0}},
				},
			}

			res := f.Call(vals, args, &enh)
			So(res[0].V, ShouldEqual, 100)
		})
		Convey("4. The length of vector equals 1 and value equals NaN", func() {
			vals := []parser.Value{
				Vector{
					Sample{Metric: labels.Labels{{Name: "cluster", Value: "2"}}, Point: Point{T: 1655346170, V: math.NaN()}},
				},
			}
			res := f.Call(vals, args, &enh)
			So(strconv.FormatFloat(res[0].V, 'f', -1, 32), ShouldEqual, "NaN")

		})

	})
}

func TestFuncRank(t *testing.T) {
	Convey("test funcRank", t, func() {
		var f funcRank = funcRank{}
		vals := []parser.Value{
			Vector{
				Sample{Metric: labels.Labels{{Name: "cluster", Value: "1"}}, Point: Point{T: 1655346170, V: 110}},
				Sample{Metric: labels.Labels{{Name: "cluster", Value: "2"}}, Point: Point{T: 1655346170, V: 11}},
				Sample{Metric: labels.Labels{{Name: "cluster", Value: "3"}}, Point: Point{T: 1655346170, V: 11}},
				Sample{Metric: labels.Labels{{Name: "cluster", Value: "4"}}, Point: Point{T: 1655346170, V: 3}},
			},
		}
		args := parser.Expressions{
			&parser.VectorSelector{},
			&parser.NumberLiteral{
				Val: 1,
			},
		}
		enh := EvalNodeHelper{}
		Convey("1. sort in descending order", func() {
			args := parser.Expressions{
				&parser.VectorSelector{},
				&parser.NumberLiteral{
					Val: 0,
				},
			}
			res := f.Call(vals, args, &enh)
			So(res[0].V, ShouldEqual, 1)
			So(res[1].V, ShouldEqual, 2)
			So(res[2].V, ShouldEqual, 2)
			So(res[3].V, ShouldEqual, 4)
		})
		Convey("2. sort in ascending order", func() {
			res := f.Call(vals, args, &enh)
			So(res[0].V, ShouldEqual, 4)
			So(res[1].V, ShouldEqual, 2)
			So(res[2].V, ShouldEqual, 2)
			So(res[3].V, ShouldEqual, 1)
		})
		Convey("3. The data contains NaN +INF -INF", func() {
			vals := []parser.Value{
				Vector{
					Sample{Metric: labels.Labels{{Name: "cluster", Value: "1"}}, Point: Point{T: 1655346170, V: math.NaN()}},
					Sample{Metric: labels.Labels{{Name: "cluster", Value: "2"}}, Point: Point{T: 1655346170, V: 0}},
					Sample{Metric: labels.Labels{{Name: "cluster", Value: "3"}}, Point: Point{T: 1655346170, V: math.Inf(1)}},
					Sample{Metric: labels.Labels{{Name: "cluster", Value: "4"}}, Point: Point{T: 1655346170, V: math.Inf(-1)}},
				},
			}
			res := f.Call(vals, args, &enh)
			So(strconv.FormatFloat(res[0].V, 'f', -1, 32), ShouldEqual, "NaN")
			So(res[1].V, ShouldEqual, 3)
			So(res[2].V, ShouldResemble, math.Inf(1))
			So(res[3].V, ShouldResemble, math.Inf(-1))
		})
	})
}

func TestFuncDictLabels(t *testing.T) {
	Convey("test funcDictLabels", t, func() {

		vals := []parser.Value{
			Vector{
				Sample{Metric: labels.Labels{{Name: "cluster", Value: "1"}, {Name: "cpu", Value: "cpu0"},
					{Name: "host_ip", Value: "host1"}, {Name: "mode", Value: "idle"}}, Point: Point{T: 1655346170, V: 110}},
				Sample{Metric: labels.Labels{{Name: "cluster", Value: "1"}, {Name: "cpu", Value: "cpu1"},
					{Name: "host_ip", Value: "host2"}, {Name: "mode", Value: "iowait"}}, Point: Point{T: 1655346170, V: 113}},
				Sample{Metric: labels.Labels{{Name: "cluster", Value: "1"}, {Name: "cpu", Value: "cpu2"},
					{Name: "host_ip", Value: "host1"}, {Name: "mode", Value: "irq"}}, Point: Point{T: 1655346170, V: 11}},
				Sample{Metric: labels.Labels{{Name: "cluster", Value: "1"}, {Name: "cpu", Value: "cpu3"},
					{Name: "host_ip", Value: "host1"}, {Name: "mode", Value: "nice"}}, Point: Point{T: 1655346170, V: 211}},
			},
		}

		expect := Vector{
			Sample{Metric: labels.Labels{
				{Name: "capacity", Value: "15G"},
				{Name: "cluster", Value: "1"},
				{Name: "cpu", Value: "cpu0"},
				{Name: "db", Value: "oracal"},
				{Name: "host_ip", Value: "host1"},
				{Name: "mode", Value: "idle"},
			},
				Point: Point{T: 1655346170, V: 110}},
			Sample{Metric: labels.Labels{
				{Name: "capacity", Value: "3G"},
				{Name: "cluster", Value: "1"},
				{Name: "cpu", Value: "cpu1"},
				{Name: "db", Value: "mysql"},
				{Name: "host_ip", Value: "host2"},
				{Name: "mode", Value: "iowait"},
			},
				Point: Point{T: 1655346170, V: 113}},
			Sample{Metric: labels.Labels{
				{Name: "cluster", Value: "1"},
				{Name: "cpu", Value: "cpu2"},
				{Name: "host_ip", Value: "host1"},
				{Name: "mode", Value: "irq"},
			},
				Point: Point{T: 1655346170, V: 11}},
			Sample{Metric: labels.Labels{
				{Name: "cluster", Value: "1"},
				{Name: "cpu", Value: "cpu3"},
				{Name: "host_ip", Value: "host1"},
				{Name: "mode", Value: "nice"},
			},
				Point: Point{T: 1655346170, V: 211}},
		}

		args := parser.Expressions{
			&parser.VectorSelector{Name: "vector"},
			&parser.StringLiteral{Val: "dict_name"},
			&parser.StringLiteral{Val: "key1"},
			&parser.StringLiteral{Val: "host_ip"},
			&parser.StringLiteral{Val: "key2"},
			&parser.StringLiteral{Val: "cpu"},
			&parser.StringLiteral{Val: "value1"},
			&parser.StringLiteral{Val: "db"},
			&parser.StringLiteral{Val: "value2"},
			&parser.StringLiteral{Val: "capacity"},
		}
		enh := EvalNodeHelper{}
		Convey("dict_labels", func() {
			patch := ApplyFunc(data_dict.GetDictByName,
				func(dictName string) (interfaces.DataDict, bool) {
					return interfaces.DataDict{
						UniqueKey: true,
						DictRecords: map[string][]map[string]string{
							"host1\x00cpu0": {
								{
									"key1":   "host1",
									"key2":   "cpu0",
									"value1": "oracal",
									"value2": "15G",
									"value3": "36",
								},
							},
							"host2\x00cpu1": {
								{
									"key1":   "host2",
									"key2":   "cpu1",
									"value1": "mysql",
									"value2": "3G",
									"value3": "89",
								},
							},
							"host3\x00cpu0": {
								{
									"key1":   "host3",
									"key2":   "cpu0",
									"value1": "redis",
									"value2": "12G",
									"value3": "99",
								},
							},
						},
						Dimension: interfaces.Dimension{
							Keys: []interfaces.DimensionItem{
								{Name: "key1"},
								{Name: "key2"},
							},
							Values: []interfaces.DimensionItem{
								{Name: "value1"},
								{Name: "value2"},
								{Name: "value3"},
							},
						},
					}, true
				},
			)
			defer patch.Reset()

			var f funcDictLabels = funcDictLabels{}
			res := f.New(args).Call(vals, args, &enh)
			So(res, ShouldResemble, expect)
		})

	})
}

func TestFuncDictValues(t *testing.T) {
	Convey("test fucnDictValues", t, func() {
		vals := []parser.Value{}
		args := parser.Expressions{
			&parser.StringLiteral{Val: "dict_name"},
			&parser.StringLiteral{Val: "value3"},
			&parser.StringLiteral{Val: "key1"},
			&parser.StringLiteral{Val: "host_ip"},
			&parser.StringLiteral{Val: "value1"},
			&parser.StringLiteral{Val: "db"},
		}
		enh := EvalNodeHelper{}
		expect := Vector{
			Sample{Metric: labels.Labels{
				{Name: "db", Value: "mysql"},
				{Name: "host_ip", Value: "host2"},
				{Name: "key2", Value: "cpu1"},
				{Name: "key3", Value: "node-1"},
			},
				Point: Point{V: 89}},
			Sample{Metric: labels.Labels{
				{Name: "db", Value: "oracal"},
				{Name: "host_ip", Value: "host1"},
				{Name: "key2", Value: "cpu0"},
				{Name: "key3", Value: "node-1"},
			},
				Point: Point{V: 36}},
			Sample{Metric: labels.Labels{
				{Name: "db", Value: "redis"},
				{Name: "host_ip", Value: "host3"},
				{Name: "key2", Value: "cpu0"},
				{Name: "key3", Value: "node-3"},
			},
				Point: Point{V: 99}},
		}
		Convey("dict_values", func() {

			patch := ApplyFunc(data_dict.GetDictByName,
				func(dictName string) (interfaces.DataDict, bool) {
					return interfaces.DataDict{
						UniqueKey: true,
						DictRecords: map[string][]map[string]string{
							"host1\x00cpu0\x00node-1": {{
								"key1":   "host1",
								"key2":   "cpu0",
								"key3":   "node-1",
								"value1": "oracal",
								"value2": "15G",
								"value3": "36",
							}},
							"host2\x00cpu1\x00node-1": {{
								"key1":   "host2",
								"key2":   "cpu1",
								"key3":   "node-1",
								"value1": "mysql",
								"value2": "3G",
								"value3": "89"}},
							"host3\x00cpu0\x00node-3": {{
								"key1":   "host3",
								"key2":   "cpu0",
								"key3":   "node-3",
								"value1": "redis",
								"value2": "12G",
								"value3": "99"}},
						},
						Dimension: interfaces.Dimension{
							Keys: []interfaces.DimensionItem{
								{Name: "key1"},
								{Name: "key2"},
								{Name: "key3"},
							},
							Values: []interfaces.DimensionItem{
								{Name: "value1"},
								{Name: "value2"},
								{Name: "value3"},
							},
						},
					}, true
				},
			)
			defer patch.Reset()

			var f funcDictValues = funcDictValues{}
			res := f.New(args).Call(vals, args, &enh)

			sort.Slice(res, func(i, j int) bool {
				return res[i].Metric[0].Value < res[j].Metric[0].Value
			})

			So(res, ShouldResemble, expect)
		})

	})
}
