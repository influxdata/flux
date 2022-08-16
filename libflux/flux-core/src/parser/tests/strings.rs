use pretty_assertions::assert_eq;

use super::*;
use crate::ast::{
    tests::Locator,
    Expression::{Array, Member},
    StringExprPart::{Interpolated, Text},
};

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
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 19),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 19),
                    ..BaseNode::default()
                },
                expression: Expression::StringExpr(Box::new(StringExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 19),
                        ..BaseNode::default()
                    },
                    parts: vec![
                        StringExprPart::Text(TextPart {
                            base: BaseNode {
                                location: loc.get(1, 2, 1, 10),
                                ..BaseNode::default()
                            },
                            value: "a + b = ".to_string(),
                        }),
                        StringExprPart::Interpolated(InterpolatedPart {
                            base: BaseNode {
                                location: loc.get(1, 10, 1, 18),
                                ..BaseNode::default()
                            },
                            expression: Expression::Binary(Box::new(BinaryExpr {
                                base: BaseNode {
                                    location: loc.get(1, 12, 1, 17),
                                    ..BaseNode::default()
                                },
                                left: Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 12, 1, 13),
                                        ..BaseNode::default()
                                    },
                                    name: "a".to_string(),
                                }),
                                right: Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 16, 1, 17),
                                        ..BaseNode::default()
                                    },
                                    name: "b".to_string(),
                                }),
                                operator: Operator::AdditionOperator,
                            })),
                        }),
                    ],
                })),
            })),],
            eof: vec![],
        },
    )
}

#[test]
fn string_interpolation_array() {
    let mut p = Parser::new(r#"a = ["influx", "test", "InfluxOfflineTimeAlert", "acu:${r.a}"]"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 63),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 63),
                    ..BaseNode::default()
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 2),
                        ..BaseNode::default()
                    },
                    name: "a".to_string(),
                },
                init: Array(Box::new(ArrayExpr {
                    base: BaseNode {
                        location: loc.get(1, 5, 1, 63),
                        ..BaseNode::default()
                    },
                    lbrack: vec![],
                    elements: vec![
                        ArrayItem {
                            expression: Expression::StringLit(StringLit {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 14),
                                    ..BaseNode::default()
                                },
                                value: "influx".to_string(),
                            }),
                            comma: vec![],
                        },
                        ArrayItem {
                            expression: Expression::StringLit(StringLit {
                                base: BaseNode {
                                    location: loc.get(1, 16, 1, 22),
                                    ..BaseNode::default()
                                },
                                value: "test".to_string(),
                            }),
                            comma: vec![],
                        },
                        ArrayItem {
                            expression: Expression::StringLit(StringLit {
                                base: BaseNode {
                                    location: loc.get(1, 24, 1, 48),
                                    ..BaseNode::default()
                                },
                                value: "InfluxOfflineTimeAlert".to_string(),
                            }),
                            comma: vec![],
                        },
                        ArrayItem {
                            expression: Expression::StringExpr(Box::new(StringExpr {
                                base: BaseNode {
                                    location: loc.get(1, 50, 1, 62),
                                    ..BaseNode::default()
                                },
                                parts: vec![
                                    Text(TextPart {
                                        base: BaseNode {
                                            location: loc.get(1, 51, 1, 55),
                                            ..BaseNode::default()
                                        },
                                        value: "acu:".to_string(),
                                    }),
                                    Interpolated(InterpolatedPart {
                                        base: BaseNode {
                                            location: loc.get(1, 55, 1, 61),
                                            ..BaseNode::default()
                                        },
                                        expression: Member(Box::new(MemberExpr {
                                            base: BaseNode {
                                                location: loc.get(1, 57, 1, 60),
                                                ..BaseNode::default()
                                            },
                                            object: Expression::Identifier(Identifier {
                                                base: BaseNode {
                                                    location: loc.get(1, 57, 1, 58),
                                                    ..BaseNode::default()
                                                },
                                                name: "r".to_string(),
                                            }),
                                            lbrack: vec![],
                                            property: PropertyKey::Identifier(Identifier {
                                                base: BaseNode {
                                                    location: loc.get(1, 59, 1, 60),
                                                    ..BaseNode::default()
                                                },
                                                name: "a".to_string(),
                                            }),
                                            rbrack: vec![],
                                        })),
                                    }),
                                ],
                            }),),
                            comma: vec![],
                        },
                    ],
                    rbrack: vec![],
                },)),
            }),),],
            eof: vec![],
        }
    )
}

#[test]
fn string_interpolation_multiple() {
    let mut p = Parser::new(r#""a = ${a} and b = ${b}""#);
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
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 24),
                    ..BaseNode::default()
                },
                expression: Expression::StringExpr(Box::new(StringExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 24),
                        ..BaseNode::default()
                    },
                    parts: vec![
                        StringExprPart::Text(TextPart {
                            base: BaseNode {
                                location: loc.get(1, 2, 1, 6),
                                ..BaseNode::default()
                            },
                            value: "a = ".to_string(),
                        }),
                        StringExprPart::Interpolated(InterpolatedPart {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 10),
                                ..BaseNode::default()
                            },
                            expression: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 8, 1, 9),
                                    ..BaseNode::default()
                                },
                                name: "a".to_string(),
                            }),
                        }),
                        StringExprPart::Text(TextPart {
                            base: BaseNode {
                                location: loc.get(1, 10, 1, 19),
                                ..BaseNode::default()
                            },
                            value: " and b = ".to_string(),
                        }),
                        StringExprPart::Interpolated(InterpolatedPart {
                            base: BaseNode {
                                location: loc.get(1, 19, 1, 23),
                                ..BaseNode::default()
                            },
                            expression: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 21, 1, 22),
                                    ..BaseNode::default()
                                },
                                name: "b".to_string(),
                            }),
                        }),
                    ],
                })),
            })),],
            eof: vec![],
        },
    )
}

#[test]
fn string_interpolation_nested() {
    let mut p = Parser::new(r#""we ${"can ${"add" + "strings"}"} together""#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 44),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 44),
                    ..BaseNode::default()
                },
                expression: Expression::StringExpr(Box::new(StringExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 44),
                        ..BaseNode::default()
                    },
                    parts: vec![
                        StringExprPart::Text(TextPart {
                            base: BaseNode {
                                location: loc.get(1, 2, 1, 5),
                                ..BaseNode::default()
                            },
                            value: "we ".to_string(),
                        }),
                        StringExprPart::Interpolated(InterpolatedPart {
                            base: BaseNode {
                                location: loc.get(1, 5, 1, 34),
                                ..BaseNode::default()
                            },
                            expression: Expression::StringExpr(Box::new(StringExpr {
                                base: BaseNode {
                                    location: loc.get(1, 7, 1, 33),
                                    ..BaseNode::default()
                                },
                                parts: vec![
                                    StringExprPart::Text(TextPart {
                                        base: BaseNode {
                                            location: loc.get(1, 8, 1, 12),
                                            ..BaseNode::default()
                                        },
                                        value: "can ".to_string(),
                                    }),
                                    StringExprPart::Interpolated(InterpolatedPart {
                                        base: BaseNode {
                                            location: loc.get(1, 12, 1, 32),
                                            ..BaseNode::default()
                                        },
                                        expression: Expression::Binary(Box::new(BinaryExpr {
                                            base: BaseNode {
                                                location: loc.get(1, 14, 1, 31),
                                                ..BaseNode::default()
                                            },
                                            left: Expression::StringLit(StringLit {
                                                base: BaseNode {
                                                    location: loc.get(1, 14, 1, 19),
                                                    ..BaseNode::default()
                                                },
                                                value: "add".to_string(),
                                            }),
                                            right: Expression::StringLit(StringLit {
                                                base: BaseNode {
                                                    location: loc.get(1, 22, 1, 31),
                                                    ..BaseNode::default()
                                                },
                                                value: "strings".to_string(),
                                            }),
                                            operator: Operator::AdditionOperator,
                                        })),
                                    }),
                                ],
                            }))
                        }),
                        StringExprPart::Text(TextPart {
                            base: BaseNode {
                                location: loc.get(1, 34, 1, 43),
                                ..BaseNode::default()
                            },
                            value: " together".to_string(),
                        }),
                    ],
                })),
            })),],
            eof: vec![],
        },
    )
}

#[test]
fn string_interp_with_escapes() {
    let mut p = Parser::new(r#""string \"interpolation with ${"escapes"}\"""#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 45),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 45),
                    ..BaseNode::default()
                },
                expression: Expression::StringExpr(Box::new(StringExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 45),
                        ..BaseNode::default()
                    },
                    parts: vec![
                        StringExprPart::Text(TextPart {
                            base: BaseNode {
                                location: loc.get(1, 2, 1, 30),
                                ..BaseNode::default()
                            },
                            value: "string \"interpolation with ".to_string(),
                        }),
                        StringExprPart::Interpolated(InterpolatedPart {
                            base: BaseNode {
                                location: loc.get(1, 30, 1, 42),
                                ..BaseNode::default()
                            },
                            expression: Expression::StringLit(StringLit {
                                base: BaseNode {
                                    location: loc.get(1, 32, 1, 41),
                                    ..BaseNode::default()
                                },
                                value: "escapes".to_string(),
                            }),
                        }),
                        StringExprPart::Text(TextPart {
                            base: BaseNode {
                                location: loc.get(1, 42, 1, 44),
                                ..BaseNode::default()
                            },
                            value: "\"".to_string(),
                        }),
                    ],
                })),
            })),],
            eof: vec![],
        },
    )
}

#[test]
fn bad_string_expression() {
    let mut p = Parser::new(r#"fn = (a) => "${a}"#);
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
                        location: loc.get(1, 1, 1, 3),
                        ..BaseNode::default()
                    },
                    name: "fn".to_string(),
                },
                init: Expression::Function(Box::new(FunctionExpr {
                    base: BaseNode {
                        location: loc.get(1, 6, 1, 18),
                        ..BaseNode::default()
                    },
                    lparen: vec![],
                    params: vec![Property {
                        base: BaseNode {
                            location: loc.get(1, 7, 1, 8),
                            ..BaseNode::default()
                        },
                        key: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 7, 1, 8),
                                ..BaseNode::default()
                            },
                            name: "a".to_string()
                        }),
                        separator: vec![],
                        value: None,
                        comma: vec![],
                    }],
                    rparen: vec![],
                    arrow: vec![],
                    body: FunctionBody::Expr(Expression::StringExpr(Box::new(StringExpr {
                        base: BaseNode {
                            location: loc.get(1, 13, 1, 18),
                            errors: vec![
                                "got unexpected token in string expression @1:18-1:18: EOF"
                                    .to_string()
                            ],
                            ..BaseNode::default()
                        },
                        parts: vec![],
                    }))),
                })),
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn string_with_utf_8() {
    let mut p = Parser::new(r#""日本語""#);
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
                expression: Expression::StringLit(StringLit {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 12),
                        ..BaseNode::default()
                    },
                    value: "日本語".to_string()
                })
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn string_with_byte_values() {
    let mut p = Parser::new(r#""\xe6\x97\xa5\xe6\x9c\xac\xe8\xaa\x9e""#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 39),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 39),
                    ..BaseNode::default()
                },
                expression: Expression::StringLit(StringLit {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 39),
                        ..BaseNode::default()
                    },
                    value: "日本語".to_string()
                })
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn string_with_mixed_values() {
    let mut p = Parser::new(r#""hello 日x本 \xe6\x97\xa5\xe6\x9c\xac\xe8\xaa\x9e \xc2\xb5s""#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 63),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 63),
                    ..BaseNode::default()
                },
                expression: Expression::StringLit(StringLit {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 63),
                        ..BaseNode::default()
                    },
                    value: "hello 日x本 日本語 µs".to_string()
                })
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn string_with_escapes() {
    let mut p = Parser::new(
        r#""newline \n
carriage return \r
horizontal tab \t
double quote \"
backslash \\
""#,
    );
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {location: loc.get(1, 1, 6, 2), .. BaseNode::default() },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {location: loc.get(1, 1, 6, 2), .. BaseNode::default() },
                expression: Expression::StringLit(StringLit {
                    base: BaseNode {location: loc.get(1, 1, 6, 2), .. BaseNode::default() },
                    value: "newline \n\ncarriage return \r\nhorizontal tab \t\ndouble quote \"\nbackslash \\\n".to_string()
                })
            }))],
            eof: vec![],
        },
    )
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
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 4, 8),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 4, 8),
                    ..BaseNode::default()
                },
                expression: Expression::StringLit(StringLit {
                    base: BaseNode {
                        location: loc.get(1, 1, 4, 8),
                        ..BaseNode::default()
                    },
                    value: "\n this is a\nmultiline\nstring".to_string()
                })
            }))],
            eof: vec![],
        },
    )
}
