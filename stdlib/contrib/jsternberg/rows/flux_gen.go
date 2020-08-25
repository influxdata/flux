// DO NOT EDIT: This file is autogenerated via the builtin command.

package rows

import (
	ast "github.com/influxdata/flux/ast"
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
					Column: 12,
					Line:   15,
				},
				File:   "rows.flux",
				Source: "package rows\n\n// map will map each of the rows to a new value.\n// The function will be invoked for each row and the\n// return value will be used as the values in the output\n// row.\n//\n// The record that is passed to the function will contain\n// all of the keys and values in the record including group\n// keys, but the group key cannot be changed. Attempts to\n// change the group key will be ignored.\n//\n// The returned record does not need to contain values that are\n// part of the group key.\nbuiltin map",
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
						Line:   15,
					},
					File:   "rows.flux",
					Source: "builtin map",
					Start: ast.Position{
						Column: 1,
						Line:   15,
					},
				},
			},
			ID: &ast.Identifier{
				BaseNode: ast.BaseNode{
					Errors: nil,
					Loc: &ast.SourceLocation{
						End: ast.Position{
							Column: 12,
							Line:   15,
						},
						File:   "rows.flux",
						Source: "map",
						Start: ast.Position{
							Column: 9,
							Line:   15,
						},
					},
				},
				Name: "map",
			},
		}},
		Imports:  nil,
		Metadata: "parser-type=rust",
		Name:     "rows.flux",
		Package: &ast.PackageClause{
			BaseNode: ast.BaseNode{
				Errors: nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 13,
						Line:   1,
					},
					File:   "rows.flux",
					Source: "package rows",
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
						File:   "rows.flux",
						Source: "rows",
						Start: ast.Position{
							Column: 9,
							Line:   1,
						},
					},
				},
				Name: "rows",
			},
		},
	}},
	Package: "rows",
	Path:    "contrib/jsternberg/rows",
}
