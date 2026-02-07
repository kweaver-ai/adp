// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package static

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"uniquery/logics/promql/labels"
)

func TestStringString(t *testing.T) {

	Convey("test String String ", t, func() {
		s := String{
			T: 1646360698347,
			V: "1",
		}
		actual := s.String()
		So(actual, ShouldEqual, "1")
	})
}

func TestScalarString(t *testing.T) {

	Convey("test Scalar String ", t, func() {
		scalar := Scalar{
			T: 1646360698347,
			V: 1.1,
		}
		So(scalar.String(), ShouldEqual, "scalar: 1.1 @[1646360698347]")
	})
}

func TestSeriesString(t *testing.T) {

	Convey("test Series String ", t, func() {
		series := Series{
			Metric: []*labels.Label{
				{
					Name:  "cluster",
					Value: "txy",
				},
				{
					Name:  "name",
					Value: "node-1",
				},
			},
			Points: []Point{
				{
					T: 1646360670000,
					V: 8,
				},
			},
		}
		So(series.String(), ShouldEqual, "{cluster=\"txy\", name=\"node-1\"} =>\n8 @[1646360670000]")
	})
}

func TestPointString(t *testing.T) {

	Convey("test Point String ", t, func() {
		point := Point{
			T: 1646360698347,
			V: 1.1,
		}
		So(point.String(), ShouldEqual, "1.1 @[1646360698347]")
	})
}

func TestVectorString(t *testing.T) {

	Convey("test Vector String ", t, func() {
		samples := []Sample{
			{
				Point: Point{
					T: 1646360698347,
					V: 1.1,
				},
				Metric: []*labels.Label{
					{
						Name:  "cluster",
						Value: "txy",
					},
					{
						Name:  "name",
						Value: "node-1",
					},
				},
			},
			{
				Point: Point{
					T: 1646360718347,
					V: 2.1,
				},
				Metric: []*labels.Label{
					{
						Name:  "cluster",
						Value: "txy",
					},
					{
						Name:  "name",
						Value: "node-1",
					},
				},
			},
		}
		actual := make(Vector, 0)
		actual = append(actual, samples...)
		So(actual.String(), ShouldEqual, "1.1 @[1646360698347]\n2.1 @[1646360718347]")
	})
}

func TestMatrixString(t *testing.T) {

	Convey("test Matrix String ", t, func() {
		series := []Series{
			{
				Metric: []*labels.Label{
					{
						Name:  "cluster",
						Value: "txy",
					},
					{
						Name:  "name",
						Value: "node-1",
					},
				},
				Points: []Point{
					{
						T: 1646360670000,
						V: 8,
					},
					{
						T: 1646360700000,
						V: 9,
					},
				},
			},
		}
		actual := make(Matrix, 0)
		actual = append(actual, series...)

		So(actual.String(), ShouldEqual, "{cluster=\"txy\", name=\"node-1\"} =>\n8 @[1646360670000]\n9 @[1646360700000]")
	})
}

func TestVectorContainsSameLabelset(t *testing.T) {
	Convey("test ContainsSameLabelset ", t, func() {
		Convey("Contains ", func() {
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
			expectVector2 := make(Vector, 0)
			expectVector2 = append(expectVector2, lhsVector2...)
			b := expectVector2.ContainsSameLabelset()

			So(b, ShouldBeTrue)
		})

		Convey("Not Contains ", func() {
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
						{Name: "node", Value: "node-2"},
					},
				},
			}
			expectVector2 := make(Vector, 0)
			expectVector2 = append(expectVector2, lhsVector2...)
			b := expectVector2.ContainsSameLabelset()

			So(b, ShouldBeFalse)
		})
	})
}

func TestMatrixContainsSameLabelset(t *testing.T) {
	Convey("test ContainsSameLabelset ", t, func() {
		Convey("Contains ", func() {
			series := []Series{
				{
					Metric: []*labels.Label{
						{
							Name:  "cluster",
							Value: "txy",
						},
						{
							Name:  "name",
							Value: "node-1",
						},
					},
					Points: []Point{
						{
							T: 1646360670000,
							V: 8,
						},
						{
							T: 1646360700000,
							V: 9,
						},
					},
				},
				{
					Metric: []*labels.Label{
						{
							Name:  "cluster",
							Value: "txy",
						},
						{
							Name:  "name",
							Value: "node-1",
						},
					},
					Points: []Point{
						{
							T: 1646360670000,
							V: 8,
						},
						{
							T: 1646360700000,
							V: 9,
						},
					},
				},
			}
			actual := make(Matrix, 0)
			actual = append(actual, series...)
			b := actual.ContainsSameLabelset()

			So(b, ShouldBeTrue)
		})

		Convey("Not Contains ", func() {
			series := []Series{
				{
					Metric: []*labels.Label{
						{
							Name:  "cluster",
							Value: "txy",
						},
						{
							Name:  "name",
							Value: "node-1",
						},
					},
					Points: []Point{
						{
							T: 1646360670000,
							V: 8,
						},
						{
							T: 1646360700000,
							V: 9,
						},
					},
				},
				{
					Metric: []*labels.Label{
						{
							Name:  "cluster",
							Value: "txy",
						},
						{
							Name:  "name",
							Value: "node-2",
						},
					},
					Points: []Point{
						{
							T: 1646360670000,
							V: 8,
						},
						{
							T: 1646360700000,
							V: 9,
						},
					},
				},
			}
			actual := make(Matrix, 0)
			actual = append(actual, series...)
			b := actual.ContainsSameLabelset()

			So(b, ShouldBeFalse)
		})
	})
}

func TestChangesPointString(t *testing.T) {

	Convey("test ChangesPoint String ", t, func() {
		p := ChangesPoint{
			FirstTimestamp: 1646360698347,
			FirstValue:     123,
			LastTimestamp:  1646360699347,
			LastValue:      124,
			Changes:        3,
		}

		So(p.String(), ShouldEqual, "FirstTimestamp:1646360698347,FirstValue:123,LastTimestamp:1646360699347,LastValue:124,Changes:3")
	})
}

func TestAGGPointString(t *testing.T) {

	Convey("test AGGPoint String ", t, func() {
		p := AGGPoint{
			Value: 124,
			Count: 3,
		}

		So(p.String(), ShouldEqual, "Value:124,Count:3")
	})
}
