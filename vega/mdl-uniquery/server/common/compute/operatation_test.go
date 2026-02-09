// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package compute

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLt(t *testing.T) {
	Convey("Test Lt", t, func() {
		Convey(" Lt interface", func() {
			res := Lt("hello", []interface{}{"hello", "world"})
			So(res, ShouldBeFalse)
		})
		Convey(" Lt float64", func() {
			res := Lt(float64(1.0), float64(2.0))
			So(res, ShouldBeTrue)
		})
		Convey(" Lt float32", func() {
			res := Lt(float32(1.0), float32(2.0))
			So(res, ShouldBeTrue)
		})
		Convey(" Lt Int32", func() {
			res := Lt(int32(1), int32(2))
			So(res, ShouldBeTrue)
		})
		Convey(" Lt string", func() {
			res := Lt("1", "2")
			So(res, ShouldBeTrue)
		})
	})
}

func TestOp_in(t *testing.T) {
	Convey("Test Op_in", t, func() {
		Convey(" Op_in return true", func() {
			res := Op_in("hello", []interface{}{"hello", "world"})
			So(res, ShouldBeTrue)
		})
	})
}

func TestOp_contain(t *testing.T) {
	Convey("Test Op_contain", t, func() {
		Convey(" Op_contain return true", func() {
			res := Op_contain([]interface{}{"hello"}, []interface{}{"hello", "world"})
			So(res, ShouldBeTrue)
		})
		Convey(" Op_contain return false", func() {
			res := Op_contain([]interface{}{"hello1"}, []interface{}{"hello", "world"})
			So(res, ShouldBeFalse)
		})
	})
}

func TestEqual(t *testing.T) {
	Convey("Test Equal", t, func() {
		Convey(" Equal return true", func() {
			res := Equal("hello", "llo")
			So(res, ShouldBeFalse)
		})
	})
}

func TestLike(t *testing.T) {
	Convey("Test Like", t, func() {
		Convey(" Like return true", func() {
			res := Like("hello", "llo")
			So(res, ShouldBeTrue)
		})
		Convey(" value is not string ", func() {
			res := Like("hello", 1)
			So(res, ShouldBeFalse)
		})
	})
}

func TestProcessSingleValue(t *testing.T) {
	Convey("Test processSingleValue", t, func() {
		Convey(" processSingleValue case like", func() {
			res := processSingleValue("like", "hello", []interface{}{"hello", "world"})
			So(res, ShouldBeFalse)
		})
		Convey(" processSingleValue case not like", func() {
			res := processSingleValue("not_like", "hello", []interface{}{"hello1", "world"})
			So(res, ShouldBeTrue)
		})
		Convey(" processSingleValue case ==", func() {
			res := processSingleValue("==", "1", "1")
			So(res, ShouldBeTrue)
		})
		Convey(" processSingleValue case !=", func() {
			res := processSingleValue("!=", "1", "2")
			So(res, ShouldBeTrue)
		})
		Convey(" processSingleValue case >", func() {
			res := processSingleValue(">", "3", "2")
			So(res, ShouldBeTrue)
		})
		Convey(" processSingleValue case >=", func() {
			res := processSingleValue(">=", "3", "2")
			So(res, ShouldBeTrue)
		})
		Convey(" processSingleValue case <", func() {
			res := processSingleValue("<", "1", "2")
			So(res, ShouldBeTrue)
		})
		Convey(" processSingleValue case <=", func() {
			res := processSingleValue("<=", "1", "2")
			So(res, ShouldBeTrue)
		})
	})
}

func TestProcessMultiple(t *testing.T) {
	Convey("Test processMultiple", t, func() {
		Convey(" processMultiple case range", func() {
			res := processMultiple("range", 1, []interface{}{1, 3})
			So(res, ShouldBeFalse)
		})
		Convey(" processMultiple case not range", func() {
			res := processMultiple("not_range", 1, []interface{}{2, 3})
			So(res, ShouldBeFalse)
		})
		Convey(" processMultiple case in", func() {
			res := processMultiple("in", 1, []interface{}{1, 2})
			So(res, ShouldBeTrue)
		})
		Convey(" processMultiple case contain", func() {
			res := processMultiple("contain", []interface{}{1, 2}, []interface{}{1})
			So(res, ShouldBeTrue)
		})
	})
}

func TestExec(t *testing.T) {
	Convey("Test Exec", t, func() {
		Convey(" Exec case range", func() {
			res := Exec("range", 1, []interface{}{1, 3})
			So(res, ShouldBeFalse)
		})
		Convey(" Exec case not range", func() {
			res := Exec("not range", 1, []interface{}{2, 3})
			So(res, ShouldBeFalse)
		})
		Convey(" Exec case in", func() {
			res := Exec("in", 1, []interface{}{1, 2})
			So(res, ShouldBeTrue)
		})
		Convey(" Exec case ==", func() {
			res := Exec("==", "1", "1")
			So(res, ShouldBeTrue)
		})
		Convey(" Exec value nil", func() {
			res := Exec("range", nil, []interface{}{1, 3})
			So(res, ShouldBeFalse)
		})
	})
}
func TestIn(t *testing.T) {
	Convey("Test In", t, func() {
		Convey(" In ", func() {
			res, err := In("hello", []string{"", "hello"})
			So(res, ShouldBeTrue)
			So(err, ShouldBeNil)
		})
	})
}
func TestContain(t *testing.T) {
	Convey("Test Contain", t, func() {
		Convey(" Contain ", func() {
			res, err := Contain([]string{"hello"}, []string{"hello", "world"})
			So(res, ShouldBeTrue)
			So(err, ShouldBeNil)
		})
		Convey(" Contain ,empty", func() {
			res, err := Contain([]string{"hello"}, []string{})
			So(res, ShouldBeFalse)
			So(err, ShouldBeNil)
		})
	})
}

func TestEqualTo(t *testing.T) {
	Convey("Test EqualTo", t, func() {
		Convey(" EqualTo ", func() {
			res, err := EqualTo(1, 1)
			So(res, ShouldBeTrue)
			So(err, ShouldBeNil)
		})
	})
}
