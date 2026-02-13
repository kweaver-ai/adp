// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package util

import (
	"testing"
)

var quotetests = []struct {
	in    string
	out   string
	ascii string
}{
	{"\a\b\f\r\n\t\v", `"\a\b\f\r\n\t\v"`, `"\a\b\f\r\n\t\v"`},
	{"\\", `"\\"`, `"\\"`},
	{"abc\xffdef", `"abc\xffdef"`, `"abc\xffdef"`},
	{"\u263a", `"☺"`, `"\u263a"`},
	{"\U0010ffff", `"\U0010ffff"`, `"\U0010ffff"`},
	{"\x04", `"\x04"`, `"\x04"`},
}

var unquotetests = []struct {
	in  string
	out string
}{
	{`""`, ""},
	{`"a"`, "a"},
	{`"abc"`, "abc"},
	{`"☺"`, "☺"},
	{`"hello world"`, "hello world"},
	{`"\xFF"`, "\xFF"},
	{`"\377"`, "\377"},
	{`"\u1234"`, "\u1234"},
	{`"\U00010111"`, "\U00010111"},
	{`"\U0001011111"`, "\U0001011111"},
	{`"\a\b\f\n\r\t\v\\\""`, "\a\b\f\n\r\t\v\\\""},
	{`"'"`, "'"},

	{`''`, ""},
	{`'a'`, "a"},
	{`'abc'`, "abc"},
	{`'☺'`, "☺"},
	{`'hello world'`, "hello world"},
	{`'\ahéllo world'`, "\ahéllo world"},
	{`'\xFF'`, "\xFF"},
	{`'\377'`, "\377"},
	{`'\u1234'`, "\u1234"},
	{`'\U00010111'`, "\U00010111"},
	{`'\U0001011111'`, "\U0001011111"},
	{`'\a\b\f\n\r\t\v\\\''`, "\a\b\f\n\r\t\v\\'"},
	{`'"'`, "\""},

	{"``", ``},
	{"`a`", `a`},
	{"`abc`", `abc`},
	{"`☺`", `☺`},
	{"`hello world`", `hello world`},
	{"`\\xFF`", `\xFF`},
	{"`\\377`", `\377`},
	{"`\\`", `\`},
	{"`\n`", "\n"},
	{"`	`", `	`},
}

var misquoted = []string{
	``,
	`"`,
	`"a`,
	`"'`,
	`b"`,
	`"\"`,
	`"\9"`,
	`"\19"`,
	`"\129"`,
	`'\'`,
	`'\9'`,
	`'\19'`,
	`'\129'`,
	`'\400'`,
	`"\x1!"`,
	`"\U12345678"`,
	`"\z"`,
	"`",
	"`xxx",
	"`\"",
	`"\'"`,
	`'\"'`,
	"\"\n\"",
	"\"\\n\n\"",
	"'\n'",
	"`1`9`",
	`1231`,
	`'\xF'`,
	`""12345"`,
}

func TestUnquote(t *testing.T) {
	for _, tt := range unquotetests {
		if out, err := Unquote(tt.in); err != nil && out != tt.out {
			t.Errorf("Unquote(%#q) = %q, %v want %q, nil", tt.in, out, err, tt.out)
		}
	}

	for _, tt := range quotetests {
		if in, err := Unquote(tt.out); in != tt.in {
			t.Errorf("Unquote(%#q) = %q, %v, want %q, nil", tt.out, in, err, tt.in)
		}
	}

	for _, s := range misquoted {
		if out, err := Unquote(s); out != "" || err != ErrSyntax {
			t.Errorf("Unquote(%#q) = %q, %v want %q, %v", s, out, err, "", ErrSyntax)
		}
	}
}
