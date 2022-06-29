use chrono;
use pretty_assertions::assert_eq;

use super::*;
use crate::{
    ast,
    ast::{tests::Locator, Statement::Variable},
};

mod arrow_function;
mod errors;
mod from;
mod literals;
mod objects;
mod operator_precedence;
mod strings;
mod types;

/// Parsed ast roundtrips across the serde boundary and generates the same ast.
#[test]
fn parse_ast_roundtrip() {
    let mut p = Parser::new(
        r#"from(bucket: "an-bucket") |> range(start: -15m) |> filter(fn: (r) => r.field == "value")"#,
    );
    let ast = p.parse_file("".to_string());

    let serialized = serde_json::to_string(&ast).unwrap();
    let new_ast = serde_json::from_str(&serialized).unwrap();

    assert_eq!(ast, new_ast);
}

#[test]
fn parse_array_expr_no_rbrack() {
    let mut p = Parser::new(r#"group(columns: ["_time", "_field]", mode: "by")"#);
    let parsed = p.parse_file("".to_string());
    let node = ast::walk::Node::File(&parsed);
    ast::check::check(node).unwrap_err();
}

#[test]
fn parse_illegal_in_array_expr() {
    let mut p = Parser::new(
        r#"
strings.joinStr(arr: [r.unit, ": time", string(v: now()), "InfluxDB Task", $taskName, "last datapoint at:", string(v: r._time)], v: " ")
    "#,
    );
    let parsed = p.parse_file("".to_string());
    let node = ast::walk::Node::File(&parsed);
    ast::check::check(node).unwrap_err();
}

#[test]
fn parse_array_expr_no_comma() {
    let mut p = Parser::new(
        r#"
sort(columns: ["_time"], desc: true)|> limit(n: [object Object])
    "#,
    );
    let parsed = p.parse_file("".to_string());
    let node = ast::walk::Node::File(&parsed);
    ast::check::check(node).unwrap_err();
}

#[test]
fn parse_invalid_unicode_bare() {
    let mut p = Parser::new(r#"®some string®"#);
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
            body: vec![
                Statement::Bad(Box::new(BadStmt {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 3),
                        ..BaseNode::default()
                    },
                    text: "®".to_string(),
                })),
                Statement::Expr(Box::new(ExprStmt {
                    base: BaseNode {
                        location: loc.get(1, 3, 1, 7),
                        ..BaseNode::default()
                    },
                    expression: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 3, 1, 7),
                            ..BaseNode::default()
                        },
                        name: "some".to_string(),
                    })
                })),
                Statement::Expr(Box::new(ExprStmt {
                    base: BaseNode {
                        location: loc.get(1, 8, 1, 14),
                        ..BaseNode::default()
                    },
                    expression: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 8, 1, 14),
                            ..BaseNode::default()
                        },
                        name: "string".to_string(),
                    })
                })),
                Statement::Bad(Box::new(BadStmt {
                    base: BaseNode {
                        location: loc.get(1, 14, 1, 16),
                        ..BaseNode::default()
                    },
                    text: "®".to_string(),
                })),
            ],
            eof: vec![],
        },
    )
}

#[test]
fn parse_invalid_unicode_paren_wrapped() {
    let mut p = Parser::new(r#"(‛some string‛)"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
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
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 20),
                    ..BaseNode::default()
                },
                expression: Expression::Paren(Box::new(ParenExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 20),
                        errors: vec!["invalid expression @1:16-1:19: ‛".to_string()],
                        ..BaseNode::default()
                    },
                    lparen: vec![],
                    expression: Expression::Binary(Box::new(BinaryExpr {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 16),
                            ..BaseNode::default()
                        },
                        operator: Operator::InvalidOperator,
                        left: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 5, 1, 9),
                                errors: vec!["invalid expression @1:2-1:5: ‛".to_string()],
                                ..BaseNode::default()
                            },
                            name: "some".to_string(),
                        }),
                        right: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 10, 1, 16),
                                ..BaseNode::default()
                            },
                            name: "string".to_string(),
                        })
                    })),
                    rparen: vec![],
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn parse_invalid_unicode_interspersed() {
    let mut p = Parser::new(r#"®s®t®r®i®n®g"#);
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
            body: vec![
                Statement::Bad(Box::new(BadStmt {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 3),
                        ..BaseNode::default()
                    },
                    text: "®".to_string(),
                })),
                Statement::Expr(Box::new(ExprStmt {
                    base: BaseNode {
                        location: loc.get(1, 3, 1, 4),
                        ..BaseNode::default()
                    },
                    expression: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 3, 1, 4),
                            ..BaseNode::default()
                        },
                        name: "s".to_string(),
                    })
                })),
                Statement::Bad(Box::new(BadStmt {
                    base: BaseNode {
                        location: loc.get(1, 4, 1, 6),
                        ..BaseNode::default()
                    },
                    text: "®".to_string(),
                })),
                Statement::Expr(Box::new(ExprStmt {
                    base: BaseNode {
                        location: loc.get(1, 6, 1, 7),
                        ..BaseNode::default()
                    },
                    expression: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 6, 1, 7),
                            ..BaseNode::default()
                        },
                        name: "t".to_string(),
                    })
                })),
                Statement::Bad(Box::new(BadStmt {
                    base: BaseNode {
                        location: loc.get(1, 7, 1, 9),
                        ..BaseNode::default()
                    },
                    text: "®".to_string(),
                })),
                Statement::Expr(Box::new(ExprStmt {
                    base: BaseNode {
                        location: loc.get(1, 9, 1, 10),
                        ..BaseNode::default()
                    },
                    expression: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 9, 1, 10),
                            ..BaseNode::default()
                        },
                        name: "r".to_string(),
                    })
                })),
                Statement::Bad(Box::new(BadStmt {
                    base: BaseNode {
                        location: loc.get(1, 10, 1, 12),
                        ..BaseNode::default()
                    },
                    text: "®".to_string(),
                })),
                Statement::Expr(Box::new(ExprStmt {
                    base: BaseNode {
                        location: loc.get(1, 12, 1, 13),
                        ..BaseNode::default()
                    },
                    expression: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 12, 1, 13),
                            ..BaseNode::default()
                        },
                        name: "i".to_string(),
                    })
                })),
                Statement::Bad(Box::new(BadStmt {
                    base: BaseNode {
                        location: loc.get(1, 13, 1, 15),
                        ..BaseNode::default()
                    },
                    text: "®".to_string(),
                })),
                Statement::Expr(Box::new(ExprStmt {
                    base: BaseNode {
                        location: loc.get(1, 15, 1, 16),
                        ..BaseNode::default()
                    },
                    expression: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 15, 1, 16),
                            ..BaseNode::default()
                        },
                        name: "n".to_string(),
                    })
                })),
                Statement::Bad(Box::new(BadStmt {
                    base: BaseNode {
                        location: loc.get(1, 16, 1, 18),
                        ..BaseNode::default()
                    },
                    text: "®".to_string(),
                })),
                Statement::Expr(Box::new(ExprStmt {
                    base: BaseNode {
                        location: loc.get(1, 18, 1, 19),
                        ..BaseNode::default()
                    },
                    expression: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 18, 1, 19),
                            ..BaseNode::default()
                        },
                        name: "g".to_string(),
                    })
                })),
            ],
            eof: vec![],
        },
    )
}

#[test]
fn parse_greedy_quotes_paren_wrapped() {
    let mut p = Parser::new(r#"(“some string”)"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
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
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 20),
                    ..BaseNode::default()
                },
                expression: Expression::Paren(Box::new(ParenExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 20),
                        errors: vec!["invalid expression @1:16-1:19: ”".to_string()],
                        ..BaseNode::default()
                    },
                    lparen: vec![],
                    expression: Expression::Binary(Box::new(BinaryExpr {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 16),
                            ..BaseNode::default()
                        },
                        operator: Operator::InvalidOperator,
                        left: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 5, 1, 9),
                                errors: vec!["invalid expression @1:2-1:5: “".to_string()],
                                ..BaseNode::default()
                            },
                            name: "some".to_string(),
                        }),
                        right: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 10, 1, 16),
                                ..BaseNode::default()
                            },
                            name: "string".to_string(),
                        })
                    })),
                    rparen: vec![],
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn parse_greedy_quotes_bare() {
    let mut p = Parser::new(r#"“some string”"#);
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
            body: vec![
                Statement::Bad(Box::new(BadStmt {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 4),
                        ..BaseNode::default()
                    },
                    text: "“".to_string(),
                })),
                Statement::Expr(Box::new(ExprStmt {
                    base: BaseNode {
                        location: loc.get(1, 4, 1, 8),
                        ..BaseNode::default()
                    },
                    expression: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 4, 1, 8),
                            ..BaseNode::default()
                        },
                        name: "some".to_string(),
                    })
                })),
                Statement::Expr(Box::new(ExprStmt {
                    base: BaseNode {
                        location: loc.get(1, 9, 1, 15),
                        ..BaseNode::default()
                    },
                    expression: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 9, 1, 15),
                            ..BaseNode::default()
                        },
                        name: "string".to_string(),
                    })
                })),
                Statement::Bad(Box::new(BadStmt {
                    base: BaseNode {
                        location: loc.get(1, 15, 1, 18),
                        ..BaseNode::default()
                    },
                    text: "”".to_string(),
                })),
            ],
            eof: vec![],
        },
    )
}

#[test]
fn parse_greedy_quotes_interspersed() {
    let mut p = Parser::new(r#"“s”t“r”i“n”g"#);
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
            body: vec![
                Statement::Bad(Box::new(BadStmt {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 4),
                        ..BaseNode::default()
                    },
                    text: "“".to_string(),
                })),
                Statement::Expr(Box::new(ExprStmt {
                    base: BaseNode {
                        location: loc.get(1, 4, 1, 5),
                        ..BaseNode::default()
                    },
                    expression: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 4, 1, 5),
                            ..BaseNode::default()
                        },
                        name: "s".to_string(),
                    })
                })),
                Statement::Bad(Box::new(BadStmt {
                    base: BaseNode {
                        location: loc.get(1, 5, 1, 8),
                        ..BaseNode::default()
                    },
                    text: "”".to_string(),
                })),
                Statement::Expr(Box::new(ExprStmt {
                    base: BaseNode {
                        location: loc.get(1, 8, 1, 9),
                        ..BaseNode::default()
                    },
                    expression: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 8, 1, 9),
                            ..BaseNode::default()
                        },
                        name: "t".to_string(),
                    })
                })),
                Statement::Bad(Box::new(BadStmt {
                    base: BaseNode {
                        location: loc.get(1, 9, 1, 12),
                        ..BaseNode::default()
                    },
                    text: "“".to_string(),
                })),
                Statement::Expr(Box::new(ExprStmt {
                    base: BaseNode {
                        location: loc.get(1, 12, 1, 13),
                        ..BaseNode::default()
                    },
                    expression: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 12, 1, 13),
                            ..BaseNode::default()
                        },
                        name: "r".to_string(),
                    })
                })),
                Statement::Bad(Box::new(BadStmt {
                    base: BaseNode {
                        location: loc.get(1, 13, 1, 16),
                        ..BaseNode::default()
                    },
                    text: "”".to_string(),
                })),
                Statement::Expr(Box::new(ExprStmt {
                    base: BaseNode {
                        location: loc.get(1, 16, 1, 17),
                        ..BaseNode::default()
                    },
                    expression: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 16, 1, 17),
                            ..BaseNode::default()
                        },
                        name: "i".to_string(),
                    })
                })),
                Statement::Bad(Box::new(BadStmt {
                    base: BaseNode {
                        location: loc.get(1, 17, 1, 20),
                        ..BaseNode::default()
                    },
                    text: "“".to_string(),
                })),
                Statement::Expr(Box::new(ExprStmt {
                    base: BaseNode {
                        location: loc.get(1, 20, 1, 21),
                        ..BaseNode::default()
                    },
                    expression: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 20, 1, 21),
                            ..BaseNode::default()
                        },
                        name: "n".to_string(),
                    })
                })),
                Statement::Bad(Box::new(BadStmt {
                    base: BaseNode {
                        location: loc.get(1, 21, 1, 24),
                        ..BaseNode::default()
                    },
                    text: "”".to_string(),
                })),
                Statement::Expr(Box::new(ExprStmt {
                    base: BaseNode {
                        location: loc.get(1, 24, 1, 25),
                        ..BaseNode::default()
                    },
                    expression: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 24, 1, 25),
                            ..BaseNode::default()
                        },
                        name: "g".to_string(),
                    })
                })),
            ],
            eof: vec![],
        },
    )
}

#[test]
fn package_clause() {
    let mut p = Parser::new(r#"package foo"#);
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
            package: Some(PackageClause {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 12),
                    ..BaseNode::default()
                },
                name: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 9, 1, 12),
                        ..BaseNode::default()
                    },
                    name: "foo".to_string()
                }
            }),
            imports: vec![],
            body: vec![],
            eof: vec![],
        },
    )
}

#[test]
fn string_interpolation_trailing_dollar() {
    let mut p = Parser::new(r#""a + b = ${a + b}$""#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
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
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 20),
                    ..BaseNode::default()
                },
                expression: Expression::StringExpr(Box::new(StringExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 20),
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
                        StringExprPart::Text(TextPart {
                            base: BaseNode {
                                location: loc.get(1, 18, 1, 19),
                                ..BaseNode::default()
                            },
                            value: "$".to_string(),
                        }),
                    ],
                })),
            })),],
            eof: vec![],
        },
    )
}

#[test]
fn import() {
    let mut p = Parser::new(r#"import "path/foo""#);
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
            imports: vec![ImportDeclaration {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 18),
                    ..BaseNode::default()
                },
                alias: None,
                path: StringLit {
                    base: BaseNode {
                        location: loc.get(1, 8, 1, 18),
                        ..BaseNode::default()
                    },
                    value: "path/foo".to_string()
                }
            }],
            body: vec![],
            eof: vec![],
        },
    )
}

#[test]
fn import_as() {
    let mut p = Parser::new(r#"import bar "path/foo""#);
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
            imports: vec![ImportDeclaration {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 22),
                    ..BaseNode::default()
                },
                alias: Some(Identifier {
                    base: BaseNode {
                        location: loc.get(1, 8, 1, 11),
                        ..BaseNode::default()
                    },
                    name: "bar".to_string()
                }),
                path: StringLit {
                    base: BaseNode {
                        location: loc.get(1, 12, 1, 22),
                        ..BaseNode::default()
                    },
                    value: "path/foo".to_string()
                }
            }],
            body: vec![],
            eof: vec![],
        }
    )
}

#[test]
fn imports() {
    let mut p = Parser::new(
        r#"import "path/foo"
import "path/bar""#,
    );
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 2, 18),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![
                ImportDeclaration {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 18),
                        ..BaseNode::default()
                    },
                    alias: None,
                    path: StringLit {
                        base: BaseNode {
                            location: loc.get(1, 8, 1, 18),
                            ..BaseNode::default()
                        },
                        value: "path/foo".to_string()
                    }
                },
                ImportDeclaration {
                    base: BaseNode {
                        location: loc.get(2, 1, 2, 18),
                        ..BaseNode::default()
                    },
                    alias: None,
                    path: StringLit {
                        base: BaseNode {
                            location: loc.get(2, 8, 2, 18),
                            ..BaseNode::default()
                        },
                        value: "path/bar".to_string()
                    }
                }
            ],
            body: vec![],
            eof: vec![],
        }
    )
}

#[test]
fn package_and_imports() {
    let mut p = Parser::new(
        r#"
package baz

import "path/foo"
import "path/bar""#,
    );
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(2, 1, 5, 18),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: Some(PackageClause {
                base: BaseNode {
                    location: loc.get(2, 1, 2, 12),
                    ..BaseNode::default()
                },
                name: Identifier {
                    base: BaseNode {
                        location: loc.get(2, 9, 2, 12),
                        ..BaseNode::default()
                    },
                    name: "baz".to_string()
                }
            }),
            imports: vec![
                ImportDeclaration {
                    base: BaseNode {
                        location: loc.get(4, 1, 4, 18),
                        ..BaseNode::default()
                    },
                    alias: None,
                    path: StringLit {
                        base: BaseNode {
                            location: loc.get(4, 8, 4, 18),
                            ..BaseNode::default()
                        },
                        value: "path/foo".to_string()
                    }
                },
                ImportDeclaration {
                    base: BaseNode {
                        location: loc.get(5, 1, 5, 18),
                        ..BaseNode::default()
                    },
                    alias: None,
                    path: StringLit {
                        base: BaseNode {
                            location: loc.get(5, 8, 5, 18),
                            ..BaseNode::default()
                        },
                        value: "path/bar".to_string()
                    }
                }
            ],
            body: vec![],
            eof: vec![],
        },
    )
}

#[test]
fn package_and_imports_and_body() {
    let mut p = Parser::new(
        r#"
package baz

import "path/foo"
import "path/bar"

1 + 1"#,
    );
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(2, 1, 7, 6),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: Some(PackageClause {
                base: BaseNode {
                    location: loc.get(2, 1, 2, 12),
                    ..BaseNode::default()
                },
                name: Identifier {
                    base: BaseNode {
                        location: loc.get(2, 9, 2, 12),
                        ..BaseNode::default()
                    },
                    name: "baz".to_string()
                }
            }),
            imports: vec![
                ImportDeclaration {
                    base: BaseNode {
                        location: loc.get(4, 1, 4, 18),
                        ..BaseNode::default()
                    },
                    alias: None,
                    path: StringLit {
                        base: BaseNode {
                            location: loc.get(4, 8, 4, 18),
                            ..BaseNode::default()
                        },
                        value: "path/foo".to_string()
                    }
                },
                ImportDeclaration {
                    base: BaseNode {
                        location: loc.get(5, 1, 5, 18),
                        ..BaseNode::default()
                    },
                    alias: None,
                    path: StringLit {
                        base: BaseNode {
                            location: loc.get(5, 8, 5, 18),
                            ..BaseNode::default()
                        },
                        value: "path/bar".to_string()
                    }
                }
            ],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(7, 1, 7, 6),
                    ..BaseNode::default()
                },
                expression: Expression::Binary(Box::new(BinaryExpr {
                    base: BaseNode {
                        location: loc.get(7, 1, 7, 6),
                        ..BaseNode::default()
                    },
                    operator: Operator::AdditionOperator,
                    left: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(7, 1, 7, 2),
                            ..BaseNode::default()
                        },
                        value: 1
                    }),
                    right: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(7, 5, 7, 6),
                            ..BaseNode::default()
                        },
                        value: 1
                    })
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn optional_query_metadata() {
    let mut p = Parser::new(
        r#"option task = {
				name: "foo",
				every: 1h,
				delay: 10m,
				cron: "0 2 * * *",
				retry: 5,
			  }"#,
    );
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 7, 7),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Option(Box::new(OptionStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 7, 7),
                    ..BaseNode::default()
                },
                assignment: Assignment::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(1, 8, 7, 7),
                        ..BaseNode::default()
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(1, 8, 1, 12),
                            ..BaseNode::default()
                        },
                        name: "task".to_string()
                    },
                    init: Expression::Object(Box::new(ObjectExpr {
                        base: BaseNode {
                            location: loc.get(1, 15, 7, 7),
                            ..BaseNode::default()
                        },
                        lbrace: vec![],
                        with: None,
                        properties: vec![
                            Property {
                                base: BaseNode {
                                    location: loc.get(2, 5, 2, 16),
                                    ..BaseNode::default()
                                },
                                key: PropertyKey::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(2, 5, 2, 9),
                                        ..BaseNode::default()
                                    },
                                    name: "name".to_string()
                                }),
                                separator: vec![],
                                value: Some(Expression::StringLit(StringLit {
                                    base: BaseNode {
                                        location: loc.get(2, 11, 2, 16),
                                        ..BaseNode::default()
                                    },
                                    value: "foo".to_string()
                                })),
                                comma: vec![],
                            },
                            Property {
                                base: BaseNode {
                                    location: loc.get(3, 5, 3, 14),
                                    ..BaseNode::default()
                                },
                                key: PropertyKey::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(3, 5, 3, 10),
                                        ..BaseNode::default()
                                    },
                                    name: "every".to_string()
                                }),
                                separator: vec![],
                                value: Some(Expression::Duration(DurationLit {
                                    base: BaseNode {
                                        location: loc.get(3, 12, 3, 14),
                                        ..BaseNode::default()
                                    },
                                    values: vec![Duration {
                                        magnitude: 1,
                                        unit: "h".to_string()
                                    }]
                                })),
                                comma: vec![],
                            },
                            Property {
                                base: BaseNode {
                                    location: loc.get(4, 5, 4, 15),
                                    ..BaseNode::default()
                                },
                                key: PropertyKey::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(4, 5, 4, 10),
                                        ..BaseNode::default()
                                    },
                                    name: "delay".to_string()
                                }),
                                separator: vec![],
                                value: Some(Expression::Duration(DurationLit {
                                    base: BaseNode {
                                        location: loc.get(4, 12, 4, 15),
                                        ..BaseNode::default()
                                    },
                                    values: vec![Duration {
                                        magnitude: 10,
                                        unit: "m".to_string()
                                    }]
                                })),
                                comma: vec![],
                            },
                            Property {
                                base: BaseNode {
                                    location: loc.get(5, 5, 5, 22),
                                    ..BaseNode::default()
                                },
                                key: PropertyKey::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(5, 5, 5, 9),
                                        ..BaseNode::default()
                                    },
                                    name: "cron".to_string()
                                }),
                                separator: vec![],
                                value: Some(Expression::StringLit(StringLit {
                                    base: BaseNode {
                                        location: loc.get(5, 11, 5, 22),
                                        ..BaseNode::default()
                                    },
                                    value: "0 2 * * *".to_string()
                                })),
                                comma: vec![],
                            },
                            Property {
                                base: BaseNode {
                                    location: loc.get(6, 5, 6, 13),
                                    ..BaseNode::default()
                                },
                                key: PropertyKey::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(6, 5, 6, 10),
                                        ..BaseNode::default()
                                    },
                                    name: "retry".to_string()
                                }),
                                separator: vec![],
                                value: Some(Expression::Integer(IntegerLit {
                                    base: BaseNode {
                                        location: loc.get(6, 12, 6, 13),
                                        ..BaseNode::default()
                                    },
                                    value: 5
                                })),
                                comma: vec![],
                            }
                        ],
                        rbrace: vec![],
                    }))
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn optional_query_metadata_preceding_query_text() {
    let mut p = Parser::new(
        r#"option task = {
     name: "foo",  // Name of task
     every: 1h,    // Execution frequency of task
    }

    // Task will execute the following query
    from() |> count()"#,
    );
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 7, 22),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Statement::Option(Box::new(OptionStmt {
                    base: BaseNode {
                        location: loc.get(1, 1, 4, 6),
                        ..BaseNode::default()
                    },
                    assignment: Assignment::Variable(Box::new(VariableAssgn {
                        base: BaseNode {
                            location: loc.get(1, 8, 4, 6),
                            ..BaseNode::default()
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: loc.get(1, 8, 1, 12),
                                ..BaseNode::default()
                            },
                            name: "task".to_string()
                        },
                        init: Expression::Object(Box::new(ObjectExpr {
                            base: BaseNode {
                                location: loc.get(1, 15, 4, 6),
                                ..BaseNode::default()
                            },
                            lbrace: vec![],
                            with: None,
                            properties: vec![
                                Property {
                                    base: BaseNode {
                                        location: loc.get(2, 6, 2, 17),
                                        ..BaseNode::default()
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(2, 6, 2, 10),
                                            ..BaseNode::default()
                                        },
                                        name: "name".to_string()
                                    }),
                                    separator: vec![],
                                    value: Some(Expression::StringLit(StringLit {
                                        base: BaseNode {
                                            location: loc.get(2, 12, 2, 17),
                                            ..BaseNode::default()
                                        },
                                        value: "foo".to_string()
                                    })),
                                    comma: vec![],
                                },
                                Property {
                                    base: BaseNode {
                                        location: loc.get(3, 6, 3, 15),
                                        ..BaseNode::default()
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(3, 6, 3, 11),
                                            comments: vec![ast::Comment {
                                                text: String::from("// Name of task\n"),
                                            }],
                                            ..BaseNode::default()
                                        },
                                        name: "every".to_string()
                                    }),
                                    separator: vec![],
                                    value: Some(Expression::Duration(DurationLit {
                                        base: BaseNode {
                                            location: loc.get(3, 13, 3, 15),
                                            ..BaseNode::default()
                                        },
                                        values: vec![Duration {
                                            magnitude: 1,
                                            unit: "h".to_string()
                                        }]
                                    })),
                                    comma: vec![],
                                }
                            ],
                            rbrace: vec![ast::Comment {
                                text: String::from("// Execution frequency of task\n"),
                            }],
                        }))
                    }))
                })),
                Statement::Expr(Box::new(ExprStmt {
                    base: BaseNode {
                        location: loc.get(7, 5, 7, 22),
                        ..BaseNode::default()
                    },
                    expression: Expression::PipeExpr(Box::new(PipeExpr {
                        base: BaseNode {
                            location: loc.get(7, 5, 7, 22),
                            ..BaseNode::default()
                        },
                        argument: Expression::Call(Box::new(CallExpr {
                            base: BaseNode {
                                location: loc.get(7, 5, 7, 11),
                                ..BaseNode::default()
                            },
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(7, 5, 7, 9),
                                    comments: vec![ast::Comment {
                                        text: String::from(
                                            "// Task will execute the following query\n"
                                        ),
                                    }],
                                    ..BaseNode::default()
                                },
                                name: "from".to_string()
                            }),
                            lparen: vec![],
                            arguments: vec![],
                            rparen: vec![],
                        })),
                        call: CallExpr {
                            base: BaseNode {
                                location: loc.get(7, 15, 7, 22),
                                ..BaseNode::default()
                            },
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(7, 15, 7, 20),
                                    ..BaseNode::default()
                                },
                                name: "count".to_string()
                            }),
                            lparen: vec![],
                            arguments: vec![],
                            rparen: vec![],
                        }
                    })),
                }))
            ],
            eof: vec![],
        },
    )
}

#[test]
fn qualified_option() {
    let mut p = Parser::new(r#"option alert.state = "Warning""#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 31),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Option(Box::new(OptionStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 31),
                    ..BaseNode::default()
                },
                assignment: Assignment::Member(Box::new(MemberAssgn {
                    base: BaseNode {
                        location: loc.get(1, 8, 1, 31),
                        ..BaseNode::default()
                    },
                    member: MemberExpr {
                        base: BaseNode {
                            location: loc.get(1, 8, 1, 19),
                            ..BaseNode::default()
                        },
                        object: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 8, 1, 13),
                                ..BaseNode::default()
                            },
                            name: "alert".to_string()
                        }),
                        lbrack: vec![],
                        property: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 14, 1, 19),
                                ..BaseNode::default()
                            },
                            name: "state".to_string()
                        }),
                        rbrack: vec![],
                    },
                    init: Expression::StringLit(StringLit {
                        base: BaseNode {
                            location: loc.get(1, 22, 1, 31),
                            ..BaseNode::default()
                        },
                        value: "Warning".to_string()
                    })
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn builtin() {
    let mut p = Parser::new(r#"builtin from : int"#);
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
            body: vec![Statement::Builtin(Box::new(BuiltinStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 19),
                    ..BaseNode::default()
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 9, 1, 13),
                        ..BaseNode::default()
                    },
                    name: "from".to_string()
                },
                colon: vec![],
                ty: TypeExpression {
                    base: BaseNode {
                        location: loc.get(1, 16, 1, 19),
                        ..BaseNode::default()
                    },
                    monotype: MonoType::Basic(NamedType {
                        base: BaseNode {
                            location: loc.get(1, 16, 1, 19),
                            ..BaseNode::default()
                        },
                        name: Identifier {
                            base: BaseNode {
                                location: loc.get(1, 16, 1, 19),
                                ..BaseNode::default()
                            },
                            name: "int".to_string()
                        },
                    }),
                    constraints: vec![]
                },
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn comment() {
    let mut p = Parser::new(
        r#"// Comment
			from()"#,
    );
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(2, 4, 2, 10),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(2, 4, 2, 10),
                    ..BaseNode::default()
                },
                expression: Expression::Call(Box::new(CallExpr {
                    base: BaseNode {
                        location: loc.get(2, 4, 2, 10),
                        ..BaseNode::default()
                    },
                    callee: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(2, 4, 2, 8),
                            comments: vec![ast::Comment {
                                text: String::from("// Comment\n"),
                            }],
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
fn comment_builtin() {
    let mut p = Parser::new(
        r#"// Comment
builtin foo
// colon comment
: int"#,
    );
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(2, 1, 4, 6),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Builtin(Box::new(BuiltinStmt {
                base: BaseNode {
                    location: loc.get(2, 1, 4, 6),
                    comments: vec![ast::Comment {
                        text: String::from("// Comment\n"),
                    }],
                    ..BaseNode::default()
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(2, 9, 2, 12),
                        ..BaseNode::default()
                    },
                    name: "foo".to_string()
                },
                colon: vec![ast::Comment {
                    text: String::from("// colon comment\n"),
                }],
                ty: TypeExpression {
                    base: BaseNode {
                        location: loc.get(4, 3, 4, 6),
                        ..BaseNode::default()
                    },
                    monotype: MonoType::Basic(NamedType {
                        base: BaseNode {
                            location: loc.get(4, 3, 4, 6),
                            ..BaseNode::default()
                        },
                        name: Identifier {
                            base: BaseNode {
                                location: loc.get(4, 3, 4, 6),
                                ..BaseNode::default()
                            },
                            name: "int".to_string()
                        },
                    }),
                    constraints: vec![]
                },
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn comment_function_body() {
    let mut p = Parser::new(
        r#"fn = (tables=<-) =>
// comment
(tables)"#,
    );
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 3, 9),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 3, 9),
                    ..BaseNode::default()
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 3),
                        ..BaseNode::default()
                    },
                    name: "fn".to_string()
                },
                init: Expression::Function(Box::new(FunctionExpr {
                    base: BaseNode {
                        location: loc.get(1, 6, 3, 9),
                        ..BaseNode::default()
                    },
                    lparen: vec![],
                    params: vec![Property {
                        base: BaseNode {
                            location: loc.get(1, 7, 1, 16),
                            ..BaseNode::default()
                        },
                        key: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 7, 1, 13),
                                ..BaseNode::default()
                            },
                            name: "tables".to_string()
                        }),
                        separator: vec![],
                        value: Some(Expression::PipeLit(PipeLit {
                            base: BaseNode {
                                location: loc.get(1, 14, 1, 16),
                                ..BaseNode::default()
                            },
                        })),
                        comma: vec![],
                    }],
                    rparen: vec![],
                    arrow: vec![],
                    body: FunctionBody::Expr(Expression::Paren(Box::new(ParenExpr {
                        base: BaseNode {
                            location: loc.get(3, 1, 3, 9),
                            ..BaseNode::default()
                        },
                        lparen: vec![ast::Comment {
                            text: String::from("// comment\n"),
                        }],
                        expression: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(3, 2, 3, 8),
                                ..BaseNode::default()
                            },
                            name: "tables".to_string(),
                        }),
                        rparen: vec![],
                    })))
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn identifier_with_number() {
    let mut p = Parser::new(r#"tan2()"#);
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
                        name: "tan2".to_string()
                    }),
                    lparen: vec![],
                    arguments: vec![],
                    rparen: vec![],
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn regex_match_operators() {
    let mut p = Parser::new(r#""a" =~ /.*/ and "b" !~ /c$/"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 28),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 28),
                    ..BaseNode::default()
                },
                expression: Expression::Logical(Box::new(LogicalExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 28),
                        ..BaseNode::default()
                    },
                    operator: LogicalOperator::AndOperator,
                    left: Expression::Binary(Box::new(BinaryExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 12),
                            ..BaseNode::default()
                        },
                        operator: Operator::RegexpMatchOperator,
                        left: Expression::StringLit(StringLit {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 4),
                                ..BaseNode::default()
                            },
                            value: "a".to_string()
                        }),
                        right: Expression::Regexp(RegexpLit {
                            base: BaseNode {
                                location: loc.get(1, 8, 1, 12),
                                ..BaseNode::default()
                            },
                            value: ".*".to_string()
                        })
                    })),
                    right: Expression::Binary(Box::new(BinaryExpr {
                        base: BaseNode {
                            location: loc.get(1, 17, 1, 28),
                            ..BaseNode::default()
                        },
                        operator: Operator::NotRegexpMatchOperator,
                        left: Expression::StringLit(StringLit {
                            base: BaseNode {
                                location: loc.get(1, 17, 1, 20),
                                ..BaseNode::default()
                            },
                            value: "b".to_string()
                        }),
                        right: Expression::Regexp(RegexpLit {
                            base: BaseNode {
                                location: loc.get(1, 24, 1, 28),
                                ..BaseNode::default()
                            },
                            value: "c$".to_string()
                        })
                    }))
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn declare_variable_as_an_int() {
    let mut p = Parser::new(r#"howdy = 1"#);
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
                        location: loc.get(1, 1, 1, 6),
                        ..BaseNode::default()
                    },
                    name: "howdy".to_string()
                },
                init: Expression::Integer(IntegerLit {
                    base: BaseNode {
                        location: loc.get(1, 9, 1, 10),
                        ..BaseNode::default()
                    },
                    value: 1
                })
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn declare_variable_as_a_float() {
    let mut p = Parser::new(r#"howdy = 1.1"#);
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
                        location: loc.get(1, 1, 1, 6),
                        ..BaseNode::default()
                    },
                    name: "howdy".to_string()
                },
                init: Expression::Float(FloatLit {
                    base: BaseNode {
                        location: loc.get(1, 9, 1, 12),
                        ..BaseNode::default()
                    },
                    value: 1.1
                })
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn declare_variable_as_an_array() {
    let mut p = Parser::new(r#"howdy = [1, 2, 3, 4]"#);
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
                        location: loc.get(1, 1, 1, 6),
                        ..BaseNode::default()
                    },
                    name: "howdy".to_string()
                },
                init: Expression::Array(Box::new(ArrayExpr {
                    base: BaseNode {
                        location: loc.get(1, 9, 1, 21),
                        ..BaseNode::default()
                    },
                    lbrack: vec![],
                    elements: vec![
                        ArrayItem {
                            expression: Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 10, 1, 11),
                                    ..BaseNode::default()
                                },
                                value: 1
                            }),
                            comma: vec![],
                        },
                        ArrayItem {
                            expression: Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 13, 1, 14),
                                    ..BaseNode::default()
                                },
                                value: 2
                            }),
                            comma: vec![],
                        },
                        ArrayItem {
                            expression: Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 16, 1, 17),
                                    ..BaseNode::default()
                                },
                                value: 3
                            }),
                            comma: vec![],
                        },
                        ArrayItem {
                            expression: Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 19, 1, 20),
                                    ..BaseNode::default()
                                },
                                value: 4
                            }),
                            comma: vec![],
                        }
                    ],
                    rbrack: vec![],
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn declare_variable_as_an_empty_array() {
    let mut p = Parser::new(r#"howdy = []"#);
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
                        location: loc.get(1, 1, 1, 6),
                        ..BaseNode::default()
                    },
                    name: "howdy".to_string()
                },
                init: Expression::Array(Box::new(ArrayExpr {
                    base: BaseNode {
                        location: loc.get(1, 9, 1, 11),
                        ..BaseNode::default()
                    },
                    lbrack: vec![],
                    elements: vec![],
                    rbrack: vec![],
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn parse_empty_dict() {
    let mut p = Parser::new(r#"[:]"#);
    let parsed = p.parse_expression();
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        Expression::Dict(Box::new(DictExpr {
            base: BaseNode {
                location: loc.get(1, 1, 1, 4),
                ..BaseNode::default()
            },
            elements: vec![],
            lbrack: vec![],
            rbrack: vec![],
        }))
    )
}

#[test]
fn parse_single_element_dict() {
    let mut p = Parser::new(r#"["a": 0]"#);
    let parsed = p.parse_expression();
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        Expression::Dict(Box::new(DictExpr {
            base: BaseNode {
                location: loc.get(1, 1, 1, 9),
                ..BaseNode::default()
            },
            elements: vec![DictItem {
                key: Expression::StringLit(StringLit {
                    base: BaseNode {
                        location: loc.get(1, 2, 1, 5),
                        ..BaseNode::default()
                    },
                    value: "a".to_string(),
                }),
                val: Expression::Integer(IntegerLit {
                    base: BaseNode {
                        location: loc.get(1, 7, 1, 8),
                        ..BaseNode::default()
                    },
                    value: 0,
                }),
                comma: vec![],
            }],
            lbrack: vec![],
            rbrack: vec![],
        }))
    )
}

#[test]
fn parse_multi_element_dict() {
    let mut p = Parser::new(r#"["a": 0, "b": 1]"#);
    let parsed = p.parse_expression();
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        Expression::Dict(Box::new(DictExpr {
            base: BaseNode {
                location: loc.get(1, 1, 1, 17),
                ..BaseNode::default()
            },
            elements: vec![
                DictItem {
                    key: Expression::StringLit(StringLit {
                        base: BaseNode {
                            location: loc.get(1, 2, 1, 5),
                            ..BaseNode::default()
                        },
                        value: "a".to_string(),
                    }),
                    val: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 7, 1, 8),
                            ..BaseNode::default()
                        },
                        value: 0,
                    }),
                    comma: vec![],
                },
                DictItem {
                    key: Expression::StringLit(StringLit {
                        base: BaseNode {
                            location: loc.get(1, 10, 1, 13),
                            ..BaseNode::default()
                        },
                        value: "b".to_string(),
                    }),
                    val: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 15, 1, 16),
                            ..BaseNode::default()
                        },
                        value: 1,
                    }),
                    comma: vec![],
                },
            ],
            lbrack: vec![],
            rbrack: vec![],
        }))
    )
}

#[test]
fn parse_dict_trailing_comma0() {
    let mut p = Parser::new(r#"["a": 0, ]"#);
    let parsed = p.parse_expression();
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        Expression::Dict(Box::new(DictExpr {
            base: BaseNode {
                location: loc.get(1, 1, 1, 11),
                ..BaseNode::default()
            },
            elements: vec![DictItem {
                key: Expression::StringLit(StringLit {
                    base: BaseNode {
                        location: loc.get(1, 2, 1, 5),
                        ..BaseNode::default()
                    },
                    value: "a".to_string(),
                }),
                val: Expression::Integer(IntegerLit {
                    base: BaseNode {
                        location: loc.get(1, 7, 1, 8),
                        ..BaseNode::default()
                    },
                    value: 0,
                }),
                comma: vec![],
            }],
            lbrack: vec![],
            rbrack: vec![],
        }))
    )
}

#[test]
fn parse_dict_trailing_comma1() {
    let mut p = Parser::new(r#"["a": 0, "b": 1, ]"#);
    let parsed = p.parse_expression();
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        Expression::Dict(Box::new(DictExpr {
            base: BaseNode {
                location: loc.get(1, 1, 1, 19),
                ..BaseNode::default()
            },
            elements: vec![
                DictItem {
                    key: Expression::StringLit(StringLit {
                        base: BaseNode {
                            location: loc.get(1, 2, 1, 5),
                            ..BaseNode::default()
                        },
                        value: "a".to_string(),
                    }),
                    val: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 7, 1, 8),
                            ..BaseNode::default()
                        },
                        value: 0,
                    }),
                    comma: vec![],
                },
                DictItem {
                    key: Expression::StringLit(StringLit {
                        base: BaseNode {
                            location: loc.get(1, 10, 1, 13),
                            ..BaseNode::default()
                        },
                        value: "b".to_string(),
                    }),
                    val: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 15, 1, 16),
                            ..BaseNode::default()
                        },
                        value: 1,
                    }),
                    comma: vec![],
                },
            ],
            lbrack: vec![],
            rbrack: vec![],
        }))
    )
}

#[test]
fn parse_dict_arbitrary_keys() {
    let mut p = Parser::new(r#"[1-1: 0, 1+1: 1]"#);
    let parsed = p.parse_expression();
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        Expression::Dict(Box::new(DictExpr {
            base: BaseNode {
                location: loc.get(1, 1, 1, 17),
                ..BaseNode::default()
            },
            elements: vec![
                DictItem {
                    key: Expression::Binary(Box::new(BinaryExpr {
                        base: BaseNode {
                            location: loc.get(1, 2, 1, 5),
                            ..BaseNode::default()
                        },
                        operator: Operator::SubtractionOperator,
                        left: Expression::Integer(IntegerLit {
                            base: BaseNode {
                                location: loc.get(1, 2, 1, 3),
                                ..BaseNode::default()
                            },
                            value: 1,
                        }),
                        right: Expression::Integer(IntegerLit {
                            base: BaseNode {
                                location: loc.get(1, 4, 1, 5),
                                ..BaseNode::default()
                            },
                            value: 1,
                        }),
                    })),
                    val: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 7, 1, 8),
                            ..BaseNode::default()
                        },
                        value: 0,
                    }),
                    comma: vec![],
                },
                DictItem {
                    key: Expression::Binary(Box::new(BinaryExpr {
                        base: BaseNode {
                            location: loc.get(1, 10, 1, 13),
                            ..BaseNode::default()
                        },
                        operator: Operator::AdditionOperator,
                        left: Expression::Integer(IntegerLit {
                            base: BaseNode {
                                location: loc.get(1, 10, 1, 11),
                                ..BaseNode::default()
                            },
                            value: 1,
                        }),
                        right: Expression::Integer(IntegerLit {
                            base: BaseNode {
                                location: loc.get(1, 12, 1, 13),
                                ..BaseNode::default()
                            },
                            value: 1,
                        }),
                    })),
                    val: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 15, 1, 16),
                            ..BaseNode::default()
                        },
                        value: 1,
                    }),
                    comma: vec![],
                },
            ],
            lbrack: vec![],
            rbrack: vec![],
        }))
    )
}

#[test]
fn use_variable_to_declare_something() {
    let mut p = Parser::new(
        r#"howdy = 1
			from()"#,
    );
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 2, 10),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 10),
                        ..BaseNode::default()
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 6),
                            ..BaseNode::default()
                        },
                        name: "howdy".to_string()
                    },
                    init: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 9, 1, 10),
                            ..BaseNode::default()
                        },
                        value: 1
                    })
                })),
                Statement::Expr(Box::new(ExprStmt {
                    base: BaseNode {
                        location: loc.get(2, 4, 2, 10),
                        ..BaseNode::default()
                    },
                    expression: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(2, 4, 2, 10),
                            ..BaseNode::default()
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(2, 4, 2, 8),
                                ..BaseNode::default()
                            },
                            name: "from".to_string()
                        }),
                        lparen: vec![],
                        arguments: vec![],
                        rparen: vec![],
                    })),
                }))
            ],
            eof: vec![],
        },
    )
}

#[test]
fn variable_is_from_statement() {
    let mut p = Parser::new(
        r#"howdy = from()
			howdy.count()"#,
    );
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 2, 17),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 15),
                        ..BaseNode::default()
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 6),
                            ..BaseNode::default()
                        },
                        name: "howdy".to_string()
                    },
                    init: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 9, 1, 15),
                            ..BaseNode::default()
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 9, 1, 13),
                                ..BaseNode::default()
                            },
                            name: "from".to_string()
                        }),
                        lparen: vec![],
                        arguments: vec![],
                        rparen: vec![],
                    })),
                })),
                Statement::Expr(Box::new(ExprStmt {
                    base: BaseNode {
                        location: loc.get(2, 4, 2, 17),
                        ..BaseNode::default()
                    },
                    expression: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(2, 4, 2, 17),
                            ..BaseNode::default()
                        },
                        callee: Expression::Member(Box::new(MemberExpr {
                            base: BaseNode {
                                location: loc.get(2, 4, 2, 15),
                                ..BaseNode::default()
                            },
                            object: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(2, 4, 2, 9),
                                    ..BaseNode::default()
                                },
                                name: "howdy".to_string()
                            }),
                            lbrack: vec![],
                            property: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(2, 10, 2, 15),
                                    ..BaseNode::default()
                                },
                                name: "count".to_string()
                            }),
                            rbrack: vec![],
                        })),
                        lparen: vec![],
                        arguments: vec![],
                        rparen: vec![],
                    }))
                }))
            ],
            eof: vec![],
        },
    )
}

#[test]
fn pipe_expression() {
    let mut p = Parser::new(r#"from() |> count()"#);
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
                expression: Expression::PipeExpr(Box::new(PipeExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 18),
                        ..BaseNode::default()
                    },
                    argument: Expression::Call(Box::new(CallExpr {
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
                    call: CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 11, 1, 18),
                            ..BaseNode::default()
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 11, 1, 16),
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
        },
    )
}

#[test]
fn pipe_expression_to_member_expression_function() {
    let mut p = Parser::new(r#"a |> b.c(d:e)"#);
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
                expression: Expression::PipeExpr(Box::new(PipeExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 14),
                        ..BaseNode::default()
                    },
                    argument: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            ..BaseNode::default()
                        },
                        name: "a".to_string()
                    }),
                    call: CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 6, 1, 14),
                            ..BaseNode::default()
                        },
                        callee: Expression::Member(Box::new(MemberExpr {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 9),
                                ..BaseNode::default()
                            },
                            object: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 7),
                                    ..BaseNode::default()
                                },
                                name: "b".to_string()
                            }),
                            lbrack: vec![],
                            property: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 8, 1, 9),
                                    ..BaseNode::default()
                                },
                                name: "c".to_string()
                            }),
                            rbrack: vec![],
                        })),
                        lparen: vec![],
                        arguments: vec![Expression::Object(Box::new(ObjectExpr {
                            base: BaseNode {
                                location: loc.get(1, 10, 1, 13),
                                ..BaseNode::default()
                            },
                            lbrace: vec![],
                            with: None,
                            properties: vec![Property {
                                base: BaseNode {
                                    location: loc.get(1, 10, 1, 13),
                                    ..BaseNode::default()
                                },
                                key: PropertyKey::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 10, 1, 11),
                                        ..BaseNode::default()
                                    },
                                    name: "d".to_string()
                                }),
                                separator: vec![],
                                value: Some(Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 12, 1, 13),
                                        ..BaseNode::default()
                                    },
                                    name: "e".to_string()
                                })),
                                comma: vec![],
                            }],
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
fn literal_pipe_expression() {
    let mut p = Parser::new(r#"5 |> pow2()"#);
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
                expression: Expression::PipeExpr(Box::new(PipeExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 12),
                        ..BaseNode::default()
                    },
                    argument: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            ..BaseNode::default()
                        },
                        value: 5,
                    }),
                    call: CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 6, 1, 12),
                            ..BaseNode::default()
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 10),
                                ..BaseNode::default()
                            },
                            name: "pow2".to_string()
                        }),
                        lparen: vec![],
                        arguments: vec![],
                        rparen: vec![],
                    }
                })),
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn member_expression_pipe_expression() {
    let mut p = Parser::new(r#"foo.bar |> baz()"#);
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
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 17),
                    ..BaseNode::default()
                },
                expression: Expression::PipeExpr(Box::new(PipeExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 17),
                        ..BaseNode::default()
                    },
                    argument: Expression::Member(Box::new(MemberExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 8),
                            ..BaseNode::default()
                        },
                        object: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 4),
                                ..BaseNode::default()
                            },
                            name: "foo".to_string()
                        }),
                        lbrack: vec![],
                        property: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 5, 1, 8),
                                ..BaseNode::default()
                            },
                            name: "bar".to_string()
                        }),
                        rbrack: vec![],
                    })),
                    call: CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 12, 1, 17),
                            ..BaseNode::default()
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 12, 1, 15),
                                ..BaseNode::default()
                            },
                            name: "baz".to_string()
                        }),
                        lparen: vec![],
                        arguments: vec![],
                        rparen: vec![],
                    }
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn multiple_pipe_expressions() {
    let mut p = Parser::new(r#"from() |> range() |> filter() |> count()"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 41),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 41),
                    ..BaseNode::default()
                },
                expression: Expression::PipeExpr(Box::new(PipeExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 41),
                        ..BaseNode::default()
                    },
                    argument: Expression::PipeExpr(Box::new(PipeExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 30),
                            ..BaseNode::default()
                        },
                        argument: Expression::PipeExpr(Box::new(PipeExpr {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 18),
                                ..BaseNode::default()
                            },
                            argument: Expression::Call(Box::new(CallExpr {
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
                            call: CallExpr {
                                base: BaseNode {
                                    location: loc.get(1, 11, 1, 18),
                                    ..BaseNode::default()
                                },
                                callee: Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 11, 1, 16),
                                        ..BaseNode::default()
                                    },
                                    name: "range".to_string()
                                }),
                                lparen: vec![],
                                arguments: vec![],
                                rparen: vec![],
                            }
                        })),
                        call: CallExpr {
                            base: BaseNode {
                                location: loc.get(1, 22, 1, 30),
                                ..BaseNode::default()
                            },
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 22, 1, 28),
                                    ..BaseNode::default()
                                },
                                name: "filter".to_string()
                            }),
                            lparen: vec![],
                            arguments: vec![],
                            rparen: vec![],
                        }
                    })),
                    call: CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 34, 1, 41),
                            ..BaseNode::default()
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 34, 1, 39),
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
        },
    )
}

#[test]
fn pipe_expression_into_non_call_expression() {
    let mut p = Parser::new(r#"foo() |> bar"#);
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
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 13),
                    ..BaseNode::default()
                },
                expression: Expression::PipeExpr(Box::new(PipeExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 13),
                        ..BaseNode::default()
                    },
                    argument: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 6),
                            ..BaseNode::default()
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 4),
                                ..BaseNode::default()
                            },
                            name: "foo".to_string()
                        }),
                        lparen: vec![],
                        arguments: vec![],
                        rparen: vec![],
                    })),
                    call: CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 10, 1, 13),
                            errors: vec!["pipe destination must be a function call".to_string()],
                            ..BaseNode::default()
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 10, 1, 13),
                                ..BaseNode::default()
                            },
                            name: "bar".to_string()
                        }),
                        lparen: vec![],
                        arguments: vec![],
                        rparen: vec![],
                    }
                })),
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn two_variables_for_two_froms() {
    let mut p = Parser::new(
        r#"howdy = from()
			doody = from()
			howdy|>count()
			doody|>sum()"#,
    );
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 4, 16),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 15),
                        ..BaseNode::default()
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 6),
                            ..BaseNode::default()
                        },
                        name: "howdy".to_string()
                    },
                    init: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 9, 1, 15),
                            ..BaseNode::default()
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 9, 1, 13),
                                ..BaseNode::default()
                            },
                            name: "from".to_string()
                        }),
                        lparen: vec![],
                        arguments: vec![],
                        rparen: vec![],
                    })),
                })),
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(2, 4, 2, 18),
                        ..BaseNode::default()
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(2, 4, 2, 9),
                            ..BaseNode::default()
                        },
                        name: "doody".to_string()
                    },
                    init: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(2, 12, 2, 18),
                            ..BaseNode::default()
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(2, 12, 2, 16),
                                ..BaseNode::default()
                            },
                            name: "from".to_string()
                        }),
                        lparen: vec![],
                        arguments: vec![],
                        rparen: vec![],
                    })),
                })),
                Statement::Expr(Box::new(ExprStmt {
                    base: BaseNode {
                        location: loc.get(3, 4, 3, 18),
                        ..BaseNode::default()
                    },
                    expression: Expression::PipeExpr(Box::new(PipeExpr {
                        base: BaseNode {
                            location: loc.get(3, 4, 3, 18),
                            ..BaseNode::default()
                        },
                        argument: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(3, 4, 3, 9),
                                ..BaseNode::default()
                            },
                            name: "howdy".to_string()
                        }),
                        call: CallExpr {
                            base: BaseNode {
                                location: loc.get(3, 11, 3, 18),
                                ..BaseNode::default()
                            },
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(3, 11, 3, 16),
                                    ..BaseNode::default()
                                },
                                name: "count".to_string()
                            }),
                            lparen: vec![],
                            arguments: vec![],
                            rparen: vec![],
                        }
                    }))
                })),
                Statement::Expr(Box::new(ExprStmt {
                    base: BaseNode {
                        location: loc.get(4, 4, 4, 16),
                        ..BaseNode::default()
                    },
                    expression: Expression::PipeExpr(Box::new(PipeExpr {
                        base: BaseNode {
                            location: loc.get(4, 4, 4, 16),
                            ..BaseNode::default()
                        },
                        argument: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(4, 4, 4, 9),
                                ..BaseNode::default()
                            },
                            name: "doody".to_string()
                        }),
                        call: CallExpr {
                            base: BaseNode {
                                location: loc.get(4, 11, 4, 16),
                                ..BaseNode::default()
                            },
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(4, 11, 4, 14),
                                    ..BaseNode::default()
                                },
                                name: "sum".to_string()
                            }),
                            lparen: vec![],
                            arguments: vec![],
                            rparen: vec![],
                        }
                    }))
                }))
            ],
            eof: vec![],
        },
    )
}

#[test]
fn index_expression() {
    let mut p = Parser::new(r#"a[3]"#);
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
                expression: Expression::Index(Box::new(IndexExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 5),
                        ..BaseNode::default()
                    },
                    array: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            ..BaseNode::default()
                        },
                        name: "a".to_string()
                    }),
                    lbrack: vec![],
                    index: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 3, 1, 4),
                            ..BaseNode::default()
                        },
                        value: 3
                    }),
                    rbrack: vec![],
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn nested_index_expression() {
    let mut p = Parser::new(r#"a[3][5]"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 8),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 8),
                    ..BaseNode::default()
                },
                expression: Expression::Index(Box::new(IndexExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 8),
                        ..BaseNode::default()
                    },
                    array: Expression::Index(Box::new(IndexExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 5),
                            ..BaseNode::default()
                        },
                        array: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 2),
                                ..BaseNode::default()
                            },
                            name: "a".to_string()
                        }),
                        lbrack: vec![],
                        index: Expression::Integer(IntegerLit {
                            base: BaseNode {
                                location: loc.get(1, 3, 1, 4),
                                ..BaseNode::default()
                            },
                            value: 3
                        }),
                        rbrack: vec![],
                    })),
                    lbrack: vec![],
                    index: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 6, 1, 7),
                            ..BaseNode::default()
                        },
                        value: 5
                    }),
                    rbrack: vec![],
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn access_indexed_object_returned_from_function_call() {
    let mut p = Parser::new(r#"f()[3]"#);
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
                expression: Expression::Index(Box::new(IndexExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 7),
                        ..BaseNode::default()
                    },
                    array: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 4),
                            ..BaseNode::default()
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 2),
                                ..BaseNode::default()
                            },
                            name: "f".to_string()
                        }),
                        lparen: vec![],
                        arguments: vec![],
                        rparen: vec![],
                    })),
                    lbrack: vec![],
                    index: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 6),
                            ..BaseNode::default()
                        },
                        value: 3
                    }),
                    rbrack: vec![],
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn index_with_member_expressions() {
    let mut p = Parser::new(r#"a.b["c"]"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
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
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 9),
                    ..BaseNode::default()
                },
                expression: Expression::Member(Box::new(MemberExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 9),
                        ..BaseNode::default()
                    },
                    object: Expression::Member(Box::new(MemberExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 4),
                            ..BaseNode::default()
                        },
                        object: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 2),
                                ..BaseNode::default()
                            },
                            name: "a".to_string()
                        }),
                        lbrack: vec![],
                        property: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 3, 1, 4),
                                ..BaseNode::default()
                            },
                            name: "b".to_string()
                        }),
                        rbrack: vec![],
                    })),
                    lbrack: vec![],
                    property: PropertyKey::StringLit(StringLit {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 8),
                            ..BaseNode::default()
                        },
                        value: "c".to_string()
                    }),
                    rbrack: vec![],
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn index_with_member_with_call_expression() {
    let mut p = Parser::new(r#"a.b()["c"]"#);
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
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 11),
                    ..BaseNode::default()
                },
                expression: Expression::Member(Box::new(MemberExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 11),
                        ..BaseNode::default()
                    },
                    object: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 6),
                            ..BaseNode::default()
                        },
                        callee: Expression::Member(Box::new(MemberExpr {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 4),
                                ..BaseNode::default()
                            },
                            object: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 1, 1, 2),
                                    ..BaseNode::default()
                                },
                                name: "a".to_string()
                            }),
                            lbrack: vec![],
                            property: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 3, 1, 4),
                                    ..BaseNode::default()
                                },
                                name: "b".to_string()
                            }),
                            rbrack: vec![],
                        })),
                        lparen: vec![],
                        arguments: vec![],
                        rparen: vec![],
                    })),
                    lbrack: vec![],
                    property: PropertyKey::StringLit(StringLit {
                        base: BaseNode {
                            location: loc.get(1, 7, 1, 10),
                            ..BaseNode::default()
                        },
                        value: "c".to_string()
                    }),
                    rbrack: vec![],
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn expressions_with_function_calls() {
    let mut p = Parser::new(r#"a = foo() == 10"#);
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
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 16),
                    ..BaseNode::default()
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 2),
                        ..BaseNode::default()
                    },
                    name: "a".to_string()
                },
                init: Expression::Binary(Box::new(BinaryExpr {
                    base: BaseNode {
                        location: loc.get(1, 5, 1, 16),
                        ..BaseNode::default()
                    },
                    operator: Operator::EqualOperator,
                    left: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 10),
                            ..BaseNode::default()
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 5, 1, 8),
                                ..BaseNode::default()
                            },
                            name: "foo".to_string()
                        }),
                        lparen: vec![],
                        arguments: vec![],
                        rparen: vec![],
                    })),
                    right: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 14, 1, 16),
                            ..BaseNode::default()
                        },
                        value: 10
                    })
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn conditional() {
    let mut p = Parser::new(r#"a = if true then 0 else 1"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 26),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 26),
                    ..BaseNode::default()
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 2),
                        ..BaseNode::default()
                    },
                    name: "a".to_string()
                },
                init: Expression::Conditional(Box::new(ConditionalExpr {
                    base: BaseNode {
                        location: loc.get(1, 5, 1, 26),
                        ..BaseNode::default()
                    },
                    test: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 8, 1, 12),
                            ..BaseNode::default()
                        },
                        name: "true".to_string()
                    }),
                    consequent: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 18, 1, 19),
                            ..BaseNode::default()
                        },
                        value: 0
                    }),
                    alternate: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 25, 1, 26),
                            ..BaseNode::default()
                        },
                        value: 1
                    }),
                    tk_if: vec![],
                    tk_then: vec![],
                    tk_else: vec![],
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn conditional_with_unary_logical_operators() {
    let mut p = Parser::new(
        r#"a = if exists b or c < d and not e == f then not exists (g - h) else exists exists i"#,
    );
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 85),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 85),
                    ..BaseNode::default()
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 2),
                        ..BaseNode::default()
                    },
                    name: "a".to_string()
                },
                init: Expression::Conditional(Box::new(ConditionalExpr {
                    base: BaseNode {
                        location: loc.get(1, 5, 1, 85),
                        ..BaseNode::default()
                    },
                    test: Expression::Logical(Box::new(LogicalExpr {
                        base: BaseNode {
                            location: loc.get(1, 8, 1, 40),
                            ..BaseNode::default()
                        },
                        operator: LogicalOperator::OrOperator,
                        left: Expression::Unary(Box::new(UnaryExpr {
                            base: BaseNode {
                                location: loc.get(1, 8, 1, 16),
                                ..BaseNode::default()
                            },
                            operator: Operator::ExistsOperator,
                            argument: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 15, 1, 16),
                                    ..BaseNode::default()
                                },
                                name: "b".to_string()
                            })
                        })),
                        right: Expression::Logical(Box::new(LogicalExpr {
                            base: BaseNode {
                                location: loc.get(1, 20, 1, 40),
                                ..BaseNode::default()
                            },
                            operator: LogicalOperator::AndOperator,
                            left: Expression::Binary(Box::new(BinaryExpr {
                                base: BaseNode {
                                    location: loc.get(1, 20, 1, 25),
                                    ..BaseNode::default()
                                },
                                operator: Operator::LessThanOperator,
                                left: Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 20, 1, 21),
                                        ..BaseNode::default()
                                    },
                                    name: "c".to_string()
                                }),
                                right: Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 24, 1, 25),
                                        ..BaseNode::default()
                                    },
                                    name: "d".to_string()
                                })
                            })),
                            right: Expression::Unary(Box::new(UnaryExpr {
                                base: BaseNode {
                                    location: loc.get(1, 30, 1, 40),
                                    ..BaseNode::default()
                                },
                                operator: Operator::NotOperator,
                                argument: Expression::Binary(Box::new(BinaryExpr {
                                    base: BaseNode {
                                        location: loc.get(1, 34, 1, 40),
                                        ..BaseNode::default()
                                    },
                                    operator: Operator::EqualOperator,
                                    left: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(1, 34, 1, 35),
                                            ..BaseNode::default()
                                        },
                                        name: "e".to_string()
                                    }),
                                    right: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(1, 39, 1, 40),
                                            ..BaseNode::default()
                                        },
                                        name: "f".to_string()
                                    })
                                }))
                            }))
                        }))
                    })),
                    consequent: Expression::Unary(Box::new(UnaryExpr {
                        base: BaseNode {
                            location: loc.get(1, 46, 1, 64),
                            ..BaseNode::default()
                        },
                        operator: Operator::NotOperator,
                        argument: Expression::Unary(Box::new(UnaryExpr {
                            base: BaseNode {
                                location: loc.get(1, 50, 1, 64),
                                ..BaseNode::default()
                            },
                            operator: Operator::ExistsOperator,
                            argument: Expression::Paren(Box::new(ParenExpr {
                                base: BaseNode {
                                    location: loc.get(1, 57, 1, 64),
                                    ..BaseNode::default()
                                },
                                lparen: vec![],
                                expression: Expression::Binary(Box::new(BinaryExpr {
                                    base: BaseNode {
                                        location: loc.get(1, 58, 1, 63),
                                        ..BaseNode::default()
                                    },
                                    operator: Operator::SubtractionOperator,
                                    left: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(1, 58, 1, 59),
                                            ..BaseNode::default()
                                        },
                                        name: "g".to_string()
                                    }),
                                    right: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(1, 62, 1, 63),
                                            ..BaseNode::default()
                                        },
                                        name: "h".to_string()
                                    })
                                })),
                                rparen: vec![],
                            }))
                        }))
                    })),
                    alternate: Expression::Unary(Box::new(UnaryExpr {
                        base: BaseNode {
                            location: loc.get(1, 70, 1, 85),
                            ..BaseNode::default()
                        },
                        operator: Operator::ExistsOperator,
                        argument: Expression::Unary(Box::new(UnaryExpr {
                            base: BaseNode {
                                location: loc.get(1, 77, 1, 85),
                                ..BaseNode::default()
                            },
                            operator: Operator::ExistsOperator,
                            argument: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 84, 1, 85),
                                    ..BaseNode::default()
                                },
                                name: "i".to_string()
                            })
                        }))
                    })),
                    tk_if: vec![],
                    tk_then: vec![],
                    tk_else: vec![],
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn nested_conditionals() {
    let mut p = Parser::new(
        r#"if if b < 0 then true else false
                  then if c > 0 then 30 else 60
                  else if d == 0 then 90 else 120"#,
    );
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(p.source);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 3, 50),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 3, 50),
                    ..BaseNode::default()
                },
                expression: Expression::Conditional(Box::new(ConditionalExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 3, 50),
                        ..BaseNode::default()
                    },
                    test: Expression::Conditional(Box::new(ConditionalExpr {
                        base: BaseNode {
                            location: loc.get(1, 4, 1, 33),
                            ..BaseNode::default()
                        },
                        test: Expression::Binary(Box::new(BinaryExpr {
                            base: BaseNode {
                                location: loc.get(1, 7, 1, 12),
                                ..BaseNode::default()
                            },
                            operator: Operator::LessThanOperator,
                            left: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 7, 1, 8),
                                    ..BaseNode::default()
                                },
                                name: "b".to_string()
                            }),
                            right: Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 11, 1, 12),
                                    ..BaseNode::default()
                                },
                                value: 0
                            })
                        })),
                        consequent: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 18, 1, 22),
                                ..BaseNode::default()
                            },
                            name: "true".to_string()
                        }),
                        alternate: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 28, 1, 33),
                                ..BaseNode::default()
                            },
                            name: "false".to_string()
                        }),
                        tk_if: vec![],
                        tk_then: vec![],
                        tk_else: vec![],
                    })),
                    consequent: Expression::Conditional(Box::new(ConditionalExpr {
                        base: BaseNode {
                            location: loc.get(2, 24, 2, 48),
                            ..BaseNode::default()
                        },
                        test: Expression::Binary(Box::new(BinaryExpr {
                            base: BaseNode {
                                location: loc.get(2, 27, 2, 32),
                                ..BaseNode::default()
                            },
                            operator: Operator::GreaterThanOperator,
                            left: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(2, 27, 2, 28),
                                    ..BaseNode::default()
                                },
                                name: "c".to_string()
                            }),
                            right: Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(2, 31, 2, 32),
                                    ..BaseNode::default()
                                },
                                value: 0
                            })
                        })),
                        consequent: Expression::Integer(IntegerLit {
                            base: BaseNode {
                                location: loc.get(2, 38, 2, 40),
                                ..BaseNode::default()
                            },
                            value: 30
                        }),
                        alternate: Expression::Integer(IntegerLit {
                            base: BaseNode {
                                location: loc.get(2, 46, 2, 48),
                                ..BaseNode::default()
                            },
                            value: 60
                        }),
                        tk_if: vec![],
                        tk_then: vec![],
                        tk_else: vec![],
                    })),
                    alternate: Expression::Conditional(Box::new(ConditionalExpr {
                        base: BaseNode {
                            location: loc.get(3, 24, 3, 50),
                            ..BaseNode::default()
                        },
                        test: Expression::Binary(Box::new(BinaryExpr {
                            base: BaseNode {
                                location: loc.get(3, 27, 3, 33),
                                ..BaseNode::default()
                            },
                            operator: Operator::EqualOperator,
                            left: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(3, 27, 3, 28),
                                    ..BaseNode::default()
                                },
                                name: "d".to_string()
                            }),
                            right: Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(3, 32, 3, 33),
                                    ..BaseNode::default()
                                },
                                value: 0
                            })
                        })),
                        consequent: Expression::Integer(IntegerLit {
                            base: BaseNode {
                                location: loc.get(3, 39, 3, 41),
                                ..BaseNode::default()
                            },
                            value: 90
                        }),
                        alternate: Expression::Integer(IntegerLit {
                            base: BaseNode {
                                location: loc.get(3, 47, 3, 50),
                                ..BaseNode::default()
                            },
                            value: 120
                        }),
                        tk_if: vec![],
                        tk_then: vec![],
                        tk_else: vec![],
                    })),
                    tk_if: vec![],
                    tk_then: vec![],
                    tk_else: vec![],
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn parse_testcase() {
    let mut parser = Parser::new(r#"testcase my_test { a = 1 }"#);
    let parsed = parser.parse_file("".to_string());
    let loc = Locator::new(parser.source);
    let expected = vec![Statement::TestCase(Box::new(TestCaseStmt {
        base: BaseNode {
            location: loc.get(1, 1, 1, 27),
            ..BaseNode::default()
        },
        id: Identifier {
            base: BaseNode {
                location: loc.get(1, 10, 1, 17),
                ..BaseNode::default()
            },
            name: "my_test".to_string(),
        },
        extends: None,
        block: Block {
            base: BaseNode {
                location: loc.get(1, 18, 1, 27),
                ..BaseNode::default()
            },
            lbrace: vec![],
            body: vec![Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 20, 1, 25),
                    ..BaseNode::default()
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 20, 1, 21),
                        ..BaseNode::default()
                    },
                    name: "a".to_string(),
                },
                init: Expression::Integer(IntegerLit {
                    base: BaseNode {
                        location: loc.get(1, 24, 1, 25),
                        ..BaseNode::default()
                    },
                    value: 1,
                }),
            }))],
            rbrace: vec![],
        },
    }))];

    assert_eq!(expected, parsed.body);
}

#[test]
fn parse_testcase_extends() {
    let mut parser = Parser::new(r#"testcase my_test extends "other_test" { a = 1 }"#);
    let parsed = parser.parse_file("".to_string());
    let loc = Locator::new(parser.source);
    let expected = vec![Statement::TestCase(Box::new(TestCaseStmt {
        base: BaseNode {
            location: loc.get(1, 1, 1, 48),
            ..BaseNode::default()
        },
        id: Identifier {
            base: BaseNode {
                location: loc.get(1, 10, 1, 17),
                ..BaseNode::default()
            },
            name: "my_test".to_string(),
        },
        extends: Some(StringLit {
            base: BaseNode {
                location: loc.get(1, 26, 1, 38),
                ..BaseNode::default()
            },
            value: "other_test".to_string(),
        }),
        block: Block {
            base: BaseNode {
                location: loc.get(1, 39, 1, 48),
                ..BaseNode::default()
            },
            lbrace: vec![],
            body: vec![Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 41, 1, 46),
                    ..BaseNode::default()
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 41, 1, 42),
                        ..BaseNode::default()
                    },
                    name: "a".to_string(),
                },
                init: Expression::Integer(IntegerLit {
                    base: BaseNode {
                        location: loc.get(1, 45, 1, 46),
                        ..BaseNode::default()
                    },
                    value: 1,
                }),
            }))],
            rbrace: vec![],
        },
    }))];

    assert_eq!(expected, parsed.body);
}
