use pretty_assertions::assert_eq;

use super::*;
use crate::ast::{self, tests::Locator};

#[test]
fn test_parse_type_expression() {
    let mut p = Parser::new(r#"(a:T, b:T) => T where T: Addable + Divisible"#);
    let parsed = p.parse_type_expression();
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        TypeExpression {
            base: BaseNode {
                location: loc.get(1, 1, 1, 45),
                ..BaseNode::default()
            },
            monotype: MonoType::Function(Box::new(FunctionType {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 16),
                    ..BaseNode::default()
                },
                parameters: vec![
                    ParameterType::Required {
                        base: BaseNode {
                            location: loc.get(1, 2, 1, 5),
                            ..BaseNode::default()
                        },
                        name: Identifier {
                            base: BaseNode {
                                location: loc.get(1, 2, 1, 3),
                                ..BaseNode::default()
                            },
                            name: "a".to_string(),
                        },
                        monotype: MonoType::Tvar(TvarType {
                            base: BaseNode {
                                location: loc.get(1, 4, 1, 5),
                                ..BaseNode::default()
                            },
                            name: Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 4, 1, 5),
                                    ..BaseNode::default()
                                },
                                name: "T".to_string(),
                            },
                        }),
                    },
                    ParameterType::Required {
                        base: BaseNode {
                            location: loc.get(1, 7, 1, 10),
                            ..BaseNode::default()
                        },
                        name: Identifier {
                            base: BaseNode {
                                location: loc.get(1, 7, 1, 8),
                                ..BaseNode::default()
                            },
                            name: "b".to_string(),
                        },
                        monotype: MonoType::Tvar(TvarType {
                            base: BaseNode {
                                location: loc.get(1, 9, 1, 10),
                                ..BaseNode::default()
                            },
                            name: Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 9, 1, 10),
                                    ..BaseNode::default()
                                },
                                name: "T".to_string(),
                            },
                        }),
                    },
                ],
                monotype: MonoType::Tvar(TvarType {
                    base: BaseNode {
                        location: loc.get(1, 15, 1, 16),
                        ..BaseNode::default()
                    },
                    name: Identifier {
                        base: BaseNode {
                            location: loc.get(1, 15, 1, 16),
                            ..BaseNode::default()
                        },
                        name: "T".to_string(),
                    },
                }),
            })),
            constraints: vec![TypeConstraint {
                base: BaseNode {
                    location: loc.get(1, 23, 1, 45),
                    ..BaseNode::default()
                },
                tvar: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 23, 1, 24),
                        ..BaseNode::default()
                    },
                    name: "T".to_string(),
                },
                kinds: vec![
                    Identifier {
                        base: BaseNode {
                            location: loc.get(1, 26, 1, 33),
                            ..BaseNode::default()
                        },
                        name: "Addable".to_string(),
                    },
                    Identifier {
                        base: BaseNode {
                            location: loc.get(1, 36, 1, 45),
                            ..BaseNode::default()
                        },
                        name: "Divisible".to_string(),
                    },
                ],
            }],
        },
    )
}

#[test]
fn test_parse_type_expression_tvar() {
    let mut p = Parser::new(r#"A"#);
    let parsed = p.parse_type_expression();
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        TypeExpression {
            base: BaseNode {
                location: loc.get(1, 1, 1, 2),
                ..BaseNode::default()
            },
            monotype: MonoType::Tvar(TvarType {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 2),
                    ..BaseNode::default()
                },
                name: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 2),
                        ..BaseNode::default()
                    },
                    name: "A".to_string(),
                }
            }),
            constraints: vec![],
        },
    )
}

#[test]
fn test_parse_type_expression_int() {
    let mut p = Parser::new(r#"int"#);
    let parsed = p.parse_type_expression();
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        TypeExpression {
            base: BaseNode {
                location: loc.get(1, 1, 1, 4),
                ..BaseNode::default()
            },
            monotype: MonoType::Basic(NamedType {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 4),
                    ..BaseNode::default()
                },
                name: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 4),
                        ..BaseNode::default()
                    },
                    name: "int".to_string(),
                }
            }),
            constraints: vec![],
        },
    )
}

#[test]
fn test_parse_type_expression_uint() {
    let mut p = Parser::new(r#"uint"#);
    let parsed = p.parse_type_expression();
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        TypeExpression {
            base: BaseNode {
                location: loc.get(1, 1, 1, 5),
                ..BaseNode::default()
            },
            monotype: MonoType::Basic(NamedType {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 5),
                    ..BaseNode::default()
                },
                name: Identifier {
                    name: "uint".to_string(),
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 5),
                        ..BaseNode::default()
                    },
                }
            }),
            constraints: vec![],
        },
    )
}

#[test]
fn test_parse_type_expression_float() {
    let mut p = Parser::new(r#"float"#);
    let parsed = p.parse_type_expression();
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        TypeExpression {
            base: BaseNode {
                location: loc.get(1, 1, 1, 6),
                ..BaseNode::default()
            },
            monotype: MonoType::Basic(NamedType {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 6),
                    ..BaseNode::default()
                },
                name: Identifier {
                    name: "float".to_string(),
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 6),
                        ..BaseNode::default()
                    },
                }
            }),
            constraints: vec![],
        },
    )
}

#[test]
fn test_parse_type_expression_string() {
    let mut p = Parser::new(r#"string"#);
    let parsed = p.parse_type_expression();
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        TypeExpression {
            base: BaseNode {
                location: loc.get(1, 1, 1, 7),
                ..BaseNode::default()
            },
            monotype: MonoType::Basic(NamedType {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 7),
                    ..BaseNode::default()
                },
                name: Identifier {
                    name: "string".to_string(),
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 7),
                        ..BaseNode::default()
                    },
                }
            }),
            constraints: vec![],
        },
    )
}

#[test]
fn test_parse_type_expression_bool() {
    let mut p = Parser::new(r#"bool"#);
    let parsed = p.parse_type_expression();
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        TypeExpression {
            base: BaseNode {
                location: loc.get(1, 1, 1, 5),
                ..BaseNode::default()
            },
            monotype: MonoType::Basic(NamedType {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 5),
                    ..BaseNode::default()
                },
                name: Identifier {
                    name: "bool".to_string(),
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 5),
                        ..BaseNode::default()
                    },
                }
            }),
            constraints: vec![],
        },
    )
}

#[test]
fn test_parse_type_expression_time() {
    let mut p = Parser::new(r#"time"#);
    let parsed = p.parse_type_expression();
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        TypeExpression {
            base: BaseNode {
                location: loc.get(1, 1, 1, 5),
                ..BaseNode::default()
            },
            monotype: MonoType::Basic(NamedType {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 5),
                    ..BaseNode::default()
                },
                name: Identifier {
                    name: "time".to_string(),
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 5),
                        ..BaseNode::default()
                    },
                }
            }),
            constraints: vec![],
        },
    )
}

#[test]
fn test_parse_type_expression_duration() {
    let mut p = Parser::new(r#"duration"#);
    let parsed = p.parse_type_expression();
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        TypeExpression {
            base: BaseNode {
                location: loc.get(1, 1, 1, 9),
                ..BaseNode::default()
            },
            monotype: MonoType::Basic(NamedType {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 9),
                    ..BaseNode::default()
                },
                name: Identifier {
                    name: "duration".to_string(),
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 9),
                        ..BaseNode::default()
                    },
                }
            }),
            constraints: vec![],
        },
    )
}

#[test]
fn test_parse_type_expression_bytes() {
    let mut p = Parser::new(r#"bytes"#);
    let parsed = p.parse_type_expression();
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        TypeExpression {
            base: BaseNode {
                location: loc.get(1, 1, 1, 6),
                ..BaseNode::default()
            },
            monotype: MonoType::Basic(NamedType {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 6),
                    ..BaseNode::default()
                },
                name: Identifier {
                    name: "bytes".to_string(),
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 6),
                        ..BaseNode::default()
                    },
                }
            }),
            constraints: vec![],
        },
    )
}

#[test]
fn test_parse_type_expression_regexp() {
    let mut p = Parser::new(r#"regexp"#);
    let parsed = p.parse_type_expression();
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        TypeExpression {
            base: BaseNode {
                location: loc.get(1, 1, 1, 7),
                ..BaseNode::default()
            },
            monotype: MonoType::Basic(NamedType {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 7),
                    ..BaseNode::default()
                },
                name: Identifier {
                    name: "regexp".to_string(),
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 7),
                        ..BaseNode::default()
                    },
                }
            }),
            constraints: vec![],
        },
    )
}

#[test]
fn test_parse_type_expression_array_int() {
    let mut p = Parser::new(r#"[int]"#);
    let parsed = p.parse_type_expression();
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        TypeExpression {
            base: BaseNode {
                location: loc.get(1, 1, 1, 6),
                ..BaseNode::default()
            },
            monotype: MonoType::Array(Box::new(ArrayType {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 6),
                    ..BaseNode::default()
                },
                element: MonoType::Basic(NamedType {
                    base: BaseNode {
                        location: loc.get(1, 2, 1, 5),
                        ..BaseNode::default()
                    },
                    name: Identifier {
                        base: BaseNode {
                            location: loc.get(1, 2, 1, 5),
                            ..BaseNode::default()
                        },
                        name: "int".to_string(),
                    }
                })
            })),
            constraints: vec![],
        },
    )
}

#[test]
fn test_parse_type_expression_array_string() {
    let mut p = Parser::new(r#"[string]"#);
    let parsed = p.parse_type_expression();
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        TypeExpression {
            base: BaseNode {
                location: loc.get(1, 1, 1, 9),
                ..BaseNode::default()
            },
            monotype: MonoType::Array(Box::new(ArrayType {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 9),
                    ..BaseNode::default()
                },
                element: MonoType::Basic(NamedType {
                    base: BaseNode {
                        location: loc.get(1, 2, 1, 8),
                        ..BaseNode::default()
                    },
                    name: Identifier {
                        base: BaseNode {
                            location: loc.get(1, 2, 1, 8),
                            ..BaseNode::default()
                        },
                        name: "string".to_string(),
                    }
                })
            })),
            constraints: vec![],
        }
    )
}

#[test]
fn test_parse_type_expression_dict() {
    let mut p = Parser::new(r#"[string:int]"#);
    let parsed = p.parse_type_expression();
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        TypeExpression {
            base: BaseNode {
                location: loc.get(1, 1, 1, 13),
                ..BaseNode::default()
            },
            monotype: MonoType::Dict(Box::new(DictType {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 13),
                    ..BaseNode::default()
                },
                key: MonoType::Basic(NamedType {
                    base: BaseNode {
                        location: loc.get(1, 2, 1, 8),
                        ..BaseNode::default()
                    },
                    name: Identifier {
                        base: BaseNode {
                            location: loc.get(1, 2, 1, 8),
                            ..BaseNode::default()
                        },
                        name: "string".to_string(),
                    }
                }),
                val: MonoType::Basic(NamedType {
                    base: BaseNode {
                        location: loc.get(1, 9, 1, 12),
                        ..BaseNode::default()
                    },
                    name: Identifier {
                        base: BaseNode {
                            location: loc.get(1, 9, 1, 12),
                            ..BaseNode::default()
                        },
                        name: "int".to_string(),
                    }
                }),
            })),
            constraints: vec![],
        }
    )
}

#[test]
fn test_parse_record_type_only_properties() {
    let mut p = Parser::new(r#"{a:int, b:uint}"#);
    let parsed = p.parse_record_type();
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        MonoType::Record(RecordType {
            base: BaseNode {
                location: loc.get(1, 1, 1, 16),
                ..BaseNode::default()
            },
            tvar: None,
            properties: vec![
                PropertyType {
                    base: BaseNode {
                        location: loc.get(1, 2, 1, 7),
                        ..BaseNode::default()
                    },
                    name: Identifier {
                        name: "a".to_string(),
                        base: BaseNode {
                            location: loc.get(1, 2, 1, 3),
                            ..BaseNode::default()
                        },
                    }
                    .into(),
                    monotype: MonoType::Basic(NamedType {
                        base: BaseNode {
                            location: loc.get(1, 4, 1, 7),
                            ..BaseNode::default()
                        },
                        name: Identifier {
                            name: "int".to_string(),
                            base: BaseNode {
                                location: loc.get(1, 4, 1, 7),
                                ..BaseNode::default()
                            },
                        }
                    })
                },
                PropertyType {
                    base: BaseNode {
                        location: loc.get(1, 9, 1, 15),
                        ..BaseNode::default()
                    },
                    name: Identifier {
                        name: "b".to_string(),
                        base: BaseNode {
                            location: loc.get(1, 9, 1, 10),
                            ..BaseNode::default()
                        },
                    }
                    .into(),
                    monotype: MonoType::Basic(NamedType {
                        base: BaseNode {
                            location: loc.get(1, 11, 1, 15),
                            ..BaseNode::default()
                        },
                        name: Identifier {
                            name: "uint".to_string(),
                            base: BaseNode {
                                location: loc.get(1, 11, 1, 15),
                                ..BaseNode::default()
                            },
                        }
                    })
                }
            ]
        },)
    )
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
                        file: None,
                        start: Position {
                            line: 1,
                            column: 1,
                        },
                        end: Position {
                            line: 1,
                            column: 18,
                        },
                        source: Some(
                            "{\"a\":int, b:uint}",
                        ),
                    },
                    comments: [],
                    errors: [],
                },
                tvar: None,
                properties: [
                    PropertyType {
                        base: BaseNode {
                            location: SourceLocation {
                                file: None,
                                start: Position {
                                    line: 1,
                                    column: 2,
                                },
                                end: Position {
                                    line: 1,
                                    column: 9,
                                },
                                source: Some(
                                    "\"a\":int",
                                ),
                            },
                            comments: [],
                            errors: [],
                        },
                        name: StringLit(
                            StringLit {
                                base: BaseNode {
                                    location: SourceLocation {
                                        file: None,
                                        start: Position {
                                            line: 1,
                                            column: 2,
                                        },
                                        end: Position {
                                            line: 1,
                                            column: 5,
                                        },
                                        source: Some(
                                            "\"a\"",
                                        ),
                                    },
                                    comments: [],
                                    errors: [],
                                },
                                value: "a",
                            },
                        ),
                        monotype: Basic(
                            NamedType {
                                base: BaseNode {
                                    location: SourceLocation {
                                        file: None,
                                        start: Position {
                                            line: 1,
                                            column: 6,
                                        },
                                        end: Position {
                                            line: 1,
                                            column: 9,
                                        },
                                        source: Some(
                                            "int",
                                        ),
                                    },
                                    comments: [],
                                    errors: [],
                                },
                                name: Identifier {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            file: None,
                                            start: Position {
                                                line: 1,
                                                column: 6,
                                            },
                                            end: Position {
                                                line: 1,
                                                column: 9,
                                            },
                                            source: Some(
                                                "int",
                                            ),
                                        },
                                        comments: [],
                                        errors: [],
                                    },
                                    name: "int",
                                },
                            },
                        ),
                    },
                    PropertyType {
                        base: BaseNode {
                            location: SourceLocation {
                                file: None,
                                start: Position {
                                    line: 1,
                                    column: 11,
                                },
                                end: Position {
                                    line: 1,
                                    column: 17,
                                },
                                source: Some(
                                    "b:uint",
                                ),
                            },
                            comments: [],
                            errors: [],
                        },
                        name: Identifier(
                            Identifier {
                                base: BaseNode {
                                    location: SourceLocation {
                                        file: None,
                                        start: Position {
                                            line: 1,
                                            column: 11,
                                        },
                                        end: Position {
                                            line: 1,
                                            column: 12,
                                        },
                                        source: Some(
                                            "b",
                                        ),
                                    },
                                    comments: [],
                                    errors: [],
                                },
                                name: "b",
                            },
                        ),
                        monotype: Basic(
                            NamedType {
                                base: BaseNode {
                                    location: SourceLocation {
                                        file: None,
                                        start: Position {
                                            line: 1,
                                            column: 13,
                                        },
                                        end: Position {
                                            line: 1,
                                            column: 17,
                                        },
                                        source: Some(
                                            "uint",
                                        ),
                                    },
                                    comments: [],
                                    errors: [],
                                },
                                name: Identifier {
                                    base: BaseNode {
                                        location: SourceLocation {
                                            file: None,
                                            start: Position {
                                                line: 1,
                                                column: 13,
                                            },
                                            end: Position {
                                                line: 1,
                                                column: 17,
                                            },
                                            source: Some(
                                                "uint",
                                            ),
                                        },
                                        comments: [],
                                        errors: [],
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
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        MonoType::Record(RecordType {
            base: BaseNode {
                location: loc.get(1, 1, 1, 9),
                ..BaseNode::default()
            },
            tvar: None,
            properties: vec![PropertyType {
                base: BaseNode {
                    location: loc.get(1, 2, 1, 7),
                    ..BaseNode::default()
                },
                name: Identifier {
                    name: "a".to_string(),
                    base: BaseNode {
                        location: loc.get(1, 2, 1, 3),
                        ..BaseNode::default()
                    },
                }
                .into(),
                monotype: MonoType::Basic(NamedType {
                    base: BaseNode {
                        location: loc.get(1, 4, 1, 7),
                        ..BaseNode::default()
                    },
                    name: Identifier {
                        name: "int".to_string(),
                        base: BaseNode {
                            location: loc.get(1, 4, 1, 7),
                            ..BaseNode::default()
                        },
                    }
                })
            },]
        },)
    )
}

#[test]
fn test_parse_record_type_invalid() {
    let mut p = Parser::new(r#"{a b}"#);
    let parsed = p.parse_record_type();
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        MonoType::Record(RecordType {
            base: BaseNode {
                location: loc.get(1, 1, 1, 5),
                errors: vec!["expected RBRACE, got IDENT".to_string()],
                ..BaseNode::default()
            },
            tvar: None,
            properties: vec![],
        })
    )
}

#[test]
fn test_parse_constraint_one_ident() {
    let mut p = Parser::new(r#"A : date"#);
    let parsed = p.parse_constraints();
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        vec![TypeConstraint {
            base: BaseNode {
                location: loc.get(1, 1, 1, 9),
                ..BaseNode::default()
            },
            tvar: Identifier {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 2),
                    ..BaseNode::default()
                },
                name: "A".to_string(),
            },
            kinds: vec![Identifier {
                base: BaseNode {
                    location: loc.get(1, 5, 1, 9),
                    ..BaseNode::default()
                },
                name: "date".to_string(),
            }]
        }],
    )
}
#[test]
fn test_parse_record_type_blank() {
    let mut p = Parser::new(r#"{}"#);
    let parsed = p.parse_record_type();
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        MonoType::Record(RecordType {
            base: BaseNode {
                location: loc.get(1, 1, 1, 3),
                ..BaseNode::default()
            },
            tvar: None,
            properties: vec![],
        },)
    )
}

#[test]
fn test_parse_type_expression_function_with_no_params() {
    let mut p = Parser::new(r#"() => int"#);
    let parsed = p.parse_type_expression();
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        TypeExpression {
            base: BaseNode {
                location: loc.get(1, 1, 1, 10),
                ..BaseNode::default()
            },
            monotype: MonoType::Function(Box::new(FunctionType {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 10),
                    ..BaseNode::default()
                },

                parameters: vec![],
                monotype: MonoType::Basic(NamedType {
                    base: BaseNode {
                        location: loc.get(1, 7, 1, 10),
                        ..BaseNode::default()
                    },
                    name: Identifier {
                        base: BaseNode {
                            location: loc.get(1, 7, 1, 10),
                            ..BaseNode::default()
                        },
                        name: "int".to_string(),
                    }
                }),
            })),
            constraints: vec![],
        },
    )
}

#[test]
fn test_parse_function_type_trailing_comma() {
    let mut p = Parser::new(r#"(a:int,) => int"#);
    let parsed = p.parse_type_expression();
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        TypeExpression {
            base: BaseNode {
                location: loc.get(1, 1, 1, 16),
                ..BaseNode::default()
            },
            monotype: MonoType::Function(Box::new(FunctionType {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 16),
                    ..BaseNode::default()
                },

                parameters: vec![ParameterType::Required {
                    base: BaseNode {
                        location: loc.get(1, 2, 1, 7),
                        ..BaseNode::default()
                    },
                    name: Identifier {
                        base: BaseNode {
                            location: loc.get(1, 2, 1, 3),
                            ..BaseNode::default()
                        },
                        name: "a".to_string(),
                    },
                    monotype: MonoType::Basic(NamedType {
                        base: BaseNode {
                            location: loc.get(1, 4, 1, 7),
                            ..BaseNode::default()
                        },
                        name: Identifier {
                            base: BaseNode {
                                location: loc.get(1, 4, 1, 7),
                                ..BaseNode::default()
                            },
                            name: "int".to_string(),
                        },
                    }),
                },],
                monotype: MonoType::Basic(NamedType {
                    base: BaseNode {
                        location: loc.get(1, 13, 1, 16),
                        ..BaseNode::default()
                    },
                    name: Identifier {
                        base: BaseNode {
                            location: loc.get(1, 13, 1, 16),
                            ..BaseNode::default()
                        },
                        name: "int".to_string(),
                    }
                }),
            })),
            constraints: vec![],
        },
    )
}

#[test]
fn test_parse_type_expression_function_with_params() {
    let mut p = Parser::new(r#"(A: int, B: uint) => int"#);
    let parsed = p.parse_type_expression();
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        TypeExpression {
            base: BaseNode {
                location: loc.get(1, 1, 1, 25),
                ..BaseNode::default()
            },
            monotype: MonoType::Function(Box::new(FunctionType {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 25),
                    ..BaseNode::default()
                },
                parameters: vec![
                    ParameterType::Required {
                        base: BaseNode {
                            location: loc.get(1, 2, 1, 8),
                            ..BaseNode::default()
                        },
                        name: Identifier {
                            base: BaseNode {
                                location: loc.get(1, 2, 1, 3),
                                ..BaseNode::default()
                            },
                            name: "A".to_string(),
                        },
                        monotype: MonoType::Basic(NamedType {
                            base: BaseNode {
                                location: loc.get(1, 5, 1, 8),
                                ..BaseNode::default()
                            },
                            name: Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 5, 1, 8),
                                    ..BaseNode::default()
                                },
                                name: "int".to_string(),
                            },
                        }),
                    },
                    ParameterType::Required {
                        base: BaseNode {
                            location: loc.get(1, 10, 1, 17),
                            ..BaseNode::default()
                        },
                        name: Identifier {
                            base: BaseNode {
                                location: loc.get(1, 10, 1, 11),
                                ..BaseNode::default()
                            },
                            name: "B".to_string(),
                        },
                        monotype: MonoType::Basic(NamedType {
                            base: BaseNode {
                                location: loc.get(1, 13, 1, 17),
                                ..BaseNode::default()
                            },
                            name: Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 13, 1, 17),
                                    ..BaseNode::default()
                                },
                                name: "uint".to_string(),
                            },
                        }),
                    }
                ],
                monotype: MonoType::Basic(NamedType {
                    base: BaseNode {
                        location: loc.get(1, 22, 1, 25),
                        ..BaseNode::default()
                    },
                    name: Identifier {
                        base: BaseNode {
                            location: loc.get(1, 22, 1, 25),
                            ..BaseNode::default()
                        },
                        name: "int".to_string(),
                    }
                }),
            })),
            constraints: vec![],
        },
    )
}

// optional parameters like (.., ?n: ..) -> ..
#[test]
fn test_parse_type_expression_function_optional_params() {
    let mut p = Parser::new(r#"(?A: int) => int"#);
    let parsed = p.parse_type_expression();
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        TypeExpression {
            base: BaseNode {
                location: loc.get(1, 1, 1, 17),
                ..BaseNode::default()
            },
            monotype: MonoType::Function(Box::new(FunctionType {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 17),
                    ..BaseNode::default()
                },
                parameters: vec![ParameterType::Optional {
                    base: BaseNode {
                        location: loc.get(1, 2, 1, 9),
                        ..BaseNode::default()
                    },
                    name: Identifier {
                        base: BaseNode {
                            location: loc.get(1, 3, 1, 4),
                            ..BaseNode::default()
                        },
                        name: "A".to_string(),
                    },
                    monotype: MonoType::Basic(NamedType {
                        base: BaseNode {
                            location: loc.get(1, 6, 1, 9),
                            ..BaseNode::default()
                        },
                        name: Identifier {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 9),
                                ..BaseNode::default()
                            },
                            name: "int".to_string(),
                        },
                    }),
                    default: None,
                }],
                monotype: MonoType::Basic(NamedType {
                    base: BaseNode {
                        location: loc.get(1, 14, 1, 17),
                        ..BaseNode::default()
                    },
                    name: Identifier {
                        base: BaseNode {
                            location: loc.get(1, 14, 1, 17),
                            ..BaseNode::default()
                        },
                        name: "int".to_string(),
                    }
                }),
            })),
            constraints: vec![],
        },
    )
}

#[test]
fn test_parse_type_expression_function_named_params() {
    let mut p = Parser::new(r#"(<-A: int) => int"#);
    let parsed = p.parse_type_expression();
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        TypeExpression {
            base: BaseNode {
                location: loc.get(1, 1, 1, 18),
                ..BaseNode::default()
            },
            monotype: MonoType::Function(Box::new(FunctionType {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 18),
                    ..BaseNode::default()
                },
                parameters: vec![ParameterType::Pipe {
                    base: BaseNode {
                        location: loc.get(1, 2, 1, 10),
                        ..BaseNode::default()
                    },
                    name: Some(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 4, 1, 5),
                            ..BaseNode::default()
                        },
                        name: "A".to_string(),
                    }),
                    monotype: MonoType::Basic(NamedType {
                        base: BaseNode {
                            location: loc.get(1, 7, 1, 10),
                            ..BaseNode::default()
                        },
                        name: Identifier {
                            base: BaseNode {
                                location: loc.get(1, 7, 1, 10),
                                ..BaseNode::default()
                            },
                            name: "int".to_string(),
                        },
                    }),
                }],
                monotype: MonoType::Basic(NamedType {
                    base: BaseNode {
                        location: loc.get(1, 15, 1, 18),
                        ..BaseNode::default()
                    },
                    name: Identifier {
                        base: BaseNode {
                            location: loc.get(1, 15, 1, 18),
                            ..BaseNode::default()
                        },
                        name: "int".to_string(),
                    }
                }),
            })),
            constraints: vec![],
        },
    )
}

#[test]
fn test_parse_type_expression_function_unnamed_params() {
    let mut p = Parser::new(r#"(<- : int) => int"#);
    let parsed = p.parse_type_expression();
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        TypeExpression {
            base: BaseNode {
                location: loc.get(1, 1, 1, 18),
                ..BaseNode::default()
            },
            monotype: MonoType::Function(Box::new(FunctionType {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 18),
                    ..BaseNode::default()
                },
                parameters: vec![ParameterType::Pipe {
                    base: BaseNode {
                        location: loc.get(1, 2, 1, 10),
                        ..BaseNode::default()
                    },
                    name: None,
                    monotype: MonoType::Basic(NamedType {
                        base: BaseNode {
                            location: loc.get(1, 7, 1, 10),
                            ..BaseNode::default()
                        },
                        name: Identifier {
                            base: BaseNode {
                                location: loc.get(1, 7, 1, 10),
                                ..BaseNode::default()
                            },
                            name: "int".to_string(),
                        },
                    }),
                }],
                monotype: MonoType::Basic(NamedType {
                    base: BaseNode {
                        location: loc.get(1, 15, 1, 18),
                        ..BaseNode::default()
                    },
                    name: Identifier {
                        base: BaseNode {
                            location: loc.get(1, 15, 1, 18),
                            ..BaseNode::default()
                        },
                        name: "int".to_string(),
                    }
                }),
            })),
            constraints: vec![],
        },
    )
}

#[test]
fn test_parse_constraint_two_ident() {
    let mut p = Parser::new(r#"A: Addable + Subtractable"#);
    let parsed = p.parse_constraints();
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        vec![TypeConstraint {
            base: BaseNode {
                location: loc.get(1, 1, 1, 26),
                ..BaseNode::default()
            },
            tvar: Identifier {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 2),
                    ..BaseNode::default()
                },
                name: "A".to_string(),
            },
            kinds: vec![
                Identifier {
                    base: BaseNode {
                        location: loc.get(1, 4, 1, 11),
                        ..BaseNode::default()
                    },
                    name: "Addable".to_string(),
                },
                Identifier {
                    base: BaseNode {
                        location: loc.get(1, 14, 1, 26),
                        ..BaseNode::default()
                    },
                    name: "Subtractable".to_string(),
                }
            ]
        }],
    )
}

#[test]
fn test_parse_constraint_two_con() {
    let mut p = Parser::new(r#"A: Addable, B: Subtractable"#);
    let parsed = p.parse_constraints();
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        vec![
            TypeConstraint {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 11),
                    ..BaseNode::default()
                },
                tvar: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 2),
                        ..BaseNode::default()
                    },
                    name: "A".to_string(),
                },
                kinds: vec![Identifier {
                    base: BaseNode {
                        location: loc.get(1, 4, 1, 11),
                        ..BaseNode::default()
                    },
                    name: "Addable".to_string(),
                }]
            },
            TypeConstraint {
                base: BaseNode {
                    location: loc.get(1, 13, 1, 28),
                    ..BaseNode::default()
                },
                tvar: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 13, 1, 14),
                        ..BaseNode::default()
                    },
                    name: "B".to_string(),
                },
                kinds: vec![Identifier {
                    base: BaseNode {
                        location: loc.get(1, 16, 1, 28),
                        ..BaseNode::default()
                    },
                    name: "Subtractable".to_string(),
                }]
            }
        ],
    )
}

#[test]
fn test_parse_record_type_tvar_properties() {
    let mut p = Parser::new(r#"{A with a:int, b:uint}"#);
    let parsed = p.parse_record_type();
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        MonoType::Record(RecordType {
            base: BaseNode {
                location: loc.get(1, 1, 1, 23),
                ..BaseNode::default()
            },
            tvar: Some(Identifier {
                base: BaseNode {
                    location: loc.get(1, 2, 1, 3),
                    ..BaseNode::default()
                },
                name: "A".to_string(),
            }),
            properties: vec![
                PropertyType {
                    base: BaseNode {
                        location: loc.get(1, 9, 1, 14),
                        ..BaseNode::default()
                    },
                    name: Identifier {
                        name: "a".to_string(),
                        base: BaseNode {
                            location: loc.get(1, 9, 1, 10),
                            ..BaseNode::default()
                        },
                    }
                    .into(),
                    monotype: MonoType::Basic(NamedType {
                        base: BaseNode {
                            location: loc.get(1, 11, 1, 14),
                            ..BaseNode::default()
                        },
                        name: Identifier {
                            name: "int".to_string(),
                            base: BaseNode {
                                location: loc.get(1, 11, 1, 14),
                                ..BaseNode::default()
                            },
                        }
                    })
                },
                PropertyType {
                    base: BaseNode {
                        location: loc.get(1, 16, 1, 22),
                        ..BaseNode::default()
                    },
                    name: Identifier {
                        name: "b".to_string(),
                        base: BaseNode {
                            location: loc.get(1, 16, 1, 17),
                            ..BaseNode::default()
                        },
                    }
                    .into(),
                    monotype: MonoType::Basic(NamedType {
                        base: BaseNode {
                            location: loc.get(1, 18, 1, 22),
                            ..BaseNode::default()
                        },
                        name: Identifier {
                            name: "uint".to_string(),
                            base: BaseNode {
                                location: loc.get(1, 18, 1, 22),
                                ..BaseNode::default()
                            },
                        }
                    })
                }
            ]
        },)
    )
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
