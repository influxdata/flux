// DO NOT EDIT: This file is autogenerated via the builtin command.

package json

import (
	ast "github.com/influxdata/flux/ast"
	runtime "github.com/influxdata/flux/runtime"
)

func init() {
	runtime.RegisterPackage(pkgAST)
}

var pkgAST = &ast.Package{
	BaseNode: ast.BaseNode{
		Comments: nil,
		Errors:   nil,
		Loc:      nil,
	},
	Files: []*ast.File{&ast.File{
		BaseNode: ast.BaseNode{
			Comments: nil,
			Errors:   nil,
			Loc: &ast.SourceLocation{
				End: ast.Position{
					Column: 14,
					Line:   6,
				},
				File:   "json.flux",
				Source: "package json\n\n// Parse will consume json data as bytes and return a value.\n// Lists, objects, strings, booleans and float values can be produced.\n// All numeric values are represented using the float type.\nbuiltin parse",
				Start: ast.Position{
					Column: 1,
					Line:   1,
				},
			},
		},
		Body: []ast.Statement{&ast.BuiltinStatement{
			BaseNode: ast.BaseNode{
				Comments: []ast.Comment{ast.Comment{Text: "// Parse will consume json data as bytes and return a value.\n"}, ast.Comment{Text: "// Lists, objects, strings, booleans and float values can be produced.\n"}, ast.Comment{Text: "// All numeric values are represented using the float type.\n"}},
				Errors:   nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 14,
						Line:   6,
					},
					File:   "json.flux",
					Source: "builtin parse",
					Start: ast.Position{
						Column: 1,
						Line:   6,
					},
				},
			},
			Colon: nil,
			ID: &ast.Identifier{
				BaseNode: ast.BaseNode{
					Comments: nil,
					Errors:   nil,
					Loc: &ast.SourceLocation{
						End: ast.Position{
							Column: 14,
							Line:   6,
						},
						File:   "json.flux",
						Source: "parse",
						Start: ast.Position{
							Column: 9,
							Line:   6,
						},
					},
				},
				Name: "parse",
			},
			Ty: ast.TypeExpression{
				BaseNode: ast.BaseNode{
					Comments: nil,
					Errors:   nil,
					Loc: &ast.SourceLocation{
						End: ast.Position{
							Column: 35,
							Line:   6,
						},
						File:   "json.flux",
						Source: "(data: bytes) => A",
						Start: ast.Position{
							Column: 17,
							Line:   6,
						},
					},
				},
				Constraints: []*ast.TypeConstraint{},
				Ty: &ast.FunctionType{
					BaseNode: ast.BaseNode{
						Comments: nil,
						Errors:   nil,
						Loc: &ast.SourceLocation{
							End: ast.Position{
								Column: 35,
								Line:   6,
							},
							File:   "json.flux",
							Source: "(data: bytes) => A",
							Start: ast.Position{
								Column: 17,
								Line:   6,
							},
						},
					},
					Parameters: []*ast.ParameterType{&ast.ParameterType{
						BaseNode: ast.BaseNode{
							Comments: nil,
							Errors:   nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 29,
									Line:   6,
								},
								File:   "json.flux",
								Source: "data: bytes",
								Start: ast.Position{
									Column: 18,
									Line:   6,
								},
							},
						},
						Kind: "Required",
						Name: &ast.Identifier{
							BaseNode: ast.BaseNode{
								Comments: nil,
								Errors:   nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 22,
										Line:   6,
									},
									File:   "json.flux",
									Source: "data",
									Start: ast.Position{
										Column: 18,
										Line:   6,
									},
								},
							},
							Name: "data",
						},
						Ty: &ast.NamedType{
							BaseNode: ast.BaseNode{
								Comments: nil,
								Errors:   nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 29,
										Line:   6,
									},
									File:   "json.flux",
									Source: "bytes",
									Start: ast.Position{
										Column: 24,
										Line:   6,
									},
								},
							},
							ID: &ast.Identifier{
								BaseNode: ast.BaseNode{
									Comments: nil,
									Errors:   nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 29,
											Line:   6,
										},
										File:   "json.flux",
										Source: "bytes",
										Start: ast.Position{
											Column: 24,
											Line:   6,
										},
									},
								},
								Name: "bytes",
							},
						},
					}},
					Return: &ast.TvarType{
						BaseNode: ast.BaseNode{
							Comments: nil,
							Errors:   nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 35,
									Line:   6,
								},
								File:   "json.flux",
								Source: "A",
								Start: ast.Position{
									Column: 34,
									Line:   6,
								},
							},
						},
						ID: &ast.Identifier{
							BaseNode: ast.BaseNode{
								Comments: nil,
								Errors:   nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 35,
										Line:   6,
									},
									File:   "json.flux",
									Source: "A",
									Start: ast.Position{
										Column: 34,
										Line:   6,
									},
								},
							},
							Name: "A",
						},
					},
				},
			},
		}},
		Eof:      nil,
		Imports:  nil,
		Metadata: "parser-type=rust",
		Name:     "json.flux",
		Package: &ast.PackageClause{
			BaseNode: ast.BaseNode{
				Comments: nil,
				Errors:   nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 13,
						Line:   1,
					},
					File:   "json.flux",
					Source: "package json",
					Start: ast.Position{
						Column: 1,
						Line:   1,
					},
				},
			},
			Name: &ast.Identifier{
				BaseNode: ast.BaseNode{
					Comments: nil,
					Errors:   nil,
					Loc: &ast.SourceLocation{
						End: ast.Position{
							Column: 13,
							Line:   1,
						},
						File:   "json.flux",
						Source: "json",
						Start: ast.Position{
							Column: 9,
							Line:   1,
						},
					},
				},
				Name: "json",
			},
		},
	}},
	Package: "json",
	Path:    "experimental/json",
}
