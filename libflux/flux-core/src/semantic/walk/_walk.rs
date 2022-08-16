use std::fmt;

use crate::{
    ast::SourceLocation,
    semantic::{nodes::*, types::MonoType},
};

mk_node!(
    /// Represents any structure that can appear in the semantic graph.
    #[derive(Clone, Copy, Debug)]
    #[allow(missing_docs)]
    Node,
    Visitor<'a>,
    walk,
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
pub trait Visitor<'a> {
    /// `visit` is called for a node.
    /// When the `Visitor` is used in [`walk`], the boolean value returned
    /// is used to continue walking (`true`) or stop (`false`).
    fn visit(&mut self, node: Node<'a>) -> bool;
    /// `done` is called for a node once it has been visited along with all of its children.
    /// The default is to do nothing.
    fn done(&mut self, _: Node<'a>) {}
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

    mod nesting {
        use super::*;
        use crate::ast::Operator::AdditionOperator;

        // NestingCounter counts the number of nested Blocks found while walking.
        struct NestingCounter {
            count: u8,
        }

        impl<'a> Visitor<'a> for NestingCounter {
            fn visit(&mut self, node: Node<'a>) -> bool {
                if let Node::Block(_) = node {
                    self.count += 1
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
                    Node::BinaryExpr(expr) => {
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
