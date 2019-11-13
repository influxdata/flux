// DO NOT EDIT: This file is autogenerated via the builtin command.

package http

import (
	flux "github.com/influxdata/flux"
	ast "github.com/influxdata/flux/ast"
)

func init() {
	flux.RegisterPackage(pkgAST)
}

var pkgAST = &ast.Package{
	BaseNode: ast.BaseNode{
		Errors: nil,
		Loc:    nil,
	},
	Files: []*ast.File{&ast.File{
		BaseNode: ast.BaseNode{
			Errors: nil,
			Loc: &ast.SourceLocation{
				End: ast.Position{
					Column: 12,
					Line:   5,
				},
				File:   "http.flux",
				Source: "package http\n\n// Get submits an HTTP get request to the specified URL with headers\n// Returns HTTP status code and body as a byte array\nbuiltin get",
				Start: ast.Position{
					Column: 1,
					Line:   1,
				},
			},
		},
		Body: []ast.Statement{&ast.BuiltinStatement{
			BaseNode: ast.BaseNode{
				Errors: nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 12,
						Line:   5,
					},
					File:   "http.flux",
					Source: "builtin get",
					Start: ast.Position{
						Column: 1,
						Line:   5,
					},
				},
			},
			ID: &ast.Identifier{
				BaseNode: ast.BaseNode{
					Errors: nil,
					Loc: &ast.SourceLocation{
						End: ast.Position{
							Column: 12,
							Line:   5,
						},
						File:   "http.flux",
						Source: "get",
						Start: ast.Position{
							Column: 9,
							Line:   5,
						},
					},
				},
				Name: "get",
			},
		}},
		Imports: nil,
		Name:    "http.flux",
		Package: &ast.PackageClause{
			BaseNode: ast.BaseNode{
				Errors: nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 13,
						Line:   1,
					},
					File:   "http.flux",
					Source: "package http",
					Start: ast.Position{
						Column: 1,
						Line:   1,
					},
				},
			},
			Name: &ast.Identifier{
				BaseNode: ast.BaseNode{
					Errors: nil,
					Loc: &ast.SourceLocation{
						End: ast.Position{
							Column: 13,
							Line:   1,
						},
						File:   "http.flux",
						Source: "http",
						Start: ast.Position{
							Column: 9,
							Line:   1,
						},
					},
				},
				Name: "http",
			},
		},
	}},
	Package: "http",
	Path:    "experimental/http",
}
