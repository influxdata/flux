//! Walking the AST.

#[cfg(test)]
mod tests;

use derive_more::Display;

use crate::ast::*;

macro_rules! mk_node {
    (
        $(#[$attr:meta])*
        $name: ident
        $visitor: ident $(<$visitor_lt: lifetime>)?
        $walk: ident
        $($mut: tt)?
    ) => {
        /// Node represents any structure that can appear in the AST.
        $(#[$attr])*
        #[derive(Debug, Display)]
        #[allow(missing_docs)]
        pub enum $name<'a> {
            #[display(fmt = "Package")]
            Package(&'a $($mut)? Package),
            #[display(fmt = "File")]
            File(&'a $($mut)? File),
            #[display(fmt = "PackageClause")]
            PackageClause(&'a $($mut)? PackageClause),
            #[display(fmt = "ImportDeclaration")]
            ImportDeclaration(&'a $($mut)? ImportDeclaration),

            // Expressions
            #[display(fmt = "Identifier")]
            Identifier(&'a $($mut)? Identifier),

            #[display(fmt = "ArrayExpr")]
            ArrayExpr(&'a $($mut)? ArrayExpr),
            #[display(fmt = "DictExpr")]
            DictExpr(&'a $($mut)? DictExpr),
            #[display(fmt = "FunctionExpr")]
            FunctionExpr(&'a $($mut)? FunctionExpr),
            #[display(fmt = "LogicalExpr")]
            LogicalExpr(&'a $($mut)? LogicalExpr),
            #[display(fmt = "ObjectExpr")]
            ObjectExpr(&'a $($mut)? ObjectExpr),
            #[display(fmt = "MemberExpr")]
            MemberExpr(&'a $($mut)? MemberExpr),
            #[display(fmt = "IndexExpr")]
            IndexExpr(&'a $($mut)? IndexExpr),
            #[display(fmt = "BinaryExpr")]
            BinaryExpr(&'a $($mut)? BinaryExpr),
            #[display(fmt = "UnaryExpr")]
            UnaryExpr(&'a $($mut)? UnaryExpr),
            #[display(fmt = "PipeExpr")]
            PipeExpr(&'a $($mut)? PipeExpr),
            #[display(fmt = "CallExpr")]
            CallExpr(&'a $($mut)? CallExpr),
            #[display(fmt = "ConditionalExpr")]
            ConditionalExpr(&'a $($mut)? ConditionalExpr),
            #[display(fmt = "StringExpr")]
            StringExpr(&'a $($mut)? StringExpr),
            #[display(fmt = "ParenExpr")]
            ParenExpr(&'a $($mut)? ParenExpr),

            #[display(fmt = "IntegerLit")]
            IntegerLit(&'a $($mut)? IntegerLit),
            #[display(fmt = "FloatLit")]
            FloatLit(&'a $($mut)? FloatLit),
            #[display(fmt = "StringLit")]
            StringLit(&'a $($mut)? StringLit),
            #[display(fmt = "DurationLit")]
            DurationLit(&'a $($mut)? DurationLit),
            #[display(fmt = "UintLit")]
            UintLit(&'a $($mut)? UintLit),
            #[display(fmt = "BooleanLit")]
            BooleanLit(&'a $($mut)? BooleanLit),
            #[display(fmt = "DateTimeLit")]
            DateTimeLit(&'a $($mut)? DateTimeLit),
            #[display(fmt = "RegexpLit")]
            RegexpLit(&'a $($mut)? RegexpLit),
            #[display(fmt = "PipeLit")]
            PipeLit(&'a $($mut)? PipeLit),

            #[display(fmt = "BadExpr")]
            BadExpr(&'a $($mut)? BadExpr),

            // Statements
            #[display(fmt = "ExprStmt")]
            ExprStmt(&'a $($mut)? ExprStmt),
            #[display(fmt = "OptionStmt")]
            OptionStmt(&'a $($mut)? OptionStmt),
            #[display(fmt = "ReturnStmt")]
            ReturnStmt(&'a $($mut)? ReturnStmt),
            #[display(fmt = "BadStmt")]
            BadStmt(&'a $($mut)? BadStmt),
            #[display(fmt = "TestStmt")]
            TestStmt(&'a $($mut)? TestStmt),
            #[display(fmt = "TestCaseStmt")]
            TestCaseStmt(&'a $($mut)? TestCaseStmt),
            #[display(fmt = "BuiltinStmt")]
            BuiltinStmt(&'a $($mut)? BuiltinStmt),

            // FunctionBlock
            #[display(fmt = "Block")]
            Block(&'a $($mut)? Block),

            // Property
            #[display(fmt = "Property")]
            Property(&'a $($mut)? Property),

            // StringExprPart
            #[display(fmt = "TextPart")]
            TextPart(&'a $($mut)? TextPart),
            #[display(fmt = "InterpolatedPart")]
            InterpolatedPart(&'a $($mut)? InterpolatedPart),

            // Assignment
            #[display(fmt = "VariableAssgn")]
            VariableAssgn(&'a $($mut)? VariableAssgn),
            #[display(fmt = "MemberAssgn")]
            MemberAssgn(&'a $($mut)? MemberAssgn),

            #[display(fmt = "TypeExpression")]
            TypeExpression(&'a $($mut)? TypeExpression),
            #[display(fmt = "MonoType")]
            MonoType(&'a $($mut)? MonoType),
            #[display(fmt = "PropertyType")]
            PropertyType(&'a $($mut)? PropertyType),
            #[display(fmt = "ParameterType")]
            ParameterType(&'a $($mut)? ParameterType),
            #[display(fmt = "TypeConstraint")]
            TypeConstraint(&'a $($mut)? TypeConstraint),
        }

        impl<'a> $name<'a> {
            #[allow(missing_docs)]
            pub fn base(&self) -> &BaseNode {
                match self {
                    Self::Package(n) => &n.base,
                    Self::File(n) => &n.base,
                    Self::PackageClause(n) => &n.base,
                    Self::ImportDeclaration(n) => &n.base,
                    Self::Identifier(n) => &n.base,
                    Self::ArrayExpr(n) => &n.base,
                    Self::DictExpr(n) => &n.base,
                    Self::FunctionExpr(n) => &n.base,
                    Self::LogicalExpr(n) => &n.base,
                    Self::ObjectExpr(n) => &n.base,
                    Self::MemberExpr(n) => &n.base,
                    Self::IndexExpr(n) => &n.base,
                    Self::BinaryExpr(n) => &n.base,
                    Self::UnaryExpr(n) => &n.base,
                    Self::PipeExpr(n) => &n.base,
                    Self::CallExpr(n) => &n.base,
                    Self::ConditionalExpr(n) => &n.base,
                    Self::StringExpr(n) => &n.base,
                    Self::ParenExpr(n) => &n.base,
                    Self::IntegerLit(n) => &n.base,
                    Self::FloatLit(n) => &n.base,
                    Self::StringLit(n) => &n.base,
                    Self::DurationLit(n) => &n.base,
                    Self::UintLit(n) => &n.base,
                    Self::BooleanLit(n) => &n.base,
                    Self::DateTimeLit(n) => &n.base,
                    Self::RegexpLit(n) => &n.base,
                    Self::PipeLit(n) => &n.base,
                    Self::BadExpr(n) => &n.base,
                    Self::ExprStmt(n) => &n.base,
                    Self::OptionStmt(n) => &n.base,
                    Self::ReturnStmt(n) => &n.base,
                    Self::BadStmt(n) => &n.base,
                    Self::TestStmt(n) => &n.base,
                    Self::TestCaseStmt(n) => &n.base,
                    Self::BuiltinStmt(n) => &n.base,
                    Self::Block(n) => &n.base,
                    Self::Property(n) => &n.base,
                    Self::TextPart(n) => &n.base,
                    Self::InterpolatedPart(n) => &n.base,
                    Self::VariableAssgn(n) => &n.base,
                    Self::MemberAssgn(n) => &n.base,
                    Self::TypeExpression(n) => &n.base,
                    Self::MonoType(n) => n.base(),
                    Self::PropertyType(n) => &n.base,
                    Self::ParameterType(n) => n.base(),
                    Self::TypeConstraint(n) => &n.base,
                }
            }
        }

        impl<'a> $name<'a> {
            #[allow(missing_docs)]
            pub fn from_expr(expr: &'a $($mut)? Expression) -> Self {
                match expr {
                    Expression::Identifier(e) => Self::Identifier(e),
                    Expression::Array(e) => Self::ArrayExpr(e),
                    Expression::Dict(e) => Self::DictExpr(e),
                    Expression::Function(e) => Self::FunctionExpr(e),
                    Expression::Logical(e) => Self::LogicalExpr(e),
                    Expression::Object(e) => Self::ObjectExpr(e),
                    Expression::Member(e) => Self::MemberExpr(e),
                    Expression::Index(e) => Self::IndexExpr(e),
                    Expression::Binary(e) => Self::BinaryExpr(e),
                    Expression::Unary(e) => Self::UnaryExpr(e),
                    Expression::PipeExpr(e) => Self::PipeExpr(e),
                    Expression::Call(e) => Self::CallExpr(e),
                    Expression::Conditional(e) => Self::ConditionalExpr(e),
                    Expression::StringExpr(e) => Self::StringExpr(e),
                    Expression::Paren(e) => Self::ParenExpr(e),
                    Expression::Integer(e) => Self::IntegerLit(e),
                    Expression::Float(e) => Self::FloatLit(e),
                    Expression::StringLit(e) => Self::StringLit(e),
                    Expression::Duration(e) => Self::DurationLit(e),
                    Expression::Uint(e) => Self::UintLit(e),
                    Expression::Boolean(e) => Self::BooleanLit(e),
                    Expression::DateTime(e) => Self::DateTimeLit(e),
                    Expression::Regexp(e) => Self::RegexpLit(e),
                    Expression::PipeLit(e) => Self::PipeLit(e),
                    Expression::Bad(e) => Self::BadExpr(e),
                }
            }
            #[allow(missing_docs)]
            pub fn from_stmt(stmt: &'a $($mut)? Statement) -> Self {
                match stmt {
                    Statement::Expr(s) => Self::ExprStmt(s),
                    Statement::Variable(s) => Self::VariableAssgn(s),
                    Statement::Option(s) => Self::OptionStmt(s),
                    Statement::Return(s) => Self::ReturnStmt(s),
                    Statement::Bad(s) => Self::BadStmt(s),
                    Statement::Test(s) => Self::TestStmt(s),
                    Statement::TestCase(s) => Self::TestCaseStmt(s),
                    Statement::Builtin(s) => Self::BuiltinStmt(s),
                }
            }
            fn from_function_body(fb: &'a $($mut)? FunctionBody) -> Self {
                match fb {
                    FunctionBody::Block(b) => Self::Block(b),
                    FunctionBody::Expr(e) => Self::from_expr(e),
                }
            }
            fn from_property_key(pk: &'a $($mut)? PropertyKey) -> Self {
                match pk {
                    PropertyKey::Identifier(i) => Self::Identifier(i),
                    PropertyKey::StringLit(s) => Self::StringLit(s),
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

        /// Walk recursively visits children of a node.
        /// Nodes are visited in depth-first order.
        pub fn $walk<'a, T>(v: &mut T, $($mut)? node: $name<'a>)
        where
            T: $visitor $(<$visitor_lt>)?,
        {
            if v.visit($(&$mut)? node) {
                match & $($mut)? node {
                    $name::Package(n) => {
                        for file in & $($mut)? n.files {
                            $walk(v, $name::File(file));
                        }
                    }
                    $name::File(n) => {
                        if let Some(pkg) = & $($mut)? n.package {
                            $walk(v, $name::PackageClause(pkg));
                        }
                        for imp in & $($mut)? n.imports {
                            $walk(v, $name::ImportDeclaration(imp));
                        }
                        for stmt in & $($mut)? n.body {
                            $walk(v, $name::from_stmt(stmt));
                        }
                    }
                    $name::PackageClause(n) => {
                        $walk(v, $name::Identifier(& $($mut)? n.name));
                    }
                    $name::ImportDeclaration(n) => {
                        if let Some(alias) = & $($mut)? n.alias {
                            $walk(v, $name::Identifier(alias));
                        }
                        $walk(v, $name::StringLit(& $($mut)? n.path));
                    }
                    $name::Identifier(_) => {}
                    $name::ArrayExpr(n) => {
                        for element in & $($mut)? n.elements {
                            $walk(v, $name::from_expr(& $($mut)? element.expression));
                        }
                    }
                    $name::DictExpr(n) => {
                        for element in & $($mut)? n.elements {
                            $walk(v, $name::from_expr(& $($mut)? element.key));
                            $walk(v, $name::from_expr(& $($mut)? element.val));
                        }
                    }
                    $name::FunctionExpr(n) => {
                        for param in & $($mut)? n.params {
                            $walk(v, $name::Property(param));
                        }
                        $walk(v, $name::from_function_body(& $($mut)? n.body));
                    }
                    $name::LogicalExpr(n) => {
                        $walk(v, $name::from_expr(& $($mut)? n.left));
                        $walk(v, $name::from_expr(& $($mut)? n.right));
                    }
                    $name::ObjectExpr(n) => {
                        if let Some(ws) = & $($mut)? n.with {
                            $walk(v, $name::Identifier(& $($mut)? ws.source));
                        }
                        for prop in & $($mut)? n.properties {
                            $walk(v, $name::Property(prop));
                        }
                    }
                    $name::MemberExpr(n) => {
                        $walk(v, $name::from_expr(& $($mut)? n.object));
                        $walk(v, $name::from_property_key(& $($mut)? n.property));
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
                    $name::PipeExpr(n) => {
                        $walk(v, $name::from_expr(& $($mut)? n.argument));
                        $walk(v, $name::CallExpr(& $($mut)? n.call));
                    }
                    $name::CallExpr(n) => {
                        $walk(v, $name::from_expr(& $($mut)? n.callee));
                        for arg in & $($mut)? n.arguments {
                            $walk(v, $name::from_expr(arg));
                        }
                    }
                    $name::ConditionalExpr(n) => {
                        $walk(v, $name::from_expr(& $($mut)? n.test));
                        $walk(v, $name::from_expr(& $($mut)? n.consequent));
                        $walk(v, $name::from_expr(& $($mut)? n.alternate));
                    }
                    $name::StringExpr(n) => {
                        for part in & $($mut)? n.parts {
                            $walk(v, $name::from_string_expr_part(part));
                        }
                    }
                    $name::ParenExpr(n) => {
                        $walk(v, $name::from_expr(& $($mut)? n.expression));
                    }
                    $name::IntegerLit(_) => {}
                    $name::FloatLit(_) => {}
                    $name::StringLit(_) => {}
                    $name::DurationLit(_) => {}
                    $name::UintLit(_) => {}
                    $name::BooleanLit(_) => {}
                    $name::DateTimeLit(_) => {}
                    $name::RegexpLit(_) => {}
                    $name::PipeLit(_) => {}
                    $name::BadExpr(n) => {
                        if let Some(e) = & $($mut)? n.expression {
                            $walk(v, $name::from_expr(e));
                        }
                    }
                    $name::ExprStmt(n) => {
                        $walk(v, $name::from_expr(& $($mut)? n.expression));
                    }
                    $name::OptionStmt(n) => {
                        $walk(v, $name::from_assignment(& $($mut)? n.assignment));
                    }
                    $name::ReturnStmt(n) => {
                        $walk(v, $name::from_expr(& $($mut)? n.argument));
                    }
                    $name::BadStmt(_) => {}
                    $name::TestStmt(n) => {
                        $walk(v, $name::VariableAssgn(& $($mut)? n.assignment));
                    }
                    $name::TestCaseStmt(n) => {
                        $walk(v, $name::Identifier(& $($mut)? n.id));
                        $walk(v, $name::Block(& $($mut)? n.block));
                    }
                    $name::BuiltinStmt(n) => {
                        $walk(v, $name::Identifier(& $($mut)? n.id));
                        $walk(v, $name::TypeExpression(& $($mut)? n.ty));
                    }
                    $name::Block(n) => {
                        for s in & $($mut)? n.body {
                            $walk(v, $name::from_stmt(s));
                        }
                    }
                    $name::Property(n) => {
                        $walk(v, $name::from_property_key(& $($mut)? n.key));
                        if let Some(value) = & $($mut)? n.value {
                            $walk(v, $name::from_expr(value));
                        }
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
                    $name::TypeExpression(n) => {
                        $walk(v, $name::MonoType(& $($mut)? n.monotype));
                        for cons in & $($mut)? n.constraints {
                            $walk(v, $name::TypeConstraint(cons));
                        }
                    }
                    $name::MonoType(n) => match n {
                        MonoType::Tvar(_) => (),
                        MonoType::Basic(_) => (),
                        MonoType::Array(a) => $walk(v, $name::MonoType(& $($mut)? a.element)),
                        MonoType::Stream(a) => $walk(v, $name::MonoType(& $($mut)? a.element)),
                        MonoType::Dict(d) => {
                            $walk(v, $name::MonoType(& $($mut)? d.key));
                            $walk(v, $name::MonoType(& $($mut)? d.val));
                        }
                        MonoType::Record(r) => {
                            if let Some(tvar) = & $($mut)? r.tvar {
                                $walk(v, $name::Identifier(tvar));
                            }

                            for property in & $($mut)? r.properties {
                                $walk(v, $name::PropertyType(property));
                            }
                        }
                        MonoType::Function(f) => {
                            for param in & $($mut)? f.parameters {
                                $walk(v, $name::ParameterType(param));
                            }

                            $walk(v, $name::MonoType(& $($mut)? f.monotype));
                        }
                        MonoType::Label(lit) => $walk(v, $name::StringLit(lit)),
                    },
                    $name::PropertyType(n) => {
                        $walk(v, $name::from_property_key(& $($mut)? n.name));
                        $walk(v, $name::MonoType(& $($mut)? n.monotype));
                    }
                    $name::ParameterType(n) => match n {
                        ParameterType::Required { name, monotype, .. } => {
                            $walk(v, $name::Identifier(name));
                            $walk(v, $name::MonoType(monotype));
                        }
                        ParameterType::Optional {
                            name,
                            monotype,
                            default,
                            ..
                        } => {
                            $walk(v, $name::Identifier(name));
                            $walk(v, $name::MonoType(monotype));
                            if let Some(default) = default {
                                $walk(v, $name::StringLit(default));
                            }
                        }
                        ParameterType::Pipe { name, monotype, .. } => {
                            if let Some(name) = name {
                                $walk(v, $name::Identifier(name));
                            }
                            $walk(v, $name::MonoType(monotype));
                        }
                    },
                    $name::TypeConstraint(n) => {
                        $walk(v, $name::Identifier(& $($mut)? n.tvar));
                        for id in & $($mut)? n.kinds {
                            $walk(v, $name::Identifier(id));
                        }
                    }
                }
            }

            v.done($(&$mut)? node)
        }
    };
}

mk_node!(
    #[derive(Clone, Copy)]
    Node
    Visitor <'a>
    walk
);
mk_node!(
    NodeMut
    VisitorMut
    walk_mut
    mut
);

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

impl<'a, F> Visitor<'a> for F
where
    F: FnMut(Node<'a>),
{
    fn visit(&mut self, node: Node<'a>) -> bool {
        self(node);
        true
    }
}

/// VisitorMut defines a visitor pattern for walking the AST.
///
/// When used with the walk function, Visit will be called for every node
/// in depth-first order. After all children for a NodeMut have been visted,
/// Done is called on that NodeMut to signal that we are done with that NodeMut.
///
/// If Visit returns None, walk will not recurse on the children.
///
/// Note: the Rc in visit and done is to allow for multiple ownership of a node, i.e.
///       a visitor can own a node as well as the walk funciton. This allows
///       for nodes to persist outside the scope of the walk function and to
///       be cleaned up once all owners have let go of the reference.
///
/// See example with `FuncVisitorMut` below in this file.
pub trait VisitorMut: Sized {
    /// Visit is called for a node.
    /// When the `VisitorMut` is used in [`walk()`], the boolean value returned
    /// is used to continue walking (`true`) or stop (`false`).
    fn visit(&mut self, node: &mut NodeMut<'_>) -> bool;
    /// Done is called for a node once it has been visited along with all of its children.
    fn done(&mut self, _: &mut NodeMut<'_>) {} // default is to do nothing
}

impl<F> VisitorMut for F
where
    F: for<'a> FnMut(&mut NodeMut<'a>),
{
    fn visit(&mut self, node: &mut NodeMut<'_>) -> bool {
        self(node);
        true
    }
}
