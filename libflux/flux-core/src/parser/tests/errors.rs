use pretty_assertions::assert_eq;

use super::*;

#[test]
fn function_call_with_unbalanced_braces() {
    let mut p = Parser::new(r#"from() |> range() |> map(fn: (r) => { return r._value )"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 56",
                    source: "from() |> range() |> map(fn: (r) => { return r._value )",
                },
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
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 56",
                                source: "from() |> range() |> map(fn: (r) => { return r._value )",
                            },
                        },
                        expression: PipeExpr(
                            PipeExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 56",
                                        source: "from() |> range() |> map(fn: (r) => { return r._value )",
                                    },
                                },
                                argument: PipeExpr(
                                    PipeExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 1",
                                                end: "line: 1, column: 18",
                                                source: "from() |> range()",
                                            },
                                        },
                                        argument: Call(
                                            CallExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 1",
                                                        end: "line: 1, column: 7",
                                                        source: "from()",
                                                    },
                                                },
                                                callee: Identifier(
                                                    Identifier {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 1, column: 1",
                                                                end: "line: 1, column: 5",
                                                                source: "from",
                                                            },
                                                        },
                                                        name: "from",
                                                    },
                                                ),
                                                lparen: [],
                                                arguments: [],
                                                rparen: [],
                                            },
                                        ),
                                        call: CallExpr {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 11",
                                                    end: "line: 1, column: 18",
                                                    source: "range()",
                                                },
                                            },
                                            callee: Identifier(
                                                Identifier {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 11",
                                                            end: "line: 1, column: 16",
                                                            source: "range",
                                                        },
                                                    },
                                                    name: "range",
                                                },
                                            ),
                                            lparen: [],
                                            arguments: [],
                                            rparen: [],
                                        },
                                    },
                                ),
                                call: CallExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 22",
                                            end: "line: 1, column: 56",
                                            source: "map(fn: (r) => { return r._value )",
                                        },
                                    },
                                    callee: Identifier(
                                        Identifier {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 22",
                                                    end: "line: 1, column: 25",
                                                    source: "map",
                                                },
                                            },
                                            name: "map",
                                        },
                                    ),
                                    lparen: [],
                                    arguments: [
                                        Object(
                                            ObjectExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 26",
                                                        end: "line: 1, column: 56",
                                                        source: "fn: (r) => { return r._value )",
                                                    },
                                                },
                                                lbrace: [],
                                                with: None,
                                                properties: [
                                                    Property {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 1, column: 26",
                                                                end: "line: 1, column: 56",
                                                                source: "fn: (r) => { return r._value )",
                                                            },
                                                        },
                                                        key: Identifier(
                                                            Identifier {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 1, column: 26",
                                                                        end: "line: 1, column: 28",
                                                                        source: "fn",
                                                                    },
                                                                },
                                                                name: "fn",
                                                            },
                                                        ),
                                                        separator: [],
                                                        value: Some(
                                                            Function(
                                                                FunctionExpr {
                                                                    base: BaseNode {
                                                                        location: SourceLocation {
                                                                            start: "line: 1, column: 30",
                                                                            end: "line: 1, column: 56",
                                                                            source: "(r) => { return r._value )",
                                                                        },
                                                                    },
                                                                    lparen: [],
                                                                    params: [
                                                                        Property {
                                                                            base: BaseNode {
                                                                                location: SourceLocation {
                                                                                    start: "line: 1, column: 31",
                                                                                    end: "line: 1, column: 32",
                                                                                    source: "r",
                                                                                },
                                                                            },
                                                                            key: Identifier(
                                                                                Identifier {
                                                                                    base: BaseNode {
                                                                                        location: SourceLocation {
                                                                                            start: "line: 1, column: 31",
                                                                                            end: "line: 1, column: 32",
                                                                                            source: "r",
                                                                                        },
                                                                                    },
                                                                                    name: "r",
                                                                                },
                                                                            ),
                                                                            separator: [],
                                                                            value: None,
                                                                            comma: [],
                                                                        },
                                                                    ],
                                                                    rparen: [],
                                                                    arrow: [],
                                                                    body: Block(
                                                                        Block {
                                                                            base: BaseNode {
                                                                                location: SourceLocation {
                                                                                    start: "line: 1, column: 37",
                                                                                    end: "line: 1, column: 56",
                                                                                    source: "{ return r._value )",
                                                                                },
                                                                                errors: [
                                                                                    "expected RBRACE, got RPAREN",
                                                                                ],
                                                                            },
                                                                            lbrace: [],
                                                                            body: [
                                                                                Return(
                                                                                    ReturnStmt {
                                                                                        base: BaseNode {
                                                                                            location: SourceLocation {
                                                                                                start: "line: 1, column: 39",
                                                                                                end: "line: 1, column: 54",
                                                                                                source: "return r._value",
                                                                                            },
                                                                                        },
                                                                                        argument: Member(
                                                                                            MemberExpr {
                                                                                                base: BaseNode {
                                                                                                    location: SourceLocation {
                                                                                                        start: "line: 1, column: 46",
                                                                                                        end: "line: 1, column: 54",
                                                                                                        source: "r._value",
                                                                                                    },
                                                                                                },
                                                                                                object: Identifier(
                                                                                                    Identifier {
                                                                                                        base: BaseNode {
                                                                                                            location: SourceLocation {
                                                                                                                start: "line: 1, column: 46",
                                                                                                                end: "line: 1, column: 47",
                                                                                                                source: "r",
                                                                                                            },
                                                                                                        },
                                                                                                        name: "r",
                                                                                                    },
                                                                                                ),
                                                                                                lbrack: [],
                                                                                                property: Identifier(
                                                                                                    Identifier {
                                                                                                        base: BaseNode {
                                                                                                            location: SourceLocation {
                                                                                                                start: "line: 1, column: 48",
                                                                                                                end: "line: 1, column: 54",
                                                                                                                source: "_value",
                                                                                                            },
                                                                                                        },
                                                                                                        name: "_value",
                                                                                                    },
                                                                                                ),
                                                                                                rbrack: [],
                                                                                            },
                                                                                        ),
                                                                                    },
                                                                                ),
                                                                            ],
                                                                            rbrace: [],
                                                                        },
                                                                    ),
                                                                },
                                                            ),
                                                        ),
                                                        comma: [],
                                                    },
                                                ],
                                                rbrace: [],
                                            },
                                        ),
                                    ],
                                    rparen: [],
                                },
                            },
                        ),
                    },
                ),
            ],
            eof: [],
        }
    "#]].assert_debug_eq(&parsed);
}

// TODO(affo): that error is injected by ast.Check().
#[test]
fn illegal_statement_token() {
    let mut p = Parser::new(r#"@ ident"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 8",
                    source: "@ ident",
                },
            },
            name: "",
            metadata: "parser-type=rust",
            package: None,
            imports: [],
            body: [
                Bad(
                    BadStmt {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 2",
                                source: "@",
                            },
                        },
                        text: "@",
                    },
                ),
                Expr(
                    ExprStmt {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 3",
                                end: "line: 1, column: 8",
                                source: "ident",
                            },
                        },
                        expression: Identifier(
                            Identifier {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 3",
                                        end: "line: 1, column: 8",
                                        source: "ident",
                                    },
                                },
                                name: "ident",
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

// TODO(affo): that error is injected by ast.Check().
#[test]
fn multiple_idents_in_parens() {
    let mut p = Parser::new(r#"(a b)"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 6",
                    source: "(a b)",
                },
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
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 6",
                                source: "(a b)",
                            },
                        },
                        expression: Paren(
                            ParenExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 6",
                                        source: "(a b)",
                                    },
                                },
                                lparen: [],
                                expression: Binary(
                                    BinaryExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 2",
                                                end: "line: 1, column: 5",
                                                source: "a b",
                                            },
                                        },
                                        operator: InvalidOperator,
                                        left: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 2",
                                                        end: "line: 1, column: 3",
                                                        source: "a",
                                                    },
                                                },
                                                name: "a",
                                            },
                                        ),
                                        right: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 4",
                                                        end: "line: 1, column: 5",
                                                        source: "b",
                                                    },
                                                },
                                                name: "b",
                                            },
                                        ),
                                    },
                                ),
                                rparen: [],
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

// TODO(affo): that error is injected by ast.Check().
// TODO(jsternberg): Parens aren't recorded correctly in the source and are mostly ignored.
#[test]
fn missing_left_hand_side() {
    let mut p = Parser::new(r#"(*b)"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 5",
                    source: "(*b)",
                },
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
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 5",
                                source: "(*b)",
                            },
                        },
                        expression: Paren(
                            ParenExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 5",
                                        source: "(*b)",
                                    },
                                },
                                lparen: [],
                                expression: Binary(
                                    BinaryExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 2",
                                                end: "line: 1, column: 4",
                                                source: "*b",
                                            },
                                        },
                                        operator: MultiplicationOperator,
                                        left: Bad(
                                            BadExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 2",
                                                        end: "line: 1, column: 3",
                                                        source: "*",
                                                    },
                                                },
                                                text: "invalid token for primary expression: MUL",
                                                expression: None,
                                            },
                                        ),
                                        right: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 3",
                                                        end: "line: 1, column: 4",
                                                        source: "b",
                                                    },
                                                },
                                                name: "b",
                                            },
                                        ),
                                    },
                                ),
                                rparen: [],
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

// TODO(affo): that error is injected by ast.Check().
// TODO(jsternberg): Parens aren't recorded correctly in the source and are mostly ignored.
#[test]
fn missing_right_hand_side() {
    let mut p = Parser::new(r#"(a*)"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 5",
                    source: "(a*)",
                },
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
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 5",
                                source: "(a*)",
                            },
                        },
                        expression: Paren(
                            ParenExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 5",
                                        source: "(a*)",
                                    },
                                },
                                lparen: [],
                                expression: Binary(
                                    BinaryExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 2",
                                                end: "line: 1, column: 5",
                                                source: "a*)",
                                            },
                                        },
                                        operator: MultiplicationOperator,
                                        left: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 2",
                                                        end: "line: 1, column: 3",
                                                        source: "a",
                                                    },
                                                },
                                                name: "a",
                                            },
                                        ),
                                        right: Bad(
                                            BadExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 4",
                                                        end: "line: 1, column: 5",
                                                        source: ")",
                                                    },
                                                },
                                                text: "invalid token for primary expression: RPAREN",
                                                expression: None,
                                            },
                                        ),
                                    },
                                ),
                                rparen: [],
                            },
                        ),
                    },
                ),
            ],
            eof: [],
        }
    "#]].assert_debug_eq(&parsed);
}

#[test]
fn illegal_expression() {
    let mut p = Parser::new(r#"(@)"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 4",
                    source: "(@)",
                },
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
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 4",
                                source: "(@)",
                            },
                        },
                        expression: Paren(
                            ParenExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 4",
                                        source: "(@)",
                                    },
                                    errors: [
                                        "invalid expression @1:2-1:3: @",
                                    ],
                                },
                                lparen: [],
                                expression: Bad(
                                    BadExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 2",
                                                end: "line: 1, column: 3",
                                                source: "@",
                                            },
                                        },
                                        text: "@",
                                        expression: None,
                                    },
                                ),
                                rparen: [],
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

// NOTE(affo): this is slightly different from Go. We have a BadExpr in the body.
#[test]
fn missing_arrow_in_function_expression() {
    let mut p = Parser::new(r#"(a, b) a + b"#);
    let parsed = p.parse_file("".to_string());
    expect_test::expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 13",
                    source: "(a, b) a + b",
                },
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
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 13",
                                source: "(a, b) a + b",
                            },
                        },
                        expression: Function(
                            FunctionExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 13",
                                        source: "(a, b) a + b",
                                    },
                                },
                                lparen: [],
                                params: [
                                    Property {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 2",
                                                end: "line: 1, column: 3",
                                                source: "a",
                                            },
                                        },
                                        key: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 2",
                                                        end: "line: 1, column: 3",
                                                        source: "a",
                                                    },
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
                                                start: "line: 1, column: 5",
                                                end: "line: 1, column: 6",
                                                source: "b",
                                            },
                                        },
                                        key: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 5",
                                                        end: "line: 1, column: 6",
                                                        source: "b",
                                                    },
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
                                                    start: "line: 1, column: 8",
                                                    end: "line: 1, column: 13",
                                                    source: "a + b",
                                                },
                                            },
                                            operator: AdditionOperator,
                                            left: Identifier(
                                                Identifier {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 8",
                                                            end: "line: 1, column: 9",
                                                            source: "a",
                                                        },
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
                                                            start: "line: 1, column: 12",
                                                            end: "line: 1, column: 13",
                                                            source: "b",
                                                        },
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
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 6",
                    source: "a[b()",
                },
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
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 6",
                                source: "a[b()",
                            },
                        },
                        expression: Index(
                            IndexExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 6",
                                        source: "a[b()",
                                    },
                                    errors: [
                                        "expected RBRACK, got EOF",
                                    ],
                                },
                                array: Identifier(
                                    Identifier {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 1",
                                                end: "line: 1, column: 2",
                                                source: "a",
                                            },
                                        },
                                        name: "a",
                                    },
                                ),
                                lbrack: [],
                                index: Call(
                                    CallExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 3",
                                                end: "line: 1, column: 6",
                                                source: "b()",
                                            },
                                        },
                                        callee: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 3",
                                                        end: "line: 1, column: 4",
                                                        source: "b",
                                                    },
                                                },
                                                name: "b",
                                            },
                                        ),
                                        lparen: [],
                                        arguments: [],
                                        rparen: [],
                                    },
                                ),
                                rbrack: [],
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
fn index_with_unbalanced_parenthesis() {
    let mut p = Parser::new(r#"a[b(]"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 6",
                    source: "a[b(]",
                },
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
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 6",
                                source: "a[b(]",
                            },
                        },
                        expression: Index(
                            IndexExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 6",
                                        source: "a[b(]",
                                    },
                                },
                                array: Identifier(
                                    Identifier {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 1",
                                                end: "line: 1, column: 2",
                                                source: "a",
                                            },
                                        },
                                        name: "a",
                                    },
                                ),
                                lbrack: [],
                                index: Call(
                                    CallExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 3",
                                                end: "line: 1, column: 6",
                                                source: "b(]",
                                            },
                                            errors: [
                                                "expected RPAREN, got RBRACK",
                                            ],
                                        },
                                        callee: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 3",
                                                        end: "line: 1, column: 4",
                                                        source: "b",
                                                    },
                                                },
                                                name: "b",
                                            },
                                        ),
                                        lparen: [],
                                        arguments: [],
                                        rparen: [],
                                    },
                                ),
                                rbrack: [],
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
fn index_with_unexpected_rparen() {
    let mut p = Parser::new(r#"a[b)]"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 6",
                    source: "a[b)]",
                },
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
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 6",
                                source: "a[b)]",
                            },
                        },
                        expression: Index(
                            IndexExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 6",
                                        source: "a[b)]",
                                    },
                                    errors: [
                                        "invalid expression @1:4-1:5: )",
                                    ],
                                },
                                array: Identifier(
                                    Identifier {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 1",
                                                end: "line: 1, column: 2",
                                                source: "a",
                                            },
                                        },
                                        name: "a",
                                    },
                                ),
                                lbrack: [],
                                index: Identifier(
                                    Identifier {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 3",
                                                end: "line: 1, column: 4",
                                                source: "b",
                                            },
                                        },
                                        name: "b",
                                    },
                                ),
                                rbrack: [],
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
fn int_literal_zero_prefix() {
    let mut p = Parser::new(r#"0123"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 5",
                    source: "0123",
                },
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
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 5",
                                source: "0123",
                            },
                        },
                        expression: Integer(
                            IntegerLit {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 5",
                                        source: "0123",
                                    },
                                    errors: [
                                        "invalid integer literal \"0123\": nonzero value cannot start with 0",
                                    ],
                                },
                                value: 0,
                            },
                        ),
                    },
                ),
            ],
            eof: [],
        }
    "#]].assert_debug_eq(&parsed);
}

fn test_error_msg(src: &str, expect: expect_test::Expect) {
    let mut p = Parser::new(src);
    let parsed = p.parse_file("".to_string());
    expect.assert_eq(
        &ast::check::check(ast::walk::Node::File(&parsed))
            .unwrap_err()
            .to_string(),
    );
}

#[test]
fn parse_invalid_call() {
    let mut p = Parser::new("json.(v: r._value)");
    let parsed = p.parse_file("".to_string());

    // Checks that the identifier in the ast after the `.` does not get assigned `(` which would
    // show up as `json.((v: r._value)`.
    assert_eq!(
        crate::formatter::format_node(&parsed).unwrap(),
        "json.(v: r._value)\n"
    );
}

#[test]
fn issue_4231() {
    test_error_msg(
        r#"
            map(fn: (r) => ({ r with _value: if true and false then 1}) )
        "#,
        expect_test::expect![[r#"error @2:46-2:70: expected ELSE, got RBRACE (}) at 2:70"#]],
    );
}

#[test]
fn missing_property() {
    test_error_msg(
        r#"
            x.

            builtin y : int
        "#,
        expect_test::expect![[
            r#"error @4:13-4:13: expected IDENT, got BUILTIN (builtin) at 4:13"#
        ]],
    );
}

#[test]
fn missing_identifier_in_option() {
    test_error_msg(
        r#"
            option =

            buckets()
        "#,
        expect_test::expect![[r#"error @2:20-2:20: expected IDENT, got ASSIGN (=) at 2:20"#]],
    );
}

#[test]
fn dont_stack_overflow() {
    let mut p = Parser::new(include_str!("stack_overflow.flux"));
    let parsed = p.parse_file("".to_string());
    assert!(&ast::check::check(ast::walk::Node::File(&parsed))
        .unwrap_err()
        .iter()
        .any(|err| err.error.is_fatal()));
}

#[test]
fn dont_stack_overflow_2() {
    let mut p = Parser::new(include_str!("stack_overflow_2.flux"));
    let parsed = p.parse_file("".to_string());
    assert!(&ast::check::check(ast::walk::Node::File(&parsed))
        .unwrap_err()
        .iter()
        .any(|err| err.error.is_fatal()));
}
