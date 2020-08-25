// DO NOT EDIT: This file is autogenerated via the builtin command.

package tasks

import (
	ast "github.com/influxdata/flux/ast"
	parser "github.com/influxdata/flux/internal/parser"
	runtime "github.com/influxdata/flux/runtime"
)

func init() {
	runtime.RegisterPackage(pkgAST)
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
					Column: 20,
					Line:   10,
				},
				File:   "tasks.flux",
				Source: "package tasks\n\noption lastSuccessTime = 0000-01-01T00:00:00Z\n\n// This is currently a noop, as its implementation is meant to be\n// overridden elsewhere.\n// As this function currently only returns an unimplemented error, and \n// flux has no support for doing this natively, this function is a builtin.\n// When fully implemented, it should be able to be implemented in pure flux.\nbuiltin lastSuccess",
				Start: ast.Position{
					Column: 1,
					Line:   1,
				},
			},
		},
		Body: []ast.Statement{&ast.OptionStatement{
			Assignment: &ast.VariableAssignment{
				BaseNode: ast.BaseNode{
					Errors: nil,
					Loc: &ast.SourceLocation{
						End: ast.Position{
							Column: 46,
							Line:   3,
						},
						File:   "tasks.flux",
						Source: "lastSuccessTime = 0000-01-01T00:00:00Z",
						Start: ast.Position{
							Column: 8,
							Line:   3,
						},
					},
				},
				ID: &ast.Identifier{
					BaseNode: ast.BaseNode{
						Errors: nil,
						Loc: &ast.SourceLocation{
							End: ast.Position{
								Column: 23,
								Line:   3,
							},
							File:   "tasks.flux",
							Source: "lastSuccessTime",
							Start: ast.Position{
								Column: 8,
								Line:   3,
							},
						},
					},
					Name: "lastSuccessTime",
				},
				Init: &ast.DateTimeLiteral{
					BaseNode: ast.BaseNode{
						Errors: nil,
						Loc: &ast.SourceLocation{
							End: ast.Position{
								Column: 46,
								Line:   3,
							},
							File:   "tasks.flux",
							Source: "0000-01-01T00:00:00Z",
							Start: ast.Position{
								Column: 26,
								Line:   3,
							},
						},
					},
					Value: parser.MustParseTime("0000-01-01T00:00:00Z"),
				},
			},
			BaseNode: ast.BaseNode{
				Errors: nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 46,
						Line:   3,
					},
					File:   "tasks.flux",
					Source: "option lastSuccessTime = 0000-01-01T00:00:00Z",
					Start: ast.Position{
						Column: 1,
						Line:   3,
					},
				},
			},
		}, &ast.BuiltinStatement{
			BaseNode: ast.BaseNode{
				Errors: nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 20,
						Line:   10,
					},
					File:   "tasks.flux",
					Source: "builtin lastSuccess",
					Start: ast.Position{
						Column: 1,
						Line:   10,
					},
				},
			},
			ID: &ast.Identifier{
				BaseNode: ast.BaseNode{
					Errors: nil,
					Loc: &ast.SourceLocation{
						End: ast.Position{
							Column: 20,
							Line:   10,
						},
						File:   "tasks.flux",
						Source: "lastSuccess",
						Start: ast.Position{
							Column: 9,
							Line:   10,
						},
					},
				},
				Name: "lastSuccess",
			},
		}},
		Imports:  nil,
		Metadata: "parser-type=rust",
		Name:     "tasks.flux",
		Package: &ast.PackageClause{
			BaseNode: ast.BaseNode{
				Errors: nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 14,
						Line:   1,
					},
					File:   "tasks.flux",
					Source: "package tasks",
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
							Column: 14,
							Line:   1,
						},
						File:   "tasks.flux",
						Source: "tasks",
						Start: ast.Position{
							Column: 9,
							Line:   1,
						},
					},
				},
				Name: "tasks",
			},
		},
	}},
	Package: "tasks",
	Path:    "influxdata/influxdb/tasks",
}
