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
    types::{Kind, MonoType, PolyType, Tvar},
};

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

// Infer the types of a flux package
pub fn infer_pkg_types(
    pkg: &mut Package,
    env: Environment,
    f: &mut Fresher,
) -> std::result::Result<(Environment, Substitution), Error> {
    let (env, cons) = pkg.infer(env, f)?;
    Ok((env, infer::solve(&cons, &mut HashMap::new())?))
}

#[derive(Debug, PartialEq, Clone)]
pub struct Package {
    pub loc: ast::SourceLocation,

    pub package: String,
    pub files: Vec<File>,
}

impl Package {
    fn infer(&mut self, env: Environment, f: &mut Fresher) -> Result {
        self.files
            .iter_mut()
            .try_fold((env, Constraints::empty()), |(env, rest), file| {
                let (env, cons) = file.infer(env, f)?;
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
    fn infer(&mut self, mut env: Environment, f: &mut Fresher) -> Result {
        // TODO: add imported types to the type environment
        self.body.iter_mut().try_fold(
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
        )
        // TODO: remove imported names from the type environment
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
        let sub = infer::solve(&cons, &mut HashMap::new())?;
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
    free: Vec<Tvar>,

    #[derivative(PartialEq = "ignore")]
    cons: HashMap<Tvar, Vec<Kind>>,

    pub loc: ast::SourceLocation,

    pub id: Identifier,
    pub init: Expression,
}

impl VariableAssgn {
    pub fn new(id: Identifier, init: Expression, loc: ast::SourceLocation) -> VariableAssgn {
        VariableAssgn {
            free: Vec::new(),
            cons: HashMap::new(),
            loc,
            id,
            init,
        }
    }
    pub fn poly_type_of(&self) -> PolyType {
        PolyType {
            free: self.free.clone(),
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
        let sub = infer::solve(&constraints, &mut kinds)?;

        // Apply substitution to the type environment
        let mut env = env.apply(&sub);

        let t = self.init.type_of().clone().apply(&sub);
        let p = infer::generalize(&env, &kinds, t);

        // Update variable assignment nodes with the free vars
        // and kind constraints obtained from generalization.
        //
        // Note these variables are fixed after generalization
        // and so it is safe to update these nodes in place.
        self.free = p.free.clone();
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
        unimplemented!();
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
    fn infer(&self, env: Environment, f: &mut Fresher) -> Result {
        unimplemented!();
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
    fn infer(&mut self, env: Environment, f: &mut Fresher) -> Result {
        unimplemented!();
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
    fn infer(&self, env: Environment, f: &mut Fresher) -> Result {
        unimplemented!();
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
    fn infer(&self, env: Environment, f: &mut Fresher) -> Result {
        unimplemented!();
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
    fn infer(&self, env: Environment, f: &mut Fresher) -> Result {
        unimplemented!();
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
    fn infer(&self, env: Environment, f: &mut Fresher) -> Result {
        unimplemented!();
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
    fn infer(&self, env: Environment, f: &mut Fresher) -> Result {
        unimplemented!();
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
    fn infer(&self, env: Environment, f: &mut Fresher) -> Result {
        unimplemented!();
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
    fn infer(&self, env: Environment, f: &mut Fresher) -> Result {
        unimplemented!();
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
    fn infer(&self, env: Environment, f: &mut Fresher) -> Result {
        unimplemented!();
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
        unimplemented!();
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
        unimplemented!();
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
        unimplemented!();
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
        unimplemented!();
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
        unimplemented!();
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
        unimplemented!();
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
        unimplemented!();
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
        unimplemented!();
    }
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
    use crate::semantic::analyze::analyze;
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
                free: Vec::new(),
                cons: HashMap::new(),
                expr: MonoType::Int,
            },
            // f = (a, b) => 2 * (a + b)
            String::from("f") => PolyType {
                free: vec![Tvar(0)],
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
        let mut pkg = analyze(ast, &mut f).unwrap();

        let (env, _) = infer_pkg_types(&mut pkg, env, &mut f).unwrap();

        let normalized: HashMap<String, PolyType> = env
            .values
            .into_iter()
            .map(|(k, v)| (k, v.normalize(&mut Fresher::new())))
            .collect();

        assert_eq!(
            normalized,
            maplit::hashmap! {
                String::from("f") => PolyType {
                    free: vec![Tvar(0)],
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
                    free: Vec::new(),
                    cons: HashMap::new(),
                    expr: MonoType::Int,
                },
                String::from("x") => PolyType {
                    free: vec![Tvar(0)],
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
                    free: vec![Tvar(0)],
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
                    free: Vec::new(),
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
