use super::*;
use ast::Assignment::*;
use ast::Expression::Bad as BadE;
use ast::Expression::*;
use ast::FunctionBody::Block as FBlock;
use ast::FunctionBody::Expr as FExpr;
use ast::LogicalOperatorKind::*;
use ast::OperatorKind::*;
use ast::PropertyKey::Identifier as PkIdt;
use ast::PropertyKey::StringLiteral as PkStr;
use ast::Statement::Bad as BadS;
use ast::Statement::*;

use chrono::DateTime;

// this would give us a colorful diff.
#[cfg(test)]
use pretty_assertions::assert_eq;

#[test]
fn package_clause() {
    let mut p = Parser::new(r#"package foo"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: Some(PackageClause {
                base: BaseNode { errors: vec![] },
                name: Identifier {
                    base: BaseNode { errors: vec![] },
                    name: "foo".to_string()
                }
            }),
            imports: vec![],
            body: vec![]
        },
    )
}

#[test]
fn import() {
    let mut p = Parser::new(r#"import "path/foo""#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![ImportDeclaration {
                base: BaseNode { errors: vec![] },
                alias: None,
                path: StringLiteral {
                    base: BaseNode { errors: vec![] },
                    value: "path/foo".to_string()
                }
            }],
            body: vec![]
        },
    )
}

#[test]
fn import_as() {
    let mut p = Parser::new(r#"import bar "path/foo""#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![ImportDeclaration {
                base: BaseNode { errors: vec![] },
                alias: Some(Identifier {
                    base: BaseNode { errors: vec![] },
                    name: "bar".to_string()
                }),
                path: StringLiteral {
                    base: BaseNode { errors: vec![] },
                    value: "path/foo".to_string()
                }
            }],
            body: vec![]
        }
    )
}

#[test]
fn imports() {
    let mut p = Parser::new(
        r#"import "path/foo"
import "path/bar""#,
    );
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![
                ImportDeclaration {
                    base: BaseNode { errors: vec![] },
                    alias: None,
                    path: StringLiteral {
                        base: BaseNode { errors: vec![] },
                        value: "path/foo".to_string()
                    }
                },
                ImportDeclaration {
                    base: BaseNode { errors: vec![] },
                    alias: None,
                    path: StringLiteral {
                        base: BaseNode { errors: vec![] },
                        value: "path/bar".to_string()
                    }
                }
            ],
            body: vec![]
        }
    )
}

#[test]
fn package_and_imports() {
    let mut p = Parser::new(
        r#"
package baz

import "path/foo"
import "path/bar""#,
    );
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: Some(PackageClause {
                base: BaseNode { errors: vec![] },
                name: Identifier {
                    base: BaseNode { errors: vec![] },
                    name: "baz".to_string()
                }
            }),
            imports: vec![
                ImportDeclaration {
                    base: BaseNode { errors: vec![] },
                    alias: None,
                    path: StringLiteral {
                        base: BaseNode { errors: vec![] },
                        value: "path/foo".to_string()
                    }
                },
                ImportDeclaration {
                    base: BaseNode { errors: vec![] },
                    alias: None,
                    path: StringLiteral {
                        base: BaseNode { errors: vec![] },
                        value: "path/bar".to_string()
                    }
                }
            ],
            body: vec![]
        },
    )
}

#[test]
fn illegal_expression() {
    let mut p = Parser::new(r#"(@)"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode {
                    errors: vec!["invalid expression @position: @".to_string()]
                },
                expression: BadE(Box::new(BadExpression {
                    base: BaseNode { errors: vec![] },
                    text: "@".to_string(),
                    expression: None
                }))
            }),]
        },
    )
}

#[test]
fn package_and_imports_and_body() {
    let mut p = Parser::new(
        r#"
package baz

import "path/foo"
import "path/bar"

1 + 1"#,
    );
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: Some(PackageClause {
                base: BaseNode { errors: vec![] },
                name: Identifier {
                    base: BaseNode { errors: vec![] },
                    name: "baz".to_string()
                }
            }),
            imports: vec![
                ImportDeclaration {
                    base: BaseNode { errors: vec![] },
                    alias: None,
                    path: StringLiteral {
                        base: BaseNode { errors: vec![] },
                        value: "path/foo".to_string()
                    }
                },
                ImportDeclaration {
                    base: BaseNode { errors: vec![] },
                    alias: None,
                    path: StringLiteral {
                        base: BaseNode { errors: vec![] },
                        value: "path/bar".to_string()
                    }
                }
            ],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Bin(Box::new(BinaryExpression {
                    base: BaseNode { errors: vec![] },
                    operator: AdditionOperator,
                    left: Int(IntegerLiteral {
                        base: BaseNode { errors: vec![] },
                        value: 1
                    }),
                    right: Int(IntegerLiteral {
                        base: BaseNode { errors: vec![] },
                        value: 1
                    })
                }))
            })]
        },
    )
}

#[test]
fn package_and_imports_and_body_2() {
    let mut p = Parser::new(
        r#"
package baz

import "path/foo"
import "path/bar"

2 ^ 4"#,
    );
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: Some(PackageClause {
                base: BaseNode { errors: vec![] },
                name: Identifier {
                    base: BaseNode { errors: vec![] },
                    name: "baz".to_string()
                }
            }),
            imports: vec![
                ImportDeclaration {
                    base: BaseNode { errors: vec![] },
                    alias: None,
                    path: StringLiteral {
                        base: BaseNode { errors: vec![] },
                        value: "path/foo".to_string()
                    }
                },
                ImportDeclaration {
                    base: BaseNode { errors: vec![] },
                    alias: None,
                    path: StringLiteral {
                        base: BaseNode { errors: vec![] },
                        value: "path/bar".to_string()
                    }
                }
            ],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Bin(Box::new(BinaryExpression {
                    base: BaseNode { errors: vec![] },
                    operator: PowerOperator,
                    left: Int(IntegerLiteral {
                        base: BaseNode { errors: vec![] },
                        value: 2
                    }),
                    right: Int(IntegerLiteral {
                        base: BaseNode { errors: vec![] },
                        value: 4
                    })
                }))
            })]
        },
    )
}

#[test]
fn optional_query_metadata() {
    let mut p = Parser::new(
        r#"option task = {
				name: "foo",
				every: 1h,
				delay: 10m,
				cron: "0 2 * * *",
				retry: 5,
			  }"#,
    );
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Opt(OptionStatement {
                base: BaseNode { errors: vec![] },
                assignment: Variable(VariableAssignment {
                    base: BaseNode { errors: vec![] },
                    id: Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "task".to_string()
                    },
                    init: Obj(Box::new(ObjectExpression {
                        base: BaseNode { errors: vec![] },
                        with: None,
                        properties: vec![
                            Property {
                                base: BaseNode { errors: vec![] },
                                key: PkIdt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "name".to_string()
                                }),
                                value: Some(Str(StringLiteral {
                                    base: BaseNode { errors: vec![] },
                                    value: "foo".to_string()
                                }))
                            },
                            Property {
                                base: BaseNode { errors: vec![] },
                                key: PkIdt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "every".to_string()
                                }),
                                value: Some(Dur(DurationLiteral {
                                    base: BaseNode { errors: vec![] },
                                    values: vec![Duration {
                                        magnitude: 1,
                                        unit: "h".to_string()
                                    }]
                                }))
                            },
                            Property {
                                base: BaseNode { errors: vec![] },
                                key: PkIdt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "delay".to_string()
                                }),
                                value: Some(Dur(DurationLiteral {
                                    base: BaseNode { errors: vec![] },
                                    values: vec![Duration {
                                        magnitude: 10,
                                        unit: "m".to_string()
                                    }]
                                }))
                            },
                            Property {
                                base: BaseNode { errors: vec![] },
                                key: PkIdt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "cron".to_string()
                                }),
                                value: Some(Str(StringLiteral {
                                    base: BaseNode { errors: vec![] },
                                    value: "0 2 * * *".to_string()
                                }))
                            },
                            Property {
                                base: BaseNode { errors: vec![] },
                                key: PkIdt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "retry".to_string()
                                }),
                                value: Some(Int(IntegerLiteral {
                                    base: BaseNode { errors: vec![] },
                                    value: 5
                                }))
                            }
                        ]
                    }))
                })
            })]
        },
    )
}

#[test]
fn optional_query_metadata_preceding_query_text() {
    let mut p = Parser::new(
        r#"option task = {
            name: "foo",  // Name of task
            every: 1h,    // Execution frequency of task
        }

        // Task will execute the following query
        from() |> count()"#,
    );
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Opt(OptionStatement {
                    base: BaseNode { errors: vec![] },
                    assignment: Variable(VariableAssignment {
                        base: BaseNode { errors: vec![] },
                        id: Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "task".to_string()
                        },
                        init: Obj(Box::new(ObjectExpression {
                            base: BaseNode { errors: vec![] },
                            with: None,
                            properties: vec![
                                Property {
                                    base: BaseNode { errors: vec![] },
                                    key: PkIdt(Identifier {
                                        base: BaseNode { errors: vec![] },
                                        name: "name".to_string()
                                    }),
                                    value: Some(Str(StringLiteral {
                                        base: BaseNode { errors: vec![] },
                                        value: "foo".to_string()
                                    }))
                                },
                                Property {
                                    base: BaseNode { errors: vec![] },
                                    key: PkIdt(Identifier {
                                        base: BaseNode { errors: vec![] },
                                        name: "every".to_string()
                                    }),
                                    value: Some(Dur(DurationLiteral {
                                        base: BaseNode { errors: vec![] },
                                        values: vec![Duration {
                                            magnitude: 1,
                                            unit: "h".to_string()
                                        }]
                                    }))
                                }
                            ]
                        }))
                    })
                }),
                Expr(ExpressionStatement {
                    base: BaseNode { errors: vec![] },
                    expression: Pipe(Box::new(PipeExpression {
                        base: BaseNode { errors: vec![] },
                        argument: Call(Box::new(CallExpression {
                            base: BaseNode { errors: vec![] },
                            callee: Idt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "from".to_string()
                            }),
                            arguments: vec![]
                        })),
                        call: CallExpression {
                            base: BaseNode { errors: vec![] },
                            callee: Idt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "count".to_string()
                            }),
                            arguments: vec![]
                        }
                    })),
                })
            ]
        },
    )
}

#[test]
fn qualified_option() {
    let mut p = Parser::new(r#"option alert.state = "Warning""#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Opt(OptionStatement {
                base: BaseNode { errors: vec![] },
                assignment: Member(MemberAssignment {
                    base: BaseNode { errors: vec![] },
                    member: MemberExpression {
                        base: BaseNode { errors: vec![] },
                        object: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "alert".to_string()
                        }),
                        property: PkIdt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "state".to_string()
                        })
                    },
                    init: Str(StringLiteral {
                        base: BaseNode { errors: vec![] },
                        value: "Warning".to_string()
                    })
                })
            })]
        },
    )
}

#[test]
fn builtin() {
    let mut p = Parser::new(r#"builtin from"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Built(BuiltinStatement {
                base: BaseNode { errors: vec![] },
                id: Identifier {
                    base: BaseNode { errors: vec![] },
                    name: "from".to_string()
                }
            })]
        },
    )
}

#[test]
fn test_statement() {
    let mut p = Parser::new(r#"test mean = {want: 0, got: 0}"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Test(TestStatement {
                base: BaseNode { errors: vec![] },
                assignment: VariableAssignment {
                    base: BaseNode { errors: vec![] },
                    id: Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "mean".to_string()
                    },
                    init: Obj(Box::new(ObjectExpression {
                        base: BaseNode { errors: vec![] },
                        with: None,
                        properties: vec![
                            Property {
                                base: BaseNode { errors: vec![] },
                                key: PkIdt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "want".to_string()
                                }),
                                value: Some(Int(IntegerLiteral {
                                    base: BaseNode { errors: vec![] },
                                    value: 0
                                }))
                            },
                            Property {
                                base: BaseNode { errors: vec![] },
                                key: PkIdt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "got".to_string()
                                }),
                                value: Some(Int(IntegerLiteral {
                                    base: BaseNode { errors: vec![] },
                                    value: 0
                                }))
                            }
                        ]
                    }))
                }
            })]
        },
    )
}

#[test]
fn from() {
    let mut p = Parser::new(r#"from()"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Call(Box::new(CallExpression {
                    base: BaseNode { errors: vec![] },
                    arguments: vec![],
                    callee: Idt(Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "from".to_string()
                    })
                })),
            })]
        },
    )
}

#[test]
fn comment() {
    let mut p = Parser::new(
        r#"// Comment
			from()"#,
    );
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Call(Box::new(CallExpression {
                    base: BaseNode { errors: vec![] },
                    arguments: vec![],
                    callee: Idt(Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "from".to_string()
                    })
                })),
            })]
        },
    )
}

#[test]
fn identifier_with_number() {
    let mut p = Parser::new(r#"tan2()"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Call(Box::new(CallExpression {
                    base: BaseNode { errors: vec![] },
                    arguments: vec![],
                    callee: Idt(Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "tan2".to_string()
                    })
                }))
            })]
        },
    )
}

#[test]
fn regex_literal() {
    let mut p = Parser::new(r#"/.*/"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Regexp(RegexpLiteral {
                    base: BaseNode { errors: vec![] },
                    value: ".*".to_string()
                })
            })]
        },
    )
}

#[test]
fn regex_literal_with_escape_sequence() {
    let mut p = Parser::new(r#"/a\/b\\c\d/"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Regexp(RegexpLiteral {
                    base: BaseNode { errors: vec![] },
                    value: "a/b\\\\c\\d".to_string()
                })
            })]
        },
    )
}

#[test]
fn regex_match_operators() {
    let mut p = Parser::new(r#""a" =~ /.*/ and "b" !~ /c$/"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Log(Box::new(LogicalExpression {
                    base: BaseNode { errors: vec![] },
                    operator: AndOperator,
                    left: Bin(Box::new(BinaryExpression {
                        base: BaseNode { errors: vec![] },
                        operator: RegexpMatchOperator,
                        left: Str(StringLiteral {
                            base: BaseNode { errors: vec![] },
                            value: "a".to_string()
                        }),
                        right: Regexp(RegexpLiteral {
                            base: BaseNode { errors: vec![] },
                            value: ".*".to_string()
                        })
                    })),
                    right: Bin(Box::new(BinaryExpression {
                        base: BaseNode { errors: vec![] },
                        operator: NotRegexpMatchOperator,
                        left: Str(StringLiteral {
                            base: BaseNode { errors: vec![] },
                            value: "b".to_string()
                        }),
                        right: Regexp(RegexpLiteral {
                            base: BaseNode { errors: vec![] },
                            value: "c$".to_string()
                        })
                    }))
                }))
            })]
        },
    )
}

#[test]
fn declare_variable_as_an_int() {
    let mut p = Parser::new(r#"howdy = 1"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Var(VariableAssignment {
                base: BaseNode { errors: vec![] },
                id: Identifier {
                    base: BaseNode { errors: vec![] },
                    name: "howdy".to_string()
                },
                init: Int(IntegerLiteral {
                    base: BaseNode { errors: vec![] },
                    value: 1
                })
            })]
        },
    )
}

#[test]
fn declare_variable_as_a_float() {
    let mut p = Parser::new(r#"howdy = 1.1"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Var(VariableAssignment {
                base: BaseNode { errors: vec![] },
                id: Identifier {
                    base: BaseNode { errors: vec![] },
                    name: "howdy".to_string()
                },
                init: Flt(FloatLiteral {
                    base: BaseNode { errors: vec![] },
                    value: 1.1
                })
            })]
        },
    )
}

#[test]
fn declare_variable_as_an_array() {
    let mut p = Parser::new(r#"howdy = [1, 2, 3, 4]"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Var(VariableAssignment {
                base: BaseNode { errors: vec![] },
                id: Identifier {
                    base: BaseNode { errors: vec![] },
                    name: "howdy".to_string()
                },
                init: Arr(Box::new(ArrayExpression {
                    base: BaseNode { errors: vec![] },
                    elements: vec![
                        Int(IntegerLiteral {
                            base: BaseNode { errors: vec![] },
                            value: 1
                        }),
                        Int(IntegerLiteral {
                            base: BaseNode { errors: vec![] },
                            value: 2
                        }),
                        Int(IntegerLiteral {
                            base: BaseNode { errors: vec![] },
                            value: 3
                        }),
                        Int(IntegerLiteral {
                            base: BaseNode { errors: vec![] },
                            value: 4
                        })
                    ]
                }))
            })]
        },
    )
}

#[test]
fn declare_variable_as_an_empty_array() {
    let mut p = Parser::new(r#"howdy = []"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Var(VariableAssignment {
                base: BaseNode { errors: vec![] },
                id: Identifier {
                    base: BaseNode { errors: vec![] },
                    name: "howdy".to_string()
                },
                init: Arr(Box::new(ArrayExpression {
                    base: BaseNode { errors: vec![] },
                    elements: vec![],
                }))
            })]
        },
    )
}

#[test]
fn use_variable_to_declare_something() {
    let mut p = Parser::new(
        r#"howdy = 1
			from()"#,
    );
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Var(VariableAssignment {
                    base: BaseNode { errors: vec![] },
                    id: Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "howdy".to_string()
                    },
                    init: Int(IntegerLiteral {
                        base: BaseNode { errors: vec![] },
                        value: 1
                    })
                }),
                Expr(ExpressionStatement {
                    base: BaseNode { errors: vec![] },
                    expression: Call(Box::new(CallExpression {
                        base: BaseNode { errors: vec![] },
                        arguments: vec![],
                        callee: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "from".to_string()
                        })
                    })),
                })
            ]
        },
    )
}

#[test]
fn variable_is_from_statement() {
    let mut p = Parser::new(
        r#"howdy = from()
			howdy.count()"#,
    );
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Var(VariableAssignment {
                    base: BaseNode { errors: vec![] },
                    id: Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "howdy".to_string()
                    },
                    init: Call(Box::new(CallExpression {
                        base: BaseNode { errors: vec![] },
                        arguments: vec![],
                        callee: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "from".to_string()
                        })
                    })),
                }),
                Expr(ExpressionStatement {
                    base: BaseNode { errors: vec![] },
                    expression: Call(Box::new(CallExpression {
                        base: BaseNode { errors: vec![] },
                        arguments: vec![],
                        callee: Mem(Box::new(MemberExpression {
                            base: BaseNode { errors: vec![] },
                            object: Idt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "howdy".to_string()
                            }),
                            property: PkIdt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "count".to_string()
                            })
                        }))
                    }))
                })
            ]
        },
    )
}

#[test]
fn pipe_expression() {
    let mut p = Parser::new(r#"from() |> count()"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Pipe(Box::new(PipeExpression {
                    base: BaseNode { errors: vec![] },
                    argument: Call(Box::new(CallExpression {
                        base: BaseNode { errors: vec![] },
                        callee: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "from".to_string()
                        }),
                        arguments: vec![]
                    })),
                    call: CallExpression {
                        base: BaseNode { errors: vec![] },
                        callee: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "count".to_string()
                        }),
                        arguments: vec![]
                    }
                }))
            })]
        },
    )
}

#[test]
fn pipe_expression_to_member_expression_function() {
    let mut p = Parser::new(r#"a |> b.c(d:e)"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Pipe(Box::new(PipeExpression {
                    base: BaseNode { errors: vec![] },
                    argument: Idt(Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "a".to_string()
                    }),
                    call: CallExpression {
                        base: BaseNode { errors: vec![] },
                        callee: Mem(Box::new(MemberExpression {
                            base: BaseNode { errors: vec![] },
                            object: Idt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "b".to_string()
                            }),
                            property: PkIdt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "c".to_string()
                            })
                        })),
                        arguments: vec![Obj(Box::new(ObjectExpression {
                            base: BaseNode { errors: vec![] },
                            with: None,
                            properties: vec![Property {
                                base: BaseNode { errors: vec![] },
                                key: PkIdt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "d".to_string()
                                }),
                                value: Some(Idt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "e".to_string()
                                }))
                            }]
                        }))]
                    }
                }))
            })]
        },
    )
}

#[test]
fn literal_pipe_expression() {
    let mut p = Parser::new(r#"5 |> pow2()"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Pipe(Box::new(PipeExpression {
                    base: BaseNode { errors: vec![] },
                    argument: Int(IntegerLiteral {
                        base: BaseNode { errors: vec![] },
                        value: 5,
                    }),
                    call: CallExpression {
                        base: BaseNode { errors: vec![] },
                        arguments: vec![],
                        callee: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "pow2".to_string()
                        })
                    },
                })),
            })]
        },
    )
}

#[test]
fn member_expression_pipe_expression() {
    let mut p = Parser::new(r#"foo.bar |> baz()"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Pipe(Box::new(PipeExpression {
                    base: BaseNode { errors: vec![] },
                    argument: Mem(Box::new(MemberExpression {
                        base: BaseNode { errors: vec![] },
                        object: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "foo".to_string()
                        }),
                        property: PkIdt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "bar".to_string()
                        })
                    })),
                    call: CallExpression {
                        base: BaseNode { errors: vec![] },
                        callee: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "baz".to_string()
                        }),
                        arguments: vec![]
                    }
                }))
            })]
        },
    )
}

#[test]
fn multiple_pipe_expressions() {
    let mut p = Parser::new(r#"from() |> range() |> filter() |> count()"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Pipe(Box::new(PipeExpression {
                    base: BaseNode { errors: vec![] },
                    argument: Pipe(Box::new(PipeExpression {
                        base: BaseNode { errors: vec![] },
                        argument: Pipe(Box::new(PipeExpression {
                            base: BaseNode { errors: vec![] },
                            argument: Call(Box::new(CallExpression {
                                base: BaseNode { errors: vec![] },
                                callee: Idt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "from".to_string()
                                }),
                                arguments: vec![]
                            })),
                            call: CallExpression {
                                base: BaseNode { errors: vec![] },
                                callee: Idt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "range".to_string()
                                }),
                                arguments: vec![]
                            }
                        })),
                        call: CallExpression {
                            base: BaseNode { errors: vec![] },
                            callee: Idt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "filter".to_string()
                            }),
                            arguments: vec![]
                        }
                    })),
                    call: CallExpression {
                        base: BaseNode { errors: vec![] },
                        callee: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "count".to_string()
                        }),
                        arguments: vec![]
                    }
                }))
            })]
        },
    )
}

#[test]
fn pipe_expression_into_non_call_expression() {
    let mut p = Parser::new(r#"foo() |> bar"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Pipe(Box::new(PipeExpression {
                    base: BaseNode { errors: vec![] },
                    argument: Call(Box::new(CallExpression {
                        base: BaseNode { errors: vec![] },
                        callee: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "foo".to_string()
                        }),
                        arguments: vec![]
                    })),
                    call: CallExpression {
                        base: BaseNode {
                            errors: vec!["pipe destination must be a function call".to_string()]
                        },
                        callee: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "bar".to_string()
                        }),
                        arguments: vec![]
                    }
                })),
            })]
        },
    )
}

#[test]
fn two_variables_for_two_froms() {
    let mut p = Parser::new(
        r#"howdy = from()
			doody = from()
			howdy|>count()
			doody|>sum()"#,
    );
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Var(VariableAssignment {
                    base: BaseNode { errors: vec![] },
                    id: Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "howdy".to_string()
                    },
                    init: Call(Box::new(CallExpression {
                        base: BaseNode { errors: vec![] },
                        arguments: vec![],
                        callee: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "from".to_string()
                        })
                    })),
                }),
                Var(VariableAssignment {
                    base: BaseNode { errors: vec![] },
                    id: Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "doody".to_string()
                    },
                    init: Call(Box::new(CallExpression {
                        base: BaseNode { errors: vec![] },
                        arguments: vec![],
                        callee: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "from".to_string()
                        })
                    })),
                }),
                Expr(ExpressionStatement {
                    base: BaseNode { errors: vec![] },
                    expression: Pipe(Box::new(PipeExpression {
                        base: BaseNode { errors: vec![] },
                        argument: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "howdy".to_string()
                        }),
                        call: CallExpression {
                            base: BaseNode { errors: vec![] },
                            callee: Idt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "count".to_string()
                            }),
                            arguments: vec![]
                        }
                    }))
                }),
                Expr(ExpressionStatement {
                    base: BaseNode { errors: vec![] },
                    expression: Pipe(Box::new(PipeExpression {
                        base: BaseNode { errors: vec![] },
                        argument: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "doody".to_string()
                        }),
                        call: CallExpression {
                            base: BaseNode { errors: vec![] },
                            callee: Idt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "sum".to_string()
                            }),
                            arguments: vec![]
                        }
                    }))
                })
            ]
        },
    )
}

#[test]
fn from_with_database() {
    let mut p = Parser::new(r#"from(bucket:"telegraf/autogen")"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Call(Box::new(CallExpression {
                    base: BaseNode { errors: vec![] },
                    callee: Idt(Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "from".to_string()
                    }),
                    arguments: vec![Obj(Box::new(ObjectExpression {
                        base: BaseNode { errors: vec![] },
                        with: None,
                        properties: vec![Property {
                            base: BaseNode { errors: vec![] },
                            key: PkIdt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "bucket".to_string()
                            }),
                            value: Some(Str(StringLiteral {
                                base: BaseNode { errors: vec![] },
                                value: "telegraf/autogen".to_string()
                            }))
                        }]
                    }))]
                }))
            })]
        },
    )
}

#[test]
fn map_member_expressions() {
    let mut p = Parser::new(
        r#"m = {key1: 1, key2:"value2"}
			m.key1
			m["key2"]
			"#,
    );
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Var(VariableAssignment {
                    base: BaseNode { errors: vec![] },
                    id: Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "m".to_string()
                    },
                    init: Obj(Box::new(ObjectExpression {
                        base: BaseNode { errors: vec![] },
                        with: None,
                        properties: vec![
                            Property {
                                base: BaseNode { errors: vec![] },
                                key: PkIdt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "key1".to_string()
                                }),
                                value: Some(Int(IntegerLiteral {
                                    base: BaseNode { errors: vec![] },
                                    value: 1
                                }))
                            },
                            Property {
                                base: BaseNode { errors: vec![] },
                                key: PkIdt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "key2".to_string()
                                }),
                                value: Some(Str(StringLiteral {
                                    base: BaseNode { errors: vec![] },
                                    value: "value2".to_string()
                                }))
                            }
                        ]
                    }))
                }),
                Expr(ExpressionStatement {
                    base: BaseNode { errors: vec![] },
                    expression: Mem(Box::new(MemberExpression {
                        base: BaseNode { errors: vec![] },
                        object: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "m".to_string()
                        }),
                        property: PkIdt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "key1".to_string()
                        })
                    }))
                }),
                Expr(ExpressionStatement {
                    base: BaseNode { errors: vec![] },
                    expression: Mem(Box::new(MemberExpression {
                        base: BaseNode { errors: vec![] },
                        object: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "m".to_string()
                        }),
                        property: PkStr(StringLiteral {
                            base: BaseNode { errors: vec![] },
                            value: "key2".to_string()
                        })
                    }))
                })
            ]
        },
    )
}

#[test]
fn object_with_string_literal_key() {
    let mut p = Parser::new(r#"x = {"a": 10}"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Var(VariableAssignment {
                base: BaseNode { errors: vec![] },
                id: Identifier {
                    base: BaseNode { errors: vec![] },
                    name: "x".to_string()
                },
                init: Obj(Box::new(ObjectExpression {
                    base: BaseNode { errors: vec![] },
                    with: None,
                    properties: vec![Property {
                        base: BaseNode { errors: vec![] },
                        key: PkStr(StringLiteral {
                            base: BaseNode { errors: vec![] },
                            value: "a".to_string()
                        }),
                        value: Some(Int(IntegerLiteral {
                            base: BaseNode { errors: vec![] },
                            value: 10
                        }))
                    }]
                }))
            })]
        },
    )
}

#[test]
fn object_with_mixed_keys() {
    let mut p = Parser::new(r#"x = {"a": 10, b: 11}"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Var(VariableAssignment {
                base: BaseNode { errors: vec![] },
                id: Identifier {
                    base: BaseNode { errors: vec![] },
                    name: "x".to_string()
                },
                init: Obj(Box::new(ObjectExpression {
                    base: BaseNode { errors: vec![] },
                    with: None,
                    properties: vec![
                        Property {
                            base: BaseNode { errors: vec![] },
                            key: PkStr(StringLiteral {
                                base: BaseNode { errors: vec![] },
                                value: "a".to_string()
                            }),
                            value: Some(Int(IntegerLiteral {
                                base: BaseNode { errors: vec![] },
                                value: 10
                            }))
                        },
                        Property {
                            base: BaseNode { errors: vec![] },
                            key: PkIdt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "b".to_string()
                            }),
                            value: Some(Int(IntegerLiteral {
                                base: BaseNode { errors: vec![] },
                                value: 11
                            }))
                        }
                    ]
                }))
            })]
        },
    )
}

#[test]
fn implicit_key_object_literal() {
    let mut p = Parser::new(r#"x = {a, b}"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Var(VariableAssignment {
                base: BaseNode { errors: vec![] },
                id: Identifier {
                    base: BaseNode { errors: vec![] },
                    name: "x".to_string()
                },
                init: Obj(Box::new(ObjectExpression {
                    base: BaseNode { errors: vec![] },
                    with: None,
                    properties: vec![
                        Property {
                            base: BaseNode { errors: vec![] },
                            key: PkIdt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "a".to_string()
                            }),
                            value: None
                        },
                        Property {
                            base: BaseNode { errors: vec![] },
                            key: PkIdt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "b".to_string()
                            }),
                            value: None
                        }
                    ]
                }))
            })]
        },
    )
}

#[test]
fn index_expression() {
    let mut p = Parser::new(r#"a[3]"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Idx(Box::new(IndexExpression {
                    base: BaseNode { errors: vec![] },
                    array: Idt(Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "a".to_string()
                    }),
                    index: Int(IntegerLiteral {
                        base: BaseNode { errors: vec![] },
                        value: 3
                    })
                }))
            })]
        },
    )
}

#[test]
fn nested_index_expression() {
    let mut p = Parser::new(r#"a[3][5]"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Idx(Box::new(IndexExpression {
                    base: BaseNode { errors: vec![] },
                    array: Idx(Box::new(IndexExpression {
                        base: BaseNode { errors: vec![] },
                        array: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "a".to_string()
                        }),
                        index: Int(IntegerLiteral {
                            base: BaseNode { errors: vec![] },
                            value: 3
                        })
                    })),
                    index: Int(IntegerLiteral {
                        base: BaseNode { errors: vec![] },
                        value: 5
                    })
                }))
            })]
        },
    )
}

#[test]
fn access_indexed_object_returned_from_function_call() {
    let mut p = Parser::new(r#"f()[3]"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Idx(Box::new(IndexExpression {
                    base: BaseNode { errors: vec![] },
                    array: Call(Box::new(CallExpression {
                        base: BaseNode { errors: vec![] },
                        arguments: vec![],
                        callee: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "f".to_string()
                        }),
                    })),
                    index: Int(IntegerLiteral {
                        base: BaseNode { errors: vec![] },
                        value: 3
                    })
                }))
            })]
        },
    )
}

#[test]
fn index_with_member_expressions() {
    let mut p = Parser::new(r#"a.b["c"]"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Mem(Box::new(MemberExpression {
                    base: BaseNode { errors: vec![] },
                    object: Mem(Box::new(MemberExpression {
                        base: BaseNode { errors: vec![] },
                        object: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "a".to_string()
                        }),
                        property: PkIdt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "b".to_string()
                        })
                    })),
                    property: PkStr(StringLiteral {
                        base: BaseNode { errors: vec![] },
                        value: "c".to_string()
                    })
                }))
            })]
        },
    )
}

#[test]
fn index_with_member_with_call_expression() {
    let mut p = Parser::new(r#"a.b()["c"]"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Mem(Box::new(MemberExpression {
                    base: BaseNode { errors: vec![] },
                    object: Call(Box::new(CallExpression {
                        base: BaseNode { errors: vec![] },
                        arguments: vec![],
                        callee: Mem(Box::new(MemberExpression {
                            base: BaseNode { errors: vec![] },
                            object: Idt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "a".to_string()
                            }),
                            property: PkIdt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "b".to_string()
                            })
                        })),
                    })),
                    property: PkStr(StringLiteral {
                        base: BaseNode { errors: vec![] },
                        value: "c".to_string()
                    })
                }))
            })]
        },
    )
}

#[test]
fn index_with_unclosed_bracket() {
    let mut p = Parser::new(r#"a[b()"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Idx(Box::new(IndexExpression {
                    base: BaseNode {
                        errors: vec!["expected RBRACK, got EOF".to_string()]
                    },
                    array: Idt(Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "a".to_string()
                    }),
                    index: Call(Box::new(CallExpression {
                        base: BaseNode { errors: vec![] },
                        arguments: vec![],
                        callee: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "b".to_string()
                        })
                    })),
                }))
            })]
        },
    )
}

#[test]
fn index_with_unbalanced_parenthesis() {
    let mut p = Parser::new(r#"a[b(]"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Idx(Box::new(IndexExpression {
                    base: BaseNode { errors: vec![] },
                    array: Idt(Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "a".to_string()
                    }),
                    index: Call(Box::new(CallExpression {
                        base: BaseNode {
                            errors: vec!["expected RPAREN, got RBRACK".to_string()]
                        },
                        arguments: vec![],
                        callee: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "b".to_string()
                        })
                    })),
                }))
            })]
        },
    )
}

#[test]
fn index_with_unexpected_rparen() {
    let mut p = Parser::new(r#"a[b)]"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Idx(Box::new(IndexExpression {
                    base: BaseNode {
                        errors: vec!["invalid expression @position: )".to_string()]
                    },
                    array: Idt(Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "a".to_string()
                    }),
                    index: Idt(Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "b".to_string()
                    })
                }))
            })]
        },
    )
}

#[test]
fn binary_expression() {
    let mut p = Parser::new(r#"_value < 10.0"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Bin(Box::new(BinaryExpression {
                    base: BaseNode { errors: vec![] },
                    operator: LessThanOperator,
                    left: Idt(Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "_value".to_string()
                    }),
                    right: Flt(FloatLiteral {
                        base: BaseNode { errors: vec![] },
                        value: 10.0
                    })
                }))
            })]
        },
    )
}

#[test]
fn member_expression_binary_expression() {
    let mut p = Parser::new(r#"r._value < 10.0"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Bin(Box::new(BinaryExpression {
                    base: BaseNode { errors: vec![] },
                    operator: LessThanOperator,
                    left: Mem(Box::new(MemberExpression {
                        base: BaseNode { errors: vec![] },
                        object: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "r".to_string()
                        }),
                        property: PkIdt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "_value".to_string()
                        })
                    })),
                    right: Flt(FloatLiteral {
                        base: BaseNode { errors: vec![] },
                        value: 10.0
                    })
                }))
            })]
        },
    )
}

#[test]
fn var_as_binary_expression_of_other_vars() {
    let mut p = Parser::new(
        r#"a = 1
            b = 2
            c = a + b
            d = a"#,
    );
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Var(VariableAssignment {
                    base: BaseNode { errors: vec![] },
                    id: Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "a".to_string()
                    },
                    init: Int(IntegerLiteral {
                        base: BaseNode { errors: vec![] },
                        value: 1
                    })
                }),
                Var(VariableAssignment {
                    base: BaseNode { errors: vec![] },
                    id: Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "b".to_string()
                    },
                    init: Int(IntegerLiteral {
                        base: BaseNode { errors: vec![] },
                        value: 2
                    })
                }),
                Var(VariableAssignment {
                    base: BaseNode { errors: vec![] },
                    id: Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "c".to_string()
                    },
                    init: Bin(Box::new(BinaryExpression {
                        base: BaseNode { errors: vec![] },
                        operator: AdditionOperator,
                        left: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "a".to_string()
                        }),
                        right: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "b".to_string()
                        })
                    }))
                }),
                Var(VariableAssignment {
                    base: BaseNode { errors: vec![] },
                    id: Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "d".to_string()
                    },
                    init: Idt(Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "a".to_string()
                    })
                })
            ]
        },
    )
}

#[test]
fn var_as_unary_expression_of_other_vars() {
    let mut p = Parser::new(
        r#"a = 5
            c = -a"#,
    );
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Var(VariableAssignment {
                    base: BaseNode { errors: vec![] },
                    id: Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "a".to_string()
                    },
                    init: Int(IntegerLiteral {
                        base: BaseNode { errors: vec![] },
                        value: 5
                    })
                }),
                Var(VariableAssignment {
                    base: BaseNode { errors: vec![] },
                    id: Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "c".to_string()
                    },
                    init: Un(Box::new(UnaryExpression {
                        base: BaseNode { errors: vec![] },
                        operator: SubtractionOperator,
                        argument: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "a".to_string()
                        })
                    }))
                })
            ]
        },
    )
}

#[test]
fn var_as_both_binary_and_unary_expressions() {
    let mut p = Parser::new(
        r#"a = 5
            c = 10 * -a"#,
    );
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Var(VariableAssignment {
                    base: BaseNode { errors: vec![] },
                    id: Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "a".to_string()
                    },
                    init: Int(IntegerLiteral {
                        base: BaseNode { errors: vec![] },
                        value: 5
                    })
                }),
                Var(VariableAssignment {
                    base: BaseNode { errors: vec![] },
                    id: Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "c".to_string()
                    },
                    init: Bin(Box::new(BinaryExpression {
                        base: BaseNode { errors: vec![] },
                        operator: MultiplicationOperator,
                        left: Int(IntegerLiteral {
                            base: BaseNode { errors: vec![] },
                            value: 10
                        }),
                        right: Un(Box::new(UnaryExpression {
                            base: BaseNode { errors: vec![] },
                            operator: SubtractionOperator,
                            argument: Idt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "a".to_string()
                            })
                        }))
                    }))
                })
            ]
        },
    )
}

#[test]
fn unary_expressions_within_logical_expression() {
    let mut p = Parser::new(
        r#"a = 5.0
            10.0 * -a == -0.5 or a == 6.0"#,
    );
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Var(VariableAssignment {
                    base: BaseNode { errors: vec![] },
                    id: Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "a".to_string()
                    },
                    init: Flt(FloatLiteral {
                        base: BaseNode { errors: vec![] },
                        value: 5.0
                    })
                }),
                Expr(ExpressionStatement {
                    base: BaseNode { errors: vec![] },
                    expression: Log(Box::new(LogicalExpression {
                        base: BaseNode { errors: vec![] },
                        operator: OrOperator,
                        left: Bin(Box::new(BinaryExpression {
                            base: BaseNode { errors: vec![] },
                            operator: EqualOperator,
                            left: Bin(Box::new(BinaryExpression {
                                base: BaseNode { errors: vec![] },
                                operator: MultiplicationOperator,
                                left: Flt(FloatLiteral {
                                    base: BaseNode { errors: vec![] },
                                    value: 10.0
                                }),
                                right: Un(Box::new(UnaryExpression {
                                    base: BaseNode { errors: vec![] },
                                    operator: SubtractionOperator,
                                    argument: Idt(Identifier {
                                        base: BaseNode { errors: vec![] },
                                        name: "a".to_string()
                                    })
                                }))
                            })),
                            right: Un(Box::new(UnaryExpression {
                                base: BaseNode { errors: vec![] },
                                operator: SubtractionOperator,
                                argument: Flt(FloatLiteral {
                                    base: BaseNode { errors: vec![] },
                                    value: 0.5
                                })
                            }))
                        })),
                        right: Bin(Box::new(BinaryExpression {
                            base: BaseNode { errors: vec![] },
                            operator: EqualOperator,
                            left: Idt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "a".to_string()
                            }),
                            right: Flt(FloatLiteral {
                                base: BaseNode { errors: vec![] },
                                value: 6.0
                            })
                        }))
                    }))
                })
            ]
        },
    )
}

#[test]
fn unary_expressions_with_too_many_comments() {
    let mut p = Parser::new(
        r#"// define a
a = 5.0
// eval this
10.0 * -a == -0.5
	// or this
	or a == 6.0"#,
    );
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Var(VariableAssignment {
                    base: BaseNode { errors: vec![] },
                    id: Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "a".to_string()
                    },
                    init: Flt(FloatLiteral {
                        base: BaseNode { errors: vec![] },
                        value: 5.0
                    })
                }),
                Expr(ExpressionStatement {
                    base: BaseNode { errors: vec![] },
                    expression: Log(Box::new(LogicalExpression {
                        base: BaseNode { errors: vec![] },
                        operator: OrOperator,
                        left: Bin(Box::new(BinaryExpression {
                            base: BaseNode { errors: vec![] },
                            operator: EqualOperator,
                            left: Bin(Box::new(BinaryExpression {
                                base: BaseNode { errors: vec![] },
                                operator: MultiplicationOperator,
                                left: Flt(FloatLiteral {
                                    base: BaseNode { errors: vec![] },
                                    value: 10.0
                                }),
                                right: Un(Box::new(UnaryExpression {
                                    base: BaseNode { errors: vec![] },
                                    operator: SubtractionOperator,
                                    argument: Idt(Identifier {
                                        base: BaseNode { errors: vec![] },
                                        name: "a".to_string()
                                    })
                                }))
                            })),
                            right: Un(Box::new(UnaryExpression {
                                base: BaseNode { errors: vec![] },
                                operator: SubtractionOperator,
                                argument: Flt(FloatLiteral {
                                    base: BaseNode { errors: vec![] },
                                    value: 0.5
                                })
                            }))
                        })),
                        right: Bin(Box::new(BinaryExpression {
                            base: BaseNode { errors: vec![] },
                            operator: EqualOperator,
                            left: Idt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "a".to_string()
                            }),
                            right: Flt(FloatLiteral {
                                base: BaseNode { errors: vec![] },
                                value: 6.0
                            })
                        }))
                    }))
                })
            ]
        },
    )
}

#[test]
fn expressions_with_function_calls() {
    let mut p = Parser::new(r#"a = foo() == 10"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Var(VariableAssignment {
                base: BaseNode { errors: vec![] },
                id: Identifier {
                    base: BaseNode { errors: vec![] },
                    name: "a".to_string()
                },
                init: Bin(Box::new(BinaryExpression {
                    base: BaseNode { errors: vec![] },
                    operator: EqualOperator,
                    left: Call(Box::new(CallExpression {
                        base: BaseNode { errors: vec![] },
                        arguments: vec![],
                        callee: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "foo".to_string()
                        })
                    })),
                    right: Int(IntegerLiteral {
                        base: BaseNode { errors: vec![] },
                        value: 10
                    })
                }))
            })]
        },
    )
}

#[test]
fn mix_unary_logical_and_binary_expressions() {
    let mut p = Parser::new(
        r#"
            not (f() == 6.0 * x) or fail()"#,
    );
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Log(Box::new(LogicalExpression {
                    base: BaseNode { errors: vec![] },
                    operator: OrOperator,
                    left: Un(Box::new(UnaryExpression {
                        base: BaseNode { errors: vec![] },
                        operator: NotOperator,
                        argument: Bin(Box::new(BinaryExpression {
                            base: BaseNode { errors: vec![] },
                            operator: EqualOperator,
                            left: Call(Box::new(CallExpression {
                                base: BaseNode { errors: vec![] },
                                arguments: vec![],
                                callee: Idt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "f".to_string()
                                })
                            })),
                            right: Bin(Box::new(BinaryExpression {
                                base: BaseNode { errors: vec![] },
                                operator: MultiplicationOperator,
                                left: Flt(FloatLiteral {
                                    base: BaseNode { errors: vec![] },
                                    value: 6.0
                                }),
                                right: Idt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "x".to_string()
                                })
                            }))
                        }))
                    })),
                    right: Call(Box::new(CallExpression {
                        base: BaseNode { errors: vec![] },
                        arguments: vec![],
                        callee: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "fail".to_string()
                        })
                    })),
                }))
            })]
        },
    )
}

#[test]
fn mix_unary_logical_and_binary_expressions_with_extra_parens() {
    let mut p = Parser::new(
        r#"
            (not (f() == 6.0 * x) or fail())"#,
    );
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Log(Box::new(LogicalExpression {
                    base: BaseNode { errors: vec![] },
                    operator: OrOperator,
                    left: Un(Box::new(UnaryExpression {
                        base: BaseNode { errors: vec![] },
                        operator: NotOperator,
                        argument: Bin(Box::new(BinaryExpression {
                            base: BaseNode { errors: vec![] },
                            operator: EqualOperator,
                            left: Call(Box::new(CallExpression {
                                base: BaseNode { errors: vec![] },
                                arguments: vec![],
                                callee: Idt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "f".to_string()
                                })
                            })),
                            right: Bin(Box::new(BinaryExpression {
                                base: BaseNode { errors: vec![] },
                                operator: MultiplicationOperator,
                                left: Flt(FloatLiteral {
                                    base: BaseNode { errors: vec![] },
                                    value: 6.0
                                }),
                                right: Idt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "x".to_string()
                                })
                            }))
                        }))
                    })),
                    right: Call(Box::new(CallExpression {
                        base: BaseNode { errors: vec![] },
                        arguments: vec![],
                        callee: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "fail".to_string()
                        })
                    })),
                }))
            })]
        },
    )
}

#[test]
fn binary_operator_precedence() {
    let mut p = Parser::new(r#"a / b - 1.0"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Bin(Box::new(BinaryExpression {
                    base: BaseNode { errors: vec![] },
                    operator: SubtractionOperator,
                    left: Bin(Box::new(BinaryExpression {
                        base: BaseNode { errors: vec![] },
                        operator: DivisionOperator,
                        left: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "a".to_string()
                        }),
                        right: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "b".to_string()
                        })
                    })),
                    right: Flt(FloatLiteral {
                        base: BaseNode { errors: vec![] },
                        value: 1.0
                    })
                }))
            })]
        },
    )
}

#[test]
fn binary_operator_precedence_literals_only() {
    let mut p = Parser::new(r#"2 / "a" - 1.0"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Bin(Box::new(BinaryExpression {
                    base: BaseNode { errors: vec![] },
                    operator: SubtractionOperator,
                    left: Bin(Box::new(BinaryExpression {
                        base: BaseNode { errors: vec![] },
                        operator: DivisionOperator,
                        left: Int(IntegerLiteral {
                            base: BaseNode { errors: vec![] },
                            value: 2
                        }),
                        right: Str(StringLiteral {
                            base: BaseNode { errors: vec![] },
                            value: "a".to_string()
                        })
                    })),
                    right: Flt(FloatLiteral {
                        base: BaseNode { errors: vec![] },
                        value: 1.0
                    })
                }))
            })]
        },
    )
}

#[test]
fn binary_operator_precedence_double_subtraction() {
    let mut p = Parser::new(r#"1 - 2 - 3"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Bin(Box::new(BinaryExpression {
                    base: BaseNode { errors: vec![] },
                    operator: SubtractionOperator,
                    left: Bin(Box::new(BinaryExpression {
                        base: BaseNode { errors: vec![] },
                        operator: SubtractionOperator,
                        left: Int(IntegerLiteral {
                            base: BaseNode { errors: vec![] },
                            value: 1
                        }),
                        right: Int(IntegerLiteral {
                            base: BaseNode { errors: vec![] },
                            value: 2
                        })
                    })),
                    right: Int(IntegerLiteral {
                        base: BaseNode { errors: vec![] },
                        value: 3
                    })
                }))
            })]
        },
    )
}

#[test]
fn binary_operator_precedence_double_subtraction_with_parens() {
    let mut p = Parser::new(r#"1 - (2 - 3)"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Bin(Box::new(BinaryExpression {
                    base: BaseNode { errors: vec![] },
                    operator: SubtractionOperator,
                    left: Int(IntegerLiteral {
                        base: BaseNode { errors: vec![] },
                        value: 1
                    }),
                    right: Bin(Box::new(BinaryExpression {
                        base: BaseNode { errors: vec![] },
                        operator: SubtractionOperator,
                        left: Int(IntegerLiteral {
                            base: BaseNode { errors: vec![] },
                            value: 2
                        }),
                        right: Int(IntegerLiteral {
                            base: BaseNode { errors: vec![] },
                            value: 3
                        })
                    }))
                }))
            })]
        },
    )
}

#[test]
fn binary_operator_precedence_double_sum() {
    let mut p = Parser::new(r#"1 + 2 + 3"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Bin(Box::new(BinaryExpression {
                    base: BaseNode { errors: vec![] },
                    operator: AdditionOperator,
                    left: Bin(Box::new(BinaryExpression {
                        base: BaseNode { errors: vec![] },
                        operator: AdditionOperator,
                        left: Int(IntegerLiteral {
                            base: BaseNode { errors: vec![] },
                            value: 1
                        }),
                        right: Int(IntegerLiteral {
                            base: BaseNode { errors: vec![] },
                            value: 2
                        })
                    })),
                    right: Int(IntegerLiteral {
                        base: BaseNode { errors: vec![] },
                        value: 3
                    })
                }))
            })]
        },
    )
}

#[test]
fn binary_operator_precedence_double_sum_with_parens() {
    let mut p = Parser::new(r#"1 + (2 + 3)"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Bin(Box::new(BinaryExpression {
                    base: BaseNode { errors: vec![] },
                    operator: AdditionOperator,
                    left: Int(IntegerLiteral {
                        base: BaseNode { errors: vec![] },
                        value: 1
                    }),
                    right: Bin(Box::new(BinaryExpression {
                        base: BaseNode { errors: vec![] },
                        operator: AdditionOperator,
                        left: Int(IntegerLiteral {
                            base: BaseNode { errors: vec![] },
                            value: 2
                        }),
                        right: Int(IntegerLiteral {
                            base: BaseNode { errors: vec![] },
                            value: 3
                        })
                    }))
                }))
            })]
        },
    )
}

#[test]
fn logical_unary_operator_precedence() {
    let mut p = Parser::new(r#"not -1 == a"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Un(Box::new(UnaryExpression {
                    base: BaseNode { errors: vec![] },
                    operator: NotOperator,
                    argument: Bin(Box::new(BinaryExpression {
                        base: BaseNode { errors: vec![] },
                        operator: EqualOperator,
                        left: Un(Box::new(UnaryExpression {
                            base: BaseNode { errors: vec![] },
                            operator: SubtractionOperator,
                            argument: Int(IntegerLiteral {
                                base: BaseNode { errors: vec![] },
                                value: 1
                            })
                        })),
                        right: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "a".to_string()
                        })
                    }))
                }))
            })]
        },
    )
}

#[test]
fn all_operators_precedence() {
    let mut p = Parser::new(
        r#"a() == b.a + b.c * d < 100 and e != f[g] and h > i * j and
k / l < m + n - o or p() <= q() or r >= s and not t =~ /a/ and u !~ /a/"#,
    );
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Log(Box::new(LogicalExpression {
                    base: BaseNode { errors: vec![] },
                    operator: OrOperator,
                    left: Log(Box::new(LogicalExpression {
                        base: BaseNode { errors: vec![] },
                        operator: OrOperator,
                        left: Log(Box::new(LogicalExpression {
                            base: BaseNode { errors: vec![] },
                            operator: AndOperator,
                            left: Log(Box::new(LogicalExpression {
                                base: BaseNode { errors: vec![] },
                                operator: AndOperator,
                                left: Log(Box::new(LogicalExpression {
                                    base: BaseNode { errors: vec![] },
                                    operator: AndOperator,
                                    left: Bin(Box::new(BinaryExpression {
                                        base: BaseNode { errors: vec![] },
                                        operator: LessThanOperator,
                                        left: Bin(Box::new(BinaryExpression {
                                            base: BaseNode { errors: vec![] },
                                            operator: EqualOperator,
                                            left: Call(Box::new(CallExpression {
                                                base: BaseNode { errors: vec![] },
                                                arguments: vec![],
                                                callee: Idt(Identifier {
                                                    base: BaseNode { errors: vec![] },
                                                    name: "a".to_string()
                                                })
                                            })),
                                            right: Bin(Box::new(BinaryExpression {
                                                base: BaseNode { errors: vec![] },
                                                operator: AdditionOperator,
                                                left: Mem(Box::new(MemberExpression {
                                                    base: BaseNode { errors: vec![] },
                                                    object: Idt(Identifier {
                                                        base: BaseNode { errors: vec![] },
                                                        name: "b".to_string()
                                                    }),
                                                    property: PkIdt(Identifier {
                                                        base: BaseNode { errors: vec![] },
                                                        name: "a".to_string()
                                                    })
                                                })),
                                                right: Bin(Box::new(BinaryExpression {
                                                    base: BaseNode { errors: vec![] },
                                                    operator: MultiplicationOperator,
                                                    left: Mem(Box::new(MemberExpression {
                                                        base: BaseNode { errors: vec![] },
                                                        object: Idt(Identifier {
                                                            base: BaseNode { errors: vec![] },
                                                            name: "b".to_string()
                                                        }),
                                                        property: PkIdt(Identifier {
                                                            base: BaseNode { errors: vec![] },
                                                            name: "c".to_string()
                                                        })
                                                    })),
                                                    right: Idt(Identifier {
                                                        base: BaseNode { errors: vec![] },
                                                        name: "d".to_string()
                                                    })
                                                }))
                                            }))
                                        })),
                                        right: Int(IntegerLiteral {
                                            base: BaseNode { errors: vec![] },
                                            value: 100
                                        })
                                    })),
                                    right: Bin(Box::new(BinaryExpression {
                                        base: BaseNode { errors: vec![] },
                                        operator: NotEqualOperator,
                                        left: Idt(Identifier {
                                            base: BaseNode { errors: vec![] },
                                            name: "e".to_string()
                                        }),
                                        right: Idx(Box::new(IndexExpression {
                                            base: BaseNode { errors: vec![] },
                                            array: Idt(Identifier {
                                                base: BaseNode { errors: vec![] },
                                                name: "f".to_string()
                                            }),
                                            index: Idt(Identifier {
                                                base: BaseNode { errors: vec![] },
                                                name: "g".to_string()
                                            })
                                        }))
                                    }))
                                })),
                                right: Bin(Box::new(BinaryExpression {
                                    base: BaseNode { errors: vec![] },
                                    operator: GreaterThanOperator,
                                    left: Idt(Identifier {
                                        base: BaseNode { errors: vec![] },
                                        name: "h".to_string()
                                    }),
                                    right: Bin(Box::new(BinaryExpression {
                                        base: BaseNode { errors: vec![] },
                                        operator: MultiplicationOperator,
                                        left: Idt(Identifier {
                                            base: BaseNode { errors: vec![] },
                                            name: "i".to_string()
                                        }),
                                        right: Idt(Identifier {
                                            base: BaseNode { errors: vec![] },
                                            name: "j".to_string()
                                        })
                                    }))
                                }))
                            })),
                            right: Bin(Box::new(BinaryExpression {
                                base: BaseNode { errors: vec![] },
                                operator: LessThanOperator,
                                left: Bin(Box::new(BinaryExpression {
                                    base: BaseNode { errors: vec![] },
                                    operator: DivisionOperator,
                                    left: Idt(Identifier {
                                        base: BaseNode { errors: vec![] },
                                        name: "k".to_string()
                                    }),
                                    right: Idt(Identifier {
                                        base: BaseNode { errors: vec![] },
                                        name: "l".to_string()
                                    })
                                })),
                                right: Bin(Box::new(BinaryExpression {
                                    base: BaseNode { errors: vec![] },
                                    operator: SubtractionOperator,
                                    left: Bin(Box::new(BinaryExpression {
                                        base: BaseNode { errors: vec![] },
                                        operator: AdditionOperator,
                                        left: Idt(Identifier {
                                            base: BaseNode { errors: vec![] },
                                            name: "m".to_string()
                                        }),
                                        right: Idt(Identifier {
                                            base: BaseNode { errors: vec![] },
                                            name: "n".to_string()
                                        })
                                    })),
                                    right: Idt(Identifier {
                                        base: BaseNode { errors: vec![] },
                                        name: "o".to_string()
                                    })
                                }))
                            }))
                        })),
                        right: Bin(Box::new(BinaryExpression {
                            base: BaseNode { errors: vec![] },
                            operator: LessThanEqualOperator,
                            left: Call(Box::new(CallExpression {
                                base: BaseNode { errors: vec![] },
                                arguments: vec![],
                                callee: Idt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "p".to_string()
                                })
                            })),
                            right: Call(Box::new(CallExpression {
                                base: BaseNode { errors: vec![] },
                                arguments: vec![],
                                callee: Idt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "q".to_string()
                                })
                            }))
                        }))
                    })),
                    right: Log(Box::new(LogicalExpression {
                        base: BaseNode { errors: vec![] },
                        operator: AndOperator,
                        left: Log(Box::new(LogicalExpression {
                            base: BaseNode { errors: vec![] },
                            operator: AndOperator,
                            left: Bin(Box::new(BinaryExpression {
                                base: BaseNode { errors: vec![] },
                                operator: GreaterThanEqualOperator,
                                left: Idt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "r".to_string()
                                }),
                                right: Idt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "s".to_string()
                                })
                            })),
                            right: Un(Box::new(UnaryExpression {
                                base: BaseNode { errors: vec![] },
                                operator: NotOperator,
                                argument: Bin(Box::new(BinaryExpression {
                                    base: BaseNode { errors: vec![] },
                                    operator: RegexpMatchOperator,
                                    left: Idt(Identifier {
                                        base: BaseNode { errors: vec![] },
                                        name: "t".to_string()
                                    }),
                                    right: Regexp(RegexpLiteral {
                                        base: BaseNode { errors: vec![] },
                                        value: "a".to_string()
                                    })
                                }))
                            }))
                        })),
                        right: Bin(Box::new(BinaryExpression {
                            base: BaseNode { errors: vec![] },
                            operator: NotRegexpMatchOperator,
                            left: Idt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "u".to_string()
                            }),
                            right: Regexp(RegexpLiteral {
                                base: BaseNode { errors: vec![] },
                                value: "a".to_string()
                            })
                        }))
                    }))
                }))
            })]
        },
    )
}

#[test]
fn logical_operators_precedence_1() {
    let mut p = Parser::new(r#"not a or b"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Log(Box::new(LogicalExpression {
                    base: BaseNode { errors: vec![] },
                    operator: OrOperator,
                    left: Un(Box::new(UnaryExpression {
                        base: BaseNode { errors: vec![] },
                        operator: NotOperator,
                        argument: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "a".to_string()
                        })
                    })),
                    right: Idt(Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "b".to_string()
                    })
                }))
            })]
        },
    )
}

#[test]
fn logical_operators_precedence_2() {
    let mut p = Parser::new(r#"a or not b"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Log(Box::new(LogicalExpression {
                    base: BaseNode { errors: vec![] },
                    operator: OrOperator,
                    left: Idt(Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "a".to_string()
                    }),
                    right: Un(Box::new(UnaryExpression {
                        base: BaseNode { errors: vec![] },
                        operator: NotOperator,
                        argument: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "b".to_string()
                        })
                    }))
                }))
            })]
        },
    )
}

#[test]
fn logical_operators_precedence_3() {
    let mut p = Parser::new(r#"not a and b"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Log(Box::new(LogicalExpression {
                    base: BaseNode { errors: vec![] },
                    operator: AndOperator,
                    left: Un(Box::new(UnaryExpression {
                        base: BaseNode { errors: vec![] },
                        operator: NotOperator,
                        argument: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "a".to_string()
                        })
                    })),
                    right: Idt(Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "b".to_string()
                    })
                }))
            })]
        },
    )
}

#[test]
fn logical_operators_precedence_4() {
    let mut p = Parser::new(r#"a and not b"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Log(Box::new(LogicalExpression {
                    base: BaseNode { errors: vec![] },
                    operator: AndOperator,
                    left: Idt(Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "a".to_string()
                    }),
                    right: Un(Box::new(UnaryExpression {
                        base: BaseNode { errors: vec![] },
                        operator: NotOperator,
                        argument: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "b".to_string()
                        })
                    }))
                }))
            })]
        },
    )
}

#[test]
fn logical_operators_precedence_5() {
    let mut p = Parser::new(r#"a and b or c"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Log(Box::new(LogicalExpression {
                    base: BaseNode { errors: vec![] },
                    operator: OrOperator,
                    left: Log(Box::new(LogicalExpression {
                        base: BaseNode { errors: vec![] },
                        operator: AndOperator,
                        left: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "a".to_string()
                        }),
                        right: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "b".to_string()
                        })
                    })),
                    right: Idt(Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "c".to_string()
                    })
                }))
            })]
        },
    )
}

#[test]
fn logical_operators_precedence_6() {
    let mut p = Parser::new(r#"a or b and c"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Log(Box::new(LogicalExpression {
                    base: BaseNode { errors: vec![] },
                    operator: OrOperator,
                    left: Idt(Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "a".to_string()
                    }),
                    right: Log(Box::new(LogicalExpression {
                        base: BaseNode { errors: vec![] },
                        operator: AndOperator,
                        left: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "b".to_string()
                        }),
                        right: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "c".to_string()
                        })
                    }))
                }))
            })]
        },
    )
}

#[test]
fn logical_operators_precedence_7() {
    let mut p = Parser::new(r#"not (a or b)"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Un(Box::new(UnaryExpression {
                    base: BaseNode { errors: vec![] },
                    operator: NotOperator,
                    argument: Log(Box::new(LogicalExpression {
                        base: BaseNode { errors: vec![] },
                        operator: OrOperator,
                        left: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "a".to_string()
                        }),
                        right: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "b".to_string()
                        })
                    }))
                }))
            })]
        },
    )
}

#[test]
fn logical_operators_precedence_8() {
    let mut p = Parser::new(r#"not (a and b)"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Un(Box::new(UnaryExpression {
                    base: BaseNode { errors: vec![] },
                    operator: NotOperator,
                    argument: Log(Box::new(LogicalExpression {
                        base: BaseNode { errors: vec![] },
                        operator: AndOperator,
                        left: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "a".to_string()
                        }),
                        right: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "b".to_string()
                        })
                    }))
                }))
            })]
        },
    )
}

#[test]
fn logical_operators_precedence_9() {
    let mut p = Parser::new(r#"(a or b) and c"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Log(Box::new(LogicalExpression {
                    base: BaseNode { errors: vec![] },
                    operator: AndOperator,
                    left: Log(Box::new(LogicalExpression {
                        base: BaseNode { errors: vec![] },
                        operator: OrOperator,
                        left: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "a".to_string()
                        }),
                        right: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "b".to_string()
                        })
                    })),
                    right: Idt(Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "c".to_string()
                    })
                }))
            })]
        },
    )
}

#[test]
fn logical_operators_precedence_10() {
    let mut p = Parser::new(r#"a and (b or c)"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Log(Box::new(LogicalExpression {
                    base: BaseNode { errors: vec![] },
                    operator: AndOperator,
                    left: Idt(Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "a".to_string()
                    }),
                    right: Log(Box::new(LogicalExpression {
                        base: BaseNode { errors: vec![] },
                        operator: OrOperator,
                        left: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "b".to_string()
                        }),
                        right: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "c".to_string()
                        })
                    }))
                }))
            })]
        },
    )
}

// The following test case demonstrates confusing behavior:
// The `(` at 2:1 begins a call, but a user might
// reasonably expect it to start a new statement.
#[test]
fn two_logical_operations_with_parens() {
    let mut p = Parser::new(
        r#"not (a and b)
(a or b) and c"#,
    );
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Log(Box::new(LogicalExpression {
                    base: BaseNode { errors: vec![] },
                    operator: AndOperator,
                    left: Un(Box::new(UnaryExpression {
                        base: BaseNode { errors: vec![] },
                        operator: NotOperator,
                        argument: Call(Box::new(CallExpression {
                            base: BaseNode { errors: vec![] },
                            callee: Log(Box::new(LogicalExpression {
                                base: BaseNode { errors: vec![] },
                                operator: AndOperator,
                                left: Idt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "a".to_string()
                                }),
                                right: Idt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "b".to_string()
                                })
                            })),
                            arguments: vec![Obj(Box::new(ObjectExpression {
                                base: BaseNode {
                                    errors: vec![
                                        "expected comma in property list, got OR".to_string()
                                    ]
                                },
                                with: None,
                                properties: vec![
                                    Property {
                                        base: BaseNode { errors: vec![] },
                                        key: PkIdt(Identifier {
                                            base: BaseNode { errors: vec![] },
                                            name: "a".to_string()
                                        }),
                                        value: None,
                                    },
                                    Property {
                                        base: BaseNode {
                                            errors: vec![
                                                "unexpected token for property key: OR (or)"
                                                    .to_string()
                                            ]
                                        },
                                        key: PkStr(StringLiteral {
                                            base: BaseNode { errors: vec![] },
                                            value: "<invalid>".to_string()
                                        }),
                                        value: None,
                                    }
                                ]
                            }))]
                        }))
                    })),
                    right: Idt(Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "c".to_string()
                    })
                }))
            })]
        },
    )
}

#[test]
fn arrow_function_called() {
    let mut p = Parser::new(
        r#"plusOne = (r) => r + 1
plusOne(r:5)"#,
    );
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Var(VariableAssignment {
                    base: BaseNode { errors: vec![] },
                    id: Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "plusOne".to_string()
                    },
                    init: Fun(Box::new(FunctionExpression {
                        base: BaseNode { errors: vec![] },
                        params: vec![Property {
                            base: BaseNode { errors: vec![] },
                            key: PkIdt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "r".to_string()
                            }),
                            value: None
                        }],
                        body: FExpr(Bin(Box::new(BinaryExpression {
                            base: BaseNode { errors: vec![] },
                            operator: AdditionOperator,
                            left: Idt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "r".to_string()
                            }),
                            right: Int(IntegerLiteral {
                                base: BaseNode { errors: vec![] },
                                value: 1
                            })
                        })))
                    }))
                }),
                Expr(ExpressionStatement {
                    base: BaseNode { errors: vec![] },
                    expression: Call(Box::new(CallExpression {
                        base: BaseNode { errors: vec![] },
                        callee: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "plusOne".to_string()
                        }),
                        arguments: vec![Obj(Box::new(ObjectExpression {
                            base: BaseNode { errors: vec![] },
                            with: None,
                            properties: vec![Property {
                                base: BaseNode { errors: vec![] },
                                key: PkIdt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "r".to_string()
                                }),
                                value: Some(Int(IntegerLiteral {
                                    base: BaseNode { errors: vec![] },
                                    value: 5
                                }))
                            }]
                        }))]
                    }))
                })
            ]
        },
    )
}

#[test]
fn arrow_function_return_map() {
    let mut p = Parser::new(r#"toMap = (r) =>({r:r})"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Var(VariableAssignment {
                base: BaseNode { errors: vec![] },
                id: Identifier {
                    base: BaseNode { errors: vec![] },
                    name: "toMap".to_string()
                },
                init: Fun(Box::new(FunctionExpression {
                    base: BaseNode { errors: vec![] },
                    params: vec![Property {
                        base: BaseNode { errors: vec![] },
                        key: PkIdt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "r".to_string()
                        }),
                        value: None
                    }],
                    body: FExpr(Obj(Box::new(ObjectExpression {
                        base: BaseNode { errors: vec![] },
                        with: None,
                        properties: vec![Property {
                            base: BaseNode { errors: vec![] },
                            key: PkIdt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "r".to_string()
                            }),
                            value: Some(Idt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "r".to_string()
                            }))
                        }]
                    })))
                }))
            })]
        },
    )
}

#[test]
fn arrow_function_with_default_arg() {
    let mut p = Parser::new(r#"addN = (r, n=5) => r + n"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Var(VariableAssignment {
                base: BaseNode { errors: vec![] },
                id: Identifier {
                    base: BaseNode { errors: vec![] },
                    name: "addN".to_string()
                },
                init: Fun(Box::new(FunctionExpression {
                    base: BaseNode { errors: vec![] },
                    params: vec![
                        Property {
                            base: BaseNode { errors: vec![] },
                            key: PkIdt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "r".to_string()
                            }),
                            value: None
                        },
                        Property {
                            base: BaseNode { errors: vec![] },
                            key: PkIdt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "n".to_string()
                            }),
                            value: Some(Int(IntegerLiteral {
                                base: BaseNode { errors: vec![] },
                                value: 5
                            }))
                        }
                    ],
                    body: FExpr(Bin(Box::new(BinaryExpression {
                        base: BaseNode { errors: vec![] },
                        operator: AdditionOperator,
                        left: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "r".to_string()
                        }),
                        right: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "n".to_string()
                        })
                    })))
                }))
            })]
        },
    )
}

#[test]
fn arrow_function_called_in_binary_expression() {
    let mut p = Parser::new(
        r#"
            plusOne = (r) => r + 1
            plusOne(r:5) == 6 or die()"#,
    );
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Var(VariableAssignment {
                    base: BaseNode { errors: vec![] },
                    id: Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "plusOne".to_string()
                    },
                    init: Fun(Box::new(FunctionExpression {
                        base: BaseNode { errors: vec![] },
                        params: vec![Property {
                            base: BaseNode { errors: vec![] },
                            key: PkIdt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "r".to_string()
                            }),
                            value: None
                        }],
                        body: FExpr(Bin(Box::new(BinaryExpression {
                            base: BaseNode { errors: vec![] },
                            operator: AdditionOperator,
                            left: Idt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "r".to_string()
                            }),
                            right: Int(IntegerLiteral {
                                base: BaseNode { errors: vec![] },
                                value: 1
                            })
                        })))
                    }))
                }),
                Expr(ExpressionStatement {
                    base: BaseNode { errors: vec![] },
                    expression: Log(Box::new(LogicalExpression {
                        base: BaseNode { errors: vec![] },
                        operator: OrOperator,
                        left: Bin(Box::new(BinaryExpression {
                            base: BaseNode { errors: vec![] },
                            operator: EqualOperator,
                            left: Call(Box::new(CallExpression {
                                base: BaseNode { errors: vec![] },
                                callee: Idt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "plusOne".to_string()
                                }),
                                arguments: vec![Obj(Box::new(ObjectExpression {
                                    base: BaseNode { errors: vec![] },
                                    with: None,
                                    properties: vec![Property {
                                        base: BaseNode { errors: vec![] },
                                        key: PkIdt(Identifier {
                                            base: BaseNode { errors: vec![] },
                                            name: "r".to_string()
                                        }),
                                        value: Some(Int(IntegerLiteral {
                                            base: BaseNode { errors: vec![] },
                                            value: 5
                                        }))
                                    }]
                                }))]
                            })),
                            right: Int(IntegerLiteral {
                                base: BaseNode { errors: vec![] },
                                value: 6
                            })
                        })),
                        right: Call(Box::new(CallExpression {
                            base: BaseNode { errors: vec![] },
                            arguments: vec![],
                            callee: Idt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "die".to_string()
                            })
                        }))
                    }))
                })
            ]
        },
    )
}

#[test]
fn arrow_function_as_single_expression() {
    let mut p = Parser::new(r#"f = (r) => r["_measurement"] == "cpu""#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Var(VariableAssignment {
                base: BaseNode { errors: vec![] },
                id: Identifier {
                    base: BaseNode { errors: vec![] },
                    name: "f".to_string()
                },
                init: Fun(Box::new(FunctionExpression {
                    base: BaseNode { errors: vec![] },
                    params: vec![Property {
                        base: BaseNode { errors: vec![] },
                        key: PkIdt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "r".to_string()
                        }),
                        value: None
                    }],
                    body: FExpr(Bin(Box::new(BinaryExpression {
                        base: BaseNode { errors: vec![] },
                        operator: EqualOperator,
                        left: Mem(Box::new(MemberExpression {
                            base: BaseNode { errors: vec![] },
                            object: Idt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "r".to_string()
                            }),
                            property: PkStr(StringLiteral {
                                base: BaseNode { errors: vec![] },
                                value: "_measurement".to_string()
                            })
                        })),
                        right: Str(StringLiteral {
                            base: BaseNode { errors: vec![] },
                            value: "cpu".to_string()
                        })
                    })))
                }))
            })]
        },
    )
}

#[test]
fn arrow_function_as_block() {
    let mut p = Parser::new(
        r#"f = (r) => { 
                m = r["_measurement"]
                return m == "cpu"
            }"#,
    );
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Var(VariableAssignment {
                base: BaseNode { errors: vec![] },
                id: Identifier {
                    base: BaseNode { errors: vec![] },
                    name: "f".to_string()
                },
                init: Fun(Box::new(FunctionExpression {
                    base: BaseNode { errors: vec![] },
                    params: vec![Property {
                        base: BaseNode { errors: vec![] },
                        key: PkIdt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "r".to_string()
                        }),
                        value: None
                    }],
                    body: FBlock(Block {
                        base: BaseNode { errors: vec![] },
                        body: vec![
                            Var(VariableAssignment {
                                base: BaseNode { errors: vec![] },
                                id: Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "m".to_string()
                                },
                                init: Mem(Box::new(MemberExpression {
                                    base: BaseNode { errors: vec![] },
                                    object: Idt(Identifier {
                                        base: BaseNode { errors: vec![] },
                                        name: "r".to_string()
                                    }),
                                    property: PkStr(StringLiteral {
                                        base: BaseNode { errors: vec![] },
                                        value: "_measurement".to_string()
                                    })
                                }))
                            }),
                            Ret(ReturnStatement {
                                base: BaseNode { errors: vec![] },
                                argument: Bin(Box::new(BinaryExpression {
                                    base: BaseNode { errors: vec![] },
                                    operator: EqualOperator,
                                    left: Idt(Identifier {
                                        base: BaseNode { errors: vec![] },
                                        name: "m".to_string()
                                    }),
                                    right: Str(StringLiteral {
                                        base: BaseNode { errors: vec![] },
                                        value: "cpu".to_string()
                                    })
                                }))
                            })
                        ]
                    })
                }))
            })]
        },
    )
}

#[test]
fn conditional() {
    let mut p = Parser::new(r#"a = if true then 0 else 1"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Var(VariableAssignment {
                base: BaseNode { errors: vec![] },
                id: Identifier {
                    base: BaseNode { errors: vec![] },
                    name: "a".to_string()
                },
                init: Cond(Box::new(ConditionalExpression {
                    base: BaseNode { errors: vec![] },
                    test: Idt(Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "true".to_string()
                    }),
                    consequent: Int(IntegerLiteral {
                        base: BaseNode { errors: vec![] },
                        value: 0
                    }),
                    alternate: Int(IntegerLiteral {
                        base: BaseNode { errors: vec![] },
                        value: 1
                    })
                }))
            })]
        },
    )
}

#[test]
fn nested_conditionals() {
    let mut p = Parser::new(
        r#"if if b < 0 then true else false
                  then if c > 0 then 30 else 60
                  else if d == 0 then 90 else 120"#,
    );
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Cond(Box::new(ConditionalExpression {
                    base: BaseNode { errors: vec![] },
                    test: Cond(Box::new(ConditionalExpression {
                        base: BaseNode { errors: vec![] },
                        test: Bin(Box::new(BinaryExpression {
                            base: BaseNode { errors: vec![] },
                            operator: LessThanOperator,
                            left: Idt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "b".to_string()
                            }),
                            right: Int(IntegerLiteral {
                                base: BaseNode { errors: vec![] },
                                value: 0
                            })
                        })),
                        consequent: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "true".to_string()
                        }),
                        alternate: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "false".to_string()
                        })
                    })),
                    consequent: Cond(Box::new(ConditionalExpression {
                        base: BaseNode { errors: vec![] },
                        test: Bin(Box::new(BinaryExpression {
                            base: BaseNode { errors: vec![] },
                            operator: GreaterThanOperator,
                            left: Idt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "c".to_string()
                            }),
                            right: Int(IntegerLiteral {
                                base: BaseNode { errors: vec![] },
                                value: 0
                            })
                        })),
                        consequent: Int(IntegerLiteral {
                            base: BaseNode { errors: vec![] },
                            value: 30
                        }),
                        alternate: Int(IntegerLiteral {
                            base: BaseNode { errors: vec![] },
                            value: 60
                        })
                    })),
                    alternate: Cond(Box::new(ConditionalExpression {
                        base: BaseNode { errors: vec![] },
                        test: Bin(Box::new(BinaryExpression {
                            base: BaseNode { errors: vec![] },
                            operator: EqualOperator,
                            left: Idt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "d".to_string()
                            }),
                            right: Int(IntegerLiteral {
                                base: BaseNode { errors: vec![] },
                                value: 0
                            })
                        })),
                        consequent: Int(IntegerLiteral {
                            base: BaseNode { errors: vec![] },
                            value: 90
                        }),
                        alternate: Int(IntegerLiteral {
                            base: BaseNode { errors: vec![] },
                            value: 120
                        })
                    }))
                }))
            })]
        },
    )
}

#[test]
fn from_with_filter_with_no_parens() {
    let mut p = Parser::new(r#"from(bucket:"telegraf/autogen").filter(fn: (r) => r["other"]=="mem" and r["this"]=="that" or r["these"]!="those")"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Call(Box::new(CallExpression {
                    base: BaseNode { errors: vec![] },
                    callee: Mem(Box::new(MemberExpression {
                        base: BaseNode { errors: vec![] },
                        object: Call(Box::new(CallExpression {
                            base: BaseNode { errors: vec![] },
                            callee: Idt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "from".to_string()
                            }),
                            arguments: vec![Obj(Box::new(ObjectExpression {
                                base: BaseNode { errors: vec![] },
                                with: None,
                                properties: vec![Property {
                                    base: BaseNode { errors: vec![] },
                                    key: PkIdt(Identifier {
                                        base: BaseNode { errors: vec![] },
                                        name: "bucket".to_string()
                                    }),
                                    value: Some(Str(StringLiteral {
                                        base: BaseNode { errors: vec![] },
                                        value: "telegraf/autogen".to_string()
                                    }))
                                }]
                            }))]
                        })),
                        property: PkIdt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "filter".to_string()
                        })
                    })),
                    arguments: vec![Obj(Box::new(ObjectExpression {
                        base: BaseNode { errors: vec![] },
                        with: None,
                        properties: vec![Property {
                            base: BaseNode { errors: vec![] },
                            key: PkIdt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "fn".to_string()
                            }),
                            value: Some(Fun(Box::new(FunctionExpression {
                                base: BaseNode { errors: vec![] },
                                params: vec![Property {
                                    base: BaseNode { errors: vec![] },
                                    key: PkIdt(Identifier {
                                        base: BaseNode { errors: vec![] },
                                        name: "r".to_string()
                                    }),
                                    value: None
                                }],
                                body: FExpr(Log(Box::new(LogicalExpression {
                                    base: BaseNode { errors: vec![] },
                                    operator: OrOperator,
                                    left: Log(Box::new(LogicalExpression {
                                        base: BaseNode { errors: vec![] },
                                        operator: AndOperator,
                                        left: Bin(Box::new(BinaryExpression {
                                            base: BaseNode { errors: vec![] },
                                            operator: EqualOperator,
                                            left: Mem(Box::new(MemberExpression {
                                                base: BaseNode { errors: vec![] },
                                                object: Idt(Identifier {
                                                    base: BaseNode { errors: vec![] },
                                                    name: "r".to_string()
                                                }),
                                                property: PkStr(StringLiteral {
                                                    base: BaseNode { errors: vec![] },
                                                    value: "other".to_string()
                                                })
                                            })),
                                            right: Str(StringLiteral {
                                                base: BaseNode { errors: vec![] },
                                                value: "mem".to_string()
                                            })
                                        })),
                                        right: Bin(Box::new(BinaryExpression {
                                            base: BaseNode { errors: vec![] },
                                            operator: EqualOperator,
                                            left: Mem(Box::new(MemberExpression {
                                                base: BaseNode { errors: vec![] },
                                                object: Idt(Identifier {
                                                    base: BaseNode { errors: vec![] },
                                                    name: "r".to_string()
                                                }),
                                                property: PkStr(StringLiteral {
                                                    base: BaseNode { errors: vec![] },
                                                    value: "this".to_string()
                                                })
                                            })),
                                            right: Str(StringLiteral {
                                                base: BaseNode { errors: vec![] },
                                                value: "that".to_string()
                                            })
                                        }))
                                    })),
                                    right: Bin(Box::new(BinaryExpression {
                                        base: BaseNode { errors: vec![] },
                                        operator: NotEqualOperator,
                                        left: Mem(Box::new(MemberExpression {
                                            base: BaseNode { errors: vec![] },
                                            object: Idt(Identifier {
                                                base: BaseNode { errors: vec![] },
                                                name: "r".to_string()
                                            }),
                                            property: PkStr(StringLiteral {
                                                base: BaseNode { errors: vec![] },
                                                value: "these".to_string()
                                            })
                                        })),
                                        right: Str(StringLiteral {
                                            base: BaseNode { errors: vec![] },
                                            value: "those".to_string()
                                        })
                                    }))
                                })))
                            })))
                        }]
                    }))]
                }))
            })]
        },
    )
}

#[test]
fn from_with_range() {
    let mut p = Parser::new(r#"from(bucket:"telegraf/autogen")|>range(start:-1h, end:10m)"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Pipe(Box::new(PipeExpression {
                    base: BaseNode { errors: vec![] },
                    argument: Call(Box::new(CallExpression {
                        base: BaseNode { errors: vec![] },
                        callee: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "from".to_string()
                        }),
                        arguments: vec![Obj(Box::new(ObjectExpression {
                            base: BaseNode { errors: vec![] },
                            with: None,
                            properties: vec![Property {
                                base: BaseNode { errors: vec![] },
                                key: PkIdt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "bucket".to_string()
                                }),
                                value: Some(Str(StringLiteral {
                                    base: BaseNode { errors: vec![] },
                                    value: "telegraf/autogen".to_string()
                                }))
                            }]
                        }))]
                    })),
                    call: CallExpression {
                        base: BaseNode { errors: vec![] },
                        callee: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "range".to_string()
                        }),
                        arguments: vec![Obj(Box::new(ObjectExpression {
                            base: BaseNode { errors: vec![] },
                            with: None,
                            properties: vec![
                                Property {
                                    base: BaseNode { errors: vec![] },
                                    key: PkIdt(Identifier {
                                        base: BaseNode { errors: vec![] },
                                        name: "start".to_string()
                                    }),
                                    value: Some(Un(Box::new(UnaryExpression {
                                        base: BaseNode { errors: vec![] },
                                        operator: SubtractionOperator,
                                        argument: Dur(DurationLiteral {
                                            base: BaseNode { errors: vec![] },
                                            values: vec![Duration {
                                                magnitude: 1,
                                                unit: "h".to_string()
                                            }]
                                        })
                                    })))
                                },
                                Property {
                                    base: BaseNode { errors: vec![] },
                                    key: PkIdt(Identifier {
                                        base: BaseNode { errors: vec![] },
                                        name: "end".to_string()
                                    }),
                                    value: Some(Dur(DurationLiteral {
                                        base: BaseNode { errors: vec![] },
                                        values: vec![Duration {
                                            magnitude: 10,
                                            unit: "m".to_string()
                                        }]
                                    }))
                                }
                            ]
                        }))]
                    }
                }))
            })]
        },
    )
}

#[test]
fn from_with_limit() {
    let mut p = Parser::new(r#"from(bucket:"telegraf/autogen")|>limit(limit:100, offset:10)"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Pipe(Box::new(PipeExpression {
                    base: BaseNode { errors: vec![] },
                    argument: Call(Box::new(CallExpression {
                        base: BaseNode { errors: vec![] },
                        callee: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "from".to_string()
                        }),
                        arguments: vec![Obj(Box::new(ObjectExpression {
                            base: BaseNode { errors: vec![] },
                            with: None,
                            properties: vec![Property {
                                base: BaseNode { errors: vec![] },
                                key: PkIdt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "bucket".to_string()
                                }),
                                value: Some(Str(StringLiteral {
                                    base: BaseNode { errors: vec![] },
                                    value: "telegraf/autogen".to_string()
                                }))
                            }]
                        }))]
                    })),
                    call: CallExpression {
                        base: BaseNode { errors: vec![] },
                        callee: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "limit".to_string()
                        }),
                        arguments: vec![Obj(Box::new(ObjectExpression {
                            base: BaseNode { errors: vec![] },
                            with: None,
                            properties: vec![
                                Property {
                                    base: BaseNode { errors: vec![] },
                                    key: PkIdt(Identifier {
                                        base: BaseNode { errors: vec![] },
                                        name: "limit".to_string()
                                    }),
                                    value: Some(Int(IntegerLiteral {
                                        base: BaseNode { errors: vec![] },
                                        value: 100
                                    }))
                                },
                                Property {
                                    base: BaseNode { errors: vec![] },
                                    key: PkIdt(Identifier {
                                        base: BaseNode { errors: vec![] },
                                        name: "offset".to_string()
                                    }),
                                    value: Some(Int(IntegerLiteral {
                                        base: BaseNode { errors: vec![] },
                                        value: 10
                                    }))
                                }
                            ]
                        }))]
                    }
                }))
            })]
        },
    )
}

#[test]
fn from_with_range_and_count() {
    let mut p = Parser::new(
        r#"from(bucket:"mydb/autogen")
						|> range(start:-4h, stop:-2h)
						|> count()"#,
    );
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Pipe(Box::new(PipeExpression {
                    base: BaseNode { errors: vec![] },
                    argument: Pipe(Box::new(PipeExpression {
                        base: BaseNode { errors: vec![] },
                        argument: Call(Box::new(CallExpression {
                            base: BaseNode { errors: vec![] },
                            callee: Idt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "from".to_string()
                            }),
                            arguments: vec![Obj(Box::new(ObjectExpression {
                                base: BaseNode { errors: vec![] },
                                with: None,
                                properties: vec![Property {
                                    base: BaseNode { errors: vec![] },
                                    key: PkIdt(Identifier {
                                        base: BaseNode { errors: vec![] },
                                        name: "bucket".to_string()
                                    }),
                                    value: Some(Str(StringLiteral {
                                        base: BaseNode { errors: vec![] },
                                        value: "mydb/autogen".to_string()
                                    }))
                                }]
                            }))]
                        })),
                        call: CallExpression {
                            base: BaseNode { errors: vec![] },
                            callee: Idt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "range".to_string()
                            }),
                            arguments: vec![Obj(Box::new(ObjectExpression {
                                base: BaseNode { errors: vec![] },
                                with: None,
                                properties: vec![
                                    Property {
                                        base: BaseNode { errors: vec![] },
                                        key: PkIdt(Identifier {
                                            base: BaseNode { errors: vec![] },
                                            name: "start".to_string()
                                        }),
                                        value: Some(Un(Box::new(UnaryExpression {
                                            base: BaseNode { errors: vec![] },
                                            operator: SubtractionOperator,
                                            argument: Dur(DurationLiteral {
                                                base: BaseNode { errors: vec![] },
                                                values: vec![Duration {
                                                    magnitude: 4,
                                                    unit: "h".to_string()
                                                }]
                                            })
                                        })))
                                    },
                                    Property {
                                        base: BaseNode { errors: vec![] },
                                        key: PkIdt(Identifier {
                                            base: BaseNode { errors: vec![] },
                                            name: "stop".to_string()
                                        }),
                                        value: Some(Un(Box::new(UnaryExpression {
                                            base: BaseNode { errors: vec![] },
                                            operator: SubtractionOperator,
                                            argument: Dur(DurationLiteral {
                                                base: BaseNode { errors: vec![] },
                                                values: vec![Duration {
                                                    magnitude: 2,
                                                    unit: "h".to_string()
                                                }]
                                            })
                                        })))
                                    }
                                ]
                            }))]
                        }
                    })),
                    call: CallExpression {
                        base: BaseNode { errors: vec![] },
                        callee: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "count".to_string()
                        }),
                        arguments: vec![]
                    }
                }))
            })]
        }
    )
}

#[test]
fn from_with_range_limit_and_count() {
    let mut p = Parser::new(
        r#"from(bucket:"mydb/autogen")
						|> range(start:-4h, stop:-2h)
						|> limit(n:10)
						|> count()"#,
    );
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Pipe(Box::new(PipeExpression {
                    base: BaseNode { errors: vec![] },
                    argument: Pipe(Box::new(PipeExpression {
                        base: BaseNode { errors: vec![] },
                        argument: Pipe(Box::new(PipeExpression {
                            base: BaseNode { errors: vec![] },
                            argument: Call(Box::new(CallExpression {
                                base: BaseNode { errors: vec![] },
                                callee: Idt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "from".to_string()
                                }),
                                arguments: vec![Obj(Box::new(ObjectExpression {
                                    base: BaseNode { errors: vec![] },
                                    with: None,
                                    properties: vec![Property {
                                        base: BaseNode { errors: vec![] },
                                        key: PkIdt(Identifier {
                                            base: BaseNode { errors: vec![] },
                                            name: "bucket".to_string()
                                        }),
                                        value: Some(Str(StringLiteral {
                                            base: BaseNode { errors: vec![] },
                                            value: "mydb/autogen".to_string()
                                        }))
                                    }]
                                }))]
                            })),
                            call: CallExpression {
                                base: BaseNode { errors: vec![] },
                                callee: Idt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "range".to_string()
                                }),
                                arguments: vec![Obj(Box::new(ObjectExpression {
                                    base: BaseNode { errors: vec![] },
                                    with: None,
                                    properties: vec![
                                        Property {
                                            base: BaseNode { errors: vec![] },
                                            key: PkIdt(Identifier {
                                                base: BaseNode { errors: vec![] },
                                                name: "start".to_string()
                                            }),
                                            value: Some(Un(Box::new(UnaryExpression {
                                                base: BaseNode { errors: vec![] },
                                                operator: SubtractionOperator,
                                                argument: Dur(DurationLiteral {
                                                    base: BaseNode { errors: vec![] },
                                                    values: vec![Duration {
                                                        magnitude: 4,
                                                        unit: "h".to_string()
                                                    }]
                                                })
                                            })))
                                        },
                                        Property {
                                            base: BaseNode { errors: vec![] },
                                            key: PkIdt(Identifier {
                                                base: BaseNode { errors: vec![] },
                                                name: "stop".to_string()
                                            }),
                                            value: Some(Un(Box::new(UnaryExpression {
                                                base: BaseNode { errors: vec![] },
                                                operator: SubtractionOperator,
                                                argument: Dur(DurationLiteral {
                                                    base: BaseNode { errors: vec![] },
                                                    values: vec![Duration {
                                                        magnitude: 2,
                                                        unit: "h".to_string()
                                                    }]
                                                })
                                            })))
                                        }
                                    ]
                                }))]
                            }
                        })),
                        call: CallExpression {
                            base: BaseNode { errors: vec![] },
                            callee: Idt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "limit".to_string()
                            }),
                            arguments: vec![Obj(Box::new(ObjectExpression {
                                base: BaseNode { errors: vec![] },
                                with: None,
                                properties: vec![Property {
                                    base: BaseNode { errors: vec![] },
                                    key: PkIdt(Identifier {
                                        base: BaseNode { errors: vec![] },
                                        name: "n".to_string()
                                    }),
                                    value: Some(Int(IntegerLiteral {
                                        base: BaseNode { errors: vec![] },
                                        value: 10
                                    }))
                                }]
                            }))]
                        }
                    })),
                    call: CallExpression {
                        base: BaseNode { errors: vec![] },
                        callee: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "count".to_string()
                        }),
                        arguments: vec![]
                    }
                }))
            })]
        }
    )
}

#[test]
fn from_with_join() {
    let mut p = Parser::new(
        r#"
a = from(bucket:"dbA/autogen") |> range(start:-1h)
b = from(bucket:"dbB/autogen") |> range(start:-1h)
join(tables:[a,b], on:["host"], fn: (a,b) => a["_field"] + b["_field"])"#,
    );
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Var(VariableAssignment {
                    base: BaseNode { errors: vec![] },
                    id: Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "a".to_string()
                    },
                    init: Pipe(Box::new(PipeExpression {
                        base: BaseNode { errors: vec![] },
                        argument: Call(Box::new(CallExpression {
                            base: BaseNode { errors: vec![] },
                            callee: Idt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "from".to_string()
                            }),
                            arguments: vec![Obj(Box::new(ObjectExpression {
                                base: BaseNode { errors: vec![] },
                                with: None,
                                properties: vec![Property {
                                    base: BaseNode { errors: vec![] },
                                    key: PkIdt(Identifier {
                                        base: BaseNode { errors: vec![] },
                                        name: "bucket".to_string()
                                    }),
                                    value: Some(Str(StringLiteral {
                                        base: BaseNode { errors: vec![] },
                                        value: "dbA/autogen".to_string()
                                    }))
                                }]
                            }))]
                        })),
                        call: CallExpression {
                            base: BaseNode { errors: vec![] },
                            callee: Idt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "range".to_string()
                            }),
                            arguments: vec![Obj(Box::new(ObjectExpression {
                                base: BaseNode { errors: vec![] },
                                with: None,
                                properties: vec![Property {
                                    base: BaseNode { errors: vec![] },
                                    key: PkIdt(Identifier {
                                        base: BaseNode { errors: vec![] },
                                        name: "start".to_string()
                                    }),
                                    value: Some(Un(Box::new(UnaryExpression {
                                        base: BaseNode { errors: vec![] },
                                        operator: SubtractionOperator,
                                        argument: Dur(DurationLiteral {
                                            base: BaseNode { errors: vec![] },
                                            values: vec![Duration {
                                                magnitude: 1,
                                                unit: "h".to_string()
                                            }]
                                        })
                                    })))
                                }]
                            }))]
                        }
                    }))
                }),
                Var(VariableAssignment {
                    base: BaseNode { errors: vec![] },
                    id: Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "b".to_string()
                    },
                    init: Pipe(Box::new(PipeExpression {
                        base: BaseNode { errors: vec![] },
                        argument: Call(Box::new(CallExpression {
                            base: BaseNode { errors: vec![] },
                            callee: Idt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "from".to_string()
                            }),
                            arguments: vec![Obj(Box::new(ObjectExpression {
                                base: BaseNode { errors: vec![] },
                                with: None,
                                properties: vec![Property {
                                    base: BaseNode { errors: vec![] },
                                    key: PkIdt(Identifier {
                                        base: BaseNode { errors: vec![] },
                                        name: "bucket".to_string()
                                    }),
                                    value: Some(Str(StringLiteral {
                                        base: BaseNode { errors: vec![] },
                                        value: "dbB/autogen".to_string()
                                    }))
                                }]
                            }))]
                        })),
                        call: CallExpression {
                            base: BaseNode { errors: vec![] },
                            callee: Idt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "range".to_string()
                            }),
                            arguments: vec![Obj(Box::new(ObjectExpression {
                                base: BaseNode { errors: vec![] },
                                with: None,
                                properties: vec![Property {
                                    base: BaseNode { errors: vec![] },
                                    key: PkIdt(Identifier {
                                        base: BaseNode { errors: vec![] },
                                        name: "start".to_string()
                                    }),
                                    value: Some(Un(Box::new(UnaryExpression {
                                        base: BaseNode { errors: vec![] },
                                        operator: SubtractionOperator,
                                        argument: Dur(DurationLiteral {
                                            base: BaseNode { errors: vec![] },
                                            values: vec![Duration {
                                                magnitude: 1,
                                                unit: "h".to_string()
                                            }]
                                        })
                                    })))
                                }]
                            }))]
                        }
                    }))
                }),
                Expr(ExpressionStatement {
                    base: BaseNode { errors: vec![] },
                    expression: Call(Box::new(CallExpression {
                        base: BaseNode { errors: vec![] },
                        callee: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "join".to_string()
                        }),
                        arguments: vec![Obj(Box::new(ObjectExpression {
                            base: BaseNode { errors: vec![] },
                            with: None,
                            properties: vec![
                                Property {
                                    base: BaseNode { errors: vec![] },
                                    key: PkIdt(Identifier {
                                        base: BaseNode { errors: vec![] },
                                        name: "tables".to_string()
                                    }),
                                    value: Some(Arr(Box::new(ArrayExpression {
                                        base: BaseNode { errors: vec![] },
                                        elements: vec![
                                            Idt(Identifier {
                                                base: BaseNode { errors: vec![] },
                                                name: "a".to_string()
                                            }),
                                            Idt(Identifier {
                                                base: BaseNode { errors: vec![] },
                                                name: "b".to_string()
                                            })
                                        ]
                                    })))
                                },
                                Property {
                                    base: BaseNode { errors: vec![] },
                                    key: PkIdt(Identifier {
                                        base: BaseNode { errors: vec![] },
                                        name: "on".to_string()
                                    }),
                                    value: Some(Arr(Box::new(ArrayExpression {
                                        base: BaseNode { errors: vec![] },
                                        elements: vec![Str(StringLiteral {
                                            base: BaseNode { errors: vec![] },
                                            value: "host".to_string()
                                        })]
                                    })))
                                },
                                Property {
                                    base: BaseNode { errors: vec![] },
                                    key: PkIdt(Identifier {
                                        base: BaseNode { errors: vec![] },
                                        name: "fn".to_string()
                                    }),
                                    value: Some(Fun(Box::new(FunctionExpression {
                                        base: BaseNode { errors: vec![] },
                                        params: vec![
                                            Property {
                                                base: BaseNode { errors: vec![] },
                                                key: PkIdt(Identifier {
                                                    base: BaseNode { errors: vec![] },
                                                    name: "a".to_string()
                                                }),
                                                value: None
                                            },
                                            Property {
                                                base: BaseNode { errors: vec![] },
                                                key: PkIdt(Identifier {
                                                    base: BaseNode { errors: vec![] },
                                                    name: "b".to_string()
                                                }),
                                                value: None
                                            }
                                        ],
                                        body: FExpr(Bin(Box::new(BinaryExpression {
                                            base: BaseNode { errors: vec![] },
                                            operator: AdditionOperator,
                                            left: Mem(Box::new(MemberExpression {
                                                base: BaseNode { errors: vec![] },
                                                object: Idt(Identifier {
                                                    base: BaseNode { errors: vec![] },
                                                    name: "a".to_string()
                                                }),
                                                property: PkStr(StringLiteral {
                                                    base: BaseNode { errors: vec![] },
                                                    value: "_field".to_string()
                                                })
                                            })),
                                            right: Mem(Box::new(MemberExpression {
                                                base: BaseNode { errors: vec![] },
                                                object: Idt(Identifier {
                                                    base: BaseNode { errors: vec![] },
                                                    name: "b".to_string()
                                                }),
                                                property: PkStr(StringLiteral {
                                                    base: BaseNode { errors: vec![] },
                                                    value: "_field".to_string()
                                                })
                                            }))
                                        })))
                                    })))
                                }
                            ]
                        }))]
                    }))
                })
            ]
        },
    )
}

#[test]
fn from_with_join_with_complex_expression() {
    let mut p = Parser::new(
        r#"
a = from(bucket:"Flux/autogen")
	|> filter(fn: (r) => r["_measurement"] == "a")
	|> range(start:-1h)

b = from(bucket:"Flux/autogen")
	|> filter(fn: (r) => r["_measurement"] == "b")
	|> range(start:-1h)

join(tables:[a,b], on:["t1"], fn: (a,b) => (a["_field"] - b["_field"]) / b["_field"])"#,
    );
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Var(VariableAssignment {
                    base: BaseNode { errors: vec![] },
                    id: Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "a".to_string()
                    },
                    init: Pipe(Box::new(PipeExpression {
                        base: BaseNode { errors: vec![] },
                        argument: Pipe(Box::new(PipeExpression {
                            base: BaseNode { errors: vec![] },
                            argument: Call(Box::new(CallExpression {
                                base: BaseNode { errors: vec![] },
                                callee: Idt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "from".to_string()
                                }),
                                arguments: vec![Obj(Box::new(ObjectExpression {
                                    base: BaseNode { errors: vec![] },
                                    with: None,
                                    properties: vec![Property {
                                        base: BaseNode { errors: vec![] },
                                        key: PkIdt(Identifier {
                                            base: BaseNode { errors: vec![] },
                                            name: "bucket".to_string()
                                        }),
                                        value: Some(Str(StringLiteral {
                                            base: BaseNode { errors: vec![] },
                                            value: "Flux/autogen".to_string()
                                        }))
                                    }]
                                }))]
                            })),
                            call: CallExpression {
                                base: BaseNode { errors: vec![] },
                                callee: Idt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "filter".to_string()
                                }),
                                arguments: vec![Obj(Box::new(ObjectExpression {
                                    base: BaseNode { errors: vec![] },
                                    with: None,
                                    properties: vec![Property {
                                        base: BaseNode { errors: vec![] },
                                        key: PkIdt(Identifier {
                                            base: BaseNode { errors: vec![] },
                                            name: "fn".to_string()
                                        }),
                                        value: Some(Fun(Box::new(FunctionExpression {
                                            base: BaseNode { errors: vec![] },
                                            params: vec![Property {
                                                base: BaseNode { errors: vec![] },
                                                key: PkIdt(Identifier {
                                                    base: BaseNode { errors: vec![] },
                                                    name: "r".to_string()
                                                }),
                                                value: None
                                            }],
                                            body: FExpr(Bin(Box::new(BinaryExpression {
                                                base: BaseNode { errors: vec![] },
                                                operator: EqualOperator,
                                                left: Mem(Box::new(MemberExpression {
                                                    base: BaseNode { errors: vec![] },
                                                    object: Idt(Identifier {
                                                        base: BaseNode { errors: vec![] },
                                                        name: "r".to_string()
                                                    }),
                                                    property: PkStr(StringLiteral {
                                                        base: BaseNode { errors: vec![] },
                                                        value: "_measurement".to_string()
                                                    })
                                                })),
                                                right: Str(StringLiteral {
                                                    base: BaseNode { errors: vec![] },
                                                    value: "a".to_string()
                                                })
                                            })))
                                        })))
                                    }]
                                }))]
                            }
                        })),
                        call: CallExpression {
                            base: BaseNode { errors: vec![] },
                            callee: Idt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "range".to_string()
                            }),
                            arguments: vec![Obj(Box::new(ObjectExpression {
                                base: BaseNode { errors: vec![] },
                                with: None,
                                properties: vec![Property {
                                    base: BaseNode { errors: vec![] },
                                    key: PkIdt(Identifier {
                                        base: BaseNode { errors: vec![] },
                                        name: "start".to_string()
                                    }),
                                    value: Some(Un(Box::new(UnaryExpression {
                                        base: BaseNode { errors: vec![] },
                                        operator: SubtractionOperator,
                                        argument: Dur(DurationLiteral {
                                            base: BaseNode { errors: vec![] },
                                            values: vec![Duration {
                                                magnitude: 1,
                                                unit: "h".to_string()
                                            }]
                                        })
                                    })))
                                }]
                            }))]
                        }
                    }))
                }),
                Var(VariableAssignment {
                    base: BaseNode { errors: vec![] },
                    id: Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "b".to_string()
                    },
                    init: Pipe(Box::new(PipeExpression {
                        base: BaseNode { errors: vec![] },
                        argument: Pipe(Box::new(PipeExpression {
                            base: BaseNode { errors: vec![] },
                            argument: Call(Box::new(CallExpression {
                                base: BaseNode { errors: vec![] },
                                callee: Idt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "from".to_string()
                                }),
                                arguments: vec![Obj(Box::new(ObjectExpression {
                                    base: BaseNode { errors: vec![] },
                                    with: None,
                                    properties: vec![Property {
                                        base: BaseNode { errors: vec![] },
                                        key: PkIdt(Identifier {
                                            base: BaseNode { errors: vec![] },
                                            name: "bucket".to_string()
                                        }),
                                        value: Some(Str(StringLiteral {
                                            base: BaseNode { errors: vec![] },
                                            value: "Flux/autogen".to_string()
                                        }))
                                    }]
                                }))]
                            })),
                            call: CallExpression {
                                base: BaseNode { errors: vec![] },
                                callee: Idt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "filter".to_string()
                                }),
                                arguments: vec![Obj(Box::new(ObjectExpression {
                                    base: BaseNode { errors: vec![] },
                                    with: None,
                                    properties: vec![Property {
                                        base: BaseNode { errors: vec![] },
                                        key: PkIdt(Identifier {
                                            base: BaseNode { errors: vec![] },
                                            name: "fn".to_string()
                                        }),
                                        value: Some(Fun(Box::new(FunctionExpression {
                                            base: BaseNode { errors: vec![] },
                                            params: vec![Property {
                                                base: BaseNode { errors: vec![] },
                                                key: PkIdt(Identifier {
                                                    base: BaseNode { errors: vec![] },
                                                    name: "r".to_string()
                                                }),
                                                value: None
                                            }],
                                            body: FExpr(Bin(Box::new(BinaryExpression {
                                                base: BaseNode { errors: vec![] },
                                                operator: EqualOperator,
                                                left: Mem(Box::new(MemberExpression {
                                                    base: BaseNode { errors: vec![] },
                                                    object: Idt(Identifier {
                                                        base: BaseNode { errors: vec![] },
                                                        name: "r".to_string()
                                                    }),
                                                    property: PkStr(StringLiteral {
                                                        base: BaseNode { errors: vec![] },
                                                        value: "_measurement".to_string()
                                                    })
                                                })),
                                                right: Str(StringLiteral {
                                                    base: BaseNode { errors: vec![] },
                                                    value: "b".to_string()
                                                })
                                            })))
                                        })))
                                    }]
                                }))]
                            }
                        })),
                        call: CallExpression {
                            base: BaseNode { errors: vec![] },
                            callee: Idt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "range".to_string()
                            }),
                            arguments: vec![Obj(Box::new(ObjectExpression {
                                base: BaseNode { errors: vec![] },
                                with: None,
                                properties: vec![Property {
                                    base: BaseNode { errors: vec![] },
                                    key: PkIdt(Identifier {
                                        base: BaseNode { errors: vec![] },
                                        name: "start".to_string()
                                    }),
                                    value: Some(Un(Box::new(UnaryExpression {
                                        base: BaseNode { errors: vec![] },
                                        operator: SubtractionOperator,
                                        argument: Dur(DurationLiteral {
                                            base: BaseNode { errors: vec![] },
                                            values: vec![Duration {
                                                magnitude: 1,
                                                unit: "h".to_string()
                                            }]
                                        })
                                    })))
                                }]
                            }))]
                        }
                    }))
                }),
                Expr(ExpressionStatement {
                    base: BaseNode { errors: vec![] },
                    expression: Call(Box::new(CallExpression {
                        base: BaseNode { errors: vec![] },
                        callee: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "join".to_string()
                        }),
                        arguments: vec![Obj(Box::new(ObjectExpression {
                            base: BaseNode { errors: vec![] },
                            with: None,
                            properties: vec![
                                Property {
                                    base: BaseNode { errors: vec![] },
                                    key: PkIdt(Identifier {
                                        base: BaseNode { errors: vec![] },
                                        name: "tables".to_string()
                                    }),
                                    value: Some(Arr(Box::new(ArrayExpression {
                                        base: BaseNode { errors: vec![] },
                                        elements: vec![
                                            Idt(Identifier {
                                                base: BaseNode { errors: vec![] },
                                                name: "a".to_string()
                                            }),
                                            Idt(Identifier {
                                                base: BaseNode { errors: vec![] },
                                                name: "b".to_string()
                                            })
                                        ]
                                    })))
                                },
                                Property {
                                    base: BaseNode { errors: vec![] },
                                    key: PkIdt(Identifier {
                                        base: BaseNode { errors: vec![] },
                                        name: "on".to_string()
                                    }),
                                    value: Some(Arr(Box::new(ArrayExpression {
                                        base: BaseNode { errors: vec![] },
                                        elements: vec![Str(StringLiteral {
                                            base: BaseNode { errors: vec![] },
                                            value: "t1".to_string()
                                        })]
                                    })))
                                },
                                Property {
                                    base: BaseNode { errors: vec![] },
                                    key: PkIdt(Identifier {
                                        base: BaseNode { errors: vec![] },
                                        name: "fn".to_string()
                                    }),
                                    value: Some(Fun(Box::new(FunctionExpression {
                                        base: BaseNode { errors: vec![] },
                                        params: vec![
                                            Property {
                                                base: BaseNode { errors: vec![] },
                                                key: PkIdt(Identifier {
                                                    base: BaseNode { errors: vec![] },
                                                    name: "a".to_string()
                                                }),
                                                value: None
                                            },
                                            Property {
                                                base: BaseNode { errors: vec![] },
                                                key: PkIdt(Identifier {
                                                    base: BaseNode { errors: vec![] },
                                                    name: "b".to_string()
                                                }),
                                                value: None
                                            }
                                        ],
                                        body: FExpr(Bin(Box::new(BinaryExpression {
                                            base: BaseNode { errors: vec![] },
                                            operator: DivisionOperator,
                                            left: Bin(Box::new(BinaryExpression {
                                                base: BaseNode { errors: vec![] },
                                                operator: SubtractionOperator,
                                                left: Mem(Box::new(MemberExpression {
                                                    base: BaseNode { errors: vec![] },
                                                    object: Idt(Identifier {
                                                        base: BaseNode { errors: vec![] },
                                                        name: "a".to_string()
                                                    }),
                                                    property: PkStr(StringLiteral {
                                                        base: BaseNode { errors: vec![] },
                                                        value: "_field".to_string()
                                                    })
                                                })),
                                                right: Mem(Box::new(MemberExpression {
                                                    base: BaseNode { errors: vec![] },
                                                    object: Idt(Identifier {
                                                        base: BaseNode { errors: vec![] },
                                                        name: "b".to_string()
                                                    }),
                                                    property: PkStr(StringLiteral {
                                                        base: BaseNode { errors: vec![] },
                                                        value: "_field".to_string()
                                                    })
                                                }))
                                            })),
                                            right: Mem(Box::new(MemberExpression {
                                                base: BaseNode { errors: vec![] },
                                                object: Idt(Identifier {
                                                    base: BaseNode { errors: vec![] },
                                                    name: "b".to_string()
                                                }),
                                                property: PkStr(StringLiteral {
                                                    base: BaseNode { errors: vec![] },
                                                    value: "_field".to_string()
                                                })
                                            }))
                                        })))
                                    })))
                                }
                            ]
                        }))]
                    }))
                })
            ]
        },
    )
}

// TODO(affo): the scanner fails in lexing µs.
// There is a related test in the scanner.
#[test]
#[ignore] // See https://github.com/influxdata/flux/issues/1448
fn duration_literal_all_units() {
    let mut p = Parser::new(r#"dur = 1y3mo2w1d4h1m30s1ms2µs70ns"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Var(VariableAssignment {
                base: BaseNode { errors: vec![] },
                id: Identifier {
                    base: BaseNode { errors: vec![] },
                    name: "dur".to_string()
                },
                init: Dur(DurationLiteral {
                    base: BaseNode { errors: vec![] },
                    values: vec![
                        Duration {
                            magnitude: 1,
                            unit: "y".to_string()
                        },
                        Duration {
                            magnitude: 3,
                            unit: "mo".to_string()
                        },
                        Duration {
                            magnitude: 2,
                            unit: "w".to_string()
                        },
                        Duration {
                            magnitude: 1,
                            unit: "d".to_string()
                        },
                        Duration {
                            magnitude: 4,
                            unit: "h".to_string()
                        },
                        Duration {
                            magnitude: 1,
                            unit: "m".to_string()
                        },
                        Duration {
                            magnitude: 30,
                            unit: "s".to_string()
                        },
                        Duration {
                            magnitude: 1,
                            unit: "ms".to_string()
                        },
                        Duration {
                            magnitude: 2,
                            unit: "us".to_string()
                        },
                        Duration {
                            magnitude: 70,
                            unit: "ns".to_string()
                        }
                    ]
                })
            })]
        },
    )
}

#[test]
fn duration_literal_months() {
    let mut p = Parser::new(r#"dur = 6mo"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Var(VariableAssignment {
                base: BaseNode { errors: vec![] },
                id: Identifier {
                    base: BaseNode { errors: vec![] },
                    name: "dur".to_string()
                },
                init: Dur(DurationLiteral {
                    base: BaseNode { errors: vec![] },
                    values: vec![Duration {
                        magnitude: 6,
                        unit: "mo".to_string()
                    }]
                })
            })]
        },
    )
}

#[test]
fn duration_literal_milliseconds() {
    let mut p = Parser::new(r#"dur = 500ms"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Var(VariableAssignment {
                base: BaseNode { errors: vec![] },
                id: Identifier {
                    base: BaseNode { errors: vec![] },
                    name: "dur".to_string()
                },
                init: Dur(DurationLiteral {
                    base: BaseNode { errors: vec![] },
                    values: vec![Duration {
                        magnitude: 500,
                        unit: "ms".to_string()
                    }]
                })
            })]
        },
    )
}

#[test]
fn duration_literal_months_minutes_milliseconds() {
    let mut p = Parser::new(r#"dur = 6mo30m500ms"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Var(VariableAssignment {
                base: BaseNode { errors: vec![] },
                id: Identifier {
                    base: BaseNode { errors: vec![] },
                    name: "dur".to_string()
                },
                init: Dur(DurationLiteral {
                    base: BaseNode { errors: vec![] },
                    values: vec![
                        Duration {
                            magnitude: 6,
                            unit: "mo".to_string()
                        },
                        Duration {
                            magnitude: 30,
                            unit: "m".to_string()
                        },
                        Duration {
                            magnitude: 500,
                            unit: "ms".to_string()
                        }
                    ]
                })
            })]
        },
    )
}

#[test]
fn date_literal_in_the_default_location() {
    let mut p = Parser::new(r#"now = 2018-11-29"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Var(VariableAssignment {
                base: BaseNode { errors: vec![] },
                id: Identifier {
                    base: BaseNode { errors: vec![] },
                    name: "now".to_string()
                },
                init: Time(DateTimeLiteral {
                    base: BaseNode { errors: vec![] },
                    value: DateTime::parse_from_rfc3339("2018-11-29T00:00:00Z").unwrap()
                })
            })]
        },
    )
}

#[test]
fn date_time_literal() {
    let mut p = Parser::new(r#"now = 2018-11-29T09:00:00Z"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Var(VariableAssignment {
                base: BaseNode { errors: vec![] },
                id: Identifier {
                    base: BaseNode { errors: vec![] },
                    name: "now".to_string()
                },
                init: Time(DateTimeLiteral {
                    base: BaseNode { errors: vec![] },
                    value: DateTime::parse_from_rfc3339("2018-11-29T09:00:00Z").unwrap()
                })
            })]
        },
    )
}

#[test]
fn date_time_literal_with_fractional_seconds() {
    let mut p = Parser::new(r#"now = 2018-11-29T09:00:00.100000000Z"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Var(VariableAssignment {
                base: BaseNode { errors: vec![] },
                id: Identifier {
                    base: BaseNode { errors: vec![] },
                    name: "now".to_string()
                },
                init: Time(DateTimeLiteral {
                    base: BaseNode { errors: vec![] },
                    value: DateTime::parse_from_rfc3339("2018-11-29T09:00:00.100000000Z").unwrap()
                })
            })]
        },
    )
}

#[test]
fn unary_expression_with_member_expression() {
    let mut p = Parser::new(r#"not m.b"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Un(Box::new(UnaryExpression {
                    base: BaseNode { errors: vec![] },
                    operator: NotOperator,
                    argument: Mem(Box::new(MemberExpression {
                        base: BaseNode { errors: vec![] },
                        object: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "m".to_string()
                        }),
                        property: PkIdt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "b".to_string()
                        })
                    }))
                }))
            })]
        },
    )
}

#[test]
fn function_call_with_unbalanced_braces() {
    let mut p = Parser::new(r#"from() |> range() |> map(fn: (r) => { return r._value )"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Pipe(Box::new(PipeExpression {
                    base: BaseNode { errors: vec![] },
                    argument: Pipe(Box::new(PipeExpression {
                        base: BaseNode { errors: vec![] },
                        argument: Call(Box::new(CallExpression {
                            base: BaseNode { errors: vec![] },
                            callee: Idt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "from".to_string()
                            }),
                            arguments: vec![]
                        })),
                        call: CallExpression {
                            base: BaseNode { errors: vec![] },
                            callee: Idt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "range".to_string()
                            }),
                            arguments: vec![]
                        }
                    })),
                    call: CallExpression {
                        base: BaseNode { errors: vec![] },
                        callee: Idt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "map".to_string()
                        }),
                        arguments: vec![Obj(Box::new(ObjectExpression {
                            base: BaseNode { errors: vec![] },
                            with: None,
                            properties: vec![Property {
                                base: BaseNode { errors: vec![] },
                                key: PkIdt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "fn".to_string()
                                }),
                                value: Some(Fun(Box::new(FunctionExpression {
                                    base: BaseNode { errors: vec![] },
                                    params: vec![Property {
                                        base: BaseNode { errors: vec![] },
                                        key: PkIdt(Identifier {
                                            base: BaseNode { errors: vec![] },
                                            name: "r".to_string()
                                        }),
                                        value: None
                                    }],
                                    body: FBlock(Block {
                                        base: BaseNode {
                                            errors: vec!["expected RBRACE, got RPAREN".to_string()]
                                        },
                                        body: vec![Ret(ReturnStatement {
                                            base: BaseNode { errors: vec![] },
                                            argument: Mem(Box::new(MemberExpression {
                                                base: BaseNode { errors: vec![] },
                                                object: Idt(Identifier {
                                                    base: BaseNode { errors: vec![] },
                                                    name: "r".to_string()
                                                }),
                                                property: PkIdt(Identifier {
                                                    base: BaseNode { errors: vec![] },
                                                    name: "_value".to_string()
                                                })
                                            }))
                                        })]
                                    })
                                })))
                            }]
                        }))]
                    }
                }))
            })]
        },
    )
}

#[test]
fn string_with_utf_8() {
    let mut p = Parser::new(r#""日本語""#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Str(StringLiteral {
                    base: BaseNode { errors: vec![] },
                    value: "日本語".to_string()
                })
            })]
        },
    )
}

#[test]
fn string_with_byte_values() {
    let mut p = Parser::new(r#""\xe6\x97\xa5\xe6\x9c\xac\xe8\xaa\x9e""#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Str(StringLiteral {
                    base: BaseNode { errors: vec![] },
                    value: "日本語".to_string()
                })
            })]
        },
    )
}

#[test]
fn string_with_unicode_values() {
    let mut p = Parser::new(r#""日本語""#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Str(StringLiteral {
                    base: BaseNode { errors: vec![] },
                    value: "日本語".to_string()
                })
            })]
        },
    )
}

#[test]
fn string_with_mixed_values() {
    let mut p = Parser::new(r#""hello 日x本 \xe6\x97\xa5\xe6\x9c\xac\xe8\xaa\x9e \xc2\xb5s""#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Str(StringLiteral {
                    base: BaseNode { errors: vec![] },
                    value: "hello 日x本 日本語 µs".to_string()
                })
            })]
        },
    )
}

#[test]
fn string_with_escapes() {
    let mut p = Parser::new(
        r#""newline \n
carriage return \r
horizontal tab \t
double quote \"
backslash \\
""#,
    );
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Str(StringLiteral {
                    base: BaseNode { errors: vec![] },
                    value: "newline \n\ncarriage return \r\nhorizontal tab \t\ndouble quote \"\nbackslash \\\n".to_string()
                })
            })]
        },
    )
}

#[test]
fn multiline_string() {
    let mut p = Parser::new(
        r#""
 this is a
multiline
string"
"#,
    );
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Str(StringLiteral {
                    base: BaseNode { errors: vec![] },
                    value: "\n this is a\nmultiline\nstring".to_string()
                })
            })]
        },
    )
}

#[test]
fn illegal_statement_token() {
    let mut p = Parser::new(r#"@ ident"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                BadS(BadStatement {
                    base: BaseNode {
                        errors: vec!["invalid statement @position: @".to_string()]
                    },
                    text: "@".to_string()
                }),
                Expr(ExpressionStatement {
                    base: BaseNode { errors: vec![] },
                    expression: Idt(Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "ident".to_string()
                    })
                })
            ]
        },
    )
}

// TODO(affo): that error is injected by ast.Check().
#[test]
fn missing_left_hand_side() {
    let mut p = Parser::new(r#"(*b)"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Bin(Box::new(BinaryExpression {
                    // TODO(affo): this should be like this:
                    // base: BaseNode { errors: vec!["missing left hand side of expression".to_string()] },
                    base: BaseNode { errors: vec![] },
                    operator: MultiplicationOperator,
                    left: BadE(Box::new(BadExpression {
                        base: BaseNode { errors: vec![] },
                        text: "invalid token for primary expression: MUL".to_string(),
                        expression: None
                    })),
                    right: Idt(Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "b".to_string()
                    })
                }))
            })]
        },
    )
}

// TODO(affo): that error is injected by ast.Check().
#[test]
fn missing_right_hand_side() {
    let mut p = Parser::new(r#"(a*)"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Bin(Box::new(BinaryExpression {
                    // TODO(affo): this should be like this:
                    // base: BaseNode { errors: vec!["missing right hand side of expression".to_string()] },
                    base: BaseNode { errors: vec![] },
                    operator: MultiplicationOperator,
                    right: BadE(Box::new(BadExpression {
                        base: BaseNode { errors: vec![] },
                        text: "invalid token for primary expression: RPAREN".to_string(),
                        expression: None
                    })),
                    left: Idt(Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "a".to_string()
                    })
                }))
            })]
        },
    )
}

// NOTE(affo): this is slightly different from Go. We have a BadExpression in the body.
#[test]
fn missing_arrow_in_function_expression() {
    let mut p = Parser::new(r#"(a, b) a + b"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Fun(Box::new(FunctionExpression {
                    base: BaseNode {
                        errors: vec![
                            "expected ARROW, got IDENT (a) at position".to_string(),
                            "expected ARROW, got ADD (+) at position".to_string(),
                            "expected ARROW, got IDENT (b) at position".to_string(),
                            "expected ARROW, got EOF".to_string()
                        ]
                    },
                    params: vec![
                        Property {
                            base: BaseNode { errors: vec![] },
                            key: PkIdt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "a".to_string()
                            }),
                            value: None
                        },
                        Property {
                            base: BaseNode { errors: vec![] },
                            key: PkIdt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "b".to_string()
                            }),
                            value: None
                        }
                    ],
                    body: FExpr(BadE(Box::new(BadExpression {
                        base: BaseNode { errors: vec![] },
                        text: "invalid token for primary expression: EOF".to_string(),
                        expression: None
                    })))
                }))
            })]
        },
    )
}

#[test]
fn property_list_missing_property() {
    let mut p = Parser::new(r#"o = {a: "a",, b: 7}"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Var(VariableAssignment {
                base: BaseNode { errors: vec![] },
                id: Identifier {
                    base: BaseNode { errors: vec![] },
                    name: "o".to_string()
                },
                init: Obj(Box::new(ObjectExpression {
                    base: BaseNode { errors: vec![] },
                    with: None,
                    properties: vec![
                        Property {
                            base: BaseNode { errors: vec![] },
                            key: PkIdt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "a".to_string()
                            }),
                            value: Some(Str(StringLiteral {
                                base: BaseNode { errors: vec![] },
                                value: "a".to_string()
                            }))
                        },
                        Property {
                            base: BaseNode {
                                errors: vec!["missing property in property list".to_string()]
                            },
                            key: PkStr(StringLiteral {
                                base: BaseNode { errors: vec![] },
                                value: "<invalid>".to_string()
                            }),
                            value: None
                        },
                        Property {
                            base: BaseNode { errors: vec![] },
                            key: PkIdt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "b".to_string()
                            }),
                            value: Some(Int(IntegerLiteral {
                                base: BaseNode { errors: vec![] },
                                value: 7
                            }))
                        }
                    ]
                }))
            })]
        },
    )
}

#[test]
fn property_list_missing_key() {
    let mut p = Parser::new(r#"o = {: "a"}"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Var(VariableAssignment {
                base: BaseNode { errors: vec![] },
                id: Identifier {
                    base: BaseNode { errors: vec![] },
                    name: "o".to_string()
                },
                init: Obj(Box::new(ObjectExpression {
                    base: BaseNode { errors: vec![] },
                    with: None,
                    properties: vec![Property {
                        base: BaseNode {
                            errors: vec!["missing property key".to_string()]
                        },
                        key: PkStr(StringLiteral {
                            base: BaseNode { errors: vec![] },
                            value: "<invalid>".to_string()
                        }),
                        value: Some(Str(StringLiteral {
                            base: BaseNode { errors: vec![] },
                            value: "a".to_string()
                        }))
                    }]
                }))
            })]
        },
    )
}

#[test]
fn property_list_missing_value() {
    let mut p = Parser::new(r#"o = {a:}"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Var(VariableAssignment {
                base: BaseNode { errors: vec![] },
                id: Identifier {
                    base: BaseNode { errors: vec![] },
                    name: "o".to_string()
                },
                init: Obj(Box::new(ObjectExpression {
                    base: BaseNode { errors: vec![] },
                    with: None,
                    properties: vec![Property {
                        base: BaseNode {
                            errors: vec!["missing property value".to_string()]
                        },
                        key: PkIdt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "a".to_string()
                        }),
                        value: None
                    }]
                }))
            })]
        },
    )
}

// TODO(affo): that error is injected by ast.Check().
#[test]
fn property_list_missing_comma() {
    let mut p = Parser::new(r#"o = {a: "a" b: 30}"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Var(VariableAssignment {
                base: BaseNode { errors: vec![] },
                id: Identifier {
                    base: BaseNode { errors: vec![] },
                    name: "o".to_string()
                },
                init: Obj(Box::new(ObjectExpression {
                    base: BaseNode { errors: vec![] },
                    with: None,
                    properties: vec![
                        Property {
                            base: BaseNode { errors: vec![] },
                            key: PkIdt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "a".to_string()
                            }),
                            // TODO(affo): ast.Check would add the error "expected an operator between two expressions".
                            value: Some(Bin(Box::new(BinaryExpression {
                                base: BaseNode { errors: vec![] },
                                operator: InvalidOperator,
                                left: Str(StringLiteral {
                                    base: BaseNode { errors: vec![] },
                                    value: "a".to_string()
                                }),
                                right: Idt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "b".to_string()
                                })
                            })))
                        },
                        Property {
                            base: BaseNode {
                                errors: vec!["missing property key".to_string()]
                            },
                            key: PkStr(StringLiteral {
                                base: BaseNode { errors: vec![] },
                                value: "<invalid>".to_string()
                            }),
                            value: Some(Int(IntegerLiteral {
                                base: BaseNode {
                                    errors: vec![
                                        "expected comma in property list, got COLON".to_string()
                                    ]
                                },
                                value: 30
                            }))
                        }
                    ]
                }))
            })]
        },
    )
}

#[test]
fn property_list_trailing_comma() {
    let mut p = Parser::new(r#"o = {a: "a",}"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Var(VariableAssignment {
                base: BaseNode { errors: vec![] },
                id: Identifier {
                    base: BaseNode { errors: vec![] },
                    name: "o".to_string()
                },
                init: Obj(Box::new(ObjectExpression {
                    base: BaseNode { errors: vec![] },
                    with: None,
                    properties: vec![Property {
                        base: BaseNode { errors: vec![] },
                        key: PkIdt(Identifier {
                            base: BaseNode { errors: vec![] },
                            name: "a".to_string()
                        }),
                        value: Some(Str(StringLiteral {
                            base: BaseNode { errors: vec![] },
                            value: "a".to_string()
                        }))
                    }]
                }))
            })]
        },
    )
}

#[test]
fn property_list_bad_property() {
    let mut p = Parser::new(r#"o = {a: "a", 30, b: 7}"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Var(VariableAssignment {
                base: BaseNode { errors: vec![] },
                id: Identifier {
                    base: BaseNode { errors: vec![] },
                    name: "o".to_string()
                },
                init: Obj(Box::new(ObjectExpression {
                    base: BaseNode { errors: vec![] },
                    with: None,
                    properties: vec![
                        Property {
                            base: BaseNode { errors: vec![] },
                            key: PkIdt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "a".to_string()
                            }),
                            value: Some(Str(StringLiteral {
                                base: BaseNode { errors: vec![] },
                                value: "a".to_string()
                            }))
                        },
                        Property {
                            base: BaseNode {
                                errors: vec![
                                    "unexpected token for property key: INT (30)".to_string()
                                ]
                            },
                            key: PkStr(StringLiteral {
                                base: BaseNode { errors: vec![] },
                                value: "<invalid>".to_string()
                            }),
                            value: None
                        },
                        Property {
                            base: BaseNode { errors: vec![] },
                            key: PkIdt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "b".to_string()
                            }),
                            value: Some(Int(IntegerLiteral {
                                base: BaseNode { errors: vec![] },
                                value: 7
                            }))
                        }
                    ]
                }))
            })]
        },
    )
}

// TODO(jsternberg): This should fill in error nodes.
// The current behavior is non-sensical.
#[test]
fn invalid_expression_in_array() {
    let mut p = Parser::new(r#"['a']"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Arr(Box::new(ArrayExpression {
                    base: BaseNode { errors: vec![] },
                    elements: vec![Idt(Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "a".to_string()
                    })]
                }))
            })]
        },
    )
}

#[test]
fn integer_literal_overflow() {
    let mut p = Parser::new(r#"100000000000000000000000000000"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Int(IntegerLiteral {
                    base: BaseNode { errors: vec!["invalid integer literal \"100000000000000000000000000000\": value out of range".to_string()] },
                    value: 0,
                })
            })]
        },
    )
}

// TODO(affo): that error is injected by ast.Check().
#[test]
fn multiple_idents_in_parens() {
    let mut p = Parser::new(r#"(a b)"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Bin(Box::new(BinaryExpression {
                    // TODO(affo): ast.Check would add the error "expected an operator between two expressions".
                    base: BaseNode { errors: vec![] },
                    operator: InvalidOperator,
                    left: Idt(Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "a".to_string()
                    }),
                    right: Idt(Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "b".to_string()
                    })
                }))
            })]
        },
    )
}

#[test]
fn bad_regex_literal() {
    let mut p = Parser::new(r#"/*/"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Regexp(RegexpLiteral {
                    base: BaseNode {
                        errors: vec![
                            "regex parse error: * error: repetition operator missing expression"
                                .to_string()
                        ]
                    },
                    value: "".to_string()
                })
            })]
        },
    )
}

// TODO(affo): that error is injected by ast.Check().
#[test]
fn implicit_key_object_literal_error() {
    let mut p = Parser::new(r#"x = {"a", b}"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Var(VariableAssignment {
                base: BaseNode { errors: vec![] },
                id: Identifier {
                    base: BaseNode { errors: vec![] },
                    name: "x".to_string()
                },
                init: Obj(Box::new(ObjectExpression {
                    base: BaseNode { errors: vec![] },
                    with: None,
                    properties: vec![
                        Property {
                            // TODO(affo): this should error with ast.Check: "string literal key "a" must have a value".
                            base: BaseNode { errors: vec![] },
                            key: PkStr(StringLiteral {
                                base: BaseNode { errors: vec![] },
                                value: "a".to_string()
                            }),
                            value: None
                        },
                        Property {
                            base: BaseNode { errors: vec![] },
                            key: PkIdt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "b".to_string()
                            }),
                            value: None
                        }
                    ]
                }))
            })]
        },
    )
}

#[test]
fn object_with() {
    let mut p = Parser::new(r#"{a with b:c, d:e}"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Obj(Box::new(ObjectExpression {
                    base: BaseNode { errors: vec![] },
                    with: Some(Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "a".to_string()
                    }),
                    properties: vec![
                        Property {
                            base: BaseNode { errors: vec![] },
                            key: PkIdt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "b".to_string()
                            }),
                            value: Some(Idt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "c".to_string()
                            }))
                        },
                        Property {
                            base: BaseNode { errors: vec![] },
                            key: PkIdt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "d".to_string()
                            }),
                            value: Some(Idt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "e".to_string()
                            }))
                        }
                    ]
                }))
            })]
        },
    )
}

#[test]
fn object_with_implicit_keys() {
    let mut p = Parser::new(r#"{a with b, c}"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Expr(ExpressionStatement {
                base: BaseNode { errors: vec![] },
                expression: Obj(Box::new(ObjectExpression {
                    base: BaseNode { errors: vec![] },
                    with: Some(Identifier {
                        base: BaseNode { errors: vec![] },
                        name: "a".to_string()
                    }),
                    properties: vec![
                        Property {
                            base: BaseNode { errors: vec![] },
                            key: PkIdt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "b".to_string()
                            }),
                            value: None
                        },
                        Property {
                            base: BaseNode { errors: vec![] },
                            key: PkIdt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "c".to_string()
                            }),
                            value: None
                        }
                    ]
                }))
            })]
        },
    )
}

#[test]
fn conditional_with_unary_logical_operators() {
    let mut p = Parser::new(
        r#"a = if exists b or c < d and not e == f then not exists (g - h) else exists exists i"#,
    );
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Var(VariableAssignment {
                base: BaseNode { errors: vec![] },
                id: Identifier {
                    base: BaseNode { errors: vec![] },
                    name: "a".to_string()
                },
                init: Cond(Box::new(ConditionalExpression {
                    base: BaseNode { errors: vec![] },
                    test: Log(Box::new(LogicalExpression {
                        base: BaseNode { errors: vec![] },
                        operator: OrOperator,
                        left: Un(Box::new(UnaryExpression {
                            base: BaseNode { errors: vec![] },
                            operator: ExistsOperator,
                            argument: Idt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "b".to_string()
                            })
                        })),
                        right: Log(Box::new(LogicalExpression {
                            base: BaseNode { errors: vec![] },
                            operator: AndOperator,
                            left: Bin(Box::new(BinaryExpression {
                                base: BaseNode { errors: vec![] },
                                operator: LessThanOperator,
                                left: Idt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "c".to_string()
                                }),
                                right: Idt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "d".to_string()
                                })
                            })),
                            right: Un(Box::new(UnaryExpression {
                                base: BaseNode { errors: vec![] },
                                operator: NotOperator,
                                argument: Bin(Box::new(BinaryExpression {
                                    base: BaseNode { errors: vec![] },
                                    operator: EqualOperator,
                                    left: Idt(Identifier {
                                        base: BaseNode { errors: vec![] },
                                        name: "e".to_string()
                                    }),
                                    right: Idt(Identifier {
                                        base: BaseNode { errors: vec![] },
                                        name: "f".to_string()
                                    })
                                }))
                            }))
                        }))
                    })),
                    consequent: Un(Box::new(UnaryExpression {
                        base: BaseNode { errors: vec![] },
                        operator: NotOperator,
                        argument: Un(Box::new(UnaryExpression {
                            base: BaseNode { errors: vec![] },
                            operator: ExistsOperator,
                            argument: Bin(Box::new(BinaryExpression {
                                base: BaseNode { errors: vec![] },
                                operator: SubtractionOperator,
                                left: Idt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "g".to_string()
                                }),
                                right: Idt(Identifier {
                                    base: BaseNode { errors: vec![] },
                                    name: "h".to_string()
                                })
                            }))
                        }))
                    })),
                    alternate: Un(Box::new(UnaryExpression {
                        base: BaseNode { errors: vec![] },
                        operator: ExistsOperator,
                        argument: Un(Box::new(UnaryExpression {
                            base: BaseNode { errors: vec![] },
                            operator: ExistsOperator,
                            argument: Idt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "i".to_string()
                            })
                        }))
                    }))
                }))
            })]
        },
    )
}

// TODO(affo): that error is injected by ast.Check().
#[test]
fn implicit_and_explicit_keys_object_literal_error() {
    let mut p = Parser::new(r#"x = {a, b:c}"#);
    let parsed = p.parse_file("".to_string());
    assert_eq!(
        parsed,
        File {
            base: BaseNode { errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Var(VariableAssignment {
                base: BaseNode { errors: vec![] },
                id: Identifier {
                    base: BaseNode { errors: vec![] },
                    name: "x".to_string()
                },
                init: Obj(Box::new(ObjectExpression {
                    // TODO(affo): this should error in ast.Check(): "cannot mix implicit and explicit properties".
                    base: BaseNode { errors: vec![] },
                    with: None,
                    properties: vec![
                        Property {
                            base: BaseNode { errors: vec![] },
                            key: PkIdt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "a".to_string()
                            }),
                            value: None
                        },
                        Property {
                            base: BaseNode { errors: vec![] },
                            key: PkIdt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "b".to_string()
                            }),
                            value: Some(Idt(Identifier {
                                base: BaseNode { errors: vec![] },
                                name: "c".to_string()
                            }))
                        }
                    ]
                }))
            })]
        },
    )
}
