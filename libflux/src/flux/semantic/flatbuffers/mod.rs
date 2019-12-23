#[allow(clippy::all)]
pub mod semantic_generated;
pub mod types;

use std::cell::RefCell;
use std::rc::Rc;

use crate::ast;

use crate::semantic;
use crate::semantic::walk;
use flatbuffers::{UnionWIPOffset, WIPOffset};
use semantic_generated::fbsemantic;

extern crate chrono;
use chrono::Offset;

pub fn serialize(semantic_pkg: &mut semantic::nodes::Package) -> Result<(Vec<u8>, usize), String> {
    let mut v = new_serializing_visitor_with_capacity(1024);
    walk::walk(&mut v, Rc::new(walk::Node::Package(semantic_pkg)));
    v.finish()
}

fn new_serializing_visitor_with_capacity<'a>(_capacity: usize) -> SerializingVisitor<'a> {
    SerializingVisitor {
        inner: Rc::new(RefCell::new(SerializingVisitorState::new_with_capacity(
            _capacity,
        ))),
    }
}

struct SerializingVisitor<'a> {
    inner: Rc<RefCell<SerializingVisitorState<'a>>>,
}

impl<'a> semantic::walk::Visitor<'_> for SerializingVisitor<'a> {
    fn visit(&mut self, _node: Rc<walk::Node<'_>>) -> bool {
        let v = self.inner.borrow();
        if v.err.is_some() {
            return false;
        }
        Rc::clone(&self.inner);
        true
    }

    fn done(&mut self, node: Rc<walk::Node<'_>>) {
        let mut v = &mut *self.inner.borrow_mut();
        if v.err.is_some() {
            return;
        }
        let node = &*node;
        let loc = v.create_loc(node.loc());
        match node {
            walk::Node::IntegerLit(int) => {
                let int_typ = int.typ.clone();
                let (typ, typ_type) = types::build_type(&mut v.builder, int_typ);

                let int = fbsemantic::IntegerLiteral::create(
                    &mut v.builder,
                    &fbsemantic::IntegerLiteralArgs {
                        loc,
                        value: int.value,
                        typ: Some(typ),
                        typ_type,
                    },
                );
                v.expr_stack
                    .push((int.as_union_value(), fbsemantic::Expression::IntegerLiteral))
            }
            walk::Node::UintLit(uint) => {
                let uint_typ = uint.typ.clone();
                let (typ, typ_type) = types::build_type(&mut v.builder, uint_typ);

                let uint = fbsemantic::UnsignedIntegerLiteral::create(
                    &mut v.builder,
                    &fbsemantic::UnsignedIntegerLiteralArgs {
                        loc,
                        value: uint.value,
                        typ: Some(typ),
                        typ_type,
                    },
                );
                v.expr_stack.push((
                    uint.as_union_value(),
                    fbsemantic::Expression::UnsignedIntegerLiteral,
                ))
            }
            walk::Node::FloatLit(float) => {
                let float_typ = float.typ.clone();
                let (typ, typ_type) = types::build_type(&mut v.builder, float_typ);

                let float = fbsemantic::FloatLiteral::create(
                    &mut v.builder,
                    &fbsemantic::FloatLiteralArgs {
                        loc,
                        value: float.value,
                        typ: Some(typ),
                        typ_type,
                    },
                );
                v.expr_stack
                    .push((float.as_union_value(), fbsemantic::Expression::FloatLiteral))
            }
            walk::Node::RegexpLit(regex) => {
                let regex_val = v.create_string(&regex.value);
                let regex_typ = regex.typ.clone();
                let (typ, typ_type) = types::build_type(&mut v.builder, regex_typ);

                let regex = fbsemantic::RegexpLiteral::create(
                    &mut v.builder,
                    &fbsemantic::RegexpLiteralArgs {
                        loc,
                        value: regex_val,
                        typ: Some(typ),
                        typ_type,
                    },
                );
                v.expr_stack.push((
                    regex.as_union_value(),
                    fbsemantic::Expression::RegexpLiteral,
                ))
            }
            walk::Node::StringLit(string) => {
                let string_val = v.create_string(&string.value);
                let string_typ = string.typ.clone();
                let (typ, typ_type) = types::build_type(&mut v.builder, string_typ);

                let string = fbsemantic::StringLiteral::create(
                    &mut v.builder,
                    &fbsemantic::StringLiteralArgs {
                        loc,
                        value: string_val,
                        typ: Some(typ),
                        typ_type,
                    },
                );
                v.expr_stack.push((
                    string.as_union_value(),
                    fbsemantic::Expression::StringLiteral,
                ))
            }

            walk::Node::DurationLit(dur_lit) => {
                let mut dur_vec: Vec<WIPOffset<fbsemantic::Duration>> = Vec::new();
                let magnitude = match dur_lit.value.num_nanoseconds() {
                    Some(mag) => mag,
                    None => {
                        v.err = Some(String::from("Empty duration value"));
                        return;
                    }
                };
                let dur = fbsemantic::Duration::create(
                    &mut v.builder,
                    &fbsemantic::DurationArgs {
                        magnitude,
                        unit: fbsemantic::TimeUnit::ns,
                    },
                );
                dur_vec.push(dur);
                let value = Some(v.builder.create_vector(dur_vec.as_slice()));

                let dur_typ = dur_lit.typ.clone();
                let (typ, typ_type) = types::build_type(&mut v.builder, dur_typ);

                let dur_lit = fbsemantic::DurationLiteral::create(
                    &mut v.builder,
                    &fbsemantic::DurationLiteralArgs {
                        loc,
                        value,
                        typ: Some(typ),
                        typ_type,
                    },
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
                let offset = datetime.value.offset().fix().local_minus_utc();

                let time = fbsemantic::Time::create(
                    &mut v.builder,
                    &fbsemantic::TimeArgs {
                        secs,
                        nsecs: nano_secs,
                        offset,
                    },
                );

                let date_typ = datetime.typ.clone();
                let (typ, typ_type) = types::build_type(&mut v.builder, date_typ);

                let datetime = fbsemantic::DateTimeLiteral::create(
                    &mut v.builder,
                    &fbsemantic::DateTimeLiteralArgs {
                        loc,
                        value: Some(time),
                        typ: Some(typ),
                        typ_type,
                    },
                );
                v.expr_stack.push((
                    datetime.as_union_value(),
                    fbsemantic::Expression::DateTimeLiteral,
                ))
            }

            walk::Node::BooleanLit(boolean) => {
                let boolean_typ = boolean.typ.clone();
                let (typ, typ_type) = types::build_type(&mut v.builder, boolean_typ);
                let boolean = fbsemantic::BooleanLiteral::create(
                    &mut v.builder,
                    &fbsemantic::BooleanLiteralArgs {
                        loc,
                        value: boolean.value,
                        typ: Some(typ),
                        typ_type,
                    },
                );
                v.expr_stack.push((
                    boolean.as_union_value(),
                    fbsemantic::Expression::BooleanLiteral,
                ))
            }

            walk::Node::IdentifierExpr(id) => {
                let name = v.create_string(&id.name);
                let id_typ = id.typ.clone();
                let (typ, typ_type) = types::build_type(&mut v.builder, id_typ);

                let ident = fbsemantic::IdentifierExpression::create(
                    &mut v.builder,
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
                let name = v.create_string(&id.name);
                let identifier = fbsemantic::Identifier::create(
                    &mut v.builder,
                    &fbsemantic::IdentifierArgs { loc, name },
                );
                v.identifiers.push(identifier)
            }

            walk::Node::Property(prop) => {
                // the value for a property is always an expression
                let key = v.pop_ident();
                let (value, value_type) = v.pop_expr();

                let prop = fbsemantic::Property::create(
                    &mut v.builder,
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
                let (typ, typ_type) = types::build_type(&mut v.builder, unary_typ);
                let unary = fbsemantic::UnaryExpression::create(
                    &mut v.builder,
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
                    let fb_prop_vec = v.builder.create_vector(&prop_vec.as_slice());
                    Some(fb_prop_vec)
                };

                let obj_type = obj.typ.clone();
                let (typ, typ_type) = types::build_type(&mut v.builder, obj_type);

                let obj = fbsemantic::ObjectExpression::create(
                    &mut v.builder,
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
                let (typ, typ_type) = types::build_type(&mut v.builder, ind_type);

                let index = fbsemantic::IndexExpression::create(
                    &mut v.builder,
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
                let property = v.create_string(&member.property);
                let (object, object_type) = v.pop_expr();

                let member_typ = member.typ.clone();
                let (typ, typ_type) = types::build_type(&mut v.builder, member_typ);

                let mem = fbsemantic::MemberExpression::create(
                    &mut v.builder,
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

                let logical_typ = logical.typ.clone();
                let (typ, typ_type) = types::build_type(&mut v.builder, logical_typ);

                let logical = fbsemantic::LogicalExpression::create(
                    &mut v.builder,
                    &fbsemantic::LogicalExpressionArgs {
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
                    logical.as_union_value(),
                    fbsemantic::Expression::LogicalExpression,
                ));
            }

            walk::Node::ConditionalExpr(cond) => {
                let (alternate, alternate_type) = v.pop_expr();
                let (consequent, consequent_type) = v.pop_expr();
                let (test, test_type) = v.pop_expr();

                let cond_typ = cond.typ.clone();
                let (typ, typ_type) = types::build_type(&mut v.builder, cond_typ);

                let cond = fbsemantic::ConditionalExpression::create(
                    &mut v.builder,
                    &fbsemantic::ConditionalExpressionArgs {
                        loc,
                        test,
                        test_type,
                        alternate,
                        alternate_type,
                        consequent,
                        consequent_type,
                        typ: Some(typ),
                        typ_type,
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
                let (typ, typ_type) = types::build_type(&mut v.builder, call_typ);

                let call = fbsemantic::CallExpression::create(
                    &mut v.builder,
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
                let (typ, typ_type) = types::build_type(&mut v.builder, bin_typ);

                let bin = fbsemantic::BinaryExpression::create(
                    &mut v.builder,
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
                loop {
                    block_len += 1;
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
                let body_vec = {
                    let stmt_vec = v.create_stmt_vector(block_len);
                    Some(v.builder.create_vector(&stmt_vec.as_slice()))
                };
                let body = Some(fbsemantic::Block::create(
                    &mut v.builder,
                    &fbsemantic::BlockArgs {
                        loc,
                        body: body_vec,
                    },
                ));

                let func_typ = func.typ.clone();
                let (typ, typ_type) = types::build_type(&mut v.builder, func_typ);

                let func = fbsemantic::FunctionExpression::create(
                    &mut v.builder,
                    &fbsemantic::FunctionExpressionArgs {
                        loc,
                        params,
                        body,
                        typ: Some(typ),
                        typ_type,
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
                    &mut v.builder,
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
                            &mut v.builder,
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
                let (typ, typ_type) = types::build_type(&mut v.builder, arr_typ);

                let array = fbsemantic::ArrayExpression::create(
                    &mut v.builder,
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

            walk::Node::TextPart(tp) => {
                let text_value = Some(v.builder.create_string(tp.value.as_str()));
                let text = fbsemantic::StringExpressionPart::create(
                    &mut v.builder,
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
                    &mut v.builder,
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
                let string_typ = string.typ.clone();
                let (typ, typ_type) = types::build_type(&mut v.builder, string_typ);

                let string = fbsemantic::StringExpression::create(
                    &mut v.builder,
                    &fbsemantic::StringExpressionArgs {
                        loc,
                        parts,
                        typ: Some(typ),
                        typ_type,
                    },
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
                    &mut v.builder,
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
                let typ = Some(types::build_polytype(&mut v.builder, poly));

                let native = fbsemantic::NativeVariableAssignment::create(
                    &mut v.builder,
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
                    &mut v.builder,
                    &fbsemantic::ReturnStatementArgs {
                        loc,
                        argument,
                        argument_type,
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
                    &mut v.builder,
                    &fbsemantic::ExpressionStatementArgs {
                        loc,
                        expression,
                        expression_type,
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
                            v.err = Some(String::from(
                                "failed to pop assignment statement from stmt vector",
                            ));
                            return;
                        }
                    }
                };

                let test = fbsemantic::TestStatement::create(
                    &mut v.builder,
                    &fbsemantic::TestStatementArgs { loc, assignment },
                );
                v.stmts
                    .push((test.as_union_value(), fbsemantic::Statement::TestStatement));
            }

            walk::Node::BuiltinStmt(builtin) => {
                let id = v.pop_ident();
                let builtin = fbsemantic::BuiltinStatement::create(
                    &mut v.builder,
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
                                v.err = Some(format!("found {:?} in stmt vector", ty));
                                return;
                            }
                            None => {
                                v.err = Some(String::from(
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
                                v.err = Some(String::from(
                                    "Member assignment was not added to SerializingVisitor",
                                ));
                                return;
                            }
                        },
                    }
                };
                let opt = fbsemantic::OptionStatement::create(
                    &mut v.builder,
                    &fbsemantic::OptionStatementArgs {
                        loc,
                        assignment,
                        assignment_type,
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
                    &mut v.builder,
                    &fbsemantic::ImportDeclarationArgs { loc, alias, path },
                );
                v.import_decls.push(import);
            }

            walk::Node::PackageClause(_) => {
                let name = v.pop_ident();
                let pc = fbsemantic::PackageClause::create(
                    &mut v.builder,
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
                let body = Some(v.builder.create_vector(&stmt_vec.as_slice()));

                let file = fbsemantic::File::create(
                    &mut v.builder,
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
                    &mut v.builder,
                    &fbsemantic::PackageArgs {
                        loc,
                        package,
                        files,
                    },
                ));
            }
        }
    }
}

impl<'a> SerializingVisitor<'a> {
    fn finish(self) -> Result<(Vec<u8>, usize), String> {
        let v = match Rc::try_unwrap(self.inner) {
            Ok(sv) => sv,
            Err(_) => return Err(String::from("error unwrapping rc")),
        };
        let mut v = v.into_inner();
        if let Some(e) = v.err {
            return Err(e);
        };
        let pkg = match v.package {
            None => return Err(String::from("missing serialized package")),
            Some(pkg) => pkg,
        };
        v.builder.finish(pkg, None);

        // Collapse releases ownership of the byte vector and returns it to caller.
        Ok(v.builder.collapse())
    }
}

struct SerializingVisitorState<'a> {
    // Any error that occurred during serialization, returned by the visitor's finish method.
    err: Option<String>,

    builder: flatbuffers::FlatBufferBuilder<'a>,

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

impl<'a> SerializingVisitorState<'a> {
    fn new_with_capacity(capacity: usize) -> SerializingVisitorState<'a> {
        SerializingVisitorState {
            err: None,
            builder: flatbuffers::FlatBufferBuilder::new_with_capacity(capacity),
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
                self.err = Some(String::from("Tried popping empty expression stack"));
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
                    self.err = Some(format!(
                        "expected {} on expr stack, got {}",
                        fbsemantic::enum_name_expression(kind),
                        fbsemantic::enum_name_expression(e)
                    ));
                    None
                }
            }
            None => {
                self.err = Some("Tried popping empty expression stack".to_string());
                None
            }
        }
    }

    fn pop_ident<T>(&mut self) -> Option<WIPOffset<T>> {
        match self.identifiers.pop() {
            None => {
                self.err = Some("Tried popping empty identifier stack".to_string());
                None
            }
            Some(wip) => Some(WIPOffset::new(wip.value())),
        }
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
                &mut self.builder,
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
                self.err = Some(String::from(
                    "Tried popping empty statement stack. Expected assignment on top of stack.",
                ));
                (None, fbsemantic::Assignment::NONE)
            }
            Some(_) => {
                self.err = Some(String::from(
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
            &mut self.builder,
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

fn fb_duration(d: &str) -> Result<fbsemantic::TimeUnit, String> {
    match d {
        "y" => Ok(fbsemantic::TimeUnit::y),
        "mo" => Ok(fbsemantic::TimeUnit::mo),
        "w" => Ok(fbsemantic::TimeUnit::w),
        "d" => Ok(fbsemantic::TimeUnit::d),
        "h" => Ok(fbsemantic::TimeUnit::h),
        "m" => Ok(fbsemantic::TimeUnit::m),
        "s" => Ok(fbsemantic::TimeUnit::s),
        "ms" => Ok(fbsemantic::TimeUnit::ms),
        "us" => Ok(fbsemantic::TimeUnit::us),
        "ns" => Ok(fbsemantic::TimeUnit::ns),
        s => Err(format!("unknown time unit {}", s)),
    }
}

#[cfg(test)]
mod tests;
