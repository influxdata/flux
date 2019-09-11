// DO NOT EDIT: This file is autogenerated via the builtin command.

package secrets

import (
	ast "github.com/influxdata/flux/ast"
	parser "github.com/influxdata/flux/internal/parser"
)

var FluxTestPackages = []*ast.Package{&ast.Package{
	BaseNode: ast.BaseNode{
		Errors: nil,
		Loc:    nil,
	},
	Files: []*ast.File{&ast.File{
		BaseNode: ast.BaseNode{
			Errors: nil,
			Loc: &ast.SourceLocation{
				End: ast.Position{
					Column: 100,
					Line:   30,
				},
				File:   "secrets_test.flux",
				Source: "package secrets_test\n\nimport \"testing\"\nimport \"influxdata/influxdb/secrets\"\n\noption now = () => (2030-01-01T00:00:00Z)\n\ninData = \"\n#datatype,string,long,dateTime:RFC3339,double,string,string\n#group,false,false,false,false,true,true\n#default,_result,,,,,\n,result,table,_time,_value,_field,_measurement\n,,0,2018-05-22T19:53:26Z,1.83,load1,system\n\"\n\noutData = \"\n#datatype,string,long,dateTime:RFC3339,double,string,string,string\n#group,false,false,false,false,true,true,false\n#default,_result,,,,,,\n,result,table,_time,_value,_field,_measurement,token\n,,0,2018-05-22T19:53:26Z,1.83,load1,system,mysecrettoken\n\"\n\ntoken = secrets.get(key: \"token\")\nt_get_secret = (table=<-) =>\n\ttable\n    |> set(key: \"token\", value: token)\n\ntest _get_secret = () =>\n\t({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_get_secret})",
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
							Column: 42,
							Line:   6,
						},
						File:   "secrets_test.flux",
						Source: "now = () => (2030-01-01T00:00:00Z)",
						Start: ast.Position{
							Column: 8,
							Line:   6,
						},
					},
				},
				ID: &ast.Identifier{
					BaseNode: ast.BaseNode{
						Errors: nil,
						Loc: &ast.SourceLocation{
							End: ast.Position{
								Column: 11,
								Line:   6,
							},
							File:   "secrets_test.flux",
							Source: "now",
							Start: ast.Position{
								Column: 8,
								Line:   6,
							},
						},
					},
					Name: "now",
				},
				Init: &ast.FunctionExpression{
					BaseNode: ast.BaseNode{
						Errors: nil,
						Loc: &ast.SourceLocation{
							End: ast.Position{
								Column: 42,
								Line:   6,
							},
							File:   "secrets_test.flux",
							Source: "() => (2030-01-01T00:00:00Z)",
							Start: ast.Position{
								Column: 14,
								Line:   6,
							},
						},
					},
					Body: &ast.ParenExpression{
						BaseNode: ast.BaseNode{
							Errors: nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 42,
									Line:   6,
								},
								File:   "secrets_test.flux",
								Source: "(2030-01-01T00:00:00Z)",
								Start: ast.Position{
									Column: 20,
									Line:   6,
								},
							},
						},
						Expression: &ast.DateTimeLiteral{
							BaseNode: ast.BaseNode{
								Errors: nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 41,
										Line:   6,
									},
									File:   "secrets_test.flux",
									Source: "2030-01-01T00:00:00Z",
									Start: ast.Position{
										Column: 21,
										Line:   6,
									},
								},
							},
							Value: parser.MustParseTime("2030-01-01T00:00:00Z"),
						},
					},
					Params: nil,
				},
			},
			BaseNode: ast.BaseNode{
				Errors: nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 42,
						Line:   6,
					},
					File:   "secrets_test.flux",
					Source: "option now = () => (2030-01-01T00:00:00Z)",
					Start: ast.Position{
						Column: 1,
						Line:   6,
					},
				},
			},
		}, &ast.VariableAssignment{
			BaseNode: ast.BaseNode{
				Errors: nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 2,
						Line:   14,
					},
					File:   "secrets_test.flux",
					Source: "inData = \"\n#datatype,string,long,dateTime:RFC3339,double,string,string\n#group,false,false,false,false,true,true\n#default,_result,,,,,\n,result,table,_time,_value,_field,_measurement\n,,0,2018-05-22T19:53:26Z,1.83,load1,system\n\"",
					Start: ast.Position{
						Column: 1,
						Line:   8,
					},
				},
			},
			ID: &ast.Identifier{
				BaseNode: ast.BaseNode{
					Errors: nil,
					Loc: &ast.SourceLocation{
						End: ast.Position{
							Column: 7,
							Line:   8,
						},
						File:   "secrets_test.flux",
						Source: "inData",
						Start: ast.Position{
							Column: 1,
							Line:   8,
						},
					},
				},
				Name: "inData",
			},
			Init: &ast.StringLiteral{
				BaseNode: ast.BaseNode{
					Errors: nil,
					Loc: &ast.SourceLocation{
						End: ast.Position{
							Column: 2,
							Line:   14,
						},
						File:   "secrets_test.flux",
						Source: "\"\n#datatype,string,long,dateTime:RFC3339,double,string,string\n#group,false,false,false,false,true,true\n#default,_result,,,,,\n,result,table,_time,_value,_field,_measurement\n,,0,2018-05-22T19:53:26Z,1.83,load1,system\n\"",
						Start: ast.Position{
							Column: 10,
							Line:   8,
						},
					},
				},
				Value: "\n#datatype,string,long,dateTime:RFC3339,double,string,string\n#group,false,false,false,false,true,true\n#default,_result,,,,,\n,result,table,_time,_value,_field,_measurement\n,,0,2018-05-22T19:53:26Z,1.83,load1,system\n",
			},
		}, &ast.VariableAssignment{
			BaseNode: ast.BaseNode{
				Errors: nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 2,
						Line:   22,
					},
					File:   "secrets_test.flux",
					Source: "outData = \"\n#datatype,string,long,dateTime:RFC3339,double,string,string,string\n#group,false,false,false,false,true,true,false\n#default,_result,,,,,,\n,result,table,_time,_value,_field,_measurement,token\n,,0,2018-05-22T19:53:26Z,1.83,load1,system,mysecrettoken\n\"",
					Start: ast.Position{
						Column: 1,
						Line:   16,
					},
				},
			},
			ID: &ast.Identifier{
				BaseNode: ast.BaseNode{
					Errors: nil,
					Loc: &ast.SourceLocation{
						End: ast.Position{
							Column: 8,
							Line:   16,
						},
						File:   "secrets_test.flux",
						Source: "outData",
						Start: ast.Position{
							Column: 1,
							Line:   16,
						},
					},
				},
				Name: "outData",
			},
			Init: &ast.StringLiteral{
				BaseNode: ast.BaseNode{
					Errors: nil,
					Loc: &ast.SourceLocation{
						End: ast.Position{
							Column: 2,
							Line:   22,
						},
						File:   "secrets_test.flux",
						Source: "\"\n#datatype,string,long,dateTime:RFC3339,double,string,string,string\n#group,false,false,false,false,true,true,false\n#default,_result,,,,,,\n,result,table,_time,_value,_field,_measurement,token\n,,0,2018-05-22T19:53:26Z,1.83,load1,system,mysecrettoken\n\"",
						Start: ast.Position{
							Column: 11,
							Line:   16,
						},
					},
				},
				Value: "\n#datatype,string,long,dateTime:RFC3339,double,string,string,string\n#group,false,false,false,false,true,true,false\n#default,_result,,,,,,\n,result,table,_time,_value,_field,_measurement,token\n,,0,2018-05-22T19:53:26Z,1.83,load1,system,mysecrettoken\n",
			},
		}, &ast.VariableAssignment{
			BaseNode: ast.BaseNode{
				Errors: nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 34,
						Line:   24,
					},
					File:   "secrets_test.flux",
					Source: "token = secrets.get(key: \"token\")",
					Start: ast.Position{
						Column: 1,
						Line:   24,
					},
				},
			},
			ID: &ast.Identifier{
				BaseNode: ast.BaseNode{
					Errors: nil,
					Loc: &ast.SourceLocation{
						End: ast.Position{
							Column: 6,
							Line:   24,
						},
						File:   "secrets_test.flux",
						Source: "token",
						Start: ast.Position{
							Column: 1,
							Line:   24,
						},
					},
				},
				Name: "token",
			},
			Init: &ast.CallExpression{
				Arguments: []ast.Expression{&ast.ObjectExpression{
					BaseNode: ast.BaseNode{
						Errors: nil,
						Loc: &ast.SourceLocation{
							End: ast.Position{
								Column: 33,
								Line:   24,
							},
							File:   "secrets_test.flux",
							Source: "key: \"token\"",
							Start: ast.Position{
								Column: 21,
								Line:   24,
							},
						},
					},
					Properties: []*ast.Property{&ast.Property{
						BaseNode: ast.BaseNode{
							Errors: nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 33,
									Line:   24,
								},
								File:   "secrets_test.flux",
								Source: "key: \"token\"",
								Start: ast.Position{
									Column: 21,
									Line:   24,
								},
							},
						},
						Key: &ast.Identifier{
							BaseNode: ast.BaseNode{
								Errors: nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 24,
										Line:   24,
									},
									File:   "secrets_test.flux",
									Source: "key",
									Start: ast.Position{
										Column: 21,
										Line:   24,
									},
								},
							},
							Name: "key",
						},
						Value: &ast.StringLiteral{
							BaseNode: ast.BaseNode{
								Errors: nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 33,
										Line:   24,
									},
									File:   "secrets_test.flux",
									Source: "\"token\"",
									Start: ast.Position{
										Column: 26,
										Line:   24,
									},
								},
							},
							Value: "token",
						},
					}},
					With: nil,
				}},
				BaseNode: ast.BaseNode{
					Errors: nil,
					Loc: &ast.SourceLocation{
						End: ast.Position{
							Column: 34,
							Line:   24,
						},
						File:   "secrets_test.flux",
						Source: "secrets.get(key: \"token\")",
						Start: ast.Position{
							Column: 9,
							Line:   24,
						},
					},
				},
				Callee: &ast.MemberExpression{
					BaseNode: ast.BaseNode{
						Errors: nil,
						Loc: &ast.SourceLocation{
							End: ast.Position{
								Column: 20,
								Line:   24,
							},
							File:   "secrets_test.flux",
							Source: "secrets.get",
							Start: ast.Position{
								Column: 9,
								Line:   24,
							},
						},
					},
					Object: &ast.Identifier{
						BaseNode: ast.BaseNode{
							Errors: nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 16,
									Line:   24,
								},
								File:   "secrets_test.flux",
								Source: "secrets",
								Start: ast.Position{
									Column: 9,
									Line:   24,
								},
							},
						},
						Name: "secrets",
					},
					Property: &ast.Identifier{
						BaseNode: ast.BaseNode{
							Errors: nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 20,
									Line:   24,
								},
								File:   "secrets_test.flux",
								Source: "get",
								Start: ast.Position{
									Column: 17,
									Line:   24,
								},
							},
						},
						Name: "get",
					},
				},
			},
		}, &ast.VariableAssignment{
			BaseNode: ast.BaseNode{
				Errors: nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 39,
						Line:   27,
					},
					File:   "secrets_test.flux",
					Source: "t_get_secret = (table=<-) =>\n\ttable\n    |> set(key: \"token\", value: token)",
					Start: ast.Position{
						Column: 1,
						Line:   25,
					},
				},
			},
			ID: &ast.Identifier{
				BaseNode: ast.BaseNode{
					Errors: nil,
					Loc: &ast.SourceLocation{
						End: ast.Position{
							Column: 13,
							Line:   25,
						},
						File:   "secrets_test.flux",
						Source: "t_get_secret",
						Start: ast.Position{
							Column: 1,
							Line:   25,
						},
					},
				},
				Name: "t_get_secret",
			},
			Init: &ast.FunctionExpression{
				BaseNode: ast.BaseNode{
					Errors: nil,
					Loc: &ast.SourceLocation{
						End: ast.Position{
							Column: 39,
							Line:   27,
						},
						File:   "secrets_test.flux",
						Source: "(table=<-) =>\n\ttable\n    |> set(key: \"token\", value: token)",
						Start: ast.Position{
							Column: 16,
							Line:   25,
						},
					},
				},
				Body: &ast.PipeExpression{
					Argument: &ast.Identifier{
						BaseNode: ast.BaseNode{
							Errors: nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 7,
									Line:   26,
								},
								File:   "secrets_test.flux",
								Source: "table",
								Start: ast.Position{
									Column: 2,
									Line:   26,
								},
							},
						},
						Name: "table",
					},
					BaseNode: ast.BaseNode{
						Errors: nil,
						Loc: &ast.SourceLocation{
							End: ast.Position{
								Column: 39,
								Line:   27,
							},
							File:   "secrets_test.flux",
							Source: "table\n    |> set(key: \"token\", value: token)",
							Start: ast.Position{
								Column: 2,
								Line:   26,
							},
						},
					},
					Call: &ast.CallExpression{
						Arguments: []ast.Expression{&ast.ObjectExpression{
							BaseNode: ast.BaseNode{
								Errors: nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 38,
										Line:   27,
									},
									File:   "secrets_test.flux",
									Source: "key: \"token\", value: token",
									Start: ast.Position{
										Column: 12,
										Line:   27,
									},
								},
							},
							Properties: []*ast.Property{&ast.Property{
								BaseNode: ast.BaseNode{
									Errors: nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 24,
											Line:   27,
										},
										File:   "secrets_test.flux",
										Source: "key: \"token\"",
										Start: ast.Position{
											Column: 12,
											Line:   27,
										},
									},
								},
								Key: &ast.Identifier{
									BaseNode: ast.BaseNode{
										Errors: nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 15,
												Line:   27,
											},
											File:   "secrets_test.flux",
											Source: "key",
											Start: ast.Position{
												Column: 12,
												Line:   27,
											},
										},
									},
									Name: "key",
								},
								Value: &ast.StringLiteral{
									BaseNode: ast.BaseNode{
										Errors: nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 24,
												Line:   27,
											},
											File:   "secrets_test.flux",
											Source: "\"token\"",
											Start: ast.Position{
												Column: 17,
												Line:   27,
											},
										},
									},
									Value: "token",
								},
							}, &ast.Property{
								BaseNode: ast.BaseNode{
									Errors: nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 38,
											Line:   27,
										},
										File:   "secrets_test.flux",
										Source: "value: token",
										Start: ast.Position{
											Column: 26,
											Line:   27,
										},
									},
								},
								Key: &ast.Identifier{
									BaseNode: ast.BaseNode{
										Errors: nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 31,
												Line:   27,
											},
											File:   "secrets_test.flux",
											Source: "value",
											Start: ast.Position{
												Column: 26,
												Line:   27,
											},
										},
									},
									Name: "value",
								},
								Value: &ast.Identifier{
									BaseNode: ast.BaseNode{
										Errors: nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 38,
												Line:   27,
											},
											File:   "secrets_test.flux",
											Source: "token",
											Start: ast.Position{
												Column: 33,
												Line:   27,
											},
										},
									},
									Name: "token",
								},
							}},
							With: nil,
						}},
						BaseNode: ast.BaseNode{
							Errors: nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 39,
									Line:   27,
								},
								File:   "secrets_test.flux",
								Source: "set(key: \"token\", value: token)",
								Start: ast.Position{
									Column: 8,
									Line:   27,
								},
							},
						},
						Callee: &ast.Identifier{
							BaseNode: ast.BaseNode{
								Errors: nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 11,
										Line:   27,
									},
									File:   "secrets_test.flux",
									Source: "set",
									Start: ast.Position{
										Column: 8,
										Line:   27,
									},
								},
							},
							Name: "set",
						},
					},
				},
				Params: []*ast.Property{&ast.Property{
					BaseNode: ast.BaseNode{
						Errors: nil,
						Loc: &ast.SourceLocation{
							End: ast.Position{
								Column: 25,
								Line:   25,
							},
							File:   "secrets_test.flux",
							Source: "table=<-",
							Start: ast.Position{
								Column: 17,
								Line:   25,
							},
						},
					},
					Key: &ast.Identifier{
						BaseNode: ast.BaseNode{
							Errors: nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 22,
									Line:   25,
								},
								File:   "secrets_test.flux",
								Source: "table",
								Start: ast.Position{
									Column: 17,
									Line:   25,
								},
							},
						},
						Name: "table",
					},
					Value: &ast.PipeLiteral{BaseNode: ast.BaseNode{
						Errors: nil,
						Loc: &ast.SourceLocation{
							End: ast.Position{
								Column: 25,
								Line:   25,
							},
							File:   "secrets_test.flux",
							Source: "<-",
							Start: ast.Position{
								Column: 23,
								Line:   25,
							},
						},
					}},
				}},
			},
		}, &ast.TestStatement{
			Assignment: &ast.VariableAssignment{
				BaseNode: ast.BaseNode{
					Errors: nil,
					Loc: &ast.SourceLocation{
						End: ast.Position{
							Column: 100,
							Line:   30,
						},
						File:   "secrets_test.flux",
						Source: "_get_secret = () =>\n\t({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_get_secret})",
						Start: ast.Position{
							Column: 6,
							Line:   29,
						},
					},
				},
				ID: &ast.Identifier{
					BaseNode: ast.BaseNode{
						Errors: nil,
						Loc: &ast.SourceLocation{
							End: ast.Position{
								Column: 17,
								Line:   29,
							},
							File:   "secrets_test.flux",
							Source: "_get_secret",
							Start: ast.Position{
								Column: 6,
								Line:   29,
							},
						},
					},
					Name: "_get_secret",
				},
				Init: &ast.FunctionExpression{
					BaseNode: ast.BaseNode{
						Errors: nil,
						Loc: &ast.SourceLocation{
							End: ast.Position{
								Column: 100,
								Line:   30,
							},
							File:   "secrets_test.flux",
							Source: "() =>\n\t({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_get_secret})",
							Start: ast.Position{
								Column: 20,
								Line:   29,
							},
						},
					},
					Body: &ast.ParenExpression{
						BaseNode: ast.BaseNode{
							Errors: nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 100,
									Line:   30,
								},
								File:   "secrets_test.flux",
								Source: "({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_get_secret})",
								Start: ast.Position{
									Column: 2,
									Line:   30,
								},
							},
						},
						Expression: &ast.ObjectExpression{
							BaseNode: ast.BaseNode{
								Errors: nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 99,
										Line:   30,
									},
									File:   "secrets_test.flux",
									Source: "{input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_get_secret}",
									Start: ast.Position{
										Column: 3,
										Line:   30,
									},
								},
							},
							Properties: []*ast.Property{&ast.Property{
								BaseNode: ast.BaseNode{
									Errors: nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 43,
											Line:   30,
										},
										File:   "secrets_test.flux",
										Source: "input: testing.loadStorage(csv: inData)",
										Start: ast.Position{
											Column: 4,
											Line:   30,
										},
									},
								},
								Key: &ast.Identifier{
									BaseNode: ast.BaseNode{
										Errors: nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 9,
												Line:   30,
											},
											File:   "secrets_test.flux",
											Source: "input",
											Start: ast.Position{
												Column: 4,
												Line:   30,
											},
										},
									},
									Name: "input",
								},
								Value: &ast.CallExpression{
									Arguments: []ast.Expression{&ast.ObjectExpression{
										BaseNode: ast.BaseNode{
											Errors: nil,
											Loc: &ast.SourceLocation{
												End: ast.Position{
													Column: 42,
													Line:   30,
												},
												File:   "secrets_test.flux",
												Source: "csv: inData",
												Start: ast.Position{
													Column: 31,
													Line:   30,
												},
											},
										},
										Properties: []*ast.Property{&ast.Property{
											BaseNode: ast.BaseNode{
												Errors: nil,
												Loc: &ast.SourceLocation{
													End: ast.Position{
														Column: 42,
														Line:   30,
													},
													File:   "secrets_test.flux",
													Source: "csv: inData",
													Start: ast.Position{
														Column: 31,
														Line:   30,
													},
												},
											},
											Key: &ast.Identifier{
												BaseNode: ast.BaseNode{
													Errors: nil,
													Loc: &ast.SourceLocation{
														End: ast.Position{
															Column: 34,
															Line:   30,
														},
														File:   "secrets_test.flux",
														Source: "csv",
														Start: ast.Position{
															Column: 31,
															Line:   30,
														},
													},
												},
												Name: "csv",
											},
											Value: &ast.Identifier{
												BaseNode: ast.BaseNode{
													Errors: nil,
													Loc: &ast.SourceLocation{
														End: ast.Position{
															Column: 42,
															Line:   30,
														},
														File:   "secrets_test.flux",
														Source: "inData",
														Start: ast.Position{
															Column: 36,
															Line:   30,
														},
													},
												},
												Name: "inData",
											},
										}},
										With: nil,
									}},
									BaseNode: ast.BaseNode{
										Errors: nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 43,
												Line:   30,
											},
											File:   "secrets_test.flux",
											Source: "testing.loadStorage(csv: inData)",
											Start: ast.Position{
												Column: 11,
												Line:   30,
											},
										},
									},
									Callee: &ast.MemberExpression{
										BaseNode: ast.BaseNode{
											Errors: nil,
											Loc: &ast.SourceLocation{
												End: ast.Position{
													Column: 30,
													Line:   30,
												},
												File:   "secrets_test.flux",
												Source: "testing.loadStorage",
												Start: ast.Position{
													Column: 11,
													Line:   30,
												},
											},
										},
										Object: &ast.Identifier{
											BaseNode: ast.BaseNode{
												Errors: nil,
												Loc: &ast.SourceLocation{
													End: ast.Position{
														Column: 18,
														Line:   30,
													},
													File:   "secrets_test.flux",
													Source: "testing",
													Start: ast.Position{
														Column: 11,
														Line:   30,
													},
												},
											},
											Name: "testing",
										},
										Property: &ast.Identifier{
											BaseNode: ast.BaseNode{
												Errors: nil,
												Loc: &ast.SourceLocation{
													End: ast.Position{
														Column: 30,
														Line:   30,
													},
													File:   "secrets_test.flux",
													Source: "loadStorage",
													Start: ast.Position{
														Column: 19,
														Line:   30,
													},
												},
											},
											Name: "loadStorage",
										},
									},
								},
							}, &ast.Property{
								BaseNode: ast.BaseNode{
									Errors: nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 80,
											Line:   30,
										},
										File:   "secrets_test.flux",
										Source: "want: testing.loadMem(csv: outData)",
										Start: ast.Position{
											Column: 45,
											Line:   30,
										},
									},
								},
								Key: &ast.Identifier{
									BaseNode: ast.BaseNode{
										Errors: nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 49,
												Line:   30,
											},
											File:   "secrets_test.flux",
											Source: "want",
											Start: ast.Position{
												Column: 45,
												Line:   30,
											},
										},
									},
									Name: "want",
								},
								Value: &ast.CallExpression{
									Arguments: []ast.Expression{&ast.ObjectExpression{
										BaseNode: ast.BaseNode{
											Errors: nil,
											Loc: &ast.SourceLocation{
												End: ast.Position{
													Column: 79,
													Line:   30,
												},
												File:   "secrets_test.flux",
												Source: "csv: outData",
												Start: ast.Position{
													Column: 67,
													Line:   30,
												},
											},
										},
										Properties: []*ast.Property{&ast.Property{
											BaseNode: ast.BaseNode{
												Errors: nil,
												Loc: &ast.SourceLocation{
													End: ast.Position{
														Column: 79,
														Line:   30,
													},
													File:   "secrets_test.flux",
													Source: "csv: outData",
													Start: ast.Position{
														Column: 67,
														Line:   30,
													},
												},
											},
											Key: &ast.Identifier{
												BaseNode: ast.BaseNode{
													Errors: nil,
													Loc: &ast.SourceLocation{
														End: ast.Position{
															Column: 70,
															Line:   30,
														},
														File:   "secrets_test.flux",
														Source: "csv",
														Start: ast.Position{
															Column: 67,
															Line:   30,
														},
													},
												},
												Name: "csv",
											},
											Value: &ast.Identifier{
												BaseNode: ast.BaseNode{
													Errors: nil,
													Loc: &ast.SourceLocation{
														End: ast.Position{
															Column: 79,
															Line:   30,
														},
														File:   "secrets_test.flux",
														Source: "outData",
														Start: ast.Position{
															Column: 72,
															Line:   30,
														},
													},
												},
												Name: "outData",
											},
										}},
										With: nil,
									}},
									BaseNode: ast.BaseNode{
										Errors: nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 80,
												Line:   30,
											},
											File:   "secrets_test.flux",
											Source: "testing.loadMem(csv: outData)",
											Start: ast.Position{
												Column: 51,
												Line:   30,
											},
										},
									},
									Callee: &ast.MemberExpression{
										BaseNode: ast.BaseNode{
											Errors: nil,
											Loc: &ast.SourceLocation{
												End: ast.Position{
													Column: 66,
													Line:   30,
												},
												File:   "secrets_test.flux",
												Source: "testing.loadMem",
												Start: ast.Position{
													Column: 51,
													Line:   30,
												},
											},
										},
										Object: &ast.Identifier{
											BaseNode: ast.BaseNode{
												Errors: nil,
												Loc: &ast.SourceLocation{
													End: ast.Position{
														Column: 58,
														Line:   30,
													},
													File:   "secrets_test.flux",
													Source: "testing",
													Start: ast.Position{
														Column: 51,
														Line:   30,
													},
												},
											},
											Name: "testing",
										},
										Property: &ast.Identifier{
											BaseNode: ast.BaseNode{
												Errors: nil,
												Loc: &ast.SourceLocation{
													End: ast.Position{
														Column: 66,
														Line:   30,
													},
													File:   "secrets_test.flux",
													Source: "loadMem",
													Start: ast.Position{
														Column: 59,
														Line:   30,
													},
												},
											},
											Name: "loadMem",
										},
									},
								},
							}, &ast.Property{
								BaseNode: ast.BaseNode{
									Errors: nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 98,
											Line:   30,
										},
										File:   "secrets_test.flux",
										Source: "fn: t_get_secret",
										Start: ast.Position{
											Column: 82,
											Line:   30,
										},
									},
								},
								Key: &ast.Identifier{
									BaseNode: ast.BaseNode{
										Errors: nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 84,
												Line:   30,
											},
											File:   "secrets_test.flux",
											Source: "fn",
											Start: ast.Position{
												Column: 82,
												Line:   30,
											},
										},
									},
									Name: "fn",
								},
								Value: &ast.Identifier{
									BaseNode: ast.BaseNode{
										Errors: nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 98,
												Line:   30,
											},
											File:   "secrets_test.flux",
											Source: "t_get_secret",
											Start: ast.Position{
												Column: 86,
												Line:   30,
											},
										},
									},
									Name: "t_get_secret",
								},
							}},
							With: nil,
						},
					},
					Params: nil,
				},
			},
			BaseNode: ast.BaseNode{
				Errors: nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 100,
						Line:   30,
					},
					File:   "secrets_test.flux",
					Source: "test _get_secret = () =>\n\t({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_get_secret})",
					Start: ast.Position{
						Column: 1,
						Line:   29,
					},
				},
			},
		}},
		Imports: []*ast.ImportDeclaration{&ast.ImportDeclaration{
			As: nil,
			BaseNode: ast.BaseNode{
				Errors: nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 17,
						Line:   3,
					},
					File:   "secrets_test.flux",
					Source: "import \"testing\"",
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
							Column: 17,
							Line:   3,
						},
						File:   "secrets_test.flux",
						Source: "\"testing\"",
						Start: ast.Position{
							Column: 8,
							Line:   3,
						},
					},
				},
				Value: "testing",
			},
		}, &ast.ImportDeclaration{
			As: nil,
			BaseNode: ast.BaseNode{
				Errors: nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 37,
						Line:   4,
					},
					File:   "secrets_test.flux",
					Source: "import \"influxdata/influxdb/secrets\"",
					Start: ast.Position{
						Column: 1,
						Line:   4,
					},
				},
			},
			Path: &ast.StringLiteral{
				BaseNode: ast.BaseNode{
					Errors: nil,
					Loc: &ast.SourceLocation{
						End: ast.Position{
							Column: 37,
							Line:   4,
						},
						File:   "secrets_test.flux",
						Source: "\"influxdata/influxdb/secrets\"",
						Start: ast.Position{
							Column: 8,
							Line:   4,
						},
					},
				},
				Value: "influxdata/influxdb/secrets",
			},
		}},
		Name: "secrets_test.flux",
		Package: &ast.PackageClause{
			BaseNode: ast.BaseNode{
				Errors: nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 21,
						Line:   1,
					},
					File:   "secrets_test.flux",
					Source: "package secrets_test",
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
						File:   "secrets_test.flux",
						Source: "secrets_test",
						Start: ast.Position{
							Column: 9,
							Line:   1,
						},
					},
				},
				Name: "main",
			},
		},
	}},
	Package: "main",
	Path:    "",
}}
