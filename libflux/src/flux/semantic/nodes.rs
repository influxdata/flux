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
    import::Importer,
    infer::{Constraint, Constraints},
    sub::{Substitutable, Substitution},
    types::{Array, Function, Kind, MonoType, PolyType, Tvar},
};

use chrono::prelude::DateTime;
use chrono::FixedOffset;
use derivative::Derivative;
use std::collections::HashMap;
use std::fmt;
use std::vec::Vec;

// Result returned from the various 'infer' methods defined in this
// module. The result of inferring an expression or statment is an
// updated type environment and a set of type constraints to be solved.
pub type Result = std::result::Result<(Environment, Constraints), Error>;

#[derive(Debug)]
pub struct Error {
    pub msg: String,
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

impl From<String> for Error {
    fn from(msg: String) -> Error {
        Error { msg }
    }
}

impl From<Error> for String {
    fn from(err: Error) -> String {
        err.to_string()
    }
}

impl Error {
    fn undeclared_variable(name: String) -> Error {
        Error {
            msg: format!("undeclared variable {}", name),
        }
    }
    fn undefined_builtin(name: &str) -> Error {
        Error {
            msg: format!("builtin identifier {} not defined", name),
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
    Variable(Box<VariableAssgn>),
    Option(Box<OptionStmt>),
    Return(ReturnStmt),
    Test(Box<TestStmt>),
    Builtin(BuiltinStmt),
}

impl Statement {
    fn apply(self, sub: &Substitution) -> Self {
        match self {
            Statement::Expr(stmt) => Statement::Expr(stmt.apply(&sub)),
            Statement::Variable(stmt) => Statement::Variable(Box::new(stmt.apply(&sub))),
            Statement::Option(stmt) => Statement::Option(Box::new(stmt.apply(&sub))),
            Statement::Return(stmt) => Statement::Return(stmt.apply(&sub)),
            Statement::Test(stmt) => Statement::Test(Box::new(stmt.apply(&sub))),
            Statement::Builtin(stmt) => Statement::Builtin(stmt.apply(&sub)),
        }
    }
}

#[derive(Debug, PartialEq, Clone)]
pub enum Assignment {
    Variable(VariableAssgn),
    Member(MemberAssgn),
}

impl Assignment {
    fn apply(self, sub: &Substitution) -> Self {
        match self {
            Assignment::Variable(assign) => Assignment::Variable(assign.apply(&sub)),
            Assignment::Member(assign) => Assignment::Member(assign.apply(&sub)),
        }
    }
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
            Expression::Integer(lit) => lit.infer(env),
            Expression::Float(lit) => lit.infer(env),
            Expression::StringLit(lit) => lit.infer(env),
            Expression::Duration(lit) => lit.infer(env),
            Expression::Uint(lit) => lit.infer(env),
            Expression::Boolean(lit) => lit.infer(env),
            Expression::DateTime(lit) => lit.infer(env),
            Expression::Regexp(lit) => lit.infer(env),
        }
    }
    fn apply(self, sub: &Substitution) -> Self {
        match self {
            Expression::Identifier(e) => Expression::Identifier(e.apply(&sub)),
            Expression::Array(e) => Expression::Array(Box::new(e.apply(&sub))),
            Expression::Function(e) => Expression::Function(Box::new(e.apply(&sub))),
            Expression::Logical(e) => Expression::Logical(Box::new(e.apply(&sub))),
            Expression::Object(e) => Expression::Object(Box::new(e.apply(&sub))),
            Expression::Member(e) => Expression::Member(Box::new(e.apply(&sub))),
            Expression::Index(e) => Expression::Index(Box::new(e.apply(&sub))),
            Expression::Binary(e) => Expression::Binary(Box::new(e.apply(&sub))),
            Expression::Unary(e) => Expression::Unary(Box::new(e.apply(&sub))),
            Expression::Call(e) => Expression::Call(Box::new(e.apply(&sub))),
            Expression::Conditional(e) => Expression::Conditional(Box::new(e.apply(&sub))),
            Expression::StringExpr(e) => Expression::StringExpr(Box::new(e.apply(&sub))),
            Expression::Integer(lit) => Expression::Integer(lit.apply(&sub)),
            Expression::Float(lit) => Expression::Float(lit.apply(&sub)),
            Expression::StringLit(lit) => Expression::StringLit(lit.apply(&sub)),
            Expression::Duration(lit) => Expression::Duration(lit.apply(&sub)),
            Expression::Uint(lit) => Expression::Uint(lit.apply(&sub)),
            Expression::Boolean(lit) => Expression::Boolean(lit.apply(&sub)),
            Expression::DateTime(lit) => Expression::DateTime(lit.apply(&sub)),
            Expression::Regexp(lit) => Expression::Regexp(lit.apply(&sub)),
        }
    }
}

// Infer the types of a flux package
pub fn infer_pkg_types<T, S>(
    pkg: &mut Package,
    env: Environment,
    f: &mut Fresher,
    importer: &T,
    builtins: &S,
) -> std::result::Result<(Environment, Substitution), Error>
where
    T: Importer,
    S: Importer,
{
    let (env, cons) = pkg.infer(env, f, importer, builtins)?;
    Ok((env, infer::solve(&cons, &mut HashMap::new(), f)?))
}

pub fn infer_file<T, S>(
    file: &mut File,
    env: Environment,
    f: &mut Fresher,
    importer: &T,
    builtins: &S,
) -> Result
where
    T: Importer,
    S: Importer,
{
    file.infer(env, f, importer, builtins)
}

pub fn inject_pkg_types(pkg: Package, sub: &Substitution) -> Package {
    pkg.apply(&sub)
}

#[derive(Debug, PartialEq, Clone)]
pub struct Package {
    pub loc: ast::SourceLocation,

    pub package: String,
    pub files: Vec<File>,
}

impl Package {
    fn infer<T, S>(
        &mut self,
        env: Environment,
        f: &mut Fresher,
        importer: &T,
        builtins: &S,
    ) -> Result
    where
        T: Importer,
        S: Importer,
    {
        self.files
            .iter_mut()
            .try_fold((env, Constraints::empty()), |(env, rest), file| {
                let (env, cons) = file.infer(env, f, importer, builtins)?;
                Ok((env, cons + rest))
            })
    }
    fn apply(mut self, sub: &Substitution) -> Self {
        self.files = self
            .files
            .into_iter()
            .map(|file| file.apply(&sub))
            .collect();
        self
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
    fn infer<T, S>(
        &mut self,
        mut env: Environment,
        f: &mut Fresher,
        importer: &T,
        builtins: &S,
    ) -> Result
    where
        T: Importer,
        S: Importer,
    {
        let mut imports = Vec::with_capacity(self.imports.len());

        for dec in &self.imports {
            let path = &dec.path.value;

            let name = match &dec.alias {
                None => path.rsplitn(2, '/').collect::<Vec<&str>>()[0],
                Some(id) => &id.name[..],
            };

            imports.push(name);

            match importer.import(path) {
                Some(poly) => env.add(name.to_owned(), poly),
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
                            let env = stmt.infer(env, builtins)?;
                            Ok((env, rest))
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
    fn apply(mut self, sub: &Substitution) -> Self {
        self.body = self.body.into_iter().map(|stmt| stmt.apply(&sub)).collect();
        self
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
    fn apply(mut self, sub: &Substitution) -> Self {
        self.assignment = self.assignment.apply(&sub);
        self
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct BuiltinStmt {
    pub loc: ast::SourceLocation,

    pub id: Identifier,
}

impl BuiltinStmt {
    fn infer<I: Importer>(
        &self,
        mut env: Environment,
        importer: &I,
    ) -> std::result::Result<Environment, Error> {
        if let Some(ty) = importer.import(&self.id.name) {
            env.add(self.id.name.clone(), ty);
            Ok(env)
        } else {
            Err(Error::undefined_builtin(&self.id.name))
        }
    }
    fn apply(self, _: &Substitution) -> Self {
        self
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
    fn apply(mut self, sub: &Substitution) -> Self {
        self.assignment = self.assignment.apply(&sub);
        self
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
    fn apply(mut self, sub: &Substitution) -> Self {
        self.expression = self.expression.apply(&sub);
        self
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct ReturnStmt {
    pub loc: ast::SourceLocation,

    pub argument: Expression,
}

impl ReturnStmt {
    #[allow(dead_code)]
    fn infer(&mut self, env: Environment, f: &mut Fresher) -> Result {
        self.argument.infer(env, f)
    }
    fn apply(mut self, sub: &Substitution) -> Self {
        self.argument = self.argument.apply(&sub);
        self
    }
}

#[derive(Debug, Derivative, Clone)]
#[derivative(PartialEq)]
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
        env.add(String::from(&self.id.name), p);
        Ok((env, constraints))
    }
    fn apply(mut self, sub: &Substitution) -> Self {
        self.init = self.init.apply(&sub);
        self
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct MemberAssgn {
    pub loc: ast::SourceLocation,

    pub member: MemberExpr,
    pub init: Expression,
}

impl MemberAssgn {
    fn apply(mut self, sub: &Substitution) -> Self {
        self.member = self.member.apply(&sub);
        self.init = self.init.apply(&sub);
        self
    }
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
        Ok((env, Constraints::from(constraints)))
    }
    fn apply(mut self, sub: &Substitution) -> Self {
        self.typ = self.typ.apply(&sub);
        self.parts = self
            .parts
            .into_iter()
            .map(|part| part.apply(&sub))
            .collect();
        self
    }
}

#[derive(Debug, PartialEq, Clone)]
pub enum StringExprPart {
    Text(TextPart),
    Interpolated(InterpolatedPart),
}

impl StringExprPart {
    fn apply(self, sub: &Substitution) -> Self {
        match self {
            StringExprPart::Interpolated(part) => StringExprPart::Interpolated(part.apply(&sub)),
            StringExprPart::Text(_) => self,
        }
    }
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

impl InterpolatedPart {
    fn apply(mut self, sub: &Substitution) -> Self {
        self.expression = self.expression.apply(&sub);
        self
    }
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
        Ok((env, cons.into()))
    }
    fn apply(mut self, sub: &Substitution) -> Self {
        self.typ = self.typ.apply(&sub);
        self.elements = self
            .elements
            .into_iter()
            .map(|element| element.apply(&sub))
            .collect();
        self
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
        Ok((env, cons))
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
            if p.default.is_some() {
                ds.push(p);
            };
        }
        ds
    }
    fn apply(mut self, sub: &Substitution) -> Self {
        self.typ = self.typ.apply(&sub);
        self.params = self
            .params
            .into_iter()
            .map(|param| param.apply(&sub))
            .collect();
        self.body = self.body.apply(&sub);
        self
    }
}

// Block represents a function block and is equivalent to a let-expression
// in other functional languages.
//
// Functions must evaluate to a value in Flux. In other words, a function
// must always have a return value. This means a function block is by
// definition an expression.
//
// A function block is an expression that evaluates to the argument of
// its terminating ReturnStmt.
#[derive(Debug, PartialEq, Clone)]
pub enum Block {
    Variable(Box<VariableAssgn>, Box<Block>),
    Expr(ExprStmt, Box<Block>),
    Return(ReturnStmt),
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
            Block::Variable(assign, _) => &assign.loc,
            Block::Expr(es, _) => es.expression.loc(),
            Block::Return(ret) => &ret.loc,
        }
    }
    pub fn type_of(&self) -> &MonoType {
        let mut n = self;
        loop {
            n = match n {
                Block::Variable(_, b) => b.as_ref(),
                Block::Expr(_, b) => b.as_ref(),
                Block::Return(r) => return r.argument.type_of(),
            }
        }
    }
    fn apply(self, sub: &Substitution) -> Self {
        match self {
            Block::Variable(assign, next) => {
                Block::Variable(Box::new(assign.apply(&sub)), Box::new(next.apply(&sub)))
            }
            Block::Expr(es, next) => Block::Expr(es.apply(&sub), Box::new(next.apply(&sub))),
            Block::Return(e) => Block::Return(e.apply(&sub)),
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

impl FunctionParameter {
    fn apply(mut self, sub: &Substitution) -> Self {
        match self.default {
            Some(e) => {
                self.default = Some(e.apply(&sub));
                self
            }
            None => self,
        }
    }
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
        Ok((env, lcons + rcons + cons))
    }
    fn apply(mut self, sub: &Substitution) -> Self {
        self.typ = self.typ.apply(&sub);
        self.left = self.left.apply(&sub);
        self.right = self.right.apply(&sub);
        self
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
    fn apply(mut self, sub: &Substitution) -> Self {
        self.typ = self.typ.apply(&sub);
        self.callee = self.callee.apply(&sub);
        self.arguments = self
            .arguments
            .into_iter()
            .map(|arg| arg.apply(&sub))
            .collect();
        match self.pipe {
            Some(e) => {
                self.pipe = Some(e.apply(&sub));
                self
            }
            None => self,
        }
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
        Ok((env, cons))
    }
    fn apply(mut self, sub: &Substitution) -> Self {
        self.typ = self.typ.apply(&sub);
        self.test = self.test.apply(&sub);
        self.consequent = self.consequent.apply(&sub);
        self.alternate = self.alternate.apply(&sub);
        self
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
        Ok((env, cons))
    }
    fn apply(mut self, sub: &Substitution) -> Self {
        self.typ = self.typ.apply(&sub);
        self.left = self.left.apply(&sub);
        self.right = self.right.apply(&sub);
        self
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
    fn apply(mut self, sub: &Substitution) -> Self {
        self.typ = self.typ.apply(&sub);
        self.object = self.object.apply(&sub);
        self
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
        Ok((env, cons))
    }
    fn apply(mut self, sub: &Substitution) -> Self {
        self.typ = self.typ.apply(&sub);
        self.array = self.array.apply(&sub);
        self.index = self.index.apply(&sub);
        self
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
    fn apply(mut self, sub: &Substitution) -> Self {
        self.typ = self.typ.apply(&sub);
        if let Some(e) = self.with {
            self.with = Some(e.apply(&sub));
        }
        self.properties = self
            .properties
            .into_iter()
            .map(|prop| prop.apply(&sub))
            .collect();
        self
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
            ast::Operator::AdditionOperator | ast::Operator::SubtractionOperator => {
                Constraints::from(vec![
                    Constraint::Equal(self.argument.type_of().clone(), self.typ.clone()),
                    Constraint::Kind(self.argument.type_of().clone(), Kind::Negatable),
                ])
            }
            _ => return Err(Error::unsupported_unary_operator(&self.operator)),
        };
        Ok((env, acons + cons))
    }
    fn apply(mut self, sub: &Substitution) -> Self {
        self.typ = self.typ.apply(&sub);
        self.argument = self.argument.apply(&sub);
        self
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct Property {
    pub loc: ast::SourceLocation,

    pub key: Identifier,
    pub value: Expression,
}

impl Property {
    fn apply(mut self, sub: &Substitution) -> Self {
        self.value = self.value.apply(&sub);
        self
    }
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
    fn apply(mut self, sub: &Substitution) -> Self {
        self.typ = self.typ.apply(&sub);
        self
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
    fn infer(&self, env: Environment) -> Result {
        infer_literal(env, &self.typ, MonoType::Bool)
    }
    fn apply(mut self, sub: &Substitution) -> Self {
        self.typ = self.typ.apply(&sub);
        self
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
    fn infer(&self, env: Environment) -> Result {
        infer_literal(env, &self.typ, MonoType::Int)
    }
    fn apply(mut self, sub: &Substitution) -> Self {
        self.typ = self.typ.apply(&sub);
        self
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
    fn infer(&self, env: Environment) -> Result {
        infer_literal(env, &self.typ, MonoType::Float)
    }
    fn apply(mut self, sub: &Substitution) -> Self {
        self.typ = self.typ.apply(&sub);
        self
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
    fn infer(&self, env: Environment) -> Result {
        infer_literal(env, &self.typ, MonoType::Regexp)
    }
    fn apply(mut self, sub: &Substitution) -> Self {
        self.typ = self.typ.apply(&sub);
        self
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
    fn infer(&self, env: Environment) -> Result {
        infer_literal(env, &self.typ, MonoType::String)
    }
    fn apply(mut self, sub: &Substitution) -> Self {
        self.typ = self.typ.apply(&sub);
        self
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
    fn infer(&self, env: Environment) -> Result {
        infer_literal(env, &self.typ, MonoType::Uint)
    }
    fn apply(mut self, sub: &Substitution) -> Self {
        self.typ = self.typ.apply(&sub);
        self
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
    fn infer(&self, env: Environment) -> Result {
        infer_literal(env, &self.typ, MonoType::Time)
    }
    fn apply(mut self, sub: &Substitution) -> Self {
        self.typ = self.typ.apply(&sub);
        self
    }
}

// Duration is a struct that keeps track of time in months and nanoseconds.
// Months and nanoseconds must be positive values. Negative is a bool to indicate
// whether the magnitude of durations converted from the AST have a positive or
// negative value
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(rename = "Duration")]
pub struct Duration {
    pub months: i64,
    pub nanoseconds: i64,
    pub negative: bool,
}

// DurationLit is a pair consisting of length of time and the unit of time measured.
// It is the atomic unit from which all duration literals are composed.
#[derive(Derivative)]
#[derivative(Debug, PartialEq, Clone)]
pub struct DurationLit {
    pub loc: ast::SourceLocation,
    #[derivative(PartialEq = "ignore")]
    pub typ: MonoType,
    #[derivative(PartialEq = "ignore")]
    pub value: Duration,
}

impl DurationLit {
    fn infer(&self, env: Environment) -> Result {
        infer_literal(env, &self.typ, MonoType::Duration)
    }
    fn apply(mut self, sub: &Substitution) -> Self {
        self.typ = self.typ.apply(&sub);
        self
    }
}

fn infer_literal(env: Environment, typ: &MonoType, is: MonoType) -> Result {
    let constraints = Constraints::from(vec![Constraint::Equal(typ.clone(), is)]);
    Ok((env, constraints))
}

// The following durations have nanosecond base units
const NANOS: i64 = 1;
const MICROS: i64 = NANOS * 1000;
const MILLIS: i64 = MICROS * 1000;
const SECONDS: i64 = MILLIS * 1000;
const MINUTES: i64 = SECONDS * 60;
const HOURS: i64 = MINUTES * 60;
const DAYS: i64 = HOURS * 24;
const WEEKS: i64 = DAYS * 7;

// The following durations have month base units
const MONTHS: i64 = 1;
const YEARS: i64 = MONTHS * 12;

pub fn convert_duration(ast_dur: &[ast::Duration]) -> std::result::Result<Duration, String> {
    if ast_dur.is_empty() {
        return Err(String::from(
            "AST duration vector must contain at least one duration value",
        ));
    };

    let negative = ast_dur[0].magnitude.is_negative();

    let (nanoseconds, months) = ast_dur.iter().try_fold((0i64, 0i64), |acc, d| {
        if (d.magnitude.is_negative() && !negative) || (!d.magnitude.is_negative() && negative) {
            return Err("all values in AST duration vector must have the same sign");
        }

        match d.unit.as_str() {
            "y" => Ok((acc.0, acc.1 + d.magnitude * YEARS)),
            "mo" => Ok((acc.0, acc.1 + d.magnitude * MONTHS)),
            "w" => Ok((acc.0 + d.magnitude * WEEKS, acc.1)),
            "d" => Ok((acc.0 + d.magnitude * DAYS, acc.1)),
            "h" => Ok((acc.0 + d.magnitude * HOURS, acc.1)),
            "m" => Ok((acc.0 + d.magnitude * MINUTES, acc.1)),
            "s" => Ok((acc.0 + d.magnitude * SECONDS, acc.1)),
            "ms" => Ok((acc.0 + d.magnitude * MILLIS, acc.1)),
            "us" | "µs" => Ok((acc.0 + d.magnitude * MICROS, acc.1)),
            "ns" => Ok((acc.0 + d.magnitude * NANOS, acc.1)),
            _ => Err("unrecognized magnitude for duration"),
        }
    })?;

    let nanoseconds = nanoseconds.abs();
    let months = months.abs();

    Ok(Duration {
        months,
        nanoseconds,
        negative,
    })
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::ast;
    use crate::semantic::types::{MonoType, Tvar};
    use crate::semantic::walk::{walk, Node};
    use maplit::hashmap;
    use std::rc::Rc;

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
        let expect_nano = 3 * WEEKS + 4 * MINUTES + 5 * NANOS;
        let expect_months = 1 * YEARS + 2 * MONTHS;

        let got = convert_duration(&t).unwrap();
        assert_eq!(expect_nano, got.nanoseconds);
        assert_eq!(expect_months, got.months);
        assert_eq!(false, got.negative);
    }

    #[test]
    fn duration_conversion_same_magnitude_twice() {
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
        let expect_nano = 0;
        let expect_months = 4 * YEARS + 2 * MONTHS;

        let got = convert_duration(&t).unwrap();
        assert_eq!(expect_nano, got.nanoseconds);
        assert_eq!(expect_months, got.months);
        assert_eq!(false, got.negative);
    }

    #[test]
    fn duration_conversion_negative_ok() {
        let t = vec![
            ast::Duration {
                magnitude: -1,
                unit: "y".to_string(),
            },
            ast::Duration {
                magnitude: -2,
                unit: "mo".to_string(),
            },
            ast::Duration {
                magnitude: -3,
                unit: "w".to_string(),
            },
        ];
        let expect_months = (-1 * YEARS + (-2 * MONTHS)).abs();
        let expect_nano = (-3 * WEEKS).abs();

        let got = convert_duration(&t).unwrap();
        assert_eq!(expect_nano, got.nanoseconds);
        assert_eq!(expect_months, got.months);
        assert_eq!(true, got.negative);
    }

    #[test]
    fn duration_conversion_unit_error() {
        let t = vec![
            ast::Duration {
                magnitude: -1,
                unit: "y".to_string(),
            },
            ast::Duration {
                magnitude: -2,
                unit: "--idk--".to_string(),
            },
            ast::Duration {
                magnitude: -3,
                unit: "w".to_string(),
            },
        ];
        let exp = "unrecognized magnitude for duration";
        let got = convert_duration(&t).err().expect("should be an error");
        assert_eq!(exp, got.to_string());
    }

    #[test]
    fn duration_conversion_different_signs_error() {
        let t = vec![
            ast::Duration {
                magnitude: -1,
                unit: "y".to_string(),
            },
            ast::Duration {
                magnitude: 2,
                unit: "ns".to_string(),
            },
            ast::Duration {
                magnitude: -3,
                unit: "w".to_string(),
            },
        ];
        let exp = "all values in AST duration vector must have the same sign";
        let got = convert_duration(&t).err().expect("should be an error");
        assert_eq!(exp, got.to_string());
    }

    #[test]
    fn duration_conversion_empty_error() {
        let t = Vec::new();
        let exp = "AST duration vector must contain at least one duration value";
        let got = convert_duration(&t).err().expect("should be an error");
        assert_eq!(exp, got.to_string());
    }

    #[test]
    fn test_inject_types() {
        let b = ast::BaseNode::default();
        let pkg = Package {
            loc: b.location.clone(),
            package: "main".to_string(),
            files: vec![File {
                loc: b.location.clone(),
                package: None,
                imports: Vec::new(),
                body: vec![
                    Statement::Variable(Box::new(VariableAssgn::new(
                        Identifier {
                            loc: b.location.clone(),
                            name: "f".to_string(),
                        },
                        Expression::Function(Box::new(FunctionExpr {
                            loc: b.location.clone(),
                            typ: MonoType::Var(Tvar(0)),
                            params: vec![
                                FunctionParameter {
                                    loc: b.location.clone(),
                                    is_pipe: true,
                                    key: Identifier {
                                        loc: b.location.clone(),
                                        name: "piped".to_string(),
                                    },
                                    default: None,
                                },
                                FunctionParameter {
                                    loc: b.location.clone(),
                                    is_pipe: false,
                                    key: Identifier {
                                        loc: b.location.clone(),
                                        name: "a".to_string(),
                                    },
                                    default: None,
                                },
                            ],
                            body: Block::Return(ReturnStmt {
                                loc: b.location.clone(),
                                argument: Expression::Binary(Box::new(BinaryExpr {
                                    loc: b.location.clone(),
                                    typ: MonoType::Var(Tvar(1)),
                                    operator: ast::Operator::AdditionOperator,
                                    left: Expression::Identifier(IdentifierExpr {
                                        loc: b.location.clone(),
                                        typ: MonoType::Var(Tvar(2)),
                                        name: "a".to_string(),
                                    }),
                                    right: Expression::Identifier(IdentifierExpr {
                                        loc: b.location.clone(),
                                        typ: MonoType::Var(Tvar(3)),
                                        name: "piped".to_string(),
                                    }),
                                })),
                            }),
                        })),
                        b.location.clone(),
                    ))),
                    Statement::Expr(ExprStmt {
                        loc: b.location.clone(),
                        expression: Expression::Call(Box::new(CallExpr {
                            loc: b.location.clone(),
                            typ: MonoType::Var(Tvar(4)),
                            pipe: Some(Expression::Integer(IntegerLit {
                                loc: b.location.clone(),
                                typ: MonoType::Var(Tvar(5)),
                                value: 3,
                            })),
                            callee: Expression::Identifier(IdentifierExpr {
                                loc: b.location.clone(),
                                typ: MonoType::Var(Tvar(6)),
                                name: "f".to_string(),
                            }),
                            arguments: vec![Property {
                                loc: b.location.clone(),
                                key: Identifier {
                                    loc: b.location.clone(),
                                    name: "a".to_string(),
                                },
                                value: Expression::Integer(IntegerLit {
                                    loc: b.location.clone(),
                                    typ: MonoType::Var(Tvar(7)),
                                    value: 2,
                                }),
                            }],
                        })),
                    }),
                ],
            }],
        };
        let sub: Substitution = hashmap! {
            Tvar(0) => MonoType::Int,
            Tvar(1) => MonoType::Int,
            Tvar(2) => MonoType::Int,
            Tvar(3) => MonoType::Int,
            Tvar(4) => MonoType::Int,
            Tvar(5) => MonoType::Int,
            Tvar(6) => MonoType::Int,
            Tvar(7) => MonoType::Int,
        }
        .into();
        let pkg = inject_pkg_types(pkg, &sub);
        let mut no_types_checked = 0;
        walk(
            &mut |node: Rc<Node>| {
                let typ = node.type_of();
                if let Some(typ) = typ {
                    assert_eq!(typ, &MonoType::Int);
                    no_types_checked += 1;
                }
            },
            Rc::new(Node::Package(&pkg)),
        );
        assert_eq!(no_types_checked, 8);
    }
}
