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
use crate::semantic::types;

use chrono::prelude::DateTime;
use chrono::Duration;
use chrono::FixedOffset;
use derivative::Derivative;
use std::vec::Vec;

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

#[derive(Debug, PartialEq, Clone)]
pub struct Package {
    pub loc: ast::SourceLocation,

    pub package: String,
    pub files: Vec<File>,
}

#[derive(Debug, PartialEq, Clone)]
pub struct File {
    pub loc: ast::SourceLocation,

    pub package: Option<PackageClause>,
    pub imports: Vec<ImportDeclaration>,
    pub body: Vec<Statement>,
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
pub struct Block {
    pub loc: ast::SourceLocation,

    pub body: Vec<Statement>,
}

impl Block {
    fn return_statement(&self) -> &ReturnStmt {
        let len = self.body.len();
        let last = self
            .body
            .get(len - 1)
            .expect("body must have at least one statement");
        let last: Option<&ReturnStmt> = match last {
            Statement::Return(rs) => Some(rs),
            _ => None,
        };
        last.expect("last statement must be a return statement")
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct OptionStmt {
    pub loc: ast::SourceLocation,

    pub assignment: Assignment,
}

#[derive(Debug, PartialEq, Clone)]
pub struct BuiltinStmt {
    pub loc: ast::SourceLocation,

    pub id: Identifier,
}

#[derive(Debug, PartialEq, Clone)]
pub struct TestStmt {
    pub loc: ast::SourceLocation,

    pub assignment: VariableAssgn,
}

#[derive(Debug, PartialEq, Clone)]
pub struct ExprStmt {
    pub loc: ast::SourceLocation,

    pub expression: Expression,
}

#[derive(Debug, PartialEq, Clone)]
pub struct ReturnStmt {
    pub loc: ast::SourceLocation,

    pub argument: Expression,
}

#[derive(Debug, PartialEq, Clone)]
pub struct VariableAssgn {
    pub loc: ast::SourceLocation,

    pub id: Identifier,
    pub init: Expression,
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
    pub typ: types::MonoType,

    pub parts: Vec<StringExprPart>,
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
    pub typ: types::MonoType,

    pub elements: Vec<Expression>,
}

// FunctionExpr represents the definition of a function
#[derive(Derivative)]
#[derivative(Debug, PartialEq, Clone)]
pub struct FunctionExpr {
    pub loc: ast::SourceLocation,
    #[derivative(PartialEq = "ignore")]
    pub typ: types::MonoType,

    pub params: Vec<FunctionParameter>,
    pub body: Block,
}

impl FunctionExpr {
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
    pub typ: types::MonoType,

    pub operator: ast::Operator,
    pub left: Expression,
    pub right: Expression,
}

#[derive(Derivative)]
#[derivative(Debug, PartialEq, Clone)]
pub struct CallExpr {
    pub loc: ast::SourceLocation,
    #[derivative(PartialEq = "ignore")]
    pub typ: types::MonoType,

    pub callee: Expression,
    pub arguments: Vec<Property>,
    pub pipe: Option<Expression>,
}

#[derive(Derivative)]
#[derivative(Debug, PartialEq, Clone)]
pub struct ConditionalExpr {
    pub loc: ast::SourceLocation,
    #[derivative(PartialEq = "ignore")]
    pub typ: types::MonoType,

    pub test: Expression,
    pub consequent: Expression,
    pub alternate: Expression,
}

#[derive(Derivative)]
#[derivative(Debug, PartialEq, Clone)]
pub struct LogicalExpr {
    pub loc: ast::SourceLocation,
    #[derivative(PartialEq = "ignore")]
    pub typ: types::MonoType,

    pub operator: ast::LogicalOperator,
    pub left: Expression,
    pub right: Expression,
}

#[derive(Derivative)]
#[derivative(Debug, PartialEq, Clone)]
pub struct MemberExpr {
    pub loc: ast::SourceLocation,
    #[derivative(PartialEq = "ignore")]
    pub typ: types::MonoType,

    pub object: Expression,
    pub property: String,
}

#[derive(Derivative)]
#[derivative(Debug, PartialEq, Clone)]
pub struct IndexExpr {
    pub loc: ast::SourceLocation,
    #[derivative(PartialEq = "ignore")]
    pub typ: types::MonoType,

    pub array: Expression,
    pub index: Expression,
}

#[derive(Derivative)]
#[derivative(Debug, PartialEq, Clone)]
pub struct ObjectExpr {
    pub loc: ast::SourceLocation,
    #[derivative(PartialEq = "ignore")]
    pub typ: types::MonoType,

    pub with: Option<IdentifierExpr>,
    pub properties: Vec<Property>,
}

#[derive(Derivative)]
#[derivative(Debug, PartialEq, Clone)]
pub struct UnaryExpr {
    pub loc: ast::SourceLocation,
    #[derivative(PartialEq = "ignore")]
    pub typ: types::MonoType,

    pub operator: ast::Operator,
    pub argument: Expression,
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
    pub typ: types::MonoType,

    pub name: String,
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
    pub typ: types::MonoType,

    pub value: bool,
}

#[derive(Derivative)]
#[derivative(Debug, PartialEq, Clone)]
pub struct IntegerLit {
    pub loc: ast::SourceLocation,
    #[derivative(PartialEq = "ignore")]
    pub typ: types::MonoType,

    pub value: i64,
}

#[derive(Derivative)]
#[derivative(Debug, PartialEq, Clone)]
pub struct FloatLit {
    pub loc: ast::SourceLocation,
    #[derivative(PartialEq = "ignore")]
    pub typ: types::MonoType,

    pub value: f64,
}

#[derive(Derivative)]
#[derivative(Debug, PartialEq, Clone)]
pub struct RegexpLit {
    pub loc: ast::SourceLocation,
    #[derivative(PartialEq = "ignore")]
    pub typ: types::MonoType,

    // TODO(affo): should this be a compiled regexp?
    pub value: String,
}

#[derive(Derivative)]
#[derivative(Debug, PartialEq, Clone)]
pub struct StringLit {
    pub loc: ast::SourceLocation,
    #[derivative(PartialEq = "ignore")]
    pub typ: types::MonoType,

    pub value: String,
}

#[derive(Derivative)]
#[derivative(Debug, PartialEq, Clone)]
pub struct UintLit {
    pub loc: ast::SourceLocation,
    #[derivative(PartialEq = "ignore")]
    pub typ: types::MonoType,

    pub value: u64,
}

#[derive(Derivative)]
#[derivative(Debug, PartialEq, Clone)]
pub struct DateTimeLit {
    pub loc: ast::SourceLocation,
    #[derivative(PartialEq = "ignore")]
    pub typ: types::MonoType,

    pub value: DateTime<FixedOffset>,
}

#[derive(Derivative)]
#[derivative(Debug, PartialEq, Clone)]
pub struct DurationLit {
    pub loc: ast::SourceLocation,
    #[derivative(PartialEq = "ignore")]
    pub typ: types::MonoType,

    pub value: Duration,
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
pub fn convert_duration(duration: &Vec<ast::Duration>) -> Result<Duration, String> {
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
