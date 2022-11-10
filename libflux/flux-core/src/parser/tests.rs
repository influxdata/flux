use pretty_assertions::assert_eq;

use super::*;
use crate::ast;

use expect_test::expect;

mod arrow_function;
mod attributes;
mod errors;
mod from;
mod literals;
mod objects;
mod operator_precedence;
mod property_list;
mod strings;
mod types;

fn test_file(source: &str, expect: expect_test::Expect) {
    let mut parser = Parser::new(source);
    let parsed = parser.parse_file("".to_string());

    expect.assert_debug_eq(&parsed);
}

/// Parsed ast roundtrips across the serde boundary and generates the same ast.
#[test]
fn parse_ast_roundtrip() {
    let mut p = Parser::new(
        r#"from(bucket: "an-bucket") |> range(start: -15m) |> filter(fn: (r) => r.field == "value")"#,
    );
    let ast = p.parse_file("".to_string());

    let serialized = serde_json::to_string(&ast).unwrap();
    let new_ast = serde_json::from_str(&serialized).unwrap();

    assert_eq!(ast, new_ast);
}

#[test]
fn parse_array_expr_no_rbrack() {
    let mut p = Parser::new(r#"group(columns: ["_time", "_field]", mode: "by")"#);
    let parsed = p.parse_file("".to_string());
    let node = ast::walk::Node::File(&parsed);
    ast::check::check(node).unwrap_err();
}

#[test]
fn parse_illegal_in_array_expr() {
    let mut p = Parser::new(
        r#"
strings.joinStr(arr: [r.unit, ": time", string(v: now()), "InfluxDB Task", $taskName, "last datapoint at:", string(v: r._time)], v: " ")
    "#,
    );
    let parsed = p.parse_file("".to_string());
    let node = ast::walk::Node::File(&parsed);
    ast::check::check(node).unwrap_err();
}

#[test]
fn parse_array_expr_no_comma() {
    let mut p = Parser::new(
        r#"
sort(columns: ["_time"], desc: true)|> limit(n: [object Object])
    "#,
    );
    let parsed = p.parse_file("".to_string());
    let node = ast::walk::Node::File(&parsed);
    ast::check::check(node).unwrap_err();
}

#[test]
fn parse_invalid_unicode_bare() {
    test_file(
        r#"®some string®"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 16",
                        source: "®some string®",
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
                                    end: "line: 1, column: 3",
                                    source: "®",
                                },
                            },
                            text: "®",
                        },
                    ),
                    Expr(
                        ExprStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 3",
                                    end: "line: 1, column: 7",
                                    source: "some",
                                },
                            },
                            expression: Identifier(
                                Identifier {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 3",
                                            end: "line: 1, column: 7",
                                            source: "some",
                                        },
                                    },
                                    name: "some",
                                },
                            ),
                        },
                    ),
                    Expr(
                        ExprStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 8",
                                    end: "line: 1, column: 14",
                                    source: "string",
                                },
                            },
                            expression: Identifier(
                                Identifier {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 8",
                                            end: "line: 1, column: 14",
                                            source: "string",
                                        },
                                    },
                                    name: "string",
                                },
                            ),
                        },
                    ),
                    Bad(
                        BadStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 14",
                                    end: "line: 1, column: 16",
                                    source: "®",
                                },
                            },
                            text: "®",
                        },
                    ),
                ],
                eof: [],
            }
        "#]],
    );
}

#[test]
fn parse_invalid_unicode_paren_wrapped() {
    test_file(
        r#"(‛some string‛)"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 20",
                        source: "(‛some string‛)",
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
                                    end: "line: 1, column: 20",
                                    source: "(‛some string‛)",
                                },
                            },
                            expression: Paren(
                                ParenExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 1",
                                            end: "line: 1, column: 20",
                                            source: "(‛some string‛)",
                                        },
                                        errors: [
                                            "invalid expression @1:16-1:19: ‛",
                                        ],
                                    },
                                    lparen: [],
                                    expression: Binary(
                                        BinaryExpr {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 5",
                                                    end: "line: 1, column: 16",
                                                    source: "some string",
                                                },
                                            },
                                            operator: InvalidOperator,
                                            left: Identifier(
                                                Identifier {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 5",
                                                            end: "line: 1, column: 9",
                                                            source: "some",
                                                        },
                                                        errors: [
                                                            "invalid expression @1:2-1:5: ‛",
                                                        ],
                                                    },
                                                    name: "some",
                                                },
                                            ),
                                            right: Identifier(
                                                Identifier {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 10",
                                                            end: "line: 1, column: 16",
                                                            source: "string",
                                                        },
                                                    },
                                                    name: "string",
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
        "#]],
    );
}

#[test]
fn parse_invalid_unicode_interspersed() {
    test_file(
        r#"®s®t®r®i®n®g"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 19",
                        source: "®s®t®r®i®n®g",
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
                                    end: "line: 1, column: 3",
                                    source: "®",
                                },
                            },
                            text: "®",
                        },
                    ),
                    Expr(
                        ExprStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 3",
                                    end: "line: 1, column: 4",
                                    source: "s",
                                },
                            },
                            expression: Identifier(
                                Identifier {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 3",
                                            end: "line: 1, column: 4",
                                            source: "s",
                                        },
                                    },
                                    name: "s",
                                },
                            ),
                        },
                    ),
                    Bad(
                        BadStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 4",
                                    end: "line: 1, column: 6",
                                    source: "®",
                                },
                            },
                            text: "®",
                        },
                    ),
                    Expr(
                        ExprStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 6",
                                    end: "line: 1, column: 7",
                                    source: "t",
                                },
                            },
                            expression: Identifier(
                                Identifier {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 6",
                                            end: "line: 1, column: 7",
                                            source: "t",
                                        },
                                    },
                                    name: "t",
                                },
                            ),
                        },
                    ),
                    Bad(
                        BadStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 7",
                                    end: "line: 1, column: 9",
                                    source: "®",
                                },
                            },
                            text: "®",
                        },
                    ),
                    Expr(
                        ExprStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 9",
                                    end: "line: 1, column: 10",
                                    source: "r",
                                },
                            },
                            expression: Identifier(
                                Identifier {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 9",
                                            end: "line: 1, column: 10",
                                            source: "r",
                                        },
                                    },
                                    name: "r",
                                },
                            ),
                        },
                    ),
                    Bad(
                        BadStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 10",
                                    end: "line: 1, column: 12",
                                    source: "®",
                                },
                            },
                            text: "®",
                        },
                    ),
                    Expr(
                        ExprStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 12",
                                    end: "line: 1, column: 13",
                                    source: "i",
                                },
                            },
                            expression: Identifier(
                                Identifier {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 12",
                                            end: "line: 1, column: 13",
                                            source: "i",
                                        },
                                    },
                                    name: "i",
                                },
                            ),
                        },
                    ),
                    Bad(
                        BadStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 13",
                                    end: "line: 1, column: 15",
                                    source: "®",
                                },
                            },
                            text: "®",
                        },
                    ),
                    Expr(
                        ExprStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 15",
                                    end: "line: 1, column: 16",
                                    source: "n",
                                },
                            },
                            expression: Identifier(
                                Identifier {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 15",
                                            end: "line: 1, column: 16",
                                            source: "n",
                                        },
                                    },
                                    name: "n",
                                },
                            ),
                        },
                    ),
                    Bad(
                        BadStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 16",
                                    end: "line: 1, column: 18",
                                    source: "®",
                                },
                            },
                            text: "®",
                        },
                    ),
                    Expr(
                        ExprStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 18",
                                    end: "line: 1, column: 19",
                                    source: "g",
                                },
                            },
                            expression: Identifier(
                                Identifier {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 18",
                                            end: "line: 1, column: 19",
                                            source: "g",
                                        },
                                    },
                                    name: "g",
                                },
                            ),
                        },
                    ),
                ],
                eof: [],
            }
        "#]],
    );
}

#[test]
fn parse_greedy_quotes_paren_wrapped() {
    test_file(
        r#"(“some string”)"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 20",
                        source: "(“some string”)",
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
                                    end: "line: 1, column: 20",
                                    source: "(“some string”)",
                                },
                            },
                            expression: Paren(
                                ParenExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 1",
                                            end: "line: 1, column: 20",
                                            source: "(“some string”)",
                                        },
                                        errors: [
                                            "invalid expression @1:16-1:19: ”",
                                        ],
                                    },
                                    lparen: [],
                                    expression: Binary(
                                        BinaryExpr {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 5",
                                                    end: "line: 1, column: 16",
                                                    source: "some string",
                                                },
                                            },
                                            operator: InvalidOperator,
                                            left: Identifier(
                                                Identifier {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 5",
                                                            end: "line: 1, column: 9",
                                                            source: "some",
                                                        },
                                                        errors: [
                                                            "invalid expression @1:2-1:5: “",
                                                        ],
                                                    },
                                                    name: "some",
                                                },
                                            ),
                                            right: Identifier(
                                                Identifier {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 10",
                                                            end: "line: 1, column: 16",
                                                            source: "string",
                                                        },
                                                    },
                                                    name: "string",
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
        "#]],
    );
}

#[test]
fn parse_greedy_quotes_bare() {
    test_file(
        r#"“some string”"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 18",
                        source: "“some string”",
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
                                    end: "line: 1, column: 4",
                                    source: "“",
                                },
                            },
                            text: "“",
                        },
                    ),
                    Expr(
                        ExprStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 4",
                                    end: "line: 1, column: 8",
                                    source: "some",
                                },
                            },
                            expression: Identifier(
                                Identifier {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 4",
                                            end: "line: 1, column: 8",
                                            source: "some",
                                        },
                                    },
                                    name: "some",
                                },
                            ),
                        },
                    ),
                    Expr(
                        ExprStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 9",
                                    end: "line: 1, column: 15",
                                    source: "string",
                                },
                            },
                            expression: Identifier(
                                Identifier {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 9",
                                            end: "line: 1, column: 15",
                                            source: "string",
                                        },
                                    },
                                    name: "string",
                                },
                            ),
                        },
                    ),
                    Bad(
                        BadStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 15",
                                    end: "line: 1, column: 18",
                                    source: "”",
                                },
                            },
                            text: "”",
                        },
                    ),
                ],
                eof: [],
            }
        "#]],
    );
}

#[test]
fn parse_greedy_quotes_interspersed() {
    test_file(
        r#"“s”t“r”i“n”g"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 25",
                        source: "“s”t“r”i“n”g",
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
                                    end: "line: 1, column: 4",
                                    source: "“",
                                },
                            },
                            text: "“",
                        },
                    ),
                    Expr(
                        ExprStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 4",
                                    end: "line: 1, column: 5",
                                    source: "s",
                                },
                            },
                            expression: Identifier(
                                Identifier {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 4",
                                            end: "line: 1, column: 5",
                                            source: "s",
                                        },
                                    },
                                    name: "s",
                                },
                            ),
                        },
                    ),
                    Bad(
                        BadStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 5",
                                    end: "line: 1, column: 8",
                                    source: "”",
                                },
                            },
                            text: "”",
                        },
                    ),
                    Expr(
                        ExprStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 8",
                                    end: "line: 1, column: 9",
                                    source: "t",
                                },
                            },
                            expression: Identifier(
                                Identifier {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 8",
                                            end: "line: 1, column: 9",
                                            source: "t",
                                        },
                                    },
                                    name: "t",
                                },
                            ),
                        },
                    ),
                    Bad(
                        BadStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 9",
                                    end: "line: 1, column: 12",
                                    source: "“",
                                },
                            },
                            text: "“",
                        },
                    ),
                    Expr(
                        ExprStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 12",
                                    end: "line: 1, column: 13",
                                    source: "r",
                                },
                            },
                            expression: Identifier(
                                Identifier {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 12",
                                            end: "line: 1, column: 13",
                                            source: "r",
                                        },
                                    },
                                    name: "r",
                                },
                            ),
                        },
                    ),
                    Bad(
                        BadStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 13",
                                    end: "line: 1, column: 16",
                                    source: "”",
                                },
                            },
                            text: "”",
                        },
                    ),
                    Expr(
                        ExprStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 16",
                                    end: "line: 1, column: 17",
                                    source: "i",
                                },
                            },
                            expression: Identifier(
                                Identifier {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 16",
                                            end: "line: 1, column: 17",
                                            source: "i",
                                        },
                                    },
                                    name: "i",
                                },
                            ),
                        },
                    ),
                    Bad(
                        BadStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 17",
                                    end: "line: 1, column: 20",
                                    source: "“",
                                },
                            },
                            text: "“",
                        },
                    ),
                    Expr(
                        ExprStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 20",
                                    end: "line: 1, column: 21",
                                    source: "n",
                                },
                            },
                            expression: Identifier(
                                Identifier {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 20",
                                            end: "line: 1, column: 21",
                                            source: "n",
                                        },
                                    },
                                    name: "n",
                                },
                            ),
                        },
                    ),
                    Bad(
                        BadStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 21",
                                    end: "line: 1, column: 24",
                                    source: "”",
                                },
                            },
                            text: "”",
                        },
                    ),
                    Expr(
                        ExprStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 24",
                                    end: "line: 1, column: 25",
                                    source: "g",
                                },
                            },
                            expression: Identifier(
                                Identifier {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 24",
                                            end: "line: 1, column: 25",
                                            source: "g",
                                        },
                                    },
                                    name: "g",
                                },
                            ),
                        },
                    ),
                ],
                eof: [],
            }
        "#]],
    );
}

#[test]
fn package_clause() {
    test_file(
        r#"package foo"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 12",
                        source: "package foo",
                    },
                },
                name: "",
                metadata: "parser-type=rust",
                package: Some(
                    PackageClause {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 12",
                                source: "package foo",
                            },
                        },
                        name: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 9",
                                    end: "line: 1, column: 12",
                                    source: "foo",
                                },
                            },
                            name: "foo",
                        },
                    },
                ),
                imports: [],
                body: [],
                eof: [],
            }
        "#]],
    );
}

#[test]
fn string_interpolation_trailing_dollar() {
    test_file(
        r#""a + b = ${a + b}$""#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 20",
                        source: "\"a + b = ${a + b}$\"",
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
                                    end: "line: 1, column: 20",
                                    source: "\"a + b = ${a + b}$\"",
                                },
                            },
                            expression: StringExpr(
                                StringExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 1",
                                            end: "line: 1, column: 20",
                                            source: "\"a + b = ${a + b}$\"",
                                        },
                                    },
                                    parts: [
                                        Text(
                                            TextPart {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 2",
                                                        end: "line: 1, column: 10",
                                                        source: "a + b = ",
                                                    },
                                                },
                                                value: "a + b = ",
                                            },
                                        ),
                                        Interpolated(
                                            InterpolatedPart {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 10",
                                                        end: "line: 1, column: 18",
                                                        source: "${a + b}",
                                                    },
                                                },
                                                expression: Binary(
                                                    BinaryExpr {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 1, column: 12",
                                                                end: "line: 1, column: 17",
                                                                source: "a + b",
                                                            },
                                                        },
                                                        operator: AdditionOperator,
                                                        left: Identifier(
                                                            Identifier {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 1, column: 12",
                                                                        end: "line: 1, column: 13",
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
                                                                        start: "line: 1, column: 16",
                                                                        end: "line: 1, column: 17",
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
                                        Text(
                                            TextPart {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 18",
                                                        end: "line: 1, column: 19",
                                                        source: "$",
                                                    },
                                                },
                                                value: "$",
                                            },
                                        ),
                                    ],
                                },
                            ),
                        },
                    ),
                ],
                eof: [],
            }
        "#]],
    );
}

#[test]
fn import() {
    test_file(
        r#"import "path/foo""#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 18",
                        source: "import \"path/foo\"",
                    },
                },
                name: "",
                metadata: "parser-type=rust",
                package: None,
                imports: [
                    ImportDeclaration {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 18",
                                source: "import \"path/foo\"",
                            },
                        },
                        alias: None,
                        path: StringLit {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 8",
                                    end: "line: 1, column: 18",
                                    source: "\"path/foo\"",
                                },
                            },
                            value: "path/foo",
                        },
                    },
                ],
                body: [],
                eof: [],
            }
        "#]],
    );
}

#[test]
fn import_as() {
    test_file(
        r#"import bar "path/foo""#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 22",
                        source: "import bar \"path/foo\"",
                    },
                },
                name: "",
                metadata: "parser-type=rust",
                package: None,
                imports: [
                    ImportDeclaration {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 22",
                                source: "import bar \"path/foo\"",
                            },
                        },
                        alias: Some(
                            Identifier {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 8",
                                        end: "line: 1, column: 11",
                                        source: "bar",
                                    },
                                },
                                name: "bar",
                            },
                        ),
                        path: StringLit {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 12",
                                    end: "line: 1, column: 22",
                                    source: "\"path/foo\"",
                                },
                            },
                            value: "path/foo",
                        },
                    },
                ],
                body: [],
                eof: [],
            }
        "#]],
    );
}

#[test]
fn imports() {
    test_file(
        r#"import "path/foo"
import "path/bar""#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 2, column: 18",
                        source: "import \"path/foo\"\nimport \"path/bar\"",
                    },
                },
                name: "",
                metadata: "parser-type=rust",
                package: None,
                imports: [
                    ImportDeclaration {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 18",
                                source: "import \"path/foo\"",
                            },
                        },
                        alias: None,
                        path: StringLit {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 8",
                                    end: "line: 1, column: 18",
                                    source: "\"path/foo\"",
                                },
                            },
                            value: "path/foo",
                        },
                    },
                    ImportDeclaration {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 2, column: 1",
                                end: "line: 2, column: 18",
                                source: "import \"path/bar\"",
                            },
                        },
                        alias: None,
                        path: StringLit {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 2, column: 8",
                                    end: "line: 2, column: 18",
                                    source: "\"path/bar\"",
                                },
                            },
                            value: "path/bar",
                        },
                    },
                ],
                body: [],
                eof: [],
            }
        "#]],
    );
}

#[test]
fn package_and_imports() {
    test_file(
        r#"
package baz

import "path/foo"
import "path/bar""#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 5, column: 18",
                        source: "\npackage baz\n\nimport \"path/foo\"\nimport \"path/bar\"",
                    },
                },
                name: "",
                metadata: "parser-type=rust",
                package: Some(
                    PackageClause {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 2, column: 1",
                                end: "line: 2, column: 12",
                                source: "package baz",
                            },
                        },
                        name: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 2, column: 9",
                                    end: "line: 2, column: 12",
                                    source: "baz",
                                },
                            },
                            name: "baz",
                        },
                    },
                ),
                imports: [
                    ImportDeclaration {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 4, column: 1",
                                end: "line: 4, column: 18",
                                source: "import \"path/foo\"",
                            },
                        },
                        alias: None,
                        path: StringLit {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 4, column: 8",
                                    end: "line: 4, column: 18",
                                    source: "\"path/foo\"",
                                },
                            },
                            value: "path/foo",
                        },
                    },
                    ImportDeclaration {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 5, column: 1",
                                end: "line: 5, column: 18",
                                source: "import \"path/bar\"",
                            },
                        },
                        alias: None,
                        path: StringLit {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 5, column: 8",
                                    end: "line: 5, column: 18",
                                    source: "\"path/bar\"",
                                },
                            },
                            value: "path/bar",
                        },
                    },
                ],
                body: [],
                eof: [],
            }
        "#]],
    );
}

#[test]
fn package_and_imports_and_body() {
    test_file(
        r#"
package baz

import "path/foo"
import "path/bar"

1 + 1"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 7, column: 6",
                        source: "\npackage baz\n\nimport \"path/foo\"\nimport \"path/bar\"\n\n1 + 1",
                    },
                },
                name: "",
                metadata: "parser-type=rust",
                package: Some(
                    PackageClause {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 2, column: 1",
                                end: "line: 2, column: 12",
                                source: "package baz",
                            },
                        },
                        name: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 2, column: 9",
                                    end: "line: 2, column: 12",
                                    source: "baz",
                                },
                            },
                            name: "baz",
                        },
                    },
                ),
                imports: [
                    ImportDeclaration {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 4, column: 1",
                                end: "line: 4, column: 18",
                                source: "import \"path/foo\"",
                            },
                        },
                        alias: None,
                        path: StringLit {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 4, column: 8",
                                    end: "line: 4, column: 18",
                                    source: "\"path/foo\"",
                                },
                            },
                            value: "path/foo",
                        },
                    },
                    ImportDeclaration {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 5, column: 1",
                                end: "line: 5, column: 18",
                                source: "import \"path/bar\"",
                            },
                        },
                        alias: None,
                        path: StringLit {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 5, column: 8",
                                    end: "line: 5, column: 18",
                                    source: "\"path/bar\"",
                                },
                            },
                            value: "path/bar",
                        },
                    },
                ],
                body: [
                    Expr(
                        ExprStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 7, column: 1",
                                    end: "line: 7, column: 6",
                                    source: "1 + 1",
                                },
                            },
                            expression: Binary(
                                BinaryExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 7, column: 1",
                                            end: "line: 7, column: 6",
                                            source: "1 + 1",
                                        },
                                    },
                                    operator: AdditionOperator,
                                    left: Integer(
                                        IntegerLit {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 7, column: 1",
                                                    end: "line: 7, column: 2",
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
                                                    start: "line: 7, column: 5",
                                                    end: "line: 7, column: 6",
                                                    source: "1",
                                                },
                                            },
                                            value: 1,
                                        },
                                    ),
                                },
                            ),
                        },
                    ),
                ],
                eof: [],
            }
        "#]],
    );
}

#[test]
fn optional_query_metadata() {
    test_file(
        r#"option task = {
				name: "foo",
				every: 1h,
				delay: 10m,
				cron: "0 2 * * *",
				retry: 5,
			  }"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 7, column: 7",
                        source: "option task = {\n\t\t\t\tname: \"foo\",\n\t\t\t\tevery: 1h,\n\t\t\t\tdelay: 10m,\n\t\t\t\tcron: \"0 2 * * *\",\n\t\t\t\tretry: 5,\n\t\t\t  }",
                    },
                },
                name: "",
                metadata: "parser-type=rust",
                package: None,
                imports: [],
                body: [
                    Option(
                        OptionStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 1",
                                    end: "line: 7, column: 7",
                                    source: "option task = {\n\t\t\t\tname: \"foo\",\n\t\t\t\tevery: 1h,\n\t\t\t\tdelay: 10m,\n\t\t\t\tcron: \"0 2 * * *\",\n\t\t\t\tretry: 5,\n\t\t\t  }",
                                },
                            },
                            assignment: Variable(
                                VariableAssgn {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 8",
                                            end: "line: 7, column: 7",
                                            source: "task = {\n\t\t\t\tname: \"foo\",\n\t\t\t\tevery: 1h,\n\t\t\t\tdelay: 10m,\n\t\t\t\tcron: \"0 2 * * *\",\n\t\t\t\tretry: 5,\n\t\t\t  }",
                                        },
                                    },
                                    id: Identifier {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 8",
                                                end: "line: 1, column: 12",
                                                source: "task",
                                            },
                                        },
                                        name: "task",
                                    },
                                    init: Object(
                                        ObjectExpr {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 15",
                                                    end: "line: 7, column: 7",
                                                    source: "{\n\t\t\t\tname: \"foo\",\n\t\t\t\tevery: 1h,\n\t\t\t\tdelay: 10m,\n\t\t\t\tcron: \"0 2 * * *\",\n\t\t\t\tretry: 5,\n\t\t\t  }",
                                                },
                                            },
                                            lbrace: [],
                                            with: None,
                                            properties: [
                                                Property {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 2, column: 5",
                                                            end: "line: 2, column: 16",
                                                            source: "name: \"foo\"",
                                                        },
                                                    },
                                                    key: Identifier(
                                                        Identifier {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 2, column: 5",
                                                                    end: "line: 2, column: 9",
                                                                    source: "name",
                                                                },
                                                            },
                                                            name: "name",
                                                        },
                                                    ),
                                                    separator: [],
                                                    value: Some(
                                                        StringLit(
                                                            StringLit {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 2, column: 11",
                                                                        end: "line: 2, column: 16",
                                                                        source: "\"foo\"",
                                                                    },
                                                                },
                                                                value: "foo",
                                                            },
                                                        ),
                                                    ),
                                                    comma: [],
                                                },
                                                Property {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 3, column: 5",
                                                            end: "line: 3, column: 14",
                                                            source: "every: 1h",
                                                        },
                                                    },
                                                    key: Identifier(
                                                        Identifier {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 3, column: 5",
                                                                    end: "line: 3, column: 10",
                                                                    source: "every",
                                                                },
                                                            },
                                                            name: "every",
                                                        },
                                                    ),
                                                    separator: [],
                                                    value: Some(
                                                        Duration(
                                                            DurationLit {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 3, column: 12",
                                                                        end: "line: 3, column: 14",
                                                                        source: "1h",
                                                                    },
                                                                },
                                                                values: [
                                                                    Duration {
                                                                        magnitude: 1,
                                                                        unit: "h",
                                                                    },
                                                                ],
                                                            },
                                                        ),
                                                    ),
                                                    comma: [],
                                                },
                                                Property {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 4, column: 5",
                                                            end: "line: 4, column: 15",
                                                            source: "delay: 10m",
                                                        },
                                                    },
                                                    key: Identifier(
                                                        Identifier {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 4, column: 5",
                                                                    end: "line: 4, column: 10",
                                                                    source: "delay",
                                                                },
                                                            },
                                                            name: "delay",
                                                        },
                                                    ),
                                                    separator: [],
                                                    value: Some(
                                                        Duration(
                                                            DurationLit {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 4, column: 12",
                                                                        end: "line: 4, column: 15",
                                                                        source: "10m",
                                                                    },
                                                                },
                                                                values: [
                                                                    Duration {
                                                                        magnitude: 10,
                                                                        unit: "m",
                                                                    },
                                                                ],
                                                            },
                                                        ),
                                                    ),
                                                    comma: [],
                                                },
                                                Property {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 5, column: 5",
                                                            end: "line: 5, column: 22",
                                                            source: "cron: \"0 2 * * *\"",
                                                        },
                                                    },
                                                    key: Identifier(
                                                        Identifier {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 5, column: 5",
                                                                    end: "line: 5, column: 9",
                                                                    source: "cron",
                                                                },
                                                            },
                                                            name: "cron",
                                                        },
                                                    ),
                                                    separator: [],
                                                    value: Some(
                                                        StringLit(
                                                            StringLit {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 5, column: 11",
                                                                        end: "line: 5, column: 22",
                                                                        source: "\"0 2 * * *\"",
                                                                    },
                                                                },
                                                                value: "0 2 * * *",
                                                            },
                                                        ),
                                                    ),
                                                    comma: [],
                                                },
                                                Property {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 6, column: 5",
                                                            end: "line: 6, column: 13",
                                                            source: "retry: 5",
                                                        },
                                                    },
                                                    key: Identifier(
                                                        Identifier {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 6, column: 5",
                                                                    end: "line: 6, column: 10",
                                                                    source: "retry",
                                                                },
                                                            },
                                                            name: "retry",
                                                        },
                                                    ),
                                                    separator: [],
                                                    value: Some(
                                                        Integer(
                                                            IntegerLit {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 6, column: 12",
                                                                        end: "line: 6, column: 13",
                                                                        source: "5",
                                                                    },
                                                                },
                                                                value: 5,
                                                            },
                                                        ),
                                                    ),
                                                    comma: [],
                                                },
                                            ],
                                            rbrace: [],
                                        },
                                    ),
                                },
                            ),
                        },
                    ),
                ],
                eof: [],
            }
        "#]],
    );
}

#[test]
fn qualified_option() {
    test_file(
        r#"option alert.state = "Warning""#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 31",
                        source: "option alert.state = \"Warning\"",
                    },
                },
                name: "",
                metadata: "parser-type=rust",
                package: None,
                imports: [],
                body: [
                    Option(
                        OptionStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 1",
                                    end: "line: 1, column: 31",
                                    source: "option alert.state = \"Warning\"",
                                },
                            },
                            assignment: Member(
                                MemberAssgn {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 8",
                                            end: "line: 1, column: 31",
                                            source: "alert.state = \"Warning\"",
                                        },
                                    },
                                    member: MemberExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 8",
                                                end: "line: 1, column: 19",
                                                source: "alert.state",
                                            },
                                        },
                                        object: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 8",
                                                        end: "line: 1, column: 13",
                                                        source: "alert",
                                                    },
                                                },
                                                name: "alert",
                                            },
                                        ),
                                        lbrack: [],
                                        property: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 14",
                                                        end: "line: 1, column: 19",
                                                        source: "state",
                                                    },
                                                },
                                                name: "state",
                                            },
                                        ),
                                        rbrack: [],
                                    },
                                    init: StringLit(
                                        StringLit {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 22",
                                                    end: "line: 1, column: 31",
                                                    source: "\"Warning\"",
                                                },
                                            },
                                            value: "Warning",
                                        },
                                    ),
                                },
                            ),
                        },
                    ),
                ],
                eof: [],
            }
        "#]],
    );
}

#[test]
fn builtin() {
    test_file(
        r#"builtin from : int"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 19",
                        source: "builtin from : int",
                    },
                },
                name: "",
                metadata: "parser-type=rust",
                package: None,
                imports: [],
                body: [
                    Builtin(
                        BuiltinStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 1",
                                    end: "line: 1, column: 19",
                                    source: "builtin from : int",
                                },
                            },
                            colon: [],
                            id: Identifier {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 9",
                                        end: "line: 1, column: 13",
                                        source: "from",
                                    },
                                },
                                name: "from",
                            },
                            ty: TypeExpression {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 16",
                                        end: "line: 1, column: 19",
                                        source: "int",
                                    },
                                },
                                monotype: Basic(
                                    NamedType {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 16",
                                                end: "line: 1, column: 19",
                                                source: "int",
                                            },
                                        },
                                        name: Identifier {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 16",
                                                    end: "line: 1, column: 19",
                                                    source: "int",
                                                },
                                            },
                                            name: "int",
                                        },
                                    },
                                ),
                                constraints: [],
                            },
                        },
                    ),
                ],
                eof: [],
            }
        "#]],
    );
}

#[test]
fn comment() {
    test_file(
        r#"// Comment
			from()"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 2, column: 10",
                        source: "// Comment\n\t\t\tfrom()",
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
                                    start: "line: 2, column: 4",
                                    end: "line: 2, column: 10",
                                    source: "from()",
                                },
                            },
                            expression: Call(
                                CallExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 2, column: 4",
                                            end: "line: 2, column: 10",
                                            source: "from()",
                                        },
                                    },
                                    callee: Identifier(
                                        Identifier {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 2, column: 4",
                                                    end: "line: 2, column: 8",
                                                    source: "from",
                                                },
                                                comments: [
                                                    Comment {
                                                        text: "// Comment\n",
                                                    },
                                                ],
                                            },
                                            name: "from",
                                        },
                                    ),
                                    lparen: [],
                                    arguments: [],
                                    rparen: [],
                                },
                            ),
                        },
                    ),
                ],
                eof: [],
            }
        "#]],
    );
}
#[test]
fn comment_builtin() {
    test_file(
        r#"// Comment
builtin foo
// colon comment
: int"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 4, column: 6",
                        source: "// Comment\nbuiltin foo\n// colon comment\n: int",
                    },
                },
                name: "",
                metadata: "parser-type=rust",
                package: None,
                imports: [],
                body: [
                    Builtin(
                        BuiltinStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 2, column: 1",
                                    end: "line: 4, column: 6",
                                    source: "builtin foo\n// colon comment\n: int",
                                },
                                comments: [
                                    Comment {
                                        text: "// Comment\n",
                                    },
                                ],
                            },
                            colon: [
                                Comment {
                                    text: "// colon comment\n",
                                },
                            ],
                            id: Identifier {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 2, column: 9",
                                        end: "line: 2, column: 12",
                                        source: "foo",
                                    },
                                },
                                name: "foo",
                            },
                            ty: TypeExpression {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 4, column: 3",
                                        end: "line: 4, column: 6",
                                        source: "int",
                                    },
                                },
                                monotype: Basic(
                                    NamedType {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 4, column: 3",
                                                end: "line: 4, column: 6",
                                                source: "int",
                                            },
                                        },
                                        name: Identifier {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 4, column: 3",
                                                    end: "line: 4, column: 6",
                                                    source: "int",
                                                },
                                            },
                                            name: "int",
                                        },
                                    },
                                ),
                                constraints: [],
                            },
                        },
                    ),
                ],
                eof: [],
            }
        "#]],
    );
}

#[test]
fn comment_function_body() {
    test_file(
        r#"fn = (tables=<-) =>
// comment
(tables)"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 3, column: 9",
                        source: "fn = (tables=<-) =>\n// comment\n(tables)",
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
                                    end: "line: 3, column: 9",
                                    source: "fn = (tables=<-) =>\n// comment\n(tables)",
                                },
                            },
                            id: Identifier {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 3",
                                        source: "fn",
                                    },
                                },
                                name: "fn",
                            },
                            init: Function(
                                FunctionExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 6",
                                            end: "line: 3, column: 9",
                                            source: "(tables=<-) =>\n// comment\n(tables)",
                                        },
                                    },
                                    lparen: [],
                                    params: [
                                        Property {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 7",
                                                    end: "line: 1, column: 16",
                                                    source: "tables=<-",
                                                },
                                            },
                                            key: Identifier(
                                                Identifier {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 7",
                                                            end: "line: 1, column: 13",
                                                            source: "tables",
                                                        },
                                                    },
                                                    name: "tables",
                                                },
                                            ),
                                            separator: [],
                                            value: Some(
                                                PipeLit(
                                                    PipeLit {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 1, column: 14",
                                                                end: "line: 1, column: 16",
                                                                source: "<-",
                                                            },
                                                        },
                                                    },
                                                ),
                                            ),
                                            comma: [],
                                        },
                                    ],
                                    rparen: [],
                                    arrow: [],
                                    body: Expr(
                                        Paren(
                                            ParenExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 3, column: 1",
                                                        end: "line: 3, column: 9",
                                                        source: "(tables)",
                                                    },
                                                },
                                                lparen: [
                                                    Comment {
                                                        text: "// comment\n",
                                                    },
                                                ],
                                                expression: Identifier(
                                                    Identifier {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 3, column: 2",
                                                                end: "line: 3, column: 8",
                                                                source: "tables",
                                                            },
                                                        },
                                                        name: "tables",
                                                    },
                                                ),
                                                rparen: [],
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
        "#]],
    );
}

#[test]
fn identifier_with_number() {
    test_file(
        r#"tan2()"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 7",
                        source: "tan2()",
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
                                    end: "line: 1, column: 7",
                                    source: "tan2()",
                                },
                            },
                            expression: Call(
                                CallExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 1",
                                            end: "line: 1, column: 7",
                                            source: "tan2()",
                                        },
                                    },
                                    callee: Identifier(
                                        Identifier {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 1",
                                                    end: "line: 1, column: 5",
                                                    source: "tan2",
                                                },
                                            },
                                            name: "tan2",
                                        },
                                    ),
                                    lparen: [],
                                    arguments: [],
                                    rparen: [],
                                },
                            ),
                        },
                    ),
                ],
                eof: [],
            }
        "#]],
    );
}

#[test]
fn regex_match_operators() {
    test_file(
        r#""a" =~ /.*/ and "b" !~ /c$/"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 28",
                        source: "\"a\" =~ /.*/ and \"b\" !~ /c$/",
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
                                    end: "line: 1, column: 28",
                                    source: "\"a\" =~ /.*/ and \"b\" !~ /c$/",
                                },
                            },
                            expression: Logical(
                                LogicalExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 1",
                                            end: "line: 1, column: 28",
                                            source: "\"a\" =~ /.*/ and \"b\" !~ /c$/",
                                        },
                                    },
                                    operator: AndOperator,
                                    left: Binary(
                                        BinaryExpr {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 1",
                                                    end: "line: 1, column: 12",
                                                    source: "\"a\" =~ /.*/",
                                                },
                                            },
                                            operator: RegexpMatchOperator,
                                            left: StringLit(
                                                StringLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 1",
                                                            end: "line: 1, column: 4",
                                                            source: "\"a\"",
                                                        },
                                                    },
                                                    value: "a",
                                                },
                                            ),
                                            right: Regexp(
                                                RegexpLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 8",
                                                            end: "line: 1, column: 12",
                                                            source: "/.*/",
                                                        },
                                                    },
                                                    value: ".*",
                                                },
                                            ),
                                        },
                                    ),
                                    right: Binary(
                                        BinaryExpr {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 17",
                                                    end: "line: 1, column: 28",
                                                    source: "\"b\" !~ /c$/",
                                                },
                                            },
                                            operator: NotRegexpMatchOperator,
                                            left: StringLit(
                                                StringLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 17",
                                                            end: "line: 1, column: 20",
                                                            source: "\"b\"",
                                                        },
                                                    },
                                                    value: "b",
                                                },
                                            ),
                                            right: Regexp(
                                                RegexpLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 24",
                                                            end: "line: 1, column: 28",
                                                            source: "/c$/",
                                                        },
                                                    },
                                                    value: "c$",
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
        "#]],
    );
}

#[test]
fn declare_variable_as_an_int() {
    test_file(
        r#"howdy = 1"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 10",
                        source: "howdy = 1",
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
                                    end: "line: 1, column: 10",
                                    source: "howdy = 1",
                                },
                            },
                            id: Identifier {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 6",
                                        source: "howdy",
                                    },
                                },
                                name: "howdy",
                            },
                            init: Integer(
                                IntegerLit {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 9",
                                            end: "line: 1, column: 10",
                                            source: "1",
                                        },
                                    },
                                    value: 1,
                                },
                            ),
                        },
                    ),
                ],
                eof: [],
            }
        "#]],
    );
}

#[test]
fn declare_variable_as_a_float() {
    test_file(
        r#"howdy = 1.1"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 12",
                        source: "howdy = 1.1",
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
                                    end: "line: 1, column: 12",
                                    source: "howdy = 1.1",
                                },
                            },
                            id: Identifier {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 6",
                                        source: "howdy",
                                    },
                                },
                                name: "howdy",
                            },
                            init: Float(
                                FloatLit {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 9",
                                            end: "line: 1, column: 12",
                                            source: "1.1",
                                        },
                                    },
                                    value: NotNan(
                                        1.1,
                                    ),
                                },
                            ),
                        },
                    ),
                ],
                eof: [],
            }
        "#]],
    );
}

#[test]
fn declare_variable_as_an_array() {
    test_file(
        r#"howdy = [1, 2, 3, 4]"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 21",
                        source: "howdy = [1, 2, 3, 4]",
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
                                    end: "line: 1, column: 21",
                                    source: "howdy = [1, 2, 3, 4]",
                                },
                            },
                            id: Identifier {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 6",
                                        source: "howdy",
                                    },
                                },
                                name: "howdy",
                            },
                            init: Array(
                                ArrayExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 9",
                                            end: "line: 1, column: 21",
                                            source: "[1, 2, 3, 4]",
                                        },
                                    },
                                    lbrack: [],
                                    elements: [
                                        ArrayItem {
                                            expression: Integer(
                                                IntegerLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 10",
                                                            end: "line: 1, column: 11",
                                                            source: "1",
                                                        },
                                                    },
                                                    value: 1,
                                                },
                                            ),
                                            comma: [],
                                        },
                                        ArrayItem {
                                            expression: Integer(
                                                IntegerLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 13",
                                                            end: "line: 1, column: 14",
                                                            source: "2",
                                                        },
                                                    },
                                                    value: 2,
                                                },
                                            ),
                                            comma: [],
                                        },
                                        ArrayItem {
                                            expression: Integer(
                                                IntegerLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 16",
                                                            end: "line: 1, column: 17",
                                                            source: "3",
                                                        },
                                                    },
                                                    value: 3,
                                                },
                                            ),
                                            comma: [],
                                        },
                                        ArrayItem {
                                            expression: Integer(
                                                IntegerLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 19",
                                                            end: "line: 1, column: 20",
                                                            source: "4",
                                                        },
                                                    },
                                                    value: 4,
                                                },
                                            ),
                                            comma: [],
                                        },
                                    ],
                                    rbrack: [],
                                },
                            ),
                        },
                    ),
                ],
                eof: [],
            }
        "#]],
    );
}

#[test]
fn declare_variable_as_an_empty_array() {
    test_file(
        r#"howdy = []"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 11",
                        source: "howdy = []",
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
                                    end: "line: 1, column: 11",
                                    source: "howdy = []",
                                },
                            },
                            id: Identifier {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 6",
                                        source: "howdy",
                                    },
                                },
                                name: "howdy",
                            },
                            init: Array(
                                ArrayExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 9",
                                            end: "line: 1, column: 11",
                                            source: "[]",
                                        },
                                    },
                                    lbrack: [],
                                    elements: [],
                                    rbrack: [],
                                },
                            ),
                        },
                    ),
                ],
                eof: [],
            }
        "#]],
    );
}

#[test]
fn parse_empty_dict() {
    test_file(
        r#"[:]"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 4",
                        source: "[:]",
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
                                    source: "[:]",
                                },
                            },
                            expression: Dict(
                                DictExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 1",
                                            end: "line: 1, column: 4",
                                            source: "[:]",
                                        },
                                    },
                                    lbrack: [],
                                    elements: [],
                                    rbrack: [],
                                },
                            ),
                        },
                    ),
                ],
                eof: [],
            }
        "#]],
    );
}

#[test]
fn parse_single_element_dict() {
    test_file(
        r#"["a": 0]"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 9",
                        source: "[\"a\": 0]",
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
                                    end: "line: 1, column: 9",
                                    source: "[\"a\": 0]",
                                },
                            },
                            expression: Dict(
                                DictExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 1",
                                            end: "line: 1, column: 9",
                                            source: "[\"a\": 0]",
                                        },
                                    },
                                    lbrack: [],
                                    elements: [
                                        DictItem {
                                            key: StringLit(
                                                StringLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 2",
                                                            end: "line: 1, column: 5",
                                                            source: "\"a\"",
                                                        },
                                                    },
                                                    value: "a",
                                                },
                                            ),
                                            val: Integer(
                                                IntegerLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 7",
                                                            end: "line: 1, column: 8",
                                                            source: "0",
                                                        },
                                                    },
                                                    value: 0,
                                                },
                                            ),
                                            comma: [],
                                        },
                                    ],
                                    rbrack: [],
                                },
                            ),
                        },
                    ),
                ],
                eof: [],
            }
        "#]],
    );
}

#[test]
fn parse_multi_element_dict() {
    test_file(
        r#"["a": 0, "b": 1]"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 17",
                        source: "[\"a\": 0, \"b\": 1]",
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
                                    end: "line: 1, column: 17",
                                    source: "[\"a\": 0, \"b\": 1]",
                                },
                            },
                            expression: Dict(
                                DictExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 1",
                                            end: "line: 1, column: 17",
                                            source: "[\"a\": 0, \"b\": 1]",
                                        },
                                    },
                                    lbrack: [],
                                    elements: [
                                        DictItem {
                                            key: StringLit(
                                                StringLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 2",
                                                            end: "line: 1, column: 5",
                                                            source: "\"a\"",
                                                        },
                                                    },
                                                    value: "a",
                                                },
                                            ),
                                            val: Integer(
                                                IntegerLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 7",
                                                            end: "line: 1, column: 8",
                                                            source: "0",
                                                        },
                                                    },
                                                    value: 0,
                                                },
                                            ),
                                            comma: [],
                                        },
                                        DictItem {
                                            key: StringLit(
                                                StringLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 10",
                                                            end: "line: 1, column: 13",
                                                            source: "\"b\"",
                                                        },
                                                    },
                                                    value: "b",
                                                },
                                            ),
                                            val: Integer(
                                                IntegerLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 15",
                                                            end: "line: 1, column: 16",
                                                            source: "1",
                                                        },
                                                    },
                                                    value: 1,
                                                },
                                            ),
                                            comma: [],
                                        },
                                    ],
                                    rbrack: [],
                                },
                            ),
                        },
                    ),
                ],
                eof: [],
            }
        "#]],
    );
}

#[test]
fn parse_dict_trailing_comma0() {
    test_file(
        r#"["a": 0, ]"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 11",
                        source: "[\"a\": 0, ]",
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
                                    source: "[\"a\": 0, ]",
                                },
                            },
                            expression: Dict(
                                DictExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 1",
                                            end: "line: 1, column: 11",
                                            source: "[\"a\": 0, ]",
                                        },
                                    },
                                    lbrack: [],
                                    elements: [
                                        DictItem {
                                            key: StringLit(
                                                StringLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 2",
                                                            end: "line: 1, column: 5",
                                                            source: "\"a\"",
                                                        },
                                                    },
                                                    value: "a",
                                                },
                                            ),
                                            val: Integer(
                                                IntegerLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 7",
                                                            end: "line: 1, column: 8",
                                                            source: "0",
                                                        },
                                                    },
                                                    value: 0,
                                                },
                                            ),
                                            comma: [],
                                        },
                                    ],
                                    rbrack: [],
                                },
                            ),
                        },
                    ),
                ],
                eof: [],
            }
        "#]],
    );
}

#[test]
fn parse_dict_trailing_comma1() {
    test_file(
        r#"["a": 0, "b": 1, ]"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 19",
                        source: "[\"a\": 0, \"b\": 1, ]",
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
                                    end: "line: 1, column: 19",
                                    source: "[\"a\": 0, \"b\": 1, ]",
                                },
                            },
                            expression: Dict(
                                DictExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 1",
                                            end: "line: 1, column: 19",
                                            source: "[\"a\": 0, \"b\": 1, ]",
                                        },
                                    },
                                    lbrack: [],
                                    elements: [
                                        DictItem {
                                            key: StringLit(
                                                StringLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 2",
                                                            end: "line: 1, column: 5",
                                                            source: "\"a\"",
                                                        },
                                                    },
                                                    value: "a",
                                                },
                                            ),
                                            val: Integer(
                                                IntegerLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 7",
                                                            end: "line: 1, column: 8",
                                                            source: "0",
                                                        },
                                                    },
                                                    value: 0,
                                                },
                                            ),
                                            comma: [],
                                        },
                                        DictItem {
                                            key: StringLit(
                                                StringLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 10",
                                                            end: "line: 1, column: 13",
                                                            source: "\"b\"",
                                                        },
                                                    },
                                                    value: "b",
                                                },
                                            ),
                                            val: Integer(
                                                IntegerLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 15",
                                                            end: "line: 1, column: 16",
                                                            source: "1",
                                                        },
                                                    },
                                                    value: 1,
                                                },
                                            ),
                                            comma: [],
                                        },
                                    ],
                                    rbrack: [],
                                },
                            ),
                        },
                    ),
                ],
                eof: [],
            }
        "#]],
    );
}

#[test]
fn parse_dict_arbitrary_keys() {
    test_file(
        r#"[1-1: 0, 1+1: 1]"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 17",
                        source: "[1-1: 0, 1+1: 1]",
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
                                    end: "line: 1, column: 17",
                                    source: "[1-1: 0, 1+1: 1]",
                                },
                            },
                            expression: Dict(
                                DictExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 1",
                                            end: "line: 1, column: 17",
                                            source: "[1-1: 0, 1+1: 1]",
                                        },
                                    },
                                    lbrack: [],
                                    elements: [
                                        DictItem {
                                            key: Binary(
                                                BinaryExpr {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 2",
                                                            end: "line: 1, column: 5",
                                                            source: "1-1",
                                                        },
                                                    },
                                                    operator: SubtractionOperator,
                                                    left: Integer(
                                                        IntegerLit {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 1, column: 2",
                                                                    end: "line: 1, column: 3",
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
                                                                    start: "line: 1, column: 4",
                                                                    end: "line: 1, column: 5",
                                                                    source: "1",
                                                                },
                                                            },
                                                            value: 1,
                                                        },
                                                    ),
                                                },
                                            ),
                                            val: Integer(
                                                IntegerLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 7",
                                                            end: "line: 1, column: 8",
                                                            source: "0",
                                                        },
                                                    },
                                                    value: 0,
                                                },
                                            ),
                                            comma: [],
                                        },
                                        DictItem {
                                            key: Binary(
                                                BinaryExpr {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 10",
                                                            end: "line: 1, column: 13",
                                                            source: "1+1",
                                                        },
                                                    },
                                                    operator: AdditionOperator,
                                                    left: Integer(
                                                        IntegerLit {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 1, column: 10",
                                                                    end: "line: 1, column: 11",
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
                                                                    start: "line: 1, column: 12",
                                                                    end: "line: 1, column: 13",
                                                                    source: "1",
                                                                },
                                                            },
                                                            value: 1,
                                                        },
                                                    ),
                                                },
                                            ),
                                            val: Integer(
                                                IntegerLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 15",
                                                            end: "line: 1, column: 16",
                                                            source: "1",
                                                        },
                                                    },
                                                    value: 1,
                                                },
                                            ),
                                            comma: [],
                                        },
                                    ],
                                    rbrack: [],
                                },
                            ),
                        },
                    ),
                ],
                eof: [],
            }
        "#]],
    );
}

#[test]
fn use_variable_to_declare_something() {
    test_file(
        r#"howdy = 1
			from()"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 2, column: 10",
                        source: "howdy = 1\n\t\t\tfrom()",
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
                                    end: "line: 1, column: 10",
                                    source: "howdy = 1",
                                },
                            },
                            id: Identifier {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 6",
                                        source: "howdy",
                                    },
                                },
                                name: "howdy",
                            },
                            init: Integer(
                                IntegerLit {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 9",
                                            end: "line: 1, column: 10",
                                            source: "1",
                                        },
                                    },
                                    value: 1,
                                },
                            ),
                        },
                    ),
                    Expr(
                        ExprStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 2, column: 4",
                                    end: "line: 2, column: 10",
                                    source: "from()",
                                },
                            },
                            expression: Call(
                                CallExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 2, column: 4",
                                            end: "line: 2, column: 10",
                                            source: "from()",
                                        },
                                    },
                                    callee: Identifier(
                                        Identifier {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 2, column: 4",
                                                    end: "line: 2, column: 8",
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
                        },
                    ),
                ],
                eof: [],
            }
        "#]],
    );
}

#[test]
fn variable_is_from_statement() {
    test_file(
        r#"howdy = from()
			howdy.count()"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 2, column: 17",
                        source: "howdy = from()\n\t\t\thowdy.count()",
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
                                    end: "line: 1, column: 15",
                                    source: "howdy = from()",
                                },
                            },
                            id: Identifier {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 6",
                                        source: "howdy",
                                    },
                                },
                                name: "howdy",
                            },
                            init: Call(
                                CallExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 9",
                                            end: "line: 1, column: 15",
                                            source: "from()",
                                        },
                                    },
                                    callee: Identifier(
                                        Identifier {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 9",
                                                    end: "line: 1, column: 13",
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
                        },
                    ),
                    Expr(
                        ExprStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 2, column: 4",
                                    end: "line: 2, column: 17",
                                    source: "howdy.count()",
                                },
                            },
                            expression: Call(
                                CallExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 2, column: 4",
                                            end: "line: 2, column: 17",
                                            source: "howdy.count()",
                                        },
                                    },
                                    callee: Member(
                                        MemberExpr {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 2, column: 4",
                                                    end: "line: 2, column: 15",
                                                    source: "howdy.count",
                                                },
                                            },
                                            object: Identifier(
                                                Identifier {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 2, column: 4",
                                                            end: "line: 2, column: 9",
                                                            source: "howdy",
                                                        },
                                                    },
                                                    name: "howdy",
                                                },
                                            ),
                                            lbrack: [],
                                            property: Identifier(
                                                Identifier {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 2, column: 10",
                                                            end: "line: 2, column: 15",
                                                            source: "count",
                                                        },
                                                    },
                                                    name: "count",
                                                },
                                            ),
                                            rbrack: [],
                                        },
                                    ),
                                    lparen: [],
                                    arguments: [],
                                    rparen: [],
                                },
                            ),
                        },
                    ),
                ],
                eof: [],
            }
        "#]],
    );
}

#[test]
fn pipe_expression() {
    test_file(
        r#"from() |> count()"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 18",
                        source: "from() |> count()",
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
                                    end: "line: 1, column: 18",
                                    source: "from() |> count()",
                                },
                            },
                            expression: PipeExpr(
                                PipeExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 1",
                                            end: "line: 1, column: 18",
                                            source: "from() |> count()",
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
                                                source: "count()",
                                            },
                                        },
                                        callee: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 11",
                                                        end: "line: 1, column: 16",
                                                        source: "count",
                                                    },
                                                },
                                                name: "count",
                                            },
                                        ),
                                        lparen: [],
                                        arguments: [],
                                        rparen: [],
                                    },
                                },
                            ),
                        },
                    ),
                ],
                eof: [],
            }
        "#]],
    );
}

#[test]
fn pipe_expression_to_member_expression_function() {
    test_file(
        r#"a |> b.c(d:e)"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 14",
                        source: "a |> b.c(d:e)",
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
                                    source: "a |> b.c(d:e)",
                                },
                            },
                            expression: PipeExpr(
                                PipeExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 1",
                                            end: "line: 1, column: 14",
                                            source: "a |> b.c(d:e)",
                                        },
                                    },
                                    argument: Identifier(
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
                                    call: CallExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 6",
                                                end: "line: 1, column: 14",
                                                source: "b.c(d:e)",
                                            },
                                        },
                                        callee: Member(
                                            MemberExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 6",
                                                        end: "line: 1, column: 9",
                                                        source: "b.c",
                                                    },
                                                },
                                                object: Identifier(
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
                                                lbrack: [],
                                                property: Identifier(
                                                    Identifier {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 1, column: 8",
                                                                end: "line: 1, column: 9",
                                                                source: "c",
                                                            },
                                                        },
                                                        name: "c",
                                                    },
                                                ),
                                                rbrack: [],
                                            },
                                        ),
                                        lparen: [],
                                        arguments: [
                                            Object(
                                                ObjectExpr {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 10",
                                                            end: "line: 1, column: 13",
                                                            source: "d:e",
                                                        },
                                                    },
                                                    lbrace: [],
                                                    with: None,
                                                    properties: [
                                                        Property {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 1, column: 10",
                                                                    end: "line: 1, column: 13",
                                                                    source: "d:e",
                                                                },
                                                            },
                                                            key: Identifier(
                                                                Identifier {
                                                                    base: BaseNode {
                                                                        location: SourceLocation {
                                                                            start: "line: 1, column: 10",
                                                                            end: "line: 1, column: 11",
                                                                            source: "d",
                                                                        },
                                                                    },
                                                                    name: "d",
                                                                },
                                                            ),
                                                            separator: [],
                                                            value: Some(
                                                                Identifier(
                                                                    Identifier {
                                                                        base: BaseNode {
                                                                            location: SourceLocation {
                                                                                start: "line: 1, column: 12",
                                                                                end: "line: 1, column: 13",
                                                                                source: "e",
                                                                            },
                                                                        },
                                                                        name: "e",
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
        "#]],
    );
}

#[test]
fn literal_pipe_expression() {
    test_file(
        r#"5 |> pow2()"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 12",
                        source: "5 |> pow2()",
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
                                    source: "5 |> pow2()",
                                },
                            },
                            expression: PipeExpr(
                                PipeExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 1",
                                            end: "line: 1, column: 12",
                                            source: "5 |> pow2()",
                                        },
                                    },
                                    argument: Integer(
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
                                    call: CallExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 6",
                                                end: "line: 1, column: 12",
                                                source: "pow2()",
                                            },
                                        },
                                        callee: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 6",
                                                        end: "line: 1, column: 10",
                                                        source: "pow2",
                                                    },
                                                },
                                                name: "pow2",
                                            },
                                        ),
                                        lparen: [],
                                        arguments: [],
                                        rparen: [],
                                    },
                                },
                            ),
                        },
                    ),
                ],
                eof: [],
            }
        "#]],
    );
}

#[test]
fn member_expression_pipe_expression() {
    test_file(
        r#"foo.bar |> baz()"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 17",
                        source: "foo.bar |> baz()",
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
                                    end: "line: 1, column: 17",
                                    source: "foo.bar |> baz()",
                                },
                            },
                            expression: PipeExpr(
                                PipeExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 1",
                                            end: "line: 1, column: 17",
                                            source: "foo.bar |> baz()",
                                        },
                                    },
                                    argument: Member(
                                        MemberExpr {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 1",
                                                    end: "line: 1, column: 8",
                                                    source: "foo.bar",
                                                },
                                            },
                                            object: Identifier(
                                                Identifier {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 1",
                                                            end: "line: 1, column: 4",
                                                            source: "foo",
                                                        },
                                                    },
                                                    name: "foo",
                                                },
                                            ),
                                            lbrack: [],
                                            property: Identifier(
                                                Identifier {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 5",
                                                            end: "line: 1, column: 8",
                                                            source: "bar",
                                                        },
                                                    },
                                                    name: "bar",
                                                },
                                            ),
                                            rbrack: [],
                                        },
                                    ),
                                    call: CallExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 12",
                                                end: "line: 1, column: 17",
                                                source: "baz()",
                                            },
                                        },
                                        callee: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 12",
                                                        end: "line: 1, column: 15",
                                                        source: "baz",
                                                    },
                                                },
                                                name: "baz",
                                            },
                                        ),
                                        lparen: [],
                                        arguments: [],
                                        rparen: [],
                                    },
                                },
                            ),
                        },
                    ),
                ],
                eof: [],
            }
        "#]],
    );
}

#[test]
fn multiple_pipe_expressions() {
    test_file(
        r#"from() |> range() |> filter() |> count()"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 41",
                        source: "from() |> range() |> filter() |> count()",
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
                                    end: "line: 1, column: 41",
                                    source: "from() |> range() |> filter() |> count()",
                                },
                            },
                            expression: PipeExpr(
                                PipeExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 1",
                                            end: "line: 1, column: 41",
                                            source: "from() |> range() |> filter() |> count()",
                                        },
                                    },
                                    argument: PipeExpr(
                                        PipeExpr {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 1",
                                                    end: "line: 1, column: 30",
                                                    source: "from() |> range() |> filter()",
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
                                                        end: "line: 1, column: 30",
                                                        source: "filter()",
                                                    },
                                                },
                                                callee: Identifier(
                                                    Identifier {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 1, column: 22",
                                                                end: "line: 1, column: 28",
                                                                source: "filter",
                                                            },
                                                        },
                                                        name: "filter",
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
                                                start: "line: 1, column: 34",
                                                end: "line: 1, column: 41",
                                                source: "count()",
                                            },
                                        },
                                        callee: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 34",
                                                        end: "line: 1, column: 39",
                                                        source: "count",
                                                    },
                                                },
                                                name: "count",
                                            },
                                        ),
                                        lparen: [],
                                        arguments: [],
                                        rparen: [],
                                    },
                                },
                            ),
                        },
                    ),
                ],
                eof: [],
            }
        "#]],
    );
}

#[test]
fn pipe_expression_into_non_call_expression() {
    test_file(
        r#"foo() |> bar"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 13",
                        source: "foo() |> bar",
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
                                    source: "foo() |> bar",
                                },
                            },
                            expression: PipeExpr(
                                PipeExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 1",
                                            end: "line: 1, column: 13",
                                            source: "foo() |> bar",
                                        },
                                    },
                                    argument: Call(
                                        CallExpr {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 1",
                                                    end: "line: 1, column: 6",
                                                    source: "foo()",
                                                },
                                            },
                                            callee: Identifier(
                                                Identifier {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 1",
                                                            end: "line: 1, column: 4",
                                                            source: "foo",
                                                        },
                                                    },
                                                    name: "foo",
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
                                                start: "line: 1, column: 10",
                                                end: "line: 1, column: 13",
                                                source: "bar",
                                            },
                                            errors: [
                                                "pipe destination must be a function call",
                                            ],
                                        },
                                        callee: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 10",
                                                        end: "line: 1, column: 13",
                                                        source: "bar",
                                                    },
                                                },
                                                name: "bar",
                                            },
                                        ),
                                        lparen: [],
                                        arguments: [],
                                        rparen: [],
                                    },
                                },
                            ),
                        },
                    ),
                ],
                eof: [],
            }
        "#]],
    );
}

#[test]
fn two_variables_for_two_froms() {
    test_file(
        r#"howdy = from()
			doody = from()
			howdy|>count()
			doody|>sum()"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 4, column: 16",
                        source: "howdy = from()\n\t\t\tdoody = from()\n\t\t\thowdy|>count()\n\t\t\tdoody|>sum()",
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
                                    end: "line: 1, column: 15",
                                    source: "howdy = from()",
                                },
                            },
                            id: Identifier {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 6",
                                        source: "howdy",
                                    },
                                },
                                name: "howdy",
                            },
                            init: Call(
                                CallExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 9",
                                            end: "line: 1, column: 15",
                                            source: "from()",
                                        },
                                    },
                                    callee: Identifier(
                                        Identifier {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 9",
                                                    end: "line: 1, column: 13",
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
                        },
                    ),
                    Variable(
                        VariableAssgn {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 2, column: 4",
                                    end: "line: 2, column: 18",
                                    source: "doody = from()",
                                },
                            },
                            id: Identifier {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 2, column: 4",
                                        end: "line: 2, column: 9",
                                        source: "doody",
                                    },
                                },
                                name: "doody",
                            },
                            init: Call(
                                CallExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 2, column: 12",
                                            end: "line: 2, column: 18",
                                            source: "from()",
                                        },
                                    },
                                    callee: Identifier(
                                        Identifier {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 2, column: 12",
                                                    end: "line: 2, column: 16",
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
                        },
                    ),
                    Expr(
                        ExprStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 3, column: 4",
                                    end: "line: 3, column: 18",
                                    source: "howdy|>count()",
                                },
                            },
                            expression: PipeExpr(
                                PipeExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 3, column: 4",
                                            end: "line: 3, column: 18",
                                            source: "howdy|>count()",
                                        },
                                    },
                                    argument: Identifier(
                                        Identifier {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 3, column: 4",
                                                    end: "line: 3, column: 9",
                                                    source: "howdy",
                                                },
                                            },
                                            name: "howdy",
                                        },
                                    ),
                                    call: CallExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 3, column: 11",
                                                end: "line: 3, column: 18",
                                                source: "count()",
                                            },
                                        },
                                        callee: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 3, column: 11",
                                                        end: "line: 3, column: 16",
                                                        source: "count",
                                                    },
                                                },
                                                name: "count",
                                            },
                                        ),
                                        lparen: [],
                                        arguments: [],
                                        rparen: [],
                                    },
                                },
                            ),
                        },
                    ),
                    Expr(
                        ExprStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 4, column: 4",
                                    end: "line: 4, column: 16",
                                    source: "doody|>sum()",
                                },
                            },
                            expression: PipeExpr(
                                PipeExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 4, column: 4",
                                            end: "line: 4, column: 16",
                                            source: "doody|>sum()",
                                        },
                                    },
                                    argument: Identifier(
                                        Identifier {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 4, column: 4",
                                                    end: "line: 4, column: 9",
                                                    source: "doody",
                                                },
                                            },
                                            name: "doody",
                                        },
                                    ),
                                    call: CallExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 4, column: 11",
                                                end: "line: 4, column: 16",
                                                source: "sum()",
                                            },
                                        },
                                        callee: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 4, column: 11",
                                                        end: "line: 4, column: 14",
                                                        source: "sum",
                                                    },
                                                },
                                                name: "sum",
                                            },
                                        ),
                                        lparen: [],
                                        arguments: [],
                                        rparen: [],
                                    },
                                },
                            ),
                        },
                    ),
                ],
                eof: [],
            }
        "#]],
    );
}

#[test]
fn index_expression() {
    test_file(
        r#"a[3]"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 5",
                        source: "a[3]",
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
                                    source: "a[3]",
                                },
                            },
                            expression: Index(
                                IndexExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 1",
                                            end: "line: 1, column: 5",
                                            source: "a[3]",
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
                                    index: Integer(
                                        IntegerLit {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 3",
                                                    end: "line: 1, column: 4",
                                                    source: "3",
                                                },
                                            },
                                            value: 3,
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
        "#]],
    );
}

#[test]
fn nested_index_expression() {
    test_file(
        r#"a[3][5]"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 8",
                        source: "a[3][5]",
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
                                    source: "a[3][5]",
                                },
                            },
                            expression: Index(
                                IndexExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 1",
                                            end: "line: 1, column: 8",
                                            source: "a[3][5]",
                                        },
                                    },
                                    array: Index(
                                        IndexExpr {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 1",
                                                    end: "line: 1, column: 5",
                                                    source: "a[3]",
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
                                            index: Integer(
                                                IntegerLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 3",
                                                            end: "line: 1, column: 4",
                                                            source: "3",
                                                        },
                                                    },
                                                    value: 3,
                                                },
                                            ),
                                            rbrack: [],
                                        },
                                    ),
                                    lbrack: [],
                                    index: Integer(
                                        IntegerLit {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 6",
                                                    end: "line: 1, column: 7",
                                                    source: "5",
                                                },
                                            },
                                            value: 5,
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
        "#]],
    );
}

#[test]
fn access_indexed_object_returned_from_function_call() {
    test_file(
        r#"f()[3]"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 7",
                        source: "f()[3]",
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
                                    end: "line: 1, column: 7",
                                    source: "f()[3]",
                                },
                            },
                            expression: Index(
                                IndexExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 1",
                                            end: "line: 1, column: 7",
                                            source: "f()[3]",
                                        },
                                    },
                                    array: Call(
                                        CallExpr {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 1",
                                                    end: "line: 1, column: 4",
                                                    source: "f()",
                                                },
                                            },
                                            callee: Identifier(
                                                Identifier {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 1",
                                                            end: "line: 1, column: 2",
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
                                    lbrack: [],
                                    index: Integer(
                                        IntegerLit {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 5",
                                                    end: "line: 1, column: 6",
                                                    source: "3",
                                                },
                                            },
                                            value: 3,
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
        "#]],
    );
}

#[test]
fn index_with_member_expressions() {
    test_file(
        r#"a.b["c"]"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 9",
                        source: "a.b[\"c\"]",
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
                                    end: "line: 1, column: 9",
                                    source: "a.b[\"c\"]",
                                },
                            },
                            expression: Member(
                                MemberExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 1",
                                            end: "line: 1, column: 9",
                                            source: "a.b[\"c\"]",
                                        },
                                    },
                                    object: Member(
                                        MemberExpr {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 1",
                                                    end: "line: 1, column: 4",
                                                    source: "a.b",
                                                },
                                            },
                                            object: Identifier(
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
                                            property: Identifier(
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
                                    lbrack: [],
                                    property: StringLit(
                                        StringLit {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 5",
                                                    end: "line: 1, column: 8",
                                                    source: "\"c\"",
                                                },
                                            },
                                            value: "c",
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
        "#]],
    );
}

#[test]
fn index_with_member_with_call_expression() {
    test_file(
        r#"a.b()["c"]"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 11",
                        source: "a.b()[\"c\"]",
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
                                    source: "a.b()[\"c\"]",
                                },
                            },
                            expression: Member(
                                MemberExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 1",
                                            end: "line: 1, column: 11",
                                            source: "a.b()[\"c\"]",
                                        },
                                    },
                                    object: Call(
                                        CallExpr {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 1",
                                                    end: "line: 1, column: 6",
                                                    source: "a.b()",
                                                },
                                            },
                                            callee: Member(
                                                MemberExpr {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 1",
                                                            end: "line: 1, column: 4",
                                                            source: "a.b",
                                                        },
                                                    },
                                                    object: Identifier(
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
                                                    property: Identifier(
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
                                            lparen: [],
                                            arguments: [],
                                            rparen: [],
                                        },
                                    ),
                                    lbrack: [],
                                    property: StringLit(
                                        StringLit {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 7",
                                                    end: "line: 1, column: 10",
                                                    source: "\"c\"",
                                                },
                                            },
                                            value: "c",
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
        "#]],
    );
}

#[test]
fn expressions_with_function_calls() {
    test_file(
        r#"a = foo() == 10"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 16",
                        source: "a = foo() == 10",
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
                                    end: "line: 1, column: 16",
                                    source: "a = foo() == 10",
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
                            init: Binary(
                                BinaryExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 5",
                                            end: "line: 1, column: 16",
                                            source: "foo() == 10",
                                        },
                                    },
                                    operator: EqualOperator,
                                    left: Call(
                                        CallExpr {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 5",
                                                    end: "line: 1, column: 10",
                                                    source: "foo()",
                                                },
                                            },
                                            callee: Identifier(
                                                Identifier {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 5",
                                                            end: "line: 1, column: 8",
                                                            source: "foo",
                                                        },
                                                    },
                                                    name: "foo",
                                                },
                                            ),
                                            lparen: [],
                                            arguments: [],
                                            rparen: [],
                                        },
                                    ),
                                    right: Integer(
                                        IntegerLit {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 14",
                                                    end: "line: 1, column: 16",
                                                    source: "10",
                                                },
                                            },
                                            value: 10,
                                        },
                                    ),
                                },
                            ),
                        },
                    ),
                ],
                eof: [],
            }
        "#]],
    );
}

#[test]
fn conditional() {
    test_file(
        r#"a = if true then 0 else 1"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 26",
                        source: "a = if true then 0 else 1",
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
                                    end: "line: 1, column: 26",
                                    source: "a = if true then 0 else 1",
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
                            init: Conditional(
                                ConditionalExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 5",
                                            end: "line: 1, column: 26",
                                            source: "if true then 0 else 1",
                                        },
                                    },
                                    tk_if: [],
                                    test: Identifier(
                                        Identifier {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 8",
                                                    end: "line: 1, column: 12",
                                                    source: "true",
                                                },
                                            },
                                            name: "true",
                                        },
                                    ),
                                    tk_then: [],
                                    consequent: Integer(
                                        IntegerLit {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 18",
                                                    end: "line: 1, column: 19",
                                                    source: "0",
                                                },
                                            },
                                            value: 0,
                                        },
                                    ),
                                    tk_else: [],
                                    alternate: Integer(
                                        IntegerLit {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 25",
                                                    end: "line: 1, column: 26",
                                                    source: "1",
                                                },
                                            },
                                            value: 1,
                                        },
                                    ),
                                },
                            ),
                        },
                    ),
                ],
                eof: [],
            }
        "#]],
    );
}

#[test]
fn conditional_with_unary_logical_operators() {
    test_file(
        r#"a = if exists b or c < d and not e == f then not exists (g - h) else exists exists i"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 85",
                        source: "a = if exists b or c < d and not e == f then not exists (g - h) else exists exists i",
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
                                    end: "line: 1, column: 85",
                                    source: "a = if exists b or c < d and not e == f then not exists (g - h) else exists exists i",
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
                            init: Conditional(
                                ConditionalExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 5",
                                            end: "line: 1, column: 85",
                                            source: "if exists b or c < d and not e == f then not exists (g - h) else exists exists i",
                                        },
                                    },
                                    tk_if: [],
                                    test: Logical(
                                        LogicalExpr {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 8",
                                                    end: "line: 1, column: 40",
                                                    source: "exists b or c < d and not e == f",
                                                },
                                            },
                                            operator: OrOperator,
                                            left: Unary(
                                                UnaryExpr {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 8",
                                                            end: "line: 1, column: 16",
                                                            source: "exists b",
                                                        },
                                                    },
                                                    operator: ExistsOperator,
                                                    argument: Identifier(
                                                        Identifier {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 1, column: 15",
                                                                    end: "line: 1, column: 16",
                                                                    source: "b",
                                                                },
                                                            },
                                                            name: "b",
                                                        },
                                                    ),
                                                },
                                            ),
                                            right: Logical(
                                                LogicalExpr {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 20",
                                                            end: "line: 1, column: 40",
                                                            source: "c < d and not e == f",
                                                        },
                                                    },
                                                    operator: AndOperator,
                                                    left: Binary(
                                                        BinaryExpr {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 1, column: 20",
                                                                    end: "line: 1, column: 25",
                                                                    source: "c < d",
                                                                },
                                                            },
                                                            operator: LessThanOperator,
                                                            left: Identifier(
                                                                Identifier {
                                                                    base: BaseNode {
                                                                        location: SourceLocation {
                                                                            start: "line: 1, column: 20",
                                                                            end: "line: 1, column: 21",
                                                                            source: "c",
                                                                        },
                                                                    },
                                                                    name: "c",
                                                                },
                                                            ),
                                                            right: Identifier(
                                                                Identifier {
                                                                    base: BaseNode {
                                                                        location: SourceLocation {
                                                                            start: "line: 1, column: 24",
                                                                            end: "line: 1, column: 25",
                                                                            source: "d",
                                                                        },
                                                                    },
                                                                    name: "d",
                                                                },
                                                            ),
                                                        },
                                                    ),
                                                    right: Unary(
                                                        UnaryExpr {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 1, column: 30",
                                                                    end: "line: 1, column: 40",
                                                                    source: "not e == f",
                                                                },
                                                            },
                                                            operator: NotOperator,
                                                            argument: Binary(
                                                                BinaryExpr {
                                                                    base: BaseNode {
                                                                        location: SourceLocation {
                                                                            start: "line: 1, column: 34",
                                                                            end: "line: 1, column: 40",
                                                                            source: "e == f",
                                                                        },
                                                                    },
                                                                    operator: EqualOperator,
                                                                    left: Identifier(
                                                                        Identifier {
                                                                            base: BaseNode {
                                                                                location: SourceLocation {
                                                                                    start: "line: 1, column: 34",
                                                                                    end: "line: 1, column: 35",
                                                                                    source: "e",
                                                                                },
                                                                            },
                                                                            name: "e",
                                                                        },
                                                                    ),
                                                                    right: Identifier(
                                                                        Identifier {
                                                                            base: BaseNode {
                                                                                location: SourceLocation {
                                                                                    start: "line: 1, column: 39",
                                                                                    end: "line: 1, column: 40",
                                                                                    source: "f",
                                                                                },
                                                                            },
                                                                            name: "f",
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
                                    tk_then: [],
                                    consequent: Unary(
                                        UnaryExpr {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 46",
                                                    end: "line: 1, column: 64",
                                                    source: "not exists (g - h)",
                                                },
                                            },
                                            operator: NotOperator,
                                            argument: Unary(
                                                UnaryExpr {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 50",
                                                            end: "line: 1, column: 64",
                                                            source: "exists (g - h)",
                                                        },
                                                    },
                                                    operator: ExistsOperator,
                                                    argument: Paren(
                                                        ParenExpr {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 1, column: 57",
                                                                    end: "line: 1, column: 64",
                                                                    source: "(g - h)",
                                                                },
                                                            },
                                                            lparen: [],
                                                            expression: Binary(
                                                                BinaryExpr {
                                                                    base: BaseNode {
                                                                        location: SourceLocation {
                                                                            start: "line: 1, column: 58",
                                                                            end: "line: 1, column: 63",
                                                                            source: "g - h",
                                                                        },
                                                                    },
                                                                    operator: SubtractionOperator,
                                                                    left: Identifier(
                                                                        Identifier {
                                                                            base: BaseNode {
                                                                                location: SourceLocation {
                                                                                    start: "line: 1, column: 58",
                                                                                    end: "line: 1, column: 59",
                                                                                    source: "g",
                                                                                },
                                                                            },
                                                                            name: "g",
                                                                        },
                                                                    ),
                                                                    right: Identifier(
                                                                        Identifier {
                                                                            base: BaseNode {
                                                                                location: SourceLocation {
                                                                                    start: "line: 1, column: 62",
                                                                                    end: "line: 1, column: 63",
                                                                                    source: "h",
                                                                                },
                                                                            },
                                                                            name: "h",
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
                                    tk_else: [],
                                    alternate: Unary(
                                        UnaryExpr {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 70",
                                                    end: "line: 1, column: 85",
                                                    source: "exists exists i",
                                                },
                                            },
                                            operator: ExistsOperator,
                                            argument: Unary(
                                                UnaryExpr {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 77",
                                                            end: "line: 1, column: 85",
                                                            source: "exists i",
                                                        },
                                                    },
                                                    operator: ExistsOperator,
                                                    argument: Identifier(
                                                        Identifier {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 1, column: 84",
                                                                    end: "line: 1, column: 85",
                                                                    source: "i",
                                                                },
                                                            },
                                                            name: "i",
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
        "#]],
    );
}

#[test]
fn nested_conditionals() {
    test_file(
        r#"if if b < 0 then true else false
                  then if c > 0 then 30 else 60
                  else if d == 0 then 90 else 120"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 3, column: 50",
                        source: "if if b < 0 then true else false\n                  then if c > 0 then 30 else 60\n                  else if d == 0 then 90 else 120",
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
                                    end: "line: 3, column: 50",
                                    source: "if if b < 0 then true else false\n                  then if c > 0 then 30 else 60\n                  else if d == 0 then 90 else 120",
                                },
                            },
                            expression: Conditional(
                                ConditionalExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 1",
                                            end: "line: 3, column: 50",
                                            source: "if if b < 0 then true else false\n                  then if c > 0 then 30 else 60\n                  else if d == 0 then 90 else 120",
                                        },
                                    },
                                    tk_if: [],
                                    test: Conditional(
                                        ConditionalExpr {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 4",
                                                    end: "line: 1, column: 33",
                                                    source: "if b < 0 then true else false",
                                                },
                                            },
                                            tk_if: [],
                                            test: Binary(
                                                BinaryExpr {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 7",
                                                            end: "line: 1, column: 12",
                                                            source: "b < 0",
                                                        },
                                                    },
                                                    operator: LessThanOperator,
                                                    left: Identifier(
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
                                                    right: Integer(
                                                        IntegerLit {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 1, column: 11",
                                                                    end: "line: 1, column: 12",
                                                                    source: "0",
                                                                },
                                                            },
                                                            value: 0,
                                                        },
                                                    ),
                                                },
                                            ),
                                            tk_then: [],
                                            consequent: Identifier(
                                                Identifier {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 18",
                                                            end: "line: 1, column: 22",
                                                            source: "true",
                                                        },
                                                    },
                                                    name: "true",
                                                },
                                            ),
                                            tk_else: [],
                                            alternate: Identifier(
                                                Identifier {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 28",
                                                            end: "line: 1, column: 33",
                                                            source: "false",
                                                        },
                                                    },
                                                    name: "false",
                                                },
                                            ),
                                        },
                                    ),
                                    tk_then: [],
                                    consequent: Conditional(
                                        ConditionalExpr {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 2, column: 24",
                                                    end: "line: 2, column: 48",
                                                    source: "if c > 0 then 30 else 60",
                                                },
                                            },
                                            tk_if: [],
                                            test: Binary(
                                                BinaryExpr {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 2, column: 27",
                                                            end: "line: 2, column: 32",
                                                            source: "c > 0",
                                                        },
                                                    },
                                                    operator: GreaterThanOperator,
                                                    left: Identifier(
                                                        Identifier {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 2, column: 27",
                                                                    end: "line: 2, column: 28",
                                                                    source: "c",
                                                                },
                                                            },
                                                            name: "c",
                                                        },
                                                    ),
                                                    right: Integer(
                                                        IntegerLit {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 2, column: 31",
                                                                    end: "line: 2, column: 32",
                                                                    source: "0",
                                                                },
                                                            },
                                                            value: 0,
                                                        },
                                                    ),
                                                },
                                            ),
                                            tk_then: [],
                                            consequent: Integer(
                                                IntegerLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 2, column: 38",
                                                            end: "line: 2, column: 40",
                                                            source: "30",
                                                        },
                                                    },
                                                    value: 30,
                                                },
                                            ),
                                            tk_else: [],
                                            alternate: Integer(
                                                IntegerLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 2, column: 46",
                                                            end: "line: 2, column: 48",
                                                            source: "60",
                                                        },
                                                    },
                                                    value: 60,
                                                },
                                            ),
                                        },
                                    ),
                                    tk_else: [],
                                    alternate: Conditional(
                                        ConditionalExpr {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 3, column: 24",
                                                    end: "line: 3, column: 50",
                                                    source: "if d == 0 then 90 else 120",
                                                },
                                            },
                                            tk_if: [],
                                            test: Binary(
                                                BinaryExpr {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 3, column: 27",
                                                            end: "line: 3, column: 33",
                                                            source: "d == 0",
                                                        },
                                                    },
                                                    operator: EqualOperator,
                                                    left: Identifier(
                                                        Identifier {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 3, column: 27",
                                                                    end: "line: 3, column: 28",
                                                                    source: "d",
                                                                },
                                                            },
                                                            name: "d",
                                                        },
                                                    ),
                                                    right: Integer(
                                                        IntegerLit {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 3, column: 32",
                                                                    end: "line: 3, column: 33",
                                                                    source: "0",
                                                                },
                                                            },
                                                            value: 0,
                                                        },
                                                    ),
                                                },
                                            ),
                                            tk_then: [],
                                            consequent: Integer(
                                                IntegerLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 3, column: 39",
                                                            end: "line: 3, column: 41",
                                                            source: "90",
                                                        },
                                                    },
                                                    value: 90,
                                                },
                                            ),
                                            tk_else: [],
                                            alternate: Integer(
                                                IntegerLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 3, column: 47",
                                                            end: "line: 3, column: 50",
                                                            source: "120",
                                                        },
                                                    },
                                                    value: 120,
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
        "#]],
    );
}

#[test]
fn parse_testcase() {
    test_file(
        r#"testcase my_test { a = 1 }"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 27",
                        source: "testcase my_test { a = 1 }",
                    },
                },
                name: "",
                metadata: "parser-type=rust",
                package: None,
                imports: [],
                body: [
                    TestCase(
                        TestCaseStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 1",
                                    end: "line: 1, column: 27",
                                    source: "testcase my_test { a = 1 }",
                                },
                            },
                            id: Identifier {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 10",
                                        end: "line: 1, column: 17",
                                        source: "my_test",
                                    },
                                },
                                name: "my_test",
                            },
                            extends: None,
                            block: Block {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 18",
                                        end: "line: 1, column: 27",
                                        source: "{ a = 1 }",
                                    },
                                },
                                lbrace: [],
                                body: [
                                    Variable(
                                        VariableAssgn {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 20",
                                                    end: "line: 1, column: 25",
                                                    source: "a = 1",
                                                },
                                            },
                                            id: Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 20",
                                                        end: "line: 1, column: 21",
                                                        source: "a",
                                                    },
                                                },
                                                name: "a",
                                            },
                                            init: Integer(
                                                IntegerLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 24",
                                                            end: "line: 1, column: 25",
                                                            source: "1",
                                                        },
                                                    },
                                                    value: 1,
                                                },
                                            ),
                                        },
                                    ),
                                ],
                                rbrace: [],
                            },
                        },
                    ),
                ],
                eof: [],
            }
        "#]],
    );
}

#[test]
fn parse_testcase_extends() {
    test_file(
        r#"testcase my_test extends "other_test" { a = 1 }"#,
        expect![[r#"
            File {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 48",
                        source: "testcase my_test extends \"other_test\" { a = 1 }",
                    },
                },
                name: "",
                metadata: "parser-type=rust",
                package: None,
                imports: [],
                body: [
                    TestCase(
                        TestCaseStmt {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 1",
                                    end: "line: 1, column: 48",
                                    source: "testcase my_test extends \"other_test\" { a = 1 }",
                                },
                            },
                            id: Identifier {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 10",
                                        end: "line: 1, column: 17",
                                        source: "my_test",
                                    },
                                },
                                name: "my_test",
                            },
                            extends: Some(
                                StringLit {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 26",
                                            end: "line: 1, column: 38",
                                            source: "\"other_test\"",
                                        },
                                    },
                                    value: "other_test",
                                },
                            ),
                            block: Block {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 39",
                                        end: "line: 1, column: 48",
                                        source: "{ a = 1 }",
                                    },
                                },
                                lbrace: [],
                                body: [
                                    Variable(
                                        VariableAssgn {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 41",
                                                    end: "line: 1, column: 46",
                                                    source: "a = 1",
                                                },
                                            },
                                            id: Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 41",
                                                        end: "line: 1, column: 42",
                                                        source: "a",
                                                    },
                                                },
                                                name: "a",
                                            },
                                            init: Integer(
                                                IntegerLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 45",
                                                            end: "line: 1, column: 46",
                                                            source: "1",
                                                        },
                                                    },
                                                    value: 1,
                                                },
                                            ),
                                        },
                                    ),
                                ],
                                rbrace: [],
                            },
                        },
                    ),
                ],
                eof: [],
            }
        "#]],
    );
}
