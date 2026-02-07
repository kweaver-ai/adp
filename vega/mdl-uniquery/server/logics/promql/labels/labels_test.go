// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package labels

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/require"
)

func TestLabelsFromStrings(t *testing.T) {
	labels := FromStrings("aaa", "111", "bbb", "222")
	expected := Labels{
		{
			Name:  "aaa",
			Value: "111",
		},
		{
			Name:  "bbb",
			Value: "222",
		},
	}

	require.Equal(t, expected, labels, "unexpected labelset")

	// require.Panics(t, func() { FromStrings("aaa", "111", "bbb") })
}

func TestLabelsCompare(t *testing.T) {
	labels := Labels{
		{
			Name:  "aaa",
			Value: "111",
		},
		{
			Name:  "bbb",
			Value: "222",
		},
	}

	tests := []struct {
		compared Labels
		expected int
	}{
		{
			compared: Labels{
				{
					Name:  "aaa",
					Value: "110",
				},
				{
					Name:  "bbb",
					Value: "222",
				},
			},
			expected: 1,
		},
		{
			compared: Labels{
				{
					Name:  "aaa",
					Value: "111",
				},
				{
					Name:  "bbb",
					Value: "233",
				},
			},
			expected: -1,
		},
		{
			compared: Labels{
				{
					Name:  "aaa",
					Value: "111",
				},
				{
					Name:  "bar",
					Value: "222",
				},
			},
			expected: 1,
		},
		{
			compared: Labels{
				{
					Name:  "aaa",
					Value: "111",
				},
				{
					Name:  "bbc",
					Value: "222",
				},
			},
			expected: -1,
		},
		{
			compared: Labels{
				{
					Name:  "aaa",
					Value: "111",
				},
			},
			expected: 1,
		},
		{
			compared: Labels{
				{
					Name:  "aaa",
					Value: "111",
				},
				{
					Name:  "bbb",
					Value: "222",
				},
				{
					Name:  "ccc",
					Value: "333",
				},
				{
					Name:  "ddd",
					Value: "444",
				},
			},
			expected: -2,
		},
		{
			compared: Labels{
				{
					Name:  "aaa",
					Value: "111",
				},
				{
					Name:  "bbb",
					Value: "222",
				},
			},
			expected: 0,
		},
	}

	for i, test := range tests {
		got := Compare(labels, test.compared)
		require.Equal(t, test.expected, got, "unexpected comparison result for test case %d", i)
	}
}

func TestLabelsHash(t *testing.T) {
	lbls := Labels{
		{Name: "foo", Value: "bar"},
		{Name: "baz", Value: "qux"},
	}
	require.Equal(t, lbls.Hash(), lbls.Hash())
	require.NotEqual(t, lbls.Hash(), Labels{lbls[1], lbls[0]}.Hash(), "unordered labels match.")
	require.NotEqual(t, lbls.Hash(), Labels{lbls[0]}.Hash(), "different labels match.")
}

func TestLabelsMap(t *testing.T) {
	lbls := Labels{
		{Name: "foo", Value: "bar"},
		{Name: "baz", Value: "qux"},
	}
	expectedMap := make(map[string]string, len(lbls))
	expectedMap["foo"] = "bar"
	expectedMap["baz"] = "qux"

	require.Equal(t, lbls.Map(), lbls.Map())
	require.Equal(t, lbls.Map(), expectedMap)
	require.Equal(t, lbls.Map(), Labels{lbls[1], lbls[0]}.Map())
	require.NotEqual(t, lbls.Map(), Labels{lbls[0]}.Map(), "different labels match.")
}

var benchmarkLabelsResult uint64

func BenchmarkLabelsHash(b *testing.B) {
	for _, tcase := range []struct {
		name string
		lbls Labels
	}{
		{
			name: "typical labels under 1KB",
			lbls: func() Labels {
				lbls := make(Labels, 10)
				for i := 0; i < len(lbls); i++ {
					lbls[i] = &Label{Name: fmt.Sprintf("abcdefghijabcdefghijabcdefghij%d", i), Value: fmt.Sprintf("abcdefghijabcdefghijabcdefghijabcdefghijabcdefghij%d", i)}
				}
				return lbls
			}(),
		},
		{
			name: "bigger labels over 1KB",
			lbls: func() Labels {
				lbls := make(Labels, 10)
				for i := 0; i < len(lbls); i++ {
					lbls[i] = &Label{Name: fmt.Sprintf("abcdefghijabcdefghijabcdefghijabcdefghijabcdefghij%d", i), Value: fmt.Sprintf("abcdefghijabcdefghijabcdefghijabcdefghijabcdefghij%d", i)}
				}
				return lbls
			}(),
		},
		{
			name: "extremely large label value 10MB",
			lbls: func() Labels {
				lbl := &strings.Builder{}
				lbl.Grow(1024 * 1024 * 10)
				word := "abcdefghij"
				for i := 0; i < lbl.Cap()/len(word); i++ {
					_, _ = lbl.WriteString(word)
				}
				return Labels{{Name: "__name__", Value: lbl.String()}}
			}(),
		},
	} {
		b.Run(tcase.name, func(b *testing.B) {
			var h uint64

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				h = tcase.lbls.Hash()
			}
			benchmarkLabelsResult = h
		})
	}
}

func TestLabelsNew(t *testing.T) {
	Convey("test labels New ", t, func() {
		labelArr := []*Label{
			{
				Name:  "aaa",
				Value: "111",
			},
			{
				Name:  "bbb",
				Value: "222",
			},
		}

		expected := Labels{
			{
				Name:  "aaa",
				Value: "111",
			},
			{
				Name:  "bbb",
				Value: "222",
			},
		}

		actual := New(labelArr...)
		So(actual.Len(), ShouldEqual, 2)
		So(actual, ShouldResemble, expected)
	})
}

func TestBuliderNewBulider(t *testing.T) {
	require.Equal(
		t,
		&Builder{
			base: Labels{{"aaa", "111"}},
			del:  []string{},
			add:  []*Label{},
		},
		NewBuilder(Labels{{"aaa", "111"}}),
	)
}

func TestBuilderDel(t *testing.T) {
	require.Equal(
		t,
		&Builder{
			del: []string{"bbb"},
			add: []*Label{{"aaa", "111"}, {"ccc", "333"}},
		},
		(&Builder{
			del: []string{},
			add: []*Label{{"aaa", "111"}, {"bbb", "222"}, {"ccc", "333"}},
		}).Del("bbb"),
	)
}

func TestBuilderKeep(t *testing.T) {
	require.Equal(
		t,
		&Builder{
			base: Labels{{"aaa", "111"}, {"bbb", "222"}, {"ccc", "333"}},
			del:  []string{"aaa", "ccc"},
			add:  []*Label{},
		},
		(&Builder{
			base: Labels{{"aaa", "111"}, {"bbb", "222"}, {"ccc", "333"}},
			del:  []string{},
			add:  []*Label{},
		}).Keep("bbb"),
	)
}

func TestBuilderSet(t *testing.T) {
	require.Equal(
		t,
		&Builder{
			base: Labels{{"aaa", "111"}, {"bbb", "222"}, {"ccc", "333"}},
			del:  []string{"bbb"},
			add:  []*Label{},
		},
		(&Builder{
			base: Labels{{"aaa", "111"}, {"bbb", "222"}, {"ccc", "333"}},
			del:  []string{},
			add:  []*Label{},
		}).Set("bbb", ""),
	)

	require.Equal(
		t,
		&Builder{
			base: Labels{{"aaa", "111"}, {"bbb", "222"}, {"ccc", "333"}},
			del:  []string{},
			add:  []*Label{{"bbb", "333"}},
		},
		(&Builder{
			base: Labels{{"aaa", "111"}, {"bbb", "222"}, {"ccc", "333"}},
			del:  []string{},
			add:  []*Label{{"bbb", "222"}},
		}).Set("bbb", "333"),
	)

	require.Equal(
		t,
		&Builder{
			base: Labels{{"aaa", "111"}, {"bbb", "222"}, {"ccc", "333"}},
			del:  []string{},
			add:  []*Label{{"ddd", "444"}},
		},
		(&Builder{
			base: Labels{{"aaa", "111"}, {"bbb", "222"}, {"ccc", "333"}},
			del:  []string{},
			add:  []*Label{{"ddd", "444"}},
		}).Set("ddd", "444"),
	)
}

func TestBuilderLabels(t *testing.T) {
	require.Equal(
		t,
		Labels{{"aaa", "111"}, {"ccc", "333"}, {"ddd", "444"}},
		(&Builder{
			base: Labels{{"aaa", "111"}, {"bbb", "222"}, {"ccc", "333"}},
			del:  []string{"bbb"},
			add:  []*Label{{"ddd", "444"}},
		}).Labels(),
	)
}

func TestLabelsWithLabels(t *testing.T) {
	require.Equal(t, Labels{{"aaa", "111"}, {"bbb", "222"}}, Labels{{"aaa", "111"}, {"bbb", "222"}, {"ccc", "333"}}.WithLabels("aaa", "bbb"))
}

func TestLabelsWithoutLabels(t *testing.T) {
	require.Equal(t, Labels{{"aaa", "111"}}, Labels{{"aaa", "111"}, {"bbb", "222"}, {"ccc", "333"}}.WithoutLabels("bbb", "ccc"))
	require.Equal(t, Labels{{"aaa", "111"}}, Labels{{"aaa", "111"}, {"bbb", "222"}, {MetricName, "333"}}.WithoutLabels("bbb"))
}

func TestLabelsMatchLabels(t *testing.T) {
	labels := Labels{
		{
			Name:  "__name__",
			Value: "ALERTS",
		},
		{
			Name:  "alertname",
			Value: "HTTPRequestRateLow",
		},
		{
			Name:  "alertstate",
			Value: "pending",
		},
		{
			Name:  "instance",
			Value: "0",
		},
		{
			Name:  "job",
			Value: "app-server",
		},
		{
			Name:  "severity",
			Value: "critical",
		},
	}

	tests := []struct {
		providedNames []string
		on            bool
		expected      Labels
	}{
		// on = true, explicitly including metric name in matching.
		{
			providedNames: []string{
				"__name__",
				"alertname",
				"alertstate",
				"instance",
			},
			on: true,
			expected: Labels{
				{
					Name:  "__name__",
					Value: "ALERTS",
				},
				{
					Name:  "alertname",
					Value: "HTTPRequestRateLow",
				},
				{
					Name:  "alertstate",
					Value: "pending",
				},
				{
					Name:  "instance",
					Value: "0",
				},
			},
		},
		// on = false, explicitly excluding metric name from matching.
		{
			providedNames: []string{
				"__name__",
				"alertname",
				"alertstate",
				"instance",
			},
			on: false,
			expected: Labels{
				{
					Name:  "job",
					Value: "app-server",
				},
				{
					Name:  "severity",
					Value: "critical",
				},
			},
		},
		// on = true, explicitly excluding metric name from matching.
		{
			providedNames: []string{
				"alertname",
				"alertstate",
				"instance",
			},
			on: true,
			expected: Labels{
				{
					Name:  "alertname",
					Value: "HTTPRequestRateLow",
				},
				{
					Name:  "alertstate",
					Value: "pending",
				},
				{
					Name:  "instance",
					Value: "0",
				},
			},
		},
		// on = false, implicitly excluding metric name from matching.
		{
			providedNames: []string{
				"alertname",
				"alertstate",
				"instance",
			},
			on: false,
			expected: Labels{
				{
					Name:  "job",
					Value: "app-server",
				},
				{
					Name:  "severity",
					Value: "critical",
				},
			},
		},
	}

	for i, test := range tests {
		got := labels.MatchLabels(test.on, test.providedNames...)
		require.Equal(t, test.expected, got, "unexpected labelset for test case %d", i)
	}
}

func TestLabelsGet(t *testing.T) {
	require.Equal(t, "", Labels{{"aaa", "111"}, {"bbb", "222"}}.Get("foo"))
	require.Equal(t, "111", Labels{{"aaa", "111"}, {"bbb", "222"}}.Get("aaa"))
}

func TestLabelsHashForLabels(t *testing.T) {
	Convey("test hashforlables", t, func() {
		labels := &Labels{{"aaa", "111"}, {"ccc", "333"}, {"ddd", "444"}}
		Convey("labels containes one of the names", func() {
			hash, buf := labels.HashForLabels(nil, "aaa")

			So(hash, ShouldEqual, uint64(17396288654302788568))
			So(buf, ShouldResemble, []byte{97, 97, 97, 255, 49, 49, 49, 255})
		})
		Convey("labels does not containes any of the names", func() {
			hash, buf := labels.HashForLabels(nil, "bbb")

			So(hash, ShouldEqual, uint64(17241709254077376921))
			So(buf, ShouldBeNil)
		})
	})
}

func TestLableHashWithoutLables(t *testing.T) {
	Convey("test hashwithoutlables", t, func() {
		labels := &Labels{{"aaa", "111"}, {"ccc", "333"}, {"ddd", "444"}}
		Convey("labels containes one of the names", func() {
			hash, buf := labels.HashWithoutLabels(nil, "aaa")

			So(hash, ShouldEqual, uint64(4139075327986457141))
			So(buf, ShouldResemble, []byte{99, 99, 99, 255, 51, 51, 51, 255, 100, 100, 100, 255, 52, 52, 52, 255})
		})
		Convey("labels does not containes any of the names", func() {
			hash, buf := labels.HashWithoutLabels(nil, "bbb")

			So(hash, ShouldEqual, uint64(15715640708263596990))
			So(buf, ShouldResemble, []byte{97, 97, 97, 255, 49, 49, 49, 255, 99, 99, 99, 255, 51, 51, 51, 255, 100, 100, 100, 255, 52, 52, 52, 255})
		})
		Convey("names equal to _name_", func() {
			hash, buf := labels.HashWithoutLabels(nil, "_name_")

			So(hash, ShouldEqual, uint64(15715640708263596990))
			So(buf, ShouldResemble, []byte{97, 97, 97, 255, 49, 49, 49, 255, 99, 99, 99, 255, 51, 51, 51, 255, 100, 100, 100, 255, 52, 52, 52, 255})
		})
	})
}

func TestLabels_String(t *testing.T) {
	cases := []struct {
		lables   Labels
		expected string
	}{
		{
			lables: Labels{
				{
					Name:  "t1",
					Value: "t1",
				},
				{
					Name:  "t2",
					Value: "t2",
				},
			},
			expected: "{t1=\"t1\", t2=\"t2\"}",
		},
		{
			lables:   Labels{},
			expected: "{}",
		},
		{
			lables:   nil,
			expected: "{}",
		},
	}
	for _, c := range cases {
		str := c.lables.String()
		require.Equal(t, c.expected, str)
	}
}

func TestIsValid(t *testing.T) {
	Convey("test IsValid", t, func() {
		Convey("label contains only a single character", func() {
			So(LabelName("a").IsValid(), ShouldBeTrue)
			So(LabelName("A").IsValid(), ShouldBeTrue)
			So(LabelName("1").IsValid(), ShouldBeFalse)
			So(LabelName("_").IsValid(), ShouldBeTrue)
			So(LabelName(".").IsValid(), ShouldBeFalse)
			So(LabelName(",").IsValid(), ShouldBeFalse)
			So(LabelName("你").IsValid(), ShouldBeFalse)
		})

		Convey("label only contains two characters", func() {
			So(LabelName("ab").IsValid(), ShouldBeTrue)
			So(LabelName("aA").IsValid(), ShouldBeTrue)
			So(LabelName("a1").IsValid(), ShouldBeTrue)
			So(LabelName("a_").IsValid(), ShouldBeTrue)
			So(LabelName("a.").IsValid(), ShouldBeFalse)
			So(LabelName("a,").IsValid(), ShouldBeFalse)
			So(LabelName("a你").IsValid(), ShouldBeFalse)

			So(LabelName("Ab").IsValid(), ShouldBeTrue)
			So(LabelName("AA").IsValid(), ShouldBeTrue)
			So(LabelName("A1").IsValid(), ShouldBeTrue)
			So(LabelName("A_").IsValid(), ShouldBeTrue)
			So(LabelName("A.").IsValid(), ShouldBeFalse)
			So(LabelName("A,").IsValid(), ShouldBeFalse)
			So(LabelName("A你").IsValid(), ShouldBeFalse)

			So(LabelName("_b").IsValid(), ShouldBeTrue)
			So(LabelName("_A").IsValid(), ShouldBeTrue)
			So(LabelName("_1").IsValid(), ShouldBeTrue)
			So(LabelName("__").IsValid(), ShouldBeTrue)
			So(LabelName("_.").IsValid(), ShouldBeFalse)
			So(LabelName("_,").IsValid(), ShouldBeFalse)
			So(LabelName("_你").IsValid(), ShouldBeFalse)

			So(LabelName("1b").IsValid(), ShouldBeFalse)
			So(LabelName("1A").IsValid(), ShouldBeFalse)
			So(LabelName("11").IsValid(), ShouldBeFalse)
			So(LabelName("1_").IsValid(), ShouldBeFalse)
			So(LabelName("1.").IsValid(), ShouldBeFalse)
			So(LabelName("1,").IsValid(), ShouldBeFalse)
			So(LabelName("1你").IsValid(), ShouldBeFalse)

			So(LabelName(".b").IsValid(), ShouldBeFalse)
			So(LabelName(".A").IsValid(), ShouldBeFalse)
			So(LabelName(".1").IsValid(), ShouldBeFalse)
			So(LabelName("._").IsValid(), ShouldBeFalse)
			So(LabelName("..").IsValid(), ShouldBeFalse)
			So(LabelName(".,").IsValid(), ShouldBeFalse)
			So(LabelName(".你").IsValid(), ShouldBeFalse)

			So(LabelName(",b").IsValid(), ShouldBeFalse)
			So(LabelName(",A").IsValid(), ShouldBeFalse)
			So(LabelName(",1").IsValid(), ShouldBeFalse)
			So(LabelName(",_").IsValid(), ShouldBeFalse)
			So(LabelName(",.").IsValid(), ShouldBeFalse)
			So(LabelName(",,").IsValid(), ShouldBeFalse)
			So(LabelName(",你").IsValid(), ShouldBeFalse)

			So(LabelName("你b").IsValid(), ShouldBeFalse)
			So(LabelName("你A").IsValid(), ShouldBeFalse)
			So(LabelName("你1").IsValid(), ShouldBeFalse)
			So(LabelName("你_").IsValid(), ShouldBeFalse)
			So(LabelName("你.").IsValid(), ShouldBeFalse)
			So(LabelName("你,").IsValid(), ShouldBeFalse)
			So(LabelName("你你").IsValid(), ShouldBeFalse)
		})

		Convey("label has two sections", func() {
			So(LabelName("a.a").IsValid(), ShouldBeTrue)
			So(LabelName("a.A").IsValid(), ShouldBeTrue)
			So(LabelName("a.1").IsValid(), ShouldBeFalse)
			So(LabelName("a._").IsValid(), ShouldBeTrue)
			So(LabelName("a..").IsValid(), ShouldBeFalse)
			So(LabelName("a.,").IsValid(), ShouldBeFalse)
			So(LabelName("a.你").IsValid(), ShouldBeFalse)

			So(LabelName("A.a").IsValid(), ShouldBeTrue)
			So(LabelName("A.A").IsValid(), ShouldBeTrue)
			So(LabelName("A.1").IsValid(), ShouldBeFalse)
			So(LabelName("A._").IsValid(), ShouldBeTrue)
			So(LabelName("A..").IsValid(), ShouldBeFalse)
			So(LabelName("A.,").IsValid(), ShouldBeFalse)
			So(LabelName("A.你").IsValid(), ShouldBeFalse)

			So(LabelName("_.a").IsValid(), ShouldBeTrue)
			So(LabelName("_.A").IsValid(), ShouldBeTrue)
			So(LabelName("_.1").IsValid(), ShouldBeFalse)
			So(LabelName("_._").IsValid(), ShouldBeTrue)
			So(LabelName("_..").IsValid(), ShouldBeFalse)
			So(LabelName("_.,").IsValid(), ShouldBeFalse)
			So(LabelName("_.你").IsValid(), ShouldBeFalse)
		})

		Convey("label has multiple segments", func() {
			So(LabelName("a.b.1").IsValid(), ShouldBeFalse)
			So(LabelName("a.b.1._").IsValid(), ShouldBeFalse)
			So(LabelName("a.b.1._.A").IsValid(), ShouldBeFalse)
			So(LabelName("a.你._.1l").IsValid(), ShouldBeFalse)
			So(LabelName(".a.你._.1l").IsValid(), ShouldBeFalse)
			So(LabelName("a.b.._.l").IsValid(), ShouldBeFalse)
			So(LabelName("a.b...._.l").IsValid(), ShouldBeFalse)
			So(LabelName("a.b..,._.l").IsValid(), ShouldBeFalse)
			So(LabelName("a.b..你._.l").IsValid(), ShouldBeFalse)
			So(LabelName("a.b..ab._.l").IsValid(), ShouldBeFalse)
			So(LabelName("_a._.b").IsValid(), ShouldBeTrue)
		})
	})

}
