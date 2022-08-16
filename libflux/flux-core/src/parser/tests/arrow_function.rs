use pretty_assertions::assert_eq;

use super::*;
use crate::ast::tests::Locator;

#[test]
fn arrow_function_called() {
    let mut p = Parser::new(
        r#"plusOne = (r) => r + 1
   plusOne(r:5)"#,
    );
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 2, 16),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 23),
                        ..BaseNode::default()
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 8),
                            ..BaseNode::default()
                        },
                        name: "plusOne".to_string()
                    },
                    init: Expression::Function(Box::new(FunctionExpr {
                        base: BaseNode {
                            location: loc.get(1, 11, 1, 23),
                            ..BaseNode::default()
                        },
                        lparen: vec![],
                        params: vec![Property {
                            base: BaseNode {
                                location: loc.get(1, 12, 1, 13),
                                ..BaseNode::default()
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 12, 1, 13),
                                    ..BaseNode::default()
                                },
                                name: "r".to_string()
                            }),
                            separator: vec![],
                            value: None,
                            comma: vec![],
                        }],
                        rparen: vec![],
                        arrow: vec![],
                        body: FunctionBody::Expr(Expression::Binary(Box::new(BinaryExpr {
                            base: BaseNode {
                                location: loc.get(1, 18, 1, 23),
                                ..BaseNode::default()
                            },
                            operator: Operator::AdditionOperator,
                            left: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 18, 1, 19),
                                    ..BaseNode::default()
                                },
                                name: "r".to_string()
                            }),
                            right: Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 22, 1, 23),
                                    ..BaseNode::default()
                                },
                                value: 1
                            })
                        }))),
                    }))
                })),
                Statement::Expr(Box::new(ExprStmt {
                    base: BaseNode {
                        location: loc.get(2, 4, 2, 16),
                        ..BaseNode::default()
                    },
                    expression: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(2, 4, 2, 16),
                            ..BaseNode::default()
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(2, 4, 2, 11),
                                ..BaseNode::default()
                            },
                            name: "plusOne".to_string()
                        }),
                        lparen: vec![],
                        arguments: vec![Expression::Object(Box::new(ObjectExpr {
                            base: BaseNode {
                                location: loc.get(2, 12, 2, 15),
                                ..BaseNode::default()
                            },
                            lbrace: vec![],
                            with: None,
                            properties: vec![Property {
                                base: BaseNode {
                                    location: loc.get(2, 12, 2, 15),
                                    ..BaseNode::default()
                                },
                                key: PropertyKey::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(2, 12, 2, 13),
                                        ..BaseNode::default()
                                    },
                                    name: "r".to_string()
                                }),
                                separator: vec![],
                                value: Some(Expression::Integer(IntegerLit {
                                    base: BaseNode {
                                        location: loc.get(2, 14, 2, 15),
                                        ..BaseNode::default()
                                    },
                                    value: 5
                                })),
                                comma: vec![],
                            }],
                            rbrace: vec![],
                        }))],
                        rparen: vec![],
                    }))
                }))
            ],
            eof: vec![],
        },
    )
}

#[test]
fn arrow_function_return_map() {
    let mut p = Parser::new(r#"toMap = (r) =>({r:r})"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 22),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 22),
                    ..BaseNode::default()
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 6),
                        ..BaseNode::default()
                    },
                    name: "toMap".to_string()
                },
                init: Expression::Function(Box::new(FunctionExpr {
                    base: BaseNode {
                        location: loc.get(1, 9, 1, 22),
                        ..BaseNode::default()
                    },
                    lparen: vec![],
                    params: vec![Property {
                        base: BaseNode {
                            location: loc.get(1, 10, 1, 11),
                            ..BaseNode::default()
                        },
                        key: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 10, 1, 11),
                                ..BaseNode::default()
                            },
                            name: "r".to_string()
                        }),
                        separator: vec![],
                        value: None,
                        comma: vec![],
                    }],
                    rparen: vec![],
                    arrow: vec![],
                    body: FunctionBody::Expr(Expression::Paren(Box::new(ParenExpr {
                        base: BaseNode {
                            location: loc.get(1, 15, 1, 22),
                            ..BaseNode::default()
                        },
                        lparen: vec![],
                        expression: Expression::Object(Box::new(ObjectExpr {
                            base: BaseNode {
                                location: loc.get(1, 16, 1, 21),
                                ..BaseNode::default()
                            },
                            lbrace: vec![],
                            with: None,
                            properties: vec![Property {
                                base: BaseNode {
                                    location: loc.get(1, 17, 1, 20),
                                    ..BaseNode::default()
                                },
                                key: PropertyKey::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 17, 1, 18),
                                        ..BaseNode::default()
                                    },
                                    name: "r".to_string()
                                }),
                                separator: vec![],
                                value: Some(Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 19, 1, 20),
                                        ..BaseNode::default()
                                    },
                                    name: "r".to_string()
                                })),
                                comma: vec![],
                            }],
                            rbrace: vec![],
                        })),
                        rparen: vec![],
                    }))),
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn arrow_function() {
    let mut p = Parser::new(r#"(x,y) => x == y"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 16),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 16),
                    ..BaseNode::default()
                },
                expression: Expression::Function(Box::new(FunctionExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 16),
                        ..BaseNode::default()
                    },
                    lparen: vec![],
                    params: vec![
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 2, 1, 3),
                                ..BaseNode::default()
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 2, 1, 3),
                                    ..BaseNode::default()
                                },
                                name: "x".to_string()
                            }),
                            separator: vec![],
                            value: None,
                            comma: vec![],
                        },
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 4, 1, 5),
                                ..BaseNode::default()
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 4, 1, 5),
                                    ..BaseNode::default()
                                },
                                name: "y".to_string()
                            }),
                            separator: vec![],
                            value: None,
                            comma: vec![],
                        }
                    ],
                    rparen: vec![],
                    arrow: vec![],
                    body: FunctionBody::Expr(Expression::Binary(Box::new(BinaryExpr {
                        base: BaseNode {
                            location: loc.get(1, 10, 1, 16),
                            ..BaseNode::default()
                        },
                        operator: Operator::EqualOperator,
                        left: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 10, 1, 11),
                                ..BaseNode::default()
                            },
                            name: "x".to_string(),
                        }),
                        right: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 15, 1, 16),
                                ..BaseNode::default()
                            },
                            name: "y".to_string(),
                        }),
                    }))),
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn arrow_function_with_default_arg() {
    let mut p = Parser::new(r#"addN = (r, n=5) => r + n"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 25),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 25),
                    ..BaseNode::default()
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 5),
                        ..BaseNode::default()
                    },
                    name: "addN".to_string()
                },
                init: Expression::Function(Box::new(FunctionExpr {
                    base: BaseNode {
                        location: loc.get(1, 8, 1, 25),
                        ..BaseNode::default()
                    },
                    lparen: vec![],
                    params: vec![
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
                                name: "r".to_string()
                            }),
                            separator: vec![],
                            value: None,
                            comma: vec![],
                        },
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 12, 1, 15),
                                ..BaseNode::default()
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 12, 1, 13),
                                    ..BaseNode::default()
                                },
                                name: "n".to_string()
                            }),
                            separator: vec![],
                            value: Some(Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 14, 1, 15),
                                    ..BaseNode::default()
                                },
                                value: 5
                            })),
                            comma: vec![],
                        }
                    ],
                    rparen: vec![],
                    arrow: vec![],
                    body: FunctionBody::Expr(Expression::Binary(Box::new(BinaryExpr {
                        base: BaseNode {
                            location: loc.get(1, 20, 1, 25),
                            ..BaseNode::default()
                        },
                        operator: Operator::AdditionOperator,
                        left: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 20, 1, 21),
                                ..BaseNode::default()
                            },
                            name: "r".to_string()
                        }),
                        right: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 24, 1, 25),
                                ..BaseNode::default()
                            },
                            name: "n".to_string()
                        })
                    }))),
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn arrow_function_called_in_binary_expression() {
    let mut p = Parser::new(
        r#"
            plusOne = (r) => r + 1
            plusOne(r:5) == 6 or die()"#,
    );
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(2, 13, 3, 39),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(2, 13, 2, 35),
                        ..BaseNode::default()
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(2, 13, 2, 20),
                            ..BaseNode::default()
                        },
                        name: "plusOne".to_string()
                    },
                    init: Expression::Function(Box::new(FunctionExpr {
                        base: BaseNode {
                            location: loc.get(2, 23, 2, 35),
                            ..BaseNode::default()
                        },
                        lparen: vec![],
                        params: vec![Property {
                            base: BaseNode {
                                location: loc.get(2, 24, 2, 25),
                                ..BaseNode::default()
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(2, 24, 2, 25),
                                    ..BaseNode::default()
                                },
                                name: "r".to_string()
                            }),
                            separator: vec![],
                            value: None,
                            comma: vec![],
                        }],
                        rparen: vec![],
                        arrow: vec![],
                        body: FunctionBody::Expr(Expression::Binary(Box::new(BinaryExpr {
                            base: BaseNode {
                                location: loc.get(2, 30, 2, 35),
                                ..BaseNode::default()
                            },
                            operator: Operator::AdditionOperator,
                            left: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(2, 30, 2, 31),
                                    ..BaseNode::default()
                                },
                                name: "r".to_string()
                            }),
                            right: Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(2, 34, 2, 35),
                                    ..BaseNode::default()
                                },
                                value: 1
                            })
                        }))),
                    }))
                })),
                Statement::Expr(Box::new(ExprStmt {
                    base: BaseNode {
                        location: loc.get(3, 13, 3, 39),
                        ..BaseNode::default()
                    },
                    expression: Expression::Logical(Box::new(LogicalExpr {
                        base: BaseNode {
                            location: loc.get(3, 13, 3, 39),
                            ..BaseNode::default()
                        },
                        operator: LogicalOperator::OrOperator,
                        left: Expression::Binary(Box::new(BinaryExpr {
                            base: BaseNode {
                                location: loc.get(3, 13, 3, 30),
                                ..BaseNode::default()
                            },
                            operator: Operator::EqualOperator,
                            left: Expression::Call(Box::new(CallExpr {
                                base: BaseNode {
                                    location: loc.get(3, 13, 3, 25),
                                    ..BaseNode::default()
                                },
                                callee: Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(3, 13, 3, 20),
                                        ..BaseNode::default()
                                    },
                                    name: "plusOne".to_string()
                                }),
                                lparen: vec![],
                                arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                    base: BaseNode {
                                        location: loc.get(3, 21, 3, 24),
                                        ..BaseNode::default()
                                    },
                                    lbrace: vec![],
                                    with: None,
                                    properties: vec![Property {
                                        base: BaseNode {
                                            location: loc.get(3, 21, 3, 24),
                                            ..BaseNode::default()
                                        },
                                        key: PropertyKey::Identifier(Identifier {
                                            base: BaseNode {
                                                location: loc.get(3, 21, 3, 22),
                                                ..BaseNode::default()
                                            },
                                            name: "r".to_string()
                                        }),
                                        separator: vec![],
                                        value: Some(Expression::Integer(IntegerLit {
                                            base: BaseNode {
                                                location: loc.get(3, 23, 3, 24),
                                                ..BaseNode::default()
                                            },
                                            value: 5
                                        })),
                                        comma: vec![],
                                    }],
                                    rbrace: vec![],
                                }))],
                                rparen: vec![],
                            })),
                            right: Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(3, 29, 3, 30),
                                    ..BaseNode::default()
                                },
                                value: 6
                            })
                        })),
                        right: Expression::Call(Box::new(CallExpr {
                            base: BaseNode {
                                location: loc.get(3, 34, 3, 39),
                                ..BaseNode::default()
                            },
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(3, 34, 3, 37),
                                    ..BaseNode::default()
                                },
                                name: "die".to_string()
                            }),
                            lparen: vec![],
                            arguments: vec![],
                            rparen: vec![],
                        }))
                    }))
                }))
            ],
            eof: vec![],
        },
    )
}

#[test]
fn arrow_function_as_single_expression() {
    let mut p = Parser::new(r#"f = (r) => r["_measurement"] == "cpu""#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 38),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 38),
                    ..BaseNode::default()
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 2),
                        ..BaseNode::default()
                    },
                    name: "f".to_string()
                },
                init: Expression::Function(Box::new(FunctionExpr {
                    base: BaseNode {
                        location: loc.get(1, 5, 1, 38),
                        ..BaseNode::default()
                    },
                    lparen: vec![],
                    params: vec![Property {
                        base: BaseNode {
                            location: loc.get(1, 6, 1, 7),
                            ..BaseNode::default()
                        },
                        key: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 7),
                                ..BaseNode::default()
                            },
                            name: "r".to_string()
                        }),
                        separator: vec![],
                        value: None,
                        comma: vec![],
                    }],
                    rparen: vec![],
                    arrow: vec![],
                    body: FunctionBody::Expr(Expression::Binary(Box::new(BinaryExpr {
                        base: BaseNode {
                            location: loc.get(1, 12, 1, 38),
                            ..BaseNode::default()
                        },
                        operator: Operator::EqualOperator,
                        left: Expression::Member(Box::new(MemberExpr {
                            base: BaseNode {
                                location: loc.get(1, 12, 1, 29),
                                ..BaseNode::default()
                            },
                            object: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 12, 1, 13),
                                    ..BaseNode::default()
                                },
                                name: "r".to_string()
                            }),
                            lbrack: vec![],
                            property: PropertyKey::StringLit(StringLit {
                                base: BaseNode {
                                    location: loc.get(1, 14, 1, 28),
                                    ..BaseNode::default()
                                },
                                value: "_measurement".to_string()
                            }),
                            rbrack: vec![],
                        })),
                        right: Expression::StringLit(StringLit {
                            base: BaseNode {
                                location: loc.get(1, 33, 1, 38),
                                ..BaseNode::default()
                            },
                            value: "cpu".to_string()
                        })
                    }))),
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn arrow_function_as_block() {
    let mut p = Parser::new(
        r#"f = (r) => {
                m = r["_measurement"]
                return m == "cpu"
            }"#,
    );
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 4, 14),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 4, 14),
                    ..BaseNode::default()
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 2),
                        ..BaseNode::default()
                    },
                    name: "f".to_string()
                },
                init: Expression::Function(Box::new(FunctionExpr {
                    base: BaseNode {
                        location: loc.get(1, 5, 4, 14),
                        ..BaseNode::default()
                    },
                    lparen: vec![],
                    params: vec![Property {
                        base: BaseNode {
                            location: loc.get(1, 6, 1, 7),
                            ..BaseNode::default()
                        },
                        key: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 7),
                                ..BaseNode::default()
                            },
                            name: "r".to_string()
                        }),
                        separator: vec![],
                        value: None,
                        comma: vec![],
                    }],
                    rparen: vec![],
                    arrow: vec![],
                    body: FunctionBody::Block(Block {
                        base: BaseNode {
                            location: loc.get(1, 12, 4, 14),
                            ..BaseNode::default()
                        },
                        lbrace: vec![],
                        body: vec![
                            Statement::Variable(Box::new(VariableAssgn {
                                base: BaseNode {
                                    location: loc.get(2, 17, 2, 38),
                                    ..BaseNode::default()
                                },
                                id: Identifier {
                                    base: BaseNode {
                                        location: loc.get(2, 17, 2, 18),
                                        ..BaseNode::default()
                                    },
                                    name: "m".to_string()
                                },
                                init: Expression::Member(Box::new(MemberExpr {
                                    base: BaseNode {
                                        location: loc.get(2, 21, 2, 38),
                                        ..BaseNode::default()
                                    },
                                    object: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(2, 21, 2, 22),
                                            ..BaseNode::default()
                                        },
                                        name: "r".to_string()
                                    }),
                                    lbrack: vec![],
                                    property: PropertyKey::StringLit(StringLit {
                                        base: BaseNode {
                                            location: loc.get(2, 23, 2, 37),
                                            ..BaseNode::default()
                                        },
                                        value: "_measurement".to_string()
                                    }),
                                    rbrack: vec![],
                                }))
                            })),
                            Statement::Return(Box::new(ReturnStmt {
                                base: BaseNode {
                                    location: loc.get(3, 17, 3, 34),
                                    ..BaseNode::default()
                                },
                                argument: Expression::Binary(Box::new(BinaryExpr {
                                    base: BaseNode {
                                        location: loc.get(3, 24, 3, 34),
                                        ..BaseNode::default()
                                    },
                                    operator: Operator::EqualOperator,
                                    left: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(3, 24, 3, 25),
                                            ..BaseNode::default()
                                        },
                                        name: "m".to_string()
                                    }),
                                    right: Expression::StringLit(StringLit {
                                        base: BaseNode {
                                            location: loc.get(3, 29, 3, 34),
                                            ..BaseNode::default()
                                        },
                                        value: "cpu".to_string()
                                    })
                                }))
                            }))
                        ],
                        rbrace: vec![],
                    }),
                }))
            }))],
            eof: vec![],
        },
    )
}
