#[cfg(test)]
mod tests;

use crate::ast::*;
use std::cell::{RefCell, RefMut};
use std::rc::Rc;

// Node represents any structure that can appear in the AST
#[derive(Debug)]
pub enum Node<'a> {
    Package(&'a Package),
    File(&'a File),
    PackageClause(&'a PackageClause),
    ImportDeclaration(&'a ImportDeclaration),

    // Expressions
    Identifier(&'a Identifier),

    ArrayExpr(&'a ArrayExpr),
    FunctionExpr(&'a FunctionExpr),
    LogicalExpr(&'a LogicalExpr),
    ObjectExpr(&'a ObjectExpr),
    MemberExpr(&'a MemberExpr),
    IndexExpr(&'a IndexExpr),
    BinaryExpr(&'a BinaryExpr),
    UnaryExpr(&'a UnaryExpr),
    PipeExpr(&'a PipeExpr),
    CallExpr(&'a CallExpr),
    ConditionalExpr(&'a ConditionalExpr),
    StringExpr(&'a StringExpr),
    ParenExpr(&'a ParenExpr),

    IntegerLit(&'a IntegerLit),
    FloatLit(&'a FloatLit),
    StringLit(&'a StringLit),
    DurationLit(&'a DurationLit),
    UintLit(&'a UintLit),
    BooleanLit(&'a BooleanLit),
    DateTimeLit(&'a DateTimeLit),
    RegexpLit(&'a RegexpLit),
    PipeLit(&'a PipeLit),

    BadExpr(&'a BadExpr),

    // Statements
    ExprStmt(&'a ExprStmt),
    OptionStmt(&'a OptionStmt),
    ReturnStmt(&'a ReturnStmt),
    BadStmt(&'a BadStmt),
    TestStmt(&'a TestStmt),
    BuiltinStmt(&'a BuiltinStmt),

    // FunctionBlock
    Block(&'a Block),

    // Property
    Property(&'a Property),

    // StringExprPart
    TextPart(&'a TextPart),
    InterpolatedPart(&'a InterpolatedPart),

    // Assignment
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
            Node::ArrayExpr(_) => write!(f, "ArrayExpr"),
            Node::FunctionExpr(_) => write!(f, "FunctionExpr"),
            Node::LogicalExpr(_) => write!(f, "LogicalExpr"),
            Node::ObjectExpr(_) => write!(f, "ObjectExpr"),
            Node::MemberExpr(_) => write!(f, "MemberExpr"),
            Node::IndexExpr(_) => write!(f, "IndexExpr"),
            Node::BinaryExpr(_) => write!(f, "BinaryExpr"),
            Node::UnaryExpr(_) => write!(f, "UnaryExpr"),
            Node::PipeExpr(_) => write!(f, "PipeExpr"),
            Node::CallExpr(_) => write!(f, "CallExpr"),
            Node::ConditionalExpr(_) => write!(f, "ConditionalExpr"),
            Node::StringExpr(_) => write!(f, "StringExpr"),
            Node::ParenExpr(_) => write!(f, "ParenExpr"),
            Node::IntegerLit(_) => write!(f, "IntegerLit"),
            Node::FloatLit(_) => write!(f, "FloatLit"),
            Node::StringLit(_) => write!(f, "StringLit"),
            Node::DurationLit(_) => write!(f, "DurationLit"),
            Node::UintLit(_) => write!(f, "UintLit"),
            Node::BooleanLit(_) => write!(f, "BooleanLit"),
            Node::DateTimeLit(_) => write!(f, "DateTimeLit"),
            Node::RegexpLit(_) => write!(f, "RegexpLit"),
            Node::PipeLit(_) => write!(f, "PipeLit"),
            Node::BadExpr(_) => write!(f, "BadExpr"),
            Node::ExprStmt(_) => write!(f, "ExprStmt"),
            Node::OptionStmt(_) => write!(f, "OptionStmt"),
            Node::ReturnStmt(_) => write!(f, "ReturnStmt"),
            Node::BadStmt(_) => write!(f, "BadStmt"),
            Node::TestStmt(_) => write!(f, "TestStmt"),
            Node::BuiltinStmt(_) => write!(f, "BuiltinStmt"),
            Node::Block(_) => write!(f, "Block"),
            Node::Property(_) => write!(f, "Property"),
            Node::TextPart(_) => write!(f, "TextPart"),
            Node::InterpolatedPart(_) => write!(f, "InterpolatedPart"),
            Node::VariableAssgn(_) => write!(f, "VariableAssgn"),
            Node::MemberAssgn(_) => write!(f, "MemberAssgn"),
        }
    }
}
impl<'a> Node<'a> {
    pub fn base(&self) -> &BaseNode {
        match self {
            Node::Package(n) => &n.base,
            Node::File(n) => &n.base,
            Node::PackageClause(n) => &n.base,
            Node::ImportDeclaration(n) => &n.base,
            Node::Identifier(n) => &n.base,
            Node::ArrayExpr(n) => &n.base,
            Node::FunctionExpr(n) => &n.base,
            Node::LogicalExpr(n) => &n.base,
            Node::ObjectExpr(n) => &n.base,
            Node::MemberExpr(n) => &n.base,
            Node::IndexExpr(n) => &n.base,
            Node::BinaryExpr(n) => &n.base,
            Node::UnaryExpr(n) => &n.base,
            Node::PipeExpr(n) => &n.base,
            Node::CallExpr(n) => &n.base,
            Node::ConditionalExpr(n) => &n.base,
            Node::StringExpr(n) => &n.base,
            Node::ParenExpr(n) => &n.base,
            Node::IntegerLit(n) => &n.base,
            Node::FloatLit(n) => &n.base,
            Node::StringLit(n) => &n.base,
            Node::DurationLit(n) => &n.base,
            Node::UintLit(n) => &n.base,
            Node::BooleanLit(n) => &n.base,
            Node::DateTimeLit(n) => &n.base,
            Node::RegexpLit(n) => &n.base,
            Node::PipeLit(n) => &n.base,
            Node::BadExpr(n) => &n.base,
            Node::ExprStmt(n) => &n.base,
            Node::OptionStmt(n) => &n.base,
            Node::ReturnStmt(n) => &n.base,
            Node::BadStmt(n) => &n.base,
            Node::TestStmt(n) => &n.base,
            Node::BuiltinStmt(n) => &n.base,
            Node::Block(n) => &n.base,
            Node::Property(n) => &n.base,
            Node::TextPart(n) => &n.base,
            Node::InterpolatedPart(n) => &n.base,
            Node::VariableAssgn(n) => &n.base,
            Node::MemberAssgn(n) => &n.base,
        }
    }
}

impl<'a> Node<'a> {
    fn from_expr(expr: &'a Expression) -> Node {
        match expr {
            Expression::Identifier(e) => Node::Identifier(e),
            Expression::Array(e) => Node::ArrayExpr(e),
            Expression::Function(e) => Node::FunctionExpr(e),
            Expression::Logical(e) => Node::LogicalExpr(e),
            Expression::Object(e) => Node::ObjectExpr(e),
            Expression::Member(e) => Node::MemberExpr(e),
            Expression::Index(e) => Node::IndexExpr(e),
            Expression::Binary(e) => Node::BinaryExpr(e),
            Expression::Unary(e) => Node::UnaryExpr(e),
            Expression::PipeExpr(e) => Node::PipeExpr(e),
            Expression::Call(e) => Node::CallExpr(e),
            Expression::Conditional(e) => Node::ConditionalExpr(e),
            Expression::StringExpr(e) => Node::StringExpr(e),
            Expression::Paren(e) => Node::ParenExpr(e),
            Expression::Integer(e) => Node::IntegerLit(e),
            Expression::Float(e) => Node::FloatLit(e),
            Expression::StringLit(e) => Node::StringLit(e),
            Expression::Duration(e) => Node::DurationLit(e),
            Expression::Uint(e) => Node::UintLit(e),
            Expression::Boolean(e) => Node::BooleanLit(e),
            Expression::DateTime(e) => Node::DateTimeLit(e),
            Expression::Regexp(e) => Node::RegexpLit(e),
            Expression::PipeLit(e) => Node::PipeLit(e),
            Expression::Bad(e) => Node::BadExpr(e),
        }
    }
    fn from_stmt(stmt: &Statement) -> Node {
        match stmt {
            Statement::Expr(s) => Node::ExprStmt(s),
            Statement::Variable(s) => Node::VariableAssgn(s),
            Statement::Option(s) => Node::OptionStmt(s),
            Statement::Return(s) => Node::ReturnStmt(s),
            Statement::Bad(s) => Node::BadStmt(s),
            Statement::Test(s) => Node::TestStmt(s),
            Statement::Builtin(s) => Node::BuiltinStmt(s),
        }
    }
    fn from_function_body(fb: &FunctionBody) -> Node {
        match fb {
            FunctionBody::Block(b) => Node::Block(b),
            FunctionBody::Expr(e) => Node::from_expr(e),
        }
    }
    fn from_property_key(pk: &PropertyKey) -> Node {
        match pk {
            PropertyKey::Identifier(i) => Node::Identifier(i),
            PropertyKey::StringLit(s) => Node::StringLit(s),
        }
    }
    fn from_string_expr_part(sp: &StringExprPart) -> Node {
        match sp {
            StringExprPart::Text(t) => Node::TextPart(t),
            StringExprPart::Interpolated(e) => Node::InterpolatedPart(e),
        }
    }
    fn from_assignment(a: &Assignment) -> Node {
        match a {
            Assignment::Variable(v) => Node::VariableAssgn(v),
            Assignment::Member(m) => Node::MemberAssgn(m),
        }
    }
}

// Visitor defines a visitor pattern for walking the AST
//
// When used with the walk function, Visit will be called for every node
// in depth-first order. After all children for a Node have been visted,
// Done is called on that Node to signal that we are done with that Node.
//
// If Visit returns None, walk will not recurse on the children.
pub trait Visitor: Sized {
    // Visit is called for a node.
    // The returned visitor will be used to walk children of the node.
    // If visit returns None, walk will not recurse on the children.
    fn visit(&self, node: &Node) -> Option<Self>;
    // Done is called for a node once it has been visited along with all of its children.
    fn done(&self, _: &Node) {} // default is to do nothing
}

// Walk recursively visits children of a node.
// Nodes are visited in depth-first order.
pub fn walk<T>(v: &T, node: Node)
where
    T: Visitor,
{
    if let Some(w) = v.visit(&node) {
        match node {
            Node::Package(n) => {
                for file in n.files.iter() {
                    walk(&w, Node::File(&file));
                }
            }
            Node::File(n) => {
                if let Some(pkg) = &n.package {
                    walk(&w, Node::PackageClause(pkg));
                }
                for imp in n.imports.iter() {
                    walk(&w, Node::ImportDeclaration(imp));
                }
                for stmt in n.body.iter() {
                    walk(&w, Node::from_stmt(stmt));
                }
            }
            Node::PackageClause(n) => {
                walk(&w, Node::Identifier(&n.name));
            }
            Node::ImportDeclaration(n) => {
                if let Some(alias) = &n.alias {
                    walk(&w, Node::Identifier(alias));
                }
                walk(&w, Node::StringLit(&n.path));
            }
            Node::Identifier(_) => {}
            Node::ArrayExpr(n) => {
                for element in n.elements.iter() {
                    walk(&w, Node::from_expr(&element));
                }
            }
            Node::FunctionExpr(n) => {
                for param in n.params.iter() {
                    walk(&w, Node::Property(&param));
                }
                walk(&w, Node::from_function_body(&n.body))
            }
            Node::LogicalExpr(n) => {
                walk(&w, Node::from_expr(&n.left));
                walk(&w, Node::from_expr(&n.right));
            }
            Node::ObjectExpr(n) => {
                if let Some(i) = &n.with {
                    walk(&w, Node::Identifier(i));
                }
                for prop in n.properties.iter() {
                    walk(&w, Node::Property(&prop))
                }
            }
            Node::MemberExpr(n) => {
                walk(&w, Node::from_expr(&n.object));
                walk(&w, Node::from_property_key(&n.property));
            }
            Node::IndexExpr(n) => {
                walk(&w, Node::from_expr(&n.array));
                walk(&w, Node::from_expr(&n.index));
            }
            Node::BinaryExpr(n) => {
                walk(&w, Node::from_expr(&n.left));
                walk(&w, Node::from_expr(&n.right));
            }
            Node::UnaryExpr(n) => {
                walk(&w, Node::from_expr(&n.argument));
            }
            Node::PipeExpr(n) => {
                walk(&w, Node::from_expr(&n.argument));
                walk(&w, Node::CallExpr(&n.call));
            }
            Node::CallExpr(n) => {
                walk(&w, Node::from_expr(&n.callee));
                for arg in n.arguments.iter() {
                    walk(&w, Node::from_expr(&arg));
                }
            }
            Node::ConditionalExpr(n) => {
                walk(&w, Node::from_expr(&n.test));
                walk(&w, Node::from_expr(&n.consequent));
                walk(&w, Node::from_expr(&n.alternate));
            }
            Node::StringExpr(n) => {
                for part in n.parts.iter() {
                    walk(&w, Node::from_string_expr_part(&part));
                }
            }
            Node::ParenExpr(n) => {
                walk(&w, Node::from_expr(&n.expression));
            }
            Node::IntegerLit(_) => {}
            Node::FloatLit(_) => {}
            Node::StringLit(_) => {}
            Node::DurationLit(_) => {}
            Node::UintLit(_) => {}
            Node::BooleanLit(_) => {}
            Node::DateTimeLit(_) => {}
            Node::RegexpLit(_) => {}
            Node::PipeLit(_) => {}
            Node::BadExpr(n) => {
                if let Some(e) = &n.expression {
                    walk(&w, Node::from_expr(e));
                }
            }
            Node::ExprStmt(n) => {
                walk(&w, Node::from_expr(&n.expression));
            }
            Node::OptionStmt(n) => {
                walk(&w, Node::from_assignment(&n.assignment));
            }
            Node::ReturnStmt(n) => {
                walk(&w, Node::from_expr(&n.argument));
            }
            Node::BadStmt(_) => {}
            Node::TestStmt(n) => {
                walk(&w, Node::VariableAssgn(&n.assignment));
            }
            Node::BuiltinStmt(n) => {
                walk(&w, Node::Identifier(&n.id));
            }
            Node::Block(n) => {
                for s in n.body.iter() {
                    walk(&w, Node::from_stmt(&s));
                }
            }
            Node::Property(n) => {
                walk(&w, Node::from_property_key(&n.key));
                if let Some(v) = &n.value {
                    walk(&w, Node::from_expr(v));
                }
            }
            Node::TextPart(_) => {}
            Node::InterpolatedPart(n) => {
                walk(&w, Node::from_expr(&n.expression));
            }
            Node::VariableAssgn(n) => {
                walk(&w, Node::Identifier(&n.id));
                walk(&w, Node::from_expr(&n.init));
            }
            Node::MemberAssgn(n) => {
                walk(&w, Node::MemberExpr(&n.member));
                walk(&w, Node::from_expr(&n.init));
            }
        }
    }
    v.done(&node)
}

type FuncVisitor<'a> = Rc<RefCell<&'a mut dyn FnMut(&Node)>>;

impl<'a> Visitor for FuncVisitor<'a> {
    fn visit(&self, node: &Node) -> Option<Self> {
        let mut func: RefMut<_> = self.borrow_mut();
        (&mut *func)(node);
        Some(Rc::clone(self))
    }
}

// Create Visitor will produce a visitor that calls the function for all nodes.
pub fn create_visitor(func: &mut dyn FnMut(&Node)) -> FuncVisitor {
    Rc::new(RefCell::new(func))
}
