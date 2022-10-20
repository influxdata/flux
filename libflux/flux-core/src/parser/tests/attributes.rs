use super::*;

#[test]
fn parse_package_attribute() {
    let mut p = Parser::new(
        r#"@attribute
package main"#,
    );
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 2, column: 13",
                    source: "@attribute\npackage main",
                },
                attributes: [
                    Attribute {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 11",
                                source: "@attribute",
                            },
                        },
                        name: "attribute",
                        params: [],
                    },
                ],
            },
            name: "",
            metadata: "parser-type=rust",
            package: Some(
                PackageClause {
                    base: BaseNode {
                        location: SourceLocation {
                            start: "line: 2, column: 1",
                            end: "line: 2, column: 13",
                            source: "package main",
                        },
                    },
                    name: Identifier {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 2, column: 9",
                                end: "line: 2, column: 13",
                                source: "main",
                            },
                        },
                        name: "main",
                    },
                },
            ),
            imports: [],
            body: [],
            eof: [],
        }
    "#]]
    .assert_debug_eq(&parsed);
}

#[test]
fn parse_package_attribute_with_args() {
    let mut p = Parser::new(
        r#"@edition("2022.1")
package main"#,
    );
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 2, column: 13",
                    source: "@edition(\"2022.1\")\npackage main",
                },
                attributes: [
                    Attribute {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 19",
                                source: "@edition(\"2022.1\")",
                            },
                        },
                        name: "edition",
                        params: [
                            AttributeParam {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 10",
                                        end: "line: 1, column: 18",
                                        source: "\"2022.1\"",
                                    },
                                },
                                value: StringLit(
                                    StringLit {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 10",
                                                end: "line: 1, column: 18",
                                                source: "\"2022.1\"",
                                            },
                                        },
                                        value: "2022.1",
                                    },
                                ),
                                comma: [],
                            },
                        ],
                    },
                ],
            },
            name: "",
            metadata: "parser-type=rust",
            package: Some(
                PackageClause {
                    base: BaseNode {
                        location: SourceLocation {
                            start: "line: 2, column: 1",
                            end: "line: 2, column: 13",
                            source: "package main",
                        },
                    },
                    name: Identifier {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 2, column: 9",
                                end: "line: 2, column: 13",
                                source: "main",
                            },
                        },
                        name: "main",
                    },
                },
            ),
            imports: [],
            body: [],
            eof: [],
        }
    "#]]
    .assert_debug_eq(&parsed);
}

#[test]
fn parse_attribute_before_import() {
    let mut p = Parser::new(
        r#"@attribute
import "date"
"#,
    );
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 2, column: 14",
                    source: "@attribute\nimport \"date\"",
                },
                attributes: [
                    Attribute {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 11",
                                source: "@attribute",
                            },
                        },
                        name: "attribute",
                        params: [],
                    },
                ],
            },
            name: "",
            metadata: "parser-type=rust",
            package: None,
            imports: [
                ImportDeclaration {
                    base: BaseNode {
                        location: SourceLocation {
                            start: "line: 2, column: 1",
                            end: "line: 2, column: 14",
                            source: "import \"date\"",
                        },
                    },
                    alias: None,
                    path: StringLit {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 2, column: 8",
                                end: "line: 2, column: 14",
                                source: "\"date\"",
                            },
                        },
                        value: "date",
                    },
                },
            ],
            body: [],
            eof: [],
        }
    "#]]
    .assert_debug_eq(&parsed);
}

#[test]
fn parse_attribute_no_package() {
    let mut p = Parser::new(
        r#"@attribute
foo = "a"
"#,
    );
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 2, column: 10",
                    source: "@attribute\nfoo = \"a\"",
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
                                end: "line: 2, column: 10",
                                source: "foo = \"a\"",
                            },
                            attributes: [
                                Attribute {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 1",
                                            end: "line: 1, column: 11",
                                            source: "@attribute",
                                        },
                                    },
                                    name: "attribute",
                                    params: [],
                                },
                            ],
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 2, column: 1",
                                    end: "line: 2, column: 4",
                                    source: "foo",
                                },
                            },
                            name: "foo",
                        },
                        init: StringLit(
                            StringLit {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 2, column: 7",
                                        end: "line: 2, column: 10",
                                        source: "\"a\"",
                                    },
                                },
                                value: "a",
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
fn parse_attribute_package_comment() {
    let mut p = Parser::new(
        r#"
// Package foo implements foo things.
@edition("2022.1")
package foo
"#,
    );
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 3, column: 1",
                    end: "line: 4, column: 12",
                    source: "@edition(\"2022.1\")\npackage foo",
                },
                attributes: [
                    Attribute {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 3, column: 1",
                                end: "line: 3, column: 19",
                                source: "@edition(\"2022.1\")",
                            },
                            comments: [
                                Comment {
                                    text: "// Package foo implements foo things.\n",
                                },
                            ],
                        },
                        name: "edition",
                        params: [
                            AttributeParam {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 3, column: 10",
                                        end: "line: 3, column: 18",
                                        source: "\"2022.1\"",
                                    },
                                },
                                value: StringLit(
                                    StringLit {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 3, column: 10",
                                                end: "line: 3, column: 18",
                                                source: "\"2022.1\"",
                                            },
                                        },
                                        value: "2022.1",
                                    },
                                ),
                                comma: [],
                            },
                        ],
                    },
                ],
            },
            name: "",
            metadata: "parser-type=rust",
            package: Some(
                PackageClause {
                    base: BaseNode {
                        location: SourceLocation {
                            start: "line: 4, column: 1",
                            end: "line: 4, column: 12",
                            source: "package foo",
                        },
                    },
                    name: Identifier {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 4, column: 9",
                                end: "line: 4, column: 12",
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
    "#]]
    .assert_debug_eq(&parsed);
}

#[test]
fn parse_attribute_doc_comment() {
    let mut p = Parser::new(
        r#"
// My documentation comment.
@deprecated
identity = (x) => x
"#,
    );
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 3, column: 1",
                    end: "line: 4, column: 20",
                    source: "@deprecated\nidentity = (x) => x",
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
                                start: "line: 4, column: 1",
                                end: "line: 4, column: 20",
                                source: "identity = (x) => x",
                            },
                            attributes: [
                                Attribute {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 3, column: 1",
                                            end: "line: 3, column: 12",
                                            source: "@deprecated",
                                        },
                                        comments: [
                                            Comment {
                                                text: "// My documentation comment.\n",
                                            },
                                        ],
                                    },
                                    name: "deprecated",
                                    params: [],
                                },
                            ],
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 4, column: 1",
                                    end: "line: 4, column: 9",
                                    source: "identity",
                                },
                            },
                            name: "identity",
                        },
                        init: Function(
                            FunctionExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 4, column: 12",
                                        end: "line: 4, column: 20",
                                        source: "(x) => x",
                                    },
                                },
                                lparen: [],
                                params: [
                                    Property {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 4, column: 13",
                                                end: "line: 4, column: 14",
                                                source: "x",
                                            },
                                        },
                                        key: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 4, column: 13",
                                                        end: "line: 4, column: 14",
                                                        source: "x",
                                                    },
                                                },
                                                name: "x",
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
                                    Identifier(
                                        Identifier {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 4, column: 19",
                                                    end: "line: 4, column: 20",
                                                    source: "x",
                                                },
                                            },
                                            name: "x",
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
