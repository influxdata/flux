//! Walking the AST.

#[cfg(test)]
mod tests;

use derive_more::Display;

use crate::ast::*;

/// Node represents any structure that can appear in the AST.
#[derive(Debug, Display, Copy, Clone, derive_more::From)]
#[allow(missing_docs)]
pub enum Node<'a> {
    #[display(fmt = "Package")]
    Package(&'a Package),
    #[display(fmt = "File")]
    File(&'a File),
    #[display(fmt = "PackageClause")]
    PackageClause(&'a PackageClause),
    #[display(fmt = "ImportDeclaration")]
    ImportDeclaration(&'a ImportDeclaration),

    // Expressions
    #[display(fmt = "Identifier")]
    Identifier(&'a Identifier),

    #[display(fmt = "ArrayExpr")]
    ArrayExpr(&'a ArrayExpr),
    #[display(fmt = "DictExpr")]
    DictExpr(&'a DictExpr),
    #[display(fmt = "FunctionExpr")]
    FunctionExpr(&'a FunctionExpr),
    #[display(fmt = "LogicalExpr")]
    LogicalExpr(&'a LogicalExpr),
    #[display(fmt = "ObjectExpr")]
    ObjectExpr(&'a ObjectExpr),
    #[display(fmt = "MemberExpr")]
    MemberExpr(&'a MemberExpr),
    #[display(fmt = "IndexExpr")]
    IndexExpr(&'a IndexExpr),
    #[display(fmt = "BinaryExpr")]
    BinaryExpr(&'a BinaryExpr),
    #[display(fmt = "UnaryExpr")]
    UnaryExpr(&'a UnaryExpr),
    #[display(fmt = "PipeExpr")]
    PipeExpr(&'a PipeExpr),
    #[display(fmt = "CallExpr")]
    CallExpr(&'a CallExpr),
    #[display(fmt = "ConditionalExpr")]
    ConditionalExpr(&'a ConditionalExpr),
    #[display(fmt = "StringExpr")]
    StringExpr(&'a StringExpr),
    #[display(fmt = "ParenExpr")]
    ParenExpr(&'a ParenExpr),

    #[display(fmt = "IntegerLit")]
    IntegerLit(&'a IntegerLit),
    #[display(fmt = "FloatLit")]
    FloatLit(&'a FloatLit),
    #[display(fmt = "StringLit")]
    StringLit(&'a StringLit),
    #[display(fmt = "DurationLit")]
    DurationLit(&'a DurationLit),
    #[display(fmt = "UintLit")]
    UintLit(&'a UintLit),
    #[display(fmt = "BooleanLit")]
    BooleanLit(&'a BooleanLit),
    #[display(fmt = "DateTimeLit")]
    DateTimeLit(&'a DateTimeLit),
    #[display(fmt = "RegexpLit")]
    RegexpLit(&'a RegexpLit),
    #[display(fmt = "PipeLit")]
    PipeLit(&'a PipeLit),
    #[display(fmt = "LabelLit")]
    LabelLit(&'a LabelLit),

    #[display(fmt = "BadExpr")]
    BadExpr(&'a BadExpr),

    // Statements
    #[display(fmt = "ExprStmt")]
    ExprStmt(&'a ExprStmt),
    #[display(fmt = "OptionStmt")]
    OptionStmt(&'a OptionStmt),
    #[display(fmt = "ReturnStmt")]
    ReturnStmt(&'a ReturnStmt),
    #[display(fmt = "BadStmt")]
    BadStmt(&'a BadStmt),
    #[display(fmt = "TestCaseStmt")]
    TestCaseStmt(&'a TestCaseStmt),
    #[display(fmt = "BuiltinStmt")]
    BuiltinStmt(&'a BuiltinStmt),

    // FunctionBlock
    #[display(fmt = "Block")]
    Block(&'a Block),

    // Property
    #[display(fmt = "Property")]
    Property(&'a Property),

    // StringExprPart
    #[display(fmt = "TextPart")]
    TextPart(&'a TextPart),
    #[display(fmt = "InterpolatedPart")]
    InterpolatedPart(&'a InterpolatedPart),

    // Assignment
    #[display(fmt = "VariableAssgn")]
    VariableAssgn(&'a VariableAssgn),
    #[display(fmt = "MemberAssgn")]
    MemberAssgn(&'a MemberAssgn),

    #[display(fmt = "TypeExpression")]
    TypeExpression(&'a TypeExpression),
    #[display(fmt = "MonoType")]
    MonoType(&'a MonoType),
    #[display(fmt = "PropertyType")]
    PropertyType(&'a PropertyType),
    #[display(fmt = "ParameterType")]
    ParameterType(&'a ParameterType),
    #[display(fmt = "TypeConstraint")]
    TypeConstraint(&'a TypeConstraint),
}

impl<'a> Node<'a> {
    #[allow(missing_docs)]
    pub fn base(&self) -> &'a BaseNode {
        match self {
            Node::Package(n) => &n.base,
            Node::File(n) => &n.base,
            Node::PackageClause(n) => &n.base,
            Node::ImportDeclaration(n) => &n.base,
            Node::Identifier(n) => &n.base,
            Node::ArrayExpr(n) => &n.base,
            Node::DictExpr(n) => &n.base,
            Node::FunctionExpr(n) => &n.base,
            Node::LogicalExpr(n) => &n.base,
            Node::ObjectExpr(n) => &n.base,
            Node::MemberExpr(n) => &n.base,
            Node::IndexExpr(n) => &n.base,
            Node::BinaryExpr(n) => &n.base,
            Node::UnaryExpr(n) => &n.base,
            Node::PipeExpr(n) => &n.base,
            Node::CallExpr(n) => &n.base,
            Node::ConditionalExpr(n) => &n.base,
            Node::StringExpr(n) => &n.base,
            Node::ParenExpr(n) => &n.base,
            Node::IntegerLit(n) => &n.base,
            Node::FloatLit(n) => &n.base,
            Node::StringLit(n) => &n.base,
            Node::DurationLit(n) => &n.base,
            Node::UintLit(n) => &n.base,
            Node::BooleanLit(n) => &n.base,
            Node::DateTimeLit(n) => &n.base,
            Node::RegexpLit(n) => &n.base,
            Node::PipeLit(n) => &n.base,
            Node::LabelLit(n) => &n.base,
            Node::BadExpr(n) => &n.base,
            Node::ExprStmt(n) => &n.base,
            Node::OptionStmt(n) => &n.base,
            Node::ReturnStmt(n) => &n.base,
            Node::BadStmt(n) => &n.base,
            Node::TestCaseStmt(n) => &n.base,
            Node::BuiltinStmt(n) => &n.base,
            Node::Block(n) => &n.base,
            Node::Property(n) => &n.base,
            Node::TextPart(n) => &n.base,
            Node::InterpolatedPart(n) => &n.base,
            Node::VariableAssgn(n) => &n.base,
            Node::MemberAssgn(n) => &n.base,
            Node::TypeExpression(n) => &n.base,
            Node::MonoType(n) => n.base(),
            Node::PropertyType(n) => &n.base,
            Node::ParameterType(n) => n.base(),
            Node::TypeConstraint(n) => &n.base,
        }
    }
}

impl<'a> Node<'a> {
    #[allow(missing_docs)]
    pub fn from_expr(expr: &'a Expression) -> Node {
        match expr {
            Expression::Identifier(e) => Node::Identifier(e),
            Expression::Array(e) => Node::ArrayExpr(e),
            Expression::Dict(e) => Node::DictExpr(e),
            Expression::Function(e) => Node::FunctionExpr(e),
            Expression::Logical(e) => Node::LogicalExpr(e),
            Expression::Object(e) => Node::ObjectExpr(e),
            Expression::Member(e) => Node::MemberExpr(e),
            Expression::Index(e) => Node::IndexExpr(e),
            Expression::Binary(e) => Node::BinaryExpr(e),
            Expression::Unary(e) => Node::UnaryExpr(e),
            Expression::PipeExpr(e) => Node::PipeExpr(e),
            Expression::Call(e) => Node::CallExpr(e),
            Expression::Conditional(e) => Node::ConditionalExpr(e),
            Expression::StringExpr(e) => Node::StringExpr(e),
            Expression::Paren(e) => Node::ParenExpr(e),
            Expression::Integer(e) => Node::IntegerLit(e),
            Expression::Float(e) => Node::FloatLit(e),
            Expression::StringLit(e) => Node::StringLit(e),
            Expression::Duration(e) => Node::DurationLit(e),
            Expression::Uint(e) => Node::UintLit(e),
            Expression::Boolean(e) => Node::BooleanLit(e),
            Expression::DateTime(e) => Node::DateTimeLit(e),
            Expression::Regexp(e) => Node::RegexpLit(e),
            Expression::PipeLit(e) => Node::PipeLit(e),
            Expression::Label(e) => Node::LabelLit(e),
            Expression::Bad(e) => Node::BadExpr(e),
        }
    }
    #[allow(missing_docs)]
    pub fn from_stmt(stmt: &Statement) -> Node {
        match stmt {
            Statement::Expr(s) => Node::ExprStmt(s),
            Statement::Variable(s) => Node::VariableAssgn(s),
            Statement::Option(s) => Node::OptionStmt(s),
            Statement::Return(s) => Node::ReturnStmt(s),
            Statement::Bad(s) => Node::BadStmt(s),
            Statement::TestCase(s) => Node::TestCaseStmt(s),
            Statement::Builtin(s) => Node::BuiltinStmt(s),
        }
    }
    fn from_function_body(fb: &FunctionBody) -> Node {
        match fb {
            FunctionBody::Block(b) => Node::Block(b),
            FunctionBody::Expr(e) => Node::from_expr(e),
        }
    }
    fn from_property_key(pk: &PropertyKey) -> Node {
        match pk {
            PropertyKey::Identifier(i) => Node::Identifier(i),
            PropertyKey::StringLit(s) => Node::StringLit(s),
        }
    }
    fn from_string_expr_part(sp: &StringExprPart) -> Node {
        match sp {
            StringExprPart::Text(t) => Node::TextPart(t),
            StringExprPart::Interpolated(e) => Node::InterpolatedPart(e),
        }
    }
    fn from_assignment(a: &Assignment) -> Node {
        match a {
            Assignment::Variable(v) => Node::VariableAssgn(v),
            Assignment::Member(m) => Node::MemberAssgn(m),
        }
    }
}

/// Visitor defines a visitor pattern for walking the AST.
///
/// When used with the walk function, Visit will be called for every node
/// in depth-first order. After all children for a Node have been visted,
/// Done is called on that Node to signal that we are done with that Node.
///
/// If Visit returns None, walk will not recurse on the children.
///
/// Note: the Rc in visit and done is to allow for multiple ownership of a node, i.e.
///       a visitor can own a node as well as the walk funciton. This allows
///       for nodes to persist outside the scope of the walk function and to
///       be cleaned up once all owners have let go of the reference.
///
/// See example with `FuncVisitor` below in this file.
pub trait Visitor<'a>: Sized {
    /// Visit is called for a node.
    /// When the `Visitor` is used in [`walk()`], the boolean value returned
    /// is used to continue walking (`true`) or stop (`false`).
    fn visit(&mut self, node: Node<'a>) -> bool;
    /// Done is called for a node once it has been visited along with all of its children.
    fn done(&mut self, _: Node<'a>) {} // default is to do nothing
}

/// Walk recursively visits children of a node.
/// Nodes are visited in depth-first order.
pub fn walk<'a, T>(v: &mut T, node: Node<'a>)
where
    T: Visitor<'a>,
{
    if v.visit(node) {
        match node {
            Node::Package(n) => {
                for file in n.files.iter() {
                    walk(v, Node::File(file));
                }
            }
            Node::File(n) => {
                if let Some(pkg) = &n.package {
                    walk(v, Node::PackageClause(pkg));
                }
                for imp in n.imports.iter() {
                    walk(v, Node::ImportDeclaration(imp));
                }
                for stmt in n.body.iter() {
                    walk(v, Node::from_stmt(stmt));
                }
            }
            Node::PackageClause(n) => {
                walk(v, Node::Identifier(&n.name));
            }
            Node::ImportDeclaration(n) => {
                if let Some(alias) = &n.alias {
                    walk(v, Node::Identifier(alias));
                }
                walk(v, Node::StringLit(&n.path));
            }
            Node::Identifier(_) => {}
            Node::ArrayExpr(n) => {
                for element in n.elements.iter() {
                    walk(v, Node::from_expr(&element.expression));
                }
            }
            Node::DictExpr(n) => {
                for element in n.elements.iter() {
                    walk(v, Node::from_expr(&element.key));
                    walk(v, Node::from_expr(&element.val));
                }
            }
            Node::FunctionExpr(n) => {
                for param in n.params.iter() {
                    walk(v, Node::Property(param));
                }
                walk(v, Node::from_function_body(&n.body));
            }
            Node::LogicalExpr(n) => {
                walk(v, Node::from_expr(&n.left));
                walk(v, Node::from_expr(&n.right));
            }
            Node::ObjectExpr(n) => {
                if let Some(ws) = &n.with {
                    walk(v, Node::Identifier(&ws.source));
                }
                for prop in n.properties.iter() {
                    walk(v, Node::Property(prop));
                }
            }
            Node::MemberExpr(n) => {
                walk(v, Node::from_expr(&n.object));
                walk(v, Node::from_property_key(&n.property));
            }
            Node::IndexExpr(n) => {
                walk(v, Node::from_expr(&n.array));
                walk(v, Node::from_expr(&n.index));
            }
            Node::BinaryExpr(n) => {
                walk(v, Node::from_expr(&n.left));
                walk(v, Node::from_expr(&n.right));
            }
            Node::UnaryExpr(n) => {
                walk(v, Node::from_expr(&n.argument));
            }
            Node::PipeExpr(n) => {
                walk(v, Node::from_expr(&n.argument));
                walk(v, Node::CallExpr(&n.call));
            }
            Node::CallExpr(n) => {
                walk(v, Node::from_expr(&n.callee));
                for arg in n.arguments.iter() {
                    walk(v, Node::from_expr(arg));
                }
            }
            Node::ConditionalExpr(n) => {
                walk(v, Node::from_expr(&n.test));
                walk(v, Node::from_expr(&n.consequent));
                walk(v, Node::from_expr(&n.alternate));
            }
            Node::StringExpr(n) => {
                for part in n.parts.iter() {
                    walk(v, Node::from_string_expr_part(part));
                }
            }
            Node::ParenExpr(n) => {
                walk(v, Node::from_expr(&n.expression));
            }
            Node::IntegerLit(_) => {}
            Node::FloatLit(_) => {}
            Node::StringLit(_) => {}
            Node::DurationLit(_) => {}
            Node::UintLit(_) => {}
            Node::BooleanLit(_) => {}
            Node::DateTimeLit(_) => {}
            Node::RegexpLit(_) => {}
            Node::PipeLit(_) => {}
            Node::LabelLit(_) => {}
            Node::BadExpr(n) => {
                if let Some(e) = &n.expression {
                    walk(v, Node::from_expr(e));
                }
            }
            Node::ExprStmt(n) => {
                walk(v, Node::from_expr(&n.expression));
            }
            Node::OptionStmt(n) => {
                walk(v, Node::from_assignment(&n.assignment));
            }
            Node::ReturnStmt(n) => {
                walk(v, Node::from_expr(&n.argument));
            }
            Node::BadStmt(_) => {}
            Node::TestCaseStmt(n) => {
                walk(v, Node::Identifier(&n.id));
                walk(v, Node::Block(&n.block));
            }
            Node::BuiltinStmt(n) => {
                walk(v, Node::Identifier(&n.id));
                walk(v, Node::TypeExpression(&n.ty));
            }
            Node::Block(n) => {
                for s in n.body.iter() {
                    walk(v, Node::from_stmt(s));
                }
            }
            Node::Property(n) => {
                walk(v, Node::from_property_key(&n.key));
                if let Some(value) = &n.value {
                    walk(v, Node::from_expr(value));
                }
            }
            Node::TextPart(_) => {}
            Node::InterpolatedPart(n) => {
                walk(v, Node::from_expr(&n.expression));
            }
            Node::VariableAssgn(n) => {
                walk(v, Node::Identifier(&n.id));
                walk(v, Node::from_expr(&n.init));
            }
            Node::MemberAssgn(n) => {
                walk(v, Node::MemberExpr(&n.member));
                walk(v, Node::from_expr(&n.init));
            }
            Node::TypeExpression(n) => {
                walk(v, Node::MonoType(&n.monotype));
                for cons in &n.constraints {
                    walk(v, Node::TypeConstraint(cons));
                }
            }
            Node::MonoType(n) => match n {
                MonoType::Tvar(_) => (),
                MonoType::Basic(_) => (),
                MonoType::Array(a) => walk(v, Node::MonoType(&a.element)),
                MonoType::Stream(a) => walk(v, Node::MonoType(&a.element)),
                MonoType::Vector(a) => walk(v, Node::MonoType(&a.element)),
                MonoType::Dict(d) => {
                    walk(v, Node::MonoType(&d.key));
                    walk(v, Node::MonoType(&d.val));
                }
                MonoType::Dynamic(_) => (),
                MonoType::Record(r) => {
                    if let Some(tvar) = &r.tvar {
                        walk(v, Node::Identifier(tvar));
                    }

                    for property in &r.properties {
                        walk(v, Node::PropertyType(property));
                    }
                }
                MonoType::Function(f) => {
                    for param in &f.parameters {
                        walk(v, Node::ParameterType(param));
                    }

                    walk(v, Node::MonoType(&f.monotype));
                }
                MonoType::Label(_) => (),
            },
            Node::PropertyType(n) => {
                walk(v, Node::from_property_key(&n.name));
                walk(v, Node::MonoType(&n.monotype));
            }
            Node::ParameterType(n) => match n {
                ParameterType::Required { name, monotype, .. } => {
                    walk(v, Node::Identifier(name));
                    walk(v, Node::MonoType(monotype));
                }
                ParameterType::Optional {
                    name,
                    monotype,
                    default,
                    ..
                } => {
                    walk(v, Node::Identifier(name));
                    walk(v, Node::MonoType(monotype));
                    if let Some(default) = default {
                        walk(v, Node::LabelLit(default));
                    }
                }
                ParameterType::Pipe { name, monotype, .. } => {
                    if let Some(name) = name {
                        walk(v, Node::Identifier(name));
                    }
                    walk(v, Node::MonoType(monotype));
                }
            },
            Node::TypeConstraint(n) => {
                walk(v, Node::Identifier(&n.tvar));
                for id in &n.kinds {
                    walk(v, Node::Identifier(id));
                }
            }
        }
    }

    v.done(node)
}

impl<'a, F> Visitor<'a> for F
where
    F: FnMut(Node<'a>),
{
    fn visit(&mut self, node: Node<'a>) -> bool {
        self(node);
        true
    }
}
