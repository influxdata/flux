use crate::ast::SourceLocation;
use crate::semantic::nodes::*;
use crate::semantic::types::MonoType;
use std::fmt;

/// NodeMut represents any structure that can appear in the semantic graph.
/// It also enables mutability of the wrapped semantic node.
#[derive(Debug)]
pub enum NodeMut<'a> {
    Package(&'a mut Package),
    File(&'a mut File),
    PackageClause(&'a mut PackageClause),
    ImportDeclaration(&'a mut ImportDeclaration),
    Identifier(&'a mut Identifier),
    FunctionParameter(&'a mut FunctionParameter),
    Block(&'a mut Block),
    Property(&'a mut Property),

    // Expressions.
    IdentifierExpr(&'a mut IdentifierExpr),
    ArrayExpr(&'a mut ArrayExpr),
    FunctionExpr(&'a mut FunctionExpr),
    LogicalExpr(&'a mut LogicalExpr),
    ObjectExpr(&'a mut ObjectExpr),
    MemberExpr(&'a mut MemberExpr),
    IndexExpr(&'a mut IndexExpr),
    BinaryExpr(&'a mut BinaryExpr),
    UnaryExpr(&'a mut UnaryExpr),
    CallExpr(&'a mut CallExpr),
    ConditionalExpr(&'a mut ConditionalExpr),
    StringExpr(&'a mut StringExpr),
    IntegerLit(&'a mut IntegerLit),
    FloatLit(&'a mut FloatLit),
    StringLit(&'a mut StringLit),
    DurationLit(&'a mut DurationLit),
    UintLit(&'a mut UintLit),
    BooleanLit(&'a mut BooleanLit),
    DateTimeLit(&'a mut DateTimeLit),
    RegexpLit(&'a mut RegexpLit),

    // Statements.
    ExprStmt(&'a mut ExprStmt),
    OptionStmt(&'a mut OptionStmt),
    ReturnStmt(&'a mut ReturnStmt),
    TestStmt(&'a mut TestStmt),
    BuiltinStmt(&'a mut BuiltinStmt),

    // StringExprPart.
    TextPart(&'a mut TextPart),
    InterpolatedPart(&'a mut InterpolatedPart),

    // Assignment.
    VariableAssgn(&'a mut VariableAssgn), // Native variable assignment
    MemberAssgn(&'a mut MemberAssgn),
}

impl<'a> fmt::Display for NodeMut<'a> {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match self {
            NodeMut::Package(_) => write!(f, "Package"),
            NodeMut::File(_) => write!(f, "File"),
            NodeMut::PackageClause(_) => write!(f, "PackageClause"),
            NodeMut::ImportDeclaration(_) => write!(f, "ImportDeclaration"),
            NodeMut::Identifier(_) => write!(f, "Identifier"),
            NodeMut::IdentifierExpr(_) => write!(f, "IdentifierExpr"),
            NodeMut::ArrayExpr(_) => write!(f, "ArrayExpr"),
            NodeMut::FunctionExpr(_) => write!(f, "FunctionExpr"),
            NodeMut::FunctionParameter(_) => write!(f, "FunctionParameter"),
            NodeMut::LogicalExpr(_) => write!(f, "LogicalExpr"),
            NodeMut::ObjectExpr(_) => write!(f, "ObjectExpr"),
            NodeMut::MemberExpr(_) => write!(f, "MemberExpr"),
            NodeMut::IndexExpr(_) => write!(f, "IndexExpr"),
            NodeMut::BinaryExpr(_) => write!(f, "BinaryExpr"),
            NodeMut::UnaryExpr(_) => write!(f, "UnaryExpr"),
            NodeMut::CallExpr(_) => write!(f, "CallExpr"),
            NodeMut::ConditionalExpr(_) => write!(f, "ConditionalExpr"),
            NodeMut::StringExpr(_) => write!(f, "StringExpr"),
            NodeMut::IntegerLit(_) => write!(f, "IntegerLit"),
            NodeMut::FloatLit(_) => write!(f, "FloatLit"),
            NodeMut::StringLit(_) => write!(f, "StringLit"),
            NodeMut::DurationLit(_) => write!(f, "DurationLit"),
            NodeMut::UintLit(_) => write!(f, "UintLit"),
            NodeMut::BooleanLit(_) => write!(f, "BooleanLit"),
            NodeMut::DateTimeLit(_) => write!(f, "DateTimeLit"),
            NodeMut::RegexpLit(_) => write!(f, "RegexpLit"),
            NodeMut::ExprStmt(_) => write!(f, "ExprStmt"),
            NodeMut::OptionStmt(_) => write!(f, "OptionStmt"),
            NodeMut::ReturnStmt(_) => write!(f, "ReturnStmt"),
            NodeMut::TestStmt(_) => write!(f, "TestStmt"),
            NodeMut::BuiltinStmt(_) => write!(f, "BuiltinStmt"),
            NodeMut::Block(n) => match n {
                Block::Variable(_, _) => write!(f, "Block::Variable"),
                Block::Expr(_, _) => write!(f, "Block::Expr"),
                Block::Return(_) => write!(f, "Block::Return"),
            },
            NodeMut::Property(_) => write!(f, "Property"),
            NodeMut::TextPart(_) => write!(f, "TextPart"),
            NodeMut::InterpolatedPart(_) => write!(f, "InterpolatedPart"),
            NodeMut::VariableAssgn(_) => write!(f, "VariableAssgn"),
            NodeMut::MemberAssgn(_) => write!(f, "MemberAssgn"),
        }
    }
}
impl<'a> NodeMut<'a> {
    pub fn loc(&self) -> &SourceLocation {
        match self {
            NodeMut::Package(n) => &n.loc,
            NodeMut::File(n) => &n.loc,
            NodeMut::PackageClause(n) => &n.loc,
            NodeMut::ImportDeclaration(n) => &n.loc,
            NodeMut::Identifier(n) => &n.loc,
            NodeMut::IdentifierExpr(n) => &n.loc,
            NodeMut::ArrayExpr(n) => &n.loc,
            NodeMut::FunctionExpr(n) => &n.loc,
            NodeMut::FunctionParameter(n) => &n.loc,
            NodeMut::LogicalExpr(n) => &n.loc,
            NodeMut::ObjectExpr(n) => &n.loc,
            NodeMut::MemberExpr(n) => &n.loc,
            NodeMut::IndexExpr(n) => &n.loc,
            NodeMut::BinaryExpr(n) => &n.loc,
            NodeMut::UnaryExpr(n) => &n.loc,
            NodeMut::CallExpr(n) => &n.loc,
            NodeMut::ConditionalExpr(n) => &n.loc,
            NodeMut::StringExpr(n) => &n.loc,
            NodeMut::IntegerLit(n) => &n.loc,
            NodeMut::FloatLit(n) => &n.loc,
            NodeMut::StringLit(n) => &n.loc,
            NodeMut::DurationLit(n) => &n.loc,
            NodeMut::UintLit(n) => &n.loc,
            NodeMut::BooleanLit(n) => &n.loc,
            NodeMut::DateTimeLit(n) => &n.loc,
            NodeMut::RegexpLit(n) => &n.loc,
            NodeMut::ExprStmt(n) => &n.loc,
            NodeMut::OptionStmt(n) => &n.loc,
            NodeMut::ReturnStmt(n) => &n.loc,
            NodeMut::TestStmt(n) => &n.loc,
            NodeMut::BuiltinStmt(n) => &n.loc,
            NodeMut::Block(n) => n.loc(),
            NodeMut::Property(n) => &n.loc,
            NodeMut::TextPart(n) => &n.loc,
            NodeMut::InterpolatedPart(n) => &n.loc,
            NodeMut::VariableAssgn(n) => &n.loc,
            NodeMut::MemberAssgn(n) => &n.loc,
        }
    }
    pub fn type_of(&self) -> Option<&MonoType> {
        match self {
            NodeMut::IdentifierExpr(n) => Some(&n.typ),
            NodeMut::ArrayExpr(n) => Some(&n.typ),
            NodeMut::FunctionExpr(n) => Some(&n.typ),
            NodeMut::LogicalExpr(n) => Some(&n.typ),
            NodeMut::ObjectExpr(n) => Some(&n.typ),
            NodeMut::MemberExpr(n) => Some(&n.typ),
            NodeMut::IndexExpr(n) => Some(&n.typ),
            NodeMut::BinaryExpr(n) => Some(&n.typ),
            NodeMut::UnaryExpr(n) => Some(&n.typ),
            NodeMut::CallExpr(n) => Some(&n.typ),
            NodeMut::ConditionalExpr(n) => Some(&n.typ),
            NodeMut::StringExpr(n) => Some(&n.typ),
            NodeMut::IntegerLit(n) => Some(&n.typ),
            NodeMut::FloatLit(n) => Some(&n.typ),
            NodeMut::StringLit(n) => Some(&n.typ),
            NodeMut::DurationLit(n) => Some(&n.typ),
            NodeMut::UintLit(n) => Some(&n.typ),
            NodeMut::BooleanLit(n) => Some(&n.typ),
            NodeMut::DateTimeLit(n) => Some(&n.typ),
            NodeMut::RegexpLit(n) => Some(&n.typ),
            _ => None,
        }
    }
    pub fn set_loc(&mut self, loc: SourceLocation) {
        match self {
            NodeMut::Package(ref mut n) => n.loc = loc,
            NodeMut::File(ref mut n) => n.loc = loc,
            NodeMut::PackageClause(ref mut n) => n.loc = loc,
            NodeMut::ImportDeclaration(ref mut n) => n.loc = loc,
            NodeMut::Identifier(ref mut n) => n.loc = loc,
            NodeMut::IdentifierExpr(ref mut n) => n.loc = loc,
            NodeMut::ArrayExpr(ref mut n) => n.loc = loc,
            NodeMut::FunctionExpr(ref mut n) => n.loc = loc,
            NodeMut::FunctionParameter(ref mut n) => n.loc = loc,
            NodeMut::LogicalExpr(ref mut n) => n.loc = loc,
            NodeMut::ObjectExpr(ref mut n) => n.loc = loc,
            NodeMut::MemberExpr(ref mut n) => n.loc = loc,
            NodeMut::IndexExpr(ref mut n) => n.loc = loc,
            NodeMut::BinaryExpr(ref mut n) => n.loc = loc,
            NodeMut::UnaryExpr(ref mut n) => n.loc = loc,
            NodeMut::CallExpr(ref mut n) => n.loc = loc,
            NodeMut::ConditionalExpr(ref mut n) => n.loc = loc,
            NodeMut::StringExpr(ref mut n) => n.loc = loc,
            NodeMut::IntegerLit(ref mut n) => n.loc = loc,
            NodeMut::FloatLit(ref mut n) => n.loc = loc,
            NodeMut::StringLit(ref mut n) => n.loc = loc,
            NodeMut::DurationLit(ref mut n) => n.loc = loc,
            NodeMut::UintLit(ref mut n) => n.loc = loc,
            NodeMut::BooleanLit(ref mut n) => n.loc = loc,
            NodeMut::DateTimeLit(ref mut n) => n.loc = loc,
            NodeMut::RegexpLit(ref mut n) => n.loc = loc,
            NodeMut::ExprStmt(ref mut n) => n.loc = loc,
            NodeMut::OptionStmt(ref mut n) => n.loc = loc,
            NodeMut::ReturnStmt(ref mut n) => n.loc = loc,
            NodeMut::TestStmt(ref mut n) => n.loc = loc,
            NodeMut::BuiltinStmt(ref mut n) => n.loc = loc,
            NodeMut::Block(_) => (),
            NodeMut::Property(ref mut n) => n.loc = loc,
            NodeMut::TextPart(ref mut n) => n.loc = loc,
            NodeMut::InterpolatedPart(ref mut n) => n.loc = loc,
            NodeMut::VariableAssgn(ref mut n) => n.loc = loc,
            NodeMut::MemberAssgn(ref mut n) => n.loc = loc,
        };
    }
}

// Private utility functions.
impl<'a> NodeMut<'a> {
    fn from_expr(expr: &'a mut Expression) -> NodeMut {
        match *expr {
            Expression::Identifier(ref mut e) => NodeMut::IdentifierExpr(e),
            Expression::Array(ref mut e) => NodeMut::ArrayExpr(e),
            Expression::Function(ref mut e) => NodeMut::FunctionExpr(e),
            Expression::Logical(ref mut e) => NodeMut::LogicalExpr(e),
            Expression::Object(ref mut e) => NodeMut::ObjectExpr(e),
            Expression::Member(ref mut e) => NodeMut::MemberExpr(e),
            Expression::Index(ref mut e) => NodeMut::IndexExpr(e),
            Expression::Binary(ref mut e) => NodeMut::BinaryExpr(e),
            Expression::Unary(ref mut e) => NodeMut::UnaryExpr(e),
            Expression::Call(ref mut e) => NodeMut::CallExpr(e),
            Expression::Conditional(ref mut e) => NodeMut::ConditionalExpr(e),
            Expression::StringExpr(ref mut e) => NodeMut::StringExpr(e),
            Expression::Integer(ref mut e) => NodeMut::IntegerLit(e),
            Expression::Float(ref mut e) => NodeMut::FloatLit(e),
            Expression::StringLit(ref mut e) => NodeMut::StringLit(e),
            Expression::Duration(ref mut e) => NodeMut::DurationLit(e),
            Expression::Uint(ref mut e) => NodeMut::UintLit(e),
            Expression::Boolean(ref mut e) => NodeMut::BooleanLit(e),
            Expression::DateTime(ref mut e) => NodeMut::DateTimeLit(e),
            Expression::Regexp(ref mut e) => NodeMut::RegexpLit(e),
        }
    }
    fn from_stmt(stmt: &'a mut Statement) -> NodeMut {
        match *stmt {
            Statement::Expr(ref mut s) => NodeMut::ExprStmt(s),
            Statement::Variable(ref mut s) => NodeMut::VariableAssgn(s),
            Statement::Option(ref mut s) => NodeMut::OptionStmt(s),
            Statement::Return(ref mut s) => NodeMut::ReturnStmt(s),
            Statement::Test(ref mut s) => NodeMut::TestStmt(s),
            Statement::Builtin(ref mut s) => NodeMut::BuiltinStmt(s),
        }
    }
    fn from_string_expr_part(sp: &'a mut StringExprPart) -> NodeMut {
        match *sp {
            StringExprPart::Text(ref mut t) => NodeMut::TextPart(t),
            StringExprPart::Interpolated(ref mut e) => NodeMut::InterpolatedPart(e),
        }
    }
    fn from_assignment(a: &'a mut Assignment) -> NodeMut {
        match *a {
            Assignment::Variable(ref mut v) => NodeMut::VariableAssgn(v),
            Assignment::Member(ref mut m) => NodeMut::MemberAssgn(m),
        }
    }
}

/// VisitorMut is used by `walk_mut` to recursively visit a semantic graph and mutate it.
/// The trait makes it possible to mutate both the Visitor and the Nodes while visiting.
/// One can implement VisitorMut or use a `FnMut(NodeMut)`.
///
/// # Examples
///
/// A Visitor that mutate node types:
///
/// ```
/// use core::semantic::walk::{NodeMut, VisitorMut};
/// use core::semantic::types::*;
///
/// struct TypeMutator {}
///
/// impl VisitorMut for TypeMutator {
///     fn visit(&mut self, node: &mut NodeMut) -> bool {
///         match node {
///             NodeMut::IdentifierExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
///             NodeMut::ArrayExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
///             NodeMut::FunctionExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
///             NodeMut::LogicalExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
///             NodeMut::ObjectExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
///             NodeMut::MemberExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
///             NodeMut::IndexExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
///             NodeMut::BinaryExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
///             NodeMut::UnaryExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
///             NodeMut::CallExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
///             NodeMut::ConditionalExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
///             NodeMut::StringExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
///             NodeMut::IntegerLit(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
///             NodeMut::FloatLit(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
///             NodeMut::StringLit(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
///             NodeMut::DurationLit(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
///             NodeMut::UintLit(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
///             NodeMut::BooleanLit(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
///             NodeMut::DateTimeLit(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
///             NodeMut::RegexpLit(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
///             _ => (),
///         };
///         true
///     }
/// }
///
/// impl TypeMutator {
///     fn new() -> Self {
///         TypeMutator {}
///     }
/// }
/// ```
pub trait VisitorMut: Sized {
    /// Visit is called for a node.
    /// When the VisitorMut is used in function `walk_mut`, the boolean value returned
    /// is used to continue (true) or stop (false) walking.
    fn visit(&mut self, node: &mut NodeMut) -> bool;
    /// Done is called for a node once it has been visited along with all of its children.
    /// The default is to do nothing.
    fn done(&mut self, _: &mut NodeMut) {}
}

/// `walk_mut` recursively visits children of a node given a VisitorMut.
/// Nodes are visited in depth-first order.
pub fn walk_mut<T>(v: &mut T, mut node: &mut NodeMut)
where
    T: VisitorMut,
{
    if v.visit(&mut node) {
        match node {
            NodeMut::Package(ref mut n) => {
                for mut file in n.files.iter_mut() {
                    walk_mut(v, &mut NodeMut::File(&mut file));
                }
            }
            NodeMut::File(ref mut n) => {
                if let Some(mut pkg) = n.package.as_mut() {
                    walk_mut(v, &mut NodeMut::PackageClause(&mut pkg));
                }
                for mut imp in n.imports.iter_mut() {
                    walk_mut(v, &mut NodeMut::ImportDeclaration(&mut imp));
                }
                for mut stmt in n.body.iter_mut() {
                    walk_mut(v, &mut NodeMut::from_stmt(&mut stmt));
                }
            }
            NodeMut::PackageClause(ref mut n) => {
                walk_mut(v, &mut NodeMut::Identifier(&mut n.name));
            }
            NodeMut::ImportDeclaration(ref mut n) => {
                if let Some(mut alias) = n.alias.as_mut() {
                    walk_mut(v, &mut NodeMut::Identifier(&mut alias));
                }
                walk_mut(v, &mut NodeMut::StringLit(&mut n.path));
            }
            NodeMut::Identifier(_) => {}
            NodeMut::IdentifierExpr(_) => {}
            NodeMut::ArrayExpr(ref mut n) => {
                for mut element in n.elements.iter_mut() {
                    walk_mut(v, &mut NodeMut::from_expr(&mut element));
                }
            }
            NodeMut::FunctionExpr(ref mut n) => {
                for mut param in n.params.iter_mut() {
                    walk_mut(v, &mut NodeMut::FunctionParameter(&mut param));
                }
                walk_mut(v, &mut NodeMut::Block(&mut n.body));
            }
            NodeMut::FunctionParameter(ref mut n) => {
                walk_mut(v, &mut NodeMut::Identifier(&mut n.key));
                if let Some(mut def) = n.default.as_mut() {
                    walk_mut(v, &mut NodeMut::from_expr(&mut def));
                }
            }
            NodeMut::LogicalExpr(ref mut n) => {
                walk_mut(v, &mut NodeMut::from_expr(&mut n.left));
                walk_mut(v, &mut NodeMut::from_expr(&mut n.right));
            }
            NodeMut::ObjectExpr(ref mut n) => {
                if let Some(mut i) = n.with.as_mut() {
                    walk_mut(v, &mut NodeMut::IdentifierExpr(&mut i));
                }
                for mut prop in n.properties.iter_mut() {
                    walk_mut(v, &mut NodeMut::Property(&mut prop));
                }
            }
            NodeMut::MemberExpr(ref mut n) => {
                walk_mut(v, &mut NodeMut::from_expr(&mut n.object));
            }
            NodeMut::IndexExpr(ref mut n) => {
                walk_mut(v, &mut NodeMut::from_expr(&mut n.array));
                walk_mut(v, &mut NodeMut::from_expr(&mut n.index));
            }
            NodeMut::BinaryExpr(ref mut n) => {
                walk_mut(v, &mut NodeMut::from_expr(&mut n.left));
                walk_mut(v, &mut NodeMut::from_expr(&mut n.right));
            }
            NodeMut::UnaryExpr(ref mut n) => {
                walk_mut(v, &mut NodeMut::from_expr(&mut n.argument));
            }
            NodeMut::CallExpr(ref mut n) => {
                walk_mut(v, &mut NodeMut::from_expr(&mut n.callee));
                if let Some(mut p) = n.pipe.as_mut() {
                    walk_mut(v, &mut NodeMut::from_expr(&mut p));
                }
                for mut arg in n.arguments.iter_mut() {
                    walk_mut(v, &mut NodeMut::Property(&mut arg));
                }
            }
            NodeMut::ConditionalExpr(ref mut n) => {
                walk_mut(v, &mut NodeMut::from_expr(&mut n.test));
                walk_mut(v, &mut NodeMut::from_expr(&mut n.consequent));
                walk_mut(v, &mut NodeMut::from_expr(&mut n.alternate));
            }
            NodeMut::StringExpr(ref mut n) => {
                for mut part in n.parts.iter_mut() {
                    walk_mut(v, &mut NodeMut::from_string_expr_part(&mut part));
                }
            }
            NodeMut::IntegerLit(_) => {}
            NodeMut::FloatLit(_) => {}
            NodeMut::StringLit(_) => {}
            NodeMut::DurationLit(_) => {}
            NodeMut::UintLit(_) => {}
            NodeMut::BooleanLit(_) => {}
            NodeMut::DateTimeLit(_) => {}
            NodeMut::RegexpLit(_) => {}
            NodeMut::ExprStmt(ref mut n) => {
                walk_mut(v, &mut NodeMut::from_expr(&mut n.expression));
            }
            NodeMut::OptionStmt(ref mut n) => {
                walk_mut(v, &mut NodeMut::from_assignment(&mut n.assignment));
            }
            NodeMut::ReturnStmt(ref mut n) => {
                walk_mut(v, &mut NodeMut::from_expr(&mut n.argument));
            }
            NodeMut::TestStmt(ref mut n) => {
                walk_mut(v, &mut NodeMut::VariableAssgn(&mut n.assignment));
            }
            NodeMut::BuiltinStmt(ref mut n) => {
                walk_mut(v, &mut NodeMut::Identifier(&mut n.id));
            }
            NodeMut::Block(ref mut n) => match n {
                Block::Variable(ref mut assgn, ref mut next) => {
                    walk_mut(v, &mut NodeMut::VariableAssgn(assgn));
                    walk_mut(v, &mut NodeMut::Block(&mut *next));
                }
                Block::Expr(ref mut estmt, ref mut next) => {
                    walk_mut(v, &mut NodeMut::ExprStmt(estmt));
                    walk_mut(v, &mut NodeMut::Block(&mut *next))
                }
                Block::Return(ref mut ret_stmt) => walk_mut(v, &mut NodeMut::ReturnStmt(ret_stmt)),
            },
            NodeMut::Property(ref mut n) => {
                walk_mut(v, &mut NodeMut::Identifier(&mut n.key));
                walk_mut(v, &mut NodeMut::from_expr(&mut n.value));
            }
            NodeMut::TextPart(_) => {}
            NodeMut::InterpolatedPart(ref mut n) => {
                walk_mut(v, &mut NodeMut::from_expr(&mut n.expression));
            }
            NodeMut::VariableAssgn(ref mut n) => {
                walk_mut(v, &mut NodeMut::Identifier(&mut n.id));
                walk_mut(v, &mut NodeMut::from_expr(&mut n.init));
            }
            NodeMut::MemberAssgn(ref mut n) => {
                walk_mut(v, &mut NodeMut::MemberExpr(&mut n.member));
                walk_mut(v, &mut NodeMut::from_expr(&mut n.init));
            }
        };
    }
    v.done(&mut node);
}

/// Implementation of VisitorMut for a mutable closure.
/// We need Higher-Rank Trait Bounds (`for<'a> ...`) here for compiling.
/// See https://doc.rust-lang.org/nomicon/hrtb.html.
impl<F> VisitorMut for F
where
    F: for<'a> FnMut(&mut NodeMut<'a>),
{
    fn visit(&mut self, node: &mut NodeMut) -> bool {
        self(node);
        true
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::ast;
    use crate::semantic::walk::test_utils::compile;

    mod node_ids {
        use super::*;

        fn test_walk(source: &str, want: Vec<&str>) {
            let mut sem_pkg = compile(source);
            let mut nodes = Vec::new();
            walk_mut(
                &mut |n: &mut NodeMut| nodes.push(format!("{}", n)),
                &mut NodeMut::File(&mut sem_pkg.files[0]),
            );
            assert_eq!(want, nodes);
        }

        #[test]
        fn test_file() {
            test_walk("", vec!["File"])
        }
        #[test]
        fn test_package_clause() {
            test_walk("package a", vec!["File", "PackageClause", "Identifier"])
        }
        #[test]
        fn test_import_declaration() {
            test_walk(
                "import \"a\"",
                vec!["File", "ImportDeclaration", "StringLit"],
            )
        }
        #[test]
        fn test_ident() {
            test_walk("a", vec!["File", "ExprStmt", "IdentifierExpr"])
        }
        #[test]
        fn test_array_expr() {
            test_walk(
                "[1,2,3]",
                vec![
                    "File",
                    "ExprStmt",
                    "ArrayExpr",
                    "IntegerLit",
                    "IntegerLit",
                    "IntegerLit",
                ],
            )
        }
        #[test]
        fn test_function_expr() {
            test_walk(
                "() => 1",
                vec![
                    "File",
                    "ExprStmt",
                    "FunctionExpr",
                    "Block::Return",
                    "ReturnStmt",
                    "IntegerLit",
                ],
            )
        }
        #[test]
        fn test_function_expr_multiline_block() {
            test_walk(
                "() => {
                    a = 1
                    b = 3 + 2
                    a + b
                    return a
                }",
                vec![
                    "File",
                    "ExprStmt",
                    "FunctionExpr",
                    "Block::Variable",
                    "VariableAssgn",
                    "Identifier",
                    "IntegerLit",
                    "Block::Variable",
                    "VariableAssgn",
                    "Identifier",
                    "BinaryExpr",
                    "IntegerLit",
                    "IntegerLit",
                    "Block::Expr",
                    "ExprStmt",
                    "BinaryExpr",
                    "IdentifierExpr",
                    "IdentifierExpr",
                    "Block::Return",
                    "ReturnStmt",
                    "IdentifierExpr",
                ],
            )
        }
        #[test]
        fn test_function_with_args() {
            test_walk(
                "(a=1) => a",
                vec![
                    "File",
                    "ExprStmt",
                    "FunctionExpr",
                    "FunctionParameter",
                    "Identifier",
                    "IntegerLit",
                    "Block::Return",
                    "ReturnStmt",
                    "IdentifierExpr",
                ],
            )
        }
        #[test]
        fn test_logical_expr() {
            test_walk(
                "true or false",
                vec![
                    "File",
                    "ExprStmt",
                    "LogicalExpr",
                    "IdentifierExpr",
                    "IdentifierExpr",
                ],
            )
        }
        #[test]
        fn test_object_expr() {
            test_walk(
                "{a:1,\"b\":false}",
                vec![
                    "File",
                    "ExprStmt",
                    "ObjectExpr",
                    "Property",
                    "Identifier",
                    "IntegerLit",
                    "Property",
                    "Identifier",
                    "IdentifierExpr",
                ],
            )
        }
        #[test]
        fn test_member_expr() {
            test_walk(
                "a.b",
                vec!["File", "ExprStmt", "MemberExpr", "IdentifierExpr"],
            )
        }
        #[test]
        fn test_index_expr() {
            test_walk(
                "a[b]",
                vec![
                    "File",
                    "ExprStmt",
                    "IndexExpr",
                    "IdentifierExpr",
                    "IdentifierExpr",
                ],
            )
        }
        #[test]
        fn test_binary_expr() {
            test_walk(
                "a+b",
                vec![
                    "File",
                    "ExprStmt",
                    "BinaryExpr",
                    "IdentifierExpr",
                    "IdentifierExpr",
                ],
            )
        }
        #[test]
        fn test_unary_expr() {
            test_walk(
                "-b",
                vec!["File", "ExprStmt", "UnaryExpr", "IdentifierExpr"],
            )
        }
        #[test]
        fn test_pipe_expr() {
            test_walk(
                "a|>b()",
                vec![
                    "File",
                    "ExprStmt",
                    "CallExpr",
                    "IdentifierExpr",
                    "IdentifierExpr",
                ],
            )
        }
        #[test]
        fn test_call_expr() {
            test_walk(
                "b(a:1)",
                vec![
                    "File",
                    "ExprStmt",
                    "CallExpr",
                    "IdentifierExpr",
                    "Property",
                    "Identifier",
                    "IntegerLit",
                ],
            )
        }
        #[test]
        fn test_conditional_expr() {
            test_walk(
                "if x then y else z",
                vec![
                    "File",
                    "ExprStmt",
                    "ConditionalExpr",
                    "IdentifierExpr",
                    "IdentifierExpr",
                    "IdentifierExpr",
                ],
            )
        }
        #[test]
        fn test_string_expr() {
            test_walk(
                "\"hello ${world}\"",
                vec![
                    "File",
                    "ExprStmt",
                    "StringExpr",
                    "TextPart",
                    "InterpolatedPart",
                    "IdentifierExpr",
                ],
            )
        }
        #[test]
        fn test_paren_expr() {
            test_walk(
                "(a + b)",
                vec![
                    "File",
                    "ExprStmt",
                    "BinaryExpr",
                    "IdentifierExpr",
                    "IdentifierExpr",
                ],
            )
        }
        #[test]
        fn test_integer_lit() {
            test_walk("1", vec!["File", "ExprStmt", "IntegerLit"])
        }
        #[test]
        fn test_float_lit() {
            test_walk("1.0", vec!["File", "ExprStmt", "FloatLit"])
        }
        #[test]
        fn test_string_lit() {
            test_walk("\"a\"", vec!["File", "ExprStmt", "StringLit"])
        }
        #[test]
        fn test_duration_lit() {
            test_walk("1m", vec!["File", "ExprStmt", "DurationLit"])
        }
        #[test]
        fn test_datetime_lit() {
            test_walk(
                "2019-01-01T00:00:00Z",
                vec!["File", "ExprStmt", "DateTimeLit"],
            )
        }
        #[test]
        fn test_regexp_lit() {
            test_walk("/./", vec!["File", "ExprStmt", "RegexpLit"])
        }
        #[test]
        fn test_pipe_lit() {
            test_walk(
                "(a=<-)=>a",
                vec![
                    "File",
                    "ExprStmt",
                    "FunctionExpr",
                    "FunctionParameter",
                    "Identifier",
                    "Block::Return",
                    "ReturnStmt",
                    "IdentifierExpr",
                ],
            )
        }

        #[test]
        fn test_option_stmt() {
            test_walk(
                "option a = b",
                vec![
                    "File",
                    "OptionStmt",
                    "VariableAssgn",
                    "Identifier",
                    "IdentifierExpr",
                ],
            )
        }
        #[test]
        fn test_return_stmt() {
            // This is quite tricky, even if there is an explicit ReturnStmt,
            // `analyze` returns a `Block::Return` when inside of a function body.
            test_walk(
                "() => {return 1}",
                vec![
                    "File",
                    "ExprStmt",
                    "FunctionExpr",
                    "Block::Return",
                    "ReturnStmt",
                    "IntegerLit",
                ],
            )
        }
        #[test]
        fn test_test_stmt() {
            test_walk(
                "test a = 1",
                vec![
                    "File",
                    "TestStmt",
                    "VariableAssgn",
                    "Identifier",
                    "IntegerLit",
                ],
            )
        }
        #[test]
        fn test_builtin_stmt() {
            test_walk("builtin a", vec!["File", "BuiltinStmt", "Identifier"])
        }
        #[test]
        fn test_variable_assgn() {
            test_walk(
                "a = b",
                vec!["File", "VariableAssgn", "Identifier", "IdentifierExpr"],
            )
        }
        #[test]
        fn test_member_assgn() {
            test_walk(
                "option a.b = c",
                vec![
                    "File",
                    "OptionStmt",
                    "MemberAssgn",
                    "MemberExpr",
                    "IdentifierExpr",
                    "IdentifierExpr",
                ],
            )
        }
    }

    mod mutate_nodes {
        use super::*;
        use crate::semantic::types::{MonoType, Tvar};
        use std::collections::HashSet;

        // LocationCollector collects the locations found in the graph while walking.
        struct LocationCollector {
            locs: Vec<ast::SourceLocation>,
        }

        impl VisitorMut for LocationCollector {
            fn visit(&mut self, node: &mut NodeMut) -> bool {
                self.locs.push(node.loc().clone());
                true
            }
        }

        impl LocationCollector {
            fn new() -> LocationCollector {
                LocationCollector { locs: Vec::new() }
            }
        }

        // TypeCollector collects the types found in the graph while walking.
        struct TypeCollector {
            types: Vec<Tvar>,
        }

        impl TypeCollector {
            fn new() -> TypeCollector {
                TypeCollector { types: Vec::new() }
            }
        }

        impl VisitorMut for TypeCollector {
            fn visit(&mut self, node: &mut NodeMut) -> bool {
                let typ = match node {
                    NodeMut::FunctionExpr(ref expr) => Some(expr.typ.clone()),
                    NodeMut::CallExpr(ref expr) => Some(expr.typ.clone()),
                    NodeMut::MemberExpr(ref expr) => Some(expr.typ.clone()),
                    NodeMut::IndexExpr(ref expr) => Some(expr.typ.clone()),
                    NodeMut::BinaryExpr(ref expr) => Some(expr.typ.clone()),
                    NodeMut::UnaryExpr(ref expr) => Some(expr.typ.clone()),
                    NodeMut::LogicalExpr(ref expr) => Some(expr.typ.clone()),
                    NodeMut::ConditionalExpr(ref expr) => Some(expr.typ.clone()),
                    NodeMut::ObjectExpr(ref expr) => Some(expr.typ.clone()),
                    NodeMut::ArrayExpr(ref expr) => Some(expr.typ.clone()),
                    NodeMut::IdentifierExpr(ref expr) => Some(expr.typ.clone()),
                    NodeMut::StringExpr(ref expr) => Some(expr.typ.clone()),
                    NodeMut::StringLit(ref lit) => Some(lit.typ.clone()),
                    NodeMut::BooleanLit(ref lit) => Some(lit.typ.clone()),
                    NodeMut::FloatLit(ref lit) => Some(lit.typ.clone()),
                    NodeMut::IntegerLit(ref lit) => Some(lit.typ.clone()),
                    NodeMut::UintLit(ref lit) => Some(lit.typ.clone()),
                    NodeMut::RegexpLit(ref lit) => Some(lit.typ.clone()),
                    NodeMut::DurationLit(ref lit) => Some(lit.typ.clone()),
                    NodeMut::DateTimeLit(ref lit) => Some(lit.typ.clone()),
                    _ => None,
                };
                if let Some(MonoType::Var(tv)) = typ {
                    self.types.push(tv);
                }
                true
            }
        }

        #[test]
        fn test_loc() {
            let base_loc = ast::BaseNode::default().location;
            let mut pkg = compile(
                r#"
a = from(bucket:"Flux/autogen")
	|> filter(fn: (r) => r["_measurement"] == "a")
	|> range(start:-1h)

b = from(bucket:"Flux/autogen")
	|> filter(fn: (r) => r["_measurement"] == "b")
	|> range(start:-1h)

join(tables:[a,b], on:["t1"], fn: (a,b) => (a["_field"] - b["_field"]) / b["_field"])"#,
            );
            let mut v = LocationCollector::new();
            walk_mut(&mut v, &mut NodeMut::Package(&mut pkg));
            let locs = v.locs;
            assert!(locs.len() > 0);
            for loc in locs {
                assert_ne!(loc, base_loc);
            }
            // now mutate the locations
            walk_mut(
                &mut |n: &mut NodeMut| n.set_loc(base_loc.clone()),
                &mut NodeMut::Package(&mut pkg),
            );
            // now assert that every location is the base one
            let mut v = LocationCollector::new();
            walk_mut(&mut v, &mut NodeMut::Package(&mut pkg));
            let locs = v.locs;
            assert!(locs.len() > 0);
            for loc in locs {
                assert_eq!(loc, base_loc);
            }
        }

        #[test]
        fn test_types() {
            let mut pkg = compile(
                r#"
a = from(bucket:"Flux/autogen")
	|> filter(fn: (r) => r["_measurement"] == "a")
	|> range(start:-1h)

b = from(bucket:"Flux/autogen")
	|> filter(fn: (r) => r["_measurement"] == "b")
	|> range(start:-1h)

join(tables:[a,b], on:["t1"], fn: (a,b) => (a["_field"] - b["_field"]) / b["_field"])"#,
            );
            let mut v = TypeCollector::new();
            walk_mut(&mut v, &mut NodeMut::Package(&mut pkg));
            let types = v.types;
            assert!(types.len() > 0);
            // every tvar has a unique id
            let mut ids: HashSet<u64> = HashSet::new();
            for tvar in types {
                assert!(!ids.contains(&tvar.0));
                ids.insert(tvar.0);
            }
            // now mutate the types
            walk_mut(
                &mut |n: &mut NodeMut| {
                    match n {
                        NodeMut::IdentifierExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        NodeMut::ArrayExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        NodeMut::FunctionExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        NodeMut::LogicalExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        NodeMut::ObjectExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        NodeMut::MemberExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        NodeMut::IndexExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        NodeMut::BinaryExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        NodeMut::UnaryExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        NodeMut::CallExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        NodeMut::ConditionalExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        NodeMut::StringExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        NodeMut::IntegerLit(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        NodeMut::FloatLit(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        NodeMut::StringLit(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        NodeMut::DurationLit(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        NodeMut::UintLit(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        NodeMut::BooleanLit(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        NodeMut::DateTimeLit(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        NodeMut::RegexpLit(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        _ => (),
                    };
                },
                &mut NodeMut::Package(&mut pkg),
            );
            // now assert that every type is the invalid one
            let mut v = TypeCollector::new();
            walk_mut(&mut v, &mut NodeMut::Package(&mut pkg));
            let types = v.types;
            assert!(types.len() > 0);
            for tvar in types {
                assert_eq!(tvar, Tvar(1234));
            }
        }
    }

    mod nesting {
        use super::*;
        use crate::ast::Operator::AdditionOperator;

        // NestingCounter counts the number of nested Blocks found while walking.
        struct NestingCounter {
            count: u8,
        }

        impl VisitorMut for NestingCounter {
            fn visit(&mut self, node: &mut NodeMut) -> bool {
                match node {
                    NodeMut::Block(_) => self.count += 1,
                    _ => (),
                }
                true
            }
        }

        #[test]
        fn test_nesting_count() {
            let mut pkg = compile(
                r#"
f = () => {
    // 1
    return () => {
        // 2
        return () => {
            // 3
            1 + 1
            // 4
            return 1
        }
    }
}

g = () => {
    // 5
    return 2
}

f()()()
g()
"#,
            );
            let mut v = NestingCounter { count: 0 };
            walk_mut(&mut v, &mut NodeMut::Package(&mut pkg));
            assert_eq!(v.count, 5);
        }

        // RepeatedPlusChecker checks if there is more than one addition in the same Function Scope.
        // If so, it errors.
        struct RepeatedPlusChecker {
            // A stack.
            plus: Vec<bool>,
            err: String,
        }

        impl RepeatedPlusChecker {
            fn new() -> RepeatedPlusChecker {
                RepeatedPlusChecker {
                    plus: vec![false],
                    err: "".to_string(),
                }
            }
        }

        impl VisitorMut for RepeatedPlusChecker {
            fn visit(&mut self, node: &mut NodeMut) -> bool {
                match node {
                    NodeMut::FunctionExpr(_) => {
                        self.plus.push(false);
                    }
                    NodeMut::BinaryExpr(ref expr) => {
                        if expr.operator == AdditionOperator {
                            if *self.plus.last().expect("there must be a last") {
                                self.err = format!("repeated + on line {}", expr.loc.start.line);
                                return false;
                            }
                            self.plus.pop();
                            self.plus.push(true);
                        }
                    }
                    _ => (),
                }
                true
            }

            fn done(&mut self, node: &mut NodeMut) {
                if let NodeMut::FunctionExpr(_) = node {
                    self.plus.pop();
                }
            }
        }

        #[test]
        fn test_nesting_scope() {
            let mut ok = compile(
                r#"
1 + 1
f = () => {
    1 + 1
    return () => {
        1 + 1
        return () => {
            return 1 + 1
        }
    }
}
"#,
            );
            let mut v = RepeatedPlusChecker::new();
            walk_mut(&mut v, &mut NodeMut::Package(&mut ok));
            assert_eq!(v.err.as_str(), "");
            let mut not_ok1 = compile(
                r#"
1 + 1
f = () => {
    1 + 1
    return () => {
        return () => {
            return 1
        }
    }
}
1 + 1 // error on line 11
"#,
            );
            let mut v = RepeatedPlusChecker::new();
            walk_mut(&mut v, &mut NodeMut::Package(&mut not_ok1));
            assert_eq!(v.err.as_str(), "repeated + on line 11");
            let mut not_ok2 = compile(
                r#"
1 + 1
f = () => {
    1 + 1
    return () => {
        1 + 1
        return 1 + () => { // error on line 7
            return 1
        }()
    }
}
"#,
            );
            let mut v = RepeatedPlusChecker::new();
            walk_mut(&mut v, &mut NodeMut::Package(&mut not_ok2));
            assert_eq!(v.err.as_str(), "repeated + on line 7");
            let mut not_ok3 = compile(
                r#"
f = () => {
    1 + 1
    return () => {
        1 + 1
        return 1 + () => { // error on line 6
            return 1 + 1 + 1 // error but should be ignored
        }()
    }
}
"#,
            );
            let mut v = RepeatedPlusChecker::new();
            walk_mut(&mut v, &mut NodeMut::Package(&mut not_ok3));
            assert_eq!(v.err.as_str(), "repeated + on line 6");
        }
    }
}
