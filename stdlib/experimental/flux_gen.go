// DO NOT EDIT: This file is autogenerated via the builtin command.

package experimental

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
					Column: 11,
					Line:   20,
				},
				File:   "experimental.flux",
				Source: "package experimental\n\nbuiltin addDuration\nbuiltin subDuration\n\n// An experimental version of group that has mode: \"extend\"\nbuiltin group\n\n// objectKeys produces a list of the keys existing on the object\nbuiltin objectKeys\n\n// set adds the values from the object onto each row of a table\nbuiltin set\n\n// An experimental version of \"to\" that:\n// - Expects pivoted data\n// - Any column in the group key is made a tag in storage\n// - All other columns are fields\n// - An error will be thrown for incompatible data types\nbuiltin to",
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
						Column: 20,
						Line:   3,
					},
					File:   "experimental.flux",
					Source: "builtin addDuration",
					Start: ast.Position{
						Column: 1,
						Line:   3,
					},
				},
			},
			ID: &ast.Identifier{
				BaseNode: ast.BaseNode{
					Errors: nil,
					Loc: &ast.SourceLocation{
						End: ast.Position{
							Column: 20,
							Line:   3,
						},
						File:   "experimental.flux",
						Source: "addDuration",
						Start: ast.Position{
							Column: 9,
							Line:   3,
						},
					},
				},
				Name: "addDuration",
			},
		}, &ast.BuiltinStatement{
			BaseNode: ast.BaseNode{
				Errors: nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 20,
						Line:   4,
					},
					File:   "experimental.flux",
					Source: "builtin subDuration",
					Start: ast.Position{
						Column: 1,
						Line:   4,
					},
				},
			},
			ID: &ast.Identifier{
				BaseNode: ast.BaseNode{
					Errors: nil,
					Loc: &ast.SourceLocation{
						End: ast.Position{
							Column: 20,
							Line:   4,
						},
						File:   "experimental.flux",
						Source: "subDuration",
						Start: ast.Position{
							Column: 9,
							Line:   4,
						},
					},
				},
				Name: "subDuration",
			},
		}, &ast.BuiltinStatement{
			BaseNode: ast.BaseNode{
				Errors: nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 14,
						Line:   7,
					},
					File:   "experimental.flux",
					Source: "builtin group",
					Start: ast.Position{
						Column: 1,
						Line:   7,
					},
				},
			},
			ID: &ast.Identifier{
				BaseNode: ast.BaseNode{
					Errors: nil,
					Loc: &ast.SourceLocation{
						End: ast.Position{
							Column: 14,
							Line:   7,
						},
						File:   "experimental.flux",
						Source: "group",
						Start: ast.Position{
							Column: 9,
							Line:   7,
						},
					},
				},
				Name: "group",
			},
		}, &ast.BuiltinStatement{
			BaseNode: ast.BaseNode{
				Errors: nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 19,
						Line:   10,
					},
					File:   "experimental.flux",
					Source: "builtin objectKeys",
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
							Column: 19,
							Line:   10,
						},
						File:   "experimental.flux",
						Source: "objectKeys",
						Start: ast.Position{
							Column: 9,
							Line:   10,
						},
					},
				},
				Name: "objectKeys",
			},
		}, &ast.BuiltinStatement{
			BaseNode: ast.BaseNode{
				Errors: nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 12,
						Line:   13,
					},
					File:   "experimental.flux",
					Source: "builtin set",
					Start: ast.Position{
						Column: 1,
						Line:   13,
					},
				},
			},
			ID: &ast.Identifier{
				BaseNode: ast.BaseNode{
					Errors: nil,
					Loc: &ast.SourceLocation{
						End: ast.Position{
							Column: 12,
							Line:   13,
						},
						File:   "experimental.flux",
						Source: "set",
						Start: ast.Position{
							Column: 9,
							Line:   13,
						},
					},
				},
				Name: "set",
			},
		}, &ast.BuiltinStatement{
			BaseNode: ast.BaseNode{
				Errors: nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 11,
						Line:   20,
					},
					File:   "experimental.flux",
					Source: "builtin to",
					Start: ast.Position{
						Column: 1,
						Line:   20,
					},
				},
			},
			ID: &ast.Identifier{
				BaseNode: ast.BaseNode{
					Errors: nil,
					Loc: &ast.SourceLocation{
						End: ast.Position{
							Column: 11,
							Line:   20,
						},
						File:   "experimental.flux",
						Source: "to",
						Start: ast.Position{
							Column: 9,
							Line:   20,
						},
					},
				},
				Name: "to",
			},
		}},
		Imports: nil,
		Name:    "experimental.flux",
		Package: &ast.PackageClause{
			BaseNode: ast.BaseNode{
				Errors: nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 21,
						Line:   1,
					},
					File:   "experimental.flux",
					Source: "package experimental",
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
							Column: 21,
							Line:   1,
						},
						File:   "experimental.flux",
						Source: "experimental",
						Start: ast.Position{
							Column: 9,
							Line:   1,
						},
					},
				},
				Name: "experimental",
			},
		},
	}},
	Package: "experimental",
	Path:    "experimental",
}
