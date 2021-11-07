use super::*;
use crate::ast;

use crate::ast::tests::Locator;
use crate::ast::Statement::Variable;
use chrono;
use pretty_assertions::assert_eq;

mod arrow_function;
mod from;
mod operator_precedence;
mod strings;
mod types;

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
    let loc = Locator::new(&p.source[..]);
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
    let loc = Locator::new(&p.source[..]);
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
    let loc = Locator::new(&p.source[..]);
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
fn import() {
    let mut p = Parser::new(r#"import "path/foo""#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
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
    let loc = Locator::new(&p.source[..]);
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
    let loc = Locator::new(&p.source[..]);
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
    let loc = Locator::new(&p.source[..]);
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
    let loc = Locator::new(&p.source[..]);
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
    let loc = Locator::new(&p.source[..]);
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
    let loc = Locator::new(&p.source[..]);
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
    let loc = Locator::new(&p.source[..]);
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
            body: vec![Statement::Builtin(Box::new(BuiltinStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 13),
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
fn test_statement() {
    let mut p = Parser::new(r#"test mean = {want: 0, got: 0}"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 30),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Test(Box::new(TestStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 30),
                    ..BaseNode::default()
                },
                assignment: VariableAssgn {
                    base: BaseNode {
                        location: loc.get(1, 6, 1, 30),
                        ..BaseNode::default()
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(1, 6, 1, 10),
                            ..BaseNode::default()
                        },
                        name: "mean".to_string()
                    },
                    init: Expression::Object(Box::new(ObjectExpr {
                        base: BaseNode {
                            location: loc.get(1, 13, 1, 30),
                            ..BaseNode::default()
                        },
                        lbrace: vec![],
                        with: None,
                        properties: vec![
                            Property {
                                base: BaseNode {
                                    location: loc.get(1, 14, 1, 21),
                                    ..BaseNode::default()
                                },
                                key: PropertyKey::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 14, 1, 18),
                                        ..BaseNode::default()
                                    },
                                    name: "want".to_string()
                                }),
                                separator: vec![],
                                value: Some(Expression::Integer(IntegerLit {
                                    base: BaseNode {
                                        location: loc.get(1, 20, 1, 21),
                                        ..BaseNode::default()
                                    },
                                    value: 0
                                })),
                                comma: vec![],
                            },
                            Property {
                                base: BaseNode {
                                    location: loc.get(1, 23, 1, 29),
                                    ..BaseNode::default()
                                },
                                key: PropertyKey::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 23, 1, 26),
                                        ..BaseNode::default()
                                    },
                                    name: "got".to_string()
                                }),
                                separator: vec![],
                                value: Some(Expression::Integer(IntegerLit {
                                    base: BaseNode {
                                        location: loc.get(1, 28, 1, 29),
                                        ..BaseNode::default()
                                    },
                                    value: 0
                                })),
                                comma: vec![],
                            }
                        ],
                        rbrace: vec![],
                    }))
                }
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn shebang() {
    let mut p = Parser::new(
        r#"#! /usr/bin/env flux
			from()"#,
    );
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
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
                                text: String::from("#! /usr/bin/env flux"),
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
fn linebreak_shebang() {
    let mut p = Parser::new(
        "\n#! /usr/bin/env flux\n// Comment"
    );
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(2,1,2,21),
                ..BaseNode::default()
            },
            name: String::new(),
            metadata: String::from("parser-type=rust"),
            package: None,
            imports: Vec::new(),
            body: vec![
                Statement::Bad(
                    Box::new(BadStmt {
                        base: BaseNode {
                            location: loc.get(2,1,2,21),
                            ..BaseNode::default()
                        },
                        text: String::from("#! /usr/bin/env flux"),
                    },
                )),
            ],
            eof: vec![
                ast::Comment {
                    text: String::from("// Comment"),
                },
            ],
        },
    )
}

#[test]
fn whitespace_shebang() {
    let mut p = Parser::new(
        " #! /usr/bin/env flux\n// Comment"
    );
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1,2,1,22),
                ..BaseNode::default()
            },
            name: String::new(),
            metadata: String::from("parser-type=rust"),
            package: None,
            imports: Vec::new(),
            body: vec![
                Statement::Bad(
                    Box::new(BadStmt {
                        base: BaseNode {
                            location: loc.get(1,2,1,22),
                            ..BaseNode::default()
                        },
                        text: String::from("#! /usr/bin/env flux"),
                    },
                )),
            ],
            eof: vec![
                ast::Comment {
                    text: String::from("// Comment"),
                },
            ],
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
    let loc = Locator::new(&p.source[..]);
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
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(2, 1, 2, 12),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Builtin(Box::new(BuiltinStmt {
                base: BaseNode {
                    location: loc.get(2, 1, 2, 12),
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
    let loc = Locator::new(&p.source[..]);
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
fn regex_literal() {
    let mut p = Parser::new(r#"/.*/"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
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
                expression: Expression::Regexp(RegexpLit {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 12),
                        ..BaseNode::default()
                    },
                    value: "a/b\\\\c\\d".to_string()
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
    let loc = Locator::new(&p.source[..]);
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
fn regex_match_operators() {
    let mut p = Parser::new(r#""a" =~ /.*/ and "b" !~ /c$/"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
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
    let loc = Locator::new(&p.source[..]);
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
    let loc = Locator::new(&p.source[..]);
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
    let loc = Locator::new(&p.source[..]);
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
    let loc = Locator::new(&p.source[..]);
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
    let loc = Locator::new(&p.source[..]);
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
    let loc = Locator::new(&p.source[..]);
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
    let loc = Locator::new(&p.source[..]);
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
    let loc = Locator::new(&p.source[..]);
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
    let loc = Locator::new(&p.source[..]);
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
    let loc = Locator::new(&p.source[..]);
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
    let loc = Locator::new(&p.source[..]);
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
    let loc = Locator::new(&p.source[..]);
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
    let loc = Locator::new(&p.source[..]);
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
fn map_member_expressions() {
    let mut p = Parser::new(
        r#"m = {key1: 1, key2:"value2"}
			m.key1
			m["key2"]
			"#,
    );
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
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
    let loc = Locator::new(&p.source[..]);
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
    let loc = Locator::new(&p.source[..]);
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

#[test]
fn index_expression() {
    let mut p = Parser::new(r#"a[3]"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
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
    let loc = Locator::new(&p.source[..]);
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
fn index_with_unclosed_bracket() {
    let mut p = Parser::new(r#"a[b()"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 6),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 6),
                    ..BaseNode::default()
                },
                expression: Expression::Index(Box::new(IndexExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 6),
                        errors: vec!["expected RBRACK, got EOF".to_string()],
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
                    index: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 3, 1, 6),
                            ..BaseNode::default()
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 3, 1, 4),
                                ..BaseNode::default()
                            },
                            name: "b".to_string()
                        }),
                        lparen: vec![],
                        arguments: vec![],
                        rparen: vec![],
                    })),
                    rbrack: vec![],
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn index_with_unbalanced_parenthesis() {
    let mut p = Parser::new(r#"a[b(]"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 6),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 6),
                    ..BaseNode::default()
                },
                expression: Expression::Index(Box::new(IndexExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 6),
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
                    index: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 3, 1, 6),
                            errors: vec!["expected RPAREN, got RBRACK".to_string()],
                            ..BaseNode::default()
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 3, 1, 4),
                                ..BaseNode::default()
                            },
                            name: "b".to_string()
                        }),
                        lparen: vec![],
                        arguments: vec![],
                        rparen: vec![],
                    })),
                    rbrack: vec![],
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn index_with_unexpected_rparen() {
    let mut p = Parser::new(r#"a[b)]"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 6),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 6),
                    ..BaseNode::default()
                },
                expression: Expression::Index(Box::new(IndexExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 6),
                        errors: vec!["invalid expression @1:4-1:5: )".to_string()],
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
                    index: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 3, 1, 4),
                            ..BaseNode::default()
                        },
                        name: "b".to_string()
                    }),
                    rbrack: vec![],
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn binary_expression() {
    let mut p = Parser::new(r#"_value < 10.0"#);
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
                    operator: Operator::LessThanOperator,
                    left: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 7),
                            ..BaseNode::default()
                        },
                        name: "_value".to_string()
                    }),
                    right: Expression::Float(FloatLit {
                        base: BaseNode {
                            location: loc.get(1, 10, 1, 14),
                            ..BaseNode::default()
                        },
                        value: 10.0
                    })
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn member_expression_binary_expression() {
    let mut p = Parser::new(r#"r._value < 10.0"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
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
                expression: Expression::Binary(Box::new(BinaryExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 16),
                        ..BaseNode::default()
                    },
                    operator: Operator::LessThanOperator,
                    left: Expression::Member(Box::new(MemberExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 9),
                            ..BaseNode::default()
                        },
                        object: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 2),
                                ..BaseNode::default()
                            },
                            name: "r".to_string()
                        }),
                        lbrack: vec![],
                        property: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 3, 1, 9),
                                ..BaseNode::default()
                            },
                            name: "_value".to_string()
                        }),
                        rbrack: vec![],
                    })),
                    right: Expression::Float(FloatLit {
                        base: BaseNode {
                            location: loc.get(1, 12, 1, 16),
                            ..BaseNode::default()
                        },
                        value: 10.0
                    })
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn var_as_binary_expression_of_other_vars() {
    let mut p = Parser::new(
        r#"a = 1
            b = 2
            c = a + b
            d = a"#,
    );
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 4, 18),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 6),
                        ..BaseNode::default()
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            ..BaseNode::default()
                        },
                        name: "a".to_string()
                    },
                    init: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 6),
                            ..BaseNode::default()
                        },
                        value: 1
                    })
                })),
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(2, 13, 2, 18),
                        ..BaseNode::default()
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(2, 13, 2, 14),
                            ..BaseNode::default()
                        },
                        name: "b".to_string()
                    },
                    init: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(2, 17, 2, 18),
                            ..BaseNode::default()
                        },
                        value: 2
                    })
                })),
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(3, 13, 3, 22),
                        ..BaseNode::default()
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(3, 13, 3, 14),
                            ..BaseNode::default()
                        },
                        name: "c".to_string()
                    },
                    init: Expression::Binary(Box::new(BinaryExpr {
                        base: BaseNode {
                            location: loc.get(3, 17, 3, 22),
                            ..BaseNode::default()
                        },
                        operator: Operator::AdditionOperator,
                        left: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(3, 17, 3, 18),
                                ..BaseNode::default()
                            },
                            name: "a".to_string()
                        }),
                        right: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(3, 21, 3, 22),
                                ..BaseNode::default()
                            },
                            name: "b".to_string()
                        })
                    }))
                })),
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(4, 13, 4, 18),
                        ..BaseNode::default()
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(4, 13, 4, 14),
                            ..BaseNode::default()
                        },
                        name: "d".to_string()
                    },
                    init: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(4, 17, 4, 18),
                            ..BaseNode::default()
                        },
                        name: "a".to_string()
                    })
                }))
            ],
            eof: vec![],
        },
    )
}

#[test]
fn var_as_unary_expression_of_other_vars() {
    let mut p = Parser::new(
        r#"a = 5
            c = -a"#,
    );
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 2, 19),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 6),
                        ..BaseNode::default()
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            ..BaseNode::default()
                        },
                        name: "a".to_string()
                    },
                    init: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 6),
                            ..BaseNode::default()
                        },
                        value: 5
                    })
                })),
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(2, 13, 2, 19),
                        ..BaseNode::default()
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(2, 13, 2, 14),
                            ..BaseNode::default()
                        },
                        name: "c".to_string()
                    },
                    init: Expression::Unary(Box::new(UnaryExpr {
                        base: BaseNode {
                            location: loc.get(2, 17, 2, 19),
                            ..BaseNode::default()
                        },
                        operator: Operator::SubtractionOperator,
                        argument: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(2, 18, 2, 19),
                                ..BaseNode::default()
                            },
                            name: "a".to_string()
                        })
                    }))
                }))
            ],
            eof: vec![],
        },
    )
}

#[test]
fn var_as_both_binary_and_unary_expressions() {
    let mut p = Parser::new(
        r#"a = 5
            c = 10 * -a"#,
    );
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 2, 24),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 6),
                        ..BaseNode::default()
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            ..BaseNode::default()
                        },
                        name: "a".to_string()
                    },
                    init: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 6),
                            ..BaseNode::default()
                        },
                        value: 5
                    })
                })),
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(2, 13, 2, 24),
                        ..BaseNode::default()
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(2, 13, 2, 14),
                            ..BaseNode::default()
                        },
                        name: "c".to_string()
                    },
                    init: Expression::Binary(Box::new(BinaryExpr {
                        base: BaseNode {
                            location: loc.get(2, 17, 2, 24),
                            ..BaseNode::default()
                        },
                        operator: Operator::MultiplicationOperator,
                        left: Expression::Integer(IntegerLit {
                            base: BaseNode {
                                location: loc.get(2, 17, 2, 19),
                                ..BaseNode::default()
                            },
                            value: 10
                        }),
                        right: Expression::Unary(Box::new(UnaryExpr {
                            base: BaseNode {
                                location: loc.get(2, 22, 2, 24),
                                ..BaseNode::default()
                            },
                            operator: Operator::SubtractionOperator,
                            argument: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(2, 23, 2, 24),
                                    ..BaseNode::default()
                                },
                                name: "a".to_string()
                            })
                        }))
                    }))
                }))
            ],
            eof: vec![],
        },
    )
}

#[test]
fn unary_expressions_within_logical_expression() {
    let mut p = Parser::new(
        r#"a = 5.0
            10.0 * -a == -0.5 or a == 6.0"#,
    );
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 2, 42),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 8),
                        ..BaseNode::default()
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            ..BaseNode::default()
                        },
                        name: "a".to_string()
                    },
                    init: Expression::Float(FloatLit {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 8),
                            ..BaseNode::default()
                        },
                        value: 5.0
                    })
                })),
                Statement::Expr(Box::new(ExprStmt {
                    base: BaseNode {
                        location: loc.get(2, 13, 2, 42),
                        ..BaseNode::default()
                    },
                    expression: Expression::Logical(Box::new(LogicalExpr {
                        base: BaseNode {
                            location: loc.get(2, 13, 2, 42),
                            ..BaseNode::default()
                        },
                        operator: LogicalOperator::OrOperator,
                        left: Expression::Binary(Box::new(BinaryExpr {
                            base: BaseNode {
                                location: loc.get(2, 13, 2, 30),
                                ..BaseNode::default()
                            },
                            operator: Operator::EqualOperator,
                            left: Expression::Binary(Box::new(BinaryExpr {
                                base: BaseNode {
                                    location: loc.get(2, 13, 2, 22),
                                    ..BaseNode::default()
                                },
                                operator: Operator::MultiplicationOperator,
                                left: Expression::Float(FloatLit {
                                    base: BaseNode {
                                        location: loc.get(2, 13, 2, 17),
                                        ..BaseNode::default()
                                    },
                                    value: 10.0
                                }),
                                right: Expression::Unary(Box::new(UnaryExpr {
                                    base: BaseNode {
                                        location: loc.get(2, 20, 2, 22),
                                        ..BaseNode::default()
                                    },
                                    operator: Operator::SubtractionOperator,
                                    argument: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(2, 21, 2, 22),
                                            ..BaseNode::default()
                                        },
                                        name: "a".to_string()
                                    })
                                }))
                            })),
                            right: Expression::Unary(Box::new(UnaryExpr {
                                base: BaseNode {
                                    location: loc.get(2, 26, 2, 30),
                                    ..BaseNode::default()
                                },
                                operator: Operator::SubtractionOperator,
                                argument: Expression::Float(FloatLit {
                                    base: BaseNode {
                                        location: loc.get(2, 27, 2, 30),
                                        ..BaseNode::default()
                                    },
                                    value: 0.5
                                })
                            }))
                        })),
                        right: Expression::Binary(Box::new(BinaryExpr {
                            base: BaseNode {
                                location: loc.get(2, 34, 2, 42),
                                ..BaseNode::default()
                            },
                            operator: Operator::EqualOperator,
                            left: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(2, 34, 2, 35),
                                    ..BaseNode::default()
                                },
                                name: "a".to_string()
                            }),
                            right: Expression::Float(FloatLit {
                                base: BaseNode {
                                    location: loc.get(2, 39, 2, 42),
                                    ..BaseNode::default()
                                },
                                value: 6.0
                            })
                        }))
                    }))
                }))
            ],
            eof: vec![],
        },
    )
}

#[test]
fn unary_expression_with_member_expression() {
    let mut p = Parser::new(r#"not m.b"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
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
                expression: Expression::Unary(Box::new(UnaryExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 8),
                        ..BaseNode::default()
                    },
                    operator: Operator::NotOperator,
                    argument: Expression::Member(Box::new(MemberExpr {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 8),
                            ..BaseNode::default()
                        },
                        object: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 5, 1, 6),
                                ..BaseNode::default()
                            },
                            name: "m".to_string()
                        }),
                        lbrack: vec![],
                        property: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 7, 1, 8),
                                ..BaseNode::default()
                            },
                            name: "b".to_string()
                        }),
                        rbrack: vec![],
                    }))
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn unary_expressions_with_too_many_comments() {
    let mut p = Parser::new(
        r#"// define a
a = 5.0
// eval this
10.0 * -a == -0.5
	// or this
	or a == 6.0"#,
    );
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(2, 1, 6, 13),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(2, 1, 2, 8),
                        ..BaseNode::default()
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(2, 1, 2, 2),
                            comments: vec![ast::Comment {
                                text: String::from("// define a\n"),
                            }],
                            ..BaseNode::default()
                        },
                        name: "a".to_string()
                    },
                    init: Expression::Float(FloatLit {
                        base: BaseNode {
                            location: loc.get(2, 5, 2, 8),
                            ..BaseNode::default()
                        },
                        value: 5.0
                    })
                })),
                Statement::Expr(Box::new(ExprStmt {
                    base: BaseNode {
                        location: loc.get(4, 1, 6, 13),
                        ..BaseNode::default()
                    },
                    expression: Expression::Logical(Box::new(LogicalExpr {
                        base: BaseNode {
                            location: loc.get(4, 1, 6, 13),
                            comments: vec![ast::Comment {
                                text: String::from("// or this\n"),
                            }],
                            ..BaseNode::default()
                        },
                        operator: LogicalOperator::OrOperator,
                        left: Expression::Binary(Box::new(BinaryExpr {
                            base: BaseNode {
                                location: loc.get(4, 1, 4, 18),
                                ..BaseNode::default()
                            },
                            operator: Operator::EqualOperator,
                            left: Expression::Binary(Box::new(BinaryExpr {
                                base: BaseNode {
                                    location: loc.get(4, 1, 4, 10),
                                    ..BaseNode::default()
                                },
                                operator: Operator::MultiplicationOperator,
                                left: Expression::Float(FloatLit {
                                    base: BaseNode {
                                        location: loc.get(4, 1, 4, 5),
                                        comments: vec![ast::Comment {
                                            text: String::from("// eval this\n"),
                                        }],
                                        ..BaseNode::default()
                                    },
                                    value: 10.0
                                }),
                                right: Expression::Unary(Box::new(UnaryExpr {
                                    base: BaseNode {
                                        location: loc.get(4, 8, 4, 10),
                                        ..BaseNode::default()
                                    },
                                    operator: Operator::SubtractionOperator,
                                    argument: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(4, 9, 4, 10),
                                            ..BaseNode::default()
                                        },
                                        name: "a".to_string()
                                    })
                                }))
                            })),
                            right: Expression::Unary(Box::new(UnaryExpr {
                                base: BaseNode {
                                    location: loc.get(4, 14, 4, 18),
                                    ..BaseNode::default()
                                },
                                operator: Operator::SubtractionOperator,
                                argument: Expression::Float(FloatLit {
                                    base: BaseNode {
                                        location: loc.get(4, 15, 4, 18),
                                        ..BaseNode::default()
                                    },
                                    value: 0.5
                                })
                            }))
                        })),
                        right: Expression::Binary(Box::new(BinaryExpr {
                            base: BaseNode {
                                location: loc.get(6, 5, 6, 13),
                                ..BaseNode::default()
                            },
                            operator: Operator::EqualOperator,
                            left: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(6, 5, 6, 6),
                                    ..BaseNode::default()
                                },
                                name: "a".to_string()
                            }),
                            right: Expression::Float(FloatLit {
                                base: BaseNode {
                                    location: loc.get(6, 10, 6, 13),
                                    ..BaseNode::default()
                                },
                                value: 6.0
                            })
                        }))
                    }))
                }))
            ],
            eof: vec![],
        },
    )
}

#[test]
fn expressions_with_function_calls() {
    let mut p = Parser::new(r#"a = foo() == 10"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
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
fn mix_unary_logical_and_binary_expressions() {
    let mut p = Parser::new(
        r#"
            not (f() == 6.0 * x) or fail()"#,
    );
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(2, 13, 2, 43),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(2, 13, 2, 43),
                    ..BaseNode::default()
                },
                expression: Expression::Logical(Box::new(LogicalExpr {
                    base: BaseNode {
                        location: loc.get(2, 13, 2, 43),
                        ..BaseNode::default()
                    },
                    operator: LogicalOperator::OrOperator,
                    left: Expression::Unary(Box::new(UnaryExpr {
                        base: BaseNode {
                            location: loc.get(2, 13, 2, 33),
                            ..BaseNode::default()
                        },
                        operator: Operator::NotOperator,
                        argument: Expression::Paren(Box::new(ParenExpr {
                            base: BaseNode {
                                location: loc.get(2, 17, 2, 33),
                                ..BaseNode::default()
                            },
                            lparen: vec![],
                            expression: Expression::Binary(Box::new(BinaryExpr {
                                base: BaseNode {
                                    location: loc.get(2, 18, 2, 32),
                                    ..BaseNode::default()
                                },
                                operator: Operator::EqualOperator,
                                left: Expression::Call(Box::new(CallExpr {
                                    base: BaseNode {
                                        location: loc.get(2, 18, 2, 21),
                                        ..BaseNode::default()
                                    },
                                    lparen: vec![],
                                    callee: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(2, 18, 2, 19),
                                            ..BaseNode::default()
                                        },
                                        name: "f".to_string()
                                    }),
                                    arguments: vec![],
                                    rparen: vec![],
                                })),
                                right: Expression::Binary(Box::new(BinaryExpr {
                                    base: BaseNode {
                                        location: loc.get(2, 25, 2, 32),
                                        ..BaseNode::default()
                                    },
                                    operator: Operator::MultiplicationOperator,
                                    left: Expression::Float(FloatLit {
                                        base: BaseNode {
                                            location: loc.get(2, 25, 2, 28),
                                            ..BaseNode::default()
                                        },
                                        value: 6.0
                                    }),
                                    right: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(2, 31, 2, 32),
                                            ..BaseNode::default()
                                        },
                                        name: "x".to_string()
                                    })
                                }))
                            })),
                            rparen: vec![],
                        }))
                    })),
                    right: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(2, 37, 2, 43),
                            ..BaseNode::default()
                        },
                        lparen: vec![],
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(2, 37, 2, 41),
                                ..BaseNode::default()
                            },
                            name: "fail".to_string()
                        }),
                        arguments: vec![],
                        rparen: vec![],
                    })),
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn mix_unary_logical_and_binary_expressions_with_extra_parens() {
    let mut p = Parser::new(
        r#"
            (not (f() == 6.0 * x) or fail())"#,
    );
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(2, 13, 2, 45),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(2, 13, 2, 45),
                    ..BaseNode::default()
                },
                expression: Expression::Paren(Box::new(ParenExpr {
                    base: BaseNode {
                        location: loc.get(2, 13, 2, 45),
                        ..BaseNode::default()
                    },
                    lparen: vec![],
                    expression: Expression::Logical(Box::new(LogicalExpr {
                        base: BaseNode {
                            location: loc.get(2, 14, 2, 44),
                            ..BaseNode::default()
                        },
                        operator: LogicalOperator::OrOperator,
                        left: Expression::Unary(Box::new(UnaryExpr {
                            base: BaseNode {
                                location: loc.get(2, 14, 2, 34),
                                ..BaseNode::default()
                            },
                            operator: Operator::NotOperator,
                            argument: Expression::Paren(Box::new(ParenExpr {
                                base: BaseNode {
                                    location: loc.get(2, 18, 2, 34),
                                    ..BaseNode::default()
                                },
                                lparen: vec![],
                                expression: Expression::Binary(Box::new(BinaryExpr {
                                    base: BaseNode {
                                        location: loc.get(2, 19, 2, 33),
                                        ..BaseNode::default()
                                    },
                                    operator: Operator::EqualOperator,
                                    left: Expression::Call(Box::new(CallExpr {
                                        base: BaseNode {
                                            location: loc.get(2, 19, 2, 22),
                                            ..BaseNode::default()
                                        },
                                        lparen: vec![],
                                        callee: Expression::Identifier(Identifier {
                                            base: BaseNode {
                                                location: loc.get(2, 19, 2, 20),
                                                ..BaseNode::default()
                                            },
                                            name: "f".to_string()
                                        }),
                                        arguments: vec![],
                                        rparen: vec![],
                                    })),
                                    right: Expression::Binary(Box::new(BinaryExpr {
                                        base: BaseNode {
                                            location: loc.get(2, 26, 2, 33),
                                            ..BaseNode::default()
                                        },
                                        operator: Operator::MultiplicationOperator,
                                        left: Expression::Float(FloatLit {
                                            base: BaseNode {
                                                location: loc.get(2, 26, 2, 29),
                                                ..BaseNode::default()
                                            },
                                            value: 6.0
                                        }),
                                        right: Expression::Identifier(Identifier {
                                            base: BaseNode {
                                                location: loc.get(2, 32, 2, 33),
                                                ..BaseNode::default()
                                            },
                                            name: "x".to_string()
                                        })
                                    }))
                                })),
                                rparen: vec![],
                            }))
                        })),
                        right: Expression::Call(Box::new(CallExpr {
                            base: BaseNode {
                                location: loc.get(2, 38, 2, 44),
                                ..BaseNode::default()
                            },
                            lparen: vec![],
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(2, 38, 2, 42),
                                    ..BaseNode::default()
                                },
                                name: "fail".to_string()
                            }),
                            arguments: vec![],
                            rparen: vec![],
                        })),
                    })),
                    rparen: vec![],
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn modulo_op_ints() {
    let mut p = Parser::new(r#"3 % 8"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 6),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 6),
                    ..BaseNode::default()
                },
                expression: Expression::Binary(Box::new(BinaryExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 6),
                        ..BaseNode::default()
                    },
                    operator: Operator::ModuloOperator,
                    left: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            ..BaseNode::default()
                        },
                        value: 3
                    }),
                    right: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 6),
                            ..BaseNode::default()
                        },
                        value: 8
                    })
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn modulo_op_floats() {
    let mut p = Parser::new(r#"8.3 % 3.1"#);
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
                    operator: Operator::ModuloOperator,
                    left: Expression::Float(FloatLit {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 4),
                            ..BaseNode::default()
                        },
                        value: 8.3
                    }),
                    right: Expression::Float(FloatLit {
                        base: BaseNode {
                            location: loc.get(1, 7, 1, 10),
                            ..BaseNode::default()
                        },
                        value: 3.1
                    })
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn power_op() {
    let mut p = Parser::new(r#"2 ^ 4"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 6),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 6),
                    ..BaseNode::default()
                },
                expression: Expression::Binary(Box::new(BinaryExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 6),
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
                    right: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 6),
                            ..BaseNode::default()
                        },
                        value: 4
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
    let loc = Locator::new(&p.source[..]);
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
    let loc = Locator::new(&p.source[..]);
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
    let loc = Locator::new(&p.source[..]);
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
fn duration_literal_all_units() {
    let mut p = Parser::new(r#"dur = 1y3mo2w1d4h1m30s1ms2µs70ns"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
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
fn duration_literal_months() {
    let mut p = Parser::new(r#"dur = 6mo"#);
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
    let loc = Locator::new(&p.source[..]);
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
    let loc = Locator::new(&p.source[..]);
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
fn data_time_literal_arg() {
    let mut p = Parser::new(r#"range(start: 2018-11-29T09:00:00)"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
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
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 34),
                    ..BaseNode::default()
                },
                expression: Expression::Call(Box::new(CallExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 34),
                        errors: vec!["expected RPAREN, got EOF".to_string()],
                        ..BaseNode::default()
                    },
                    arguments: vec![Expression::Object(Box::new(ObjectExpr {
                        base: BaseNode {
                            location: loc.get(1, 7, 1, 12),
                            ..BaseNode::default()
                        },
                        lbrace: vec![],
                        rbrace: vec![],
                        with: None,
                        properties: vec![Property {
                            base: BaseNode {
                                location: loc.get(1, 7, 1, 12),
                                errors: vec![
                                    "invalid expression @1:33-1:34: )".to_string(),
                                    "missing property value".to_string(),
                                ],
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
                            value: None,
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
fn date_time_literal() {
    let mut p = Parser::new(r#"now = 2018-11-29T09:00:00Z"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
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
    let loc = Locator::new(&p.source[..]);
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
fn function_call_with_unbalanced_braces() {
    let mut p = Parser::new(r#"from() |> range() |> map(fn: (r) => { return r._value )"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 56),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 56),
                    ..BaseNode::default()
                },
                expression: Expression::PipeExpr(Box::new(PipeExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 56),
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
                            location: loc.get(1, 22, 1, 56),
                            ..BaseNode::default()
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 22, 1, 25),
                                ..BaseNode::default()
                            },
                            name: "map".to_string()
                        }),
                        lparen: vec![],
                        arguments: vec![Expression::Object(Box::new(ObjectExpr {
                            base: BaseNode {
                                location: loc.get(1, 26, 1, 56),
                                ..BaseNode::default()
                            },
                            lbrace: vec![],
                            with: None,
                            properties: vec![Property {
                                base: BaseNode {
                                    location: loc.get(1, 26, 1, 56),
                                    ..BaseNode::default()
                                },
                                key: PropertyKey::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 26, 1, 28),
                                        ..BaseNode::default()
                                    },
                                    name: "fn".to_string()
                                }),
                                separator: vec![],
                                value: Some(Expression::Function(Box::new(FunctionExpr {
                                    base: BaseNode {
                                        location: loc.get(1, 30, 1, 56),
                                        ..BaseNode::default()
                                    },
                                    lparen: vec![],
                                    params: vec![Property {
                                        base: BaseNode {
                                            location: loc.get(1, 31, 1, 32),
                                            ..BaseNode::default()
                                        },
                                        key: PropertyKey::Identifier(Identifier {
                                            base: BaseNode {
                                                location: loc.get(1, 31, 1, 32),
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
                                            location: loc.get(1, 37, 1, 56),
                                            errors: vec!["expected RBRACE, got RPAREN".to_string()],
                                            ..BaseNode::default()
                                        },
                                        lbrace: vec![],
                                        body: vec![Statement::Return(Box::new(ReturnStmt {
                                            base: BaseNode {
                                                location: loc.get(1, 39, 1, 54),
                                                ..BaseNode::default()
                                            },
                                            argument: Expression::Member(Box::new(MemberExpr {
                                                base: BaseNode {
                                                    location: loc.get(1, 46, 1, 54),
                                                    ..BaseNode::default()
                                                },
                                                object: Expression::Identifier(Identifier {
                                                    base: BaseNode {
                                                        location: loc.get(1, 46, 1, 47),
                                                        ..BaseNode::default()
                                                    },
                                                    name: "r".to_string()
                                                }),
                                                lbrack: vec![],
                                                property: PropertyKey::Identifier(Identifier {
                                                    base: BaseNode {
                                                        location: loc.get(1, 48, 1, 54),
                                                        ..BaseNode::default()
                                                    },
                                                    name: "_value".to_string()
                                                }),
                                                rbrack: vec![],
                                            }))
                                        }))],
                                        rbrace: vec![],
                                    }),
                                }))),
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

// TODO(affo): that error is injected by ast.Check().
#[test]
fn illegal_statement_token() {
    let mut p = Parser::new(r#"@ ident"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
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
            body: vec![
                Statement::Bad(Box::new(BadStmt {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 2),
                        // errors: vec!["invalid statement @1:1-1:2: @".to_string()]
                        ..BaseNode::default()
                    },
                    text: "@".to_string()
                })),
                Statement::Expr(Box::new(ExprStmt {
                    base: BaseNode {
                        location: loc.get(1, 3, 1, 8),
                        ..BaseNode::default()
                    },
                    expression: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 3, 1, 8),
                            ..BaseNode::default()
                        },
                        name: "ident".to_string()
                    })
                }))
            ],
            eof: vec![],
        },
    )
}

// TODO(affo): that error is injected by ast.Check().
#[test]
fn multiple_idents_in_parens() {
    let mut p = Parser::new(r#"(a b)"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(1, 1, 1, 6),
                ..BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(Box::new(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 6),
                    ..BaseNode::default()
                },
                expression: Expression::Paren(Box::new(ParenExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 6),
                        ..BaseNode::default()
                    },
                    lparen: vec![],
                    expression: Expression::Binary(Box::new(BinaryExpr {
                        // TODO(affo): ast.Check would add the error "expected an operator between two expressions".
                        base: BaseNode {
                            location: loc.get(1, 2, 1, 5),
                            ..BaseNode::default()
                        },
                        operator: Operator::InvalidOperator,
                        left: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 2, 1, 3),
                                ..BaseNode::default()
                            },
                            name: "a".to_string()
                        }),
                        right: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 4, 1, 5),
                                ..BaseNode::default()
                            },
                            name: "b".to_string()
                        })
                    })),
                    rparen: vec![],
                }))
            }))],
            eof: vec![],
        },
    )
}

// TODO(affo): that error is injected by ast.Check().
// TODO(jsternberg): Parens aren't recorded correctly in the source and are mostly ignored.
#[test]
fn missing_left_hand_side() {
    let mut p = Parser::new(r#"(*b)"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
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
                expression: Expression::Paren(Box::new(ParenExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 5),
                        ..BaseNode::default()
                    },
                    lparen: vec![],
                    expression: Expression::Binary(Box::new(BinaryExpr {
                        // TODO(affo): this should be like this:
                        // base: BaseNode {location: ..., errors: vec!["missing left hand side of expression".to_string()] },
                        base: BaseNode {
                            location: loc.get(1, 2, 1, 4),
                            ..BaseNode::default()
                        },
                        operator: Operator::MultiplicationOperator,
                        left: Expression::Bad(Box::new(BadExpr {
                            base: BaseNode {
                                location: loc.get(1, 2, 1, 3),
                                ..BaseNode::default()
                            },
                            text: "invalid token for primary expression: MUL".to_string(),
                            expression: None
                        })),
                        right: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 3, 1, 4),
                                ..BaseNode::default()
                            },
                            name: "b".to_string()
                        })
                    })),
                    rparen: vec![],
                }))
            }))],
            eof: vec![],
        },
    )
}

// TODO(affo): that error is injected by ast.Check().
// TODO(jsternberg): Parens aren't recorded correctly in the source and are mostly ignored.
#[test]
fn missing_right_hand_side() {
    let mut p = Parser::new(r#"(a*)"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
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
                expression: Expression::Paren(Box::new(ParenExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 5),
                        ..BaseNode::default()
                    },
                    lparen: vec![],
                    expression: Expression::Binary(Box::new(BinaryExpr {
                        // TODO(affo): this should be like this:
                        // base: BaseNode {location: ..., errors: vec!["missing right hand side of expression".to_string()] },
                        base: BaseNode {
                            location: loc.get(1, 2, 1, 5),
                            ..BaseNode::default()
                        },
                        operator: Operator::MultiplicationOperator,
                        left: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 2, 1, 3),
                                ..BaseNode::default()
                            },
                            name: "a".to_string()
                        }),
                        right: Expression::Bad(Box::new(BadExpr {
                            base: BaseNode {
                                location: loc.get(1, 4, 1, 5),
                                ..BaseNode::default()
                            },
                            text: "invalid token for primary expression: RPAREN".to_string(),
                            expression: None
                        })),
                    })),
                    rparen: vec![],
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn illegal_expression() {
    let mut p = Parser::new(r#"(@)"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
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
                expression: Expression::Paren(Box::new(ParenExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 4),
                        errors: vec!["invalid expression @1:2-1:3: @".to_string()],
                        ..BaseNode::default()
                    },
                    lparen: vec![],
                    expression: Expression::Bad(Box::new(BadExpr {
                        base: BaseNode {
                            location: loc.get(1, 2, 1, 3),
                            ..BaseNode::default()
                        },
                        text: "@".to_string(),
                        expression: None
                    })),
                    rparen: vec![],
                }))
            }))],
            eof: vec![],
        },
    )
}

// NOTE(affo): this is slightly different from Go. We have a BadExpr in the body.
#[test]
fn missing_arrow_in_function_expression() {
    let mut p = Parser::new(r#"(a, b) a + b"#);
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
                expression: Expression::Function(Box::new(FunctionExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 13),
                        errors: vec![
                            "expected ARROW, got IDENT (a) at 1:8".to_string(),
                            "expected ARROW, got ADD (+) at 1:10".to_string(),
                            "expected ARROW, got IDENT (b) at 1:12".to_string(),
                            "expected ARROW, got EOF".to_string()
                        ],
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
                                name: "a".to_string()
                            }),
                            separator: vec![],
                            value: None,
                            comma: vec![],
                        },
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 5, 1, 6),
                                ..BaseNode::default()
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 5, 1, 6),
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
                    body: FunctionBody::Expr(Expression::Bad(Box::new(BadExpr {
                        base: BaseNode {
                            location: loc.get(1, 13, 1, 13),
                            ..BaseNode::default()
                        },
                        text: "invalid token for primary expression: EOF".to_string(),
                        expression: None
                    }))),
                }))
            }))],
            eof: vec![],
        },
    )
}

#[test]
fn integer_literal_overflow() {
    let mut p = Parser::new(r#"100000000000000000000000000000"#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
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

#[test]
fn parse_testcase() {
    let mut parser = Parser::new(r#"testcase my_test { a = 1 }"#);
    let parsed = parser.parse_file("".to_string());
    let loc = Locator::new(&parser.source[..]);
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
    let loc = Locator::new(&parser.source[..]);
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
