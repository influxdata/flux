#![allow(missing_docs)]
pub mod check;

pub mod flatbuffers;
pub mod walk;

use crate::scanner;
use std::collections::HashMap;
use std::fmt;
use std::str::FromStr;
use std::vec::Vec;

use chrono::FixedOffset;

use serde::de::{Deserialize, Deserializer, Error, Visitor};
use serde::ser::{Serialize, SerializeSeq, Serializer};
use serde_aux::prelude::*;

pub const DEFAULT_PACKAGE_NAME: &str = "main";

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
        Self::invalid()
    }
}

// SourceLocation represents the location of a node in the AST
#[derive(Debug, Default, PartialEq, Clone, Serialize, Deserialize)]
pub struct SourceLocation {
    #[serde(skip_serializing_if = "skip_string_option")]
    pub file: Option<String>, // File is the optional file name.
    pub start: Position, // Start is the location in the source the node starts.
    pub end: Position,   // End is the location in the source the node ends.
    #[serde(skip_serializing_if = "skip_string_option")]
    pub source: Option<String>, // Source is optional raw source.
}

impl SourceLocation {
    pub fn is_valid(&self) -> bool {
        self.start.is_valid() && self.end.is_valid()
    }
}

fn skip_string_option(opt_str: &Option<String>) -> bool {
    opt_str.is_none() || opt_str.as_ref().unwrap().is_empty()
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
#[serde(tag = "type")]
pub enum Expression {
    Identifier(Identifier),
    #[serde(rename = "ArrayExpression")]
    Array(Box<ArrayExpr>),
    #[serde(rename = "FunctionExpression")]
    Function(Box<FunctionExpr>),
    #[serde(rename = "LogicalExpression")]
    Logical(Box<LogicalExpr>),
    #[serde(rename = "ObjectExpression")]
    Object(Box<ObjectExpr>),
    #[serde(rename = "MemberExpression")]
    Member(Box<MemberExpr>),
    #[serde(rename = "IndexExpression")]
    Index(Box<IndexExpr>),
    #[serde(rename = "BinaryExpression")]
    Binary(Box<BinaryExpr>),
    #[serde(rename = "UnaryExpression")]
    Unary(Box<UnaryExpr>),
    #[serde(rename = "PipeExpression")]
    PipeExpr(Box<PipeExpr>),
    #[serde(rename = "CallExpression")]
    Call(Box<CallExpr>),
    #[serde(rename = "ConditionalExpression")]
    Conditional(Box<ConditionalExpr>),
    #[serde(rename = "StringExpression")]
    StringExpr(Box<StringExpr>),
    #[serde(rename = "ParenExpression")]
    Paren(Box<ParenExpr>),

    #[serde(rename = "IntegerLiteral")]
    Integer(IntegerLit),
    #[serde(rename = "FloatLiteral")]
    Float(FloatLit),
    #[serde(rename = "StringLiteral")]
    StringLit(StringLit),
    #[serde(rename = "DurationLiteral")]
    Duration(DurationLit),
    #[serde(rename = "UnsignedIntegerLiteral")]
    Uint(UintLit),
    #[serde(rename = "BooleanLiteral")]
    Boolean(BooleanLit),
    #[serde(rename = "DateTimeLiteral")]
    DateTime(DateTimeLit),
    #[serde(rename = "RegexpLiteral")]
    Regexp(RegexpLit),
    #[serde(rename = "PipeLiteral")]
    PipeLit(PipeLit),

    #[serde(rename = "BadExpression")]
    Bad(Box<BadExpr>),
}

impl Expression {
    // `base` is an utility method that returns the BaseNode for an Expression.
    pub fn base(&self) -> &BaseNode {
        match self {
            Expression::Identifier(wrapped) => &wrapped.base,
            Expression::Array(wrapped) => &wrapped.base,
            Expression::Function(wrapped) => &wrapped.base,
            Expression::Logical(wrapped) => &wrapped.base,
            Expression::Object(wrapped) => &wrapped.base,
            Expression::Member(wrapped) => &wrapped.base,
            Expression::Index(wrapped) => &wrapped.base,
            Expression::Binary(wrapped) => &wrapped.base,
            Expression::Unary(wrapped) => &wrapped.base,
            Expression::PipeExpr(wrapped) => &wrapped.base,
            Expression::Call(wrapped) => &wrapped.base,
            Expression::Conditional(wrapped) => &wrapped.base,
            Expression::Integer(wrapped) => &wrapped.base,
            Expression::Float(wrapped) => &wrapped.base,
            Expression::StringLit(wrapped) => &wrapped.base,
            Expression::Duration(wrapped) => &wrapped.base,
            Expression::Uint(wrapped) => &wrapped.base,
            Expression::Boolean(wrapped) => &wrapped.base,
            Expression::DateTime(wrapped) => &wrapped.base,
            Expression::Regexp(wrapped) => &wrapped.base,
            Expression::PipeLit(wrapped) => &wrapped.base,
            Expression::Bad(wrapped) => &wrapped.base,
            Expression::StringExpr(wrapped) => &wrapped.base,
            Expression::Paren(wrapped) => &wrapped.base,
        }
    }
}

#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub enum Statement {
    #[serde(rename = "ExpressionStatement")]
    Expr(Box<ExprStmt>),
    #[serde(rename = "VariableAssignment")]
    Variable(Box<VariableAssgn>),
    #[serde(rename = "OptionStatement")]
    Option(Box<OptionStmt>),
    #[serde(rename = "ReturnStatement")]
    Return(Box<ReturnStmt>),
    #[serde(rename = "BadStatement")]
    Bad(Box<BadStmt>),
    #[serde(rename = "TestStatement")]
    Test(Box<TestStmt>),
    #[serde(rename = "BuiltinStatement")]
    Builtin(Box<BuiltinStmt>),
}

impl Statement {
    // `base` is an utility method that returns the BaseNode for a Statement.
    pub fn base(&self) -> &BaseNode {
        match self {
            Statement::Expr(wrapped) => &wrapped.base,
            Statement::Variable(wrapped) => &wrapped.base,
            Statement::Option(wrapped) => &wrapped.base,
            Statement::Return(wrapped) => &wrapped.base,
            Statement::Bad(wrapped) => &wrapped.base,
            Statement::Test(wrapped) => &wrapped.base,
            Statement::Builtin(wrapped) => &wrapped.base,
        }
    }

    // returns a integer based type value.
    pub fn typ(&self) -> i8 {
        match self {
            Statement::Expr(_) => 0,
            Statement::Variable(_) => 1,
            Statement::Option(_) => 2,
            Statement::Return(_) => 3,
            Statement::Bad(_) => 4,
            Statement::Test(_) => 5,
            Statement::Builtin(_) => 6,
        }
    }
}

#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub enum Assignment {
    #[serde(rename = "VariableAssignment")]
    Variable(Box<VariableAssgn>),
    #[serde(rename = "MemberAssignment")]
    Member(Box<MemberAssgn>),
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
#[serde(tag = "type")]
pub enum PropertyKey {
    Identifier(Identifier),
    #[serde(rename = "StringLiteral")]
    StringLit(StringLit),
}

impl PropertyKey {
    // `base` is an utility method that returns the BaseNode for a PropertyKey.
    pub fn base(&self) -> &BaseNode {
        match self {
            PropertyKey::Identifier(wrapped) => &wrapped.base,
            PropertyKey::StringLit(wrapped) => &wrapped.base,
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

fn serialize_errors<S>(errors: &[String], ser: S) -> Result<S::Ok, S::Error>
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

#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct Comment {
    pub lit: String,
    pub next: Option<Box<Comment>>,
}

pub type CommentList = Option<Box<Comment>>;

// BaseNode holds the attributes every expression or statement must have
#[derive(Debug, Default, PartialEq, Clone, Serialize, Deserialize)]
pub struct BaseNode {
    #[serde(default)]
    pub location: SourceLocation,
    // If the base node is for a terminal the comments will be here. We also
    // use the base node comments when a non-terminal contains just one
    // terminal on the right hand side. This saves us populating the
    // type-specific AST nodes with comment lists when we can avoid it..
    #[serde(skip_serializing_if = "Option::is_none")]
    pub comments: CommentList,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[serde(serialize_with = "serialize_errors")]
    #[serde(default)]
    pub errors: Vec<String>,
}

impl BaseNode {
    pub fn is_empty(&self) -> bool {
        self.errors.is_empty() && !self.location.is_valid()
    }

    pub fn set_comments(&mut self, comments: CommentList) {
        self.comments = comments;
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

impl From<File> for Package {
    fn from(file: File) -> Self {
        Package {
            base: BaseNode {
                ..BaseNode::default()
            },
            path: String::from(""),
            package: String::from(file.get_package()),
            files: vec![file],
        }
    }
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
    #[serde(skip_serializing_if = "String::is_empty")]
    #[serde(default)]
    pub metadata: String,
    pub package: Option<PackageClause>,
    #[serde(deserialize_with = "deserialize_default_from_null")]
    pub imports: Vec<ImportDeclaration>,
    pub body: Vec<Statement>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub eof: CommentList,
}

impl File {
    fn get_package(self: &File) -> &str {
        match &self.package {
            Some(pkg_clause) => pkg_clause.name.name.as_str(),
            None => DEFAULT_PACKAGE_NAME,
        }
    }
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
    pub path: StringLit,
}

// Block is a set of statements
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub struct Block {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub lbrace: CommentList,
    pub body: Vec<Statement>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub rbrace: CommentList,
}

// BadStmt is a placeholder for statements for which no correct statement nodes
// can be created.
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(rename = "BadStatement")]
pub struct BadStmt {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub text: String,
}

// ExprStmt may consist of an expression that does not return a value and is executed solely for its side-effects.
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct ExprStmt {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub expression: Expression,
}

// ReturnStmt defines an Expression to return
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct ReturnStmt {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub argument: Expression,
}

// OptionStmt syntactically is a single variable declaration
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct OptionStmt {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub assignment: Assignment,
}

// BuiltinStmt declares a builtin identifier and its struct
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct BuiltinStmt {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub id: Identifier,
}

#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub enum MonoType {
    #[serde(rename = "TvarType")]
    Tvar(TvarType),
    #[serde(rename = "NamedType")]
    Basic(NamedType),
    #[serde(rename = "ArrayType")]
    Array(Box<ArrayType>),
    #[serde(rename = "RecordType")]
    Record(RecordType),
    #[serde(rename = "FunctionType")]
    Function(Box<FunctionType>),
    Invalid,
}

pub fn base_from_monotype(m: &MonoType) -> BaseNode {
    match m {
        MonoType::Basic(t) => t.base.clone(),
        MonoType::Tvar(t) => t.base.clone(),
        MonoType::Array(t) => t.base.clone(),
        MonoType::Record(t) => t.base.clone(),
        MonoType::Function(t) => t.base.clone(),
        _ => BaseNode::default(),
    }
}
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct NamedType {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub name: Identifier,
}
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct TvarType {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub name: Identifier,
}
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct ArrayType {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub monotype: MonoType,
}

#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct FunctionType {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub parameters: Vec<ParameterType>,
    pub monotype: MonoType,
}

#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub enum ParameterType {
    #[serde(rename = "Required")]
    Required {
        #[serde(skip_serializing_if = "BaseNode::is_empty")]
        #[serde(default)]
        #[serde(flatten)]
        base: BaseNode,
        name: Identifier,
        ty: MonoType,
    },
    #[serde(rename = "Optional")]
    Optional {
        #[serde(skip_serializing_if = "BaseNode::is_empty")]
        #[serde(default)]
        #[serde(flatten)]
        base: BaseNode,
        name: Identifier,
        ty: MonoType,
    },
    #[serde(rename = "Pipe")]
    Pipe {
        #[serde(skip_serializing_if = "BaseNode::is_empty")]
        #[serde(default)]
        #[serde(flatten)]
        base: BaseNode,
        #[serde(skip_serializing_if = "Option::is_none")]
        name: Option<Identifier>,
        ty: MonoType,
    },
    Invalid,
}

#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct TypeExpression {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub monotype: MonoType,
    pub constraint: Option<TypeConstraint>,
}
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct TypeConstraint {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub tvar: Identifier,
    pub kinds: Vec<Identifier>,
}
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct RecordType {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub tvar: Option<Identifier>,
    pub properties: Option<Vec<PropertyType>>,
}
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct PropertyType {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub identifier: Identifier,
    pub monotype: MonoType,
}
// TestStmt declares a Flux test case
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct TestStmt {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub assignment: VariableAssgn,
}

// VariableAssgn represents the declaration of a variable
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct VariableAssgn {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub id: Identifier,
    pub init: Expression,
}

// MemberAssgn represents an assignement into a member of an object.
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct MemberAssgn {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub member: MemberExpr,
    pub init: Expression,
}

// StringExpr represents an interpolated string
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct StringExpr {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub parts: Vec<StringExprPart>,
}

// StringExprPart represents part of an interpolated string
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub enum StringExprPart {
    #[serde(rename = "TextPart")]
    Text(TextPart),
    #[serde(rename = "InterpolatedPart")]
    Interpolated(InterpolatedPart),
}

// TextPart represents the text part of an interpolated string
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct TextPart {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub value: String,
}

// InterpolatedPart represents the expression part of an interpolated string
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct InterpolatedPart {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub expression: Expression,
}

// ParenExpr represents an expression wrapped in parenthesis
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct ParenExpr {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub lparen: CommentList,
    pub expression: Expression,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub rparen: CommentList,
}

// CallExpr represents a function call
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct CallExpr {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub callee: Expression,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub lparen: CommentList,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[serde(default)]
    pub arguments: Vec<Expression>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub rparen: CommentList,
}

// PipeExpr represents a call expression using the pipe forward syntax.
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct PipeExpr {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub argument: Expression,
    pub call: CallExpr,
}

// MemberExpr represents calling a property of a Call
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct MemberExpr {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub object: Expression,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub lbrack: CommentList,
    pub property: PropertyKey,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub rbrack: CommentList,
}

// IndexExpr represents indexing into an array
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct IndexExpr {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub array: Expression,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub lbrack: CommentList,
    pub index: Expression,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub rbrack: CommentList,
}

#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct FunctionExpr {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub lparen: CommentList,
    #[serde(deserialize_with = "deserialize_default_from_null")]
    pub params: Vec<Property>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub rparen: CommentList,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub arrow: CommentList,
    pub body: FunctionBody,
}

// Operator are Equality and Arithmetic operators.
// Result of evaluating an equality operator is always of type Boolean based on whether the
// comparison is true.
// Arithmetic operators take numerical values (either literals or variables) as their operands
// and return a single numerical value.
#[derive(Debug, PartialEq, Clone)]
pub enum Operator {
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

impl ToString for Operator {
    fn to_string(&self) -> String {
        match self {
            Operator::MultiplicationOperator => "*".to_string(),
            Operator::DivisionOperator => "/".to_string(),
            Operator::ModuloOperator => "%".to_string(),
            Operator::PowerOperator => "^".to_string(),
            Operator::AdditionOperator => "+".to_string(),
            Operator::SubtractionOperator => "-".to_string(),
            Operator::LessThanEqualOperator => "<=".to_string(),
            Operator::LessThanOperator => "<".to_string(),
            Operator::GreaterThanEqualOperator => ">=".to_string(),
            Operator::GreaterThanOperator => ">".to_string(),
            Operator::StartsWithOperator => "startswith".to_string(),
            Operator::InOperator => "in".to_string(),
            Operator::NotOperator => "not".to_string(),
            Operator::ExistsOperator => "exists".to_string(),
            Operator::NotEmptyOperator => "not empty".to_string(),
            Operator::EmptyOperator => "empty".to_string(),
            Operator::EqualOperator => "==".to_string(),
            Operator::NotEqualOperator => "!=".to_string(),
            Operator::RegexpMatchOperator => "=~".to_string(),
            Operator::NotRegexpMatchOperator => "!~".to_string(),
            Operator::InvalidOperator => "<INVALID_OP>".to_string(),
        }
    }
}

impl Serialize for Operator {
    fn serialize<S>(&self, serializer: S) -> Result<<S as Serializer>::Ok, <S as Serializer>::Error>
    where
        S: Serializer,
    {
        serialize_to_string(self, serializer)
    }
}

impl FromStr for Operator {
    type Err = String;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s {
            "*" => Ok(Operator::MultiplicationOperator),
            "/" => Ok(Operator::DivisionOperator),
            "%" => Ok(Operator::ModuloOperator),
            "^" => Ok(Operator::PowerOperator),
            "+" => Ok(Operator::AdditionOperator),
            "-" => Ok(Operator::SubtractionOperator),
            "<=" => Ok(Operator::LessThanEqualOperator),
            "<" => Ok(Operator::LessThanOperator),
            ">=" => Ok(Operator::GreaterThanEqualOperator),
            ">" => Ok(Operator::GreaterThanOperator),
            "startswith" => Ok(Operator::StartsWithOperator),
            "in" => Ok(Operator::InOperator),
            "not" => Ok(Operator::NotOperator),
            "exists" => Ok(Operator::ExistsOperator),
            "not empty" => Ok(Operator::NotEmptyOperator),
            "empty" => Ok(Operator::EmptyOperator),
            "==" => Ok(Operator::EqualOperator),
            "!=" => Ok(Operator::NotEqualOperator),
            "=~" => Ok(Operator::RegexpMatchOperator),
            "!~" => Ok(Operator::NotRegexpMatchOperator),
            "<INVALID_OP>" => Ok(Operator::InvalidOperator),
            _ => Err(format!("unknown operator: {}", s)),
        }
    }
}

struct OperatorVisitor;

impl<'de> Visitor<'de> for OperatorVisitor {
    type Value = Operator;

    fn expecting(&self, formatter: &mut fmt::Formatter) -> fmt::Result {
        formatter.write_str("a valid string valid for an operator")
    }

    fn visit_str<E>(self, value: &str) -> Result<Self::Value, E>
    where
        E: Error,
    {
        let r = value.parse::<Operator>();
        match r {
            Ok(v) => Ok(v),
            Err(s) => Err(E::custom(s)),
        }
    }
}

impl<'de> Deserialize<'de> for Operator {
    fn deserialize<D>(d: D) -> Result<Self, D::Error>
    where
        D: Deserializer<'de>,
    {
        d.deserialize_str(OperatorVisitor)
    }
}

// BinaryExpr use binary operators act on two operands in an expression.
// BinaryExpr includes relational and arithmetic operators
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(rename = "BinaryExpression")]
pub struct BinaryExpr {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub operator: Operator,
    pub left: Expression,
    pub right: Expression,
}

// UnaryExpr use operators act on a single operand in an expression.
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(rename = "UnaryExpression")]
pub struct UnaryExpr {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub operator: Operator,
    pub argument: Expression,
}

// LogicalOperator are used with boolean (logical) values
#[derive(Debug, PartialEq, Clone)]
pub enum LogicalOperator {
    AndOperator,
    OrOperator,
}

impl ToString for LogicalOperator {
    fn to_string(&self) -> String {
        match self {
            LogicalOperator::AndOperator => "and".to_string(),
            LogicalOperator::OrOperator => "or".to_string(),
        }
    }
}

impl Serialize for LogicalOperator {
    fn serialize<S>(&self, serializer: S) -> Result<<S as Serializer>::Ok, <S as Serializer>::Error>
    where
        S: Serializer,
    {
        serialize_to_string(self, serializer)
    }
}

impl FromStr for LogicalOperator {
    type Err = String;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s {
            "and" => Ok(LogicalOperator::AndOperator),
            "or" => Ok(LogicalOperator::OrOperator),
            _ => Err(format!("unknown logical operator: {}", s)),
        }
    }
}

struct LogicalOperatorVisitor;

impl<'de> Visitor<'de> for LogicalOperatorVisitor {
    type Value = LogicalOperator;

    fn expecting(&self, formatter: &mut fmt::Formatter) -> fmt::Result {
        formatter.write_str("a valid string valid for a logical operator")
    }

    fn visit_str<E>(self, value: &str) -> Result<Self::Value, E>
    where
        E: Error,
    {
        let r = value.parse::<LogicalOperator>();
        match r {
            Ok(v) => Ok(v),
            Err(s) => Err(E::custom(s)),
        }
    }
}

impl<'de> Deserialize<'de> for LogicalOperator {
    fn deserialize<D>(d: D) -> Result<Self, D::Error>
    where
        D: Deserializer<'de>,
    {
        d.deserialize_str(LogicalOperatorVisitor)
    }
}

// LogicalExpr represent the rule conditions that collectively evaluate to either true or false.
// `or` expressions compute the disjunction of two boolean expressions and return boolean values.
// `and`` expressions compute the conjunction of two boolean expressions and return boolean values.
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct LogicalExpr {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub operator: LogicalOperator,
    pub left: Expression,
    pub right: Expression,
}

// ArrayExpr is used to create and directly specify the elements of an array object
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct ArrayItem {
    #[serde(default)]
    #[serde(flatten)]
    pub expression: Expression,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub comma: CommentList,
}

// ArrayExpr is used to create and directly specify the elements of an array object
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct ArrayExpr {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub lbrack: CommentList,
    #[serde(deserialize_with = "deserialize_default_from_null")]
    pub elements: Vec<ArrayItem>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub rbrack: CommentList,
}

#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct WithSource {
    #[serde(default)]
    #[serde(flatten)]
    pub source: Identifier,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub with: CommentList,
}

// ObjectExpr allows the declaration of an anonymous object within a declaration.
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct ObjectExpr {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub lbrace: CommentList,
    #[serde(skip_serializing_if = "Option::is_none")]
    #[serde(default)]
    pub with: Option<WithSource>,
    #[serde(deserialize_with = "deserialize_default_from_null")]
    pub properties: Vec<Property>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub rbrace: CommentList,
}

// ConditionalExpr selects one of two expressions, `Alternate` or `Consequent`
// depending on a third, boolean, expression, `Test`.
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct ConditionalExpr {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub tk_if: CommentList,
    pub test: Expression,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub tk_then: CommentList,
    pub consequent: Expression,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub tk_else: CommentList,
    pub alternate: Expression,
}

// BadExpr is a malformed expression that contains the reason why in `text`.
// It can contain another expression, so that the parser can make a chained list of bad expressions.
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct BadExpr {
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
    #[serde(skip_serializing_if = "Option::is_none")]
    pub separator: CommentList,
    // `value` is optional, because of the shortcut: {a} <--> {a: a}
    pub value: Option<Expression>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub comma: CommentList,
}

// Identifier represents a name that identifies a unique Node
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct Identifier {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub name: String,
}

// PipeLit represents an specialized literal value, indicating the left hand value of a pipe expression.
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct PipeLit {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
}

// StringLit expressions begin and end with double quote marks.
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct StringLit {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub value: String,
}

// Boolean represent boolean values
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct BooleanLit {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub value: bool,
}

// FloatLit represent floating point numbers according to the double representations defined by the IEEE-754-1985
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct FloatLit {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub value: f64,
}

// IntegerLit represent integer numbers.
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct IntegerLit {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    #[serde(serialize_with = "serialize_to_string")]
    #[serde(deserialize_with = "deserialize_str_i64")]
    pub value: i64,
}

// UintLit represent integer numbers.
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct UintLit {
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

// RegexpLit expressions begin and end with `/` and are regular expressions with syntax accepted by RE2
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct RegexpLit {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub value: String,
}

// DurationLit is a pair consisting of length of time and the unit of time measured.
// It is the atomic unit from which all duration literals are composed.
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
#[serde(rename = "Duration")]
pub struct Duration {
    pub magnitude: i64,
    pub unit: String,
}

// DurationLit represents the elapsed time between two instants as an
// int64 nanosecond count with syntax of golang's time.Duration
// TODO: this may be better as a class initialization
// All magnitudes in Duration vector should have the same sign
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct DurationLit {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub values: Vec<Duration>,
}

// TODO: we need a "duration from" that takes a time and a durationliteral, and gives an exact time.DurationLit instead of an approximation
//
// DateTimeLit represents an instant in time with nanosecond precision using
// the syntax of golang's RFC3339 Nanosecond variant
// TODO: this may be better as a class initialization
#[derive(Debug, PartialEq, Clone, Serialize, Deserialize)]
pub struct DateTimeLit {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub value: chrono::DateTime<FixedOffset>,
}

#[cfg(test)]
mod tests;
