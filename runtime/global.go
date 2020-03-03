package runtime

import (
	"context"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/parser"
	"github.com/influxdata/flux/values"
)

// RegisterPackage adds a builtin package
func RegisterPackage(pkg *ast.Package) {
	if err := defaultRuntime.RegisterPackage(pkg); err != nil {
		panic(err)
	}
}

// RegisterPackageValue adds a value for an identifier in a builtin package
func RegisterPackageValue(pkgpath, name string, value values.Value) {
	if err := defaultRuntime.RegisterPackageValue(pkgpath, name, value); err != nil {
		panic(err)
	}
}

// ReplacePackageValue replaces a value for an identifier in a builtin package
func ReplacePackageValue(pkgpath, name string, value values.Value) {
	if err := defaultRuntime.ReplacePackageValue(pkgpath, name, value); err != nil {
		panic(err)
	}
}

// StdLib returns an importer for the Flux standard library.
func StdLib() interpreter.Importer {
	return defaultRuntime.Stdlib()
}

// Prelude returns a scope object representing the Flux universe block
func Prelude() values.Scope {
	return defaultRuntime.Prelude()
}

// Eval accepts a Flux script and evaluates it to produce a set of side effects (as a slice of values) and a scope.
func Eval(ctx context.Context, flux string, opts ...ScopeMutator) ([]interpreter.SideEffect, values.Scope, error) {
	h := parser.ParseToHandle([]byte(flux))
	return defaultRuntime.evalHandle(ctx, h, opts...)
}

// EvalAST accepts a Flux AST and evaluates it to produce a set of side effects (as a slice of values) and a scope.
func EvalAST(ctx context.Context, astPkg *ast.Package, opts ...ScopeMutator) ([]interpreter.SideEffect, values.Scope, error) {
	return defaultRuntime.Eval(ctx, astPkg, opts...)
}

// EvalOptions is like EvalAST, but only evaluates options.
func EvalOptions(ctx context.Context, astPkg *ast.Package, opts ...ScopeMutator) ([]interpreter.SideEffect, values.Scope, error) {
	return EvalAST(ctx, options(astPkg), opts...)
}

// options returns a shallow copy of the AST, trimmed to include only option statements.
func options(astPkg *ast.Package) *ast.Package {
	trimmed := &ast.Package{
		BaseNode: astPkg.BaseNode,
		Path:     astPkg.Path,
		Package:  astPkg.Package,
	}
	for _, f := range astPkg.Files {
		var body []ast.Statement
		for _, s := range f.Body {
			if opt, ok := s.(*ast.OptionStatement); ok {
				body = append(body, opt)
			}
		}
		if len(body) > 0 {
			trimmed.Files = append(trimmed.Files, &ast.File{
				Body:     body,
				BaseNode: f.BaseNode,
				Name:     f.Name,
				Package:  f.Package,
				Imports:  f.Imports,
			})
		}
	}

	return trimmed
}

// FinalizeBuiltIns must be called to complete registration.
// Future calls to RegisterFunction or RegisterPackageValue will panic.
func FinalizeBuiltIns() {
	if err := defaultRuntime.Finalize(); err != nil {
		panic(err)
	}
}
