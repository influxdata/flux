use crate::ast::SourceLocation;
use crate::semantic::nodes::*;
use crate::semantic::types::MonoType;
use std::fmt;
use std::rc::Rc;

/// Node represents any structure that can appear in the semantic graph.
#[derive(Debug)]
pub enum Node<'a> {
    Package(&'a Package),
    File(&'a File),
    PackageClause(&'a PackageClause),
    ImportDeclaration(&'a ImportDeclaration),
    Identifier(&'a Identifier),
    FunctionParameter(&'a FunctionParameter),
    Block(&'a Block),
    Property(&'a Property),

    // Expressions.
    IdentifierExpr(&'a IdentifierExpr),
    ArrayExpr(&'a ArrayExpr),
    FunctionExpr(&'a FunctionExpr),
    LogicalExpr(&'a LogicalExpr),
    ObjectExpr(&'a ObjectExpr),
    MemberExpr(&'a MemberExpr),
    IndexExpr(&'a IndexExpr),
    BinaryExpr(&'a BinaryExpr),
    UnaryExpr(&'a UnaryExpr),
    CallExpr(&'a CallExpr),
    ConditionalExpr(&'a ConditionalExpr),
    StringExpr(&'a StringExpr),
    IntegerLit(&'a IntegerLit),
    FloatLit(&'a FloatLit),
    StringLit(&'a StringLit),
    DurationLit(&'a DurationLit),
    UintLit(&'a UintLit),
    BooleanLit(&'a BooleanLit),
    DateTimeLit(&'a DateTimeLit),
    RegexpLit(&'a RegexpLit),

    // Statements.
    ExprStmt(&'a ExprStmt),
    OptionStmt(&'a OptionStmt),
    ReturnStmt(&'a ReturnStmt),
    TestStmt(&'a TestStmt),
    BuiltinStmt(&'a BuiltinStmt),

    // StringExprPart.
    TextPart(&'a TextPart),
    InterpolatedPart(&'a InterpolatedPart),

    // Assignment.
    VariableAssgn(&'a VariableAssgn),
    MemberAssgn(&'a MemberAssgn),
}

impl<'a> fmt::Display for Node<'a> {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match self {
            Node::Package(_) => write!(f, "Package"),
            Node::File(_) => write!(f, "File"),
            Node::PackageClause(_) => write!(f, "PackageClause"),
            Node::ImportDeclaration(_) => write!(f, "ImportDeclaration"),
            Node::Identifier(_) => write!(f, "Identifier"),
            Node::IdentifierExpr(_) => write!(f, "IdentifierExpr"),
            Node::ArrayExpr(_) => write!(f, "ArrayExpr"),
            Node::FunctionExpr(_) => write!(f, "FunctionExpr"),
            Node::FunctionParameter(_) => write!(f, "FunctionParameter"),
            Node::LogicalExpr(_) => write!(f, "LogicalExpr"),
            Node::ObjectExpr(_) => write!(f, "ObjectExpr"),
            Node::MemberExpr(_) => write!(f, "MemberExpr"),
            Node::IndexExpr(_) => write!(f, "IndexExpr"),
            Node::BinaryExpr(_) => write!(f, "BinaryExpr"),
            Node::UnaryExpr(_) => write!(f, "UnaryExpr"),
            Node::CallExpr(_) => write!(f, "CallExpr"),
            Node::ConditionalExpr(_) => write!(f, "ConditionalExpr"),
            Node::StringExpr(_) => write!(f, "StringExpr"),
            Node::IntegerLit(_) => write!(f, "IntegerLit"),
            Node::FloatLit(_) => write!(f, "FloatLit"),
            Node::StringLit(_) => write!(f, "StringLit"),
            Node::DurationLit(_) => write!(f, "DurationLit"),
            Node::UintLit(_) => write!(f, "UintLit"),
            Node::BooleanLit(_) => write!(f, "BooleanLit"),
            Node::DateTimeLit(_) => write!(f, "DateTimeLit"),
            Node::RegexpLit(_) => write!(f, "RegexpLit"),
            Node::ExprStmt(_) => write!(f, "ExprStmt"),
            Node::OptionStmt(_) => write!(f, "OptionStmt"),
            Node::ReturnStmt(_) => write!(f, "ReturnStmt"),
            Node::TestStmt(_) => write!(f, "TestStmt"),
            Node::BuiltinStmt(_) => write!(f, "BuiltinStmt"),
            Node::Block(n) => match n {
                Block::Variable(_, _) => write!(f, "Block::Variable"),
                Block::Expr(_, _) => write!(f, "Block::Expr"),
                Block::Return(_) => write!(f, "Block::Return"),
            },
            Node::Property(_) => write!(f, "Property"),
            Node::TextPart(_) => write!(f, "TextPart"),
            Node::InterpolatedPart(_) => write!(f, "InterpolatedPart"),
            Node::VariableAssgn(_) => write!(f, "VariableAssgn"),
            Node::MemberAssgn(_) => write!(f, "MemberAssgn"),
        }
    }
}

impl<'a> Node<'a> {
    pub fn loc(&self) -> &SourceLocation {
        match self {
            Node::Package(n) => &n.loc,
            Node::File(n) => &n.loc,
            Node::PackageClause(n) => &n.loc,
            Node::ImportDeclaration(n) => &n.loc,
            Node::Identifier(n) => &n.loc,
            Node::IdentifierExpr(n) => &n.loc,
            Node::ArrayExpr(n) => &n.loc,
            Node::FunctionExpr(n) => &n.loc,
            Node::FunctionParameter(n) => &n.loc,
            Node::LogicalExpr(n) => &n.loc,
            Node::ObjectExpr(n) => &n.loc,
            Node::MemberExpr(n) => &n.loc,
            Node::IndexExpr(n) => &n.loc,
            Node::BinaryExpr(n) => &n.loc,
            Node::UnaryExpr(n) => &n.loc,
            Node::CallExpr(n) => &n.loc,
            Node::ConditionalExpr(n) => &n.loc,
            Node::StringExpr(n) => &n.loc,
            Node::IntegerLit(n) => &n.loc,
            Node::FloatLit(n) => &n.loc,
            Node::StringLit(n) => &n.loc,
            Node::DurationLit(n) => &n.loc,
            Node::UintLit(n) => &n.loc,
            Node::BooleanLit(n) => &n.loc,
            Node::DateTimeLit(n) => &n.loc,
            Node::RegexpLit(n) => &n.loc,
            Node::ExprStmt(n) => &n.loc,
            Node::OptionStmt(n) => &n.loc,
            Node::ReturnStmt(n) => &n.loc,
            Node::TestStmt(n) => &n.loc,
            Node::BuiltinStmt(n) => &n.loc,
            Node::Block(n) => n.loc(),
            Node::Property(n) => &n.loc,
            Node::TextPart(n) => &n.loc,
            Node::InterpolatedPart(n) => &n.loc,
            Node::VariableAssgn(n) => &n.loc,
            Node::MemberAssgn(n) => &n.loc,
        }
    }
    pub fn type_of(&self) -> Option<MonoType> {
        match self {
            Node::IdentifierExpr(n) => Some(Expression::Identifier((*n).clone()).type_of()),
            Node::ArrayExpr(n) => Some(Expression::Array(Box::new((*n).clone())).type_of()),
            Node::FunctionExpr(n) => Some(Expression::Function(Box::new((*n).clone())).type_of()),
            Node::LogicalExpr(n) => Some(Expression::Logical(Box::new((*n).clone())).type_of()),
            Node::ObjectExpr(n) => Some(Expression::Object(Box::new((*n).clone())).type_of()),
            Node::MemberExpr(n) => Some(Expression::Member(Box::new((*n).clone())).type_of()),
            Node::IndexExpr(n) => Some(Expression::Index(Box::new((*n).clone())).type_of()),
            Node::BinaryExpr(n) => Some(Expression::Binary(Box::new((*n).clone())).type_of()),
            Node::UnaryExpr(n) => Some(Expression::Unary(Box::new((*n).clone())).type_of()),
            Node::CallExpr(n) => Some(Expression::Call(Box::new((*n).clone())).type_of()),
            Node::ConditionalExpr(n) => {
                Some(Expression::Conditional(Box::new((*n).clone())).type_of())
            }
            Node::StringExpr(n) => Some(Expression::StringExpr(Box::new((*n).clone())).type_of()),
            Node::IntegerLit(n) => Some(Expression::Integer((*n).clone()).type_of()),
            Node::FloatLit(n) => Some(Expression::Float((*n).clone()).type_of()),
            Node::StringLit(n) => Some(Expression::StringLit((*n).clone()).type_of()),
            Node::DurationLit(n) => Some(Expression::Duration((*n).clone()).type_of()),
            Node::UintLit(n) => Some(Expression::Uint((*n).clone()).type_of()),
            Node::BooleanLit(n) => Some(Expression::Boolean((*n).clone()).type_of()),
            Node::DateTimeLit(n) => Some(Expression::DateTime((*n).clone()).type_of()),
            Node::RegexpLit(n) => Some(Expression::Regexp((*n).clone()).type_of()),
            _ => None,
        }
    }
}

// Utility functions.
impl<'a> Node<'a> {
    pub fn from_expr(expr: &'a Expression) -> Node {
        match *expr {
            Expression::Identifier(ref e) => Node::IdentifierExpr(e),
            Expression::Array(ref e) => Node::ArrayExpr(e),
            Expression::Function(ref e) => Node::FunctionExpr(e),
            Expression::Logical(ref e) => Node::LogicalExpr(e),
            Expression::Object(ref e) => Node::ObjectExpr(e),
            Expression::Member(ref e) => Node::MemberExpr(e),
            Expression::Index(ref e) => Node::IndexExpr(e),
            Expression::Binary(ref e) => Node::BinaryExpr(e),
            Expression::Unary(ref e) => Node::UnaryExpr(e),
            Expression::Call(ref e) => Node::CallExpr(e),
            Expression::Conditional(ref e) => Node::ConditionalExpr(e),
            Expression::StringExpr(ref e) => Node::StringExpr(e),
            Expression::Integer(ref e) => Node::IntegerLit(e),
            Expression::Float(ref e) => Node::FloatLit(e),
            Expression::StringLit(ref e) => Node::StringLit(e),
            Expression::Duration(ref e) => Node::DurationLit(e),
            Expression::Uint(ref e) => Node::UintLit(e),
            Expression::Boolean(ref e) => Node::BooleanLit(e),
            Expression::DateTime(ref e) => Node::DateTimeLit(e),
            Expression::Regexp(ref e) => Node::RegexpLit(e),
        }
    }
    pub fn from_stmt(stmt: &'a Statement) -> Node {
        match *stmt {
            Statement::Expr(ref s) => Node::ExprStmt(s),
            Statement::Variable(ref s) => Node::VariableAssgn(s),
            Statement::Option(ref s) => Node::OptionStmt(s),
            Statement::Return(ref s) => Node::ReturnStmt(s),
            Statement::Test(ref s) => Node::TestStmt(s),
            Statement::Builtin(ref s) => Node::BuiltinStmt(s),
        }
    }
    fn from_string_expr_part(sp: &'a StringExprPart) -> Node {
        match *sp {
            StringExprPart::Text(ref t) => Node::TextPart(t),
            StringExprPart::Interpolated(ref e) => Node::InterpolatedPart(e),
        }
    }
    fn from_assignment(a: &'a Assignment) -> Node {
        match *a {
            Assignment::Variable(ref v) => Node::VariableAssgn(v),
            Assignment::Member(ref m) => Node::MemberAssgn(m),
        }
    }
}

/// Visitor is used by `walk` to recursively visit a semantic graph.
/// One can implement Visitor or use a `FnMut(Node)`.
///
/// # Examples
///
/// Print out the nodes of a semantic graph:
///
/// ```
/// use core::ast;
/// use core::semantic::walk::{Node, walk};
/// use core::semantic::nodes::*;
/// use std::rc::Rc;
///
/// let mut pkg = Package {
///     loc:  ast::BaseNode::default().location,
///     package: "main".to_string(),
///     files: vec![],
/// };
/// walk(
///     &mut |n: Rc<Node>| println!("{}", *n),
///     Rc::new(Node::Package(&pkg)),
/// );
/// ```
///
/// A "scoped" Visitor that errors if finds more than one addition operation in the same scope:
///
/// ```
/// use core::ast::Operator::AdditionOperator;
/// use core::semantic::walk::{Node, Visitor};
/// use core::semantic::nodes::*;
/// use std::rc::Rc;
///
/// struct RepeatedPlusChecker {
///     // A stack.
///     plus: Vec<bool>,
///     err: String,
/// }
///
/// impl RepeatedPlusChecker {
///     fn new() -> RepeatedPlusChecker {
///         RepeatedPlusChecker {
///             plus: vec![false],
///             err: "".to_string(),
///         }
///     }
/// }
///
/// impl <'a> Visitor<'a> for RepeatedPlusChecker {
///     fn visit(&mut self, node: Rc<Node<'a>>) -> bool {
///         match *node {
///             Node::Block(_) => {
///                 self.plus.push(false);
///             }
///             Node::BinaryExpr(ref expr) => {
///                 if expr.operator == AdditionOperator {
///                     if *self.plus.last().expect("there must be a last") {
///                         self.err = format!("repeated + on line {}", expr.loc.start.line);
///                         return false;
///                     }
///                     self.plus.pop();
///                     self.plus.push(true);
///                 }
///             }
///             _ => (),
///         }
///         true
///     }
///
///     fn done(&mut self, node: Rc<Node>) {
///         if let Node::Block(_) = *node {
///             self.plus.pop();
///         }
///     }
/// }
/// ```
pub trait Visitor<'a>: Sized {
    /// Visit is called for a node.
    /// When the Visitor is used in function `walk`, the boolean value returned
    /// is used to continue (true) or stop (false) walking.
    fn visit(&mut self, node: Rc<Node<'a>>) -> bool;
    /// Done is called for a node once it has been visited along with all of its children.
    /// The default is to do nothing.
    fn done(&mut self, _: Rc<Node<'a>>) {}
}

/// `walk` recursively visits children of a node given a Visitor.
/// Nodes are visited in depth-first order.
pub fn walk<'a, T>(v: &mut T, node: Rc<Node<'a>>)
where
    T: Visitor<'a>,
{
    if v.visit(node.clone()) {
        match *node.clone() {
            Node::Package(ref n) => {
                for file in n.files.iter() {
                    walk(v, Rc::new(Node::File(file)));
                }
            }
            Node::File(ref n) => {
                if let Some(ref pkg) = n.package {
                    walk(v, Rc::new(Node::PackageClause(pkg)));
                }
                for imp in n.imports.iter() {
                    walk(v, Rc::new(Node::ImportDeclaration(imp)));
                }
                for stmt in n.body.iter() {
                    walk(v, Rc::new(Node::from_stmt(stmt)));
                }
            }
            Node::PackageClause(ref n) => {
                walk(v, Rc::new(Node::Identifier(&n.name)));
            }
            Node::ImportDeclaration(ref n) => {
                if let Some(ref alias) = n.alias {
                    walk(v, Rc::new(Node::Identifier(alias)));
                }
                walk(v, Rc::new(Node::StringLit(&n.path)));
            }
            Node::Identifier(_) => {}
            Node::IdentifierExpr(_) => {}
            Node::ArrayExpr(ref n) => {
                for element in n.elements.iter() {
                    walk(v, Rc::new(Node::from_expr(element)));
                }
            }
            Node::FunctionExpr(ref n) => {
                for param in n.params.iter() {
                    walk(v, Rc::new(Node::FunctionParameter(param)));
                }
                walk(v, Rc::new(Node::Block(&n.body)));
            }
            Node::FunctionParameter(ref n) => {
                walk(v, Rc::new(Node::Identifier(&n.key)));
                if let Some(ref def) = n.default {
                    walk(v, Rc::new(Node::from_expr(def)));
                }
            }
            Node::LogicalExpr(ref n) => {
                walk(v, Rc::new(Node::from_expr(&n.left)));
                walk(v, Rc::new(Node::from_expr(&n.right)));
            }
            Node::ObjectExpr(ref n) => {
                if let Some(ref i) = n.with {
                    walk(v, Rc::new(Node::IdentifierExpr(i)));
                }
                for prop in n.properties.iter() {
                    walk(v, Rc::new(Node::Property(prop)));
                }
            }
            Node::MemberExpr(ref n) => {
                walk(v, Rc::new(Node::from_expr(&n.object)));
            }
            Node::IndexExpr(ref n) => {
                walk(v, Rc::new(Node::from_expr(&n.array)));
                walk(v, Rc::new(Node::from_expr(&n.index)));
            }
            Node::BinaryExpr(ref n) => {
                walk(v, Rc::new(Node::from_expr(&n.left)));
                walk(v, Rc::new(Node::from_expr(&n.right)));
            }
            Node::UnaryExpr(ref n) => {
                walk(v, Rc::new(Node::from_expr(&n.argument)));
            }
            Node::CallExpr(ref n) => {
                walk(v, Rc::new(Node::from_expr(&n.callee)));
                if let Some(ref p) = n.pipe {
                    walk(v, Rc::new(Node::from_expr(p)));
                }
                for arg in n.arguments.iter() {
                    walk(v, Rc::new(Node::Property(arg)));
                }
            }
            Node::ConditionalExpr(ref n) => {
                walk(v, Rc::new(Node::from_expr(&n.test)));
                walk(v, Rc::new(Node::from_expr(&n.consequent)));
                walk(v, Rc::new(Node::from_expr(&n.alternate)));
            }
            Node::StringExpr(ref n) => {
                for part in n.parts.iter() {
                    walk(v, Rc::new(Node::from_string_expr_part(part)));
                }
            }
            Node::IntegerLit(_) => {}
            Node::FloatLit(_) => {}
            Node::StringLit(_) => {}
            Node::DurationLit(_) => {}
            Node::UintLit(_) => {}
            Node::BooleanLit(_) => {}
            Node::DateTimeLit(_) => {}
            Node::RegexpLit(_) => {}
            Node::ExprStmt(ref n) => {
                walk(v, Rc::new(Node::from_expr(&n.expression)));
            }
            Node::OptionStmt(ref n) => {
                walk(v, Rc::new(Node::from_assignment(&n.assignment)));
            }
            Node::ReturnStmt(ref n) => {
                walk(v, Rc::new(Node::from_expr(&n.argument)));
            }
            Node::TestStmt(ref n) => {
                walk(v, Rc::new(Node::VariableAssgn(&n.assignment)));
            }
            Node::BuiltinStmt(ref n) => {
                walk(v, Rc::new(Node::Identifier(&n.id)));
            }
            Node::Block(ref n) => match n {
                Block::Variable(ref assgn, ref next) => {
                    walk(v, Rc::new(Node::VariableAssgn(assgn)));
                    walk(v, Rc::new(Node::Block(&*next)));
                }
                Block::Expr(ref estmt, ref next) => {
                    walk(v, Rc::new(Node::ExprStmt(estmt)));
                    walk(v, Rc::new(Node::Block(&*next)))
                }
                Block::Return(ref ret_stmt) => walk(v, Rc::new(Node::ReturnStmt(ret_stmt))),
            },
            Node::Property(ref n) => {
                walk(v, Rc::new(Node::Identifier(&n.key)));
                walk(v, Rc::new(Node::from_expr(&n.value)));
            }
            Node::TextPart(_) => {}
            Node::InterpolatedPart(ref n) => {
                walk(v, Rc::new(Node::from_expr(&n.expression)));
            }
            Node::VariableAssgn(ref n) => {
                walk(v, Rc::new(Node::Identifier(&n.id)));
                walk(v, Rc::new(Node::from_expr(&n.init)));
            }
            Node::MemberAssgn(ref n) => {
                walk(v, Rc::new(Node::MemberExpr(&n.member)));
                walk(v, Rc::new(Node::from_expr(&n.init)));
            }
        };
    }
    v.done(node.clone());
}

/// Implementation of Visitor for a mutable closure.
/// We need Higher-Rank Trait Bounds (`for<'a> ...`) here for compiling.
/// See https://doc.rust-lang.org/nomicon/hrtb.html.
impl<'a, F> Visitor<'a> for F
where
    F: FnMut(Rc<Node<'a>>),
{
    fn visit(&mut self, node: Rc<Node<'a>>) -> bool {
        self(node);
        true
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::semantic::walk::test_utils::compile;

    mod node_ids {
        use super::*;

        fn test_walk(source: &str, want: Vec<&str>) {
            let sem_pkg = compile(source);
            let mut nodes = Vec::new();
            walk(
                &mut |n: Rc<Node>| nodes.push(format!("{}", *n)),
                Rc::new(Node::File(&sem_pkg.files[0])),
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

    mod nesting {
        use super::*;
        use crate::ast::Operator::AdditionOperator;

        // NestingCounter counts the number of nested Blocks found while walking.
        struct NestingCounter {
            count: u8,
        }

        impl<'a> Visitor<'a> for NestingCounter {
            fn visit(&mut self, node: Rc<Node<'a>>) -> bool {
                match *node {
                    Node::Block(_) => self.count += 1,
                    _ => (),
                }
                true
            }
        }

        #[test]
        fn test_nesting_count() {
            let pkg = compile(
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
            walk(&mut v, Rc::new(Node::Package(&pkg)));
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

        impl<'a> Visitor<'a> for RepeatedPlusChecker {
            fn visit(&mut self, node: Rc<Node<'a>>) -> bool {
                match *node {
                    Node::FunctionExpr(_) => {
                        self.plus.push(false);
                    }
                    Node::BinaryExpr(ref expr) => {
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

            fn done(&mut self, node: Rc<Node>) {
                if let Node::FunctionExpr(_) = *node {
                    self.plus.pop();
                }
            }
        }

        #[test]
        fn test_nesting_scope() {
            let ok = compile(
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
            walk(&mut v, Rc::new(Node::Package(&ok)));
            assert_eq!(v.err.as_str(), "");
            let not_ok1 = compile(
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
            walk(&mut v, Rc::new(Node::Package(&not_ok1)));
            assert_eq!(v.err.as_str(), "repeated + on line 11");
            let not_ok2 = compile(
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
            walk(&mut v, Rc::new(Node::Package(&not_ok2)));
            assert_eq!(v.err.as_str(), "repeated + on line 7");
            let not_ok3 = compile(
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
            walk(&mut v, Rc::new(Node::Package(&not_ok3)));
            assert_eq!(v.err.as_str(), "repeated + on line 6");
        }
    }
}
