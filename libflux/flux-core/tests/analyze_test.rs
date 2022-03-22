use std::collections::BTreeMap;

use fluxcore::{
    ast,
    semantic::{
        import::Packages,
        nodes::*,
        types::{Function, MonoType, SemanticMap, Tvar},
        walk::{self, walk_mut, NodeMut},
        Analyzer,
    },
};
use pretty_assertions::assert_eq;

fn collect_symbols(pkg: &Package) -> BTreeMap<String, Symbol> {
    let mut map = BTreeMap::new();

    walk::walk(
        &mut |node| {
            let symbol = match node {
                walk::Node::Identifier(id) => &id.name,
                walk::Node::ImportDeclaration(import) => &import.import_symbol,
                walk::Node::IdentifierExpr(id) => &id.name,
                walk::Node::MemberExpr(member) => &member.property,
                _ => return,
            };

            if !map.contains_key(symbol.full_name()) {
                map.insert(symbol.full_name().to_string(), symbol.clone());
            }
        },
        walk::Node::Package(pkg),
    );

    map
}

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
    let symbols = collect_symbols(&got);
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
                        name: symbols["n@main"].clone(),
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
                        name: symbols["s@main"].clone(),
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
                        name: symbols["f@main"].clone(),
                    },
                    Expression::Function(Box::new(FunctionExpr {
                        loc: ast::BaseNode::default().location,
                        typ: MonoType::from(f_type),
                        params: vec![FunctionParameter {
                            loc: ast::BaseNode::default().location,
                            is_pipe: false,
                            key: Identifier {
                                loc: ast::BaseNode::default().location,
                                name: symbols["a"].clone(),
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
                                    name: symbols["a"].clone(),
                                }),
                                right: Expression::Identifier(IdentifierExpr {
                                    loc: ast::BaseNode::default().location,
                                    typ: MonoType::Var(Tvar(4)),
                                    name: symbols["a"].clone(),
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
                            name: symbols["f@main"].clone(),
                        }),
                        arguments: vec![Property {
                            loc: ast::BaseNode::default().location,
                            key: Identifier {
                                loc: ast::BaseNode::default().location,
                                name: symbols["a"].clone(),
                            },
                            value: Expression::Identifier(IdentifierExpr {
                                loc: ast::BaseNode::default().location,
                                typ: MonoType::INT,
                                name: symbols["n@main"].clone(),
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
                            name: symbols["f@main"].clone(),
                        }),
                        arguments: vec![Property {
                            loc: ast::BaseNode::default().location,
                            key: Identifier {
                                loc: ast::BaseNode::default().location,
                                name: symbols["a"].clone(),
                            },
                            value: Expression::Identifier(IdentifierExpr {
                                loc: ast::BaseNode::default().location,
                                typ: MonoType::STRING,
                                name: symbols["s@main"].clone(),
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
        NodeMut::Package(&mut got),
    );
    assert_eq!(want, got);
}
