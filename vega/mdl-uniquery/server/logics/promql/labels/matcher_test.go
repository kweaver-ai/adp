// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package labels

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/require"
)

func mustNewMatcher(t *testing.T, mType MatchType, value string) *Matcher {
	m, err := NewMatcher(mType, "", value)
	require.NoError(t, err)
	return m
}

func TestMatcher(t *testing.T) {
	Convey("test matcher ", t, func() {
		tests := []struct {
			matcher *Matcher
			value   string
			match   bool
		}{
			{
				matcher: mustNewMatcher(t, MatchEqual, "bar"),
				value:   "bar",
				match:   true,
			},
			{
				matcher: mustNewMatcher(t, MatchEqual, "bar"),
				value:   "foo-bar",
				match:   false,
			},
			{
				matcher: mustNewMatcher(t, MatchNotEqual, "bar"),
				value:   "bar",
				match:   false,
			},
			{
				matcher: mustNewMatcher(t, MatchNotEqual, "bar"),
				value:   "foo-bar",
				match:   true,
			},
			{
				matcher: mustNewMatcher(t, MatchRegexp, "bar"),
				value:   "bar",
				match:   true,
			},
			{
				matcher: mustNewMatcher(t, MatchRegexp, "bar"),
				value:   "foo-bar",
				match:   false,
			},
			{
				matcher: mustNewMatcher(t, MatchRegexp, ".*bar"),
				value:   "foo-bar",
				match:   true,
			},
			{
				matcher: mustNewMatcher(t, MatchNotRegexp, "bar"),
				value:   "bar",
				match:   false,
			},
			{
				matcher: mustNewMatcher(t, MatchNotRegexp, "bar"),
				value:   "foo-bar",
				match:   true,
			},
			{
				matcher: mustNewMatcher(t, MatchNotRegexp, ".*bar"),
				value:   "foo-bar",
				match:   false,
			},
		}

		for _, test := range tests {
			So(test.matcher.Matches(test.value), ShouldEqual, test.match)
		}
	})
}

func TestMatcherTypeString(t *testing.T) {
	Convey("Test MatcherType String ", t, func() {
		eq := MatchEqual.String()
		So(eq, ShouldEqual, "=")

		noteq := MatchNotEqual.String()
		So(noteq, ShouldEqual, "!=")

		regeq := MatchRegexp.String()
		So(regeq, ShouldEqual, "=~")

		notregeq := MatchNotRegexp.String()
		So(notregeq, ShouldEqual, "!~")
	})
}

func TestMatcherString(t *testing.T) {
	Convey("Test MatcherType String ", t, func() {
		meq, err := NewMatcher(MatchEqual, "cluster", "txy")
		So(err, ShouldBeNil)
		So(meq.String(), ShouldEqual, `cluster="txy"`)

		mnoteq, err := NewMatcher(MatchNotEqual, "cluster", "txy")
		So(err, ShouldBeNil)
		So(mnoteq.String(), ShouldEqual, `cluster!="txy"`)

		mreq, err := NewMatcher(MatchRegexp, "cluster", "txy")
		So(err, ShouldBeNil)
		So(mreq.String(), ShouldEqual, `cluster=~"txy"`)

		mnotreq, err := NewMatcher(MatchNotRegexp, "cluster", "txy")
		So(err, ShouldBeNil)
		So(mnotreq.String(), ShouldEqual, `cluster!~"txy"`)
	})
}
