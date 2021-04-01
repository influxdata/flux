// DO NOT EDIT: This file is autogenerated via the builtin command.

package array

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
					Column: 13,
					Line:   16,
				},
				File:   "array.flux",
				Source: "package array\n\n\n// from will construct a table from the input rows.\n//\n// This function takes the `rows` parameter. The rows\n// parameter is an array of records that will be constructed.\n// All of the records must have the same keys and the same types\n// for the values.\n//\n// Example:\n//\n//    import \"array\"\n//    array.from(rows:[{a:1, b: false, c: \"hi\"}, {a:2, b: true, c: \"bye\"}])\n//\nbuiltin from",
				Start: ast.Position{
					Column: 1,
					Line:   1,
				},
			},
		},
		Body: []ast.Statement{&ast.BuiltinStatement{
			BaseNode: ast.BaseNode{
				Comments: nil,
				Errors:   nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 13,
						Line:   16,
					},
					File:   "array.flux",
					Source: "builtin from",
					Start: ast.Position{
						Column: 1,
						Line:   16,
					},
				},
			},
			ID: &ast.Identifier{
				BaseNode: ast.BaseNode{
					Comments: nil,
					Errors:   nil,
					Loc: &ast.SourceLocation{
						End: ast.Position{
							Column: 13,
							Line:   16,
						},
						File:   "array.flux",
						Source: "from",
						Start: ast.Position{
							Column: 9,
							Line:   16,
						},
					},
				},
				Name: "from",
			},
			Ty: ast.TypeExpression{
				BaseNode: ast.BaseNode{
					Comments: nil,
					Errors:   nil,
					Loc: &ast.SourceLocation{
						End: ast.Position{
							Column: 50,
							Line:   16,
						},
						File:   "array.flux",
						Source: "(rows: [A]) => [A] where A: Record",
						Start: ast.Position{
							Column: 16,
							Line:   16,
						},
					},
				},
				Constraints: []*ast.TypeConstraint{&ast.TypeConstraint{
					BaseNode: ast.BaseNode{
						Comments: nil,
						Errors:   nil,
						Loc: &ast.SourceLocation{
							End: ast.Position{
								Column: 50,
								Line:   16,
							},
							File:   "array.flux",
							Source: "A: Record",
							Start: ast.Position{
								Column: 41,
								Line:   16,
							},
						},
					},
					Kinds: []*ast.Identifier{&ast.Identifier{
						BaseNode: ast.BaseNode{
							Comments: nil,
							Errors:   nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 50,
									Line:   16,
								},
								File:   "array.flux",
								Source: "Record",
								Start: ast.Position{
									Column: 44,
									Line:   16,
								},
							},
						},
						Name: "Record",
					}},
					Tvar: &ast.Identifier{
						BaseNode: ast.BaseNode{
							Comments: nil,
							Errors:   nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 42,
									Line:   16,
								},
								File:   "array.flux",
								Source: "A",
								Start: ast.Position{
									Column: 41,
									Line:   16,
								},
							},
						},
						Name: "A",
					},
				}},
				Ty: &ast.FunctionType{
					BaseNode: ast.BaseNode{
						Comments: nil,
						Errors:   nil,
						Loc: &ast.SourceLocation{
							End: ast.Position{
								Column: 34,
								Line:   16,
							},
							File:   "array.flux",
							Source: "(rows: [A]) => [A]",
							Start: ast.Position{
								Column: 16,
								Line:   16,
							},
						},
					},
					Parameters: []*ast.ParameterType{&ast.ParameterType{
						BaseNode: ast.BaseNode{
							Comments: nil,
							Errors:   nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 26,
									Line:   16,
								},
								File:   "array.flux",
								Source: "rows: [A]",
								Start: ast.Position{
									Column: 17,
									Line:   16,
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
										Column: 21,
										Line:   16,
									},
									File:   "array.flux",
									Source: "rows",
									Start: ast.Position{
										Column: 17,
										Line:   16,
									},
								},
							},
							Name: "rows",
						},
						Ty: &ast.ArrayType{
							BaseNode: ast.BaseNode{
								Comments: nil,
								Errors:   nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 26,
										Line:   16,
									},
									File:   "array.flux",
									Source: "[A]",
									Start: ast.Position{
										Column: 23,
										Line:   16,
									},
								},
							},
							ElementType: &ast.TvarType{
								BaseNode: ast.BaseNode{
									Comments: nil,
									Errors:   nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 25,
											Line:   16,
										},
										File:   "array.flux",
										Source: "A",
										Start: ast.Position{
											Column: 24,
											Line:   16,
										},
									},
								},
								ID: &ast.Identifier{
									BaseNode: ast.BaseNode{
										Comments: nil,
										Errors:   nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 25,
												Line:   16,
											},
											File:   "array.flux",
											Source: "A",
											Start: ast.Position{
												Column: 24,
												Line:   16,
											},
										},
									},
									Name: "A",
								},
							},
						},
					}},
					Return: &ast.ArrayType{
						BaseNode: ast.BaseNode{
							Comments: nil,
							Errors:   nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 34,
									Line:   16,
								},
								File:   "array.flux",
								Source: "[A]",
								Start: ast.Position{
									Column: 31,
									Line:   16,
								},
							},
						},
						ElementType: &ast.TvarType{
							BaseNode: ast.BaseNode{
								Comments: nil,
								Errors:   nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 33,
										Line:   16,
									},
									File:   "array.flux",
									Source: "A",
									Start: ast.Position{
										Column: 32,
										Line:   16,
									},
								},
							},
							ID: &ast.Identifier{
								BaseNode: ast.BaseNode{
									Comments: nil,
									Errors:   nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 33,
											Line:   16,
										},
										File:   "array.flux",
										Source: "A",
										Start: ast.Position{
											Column: 32,
											Line:   16,
										},
									},
								},
								Name: "A",
							},
						},
					},
				},
			},
		}},
		Imports:  nil,
		Metadata: "parser-type=rust",
		Name:     "array.flux",
		Package: &ast.PackageClause{
			BaseNode: ast.BaseNode{
				Comments: nil,
				Errors:   nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 14,
						Line:   1,
					},
					File:   "array.flux",
					Source: "package array",
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
							Column: 14,
							Line:   1,
						},
						File:   "array.flux",
						Source: "array",
						Start: ast.Position{
							Column: 9,
							Line:   1,
						},
					},
				},
				Name: "array",
			},
		},
	}},
	Package: "array",
	Path:    "array",
}
