// Module nodes contains the nodes for the semantic graph.
//
// NOTE(affo): At this stage some nodes are a clone of the AST nodes with some type information added.
//  Nevertheless, new node types allow us to decouple this step of compilation from the parsing.
//  This is of paramount importance if we decide to add responsibilities to the semantic analysis and
//  change it independently from the parsing bits.
//  Uncommented node types are a direct port of the AST ones.

extern crate chrono;
extern crate derivative;

use crate::ast;
use crate::semantic::infer;
use crate::semantic::types;
use crate::semantic::{
    env::Environment,
    fresh::Fresher,
    infer::{Constraint, Constraints},
    sub::{Substitutable, Substitution},
    types::{Array, Function, Kind, MonoType, PolyType, Tvar},
};

use crate::semantic::types::Kind::*;
use chrono::prelude::DateTime;
use chrono::Duration;
use chrono::FixedOffset;
use derivative::Derivative;
use std::collections::HashMap;
use std::fmt;
use std::vec::Vec;

// Result returned from the various 'infer' methods defined in this
// module. The result of inferring an expression or statment is an
// updated type environment and a set of type constraints to be solved.
type Result = std::result::Result<(Environment, Constraints), Error>;

#[derive(Debug)]
pub struct Error {
    msg: String,
}

impl fmt::Display for Error {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        f.write_str(&self.msg)
    }
}

impl From<types::Error> for Error {
    fn from(err: types::Error) -> Error {
        Error {
            msg: err.to_string(),
        }
    }
}

impl Error {
    fn undeclared_variable(name: String) -> Error {
        Error {
            msg: format!("undeclared variable {}", name),
        }
    }
    fn invalid_statement(msg: String) -> Error {
        Error { msg }
    }
    fn unsupported_binary_operator(op: &ast::Operator) -> Error {
        Error {
            msg: format!("unsupported binary operator {}", op.to_string()),
        }
    }
    fn unsupported_unary_operator(op: &ast::Operator) -> Error {
        Error {
            msg: format!("unsupported unary operator {}", op.to_string()),
        }
    }
    fn unknown_import_path(path: &str) -> Error {
        Error {
            msg: format!("\"{}\" is not a known import path", path),
        }
    }
}

#[derive(Debug, PartialEq, Clone)]
pub enum Statement {
    Expr(ExprStmt),
    Variable(VariableAssgn),
    Option(OptionStmt),
    Return(ReturnStmt),
    Test(TestStmt),
    Builtin(BuiltinStmt),
}

#[derive(Debug, PartialEq, Clone)]
pub enum Assignment {
    Variable(VariableAssgn),
    Member(MemberAssgn),
}

#[derive(Debug, PartialEq, Clone)]
pub enum Expression {
    Identifier(IdentifierExpr),
    Array(Box<ArrayExpr>),
    Function(Box<FunctionExpr>),
    Logical(Box<LogicalExpr>),
    Object(Box<ObjectExpr>),
    Member(Box<MemberExpr>),
    Index(Box<IndexExpr>),
    Binary(Box<BinaryExpr>),
    Unary(Box<UnaryExpr>),
    Call(Box<CallExpr>),
    Conditional(Box<ConditionalExpr>),
    StringExpr(Box<StringExpr>),

    Integer(IntegerLit),
    Float(FloatLit),
    StringLit(StringLit),
    Duration(DurationLit),
    Uint(UintLit),
    Boolean(BooleanLit),
    DateTime(DateTimeLit),
    Regexp(RegexpLit),
}

impl Expression {
    pub fn type_of(&self) -> &MonoType {
        match self {
            Expression::Identifier(e) => &e.typ,
            Expression::Array(e) => &e.typ,
            Expression::Function(e) => &e.typ,
            Expression::Logical(e) => &e.typ,
            Expression::Object(e) => &e.typ,
            Expression::Member(e) => &e.typ,
            Expression::Index(e) => &e.typ,
            Expression::Binary(e) => &e.typ,
            Expression::Unary(e) => &e.typ,
            Expression::Call(e) => &e.typ,
            Expression::Conditional(e) => &e.typ,
            Expression::StringExpr(e) => &e.typ,
            Expression::Integer(lit) => &lit.typ,
            Expression::Float(lit) => &lit.typ,
            Expression::StringLit(lit) => &lit.typ,
            Expression::Duration(lit) => &lit.typ,
            Expression::Uint(lit) => &lit.typ,
            Expression::Boolean(lit) => &lit.typ,
            Expression::DateTime(lit) => &lit.typ,
            Expression::Regexp(lit) => &lit.typ,
        }
    }
    pub fn loc(&self) -> &ast::SourceLocation {
        match self {
            Expression::Identifier(e) => &e.loc,
            Expression::Array(e) => &e.loc,
            Expression::Function(e) => &e.loc,
            Expression::Logical(e) => &e.loc,
            Expression::Object(e) => &e.loc,
            Expression::Member(e) => &e.loc,
            Expression::Index(e) => &e.loc,
            Expression::Binary(e) => &e.loc,
            Expression::Unary(e) => &e.loc,
            Expression::Call(e) => &e.loc,
            Expression::Conditional(e) => &e.loc,
            Expression::StringExpr(e) => &e.loc,
            Expression::Integer(lit) => &lit.loc,
            Expression::Float(lit) => &lit.loc,
            Expression::StringLit(lit) => &lit.loc,
            Expression::Duration(lit) => &lit.loc,
            Expression::Uint(lit) => &lit.loc,
            Expression::Boolean(lit) => &lit.loc,
            Expression::DateTime(lit) => &lit.loc,
            Expression::Regexp(lit) => &lit.loc,
        }
    }
    fn infer(&mut self, env: Environment, f: &mut Fresher) -> Result {
        match self {
            Expression::Identifier(e) => e.infer(env, f),
            Expression::Array(e) => e.infer(env, f),
            Expression::Function(e) => e.infer(env, f),
            Expression::Logical(e) => e.infer(env, f),
            Expression::Object(e) => e.infer(env, f),
            Expression::Member(e) => e.infer(env, f),
            Expression::Index(e) => e.infer(env, f),
            Expression::Binary(e) => e.infer(env, f),
            Expression::Unary(e) => e.infer(env, f),
            Expression::Call(e) => e.infer(env, f),
            Expression::Conditional(e) => e.infer(env, f),
            Expression::StringExpr(e) => e.infer(env, f),
            Expression::Integer(lit) => lit.infer(env, f),
            Expression::Float(lit) => lit.infer(env, f),
            Expression::StringLit(lit) => lit.infer(env, f),
            Expression::Duration(lit) => lit.infer(env, f),
            Expression::Uint(lit) => lit.infer(env, f),
            Expression::Boolean(lit) => lit.infer(env, f),
            Expression::DateTime(lit) => lit.infer(env, f),
            Expression::Regexp(lit) => lit.infer(env, f),
        }
    }
}

pub struct Importer<'a> {
    values: HashMap<&'a str, PolyType>,
}

impl<'a> From<HashMap<&'a str, PolyType>> for Importer<'a> {
    fn from(m: HashMap<&'a str, PolyType>) -> Importer<'a> {
        Importer { values: m }
    }
}

impl<'a> Importer<'a> {
    pub fn import(&'a self, name: &'a str) -> Option<&'a PolyType> {
        self.values.get(name)
    }
}

// Infer the types of a flux package
pub fn infer_pkg_types(
    pkg: &mut Package,
    env: Environment,
    f: &mut Fresher,
    i: &Importer,
) -> std::result::Result<(Environment, Substitution), Error> {
    let (env, cons) = pkg.infer(env, f, i)?;
    Ok((env, infer::solve(&cons, &mut HashMap::new(), f)?))
}

#[derive(Debug, PartialEq, Clone)]
pub struct Package {
    pub loc: ast::SourceLocation,

    pub package: String,
    pub files: Vec<File>,
}

impl Package {
    fn infer(&mut self, env: Environment, f: &mut Fresher, i: &Importer) -> Result {
        self.files
            .iter_mut()
            .try_fold((env, Constraints::empty()), |(env, rest), file| {
                let (env, cons) = file.infer(env, f, i)?;
                Ok((env, cons + rest))
            })
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct File {
    pub loc: ast::SourceLocation,

    pub package: Option<PackageClause>,
    pub imports: Vec<ImportDeclaration>,
    pub body: Vec<Statement>,
}

impl File {
    fn infer(&mut self, mut env: Environment, f: &mut Fresher, i: &Importer) -> Result {
        let mut imports = Vec::with_capacity(self.imports.len());

        for dec in &self.imports {
            let path = &dec.path.value;
            let name = pkg_name_from_path(&path);

            imports.push(name);

            match i.import(path) {
                Some(poly) => env.add(name.to_owned(), poly.clone()),
                None => return Err(Error::unknown_import_path(path)),
            };
        }

        let (mut env, constraints) =
            self.body
                .iter_mut()
                .try_fold(
                    (env, Constraints::empty()),
                    |(env, rest), node| match node {
                        Statement::Builtin(stmt) => {
                            let (env, cons) = stmt.infer(env, f)?;
                            Ok((env, cons + rest))
                        }
                        Statement::Variable(stmt) => {
                            let (env, cons) = stmt.infer(env, f)?;
                            Ok((env, cons + rest))
                        }
                        Statement::Option(stmt) => {
                            let (env, cons) = stmt.infer(env, f)?;
                            Ok((env, cons + rest))
                        }
                        Statement::Expr(stmt) => {
                            let (env, cons) = stmt.infer(env, f)?;
                            Ok((env, cons + rest))
                        }
                        Statement::Test(stmt) => {
                            let (env, cons) = stmt.infer(env, f)?;
                            Ok((env, cons + rest))
                        }
                        Statement::Return(_) => Err(Error::invalid_statement(String::from(
                            "cannot have return statement in file block",
                        ))),
                    },
                )?;

        for name in imports {
            env.remove(name);
        }
        Ok((env, constraints))
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct PackageClause {
    pub loc: ast::SourceLocation,

    pub name: Identifier,
}

#[derive(Debug, PartialEq, Clone)]
pub struct ImportDeclaration {
    pub loc: ast::SourceLocation,

    pub alias: Option<Identifier>,
    pub path: StringLit,
}

fn pkg_name_from_path(path: &str) -> &str {
    path.rsplitn(2, '/').next().unwrap()
}

#[derive(Debug, PartialEq, Clone)]
pub struct OptionStmt {
    pub loc: ast::SourceLocation,

    pub assignment: Assignment,
}

impl OptionStmt {
    fn infer(&mut self, env: Environment, f: &mut Fresher) -> Result {
        match &mut self.assignment {
            Assignment::Member(stmt) => {
                let (env, cons) = stmt.init.infer(env, f)?;
                let (env, rest) = stmt.member.infer(env, f)?;

                let l = stmt.member.typ.clone();
                let r = stmt.init.type_of().clone();

                Ok((env, cons + rest + vec![Constraint::Equal(l, r)].into()))
            }
            Assignment::Variable(stmt) => stmt.infer(env, f),
        }
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct BuiltinStmt {
    pub loc: ast::SourceLocation,

    pub id: Identifier,
}

impl BuiltinStmt {
    fn infer(&self, _: Environment, _: &mut Fresher) -> Result {
        unimplemented!();
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct TestStmt {
    pub loc: ast::SourceLocation,

    pub assignment: VariableAssgn,
}

impl TestStmt {
    fn infer(&mut self, env: Environment, f: &mut Fresher) -> Result {
        self.assignment.infer(env, f)
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct ExprStmt {
    pub loc: ast::SourceLocation,

    pub expression: Expression,
}

impl ExprStmt {
    fn infer(&mut self, env: Environment, f: &mut Fresher) -> Result {
        let (env, cons) = self.expression.infer(env, f)?;
        let sub = infer::solve(&cons, &mut HashMap::new(), f)?;
        Ok((env.apply(&sub), cons))
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct ReturnStmt {
    pub loc: ast::SourceLocation,

    pub argument: Expression,
}

impl ReturnStmt {
    fn infer(&mut self, env: Environment, f: &mut Fresher) -> Result {
        self.argument.infer(env, f)
    }
}

#[derive(Debug, Derivative, PartialEq, Clone)]
pub struct VariableAssgn {
    #[derivative(PartialEq = "ignore")]
    vars: Vec<Tvar>,

    #[derivative(PartialEq = "ignore")]
    cons: HashMap<Tvar, Vec<Kind>>,

    pub loc: ast::SourceLocation,

    pub id: Identifier,
    pub init: Expression,
}

impl VariableAssgn {
    pub fn new(id: Identifier, init: Expression, loc: ast::SourceLocation) -> VariableAssgn {
        VariableAssgn {
            vars: Vec::new(),
            cons: HashMap::new(),
            loc,
            id,
            init,
        }
    }
    pub fn poly_type_of(&self) -> PolyType {
        PolyType {
            vars: self.vars.clone(),
            cons: self.cons.clone(),
            expr: self.init.type_of().clone(),
        }
    }
    // Polymorphic generalization, necessary for let-polymorphism, is
    // implemented here.
    //
    // In particular, for every variable assignment we infer the type of
    // its corresponding expression. We then generalize that type by
    // quantifying over all of its free type variables. Finally we bind
    // the variable to its newly generalized type in the type environment
    // before inferring the rest of the program.
    //
    fn infer(&mut self, env: Environment, f: &mut Fresher) -> Result {
        let (env, constraints) = self.init.infer(env, f)?;

        let mut kinds = HashMap::new();
        let sub = infer::solve(&constraints, &mut kinds, f)?;

        // Apply substitution to the type environment
        let mut env = env.apply(&sub);

        let t = self.init.type_of().clone().apply(&sub);
        let p = infer::generalize(&env, &kinds, t);

        // Update variable assignment nodes with the free vars
        // and kind constraints obtained from generalization.
        //
        // Note these variables are fixed after generalization
        // and so it is safe to update these nodes in place.
        self.vars = p.vars.clone();
        self.cons = p.cons.clone();

        // Update the type environment
        &mut env.add(String::from(&self.id.name), p);
        Ok((env, constraints))
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct MemberAssgn {
    pub loc: ast::SourceLocation,

    pub member: MemberExpr,
    pub init: Expression,
}

#[derive(Derivative)]
#[derivative(Debug, PartialEq, Clone)]
pub struct StringExpr {
    pub loc: ast::SourceLocation,
    #[derivative(PartialEq = "ignore")]
    pub typ: MonoType,

    pub parts: Vec<StringExprPart>,
}

impl StringExpr {
    fn infer(&mut self, env: Environment, f: &mut Fresher) -> Result {
        let mut env = env;
        let mut constraints = Vec::new();
        for p in &mut self.parts {
            if let StringExprPart::Interpolated(ref mut ip) = p {
                let (e, cons) = ip.expression.infer(env, f)?;
                constraints.append(&mut Vec::from(cons));
                constraints.push(Constraint::Equal(
                    ip.expression.type_of().clone(),
                    MonoType::String,
                ));
                env = e
            }
        }
        constraints.push(Constraint::Equal(self.typ.clone(), MonoType::String));
        return Ok((env, Constraints::from(constraints)));
    }
}

#[derive(Debug, PartialEq, Clone)]
pub enum StringExprPart {
    Text(TextPart),
    Interpolated(InterpolatedPart),
}

#[derive(Debug, PartialEq, Clone)]
pub struct TextPart {
    pub loc: ast::SourceLocation,

    pub value: String,
}

#[derive(Debug, PartialEq, Clone)]
pub struct InterpolatedPart {
    pub loc: ast::SourceLocation,

    pub expression: Expression,
}

#[derive(Derivative)]
#[derivative(Debug, PartialEq, Clone)]
pub struct ArrayExpr {
    pub loc: ast::SourceLocation,
    #[derivative(PartialEq = "ignore")]
    pub typ: MonoType,

    pub elements: Vec<Expression>,
}

impl ArrayExpr {
    fn infer(&mut self, mut env: Environment, f: &mut Fresher) -> Result {
        let mut cons = Vec::new();
        let elt = MonoType::Var(f.fresh());
        for el in &mut self.elements {
            let (e, c) = el.infer(env, f)?;
            cons.append(&mut c.into());
            cons.push(Constraint::Equal(el.type_of().clone(), elt.clone()));
            env = e;
        }
        let at = MonoType::Arr(Box::new(Array(elt)));
        cons.push(Constraint::Equal(at, self.typ.clone()));
        return Ok((env, cons.into()));
    }
}

// FunctionExpr represents the definition of a function
#[derive(Derivative)]
#[derivative(Debug, PartialEq, Clone)]
pub struct FunctionExpr {
    pub loc: ast::SourceLocation,
    #[derivative(PartialEq = "ignore")]
    pub typ: MonoType,

    pub params: Vec<FunctionParameter>,
    pub body: Block,
}

impl FunctionExpr {
    fn infer(&mut self, mut env: Environment, f: &mut Fresher) -> Result {
        let mut cons = Constraints::empty();
        let mut pipe = None;
        let mut req = HashMap::new();
        let mut opt = HashMap::new();
        // This params will build the nested env when inferring the function body.
        let mut params = HashMap::new();
        for param in &mut self.params {
            match param.default {
                Some(ref mut e) => {
                    let (nenv, ncons) = e.infer(env, f)?;
                    cons = cons + ncons;
                    let id = param.key.name.clone();
                    // We are here: `f = (a=1) => {...}`.
                    // So, this PolyType is actually a MonoType, whose type
                    // is the one of the default value ("1" in "a=1").
                    let typ = PolyType {
                        vars: Vec::new(),
                        cons: HashMap::new(),
                        expr: e.type_of().clone(),
                    };
                    params.insert(id.clone(), typ);
                    opt.insert(id, e.type_of().clone());
                    env = nenv;
                }
                None => {
                    // We are here: `f = (a) => {...}`.
                    // So, we do not know the type of "a". Let's use a fresh TVar.
                    let id = param.key.name.clone();
                    let ftvar = f.fresh();
                    let typ = PolyType {
                        vars: Vec::new(),
                        cons: HashMap::new(),
                        expr: MonoType::Var(ftvar),
                    };
                    params.insert(id.clone(), typ.clone());
                    // Piped arguments cannot have a default value.
                    // So check if this is a piped argument.
                    if param.is_pipe {
                        pipe = Some(types::Property {
                            k: id,
                            v: MonoType::Var(ftvar),
                        });
                    } else {
                        req.insert(id, MonoType::Var(ftvar));
                    }
                }
            };
        }
        // Add the parameters to some nested environment.
        let mut nenv = Environment::new(env);
        for (id, param) in params.into_iter() {
            nenv.add(id, param);
        }
        // And use it to infer the body.
        let (nenv, bcons) = self.body.infer(nenv, f)?;
        // Now pop the nested environment, we don't need it anymore.
        let env = nenv.pop();
        let retn = self.body.type_of().clone();
        let func = MonoType::Fun(Box::new(Function {
            req,
            opt,
            pipe,
            retn,
        }));
        cons = cons + bcons;
        cons.add(Constraint::Equal(func, self.typ.clone()));
        return Ok((env, cons));
    }

    pub fn pipe(&self) -> Option<&FunctionParameter> {
        for p in &self.params {
            if p.is_pipe {
                return Some(p);
            }
        }
        None
    }

    pub fn defaults(&self) -> Vec<&FunctionParameter> {
        let mut ds = Vec::new();
        for p in &self.params {
            match p.default {
                Some(_) => ds.push(p),
                None => (),
            }
        }
        ds
    }
}

// Block represents a function block and is equivalent to a let-expression
// in other functional languages.
//
// Functions must evaluate to a value in Flux. In other words, a function
// must always have a return value. This means a function block is by
// definition an expression.
//
#[derive(Debug, PartialEq, Clone)]
pub enum Block {
    Variable(VariableAssgn, Box<Block>),
    Expr(ExprStmt, Box<Block>),
    Return(Expression),
}

impl Block {
    fn infer(&mut self, env: Environment, f: &mut Fresher) -> Result {
        match self {
            Block::Variable(stmt, block) => {
                let (env, cons) = stmt.infer(env, f)?;
                let (env, rest) = block.infer(env, f)?;

                Ok((env, cons + rest))
            }
            Block::Expr(stmt, block) => {
                let (env, cons) = stmt.infer(env, f)?;
                let (env, rest) = block.infer(env, f)?;

                Ok((env, cons + rest))
            }
            Block::Return(e) => e.infer(env, f),
        }
    }
    pub fn loc(&self) -> &ast::SourceLocation {
        match self {
            Block::Variable(ass, _) => &ass.loc,
            Block::Expr(es, _) => es.expression.loc(),
            Block::Return(expr) => expr.loc(),
        }
    }
    pub fn type_of(&self) -> &MonoType {
        let mut n = self;
        loop {
            n = match n {
                Block::Variable(_, b) => b.as_ref(),
                Block::Expr(_, b) => b.as_ref(),
                Block::Return(e) => return e.type_of(),
            }
        }
    }
}

// FunctionParameter represents a function parameter.
#[derive(Debug, PartialEq, Clone)]
pub struct FunctionParameter {
    pub loc: ast::SourceLocation,

    pub is_pipe: bool,
    pub key: Identifier,
    pub default: Option<Expression>,
}

#[derive(Derivative)]
#[derivative(Debug, PartialEq, Clone)]
pub struct BinaryExpr {
    pub loc: ast::SourceLocation,
    #[derivative(PartialEq = "ignore")]
    pub typ: MonoType,

    pub operator: ast::Operator,
    pub left: Expression,
    pub right: Expression,
}

impl BinaryExpr {
    fn infer(&mut self, env: Environment, f: &mut Fresher) -> Result {
        // Compute the left and right constraints.
        // Do this first so that we can return an error if one occurs.
        let (env, lcons) = self.left.infer(env, f)?;
        let (env, rcons) = self.right.infer(env, f)?;

        let cons = match self.operator {
            // The following operators require both sides to be equal.
            ast::Operator::AdditionOperator => Constraints::from(vec![
                Constraint::Equal(self.left.type_of().clone(), self.right.type_of().clone()),
                Constraint::Equal(self.left.type_of().clone(), self.typ.clone()),
                Constraint::Kind(self.typ.clone(), Kind::Addable),
            ]),
            ast::Operator::SubtractionOperator => Constraints::from(vec![
                Constraint::Equal(self.left.type_of().clone(), self.right.type_of().clone()),
                Constraint::Equal(self.left.type_of().clone(), self.typ.clone()),
                Constraint::Kind(self.typ.clone(), Kind::Subtractable),
            ]),
            ast::Operator::MultiplicationOperator => Constraints::from(vec![
                Constraint::Equal(self.left.type_of().clone(), self.right.type_of().clone()),
                Constraint::Equal(self.left.type_of().clone(), self.typ.clone()),
                Constraint::Kind(self.typ.clone(), Kind::Divisible),
            ]),
            ast::Operator::DivisionOperator => Constraints::from(vec![
                Constraint::Equal(self.left.type_of().clone(), self.right.type_of().clone()),
                Constraint::Equal(self.left.type_of().clone(), self.typ.clone()),
                Constraint::Kind(self.typ.clone(), Kind::Divisible),
            ]),
            ast::Operator::PowerOperator => Constraints::from(vec![
                Constraint::Equal(self.left.type_of().clone(), self.right.type_of().clone()),
                Constraint::Equal(self.left.type_of().clone(), self.typ.clone()),
                Constraint::Kind(self.typ.clone(), Kind::Divisible),
            ]),
            ast::Operator::ModuloOperator => Constraints::from(vec![
                Constraint::Equal(self.left.type_of().clone(), self.right.type_of().clone()),
                Constraint::Equal(self.left.type_of().clone(), self.typ.clone()),
                Constraint::Kind(self.typ.clone(), Kind::Divisible),
            ]),
            ast::Operator::GreaterThanOperator => Constraints::from(vec![
                Constraint::Equal(self.left.type_of().clone(), self.right.type_of().clone()),
                Constraint::Equal(self.typ.clone(), MonoType::Bool),
                Constraint::Kind(self.left.type_of().clone(), Kind::Comparable),
            ]),
            ast::Operator::LessThanOperator => Constraints::from(vec![
                Constraint::Equal(self.left.type_of().clone(), self.right.type_of().clone()),
                Constraint::Equal(self.typ.clone(), MonoType::Bool),
                Constraint::Kind(self.left.type_of().clone(), Kind::Comparable),
            ]),
            ast::Operator::EqualOperator => Constraints::from(vec![
                Constraint::Equal(self.left.type_of().clone(), self.right.type_of().clone()),
                Constraint::Equal(self.typ.clone(), MonoType::Bool),
                Constraint::Kind(self.left.type_of().clone(), Kind::Equatable),
            ]),
            ast::Operator::NotEqualOperator => Constraints::from(vec![
                Constraint::Equal(self.left.type_of().clone(), self.right.type_of().clone()),
                Constraint::Equal(self.typ.clone(), MonoType::Bool),
                Constraint::Kind(self.left.type_of().clone(), Kind::Equatable),
            ]),
            ast::Operator::GreaterThanEqualOperator => Constraints::from(vec![
                Constraint::Equal(self.left.type_of().clone(), self.right.type_of().clone()),
                Constraint::Equal(self.typ.clone(), MonoType::Bool),
                Constraint::Kind(self.left.type_of().clone(), Kind::Equatable),
                Constraint::Kind(self.left.type_of().clone(), Kind::Comparable),
            ]),
            ast::Operator::LessThanEqualOperator => Constraints::from(vec![
                Constraint::Equal(self.left.type_of().clone(), self.right.type_of().clone()),
                Constraint::Equal(self.typ.clone(), MonoType::Bool),
                Constraint::Kind(self.left.type_of().clone(), Kind::Equatable),
                Constraint::Kind(self.left.type_of().clone(), Kind::Comparable),
            ]),
            // Regular expression operators.
            ast::Operator::RegexpMatchOperator => Constraints::from(vec![
                Constraint::Equal(self.left.type_of().clone(), MonoType::String),
                Constraint::Equal(self.right.type_of().clone(), MonoType::Regexp),
                Constraint::Equal(self.typ.clone(), MonoType::Bool),
            ]),
            ast::Operator::NotRegexpMatchOperator => Constraints::from(vec![
                Constraint::Equal(self.left.type_of().clone(), MonoType::String),
                Constraint::Equal(self.right.type_of().clone(), MonoType::Regexp),
                Constraint::Equal(self.typ.clone(), MonoType::Bool),
            ]),
            _ => return Err(Error::unsupported_binary_operator(&self.operator)),
        };

        // Otherwise, add the constraints together and return them.
        return Ok((env, lcons + rcons + cons));
    }
}

#[derive(Derivative)]
#[derivative(Debug, PartialEq, Clone)]
pub struct CallExpr {
    pub loc: ast::SourceLocation,
    #[derivative(PartialEq = "ignore")]
    pub typ: MonoType,

    pub callee: Expression,
    pub arguments: Vec<Property>,
    pub pipe: Option<Expression>,
}

impl CallExpr {
    fn infer(&mut self, env: Environment, f: &mut Fresher) -> Result {
        // First, recursively infer every type of the children of this call expression,
        // update the environment and the constraints, and use the inferred types to
        // build the fields of the type for this call expression.
        let (mut env, mut cons) = self.callee.infer(env, f)?;
        let mut req = HashMap::new();
        let mut pipe = None;
        for Property {
            key: ref mut id,
            value: ref mut expr,
            ..
        } in &mut self.arguments
        {
            let (nenv, ncons) = expr.infer(env, f)?;
            cons = cons + ncons;
            env = nenv;
            // Every argument is required in a function call.
            req.insert(id.name.clone(), expr.type_of().clone());
        }
        if let Some(ref mut p) = &mut self.pipe {
            let (nenv, ncons) = p.infer(env, f)?;
            cons = cons + ncons;
            env = nenv;
            pipe = Some(types::Property {
                k: "<-".to_string(),
                v: p.type_of().clone(),
            });
        }
        // Constrain the callee to be a Function.
        cons.add(Constraint::Equal(
            self.callee.type_of().clone(),
            MonoType::Fun(Box::new(Function {
                opt: HashMap::new(),
                req,
                pipe,
                // The return type of a function call is the type of the call itself.
                // Remind that, when two functions are unified, their return types are unified too.
                // As an example take:
                //   f = (a) => a + 1
                //   f(a: 0)
                // The return type of `f` is `int`.
                // The return type of `f(a: 0)` is `t0` (a fresh type variable).
                // Upon unification a substitution "t0 => int" is created, so that the compiler
                // can infer that, for instance, `f(a: 0) + 1` is legal.
                retn: self.typ.clone(),
            })),
        ));
        Ok((env, cons))
    }
}

#[derive(Derivative)]
#[derivative(Debug, PartialEq, Clone)]
pub struct ConditionalExpr {
    pub loc: ast::SourceLocation,
    #[derivative(PartialEq = "ignore")]
    pub typ: MonoType,

    pub test: Expression,
    pub consequent: Expression,
    pub alternate: Expression,
}

impl ConditionalExpr {
    fn infer(&mut self, env: Environment, f: &mut Fresher) -> Result {
        let (env, tcons) = self.test.infer(env, f)?;
        let (env, ccons) = self.consequent.infer(env, f)?;
        let (env, acons) = self.alternate.infer(env, f)?;
        let cons = tcons
            + ccons
            + acons
            + Constraints::from(vec![
                Constraint::Equal(self.test.type_of().clone(), MonoType::Bool),
                Constraint::Equal(
                    self.consequent.type_of().clone(),
                    self.alternate.type_of().clone(),
                ),
                Constraint::Equal(self.consequent.type_of().clone(), self.typ.clone()),
            ]);
        return Ok((env, cons));
    }
}

#[derive(Derivative)]
#[derivative(Debug, PartialEq, Clone)]
pub struct LogicalExpr {
    pub loc: ast::SourceLocation,
    #[derivative(PartialEq = "ignore")]
    pub typ: MonoType,

    pub operator: ast::LogicalOperator,
    pub left: Expression,
    pub right: Expression,
}

impl LogicalExpr {
    fn infer(&mut self, env: Environment, f: &mut Fresher) -> Result {
        let (env, lcons) = self.left.infer(env, f)?;
        let (env, rcons) = self.right.infer(env, f)?;
        let cons = lcons
            + rcons
            + Constraints::from(vec![
                Constraint::Equal(self.left.type_of().clone(), MonoType::Bool),
                Constraint::Equal(self.right.type_of().clone(), MonoType::Bool),
                Constraint::Equal(self.typ.clone(), MonoType::Bool),
            ]);
        return Ok((env, cons));
    }
}

#[derive(Derivative)]
#[derivative(Debug, PartialEq, Clone)]
pub struct MemberExpr {
    pub loc: ast::SourceLocation,
    #[derivative(PartialEq = "ignore")]
    pub typ: MonoType,

    pub object: Expression,
    pub property: String,
}

impl MemberExpr {
    // A member expression such as `r.a` produces the constraint:
    //
    //     type_of(r) = {a: type_of(r.a) | 'r}
    //
    // where 'r is a fresh type variable.
    //
    fn infer(&mut self, env: Environment, f: &mut Fresher) -> Result {
        let head = types::Property {
            k: self.property.to_owned(),
            v: self.typ.to_owned(),
        };
        let tail = MonoType::Var(f.fresh());

        let r = MonoType::from(types::Row::Extension { head, tail });
        let t = self.object.type_of().to_owned();

        let (env, cons) = self.object.infer(env, f)?;
        Ok((env, cons + vec![Constraint::Equal(t, r)].into()))
    }
}

#[derive(Derivative)]
#[derivative(Debug, PartialEq, Clone)]
pub struct IndexExpr {
    pub loc: ast::SourceLocation,
    #[derivative(PartialEq = "ignore")]
    pub typ: MonoType,

    pub array: Expression,
    pub index: Expression,
}

impl IndexExpr {
    fn infer(&mut self, env: Environment, f: &mut Fresher) -> Result {
        let (env, acons) = self.array.infer(env, f)?;
        let (env, icons) = self.index.infer(env, f)?;
        let cons = acons
            + icons
            + Constraints::from(vec![
                Constraint::Equal(self.index.type_of().clone(), MonoType::Int),
                Constraint::Equal(
                    self.array.type_of().clone(),
                    MonoType::Arr(Box::new(Array(self.typ.clone()))),
                ),
            ]);
        return Ok((env, cons));
    }
}

#[derive(Derivative)]
#[derivative(Debug, PartialEq, Clone)]
pub struct ObjectExpr {
    pub loc: ast::SourceLocation,
    #[derivative(PartialEq = "ignore")]
    pub typ: MonoType,

    pub with: Option<IdentifierExpr>,
    pub properties: Vec<Property>,
}

impl ObjectExpr {
    fn infer(&mut self, mut env: Environment, f: &mut Fresher) -> Result {
        // If record extension, infer constraints for base
        let (mut r, mut cons) = match &mut self.with {
            Some(expr) => {
                let (e, cons) = expr.infer(env, f)?;
                env = e;
                (expr.typ.to_owned(), cons)
            }
            None => (
                MonoType::Row(Box::new(types::Row::Empty)),
                Constraints::empty(),
            ),
        };
        // Infer constraints for properties
        for prop in self.properties.iter_mut().rev() {
            let (e, rest) = prop.value.infer(env, f)?;
            env = e;
            cons = cons + rest;
            r = MonoType::Row(Box::new(types::Row::Extension {
                head: types::Property {
                    k: prop.key.name.to_owned(),
                    v: prop.value.type_of().to_owned(),
                },
                tail: r,
            }));
        }
        Ok((
            env,
            cons + vec![Constraint::Equal(self.typ.to_owned(), r)].into(),
        ))
    }
}

#[derive(Derivative)]
#[derivative(Debug, PartialEq, Clone)]
pub struct UnaryExpr {
    pub loc: ast::SourceLocation,
    #[derivative(PartialEq = "ignore")]
    pub typ: MonoType,

    pub operator: ast::Operator,
    pub argument: Expression,
}

impl UnaryExpr {
    fn infer(&mut self, env: Environment, f: &mut Fresher) -> Result {
        let (env, acons) = self.argument.infer(env, f)?;
        let cons = match self.operator {
            ast::Operator::NotOperator => Constraints::from(vec![
                Constraint::Equal(self.argument.type_of().clone(), MonoType::Bool),
                Constraint::Equal(self.typ.clone(), MonoType::Bool),
            ]),
            ast::Operator::ExistsOperator => {
                Constraints::from(Constraint::Equal(self.typ.clone(), MonoType::Bool))
            }
            ast::Operator::AdditionOperator => Constraints::from(vec![
                Constraint::Equal(self.argument.type_of().clone(), self.typ.clone()),
                Constraint::Kind(self.argument.type_of().clone(), Kind::Addable),
            ]),
            ast::Operator::SubtractionOperator => Constraints::from(vec![
                Constraint::Equal(self.argument.type_of().clone(), self.typ.clone()),
                Constraint::Kind(self.argument.type_of().clone(), Kind::Subtractable),
            ]),
            _ => return Err(Error::unsupported_unary_operator(&self.operator)),
        };
        return Ok((env, acons + cons));
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct Property {
    pub loc: ast::SourceLocation,

    pub key: Identifier,
    pub value: Expression,
}

#[derive(Derivative)]
#[derivative(Debug, PartialEq, Clone)]
pub struct IdentifierExpr {
    pub loc: ast::SourceLocation,
    #[derivative(PartialEq = "ignore")]
    pub typ: MonoType,

    pub name: String,
}

impl IdentifierExpr {
    fn infer(&self, env: Environment, f: &mut Fresher) -> Result {
        match env.lookup(&self.name) {
            Some(poly) => {
                let (t, cons) = infer::instantiate(poly.clone(), f);
                Ok((
                    env,
                    cons + Constraints::from(vec![Constraint::Equal(t, self.typ.clone())]),
                ))
            }
            None => Err(Error::undeclared_variable(self.name.to_string())),
        }
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct Identifier {
    pub loc: ast::SourceLocation,

    pub name: String,
}

#[derive(Derivative)]
#[derivative(Debug, PartialEq, Clone)]
pub struct BooleanLit {
    pub loc: ast::SourceLocation,
    #[derivative(PartialEq = "ignore")]
    pub typ: MonoType,

    pub value: bool,
}

impl BooleanLit {
    fn infer(&self, env: Environment, f: &mut Fresher) -> Result {
        return infer_literal(env, &self.typ, MonoType::Bool);
    }
}

#[derive(Derivative)]
#[derivative(Debug, PartialEq, Clone)]
pub struct IntegerLit {
    pub loc: ast::SourceLocation,
    #[derivative(PartialEq = "ignore")]
    pub typ: MonoType,

    pub value: i64,
}

impl IntegerLit {
    fn infer(&self, env: Environment, f: &mut Fresher) -> Result {
        return infer_literal(env, &self.typ, MonoType::Int);
    }
}

#[derive(Derivative)]
#[derivative(Debug, PartialEq, Clone)]
pub struct FloatLit {
    pub loc: ast::SourceLocation,
    #[derivative(PartialEq = "ignore")]
    pub typ: MonoType,

    pub value: f64,
}

impl FloatLit {
    fn infer(&self, env: Environment, f: &mut Fresher) -> Result {
        return infer_literal(env, &self.typ, MonoType::Float);
    }
}

#[derive(Derivative)]
#[derivative(Debug, PartialEq, Clone)]
pub struct RegexpLit {
    pub loc: ast::SourceLocation,
    #[derivative(PartialEq = "ignore")]
    pub typ: MonoType,

    // TODO(affo): should this be a compiled regexp?
    pub value: String,
}

impl RegexpLit {
    fn infer(&self, env: Environment, f: &mut Fresher) -> Result {
        return infer_literal(env, &self.typ, MonoType::Regexp);
    }
}

#[derive(Derivative)]
#[derivative(Debug, PartialEq, Clone)]
pub struct StringLit {
    pub loc: ast::SourceLocation,
    #[derivative(PartialEq = "ignore")]
    pub typ: MonoType,

    pub value: String,
}

impl StringLit {
    fn infer(&self, env: Environment, f: &mut Fresher) -> Result {
        return infer_literal(env, &self.typ, MonoType::String);
    }
}

#[derive(Derivative)]
#[derivative(Debug, PartialEq, Clone)]
pub struct UintLit {
    pub loc: ast::SourceLocation,
    #[derivative(PartialEq = "ignore")]
    pub typ: MonoType,

    pub value: u64,
}

impl UintLit {
    fn infer(&self, env: Environment, f: &mut Fresher) -> Result {
        return infer_literal(env, &self.typ, MonoType::Uint);
    }
}

#[derive(Derivative)]
#[derivative(Debug, PartialEq, Clone)]
pub struct DateTimeLit {
    pub loc: ast::SourceLocation,
    #[derivative(PartialEq = "ignore")]
    pub typ: MonoType,

    pub value: DateTime<FixedOffset>,
}

impl DateTimeLit {
    fn infer(&self, env: Environment, f: &mut Fresher) -> Result {
        return infer_literal(env, &self.typ, MonoType::Time);
    }
}

#[derive(Derivative)]
#[derivative(Debug, PartialEq, Clone)]
pub struct DurationLit {
    pub loc: ast::SourceLocation,
    #[derivative(PartialEq = "ignore")]
    pub typ: MonoType,

    pub value: Duration,
}

impl DurationLit {
    fn infer(&self, env: Environment, f: &mut Fresher) -> Result {
        return infer_literal(env, &self.typ, MonoType::Duration);
    }
}

fn infer_literal(env: Environment, typ: &MonoType, is: MonoType) -> Result {
    let constraints = Constraints::from(vec![Constraint::Equal(typ.clone(), is)]);
    return Ok((env, constraints));
}

const NANOS: i64 = 1;
const MICROS: i64 = NANOS * 1000;
const MILLIS: i64 = MICROS * 1000;
const SECONDS: i64 = MILLIS * 1000;
const MINUTES: i64 = SECONDS * 60;
const HOURS: i64 = MINUTES * 60;
const DAYS: i64 = HOURS * 24;
const WEEKS: i64 = DAYS * 7;
const MONTHS: f64 = WEEKS as f64 * (365.25 / 12.0 / 7.0);
const YEARS: f64 = MONTHS * 12.0;

// TODO(affo): this is not accurate, a duration value depends on the time in which it is calculated.
// 1 month is different if now is the 1st of January, or the 1st of February.
// Some days do not last 24 hours because of light savings.
pub fn convert_duration(duration: &Vec<ast::Duration>) -> std::result::Result<Duration, String> {
    let d = duration
        .iter()
        .try_fold(0 as i64, |acc, d| match d.unit.as_str() {
            "y" => Ok(acc + (d.magnitude as f64 * YEARS) as i64),
            "mo" => Ok(acc + (d.magnitude as f64 * MONTHS) as i64),
            "w" => Ok(acc + d.magnitude * WEEKS),
            "d" => Ok(acc + d.magnitude * DAYS),
            "h" => Ok(acc + d.magnitude * HOURS),
            "m" => Ok(acc + d.magnitude * MINUTES),
            "s" => Ok(acc + d.magnitude * SECONDS),
            "ms" => Ok(acc + d.magnitude * MILLIS),
            "us" | "µs" => Ok(acc + d.magnitude * MICROS),
            "ns" => Ok(acc + d.magnitude * NANOS),
            _ => Err(format!("unrecognized magnitude for duration: {}", d.unit)),
        })?;
    Ok(Duration::nanoseconds(d))
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::parser::parse_string;
    use crate::semantic::analyze::analyze_with;
    use crate::semantic::types::{MonoType, PolyType};
    use std::collections::HashMap;

    #[test]
    fn test_infer_instantiation() {
        let src = r#"
            x = f
            y = f
            z = a
        "#;

        // This environment represents our prelude
        //
        //     a = 5
        //     f = (a, b) => 2 * (a + b)
        //
        let env = Environment::from(maplit::hashmap! {
            // a = 5
            String::from("a") => PolyType {
                vars: Vec::new(),
                cons: HashMap::new(),
                expr: MonoType::Int,
            },
            // f = (a, b) => 2 * (a + b)
            String::from("f") => PolyType {
                vars: vec![Tvar(0)],
                cons: maplit::hashmap! { Tvar(0) => vec![Kind::Addable, Kind::Divisible]},
                expr: MonoType::Fun(Box::new(types::Function {
                    req: maplit::hashmap! {
                        String::from("a") => MonoType::Var(Tvar(0)),
                        String::from("b") => MonoType::Var(Tvar(0)),
                    },
                    opt: HashMap::new(),
                    pipe: None,
                    retn: MonoType::Var(Tvar(0)),
                })),
            },
        });

        let file = parse_string("file", src);

        let ast = ast::Package {
            base: file.base.clone(),
            path: "path/to/pkg".to_string(),
            package: "main".to_string(),
            files: vec![file],
        };

        let mut f: Fresher = 1.into();
        let mut pkg = analyze_with(ast, &mut f).unwrap();

        let (env, _) = infer_pkg_types(
            &mut pkg,
            env,
            &mut f,
            &Importer {
                values: HashMap::new(),
            },
        )
        .unwrap();

        let normalized: HashMap<String, PolyType> = env
            .values
            .into_iter()
            .map(|(k, v)| (k, v.fresh(&mut Fresher::new())))
            .collect();

        assert_eq!(
            normalized,
            maplit::hashmap! {
                String::from("f") => PolyType {
                    vars: vec![Tvar(0)],
                    cons: maplit::hashmap! { Tvar(0) => vec![Kind::Addable, Kind::Divisible]},
                    expr: MonoType::Fun(Box::new(types::Function {
                        req: maplit::hashmap! {
                            String::from("a") => MonoType::Var(Tvar(0)),
                            String::from("b") => MonoType::Var(Tvar(0)),
                        },
                        opt: HashMap::new(),
                        pipe: None,
                        retn: MonoType::Var(Tvar(0)),
                    })),
                },
                String::from("a") => PolyType {
                    vars: Vec::new(),
                    cons: HashMap::new(),
                    expr: MonoType::Int,
                },
                String::from("x") => PolyType {
                    vars: vec![Tvar(0)],
                    cons: maplit::hashmap! { Tvar(0) => vec![Kind::Addable, Kind::Divisible]},
                    expr: MonoType::Fun(Box::new(types::Function {
                        req: maplit::hashmap! {
                            String::from("a") => MonoType::Var(Tvar(0)),
                            String::from("b") => MonoType::Var(Tvar(0)),
                        },
                        opt: HashMap::new(),
                        pipe: None,
                        retn: MonoType::Var(Tvar(0)),
                    })),
                },
                String::from("y") => PolyType {
                    vars: vec![Tvar(0)],
                    cons: maplit::hashmap! { Tvar(0) => vec![Kind::Addable, Kind::Divisible]},
                    expr: MonoType::Fun(Box::new(types::Function {
                        req: maplit::hashmap! {
                            String::from("a") => MonoType::Var(Tvar(0)),
                            String::from("b") => MonoType::Var(Tvar(0)),
                        },
                        opt: HashMap::new(),
                        pipe: None,
                        retn: MonoType::Var(Tvar(0)),
                    })),
                },
                String::from("z") => PolyType {
                    vars: Vec::new(),
                    cons: HashMap::new(),
                    expr: MonoType::Int,
                },
            }
        );
    }

    #[test]
    fn duration_conversion_ok() {
        let t = vec![
            ast::Duration {
                magnitude: 1,
                unit: "y".to_string(),
            },
            ast::Duration {
                magnitude: 2,
                unit: "mo".to_string(),
            },
            ast::Duration {
                magnitude: 3,
                unit: "w".to_string(),
            },
            ast::Duration {
                magnitude: 4,
                unit: "m".to_string(),
            },
            ast::Duration {
                magnitude: 5,
                unit: "ns".to_string(),
            },
        ];
        let exp = (1.0 * YEARS + 2.0 * MONTHS) as i64 + 3 * WEEKS + 4 * MINUTES + 5 * NANOS;
        let got = convert_duration(&t).unwrap();
        assert_eq!(exp, got.num_nanoseconds().expect("should not overflow"));
    }

    #[test]
    fn duration_conversion_doubled_magnitude() {
        let t = vec![
            ast::Duration {
                magnitude: 1,
                unit: "y".to_string(),
            },
            ast::Duration {
                magnitude: 2,
                unit: "mo".to_string(),
            },
            ast::Duration {
                magnitude: 3,
                unit: "y".to_string(),
            },
        ];
        let exp = (4.0 * YEARS + 2.0 * MONTHS) as i64;
        let got = convert_duration(&t).unwrap();
        assert_eq!(exp, got.num_nanoseconds().expect("should not overflow"));
    }

    #[test]
    fn duration_conversion_negative() {
        let t = vec![
            ast::Duration {
                magnitude: -1,
                unit: "y".to_string(),
            },
            ast::Duration {
                magnitude: 2,
                unit: "mo".to_string(),
            },
            ast::Duration {
                magnitude: -3,
                unit: "w".to_string(),
            },
        ];
        let exp = (-1.0 * YEARS + 2.0 * MONTHS) as i64 - 3 * WEEKS;
        let got = convert_duration(&t).unwrap();
        assert_eq!(exp, got.num_nanoseconds().expect("should not overflow"));
    }

    #[test]
    fn duration_conversion_error() {
        let t = vec![
            ast::Duration {
                magnitude: -1,
                unit: "y".to_string(),
            },
            ast::Duration {
                magnitude: 2,
                unit: "--idk--".to_string(),
            },
            ast::Duration {
                magnitude: -3,
                unit: "w".to_string(),
            },
        ];
        let exp = "unrecognized magnitude for duration: --idk--";
        let got = convert_duration(&t).err().expect("should be an error");
        assert_eq!(exp, got.to_string());
    }
}
