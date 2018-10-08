package semantictest

import (
	"regexp"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/influxdata/flux/semantic"
)

var CmpOptions = []cmp.Option{
	cmp.Comparer(func(x, y *regexp.Regexp) bool { return x.String() == y.String() }),

	cmpopts.IgnoreUnexported(semantic.ArrayExpression{}),
	cmpopts.IgnoreUnexported(semantic.FunctionExpression{}),
	cmpopts.IgnoreUnexported(semantic.IdentifierExpression{}),
	cmpopts.IgnoreUnexported(semantic.BinaryExpression{}),
	cmpopts.IgnoreUnexported(semantic.CallExpression{}),
	cmpopts.IgnoreUnexported(semantic.ConditionalExpression{}),
	cmpopts.IgnoreUnexported(semantic.LogicalExpression{}),
	cmpopts.IgnoreUnexported(semantic.MemberExpression{}),
	cmpopts.IgnoreUnexported(semantic.ObjectExpression{}),
	cmpopts.IgnoreUnexported(semantic.UnaryExpression{}),
	cmpopts.IgnoreUnexported(semantic.IdentifierExpression{}),
}
