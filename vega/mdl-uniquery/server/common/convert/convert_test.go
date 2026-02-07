// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package convert

import (
	"fmt"
	"math"
	"os"
	"strings"
	"testing"
	"time"
	"uniquery/common"

	"github.com/bytedance/sonic"
	libCommon "github.com/kweaver-ai/kweaver-go-lib/common"
	. "github.com/smartystreets/goconvey/convey"
)

func TestIntToDuration(t *testing.T) {
	Convey("Test int str to time.duration", t, func() {
		Convey("1m", func() {
			in := "1m"
			actual, _ := IntToDuration(in)
			So(actual, ShouldEqual, time.Duration(1)*time.Minute)
		})
		Convey("30s", func() {
			in := "30s"
			actual, _ := IntToDuration(in)
			So(actual, ShouldEqual, time.Duration(30)*time.Second)
		})
		Convey("1h", func() {
			in := "1h"
			actual, _ := IntToDuration(in)
			So(actual, ShouldEqual, time.Duration(1)*time.Hour)
		})
		Convey("1", func() {
			in := "1"
			actual, err := IntToDuration(in)
			So(actual, ShouldEqual, 0)
			So(err, ShouldNotBeNil)
		})
		Convey("1y", func() {
			in := "1y"
			actual, err := IntToDuration(in)
			So(actual, ShouldEqual, 0)
			So(err, ShouldNotBeNil)
		})
		Convey("65534s", func() {
			in := "65534s"
			actual, err := IntToDuration(in)
			So(actual, ShouldEqual, time.Duration(65534)*time.Second)
			So(err, ShouldBeNil)
		})
		Convey("2mmmm", func() {
			in := "2mmmm"
			actual, err := IntToDuration(in)
			So(actual, ShouldEqual, 0)
			So(err, ShouldNotBeNil)
		})
	})

}

func TestJsonToMap(t *testing.T) {
	Convey("Test json to map", t, func() {
		Convey("json str", func() {
			jsonStr := "{\n\"aggs\": {\n\"set_count\": {\n" +
				"\"cardinality\": {\n\"script\": {\n" +
				"\"inline\": \"params['_source']['city']['name'];\"\n" +
				"}\n}\n},\n\"\\u57ce\\u5e02\": {\n" +
				"\"terms\": {\n\"script\": {\n" +
				"\"inline\": \"params['_source']['city']['name'];\"\n},\n " +
				"\"shard_size\": 100,\n\"size\": 10\n}\n " +
				"}\n},\n\"query\": {\n\"bool\": {\n\"filter\": [],\n" +
				"\"must\": [\n{\n" +
				"\"range\": {\n\"@timestamp\": {\n" +
				" \"format\": \"epoch_millis||yyyy-MM-dd HH:mm:ss\",\n" +
				"\"gte\": 1580200956673,\n" +
				"\"lt\": 1611823356673\n" +
				" }\n}\n                " +
				"},\n{ \"bool\": {\n\"should\": [\n" +
				"{\n\"match\": {\n " +
				"\"_index\": \"test_script-0\"\n      " +
				"}\n}\n" +
				"]\n}\n}\n],\n\"must_not\": [],\n" +
				"\"should\": []\n    }\n},\n\"size\": 0\n}"
			_, err := JsonToMap(jsonStr)
			So(err, ShouldBeNil)
		})
		Convey("not json map", func() {
			jsonStr := "{ 'a':'b'}"
			_, err := JsonToMap(jsonStr)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestInterToArray(t *testing.T) {
	Convey("Test interToArray:", t, func() {
		Convey("correct array", func() {
			m := map[string]interface{}{
				"alias": []string{"a", "b", "c"},
			}
			jsondata, _ := sonic.Marshal(m)
			jsonStr := strings.NewReader(string(jsondata))
			var mapJson map[string]interface{}
			if err := sonic.ConfigDefault.NewDecoder(jsonStr).Decode(&mapJson); err != nil {
				t.Errorf("Error parsing the response body: %s", err)
			}
			alias := mapJson["alias"]

			array := InterToArray(alias)
			So(array, ShouldResemble, m["alias"])
		})
		Convey("incorrect array", func() {
			m := map[string]interface{}{
				"alias": "a,b,c",
			}
			jsondata, _ := sonic.Marshal(m)
			jsonStr := strings.NewReader(string(jsondata))
			var mapJson map[string]interface{}
			if err := sonic.ConfigDefault.NewDecoder(jsonStr).Decode(&mapJson); err != nil {
				t.Errorf("Error parsing the response body: %s", err)
			}
			alias := mapJson["alias"]
			array := InterToArray(alias)
			So(array, ShouldBeNil)
		})
		Convey("empty array", func() {
			m := map[string]interface{}{
				"alias": []string{},
			}
			jsondata, _ := sonic.Marshal(m)
			jsonStr := strings.NewReader(string(jsondata))
			var mapJson map[string]interface{}
			if err := sonic.ConfigDefault.NewDecoder(jsonStr).Decode(&mapJson); err != nil {
				t.Errorf("Error parsing the response body: %s", err)
			}
			alias := mapJson["alias"]
			array := InterToArray(alias)
			So(array, ShouldResemble, m["alias"])
		})
	})
}

func TestIntersectArray(t *testing.T) {
	Convey("Test intersect array", t, func() {
		Convey("test1", func() {
			slice1 := []string{"test", "ar_audit_log"}
			slice2 := []string{"kc"}
			inter := IntersectArray(slice1, slice2)
			So(inter, ShouldResemble, []string{})
		})
		Convey("test2", func() {
			slice1 := []string{"1", "2", "3", "6", "8"}
			slice2 := []string{"2", "3", "5", "0"}
			inter := IntersectArray(slice1, slice2)
			So(inter, ShouldResemble, []string{"2", "3"})
		})

	})
}

func TestStructConvertMap(t *testing.T) {
	Convey("Test struct to map", t, func() {
		type Birthday struct {
			Year  int // Year (e.g., 2014).
			Month int // Month of the year (January = 1, ...).
			Day   int // Day of the month, starting at 1.
		}
		type Person struct {
			Name     string   `json:"name"`
			Age      int      `json:"age"`
			Birthday Birthday `json:"birthday"`
		}
		p1 := &Person{
			Name:     "cassie",
			Age:      18,
			Birthday: Birthday{Year: 1996, Month: 12, Day: 25},
		}
		m1 := map[string]interface{}{
			"name":     p1.Name,
			"age":      p1.Age,
			"birthday": p1.Birthday,
		}
		data := StructConvertMap(*p1, "json")
		So(data, ShouldResemble, m1)
	})
}

func TestParseTime(t *testing.T) {
	ts, err := time.Parse(time.RFC3339Nano, "2015-06-03T13:21:58.555Z")
	if err != nil {
		panic(err)
	}

	var tests = []struct {
		input  string
		fail   bool
		result time.Time
	}{
		{
			input: "",
			fail:  true,
		}, {
			input: "abc",
			fail:  true,
		}, {
			input: "30s",
			fail:  true,
		}, {
			input:  "123",
			result: time.Unix(123, 0),
		}, {
			input:  "123.123",
			result: time.Unix(123, 123000000),
		}, {
			input:  "2015-06-03T13:21:58.555Z",
			result: ts,
		}, {
			input:  "2015-06-03T14:21:58.555+01:00",
			result: ts,
		}, {
			input:  "1543578564.705",
			result: time.Unix(1543578564, 705*1e6),
		}, {
			input:  "1543578564",
			result: time.Unix(1543578564, 0),
		},
		{
			input:  MinTime.Format(time.RFC3339Nano),
			result: MinTime,
		},
		{
			input:  MaxTime.Format(time.RFC3339Nano),
			result: MaxTime,
		},
	}

	for _, test := range tests {
		ts, err := ParseTime(test.input)
		if err != nil && !test.fail {
			t.Errorf("Unexpected error for %q: %s", test.input, err)
			continue
		}
		if err == nil && test.fail {
			t.Errorf("Expected error for %q but got none", test.input)
			continue
		}
		if !test.fail && !ts.Equal(test.result) {
			t.Errorf("Expected time %v for input %q but got %v", test.result, test.input, ts)
		}
	}
}

func TestParseDuration(t *testing.T) {
	cases := []struct {
		in  string
		out time.Duration

		expectedString string
	}{
		{
			in:             "0",
			out:            0,
			expectedString: "0s",
		}, {
			in:             "0w",
			out:            0,
			expectedString: "0s",
		}, {
			in:  "0s",
			out: 0,
		}, {
			in:  "324ms",
			out: 324 * time.Millisecond,
		}, {
			in:  "3s",
			out: 3 * time.Second,
		}, {
			in:  "5m",
			out: 5 * time.Minute,
		}, {
			in:  "1h",
			out: time.Hour,
		}, {
			in:  "4d",
			out: 4 * 24 * time.Hour,
		}, {
			in:  "4d1h",
			out: 4*24*time.Hour + time.Hour,
		}, {
			in:             "14d",
			out:            14 * 24 * time.Hour,
			expectedString: "2w",
		}, {
			in:  "3w",
			out: 3 * 7 * 24 * time.Hour,
		}, {
			in:             "3w2d1h",
			out:            3*7*24*time.Hour + 2*24*time.Hour + time.Hour,
			expectedString: "23d1h",
		}, {
			in:  "10y",
			out: 10 * 365 * 24 * time.Hour,
		},
	}

	for _, c := range cases {
		d, err := ParseDuration(c.in)
		if err != nil {
			t.Errorf("Unexpected error on input %q", c.in)
		}
		if time.Duration(d) != c.out {
			t.Errorf("Expected %v but got %v", c.out, d)
		}
	}
}

func TestParseTimeParam(t *testing.T) {
	Convey("Test ParseTimeParam", t, func() {

		Convey("val is empty ", func() {
			ts, err := ParseTimeParam("", "start", MinTime)

			So(ts, ShouldEqual, MinTime)
			So(err, ShouldBeNil)
		})

		Convey("Invalid time value ", func() {
			ts, err := ParseTimeParam("1646360670a", "start", MinTime)

			So(ts, ShouldResemble, time.Time{})
			So(err.Error(), ShouldEqual, `invalid time value for 'start': cannot parse "1646360670a" to a valid timestamp`)
		})

		Convey("success ", func() {
			ts, err := ParseTimeParam("1646360670", "start", MinTime)

			So(ts, ShouldEqual, time.Unix(1646360670, 0))
			So(err, ShouldBeNil)
		})
	})
}

func TestFromFloatSeconds(t *testing.T) {
	Convey("Test FromFloatSeconds", t, func() {
		result := FromFloatSeconds(1543578564.705)
		So(result, ShouldEqual, 1543578564705)
	})
}

func TestFromTime(t *testing.T) {
	Convey("Test FromTime", t, func() {
		result := FromTime(time.Unix(1646360670, 0))
		So(result, ShouldEqual, 1646360670000)
	})
}

func TestRFC3339ToMicroTimestamp(t *testing.T) {
	Convey("Test RFC3339ToMicroTimestamp", t, func() {
		Convey("Parse error", func() {
			expectedErr := &time.ParseError{
				Layout:     "2006-01-02T15:04:05.999999999Z07:00",
				Value:      "12334",
				LayoutElem: "-",
				ValueElem:  "4",
				Message:    "",
			}

			timeStrs := []string{"12334"}
			res, err := RFC3339ToMicroTimestamp(timeStrs)
			So(res, ShouldBeNil)
			So(err, ShouldResemble, expectedErr)
		})
		Convey("Parse success", func() {
			timeStrs := []string{"2023-02-20T18:51:58.114304Z"}
			res, err := RFC3339ToMicroTimestamp(timeStrs)
			So(res[0], ShouldEqual, int64(1676919118114304))
			So(err, ShouldBeNil)
		})
	})
}

func TestGetLookBackDelta(t *testing.T) {
	Convey("Test GetLookBackDelta", t, func() {
		Convey("GetLookBackDelta eq 0 && configDelta eq 0", func() {
			res := GetLookBackDelta(0, time.Duration(0))
			So(res, ShouldEqual, 300000)
		})

		Convey("GetLookBackDelta eq 0 && configDelta ne 0", func() {
			res := GetLookBackDelta(0, time.Duration(2*time.Minute))
			So(res, ShouldEqual, 120000)
		})

		Convey("GetLookBackDelta ne 0 ", func() {
			res := GetLookBackDelta(100000, time.Duration(2*time.Minute))
			So(res, ShouldEqual, 100000)
		})
	})
}

func TestStringToInt64Slice(t *testing.T) {
	Convey("Test StringToInt64Slice", t, func() {
		Convey("Success", func() {
			res, err := StringToInt64Slice("1,2,3,4")
			So(res, ShouldEqual, []uint64{1, 2, 3, 4})
			So(err, ShouldBeNil)

		})
		Convey("str is empty", func() {
			res, err := StringToInt64Slice("")
			So(res, ShouldEqual, []uint64{})
			So(err, ShouldBeNil)

		})
		Convey("str is invalid", func() {
			res, err := StringToInt64Slice("a,bc")
			So(res, ShouldEqual, []uint64{})
			So(err, ShouldEqual, fmt.Errorf("[]string [a bc] to []uint64 parse failed"))
		})
	})
}

func TestStringSliceToInt64Slice(t *testing.T) {
	Convey("Test StringSliceToInt64Slice", t, func() {
		Convey("Success", func() {
			res, err := StringSliceToInt64Slice([]string{"1", "2"})
			So(res, ShouldEqual, []uint64{1, 2})
			So(err, ShouldBeNil)

		})
		Convey("str is empty", func() {
			res, err := StringSliceToInt64Slice([]string{})
			So(res, ShouldEqual, []uint64{})
			So(err, ShouldBeNil)

		})
		Convey("str is invalid", func() {
			res, err := StringSliceToInt64Slice([]string{"a", "bc"})
			So(res, ShouldEqual, []uint64{})
			So(err, ShouldEqual, fmt.Errorf("[]string [a bc] to []uint64 parse failed"))
		})
	})
}

func TestWrapMetricValue(t *testing.T) {
	Convey("Test WrapMetricValue", t, func() {
		Convey("Success", func() {
			res := WrapMetricValue(1.0)
			So(res, ShouldEqual, float64(1.0))

		})
		Convey("+inf", func() {
			res := WrapMetricValue(math.Inf(1))
			So(res, ShouldEqual, "+Inf")

		})
		Convey("-inf", func() {
			res := WrapMetricValue(math.Inf(-1))
			So(res, ShouldEqual, "-Inf")

		})
		Convey("NAN", func() {
			res := WrapMetricValue(math.NaN())
			So(res, ShouldEqual, "NaN")

		})
	})
}

func TestParseTimeToMillis(t *testing.T) {
	Convey("Test ParseTimeToMillis", t, func() {

		loc, _ := time.LoadLocation(os.Getenv("TZ"))
		common.APP_LOCATION = loc
		Convey("minute format", func() {
			ts, err := ParseTimeToMillis("2024-06-01 12:34", "minute")
			So(err, ShouldBeNil)
			So(ts, ShouldEqual, time.Date(2024, 6, 1, 12, 34, 0, 0, time.Local).UnixMilli())
		})

		Convey("hour format", func() {
			ts, err := ParseTimeToMillis("2024-06-01 12", "hour")
			So(err, ShouldBeNil)
			So(ts, ShouldEqual, time.Date(2024, 6, 1, 12, 0, 0, 0, time.Local).UnixMilli())
		})

		Convey("day format", func() {
			ts, err := ParseTimeToMillis("2024-06-01", "day")
			So(err, ShouldBeNil)
			So(ts, ShouldEqual, time.Date(2024, 6, 1, 0, 0, 0, 0, time.Local).UnixMilli())
		})

		Convey("week format", func() {
			ts, err := ParseTimeToMillis("2024-38", "week")
			So(err, ShouldBeNil)
			// 2024-22 means 2024 year, week 22, Monday
			// Let's parse it the same way as the function does
			// expected, _ := time.ParseInLocation("2006W021", "2024W021", time.Local)
			So(ts, ShouldEqual, 1726416000000)
		})

		Convey("month format", func() {
			ts, err := ParseTimeToMillis("2024-06", "month")
			So(err, ShouldBeNil)
			So(ts, ShouldEqual, time.Date(2024, 6, 1, 0, 0, 0, 0, time.Local).UnixMilli())
		})

		Convey("quarter format", func() {
			ts, err := ParseTimeToMillis("2024-Q2", "quarter")
			So(err, ShouldBeNil)
			So(ts, ShouldEqual, time.Date(2024, 4, 1, 0, 0, 0, 0, time.Local).UnixMilli())
		})

		Convey("invalid week format", func() {
			ts, err := ParseTimeToMillis("2024", "week")
			So(err, ShouldNotBeNil)
			So(ts, ShouldEqual, 0)
		})

		Convey("invalid quarter format", func() {
			ts, err := ParseTimeToMillis("2024-Q5", "quarter")
			So(err, ShouldNotBeNil)
			So(ts, ShouldEqual, 0)
		})

		Convey("unknown formatType", func() {
			ts, err := ParseTimeToMillis("2024-06-01", "unknown")
			So(err, ShouldNotBeNil)
			So(ts, ShouldEqual, 0)
		})

		Convey("empty string", func() {
			ts, err := ParseTimeToMillis("", "day")
			So(err, ShouldNotBeNil)
			So(ts, ShouldEqual, 0)
		})
	})
}

func TestFormatTimeMiliis(t *testing.T) {
	Convey("Test FormatTimeMiliis", t, func() {
		loc, _ := time.LoadLocation(os.Getenv("TZ"))
		common.APP_LOCATION = loc

		Convey("minute format", func() {
			ts := time.Date(2024, 6, 1, 15, 4, 0, 0, loc).UnixMilli()
			formatted := FormatTimeMiliis(ts, "minute")
			So(formatted, ShouldEqual, "2024-06-01 15:04")
		})

		Convey("hour format", func() {
			ts := time.Date(2024, 6, 1, 15, 0, 0, 0, loc).UnixMilli()
			formatted := FormatTimeMiliis(ts, "hour")
			So(formatted, ShouldEqual, "2024-06-01 15")
		})

		Convey("day format", func() {
			ts := time.Date(2024, 6, 1, 0, 0, 0, 0, loc).UnixMilli()
			formatted := FormatTimeMiliis(ts, "day")
			So(formatted, ShouldEqual, "2024-06-01")
		})

		Convey("week format", func() {
			// 2024-22nd week, Monday
			ts := time.Date(2024, 5, 27, 0, 0, 0, 0, loc).UnixMilli()
			formatted := FormatTimeMiliis(ts, "week")
			year, week := time.UnixMilli(ts).In(loc).ISOWeek()
			expected := fmt.Sprintf("%d-%02d", year, week)
			So(formatted, ShouldEqual, expected)
		})

		Convey("month format", func() {
			ts := time.Date(2024, 6, 1, 0, 0, 0, 0, loc).UnixMilli()
			formatted := FormatTimeMiliis(ts, "month")
			So(formatted, ShouldEqual, "2024-06")
		})

		Convey("quarter format", func() {
			ts := time.Date(2024, 4, 1, 0, 0, 0, 0, loc).UnixMilli()
			formatted := FormatTimeMiliis(ts, "quarter")
			So(formatted, ShouldEqual, "2024-Q2")
		})

		Convey("year format", func() {
			ts := time.Date(2024, 1, 1, 0, 0, 0, 0, loc).UnixMilli()
			formatted := FormatTimeMiliis(ts, "year")
			So(formatted, ShouldEqual, "2024")
		})

		Convey("unknown formatType", func() {
			ts := time.Date(2024, 6, 1, 0, 0, 0, 0, loc).UnixMilli()
			formatted := FormatTimeMiliis(ts, "unknown")
			// Should fallback to RFC3339Milli
			expected := time.UnixMilli(ts).In(loc).Format(libCommon.RFC3339Milli)
			So(formatted, ShouldEqual, expected)
		})
	})
}
