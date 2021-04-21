// DO NOT EDIT: This file is autogenerated via the builtin command.

package influxdb

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
					Line:   5,
				},
				File:   "influxdb.flux",
				Source: "package influxdb\n\n// api submits an HTTP request to the specified API path.\n// Returns HTTP status code, response headers, and body as a byte array.\nbuiltin api",
				Start: ast.Position{
					Column: 1,
					Line:   1,
				},
			},
		},
		Body: []ast.Statement{&ast.BuiltinStatement{
			BaseNode: ast.BaseNode{
				Comments: []ast.Comment{ast.Comment{Text: "// api submits an HTTP request to the specified API path.\n"}, ast.Comment{Text: "// Returns HTTP status code, response headers, and body as a byte array.\n"}},
				Errors:   nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 12,
						Line:   5,
					},
					File:   "influxdb.flux",
					Source: "builtin api",
					Start: ast.Position{
						Column: 1,
						Line:   5,
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
							Line:   5,
						},
						File:   "influxdb.flux",
						Source: "api",
						Start: ast.Position{
							Column: 9,
							Line:   5,
						},
					},
				},
				Name: "api",
			},
			Ty: ast.TypeExpression{
				BaseNode: ast.BaseNode{
					Comments: nil,
					Errors:   nil,
					Loc: &ast.SourceLocation{
						End: ast.Position{
							Column: 2,
							Line:   18,
						},
						File:   "influxdb.flux",
						Source: "(\n\tmethod: string,\n\tpath: string,\n\t?host: string,\n\t?token: string,\n\t?body: bytes,\n\t?headers: [string: string],\n\t?query: [string: string],\n\t?timeout: duration\n) => {\n\tstatusCode: int,\n\tbody: bytes,\n\theaders: [string: string],\n}",
						Start: ast.Position{
							Column: 15,
							Line:   5,
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
								Column: 2,
								Line:   18,
							},
							File:   "influxdb.flux",
							Source: "(\n\tmethod: string,\n\tpath: string,\n\t?host: string,\n\t?token: string,\n\t?body: bytes,\n\t?headers: [string: string],\n\t?query: [string: string],\n\t?timeout: duration\n) => {\n\tstatusCode: int,\n\tbody: bytes,\n\theaders: [string: string],\n}",
							Start: ast.Position{
								Column: 15,
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
									Column: 16,
									Line:   6,
								},
								File:   "influxdb.flux",
								Source: "method: string",
								Start: ast.Position{
									Column: 2,
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
										Column: 8,
										Line:   6,
									},
									File:   "influxdb.flux",
									Source: "method",
									Start: ast.Position{
										Column: 2,
										Line:   6,
									},
								},
							},
							Name: "method",
						},
						Ty: &ast.NamedType{
							BaseNode: ast.BaseNode{
								Comments: nil,
								Errors:   nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 16,
										Line:   6,
									},
									File:   "influxdb.flux",
									Source: "string",
									Start: ast.Position{
										Column: 10,
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
											Column: 16,
											Line:   6,
										},
										File:   "influxdb.flux",
										Source: "string",
										Start: ast.Position{
											Column: 10,
											Line:   6,
										},
									},
								},
								Name: "string",
							},
						},
					}, &ast.ParameterType{
						BaseNode: ast.BaseNode{
							Comments: nil,
							Errors:   nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 14,
									Line:   7,
								},
								File:   "influxdb.flux",
								Source: "path: string",
								Start: ast.Position{
									Column: 2,
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
										Column: 6,
										Line:   7,
									},
									File:   "influxdb.flux",
									Source: "path",
									Start: ast.Position{
										Column: 2,
										Line:   7,
									},
								},
							},
							Name: "path",
						},
						Ty: &ast.NamedType{
							BaseNode: ast.BaseNode{
								Comments: nil,
								Errors:   nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 14,
										Line:   7,
									},
									File:   "influxdb.flux",
									Source: "string",
									Start: ast.Position{
										Column: 8,
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
											Column: 14,
											Line:   7,
										},
										File:   "influxdb.flux",
										Source: "string",
										Start: ast.Position{
											Column: 8,
											Line:   7,
										},
									},
								},
								Name: "string",
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
								File:   "influxdb.flux",
								Source: "?host: string",
								Start: ast.Position{
									Column: 2,
									Line:   8,
								},
							},
						},
						Kind: "Optional",
						Name: &ast.Identifier{
							BaseNode: ast.BaseNode{
								Comments: nil,
								Errors:   nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 7,
										Line:   8,
									},
									File:   "influxdb.flux",
									Source: "host",
									Start: ast.Position{
										Column: 3,
										Line:   8,
									},
								},
							},
							Name: "host",
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
									File:   "influxdb.flux",
									Source: "string",
									Start: ast.Position{
										Column: 9,
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
										File:   "influxdb.flux",
										Source: "string",
										Start: ast.Position{
											Column: 9,
											Line:   8,
										},
									},
								},
								Name: "string",
							},
						},
					}, &ast.ParameterType{
						BaseNode: ast.BaseNode{
							Comments: nil,
							Errors:   nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 16,
									Line:   9,
								},
								File:   "influxdb.flux",
								Source: "?token: string",
								Start: ast.Position{
									Column: 2,
									Line:   9,
								},
							},
						},
						Kind: "Optional",
						Name: &ast.Identifier{
							BaseNode: ast.BaseNode{
								Comments: nil,
								Errors:   nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 8,
										Line:   9,
									},
									File:   "influxdb.flux",
									Source: "token",
									Start: ast.Position{
										Column: 3,
										Line:   9,
									},
								},
							},
							Name: "token",
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
									File:   "influxdb.flux",
									Source: "string",
									Start: ast.Position{
										Column: 10,
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
										File:   "influxdb.flux",
										Source: "string",
										Start: ast.Position{
											Column: 10,
											Line:   9,
										},
									},
								},
								Name: "string",
							},
						},
					}, &ast.ParameterType{
						BaseNode: ast.BaseNode{
							Comments: nil,
							Errors:   nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 14,
									Line:   10,
								},
								File:   "influxdb.flux",
								Source: "?body: bytes",
								Start: ast.Position{
									Column: 2,
									Line:   10,
								},
							},
						},
						Kind: "Optional",
						Name: &ast.Identifier{
							BaseNode: ast.BaseNode{
								Comments: nil,
								Errors:   nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 7,
										Line:   10,
									},
									File:   "influxdb.flux",
									Source: "body",
									Start: ast.Position{
										Column: 3,
										Line:   10,
									},
								},
							},
							Name: "body",
						},
						Ty: &ast.NamedType{
							BaseNode: ast.BaseNode{
								Comments: nil,
								Errors:   nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 14,
										Line:   10,
									},
									File:   "influxdb.flux",
									Source: "bytes",
									Start: ast.Position{
										Column: 9,
										Line:   10,
									},
								},
							},
							ID: &ast.Identifier{
								BaseNode: ast.BaseNode{
									Comments: nil,
									Errors:   nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 14,
											Line:   10,
										},
										File:   "influxdb.flux",
										Source: "bytes",
										Start: ast.Position{
											Column: 9,
											Line:   10,
										},
									},
								},
								Name: "bytes",
							},
						},
					}, &ast.ParameterType{
						BaseNode: ast.BaseNode{
							Comments: nil,
							Errors:   nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 28,
									Line:   11,
								},
								File:   "influxdb.flux",
								Source: "?headers: [string: string]",
								Start: ast.Position{
									Column: 2,
									Line:   11,
								},
							},
						},
						Kind: "Optional",
						Name: &ast.Identifier{
							BaseNode: ast.BaseNode{
								Comments: nil,
								Errors:   nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 10,
										Line:   11,
									},
									File:   "influxdb.flux",
									Source: "headers",
									Start: ast.Position{
										Column: 3,
										Line:   11,
									},
								},
							},
							Name: "headers",
						},
						Ty: &ast.DictType{
							BaseNode: ast.BaseNode{
								Comments: nil,
								Errors:   nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 28,
										Line:   11,
									},
									File:   "influxdb.flux",
									Source: "[string: string]",
									Start: ast.Position{
										Column: 12,
										Line:   11,
									},
								},
							},
							KeyType: &ast.NamedType{
								BaseNode: ast.BaseNode{
									Comments: nil,
									Errors:   nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 19,
											Line:   11,
										},
										File:   "influxdb.flux",
										Source: "string",
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
												Column: 19,
												Line:   11,
											},
											File:   "influxdb.flux",
											Source: "string",
											Start: ast.Position{
												Column: 13,
												Line:   11,
											},
										},
									},
									Name: "string",
								},
							},
							ValueType: &ast.NamedType{
								BaseNode: ast.BaseNode{
									Comments: nil,
									Errors:   nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 27,
											Line:   11,
										},
										File:   "influxdb.flux",
										Source: "string",
										Start: ast.Position{
											Column: 21,
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
												Column: 27,
												Line:   11,
											},
											File:   "influxdb.flux",
											Source: "string",
											Start: ast.Position{
												Column: 21,
												Line:   11,
											},
										},
									},
									Name: "string",
								},
							},
						},
					}, &ast.ParameterType{
						BaseNode: ast.BaseNode{
							Comments: nil,
							Errors:   nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 26,
									Line:   12,
								},
								File:   "influxdb.flux",
								Source: "?query: [string: string]",
								Start: ast.Position{
									Column: 2,
									Line:   12,
								},
							},
						},
						Kind: "Optional",
						Name: &ast.Identifier{
							BaseNode: ast.BaseNode{
								Comments: nil,
								Errors:   nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 8,
										Line:   12,
									},
									File:   "influxdb.flux",
									Source: "query",
									Start: ast.Position{
										Column: 3,
										Line:   12,
									},
								},
							},
							Name: "query",
						},
						Ty: &ast.DictType{
							BaseNode: ast.BaseNode{
								Comments: nil,
								Errors:   nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 26,
										Line:   12,
									},
									File:   "influxdb.flux",
									Source: "[string: string]",
									Start: ast.Position{
										Column: 10,
										Line:   12,
									},
								},
							},
							KeyType: &ast.NamedType{
								BaseNode: ast.BaseNode{
									Comments: nil,
									Errors:   nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 17,
											Line:   12,
										},
										File:   "influxdb.flux",
										Source: "string",
										Start: ast.Position{
											Column: 11,
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
												Column: 17,
												Line:   12,
											},
											File:   "influxdb.flux",
											Source: "string",
											Start: ast.Position{
												Column: 11,
												Line:   12,
											},
										},
									},
									Name: "string",
								},
							},
							ValueType: &ast.NamedType{
								BaseNode: ast.BaseNode{
									Comments: nil,
									Errors:   nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 25,
											Line:   12,
										},
										File:   "influxdb.flux",
										Source: "string",
										Start: ast.Position{
											Column: 19,
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
												Column: 25,
												Line:   12,
											},
											File:   "influxdb.flux",
											Source: "string",
											Start: ast.Position{
												Column: 19,
												Line:   12,
											},
										},
									},
									Name: "string",
								},
							},
						},
					}, &ast.ParameterType{
						BaseNode: ast.BaseNode{
							Comments: nil,
							Errors:   nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 20,
									Line:   13,
								},
								File:   "influxdb.flux",
								Source: "?timeout: duration",
								Start: ast.Position{
									Column: 2,
									Line:   13,
								},
							},
						},
						Kind: "Optional",
						Name: &ast.Identifier{
							BaseNode: ast.BaseNode{
								Comments: nil,
								Errors:   nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 10,
										Line:   13,
									},
									File:   "influxdb.flux",
									Source: "timeout",
									Start: ast.Position{
										Column: 3,
										Line:   13,
									},
								},
							},
							Name: "timeout",
						},
						Ty: &ast.NamedType{
							BaseNode: ast.BaseNode{
								Comments: nil,
								Errors:   nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 20,
										Line:   13,
									},
									File:   "influxdb.flux",
									Source: "duration",
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
											Column: 20,
											Line:   13,
										},
										File:   "influxdb.flux",
										Source: "duration",
										Start: ast.Position{
											Column: 12,
											Line:   13,
										},
									},
								},
								Name: "duration",
							},
						},
					}},
					Return: &ast.RecordType{
						BaseNode: ast.BaseNode{
							Comments: nil,
							Errors:   nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 2,
									Line:   18,
								},
								File:   "influxdb.flux",
								Source: "{\n\tstatusCode: int,\n\tbody: bytes,\n\theaders: [string: string],\n}",
								Start: ast.Position{
									Column: 6,
									Line:   14,
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
										Line:   15,
									},
									File:   "influxdb.flux",
									Source: "statusCode: int",
									Start: ast.Position{
										Column: 2,
										Line:   15,
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
											Line:   15,
										},
										File:   "influxdb.flux",
										Source: "statusCode",
										Start: ast.Position{
											Column: 2,
											Line:   15,
										},
									},
								},
								Name: "statusCode",
							},
							Ty: &ast.NamedType{
								BaseNode: ast.BaseNode{
									Comments: nil,
									Errors:   nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 17,
											Line:   15,
										},
										File:   "influxdb.flux",
										Source: "int",
										Start: ast.Position{
											Column: 14,
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
												Column: 17,
												Line:   15,
											},
											File:   "influxdb.flux",
											Source: "int",
											Start: ast.Position{
												Column: 14,
												Line:   15,
											},
										},
									},
									Name: "int",
								},
							},
						}, &ast.PropertyType{
							BaseNode: ast.BaseNode{
								Comments: nil,
								Errors:   nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 13,
										Line:   16,
									},
									File:   "influxdb.flux",
									Source: "body: bytes",
									Start: ast.Position{
										Column: 2,
										Line:   16,
									},
								},
							},
							Name: &ast.Identifier{
								BaseNode: ast.BaseNode{
									Comments: nil,
									Errors:   nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 6,
											Line:   16,
										},
										File:   "influxdb.flux",
										Source: "body",
										Start: ast.Position{
											Column: 2,
											Line:   16,
										},
									},
								},
								Name: "body",
							},
							Ty: &ast.NamedType{
								BaseNode: ast.BaseNode{
									Comments: nil,
									Errors:   nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 13,
											Line:   16,
										},
										File:   "influxdb.flux",
										Source: "bytes",
										Start: ast.Position{
											Column: 8,
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
											File:   "influxdb.flux",
											Source: "bytes",
											Start: ast.Position{
												Column: 8,
												Line:   16,
											},
										},
									},
									Name: "bytes",
								},
							},
						}, &ast.PropertyType{
							BaseNode: ast.BaseNode{
								Comments: nil,
								Errors:   nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 27,
										Line:   17,
									},
									File:   "influxdb.flux",
									Source: "headers: [string: string]",
									Start: ast.Position{
										Column: 2,
										Line:   17,
									},
								},
							},
							Name: &ast.Identifier{
								BaseNode: ast.BaseNode{
									Comments: nil,
									Errors:   nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 9,
											Line:   17,
										},
										File:   "influxdb.flux",
										Source: "headers",
										Start: ast.Position{
											Column: 2,
											Line:   17,
										},
									},
								},
								Name: "headers",
							},
							Ty: &ast.DictType{
								BaseNode: ast.BaseNode{
									Comments: nil,
									Errors:   nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 27,
											Line:   17,
										},
										File:   "influxdb.flux",
										Source: "[string: string]",
										Start: ast.Position{
											Column: 11,
											Line:   17,
										},
									},
								},
								KeyType: &ast.NamedType{
									BaseNode: ast.BaseNode{
										Comments: nil,
										Errors:   nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 18,
												Line:   17,
											},
											File:   "influxdb.flux",
											Source: "string",
											Start: ast.Position{
												Column: 12,
												Line:   17,
											},
										},
									},
									ID: &ast.Identifier{
										BaseNode: ast.BaseNode{
											Comments: nil,
											Errors:   nil,
											Loc: &ast.SourceLocation{
												End: ast.Position{
													Column: 18,
													Line:   17,
												},
												File:   "influxdb.flux",
												Source: "string",
												Start: ast.Position{
													Column: 12,
													Line:   17,
												},
											},
										},
										Name: "string",
									},
								},
								ValueType: &ast.NamedType{
									BaseNode: ast.BaseNode{
										Comments: nil,
										Errors:   nil,
										Loc: &ast.SourceLocation{
											End: ast.Position{
												Column: 26,
												Line:   17,
											},
											File:   "influxdb.flux",
											Source: "string",
											Start: ast.Position{
												Column: 20,
												Line:   17,
											},
										},
									},
									ID: &ast.Identifier{
										BaseNode: ast.BaseNode{
											Comments: nil,
											Errors:   nil,
											Loc: &ast.SourceLocation{
												End: ast.Position{
													Column: 26,
													Line:   17,
												},
												File:   "influxdb.flux",
												Source: "string",
												Start: ast.Position{
													Column: 20,
													Line:   17,
												},
											},
										},
										Name: "string",
									},
								},
							},
						}},
						Tvar: nil,
					},
				},
			},
		}},
		Eof:      nil,
		Imports:  nil,
		Metadata: "parser-type=rust",
		Name:     "influxdb.flux",
		Package: &ast.PackageClause{
			BaseNode: ast.BaseNode{
				Comments: nil,
				Errors:   nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 17,
						Line:   1,
					},
					File:   "influxdb.flux",
					Source: "package influxdb",
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
						File:   "influxdb.flux",
						Source: "influxdb",
						Start: ast.Position{
							Column: 9,
							Line:   1,
						},
					},
				},
				Name: "influxdb",
			},
		},
	}},
	Package: "influxdb",
	Path:    "experimental/influxdb",
}
