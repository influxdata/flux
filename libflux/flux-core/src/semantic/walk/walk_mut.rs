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
    NodeMut mut
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
            NodeMut::ErrorExpr(ref mut n) => **n = loc,
            NodeMut::ExprStmt(ref mut n) => n.loc = loc,
            NodeMut::OptionStmt(ref mut n) => n.loc = loc,
            NodeMut::ReturnStmt(ref mut n) => n.loc = loc,
            NodeMut::TestStmt(ref mut n) => n.loc = loc,
            NodeMut::TestCaseStmt(ref mut n) => n.loc = loc,
            NodeMut::BuiltinStmt(ref mut n) => n.loc = loc,
            NodeMut::ErrorStmt(ref mut n) => **n = loc,
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
///     fn visit(&mut self, node: &mut NodeMut) -> bool {
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
pub trait VisitorMut: Sized {
    /// `visit` is called for a node.
    /// When the `VisitorMut` is used in [`walk_mut`], the boolean value returned
    /// is used to continue walking (`true`) or stop (`false`).
    fn visit(&mut self, node: &mut NodeMut) -> bool;
    /// `done` is called for a node once it has been visited along with all of its children.
    /// The default is to do nothing.
    fn done(&mut self, _: &mut NodeMut) {}
}

/// Recursively visits children of a node given a [`VisitorMut`].
/// Nodes are visited in depth-first order.
pub fn walk_mut<T>(v: &mut T, node: &mut NodeMut)
where
    T: VisitorMut,
{
    if v.visit(node) {
        match node {
            NodeMut::Package(ref mut n) => {
                for file in n.files.iter_mut() {
                    walk_mut(v, &mut NodeMut::File(file));
                }
            }
            NodeMut::File(ref mut n) => {
                if let Some(pkg) = n.package.as_mut() {
                    walk_mut(v, &mut NodeMut::PackageClause(pkg));
                }
                for imp in n.imports.iter_mut() {
                    walk_mut(v, &mut NodeMut::ImportDeclaration(imp));
                }
                for stmt in n.body.iter_mut() {
                    walk_mut(v, &mut NodeMut::from_stmt(stmt));
                }
            }
            NodeMut::PackageClause(ref mut n) => {
                walk_mut(v, &mut NodeMut::Identifier(&mut n.name));
            }
            NodeMut::ImportDeclaration(ref mut n) => {
                if let Some(alias) = n.alias.as_mut() {
                    walk_mut(v, &mut NodeMut::Identifier(alias));
                }
                walk_mut(v, &mut NodeMut::StringLit(&mut n.path));
            }
            NodeMut::Identifier(_) => {}
            NodeMut::IdentifierExpr(_) => {}
            NodeMut::ArrayExpr(ref mut n) => {
                for element in n.elements.iter_mut() {
                    walk_mut(v, &mut NodeMut::from_expr(element));
                }
            }
            NodeMut::DictExpr(ref mut n) => {
                for item in n.elements.iter_mut() {
                    walk_mut(v, &mut NodeMut::from_expr(&mut item.0));
                    walk_mut(v, &mut NodeMut::from_expr(&mut item.1));
                }
            }
            NodeMut::FunctionExpr(ref mut n) => {
                for param in n.params.iter_mut() {
                    walk_mut(v, &mut NodeMut::FunctionParameter(param));
                }
                walk_mut(v, &mut NodeMut::Block(&mut n.body));
            }
            NodeMut::FunctionParameter(ref mut n) => {
                walk_mut(v, &mut NodeMut::Identifier(&mut n.key));
                if let Some(def) = n.default.as_mut() {
                    walk_mut(v, &mut NodeMut::from_expr(def));
                }
            }
            NodeMut::LogicalExpr(ref mut n) => {
                walk_mut(v, &mut NodeMut::from_expr(&mut n.left));
                walk_mut(v, &mut NodeMut::from_expr(&mut n.right));
            }
            NodeMut::ObjectExpr(ref mut n) => {
                if let Some(i) = n.with.as_mut() {
                    walk_mut(v, &mut NodeMut::IdentifierExpr(i));
                }
                for prop in n.properties.iter_mut() {
                    walk_mut(v, &mut NodeMut::Property(prop));
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
                if let Some(p) = n.pipe.as_mut() {
                    walk_mut(v, &mut NodeMut::from_expr(p));
                }
                for arg in n.arguments.iter_mut() {
                    walk_mut(v, &mut NodeMut::Property(arg));
                }
            }
            NodeMut::ConditionalExpr(ref mut n) => {
                walk_mut(v, &mut NodeMut::from_expr(&mut n.test));
                walk_mut(v, &mut NodeMut::from_expr(&mut n.consequent));
                walk_mut(v, &mut NodeMut::from_expr(&mut n.alternate));
            }
            NodeMut::StringExpr(ref mut n) => {
                for part in n.parts.iter_mut() {
                    walk_mut(v, &mut NodeMut::from_string_expr_part(part));
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
            NodeMut::TestCaseStmt(ref mut n) => {
                walk_mut(v, &mut NodeMut::Identifier(&mut n.id));
                walk_mut(v, &mut NodeMut::Block(&mut n.block));
            }
            NodeMut::BuiltinStmt(ref mut n) => {
                walk_mut(v, &mut NodeMut::Identifier(&mut n.id));
            }
            NodeMut::ErrorStmt(_) => {}
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
            NodeMut::ErrorExpr(_) => (),
        };
    }
    v.done(node);
}

/// Implementation of VisitorMut for a mutable closure.
/// We need Higher-Rank Trait Bounds (`for<'a> ...`) here for compiling.
/// See <https://doc.rust-lang.org/nomicon/hrtb.html>.
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
    use crate::{ast, semantic::walk::test_utils::compile};

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
                &mut NodeMut::Package(&mut pkg),
            );
            // now assert that every type is the invalid one
            let mut v = TypeCollector::new();
            walk_mut(&mut v, &mut NodeMut::Package(&mut pkg));
            let types = v.types;
            assert!(types.len() > 0);
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
