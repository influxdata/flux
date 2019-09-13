extern crate chrono;
#[macro_use]
extern crate serde_derive;
extern crate serde_aux;

use std::collections::HashMap;
use std::fmt;
use std::str::FromStr;
use std::vec::Vec;

use scanner;

use chrono::prelude::DateTime;
use chrono::FixedOffset;

use serde::de::{Deserialize, Deserializer, Error, Visitor};
use serde::ser::{Serialize, SerializeSeq, Serializer};
use serde_aux::prelude::*;

// Position is the AST counterpart of Scanner's Position.
// It adds serde capabilities.
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct Position {
    pub line: u32,
    pub column: u32,
}

impl Position {
    pub fn is_valid(&self) -> bool {
        self.line > 0 && self.column > 0
    }

    pub fn invalid() -> Self {
        Position { line: 0, column: 0 }
    }
}

impl From<&scanner::Position> for Position {
    fn from(item: &scanner::Position) -> Self {
        Position {
            line: item.line,
            column: item.column,
        }
    }
}

impl From<&Position> for scanner::Position {
    fn from(item: &Position) -> Self {
        scanner::Position {
            line: item.line,
            column: item.column,
        }
    }
}

impl Default for Position {
    fn default() -> Self {
        return Self::invalid();
    }
}

// SourceLocation represents the location of a node in the AST
#[derive(Debug, Default, PartialEq, Clone, Serialize, Deserialize)]
pub struct SourceLocation {
    pub file: Option<String>,   // File is the optional file name.
    pub start: Position,        // Start is the location in the source the node starts.
    pub end: Position,          // End is the location in the source the node ends.
    pub source: Option<String>, // Source is optional raw source.
}

impl SourceLocation {
    pub fn is_valid(&self) -> bool {
        self.start.is_valid() && self.end.is_valid()
    }
}

impl fmt::Display for SourceLocation {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        let fname = match &self.file {
            Some(s) => s.clone(),
            None => "".to_string(),
        };
        write!(
            f,
            "{}@{}:{}-{}:{}",
            fname, self.start.line, self.start.column, self.end.line, self.end.column
        )
    }
}

// serialize_to_string serializes an object that implements ToString to its string representation.
fn serialize_to_string<T, S>(field: &T, ser: S) -> Result<S::Ok, S::Error>
where
    S: Serializer,
    T: ToString,
{
    let s = field.to_string();
    ser.serialize_str(s.as_str())
}

#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(untagged)]
pub enum Expression {
    Idt(Identifier),

    Arr(Box<ArrayExpression>),
    Fun(Box<FunctionExpression>),
    Log(Box<LogicalExpression>),
    Obj(Box<ObjectExpression>),
    Mem(Box<MemberExpression>),
    Idx(Box<IndexExpression>),
    Bin(Box<BinaryExpression>),
    Un(Box<UnaryExpression>),
    Pipe(Box<PipeExpression>),
    Call(Box<CallExpression>),
    Cond(Box<ConditionalExpression>),
    StringExp(Box<StringExpression>),
    Paren(Box<ParenExpression>),

    Int(IntegerLiteral),
    Flt(FloatLiteral),
    Str(StringLiteral),
    Dur(DurationLiteral),
    Uint(UnsignedIntegerLiteral),
    Bool(BooleanLiteral),
    Time(DateTimeLiteral),
    Regexp(RegexpLiteral),
    PipeLit(PipeLiteral),

    Bad(Box<BadExpression>),
}

impl Expression {
    // `base` is an utility method that returns the BaseNode for an Expression.
    pub fn base(&self) -> &BaseNode {
        match self {
            Expression::Idt(wrapped) => &wrapped.base,
            Expression::Arr(wrapped) => &wrapped.base,
            Expression::Fun(wrapped) => &wrapped.base,
            Expression::Log(wrapped) => &wrapped.base,
            Expression::Obj(wrapped) => &wrapped.base,
            Expression::Mem(wrapped) => &wrapped.base,
            Expression::Idx(wrapped) => &wrapped.base,
            Expression::Bin(wrapped) => &wrapped.base,
            Expression::Un(wrapped) => &wrapped.base,
            Expression::Pipe(wrapped) => &wrapped.base,
            Expression::Call(wrapped) => &wrapped.base,
            Expression::Cond(wrapped) => &wrapped.base,
            Expression::Int(wrapped) => &wrapped.base,
            Expression::Flt(wrapped) => &wrapped.base,
            Expression::Str(wrapped) => &wrapped.base,
            Expression::Dur(wrapped) => &wrapped.base,
            Expression::Uint(wrapped) => &wrapped.base,
            Expression::Bool(wrapped) => &wrapped.base,
            Expression::Time(wrapped) => &wrapped.base,
            Expression::Regexp(wrapped) => &wrapped.base,
            Expression::PipeLit(wrapped) => &wrapped.base,
            Expression::Bad(wrapped) => &wrapped.base,
            Expression::StringExp(wrapped) => &wrapped.base,
            Expression::Paren(wrapped) => &wrapped.base,
        }
    }
}

#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(untagged)]
pub enum Statement {
    Expr(ExpressionStatement),
    Var(VariableAssignment),
    Opt(OptionStatement),
    Ret(ReturnStatement),
    Bad(BadStatement),
    Test(TestStatement),
    Built(BuiltinStatement),
}

impl Statement {
    // `base` is an utility method that returns the BaseNode for a Statement.
    pub fn base(&self) -> &BaseNode {
        match self {
            Statement::Expr(wrapped) => &wrapped.base,
            Statement::Var(wrapped) => &wrapped.base,
            Statement::Opt(wrapped) => &wrapped.base,
            Statement::Ret(wrapped) => &wrapped.base,
            Statement::Bad(wrapped) => &wrapped.base,
            Statement::Test(wrapped) => &wrapped.base,
            Statement::Built(wrapped) => &wrapped.base,
        }
    }
}

#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(untagged)]
pub enum Assignment {
    Variable(VariableAssignment),
    Member(MemberAssignment),
}

impl Assignment {
    // `base` is an utility method that returns the BaseNode for an Assignment.
    pub fn base(&self) -> &BaseNode {
        match self {
            Assignment::Variable(wrapped) => &wrapped.base,
            Assignment::Member(wrapped) => &wrapped.base,
        }
    }
}

#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(untagged)]
pub enum PropertyKey {
    Identifier(Identifier),
    StringLiteral(StringLiteral),
}

impl PropertyKey {
    // `base` is an utility method that returns the BaseNode for a PropertyKey.
    pub fn base(&self) -> &BaseNode {
        match self {
            PropertyKey::Identifier(wrapped) => &wrapped.base,
            PropertyKey::StringLiteral(wrapped) => &wrapped.base,
        }
    }
}

// This matches the grammar, and not ast.go:
//  ParenExpression                = "(" Expression ")" .
//  FunctionExpressionSuffix       = "=>" FunctionBodyExpression .
//  FunctionBodyExpression         = Block | Expression .
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(untagged)]
pub enum FunctionBody {
    Block(Block),
    Expr(Expression),
}

impl FunctionBody {
    // `base` is an utility method that returns the BaseNode for a FunctionBody.
    pub fn base(&self) -> &BaseNode {
        match self {
            FunctionBody::Block(wrapped) => &wrapped.base,
            FunctionBody::Expr(wrapped) => &wrapped.base(),
        }
    }
}

fn serialize_errors<S>(errors: &Vec<String>, ser: S) -> Result<S::Ok, S::Error>
where
    S: Serializer,
{
    let mut seq = ser.serialize_seq(Some(errors.len()))?;
    for e in errors {
        let mut me = HashMap::new();
        me.insert("msg".to_string(), e);
        seq.serialize_element(&me)?;
    }
    seq.end()
}

// BaseNode holds the attributes every expression or statement should have
#[derive(Debug, Default, PartialEq, Clone, Serialize, Deserialize)]
pub struct BaseNode {
    #[serde(default)]
    pub location: SourceLocation,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[serde(serialize_with = "serialize_errors")]
    #[serde(default)]
    pub errors: Vec<String>,
}

impl BaseNode {
    pub fn is_empty(&self) -> bool {
        self.errors.is_empty() && !self.location.is_valid()
    }
}

// Package represents a complete package source tree
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct Package {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    #[serde(skip_serializing_if = "String::is_empty")]
    #[serde(default)]
    pub path: String,
    pub package: String,
    pub files: Vec<File>,
}

// File represents a source from a single file
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct File {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    #[serde(skip_serializing_if = "String::is_empty")]
    #[serde(default)]
    pub name: String,
    pub package: Option<PackageClause>,
    #[serde(deserialize_with = "deserialize_default_from_null")]
    pub imports: Vec<ImportDeclaration>,
    pub body: Vec<Statement>,
}

// PackageClause defines the current package identifier.
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct PackageClause {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub name: Identifier,
}

// ImportDeclaration declares a single import
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct ImportDeclaration {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    #[serde(rename = "as")]
    pub alias: Option<Identifier>,
    pub path: StringLiteral,
}

// Block is a set of statements
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct Block {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub body: Vec<Statement>,
}

// BadStatement is a placeholder for statements for which no correct statement nodes
// can be created.
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct BadStatement {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub text: String,
}

// ExpressionStatement may consist of an expression that does not return a value and is executed solely for its side-effects.
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct ExpressionStatement {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub expression: Expression,
}

// ReturnStatement defines an Expression to return
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct ReturnStatement {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub argument: Expression,
}

// OptionStatement syntactically is a single variable declaration
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct OptionStatement {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub assignment: Assignment,
}

// BuiltinStatement declares a builtin identifier and its struct
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct BuiltinStatement {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub id: Identifier,
}

// TestStatement declares a Flux test case
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct TestStatement {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub assignment: VariableAssignment,
}

// VariableAssignment represents the declaration of a variable
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct VariableAssignment {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub id: Identifier,
    pub init: Expression,
}

#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct MemberAssignment {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub member: MemberExpression,
    pub init: Expression,
}

// StringExpression represents an interpolated string
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct StringExpression {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    pub base: BaseNode,
    pub parts: Vec<StringExpressionPart>,
}

// StringExpressionPart represents part of an interpolated string
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(untagged)]
pub enum StringExpressionPart {
    Text(TextPart),
    Expr(InterpolatedPart),
}

// TextPart represents the text part of an interpolated string
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct TextPart {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    pub base: BaseNode,
    pub value: String,
}

// InterpolatedPart represents the expression part of an interpolated string
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct InterpolatedPart {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    pub base: BaseNode,
    pub expression: Expression,
}

// ParenExpression represents an expression wrapped in parenthesis
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct ParenExpression {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    pub base: BaseNode,
    pub expression: Expression,
}

// CallExpression represents a function call
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct CallExpression {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub callee: Expression,
    pub arguments: Vec<Expression>,
}

#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct PipeExpression {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub argument: Expression,
    pub call: CallExpression,
}

// MemberExpression represents calling a property of a CallExpression
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct MemberExpression {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub object: Expression,
    pub property: PropertyKey,
}

// IndexExpression represents indexing into an array
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct IndexExpression {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub array: Expression,
    pub index: Expression,
}

#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct FunctionExpression {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub params: Vec<Property>,
    pub body: FunctionBody,
}

// OperatorKind are Equality and Arithmetic operators.
// Result of evaluating an equality operator is always of type Boolean based on whether the
// comparison is true.
// Arithmetic operators take numerical values (either literals or variables) as their operands
// and return a single numerical value.
#[derive(Debug, PartialEq, Clone)]
pub enum OperatorKind {
    MultiplicationOperator,
    DivisionOperator,
    ModuloOperator,
    PowerOperator,
    AdditionOperator,
    SubtractionOperator,
    LessThanEqualOperator,
    LessThanOperator,
    GreaterThanEqualOperator,
    GreaterThanOperator,
    StartsWithOperator,
    InOperator,
    NotOperator,
    ExistsOperator,
    NotEmptyOperator,
    EmptyOperator,
    EqualOperator,
    NotEqualOperator,
    RegexpMatchOperator,
    NotRegexpMatchOperator,

    // this is necessary for bad binary expressions.
    InvalidOperator,
}

impl ToString for OperatorKind {
    fn to_string(&self) -> String {
        match self {
            OperatorKind::MultiplicationOperator => "*".to_string(),
            OperatorKind::DivisionOperator => "/".to_string(),
            OperatorKind::ModuloOperator => "%".to_string(),
            OperatorKind::PowerOperator => "^".to_string(),
            OperatorKind::AdditionOperator => "+".to_string(),
            OperatorKind::SubtractionOperator => "-".to_string(),
            OperatorKind::LessThanEqualOperator => "<=".to_string(),
            OperatorKind::LessThanOperator => "<".to_string(),
            OperatorKind::GreaterThanEqualOperator => ">=".to_string(),
            OperatorKind::GreaterThanOperator => ">".to_string(),
            OperatorKind::StartsWithOperator => "startswith".to_string(),
            OperatorKind::InOperator => "in".to_string(),
            OperatorKind::NotOperator => "not".to_string(),
            OperatorKind::ExistsOperator => "exists".to_string(),
            OperatorKind::NotEmptyOperator => "not empty".to_string(),
            OperatorKind::EmptyOperator => "empty".to_string(),
            OperatorKind::EqualOperator => "==".to_string(),
            OperatorKind::NotEqualOperator => "!=".to_string(),
            OperatorKind::RegexpMatchOperator => "=~".to_string(),
            OperatorKind::NotRegexpMatchOperator => "!~".to_string(),
            OperatorKind::InvalidOperator => "<INVALID_OP>".to_string(),
        }
    }
}

impl Serialize for OperatorKind {
    fn serialize<S>(&self, serializer: S) -> Result<<S as Serializer>::Ok, <S as Serializer>::Error>
    where
        S: Serializer,
    {
        serialize_to_string(self, serializer)
    }
}

impl FromStr for OperatorKind {
    type Err = String;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s {
            "*" => Ok(OperatorKind::MultiplicationOperator),
            "/" => Ok(OperatorKind::DivisionOperator),
            "%" => Ok(OperatorKind::ModuloOperator),
            "^" => Ok(OperatorKind::PowerOperator),
            "+" => Ok(OperatorKind::AdditionOperator),
            "-" => Ok(OperatorKind::SubtractionOperator),
            "<=" => Ok(OperatorKind::LessThanEqualOperator),
            "<" => Ok(OperatorKind::LessThanOperator),
            ">=" => Ok(OperatorKind::GreaterThanEqualOperator),
            ">" => Ok(OperatorKind::GreaterThanOperator),
            "startswith" => Ok(OperatorKind::StartsWithOperator),
            "in" => Ok(OperatorKind::InOperator),
            "not" => Ok(OperatorKind::NotOperator),
            "exists" => Ok(OperatorKind::ExistsOperator),
            "not empty" => Ok(OperatorKind::NotEmptyOperator),
            "empty" => Ok(OperatorKind::EmptyOperator),
            "==" => Ok(OperatorKind::EqualOperator),
            "!=" => Ok(OperatorKind::NotEqualOperator),
            "=~" => Ok(OperatorKind::RegexpMatchOperator),
            "!~" => Ok(OperatorKind::NotRegexpMatchOperator),
            "<INVALID_OP>" => Ok(OperatorKind::InvalidOperator),
            _ => Err(format!("unknown operator: {}", s)),
        }
    }
}

struct OperatorKindVisitor;

impl<'de> Visitor<'de> for OperatorKindVisitor {
    type Value = OperatorKind;

    fn expecting(&self, formatter: &mut fmt::Formatter) -> fmt::Result {
        formatter.write_str("a valid string valid for an operator")
    }

    fn visit_str<E>(self, value: &str) -> Result<Self::Value, E>
    where
        E: Error,
    {
        let r = value.parse::<OperatorKind>();
        match r {
            Ok(v) => Ok(v),
            Err(s) => Err(E::custom(s)),
        }
    }
}

impl<'de> Deserialize<'de> for OperatorKind {
    fn deserialize<D>(d: D) -> Result<Self, D::Error>
    where
        D: Deserializer<'de>,
    {
        d.deserialize_str(OperatorKindVisitor)
    }
}

// BinaryExpression use binary operators act on two operands in an expression.
// BinaryExpression includes relational and arithmetic operators
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct BinaryExpression {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub operator: OperatorKind,
    pub left: Expression,
    pub right: Expression,
}

// UnaryExpression use operators act on a single operand in an expression.
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct UnaryExpression {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub operator: OperatorKind,
    pub argument: Expression,
}

// LogicalOperatorKind are used with boolean (logical) values
#[derive(Debug, PartialEq, Clone)]
pub enum LogicalOperatorKind {
    AndOperator,
    OrOperator,
}

impl ToString for LogicalOperatorKind {
    fn to_string(&self) -> String {
        match self {
            LogicalOperatorKind::AndOperator => "and".to_string(),
            LogicalOperatorKind::OrOperator => "or".to_string(),
        }
    }
}

impl Serialize for LogicalOperatorKind {
    fn serialize<S>(&self, serializer: S) -> Result<<S as Serializer>::Ok, <S as Serializer>::Error>
    where
        S: Serializer,
    {
        serialize_to_string(self, serializer)
    }
}

impl FromStr for LogicalOperatorKind {
    type Err = String;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s {
            "and" => Ok(LogicalOperatorKind::AndOperator),
            "or" => Ok(LogicalOperatorKind::OrOperator),
            _ => Err(format!("unknown logical operator: {}", s)),
        }
    }
}

struct LogicalOperatorKindVisitor;

impl<'de> Visitor<'de> for LogicalOperatorKindVisitor {
    type Value = LogicalOperatorKind;

    fn expecting(&self, formatter: &mut fmt::Formatter) -> fmt::Result {
        formatter.write_str("a valid string valid for a logical operator")
    }

    fn visit_str<E>(self, value: &str) -> Result<Self::Value, E>
    where
        E: Error,
    {
        let r = value.parse::<LogicalOperatorKind>();
        match r {
            Ok(v) => Ok(v),
            Err(s) => Err(E::custom(s)),
        }
    }
}

impl<'de> Deserialize<'de> for LogicalOperatorKind {
    fn deserialize<D>(d: D) -> Result<Self, D::Error>
    where
        D: Deserializer<'de>,
    {
        d.deserialize_str(LogicalOperatorKindVisitor)
    }
}

// LogicalExpression represent the rule conditions that collectively evaluate to either true or false.
// `or` expressions compute the disjunction of two boolean expressions and return boolean values.
// `and`` expressions compute the conjunction of two boolean expressions and return boolean values.
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct LogicalExpression {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub operator: LogicalOperatorKind,
    pub left: Expression,
    pub right: Expression,
}

// ArrayExpression is used to create and directly specify the elements of an array object
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct ArrayExpression {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub elements: Vec<Expression>,
}

// ObjectExpression allows the declaration of an anonymous object within a declaration.
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct ObjectExpression {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub with: Option<Identifier>,
    pub properties: Vec<Property>,
}

// ConditionalExpression selects one of two expressions, `Alternate` or `Consequent`
// depending on a third, boolean, expression, `Test`.
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct ConditionalExpression {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub test: Expression,
    pub consequent: Expression,
    pub alternate: Expression,
}

// BadExpression is a malformed expression that contains the reason why in `text`.
// It can contain another expression, so that the parser can make a chained list of bad expressions.
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct BadExpression {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub text: String,
    pub expression: Option<Expression>,
}

// Property is the value associated with a key.
// A property's key can be either an identifier or string literal.
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct Property {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub key: PropertyKey,
    // `value` is optional, because of the shortcut: {a} <--> {a: a}
    pub value: Option<Expression>,
}

// Identifier represents a name that identifies a unique Node
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct Identifier {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub name: String,
}

// PipeLiteral represents an specialized literal value, indicating the left hand value of a pipe expression.
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct PipeLiteral {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
}

// StringLiteral expressions begin and end with double quote marks.
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct StringLiteral {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub value: String,
}

// BooleanLiteral represent boolean values
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct BooleanLiteral {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub value: bool,
}

// FloatLiteral  represent floating point numbers according to the double representations defined by the IEEE-754-1985
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct FloatLiteral {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub value: f64,
}

// IntegerLiteral represent integer numbers.
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct IntegerLiteral {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    #[serde(serialize_with = "serialize_to_string")]
    #[serde(deserialize_with = "deserialize_str_i64")]
    pub value: i64,
}

// UnsignedIntegerLiteral represent integer numbers.
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct UnsignedIntegerLiteral {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    #[serde(serialize_with = "serialize_to_string")]
    #[serde(deserialize_with = "deserialize_str_u64")]
    pub value: u64,
}

struct U64Visitor;

impl<'de> Visitor<'de> for U64Visitor {
    type Value = u64;

    fn expecting(&self, formatter: &mut fmt::Formatter) -> fmt::Result {
        formatter.write_str("a string representation for an unsigned integer")
    }

    fn visit_str<E>(self, value: &str) -> Result<Self::Value, E>
    where
        E: Error,
    {
        let r = value.parse::<u64>();
        match r {
            Ok(v) => Ok(v),
            Err(s) => Err(E::custom(s)),
        }
    }
}

fn deserialize_str_u64<'de, D>(d: D) -> Result<u64, D::Error>
where
    D: Deserializer<'de>,
{
    d.deserialize_str(U64Visitor)
}

struct I64Visitor;

impl<'de> Visitor<'de> for I64Visitor {
    type Value = i64;

    fn expecting(&self, formatter: &mut fmt::Formatter) -> fmt::Result {
        formatter.write_str("a string representation for an integer")
    }

    fn visit_str<E>(self, value: &str) -> Result<Self::Value, E>
    where
        E: Error,
    {
        let r = value.parse::<i64>();
        match r {
            Ok(v) => Ok(v),
            Err(s) => Err(E::custom(s)),
        }
    }
    fn visit_string<E>(self, value: String) -> Result<Self::Value, E>
    where
        E: Error,
    {
        let r = value.parse::<i64>();
        match r {
            Ok(v) => Ok(v),
            Err(s) => Err(E::custom(s)),
        }
    }
}

fn deserialize_str_i64<'de, D>(d: D) -> Result<i64, D::Error>
where
    D: Deserializer<'de>,
{
    d.deserialize_str(I64Visitor)
}

// RegexpLiteral expressions begin and end with `/` and are regular expressions with syntax accepted by RE2
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct RegexpLiteral {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub value: String,
}

// Duration is a pair consisting of length of time and the unit of time measured.
// It is the atomic unit from which all duration literals are composed.
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct Duration {
    pub magnitude: i64,
    pub unit: String,
}

// DurationLiteral represents the elapsed time between two instants as an
// int64 nanosecond count with syntax of golang's time.Duration
// TODO: this may be better as a class initialization
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct DurationLiteral {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub values: Vec<Duration>,
}

// TODO: we need a "duration from" that takes a time and a durationliteral, and gives an exact time.Duration instead of an approximation
//
// DateTimeLiteral represents an instant in time with nanosecond precision using
// the syntax of golang's RFC3339 Nanosecond variant
// TODO: this may be better as a class initialization
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct DateTimeLiteral {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub value: DateTime<FixedOffset>,
}

#[cfg(test)]
mod tests;
