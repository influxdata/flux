//! Walking the semantic graph.

macro_rules! mk_node {
    (
        $(#[$attr:meta])*
        $name: ident,
        $visitor: ident $(<$visitor_lt: lifetime>)?,
        $walk: ident,
        $($mut: tt)?
    ) => {

        $(#[$attr])*
        #[derive(derive_more::From)]
        pub enum $name<'a> {
            Package(&'a $($mut)? Package),
            File(&'a $($mut)? File),
            PackageClause(&'a $($mut)? PackageClause),
            ImportDeclaration(&'a $($mut)? ImportDeclaration),
            Identifier(&'a $($mut)? Identifier),
            FunctionParameter(&'a $($mut)? FunctionParameter),
            Block(&'a $($mut)? Block),
            Property(&'a $($mut)? Property),

            // Expressions.
            Expr(&'a $($mut)? Expression),

            IdentifierExpr(&'a $($mut)? IdentifierExpr),
            ArrayExpr(&'a $($mut)? ArrayExpr),
            DictExpr(&'a $($mut)? DictExpr),
            FunctionExpr(&'a $($mut)? FunctionExpr),
            LogicalExpr(&'a $($mut)? LogicalExpr),
            ObjectExpr(&'a $($mut)? ObjectExpr),
            MemberExpr(&'a $($mut)? MemberExpr),
            IndexExpr(&'a $($mut)? IndexExpr),
            BinaryExpr(&'a $($mut)? BinaryExpr),
            UnaryExpr(&'a $($mut)? UnaryExpr),
            CallExpr(&'a $($mut)? CallExpr),
            ConditionalExpr(&'a $($mut)? ConditionalExpr),
            StringExpr(&'a $($mut)? StringExpr),
            IntegerLit(&'a $($mut)? IntegerLit),
            FloatLit(&'a $($mut)? FloatLit),
            StringLit(&'a $($mut)? StringLit),
            DurationLit(&'a $($mut)? DurationLit),
            UintLit(&'a $($mut)? UintLit),
            BooleanLit(&'a $($mut)? BooleanLit),
            DateTimeLit(&'a $($mut)? DateTimeLit),
            RegexpLit(&'a $($mut)? RegexpLit),
            LabelLit(&'a $($mut)? LabelLit),
            ErrorExpr(&'a $($mut)? BadExpr),

            // Statements.
            ExprStmt(&'a $($mut)? ExprStmt),
            OptionStmt(&'a $($mut)? OptionStmt),
            ReturnStmt(&'a $($mut)? ReturnStmt),
            TestCaseStmt(&'a $($mut)? TestCaseStmt),
            BuiltinStmt(&'a $($mut)? BuiltinStmt),
            ErrorStmt(&'a $($mut)? BadStmt),

            // StringExprPart.
            TextPart(&'a $($mut)? TextPart),
            InterpolatedPart(&'a $($mut)? InterpolatedPart),

            // Assignment.
            VariableAssgn(&'a $($mut)? VariableAssgn), // Native variable assignment
            MemberAssgn(&'a $($mut)? MemberAssgn),
        }

        impl<'a> fmt::Display for $name<'a> {
            fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
                match self {
                    Self::Package(_) => write!(f, "Package"),
                    Self::File(_) => write!(f, "File"),
                    Self::PackageClause(_) => write!(f, "PackageClause"),
                    Self::ImportDeclaration(_) => write!(f, "ImportDeclaration"),
                    Self::Identifier(_) => write!(f, "Identifier"),
                    Self::Expr(_) => write!(f, "Expr"),
                    Self::IdentifierExpr(_) => write!(f, "IdentifierExpr"),
                    Self::ArrayExpr(_) => write!(f, "ArrayExpr"),
                    Self::DictExpr(_) => write!(f, "DictExpr"),
                    Self::FunctionExpr(_) => write!(f, "FunctionExpr"),
                    Self::FunctionParameter(_) => write!(f, "FunctionParameter"),
                    Self::LogicalExpr(_) => write!(f, "LogicalExpr"),
                    Self::ObjectExpr(_) => write!(f, "ObjectExpr"),
                    Self::MemberExpr(_) => write!(f, "MemberExpr"),
                    Self::IndexExpr(_) => write!(f, "IndexExpr"),
                    Self::BinaryExpr(_) => write!(f, "BinaryExpr"),
                    Self::UnaryExpr(_) => write!(f, "UnaryExpr"),
                    Self::CallExpr(_) => write!(f, "CallExpr"),
                    Self::ConditionalExpr(_) => write!(f, "ConditionalExpr"),
                    Self::StringExpr(_) => write!(f, "StringExpr"),
                    Self::IntegerLit(_) => write!(f, "IntegerLit"),
                    Self::FloatLit(_) => write!(f, "FloatLit"),
                    Self::StringLit(_) => write!(f, "StringLit"),
                    Self::DurationLit(_) => write!(f, "DurationLit"),
                    Self::UintLit(_) => write!(f, "UintLit"),
                    Self::BooleanLit(_) => write!(f, "BooleanLit"),
                    Self::DateTimeLit(_) => write!(f, "DateTimeLit"),
                    Self::RegexpLit(_) => write!(f, "RegexpLit"),
                    Self::LabelLit(_) => write!(f, "LabelLit"),
                    Self::ErrorExpr(_) => write!(f, "ErrorExpr"),
                    Self::ExprStmt(_) => write!(f, "ExprStmt"),
                    Self::OptionStmt(_) => write!(f, "OptionStmt"),
                    Self::ReturnStmt(_) => write!(f, "ReturnStmt"),
                    Self::TestCaseStmt(_) => write!(f, "TestCaseStmt"),
                    Self::BuiltinStmt(_) => write!(f, "BuiltinStmt"),
                    Self::ErrorStmt(_) => write!(f, "ErrorStmt"),
                    Self::Block(n) => match n {
                        Block::Variable(_, _) => write!(f, "Block::Variable"),
                        Block::Expr(_, _) => write!(f, "Block::Expr"),
                        Block::Return(_) => write!(f, "Block::Return"),
                    },
                    Self::Property(_) => write!(f, "Property"),
                    Self::TextPart(_) => write!(f, "TextPart"),
                    Self::InterpolatedPart(_) => write!(f, "InterpolatedPart"),
                    Self::VariableAssgn(_) => write!(f, "VariableAssgn"),
                    Self::MemberAssgn(_) => write!(f, "MemberAssgn"),
                }
            }
        }
        impl<'a> $name<'a> {
            /// Returns the source location of a semantic graph node.
            pub fn loc(&self) -> &SourceLocation {
                match self {
                    Self::Package(n) => &n.loc,
                    Self::File(n) => &n.loc,
                    Self::PackageClause(n) => &n.loc,
                    Self::ImportDeclaration(n) => &n.loc,
                    Self::Identifier(n) => &n.loc,
                    Self::IdentifierExpr(n) => &n.loc,
                    Self::ArrayExpr(n) => &n.loc,
                    Self::DictExpr(n) => &n.loc,
                    Self::FunctionExpr(n) => &n.loc,
                    Self::FunctionParameter(n) => &n.loc,
                    Self::Expr(n) => n.loc(),
                    Self::LogicalExpr(n) => &n.loc,
                    Self::ObjectExpr(n) => &n.loc,
                    Self::MemberExpr(n) => &n.loc,
                    Self::IndexExpr(n) => &n.loc,
                    Self::BinaryExpr(n) => &n.loc,
                    Self::UnaryExpr(n) => &n.loc,
                    Self::CallExpr(n) => &n.loc,
                    Self::ConditionalExpr(n) => &n.loc,
                    Self::StringExpr(n) => &n.loc,
                    Self::IntegerLit(n) => &n.loc,
                    Self::FloatLit(n) => &n.loc,
                    Self::StringLit(n) => &n.loc,
                    Self::DurationLit(n) => &n.loc,
                    Self::UintLit(n) => &n.loc,
                    Self::BooleanLit(n) => &n.loc,
                    Self::DateTimeLit(n) => &n.loc,
                    Self::RegexpLit(n) => &n.loc,
                    Self::LabelLit(n) => &n.loc,
                    Self::ExprStmt(n) => &n.loc,
                    Self::OptionStmt(n) => &n.loc,
                    Self::ReturnStmt(n) => &n.loc,
                    Self::TestCaseStmt(n) => &n.loc,
                    Self::BuiltinStmt(n) => &n.loc,
                    Self::ErrorStmt(n) => &n.loc,
                    Self::Block(n) => n.loc(),
                    Self::Property(n) => &n.loc,
                    Self::TextPart(n) => &n.loc,
                    Self::InterpolatedPart(n) => &n.loc,
                    Self::VariableAssgn(n) => &n.loc,
                    Self::MemberAssgn(n) => &n.loc,
                    Self::ErrorExpr(n) => &n.loc,
                }
            }

            /// Returns the type of a semantic graph node.
            pub fn type_of(&self) -> Option<MonoType> {
                match self {
                    Self::IdentifierExpr(n) => Some(Expression::Identifier((*n).clone()).type_of()),
                    Self::ArrayExpr(n) => Some(Expression::Array(Box::new((*n).clone())).type_of()),
                    Self::DictExpr(n) => Some(Expression::Dict(Box::new((*n).clone())).type_of()),
                    Self::FunctionExpr(n) => {
                        Some(Expression::Function(Box::new((*n).clone())).type_of())
                    }
                    Self::LogicalExpr(n) => {
                        Some(Expression::Logical(Box::new((*n).clone())).type_of())
                    }
                    Self::ObjectExpr(n) => {
                        Some(Expression::Object(Box::new((*n).clone())).type_of())
                    }
                    Self::MemberExpr(n) => {
                        Some(Expression::Member(Box::new((*n).clone())).type_of())
                    }
                    Self::IndexExpr(n) => Some(Expression::Index(Box::new((*n).clone())).type_of()),
                    Self::BinaryExpr(n) => {
                        Some(Expression::Binary(Box::new((*n).clone())).type_of())
                    }
                    Self::UnaryExpr(n) => Some(Expression::Unary(Box::new((*n).clone())).type_of()),
                    Self::CallExpr(n) => Some(Expression::Call(Box::new((*n).clone())).type_of()),
                    Self::ConditionalExpr(n) => {
                        Some(Expression::Conditional(Box::new((*n).clone())).type_of())
                    }
                    Self::StringExpr(n) => {
                        Some(Expression::StringExpr(Box::new((*n).clone())).type_of())
                    }
                    Self::IntegerLit(n) => Some(Expression::Integer((*n).clone()).type_of()),
                    Self::FloatLit(n) => Some(Expression::Float((*n).clone()).type_of()),
                    Self::StringLit(n) => Some(Expression::StringLit((*n).clone()).type_of()),
                    Self::DurationLit(n) => Some(Expression::Duration((*n).clone()).type_of()),
                    Self::UintLit(n) => Some(Expression::Uint((*n).clone()).type_of()),
                    Self::BooleanLit(n) => Some(Expression::Boolean((*n).clone()).type_of()),
                    Self::DateTimeLit(n) => Some(Expression::DateTime((*n).clone()).type_of()),
                    Self::RegexpLit(n) => Some(Expression::Regexp((*n).clone()).type_of()),
                    _ => None,
                }
            }
        }

        // Utility functions.
        impl<'a> $name<'a> {
            pub(crate) fn from_expr(expr: &'a $($mut)? Expression) -> Self {
                Self::Expr(expr)
            }

            pub(crate) fn reduce_expr(expr: &'a $($mut)? Expression) -> Self {
                match expr {
                    Expression::Identifier(e) => Self::IdentifierExpr(e),
                    Expression::Array(e) => Self::ArrayExpr(e),
                    Expression::Dict(e) => Self::DictExpr(e),
                    Expression::Function(e) => Self::FunctionExpr(e),
                    Expression::Logical(e) => Self::LogicalExpr(e),
                    Expression::Object(e) => Self::ObjectExpr(e),
                    Expression::Member(e) => Self::MemberExpr(e),
                    Expression::Index(e) => Self::IndexExpr(e),
                    Expression::Binary(e) => Self::BinaryExpr(e),
                    Expression::Unary(e) => Self::UnaryExpr(e),
                    Expression::Call(e) => Self::CallExpr(e),
                    Expression::Conditional(e) => Self::ConditionalExpr(e),
                    Expression::StringExpr(e) => Self::StringExpr(e),
                    Expression::Integer(e) => Self::IntegerLit(e),
                    Expression::Float(e) => Self::FloatLit(e),
                    Expression::StringLit(e) => Self::StringLit(e),
                    Expression::Duration(e) => Self::DurationLit(e),
                    Expression::Uint(e) => Self::UintLit(e),
                    Expression::Boolean(e) => Self::BooleanLit(e),
                    Expression::DateTime(e) => Self::DateTimeLit(e),
                    Expression::Regexp(e) => Self::RegexpLit(e),
                    Expression::Label(e) => Self::LabelLit(e),
                    Expression::Error(e) => Self::ErrorExpr(e),
                }
            }
            pub(crate) fn from_stmt(stmt: &'a $($mut)? Statement) -> Self {
                match stmt {
                    Statement::Expr(s) => Self::ExprStmt(s),
                    Statement::Variable(s) => Self::VariableAssgn(s),
                    Statement::Option(s) => Self::OptionStmt(s),
                    Statement::Return(s) => Self::ReturnStmt(s),
                    Statement::TestCase(s) => Self::TestCaseStmt(s),
                    Statement::Builtin(s) => Self::BuiltinStmt(s),
                    Statement::Error(s) => Self::ErrorStmt(s),
                }
            }
            fn from_string_expr_part(sp: &'a $($mut)? StringExprPart) -> Self {
                match sp {
                    StringExprPart::Text(t) => Self::TextPart(t),
                    StringExprPart::Interpolated(e) => Self::InterpolatedPart(e),
                }
            }
            fn from_assignment(a: &'a $($mut)? Assignment) -> Self {
                match a {
                    Assignment::Variable(v) => Self::VariableAssgn(v),
                    Assignment::Member(m) => Self::MemberAssgn(m),
                }
            }
        }

        /// Recursively visits children of a node given a Visitor.
        /// Nodes are visited in depth-first order.
        #[allow(clippy::needless_lifetimes)]
        pub fn $walk<'a, T>(v: &mut T, $($mut)? node: $name<'a>)
        where
            T: ?Sized + $visitor $(<$visitor_lt>)?,
        {
            if v.visit($(&$mut)? node) {
                match &$($mut)? node {
                    $name::Package(n) => {
                        for file in &$($mut)? n.files {
                            $walk(v, $name::File(file));
                        }
                    }
                    $name::File(n) => {
                        if let Some(pkg) = &$($mut)? n.package {
                            $walk(v, $name::PackageClause(pkg));
                        }
                        for imp in &$($mut)? n.imports {
                            $walk(v, $name::ImportDeclaration(imp));
                        }
                        for stmt in &$($mut)? n.body {
                            $walk(v, $name::from_stmt(stmt));
                        }
                    }
                    $name::PackageClause(n) => {
                        $walk(v, $name::Identifier(& $($mut)? n.name));
                    }
                    $name::ImportDeclaration(n) => {
                        if let Some(alias) = &$($mut)? n.alias {
                            $walk(v, $name::Identifier(alias));
                        }
                        $walk(v, $name::StringLit(& $($mut)? n.path));
                    }
                    $name::Identifier(_) => {}
                    $name::Expr(n) => {
                        $walk(v, $name::reduce_expr(n));
                    }
                    $name::IdentifierExpr(_) => {}
                    $name::ArrayExpr(n) => {
                        for element in &$($mut)? n.elements {
                            $walk(v, $name::from_expr(element));
                        }
                    }
                    $name::DictExpr(n) => {
                        for (key, val) in &$($mut)? n.elements {
                            $walk(v, $name::from_expr(key));
                            $walk(v, $name::from_expr(val));
                        }
                    }
                    $name::FunctionExpr(n) => {
                        for param in &$($mut)? n.params {
                            $walk(v, $name::FunctionParameter(param));
                        }
                        $walk(v, $name::Block(& $($mut)? n.body));
                        if let Some(vectorized) = &$($mut)? n.vectorized {
                            $walk(v, $name::FunctionExpr(vectorized));
                        }
                    }
                    $name::FunctionParameter(n) => {
                        $walk(v, $name::Identifier(& $($mut)? n.key));
                        if let Some(def) = &$($mut)? n.default {
                            $walk(v, $name::from_expr(def));
                        }
                    }
                    $name::LogicalExpr(n) => {
                        $walk(v, $name::from_expr(& $($mut)? n.left));
                        $walk(v, $name::from_expr(& $($mut)? n.right));
                    }
                    $name::ObjectExpr(n) => {
                        if let Some(i) = &$($mut)? n.with {
                            $walk(v, $name::IdentifierExpr(i));
                        }
                        for prop in &$($mut)? n.properties {
                            $walk(v, $name::Property(prop));
                        }
                    }
                    $name::MemberExpr(n) => {
                        $walk(v, $name::from_expr(& $($mut)? n.object));
                    }
                    $name::IndexExpr(n) => {
                        $walk(v, $name::from_expr(& $($mut)? n.array));
                        $walk(v, $name::from_expr(& $($mut)? n.index));
                    }
                    $name::BinaryExpr(n) => {
                        $walk(v, $name::from_expr(& $($mut)? n.left));
                        $walk(v, $name::from_expr(& $($mut)? n.right));
                    }
                    $name::UnaryExpr(n) => {
                        $walk(v, $name::from_expr(& $($mut)? n.argument));
                    }
                    $name::CallExpr(n) => {
                        $walk(v, $name::from_expr(& $($mut)? n.callee));
                        if let Some(p) = &$($mut)? n.pipe {
                            $walk(v, $name::from_expr(p));
                        }
                        for arg in &$($mut)? n.arguments {
                            $walk(v, $name::Property(arg));
                        }
                    }
                    $name::ConditionalExpr(n) => {
                        $walk(v, $name::from_expr(& $($mut)? n.test));
                        $walk(v, $name::from_expr(& $($mut)? n.consequent));
                        $walk(v, $name::from_expr(& $($mut)? n.alternate));
                    }
                    $name::StringExpr(n) => {
                        for part in &$($mut)? n.parts {
                            $walk(v, $name::from_string_expr_part(part));
                        }
                    }
                    $name::IntegerLit(_) => {}
                    $name::FloatLit(_) => {}
                    $name::StringLit(_) => {}
                    $name::DurationLit(_) => {}
                    $name::UintLit(_) => {}
                    $name::BooleanLit(_) => {}
                    $name::DateTimeLit(_) => {}
                    $name::RegexpLit(_) => {}
                    $name::LabelLit(_) => {}
                    $name::ExprStmt(n) => {
                        $walk(v, $name::from_expr(& $($mut)? n.expression));
                    }
                    $name::OptionStmt(n) => {
                        $walk(v, $name::from_assignment(& $($mut)? n.assignment));
                    }
                    $name::ReturnStmt(n) => {
                        $walk(v, $name::from_expr(& $($mut)? n.argument));
                    }
                    $name::TestCaseStmt(n) => {
                        $walk(v, $name::Identifier(& $($mut)? n.id));
                        if let Some(e) = & $($mut)? n.extends {
                            $walk(v, $name::StringLit(e));
                        }
                        for stmt in & $($mut)? n.body {
                            $walk(v, $name::from_stmt(stmt));
                        }
                    }
                    $name::BuiltinStmt(n) => {
                        $walk(v, $name::Identifier(& $($mut)? n.id));
                    }
                    $name::ErrorStmt(_) => {}
                    $name::Block(n) => match n {
                        Block::Variable(assgn, next) => {
                            $walk(v, $name::VariableAssgn(assgn));
                            $walk(v, $name::Block(& $($mut)? *next));
                        }
                        Block::Expr(estmt, next) => {
                            $walk(v, $name::ExprStmt(estmt));
                            $walk(v, $name::Block(& $($mut)? *next))
                        }
                        Block::Return(ret_stmt) => $walk(v, $name::ReturnStmt(ret_stmt)),
                    },
                    $name::Property(n) => {
                        $walk(v, $name::Identifier(& $($mut)? n.key));
                        $walk(v, $name::from_expr(& $($mut)? n.value));
                    }
                    $name::TextPart(_) => {}
                    $name::InterpolatedPart(n) => {
                        $walk(v, $name::from_expr(& $($mut)? n.expression));
                    }
                    $name::VariableAssgn(n) => {
                        $walk(v, $name::Identifier(& $($mut)? n.id));
                        $walk(v, $name::from_expr(& $($mut)? n.init));
                    }
                    $name::MemberAssgn(n) => {
                        $walk(v, $name::MemberExpr(& $($mut)? n.member));
                        $walk(v, $name::from_expr(& $($mut)? n.init));
                    }
                    $name::ErrorExpr(_) => (),
                };
            }
            v.done($(&$mut)? node);
        }
    };
}

mod _walk;
mod walk_mut;
pub use _walk::*;
pub use walk_mut::*;

#[cfg(test)]
mod test_utils;

#[cfg(test)]
mod test_node_ids {
    use super::*;
    use crate::semantic::walk::{test_utils::compile, walk, walk_mut};

    fn test_walk(source: &str, want: expect_test::Expect) {
        let mut sem_pkg = compile(source);
        let nodes = {
            let mut nodes = Vec::new();
            walk(
                &mut |n: Node<'_>| nodes.push(format!("{}", n)),
                Node::File(&sem_pkg.files[0]),
            );
            nodes
        };

        want.assert_debug_eq(&nodes);

        let mut_nodes = {
            let mut nodes = Vec::new();
            walk_mut(
                &mut |n: &mut NodeMut<'_>| nodes.push(format!("{}", n)),
                NodeMut::File(&mut sem_pkg.files[0]),
            );
            nodes
        };

        assert_eq!(nodes, mut_nodes);
    }

    #[test]
    fn test_file() {
        test_walk(
            "",
            expect_test::expect![[r#"
            [
                "File",
            ]
        "#]],
        )
    }
    #[test]
    fn test_package_clause() {
        test_walk(
            "package a",
            expect_test::expect![[r#"
            [
                "File",
                "PackageClause",
                "Identifier",
            ]
        "#]],
        )
    }
    #[test]
    fn test_import_declaration() {
        test_walk(
            "import \"a\"",
            expect_test::expect![[r#"
            [
                "File",
                "ImportDeclaration",
                "StringLit",
            ]
        "#]],
        )
    }
    #[test]
    fn test_ident() {
        test_walk(
            "a",
            expect_test::expect![[r#"
            [
                "File",
                "ExprStmt",
                "Expr",
                "IdentifierExpr",
            ]
        "#]],
        )
    }
    #[test]
    fn test_array_expr() {
        test_walk(
            "[1,2,3]",
            expect_test::expect![[r#"
            [
                "File",
                "ExprStmt",
                "Expr",
                "ArrayExpr",
                "Expr",
                "IntegerLit",
                "Expr",
                "IntegerLit",
                "Expr",
                "IntegerLit",
            ]
        "#]],
        )
    }
    #[test]
    fn test_function_expr() {
        test_walk(
            "() => 1",
            expect_test::expect![[r#"
            [
                "File",
                "ExprStmt",
                "Expr",
                "FunctionExpr",
                "Block::Return",
                "ReturnStmt",
                "Expr",
                "IntegerLit",
            ]
        "#]],
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
            expect_test::expect![[r#"
                [
                    "File",
                    "ExprStmt",
                    "Expr",
                    "FunctionExpr",
                    "Block::Variable",
                    "VariableAssgn",
                    "Identifier",
                    "Expr",
                    "IntegerLit",
                    "Block::Variable",
                    "VariableAssgn",
                    "Identifier",
                    "Expr",
                    "BinaryExpr",
                    "Expr",
                    "IntegerLit",
                    "Expr",
                    "IntegerLit",
                    "Block::Expr",
                    "ExprStmt",
                    "Expr",
                    "BinaryExpr",
                    "Expr",
                    "IdentifierExpr",
                    "Expr",
                    "IdentifierExpr",
                    "Block::Return",
                    "ReturnStmt",
                    "Expr",
                    "IdentifierExpr",
                ]
            "#]],
        )
    }
    #[test]
    fn test_function_with_args() {
        test_walk(
            "(a=1) => a",
            expect_test::expect![[r#"
            [
                "File",
                "ExprStmt",
                "Expr",
                "FunctionExpr",
                "FunctionParameter",
                "Identifier",
                "Expr",
                "IntegerLit",
                "Block::Return",
                "ReturnStmt",
                "Expr",
                "IdentifierExpr",
            ]
        "#]],
        )
    }
    #[test]
    fn test_logical_expr() {
        test_walk(
            "true or false",
            expect_test::expect![[r#"
            [
                "File",
                "ExprStmt",
                "Expr",
                "LogicalExpr",
                "Expr",
                "IdentifierExpr",
                "Expr",
                "IdentifierExpr",
            ]
        "#]],
        )
    }
    #[test]
    fn test_object_expr() {
        test_walk(
            "{a:1,\"b\":false}",
            expect_test::expect![[r#"
            [
                "File",
                "ExprStmt",
                "Expr",
                "ObjectExpr",
                "Property",
                "Identifier",
                "Expr",
                "IntegerLit",
                "Property",
                "Identifier",
                "Expr",
                "IdentifierExpr",
            ]
        "#]],
        )
    }
    #[test]
    fn test_member_expr() {
        test_walk(
            "a.b",
            expect_test::expect![[r#"
            [
                "File",
                "ExprStmt",
                "Expr",
                "MemberExpr",
                "Expr",
                "IdentifierExpr",
            ]
        "#]],
        )
    }
    #[test]
    fn test_index_expr() {
        test_walk(
            "a[b]",
            expect_test::expect![[r#"
            [
                "File",
                "ExprStmt",
                "Expr",
                "IndexExpr",
                "Expr",
                "IdentifierExpr",
                "Expr",
                "IdentifierExpr",
            ]
        "#]],
        )
    }
    #[test]
    fn test_binary_expr() {
        test_walk(
            "a+b",
            expect_test::expect![[r#"
            [
                "File",
                "ExprStmt",
                "Expr",
                "BinaryExpr",
                "Expr",
                "IdentifierExpr",
                "Expr",
                "IdentifierExpr",
            ]
        "#]],
        )
    }
    #[test]
    fn test_unary_expr() {
        test_walk(
            "-b",
            expect_test::expect![[r#"
            [
                "File",
                "ExprStmt",
                "Expr",
                "UnaryExpr",
                "Expr",
                "IdentifierExpr",
            ]
        "#]],
        )
    }
    #[test]
    fn test_pipe_expr() {
        test_walk(
            "a|>b()",
            expect_test::expect![[r#"
            [
                "File",
                "ExprStmt",
                "Expr",
                "CallExpr",
                "Expr",
                "IdentifierExpr",
                "Expr",
                "IdentifierExpr",
            ]
        "#]],
        )
    }
    #[test]
    fn test_call_expr() {
        test_walk(
            "b(a:1)",
            expect_test::expect![[r#"
            [
                "File",
                "ExprStmt",
                "Expr",
                "CallExpr",
                "Expr",
                "IdentifierExpr",
                "Property",
                "Identifier",
                "Expr",
                "IntegerLit",
            ]
        "#]],
        )
    }
    #[test]
    fn test_conditional_expr() {
        test_walk(
            "if x then y else z",
            expect_test::expect![[r#"
            [
                "File",
                "ExprStmt",
                "Expr",
                "ConditionalExpr",
                "Expr",
                "IdentifierExpr",
                "Expr",
                "IdentifierExpr",
                "Expr",
                "IdentifierExpr",
            ]
        "#]],
        )
    }
    #[test]
    fn test_string_expr() {
        test_walk(
            "\"hello ${world}\"",
            expect_test::expect![[r#"
            [
                "File",
                "ExprStmt",
                "Expr",
                "StringExpr",
                "TextPart",
                "InterpolatedPart",
                "Expr",
                "IdentifierExpr",
            ]
        "#]],
        )
    }
    #[test]
    fn test_paren_expr() {
        test_walk(
            "(a + b)",
            expect_test::expect![[r#"
            [
                "File",
                "ExprStmt",
                "Expr",
                "BinaryExpr",
                "Expr",
                "IdentifierExpr",
                "Expr",
                "IdentifierExpr",
            ]
        "#]],
        )
    }
    #[test]
    fn test_integer_lit() {
        test_walk(
            "1",
            expect_test::expect![[r#"
            [
                "File",
                "ExprStmt",
                "Expr",
                "IntegerLit",
            ]
        "#]],
        )
    }
    #[test]
    fn test_float_lit() {
        test_walk(
            "1.0",
            expect_test::expect![[r#"
            [
                "File",
                "ExprStmt",
                "Expr",
                "FloatLit",
            ]
        "#]],
        )
    }
    #[test]
    fn test_string_lit() {
        test_walk(
            "\"a\"",
            expect_test::expect![[r#"
            [
                "File",
                "ExprStmt",
                "Expr",
                "StringLit",
            ]
        "#]],
        )
    }
    #[test]
    fn test_duration_lit() {
        test_walk(
            "1m",
            expect_test::expect![[r#"
            [
                "File",
                "ExprStmt",
                "Expr",
                "DurationLit",
            ]
        "#]],
        )
    }
    #[test]
    fn test_datetime_lit() {
        test_walk(
            "2019-01-01T00:00:00Z",
            expect_test::expect![[r#"
            [
                "File",
                "ExprStmt",
                "Expr",
                "DateTimeLit",
            ]
        "#]],
        )
    }
    #[test]
    fn test_regexp_lit() {
        test_walk(
            "/./",
            expect_test::expect![[r#"
            [
                "File",
                "ExprStmt",
                "Expr",
                "RegexpLit",
            ]
        "#]],
        )
    }
    #[test]
    fn test_pipe_lit() {
        test_walk(
            "(a=<-)=>a",
            expect_test::expect![[r#"
            [
                "File",
                "ExprStmt",
                "Expr",
                "FunctionExpr",
                "FunctionParameter",
                "Identifier",
                "Block::Return",
                "ReturnStmt",
                "Expr",
                "IdentifierExpr",
            ]
        "#]],
        )
    }

    #[test]
    fn test_option_stmt() {
        test_walk(
            "option a = b",
            expect_test::expect![[r#"
            [
                "File",
                "OptionStmt",
                "VariableAssgn",
                "Identifier",
                "Expr",
                "IdentifierExpr",
            ]
        "#]],
        )
    }
    #[test]
    fn test_return_stmt() {
        // This is quite tricky, even if there is an explicit ReturnStmt,
        // `analyze` returns a `Block::Return` when inside of a function body.
        test_walk(
            "() => {return 1}",
            expect_test::expect![[r#"
            [
                "File",
                "ExprStmt",
                "Expr",
                "FunctionExpr",
                "Block::Return",
                "ReturnStmt",
                "Expr",
                "IntegerLit",
            ]
        "#]],
        )
    }
    #[test]
    fn test_builtin_stmt() {
        test_walk(
            "builtin a : int",
            expect_test::expect![[r#"
            [
                "File",
                "BuiltinStmt",
                "Identifier",
            ]
        "#]],
        )
    }
    #[test]
    fn test_variable_assgn() {
        test_walk(
            "a = b",
            expect_test::expect![[r#"
            [
                "File",
                "VariableAssgn",
                "Identifier",
                "Expr",
                "IdentifierExpr",
            ]
        "#]],
        )
    }
    #[test]
    fn test_member_assgn() {
        test_walk(
            "option a.b = c",
            expect_test::expect![[r#"
            [
                "File",
                "OptionStmt",
                "MemberAssgn",
                "MemberExpr",
                "Expr",
                "IdentifierExpr",
                "Expr",
                "IdentifierExpr",
            ]
        "#]],
        )
    }
}
