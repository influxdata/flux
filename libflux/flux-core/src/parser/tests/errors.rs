use pretty_assertions::assert_eq;

use super::*;
use crate::locator::Locator;

#[test]
fn function_call_with_unbalanced_braces() {
    let mut p = Parser::new(r#"from() |> range() |> map(fn: (r) => { return r._value )"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 56),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 56),
                    ..BaseNode::default()
                },
                expression: Expression::PipeExpr(Box::new(PipeExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 56),
                        ..BaseNode::default()
                    },
                    argument: Expression::PipeExpr(Box::new(PipeExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 18),
                            ..BaseNode::default()
                        },
                        argument: Expression::Call(Box::new(CallExpr {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 7),
                                ..BaseNode::default()
                            },
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 1, 1, 5),
                                    ..BaseNode::default()
                                },
                                name: "from".to_string()
                            }),
                            lparen: vec![],
                            arguments: vec![],
                            rparen: vec![],
                        })),
                        call: CallExpr {
                            base: BaseNode {
                                location: loc.get(1, 11, 1, 18),
                                ..BaseNode::default()
                            },
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 11, 1, 16),
                                    ..BaseNode::default()
                                },
                                name: "range".to_string()
                            }),
                            lparen: vec![],
                            arguments: vec![],
                            rparen: vec![],
                        }
                    })),
                    call: CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 22, 1, 56),
                            ..BaseNode::default()
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 22, 1, 25),
                                ..BaseNode::default()
                            },
                            name: "map".to_string()
                        }),
                        lparen: vec![],
                        arguments: vec![Expression::Object(Box::new(ObjectExpr {
                            base: BaseNode {
                                location: loc.get(1, 26, 1, 56),
                                ..BaseNode::default()
                            },
                            lbrace: vec![],
                            with: None,
                            properties: vec![Property {
                                base: BaseNode {
                                    location: loc.get(1, 26, 1, 56),
                                    ..BaseNode::default()
                                },
                                key: PropertyKey::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 26, 1, 28),
                                        ..BaseNode::default()
                                    },
                                    name: "fn".to_string()
                                }),
                                separator: vec![],
                                value: Some(Expression::Function(Box::new(FunctionExpr {
                                    base: BaseNode {
                                        location: loc.get(1, 30, 1, 56),
                                        ..BaseNode::default()
                                    },
                                    lparen: vec![],
                                    params: vec![Property {
                                        base: BaseNode {
                                            location: loc.get(1, 31, 1, 32),
                                            ..BaseNode::default()
                                        },
                                        key: PropertyKey::Identifier(Identifier {
                                            base: BaseNode {
                                                location: loc.get(1, 31, 1, 32),
                                                ..BaseNode::default()
                                            },
                                            name: "r".to_string()
                                        }),
                                        separator: vec![],
                                        value: None,
                                        comma: vec![],
                                    }],
                                    rparen: vec![],
                                    arrow: vec![],
                                    body: FunctionBody::Block(Block {
                                        base: BaseNode {
                                            location: loc.get(1, 37, 1, 56),
                                            errors: vec!["expected RBRACE, got RPAREN".to_string()],
                                            ..BaseNode::default()
                                        },
                                        lbrace: vec![],
                                        body: vec![Statement::Return(Box::new(ReturnStmt {
                                            base: BaseNode {
                                                location: loc.get(1, 39, 1, 54),
                                                ..BaseNode::default()
                                            },
                                            argument: Expression::Member(Box::new(MemberExpr {
                                                base: BaseNode {
                                                    location: loc.get(1, 46, 1, 54),
                                                    ..BaseNode::default()
                                                },
                                                object: Expression::Identifier(Identifier {
                                                    base: BaseNode {
                                                        location: loc.get(1, 46, 1, 47),
                                                        ..BaseNode::default()
                                                    },
                                                    name: "r".to_string()
                                                }),
                                                lbrack: vec![],
                                                property: PropertyKey::Identifier(Identifier {
                                                    base: BaseNode {
                                                        location: loc.get(1, 48, 1, 54),
                                                        ..BaseNode::default()
                                                    },
                                                    name: "_value".to_string()
                                                }),
                                                rbrack: vec![],
                                            }))
                                        }))],
                                        rbrace: vec![],
                                    }),
                                }))),
                                comma: vec![],
                            }],
                            rbrace: vec![],
                        }))],
                        rparen: vec![],
                    }
                }))
            }))],
            eof: vec![],
        },
    )
}

// TODO(affo): that error is injected by ast.Check().
#[test]
fn illegal_statement_token() {
    let mut p = Parser::new(r#"@ ident"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 8),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Statement::Bad(Box::new(BadStmt {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 2),
                        // errors: vec!["invalid statement @1:1-1:2: @".to_string()]
                        ..BaseNode::default()
                    },
                    text: "@".to_string()
                })),
                Statement::Expr(Box::new(ExprStmt {
                    base: BaseNode {
                        location: loc.get(1, 3, 1, 8),
                        ..BaseNode::default()
                    },
                    expression: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 3, 1, 8),
                            ..BaseNode::default()
                        },
                        name: "ident".to_string()
                    })
                }))
            ],
            eof: vec![],
        },
    )
}

// TODO(affo): that error is injected by ast.Check().
#[test]
fn multiple_idents_in_parens() {
    let mut p = Parser::new(r#"(a b)"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 6),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 6),
                    ..BaseNode::default()
                },
                expression: Expression::Paren(Box::new(ParenExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 6),
                        ..BaseNode::default()
                    },
                    lparen: vec![],
                    expression: Expression::Binary(Box::new(BinaryExpr {
                        // TODO(affo): ast.Check would add the error "expected an operator between two expressions".
                        base: BaseNode {
                            location: loc.get(1, 2, 1, 5),
                            ..BaseNode::default()
                        },
                        operator: Operator::InvalidOperator,
                        left: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 2, 1, 3),
                                ..BaseNode::default()
                            },
                            name: "a".to_string()
                        }),
                        right: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 4, 1, 5),
                                ..BaseNode::default()
                            },
                            name: "b".to_string()
                        })
                    })),
                    rparen: vec![],
                }))
            }))],
            eof: vec![],
        },
    )
}

// TODO(affo): that error is injected by ast.Check().
// TODO(jsternberg): Parens aren't recorded correctly in the source and are mostly ignored.
#[test]
fn missing_left_hand_side() {
    let mut p = Parser::new(r#"(*b)"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 5),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 5),
                    ..BaseNode::default()
                },
                expression: Expression::Paren(Box::new(ParenExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 5),
                        ..BaseNode::default()
                    },
                    lparen: vec![],
                    expression: Expression::Binary(Box::new(BinaryExpr {
                        // TODO(affo): this should be like this:
                        // base: BaseNode {location: ..., errors: vec!["missing left hand side of expression".to_string()] },
                        base: BaseNode {
                            location: loc.get(1, 2, 1, 4),
                            ..BaseNode::default()
                        },
                        operator: Operator::MultiplicationOperator,
                        left: Expression::Bad(Box::new(BadExpr {
                            base: BaseNode {
                                location: loc.get(1, 2, 1, 3),
                                ..BaseNode::default()
                            },
                            text: "invalid token for primary expression: MUL".to_string(),
                            expression: None
                        })),
                        right: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 3, 1, 4),
                                ..BaseNode::default()
                            },
                            name: "b".to_string()
                        })
                    })),
                    rparen: vec![],
                }))
            }))],
            eof: vec![],
        },
    )
}

// TODO(affo): that error is injected by ast.Check().
// TODO(jsternberg): Parens aren't recorded correctly in the source and are mostly ignored.
#[test]
fn missing_right_hand_side() {
    let mut p = Parser::new(r#"(a*)"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 5),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 5),
                    ..BaseNode::default()
                },
                expression: Expression::Paren(Box::new(ParenExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 5),
                        ..BaseNode::default()
                    },
                    lparen: vec![],
                    expression: Expression::Binary(Box::new(BinaryExpr {
                        // TODO(affo): this should be like this:
                        // base: BaseNode {location: ..., errors: vec!["missing right hand side of expression".to_string()] },
                        base: BaseNode {
                            location: loc.get(1, 2, 1, 5),
                            ..BaseNode::default()
                        },
                        operator: Operator::MultiplicationOperator,
                        left: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 2, 1, 3),
                                ..BaseNode::default()
                            },
                            name: "a".to_string()
                        }),
                        right: Expression::Bad(Box::new(BadExpr {
                            base: BaseNode {
                                location: loc.get(1, 4, 1, 5),
                                ..BaseNode::default()
                            },
                            text: "invalid token for primary expression: RPAREN".to_string(),
                            expression: None
                        })),
                    })),
                    rparen: vec![],
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn illegal_expression() {
    let mut p = Parser::new(r#"(@)"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 4),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 4),
                    ..BaseNode::default()
                },
                expression: Expression::Paren(Box::new(ParenExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 4),
                        errors: vec!["invalid expression @1:2-1:3: @".to_string()],
                        ..BaseNode::default()
                    },
                    lparen: vec![],
                    expression: Expression::Bad(Box::new(BadExpr {
                        base: BaseNode {
                            location: loc.get(1, 2, 1, 3),
                            ..BaseNode::default()
                        },
                        text: "@".to_string(),
                        expression: None
                    })),
                    rparen: vec![],
                }))
            }))],
            eof: vec![],
        },
    )
}

// NOTE(affo): this is slightly different from Go. We have a BadExpr in the body.
#[test]
fn missing_arrow_in_function_expression() {
    let mut p = Parser::new(r#"(a, b) a + b"#);
    let parsed = p.parse_file("".to_string());
    expect_test::expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    file: Some(
                        "",
                    ),
                    start: Position {
                        line: 1,
                        column: 1,
                    },
                    end: Position {
                        line: 1,
                        column: 13,
                    },
                },
                comments: [],
                errors: [],
            },
            name: "",
            metadata: "parser-type=rust",
            package: None,
            imports: [],
            body: [
                Expr(
                    ExprStmt {
                        base: BaseNode {
                            location: SourceLocation {
                                file: Some(
                                    "",
                                ),
                                start: Position {
                                    line: 1,
                                    column: 1,
                                },
                                end: Position {
                                    line: 1,
                                    column: 13,
                                },
                            },
                            comments: [],
                            errors: [],
                        },
                        expression: Function(
                            FunctionExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        file: Some(
                                            "",
                                        ),
                                        start: Position {
                                            line: 1,
                                            column: 1,
                                        },
                                        end: Position {
                                            line: 1,
                                            column: 13,
                                        },
                                    },
                                    comments: [],
                                    errors: [],
                                },
                                lparen: [],
                                params: [
                                    Property {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                file: Some(
                                                    "",
                                                ),
                                                start: Position {
                                                    line: 1,
                                                    column: 2,
                                                },
                                                end: Position {
                                                    line: 1,
                                                    column: 3,
                                                },
                                            },
                                            comments: [],
                                            errors: [],
                                        },
                                        key: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        file: Some(
                                                            "",
                                                        ),
                                                        start: Position {
                                                            line: 1,
                                                            column: 2,
                                                        },
                                                        end: Position {
                                                            line: 1,
                                                            column: 3,
                                                        },
                                                    },
                                                    comments: [],
                                                    errors: [],
                                                },
                                                name: "a",
                                            },
                                        ),
                                        separator: [],
                                        value: None,
                                        comma: [],
                                    },
                                    Property {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                file: Some(
                                                    "",
                                                ),
                                                start: Position {
                                                    line: 1,
                                                    column: 5,
                                                },
                                                end: Position {
                                                    line: 1,
                                                    column: 6,
                                                },
                                            },
                                            comments: [],
                                            errors: [],
                                        },
                                        key: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        file: Some(
                                                            "",
                                                        ),
                                                        start: Position {
                                                            line: 1,
                                                            column: 5,
                                                        },
                                                        end: Position {
                                                            line: 1,
                                                            column: 6,
                                                        },
                                                    },
                                                    comments: [],
                                                    errors: [],
                                                },
                                                name: "b",
                                            },
                                        ),
                                        separator: [],
                                        value: None,
                                        comma: [],
                                    },
                                ],
                                rparen: [],
                                arrow: [],
                                body: Expr(
                                    Binary(
                                        BinaryExpr {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    file: Some(
                                                        "",
                                                    ),
                                                    start: Position {
                                                        line: 1,
                                                        column: 8,
                                                    },
                                                    end: Position {
                                                        line: 1,
                                                        column: 13,
                                                    },
                                                },
                                                comments: [],
                                                errors: [],
                                            },
                                            operator: AdditionOperator,
                                            left: Identifier(
                                                Identifier {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            file: Some(
                                                                "",
                                                            ),
                                                            start: Position {
                                                                line: 1,
                                                                column: 8,
                                                            },
                                                            end: Position {
                                                                line: 1,
                                                                column: 9,
                                                            },
                                                        },
                                                        comments: [],
                                                        errors: [
                                                            "expected ARROW, got IDENT (a) at 1:8",
                                                        ],
                                                    },
                                                    name: "a",
                                                },
                                            ),
                                            right: Identifier(
                                                Identifier {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            file: Some(
                                                                "",
                                                            ),
                                                            start: Position {
                                                                line: 1,
                                                                column: 12,
                                                            },
                                                            end: Position {
                                                                line: 1,
                                                                column: 13,
                                                            },
                                                        },
                                                        comments: [],
                                                        errors: [],
                                                    },
                                                    name: "b",
                                                },
                                            ),
                                        },
                                    ),
                                ),
                            },
                        ),
                    },
                ),
            ],
            eof: [],
        }
    "#]]
    .assert_debug_eq(&parsed);
}

#[test]
fn index_with_unclosed_bracket() {
    let mut p = Parser::new(r#"a[b()"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 6),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 6),
                    ..BaseNode::default()
                },
                expression: Expression::Index(Box::new(IndexExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 6),
                        errors: vec!["expected RBRACK, got EOF".to_string()],
                        ..BaseNode::default()
                    },
                    array: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            ..BaseNode::default()
                        },
                        name: "a".to_string()
                    }),
                    lbrack: vec![],
                    index: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 3, 1, 6),
                            ..BaseNode::default()
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 3, 1, 4),
                                ..BaseNode::default()
                            },
                            name: "b".to_string()
                        }),
                        lparen: vec![],
                        arguments: vec![],
                        rparen: vec![],
                    })),
                    rbrack: vec![],
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn index_with_unbalanced_parenthesis() {
    let mut p = Parser::new(r#"a[b(]"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 6),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 6),
                    ..BaseNode::default()
                },
                expression: Expression::Index(Box::new(IndexExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 6),
                        ..BaseNode::default()
                    },
                    array: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            ..BaseNode::default()
                        },
                        name: "a".to_string()
                    }),
                    lbrack: vec![],
                    index: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 3, 1, 6),
                            errors: vec!["expected RPAREN, got RBRACK".to_string()],
                            ..BaseNode::default()
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 3, 1, 4),
                                ..BaseNode::default()
                            },
                            name: "b".to_string()
                        }),
                        lparen: vec![],
                        arguments: vec![],
                        rparen: vec![],
                    })),
                    rbrack: vec![],
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn index_with_unexpected_rparen() {
    let mut p = Parser::new(r#"a[b)]"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 6),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 6),
                    ..BaseNode::default()
                },
                expression: Expression::Index(Box::new(IndexExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 6),
                        errors: vec!["invalid expression @1:4-1:5: )".to_string()],
                        ..BaseNode::default()
                    },
                    array: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            ..BaseNode::default()
                        },
                        name: "a".to_string()
                    }),
                    lbrack: vec![],
                    index: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 3, 1, 4),
                            ..BaseNode::default()
                        },
                        name: "b".to_string()
                    }),
                    rbrack: vec![],
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn issue_4231() {
    let mut p = Parser::new(
        r#"
map(fn: (r) => ({ r with _value: if true and false then 1}) )
"#,
    );
    let parsed = p.parse_file("".to_string());
    expect_test::expect![[r#"error @2:34-2:59: expected ELSE, got RBRACE (}) at 2:58"#]].assert_eq(
        &ast::check::check(ast::walk::Node::File(&parsed))
            .unwrap_err()
            .to_string(),
    );
}
