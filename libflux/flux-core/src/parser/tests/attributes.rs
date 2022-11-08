use super::*;

#[test]
fn parse_attribute_inner_nothing_follows() {
    let mut p = Parser::new(
        r#"@attribute
"#,
    );
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 2, column: 1",
                    source: "@attribute\n",
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
                                end: "line: 1, column: 11",
                                source: "@attribute",
                            },
                        },
                        text: "extra attributes not associated with anything",
                    },
                ),
            ],
            eof: [],
        }
    "#]]
    .assert_debug_eq(&parsed);
}
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
                    end: "line: 3, column: 1",
                    source: "@attribute\nimport \"date\"\n",
                },
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
                    end: "line: 3, column: 1",
                    source: "@attribute\nfoo = \"a\"\n",
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
                    start: "line: 1, column: 1",
                    end: "line: 5, column: 1",
                    source: "\n// Package foo implements foo things.\n@edition(\"2022.1\")\npackage foo\n",
                },
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
fn parse_import_attributes() {
    let mut p = Parser::new(
        r#"
@registry("stdlib")
import "date"
@registry("fluxlang.dev")
import "foo"
@registry("baz")
import "bar"
"#,
    );
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 8, column: 1",
                    source: "\n@registry(\"stdlib\")\nimport \"date\"\n@registry(\"fluxlang.dev\")\nimport \"foo\"\n@registry(\"baz\")\nimport \"bar\"\n",
                },
            },
            name: "",
            metadata: "parser-type=rust",
            package: None,
            imports: [
                ImportDeclaration {
                    base: BaseNode {
                        location: SourceLocation {
                            start: "line: 3, column: 1",
                            end: "line: 3, column: 14",
                            source: "import \"date\"",
                        },
                        attributes: [
                            Attribute {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 2, column: 1",
                                        end: "line: 2, column: 20",
                                        source: "@registry(\"stdlib\")",
                                    },
                                },
                                name: "registry",
                                params: [
                                    AttributeParam {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 2, column: 11",
                                                end: "line: 2, column: 19",
                                                source: "\"stdlib\"",
                                            },
                                        },
                                        value: StringLit(
                                            StringLit {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 2, column: 11",
                                                        end: "line: 2, column: 19",
                                                        source: "\"stdlib\"",
                                                    },
                                                },
                                                value: "stdlib",
                                            },
                                        ),
                                        comma: [],
                                    },
                                ],
                            },
                        ],
                    },
                    alias: None,
                    path: StringLit {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 3, column: 8",
                                end: "line: 3, column: 14",
                                source: "\"date\"",
                            },
                        },
                        value: "date",
                    },
                },
                ImportDeclaration {
                    base: BaseNode {
                        location: SourceLocation {
                            start: "line: 5, column: 1",
                            end: "line: 5, column: 13",
                            source: "import \"foo\"",
                        },
                        attributes: [
                            Attribute {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 4, column: 1",
                                        end: "line: 4, column: 26",
                                        source: "@registry(\"fluxlang.dev\")",
                                    },
                                },
                                name: "registry",
                                params: [
                                    AttributeParam {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 4, column: 11",
                                                end: "line: 4, column: 25",
                                                source: "\"fluxlang.dev\"",
                                            },
                                        },
                                        value: StringLit(
                                            StringLit {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 4, column: 11",
                                                        end: "line: 4, column: 25",
                                                        source: "\"fluxlang.dev\"",
                                                    },
                                                },
                                                value: "fluxlang.dev",
                                            },
                                        ),
                                        comma: [],
                                    },
                                ],
                            },
                        ],
                    },
                    alias: None,
                    path: StringLit {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 5, column: 8",
                                end: "line: 5, column: 13",
                                source: "\"foo\"",
                            },
                        },
                        value: "foo",
                    },
                },
                ImportDeclaration {
                    base: BaseNode {
                        location: SourceLocation {
                            start: "line: 7, column: 1",
                            end: "line: 7, column: 13",
                            source: "import \"bar\"",
                        },
                        attributes: [
                            Attribute {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 6, column: 1",
                                        end: "line: 6, column: 17",
                                        source: "@registry(\"baz\")",
                                    },
                                },
                                name: "registry",
                                params: [
                                    AttributeParam {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 6, column: 11",
                                                end: "line: 6, column: 16",
                                                source: "\"baz\"",
                                            },
                                        },
                                        value: StringLit(
                                            StringLit {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 6, column: 11",
                                                        end: "line: 6, column: 16",
                                                        source: "\"baz\"",
                                                    },
                                                },
                                                value: "baz",
                                            },
                                        ),
                                        comma: [],
                                    },
                                ],
                            },
                        ],
                    },
                    alias: None,
                    path: StringLit {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 7, column: 8",
                                end: "line: 7, column: 13",
                                source: "\"bar\"",
                            },
                        },
                        value: "bar",
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
fn parse_many_attributes() {
    let mut p = Parser::new(
        r#"// Package comments
@mount("fluxlang.dev", "https://fluxlang.dev/api/modules")
@two
@three
@four
package foo

// Comments for import
@registry("fluxlang.dev")
@double
import "date"

// x is one
@deprecated("0.123.0")
x = 1
"#,
    );
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 16, column: 1",
                    source: "// Package comments\n@mount(\"fluxlang.dev\", \"https://fluxlang.dev/api/modules\")\n@two\n@three\n@four\npackage foo\n\n// Comments for import\n@registry(\"fluxlang.dev\")\n@double\nimport \"date\"\n\n// x is one\n@deprecated(\"0.123.0\")\nx = 1\n",
                },
            },
            name: "",
            metadata: "parser-type=rust",
            package: Some(
                PackageClause {
                    base: BaseNode {
                        location: SourceLocation {
                            start: "line: 6, column: 1",
                            end: "line: 6, column: 12",
                            source: "package foo",
                        },
                        attributes: [
                            Attribute {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 2, column: 1",
                                        end: "line: 2, column: 59",
                                        source: "@mount(\"fluxlang.dev\", \"https://fluxlang.dev/api/modules\")",
                                    },
                                    comments: [
                                        Comment {
                                            text: "// Package comments\n",
                                        },
                                    ],
                                },
                                name: "mount",
                                params: [
                                    AttributeParam {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 2, column: 8",
                                                end: "line: 2, column: 23",
                                                source: "\"fluxlang.dev\",",
                                            },
                                        },
                                        value: StringLit(
                                            StringLit {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 2, column: 8",
                                                        end: "line: 2, column: 22",
                                                        source: "\"fluxlang.dev\"",
                                                    },
                                                },
                                                value: "fluxlang.dev",
                                            },
                                        ),
                                        comma: [],
                                    },
                                    AttributeParam {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 2, column: 24",
                                                end: "line: 2, column: 58",
                                                source: "\"https://fluxlang.dev/api/modules\"",
                                            },
                                        },
                                        value: StringLit(
                                            StringLit {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 2, column: 24",
                                                        end: "line: 2, column: 58",
                                                        source: "\"https://fluxlang.dev/api/modules\"",
                                                    },
                                                },
                                                value: "https://fluxlang.dev/api/modules",
                                            },
                                        ),
                                        comma: [],
                                    },
                                ],
                            },
                            Attribute {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 3, column: 1",
                                        end: "line: 3, column: 5",
                                        source: "@two",
                                    },
                                },
                                name: "two",
                                params: [],
                            },
                            Attribute {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 4, column: 1",
                                        end: "line: 4, column: 7",
                                        source: "@three",
                                    },
                                },
                                name: "three",
                                params: [],
                            },
                            Attribute {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 5, column: 1",
                                        end: "line: 5, column: 6",
                                        source: "@four",
                                    },
                                },
                                name: "four",
                                params: [],
                            },
                        ],
                    },
                    name: Identifier {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 6, column: 9",
                                end: "line: 6, column: 12",
                                source: "foo",
                            },
                        },
                        name: "foo",
                    },
                },
            ),
            imports: [
                ImportDeclaration {
                    base: BaseNode {
                        location: SourceLocation {
                            start: "line: 11, column: 1",
                            end: "line: 11, column: 14",
                            source: "import \"date\"",
                        },
                        attributes: [
                            Attribute {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 9, column: 1",
                                        end: "line: 9, column: 26",
                                        source: "@registry(\"fluxlang.dev\")",
                                    },
                                    comments: [
                                        Comment {
                                            text: "// Comments for import\n",
                                        },
                                    ],
                                },
                                name: "registry",
                                params: [
                                    AttributeParam {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 9, column: 11",
                                                end: "line: 9, column: 25",
                                                source: "\"fluxlang.dev\"",
                                            },
                                        },
                                        value: StringLit(
                                            StringLit {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 9, column: 11",
                                                        end: "line: 9, column: 25",
                                                        source: "\"fluxlang.dev\"",
                                                    },
                                                },
                                                value: "fluxlang.dev",
                                            },
                                        ),
                                        comma: [],
                                    },
                                ],
                            },
                            Attribute {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 10, column: 1",
                                        end: "line: 10, column: 8",
                                        source: "@double",
                                    },
                                },
                                name: "double",
                                params: [],
                            },
                        ],
                    },
                    alias: None,
                    path: StringLit {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 11, column: 8",
                                end: "line: 11, column: 14",
                                source: "\"date\"",
                            },
                        },
                        value: "date",
                    },
                },
            ],
            body: [
                Variable(
                    VariableAssgn {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 15, column: 1",
                                end: "line: 15, column: 6",
                                source: "x = 1",
                            },
                            attributes: [
                                Attribute {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 14, column: 1",
                                            end: "line: 14, column: 23",
                                            source: "@deprecated(\"0.123.0\")",
                                        },
                                        comments: [
                                            Comment {
                                                text: "// x is one\n",
                                            },
                                        ],
                                    },
                                    name: "deprecated",
                                    params: [
                                        AttributeParam {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 14, column: 13",
                                                    end: "line: 14, column: 22",
                                                    source: "\"0.123.0\"",
                                                },
                                            },
                                            value: StringLit(
                                                StringLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 14, column: 13",
                                                            end: "line: 14, column: 22",
                                                            source: "\"0.123.0\"",
                                                        },
                                                    },
                                                    value: "0.123.0",
                                                },
                                            ),
                                            comma: [],
                                        },
                                    ],
                                },
                            ],
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 15, column: 1",
                                    end: "line: 15, column: 2",
                                    source: "x",
                                },
                            },
                            name: "x",
                        },
                        init: Integer(
                            IntegerLit {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 15, column: 5",
                                        end: "line: 15, column: 6",
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
                    start: "line: 1, column: 1",
                    end: "line: 5, column: 1",
                    source: "\n// My documentation comment.\n@deprecated\nidentity = (x) => x\n",
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
