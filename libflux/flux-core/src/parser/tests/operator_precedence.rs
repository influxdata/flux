use super::*;

use crate::ast::tests::Locator;
use pretty_assertions::assert_eq;

#[test]
fn binary_operator_precedence() {
    let mut p = Parser::new(r#"a / b - 1.0"#);
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
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 12),
                    ..BaseNode::default()
                },
                expression: Expression::Binary(Box::new(BinaryExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 12),
                        ..BaseNode::default()
                    },
                    operator: Operator::SubtractionOperator,
                    left: Expression::Binary(Box::new(BinaryExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 6),
                            ..BaseNode::default()
                        },
                        operator: Operator::DivisionOperator,
                        left: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 2),
                                ..BaseNode::default()
                            },
                            name: "a".to_string()
                        }),
                        right: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 5, 1, 6),
                                ..BaseNode::default()
                            },
                            name: "b".to_string()
                        })
                    })),
                    right: Expression::Float(FloatLit {
                        base: BaseNode {
                            location: loc.get(1, 9, 1, 12),
                            ..BaseNode::default()
                        },
                        value: 1.0
                    })
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn binary_operator_precedence_literals_only() {
    let mut p = Parser::new(r#"2 / "a" - 1.0"#);
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
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 14),
                    ..BaseNode::default()
                },
                expression: Expression::Binary(Box::new(BinaryExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 14),
                        ..BaseNode::default()
                    },
                    operator: Operator::SubtractionOperator,
                    left: Expression::Binary(Box::new(BinaryExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 8),
                            ..BaseNode::default()
                        },
                        operator: Operator::DivisionOperator,
                        left: Expression::Integer(IntegerLit {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 2),
                                ..BaseNode::default()
                            },
                            value: 2
                        }),
                        right: Expression::StringLit(StringLit {
                            base: BaseNode {
                                location: loc.get(1, 5, 1, 8),
                                ..BaseNode::default()
                            },
                            value: "a".to_string()
                        })
                    })),
                    right: Expression::Float(FloatLit {
                        base: BaseNode {
                            location: loc.get(1, 11, 1, 14),
                            ..BaseNode::default()
                        },
                        value: 1.0
                    })
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn binary_operator_precedence_double_subtraction() {
    let mut p = Parser::new(r#"1 - 2 - 3"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
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
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 10),
                    ..BaseNode::default()
                },
                expression: Expression::Binary(Box::new(BinaryExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 10),
                        ..BaseNode::default()
                    },
                    operator: Operator::SubtractionOperator,
                    left: Expression::Binary(Box::new(BinaryExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 6),
                            ..BaseNode::default()
                        },
                        operator: Operator::SubtractionOperator,
                        left: Expression::Integer(IntegerLit {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 2),
                                ..BaseNode::default()
                            },
                            value: 1
                        }),
                        right: Expression::Integer(IntegerLit {
                            base: BaseNode {
                                location: loc.get(1, 5, 1, 6),
                                ..BaseNode::default()
                            },
                            value: 2
                        })
                    })),
                    right: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 9, 1, 10),
                            ..BaseNode::default()
                        },
                        value: 3
                    })
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn binary_operator_precedence_double_subtraction_with_parens() {
    let mut p = Parser::new(r#"1 - (2 - 3)"#);
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
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 12),
                    ..BaseNode::default()
                },
                expression: Expression::Binary(Box::new(BinaryExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 12),
                        ..BaseNode::default()
                    },
                    operator: Operator::SubtractionOperator,
                    left: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            ..BaseNode::default()
                        },
                        value: 1
                    }),
                    right: Expression::Paren(Box::new(ParenExpr {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 12),
                            ..BaseNode::default()
                        },
                        lparen: vec![],
                        expression: Expression::Binary(Box::new(BinaryExpr {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 11),
                                ..BaseNode::default()
                            },
                            operator: Operator::SubtractionOperator,
                            left: Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 7),
                                    ..BaseNode::default()
                                },
                                value: 2
                            }),
                            right: Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 10, 1, 11),
                                    ..BaseNode::default()
                                },
                                value: 3
                            })
                        })),
                        rparen: vec![],
                    }))
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn binary_operator_precedence_double_sum() {
    let mut p = Parser::new(r#"1 + 2 + 3"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
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
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 10),
                    ..BaseNode::default()
                },
                expression: Expression::Binary(Box::new(BinaryExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 10),
                        ..BaseNode::default()
                    },
                    operator: Operator::AdditionOperator,
                    left: Expression::Binary(Box::new(BinaryExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 6),
                            ..BaseNode::default()
                        },
                        operator: Operator::AdditionOperator,
                        left: Expression::Integer(IntegerLit {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 2),
                                ..BaseNode::default()
                            },
                            value: 1
                        }),
                        right: Expression::Integer(IntegerLit {
                            base: BaseNode {
                                location: loc.get(1, 5, 1, 6),
                                ..BaseNode::default()
                            },
                            value: 2
                        })
                    })),
                    right: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 9, 1, 10),
                            ..BaseNode::default()
                        },
                        value: 3
                    })
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn binary_operator_precedence_exponent() {
    let mut p = Parser::new(r#"5 * 1 ^ 5"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
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
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 10),
                    ..BaseNode::default()
                },
                expression: Expression::Binary(Box::new(BinaryExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 10),
                        ..BaseNode::default()
                    },
                    operator: Operator::MultiplicationOperator,
                    left: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            ..BaseNode::default()
                        },
                        value: 5
                    }),
                    right: Expression::Binary(Box::new(BinaryExpr {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 10),
                            ..BaseNode::default()
                        },
                        operator: Operator::PowerOperator,
                        left: Expression::Integer(IntegerLit {
                            base: BaseNode {
                                location: loc.get(1, 5, 1, 6),
                                ..BaseNode::default()
                            },
                            value: 1
                        }),
                        right: Expression::Integer(IntegerLit {
                            base: BaseNode {
                                location: loc.get(1, 9, 1, 10),
                                ..BaseNode::default()
                            },
                            value: 5
                        })
                    })),
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn binary_operator_precedence_double_sum_with_parens() {
    let mut p = Parser::new(r#"1 + (2 + 3)"#);
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
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 12),
                    ..BaseNode::default()
                },
                expression: Expression::Binary(Box::new(BinaryExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 12),
                        ..BaseNode::default()
                    },
                    operator: Operator::AdditionOperator,
                    left: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            ..BaseNode::default()
                        },
                        value: 1
                    }),
                    right: Expression::Paren(Box::new(ParenExpr {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 12),
                            ..BaseNode::default()
                        },
                        lparen: vec![],
                        expression: Expression::Binary(Box::new(BinaryExpr {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 11),
                                ..BaseNode::default()
                            },
                            operator: Operator::AdditionOperator,
                            left: Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 7),
                                    ..BaseNode::default()
                                },
                                value: 2
                            }),
                            right: Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 10, 1, 11),
                                    ..BaseNode::default()
                                },
                                value: 3
                            })
                        })),
                        rparen: vec![],
                    }))
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn binary_operator_precedence_exponent_with_parens() {
    let mut p = Parser::new(r#"2 ^ (1 + 3)"#);
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
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 12),
                    ..BaseNode::default()
                },
                expression: Expression::Binary(Box::new(BinaryExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 12),
                        ..BaseNode::default()
                    },
                    operator: Operator::PowerOperator,
                    left: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            ..BaseNode::default()
                        },
                        value: 2
                    }),
                    right: Expression::Paren(Box::new(ParenExpr {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 12),
                            ..BaseNode::default()
                        },
                        lparen: vec![],
                        expression: Expression::Binary(Box::new(BinaryExpr {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 11),
                                ..BaseNode::default()
                            },
                            operator: Operator::AdditionOperator,
                            left: Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 7),
                                    ..BaseNode::default()
                                },
                                value: 1
                            }),
                            right: Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 10, 1, 11),
                                    ..BaseNode::default()
                                },
                                value: 3
                            })
                        })),
                        rparen: vec![],
                    }))
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn logical_unary_operator_precedence() {
    let mut p = Parser::new(r#"not -1 == a"#);
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
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 12),
                    ..BaseNode::default()
                },
                expression: Expression::Unary(Box::new(UnaryExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 12),
                        ..BaseNode::default()
                    },
                    operator: Operator::NotOperator,
                    argument: Expression::Binary(Box::new(BinaryExpr {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 12),
                            ..BaseNode::default()
                        },
                        operator: Operator::EqualOperator,
                        left: Expression::Unary(Box::new(UnaryExpr {
                            base: BaseNode {
                                location: loc.get(1, 5, 1, 7),
                                ..BaseNode::default()
                            },
                            operator: Operator::SubtractionOperator,
                            argument: Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 7),
                                    ..BaseNode::default()
                                },
                                value: 1
                            })
                        })),
                        right: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 11, 1, 12),
                                ..BaseNode::default()
                            },
                            name: "a".to_string()
                        })
                    }))
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn all_operators_precedence() {
    let mut p = Parser::new(
        r#"a() == b.a + b.c * d < 100 and e != f[g] and h > i * j and
k / l < m + n - o or p() <= q() or r >= s and not t =~ /a/ and u !~ /a/"#,
    );
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 2, 72),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 2, 72),
                    ..BaseNode::default()
                },
                expression: Expression::Logical(Box::new(LogicalExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 2, 72),
                        ..BaseNode::default()
                    },
                    operator: LogicalOperator::OrOperator,
                    left: Expression::Logical(Box::new(LogicalExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 2, 32),
                            ..BaseNode::default()
                        },
                        operator: LogicalOperator::OrOperator,
                        left: Expression::Logical(Box::new(LogicalExpr {
                            base: BaseNode {
                                location: loc.get(1, 1, 2, 18),
                                ..BaseNode::default()
                            },
                            operator: LogicalOperator::AndOperator,
                            left: Expression::Logical(Box::new(LogicalExpr {
                                base: BaseNode {
                                    location: loc.get(1, 1, 1, 55),
                                    ..BaseNode::default()
                                },
                                operator: LogicalOperator::AndOperator,
                                left: Expression::Logical(Box::new(LogicalExpr {
                                    base: BaseNode {
                                        location: loc.get(1, 1, 1, 41),
                                        ..BaseNode::default()
                                    },
                                    operator: LogicalOperator::AndOperator,
                                    left: Expression::Binary(Box::new(BinaryExpr {
                                        base: BaseNode {
                                            location: loc.get(1, 1, 1, 27),
                                            ..BaseNode::default()
                                        },
                                        operator: Operator::LessThanOperator,
                                        left: Expression::Binary(Box::new(BinaryExpr {
                                            base: BaseNode {
                                                location: loc.get(1, 1, 1, 21),
                                                ..BaseNode::default()
                                            },
                                            operator: Operator::EqualOperator,
                                            left: Expression::Call(Box::new(CallExpr {
                                                base: BaseNode {
                                                    location: loc.get(1, 1, 1, 4),
                                                    ..BaseNode::default()
                                                },
                                                callee: Expression::Identifier(Identifier {
                                                    base: BaseNode {
                                                        location: loc.get(1, 1, 1, 2),
                                                        ..BaseNode::default()
                                                    },
                                                    name: "a".to_string()
                                                }),
                                                lparen: vec![],
                                                arguments: vec![],
                                                rparen: vec![],
                                            })),
                                            right: Expression::Binary(Box::new(BinaryExpr {
                                                base: BaseNode {
                                                    location: loc.get(1, 8, 1, 21),
                                                    ..BaseNode::default()
                                                },
                                                operator: Operator::AdditionOperator,
                                                left: Expression::Member(Box::new(MemberExpr {
                                                    base: BaseNode {
                                                        location: loc.get(1, 8, 1, 11),
                                                        ..BaseNode::default()
                                                    },
                                                    object: Expression::Identifier(Identifier {
                                                        base: BaseNode {
                                                            location: loc.get(1, 8, 1, 9),
                                                            ..BaseNode::default()
                                                        },
                                                        name: "b".to_string()
                                                    }),
                                                    lbrack: vec![],
                                                    property: PropertyKey::Identifier(Identifier {
                                                        base: BaseNode {
                                                            location: loc.get(1, 10, 1, 11),
                                                            ..BaseNode::default()
                                                        },
                                                        name: "a".to_string()
                                                    }),
                                                    rbrack: vec![],
                                                })),
                                                right: Expression::Binary(Box::new(BinaryExpr {
                                                    base: BaseNode {
                                                        location: loc.get(1, 14, 1, 21),
                                                        ..BaseNode::default()
                                                    },
                                                    operator: Operator::MultiplicationOperator,
                                                    left: Expression::Member(Box::new(
                                                        MemberExpr {
                                                            base: BaseNode {
                                                                location: loc.get(1, 14, 1, 17),
                                                                ..BaseNode::default()
                                                            },
                                                            object: Expression::Identifier(
                                                                Identifier {
                                                                    base: BaseNode {
                                                                        location: loc
                                                                            .get(1, 14, 1, 15),
                                                                        ..BaseNode::default()
                                                                    },
                                                                    name: "b".to_string()
                                                                }
                                                            ),
                                                            lbrack: vec![],
                                                            property: PropertyKey::Identifier(
                                                                Identifier {
                                                                    base: BaseNode {
                                                                        location: loc
                                                                            .get(1, 16, 1, 17),
                                                                        ..BaseNode::default()
                                                                    },
                                                                    name: "c".to_string()
                                                                }
                                                            ),
                                                            rbrack: vec![],
                                                        }
                                                    )),
                                                    right: Expression::Identifier(Identifier {
                                                        base: BaseNode {
                                                            location: loc.get(1, 20, 1, 21),
                                                            ..BaseNode::default()
                                                        },
                                                        name: "d".to_string()
                                                    })
                                                }))
                                            }))
                                        })),
                                        right: Expression::Integer(IntegerLit {
                                            base: BaseNode {
                                                location: loc.get(1, 24, 1, 27),
                                                ..BaseNode::default()
                                            },
                                            value: 100
                                        })
                                    })),
                                    right: Expression::Binary(Box::new(BinaryExpr {
                                        base: BaseNode {
                                            location: loc.get(1, 32, 1, 41),
                                            ..BaseNode::default()
                                        },
                                        operator: Operator::NotEqualOperator,
                                        left: Expression::Identifier(Identifier {
                                            base: BaseNode {
                                                location: loc.get(1, 32, 1, 33),
                                                ..BaseNode::default()
                                            },
                                            name: "e".to_string()
                                        }),
                                        right: Expression::Index(Box::new(IndexExpr {
                                            base: BaseNode {
                                                location: loc.get(1, 37, 1, 41),
                                                ..BaseNode::default()
                                            },
                                            array: Expression::Identifier(Identifier {
                                                base: BaseNode {
                                                    location: loc.get(1, 37, 1, 38),
                                                    ..BaseNode::default()
                                                },
                                                name: "f".to_string()
                                            }),
                                            lbrack: vec![],
                                            index: Expression::Identifier(Identifier {
                                                base: BaseNode {
                                                    location: loc.get(1, 39, 1, 40),
                                                    ..BaseNode::default()
                                                },
                                                name: "g".to_string()
                                            }),
                                            rbrack: vec![],
                                        }))
                                    }))
                                })),
                                right: Expression::Binary(Box::new(BinaryExpr {
                                    base: BaseNode {
                                        location: loc.get(1, 46, 1, 55),
                                        ..BaseNode::default()
                                    },
                                    operator: Operator::GreaterThanOperator,
                                    left: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(1, 46, 1, 47),
                                            ..BaseNode::default()
                                        },
                                        name: "h".to_string()
                                    }),
                                    right: Expression::Binary(Box::new(BinaryExpr {
                                        base: BaseNode {
                                            location: loc.get(1, 50, 1, 55),
                                            ..BaseNode::default()
                                        },
                                        operator: Operator::MultiplicationOperator,
                                        left: Expression::Identifier(Identifier {
                                            base: BaseNode {
                                                location: loc.get(1, 50, 1, 51),
                                                ..BaseNode::default()
                                            },
                                            name: "i".to_string()
                                        }),
                                        right: Expression::Identifier(Identifier {
                                            base: BaseNode {
                                                location: loc.get(1, 54, 1, 55),
                                                ..BaseNode::default()
                                            },
                                            name: "j".to_string()
                                        })
                                    }))
                                }))
                            })),
                            right: Expression::Binary(Box::new(BinaryExpr {
                                base: BaseNode {
                                    location: loc.get(2, 1, 2, 18),
                                    ..BaseNode::default()
                                },
                                operator: Operator::LessThanOperator,
                                left: Expression::Binary(Box::new(BinaryExpr {
                                    base: BaseNode {
                                        location: loc.get(2, 1, 2, 6),
                                        ..BaseNode::default()
                                    },
                                    operator: Operator::DivisionOperator,
                                    left: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(2, 1, 2, 2),
                                            ..BaseNode::default()
                                        },
                                        name: "k".to_string()
                                    }),
                                    right: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(2, 5, 2, 6),
                                            ..BaseNode::default()
                                        },
                                        name: "l".to_string()
                                    })
                                })),
                                right: Expression::Binary(Box::new(BinaryExpr {
                                    base: BaseNode {
                                        location: loc.get(2, 9, 2, 18),
                                        ..BaseNode::default()
                                    },
                                    operator: Operator::SubtractionOperator,
                                    left: Expression::Binary(Box::new(BinaryExpr {
                                        base: BaseNode {
                                            location: loc.get(2, 9, 2, 14),
                                            ..BaseNode::default()
                                        },
                                        operator: Operator::AdditionOperator,
                                        left: Expression::Identifier(Identifier {
                                            base: BaseNode {
                                                location: loc.get(2, 9, 2, 10),
                                                ..BaseNode::default()
                                            },
                                            name: "m".to_string()
                                        }),
                                        right: Expression::Identifier(Identifier {
                                            base: BaseNode {
                                                location: loc.get(2, 13, 2, 14),
                                                ..BaseNode::default()
                                            },
                                            name: "n".to_string()
                                        })
                                    })),
                                    right: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(2, 17, 2, 18),
                                            ..BaseNode::default()
                                        },
                                        name: "o".to_string()
                                    })
                                }))
                            }))
                        })),
                        right: Expression::Binary(Box::new(BinaryExpr {
                            base: BaseNode {
                                location: loc.get(2, 22, 2, 32),
                                ..BaseNode::default()
                            },
                            operator: Operator::LessThanEqualOperator,
                            left: Expression::Call(Box::new(CallExpr {
                                base: BaseNode {
                                    location: loc.get(2, 22, 2, 25),
                                    ..BaseNode::default()
                                },
                                callee: Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(2, 22, 2, 23),
                                        ..BaseNode::default()
                                    },
                                    name: "p".to_string()
                                }),
                                lparen: vec![],
                                arguments: vec![],
                                rparen: vec![],
                            })),
                            right: Expression::Call(Box::new(CallExpr {
                                base: BaseNode {
                                    location: loc.get(2, 29, 2, 32),
                                    ..BaseNode::default()
                                },
                                callee: Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(2, 29, 2, 30),
                                        ..BaseNode::default()
                                    },
                                    name: "q".to_string()
                                }),
                                lparen: vec![],
                                arguments: vec![],
                                rparen: vec![],
                            }))
                        }))
                    })),
                    right: Expression::Logical(Box::new(LogicalExpr {
                        base: BaseNode {
                            location: loc.get(2, 36, 2, 72),
                            ..BaseNode::default()
                        },
                        operator: LogicalOperator::AndOperator,
                        left: Expression::Logical(Box::new(LogicalExpr {
                            base: BaseNode {
                                location: loc.get(2, 36, 2, 59),
                                ..BaseNode::default()
                            },
                            operator: LogicalOperator::AndOperator,
                            left: Expression::Binary(Box::new(BinaryExpr {
                                base: BaseNode {
                                    location: loc.get(2, 36, 2, 42),
                                    ..BaseNode::default()
                                },
                                operator: Operator::GreaterThanEqualOperator,
                                left: Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(2, 36, 2, 37),
                                        ..BaseNode::default()
                                    },
                                    name: "r".to_string()
                                }),
                                right: Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(2, 41, 2, 42),
                                        ..BaseNode::default()
                                    },
                                    name: "s".to_string()
                                })
                            })),
                            right: Expression::Unary(Box::new(UnaryExpr {
                                base: BaseNode {
                                    location: loc.get(2, 47, 2, 59),
                                    ..BaseNode::default()
                                },
                                operator: Operator::NotOperator,
                                argument: Expression::Binary(Box::new(BinaryExpr {
                                    base: BaseNode {
                                        location: loc.get(2, 51, 2, 59),
                                        ..BaseNode::default()
                                    },
                                    operator: Operator::RegexpMatchOperator,
                                    left: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(2, 51, 2, 52),
                                            ..BaseNode::default()
                                        },
                                        name: "t".to_string()
                                    }),
                                    right: Expression::Regexp(RegexpLit {
                                        base: BaseNode {
                                            location: loc.get(2, 56, 2, 59),
                                            ..BaseNode::default()
                                        },
                                        value: "a".to_string()
                                    })
                                }))
                            }))
                        })),
                        right: Expression::Binary(Box::new(BinaryExpr {
                            base: BaseNode {
                                location: loc.get(2, 64, 2, 72),
                                ..BaseNode::default()
                            },
                            operator: Operator::NotRegexpMatchOperator,
                            left: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(2, 64, 2, 65),
                                    ..BaseNode::default()
                                },
                                name: "u".to_string()
                            }),
                            right: Expression::Regexp(RegexpLit {
                                base: BaseNode {
                                    location: loc.get(2, 69, 2, 72),
                                    ..BaseNode::default()
                                },
                                value: "a".to_string()
                            })
                        }))
                    }))
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn logical_operators_precedence_1() {
    let mut p = Parser::new(r#"not a or b"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
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
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 11),
                    ..BaseNode::default()
                },
                expression: Expression::Logical(Box::new(LogicalExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 11),
                        ..BaseNode::default()
                    },
                    operator: LogicalOperator::OrOperator,
                    left: Expression::Unary(Box::new(UnaryExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 6),
                            ..BaseNode::default()
                        },
                        operator: Operator::NotOperator,
                        argument: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 5, 1, 6),
                                ..BaseNode::default()
                            },
                            name: "a".to_string()
                        })
                    })),
                    right: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 10, 1, 11),
                            ..BaseNode::default()
                        },
                        name: "b".to_string()
                    })
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn logical_operators_precedence_2() {
    let mut p = Parser::new(r#"a or not b"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
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
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 11),
                    ..BaseNode::default()
                },
                expression: Expression::Logical(Box::new(LogicalExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 11),
                        ..BaseNode::default()
                    },
                    operator: LogicalOperator::OrOperator,
                    left: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            ..BaseNode::default()
                        },
                        name: "a".to_string()
                    }),
                    right: Expression::Unary(Box::new(UnaryExpr {
                        base: BaseNode {
                            location: loc.get(1, 6, 1, 11),
                            ..BaseNode::default()
                        },
                        operator: Operator::NotOperator,
                        argument: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 10, 1, 11),
                                ..BaseNode::default()
                            },
                            name: "b".to_string()
                        })
                    }))
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn logical_operators_precedence_3() {
    let mut p = Parser::new(r#"not a and b"#);
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
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 12),
                    ..BaseNode::default()
                },
                expression: Expression::Logical(Box::new(LogicalExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 12),
                        ..BaseNode::default()
                    },
                    operator: LogicalOperator::AndOperator,
                    left: Expression::Unary(Box::new(UnaryExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 6),
                            ..BaseNode::default()
                        },
                        operator: Operator::NotOperator,
                        argument: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 5, 1, 6),
                                ..BaseNode::default()
                            },
                            name: "a".to_string()
                        })
                    })),
                    right: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 11, 1, 12),
                            ..BaseNode::default()
                        },
                        name: "b".to_string()
                    })
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn logical_operators_precedence_4() {
    let mut p = Parser::new(r#"a and not b"#);
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
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 12),
                    ..BaseNode::default()
                },
                expression: Expression::Logical(Box::new(LogicalExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 12),
                        ..BaseNode::default()
                    },
                    operator: LogicalOperator::AndOperator,
                    left: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            ..BaseNode::default()
                        },
                        name: "a".to_string()
                    }),
                    right: Expression::Unary(Box::new(UnaryExpr {
                        base: BaseNode {
                            location: loc.get(1, 7, 1, 12),
                            ..BaseNode::default()
                        },
                        operator: Operator::NotOperator,
                        argument: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 11, 1, 12),
                                ..BaseNode::default()
                            },
                            name: "b".to_string()
                        })
                    }))
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn logical_operators_precedence_5() {
    let mut p = Parser::new(r#"a and b or c"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
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
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 13),
                    ..BaseNode::default()
                },
                expression: Expression::Logical(Box::new(LogicalExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 13),
                        ..BaseNode::default()
                    },
                    operator: LogicalOperator::OrOperator,
                    left: Expression::Logical(Box::new(LogicalExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 8),
                            ..BaseNode::default()
                        },
                        operator: LogicalOperator::AndOperator,
                        left: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 2),
                                ..BaseNode::default()
                            },
                            name: "a".to_string()
                        }),
                        right: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 7, 1, 8),
                                ..BaseNode::default()
                            },
                            name: "b".to_string()
                        })
                    })),
                    right: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 12, 1, 13),
                            ..BaseNode::default()
                        },
                        name: "c".to_string()
                    })
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn logical_operators_precedence_6() {
    let mut p = Parser::new(r#"a or b and c"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
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
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 13),
                    ..BaseNode::default()
                },
                expression: Expression::Logical(Box::new(LogicalExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 13),
                        ..BaseNode::default()
                    },
                    operator: LogicalOperator::OrOperator,
                    left: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            ..BaseNode::default()
                        },
                        name: "a".to_string()
                    }),
                    right: Expression::Logical(Box::new(LogicalExpr {
                        base: BaseNode {
                            location: loc.get(1, 6, 1, 13),
                            ..BaseNode::default()
                        },
                        operator: LogicalOperator::AndOperator,
                        left: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 7),
                                ..BaseNode::default()
                            },
                            name: "b".to_string()
                        }),
                        right: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 12, 1, 13),
                                ..BaseNode::default()
                            },
                            name: "c".to_string()
                        })
                    }))
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn logical_operators_precedence_7() {
    let mut p = Parser::new(r#"not (a or b)"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
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
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 13),
                    ..BaseNode::default()
                },
                expression: Expression::Unary(Box::new(UnaryExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 13),
                        ..BaseNode::default()
                    },
                    operator: Operator::NotOperator,
                    argument: Expression::Paren(Box::new(ParenExpr {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 13),
                            ..BaseNode::default()
                        },
                        lparen: vec![],
                        expression: Expression::Logical(Box::new(LogicalExpr {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 12),
                                ..BaseNode::default()
                            },
                            operator: LogicalOperator::OrOperator,
                            left: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 7),
                                    ..BaseNode::default()
                                },
                                name: "a".to_string()
                            }),
                            right: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 11, 1, 12),
                                    ..BaseNode::default()
                                },
                                name: "b".to_string()
                            })
                        })),
                        rparen: vec![],
                    }))
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn logical_operators_precedence_8() {
    let mut p = Parser::new(r#"not (a and b)"#);
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
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 14),
                    ..BaseNode::default()
                },
                expression: Expression::Unary(Box::new(UnaryExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 14),
                        ..BaseNode::default()
                    },
                    operator: Operator::NotOperator,
                    argument: Expression::Paren(Box::new(ParenExpr {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 14),
                            ..BaseNode::default()
                        },
                        lparen: vec![],
                        expression: Expression::Logical(Box::new(LogicalExpr {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 13),
                                ..BaseNode::default()
                            },
                            operator: LogicalOperator::AndOperator,
                            left: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 7),
                                    ..BaseNode::default()
                                },
                                name: "a".to_string()
                            }),
                            right: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 12, 1, 13),
                                    ..BaseNode::default()
                                },
                                name: "b".to_string()
                            })
                        })),
                        rparen: vec![],
                    }))
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn logical_operators_precedence_9() {
    let mut p = Parser::new(r#"(a or b) and c"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 15),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 15),
                    ..BaseNode::default()
                },
                expression: Expression::Logical(Box::new(LogicalExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 15),
                        ..BaseNode::default()
                    },
                    operator: LogicalOperator::AndOperator,
                    left: Expression::Paren(Box::new(ParenExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 9),
                            ..BaseNode::default()
                        },
                        lparen: vec![],
                        expression: Expression::Logical(Box::new(LogicalExpr {
                            base: BaseNode {
                                location: loc.get(1, 2, 1, 8),
                                ..BaseNode::default()
                            },
                            operator: LogicalOperator::OrOperator,
                            left: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 2, 1, 3),
                                    ..BaseNode::default()
                                },
                                name: "a".to_string()
                            }),
                            right: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 7, 1, 8),
                                    ..BaseNode::default()
                                },
                                name: "b".to_string()
                            })
                        })),
                        rparen: vec![],
                    })),
                    right: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 14, 1, 15),
                            ..BaseNode::default()
                        },
                        name: "c".to_string()
                    })
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn logical_operators_precedence_10() {
    let mut p = Parser::new(r#"a and (b or c)"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 15),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 15),
                    ..BaseNode::default()
                },
                expression: Expression::Logical(Box::new(LogicalExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 15),
                        ..BaseNode::default()
                    },
                    operator: LogicalOperator::AndOperator,
                    left: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            ..BaseNode::default()
                        },
                        name: "a".to_string()
                    }),
                    right: Expression::Paren(Box::new(ParenExpr {
                        base: BaseNode {
                            location: loc.get(1, 7, 1, 15),
                            ..BaseNode::default()
                        },
                        lparen: vec![],
                        expression: Expression::Logical(Box::new(LogicalExpr {
                            base: BaseNode {
                                location: loc.get(1, 8, 1, 14),
                                ..BaseNode::default()
                            },
                            operator: LogicalOperator::OrOperator,
                            left: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 8, 1, 9),
                                    ..BaseNode::default()
                                },
                                name: "b".to_string()
                            }),
                            right: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 13, 1, 14),
                                    ..BaseNode::default()
                                },
                                name: "c".to_string()
                            })
                        })),
                        rparen: vec![],
                    }))
                }))
            }))],
            eof: vec![],
        },
    )
}

// The following test case demonstrates confusing behavior:
// The `(` at 2:1 begins a call, but a user might
// reasonably expect it to start a new statement.
#[test]
fn two_logical_operations_with_parens() {
    let mut p = Parser::new(
        r#"not (a and b)
(a or b) and c"#,
    );
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 2, 15),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 2, 15),
                    ..BaseNode::default()
                },
                expression: Expression::Logical(Box::new(LogicalExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 2, 15),
                        ..BaseNode::default()
                    },
                    operator: LogicalOperator::AndOperator,
                    left: Expression::Unary(Box::new(UnaryExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 2, 9),
                            ..BaseNode::default()
                        },
                        operator: Operator::NotOperator,
                        argument: Expression::Call(Box::new(CallExpr {
                            base: BaseNode {
                                location: loc.get(1, 5, 2, 9),
                                errors: vec!["expected comma in property list, got OR".to_string()],
                                ..BaseNode::default()
                            },
                            callee: Expression::Paren(Box::new(ParenExpr {
                                base: BaseNode {
                                    location: loc.get(1, 5, 1, 14),
                                    ..BaseNode::default()
                                },
                                lparen: vec![],
                                expression: Expression::Logical(Box::new(LogicalExpr {
                                    base: BaseNode {
                                        location: loc.get(1, 6, 1, 13),
                                        ..BaseNode::default()
                                    },
                                    operator: LogicalOperator::AndOperator,
                                    left: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(1, 6, 1, 7),
                                            ..BaseNode::default()
                                        },
                                        name: "a".to_string()
                                    }),
                                    right: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(1, 12, 1, 13),
                                            ..BaseNode::default()
                                        },
                                        name: "b".to_string()
                                    })
                                })),
                                rparen: vec![],
                            })),
                            lparen: vec![],
                            arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                base: BaseNode {
                                    location: loc.get(2, 2, 2, 8),
                                    ..BaseNode::default()
                                },
                                lbrace: vec![],
                                with: None,
                                properties: vec![
                                    Property {
                                        base: BaseNode {
                                            location: loc.get(2, 2, 2, 3),
                                            ..BaseNode::default()
                                        },
                                        key: PropertyKey::Identifier(Identifier {
                                            base: BaseNode {
                                                location: loc.get(2, 2, 2, 3),
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
                                            location: loc.get(2, 4, 2, 8),
                                            errors: vec![
                                                "unexpected token for property key: OR (or)"
                                                    .to_string()
                                            ],
                                            ..BaseNode::default()
                                        },
                                        key: PropertyKey::StringLit(StringLit {
                                            base: BaseNode {
                                                location: loc.get(2, 4, 2, 4),
                                                ..BaseNode::default()
                                            },
                                            value: "<invalid>".to_string()
                                        }),
                                        separator: vec![],
                                        value: None,
                                        comma: vec![],
                                    }
                                ],
                                rbrace: vec![],
                            }))],
                            rparen: vec![],
                        }))
                    })),
                    right: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(2, 14, 2, 15),
                            ..BaseNode::default()
                        },
                        name: "c".to_string()
                    })
                }))
            }))],
            eof: vec![],
        },
    )
}
