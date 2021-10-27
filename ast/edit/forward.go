// Package edit provides support for editing AST trees.
//
// It's a thin wrapper over the editlite and testcase packages,
// maintained to keep API compatibility.
//
// See those packages for documentation.
package edit

import (
	"context"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/ast/editlite"
	"github.com/influxdata/flux/ast/testcase"
)

func TestcaseTransform(ctx context.Context, pkg *ast.Package, modules testcase.TestModules) ([]string, []*ast.Package, error) {
	return testcase.Transform(ctx, pkg, modules)
}

type TestModules = testcase.TestModules

var OptionNotFoundError = editlite.OptionNotFoundError

func DeleteOption(file *ast.File, name string) {
	editlite.DeleteOption(file, name)
}

func DeleteProperty(obj *ast.ObjectExpression, key string) {
	editlite.DeleteProperty(obj, key)
}

func GetOption(file *ast.File, name string) (ast.Expression, error) {
	return editlite.GetOption(file, name)
}

func GetProperty(obj *ast.ObjectExpression, key string) (ast.Expression, error) {
	return editlite.GetProperty(obj, key)
}

func HasDuplicateOptions(file *ast.File, name string) bool {
	return editlite.HasDuplicateOptions(file, name)
}

func Match(node ast.Node, pattern ast.Node, matchSlicesFuzzy bool) []ast.Node {
	return editlite.Match(node, pattern, matchSlicesFuzzy)
}

func Option(node ast.Node, optionIdentifier string, fn OptionFn) (bool, error) {
	return editlite.Option(node, optionIdentifier, fn)
}

func SetOption(file *ast.File, name string, expr ast.Expression) {
	editlite.SetOption(file, name, expr)
}

func SetProperty(obj *ast.ObjectExpression, key string, value ast.Expression) {
	editlite.SetProperty(obj, key, value)
}

type OptionFn = editlite.OptionFn

func OptionObjectFn(keyMap map[string]ast.Expression) OptionFn {
	return editlite.OptionObjectFn(keyMap)
}

func OptionValueFn(expr ast.Expression) OptionFn {
	return editlite.OptionValueFn(expr)
}
