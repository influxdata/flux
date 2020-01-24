package executetest

import (
	"testing"

	"github.com/influxdata/flux/semantic"
)

// FunctionExpression will take a function expression as a string
// and return the *semantic.FunctionExpression.
//
// This will cause a fatal error in the test on failure.
func FunctionExpression(t testing.TB, source string) *semantic.FunctionExpression {
	t.Helper()

	pkg, err := semantic.AnalyzeSource(source)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	file := pkg.Files[0]
	if len(file.Body) != 1 {
		t.Fatal("function expression must have exactly one statement")
	}

	stmt, ok := file.Body[0].(*semantic.ExpressionStatement)
	if !ok {
		t.Fatal("statement must be an expression statement with a function expression")
	}

	fn, ok := stmt.Expression.(*semantic.FunctionExpression)
	if !ok {
		t.Fatal("expression must be a function expression")
	}
	return fn
}
