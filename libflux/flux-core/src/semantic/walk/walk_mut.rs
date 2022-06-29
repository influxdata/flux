use std::fmt;

use crate::{
    ast::SourceLocation,
    semantic::{nodes::*, types::MonoType},
};

mk_node!(
    /// Represents any structure that can appear in the semantic graph.
    /// It also enables mutability of the wrapped semantic node.
    #[derive(Debug)]
    #[allow(missing_docs)]
    NodeMut,
    VisitorMut,
    walk_mut,
    mut
);

impl NodeMut<'_> {
    #[allow(missing_docs)]
    pub fn set_loc(&mut self, loc: SourceLocation) {
        match self {
            NodeMut::Package(ref mut n) => n.loc = loc,
            NodeMut::File(ref mut n) => n.loc = loc,
            NodeMut::PackageClause(ref mut n) => n.loc = loc,
            NodeMut::ImportDeclaration(ref mut n) => n.loc = loc,
            NodeMut::Identifier(ref mut n) => n.loc = loc,
            NodeMut::Expr(ref mut n) => NodeMut::reduce_expr(n).set_loc(loc),
            NodeMut::IdentifierExpr(ref mut n) => n.loc = loc,
            NodeMut::ArrayExpr(ref mut n) => n.loc = loc,
            NodeMut::DictExpr(ref mut n) => n.loc = loc,
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
            NodeMut::ErrorExpr(ref mut n) => n.loc = loc,
            NodeMut::ExprStmt(ref mut n) => n.loc = loc,
            NodeMut::OptionStmt(ref mut n) => n.loc = loc,
            NodeMut::ReturnStmt(ref mut n) => n.loc = loc,
            NodeMut::TestCaseStmt(ref mut n) => n.loc = loc,
            NodeMut::BuiltinStmt(ref mut n) => n.loc = loc,
            NodeMut::ErrorStmt(ref mut n) => n.loc = loc,
            NodeMut::Block(_) => (),
            NodeMut::Property(ref mut n) => n.loc = loc,
            NodeMut::TextPart(ref mut n) => n.loc = loc,
            NodeMut::InterpolatedPart(ref mut n) => n.loc = loc,
            NodeMut::VariableAssgn(ref mut n) => n.loc = loc,
            NodeMut::MemberAssgn(ref mut n) => n.loc = loc,
        };
    }
}

/// Used by [`walk_mut`] to recursively visit a semantic graph and mutate it.
/// The trait makes it possible to mutate both the visitor and the nodes while visiting.
/// One can implement `VisitorMut` or use a `FnMut(NodeMut)`.
///
/// # Examples
///
/// A visitor that mutates node types:
///
/// ```no_run
/// use fluxcore::semantic::walk::{NodeMut, VisitorMut};
/// use fluxcore::semantic::types::*;
///
/// struct TypeMutator {}
///
/// impl VisitorMut for TypeMutator {
///     fn visit(&mut self, node: &mut NodeMut<'_>) -> bool {
///         match node {
///             NodeMut::IdentifierExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
///             NodeMut::ArrayExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
///             NodeMut::FunctionExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
///             NodeMut::ObjectExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
///             NodeMut::MemberExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
///             NodeMut::IndexExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
///             NodeMut::BinaryExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
///             NodeMut::UnaryExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
///             NodeMut::CallExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
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
pub trait VisitorMut {
    /// `visit` is called for a node.
    /// When the `VisitorMut` is used in [`walk_mut`], the boolean value returned
    /// is used to continue walking (`true`) or stop (`false`).
    fn visit(&mut self, node: &mut NodeMut<'_>) -> bool;
    /// `done` is called for a node once it has been visited along with all of its children.
    /// The default is to do nothing.
    fn done(&mut self, _: &mut NodeMut<'_>) {}
}

/// Implementation of VisitorMut for a mutable closure.
/// We need Higher-Rank Trait Bounds (`for<'a> ...`) here for compiling.
/// See <https://doc.rust-lang.org/nomicon/hrtb.html>.
impl<F> VisitorMut for F
where
    F: for<'b> FnMut(&mut NodeMut<'_>),
{
    fn visit(&mut self, node: &mut NodeMut<'_>) -> bool {
        self(node);
        true
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::{ast, semantic::walk::test_utils::compile};

    mod mutate_nodes {

        use super::*;
        use crate::semantic::types::{MonoType, Tvar};

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
            types: Vec<MonoType>,
        }

        impl TypeCollector {
            fn new() -> TypeCollector {
                TypeCollector { types: Vec::new() }
            }
        }

        impl VisitorMut for TypeCollector {
            fn visit(&mut self, node: &mut NodeMut) -> bool {
                let typ = match node {
                    NodeMut::IdentifierExpr(n) => {
                        Some(Expression::Identifier((*n).clone()).type_of())
                    }
                    NodeMut::ArrayExpr(n) => {
                        Some(Expression::Array(Box::new((*n).clone())).type_of())
                    }
                    NodeMut::FunctionExpr(n) => {
                        Some(Expression::Function(Box::new((*n).clone())).type_of())
                    }
                    NodeMut::LogicalExpr(n) => {
                        Some(Expression::Logical(Box::new((*n).clone())).type_of())
                    }
                    NodeMut::ObjectExpr(n) => {
                        Some(Expression::Object(Box::new((*n).clone())).type_of())
                    }
                    NodeMut::MemberExpr(n) => {
                        Some(Expression::Member(Box::new((*n).clone())).type_of())
                    }
                    NodeMut::IndexExpr(n) => {
                        Some(Expression::Index(Box::new((*n).clone())).type_of())
                    }
                    NodeMut::BinaryExpr(n) => {
                        Some(Expression::Binary(Box::new((*n).clone())).type_of())
                    }
                    NodeMut::UnaryExpr(n) => {
                        Some(Expression::Unary(Box::new((*n).clone())).type_of())
                    }
                    NodeMut::CallExpr(n) => {
                        Some(Expression::Call(Box::new((*n).clone())).type_of())
                    }
                    NodeMut::ConditionalExpr(n) => {
                        Some(Expression::Conditional(Box::new((*n).clone())).type_of())
                    }
                    NodeMut::StringExpr(n) => {
                        Some(Expression::StringExpr(Box::new((*n).clone())).type_of())
                    }
                    NodeMut::IntegerLit(n) => Some(Expression::Integer((*n).clone()).type_of()),
                    NodeMut::FloatLit(n) => Some(Expression::Float((*n).clone()).type_of()),
                    NodeMut::StringLit(n) => Some(Expression::StringLit((*n).clone()).type_of()),
                    NodeMut::DurationLit(n) => Some(Expression::Duration((*n).clone()).type_of()),
                    NodeMut::UintLit(n) => Some(Expression::Uint((*n).clone()).type_of()),
                    NodeMut::BooleanLit(n) => Some(Expression::Boolean((*n).clone()).type_of()),
                    NodeMut::DateTimeLit(n) => Some(Expression::DateTime((*n).clone()).type_of()),
                    NodeMut::RegexpLit(n) => Some(Expression::Regexp((*n).clone()).type_of()),
                    _ => None,
                };
                if let Some(typ) = typ {
                    self.types.push(typ);
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
            walk_mut(&mut v, NodeMut::Package(&mut pkg));
            let locs = v.locs;
            assert!(!locs.is_empty());
            for loc in locs {
                assert_ne!(loc, base_loc);
            }
            // now mutate the locations
            walk_mut(
                &mut |n: &mut NodeMut| n.set_loc(base_loc.clone()),
                NodeMut::Package(&mut pkg),
            );
            // now assert that every location is the base one
            let mut v = LocationCollector::new();
            walk_mut(&mut v, NodeMut::Package(&mut pkg));
            let locs = v.locs;
            assert!(!locs.is_empty());
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
            walk_mut(&mut v, NodeMut::Package(&mut pkg));
            let types = v.types;
            assert!(!types.is_empty());
            // no type is a type variable
            assert!(types.iter().all(|t| !matches!(t, MonoType::Var(_))));
            // now mutate the types
            walk_mut(
                &mut |n: &mut NodeMut| {
                    match n {
                        NodeMut::IdentifierExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        NodeMut::ArrayExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        NodeMut::FunctionExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        NodeMut::ObjectExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        NodeMut::MemberExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        NodeMut::IndexExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        NodeMut::BinaryExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        NodeMut::UnaryExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        NodeMut::CallExpr(ref mut n) => n.typ = MonoType::Var(Tvar(1234)),
                        _ => (),
                    };
                },
                NodeMut::Package(&mut pkg),
            );
            // now assert that every type is the invalid one
            let mut v = TypeCollector::new();
            walk_mut(&mut v, NodeMut::Package(&mut pkg));
            let types = v.types;
            assert!(!types.is_empty());
            for tvar in types {
                if let MonoType::Var(tvar) = tvar {
                    assert_eq!(tvar, Tvar(1234));
                }
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
                if let NodeMut::Block(_) = node {
                    self.count += 1
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
            walk_mut(&mut v, NodeMut::Package(&mut pkg));
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
            walk_mut(&mut v, NodeMut::Package(&mut ok));
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
            walk_mut(&mut v, NodeMut::Package(&mut not_ok1));
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
            walk_mut(&mut v, NodeMut::Package(&mut not_ok2));
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
            walk_mut(&mut v, NodeMut::Package(&mut not_ok3));
            assert_eq!(v.err.as_str(), "repeated + on line 6");
        }
    }
}
