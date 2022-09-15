use pretty_assertions::assert_eq;

use super::*;
use crate::ast::tests::Locator;

#[test]
fn regex_literal() {
    let mut p = Parser::new(r#"/.*/"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 5),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 5),
                    ..BaseNode::default()
                },
                expression: Expression::Regexp(RegexpLit {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 5),
                        ..BaseNode::default()
                    },
                    value: ".*".to_string()
                })
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn regex_literal_with_escape_sequence() {
    let mut p = Parser::new(r#"/a\/b\\c\d/"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 12),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 12),
                    ..BaseNode::default()
                },
                expression: Expression::Regexp(RegexpLit {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 12),
                        ..BaseNode::default()
                    },
                    value: r#"a/b\\c\d"#.to_string()
                })
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn regex_literal_with_hex_escape() {
    let mut p = Parser::new(r#"/^\xe6\x97\xa5\xe6\x9c\xac\xe8\xaa\x9e(ZZ)?$/"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 46),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 46),
                    ..BaseNode::default()
                },
                expression: Expression::Regexp(RegexpLit {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 46),
                        ..BaseNode::default()
                    },
                    value: r#"^日本語(ZZ)?$"#.to_string()
                })
            }))],
            eof: vec![],
        },
    )
}
#[test]
fn regex_literal_empty_pattern() {
    let mut p = Parser::new(r#"/(:?)/"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 7),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 7),
                    ..BaseNode::default()
                },
                expression: Expression::Regexp(RegexpLit {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 7),
                        ..BaseNode::default()
                    },
                    value: r#"(:?)"#.to_string()
                })
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn bad_regex_literal() {
    let mut p = Parser::new(r#"/*/"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 4),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 4),
                    ..BaseNode::default()
                },
                expression: Expression::Regexp(RegexpLit {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 4),
                        errors: vec![
                            "regex parse error: * error: repetition operator missing expression"
                                .to_string()
                        ],
                        ..BaseNode::default()
                    },
                    value: "".to_string()
                })
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn duration_literal_all_units() {
    let mut p = Parser::new(r#"dur = 1y3mo2w1d4h1m30s1ms2µs70ns"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 34),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 34),
                    ..BaseNode::default()
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 4),
                        ..BaseNode::default()
                    },
                    name: "dur".to_string()
                },
                init: Expression::Duration(DurationLit {
                    base: BaseNode {
                        location: loc.get(1, 7, 1, 34),
                        ..BaseNode::default()
                    },
                    values: vec![
                        Duration {
                            magnitude: 1,
                            unit: "y".to_string()
                        },
                        Duration {
                            magnitude: 3,
                            unit: "mo".to_string()
                        },
                        Duration {
                            magnitude: 2,
                            unit: "w".to_string()
                        },
                        Duration {
                            magnitude: 1,
                            unit: "d".to_string()
                        },
                        Duration {
                            magnitude: 4,
                            unit: "h".to_string()
                        },
                        Duration {
                            magnitude: 1,
                            unit: "m".to_string()
                        },
                        Duration {
                            magnitude: 30,
                            unit: "s".to_string()
                        },
                        Duration {
                            magnitude: 1,
                            unit: "ms".to_string()
                        },
                        Duration {
                            magnitude: 2,
                            unit: "us".to_string()
                        },
                        Duration {
                            magnitude: 70,
                            unit: "ns".to_string()
                        }
                    ]
                })
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn duration_literal_leading_zero() {
    let mut p = Parser::new(r#"dur = 01y02mo03w04d05h06m07s08ms09µs010ns"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 43),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 43),
                    ..BaseNode::default()
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 4),
                        ..BaseNode::default()
                    },
                    name: "dur".to_string()
                },
                init: Expression::Duration(DurationLit {
                    base: BaseNode {
                        location: loc.get(1, 7, 1, 43),
                        ..BaseNode::default()
                    },
                    values: vec![
                        Duration {
                            magnitude: 1,
                            unit: "y".to_string()
                        },
                        Duration {
                            magnitude: 2,
                            unit: "mo".to_string()
                        },
                        Duration {
                            magnitude: 3,
                            unit: "w".to_string()
                        },
                        Duration {
                            magnitude: 4,
                            unit: "d".to_string()
                        },
                        Duration {
                            magnitude: 5,
                            unit: "h".to_string()
                        },
                        Duration {
                            magnitude: 6,
                            unit: "m".to_string()
                        },
                        Duration {
                            magnitude: 7,
                            unit: "s".to_string()
                        },
                        Duration {
                            magnitude: 8,
                            unit: "ms".to_string()
                        },
                        Duration {
                            magnitude: 9,
                            unit: "us".to_string()
                        },
                        Duration {
                            magnitude: 10,
                            unit: "ns".to_string()
                        }
                    ]
                })
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn duration_literal_months() {
    let mut p = Parser::new(r#"dur = 6mo"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 10),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 10),
                    ..BaseNode::default()
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 4),
                        ..BaseNode::default()
                    },
                    name: "dur".to_string()
                },
                init: Expression::Duration(DurationLit {
                    base: BaseNode {
                        location: loc.get(1, 7, 1, 10),
                        ..BaseNode::default()
                    },
                    values: vec![Duration {
                        magnitude: 6,
                        unit: "mo".to_string()
                    }]
                })
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn duration_literal_milliseconds() {
    let mut p = Parser::new(r#"dur = 500ms"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 12),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 12),
                    ..BaseNode::default()
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 4),
                        ..BaseNode::default()
                    },
                    name: "dur".to_string()
                },
                init: Expression::Duration(DurationLit {
                    base: BaseNode {
                        location: loc.get(1, 7, 1, 12),
                        ..BaseNode::default()
                    },
                    values: vec![Duration {
                        magnitude: 500,
                        unit: "ms".to_string()
                    }]
                })
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn duration_literal_months_minutes_milliseconds() {
    let mut p = Parser::new(r#"dur = 6mo30m500ms"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 18),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 18),
                    ..BaseNode::default()
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 4),
                        ..BaseNode::default()
                    },
                    name: "dur".to_string()
                },
                init: Expression::Duration(DurationLit {
                    base: BaseNode {
                        location: loc.get(1, 7, 1, 18),
                        ..BaseNode::default()
                    },
                    values: vec![
                        Duration {
                            magnitude: 6,
                            unit: "mo".to_string()
                        },
                        Duration {
                            magnitude: 30,
                            unit: "m".to_string()
                        },
                        Duration {
                            magnitude: 500,
                            unit: "ms".to_string()
                        }
                    ]
                })
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn date_literal_in_the_default_location() {
    let mut p = Parser::new(r#"now = 2018-11-29"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 17),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 17),
                    ..BaseNode::default()
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 4),
                        ..BaseNode::default()
                    },
                    name: "now".to_string()
                },
                init: Expression::DateTime(DateTimeLit {
                    base: BaseNode {
                        location: loc.get(1, 7, 1, 17),
                        ..BaseNode::default()
                    },
                    value: chrono::DateTime::parse_from_rfc3339("2018-11-29T00:00:00Z").unwrap()
                })
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn date_time_literal_arg() {
    let mut p = Parser::new(r#"range(start: 2018-11-29T09:00:00Z)"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 35),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 35),
                    ..BaseNode::default()
                },
                expression: Expression::Call(Box::new(CallExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 35),
                        errors: vec![],
                        ..BaseNode::default()
                    },
                    arguments: vec![Expression::Object(Box::new(ObjectExpr {
                        base: BaseNode {
                            location: loc.get(1, 7, 1, 34),
                            ..BaseNode::default()
                        },
                        lbrace: vec![],
                        rbrace: vec![],
                        with: None,
                        properties: vec![Property {
                            base: BaseNode {
                                location: loc.get(1, 7, 1, 34),
                                errors: vec![],
                                ..BaseNode::default()
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 7, 1, 12),
                                    ..BaseNode::default()
                                },
                                name: "start".to_string(),
                            }),
                            comma: vec![],
                            separator: vec![],
                            value: Some(Expression::DateTime(DateTimeLit {
                                base: BaseNode {
                                    location: loc.get(1, 14, 1, 34),
                                    ..BaseNode::default()
                                },
                                value: chrono::DateTime::parse_from_rfc3339("2018-11-29T09:00:00Z")
                                    .unwrap()
                            })),
                        }],
                    }))],
                    callee: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 6),
                            ..BaseNode::default()
                        },
                        name: "range".to_string(),
                    }),
                    lparen: vec![],
                    rparen: vec![],
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn date_time_literal_no_offset_error() {
    let mut p = Parser::new(r#"t = 2018-11-29T09:00:00"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 24),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 24),
                    ..BaseNode::default()
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 2),
                        ..BaseNode::default()
                    },
                    name: "t".to_string(),
                },
                init: Expression::Bad(Box::new(BadExpr {
                    base: BaseNode {
                        location: loc.get(1, 5, 1, 24),
                        ..BaseNode::default()
                    },
                    text: "invalid date time literal, missing time offset".to_string(),
                    expression: None
                })),
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn date_time_literal() {
    let mut p = Parser::new(r#"now = 2018-11-29T09:00:00Z"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 27),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 27),
                    ..BaseNode::default()
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 4),
                        ..BaseNode::default()
                    },
                    name: "now".to_string()
                },
                init: Expression::DateTime(DateTimeLit {
                    base: BaseNode {
                        location: loc.get(1, 7, 1, 27),
                        ..BaseNode::default()
                    },
                    value: chrono::DateTime::parse_from_rfc3339("2018-11-29T09:00:00Z").unwrap()
                })
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn date_time_literal_with_fractional_seconds() {
    let mut p = Parser::new(r#"now = 2018-11-29T09:00:00.100000000Z"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 37),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 37),
                    ..BaseNode::default()
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 4),
                        ..BaseNode::default()
                    },
                    name: "now".to_string()
                },
                init: Expression::DateTime(DateTimeLit {
                    base: BaseNode {
                        location: loc.get(1, 7, 1, 37),
                        ..BaseNode::default()
                    },
                    value: chrono::DateTime::parse_from_rfc3339("2018-11-29T09:00:00.100000000Z")
                        .unwrap()
                })
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn integer_literal_overflow() {
    let mut p = Parser::new(r#"100000000000000000000000000000"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 31),
                .. BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 31),
                    .. BaseNode::default() },
                expression: Expression::Integer(IntegerLit {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 31),
                        errors: vec!["invalid integer literal \"100000000000000000000000000000\": value out of range".to_string()],
                        .. BaseNode::default()
                    },
                    value: 0,
                })
            }))],
            eof: vec![],
        },
    )
}
