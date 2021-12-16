use fluxcore::{
    ast,
    semantic::{
        import::Packages,
        nodes::*,
        types::{Function, MonoType, SemanticMap, Tvar},
        walk::{walk_mut, NodeMut},
        Analyzer,
    },
};
use pretty_assertions::assert_eq;

#[test]
fn analyze_end_to_end() {
    let mut analyzer = Analyzer::new_with_defaults(Default::default(), Packages::new());
    let (_, mut got) = analyzer
        .analyze_source(
            "main".to_string(),
            "main.flux".to_string(),
            r#"
n = 1
s = "string"
f = (a) => a + a
f(a: n)
f(a: s)
        "#,
        )
        .unwrap();
    let f_type = Function {
        req: fluxcore::semantic_map! {
            "a".to_string() => MonoType::Var(Tvar(4)),
        },
        opt: SemanticMap::new(),
        pipe: None,
        retn: MonoType::Var(Tvar(4)),
    };
    let f_call_int_type = Function {
        req: fluxcore::semantic_map! {
            "a".to_string() => MonoType::INT,
        },
        opt: SemanticMap::new(),
        pipe: None,
        retn: MonoType::INT,
    };
    let f_call_string_type = Function {
        req: fluxcore::semantic_map! {
            "a".to_string() => MonoType::STRING,
        },
        opt: SemanticMap::new(),
        pipe: None,
        retn: MonoType::STRING,
    };
    let want = Package {
        loc: ast::BaseNode::default().location,
        package: "main".to_string(),
        files: vec![File {
            loc: ast::BaseNode::default().location,
            package: None,
            imports: Vec::new(),
            body: vec![
                Statement::Variable(Box::new(VariableAssgn::new(
                    Identifier {
                        loc: ast::BaseNode::default().location,
                        name: Symbol::from("n@main"),
                    },
                    Expression::Integer(IntegerLit {
                        loc: ast::BaseNode::default().location,
                        value: 1,
                    }),
                    ast::BaseNode::default().location,
                ))),
                Statement::Variable(Box::new(VariableAssgn::new(
                    Identifier {
                        loc: ast::BaseNode::default().location,
                        name: Symbol::from("s@main"),
                    },
                    Expression::StringLit(StringLit {
                        loc: ast::BaseNode::default().location,
                        value: "string".to_string(),
                    }),
                    ast::BaseNode::default().location,
                ))),
                Statement::Variable(Box::new(VariableAssgn::new(
                    Identifier {
                        loc: ast::BaseNode::default().location,
                        name: Symbol::from("f@main"),
                    },
                    Expression::Function(Box::new(FunctionExpr {
                        loc: ast::BaseNode::default().location,
                        typ: MonoType::from(f_type),
                        params: vec![FunctionParameter {
                            loc: ast::BaseNode::default().location,
                            is_pipe: false,
                            key: Identifier {
                                loc: ast::BaseNode::default().location,
                                name: Symbol::from("a"),
                            },
                            default: None,
                        }],
                        body: Block::Return(ReturnStmt {
                            loc: ast::BaseNode::default().location,
                            argument: Expression::Binary(Box::new(BinaryExpr {
                                loc: ast::BaseNode::default().location,
                                typ: MonoType::Var(Tvar(4)),
                                operator: ast::Operator::AdditionOperator,
                                left: Expression::Identifier(IdentifierExpr {
                                    loc: ast::BaseNode::default().location,
                                    typ: MonoType::Var(Tvar(4)),
                                    name: Symbol::from("a"),
                                }),
                                right: Expression::Identifier(IdentifierExpr {
                                    loc: ast::BaseNode::default().location,
                                    typ: MonoType::Var(Tvar(4)),
                                    name: Symbol::from("a"),
                                }),
                            })),
                        }),
                        vectorized: None,
                    })),
                    ast::BaseNode::default().location,
                ))),
                Statement::Expr(ExprStmt {
                    loc: ast::BaseNode::default().location,
                    expression: Expression::Call(Box::new(CallExpr {
                        loc: ast::BaseNode::default().location,
                        typ: MonoType::INT,
                        pipe: None,
                        callee: Expression::Identifier(IdentifierExpr {
                            loc: ast::BaseNode::default().location,
                            typ: MonoType::from(f_call_int_type),
                            name: Symbol::from("f@main"),
                        }),
                        arguments: vec![Property {
                            loc: ast::BaseNode::default().location,
                            key: Identifier {
                                loc: ast::BaseNode::default().location,
                                name: Symbol::from("a"),
                            },
                            value: Expression::Identifier(IdentifierExpr {
                                loc: ast::BaseNode::default().location,
                                typ: MonoType::INT,
                                name: Symbol::from("n@main"),
                            }),
                        }],
                    })),
                }),
                Statement::Expr(ExprStmt {
                    loc: ast::BaseNode::default().location,
                    expression: Expression::Call(Box::new(CallExpr {
                        loc: ast::BaseNode::default().location,
                        typ: MonoType::STRING,
                        pipe: None,
                        callee: Expression::Identifier(IdentifierExpr {
                            loc: ast::BaseNode::default().location,
                            typ: MonoType::from(f_call_string_type),
                            name: Symbol::from("f@main"),
                        }),
                        arguments: vec![Property {
                            loc: ast::BaseNode::default().location,
                            key: Identifier {
                                loc: ast::BaseNode::default().location,
                                name: Symbol::from("a"),
                            },
                            value: Expression::Identifier(IdentifierExpr {
                                loc: ast::BaseNode::default().location,
                                typ: MonoType::STRING,
                                name: Symbol::from("s@main"),
                            }),
                        }],
                    })),
                }),
            ],
        }],
    };
    // We don't want to test the locations, so we override those with the base one.
    walk_mut(
        &mut |n: &mut NodeMut| n.set_loc(ast::BaseNode::default().location),
        &mut NodeMut::Package(&mut got),
    );
    assert_eq!(want, got);
}
