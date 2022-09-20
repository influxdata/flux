use super::*;

#[test]
fn property_list_missing_property() {
    let mut p = Parser::new(r#"o = {a: "a",, b: 7}"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 20",
                    source: "o = {a: \"a\",, b: 7}",
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
                                end: "line: 1, column: 20",
                                source: "o = {a: \"a\",, b: 7}",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 1",
                                    end: "line: 1, column: 2",
                                    source: "o",
                                },
                            },
                            name: "o",
                        },
                        init: Object(
                            ObjectExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 5",
                                        end: "line: 1, column: 20",
                                        source: "{a: \"a\",, b: 7}",
                                    },
                                },
                                lbrace: [],
                                with: None,
                                properties: [
                                    Property {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 6",
                                                end: "line: 1, column: 12",
                                                source: "a: \"a\"",
                                            },
                                        },
                                        key: Identifier(
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
                                        separator: [],
                                        value: Some(
                                            StringLit(
                                                StringLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 9",
                                                            end: "line: 1, column: 12",
                                                            source: "\"a\"",
                                                        },
                                                    },
                                                    value: "a",
                                                },
                                            ),
                                        ),
                                        comma: [],
                                    },
                                    Property {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 13",
                                                end: "line: 1, column: 13",
                                                source: "",
                                            },
                                            errors: [
                                                "missing property in property list",
                                            ],
                                        },
                                        key: StringLit(
                                            StringLit {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 13",
                                                        end: "line: 1, column: 13",
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
                                    Property {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 15",
                                                end: "line: 1, column: 19",
                                                source: "b: 7",
                                            },
                                        },
                                        key: Identifier(
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
                                        separator: [],
                                        value: Some(
                                            Integer(
                                                IntegerLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 18",
                                                            end: "line: 1, column: 19",
                                                            source: "7",
                                                        },
                                                    },
                                                    value: 7,
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
            ],
            eof: [],
        }
    "#]]
    .assert_debug_eq(&parsed);
}

#[test]
fn property_list_missing_key() {
    let mut p = Parser::new(r#"o = {: "a"}"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 12",
                    source: "o = {: \"a\"}",
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
                                source: "o = {: \"a\"}",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 1",
                                    end: "line: 1, column: 2",
                                    source: "o",
                                },
                            },
                            name: "o",
                        },
                        init: Object(
                            ObjectExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 5",
                                        end: "line: 1, column: 12",
                                        source: "{: \"a\"}",
                                    },
                                },
                                lbrace: [],
                                with: None,
                                properties: [
                                    Property {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 6",
                                                end: "line: 1, column: 11",
                                                source: ": \"a\"",
                                            },
                                            errors: [
                                                "missing property key",
                                            ],
                                        },
                                        key: StringLit(
                                            StringLit {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 6",
                                                        end: "line: 1, column: 6",
                                                        source: "",
                                                    },
                                                },
                                                value: "<invalid>",
                                            },
                                        ),
                                        separator: [],
                                        value: Some(
                                            StringLit(
                                                StringLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 8",
                                                            end: "line: 1, column: 11",
                                                            source: "\"a\"",
                                                        },
                                                    },
                                                    value: "a",
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
            ],
            eof: [],
        }
    "#]]
    .assert_debug_eq(&parsed);
}

#[test]
fn property_list_missing_value() {
    let mut p = Parser::new(r#"o = {a:}"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 9",
                    source: "o = {a:}",
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
                                end: "line: 1, column: 9",
                                source: "o = {a:}",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 1",
                                    end: "line: 1, column: 2",
                                    source: "o",
                                },
                            },
                            name: "o",
                        },
                        init: Object(
                            ObjectExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 5",
                                        end: "line: 1, column: 9",
                                        source: "{a:}",
                                    },
                                },
                                lbrace: [],
                                with: None,
                                properties: [
                                    Property {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 6",
                                                end: "line: 1, column: 7",
                                                source: "a",
                                            },
                                            errors: [
                                                "missing property value",
                                            ],
                                        },
                                        key: Identifier(
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
                                        separator: [],
                                        value: None,
                                        comma: [],
                                    },
                                ],
                                rbrace: [],
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
fn property_list_missing_comma() {
    let mut p = Parser::new(r#"o = {a: "a" b: 30}"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 19",
                    source: "o = {a: \"a\" b: 30}",
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
                                end: "line: 1, column: 19",
                                source: "o = {a: \"a\" b: 30}",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 1",
                                    end: "line: 1, column: 2",
                                    source: "o",
                                },
                            },
                            name: "o",
                        },
                        init: Object(
                            ObjectExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 5",
                                        end: "line: 1, column: 19",
                                        source: "{a: \"a\" b: 30}",
                                    },
                                },
                                lbrace: [],
                                with: None,
                                properties: [
                                    Property {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 6",
                                                end: "line: 1, column: 14",
                                                source: "a: \"a\" b",
                                            },
                                        },
                                        key: Identifier(
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
                                        separator: [],
                                        value: Some(
                                            Binary(
                                                BinaryExpr {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 9",
                                                            end: "line: 1, column: 14",
                                                            source: "\"a\" b",
                                                        },
                                                    },
                                                    operator: InvalidOperator,
                                                    left: StringLit(
                                                        StringLit {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 1, column: 9",
                                                                    end: "line: 1, column: 12",
                                                                    source: "\"a\"",
                                                                },
                                                            },
                                                            value: "a",
                                                        },
                                                    ),
                                                    right: Identifier(
                                                        Identifier {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 1, column: 13",
                                                                    end: "line: 1, column: 14",
                                                                    source: "b",
                                                                },
                                                            },
                                                            name: "b",
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
                                                start: "line: 1, column: 14",
                                                end: "line: 1, column: 18",
                                                source: ": 30",
                                            },
                                            errors: [
                                                "missing property key",
                                            ],
                                        },
                                        key: StringLit(
                                            StringLit {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 14",
                                                        end: "line: 1, column: 14",
                                                        source: "",
                                                    },
                                                },
                                                value: "<invalid>",
                                            },
                                        ),
                                        separator: [],
                                        value: Some(
                                            Integer(
                                                IntegerLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 16",
                                                            end: "line: 1, column: 18",
                                                            source: "30",
                                                        },
                                                        errors: [
                                                            "expected comma in property list, got COLON",
                                                        ],
                                                    },
                                                    value: 30,
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
            ],
            eof: [],
        }
    "#]].assert_debug_eq(&parsed);
}

#[test]
fn property_list_trailing_comma() {
    let mut p = Parser::new(r#"o = {a: "a",}"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 14",
                    source: "o = {a: \"a\",}",
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
                                end: "line: 1, column: 14",
                                source: "o = {a: \"a\",}",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 1",
                                    end: "line: 1, column: 2",
                                    source: "o",
                                },
                            },
                            name: "o",
                        },
                        init: Object(
                            ObjectExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 5",
                                        end: "line: 1, column: 14",
                                        source: "{a: \"a\",}",
                                    },
                                },
                                lbrace: [],
                                with: None,
                                properties: [
                                    Property {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 6",
                                                end: "line: 1, column: 12",
                                                source: "a: \"a\"",
                                            },
                                        },
                                        key: Identifier(
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
                                        separator: [],
                                        value: Some(
                                            StringLit(
                                                StringLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 9",
                                                            end: "line: 1, column: 12",
                                                            source: "\"a\"",
                                                        },
                                                    },
                                                    value: "a",
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
            ],
            eof: [],
        }
    "#]]
    .assert_debug_eq(&parsed);
}

#[test]
fn property_list_bad_property() {
    let mut p = Parser::new(r#"o = {a: "a", 30, b: 7}"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 23",
                    source: "o = {a: \"a\", 30, b: 7}",
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
                                end: "line: 1, column: 23",
                                source: "o = {a: \"a\", 30, b: 7}",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 1",
                                    end: "line: 1, column: 2",
                                    source: "o",
                                },
                            },
                            name: "o",
                        },
                        init: Object(
                            ObjectExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 5",
                                        end: "line: 1, column: 23",
                                        source: "{a: \"a\", 30, b: 7}",
                                    },
                                },
                                lbrace: [],
                                with: None,
                                properties: [
                                    Property {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 6",
                                                end: "line: 1, column: 12",
                                                source: "a: \"a\"",
                                            },
                                        },
                                        key: Identifier(
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
                                        separator: [],
                                        value: Some(
                                            StringLit(
                                                StringLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 9",
                                                            end: "line: 1, column: 12",
                                                            source: "\"a\"",
                                                        },
                                                    },
                                                    value: "a",
                                                },
                                            ),
                                        ),
                                        comma: [],
                                    },
                                    Property {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 14",
                                                end: "line: 1, column: 16",
                                                source: "30",
                                            },
                                            errors: [
                                                "unexpected token for property key: INT (30)",
                                            ],
                                        },
                                        key: StringLit(
                                            StringLit {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 14",
                                                        end: "line: 1, column: 14",
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
                                    Property {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 18",
                                                end: "line: 1, column: 22",
                                                source: "b: 7",
                                            },
                                        },
                                        key: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 18",
                                                        end: "line: 1, column: 19",
                                                        source: "b",
                                                    },
                                                },
                                                name: "b",
                                            },
                                        ),
                                        separator: [],
                                        value: Some(
                                            Integer(
                                                IntegerLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 21",
                                                            end: "line: 1, column: 22",
                                                            source: "7",
                                                        },
                                                    },
                                                    value: 7,
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
            ],
            eof: [],
        }
    "#]]
    .assert_debug_eq(&parsed);
}
