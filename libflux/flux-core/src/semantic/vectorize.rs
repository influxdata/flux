use std::collections::HashMap;

use crate::{
    ast::{Operator, SourceLocation},
    errors::{located, Errors},
    semantic::{
        nodes::{
            BinaryExpr, Block, CallExpr, ConditionalExpr, Error, ErrorKind, Expression,
            FunctionExpr, Identifier, IdentifierExpr, LogicalExpr, MemberExpr, ObjectExpr, Package,
            Property, Result, ReturnStmt,
        },
        types::{self, Function, Label, MonoType},
        AnalyzerConfig, Feature, Symbol,
    },
};

/// The name of a function used to signal to the runtime that the value in the
/// `v` parameter should be rewritten as a "vector repeat" value.
const VEC_REPEAT_FN: &str = "~~vecRepeat~~";

/// Vectorizes a pkg
pub fn vectorize(
    config: &AnalyzerConfig,
    pkg: &mut Package,
) -> std::result::Result<(), Errors<Error>> {
    use crate::semantic::walk::{walk_mut, NodeMut, VisitorMut};
    struct Vectorizer<'a> {
        #[allow(dead_code)]
        config: &'a AnalyzerConfig,
        errors: Errors<Error>,
    }
    impl VisitorMut for Vectorizer<'_> {
        fn visit(&mut self, _node: &mut NodeMut) -> bool {
            true
        }

        fn done(&mut self, node: &mut NodeMut) {
            if let NodeMut::FunctionExpr(function) = node {
                match function.vectorize(self.config) {
                    Ok(vectorized) => function.vectorized = Some(Box::new(vectorized)),
                    Err(err) => self.errors.push(err),
                }
            }
        }
    }

    let mut visitor = Vectorizer {
        config,
        errors: Errors::new(),
    };
    walk_mut(&mut visitor, NodeMut::Package(pkg));
    if visitor.errors.has_errors() {
        Err(visitor.errors)
    } else {
        Ok(())
    }
}

struct VectorizeEnv<'a> {
    #[allow(dead_code)]
    config: &'a AnalyzerConfig,
    symbols: HashMap<Symbol, MonoType>,
}

impl Expression {
    fn vectorize(&self, env: &VectorizeEnv<'_>) -> Result<Self> {
        Ok(match self {
            Expression::Identifier(identifier) => {
                Expression::Identifier(identifier.vectorize(env)?)
            }
            Expression::Member(member) => {
                let object = member.object.vectorize(env)?;
                let typ = object.type_of();
                Expression::Member(Box::new(MemberExpr {
                    loc: member.loc.clone(),
                    typ: typ
                        .field(&member.property)
                        .ok_or_else(|| {
                            located(
                                member.object.loc().clone(),
                                ErrorKind::UnableToVectorize(format!(
                                    "Expected record type, got `{}`",
                                    typ
                                )),
                            )
                        })?
                        .v
                        .clone(),
                    object,
                    property: member.property.clone(),
                }))
            }
            Expression::Binary(binary) => {
                let left = binary.left.vectorize(env)?;
                let right = binary.right.vectorize(env)?;

                if !op_is_vectorizable(&binary.operator) {
                    return Err(located(
                        binary.loc.clone(),
                        ErrorKind::UnableToVectorize(format!(
                            "unsupported operator {}",
                            binary.operator
                        )),
                    ));
                }

                Expression::Binary(Box::new(BinaryExpr {
                    loc: binary.loc.clone(),
                    typ: MonoType::vector(binary.typ.clone()),
                    operator: binary.operator.clone(),
                    left,
                    right,
                }))
            }
            Expression::Logical(expr) => {
                let left = expr.left.vectorize(env)?;
                let right = expr.right.vectorize(env)?;
                Expression::Logical(Box::new(LogicalExpr {
                    loc: expr.loc.clone(),
                    typ: MonoType::vector(expr.typ.clone()),
                    operator: expr.operator.clone(),
                    left,
                    right,
                }))
            }
            Expression::Conditional(expr) => Expression::Conditional(Box::new(ConditionalExpr {
                loc: expr.loc.clone(),
                test: expr.test.vectorize(env)?,
                consequent: expr.consequent.vectorize(env)?,
                alternate: expr.alternate.vectorize(env)?,
                typ: MonoType::vector(expr.typ.clone()),
            })),
            expr @ Expression::Integer(_)
                if env.config.features.contains(&Feature::VectorizedConst) =>
            {
                wrap_vec_repeat(expr.clone())
            }
            expr @ Expression::DateTime(_)
                if env.config.features.contains(&Feature::VectorizedConst) =>
            {
                wrap_vec_repeat(expr.clone())
            }
            expr @ Expression::Float(_)
                if env.config.features.contains(&Feature::VectorizedConst) =>
            {
                wrap_vec_repeat(expr.clone())
            }
            expr @ Expression::StringLit(_)
                if env.config.features.contains(&Feature::VectorizedConst) =>
            {
                wrap_vec_repeat(expr.clone())
            }
            _ => {
                return Err(located(
                    self.loc().clone(),
                    ErrorKind::UnableToVectorize("Unable to vectorize expression".into()),
                ));
            }
        })
    }
}

/// Check to see if a given operator is vectorizable.
fn op_is_vectorizable(op: &Operator) -> bool {
    // Note that only certain operators can be vectorized today.
    // See `array/binary.tmpldata` for the currently supported ops.
    // As new ops are implemented, this match should be updated.
    matches!(
        op,
        Operator::AdditionOperator
            | Operator::SubtractionOperator
            | Operator::MultiplicationOperator
            | Operator::DivisionOperator
            | Operator::ModuloOperator
            | Operator::PowerOperator
    )
}

fn wrap_vec_repeat(expr: Expression) -> Expression {
    // The call expression is just a way to trigger a rewrite in Go during evaluation.
    // The only details that matter are the parameter `v` (the original expression) and
    // the name of the function.
    let typ = expr.type_of();
    let call = CallExpr {
        loc: expr.loc().clone(),
        typ: MonoType::vector(typ.clone()),
        pipe: None,
        callee: Expression::Identifier(IdentifierExpr {
            loc: SourceLocation::default(),
            typ,
            name: Symbol::from(VEC_REPEAT_FN),
        }),
        arguments: vec![Property {
            loc: SourceLocation::default(),
            key: Identifier {
                loc: SourceLocation::default(),
                name: Symbol::from("v"),
            },
            value: expr.clone(),
        }],
    };
    Expression::Call(Box::new(call))
}

impl IdentifierExpr {
    fn vectorize(&self, env: &VectorizeEnv<'_>) -> Result<Self> {
        let typ = env
            .symbols
            .get(&self.name)
            .ok_or_else(|| {
                located(
                    self.loc.clone(),
                    ErrorKind::UnableToVectorize(format!(
                        "Unable to vectorize non-vector symbol `{}`",
                        self.name
                    )),
                )
            })?
            .clone();

        Ok(IdentifierExpr {
            loc: self.loc.clone(),
            typ,
            name: self.name.clone(),
        })
    }
}

impl FunctionExpr {
    fn vectorize(&self, config: &AnalyzerConfig) -> Result<Self> {
        if self.params.len() == 1 && self.params[0].key.name == "r" {
            fn vectorize_fields(record: &MonoType) -> MonoType {
                use crate::semantic::types::Record;
                match record {
                    MonoType::Record(record) => MonoType::from(match &**record {
                        Record::Empty => Record::Empty,
                        Record::Extension { head, tail } => Record::Extension {
                            head: types::Property {
                                k: head.k.clone(),
                                v: MonoType::vector(head.v.clone()),
                            },
                            tail: vectorize_fields(tail),
                        },
                    }),
                    _ => record.clone(),
                }
            }
            let params: Vec<_> = self
                .params
                .iter()
                .map(|param| {
                    let parameter_type =
                        vectorize_fields(self.typ.parameter(&param.key.name).unwrap());
                    (param.key.name.clone(), parameter_type)
                })
                .collect();
            let env = VectorizeEnv {
                config,
                symbols: params.iter().cloned().collect(),
            };

            let body = match &self.body {
                Block::Variable(..) | Block::Expr(..) => {
                    return Err(located(
                        self.body.loc().clone(),
                        ErrorKind::UnableToVectorize("Unable to vectorize statements".into()),
                    ));
                }
                // XXX: sean (January 14 2022) - The only type of function expression
                // currently supported for vectorization is one whose body contains only
                // a single object expression, the fields of which only reference members of
                // `r` and do not include any kind of operation, literal, or logical expression.
                //
                // We may support other expression types in the future.
                Block::Return(e) => {
                    let argument = match &e.argument {
                        Expression::Object(e) => {
                            let properties = e
                                .properties
                                .iter()
                                .map(|p| {
                                    Ok(Property {
                                        loc: p.loc.clone(),
                                        key: p.key.clone(),
                                        value: p.value.vectorize(&env)?,
                                    })
                                })
                                .collect::<Result<Vec<_>>>()?;

                            let with = e
                                .with
                                .as_ref()
                                .map(|with| with.vectorize(&env))
                                .transpose()?;

                            Expression::Object(Box::new(ObjectExpr {
                                loc: e.loc.clone(),
                                typ: MonoType::from(types::Record::new(
                                    properties.iter().map(|p| types::Property {
                                        k: Label::from(p.key.name.clone()).into(),
                                        v: p.value.type_of(),
                                    }),
                                    with.as_ref().map(|with| with.typ.clone()),
                                )),
                                with,
                                properties,
                            }))
                        }
                        _ => {
                            return Err(located(
                                e.argument.loc().clone(),
                                ErrorKind::UnableToVectorize(
                                    "Vectorization only supports returning a record".into(),
                                ),
                            ));
                        }
                    };
                    Block::Return(ReturnStmt {
                        loc: e.loc.clone(),
                        argument,
                    })
                }
            };
            Ok(FunctionExpr {
                loc: self.loc.clone(),
                typ: MonoType::from(Function {
                    pipe: None,
                    req: params
                        .into_iter()
                        .map(|(key, value)| (key.to_string(), value))
                        .collect(),
                    opt: Default::default(),
                    retn: body.type_of(),
                }),
                params: self.params.clone(),
                body,
                vectorized: None,
            })
        } else {
            // Only `map` will get vectorized to start with, so only try to vectorize such functions
            Err(located(
                self.loc.clone(),
                ErrorKind::UnableToVectorize("Does not match the `map` signature".into()),
            ))
        }
    }
}
