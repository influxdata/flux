use pretty_assertions::assert_eq;

use super::*;
use crate::locator::Locator;

#[test]
fn from() {
    let mut p = Parser::new(r#"from()"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
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
                expression: Expression::Call(Box::new(CallExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 7),
                        ..BaseNode::default()
                    },
                    callee: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 5),
                            ..BaseNode::default()
                        },
                        name: "from".to_string()
                    }),
                    lparen: vec![],
                    arguments: vec![],
                    rparen: vec![],
                })),
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn from_with_database() {
    let mut p = Parser::new(r#"from(bucket:"telegraf/autogen")"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 32),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 32),
                    ..BaseNode::default()
                },
                expression: Expression::Call(Box::new(CallExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 32),
                        ..BaseNode::default()
                    },
                    callee: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 5),
                            ..BaseNode::default()
                        },
                        name: "from".to_string()
                    }),
                    lparen: vec![],
                    arguments: vec![Expression::Object(Box::new(ObjectExpr {
                        base: BaseNode {
                            location: loc.get(1, 6, 1, 31),
                            ..BaseNode::default()
                        },
                        lbrace: vec![],
                        with: None,
                        properties: vec![Property {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 31),
                                ..BaseNode::default()
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 12),
                                    ..BaseNode::default()
                                },
                                name: "bucket".to_string()
                            }),
                            separator: vec![],
                            value: Some(Expression::StringLit(StringLit {
                                base: BaseNode {
                                    location: loc.get(1, 13, 1, 31),
                                    ..BaseNode::default()
                                },
                                value: "telegraf/autogen".to_string()
                            })),
                            comma: vec![],
                        }],
                        rbrace: vec![],
                    }))],
                    rparen: vec![],
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn from_with_filter_with_no_parens() {
    let mut p = Parser::new(
        r#"from(bucket:"telegraf/autogen").filter(fn: (r) => r["other"]=="mem" and r["this"]=="that" or r["these"]!="those")"#,
    );
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 114),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 114),
                    ..BaseNode::default()
                },
                expression: Expression::Call(Box::new(CallExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 114),
                        ..BaseNode::default()
                    },
                    callee: Expression::Member(Box::new(MemberExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 39),
                            ..BaseNode::default()
                        },
                        property: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 33, 1, 39),
                                ..BaseNode::default()
                            },
                            name: "filter".to_string()
                        }),
                        lbrack: vec![],
                        object: Expression::Call(Box::new(CallExpr {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 32),
                                ..BaseNode::default()
                            },
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 1, 1, 5),
                                    ..BaseNode::default()
                                },
                                name: "from".to_string()
                            }),
                            lparen: vec![],
                            arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 31),
                                    ..BaseNode::default()
                                },
                                lbrace: vec![],
                                with: None,
                                properties: vec![Property {
                                    base: BaseNode {
                                        location: loc.get(1, 6, 1, 31),
                                        ..BaseNode::default()
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(1, 6, 1, 12),
                                            ..BaseNode::default()
                                        },
                                        name: "bucket".to_string()
                                    }),
                                    separator: vec![],
                                    value: Some(Expression::StringLit(StringLit {
                                        base: BaseNode {
                                            location: loc.get(1, 13, 1, 31),
                                            ..BaseNode::default()
                                        },
                                        value: "telegraf/autogen".to_string()
                                    })),
                                    comma: vec![],
                                }],
                                rbrace: vec![],
                            }))],
                            rparen: vec![],
                        })),
                        rbrack: vec![],
                    })),
                    lparen: vec![],
                    arguments: vec![Expression::Object(Box::new(ObjectExpr {
                        base: BaseNode {
                            location: loc.get(1, 40, 1, 113),
                            ..BaseNode::default()
                        },
                        lbrace: vec![],
                        with: None,
                        properties: vec![Property {
                            base: BaseNode {
                                location: loc.get(1, 40, 1, 113),
                                ..BaseNode::default()
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 40, 1, 42),
                                    ..BaseNode::default()
                                },
                                name: "fn".to_string()
                            }),
                            separator: vec![],
                            value: Some(Expression::Function(Box::new(FunctionExpr {
                                base: BaseNode {
                                    location: loc.get(1, 44, 1, 113),
                                    ..BaseNode::default()
                                },
                                lparen: vec![],
                                params: vec![Property {
                                    base: BaseNode {
                                        location: loc.get(1, 45, 1, 46),
                                        ..BaseNode::default()
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(1, 45, 1, 46),
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
                                body: FunctionBody::Expr(Expression::Logical(Box::new(
                                    LogicalExpr {
                                        base: BaseNode {
                                            location: loc.get(1, 51, 1, 113),
                                            ..BaseNode::default()
                                        },
                                        operator: LogicalOperator::OrOperator,
                                        left: Expression::Logical(Box::new(LogicalExpr {
                                            base: BaseNode {
                                                location: loc.get(1, 51, 1, 90),
                                                ..BaseNode::default()
                                            },
                                            operator: LogicalOperator::AndOperator,
                                            left: Expression::Binary(Box::new(BinaryExpr {
                                                base: BaseNode {
                                                    location: loc.get(1, 51, 1, 68),
                                                    ..BaseNode::default()
                                                },
                                                operator: Operator::EqualOperator,
                                                left: Expression::Member(Box::new(MemberExpr {
                                                    base: BaseNode {
                                                        location: loc.get(1, 51, 1, 61),
                                                        ..BaseNode::default()
                                                    },
                                                    object: Expression::Identifier(Identifier {
                                                        base: BaseNode {
                                                            location: loc.get(1, 51, 1, 52),
                                                            ..BaseNode::default()
                                                        },
                                                        name: "r".to_string()
                                                    }),
                                                    lbrack: vec![],
                                                    property: PropertyKey::StringLit(StringLit {
                                                        base: BaseNode {
                                                            location: loc.get(1, 53, 1, 60),
                                                            ..BaseNode::default()
                                                        },
                                                        value: "other".to_string()
                                                    }),
                                                    rbrack: vec![],
                                                })),
                                                right: Expression::StringLit(StringLit {
                                                    base: BaseNode {
                                                        location: loc.get(1, 63, 1, 68),
                                                        ..BaseNode::default()
                                                    },
                                                    value: "mem".to_string()
                                                })
                                            })),
                                            right: Expression::Binary(Box::new(BinaryExpr {
                                                base: BaseNode {
                                                    location: loc.get(1, 73, 1, 90),
                                                    ..BaseNode::default()
                                                },
                                                operator: Operator::EqualOperator,
                                                left: Expression::Member(Box::new(MemberExpr {
                                                    base: BaseNode {
                                                        location: loc.get(1, 73, 1, 82),
                                                        ..BaseNode::default()
                                                    },
                                                    object: Expression::Identifier(Identifier {
                                                        base: BaseNode {
                                                            location: loc.get(1, 73, 1, 74),
                                                            ..BaseNode::default()
                                                        },
                                                        name: "r".to_string()
                                                    }),
                                                    lbrack: vec![],
                                                    property: PropertyKey::StringLit(StringLit {
                                                        base: BaseNode {
                                                            location: loc.get(1, 75, 1, 81),
                                                            ..BaseNode::default()
                                                        },
                                                        value: "this".to_string()
                                                    }),
                                                    rbrack: vec![],
                                                })),
                                                right: Expression::StringLit(StringLit {
                                                    base: BaseNode {
                                                        location: loc.get(1, 84, 1, 90),
                                                        ..BaseNode::default()
                                                    },
                                                    value: "that".to_string()
                                                })
                                            }))
                                        })),
                                        right: Expression::Binary(Box::new(BinaryExpr {
                                            base: BaseNode {
                                                location: loc.get(1, 94, 1, 113),
                                                ..BaseNode::default()
                                            },
                                            operator: Operator::NotEqualOperator,
                                            left: Expression::Member(Box::new(MemberExpr {
                                                base: BaseNode {
                                                    location: loc.get(1, 94, 1, 104),
                                                    ..BaseNode::default()
                                                },
                                                object: Expression::Identifier(Identifier {
                                                    base: BaseNode {
                                                        location: loc.get(1, 94, 1, 95),
                                                        ..BaseNode::default()
                                                    },
                                                    name: "r".to_string()
                                                }),
                                                lbrack: vec![],
                                                property: PropertyKey::StringLit(StringLit {
                                                    base: BaseNode {
                                                        location: loc.get(1, 96, 1, 103),
                                                        ..BaseNode::default()
                                                    },
                                                    value: "these".to_string()
                                                }),
                                                rbrack: vec![],
                                            })),
                                            right: Expression::StringLit(StringLit {
                                                base: BaseNode {
                                                    location: loc.get(1, 106, 1, 113),
                                                    ..BaseNode::default()
                                                },
                                                value: "those".to_string()
                                            })
                                        }))
                                    }
                                ))),
                            }))),
                            comma: vec![],
                        }],
                        rbrace: vec![],
                    }))],
                    rparen: vec![],
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn from_with_range() {
    let mut p = Parser::new(r#"from(bucket:"telegraf/autogen")|>range(start:-1h, end:10m)"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 59),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 59),
                    ..BaseNode::default()
                },
                expression: Expression::PipeExpr(Box::new(PipeExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 59),
                        ..BaseNode::default()
                    },
                    argument: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 32),
                            ..BaseNode::default()
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 5),
                                ..BaseNode::default()
                            },
                            name: "from".to_string()
                        }),
                        lparen: vec![],
                        arguments: vec![Expression::Object(Box::new(ObjectExpr {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 31),
                                ..BaseNode::default()
                            },
                            lbrace: vec![],
                            with: None,
                            properties: vec![Property {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 31),
                                    ..BaseNode::default()
                                },
                                key: PropertyKey::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 6, 1, 12),
                                        ..BaseNode::default()
                                    },
                                    name: "bucket".to_string()
                                }),
                                separator: vec![],
                                value: Some(Expression::StringLit(StringLit {
                                    base: BaseNode {
                                        location: loc.get(1, 13, 1, 31),
                                        ..BaseNode::default()
                                    },
                                    value: "telegraf/autogen".to_string()
                                })),
                                comma: vec![],
                            }],
                            rbrace: vec![],
                        }))],
                        rparen: vec![],
                    })),
                    call: CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 34, 1, 59),
                            ..BaseNode::default()
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 34, 1, 39),
                                ..BaseNode::default()
                            },
                            name: "range".to_string()
                        }),
                        lparen: vec![],
                        arguments: vec![Expression::Object(Box::new(ObjectExpr {
                            base: BaseNode {
                                location: loc.get(1, 40, 1, 58),
                                ..BaseNode::default()
                            },
                            lbrace: vec![],
                            with: None,
                            properties: vec![
                                Property {
                                    base: BaseNode {
                                        location: loc.get(1, 40, 1, 49),
                                        ..BaseNode::default()
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(1, 40, 1, 45),
                                            ..BaseNode::default()
                                        },
                                        name: "start".to_string()
                                    }),
                                    separator: vec![],
                                    value: Some(Expression::Unary(Box::new(UnaryExpr {
                                        base: BaseNode {
                                            location: loc.get(1, 46, 1, 49),
                                            ..BaseNode::default()
                                        },
                                        operator: Operator::SubtractionOperator,
                                        argument: Expression::Duration(DurationLit {
                                            base: BaseNode {
                                                location: loc.get(1, 47, 1, 49),
                                                ..BaseNode::default()
                                            },
                                            values: vec![Duration {
                                                magnitude: 1,
                                                unit: "h".to_string()
                                            }]
                                        })
                                    }))),
                                    comma: vec![],
                                },
                                Property {
                                    base: BaseNode {
                                        location: loc.get(1, 51, 1, 58),
                                        ..BaseNode::default()
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(1, 51, 1, 54),
                                            ..BaseNode::default()
                                        },
                                        name: "end".to_string()
                                    }),
                                    separator: vec![],
                                    value: Some(Expression::Duration(DurationLit {
                                        base: BaseNode {
                                            location: loc.get(1, 55, 1, 58),
                                            ..BaseNode::default()
                                        },
                                        values: vec![Duration {
                                            magnitude: 10,
                                            unit: "m".to_string()
                                        }]
                                    })),
                                    comma: vec![],
                                }
                            ],
                            rbrace: vec![],
                        }))],
                        rparen: vec![],
                    }
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn from_with_limit() {
    let mut p = Parser::new(r#"from(bucket:"telegraf/autogen")|>limit(limit:100, offset:10)"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 61),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 61),
                    ..BaseNode::default()
                },
                expression: Expression::PipeExpr(Box::new(PipeExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 61),
                        ..BaseNode::default()
                    },
                    argument: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 32),
                            ..BaseNode::default()
                        },
                        lparen: vec![],
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 5),
                                ..BaseNode::default()
                            },
                            name: "from".to_string()
                        }),
                        rparen: vec![],
                        arguments: vec![Expression::Object(Box::new(ObjectExpr {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 31),
                                ..BaseNode::default()
                            },
                            lbrace: vec![],
                            with: None,
                            properties: vec![Property {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 31),
                                    ..BaseNode::default()
                                },
                                key: PropertyKey::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 6, 1, 12),
                                        ..BaseNode::default()
                                    },
                                    name: "bucket".to_string()
                                }),
                                separator: vec![],
                                value: Some(Expression::StringLit(StringLit {
                                    base: BaseNode {
                                        location: loc.get(1, 13, 1, 31),
                                        ..BaseNode::default()
                                    },
                                    value: "telegraf/autogen".to_string()
                                })),
                                comma: vec![],
                            }],
                            rbrace: vec![],
                        }))]
                    })),
                    call: CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 34, 1, 61),
                            ..BaseNode::default()
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 34, 1, 39),
                                ..BaseNode::default()
                            },
                            name: "limit".to_string()
                        }),
                        lparen: vec![],
                        arguments: vec![Expression::Object(Box::new(ObjectExpr {
                            base: BaseNode {
                                location: loc.get(1, 40, 1, 60),
                                ..BaseNode::default()
                            },
                            lbrace: vec![],
                            with: None,
                            properties: vec![
                                Property {
                                    base: BaseNode {
                                        location: loc.get(1, 40, 1, 49),
                                        ..BaseNode::default()
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(1, 40, 1, 45),
                                            ..BaseNode::default()
                                        },
                                        name: "limit".to_string()
                                    }),
                                    separator: vec![],
                                    value: Some(Expression::Integer(IntegerLit {
                                        base: BaseNode {
                                            location: loc.get(1, 46, 1, 49),
                                            ..BaseNode::default()
                                        },
                                        value: 100
                                    })),
                                    comma: vec![],
                                },
                                Property {
                                    base: BaseNode {
                                        location: loc.get(1, 51, 1, 60),
                                        ..BaseNode::default()
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(1, 51, 1, 57),
                                            ..BaseNode::default()
                                        },
                                        name: "offset".to_string()
                                    }),
                                    separator: vec![],
                                    value: Some(Expression::Integer(IntegerLit {
                                        base: BaseNode {
                                            location: loc.get(1, 58, 1, 60),
                                            ..BaseNode::default()
                                        },
                                        value: 10
                                    })),
                                    comma: vec![],
                                }
                            ],
                            rbrace: vec![],
                        }))],
                        rparen: vec![],
                    }
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn from_with_range_and_count() {
    let mut p = Parser::new(
        r#"from(bucket:"mydb/autogen")
						|> range(start:-4h, stop:-2h)
						|> count()"#,
    );
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 3, 17),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 3, 17),
                    ..BaseNode::default()
                },
                expression: Expression::PipeExpr(Box::new(PipeExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 3, 17),
                        ..BaseNode::default()
                    },
                    argument: Expression::PipeExpr(Box::new(PipeExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 2, 36),
                            ..BaseNode::default()
                        },
                        argument: Expression::Call(Box::new(CallExpr {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 28),
                                ..BaseNode::default()
                            },
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 1, 1, 5),
                                    ..BaseNode::default()
                                },
                                name: "from".to_string()
                            }),
                            lparen: vec![],
                            arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 27),
                                    ..BaseNode::default()
                                },
                                lbrace: vec![],
                                with: None,
                                properties: vec![Property {
                                    base: BaseNode {
                                        location: loc.get(1, 6, 1, 27),
                                        ..BaseNode::default()
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(1, 6, 1, 12),
                                            ..BaseNode::default()
                                        },
                                        name: "bucket".to_string()
                                    }),
                                    separator: vec![],
                                    value: Some(Expression::StringLit(StringLit {
                                        base: BaseNode {
                                            location: loc.get(1, 13, 1, 27),
                                            ..BaseNode::default()
                                        },
                                        value: "mydb/autogen".to_string()
                                    })),
                                    comma: vec![],
                                }],
                                rbrace: vec![],
                            }))],
                            rparen: vec![],
                        })),
                        call: CallExpr {
                            base: BaseNode {
                                location: loc.get(2, 10, 2, 36),
                                ..BaseNode::default()
                            },
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(2, 10, 2, 15),
                                    ..BaseNode::default()
                                },
                                name: "range".to_string()
                            }),
                            lparen: vec![],
                            arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                base: BaseNode {
                                    location: loc.get(2, 16, 2, 35),
                                    ..BaseNode::default()
                                },
                                lbrace: vec![],
                                with: None,
                                properties: vec![
                                    Property {
                                        base: BaseNode {
                                            location: loc.get(2, 16, 2, 25),
                                            ..BaseNode::default()
                                        },
                                        key: PropertyKey::Identifier(Identifier {
                                            base: BaseNode {
                                                location: loc.get(2, 16, 2, 21),
                                                ..BaseNode::default()
                                            },
                                            name: "start".to_string()
                                        }),
                                        separator: vec![],
                                        value: Some(Expression::Unary(Box::new(UnaryExpr {
                                            base: BaseNode {
                                                location: loc.get(2, 22, 2, 25),
                                                ..BaseNode::default()
                                            },
                                            operator: Operator::SubtractionOperator,
                                            argument: Expression::Duration(DurationLit {
                                                base: BaseNode {
                                                    location: loc.get(2, 23, 2, 25),
                                                    ..BaseNode::default()
                                                },
                                                values: vec![Duration {
                                                    magnitude: 4,
                                                    unit: "h".to_string()
                                                }]
                                            })
                                        }))),
                                        comma: vec![],
                                    },
                                    Property {
                                        base: BaseNode {
                                            location: loc.get(2, 27, 2, 35),
                                            ..BaseNode::default()
                                        },
                                        key: PropertyKey::Identifier(Identifier {
                                            base: BaseNode {
                                                location: loc.get(2, 27, 2, 31),
                                                ..BaseNode::default()
                                            },
                                            name: "stop".to_string()
                                        }),
                                        separator: vec![],
                                        value: Some(Expression::Unary(Box::new(UnaryExpr {
                                            base: BaseNode {
                                                location: loc.get(2, 32, 2, 35),
                                                ..BaseNode::default()
                                            },
                                            operator: Operator::SubtractionOperator,
                                            argument: Expression::Duration(DurationLit {
                                                base: BaseNode {
                                                    location: loc.get(2, 33, 2, 35),
                                                    ..BaseNode::default()
                                                },
                                                values: vec![Duration {
                                                    magnitude: 2,
                                                    unit: "h".to_string()
                                                }]
                                            })
                                        }))),
                                        comma: vec![],
                                    }
                                ],
                                rbrace: vec![],
                            }))],
                            rparen: vec![],
                        }
                    })),
                    call: CallExpr {
                        base: BaseNode {
                            location: loc.get(3, 10, 3, 17),
                            ..BaseNode::default()
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(3, 10, 3, 15),
                                ..BaseNode::default()
                            },
                            name: "count".to_string()
                        }),
                        lparen: vec![],
                        arguments: vec![],
                        rparen: vec![],
                    }
                }))
            }))],
            eof: vec![],
        }
    )
}

#[test]
fn from_with_range_limit_and_count() {
    let mut p = Parser::new(
        r#"from(bucket:"mydb/autogen")
						|> range(start:-4h, stop:-2h)
						|> limit(n:10)
						|> count()"#,
    );
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 4, 17),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 4, 17),
                    ..BaseNode::default()
                },
                expression: Expression::PipeExpr(Box::new(PipeExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 4, 17),
                        ..BaseNode::default()
                    },
                    argument: Expression::PipeExpr(Box::new(PipeExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 3, 21),
                            ..BaseNode::default()
                        },
                        argument: Expression::PipeExpr(Box::new(PipeExpr {
                            base: BaseNode {
                                location: loc.get(1, 1, 2, 36),
                                ..BaseNode::default()
                            },
                            argument: Expression::Call(Box::new(CallExpr {
                                base: BaseNode {
                                    location: loc.get(1, 1, 1, 28),
                                    ..BaseNode::default()
                                },
                                callee: Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 1, 1, 5),
                                        ..BaseNode::default()
                                    },
                                    name: "from".to_string()
                                }),
                                lparen: vec![],
                                arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                    base: BaseNode {
                                        location: loc.get(1, 6, 1, 27),
                                        ..BaseNode::default()
                                    },
                                    lbrace: vec![],
                                    with: None,
                                    properties: vec![Property {
                                        base: BaseNode {
                                            location: loc.get(1, 6, 1, 27),
                                            ..BaseNode::default()
                                        },
                                        key: PropertyKey::Identifier(Identifier {
                                            base: BaseNode {
                                                location: loc.get(1, 6, 1, 12),
                                                ..BaseNode::default()
                                            },
                                            name: "bucket".to_string()
                                        }),
                                        separator: vec![],
                                        value: Some(Expression::StringLit(StringLit {
                                            base: BaseNode {
                                                location: loc.get(1, 13, 1, 27),
                                                ..BaseNode::default()
                                            },
                                            value: "mydb/autogen".to_string()
                                        })),
                                        comma: vec![],
                                    }],
                                    rbrace: vec![],
                                }))],
                                rparen: vec![],
                            })),
                            call: CallExpr {
                                base: BaseNode {
                                    location: loc.get(2, 10, 2, 36),
                                    ..BaseNode::default()
                                },
                                callee: Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(2, 10, 2, 15),
                                        ..BaseNode::default()
                                    },
                                    name: "range".to_string()
                                }),
                                lparen: vec![],
                                arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                    base: BaseNode {
                                        location: loc.get(2, 16, 2, 35),
                                        ..BaseNode::default()
                                    },
                                    lbrace: vec![],
                                    with: None,
                                    properties: vec![
                                        Property {
                                            base: BaseNode {
                                                location: loc.get(2, 16, 2, 25),
                                                ..BaseNode::default()
                                            },
                                            key: PropertyKey::Identifier(Identifier {
                                                base: BaseNode {
                                                    location: loc.get(2, 16, 2, 21),
                                                    ..BaseNode::default()
                                                },
                                                name: "start".to_string()
                                            }),
                                            separator: vec![],
                                            value: Some(Expression::Unary(Box::new(UnaryExpr {
                                                base: BaseNode {
                                                    location: loc.get(2, 22, 2, 25),
                                                    ..BaseNode::default()
                                                },
                                                operator: Operator::SubtractionOperator,
                                                argument: Expression::Duration(DurationLit {
                                                    base: BaseNode {
                                                        location: loc.get(2, 23, 2, 25),
                                                        ..BaseNode::default()
                                                    },
                                                    values: vec![Duration {
                                                        magnitude: 4,
                                                        unit: "h".to_string()
                                                    }]
                                                })
                                            }))),
                                            comma: vec![],
                                        },
                                        Property {
                                            base: BaseNode {
                                                location: loc.get(2, 27, 2, 35),
                                                ..BaseNode::default()
                                            },
                                            key: PropertyKey::Identifier(Identifier {
                                                base: BaseNode {
                                                    location: loc.get(2, 27, 2, 31),
                                                    ..BaseNode::default()
                                                },
                                                name: "stop".to_string()
                                            }),
                                            separator: vec![],
                                            value: Some(Expression::Unary(Box::new(UnaryExpr {
                                                base: BaseNode {
                                                    location: loc.get(2, 32, 2, 35),
                                                    ..BaseNode::default()
                                                },
                                                operator: Operator::SubtractionOperator,
                                                argument: Expression::Duration(DurationLit {
                                                    base: BaseNode {
                                                        location: loc.get(2, 33, 2, 35),
                                                        ..BaseNode::default()
                                                    },
                                                    values: vec![Duration {
                                                        magnitude: 2,
                                                        unit: "h".to_string()
                                                    }]
                                                })
                                            }))),
                                            comma: vec![],
                                        }
                                    ],
                                    rbrace: vec![],
                                }))],
                                rparen: vec![],
                            }
                        })),
                        call: CallExpr {
                            base: BaseNode {
                                location: loc.get(3, 10, 3, 21),
                                ..BaseNode::default()
                            },
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(3, 10, 3, 15),
                                    ..BaseNode::default()
                                },
                                name: "limit".to_string()
                            }),
                            lparen: vec![],
                            arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                base: BaseNode {
                                    location: loc.get(3, 16, 3, 20),
                                    ..BaseNode::default()
                                },
                                lbrace: vec![],
                                with: None,
                                properties: vec![Property {
                                    base: BaseNode {
                                        location: loc.get(3, 16, 3, 20),
                                        ..BaseNode::default()
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(3, 16, 3, 17),
                                            ..BaseNode::default()
                                        },
                                        name: "n".to_string()
                                    }),
                                    separator: vec![],
                                    value: Some(Expression::Integer(IntegerLit {
                                        base: BaseNode {
                                            location: loc.get(3, 18, 3, 20),
                                            ..BaseNode::default()
                                        },
                                        value: 10
                                    })),
                                    comma: vec![],
                                }],
                                rbrace: vec![],
                            }))],
                            rparen: vec![],
                        }
                    })),
                    call: CallExpr {
                        base: BaseNode {
                            location: loc.get(4, 10, 4, 17),
                            ..BaseNode::default()
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(4, 10, 4, 15),
                                ..BaseNode::default()
                            },
                            name: "count".to_string()
                        }),
                        lparen: vec![],
                        arguments: vec![],
                        rparen: vec![],
                    }
                }))
            }))],
            eof: vec![],
        }
    )
}

#[test]
fn from_with_join() {
    let mut p = Parser::new(
        r#"
a = from(bucket:"dbA/autogen") |> range(start:-1h)
b = from(bucket:"dbB/autogen") |> range(start:-1h)
join(tables:[a,b], on:["host"], fn: (a,b) => a["_field"] + b["_field"])"#,
    );
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(2, 1, 4, 72),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(2, 1, 2, 51),
                        ..BaseNode::default()
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(2, 1, 2, 2),
                            ..BaseNode::default()
                        },
                        name: "a".to_string()
                    },
                    init: Expression::PipeExpr(Box::new(PipeExpr {
                        base: BaseNode {
                            location: loc.get(2, 5, 2, 51),
                            ..BaseNode::default()
                        },
                        argument: Expression::Call(Box::new(CallExpr {
                            base: BaseNode {
                                location: loc.get(2, 5, 2, 31),
                                ..BaseNode::default()
                            },
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(2, 5, 2, 9),
                                    ..BaseNode::default()
                                },
                                name: "from".to_string()
                            }),
                            lparen: vec![],
                            arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                base: BaseNode {
                                    location: loc.get(2, 10, 2, 30),
                                    ..BaseNode::default()
                                },
                                lbrace: vec![],
                                with: None,
                                properties: vec![Property {
                                    base: BaseNode {
                                        location: loc.get(2, 10, 2, 30),
                                        ..BaseNode::default()
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(2, 10, 2, 16),
                                            ..BaseNode::default()
                                        },
                                        name: "bucket".to_string()
                                    }),
                                    separator: vec![],
                                    value: Some(Expression::StringLit(StringLit {
                                        base: BaseNode {
                                            location: loc.get(2, 17, 2, 30),
                                            ..BaseNode::default()
                                        },
                                        value: "dbA/autogen".to_string()
                                    })),
                                    comma: vec![],
                                }],
                                rbrace: vec![],
                            }))],
                            rparen: vec![],
                        })),
                        call: CallExpr {
                            base: BaseNode {
                                location: loc.get(2, 35, 2, 51),
                                ..BaseNode::default()
                            },
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(2, 35, 2, 40),
                                    ..BaseNode::default()
                                },
                                name: "range".to_string()
                            }),
                            lparen: vec![],
                            arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                base: BaseNode {
                                    location: loc.get(2, 41, 2, 50),
                                    ..BaseNode::default()
                                },
                                lbrace: vec![],
                                with: None,
                                properties: vec![Property {
                                    base: BaseNode {
                                        location: loc.get(2, 41, 2, 50),
                                        ..BaseNode::default()
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(2, 41, 2, 46),
                                            ..BaseNode::default()
                                        },
                                        name: "start".to_string()
                                    }),
                                    separator: vec![],
                                    value: Some(Expression::Unary(Box::new(UnaryExpr {
                                        base: BaseNode {
                                            location: loc.get(2, 47, 2, 50),
                                            ..BaseNode::default()
                                        },
                                        operator: Operator::SubtractionOperator,
                                        argument: Expression::Duration(DurationLit {
                                            base: BaseNode {
                                                location: loc.get(2, 48, 2, 50),
                                                ..BaseNode::default()
                                            },
                                            values: vec![Duration {
                                                magnitude: 1,
                                                unit: "h".to_string()
                                            }]
                                        })
                                    }))),
                                    comma: vec![],
                                }],
                                rbrace: vec![],
                            }))],
                            rparen: vec![],
                        }
                    }))
                })),
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(3, 1, 3, 51),
                        ..BaseNode::default()
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(3, 1, 3, 2),
                            ..BaseNode::default()
                        },
                        name: "b".to_string()
                    },
                    init: Expression::PipeExpr(Box::new(PipeExpr {
                        base: BaseNode {
                            location: loc.get(3, 5, 3, 51),
                            ..BaseNode::default()
                        },
                        argument: Expression::Call(Box::new(CallExpr {
                            base: BaseNode {
                                location: loc.get(3, 5, 3, 31),
                                ..BaseNode::default()
                            },
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(3, 5, 3, 9),
                                    ..BaseNode::default()
                                },
                                name: "from".to_string()
                            }),
                            lparen: vec![],
                            arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                base: BaseNode {
                                    location: loc.get(3, 10, 3, 30),
                                    ..BaseNode::default()
                                },
                                lbrace: vec![],
                                with: None,
                                properties: vec![Property {
                                    base: BaseNode {
                                        location: loc.get(3, 10, 3, 30),
                                        ..BaseNode::default()
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(3, 10, 3, 16),
                                            ..BaseNode::default()
                                        },
                                        name: "bucket".to_string()
                                    }),
                                    separator: vec![],
                                    value: Some(Expression::StringLit(StringLit {
                                        base: BaseNode {
                                            location: loc.get(3, 17, 3, 30),
                                            ..BaseNode::default()
                                        },
                                        value: "dbB/autogen".to_string()
                                    })),
                                    comma: vec![],
                                }],
                                rbrace: vec![],
                            }))],
                            rparen: vec![],
                        })),
                        call: CallExpr {
                            base: BaseNode {
                                location: loc.get(3, 35, 3, 51),
                                ..BaseNode::default()
                            },
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(3, 35, 3, 40),
                                    ..BaseNode::default()
                                },
                                name: "range".to_string()
                            }),
                            lparen: vec![],
                            arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                base: BaseNode {
                                    location: loc.get(3, 41, 3, 50),
                                    ..BaseNode::default()
                                },
                                lbrace: vec![],
                                with: None,
                                properties: vec![Property {
                                    base: BaseNode {
                                        location: loc.get(3, 41, 3, 50),
                                        ..BaseNode::default()
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(3, 41, 3, 46),
                                            ..BaseNode::default()
                                        },
                                        name: "start".to_string()
                                    }),
                                    separator: vec![],
                                    value: Some(Expression::Unary(Box::new(UnaryExpr {
                                        base: BaseNode {
                                            location: loc.get(3, 47, 3, 50),
                                            ..BaseNode::default()
                                        },
                                        operator: Operator::SubtractionOperator,
                                        argument: Expression::Duration(DurationLit {
                                            base: BaseNode {
                                                location: loc.get(3, 48, 3, 50),
                                                ..BaseNode::default()
                                            },
                                            values: vec![Duration {
                                                magnitude: 1,
                                                unit: "h".to_string()
                                            }]
                                        })
                                    }))),
                                    comma: vec![],
                                }],
                                rbrace: vec![],
                            }))],
                            rparen: vec![],
                        }
                    }))
                })),
                Statement::Expr(Box::new(ExprStmt {
                    base: BaseNode {
                        location: loc.get(4, 1, 4, 72),
                        ..BaseNode::default()
                    },
                    expression: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(4, 1, 4, 72),
                            ..BaseNode::default()
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(4, 1, 4, 5),
                                ..BaseNode::default()
                            },
                            name: "join".to_string()
                        }),
                        lparen: vec![],
                        arguments: vec![Expression::Object(Box::new(ObjectExpr {
                            base: BaseNode {
                                location: loc.get(4, 6, 4, 71),
                                ..BaseNode::default()
                            },
                            lbrace: vec![],
                            with: None,
                            properties: vec![
                                Property {
                                    base: BaseNode {
                                        location: loc.get(4, 6, 4, 18),
                                        ..BaseNode::default()
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(4, 6, 4, 12),
                                            ..BaseNode::default()
                                        },
                                        name: "tables".to_string()
                                    }),
                                    separator: vec![],
                                    value: Some(Expression::Array(Box::new(ArrayExpr {
                                        base: BaseNode {
                                            location: loc.get(4, 13, 4, 18),
                                            ..BaseNode::default()
                                        },
                                        lbrack: vec![],
                                        elements: vec![
                                            ArrayItem {
                                                expression: Expression::Identifier(Identifier {
                                                    base: BaseNode {
                                                        location: loc.get(4, 14, 4, 15),
                                                        ..BaseNode::default()
                                                    },
                                                    name: "a".to_string()
                                                }),
                                                comma: vec![],
                                            },
                                            ArrayItem {
                                                expression: Expression::Identifier(Identifier {
                                                    base: BaseNode {
                                                        location: loc.get(4, 16, 4, 17),
                                                        ..BaseNode::default()
                                                    },
                                                    name: "b".to_string()
                                                }),
                                                comma: vec![],
                                            }
                                        ],
                                        rbrack: vec![],
                                    }))),
                                    comma: vec![],
                                },
                                Property {
                                    base: BaseNode {
                                        location: loc.get(4, 20, 4, 31),
                                        ..BaseNode::default()
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(4, 20, 4, 22),
                                            ..BaseNode::default()
                                        },
                                        name: "on".to_string()
                                    }),
                                    separator: vec![],
                                    value: Some(Expression::Array(Box::new(ArrayExpr {
                                        base: BaseNode {
                                            location: loc.get(4, 23, 4, 31),
                                            ..BaseNode::default()
                                        },
                                        lbrack: vec![],
                                        elements: vec![ArrayItem {
                                            expression: Expression::StringLit(StringLit {
                                                base: BaseNode {
                                                    location: loc.get(4, 24, 4, 30),
                                                    ..BaseNode::default()
                                                },
                                                value: "host".to_string()
                                            }),
                                            comma: vec![],
                                        }],
                                        rbrack: vec![],
                                    }))),
                                    comma: vec![],
                                },
                                Property {
                                    base: BaseNode {
                                        location: loc.get(4, 33, 4, 71),
                                        ..BaseNode::default()
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(4, 33, 4, 35),
                                            ..BaseNode::default()
                                        },
                                        name: "fn".to_string()
                                    }),
                                    separator: vec![],
                                    value: Some(Expression::Function(Box::new(FunctionExpr {
                                        base: BaseNode {
                                            location: loc.get(4, 37, 4, 71),
                                            ..BaseNode::default()
                                        },
                                        lparen: vec![],
                                        params: vec![
                                            Property {
                                                base: BaseNode {
                                                    location: loc.get(4, 38, 4, 39),
                                                    ..BaseNode::default()
                                                },
                                                key: PropertyKey::Identifier(Identifier {
                                                    base: BaseNode {
                                                        location: loc.get(4, 38, 4, 39),
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
                                                    location: loc.get(4, 40, 4, 41),
                                                    ..BaseNode::default()
                                                },
                                                key: PropertyKey::Identifier(Identifier {
                                                    base: BaseNode {
                                                        location: loc.get(4, 40, 4, 41),
                                                        ..BaseNode::default()
                                                    },
                                                    name: "b".to_string()
                                                }),
                                                separator: vec![],
                                                value: None,
                                                comma: vec![],
                                            }
                                        ],
                                        rparen: vec![],
                                        arrow: vec![],
                                        body: FunctionBody::Expr(Expression::Binary(Box::new(
                                            BinaryExpr {
                                                base: BaseNode {
                                                    location: loc.get(4, 46, 4, 71),
                                                    ..BaseNode::default()
                                                },
                                                operator: Operator::AdditionOperator,
                                                left: Expression::Member(Box::new(MemberExpr {
                                                    base: BaseNode {
                                                        location: loc.get(4, 46, 4, 57),
                                                        ..BaseNode::default()
                                                    },
                                                    object: Expression::Identifier(Identifier {
                                                        base: BaseNode {
                                                            location: loc.get(4, 46, 4, 47),
                                                            ..BaseNode::default()
                                                        },
                                                        name: "a".to_string()
                                                    }),
                                                    lbrack: vec![],
                                                    property: PropertyKey::StringLit(StringLit {
                                                        base: BaseNode {
                                                            location: loc.get(4, 48, 4, 56),
                                                            ..BaseNode::default()
                                                        },
                                                        value: "_field".to_string()
                                                    }),
                                                    rbrack: vec![],
                                                })),
                                                right: Expression::Member(Box::new(MemberExpr {
                                                    base: BaseNode {
                                                        location: loc.get(4, 60, 4, 71),
                                                        ..BaseNode::default()
                                                    },
                                                    object: Expression::Identifier(Identifier {
                                                        base: BaseNode {
                                                            location: loc.get(4, 60, 4, 61),
                                                            ..BaseNode::default()
                                                        },
                                                        name: "b".to_string()
                                                    }),
                                                    lbrack: vec![],
                                                    property: PropertyKey::StringLit(StringLit {
                                                        base: BaseNode {
                                                            location: loc.get(4, 62, 4, 70),
                                                            ..BaseNode::default()
                                                        },
                                                        value: "_field".to_string()
                                                    }),
                                                    rbrack: vec![],
                                                }))
                                            }
                                        ))),
                                    }))),
                                    comma: vec![],
                                }
                            ],
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
fn from_with_join_with_complex_expression() {
    let mut p = Parser::new(
        r#"
a = from(bucket:"Flux/autogen")
	|> filter(fn: (r) => r["_measurement"] == "a")
	|> range(start:-1h)

b = from(bucket:"Flux/autogen")
	|> filter(fn: (r) => r["_measurement"] == "b")
	|> range(start:-1h)

join(tables:[a,b], on:["t1"], fn: (a,b) => (a["_field"] - b["_field"]) / b["_field"])"#,
    );
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);

    let stmt_a = Statement::Variable(Box::new(VariableAssgn {
        base: BaseNode {
            location: loc.get(2, 1, 4, 21),
            ..BaseNode::default()
        },
        id: Identifier {
            base: BaseNode {
                location: loc.get(2, 1, 2, 2),
                ..BaseNode::default()
            },
            name: "a".to_string(),
        },
        init: Expression::PipeExpr(Box::new(PipeExpr {
            base: BaseNode {
                location: loc.get(2, 5, 4, 21),
                ..BaseNode::default()
            },
            argument: Expression::PipeExpr(Box::new(PipeExpr {
                base: BaseNode {
                    location: loc.get(2, 5, 3, 48),
                    ..BaseNode::default()
                },
                argument: Expression::Call(Box::new(CallExpr {
                    base: BaseNode {
                        location: loc.get(2, 5, 2, 32),
                        ..BaseNode::default()
                    },
                    callee: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(2, 5, 2, 9),
                            ..BaseNode::default()
                        },
                        name: "from".to_string(),
                    }),
                    lparen: vec![],
                    arguments: vec![Expression::Object(Box::new(ObjectExpr {
                        base: BaseNode {
                            location: loc.get(2, 10, 2, 31),
                            ..BaseNode::default()
                        },
                        lbrace: vec![],
                        with: None,
                        properties: vec![Property {
                            base: BaseNode {
                                location: loc.get(2, 10, 2, 31),
                                ..BaseNode::default()
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(2, 10, 2, 16),
                                    ..BaseNode::default()
                                },
                                name: "bucket".to_string(),
                            }),
                            separator: vec![],
                            value: Some(Expression::StringLit(StringLit {
                                base: BaseNode {
                                    location: loc.get(2, 17, 2, 31),
                                    ..BaseNode::default()
                                },
                                value: "Flux/autogen".to_string(),
                            })),
                            comma: vec![],
                        }],
                        rbrace: vec![],
                    }))],
                    rparen: vec![],
                })),
                call: CallExpr {
                    base: BaseNode {
                        location: loc.get(3, 5, 3, 48),
                        ..BaseNode::default()
                    },
                    callee: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(3, 5, 3, 11),
                            ..BaseNode::default()
                        },
                        name: "filter".to_string(),
                    }),
                    lparen: vec![],
                    arguments: vec![Expression::Object(Box::new(ObjectExpr {
                        base: BaseNode {
                            location: loc.get(3, 12, 3, 47),
                            ..BaseNode::default()
                        },
                        lbrace: vec![],
                        with: None,
                        properties: vec![Property {
                            base: BaseNode {
                                location: loc.get(3, 12, 3, 47),
                                ..BaseNode::default()
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(3, 12, 3, 14),
                                    ..BaseNode::default()
                                },
                                name: "fn".to_string(),
                            }),
                            separator: vec![],
                            value: Some(Expression::Function(Box::new(FunctionExpr {
                                base: BaseNode {
                                    location: loc.get(3, 16, 3, 47),
                                    ..BaseNode::default()
                                },
                                lparen: vec![],
                                params: vec![Property {
                                    base: BaseNode {
                                        location: loc.get(3, 17, 3, 18),
                                        ..BaseNode::default()
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(3, 17, 3, 18),
                                            ..BaseNode::default()
                                        },
                                        name: "r".to_string(),
                                    }),
                                    separator: vec![],
                                    value: None,
                                    comma: vec![],
                                }],
                                rparen: vec![],
                                arrow: vec![],
                                body: FunctionBody::Expr(Expression::Binary(Box::new(
                                    BinaryExpr {
                                        base: BaseNode {
                                            location: loc.get(3, 23, 3, 47),
                                            ..BaseNode::default()
                                        },
                                        operator: Operator::EqualOperator,
                                        left: Expression::Member(Box::new(MemberExpr {
                                            base: BaseNode {
                                                location: loc.get(3, 23, 3, 40),
                                                ..BaseNode::default()
                                            },
                                            object: Expression::Identifier(Identifier {
                                                base: BaseNode {
                                                    location: loc.get(3, 23, 3, 24),
                                                    ..BaseNode::default()
                                                },
                                                name: "r".to_string(),
                                            }),
                                            lbrack: vec![],
                                            property: PropertyKey::StringLit(StringLit {
                                                base: BaseNode {
                                                    location: loc.get(3, 25, 3, 39),
                                                    ..BaseNode::default()
                                                },
                                                value: "_measurement".to_string(),
                                            }),
                                            rbrack: vec![],
                                        })),
                                        right: Expression::StringLit(StringLit {
                                            base: BaseNode {
                                                location: loc.get(3, 44, 3, 47),
                                                ..BaseNode::default()
                                            },
                                            value: "a".to_string(),
                                        }),
                                    },
                                ))),
                            }))),
                            comma: vec![],
                        }],
                        rbrace: vec![],
                    }))],
                    rparen: vec![],
                },
            })),
            call: CallExpr {
                base: BaseNode {
                    location: loc.get(4, 5, 4, 21),
                    ..BaseNode::default()
                },
                callee: Expression::Identifier(Identifier {
                    base: BaseNode {
                        location: loc.get(4, 5, 4, 10),
                        ..BaseNode::default()
                    },
                    name: "range".to_string(),
                }),
                lparen: vec![],
                arguments: vec![Expression::Object(Box::new(ObjectExpr {
                    base: BaseNode {
                        location: loc.get(4, 11, 4, 20),
                        ..BaseNode::default()
                    },
                    lbrace: vec![],
                    with: None,
                    properties: vec![Property {
                        base: BaseNode {
                            location: loc.get(4, 11, 4, 20),
                            ..BaseNode::default()
                        },
                        key: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(4, 11, 4, 16),
                                ..BaseNode::default()
                            },
                            name: "start".to_string(),
                        }),
                        separator: vec![],
                        value: Some(Expression::Unary(Box::new(UnaryExpr {
                            base: BaseNode {
                                location: loc.get(4, 17, 4, 20),
                                ..BaseNode::default()
                            },
                            operator: Operator::SubtractionOperator,
                            argument: Expression::Duration(DurationLit {
                                base: BaseNode {
                                    location: loc.get(4, 18, 4, 20),
                                    ..BaseNode::default()
                                },
                                values: vec![Duration {
                                    magnitude: 1,
                                    unit: "h".to_string(),
                                }],
                            }),
                        }))),
                        comma: vec![],
                    }],
                    rbrace: vec![],
                }))],
                rparen: vec![],
            },
        })),
    }));

    let stmt_b = Statement::Variable(Box::new(VariableAssgn {
        base: BaseNode {
            location: loc.get(6, 1, 8, 21),
            ..BaseNode::default()
        },
        id: Identifier {
            base: BaseNode {
                location: loc.get(6, 1, 6, 2),
                ..BaseNode::default()
            },
            name: "b".to_string(),
        },
        init: Expression::PipeExpr(Box::new(PipeExpr {
            base: BaseNode {
                location: loc.get(6, 5, 8, 21),
                ..BaseNode::default()
            },
            argument: Expression::PipeExpr(Box::new(PipeExpr {
                base: BaseNode {
                    location: loc.get(6, 5, 7, 48),
                    ..BaseNode::default()
                },
                argument: Expression::Call(Box::new(CallExpr {
                    base: BaseNode {
                        location: loc.get(6, 5, 6, 32),
                        ..BaseNode::default()
                    },
                    callee: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(6, 5, 6, 9),
                            ..BaseNode::default()
                        },
                        name: "from".to_string(),
                    }),
                    lparen: vec![],
                    arguments: vec![Expression::Object(Box::new(ObjectExpr {
                        base: BaseNode {
                            location: loc.get(6, 10, 6, 31),
                            ..BaseNode::default()
                        },
                        lbrace: vec![],
                        with: None,
                        properties: vec![Property {
                            base: BaseNode {
                                location: loc.get(6, 10, 6, 31),
                                ..BaseNode::default()
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(6, 10, 6, 16),
                                    ..BaseNode::default()
                                },
                                name: "bucket".to_string(),
                            }),
                            separator: vec![],
                            value: Some(Expression::StringLit(StringLit {
                                base: BaseNode {
                                    location: loc.get(6, 17, 6, 31),
                                    ..BaseNode::default()
                                },
                                value: "Flux/autogen".to_string(),
                            })),
                            comma: vec![],
                        }],
                        rbrace: vec![],
                    }))],
                    rparen: vec![],
                })),
                call: CallExpr {
                    base: BaseNode {
                        location: loc.get(7, 5, 7, 48),
                        ..BaseNode::default()
                    },
                    callee: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(7, 5, 7, 11),
                            ..BaseNode::default()
                        },
                        name: "filter".to_string(),
                    }),
                    lparen: vec![],
                    arguments: vec![Expression::Object(Box::new(ObjectExpr {
                        base: BaseNode {
                            location: loc.get(7, 12, 7, 47),
                            ..BaseNode::default()
                        },
                        lbrace: vec![],
                        with: None,
                        properties: vec![Property {
                            base: BaseNode {
                                location: loc.get(7, 12, 7, 47),
                                ..BaseNode::default()
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(7, 12, 7, 14),
                                    ..BaseNode::default()
                                },
                                name: "fn".to_string(),
                            }),
                            separator: vec![],
                            value: Some(Expression::Function(Box::new(FunctionExpr {
                                base: BaseNode {
                                    location: loc.get(7, 16, 7, 47),
                                    ..BaseNode::default()
                                },
                                lparen: vec![],
                                params: vec![Property {
                                    base: BaseNode {
                                        location: loc.get(7, 17, 7, 18),
                                        ..BaseNode::default()
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(7, 17, 7, 18),
                                            ..BaseNode::default()
                                        },
                                        name: "r".to_string(),
                                    }),
                                    separator: vec![],
                                    value: None,
                                    comma: vec![],
                                }],
                                rparen: vec![],
                                arrow: vec![],
                                body: FunctionBody::Expr(Expression::Binary(Box::new(
                                    BinaryExpr {
                                        base: BaseNode {
                                            location: loc.get(7, 23, 7, 47),
                                            ..BaseNode::default()
                                        },
                                        operator: Operator::EqualOperator,
                                        left: Expression::Member(Box::new(MemberExpr {
                                            base: BaseNode {
                                                location: loc.get(7, 23, 7, 40),
                                                ..BaseNode::default()
                                            },
                                            object: Expression::Identifier(Identifier {
                                                base: BaseNode {
                                                    location: loc.get(7, 23, 7, 24),
                                                    ..BaseNode::default()
                                                },
                                                name: "r".to_string(),
                                            }),
                                            lbrack: vec![],
                                            property: PropertyKey::StringLit(StringLit {
                                                base: BaseNode {
                                                    location: loc.get(7, 25, 7, 39),
                                                    ..BaseNode::default()
                                                },
                                                value: "_measurement".to_string(),
                                            }),
                                            rbrack: vec![],
                                        })),
                                        right: Expression::StringLit(StringLit {
                                            base: BaseNode {
                                                location: loc.get(7, 44, 7, 47),
                                                ..BaseNode::default()
                                            },
                                            value: "b".to_string(),
                                        }),
                                    },
                                ))),
                            }))),
                            comma: vec![],
                        }],
                        rbrace: vec![],
                    }))],
                    rparen: vec![],
                },
            })),
            call: CallExpr {
                base: BaseNode {
                    location: loc.get(8, 5, 8, 21),
                    ..BaseNode::default()
                },
                callee: Expression::Identifier(Identifier {
                    base: BaseNode {
                        location: loc.get(8, 5, 8, 10),
                        ..BaseNode::default()
                    },
                    name: "range".to_string(),
                }),
                lparen: vec![],
                arguments: vec![Expression::Object(Box::new(ObjectExpr {
                    base: BaseNode {
                        location: loc.get(8, 11, 8, 20),
                        ..BaseNode::default()
                    },
                    lbrace: vec![],
                    with: None,
                    properties: vec![Property {
                        base: BaseNode {
                            location: loc.get(8, 11, 8, 20),
                            ..BaseNode::default()
                        },
                        key: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(8, 11, 8, 16),
                                ..BaseNode::default()
                            },
                            name: "start".to_string(),
                        }),
                        separator: vec![],
                        value: Some(Expression::Unary(Box::new(UnaryExpr {
                            base: BaseNode {
                                location: loc.get(8, 17, 8, 20),
                                ..BaseNode::default()
                            },
                            operator: Operator::SubtractionOperator,
                            argument: Expression::Duration(DurationLit {
                                base: BaseNode {
                                    location: loc.get(8, 18, 8, 20),
                                    ..BaseNode::default()
                                },
                                values: vec![Duration {
                                    magnitude: 1,
                                    unit: "h".to_string(),
                                }],
                            }),
                        }))),
                        comma: vec![],
                    }],
                    rbrace: vec![],
                }))],
                rparen: vec![],
            },
        })),
    }));

    let stmt_expr = Statement::Expr(Box::new(ExprStmt {
        base: BaseNode {
            location: loc.get(10, 1, 10, 86),
            ..BaseNode::default()
        },
        expression: Expression::Call(Box::new(CallExpr {
            base: BaseNode {
                location: loc.get(10, 1, 10, 86),
                ..BaseNode::default()
            },
            callee: Expression::Identifier(Identifier {
                base: BaseNode {
                    location: loc.get(10, 1, 10, 5),
                    ..BaseNode::default()
                },
                name: "join".to_string(),
            }),
            lparen: vec![],
            arguments: vec![Expression::Object(Box::new(ObjectExpr {
                base: BaseNode {
                    location: loc.get(10, 6, 10, 85),
                    ..BaseNode::default()
                },
                lbrace: vec![],
                with: None,
                properties: vec![
                    Property {
                        base: BaseNode {
                            location: loc.get(10, 6, 10, 18),
                            ..BaseNode::default()
                        },
                        key: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(10, 6, 10, 12),
                                ..BaseNode::default()
                            },
                            name: "tables".to_string(),
                        }),
                        separator: vec![],
                        value: Some(Expression::Array(Box::new(ArrayExpr {
                            base: BaseNode {
                                location: loc.get(10, 13, 10, 18),
                                ..BaseNode::default()
                            },
                            lbrack: vec![],
                            elements: vec![
                                ArrayItem {
                                    expression: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(10, 14, 10, 15),
                                            ..BaseNode::default()
                                        },
                                        name: "a".to_string(),
                                    }),
                                    comma: vec![],
                                },
                                ArrayItem {
                                    expression: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(10, 16, 10, 17),
                                            ..BaseNode::default()
                                        },
                                        name: "b".to_string(),
                                    }),
                                    comma: vec![],
                                },
                            ],
                            rbrack: vec![],
                        }))),
                        comma: vec![],
                    },
                    Property {
                        base: BaseNode {
                            location: loc.get(10, 20, 10, 29),
                            ..BaseNode::default()
                        },
                        key: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(10, 20, 10, 22),
                                ..BaseNode::default()
                            },
                            name: "on".to_string(),
                        }),
                        separator: vec![],
                        value: Some(Expression::Array(Box::new(ArrayExpr {
                            base: BaseNode {
                                location: loc.get(10, 23, 10, 29),
                                ..BaseNode::default()
                            },
                            lbrack: vec![],
                            elements: vec![ArrayItem {
                                expression: Expression::StringLit(StringLit {
                                    base: BaseNode {
                                        location: loc.get(10, 24, 10, 28),
                                        ..BaseNode::default()
                                    },
                                    value: "t1".to_string(),
                                }),
                                comma: vec![],
                            }],
                            rbrack: vec![],
                        }))),
                        comma: vec![],
                    },
                    Property {
                        base: BaseNode {
                            location: loc.get(10, 31, 10, 85),
                            ..BaseNode::default()
                        },
                        key: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(10, 31, 10, 33),
                                ..BaseNode::default()
                            },
                            name: "fn".to_string(),
                        }),
                        separator: vec![],
                        value: Some(Expression::Function(Box::new(FunctionExpr {
                            base: BaseNode {
                                location: loc.get(10, 35, 10, 85),
                                ..BaseNode::default()
                            },
                            lparen: vec![],
                            params: vec![
                                Property {
                                    base: BaseNode {
                                        location: loc.get(10, 36, 10, 37),
                                        ..BaseNode::default()
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(10, 36, 10, 37),
                                            ..BaseNode::default()
                                        },
                                        name: "a".to_string(),
                                    }),
                                    separator: vec![],
                                    value: None,
                                    comma: vec![],
                                },
                                Property {
                                    base: BaseNode {
                                        location: loc.get(10, 38, 10, 39),
                                        ..BaseNode::default()
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(10, 38, 10, 39),
                                            ..BaseNode::default()
                                        },
                                        name: "b".to_string(),
                                    }),
                                    separator: vec![],
                                    value: None,
                                    comma: vec![],
                                },
                            ],
                            rparen: vec![],
                            arrow: vec![],
                            body: FunctionBody::Expr(Expression::Binary(Box::new(BinaryExpr {
                                base: BaseNode {
                                    location: loc.get(10, 44, 10, 85),
                                    ..BaseNode::default()
                                },
                                operator: Operator::DivisionOperator,
                                left: Expression::Paren(Box::new(ParenExpr {
                                    base: BaseNode {
                                        location: loc.get(10, 44, 10, 71),
                                        ..BaseNode::default()
                                    },
                                    lparen: vec![],
                                    expression: Expression::Binary(Box::new(BinaryExpr {
                                        base: BaseNode {
                                            location: loc.get(10, 45, 10, 70),
                                            ..BaseNode::default()
                                        },
                                        operator: Operator::SubtractionOperator,
                                        left: Expression::Member(Box::new(MemberExpr {
                                            base: BaseNode {
                                                location: loc.get(10, 45, 10, 56),
                                                ..BaseNode::default()
                                            },
                                            object: Expression::Identifier(Identifier {
                                                base: BaseNode {
                                                    location: loc.get(10, 45, 10, 46),
                                                    ..BaseNode::default()
                                                },
                                                name: "a".to_string(),
                                            }),
                                            lbrack: vec![],
                                            property: PropertyKey::StringLit(StringLit {
                                                base: BaseNode {
                                                    location: loc.get(10, 47, 10, 55),
                                                    ..BaseNode::default()
                                                },
                                                value: "_field".to_string(),
                                            }),
                                            rbrack: vec![],
                                        })),
                                        right: Expression::Member(Box::new(MemberExpr {
                                            base: BaseNode {
                                                location: loc.get(10, 59, 10, 70),
                                                ..BaseNode::default()
                                            },
                                            object: Expression::Identifier(Identifier {
                                                base: BaseNode {
                                                    location: loc.get(10, 59, 10, 60),
                                                    ..BaseNode::default()
                                                },
                                                name: "b".to_string(),
                                            }),
                                            lbrack: vec![],
                                            property: PropertyKey::StringLit(StringLit {
                                                base: BaseNode {
                                                    location: loc.get(10, 61, 10, 69),
                                                    ..BaseNode::default()
                                                },
                                                value: "_field".to_string(),
                                            }),
                                            rbrack: vec![],
                                        })),
                                    })),
                                    rparen: vec![],
                                })),
                                right: Expression::Member(Box::new(MemberExpr {
                                    base: BaseNode {
                                        location: loc.get(10, 74, 10, 85),
                                        ..BaseNode::default()
                                    },
                                    object: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(10, 74, 10, 75),
                                            ..BaseNode::default()
                                        },
                                        name: "b".to_string(),
                                    }),
                                    lbrack: vec![],
                                    property: PropertyKey::StringLit(StringLit {
                                        base: BaseNode {
                                            location: loc.get(10, 76, 10, 84),
                                            ..BaseNode::default()
                                        },
                                        value: "_field".to_string(),
                                    }),
                                    rbrack: vec![],
                                })),
                            }))),
                        }))),
                        comma: vec![],
                    },
                ],
                rbrace: vec![],
            }))],
            rparen: vec![],
        })),
    }));

    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(2, 1, 10, 86),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![stmt_a, stmt_b, stmt_expr],
            eof: vec![],
        },
    )
}
