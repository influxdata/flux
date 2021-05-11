// DO NOT EDIT: This file is autogenerated via the builtin command.

package csv

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
					Column: 64,
					Line:   7,
				},
				File:   "csv.flux",
				Source: "package csv\n\n\nimport c \"csv\"\nimport \"experimental/http\"\n\nfrom = (url) => c.from(csv: string(v: http.get(url: url).body))",
				Start: ast.Position{
					Column: 1,
					Line:   1,
				},
			},
		},
		Body: []ast.Statement{&ast.VariableAssignment{
			BaseNode: ast.BaseNode{
				Comments: nil,
				Errors:   nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 64,
						Line:   7,
					},
					File:   "csv.flux",
					Source: "from = (url) => c.from(csv: string(v: http.get(url: url).body))",
					Start: ast.Position{
						Column: 1,
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
							Column: 5,
							Line:   7,
						},
						File:   "csv.flux",
						Source: "from",
						Start: ast.Position{
							Column: 1,
							Line:   7,
						},
					},
				},
				Name: "from",
			},
			Init: &ast.FunctionExpression{
				Arrow: nil,
				BaseNode: ast.BaseNode{
					Comments: nil,
					Errors:   nil,
					Loc: &ast.SourceLocation{
						End: ast.Position{
							Column: 64,
							Line:   7,
						},
						File:   "csv.flux",
						Source: "(url) => c.from(csv: string(v: http.get(url: url).body))",
						Start: ast.Position{
							Column: 8,
							Line:   7,
						},
					},
				},
				Body: &ast.CallExpression{
					Arguments: []ast.Expression{&ast.ObjectExpression{
						BaseNode: ast.BaseNode{
							Comments: nil,
							Errors:   nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 63,
									Line:   7,
								},
								File:   "csv.flux",
								Source: "csv: string(v: http.get(url: url).body)",
								Start: ast.Position{
									Column: 24,
									Line:   7,
								},
							},
						},
						Lbrace: nil,
						Properties: []*ast.Property{&ast.Property{
							BaseNode: ast.BaseNode{
								Comments: nil,
								Errors:   nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 63,
										Line:   7,
									},
									File:   "csv.flux",
									Source: "csv: string(v: http.get(url: url).body)",
									Start: ast.Position{
										Column: 24,
										Line:   7,
									},
								},
							},
							Comma: nil,
							Key: &ast.Identifier{
								BaseNode: ast.BaseNode{
									Comments: nil,
									Errors:   nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 27,
											Line:   7,
										},
										File:   "csv.flux",
										Source: "csv",
										Start: ast.Position{
											Column: 24,
											Line:   7,
										},
									},
								},
								Name: "csv",
							},
							Separator: nil,
							Value: &ast.CallExpression{
								Arguments: []ast.Expression{&ast.ObjectExpression{
									BaseNode: ast.BaseNode{
										Comments: nil,
										Errors:   nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 62,
												Line:   7,
											},
											File:   "csv.flux",
											Source: "v: http.get(url: url).body",
											Start: ast.Position{
												Column: 36,
												Line:   7,
											},
										},
									},
									Lbrace: nil,
									Properties: []*ast.Property{&ast.Property{
										BaseNode: ast.BaseNode{
											Comments: nil,
											Errors:   nil,
											Loc: &ast.SourceLocation{
												End: ast.Position{
													Column: 62,
													Line:   7,
												},
												File:   "csv.flux",
												Source: "v: http.get(url: url).body",
												Start: ast.Position{
													Column: 36,
													Line:   7,
												},
											},
										},
										Comma: nil,
										Key: &ast.Identifier{
											BaseNode: ast.BaseNode{
												Comments: nil,
												Errors:   nil,
												Loc: &ast.SourceLocation{
													End: ast.Position{
														Column: 37,
														Line:   7,
													},
													File:   "csv.flux",
													Source: "v",
													Start: ast.Position{
														Column: 36,
														Line:   7,
													},
												},
											},
											Name: "v",
										},
										Separator: nil,
										Value: &ast.MemberExpression{
											BaseNode: ast.BaseNode{
												Comments: nil,
												Errors:   nil,
												Loc: &ast.SourceLocation{
													End: ast.Position{
														Column: 62,
														Line:   7,
													},
													File:   "csv.flux",
													Source: "http.get(url: url).body",
													Start: ast.Position{
														Column: 39,
														Line:   7,
													},
												},
											},
											Lbrack: nil,
											Object: &ast.CallExpression{
												Arguments: []ast.Expression{&ast.ObjectExpression{
													BaseNode: ast.BaseNode{
														Comments: nil,
														Errors:   nil,
														Loc: &ast.SourceLocation{
															End: ast.Position{
																Column: 56,
																Line:   7,
															},
															File:   "csv.flux",
															Source: "url: url",
															Start: ast.Position{
																Column: 48,
																Line:   7,
															},
														},
													},
													Lbrace: nil,
													Properties: []*ast.Property{&ast.Property{
														BaseNode: ast.BaseNode{
															Comments: nil,
															Errors:   nil,
															Loc: &ast.SourceLocation{
																End: ast.Position{
																	Column: 56,
																	Line:   7,
																},
																File:   "csv.flux",
																Source: "url: url",
																Start: ast.Position{
																	Column: 48,
																	Line:   7,
																},
															},
														},
														Comma: nil,
														Key: &ast.Identifier{
															BaseNode: ast.BaseNode{
																Comments: nil,
																Errors:   nil,
																Loc: &ast.SourceLocation{
																	End: ast.Position{
																		Column: 51,
																		Line:   7,
																	},
																	File:   "csv.flux",
																	Source: "url",
																	Start: ast.Position{
																		Column: 48,
																		Line:   7,
																	},
																},
															},
															Name: "url",
														},
														Separator: nil,
														Value: &ast.Identifier{
															BaseNode: ast.BaseNode{
																Comments: nil,
																Errors:   nil,
																Loc: &ast.SourceLocation{
																	End: ast.Position{
																		Column: 56,
																		Line:   7,
																	},
																	File:   "csv.flux",
																	Source: "url",
																	Start: ast.Position{
																		Column: 53,
																		Line:   7,
																	},
																},
															},
															Name: "url",
														},
													}},
													Rbrace: nil,
													With:   nil,
												}},
												BaseNode: ast.BaseNode{
													Comments: nil,
													Errors:   nil,
													Loc: &ast.SourceLocation{
														End: ast.Position{
															Column: 57,
															Line:   7,
														},
														File:   "csv.flux",
														Source: "http.get(url: url)",
														Start: ast.Position{
															Column: 39,
															Line:   7,
														},
													},
												},
												Callee: &ast.MemberExpression{
													BaseNode: ast.BaseNode{
														Comments: nil,
														Errors:   nil,
														Loc: &ast.SourceLocation{
															End: ast.Position{
																Column: 47,
																Line:   7,
															},
															File:   "csv.flux",
															Source: "http.get",
															Start: ast.Position{
																Column: 39,
																Line:   7,
															},
														},
													},
													Lbrack: nil,
													Object: &ast.Identifier{
														BaseNode: ast.BaseNode{
															Comments: nil,
															Errors:   nil,
															Loc: &ast.SourceLocation{
																End: ast.Position{
																	Column: 43,
																	Line:   7,
																},
																File:   "csv.flux",
																Source: "http",
																Start: ast.Position{
																	Column: 39,
																	Line:   7,
																},
															},
														},
														Name: "http",
													},
													Property: &ast.Identifier{
														BaseNode: ast.BaseNode{
															Comments: nil,
															Errors:   nil,
															Loc: &ast.SourceLocation{
																End: ast.Position{
																	Column: 47,
																	Line:   7,
																},
																File:   "csv.flux",
																Source: "get",
																Start: ast.Position{
																	Column: 44,
																	Line:   7,
																},
															},
														},
														Name: "get",
													},
													Rbrack: nil,
												},
												Lparen: nil,
												Rparen: nil,
											},
											Property: &ast.Identifier{
												BaseNode: ast.BaseNode{
													Comments: nil,
													Errors:   nil,
													Loc: &ast.SourceLocation{
														End: ast.Position{
															Column: 62,
															Line:   7,
														},
														File:   "csv.flux",
														Source: "body",
														Start: ast.Position{
															Column: 58,
															Line:   7,
														},
													},
												},
												Name: "body",
											},
											Rbrack: nil,
										},
									}},
									Rbrace: nil,
									With:   nil,
								}},
								BaseNode: ast.BaseNode{
									Comments: nil,
									Errors:   nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 63,
											Line:   7,
										},
										File:   "csv.flux",
										Source: "string(v: http.get(url: url).body)",
										Start: ast.Position{
											Column: 29,
											Line:   7,
										},
									},
								},
								Callee: &ast.Identifier{
									BaseNode: ast.BaseNode{
										Comments: nil,
										Errors:   nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 35,
												Line:   7,
											},
											File:   "csv.flux",
											Source: "string",
											Start: ast.Position{
												Column: 29,
												Line:   7,
											},
										},
									},
									Name: "string",
								},
								Lparen: nil,
								Rparen: nil,
							},
						}},
						Rbrace: nil,
						With:   nil,
					}},
					BaseNode: ast.BaseNode{
						Comments: nil,
						Errors:   nil,
						Loc: &ast.SourceLocation{
							End: ast.Position{
								Column: 64,
								Line:   7,
							},
							File:   "csv.flux",
							Source: "c.from(csv: string(v: http.get(url: url).body))",
							Start: ast.Position{
								Column: 17,
								Line:   7,
							},
						},
					},
					Callee: &ast.MemberExpression{
						BaseNode: ast.BaseNode{
							Comments: nil,
							Errors:   nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 23,
									Line:   7,
								},
								File:   "csv.flux",
								Source: "c.from",
								Start: ast.Position{
									Column: 17,
									Line:   7,
								},
							},
						},
						Lbrack: nil,
						Object: &ast.Identifier{
							BaseNode: ast.BaseNode{
								Comments: nil,
								Errors:   nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 18,
										Line:   7,
									},
									File:   "csv.flux",
									Source: "c",
									Start: ast.Position{
										Column: 17,
										Line:   7,
									},
								},
							},
							Name: "c",
						},
						Property: &ast.Identifier{
							BaseNode: ast.BaseNode{
								Comments: nil,
								Errors:   nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 23,
										Line:   7,
									},
									File:   "csv.flux",
									Source: "from",
									Start: ast.Position{
										Column: 19,
										Line:   7,
									},
								},
							},
							Name: "from",
						},
						Rbrack: nil,
					},
					Lparen: nil,
					Rparen: nil,
				},
				Lparen: nil,
				Params: []*ast.Property{&ast.Property{
					BaseNode: ast.BaseNode{
						Comments: nil,
						Errors:   nil,
						Loc: &ast.SourceLocation{
							End: ast.Position{
								Column: 12,
								Line:   7,
							},
							File:   "csv.flux",
							Source: "url",
							Start: ast.Position{
								Column: 9,
								Line:   7,
							},
						},
					},
					Comma: nil,
					Key: &ast.Identifier{
						BaseNode: ast.BaseNode{
							Comments: nil,
							Errors:   nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 12,
									Line:   7,
								},
								File:   "csv.flux",
								Source: "url",
								Start: ast.Position{
									Column: 9,
									Line:   7,
								},
							},
						},
						Name: "url",
					},
					Separator: nil,
					Value:     nil,
				}},
				Rparan: nil,
			},
		}},
		Eof: nil,
		Imports: []*ast.ImportDeclaration{&ast.ImportDeclaration{
			As: &ast.Identifier{
				BaseNode: ast.BaseNode{
					Comments: nil,
					Errors:   nil,
					Loc: &ast.SourceLocation{
						End: ast.Position{
							Column: 9,
							Line:   4,
						},
						File:   "csv.flux",
						Source: "c",
						Start: ast.Position{
							Column: 8,
							Line:   4,
						},
					},
				},
				Name: "c",
			},
			BaseNode: ast.BaseNode{
				Comments: nil,
				Errors:   nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 15,
						Line:   4,
					},
					File:   "csv.flux",
					Source: "import c \"csv\"",
					Start: ast.Position{
						Column: 1,
						Line:   4,
					},
				},
			},
			Path: &ast.StringLiteral{
				BaseNode: ast.BaseNode{
					Comments: nil,
					Errors:   nil,
					Loc: &ast.SourceLocation{
						End: ast.Position{
							Column: 15,
							Line:   4,
						},
						File:   "csv.flux",
						Source: "\"csv\"",
						Start: ast.Position{
							Column: 10,
							Line:   4,
						},
					},
				},
				Value: "csv",
			},
		}, &ast.ImportDeclaration{
			As: nil,
			BaseNode: ast.BaseNode{
				Comments: nil,
				Errors:   nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 27,
						Line:   5,
					},
					File:   "csv.flux",
					Source: "import \"experimental/http\"",
					Start: ast.Position{
						Column: 1,
						Line:   5,
					},
				},
			},
			Path: &ast.StringLiteral{
				BaseNode: ast.BaseNode{
					Comments: nil,
					Errors:   nil,
					Loc: &ast.SourceLocation{
						End: ast.Position{
							Column: 27,
							Line:   5,
						},
						File:   "csv.flux",
						Source: "\"experimental/http\"",
						Start: ast.Position{
							Column: 8,
							Line:   5,
						},
					},
				},
				Value: "experimental/http",
			},
		}},
		Metadata: "parser-type=rust",
		Name:     "csv.flux",
		Package: &ast.PackageClause{
			BaseNode: ast.BaseNode{
				Comments: nil,
				Errors:   nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 12,
						Line:   1,
					},
					File:   "csv.flux",
					Source: "package csv",
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
							Column: 12,
							Line:   1,
						},
						File:   "csv.flux",
						Source: "csv",
						Start: ast.Position{
							Column: 9,
							Line:   1,
						},
					},
				},
				Name: "csv",
			},
		},
	}},
	Package: "csv",
	Path:    "experimental/csv",
}
