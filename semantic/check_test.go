package semantic

import (
	"fmt"
	"testing"

	"github.com/influxdata/flux/ast"
)

func TestOptionDeclarations(t *testing.T) {
	testcases := []struct {
		name string
		pkg  *Package
		err  error
	}{
		{
			// package foo
			// option a = 0
			// f = () => {
			//   a = 0
			//   return a + 1
			// }
			//
			name: "no error",
			pkg: &Package{
				Package: "foo",
				Files: []*File{
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "a"},
									Init:       &IntegerLiteral{Value: 0},
								},
							},
						},
					},
				},
			},
		},
		{
			// package foo
			// f = () => {
			//   option bar = 0
			// }
			//
			name: "function block",
			pkg: &Package{
				Package: "foo",
				Files: []*File{
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&NativeVariableAssignment{
								Identifier: &Identifier{Name: "f"},
								Init: &FunctionExpression{
									Block: &FunctionBlock{
										Body: &Block{
											Body: []Statement{
												&OptionStatement{
													Loc: Loc{
														Start: ast.Position{
															Line:   3,
															Column: 3,
														},
														End: ast.Position{
															Line:   3,
															Column: 17,
														},
													},
													Assignment: &NativeVariableAssignment{
														Identifier: &Identifier{Name: "bar"},
														Init:       &IntegerLiteral{Value: 0},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			err: fmt.Errorf(`option "bar" declared below package block at 3:3-3:17`),
		},
		{
			// package foo
			// f = () => {
			//   g = () => {
			//     option bar = 0
			//   }
			// }
			//
			name: "nested function block",
			pkg: &Package{
				Package: "foo",
				Files: []*File{
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&NativeVariableAssignment{
								Identifier: &Identifier{Name: "f"},
								Init: &FunctionExpression{
									Block: &FunctionBlock{
										Body: &Block{
											Body: []Statement{
												&NativeVariableAssignment{
													Identifier: &Identifier{Name: "g"},
													Init: &FunctionExpression{
														Block: &FunctionBlock{
															Body: &Block{
																Body: []Statement{
																	&OptionStatement{
																		Loc: Loc{
																			Start: ast.Position{
																				Line:   4,
																				Column: 5,
																			},
																			End: ast.Position{
																				Line:   4,
																				Column: 19,
																			},
																		},
																		Assignment: &NativeVariableAssignment{
																			Identifier: &Identifier{Name: "bar"},
																			Init:       &IntegerLiteral{Value: 0},
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			err: fmt.Errorf(`option "bar" declared below package block at 4:5-4:19`),
		},
		{
			// package foo
			// import "bar"
			// f = () => {
			//   option bar.baz = 0
			// }
			//
			name: "qualified option",
			pkg: &Package{
				Package: "foo",
				Files: []*File{
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Imports: []*ImportDeclaration{
							{
								Path: &StringLiteral{Value: "bar"},
							},
						},
						Body: []Statement{
							&NativeVariableAssignment{
								Identifier: &Identifier{Name: "f"},
								Init: &FunctionExpression{
									Block: &FunctionBlock{
										Body: &Block{
											Body: []Statement{
												&OptionStatement{
													Loc: Loc{
														Start: ast.Position{
															Line:   4,
															Column: 3,
														},
														End: ast.Position{
															Line:   4,
															Column: 21,
														},
													},
													Assignment: &MemberAssignment{
														Member: &MemberExpression{
															Object:   &IdentifierExpression{Name: "bar"},
															Property: "baz",
														},
														Init: &IntegerLiteral{Value: 0},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			err: fmt.Errorf(`option "bar.baz" declared below package block at 4:3-4:21`),
		},
		{
			// package foo
			// option a = 0
			//
			// package foo
			// import "bar"
			//
			// x = bar.x
			// option bar.x = 0
			//
			// package foo
			// option b = 0
			//
			// f = () => {
			//   a = 1
			//   b = 1
			//   c = 1
			//   return a + b - c
			// }
			//
			// package foo
			// option c = 0
			// g = () => {
			//   option d = "d"
			//   return 0
			// }
			//
			name: "multiple files",
			pkg: &Package{
				Package: "foo",
				Files: []*File{
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "a"},
									Init:       &IntegerLiteral{Value: 0},
								},
							},
						},
					},
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Imports: []*ImportDeclaration{
							{
								Path: &StringLiteral{Value: "bar"},
							},
						},
						Body: []Statement{
							&NativeVariableAssignment{
								Identifier: &Identifier{Name: "x"},
								Init: &MemberExpression{
									Object:   &IdentifierExpression{Name: "bar"},
									Property: "x",
								},
							},
							&OptionStatement{
								Assignment: &MemberAssignment{
									Member: &MemberExpression{
										Object:   &IdentifierExpression{Name: "bar"},
										Property: "x",
									},
									Init: &IntegerLiteral{Value: 0},
								},
							},
						},
					},
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Imports: []*ImportDeclaration{
							{
								Path: &StringLiteral{Value: "bar"},
							},
						},
						Body: []Statement{
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "b"},
									Init:       &IntegerLiteral{Value: 0},
								},
							},
							&NativeVariableAssignment{
								Identifier: &Identifier{Name: "f"},
								Init: &FunctionExpression{
									Block: &FunctionBlock{
										Body: &Block{
											Body: []Statement{
												&NativeVariableAssignment{
													Identifier: &Identifier{Name: "a"},
													Init:       &IntegerLiteral{Value: 1},
												},
												&NativeVariableAssignment{
													Identifier: &Identifier{Name: "b"},
													Init:       &IntegerLiteral{Value: 1},
												},
												&NativeVariableAssignment{
													Identifier: &Identifier{Name: "c"},
													Init:       &IntegerLiteral{Value: 1},
												},
												&ReturnStatement{
													Argument: &BinaryExpression{
														Operator: ast.AdditionOperator,
														Left:     &IdentifierExpression{Name: "a"},
														Right: &BinaryExpression{
															Operator: ast.SubtractionOperator,
															Left:     &IdentifierExpression{Name: "b"},
															Right:    &IdentifierExpression{Name: "c"},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Imports: []*ImportDeclaration{
							{
								Path: &StringLiteral{Value: "bar"},
							},
						},
						Body: []Statement{
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "c"},
									Init:       &IntegerLiteral{Value: 0},
								},
							},
							&NativeVariableAssignment{
								Identifier: &Identifier{Name: "g"},
								Init: &FunctionExpression{
									Block: &FunctionBlock{
										Body: &Block{
											Body: []Statement{
												&OptionStatement{
													Loc: Loc{
														File: "g.flux",
														Start: ast.Position{
															Line:   4,
															Column: 3,
														},
														End: ast.Position{
															Line:   4,
															Column: 17,
														},
													},
													Assignment: &NativeVariableAssignment{
														Identifier: &Identifier{Name: "d"},
														Init:       &StringLiteral{Value: "d"},
													},
												},
												&ReturnStatement{
													Argument: &IntegerLiteral{Value: 0},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			err: fmt.Errorf(`option "d" declared below package block at g.flux|4:3-4:17`),
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := optionStatements(tc.pkg)
			switch {
			case err == nil && tc.err == nil:
				// Test passes
			case err == nil && tc.err != nil:
				t.Errorf("expected error: %v", tc.err)
			case err != nil && tc.err == nil:
				t.Errorf("unexpected error: %v", err)
			case err != nil && tc.err != nil:
				if err.Error() != tc.err.Error() {
					t.Errorf("unexpected result; want err=%v, got err=%v", tc.err, err)
				}
				// else test passes
			}
		})
	}
}

func TestOptionReAssignments(t *testing.T) {
	testcases := []struct {
		name string
		pkg  *Package
		err  error
	}{
		{
			// package foo
			// option a = 0
			// option a = 1
			//
			name: "simple",
			pkg: &Package{
				Package: "foo",
				Files: []*File{
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "a"},
									Init:       &IntegerLiteral{Value: 0},
								},
							},
							&OptionStatement{
								Loc: Loc{
									Start: ast.Position{
										Line:   3,
										Column: 1,
									},
									End: ast.Position{
										Line:   3,
										Column: 13,
									},
								},
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "a"},
									Init:       &IntegerLiteral{Value: 1},
								},
							},
						},
					},
				},
			},
			err: fmt.Errorf(`option "a" redeclared at 3:1-3:13`),
		},
		{
			// package foo
			// option a = 0
			//
			// package foo
			// b = 0
			//
			// package foo
			// option c = 0
			//
			// package foo
			// option c = 1
			//
			name: "multiple files",
			pkg: &Package{
				Package: "foo",
				Files: []*File{
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "a"},
									Init:       &IntegerLiteral{Value: 0},
								},
							},
						},
					},
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&NativeVariableAssignment{
								Identifier: &Identifier{Name: "b"},
								Init:       &IntegerLiteral{Value: 0},
							},
						},
					},
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "c"},
									Init:       &IntegerLiteral{Value: 0},
								},
							},
						},
					},
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&OptionStatement{
								Loc: Loc{
									File: "c.flux",
									Start: ast.Position{
										Line:   2,
										Column: 1,
									},
									End: ast.Position{
										Line:   2,
										Column: 13,
									},
								},
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "c"},
									Init:       &IntegerLiteral{Value: 1},
								},
							},
						},
					},
				},
			},
			err: fmt.Errorf(`option "c" redeclared at c.flux|2:1-2:13`),
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			vars := make(map[string]bool)
			opts := make(map[string]bool)
			err := runChecks(tc.pkg, vars, opts)
			switch {
			case err == nil && tc.err == nil:
				// Test passes
			case err == nil && tc.err != nil:
				t.Errorf("expected error: %v", tc.err)
			case err != nil && tc.err == nil:
				t.Errorf("unexpected error: %v", err)
			case err != nil && tc.err != nil:
				if err.Error() != tc.err.Error() {
					t.Errorf("unexpected result; want err=%v, got err=%v", tc.err, err)
				}
				// else test passes
			}
		})
	}
}

func TestVarReAssignments(t *testing.T) {
	testcases := []struct {
		name string
		skip string
		pkg  *Package
		err  error
	}{
		{
			// package foo
			// a = 0
			// b = 1
			// c = 2
			//
			name: "no error",
			pkg: &Package{
				Package: "foo",
				Files: []*File{
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&NativeVariableAssignment{
								Identifier: &Identifier{Name: "a"},
								Init:       &IntegerLiteral{Value: 0},
							},
							&NativeVariableAssignment{
								Identifier: &Identifier{Name: "b"},
								Init:       &IntegerLiteral{Value: 1},
							},
							&NativeVariableAssignment{
								Identifier: &Identifier{Name: "c"},
								Init:       &IntegerLiteral{Value: 2},
							},
						},
					},
				},
			},
		},
		{
			// package foo
			// a = 0
			//
			// package foo
			// b = 0
			//
			name: "no error multiple files",
			pkg: &Package{
				Package: "foo",
				Files: []*File{
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&NativeVariableAssignment{
								Identifier: &Identifier{Name: "a"},
								Init:       &IntegerLiteral{Value: 0},
							},
						},
					},
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&NativeVariableAssignment{
								Identifier: &Identifier{Name: "b"},
								Init:       &IntegerLiteral{Value: 0},
							},
						},
					},
				},
			},
		},
		{
			// package foo
			// a = 0
			// a = 1
			//
			name: "redeclaration",
			pkg: &Package{
				Package: "foo",
				Files: []*File{
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&NativeVariableAssignment{
								Identifier: &Identifier{Name: "a"},
								Init:       &IntegerLiteral{Value: 0},
							},
							&NativeVariableAssignment{
								Loc: Loc{
									Start: ast.Position{
										Line:   3,
										Column: 1,
									},
									End: ast.Position{
										Line:   3,
										Column: 6,
									},
								},
								Identifier: &Identifier{Name: "a"},
								Init:       &IntegerLiteral{Value: 1},
							},
						},
					},
				},
			},
			err: fmt.Errorf(`var "a" redeclared at 3:1-3:6`),
		},
		{
			// package foo
			// option a = 0
			// a = 1
			//
			name: "redec option",
			pkg: &Package{
				Package: "foo",
				Files: []*File{
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "a"},
									Init:       &IntegerLiteral{Value: 0},
								},
							},
							&NativeVariableAssignment{
								Loc: Loc{
									Start: ast.Position{
										Line:   3,
										Column: 1,
									},
									End: ast.Position{
										Line:   3,
										Column: 6,
									},
								},
								Identifier: &Identifier{Name: "a"},
								Init:       &IntegerLiteral{Value: 1},
							},
						},
					},
				},
			},
			err: fmt.Errorf(`cannot declare variable "a" at 3:1-3:6; option with same name already declared`),
		},
		{
			// package foo
			// a = 0
			// f = () => {
			//   a = 2
			//   return a
			// }
			//
			name: "shadow",
			pkg: &Package{
				Package: "foo",
				Files: []*File{
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&NativeVariableAssignment{
								Identifier: &Identifier{Name: "a"},
								Init:       &IntegerLiteral{Value: 0},
							},
							&NativeVariableAssignment{
								Identifier: &Identifier{Name: "f"},
								Init: &FunctionExpression{
									Block: &FunctionBlock{
										Body: &Block{
											Body: []Statement{
												&NativeVariableAssignment{
													Identifier: &Identifier{Name: "a"},
													Init:       &IntegerLiteral{Value: 2},
												},
												&ReturnStatement{
													Argument: &IdentifierExpression{Name: "a"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			// package foo
			// f = () => {
			//   a = 2
			//   return a
			// }
			// a = 0
			//
			name: "after shadow",
			pkg: &Package{
				Package: "foo",
				Files: []*File{
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&NativeVariableAssignment{
								Identifier: &Identifier{Name: "f"},
								Init: &FunctionExpression{
									Block: &FunctionBlock{
										Body: &Block{
											Body: []Statement{
												&NativeVariableAssignment{
													Identifier: &Identifier{Name: "a"},
													Init:       &IntegerLiteral{Value: 2},
												},
												&ReturnStatement{
													Argument: &IdentifierExpression{Name: "a"},
												},
											},
										},
									},
								},
							},
							&NativeVariableAssignment{
								Identifier: &Identifier{Name: "a"},
								Init:       &IntegerLiteral{Value: 0},
							},
						},
					},
				},
			},
		},
		{
			// package foo
			// a = 0
			// f = () => {
			//   a = 1
			//   b = a
			//   b = 1
			//   return b
			// }
			//
			name: "redeclaration inside function",
			pkg: &Package{
				Package: "foo",
				Files: []*File{
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&NativeVariableAssignment{
								Identifier: &Identifier{Name: "a"},
								Init:       &IntegerLiteral{Value: 0},
							},
							&NativeVariableAssignment{
								Identifier: &Identifier{Name: "f"},
								Init: &FunctionExpression{
									Block: &FunctionBlock{
										Body: &Block{
											Body: []Statement{
												&NativeVariableAssignment{
													Identifier: &Identifier{Name: "a"},
													Init:       &IntegerLiteral{Value: 1},
												},
												&NativeVariableAssignment{
													Identifier: &Identifier{Name: "b"},
													Init:       &IdentifierExpression{Name: "a"},
												},
												&NativeVariableAssignment{
													Loc: Loc{
														Start: ast.Position{
															Line:   6,
															Column: 3,
														},
														End: ast.Position{
															Line:   6,
															Column: 8,
														},
													},
													Identifier: &Identifier{Name: "b"},
													Init:       &IntegerLiteral{Value: 1},
												},
												&ReturnStatement{
													Argument: &IdentifierExpression{Name: "b"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			err: fmt.Errorf(`var "b" redeclared at 6:3-6:8`),
		},
		{
			// package foo
			// a = 0
			// option f = () => {
			//   a = 1
			//   b = a
			//   b = 1
			//   return b
			// }
			//
			name: "redeclaration inside option expression",
			pkg: &Package{
				Package: "foo",
				Files: []*File{
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&NativeVariableAssignment{
								Identifier: &Identifier{Name: "a"},
								Init:       &IntegerLiteral{Value: 0},
							},
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "f"},
									Init: &FunctionExpression{
										Block: &FunctionBlock{
											Body: &Block{
												Body: []Statement{
													&NativeVariableAssignment{
														Identifier: &Identifier{Name: "a"},
														Init:       &IntegerLiteral{Value: 1},
													},
													&NativeVariableAssignment{
														Identifier: &Identifier{Name: "b"},
														Init:       &IdentifierExpression{Name: "a"},
													},
													&NativeVariableAssignment{
														Loc: Loc{
															Start: ast.Position{
																Line:   6,
																Column: 3,
															},
															End: ast.Position{
																Line:   6,
																Column: 8,
															},
														},
														Identifier: &Identifier{Name: "b"},
														Init:       &IntegerLiteral{Value: 1},
													},
													&ReturnStatement{
														Argument: &IdentifierExpression{Name: "b"},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			err: fmt.Errorf(`var "b" redeclared at 6:3-6:8`),
		},
		{
			// package foo
			// f = (a) => {
			//   a = 1
			//   return a
			// }
			//
			name: "reassign parameter",
			skip: "reassigning a param inside a function's body should error (https://github.com/influxdata/flux/issues/857)",
			pkg: &Package{
				Package: "foo",
				Files: []*File{
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&NativeVariableAssignment{
								Identifier: &Identifier{Name: "f"},
								Init: &FunctionExpression{
									Block: &FunctionBlock{
										Parameters: &FunctionParameters{
											List: []*FunctionParameter{
												{
													Key: &Identifier{Name: "a"},
												},
											},
										},
										Body: &Block{
											Body: []Statement{
												&NativeVariableAssignment{
													Loc: Loc{
														Start: ast.Position{
															Line:   3,
															Column: 3,
														},
														End: ast.Position{
															Line:   3,
															Column: 8,
														},
													},
													Identifier: &Identifier{Name: "a"},
													Init:       &IntegerLiteral{Value: 1},
												},
												&ReturnStatement{
													Argument: &IdentifierExpression{Name: "a"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			err: fmt.Errorf(`var "a" redeclared at 3:3-3:8`),
		},
		{
			// package foo
			// option bar = () => {
			//   bar = 0
			//   return bar
			// }
			//
			name: "no error option",
			pkg: &Package{
				Package: "foo",
				Files: []*File{
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "bar"},
									Init: &FunctionExpression{
										Block: &FunctionBlock{
											Body: &Block{
												Body: []Statement{
													&NativeVariableAssignment{
														Identifier: &Identifier{Name: "bar"},
														Init:       &IntegerLiteral{Value: 0},
													},
													&ReturnStatement{
														Argument: &IdentifierExpression{Name: "bar"},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			// package foo
			//
			// x = 0
			//
			// f = () => {
			//   a = 0
			//   b = 0
			//   return a + b
			// }
			//
			// x = 1
			//
			name: "redec after function",
			pkg: &Package{
				Package: "foo",
				Files: []*File{
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&NativeVariableAssignment{
								Identifier: &Identifier{Name: "x"},
								Init:       &IntegerLiteral{Value: 0},
							},
							&NativeVariableAssignment{
								Identifier: &Identifier{Name: "f"},
								Init: &FunctionExpression{
									Block: &FunctionBlock{
										Body: &Block{
											Body: []Statement{
												&NativeVariableAssignment{
													Identifier: &Identifier{Name: "a"},
													Init:       &IntegerLiteral{Value: 0},
												},
												&NativeVariableAssignment{
													Identifier: &Identifier{Name: "b"},
													Init:       &IntegerLiteral{Value: 0},
												},
												&ReturnStatement{
													Argument: &BinaryExpression{
														Operator: ast.AdditionOperator,
														Left:     &IdentifierExpression{Name: "a"},
														Right:    &IdentifierExpression{Name: "b"},
													},
												},
											},
										},
									},
								},
							},
							&NativeVariableAssignment{
								Loc: Loc{
									Start: ast.Position{
										Line:   11,
										Column: 1,
									},
									End: ast.Position{
										Line:   11,
										Column: 6,
									},
								},
								Identifier: &Identifier{Name: "x"},
								Init:       &IntegerLiteral{Value: 1},
							},
						},
					},
				},
			},
			err: fmt.Errorf(`var "x" redeclared at 11:1-11:6`),
		},
		{
			// package foo
			// a = 0
			// d = a
			//
			// package foo
			//
			// b = 0
			//
			// f = () => {
			//   a = 0
			//   b = 0
			//   c = 0
			//   return a + b + c
			// }
			//
			// c = 0
			//
			// package foo
			//
			// g = (a, b, c) => {
			//   f = (a, b, c) => a + b + c
			//   return f(a: a, b: b, c: c)
			// }
			//
			// d = g(a: 0, b: 1, c: 2)
			//
			name: "redeclaration multiple files",
			pkg: &Package{
				Package: "foo",
				Files: []*File{
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&NativeVariableAssignment{
								Identifier: &Identifier{Name: "a"},
								Init:       &IntegerLiteral{Value: 0},
							},
							&NativeVariableAssignment{
								Identifier: &Identifier{Name: "d"},
								Init:       &IdentifierExpression{Name: "a"},
							},
						},
					},
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&NativeVariableAssignment{
								Identifier: &Identifier{Name: "b"},
								Init:       &IntegerLiteral{Value: 0},
							},
							&NativeVariableAssignment{
								Identifier: &Identifier{Name: "f"},
								Init: &FunctionExpression{
									Block: &FunctionBlock{
										Body: &Block{
											Body: []Statement{
												&NativeVariableAssignment{
													Identifier: &Identifier{Name: "a"},
													Init:       &IntegerLiteral{Value: 0},
												},
												&NativeVariableAssignment{
													Identifier: &Identifier{Name: "b"},
													Init:       &IntegerLiteral{Value: 0},
												},
												&NativeVariableAssignment{
													Identifier: &Identifier{Name: "c"},
													Init:       &IntegerLiteral{Value: 0},
												},
												&ReturnStatement{
													Argument: &BinaryExpression{
														Operator: ast.AdditionOperator,
														Left:     &IdentifierExpression{Name: "a"},
														Right: &BinaryExpression{
															Operator: ast.AdditionOperator,
															Left:     &IdentifierExpression{Name: "b"},
															Right:    &IdentifierExpression{Name: "c"},
														},
													},
												},
											},
										},
									},
								},
							},
							&NativeVariableAssignment{
								Identifier: &Identifier{Name: "c"},
								Init:       &IntegerLiteral{Value: 0},
							},
						},
					},
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&NativeVariableAssignment{
								Identifier: &Identifier{Name: "g"},
								Init: &FunctionExpression{
									Block: &FunctionBlock{
										Parameters: &FunctionParameters{
											List: []*FunctionParameter{
												{
													Key: &Identifier{Name: "a"},
												},
												{
													Key: &Identifier{Name: "b"},
												},
												{
													Key: &Identifier{Name: "c"},
												},
											},
										},
										Body: &Block{
											Body: []Statement{
												&NativeVariableAssignment{
													Identifier: &Identifier{Name: "f"},
													Init: &FunctionExpression{
														Block: &FunctionBlock{
															Parameters: &FunctionParameters{
																List: []*FunctionParameter{
																	{
																		Key: &Identifier{Name: "a"},
																	},
																	{
																		Key: &Identifier{Name: "b"},
																	},
																	{
																		Key: &Identifier{Name: "c"},
																	},
																},
															},
															Body: &BinaryExpression{
																Operator: ast.AdditionOperator,
																Left:     &IdentifierExpression{Name: "a"},
																Right: &BinaryExpression{
																	Operator: ast.AdditionOperator,
																	Left:     &IdentifierExpression{Name: "b"},
																	Right:    &IdentifierExpression{Name: "c"},
																},
															},
														},
													},
												},
												&ReturnStatement{
													Argument: &CallExpression{
														Callee: &IdentifierExpression{Name: "f"},
														Arguments: &ObjectExpression{
															Properties: []*Property{
																{
																	Key:   &Identifier{Name: "a"},
																	Value: &IdentifierExpression{Name: "a"},
																},
																{
																	Key:   &Identifier{Name: "b"},
																	Value: &IdentifierExpression{Name: "b"},
																},
																{
																	Key:   &Identifier{Name: "b"},
																	Value: &IdentifierExpression{Name: "b"},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
							&NativeVariableAssignment{
								Loc: Loc{
									File: "d.flux",
									Start: ast.Position{
										Line:   8,
										Column: 1,
									},
									End: ast.Position{
										Line:   8,
										Column: 24,
									},
								},
								Identifier: &Identifier{Name: "d"},
								Init: &CallExpression{
									Callee: &IdentifierExpression{Name: "g"},
									Arguments: &ObjectExpression{
										Properties: []*Property{
											{
												Key:   &Identifier{Name: "a"},
												Value: &IntegerLiteral{Value: 0},
											},
											{
												Key:   &Identifier{Name: "b"},
												Value: &IntegerLiteral{Value: 1},
											},
											{
												Key:   &Identifier{Name: "b"},
												Value: &IntegerLiteral{Value: 2},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			err: fmt.Errorf(`var "d" redeclared at d.flux|8:1-8:24`),
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skipf("skipping %s: %s", tc.name, tc.skip)
			}
			vars := make(map[string]bool)
			opts := make(map[string]bool)
			err := runChecks(tc.pkg, vars, opts)
			switch {
			case err == nil && tc.err == nil:
				// Test passes
			case err == nil && tc.err != nil:
				t.Errorf("expected error: %v", tc.err)
			case err != nil && tc.err == nil:
				t.Errorf("unexpected error: %v", err)
			case err != nil && tc.err != nil:
				if err.Error() != tc.err.Error() {
					t.Errorf("unexpected result; want err=%v, got err=%v", tc.err, err)
				}
				// else test passes
			}
		})
	}
}

func TestOptionDependencies(t *testing.T) {
	testcases := []struct {
		name string
		pkg  *Package
		err  error
	}{
		{
			// package foo
			// option bar = 0
			//
			// package foo
			// option baz = 0
			//
			name: "no error",
			pkg: &Package{
				Package: "foo",
				Files: []*File{
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "bar"},
									Init:       &IntegerLiteral{Value: 0},
								},
							},
						},
					},
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "baz"},
									Init:       &IntegerLiteral{Value: 0},
								},
							},
						},
					},
				},
			},
		},
		{
			// package foo
			// option a = 0
			// option b = a
			//
			name: "dependency",
			pkg: &Package{
				Package: "foo",
				Files: []*File{
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "a"},
									Init:       &IntegerLiteral{Value: 0},
								},
							},
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "b"},
									Init: &IdentifierExpression{
										Loc: Loc{
											Start: ast.Position{
												Line:   3,
												Column: 12,
											},
											End: ast.Position{
												Line:   3,
												Column: 13,
											},
										},
										Name: "a",
									},
								},
							},
						},
					},
				},
			},
			err: fmt.Errorf(`option dependency: option "b" depends on option "a" defined in the same package at 3:12-3:13`),
		},
		{
			// package foo
			// option a = 0
			//
			// package foo
			// option b = a
			//
			name: "dependency across files",
			pkg: &Package{
				Package: "foo",
				Files: []*File{
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "a"},
									Init:       &IntegerLiteral{Value: 0},
								},
							},
						},
					},
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "b"},
									Init: &IdentifierExpression{
										Loc: Loc{
											File: "b.flux",
											Start: ast.Position{
												Line:   2,
												Column: 12,
											},
											End: ast.Position{
												Line:   2,
												Column: 13,
											},
										},
										Name: "a",
									},
								},
							},
						},
					},
				},
			},
			err: fmt.Errorf(`option dependency: option "b" depends on option "a" defined in the same package at b.flux|2:12-2:13`),
		},
		{
			// package foo
			// import "bar"
			// option a = bar.x
			//
			name: "dependency on export",
			pkg: &Package{
				Package: "foo",
				Files: []*File{
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Imports: []*ImportDeclaration{
							{
								Path: &StringLiteral{Value: "bar"},
							},
						},
						Body: []Statement{
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "a"},
									Init: &MemberExpression{
										Object:   &IdentifierExpression{Name: "bar"},
										Property: "x",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			// package foo
			// import "bar"
			//
			// option a = bar.a.x
			//
			name: "option with same name as export",
			pkg: &Package{
				Package: "foo",
				Files: []*File{
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Imports: []*ImportDeclaration{
							{
								Path: &StringLiteral{Value: "bar"},
							},
						},
						Body: []Statement{
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "a"},
									Init: &MemberExpression{
										Object: &MemberExpression{
											Object:   &IdentifierExpression{Name: "bar"},
											Property: "a",
										},
										Property: "x",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			// package foo
			// option a = 0
			// option f = () => a
			//
			name: "nested dependency",
			pkg: &Package{
				Package: "foo",
				Files: []*File{
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "a"},
									Init:       &IntegerLiteral{Value: 0},
								},
							},
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "f"},
									Init: &FunctionExpression{
										Block: &FunctionBlock{
											Body: &IdentifierExpression{
												Loc: Loc{
													Start: ast.Position{
														Line:   3,
														Column: 18,
													},
													End: ast.Position{
														Line:   3,
														Column: 19,
													},
												},
												Name: "a",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			err: fmt.Errorf(`option dependency: option "f" depends on option "a" defined in the same package at 3:18-3:19`),
		},
		{
			// package foo
			// option a = 0
			// option f = () => (() => a)()
			//
			name: "nested nested dependency",
			pkg: &Package{
				Package: "foo",
				Files: []*File{
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "a"},
									Init:       &IntegerLiteral{Value: 0},
								},
							},
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "f"},
									Init: &FunctionExpression{
										Block: &FunctionBlock{
											Body: &CallExpression{
												Callee: &FunctionExpression{
													Block: &FunctionBlock{
														Body: &IdentifierExpression{
															Loc: Loc{
																Start: ast.Position{
																	Line:   3,
																	Column: 25,
																},
																End: ast.Position{
																	Line:   3,
																	Column: 26,
																},
															},
															Name: "a",
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			err: fmt.Errorf(`option dependency: option "f" depends on option "a" defined in the same package at 3:25-3:26`),
		},
		{
			// package foo
			// option a = 0
			//
			// option f = () => {
			//   a = 1
			//   return a + 1
			// }
			//
			name: "shadow",
			pkg: &Package{
				Package: "foo",
				Files: []*File{
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "a"},
									Init:       &IntegerLiteral{Value: 0},
								},
							},
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "f"},
									Init: &FunctionExpression{
										Block: &FunctionBlock{
											Body: &Block{
												Body: []Statement{
													&NativeVariableAssignment{
														Identifier: &Identifier{Name: "a"},
														Init:       &IntegerLiteral{Value: 1},
													},
													&ReturnStatement{
														Argument: &BinaryExpression{
															Operator: ast.AdditionOperator,
															Left:     &IdentifierExpression{Name: "a"},
															Right:    &IntegerLiteral{Value: 1},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			// package foo
			// option a = 0
			// option f = (a) => a
			//
			name: "param shadow",
			pkg: &Package{
				Package: "foo",
				Files: []*File{
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "a"},
									Init:       &IntegerLiteral{Value: 0},
								},
							},
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "f"},
									Init: &FunctionExpression{
										Block: &FunctionBlock{
											Parameters: &FunctionParameters{
												List: []*FunctionParameter{
													{
														Key: &Identifier{Name: "a"},
													},
												},
											},
											Body: &IdentifierExpression{Name: "a"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			// package foo
			// option a = 0
			//
			// option f = () => {
			//   a = 1
			//   return (() => a + 1)()
			// }
			//
			name: "nested shadow",
			pkg: &Package{
				Package: "foo",
				Files: []*File{
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "a"},
									Init:       &IntegerLiteral{Value: 0},
								},
							},
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "f"},
									Init: &FunctionExpression{
										Block: &FunctionBlock{
											Body: &Block{
												Body: []Statement{
													&NativeVariableAssignment{
														Identifier: &Identifier{Name: "a"},
														Init:       &IntegerLiteral{Value: 1},
													},
													&ReturnStatement{
														Argument: &CallExpression{
															Callee: &FunctionExpression{
																Block: &FunctionBlock{
																	Body: &BinaryExpression{
																		Operator: ast.AdditionOperator,
																		Left:     &IdentifierExpression{Name: "a"},
																		Right:    &IntegerLiteral{Value: 1},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			// package foo
			// import "bar"
			//
			// option bar = {x: 0}
			// option a = bar.x
			//
			name: "option that shadows import",
			pkg: &Package{
				Package: "foo",
				Files: []*File{
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Imports: []*ImportDeclaration{
							{
								Path: &StringLiteral{Value: "bar"},
							},
						},
						Body: []Statement{
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "bar"},
									Init: &ObjectExpression{
										Properties: []*Property{
											{
												Key:   &Identifier{Name: "x"},
												Value: &IntegerLiteral{Value: 0},
											},
										},
									},
								},
							},
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "a"},
									Init: &MemberExpression{
										Object: &IdentifierExpression{
											Loc: Loc{
												Start: ast.Position{
													Line:   5,
													Column: 12,
												},
												End: ast.Position{
													Line:   5,
													Column: 15,
												},
											},
											Name: "bar",
										},
										Property: "x",
									},
								},
							},
						},
					},
				},
			},
			err: fmt.Errorf(`option dependency: option "a" depends on option "bar" defined in the same package at 5:12-5:15`),
		},
		{
			// package foo
			// option a = 0
			//
			// option f = () => {
			//   a = 1
			//   return a + 1
			// }
			//
			// option b = a
			//
			name: "dependency after shadow",
			pkg: &Package{
				Package: "foo",
				Files: []*File{
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "a"},
									Init:       &IntegerLiteral{Value: 0},
								},
							},
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "f"},
									Init: &FunctionExpression{
										Block: &FunctionBlock{
											Body: &Block{
												Body: []Statement{
													&NativeVariableAssignment{
														Identifier: &Identifier{Name: "a"},
														Init:       &IntegerLiteral{Value: 0},
													},
													&ReturnStatement{
														Argument: &BinaryExpression{
															Operator: ast.AdditionOperator,
															Left:     &IdentifierExpression{Name: "a"},
															Right:    &IntegerLiteral{Value: 1},
														},
													},
												},
											},
										},
									},
								},
							},
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "b"},
									Init: &IdentifierExpression{
										Loc: Loc{
											Start: ast.Position{
												Line:   9,
												Column: 12,
											},
											End: ast.Position{
												Line:   9,
												Column: 13,
											},
										},
										Name: "a",
									},
								},
							},
						},
					},
				},
			},
			err: fmt.Errorf(`option dependency: option "b" depends on option "a" defined in the same package at 9:12-9:13`),
		},
		{
			// package foo
			// option a = 0
			//
			// package foo
			// option f = () => {
			//   a = 1
			//   return a + 1
			// }
			//
			// package foo
			// option g = (f) => {
			//   a = 1
			//   g = (g) => g |> f
			//   h = (b) => g(g: b)
			//   return h(b: a)
			// }
			//
			// package foo
			// option b = a
			//
			name: "dependency with multiple files and shadows",
			pkg: &Package{
				Package: "foo",
				Files: []*File{
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "a"},
									Init:       &IntegerLiteral{Value: 0},
								},
							},
						},
					},
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "f"},
									Init: &FunctionExpression{
										Block: &FunctionBlock{
											Body: &Block{
												Body: []Statement{
													&NativeVariableAssignment{
														Identifier: &Identifier{Name: "a"},
														Init:       &IntegerLiteral{Value: 1},
													},
													&ReturnStatement{
														Argument: &BinaryExpression{
															Operator: ast.AdditionOperator,
															Left:     &IdentifierExpression{Name: "a"},
															Right:    &IntegerLiteral{Value: 1},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "g"},
									Init: &FunctionExpression{
										Block: &FunctionBlock{
											Parameters: &FunctionParameters{
												List: []*FunctionParameter{
													{
														Key: &Identifier{Name: "f"},
													},
												},
											},
											Body: &Block{
												Body: []Statement{
													&NativeVariableAssignment{
														Identifier: &Identifier{Name: "a"},
														Init:       &IntegerLiteral{Value: 1},
													},
													&NativeVariableAssignment{
														Identifier: &Identifier{Name: "g"},
														Init: &FunctionExpression{
															Block: &FunctionBlock{
																Parameters: &FunctionParameters{
																	List: []*FunctionParameter{
																		{
																			Key: &Identifier{Name: "g"},
																		},
																	},
																},
																Body: &CallExpression{
																	Callee: &IdentifierExpression{Name: "f"},
																	Pipe:   &IdentifierExpression{Name: "g"},
																},
															},
														},
													},
													&NativeVariableAssignment{
														Identifier: &Identifier{Name: "h"},
														Init: &FunctionExpression{
															Block: &FunctionBlock{
																Parameters: &FunctionParameters{
																	List: []*FunctionParameter{
																		{
																			Key: &Identifier{Name: "b"},
																		},
																	},
																},
																Body: &CallExpression{
																	Callee: &IdentifierExpression{Name: "g"},
																	Arguments: &ObjectExpression{
																		Properties: []*Property{
																			{
																				Key:   &Identifier{Name: "g"},
																				Value: &IdentifierExpression{Name: "b"},
																			},
																		},
																	},
																},
															},
														},
													},
													&ReturnStatement{
														Argument: &CallExpression{
															Callee: &IdentifierExpression{Name: "h"},
															Arguments: &ObjectExpression{
																Properties: []*Property{
																	{
																		Key:   &Identifier{Name: "b"},
																		Value: &IdentifierExpression{Name: "a"},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					{
						Package: &PackageClause{
							Name: &Identifier{Name: "foo"},
						},
						Body: []Statement{
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "b"},
									Init: &IdentifierExpression{
										Loc: Loc{
											File: "b.flux",
											Start: ast.Position{
												Line:   2,
												Column: 12,
											},
											End: ast.Position{
												Line:   2,
												Column: 13,
											},
										},
										Name: "a",
									},
								},
							},
						},
					},
				},
			},
			err: fmt.Errorf(`option dependency: option "b" depends on option "a" defined in the same package at b.flux|2:12-2:13`),
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			vars := make(map[string]bool)
			opts := make(map[string]bool)
			err := runChecks(tc.pkg, vars, opts)
			switch {
			case err == nil && tc.err == nil:
				// Test passes
			case err == nil && tc.err != nil:
				t.Errorf("expected error: %v", tc.err)
			case err != nil && tc.err == nil:
				t.Errorf("unexpected error: %v", err)
			case err != nil && tc.err != nil:
				if err.Error() != tc.err.Error() {
					t.Errorf("unexpected result; want err=%v, got err=%v", tc.err, err)
				}
				// else test passes
			}
		})
	}
}
