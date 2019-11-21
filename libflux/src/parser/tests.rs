use super::*;
use crate::ast;

use chrono;

// This gives us a colorful diff.
#[cfg(test)]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 19),
                    errors: vec![]
                },
                expression: Expression::StringExpr(Box::new(StringExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 19),
                        errors: vec![]
                    },
                    parts: vec![
                        StringExprPart::Text(TextPart {
                            base: BaseNode {
                                location: loc.get(1, 2, 1, 10),
                                errors: vec![]
                            },
                            value: "a + b = ".to_string(),
                        }),
                        StringExprPart::Interpolated(InterpolatedPart {
                            base: BaseNode {
                                location: loc.get(1, 10, 1, 18),
                                errors: vec![]
                            },
                            expression: Expression::Binary(Box::new(BinaryExpr {
                                base: BaseNode {
                                    location: loc.get(1, 12, 1, 17),
                                    errors: vec![]
                                },
                                left: Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 12, 1, 13),
                                        errors: vec![]
                                    },
                                    name: "a".to_string(),
                                }),
                                right: Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 16, 1, 17),
                                        errors: vec![]
                                    },
                                    name: "b".to_string(),
                                }),
                                operator: Operator::AdditionOperator,
                            })),
                        }),
                    ],
                })),
            }),],
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 24),
                    errors: vec![]
                },
                expression: Expression::StringExpr(Box::new(StringExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 24),
                        errors: vec![]
                    },
                    parts: vec![
                        StringExprPart::Text(TextPart {
                            base: BaseNode {
                                location: loc.get(1, 2, 1, 6),
                                errors: vec![]
                            },
                            value: "a = ".to_string(),
                        }),
                        StringExprPart::Interpolated(InterpolatedPart {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 10),
                                errors: vec![]
                            },
                            expression: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 8, 1, 9),
                                    errors: vec![]
                                },
                                name: "a".to_string(),
                            }),
                        }),
                        StringExprPart::Text(TextPart {
                            base: BaseNode {
                                location: loc.get(1, 10, 1, 19),
                                errors: vec![]
                            },
                            value: " and b = ".to_string(),
                        }),
                        StringExprPart::Interpolated(InterpolatedPart {
                            base: BaseNode {
                                location: loc.get(1, 19, 1, 23),
                                errors: vec![]
                            },
                            expression: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 21, 1, 22),
                                    errors: vec![]
                                },
                                name: "b".to_string(),
                            }),
                        }),
                    ],
                })),
            }),],
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 44),
                    errors: vec![]
                },
                expression: Expression::StringExpr(Box::new(StringExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 44),
                        errors: vec![]
                    },
                    parts: vec![
                        StringExprPart::Text(TextPart {
                            base: BaseNode {
                                location: loc.get(1, 2, 1, 5),
                                errors: vec![]
                            },
                            value: "we ".to_string(),
                        }),
                        StringExprPart::Interpolated(InterpolatedPart {
                            base: BaseNode {
                                location: loc.get(1, 5, 1, 34),
                                errors: vec![]
                            },
                            expression: Expression::StringExpr(Box::new(StringExpr {
                                base: BaseNode {
                                    location: loc.get(1, 7, 1, 33),
                                    errors: vec![]
                                },
                                parts: vec![
                                    StringExprPart::Text(TextPart {
                                        base: BaseNode {
                                            location: loc.get(1, 8, 1, 12),
                                            errors: vec![]
                                        },
                                        value: "can ".to_string(),
                                    }),
                                    StringExprPart::Interpolated(InterpolatedPart {
                                        base: BaseNode {
                                            location: loc.get(1, 12, 1, 32),
                                            errors: vec![]
                                        },
                                        expression: Expression::Binary(Box::new(BinaryExpr {
                                            base: BaseNode {
                                                location: loc.get(1, 14, 1, 31),
                                                errors: vec![]
                                            },
                                            left: Expression::StringLit(StringLit {
                                                base: BaseNode {
                                                    location: loc.get(1, 14, 1, 19),
                                                    errors: vec![]
                                                },
                                                value: "add".to_string(),
                                            }),
                                            right: Expression::StringLit(StringLit {
                                                base: BaseNode {
                                                    location: loc.get(1, 22, 1, 31),
                                                    errors: vec![]
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
                                errors: vec![]
                            },
                            value: " together".to_string(),
                        }),
                    ],
                })),
            }),],
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 45),
                    errors: vec![]
                },
                expression: Expression::StringExpr(Box::new(StringExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 45),
                        errors: vec![]
                    },
                    parts: vec![
                        StringExprPart::Text(TextPart {
                            base: BaseNode {
                                location: loc.get(1, 2, 1, 30),
                                errors: vec![]
                            },
                            value: "string \"interpolation with ".to_string(),
                        }),
                        StringExprPart::Interpolated(InterpolatedPart {
                            base: BaseNode {
                                location: loc.get(1, 30, 1, 42),
                                errors: vec![]
                            },
                            expression: Expression::StringLit(StringLit {
                                base: BaseNode {
                                    location: loc.get(1, 32, 1, 41),
                                    errors: vec![]
                                },
                                value: "escapes".to_string(),
                            }),
                        }),
                        StringExprPart::Text(TextPart {
                            base: BaseNode {
                                location: loc.get(1, 42, 1, 44),
                                errors: vec![]
                            },
                            value: "\"".to_string(),
                        }),
                    ],
                })),
            }),],
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 18),
                    errors: vec![]
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 3),
                        errors: vec![]
                    },
                    name: "fn".to_string(),
                },
                init: Expression::Function(Box::new(FunctionExpr {
                    base: BaseNode {
                        location: loc.get(1, 6, 1, 18),
                        errors: vec![]
                    },
                    params: vec![Property {
                        base: BaseNode {
                            location: loc.get(1, 7, 1, 8),
                            errors: vec![]
                        },
                        key: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 7, 1, 8),
                                errors: vec![]
                            },
                            name: "a".to_string()
                        }),
                        value: None
                    }],
                    body: FunctionBody::Expr(Expression::StringExpr(Box::new(StringExpr {
                        base: BaseNode {
                            location: loc.get(1, 13, 1, 18),
                            errors: vec![
                                "got unexpected token in string expression @1:18-1:18: EOF"
                                    .to_string()
                            ]
                        },
                        parts: vec![],
                    })))
                })),
            }))],
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
                errors: vec![]
            },
            name: "".to_string(),
            package: Some(PackageClause {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 12),
                    errors: vec![]
                },
                name: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 9, 1, 12),
                        errors: vec![]
                    },
                    name: "foo".to_string()
                }
            }),
            imports: vec![],
            body: vec![]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![ImportDeclaration {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 18),
                    errors: vec![]
                },
                alias: None,
                path: StringLit {
                    base: BaseNode {
                        location: loc.get(1, 8, 1, 18),
                        errors: vec![]
                    },
                    value: "path/foo".to_string()
                }
            }],
            body: vec![]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![ImportDeclaration {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 22),
                    errors: vec![]
                },
                alias: Some(Identifier {
                    base: BaseNode {
                        location: loc.get(1, 8, 1, 11),
                        errors: vec![]
                    },
                    name: "bar".to_string()
                }),
                path: StringLit {
                    base: BaseNode {
                        location: loc.get(1, 12, 1, 22),
                        errors: vec![]
                    },
                    value: "path/foo".to_string()
                }
            }],
            body: vec![]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![
                ImportDeclaration {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 18),
                        errors: vec![]
                    },
                    alias: None,
                    path: StringLit {
                        base: BaseNode {
                            location: loc.get(1, 8, 1, 18),
                            errors: vec![]
                        },
                        value: "path/foo".to_string()
                    }
                },
                ImportDeclaration {
                    base: BaseNode {
                        location: loc.get(2, 1, 2, 18),
                        errors: vec![]
                    },
                    alias: None,
                    path: StringLit {
                        base: BaseNode {
                            location: loc.get(2, 8, 2, 18),
                            errors: vec![]
                        },
                        value: "path/bar".to_string()
                    }
                }
            ],
            body: vec![]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: Some(PackageClause {
                base: BaseNode {
                    location: loc.get(2, 1, 2, 12),
                    errors: vec![]
                },
                name: Identifier {
                    base: BaseNode {
                        location: loc.get(2, 9, 2, 12),
                        errors: vec![]
                    },
                    name: "baz".to_string()
                }
            }),
            imports: vec![
                ImportDeclaration {
                    base: BaseNode {
                        location: loc.get(4, 1, 4, 18),
                        errors: vec![]
                    },
                    alias: None,
                    path: StringLit {
                        base: BaseNode {
                            location: loc.get(4, 8, 4, 18),
                            errors: vec![]
                        },
                        value: "path/foo".to_string()
                    }
                },
                ImportDeclaration {
                    base: BaseNode {
                        location: loc.get(5, 1, 5, 18),
                        errors: vec![]
                    },
                    alias: None,
                    path: StringLit {
                        base: BaseNode {
                            location: loc.get(5, 8, 5, 18),
                            errors: vec![]
                        },
                        value: "path/bar".to_string()
                    }
                }
            ],
            body: vec![]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: Some(PackageClause {
                base: BaseNode {
                    location: loc.get(2, 1, 2, 12),
                    errors: vec![]
                },
                name: Identifier {
                    base: BaseNode {
                        location: loc.get(2, 9, 2, 12),
                        errors: vec![]
                    },
                    name: "baz".to_string()
                }
            }),
            imports: vec![
                ImportDeclaration {
                    base: BaseNode {
                        location: loc.get(4, 1, 4, 18),
                        errors: vec![]
                    },
                    alias: None,
                    path: StringLit {
                        base: BaseNode {
                            location: loc.get(4, 8, 4, 18),
                            errors: vec![]
                        },
                        value: "path/foo".to_string()
                    }
                },
                ImportDeclaration {
                    base: BaseNode {
                        location: loc.get(5, 1, 5, 18),
                        errors: vec![]
                    },
                    alias: None,
                    path: StringLit {
                        base: BaseNode {
                            location: loc.get(5, 8, 5, 18),
                            errors: vec![]
                        },
                        value: "path/bar".to_string()
                    }
                }
            ],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(7, 1, 7, 6),
                    errors: vec![]
                },
                expression: Expression::Binary(Box::new(BinaryExpr {
                    base: BaseNode {
                        location: loc.get(7, 1, 7, 6),
                        errors: vec![]
                    },
                    operator: Operator::AdditionOperator,
                    left: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(7, 1, 7, 2),
                            errors: vec![]
                        },
                        value: 1
                    }),
                    right: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(7, 5, 7, 6),
                            errors: vec![]
                        },
                        value: 1
                    })
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Option(Box::new(OptionStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 7, 7),
                    errors: vec![]
                },
                assignment: Assignment::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(1, 8, 7, 7),
                        errors: vec![]
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(1, 8, 1, 12),
                            errors: vec![]
                        },
                        name: "task".to_string()
                    },
                    init: Expression::Object(Box::new(ObjectExpr {
                        base: BaseNode {
                            location: loc.get(1, 15, 7, 7),
                            errors: vec![]
                        },
                        with: None,
                        properties: vec![
                            Property {
                                base: BaseNode {
                                    location: loc.get(2, 5, 2, 16),
                                    errors: vec![]
                                },
                                key: PropertyKey::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(2, 5, 2, 9),
                                        errors: vec![]
                                    },
                                    name: "name".to_string()
                                }),
                                value: Some(Expression::StringLit(StringLit {
                                    base: BaseNode {
                                        location: loc.get(2, 11, 2, 16),
                                        errors: vec![]
                                    },
                                    value: "foo".to_string()
                                }))
                            },
                            Property {
                                base: BaseNode {
                                    location: loc.get(3, 5, 3, 14),
                                    errors: vec![]
                                },
                                key: PropertyKey::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(3, 5, 3, 10),
                                        errors: vec![]
                                    },
                                    name: "every".to_string()
                                }),
                                value: Some(Expression::Duration(DurationLit {
                                    base: BaseNode {
                                        location: loc.get(3, 12, 3, 14),
                                        errors: vec![]
                                    },
                                    values: vec![Duration {
                                        magnitude: 1,
                                        unit: "h".to_string()
                                    }]
                                }))
                            },
                            Property {
                                base: BaseNode {
                                    location: loc.get(4, 5, 4, 15),
                                    errors: vec![]
                                },
                                key: PropertyKey::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(4, 5, 4, 10),
                                        errors: vec![]
                                    },
                                    name: "delay".to_string()
                                }),
                                value: Some(Expression::Duration(DurationLit {
                                    base: BaseNode {
                                        location: loc.get(4, 12, 4, 15),
                                        errors: vec![]
                                    },
                                    values: vec![Duration {
                                        magnitude: 10,
                                        unit: "m".to_string()
                                    }]
                                }))
                            },
                            Property {
                                base: BaseNode {
                                    location: loc.get(5, 5, 5, 22),
                                    errors: vec![]
                                },
                                key: PropertyKey::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(5, 5, 5, 9),
                                        errors: vec![]
                                    },
                                    name: "cron".to_string()
                                }),
                                value: Some(Expression::StringLit(StringLit {
                                    base: BaseNode {
                                        location: loc.get(5, 11, 5, 22),
                                        errors: vec![]
                                    },
                                    value: "0 2 * * *".to_string()
                                }))
                            },
                            Property {
                                base: BaseNode {
                                    location: loc.get(6, 5, 6, 13),
                                    errors: vec![]
                                },
                                key: PropertyKey::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(6, 5, 6, 10),
                                        errors: vec![]
                                    },
                                    name: "retry".to_string()
                                }),
                                value: Some(Expression::Integer(IntegerLit {
                                    base: BaseNode {
                                        location: loc.get(6, 12, 6, 13),
                                        errors: vec![]
                                    },
                                    value: 5
                                }))
                            }
                        ]
                    }))
                }))
            }))]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Statement::Option(Box::new(OptionStmt {
                    base: BaseNode {
                        location: loc.get(1, 1, 4, 6),
                        errors: vec![]
                    },
                    assignment: Assignment::Variable(Box::new(VariableAssgn {
                        base: BaseNode {
                            location: loc.get(1, 8, 4, 6),
                            errors: vec![]
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: loc.get(1, 8, 1, 12),
                                errors: vec![]
                            },
                            name: "task".to_string()
                        },
                        init: Expression::Object(Box::new(ObjectExpr {
                            base: BaseNode {
                                location: loc.get(1, 15, 4, 6),
                                errors: vec![]
                            },
                            with: None,
                            properties: vec![
                                Property {
                                    base: BaseNode {
                                        location: loc.get(2, 6, 2, 17),
                                        errors: vec![]
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(2, 6, 2, 10),
                                            errors: vec![]
                                        },
                                        name: "name".to_string()
                                    }),
                                    value: Some(Expression::StringLit(StringLit {
                                        base: BaseNode {
                                            location: loc.get(2, 12, 2, 17),
                                            errors: vec![]
                                        },
                                        value: "foo".to_string()
                                    }))
                                },
                                Property {
                                    base: BaseNode {
                                        location: loc.get(3, 6, 3, 15),
                                        errors: vec![]
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(3, 6, 3, 11),
                                            errors: vec![]
                                        },
                                        name: "every".to_string()
                                    }),
                                    value: Some(Expression::Duration(DurationLit {
                                        base: BaseNode {
                                            location: loc.get(3, 13, 3, 15),
                                            errors: vec![]
                                        },
                                        values: vec![Duration {
                                            magnitude: 1,
                                            unit: "h".to_string()
                                        }]
                                    }))
                                }
                            ]
                        }))
                    }))
                })),
                Statement::Expr(ExprStmt {
                    base: BaseNode {
                        location: loc.get(7, 5, 7, 22),
                        errors: vec![]
                    },
                    expression: Expression::PipeExpr(Box::new(PipeExpr {
                        base: BaseNode {
                            location: loc.get(7, 5, 7, 22),
                            errors: vec![]
                        },
                        argument: Expression::Call(Box::new(CallExpr {
                            base: BaseNode {
                                location: loc.get(7, 5, 7, 11),
                                errors: vec![]
                            },
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(7, 5, 7, 9),
                                    errors: vec![]
                                },
                                name: "from".to_string()
                            }),
                            arguments: vec![]
                        })),
                        call: CallExpr {
                            base: BaseNode {
                                location: loc.get(7, 15, 7, 22),
                                errors: vec![]
                            },
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(7, 15, 7, 20),
                                    errors: vec![]
                                },
                                name: "count".to_string()
                            }),
                            arguments: vec![]
                        }
                    })),
                })
            ]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Option(Box::new(OptionStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 31),
                    errors: vec![]
                },
                assignment: Assignment::Member(Box::new(MemberAssgn {
                    base: BaseNode {
                        location: loc.get(1, 8, 1, 31),
                        errors: vec![]
                    },
                    member: MemberExpr {
                        base: BaseNode {
                            location: loc.get(1, 8, 1, 19),
                            errors: vec![]
                        },
                        object: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 8, 1, 13),
                                errors: vec![]
                            },
                            name: "alert".to_string()
                        }),
                        property: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 14, 1, 19),
                                errors: vec![]
                            },
                            name: "state".to_string()
                        })
                    },
                    init: Expression::StringLit(StringLit {
                        base: BaseNode {
                            location: loc.get(1, 22, 1, 31),
                            errors: vec![]
                        },
                        value: "Warning".to_string()
                    })
                }))
            }))]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Builtin(BuiltinStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 13),
                    errors: vec![]
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 9, 1, 13),
                        errors: vec![]
                    },
                    name: "from".to_string()
                }
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Test(Box::new(TestStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 30),
                    errors: vec![]
                },
                assignment: VariableAssgn {
                    base: BaseNode {
                        location: loc.get(1, 6, 1, 30),
                        errors: vec![]
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(1, 6, 1, 10),
                            errors: vec![]
                        },
                        name: "mean".to_string()
                    },
                    init: Expression::Object(Box::new(ObjectExpr {
                        base: BaseNode {
                            location: loc.get(1, 13, 1, 30),
                            errors: vec![]
                        },
                        with: None,
                        properties: vec![
                            Property {
                                base: BaseNode {
                                    location: loc.get(1, 14, 1, 21),
                                    errors: vec![]
                                },
                                key: PropertyKey::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 14, 1, 18),
                                        errors: vec![]
                                    },
                                    name: "want".to_string()
                                }),
                                value: Some(Expression::Integer(IntegerLit {
                                    base: BaseNode {
                                        location: loc.get(1, 20, 1, 21),
                                        errors: vec![]
                                    },
                                    value: 0
                                }))
                            },
                            Property {
                                base: BaseNode {
                                    location: loc.get(1, 23, 1, 29),
                                    errors: vec![]
                                },
                                key: PropertyKey::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 23, 1, 26),
                                        errors: vec![]
                                    },
                                    name: "got".to_string()
                                }),
                                value: Some(Expression::Integer(IntegerLit {
                                    base: BaseNode {
                                        location: loc.get(1, 28, 1, 29),
                                        errors: vec![]
                                    },
                                    value: 0
                                }))
                            }
                        ]
                    }))
                }
            }))]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 7),
                    errors: vec![]
                },
                expression: Expression::Call(Box::new(CallExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 7),
                        errors: vec![]
                    },
                    arguments: vec![],
                    callee: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 5),
                            errors: vec![]
                        },
                        name: "from".to_string()
                    })
                })),
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(2, 4, 2, 10),
                    errors: vec![]
                },
                expression: Expression::Call(Box::new(CallExpr {
                    base: BaseNode {
                        location: loc.get(2, 4, 2, 10),
                        errors: vec![]
                    },
                    arguments: vec![],
                    callee: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(2, 4, 2, 8),
                            errors: vec![]
                        },
                        name: "from".to_string()
                    })
                })),
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 7),
                    errors: vec![]
                },
                expression: Expression::Call(Box::new(CallExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 7),
                        errors: vec![]
                    },
                    arguments: vec![],
                    callee: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 5),
                            errors: vec![]
                        },
                        name: "tan2".to_string()
                    })
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 5),
                    errors: vec![]
                },
                expression: Expression::Regexp(RegexpLit {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 5),
                        errors: vec![]
                    },
                    value: ".*".to_string()
                })
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 12),
                    errors: vec![]
                },
                expression: Expression::Regexp(RegexpLit {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 12),
                        errors: vec![]
                    },
                    value: "a/b\\\\c\\d".to_string()
                })
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 4),
                    errors: vec![]
                },
                expression: Expression::Regexp(RegexpLit {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 4),
                        errors: vec![
                            "regex parse error: * error: repetition operator missing expression"
                                .to_string()
                        ]
                    },
                    value: "".to_string()
                })
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 28),
                    errors: vec![]
                },
                expression: Expression::Logical(Box::new(LogicalExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 28),
                        errors: vec![]
                    },
                    operator: LogicalOperator::AndOperator,
                    left: Expression::Binary(Box::new(BinaryExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 12),
                            errors: vec![]
                        },
                        operator: Operator::RegexpMatchOperator,
                        left: Expression::StringLit(StringLit {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 4),
                                errors: vec![]
                            },
                            value: "a".to_string()
                        }),
                        right: Expression::Regexp(RegexpLit {
                            base: BaseNode {
                                location: loc.get(1, 8, 1, 12),
                                errors: vec![]
                            },
                            value: ".*".to_string()
                        })
                    })),
                    right: Expression::Binary(Box::new(BinaryExpr {
                        base: BaseNode {
                            location: loc.get(1, 17, 1, 28),
                            errors: vec![]
                        },
                        operator: Operator::NotRegexpMatchOperator,
                        left: Expression::StringLit(StringLit {
                            base: BaseNode {
                                location: loc.get(1, 17, 1, 20),
                                errors: vec![]
                            },
                            value: "b".to_string()
                        }),
                        right: Expression::Regexp(RegexpLit {
                            base: BaseNode {
                                location: loc.get(1, 24, 1, 28),
                                errors: vec![]
                            },
                            value: "c$".to_string()
                        })
                    }))
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 10),
                    errors: vec![]
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 6),
                        errors: vec![]
                    },
                    name: "howdy".to_string()
                },
                init: Expression::Integer(IntegerLit {
                    base: BaseNode {
                        location: loc.get(1, 9, 1, 10),
                        errors: vec![]
                    },
                    value: 1
                })
            }))]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 12),
                    errors: vec![]
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 6),
                        errors: vec![]
                    },
                    name: "howdy".to_string()
                },
                init: Expression::Float(FloatLit {
                    base: BaseNode {
                        location: loc.get(1, 9, 1, 12),
                        errors: vec![]
                    },
                    value: 1.1
                })
            }))]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 21),
                    errors: vec![]
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 6),
                        errors: vec![]
                    },
                    name: "howdy".to_string()
                },
                init: Expression::Array(Box::new(ArrayExpr {
                    base: BaseNode {
                        location: loc.get(1, 9, 1, 21),
                        errors: vec![]
                    },
                    elements: vec![
                        Expression::Integer(IntegerLit {
                            base: BaseNode {
                                location: loc.get(1, 10, 1, 11),
                                errors: vec![]
                            },
                            value: 1
                        }),
                        Expression::Integer(IntegerLit {
                            base: BaseNode {
                                location: loc.get(1, 13, 1, 14),
                                errors: vec![]
                            },
                            value: 2
                        }),
                        Expression::Integer(IntegerLit {
                            base: BaseNode {
                                location: loc.get(1, 16, 1, 17),
                                errors: vec![]
                            },
                            value: 3
                        }),
                        Expression::Integer(IntegerLit {
                            base: BaseNode {
                                location: loc.get(1, 19, 1, 20),
                                errors: vec![]
                            },
                            value: 4
                        })
                    ]
                }))
            }))]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 11),
                    errors: vec![]
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 6),
                        errors: vec![]
                    },
                    name: "howdy".to_string()
                },
                init: Expression::Array(Box::new(ArrayExpr {
                    base: BaseNode {
                        location: loc.get(1, 9, 1, 11),
                        errors: vec![]
                    },
                    elements: vec![],
                }))
            }))]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 10),
                        errors: vec![]
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 6),
                            errors: vec![]
                        },
                        name: "howdy".to_string()
                    },
                    init: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 9, 1, 10),
                            errors: vec![]
                        },
                        value: 1
                    })
                })),
                Statement::Expr(ExprStmt {
                    base: BaseNode {
                        location: loc.get(2, 4, 2, 10),
                        errors: vec![]
                    },
                    expression: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(2, 4, 2, 10),
                            errors: vec![]
                        },
                        arguments: vec![],
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(2, 4, 2, 8),
                                errors: vec![]
                            },
                            name: "from".to_string()
                        })
                    })),
                })
            ]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 15),
                        errors: vec![]
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 6),
                            errors: vec![]
                        },
                        name: "howdy".to_string()
                    },
                    init: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 9, 1, 15),
                            errors: vec![]
                        },
                        arguments: vec![],
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 9, 1, 13),
                                errors: vec![]
                            },
                            name: "from".to_string()
                        })
                    })),
                })),
                Statement::Expr(ExprStmt {
                    base: BaseNode {
                        location: loc.get(2, 4, 2, 17),
                        errors: vec![]
                    },
                    expression: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(2, 4, 2, 17),
                            errors: vec![]
                        },
                        arguments: vec![],
                        callee: Expression::Member(Box::new(MemberExpr {
                            base: BaseNode {
                                location: loc.get(2, 4, 2, 15),
                                errors: vec![]
                            },
                            object: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(2, 4, 2, 9),
                                    errors: vec![]
                                },
                                name: "howdy".to_string()
                            }),
                            property: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(2, 10, 2, 15),
                                    errors: vec![]
                                },
                                name: "count".to_string()
                            })
                        }))
                    }))
                })
            ]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 18),
                    errors: vec![]
                },
                expression: Expression::PipeExpr(Box::new(PipeExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 18),
                        errors: vec![]
                    },
                    argument: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 7),
                            errors: vec![]
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 5),
                                errors: vec![]
                            },
                            name: "from".to_string()
                        }),
                        arguments: vec![]
                    })),
                    call: CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 11, 1, 18),
                            errors: vec![]
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 11, 1, 16),
                                errors: vec![]
                            },
                            name: "count".to_string()
                        }),
                        arguments: vec![]
                    }
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 14),
                    errors: vec![]
                },
                expression: Expression::PipeExpr(Box::new(PipeExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 14),
                        errors: vec![]
                    },
                    argument: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            errors: vec![]
                        },
                        name: "a".to_string()
                    }),
                    call: CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 6, 1, 14),
                            errors: vec![]
                        },
                        callee: Expression::Member(Box::new(MemberExpr {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 9),
                                errors: vec![]
                            },
                            object: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 7),
                                    errors: vec![]
                                },
                                name: "b".to_string()
                            }),
                            property: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 8, 1, 9),
                                    errors: vec![]
                                },
                                name: "c".to_string()
                            })
                        })),
                        arguments: vec![Expression::Object(Box::new(ObjectExpr {
                            base: BaseNode {
                                location: loc.get(1, 10, 1, 13),
                                errors: vec![]
                            },
                            with: None,
                            properties: vec![Property {
                                base: BaseNode {
                                    location: loc.get(1, 10, 1, 13),
                                    errors: vec![]
                                },
                                key: PropertyKey::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 10, 1, 11),
                                        errors: vec![]
                                    },
                                    name: "d".to_string()
                                }),
                                value: Some(Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 12, 1, 13),
                                        errors: vec![]
                                    },
                                    name: "e".to_string()
                                }))
                            }]
                        }))]
                    }
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 12),
                    errors: vec![]
                },
                expression: Expression::PipeExpr(Box::new(PipeExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 12),
                        errors: vec![]
                    },
                    argument: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            errors: vec![]
                        },
                        value: 5,
                    }),
                    call: CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 6, 1, 12),
                            errors: vec![]
                        },
                        arguments: vec![],
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 10),
                                errors: vec![]
                            },
                            name: "pow2".to_string()
                        })
                    },
                })),
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 17),
                    errors: vec![]
                },
                expression: Expression::PipeExpr(Box::new(PipeExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 17),
                        errors: vec![]
                    },
                    argument: Expression::Member(Box::new(MemberExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 8),
                            errors: vec![]
                        },
                        object: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 4),
                                errors: vec![]
                            },
                            name: "foo".to_string()
                        }),
                        property: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 5, 1, 8),
                                errors: vec![]
                            },
                            name: "bar".to_string()
                        })
                    })),
                    call: CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 12, 1, 17),
                            errors: vec![]
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 12, 1, 15),
                                errors: vec![]
                            },
                            name: "baz".to_string()
                        }),
                        arguments: vec![]
                    }
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 41),
                    errors: vec![]
                },
                expression: Expression::PipeExpr(Box::new(PipeExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 41),
                        errors: vec![]
                    },
                    argument: Expression::PipeExpr(Box::new(PipeExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 30),
                            errors: vec![]
                        },
                        argument: Expression::PipeExpr(Box::new(PipeExpr {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 18),
                                errors: vec![]
                            },
                            argument: Expression::Call(Box::new(CallExpr {
                                base: BaseNode {
                                    location: loc.get(1, 1, 1, 7),
                                    errors: vec![]
                                },
                                callee: Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 1, 1, 5),
                                        errors: vec![]
                                    },
                                    name: "from".to_string()
                                }),
                                arguments: vec![]
                            })),
                            call: CallExpr {
                                base: BaseNode {
                                    location: loc.get(1, 11, 1, 18),
                                    errors: vec![]
                                },
                                callee: Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 11, 1, 16),
                                        errors: vec![]
                                    },
                                    name: "range".to_string()
                                }),
                                arguments: vec![]
                            }
                        })),
                        call: CallExpr {
                            base: BaseNode {
                                location: loc.get(1, 22, 1, 30),
                                errors: vec![]
                            },
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 22, 1, 28),
                                    errors: vec![]
                                },
                                name: "filter".to_string()
                            }),
                            arguments: vec![]
                        }
                    })),
                    call: CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 34, 1, 41),
                            errors: vec![]
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 34, 1, 39),
                                errors: vec![]
                            },
                            name: "count".to_string()
                        }),
                        arguments: vec![]
                    }
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 13),
                    errors: vec![]
                },
                expression: Expression::PipeExpr(Box::new(PipeExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 13),
                        errors: vec![]
                    },
                    argument: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 6),
                            errors: vec![]
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 4),
                                errors: vec![]
                            },
                            name: "foo".to_string()
                        }),
                        arguments: vec![]
                    })),
                    call: CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 10, 1, 13),
                            errors: vec!["pipe destination must be a function call".to_string()]
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 10, 1, 13),
                                errors: vec![]
                            },
                            name: "bar".to_string()
                        }),
                        arguments: vec![]
                    }
                })),
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 15),
                        errors: vec![]
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 6),
                            errors: vec![]
                        },
                        name: "howdy".to_string()
                    },
                    init: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 9, 1, 15),
                            errors: vec![]
                        },
                        arguments: vec![],
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 9, 1, 13),
                                errors: vec![]
                            },
                            name: "from".to_string()
                        })
                    })),
                })),
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(2, 4, 2, 18),
                        errors: vec![]
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(2, 4, 2, 9),
                            errors: vec![]
                        },
                        name: "doody".to_string()
                    },
                    init: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(2, 12, 2, 18),
                            errors: vec![]
                        },
                        arguments: vec![],
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(2, 12, 2, 16),
                                errors: vec![]
                            },
                            name: "from".to_string()
                        })
                    })),
                })),
                Statement::Expr(ExprStmt {
                    base: BaseNode {
                        location: loc.get(3, 4, 3, 18),
                        errors: vec![]
                    },
                    expression: Expression::PipeExpr(Box::new(PipeExpr {
                        base: BaseNode {
                            location: loc.get(3, 4, 3, 18),
                            errors: vec![]
                        },
                        argument: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(3, 4, 3, 9),
                                errors: vec![]
                            },
                            name: "howdy".to_string()
                        }),
                        call: CallExpr {
                            base: BaseNode {
                                location: loc.get(3, 11, 3, 18),
                                errors: vec![]
                            },
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(3, 11, 3, 16),
                                    errors: vec![]
                                },
                                name: "count".to_string()
                            }),
                            arguments: vec![]
                        }
                    }))
                }),
                Statement::Expr(ExprStmt {
                    base: BaseNode {
                        location: loc.get(4, 4, 4, 16),
                        errors: vec![]
                    },
                    expression: Expression::PipeExpr(Box::new(PipeExpr {
                        base: BaseNode {
                            location: loc.get(4, 4, 4, 16),
                            errors: vec![]
                        },
                        argument: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(4, 4, 4, 9),
                                errors: vec![]
                            },
                            name: "doody".to_string()
                        }),
                        call: CallExpr {
                            base: BaseNode {
                                location: loc.get(4, 11, 4, 16),
                                errors: vec![]
                            },
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(4, 11, 4, 14),
                                    errors: vec![]
                                },
                                name: "sum".to_string()
                            }),
                            arguments: vec![]
                        }
                    }))
                })
            ]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 32),
                    errors: vec![]
                },
                expression: Expression::Call(Box::new(CallExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 32),
                        errors: vec![]
                    },
                    callee: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 5),
                            errors: vec![]
                        },
                        name: "from".to_string()
                    }),
                    arguments: vec![Expression::Object(Box::new(ObjectExpr {
                        base: BaseNode {
                            location: loc.get(1, 6, 1, 31),
                            errors: vec![]
                        },
                        with: None,
                        properties: vec![Property {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 31),
                                errors: vec![]
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 12),
                                    errors: vec![]
                                },
                                name: "bucket".to_string()
                            }),
                            value: Some(Expression::StringLit(StringLit {
                                base: BaseNode {
                                    location: loc.get(1, 13, 1, 31),
                                    errors: vec![]
                                },
                                value: "telegraf/autogen".to_string()
                            }))
                        }]
                    }))]
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 29),
                        errors: vec![]
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            errors: vec![]
                        },
                        name: "m".to_string()
                    },
                    init: Expression::Object(Box::new(ObjectExpr {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 29),
                            errors: vec![]
                        },
                        with: None,
                        properties: vec![
                            Property {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 13),
                                    errors: vec![]
                                },
                                key: PropertyKey::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 6, 1, 10),
                                        errors: vec![]
                                    },
                                    name: "key1".to_string()
                                }),
                                value: Some(Expression::Integer(IntegerLit {
                                    base: BaseNode {
                                        location: loc.get(1, 12, 1, 13),
                                        errors: vec![]
                                    },
                                    value: 1
                                }))
                            },
                            Property {
                                base: BaseNode {
                                    location: loc.get(1, 15, 1, 28),
                                    errors: vec![]
                                },
                                key: PropertyKey::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 15, 1, 19),
                                        errors: vec![]
                                    },
                                    name: "key2".to_string()
                                }),
                                value: Some(Expression::StringLit(StringLit {
                                    base: BaseNode {
                                        location: loc.get(1, 20, 1, 28),
                                        errors: vec![]
                                    },
                                    value: "value2".to_string()
                                }))
                            }
                        ]
                    }))
                })),
                Statement::Expr(ExprStmt {
                    base: BaseNode {
                        location: loc.get(2, 4, 2, 10),
                        errors: vec![]
                    },
                    expression: Expression::Member(Box::new(MemberExpr {
                        base: BaseNode {
                            location: loc.get(2, 4, 2, 10),
                            errors: vec![]
                        },
                        object: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(2, 4, 2, 5),
                                errors: vec![]
                            },
                            name: "m".to_string()
                        }),
                        property: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(2, 6, 2, 10),
                                errors: vec![]
                            },
                            name: "key1".to_string()
                        })
                    }))
                }),
                Statement::Expr(ExprStmt {
                    base: BaseNode {
                        location: loc.get(3, 4, 3, 13),
                        errors: vec![]
                    },
                    expression: Expression::Member(Box::new(MemberExpr {
                        base: BaseNode {
                            location: loc.get(3, 4, 3, 13),
                            errors: vec![]
                        },
                        object: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(3, 4, 3, 5),
                                errors: vec![]
                            },
                            name: "m".to_string()
                        }),
                        property: PropertyKey::StringLit(StringLit {
                            base: BaseNode {
                                location: loc.get(3, 6, 3, 12),
                                errors: vec![]
                            },
                            value: "key2".to_string()
                        })
                    }))
                })
            ]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 14),
                    errors: vec![]
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 2),
                        errors: vec![]
                    },
                    name: "x".to_string()
                },
                init: Expression::Object(Box::new(ObjectExpr {
                    base: BaseNode {
                        location: loc.get(1, 5, 1, 14),
                        errors: vec![]
                    },
                    with: None,
                    properties: vec![Property {
                        base: BaseNode {
                            location: loc.get(1, 6, 1, 13),
                            errors: vec![]
                        },
                        key: PropertyKey::StringLit(StringLit {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 9),
                                errors: vec![]
                            },
                            value: "a".to_string()
                        }),
                        value: Some(Expression::Integer(IntegerLit {
                            base: BaseNode {
                                location: loc.get(1, 11, 1, 13),
                                errors: vec![]
                            },
                            value: 10
                        }))
                    }]
                }))
            }))]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 21),
                    errors: vec![]
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 2),
                        errors: vec![]
                    },
                    name: "x".to_string()
                },
                init: Expression::Object(Box::new(ObjectExpr {
                    base: BaseNode {
                        location: loc.get(1, 5, 1, 21),
                        errors: vec![]
                    },
                    with: None,
                    properties: vec![
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 13),
                                errors: vec![]
                            },
                            key: PropertyKey::StringLit(StringLit {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 9),
                                    errors: vec![]
                                },
                                value: "a".to_string()
                            }),
                            value: Some(Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 11, 1, 13),
                                    errors: vec![]
                                },
                                value: 10
                            }))
                        },
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 15, 1, 20),
                                errors: vec![]
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 15, 1, 16),
                                    errors: vec![]
                                },
                                name: "b".to_string()
                            }),
                            value: Some(Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 18, 1, 20),
                                    errors: vec![]
                                },
                                value: 11
                            }))
                        }
                    ]
                }))
            }))]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 11),
                    errors: vec![]
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 2),
                        errors: vec![]
                    },
                    name: "x".to_string()
                },
                init: Expression::Object(Box::new(ObjectExpr {
                    base: BaseNode {
                        location: loc.get(1, 5, 1, 11),
                        errors: vec![]
                    },
                    with: None,
                    properties: vec![
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 7),
                                errors: vec![]
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 7),
                                    errors: vec![]
                                },
                                name: "a".to_string()
                            }),
                            value: None
                        },
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 9, 1, 10),
                                errors: vec![]
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 9, 1, 10),
                                    errors: vec![]
                                },
                                name: "b".to_string()
                            }),
                            value: None
                        }
                    ]
                }))
            }))]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 13),
                    errors: vec![]
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 2),
                        errors: vec![]
                    },
                    name: "x".to_string()
                },
                init: Expression::Object(Box::new(ObjectExpr {
                    base: BaseNode {
                        location: loc.get(1, 5, 1, 13),
                        errors: vec![]
                    },
                    with: None,
                    properties: vec![
                        Property {
                            // TODO(affo): this should error with ast.Check: "string literal key "a" must have a value".
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 9),
                                errors: vec![]
                            },
                            key: PropertyKey::StringLit(StringLit {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 9),
                                    errors: vec![]
                                },
                                value: "a".to_string()
                            }),
                            value: None
                        },
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 11, 1, 12),
                                errors: vec![]
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 11, 1, 12),
                                    errors: vec![]
                                },
                                name: "b".to_string()
                            }),
                            value: None
                        }
                    ]
                }))
            }))]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 13),
                    errors: vec![]
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 2),
                        errors: vec![]
                    },
                    name: "x".to_string()
                },
                init: Expression::Object(Box::new(ObjectExpr {
                    // TODO(affo): this should error in ast.Check(): "cannot mix implicit and explicit properties".
                    base: BaseNode {
                        location: loc.get(1, 5, 1, 13),
                        errors: vec![]
                    },
                    with: None,
                    properties: vec![
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 7),
                                errors: vec![]
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 7),
                                    errors: vec![]
                                },
                                name: "a".to_string()
                            }),
                            value: None
                        },
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 9, 1, 12),
                                errors: vec![]
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 9, 1, 10),
                                    errors: vec![]
                                },
                                name: "b".to_string()
                            }),
                            value: Some(Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 11, 1, 12),
                                    errors: vec![]
                                },
                                name: "c".to_string()
                            }))
                        }
                    ]
                }))
            }))]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 18),
                    errors: vec![]
                },
                expression: Expression::Object(Box::new(ObjectExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 18),
                        errors: vec![]
                    },
                    with: Some(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 2, 1, 3),
                            errors: vec![]
                        },
                        name: "a".to_string()
                    }),
                    properties: vec![
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 9, 1, 12),
                                errors: vec![]
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 9, 1, 10),
                                    errors: vec![]
                                },
                                name: "b".to_string()
                            }),
                            value: Some(Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 11, 1, 12),
                                    errors: vec![]
                                },
                                name: "c".to_string()
                            }))
                        },
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 14, 1, 17),
                                errors: vec![]
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 14, 1, 15),
                                    errors: vec![]
                                },
                                name: "d".to_string()
                            }),
                            value: Some(Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 16, 1, 17),
                                    errors: vec![]
                                },
                                name: "e".to_string()
                            }))
                        }
                    ]
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 14),
                    errors: vec![]
                },
                expression: Expression::Object(Box::new(ObjectExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 14),
                        errors: vec![]
                    },
                    with: Some(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 2, 1, 3),
                            errors: vec![]
                        },
                        name: "a".to_string()
                    }),
                    properties: vec![
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 9, 1, 10),
                                errors: vec![]
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 9, 1, 10),
                                    errors: vec![]
                                },
                                name: "b".to_string()
                            }),
                            value: None
                        },
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 12, 1, 13),
                                errors: vec![]
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 12, 1, 13),
                                    errors: vec![]
                                },
                                name: "c".to_string()
                            }),
                            value: None
                        }
                    ]
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 5),
                    errors: vec![]
                },
                expression: Expression::Index(Box::new(IndexExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 5),
                        errors: vec![]
                    },
                    array: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            errors: vec![]
                        },
                        name: "a".to_string()
                    }),
                    index: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 3, 1, 4),
                            errors: vec![]
                        },
                        value: 3
                    })
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 8),
                    errors: vec![]
                },
                expression: Expression::Index(Box::new(IndexExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 8),
                        errors: vec![]
                    },
                    array: Expression::Index(Box::new(IndexExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 5),
                            errors: vec![]
                        },
                        array: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 2),
                                errors: vec![]
                            },
                            name: "a".to_string()
                        }),
                        index: Expression::Integer(IntegerLit {
                            base: BaseNode {
                                location: loc.get(1, 3, 1, 4),
                                errors: vec![]
                            },
                            value: 3
                        })
                    })),
                    index: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 6, 1, 7),
                            errors: vec![]
                        },
                        value: 5
                    })
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 7),
                    errors: vec![]
                },
                expression: Expression::Index(Box::new(IndexExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 7),
                        errors: vec![]
                    },
                    array: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 4),
                            errors: vec![]
                        },
                        arguments: vec![],
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 2),
                                errors: vec![]
                            },
                            name: "f".to_string()
                        }),
                    })),
                    index: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 6),
                            errors: vec![]
                        },
                        value: 3
                    })
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 9),
                    errors: vec![]
                },
                expression: Expression::Member(Box::new(MemberExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 9),
                        errors: vec![]
                    },
                    object: Expression::Member(Box::new(MemberExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 4),
                            errors: vec![]
                        },
                        object: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 2),
                                errors: vec![]
                            },
                            name: "a".to_string()
                        }),
                        property: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 3, 1, 4),
                                errors: vec![]
                            },
                            name: "b".to_string()
                        })
                    })),
                    property: PropertyKey::StringLit(StringLit {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 8),
                            errors: vec![]
                        },
                        value: "c".to_string()
                    })
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 11),
                    errors: vec![]
                },
                expression: Expression::Member(Box::new(MemberExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 11),
                        errors: vec![]
                    },
                    object: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 6),
                            errors: vec![]
                        },
                        arguments: vec![],
                        callee: Expression::Member(Box::new(MemberExpr {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 4),
                                errors: vec![]
                            },
                            object: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 1, 1, 2),
                                    errors: vec![]
                                },
                                name: "a".to_string()
                            }),
                            property: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 3, 1, 4),
                                    errors: vec![]
                                },
                                name: "b".to_string()
                            })
                        })),
                    })),
                    property: PropertyKey::StringLit(StringLit {
                        base: BaseNode {
                            location: loc.get(1, 7, 1, 10),
                            errors: vec![]
                        },
                        value: "c".to_string()
                    })
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 6),
                    errors: vec![]
                },
                expression: Expression::Index(Box::new(IndexExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 6),
                        errors: vec!["expected RBRACK, got EOF".to_string()]
                    },
                    array: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            errors: vec![]
                        },
                        name: "a".to_string()
                    }),
                    index: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 3, 1, 6),
                            errors: vec![]
                        },
                        arguments: vec![],
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 3, 1, 4),
                                errors: vec![]
                            },
                            name: "b".to_string()
                        })
                    })),
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 6),
                    errors: vec![]
                },
                expression: Expression::Index(Box::new(IndexExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 6),
                        errors: vec![]
                    },
                    array: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            errors: vec![]
                        },
                        name: "a".to_string()
                    }),
                    index: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 3, 1, 6),
                            errors: vec!["expected RPAREN, got RBRACK".to_string()]
                        },
                        arguments: vec![],
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 3, 1, 4),
                                errors: vec![]
                            },
                            name: "b".to_string()
                        })
                    })),
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 6),
                    errors: vec![]
                },
                expression: Expression::Index(Box::new(IndexExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 6),
                        errors: vec!["invalid expression @1:4-1:5: )".to_string()]
                    },
                    array: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            errors: vec![]
                        },
                        name: "a".to_string()
                    }),
                    index: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 3, 1, 4),
                            errors: vec![]
                        },
                        name: "b".to_string()
                    })
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 14),
                    errors: vec![]
                },
                expression: Expression::Binary(Box::new(BinaryExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 14),
                        errors: vec![]
                    },
                    operator: Operator::LessThanOperator,
                    left: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 7),
                            errors: vec![]
                        },
                        name: "_value".to_string()
                    }),
                    right: Expression::Float(FloatLit {
                        base: BaseNode {
                            location: loc.get(1, 10, 1, 14),
                            errors: vec![]
                        },
                        value: 10.0
                    })
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 16),
                    errors: vec![]
                },
                expression: Expression::Binary(Box::new(BinaryExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 16),
                        errors: vec![]
                    },
                    operator: Operator::LessThanOperator,
                    left: Expression::Member(Box::new(MemberExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 9),
                            errors: vec![]
                        },
                        object: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 2),
                                errors: vec![]
                            },
                            name: "r".to_string()
                        }),
                        property: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 3, 1, 9),
                                errors: vec![]
                            },
                            name: "_value".to_string()
                        })
                    })),
                    right: Expression::Float(FloatLit {
                        base: BaseNode {
                            location: loc.get(1, 12, 1, 16),
                            errors: vec![]
                        },
                        value: 10.0
                    })
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 6),
                        errors: vec![]
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            errors: vec![]
                        },
                        name: "a".to_string()
                    },
                    init: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 6),
                            errors: vec![]
                        },
                        value: 1
                    })
                })),
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(2, 13, 2, 18),
                        errors: vec![]
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(2, 13, 2, 14),
                            errors: vec![]
                        },
                        name: "b".to_string()
                    },
                    init: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(2, 17, 2, 18),
                            errors: vec![]
                        },
                        value: 2
                    })
                })),
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(3, 13, 3, 22),
                        errors: vec![]
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(3, 13, 3, 14),
                            errors: vec![]
                        },
                        name: "c".to_string()
                    },
                    init: Expression::Binary(Box::new(BinaryExpr {
                        base: BaseNode {
                            location: loc.get(3, 17, 3, 22),
                            errors: vec![]
                        },
                        operator: Operator::AdditionOperator,
                        left: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(3, 17, 3, 18),
                                errors: vec![]
                            },
                            name: "a".to_string()
                        }),
                        right: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(3, 21, 3, 22),
                                errors: vec![]
                            },
                            name: "b".to_string()
                        })
                    }))
                })),
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(4, 13, 4, 18),
                        errors: vec![]
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(4, 13, 4, 14),
                            errors: vec![]
                        },
                        name: "d".to_string()
                    },
                    init: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(4, 17, 4, 18),
                            errors: vec![]
                        },
                        name: "a".to_string()
                    })
                }))
            ]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 6),
                        errors: vec![]
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            errors: vec![]
                        },
                        name: "a".to_string()
                    },
                    init: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 6),
                            errors: vec![]
                        },
                        value: 5
                    })
                })),
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(2, 13, 2, 19),
                        errors: vec![]
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(2, 13, 2, 14),
                            errors: vec![]
                        },
                        name: "c".to_string()
                    },
                    init: Expression::Unary(Box::new(UnaryExpr {
                        base: BaseNode {
                            location: loc.get(2, 17, 2, 19),
                            errors: vec![]
                        },
                        operator: Operator::SubtractionOperator,
                        argument: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(2, 18, 2, 19),
                                errors: vec![]
                            },
                            name: "a".to_string()
                        })
                    }))
                }))
            ]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 6),
                        errors: vec![]
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            errors: vec![]
                        },
                        name: "a".to_string()
                    },
                    init: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 6),
                            errors: vec![]
                        },
                        value: 5
                    })
                })),
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(2, 13, 2, 24),
                        errors: vec![]
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(2, 13, 2, 14),
                            errors: vec![]
                        },
                        name: "c".to_string()
                    },
                    init: Expression::Binary(Box::new(BinaryExpr {
                        base: BaseNode {
                            location: loc.get(2, 17, 2, 24),
                            errors: vec![]
                        },
                        operator: Operator::MultiplicationOperator,
                        left: Expression::Integer(IntegerLit {
                            base: BaseNode {
                                location: loc.get(2, 17, 2, 19),
                                errors: vec![]
                            },
                            value: 10
                        }),
                        right: Expression::Unary(Box::new(UnaryExpr {
                            base: BaseNode {
                                location: loc.get(2, 22, 2, 24),
                                errors: vec![]
                            },
                            operator: Operator::SubtractionOperator,
                            argument: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(2, 23, 2, 24),
                                    errors: vec![]
                                },
                                name: "a".to_string()
                            })
                        }))
                    }))
                }))
            ]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 8),
                        errors: vec![]
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            errors: vec![]
                        },
                        name: "a".to_string()
                    },
                    init: Expression::Float(FloatLit {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 8),
                            errors: vec![]
                        },
                        value: 5.0
                    })
                })),
                Statement::Expr(ExprStmt {
                    base: BaseNode {
                        location: loc.get(2, 13, 2, 42),
                        errors: vec![]
                    },
                    expression: Expression::Logical(Box::new(LogicalExpr {
                        base: BaseNode {
                            location: loc.get(2, 13, 2, 42),
                            errors: vec![]
                        },
                        operator: LogicalOperator::OrOperator,
                        left: Expression::Binary(Box::new(BinaryExpr {
                            base: BaseNode {
                                location: loc.get(2, 13, 2, 30),
                                errors: vec![]
                            },
                            operator: Operator::EqualOperator,
                            left: Expression::Binary(Box::new(BinaryExpr {
                                base: BaseNode {
                                    location: loc.get(2, 13, 2, 22),
                                    errors: vec![]
                                },
                                operator: Operator::MultiplicationOperator,
                                left: Expression::Float(FloatLit {
                                    base: BaseNode {
                                        location: loc.get(2, 13, 2, 17),
                                        errors: vec![]
                                    },
                                    value: 10.0
                                }),
                                right: Expression::Unary(Box::new(UnaryExpr {
                                    base: BaseNode {
                                        location: loc.get(2, 20, 2, 22),
                                        errors: vec![]
                                    },
                                    operator: Operator::SubtractionOperator,
                                    argument: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(2, 21, 2, 22),
                                            errors: vec![]
                                        },
                                        name: "a".to_string()
                                    })
                                }))
                            })),
                            right: Expression::Unary(Box::new(UnaryExpr {
                                base: BaseNode {
                                    location: loc.get(2, 26, 2, 30),
                                    errors: vec![]
                                },
                                operator: Operator::SubtractionOperator,
                                argument: Expression::Float(FloatLit {
                                    base: BaseNode {
                                        location: loc.get(2, 27, 2, 30),
                                        errors: vec![]
                                    },
                                    value: 0.5
                                })
                            }))
                        })),
                        right: Expression::Binary(Box::new(BinaryExpr {
                            base: BaseNode {
                                location: loc.get(2, 34, 2, 42),
                                errors: vec![]
                            },
                            operator: Operator::EqualOperator,
                            left: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(2, 34, 2, 35),
                                    errors: vec![]
                                },
                                name: "a".to_string()
                            }),
                            right: Expression::Float(FloatLit {
                                base: BaseNode {
                                    location: loc.get(2, 39, 2, 42),
                                    errors: vec![]
                                },
                                value: 6.0
                            })
                        }))
                    }))
                })
            ]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 8),
                    errors: vec![]
                },
                expression: Expression::Unary(Box::new(UnaryExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 8),
                        errors: vec![]
                    },
                    operator: Operator::NotOperator,
                    argument: Expression::Member(Box::new(MemberExpr {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 8),
                            errors: vec![]
                        },
                        object: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 5, 1, 6),
                                errors: vec![]
                            },
                            name: "m".to_string()
                        }),
                        property: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 7, 1, 8),
                                errors: vec![]
                            },
                            name: "b".to_string()
                        })
                    }))
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(2, 1, 2, 8),
                        errors: vec![]
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(2, 1, 2, 2),
                            errors: vec![]
                        },
                        name: "a".to_string()
                    },
                    init: Expression::Float(FloatLit {
                        base: BaseNode {
                            location: loc.get(2, 5, 2, 8),
                            errors: vec![]
                        },
                        value: 5.0
                    })
                })),
                Statement::Expr(ExprStmt {
                    base: BaseNode {
                        location: loc.get(4, 1, 6, 13),
                        errors: vec![]
                    },
                    expression: Expression::Logical(Box::new(LogicalExpr {
                        base: BaseNode {
                            location: loc.get(4, 1, 6, 13),
                            errors: vec![]
                        },
                        operator: LogicalOperator::OrOperator,
                        left: Expression::Binary(Box::new(BinaryExpr {
                            base: BaseNode {
                                location: loc.get(4, 1, 4, 18),
                                errors: vec![]
                            },
                            operator: Operator::EqualOperator,
                            left: Expression::Binary(Box::new(BinaryExpr {
                                base: BaseNode {
                                    location: loc.get(4, 1, 4, 10),
                                    errors: vec![]
                                },
                                operator: Operator::MultiplicationOperator,
                                left: Expression::Float(FloatLit {
                                    base: BaseNode {
                                        location: loc.get(4, 1, 4, 5),
                                        errors: vec![]
                                    },
                                    value: 10.0
                                }),
                                right: Expression::Unary(Box::new(UnaryExpr {
                                    base: BaseNode {
                                        location: loc.get(4, 8, 4, 10),
                                        errors: vec![]
                                    },
                                    operator: Operator::SubtractionOperator,
                                    argument: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(4, 9, 4, 10),
                                            errors: vec![]
                                        },
                                        name: "a".to_string()
                                    })
                                }))
                            })),
                            right: Expression::Unary(Box::new(UnaryExpr {
                                base: BaseNode {
                                    location: loc.get(4, 14, 4, 18),
                                    errors: vec![]
                                },
                                operator: Operator::SubtractionOperator,
                                argument: Expression::Float(FloatLit {
                                    base: BaseNode {
                                        location: loc.get(4, 15, 4, 18),
                                        errors: vec![]
                                    },
                                    value: 0.5
                                })
                            }))
                        })),
                        right: Expression::Binary(Box::new(BinaryExpr {
                            base: BaseNode {
                                location: loc.get(6, 5, 6, 13),
                                errors: vec![]
                            },
                            operator: Operator::EqualOperator,
                            left: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(6, 5, 6, 6),
                                    errors: vec![]
                                },
                                name: "a".to_string()
                            }),
                            right: Expression::Float(FloatLit {
                                base: BaseNode {
                                    location: loc.get(6, 10, 6, 13),
                                    errors: vec![]
                                },
                                value: 6.0
                            })
                        }))
                    }))
                })
            ]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 16),
                    errors: vec![]
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 2),
                        errors: vec![]
                    },
                    name: "a".to_string()
                },
                init: Expression::Binary(Box::new(BinaryExpr {
                    base: BaseNode {
                        location: loc.get(1, 5, 1, 16),
                        errors: vec![]
                    },
                    operator: Operator::EqualOperator,
                    left: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 10),
                            errors: vec![]
                        },
                        arguments: vec![],
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 5, 1, 8),
                                errors: vec![]
                            },
                            name: "foo".to_string()
                        })
                    })),
                    right: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 14, 1, 16),
                            errors: vec![]
                        },
                        value: 10
                    })
                }))
            }))]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(2, 13, 2, 43),
                    errors: vec![]
                },
                expression: Expression::Logical(Box::new(LogicalExpr {
                    base: BaseNode {
                        location: loc.get(2, 13, 2, 43),
                        errors: vec![]
                    },
                    operator: LogicalOperator::OrOperator,
                    left: Expression::Unary(Box::new(UnaryExpr {
                        base: BaseNode {
                            location: loc.get(2, 13, 2, 33),
                            errors: vec![]
                        },
                        operator: Operator::NotOperator,
                        argument: Expression::Paren(Box::new(ParenExpr {
                            base: BaseNode {
                                location: loc.get(2, 17, 2, 33),
                                errors: vec![]
                            },
                            expression: Expression::Binary(Box::new(BinaryExpr {
                                base: BaseNode {
                                    location: loc.get(2, 18, 2, 32),
                                    errors: vec![]
                                },
                                operator: Operator::EqualOperator,
                                left: Expression::Call(Box::new(CallExpr {
                                    base: BaseNode {
                                        location: loc.get(2, 18, 2, 21),
                                        errors: vec![]
                                    },
                                    arguments: vec![],
                                    callee: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(2, 18, 2, 19),
                                            errors: vec![]
                                        },
                                        name: "f".to_string()
                                    })
                                })),
                                right: Expression::Binary(Box::new(BinaryExpr {
                                    base: BaseNode {
                                        location: loc.get(2, 25, 2, 32),
                                        errors: vec![]
                                    },
                                    operator: Operator::MultiplicationOperator,
                                    left: Expression::Float(FloatLit {
                                        base: BaseNode {
                                            location: loc.get(2, 25, 2, 28),
                                            errors: vec![]
                                        },
                                        value: 6.0
                                    }),
                                    right: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(2, 31, 2, 32),
                                            errors: vec![]
                                        },
                                        name: "x".to_string()
                                    })
                                }))
                            }))
                        }))
                    })),
                    right: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(2, 37, 2, 43),
                            errors: vec![]
                        },
                        arguments: vec![],
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(2, 37, 2, 41),
                                errors: vec![]
                            },
                            name: "fail".to_string()
                        })
                    })),
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(2, 13, 2, 45),
                    errors: vec![]
                },
                expression: Expression::Paren(Box::new(ParenExpr {
                    base: BaseNode {
                        location: loc.get(2, 13, 2, 45),
                        errors: vec![]
                    },
                    expression: Expression::Logical(Box::new(LogicalExpr {
                        base: BaseNode {
                            location: loc.get(2, 14, 2, 44),
                            errors: vec![]
                        },
                        operator: LogicalOperator::OrOperator,
                        left: Expression::Unary(Box::new(UnaryExpr {
                            base: BaseNode {
                                location: loc.get(2, 14, 2, 34),
                                errors: vec![]
                            },
                            operator: Operator::NotOperator,
                            argument: Expression::Paren(Box::new(ParenExpr {
                                base: BaseNode {
                                    location: loc.get(2, 18, 2, 34),
                                    errors: vec![]
                                },
                                expression: Expression::Binary(Box::new(BinaryExpr {
                                    base: BaseNode {
                                        location: loc.get(2, 19, 2, 33),
                                        errors: vec![]
                                    },
                                    operator: Operator::EqualOperator,
                                    left: Expression::Call(Box::new(CallExpr {
                                        base: BaseNode {
                                            location: loc.get(2, 19, 2, 22),
                                            errors: vec![]
                                        },
                                        arguments: vec![],
                                        callee: Expression::Identifier(Identifier {
                                            base: BaseNode {
                                                location: loc.get(2, 19, 2, 20),
                                                errors: vec![]
                                            },
                                            name: "f".to_string()
                                        })
                                    })),
                                    right: Expression::Binary(Box::new(BinaryExpr {
                                        base: BaseNode {
                                            location: loc.get(2, 26, 2, 33),
                                            errors: vec![]
                                        },
                                        operator: Operator::MultiplicationOperator,
                                        left: Expression::Float(FloatLit {
                                            base: BaseNode {
                                                location: loc.get(2, 26, 2, 29),
                                                errors: vec![]
                                            },
                                            value: 6.0
                                        }),
                                        right: Expression::Identifier(Identifier {
                                            base: BaseNode {
                                                location: loc.get(2, 32, 2, 33),
                                                errors: vec![]
                                            },
                                            name: "x".to_string()
                                        })
                                    }))
                                }))
                            }))
                        })),
                        right: Expression::Call(Box::new(CallExpr {
                            base: BaseNode {
                                location: loc.get(2, 38, 2, 44),
                                errors: vec![]
                            },
                            arguments: vec![],
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(2, 38, 2, 42),
                                    errors: vec![]
                                },
                                name: "fail".to_string()
                            })
                        })),
                    }))
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 6),
                    errors: vec![]
                },
                expression: Expression::Binary(Box::new(BinaryExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 6),
                        errors: vec![]
                    },
                    operator: Operator::ModuloOperator,
                    left: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            errors: vec![]
                        },
                        value: 3
                    }),
                    right: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 6),
                            errors: vec![]
                        },
                        value: 8
                    })
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 10),
                    errors: vec![]
                },
                expression: Expression::Binary(Box::new(BinaryExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 10),
                        errors: vec![]
                    },
                    operator: Operator::ModuloOperator,
                    left: Expression::Float(FloatLit {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 4),
                            errors: vec![]
                        },
                        value: 8.3
                    }),
                    right: Expression::Float(FloatLit {
                        base: BaseNode {
                            location: loc.get(1, 7, 1, 10),
                            errors: vec![]
                        },
                        value: 3.1
                    })
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 6),
                    errors: vec![]
                },
                expression: Expression::Binary(Box::new(BinaryExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 6),
                        errors: vec![]
                    },
                    operator: Operator::PowerOperator,
                    left: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            errors: vec![]
                        },
                        value: 2
                    }),
                    right: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 6),
                            errors: vec![]
                        },
                        value: 4
                    })
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 12),
                    errors: vec![]
                },
                expression: Expression::Binary(Box::new(BinaryExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 12),
                        errors: vec![]
                    },
                    operator: Operator::SubtractionOperator,
                    left: Expression::Binary(Box::new(BinaryExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 6),
                            errors: vec![]
                        },
                        operator: Operator::DivisionOperator,
                        left: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 2),
                                errors: vec![]
                            },
                            name: "a".to_string()
                        }),
                        right: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 5, 1, 6),
                                errors: vec![]
                            },
                            name: "b".to_string()
                        })
                    })),
                    right: Expression::Float(FloatLit {
                        base: BaseNode {
                            location: loc.get(1, 9, 1, 12),
                            errors: vec![]
                        },
                        value: 1.0
                    })
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 14),
                    errors: vec![]
                },
                expression: Expression::Binary(Box::new(BinaryExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 14),
                        errors: vec![]
                    },
                    operator: Operator::SubtractionOperator,
                    left: Expression::Binary(Box::new(BinaryExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 8),
                            errors: vec![]
                        },
                        operator: Operator::DivisionOperator,
                        left: Expression::Integer(IntegerLit {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 2),
                                errors: vec![]
                            },
                            value: 2
                        }),
                        right: Expression::StringLit(StringLit {
                            base: BaseNode {
                                location: loc.get(1, 5, 1, 8),
                                errors: vec![]
                            },
                            value: "a".to_string()
                        })
                    })),
                    right: Expression::Float(FloatLit {
                        base: BaseNode {
                            location: loc.get(1, 11, 1, 14),
                            errors: vec![]
                        },
                        value: 1.0
                    })
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 10),
                    errors: vec![]
                },
                expression: Expression::Binary(Box::new(BinaryExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 10),
                        errors: vec![]
                    },
                    operator: Operator::SubtractionOperator,
                    left: Expression::Binary(Box::new(BinaryExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 6),
                            errors: vec![]
                        },
                        operator: Operator::SubtractionOperator,
                        left: Expression::Integer(IntegerLit {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 2),
                                errors: vec![]
                            },
                            value: 1
                        }),
                        right: Expression::Integer(IntegerLit {
                            base: BaseNode {
                                location: loc.get(1, 5, 1, 6),
                                errors: vec![]
                            },
                            value: 2
                        })
                    })),
                    right: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 9, 1, 10),
                            errors: vec![]
                        },
                        value: 3
                    })
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 12),
                    errors: vec![]
                },
                expression: Expression::Binary(Box::new(BinaryExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 12),
                        errors: vec![]
                    },
                    operator: Operator::SubtractionOperator,
                    left: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            errors: vec![]
                        },
                        value: 1
                    }),
                    right: Expression::Paren(Box::new(ParenExpr {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 12),
                            errors: vec![]
                        },
                        expression: Expression::Binary(Box::new(BinaryExpr {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 11),
                                errors: vec![]
                            },
                            operator: Operator::SubtractionOperator,
                            left: Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 7),
                                    errors: vec![]
                                },
                                value: 2
                            }),
                            right: Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 10, 1, 11),
                                    errors: vec![]
                                },
                                value: 3
                            })
                        }))
                    }))
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 10),
                    errors: vec![]
                },
                expression: Expression::Binary(Box::new(BinaryExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 10),
                        errors: vec![]
                    },
                    operator: Operator::AdditionOperator,
                    left: Expression::Binary(Box::new(BinaryExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 6),
                            errors: vec![]
                        },
                        operator: Operator::AdditionOperator,
                        left: Expression::Integer(IntegerLit {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 2),
                                errors: vec![]
                            },
                            value: 1
                        }),
                        right: Expression::Integer(IntegerLit {
                            base: BaseNode {
                                location: loc.get(1, 5, 1, 6),
                                errors: vec![]
                            },
                            value: 2
                        })
                    })),
                    right: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 9, 1, 10),
                            errors: vec![]
                        },
                        value: 3
                    })
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 12),
                    errors: vec![]
                },
                expression: Expression::Binary(Box::new(BinaryExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 12),
                        errors: vec![]
                    },
                    operator: Operator::AdditionOperator,
                    left: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            errors: vec![]
                        },
                        value: 1
                    }),
                    right: Expression::Paren(Box::new(ParenExpr {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 12),
                            errors: vec![]
                        },
                        expression: Expression::Binary(Box::new(BinaryExpr {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 11),
                                errors: vec![]
                            },
                            operator: Operator::AdditionOperator,
                            left: Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 7),
                                    errors: vec![]
                                },
                                value: 2
                            }),
                            right: Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 10, 1, 11),
                                    errors: vec![]
                                },
                                value: 3
                            })
                        }))
                    }))
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 12),
                    errors: vec![]
                },
                expression: Expression::Unary(Box::new(UnaryExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 12),
                        errors: vec![]
                    },
                    operator: Operator::NotOperator,
                    argument: Expression::Binary(Box::new(BinaryExpr {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 12),
                            errors: vec![]
                        },
                        operator: Operator::EqualOperator,
                        left: Expression::Unary(Box::new(UnaryExpr {
                            base: BaseNode {
                                location: loc.get(1, 5, 1, 7),
                                errors: vec![]
                            },
                            operator: Operator::SubtractionOperator,
                            argument: Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 7),
                                    errors: vec![]
                                },
                                value: 1
                            })
                        })),
                        right: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 11, 1, 12),
                                errors: vec![]
                            },
                            name: "a".to_string()
                        })
                    }))
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 2, 72),
                    errors: vec![]
                },
                expression: Expression::Logical(Box::new(LogicalExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 2, 72),
                        errors: vec![]
                    },
                    operator: LogicalOperator::OrOperator,
                    left: Expression::Logical(Box::new(LogicalExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 2, 32),
                            errors: vec![]
                        },
                        operator: LogicalOperator::OrOperator,
                        left: Expression::Logical(Box::new(LogicalExpr {
                            base: BaseNode {
                                location: loc.get(1, 1, 2, 18),
                                errors: vec![]
                            },
                            operator: LogicalOperator::AndOperator,
                            left: Expression::Logical(Box::new(LogicalExpr {
                                base: BaseNode {
                                    location: loc.get(1, 1, 1, 55),
                                    errors: vec![]
                                },
                                operator: LogicalOperator::AndOperator,
                                left: Expression::Logical(Box::new(LogicalExpr {
                                    base: BaseNode {
                                        location: loc.get(1, 1, 1, 41),
                                        errors: vec![]
                                    },
                                    operator: LogicalOperator::AndOperator,
                                    left: Expression::Binary(Box::new(BinaryExpr {
                                        base: BaseNode {
                                            location: loc.get(1, 1, 1, 27),
                                            errors: vec![]
                                        },
                                        operator: Operator::LessThanOperator,
                                        left: Expression::Binary(Box::new(BinaryExpr {
                                            base: BaseNode {
                                                location: loc.get(1, 1, 1, 21),
                                                errors: vec![]
                                            },
                                            operator: Operator::EqualOperator,
                                            left: Expression::Call(Box::new(CallExpr {
                                                base: BaseNode {
                                                    location: loc.get(1, 1, 1, 4),
                                                    errors: vec![]
                                                },
                                                arguments: vec![],
                                                callee: Expression::Identifier(Identifier {
                                                    base: BaseNode {
                                                        location: loc.get(1, 1, 1, 2),
                                                        errors: vec![]
                                                    },
                                                    name: "a".to_string()
                                                })
                                            })),
                                            right: Expression::Binary(Box::new(BinaryExpr {
                                                base: BaseNode {
                                                    location: loc.get(1, 8, 1, 21),
                                                    errors: vec![]
                                                },
                                                operator: Operator::AdditionOperator,
                                                left: Expression::Member(Box::new(MemberExpr {
                                                    base: BaseNode {
                                                        location: loc.get(1, 8, 1, 11),
                                                        errors: vec![]
                                                    },
                                                    object: Expression::Identifier(Identifier {
                                                        base: BaseNode {
                                                            location: loc.get(1, 8, 1, 9),
                                                            errors: vec![]
                                                        },
                                                        name: "b".to_string()
                                                    }),
                                                    property: PropertyKey::Identifier(Identifier {
                                                        base: BaseNode {
                                                            location: loc.get(1, 10, 1, 11),
                                                            errors: vec![]
                                                        },
                                                        name: "a".to_string()
                                                    })
                                                })),
                                                right: Expression::Binary(Box::new(BinaryExpr {
                                                    base: BaseNode {
                                                        location: loc.get(1, 14, 1, 21),
                                                        errors: vec![]
                                                    },
                                                    operator: Operator::MultiplicationOperator,
                                                    left: Expression::Member(Box::new(
                                                        MemberExpr {
                                                            base: BaseNode {
                                                                location: loc.get(1, 14, 1, 17),
                                                                errors: vec![]
                                                            },
                                                            object: Expression::Identifier(
                                                                Identifier {
                                                                    base: BaseNode {
                                                                        location: loc
                                                                            .get(1, 14, 1, 15),
                                                                        errors: vec![]
                                                                    },
                                                                    name: "b".to_string()
                                                                }
                                                            ),
                                                            property: PropertyKey::Identifier(
                                                                Identifier {
                                                                    base: BaseNode {
                                                                        location: loc
                                                                            .get(1, 16, 1, 17),
                                                                        errors: vec![]
                                                                    },
                                                                    name: "c".to_string()
                                                                }
                                                            )
                                                        }
                                                    )),
                                                    right: Expression::Identifier(Identifier {
                                                        base: BaseNode {
                                                            location: loc.get(1, 20, 1, 21),
                                                            errors: vec![]
                                                        },
                                                        name: "d".to_string()
                                                    })
                                                }))
                                            }))
                                        })),
                                        right: Expression::Integer(IntegerLit {
                                            base: BaseNode {
                                                location: loc.get(1, 24, 1, 27),
                                                errors: vec![]
                                            },
                                            value: 100
                                        })
                                    })),
                                    right: Expression::Binary(Box::new(BinaryExpr {
                                        base: BaseNode {
                                            location: loc.get(1, 32, 1, 41),
                                            errors: vec![]
                                        },
                                        operator: Operator::NotEqualOperator,
                                        left: Expression::Identifier(Identifier {
                                            base: BaseNode {
                                                location: loc.get(1, 32, 1, 33),
                                                errors: vec![]
                                            },
                                            name: "e".to_string()
                                        }),
                                        right: Expression::Index(Box::new(IndexExpr {
                                            base: BaseNode {
                                                location: loc.get(1, 37, 1, 41),
                                                errors: vec![]
                                            },
                                            array: Expression::Identifier(Identifier {
                                                base: BaseNode {
                                                    location: loc.get(1, 37, 1, 38),
                                                    errors: vec![]
                                                },
                                                name: "f".to_string()
                                            }),
                                            index: Expression::Identifier(Identifier {
                                                base: BaseNode {
                                                    location: loc.get(1, 39, 1, 40),
                                                    errors: vec![]
                                                },
                                                name: "g".to_string()
                                            })
                                        }))
                                    }))
                                })),
                                right: Expression::Binary(Box::new(BinaryExpr {
                                    base: BaseNode {
                                        location: loc.get(1, 46, 1, 55),
                                        errors: vec![]
                                    },
                                    operator: Operator::GreaterThanOperator,
                                    left: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(1, 46, 1, 47),
                                            errors: vec![]
                                        },
                                        name: "h".to_string()
                                    }),
                                    right: Expression::Binary(Box::new(BinaryExpr {
                                        base: BaseNode {
                                            location: loc.get(1, 50, 1, 55),
                                            errors: vec![]
                                        },
                                        operator: Operator::MultiplicationOperator,
                                        left: Expression::Identifier(Identifier {
                                            base: BaseNode {
                                                location: loc.get(1, 50, 1, 51),
                                                errors: vec![]
                                            },
                                            name: "i".to_string()
                                        }),
                                        right: Expression::Identifier(Identifier {
                                            base: BaseNode {
                                                location: loc.get(1, 54, 1, 55),
                                                errors: vec![]
                                            },
                                            name: "j".to_string()
                                        })
                                    }))
                                }))
                            })),
                            right: Expression::Binary(Box::new(BinaryExpr {
                                base: BaseNode {
                                    location: loc.get(2, 1, 2, 18),
                                    errors: vec![]
                                },
                                operator: Operator::LessThanOperator,
                                left: Expression::Binary(Box::new(BinaryExpr {
                                    base: BaseNode {
                                        location: loc.get(2, 1, 2, 6),
                                        errors: vec![]
                                    },
                                    operator: Operator::DivisionOperator,
                                    left: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(2, 1, 2, 2),
                                            errors: vec![]
                                        },
                                        name: "k".to_string()
                                    }),
                                    right: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(2, 5, 2, 6),
                                            errors: vec![]
                                        },
                                        name: "l".to_string()
                                    })
                                })),
                                right: Expression::Binary(Box::new(BinaryExpr {
                                    base: BaseNode {
                                        location: loc.get(2, 9, 2, 18),
                                        errors: vec![]
                                    },
                                    operator: Operator::SubtractionOperator,
                                    left: Expression::Binary(Box::new(BinaryExpr {
                                        base: BaseNode {
                                            location: loc.get(2, 9, 2, 14),
                                            errors: vec![]
                                        },
                                        operator: Operator::AdditionOperator,
                                        left: Expression::Identifier(Identifier {
                                            base: BaseNode {
                                                location: loc.get(2, 9, 2, 10),
                                                errors: vec![]
                                            },
                                            name: "m".to_string()
                                        }),
                                        right: Expression::Identifier(Identifier {
                                            base: BaseNode {
                                                location: loc.get(2, 13, 2, 14),
                                                errors: vec![]
                                            },
                                            name: "n".to_string()
                                        })
                                    })),
                                    right: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(2, 17, 2, 18),
                                            errors: vec![]
                                        },
                                        name: "o".to_string()
                                    })
                                }))
                            }))
                        })),
                        right: Expression::Binary(Box::new(BinaryExpr {
                            base: BaseNode {
                                location: loc.get(2, 22, 2, 32),
                                errors: vec![]
                            },
                            operator: Operator::LessThanEqualOperator,
                            left: Expression::Call(Box::new(CallExpr {
                                base: BaseNode {
                                    location: loc.get(2, 22, 2, 25),
                                    errors: vec![]
                                },
                                arguments: vec![],
                                callee: Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(2, 22, 2, 23),
                                        errors: vec![]
                                    },
                                    name: "p".to_string()
                                })
                            })),
                            right: Expression::Call(Box::new(CallExpr {
                                base: BaseNode {
                                    location: loc.get(2, 29, 2, 32),
                                    errors: vec![]
                                },
                                arguments: vec![],
                                callee: Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(2, 29, 2, 30),
                                        errors: vec![]
                                    },
                                    name: "q".to_string()
                                })
                            }))
                        }))
                    })),
                    right: Expression::Logical(Box::new(LogicalExpr {
                        base: BaseNode {
                            location: loc.get(2, 36, 2, 72),
                            errors: vec![]
                        },
                        operator: LogicalOperator::AndOperator,
                        left: Expression::Logical(Box::new(LogicalExpr {
                            base: BaseNode {
                                location: loc.get(2, 36, 2, 59),
                                errors: vec![]
                            },
                            operator: LogicalOperator::AndOperator,
                            left: Expression::Binary(Box::new(BinaryExpr {
                                base: BaseNode {
                                    location: loc.get(2, 36, 2, 42),
                                    errors: vec![]
                                },
                                operator: Operator::GreaterThanEqualOperator,
                                left: Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(2, 36, 2, 37),
                                        errors: vec![]
                                    },
                                    name: "r".to_string()
                                }),
                                right: Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(2, 41, 2, 42),
                                        errors: vec![]
                                    },
                                    name: "s".to_string()
                                })
                            })),
                            right: Expression::Unary(Box::new(UnaryExpr {
                                base: BaseNode {
                                    location: loc.get(2, 47, 2, 59),
                                    errors: vec![]
                                },
                                operator: Operator::NotOperator,
                                argument: Expression::Binary(Box::new(BinaryExpr {
                                    base: BaseNode {
                                        location: loc.get(2, 51, 2, 59),
                                        errors: vec![]
                                    },
                                    operator: Operator::RegexpMatchOperator,
                                    left: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(2, 51, 2, 52),
                                            errors: vec![]
                                        },
                                        name: "t".to_string()
                                    }),
                                    right: Expression::Regexp(RegexpLit {
                                        base: BaseNode {
                                            location: loc.get(2, 56, 2, 59),
                                            errors: vec![]
                                        },
                                        value: "a".to_string()
                                    })
                                }))
                            }))
                        })),
                        right: Expression::Binary(Box::new(BinaryExpr {
                            base: BaseNode {
                                location: loc.get(2, 64, 2, 72),
                                errors: vec![]
                            },
                            operator: Operator::NotRegexpMatchOperator,
                            left: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(2, 64, 2, 65),
                                    errors: vec![]
                                },
                                name: "u".to_string()
                            }),
                            right: Expression::Regexp(RegexpLit {
                                base: BaseNode {
                                    location: loc.get(2, 69, 2, 72),
                                    errors: vec![]
                                },
                                value: "a".to_string()
                            })
                        }))
                    }))
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 11),
                    errors: vec![]
                },
                expression: Expression::Logical(Box::new(LogicalExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 11),
                        errors: vec![]
                    },
                    operator: LogicalOperator::OrOperator,
                    left: Expression::Unary(Box::new(UnaryExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 6),
                            errors: vec![]
                        },
                        operator: Operator::NotOperator,
                        argument: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 5, 1, 6),
                                errors: vec![]
                            },
                            name: "a".to_string()
                        })
                    })),
                    right: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 10, 1, 11),
                            errors: vec![]
                        },
                        name: "b".to_string()
                    })
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 11),
                    errors: vec![]
                },
                expression: Expression::Logical(Box::new(LogicalExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 11),
                        errors: vec![]
                    },
                    operator: LogicalOperator::OrOperator,
                    left: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            errors: vec![]
                        },
                        name: "a".to_string()
                    }),
                    right: Expression::Unary(Box::new(UnaryExpr {
                        base: BaseNode {
                            location: loc.get(1, 6, 1, 11),
                            errors: vec![]
                        },
                        operator: Operator::NotOperator,
                        argument: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 10, 1, 11),
                                errors: vec![]
                            },
                            name: "b".to_string()
                        })
                    }))
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 12),
                    errors: vec![]
                },
                expression: Expression::Logical(Box::new(LogicalExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 12),
                        errors: vec![]
                    },
                    operator: LogicalOperator::AndOperator,
                    left: Expression::Unary(Box::new(UnaryExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 6),
                            errors: vec![]
                        },
                        operator: Operator::NotOperator,
                        argument: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 5, 1, 6),
                                errors: vec![]
                            },
                            name: "a".to_string()
                        })
                    })),
                    right: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 11, 1, 12),
                            errors: vec![]
                        },
                        name: "b".to_string()
                    })
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 12),
                    errors: vec![]
                },
                expression: Expression::Logical(Box::new(LogicalExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 12),
                        errors: vec![]
                    },
                    operator: LogicalOperator::AndOperator,
                    left: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            errors: vec![]
                        },
                        name: "a".to_string()
                    }),
                    right: Expression::Unary(Box::new(UnaryExpr {
                        base: BaseNode {
                            location: loc.get(1, 7, 1, 12),
                            errors: vec![]
                        },
                        operator: Operator::NotOperator,
                        argument: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 11, 1, 12),
                                errors: vec![]
                            },
                            name: "b".to_string()
                        })
                    }))
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 13),
                    errors: vec![]
                },
                expression: Expression::Logical(Box::new(LogicalExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 13),
                        errors: vec![]
                    },
                    operator: LogicalOperator::OrOperator,
                    left: Expression::Logical(Box::new(LogicalExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 8),
                            errors: vec![]
                        },
                        operator: LogicalOperator::AndOperator,
                        left: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 2),
                                errors: vec![]
                            },
                            name: "a".to_string()
                        }),
                        right: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 7, 1, 8),
                                errors: vec![]
                            },
                            name: "b".to_string()
                        })
                    })),
                    right: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 12, 1, 13),
                            errors: vec![]
                        },
                        name: "c".to_string()
                    })
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 13),
                    errors: vec![]
                },
                expression: Expression::Logical(Box::new(LogicalExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 13),
                        errors: vec![]
                    },
                    operator: LogicalOperator::OrOperator,
                    left: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            errors: vec![]
                        },
                        name: "a".to_string()
                    }),
                    right: Expression::Logical(Box::new(LogicalExpr {
                        base: BaseNode {
                            location: loc.get(1, 6, 1, 13),
                            errors: vec![]
                        },
                        operator: LogicalOperator::AndOperator,
                        left: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 7),
                                errors: vec![]
                            },
                            name: "b".to_string()
                        }),
                        right: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 12, 1, 13),
                                errors: vec![]
                            },
                            name: "c".to_string()
                        })
                    }))
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 13),
                    errors: vec![]
                },
                expression: Expression::Unary(Box::new(UnaryExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 13),
                        errors: vec![]
                    },
                    operator: Operator::NotOperator,
                    argument: Expression::Paren(Box::new(ParenExpr {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 13),
                            errors: vec![]
                        },
                        expression: Expression::Logical(Box::new(LogicalExpr {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 12),
                                errors: vec![]
                            },
                            operator: LogicalOperator::OrOperator,
                            left: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 7),
                                    errors: vec![]
                                },
                                name: "a".to_string()
                            }),
                            right: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 11, 1, 12),
                                    errors: vec![]
                                },
                                name: "b".to_string()
                            })
                        }))
                    }))
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 14),
                    errors: vec![]
                },
                expression: Expression::Unary(Box::new(UnaryExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 14),
                        errors: vec![]
                    },
                    operator: Operator::NotOperator,
                    argument: Expression::Paren(Box::new(ParenExpr {
                        base: BaseNode {
                            location: loc.get(1, 5, 1, 14),
                            errors: vec![]
                        },
                        expression: Expression::Logical(Box::new(LogicalExpr {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 13),
                                errors: vec![]
                            },
                            operator: LogicalOperator::AndOperator,
                            left: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 7),
                                    errors: vec![]
                                },
                                name: "a".to_string()
                            }),
                            right: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 12, 1, 13),
                                    errors: vec![]
                                },
                                name: "b".to_string()
                            })
                        }))
                    }))
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 15),
                    errors: vec![]
                },
                expression: Expression::Logical(Box::new(LogicalExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 15),
                        errors: vec![]
                    },
                    operator: LogicalOperator::AndOperator,
                    left: Expression::Paren(Box::new(ParenExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 9),
                            errors: vec![]
                        },
                        expression: Expression::Logical(Box::new(LogicalExpr {
                            base: BaseNode {
                                location: loc.get(1, 2, 1, 8),
                                errors: vec![]
                            },
                            operator: LogicalOperator::OrOperator,
                            left: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 2, 1, 3),
                                    errors: vec![]
                                },
                                name: "a".to_string()
                            }),
                            right: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 7, 1, 8),
                                    errors: vec![]
                                },
                                name: "b".to_string()
                            })
                        }))
                    })),
                    right: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 14, 1, 15),
                            errors: vec![]
                        },
                        name: "c".to_string()
                    })
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 15),
                    errors: vec![]
                },
                expression: Expression::Logical(Box::new(LogicalExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 15),
                        errors: vec![]
                    },
                    operator: LogicalOperator::AndOperator,
                    left: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 2),
                            errors: vec![]
                        },
                        name: "a".to_string()
                    }),
                    right: Expression::Paren(Box::new(ParenExpr {
                        base: BaseNode {
                            location: loc.get(1, 7, 1, 15),
                            errors: vec![]
                        },
                        expression: Expression::Logical(Box::new(LogicalExpr {
                            base: BaseNode {
                                location: loc.get(1, 8, 1, 14),
                                errors: vec![]
                            },
                            operator: LogicalOperator::OrOperator,
                            left: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 8, 1, 9),
                                    errors: vec![]
                                },
                                name: "b".to_string()
                            }),
                            right: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 13, 1, 14),
                                    errors: vec![]
                                },
                                name: "c".to_string()
                            })
                        }))
                    }))
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 2, 15),
                    errors: vec![]
                },
                expression: Expression::Logical(Box::new(LogicalExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 2, 15),
                        errors: vec![]
                    },
                    operator: LogicalOperator::AndOperator,
                    left: Expression::Unary(Box::new(UnaryExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 2, 9),
                            errors: vec![]
                        },
                        operator: Operator::NotOperator,
                        argument: Expression::Call(Box::new(CallExpr {
                            base: BaseNode {
                                location: loc.get(1, 5, 2, 9),
                                errors: vec!["expected comma in property list, got OR".to_string()]
                            },
                            callee: Expression::Paren(Box::new(ParenExpr {
                                base: BaseNode {
                                    location: loc.get(1, 5, 1, 14),
                                    errors: vec![]
                                },
                                expression: Expression::Logical(Box::new(LogicalExpr {
                                    base: BaseNode {
                                        location: loc.get(1, 6, 1, 13),
                                        errors: vec![]
                                    },
                                    operator: LogicalOperator::AndOperator,
                                    left: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(1, 6, 1, 7),
                                            errors: vec![]
                                        },
                                        name: "a".to_string()
                                    }),
                                    right: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(1, 12, 1, 13),
                                            errors: vec![]
                                        },
                                        name: "b".to_string()
                                    })
                                }))
                            })),
                            arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                base: BaseNode {
                                    location: loc.get(2, 2, 2, 8),
                                    errors: vec![],
                                },
                                with: None,
                                properties: vec![
                                    Property {
                                        base: BaseNode {
                                            location: loc.get(2, 2, 2, 3),
                                            errors: vec![]
                                        },
                                        key: PropertyKey::Identifier(Identifier {
                                            base: BaseNode {
                                                location: loc.get(2, 2, 2, 3),
                                                errors: vec![]
                                            },
                                            name: "a".to_string()
                                        }),
                                        value: None,
                                    },
                                    Property {
                                        base: BaseNode {
                                            location: loc.get(2, 4, 2, 8),
                                            errors: vec![
                                                "unexpected token for property key: OR (or)"
                                                    .to_string()
                                            ]
                                        },
                                        key: PropertyKey::StringLit(StringLit {
                                            base: BaseNode {
                                                location: loc.get(2, 4, 2, 4),
                                                errors: vec![]
                                            },
                                            value: "<invalid>".to_string()
                                        }),
                                        value: None,
                                    }
                                ]
                            }))]
                        }))
                    })),
                    right: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(2, 14, 2, 15),
                            errors: vec![]
                        },
                        name: "c".to_string()
                    })
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 23),
                        errors: vec![]
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 8),
                            errors: vec![]
                        },
                        name: "plusOne".to_string()
                    },
                    init: Expression::Function(Box::new(FunctionExpr {
                        base: BaseNode {
                            location: loc.get(1, 11, 1, 23),
                            errors: vec![]
                        },
                        params: vec![Property {
                            base: BaseNode {
                                location: loc.get(1, 12, 1, 13),
                                errors: vec![]
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 12, 1, 13),
                                    errors: vec![]
                                },
                                name: "r".to_string()
                            }),
                            value: None
                        }],
                        body: FunctionBody::Expr(Expression::Binary(Box::new(BinaryExpr {
                            base: BaseNode {
                                location: loc.get(1, 18, 1, 23),
                                errors: vec![]
                            },
                            operator: Operator::AdditionOperator,
                            left: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 18, 1, 19),
                                    errors: vec![]
                                },
                                name: "r".to_string()
                            }),
                            right: Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 22, 1, 23),
                                    errors: vec![]
                                },
                                value: 1
                            })
                        })))
                    }))
                })),
                Statement::Expr(ExprStmt {
                    base: BaseNode {
                        location: loc.get(2, 4, 2, 16),
                        errors: vec![]
                    },
                    expression: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(2, 4, 2, 16),
                            errors: vec![]
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(2, 4, 2, 11),
                                errors: vec![]
                            },
                            name: "plusOne".to_string()
                        }),
                        arguments: vec![Expression::Object(Box::new(ObjectExpr {
                            base: BaseNode {
                                location: loc.get(2, 12, 2, 15),
                                errors: vec![]
                            },
                            with: None,
                            properties: vec![Property {
                                base: BaseNode {
                                    location: loc.get(2, 12, 2, 15),
                                    errors: vec![]
                                },
                                key: PropertyKey::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(2, 12, 2, 13),
                                        errors: vec![]
                                    },
                                    name: "r".to_string()
                                }),
                                value: Some(Expression::Integer(IntegerLit {
                                    base: BaseNode {
                                        location: loc.get(2, 14, 2, 15),
                                        errors: vec![]
                                    },
                                    value: 5
                                }))
                            }]
                        }))]
                    }))
                })
            ]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 22),
                    errors: vec![]
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 6),
                        errors: vec![]
                    },
                    name: "toMap".to_string()
                },
                init: Expression::Function(Box::new(FunctionExpr {
                    base: BaseNode {
                        location: loc.get(1, 9, 1, 22),
                        errors: vec![]
                    },
                    params: vec![Property {
                        base: BaseNode {
                            location: loc.get(1, 10, 1, 11),
                            errors: vec![]
                        },
                        key: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 10, 1, 11),
                                errors: vec![]
                            },
                            name: "r".to_string()
                        }),
                        value: None
                    }],
                    body: FunctionBody::Expr(Expression::Paren(Box::new(ParenExpr {
                        base: BaseNode {
                            location: loc.get(1, 15, 1, 22),
                            errors: vec![]
                        },
                        expression: Expression::Object(Box::new(ObjectExpr {
                            base: BaseNode {
                                location: loc.get(1, 16, 1, 21),
                                errors: vec![]
                            },
                            with: None,
                            properties: vec![Property {
                                base: BaseNode {
                                    location: loc.get(1, 17, 1, 20),
                                    errors: vec![]
                                },
                                key: PropertyKey::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 17, 1, 18),
                                        errors: vec![]
                                    },
                                    name: "r".to_string()
                                }),
                                value: Some(Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 19, 1, 20),
                                        errors: vec![]
                                    },
                                    name: "r".to_string()
                                }))
                            }]
                        }))
                    })))
                }))
            }))]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 25),
                    errors: vec![]
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 5),
                        errors: vec![]
                    },
                    name: "addN".to_string()
                },
                init: Expression::Function(Box::new(FunctionExpr {
                    base: BaseNode {
                        location: loc.get(1, 8, 1, 25),
                        errors: vec![]
                    },
                    params: vec![
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 9, 1, 10),
                                errors: vec![]
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 9, 1, 10),
                                    errors: vec![]
                                },
                                name: "r".to_string()
                            }),
                            value: None
                        },
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 12, 1, 15),
                                errors: vec![]
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 12, 1, 13),
                                    errors: vec![]
                                },
                                name: "n".to_string()
                            }),
                            value: Some(Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 14, 1, 15),
                                    errors: vec![]
                                },
                                value: 5
                            }))
                        }
                    ],
                    body: FunctionBody::Expr(Expression::Binary(Box::new(BinaryExpr {
                        base: BaseNode {
                            location: loc.get(1, 20, 1, 25),
                            errors: vec![]
                        },
                        operator: Operator::AdditionOperator,
                        left: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 20, 1, 21),
                                errors: vec![]
                            },
                            name: "r".to_string()
                        }),
                        right: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 24, 1, 25),
                                errors: vec![]
                            },
                            name: "n".to_string()
                        })
                    })))
                }))
            }))]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(2, 13, 2, 35),
                        errors: vec![]
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(2, 13, 2, 20),
                            errors: vec![]
                        },
                        name: "plusOne".to_string()
                    },
                    init: Expression::Function(Box::new(FunctionExpr {
                        base: BaseNode {
                            location: loc.get(2, 23, 2, 35),
                            errors: vec![]
                        },
                        params: vec![Property {
                            base: BaseNode {
                                location: loc.get(2, 24, 2, 25),
                                errors: vec![]
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(2, 24, 2, 25),
                                    errors: vec![]
                                },
                                name: "r".to_string()
                            }),
                            value: None
                        }],
                        body: FunctionBody::Expr(Expression::Binary(Box::new(BinaryExpr {
                            base: BaseNode {
                                location: loc.get(2, 30, 2, 35),
                                errors: vec![]
                            },
                            operator: Operator::AdditionOperator,
                            left: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(2, 30, 2, 31),
                                    errors: vec![]
                                },
                                name: "r".to_string()
                            }),
                            right: Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(2, 34, 2, 35),
                                    errors: vec![]
                                },
                                value: 1
                            })
                        })))
                    }))
                })),
                Statement::Expr(ExprStmt {
                    base: BaseNode {
                        location: loc.get(3, 13, 3, 39),
                        errors: vec![]
                    },
                    expression: Expression::Logical(Box::new(LogicalExpr {
                        base: BaseNode {
                            location: loc.get(3, 13, 3, 39),
                            errors: vec![]
                        },
                        operator: LogicalOperator::OrOperator,
                        left: Expression::Binary(Box::new(BinaryExpr {
                            base: BaseNode {
                                location: loc.get(3, 13, 3, 30),
                                errors: vec![]
                            },
                            operator: Operator::EqualOperator,
                            left: Expression::Call(Box::new(CallExpr {
                                base: BaseNode {
                                    location: loc.get(3, 13, 3, 25),
                                    errors: vec![]
                                },
                                callee: Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(3, 13, 3, 20),
                                        errors: vec![]
                                    },
                                    name: "plusOne".to_string()
                                }),
                                arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                    base: BaseNode {
                                        location: loc.get(3, 21, 3, 24),
                                        errors: vec![]
                                    },
                                    with: None,
                                    properties: vec![Property {
                                        base: BaseNode {
                                            location: loc.get(3, 21, 3, 24),
                                            errors: vec![]
                                        },
                                        key: PropertyKey::Identifier(Identifier {
                                            base: BaseNode {
                                                location: loc.get(3, 21, 3, 22),
                                                errors: vec![]
                                            },
                                            name: "r".to_string()
                                        }),
                                        value: Some(Expression::Integer(IntegerLit {
                                            base: BaseNode {
                                                location: loc.get(3, 23, 3, 24),
                                                errors: vec![]
                                            },
                                            value: 5
                                        }))
                                    }]
                                }))]
                            })),
                            right: Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(3, 29, 3, 30),
                                    errors: vec![]
                                },
                                value: 6
                            })
                        })),
                        right: Expression::Call(Box::new(CallExpr {
                            base: BaseNode {
                                location: loc.get(3, 34, 3, 39),
                                errors: vec![]
                            },
                            arguments: vec![],
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(3, 34, 3, 37),
                                    errors: vec![]
                                },
                                name: "die".to_string()
                            })
                        }))
                    }))
                })
            ]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 38),
                    errors: vec![]
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 2),
                        errors: vec![]
                    },
                    name: "f".to_string()
                },
                init: Expression::Function(Box::new(FunctionExpr {
                    base: BaseNode {
                        location: loc.get(1, 5, 1, 38),
                        errors: vec![]
                    },
                    params: vec![Property {
                        base: BaseNode {
                            location: loc.get(1, 6, 1, 7),
                            errors: vec![]
                        },
                        key: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 7),
                                errors: vec![]
                            },
                            name: "r".to_string()
                        }),
                        value: None
                    }],
                    body: FunctionBody::Expr(Expression::Binary(Box::new(BinaryExpr {
                        base: BaseNode {
                            location: loc.get(1, 12, 1, 38),
                            errors: vec![]
                        },
                        operator: Operator::EqualOperator,
                        left: Expression::Member(Box::new(MemberExpr {
                            base: BaseNode {
                                location: loc.get(1, 12, 1, 29),
                                errors: vec![]
                            },
                            object: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 12, 1, 13),
                                    errors: vec![]
                                },
                                name: "r".to_string()
                            }),
                            property: PropertyKey::StringLit(StringLit {
                                base: BaseNode {
                                    location: loc.get(1, 14, 1, 28),
                                    errors: vec![]
                                },
                                value: "_measurement".to_string()
                            })
                        })),
                        right: Expression::StringLit(StringLit {
                            base: BaseNode {
                                location: loc.get(1, 33, 1, 38),
                                errors: vec![]
                            },
                            value: "cpu".to_string()
                        })
                    })))
                }))
            }))]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 4, 14),
                    errors: vec![]
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 2),
                        errors: vec![]
                    },
                    name: "f".to_string()
                },
                init: Expression::Function(Box::new(FunctionExpr {
                    base: BaseNode {
                        location: loc.get(1, 5, 4, 14),
                        errors: vec![]
                    },
                    params: vec![Property {
                        base: BaseNode {
                            location: loc.get(1, 6, 1, 7),
                            errors: vec![]
                        },
                        key: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 7),
                                errors: vec![]
                            },
                            name: "r".to_string()
                        }),
                        value: None
                    }],
                    body: FunctionBody::Block(Block {
                        base: BaseNode {
                            location: loc.get(1, 12, 4, 14),
                            errors: vec![]
                        },
                        body: vec![
                            Statement::Variable(Box::new(VariableAssgn {
                                base: BaseNode {
                                    location: loc.get(2, 17, 2, 38),
                                    errors: vec![]
                                },
                                id: Identifier {
                                    base: BaseNode {
                                        location: loc.get(2, 17, 2, 18),
                                        errors: vec![]
                                    },
                                    name: "m".to_string()
                                },
                                init: Expression::Member(Box::new(MemberExpr {
                                    base: BaseNode {
                                        location: loc.get(2, 21, 2, 38),
                                        errors: vec![]
                                    },
                                    object: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(2, 21, 2, 22),
                                            errors: vec![]
                                        },
                                        name: "r".to_string()
                                    }),
                                    property: PropertyKey::StringLit(StringLit {
                                        base: BaseNode {
                                            location: loc.get(2, 23, 2, 37),
                                            errors: vec![]
                                        },
                                        value: "_measurement".to_string()
                                    })
                                }))
                            })),
                            Statement::Return(ReturnStmt {
                                base: BaseNode {
                                    location: loc.get(3, 17, 3, 34),
                                    errors: vec![]
                                },
                                argument: Expression::Binary(Box::new(BinaryExpr {
                                    base: BaseNode {
                                        location: loc.get(3, 24, 3, 34),
                                        errors: vec![]
                                    },
                                    operator: Operator::EqualOperator,
                                    left: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(3, 24, 3, 25),
                                            errors: vec![]
                                        },
                                        name: "m".to_string()
                                    }),
                                    right: Expression::StringLit(StringLit {
                                        base: BaseNode {
                                            location: loc.get(3, 29, 3, 34),
                                            errors: vec![]
                                        },
                                        value: "cpu".to_string()
                                    })
                                }))
                            })
                        ]
                    })
                }))
            }))]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 26),
                    errors: vec![]
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 2),
                        errors: vec![]
                    },
                    name: "a".to_string()
                },
                init: Expression::Conditional(Box::new(ConditionalExpr {
                    base: BaseNode {
                        location: loc.get(1, 5, 1, 26),
                        errors: vec![]
                    },
                    test: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 8, 1, 12),
                            errors: vec![]
                        },
                        name: "true".to_string()
                    }),
                    consequent: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 18, 1, 19),
                            errors: vec![]
                        },
                        value: 0
                    }),
                    alternate: Expression::Integer(IntegerLit {
                        base: BaseNode {
                            location: loc.get(1, 25, 1, 26),
                            errors: vec![]
                        },
                        value: 1
                    })
                }))
            }))]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 85),
                    errors: vec![]
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 2),
                        errors: vec![]
                    },
                    name: "a".to_string()
                },
                init: Expression::Conditional(Box::new(ConditionalExpr {
                    base: BaseNode {
                        location: loc.get(1, 5, 1, 85),
                        errors: vec![]
                    },
                    test: Expression::Logical(Box::new(LogicalExpr {
                        base: BaseNode {
                            location: loc.get(1, 8, 1, 40),
                            errors: vec![]
                        },
                        operator: LogicalOperator::OrOperator,
                        left: Expression::Unary(Box::new(UnaryExpr {
                            base: BaseNode {
                                location: loc.get(1, 8, 1, 16),
                                errors: vec![]
                            },
                            operator: Operator::ExistsOperator,
                            argument: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 15, 1, 16),
                                    errors: vec![]
                                },
                                name: "b".to_string()
                            })
                        })),
                        right: Expression::Logical(Box::new(LogicalExpr {
                            base: BaseNode {
                                location: loc.get(1, 20, 1, 40),
                                errors: vec![]
                            },
                            operator: LogicalOperator::AndOperator,
                            left: Expression::Binary(Box::new(BinaryExpr {
                                base: BaseNode {
                                    location: loc.get(1, 20, 1, 25),
                                    errors: vec![]
                                },
                                operator: Operator::LessThanOperator,
                                left: Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 20, 1, 21),
                                        errors: vec![]
                                    },
                                    name: "c".to_string()
                                }),
                                right: Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 24, 1, 25),
                                        errors: vec![]
                                    },
                                    name: "d".to_string()
                                })
                            })),
                            right: Expression::Unary(Box::new(UnaryExpr {
                                base: BaseNode {
                                    location: loc.get(1, 30, 1, 40),
                                    errors: vec![]
                                },
                                operator: Operator::NotOperator,
                                argument: Expression::Binary(Box::new(BinaryExpr {
                                    base: BaseNode {
                                        location: loc.get(1, 34, 1, 40),
                                        errors: vec![]
                                    },
                                    operator: Operator::EqualOperator,
                                    left: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(1, 34, 1, 35),
                                            errors: vec![]
                                        },
                                        name: "e".to_string()
                                    }),
                                    right: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(1, 39, 1, 40),
                                            errors: vec![]
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
                            errors: vec![]
                        },
                        operator: Operator::NotOperator,
                        argument: Expression::Unary(Box::new(UnaryExpr {
                            base: BaseNode {
                                location: loc.get(1, 50, 1, 64),
                                errors: vec![]
                            },
                            operator: Operator::ExistsOperator,
                            argument: Expression::Paren(Box::new(ParenExpr {
                                base: BaseNode {
                                    location: loc.get(1, 57, 1, 64),
                                    errors: vec![]
                                },
                                expression: Expression::Binary(Box::new(BinaryExpr {
                                    base: BaseNode {
                                        location: loc.get(1, 58, 1, 63),
                                        errors: vec![]
                                    },
                                    operator: Operator::SubtractionOperator,
                                    left: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(1, 58, 1, 59),
                                            errors: vec![]
                                        },
                                        name: "g".to_string()
                                    }),
                                    right: Expression::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(1, 62, 1, 63),
                                            errors: vec![]
                                        },
                                        name: "h".to_string()
                                    })
                                }))
                            }))
                        }))
                    })),
                    alternate: Expression::Unary(Box::new(UnaryExpr {
                        base: BaseNode {
                            location: loc.get(1, 70, 1, 85),
                            errors: vec![]
                        },
                        operator: Operator::ExistsOperator,
                        argument: Expression::Unary(Box::new(UnaryExpr {
                            base: BaseNode {
                                location: loc.get(1, 77, 1, 85),
                                errors: vec![]
                            },
                            operator: Operator::ExistsOperator,
                            argument: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 84, 1, 85),
                                    errors: vec![]
                                },
                                name: "i".to_string()
                            })
                        }))
                    }))
                }))
            }))]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 3, 50),
                    errors: vec![]
                },
                expression: Expression::Conditional(Box::new(ConditionalExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 3, 50),
                        errors: vec![]
                    },
                    test: Expression::Conditional(Box::new(ConditionalExpr {
                        base: BaseNode {
                            location: loc.get(1, 4, 1, 33),
                            errors: vec![]
                        },
                        test: Expression::Binary(Box::new(BinaryExpr {
                            base: BaseNode {
                                location: loc.get(1, 7, 1, 12),
                                errors: vec![]
                            },
                            operator: Operator::LessThanOperator,
                            left: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 7, 1, 8),
                                    errors: vec![]
                                },
                                name: "b".to_string()
                            }),
                            right: Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 11, 1, 12),
                                    errors: vec![]
                                },
                                value: 0
                            })
                        })),
                        consequent: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 18, 1, 22),
                                errors: vec![]
                            },
                            name: "true".to_string()
                        }),
                        alternate: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 28, 1, 33),
                                errors: vec![]
                            },
                            name: "false".to_string()
                        })
                    })),
                    consequent: Expression::Conditional(Box::new(ConditionalExpr {
                        base: BaseNode {
                            location: loc.get(2, 24, 2, 48),
                            errors: vec![]
                        },
                        test: Expression::Binary(Box::new(BinaryExpr {
                            base: BaseNode {
                                location: loc.get(2, 27, 2, 32),
                                errors: vec![]
                            },
                            operator: Operator::GreaterThanOperator,
                            left: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(2, 27, 2, 28),
                                    errors: vec![]
                                },
                                name: "c".to_string()
                            }),
                            right: Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(2, 31, 2, 32),
                                    errors: vec![]
                                },
                                value: 0
                            })
                        })),
                        consequent: Expression::Integer(IntegerLit {
                            base: BaseNode {
                                location: loc.get(2, 38, 2, 40),
                                errors: vec![]
                            },
                            value: 30
                        }),
                        alternate: Expression::Integer(IntegerLit {
                            base: BaseNode {
                                location: loc.get(2, 46, 2, 48),
                                errors: vec![]
                            },
                            value: 60
                        })
                    })),
                    alternate: Expression::Conditional(Box::new(ConditionalExpr {
                        base: BaseNode {
                            location: loc.get(3, 24, 3, 50),
                            errors: vec![]
                        },
                        test: Expression::Binary(Box::new(BinaryExpr {
                            base: BaseNode {
                                location: loc.get(3, 27, 3, 33),
                                errors: vec![]
                            },
                            operator: Operator::EqualOperator,
                            left: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(3, 27, 3, 28),
                                    errors: vec![]
                                },
                                name: "d".to_string()
                            }),
                            right: Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(3, 32, 3, 33),
                                    errors: vec![]
                                },
                                value: 0
                            })
                        })),
                        consequent: Expression::Integer(IntegerLit {
                            base: BaseNode {
                                location: loc.get(3, 39, 3, 41),
                                errors: vec![]
                            },
                            value: 90
                        }),
                        alternate: Expression::Integer(IntegerLit {
                            base: BaseNode {
                                location: loc.get(3, 47, 3, 50),
                                errors: vec![]
                            },
                            value: 120
                        })
                    }))
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 114),
                    errors: vec![]
                },
                expression: Expression::Call(Box::new(CallExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 114),
                        errors: vec![]
                    },
                    callee: Expression::Member(Box::new(MemberExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 39),
                            errors: vec![]
                        },
                        property: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 33, 1, 39),
                                errors: vec![]
                            },
                            name: "filter".to_string()
                        }),
                        object: Expression::Call(Box::new(CallExpr {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 32),
                                errors: vec![]
                            },
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 1, 1, 5),
                                    errors: vec![]
                                },
                                name: "from".to_string()
                            }),
                            arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 31),
                                    errors: vec![]
                                },
                                with: None,
                                properties: vec![Property {
                                    base: BaseNode {
                                        location: loc.get(1, 6, 1, 31),
                                        errors: vec![]
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(1, 6, 1, 12),
                                            errors: vec![]
                                        },
                                        name: "bucket".to_string()
                                    }),
                                    value: Some(Expression::StringLit(StringLit {
                                        base: BaseNode {
                                            location: loc.get(1, 13, 1, 31),
                                            errors: vec![]
                                        },
                                        value: "telegraf/autogen".to_string()
                                    }))
                                }]
                            }))]
                        })),
                    })),
                    arguments: vec![Expression::Object(Box::new(ObjectExpr {
                        base: BaseNode {
                            location: loc.get(1, 40, 1, 113),
                            errors: vec![]
                        },
                        with: None,
                        properties: vec![Property {
                            base: BaseNode {
                                location: loc.get(1, 40, 1, 113),
                                errors: vec![]
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 40, 1, 42),
                                    errors: vec![]
                                },
                                name: "fn".to_string()
                            }),
                            value: Some(Expression::Function(Box::new(FunctionExpr {
                                base: BaseNode {
                                    location: loc.get(1, 44, 1, 113),
                                    errors: vec![]
                                },
                                params: vec![Property {
                                    base: BaseNode {
                                        location: loc.get(1, 45, 1, 46),
                                        errors: vec![]
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(1, 45, 1, 46),
                                            errors: vec![]
                                        },
                                        name: "r".to_string()
                                    }),
                                    value: None
                                }],
                                body: FunctionBody::Expr(Expression::Logical(Box::new(
                                    LogicalExpr {
                                        base: BaseNode {
                                            location: loc.get(1, 51, 1, 113),
                                            errors: vec![]
                                        },
                                        operator: LogicalOperator::OrOperator,
                                        left: Expression::Logical(Box::new(LogicalExpr {
                                            base: BaseNode {
                                                location: loc.get(1, 51, 1, 90),
                                                errors: vec![]
                                            },
                                            operator: LogicalOperator::AndOperator,
                                            left: Expression::Binary(Box::new(BinaryExpr {
                                                base: BaseNode {
                                                    location: loc.get(1, 51, 1, 68),
                                                    errors: vec![]
                                                },
                                                operator: Operator::EqualOperator,
                                                left: Expression::Member(Box::new(MemberExpr {
                                                    base: BaseNode {
                                                        location: loc.get(1, 51, 1, 61),
                                                        errors: vec![]
                                                    },
                                                    object: Expression::Identifier(Identifier {
                                                        base: BaseNode {
                                                            location: loc.get(1, 51, 1, 52),
                                                            errors: vec![]
                                                        },
                                                        name: "r".to_string()
                                                    }),
                                                    property: PropertyKey::StringLit(StringLit {
                                                        base: BaseNode {
                                                            location: loc.get(1, 53, 1, 60),
                                                            errors: vec![]
                                                        },
                                                        value: "other".to_string()
                                                    })
                                                })),
                                                right: Expression::StringLit(StringLit {
                                                    base: BaseNode {
                                                        location: loc.get(1, 63, 1, 68),
                                                        errors: vec![]
                                                    },
                                                    value: "mem".to_string()
                                                })
                                            })),
                                            right: Expression::Binary(Box::new(BinaryExpr {
                                                base: BaseNode {
                                                    location: loc.get(1, 73, 1, 90),
                                                    errors: vec![]
                                                },
                                                operator: Operator::EqualOperator,
                                                left: Expression::Member(Box::new(MemberExpr {
                                                    base: BaseNode {
                                                        location: loc.get(1, 73, 1, 82),
                                                        errors: vec![]
                                                    },
                                                    object: Expression::Identifier(Identifier {
                                                        base: BaseNode {
                                                            location: loc.get(1, 73, 1, 74),
                                                            errors: vec![]
                                                        },
                                                        name: "r".to_string()
                                                    }),
                                                    property: PropertyKey::StringLit(StringLit {
                                                        base: BaseNode {
                                                            location: loc.get(1, 75, 1, 81),
                                                            errors: vec![]
                                                        },
                                                        value: "this".to_string()
                                                    })
                                                })),
                                                right: Expression::StringLit(StringLit {
                                                    base: BaseNode {
                                                        location: loc.get(1, 84, 1, 90),
                                                        errors: vec![]
                                                    },
                                                    value: "that".to_string()
                                                })
                                            }))
                                        })),
                                        right: Expression::Binary(Box::new(BinaryExpr {
                                            base: BaseNode {
                                                location: loc.get(1, 94, 1, 113),
                                                errors: vec![]
                                            },
                                            operator: Operator::NotEqualOperator,
                                            left: Expression::Member(Box::new(MemberExpr {
                                                base: BaseNode {
                                                    location: loc.get(1, 94, 1, 104),
                                                    errors: vec![]
                                                },
                                                object: Expression::Identifier(Identifier {
                                                    base: BaseNode {
                                                        location: loc.get(1, 94, 1, 95),
                                                        errors: vec![]
                                                    },
                                                    name: "r".to_string()
                                                }),
                                                property: PropertyKey::StringLit(StringLit {
                                                    base: BaseNode {
                                                        location: loc.get(1, 96, 1, 103),
                                                        errors: vec![]
                                                    },
                                                    value: "these".to_string()
                                                })
                                            })),
                                            right: Expression::StringLit(StringLit {
                                                base: BaseNode {
                                                    location: loc.get(1, 106, 1, 113),
                                                    errors: vec![]
                                                },
                                                value: "those".to_string()
                                            })
                                        }))
                                    }
                                )))
                            })))
                        }]
                    }))]
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 59),
                    errors: vec![]
                },
                expression: Expression::PipeExpr(Box::new(PipeExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 59),
                        errors: vec![]
                    },
                    argument: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 32),
                            errors: vec![]
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 5),
                                errors: vec![]
                            },
                            name: "from".to_string()
                        }),
                        arguments: vec![Expression::Object(Box::new(ObjectExpr {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 31),
                                errors: vec![]
                            },
                            with: None,
                            properties: vec![Property {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 31),
                                    errors: vec![]
                                },
                                key: PropertyKey::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 6, 1, 12),
                                        errors: vec![]
                                    },
                                    name: "bucket".to_string()
                                }),
                                value: Some(Expression::StringLit(StringLit {
                                    base: BaseNode {
                                        location: loc.get(1, 13, 1, 31),
                                        errors: vec![]
                                    },
                                    value: "telegraf/autogen".to_string()
                                }))
                            }]
                        }))]
                    })),
                    call: CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 34, 1, 59),
                            errors: vec![]
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 34, 1, 39),
                                errors: vec![]
                            },
                            name: "range".to_string()
                        }),
                        arguments: vec![Expression::Object(Box::new(ObjectExpr {
                            base: BaseNode {
                                location: loc.get(1, 40, 1, 58),
                                errors: vec![]
                            },
                            with: None,
                            properties: vec![
                                Property {
                                    base: BaseNode {
                                        location: loc.get(1, 40, 1, 49),
                                        errors: vec![]
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(1, 40, 1, 45),
                                            errors: vec![]
                                        },
                                        name: "start".to_string()
                                    }),
                                    value: Some(Expression::Unary(Box::new(UnaryExpr {
                                        base: BaseNode {
                                            location: loc.get(1, 46, 1, 49),
                                            errors: vec![]
                                        },
                                        operator: Operator::SubtractionOperator,
                                        argument: Expression::Duration(DurationLit {
                                            base: BaseNode {
                                                location: loc.get(1, 47, 1, 49),
                                                errors: vec![]
                                            },
                                            values: vec![Duration {
                                                magnitude: 1,
                                                unit: "h".to_string()
                                            }]
                                        })
                                    })))
                                },
                                Property {
                                    base: BaseNode {
                                        location: loc.get(1, 51, 1, 58),
                                        errors: vec![]
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(1, 51, 1, 54),
                                            errors: vec![]
                                        },
                                        name: "end".to_string()
                                    }),
                                    value: Some(Expression::Duration(DurationLit {
                                        base: BaseNode {
                                            location: loc.get(1, 55, 1, 58),
                                            errors: vec![]
                                        },
                                        values: vec![Duration {
                                            magnitude: 10,
                                            unit: "m".to_string()
                                        }]
                                    }))
                                }
                            ]
                        }))]
                    }
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 61),
                    errors: vec![]
                },
                expression: Expression::PipeExpr(Box::new(PipeExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 61),
                        errors: vec![]
                    },
                    argument: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 32),
                            errors: vec![]
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 5),
                                errors: vec![]
                            },
                            name: "from".to_string()
                        }),
                        arguments: vec![Expression::Object(Box::new(ObjectExpr {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 31),
                                errors: vec![]
                            },
                            with: None,
                            properties: vec![Property {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 31),
                                    errors: vec![]
                                },
                                key: PropertyKey::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 6, 1, 12),
                                        errors: vec![]
                                    },
                                    name: "bucket".to_string()
                                }),
                                value: Some(Expression::StringLit(StringLit {
                                    base: BaseNode {
                                        location: loc.get(1, 13, 1, 31),
                                        errors: vec![]
                                    },
                                    value: "telegraf/autogen".to_string()
                                }))
                            }]
                        }))]
                    })),
                    call: CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 34, 1, 61),
                            errors: vec![]
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 34, 1, 39),
                                errors: vec![]
                            },
                            name: "limit".to_string()
                        }),
                        arguments: vec![Expression::Object(Box::new(ObjectExpr {
                            base: BaseNode {
                                location: loc.get(1, 40, 1, 60),
                                errors: vec![]
                            },
                            with: None,
                            properties: vec![
                                Property {
                                    base: BaseNode {
                                        location: loc.get(1, 40, 1, 49),
                                        errors: vec![]
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(1, 40, 1, 45),
                                            errors: vec![]
                                        },
                                        name: "limit".to_string()
                                    }),
                                    value: Some(Expression::Integer(IntegerLit {
                                        base: BaseNode {
                                            location: loc.get(1, 46, 1, 49),
                                            errors: vec![]
                                        },
                                        value: 100
                                    }))
                                },
                                Property {
                                    base: BaseNode {
                                        location: loc.get(1, 51, 1, 60),
                                        errors: vec![]
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(1, 51, 1, 57),
                                            errors: vec![]
                                        },
                                        name: "offset".to_string()
                                    }),
                                    value: Some(Expression::Integer(IntegerLit {
                                        base: BaseNode {
                                            location: loc.get(1, 58, 1, 60),
                                            errors: vec![]
                                        },
                                        value: 10
                                    }))
                                }
                            ]
                        }))]
                    }
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 3, 17),
                    errors: vec![]
                },
                expression: Expression::PipeExpr(Box::new(PipeExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 3, 17),
                        errors: vec![]
                    },
                    argument: Expression::PipeExpr(Box::new(PipeExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 2, 36),
                            errors: vec![]
                        },
                        argument: Expression::Call(Box::new(CallExpr {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 28),
                                errors: vec![]
                            },
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 1, 1, 5),
                                    errors: vec![]
                                },
                                name: "from".to_string()
                            }),
                            arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 27),
                                    errors: vec![]
                                },
                                with: None,
                                properties: vec![Property {
                                    base: BaseNode {
                                        location: loc.get(1, 6, 1, 27),
                                        errors: vec![]
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(1, 6, 1, 12),
                                            errors: vec![]
                                        },
                                        name: "bucket".to_string()
                                    }),
                                    value: Some(Expression::StringLit(StringLit {
                                        base: BaseNode {
                                            location: loc.get(1, 13, 1, 27),
                                            errors: vec![]
                                        },
                                        value: "mydb/autogen".to_string()
                                    }))
                                }]
                            }))]
                        })),
                        call: CallExpr {
                            base: BaseNode {
                                location: loc.get(2, 10, 2, 36),
                                errors: vec![]
                            },
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(2, 10, 2, 15),
                                    errors: vec![]
                                },
                                name: "range".to_string()
                            }),
                            arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                base: BaseNode {
                                    location: loc.get(2, 16, 2, 35),
                                    errors: vec![]
                                },
                                with: None,
                                properties: vec![
                                    Property {
                                        base: BaseNode {
                                            location: loc.get(2, 16, 2, 25),
                                            errors: vec![]
                                        },
                                        key: PropertyKey::Identifier(Identifier {
                                            base: BaseNode {
                                                location: loc.get(2, 16, 2, 21),
                                                errors: vec![]
                                            },
                                            name: "start".to_string()
                                        }),
                                        value: Some(Expression::Unary(Box::new(UnaryExpr {
                                            base: BaseNode {
                                                location: loc.get(2, 22, 2, 25),
                                                errors: vec![]
                                            },
                                            operator: Operator::SubtractionOperator,
                                            argument: Expression::Duration(DurationLit {
                                                base: BaseNode {
                                                    location: loc.get(2, 23, 2, 25),
                                                    errors: vec![]
                                                },
                                                values: vec![Duration {
                                                    magnitude: 4,
                                                    unit: "h".to_string()
                                                }]
                                            })
                                        })))
                                    },
                                    Property {
                                        base: BaseNode {
                                            location: loc.get(2, 27, 2, 35),
                                            errors: vec![]
                                        },
                                        key: PropertyKey::Identifier(Identifier {
                                            base: BaseNode {
                                                location: loc.get(2, 27, 2, 31),
                                                errors: vec![]
                                            },
                                            name: "stop".to_string()
                                        }),
                                        value: Some(Expression::Unary(Box::new(UnaryExpr {
                                            base: BaseNode {
                                                location: loc.get(2, 32, 2, 35),
                                                errors: vec![]
                                            },
                                            operator: Operator::SubtractionOperator,
                                            argument: Expression::Duration(DurationLit {
                                                base: BaseNode {
                                                    location: loc.get(2, 33, 2, 35),
                                                    errors: vec![]
                                                },
                                                values: vec![Duration {
                                                    magnitude: 2,
                                                    unit: "h".to_string()
                                                }]
                                            })
                                        })))
                                    }
                                ]
                            }))]
                        }
                    })),
                    call: CallExpr {
                        base: BaseNode {
                            location: loc.get(3, 10, 3, 17),
                            errors: vec![]
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(3, 10, 3, 15),
                                errors: vec![]
                            },
                            name: "count".to_string()
                        }),
                        arguments: vec![]
                    }
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 4, 17),
                    errors: vec![]
                },
                expression: Expression::PipeExpr(Box::new(PipeExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 4, 17),
                        errors: vec![]
                    },
                    argument: Expression::PipeExpr(Box::new(PipeExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 3, 21),
                            errors: vec![]
                        },
                        argument: Expression::PipeExpr(Box::new(PipeExpr {
                            base: BaseNode {
                                location: loc.get(1, 1, 2, 36),
                                errors: vec![]
                            },
                            argument: Expression::Call(Box::new(CallExpr {
                                base: BaseNode {
                                    location: loc.get(1, 1, 1, 28),
                                    errors: vec![]
                                },
                                callee: Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 1, 1, 5),
                                        errors: vec![]
                                    },
                                    name: "from".to_string()
                                }),
                                arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                    base: BaseNode {
                                        location: loc.get(1, 6, 1, 27),
                                        errors: vec![]
                                    },
                                    with: None,
                                    properties: vec![Property {
                                        base: BaseNode {
                                            location: loc.get(1, 6, 1, 27),
                                            errors: vec![]
                                        },
                                        key: PropertyKey::Identifier(Identifier {
                                            base: BaseNode {
                                                location: loc.get(1, 6, 1, 12),
                                                errors: vec![]
                                            },
                                            name: "bucket".to_string()
                                        }),
                                        value: Some(Expression::StringLit(StringLit {
                                            base: BaseNode {
                                                location: loc.get(1, 13, 1, 27),
                                                errors: vec![]
                                            },
                                            value: "mydb/autogen".to_string()
                                        }))
                                    }]
                                }))]
                            })),
                            call: CallExpr {
                                base: BaseNode {
                                    location: loc.get(2, 10, 2, 36),
                                    errors: vec![]
                                },
                                callee: Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(2, 10, 2, 15),
                                        errors: vec![]
                                    },
                                    name: "range".to_string()
                                }),
                                arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                    base: BaseNode {
                                        location: loc.get(2, 16, 2, 35),
                                        errors: vec![]
                                    },
                                    with: None,
                                    properties: vec![
                                        Property {
                                            base: BaseNode {
                                                location: loc.get(2, 16, 2, 25),
                                                errors: vec![]
                                            },
                                            key: PropertyKey::Identifier(Identifier {
                                                base: BaseNode {
                                                    location: loc.get(2, 16, 2, 21),
                                                    errors: vec![]
                                                },
                                                name: "start".to_string()
                                            }),
                                            value: Some(Expression::Unary(Box::new(UnaryExpr {
                                                base: BaseNode {
                                                    location: loc.get(2, 22, 2, 25),
                                                    errors: vec![]
                                                },
                                                operator: Operator::SubtractionOperator,
                                                argument: Expression::Duration(DurationLit {
                                                    base: BaseNode {
                                                        location: loc.get(2, 23, 2, 25),
                                                        errors: vec![]
                                                    },
                                                    values: vec![Duration {
                                                        magnitude: 4,
                                                        unit: "h".to_string()
                                                    }]
                                                })
                                            })))
                                        },
                                        Property {
                                            base: BaseNode {
                                                location: loc.get(2, 27, 2, 35),
                                                errors: vec![]
                                            },
                                            key: PropertyKey::Identifier(Identifier {
                                                base: BaseNode {
                                                    location: loc.get(2, 27, 2, 31),
                                                    errors: vec![]
                                                },
                                                name: "stop".to_string()
                                            }),
                                            value: Some(Expression::Unary(Box::new(UnaryExpr {
                                                base: BaseNode {
                                                    location: loc.get(2, 32, 2, 35),
                                                    errors: vec![]
                                                },
                                                operator: Operator::SubtractionOperator,
                                                argument: Expression::Duration(DurationLit {
                                                    base: BaseNode {
                                                        location: loc.get(2, 33, 2, 35),
                                                        errors: vec![]
                                                    },
                                                    values: vec![Duration {
                                                        magnitude: 2,
                                                        unit: "h".to_string()
                                                    }]
                                                })
                                            })))
                                        }
                                    ]
                                }))]
                            }
                        })),
                        call: CallExpr {
                            base: BaseNode {
                                location: loc.get(3, 10, 3, 21),
                                errors: vec![]
                            },
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(3, 10, 3, 15),
                                    errors: vec![]
                                },
                                name: "limit".to_string()
                            }),
                            arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                base: BaseNode {
                                    location: loc.get(3, 16, 3, 20),
                                    errors: vec![]
                                },
                                with: None,
                                properties: vec![Property {
                                    base: BaseNode {
                                        location: loc.get(3, 16, 3, 20),
                                        errors: vec![]
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(3, 16, 3, 17),
                                            errors: vec![]
                                        },
                                        name: "n".to_string()
                                    }),
                                    value: Some(Expression::Integer(IntegerLit {
                                        base: BaseNode {
                                            location: loc.get(3, 18, 3, 20),
                                            errors: vec![]
                                        },
                                        value: 10
                                    }))
                                }]
                            }))]
                        }
                    })),
                    call: CallExpr {
                        base: BaseNode {
                            location: loc.get(4, 10, 4, 17),
                            errors: vec![]
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(4, 10, 4, 15),
                                errors: vec![]
                            },
                            name: "count".to_string()
                        }),
                        arguments: vec![]
                    }
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(2, 1, 2, 51),
                        errors: vec![]
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(2, 1, 2, 2),
                            errors: vec![]
                        },
                        name: "a".to_string()
                    },
                    init: Expression::PipeExpr(Box::new(PipeExpr {
                        base: BaseNode {
                            location: loc.get(2, 5, 2, 51),
                            errors: vec![]
                        },
                        argument: Expression::Call(Box::new(CallExpr {
                            base: BaseNode {
                                location: loc.get(2, 5, 2, 31),
                                errors: vec![]
                            },
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(2, 5, 2, 9),
                                    errors: vec![]
                                },
                                name: "from".to_string()
                            }),
                            arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                base: BaseNode {
                                    location: loc.get(2, 10, 2, 30),
                                    errors: vec![]
                                },
                                with: None,
                                properties: vec![Property {
                                    base: BaseNode {
                                        location: loc.get(2, 10, 2, 30),
                                        errors: vec![]
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(2, 10, 2, 16),
                                            errors: vec![]
                                        },
                                        name: "bucket".to_string()
                                    }),
                                    value: Some(Expression::StringLit(StringLit {
                                        base: BaseNode {
                                            location: loc.get(2, 17, 2, 30),
                                            errors: vec![]
                                        },
                                        value: "dbA/autogen".to_string()
                                    }))
                                }]
                            }))]
                        })),
                        call: CallExpr {
                            base: BaseNode {
                                location: loc.get(2, 35, 2, 51),
                                errors: vec![]
                            },
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(2, 35, 2, 40),
                                    errors: vec![]
                                },
                                name: "range".to_string()
                            }),
                            arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                base: BaseNode {
                                    location: loc.get(2, 41, 2, 50),
                                    errors: vec![]
                                },
                                with: None,
                                properties: vec![Property {
                                    base: BaseNode {
                                        location: loc.get(2, 41, 2, 50),
                                        errors: vec![]
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(2, 41, 2, 46),
                                            errors: vec![]
                                        },
                                        name: "start".to_string()
                                    }),
                                    value: Some(Expression::Unary(Box::new(UnaryExpr {
                                        base: BaseNode {
                                            location: loc.get(2, 47, 2, 50),
                                            errors: vec![]
                                        },
                                        operator: Operator::SubtractionOperator,
                                        argument: Expression::Duration(DurationLit {
                                            base: BaseNode {
                                                location: loc.get(2, 48, 2, 50),
                                                errors: vec![]
                                            },
                                            values: vec![Duration {
                                                magnitude: 1,
                                                unit: "h".to_string()
                                            }]
                                        })
                                    })))
                                }]
                            }))]
                        }
                    }))
                })),
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(3, 1, 3, 51),
                        errors: vec![]
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(3, 1, 3, 2),
                            errors: vec![]
                        },
                        name: "b".to_string()
                    },
                    init: Expression::PipeExpr(Box::new(PipeExpr {
                        base: BaseNode {
                            location: loc.get(3, 5, 3, 51),
                            errors: vec![]
                        },
                        argument: Expression::Call(Box::new(CallExpr {
                            base: BaseNode {
                                location: loc.get(3, 5, 3, 31),
                                errors: vec![]
                            },
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(3, 5, 3, 9),
                                    errors: vec![]
                                },
                                name: "from".to_string()
                            }),
                            arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                base: BaseNode {
                                    location: loc.get(3, 10, 3, 30),
                                    errors: vec![]
                                },
                                with: None,
                                properties: vec![Property {
                                    base: BaseNode {
                                        location: loc.get(3, 10, 3, 30),
                                        errors: vec![]
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(3, 10, 3, 16),
                                            errors: vec![]
                                        },
                                        name: "bucket".to_string()
                                    }),
                                    value: Some(Expression::StringLit(StringLit {
                                        base: BaseNode {
                                            location: loc.get(3, 17, 3, 30),
                                            errors: vec![]
                                        },
                                        value: "dbB/autogen".to_string()
                                    }))
                                }]
                            }))]
                        })),
                        call: CallExpr {
                            base: BaseNode {
                                location: loc.get(3, 35, 3, 51),
                                errors: vec![]
                            },
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(3, 35, 3, 40),
                                    errors: vec![]
                                },
                                name: "range".to_string()
                            }),
                            arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                base: BaseNode {
                                    location: loc.get(3, 41, 3, 50),
                                    errors: vec![]
                                },
                                with: None,
                                properties: vec![Property {
                                    base: BaseNode {
                                        location: loc.get(3, 41, 3, 50),
                                        errors: vec![]
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(3, 41, 3, 46),
                                            errors: vec![]
                                        },
                                        name: "start".to_string()
                                    }),
                                    value: Some(Expression::Unary(Box::new(UnaryExpr {
                                        base: BaseNode {
                                            location: loc.get(3, 47, 3, 50),
                                            errors: vec![]
                                        },
                                        operator: Operator::SubtractionOperator,
                                        argument: Expression::Duration(DurationLit {
                                            base: BaseNode {
                                                location: loc.get(3, 48, 3, 50),
                                                errors: vec![]
                                            },
                                            values: vec![Duration {
                                                magnitude: 1,
                                                unit: "h".to_string()
                                            }]
                                        })
                                    })))
                                }]
                            }))]
                        }
                    }))
                })),
                Statement::Expr(ExprStmt {
                    base: BaseNode {
                        location: loc.get(4, 1, 4, 72),
                        errors: vec![]
                    },
                    expression: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(4, 1, 4, 72),
                            errors: vec![]
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(4, 1, 4, 5),
                                errors: vec![]
                            },
                            name: "join".to_string()
                        }),
                        arguments: vec![Expression::Object(Box::new(ObjectExpr {
                            base: BaseNode {
                                location: loc.get(4, 6, 4, 71),
                                errors: vec![]
                            },
                            with: None,
                            properties: vec![
                                Property {
                                    base: BaseNode {
                                        location: loc.get(4, 6, 4, 18),
                                        errors: vec![]
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(4, 6, 4, 12),
                                            errors: vec![]
                                        },
                                        name: "tables".to_string()
                                    }),
                                    value: Some(Expression::Array(Box::new(ArrayExpr {
                                        base: BaseNode {
                                            location: loc.get(4, 13, 4, 18),
                                            errors: vec![]
                                        },
                                        elements: vec![
                                            Expression::Identifier(Identifier {
                                                base: BaseNode {
                                                    location: loc.get(4, 14, 4, 15),
                                                    errors: vec![]
                                                },
                                                name: "a".to_string()
                                            }),
                                            Expression::Identifier(Identifier {
                                                base: BaseNode {
                                                    location: loc.get(4, 16, 4, 17),
                                                    errors: vec![]
                                                },
                                                name: "b".to_string()
                                            })
                                        ]
                                    })))
                                },
                                Property {
                                    base: BaseNode {
                                        location: loc.get(4, 20, 4, 31),
                                        errors: vec![]
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(4, 20, 4, 22),
                                            errors: vec![]
                                        },
                                        name: "on".to_string()
                                    }),
                                    value: Some(Expression::Array(Box::new(ArrayExpr {
                                        base: BaseNode {
                                            location: loc.get(4, 23, 4, 31),
                                            errors: vec![]
                                        },
                                        elements: vec![Expression::StringLit(StringLit {
                                            base: BaseNode {
                                                location: loc.get(4, 24, 4, 30),
                                                errors: vec![]
                                            },
                                            value: "host".to_string()
                                        })]
                                    })))
                                },
                                Property {
                                    base: BaseNode {
                                        location: loc.get(4, 33, 4, 71),
                                        errors: vec![]
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(4, 33, 4, 35),
                                            errors: vec![]
                                        },
                                        name: "fn".to_string()
                                    }),
                                    value: Some(Expression::Function(Box::new(FunctionExpr {
                                        base: BaseNode {
                                            location: loc.get(4, 37, 4, 71),
                                            errors: vec![]
                                        },
                                        params: vec![
                                            Property {
                                                base: BaseNode {
                                                    location: loc.get(4, 38, 4, 39),
                                                    errors: vec![]
                                                },
                                                key: PropertyKey::Identifier(Identifier {
                                                    base: BaseNode {
                                                        location: loc.get(4, 38, 4, 39),
                                                        errors: vec![]
                                                    },
                                                    name: "a".to_string()
                                                }),
                                                value: None
                                            },
                                            Property {
                                                base: BaseNode {
                                                    location: loc.get(4, 40, 4, 41),
                                                    errors: vec![]
                                                },
                                                key: PropertyKey::Identifier(Identifier {
                                                    base: BaseNode {
                                                        location: loc.get(4, 40, 4, 41),
                                                        errors: vec![]
                                                    },
                                                    name: "b".to_string()
                                                }),
                                                value: None
                                            }
                                        ],
                                        body: FunctionBody::Expr(Expression::Binary(Box::new(
                                            BinaryExpr {
                                                base: BaseNode {
                                                    location: loc.get(4, 46, 4, 71),
                                                    errors: vec![]
                                                },
                                                operator: Operator::AdditionOperator,
                                                left: Expression::Member(Box::new(MemberExpr {
                                                    base: BaseNode {
                                                        location: loc.get(4, 46, 4, 57),
                                                        errors: vec![]
                                                    },
                                                    object: Expression::Identifier(Identifier {
                                                        base: BaseNode {
                                                            location: loc.get(4, 46, 4, 47),
                                                            errors: vec![]
                                                        },
                                                        name: "a".to_string()
                                                    }),
                                                    property: PropertyKey::StringLit(StringLit {
                                                        base: BaseNode {
                                                            location: loc.get(4, 48, 4, 56),
                                                            errors: vec![]
                                                        },
                                                        value: "_field".to_string()
                                                    })
                                                })),
                                                right: Expression::Member(Box::new(MemberExpr {
                                                    base: BaseNode {
                                                        location: loc.get(4, 60, 4, 71),
                                                        errors: vec![]
                                                    },
                                                    object: Expression::Identifier(Identifier {
                                                        base: BaseNode {
                                                            location: loc.get(4, 60, 4, 61),
                                                            errors: vec![]
                                                        },
                                                        name: "b".to_string()
                                                    }),
                                                    property: PropertyKey::StringLit(StringLit {
                                                        base: BaseNode {
                                                            location: loc.get(4, 62, 4, 70),
                                                            errors: vec![]
                                                        },
                                                        value: "_field".to_string()
                                                    })
                                                }))
                                            }
                                        )))
                                    })))
                                }
                            ]
                        }))]
                    }))
                })
            ]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(2, 1, 4, 21),
                        errors: vec![]
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(2, 1, 2, 2),
                            errors: vec![]
                        },
                        name: "a".to_string()
                    },
                    init: Expression::PipeExpr(Box::new(PipeExpr {
                        base: BaseNode {
                            location: loc.get(2, 5, 4, 21),
                            errors: vec![]
                        },
                        argument: Expression::PipeExpr(Box::new(PipeExpr {
                            base: BaseNode {
                                location: loc.get(2, 5, 3, 48),
                                errors: vec![]
                            },
                            argument: Expression::Call(Box::new(CallExpr {
                                base: BaseNode {
                                    location: loc.get(2, 5, 2, 32),
                                    errors: vec![]
                                },
                                callee: Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(2, 5, 2, 9),
                                        errors: vec![]
                                    },
                                    name: "from".to_string()
                                }),
                                arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                    base: BaseNode {
                                        location: loc.get(2, 10, 2, 31),
                                        errors: vec![]
                                    },
                                    with: None,
                                    properties: vec![Property {
                                        base: BaseNode {
                                            location: loc.get(2, 10, 2, 31),
                                            errors: vec![]
                                        },
                                        key: PropertyKey::Identifier(Identifier {
                                            base: BaseNode {
                                                location: loc.get(2, 10, 2, 16),
                                                errors: vec![]
                                            },
                                            name: "bucket".to_string()
                                        }),
                                        value: Some(Expression::StringLit(StringLit {
                                            base: BaseNode {
                                                location: loc.get(2, 17, 2, 31),
                                                errors: vec![]
                                            },
                                            value: "Flux/autogen".to_string()
                                        }))
                                    }]
                                }))]
                            })),
                            call: CallExpr {
                                base: BaseNode {
                                    location: loc.get(3, 5, 3, 48),
                                    errors: vec![]
                                },
                                callee: Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(3, 5, 3, 11),
                                        errors: vec![]
                                    },
                                    name: "filter".to_string()
                                }),
                                arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                    base: BaseNode {
                                        location: loc.get(3, 12, 3, 47),
                                        errors: vec![]
                                    },
                                    with: None,
                                    properties: vec![Property {
                                        base: BaseNode {
                                            location: loc.get(3, 12, 3, 47),
                                            errors: vec![]
                                        },
                                        key: PropertyKey::Identifier(Identifier {
                                            base: BaseNode {
                                                location: loc.get(3, 12, 3, 14),
                                                errors: vec![]
                                            },
                                            name: "fn".to_string()
                                        }),
                                        value: Some(Expression::Function(Box::new(FunctionExpr {
                                            base: BaseNode {
                                                location: loc.get(3, 16, 3, 47),
                                                errors: vec![]
                                            },
                                            params: vec![Property {
                                                base: BaseNode {
                                                    location: loc.get(3, 17, 3, 18),
                                                    errors: vec![]
                                                },
                                                key: PropertyKey::Identifier(Identifier {
                                                    base: BaseNode {
                                                        location: loc.get(3, 17, 3, 18),
                                                        errors: vec![]
                                                    },
                                                    name: "r".to_string()
                                                }),
                                                value: None
                                            }],
                                            body: FunctionBody::Expr(Expression::Binary(Box::new(
                                                BinaryExpr {
                                                    base: BaseNode {
                                                        location: loc.get(3, 23, 3, 47),
                                                        errors: vec![]
                                                    },
                                                    operator: Operator::EqualOperator,
                                                    left: Expression::Member(Box::new(
                                                        MemberExpr {
                                                            base: BaseNode {
                                                                location: loc.get(3, 23, 3, 40),
                                                                errors: vec![]
                                                            },
                                                            object: Expression::Identifier(
                                                                Identifier {
                                                                    base: BaseNode {
                                                                        location: loc
                                                                            .get(3, 23, 3, 24),
                                                                        errors: vec![]
                                                                    },
                                                                    name: "r".to_string()
                                                                }
                                                            ),
                                                            property: PropertyKey::StringLit(
                                                                StringLit {
                                                                    base: BaseNode {
                                                                        location: loc
                                                                            .get(3, 25, 3, 39),
                                                                        errors: vec![]
                                                                    },
                                                                    value: "_measurement"
                                                                        .to_string()
                                                                }
                                                            )
                                                        }
                                                    )),
                                                    right: Expression::StringLit(StringLit {
                                                        base: BaseNode {
                                                            location: loc.get(3, 44, 3, 47),
                                                            errors: vec![]
                                                        },
                                                        value: "a".to_string()
                                                    })
                                                }
                                            )))
                                        })))
                                    }]
                                }))]
                            }
                        })),
                        call: CallExpr {
                            base: BaseNode {
                                location: loc.get(4, 5, 4, 21),
                                errors: vec![]
                            },
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(4, 5, 4, 10),
                                    errors: vec![]
                                },
                                name: "range".to_string()
                            }),
                            arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                base: BaseNode {
                                    location: loc.get(4, 11, 4, 20),
                                    errors: vec![]
                                },
                                with: None,
                                properties: vec![Property {
                                    base: BaseNode {
                                        location: loc.get(4, 11, 4, 20),
                                        errors: vec![]
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(4, 11, 4, 16),
                                            errors: vec![]
                                        },
                                        name: "start".to_string()
                                    }),
                                    value: Some(Expression::Unary(Box::new(UnaryExpr {
                                        base: BaseNode {
                                            location: loc.get(4, 17, 4, 20),
                                            errors: vec![]
                                        },
                                        operator: Operator::SubtractionOperator,
                                        argument: Expression::Duration(DurationLit {
                                            base: BaseNode {
                                                location: loc.get(4, 18, 4, 20),
                                                errors: vec![]
                                            },
                                            values: vec![Duration {
                                                magnitude: 1,
                                                unit: "h".to_string()
                                            }]
                                        })
                                    })))
                                }]
                            }))]
                        }
                    }))
                })),
                Statement::Variable(Box::new(VariableAssgn {
                    base: BaseNode {
                        location: loc.get(6, 1, 8, 21),
                        errors: vec![]
                    },
                    id: Identifier {
                        base: BaseNode {
                            location: loc.get(6, 1, 6, 2),
                            errors: vec![]
                        },
                        name: "b".to_string()
                    },
                    init: Expression::PipeExpr(Box::new(PipeExpr {
                        base: BaseNode {
                            location: loc.get(6, 5, 8, 21),
                            errors: vec![]
                        },
                        argument: Expression::PipeExpr(Box::new(PipeExpr {
                            base: BaseNode {
                                location: loc.get(6, 5, 7, 48),
                                errors: vec![]
                            },
                            argument: Expression::Call(Box::new(CallExpr {
                                base: BaseNode {
                                    location: loc.get(6, 5, 6, 32),
                                    errors: vec![]
                                },
                                callee: Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(6, 5, 6, 9),
                                        errors: vec![]
                                    },
                                    name: "from".to_string()
                                }),
                                arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                    base: BaseNode {
                                        location: loc.get(6, 10, 6, 31),
                                        errors: vec![]
                                    },
                                    with: None,
                                    properties: vec![Property {
                                        base: BaseNode {
                                            location: loc.get(6, 10, 6, 31),
                                            errors: vec![]
                                        },
                                        key: PropertyKey::Identifier(Identifier {
                                            base: BaseNode {
                                                location: loc.get(6, 10, 6, 16),
                                                errors: vec![]
                                            },
                                            name: "bucket".to_string()
                                        }),
                                        value: Some(Expression::StringLit(StringLit {
                                            base: BaseNode {
                                                location: loc.get(6, 17, 6, 31),
                                                errors: vec![]
                                            },
                                            value: "Flux/autogen".to_string()
                                        }))
                                    }]
                                }))]
                            })),
                            call: CallExpr {
                                base: BaseNode {
                                    location: loc.get(7, 5, 7, 48),
                                    errors: vec![]
                                },
                                callee: Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(7, 5, 7, 11),
                                        errors: vec![]
                                    },
                                    name: "filter".to_string()
                                }),
                                arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                    base: BaseNode {
                                        location: loc.get(7, 12, 7, 47),
                                        errors: vec![]
                                    },
                                    with: None,
                                    properties: vec![Property {
                                        base: BaseNode {
                                            location: loc.get(7, 12, 7, 47),
                                            errors: vec![]
                                        },
                                        key: PropertyKey::Identifier(Identifier {
                                            base: BaseNode {
                                                location: loc.get(7, 12, 7, 14),
                                                errors: vec![]
                                            },
                                            name: "fn".to_string()
                                        }),
                                        value: Some(Expression::Function(Box::new(FunctionExpr {
                                            base: BaseNode {
                                                location: loc.get(7, 16, 7, 47),
                                                errors: vec![]
                                            },
                                            params: vec![Property {
                                                base: BaseNode {
                                                    location: loc.get(7, 17, 7, 18),
                                                    errors: vec![]
                                                },
                                                key: PropertyKey::Identifier(Identifier {
                                                    base: BaseNode {
                                                        location: loc.get(7, 17, 7, 18),
                                                        errors: vec![]
                                                    },
                                                    name: "r".to_string()
                                                }),
                                                value: None
                                            }],
                                            body: FunctionBody::Expr(Expression::Binary(Box::new(
                                                BinaryExpr {
                                                    base: BaseNode {
                                                        location: loc.get(7, 23, 7, 47),
                                                        errors: vec![]
                                                    },
                                                    operator: Operator::EqualOperator,
                                                    left: Expression::Member(Box::new(
                                                        MemberExpr {
                                                            base: BaseNode {
                                                                location: loc.get(7, 23, 7, 40),
                                                                errors: vec![]
                                                            },
                                                            object: Expression::Identifier(
                                                                Identifier {
                                                                    base: BaseNode {
                                                                        location: loc
                                                                            .get(7, 23, 7, 24),
                                                                        errors: vec![]
                                                                    },
                                                                    name: "r".to_string()
                                                                }
                                                            ),
                                                            property: PropertyKey::StringLit(
                                                                StringLit {
                                                                    base: BaseNode {
                                                                        location: loc
                                                                            .get(7, 25, 7, 39),
                                                                        errors: vec![]
                                                                    },
                                                                    value: "_measurement"
                                                                        .to_string()
                                                                }
                                                            )
                                                        }
                                                    )),
                                                    right: Expression::StringLit(StringLit {
                                                        base: BaseNode {
                                                            location: loc.get(7, 44, 7, 47),
                                                            errors: vec![]
                                                        },
                                                        value: "b".to_string()
                                                    })
                                                }
                                            )))
                                        })))
                                    }]
                                }))]
                            }
                        })),
                        call: CallExpr {
                            base: BaseNode {
                                location: loc.get(8, 5, 8, 21),
                                errors: vec![]
                            },
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(8, 5, 8, 10),
                                    errors: vec![]
                                },
                                name: "range".to_string()
                            }),
                            arguments: vec![Expression::Object(Box::new(ObjectExpr {
                                base: BaseNode {
                                    location: loc.get(8, 11, 8, 20),
                                    errors: vec![]
                                },
                                with: None,
                                properties: vec![Property {
                                    base: BaseNode {
                                        location: loc.get(8, 11, 8, 20),
                                        errors: vec![]
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(8, 11, 8, 16),
                                            errors: vec![]
                                        },
                                        name: "start".to_string()
                                    }),
                                    value: Some(Expression::Unary(Box::new(UnaryExpr {
                                        base: BaseNode {
                                            location: loc.get(8, 17, 8, 20),
                                            errors: vec![]
                                        },
                                        operator: Operator::SubtractionOperator,
                                        argument: Expression::Duration(DurationLit {
                                            base: BaseNode {
                                                location: loc.get(8, 18, 8, 20),
                                                errors: vec![]
                                            },
                                            values: vec![Duration {
                                                magnitude: 1,
                                                unit: "h".to_string()
                                            }]
                                        })
                                    })))
                                }]
                            }))]
                        }
                    }))
                })),
                Statement::Expr(ExprStmt {
                    base: BaseNode {
                        location: loc.get(10, 1, 10, 86),
                        errors: vec![]
                    },
                    expression: Expression::Call(Box::new(CallExpr {
                        base: BaseNode {
                            location: loc.get(10, 1, 10, 86),
                            errors: vec![]
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(10, 1, 10, 5),
                                errors: vec![]
                            },
                            name: "join".to_string()
                        }),
                        arguments: vec![Expression::Object(Box::new(ObjectExpr {
                            base: BaseNode {
                                location: loc.get(10, 6, 10, 85),
                                errors: vec![]
                            },
                            with: None,
                            properties: vec![
                                Property {
                                    base: BaseNode {
                                        location: loc.get(10, 6, 10, 18),
                                        errors: vec![]
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(10, 6, 10, 12),
                                            errors: vec![]
                                        },
                                        name: "tables".to_string()
                                    }),
                                    value: Some(Expression::Array(Box::new(ArrayExpr {
                                        base: BaseNode {
                                            location: loc.get(10, 13, 10, 18),
                                            errors: vec![]
                                        },
                                        elements: vec![
                                            Expression::Identifier(Identifier {
                                                base: BaseNode {
                                                    location: loc.get(10, 14, 10, 15),
                                                    errors: vec![]
                                                },
                                                name: "a".to_string()
                                            }),
                                            Expression::Identifier(Identifier {
                                                base: BaseNode {
                                                    location: loc.get(10, 16, 10, 17),
                                                    errors: vec![]
                                                },
                                                name: "b".to_string()
                                            })
                                        ]
                                    })))
                                },
                                Property {
                                    base: BaseNode {
                                        location: loc.get(10, 20, 10, 29),
                                        errors: vec![]
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(10, 20, 10, 22),
                                            errors: vec![]
                                        },
                                        name: "on".to_string()
                                    }),
                                    value: Some(Expression::Array(Box::new(ArrayExpr {
                                        base: BaseNode {
                                            location: loc.get(10, 23, 10, 29),
                                            errors: vec![]
                                        },
                                        elements: vec![Expression::StringLit(StringLit {
                                            base: BaseNode {
                                                location: loc.get(10, 24, 10, 28),
                                                errors: vec![]
                                            },
                                            value: "t1".to_string()
                                        })]
                                    })))
                                },
                                Property {
                                    base: BaseNode {
                                        location: loc.get(10, 31, 10, 85),
                                        errors: vec![]
                                    },
                                    key: PropertyKey::Identifier(Identifier {
                                        base: BaseNode {
                                            location: loc.get(10, 31, 10, 33),
                                            errors: vec![]
                                        },
                                        name: "fn".to_string()
                                    }),
                                    value: Some(Expression::Function(Box::new(FunctionExpr {
                                        base: BaseNode {
                                            location: loc.get(10, 35, 10, 85),
                                            errors: vec![]
                                        },
                                        params: vec![
                                            Property {
                                                base: BaseNode {
                                                    location: loc.get(10, 36, 10, 37),
                                                    errors: vec![]
                                                },
                                                key: PropertyKey::Identifier(Identifier {
                                                    base: BaseNode {
                                                        location: loc.get(10, 36, 10, 37),
                                                        errors: vec![]
                                                    },
                                                    name: "a".to_string()
                                                }),
                                                value: None
                                            },
                                            Property {
                                                base: BaseNode {
                                                    location: loc.get(10, 38, 10, 39),
                                                    errors: vec![]
                                                },
                                                key: PropertyKey::Identifier(Identifier {
                                                    base: BaseNode {
                                                        location: loc.get(10, 38, 10, 39),
                                                        errors: vec![]
                                                    },
                                                    name: "b".to_string()
                                                }),
                                                value: None
                                            }
                                        ],
                                        body: FunctionBody::Expr(Expression::Binary(Box::new(
                                            BinaryExpr {
                                                base: BaseNode {
                                                    location: loc.get(10, 44, 10, 85),
                                                    errors: vec![]
                                                },
                                                operator: Operator::DivisionOperator,
                                                left: Expression::Paren(Box::new(ParenExpr {
                                                    base: BaseNode {
                                                        location: loc.get(10, 44, 10, 71),
                                                        errors: vec![]
                                                    },
                                                    expression: Expression::Binary(Box::new(
                                                        BinaryExpr {
                                                            base: BaseNode {
                                                                location: loc.get(10, 45, 10, 70),
                                                                errors: vec![]
                                                            },
                                                            operator: Operator::SubtractionOperator,
                                                            left: Expression::Member(Box::new(
                                                                MemberExpr {
                                                                    base: BaseNode {
                                                                        location: loc
                                                                            .get(10, 45, 10, 56),
                                                                        errors: vec![]
                                                                    },
                                                                    object: Expression::Identifier(
                                                                        Identifier {
                                                                            base: BaseNode {
                                                                                location: loc.get(
                                                                                    10, 45, 10, 46
                                                                                ),
                                                                                errors: vec![]
                                                                            },
                                                                            name: "a".to_string()
                                                                        }
                                                                    ),
                                                                    property:
                                                                        PropertyKey::StringLit(
                                                                            StringLit {
                                                                                base: BaseNode {
                                                                                    location: loc
                                                                                        .get(
                                                                                            10, 47,
                                                                                            10, 55
                                                                                        ),
                                                                                    errors: vec![]
                                                                                },
                                                                                value: "_field"
                                                                                    .to_string()
                                                                            }
                                                                        )
                                                                }
                                                            )),
                                                            right: Expression::Member(Box::new(
                                                                MemberExpr {
                                                                    base: BaseNode {
                                                                        location: loc
                                                                            .get(10, 59, 10, 70),
                                                                        errors: vec![]
                                                                    },
                                                                    object: Expression::Identifier(
                                                                        Identifier {
                                                                            base: BaseNode {
                                                                                location: loc.get(
                                                                                    10, 59, 10, 60
                                                                                ),
                                                                                errors: vec![]
                                                                            },
                                                                            name: "b".to_string()
                                                                        }
                                                                    ),
                                                                    property:
                                                                        PropertyKey::StringLit(
                                                                            StringLit {
                                                                                base: BaseNode {
                                                                                    location: loc
                                                                                        .get(
                                                                                            10, 61,
                                                                                            10, 69
                                                                                        ),
                                                                                    errors: vec![]
                                                                                },
                                                                                value: "_field"
                                                                                    .to_string()
                                                                            }
                                                                        )
                                                                }
                                                            ))
                                                        }
                                                    ))
                                                })),
                                                right: Expression::Member(Box::new(MemberExpr {
                                                    base: BaseNode {
                                                        location: loc.get(10, 74, 10, 85),
                                                        errors: vec![]
                                                    },
                                                    object: Expression::Identifier(Identifier {
                                                        base: BaseNode {
                                                            location: loc.get(10, 74, 10, 75),
                                                            errors: vec![]
                                                        },
                                                        name: "b".to_string()
                                                    }),
                                                    property: PropertyKey::StringLit(StringLit {
                                                        base: BaseNode {
                                                            location: loc.get(10, 76, 10, 84),
                                                            errors: vec![]
                                                        },
                                                        value: "_field".to_string()
                                                    })
                                                }))
                                            }
                                        )))
                                    })))
                                }
                            ]
                        }))]
                    }))
                })
            ]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 34),
                    errors: vec![]
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 4),
                        errors: vec![]
                    },
                    name: "dur".to_string()
                },
                init: Expression::Duration(DurationLit {
                    base: BaseNode {
                        location: loc.get(1, 7, 1, 34),
                        errors: vec![]
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
            }))]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 10),
                    errors: vec![]
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 4),
                        errors: vec![]
                    },
                    name: "dur".to_string()
                },
                init: Expression::Duration(DurationLit {
                    base: BaseNode {
                        location: loc.get(1, 7, 1, 10),
                        errors: vec![]
                    },
                    values: vec![Duration {
                        magnitude: 6,
                        unit: "mo".to_string()
                    }]
                })
            }))]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 12),
                    errors: vec![]
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 4),
                        errors: vec![]
                    },
                    name: "dur".to_string()
                },
                init: Expression::Duration(DurationLit {
                    base: BaseNode {
                        location: loc.get(1, 7, 1, 12),
                        errors: vec![]
                    },
                    values: vec![Duration {
                        magnitude: 500,
                        unit: "ms".to_string()
                    }]
                })
            }))]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 18),
                    errors: vec![]
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 4),
                        errors: vec![]
                    },
                    name: "dur".to_string()
                },
                init: Expression::Duration(DurationLit {
                    base: BaseNode {
                        location: loc.get(1, 7, 1, 18),
                        errors: vec![]
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
            }))]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 17),
                    errors: vec![]
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 4),
                        errors: vec![]
                    },
                    name: "now".to_string()
                },
                init: Expression::DateTime(DateTimeLit {
                    base: BaseNode {
                        location: loc.get(1, 7, 1, 17),
                        errors: vec![]
                    },
                    value: chrono::DateTime::parse_from_rfc3339("2018-11-29T00:00:00Z").unwrap()
                })
            }))]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 27),
                    errors: vec![]
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 4),
                        errors: vec![]
                    },
                    name: "now".to_string()
                },
                init: Expression::DateTime(DateTimeLit {
                    base: BaseNode {
                        location: loc.get(1, 7, 1, 27),
                        errors: vec![]
                    },
                    value: chrono::DateTime::parse_from_rfc3339("2018-11-29T09:00:00Z").unwrap()
                })
            }))]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 37),
                    errors: vec![]
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 4),
                        errors: vec![]
                    },
                    name: "now".to_string()
                },
                init: Expression::DateTime(DateTimeLit {
                    base: BaseNode {
                        location: loc.get(1, 7, 1, 37),
                        errors: vec![]
                    },
                    value: chrono::DateTime::parse_from_rfc3339("2018-11-29T09:00:00.100000000Z")
                        .unwrap()
                })
            }))]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 56),
                    errors: vec![]
                },
                expression: Expression::PipeExpr(Box::new(PipeExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 56),
                        errors: vec![]
                    },
                    argument: Expression::PipeExpr(Box::new(PipeExpr {
                        base: BaseNode {
                            location: loc.get(1, 1, 1, 18),
                            errors: vec![]
                        },
                        argument: Expression::Call(Box::new(CallExpr {
                            base: BaseNode {
                                location: loc.get(1, 1, 1, 7),
                                errors: vec![]
                            },
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 1, 1, 5),
                                    errors: vec![]
                                },
                                name: "from".to_string()
                            }),
                            arguments: vec![]
                        })),
                        call: CallExpr {
                            base: BaseNode {
                                location: loc.get(1, 11, 1, 18),
                                errors: vec![]
                            },
                            callee: Expression::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 11, 1, 16),
                                    errors: vec![]
                                },
                                name: "range".to_string()
                            }),
                            arguments: vec![]
                        }
                    })),
                    call: CallExpr {
                        base: BaseNode {
                            location: loc.get(1, 22, 1, 56),
                            errors: vec![]
                        },
                        callee: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 22, 1, 25),
                                errors: vec![]
                            },
                            name: "map".to_string()
                        }),
                        arguments: vec![Expression::Object(Box::new(ObjectExpr {
                            base: BaseNode {
                                location: loc.get(1, 26, 1, 56),
                                errors: vec![]
                            },
                            with: None,
                            properties: vec![Property {
                                base: BaseNode {
                                    location: loc.get(1, 26, 1, 56),
                                    errors: vec![]
                                },
                                key: PropertyKey::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 26, 1, 28),
                                        errors: vec![]
                                    },
                                    name: "fn".to_string()
                                }),
                                value: Some(Expression::Function(Box::new(FunctionExpr {
                                    base: BaseNode {
                                        location: loc.get(1, 30, 1, 56),
                                        errors: vec![]
                                    },
                                    params: vec![Property {
                                        base: BaseNode {
                                            location: loc.get(1, 31, 1, 32),
                                            errors: vec![]
                                        },
                                        key: PropertyKey::Identifier(Identifier {
                                            base: BaseNode {
                                                location: loc.get(1, 31, 1, 32),
                                                errors: vec![]
                                            },
                                            name: "r".to_string()
                                        }),
                                        value: None
                                    }],
                                    body: FunctionBody::Block(Block {
                                        base: BaseNode {
                                            location: loc.get(1, 37, 1, 56),
                                            errors: vec!["expected RBRACE, got RPAREN".to_string()]
                                        },
                                        body: vec![Statement::Return(ReturnStmt {
                                            base: BaseNode {
                                                location: loc.get(1, 39, 1, 54),
                                                errors: vec![]
                                            },
                                            argument: Expression::Member(Box::new(MemberExpr {
                                                base: BaseNode {
                                                    location: loc.get(1, 46, 1, 54),
                                                    errors: vec![]
                                                },
                                                object: Expression::Identifier(Identifier {
                                                    base: BaseNode {
                                                        location: loc.get(1, 46, 1, 47),
                                                        errors: vec![]
                                                    },
                                                    name: "r".to_string()
                                                }),
                                                property: PropertyKey::Identifier(Identifier {
                                                    base: BaseNode {
                                                        location: loc.get(1, 48, 1, 54),
                                                        errors: vec![]
                                                    },
                                                    name: "_value".to_string()
                                                })
                                            }))
                                        })]
                                    })
                                })))
                            }]
                        }))]
                    }
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 12),
                    errors: vec![]
                },
                expression: Expression::StringLit(StringLit {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 12),
                        errors: vec![]
                    },
                    value: "日本語".to_string()
                })
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 39),
                    errors: vec![]
                },
                expression: Expression::StringLit(StringLit {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 39),
                        errors: vec![]
                    },
                    value: "日本語".to_string()
                })
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 63),
                    errors: vec![]
                },
                expression: Expression::StringLit(StringLit {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 63),
                        errors: vec![]
                    },
                    value: "hello 日x本 日本語 µs".to_string()
                })
            })]
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
            base: BaseNode {location: loc.get(1, 1, 6, 2), errors: vec![] },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {location: loc.get(1, 1, 6, 2), errors: vec![] },
                expression: Expression::StringLit(StringLit {
                    base: BaseNode {location: loc.get(1, 1, 6, 2), errors: vec![] },
                    value: "newline \n\ncarriage return \r\nhorizontal tab \t\ndouble quote \"\nbackslash \\\n".to_string()
                })
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 4, 8),
                    errors: vec![]
                },
                expression: Expression::StringLit(StringLit {
                    base: BaseNode {
                        location: loc.get(1, 1, 4, 8),
                        errors: vec![]
                    },
                    value: "\n this is a\nmultiline\nstring".to_string()
                })
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![
                Statement::Bad(BadStmt {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 2),
                        // errors: vec!["invalid statement @1:1-1:2: @".to_string()]
                        errors: vec![],
                    },
                    text: "@".to_string()
                }),
                Statement::Expr(ExprStmt {
                    base: BaseNode {
                        location: loc.get(1, 3, 1, 8),
                        errors: vec![]
                    },
                    expression: Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 3, 1, 8),
                            errors: vec![]
                        },
                        name: "ident".to_string()
                    })
                })
            ]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 6),
                    errors: vec![]
                },
                expression: Expression::Paren(Box::new(ParenExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 6),
                        errors: vec![]
                    },
                    expression: Expression::Binary(Box::new(BinaryExpr {
                        // TODO(affo): ast.Check would add the error "expected an operator between two expressions".
                        base: BaseNode {
                            location: loc.get(1, 2, 1, 5),
                            errors: vec![]
                        },
                        operator: Operator::InvalidOperator,
                        left: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 2, 1, 3),
                                errors: vec![]
                            },
                            name: "a".to_string()
                        }),
                        right: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 4, 1, 5),
                                errors: vec![]
                            },
                            name: "b".to_string()
                        })
                    }))
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 5),
                    errors: vec![]
                },
                expression: Expression::Paren(Box::new(ParenExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 5),
                        errors: vec![]
                    },
                    expression: Expression::Binary(Box::new(BinaryExpr {
                        // TODO(affo): this should be like this:
                        // base: BaseNode {location: ..., errors: vec!["missing left hand side of expression".to_string()] },
                        base: BaseNode {
                            location: loc.get(1, 2, 1, 4),
                            errors: vec![]
                        },
                        operator: Operator::MultiplicationOperator,
                        left: Expression::Bad(Box::new(BadExpr {
                            base: BaseNode {
                                location: loc.get(1, 2, 1, 3),
                                errors: vec![]
                            },
                            text: "invalid token for primary expression: MUL".to_string(),
                            expression: None
                        })),
                        right: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 3, 1, 4),
                                errors: vec![]
                            },
                            name: "b".to_string()
                        })
                    }))
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 5),
                    errors: vec![]
                },
                expression: Expression::Paren(Box::new(ParenExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 5),
                        errors: vec![]
                    },
                    expression: Expression::Binary(Box::new(BinaryExpr {
                        // TODO(affo): this should be like this:
                        // base: BaseNode {location: ..., errors: vec!["missing right hand side of expression".to_string()] },
                        base: BaseNode {
                            location: loc.get(1, 2, 1, 5),
                            errors: vec![]
                        },
                        operator: Operator::MultiplicationOperator,
                        left: Expression::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 2, 1, 3),
                                errors: vec![]
                            },
                            name: "a".to_string()
                        }),
                        right: Expression::Bad(Box::new(BadExpr {
                            base: BaseNode {
                                location: loc.get(1, 4, 1, 5),
                                errors: vec![]
                            },
                            text: "invalid token for primary expression: RPAREN".to_string(),
                            expression: None
                        })),
                    }))
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 4),
                    errors: vec![]
                },
                expression: Expression::Paren(Box::new(ParenExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 4),
                        errors: vec!["invalid expression @1:2-1:3: @".to_string()]
                    },
                    expression: Expression::Bad(Box::new(BadExpr {
                        base: BaseNode {
                            location: loc.get(1, 2, 1, 3),
                            errors: vec![]
                        },
                        text: "@".to_string(),
                        expression: None
                    }))
                }))
            }),]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 13),
                    errors: vec![]
                },
                expression: Expression::Function(Box::new(FunctionExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 13),
                        errors: vec![
                            "expected ARROW, got IDENT (a) at 1:8".to_string(),
                            "expected ARROW, got ADD (+) at 1:10".to_string(),
                            "expected ARROW, got IDENT (b) at 1:12".to_string(),
                            "expected ARROW, got EOF".to_string()
                        ]
                    },
                    params: vec![
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 2, 1, 3),
                                errors: vec![]
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 2, 1, 3),
                                    errors: vec![]
                                },
                                name: "a".to_string()
                            }),
                            value: None
                        },
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 5, 1, 6),
                                errors: vec![]
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 5, 1, 6),
                                    errors: vec![]
                                },
                                name: "b".to_string()
                            }),
                            value: None
                        }
                    ],
                    body: FunctionBody::Expr(Expression::Bad(Box::new(BadExpr {
                        base: BaseNode {
                            location: loc.get(1, 13, 1, 13),
                            errors: vec![]
                        },
                        text: "invalid token for primary expression: EOF".to_string(),
                        expression: None
                    })))
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 20),
                    errors: vec![]
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 2),
                        errors: vec![]
                    },
                    name: "o".to_string()
                },
                init: Expression::Object(Box::new(ObjectExpr {
                    base: BaseNode {
                        location: loc.get(1, 5, 1, 20),
                        errors: vec![]
                    },
                    with: None,
                    properties: vec![
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 12),
                                errors: vec![]
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 7),
                                    errors: vec![]
                                },
                                name: "a".to_string()
                            }),
                            value: Some(Expression::StringLit(StringLit {
                                base: BaseNode {
                                    location: loc.get(1, 9, 1, 12),
                                    errors: vec![]
                                },
                                value: "a".to_string()
                            }))
                        },
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 13, 1, 13),
                                errors: vec!["missing property in property list".to_string()]
                            },
                            key: PropertyKey::StringLit(StringLit {
                                base: BaseNode {
                                    location: loc.get(1, 13, 1, 13),
                                    errors: vec![]
                                },
                                value: "<invalid>".to_string()
                            }),
                            value: None
                        },
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 15, 1, 19),
                                errors: vec![]
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 15, 1, 16),
                                    errors: vec![]
                                },
                                name: "b".to_string()
                            }),
                            value: Some(Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 18, 1, 19),
                                    errors: vec![]
                                },
                                value: 7
                            }))
                        }
                    ]
                }))
            }))]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 12),
                    errors: vec![]
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 2),
                        errors: vec![]
                    },
                    name: "o".to_string()
                },
                init: Expression::Object(Box::new(ObjectExpr {
                    base: BaseNode {
                        location: loc.get(1, 5, 1, 12),
                        errors: vec![]
                    },
                    with: None,
                    properties: vec![Property {
                        base: BaseNode {
                            location: loc.get(1, 6, 1, 11),
                            errors: vec!["missing property key".to_string()]
                        },
                        key: PropertyKey::StringLit(StringLit {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 6),
                                errors: vec![]
                            },
                            value: "<invalid>".to_string()
                        }),
                        value: Some(Expression::StringLit(StringLit {
                            base: BaseNode {
                                location: loc.get(1, 8, 1, 11),
                                errors: vec![]
                            },
                            value: "a".to_string()
                        }))
                    }]
                }))
            }))]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 9),
                    errors: vec![]
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 2),
                        errors: vec![]
                    },
                    name: "o".to_string()
                },
                init: Expression::Object(Box::new(ObjectExpr {
                    base: BaseNode {
                        location: loc.get(1, 5, 1, 9),
                        errors: vec![]
                    },
                    with: None,
                    properties: vec![Property {
                        base: BaseNode {
                            location: loc.get(1, 6, 1, 7),
                            errors: vec!["missing property value".to_string()]
                        },
                        key: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 7),
                                errors: vec![]
                            },
                            name: "a".to_string()
                        }),
                        value: None
                    }]
                }))
            }))]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 19),
                    errors: vec![]
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 2),
                        errors: vec![]
                    },
                    name: "o".to_string()
                },
                init: Expression::Object(Box::new(ObjectExpr {
                    base: BaseNode {
                        location: loc.get(1, 5, 1, 19),
                        errors: vec![]
                    },
                    with: None,
                    properties: vec![
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 14),
                                errors: vec![]
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 7),
                                    errors: vec![]
                                },
                                name: "a".to_string()
                            }),
                            // TODO(affo): ast.Check would add the error "expected an operator between two expressions".
                            value: Some(Expression::Binary(Box::new(BinaryExpr {
                                base: BaseNode {
                                    location: loc.get(1, 9, 1, 14),
                                    errors: vec![]
                                },
                                operator: Operator::InvalidOperator,
                                left: Expression::StringLit(StringLit {
                                    base: BaseNode {
                                        location: loc.get(1, 9, 1, 12),
                                        errors: vec![]
                                    },
                                    value: "a".to_string()
                                }),
                                right: Expression::Identifier(Identifier {
                                    base: BaseNode {
                                        location: loc.get(1, 13, 1, 14),
                                        errors: vec![]
                                    },
                                    name: "b".to_string()
                                })
                            })))
                        },
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 14, 1, 18),
                                errors: vec!["missing property key".to_string()]
                            },
                            key: PropertyKey::StringLit(StringLit {
                                base: BaseNode {
                                    location: loc.get(1, 14, 1, 14),
                                    errors: vec![]
                                },
                                value: "<invalid>".to_string()
                            }),
                            value: Some(Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 16, 1, 18),
                                    errors: vec![
                                        "expected comma in property list, got COLON".to_string()
                                    ]
                                },
                                value: 30
                            }))
                        }
                    ]
                }))
            }))]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 14),
                    errors: vec![]
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 2),
                        errors: vec![]
                    },
                    name: "o".to_string()
                },
                init: Expression::Object(Box::new(ObjectExpr {
                    base: BaseNode {
                        location: loc.get(1, 5, 1, 14),
                        errors: vec![]
                    },
                    with: None,
                    properties: vec![Property {
                        base: BaseNode {
                            location: loc.get(1, 6, 1, 12),
                            errors: vec![]
                        },
                        key: PropertyKey::Identifier(Identifier {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 7),
                                errors: vec![]
                            },
                            name: "a".to_string()
                        }),
                        value: Some(Expression::StringLit(StringLit {
                            base: BaseNode {
                                location: loc.get(1, 9, 1, 12),
                                errors: vec![]
                            },
                            value: "a".to_string()
                        }))
                    }]
                }))
            }))]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Variable(Box::new(VariableAssgn {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 23),
                    errors: vec![]
                },
                id: Identifier {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 2),
                        errors: vec![]
                    },
                    name: "o".to_string()
                },
                init: Expression::Object(Box::new(ObjectExpr {
                    base: BaseNode {
                        location: loc.get(1, 5, 1, 23),
                        errors: vec![]
                    },
                    with: None,
                    properties: vec![
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 6, 1, 12),
                                errors: vec![]
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 6, 1, 7),
                                    errors: vec![]
                                },
                                name: "a".to_string()
                            }),
                            value: Some(Expression::StringLit(StringLit {
                                base: BaseNode {
                                    location: loc.get(1, 9, 1, 12),
                                    errors: vec![]
                                },
                                value: "a".to_string()
                            }))
                        },
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 14, 1, 16),
                                errors: vec![
                                    "unexpected token for property key: INT (30)".to_string()
                                ]
                            },
                            key: PropertyKey::StringLit(StringLit {
                                base: BaseNode {
                                    location: loc.get(1, 14, 1, 14),
                                    errors: vec![]
                                },
                                value: "<invalid>".to_string()
                            }),
                            value: None
                        },
                        Property {
                            base: BaseNode {
                                location: loc.get(1, 18, 1, 22),
                                errors: vec![]
                            },
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode {
                                    location: loc.get(1, 18, 1, 19),
                                    errors: vec![]
                                },
                                name: "b".to_string()
                            }),
                            value: Some(Expression::Integer(IntegerLit {
                                base: BaseNode {
                                    location: loc.get(1, 21, 1, 22),
                                    errors: vec![]
                                },
                                value: 7
                            }))
                        }
                    ]
                }))
            }))]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 6),
                    errors: vec![]
                },
                expression: Expression::Array(Box::new(ArrayExpr {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 6),
                        errors: vec![]
                    },
                    elements: vec![Expression::Identifier(Identifier {
                        base: BaseNode {
                            location: loc.get(1, 3, 1, 4),
                            errors: vec![]
                        },
                        name: "a".to_string()
                    })]
                }))
            })]
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
                errors: vec![]
            },
            name: "".to_string(),
            package: None,
            imports: vec![],
            body: vec![Statement::Expr(ExprStmt {
                base: BaseNode {
                    location: loc.get(1, 1, 1, 31),
                    errors: vec![] },
                expression: Expression::Integer(IntegerLit {
                    base: BaseNode {
                        location: loc.get(1, 1, 1, 31),
                        errors: vec!["invalid integer literal \"100000000000000000000000000000\": value out of range".to_string()]
                    },
                    value: 0,
                })
            })]
        },
    )
}
