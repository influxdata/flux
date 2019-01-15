package semantic

import (
	"fmt"
	"testing"

	"github.com/influxdata/flux/ast"
)

func TestOptionDeclarationChecks(t *testing.T) {
	testcases := []struct {
		name string
		pkg  *Package
		err  error
	}{
		{
			// package foo
			// option bar = 0
			// option bar = 1
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
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "bar"},
									Init:       &IntegerLiteral{Value: 1},
								},
							},
						},
					},
				},
			},
		},
		{
			// package foo
			// option bar = 0
			// f = () => {
			//   a = 0
			//   b = 0
			//   return a + b
			// }
			// option bar = 1
			//
			name: "no error after block",
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
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "bar"},
									Init:       &IntegerLiteral{Value: 1},
								},
							},
						},
					},
				},
			},
		},
		{
			// package foo
			// option bar = 0
			//
			// package foo
			// option baz = 0
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
			err: fmt.Errorf(`option "bar" declared below package block at 0:0-0:0`),
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
			err: fmt.Errorf(`option "bar" declared below package block at 0:0-0:0`),
		},
		{
			// package foo
			// option bar = 0
			//
			// package foo
			// f = () => {
			//   option bar = 1
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
							&NativeVariableAssignment{
								Identifier: &Identifier{Name: "f"},
								Init: &FunctionExpression{
									Block: &FunctionBlock{
										Body: &Block{
											Body: []Statement{
												&OptionStatement{
													Assignment: &NativeVariableAssignment{
														Identifier: &Identifier{Name: "bar"},
														Init:       &IntegerLiteral{Value: 1},
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
			err: fmt.Errorf(`option "bar" declared below package block at 0:0-0:0`),
		},
		{
			// package foo
			// import "bar"
			// f = () => {
			//   option bar.baz = 0
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
			err: fmt.Errorf(`option "bar.baz" declared below package block at 0:0-0:0`),
		},
		{
			// package foo
			//
			// option bar = 0
			//
			// f = () => {
			//   a = 0
			//   b = 0
			//   return a + b
			// }
			//
			// option baz = 0
			//
			// g = () => {
			//   option baz = 0
			// }
			//
			name: "after function block",
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
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "baz"},
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
						},
					},
				},
			},
			err: fmt.Errorf(`option "baz" declared below package block at 0:0-0:0`),
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := checkOptionDecs(tc.pkg)
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

func TestVarAssignmentChecks(t *testing.T) {
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
			// option a = 0
			// option a = 1
			// b = 2
			//
			name: "no error with options",
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
									Identifier: &Identifier{Name: "a"},
									Init:       &IntegerLiteral{Value: 1},
								},
							},
							&NativeVariableAssignment{
								Identifier: &Identifier{Name: "b"},
								Init:       &IntegerLiteral{Value: 2},
							},
						},
					},
				},
			},
		},
		{
			// package foo
			// option a = 0
			// option a = 1
			// option b = 2
			// b = 3
			//
			name: "error with options",
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
									Identifier: &Identifier{Name: "a"},
									Init:       &IntegerLiteral{Value: 1},
								},
							},
							&OptionStatement{
								Assignment: &NativeVariableAssignment{
									Identifier: &Identifier{Name: "b"},
									Init:       &IntegerLiteral{Value: 2},
								},
							},
							&NativeVariableAssignment{
								Identifier: &Identifier{Name: "b"},
								Init:       &IntegerLiteral{Value: 3},
							},
						},
					},
				},
			},
			err: fmt.Errorf(`var "b" redeclared at 0:0-0:0`),
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
								Identifier: &Identifier{Name: "a"},
								Init:       &IntegerLiteral{Value: 1},
							},
						},
					},
				},
			},
			err: fmt.Errorf(`var "a" redeclared at 0:0-0:0`),
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
			err: fmt.Errorf(`var "b" redeclared at 0:0-0:0`),
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
			err: fmt.Errorf(`var "b" redeclared at 0:0-0:0`),
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
			err: fmt.Errorf(`var "a" redeclared at 0:0-0:0`),
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
								Identifier: &Identifier{Name: "x"},
								Init:       &IntegerLiteral{Value: 1},
							},
						},
					},
				},
			},
			err: fmt.Errorf(`var "x" redeclared at 0:0-0:0`),
		},
		{
			// package foo
			//
			// f = () => {
			//   a = 0
			//   b = 0
			//   return a + b
			// }
			//
			// a = 1
			//
			name: "redec after block",
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
								Identifier: &Identifier{Name: "a"},
								Init:       &IntegerLiteral{Value: 1},
							},
						},
					},
				},
			},
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
			err: fmt.Errorf(`var "d" redeclared at 0:0-0:0`),
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skipf("skipping %s: %s", tc.name, tc.skip)
			}
			err := checkVarDecs(tc.pkg)
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
