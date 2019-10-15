// DO NOT EDIT: This file is autogenerated via the builtin command.

package prometheus

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
					Column: 58,
					Line:   17,
				},
				File:   "prometheus.flux",
				Source: "package prometheus\nimport \"universe\" \n\n// scrape enables scraping of a prometheus metrics endpoint and converts \n// that input into flux tables. Each metric is put into an individual flux \n// table, including each histogram and summary value.  \nbuiltin scrape\n\n// histogramQuantile enables the user to calculate quantiles on a set of given values\n// This function assumes that the given histogram data is being scraped or read from a \n// Prometheus source. \nhistogramQuantile = (tables=<-, quantile) => \n    tables\n        |> filter(fn: (r) => r._measurement == \"prometheus\")\n        |> group(mode: \"except\", columns: [\"le\", \"_value\", \"_time\"]) \n        |> map(fn:(r) => ({r with le: float(v:r.le)})) \n        |> universe.histogramQuantile(quantile: quantile)",
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
						Column: 15,
						Line:   7,
					},
					File:   "prometheus.flux",
					Source: "builtin scrape",
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
							Column: 15,
							Line:   7,
						},
						File:   "prometheus.flux",
						Source: "scrape",
						Start: ast.Position{
							Column: 9,
							Line:   7,
						},
					},
				},
				Name: "scrape",
			},
		}, &ast.VariableAssignment{
			BaseNode: ast.BaseNode{
				Errors: nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 58,
						Line:   17,
					},
					File:   "prometheus.flux",
					Source: "histogramQuantile = (tables=<-, quantile) => \n    tables\n        |> filter(fn: (r) => r._measurement == \"prometheus\")\n        |> group(mode: \"except\", columns: [\"le\", \"_value\", \"_time\"]) \n        |> map(fn:(r) => ({r with le: float(v:r.le)})) \n        |> universe.histogramQuantile(quantile: quantile)",
					Start: ast.Position{
						Column: 1,
						Line:   12,
					},
				},
			},
			ID: &ast.Identifier{
				BaseNode: ast.BaseNode{
					Errors: nil,
					Loc: &ast.SourceLocation{
						End: ast.Position{
							Column: 18,
							Line:   12,
						},
						File:   "prometheus.flux",
						Source: "histogramQuantile",
						Start: ast.Position{
							Column: 1,
							Line:   12,
						},
					},
				},
				Name: "histogramQuantile",
			},
			Init: &ast.FunctionExpression{
				BaseNode: ast.BaseNode{
					Errors: nil,
					Loc: &ast.SourceLocation{
						End: ast.Position{
							Column: 58,
							Line:   17,
						},
						File:   "prometheus.flux",
						Source: "(tables=<-, quantile) => \n    tables\n        |> filter(fn: (r) => r._measurement == \"prometheus\")\n        |> group(mode: \"except\", columns: [\"le\", \"_value\", \"_time\"]) \n        |> map(fn:(r) => ({r with le: float(v:r.le)})) \n        |> universe.histogramQuantile(quantile: quantile)",
						Start: ast.Position{
							Column: 21,
							Line:   12,
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
												Column: 11,
												Line:   13,
											},
											File:   "prometheus.flux",
											Source: "tables",
											Start: ast.Position{
												Column: 5,
												Line:   13,
											},
										},
									},
									Name: "tables",
								},
								BaseNode: ast.BaseNode{
									Errors: nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 61,
											Line:   14,
										},
										File:   "prometheus.flux",
										Source: "tables\n        |> filter(fn: (r) => r._measurement == \"prometheus\")",
										Start: ast.Position{
											Column: 5,
											Line:   13,
										},
									},
								},
								Call: &ast.CallExpression{
									Arguments: []ast.Expression{&ast.ObjectExpression{
										BaseNode: ast.BaseNode{
											Errors: nil,
											Loc: &ast.SourceLocation{
												End: ast.Position{
													Column: 60,
													Line:   14,
												},
												File:   "prometheus.flux",
												Source: "fn: (r) => r._measurement == \"prometheus\"",
												Start: ast.Position{
													Column: 19,
													Line:   14,
												},
											},
										},
										Properties: []*ast.Property{&ast.Property{
											BaseNode: ast.BaseNode{
												Errors: nil,
												Loc: &ast.SourceLocation{
													End: ast.Position{
														Column: 60,
														Line:   14,
													},
													File:   "prometheus.flux",
													Source: "fn: (r) => r._measurement == \"prometheus\"",
													Start: ast.Position{
														Column: 19,
														Line:   14,
													},
												},
											},
											Key: &ast.Identifier{
												BaseNode: ast.BaseNode{
													Errors: nil,
													Loc: &ast.SourceLocation{
														End: ast.Position{
															Column: 21,
															Line:   14,
														},
														File:   "prometheus.flux",
														Source: "fn",
														Start: ast.Position{
															Column: 19,
															Line:   14,
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
															Column: 60,
															Line:   14,
														},
														File:   "prometheus.flux",
														Source: "(r) => r._measurement == \"prometheus\"",
														Start: ast.Position{
															Column: 23,
															Line:   14,
														},
													},
												},
												Body: &ast.BinaryExpression{
													BaseNode: ast.BaseNode{
														Errors: nil,
														Loc: &ast.SourceLocation{
															End: ast.Position{
																Column: 60,
																Line:   14,
															},
															File:   "prometheus.flux",
															Source: "r._measurement == \"prometheus\"",
															Start: ast.Position{
																Column: 30,
																Line:   14,
															},
														},
													},
													Left: &ast.MemberExpression{
														BaseNode: ast.BaseNode{
															Errors: nil,
															Loc: &ast.SourceLocation{
																End: ast.Position{
																	Column: 44,
																	Line:   14,
																},
																File:   "prometheus.flux",
																Source: "r._measurement",
																Start: ast.Position{
																	Column: 30,
																	Line:   14,
																},
															},
														},
														Object: &ast.Identifier{
															BaseNode: ast.BaseNode{
																Errors: nil,
																Loc: &ast.SourceLocation{
																	End: ast.Position{
																		Column: 31,
																		Line:   14,
																	},
																	File:   "prometheus.flux",
																	Source: "r",
																	Start: ast.Position{
																		Column: 30,
																		Line:   14,
																	},
																},
															},
															Name: "r",
														},
														Property: &ast.Identifier{
															BaseNode: ast.BaseNode{
																Errors: nil,
																Loc: &ast.SourceLocation{
																	End: ast.Position{
																		Column: 44,
																		Line:   14,
																	},
																	File:   "prometheus.flux",
																	Source: "_measurement",
																	Start: ast.Position{
																		Column: 32,
																		Line:   14,
																	},
																},
															},
															Name: "_measurement",
														},
													},
													Operator: 17,
													Right: &ast.StringLiteral{
														BaseNode: ast.BaseNode{
															Errors: nil,
															Loc: &ast.SourceLocation{
																End: ast.Position{
																	Column: 60,
																	Line:   14,
																},
																File:   "prometheus.flux",
																Source: "\"prometheus\"",
																Start: ast.Position{
																	Column: 48,
																	Line:   14,
																},
															},
														},
														Value: "prometheus",
													},
												},
												Params: []*ast.Property{&ast.Property{
													BaseNode: ast.BaseNode{
														Errors: nil,
														Loc: &ast.SourceLocation{
															End: ast.Position{
																Column: 25,
																Line:   14,
															},
															File:   "prometheus.flux",
															Source: "r",
															Start: ast.Position{
																Column: 24,
																Line:   14,
															},
														},
													},
													Key: &ast.Identifier{
														BaseNode: ast.BaseNode{
															Errors: nil,
															Loc: &ast.SourceLocation{
																End: ast.Position{
																	Column: 25,
																	Line:   14,
																},
																File:   "prometheus.flux",
																Source: "r",
																Start: ast.Position{
																	Column: 24,
																	Line:   14,
																},
															},
														},
														Name: "r",
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
												Column: 61,
												Line:   14,
											},
											File:   "prometheus.flux",
											Source: "filter(fn: (r) => r._measurement == \"prometheus\")",
											Start: ast.Position{
												Column: 12,
												Line:   14,
											},
										},
									},
									Callee: &ast.Identifier{
										BaseNode: ast.BaseNode{
											Errors: nil,
											Loc: &ast.SourceLocation{
												End: ast.Position{
													Column: 18,
													Line:   14,
												},
												File:   "prometheus.flux",
												Source: "filter",
												Start: ast.Position{
													Column: 12,
													Line:   14,
												},
											},
										},
										Name: "filter",
									},
								},
							},
							BaseNode: ast.BaseNode{
								Errors: nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 69,
										Line:   15,
									},
									File:   "prometheus.flux",
									Source: "tables\n        |> filter(fn: (r) => r._measurement == \"prometheus\")\n        |> group(mode: \"except\", columns: [\"le\", \"_value\", \"_time\"])",
									Start: ast.Position{
										Column: 5,
										Line:   13,
									},
								},
							},
							Call: &ast.CallExpression{
								Arguments: []ast.Expression{&ast.ObjectExpression{
									BaseNode: ast.BaseNode{
										Errors: nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 68,
												Line:   15,
											},
											File:   "prometheus.flux",
											Source: "mode: \"except\", columns: [\"le\", \"_value\", \"_time\"]",
											Start: ast.Position{
												Column: 18,
												Line:   15,
											},
										},
									},
									Properties: []*ast.Property{&ast.Property{
										BaseNode: ast.BaseNode{
											Errors: nil,
											Loc: &ast.SourceLocation{
												End: ast.Position{
													Column: 32,
													Line:   15,
												},
												File:   "prometheus.flux",
												Source: "mode: \"except\"",
												Start: ast.Position{
													Column: 18,
													Line:   15,
												},
											},
										},
										Key: &ast.Identifier{
											BaseNode: ast.BaseNode{
												Errors: nil,
												Loc: &ast.SourceLocation{
													End: ast.Position{
														Column: 22,
														Line:   15,
													},
													File:   "prometheus.flux",
													Source: "mode",
													Start: ast.Position{
														Column: 18,
														Line:   15,
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
														Column: 32,
														Line:   15,
													},
													File:   "prometheus.flux",
													Source: "\"except\"",
													Start: ast.Position{
														Column: 24,
														Line:   15,
													},
												},
											},
											Value: "except",
										},
									}, &ast.Property{
										BaseNode: ast.BaseNode{
											Errors: nil,
											Loc: &ast.SourceLocation{
												End: ast.Position{
													Column: 68,
													Line:   15,
												},
												File:   "prometheus.flux",
												Source: "columns: [\"le\", \"_value\", \"_time\"]",
												Start: ast.Position{
													Column: 34,
													Line:   15,
												},
											},
										},
										Key: &ast.Identifier{
											BaseNode: ast.BaseNode{
												Errors: nil,
												Loc: &ast.SourceLocation{
													End: ast.Position{
														Column: 41,
														Line:   15,
													},
													File:   "prometheus.flux",
													Source: "columns",
													Start: ast.Position{
														Column: 34,
														Line:   15,
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
														Column: 68,
														Line:   15,
													},
													File:   "prometheus.flux",
													Source: "[\"le\", \"_value\", \"_time\"]",
													Start: ast.Position{
														Column: 43,
														Line:   15,
													},
												},
											},
											Elements: []ast.Expression{&ast.StringLiteral{
												BaseNode: ast.BaseNode{
													Errors: nil,
													Loc: &ast.SourceLocation{
														End: ast.Position{
															Column: 48,
															Line:   15,
														},
														File:   "prometheus.flux",
														Source: "\"le\"",
														Start: ast.Position{
															Column: 44,
															Line:   15,
														},
													},
												},
												Value: "le",
											}, &ast.StringLiteral{
												BaseNode: ast.BaseNode{
													Errors: nil,
													Loc: &ast.SourceLocation{
														End: ast.Position{
															Column: 58,
															Line:   15,
														},
														File:   "prometheus.flux",
														Source: "\"_value\"",
														Start: ast.Position{
															Column: 50,
															Line:   15,
														},
													},
												},
												Value: "_value",
											}, &ast.StringLiteral{
												BaseNode: ast.BaseNode{
													Errors: nil,
													Loc: &ast.SourceLocation{
														End: ast.Position{
															Column: 67,
															Line:   15,
														},
														File:   "prometheus.flux",
														Source: "\"_time\"",
														Start: ast.Position{
															Column: 60,
															Line:   15,
														},
													},
												},
												Value: "_time",
											}},
										},
									}},
									With: nil,
								}},
								BaseNode: ast.BaseNode{
									Errors: nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 69,
											Line:   15,
										},
										File:   "prometheus.flux",
										Source: "group(mode: \"except\", columns: [\"le\", \"_value\", \"_time\"])",
										Start: ast.Position{
											Column: 12,
											Line:   15,
										},
									},
								},
								Callee: &ast.Identifier{
									BaseNode: ast.BaseNode{
										Errors: nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 17,
												Line:   15,
											},
											File:   "prometheus.flux",
											Source: "group",
											Start: ast.Position{
												Column: 12,
												Line:   15,
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
									Column: 55,
									Line:   16,
								},
								File:   "prometheus.flux",
								Source: "tables\n        |> filter(fn: (r) => r._measurement == \"prometheus\")\n        |> group(mode: \"except\", columns: [\"le\", \"_value\", \"_time\"]) \n        |> map(fn:(r) => ({r with le: float(v:r.le)}))",
								Start: ast.Position{
									Column: 5,
									Line:   13,
								},
							},
						},
						Call: &ast.CallExpression{
							Arguments: []ast.Expression{&ast.ObjectExpression{
								BaseNode: ast.BaseNode{
									Errors: nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 54,
											Line:   16,
										},
										File:   "prometheus.flux",
										Source: "fn:(r) => ({r with le: float(v:r.le)})",
										Start: ast.Position{
											Column: 16,
											Line:   16,
										},
									},
								},
								Properties: []*ast.Property{&ast.Property{
									BaseNode: ast.BaseNode{
										Errors: nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 54,
												Line:   16,
											},
											File:   "prometheus.flux",
											Source: "fn:(r) => ({r with le: float(v:r.le)})",
											Start: ast.Position{
												Column: 16,
												Line:   16,
											},
										},
									},
									Key: &ast.Identifier{
										BaseNode: ast.BaseNode{
											Errors: nil,
											Loc: &ast.SourceLocation{
												End: ast.Position{
													Column: 18,
													Line:   16,
												},
												File:   "prometheus.flux",
												Source: "fn",
												Start: ast.Position{
													Column: 16,
													Line:   16,
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
													Column: 54,
													Line:   16,
												},
												File:   "prometheus.flux",
												Source: "(r) => ({r with le: float(v:r.le)})",
												Start: ast.Position{
													Column: 19,
													Line:   16,
												},
											},
										},
										Body: &ast.ParenExpression{
											BaseNode: ast.BaseNode{
												Errors: nil,
												Loc: &ast.SourceLocation{
													End: ast.Position{
														Column: 54,
														Line:   16,
													},
													File:   "prometheus.flux",
													Source: "({r with le: float(v:r.le)})",
													Start: ast.Position{
														Column: 26,
														Line:   16,
													},
												},
											},
											Expression: &ast.ObjectExpression{
												BaseNode: ast.BaseNode{
													Errors: nil,
													Loc: &ast.SourceLocation{
														End: ast.Position{
															Column: 53,
															Line:   16,
														},
														File:   "prometheus.flux",
														Source: "{r with le: float(v:r.le)}",
														Start: ast.Position{
															Column: 27,
															Line:   16,
														},
													},
												},
												Properties: []*ast.Property{&ast.Property{
													BaseNode: ast.BaseNode{
														Errors: nil,
														Loc: &ast.SourceLocation{
															End: ast.Position{
																Column: 52,
																Line:   16,
															},
															File:   "prometheus.flux",
															Source: "le: float(v:r.le)",
															Start: ast.Position{
																Column: 35,
																Line:   16,
															},
														},
													},
													Key: &ast.Identifier{
														BaseNode: ast.BaseNode{
															Errors: nil,
															Loc: &ast.SourceLocation{
																End: ast.Position{
																	Column: 37,
																	Line:   16,
																},
																File:   "prometheus.flux",
																Source: "le",
																Start: ast.Position{
																	Column: 35,
																	Line:   16,
																},
															},
														},
														Name: "le",
													},
													Value: &ast.CallExpression{
														Arguments: []ast.Expression{&ast.ObjectExpression{
															BaseNode: ast.BaseNode{
																Errors: nil,
																Loc: &ast.SourceLocation{
																	End: ast.Position{
																		Column: 51,
																		Line:   16,
																	},
																	File:   "prometheus.flux",
																	Source: "v:r.le",
																	Start: ast.Position{
																		Column: 45,
																		Line:   16,
																	},
																},
															},
															Properties: []*ast.Property{&ast.Property{
																BaseNode: ast.BaseNode{
																	Errors: nil,
																	Loc: &ast.SourceLocation{
																		End: ast.Position{
																			Column: 51,
																			Line:   16,
																		},
																		File:   "prometheus.flux",
																		Source: "v:r.le",
																		Start: ast.Position{
																			Column: 45,
																			Line:   16,
																		},
																	},
																},
																Key: &ast.Identifier{
																	BaseNode: ast.BaseNode{
																		Errors: nil,
																		Loc: &ast.SourceLocation{
																			End: ast.Position{
																				Column: 46,
																				Line:   16,
																			},
																			File:   "prometheus.flux",
																			Source: "v",
																			Start: ast.Position{
																				Column: 45,
																				Line:   16,
																			},
																		},
																	},
																	Name: "v",
																},
																Value: &ast.MemberExpression{
																	BaseNode: ast.BaseNode{
																		Errors: nil,
																		Loc: &ast.SourceLocation{
																			End: ast.Position{
																				Column: 51,
																				Line:   16,
																			},
																			File:   "prometheus.flux",
																			Source: "r.le",
																			Start: ast.Position{
																				Column: 47,
																				Line:   16,
																			},
																		},
																	},
																	Object: &ast.Identifier{
																		BaseNode: ast.BaseNode{
																			Errors: nil,
																			Loc: &ast.SourceLocation{
																				End: ast.Position{
																					Column: 48,
																					Line:   16,
																				},
																				File:   "prometheus.flux",
																				Source: "r",
																				Start: ast.Position{
																					Column: 47,
																					Line:   16,
																				},
																			},
																		},
																		Name: "r",
																	},
																	Property: &ast.Identifier{
																		BaseNode: ast.BaseNode{
																			Errors: nil,
																			Loc: &ast.SourceLocation{
																				End: ast.Position{
																					Column: 51,
																					Line:   16,
																				},
																				File:   "prometheus.flux",
																				Source: "le",
																				Start: ast.Position{
																					Column: 49,
																					Line:   16,
																				},
																			},
																		},
																		Name: "le",
																	},
																},
															}},
															With: nil,
														}},
														BaseNode: ast.BaseNode{
															Errors: nil,
															Loc: &ast.SourceLocation{
																End: ast.Position{
																	Column: 52,
																	Line:   16,
																},
																File:   "prometheus.flux",
																Source: "float(v:r.le)",
																Start: ast.Position{
																	Column: 39,
																	Line:   16,
																},
															},
														},
														Callee: &ast.Identifier{
															BaseNode: ast.BaseNode{
																Errors: nil,
																Loc: &ast.SourceLocation{
																	End: ast.Position{
																		Column: 44,
																		Line:   16,
																	},
																	File:   "prometheus.flux",
																	Source: "float",
																	Start: ast.Position{
																		Column: 39,
																		Line:   16,
																	},
																},
															},
															Name: "float",
														},
													},
												}},
												With: &ast.Identifier{
													BaseNode: ast.BaseNode{
														Errors: nil,
														Loc: &ast.SourceLocation{
															End: ast.Position{
																Column: 29,
																Line:   16,
															},
															File:   "prometheus.flux",
															Source: "r",
															Start: ast.Position{
																Column: 28,
																Line:   16,
															},
														},
													},
													Name: "r",
												},
											},
										},
										Params: []*ast.Property{&ast.Property{
											BaseNode: ast.BaseNode{
												Errors: nil,
												Loc: &ast.SourceLocation{
													End: ast.Position{
														Column: 21,
														Line:   16,
													},
													File:   "prometheus.flux",
													Source: "r",
													Start: ast.Position{
														Column: 20,
														Line:   16,
													},
												},
											},
											Key: &ast.Identifier{
												BaseNode: ast.BaseNode{
													Errors: nil,
													Loc: &ast.SourceLocation{
														End: ast.Position{
															Column: 21,
															Line:   16,
														},
														File:   "prometheus.flux",
														Source: "r",
														Start: ast.Position{
															Column: 20,
															Line:   16,
														},
													},
												},
												Name: "r",
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
										Column: 55,
										Line:   16,
									},
									File:   "prometheus.flux",
									Source: "map(fn:(r) => ({r with le: float(v:r.le)}))",
									Start: ast.Position{
										Column: 12,
										Line:   16,
									},
								},
							},
							Callee: &ast.Identifier{
								BaseNode: ast.BaseNode{
									Errors: nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 15,
											Line:   16,
										},
										File:   "prometheus.flux",
										Source: "map",
										Start: ast.Position{
											Column: 12,
											Line:   16,
										},
									},
								},
								Name: "map",
							},
						},
					},
					BaseNode: ast.BaseNode{
						Errors: nil,
						Loc: &ast.SourceLocation{
							End: ast.Position{
								Column: 58,
								Line:   17,
							},
							File:   "prometheus.flux",
							Source: "tables\n        |> filter(fn: (r) => r._measurement == \"prometheus\")\n        |> group(mode: \"except\", columns: [\"le\", \"_value\", \"_time\"]) \n        |> map(fn:(r) => ({r with le: float(v:r.le)})) \n        |> universe.histogramQuantile(quantile: quantile)",
							Start: ast.Position{
								Column: 5,
								Line:   13,
							},
						},
					},
					Call: &ast.CallExpression{
						Arguments: []ast.Expression{&ast.ObjectExpression{
							BaseNode: ast.BaseNode{
								Errors: nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 57,
										Line:   17,
									},
									File:   "prometheus.flux",
									Source: "quantile: quantile",
									Start: ast.Position{
										Column: 39,
										Line:   17,
									},
								},
							},
							Properties: []*ast.Property{&ast.Property{
								BaseNode: ast.BaseNode{
									Errors: nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 57,
											Line:   17,
										},
										File:   "prometheus.flux",
										Source: "quantile: quantile",
										Start: ast.Position{
											Column: 39,
											Line:   17,
										},
									},
								},
								Key: &ast.Identifier{
									BaseNode: ast.BaseNode{
										Errors: nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 47,
												Line:   17,
											},
											File:   "prometheus.flux",
											Source: "quantile",
											Start: ast.Position{
												Column: 39,
												Line:   17,
											},
										},
									},
									Name: "quantile",
								},
								Value: &ast.Identifier{
									BaseNode: ast.BaseNode{
										Errors: nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 57,
												Line:   17,
											},
											File:   "prometheus.flux",
											Source: "quantile",
											Start: ast.Position{
												Column: 49,
												Line:   17,
											},
										},
									},
									Name: "quantile",
								},
							}},
							With: nil,
						}},
						BaseNode: ast.BaseNode{
							Errors: nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 58,
									Line:   17,
								},
								File:   "prometheus.flux",
								Source: "universe.histogramQuantile(quantile: quantile)",
								Start: ast.Position{
									Column: 12,
									Line:   17,
								},
							},
						},
						Callee: &ast.MemberExpression{
							BaseNode: ast.BaseNode{
								Errors: nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 38,
										Line:   17,
									},
									File:   "prometheus.flux",
									Source: "universe.histogramQuantile",
									Start: ast.Position{
										Column: 12,
										Line:   17,
									},
								},
							},
							Object: &ast.Identifier{
								BaseNode: ast.BaseNode{
									Errors: nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 20,
											Line:   17,
										},
										File:   "prometheus.flux",
										Source: "universe",
										Start: ast.Position{
											Column: 12,
											Line:   17,
										},
									},
								},
								Name: "universe",
							},
							Property: &ast.Identifier{
								BaseNode: ast.BaseNode{
									Errors: nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 38,
											Line:   17,
										},
										File:   "prometheus.flux",
										Source: "histogramQuantile",
										Start: ast.Position{
											Column: 21,
											Line:   17,
										},
									},
								},
								Name: "histogramQuantile",
							},
						},
					},
				},
				Params: []*ast.Property{&ast.Property{
					BaseNode: ast.BaseNode{
						Errors: nil,
						Loc: &ast.SourceLocation{
							End: ast.Position{
								Column: 31,
								Line:   12,
							},
							File:   "prometheus.flux",
							Source: "tables=<-",
							Start: ast.Position{
								Column: 22,
								Line:   12,
							},
						},
					},
					Key: &ast.Identifier{
						BaseNode: ast.BaseNode{
							Errors: nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 28,
									Line:   12,
								},
								File:   "prometheus.flux",
								Source: "tables",
								Start: ast.Position{
									Column: 22,
									Line:   12,
								},
							},
						},
						Name: "tables",
					},
					Value: &ast.PipeLiteral{BaseNode: ast.BaseNode{
						Errors: nil,
						Loc: &ast.SourceLocation{
							End: ast.Position{
								Column: 31,
								Line:   12,
							},
							File:   "prometheus.flux",
							Source: "<-",
							Start: ast.Position{
								Column: 29,
								Line:   12,
							},
						},
					}},
				}, &ast.Property{
					BaseNode: ast.BaseNode{
						Errors: nil,
						Loc: &ast.SourceLocation{
							End: ast.Position{
								Column: 41,
								Line:   12,
							},
							File:   "prometheus.flux",
							Source: "quantile",
							Start: ast.Position{
								Column: 33,
								Line:   12,
							},
						},
					},
					Key: &ast.Identifier{
						BaseNode: ast.BaseNode{
							Errors: nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 41,
									Line:   12,
								},
								File:   "prometheus.flux",
								Source: "quantile",
								Start: ast.Position{
									Column: 33,
									Line:   12,
								},
							},
						},
						Name: "quantile",
					},
					Value: nil,
				}},
			},
		}},
		Imports: []*ast.ImportDeclaration{&ast.ImportDeclaration{
			As: nil,
			BaseNode: ast.BaseNode{
				Errors: nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 18,
						Line:   2,
					},
					File:   "prometheus.flux",
					Source: "import \"universe\"",
					Start: ast.Position{
						Column: 1,
						Line:   2,
					},
				},
			},
			Path: &ast.StringLiteral{
				BaseNode: ast.BaseNode{
					Errors: nil,
					Loc: &ast.SourceLocation{
						End: ast.Position{
							Column: 18,
							Line:   2,
						},
						File:   "prometheus.flux",
						Source: "\"universe\"",
						Start: ast.Position{
							Column: 8,
							Line:   2,
						},
					},
				},
				Value: "universe",
			},
		}},
		Name: "prometheus.flux",
		Package: &ast.PackageClause{
			BaseNode: ast.BaseNode{
				Errors: nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 19,
						Line:   1,
					},
					File:   "prometheus.flux",
					Source: "package prometheus",
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
							Column: 19,
							Line:   1,
						},
						File:   "prometheus.flux",
						Source: "prometheus",
						Start: ast.Position{
							Column: 9,
							Line:   1,
						},
					},
				},
				Name: "prometheus",
			},
		},
	}},
	Package: "prometheus",
	Path:    "experimental/prometheus",
}
