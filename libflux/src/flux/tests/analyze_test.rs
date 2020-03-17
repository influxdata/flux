use flux::ast;
use flux::semantic::convert_source;
use flux::semantic::nodes::*;
use flux::semantic::types::{Function, MonoType, Tvar};
use flux::semantic::walk::{walk_mut, NodeMut};
use maplit;
use std::collections::BTreeMap;

use pretty_assertions::assert_eq;

#[test]
fn analyze_end_to_end() {
    let mut got = convert_source(
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
        req: maplit::btreemap! {
            "a".to_string() => MonoType::Var(Tvar(4)),
        },
        opt: BTreeMap::new(),
        pipe: None,
        retn: MonoType::Var(Tvar(4)),
    };
    let f_call_int_type = Function {
        req: maplit::btreemap! {
            "a".to_string() => MonoType::Int,
        },
        opt: BTreeMap::new(),
        pipe: None,
        retn: MonoType::Int,
    };
    let f_call_string_type = Function {
        req: maplit::btreemap! {
            "a".to_string() => MonoType::String,
        },
        opt: BTreeMap::new(),
        pipe: None,
        retn: MonoType::String,
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
                        name: "n".to_string(),
                    },
                    Expression::Integer(IntegerLit {
                        loc: ast::BaseNode::default().location,
                        typ: MonoType::Int,
                        value: 1,
                    }),
                    ast::BaseNode::default().location,
                ))),
                Statement::Variable(Box::new(VariableAssgn::new(
                    Identifier {
                        loc: ast::BaseNode::default().location,
                        name: "s".to_string(),
                    },
                    Expression::StringLit(StringLit {
                        loc: ast::BaseNode::default().location,
                        typ: MonoType::String,
                        value: "string".to_string(),
                    }),
                    ast::BaseNode::default().location,
                ))),
                Statement::Variable(Box::new(VariableAssgn::new(
                    Identifier {
                        loc: ast::BaseNode::default().location,
                        name: "f".to_string(),
                    },
                    Expression::Function(Box::new(FunctionExpr {
                        loc: ast::BaseNode::default().location,
                        typ: MonoType::Fun(Box::new(f_type)),
                        params: vec![FunctionParameter {
                            loc: ast::BaseNode::default().location,
                            is_pipe: false,
                            key: Identifier {
                                loc: ast::BaseNode::default().location,
                                name: "a".to_string(),
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
                                    name: "a".to_string(),
                                }),
                                right: Expression::Identifier(IdentifierExpr {
                                    loc: ast::BaseNode::default().location,
                                    typ: MonoType::Var(Tvar(4)),
                                    name: "a".to_string(),
                                }),
                            })),
                        }),
                    })),
                    ast::BaseNode::default().location,
                ))),
                Statement::Expr(ExprStmt {
                    loc: ast::BaseNode::default().location,
                    expression: Expression::Call(Box::new(CallExpr {
                        loc: ast::BaseNode::default().location,
                        typ: MonoType::Int,
                        pipe: None,
                        callee: Expression::Identifier(IdentifierExpr {
                            loc: ast::BaseNode::default().location,
                            typ: MonoType::Fun(Box::new(f_call_int_type)),
                            name: "f".to_string(),
                        }),
                        arguments: vec![Property {
                            loc: ast::BaseNode::default().location,
                            key: Identifier {
                                loc: ast::BaseNode::default().location,
                                name: "a".to_string(),
                            },
                            value: Expression::Identifier(IdentifierExpr {
                                loc: ast::BaseNode::default().location,
                                typ: MonoType::Int,
                                name: "n".to_string(),
                            }),
                        }],
                    })),
                }),
                Statement::Expr(ExprStmt {
                    loc: ast::BaseNode::default().location,
                    expression: Expression::Call(Box::new(CallExpr {
                        loc: ast::BaseNode::default().location,
                        typ: MonoType::String,
                        pipe: None,
                        callee: Expression::Identifier(IdentifierExpr {
                            loc: ast::BaseNode::default().location,
                            typ: MonoType::Fun(Box::new(f_call_string_type)),
                            name: "f".to_string(),
                        }),
                        arguments: vec![Property {
                            loc: ast::BaseNode::default().location,
                            key: Identifier {
                                loc: ast::BaseNode::default().location,
                                name: "a".to_string(),
                            },
                            value: Expression::Identifier(IdentifierExpr {
                                loc: ast::BaseNode::default().location,
                                typ: MonoType::String,
                                name: "s".to_string(),
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
