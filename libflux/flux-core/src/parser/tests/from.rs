use super::*;

#[test]
fn from() {
    let mut p = Parser::new(r#"from()"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 7",
                    source: "from()",
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
                                source: "from()",
                            },
                        },
                        expression: Call(
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
                    },
                ),
            ],
            eof: [],
        }
    "#]]
    .assert_debug_eq(&parsed);
}

#[test]
fn from_with_database() {
    let mut p = Parser::new(r#"from(bucket:"telegraf/autogen")"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 32",
                    source: "from(bucket:\"telegraf/autogen\")",
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
                                end: "line: 1, column: 32",
                                source: "from(bucket:\"telegraf/autogen\")",
                            },
                        },
                        expression: Call(
                            CallExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 32",
                                        source: "from(bucket:\"telegraf/autogen\")",
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
                                arguments: [
                                    Object(
                                        ObjectExpr {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 6",
                                                    end: "line: 1, column: 31",
                                                    source: "bucket:\"telegraf/autogen\"",
                                                },
                                            },
                                            lbrace: [],
                                            with: None,
                                            properties: [
                                                Property {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 6",
                                                            end: "line: 1, column: 31",
                                                            source: "bucket:\"telegraf/autogen\"",
                                                        },
                                                    },
                                                    key: Identifier(
                                                        Identifier {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 1, column: 6",
                                                                    end: "line: 1, column: 12",
                                                                    source: "bucket",
                                                                },
                                                            },
                                                            name: "bucket",
                                                        },
                                                    ),
                                                    separator: [],
                                                    value: Some(
                                                        StringLit(
                                                            StringLit {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 1, column: 13",
                                                                        end: "line: 1, column: 31",
                                                                        source: "\"telegraf/autogen\"",
                                                                    },
                                                                },
                                                                value: "telegraf/autogen",
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
                        ),
                    },
                ),
            ],
            eof: [],
        }
    "#]].assert_debug_eq(&parsed);
}

#[test]
fn from_with_filter_with_no_parens() {
    let mut p = Parser::new(
        r#"from(bucket:"telegraf/autogen").filter(fn: (r) => r["other"]=="mem" and r["this"]=="that" or r["these"]!="those")"#,
    );
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 114",
                    source: "from(bucket:\"telegraf/autogen\").filter(fn: (r) => r[\"other\"]==\"mem\" and r[\"this\"]==\"that\" or r[\"these\"]!=\"those\")",
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
                                end: "line: 1, column: 114",
                                source: "from(bucket:\"telegraf/autogen\").filter(fn: (r) => r[\"other\"]==\"mem\" and r[\"this\"]==\"that\" or r[\"these\"]!=\"those\")",
                            },
                        },
                        expression: Call(
                            CallExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 114",
                                        source: "from(bucket:\"telegraf/autogen\").filter(fn: (r) => r[\"other\"]==\"mem\" and r[\"this\"]==\"that\" or r[\"these\"]!=\"those\")",
                                    },
                                },
                                callee: Member(
                                    MemberExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 1",
                                                end: "line: 1, column: 39",
                                                source: "from(bucket:\"telegraf/autogen\").filter",
                                            },
                                        },
                                        object: Call(
                                            CallExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 1",
                                                        end: "line: 1, column: 32",
                                                        source: "from(bucket:\"telegraf/autogen\")",
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
                                                arguments: [
                                                    Object(
                                                        ObjectExpr {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 1, column: 6",
                                                                    end: "line: 1, column: 31",
                                                                    source: "bucket:\"telegraf/autogen\"",
                                                                },
                                                            },
                                                            lbrace: [],
                                                            with: None,
                                                            properties: [
                                                                Property {
                                                                    base: BaseNode {
                                                                        location: SourceLocation {
                                                                            start: "line: 1, column: 6",
                                                                            end: "line: 1, column: 31",
                                                                            source: "bucket:\"telegraf/autogen\"",
                                                                        },
                                                                    },
                                                                    key: Identifier(
                                                                        Identifier {
                                                                            base: BaseNode {
                                                                                location: SourceLocation {
                                                                                    start: "line: 1, column: 6",
                                                                                    end: "line: 1, column: 12",
                                                                                    source: "bucket",
                                                                                },
                                                                            },
                                                                            name: "bucket",
                                                                        },
                                                                    ),
                                                                    separator: [],
                                                                    value: Some(
                                                                        StringLit(
                                                                            StringLit {
                                                                                base: BaseNode {
                                                                                    location: SourceLocation {
                                                                                        start: "line: 1, column: 13",
                                                                                        end: "line: 1, column: 31",
                                                                                        source: "\"telegraf/autogen\"",
                                                                                    },
                                                                                },
                                                                                value: "telegraf/autogen",
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
                                        ),
                                        lbrack: [],
                                        property: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 33",
                                                        end: "line: 1, column: 39",
                                                        source: "filter",
                                                    },
                                                },
                                                name: "filter",
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
                                                    start: "line: 1, column: 40",
                                                    end: "line: 1, column: 113",
                                                    source: "fn: (r) => r[\"other\"]==\"mem\" and r[\"this\"]==\"that\" or r[\"these\"]!=\"those\"",
                                                },
                                            },
                                            lbrace: [],
                                            with: None,
                                            properties: [
                                                Property {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 40",
                                                            end: "line: 1, column: 113",
                                                            source: "fn: (r) => r[\"other\"]==\"mem\" and r[\"this\"]==\"that\" or r[\"these\"]!=\"those\"",
                                                        },
                                                    },
                                                    key: Identifier(
                                                        Identifier {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 1, column: 40",
                                                                    end: "line: 1, column: 42",
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
                                                                        start: "line: 1, column: 44",
                                                                        end: "line: 1, column: 113",
                                                                        source: "(r) => r[\"other\"]==\"mem\" and r[\"this\"]==\"that\" or r[\"these\"]!=\"those\"",
                                                                    },
                                                                },
                                                                lparen: [],
                                                                params: [
                                                                    Property {
                                                                        base: BaseNode {
                                                                            location: SourceLocation {
                                                                                start: "line: 1, column: 45",
                                                                                end: "line: 1, column: 46",
                                                                                source: "r",
                                                                            },
                                                                        },
                                                                        key: Identifier(
                                                                            Identifier {
                                                                                base: BaseNode {
                                                                                    location: SourceLocation {
                                                                                        start: "line: 1, column: 45",
                                                                                        end: "line: 1, column: 46",
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
                                                                body: Expr(
                                                                    Logical(
                                                                        LogicalExpr {
                                                                            base: BaseNode {
                                                                                location: SourceLocation {
                                                                                    start: "line: 1, column: 51",
                                                                                    end: "line: 1, column: 113",
                                                                                    source: "r[\"other\"]==\"mem\" and r[\"this\"]==\"that\" or r[\"these\"]!=\"those\"",
                                                                                },
                                                                            },
                                                                            operator: OrOperator,
                                                                            left: Logical(
                                                                                LogicalExpr {
                                                                                    base: BaseNode {
                                                                                        location: SourceLocation {
                                                                                            start: "line: 1, column: 51",
                                                                                            end: "line: 1, column: 90",
                                                                                            source: "r[\"other\"]==\"mem\" and r[\"this\"]==\"that\"",
                                                                                        },
                                                                                    },
                                                                                    operator: AndOperator,
                                                                                    left: Binary(
                                                                                        BinaryExpr {
                                                                                            base: BaseNode {
                                                                                                location: SourceLocation {
                                                                                                    start: "line: 1, column: 51",
                                                                                                    end: "line: 1, column: 68",
                                                                                                    source: "r[\"other\"]==\"mem\"",
                                                                                                },
                                                                                            },
                                                                                            operator: EqualOperator,
                                                                                            left: Member(
                                                                                                MemberExpr {
                                                                                                    base: BaseNode {
                                                                                                        location: SourceLocation {
                                                                                                            start: "line: 1, column: 51",
                                                                                                            end: "line: 1, column: 61",
                                                                                                            source: "r[\"other\"]",
                                                                                                        },
                                                                                                    },
                                                                                                    object: Identifier(
                                                                                                        Identifier {
                                                                                                            base: BaseNode {
                                                                                                                location: SourceLocation {
                                                                                                                    start: "line: 1, column: 51",
                                                                                                                    end: "line: 1, column: 52",
                                                                                                                    source: "r",
                                                                                                                },
                                                                                                            },
                                                                                                            name: "r",
                                                                                                        },
                                                                                                    ),
                                                                                                    lbrack: [],
                                                                                                    property: StringLit(
                                                                                                        StringLit {
                                                                                                            base: BaseNode {
                                                                                                                location: SourceLocation {
                                                                                                                    start: "line: 1, column: 53",
                                                                                                                    end: "line: 1, column: 60",
                                                                                                                    source: "\"other\"",
                                                                                                                },
                                                                                                            },
                                                                                                            value: "other",
                                                                                                        },
                                                                                                    ),
                                                                                                    rbrack: [],
                                                                                                },
                                                                                            ),
                                                                                            right: StringLit(
                                                                                                StringLit {
                                                                                                    base: BaseNode {
                                                                                                        location: SourceLocation {
                                                                                                            start: "line: 1, column: 63",
                                                                                                            end: "line: 1, column: 68",
                                                                                                            source: "\"mem\"",
                                                                                                        },
                                                                                                    },
                                                                                                    value: "mem",
                                                                                                },
                                                                                            ),
                                                                                        },
                                                                                    ),
                                                                                    right: Binary(
                                                                                        BinaryExpr {
                                                                                            base: BaseNode {
                                                                                                location: SourceLocation {
                                                                                                    start: "line: 1, column: 73",
                                                                                                    end: "line: 1, column: 90",
                                                                                                    source: "r[\"this\"]==\"that\"",
                                                                                                },
                                                                                            },
                                                                                            operator: EqualOperator,
                                                                                            left: Member(
                                                                                                MemberExpr {
                                                                                                    base: BaseNode {
                                                                                                        location: SourceLocation {
                                                                                                            start: "line: 1, column: 73",
                                                                                                            end: "line: 1, column: 82",
                                                                                                            source: "r[\"this\"]",
                                                                                                        },
                                                                                                    },
                                                                                                    object: Identifier(
                                                                                                        Identifier {
                                                                                                            base: BaseNode {
                                                                                                                location: SourceLocation {
                                                                                                                    start: "line: 1, column: 73",
                                                                                                                    end: "line: 1, column: 74",
                                                                                                                    source: "r",
                                                                                                                },
                                                                                                            },
                                                                                                            name: "r",
                                                                                                        },
                                                                                                    ),
                                                                                                    lbrack: [],
                                                                                                    property: StringLit(
                                                                                                        StringLit {
                                                                                                            base: BaseNode {
                                                                                                                location: SourceLocation {
                                                                                                                    start: "line: 1, column: 75",
                                                                                                                    end: "line: 1, column: 81",
                                                                                                                    source: "\"this\"",
                                                                                                                },
                                                                                                            },
                                                                                                            value: "this",
                                                                                                        },
                                                                                                    ),
                                                                                                    rbrack: [],
                                                                                                },
                                                                                            ),
                                                                                            right: StringLit(
                                                                                                StringLit {
                                                                                                    base: BaseNode {
                                                                                                        location: SourceLocation {
                                                                                                            start: "line: 1, column: 84",
                                                                                                            end: "line: 1, column: 90",
                                                                                                            source: "\"that\"",
                                                                                                        },
                                                                                                    },
                                                                                                    value: "that",
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
                                                                                            start: "line: 1, column: 94",
                                                                                            end: "line: 1, column: 113",
                                                                                            source: "r[\"these\"]!=\"those\"",
                                                                                        },
                                                                                    },
                                                                                    operator: NotEqualOperator,
                                                                                    left: Member(
                                                                                        MemberExpr {
                                                                                            base: BaseNode {
                                                                                                location: SourceLocation {
                                                                                                    start: "line: 1, column: 94",
                                                                                                    end: "line: 1, column: 104",
                                                                                                    source: "r[\"these\"]",
                                                                                                },
                                                                                            },
                                                                                            object: Identifier(
                                                                                                Identifier {
                                                                                                    base: BaseNode {
                                                                                                        location: SourceLocation {
                                                                                                            start: "line: 1, column: 94",
                                                                                                            end: "line: 1, column: 95",
                                                                                                            source: "r",
                                                                                                        },
                                                                                                    },
                                                                                                    name: "r",
                                                                                                },
                                                                                            ),
                                                                                            lbrack: [],
                                                                                            property: StringLit(
                                                                                                StringLit {
                                                                                                    base: BaseNode {
                                                                                                        location: SourceLocation {
                                                                                                            start: "line: 1, column: 96",
                                                                                                            end: "line: 1, column: 103",
                                                                                                            source: "\"these\"",
                                                                                                        },
                                                                                                    },
                                                                                                    value: "these",
                                                                                                },
                                                                                            ),
                                                                                            rbrack: [],
                                                                                        },
                                                                                    ),
                                                                                    right: StringLit(
                                                                                        StringLit {
                                                                                            base: BaseNode {
                                                                                                location: SourceLocation {
                                                                                                    start: "line: 1, column: 106",
                                                                                                    end: "line: 1, column: 113",
                                                                                                    source: "\"those\"",
                                                                                                },
                                                                                            },
                                                                                            value: "those",
                                                                                        },
                                                                                    ),
                                                                                },
                                                                            ),
                                                                        },
                                                                    ),
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
                        ),
                    },
                ),
            ],
            eof: [],
        }
    "#]].assert_debug_eq(&parsed);
}

#[test]
fn from_with_range() {
    let mut p = Parser::new(r#"from(bucket:"telegraf/autogen")|>range(start:-1h, end:10m)"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 59",
                    source: "from(bucket:\"telegraf/autogen\")|>range(start:-1h, end:10m)",
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
                                end: "line: 1, column: 59",
                                source: "from(bucket:\"telegraf/autogen\")|>range(start:-1h, end:10m)",
                            },
                        },
                        expression: PipeExpr(
                            PipeExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 59",
                                        source: "from(bucket:\"telegraf/autogen\")|>range(start:-1h, end:10m)",
                                    },
                                },
                                argument: Call(
                                    CallExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 1",
                                                end: "line: 1, column: 32",
                                                source: "from(bucket:\"telegraf/autogen\")",
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
                                        arguments: [
                                            Object(
                                                ObjectExpr {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 6",
                                                            end: "line: 1, column: 31",
                                                            source: "bucket:\"telegraf/autogen\"",
                                                        },
                                                    },
                                                    lbrace: [],
                                                    with: None,
                                                    properties: [
                                                        Property {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 1, column: 6",
                                                                    end: "line: 1, column: 31",
                                                                    source: "bucket:\"telegraf/autogen\"",
                                                                },
                                                            },
                                                            key: Identifier(
                                                                Identifier {
                                                                    base: BaseNode {
                                                                        location: SourceLocation {
                                                                            start: "line: 1, column: 6",
                                                                            end: "line: 1, column: 12",
                                                                            source: "bucket",
                                                                        },
                                                                    },
                                                                    name: "bucket",
                                                                },
                                                            ),
                                                            separator: [],
                                                            value: Some(
                                                                StringLit(
                                                                    StringLit {
                                                                        base: BaseNode {
                                                                            location: SourceLocation {
                                                                                start: "line: 1, column: 13",
                                                                                end: "line: 1, column: 31",
                                                                                source: "\"telegraf/autogen\"",
                                                                            },
                                                                        },
                                                                        value: "telegraf/autogen",
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
                                ),
                                call: CallExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 34",
                                            end: "line: 1, column: 59",
                                            source: "range(start:-1h, end:10m)",
                                        },
                                    },
                                    callee: Identifier(
                                        Identifier {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 34",
                                                    end: "line: 1, column: 39",
                                                    source: "range",
                                                },
                                            },
                                            name: "range",
                                        },
                                    ),
                                    lparen: [],
                                    arguments: [
                                        Object(
                                            ObjectExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 40",
                                                        end: "line: 1, column: 58",
                                                        source: "start:-1h, end:10m",
                                                    },
                                                },
                                                lbrace: [],
                                                with: None,
                                                properties: [
                                                    Property {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 1, column: 40",
                                                                end: "line: 1, column: 49",
                                                                source: "start:-1h",
                                                            },
                                                        },
                                                        key: Identifier(
                                                            Identifier {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 1, column: 40",
                                                                        end: "line: 1, column: 45",
                                                                        source: "start",
                                                                    },
                                                                },
                                                                name: "start",
                                                            },
                                                        ),
                                                        separator: [],
                                                        value: Some(
                                                            Unary(
                                                                UnaryExpr {
                                                                    base: BaseNode {
                                                                        location: SourceLocation {
                                                                            start: "line: 1, column: 46",
                                                                            end: "line: 1, column: 49",
                                                                            source: "-1h",
                                                                        },
                                                                    },
                                                                    operator: SubtractionOperator,
                                                                    argument: Duration(
                                                                        DurationLit {
                                                                            base: BaseNode {
                                                                                location: SourceLocation {
                                                                                    start: "line: 1, column: 47",
                                                                                    end: "line: 1, column: 49",
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
                                                                },
                                                            ),
                                                        ),
                                                        comma: [],
                                                    },
                                                    Property {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 1, column: 51",
                                                                end: "line: 1, column: 58",
                                                                source: "end:10m",
                                                            },
                                                        },
                                                        key: Identifier(
                                                            Identifier {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 1, column: 51",
                                                                        end: "line: 1, column: 54",
                                                                        source: "end",
                                                                    },
                                                                },
                                                                name: "end",
                                                            },
                                                        ),
                                                        separator: [],
                                                        value: Some(
                                                            Duration(
                                                                DurationLit {
                                                                    base: BaseNode {
                                                                        location: SourceLocation {
                                                                            start: "line: 1, column: 55",
                                                                            end: "line: 1, column: 58",
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

#[test]
fn from_with_limit() {
    let mut p = Parser::new(r#"from(bucket:"telegraf/autogen")|>limit(limit:100, offset:10)"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 61",
                    source: "from(bucket:\"telegraf/autogen\")|>limit(limit:100, offset:10)",
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
                                end: "line: 1, column: 61",
                                source: "from(bucket:\"telegraf/autogen\")|>limit(limit:100, offset:10)",
                            },
                        },
                        expression: PipeExpr(
                            PipeExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 61",
                                        source: "from(bucket:\"telegraf/autogen\")|>limit(limit:100, offset:10)",
                                    },
                                },
                                argument: Call(
                                    CallExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 1",
                                                end: "line: 1, column: 32",
                                                source: "from(bucket:\"telegraf/autogen\")",
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
                                        arguments: [
                                            Object(
                                                ObjectExpr {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 6",
                                                            end: "line: 1, column: 31",
                                                            source: "bucket:\"telegraf/autogen\"",
                                                        },
                                                    },
                                                    lbrace: [],
                                                    with: None,
                                                    properties: [
                                                        Property {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 1, column: 6",
                                                                    end: "line: 1, column: 31",
                                                                    source: "bucket:\"telegraf/autogen\"",
                                                                },
                                                            },
                                                            key: Identifier(
                                                                Identifier {
                                                                    base: BaseNode {
                                                                        location: SourceLocation {
                                                                            start: "line: 1, column: 6",
                                                                            end: "line: 1, column: 12",
                                                                            source: "bucket",
                                                                        },
                                                                    },
                                                                    name: "bucket",
                                                                },
                                                            ),
                                                            separator: [],
                                                            value: Some(
                                                                StringLit(
                                                                    StringLit {
                                                                        base: BaseNode {
                                                                            location: SourceLocation {
                                                                                start: "line: 1, column: 13",
                                                                                end: "line: 1, column: 31",
                                                                                source: "\"telegraf/autogen\"",
                                                                            },
                                                                        },
                                                                        value: "telegraf/autogen",
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
                                ),
                                call: CallExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 34",
                                            end: "line: 1, column: 61",
                                            source: "limit(limit:100, offset:10)",
                                        },
                                    },
                                    callee: Identifier(
                                        Identifier {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 34",
                                                    end: "line: 1, column: 39",
                                                    source: "limit",
                                                },
                                            },
                                            name: "limit",
                                        },
                                    ),
                                    lparen: [],
                                    arguments: [
                                        Object(
                                            ObjectExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 40",
                                                        end: "line: 1, column: 60",
                                                        source: "limit:100, offset:10",
                                                    },
                                                },
                                                lbrace: [],
                                                with: None,
                                                properties: [
                                                    Property {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 1, column: 40",
                                                                end: "line: 1, column: 49",
                                                                source: "limit:100",
                                                            },
                                                        },
                                                        key: Identifier(
                                                            Identifier {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 1, column: 40",
                                                                        end: "line: 1, column: 45",
                                                                        source: "limit",
                                                                    },
                                                                },
                                                                name: "limit",
                                                            },
                                                        ),
                                                        separator: [],
                                                        value: Some(
                                                            Integer(
                                                                IntegerLit {
                                                                    base: BaseNode {
                                                                        location: SourceLocation {
                                                                            start: "line: 1, column: 46",
                                                                            end: "line: 1, column: 49",
                                                                            source: "100",
                                                                        },
                                                                    },
                                                                    value: 100,
                                                                },
                                                            ),
                                                        ),
                                                        comma: [],
                                                    },
                                                    Property {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 1, column: 51",
                                                                end: "line: 1, column: 60",
                                                                source: "offset:10",
                                                            },
                                                        },
                                                        key: Identifier(
                                                            Identifier {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 1, column: 51",
                                                                        end: "line: 1, column: 57",
                                                                        source: "offset",
                                                                    },
                                                                },
                                                                name: "offset",
                                                            },
                                                        ),
                                                        separator: [],
                                                        value: Some(
                                                            Integer(
                                                                IntegerLit {
                                                                    base: BaseNode {
                                                                        location: SourceLocation {
                                                                            start: "line: 1, column: 58",
                                                                            end: "line: 1, column: 60",
                                                                            source: "10",
                                                                        },
                                                                    },
                                                                    value: 10,
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

#[test]
fn from_with_range_and_count() {
    let mut p = Parser::new(
        r#"from(bucket:"mydb/autogen")
						|> range(start:-4h, stop:-2h)
						|> count()"#,
    );
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 3, column: 17",
                    source: "from(bucket:\"mydb/autogen\")\n\t\t\t\t\t\t|> range(start:-4h, stop:-2h)\n\t\t\t\t\t\t|> count()",
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
                                end: "line: 3, column: 17",
                                source: "from(bucket:\"mydb/autogen\")\n\t\t\t\t\t\t|> range(start:-4h, stop:-2h)\n\t\t\t\t\t\t|> count()",
                            },
                        },
                        expression: PipeExpr(
                            PipeExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 3, column: 17",
                                        source: "from(bucket:\"mydb/autogen\")\n\t\t\t\t\t\t|> range(start:-4h, stop:-2h)\n\t\t\t\t\t\t|> count()",
                                    },
                                },
                                argument: PipeExpr(
                                    PipeExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 1",
                                                end: "line: 2, column: 36",
                                                source: "from(bucket:\"mydb/autogen\")\n\t\t\t\t\t\t|> range(start:-4h, stop:-2h)",
                                            },
                                        },
                                        argument: Call(
                                            CallExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 1",
                                                        end: "line: 1, column: 28",
                                                        source: "from(bucket:\"mydb/autogen\")",
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
                                                arguments: [
                                                    Object(
                                                        ObjectExpr {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 1, column: 6",
                                                                    end: "line: 1, column: 27",
                                                                    source: "bucket:\"mydb/autogen\"",
                                                                },
                                                            },
                                                            lbrace: [],
                                                            with: None,
                                                            properties: [
                                                                Property {
                                                                    base: BaseNode {
                                                                        location: SourceLocation {
                                                                            start: "line: 1, column: 6",
                                                                            end: "line: 1, column: 27",
                                                                            source: "bucket:\"mydb/autogen\"",
                                                                        },
                                                                    },
                                                                    key: Identifier(
                                                                        Identifier {
                                                                            base: BaseNode {
                                                                                location: SourceLocation {
                                                                                    start: "line: 1, column: 6",
                                                                                    end: "line: 1, column: 12",
                                                                                    source: "bucket",
                                                                                },
                                                                            },
                                                                            name: "bucket",
                                                                        },
                                                                    ),
                                                                    separator: [],
                                                                    value: Some(
                                                                        StringLit(
                                                                            StringLit {
                                                                                base: BaseNode {
                                                                                    location: SourceLocation {
                                                                                        start: "line: 1, column: 13",
                                                                                        end: "line: 1, column: 27",
                                                                                        source: "\"mydb/autogen\"",
                                                                                    },
                                                                                },
                                                                                value: "mydb/autogen",
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
                                        ),
                                        call: CallExpr {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 2, column: 10",
                                                    end: "line: 2, column: 36",
                                                    source: "range(start:-4h, stop:-2h)",
                                                },
                                            },
                                            callee: Identifier(
                                                Identifier {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 2, column: 10",
                                                            end: "line: 2, column: 15",
                                                            source: "range",
                                                        },
                                                    },
                                                    name: "range",
                                                },
                                            ),
                                            lparen: [],
                                            arguments: [
                                                Object(
                                                    ObjectExpr {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 2, column: 16",
                                                                end: "line: 2, column: 35",
                                                                source: "start:-4h, stop:-2h",
                                                            },
                                                        },
                                                        lbrace: [],
                                                        with: None,
                                                        properties: [
                                                            Property {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 2, column: 16",
                                                                        end: "line: 2, column: 25",
                                                                        source: "start:-4h",
                                                                    },
                                                                },
                                                                key: Identifier(
                                                                    Identifier {
                                                                        base: BaseNode {
                                                                            location: SourceLocation {
                                                                                start: "line: 2, column: 16",
                                                                                end: "line: 2, column: 21",
                                                                                source: "start",
                                                                            },
                                                                        },
                                                                        name: "start",
                                                                    },
                                                                ),
                                                                separator: [],
                                                                value: Some(
                                                                    Unary(
                                                                        UnaryExpr {
                                                                            base: BaseNode {
                                                                                location: SourceLocation {
                                                                                    start: "line: 2, column: 22",
                                                                                    end: "line: 2, column: 25",
                                                                                    source: "-4h",
                                                                                },
                                                                            },
                                                                            operator: SubtractionOperator,
                                                                            argument: Duration(
                                                                                DurationLit {
                                                                                    base: BaseNode {
                                                                                        location: SourceLocation {
                                                                                            start: "line: 2, column: 23",
                                                                                            end: "line: 2, column: 25",
                                                                                            source: "4h",
                                                                                        },
                                                                                    },
                                                                                    values: [
                                                                                        Duration {
                                                                                            magnitude: 4,
                                                                                            unit: "h",
                                                                                        },
                                                                                    ],
                                                                                },
                                                                            ),
                                                                        },
                                                                    ),
                                                                ),
                                                                comma: [],
                                                            },
                                                            Property {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 2, column: 27",
                                                                        end: "line: 2, column: 35",
                                                                        source: "stop:-2h",
                                                                    },
                                                                },
                                                                key: Identifier(
                                                                    Identifier {
                                                                        base: BaseNode {
                                                                            location: SourceLocation {
                                                                                start: "line: 2, column: 27",
                                                                                end: "line: 2, column: 31",
                                                                                source: "stop",
                                                                            },
                                                                        },
                                                                        name: "stop",
                                                                    },
                                                                ),
                                                                separator: [],
                                                                value: Some(
                                                                    Unary(
                                                                        UnaryExpr {
                                                                            base: BaseNode {
                                                                                location: SourceLocation {
                                                                                    start: "line: 2, column: 32",
                                                                                    end: "line: 2, column: 35",
                                                                                    source: "-2h",
                                                                                },
                                                                            },
                                                                            operator: SubtractionOperator,
                                                                            argument: Duration(
                                                                                DurationLit {
                                                                                    base: BaseNode {
                                                                                        location: SourceLocation {
                                                                                            start: "line: 2, column: 33",
                                                                                            end: "line: 2, column: 35",
                                                                                            source: "2h",
                                                                                        },
                                                                                    },
                                                                                    values: [
                                                                                        Duration {
                                                                                            magnitude: 2,
                                                                                            unit: "h",
                                                                                        },
                                                                                    ],
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
                                call: CallExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 3, column: 10",
                                            end: "line: 3, column: 17",
                                            source: "count()",
                                        },
                                    },
                                    callee: Identifier(
                                        Identifier {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 3, column: 10",
                                                    end: "line: 3, column: 15",
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
    "#]].assert_debug_eq(&parsed);
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
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 4, column: 17",
                    source: "from(bucket:\"mydb/autogen\")\n\t\t\t\t\t\t|> range(start:-4h, stop:-2h)\n\t\t\t\t\t\t|> limit(n:10)\n\t\t\t\t\t\t|> count()",
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
                                end: "line: 4, column: 17",
                                source: "from(bucket:\"mydb/autogen\")\n\t\t\t\t\t\t|> range(start:-4h, stop:-2h)\n\t\t\t\t\t\t|> limit(n:10)\n\t\t\t\t\t\t|> count()",
                            },
                        },
                        expression: PipeExpr(
                            PipeExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 4, column: 17",
                                        source: "from(bucket:\"mydb/autogen\")\n\t\t\t\t\t\t|> range(start:-4h, stop:-2h)\n\t\t\t\t\t\t|> limit(n:10)\n\t\t\t\t\t\t|> count()",
                                    },
                                },
                                argument: PipeExpr(
                                    PipeExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 1",
                                                end: "line: 3, column: 21",
                                                source: "from(bucket:\"mydb/autogen\")\n\t\t\t\t\t\t|> range(start:-4h, stop:-2h)\n\t\t\t\t\t\t|> limit(n:10)",
                                            },
                                        },
                                        argument: PipeExpr(
                                            PipeExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 1",
                                                        end: "line: 2, column: 36",
                                                        source: "from(bucket:\"mydb/autogen\")\n\t\t\t\t\t\t|> range(start:-4h, stop:-2h)",
                                                    },
                                                },
                                                argument: Call(
                                                    CallExpr {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 1, column: 1",
                                                                end: "line: 1, column: 28",
                                                                source: "from(bucket:\"mydb/autogen\")",
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
                                                        arguments: [
                                                            Object(
                                                                ObjectExpr {
                                                                    base: BaseNode {
                                                                        location: SourceLocation {
                                                                            start: "line: 1, column: 6",
                                                                            end: "line: 1, column: 27",
                                                                            source: "bucket:\"mydb/autogen\"",
                                                                        },
                                                                    },
                                                                    lbrace: [],
                                                                    with: None,
                                                                    properties: [
                                                                        Property {
                                                                            base: BaseNode {
                                                                                location: SourceLocation {
                                                                                    start: "line: 1, column: 6",
                                                                                    end: "line: 1, column: 27",
                                                                                    source: "bucket:\"mydb/autogen\"",
                                                                                },
                                                                            },
                                                                            key: Identifier(
                                                                                Identifier {
                                                                                    base: BaseNode {
                                                                                        location: SourceLocation {
                                                                                            start: "line: 1, column: 6",
                                                                                            end: "line: 1, column: 12",
                                                                                            source: "bucket",
                                                                                        },
                                                                                    },
                                                                                    name: "bucket",
                                                                                },
                                                                            ),
                                                                            separator: [],
                                                                            value: Some(
                                                                                StringLit(
                                                                                    StringLit {
                                                                                        base: BaseNode {
                                                                                            location: SourceLocation {
                                                                                                start: "line: 1, column: 13",
                                                                                                end: "line: 1, column: 27",
                                                                                                source: "\"mydb/autogen\"",
                                                                                            },
                                                                                        },
                                                                                        value: "mydb/autogen",
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
                                                ),
                                                call: CallExpr {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 2, column: 10",
                                                            end: "line: 2, column: 36",
                                                            source: "range(start:-4h, stop:-2h)",
                                                        },
                                                    },
                                                    callee: Identifier(
                                                        Identifier {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 2, column: 10",
                                                                    end: "line: 2, column: 15",
                                                                    source: "range",
                                                                },
                                                            },
                                                            name: "range",
                                                        },
                                                    ),
                                                    lparen: [],
                                                    arguments: [
                                                        Object(
                                                            ObjectExpr {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 2, column: 16",
                                                                        end: "line: 2, column: 35",
                                                                        source: "start:-4h, stop:-2h",
                                                                    },
                                                                },
                                                                lbrace: [],
                                                                with: None,
                                                                properties: [
                                                                    Property {
                                                                        base: BaseNode {
                                                                            location: SourceLocation {
                                                                                start: "line: 2, column: 16",
                                                                                end: "line: 2, column: 25",
                                                                                source: "start:-4h",
                                                                            },
                                                                        },
                                                                        key: Identifier(
                                                                            Identifier {
                                                                                base: BaseNode {
                                                                                    location: SourceLocation {
                                                                                        start: "line: 2, column: 16",
                                                                                        end: "line: 2, column: 21",
                                                                                        source: "start",
                                                                                    },
                                                                                },
                                                                                name: "start",
                                                                            },
                                                                        ),
                                                                        separator: [],
                                                                        value: Some(
                                                                            Unary(
                                                                                UnaryExpr {
                                                                                    base: BaseNode {
                                                                                        location: SourceLocation {
                                                                                            start: "line: 2, column: 22",
                                                                                            end: "line: 2, column: 25",
                                                                                            source: "-4h",
                                                                                        },
                                                                                    },
                                                                                    operator: SubtractionOperator,
                                                                                    argument: Duration(
                                                                                        DurationLit {
                                                                                            base: BaseNode {
                                                                                                location: SourceLocation {
                                                                                                    start: "line: 2, column: 23",
                                                                                                    end: "line: 2, column: 25",
                                                                                                    source: "4h",
                                                                                                },
                                                                                            },
                                                                                            values: [
                                                                                                Duration {
                                                                                                    magnitude: 4,
                                                                                                    unit: "h",
                                                                                                },
                                                                                            ],
                                                                                        },
                                                                                    ),
                                                                                },
                                                                            ),
                                                                        ),
                                                                        comma: [],
                                                                    },
                                                                    Property {
                                                                        base: BaseNode {
                                                                            location: SourceLocation {
                                                                                start: "line: 2, column: 27",
                                                                                end: "line: 2, column: 35",
                                                                                source: "stop:-2h",
                                                                            },
                                                                        },
                                                                        key: Identifier(
                                                                            Identifier {
                                                                                base: BaseNode {
                                                                                    location: SourceLocation {
                                                                                        start: "line: 2, column: 27",
                                                                                        end: "line: 2, column: 31",
                                                                                        source: "stop",
                                                                                    },
                                                                                },
                                                                                name: "stop",
                                                                            },
                                                                        ),
                                                                        separator: [],
                                                                        value: Some(
                                                                            Unary(
                                                                                UnaryExpr {
                                                                                    base: BaseNode {
                                                                                        location: SourceLocation {
                                                                                            start: "line: 2, column: 32",
                                                                                            end: "line: 2, column: 35",
                                                                                            source: "-2h",
                                                                                        },
                                                                                    },
                                                                                    operator: SubtractionOperator,
                                                                                    argument: Duration(
                                                                                        DurationLit {
                                                                                            base: BaseNode {
                                                                                                location: SourceLocation {
                                                                                                    start: "line: 2, column: 33",
                                                                                                    end: "line: 2, column: 35",
                                                                                                    source: "2h",
                                                                                                },
                                                                                            },
                                                                                            values: [
                                                                                                Duration {
                                                                                                    magnitude: 2,
                                                                                                    unit: "h",
                                                                                                },
                                                                                            ],
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
                                        call: CallExpr {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 3, column: 10",
                                                    end: "line: 3, column: 21",
                                                    source: "limit(n:10)",
                                                },
                                            },
                                            callee: Identifier(
                                                Identifier {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 3, column: 10",
                                                            end: "line: 3, column: 15",
                                                            source: "limit",
                                                        },
                                                    },
                                                    name: "limit",
                                                },
                                            ),
                                            lparen: [],
                                            arguments: [
                                                Object(
                                                    ObjectExpr {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 3, column: 16",
                                                                end: "line: 3, column: 20",
                                                                source: "n:10",
                                                            },
                                                        },
                                                        lbrace: [],
                                                        with: None,
                                                        properties: [
                                                            Property {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 3, column: 16",
                                                                        end: "line: 3, column: 20",
                                                                        source: "n:10",
                                                                    },
                                                                },
                                                                key: Identifier(
                                                                    Identifier {
                                                                        base: BaseNode {
                                                                            location: SourceLocation {
                                                                                start: "line: 3, column: 16",
                                                                                end: "line: 3, column: 17",
                                                                                source: "n",
                                                                            },
                                                                        },
                                                                        name: "n",
                                                                    },
                                                                ),
                                                                separator: [],
                                                                value: Some(
                                                                    Integer(
                                                                        IntegerLit {
                                                                            base: BaseNode {
                                                                                location: SourceLocation {
                                                                                    start: "line: 3, column: 18",
                                                                                    end: "line: 3, column: 20",
                                                                                    source: "10",
                                                                                },
                                                                            },
                                                                            value: 10,
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
                                call: CallExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 4, column: 10",
                                            end: "line: 4, column: 17",
                                            source: "count()",
                                        },
                                    },
                                    callee: Identifier(
                                        Identifier {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 4, column: 10",
                                                    end: "line: 4, column: 15",
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
    "#]].assert_debug_eq(&parsed);
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
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 4, column: 72",
                    source: "\na = from(bucket:\"dbA/autogen\") |> range(start:-1h)\nb = from(bucket:\"dbB/autogen\") |> range(start:-1h)\njoin(tables:[a,b], on:[\"host\"], fn: (a,b) => a[\"_field\"] + b[\"_field\"])",
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
                                end: "line: 2, column: 51",
                                source: "a = from(bucket:\"dbA/autogen\") |> range(start:-1h)",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 2, column: 1",
                                    end: "line: 2, column: 2",
                                    source: "a",
                                },
                            },
                            name: "a",
                        },
                        init: PipeExpr(
                            PipeExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 2, column: 5",
                                        end: "line: 2, column: 51",
                                        source: "from(bucket:\"dbA/autogen\") |> range(start:-1h)",
                                    },
                                },
                                argument: Call(
                                    CallExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 2, column: 5",
                                                end: "line: 2, column: 31",
                                                source: "from(bucket:\"dbA/autogen\")",
                                            },
                                        },
                                        callee: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 2, column: 5",
                                                        end: "line: 2, column: 9",
                                                        source: "from",
                                                    },
                                                },
                                                name: "from",
                                            },
                                        ),
                                        lparen: [],
                                        arguments: [
                                            Object(
                                                ObjectExpr {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 2, column: 10",
                                                            end: "line: 2, column: 30",
                                                            source: "bucket:\"dbA/autogen\"",
                                                        },
                                                    },
                                                    lbrace: [],
                                                    with: None,
                                                    properties: [
                                                        Property {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 2, column: 10",
                                                                    end: "line: 2, column: 30",
                                                                    source: "bucket:\"dbA/autogen\"",
                                                                },
                                                            },
                                                            key: Identifier(
                                                                Identifier {
                                                                    base: BaseNode {
                                                                        location: SourceLocation {
                                                                            start: "line: 2, column: 10",
                                                                            end: "line: 2, column: 16",
                                                                            source: "bucket",
                                                                        },
                                                                    },
                                                                    name: "bucket",
                                                                },
                                                            ),
                                                            separator: [],
                                                            value: Some(
                                                                StringLit(
                                                                    StringLit {
                                                                        base: BaseNode {
                                                                            location: SourceLocation {
                                                                                start: "line: 2, column: 17",
                                                                                end: "line: 2, column: 30",
                                                                                source: "\"dbA/autogen\"",
                                                                            },
                                                                        },
                                                                        value: "dbA/autogen",
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
                                ),
                                call: CallExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 2, column: 35",
                                            end: "line: 2, column: 51",
                                            source: "range(start:-1h)",
                                        },
                                    },
                                    callee: Identifier(
                                        Identifier {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 2, column: 35",
                                                    end: "line: 2, column: 40",
                                                    source: "range",
                                                },
                                            },
                                            name: "range",
                                        },
                                    ),
                                    lparen: [],
                                    arguments: [
                                        Object(
                                            ObjectExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 2, column: 41",
                                                        end: "line: 2, column: 50",
                                                        source: "start:-1h",
                                                    },
                                                },
                                                lbrace: [],
                                                with: None,
                                                properties: [
                                                    Property {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 2, column: 41",
                                                                end: "line: 2, column: 50",
                                                                source: "start:-1h",
                                                            },
                                                        },
                                                        key: Identifier(
                                                            Identifier {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 2, column: 41",
                                                                        end: "line: 2, column: 46",
                                                                        source: "start",
                                                                    },
                                                                },
                                                                name: "start",
                                                            },
                                                        ),
                                                        separator: [],
                                                        value: Some(
                                                            Unary(
                                                                UnaryExpr {
                                                                    base: BaseNode {
                                                                        location: SourceLocation {
                                                                            start: "line: 2, column: 47",
                                                                            end: "line: 2, column: 50",
                                                                            source: "-1h",
                                                                        },
                                                                    },
                                                                    operator: SubtractionOperator,
                                                                    argument: Duration(
                                                                        DurationLit {
                                                                            base: BaseNode {
                                                                                location: SourceLocation {
                                                                                    start: "line: 2, column: 48",
                                                                                    end: "line: 2, column: 50",
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
                Variable(
                    VariableAssgn {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 3, column: 1",
                                end: "line: 3, column: 51",
                                source: "b = from(bucket:\"dbB/autogen\") |> range(start:-1h)",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 3, column: 1",
                                    end: "line: 3, column: 2",
                                    source: "b",
                                },
                            },
                            name: "b",
                        },
                        init: PipeExpr(
                            PipeExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 3, column: 5",
                                        end: "line: 3, column: 51",
                                        source: "from(bucket:\"dbB/autogen\") |> range(start:-1h)",
                                    },
                                },
                                argument: Call(
                                    CallExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 3, column: 5",
                                                end: "line: 3, column: 31",
                                                source: "from(bucket:\"dbB/autogen\")",
                                            },
                                        },
                                        callee: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 3, column: 5",
                                                        end: "line: 3, column: 9",
                                                        source: "from",
                                                    },
                                                },
                                                name: "from",
                                            },
                                        ),
                                        lparen: [],
                                        arguments: [
                                            Object(
                                                ObjectExpr {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 3, column: 10",
                                                            end: "line: 3, column: 30",
                                                            source: "bucket:\"dbB/autogen\"",
                                                        },
                                                    },
                                                    lbrace: [],
                                                    with: None,
                                                    properties: [
                                                        Property {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 3, column: 10",
                                                                    end: "line: 3, column: 30",
                                                                    source: "bucket:\"dbB/autogen\"",
                                                                },
                                                            },
                                                            key: Identifier(
                                                                Identifier {
                                                                    base: BaseNode {
                                                                        location: SourceLocation {
                                                                            start: "line: 3, column: 10",
                                                                            end: "line: 3, column: 16",
                                                                            source: "bucket",
                                                                        },
                                                                    },
                                                                    name: "bucket",
                                                                },
                                                            ),
                                                            separator: [],
                                                            value: Some(
                                                                StringLit(
                                                                    StringLit {
                                                                        base: BaseNode {
                                                                            location: SourceLocation {
                                                                                start: "line: 3, column: 17",
                                                                                end: "line: 3, column: 30",
                                                                                source: "\"dbB/autogen\"",
                                                                            },
                                                                        },
                                                                        value: "dbB/autogen",
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
                                ),
                                call: CallExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 3, column: 35",
                                            end: "line: 3, column: 51",
                                            source: "range(start:-1h)",
                                        },
                                    },
                                    callee: Identifier(
                                        Identifier {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 3, column: 35",
                                                    end: "line: 3, column: 40",
                                                    source: "range",
                                                },
                                            },
                                            name: "range",
                                        },
                                    ),
                                    lparen: [],
                                    arguments: [
                                        Object(
                                            ObjectExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 3, column: 41",
                                                        end: "line: 3, column: 50",
                                                        source: "start:-1h",
                                                    },
                                                },
                                                lbrace: [],
                                                with: None,
                                                properties: [
                                                    Property {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 3, column: 41",
                                                                end: "line: 3, column: 50",
                                                                source: "start:-1h",
                                                            },
                                                        },
                                                        key: Identifier(
                                                            Identifier {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 3, column: 41",
                                                                        end: "line: 3, column: 46",
                                                                        source: "start",
                                                                    },
                                                                },
                                                                name: "start",
                                                            },
                                                        ),
                                                        separator: [],
                                                        value: Some(
                                                            Unary(
                                                                UnaryExpr {
                                                                    base: BaseNode {
                                                                        location: SourceLocation {
                                                                            start: "line: 3, column: 47",
                                                                            end: "line: 3, column: 50",
                                                                            source: "-1h",
                                                                        },
                                                                    },
                                                                    operator: SubtractionOperator,
                                                                    argument: Duration(
                                                                        DurationLit {
                                                                            base: BaseNode {
                                                                                location: SourceLocation {
                                                                                    start: "line: 3, column: 48",
                                                                                    end: "line: 3, column: 50",
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
                Expr(
                    ExprStmt {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 4, column: 1",
                                end: "line: 4, column: 72",
                                source: "join(tables:[a,b], on:[\"host\"], fn: (a,b) => a[\"_field\"] + b[\"_field\"])",
                            },
                        },
                        expression: Call(
                            CallExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 4, column: 1",
                                        end: "line: 4, column: 72",
                                        source: "join(tables:[a,b], on:[\"host\"], fn: (a,b) => a[\"_field\"] + b[\"_field\"])",
                                    },
                                },
                                callee: Identifier(
                                    Identifier {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 4, column: 1",
                                                end: "line: 4, column: 5",
                                                source: "join",
                                            },
                                        },
                                        name: "join",
                                    },
                                ),
                                lparen: [],
                                arguments: [
                                    Object(
                                        ObjectExpr {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 4, column: 6",
                                                    end: "line: 4, column: 71",
                                                    source: "tables:[a,b], on:[\"host\"], fn: (a,b) => a[\"_field\"] + b[\"_field\"]",
                                                },
                                            },
                                            lbrace: [],
                                            with: None,
                                            properties: [
                                                Property {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 4, column: 6",
                                                            end: "line: 4, column: 18",
                                                            source: "tables:[a,b]",
                                                        },
                                                    },
                                                    key: Identifier(
                                                        Identifier {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 4, column: 6",
                                                                    end: "line: 4, column: 12",
                                                                    source: "tables",
                                                                },
                                                            },
                                                            name: "tables",
                                                        },
                                                    ),
                                                    separator: [],
                                                    value: Some(
                                                        Array(
                                                            ArrayExpr {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 4, column: 13",
                                                                        end: "line: 4, column: 18",
                                                                        source: "[a,b]",
                                                                    },
                                                                },
                                                                lbrack: [],
                                                                elements: [
                                                                    ArrayItem {
                                                                        expression: Identifier(
                                                                            Identifier {
                                                                                base: BaseNode {
                                                                                    location: SourceLocation {
                                                                                        start: "line: 4, column: 14",
                                                                                        end: "line: 4, column: 15",
                                                                                        source: "a",
                                                                                    },
                                                                                },
                                                                                name: "a",
                                                                            },
                                                                        ),
                                                                        comma: [],
                                                                    },
                                                                    ArrayItem {
                                                                        expression: Identifier(
                                                                            Identifier {
                                                                                base: BaseNode {
                                                                                    location: SourceLocation {
                                                                                        start: "line: 4, column: 16",
                                                                                        end: "line: 4, column: 17",
                                                                                        source: "b",
                                                                                    },
                                                                                },
                                                                                name: "b",
                                                                            },
                                                                        ),
                                                                        comma: [],
                                                                    },
                                                                ],
                                                                rbrack: [],
                                                            },
                                                        ),
                                                    ),
                                                    comma: [],
                                                },
                                                Property {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 4, column: 20",
                                                            end: "line: 4, column: 31",
                                                            source: "on:[\"host\"]",
                                                        },
                                                    },
                                                    key: Identifier(
                                                        Identifier {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 4, column: 20",
                                                                    end: "line: 4, column: 22",
                                                                    source: "on",
                                                                },
                                                            },
                                                            name: "on",
                                                        },
                                                    ),
                                                    separator: [],
                                                    value: Some(
                                                        Array(
                                                            ArrayExpr {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 4, column: 23",
                                                                        end: "line: 4, column: 31",
                                                                        source: "[\"host\"]",
                                                                    },
                                                                },
                                                                lbrack: [],
                                                                elements: [
                                                                    ArrayItem {
                                                                        expression: StringLit(
                                                                            StringLit {
                                                                                base: BaseNode {
                                                                                    location: SourceLocation {
                                                                                        start: "line: 4, column: 24",
                                                                                        end: "line: 4, column: 30",
                                                                                        source: "\"host\"",
                                                                                    },
                                                                                },
                                                                                value: "host",
                                                                            },
                                                                        ),
                                                                        comma: [],
                                                                    },
                                                                ],
                                                                rbrack: [],
                                                            },
                                                        ),
                                                    ),
                                                    comma: [],
                                                },
                                                Property {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 4, column: 33",
                                                            end: "line: 4, column: 71",
                                                            source: "fn: (a,b) => a[\"_field\"] + b[\"_field\"]",
                                                        },
                                                    },
                                                    key: Identifier(
                                                        Identifier {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 4, column: 33",
                                                                    end: "line: 4, column: 35",
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
                                                                        start: "line: 4, column: 37",
                                                                        end: "line: 4, column: 71",
                                                                        source: "(a,b) => a[\"_field\"] + b[\"_field\"]",
                                                                    },
                                                                },
                                                                lparen: [],
                                                                params: [
                                                                    Property {
                                                                        base: BaseNode {
                                                                            location: SourceLocation {
                                                                                start: "line: 4, column: 38",
                                                                                end: "line: 4, column: 39",
                                                                                source: "a",
                                                                            },
                                                                        },
                                                                        key: Identifier(
                                                                            Identifier {
                                                                                base: BaseNode {
                                                                                    location: SourceLocation {
                                                                                        start: "line: 4, column: 38",
                                                                                        end: "line: 4, column: 39",
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
                                                                                start: "line: 4, column: 40",
                                                                                end: "line: 4, column: 41",
                                                                                source: "b",
                                                                            },
                                                                        },
                                                                        key: Identifier(
                                                                            Identifier {
                                                                                base: BaseNode {
                                                                                    location: SourceLocation {
                                                                                        start: "line: 4, column: 40",
                                                                                        end: "line: 4, column: 41",
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
                                                                                    start: "line: 4, column: 46",
                                                                                    end: "line: 4, column: 71",
                                                                                    source: "a[\"_field\"] + b[\"_field\"]",
                                                                                },
                                                                            },
                                                                            operator: AdditionOperator,
                                                                            left: Member(
                                                                                MemberExpr {
                                                                                    base: BaseNode {
                                                                                        location: SourceLocation {
                                                                                            start: "line: 4, column: 46",
                                                                                            end: "line: 4, column: 57",
                                                                                            source: "a[\"_field\"]",
                                                                                        },
                                                                                    },
                                                                                    object: Identifier(
                                                                                        Identifier {
                                                                                            base: BaseNode {
                                                                                                location: SourceLocation {
                                                                                                    start: "line: 4, column: 46",
                                                                                                    end: "line: 4, column: 47",
                                                                                                    source: "a",
                                                                                                },
                                                                                            },
                                                                                            name: "a",
                                                                                        },
                                                                                    ),
                                                                                    lbrack: [],
                                                                                    property: StringLit(
                                                                                        StringLit {
                                                                                            base: BaseNode {
                                                                                                location: SourceLocation {
                                                                                                    start: "line: 4, column: 48",
                                                                                                    end: "line: 4, column: 56",
                                                                                                    source: "\"_field\"",
                                                                                                },
                                                                                            },
                                                                                            value: "_field",
                                                                                        },
                                                                                    ),
                                                                                    rbrack: [],
                                                                                },
                                                                            ),
                                                                            right: Member(
                                                                                MemberExpr {
                                                                                    base: BaseNode {
                                                                                        location: SourceLocation {
                                                                                            start: "line: 4, column: 60",
                                                                                            end: "line: 4, column: 71",
                                                                                            source: "b[\"_field\"]",
                                                                                        },
                                                                                    },
                                                                                    object: Identifier(
                                                                                        Identifier {
                                                                                            base: BaseNode {
                                                                                                location: SourceLocation {
                                                                                                    start: "line: 4, column: 60",
                                                                                                    end: "line: 4, column: 61",
                                                                                                    source: "b",
                                                                                                },
                                                                                            },
                                                                                            name: "b",
                                                                                        },
                                                                                    ),
                                                                                    lbrack: [],
                                                                                    property: StringLit(
                                                                                        StringLit {
                                                                                            base: BaseNode {
                                                                                                location: SourceLocation {
                                                                                                    start: "line: 4, column: 62",
                                                                                                    end: "line: 4, column: 70",
                                                                                                    source: "\"_field\"",
                                                                                                },
                                                                                            },
                                                                                            value: "_field",
                                                                                        },
                                                                                    ),
                                                                                    rbrack: [],
                                                                                },
                                                                            ),
                                                                        },
                                                                    ),
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
                        ),
                    },
                ),
            ],
            eof: [],
        }
    "#]].assert_debug_eq(&parsed);
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

    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 10, column: 86",
                    source: "\na = from(bucket:\"Flux/autogen\")\n\t|> filter(fn: (r) => r[\"_measurement\"] == \"a\")\n\t|> range(start:-1h)\n\nb = from(bucket:\"Flux/autogen\")\n\t|> filter(fn: (r) => r[\"_measurement\"] == \"b\")\n\t|> range(start:-1h)\n\njoin(tables:[a,b], on:[\"t1\"], fn: (a,b) => (a[\"_field\"] - b[\"_field\"]) / b[\"_field\"])",
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
                                end: "line: 4, column: 21",
                                source: "a = from(bucket:\"Flux/autogen\")\n\t|> filter(fn: (r) => r[\"_measurement\"] == \"a\")\n\t|> range(start:-1h)",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 2, column: 1",
                                    end: "line: 2, column: 2",
                                    source: "a",
                                },
                            },
                            name: "a",
                        },
                        init: PipeExpr(
                            PipeExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 2, column: 5",
                                        end: "line: 4, column: 21",
                                        source: "from(bucket:\"Flux/autogen\")\n\t|> filter(fn: (r) => r[\"_measurement\"] == \"a\")\n\t|> range(start:-1h)",
                                    },
                                },
                                argument: PipeExpr(
                                    PipeExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 2, column: 5",
                                                end: "line: 3, column: 48",
                                                source: "from(bucket:\"Flux/autogen\")\n\t|> filter(fn: (r) => r[\"_measurement\"] == \"a\")",
                                            },
                                        },
                                        argument: Call(
                                            CallExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 2, column: 5",
                                                        end: "line: 2, column: 32",
                                                        source: "from(bucket:\"Flux/autogen\")",
                                                    },
                                                },
                                                callee: Identifier(
                                                    Identifier {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 2, column: 5",
                                                                end: "line: 2, column: 9",
                                                                source: "from",
                                                            },
                                                        },
                                                        name: "from",
                                                    },
                                                ),
                                                lparen: [],
                                                arguments: [
                                                    Object(
                                                        ObjectExpr {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 2, column: 10",
                                                                    end: "line: 2, column: 31",
                                                                    source: "bucket:\"Flux/autogen\"",
                                                                },
                                                            },
                                                            lbrace: [],
                                                            with: None,
                                                            properties: [
                                                                Property {
                                                                    base: BaseNode {
                                                                        location: SourceLocation {
                                                                            start: "line: 2, column: 10",
                                                                            end: "line: 2, column: 31",
                                                                            source: "bucket:\"Flux/autogen\"",
                                                                        },
                                                                    },
                                                                    key: Identifier(
                                                                        Identifier {
                                                                            base: BaseNode {
                                                                                location: SourceLocation {
                                                                                    start: "line: 2, column: 10",
                                                                                    end: "line: 2, column: 16",
                                                                                    source: "bucket",
                                                                                },
                                                                            },
                                                                            name: "bucket",
                                                                        },
                                                                    ),
                                                                    separator: [],
                                                                    value: Some(
                                                                        StringLit(
                                                                            StringLit {
                                                                                base: BaseNode {
                                                                                    location: SourceLocation {
                                                                                        start: "line: 2, column: 17",
                                                                                        end: "line: 2, column: 31",
                                                                                        source: "\"Flux/autogen\"",
                                                                                    },
                                                                                },
                                                                                value: "Flux/autogen",
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
                                        ),
                                        call: CallExpr {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 3, column: 5",
                                                    end: "line: 3, column: 48",
                                                    source: "filter(fn: (r) => r[\"_measurement\"] == \"a\")",
                                                },
                                            },
                                            callee: Identifier(
                                                Identifier {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 3, column: 5",
                                                            end: "line: 3, column: 11",
                                                            source: "filter",
                                                        },
                                                    },
                                                    name: "filter",
                                                },
                                            ),
                                            lparen: [],
                                            arguments: [
                                                Object(
                                                    ObjectExpr {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 3, column: 12",
                                                                end: "line: 3, column: 47",
                                                                source: "fn: (r) => r[\"_measurement\"] == \"a\"",
                                                            },
                                                        },
                                                        lbrace: [],
                                                        with: None,
                                                        properties: [
                                                            Property {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 3, column: 12",
                                                                        end: "line: 3, column: 47",
                                                                        source: "fn: (r) => r[\"_measurement\"] == \"a\"",
                                                                    },
                                                                },
                                                                key: Identifier(
                                                                    Identifier {
                                                                        base: BaseNode {
                                                                            location: SourceLocation {
                                                                                start: "line: 3, column: 12",
                                                                                end: "line: 3, column: 14",
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
                                                                                    start: "line: 3, column: 16",
                                                                                    end: "line: 3, column: 47",
                                                                                    source: "(r) => r[\"_measurement\"] == \"a\"",
                                                                                },
                                                                            },
                                                                            lparen: [],
                                                                            params: [
                                                                                Property {
                                                                                    base: BaseNode {
                                                                                        location: SourceLocation {
                                                                                            start: "line: 3, column: 17",
                                                                                            end: "line: 3, column: 18",
                                                                                            source: "r",
                                                                                        },
                                                                                    },
                                                                                    key: Identifier(
                                                                                        Identifier {
                                                                                            base: BaseNode {
                                                                                                location: SourceLocation {
                                                                                                    start: "line: 3, column: 17",
                                                                                                    end: "line: 3, column: 18",
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
                                                                            body: Expr(
                                                                                Binary(
                                                                                    BinaryExpr {
                                                                                        base: BaseNode {
                                                                                            location: SourceLocation {
                                                                                                start: "line: 3, column: 23",
                                                                                                end: "line: 3, column: 47",
                                                                                                source: "r[\"_measurement\"] == \"a\"",
                                                                                            },
                                                                                        },
                                                                                        operator: EqualOperator,
                                                                                        left: Member(
                                                                                            MemberExpr {
                                                                                                base: BaseNode {
                                                                                                    location: SourceLocation {
                                                                                                        start: "line: 3, column: 23",
                                                                                                        end: "line: 3, column: 40",
                                                                                                        source: "r[\"_measurement\"]",
                                                                                                    },
                                                                                                },
                                                                                                object: Identifier(
                                                                                                    Identifier {
                                                                                                        base: BaseNode {
                                                                                                            location: SourceLocation {
                                                                                                                start: "line: 3, column: 23",
                                                                                                                end: "line: 3, column: 24",
                                                                                                                source: "r",
                                                                                                            },
                                                                                                        },
                                                                                                        name: "r",
                                                                                                    },
                                                                                                ),
                                                                                                lbrack: [],
                                                                                                property: StringLit(
                                                                                                    StringLit {
                                                                                                        base: BaseNode {
                                                                                                            location: SourceLocation {
                                                                                                                start: "line: 3, column: 25",
                                                                                                                end: "line: 3, column: 39",
                                                                                                                source: "\"_measurement\"",
                                                                                                            },
                                                                                                        },
                                                                                                        value: "_measurement",
                                                                                                    },
                                                                                                ),
                                                                                                rbrack: [],
                                                                                            },
                                                                                        ),
                                                                                        right: StringLit(
                                                                                            StringLit {
                                                                                                base: BaseNode {
                                                                                                    location: SourceLocation {
                                                                                                        start: "line: 3, column: 44",
                                                                                                        end: "line: 3, column: 47",
                                                                                                        source: "\"a\"",
                                                                                                    },
                                                                                                },
                                                                                                value: "a",
                                                                                            },
                                                                                        ),
                                                                                    },
                                                                                ),
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
                                call: CallExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 4, column: 5",
                                            end: "line: 4, column: 21",
                                            source: "range(start:-1h)",
                                        },
                                    },
                                    callee: Identifier(
                                        Identifier {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 4, column: 5",
                                                    end: "line: 4, column: 10",
                                                    source: "range",
                                                },
                                            },
                                            name: "range",
                                        },
                                    ),
                                    lparen: [],
                                    arguments: [
                                        Object(
                                            ObjectExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 4, column: 11",
                                                        end: "line: 4, column: 20",
                                                        source: "start:-1h",
                                                    },
                                                },
                                                lbrace: [],
                                                with: None,
                                                properties: [
                                                    Property {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 4, column: 11",
                                                                end: "line: 4, column: 20",
                                                                source: "start:-1h",
                                                            },
                                                        },
                                                        key: Identifier(
                                                            Identifier {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 4, column: 11",
                                                                        end: "line: 4, column: 16",
                                                                        source: "start",
                                                                    },
                                                                },
                                                                name: "start",
                                                            },
                                                        ),
                                                        separator: [],
                                                        value: Some(
                                                            Unary(
                                                                UnaryExpr {
                                                                    base: BaseNode {
                                                                        location: SourceLocation {
                                                                            start: "line: 4, column: 17",
                                                                            end: "line: 4, column: 20",
                                                                            source: "-1h",
                                                                        },
                                                                    },
                                                                    operator: SubtractionOperator,
                                                                    argument: Duration(
                                                                        DurationLit {
                                                                            base: BaseNode {
                                                                                location: SourceLocation {
                                                                                    start: "line: 4, column: 18",
                                                                                    end: "line: 4, column: 20",
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
                Variable(
                    VariableAssgn {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 6, column: 1",
                                end: "line: 8, column: 21",
                                source: "b = from(bucket:\"Flux/autogen\")\n\t|> filter(fn: (r) => r[\"_measurement\"] == \"b\")\n\t|> range(start:-1h)",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 6, column: 1",
                                    end: "line: 6, column: 2",
                                    source: "b",
                                },
                            },
                            name: "b",
                        },
                        init: PipeExpr(
                            PipeExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 6, column: 5",
                                        end: "line: 8, column: 21",
                                        source: "from(bucket:\"Flux/autogen\")\n\t|> filter(fn: (r) => r[\"_measurement\"] == \"b\")\n\t|> range(start:-1h)",
                                    },
                                },
                                argument: PipeExpr(
                                    PipeExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 6, column: 5",
                                                end: "line: 7, column: 48",
                                                source: "from(bucket:\"Flux/autogen\")\n\t|> filter(fn: (r) => r[\"_measurement\"] == \"b\")",
                                            },
                                        },
                                        argument: Call(
                                            CallExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 6, column: 5",
                                                        end: "line: 6, column: 32",
                                                        source: "from(bucket:\"Flux/autogen\")",
                                                    },
                                                },
                                                callee: Identifier(
                                                    Identifier {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 6, column: 5",
                                                                end: "line: 6, column: 9",
                                                                source: "from",
                                                            },
                                                        },
                                                        name: "from",
                                                    },
                                                ),
                                                lparen: [],
                                                arguments: [
                                                    Object(
                                                        ObjectExpr {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 6, column: 10",
                                                                    end: "line: 6, column: 31",
                                                                    source: "bucket:\"Flux/autogen\"",
                                                                },
                                                            },
                                                            lbrace: [],
                                                            with: None,
                                                            properties: [
                                                                Property {
                                                                    base: BaseNode {
                                                                        location: SourceLocation {
                                                                            start: "line: 6, column: 10",
                                                                            end: "line: 6, column: 31",
                                                                            source: "bucket:\"Flux/autogen\"",
                                                                        },
                                                                    },
                                                                    key: Identifier(
                                                                        Identifier {
                                                                            base: BaseNode {
                                                                                location: SourceLocation {
                                                                                    start: "line: 6, column: 10",
                                                                                    end: "line: 6, column: 16",
                                                                                    source: "bucket",
                                                                                },
                                                                            },
                                                                            name: "bucket",
                                                                        },
                                                                    ),
                                                                    separator: [],
                                                                    value: Some(
                                                                        StringLit(
                                                                            StringLit {
                                                                                base: BaseNode {
                                                                                    location: SourceLocation {
                                                                                        start: "line: 6, column: 17",
                                                                                        end: "line: 6, column: 31",
                                                                                        source: "\"Flux/autogen\"",
                                                                                    },
                                                                                },
                                                                                value: "Flux/autogen",
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
                                        ),
                                        call: CallExpr {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 7, column: 5",
                                                    end: "line: 7, column: 48",
                                                    source: "filter(fn: (r) => r[\"_measurement\"] == \"b\")",
                                                },
                                            },
                                            callee: Identifier(
                                                Identifier {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 7, column: 5",
                                                            end: "line: 7, column: 11",
                                                            source: "filter",
                                                        },
                                                    },
                                                    name: "filter",
                                                },
                                            ),
                                            lparen: [],
                                            arguments: [
                                                Object(
                                                    ObjectExpr {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 7, column: 12",
                                                                end: "line: 7, column: 47",
                                                                source: "fn: (r) => r[\"_measurement\"] == \"b\"",
                                                            },
                                                        },
                                                        lbrace: [],
                                                        with: None,
                                                        properties: [
                                                            Property {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 7, column: 12",
                                                                        end: "line: 7, column: 47",
                                                                        source: "fn: (r) => r[\"_measurement\"] == \"b\"",
                                                                    },
                                                                },
                                                                key: Identifier(
                                                                    Identifier {
                                                                        base: BaseNode {
                                                                            location: SourceLocation {
                                                                                start: "line: 7, column: 12",
                                                                                end: "line: 7, column: 14",
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
                                                                                    start: "line: 7, column: 16",
                                                                                    end: "line: 7, column: 47",
                                                                                    source: "(r) => r[\"_measurement\"] == \"b\"",
                                                                                },
                                                                            },
                                                                            lparen: [],
                                                                            params: [
                                                                                Property {
                                                                                    base: BaseNode {
                                                                                        location: SourceLocation {
                                                                                            start: "line: 7, column: 17",
                                                                                            end: "line: 7, column: 18",
                                                                                            source: "r",
                                                                                        },
                                                                                    },
                                                                                    key: Identifier(
                                                                                        Identifier {
                                                                                            base: BaseNode {
                                                                                                location: SourceLocation {
                                                                                                    start: "line: 7, column: 17",
                                                                                                    end: "line: 7, column: 18",
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
                                                                            body: Expr(
                                                                                Binary(
                                                                                    BinaryExpr {
                                                                                        base: BaseNode {
                                                                                            location: SourceLocation {
                                                                                                start: "line: 7, column: 23",
                                                                                                end: "line: 7, column: 47",
                                                                                                source: "r[\"_measurement\"] == \"b\"",
                                                                                            },
                                                                                        },
                                                                                        operator: EqualOperator,
                                                                                        left: Member(
                                                                                            MemberExpr {
                                                                                                base: BaseNode {
                                                                                                    location: SourceLocation {
                                                                                                        start: "line: 7, column: 23",
                                                                                                        end: "line: 7, column: 40",
                                                                                                        source: "r[\"_measurement\"]",
                                                                                                    },
                                                                                                },
                                                                                                object: Identifier(
                                                                                                    Identifier {
                                                                                                        base: BaseNode {
                                                                                                            location: SourceLocation {
                                                                                                                start: "line: 7, column: 23",
                                                                                                                end: "line: 7, column: 24",
                                                                                                                source: "r",
                                                                                                            },
                                                                                                        },
                                                                                                        name: "r",
                                                                                                    },
                                                                                                ),
                                                                                                lbrack: [],
                                                                                                property: StringLit(
                                                                                                    StringLit {
                                                                                                        base: BaseNode {
                                                                                                            location: SourceLocation {
                                                                                                                start: "line: 7, column: 25",
                                                                                                                end: "line: 7, column: 39",
                                                                                                                source: "\"_measurement\"",
                                                                                                            },
                                                                                                        },
                                                                                                        value: "_measurement",
                                                                                                    },
                                                                                                ),
                                                                                                rbrack: [],
                                                                                            },
                                                                                        ),
                                                                                        right: StringLit(
                                                                                            StringLit {
                                                                                                base: BaseNode {
                                                                                                    location: SourceLocation {
                                                                                                        start: "line: 7, column: 44",
                                                                                                        end: "line: 7, column: 47",
                                                                                                        source: "\"b\"",
                                                                                                    },
                                                                                                },
                                                                                                value: "b",
                                                                                            },
                                                                                        ),
                                                                                    },
                                                                                ),
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
                                call: CallExpr {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 8, column: 5",
                                            end: "line: 8, column: 21",
                                            source: "range(start:-1h)",
                                        },
                                    },
                                    callee: Identifier(
                                        Identifier {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 8, column: 5",
                                                    end: "line: 8, column: 10",
                                                    source: "range",
                                                },
                                            },
                                            name: "range",
                                        },
                                    ),
                                    lparen: [],
                                    arguments: [
                                        Object(
                                            ObjectExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 8, column: 11",
                                                        end: "line: 8, column: 20",
                                                        source: "start:-1h",
                                                    },
                                                },
                                                lbrace: [],
                                                with: None,
                                                properties: [
                                                    Property {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 8, column: 11",
                                                                end: "line: 8, column: 20",
                                                                source: "start:-1h",
                                                            },
                                                        },
                                                        key: Identifier(
                                                            Identifier {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 8, column: 11",
                                                                        end: "line: 8, column: 16",
                                                                        source: "start",
                                                                    },
                                                                },
                                                                name: "start",
                                                            },
                                                        ),
                                                        separator: [],
                                                        value: Some(
                                                            Unary(
                                                                UnaryExpr {
                                                                    base: BaseNode {
                                                                        location: SourceLocation {
                                                                            start: "line: 8, column: 17",
                                                                            end: "line: 8, column: 20",
                                                                            source: "-1h",
                                                                        },
                                                                    },
                                                                    operator: SubtractionOperator,
                                                                    argument: Duration(
                                                                        DurationLit {
                                                                            base: BaseNode {
                                                                                location: SourceLocation {
                                                                                    start: "line: 8, column: 18",
                                                                                    end: "line: 8, column: 20",
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
                Expr(
                    ExprStmt {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 10, column: 1",
                                end: "line: 10, column: 86",
                                source: "join(tables:[a,b], on:[\"t1\"], fn: (a,b) => (a[\"_field\"] - b[\"_field\"]) / b[\"_field\"])",
                            },
                        },
                        expression: Call(
                            CallExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 10, column: 1",
                                        end: "line: 10, column: 86",
                                        source: "join(tables:[a,b], on:[\"t1\"], fn: (a,b) => (a[\"_field\"] - b[\"_field\"]) / b[\"_field\"])",
                                    },
                                },
                                callee: Identifier(
                                    Identifier {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 10, column: 1",
                                                end: "line: 10, column: 5",
                                                source: "join",
                                            },
                                        },
                                        name: "join",
                                    },
                                ),
                                lparen: [],
                                arguments: [
                                    Object(
                                        ObjectExpr {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 10, column: 6",
                                                    end: "line: 10, column: 85",
                                                    source: "tables:[a,b], on:[\"t1\"], fn: (a,b) => (a[\"_field\"] - b[\"_field\"]) / b[\"_field\"]",
                                                },
                                            },
                                            lbrace: [],
                                            with: None,
                                            properties: [
                                                Property {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 10, column: 6",
                                                            end: "line: 10, column: 18",
                                                            source: "tables:[a,b]",
                                                        },
                                                    },
                                                    key: Identifier(
                                                        Identifier {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 10, column: 6",
                                                                    end: "line: 10, column: 12",
                                                                    source: "tables",
                                                                },
                                                            },
                                                            name: "tables",
                                                        },
                                                    ),
                                                    separator: [],
                                                    value: Some(
                                                        Array(
                                                            ArrayExpr {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 10, column: 13",
                                                                        end: "line: 10, column: 18",
                                                                        source: "[a,b]",
                                                                    },
                                                                },
                                                                lbrack: [],
                                                                elements: [
                                                                    ArrayItem {
                                                                        expression: Identifier(
                                                                            Identifier {
                                                                                base: BaseNode {
                                                                                    location: SourceLocation {
                                                                                        start: "line: 10, column: 14",
                                                                                        end: "line: 10, column: 15",
                                                                                        source: "a",
                                                                                    },
                                                                                },
                                                                                name: "a",
                                                                            },
                                                                        ),
                                                                        comma: [],
                                                                    },
                                                                    ArrayItem {
                                                                        expression: Identifier(
                                                                            Identifier {
                                                                                base: BaseNode {
                                                                                    location: SourceLocation {
                                                                                        start: "line: 10, column: 16",
                                                                                        end: "line: 10, column: 17",
                                                                                        source: "b",
                                                                                    },
                                                                                },
                                                                                name: "b",
                                                                            },
                                                                        ),
                                                                        comma: [],
                                                                    },
                                                                ],
                                                                rbrack: [],
                                                            },
                                                        ),
                                                    ),
                                                    comma: [],
                                                },
                                                Property {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 10, column: 20",
                                                            end: "line: 10, column: 29",
                                                            source: "on:[\"t1\"]",
                                                        },
                                                    },
                                                    key: Identifier(
                                                        Identifier {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 10, column: 20",
                                                                    end: "line: 10, column: 22",
                                                                    source: "on",
                                                                },
                                                            },
                                                            name: "on",
                                                        },
                                                    ),
                                                    separator: [],
                                                    value: Some(
                                                        Array(
                                                            ArrayExpr {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 10, column: 23",
                                                                        end: "line: 10, column: 29",
                                                                        source: "[\"t1\"]",
                                                                    },
                                                                },
                                                                lbrack: [],
                                                                elements: [
                                                                    ArrayItem {
                                                                        expression: StringLit(
                                                                            StringLit {
                                                                                base: BaseNode {
                                                                                    location: SourceLocation {
                                                                                        start: "line: 10, column: 24",
                                                                                        end: "line: 10, column: 28",
                                                                                        source: "\"t1\"",
                                                                                    },
                                                                                },
                                                                                value: "t1",
                                                                            },
                                                                        ),
                                                                        comma: [],
                                                                    },
                                                                ],
                                                                rbrack: [],
                                                            },
                                                        ),
                                                    ),
                                                    comma: [],
                                                },
                                                Property {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 10, column: 31",
                                                            end: "line: 10, column: 85",
                                                            source: "fn: (a,b) => (a[\"_field\"] - b[\"_field\"]) / b[\"_field\"]",
                                                        },
                                                    },
                                                    key: Identifier(
                                                        Identifier {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 10, column: 31",
                                                                    end: "line: 10, column: 33",
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
                                                                        start: "line: 10, column: 35",
                                                                        end: "line: 10, column: 85",
                                                                        source: "(a,b) => (a[\"_field\"] - b[\"_field\"]) / b[\"_field\"]",
                                                                    },
                                                                },
                                                                lparen: [],
                                                                params: [
                                                                    Property {
                                                                        base: BaseNode {
                                                                            location: SourceLocation {
                                                                                start: "line: 10, column: 36",
                                                                                end: "line: 10, column: 37",
                                                                                source: "a",
                                                                            },
                                                                        },
                                                                        key: Identifier(
                                                                            Identifier {
                                                                                base: BaseNode {
                                                                                    location: SourceLocation {
                                                                                        start: "line: 10, column: 36",
                                                                                        end: "line: 10, column: 37",
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
                                                                                start: "line: 10, column: 38",
                                                                                end: "line: 10, column: 39",
                                                                                source: "b",
                                                                            },
                                                                        },
                                                                        key: Identifier(
                                                                            Identifier {
                                                                                base: BaseNode {
                                                                                    location: SourceLocation {
                                                                                        start: "line: 10, column: 38",
                                                                                        end: "line: 10, column: 39",
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
                                                                                    start: "line: 10, column: 44",
                                                                                    end: "line: 10, column: 85",
                                                                                    source: "(a[\"_field\"] - b[\"_field\"]) / b[\"_field\"]",
                                                                                },
                                                                            },
                                                                            operator: DivisionOperator,
                                                                            left: Paren(
                                                                                ParenExpr {
                                                                                    base: BaseNode {
                                                                                        location: SourceLocation {
                                                                                            start: "line: 10, column: 44",
                                                                                            end: "line: 10, column: 71",
                                                                                            source: "(a[\"_field\"] - b[\"_field\"])",
                                                                                        },
                                                                                    },
                                                                                    lparen: [],
                                                                                    expression: Binary(
                                                                                        BinaryExpr {
                                                                                            base: BaseNode {
                                                                                                location: SourceLocation {
                                                                                                    start: "line: 10, column: 45",
                                                                                                    end: "line: 10, column: 70",
                                                                                                    source: "a[\"_field\"] - b[\"_field\"]",
                                                                                                },
                                                                                            },
                                                                                            operator: SubtractionOperator,
                                                                                            left: Member(
                                                                                                MemberExpr {
                                                                                                    base: BaseNode {
                                                                                                        location: SourceLocation {
                                                                                                            start: "line: 10, column: 45",
                                                                                                            end: "line: 10, column: 56",
                                                                                                            source: "a[\"_field\"]",
                                                                                                        },
                                                                                                    },
                                                                                                    object: Identifier(
                                                                                                        Identifier {
                                                                                                            base: BaseNode {
                                                                                                                location: SourceLocation {
                                                                                                                    start: "line: 10, column: 45",
                                                                                                                    end: "line: 10, column: 46",
                                                                                                                    source: "a",
                                                                                                                },
                                                                                                            },
                                                                                                            name: "a",
                                                                                                        },
                                                                                                    ),
                                                                                                    lbrack: [],
                                                                                                    property: StringLit(
                                                                                                        StringLit {
                                                                                                            base: BaseNode {
                                                                                                                location: SourceLocation {
                                                                                                                    start: "line: 10, column: 47",
                                                                                                                    end: "line: 10, column: 55",
                                                                                                                    source: "\"_field\"",
                                                                                                                },
                                                                                                            },
                                                                                                            value: "_field",
                                                                                                        },
                                                                                                    ),
                                                                                                    rbrack: [],
                                                                                                },
                                                                                            ),
                                                                                            right: Member(
                                                                                                MemberExpr {
                                                                                                    base: BaseNode {
                                                                                                        location: SourceLocation {
                                                                                                            start: "line: 10, column: 59",
                                                                                                            end: "line: 10, column: 70",
                                                                                                            source: "b[\"_field\"]",
                                                                                                        },
                                                                                                    },
                                                                                                    object: Identifier(
                                                                                                        Identifier {
                                                                                                            base: BaseNode {
                                                                                                                location: SourceLocation {
                                                                                                                    start: "line: 10, column: 59",
                                                                                                                    end: "line: 10, column: 60",
                                                                                                                    source: "b",
                                                                                                                },
                                                                                                            },
                                                                                                            name: "b",
                                                                                                        },
                                                                                                    ),
                                                                                                    lbrack: [],
                                                                                                    property: StringLit(
                                                                                                        StringLit {
                                                                                                            base: BaseNode {
                                                                                                                location: SourceLocation {
                                                                                                                    start: "line: 10, column: 61",
                                                                                                                    end: "line: 10, column: 69",
                                                                                                                    source: "\"_field\"",
                                                                                                                },
                                                                                                            },
                                                                                                            value: "_field",
                                                                                                        },
                                                                                                    ),
                                                                                                    rbrack: [],
                                                                                                },
                                                                                            ),
                                                                                        },
                                                                                    ),
                                                                                    rparen: [],
                                                                                },
                                                                            ),
                                                                            right: Member(
                                                                                MemberExpr {
                                                                                    base: BaseNode {
                                                                                        location: SourceLocation {
                                                                                            start: "line: 10, column: 74",
                                                                                            end: "line: 10, column: 85",
                                                                                            source: "b[\"_field\"]",
                                                                                        },
                                                                                    },
                                                                                    object: Identifier(
                                                                                        Identifier {
                                                                                            base: BaseNode {
                                                                                                location: SourceLocation {
                                                                                                    start: "line: 10, column: 74",
                                                                                                    end: "line: 10, column: 75",
                                                                                                    source: "b",
                                                                                                },
                                                                                            },
                                                                                            name: "b",
                                                                                        },
                                                                                    ),
                                                                                    lbrack: [],
                                                                                    property: StringLit(
                                                                                        StringLit {
                                                                                            base: BaseNode {
                                                                                                location: SourceLocation {
                                                                                                    start: "line: 10, column: 76",
                                                                                                    end: "line: 10, column: 84",
                                                                                                    source: "\"_field\"",
                                                                                                },
                                                                                            },
                                                                                            value: "_field",
                                                                                        },
                                                                                    ),
                                                                                    rbrack: [],
                                                                                },
                                                                            ),
                                                                        },
                                                                    ),
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
                        ),
                    },
                ),
            ],
            eof: [],
        }
    "#]].assert_debug_eq(&parsed);
}
