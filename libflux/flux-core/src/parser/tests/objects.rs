use pretty_assertions::assert_eq;

use super::*;
use crate::ast::tests::Locator;

#[test]
fn map_member_expressions() {
    let mut p = Parser::new(
        r#"m = {key1: 1, key2:"value2"}
			m.key1
			m["key2"]
			"#,
    );
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 3, 13),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 29),
                        ..BaseNode::default()
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            ..BaseNode::default()
                        },
                        name: "m".to_string()
                    },
                    init: Expression::Object(Box::new(ObjectExpr {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 29),
                            ..BaseNode::default()
                        },
                        lbrace: vec![],
                        with: None,
                        properties: vec![
                            Property {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 13),
                                    ..BaseNode::default()
                                },
                                key: PropertyKey::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 6, 1, 10),
                                        ..BaseNode::default()
                                    },
                                    name: "key1".to_string()
                                }),
                                separator: vec![],
                                value: Some(Expression::Integer(IntegerLit {
                                    base: BaseNode {
                                        location: loc.get(1, 12, 1, 13),
                                        ..BaseNode::default()
                                    },
                                    value: 1
                                })),
                                comma: vec![],
                            },
                            Property {
                                base: BaseNode {
                                    location: loc.get(1, 15, 1, 28),
                                    ..BaseNode::default()
                                },
                                key: PropertyKey::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 15, 1, 19),
                                        ..BaseNode::default()
                                    },
                                    name: "key2".to_string()
                                }),
                                separator: vec![],
                                value: Some(Expression::StringLit(StringLit {
                                    base: BaseNode {
                                        location: loc.get(1, 20, 1, 28),
                                        ..BaseNode::default()
                                    },
                                    value: "value2".to_string()
                                })),
                                comma: vec![],
                            }
                        ],
                        rbrace: vec![],
                    }))
                })),
                Statement::Expr(Box::new(ExprStmt {
                    base: BaseNode {
                        location: loc.get(2, 4, 2, 10),
                        ..BaseNode::default()
                    },
                    expression: Expression::Member(Box::new(MemberExpr {
                        base: BaseNode {
                            location: loc.get(2, 4, 2, 10),
                            ..BaseNode::default()
                        },
                        object: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(2, 4, 2, 5),
                                ..BaseNode::default()
                            },
                            name: "m".to_string()
                        }),
                        lbrack: vec![],
                        property: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(2, 6, 2, 10),
                                ..BaseNode::default()
                            },
                            name: "key1".to_string()
                        }),
                        rbrack: vec![],
                    }))
                })),
                Statement::Expr(Box::new(ExprStmt {
                    base: BaseNode {
                        location: loc.get(3, 4, 3, 13),
                        ..BaseNode::default()
                    },
                    expression: Expression::Member(Box::new(MemberExpr {
                        base: BaseNode {
                            location: loc.get(3, 4, 3, 13),
                            ..BaseNode::default()
                        },
                        object: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(3, 4, 3, 5),
                                ..BaseNode::default()
                            },
                            name: "m".to_string()
                        }),
                        lbrack: vec![],
                        property: PropertyKey::StringLit(StringLit {
                            base: BaseNode {
                                location: loc.get(3, 6, 3, 12),
                                ..BaseNode::default()
                            },
                            value: "key2".to_string()
                        }),
                        rbrack: vec![],
                    }))
                }))
            ],
            eof: vec![],
        },
    )
}

#[test]
fn object_with_string_literal_key() {
    let mut p = Parser::new(r#"x = {"a": 10}"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
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
                    name: "x".to_string()
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
                            location: loc.get(1, 6, 1, 13),
                            ..BaseNode::default()
                        },
                        key: PropertyKey::StringLit(StringLit {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 9),
                                ..BaseNode::default()
                            },
                            value: "a".to_string()
                        }),
                        separator: vec![],
                        value: Some(Expression::Integer(IntegerLit {
                            base: BaseNode {
                                location: loc.get(1, 11, 1, 13),
                                ..BaseNode::default()
                            },
                            value: 10
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
fn object_with_mixed_keys() {
    let mut p = Parser::new(r#"x = {"a": 10, b: 11}"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 21),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 21),
                    ..BaseNode::default()
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 2),
                        ..BaseNode::default()
                    },
                    name: "x".to_string()
                },
                init: Expression::Object(Box::new(ObjectExpr {
                    base: BaseNode {
                        location: loc.get(1, 5, 1, 21),
                        ..BaseNode::default()
                    },
                    lbrace: vec![],
                    with: None,
                    properties: vec![
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 13),
                                ..BaseNode::default()
                            },
                            key: PropertyKey::StringLit(StringLit {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 9),
                                    ..BaseNode::default()
                                },
                                value: "a".to_string()
                            }),
                            separator: vec![],
                            value: Some(Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 11, 1, 13),
                                    ..BaseNode::default()
                                },
                                value: 10
                            })),
                            comma: vec![],
                        },
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 15, 1, 20),
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
                                    location: loc.get(1, 18, 1, 20),
                                    ..BaseNode::default()
                                },
                                value: 11
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
fn implicit_key_object_literal() {
    let mut p = Parser::new(r#"x = {a, b}"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 11),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 11),
                    ..BaseNode::default()
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 2),
                        ..BaseNode::default()
                    },
                    name: "x".to_string()
                },
                init: Expression::Object(Box::new(ObjectExpr {
                    base: BaseNode {
                        location: loc.get(1, 5, 1, 11),
                        ..BaseNode::default()
                    },
                    lbrace: vec![],
                    with: None,
                    properties: vec![
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 7),
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
                        },
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 9, 1, 10),
                                ..BaseNode::default()
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 9, 1, 10),
                                    ..BaseNode::default()
                                },
                                name: "b".to_string()
                            }),
                            separator: vec![],
                            value: None,
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

// TODO(affo): that error is injected by ast.Check().
#[test]
fn implicit_key_object_literal_error() {
    let mut p = Parser::new(r#"x = {"a", b}"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 13),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 13),
                    ..BaseNode::default()
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 2),
                        ..BaseNode::default()
                    },
                    name: "x".to_string()
                },
                init: Expression::Object(Box::new(ObjectExpr {
                    base: BaseNode {
                        location: loc.get(1, 5, 1, 13),
                        ..BaseNode::default()
                    },
                    lbrace: vec![],
                    with: None,
                    properties: vec![
                        Property {
                            // TODO(affo): this should error with ast.Check: "string literal key "a" must have a value".
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 9),
                                ..BaseNode::default()
                            },
                            key: PropertyKey::StringLit(StringLit {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 9),
                                    ..BaseNode::default()
                                },
                                value: "a".to_string()
                            }),
                            separator: vec![],
                            value: None,
                            comma: vec![],
                        },
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 11, 1, 12),
                                ..BaseNode::default()
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 11, 1, 12),
                                    ..BaseNode::default()
                                },
                                name: "b".to_string()
                            }),
                            separator: vec![],
                            value: None,
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

// TODO(affo): that error is injected by ast.Check().
#[test]
fn implicit_and_explicit_keys_object_literal_error() {
    let mut p = Parser::new(r#"x = {a, b:c}"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 13),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 13),
                    ..BaseNode::default()
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 2),
                        ..BaseNode::default()
                    },
                    name: "x".to_string()
                },
                init: Expression::Object(Box::new(ObjectExpr {
                    // TODO(affo): this should error in ast.Check(): "cannot mix implicit and explicit properties".
                    base: BaseNode {
                        location: loc.get(1, 5, 1, 13),
                        ..BaseNode::default()
                    },
                    lbrace: vec![],
                    with: None,
                    properties: vec![
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 7),
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
                        },
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 9, 1, 12),
                                ..BaseNode::default()
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 9, 1, 10),
                                    ..BaseNode::default()
                                },
                                name: "b".to_string()
                            }),
                            separator: vec![],
                            value: Some(Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 11, 1, 12),
                                    ..BaseNode::default()
                                },
                                name: "c".to_string()
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
fn object_with() {
    let mut p = Parser::new(r#"{a with b:c, d:e}"#);
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
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 18),
                    ..BaseNode::default()
                },
                expression: Expression::Object(Box::new(ObjectExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 18),
                        ..BaseNode::default()
                    },
                    lbrace: vec![],
                    with: Some(WithSource {
                        source: Identifier {
                            base: BaseNode {
                                location: loc.get(1, 2, 1, 3),
                                ..BaseNode::default()
                            },
                            name: "a".to_string()
                        },
                        with: vec![]
                    }),
                    properties: vec![
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 9, 1, 12),
                                ..BaseNode::default()
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 9, 1, 10),
                                    ..BaseNode::default()
                                },
                                name: "b".to_string()
                            }),
                            separator: vec![],
                            value: Some(Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 11, 1, 12),
                                    ..BaseNode::default()
                                },
                                name: "c".to_string()
                            })),
                            comma: vec![],
                        },
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 14, 1, 17),
                                ..BaseNode::default()
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 14, 1, 15),
                                    ..BaseNode::default()
                                },
                                name: "d".to_string()
                            }),
                            separator: vec![],
                            value: Some(Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 16, 1, 17),
                                    ..BaseNode::default()
                                },
                                name: "e".to_string()
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
fn object_with_implicit_keys() {
    let mut p = Parser::new(r#"{a with b, c}"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
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
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 14),
                    ..BaseNode::default()
                },
                expression: Expression::Object(Box::new(ObjectExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 14),
                        ..BaseNode::default()
                    },
                    lbrace: vec![],
                    with: Some(WithSource {
                        source: Identifier {
                            base: BaseNode {
                                location: loc.get(1, 2, 1, 3),
                                ..BaseNode::default()
                            },
                            name: "a".to_string()
                        },
                        with: vec![],
                    }),
                    properties: vec![
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 9, 1, 10),
                                ..BaseNode::default()
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 9, 1, 10),
                                    ..BaseNode::default()
                                },
                                name: "b".to_string()
                            }),
                            separator: vec![],
                            value: None,
                            comma: vec![],
                        },
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 12, 1, 13),
                                ..BaseNode::default()
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 12, 1, 13),
                                    ..BaseNode::default()
                                },
                                name: "c".to_string()
                            }),
                            separator: vec![],
                            value: None,
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
