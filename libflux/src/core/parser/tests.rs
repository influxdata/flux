use super::*;
use crate::ast;

use chrono;
use pretty_assertions::assert_eq;

struct Locator<'a> {
    source: &'a str,
    lines: Vec<u32>,
}

impl<'a> Locator<'a> {
    fn new(source: &'a str) -> Self {
        let mut lines = Vec::new();
        lines.push(0);
        let ci = source.char_indices();
        for (i, c) in ci {
            match c {
                '\n' => lines.push((i + 1) as u32),
                _ => (),
            }
        }
        Self { source, lines }
    }

    fn get(&self, sl: u32, sc: u32, el: u32, ec: u32) -> SourceLocation {
        SourceLocation {
            file: Some("".to_string()),
            source: Some(self.get_src(sl, sc, el, ec).to_string()),
            start: ast::Position {
                line: sl,
                column: sc,
            },
            end: ast::Position {
                line: el,
                column: ec,
            },
        }
    }

    fn get_src(&self, sl: u32, sc: u32, el: u32, ec: u32) -> &str {
        let start_offset = self.lines.get(sl as usize - 1).expect("line not found") + sc - 1;
        let end_offset = self.lines.get(el as usize - 1).expect("line not found") + ec - 1;
        return &self.source[start_offset as usize..end_offset as usize];
    }
}

#[test]
fn string_interpolation_simple() {
    let mut p = Parser::new(r#""a + b = ${a + b}""#);
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
            eof: None,
        },
    )
}

#[test]
fn string_interpolation_multiple() {
    let mut p = Parser::new(r#""a = ${a} and b = ${b}""#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
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
            eof: None,
        },
    )
}

#[test]
fn string_interpolation_nested() {
    let mut p = Parser::new(r#""we ${"can ${"add" + "strings"}"} together""#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
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
            eof: None,
        },
    )
}

#[test]
fn string_interp_with_escapes() {
    let mut p = Parser::new(r#""string \"interpolation with ${"escapes"}\"""#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
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
            eof: None,
        },
    )
}

#[test]
fn bad_string_expression() {
    let mut p = Parser::new(r#"fn = (a) => "${a}"#);
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
                    lparen: None,
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
                        separator: None,
                        value: None,
                        comma: None,
                    }],
                    rparen: None,
                    arrow: None,
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
            eof: None,
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
            eof: None,
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
            eof: None,
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
            eof: None,
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
            eof: None,
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
            eof: None,
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
            eof: None,
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
                        lbrace: None,
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
                                separator: None,
                                value: Some(Expression::StringLit(StringLit {
                                    base: BaseNode {
                                        location: loc.get(2, 11, 2, 16),
                                        ..BaseNode::default()
                                    },
                                    value: "foo".to_string()
                                })),
                                comma: None,
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
                                separator: None,
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
                                comma: None,
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
                                separator: None,
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
                                comma: None,
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
                                separator: None,
                                value: Some(Expression::StringLit(StringLit {
                                    base: BaseNode {
                                        location: loc.get(5, 11, 5, 22),
                                        ..BaseNode::default()
                                    },
                                    value: "0 2 * * *".to_string()
                                })),
                                comma: None,
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
                                separator: None,
                                value: Some(Expression::Integer(IntegerLit {
                                    base: BaseNode {
                                        location: loc.get(6, 12, 6, 13),
                                        ..BaseNode::default()
                                    },
                                    value: 5
                                })),
                                comma: None,
                            }
                        ],
                        rbrace: None,
                    }))
                }))
            }))],
            eof: None,
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
                            lbrace: None,
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
                                    separator: None,
                                    value: Some(Expression::StringLit(StringLit {
                                        base: BaseNode {
                                            location: loc.get(2, 12, 2, 17),
                                            ..BaseNode::default()
                                        },
                                        value: "foo".to_string()
                                    })),
                                    comma: None,
                                },
                                Property {
                                    base: BaseNode {
                                        location: loc.get(3, 6, 3, 15),
                                        ..BaseNode::default()
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(3, 6, 3, 11),
                                            comments: Some(Box::new(Comment {
                                                lit: "// Name of task\n".to_string(),
                                                next: None,
                                            })),
                                            ..BaseNode::default()
                                        },
                                        name: "every".to_string()
                                    }),
                                    separator: None,
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
                                    comma: None,
                                }
                            ],
                            rbrace: Some(Box::new(Comment {
                                lit: "// Execution frequency of task\n".to_string(),
                                next: None,
                            })),
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
                                    comments: Some(Box::new(Comment {
                                        lit: "// Task will execute the following query\n"
                                            .to_string(),
                                        next: None,
                                    })),
                                    ..BaseNode::default()
                                },
                                name: "from".to_string()
                            }),
                            lparen: None,
                            arguments: vec![],
                            rparen: None,
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
                            lparen: None,
                            arguments: vec![],
                            rparen: None,
                        }
                    })),
                }))
            ],
            eof: None,
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
                        lbrack: None,
                        property: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 14, 1, 19),
                                ..BaseNode::default()
                            },
                            name: "state".to_string()
                        }),
                        rbrack: None,
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
            eof: None,
        },
    )
}

#[test]
fn builtin() {
    let mut p = Parser::new(r#"builtin from"#);
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
                }
            }))],
            eof: None,
        },
    )
}

#[test]
fn test_parse_type_expression_tvar() {
    let mut p = Parser::new(r#"A"#);
    let parsed = p.parse_type_expression();
    let loc = Locator::new(&p.source[..]);
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
            }),
            constraint: None,
        },
    )
}

#[test]
fn test_parse_type_expression_int() {
    let mut p = Parser::new(r#"int"#);
    let parsed = p.parse_type_expression();
    let loc = Locator::new(&p.source[..]);
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
            constraint: None,
        },
    )
}

#[test]
fn test_parse_type_expression_uint() {
    let mut p = Parser::new(r#"uint"#);
    let parsed = p.parse_type_expression();
    let loc = Locator::new(&p.source[..]);
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
            constraint: None,
        },
    )
}

#[test]
fn test_parse_type_expression_float() {
    let mut p = Parser::new(r#"float"#);
    let parsed = p.parse_type_expression();
    let loc = Locator::new(&p.source[..]);
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
            constraint: None,
        },
    )
}

#[test]
fn test_parse_type_expression_string() {
    let mut p = Parser::new(r#"string"#);
    let parsed = p.parse_type_expression();
    let loc = Locator::new(&p.source[..]);
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
            constraint: None,
        },
    )
}

#[test]
fn test_parse_type_expression_bool() {
    let mut p = Parser::new(r#"bool"#);
    let parsed = p.parse_type_expression();
    let loc = Locator::new(&p.source[..]);
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
            constraint: None
        },
    )
}

#[test]
fn test_parse_type_expression_time() {
    let mut p = Parser::new(r#"time"#);
    let parsed = p.parse_type_expression();
    let loc = Locator::new(&p.source[..]);
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
            constraint: None
        },
    )
}

#[test]
fn test_parse_type_expression_duration() {
    let mut p = Parser::new(r#"duration"#);
    let parsed = p.parse_type_expression();
    let loc = Locator::new(&p.source[..]);
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
            constraint: None
        },
    )
}

#[test]
fn test_parse_type_expression_bytes() {
    let mut p = Parser::new(r#"bytes"#);
    let parsed = p.parse_type_expression();
    let loc = Locator::new(&p.source[..]);
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
            constraint: None
        },
    )
}

#[test]
fn test_parse_type_expression_regexp() {
    let mut p = Parser::new(r#"regexp"#);
    let parsed = p.parse_type_expression();
    let loc = Locator::new(&p.source[..]);
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
            constraint: None,
        },
    )
}

#[test]
fn test_parse_type_expression_array_int() {
    let mut p = Parser::new(r#"[int]"#);
    let parsed = p.parse_type_expression();
    let loc = Locator::new(&p.source[..]);
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
                monotype: MonoType::Basic(NamedType {
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
            constraint: None,
        },
    )
}

#[test]
fn test_parse_type_expression_array_string() {
    let mut p = Parser::new(r#"[string]"#);
    let parsed = p.parse_type_expression();
    let loc = Locator::new(&p.source[..]);
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
                monotype: MonoType::Basic(NamedType {
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
            constraint: None
        },
    )
}

#[test]
fn test_parse_constraint_one_ident() {
    let mut p = Parser::new(r#"A : date"#);
    let parsed = p.parse_constraints();
    let loc = Locator::new(&p.source[..]);
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
fn test_parse_constraint_two_ident() {
    let mut p = Parser::new(r#"A: Addable + Subtractable"#);
    let parsed = p.parse_constraints();
    let loc = Locator::new(&p.source[..]);
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
    let loc = Locator::new(&p.source[..]);
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
                        lbrace: None,
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
                                separator: None,
                                value: Some(Expression::Integer(IntegerLit {
                                    base: BaseNode {
                                        location: loc.get(1, 20, 1, 21),
                                        ..BaseNode::default()
                                    },
                                    value: 0
                                })),
                                comma: None,
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
                                separator: None,
                                value: Some(Expression::Integer(IntegerLit {
                                    base: BaseNode {
                                        location: loc.get(1, 28, 1, 29),
                                        ..BaseNode::default()
                                    },
                                    value: 0
                                })),
                                comma: None,
                            }
                        ],
                        rbrace: None,
                    }))
                }
            }))],
            eof: None,
        },
    )
}

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
                    lparen: None,
                    arguments: vec![],
                    rparen: None,
                })),
            }))],
            eof: None,
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
                            comments: Some(Box::new(Comment {
                                lit: "// Comment\n".to_string(),
                                next: None,
                            })),
                            ..BaseNode::default()
                        },
                        name: "from".to_string()
                    }),
                    lparen: None,
                    arguments: vec![],
                    rparen: None,
                })),
            }))],
            eof: None,
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
                    lparen: None,
                    arguments: vec![],
                    rparen: None,
                }))
            }))],
            eof: None,
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
            eof: None,
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
            eof: None,
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
            eof: None,
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
            eof: None,
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
            eof: None,
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
            eof: None,
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
                    lbrack: None,
                    elements: vec![
                        ArrayItem {
                            expression: Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 10, 1, 11),
                                    ..BaseNode::default()
                                },
                                value: 1
                            }),
                            comma: None,
                        },
                        ArrayItem {
                            expression: Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 13, 1, 14),
                                    ..BaseNode::default()
                                },
                                value: 2
                            }),
                            comma: None,
                        },
                        ArrayItem {
                            expression: Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 16, 1, 17),
                                    ..BaseNode::default()
                                },
                                value: 3
                            }),
                            comma: None,
                        },
                        ArrayItem {
                            expression: Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 19, 1, 20),
                                    ..BaseNode::default()
                                },
                                value: 4
                            }),
                            comma: None,
                        }
                    ],
                    rbrack: None,
                }))
            }))],
            eof: None,
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
                    lbrack: None,
                    elements: vec![],
                    rbrack: None,
                }))
            }))],
            eof: None,
        },
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
                        lparen: None,
                        arguments: vec![],
                        rparen: None,
                    })),
                }))
            ],
            eof: None,
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
                        lparen: None,
                        arguments: vec![],
                        rparen: None,
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
                            lbrack: None,
                            property: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(2, 10, 2, 15),
                                    ..BaseNode::default()
                                },
                                name: "count".to_string()
                            }),
                            rbrack: None,
                        })),
                        lparen: None,
                        arguments: vec![],
                        rparen: None,
                    }))
                }))
            ],
            eof: None,
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
                        lparen: None,
                        arguments: vec![],
                        rparen: None,
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
                        lparen: None,
                        arguments: vec![],
                        rparen: None,
                    }
                }))
            }))],
            eof: None,
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
                            lbrack: None,
                            property: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 8, 1, 9),
                                    ..BaseNode::default()
                                },
                                name: "c".to_string()
                            }),
                            rbrack: None,
                        })),
                        lparen: None,
                        arguments: vec![Expression::Object(Box::new(ObjectExpr {
                            base: BaseNode {
                                location: loc.get(1, 10, 1, 13),
                                ..BaseNode::default()
                            },
                            lbrace: None,
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
                                separator: None,
                                value: Some(Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 12, 1, 13),
                                        ..BaseNode::default()
                                    },
                                    name: "e".to_string()
                                })),
                                comma: None,
                            }],
                            rbrace: None,
                        }))],
                        rparen: None,
                    }
                }))
            }))],
            eof: None,
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
                        lparen: None,
                        arguments: vec![],
                        rparen: None,
                    }
                })),
            }))],
            eof: None,
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
                        lbrack: None,
                        property: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 5, 1, 8),
                                ..BaseNode::default()
                            },
                            name: "bar".to_string()
                        }),
                        rbrack: None,
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
                        lparen: None,
                        arguments: vec![],
                        rparen: None,
                    }
                }))
            }))],
            eof: None,
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
                                lparen: None,
                                arguments: vec![],
                                rparen: None,
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
                                lparen: None,
                                arguments: vec![],
                                rparen: None,
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
                            lparen: None,
                            arguments: vec![],
                            rparen: None,
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
                        lparen: None,
                        arguments: vec![],
                        rparen: None,
                    }
                }))
            }))],
            eof: None,
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
                        lparen: None,
                        arguments: vec![],
                        rparen: None,
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
                        lparen: None,
                        arguments: vec![],
                        rparen: None,
                    }
                })),
            }))],
            eof: None,
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
                        lparen: None,
                        arguments: vec![],
                        rparen: None,
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
                        lparen: None,
                        arguments: vec![],
                        rparen: None,
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
                            lparen: None,
                            arguments: vec![],
                            rparen: None,
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
                            lparen: None,
                            arguments: vec![],
                            rparen: None,
                        }
                    }))
                }))
            ],
            eof: None,
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
                    lparen: None,
                    arguments: vec![Expression::Object(Box::new(ObjectExpr {
                        base: BaseNode {
                            location: loc.get(1, 6, 1, 31),
                            ..BaseNode::default()
                        },
                        lbrace: None,
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
                            separator: None,
                            value: Some(Expression::StringLit(StringLit {
                                base: BaseNode {
                                    location: loc.get(1, 13, 1, 31),
                                    ..BaseNode::default()
                                },
                                value: "telegraf/autogen".to_string()
                            })),
                            comma: None,
                        }],
                        rbrace: None,
                    }))],
                    rparen: None,
                }))
            }))],
            eof: None,
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
                        lbrace: None,
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
                                separator: None,
                                value: Some(Expression::Integer(IntegerLit {
                                    base: BaseNode {
                                        location: loc.get(1, 12, 1, 13),
                                        ..BaseNode::default()
                                    },
                                    value: 1
                                })),
                                comma: None,
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
                                separator: None,
                                value: Some(Expression::StringLit(StringLit {
                                    base: BaseNode {
                                        location: loc.get(1, 20, 1, 28),
                                        ..BaseNode::default()
                                    },
                                    value: "value2".to_string()
                                })),
                                comma: None,
                            }
                        ],
                        rbrace: None,
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
                        lbrack: None,
                        property: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(2, 6, 2, 10),
                                ..BaseNode::default()
                            },
                            name: "key1".to_string()
                        }),
                        rbrack: None,
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
                        lbrack: None,
                        property: PropertyKey::StringLit(StringLit {
                            base: BaseNode {
                                location: loc.get(3, 6, 3, 12),
                                ..BaseNode::default()
                            },
                            value: "key2".to_string()
                        }),
                        rbrack: None,
                    }))
                }))
            ],
            eof: None,
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
                    lbrace: None,
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
                        separator: None,
                        value: Some(Expression::Integer(IntegerLit {
                            base: BaseNode {
                                location: loc.get(1, 11, 1, 13),
                                ..BaseNode::default()
                            },
                            value: 10
                        })),
                        comma: None,
                    }],
                    rbrace: None,
                }))
            }))],
            eof: None,
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
                    lbrace: None,
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
                            separator: None,
                            value: Some(Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 11, 1, 13),
                                    ..BaseNode::default()
                                },
                                value: 10
                            })),
                            comma: None,
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
                            separator: None,
                            value: Some(Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 18, 1, 20),
                                    ..BaseNode::default()
                                },
                                value: 11
                            })),
                            comma: None,
                        }
                    ],
                    rbrace: None,
                }))
            }))],
            eof: None,
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
                    lbrace: None,
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
                            separator: None,
                            value: None,
                            comma: None,
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
                            separator: None,
                            value: None,
                            comma: None,
                        }
                    ],
                    rbrace: None,
                }))
            }))],
            eof: None,
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
                    lbrace: None,
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
                            separator: None,
                            value: None,
                            comma: None,
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
                            separator: None,
                            value: None,
                            comma: None,
                        }
                    ],
                    rbrace: None,
                }))
            }))],
            eof: None,
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
                    lbrace: None,
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
                            separator: None,
                            value: None,
                            comma: None,
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
                            separator: None,
                            value: Some(Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 11, 1, 12),
                                    ..BaseNode::default()
                                },
                                name: "c".to_string()
                            })),
                            comma: None,
                        }
                    ],
                    rbrace: None,
                }))
            }))],
            eof: None,
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
                    lbrace: None,
                    with: Some(WithSource {
                        source: Identifier {
                            base: BaseNode {
                                location: loc.get(1, 2, 1, 3),
                                ..BaseNode::default()
                            },
                            name: "a".to_string()
                        },
                        with: None
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
                            separator: None,
                            value: Some(Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 11, 1, 12),
                                    ..BaseNode::default()
                                },
                                name: "c".to_string()
                            })),
                            comma: None,
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
                            separator: None,
                            value: Some(Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 16, 1, 17),
                                    ..BaseNode::default()
                                },
                                name: "e".to_string()
                            })),
                            comma: None,
                        }
                    ],
                    rbrace: None,
                }))
            }))],
            eof: None,
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
                    lbrace: None,
                    with: Some(WithSource {
                        source: Identifier {
                            base: BaseNode {
                                location: loc.get(1, 2, 1, 3),
                                ..BaseNode::default()
                            },
                            name: "a".to_string()
                        },
                        with: None,
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
                            separator: None,
                            value: None,
                            comma: None,
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
                            separator: None,
                            value: None,
                            comma: None,
                        }
                    ],
                    rbrace: None,
                }))
            }))],
            eof: None,
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
                    lbrack: None,
                    index: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 3, 1, 4),
                            ..BaseNode::default()
                        },
                        value: 3
                    }),
                    rbrack: None,
                }))
            }))],
            eof: None,
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
                        lbrack: None,
                        index: Expression::Integer(IntegerLit {
                            base: BaseNode {
                                location: loc.get(1, 3, 1, 4),
                                ..BaseNode::default()
                            },
                            value: 3
                        }),
                        rbrack: None,
                    })),
                    lbrack: None,
                    index: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 6, 1, 7),
                            ..BaseNode::default()
                        },
                        value: 5
                    }),
                    rbrack: None,
                }))
            }))],
            eof: None,
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
                        lparen: None,
                        arguments: vec![],
                        rparen: None,
                    })),
                    lbrack: None,
                    index: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 6),
                            ..BaseNode::default()
                        },
                        value: 3
                    }),
                    rbrack: None,
                }))
            }))],
            eof: None,
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
                        lbrack: None,
                        property: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 3, 1, 4),
                                ..BaseNode::default()
                            },
                            name: "b".to_string()
                        }),
                        rbrack: None,
                    })),
                    lbrack: None,
                    property: PropertyKey::StringLit(StringLit {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 8),
                            ..BaseNode::default()
                        },
                        value: "c".to_string()
                    }),
                    rbrack: None,
                }))
            }))],
            eof: None,
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
                            lbrack: None,
                            property: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 3, 1, 4),
                                    ..BaseNode::default()
                                },
                                name: "b".to_string()
                            }),
                            rbrack: None,
                        })),
                        lparen: None,
                        arguments: vec![],
                        rparen: None,
                    })),
                    lbrack: None,
                    property: PropertyKey::StringLit(StringLit {
                        base: BaseNode {
                            location: loc.get(1, 7, 1, 10),
                            ..BaseNode::default()
                        },
                        value: "c".to_string()
                    }),
                    rbrack: None,
                }))
            }))],
            eof: None,
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
                    lbrack: None,
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
                        lparen: None,
                        arguments: vec![],
                        rparen: None,
                    })),
                    rbrack: None,
                }))
            }))],
            eof: None,
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
                    lbrack: None,
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
                        lparen: None,
                        arguments: vec![],
                        rparen: None,
                    })),
                    rbrack: None,
                }))
            }))],
            eof: None,
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
                    lbrack: None,
                    index: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 3, 1, 4),
                            ..BaseNode::default()
                        },
                        name: "b".to_string()
                    }),
                    rbrack: None,
                }))
            }))],
            eof: None,
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
            eof: None,
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
                        lbrack: None,
                        property: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 3, 1, 9),
                                ..BaseNode::default()
                            },
                            name: "_value".to_string()
                        }),
                        rbrack: None,
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
            eof: None,
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
            eof: None,
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
            eof: None,
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
            eof: None,
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
            eof: None,
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
                        lbrack: None,
                        property: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 7, 1, 8),
                                ..BaseNode::default()
                            },
                            name: "b".to_string()
                        }),
                        rbrack: None,
                    }))
                }))
            }))],
            eof: None,
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
                            comments: Some(Box::new(Comment {
                                lit: "// define a\n".to_string(),
                                next: None,
                            })),
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
                            comments: Some(Box::new(Comment {
                                lit: "// or this\n".to_string(),
                                next: None,
                            })),
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
                                        comments: Some(Box::new(Comment {
                                            lit: "// eval this\n".to_string(),
                                            next: None,
                                        })),
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
            eof: None,
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
                        lparen: None,
                        arguments: vec![],
                        rparen: None,
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
            eof: None,
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
                            lparen: None,
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
                                    lparen: None,
                                    callee: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(2, 18, 2, 19),
                                            ..BaseNode::default()
                                        },
                                        name: "f".to_string()
                                    }),
                                    arguments: vec![],
                                    rparen: None,
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
                            rparen: None,
                        }))
                    })),
                    right: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(2, 37, 2, 43),
                            ..BaseNode::default()
                        },
                        lparen: None,
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(2, 37, 2, 41),
                                ..BaseNode::default()
                            },
                            name: "fail".to_string()
                        }),
                        arguments: vec![],
                        rparen: None,
                    })),
                }))
            }))],
            eof: None,
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
                    lparen: None,
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
                                lparen: None,
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
                                        lparen: None,
                                        callee: Expression::Identifier(Identifier {
                                            base: BaseNode {
                                                location: loc.get(2, 19, 2, 20),
                                                ..BaseNode::default()
                                            },
                                            name: "f".to_string()
                                        }),
                                        arguments: vec![],
                                        rparen: None,
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
                                rparen: None,
                            }))
                        })),
                        right: Expression::Call(Box::new(CallExpr {
                            base: BaseNode {
                                location: loc.get(2, 38, 2, 44),
                                ..BaseNode::default()
                            },
                            lparen: None,
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(2, 38, 2, 42),
                                    ..BaseNode::default()
                                },
                                name: "fail".to_string()
                            }),
                            arguments: vec![],
                            rparen: None,
                        })),
                    })),
                    rparen: None,
                }))
            }))],
            eof: None,
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
            eof: None,
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
            eof: None,
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
            eof: None,
        },
    )
}

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
            eof: None,
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
            eof: None,
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
            eof: None,
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
                        lparen: None,
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
                        rparen: None,
                    }))
                }))
            }))],
            eof: None,
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
            eof: None,
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
                        lparen: None,
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
                        rparen: None,
                    }))
                }))
            }))],
            eof: None,
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
            eof: None,
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
                                                lparen: None,
                                                arguments: vec![],
                                                rparen: None,
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
                                                    lbrack: None,
                                                    property: PropertyKey::Identifier(Identifier {
                                                        base: BaseNode {
                                                            location: loc.get(1, 10, 1, 11),
                                                            ..BaseNode::default()
                                                        },
                                                        name: "a".to_string()
                                                    }),
                                                    rbrack: None,
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
                                                            lbrack: None,
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
                                                            rbrack: None,
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
                                            lbrack: None,
                                            index: Expression::Identifier(Identifier {
                                                base: BaseNode {
                                                    location: loc.get(1, 39, 1, 40),
                                                    ..BaseNode::default()
                                                },
                                                name: "g".to_string()
                                            }),
                                            rbrack: None,
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
                                lparen: None,
                                arguments: vec![],
                                rparen: None,
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
                                lparen: None,
                                arguments: vec![],
                                rparen: None,
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
            eof: None,
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
            eof: None,
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
            eof: None,
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
            eof: None,
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
            eof: None,
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
            eof: None,
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
            eof: None,
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
                        lparen: None,
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
                        rparen: None,
                    }))
                }))
            }))],
            eof: None,
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
                        lparen: None,
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
                        rparen: None,
                    }))
                }))
            }))],
            eof: None,
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
                        lparen: None,
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
                        rparen: None,
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
            eof: None,
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
                        lparen: None,
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
                        rparen: None,
                    }))
                }))
            }))],
            eof: None,
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
                                lparen: None,
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
                                rparen: None,
                            })),
                            lparen: None,
                            arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                base: BaseNode {
                                    location: loc.get(2, 2, 2, 8),
                                    ..BaseNode::default()
                                },
                                lbrace: None,
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
                                        separator: None,
                                        value: None,
                                        comma: None,
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
                                        separator: None,
                                        value: None,
                                        comma: None,
                                    }
                                ],
                                rbrace: None,
                            }))],
                            rparen: None,
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
            eof: None,
        },
    )
}

#[test]
fn arrow_function_called() {
    let mut p = Parser::new(
        r#"plusOne = (r) => r + 1
   plusOne(r:5)"#,
    );
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
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
                        lparen: None,
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
                            separator: None,
                            value: None,
                            comma: None,
                        }],
                        rparen: None,
                        arrow: None,
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
                        lparen: None,
                        arguments: vec![Expression::Object(Box::new(ObjectExpr {
                            base: BaseNode {
                                location: loc.get(2, 12, 2, 15),
                                ..BaseNode::default()
                            },
                            lbrace: None,
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
                                separator: None,
                                value: Some(Expression::Integer(IntegerLit {
                                    base: BaseNode {
                                        location: loc.get(2, 14, 2, 15),
                                        ..BaseNode::default()
                                    },
                                    value: 5
                                })),
                                comma: None,
                            }],
                            rbrace: None,
                        }))],
                        rparen: None,
                    }))
                }))
            ],
            eof: None,
        },
    )
}

#[test]
fn arrow_function_return_map() {
    let mut p = Parser::new(r#"toMap = (r) =>({r:r})"#);
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
                    lparen: None,
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
                        separator: None,
                        value: None,
                        comma: None,
                    }],
                    rparen: None,
                    arrow: None,
                    body: FunctionBody::Expr(Expression::Paren(Box::new(ParenExpr {
                        base: BaseNode {
                            location: loc.get(1, 15, 1, 22),
                            ..BaseNode::default()
                        },
                        lparen: None,
                        expression: Expression::Object(Box::new(ObjectExpr {
                            base: BaseNode {
                                location: loc.get(1, 16, 1, 21),
                                ..BaseNode::default()
                            },
                            lbrace: None,
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
                                separator: None,
                                value: Some(Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 19, 1, 20),
                                        ..BaseNode::default()
                                    },
                                    name: "r".to_string()
                                })),
                                comma: None,
                            }],
                            rbrace: None,
                        })),
                        rparen: None,
                    }))),
                }))
            }))],
            eof: None,
        },
    )
}

#[test]
fn arrow_function_with_default_arg() {
    let mut p = Parser::new(r#"addN = (r, n=5) => r + n"#);
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
                    lparen: None,
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
                            separator: None,
                            value: None,
                            comma: None,
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
                            separator: None,
                            value: Some(Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 14, 1, 15),
                                    ..BaseNode::default()
                                },
                                value: 5
                            })),
                            comma: None,
                        }
                    ],
                    rparen: None,
                    arrow: None,
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
            eof: None,
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
    let loc = Locator::new(&p.source[..]);
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
                        lparen: None,
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
                            separator: None,
                            value: None,
                            comma: None,
                        }],
                        rparen: None,
                        arrow: None,
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
                                lparen: None,
                                arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                    base: BaseNode {
                                        location: loc.get(3, 21, 3, 24),
                                        ..BaseNode::default()
                                    },
                                    lbrace: None,
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
                                        separator: None,
                                        value: Some(Expression::Integer(IntegerLit {
                                            base: BaseNode {
                                                location: loc.get(3, 23, 3, 24),
                                                ..BaseNode::default()
                                            },
                                            value: 5
                                        })),
                                        comma: None,
                                    }],
                                    rbrace: None,
                                }))],
                                rparen: None,
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
                            lparen: None,
                            arguments: vec![],
                            rparen: None,
                        }))
                    }))
                }))
            ],
            eof: None,
        },
    )
}

#[test]
fn arrow_function_as_single_expression() {
    let mut p = Parser::new(r#"f = (r) => r["_measurement"] == "cpu""#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
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
                    lparen: None,
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
                        separator: None,
                        value: None,
                        comma: None,
                    }],
                    rparen: None,
                    arrow: None,
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
                            lbrack: None,
                            property: PropertyKey::StringLit(StringLit {
                                base: BaseNode {
                                    location: loc.get(1, 14, 1, 28),
                                    ..BaseNode::default()
                                },
                                value: "_measurement".to_string()
                            }),
                            rbrack: None,
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
            eof: None,
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
    let loc = Locator::new(&p.source[..]);
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
                    lparen: None,
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
                        separator: None,
                        value: None,
                        comma: None,
                    }],
                    rparen: None,
                    arrow: None,
                    body: FunctionBody::Block(Block {
                        base: BaseNode {
                            location: loc.get(1, 12, 4, 14),
                            ..BaseNode::default()
                        },
                        lbrace: None,
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
                                    lbrack: None,
                                    property: PropertyKey::StringLit(StringLit {
                                        base: BaseNode {
                                            location: loc.get(2, 23, 2, 37),
                                            ..BaseNode::default()
                                        },
                                        value: "_measurement".to_string()
                                    }),
                                    rbrack: None,
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
                        rbrace: None,
                    }),
                }))
            }))],
            eof: None,
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
                    tk_if: None,
                    tk_then: None,
                    tk_else: None,
                }))
            }))],
            eof: None,
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
                                lparen: None,
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
                                rparen: None,
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
                    tk_if: None,
                    tk_then: None,
                    tk_else: None,
                }))
            }))],
            eof: None,
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
                        tk_if: None,
                        tk_then: None,
                        tk_else: None,
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
                        tk_if: None,
                        tk_then: None,
                        tk_else: None,
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
                        tk_if: None,
                        tk_then: None,
                        tk_else: None,
                    })),
                    tk_if: None,
                    tk_then: None,
                    tk_else: None,
                }))
            }))],
            eof: None,
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
                        lbrack: None,
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
                            lparen: None,
                            arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 31),
                                    ..BaseNode::default()
                                },
                                lbrace: None,
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
                                    separator: None,
                                    value: Some(Expression::StringLit(StringLit {
                                        base: BaseNode {
                                            location: loc.get(1, 13, 1, 31),
                                            ..BaseNode::default()
                                        },
                                        value: "telegraf/autogen".to_string()
                                    })),
                                    comma: None,
                                }],
                                rbrace: None,
                            }))],
                            rparen: None,
                        })),
                        rbrack: None,
                    })),
                    lparen: None,
                    arguments: vec![Expression::Object(Box::new(ObjectExpr {
                        base: BaseNode {
                            location: loc.get(1, 40, 1, 113),
                            ..BaseNode::default()
                        },
                        lbrace: None,
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
                            separator: None,
                            value: Some(Expression::Function(Box::new(FunctionExpr {
                                base: BaseNode {
                                    location: loc.get(1, 44, 1, 113),
                                    ..BaseNode::default()
                                },
                                lparen: None,
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
                                    separator: None,
                                    value: None,
                                    comma: None,
                                }],
                                rparen: None,
                                arrow: None,
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
                                                    lbrack: None,
                                                    property: PropertyKey::StringLit(StringLit {
                                                        base: BaseNode {
                                                            location: loc.get(1, 53, 1, 60),
                                                            ..BaseNode::default()
                                                        },
                                                        value: "other".to_string()
                                                    }),
                                                    rbrack: None,
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
                                                    lbrack: None,
                                                    property: PropertyKey::StringLit(StringLit {
                                                        base: BaseNode {
                                                            location: loc.get(1, 75, 1, 81),
                                                            ..BaseNode::default()
                                                        },
                                                        value: "this".to_string()
                                                    }),
                                                    rbrack: None,
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
                                                lbrack: None,
                                                property: PropertyKey::StringLit(StringLit {
                                                    base: BaseNode {
                                                        location: loc.get(1, 96, 1, 103),
                                                        ..BaseNode::default()
                                                    },
                                                    value: "these".to_string()
                                                }),
                                                rbrack: None,
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
                            comma: None,
                        }],
                        rbrace: None,
                    }))],
                    rparen: None,
                }))
            }))],
            eof: None,
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
                        lparen: None,
                        arguments: vec![Expression::Object(Box::new(ObjectExpr {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 31),
                                ..BaseNode::default()
                            },
                            lbrace: None,
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
                                separator: None,
                                value: Some(Expression::StringLit(StringLit {
                                    base: BaseNode {
                                        location: loc.get(1, 13, 1, 31),
                                        ..BaseNode::default()
                                    },
                                    value: "telegraf/autogen".to_string()
                                })),
                                comma: None,
                            }],
                            rbrace: None,
                        }))],
                        rparen: None,
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
                        lparen: None,
                        arguments: vec![Expression::Object(Box::new(ObjectExpr {
                            base: BaseNode {
                                location: loc.get(1, 40, 1, 58),
                                ..BaseNode::default()
                            },
                            lbrace: None,
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
                                    separator: None,
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
                                    comma: None,
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
                                    separator: None,
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
                                    comma: None,
                                }
                            ],
                            rbrace: None,
                        }))],
                        rparen: None,
                    }
                }))
            }))],
            eof: None,
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
                        lparen: None,
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 5),
                                ..BaseNode::default()
                            },
                            name: "from".to_string()
                        }),
                        rparen: None,
                        arguments: vec![Expression::Object(Box::new(ObjectExpr {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 31),
                                ..BaseNode::default()
                            },
                            lbrace: None,
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
                                separator: None,
                                value: Some(Expression::StringLit(StringLit {
                                    base: BaseNode {
                                        location: loc.get(1, 13, 1, 31),
                                        ..BaseNode::default()
                                    },
                                    value: "telegraf/autogen".to_string()
                                })),
                                comma: None,
                            }],
                            rbrace: None,
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
                        lparen: None,
                        arguments: vec![Expression::Object(Box::new(ObjectExpr {
                            base: BaseNode {
                                location: loc.get(1, 40, 1, 60),
                                ..BaseNode::default()
                            },
                            lbrace: None,
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
                                    separator: None,
                                    value: Some(Expression::Integer(IntegerLit {
                                        base: BaseNode {
                                            location: loc.get(1, 46, 1, 49),
                                            ..BaseNode::default()
                                        },
                                        value: 100
                                    })),
                                    comma: None,
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
                                    separator: None,
                                    value: Some(Expression::Integer(IntegerLit {
                                        base: BaseNode {
                                            location: loc.get(1, 58, 1, 60),
                                            ..BaseNode::default()
                                        },
                                        value: 10
                                    })),
                                    comma: None,
                                }
                            ],
                            rbrace: None,
                        }))],
                        rparen: None,
                    }
                }))
            }))],
            eof: None,
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
                            lparen: None,
                            arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 27),
                                    ..BaseNode::default()
                                },
                                lbrace: None,
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
                                    separator: None,
                                    value: Some(Expression::StringLit(StringLit {
                                        base: BaseNode {
                                            location: loc.get(1, 13, 1, 27),
                                            ..BaseNode::default()
                                        },
                                        value: "mydb/autogen".to_string()
                                    })),
                                    comma: None,
                                }],
                                rbrace: None,
                            }))],
                            rparen: None,
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
                            lparen: None,
                            arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                base: BaseNode {
                                    location: loc.get(2, 16, 2, 35),
                                    ..BaseNode::default()
                                },
                                lbrace: None,
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
                                        separator: None,
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
                                        comma: None,
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
                                        separator: None,
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
                                        comma: None,
                                    }
                                ],
                                rbrace: None,
                            }))],
                            rparen: None,
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
                        lparen: None,
                        arguments: vec![],
                        rparen: None,
                    }
                }))
            }))],
            eof: None,
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
                                lparen: None,
                                arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                    base: BaseNode {
                                        location: loc.get(1, 6, 1, 27),
                                        ..BaseNode::default()
                                    },
                                    lbrace: None,
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
                                        separator: None,
                                        value: Some(Expression::StringLit(StringLit {
                                            base: BaseNode {
                                                location: loc.get(1, 13, 1, 27),
                                                ..BaseNode::default()
                                            },
                                            value: "mydb/autogen".to_string()
                                        })),
                                        comma: None,
                                    }],
                                    rbrace: None,
                                }))],
                                rparen: None,
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
                                lparen: None,
                                arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                    base: BaseNode {
                                        location: loc.get(2, 16, 2, 35),
                                        ..BaseNode::default()
                                    },
                                    lbrace: None,
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
                                            separator: None,
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
                                            comma: None,
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
                                            separator: None,
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
                                            comma: None,
                                        }
                                    ],
                                    rbrace: None,
                                }))],
                                rparen: None,
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
                            lparen: None,
                            arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                base: BaseNode {
                                    location: loc.get(3, 16, 3, 20),
                                    ..BaseNode::default()
                                },
                                lbrace: None,
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
                                    separator: None,
                                    value: Some(Expression::Integer(IntegerLit {
                                        base: BaseNode {
                                            location: loc.get(3, 18, 3, 20),
                                            ..BaseNode::default()
                                        },
                                        value: 10
                                    })),
                                    comma: None,
                                }],
                                rbrace: None,
                            }))],
                            rparen: None,
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
                        lparen: None,
                        arguments: vec![],
                        rparen: None,
                    }
                }))
            }))],
            eof: None,
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
                            lparen: None,
                            arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                base: BaseNode {
                                    location: loc.get(2, 10, 2, 30),
                                    ..BaseNode::default()
                                },
                                lbrace: None,
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
                                    separator: None,
                                    value: Some(Expression::StringLit(StringLit {
                                        base: BaseNode {
                                            location: loc.get(2, 17, 2, 30),
                                            ..BaseNode::default()
                                        },
                                        value: "dbA/autogen".to_string()
                                    })),
                                    comma: None,
                                }],
                                rbrace: None,
                            }))],
                            rparen: None,
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
                            lparen: None,
                            arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                base: BaseNode {
                                    location: loc.get(2, 41, 2, 50),
                                    ..BaseNode::default()
                                },
                                lbrace: None,
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
                                    separator: None,
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
                                    comma: None,
                                }],
                                rbrace: None,
                            }))],
                            rparen: None,
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
                            lparen: None,
                            arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                base: BaseNode {
                                    location: loc.get(3, 10, 3, 30),
                                    ..BaseNode::default()
                                },
                                lbrace: None,
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
                                    separator: None,
                                    value: Some(Expression::StringLit(StringLit {
                                        base: BaseNode {
                                            location: loc.get(3, 17, 3, 30),
                                            ..BaseNode::default()
                                        },
                                        value: "dbB/autogen".to_string()
                                    })),
                                    comma: None,
                                }],
                                rbrace: None,
                            }))],
                            rparen: None,
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
                            lparen: None,
                            arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                base: BaseNode {
                                    location: loc.get(3, 41, 3, 50),
                                    ..BaseNode::default()
                                },
                                lbrace: None,
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
                                    separator: None,
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
                                    comma: None,
                                }],
                                rbrace: None,
                            }))],
                            rparen: None,
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
                        lparen: None,
                        arguments: vec![Expression::Object(Box::new(ObjectExpr {
                            base: BaseNode {
                                location: loc.get(4, 6, 4, 71),
                                ..BaseNode::default()
                            },
                            lbrace: None,
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
                                    separator: None,
                                    value: Some(Expression::Array(Box::new(ArrayExpr {
                                        base: BaseNode {
                                            location: loc.get(4, 13, 4, 18),
                                            ..BaseNode::default()
                                        },
                                        lbrack: None,
                                        elements: vec![
                                            ArrayItem {
                                                expression: Expression::Identifier(Identifier {
                                                    base: BaseNode {
                                                        location: loc.get(4, 14, 4, 15),
                                                        ..BaseNode::default()
                                                    },
                                                    name: "a".to_string()
                                                }),
                                                comma: None,
                                            },
                                            ArrayItem {
                                                expression: Expression::Identifier(Identifier {
                                                    base: BaseNode {
                                                        location: loc.get(4, 16, 4, 17),
                                                        ..BaseNode::default()
                                                    },
                                                    name: "b".to_string()
                                                }),
                                                comma: None,
                                            }
                                        ],
                                        rbrack: None,
                                    }))),
                                    comma: None,
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
                                    separator: None,
                                    value: Some(Expression::Array(Box::new(ArrayExpr {
                                        base: BaseNode {
                                            location: loc.get(4, 23, 4, 31),
                                            ..BaseNode::default()
                                        },
                                        lbrack: None,
                                        elements: vec![ArrayItem {
                                            expression: Expression::StringLit(StringLit {
                                                base: BaseNode {
                                                    location: loc.get(4, 24, 4, 30),
                                                    ..BaseNode::default()
                                                },
                                                value: "host".to_string()
                                            }),
                                            comma: None,
                                        }],
                                        rbrack: None,
                                    }))),
                                    comma: None,
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
                                    separator: None,
                                    value: Some(Expression::Function(Box::new(FunctionExpr {
                                        base: BaseNode {
                                            location: loc.get(4, 37, 4, 71),
                                            ..BaseNode::default()
                                        },
                                        lparen: None,
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
                                                separator: None,
                                                value: None,
                                                comma: None,
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
                                                separator: None,
                                                value: None,
                                                comma: None,
                                            }
                                        ],
                                        rparen: None,
                                        arrow: None,
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
                                                    lbrack: None,
                                                    property: PropertyKey::StringLit(StringLit {
                                                        base: BaseNode {
                                                            location: loc.get(4, 48, 4, 56),
                                                            ..BaseNode::default()
                                                        },
                                                        value: "_field".to_string()
                                                    }),
                                                    rbrack: None,
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
                                                    lbrack: None,
                                                    property: PropertyKey::StringLit(StringLit {
                                                        base: BaseNode {
                                                            location: loc.get(4, 62, 4, 70),
                                                            ..BaseNode::default()
                                                        },
                                                        value: "_field".to_string()
                                                    }),
                                                    rbrack: None,
                                                }))
                                            }
                                        ))),
                                    }))),
                                    comma: None,
                                }
                            ],
                            rbrace: None,
                        }))],
                        rparen: None,
                    }))
                }))
            ],
            eof: None,
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
    assert_eq!(
        parsed,
        File {
            base: BaseNode {
                location: loc.get(2, 1, 10, 86),
                .. BaseNode::default()
            },
            name: "".to_string(),
            metadata: "parser-type=rust".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(2, 1, 4, 21),
                        .. BaseNode::default()
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(2, 1, 2, 2),
                            .. BaseNode::default()
                        },
                        name: "a".to_string()
                    },
                    init: Expression::PipeExpr(Box::new(PipeExpr {
                        base: BaseNode {
                            location: loc.get(2, 5, 4, 21),
                            .. BaseNode::default()
                        },
                        argument: Expression::PipeExpr(Box::new(PipeExpr {
                            base: BaseNode {
                                location: loc.get(2, 5, 3, 48),
                                .. BaseNode::default()
                            },
                            argument: Expression::Call(Box::new(CallExpr {
                                base: BaseNode {
                                    location: loc.get(2, 5, 2, 32),
                                    .. BaseNode::default()
                                },
                                callee: Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(2, 5, 2, 9),
                                        .. BaseNode::default()
                                    },
                                    name: "from".to_string()
                                }),
                                lparen: None,
                                arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                    base: BaseNode {
                                        location: loc.get(2, 10, 2, 31),
                                        .. BaseNode::default()
                                    },
                                    lbrace: None,
                                    with: None,
                                    properties: vec![Property {
                                        base: BaseNode {
                                            location: loc.get(2, 10, 2, 31),
                                            .. BaseNode::default()
                                        },
                                        key: PropertyKey::Identifier(Identifier {
                                            base: BaseNode {
                                                location: loc.get(2, 10, 2, 16),
                                                .. BaseNode::default()
                                            },
                                            name: "bucket".to_string()
                                        }),
                                        separator: None,
                                        value: Some(Expression::StringLit(StringLit {
                                            base: BaseNode {
                                                location: loc.get(2, 17, 2, 31),
                                                .. BaseNode::default()
                                            },
                                            value: "Flux/autogen".to_string()
                                        })),
                                        comma: None,
                                    }],
                                    rbrace: None,
                                }))],
                                rparen: None,
                            })),
                            call: CallExpr {
                                base: BaseNode {
                                    location: loc.get(3, 5, 3, 48),
                                    .. BaseNode::default()
                                },
                                callee: Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(3, 5, 3, 11),
                                        .. BaseNode::default()
                                    },
                                    name: "filter".to_string()
                                }),
                                lparen: None,
                                arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                    base: BaseNode {
                                        location: loc.get(3, 12, 3, 47),
                                        .. BaseNode::default()
                                    },
                                    lbrace: None,
                                    with: None,
                                    properties: vec![Property {
                                        base: BaseNode {
                                            location: loc.get(3, 12, 3, 47),
                                            .. BaseNode::default()
                                        },
                                        key: PropertyKey::Identifier(Identifier {
                                            base: BaseNode {
                                                location: loc.get(3, 12, 3, 14),
                                                .. BaseNode::default()
                                            },
                                            name: "fn".to_string()
                                        }),
                                        separator: None,
                                        value: Some(Expression::Function(Box::new(FunctionExpr {
                                            base: BaseNode {
                                                location: loc.get(3, 16, 3, 47),
                                                .. BaseNode::default()
                                            },
                                            lparen: None,
                                            params: vec![Property {
                                                base: BaseNode {
                                                    location: loc.get(3, 17, 3, 18),
                                                    .. BaseNode::default()
                                                },
                                                key: PropertyKey::Identifier(Identifier {
                                                    base: BaseNode {
                                                        location: loc.get(3, 17, 3, 18),
                                                        .. BaseNode::default()
                                                    },
                                                    name: "r".to_string()
                                                }),
                                                separator: None,
                                                value: None,
                                                comma: None,
                                            }],
                                            rparen: None,
                                            arrow: None,
                                            body: FunctionBody::Expr(Expression::Binary(Box::new(
                                                BinaryExpr {
                                                    base: BaseNode {
                                                        location: loc.get(3, 23, 3, 47),
                                                        .. BaseNode::default()
                                                    },
                                                    operator: Operator::EqualOperator,
                                                    left: Expression::Member(Box::new(
                                                        MemberExpr {
                                                            base: BaseNode {
                                                                location: loc.get(3, 23, 3, 40),
                                                                .. BaseNode::default()
                                                            },
                                                            object: Expression::Identifier(
                                                                Identifier {
                                                                    base: BaseNode {
                                                                        location: loc
                                                                            .get(3, 23, 3, 24),
                                                                        .. BaseNode::default()
                                                                    },
                                                                    name: "r".to_string()
                                                                }
                                                            ),
                                                            lbrack: None,
                                                            property: PropertyKey::StringLit(
                                                                StringLit {
                                                                    base: BaseNode {
                                                                        location: loc
                                                                            .get(3, 25, 3, 39),
                                                                        .. BaseNode::default()
                                                                    },
                                                                    value: "_measurement"
                                                                        .to_string()
                                                                }
                                                            ),
                                                            rbrack: None,
                                                        }
                                                    )),
                                                    right: Expression::StringLit(StringLit {
                                                        base: BaseNode {
                                                            location: loc.get(3, 44, 3, 47),
                                                            .. BaseNode::default()
                                                        },
                                                        value: "a".to_string()
                                                    })
                                                }
                                            ))),
                                        }))),
                                        comma: None,
                                    }],
                                    rbrace: None,
                                }))],
                                rparen: None,
                            }
                        })),
                        call: CallExpr {
                            base: BaseNode {
                                location: loc.get(4, 5, 4, 21),
                                .. BaseNode::default()
                            },
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(4, 5, 4, 10),
                                    .. BaseNode::default()
                                },
                                name: "range".to_string()
                            }),
                            lparen: None,
                            arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                base: BaseNode {
                                    location: loc.get(4, 11, 4, 20),
                                    .. BaseNode::default()
                                },
                                lbrace: None,
                                with: None,
                                properties: vec![Property {
                                    base: BaseNode {
                                        location: loc.get(4, 11, 4, 20),
                                        .. BaseNode::default()
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(4, 11, 4, 16),
                                            .. BaseNode::default()
                                        },
                                        name: "start".to_string()
                                    }),
                                    separator: None,
                                    value: Some(Expression::Unary(Box::new(UnaryExpr {
                                        base: BaseNode {
                                            location: loc.get(4, 17, 4, 20),
                                            .. BaseNode::default()
                                        },
                                        operator: Operator::SubtractionOperator,
                                        argument: Expression::Duration(DurationLit {
                                            base: BaseNode {
                                                location: loc.get(4, 18, 4, 20),
                                                .. BaseNode::default()
                                            },
                                            values: vec![Duration {
                                                magnitude: 1,
                                                unit: "h".to_string()
                                            }]
                                        })
                                    }))),
                                    comma: None,
                                }],
                                rbrace: None,
                            }))],
                            rparen: None,
                        }
                    }))
                })),
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(6, 1, 8, 21),
                        .. BaseNode::default()
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(6, 1, 6, 2),
                            .. BaseNode::default()
                        },
                        name: "b".to_string()
                    },
                    init: Expression::PipeExpr(Box::new(PipeExpr {
                        base: BaseNode {
                            location: loc.get(6, 5, 8, 21),
                            .. BaseNode::default()
                        },
                        argument: Expression::PipeExpr(Box::new(PipeExpr {
                            base: BaseNode {
                                location: loc.get(6, 5, 7, 48),
                                .. BaseNode::default()
                            },
                            argument: Expression::Call(Box::new(CallExpr {
                                base: BaseNode {
                                    location: loc.get(6, 5, 6, 32),
                                    .. BaseNode::default()
                                },
                                callee: Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(6, 5, 6, 9),
                                        .. BaseNode::default()
                                    },
                                    name: "from".to_string()
                                }),
                                lparen: None,
                                arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                    base: BaseNode {
                                        location: loc.get(6, 10, 6, 31),
                                        .. BaseNode::default()
                                    },
                                    lbrace: None,
                                    with: None,
                                    properties: vec![Property {
                                        base: BaseNode {
                                            location: loc.get(6, 10, 6, 31),
                                            .. BaseNode::default()
                                        },
                                        key: PropertyKey::Identifier(Identifier {
                                            base: BaseNode {
                                                location: loc.get(6, 10, 6, 16),
                                                .. BaseNode::default()
                                            },
                                            name: "bucket".to_string()
                                        }),
                                        separator: None,
                                        value: Some(Expression::StringLit(StringLit {
                                            base: BaseNode {
                                                location: loc.get(6, 17, 6, 31),
                                                .. BaseNode::default()
                                            },
                                            value: "Flux/autogen".to_string()
                                        })),
                                        comma: None,
                                    }],
                                    rbrace: None,
                                }))],
                                rparen: None,
                            })),
                            call: CallExpr {
                                base: BaseNode {
                                    location: loc.get(7, 5, 7, 48),
                                    .. BaseNode::default()
                                },
                                callee: Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(7, 5, 7, 11),
                                        .. BaseNode::default()
                                    },
                                    name: "filter".to_string()
                                }),
                                lparen: None,
                                arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                    base: BaseNode {
                                        location: loc.get(7, 12, 7, 47),
                                        .. BaseNode::default()
                                    },
                                    lbrace: None,
                                    with: None,
                                    properties: vec![Property {
                                        base: BaseNode {
                                            location: loc.get(7, 12, 7, 47),
                                            .. BaseNode::default()
                                        },
                                        key: PropertyKey::Identifier(Identifier {
                                            base: BaseNode {
                                                location: loc.get(7, 12, 7, 14),
                                                .. BaseNode::default()
                                            },
                                            name: "fn".to_string()
                                        }),
                                        separator: None,
                                        value: Some(Expression::Function(Box::new(FunctionExpr {
                                            base: BaseNode {
                                                location: loc.get(7, 16, 7, 47),
                                                .. BaseNode::default()
                                            },
                                            lparen: None,
                                            params: vec![Property {
                                                base: BaseNode {
                                                    location: loc.get(7, 17, 7, 18),
                                                    .. BaseNode::default()
                                                },
                                                key: PropertyKey::Identifier(Identifier {
                                                    base: BaseNode {
                                                        location: loc.get(7, 17, 7, 18),
                                                        .. BaseNode::default()
                                                    },
                                                    name: "r".to_string()
                                                }),
                                                separator: None,
                                                value: None,
                                                comma: None,
                                            }],
                                            rparen: None,
                                            arrow: None,
                                            body: FunctionBody::Expr(Expression::Binary(Box::new(
                                                BinaryExpr {
                                                    base: BaseNode {
                                                        location: loc.get(7, 23, 7, 47),
                                                        .. BaseNode::default()
                                                    },
                                                    operator: Operator::EqualOperator,
                                                    left: Expression::Member(Box::new(
                                                        MemberExpr {
                                                            base: BaseNode {
                                                                location: loc.get(7, 23, 7, 40),
                                                                .. BaseNode::default()
                                                            },
                                                            object: Expression::Identifier(
                                                                Identifier {
                                                                    base: BaseNode {
                                                                        location: loc
                                                                            .get(7, 23, 7, 24),
                                                                        .. BaseNode::default()
                                                                    },
                                                                    name: "r".to_string()
                                                                }
                                                            ),
                                                            lbrack: None,
                                                            property: PropertyKey::StringLit(
                                                                StringLit {
                                                                    base: BaseNode {
                                                                        location: loc
                                                                            .get(7, 25, 7, 39),
                                                                        .. BaseNode::default()
                                                                    },
                                                                    value: "_measurement"
                                                                        .to_string()
                                                                }
                                                            ),
                                                            rbrack: None,
                                                        }
                                                    )),
                                                    right: Expression::StringLit(StringLit {
                                                        base: BaseNode {
                                                            location: loc.get(7, 44, 7, 47),
                                                            .. BaseNode::default()
                                                        },
                                                        value: "b".to_string()
                                                    })
                                                }
                                            ))),
                                        }))),
                                        comma: None,
                                    }],
                                    rbrace: None,
                                }))],
                                rparen: None,
                            }
                        })),
                        call: CallExpr {
                            base: BaseNode {
                                location: loc.get(8, 5, 8, 21),
                                .. BaseNode::default()
                            },
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(8, 5, 8, 10),
                                    .. BaseNode::default()
                                },
                                name: "range".to_string()
                            }),
                            lparen: None,
                            arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                base: BaseNode {
                                    location: loc.get(8, 11, 8, 20),
                                    .. BaseNode::default()
                                },
                                lbrace: None,
                                with: None,
                                properties: vec![Property {
                                    base: BaseNode {
                                        location: loc.get(8, 11, 8, 20),
                                        .. BaseNode::default()
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(8, 11, 8, 16),
                                            .. BaseNode::default()
                                        },
                                        name: "start".to_string()
                                    }),
                                    separator: None,
                                    value: Some(Expression::Unary(Box::new(UnaryExpr {
                                        base: BaseNode {
                                            location: loc.get(8, 17, 8, 20),
                                            .. BaseNode::default()
                                        },
                                        operator: Operator::SubtractionOperator,
                                        argument: Expression::Duration(DurationLit {
                                            base: BaseNode {
                                                location: loc.get(8, 18, 8, 20),
                                                .. BaseNode::default()
                                            },
                                            values: vec![Duration {
                                                magnitude: 1,
                                                unit: "h".to_string()
                                            }]
                                        })
                                    }))),
                                    comma: None,
                                }],
                                rbrace: None,
                            }))],
                            rparen: None,
                        }
                    }))
                })),
                Statement::Expr(Box::new(ExprStmt {
                    base: BaseNode {
                        location: loc.get(10, 1, 10, 86),
                        .. BaseNode::default()
                    },
                    expression: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(10, 1, 10, 86),
                            .. BaseNode::default()
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(10, 1, 10, 5),
                                .. BaseNode::default()
                            },
                            name: "join".to_string()
                        }),
                        lparen: None,
                        arguments: vec![Expression::Object(Box::new(ObjectExpr {
                            base: BaseNode {
                                location: loc.get(10, 6, 10, 85),
                                .. BaseNode::default()
                            },
                            lbrace: None,
                            with: None,
                            properties: vec![
                                Property {
                                    base: BaseNode {
                                        location: loc.get(10, 6, 10, 18),
                                        .. BaseNode::default()
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(10, 6, 10, 12),
                                            .. BaseNode::default()
                                        },
                                        name: "tables".to_string()
                                    }),
                                    separator: None,
                                    value: Some(Expression::Array(Box::new(ArrayExpr {
                                        base: BaseNode {
                                            location: loc.get(10, 13, 10, 18),
                                            .. BaseNode::default()
                                        },
                                        lbrack: None,
                                        elements: vec![
                                            ArrayItem {
                                                expression: Expression::Identifier(Identifier {
                                                    base: BaseNode {
                                                        location: loc.get(10, 14, 10, 15),
                                                        .. BaseNode::default()
                                                    },
                                                    name: "a".to_string()
                                                }),
                                                comma: None,
                                            },
                                            ArrayItem {
                                                expression: Expression::Identifier(Identifier {
                                                    base: BaseNode {
                                                        location: loc.get(10, 16, 10, 17),
                                                        .. BaseNode::default()
                                                    },
                                                    name: "b".to_string()
                                                }),
                                                comma: None,
                                            }
                                        ],
                                        rbrack: None,
                                    }))),
                                    comma: None,
                                },
                                Property {
                                    base: BaseNode {
                                        location: loc.get(10, 20, 10, 29),
                                        .. BaseNode::default()
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(10, 20, 10, 22),
                                            .. BaseNode::default()
                                        },
                                        name: "on".to_string()
                                    }),
                                    separator: None,
                                    value: Some(Expression::Array(Box::new(ArrayExpr {
                                        base: BaseNode {
                                            location: loc.get(10, 23, 10, 29),
                                            .. BaseNode::default()
                                        },
                                        lbrack: None,
                                        elements: vec![ArrayItem {
                                            expression: Expression::StringLit(StringLit {
                                                base: BaseNode {
                                                    location: loc.get(10, 24, 10, 28),
                                                    .. BaseNode::default()
                                                },
                                                value: "t1".to_string()
                                            }),
                                            comma: None,
                                        }],
                                        rbrack: None,
                                    }))),
                                    comma: None,
                                },
                                Property {
                                    base: BaseNode {
                                        location: loc.get(10, 31, 10, 85),
                                        .. BaseNode::default()
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(10, 31, 10, 33),
                                            .. BaseNode::default()
                                        },
                                        name: "fn".to_string()
                                    }),
                                    separator: None,
                                    value: Some(Expression::Function(Box::new(FunctionExpr {
                                        base: BaseNode {
                                            location: loc.get(10, 35, 10, 85),
                                            .. BaseNode::default()
                                        },
                                        lparen: None,
                                        params: vec![
                                            Property {
                                                base: BaseNode {
                                                    location: loc.get(10, 36, 10, 37),
                                                    .. BaseNode::default()
                                                },
                                                key: PropertyKey::Identifier(Identifier {
                                                    base: BaseNode {
                                                        location: loc.get(10, 36, 10, 37),
                                                        .. BaseNode::default()
                                                    },
                                                    name: "a".to_string()
                                                }),
                                                separator: None,
                                                value: None,
                                                comma: None,
                                            },
                                            Property {
                                                base: BaseNode {
                                                    location: loc.get(10, 38, 10, 39),
                                                    .. BaseNode::default()
                                                },
                                                key: PropertyKey::Identifier(Identifier {
                                                    base: BaseNode {
                                                        location: loc.get(10, 38, 10, 39),
                                                        .. BaseNode::default()
                                                    },
                                                    name: "b".to_string()
                                                }),
                                                separator: None,
                                                value: None,
                                                comma: None,
                                            }
                                        ],
                                        rparen: None,
                                        arrow: None,
                                        body: FunctionBody::Expr(Expression::Binary(Box::new(
                                            BinaryExpr {
                                                base: BaseNode {
                                                    location: loc.get(10, 44, 10, 85),
                                                    .. BaseNode::default()
                                                },
                                                operator: Operator::DivisionOperator,
                                                left: Expression::Paren(Box::new(ParenExpr {
                                                    base: BaseNode {
                                                        location: loc.get(10, 44, 10, 71),
                                                        .. BaseNode::default()
                                                    },
                                                    lparen: None,
                                                    expression: Expression::Binary(Box::new(
                                                        BinaryExpr {
                                                            base: BaseNode {
                                                                location: loc.get(10, 45, 10, 70),
                                                                .. BaseNode::default()
                                                            },
                                                            operator: Operator::SubtractionOperator,
                                                            left: Expression::Member(Box::new(
                                                                MemberExpr {
                                                                    base: BaseNode {
                                                                        location: loc
                                                                            .get(10, 45, 10, 56),
                                                                        .. BaseNode::default()
                                                                    },
                                                                    object: Expression::Identifier(
                                                                        Identifier {
                                                                            base: BaseNode {
                                                                                location: loc.get(
                                                                                    10, 45, 10, 46
                                                                                ),
                                                                                .. BaseNode::default()
                                                                            },
                                                                            name: "a".to_string()
                                                                        }
                                                                    ),
                                                                    lbrack: None,
                                                                    property:
                                                                        PropertyKey::StringLit(
                                                                            StringLit {
                                                                                base: BaseNode {
                                                                                    location: loc
                                                                                        .get(
                                                                                            10, 47,
                                                                                            10, 55
                                                                                        ),
                                                                                    .. BaseNode::default()
                                                                                },
                                                                                value: "_field"
                                                                                    .to_string()
                                                                            }
                                                                        ),
                                                                    rbrack: None,
                                                                }
                                                            )),
                                                            right: Expression::Member(Box::new(
                                                                MemberExpr {
                                                                    base: BaseNode {
                                                                        location: loc
                                                                            .get(10, 59, 10, 70),
                                                                        .. BaseNode::default()
                                                                    },
                                                                    object: Expression::Identifier(
                                                                        Identifier {
                                                                            base: BaseNode {
                                                                                location: loc.get(
                                                                                    10, 59, 10, 60
                                                                                ),
                                                                                .. BaseNode::default()
                                                                            },
                                                                            name: "b".to_string()
                                                                        }
                                                                    ),
                                                                    lbrack: None,
                                                                    property:
                                                                        PropertyKey::StringLit(
                                                                            StringLit {
                                                                                base: BaseNode {
                                                                                    location: loc
                                                                                        .get(
                                                                                            10, 61,
                                                                                            10, 69
                                                                                        ),
                                                                                    .. BaseNode::default()
                                                                                },
                                                                                value: "_field"
                                                                                    .to_string()
                                                                            }
                                                                        ),
                                                                    rbrack: None,
                                                                }
                                                            ))
                                                        }
                                                    )),
                                                    rparen: None,
                                                })),
                                                right: Expression::Member(Box::new(MemberExpr {
                                                    base: BaseNode {
                                                        location: loc.get(10, 74, 10, 85),
                                                        .. BaseNode::default()
                                                    },
                                                    object: Expression::Identifier(Identifier {
                                                        base: BaseNode {
                                                            location: loc.get(10, 74, 10, 75),
                                                            .. BaseNode::default()
                                                        },
                                                        name: "b".to_string()
                                                    }),
                                                    lbrack: None,
                                                    property: PropertyKey::StringLit(StringLit {
                                                        base: BaseNode {
                                                            location: loc.get(10, 76, 10, 84),
                                                            .. BaseNode::default()
                                                        },
                                                        value: "_field".to_string()
                                                    }),
                                                    rbrack: None,
                                                }))
                                            }
                                        ))),
                                    }))),
                                    comma: None,
                                }
                            ],
                            rbrace: None,
                        }))],
                        rparen: None,
                    }))
                }))
            ],
            eof: None,
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
            eof: None,
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
            eof: None,
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
            eof: None,
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
            eof: None,
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
            eof: None,
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
                        lbrace: None,
                        rbrace: None,
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
                            comma: None,
                            separator: None,
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
                    lparen: None,
                    rparen: None,
                }))
            }))],
            eof: None,
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
            eof: None,
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
            eof: None,
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
                            lparen: None,
                            arguments: vec![],
                            rparen: None,
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
                            lparen: None,
                            arguments: vec![],
                            rparen: None,
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
                        lparen: None,
                        arguments: vec![Expression::Object(Box::new(ObjectExpr {
                            base: BaseNode {
                                location: loc.get(1, 26, 1, 56),
                                ..BaseNode::default()
                            },
                            lbrace: None,
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
                                separator: None,
                                value: Some(Expression::Function(Box::new(FunctionExpr {
                                    base: BaseNode {
                                        location: loc.get(1, 30, 1, 56),
                                        ..BaseNode::default()
                                    },
                                    lparen: None,
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
                                        separator: None,
                                        value: None,
                                        comma: None,
                                    }],
                                    rparen: None,
                                    arrow: None,
                                    body: FunctionBody::Block(Block {
                                        base: BaseNode {
                                            location: loc.get(1, 37, 1, 56),
                                            errors: vec!["expected RBRACE, got RPAREN".to_string()],
                                            ..BaseNode::default()
                                        },
                                        lbrace: None,
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
                                                lbrack: None,
                                                property: PropertyKey::Identifier(Identifier {
                                                    base: BaseNode {
                                                        location: loc.get(1, 48, 1, 54),
                                                        ..BaseNode::default()
                                                    },
                                                    name: "_value".to_string()
                                                }),
                                                rbrack: None,
                                            }))
                                        }))],
                                        rbrace: None,
                                    }),
                                }))),
                                comma: None,
                            }],
                            rbrace: None,
                        }))],
                        rparen: None,
                    }
                }))
            }))],
            eof: None,
        },
    )
}

#[test]
fn string_with_utf_8() {
    let mut p = Parser::new(r#""日本語""#);
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
                expression: Expression::StringLit(StringLit {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 12),
                        ..BaseNode::default()
                    },
                    value: "日本語".to_string()
                })
            }))],
            eof: None,
        },
    )
}

#[test]
fn string_with_byte_values() {
    let mut p = Parser::new(r#""\xe6\x97\xa5\xe6\x9c\xac\xe8\xaa\x9e""#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
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
            eof: None,
        },
    )
}

#[test]
fn string_with_mixed_values() {
    let mut p = Parser::new(r#""hello 日x本 \xe6\x97\xa5\xe6\x9c\xac\xe8\xaa\x9e \xc2\xb5s""#);
    let parsed = p.parse_file("".to_string());
    let loc = Locator::new(&p.source[..]);
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
            eof: None,
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
    let loc = Locator::new(&p.source[..]);
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
            eof: None,
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
    let loc = Locator::new(&p.source[..]);
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
            eof: None,
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
            eof: None,
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
                    lparen: None,
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
                    rparen: None,
                }))
            }))],
            eof: None,
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
                    lparen: None,
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
                    rparen: None,
                }))
            }))],
            eof: None,
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
                    lparen: None,
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
                    rparen: None,
                }))
            }))],
            eof: None,
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
                    lparen: None,
                    expression: Expression::Bad(Box::new(BadExpr {
                        base: BaseNode {
                            location: loc.get(1, 2, 1, 3),
                            ..BaseNode::default()
                        },
                        text: "@".to_string(),
                        expression: None
                    })),
                    rparen: None,
                }))
            }))],
            eof: None,
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
                    lparen: None,
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
                            separator: None,
                            value: None,
                            comma: None,
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
                            separator: None,
                            value: None,
                            comma: None,
                        }
                    ],
                    rparen: None,
                    arrow: None,
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
            eof: None,
        },
    )
}

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
                    lbrace: None,
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
                            separator: None,
                            value: Some(Expression::StringLit(StringLit {
                                base: BaseNode {
                                    location: loc.get(1, 9, 1, 12),
                                    ..BaseNode::default()
                                },
                                value: "a".to_string()
                            })),
                            comma: None,
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
                            separator: None,
                            value: None,
                            comma: None,
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
                            separator: None,
                            value: Some(Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 18, 1, 19),
                                    ..BaseNode::default()
                                },
                                value: 7
                            })),
                            comma: None,
                        }
                    ],
                    rbrace: None,
                }))
            }))],
            eof: None,
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
                    lbrace: None,
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
                        separator: None,
                        value: Some(Expression::StringLit(StringLit {
                            base: BaseNode {
                                location: loc.get(1, 8, 1, 11),
                                ..BaseNode::default()
                            },
                            value: "a".to_string()
                        })),
                        comma: None,
                    }],
                    rbrace: None,
                }))
            }))],
            eof: None,
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
                    lbrace: None,
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
                        separator: None,
                        value: None,
                        comma: None,
                    }],
                    rbrace: None,
                }))
            }))],
            eof: None,
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
                    lbrace: None,
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
                            separator: None,
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
                            comma: None,
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
                            separator: None,
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
                            comma: None,
                        }
                    ],
                    rbrace: None,
                }))
            }))],
            eof: None,
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
                    lbrace: None,
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
                        separator: None,
                        value: Some(Expression::StringLit(StringLit {
                            base: BaseNode {
                                location: loc.get(1, 9, 1, 12),
                                ..BaseNode::default()
                            },
                            value: "a".to_string()
                        })),
                        comma: None,
                    }],
                    rbrace: None,
                }))
            }))],
            eof: None,
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
                    lbrace: None,
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
                            separator: None,
                            value: Some(Expression::StringLit(StringLit {
                                base: BaseNode {
                                    location: loc.get(1, 9, 1, 12),
                                    ..BaseNode::default()
                                },
                                value: "a".to_string()
                            })),
                            comma: None,
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
                            separator: None,
                            value: None,
                            comma: None,
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
                            separator: None,
                            value: Some(Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 21, 1, 22),
                                    ..BaseNode::default()
                                },
                                value: 7
                            })),
                            comma: None,
                        }
                    ],
                    rbrace: None,
                }))
            }))],
            eof: None,
        },
    )
}

// TODO(jsternberg): This should fill in error nodes.
// The current behavior is non-sensical.
#[test]
fn invalid_expression_in_array() {
    let mut p = Parser::new(r#"['a']"#);
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
                expression: Expression::Array(Box::new(ArrayExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 6),
                        ..BaseNode::default()
                    },
                    lbrack: None,
                    elements: vec![ArrayItem {
                        expression: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 3, 1, 4),
                                ..BaseNode::default()
                            },
                            name: "a".to_string()
                        }),
                        comma: None,
                    }],
                    rbrack: None,
                }))
            }))],
            eof: None,
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
            eof: None,
        },
    )
}

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
        "expected STRING, got EOF",
        "invalid string literal",
    ];

    let mut p = Parser::new(r#"""#);
    let result = p.parse_string_literal();
    assert_eq!("".to_string(), result.value);
    assert_eq!(errors, result.base.errors);
}
