use std::fmt;

use crate::{
    ast::SourceLocation,
    semantic::{nodes::*, types::MonoType},
};

mk_node!(
    /// Represents any structure that can appear in the semantic graph.
    #[derive(Clone, Debug)]
    #[allow(missing_docs)]
    Node
);

/// Used by [`walk`] to recursively visit a semantic graph.
/// One can implement `Visitor` or use a `FnMut(Node)`.
///
/// ## Example
///
/// Print out the nodes of a semantic graph
///
/// ```no_run
/// use fluxcore::ast;
/// use fluxcore::semantic::walk::{Node, walk};
/// use fluxcore::semantic::nodes::*;
///
/// let mut pkg = Package {
///     loc:  ast::BaseNode::default().location,
///     package: "main".to_string(),
///     files: vec![],
/// };
/// walk(
///     &mut |n: Node| println!("{}", n),
///     Node::Package(&pkg),
/// );
/// ```
///
/// ## Example
///
/// A "scoped" visitor that errors if finds more than one addition operation in the same scope:
///
/// ```no_run
/// use fluxcore::ast::Operator::AdditionOperator;
/// use fluxcore::semantic::walk::{Node, Visitor};
/// use fluxcore::semantic::nodes::*;
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
///     fn visit(&mut self, node: Node<'a>) -> bool {
///         match node {
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
///     fn done(&mut self, node: Node) {
///         if let Node::Block(_) = node {
///             self.plus.pop();
///         }
///     }
/// }
/// ```
pub trait Visitor<'a>: Sized {
    /// `visit` is called for a node.
    /// When the `Visitor` is used in [`walk`], the boolean value returned
    /// is used to continue walking (`true`) or stop (`false`).
    fn visit(&mut self, node: Node<'a>) -> bool;
    /// `done` is called for a node once it has been visited along with all of its children.
    /// The default is to do nothing.
    fn done(&mut self, _: Node<'a>) {}
}

/// Recursively visits children of a node given a [`Visitor`].
/// Nodes are visited in depth-first order.
pub fn walk<'a, T>(v: &mut T, node: Node<'a>)
where
    T: Visitor<'a>,
{
    if v.visit(node.clone()) {
        match node.clone() {
            Node::Package(n) => {
                for file in n.files.iter() {
                    walk(v, Node::File(file));
                }
            }
            Node::File(n) => {
                if let Some(ref pkg) = n.package {
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
                if let Some(ref alias) = n.alias {
                    walk(v, Node::Identifier(alias));
                }
                walk(v, Node::StringLit(&n.path));
            }
            Node::Identifier(_) => {}
            Node::IdentifierExpr(_) => {}
            Node::ArrayExpr(n) => {
                for element in n.elements.iter() {
                    walk(v, Node::from_expr(element));
                }
            }
            Node::DictExpr(n) => {
                for (key, val) in n.elements.iter() {
                    walk(v, Node::from_expr(key));
                    walk(v, Node::from_expr(val));
                }
            }
            Node::FunctionExpr(n) => {
                for param in n.params.iter() {
                    walk(v, Node::FunctionParameter(param));
                }
                walk(v, Node::Block(&n.body));
                if let Some(ref vectorized) = n.vectorized {
                    walk(v, Node::from_expr(vectorized));
                }
            }
            Node::FunctionParameter(n) => {
                walk(v, Node::Identifier(&n.key));
                if let Some(ref def) = n.default {
                    walk(v, Node::from_expr(def));
                }
            }
            Node::LogicalExpr(n) => {
                walk(v, Node::from_expr(&n.left));
                walk(v, Node::from_expr(&n.right));
            }
            Node::ObjectExpr(n) => {
                if let Some(ref i) = n.with {
                    walk(v, Node::IdentifierExpr(i));
                }
                for prop in n.properties.iter() {
                    walk(v, Node::Property(prop));
                }
            }
            Node::MemberExpr(n) => {
                walk(v, Node::from_expr(&n.object));
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
            Node::CallExpr(n) => {
                walk(v, Node::from_expr(&n.callee));
                if let Some(ref p) = n.pipe {
                    walk(v, Node::from_expr(p));
                }
                for arg in n.arguments.iter() {
                    walk(v, Node::Property(arg));
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
            Node::IntegerLit(_) => {}
            Node::FloatLit(_) => {}
            Node::StringLit(_) => {}
            Node::DurationLit(_) => {}
            Node::UintLit(_) => {}
            Node::BooleanLit(_) => {}
            Node::DateTimeLit(_) => {}
            Node::RegexpLit(_) => {}
            Node::ExprStmt(n) => {
                walk(v, Node::from_expr(&n.expression));
            }
            Node::OptionStmt(n) => {
                walk(v, Node::from_assignment(&n.assignment));
            }
            Node::ReturnStmt(n) => {
                walk(v, Node::from_expr(&n.argument));
            }
            Node::TestStmt(n) => {
                walk(v, Node::VariableAssgn(&n.assignment));
            }
            Node::TestCaseStmt(n) => {
                walk(v, Node::Identifier(&n.id));
                walk(v, Node::Block(&n.block));
            }
            Node::BuiltinStmt(n) => {
                walk(v, Node::Identifier(&n.id));
            }
            Node::ErrorStmt(_) => {}
            Node::Block(n) => match n {
                Block::Variable(ref assgn, ref next) => {
                    walk(v, Node::VariableAssgn(assgn));
                    walk(v, Node::Block(&*next));
                }
                Block::Expr(estmt, next) => {
                    walk(v, Node::ExprStmt(estmt));
                    walk(v, Node::Block(&*next))
                }
                Block::Return(ref ret_stmt) => walk(v, Node::ReturnStmt(ret_stmt)),
            },
            Node::Property(n) => {
                walk(v, Node::Identifier(&n.key));
                walk(v, Node::from_expr(&n.value));
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
            Node::ErrorExpr(_) => (),
        };
    }
    v.done(node.clone());
}

/// Implementation of Visitor for a mutable closure.
/// We need Higher-Rank Trait Bounds (`for<'a> ...`) here for compiling.
/// See <https://doc.rust-lang.org/nomicon/hrtb.html>.
impl<'a, F> Visitor<'a> for F
where
    F: FnMut(Node<'a>),
{
    fn visit(&mut self, node: Node<'a>) -> bool {
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
                &mut |n: Node| nodes.push(format!("{}", n)),
                Node::File(&sem_pkg.files[0]),
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
            test_walk("builtin a : int", vec!["File", "BuiltinStmt", "Identifier"])
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
            fn visit(&mut self, node: Node<'a>) -> bool {
                match node {
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
            walk(&mut v, Node::Package(&pkg));
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
            fn visit(&mut self, node: Node<'a>) -> bool {
                match node {
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

            fn done(&mut self, node: Node) {
                if let Node::FunctionExpr(_) = node {
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
            walk(&mut v, Node::Package(&ok));
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
            walk(&mut v, Node::Package(&not_ok1));
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
            walk(&mut v, Node::Package(&not_ok2));
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
            walk(&mut v, Node::Package(&not_ok3));
            assert_eq!(v.err.as_str(), "repeated + on line 6");
        }
    }
}
