use pretty_assertions::assert_eq;

use super::*;

#[test]
fn parse_string_literal() {
    let errors: Vec<String> = vec![];

    let mut p = Parser::new(r#""Hello world""#);
    let result = p.parse_string_literal();
    assert_eq!("Hello world".to_string(), result.value);
    assert_eq!(errors, result.base.errors);
}

#[test]
fn parse_string_literal_invalid_string() {
    let errors = vec![
        "expected STRING, got QUOTE (\") at 1:1",
        "invalid string literal",
    ];

    let mut p = Parser::new(r#"""#);
    let result = p.parse_string_literal();
    assert_eq!("".to_string(), result.value);
    assert_eq!(errors, result.base.errors);
}

#[test]
fn string_interpolation_simple() {
    let mut p = Parser::new(r#""a + b = ${a + b}""#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 19",
                    source: "\"a + b = ${a + b}\"",
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
                                source: "\"a + b = ${a + b}\"",
                            },
                        },
                        expression: StringExpr(
                            StringExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 19",
                                        source: "\"a + b = ${a + b}\"",
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
                                ],
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
fn string_interpolation_array() {
    let mut p = Parser::new(r#"a = ["influx", "test", "InfluxOfflineTimeAlert", "acu:${r.a}"]"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 63",
                    source: "a = [\"influx\", \"test\", \"InfluxOfflineTimeAlert\", \"acu:${r.a}\"]",
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
                                end: "line: 1, column: 63",
                                source: "a = [\"influx\", \"test\", \"InfluxOfflineTimeAlert\", \"acu:${r.a}\"]",
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
                        init: Array(
                            ArrayExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 5",
                                        end: "line: 1, column: 63",
                                        source: "[\"influx\", \"test\", \"InfluxOfflineTimeAlert\", \"acu:${r.a}\"]",
                                    },
                                },
                                lbrack: [],
                                elements: [
                                    ArrayItem {
                                        expression: StringLit(
                                            StringLit {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 6",
                                                        end: "line: 1, column: 14",
                                                        source: "\"influx\"",
                                                    },
                                                },
                                                value: "influx",
                                            },
                                        ),
                                        comma: [],
                                    },
                                    ArrayItem {
                                        expression: StringLit(
                                            StringLit {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 16",
                                                        end: "line: 1, column: 22",
                                                        source: "\"test\"",
                                                    },
                                                },
                                                value: "test",
                                            },
                                        ),
                                        comma: [],
                                    },
                                    ArrayItem {
                                        expression: StringLit(
                                            StringLit {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 24",
                                                        end: "line: 1, column: 48",
                                                        source: "\"InfluxOfflineTimeAlert\"",
                                                    },
                                                },
                                                value: "InfluxOfflineTimeAlert",
                                            },
                                        ),
                                        comma: [],
                                    },
                                    ArrayItem {
                                        expression: StringExpr(
                                            StringExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 50",
                                                        end: "line: 1, column: 62",
                                                        source: "\"acu:${r.a}\"",
                                                    },
                                                },
                                                parts: [
                                                    Text(
                                                        TextPart {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 1, column: 51",
                                                                    end: "line: 1, column: 55",
                                                                    source: "acu:",
                                                                },
                                                            },
                                                            value: "acu:",
                                                        },
                                                    ),
                                                    Interpolated(
                                                        InterpolatedPart {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 1, column: 55",
                                                                    end: "line: 1, column: 61",
                                                                    source: "${r.a}",
                                                                },
                                                            },
                                                            expression: Member(
                                                                MemberExpr {
                                                                    base: BaseNode {
                                                                        location: SourceLocation {
                                                                            start: "line: 1, column: 57",
                                                                            end: "line: 1, column: 60",
                                                                            source: "r.a",
                                                                        },
                                                                    },
                                                                    object: Identifier(
                                                                        Identifier {
                                                                            base: BaseNode {
                                                                                location: SourceLocation {
                                                                                    start: "line: 1, column: 57",
                                                                                    end: "line: 1, column: 58",
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
                                                                                    start: "line: 1, column: 59",
                                                                                    end: "line: 1, column: 60",
                                                                                    source: "a",
                                                                                },
                                                                            },
                                                                            name: "a",
                                                                        },
                                                                    ),
                                                                    rbrack: [],
                                                                },
                                                            ),
                                                        },
                                                    ),
                                                ],
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
    "#]].assert_debug_eq(&parsed);
}

#[test]
fn string_interpolation_multiple() {
    let mut p = Parser::new(r#""a = ${a} and b = ${b}""#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 24",
                    source: "\"a = ${a} and b = ${b}\"",
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
                                end: "line: 1, column: 24",
                                source: "\"a = ${a} and b = ${b}\"",
                            },
                        },
                        expression: StringExpr(
                            StringExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 24",
                                        source: "\"a = ${a} and b = ${b}\"",
                                    },
                                },
                                parts: [
                                    Text(
                                        TextPart {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 2",
                                                    end: "line: 1, column: 6",
                                                    source: "a = ",
                                                },
                                            },
                                            value: "a = ",
                                        },
                                    ),
                                    Interpolated(
                                        InterpolatedPart {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 6",
                                                    end: "line: 1, column: 10",
                                                    source: "${a}",
                                                },
                                            },
                                            expression: Identifier(
                                                Identifier {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 8",
                                                            end: "line: 1, column: 9",
                                                            source: "a",
                                                        },
                                                    },
                                                    name: "a",
                                                },
                                            ),
                                        },
                                    ),
                                    Text(
                                        TextPart {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 10",
                                                    end: "line: 1, column: 19",
                                                    source: " and b = ",
                                                },
                                            },
                                            value: " and b = ",
                                        },
                                    ),
                                    Interpolated(
                                        InterpolatedPart {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 19",
                                                    end: "line: 1, column: 23",
                                                    source: "${b}",
                                                },
                                            },
                                            expression: Identifier(
                                                Identifier {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 21",
                                                            end: "line: 1, column: 22",
                                                            source: "b",
                                                        },
                                                    },
                                                    name: "b",
                                                },
                                            ),
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
    "#]]
    .assert_debug_eq(&parsed);
}

#[test]
fn string_interpolation_nested() {
    let mut p = Parser::new(r#""we ${"can ${"add" + "strings"}"} together""#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 44",
                    source: "\"we ${\"can ${\"add\" + \"strings\"}\"} together\"",
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
                                end: "line: 1, column: 44",
                                source: "\"we ${\"can ${\"add\" + \"strings\"}\"} together\"",
                            },
                        },
                        expression: StringExpr(
                            StringExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 44",
                                        source: "\"we ${\"can ${\"add\" + \"strings\"}\"} together\"",
                                    },
                                },
                                parts: [
                                    Text(
                                        TextPart {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 2",
                                                    end: "line: 1, column: 5",
                                                    source: "we ",
                                                },
                                            },
                                            value: "we ",
                                        },
                                    ),
                                    Interpolated(
                                        InterpolatedPart {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 5",
                                                    end: "line: 1, column: 34",
                                                    source: "${\"can ${\"add\" + \"strings\"}\"}",
                                                },
                                            },
                                            expression: StringExpr(
                                                StringExpr {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 7",
                                                            end: "line: 1, column: 33",
                                                            source: "\"can ${\"add\" + \"strings\"}\"",
                                                        },
                                                    },
                                                    parts: [
                                                        Text(
                                                            TextPart {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 1, column: 8",
                                                                        end: "line: 1, column: 12",
                                                                        source: "can ",
                                                                    },
                                                                },
                                                                value: "can ",
                                                            },
                                                        ),
                                                        Interpolated(
                                                            InterpolatedPart {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 1, column: 12",
                                                                        end: "line: 1, column: 32",
                                                                        source: "${\"add\" + \"strings\"}",
                                                                    },
                                                                },
                                                                expression: Binary(
                                                                    BinaryExpr {
                                                                        base: BaseNode {
                                                                            location: SourceLocation {
                                                                                start: "line: 1, column: 14",
                                                                                end: "line: 1, column: 31",
                                                                                source: "\"add\" + \"strings\"",
                                                                            },
                                                                        },
                                                                        operator: AdditionOperator,
                                                                        left: StringLit(
                                                                            StringLit {
                                                                                base: BaseNode {
                                                                                    location: SourceLocation {
                                                                                        start: "line: 1, column: 14",
                                                                                        end: "line: 1, column: 19",
                                                                                        source: "\"add\"",
                                                                                    },
                                                                                },
                                                                                value: "add",
                                                                            },
                                                                        ),
                                                                        right: StringLit(
                                                                            StringLit {
                                                                                base: BaseNode {
                                                                                    location: SourceLocation {
                                                                                        start: "line: 1, column: 22",
                                                                                        end: "line: 1, column: 31",
                                                                                        source: "\"strings\"",
                                                                                    },
                                                                                },
                                                                                value: "strings",
                                                                            },
                                                                        ),
                                                                    },
                                                                ),
                                                            },
                                                        ),
                                                    ],
                                                },
                                            ),
                                        },
                                    ),
                                    Text(
                                        TextPart {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 34",
                                                    end: "line: 1, column: 43",
                                                    source: " together",
                                                },
                                            },
                                            value: " together",
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
    "#]].assert_debug_eq(&parsed);
}

#[test]
fn string_interp_with_escapes() {
    let mut p = Parser::new(r#""string \"interpolation with ${"escapes"}\"""#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 45",
                    source: "\"string \\\"interpolation with ${\"escapes\"}\\\"\"",
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
                                end: "line: 1, column: 45",
                                source: "\"string \\\"interpolation with ${\"escapes\"}\\\"\"",
                            },
                        },
                        expression: StringExpr(
                            StringExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 45",
                                        source: "\"string \\\"interpolation with ${\"escapes\"}\\\"\"",
                                    },
                                },
                                parts: [
                                    Text(
                                        TextPart {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 2",
                                                    end: "line: 1, column: 30",
                                                    source: "string \\\"interpolation with ",
                                                },
                                            },
                                            value: "string \"interpolation with ",
                                        },
                                    ),
                                    Interpolated(
                                        InterpolatedPart {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 30",
                                                    end: "line: 1, column: 42",
                                                    source: "${\"escapes\"}",
                                                },
                                            },
                                            expression: StringLit(
                                                StringLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 32",
                                                            end: "line: 1, column: 41",
                                                            source: "\"escapes\"",
                                                        },
                                                    },
                                                    value: "escapes",
                                                },
                                            ),
                                        },
                                    ),
                                    Text(
                                        TextPart {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 42",
                                                    end: "line: 1, column: 44",
                                                    source: "\\\"",
                                                },
                                            },
                                            value: "\"",
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
    "#]].assert_debug_eq(&parsed);
}

#[test]
fn bad_string_expression() {
    let mut p = Parser::new(r#"fn = (a) => "${a}"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 18",
                    source: "fn = (a) => \"${a}",
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
                                end: "line: 1, column: 18",
                                source: "fn = (a) => \"${a}",
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
                                        end: "line: 1, column: 18",
                                        source: "(a) => \"${a}",
                                    },
                                },
                                lparen: [],
                                params: [
                                    Property {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 7",
                                                end: "line: 1, column: 8",
                                                source: "a",
                                            },
                                        },
                                        key: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 7",
                                                        end: "line: 1, column: 8",
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
                                ],
                                rparen: [],
                                arrow: [],
                                body: Expr(
                                    StringExpr(
                                        StringExpr {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 13",
                                                    end: "line: 1, column: 18",
                                                    source: "\"${a}",
                                                },
                                                errors: [
                                                    "got unexpected token in string expression @1:18-1:18: EOF",
                                                ],
                                            },
                                            parts: [],
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
    "#]].assert_debug_eq(&parsed);
}

#[test]
fn string_with_utf_8() {
    let mut p = Parser::new(r#""日本語""#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 12",
                    source: "\"日本語\"",
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
                                source: "\"日本語\"",
                            },
                        },
                        expression: StringLit(
                            StringLit {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 12",
                                        source: "\"日本語\"",
                                    },
                                },
                                value: "日本語",
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
fn string_with_byte_values() {
    let mut p = Parser::new(r#""\xe6\x97\xa5\xe6\x9c\xac\xe8\xaa\x9e""#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 39",
                    source: "\"\\xe6\\x97\\xa5\\xe6\\x9c\\xac\\xe8\\xaa\\x9e\"",
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
                                end: "line: 1, column: 39",
                                source: "\"\\xe6\\x97\\xa5\\xe6\\x9c\\xac\\xe8\\xaa\\x9e\"",
                            },
                        },
                        expression: StringLit(
                            StringLit {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 39",
                                        source: "\"\\xe6\\x97\\xa5\\xe6\\x9c\\xac\\xe8\\xaa\\x9e\"",
                                    },
                                },
                                value: "日本語",
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
fn string_with_mixed_values() {
    let mut p = Parser::new(r#""hello 日x本 \xe6\x97\xa5\xe6\x9c\xac\xe8\xaa\x9e \xc2\xb5s""#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 63",
                    source: "\"hello 日x本 \\xe6\\x97\\xa5\\xe6\\x9c\\xac\\xe8\\xaa\\x9e \\xc2\\xb5s\"",
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
                                end: "line: 1, column: 63",
                                source: "\"hello 日x本 \\xe6\\x97\\xa5\\xe6\\x9c\\xac\\xe8\\xaa\\x9e \\xc2\\xb5s\"",
                            },
                        },
                        expression: StringLit(
                            StringLit {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 63",
                                        source: "\"hello 日x本 \\xe6\\x97\\xa5\\xe6\\x9c\\xac\\xe8\\xaa\\x9e \\xc2\\xb5s\"",
                                    },
                                },
                                value: "hello 日x本 日本語 µs",
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
fn string_with_escapes() {
    let mut p = Parser::new(
        r#""newline \n
carriage return \r
horizontal tab \t
double quote \"
backslash \\
dollar curly bracket \${
""#,
    );
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 7, column: 2",
                    source: "\"newline \\n\ncarriage return \\r\nhorizontal tab \\t\ndouble quote \\\"\nbackslash \\\\\ndollar curly bracket \\${\n\"",
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
                                end: "line: 7, column: 2",
                                source: "\"newline \\n\ncarriage return \\r\nhorizontal tab \\t\ndouble quote \\\"\nbackslash \\\\\ndollar curly bracket \\${\n\"",
                            },
                        },
                        expression: StringLit(
                            StringLit {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 7, column: 2",
                                        source: "\"newline \\n\ncarriage return \\r\nhorizontal tab \\t\ndouble quote \\\"\nbackslash \\\\\ndollar curly bracket \\${\n\"",
                                    },
                                },
                                value: "newline \n\ncarriage return \r\nhorizontal tab \t\ndouble quote \"\nbackslash \\\ndollar curly bracket ${\n",
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
fn multiline_string() {
    let mut p = Parser::new(
        r#""
 this is a
multiline
string"
"#,
    );
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 5, column: 1",
                    source: "\"\n this is a\nmultiline\nstring\"\n",
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
                                end: "line: 4, column: 8",
                                source: "\"\n this is a\nmultiline\nstring\"",
                            },
                        },
                        expression: StringLit(
                            StringLit {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 4, column: 8",
                                        source: "\"\n this is a\nmultiline\nstring\"",
                                    },
                                },
                                value: "\n this is a\nmultiline\nstring",
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
