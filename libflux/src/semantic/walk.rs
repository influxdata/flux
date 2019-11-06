use crate::ast::SourceLocation;
use crate::semantic::nodes::*;
use std::cell::RefCell;
use std::fmt;
use std::rc::Rc;

/// Node represents any structure that can appear in the AST.
#[derive(Debug)]
pub enum Node<'a> {
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
    VariableAssgn(&'a mut VariableAssgn),
    MemberAssgn(&'a mut MemberAssgn),
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
                Block::Return(expr) => write!(f, "Block::Return"),
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
}

// Private utility functions for node conversion.
impl<'a> Node<'a> {
    fn from_expr(expr: &'a mut Expression) -> Node {
        match *expr {
            Expression::Identifier(ref mut e) => Node::IdentifierExpr(e),
            Expression::Array(ref mut e) => Node::ArrayExpr(e),
            Expression::Function(ref mut e) => Node::FunctionExpr(e),
            Expression::Logical(ref mut e) => Node::LogicalExpr(e),
            Expression::Object(ref mut e) => Node::ObjectExpr(e),
            Expression::Member(ref mut e) => Node::MemberExpr(e),
            Expression::Index(ref mut e) => Node::IndexExpr(e),
            Expression::Binary(ref mut e) => Node::BinaryExpr(e),
            Expression::Unary(ref mut e) => Node::UnaryExpr(e),
            Expression::Call(ref mut e) => Node::CallExpr(e),
            Expression::Conditional(ref mut e) => Node::ConditionalExpr(e),
            Expression::StringExpr(ref mut e) => Node::StringExpr(e),
            Expression::Integer(ref mut e) => Node::IntegerLit(e),
            Expression::Float(ref mut e) => Node::FloatLit(e),
            Expression::StringLit(ref mut e) => Node::StringLit(e),
            Expression::Duration(ref mut e) => Node::DurationLit(e),
            Expression::Uint(ref mut e) => Node::UintLit(e),
            Expression::Boolean(ref mut e) => Node::BooleanLit(e),
            Expression::DateTime(ref mut e) => Node::DateTimeLit(e),
            Expression::Regexp(ref mut e) => Node::RegexpLit(e),
        }
    }
    fn from_stmt(stmt: &'a mut Statement) -> Node {
        match *stmt {
            Statement::Expr(ref mut s) => Node::ExprStmt(s),
            Statement::Variable(ref mut s) => Node::VariableAssgn(s),
            Statement::Option(ref mut s) => Node::OptionStmt(s),
            Statement::Return(ref mut s) => Node::ReturnStmt(s),
            Statement::Test(ref mut s) => Node::TestStmt(s),
            Statement::Builtin(ref mut s) => Node::BuiltinStmt(s),
        }
    }
    fn from_string_expr_part(sp: &'a mut StringExprPart) -> Node {
        match *sp {
            StringExprPart::Text(ref mut t) => Node::TextPart(t),
            StringExprPart::Interpolated(ref mut e) => Node::InterpolatedPart(e),
        }
    }
    fn from_assignment(a: &'a mut Assignment) -> Node {
        match *a {
            Assignment::Variable(ref mut v) => Node::VariableAssgn(v),
            Assignment::Member(ref mut m) => Node::MemberAssgn(m),
        }
    }
}

/// Visitor is used by `walk` to recursively visit a semantic graph.
/// The trait makes it possible to mutate both the Visitor and the Nodes while visiting.
/// One can implement Visitor or use a `FnMut(Rc<RefCell<Node>>)`.
///
/// # Examples
///
/// Print out the nodes of a semantic graph:
///
/// ```
/// use flux::ast;
/// use flux::semantic::walk::{Node, walk};
/// use flux::semantic::nodes::*;
/// use std::rc::Rc;
/// use std::cell::RefCell;
///
/// let mut pkg = Package {
///     loc:  ast::BaseNode::default().location,
///     package: "main".to_string(),
///     files: vec![],
/// };
/// walk(
///     &mut |n: Rc<RefCell<Node>>| println!("{}", *n.borrow()),
///     Rc::new(RefCell::new(Node::Package(&mut pkg))),
/// );
/// ```
///
/// A Visitor that collects the locations of nodes:
///
/// ```
/// use flux::ast;
/// use flux::semantic::walk::{Node, Visitor};
/// use flux::semantic::nodes::*;
/// use std::rc::Rc;
/// use std::cell::RefCell;
///
/// struct LocationCollector {
///     locs: Vec<ast::SourceLocation>,
/// }
///
/// impl Visitor for LocationCollector {
///     fn visit(&mut self, node: Rc<RefCell<Node>>) -> bool {
///         let node = node.borrow();
///         self.locs.push(node.loc().clone());
///         true
///     }
/// }
///
/// impl LocationCollector {
///     fn new() -> LocationCollector {
///         LocationCollector { locs: Vec::new() }
///     }
/// }
/// ```
///
/// A "scoped" Visitor that errors if finds more than one addition operation in the same scope:
///
/// ```
/// use flux::ast::Operator::AdditionOperator;
/// use flux::semantic::walk::{Node, Visitor};
/// use flux::semantic::nodes::*;
/// use std::rc::Rc;
/// use std::cell::RefCell;
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
/// impl Visitor for RepeatedPlusChecker {
///     fn visit(&mut self, node: Rc<RefCell<Node>>) -> bool {
///         match *node.borrow() {
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
///     fn done(&mut self, node: Rc<RefCell<Node>>) {
///         let node = node.borrow();
///         if let Node::Block(_) = *node {
///             self.plus.pop();
///         }
///     }
/// }
/// ```
pub trait Visitor: Sized {
    /// Visit is called for a node.
    /// When the Visitor is used in function `walk`, the boolean value returned
    /// is used to continue (true) or stop (false) walking.
    fn visit(&mut self, node: Rc<RefCell<Node>>) -> bool;
    /// Done is called for a node once it has been visited along with all of its children.
    /// The default is to do nothing
    fn done(&mut self, _: Rc<RefCell<Node>>) {}
}

/// Walk recursively visits children of a node given a Visitor.
/// Nodes are visited in depth-first order.
pub fn walk<T>(v: &mut T, node: Rc<RefCell<Node>>)
where
    T: Visitor,
{
    if v.visit(Rc::clone(&node)) {
        match *node.borrow_mut() {
            Node::Package(ref mut n) => {
                for mut file in n.files.iter_mut() {
                    walk(v, Rc::new(RefCell::new(Node::File(&mut file))));
                }
            }
            Node::File(ref mut n) => {
                if let Some(mut pkg) = n.package.as_mut() {
                    walk(v, Rc::new(RefCell::new(Node::PackageClause(&mut pkg))));
                }
                for mut imp in n.imports.iter_mut() {
                    walk(v, Rc::new(RefCell::new(Node::ImportDeclaration(&mut imp))));
                }
                for mut stmt in n.body.iter_mut() {
                    walk(v, Rc::new(RefCell::new(Node::from_stmt(&mut stmt))));
                }
            }
            Node::PackageClause(ref mut n) => {
                walk(v, Rc::new(RefCell::new(Node::Identifier(&mut n.name))));
            }
            Node::ImportDeclaration(ref mut n) => {
                if let Some(mut alias) = n.alias.as_mut() {
                    walk(v, Rc::new(RefCell::new(Node::Identifier(&mut alias))));
                }
                walk(v, Rc::new(RefCell::new(Node::StringLit(&mut n.path))));
            }
            Node::Identifier(_) => {}
            Node::IdentifierExpr(_) => {}
            Node::ArrayExpr(ref mut n) => {
                for mut element in n.elements.iter_mut() {
                    walk(v, Rc::new(RefCell::new(Node::from_expr(&mut element))));
                }
            }
            Node::FunctionExpr(ref mut n) => {
                for mut param in n.params.iter_mut() {
                    walk(
                        v,
                        Rc::new(RefCell::new(Node::FunctionParameter(&mut param))),
                    );
                }
                walk(v, Rc::new(RefCell::new(Node::Block(&mut n.body))));
            }
            Node::FunctionParameter(ref mut n) => {
                walk(v, Rc::new(RefCell::new(Node::Identifier(&mut n.key))));
                if let Some(mut def) = n.default.as_mut() {
                    walk(v, Rc::new(RefCell::new(Node::from_expr(&mut def))));
                }
            }
            Node::LogicalExpr(ref mut n) => {
                walk(v, Rc::new(RefCell::new(Node::from_expr(&mut n.left))));
                walk(v, Rc::new(RefCell::new(Node::from_expr(&mut n.right))));
            }
            Node::ObjectExpr(ref mut n) => {
                if let Some(mut i) = n.with.as_mut() {
                    walk(v, Rc::new(RefCell::new(Node::IdentifierExpr(&mut i))));
                }
                for mut prop in n.properties.iter_mut() {
                    walk(v, Rc::new(RefCell::new(Node::Property(&mut prop))));
                }
            }
            Node::MemberExpr(ref mut n) => {
                walk(v, Rc::new(RefCell::new(Node::from_expr(&mut n.object))));
            }
            Node::IndexExpr(ref mut n) => {
                walk(v, Rc::new(RefCell::new(Node::from_expr(&mut n.array))));
                walk(v, Rc::new(RefCell::new(Node::from_expr(&mut n.index))));
            }
            Node::BinaryExpr(ref mut n) => {
                walk(v, Rc::new(RefCell::new(Node::from_expr(&mut n.left))));
                walk(v, Rc::new(RefCell::new(Node::from_expr(&mut n.right))));
            }
            Node::UnaryExpr(ref mut n) => {
                walk(v, Rc::new(RefCell::new(Node::from_expr(&mut n.argument))));
            }
            Node::CallExpr(ref mut n) => {
                walk(v, Rc::new(RefCell::new(Node::from_expr(&mut n.callee))));
                if let Some(mut p) = n.pipe.as_mut() {
                    walk(v, Rc::new(RefCell::new(Node::from_expr(&mut p))));
                }
                for mut arg in n.arguments.iter_mut() {
                    walk(v, Rc::new(RefCell::new(Node::Property(&mut arg))));
                }
            }
            Node::ConditionalExpr(ref mut n) => {
                walk(v, Rc::new(RefCell::new(Node::from_expr(&mut n.test))));
                walk(v, Rc::new(RefCell::new(Node::from_expr(&mut n.consequent))));
                walk(v, Rc::new(RefCell::new(Node::from_expr(&mut n.alternate))));
            }
            Node::StringExpr(ref mut n) => {
                for mut part in n.parts.iter_mut() {
                    walk(
                        v,
                        Rc::new(RefCell::new(Node::from_string_expr_part(&mut part))),
                    );
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
            Node::ExprStmt(ref mut n) => {
                walk(v, Rc::new(RefCell::new(Node::from_expr(&mut n.expression))));
            }
            Node::OptionStmt(ref mut n) => {
                walk(
                    v,
                    Rc::new(RefCell::new(Node::from_assignment(&mut n.assignment))),
                );
            }
            Node::ReturnStmt(ref mut n) => {
                walk(v, Rc::new(RefCell::new(Node::from_expr(&mut n.argument))));
            }
            Node::TestStmt(ref mut n) => {
                walk(
                    v,
                    Rc::new(RefCell::new(Node::VariableAssgn(&mut n.assignment))),
                );
            }
            Node::BuiltinStmt(ref mut n) => {
                walk(v, Rc::new(RefCell::new(Node::Identifier(&mut n.id))));
            }
            Node::Block(ref mut n) => match n {
                Block::Variable(ref mut assgn, ref mut next) => {
                    walk(v, Rc::new(RefCell::new(Node::VariableAssgn(assgn))));
                    walk(v, Rc::new(RefCell::new(Node::Block(&mut *next))));
                }
                Block::Expr(ref mut estmt, ref mut next) => {
                    walk(v, Rc::new(RefCell::new(Node::ExprStmt(estmt))));
                    walk(v, Rc::new(RefCell::new(Node::Block(&mut *next))))
                }
                Block::Return(ref mut expr) => {
                    walk(v, Rc::new(RefCell::new(Node::from_expr(expr))))
                }
            },
            Node::Property(ref mut n) => {
                walk(v, Rc::new(RefCell::new(Node::Identifier(&mut n.key))));
                walk(v, Rc::new(RefCell::new(Node::from_expr(&mut n.value))));
            }
            Node::TextPart(_) => {}
            Node::InterpolatedPart(ref mut n) => {
                walk(v, Rc::new(RefCell::new(Node::from_expr(&mut n.expression))));
            }
            Node::VariableAssgn(ref mut n) => {
                walk(v, Rc::new(RefCell::new(Node::Identifier(&mut n.id))));
                walk(v, Rc::new(RefCell::new(Node::from_expr(&mut n.init))));
            }
            Node::MemberAssgn(ref mut n) => {
                walk(v, Rc::new(RefCell::new(Node::MemberExpr(&mut n.member))));
                walk(v, Rc::new(RefCell::new(Node::from_expr(&mut n.init))));
            }
        };
    }
    v.done(Rc::clone(&node));
}

/// Implementation of Visitor for a mutable closure.
/// We needed Higher-Rank Trait Bounds (`for<'a> ...`) here for compiling.
/// See https://doc.rust-lang.org/nomicon/hrtb.html.
impl<F> Visitor for F
where
    F: for<'a> FnMut(Rc<RefCell<Node<'a>>>),
{
    fn visit(&mut self, node: Rc<RefCell<Node>>) -> bool {
        self(node);
        true
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::ast;
    use crate::ast::check::check;
    use crate::ast::walk;
    use crate::parser::parse_string;
    use crate::semantic::analyze::analyze;
    use crate::semantic::fresh::Fresher;
    use crate::semantic::nodes;

    fn compile(source: &str) -> nodes::Package {
        let file = parse_string("test_walk", source);
        let errs = check(walk::Node::File(&file));
        if errs.len() > 0 {
            panic!(format!("got errors on parsing: {:?}", errs));
        }
        let ast_pkg = ast::Package {
            base: file.base.clone(),
            path: "path/to/pkg".to_string(),
            package: "main".to_string(),
            files: vec![file],
        };
        analyze(ast_pkg, &mut Fresher::new()).unwrap()
    }

    mod node_ids {
        use super::*;

        fn test_walk(source: &str, want: Vec<&str>) {
            let mut sem_pkg = compile(source);
            let mut nodes = Vec::new();
            walk(
                &mut |n: Rc<RefCell<Node>>| nodes.push(format!("{}", *n.borrow())),
                Rc::new(RefCell::new(Node::File(&mut sem_pkg.files[0]))),
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

        impl Visitor for LocationCollector {
            fn visit(&mut self, node: Rc<RefCell<Node>>) -> bool {
                let node = node.borrow();
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

        impl Visitor for TypeCollector {
            fn visit(&mut self, node: Rc<RefCell<Node>>) -> bool {
                let typ = match *node.borrow() {
                    Node::FunctionExpr(ref expr) => Some(expr.typ.clone()),
                    Node::CallExpr(ref expr) => Some(expr.typ.clone()),
                    Node::MemberExpr(ref expr) => Some(expr.typ.clone()),
                    Node::IndexExpr(ref expr) => Some(expr.typ.clone()),
                    Node::BinaryExpr(ref expr) => Some(expr.typ.clone()),
                    Node::UnaryExpr(ref expr) => Some(expr.typ.clone()),
                    Node::LogicalExpr(ref expr) => Some(expr.typ.clone()),
                    Node::ConditionalExpr(ref expr) => Some(expr.typ.clone()),
                    Node::ObjectExpr(ref expr) => Some(expr.typ.clone()),
                    Node::ArrayExpr(ref expr) => Some(expr.typ.clone()),
                    Node::IdentifierExpr(ref expr) => Some(expr.typ.clone()),
                    Node::StringExpr(ref expr) => Some(expr.typ.clone()),
                    Node::StringLit(ref lit) => Some(lit.typ.clone()),
                    Node::BooleanLit(ref lit) => Some(lit.typ.clone()),
                    Node::FloatLit(ref lit) => Some(lit.typ.clone()),
                    Node::IntegerLit(ref lit) => Some(lit.typ.clone()),
                    Node::UintLit(ref lit) => Some(lit.typ.clone()),
                    Node::RegexpLit(ref lit) => Some(lit.typ.clone()),
                    Node::DurationLit(ref lit) => Some(lit.typ.clone()),
                    Node::DateTimeLit(ref lit) => Some(lit.typ.clone()),
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
            walk(&mut v, Rc::new(RefCell::new(Node::Package(&mut pkg))));
            let locs = v.locs;
            assert!(locs.len() > 0);
            for loc in locs {
                assert_ne!(loc, base_loc);
            }
            // now mutate the locations
            let mut_expr_loc = |expr: &mut Expression| {
                match expr {
                    Expression::Identifier(ref mut e) => e.loc = base_loc.clone(),
                    Expression::Array(ref mut e) => e.loc = base_loc.clone(),
                    Expression::Function(ref mut e) => e.loc = base_loc.clone(),
                    Expression::Logical(ref mut e) => e.loc = base_loc.clone(),
                    Expression::Object(ref mut e) => e.loc = base_loc.clone(),
                    Expression::Member(ref mut e) => e.loc = base_loc.clone(),
                    Expression::Index(ref mut e) => e.loc = base_loc.clone(),
                    Expression::Binary(ref mut e) => e.loc = base_loc.clone(),
                    Expression::Unary(ref mut e) => e.loc = base_loc.clone(),
                    Expression::Call(ref mut e) => e.loc = base_loc.clone(),
                    Expression::Conditional(ref mut e) => e.loc = base_loc.clone(),
                    Expression::StringExpr(ref mut e) => e.loc = base_loc.clone(),
                    Expression::Integer(ref mut e) => e.loc = base_loc.clone(),
                    Expression::Float(ref mut e) => e.loc = base_loc.clone(),
                    Expression::StringLit(ref mut e) => e.loc = base_loc.clone(),
                    Expression::Duration(ref mut e) => e.loc = base_loc.clone(),
                    Expression::Uint(ref mut e) => e.loc = base_loc.clone(),
                    Expression::Boolean(ref mut e) => e.loc = base_loc.clone(),
                    Expression::DateTime(ref mut e) => e.loc = base_loc.clone(),
                    Expression::Regexp(ref mut e) => e.loc = base_loc.clone(),
                    _ => (),
                };
            };
            walk(
                &mut |n: Rc<RefCell<Node>>| {
                    match *n.borrow_mut() {
                        Node::Package(ref mut n) => n.loc = base_loc.clone(),
                        Node::File(ref mut n) => n.loc = base_loc.clone(),
                        Node::PackageClause(ref mut n) => n.loc = base_loc.clone(),
                        Node::ImportDeclaration(ref mut n) => n.loc = base_loc.clone(),
                        Node::Identifier(ref mut n) => n.loc = base_loc.clone(),
                        Node::IdentifierExpr(ref mut n) => n.loc = base_loc.clone(),
                        Node::ArrayExpr(ref mut n) => n.loc = base_loc.clone(),
                        Node::FunctionExpr(ref mut n) => n.loc = base_loc.clone(),
                        Node::FunctionParameter(ref mut n) => n.loc = base_loc.clone(),
                        Node::LogicalExpr(ref mut n) => n.loc = base_loc.clone(),
                        Node::ObjectExpr(ref mut n) => n.loc = base_loc.clone(),
                        Node::MemberExpr(ref mut n) => n.loc = base_loc.clone(),
                        Node::IndexExpr(ref mut n) => n.loc = base_loc.clone(),
                        Node::BinaryExpr(ref mut n) => n.loc = base_loc.clone(),
                        Node::UnaryExpr(ref mut n) => n.loc = base_loc.clone(),
                        Node::CallExpr(ref mut n) => n.loc = base_loc.clone(),
                        Node::ConditionalExpr(ref mut n) => n.loc = base_loc.clone(),
                        Node::StringExpr(ref mut n) => n.loc = base_loc.clone(),
                        Node::IntegerLit(ref mut n) => n.loc = base_loc.clone(),
                        Node::FloatLit(ref mut n) => n.loc = base_loc.clone(),
                        Node::StringLit(ref mut n) => n.loc = base_loc.clone(),
                        Node::DurationLit(ref mut n) => n.loc = base_loc.clone(),
                        Node::UintLit(ref mut n) => n.loc = base_loc.clone(),
                        Node::BooleanLit(ref mut n) => n.loc = base_loc.clone(),
                        Node::DateTimeLit(ref mut n) => n.loc = base_loc.clone(),
                        Node::RegexpLit(ref mut n) => n.loc = base_loc.clone(),
                        Node::ExprStmt(ref mut n) => n.loc = base_loc.clone(),
                        Node::OptionStmt(ref mut n) => n.loc = base_loc.clone(),
                        Node::ReturnStmt(ref mut n) => n.loc = base_loc.clone(),
                        Node::TestStmt(ref mut n) => n.loc = base_loc.clone(),
                        Node::BuiltinStmt(ref mut n) => n.loc = base_loc.clone(),
                        Node::Block(Block::Variable(ref mut assgn, _)) => {
                            assgn.loc = base_loc.clone()
                        }
                        Node::Block(Block::Expr(ref mut estmt, _)) => {
                            mut_expr_loc(&mut estmt.expression)
                        }
                        Node::Block(Block::Return(ref mut expr)) => mut_expr_loc(expr),
                        Node::Property(ref mut n) => n.loc = base_loc.clone(),
                        Node::TextPart(ref mut n) => n.loc = base_loc.clone(),
                        Node::InterpolatedPart(ref mut n) => n.loc = base_loc.clone(),
                        Node::VariableAssgn(ref mut n) => n.loc = base_loc.clone(),
                        Node::MemberAssgn(ref mut n) => n.loc = base_loc.clone(),
                    };
                },
                Rc::new(RefCell::new(Node::Package(&mut pkg))),
            );
            // now assert that every location is the base one
            let mut v = LocationCollector::new();
            walk(&mut v, Rc::new(RefCell::new(Node::Package(&mut pkg))));
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
            walk(&mut v, Rc::new(RefCell::new(Node::Package(&mut pkg))));
            let types = v.types;
            assert!(types.len() > 0);
            // every tvar has a unique id
            let mut ids: HashSet<u64> = HashSet::new();
            for tvar in types {
                assert!(!ids.contains(&tvar.0));
                ids.insert(tvar.0);
            }
            // now mutate the types
            walk(
                &mut |n: Rc<RefCell<Node>>| {
                    match *n.borrow_mut() {
                        Node::IdentifierExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        Node::ArrayExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        Node::FunctionExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        Node::LogicalExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        Node::ObjectExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        Node::MemberExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        Node::IndexExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        Node::BinaryExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        Node::UnaryExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        Node::CallExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        Node::ConditionalExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        Node::StringExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        Node::IntegerLit(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        Node::FloatLit(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        Node::StringLit(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        Node::DurationLit(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        Node::UintLit(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        Node::BooleanLit(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        Node::DateTimeLit(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        Node::RegexpLit(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        _ => (),
                    };
                },
                Rc::new(RefCell::new(Node::Package(&mut pkg))),
            );
            // now assert that every type is the invalid one
            let mut v = TypeCollector::new();
            walk(&mut v, Rc::new(RefCell::new(Node::Package(&mut pkg))));
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

        impl Visitor for NestingCounter {
            fn visit(&mut self, node: Rc<RefCell<Node>>) -> bool {
                match *node.borrow() {
                    Node::Block(_) => self.count += 1,
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
            walk(&mut v, Rc::new(RefCell::new(Node::Package(&mut pkg))));
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

        impl Visitor for RepeatedPlusChecker {
            fn visit(&mut self, node: Rc<RefCell<Node>>) -> bool {
                match *node.borrow() {
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

            fn done(&mut self, node: Rc<RefCell<Node>>) {
                let node = node.borrow();
                if let Node::FunctionExpr(_) = *node {
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
            walk(&mut v, Rc::new(RefCell::new(Node::Package(&mut ok))));
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
            walk(&mut v, Rc::new(RefCell::new(Node::Package(&mut not_ok1))));
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
            walk(&mut v, Rc::new(RefCell::new(Node::Package(&mut not_ok2))));
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
            walk(&mut v, Rc::new(RefCell::new(Node::Package(&mut not_ok3))));
            assert_eq!(v.err.as_str(), "repeated + on line 6");
        }
    }
}
