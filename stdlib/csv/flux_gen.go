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
					Column: 13,
					Line:   78,
				},
				File:   "csv.flux",
				Source: "package csv\n\n\n// from is a function that retrieves data from a comma separated value (CSV) data source.\n//\n// A stream of tables are returned, each unique series contained within its own table.\n// Each record in the table represents a single point in the series.\n//\n// ## Parameters\n// - `csv` is CSV data\n//\n//   Supports anonotated CSV or raw CSV. Use mode to specify the parsing mode.\n//\n// - `file` if the file path of the CSV file to query\n//\n//   The path can be absolute or relative. If relative, it is relative to the working\n//   directory of the `fluxd` process. The CSV file must exist in the same file\n//   system running the `fluxd` process.\n//\n// - `mode` is the CSV parsing mode. Default is annotations\n//\n//   Available annotation modes:\n//    annotations: Use CSV notations to determine column data types\n//    raw: Parse all columns as strings and use the first row as the header row\n//    and all subsequent rows as data.\n//\n// ## Query anotated CSV data from file\n// ```\n// import \"csv\"\n//\n// csv.from(file: \"path/to/data-file.csv\")\n// ```\n//\n// ## Query raw data from CSV file\n// ```\n// import \"csv\"\n//\n// csv.from(\n//   file: \"/path/to/data-file.csv\",\n//   mode: \"raw\"\n// )\n// ```\n//\n// ## Query an annotated CSV string\n// ```\n// import \"csv\"\n//\n// csvData = \"\n// #datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,double\n// #group,false,false,false,false,false,false,false,false\n// #default,,,,,,,,\n// ,result,table,_start,_stop,_time,region,host,_value\n// ,mean,0,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:00Z,east,A,15.43\n// ,mean,0,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:20Z,east,B,59.25\n// ,mean,0,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:40Z,east,C,52.62\n// \"\n//\n// csv.from(csv: csvData)\n// ```\n//\n// ## Query a raw CSV string\n// ```\n// import \"csv\"\n//\n// csvData = \"\n// _start,_stop,_time,region,host,_value\n// 2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:00Z,east,A,15.43\n// 2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:20Z,east,B,59.25\n// 2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:40Z,east,C,52.62\n// \"\n//\n// csv.from(\n//   csv: csvData,\n//   mode: \"raw\"\n// )\n// ```\nbuiltin from",
				Start: ast.Position{
					Column: 1,
					Line:   2,
				},
			},
		},
		Body: []ast.Statement{&ast.BuiltinStatement{
			BaseNode: ast.BaseNode{
				Comments: []ast.Comment{ast.Comment{Text: "// from is a function that retrieves data from a comma separated value (CSV) data source.\n"}, ast.Comment{Text: "//\n"}, ast.Comment{Text: "// A stream of tables are returned, each unique series contained within its own table.\n"}, ast.Comment{Text: "// Each record in the table represents a single point in the series.\n"}, ast.Comment{Text: "//\n"}, ast.Comment{Text: "// ## Parameters\n"}, ast.Comment{Text: "// - `csv` is CSV data\n"}, ast.Comment{Text: "//\n"}, ast.Comment{Text: "//   Supports anonotated CSV or raw CSV. Use mode to specify the parsing mode.\n"}, ast.Comment{Text: "//\n"}, ast.Comment{Text: "// - `file` if the file path of the CSV file to query\n"}, ast.Comment{Text: "//\n"}, ast.Comment{Text: "//   The path can be absolute or relative. If relative, it is relative to the working\n"}, ast.Comment{Text: "//   directory of the `fluxd` process. The CSV file must exist in the same file\n"}, ast.Comment{Text: "//   system running the `fluxd` process.\n"}, ast.Comment{Text: "//\n"}, ast.Comment{Text: "// - `mode` is the CSV parsing mode. Default is annotations\n"}, ast.Comment{Text: "//\n"}, ast.Comment{Text: "//   Available annotation modes:\n"}, ast.Comment{Text: "//    annotations: Use CSV notations to determine column data types\n"}, ast.Comment{Text: "//    raw: Parse all columns as strings and use the first row as the header row\n"}, ast.Comment{Text: "//    and all subsequent rows as data.\n"}, ast.Comment{Text: "//\n"}, ast.Comment{Text: "// ## Query anotated CSV data from file\n"}, ast.Comment{Text: "// ```\n"}, ast.Comment{Text: "// import \"csv\"\n"}, ast.Comment{Text: "//\n"}, ast.Comment{Text: "// csv.from(file: \"path/to/data-file.csv\")\n"}, ast.Comment{Text: "// ```\n"}, ast.Comment{Text: "//\n"}, ast.Comment{Text: "// ## Query raw data from CSV file\n"}, ast.Comment{Text: "// ```\n"}, ast.Comment{Text: "// import \"csv\"\n"}, ast.Comment{Text: "//\n"}, ast.Comment{Text: "// csv.from(\n"}, ast.Comment{Text: "//   file: \"/path/to/data-file.csv\",\n"}, ast.Comment{Text: "//   mode: \"raw\"\n"}, ast.Comment{Text: "// )\n"}, ast.Comment{Text: "// ```\n"}, ast.Comment{Text: "//\n"}, ast.Comment{Text: "// ## Query an annotated CSV string\n"}, ast.Comment{Text: "// ```\n"}, ast.Comment{Text: "// import \"csv\"\n"}, ast.Comment{Text: "//\n"}, ast.Comment{Text: "// csvData = \"\n"}, ast.Comment{Text: "// #datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,double\n"}, ast.Comment{Text: "// #group,false,false,false,false,false,false,false,false\n"}, ast.Comment{Text: "// #default,,,,,,,,\n"}, ast.Comment{Text: "// ,result,table,_start,_stop,_time,region,host,_value\n"}, ast.Comment{Text: "// ,mean,0,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:00Z,east,A,15.43\n"}, ast.Comment{Text: "// ,mean,0,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:20Z,east,B,59.25\n"}, ast.Comment{Text: "// ,mean,0,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:40Z,east,C,52.62\n"}, ast.Comment{Text: "// \"\n"}, ast.Comment{Text: "//\n"}, ast.Comment{Text: "// csv.from(csv: csvData)\n"}, ast.Comment{Text: "// ```\n"}, ast.Comment{Text: "//\n"}, ast.Comment{Text: "// ## Query a raw CSV string\n"}, ast.Comment{Text: "// ```\n"}, ast.Comment{Text: "// import \"csv\"\n"}, ast.Comment{Text: "//\n"}, ast.Comment{Text: "// csvData = \"\n"}, ast.Comment{Text: "// _start,_stop,_time,region,host,_value\n"}, ast.Comment{Text: "// 2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:00Z,east,A,15.43\n"}, ast.Comment{Text: "// 2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:20Z,east,B,59.25\n"}, ast.Comment{Text: "// 2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:40Z,east,C,52.62\n"}, ast.Comment{Text: "// \"\n"}, ast.Comment{Text: "//\n"}, ast.Comment{Text: "// csv.from(\n"}, ast.Comment{Text: "//   csv: csvData,\n"}, ast.Comment{Text: "//   mode: \"raw\"\n"}, ast.Comment{Text: "// )\n"}, ast.Comment{Text: "// ```\n"}},
				Errors:   nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 13,
						Line:   78,
					},
					File:   "csv.flux",
					Source: "builtin from",
					Start: ast.Position{
						Column: 1,
						Line:   78,
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
							Column: 13,
							Line:   78,
						},
						File:   "csv.flux",
						Source: "from",
						Start: ast.Position{
							Column: 9,
							Line:   78,
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
							Column: 83,
							Line:   78,
						},
						File:   "csv.flux",
						Source: "(?csv: string, ?file: string, ?mode: string) => [A] where A: Record",
						Start: ast.Position{
							Column: 16,
							Line:   78,
						},
					},
				},
				Constraints: []*ast.TypeConstraint{&ast.TypeConstraint{
					BaseNode: ast.BaseNode{
						Comments: nil,
						Errors:   nil,
						Loc: &ast.SourceLocation{
							End: ast.Position{
								Column: 83,
								Line:   78,
							},
							File:   "csv.flux",
							Source: "A: Record",
							Start: ast.Position{
								Column: 74,
								Line:   78,
							},
						},
					},
					Kinds: []*ast.Identifier{&ast.Identifier{
						BaseNode: ast.BaseNode{
							Comments: nil,
							Errors:   nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 83,
									Line:   78,
								},
								File:   "csv.flux",
								Source: "Record",
								Start: ast.Position{
									Column: 77,
									Line:   78,
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
									Column: 75,
									Line:   78,
								},
								File:   "csv.flux",
								Source: "A",
								Start: ast.Position{
									Column: 74,
									Line:   78,
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
								Column: 67,
								Line:   78,
							},
							File:   "csv.flux",
							Source: "(?csv: string, ?file: string, ?mode: string) => [A]",
							Start: ast.Position{
								Column: 16,
								Line:   78,
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
									Line:   78,
								},
								File:   "csv.flux",
								Source: "?csv: string",
								Start: ast.Position{
									Column: 17,
									Line:   78,
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
										Column: 21,
										Line:   78,
									},
									File:   "csv.flux",
									Source: "csv",
									Start: ast.Position{
										Column: 18,
										Line:   78,
									},
								},
							},
							Name: "csv",
						},
						Ty: &ast.NamedType{
							BaseNode: ast.BaseNode{
								Comments: nil,
								Errors:   nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 29,
										Line:   78,
									},
									File:   "csv.flux",
									Source: "string",
									Start: ast.Position{
										Column: 23,
										Line:   78,
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
											Line:   78,
										},
										File:   "csv.flux",
										Source: "string",
										Start: ast.Position{
											Column: 23,
											Line:   78,
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
									Column: 44,
									Line:   78,
								},
								File:   "csv.flux",
								Source: "?file: string",
								Start: ast.Position{
									Column: 31,
									Line:   78,
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
										Column: 36,
										Line:   78,
									},
									File:   "csv.flux",
									Source: "file",
									Start: ast.Position{
										Column: 32,
										Line:   78,
									},
								},
							},
							Name: "file",
						},
						Ty: &ast.NamedType{
							BaseNode: ast.BaseNode{
								Comments: nil,
								Errors:   nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 44,
										Line:   78,
									},
									File:   "csv.flux",
									Source: "string",
									Start: ast.Position{
										Column: 38,
										Line:   78,
									},
								},
							},
							ID: &ast.Identifier{
								BaseNode: ast.BaseNode{
									Comments: nil,
									Errors:   nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 44,
											Line:   78,
										},
										File:   "csv.flux",
										Source: "string",
										Start: ast.Position{
											Column: 38,
											Line:   78,
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
									Column: 59,
									Line:   78,
								},
								File:   "csv.flux",
								Source: "?mode: string",
								Start: ast.Position{
									Column: 46,
									Line:   78,
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
										Column: 51,
										Line:   78,
									},
									File:   "csv.flux",
									Source: "mode",
									Start: ast.Position{
										Column: 47,
										Line:   78,
									},
								},
							},
							Name: "mode",
						},
						Ty: &ast.NamedType{
							BaseNode: ast.BaseNode{
								Comments: nil,
								Errors:   nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 59,
										Line:   78,
									},
									File:   "csv.flux",
									Source: "string",
									Start: ast.Position{
										Column: 53,
										Line:   78,
									},
								},
							},
							ID: &ast.Identifier{
								BaseNode: ast.BaseNode{
									Comments: nil,
									Errors:   nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 59,
											Line:   78,
										},
										File:   "csv.flux",
										Source: "string",
										Start: ast.Position{
											Column: 53,
											Line:   78,
										},
									},
								},
								Name: "string",
							},
						},
					}},
					Return: &ast.ArrayType{
						BaseNode: ast.BaseNode{
							Comments: nil,
							Errors:   nil,
							Loc: &ast.SourceLocation{
								End: ast.Position{
									Column: 67,
									Line:   78,
								},
								File:   "csv.flux",
								Source: "[A]",
								Start: ast.Position{
									Column: 64,
									Line:   78,
								},
							},
						},
						ElementType: &ast.TvarType{
							BaseNode: ast.BaseNode{
								Comments: nil,
								Errors:   nil,
								Loc: &ast.SourceLocation{
									End: ast.Position{
										Column: 66,
										Line:   78,
									},
									File:   "csv.flux",
									Source: "A",
									Start: ast.Position{
										Column: 65,
										Line:   78,
									},
								},
							},
							ID: &ast.Identifier{
								BaseNode: ast.BaseNode{
									Comments: nil,
									Errors:   nil,
									Loc: &ast.SourceLocation{
										End: ast.Position{
											Column: 66,
											Line:   78,
										},
										File:   "csv.flux",
										Source: "A",
										Start: ast.Position{
											Column: 65,
											Line:   78,
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
		Eof:      nil,
		Imports:  nil,
		Metadata: "parser-type=rust",
		Name:     "csv.flux",
		Package: &ast.PackageClause{
			BaseNode: ast.BaseNode{
				Comments: []ast.Comment{ast.Comment{Text: "// CSV provides an API for working with [annotated CSV](https://github.com/influxdata/flux/blob/master/docs/SPEC.md#csv) files.\n"}},
				Errors:   nil,
				Loc: &ast.SourceLocation{
					End: ast.Position{
						Column: 12,
						Line:   2,
					},
					File:   "csv.flux",
					Source: "package csv",
					Start: ast.Position{
						Column: 1,
						Line:   2,
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
							Line:   2,
						},
						File:   "csv.flux",
						Source: "csv",
						Start: ast.Position{
							Column: 9,
							Line:   2,
						},
					},
				},
				Name: "csv",
			},
		},
	}},
	Package: "csv",
	Path:    "csv",
}
