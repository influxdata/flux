// DO NOT EDIT: This file is autogenerated via the builtin command.

package aggregate

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
					Column: 10,
					Line:   14,
				},
				File:   "",
				Source: "package aggregate\n\nimport \"experimental\"\n\nrate = (tables=<-, every, groupColumns=[], unit=1s) =>\n    tables\n        |> derivative(nonNegative:true, unit:unit)\n        |> aggregateWindow(every: every, fn : (tables=<-, column) =>\n            tables\n                |> mean(column: column)\n                |> group(columns: groupColumns)\n                |> experimental.group(columns: [\"_start\", \"_stop\"], mode:\"extend\")\n                |> sum()\n        )",
				Start: ast.Position{
					Column: 1,
					Line:   1,
				},
			},
		},
		Body: []ast.Statement{&ast.VariableAssignment{
			BaseNode: ast.BaseNode{
				Errors: nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 10,
						Line:   14,
					},
					File:   "",
					Source: "rate = (tables=<-, every, groupColumns=[], unit=1s) =>\n    tables\n        |> derivative(nonNegative:true, unit:unit)\n        |> aggregateWindow(every: every, fn : (tables=<-, column) =>\n            tables\n                |> mean(column: column)\n                |> group(columns: groupColumns)\n                |> experimental.group(columns: [\"_start\", \"_stop\"], mode:\"extend\")\n                |> sum()\n        )",
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
							Column: 5,
							Line:   5,
						},
						File:   "",
						Source: "rate",
						Start: ast.Position{
							Column: 1,
							Line:   5,
						},
					},
				},
				Name: "rate",
			},
			Init: &ast.FunctionExpression{
				BaseNode: ast.BaseNode{
					Errors: nil,
					Loc: &ast.SourceLocation{
						End: ast.Position{
							Column: 10,
							Line:   14,
						},
						File:   "",
						Source: "(tables=<-, every, groupColumns=[], unit=1s) =>\n    tables\n        |> derivative(nonNegative:true, unit:unit)\n        |> aggregateWindow(every: every, fn : (tables=<-, column) =>\n            tables\n                |> mean(column: column)\n                |> group(columns: groupColumns)\n                |> experimental.group(columns: [\"_start\", \"_stop\"], mode:\"extend\")\n                |> sum()\n        )",
						Start: ast.Position{
							Column: 8,
							Line:   5,
						},
					},
				},
				Body: &ast.PipeExpression{
					Argument: &ast.PipeExpression{
						Argument: &ast.Identifier{
							BaseNode: ast.BaseNode{
								Errors: nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 11,
										Line:   6,
									},
									File:   "",
									Source: "tables",
									Start: ast.Position{
										Column: 5,
										Line:   6,
									},
								},
							},
							Name: "tables",
						},
						BaseNode: ast.BaseNode{
							Errors: nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 51,
									Line:   7,
								},
								File:   "",
								Source: "tables\n        |> derivative(nonNegative:true, unit:unit)",
								Start: ast.Position{
									Column: 5,
									Line:   6,
								},
							},
						},
						Call: &ast.CallExpression{
							Arguments: []ast.Expression{&ast.ObjectExpression{
								BaseNode: ast.BaseNode{
									Errors: nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 50,
											Line:   7,
										},
										File:   "",
										Source: "nonNegative:true, unit:unit",
										Start: ast.Position{
											Column: 23,
											Line:   7,
										},
									},
								},
								Properties: []*ast.Property{&ast.Property{
									BaseNode: ast.BaseNode{
										Errors: nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 39,
												Line:   7,
											},
											File:   "",
											Source: "nonNegative:true",
											Start: ast.Position{
												Column: 23,
												Line:   7,
											},
										},
									},
									Key: &ast.Identifier{
										BaseNode: ast.BaseNode{
											Errors: nil,
											Loc: &ast.SourceLocation{
												End: ast.Position{
													Column: 34,
													Line:   7,
												},
												File:   "",
												Source: "nonNegative",
												Start: ast.Position{
													Column: 23,
													Line:   7,
												},
											},
										},
										Name: "nonNegative",
									},
									Value: &ast.Identifier{
										BaseNode: ast.BaseNode{
											Errors: nil,
											Loc: &ast.SourceLocation{
												End: ast.Position{
													Column: 39,
													Line:   7,
												},
												File:   "",
												Source: "true",
												Start: ast.Position{
													Column: 35,
													Line:   7,
												},
											},
										},
										Name: "true",
									},
								}, &ast.Property{
									BaseNode: ast.BaseNode{
										Errors: nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 50,
												Line:   7,
											},
											File:   "",
											Source: "unit:unit",
											Start: ast.Position{
												Column: 41,
												Line:   7,
											},
										},
									},
									Key: &ast.Identifier{
										BaseNode: ast.BaseNode{
											Errors: nil,
											Loc: &ast.SourceLocation{
												End: ast.Position{
													Column: 45,
													Line:   7,
												},
												File:   "",
												Source: "unit",
												Start: ast.Position{
													Column: 41,
													Line:   7,
												},
											},
										},
										Name: "unit",
									},
									Value: &ast.Identifier{
										BaseNode: ast.BaseNode{
											Errors: nil,
											Loc: &ast.SourceLocation{
												End: ast.Position{
													Column: 50,
													Line:   7,
												},
												File:   "",
												Source: "unit",
												Start: ast.Position{
													Column: 46,
													Line:   7,
												},
											},
										},
										Name: "unit",
									},
								}},
								With: nil,
							}},
							BaseNode: ast.BaseNode{
								Errors: nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 51,
										Line:   7,
									},
									File:   "",
									Source: "derivative(nonNegative:true, unit:unit)",
									Start: ast.Position{
										Column: 12,
										Line:   7,
									},
								},
							},
							Callee: &ast.Identifier{
								BaseNode: ast.BaseNode{
									Errors: nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 22,
											Line:   7,
										},
										File:   "",
										Source: "derivative",
										Start: ast.Position{
											Column: 12,
											Line:   7,
										},
									},
								},
								Name: "derivative",
							},
						},
					},
					BaseNode: ast.BaseNode{
						Errors: nil,
						Loc: &ast.SourceLocation{
							End: ast.Position{
								Column: 10,
								Line:   14,
							},
							File:   "",
							Source: "tables\n        |> derivative(nonNegative:true, unit:unit)\n        |> aggregateWindow(every: every, fn : (tables=<-, column) =>\n            tables\n                |> mean(column: column)\n                |> group(columns: groupColumns)\n                |> experimental.group(columns: [\"_start\", \"_stop\"], mode:\"extend\")\n                |> sum()\n        )",
							Start: ast.Position{
								Column: 5,
								Line:   6,
							},
						},
					},
					Call: &ast.CallExpression{
						Arguments: []ast.Expression{&ast.ObjectExpression{
							BaseNode: ast.BaseNode{
								Errors: nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 25,
										Line:   13,
									},
									File:   "",
									Source: "every: every, fn : (tables=<-, column) =>\n            tables\n                |> mean(column: column)\n                |> group(columns: groupColumns)\n                |> experimental.group(columns: [\"_start\", \"_stop\"], mode:\"extend\")\n                |> sum()",
									Start: ast.Position{
										Column: 28,
										Line:   8,
									},
								},
							},
							Properties: []*ast.Property{&ast.Property{
								BaseNode: ast.BaseNode{
									Errors: nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 40,
											Line:   8,
										},
										File:   "",
										Source: "every: every",
										Start: ast.Position{
											Column: 28,
											Line:   8,
										},
									},
								},
								Key: &ast.Identifier{
									BaseNode: ast.BaseNode{
										Errors: nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 33,
												Line:   8,
											},
											File:   "",
											Source: "every",
											Start: ast.Position{
												Column: 28,
												Line:   8,
											},
										},
									},
									Name: "every",
								},
								Value: &ast.Identifier{
									BaseNode: ast.BaseNode{
										Errors: nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 40,
												Line:   8,
											},
											File:   "",
											Source: "every",
											Start: ast.Position{
												Column: 35,
												Line:   8,
											},
										},
									},
									Name: "every",
								},
							}, &ast.Property{
								BaseNode: ast.BaseNode{
									Errors: nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 25,
											Line:   13,
										},
										File:   "",
										Source: "fn : (tables=<-, column) =>\n            tables\n                |> mean(column: column)\n                |> group(columns: groupColumns)\n                |> experimental.group(columns: [\"_start\", \"_stop\"], mode:\"extend\")\n                |> sum()",
										Start: ast.Position{
											Column: 42,
											Line:   8,
										},
									},
								},
								Key: &ast.Identifier{
									BaseNode: ast.BaseNode{
										Errors: nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 44,
												Line:   8,
											},
											File:   "",
											Source: "fn",
											Start: ast.Position{
												Column: 42,
												Line:   8,
											},
										},
									},
									Name: "fn",
								},
								Value: &ast.FunctionExpression{
									BaseNode: ast.BaseNode{
										Errors: nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 25,
												Line:   13,
											},
											File:   "",
											Source: "(tables=<-, column) =>\n            tables\n                |> mean(column: column)\n                |> group(columns: groupColumns)\n                |> experimental.group(columns: [\"_start\", \"_stop\"], mode:\"extend\")\n                |> sum()",
											Start: ast.Position{
												Column: 47,
												Line:   8,
											},
										},
									},
									Body: &ast.PipeExpression{
										Argument: &ast.PipeExpression{
											Argument: &ast.PipeExpression{
												Argument: &ast.PipeExpression{
													Argument: &ast.Identifier{
														BaseNode: ast.BaseNode{
															Errors: nil,
															Loc: &ast.SourceLocation{
																End: ast.Position{
																	Column: 19,
																	Line:   9,
																},
																File:   "",
																Source: "tables",
																Start: ast.Position{
																	Column: 13,
																	Line:   9,
																},
															},
														},
														Name: "tables",
													},
													BaseNode: ast.BaseNode{
														Errors: nil,
														Loc: &ast.SourceLocation{
															End: ast.Position{
																Column: 40,
																Line:   10,
															},
															File:   "",
															Source: "tables\n                |> mean(column: column)",
															Start: ast.Position{
																Column: 13,
																Line:   9,
															},
														},
													},
													Call: &ast.CallExpression{
														Arguments: []ast.Expression{&ast.ObjectExpression{
															BaseNode: ast.BaseNode{
																Errors: nil,
																Loc: &ast.SourceLocation{
																	End: ast.Position{
																		Column: 39,
																		Line:   10,
																	},
																	File:   "",
																	Source: "column: column",
																	Start: ast.Position{
																		Column: 25,
																		Line:   10,
																	},
																},
															},
															Properties: []*ast.Property{&ast.Property{
																BaseNode: ast.BaseNode{
																	Errors: nil,
																	Loc: &ast.SourceLocation{
																		End: ast.Position{
																			Column: 39,
																			Line:   10,
																		},
																		File:   "",
																		Source: "column: column",
																		Start: ast.Position{
																			Column: 25,
																			Line:   10,
																		},
																	},
																},
																Key: &ast.Identifier{
																	BaseNode: ast.BaseNode{
																		Errors: nil,
																		Loc: &ast.SourceLocation{
																			End: ast.Position{
																				Column: 31,
																				Line:   10,
																			},
																			File:   "",
																			Source: "column",
																			Start: ast.Position{
																				Column: 25,
																				Line:   10,
																			},
																		},
																	},
																	Name: "column",
																},
																Value: &ast.Identifier{
																	BaseNode: ast.BaseNode{
																		Errors: nil,
																		Loc: &ast.SourceLocation{
																			End: ast.Position{
																				Column: 39,
																				Line:   10,
																			},
																			File:   "",
																			Source: "column",
																			Start: ast.Position{
																				Column: 33,
																				Line:   10,
																			},
																		},
																	},
																	Name: "column",
																},
															}},
															With: nil,
														}},
														BaseNode: ast.BaseNode{
															Errors: nil,
															Loc: &ast.SourceLocation{
																End: ast.Position{
																	Column: 40,
																	Line:   10,
																},
																File:   "",
																Source: "mean(column: column)",
																Start: ast.Position{
																	Column: 20,
																	Line:   10,
																},
															},
														},
														Callee: &ast.Identifier{
															BaseNode: ast.BaseNode{
																Errors: nil,
																Loc: &ast.SourceLocation{
																	End: ast.Position{
																		Column: 24,
																		Line:   10,
																	},
																	File:   "",
																	Source: "mean",
																	Start: ast.Position{
																		Column: 20,
																		Line:   10,
																	},
																},
															},
															Name: "mean",
														},
													},
												},
												BaseNode: ast.BaseNode{
													Errors: nil,
													Loc: &ast.SourceLocation{
														End: ast.Position{
															Column: 48,
															Line:   11,
														},
														File:   "",
														Source: "tables\n                |> mean(column: column)\n                |> group(columns: groupColumns)",
														Start: ast.Position{
															Column: 13,
															Line:   9,
														},
													},
												},
												Call: &ast.CallExpression{
													Arguments: []ast.Expression{&ast.ObjectExpression{
														BaseNode: ast.BaseNode{
															Errors: nil,
															Loc: &ast.SourceLocation{
																End: ast.Position{
																	Column: 47,
																	Line:   11,
																},
																File:   "",
																Source: "columns: groupColumns",
																Start: ast.Position{
																	Column: 26,
																	Line:   11,
																},
															},
														},
														Properties: []*ast.Property{&ast.Property{
															BaseNode: ast.BaseNode{
																Errors: nil,
																Loc: &ast.SourceLocation{
																	End: ast.Position{
																		Column: 47,
																		Line:   11,
																	},
																	File:   "",
																	Source: "columns: groupColumns",
																	Start: ast.Position{
																		Column: 26,
																		Line:   11,
																	},
																},
															},
															Key: &ast.Identifier{
																BaseNode: ast.BaseNode{
																	Errors: nil,
																	Loc: &ast.SourceLocation{
																		End: ast.Position{
																			Column: 33,
																			Line:   11,
																		},
																		File:   "",
																		Source: "columns",
																		Start: ast.Position{
																			Column: 26,
																			Line:   11,
																		},
																	},
																},
																Name: "columns",
															},
															Value: &ast.Identifier{
																BaseNode: ast.BaseNode{
																	Errors: nil,
																	Loc: &ast.SourceLocation{
																		End: ast.Position{
																			Column: 47,
																			Line:   11,
																		},
																		File:   "",
																		Source: "groupColumns",
																		Start: ast.Position{
																			Column: 35,
																			Line:   11,
																		},
																	},
																},
																Name: "groupColumns",
															},
														}},
														With: nil,
													}},
													BaseNode: ast.BaseNode{
														Errors: nil,
														Loc: &ast.SourceLocation{
															End: ast.Position{
																Column: 48,
																Line:   11,
															},
															File:   "",
															Source: "group(columns: groupColumns)",
															Start: ast.Position{
																Column: 20,
																Line:   11,
															},
														},
													},
													Callee: &ast.Identifier{
														BaseNode: ast.BaseNode{
															Errors: nil,
															Loc: &ast.SourceLocation{
																End: ast.Position{
																	Column: 25,
																	Line:   11,
																},
																File:   "",
																Source: "group",
																Start: ast.Position{
																	Column: 20,
																	Line:   11,
																},
															},
														},
														Name: "group",
													},
												},
											},
											BaseNode: ast.BaseNode{
												Errors: nil,
												Loc: &ast.SourceLocation{
													End: ast.Position{
														Column: 83,
														Line:   12,
													},
													File:   "",
													Source: "tables\n                |> mean(column: column)\n                |> group(columns: groupColumns)\n                |> experimental.group(columns: [\"_start\", \"_stop\"], mode:\"extend\")",
													Start: ast.Position{
														Column: 13,
														Line:   9,
													},
												},
											},
											Call: &ast.CallExpression{
												Arguments: []ast.Expression{&ast.ObjectExpression{
													BaseNode: ast.BaseNode{
														Errors: nil,
														Loc: &ast.SourceLocation{
															End: ast.Position{
																Column: 82,
																Line:   12,
															},
															File:   "",
															Source: "columns: [\"_start\", \"_stop\"], mode:\"extend\"",
															Start: ast.Position{
																Column: 39,
																Line:   12,
															},
														},
													},
													Properties: []*ast.Property{&ast.Property{
														BaseNode: ast.BaseNode{
															Errors: nil,
															Loc: &ast.SourceLocation{
																End: ast.Position{
																	Column: 67,
																	Line:   12,
																},
																File:   "",
																Source: "columns: [\"_start\", \"_stop\"]",
																Start: ast.Position{
																	Column: 39,
																	Line:   12,
																},
															},
														},
														Key: &ast.Identifier{
															BaseNode: ast.BaseNode{
																Errors: nil,
																Loc: &ast.SourceLocation{
																	End: ast.Position{
																		Column: 46,
																		Line:   12,
																	},
																	File:   "",
																	Source: "columns",
																	Start: ast.Position{
																		Column: 39,
																		Line:   12,
																	},
																},
															},
															Name: "columns",
														},
														Value: &ast.ArrayExpression{
															BaseNode: ast.BaseNode{
																Errors: nil,
																Loc: &ast.SourceLocation{
																	End: ast.Position{
																		Column: 67,
																		Line:   12,
																	},
																	File:   "",
																	Source: "[\"_start\", \"_stop\"]",
																	Start: ast.Position{
																		Column: 48,
																		Line:   12,
																	},
																},
															},
															Elements: []ast.Expression{&ast.StringLiteral{
																BaseNode: ast.BaseNode{
																	Errors: nil,
																	Loc: &ast.SourceLocation{
																		End: ast.Position{
																			Column: 57,
																			Line:   12,
																		},
																		File:   "",
																		Source: "\"_start\"",
																		Start: ast.Position{
																			Column: 49,
																			Line:   12,
																		},
																	},
																},
																Value: "_start",
															}, &ast.StringLiteral{
																BaseNode: ast.BaseNode{
																	Errors: nil,
																	Loc: &ast.SourceLocation{
																		End: ast.Position{
																			Column: 66,
																			Line:   12,
																		},
																		File:   "",
																		Source: "\"_stop\"",
																		Start: ast.Position{
																			Column: 59,
																			Line:   12,
																		},
																	},
																},
																Value: "_stop",
															}},
														},
													}, &ast.Property{
														BaseNode: ast.BaseNode{
															Errors: nil,
															Loc: &ast.SourceLocation{
																End: ast.Position{
																	Column: 82,
																	Line:   12,
																},
																File:   "",
																Source: "mode:\"extend\"",
																Start: ast.Position{
																	Column: 69,
																	Line:   12,
																},
															},
														},
														Key: &ast.Identifier{
															BaseNode: ast.BaseNode{
																Errors: nil,
																Loc: &ast.SourceLocation{
																	End: ast.Position{
																		Column: 73,
																		Line:   12,
																	},
																	File:   "",
																	Source: "mode",
																	Start: ast.Position{
																		Column: 69,
																		Line:   12,
																	},
																},
															},
															Name: "mode",
														},
														Value: &ast.StringLiteral{
															BaseNode: ast.BaseNode{
																Errors: nil,
																Loc: &ast.SourceLocation{
																	End: ast.Position{
																		Column: 82,
																		Line:   12,
																	},
																	File:   "",
																	Source: "\"extend\"",
																	Start: ast.Position{
																		Column: 74,
																		Line:   12,
																	},
																},
															},
															Value: "extend",
														},
													}},
													With: nil,
												}},
												BaseNode: ast.BaseNode{
													Errors: nil,
													Loc: &ast.SourceLocation{
														End: ast.Position{
															Column: 83,
															Line:   12,
														},
														File:   "",
														Source: "experimental.group(columns: [\"_start\", \"_stop\"], mode:\"extend\")",
														Start: ast.Position{
															Column: 20,
															Line:   12,
														},
													},
												},
												Callee: &ast.MemberExpression{
													BaseNode: ast.BaseNode{
														Errors: nil,
														Loc: &ast.SourceLocation{
															End: ast.Position{
																Column: 38,
																Line:   12,
															},
															File:   "",
															Source: "experimental.group",
															Start: ast.Position{
																Column: 20,
																Line:   12,
															},
														},
													},
													Object: &ast.Identifier{
														BaseNode: ast.BaseNode{
															Errors: nil,
															Loc: &ast.SourceLocation{
																End: ast.Position{
																	Column: 32,
																	Line:   12,
																},
																File:   "",
																Source: "experimental",
																Start: ast.Position{
																	Column: 20,
																	Line:   12,
																},
															},
														},
														Name: "experimental",
													},
													Property: &ast.Identifier{
														BaseNode: ast.BaseNode{
															Errors: nil,
															Loc: &ast.SourceLocation{
																End: ast.Position{
																	Column: 38,
																	Line:   12,
																},
																File:   "",
																Source: "group",
																Start: ast.Position{
																	Column: 33,
																	Line:   12,
																},
															},
														},
														Name: "group",
													},
												},
											},
										},
										BaseNode: ast.BaseNode{
											Errors: nil,
											Loc: &ast.SourceLocation{
												End: ast.Position{
													Column: 25,
													Line:   13,
												},
												File:   "",
												Source: "tables\n                |> mean(column: column)\n                |> group(columns: groupColumns)\n                |> experimental.group(columns: [\"_start\", \"_stop\"], mode:\"extend\")\n                |> sum()",
												Start: ast.Position{
													Column: 13,
													Line:   9,
												},
											},
										},
										Call: &ast.CallExpression{
											Arguments: nil,
											BaseNode: ast.BaseNode{
												Errors: nil,
												Loc: &ast.SourceLocation{
													End: ast.Position{
														Column: 25,
														Line:   13,
													},
													File:   "",
													Source: "sum()",
													Start: ast.Position{
														Column: 20,
														Line:   13,
													},
												},
											},
											Callee: &ast.Identifier{
												BaseNode: ast.BaseNode{
													Errors: nil,
													Loc: &ast.SourceLocation{
														End: ast.Position{
															Column: 23,
															Line:   13,
														},
														File:   "",
														Source: "sum",
														Start: ast.Position{
															Column: 20,
															Line:   13,
														},
													},
												},
												Name: "sum",
											},
										},
									},
									Params: []*ast.Property{&ast.Property{
										BaseNode: ast.BaseNode{
											Errors: nil,
											Loc: &ast.SourceLocation{
												End: ast.Position{
													Column: 57,
													Line:   8,
												},
												File:   "",
												Source: "tables=<-",
												Start: ast.Position{
													Column: 48,
													Line:   8,
												},
											},
										},
										Key: &ast.Identifier{
											BaseNode: ast.BaseNode{
												Errors: nil,
												Loc: &ast.SourceLocation{
													End: ast.Position{
														Column: 54,
														Line:   8,
													},
													File:   "",
													Source: "tables",
													Start: ast.Position{
														Column: 48,
														Line:   8,
													},
												},
											},
											Name: "tables",
										},
										Value: &ast.PipeLiteral{BaseNode: ast.BaseNode{
											Errors: nil,
											Loc: &ast.SourceLocation{
												End: ast.Position{
													Column: 57,
													Line:   8,
												},
												File:   "",
												Source: "<-",
												Start: ast.Position{
													Column: 55,
													Line:   8,
												},
											},
										}},
									}, &ast.Property{
										BaseNode: ast.BaseNode{
											Errors: nil,
											Loc: &ast.SourceLocation{
												End: ast.Position{
													Column: 65,
													Line:   8,
												},
												File:   "",
												Source: "column",
												Start: ast.Position{
													Column: 59,
													Line:   8,
												},
											},
										},
										Key: &ast.Identifier{
											BaseNode: ast.BaseNode{
												Errors: nil,
												Loc: &ast.SourceLocation{
													End: ast.Position{
														Column: 65,
														Line:   8,
													},
													File:   "",
													Source: "column",
													Start: ast.Position{
														Column: 59,
														Line:   8,
													},
												},
											},
											Name: "column",
										},
										Value: nil,
									}},
								},
							}},
							With: nil,
						}},
						BaseNode: ast.BaseNode{
							Errors: nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 10,
									Line:   14,
								},
								File:   "",
								Source: "aggregateWindow(every: every, fn : (tables=<-, column) =>\n            tables\n                |> mean(column: column)\n                |> group(columns: groupColumns)\n                |> experimental.group(columns: [\"_start\", \"_stop\"], mode:\"extend\")\n                |> sum()\n        )",
								Start: ast.Position{
									Column: 12,
									Line:   8,
								},
							},
						},
						Callee: &ast.Identifier{
							BaseNode: ast.BaseNode{
								Errors: nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 27,
										Line:   8,
									},
									File:   "",
									Source: "aggregateWindow",
									Start: ast.Position{
										Column: 12,
										Line:   8,
									},
								},
							},
							Name: "aggregateWindow",
						},
					},
				},
				Params: []*ast.Property{&ast.Property{
					BaseNode: ast.BaseNode{
						Errors: nil,
						Loc: &ast.SourceLocation{
							End: ast.Position{
								Column: 18,
								Line:   5,
							},
							File:   "",
							Source: "tables=<-",
							Start: ast.Position{
								Column: 9,
								Line:   5,
							},
						},
					},
					Key: &ast.Identifier{
						BaseNode: ast.BaseNode{
							Errors: nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 15,
									Line:   5,
								},
								File:   "",
								Source: "tables",
								Start: ast.Position{
									Column: 9,
									Line:   5,
								},
							},
						},
						Name: "tables",
					},
					Value: &ast.PipeLiteral{BaseNode: ast.BaseNode{
						Errors: nil,
						Loc: &ast.SourceLocation{
							End: ast.Position{
								Column: 18,
								Line:   5,
							},
							File:   "",
							Source: "<-",
							Start: ast.Position{
								Column: 16,
								Line:   5,
							},
						},
					}},
				}, &ast.Property{
					BaseNode: ast.BaseNode{
						Errors: nil,
						Loc: &ast.SourceLocation{
							End: ast.Position{
								Column: 25,
								Line:   5,
							},
							File:   "",
							Source: "every",
							Start: ast.Position{
								Column: 20,
								Line:   5,
							},
						},
					},
					Key: &ast.Identifier{
						BaseNode: ast.BaseNode{
							Errors: nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 25,
									Line:   5,
								},
								File:   "",
								Source: "every",
								Start: ast.Position{
									Column: 20,
									Line:   5,
								},
							},
						},
						Name: "every",
					},
					Value: nil,
				}, &ast.Property{
					BaseNode: ast.BaseNode{
						Errors: nil,
						Loc: &ast.SourceLocation{
							End: ast.Position{
								Column: 42,
								Line:   5,
							},
							File:   "",
							Source: "groupColumns=[]",
							Start: ast.Position{
								Column: 27,
								Line:   5,
							},
						},
					},
					Key: &ast.Identifier{
						BaseNode: ast.BaseNode{
							Errors: nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 39,
									Line:   5,
								},
								File:   "",
								Source: "groupColumns",
								Start: ast.Position{
									Column: 27,
									Line:   5,
								},
							},
						},
						Name: "groupColumns",
					},
					Value: &ast.ArrayExpression{
						BaseNode: ast.BaseNode{
							Errors: nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 42,
									Line:   5,
								},
								File:   "",
								Source: "[]",
								Start: ast.Position{
									Column: 40,
									Line:   5,
								},
							},
						},
						Elements: []ast.Expression{},
					},
				}, &ast.Property{
					BaseNode: ast.BaseNode{
						Errors: nil,
						Loc: &ast.SourceLocation{
							End: ast.Position{
								Column: 51,
								Line:   5,
							},
							File:   "",
							Source: "unit=1s",
							Start: ast.Position{
								Column: 44,
								Line:   5,
							},
						},
					},
					Key: &ast.Identifier{
						BaseNode: ast.BaseNode{
							Errors: nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 48,
									Line:   5,
								},
								File:   "",
								Source: "unit",
								Start: ast.Position{
									Column: 44,
									Line:   5,
								},
							},
						},
						Name: "unit",
					},
					Value: &ast.DurationLiteral{
						BaseNode: ast.BaseNode{
							Errors: nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 51,
									Line:   5,
								},
								File:   "",
								Source: "1s",
								Start: ast.Position{
									Column: 49,
									Line:   5,
								},
							},
						},
						Values: []ast.Duration{ast.Duration{
							Magnitude: int64(1),
							Unit:      "s",
						}},
					},
				}},
			},
		}},
		Imports: []*ast.ImportDeclaration{&ast.ImportDeclaration{
			As: nil,
			BaseNode: ast.BaseNode{
				Errors: nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 22,
						Line:   3,
					},
					File:   "",
					Source: "import \"experimental\"",
					Start: ast.Position{
						Column: 1,
						Line:   3,
					},
				},
			},
			Path: &ast.StringLiteral{
				BaseNode: ast.BaseNode{
					Errors: nil,
					Loc: &ast.SourceLocation{
						End: ast.Position{
							Column: 22,
							Line:   3,
						},
						File:   "",
						Source: "\"experimental\"",
						Start: ast.Position{
							Column: 8,
							Line:   3,
						},
					},
				},
				Value: "experimental",
			},
		}},
		Metadata: "parser-type=rust",
		Name:     "aggregate.flux",
		Package: &ast.PackageClause{
			BaseNode: ast.BaseNode{
				Errors: nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 18,
						Line:   1,
					},
					File:   "",
					Source: "package aggregate",
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
							Column: 18,
							Line:   1,
						},
						File:   "",
						Source: "aggregate",
						Start: ast.Position{
							Column: 9,
							Line:   1,
						},
					},
				},
				Name: "aggregate",
			},
		},
	}},
	Package: "aggregate",
	Path:    "experimental/aggregate",
}
