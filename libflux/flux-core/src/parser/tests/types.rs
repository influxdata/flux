use super::*;
use crate::ast;

fn test_type_expression(source: &str, expect: expect_test::Expect) {
    let mut parser = Parser::new(source);
    let parsed = parser.parse_type_expression();

    expect.assert_debug_eq(&parsed);
}

#[test]
fn test_parse_type_expression() {
    test_type_expression(
        r#"(a:T, b:T) => T where T: Addable + Divisible"#,
        expect![[r#"
            TypeExpression {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 45",
                        source: "(a:T, b:T) => T where T: Addable + Divisible",
                    },
                },
                monotype: Function(
                    FunctionType {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 16",
                                source: "(a:T, b:T) => T",
                            },
                        },
                        parameters: [
                            Required {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 2",
                                        end: "line: 1, column: 5",
                                        source: "a:T",
                                    },
                                },
                                name: Identifier {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 2",
                                            end: "line: 1, column: 3",
                                            source: "a",
                                        },
                                    },
                                    name: "a",
                                },
                                monotype: Tvar(
                                    TvarType {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 4",
                                                end: "line: 1, column: 5",
                                                source: "T",
                                            },
                                        },
                                        name: Identifier {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 4",
                                                    end: "line: 1, column: 5",
                                                    source: "T",
                                                },
                                            },
                                            name: "T",
                                        },
                                    },
                                ),
                            },
                            Required {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 7",
                                        end: "line: 1, column: 10",
                                        source: "b:T",
                                    },
                                },
                                name: Identifier {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 7",
                                            end: "line: 1, column: 8",
                                            source: "b",
                                        },
                                    },
                                    name: "b",
                                },
                                monotype: Tvar(
                                    TvarType {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 9",
                                                end: "line: 1, column: 10",
                                                source: "T",
                                            },
                                        },
                                        name: Identifier {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 9",
                                                    end: "line: 1, column: 10",
                                                    source: "T",
                                                },
                                            },
                                            name: "T",
                                        },
                                    },
                                ),
                            },
                        ],
                        monotype: Tvar(
                            TvarType {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 15",
                                        end: "line: 1, column: 16",
                                        source: "T",
                                    },
                                },
                                name: Identifier {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 15",
                                            end: "line: 1, column: 16",
                                            source: "T",
                                        },
                                    },
                                    name: "T",
                                },
                            },
                        ),
                    },
                ),
                constraints: [
                    TypeConstraint {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 23",
                                end: "line: 1, column: 45",
                                source: "T: Addable + Divisible",
                            },
                        },
                        tvar: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 23",
                                    end: "line: 1, column: 24",
                                    source: "T",
                                },
                            },
                            name: "T",
                        },
                        kinds: [
                            Identifier {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 26",
                                        end: "line: 1, column: 33",
                                        source: "Addable",
                                    },
                                },
                                name: "Addable",
                            },
                            Identifier {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 36",
                                        end: "line: 1, column: 45",
                                        source: "Divisible",
                                    },
                                },
                                name: "Divisible",
                            },
                        ],
                    },
                ],
            }
        "#]],
    );
}

#[test]
fn test_parse_type_expression_tvar() {
    test_type_expression(r#"A"#, expect![[r#"
        TypeExpression {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 2",
                    source: "A",
                },
            },
            monotype: Tvar(
                TvarType {
                    base: BaseNode {
                        location: SourceLocation {
                            start: "line: 1, column: 1",
                            end: "line: 1, column: 2",
                            source: "A",
                        },
                    },
                    name: Identifier {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 2",
                                source: "A",
                            },
                        },
                        name: "A",
                    },
                },
            ),
            constraints: [],
        }
    "#]]);
}

#[test]
fn test_parse_type_expression_int() {
    test_type_expression(r#"int"#, expect![[r#"
        TypeExpression {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 4",
                    source: "int",
                },
            },
            monotype: Basic(
                NamedType {
                    base: BaseNode {
                        location: SourceLocation {
                            start: "line: 1, column: 1",
                            end: "line: 1, column: 4",
                            source: "int",
                        },
                    },
                    name: Identifier {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 4",
                                source: "int",
                            },
                        },
                        name: "int",
                    },
                },
            ),
            constraints: [],
        }
    "#]]);
}

#[test]
fn test_parse_type_expression_uint() {
    test_type_expression(r#"uint"#, expect![[r#"
        TypeExpression {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 5",
                    source: "uint",
                },
            },
            monotype: Basic(
                NamedType {
                    base: BaseNode {
                        location: SourceLocation {
                            start: "line: 1, column: 1",
                            end: "line: 1, column: 5",
                            source: "uint",
                        },
                    },
                    name: Identifier {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 5",
                                source: "uint",
                            },
                        },
                        name: "uint",
                    },
                },
            ),
            constraints: [],
        }
    "#]]);
}

#[test]
fn test_parse_type_expression_float() {
    test_type_expression(r#"float"#, expect![[r#"
        TypeExpression {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 6",
                    source: "float",
                },
            },
            monotype: Basic(
                NamedType {
                    base: BaseNode {
                        location: SourceLocation {
                            start: "line: 1, column: 1",
                            end: "line: 1, column: 6",
                            source: "float",
                        },
                    },
                    name: Identifier {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 6",
                                source: "float",
                            },
                        },
                        name: "float",
                    },
                },
            ),
            constraints: [],
        }
    "#]]);
}

#[test]
fn test_parse_type_expression_string() {
    test_type_expression(r#"string"#, expect![[r#"
        TypeExpression {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 7",
                    source: "string",
                },
            },
            monotype: Basic(
                NamedType {
                    base: BaseNode {
                        location: SourceLocation {
                            start: "line: 1, column: 1",
                            end: "line: 1, column: 7",
                            source: "string",
                        },
                    },
                    name: Identifier {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 7",
                                source: "string",
                            },
                        },
                        name: "string",
                    },
                },
            ),
            constraints: [],
        }
    "#]]);
}

#[test]
fn test_parse_type_expression_bool() {
    test_type_expression(r#"bool"#, expect![[r#"
        TypeExpression {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 5",
                    source: "bool",
                },
            },
            monotype: Basic(
                NamedType {
                    base: BaseNode {
                        location: SourceLocation {
                            start: "line: 1, column: 1",
                            end: "line: 1, column: 5",
                            source: "bool",
                        },
                    },
                    name: Identifier {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 5",
                                source: "bool",
                            },
                        },
                        name: "bool",
                    },
                },
            ),
            constraints: [],
        }
    "#]]);
}

#[test]
fn test_parse_type_expression_time() {
    test_type_expression(r#"time"#, expect![[r#"
        TypeExpression {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 5",
                    source: "time",
                },
            },
            monotype: Basic(
                NamedType {
                    base: BaseNode {
                        location: SourceLocation {
                            start: "line: 1, column: 1",
                            end: "line: 1, column: 5",
                            source: "time",
                        },
                    },
                    name: Identifier {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 5",
                                source: "time",
                            },
                        },
                        name: "time",
                    },
                },
            ),
            constraints: [],
        }
    "#]]);
}

#[test]
fn test_parse_type_expression_duration() {
    test_type_expression(r#"duration"#, expect![[r#"
        TypeExpression {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 9",
                    source: "duration",
                },
            },
            monotype: Basic(
                NamedType {
                    base: BaseNode {
                        location: SourceLocation {
                            start: "line: 1, column: 1",
                            end: "line: 1, column: 9",
                            source: "duration",
                        },
                    },
                    name: Identifier {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 9",
                                source: "duration",
                            },
                        },
                        name: "duration",
                    },
                },
            ),
            constraints: [],
        }
    "#]]);
}

#[test]
fn test_parse_type_expression_bytes() {
    test_type_expression(r#"bytes"#, expect![[r#"
        TypeExpression {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 6",
                    source: "bytes",
                },
            },
            monotype: Basic(
                NamedType {
                    base: BaseNode {
                        location: SourceLocation {
                            start: "line: 1, column: 1",
                            end: "line: 1, column: 6",
                            source: "bytes",
                        },
                    },
                    name: Identifier {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 6",
                                source: "bytes",
                            },
                        },
                        name: "bytes",
                    },
                },
            ),
            constraints: [],
        }
    "#]]);
}

#[test]
fn test_parse_type_expression_regexp() {
    test_type_expression(r#"regexp"#, expect![[r#"
        TypeExpression {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 7",
                    source: "regexp",
                },
            },
            monotype: Basic(
                NamedType {
                    base: BaseNode {
                        location: SourceLocation {
                            start: "line: 1, column: 1",
                            end: "line: 1, column: 7",
                            source: "regexp",
                        },
                    },
                    name: Identifier {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 7",
                                source: "regexp",
                            },
                        },
                        name: "regexp",
                    },
                },
            ),
            constraints: [],
        }
    "#]]);
}

#[test]
fn test_parse_type_expression_array_int() {
    test_type_expression(r#"[int]"#, expect![[r#"
        TypeExpression {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 6",
                    source: "[int]",
                },
            },
            monotype: Array(
                ArrayType {
                    base: BaseNode {
                        location: SourceLocation {
                            start: "line: 1, column: 1",
                            end: "line: 1, column: 6",
                            source: "[int]",
                        },
                    },
                    element: Basic(
                        NamedType {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 2",
                                    end: "line: 1, column: 5",
                                    source: "int",
                                },
                            },
                            name: Identifier {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 2",
                                        end: "line: 1, column: 5",
                                        source: "int",
                                    },
                                },
                                name: "int",
                            },
                        },
                    ),
                },
            ),
            constraints: [],
        }
    "#]]);
}

#[test]
fn test_parse_type_expression_array_string() {
    test_type_expression(r#"[string]"#, expect![[r#"
        TypeExpression {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 9",
                    source: "[string]",
                },
            },
            monotype: Array(
                ArrayType {
                    base: BaseNode {
                        location: SourceLocation {
                            start: "line: 1, column: 1",
                            end: "line: 1, column: 9",
                            source: "[string]",
                        },
                    },
                    element: Basic(
                        NamedType {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 2",
                                    end: "line: 1, column: 8",
                                    source: "string",
                                },
                            },
                            name: Identifier {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 2",
                                        end: "line: 1, column: 8",
                                        source: "string",
                                    },
                                },
                                name: "string",
                            },
                        },
                    ),
                },
            ),
            constraints: [],
        }
    "#]]);
}

#[test]
fn test_parse_type_expression_dict() {
    test_type_expression(r#"[string:int]"#, expect![[r#"
        TypeExpression {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 13",
                    source: "[string:int]",
                },
            },
            monotype: Dict(
                DictType {
                    base: BaseNode {
                        location: SourceLocation {
                            start: "line: 1, column: 1",
                            end: "line: 1, column: 13",
                            source: "[string:int]",
                        },
                    },
                    key: Basic(
                        NamedType {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 2",
                                    end: "line: 1, column: 8",
                                    source: "string",
                                },
                            },
                            name: Identifier {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 2",
                                        end: "line: 1, column: 8",
                                        source: "string",
                                    },
                                },
                                name: "string",
                            },
                        },
                    ),
                    val: Basic(
                        NamedType {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 9",
                                    end: "line: 1, column: 12",
                                    source: "int",
                                },
                            },
                            name: Identifier {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 9",
                                        end: "line: 1, column: 12",
                                        source: "int",
                                    },
                                },
                                name: "int",
                            },
                        },
                    ),
                },
            ),
            constraints: [],
        }
    "#]]);
}

#[test]
fn test_parse_record_type_only_properties() {
    let mut p = Parser::new(r#"{a:int, b:uint}"#);
    let parsed = p.parse_record_type();
    expect![[r#"
        Record(
            RecordType {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 16",
                        source: "{a:int, b:uint}",
                    },
                },
                tvar: None,
                properties: [
                    PropertyType {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 2",
                                end: "line: 1, column: 7",
                                source: "a:int",
                            },
                        },
                        name: Identifier(
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
                        monotype: Basic(
                            NamedType {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 4",
                                        end: "line: 1, column: 7",
                                        source: "int",
                                    },
                                },
                                name: Identifier {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 4",
                                            end: "line: 1, column: 7",
                                            source: "int",
                                        },
                                    },
                                    name: "int",
                                },
                            },
                        ),
                    },
                    PropertyType {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 9",
                                end: "line: 1, column: 15",
                                source: "b:uint",
                            },
                        },
                        name: Identifier(
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
                        monotype: Basic(
                            NamedType {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 11",
                                        end: "line: 1, column: 15",
                                        source: "uint",
                                    },
                                },
                                name: Identifier {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 11",
                                            end: "line: 1, column: 15",
                                            source: "uint",
                                        },
                                    },
                                    name: "uint",
                                },
                            },
                        ),
                    },
                ],
            },
        )
    "#]].assert_debug_eq(&parsed);
}

#[test]
fn test_parse_record_type_string_literal_property() {
    let mut p = Parser::new(r#"{"a":int, b:uint}"#);
    let parsed = p.parse_record_type();
    expect_test::expect![[r#"
        Record(
            RecordType {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 18",
                        source: "{\"a\":int, b:uint}",
                    },
                },
                tvar: None,
                properties: [
                    PropertyType {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 2",
                                end: "line: 1, column: 9",
                                source: "\"a\":int",
                            },
                        },
                        name: StringLit(
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
                        monotype: Basic(
                            NamedType {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 6",
                                        end: "line: 1, column: 9",
                                        source: "int",
                                    },
                                },
                                name: Identifier {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 6",
                                            end: "line: 1, column: 9",
                                            source: "int",
                                        },
                                    },
                                    name: "int",
                                },
                            },
                        ),
                    },
                    PropertyType {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 11",
                                end: "line: 1, column: 17",
                                source: "b:uint",
                            },
                        },
                        name: Identifier(
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
                        monotype: Basic(
                            NamedType {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 13",
                                        end: "line: 1, column: 17",
                                        source: "uint",
                                    },
                                },
                                name: Identifier {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 13",
                                            end: "line: 1, column: 17",
                                            source: "uint",
                                        },
                                    },
                                    name: "uint",
                                },
                            },
                        ),
                    },
                ],
            },
        )
    "#]]
    .assert_debug_eq(&parsed);
}

#[test]
fn test_parse_record_type_trailing_comma() {
    let mut p = Parser::new(r#"{a:int,}"#);
    let parsed = p.parse_record_type();
    expect![[r#"
        Record(
            RecordType {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 9",
                        source: "{a:int,}",
                    },
                },
                tvar: None,
                properties: [
                    PropertyType {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 2",
                                end: "line: 1, column: 7",
                                source: "a:int",
                            },
                        },
                        name: Identifier(
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
                        monotype: Basic(
                            NamedType {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 4",
                                        end: "line: 1, column: 7",
                                        source: "int",
                                    },
                                },
                                name: Identifier {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 4",
                                            end: "line: 1, column: 7",
                                            source: "int",
                                        },
                                    },
                                    name: "int",
                                },
                            },
                        ),
                    },
                ],
            },
        )
    "#]].assert_debug_eq(&parsed);
}

#[test]
fn test_parse_record_type_invalid() {
    let mut p = Parser::new(r#"{a b}"#);
    let parsed = p.parse_record_type();
    expect![[r#"
        Record(
            RecordType {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 5",
                        source: "{a b",
                    },
                    errors: [
                        "expected RBRACE, got IDENT",
                    ],
                },
                tvar: None,
                properties: [],
            },
        )
    "#]].assert_debug_eq(&parsed);
}

#[test]
fn test_parse_constraint_one_ident() {
    let mut p = Parser::new(r#"A : date"#);
    let parsed = p.parse_constraints();
    expect![[r#"
        [
            TypeConstraint {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 9",
                        source: "A : date",
                    },
                },
                tvar: Identifier {
                    base: BaseNode {
                        location: SourceLocation {
                            start: "line: 1, column: 1",
                            end: "line: 1, column: 2",
                            source: "A",
                        },
                    },
                    name: "A",
                },
                kinds: [
                    Identifier {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 5",
                                end: "line: 1, column: 9",
                                source: "date",
                            },
                        },
                        name: "date",
                    },
                ],
            },
        ]
    "#]].assert_debug_eq(&parsed);
}
#[test]
fn test_parse_record_type_blank() {
    let mut p = Parser::new(r#"{}"#);
    let parsed = p.parse_record_type();
    expect![[r#"
        Record(
            RecordType {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 3",
                        source: "{}",
                    },
                },
                tvar: None,
                properties: [],
            },
        )
    "#]].assert_debug_eq(&parsed);
}

#[test]
fn test_parse_type_expression_function_with_no_params() {
    test_type_expression(r#"() => int"#, expect![[r#"
        TypeExpression {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 10",
                    source: "() => int",
                },
            },
            monotype: Function(
                FunctionType {
                    base: BaseNode {
                        location: SourceLocation {
                            start: "line: 1, column: 1",
                            end: "line: 1, column: 10",
                            source: "() => int",
                        },
                    },
                    parameters: [],
                    monotype: Basic(
                        NamedType {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 7",
                                    end: "line: 1, column: 10",
                                    source: "int",
                                },
                            },
                            name: Identifier {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 7",
                                        end: "line: 1, column: 10",
                                        source: "int",
                                    },
                                },
                                name: "int",
                            },
                        },
                    ),
                },
            ),
            constraints: [],
        }
    "#]]);
}

#[test]
fn test_parse_function_type_trailing_comma() {
    test_type_expression(r#"(a:int,) => int"#, expect![[r#"
        TypeExpression {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 16",
                    source: "(a:int,) => int",
                },
            },
            monotype: Function(
                FunctionType {
                    base: BaseNode {
                        location: SourceLocation {
                            start: "line: 1, column: 1",
                            end: "line: 1, column: 16",
                            source: "(a:int,) => int",
                        },
                    },
                    parameters: [
                        Required {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 2",
                                    end: "line: 1, column: 7",
                                    source: "a:int",
                                },
                            },
                            name: Identifier {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 2",
                                        end: "line: 1, column: 3",
                                        source: "a",
                                    },
                                },
                                name: "a",
                            },
                            monotype: Basic(
                                NamedType {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 4",
                                            end: "line: 1, column: 7",
                                            source: "int",
                                        },
                                    },
                                    name: Identifier {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 4",
                                                end: "line: 1, column: 7",
                                                source: "int",
                                            },
                                        },
                                        name: "int",
                                    },
                                },
                            ),
                        },
                    ],
                    monotype: Basic(
                        NamedType {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 13",
                                    end: "line: 1, column: 16",
                                    source: "int",
                                },
                            },
                            name: Identifier {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 13",
                                        end: "line: 1, column: 16",
                                        source: "int",
                                    },
                                },
                                name: "int",
                            },
                        },
                    ),
                },
            ),
            constraints: [],
        }
    "#]]);
}

#[test]
fn test_parse_type_expression_function_with_params() {
    test_type_expression(r#"(A: int, B: uint) => int"#, expect![[r#"
        TypeExpression {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 25",
                    source: "(A: int, B: uint) => int",
                },
            },
            monotype: Function(
                FunctionType {
                    base: BaseNode {
                        location: SourceLocation {
                            start: "line: 1, column: 1",
                            end: "line: 1, column: 25",
                            source: "(A: int, B: uint) => int",
                        },
                    },
                    parameters: [
                        Required {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 2",
                                    end: "line: 1, column: 8",
                                    source: "A: int",
                                },
                            },
                            name: Identifier {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 2",
                                        end: "line: 1, column: 3",
                                        source: "A",
                                    },
                                },
                                name: "A",
                            },
                            monotype: Basic(
                                NamedType {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 5",
                                            end: "line: 1, column: 8",
                                            source: "int",
                                        },
                                    },
                                    name: Identifier {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 5",
                                                end: "line: 1, column: 8",
                                                source: "int",
                                            },
                                        },
                                        name: "int",
                                    },
                                },
                            ),
                        },
                        Required {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 10",
                                    end: "line: 1, column: 17",
                                    source: "B: uint",
                                },
                            },
                            name: Identifier {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 10",
                                        end: "line: 1, column: 11",
                                        source: "B",
                                    },
                                },
                                name: "B",
                            },
                            monotype: Basic(
                                NamedType {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 13",
                                            end: "line: 1, column: 17",
                                            source: "uint",
                                        },
                                    },
                                    name: Identifier {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 13",
                                                end: "line: 1, column: 17",
                                                source: "uint",
                                            },
                                        },
                                        name: "uint",
                                    },
                                },
                            ),
                        },
                    ],
                    monotype: Basic(
                        NamedType {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 22",
                                    end: "line: 1, column: 25",
                                    source: "int",
                                },
                            },
                            name: Identifier {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 22",
                                        end: "line: 1, column: 25",
                                        source: "int",
                                    },
                                },
                                name: "int",
                            },
                        },
                    ),
                },
            ),
            constraints: [],
        }
    "#]]);
}

// optional parameters like (.., ?n: ..) -> ..
#[test]
fn test_parse_type_expression_function_optional_params() {
    test_type_expression(r#"(?A: int) => int"#, expect![[r#"
        TypeExpression {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 17",
                    source: "(?A: int) => int",
                },
            },
            monotype: Function(
                FunctionType {
                    base: BaseNode {
                        location: SourceLocation {
                            start: "line: 1, column: 1",
                            end: "line: 1, column: 17",
                            source: "(?A: int) => int",
                        },
                    },
                    parameters: [
                        Optional {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 2",
                                    end: "line: 1, column: 9",
                                    source: "?A: int",
                                },
                            },
                            name: Identifier {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 3",
                                        end: "line: 1, column: 4",
                                        source: "A",
                                    },
                                },
                                name: "A",
                            },
                            monotype: Basic(
                                NamedType {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 6",
                                            end: "line: 1, column: 9",
                                            source: "int",
                                        },
                                    },
                                    name: Identifier {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 6",
                                                end: "line: 1, column: 9",
                                                source: "int",
                                            },
                                        },
                                        name: "int",
                                    },
                                },
                            ),
                            default: None,
                        },
                    ],
                    monotype: Basic(
                        NamedType {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 14",
                                    end: "line: 1, column: 17",
                                    source: "int",
                                },
                            },
                            name: Identifier {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 14",
                                        end: "line: 1, column: 17",
                                        source: "int",
                                    },
                                },
                                name: "int",
                            },
                        },
                    ),
                },
            ),
            constraints: [],
        }
    "#]]);
}

#[test]
fn test_parse_type_expression_function_named_params() {
    test_type_expression(r#"(<-A: int) => int"#, expect![[r#"
        TypeExpression {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 18",
                    source: "(<-A: int) => int",
                },
            },
            monotype: Function(
                FunctionType {
                    base: BaseNode {
                        location: SourceLocation {
                            start: "line: 1, column: 1",
                            end: "line: 1, column: 18",
                            source: "(<-A: int) => int",
                        },
                    },
                    parameters: [
                        Pipe {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 2",
                                    end: "line: 1, column: 10",
                                    source: "<-A: int",
                                },
                            },
                            name: Some(
                                Identifier {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 4",
                                            end: "line: 1, column: 5",
                                            source: "A",
                                        },
                                    },
                                    name: "A",
                                },
                            ),
                            monotype: Basic(
                                NamedType {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 7",
                                            end: "line: 1, column: 10",
                                            source: "int",
                                        },
                                    },
                                    name: Identifier {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 7",
                                                end: "line: 1, column: 10",
                                                source: "int",
                                            },
                                        },
                                        name: "int",
                                    },
                                },
                            ),
                        },
                    ],
                    monotype: Basic(
                        NamedType {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 15",
                                    end: "line: 1, column: 18",
                                    source: "int",
                                },
                            },
                            name: Identifier {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 15",
                                        end: "line: 1, column: 18",
                                        source: "int",
                                    },
                                },
                                name: "int",
                            },
                        },
                    ),
                },
            ),
            constraints: [],
        }
    "#]]);
}

#[test]
fn test_parse_type_expression_function_unnamed_params() {
    test_type_expression(r#"(<- : int) => int"#, expect![[r#"
        TypeExpression {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 18",
                    source: "(<- : int) => int",
                },
            },
            monotype: Function(
                FunctionType {
                    base: BaseNode {
                        location: SourceLocation {
                            start: "line: 1, column: 1",
                            end: "line: 1, column: 18",
                            source: "(<- : int) => int",
                        },
                    },
                    parameters: [
                        Pipe {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 2",
                                    end: "line: 1, column: 10",
                                    source: "<- : int",
                                },
                            },
                            name: None,
                            monotype: Basic(
                                NamedType {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 7",
                                            end: "line: 1, column: 10",
                                            source: "int",
                                        },
                                    },
                                    name: Identifier {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 7",
                                                end: "line: 1, column: 10",
                                                source: "int",
                                            },
                                        },
                                        name: "int",
                                    },
                                },
                            ),
                        },
                    ],
                    monotype: Basic(
                        NamedType {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 15",
                                    end: "line: 1, column: 18",
                                    source: "int",
                                },
                            },
                            name: Identifier {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 15",
                                        end: "line: 1, column: 18",
                                        source: "int",
                                    },
                                },
                                name: "int",
                            },
                        },
                    ),
                },
            ),
            constraints: [],
        }
    "#]]);
}

#[test]
fn test_parse_constraint_two_ident() {
    let mut p = Parser::new(r#"A: Addable + Subtractable"#);
    let parsed = p.parse_constraints();
    expect![[r#"
        [
            TypeConstraint {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 26",
                        source: "A: Addable + Subtractable",
                    },
                },
                tvar: Identifier {
                    base: BaseNode {
                        location: SourceLocation {
                            start: "line: 1, column: 1",
                            end: "line: 1, column: 2",
                            source: "A",
                        },
                    },
                    name: "A",
                },
                kinds: [
                    Identifier {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 4",
                                end: "line: 1, column: 11",
                                source: "Addable",
                            },
                        },
                        name: "Addable",
                    },
                    Identifier {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 14",
                                end: "line: 1, column: 26",
                                source: "Subtractable",
                            },
                        },
                        name: "Subtractable",
                    },
                ],
            },
        ]
    "#]].assert_debug_eq(&parsed);
}

#[test]
fn test_parse_constraint_two_con() {
    let mut p = Parser::new(r#"A: Addable, B: Subtractable"#);
    let parsed = p.parse_constraints();
    expect![[r#"
        [
            TypeConstraint {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 11",
                        source: "A: Addable",
                    },
                },
                tvar: Identifier {
                    base: BaseNode {
                        location: SourceLocation {
                            start: "line: 1, column: 1",
                            end: "line: 1, column: 2",
                            source: "A",
                        },
                    },
                    name: "A",
                },
                kinds: [
                    Identifier {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 4",
                                end: "line: 1, column: 11",
                                source: "Addable",
                            },
                        },
                        name: "Addable",
                    },
                ],
            },
            TypeConstraint {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 13",
                        end: "line: 1, column: 28",
                        source: "B: Subtractable",
                    },
                },
                tvar: Identifier {
                    base: BaseNode {
                        location: SourceLocation {
                            start: "line: 1, column: 13",
                            end: "line: 1, column: 14",
                            source: "B",
                        },
                    },
                    name: "B",
                },
                kinds: [
                    Identifier {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 16",
                                end: "line: 1, column: 28",
                                source: "Subtractable",
                            },
                        },
                        name: "Subtractable",
                    },
                ],
            },
        ]
    "#]].assert_debug_eq(&parsed);
}

#[test]
fn test_parse_record_type_tvar_properties() {
    let mut p = Parser::new(r#"{A with a:int, b:uint}"#);
    let parsed = p.parse_record_type();

    expect![[r#"
        Record(
            RecordType {
                base: BaseNode {
                    location: SourceLocation {
                        start: "line: 1, column: 1",
                        end: "line: 1, column: 23",
                        source: "{A with a:int, b:uint}",
                    },
                },
                tvar: Some(
                    Identifier {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 2",
                                end: "line: 1, column: 3",
                                source: "A",
                            },
                        },
                        name: "A",
                    },
                ),
                properties: [
                    PropertyType {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 9",
                                end: "line: 1, column: 14",
                                source: "a:int",
                            },
                        },
                        name: Identifier(
                            Identifier {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 9",
                                        end: "line: 1, column: 10",
                                        source: "a",
                                    },
                                },
                                name: "a",
                            },
                        ),
                        monotype: Basic(
                            NamedType {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 11",
                                        end: "line: 1, column: 14",
                                        source: "int",
                                    },
                                },
                                name: Identifier {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 11",
                                            end: "line: 1, column: 14",
                                            source: "int",
                                        },
                                    },
                                    name: "int",
                                },
                            },
                        ),
                    },
                    PropertyType {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 16",
                                end: "line: 1, column: 22",
                                source: "b:uint",
                            },
                        },
                        name: Identifier(
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
                        monotype: Basic(
                            NamedType {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 18",
                                        end: "line: 1, column: 22",
                                        source: "uint",
                                    },
                                },
                                name: Identifier {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            start: "line: 1, column: 18",
                                            end: "line: 1, column: 22",
                                            source: "uint",
                                        },
                                    },
                                    name: "uint",
                                },
                            },
                        ),
                    },
                ],
            },
        )
    "#]].assert_debug_eq(&parsed);
}

#[test]
fn test_parse_record_unclosed_error() {
    let mut p = Parser::new(r#"(r:{A with a:int) => int"#);
    let parsed = p.parse_type_expression();
    expect_test::expect![["error @1:4-1:18: expected RBRACE, got RPAREN"]].assert_eq(
        &ast::check::check(ast::walk::Node::TypeExpression(&parsed))
            .unwrap_err()
            .to_string(),
    );
}
