use super::*;

#[test]
fn regex_literal() {
    let mut p = Parser::new(r#"/.*/"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 5",
                    source: "/.*/",
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
                                source: "/.*/",
                            },
                        },
                        expression: Regexp(
                            RegexpLit {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 5",
                                        source: "/.*/",
                                    },
                                },
                                value: ".*",
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
fn regex_literal_with_escape_sequence() {
    let mut p = Parser::new(r"/a\/b\\c\d/");
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 12",
                    source: "/a\\/b\\\\c\\d/",
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
                                source: "/a\\/b\\\\c\\d/",
                            },
                        },
                        expression: Regexp(
                            RegexpLit {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 12",
                                        source: "/a\\/b\\\\c\\d/",
                                    },
                                },
                                value: "a/b\\\\c\\d",
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
fn regex_literal_with_hex_escape() {
    let mut p = Parser::new(r"/^\xe6\x97\xa5\xe6\x9c\xac\xe8\xaa\x9e(ZZ)?$/");
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 46",
                    source: "/^\\xe6\\x97\\xa5\\xe6\\x9c\\xac\\xe8\\xaa\\x9e(ZZ)?$/",
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
                                end: "line: 1, column: 46",
                                source: "/^\\xe6\\x97\\xa5\\xe6\\x9c\\xac\\xe8\\xaa\\x9e(ZZ)?$/",
                            },
                        },
                        expression: Regexp(
                            RegexpLit {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 46",
                                        source: "/^\\xe6\\x97\\xa5\\xe6\\x9c\\xac\\xe8\\xaa\\x9e(ZZ)?$/",
                                    },
                                },
                                value: "^日本語(ZZ)?$",
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
fn regex_literal_empty_pattern() {
    let mut p = Parser::new(r#"/(:?)/"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 7",
                    source: "/(:?)/",
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
                                source: "/(:?)/",
                            },
                        },
                        expression: Regexp(
                            RegexpLit {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 7",
                                        source: "/(:?)/",
                                    },
                                },
                                value: "(:?)",
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
fn bad_regex_literal() {
    let mut p = Parser::new(r#"/*/"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 4",
                    source: "/*/",
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
                                source: "/*/",
                            },
                        },
                        expression: Regexp(
                            RegexpLit {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 4",
                                        source: "/*/",
                                    },
                                    errors: [
                                        "regex parse error: * error: repetition operator missing expression",
                                    ],
                                },
                                value: "",
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
fn duration_literal_all_units() {
    let mut p = Parser::new(r#"dur = 1y3mo2w1d4h1m30s1ms2µs70ns"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 34",
                    source: "dur = 1y3mo2w1d4h1m30s1ms2µs70ns",
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
                                end: "line: 1, column: 34",
                                source: "dur = 1y3mo2w1d4h1m30s1ms2µs70ns",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 1",
                                    end: "line: 1, column: 4",
                                    source: "dur",
                                },
                            },
                            name: "dur",
                        },
                        init: Duration(
                            DurationLit {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 7",
                                        end: "line: 1, column: 34",
                                        source: "1y3mo2w1d4h1m30s1ms2µs70ns",
                                    },
                                },
                                values: [
                                    Duration {
                                        magnitude: 1,
                                        unit: "y",
                                    },
                                    Duration {
                                        magnitude: 3,
                                        unit: "mo",
                                    },
                                    Duration {
                                        magnitude: 2,
                                        unit: "w",
                                    },
                                    Duration {
                                        magnitude: 1,
                                        unit: "d",
                                    },
                                    Duration {
                                        magnitude: 4,
                                        unit: "h",
                                    },
                                    Duration {
                                        magnitude: 1,
                                        unit: "m",
                                    },
                                    Duration {
                                        magnitude: 30,
                                        unit: "s",
                                    },
                                    Duration {
                                        magnitude: 1,
                                        unit: "ms",
                                    },
                                    Duration {
                                        magnitude: 2,
                                        unit: "us",
                                    },
                                    Duration {
                                        magnitude: 70,
                                        unit: "ns",
                                    },
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
fn duration_literal_leading_zero() {
    let mut p = Parser::new(r#"dur = 01y02mo03w04d05h06m07s08ms09µs010ns"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 43",
                    source: "dur = 01y02mo03w04d05h06m07s08ms09µs010ns",
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
                                end: "line: 1, column: 43",
                                source: "dur = 01y02mo03w04d05h06m07s08ms09µs010ns",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 1",
                                    end: "line: 1, column: 4",
                                    source: "dur",
                                },
                            },
                            name: "dur",
                        },
                        init: Duration(
                            DurationLit {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 7",
                                        end: "line: 1, column: 43",
                                        source: "01y02mo03w04d05h06m07s08ms09µs010ns",
                                    },
                                },
                                values: [
                                    Duration {
                                        magnitude: 1,
                                        unit: "y",
                                    },
                                    Duration {
                                        magnitude: 2,
                                        unit: "mo",
                                    },
                                    Duration {
                                        magnitude: 3,
                                        unit: "w",
                                    },
                                    Duration {
                                        magnitude: 4,
                                        unit: "d",
                                    },
                                    Duration {
                                        magnitude: 5,
                                        unit: "h",
                                    },
                                    Duration {
                                        magnitude: 6,
                                        unit: "m",
                                    },
                                    Duration {
                                        magnitude: 7,
                                        unit: "s",
                                    },
                                    Duration {
                                        magnitude: 8,
                                        unit: "ms",
                                    },
                                    Duration {
                                        magnitude: 9,
                                        unit: "us",
                                    },
                                    Duration {
                                        magnitude: 10,
                                        unit: "ns",
                                    },
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
fn duration_literal_months() {
    let mut p = Parser::new(r#"dur = 6mo"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 10",
                    source: "dur = 6mo",
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
                                source: "dur = 6mo",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 1",
                                    end: "line: 1, column: 4",
                                    source: "dur",
                                },
                            },
                            name: "dur",
                        },
                        init: Duration(
                            DurationLit {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 7",
                                        end: "line: 1, column: 10",
                                        source: "6mo",
                                    },
                                },
                                values: [
                                    Duration {
                                        magnitude: 6,
                                        unit: "mo",
                                    },
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
fn duration_literal_milliseconds() {
    let mut p = Parser::new(r#"dur = 500ms"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 12",
                    source: "dur = 500ms",
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
                                source: "dur = 500ms",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 1",
                                    end: "line: 1, column: 4",
                                    source: "dur",
                                },
                            },
                            name: "dur",
                        },
                        init: Duration(
                            DurationLit {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 7",
                                        end: "line: 1, column: 12",
                                        source: "500ms",
                                    },
                                },
                                values: [
                                    Duration {
                                        magnitude: 500,
                                        unit: "ms",
                                    },
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
fn duration_literal_months_minutes_milliseconds() {
    let mut p = Parser::new(r#"dur = 6mo30m500ms"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 18",
                    source: "dur = 6mo30m500ms",
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
                                source: "dur = 6mo30m500ms",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 1",
                                    end: "line: 1, column: 4",
                                    source: "dur",
                                },
                            },
                            name: "dur",
                        },
                        init: Duration(
                            DurationLit {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 7",
                                        end: "line: 1, column: 18",
                                        source: "6mo30m500ms",
                                    },
                                },
                                values: [
                                    Duration {
                                        magnitude: 6,
                                        unit: "mo",
                                    },
                                    Duration {
                                        magnitude: 30,
                                        unit: "m",
                                    },
                                    Duration {
                                        magnitude: 500,
                                        unit: "ms",
                                    },
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
fn date_literal_in_the_default_location() {
    let mut p = Parser::new(r#"now = 2018-11-29"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 17",
                    source: "now = 2018-11-29",
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
                                end: "line: 1, column: 17",
                                source: "now = 2018-11-29",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 1",
                                    end: "line: 1, column: 4",
                                    source: "now",
                                },
                            },
                            name: "now",
                        },
                        init: DateTime(
                            DateTimeLit {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 7",
                                        end: "line: 1, column: 17",
                                        source: "2018-11-29",
                                    },
                                },
                                value: 2018-11-29T00:00:00+00:00,
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
fn date_time_literal_arg() {
    let mut p = Parser::new(r#"range(start: 2018-11-29T09:00:00Z)"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 35",
                    source: "range(start: 2018-11-29T09:00:00Z)",
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
                                end: "line: 1, column: 35",
                                source: "range(start: 2018-11-29T09:00:00Z)",
                            },
                        },
                        expression: Call(
                            CallExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 35",
                                        source: "range(start: 2018-11-29T09:00:00Z)",
                                    },
                                },
                                callee: Identifier(
                                    Identifier {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 1",
                                                end: "line: 1, column: 6",
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
                                                    start: "line: 1, column: 7",
                                                    end: "line: 1, column: 34",
                                                    source: "start: 2018-11-29T09:00:00Z",
                                                },
                                            },
                                            lbrace: [],
                                            with: None,
                                            properties: [
                                                Property {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 7",
                                                            end: "line: 1, column: 34",
                                                            source: "start: 2018-11-29T09:00:00Z",
                                                        },
                                                    },
                                                    key: Identifier(
                                                        Identifier {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 1, column: 7",
                                                                    end: "line: 1, column: 12",
                                                                    source: "start",
                                                                },
                                                            },
                                                            name: "start",
                                                        },
                                                    ),
                                                    separator: [],
                                                    value: Some(
                                                        DateTime(
                                                            DateTimeLit {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 1, column: 14",
                                                                        end: "line: 1, column: 34",
                                                                        source: "2018-11-29T09:00:00Z",
                                                                    },
                                                                },
                                                                value: 2018-11-29T09:00:00+00:00,
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
fn date_time_literal_no_offset_error() {
    let mut p = Parser::new(r#"t = 2018-11-29T09:00:00"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 24",
                    source: "t = 2018-11-29T09:00:00",
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
                                end: "line: 1, column: 24",
                                source: "t = 2018-11-29T09:00:00",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 1",
                                    end: "line: 1, column: 2",
                                    source: "t",
                                },
                            },
                            name: "t",
                        },
                        init: Bad(
                            BadExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 5",
                                        end: "line: 1, column: 24",
                                        source: "2018-11-29T09:00:00",
                                    },
                                },
                                text: "invalid date time literal, missing time offset",
                                expression: None,
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
fn date_time_literal() {
    let mut p = Parser::new(r#"now = 2018-11-29T09:00:00Z"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 27",
                    source: "now = 2018-11-29T09:00:00Z",
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
                                end: "line: 1, column: 27",
                                source: "now = 2018-11-29T09:00:00Z",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 1",
                                    end: "line: 1, column: 4",
                                    source: "now",
                                },
                            },
                            name: "now",
                        },
                        init: DateTime(
                            DateTimeLit {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 7",
                                        end: "line: 1, column: 27",
                                        source: "2018-11-29T09:00:00Z",
                                    },
                                },
                                value: 2018-11-29T09:00:00+00:00,
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
fn date_time_literal_with_fractional_seconds() {
    let mut p = Parser::new(r#"now = 2018-11-29T09:00:00.100000000Z"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 37",
                    source: "now = 2018-11-29T09:00:00.100000000Z",
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
                                end: "line: 1, column: 37",
                                source: "now = 2018-11-29T09:00:00.100000000Z",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 1",
                                    end: "line: 1, column: 4",
                                    source: "now",
                                },
                            },
                            name: "now",
                        },
                        init: DateTime(
                            DateTimeLit {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 7",
                                        end: "line: 1, column: 37",
                                        source: "2018-11-29T09:00:00.100000000Z",
                                    },
                                },
                                value: 2018-11-29T09:00:00.100+00:00,
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
fn integer_literal_overflow() {
    let mut p = Parser::new(r#"100000000000000000000000000000"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 31",
                    source: "100000000000000000000000000000",
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
                                end: "line: 1, column: 31",
                                source: "100000000000000000000000000000",
                            },
                        },
                        expression: Integer(
                            IntegerLit {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 31",
                                        source: "100000000000000000000000000000",
                                    },
                                    errors: [
                                        "invalid integer literal \"100000000000000000000000000000\": value out of range",
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

#[test]
fn dictionary_literal() {
    let mut p = Parser::new(r#"["a":1, "b":2]"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 15",
                    source: "[\"a\":1, \"b\":2]",
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
                                source: "[\"a\":1, \"b\":2]",
                            },
                        },
                        expression: Dict(
                            DictExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 15",
                                        source: "[\"a\":1, \"b\":2]",
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
                                                        start: "line: 1, column: 6",
                                                        end: "line: 1, column: 7",
                                                        source: "1",
                                                    },
                                                },
                                                value: 1,
                                            },
                                        ),
                                        comma: [],
                                    },
                                    DictItem {
                                        key: StringLit(
                                            StringLit {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 9",
                                                        end: "line: 1, column: 12",
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
                                ],
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
fn unclosed_dictionary_literal() {
    let mut p = Parser::new(
        r#"
        A = ["a":1, "b":2
        B = 100
"#,
    );
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 2, column: 9",
                    end: "line: 4, column: 1",
                    source: "A = [\"a\":1, \"b\":2\n        B = 100\n",
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
                                start: "line: 2, column: 9",
                                end: "line: 4, column: 1",
                                source: "A = [\"a\":1, \"b\":2\n        B = 100\n",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 2, column: 9",
                                    end: "line: 2, column: 10",
                                    source: "A",
                                },
                            },
                            name: "A",
                        },
                        init: Dict(
                            DictExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 2, column: 13",
                                        end: "line: 4, column: 1",
                                        source: "[\"a\":1, \"b\":2\n        B = 100\n",
                                    },
                                    errors: [
                                        "expected RBRACK, got EOF",
                                    ],
                                },
                                lbrack: [],
                                elements: [
                                    DictItem {
                                        key: StringLit(
                                            StringLit {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 2, column: 14",
                                                        end: "line: 2, column: 17",
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
                                                        start: "line: 2, column: 18",
                                                        end: "line: 2, column: 19",
                                                        source: "1",
                                                    },
                                                },
                                                value: 1,
                                            },
                                        ),
                                        comma: [],
                                    },
                                    DictItem {
                                        key: StringLit(
                                            StringLit {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 2, column: 21",
                                                        end: "line: 2, column: 24",
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
                                                        start: "line: 2, column: 25",
                                                        end: "line: 2, column: 26",
                                                        source: "2",
                                                    },
                                                },
                                                value: 2,
                                            },
                                        ),
                                        comma: [],
                                    },
                                    DictItem {
                                        key: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 3, column: 9",
                                                        end: "line: 3, column: 10",
                                                        source: "B",
                                                    },
                                                },
                                                name: "B",
                                            },
                                        ),
                                        val: Integer(
                                            IntegerLit {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 3, column: 13",
                                                        end: "line: 3, column: 16",
                                                        source: "100",
                                                    },
                                                    errors: [
                                                        "expected COLON, got ASSIGN (=) at 3:11",
                                                    ],
                                                },
                                                value: 100,
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
    "#]]
    .assert_debug_eq(&parsed);
}
