//! FlatBuffers serialization for the semantic graph.

#[allow(clippy::all, missing_docs, clippy::undocumented_unsafe_blocks)]
pub mod semantic_generated;
#[allow(missing_docs)]
pub mod types;

use std::{cell::RefCell, rc::Rc};

use anyhow::{anyhow, bail, Error, Result};
use flatbuffers::{UnionWIPOffset, WIPOffset};
use semantic_generated::fbsemantic;

use crate::{ast, semantic, semantic::walk};

extern crate chrono;
use chrono::Duration as ChronoDuration;

const UNKNOWNVARIANTNAME: &str = "UNKNOWNSEMANTIC";

/// Serializes a [`semantic::nodes::Package`].
pub fn serialize_pkg(semantic_pkg: &semantic::nodes::Package) -> Result<(Vec<u8>, usize)> {
    let mut builder = flatbuffers::FlatBufferBuilder::with_capacity(1024);
    let offset = serialize(semantic_pkg, &mut builder)?;
    builder.finish(offset, None);
    // Collapse releases ownership of the byte vector and returns it to caller.
    Ok(builder.collapse())
}

/// Serializes a [`semantic::nodes::Package`] into an existing builder.
pub fn serialize_pkg_into<'a>(
    semantic_pkg: &semantic::nodes::Package,
    builder: &mut flatbuffers::FlatBufferBuilder<'a>,
) -> Result<flatbuffers::WIPOffset<fbsemantic::Package<'a>>> {
    serialize(semantic_pkg, builder)
}

/// Serializes a [`semantic::nodes::Package`] into an existing builder.
fn serialize<'a, 'b>(
    semantic_pkg: &semantic::nodes::Package,
    builder: &'b mut flatbuffers::FlatBufferBuilder<'a>,
) -> Result<flatbuffers::WIPOffset<fbsemantic::Package<'a>>> {
    let mut v = SerializingVisitor {
        inner: SerializingVisitorState::with_builder(builder),
    };
    walk::walk(&mut v, walk::Node::Package(semantic_pkg));
    v.offset()
}

struct SerializingVisitor<'a, 'b> {
    inner: SerializingVisitorState<'a, 'b>,
}

impl<'a, 'b> semantic::walk::Visitor<'_> for SerializingVisitor<'a, 'b> {
    fn visit(&mut self, _node: walk::Node<'_>) -> bool {
        if self.inner.err.is_some() {
            return false;
        }
        true
    }

    fn done(&mut self, node: walk::Node<'_>) {
        let v = &mut self.inner;
        if v.err.is_some() {
            return;
        }
        let loc = v.create_loc(node.loc());
        match &node {
            walk::Node::IntegerLit(int) => {
                let int = fbsemantic::IntegerLiteral::create(
                    v.builder,
                    &fbsemantic::IntegerLiteralArgs {
                        loc,
                        value: int.value,
                    },
                );
                v.expr_stack
                    .push((int.as_union_value(), fbsemantic::Expression::IntegerLiteral))
            }
            walk::Node::UintLit(uint) => {
                let uint = fbsemantic::UnsignedIntegerLiteral::create(
                    v.builder,
                    &fbsemantic::UnsignedIntegerLiteralArgs {
                        loc,
                        value: uint.value,
                    },
                );
                v.expr_stack.push((
                    uint.as_union_value(),
                    fbsemantic::Expression::UnsignedIntegerLiteral,
                ))
            }
            walk::Node::FloatLit(float) => {
                let float = fbsemantic::FloatLiteral::create(
                    v.builder,
                    &fbsemantic::FloatLiteralArgs {
                        loc,
                        value: float.value,
                    },
                );
                v.expr_stack
                    .push((float.as_union_value(), fbsemantic::Expression::FloatLiteral))
            }
            walk::Node::RegexpLit(regex) => {
                let regex_val = v.create_string(&regex.value);
                let regex = fbsemantic::RegexpLiteral::create(
                    v.builder,
                    &fbsemantic::RegexpLiteralArgs {
                        loc,
                        value: regex_val,
                    },
                );
                v.expr_stack.push((
                    regex.as_union_value(),
                    fbsemantic::Expression::RegexpLiteral,
                ))
            }
            walk::Node::StringLit(string) => {
                let string_val = v.create_string(&string.value);
                let string = fbsemantic::StringLiteral::create(
                    v.builder,
                    &fbsemantic::StringLiteralArgs {
                        loc,
                        value: string_val,
                    },
                );
                v.expr_stack.push((
                    string.as_union_value(),
                    fbsemantic::Expression::StringLiteral,
                ))
            }
            walk::Node::DurationLit(dur_lit) => {
                let mut dur_vec: Vec<WIPOffset<fbsemantic::Duration>> = Vec::new();
                let nanoseconds = dur_lit.value.nanoseconds;
                let months = dur_lit.value.months;
                let negative = dur_lit.value.negative;

                let dur = fbsemantic::Duration::create(
                    v.builder,
                    &fbsemantic::DurationArgs {
                        months,
                        nanoseconds,
                        negative,
                    },
                );

                dur_vec.push(dur);
                let value = Some(v.builder.create_vector(dur_vec.as_slice()));
                let dur_lit = fbsemantic::DurationLiteral::create(
                    v.builder,
                    &fbsemantic::DurationLiteralArgs { loc, value },
                );
                v.expr_stack.push((
                    dur_lit.as_union_value(),
                    fbsemantic::Expression::DurationLiteral,
                ))
            }

            walk::Node::DateTimeLit(datetime) => {
                let val = datetime.value.to_rfc3339();
                let val = v.create_string(&val);

                let secs = datetime.value.timestamp();
                let nano_secs = datetime.value.timestamp_subsec_nanos();
                let offset = datetime.value.offset().local_minus_utc();

                let time = fbsemantic::Time::create(
                    v.builder,
                    &fbsemantic::TimeArgs {
                        secs,
                        nsecs: nano_secs,
                        offset,
                    },
                );

                let datetime = fbsemantic::DateTimeLiteral::create(
                    v.builder,
                    &fbsemantic::DateTimeLiteralArgs {
                        loc,
                        value: Some(time),
                    },
                );
                v.expr_stack.push((
                    datetime.as_union_value(),
                    fbsemantic::Expression::DateTimeLiteral,
                ))
            }

            walk::Node::BooleanLit(boolean) => {
                let boolean = fbsemantic::BooleanLiteral::create(
                    v.builder,
                    &fbsemantic::BooleanLiteralArgs {
                        loc,
                        value: boolean.value,
                    },
                );
                v.expr_stack.push((
                    boolean.as_union_value(),
                    fbsemantic::Expression::BooleanLiteral,
                ))
            }

            walk::Node::IdentifierExpr(id) => {
                let name = v.create_symbol(&id.name);
                let id_typ = id.typ.clone();
                let (typ, typ_type) = types::build_type(v.builder, &id_typ);

                let ident = fbsemantic::IdentifierExpression::create(
                    v.builder,
                    &fbsemantic::IdentifierExpressionArgs {
                        loc,
                        name,
                        typ: Some(typ),
                        typ_type,
                    },
                );
                v.expr_stack.push((
                    ident.as_union_value(),
                    fbsemantic::Expression::IdentifierExpression,
                ))
            }

            walk::Node::Identifier(id) => {
                let name = v.create_symbol(&id.name);
                let identifier = fbsemantic::Identifier::create(
                    v.builder,
                    &fbsemantic::IdentifierArgs { loc, name },
                );
                v.identifiers.push(identifier)
            }

            walk::Node::Property(prop) => {
                // the value for a property is always an expression
                let key = v.pop_ident();
                let (value, value_type) = v.pop_expr();

                let prop = fbsemantic::Property::create(
                    v.builder,
                    &fbsemantic::PropertyArgs {
                        loc,
                        key,
                        value_type,
                        value,
                    },
                );
                v.properties.push(prop);
            }

            walk::Node::UnaryExpr(unary) => {
                let operator = fb_operator(&unary.operator);
                let (argument, argument_type) = v.pop_expr();

                let unary_typ = unary.typ.clone();
                let (typ, typ_type) = types::build_type(v.builder, &unary_typ);
                let unary = fbsemantic::UnaryExpression::create(
                    v.builder,
                    &fbsemantic::UnaryExpressionArgs {
                        loc,
                        operator,
                        argument,
                        argument_type,
                        typ: Some(typ),
                        typ_type,
                    },
                );
                v.expr_stack.push((
                    unary.as_union_value(),
                    fbsemantic::Expression::UnaryExpression,
                ));
            }

            walk::Node::ObjectExpr(obj) => {
                let with = match obj.with {
                    None => None,
                    Some(_) => v.pop_expr_with_kind(fbsemantic::Expression::IdentifierExpression),
                };

                let properties = {
                    let prop_vec = v.create_property_vector(obj.properties.len());
                    let fb_prop_vec = v.builder.create_vector(prop_vec.as_slice());
                    Some(fb_prop_vec)
                };

                let obj_type = obj.typ.clone();
                let (typ, typ_type) = types::build_type(v.builder, &obj_type);

                let obj = fbsemantic::ObjectExpression::create(
                    v.builder,
                    &fbsemantic::ObjectExpressionArgs {
                        loc,
                        with,
                        properties,
                        typ: Some(typ),
                        typ_type,
                    },
                );
                v.expr_stack.push((
                    obj.as_union_value(),
                    fbsemantic::Expression::ObjectExpression,
                ));
            }

            walk::Node::IndexExpr(ind) => {
                let (index, index_type) = v.pop_expr();
                let (array, array_type) = v.pop_expr();
                let ind_type = ind.typ.clone();
                let (typ, typ_type) = types::build_type(v.builder, &ind_type);

                let index = fbsemantic::IndexExpression::create(
                    v.builder,
                    &fbsemantic::IndexExpressionArgs {
                        loc,
                        array,
                        array_type,
                        index,
                        index_type,
                        typ: Some(typ),
                        typ_type,
                    },
                );
                v.expr_stack.push((
                    index.as_union_value(),
                    fbsemantic::Expression::IndexExpression,
                ));
            }

            walk::Node::MemberExpr(member) => {
                let property = v.create_symbol(&member.property);
                let (object, object_type) = v.pop_expr();

                let member_typ = member.typ.clone();
                let (typ, typ_type) = types::build_type(v.builder, &member_typ);

                let mem = fbsemantic::MemberExpression::create(
                    v.builder,
                    &fbsemantic::MemberExpressionArgs {
                        loc,
                        object,
                        object_type,
                        property,
                        typ: Some(typ),
                        typ_type,
                    },
                );
                v.expr_stack.push((
                    mem.as_union_value(),
                    fbsemantic::Expression::MemberExpression,
                ));
            }

            walk::Node::LogicalExpr(logical) => {
                let operator = fb_logical_operator(&logical.operator);
                let (right, right_type) = v.pop_expr();
                let (left, left_type) = v.pop_expr();

                let logical = fbsemantic::LogicalExpression::create(
                    v.builder,
                    &fbsemantic::LogicalExpressionArgs {
                        loc,
                        operator,
                        left_type,
                        left,
                        right_type,
                        right,
                    },
                );
                v.expr_stack.push((
                    logical.as_union_value(),
                    fbsemantic::Expression::LogicalExpression,
                ));
            }

            walk::Node::ConditionalExpr(cond) => {
                let (alternate, alternate_type) = v.pop_expr();
                let (consequent, consequent_type) = v.pop_expr();
                let (test, test_type) = v.pop_expr();

                let cond = fbsemantic::ConditionalExpression::create(
                    v.builder,
                    &fbsemantic::ConditionalExpressionArgs {
                        loc,
                        test_type,
                        test,
                        alternate_type,
                        alternate,
                        consequent_type,
                        consequent,
                    },
                );
                v.expr_stack.push((
                    cond.as_union_value(),
                    fbsemantic::Expression::ConditionalExpression,
                ));
            }

            walk::Node::CallExpr(call) => {
                let (pipe, pipe_type) = {
                    match &call.pipe {
                        Some(_) => v.pop_expr(),
                        _ => (None, fbsemantic::Expression::NONE),
                    }
                };

                let (callee, callee_type) = v.pop_expr();

                let arguments = {
                    let arg_num = call.arguments.len();
                    let start = v.properties.len() - arg_num;
                    let arg_slice = &v.properties.as_slice()[start..];
                    let vec = v.builder.create_vector(arg_slice);
                    v.properties.truncate(start);
                    Some(vec)
                };

                let call_typ = call.typ.clone();
                let (typ, typ_type) = types::build_type(v.builder, &call_typ);

                let call = fbsemantic::CallExpression::create(
                    v.builder,
                    &fbsemantic::CallExpressionArgs {
                        loc,
                        callee,
                        callee_type,
                        arguments,
                        pipe,
                        pipe_type,
                        typ: Some(typ),
                        typ_type,
                    },
                );
                v.expr_stack.push((
                    call.as_union_value(),
                    fbsemantic::Expression::CallExpression,
                ));
            }

            walk::Node::BinaryExpr(bin) => {
                let operator = fb_operator(&bin.operator);
                let (right, right_type) = v.pop_expr();
                let (left, left_type) = v.pop_expr();

                let bin_typ = bin.typ.clone();
                let (typ, typ_type) = types::build_type(v.builder, &bin_typ);

                let bin = fbsemantic::BinaryExpression::create(
                    v.builder,
                    &fbsemantic::BinaryExpressionArgs {
                        loc,
                        operator,
                        left_type,
                        left,
                        right_type,
                        right,
                        typ: Some(typ),
                        typ_type,
                    },
                );
                v.expr_stack.push((
                    bin.as_union_value(),
                    fbsemantic::Expression::BinaryExpression,
                ));
            }

            walk::Node::FunctionExpr(func) => {
                let params = {
                    let par_num = func.params.len();
                    let start = v.params.len() - par_num;
                    let par_slice = &v.params.as_slice()[start..];
                    let vec = v.builder.create_vector(par_slice);
                    v.params.truncate(start);
                    Some(vec)
                };

                let mut block_len = 0;
                let mut current = &func.body;
                let body_start_pos = &current.loc().start;
                let mut body_end_pos = &current.loc().end;
                loop {
                    block_len += 1;
                    body_end_pos = &current.loc().end;
                    match current {
                        semantic::nodes::Block::Expr(_, next) => {
                            current = next.as_ref();
                        }
                        semantic::nodes::Block::Variable(_, next) => {
                            current = next.as_ref();
                        }
                        semantic::nodes::Block::Return(retn) => {
                            break;
                        }
                    }
                }
                let body_loc = ast::SourceLocation {
                    file: func.loc.file.clone(),
                    start: *body_start_pos,
                    end: *body_end_pos,
                    source: None,
                };
                let body_loc = v.create_loc(&body_loc);
                let body_vec = {
                    let stmt_vec = v.create_stmt_vector(block_len);
                    Some(v.builder.create_vector(stmt_vec.as_slice()))
                };
                let body = Some(fbsemantic::Block::create(
                    v.builder,
                    &fbsemantic::BlockArgs {
                        loc: body_loc,
                        body: body_vec,
                    },
                ));

                let func_typ = func.typ.clone();
                let (typ, typ_type) = types::build_type(v.builder, &func_typ);
                let vectorized = if func.vectorized.is_some() {
                    v.pop_expr_with_kind(fbsemantic::Expression::FunctionExpression)
                } else {
                    None
                };

                let func = fbsemantic::FunctionExpression::create(
                    v.builder,
                    &fbsemantic::FunctionExpressionArgs {
                        loc,
                        params,
                        body,
                        typ: Some(typ),
                        typ_type,
                        vectorized,
                    },
                );
                v.expr_stack.push((
                    func.as_union_value(),
                    fbsemantic::Expression::FunctionExpression,
                ));
            }

            walk::Node::FunctionParameter(func_param) => {
                let key = v.pop_ident();

                let (default, default_type) = {
                    match &func_param.default {
                        Some(_) => v.pop_expr(),
                        _ => (None, fbsemantic::Expression::NONE),
                    }
                };

                let func_param = fbsemantic::FunctionParameter::create(
                    v.builder,
                    &fbsemantic::FunctionParameterArgs {
                        loc,
                        key,
                        is_pipe: func_param.is_pipe,
                        default,
                        default_type,
                    },
                );
                v.params.push(func_param);
            }

            walk::Node::ArrayExpr(array) => {
                let num_elems = array.elements.len();
                let start = v.expr_stack.len() - num_elems;
                let elements = {
                    let elems = &v.expr_stack.as_slice()[start..];
                    let mut wrapped_elems = Vec::with_capacity(num_elems);
                    for (e, et) in elems {
                        wrapped_elems.push(fbsemantic::WrappedExpression::create(
                            v.builder,
                            &fbsemantic::WrappedExpressionArgs {
                                expression_type: *et,
                                expression: Some(*e),
                            },
                        ));
                    }
                    Some(v.builder.create_vector(wrapped_elems.as_slice()))
                };
                v.expr_stack.truncate(start);
                let arr_typ = array.typ.clone();
                let (typ, typ_type) = types::build_type(v.builder, &arr_typ);

                let array = fbsemantic::ArrayExpression::create(
                    v.builder,
                    &fbsemantic::ArrayExpressionArgs {
                        loc,
                        elements,
                        typ: Some(typ),
                        typ_type,
                    },
                );
                v.expr_stack.push((
                    array.as_union_value(),
                    fbsemantic::Expression::ArrayExpression,
                ))
            }

            walk::Node::DictExpr(dict) => {
                let num_elems = dict.elements.len();
                let stop = v.expr_stack.len();
                let start = stop - 2 * num_elems;
                let elements = {
                    let elems = &v.expr_stack.as_slice();
                    let mut items = Vec::with_capacity(num_elems);
                    for i in (start..stop).step_by(2) {
                        let (key, key_type) = elems[i];
                        let (val, val_type) = elems[i + 1];
                        items.push(fbsemantic::DictItem::create(
                            v.builder,
                            &fbsemantic::DictItemArgs {
                                key_type,
                                val_type,
                                key: Some(key),
                                val: Some(val),
                            },
                        ));
                    }
                    Some(v.builder.create_vector(items.as_slice()))
                };
                v.expr_stack.truncate(start);
                let (typ, typ_type) = types::build_type(v.builder, &dict.typ.clone());

                let dict = fbsemantic::DictExpression::create(
                    v.builder,
                    &fbsemantic::DictExpressionArgs {
                        loc,
                        elements,
                        typ: Some(typ),
                        typ_type,
                    },
                );
                v.expr_stack.push((
                    dict.as_union_value(),
                    fbsemantic::Expression::DictExpression,
                ))
            }

            walk::Node::TextPart(tp) => {
                let text_value = Some(v.builder.create_string(tp.value.as_str()));
                let text = fbsemantic::StringExpressionPart::create(
                    v.builder,
                    &fbsemantic::StringExpressionPartArgs {
                        loc,
                        text_value,
                        ..fbsemantic::StringExpressionPartArgs::default()
                    },
                );
                v.string_expr_parts.push(text);
            }

            walk::Node::InterpolatedPart(_) => {
                let (interpolated_expression, interpolated_expression_type) = v.pop_expr();
                let inter = fbsemantic::StringExpressionPart::create(
                    v.builder,
                    &fbsemantic::StringExpressionPartArgs {
                        loc,
                        interpolated_expression_type,
                        interpolated_expression,
                        ..fbsemantic::StringExpressionPartArgs::default()
                    },
                );
                v.string_expr_parts.push(inter);
            }

            walk::Node::StringExpr(string) => {
                let parts = {
                    let num_parts = string.parts.len();
                    let start = v.string_expr_parts.len() - num_parts;
                    let parts_sl = &v.string_expr_parts.as_slice()[start..];
                    let vec = v.builder.create_vector(parts_sl);
                    v.string_expr_parts.truncate(start);
                    Some(vec)
                };
                let string = fbsemantic::StringExpression::create(
                    v.builder,
                    &fbsemantic::StringExpressionArgs { loc, parts },
                );
                v.expr_stack.push((
                    string.as_union_value(),
                    fbsemantic::Expression::StringExpression,
                ));
            }

            walk::Node::MemberAssgn(mem) => {
                let (init_, init__type) = v.pop_expr();
                let member = v.pop_expr_with_kind(fbsemantic::Expression::MemberExpression);
                let mem = fbsemantic::MemberAssignment::create(
                    v.builder,
                    &fbsemantic::MemberAssignmentArgs {
                        loc,
                        member,
                        init__type,
                        init_,
                    },
                );
                v.stmts.push((
                    mem.as_union_value(),
                    fbsemantic::Statement::MemberAssignment,
                ));
            }

            walk::Node::VariableAssgn(native) => {
                let (init_, init__type) = v.pop_expr();
                let identifier = v.pop_ident();

                let poly = native.poly_type_of();
                let typ = Some(types::build_polytype(v.builder, poly));

                let native = fbsemantic::NativeVariableAssignment::create(
                    v.builder,
                    &fbsemantic::NativeVariableAssignmentArgs {
                        loc,
                        identifier,
                        init__type,
                        init_,
                        typ,
                    },
                );
                v.stmts.push((
                    native.as_union_value(),
                    fbsemantic::Statement::NativeVariableAssignment,
                ));
            }

            walk::Node::ReturnStmt(_) => {
                let (argument, argument_type) = v.pop_expr();
                let return_st = fbsemantic::ReturnStatement::create(
                    v.builder,
                    &fbsemantic::ReturnStatementArgs {
                        loc,
                        argument_type,
                        argument,
                    },
                );
                v.stmts.push((
                    return_st.as_union_value(),
                    fbsemantic::Statement::ReturnStatement,
                ));
            }

            walk::Node::ExprStmt(_) => {
                let (expression, expression_type) = v.pop_expr();
                let expr = fbsemantic::ExpressionStatement::create(
                    v.builder,
                    &fbsemantic::ExpressionStatementArgs {
                        loc,
                        expression_type,
                        expression,
                    },
                );
                v.stmts.push((
                    expr.as_union_value(),
                    fbsemantic::Statement::ExpressionStatement,
                ));
            }

            walk::Node::TestStmt(test) => {
                let assignment = {
                    match v.stmts.pop() {
                        Some((union, fbsemantic::Statement::NativeVariableAssignment)) => {
                            Some(WIPOffset::new(union.value()))
                        }
                        _ => {
                            v.err = Some(anyhow!(
                                "failed to pop assignment statement from stmt vector",
                            ));
                            return;
                        }
                    }
                };

                let test = fbsemantic::TestStatement::create(
                    v.builder,
                    &fbsemantic::TestStatementArgs { loc, assignment },
                );
                v.stmts
                    .push((test.as_union_value(), fbsemantic::Statement::TestStatement));
            }

            walk::Node::TestCaseStmt(test) => {
                // TestCase statements should be transformed before the semantic phase. Even without
                // the explicit panic here, an error would occur because the block has no function
                // scope.
                panic!("TestCaseStmt is not supported in semantic analysis.");
            }

            walk::Node::BuiltinStmt(builtin) => {
                let id = v.pop_ident();
                let builtin = fbsemantic::BuiltinStatement::create(
                    v.builder,
                    &fbsemantic::BuiltinStatementArgs { loc, id },
                );
                v.stmts.push((
                    builtin.as_union_value(),
                    fbsemantic::Statement::BuiltinStatement,
                ));
            }

            walk::Node::OptionStmt(opt) => {
                let (assignment, assignment_type) = {
                    match &opt.assignment {
                        semantic::nodes::Assignment::Variable(_) => match v.stmts.pop() {
                            Some((nva, fbsemantic::Statement::NativeVariableAssignment)) => {
                                (Some(nva), fbsemantic::Assignment::NativeVariableAssignment)
                            }
                            Some((_, ty)) => {
                                v.err = Some(anyhow!("found {:?} in stmt vector", ty));
                                return;
                            }
                            None => {
                                v.err = Some(anyhow!(
                                    "Native assignment was not added to SerializingVisitor",
                                ));
                                return;
                            }
                        },
                        semantic::nodes::Assignment::Member(_) => match v.stmts.pop() {
                            Some((member, fbsemantic::Statement::MemberAssignment)) => {
                                (Some(member), fbsemantic::Assignment::MemberAssignment)
                            }
                            _ => {
                                v.err = Some(anyhow!(
                                    "Member assignment was not added to SerializingVisitor",
                                ));
                                return;
                            }
                        },
                    }
                };
                let opt = fbsemantic::OptionStatement::create(
                    v.builder,
                    &fbsemantic::OptionStatementArgs {
                        loc,
                        assignment_type,
                        assignment,
                    },
                );
                v.stmts
                    .push((opt.as_union_value(), fbsemantic::Statement::OptionStatement));
            }

            walk::Node::Block(block) => {
                // Block statements must be convered in this enum in order to also walk all child nodes.
                // However, block statements are consumed by the function expression node so that
                // blocks are only constructed once all child nodes have been traversed.
            }

            walk::Node::ImportDeclaration(imp) => {
                let alias = {
                    match imp.alias {
                        Some(_) => v.pop_ident(),
                        _ => None,
                    }
                };

                let path = v.pop_expr_with_kind(fbsemantic::Expression::StringLiteral);
                let import = fbsemantic::ImportDeclaration::create(
                    v.builder,
                    &fbsemantic::ImportDeclarationArgs { loc, alias, path },
                );
                v.import_decls.push(import);
            }

            walk::Node::PackageClause(_) => {
                let name = v.pop_ident();
                let pc = fbsemantic::PackageClause::create(
                    v.builder,
                    &fbsemantic::PackageClauseArgs { loc, name },
                );
                v.package_clause = Some(pc);
            }

            walk::Node::File(file) => {
                let package = v.package_clause;
                v.package_clause = None;

                let imports = Some(v.builder.create_vector(v.import_decls.as_slice()));
                v.import_decls.clear();

                let stmt_vec = v.create_stmt_vector(file.body.len());
                let body = Some(v.builder.create_vector(stmt_vec.as_slice()));

                let file = fbsemantic::File::create(
                    v.builder,
                    &fbsemantic::FileArgs {
                        loc,
                        package,
                        imports,
                        body,
                    },
                );
                v.files.push(file);
            }

            walk::Node::Package(pac) => {
                let package = v.create_string(&pac.package);
                let files = {
                    let mut fs: Vec<WIPOffset<fbsemantic::File>> = Vec::new();
                    std::mem::swap(&mut v.files, &mut fs);
                    Some(v.builder.create_vector(fs.as_slice()))
                };
                v.package = Some(fbsemantic::Package::create(
                    v.builder,
                    &fbsemantic::PackageArgs {
                        loc,
                        package,
                        files,
                    },
                ));
            }

            walk::Node::ErrorStmt(_) | walk::Node::ErrorExpr(_) => {
                unreachable!("We should never try to serialize error nodes")
            }
        }
    }
}

impl<'a, 'b> SerializingVisitor<'a, 'b> {
    // Return the offset for the package, checking for any error that may have occurred during
    // serialization.
    fn offset(self) -> Result<flatbuffers::WIPOffset<fbsemantic::Package<'a>>> {
        let v = self.inner;
        if let Some(e) = v.err {
            return Err(e);
        };
        match v.package {
            None => Err(anyhow!("missing serialized package")),
            Some(offset) => Ok(offset),
        }
    }
}

struct SerializingVisitorState<'a: 'b, 'b> {
    // Any error that occurred during serialization, returned by the visitor's check method.
    err: Option<Error>,

    builder: &'b mut flatbuffers::FlatBufferBuilder<'a>,

    package: Option<WIPOffset<fbsemantic::Package<'a>>>,
    package_clause: Option<WIPOffset<fbsemantic::PackageClause<'a>>>,

    import_decls: Vec<WIPOffset<fbsemantic::ImportDeclaration<'a>>>,
    files: Vec<WIPOffset<fbsemantic::File<'a>>>,
    blocks: Vec<WIPOffset<fbsemantic::Block<'a>>>,
    stmts: Vec<(WIPOffset<UnionWIPOffset>, fbsemantic::Statement)>,
    vars: Vec<(WIPOffset<UnionWIPOffset>, fbsemantic::Var<'a>)>,
    params: Vec<WIPOffset<fbsemantic::FunctionParameter<'a>>>,

    expr_stack: Vec<(WIPOffset<UnionWIPOffset>, fbsemantic::Expression)>,
    properties: Vec<WIPOffset<fbsemantic::Property<'a>>>,
    identifiers: Vec<WIPOffset<fbsemantic::Identifier<'a>>>,
    string_expr_parts: Vec<WIPOffset<fbsemantic::StringExpressionPart<'a>>>,
}

impl<'a, 'b> SerializingVisitorState<'a, 'b> {
    fn with_builder(
        builder: &'b mut flatbuffers::FlatBufferBuilder<'a>,
    ) -> SerializingVisitorState<'a, 'b> {
        SerializingVisitorState {
            err: None,
            builder,
            package: None,
            package_clause: None,
            import_decls: Vec::new(),
            files: Vec::new(),
            blocks: Vec::new(),
            stmts: Vec::new(),
            vars: Vec::new(),
            params: Vec::new(),
            expr_stack: Vec::new(),
            properties: Vec::new(),
            identifiers: Vec::new(),
            string_expr_parts: Vec::new(),
        }
    }

    fn pop_expr(&mut self) -> (Option<WIPOffset<UnionWIPOffset>>, fbsemantic::Expression) {
        match self.expr_stack.pop() {
            None => {
                self.err = Some(anyhow!("Tried popping empty expression stack"));
                (None, fbsemantic::Expression::NONE)
            }
            Some((o, e)) => (Some(o), e),
        }
    }

    fn pop_expr_with_kind<T>(&mut self, kind: fbsemantic::Expression) -> Option<WIPOffset<T>> {
        match self.expr_stack.pop() {
            Some((wipo, e)) => {
                if e == kind {
                    Some(WIPOffset::new(wipo.value()))
                } else {
                    self.err = Some(anyhow!(
                        "expected {} on expr stack, got {}",
                        kind.variant_name().unwrap_or(UNKNOWNVARIANTNAME),
                        e.variant_name().unwrap_or(UNKNOWNVARIANTNAME)
                    ));
                    None
                }
            }
            None => {
                self.err = Some(anyhow!("Tried popping empty expression stack"));
                None
            }
        }
    }

    fn pop_ident<T>(&mut self) -> Option<WIPOffset<T>> {
        match self.identifiers.pop() {
            None => {
                self.err = Some(anyhow!("Tried popping empty identifier stack"));
                None
            }
            Some(wip) => Some(WIPOffset::new(wip.value())),
        }
    }

    fn create_symbol(&mut self, symbol: &semantic::nodes::Symbol) -> Option<WIPOffset<&'a str>> {
        Some(self.builder.create_shared_string(symbol.full_name()))
    }

    fn create_string(&mut self, string: &str) -> Option<WIPOffset<&'a str>> {
        Some(self.builder.create_string(string))
    }

    fn create_opt_string(&mut self, str: &Option<String>) -> Option<WIPOffset<&'a str>> {
        match str {
            None => None,
            Some(str) => Some(self.builder.create_string(str.as_str())),
        }
    }

    fn create_stmt_vector(
        &mut self,
        num_of_stmts: usize,
    ) -> Vec<WIPOffset<fbsemantic::WrappedStatement<'a>>> {
        let start = self.stmts.len() - num_of_stmts;
        let union_stmts = &self.stmts.as_slice()[start..];
        let mut wrapped_stmts: Vec<WIPOffset<fbsemantic::WrappedStatement>> =
            Vec::with_capacity(num_of_stmts);

        for (stmt, stmt_type) in union_stmts {
            let wrapped_st = fbsemantic::WrappedStatement::create(
                self.builder,
                &fbsemantic::WrappedStatementArgs {
                    statement_type: *stmt_type,
                    statement: Some(*stmt),
                },
            );
            wrapped_stmts.push(wrapped_st);
        }
        self.stmts.truncate(start);
        wrapped_stmts
    }

    fn create_property_vector(
        &mut self,
        n_props: usize,
    ) -> Vec<WIPOffset<fbsemantic::Property<'a>>> {
        let start = self.properties.len() - n_props;
        self.properties.split_off(start)
    }

    fn pop_assignment_stmt(
        &mut self,
    ) -> (Option<WIPOffset<UnionWIPOffset>>, fbsemantic::Assignment) {
        match self.stmts.pop() {
            Some((va, fbsemantic::Statement::NativeVariableAssignment)) => {
                (Some(va), fbsemantic::Assignment::NativeVariableAssignment)
            }
            None => {
                self.err = Some(anyhow!(
                    "Tried popping empty statement stack. Expected assignment on top of stack.",
                ));
                (None, fbsemantic::Assignment::NONE)
            }
            Some(_) => {
                self.err = Some(anyhow!(
                    "Expected assignment on top of stack statement stack.",
                ));
                (None, fbsemantic::Assignment::NONE)
            }
        }
    }

    fn create_loc(
        &mut self,
        loc: &ast::SourceLocation,
    ) -> Option<WIPOffset<fbsemantic::SourceLocation<'a>>> {
        let file = self.create_opt_string(&loc.file);
        let source = self.create_opt_string(&loc.source);

        Some(fbsemantic::SourceLocation::create(
            self.builder,
            &fbsemantic::SourceLocationArgs {
                file,
                start: Some(&fbsemantic::Position::new(
                    loc.start.line as i32,
                    loc.start.column as i32,
                )),
                end: Some(&fbsemantic::Position::new(
                    loc.end.line as i32,
                    loc.end.column as i32,
                )),
                source,
            },
        ))
    }
}

fn fb_operator(o: &ast::Operator) -> fbsemantic::Operator {
    match o {
        ast::Operator::MultiplicationOperator => fbsemantic::Operator::MultiplicationOperator,
        ast::Operator::DivisionOperator => fbsemantic::Operator::DivisionOperator,
        ast::Operator::ModuloOperator => fbsemantic::Operator::ModuloOperator,
        ast::Operator::PowerOperator => fbsemantic::Operator::PowerOperator,
        ast::Operator::AdditionOperator => fbsemantic::Operator::AdditionOperator,
        ast::Operator::SubtractionOperator => fbsemantic::Operator::SubtractionOperator,
        ast::Operator::LessThanEqualOperator => fbsemantic::Operator::LessThanEqualOperator,
        ast::Operator::LessThanOperator => fbsemantic::Operator::LessThanOperator,
        ast::Operator::GreaterThanEqualOperator => fbsemantic::Operator::GreaterThanEqualOperator,
        ast::Operator::GreaterThanOperator => fbsemantic::Operator::GreaterThanOperator,
        ast::Operator::StartsWithOperator => fbsemantic::Operator::StartsWithOperator,
        ast::Operator::InOperator => fbsemantic::Operator::InOperator,
        ast::Operator::NotOperator => fbsemantic::Operator::NotOperator,
        ast::Operator::ExistsOperator => fbsemantic::Operator::ExistsOperator,
        ast::Operator::NotEmptyOperator => fbsemantic::Operator::NotEmptyOperator,
        ast::Operator::EmptyOperator => fbsemantic::Operator::EmptyOperator,
        ast::Operator::EqualOperator => fbsemantic::Operator::EqualOperator,
        ast::Operator::NotEqualOperator => fbsemantic::Operator::NotEqualOperator,
        ast::Operator::RegexpMatchOperator => fbsemantic::Operator::RegexpMatchOperator,
        ast::Operator::NotRegexpMatchOperator => fbsemantic::Operator::NotRegexpMatchOperator,
        ast::Operator::InvalidOperator => fbsemantic::Operator::InvalidOperator,
    }
}

fn fb_logical_operator(lo: &ast::LogicalOperator) -> fbsemantic::LogicalOperator {
    match lo {
        ast::LogicalOperator::AndOperator => fbsemantic::LogicalOperator::AndOperator,
        ast::LogicalOperator::OrOperator => fbsemantic::LogicalOperator::OrOperator,
    }
}

#[cfg(test)]
mod tests;
