// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package common

import (
	"errors"
	"testing"
	"time"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/bytedance/sonic"
	. "github.com/smartystreets/goconvey/convey"
)

func Test_Convert_Map2String(t *testing.T) {
	Convey("Test Map2String", t, func() {
		Convey("empty", func() {
			actual, err := Map2String(nil)
			So(actual, ShouldEqual, "")
			So(err, ShouldBeNil)
		})

		Convey("marshal error", func() {
			patch := ApplyFuncReturn(sonic.Marshal, nil, errors.New("error"))
			defer patch.Reset()

			actual, err := Map2String(map[string]string{})
			So(actual, ShouldEqual, "")
			So(err.Error(), ShouldEqual, "error")
		})

		Convey("success", func() {

			actual, err := Map2String(map[string]string{"a": "b"})
			So(actual, ShouldResemble, `{"a":"b"}`)
			So(err, ShouldBeNil)
		})
	})
}

func Test_Convert_String2Map(t *testing.T) {
	Convey("Test String2Map", t, func() {
		var expect map[string]string

		Convey("empty", func() {
			actual, err := String2Map("")
			So(actual, ShouldResemble, expect)
			So(err, ShouldBeNil)
		})

		Convey("marshal error", func() {
			patch := ApplyFuncReturn(sonic.Unmarshal, errors.New("error"))
			defer patch.Reset()

			actual, err := String2Map("nil")
			So(actual, ShouldEqual, expect)
			So(err.Error(), ShouldEqual, "error")
		})

		Convey("success", func() {
			actual, err := String2Map(`{"a":"b"}`)
			So(actual, ShouldResemble, map[string]string{"a": "b"})
			So(err, ShouldBeNil)
		})
	})
}

func Test_Convert_ParseDuration(t *testing.T) {
	var tests = []struct {
		input  string
		fail   bool
		result time.Duration
	}{
		{
			input: "",
			fail:  true,
		}, {
			input: "abc",
			fail:  true,
		}, {
			input: "2015-06-03T13:21:58.555Z",
			fail:  true,
		}, {
			input: "-148966367200.372",
			fail:  true,
		}, {
			input: "148966367200.372",
			fail:  true,
		}, {
			input:  "123",
			result: 123 * time.Second,
		}, {
			input:  "123.333",
			result: 123*time.Second + 333*time.Millisecond,
		}, {
			input:  "5m",
			result: 5 * time.Minute,
		}, {
			input: "5s",
			fail:  true,
		}, {
			input:  "1h",
			result: 1 * time.Hour,
		}, {
			input:  "1d",
			result: 24 * time.Hour,
		},
	}

	for _, test := range tests {
		d, err := ParseDuration(test.input, DurationDayHourMinuteRE, true)
		if err != nil && !test.fail {
			t.Errorf("Unexpected error for %q: %s", test.input, err)
			continue
		}
		if err == nil && test.fail {
			t.Errorf("Expected error for %q but got none", test.input)
			continue
		}
		if !test.fail && d != test.result {
			t.Errorf("Expected duration %v for input %q but got %v", test.result, test.input, d)
		}
	}

	var tests2 = []struct {
		input  string
		fail   bool
		result time.Duration
	}{
		{
			input: "",
			fail:  true,
		}, {
			input: "abc",
			fail:  true,
		}, {
			input: "2015-06-03T13:21:58.555Z",
			fail:  true,
		}, {
			input: "-148966367200.372",
			fail:  true,
		}, {
			input: "148966367200.372",
			fail:  true,
		}, {
			input:  "123",
			result: 123 * time.Second,
		}, {
			input:  "123.333",
			result: 123*time.Second + 333*time.Millisecond,
		}, {
			input: "5m",
			fail:  true,
		}, {
			input: "5s",
			fail:  true,
		}, {
			input:  "1h",
			result: 1 * time.Hour,
		}, {
			input:  "1d",
			result: 24 * time.Hour,
		},
	}

	for _, test := range tests2 {
		d, err := ParseDuration(test.input, DurationDayHourRE, false)
		if err != nil && !test.fail {
			t.Errorf("Unexpected error for %q: %s", test.input, err)
			continue
		}
		if err == nil && test.fail {
			t.Errorf("Expected error for %q but got none", test.input)
			continue
		}
		if !test.fail && d != test.result {
			t.Errorf("Expected duration %v for input %q but got %v", test.result, test.input, d)
		}
	}
}
