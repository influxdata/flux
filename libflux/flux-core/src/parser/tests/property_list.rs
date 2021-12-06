use super::*;

use crate::locator::Locator;
use pretty_assertions::assert_eq;

#[test]
fn property_list_missing_property() {
    let mut p = Parser::new(r#"o = {a: "a",, b: 7}"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 20),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 20),
                    ..BaseNode::default()
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 2),
                        ..BaseNode::default()
                    },
                    name: "o".to_string()
                },
                init: Expression::Object(Box::new(ObjectExpr {
                    base: BaseNode {
                        location: loc.get(1, 5, 1, 20),
                        ..BaseNode::default()
                    },
                    lbrace: vec![],
                    with: None,
                    properties: vec![
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 12),
                                ..BaseNode::default()
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 7),
                                    ..BaseNode::default()
                                },
                                name: "a".to_string()
                            }),
                            separator: vec![],
                            value: Some(Expression::StringLit(StringLit {
                                base: BaseNode {
                                    location: loc.get(1, 9, 1, 12),
                                    ..BaseNode::default()
                                },
                                value: "a".to_string()
                            })),
                            comma: vec![],
                        },
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 13, 1, 13),
                                errors: vec!["missing property in property list".to_string()],
                                ..BaseNode::default()
                            },
                            key: PropertyKey::StringLit(StringLit {
                                base: BaseNode {
                                    location: loc.get(1, 13, 1, 13),
                                    ..BaseNode::default()
                                },
                                value: "<invalid>".to_string()
                            }),
                            separator: vec![],
                            value: None,
                            comma: vec![],
                        },
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 15, 1, 19),
                                ..BaseNode::default()
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 15, 1, 16),
                                    ..BaseNode::default()
                                },
                                name: "b".to_string()
                            }),
                            separator: vec![],
                            value: Some(Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 18, 1, 19),
                                    ..BaseNode::default()
                                },
                                value: 7
                            })),
                            comma: vec![],
                        }
                    ],
                    rbrace: vec![],
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn property_list_missing_key() {
    let mut p = Parser::new(r#"o = {: "a"}"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
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
                        location: loc.get(1, 1, 1, 2),
                        ..BaseNode::default()
                    },
                    name: "o".to_string()
                },
                init: Expression::Object(Box::new(ObjectExpr {
                    base: BaseNode {
                        location: loc.get(1, 5, 1, 12),
                        ..BaseNode::default()
                    },
                    lbrace: vec![],
                    with: None,
                    properties: vec![Property {
                        base: BaseNode {
                            location: loc.get(1, 6, 1, 11),
                            errors: vec!["missing property key".to_string()],
                            ..BaseNode::default()
                        },
                        key: PropertyKey::StringLit(StringLit {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 6),
                                ..BaseNode::default()
                            },
                            value: "<invalid>".to_string()
                        }),
                        separator: vec![],
                        value: Some(Expression::StringLit(StringLit {
                            base: BaseNode {
                                location: loc.get(1, 8, 1, 11),
                                ..BaseNode::default()
                            },
                            value: "a".to_string()
                        })),
                        comma: vec![],
                    }],
                    rbrace: vec![],
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn property_list_missing_value() {
    let mut p = Parser::new(r#"o = {a:}"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 9),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 9),
                    ..BaseNode::default()
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 2),
                        ..BaseNode::default()
                    },
                    name: "o".to_string()
                },
                init: Expression::Object(Box::new(ObjectExpr {
                    base: BaseNode {
                        location: loc.get(1, 5, 1, 9),
                        ..BaseNode::default()
                    },
                    lbrace: vec![],
                    with: None,
                    properties: vec![Property {
                        base: BaseNode {
                            location: loc.get(1, 6, 1, 7),
                            errors: vec!["missing property value".to_string()],
                            ..BaseNode::default()
                        },
                        key: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 7),
                                ..BaseNode::default()
                            },
                            name: "a".to_string()
                        }),
                        separator: vec![],
                        value: None,
                        comma: vec![],
                    }],
                    rbrace: vec![],
                }))
            }))],
            eof: vec![],
        },
    )
}

// TODO(affo): that error is injected by ast.Check().
#[test]
fn property_list_missing_comma() {
    let mut p = Parser::new(r#"o = {a: "a" b: 30}"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
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
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 19),
                    ..BaseNode::default()
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 2),
                        ..BaseNode::default()
                    },
                    name: "o".to_string()
                },
                init: Expression::Object(Box::new(ObjectExpr {
                    base: BaseNode {
                        location: loc.get(1, 5, 1, 19),
                        ..BaseNode::default()
                    },
                    lbrace: vec![],
                    with: None,
                    properties: vec![
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 14),
                                ..BaseNode::default()
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 7),
                                    ..BaseNode::default()
                                },
                                name: "a".to_string()
                            }),
                            separator: vec![],
                            // TODO(affo): ast.Check would add the error "expected an operator between two expressions".
                            value: Some(Expression::Binary(Box::new(BinaryExpr {
                                base: BaseNode {
                                    location: loc.get(1, 9, 1, 14),
                                    ..BaseNode::default()
                                },
                                operator: Operator::InvalidOperator,
                                left: Expression::StringLit(StringLit {
                                    base: BaseNode {
                                        location: loc.get(1, 9, 1, 12),
                                        ..BaseNode::default()
                                    },
                                    value: "a".to_string()
                                }),
                                right: Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 13, 1, 14),
                                        ..BaseNode::default()
                                    },
                                    name: "b".to_string()
                                })
                            }))),
                            comma: vec![],
                        },
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 14, 1, 18),
                                errors: vec!["missing property key".to_string()],
                                ..BaseNode::default()
                            },
                            key: PropertyKey::StringLit(StringLit {
                                base: BaseNode {
                                    location: loc.get(1, 14, 1, 14),
                                    ..BaseNode::default()
                                },
                                value: "<invalid>".to_string()
                            }),
                            separator: vec![],
                            value: Some(Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 16, 1, 18),
                                    errors: vec![
                                        "expected comma in property list, got COLON".to_string()
                                    ],
                                    ..BaseNode::default()
                                },
                                value: 30
                            })),
                            comma: vec![],
                        }
                    ],
                    rbrace: vec![],
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn property_list_trailing_comma() {
    let mut p = Parser::new(r#"o = {a: "a",}"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 14),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 14),
                    ..BaseNode::default()
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 2),
                        ..BaseNode::default()
                    },
                    name: "o".to_string()
                },
                init: Expression::Object(Box::new(ObjectExpr {
                    base: BaseNode {
                        location: loc.get(1, 5, 1, 14),
                        ..BaseNode::default()
                    },
                    lbrace: vec![],
                    with: None,
                    properties: vec![Property {
                        base: BaseNode {
                            location: loc.get(1, 6, 1, 12),
                            ..BaseNode::default()
                        },
                        key: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 7),
                                ..BaseNode::default()
                            },
                            name: "a".to_string()
                        }),
                        separator: vec![],
                        value: Some(Expression::StringLit(StringLit {
                            base: BaseNode {
                                location: loc.get(1, 9, 1, 12),
                                ..BaseNode::default()
                            },
                            value: "a".to_string()
                        })),
                        comma: vec![],
                    }],
                    rbrace: vec![],
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn property_list_bad_property() {
    let mut p = Parser::new(r#"o = {a: "a", 30, b: 7}"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 23),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 23),
                    ..BaseNode::default()
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 2),
                        ..BaseNode::default()
                    },
                    name: "o".to_string()
                },
                init: Expression::Object(Box::new(ObjectExpr {
                    base: BaseNode {
                        location: loc.get(1, 5, 1, 23),
                        ..BaseNode::default()
                    },
                    lbrace: vec![],
                    with: None,
                    properties: vec![
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 12),
                                ..BaseNode::default()
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 7),
                                    ..BaseNode::default()
                                },
                                name: "a".to_string()
                            }),
                            separator: vec![],
                            value: Some(Expression::StringLit(StringLit {
                                base: BaseNode {
                                    location: loc.get(1, 9, 1, 12),
                                    ..BaseNode::default()
                                },
                                value: "a".to_string()
                            })),
                            comma: vec![],
                        },
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 14, 1, 16),
                                errors: vec![
                                    "unexpected token for property key: INT (30)".to_string()
                                ],
                                ..BaseNode::default()
                            },
                            key: PropertyKey::StringLit(StringLit {
                                base: BaseNode {
                                    location: loc.get(1, 14, 1, 14),
                                    ..BaseNode::default()
                                },
                                value: "<invalid>".to_string()
                            }),
                            separator: vec![],
                            value: None,
                            comma: vec![],
                        },
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 18, 1, 22),
                                ..BaseNode::default()
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 18, 1, 19),
                                    ..BaseNode::default()
                                },
                                name: "b".to_string()
                            }),
                            separator: vec![],
                            value: Some(Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 21, 1, 22),
                                    ..BaseNode::default()
                                },
                                value: 7
                            })),
                            comma: vec![],
                        }
                    ],
                    rbrace: vec![],
                }))
            }))],
            eof: vec![],
        },
    )
}
