use super::*;

#[test]
fn binary_operator_precedence() {
    let mut p = Parser::new(r#"a / b - 1.0"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 12",
                    source: "a / b - 1.0",
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
                                end: "line: 1, column: 12",
                                source: "a / b - 1.0",
                            },
                        },
                        expression: Binary(
                            BinaryExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 12",
                                        source: "a / b - 1.0",
                                    },
                                },
                                operator: SubtractionOperator,
                                left: Binary(
                                    BinaryExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 1",
                                                end: "line: 1, column: 6",
                                                source: "a / b",
                                            },
                                        },
                                        operator: DivisionOperator,
                                        left: Identifier(
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
                                        right: Identifier(
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
                                    },
                                ),
                                right: Float(
                                    FloatLit {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 9",
                                                end: "line: 1, column: 12",
                                                source: "1.0",
                                            },
                                        },
                                        value: NotNan(
                                            1.0,
                                        ),
                                    },
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
fn binary_operator_precedence_literals_only() {
    let mut p = Parser::new(r#"2 / "a" - 1.0"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 14",
                    source: "2 / \"a\" - 1.0",
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
                                end: "line: 1, column: 14",
                                source: "2 / \"a\" - 1.0",
                            },
                        },
                        expression: Binary(
                            BinaryExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 14",
                                        source: "2 / \"a\" - 1.0",
                                    },
                                },
                                operator: SubtractionOperator,
                                left: Binary(
                                    BinaryExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 1",
                                                end: "line: 1, column: 8",
                                                source: "2 / \"a\"",
                                            },
                                        },
                                        operator: DivisionOperator,
                                        left: Integer(
                                            IntegerLit {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 1",
                                                        end: "line: 1, column: 2",
                                                        source: "2",
                                                    },
                                                },
                                                value: 2,
                                            },
                                        ),
                                        right: StringLit(
                                            StringLit {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 5",
                                                        end: "line: 1, column: 8",
                                                        source: "\"a\"",
                                                    },
                                                },
                                                value: "a",
                                            },
                                        ),
                                    },
                                ),
                                right: Float(
                                    FloatLit {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 11",
                                                end: "line: 1, column: 14",
                                                source: "1.0",
                                            },
                                        },
                                        value: NotNan(
                                            1.0,
                                        ),
                                    },
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
fn binary_operator_precedence_double_subtraction() {
    let mut p = Parser::new(r#"1 - 2 - 3"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 10",
                    source: "1 - 2 - 3",
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
                                end: "line: 1, column: 10",
                                source: "1 - 2 - 3",
                            },
                        },
                        expression: Binary(
                            BinaryExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 10",
                                        source: "1 - 2 - 3",
                                    },
                                },
                                operator: SubtractionOperator,
                                left: Binary(
                                    BinaryExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 1",
                                                end: "line: 1, column: 6",
                                                source: "1 - 2",
                                            },
                                        },
                                        operator: SubtractionOperator,
                                        left: Integer(
                                            IntegerLit {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 1",
                                                        end: "line: 1, column: 2",
                                                        source: "1",
                                                    },
                                                },
                                                value: 1,
                                            },
                                        ),
                                        right: Integer(
                                            IntegerLit {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 5",
                                                        end: "line: 1, column: 6",
                                                        source: "2",
                                                    },
                                                },
                                                value: 2,
                                            },
                                        ),
                                    },
                                ),
                                right: Integer(
                                    IntegerLit {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 9",
                                                end: "line: 1, column: 10",
                                                source: "3",
                                            },
                                        },
                                        value: 3,
                                    },
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
fn binary_operator_precedence_double_subtraction_with_parens() {
    let mut p = Parser::new(r#"1 - (2 - 3)"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 12",
                    source: "1 - (2 - 3)",
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
                                end: "line: 1, column: 12",
                                source: "1 - (2 - 3)",
                            },
                        },
                        expression: Binary(
                            BinaryExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 12",
                                        source: "1 - (2 - 3)",
                                    },
                                },
                                operator: SubtractionOperator,
                                left: Integer(
                                    IntegerLit {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 1",
                                                end: "line: 1, column: 2",
                                                source: "1",
                                            },
                                        },
                                        value: 1,
                                    },
                                ),
                                right: Paren(
                                    ParenExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 5",
                                                end: "line: 1, column: 12",
                                                source: "(2 - 3)",
                                            },
                                        },
                                        lparen: [],
                                        expression: Binary(
                                            BinaryExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 6",
                                                        end: "line: 1, column: 11",
                                                        source: "2 - 3",
                                                    },
                                                },
                                                operator: SubtractionOperator,
                                                left: Integer(
                                                    IntegerLit {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 1, column: 6",
                                                                end: "line: 1, column: 7",
                                                                source: "2",
                                                            },
                                                        },
                                                        value: 2,
                                                    },
                                                ),
                                                right: Integer(
                                                    IntegerLit {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 1, column: 10",
                                                                end: "line: 1, column: 11",
                                                                source: "3",
                                                            },
                                                        },
                                                        value: 3,
                                                    },
                                                ),
                                            },
                                        ),
                                        rparen: [],
                                    },
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
fn binary_operator_precedence_double_sum() {
    let mut p = Parser::new(r#"1 + 2 + 3"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 10",
                    source: "1 + 2 + 3",
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
                                end: "line: 1, column: 10",
                                source: "1 + 2 + 3",
                            },
                        },
                        expression: Binary(
                            BinaryExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 10",
                                        source: "1 + 2 + 3",
                                    },
                                },
                                operator: AdditionOperator,
                                left: Binary(
                                    BinaryExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 1",
                                                end: "line: 1, column: 6",
                                                source: "1 + 2",
                                            },
                                        },
                                        operator: AdditionOperator,
                                        left: Integer(
                                            IntegerLit {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 1",
                                                        end: "line: 1, column: 2",
                                                        source: "1",
                                                    },
                                                },
                                                value: 1,
                                            },
                                        ),
                                        right: Integer(
                                            IntegerLit {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 5",
                                                        end: "line: 1, column: 6",
                                                        source: "2",
                                                    },
                                                },
                                                value: 2,
                                            },
                                        ),
                                    },
                                ),
                                right: Integer(
                                    IntegerLit {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 9",
                                                end: "line: 1, column: 10",
                                                source: "3",
                                            },
                                        },
                                        value: 3,
                                    },
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
fn binary_operator_precedence_exponent() {
    let mut p = Parser::new(r#"5 * 1 ^ 5"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 10",
                    source: "5 * 1 ^ 5",
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
                                end: "line: 1, column: 10",
                                source: "5 * 1 ^ 5",
                            },
                        },
                        expression: Binary(
                            BinaryExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 10",
                                        source: "5 * 1 ^ 5",
                                    },
                                },
                                operator: MultiplicationOperator,
                                left: Integer(
                                    IntegerLit {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 1",
                                                end: "line: 1, column: 2",
                                                source: "5",
                                            },
                                        },
                                        value: 5,
                                    },
                                ),
                                right: Binary(
                                    BinaryExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 5",
                                                end: "line: 1, column: 10",
                                                source: "1 ^ 5",
                                            },
                                        },
                                        operator: PowerOperator,
                                        left: Integer(
                                            IntegerLit {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 5",
                                                        end: "line: 1, column: 6",
                                                        source: "1",
                                                    },
                                                },
                                                value: 1,
                                            },
                                        ),
                                        right: Integer(
                                            IntegerLit {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 9",
                                                        end: "line: 1, column: 10",
                                                        source: "5",
                                                    },
                                                },
                                                value: 5,
                                            },
                                        ),
                                    },
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
fn binary_operator_precedence_double_sum_with_parens() {
    let mut p = Parser::new(r#"1 + (2 + 3)"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 12",
                    source: "1 + (2 + 3)",
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
                                end: "line: 1, column: 12",
                                source: "1 + (2 + 3)",
                            },
                        },
                        expression: Binary(
                            BinaryExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 12",
                                        source: "1 + (2 + 3)",
                                    },
                                },
                                operator: AdditionOperator,
                                left: Integer(
                                    IntegerLit {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 1",
                                                end: "line: 1, column: 2",
                                                source: "1",
                                            },
                                        },
                                        value: 1,
                                    },
                                ),
                                right: Paren(
                                    ParenExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 5",
                                                end: "line: 1, column: 12",
                                                source: "(2 + 3)",
                                            },
                                        },
                                        lparen: [],
                                        expression: Binary(
                                            BinaryExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 6",
                                                        end: "line: 1, column: 11",
                                                        source: "2 + 3",
                                                    },
                                                },
                                                operator: AdditionOperator,
                                                left: Integer(
                                                    IntegerLit {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 1, column: 6",
                                                                end: "line: 1, column: 7",
                                                                source: "2",
                                                            },
                                                        },
                                                        value: 2,
                                                    },
                                                ),
                                                right: Integer(
                                                    IntegerLit {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 1, column: 10",
                                                                end: "line: 1, column: 11",
                                                                source: "3",
                                                            },
                                                        },
                                                        value: 3,
                                                    },
                                                ),
                                            },
                                        ),
                                        rparen: [],
                                    },
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
fn binary_operator_precedence_exponent_with_parens() {
    let mut p = Parser::new(r#"2 ^ (1 + 3)"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 12",
                    source: "2 ^ (1 + 3)",
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
                                end: "line: 1, column: 12",
                                source: "2 ^ (1 + 3)",
                            },
                        },
                        expression: Binary(
                            BinaryExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 12",
                                        source: "2 ^ (1 + 3)",
                                    },
                                },
                                operator: PowerOperator,
                                left: Integer(
                                    IntegerLit {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 1",
                                                end: "line: 1, column: 2",
                                                source: "2",
                                            },
                                        },
                                        value: 2,
                                    },
                                ),
                                right: Paren(
                                    ParenExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 5",
                                                end: "line: 1, column: 12",
                                                source: "(1 + 3)",
                                            },
                                        },
                                        lparen: [],
                                        expression: Binary(
                                            BinaryExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 6",
                                                        end: "line: 1, column: 11",
                                                        source: "1 + 3",
                                                    },
                                                },
                                                operator: AdditionOperator,
                                                left: Integer(
                                                    IntegerLit {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 1, column: 6",
                                                                end: "line: 1, column: 7",
                                                                source: "1",
                                                            },
                                                        },
                                                        value: 1,
                                                    },
                                                ),
                                                right: Integer(
                                                    IntegerLit {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 1, column: 10",
                                                                end: "line: 1, column: 11",
                                                                source: "3",
                                                            },
                                                        },
                                                        value: 3,
                                                    },
                                                ),
                                            },
                                        ),
                                        rparen: [],
                                    },
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
fn logical_unary_operator_precedence() {
    let mut p = Parser::new(r#"not -1 == a"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 12",
                    source: "not -1 == a",
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
                                end: "line: 1, column: 12",
                                source: "not -1 == a",
                            },
                        },
                        expression: Unary(
                            UnaryExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 12",
                                        source: "not -1 == a",
                                    },
                                },
                                operator: NotOperator,
                                argument: Binary(
                                    BinaryExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 5",
                                                end: "line: 1, column: 12",
                                                source: "-1 == a",
                                            },
                                        },
                                        operator: EqualOperator,
                                        left: Unary(
                                            UnaryExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 5",
                                                        end: "line: 1, column: 7",
                                                        source: "-1",
                                                    },
                                                },
                                                operator: SubtractionOperator,
                                                argument: Integer(
                                                    IntegerLit {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 1, column: 6",
                                                                end: "line: 1, column: 7",
                                                                source: "1",
                                                            },
                                                        },
                                                        value: 1,
                                                    },
                                                ),
                                            },
                                        ),
                                        right: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 11",
                                                        end: "line: 1, column: 12",
                                                        source: "a",
                                                    },
                                                },
                                                name: "a",
                                            },
                                        ),
                                    },
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
fn all_operators_precedence() {
    let mut p = Parser::new(
        r#"a() == b.a + b.c * d < 100 and e != f[g] and h > i * j and
k / l < m + n - o or p() <= q() or r >= s and not t =~ /a/ and u !~ /a/"#,
    );
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 2, column: 72",
                    source: "a() == b.a + b.c * d < 100 and e != f[g] and h > i * j and\nk / l < m + n - o or p() <= q() or r >= s and not t =~ /a/ and u !~ /a/",
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
                                end: "line: 2, column: 72",
                                source: "a() == b.a + b.c * d < 100 and e != f[g] and h > i * j and\nk / l < m + n - o or p() <= q() or r >= s and not t =~ /a/ and u !~ /a/",
                            },
                        },
                        expression: Logical(
                            LogicalExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 2, column: 72",
                                        source: "a() == b.a + b.c * d < 100 and e != f[g] and h > i * j and\nk / l < m + n - o or p() <= q() or r >= s and not t =~ /a/ and u !~ /a/",
                                    },
                                },
                                operator: OrOperator,
                                left: Logical(
                                    LogicalExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 1",
                                                end: "line: 2, column: 32",
                                                source: "a() == b.a + b.c * d < 100 and e != f[g] and h > i * j and\nk / l < m + n - o or p() <= q()",
                                            },
                                        },
                                        operator: OrOperator,
                                        left: Logical(
                                            LogicalExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 1",
                                                        end: "line: 2, column: 18",
                                                        source: "a() == b.a + b.c * d < 100 and e != f[g] and h > i * j and\nk / l < m + n - o",
                                                    },
                                                },
                                                operator: AndOperator,
                                                left: Logical(
                                                    LogicalExpr {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 1, column: 1",
                                                                end: "line: 1, column: 55",
                                                                source: "a() == b.a + b.c * d < 100 and e != f[g] and h > i * j",
                                                            },
                                                        },
                                                        operator: AndOperator,
                                                        left: Logical(
                                                            LogicalExpr {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 1, column: 1",
                                                                        end: "line: 1, column: 41",
                                                                        source: "a() == b.a + b.c * d < 100 and e != f[g]",
                                                                    },
                                                                },
                                                                operator: AndOperator,
                                                                left: Binary(
                                                                    BinaryExpr {
                                                                        base: BaseNode {
                                                                            location: SourceLocation {
                                                                                start: "line: 1, column: 1",
                                                                                end: "line: 1, column: 27",
                                                                                source: "a() == b.a + b.c * d < 100",
                                                                            },
                                                                        },
                                                                        operator: LessThanOperator,
                                                                        left: Binary(
                                                                            BinaryExpr {
                                                                                base: BaseNode {
                                                                                    location: SourceLocation {
                                                                                        start: "line: 1, column: 1",
                                                                                        end: "line: 1, column: 21",
                                                                                        source: "a() == b.a + b.c * d",
                                                                                    },
                                                                                },
                                                                                operator: EqualOperator,
                                                                                left: Call(
                                                                                    CallExpr {
                                                                                        base: BaseNode {
                                                                                            location: SourceLocation {
                                                                                                start: "line: 1, column: 1",
                                                                                                end: "line: 1, column: 4",
                                                                                                source: "a()",
                                                                                            },
                                                                                        },
                                                                                        callee: Identifier(
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
                                                                                        lparen: [],
                                                                                        arguments: [],
                                                                                        rparen: [],
                                                                                    },
                                                                                ),
                                                                                right: Binary(
                                                                                    BinaryExpr {
                                                                                        base: BaseNode {
                                                                                            location: SourceLocation {
                                                                                                start: "line: 1, column: 8",
                                                                                                end: "line: 1, column: 21",
                                                                                                source: "b.a + b.c * d",
                                                                                            },
                                                                                        },
                                                                                        operator: AdditionOperator,
                                                                                        left: Member(
                                                                                            MemberExpr {
                                                                                                base: BaseNode {
                                                                                                    location: SourceLocation {
                                                                                                        start: "line: 1, column: 8",
                                                                                                        end: "line: 1, column: 11",
                                                                                                        source: "b.a",
                                                                                                    },
                                                                                                },
                                                                                                object: Identifier(
                                                                                                    Identifier {
                                                                                                        base: BaseNode {
                                                                                                            location: SourceLocation {
                                                                                                                start: "line: 1, column: 8",
                                                                                                                end: "line: 1, column: 9",
                                                                                                                source: "b",
                                                                                                            },
                                                                                                        },
                                                                                                        name: "b",
                                                                                                    },
                                                                                                ),
                                                                                                lbrack: [],
                                                                                                property: Identifier(
                                                                                                    Identifier {
                                                                                                        base: BaseNode {
                                                                                                            location: SourceLocation {
                                                                                                                start: "line: 1, column: 10",
                                                                                                                end: "line: 1, column: 11",
                                                                                                                source: "a",
                                                                                                            },
                                                                                                        },
                                                                                                        name: "a",
                                                                                                    },
                                                                                                ),
                                                                                                rbrack: [],
                                                                                            },
                                                                                        ),
                                                                                        right: Binary(
                                                                                            BinaryExpr {
                                                                                                base: BaseNode {
                                                                                                    location: SourceLocation {
                                                                                                        start: "line: 1, column: 14",
                                                                                                        end: "line: 1, column: 21",
                                                                                                        source: "b.c * d",
                                                                                                    },
                                                                                                },
                                                                                                operator: MultiplicationOperator,
                                                                                                left: Member(
                                                                                                    MemberExpr {
                                                                                                        base: BaseNode {
                                                                                                            location: SourceLocation {
                                                                                                                start: "line: 1, column: 14",
                                                                                                                end: "line: 1, column: 17",
                                                                                                                source: "b.c",
                                                                                                            },
                                                                                                        },
                                                                                                        object: Identifier(
                                                                                                            Identifier {
                                                                                                                base: BaseNode {
                                                                                                                    location: SourceLocation {
                                                                                                                        start: "line: 1, column: 14",
                                                                                                                        end: "line: 1, column: 15",
                                                                                                                        source: "b",
                                                                                                                    },
                                                                                                                },
                                                                                                                name: "b",
                                                                                                            },
                                                                                                        ),
                                                                                                        lbrack: [],
                                                                                                        property: Identifier(
                                                                                                            Identifier {
                                                                                                                base: BaseNode {
                                                                                                                    location: SourceLocation {
                                                                                                                        start: "line: 1, column: 16",
                                                                                                                        end: "line: 1, column: 17",
                                                                                                                        source: "c",
                                                                                                                    },
                                                                                                                },
                                                                                                                name: "c",
                                                                                                            },
                                                                                                        ),
                                                                                                        rbrack: [],
                                                                                                    },
                                                                                                ),
                                                                                                right: Identifier(
                                                                                                    Identifier {
                                                                                                        base: BaseNode {
                                                                                                            location: SourceLocation {
                                                                                                                start: "line: 1, column: 20",
                                                                                                                end: "line: 1, column: 21",
                                                                                                                source: "d",
                                                                                                            },
                                                                                                        },
                                                                                                        name: "d",
                                                                                                    },
                                                                                                ),
                                                                                            },
                                                                                        ),
                                                                                    },
                                                                                ),
                                                                            },
                                                                        ),
                                                                        right: Integer(
                                                                            IntegerLit {
                                                                                base: BaseNode {
                                                                                    location: SourceLocation {
                                                                                        start: "line: 1, column: 24",
                                                                                        end: "line: 1, column: 27",
                                                                                        source: "100",
                                                                                    },
                                                                                },
                                                                                value: 100,
                                                                            },
                                                                        ),
                                                                    },
                                                                ),
                                                                right: Binary(
                                                                    BinaryExpr {
                                                                        base: BaseNode {
                                                                            location: SourceLocation {
                                                                                start: "line: 1, column: 32",
                                                                                end: "line: 1, column: 41",
                                                                                source: "e != f[g]",
                                                                            },
                                                                        },
                                                                        operator: NotEqualOperator,
                                                                        left: Identifier(
                                                                            Identifier {
                                                                                base: BaseNode {
                                                                                    location: SourceLocation {
                                                                                        start: "line: 1, column: 32",
                                                                                        end: "line: 1, column: 33",
                                                                                        source: "e",
                                                                                    },
                                                                                },
                                                                                name: "e",
                                                                            },
                                                                        ),
                                                                        right: Index(
                                                                            IndexExpr {
                                                                                base: BaseNode {
                                                                                    location: SourceLocation {
                                                                                        start: "line: 1, column: 37",
                                                                                        end: "line: 1, column: 41",
                                                                                        source: "f[g]",
                                                                                    },
                                                                                },
                                                                                array: Identifier(
                                                                                    Identifier {
                                                                                        base: BaseNode {
                                                                                            location: SourceLocation {
                                                                                                start: "line: 1, column: 37",
                                                                                                end: "line: 1, column: 38",
                                                                                                source: "f",
                                                                                            },
                                                                                        },
                                                                                        name: "f",
                                                                                    },
                                                                                ),
                                                                                lbrack: [],
                                                                                index: Identifier(
                                                                                    Identifier {
                                                                                        base: BaseNode {
                                                                                            location: SourceLocation {
                                                                                                start: "line: 1, column: 39",
                                                                                                end: "line: 1, column: 40",
                                                                                                source: "g",
                                                                                            },
                                                                                        },
                                                                                        name: "g",
                                                                                    },
                                                                                ),
                                                                                rbrack: [],
                                                                            },
                                                                        ),
                                                                    },
                                                                ),
                                                            },
                                                        ),
                                                        right: Binary(
                                                            BinaryExpr {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 1, column: 46",
                                                                        end: "line: 1, column: 55",
                                                                        source: "h > i * j",
                                                                    },
                                                                },
                                                                operator: GreaterThanOperator,
                                                                left: Identifier(
                                                                    Identifier {
                                                                        base: BaseNode {
                                                                            location: SourceLocation {
                                                                                start: "line: 1, column: 46",
                                                                                end: "line: 1, column: 47",
                                                                                source: "h",
                                                                            },
                                                                        },
                                                                        name: "h",
                                                                    },
                                                                ),
                                                                right: Binary(
                                                                    BinaryExpr {
                                                                        base: BaseNode {
                                                                            location: SourceLocation {
                                                                                start: "line: 1, column: 50",
                                                                                end: "line: 1, column: 55",
                                                                                source: "i * j",
                                                                            },
                                                                        },
                                                                        operator: MultiplicationOperator,
                                                                        left: Identifier(
                                                                            Identifier {
                                                                                base: BaseNode {
                                                                                    location: SourceLocation {
                                                                                        start: "line: 1, column: 50",
                                                                                        end: "line: 1, column: 51",
                                                                                        source: "i",
                                                                                    },
                                                                                },
                                                                                name: "i",
                                                                            },
                                                                        ),
                                                                        right: Identifier(
                                                                            Identifier {
                                                                                base: BaseNode {
                                                                                    location: SourceLocation {
                                                                                        start: "line: 1, column: 54",
                                                                                        end: "line: 1, column: 55",
                                                                                        source: "j",
                                                                                    },
                                                                                },
                                                                                name: "j",
                                                                            },
                                                                        ),
                                                                    },
                                                                ),
                                                            },
                                                        ),
                                                    },
                                                ),
                                                right: Binary(
                                                    BinaryExpr {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 2, column: 1",
                                                                end: "line: 2, column: 18",
                                                                source: "k / l < m + n - o",
                                                            },
                                                        },
                                                        operator: LessThanOperator,
                                                        left: Binary(
                                                            BinaryExpr {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 2, column: 1",
                                                                        end: "line: 2, column: 6",
                                                                        source: "k / l",
                                                                    },
                                                                },
                                                                operator: DivisionOperator,
                                                                left: Identifier(
                                                                    Identifier {
                                                                        base: BaseNode {
                                                                            location: SourceLocation {
                                                                                start: "line: 2, column: 1",
                                                                                end: "line: 2, column: 2",
                                                                                source: "k",
                                                                            },
                                                                        },
                                                                        name: "k",
                                                                    },
                                                                ),
                                                                right: Identifier(
                                                                    Identifier {
                                                                        base: BaseNode {
                                                                            location: SourceLocation {
                                                                                start: "line: 2, column: 5",
                                                                                end: "line: 2, column: 6",
                                                                                source: "l",
                                                                            },
                                                                        },
                                                                        name: "l",
                                                                    },
                                                                ),
                                                            },
                                                        ),
                                                        right: Binary(
                                                            BinaryExpr {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 2, column: 9",
                                                                        end: "line: 2, column: 18",
                                                                        source: "m + n - o",
                                                                    },
                                                                },
                                                                operator: SubtractionOperator,
                                                                left: Binary(
                                                                    BinaryExpr {
                                                                        base: BaseNode {
                                                                            location: SourceLocation {
                                                                                start: "line: 2, column: 9",
                                                                                end: "line: 2, column: 14",
                                                                                source: "m + n",
                                                                            },
                                                                        },
                                                                        operator: AdditionOperator,
                                                                        left: Identifier(
                                                                            Identifier {
                                                                                base: BaseNode {
                                                                                    location: SourceLocation {
                                                                                        start: "line: 2, column: 9",
                                                                                        end: "line: 2, column: 10",
                                                                                        source: "m",
                                                                                    },
                                                                                },
                                                                                name: "m",
                                                                            },
                                                                        ),
                                                                        right: Identifier(
                                                                            Identifier {
                                                                                base: BaseNode {
                                                                                    location: SourceLocation {
                                                                                        start: "line: 2, column: 13",
                                                                                        end: "line: 2, column: 14",
                                                                                        source: "n",
                                                                                    },
                                                                                },
                                                                                name: "n",
                                                                            },
                                                                        ),
                                                                    },
                                                                ),
                                                                right: Identifier(
                                                                    Identifier {
                                                                        base: BaseNode {
                                                                            location: SourceLocation {
                                                                                start: "line: 2, column: 17",
                                                                                end: "line: 2, column: 18",
                                                                                source: "o",
                                                                            },
                                                                        },
                                                                        name: "o",
                                                                    },
                                                                ),
                                                            },
                                                        ),
                                                    },
                                                ),
                                            },
                                        ),
                                        right: Binary(
                                            BinaryExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 2, column: 22",
                                                        end: "line: 2, column: 32",
                                                        source: "p() <= q()",
                                                    },
                                                },
                                                operator: LessThanEqualOperator,
                                                left: Call(
                                                    CallExpr {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 2, column: 22",
                                                                end: "line: 2, column: 25",
                                                                source: "p()",
                                                            },
                                                        },
                                                        callee: Identifier(
                                                            Identifier {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 2, column: 22",
                                                                        end: "line: 2, column: 23",
                                                                        source: "p",
                                                                    },
                                                                },
                                                                name: "p",
                                                            },
                                                        ),
                                                        lparen: [],
                                                        arguments: [],
                                                        rparen: [],
                                                    },
                                                ),
                                                right: Call(
                                                    CallExpr {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 2, column: 29",
                                                                end: "line: 2, column: 32",
                                                                source: "q()",
                                                            },
                                                        },
                                                        callee: Identifier(
                                                            Identifier {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 2, column: 29",
                                                                        end: "line: 2, column: 30",
                                                                        source: "q",
                                                                    },
                                                                },
                                                                name: "q",
                                                            },
                                                        ),
                                                        lparen: [],
                                                        arguments: [],
                                                        rparen: [],
                                                    },
                                                ),
                                            },
                                        ),
                                    },
                                ),
                                right: Logical(
                                    LogicalExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 2, column: 36",
                                                end: "line: 2, column: 72",
                                                source: "r >= s and not t =~ /a/ and u !~ /a/",
                                            },
                                        },
                                        operator: AndOperator,
                                        left: Logical(
                                            LogicalExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 2, column: 36",
                                                        end: "line: 2, column: 59",
                                                        source: "r >= s and not t =~ /a/",
                                                    },
                                                },
                                                operator: AndOperator,
                                                left: Binary(
                                                    BinaryExpr {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 2, column: 36",
                                                                end: "line: 2, column: 42",
                                                                source: "r >= s",
                                                            },
                                                        },
                                                        operator: GreaterThanEqualOperator,
                                                        left: Identifier(
                                                            Identifier {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 2, column: 36",
                                                                        end: "line: 2, column: 37",
                                                                        source: "r",
                                                                    },
                                                                },
                                                                name: "r",
                                                            },
                                                        ),
                                                        right: Identifier(
                                                            Identifier {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 2, column: 41",
                                                                        end: "line: 2, column: 42",
                                                                        source: "s",
                                                                    },
                                                                },
                                                                name: "s",
                                                            },
                                                        ),
                                                    },
                                                ),
                                                right: Unary(
                                                    UnaryExpr {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 2, column: 47",
                                                                end: "line: 2, column: 59",
                                                                source: "not t =~ /a/",
                                                            },
                                                        },
                                                        operator: NotOperator,
                                                        argument: Binary(
                                                            BinaryExpr {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 2, column: 51",
                                                                        end: "line: 2, column: 59",
                                                                        source: "t =~ /a/",
                                                                    },
                                                                },
                                                                operator: RegexpMatchOperator,
                                                                left: Identifier(
                                                                    Identifier {
                                                                        base: BaseNode {
                                                                            location: SourceLocation {
                                                                                start: "line: 2, column: 51",
                                                                                end: "line: 2, column: 52",
                                                                                source: "t",
                                                                            },
                                                                        },
                                                                        name: "t",
                                                                    },
                                                                ),
                                                                right: Regexp(
                                                                    RegexpLit {
                                                                        base: BaseNode {
                                                                            location: SourceLocation {
                                                                                start: "line: 2, column: 56",
                                                                                end: "line: 2, column: 59",
                                                                                source: "/a/",
                                                                            },
                                                                        },
                                                                        value: "a",
                                                                    },
                                                                ),
                                                            },
                                                        ),
                                                    },
                                                ),
                                            },
                                        ),
                                        right: Binary(
                                            BinaryExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 2, column: 64",
                                                        end: "line: 2, column: 72",
                                                        source: "u !~ /a/",
                                                    },
                                                },
                                                operator: NotRegexpMatchOperator,
                                                left: Identifier(
                                                    Identifier {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 2, column: 64",
                                                                end: "line: 2, column: 65",
                                                                source: "u",
                                                            },
                                                        },
                                                        name: "u",
                                                    },
                                                ),
                                                right: Regexp(
                                                    RegexpLit {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 2, column: 69",
                                                                end: "line: 2, column: 72",
                                                                source: "/a/",
                                                            },
                                                        },
                                                        value: "a",
                                                    },
                                                ),
                                            },
                                        ),
                                    },
                                ),
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
fn logical_operators_precedence_1() {
    let mut p = Parser::new(r#"not a or b"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 11",
                    source: "not a or b",
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
                                end: "line: 1, column: 11",
                                source: "not a or b",
                            },
                        },
                        expression: Logical(
                            LogicalExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 11",
                                        source: "not a or b",
                                    },
                                },
                                operator: OrOperator,
                                left: Unary(
                                    UnaryExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 1",
                                                end: "line: 1, column: 6",
                                                source: "not a",
                                            },
                                        },
                                        operator: NotOperator,
                                        argument: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 5",
                                                        end: "line: 1, column: 6",
                                                        source: "a",
                                                    },
                                                },
                                                name: "a",
                                            },
                                        ),
                                    },
                                ),
                                right: Identifier(
                                    Identifier {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 10",
                                                end: "line: 1, column: 11",
                                                source: "b",
                                            },
                                        },
                                        name: "b",
                                    },
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
fn logical_operators_precedence_2() {
    let mut p = Parser::new(r#"a or not b"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 11",
                    source: "a or not b",
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
                                end: "line: 1, column: 11",
                                source: "a or not b",
                            },
                        },
                        expression: Logical(
                            LogicalExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 11",
                                        source: "a or not b",
                                    },
                                },
                                operator: OrOperator,
                                left: Identifier(
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
                                right: Unary(
                                    UnaryExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 6",
                                                end: "line: 1, column: 11",
                                                source: "not b",
                                            },
                                        },
                                        operator: NotOperator,
                                        argument: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 10",
                                                        end: "line: 1, column: 11",
                                                        source: "b",
                                                    },
                                                },
                                                name: "b",
                                            },
                                        ),
                                    },
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
fn logical_operators_precedence_3() {
    let mut p = Parser::new(r#"not a and b"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 12",
                    source: "not a and b",
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
                                end: "line: 1, column: 12",
                                source: "not a and b",
                            },
                        },
                        expression: Logical(
                            LogicalExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 12",
                                        source: "not a and b",
                                    },
                                },
                                operator: AndOperator,
                                left: Unary(
                                    UnaryExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 1",
                                                end: "line: 1, column: 6",
                                                source: "not a",
                                            },
                                        },
                                        operator: NotOperator,
                                        argument: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 5",
                                                        end: "line: 1, column: 6",
                                                        source: "a",
                                                    },
                                                },
                                                name: "a",
                                            },
                                        ),
                                    },
                                ),
                                right: Identifier(
                                    Identifier {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 11",
                                                end: "line: 1, column: 12",
                                                source: "b",
                                            },
                                        },
                                        name: "b",
                                    },
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
fn logical_operators_precedence_4() {
    let mut p = Parser::new(r#"a and not b"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 12",
                    source: "a and not b",
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
                                end: "line: 1, column: 12",
                                source: "a and not b",
                            },
                        },
                        expression: Logical(
                            LogicalExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 12",
                                        source: "a and not b",
                                    },
                                },
                                operator: AndOperator,
                                left: Identifier(
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
                                right: Unary(
                                    UnaryExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 7",
                                                end: "line: 1, column: 12",
                                                source: "not b",
                                            },
                                        },
                                        operator: NotOperator,
                                        argument: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 11",
                                                        end: "line: 1, column: 12",
                                                        source: "b",
                                                    },
                                                },
                                                name: "b",
                                            },
                                        ),
                                    },
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
fn logical_operators_precedence_5() {
    let mut p = Parser::new(r#"a and b or c"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 13",
                    source: "a and b or c",
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
                                source: "a and b or c",
                            },
                        },
                        expression: Logical(
                            LogicalExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 13",
                                        source: "a and b or c",
                                    },
                                },
                                operator: OrOperator,
                                left: Logical(
                                    LogicalExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 1",
                                                end: "line: 1, column: 8",
                                                source: "a and b",
                                            },
                                        },
                                        operator: AndOperator,
                                        left: Identifier(
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
                                        right: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 7",
                                                        end: "line: 1, column: 8",
                                                        source: "b",
                                                    },
                                                },
                                                name: "b",
                                            },
                                        ),
                                    },
                                ),
                                right: Identifier(
                                    Identifier {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 12",
                                                end: "line: 1, column: 13",
                                                source: "c",
                                            },
                                        },
                                        name: "c",
                                    },
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
fn logical_operators_precedence_6() {
    let mut p = Parser::new(r#"a or b and c"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 13",
                    source: "a or b and c",
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
                                source: "a or b and c",
                            },
                        },
                        expression: Logical(
                            LogicalExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 13",
                                        source: "a or b and c",
                                    },
                                },
                                operator: OrOperator,
                                left: Identifier(
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
                                right: Logical(
                                    LogicalExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 6",
                                                end: "line: 1, column: 13",
                                                source: "b and c",
                                            },
                                        },
                                        operator: AndOperator,
                                        left: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 6",
                                                        end: "line: 1, column: 7",
                                                        source: "b",
                                                    },
                                                },
                                                name: "b",
                                            },
                                        ),
                                        right: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 12",
                                                        end: "line: 1, column: 13",
                                                        source: "c",
                                                    },
                                                },
                                                name: "c",
                                            },
                                        ),
                                    },
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
fn logical_operators_precedence_7() {
    let mut p = Parser::new(r#"not (a or b)"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 13",
                    source: "not (a or b)",
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
                                source: "not (a or b)",
                            },
                        },
                        expression: Unary(
                            UnaryExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 13",
                                        source: "not (a or b)",
                                    },
                                },
                                operator: NotOperator,
                                argument: Paren(
                                    ParenExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 5",
                                                end: "line: 1, column: 13",
                                                source: "(a or b)",
                                            },
                                        },
                                        lparen: [],
                                        expression: Logical(
                                            LogicalExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 6",
                                                        end: "line: 1, column: 12",
                                                        source: "a or b",
                                                    },
                                                },
                                                operator: OrOperator,
                                                left: Identifier(
                                                    Identifier {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 1, column: 6",
                                                                end: "line: 1, column: 7",
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
                                                                start: "line: 1, column: 11",
                                                                end: "line: 1, column: 12",
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
                    },
                ),
            ],
            eof: [],
        }
    "#]]
    .assert_debug_eq(&parsed);
}

#[test]
fn logical_operators_precedence_8() {
    let mut p = Parser::new(r#"not (a and b)"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 14",
                    source: "not (a and b)",
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
                                end: "line: 1, column: 14",
                                source: "not (a and b)",
                            },
                        },
                        expression: Unary(
                            UnaryExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 14",
                                        source: "not (a and b)",
                                    },
                                },
                                operator: NotOperator,
                                argument: Paren(
                                    ParenExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 5",
                                                end: "line: 1, column: 14",
                                                source: "(a and b)",
                                            },
                                        },
                                        lparen: [],
                                        expression: Logical(
                                            LogicalExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 6",
                                                        end: "line: 1, column: 13",
                                                        source: "a and b",
                                                    },
                                                },
                                                operator: AndOperator,
                                                left: Identifier(
                                                    Identifier {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 1, column: 6",
                                                                end: "line: 1, column: 7",
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
                                        rparen: [],
                                    },
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
fn logical_operators_precedence_9() {
    let mut p = Parser::new(r#"(a or b) and c"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 15",
                    source: "(a or b) and c",
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
                                end: "line: 1, column: 15",
                                source: "(a or b) and c",
                            },
                        },
                        expression: Logical(
                            LogicalExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 15",
                                        source: "(a or b) and c",
                                    },
                                },
                                operator: AndOperator,
                                left: Paren(
                                    ParenExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 1",
                                                end: "line: 1, column: 9",
                                                source: "(a or b)",
                                            },
                                        },
                                        lparen: [],
                                        expression: Logical(
                                            LogicalExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 2",
                                                        end: "line: 1, column: 8",
                                                        source: "a or b",
                                                    },
                                                },
                                                operator: OrOperator,
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
                                                                start: "line: 1, column: 7",
                                                                end: "line: 1, column: 8",
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
                                right: Identifier(
                                    Identifier {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 14",
                                                end: "line: 1, column: 15",
                                                source: "c",
                                            },
                                        },
                                        name: "c",
                                    },
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
fn logical_operators_precedence_10() {
    let mut p = Parser::new(r#"a and (b or c)"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 15",
                    source: "a and (b or c)",
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
                                end: "line: 1, column: 15",
                                source: "a and (b or c)",
                            },
                        },
                        expression: Logical(
                            LogicalExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 15",
                                        source: "a and (b or c)",
                                    },
                                },
                                operator: AndOperator,
                                left: Identifier(
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
                                right: Paren(
                                    ParenExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 7",
                                                end: "line: 1, column: 15",
                                                source: "(b or c)",
                                            },
                                        },
                                        lparen: [],
                                        expression: Logical(
                                            LogicalExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 8",
                                                        end: "line: 1, column: 14",
                                                        source: "b or c",
                                                    },
                                                },
                                                operator: OrOperator,
                                                left: Identifier(
                                                    Identifier {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 1, column: 8",
                                                                end: "line: 1, column: 9",
                                                                source: "b",
                                                            },
                                                        },
                                                        name: "b",
                                                    },
                                                ),
                                                right: Identifier(
                                                    Identifier {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 1, column: 13",
                                                                end: "line: 1, column: 14",
                                                                source: "c",
                                                            },
                                                        },
                                                        name: "c",
                                                    },
                                                ),
                                            },
                                        ),
                                        rparen: [],
                                    },
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
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 2, column: 15",
                    source: "not (a and b)\n(a or b) and c",
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
                                end: "line: 2, column: 15",
                                source: "not (a and b)\n(a or b) and c",
                            },
                        },
                        expression: Logical(
                            LogicalExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 2, column: 15",
                                        source: "not (a and b)\n(a or b) and c",
                                    },
                                },
                                operator: AndOperator,
                                left: Unary(
                                    UnaryExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 1",
                                                end: "line: 2, column: 9",
                                                source: "not (a and b)\n(a or b)",
                                            },
                                        },
                                        operator: NotOperator,
                                        argument: Call(
                                            CallExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 5",
                                                        end: "line: 2, column: 9",
                                                        source: "(a and b)\n(a or b)",
                                                    },
                                                    errors: [
                                                        "expected comma in property list, got OR",
                                                    ],
                                                },
                                                callee: Paren(
                                                    ParenExpr {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 1, column: 5",
                                                                end: "line: 1, column: 14",
                                                                source: "(a and b)",
                                                            },
                                                        },
                                                        lparen: [],
                                                        expression: Logical(
                                                            LogicalExpr {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 1, column: 6",
                                                                        end: "line: 1, column: 13",
                                                                        source: "a and b",
                                                                    },
                                                                },
                                                                operator: AndOperator,
                                                                left: Identifier(
                                                                    Identifier {
                                                                        base: BaseNode {
                                                                            location: SourceLocation {
                                                                                start: "line: 1, column: 6",
                                                                                end: "line: 1, column: 7",
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
                                                        rparen: [],
                                                    },
                                                ),
                                                lparen: [],
                                                arguments: [
                                                    Object(
                                                        ObjectExpr {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 2, column: 2",
                                                                    end: "line: 2, column: 8",
                                                                    source: "a or b",
                                                                },
                                                            },
                                                            lbrace: [],
                                                            with: None,
                                                            properties: [
                                                                Property {
                                                                    base: BaseNode {
                                                                        location: SourceLocation {
                                                                            start: "line: 2, column: 2",
                                                                            end: "line: 2, column: 3",
                                                                            source: "a",
                                                                        },
                                                                    },
                                                                    key: Identifier(
                                                                        Identifier {
                                                                            base: BaseNode {
                                                                                location: SourceLocation {
                                                                                    start: "line: 2, column: 2",
                                                                                    end: "line: 2, column: 3",
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
                                                                            start: "line: 2, column: 4",
                                                                            end: "line: 2, column: 8",
                                                                            source: "or b",
                                                                        },
                                                                        errors: [
                                                                            "unexpected token for property key: OR (or)",
                                                                        ],
                                                                    },
                                                                    key: StringLit(
                                                                        StringLit {
                                                                            base: BaseNode {
                                                                                location: SourceLocation {
                                                                                    start: "line: 2, column: 4",
                                                                                    end: "line: 2, column: 4",
                                                                                    source: "",
                                                                                },
                                                                            },
                                                                            value: "<invalid>",
                                                                        },
                                                                    ),
                                                                    separator: [],
                                                                    value: None,
                                                                    comma: [],
                                                                },
                                                            ],
                                                            rbrace: [],
                                                        },
                                                    ),
                                                ],
                                                rparen: [],
                                            },
                                        ),
                                    },
                                ),
                                right: Identifier(
                                    Identifier {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 2, column: 14",
                                                end: "line: 2, column: 15",
                                                source: "c",
                                            },
                                        },
                                        name: "c",
                                    },
                                ),
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
fn binary_expression() {
    let mut p = Parser::new(r#"_value < 10.0"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 14",
                    source: "_value < 10.0",
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
                                end: "line: 1, column: 14",
                                source: "_value < 10.0",
                            },
                        },
                        expression: Binary(
                            BinaryExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 14",
                                        source: "_value < 10.0",
                                    },
                                },
                                operator: LessThanOperator,
                                left: Identifier(
                                    Identifier {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 1",
                                                end: "line: 1, column: 7",
                                                source: "_value",
                                            },
                                        },
                                        name: "_value",
                                    },
                                ),
                                right: Float(
                                    FloatLit {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 10",
                                                end: "line: 1, column: 14",
                                                source: "10.0",
                                            },
                                        },
                                        value: NotNan(
                                            10.0,
                                        ),
                                    },
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
fn member_expression_binary_expression() {
    let mut p = Parser::new(r#"r._value < 10.0"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 16",
                    source: "r._value < 10.0",
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
                                end: "line: 1, column: 16",
                                source: "r._value < 10.0",
                            },
                        },
                        expression: Binary(
                            BinaryExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 16",
                                        source: "r._value < 10.0",
                                    },
                                },
                                operator: LessThanOperator,
                                left: Member(
                                    MemberExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 1",
                                                end: "line: 1, column: 9",
                                                source: "r._value",
                                            },
                                        },
                                        object: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 1",
                                                        end: "line: 1, column: 2",
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
                                                        start: "line: 1, column: 3",
                                                        end: "line: 1, column: 9",
                                                        source: "_value",
                                                    },
                                                },
                                                name: "_value",
                                            },
                                        ),
                                        rbrack: [],
                                    },
                                ),
                                right: Float(
                                    FloatLit {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 12",
                                                end: "line: 1, column: 16",
                                                source: "10.0",
                                            },
                                        },
                                        value: NotNan(
                                            10.0,
                                        ),
                                    },
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
fn var_as_binary_expression_of_other_vars() {
    let mut p = Parser::new(
        r#"a = 1
            b = 2
            c = a + b
            d = a"#,
    );
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 4, column: 18",
                    source: "a = 1\n            b = 2\n            c = a + b\n            d = a",
                },
            },
            name: "",
            metadata: "parser-type=rust",
            package: None,
            imports: [],
            body: [
                Variable(
                    VariableAssgn {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 6",
                                source: "a = 1",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 1",
                                    end: "line: 1, column: 2",
                                    source: "a",
                                },
                            },
                            name: "a",
                        },
                        init: Integer(
                            IntegerLit {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 5",
                                        end: "line: 1, column: 6",
                                        source: "1",
                                    },
                                },
                                value: 1,
                            },
                        ),
                    },
                ),
                Variable(
                    VariableAssgn {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 2, column: 13",
                                end: "line: 2, column: 18",
                                source: "b = 2",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 2, column: 13",
                                    end: "line: 2, column: 14",
                                    source: "b",
                                },
                            },
                            name: "b",
                        },
                        init: Integer(
                            IntegerLit {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 2, column: 17",
                                        end: "line: 2, column: 18",
                                        source: "2",
                                    },
                                },
                                value: 2,
                            },
                        ),
                    },
                ),
                Variable(
                    VariableAssgn {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 3, column: 13",
                                end: "line: 3, column: 22",
                                source: "c = a + b",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 3, column: 13",
                                    end: "line: 3, column: 14",
                                    source: "c",
                                },
                            },
                            name: "c",
                        },
                        init: Binary(
                            BinaryExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 3, column: 17",
                                        end: "line: 3, column: 22",
                                        source: "a + b",
                                    },
                                },
                                operator: AdditionOperator,
                                left: Identifier(
                                    Identifier {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 3, column: 17",
                                                end: "line: 3, column: 18",
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
                                                start: "line: 3, column: 21",
                                                end: "line: 3, column: 22",
                                                source: "b",
                                            },
                                        },
                                        name: "b",
                                    },
                                ),
                            },
                        ),
                    },
                ),
                Variable(
                    VariableAssgn {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 4, column: 13",
                                end: "line: 4, column: 18",
                                source: "d = a",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 4, column: 13",
                                    end: "line: 4, column: 14",
                                    source: "d",
                                },
                            },
                            name: "d",
                        },
                        init: Identifier(
                            Identifier {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 4, column: 17",
                                        end: "line: 4, column: 18",
                                        source: "a",
                                    },
                                },
                                name: "a",
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
fn var_as_unary_expression_of_other_vars() {
    let mut p = Parser::new(
        r#"a = 5
            c = -a"#,
    );
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 2, column: 19",
                    source: "a = 5\n            c = -a",
                },
            },
            name: "",
            metadata: "parser-type=rust",
            package: None,
            imports: [],
            body: [
                Variable(
                    VariableAssgn {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 6",
                                source: "a = 5",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 1",
                                    end: "line: 1, column: 2",
                                    source: "a",
                                },
                            },
                            name: "a",
                        },
                        init: Integer(
                            IntegerLit {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 5",
                                        end: "line: 1, column: 6",
                                        source: "5",
                                    },
                                },
                                value: 5,
                            },
                        ),
                    },
                ),
                Variable(
                    VariableAssgn {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 2, column: 13",
                                end: "line: 2, column: 19",
                                source: "c = -a",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 2, column: 13",
                                    end: "line: 2, column: 14",
                                    source: "c",
                                },
                            },
                            name: "c",
                        },
                        init: Unary(
                            UnaryExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 2, column: 17",
                                        end: "line: 2, column: 19",
                                        source: "-a",
                                    },
                                },
                                operator: SubtractionOperator,
                                argument: Identifier(
                                    Identifier {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 2, column: 18",
                                                end: "line: 2, column: 19",
                                                source: "a",
                                            },
                                        },
                                        name: "a",
                                    },
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
fn var_as_both_binary_and_unary_expressions() {
    let mut p = Parser::new(
        r#"a = 5
            c = 10 * -a"#,
    );
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 2, column: 24",
                    source: "a = 5\n            c = 10 * -a",
                },
            },
            name: "",
            metadata: "parser-type=rust",
            package: None,
            imports: [],
            body: [
                Variable(
                    VariableAssgn {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 6",
                                source: "a = 5",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 1",
                                    end: "line: 1, column: 2",
                                    source: "a",
                                },
                            },
                            name: "a",
                        },
                        init: Integer(
                            IntegerLit {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 5",
                                        end: "line: 1, column: 6",
                                        source: "5",
                                    },
                                },
                                value: 5,
                            },
                        ),
                    },
                ),
                Variable(
                    VariableAssgn {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 2, column: 13",
                                end: "line: 2, column: 24",
                                source: "c = 10 * -a",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 2, column: 13",
                                    end: "line: 2, column: 14",
                                    source: "c",
                                },
                            },
                            name: "c",
                        },
                        init: Binary(
                            BinaryExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 2, column: 17",
                                        end: "line: 2, column: 24",
                                        source: "10 * -a",
                                    },
                                },
                                operator: MultiplicationOperator,
                                left: Integer(
                                    IntegerLit {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 2, column: 17",
                                                end: "line: 2, column: 19",
                                                source: "10",
                                            },
                                        },
                                        value: 10,
                                    },
                                ),
                                right: Unary(
                                    UnaryExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 2, column: 22",
                                                end: "line: 2, column: 24",
                                                source: "-a",
                                            },
                                        },
                                        operator: SubtractionOperator,
                                        argument: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 2, column: 23",
                                                        end: "line: 2, column: 24",
                                                        source: "a",
                                                    },
                                                },
                                                name: "a",
                                            },
                                        ),
                                    },
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
fn unary_expressions_within_logical_expression() {
    let mut p = Parser::new(
        r#"a = 5.0
            10.0 * -a == -0.5 or a == 6.0"#,
    );
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 2, column: 42",
                    source: "a = 5.0\n            10.0 * -a == -0.5 or a == 6.0",
                },
            },
            name: "",
            metadata: "parser-type=rust",
            package: None,
            imports: [],
            body: [
                Variable(
                    VariableAssgn {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 8",
                                source: "a = 5.0",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 1",
                                    end: "line: 1, column: 2",
                                    source: "a",
                                },
                            },
                            name: "a",
                        },
                        init: Float(
                            FloatLit {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 5",
                                        end: "line: 1, column: 8",
                                        source: "5.0",
                                    },
                                },
                                value: NotNan(
                                    5.0,
                                ),
                            },
                        ),
                    },
                ),
                Expr(
                    ExprStmt {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 2, column: 13",
                                end: "line: 2, column: 42",
                                source: "10.0 * -a == -0.5 or a == 6.0",
                            },
                        },
                        expression: Logical(
                            LogicalExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 2, column: 13",
                                        end: "line: 2, column: 42",
                                        source: "10.0 * -a == -0.5 or a == 6.0",
                                    },
                                },
                                operator: OrOperator,
                                left: Binary(
                                    BinaryExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 2, column: 13",
                                                end: "line: 2, column: 30",
                                                source: "10.0 * -a == -0.5",
                                            },
                                        },
                                        operator: EqualOperator,
                                        left: Binary(
                                            BinaryExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 2, column: 13",
                                                        end: "line: 2, column: 22",
                                                        source: "10.0 * -a",
                                                    },
                                                },
                                                operator: MultiplicationOperator,
                                                left: Float(
                                                    FloatLit {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 2, column: 13",
                                                                end: "line: 2, column: 17",
                                                                source: "10.0",
                                                            },
                                                        },
                                                        value: NotNan(
                                                            10.0,
                                                        ),
                                                    },
                                                ),
                                                right: Unary(
                                                    UnaryExpr {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 2, column: 20",
                                                                end: "line: 2, column: 22",
                                                                source: "-a",
                                                            },
                                                        },
                                                        operator: SubtractionOperator,
                                                        argument: Identifier(
                                                            Identifier {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 2, column: 21",
                                                                        end: "line: 2, column: 22",
                                                                        source: "a",
                                                                    },
                                                                },
                                                                name: "a",
                                                            },
                                                        ),
                                                    },
                                                ),
                                            },
                                        ),
                                        right: Unary(
                                            UnaryExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 2, column: 26",
                                                        end: "line: 2, column: 30",
                                                        source: "-0.5",
                                                    },
                                                },
                                                operator: SubtractionOperator,
                                                argument: Float(
                                                    FloatLit {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 2, column: 27",
                                                                end: "line: 2, column: 30",
                                                                source: "0.5",
                                                            },
                                                        },
                                                        value: NotNan(
                                                            0.5,
                                                        ),
                                                    },
                                                ),
                                            },
                                        ),
                                    },
                                ),
                                right: Binary(
                                    BinaryExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 2, column: 34",
                                                end: "line: 2, column: 42",
                                                source: "a == 6.0",
                                            },
                                        },
                                        operator: EqualOperator,
                                        left: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 2, column: 34",
                                                        end: "line: 2, column: 35",
                                                        source: "a",
                                                    },
                                                },
                                                name: "a",
                                            },
                                        ),
                                        right: Float(
                                            FloatLit {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 2, column: 39",
                                                        end: "line: 2, column: 42",
                                                        source: "6.0",
                                                    },
                                                },
                                                value: NotNan(
                                                    6.0,
                                                ),
                                            },
                                        ),
                                    },
                                ),
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
fn unary_expression_with_member_expression() {
    let mut p = Parser::new(r#"not m.b"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 8",
                    source: "not m.b",
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
                                end: "line: 1, column: 8",
                                source: "not m.b",
                            },
                        },
                        expression: Unary(
                            UnaryExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 8",
                                        source: "not m.b",
                                    },
                                },
                                operator: NotOperator,
                                argument: Member(
                                    MemberExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 5",
                                                end: "line: 1, column: 8",
                                                source: "m.b",
                                            },
                                        },
                                        object: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 5",
                                                        end: "line: 1, column: 6",
                                                        source: "m",
                                                    },
                                                },
                                                name: "m",
                                            },
                                        ),
                                        lbrack: [],
                                        property: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 7",
                                                        end: "line: 1, column: 8",
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
                    },
                ),
            ],
            eof: [],
        }
    "#]]
    .assert_debug_eq(&parsed);
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
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 6, column: 13",
                    source: "// define a\na = 5.0\n// eval this\n10.0 * -a == -0.5\n\t// or this\n\tor a == 6.0",
                },
            },
            name: "",
            metadata: "parser-type=rust",
            package: None,
            imports: [],
            body: [
                Variable(
                    VariableAssgn {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 2, column: 1",
                                end: "line: 2, column: 8",
                                source: "a = 5.0",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 2, column: 1",
                                    end: "line: 2, column: 2",
                                    source: "a",
                                },
                                comments: [
                                    Comment {
                                        text: "// define a\n",
                                    },
                                ],
                            },
                            name: "a",
                        },
                        init: Float(
                            FloatLit {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 2, column: 5",
                                        end: "line: 2, column: 8",
                                        source: "5.0",
                                    },
                                },
                                value: NotNan(
                                    5.0,
                                ),
                            },
                        ),
                    },
                ),
                Expr(
                    ExprStmt {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 4, column: 1",
                                end: "line: 6, column: 13",
                                source: "10.0 * -a == -0.5\n\t// or this\n\tor a == 6.0",
                            },
                        },
                        expression: Logical(
                            LogicalExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 4, column: 1",
                                        end: "line: 6, column: 13",
                                        source: "10.0 * -a == -0.5\n\t// or this\n\tor a == 6.0",
                                    },
                                    comments: [
                                        Comment {
                                            text: "// or this\n",
                                        },
                                    ],
                                },
                                operator: OrOperator,
                                left: Binary(
                                    BinaryExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 4, column: 1",
                                                end: "line: 4, column: 18",
                                                source: "10.0 * -a == -0.5",
                                            },
                                        },
                                        operator: EqualOperator,
                                        left: Binary(
                                            BinaryExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 4, column: 1",
                                                        end: "line: 4, column: 10",
                                                        source: "10.0 * -a",
                                                    },
                                                },
                                                operator: MultiplicationOperator,
                                                left: Float(
                                                    FloatLit {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 4, column: 1",
                                                                end: "line: 4, column: 5",
                                                                source: "10.0",
                                                            },
                                                            comments: [
                                                                Comment {
                                                                    text: "// eval this\n",
                                                                },
                                                            ],
                                                        },
                                                        value: NotNan(
                                                            10.0,
                                                        ),
                                                    },
                                                ),
                                                right: Unary(
                                                    UnaryExpr {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 4, column: 8",
                                                                end: "line: 4, column: 10",
                                                                source: "-a",
                                                            },
                                                        },
                                                        operator: SubtractionOperator,
                                                        argument: Identifier(
                                                            Identifier {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 4, column: 9",
                                                                        end: "line: 4, column: 10",
                                                                        source: "a",
                                                                    },
                                                                },
                                                                name: "a",
                                                            },
                                                        ),
                                                    },
                                                ),
                                            },
                                        ),
                                        right: Unary(
                                            UnaryExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 4, column: 14",
                                                        end: "line: 4, column: 18",
                                                        source: "-0.5",
                                                    },
                                                },
                                                operator: SubtractionOperator,
                                                argument: Float(
                                                    FloatLit {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 4, column: 15",
                                                                end: "line: 4, column: 18",
                                                                source: "0.5",
                                                            },
                                                        },
                                                        value: NotNan(
                                                            0.5,
                                                        ),
                                                    },
                                                ),
                                            },
                                        ),
                                    },
                                ),
                                right: Binary(
                                    BinaryExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 6, column: 5",
                                                end: "line: 6, column: 13",
                                                source: "a == 6.0",
                                            },
                                        },
                                        operator: EqualOperator,
                                        left: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 6, column: 5",
                                                        end: "line: 6, column: 6",
                                                        source: "a",
                                                    },
                                                },
                                                name: "a",
                                            },
                                        ),
                                        right: Float(
                                            FloatLit {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 6, column: 10",
                                                        end: "line: 6, column: 13",
                                                        source: "6.0",
                                                    },
                                                },
                                                value: NotNan(
                                                    6.0,
                                                ),
                                            },
                                        ),
                                    },
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
fn mix_unary_logical_and_binary_expressions() {
    let mut p = Parser::new(
        r#"
            not (f() == 6.0 * x) or fail()"#,
    );
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 2, column: 43",
                    source: "\n            not (f() == 6.0 * x) or fail()",
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
                                start: "line: 2, column: 13",
                                end: "line: 2, column: 43",
                                source: "not (f() == 6.0 * x) or fail()",
                            },
                        },
                        expression: Logical(
                            LogicalExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 2, column: 13",
                                        end: "line: 2, column: 43",
                                        source: "not (f() == 6.0 * x) or fail()",
                                    },
                                },
                                operator: OrOperator,
                                left: Unary(
                                    UnaryExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 2, column: 13",
                                                end: "line: 2, column: 33",
                                                source: "not (f() == 6.0 * x)",
                                            },
                                        },
                                        operator: NotOperator,
                                        argument: Paren(
                                            ParenExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 2, column: 17",
                                                        end: "line: 2, column: 33",
                                                        source: "(f() == 6.0 * x)",
                                                    },
                                                },
                                                lparen: [],
                                                expression: Binary(
                                                    BinaryExpr {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 2, column: 18",
                                                                end: "line: 2, column: 32",
                                                                source: "f() == 6.0 * x",
                                                            },
                                                        },
                                                        operator: EqualOperator,
                                                        left: Call(
                                                            CallExpr {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 2, column: 18",
                                                                        end: "line: 2, column: 21",
                                                                        source: "f()",
                                                                    },
                                                                },
                                                                callee: Identifier(
                                                                    Identifier {
                                                                        base: BaseNode {
                                                                            location: SourceLocation {
                                                                                start: "line: 2, column: 18",
                                                                                end: "line: 2, column: 19",
                                                                                source: "f",
                                                                            },
                                                                        },
                                                                        name: "f",
                                                                    },
                                                                ),
                                                                lparen: [],
                                                                arguments: [],
                                                                rparen: [],
                                                            },
                                                        ),
                                                        right: Binary(
                                                            BinaryExpr {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 2, column: 25",
                                                                        end: "line: 2, column: 32",
                                                                        source: "6.0 * x",
                                                                    },
                                                                },
                                                                operator: MultiplicationOperator,
                                                                left: Float(
                                                                    FloatLit {
                                                                        base: BaseNode {
                                                                            location: SourceLocation {
                                                                                start: "line: 2, column: 25",
                                                                                end: "line: 2, column: 28",
                                                                                source: "6.0",
                                                                            },
                                                                        },
                                                                        value: NotNan(
                                                                            6.0,
                                                                        ),
                                                                    },
                                                                ),
                                                                right: Identifier(
                                                                    Identifier {
                                                                        base: BaseNode {
                                                                            location: SourceLocation {
                                                                                start: "line: 2, column: 31",
                                                                                end: "line: 2, column: 32",
                                                                                source: "x",
                                                                            },
                                                                        },
                                                                        name: "x",
                                                                    },
                                                                ),
                                                            },
                                                        ),
                                                    },
                                                ),
                                                rparen: [],
                                            },
                                        ),
                                    },
                                ),
                                right: Call(
                                    CallExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 2, column: 37",
                                                end: "line: 2, column: 43",
                                                source: "fail()",
                                            },
                                        },
                                        callee: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 2, column: 37",
                                                        end: "line: 2, column: 41",
                                                        source: "fail",
                                                    },
                                                },
                                                name: "fail",
                                            },
                                        ),
                                        lparen: [],
                                        arguments: [],
                                        rparen: [],
                                    },
                                ),
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
fn mix_unary_logical_and_binary_expressions_with_extra_parens() {
    let mut p = Parser::new(
        r#"
            (not (f() == 6.0 * x) or fail())"#,
    );
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 2, column: 45",
                    source: "\n            (not (f() == 6.0 * x) or fail())",
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
                                start: "line: 2, column: 13",
                                end: "line: 2, column: 45",
                                source: "(not (f() == 6.0 * x) or fail())",
                            },
                        },
                        expression: Paren(
                            ParenExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 2, column: 13",
                                        end: "line: 2, column: 45",
                                        source: "(not (f() == 6.0 * x) or fail())",
                                    },
                                },
                                lparen: [],
                                expression: Logical(
                                    LogicalExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 2, column: 14",
                                                end: "line: 2, column: 44",
                                                source: "not (f() == 6.0 * x) or fail()",
                                            },
                                        },
                                        operator: OrOperator,
                                        left: Unary(
                                            UnaryExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 2, column: 14",
                                                        end: "line: 2, column: 34",
                                                        source: "not (f() == 6.0 * x)",
                                                    },
                                                },
                                                operator: NotOperator,
                                                argument: Paren(
                                                    ParenExpr {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 2, column: 18",
                                                                end: "line: 2, column: 34",
                                                                source: "(f() == 6.0 * x)",
                                                            },
                                                        },
                                                        lparen: [],
                                                        expression: Binary(
                                                            BinaryExpr {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 2, column: 19",
                                                                        end: "line: 2, column: 33",
                                                                        source: "f() == 6.0 * x",
                                                                    },
                                                                },
                                                                operator: EqualOperator,
                                                                left: Call(
                                                                    CallExpr {
                                                                        base: BaseNode {
                                                                            location: SourceLocation {
                                                                                start: "line: 2, column: 19",
                                                                                end: "line: 2, column: 22",
                                                                                source: "f()",
                                                                            },
                                                                        },
                                                                        callee: Identifier(
                                                                            Identifier {
                                                                                base: BaseNode {
                                                                                    location: SourceLocation {
                                                                                        start: "line: 2, column: 19",
                                                                                        end: "line: 2, column: 20",
                                                                                        source: "f",
                                                                                    },
                                                                                },
                                                                                name: "f",
                                                                            },
                                                                        ),
                                                                        lparen: [],
                                                                        arguments: [],
                                                                        rparen: [],
                                                                    },
                                                                ),
                                                                right: Binary(
                                                                    BinaryExpr {
                                                                        base: BaseNode {
                                                                            location: SourceLocation {
                                                                                start: "line: 2, column: 26",
                                                                                end: "line: 2, column: 33",
                                                                                source: "6.0 * x",
                                                                            },
                                                                        },
                                                                        operator: MultiplicationOperator,
                                                                        left: Float(
                                                                            FloatLit {
                                                                                base: BaseNode {
                                                                                    location: SourceLocation {
                                                                                        start: "line: 2, column: 26",
                                                                                        end: "line: 2, column: 29",
                                                                                        source: "6.0",
                                                                                    },
                                                                                },
                                                                                value: NotNan(
                                                                                    6.0,
                                                                                ),
                                                                            },
                                                                        ),
                                                                        right: Identifier(
                                                                            Identifier {
                                                                                base: BaseNode {
                                                                                    location: SourceLocation {
                                                                                        start: "line: 2, column: 32",
                                                                                        end: "line: 2, column: 33",
                                                                                        source: "x",
                                                                                    },
                                                                                },
                                                                                name: "x",
                                                                            },
                                                                        ),
                                                                    },
                                                                ),
                                                            },
                                                        ),
                                                        rparen: [],
                                                    },
                                                ),
                                            },
                                        ),
                                        right: Call(
                                            CallExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 2, column: 38",
                                                        end: "line: 2, column: 44",
                                                        source: "fail()",
                                                    },
                                                },
                                                callee: Identifier(
                                                    Identifier {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 2, column: 38",
                                                                end: "line: 2, column: 42",
                                                                source: "fail",
                                                            },
                                                        },
                                                        name: "fail",
                                                    },
                                                ),
                                                lparen: [],
                                                arguments: [],
                                                rparen: [],
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
fn modulo_op_ints() {
    let mut p = Parser::new(r#"3 % 8"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 6",
                    source: "3 % 8",
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
                                source: "3 % 8",
                            },
                        },
                        expression: Binary(
                            BinaryExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 6",
                                        source: "3 % 8",
                                    },
                                },
                                operator: ModuloOperator,
                                left: Integer(
                                    IntegerLit {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 1",
                                                end: "line: 1, column: 2",
                                                source: "3",
                                            },
                                        },
                                        value: 3,
                                    },
                                ),
                                right: Integer(
                                    IntegerLit {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 5",
                                                end: "line: 1, column: 6",
                                                source: "8",
                                            },
                                        },
                                        value: 8,
                                    },
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
fn modulo_op_floats() {
    let mut p = Parser::new(r#"8.3 % 3.1"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 10",
                    source: "8.3 % 3.1",
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
                                end: "line: 1, column: 10",
                                source: "8.3 % 3.1",
                            },
                        },
                        expression: Binary(
                            BinaryExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 10",
                                        source: "8.3 % 3.1",
                                    },
                                },
                                operator: ModuloOperator,
                                left: Float(
                                    FloatLit {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 1",
                                                end: "line: 1, column: 4",
                                                source: "8.3",
                                            },
                                        },
                                        value: NotNan(
                                            8.3,
                                        ),
                                    },
                                ),
                                right: Float(
                                    FloatLit {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 7",
                                                end: "line: 1, column: 10",
                                                source: "3.1",
                                            },
                                        },
                                        value: NotNan(
                                            3.1,
                                        ),
                                    },
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
fn power_op() {
    let mut p = Parser::new(r#"2 ^ 4"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 6",
                    source: "2 ^ 4",
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
                                source: "2 ^ 4",
                            },
                        },
                        expression: Binary(
                            BinaryExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 6",
                                        source: "2 ^ 4",
                                    },
                                },
                                operator: PowerOperator,
                                left: Integer(
                                    IntegerLit {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 1",
                                                end: "line: 1, column: 2",
                                                source: "2",
                                            },
                                        },
                                        value: 2,
                                    },
                                ),
                                right: Integer(
                                    IntegerLit {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 5",
                                                end: "line: 1, column: 6",
                                                source: "4",
                                            },
                                        },
                                        value: 4,
                                    },
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
