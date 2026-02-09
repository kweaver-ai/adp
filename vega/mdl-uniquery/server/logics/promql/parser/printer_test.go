// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExprString(t *testing.T) {

	inputs := []struct {
		in, out string
	}{
		{
			in:  `sum by() (task:errors:rate10s{job="s"})`,
			out: `sum(task:errors:rate10s{job="s"})`,
		},
		{
			in: `sum by(code) (task:errors:rate10s{job="s"})`,
		},
		{
			in: `sum without() (task:errors:rate10s{job="s"})`,
		},
		{
			in: `sum without(instance) (task:errors:rate10s{job="s"})`,
		},
		{
			in: `topk(5, task:errors:rate10s{job="s"})`,
		},
		{
			in: `count_values("value", task:errors:rate10s{job="s"})`,
		},
		{
			in: `a - on() c`,
		},
		{
			in: `a - on(b) c`,
		},
		{
			in: `a - on(b) group_left(x) c`,
		},
		{
			in: `a - on(b) group_left(x, y) c`,
		},
		{
			in:  `a - on(b) group_left c`,
			out: `a - on(b) group_left() c`,
		},
		{
			in: `a - on(b) group_left() (c)`,
		},
		{
			in: `a - ignoring(b) c`,
		},
		{
			in:  `a - ignoring() c`,
			out: `a - c`,
		},
		{
			in: `up > bool 0`,
		},
		{
			in: `a offset 1m`,
		},
		{
			in: `a offset -7m`,
		},
		{
			in: `a{c="d"}[5m] offset 1m`,
		},
		{
			in: `a[5m] offset 1m`,
		},
		{
			in: `a[12m] offset -3m`,
		},
		{
			in: `a[1h:5m] offset 1m`,
		},
		{
			in: `{__name__="a"}`,
		},
		{
			in: `a{b!="c"}[1m]`,
		},
		{
			in: `a{b=~"c"}[1m]`,
		},
		{
			in: `a{b!~"c"}[1m]`,
		},
		{
			in:  `a @ 10`,
			out: `a @ 10.000`,
		},
		{
			in:  `a[1m] @ 10`,
			out: `a[1m] @ 10.000`,
		},
		{
			in: `a @ start()`,
		},
		{
			in: `a @ end()`,
		},
		{
			in: `a[1m] @ start()`,
		},
		{
			in: `a[1m] @ end()`,
		},
	}

	for _, test := range inputs {
		expr, err := ParseExpr(testCtx, test.in)
		require.NoError(t, err)

		exp := test.in
		if test.out != "" {
			exp = test.out
		}

		require.Equal(t, exp, expr.String())
	}
}
