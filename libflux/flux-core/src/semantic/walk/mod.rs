//! Walking the semantic graph.

macro_rules! mk_node {
    ($(#[$attr:meta])* $name: ident $($mut: tt)?) => {

        $(#[$attr])*
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
            DateTimeLit(&'a $($mut)? DateTimeLit), RegexpLit(&'a $($mut)? RegexpLit),

            // Statements.
            ExprStmt(&'a $($mut)? ExprStmt),
            OptionStmt(&'a $($mut)? OptionStmt),
            ReturnStmt(&'a $($mut)? ReturnStmt),
            TestStmt(&'a $($mut)? TestStmt),
            TestCaseStmt(&'a $($mut)? TestCaseStmt),
            BuiltinStmt(&'a $($mut)? BuiltinStmt),

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
                    Self::ExprStmt(_) => write!(f, "ExprStmt"),
                    Self::OptionStmt(_) => write!(f, "OptionStmt"),
                    Self::ReturnStmt(_) => write!(f, "ReturnStmt"),
                    Self::TestStmt(_) => write!(f, "TestStmt"),
                    Self::TestCaseStmt(_) => write!(f, "TestCaseStmt"),
                    Self::BuiltinStmt(_) => write!(f, "BuiltinStmt"),
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
                    Self::ExprStmt(n) => &n.loc,
                    Self::OptionStmt(n) => &n.loc,
                    Self::ReturnStmt(n) => &n.loc,
                    Self::TestStmt(n) => &n.loc,
                    Self::TestCaseStmt(n) => &n.loc,
                    Self::BuiltinStmt(n) => &n.loc,
                    Self::Block(n) => n.loc(),
                    Self::Property(n) => &n.loc,
                    Self::TextPart(n) => &n.loc,
                    Self::InterpolatedPart(n) => &n.loc,
                    Self::VariableAssgn(n) => &n.loc,
                    Self::MemberAssgn(n) => &n.loc,
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
                }
            }
            pub(crate) fn from_stmt(stmt: &'a $($mut)? Statement) -> Self {
                match stmt {
                    Statement::Expr(s) => Self::ExprStmt(s),
                    Statement::Variable(s) => Self::VariableAssgn(s),
                    Statement::Option(s) => Self::OptionStmt(s),
                    Statement::Return(s) => Self::ReturnStmt(s),
                    Statement::Test(s) => Self::TestStmt(s),
                    Statement::TestCase(s) => Self::TestCaseStmt(s),
                    Statement::Builtin(s) => Self::BuiltinStmt(s),
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
    };
}

mod _walk;
mod walk_mut;
pub use _walk::*;
pub use walk_mut::*;

#[cfg(test)]
mod test_utils;
