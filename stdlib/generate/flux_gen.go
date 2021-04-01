// DO NOT EDIT: This file is autogenerated via the builtin command.

package generate

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
					Line:   5,
				},
				File:   "generate.flux",
				Source: "package generate\n\n\n// From generates a table with count rows using fn to determine the value of each row.\nbuiltin from",
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
						Line:   5,
					},
					File:   "generate.flux",
					Source: "builtin from",
					Start: ast.Position{
						Column: 1,
						Line:   5,
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
							Line:   5,
						},
						File:   "generate.flux",
						Source: "from",
						Start: ast.Position{
							Column: 9,
							Line:   5,
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
							Column: 16,
							Line:   16,
						},
						File:   "generate.flux",
						Source: "(\n    start: A,\n    stop: A,\n    count: int,\n    fn: (n: int) => int,\n) => [{\n    _start: time,\n    _stop: time,\n    _time: time,\n    _value: int,\n}] where\n    A: Timeable",
						Start: ast.Position{
							Column: 16,
							Line:   5,
						},
					},
				},
				Constraints: []*ast.TypeConstraint{&ast.TypeConstraint{
					BaseNode: ast.BaseNode{
						Comments: nil,
						Errors:   nil,
						Loc: &ast.SourceLocation{
							End: ast.Position{
								Column: 16,
								Line:   16,
							},
							File:   "generate.flux",
							Source: "A: Timeable",
							Start: ast.Position{
								Column: 5,
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
									Column: 16,
									Line:   16,
								},
								File:   "generate.flux",
								Source: "Timeable",
								Start: ast.Position{
									Column: 8,
									Line:   16,
								},
							},
						},
						Name: "Timeable",
					}},
					Tvar: &ast.Identifier{
						BaseNode: ast.BaseNode{
							Comments: nil,
							Errors:   nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 6,
									Line:   16,
								},
								File:   "generate.flux",
								Source: "A",
								Start: ast.Position{
									Column: 5,
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
								Column: 3,
								Line:   15,
							},
							File:   "generate.flux",
							Source: "(\n    start: A,\n    stop: A,\n    count: int,\n    fn: (n: int) => int,\n) => [{\n    _start: time,\n    _stop: time,\n    _time: time,\n    _value: int,\n}]",
							Start: ast.Position{
								Column: 16,
								Line:   5,
							},
						},
					},
					Parameters: []*ast.ParameterType{&ast.ParameterType{
						BaseNode: ast.BaseNode{
							Comments: nil,
							Errors:   nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 13,
									Line:   6,
								},
								File:   "generate.flux",
								Source: "start: A",
								Start: ast.Position{
									Column: 5,
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
										Column: 10,
										Line:   6,
									},
									File:   "generate.flux",
									Source: "start",
									Start: ast.Position{
										Column: 5,
										Line:   6,
									},
								},
							},
							Name: "start",
						},
						Ty: &ast.TvarType{
							BaseNode: ast.BaseNode{
								Comments: nil,
								Errors:   nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 13,
										Line:   6,
									},
									File:   "generate.flux",
									Source: "A",
									Start: ast.Position{
										Column: 12,
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
											Column: 13,
											Line:   6,
										},
										File:   "generate.flux",
										Source: "A",
										Start: ast.Position{
											Column: 12,
											Line:   6,
										},
									},
								},
								Name: "A",
							},
						},
					}, &ast.ParameterType{
						BaseNode: ast.BaseNode{
							Comments: nil,
							Errors:   nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 12,
									Line:   7,
								},
								File:   "generate.flux",
								Source: "stop: A",
								Start: ast.Position{
									Column: 5,
									Line:   7,
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
										Column: 9,
										Line:   7,
									},
									File:   "generate.flux",
									Source: "stop",
									Start: ast.Position{
										Column: 5,
										Line:   7,
									},
								},
							},
							Name: "stop",
						},
						Ty: &ast.TvarType{
							BaseNode: ast.BaseNode{
								Comments: nil,
								Errors:   nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 12,
										Line:   7,
									},
									File:   "generate.flux",
									Source: "A",
									Start: ast.Position{
										Column: 11,
										Line:   7,
									},
								},
							},
							ID: &ast.Identifier{
								BaseNode: ast.BaseNode{
									Comments: nil,
									Errors:   nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 12,
											Line:   7,
										},
										File:   "generate.flux",
										Source: "A",
										Start: ast.Position{
											Column: 11,
											Line:   7,
										},
									},
								},
								Name: "A",
							},
						},
					}, &ast.ParameterType{
						BaseNode: ast.BaseNode{
							Comments: nil,
							Errors:   nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 15,
									Line:   8,
								},
								File:   "generate.flux",
								Source: "count: int",
								Start: ast.Position{
									Column: 5,
									Line:   8,
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
										Column: 10,
										Line:   8,
									},
									File:   "generate.flux",
									Source: "count",
									Start: ast.Position{
										Column: 5,
										Line:   8,
									},
								},
							},
							Name: "count",
						},
						Ty: &ast.NamedType{
							BaseNode: ast.BaseNode{
								Comments: nil,
								Errors:   nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 15,
										Line:   8,
									},
									File:   "generate.flux",
									Source: "int",
									Start: ast.Position{
										Column: 12,
										Line:   8,
									},
								},
							},
							ID: &ast.Identifier{
								BaseNode: ast.BaseNode{
									Comments: nil,
									Errors:   nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 15,
											Line:   8,
										},
										File:   "generate.flux",
										Source: "int",
										Start: ast.Position{
											Column: 12,
											Line:   8,
										},
									},
								},
								Name: "int",
							},
						},
					}, &ast.ParameterType{
						BaseNode: ast.BaseNode{
							Comments: nil,
							Errors:   nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 24,
									Line:   9,
								},
								File:   "generate.flux",
								Source: "fn: (n: int) => int",
								Start: ast.Position{
									Column: 5,
									Line:   9,
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
										Column: 7,
										Line:   9,
									},
									File:   "generate.flux",
									Source: "fn",
									Start: ast.Position{
										Column: 5,
										Line:   9,
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
										Column: 24,
										Line:   9,
									},
									File:   "generate.flux",
									Source: "(n: int) => int",
									Start: ast.Position{
										Column: 9,
										Line:   9,
									},
								},
							},
							Parameters: []*ast.ParameterType{&ast.ParameterType{
								BaseNode: ast.BaseNode{
									Comments: nil,
									Errors:   nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 16,
											Line:   9,
										},
										File:   "generate.flux",
										Source: "n: int",
										Start: ast.Position{
											Column: 10,
											Line:   9,
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
												Column: 11,
												Line:   9,
											},
											File:   "generate.flux",
											Source: "n",
											Start: ast.Position{
												Column: 10,
												Line:   9,
											},
										},
									},
									Name: "n",
								},
								Ty: &ast.NamedType{
									BaseNode: ast.BaseNode{
										Comments: nil,
										Errors:   nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 16,
												Line:   9,
											},
											File:   "generate.flux",
											Source: "int",
											Start: ast.Position{
												Column: 13,
												Line:   9,
											},
										},
									},
									ID: &ast.Identifier{
										BaseNode: ast.BaseNode{
											Comments: nil,
											Errors:   nil,
											Loc: &ast.SourceLocation{
												End: ast.Position{
													Column: 16,
													Line:   9,
												},
												File:   "generate.flux",
												Source: "int",
												Start: ast.Position{
													Column: 13,
													Line:   9,
												},
											},
										},
										Name: "int",
									},
								},
							}},
							Return: &ast.NamedType{
								BaseNode: ast.BaseNode{
									Comments: nil,
									Errors:   nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 24,
											Line:   9,
										},
										File:   "generate.flux",
										Source: "int",
										Start: ast.Position{
											Column: 21,
											Line:   9,
										},
									},
								},
								ID: &ast.Identifier{
									BaseNode: ast.BaseNode{
										Comments: nil,
										Errors:   nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 24,
												Line:   9,
											},
											File:   "generate.flux",
											Source: "int",
											Start: ast.Position{
												Column: 21,
												Line:   9,
											},
										},
									},
									Name: "int",
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
									Column: 3,
									Line:   15,
								},
								File:   "generate.flux",
								Source: "[{\n    _start: time,\n    _stop: time,\n    _time: time,\n    _value: int,\n}]",
								Start: ast.Position{
									Column: 6,
									Line:   10,
								},
							},
						},
						ElementType: &ast.RecordType{
							BaseNode: ast.BaseNode{
								Comments: nil,
								Errors:   nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 2,
										Line:   15,
									},
									File:   "generate.flux",
									Source: "{\n    _start: time,\n    _stop: time,\n    _time: time,\n    _value: int,\n}",
									Start: ast.Position{
										Column: 7,
										Line:   10,
									},
								},
							},
							Properties: []*ast.PropertyType{&ast.PropertyType{
								BaseNode: ast.BaseNode{
									Comments: nil,
									Errors:   nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 17,
											Line:   11,
										},
										File:   "generate.flux",
										Source: "_start: time",
										Start: ast.Position{
											Column: 5,
											Line:   11,
										},
									},
								},
								Name: &ast.Identifier{
									BaseNode: ast.BaseNode{
										Comments: nil,
										Errors:   nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 11,
												Line:   11,
											},
											File:   "generate.flux",
											Source: "_start",
											Start: ast.Position{
												Column: 5,
												Line:   11,
											},
										},
									},
									Name: "_start",
								},
								Ty: &ast.NamedType{
									BaseNode: ast.BaseNode{
										Comments: nil,
										Errors:   nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 17,
												Line:   11,
											},
											File:   "generate.flux",
											Source: "time",
											Start: ast.Position{
												Column: 13,
												Line:   11,
											},
										},
									},
									ID: &ast.Identifier{
										BaseNode: ast.BaseNode{
											Comments: nil,
											Errors:   nil,
											Loc: &ast.SourceLocation{
												End: ast.Position{
													Column: 17,
													Line:   11,
												},
												File:   "generate.flux",
												Source: "time",
												Start: ast.Position{
													Column: 13,
													Line:   11,
												},
											},
										},
										Name: "time",
									},
								},
							}, &ast.PropertyType{
								BaseNode: ast.BaseNode{
									Comments: nil,
									Errors:   nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 16,
											Line:   12,
										},
										File:   "generate.flux",
										Source: "_stop: time",
										Start: ast.Position{
											Column: 5,
											Line:   12,
										},
									},
								},
								Name: &ast.Identifier{
									BaseNode: ast.BaseNode{
										Comments: nil,
										Errors:   nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 10,
												Line:   12,
											},
											File:   "generate.flux",
											Source: "_stop",
											Start: ast.Position{
												Column: 5,
												Line:   12,
											},
										},
									},
									Name: "_stop",
								},
								Ty: &ast.NamedType{
									BaseNode: ast.BaseNode{
										Comments: nil,
										Errors:   nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 16,
												Line:   12,
											},
											File:   "generate.flux",
											Source: "time",
											Start: ast.Position{
												Column: 12,
												Line:   12,
											},
										},
									},
									ID: &ast.Identifier{
										BaseNode: ast.BaseNode{
											Comments: nil,
											Errors:   nil,
											Loc: &ast.SourceLocation{
												End: ast.Position{
													Column: 16,
													Line:   12,
												},
												File:   "generate.flux",
												Source: "time",
												Start: ast.Position{
													Column: 12,
													Line:   12,
												},
											},
										},
										Name: "time",
									},
								},
							}, &ast.PropertyType{
								BaseNode: ast.BaseNode{
									Comments: nil,
									Errors:   nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 16,
											Line:   13,
										},
										File:   "generate.flux",
										Source: "_time: time",
										Start: ast.Position{
											Column: 5,
											Line:   13,
										},
									},
								},
								Name: &ast.Identifier{
									BaseNode: ast.BaseNode{
										Comments: nil,
										Errors:   nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 10,
												Line:   13,
											},
											File:   "generate.flux",
											Source: "_time",
											Start: ast.Position{
												Column: 5,
												Line:   13,
											},
										},
									},
									Name: "_time",
								},
								Ty: &ast.NamedType{
									BaseNode: ast.BaseNode{
										Comments: nil,
										Errors:   nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 16,
												Line:   13,
											},
											File:   "generate.flux",
											Source: "time",
											Start: ast.Position{
												Column: 12,
												Line:   13,
											},
										},
									},
									ID: &ast.Identifier{
										BaseNode: ast.BaseNode{
											Comments: nil,
											Errors:   nil,
											Loc: &ast.SourceLocation{
												End: ast.Position{
													Column: 16,
													Line:   13,
												},
												File:   "generate.flux",
												Source: "time",
												Start: ast.Position{
													Column: 12,
													Line:   13,
												},
											},
										},
										Name: "time",
									},
								},
							}, &ast.PropertyType{
								BaseNode: ast.BaseNode{
									Comments: nil,
									Errors:   nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 16,
											Line:   14,
										},
										File:   "generate.flux",
										Source: "_value: int",
										Start: ast.Position{
											Column: 5,
											Line:   14,
										},
									},
								},
								Name: &ast.Identifier{
									BaseNode: ast.BaseNode{
										Comments: nil,
										Errors:   nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 11,
												Line:   14,
											},
											File:   "generate.flux",
											Source: "_value",
											Start: ast.Position{
												Column: 5,
												Line:   14,
											},
										},
									},
									Name: "_value",
								},
								Ty: &ast.NamedType{
									BaseNode: ast.BaseNode{
										Comments: nil,
										Errors:   nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 16,
												Line:   14,
											},
											File:   "generate.flux",
											Source: "int",
											Start: ast.Position{
												Column: 13,
												Line:   14,
											},
										},
									},
									ID: &ast.Identifier{
										BaseNode: ast.BaseNode{
											Comments: nil,
											Errors:   nil,
											Loc: &ast.SourceLocation{
												End: ast.Position{
													Column: 16,
													Line:   14,
												},
												File:   "generate.flux",
												Source: "int",
												Start: ast.Position{
													Column: 13,
													Line:   14,
												},
											},
										},
										Name: "int",
									},
								},
							}},
							Tvar: nil,
						},
					},
				},
			},
		}},
		Imports:  nil,
		Metadata: "parser-type=rust",
		Name:     "generate.flux",
		Package: &ast.PackageClause{
			BaseNode: ast.BaseNode{
				Comments: nil,
				Errors:   nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 17,
						Line:   1,
					},
					File:   "generate.flux",
					Source: "package generate",
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
							Column: 17,
							Line:   1,
						},
						File:   "generate.flux",
						Source: "generate",
						Start: ast.Position{
							Column: 9,
							Line:   1,
						},
					},
				},
				Name: "generate",
			},
		},
	}},
	Package: "generate",
	Path:    "generate",
}
