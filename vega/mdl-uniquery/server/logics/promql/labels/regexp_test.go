// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package labels

import (
	"errors"
	"regexp"
	"regexp/syntax"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/require"
)

func TestNewFastRegexMatcher(t *testing.T) {
	cases := []struct {
		regex    string
		value    string
		expected bool
	}{
		{regex: "(foo|bar)", value: "foo", expected: true},
		{regex: "(foo|bar)", value: "foo bar", expected: false},
		{regex: "(foo|bar)", value: "bar", expected: true},
		{regex: "foo.*", value: "foo bar", expected: true},
		{regex: "foo.*", value: "bar foo", expected: false},
		{regex: ".*foo", value: "foo bar", expected: false},
		{regex: ".*foo", value: "bar foo", expected: true},
		{regex: ".*foo", value: "foo", expected: true},
		{regex: "^.*foo$", value: "foo", expected: true},
		{regex: "^.+foo$", value: "foo", expected: false},
		{regex: "^.+foo$", value: "bfoo", expected: true},
		{regex: ".*", value: "\n", expected: false},
		{regex: ".*", value: "\nfoo", expected: false},
		{regex: ".*foo", value: "\nfoo", expected: false},
		{regex: "foo.*", value: "foo\n", expected: false},
		{regex: "foo\n.*", value: "foo\n", expected: true},
		{regex: ".*foo.*", value: "foo", expected: true},
		{regex: ".*foo.*", value: "foo bar", expected: true},
		{regex: ".*foo.*", value: "hello foo world", expected: true},
		{regex: ".*foo.*", value: "hello foo\n world", expected: false},
		{regex: ".*foo\n.*", value: "hello foo\n world", expected: true},
		{regex: ".*", value: "foo", expected: true},
		{regex: "", value: "foo", expected: false},
		{regex: "", value: "", expected: true},
	}

	for _, c := range cases {
		m, err := NewFastRegexMatcher(c.regex)
		require.NoError(t, err)
		require.Equal(t, c.expected, m.MatchString(c.value))
	}

}

func TestNewFastRegexMatcherError(t *testing.T) {
	Convey("test NewFastRegexMatcher error ", t, func() {
		Convey("test Compile error", func() {
			errRegex := struct {
				regex    string
				value    string
				expected bool
			}{
				regex: "[{", value: "{", expected: true,
			}
			m, err := NewFastRegexMatcher(errRegex.regex)
			So(err, ShouldNotBeNil)
			So(m, ShouldBeNil)
		})

		Convey("test Parse error", func() {
			cpatches := ApplyFunc(regexp.Compile,
				func(expr string) (*regexp.Regexp, error) {
					return nil, nil
				},
			)
			defer cpatches.Reset()

			patches := ApplyFunc(syntax.Parse,
				func(s string, flags syntax.Flags) (*syntax.Regexp, error) {
					return nil, errors.New("error")
				},
			)
			defer patches.Reset()

			errRegex := struct {
				regex    string
				value    string
				expected bool
			}{
				regex: ".*", value: "", expected: true,
			}
			m, err := NewFastRegexMatcher(errRegex.regex)
			So(err, ShouldNotBeNil)
			So(m, ShouldBeNil)
		})
	})
}

func TestOptimizeConcatRegex(t *testing.T) {
	cases := []struct {
		regex    string
		prefix   string
		suffix   string
		contains string
	}{
		{regex: "foo", prefix: "", suffix: "", contains: ""},
		{regex: "foo(hello|bar)", prefix: "foo", suffix: "", contains: ""},
		{regex: "foo(hello|bar)world", prefix: "foo", suffix: "world", contains: ""},
		{regex: "foo.*", prefix: "foo", suffix: "", contains: ""},
		{regex: "foo.*hello.*bar", prefix: "foo", suffix: "bar", contains: "hello"},
		{regex: ".*foo", prefix: "", suffix: "foo", contains: ""},
		{regex: "^.*foo$", prefix: "", suffix: "foo", contains: ""},
		{regex: ".*foo.*", prefix: "", suffix: "", contains: "foo"},
		{regex: ".*foo.*bar.*", prefix: "", suffix: "", contains: "foo"},
		{regex: ".*(foo|bar).*", prefix: "", suffix: "", contains: ""},
		{regex: ".*[abc].*", prefix: "", suffix: "", contains: ""},
		{regex: ".*((?i)abc).*", prefix: "", suffix: "", contains: ""},
		{regex: ".*(?i:abc).*", prefix: "", suffix: "", contains: ""},
		{regex: "(?i:abc).*", prefix: "", suffix: "", contains: ""},
		{regex: ".*(?i:abc)", prefix: "", suffix: "", contains: ""},
		{regex: ".*(?i:abc)def.*", prefix: "", suffix: "", contains: "def"},
		{regex: "(?i).*(?-i:abc)def", prefix: "", suffix: "", contains: "abc"},
		{regex: ".*(?msU:abc).*", prefix: "", suffix: "", contains: "abc"},
		{regex: "[aA]bc.*", prefix: "", suffix: "", contains: "bc"},
	}

	for _, c := range cases {
		parsed, err := syntax.Parse(c.regex, syntax.Perl)
		require.NoError(t, err)

		prefix, suffix, contains := optimizeConcatRegex(parsed)
		require.Equal(t, c.prefix, prefix)
		require.Equal(t, c.suffix, suffix)
		require.Equal(t, c.contains, contains)
	}
}
