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
				Comments: []ast.Comment{ast.Comment{Text: "// map will map each of the rows to a new value.\n"}, ast.Comment{Text: "// The function will be invoked for each row and the\n"}, ast.Comment{Text: "// return value will be used as the values in the output\n"}, ast.Comment{Text: "// row.\n"}, ast.Comment{Text: "//\n"}, ast.Comment{Text: "// The record that is passed to the function will contain\n"}, ast.Comment{Text: "// all of the keys and values in the record including group\n"}, ast.Comment{Text: "// keys, but the group key cannot be changed. Attempts to\n"}, ast.Comment{Text: "// change the group key will be ignored.\n"}, ast.Comment{Text: "//\n"}, ast.Comment{Text: "// The returned record does not need to contain values that are\n"}, ast.Comment{Text: "// part of the group key.\n"}},
				Errors:   nil,
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
			Colon: nil,
			ID: &ast.Identifier{
				BaseNode: ast.BaseNode{
					Comments: nil,
					Errors:   nil,
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
			Ty: ast.TypeExpression{
				BaseNode: ast.BaseNode{
					Comments: nil,
					Errors:   nil,
					Loc: &ast.SourceLocation{
						End: ast.Position{
							Column: 81,
							Line:   15,
						},
						File:   "rows.flux",
						Source: "(<-tables: [A], fn: (r: A) => B) => [B] where A: Record, B: Record",
						Start: ast.Position{
							Column: 15,
							Line:   15,
						},
					},
				},
				Constraints: []*ast.TypeConstraint{&ast.TypeConstraint{
					BaseNode: ast.BaseNode{
						Comments: nil,
						Errors:   nil,
						Loc: &ast.SourceLocation{
							End: ast.Position{
								Column: 70,
								Line:   15,
							},
							File:   "rows.flux",
							Source: "A: Record",
							Start: ast.Position{
								Column: 61,
								Line:   15,
							},
						},
					},
					Kinds: []*ast.Identifier{&ast.Identifier{
						BaseNode: ast.BaseNode{
							Comments: nil,
							Errors:   nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 70,
									Line:   15,
								},
								File:   "rows.flux",
								Source: "Record",
								Start: ast.Position{
									Column: 64,
									Line:   15,
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
									Column: 62,
									Line:   15,
								},
								File:   "rows.flux",
								Source: "A",
								Start: ast.Position{
									Column: 61,
									Line:   15,
								},
							},
						},
						Name: "A",
					},
				}, &ast.TypeConstraint{
					BaseNode: ast.BaseNode{
						Comments: nil,
						Errors:   nil,
						Loc: &ast.SourceLocation{
							End: ast.Position{
								Column: 81,
								Line:   15,
							},
							File:   "rows.flux",
							Source: "B: Record",
							Start: ast.Position{
								Column: 72,
								Line:   15,
							},
						},
					},
					Kinds: []*ast.Identifier{&ast.Identifier{
						BaseNode: ast.BaseNode{
							Comments: nil,
							Errors:   nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 81,
									Line:   15,
								},
								File:   "rows.flux",
								Source: "Record",
								Start: ast.Position{
									Column: 75,
									Line:   15,
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
									Column: 73,
									Line:   15,
								},
								File:   "rows.flux",
								Source: "B",
								Start: ast.Position{
									Column: 72,
									Line:   15,
								},
							},
						},
						Name: "B",
					},
				}},
				Ty: &ast.FunctionType{
					BaseNode: ast.BaseNode{
						Comments: nil,
						Errors:   nil,
						Loc: &ast.SourceLocation{
							End: ast.Position{
								Column: 54,
								Line:   15,
							},
							File:   "rows.flux",
							Source: "(<-tables: [A], fn: (r: A) => B) => [B]",
							Start: ast.Position{
								Column: 15,
								Line:   15,
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
									Line:   15,
								},
								File:   "rows.flux",
								Source: "<-tables: [A]",
								Start: ast.Position{
									Column: 16,
									Line:   15,
								},
							},
						},
						Kind: "Pipe",
						Name: &ast.Identifier{
							BaseNode: ast.BaseNode{
								Comments: nil,
								Errors:   nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 24,
										Line:   15,
									},
									File:   "rows.flux",
									Source: "tables",
									Start: ast.Position{
										Column: 18,
										Line:   15,
									},
								},
							},
							Name: "tables",
						},
						Ty: &ast.ArrayType{
							BaseNode: ast.BaseNode{
								Comments: nil,
								Errors:   nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 29,
										Line:   15,
									},
									File:   "rows.flux",
									Source: "[A]",
									Start: ast.Position{
										Column: 26,
										Line:   15,
									},
								},
							},
							ElementType: &ast.TvarType{
								BaseNode: ast.BaseNode{
									Comments: nil,
									Errors:   nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 28,
											Line:   15,
										},
										File:   "rows.flux",
										Source: "A",
										Start: ast.Position{
											Column: 27,
											Line:   15,
										},
									},
								},
								ID: &ast.Identifier{
									BaseNode: ast.BaseNode{
										Comments: nil,
										Errors:   nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 28,
												Line:   15,
											},
											File:   "rows.flux",
											Source: "A",
											Start: ast.Position{
												Column: 27,
												Line:   15,
											},
										},
									},
									Name: "A",
								},
							},
						},
					}, &ast.ParameterType{
						BaseNode: ast.BaseNode{
							Comments: nil,
							Errors:   nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 46,
									Line:   15,
								},
								File:   "rows.flux",
								Source: "fn: (r: A) => B",
								Start: ast.Position{
									Column: 31,
									Line:   15,
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
										Column: 33,
										Line:   15,
									},
									File:   "rows.flux",
									Source: "fn",
									Start: ast.Position{
										Column: 31,
										Line:   15,
									},
								},
							},
							Name: "fn",
						},
						Ty: &ast.FunctionType{
							BaseNode: ast.BaseNode{
								Comments: nil,
								Errors:   nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 46,
										Line:   15,
									},
									File:   "rows.flux",
									Source: "(r: A) => B",
									Start: ast.Position{
										Column: 35,
										Line:   15,
									},
								},
							},
							Parameters: []*ast.ParameterType{&ast.ParameterType{
								BaseNode: ast.BaseNode{
									Comments: nil,
									Errors:   nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 40,
											Line:   15,
										},
										File:   "rows.flux",
										Source: "r: A",
										Start: ast.Position{
											Column: 36,
											Line:   15,
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
												Column: 37,
												Line:   15,
											},
											File:   "rows.flux",
											Source: "r",
											Start: ast.Position{
												Column: 36,
												Line:   15,
											},
										},
									},
									Name: "r",
								},
								Ty: &ast.TvarType{
									BaseNode: ast.BaseNode{
										Comments: nil,
										Errors:   nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 40,
												Line:   15,
											},
											File:   "rows.flux",
											Source: "A",
											Start: ast.Position{
												Column: 39,
												Line:   15,
											},
										},
									},
									ID: &ast.Identifier{
										BaseNode: ast.BaseNode{
											Comments: nil,
											Errors:   nil,
											Loc: &ast.SourceLocation{
												End: ast.Position{
													Column: 40,
													Line:   15,
												},
												File:   "rows.flux",
												Source: "A",
												Start: ast.Position{
													Column: 39,
													Line:   15,
												},
											},
										},
										Name: "A",
									},
								},
							}},
							Return: &ast.TvarType{
								BaseNode: ast.BaseNode{
									Comments: nil,
									Errors:   nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 46,
											Line:   15,
										},
										File:   "rows.flux",
										Source: "B",
										Start: ast.Position{
											Column: 45,
											Line:   15,
										},
									},
								},
								ID: &ast.Identifier{
									BaseNode: ast.BaseNode{
										Comments: nil,
										Errors:   nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 46,
												Line:   15,
											},
											File:   "rows.flux",
											Source: "B",
											Start: ast.Position{
												Column: 45,
												Line:   15,
											},
										},
									},
									Name: "B",
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
									Column: 54,
									Line:   15,
								},
								File:   "rows.flux",
								Source: "[B]",
								Start: ast.Position{
									Column: 51,
									Line:   15,
								},
							},
						},
						ElementType: &ast.TvarType{
							BaseNode: ast.BaseNode{
								Comments: nil,
								Errors:   nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 53,
										Line:   15,
									},
									File:   "rows.flux",
									Source: "B",
									Start: ast.Position{
										Column: 52,
										Line:   15,
									},
								},
							},
							ID: &ast.Identifier{
								BaseNode: ast.BaseNode{
									Comments: nil,
									Errors:   nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 53,
											Line:   15,
										},
										File:   "rows.flux",
										Source: "B",
										Start: ast.Position{
											Column: 52,
											Line:   15,
										},
									},
								},
								Name: "B",
							},
						},
					},
				},
			},
		}},
		Eof:      nil,
		Imports:  nil,
		Metadata: "parser-type=rust",
		Name:     "rows.flux",
		Package: &ast.PackageClause{
			BaseNode: ast.BaseNode{
				Comments: nil,
				Errors:   nil,
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
					Comments: nil,
					Errors:   nil,
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
