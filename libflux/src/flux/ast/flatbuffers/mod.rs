#![allow(clippy::all)]
#[allow(non_snake_case, unused)]
mod ast_generated;

use std::cell::RefCell;
use std::rc::Rc;

use crate::ast;
use crate::ast::walk;
use ast_generated::fbast;
use chrono::Offset;
use flatbuffers::{UnionWIPOffset, WIPOffset};

/// Accept the given AST package and return a FlatBuffers serialization of it as a Vec<u8>.
/// The FlatBuffers builder starts from the end of a buffer towards the beginning, so we must
/// also return a `usize` value which is the number of unused bytes at the start of the buffer.
pub fn serialize(ast_pkg: &ast::Package) -> Result<(Vec<u8>, usize), String> {
    // What would a good starting capacity be?
    let v = new_serializing_visitor_with_capacity(1024);
    walk::walk(&v, walk::Node::Package(ast_pkg));
    v.finish()
}

fn new_serializing_visitor_with_capacity<'a>(_capacity: usize) -> SerializingVisitor<'a> {
    SerializingVisitor {
        inner: Rc::new(RefCell::new(SerializingVisitorState::new_with_capacity(
            1024,
        ))),
    }
}

// Serializing to flatbuffers works like this:
// A node can only be serialized once all of its internal components are also serialized.
// This means a bottom-up walk of the tree.  We achieve this my matching on each node type
// in the `done()` method of an AST visitor.
//
// As elements are serialized, state is maintained in an instance of `SerializingVisitorInner`.
// In particular, each node in an expression tree is pushed on to a stack to be popped off
// by the node that consumes it.
struct SerializingVisitor<'a> {
    inner: Rc<RefCell<SerializingVisitorState<'a>>>,
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

impl<'a> ast::walk::Visitor<'a> for SerializingVisitor<'a> {
    fn visit(&self, _node: Rc<walk::Node<'a>>) -> Option<Self> {
        let v = self.inner.borrow();
        if let Some(_) = &v.err {
            return None;
        }
        Some(SerializingVisitor {
            inner: Rc::clone(&self.inner),
        })
    }

    fn done(&self, node: Rc<walk::Node<'a>>) {
        let mut v = &mut *self.inner.borrow_mut();
        if let Some(_) = &v.err {
            return;
        }
        let node = node.as_ref();
        let base_node = v.create_base_node(node.base());
        match node {
            walk::Node::IntegerLit(i) => {
                let i = fbast::IntegerLiteral::create(
                    &mut v.builder,
                    &fbast::IntegerLiteralArgs {
                        base_node,
                        value: i.value,
                    },
                );
                v.expr_stack
                    .push((i.as_union_value(), fbast::Expression::IntegerLiteral))
            }
            walk::Node::UintLit(ui) => {
                let ui = fbast::UnsignedIntegerLiteral::create(
                    &mut v.builder,
                    &fbast::UnsignedIntegerLiteralArgs {
                        base_node,
                        value: ui.value,
                    },
                );
                v.expr_stack.push((
                    ui.as_union_value(),
                    fbast::Expression::UnsignedIntegerLiteral,
                ))
            }
            walk::Node::FloatLit(f) => {
                let i = fbast::FloatLiteral::create(
                    &mut v.builder,
                    &fbast::FloatLiteralArgs {
                        base_node,
                        value: f.value,
                    },
                );
                v.expr_stack
                    .push((i.as_union_value(), fbast::Expression::FloatLiteral))
            }
            walk::Node::StringLit(s) => {
                let value = v.create_string(&s.value);
                let s = fbast::StringLiteral::create(
                    &mut v.builder,
                    &fbast::StringLiteralArgs { base_node, value },
                );
                v.expr_stack
                    .push((s.as_union_value(), fbast::Expression::StringLiteral))
            }
            walk::Node::DurationLit(dur_lit) => {
                let mut dur_vec: Vec<WIPOffset<fbast::Duration>> =
                    Vec::with_capacity(dur_lit.values.len());
                for dur in &dur_lit.values {
                    let unit = match fb_duration(&dur.unit) {
                        Ok(unit) => unit,
                        Err(s) => {
                            v.err = Some(s);
                            return;
                        }
                    };
                    let fb_dur = fbast::Duration::create(
                        &mut v.builder,
                        &fbast::DurationArgs {
                            magnitude: dur.magnitude,
                            unit,
                        },
                    );
                    dur_vec.push(fb_dur);
                }
                let values = Some(v.builder.create_vector(dur_vec.as_slice()));
                let dur_lit = fbast::DurationLiteral::create(
                    &mut v.builder,
                    &fbast::DurationLiteralArgs { base_node, values },
                );
                v.expr_stack
                    .push((dur_lit.as_union_value(), fbast::Expression::DurationLiteral))
            }
            walk::Node::BooleanLit(b) => {
                let b = fbast::BooleanLiteral::create(
                    &mut v.builder,
                    &fbast::BooleanLiteralArgs {
                        base_node,
                        value: b.value,
                    },
                );
                v.expr_stack
                    .push((b.as_union_value(), fbast::Expression::BooleanLiteral))
            }
            walk::Node::DateTimeLit(dtl) => {
                let secs = dtl.value.timestamp();
                let nsecs = dtl.value.timestamp_subsec_nanos();
                let offset = dtl.value.offset().fix().local_minus_utc();
                let dtl = fbast::DateTimeLiteral::create(
                    &mut v.builder,
                    &fbast::DateTimeLiteralArgs {
                        base_node,
                        secs,
                        nsecs,
                        offset,
                    },
                );
                v.expr_stack
                    .push((dtl.as_union_value(), fbast::Expression::DateTimeLiteral));
            }
            walk::Node::Identifier(id) => {
                let name = v.create_string(&id.name);
                let id = fbast::Identifier::create(
                    &mut v.builder,
                    &fbast::IdentifierArgs { base_node, name },
                );
                v.expr_stack
                    .push((id.as_union_value(), fbast::Expression::Identifier))
            }
            walk::Node::RegexpLit(rel) => {
                let value = v.create_string(&rel.value);
                let rel = fbast::RegexpLiteral::create(
                    &mut v.builder,
                    &fbast::RegexpLiteralArgs { base_node, value },
                );
                v.expr_stack
                    .push((rel.as_union_value(), fbast::Expression::RegexpLiteral))
            }
            walk::Node::PipeLit(_) => {
                let pl = fbast::PipeLiteral::create(
                    &mut v.builder,
                    &fbast::PipeLiteralArgs { base_node },
                );
                v.expr_stack
                    .push((pl.as_union_value(), fbast::Expression::PipeLiteral))
            }
            walk::Node::ArrayExpr(ae) => {
                let n_elems = ae.elements.len();
                let elements = {
                    let start = v.expr_stack.len() - n_elems;
                    let elems = &v.expr_stack.as_slice()[start..];
                    let mut wrapped_elems = Vec::with_capacity(n_elems);
                    for (e, et) in elems {
                        wrapped_elems.push(fbast::WrappedExpression::create(
                            &mut v.builder,
                            &fbast::WrappedExpressionArgs {
                                expr_type: *et,
                                expr: Some(*e),
                            },
                        ));
                    }
                    Some(v.builder.create_vector(wrapped_elems.as_slice()))
                };
                v.expr_stack.truncate(v.expr_stack.len() - n_elems);
                let ae = fbast::ArrayExpression::create(
                    &mut v.builder,
                    &fbast::ArrayExpressionArgs {
                        base_node,
                        elements,
                    },
                );
                v.expr_stack
                    .push((ae.as_union_value(), fbast::Expression::ArrayExpression))
            }
            walk::Node::Property(p) => {
                let (value, value_type) = match p.value {
                    None => (None, fbast::Expression::NONE),
                    Some(_) => v.pop_expr(),
                };
                let (key, key_type) = v.pop_property_key();
                let p = fbast::Property::create(
                    &mut v.builder,
                    &fbast::PropertyArgs {
                        base_node,
                        key_type,
                        key,
                        value_type,
                        value,
                    },
                );
                v.properties.push(p);
            }
            walk::Node::Block(bl) => {
                let body = {
                    let stmt_vec = v.create_stmt_vector(bl.body.len());
                    Some(v.builder.create_vector(&stmt_vec.as_slice()))
                };
                let bl =
                    fbast::Block::create(&mut v.builder, &fbast::BlockArgs { base_node, body });
                v.blocks.push(bl);
            }
            walk::Node::FunctionExpr(fe) => {
                let params = {
                    let params_vec = v.create_property_vector(fe.params.len());
                    let params_sl = params_vec.as_slice();
                    let vec = v.builder.create_vector(params_sl);
                    Some(vec)
                };
                let (body_type, body) = match &fe.body {
                    ast::FunctionBody::Expr(_) => {
                        // create a WrappedExpression for the body
                        let (expr, expr_type) = v.pop_expr();
                        let we = fbast::WrappedExpression::create(
                            &mut v.builder,
                            &fbast::WrappedExpressionArgs { expr_type, expr },
                        );
                        (
                            fbast::ExpressionOrBlock::WrappedExpression,
                            Some(we.as_union_value()),
                        )
                    }
                    ast::FunctionBody::Block(_) => {
                        let block = match v.blocks.pop() {
                            None => {
                                v.err = Some(String::from("pop empty block stack"));
                                return;
                            }
                            Some(b) => b,
                        };
                        (
                            fbast::ExpressionOrBlock::Block,
                            Some(block.as_union_value()),
                        )
                    }
                };
                let fe = fbast::FunctionExpression::create(
                    &mut v.builder,
                    &fbast::FunctionExpressionArgs {
                        base_node,
                        params,
                        body_type,
                        body,
                    },
                );
                v.expr_stack
                    .push((fe.as_union_value(), fbast::Expression::FunctionExpression));
            }
            walk::Node::LogicalExpr(le) => {
                let operator = fb_logical_operator(&le.operator);
                let (right, right_type) = v.pop_expr();
                let (left, left_type) = v.pop_expr();
                let le = fbast::LogicalExpression::create(
                    &mut v.builder,
                    &fbast::LogicalExpressionArgs {
                        base_node,
                        operator,
                        left_type,
                        left,
                        right_type,
                        right,
                    },
                );
                v.expr_stack
                    .push((le.as_union_value(), fbast::Expression::LogicalExpression));
            }
            walk::Node::ObjectExpr(oe) => {
                let properties = {
                    let prop_vec = v.create_property_vector(oe.properties.len());
                    let fb_prop_vec = v.builder.create_vector(&prop_vec.as_slice());
                    Some(fb_prop_vec)
                };
                let with = match oe.with {
                    None => None,
                    Some(_) => v.pop_expr_with_kind(fbast::Expression::Identifier),
                };
                let oe = fbast::ObjectExpression::create(
                    &mut v.builder,
                    &fbast::ObjectExpressionArgs {
                        base_node,
                        with,
                        properties,
                    },
                );
                v.expr_stack
                    .push((oe.as_union_value(), fbast::Expression::ObjectExpression));
            }
            walk::Node::MemberExpr(_) => {
                let (property, property_type) = v.pop_property_key();
                let (object, object_type) = v.pop_expr();
                let me = fbast::MemberExpression::create(
                    &mut v.builder,
                    &fbast::MemberExpressionArgs {
                        base_node,
                        object_type,
                        object,
                        property_type,
                        property,
                    },
                );
                v.expr_stack
                    .push((me.as_union_value(), fbast::Expression::MemberExpression));
            }
            walk::Node::IndexExpr(_) => {
                let (index, index_type) = v.pop_expr();
                let (array, array_type) = v.pop_expr();
                let ie = fbast::IndexExpression::create(
                    &mut v.builder,
                    &fbast::IndexExpressionArgs {
                        base_node,
                        array_type,
                        array,
                        index_type,
                        index,
                    },
                );
                v.expr_stack
                    .push((ie.as_union_value(), fbast::Expression::IndexExpression));
            }
            walk::Node::BinaryExpr(be) => {
                let (right, right_type) = v.pop_expr();
                let (left, left_type) = v.pop_expr();
                let be = fbast::BinaryExpression::create(
                    &mut v.builder,
                    &fbast::BinaryExpressionArgs {
                        base_node,
                        operator: fb_operator(&be.operator),
                        left_type,
                        left,
                        right_type,
                        right,
                    },
                );
                v.expr_stack
                    .push((be.as_union_value(), fbast::Expression::BinaryExpression))
            }
            walk::Node::UnaryExpr(ue) => {
                let operator = fb_operator(&ue.operator);
                let (argument, argument_type) = v.pop_expr();
                let ue = fbast::UnaryExpression::create(
                    &mut v.builder,
                    &fbast::UnaryExpressionArgs {
                        base_node,
                        operator,
                        argument,
                        argument_type,
                    },
                );
                v.expr_stack
                    .push((ue.as_union_value(), fbast::Expression::UnaryExpression));
            }
            walk::Node::PipeExpr(_) => {
                let call = v.pop_expr_with_kind(fbast::Expression::CallExpression);
                let (argument, argument_type) = v.pop_expr();
                let pe = fbast::PipeExpression::create(
                    &mut v.builder,
                    &fbast::PipeExpressionArgs {
                        base_node,
                        argument_type,
                        argument,
                        call,
                    },
                );
                v.expr_stack
                    .push((pe.as_union_value(), fbast::Expression::PipeExpression));
            }
            walk::Node::CallExpr(ce) => {
                let arguments = match ce.arguments.len() {
                    0 => None,
                    1 => v.pop_expr_with_kind(fbast::Expression::ObjectExpression),
                    _ => {
                        v.err = Some(String::from("found call with more than one argument"));
                        return;
                    }
                };
                let (callee, callee_type) = v.pop_expr();
                let ce = fbast::CallExpression::create(
                    &mut v.builder,
                    &fbast::CallExpressionArgs {
                        base_node,
                        callee,
                        callee_type,
                        arguments,
                    },
                );
                v.expr_stack
                    .push((ce.as_union_value(), fbast::Expression::CallExpression));
            }
            walk::Node::ConditionalExpr(_) => {
                let (alternate, alternate_type) = v.pop_expr();
                let (consequent, consequent_type) = v.pop_expr();
                let (test, test_type) = v.pop_expr();

                let ce = fbast::ConditionalExpression::create(
                    &mut v.builder,
                    &fbast::ConditionalExpressionArgs {
                        base_node,
                        test_type,
                        test,
                        consequent_type,
                        consequent,
                        alternate_type,
                        alternate,
                    },
                );
                v.expr_stack.push((
                    ce.as_union_value(),
                    fbast::Expression::ConditionalExpression,
                ))
            }
            walk::Node::TextPart(tp) => {
                let text_value = Some(v.builder.create_string(tp.value.as_str()));
                let sep = fbast::StringExpressionPart::create(
                    &mut v.builder,
                    &fbast::StringExpressionPartArgs {
                        base_node,
                        text_value,
                        ..fbast::StringExpressionPartArgs::default()
                    },
                );
                v.string_expr_parts.push(sep);
            }
            walk::Node::InterpolatedPart(_) => {
                let (interpolated_expression, interpolated_expression_type) = v.pop_expr();
                let sep = fbast::StringExpressionPart::create(
                    &mut v.builder,
                    &fbast::StringExpressionPartArgs {
                        base_node,
                        interpolated_expression_type,
                        interpolated_expression,
                        ..fbast::StringExpressionPartArgs::default()
                    },
                );
                v.string_expr_parts.push(sep);
            }
            walk::Node::StringExpr(se) => {
                let parts = {
                    let n_parts = se.parts.len();
                    let start = v.string_expr_parts.len() - n_parts;
                    let parts_sl = &v.string_expr_parts.as_slice()[start..];
                    let vec = v.builder.create_vector(parts_sl);
                    v.string_expr_parts.truncate(start);
                    Some(vec)
                };
                let se = fbast::StringExpression::create(
                    &mut v.builder,
                    &fbast::StringExpressionArgs { base_node, parts },
                );
                v.expr_stack
                    .push((se.as_union_value(), fbast::Expression::StringExpression));
            }
            walk::Node::ParenExpr(_) => {
                let (expression, expression_type) = v.pop_expr();
                let pe = fbast::ParenExpression::create(
                    &mut v.builder,
                    &fbast::ParenExpressionArgs {
                        base_node,
                        expression_type,
                        expression,
                    },
                );
                v.expr_stack
                    .push((pe.as_union_value(), fbast::Expression::ParenExpression));
            }
            walk::Node::BadExpr(be) => {
                let (expression, expression_type) = match &be.expression {
                    None => (None, fbast::Expression::NONE),
                    Some(_) => v.pop_expr(),
                };
                let text = v.create_string(&be.text);
                let be = fbast::BadExpression::create(
                    &mut v.builder,
                    &fbast::BadExpressionArgs {
                        base_node,
                        expression_type,
                        expression,
                        text,
                    },
                );
                v.expr_stack
                    .push((be.as_union_value(), fbast::Expression::BadExpression));
            }
            walk::Node::VariableAssgn(_) => {
                let (init_, init_type) = v.pop_expr();
                let id = v.pop_expr_with_kind(fbast::Expression::Identifier);
                let va = fbast::VariableAssignment::create(
                    &mut v.builder,
                    &fbast::VariableAssignmentArgs {
                        base_node,
                        id,
                        init__type: init_type,
                        init_,
                    },
                );
                v.stmts
                    .push((va.as_union_value(), fbast::Statement::VariableAssignment));
            }
            walk::Node::MemberAssgn(_) => {
                let (init_, init_type) = v.pop_expr();
                let member = v.pop_expr_with_kind(fbast::Expression::MemberExpression);
                let ma = fbast::MemberAssignment::create(
                    &mut v.builder,
                    &fbast::MemberAssignmentArgs {
                        base_node,
                        member,
                        init__type: init_type,
                        init_,
                    },
                );
                v.member_assign = Some(ma.as_union_value());
            }
            walk::Node::ExprStmt(_) => {
                let (expression, expression_type) = v.pop_expr();
                let es = fbast::ExpressionStatement::create(
                    &mut v.builder,
                    &fbast::ExpressionStatementArgs {
                        base_node,
                        expression_type,
                        expression,
                    },
                );
                v.stmts
                    .push((es.as_union_value(), fbast::Statement::ExpressionStatement));
            }
            walk::Node::OptionStmt(os) => {
                let (assignment, assignment_type) = match os.assignment {
                    ast::Assignment::Variable(_) => v.pop_assignment_stmt(),
                    ast::Assignment::Member(_) => match v.member_assign {
                        None => {
                            v.err = Some(String::from("expected member assignment"));
                            return;
                        }
                        ma => (ma, fbast::Assignment::MemberAssignment),
                    },
                };
                let os = fbast::OptionStatement::create(
                    &mut v.builder,
                    &fbast::OptionStatementArgs {
                        base_node,
                        assignment_type,
                        assignment,
                    },
                );
                v.stmts
                    .push((os.as_union_value(), fbast::Statement::OptionStatement));
            }
            walk::Node::ReturnStmt(_) => {
                let (argument, argument_type) = v.pop_expr();
                let rs = fbast::ReturnStatement::create(
                    &mut v.builder,
                    &fbast::ReturnStatementArgs {
                        base_node,
                        argument_type,
                        argument,
                    },
                );
                v.stmts
                    .push((rs.as_union_value(), fbast::Statement::ReturnStatement));
            }
            walk::Node::BadStmt(bs) => {
                let text = Some(v.builder.create_string(bs.text.as_str()));
                let bs = fbast::BadStatement::create(
                    &mut v.builder,
                    &fbast::BadStatementArgs { base_node, text },
                );
                v.stmts
                    .push((bs.as_union_value(), fbast::Statement::BadStatement))
            }
            walk::Node::TestStmt(_) => {
                let (assignment, assignment_type) = v.pop_assignment_stmt();
                let ts = fbast::TestStatement::create(
                    &mut v.builder,
                    &fbast::TestStatementArgs {
                        base_node,
                        assignment_type,
                        assignment,
                    },
                );
                v.stmts
                    .push((ts.as_union_value(), fbast::Statement::TestStatement));
            }
            walk::Node::BuiltinStmt(_) => {
                let id = v.pop_expr_with_kind(fbast::Expression::Identifier);
                let bs = fbast::BuiltinStatement::create(
                    &mut v.builder,
                    &fbast::BuiltinStatementArgs { base_node, id },
                );
                v.stmts
                    .push((bs.as_union_value(), fbast::Statement::BuiltinStatement));
            }
            walk::Node::ImportDeclaration(id) => {
                let path = v.pop_expr_with_kind(fbast::Expression::StringLiteral);
                let as_ = match id.alias {
                    Some(_) => v.pop_expr_with_kind(fbast::Expression::Identifier),
                    None => None,
                };
                let id = fbast::ImportDeclaration::create(
                    &mut v.builder,
                    &fbast::ImportDeclarationArgs {
                        base_node,
                        as_,
                        path,
                    },
                );
                v.import_decls.push(id)
            }
            walk::Node::PackageClause(_) => {
                let name = v.pop_expr_with_kind(fbast::Expression::Identifier);
                v.package_clause = Some(fbast::PackageClause::create(
                    &mut v.builder,
                    &fbast::PackageClauseArgs { name, base_node },
                ));
            }
            walk::Node::File(f) => {
                let name = v.create_string(&f.name);
                let metadata = v.create_string(&f.metadata);
                let package = v.package_clause;
                v.package_clause = None;

                let imports = { Some(v.builder.create_vector(v.import_decls.as_slice())) };
                v.import_decls.clear();

                let stmt_vec = v.create_stmt_vector(f.body.len());
                let body = Some(v.builder.create_vector(&stmt_vec.as_slice()));
                let f = fbast::File::create(
                    &mut v.builder,
                    &fbast::FileArgs {
                        base_node,
                        name,
                        metadata,
                        package,
                        imports,
                        body,
                        ..fbast::FileArgs::default()
                    },
                );
                v.files.push(f);
            }
            walk::Node::Package(p) => {
                let path = v.create_string(&p.path);
                let package = v.create_string(&p.package);
                let files = {
                    let mut fs: Vec<WIPOffset<fbast::File>> = Vec::new();
                    std::mem::swap(&mut v.files, &mut fs);
                    Some(v.builder.create_vector(fs.as_slice()))
                };
                v.package = Some(fbast::Package::create(
                    &mut v.builder,
                    &fbast::PackageArgs {
                        base_node,
                        path,
                        package,
                        files,
                    },
                ));
            }
        };
    }
}

struct SerializingVisitorState<'a> {
    // Any error that occurred during serialization, returned by the visitor's finish method.
    err: Option<String>,

    builder: flatbuffers::FlatBufferBuilder<'a>,

    package: Option<WIPOffset<fbast::Package<'a>>>,
    package_clause: Option<WIPOffset<fbast::PackageClause<'a>>>,
    import_decls: Vec<WIPOffset<fbast::ImportDeclaration<'a>>>,
    files: Vec<WIPOffset<fbast::File<'a>>>,
    blocks: Vec<WIPOffset<fbast::Block<'a>>>,
    stmts: Vec<(WIPOffset<UnionWIPOffset>, fbast::Statement)>,

    expr_stack: Vec<(WIPOffset<UnionWIPOffset>, fbast::Expression)>,
    properties: Vec<WIPOffset<fbast::Property<'a>>>,
    string_expr_parts: Vec<WIPOffset<fbast::StringExpressionPart<'a>>>,
    member_assign: Option<WIPOffset<UnionWIPOffset>>,
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
            expr_stack: Vec::new(),
            properties: Vec::new(),
            string_expr_parts: Vec::new(),
            member_assign: None,
        }
    }

    fn pop_expr(&mut self) -> (Option<WIPOffset<UnionWIPOffset>>, fbast::Expression) {
        match self.expr_stack.pop() {
            None => {
                self.err = Some(String::from("pop empty expr stack"));
                return (None, fbast::Expression::NONE);
            }
            Some((o, e)) => (Some(o), e),
        }
    }

    fn pop_expr_with_kind<T>(&mut self, kind: fbast::Expression) -> Option<WIPOffset<T>> {
        match self.expr_stack.pop() {
            Some((wipo, e)) => {
                if e == kind {
                    Some(WIPOffset::new(wipo.value()))
                } else {
                    self.err = Some(String::from(format!(
                        "expected {} on expr stack, got {}",
                        fbast::enum_name_expression(kind),
                        fbast::enum_name_expression(e)
                    )));
                    return None;
                }
            }
            None => {
                self.err = Some(String::from("pop empty expr stack"));
                return None;
            }
        }
    }

    fn create_string(&mut self, str: &String) -> Option<WIPOffset<&'a str>> {
        Some(self.builder.create_string(str.as_str()))
    }

    fn create_opt_string(&mut self, str: &Option<String>) -> Option<WIPOffset<&'a str>> {
        match str {
            None => None,
            Some(str) => Some(self.builder.create_string(str.as_str())),
        }
    }

    fn create_stmt_vector(
        &mut self,
        n_stmts: usize,
    ) -> Vec<WIPOffset<fbast::WrappedStatement<'a>>> {
        let start = self.stmts.len() - n_stmts;
        let union_stmts = &self.stmts.as_slice()[start..];
        let mut wrapped_stmts: Vec<WIPOffset<fbast::WrappedStatement>> =
            Vec::with_capacity(n_stmts);
        for (st, st_ty) in union_stmts {
            let wrapped_st = fbast::WrappedStatement::create(
                &mut self.builder,
                &fbast::WrappedStatementArgs {
                    statement_type: *st_ty,
                    statement: Some(*st),
                },
            );
            wrapped_stmts.push(wrapped_st);
        }
        self.stmts.truncate(start);
        wrapped_stmts
    }

    fn create_property_vector(&mut self, n_props: usize) -> Vec<WIPOffset<fbast::Property<'a>>> {
        let start = self.properties.len() - n_props;
        self.properties.split_off(start)
    }

    fn pop_property_key(&mut self) -> (Option<WIPOffset<UnionWIPOffset>>, fbast::PropertyKey) {
        match self.pop_expr() {
            (offset, fbast::Expression::Identifier) => (offset, fbast::PropertyKey::Identifier),
            (offset, fbast::Expression::StringLiteral) => {
                (offset, fbast::PropertyKey::StringLiteral)
            }
            _ => {
                self.err = Some(String::from(
                    "unexpected expression on stack for property key",
                ));
                (None, fbast::PropertyKey::NONE)
            }
        }
    }

    fn pop_assignment_stmt(&mut self) -> (Option<WIPOffset<UnionWIPOffset>>, fbast::Assignment) {
        match self.stmts.pop() {
            Some((va, fbast::Statement::VariableAssignment)) => {
                (Some(va), fbast::Assignment::VariableAssignment)
            }
            None => {
                self.err = Some(String::from("pop empty stmt stack; expected assignment"));
                (None, fbast::Assignment::NONE)
            }
            Some(_) => {
                self.err = Some(String::from("expected assignment on top of stmt stack"));
                (None, fbast::Assignment::NONE)
            }
        }
    }

    fn create_base_node(
        &mut self,
        base_node: &ast::BaseNode,
    ) -> Option<WIPOffset<fbast::BaseNode<'a>>> {
        let loc = self.create_loc(&base_node.location);
        let errors = self.create_base_node_errs(&base_node.errors);
        Some(fbast::BaseNode::create(
            &mut self.builder,
            &fbast::BaseNodeArgs { loc, errors },
        ))
    }

    fn create_loc(
        &mut self,
        loc: &ast::SourceLocation,
    ) -> Option<WIPOffset<fbast::SourceLocation<'a>>> {
        let file = self.create_opt_string(&loc.file);
        let source = self.create_opt_string(&loc.source);
        Some(fbast::SourceLocation::create(
            &mut self.builder,
            &fbast::SourceLocationArgs {
                file,
                start: Some(&fbast::Position::new(
                    loc.start.line as i32,
                    loc.start.column as i32,
                )),
                end: Some(&fbast::Position::new(
                    loc.end.line as i32,
                    loc.end.column as i32,
                )),
                source,
            },
        ))
    }

    fn create_base_node_errs(
        &mut self,
        ast_errs: &Vec<String>,
    ) -> Option<
        flatbuffers::WIPOffset<flatbuffers::Vector<'a, flatbuffers::ForwardsUOffset<&'a str>>>,
    > {
        Some(
            self.builder.create_vector_of_strings(
                ast_errs
                    .iter()
                    .map(|s| s.as_str())
                    .collect::<Vec<&str>>()
                    .as_slice(),
            ),
        )
    }
}

// This is a convenience function for debugging.
#[allow(dead_code)]
fn print_expr_stack(st: &Vec<(WIPOffset<UnionWIPOffset>, fbast::Expression)>) {
    if st.len() == 0 {
        return;
    }
    for (_, et) in &st[..st.len() - 1] {
        print!(" {}", fbast::enum_name_expression(*et));
    }
    let (_, et) = st[st.len() - 1];
    println!(" -> {}", fbast::enum_name_expression(et));
}

fn fb_operator(o: &ast::Operator) -> fbast::Operator {
    match o {
        ast::Operator::MultiplicationOperator => fbast::Operator::MultiplicationOperator,
        ast::Operator::DivisionOperator => fbast::Operator::DivisionOperator,
        ast::Operator::ModuloOperator => fbast::Operator::ModuloOperator,
        ast::Operator::PowerOperator => fbast::Operator::PowerOperator,
        ast::Operator::AdditionOperator => fbast::Operator::AdditionOperator,
        ast::Operator::SubtractionOperator => fbast::Operator::SubtractionOperator,
        ast::Operator::LessThanEqualOperator => fbast::Operator::LessThanEqualOperator,
        ast::Operator::LessThanOperator => fbast::Operator::LessThanOperator,
        ast::Operator::GreaterThanEqualOperator => fbast::Operator::GreaterThanEqualOperator,
        ast::Operator::GreaterThanOperator => fbast::Operator::GreaterThanOperator,
        ast::Operator::StartsWithOperator => fbast::Operator::StartsWithOperator,
        ast::Operator::InOperator => fbast::Operator::InOperator,
        ast::Operator::NotOperator => fbast::Operator::NotOperator,
        ast::Operator::ExistsOperator => fbast::Operator::ExistsOperator,
        ast::Operator::NotEmptyOperator => fbast::Operator::NotEmptyOperator,
        ast::Operator::EmptyOperator => fbast::Operator::EmptyOperator,
        ast::Operator::EqualOperator => fbast::Operator::EqualOperator,
        ast::Operator::NotEqualOperator => fbast::Operator::NotEqualOperator,
        ast::Operator::RegexpMatchOperator => fbast::Operator::RegexpMatchOperator,
        ast::Operator::NotRegexpMatchOperator => fbast::Operator::NotRegexpMatchOperator,
        ast::Operator::InvalidOperator => fbast::Operator::InvalidOperator,
    }
}

fn fb_logical_operator(lo: &ast::LogicalOperator) -> fbast::LogicalOperator {
    match lo {
        ast::LogicalOperator::AndOperator => fbast::LogicalOperator::AndOperator,
        ast::LogicalOperator::OrOperator => fbast::LogicalOperator::OrOperator,
    }
}

fn fb_duration(d: &String) -> Result<fbast::TimeUnit, String> {
    match d.as_str() {
        "y" => Ok(fbast::TimeUnit::y),
        "mo" => Ok(fbast::TimeUnit::mo),
        "w" => Ok(fbast::TimeUnit::w),
        "d" => Ok(fbast::TimeUnit::d),
        "h" => Ok(fbast::TimeUnit::h),
        "m" => Ok(fbast::TimeUnit::m),
        "s" => Ok(fbast::TimeUnit::s),
        "ms" => Ok(fbast::TimeUnit::ms),
        "us" => Ok(fbast::TimeUnit::us),
        "ns" => Ok(fbast::TimeUnit::ns),
        s => Err(String::from(format!("unknown time unit {}", s))),
    }
}

#[cfg(test)]
mod tests;
