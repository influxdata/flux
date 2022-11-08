use super::*;

#[test]
fn map_member_expressions() {
    let mut p = Parser::new(
        r#"m = {key1: 1, key2:"value2"}
			m.key1
			m["key2"]
			"#,
    );
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 4, column: 4",
                    source: "m = {key1: 1, key2:\"value2\"}\n\t\t\tm.key1\n\t\t\tm[\"key2\"]\n\t\t\t",
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
                                end: "line: 1, column: 29",
                                source: "m = {key1: 1, key2:\"value2\"}",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 1",
                                    end: "line: 1, column: 2",
                                    source: "m",
                                },
                            },
                            name: "m",
                        },
                        init: Object(
                            ObjectExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 5",
                                        end: "line: 1, column: 29",
                                        source: "{key1: 1, key2:\"value2\"}",
                                    },
                                },
                                lbrace: [],
                                with: None,
                                properties: [
                                    Property {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 6",
                                                end: "line: 1, column: 13",
                                                source: "key1: 1",
                                            },
                                        },
                                        key: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 6",
                                                        end: "line: 1, column: 10",
                                                        source: "key1",
                                                    },
                                                },
                                                name: "key1",
                                            },
                                        ),
                                        separator: [],
                                        value: Some(
                                            Integer(
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
                                        ),
                                        comma: [],
                                    },
                                    Property {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 15",
                                                end: "line: 1, column: 28",
                                                source: "key2:\"value2\"",
                                            },
                                        },
                                        key: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 15",
                                                        end: "line: 1, column: 19",
                                                        source: "key2",
                                                    },
                                                },
                                                name: "key2",
                                            },
                                        ),
                                        separator: [],
                                        value: Some(
                                            StringLit(
                                                StringLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 20",
                                                            end: "line: 1, column: 28",
                                                            source: "\"value2\"",
                                                        },
                                                    },
                                                    value: "value2",
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
                Expr(
                    ExprStmt {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 2, column: 4",
                                end: "line: 2, column: 10",
                                source: "m.key1",
                            },
                        },
                        expression: Member(
                            MemberExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 2, column: 4",
                                        end: "line: 2, column: 10",
                                        source: "m.key1",
                                    },
                                },
                                object: Identifier(
                                    Identifier {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 2, column: 4",
                                                end: "line: 2, column: 5",
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
                                                start: "line: 2, column: 6",
                                                end: "line: 2, column: 10",
                                                source: "key1",
                                            },
                                        },
                                        name: "key1",
                                    },
                                ),
                                rbrack: [],
                            },
                        ),
                    },
                ),
                Expr(
                    ExprStmt {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 3, column: 4",
                                end: "line: 3, column: 13",
                                source: "m[\"key2\"]",
                            },
                        },
                        expression: Member(
                            MemberExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 3, column: 4",
                                        end: "line: 3, column: 13",
                                        source: "m[\"key2\"]",
                                    },
                                },
                                object: Identifier(
                                    Identifier {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 3, column: 4",
                                                end: "line: 3, column: 5",
                                                source: "m",
                                            },
                                        },
                                        name: "m",
                                    },
                                ),
                                lbrack: [],
                                property: StringLit(
                                    StringLit {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 3, column: 6",
                                                end: "line: 3, column: 12",
                                                source: "\"key2\"",
                                            },
                                        },
                                        value: "key2",
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
fn object_with_string_literal_key() {
    let mut p = Parser::new(r#"x = {"a": 10}"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 14",
                    source: "x = {\"a\": 10}",
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
                                source: "x = {\"a\": 10}",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 1",
                                    end: "line: 1, column: 2",
                                    source: "x",
                                },
                            },
                            name: "x",
                        },
                        init: Object(
                            ObjectExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 5",
                                        end: "line: 1, column: 14",
                                        source: "{\"a\": 10}",
                                    },
                                },
                                lbrace: [],
                                with: None,
                                properties: [
                                    Property {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 6",
                                                end: "line: 1, column: 13",
                                                source: "\"a\": 10",
                                            },
                                        },
                                        key: StringLit(
                                            StringLit {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 6",
                                                        end: "line: 1, column: 9",
                                                        source: "\"a\"",
                                                    },
                                                },
                                                value: "a",
                                            },
                                        ),
                                        separator: [],
                                        value: Some(
                                            Integer(
                                                IntegerLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 11",
                                                            end: "line: 1, column: 13",
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
                    },
                ),
            ],
            eof: [],
        }
    "#]]
    .assert_debug_eq(&parsed);
}

#[test]
fn object_with_mixed_keys() {
    let mut p = Parser::new(r#"x = {"a": 10, b: 11}"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 21",
                    source: "x = {\"a\": 10, b: 11}",
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
                                source: "x = {\"a\": 10, b: 11}",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 1",
                                    end: "line: 1, column: 2",
                                    source: "x",
                                },
                            },
                            name: "x",
                        },
                        init: Object(
                            ObjectExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 5",
                                        end: "line: 1, column: 21",
                                        source: "{\"a\": 10, b: 11}",
                                    },
                                },
                                lbrace: [],
                                with: None,
                                properties: [
                                    Property {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 6",
                                                end: "line: 1, column: 13",
                                                source: "\"a\": 10",
                                            },
                                        },
                                        key: StringLit(
                                            StringLit {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 6",
                                                        end: "line: 1, column: 9",
                                                        source: "\"a\"",
                                                    },
                                                },
                                                value: "a",
                                            },
                                        ),
                                        separator: [],
                                        value: Some(
                                            Integer(
                                                IntegerLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 11",
                                                            end: "line: 1, column: 13",
                                                            source: "10",
                                                        },
                                                    },
                                                    value: 10,
                                                },
                                            ),
                                        ),
                                        comma: [],
                                    },
                                    Property {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 15",
                                                end: "line: 1, column: 20",
                                                source: "b: 11",
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
                                                            end: "line: 1, column: 20",
                                                            source: "11",
                                                        },
                                                    },
                                                    value: 11,
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
fn implicit_key_object_literal() {
    let mut p = Parser::new(r#"x = {a, b}"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 11",
                    source: "x = {a, b}",
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
                                source: "x = {a, b}",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 1",
                                    end: "line: 1, column: 2",
                                    source: "x",
                                },
                            },
                            name: "x",
                        },
                        init: Object(
                            ObjectExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 5",
                                        end: "line: 1, column: 11",
                                        source: "{a, b}",
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
                                    Property {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 9",
                                                end: "line: 1, column: 10",
                                                source: "b",
                                            },
                                        },
                                        key: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 9",
                                                        end: "line: 1, column: 10",
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
fn implicit_key_object_literal_error() {
    let mut p = Parser::new(r#"x = {"a", b}"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 13",
                    source: "x = {\"a\", b}",
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
                                end: "line: 1, column: 13",
                                source: "x = {\"a\", b}",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 1",
                                    end: "line: 1, column: 2",
                                    source: "x",
                                },
                            },
                            name: "x",
                        },
                        init: Object(
                            ObjectExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 5",
                                        end: "line: 1, column: 13",
                                        source: "{\"a\", b}",
                                    },
                                },
                                lbrace: [],
                                with: None,
                                properties: [
                                    Property {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 6",
                                                end: "line: 1, column: 9",
                                                source: "\"a\"",
                                            },
                                        },
                                        key: StringLit(
                                            StringLit {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 6",
                                                        end: "line: 1, column: 9",
                                                        source: "\"a\"",
                                                    },
                                                },
                                                value: "a",
                                            },
                                        ),
                                        separator: [],
                                        value: None,
                                        comma: [],
                                    },
                                    Property {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 11",
                                                end: "line: 1, column: 12",
                                                source: "b",
                                            },
                                        },
                                        key: Identifier(
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
fn implicit_and_explicit_keys_object_literal_error() {
    let mut p = Parser::new(r#"x = {a, b:c}"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 13",
                    source: "x = {a, b:c}",
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
                                end: "line: 1, column: 13",
                                source: "x = {a, b:c}",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 1",
                                    end: "line: 1, column: 2",
                                    source: "x",
                                },
                            },
                            name: "x",
                        },
                        init: Object(
                            ObjectExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 5",
                                        end: "line: 1, column: 13",
                                        source: "{a, b:c}",
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
                                    Property {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 9",
                                                end: "line: 1, column: 12",
                                                source: "b:c",
                                            },
                                        },
                                        key: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 9",
                                                        end: "line: 1, column: 10",
                                                        source: "b",
                                                    },
                                                },
                                                name: "b",
                                            },
                                        ),
                                        separator: [],
                                        value: Some(
                                            Identifier(
                                                Identifier {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 11",
                                                            end: "line: 1, column: 12",
                                                            source: "c",
                                                        },
                                                    },
                                                    name: "c",
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
fn object_with() {
    let mut p = Parser::new(r#"{a with b:c, d:e}"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 18",
                    source: "{a with b:c, d:e}",
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
                                source: "{a with b:c, d:e}",
                            },
                        },
                        expression: Object(
                            ObjectExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 18",
                                        source: "{a with b:c, d:e}",
                                    },
                                },
                                lbrace: [],
                                with: Some(
                                    WithSource {
                                        source: Identifier {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 2",
                                                    end: "line: 1, column: 3",
                                                    source: "a",
                                                },
                                            },
                                            name: "a",
                                        },
                                        with: [],
                                    },
                                ),
                                properties: [
                                    Property {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 9",
                                                end: "line: 1, column: 12",
                                                source: "b:c",
                                            },
                                        },
                                        key: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 9",
                                                        end: "line: 1, column: 10",
                                                        source: "b",
                                                    },
                                                },
                                                name: "b",
                                            },
                                        ),
                                        separator: [],
                                        value: Some(
                                            Identifier(
                                                Identifier {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 11",
                                                            end: "line: 1, column: 12",
                                                            source: "c",
                                                        },
                                                    },
                                                    name: "c",
                                                },
                                            ),
                                        ),
                                        comma: [],
                                    },
                                    Property {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 14",
                                                end: "line: 1, column: 17",
                                                source: "d:e",
                                            },
                                        },
                                        key: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 14",
                                                        end: "line: 1, column: 15",
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
                                                            start: "line: 1, column: 16",
                                                            end: "line: 1, column: 17",
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
                    },
                ),
            ],
            eof: [],
        }
    "#]]
    .assert_debug_eq(&parsed);
}

#[test]
fn object_with_implicit_keys() {
    let mut p = Parser::new(r#"{a with b, c}"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 14",
                    source: "{a with b, c}",
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
                                source: "{a with b, c}",
                            },
                        },
                        expression: Object(
                            ObjectExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 14",
                                        source: "{a with b, c}",
                                    },
                                },
                                lbrace: [],
                                with: Some(
                                    WithSource {
                                        source: Identifier {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 2",
                                                    end: "line: 1, column: 3",
                                                    source: "a",
                                                },
                                            },
                                            name: "a",
                                        },
                                        with: [],
                                    },
                                ),
                                properties: [
                                    Property {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 9",
                                                end: "line: 1, column: 10",
                                                source: "b",
                                            },
                                        },
                                        key: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 9",
                                                        end: "line: 1, column: 10",
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
                                    Property {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 12",
                                                end: "line: 1, column: 13",
                                                source: "c",
                                            },
                                        },
                                        key: Identifier(
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
