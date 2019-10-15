package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

func analyzePackage(pkg *packages.Package) (count int) {
	for _, f := range pkg.Syntax {
		count += analyzeFile(f, pkg)
	}
	return count
}

type callVisitor struct {
	Fset     *token.FileSet
	TypeInfo *types.Info
	Errors   []error
}

func (c *callVisitor) Visit(node ast.Node) (w ast.Visitor) {
	switch node := node.(type) {
	case *ast.CallExpr:
		var ident *ast.Ident
		switch fun := node.Fun.(type) {
		case *ast.SelectorExpr:
			ident = fun.Sel
		case *ast.Ident:
			ident = fun
		}

		if obj := c.TypeInfo.ObjectOf(ident); obj != nil {
			if fn, ok := obj.(*types.Func); ok {
				c.check(node, fn)
			}
		}
		return nil
	}
	return c
}

func (c *callVisitor) check(node ast.Node, fn *types.Func) {
	if fn.Pkg() == nil {
		return
	}

	pos := c.Fset.Position(node.Pos())
	if cwd, err := os.Getwd(); err == nil {
		if filename, err := filepath.Rel(cwd, pos.Filename); err == nil {
			pos.Filename = filename
		}
	}
	if fn.Pkg().Path() == "errors" && fn.Name() == "New" {
		c.Errors = append(c.Errors, fmt.Errorf("%s: found usage of errors.New", pos))
	}
	if fn.Pkg().Path() == "fmt" && fn.Name() == "Errorf" {
		c.Errors = append(c.Errors, fmt.Errorf("%s: found usage of fmt.Errorf", pos))
	}
}

func analyzeFile(file *ast.File, pkg *packages.Package) int {
	v := callVisitor{
		Fset:     pkg.Fset,
		TypeInfo: pkg.TypesInfo,
	}
	ast.Walk(&v, file)

	if len(v.Errors) > 0 {
		for _, err := range v.Errors {
			fmt.Println(err)
		}
		return 1
	}
	return 0
}

func main() {
	cfg := packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedImports |
			packages.NeedDeps |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo,
		Env: os.Environ(),
	}
	pkgs, err := packages.Load(&cfg, os.Args[1:]...)
	if err != nil {
		panic(err)
	}

	count := 0
	for _, pkg := range pkgs {
		count += analyzePackage(pkg)
	}
	os.Exit(count)
}
