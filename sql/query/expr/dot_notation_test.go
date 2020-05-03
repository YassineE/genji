package expr_test

import (
	"testing"

	"github.com/asdine/genji/document"
	"github.com/asdine/genji/sql/query/expr"
	"github.com/stretchr/testify/require"
)

func TestDotNotationExpr(t *testing.T) {
	tests := []struct {
		expr  string
		res   document.Value
		fails bool
	}{
		{"a", document.NewIntValue(1), false},
		{"b", func() document.Value {
			d, _ := document.NewFromJSON([]byte(`{"foo bar": [1, 2]}`))
			return document.NewDocumentValue(d)
		}(),
			false},
		{"b.`foo bar`.0", document.NewIntValue(1), false},
		{"b.`foo bar`.1", document.NewIntValue(2), false},
		{"b.`foo bar`.2", nullLitteral, false},
		{"b.0", nullLitteral, false},
		{"c.0", document.NewIntValue(1), false},
		{"c.1.foo", document.NewTextValue("bar"), false},
		{"c.foo", nullLitteral, false},
		{"d", nullLitteral, false},
	}

	d, err := document.NewFromJSON([]byte(`{
		"a": 1,
		"b": {"foo bar": [1, 2]},
		"c": [1, {"foo": "bar"}, [1, 2]]
	}`))
	require.NoError(t, err)

	for _, test := range tests {
		t.Run(test.expr, func(t *testing.T) {
			testExpr(t, test.expr, expr.EvalStack{Document: d}, test.res, test.fails)
		})
	}

	t.Run("empty stack", func(t *testing.T) {
		testExpr(t, "a", expr.EvalStack{}, nullLitteral, true)
	})
}
