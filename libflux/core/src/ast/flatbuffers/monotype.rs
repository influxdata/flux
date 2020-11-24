//! Flatbuffer serialization for MonoType AST nodes
//!
use crate::ast::flatbuffers::ast_generated::fbast as fb;

#[rustfmt::skip]
use crate::ast::{
    SourceLocation,
    BaseNode,
    TypeExpression,
    TypeConstraint,
    MonoType,
    Identifier,
    NamedType,
    TvarType,
    ArrayType,
    PropertyType,
    RecordType,
    ParameterType,
    FunctionType,
};

fn build_vec<T, S, F, B>(v: Vec<T>, b: &mut B, f: F) -> Vec<S>
where
    F: Fn(&mut B, T) -> S,
{
    let mut mapped = Vec::new();
    for t in v {
        mapped.push(f(b, t));
    }
    mapped
}

fn build_base_node<'a>(
    builder: &mut flatbuffers::FlatBufferBuilder<'a>,
    base_node: BaseNode,
) -> flatbuffers::WIPOffset<fb::BaseNode<'a>> {
    let loc = Some(build_loc(builder, base_node.location));
    let errors = build_vec(base_node.errors, builder, |builder, s| {
        builder.create_string(&s)
    });
    let errors = Some(builder.create_vector(errors.as_slice()));
    fb::BaseNode::create(builder, &fb::BaseNodeArgs { loc, errors })
}

fn build_loc<'a>(
    builder: &mut flatbuffers::FlatBufferBuilder<'a>,
    loc: SourceLocation,
) -> flatbuffers::WIPOffset<fb::SourceLocation<'a>> {
    let file = match loc.file {
        None => None,
        Some(name) => Some(builder.create_string(&name)),
    };
    let source = match loc.source {
        None => None,
        Some(src) => Some(builder.create_string(&src)),
    };
    fb::SourceLocation::create(
        builder,
        &fb::SourceLocationArgs {
            file,
            start: Some(&fb::Position::new(
                loc.start.line as i32,
                loc.start.column as i32,
            )),
            end: Some(&fb::Position::new(
                loc.end.line as i32,
                loc.end.column as i32,
            )),
            source,
        },
    )
}

fn build_type_expression<'a>(
    builder: &mut flatbuffers::FlatBufferBuilder<'a>,
    expr: TypeExpression,
) -> flatbuffers::WIPOffset<fb::TypeExpression<'a>> {
    let base_node = build_base_node(builder, expr.base);
    let (offset, t) = build_monotype(builder, expr.monotype);
    let constraints = build_vec(expr.constraints, builder, build_type_constraint);
    let constraints = builder.create_vector(constraints.as_slice());
    fb::TypeExpression::create(
        builder,
        &fb::TypeExpressionArgs {
            base_node: Some(base_node),
            monotype: Some(offset),
            monotype_type: t,
            constraints: Some(constraints),
        },
    )
}

fn build_type_constraint<'a>(
    builder: &mut flatbuffers::FlatBufferBuilder<'a>,
    c: TypeConstraint,
) -> flatbuffers::WIPOffset<fb::TypeConstraint<'a>> {
    let base_node = build_base_node(builder, c.base);
    let tvar = build_identifier(builder, c.tvar);
    let kinds = build_vec(c.kinds, builder, build_identifier);
    let kinds = builder.create_vector(kinds.as_slice());
    fb::TypeConstraint::create(
        builder,
        &fb::TypeConstraintArgs {
            base_node: Some(base_node),
            tvar: Some(tvar),
            kinds: Some(kinds),
        },
    )
}

fn build_monotype<'a>(
    builder: &mut flatbuffers::FlatBufferBuilder<'a>,
    t: MonoType,
) -> (
    flatbuffers::WIPOffset<flatbuffers::UnionWIPOffset>,
    fb::MonoType,
) {
    match t {
        MonoType::Basic(t) => {
            let offset = build_named_type(builder, t);
            (offset.as_union_value(), fb::MonoType::NamedType)
        }
        MonoType::Tvar(t) => {
            let offset = build_tvar_type(builder, t);
            (offset.as_union_value(), fb::MonoType::TvarType)
        }
        MonoType::Array(t) => {
            let offset = build_array_type(builder, *t);
            (offset.as_union_value(), fb::MonoType::ArrayType)
        }
        MonoType::Record(t) => {
            let offset = build_record_type(builder, t);
            (offset.as_union_value(), fb::MonoType::RecordType)
        }
        MonoType::Function(t) => {
            let offset = build_function_type(builder, *t);
            (offset.as_union_value(), fb::MonoType::FunctionType)
        }
    }
}

fn build_identifier<'a>(
    builder: &mut flatbuffers::FlatBufferBuilder<'a>,
    id: Identifier,
) -> flatbuffers::WIPOffset<fb::Identifier<'a>> {
    let base_node = Some(build_base_node(builder, id.base));
    let name = Some(builder.create_string(&id.name));
    fb::Identifier::create(builder, &fb::IdentifierArgs { base_node, name })
}

fn build_named_type<'a>(
    builder: &mut flatbuffers::FlatBufferBuilder<'a>,
    t: NamedType,
) -> flatbuffers::WIPOffset<fb::NamedType<'a>> {
    let base_node = Some(build_base_node(builder, t.base));
    let id = Some(build_identifier(builder, t.name));
    fb::NamedType::create(builder, &fb::NamedTypeArgs { base_node, id })
}

fn build_tvar_type<'a>(
    builder: &mut flatbuffers::FlatBufferBuilder<'a>,
    t: TvarType,
) -> flatbuffers::WIPOffset<fb::TvarType<'a>> {
    let base_node = Some(build_base_node(builder, t.base));
    let id = Some(build_identifier(builder, t.name));
    fb::TvarType::create(builder, &fb::TvarTypeArgs { base_node, id })
}

fn build_array_type<'a>(
    builder: &mut flatbuffers::FlatBufferBuilder<'a>,
    a: ArrayType,
) -> flatbuffers::WIPOffset<fb::ArrayType<'a>> {
    let base_node = build_base_node(builder, a.base);
    let (offset, t) = build_monotype(builder, a.element);
    fb::ArrayType::create(
        builder,
        &fb::ArrayTypeArgs {
            base_node: Some(base_node),
            element: Some(offset),
            element_type: t,
        },
    )
}

fn build_record_type<'a>(
    builder: &mut flatbuffers::FlatBufferBuilder<'a>,
    r: RecordType,
) -> flatbuffers::WIPOffset<fb::RecordType<'a>> {
    let base_node = Some(build_base_node(builder, r.base));
    let tvar = match r.tvar {
        None => None,
        Some(id) => Some(build_identifier(builder, id)),
    };
    let properties = build_vec(r.properties, builder, build_property_type);
    let properties = Some(builder.create_vector(properties.as_slice()));
    fb::RecordType::create(
        builder,
        &fb::RecordTypeArgs {
            base_node,
            tvar,
            properties,
        },
    )
}

fn build_property_type<'a>(
    builder: &mut flatbuffers::FlatBufferBuilder<'a>,
    p: PropertyType,
) -> flatbuffers::WIPOffset<fb::PropertyType<'a>> {
    let base_node = build_base_node(builder, p.base);
    let id = build_identifier(builder, p.name);
    let (offset, t) = build_monotype(builder, p.monotype);
    fb::PropertyType::create(
        builder,
        &fb::PropertyTypeArgs {
            base_node: Some(base_node),
            id: Some(id),
            monotype: Some(offset),
            monotype_type: t,
        },
    )
}

fn build_function_type<'a>(
    builder: &mut flatbuffers::FlatBufferBuilder<'a>,
    f: FunctionType,
) -> flatbuffers::WIPOffset<fb::FunctionType<'a>> {
    let base_node = Some(build_base_node(builder, f.base));
    let parameters = build_vec(f.parameters, builder, build_parameter_type);
    let parameters = Some(builder.create_vector(parameters.as_slice()));
    let (offset, t) = build_monotype(builder, f.monotype);
    fb::FunctionType::create(
        builder,
        &fb::FunctionTypeArgs {
            base_node,
            parameters,
            monotype: Some(offset),
            monotype_type: t,
        },
    )
}

fn build_parameter_type<'a>(
    builder: &mut flatbuffers::FlatBufferBuilder<'a>,
    p: ParameterType,
) -> flatbuffers::WIPOffset<fb::ParameterType<'a>> {
    match p {
        ParameterType::Required {
            base,
            name,
            monotype,
        } => {
            let base_node = build_base_node(builder, base);
            let id = build_identifier(builder, name);
            let (offset, t) = build_monotype(builder, monotype);
            fb::ParameterType::create(
                builder,
                &fb::ParameterTypeArgs {
                    base_node: Some(base_node),
                    id: Some(id),
                    monotype: Some(offset),
                    monotype_type: t,
                    kind: fb::ParameterKind::Required,
                },
            )
        }
        ParameterType::Optional {
            base,
            name,
            monotype,
        } => {
            let base_node = build_base_node(builder, base);
            let id = build_identifier(builder, name);
            let (offset, t) = build_monotype(builder, monotype);
            fb::ParameterType::create(
                builder,
                &fb::ParameterTypeArgs {
                    base_node: Some(base_node),
                    id: Some(id),
                    monotype: Some(offset),
                    monotype_type: t,
                    kind: fb::ParameterKind::Optional,
                },
            )
        }
        ParameterType::Pipe {
            base,
            name: Some(id),
            monotype,
        } => {
            let base_node = build_base_node(builder, base);
            let id = build_identifier(builder, id);
            let (offset, t) = build_monotype(builder, monotype);
            fb::ParameterType::create(
                builder,
                &fb::ParameterTypeArgs {
                    base_node: Some(base_node),
                    id: Some(id),
                    monotype: Some(offset),
                    monotype_type: t,
                    kind: fb::ParameterKind::Pipe,
                },
            )
        }
        ParameterType::Pipe {
            base,
            name: None,
            monotype,
        } => {
            let base_node = build_base_node(builder, base);
            let (offset, t) = build_monotype(builder, monotype);
            fb::ParameterType::create(
                builder,
                &fb::ParameterTypeArgs {
                    base_node: Some(base_node),
                    id: None,
                    monotype: Some(offset),
                    monotype_type: t,
                    kind: fb::ParameterKind::Pipe,
                },
            )
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn monotype_from_table(table: flatbuffers::Table, t: fb::MonoType) -> MonoType {
        match t {
            fb::MonoType::NamedType => {
                MonoType::Basic(fb::NamedType::init_from_table(table).into())
            }
            fb::MonoType::TvarType => MonoType::Tvar(fb::TvarType::init_from_table(table).into()),
            fb::MonoType::ArrayType => {
                MonoType::Array(Box::new(fb::ArrayType::init_from_table(table).into()))
            }
            fb::MonoType::RecordType => {
                MonoType::Record(fb::RecordType::init_from_table(table).into())
            }
            fb::MonoType::FunctionType => {
                MonoType::Function(Box::new(fb::FunctionType::init_from_table(table).into()))
            }
            fb::MonoType::NONE => unimplemented!("cannot convert fb::MonoType::NONE"),
        }
    }
    impl From<fb::Identifier<'_>> for Identifier {
        fn from(id: fb::Identifier) -> Identifier {
            Identifier {
                base: BaseNode::default(),
                name: id.name().unwrap().to_string(),
            }
        }
    }
    impl From<fb::NamedType<'_>> for NamedType {
        fn from(t: fb::NamedType) -> NamedType {
            NamedType {
                base: BaseNode::default(),
                name: t.id().unwrap().into(),
            }
        }
    }
    impl From<fb::TvarType<'_>> for TvarType {
        fn from(t: fb::TvarType) -> TvarType {
            TvarType {
                base: BaseNode::default(),
                name: t.id().unwrap().into(),
            }
        }
    }
    impl From<fb::ArrayType<'_>> for ArrayType {
        fn from(t: fb::ArrayType) -> ArrayType {
            ArrayType {
                base: BaseNode::default(),
                element: monotype_from_table(t.element().unwrap(), t.element_type()),
            }
        }
    }
    impl From<fb::PropertyType<'_>> for PropertyType {
        fn from(p: fb::PropertyType) -> PropertyType {
            PropertyType {
                base: BaseNode::default(),
                name: p.id().unwrap().into(),
                monotype: monotype_from_table(p.monotype().unwrap(), p.monotype_type()),
            }
        }
    }
    impl From<fb::RecordType<'_>> for RecordType {
        fn from(t: fb::RecordType) -> RecordType {
            let mut properties = Vec::new();
            for p in t.properties().unwrap().iter() {
                properties.push(p.into());
            }
            RecordType {
                base: BaseNode::default(),
                tvar: match t.tvar() {
                    None => None,
                    Some(id) => Some(id.into()),
                },
                properties,
            }
        }
    }
    impl From<fb::ParameterType<'_>> for ParameterType {
        fn from(p: fb::ParameterType) -> ParameterType {
            match p.kind() {
                fb::ParameterKind::Required => ParameterType::Required {
                    base: BaseNode::default(),
                    name: p.id().unwrap().into(),
                    monotype: monotype_from_table(p.monotype().unwrap(), p.monotype_type()),
                },
                fb::ParameterKind::Optional => ParameterType::Optional {
                    base: BaseNode::default(),
                    name: p.id().unwrap().into(),
                    monotype: monotype_from_table(p.monotype().unwrap(), p.monotype_type()),
                },
                fb::ParameterKind::Pipe => ParameterType::Pipe {
                    base: BaseNode::default(),
                    name: match p.id() {
                        None => None,
                        Some(id) => Some(id.into()),
                    },
                    monotype: monotype_from_table(p.monotype().unwrap(), p.monotype_type()),
                },
            }
        }
    }
    impl From<fb::FunctionType<'_>> for FunctionType {
        fn from(t: fb::FunctionType) -> FunctionType {
            let mut parameters = Vec::new();
            for p in t.parameters().unwrap().iter() {
                parameters.push(p.into());
            }
            FunctionType {
                base: BaseNode::default(),
                parameters,
                monotype: monotype_from_table(t.monotype().unwrap(), t.monotype_type()),
            }
        }
    }
    impl From<fb::TypeConstraint<'_>> for TypeConstraint {
        fn from(c: fb::TypeConstraint) -> TypeConstraint {
            let mut kinds = Vec::new();
            for id in c.kinds().unwrap().iter() {
                kinds.push(id.into());
            }
            TypeConstraint {
                base: BaseNode::default(),
                tvar: c.tvar().unwrap().into(),
                kinds,
            }
        }
    }
    impl From<fb::TypeExpression<'_>> for TypeExpression {
        fn from(t: fb::TypeExpression) -> TypeExpression {
            let mut constraints = Vec::new();
            for c in t.constraints().unwrap().iter() {
                constraints.push(c.into());
            }
            TypeExpression {
                base: BaseNode::default(),
                monotype: monotype_from_table(t.monotype().unwrap(), t.monotype_type()),
                constraints,
            }
        }
    }

    fn serialize<'a, 'b, T, S, F>(
        builder: &'a mut flatbuffers::FlatBufferBuilder<'b>,
        t: T,
        f: F,
    ) -> &'a [u8]
    where
        F: Fn(&mut flatbuffers::FlatBufferBuilder<'b>, T) -> flatbuffers::WIPOffset<S>,
    {
        let offset = f(builder, t);
        builder.finish(offset, None);
        builder.finished_data()
    }

    #[test]
    fn test_named_type() {
        let want = NamedType {
            base: BaseNode::default(),
            name: Identifier {
                base: BaseNode::default(),
                name: "int".to_string(),
            },
        };

        let mut builder = flatbuffers::FlatBufferBuilder::new();

        assert_eq!(
            want,
            flatbuffers::get_root::<fb::NamedType>(serialize(
                &mut builder,
                want.clone(),
                build_named_type
            ))
            .into()
        );
    }
    #[test]
    fn test_tvar_type() {
        let want = TvarType {
            base: BaseNode::default(),
            name: Identifier {
                base: BaseNode::default(),
                name: "T".to_string(),
            },
        };

        let mut builder = flatbuffers::FlatBufferBuilder::new();

        assert_eq!(
            want,
            flatbuffers::get_root::<fb::TvarType>(serialize(
                &mut builder,
                want.clone(),
                build_tvar_type
            ))
            .into()
        );
    }
    #[test]
    fn test_array_type() {
        let want = ArrayType {
            base: BaseNode::default(),
            element: MonoType::Basic(NamedType {
                base: BaseNode::default(),
                name: Identifier {
                    base: BaseNode::default(),
                    name: "int".to_string(),
                },
            }),
        };

        let mut builder = flatbuffers::FlatBufferBuilder::new();

        assert_eq!(
            want,
            flatbuffers::get_root::<fb::ArrayType>(serialize(
                &mut builder,
                want.clone(),
                build_array_type,
            ))
            .into()
        );
    }
    #[test]
    fn test_empty_record_type() {
        let want = RecordType {
            base: BaseNode::default(),
            tvar: None,
            properties: vec![],
        };

        let mut builder = flatbuffers::FlatBufferBuilder::new();

        assert_eq!(
            want,
            flatbuffers::get_root::<fb::RecordType>(serialize(
                &mut builder,
                want.clone(),
                build_record_type,
            ))
            .into()
        );
    }
    #[test]
    fn test_record_type() {
        let want = RecordType {
            base: BaseNode::default(),
            tvar: None,
            properties: vec![
                PropertyType {
                    base: BaseNode::default(),
                    name: Identifier {
                        base: BaseNode::default(),
                        name: "a".to_string(),
                    },
                    monotype: MonoType::Basic(NamedType {
                        base: BaseNode::default(),
                        name: Identifier {
                            base: BaseNode::default(),
                            name: "int".to_string(),
                        },
                    }),
                },
                PropertyType {
                    base: BaseNode::default(),
                    name: Identifier {
                        base: BaseNode::default(),
                        name: "b".to_string(),
                    },
                    monotype: MonoType::Basic(NamedType {
                        base: BaseNode::default(),
                        name: Identifier {
                            base: BaseNode::default(),
                            name: "string".to_string(),
                        },
                    }),
                },
            ],
        };

        let mut builder = flatbuffers::FlatBufferBuilder::new();

        assert_eq!(
            want,
            flatbuffers::get_root::<fb::RecordType>(serialize(
                &mut builder,
                want.clone(),
                build_record_type,
            ))
            .into()
        );
    }
    #[test]
    fn test_record_extension_type() {
        let want = RecordType {
            base: BaseNode::default(),
            tvar: Some(Identifier {
                base: BaseNode::default(),
                name: "T".to_string(),
            }),
            properties: vec![
                PropertyType {
                    base: BaseNode::default(),
                    name: Identifier {
                        base: BaseNode::default(),
                        name: "a".to_string(),
                    },
                    monotype: MonoType::Basic(NamedType {
                        base: BaseNode::default(),
                        name: Identifier {
                            base: BaseNode::default(),
                            name: "int".to_string(),
                        },
                    }),
                },
                PropertyType {
                    base: BaseNode::default(),
                    name: Identifier {
                        base: BaseNode::default(),
                        name: "b".to_string(),
                    },
                    monotype: MonoType::Basic(NamedType {
                        base: BaseNode::default(),
                        name: Identifier {
                            base: BaseNode::default(),
                            name: "string".to_string(),
                        },
                    }),
                },
            ],
        };

        let mut builder = flatbuffers::FlatBufferBuilder::new();

        assert_eq!(
            want,
            flatbuffers::get_root::<fb::RecordType>(serialize(
                &mut builder,
                want.clone(),
                build_record_type,
            ))
            .into()
        );
    }
    #[test]
    fn test_function_type_no_params() {
        let want = FunctionType {
            base: BaseNode::default(),
            parameters: vec![],
            monotype: MonoType::Basic(NamedType {
                base: BaseNode::default(),
                name: Identifier {
                    base: BaseNode::default(),
                    name: "int".to_string(),
                },
            }),
        };

        let mut builder = flatbuffers::FlatBufferBuilder::new();

        assert_eq!(
            want,
            flatbuffers::get_root::<fb::FunctionType>(serialize(
                &mut builder,
                want.clone(),
                build_function_type,
            ))
            .into()
        );
    }
    #[test]
    fn test_function_type_many_params() {
        let want = FunctionType {
            base: BaseNode::default(),
            parameters: vec![
                ParameterType::Pipe {
                    base: BaseNode::default(),
                    name: Some(Identifier {
                        base: BaseNode::default(),
                        name: "tables".to_string(),
                    }),
                    monotype: MonoType::Array(Box::new(ArrayType {
                        base: BaseNode::default(),
                        element: MonoType::Basic(NamedType {
                            base: BaseNode::default(),
                            name: Identifier {
                                base: BaseNode::default(),
                                name: "int".to_string(),
                            },
                        }),
                    })),
                },
                ParameterType::Required {
                    base: BaseNode::default(),
                    name: Identifier {
                        base: BaseNode::default(),
                        name: "x".to_string(),
                    },
                    monotype: MonoType::Basic(NamedType {
                        base: BaseNode::default(),
                        name: Identifier {
                            base: BaseNode::default(),
                            name: "int".to_string(),
                        },
                    }),
                },
                ParameterType::Optional {
                    base: BaseNode::default(),
                    name: Identifier {
                        base: BaseNode::default(),
                        name: "y".to_string(),
                    },
                    monotype: MonoType::Basic(NamedType {
                        base: BaseNode::default(),
                        name: Identifier {
                            base: BaseNode::default(),
                            name: "bool".to_string(),
                        },
                    }),
                },
            ],
            monotype: MonoType::Basic(NamedType {
                base: BaseNode::default(),
                name: Identifier {
                    base: BaseNode::default(),
                    name: "int".to_string(),
                },
            }),
        };

        let mut builder = flatbuffers::FlatBufferBuilder::new();

        assert_eq!(
            want,
            flatbuffers::get_root::<fb::FunctionType>(serialize(
                &mut builder,
                want.clone(),
                build_function_type,
            ))
            .into()
        );
    }
    #[test]
    fn test_function_type_pipe_param() {
        let want = FunctionType {
            base: BaseNode::default(),
            parameters: vec![ParameterType::Pipe {
                base: BaseNode::default(),
                name: None,
                monotype: MonoType::Array(Box::new(ArrayType {
                    base: BaseNode::default(),
                    element: MonoType::Basic(NamedType {
                        base: BaseNode::default(),
                        name: Identifier {
                            base: BaseNode::default(),
                            name: "int".to_string(),
                        },
                    }),
                })),
            }],
            monotype: MonoType::Basic(NamedType {
                base: BaseNode::default(),
                name: Identifier {
                    base: BaseNode::default(),
                    name: "int".to_string(),
                },
            }),
        };

        let mut builder = flatbuffers::FlatBufferBuilder::new();

        assert_eq!(
            want,
            flatbuffers::get_root::<fb::FunctionType>(serialize(
                &mut builder,
                want.clone(),
                build_function_type,
            ))
            .into()
        );
    }
    #[test]
    fn test_type_expression() {
        let want = TypeExpression {
            base: BaseNode::default(),
            constraints: vec![TypeConstraint {
                base: BaseNode::default(),
                tvar: Identifier {
                    base: BaseNode::default(),
                    name: "T".to_string(),
                },
                kinds: vec![
                    Identifier {
                        base: BaseNode::default(),
                        name: "Addable".to_string(),
                    },
                    Identifier {
                        base: BaseNode::default(),
                        name: "Divisible".to_string(),
                    },
                ],
            }],
            monotype: MonoType::Function(Box::new(FunctionType {
                base: BaseNode::default(),
                parameters: vec![
                    ParameterType::Required {
                        base: BaseNode::default(),
                        name: Identifier {
                            base: BaseNode::default(),
                            name: "a".to_string(),
                        },
                        monotype: MonoType::Tvar(TvarType {
                            base: BaseNode::default(),
                            name: Identifier {
                                base: BaseNode::default(),
                                name: "T".to_string(),
                            },
                        }),
                    },
                    ParameterType::Required {
                        base: BaseNode::default(),
                        name: Identifier {
                            base: BaseNode::default(),
                            name: "b".to_string(),
                        },
                        monotype: MonoType::Tvar(TvarType {
                            base: BaseNode::default(),
                            name: Identifier {
                                base: BaseNode::default(),
                                name: "T".to_string(),
                            },
                        }),
                    },
                ],
                monotype: MonoType::Tvar(TvarType {
                    base: BaseNode::default(),
                    name: Identifier {
                        base: BaseNode::default(),
                        name: "T".to_string(),
                    },
                }),
            })),
        };

        let mut builder = flatbuffers::FlatBufferBuilder::new();

        assert_eq!(
            want,
            flatbuffers::get_root::<fb::TypeExpression>(serialize(
                &mut builder,
                want.clone(),
                build_type_expression,
            ))
            .into()
        );
    }
}
