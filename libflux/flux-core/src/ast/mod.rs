//! Abstract syntax tree (AST).

pub mod check;
pub mod walk;

use std::{collections::HashMap, fmt, str::FromStr, vec::Vec};

use chrono::FixedOffset;
use ordered_float::NotNan;
use serde::{
    de::{Deserialize, Deserializer, Error, Visitor},
    ser::{Serialize, SerializeSeq, Serializer},
};

use super::DefaultHasher;
use crate::scanner;

/// The default package name.
pub const DEFAULT_PACKAGE_NAME: &str = "main";

/// Position is the AST counterpart of [`scanner::Position`].
/// It adds serde capabilities.
#[derive(Debug, PartialEq, Eq, Copy, Clone, Serialize, Deserialize, PartialOrd, Ord)]
#[allow(missing_docs)]
pub struct Position {
    pub line: u32,
    pub column: u32,
}

impl Position {
    #[allow(missing_docs)]
    pub fn is_valid(&self) -> bool {
        self.line > 0 && self.column > 0
    }
    #[allow(missing_docs)]
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

/// Convert a Position to a lsp_types::Position
/// https://microsoft.github.io/language-server-protocol/specification#position
#[cfg(feature = "lsp")]
impl From<Position> for lsp_types::Position {
    fn from(position: Position) -> Self {
        Self {
            line: position.line - 1,
            character: position.column - 1,
        }
    }
}

/// Represents the location of a node in the AST.
#[derive(Default, PartialEq, Eq, Clone, Serialize, Deserialize)]
pub struct SourceLocation {
    /// File is the optional file name.
    #[serde(skip_serializing_if = "skip_string_option")]
    pub file: Option<String>,
    /// Start is the location in the source the node starts.
    pub start: Position,
    /// End is the location in the source the node ends.
    pub end: Position,
    /// Source is optional raw source.
    #[serde(skip_serializing_if = "skip_string_option")]
    pub source: Option<String>,
}

// Custom debug implentation which reduces the size of `Debug` printing `AST`s
impl fmt::Debug for SourceLocation {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        let mut f = f.debug_struct("SourceLocation");

        if let Some(file) = &self.file {
            f.field("file", file);
        }

        // Render the positions on a single line so that `Debug` printing `AST`s are less verbose
        f.field(
            "start",
            &format!("line: {}, column: {}", self.start.line, self.start.column),
        );
        f.field(
            "end",
            &format!("line: {}, column: {}", self.end.line, self.end.column),
        );

        if let Some(source) = &self.source {
            f.field("source", source);
        }

        f.finish()
    }
}

impl SourceLocation {
    #[allow(missing_docs)]
    pub fn is_valid(&self) -> bool {
        self.start.is_valid() && self.end.is_valid()
    }
    #[allow(missing_docs)]
    pub fn is_multiline(&self) -> bool {
        self.start.line != self.end.line
    }
}

#[cfg(feature = "lsp")]
impl From<SourceLocation> for lsp_types::Range {
    fn from(range: SourceLocation) -> Self {
        Self {
            start: range.start.into(),
            end: range.end.into(),
        }
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

#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
#[allow(missing_docs)]
pub enum Expression {
    Identifier(Identifier),
    #[serde(rename = "ArrayExpression")]
    Array(Box<ArrayExpr>),
    #[serde(rename = "DictExpression")]
    Dict(Box<DictExpr>),
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
    #[serde(rename = "LabelLiteral")]
    Label(LabelLit),

    #[serde(rename = "BadExpression")]
    Bad(Box<BadExpr>),
}

impl Expression {
    /// Returns the [`BaseNode`] for an [`Expression`].
    pub fn base(&self) -> &BaseNode {
        match self {
            Expression::Identifier(wrapped) => &wrapped.base,
            Expression::Array(wrapped) => &wrapped.base,
            Expression::Dict(wrapped) => &wrapped.base,
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
            Expression::Label(wrapped) => &wrapped.base,
            Expression::Bad(wrapped) => &wrapped.base,
            Expression::StringExpr(wrapped) => &wrapped.base,
            Expression::Paren(wrapped) => &wrapped.base,
        }
    }
}

#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
#[allow(missing_docs)]
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
    #[serde(rename = "TestCaseStatement")]
    TestCase(Box<TestCaseStmt>),
    #[serde(rename = "BuiltinStatement")]
    Builtin(Box<BuiltinStmt>),
}

impl Statement {
    /// Returns the [`BaseNode`] for a [`Statement`].
    pub fn base(&self) -> &BaseNode {
        match self {
            Statement::Expr(wrapped) => &wrapped.base,
            Statement::Variable(wrapped) => &wrapped.base,
            Statement::Option(wrapped) => &wrapped.base,
            Statement::Return(wrapped) => &wrapped.base,
            Statement::Bad(wrapped) => &wrapped.base,
            Statement::TestCase(wrapped) => &wrapped.base,
            Statement::Builtin(wrapped) => &wrapped.base,
        }
    }

    /// Returns the [`BaseNode`] for a [`Statement`].
    pub fn base_mut(&mut self) -> &mut BaseNode {
        match self {
            Statement::Expr(wrapped) => &mut wrapped.base,
            Statement::Variable(wrapped) => &mut wrapped.base,
            Statement::Option(wrapped) => &mut wrapped.base,
            Statement::Return(wrapped) => &mut wrapped.base,
            Statement::Bad(wrapped) => &mut wrapped.base,
            Statement::TestCase(wrapped) => &mut wrapped.base,
            Statement::Builtin(wrapped) => &mut wrapped.base,
        }
    }

    /// Returns an integer-based type value.
    pub fn typ(&self) -> i8 {
        match self {
            Statement::Expr(_) => 0,
            Statement::Variable(_) => 1,
            Statement::Option(_) => 2,
            Statement::Return(_) => 3,
            Statement::Bad(_) => 4,
            Statement::TestCase(_) => 7,
            Statement::Builtin(_) => 6,
        }
    }
    /// Returns the name of the type of statement.
    pub fn type_name(&self) -> &'static str {
        match self {
            Statement::Expr(_) => "expression",
            Statement::Variable(_) => "variable",
            Statement::Option(_) => "option",
            Statement::Return(_) => "return",
            Statement::Bad(_) => "bad",
            Statement::TestCase(_) => "testcase",
            Statement::Builtin(_) => "builtin",
        }
    }
}

#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
#[allow(missing_docs)]
pub enum Assignment {
    #[serde(rename = "VariableAssignment")]
    Variable(Box<VariableAssgn>),
    #[serde(rename = "MemberAssignment")]
    Member(Box<MemberAssgn>),
}

impl Assignment {
    /// Returns the [`BaseNode`] for an [`Assignment`].
    pub fn base(&self) -> &BaseNode {
        match self {
            Assignment::Variable(wrapped) => &wrapped.base,
            Assignment::Member(wrapped) => &wrapped.base,
        }
    }
}

#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
#[allow(missing_docs)]
pub enum PropertyKey {
    Identifier(Identifier),
    #[serde(rename = "StringLiteral")]
    StringLit(StringLit),
}

impl From<Identifier> for PropertyKey {
    fn from(id: Identifier) -> Self {
        Self::Identifier(id)
    }
}

impl From<StringLit> for PropertyKey {
    fn from(lit: StringLit) -> Self {
        Self::StringLit(lit)
    }
}

impl PropertyKey {
    /// Returns the [`BaseNode`] for a [`PropertyKey`].
    pub fn base(&self) -> &BaseNode {
        match self {
            PropertyKey::Identifier(wrapped) => &wrapped.base,
            PropertyKey::StringLit(wrapped) => &wrapped.base,
        }
    }

    /// Returns the key
    pub fn key(&self) -> &str {
        match self {
            PropertyKey::Identifier(wrapped) => &wrapped.name,
            PropertyKey::StringLit(wrapped) => &wrapped.value,
        }
    }
}

// This matches the grammar, and not ast.go:
//  ParenExpression                = "(" Expression ")" .
//  FunctionExpressionSuffix       = "=>" FunctionBodyExpression .
//  FunctionBodyExpression         = Block | Expression .
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[serde(untagged)]
#[allow(missing_docs)]
pub enum FunctionBody {
    Block(Block),
    Expr(Expression),
}

impl FunctionBody {
    /// Returns the [`BaseNode`] for a [`FunctionBody`].
    pub fn base(&self) -> &BaseNode {
        match self {
            FunctionBody::Block(wrapped) => &wrapped.base,
            FunctionBody::Expr(wrapped) => wrapped.base(),
        }
    }
}

fn serialize_errors<S>(errors: &[String], ser: S) -> Result<S::Ok, S::Error>
where
    S: Serializer,
{
    let mut seq = ser.serialize_seq(Some(errors.len()))?;
    for e in errors {
        let mut me: HashMap<String, &String, DefaultHasher> = HashMap::default();
        me.insert("msg".to_string(), e);
        seq.serialize_element(&me)?;
    }
    seq.end()
}

#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct Comment {
    pub text: String,
}

/// BaseNode holds the attributes every expression or statement must have.
#[derive(Default, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct BaseNode {
    #[serde(default)]
    pub location: SourceLocation,
    // If the base node is for a terminal the comments will be here. We also
    // use the base node comments when a non-terminal contains just one
    // terminal on the right hand side. This saves us populating the
    // type-specific AST nodes with comment lists when we can avoid it..
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[serde(default)]
    pub comments: Vec<Comment>,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[serde(default)]
    pub attributes: Vec<Attribute>,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[serde(serialize_with = "serialize_errors")]
    #[serde(default)]
    pub errors: Vec<String>,
}

// Custom debug implentation which reduces the size of `Debug` printing `AST`s
impl fmt::Debug for BaseNode {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        let mut f = f.debug_struct("BaseNode");
        f.field("location", &self.location);

        if !self.attributes.is_empty() {
            f.field("attributes", &self.attributes);
        }

        if !self.comments.is_empty() {
            f.field("comments", &self.comments);
        }

        if !self.errors.is_empty() {
            f.field("errors", &self.errors);
        }

        f.finish()
    }
}

impl BaseNode {
    #[allow(missing_docs)]
    pub fn is_empty(&self) -> bool {
        self.errors.is_empty() && !self.location.is_valid()
    }
    #[allow(missing_docs)]
    pub fn is_multiline(&self) -> bool {
        self.location.is_multiline()
    }
    #[allow(missing_docs)]
    pub fn set_comments(&mut self, comments: Vec<Comment>) {
        self.comments = comments;
    }
}

#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct Attribute {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,

    // Name of the attribute (such as @edition).
    // Does not include the @ symbol.
    pub name: String,

    // Attribute parameters.
    pub params: Vec<AttributeParam>,
}

#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct AttributeParam {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,

    // Represented as an Expression for simplicity when parsing,
    // but must be a literal.
    pub value: Expression,

    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[serde(default)]
    pub comma: Vec<Comment>,
}

/// Package represents a complete package source tree.
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
#[allow(missing_docs)]
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

/// Represents a source from a single file.
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
#[allow(missing_docs)]
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
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[serde(default)]
    pub eof: Vec<Comment>,
}

impl File {
    /// Reports the package name defined in the file or the default package name if not defined
    pub fn get_package(self: &File) -> &str {
        match &self.package {
            Some(pkg_clause) => pkg_clause.name.name.as_str(),
            None => DEFAULT_PACKAGE_NAME,
        }
    }
}

/// Defines the current package identifier.
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
#[allow(missing_docs)]
pub struct PackageClause {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub name: Identifier,
}

/// Declares a single import.
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
#[allow(missing_docs)]
pub struct ImportDeclaration {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    #[serde(rename = "as")]
    pub alias: Option<Identifier>,
    pub path: StringLit,
}

/// Block is a set of statements.
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
#[allow(missing_docs)]
pub struct Block {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[serde(default)]
    pub lbrace: Vec<Comment>,
    pub body: Vec<Statement>,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[serde(default)]
    pub rbrace: Vec<Comment>,
}

/// BadStmt is a placeholder for statements for which no correct statement nodes
/// can be created.
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[serde(rename = "BadStatement")]
#[allow(missing_docs)]
pub struct BadStmt {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub text: String,
}

/// ExprStmt may consist of an expression that does not return a value
/// and is executed solely for its side-effects.
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct ExprStmt {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub expression: Expression,
}

/// Defines an Expression to return.
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct ReturnStmt {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub argument: Expression,
}

/// An option statement.
///
/// Syntactically, is a single variable declaration.
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct OptionStmt {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub assignment: Assignment,
}

/// BuiltinStmt declares a builtin identifier and its struct.
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct BuiltinStmt {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[serde(default)]
    pub colon: Vec<Comment>,
    pub id: Identifier,
    pub ty: TypeExpression,
}

/// A monotype.
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
#[allow(missing_docs)]
pub enum MonoType {
    #[serde(rename = "TvarType")]
    Tvar(TvarType),
    #[serde(rename = "NamedType")]
    Basic(NamedType),
    #[serde(rename = "ArrayType")]
    Array(Box<ArrayType>),
    #[serde(rename = "StreamType")]
    Stream(Box<StreamType>),
    #[serde(rename = "VectorType")]
    Vector(Box<VectorType>),
    #[serde(rename = "DictType")]
    Dict(Box<DictType>),
    #[serde(rename = "DynamicType")]
    Dynamic(Box<DynamicType>),
    #[serde(rename = "RecordType")]
    Record(RecordType),
    #[serde(rename = "FunctionType")]
    Function(Box<FunctionType>),
    #[serde(rename = "LabelType")]
    Label(Box<LabelLit>),
}

impl MonoType {
    #[allow(missing_docs)]
    pub fn base(&self) -> &BaseNode {
        match self {
            MonoType::Basic(t) => &t.base,
            MonoType::Tvar(t) => &t.base,
            MonoType::Array(t) => &t.base,
            MonoType::Stream(t) => &t.base,
            MonoType::Vector(t) => &t.base,
            MonoType::Dict(t) => &t.base,
            MonoType::Dynamic(t) => &t.base,
            MonoType::Record(t) => &t.base,
            MonoType::Function(t) => &t.base,
            MonoType::Label(t) => &t.base,
        }
    }
}

#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct NamedType {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub name: Identifier,
}

#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct TvarType {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub name: Identifier,
}

#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct ArrayType {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub element: MonoType,
}

#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct StreamType {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub element: MonoType,
}

#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct VectorType {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub element: MonoType,
}

#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct DictType {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub key: MonoType,
    pub val: MonoType,
}

#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct DynamicType {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
}

#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct FunctionType {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub parameters: Vec<ParameterType>,
    pub monotype: MonoType,
}

#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
#[allow(missing_docs)]
pub enum ParameterType {
    Required {
        #[serde(skip_serializing_if = "BaseNode::is_empty")]
        #[serde(default)]
        #[serde(flatten)]
        base: BaseNode,
        name: Identifier,
        monotype: MonoType,
    },
    Optional {
        #[serde(skip_serializing_if = "BaseNode::is_empty")]
        #[serde(default)]
        #[serde(flatten)]
        base: BaseNode,
        name: Identifier,
        monotype: MonoType,
        // Default value for this parameter. Currently only string literals are supported
        // (to allow default labels to be specified)
        default: Option<LabelLit>,
    },
    Pipe {
        #[serde(skip_serializing_if = "BaseNode::is_empty")]
        #[serde(default)]
        #[serde(flatten)]
        base: BaseNode,
        #[serde(skip_serializing_if = "Option::is_none")]
        name: Option<Identifier>,
        monotype: MonoType,
    },
}

impl ParameterType {
    /// Returns the [`BaseNode`] for an [`ParameterType`].
    pub fn base(&self) -> &BaseNode {
        match self {
            Self::Required { base, .. } | Self::Optional { base, .. } | Self::Pipe { base, .. } => {
                base
            }
        }
    }
}

#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
#[allow(missing_docs)]
pub struct TypeExpression {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub monotype: MonoType,
    pub constraints: Vec<TypeConstraint>,
}

#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
#[allow(missing_docs)]
pub struct TypeConstraint {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub tvar: Identifier,
    pub kinds: Vec<Identifier>,
}

#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct RecordType {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub tvar: Option<Identifier>,
    pub properties: Vec<PropertyType>,
}

#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct PropertyType {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub name: PropertyKey,
    pub monotype: MonoType,
}

/// Declares a Flux test case.
// XXX: rockstar (17 Nov 2020) - This should replace the TestStmt above, once
// it has been extended enough to cover the existing use cases.
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct TestCaseStmt {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub id: Identifier,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub extends: Option<StringLit>,
    pub block: Block,
}

/// Represents the declaration of a variable.
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct VariableAssgn {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub id: Identifier,
    pub init: Expression,
}

/// Represents an assignement into a member of an object.
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct MemberAssgn {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub member: MemberExpr,
    pub init: Expression,
}

/// Represents an interpolated string.
#[allow(missing_docs)]
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
pub struct StringExpr {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub parts: Vec<StringExprPart>,
}

/// Represents part of an interpolated string.
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
#[allow(missing_docs)]
pub enum StringExprPart {
    #[serde(rename = "TextPart")]
    Text(TextPart),
    #[serde(rename = "InterpolatedPart")]
    Interpolated(InterpolatedPart),
}

/// Represents the text part of an interpolated string.
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct TextPart {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub value: String,
}

/// Represents the expression part of an interpolated string.
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct InterpolatedPart {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub expression: Expression,
}

/// Represents an expression wrapped in parenthesis.
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct ParenExpr {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[serde(default)]
    pub lparen: Vec<Comment>,
    pub expression: Expression,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[serde(default)]
    pub rparen: Vec<Comment>,
}

/// Represents a function call.
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct CallExpr {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub callee: Expression,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[serde(default)]
    pub lparen: Vec<Comment>,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[serde(default)]
    pub arguments: Vec<Expression>,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[serde(default)]
    pub rparen: Vec<Comment>,
}

/// Represents a call expression using the pipe forward syntax.
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct PipeExpr {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub argument: Expression,
    pub call: CallExpr,
}

/// Represents calling a property of a Call.
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct MemberExpr {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub object: Expression,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[serde(default)]
    pub lbrack: Vec<Comment>,
    pub property: PropertyKey,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[serde(default)]
    pub rbrack: Vec<Comment>,
}

/// Represents indexing into an array.
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct IndexExpr {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub array: Expression,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[serde(default)]
    pub lbrack: Vec<Comment>,
    pub index: Expression,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[serde(default)]
    pub rbrack: Vec<Comment>,
}

#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct FunctionExpr {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[serde(default)]
    pub lparen: Vec<Comment>,
    #[serde(deserialize_with = "deserialize_default_from_null")]
    pub params: Vec<Property>,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[serde(default)]
    pub rparen: Vec<Comment>,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[serde(default)]
    pub arrow: Vec<Comment>,
    pub body: FunctionBody,
}

/// Represents Equality and Arithmetic operators.
///
/// Result of evaluating an equality operator is always of type `bool`
/// based on whether the comparison is true.
/// Arithmetic operators take numerical values (either literals or variables)
/// as their operands and return a single numerical value.
#[derive(Debug, PartialEq, Eq, Clone)]
#[allow(missing_docs)]
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

impl fmt::Display for Operator {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        f.write_str(self.as_str())
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

macro_rules! from_to_str {
    ($name: ident, $($str: tt => $op: tt),* $(,)?) => {
        impl FromStr for $name {
            type Err = String;

            fn from_str(s: &str) -> Result<Self, Self::Err> {
                Ok(match s {
                    $(
                        $str => $name :: $op,
                    )*
                    _ => return Err(format!("unknown operator: {}", s)),
                })
            }
        }

        impl $name {
            pub(crate) fn as_str(&self) -> &'static str {
                match self {
                    $(
                    $name :: $op => $str,
                    )*
                }
            }
        }
    };
}

from_to_str! {
    Operator,
    "*" => MultiplicationOperator,
    "/" => DivisionOperator,
    "%" => ModuloOperator,
    "^" => PowerOperator,
    "+" => AdditionOperator,
    "-" => SubtractionOperator,
    "<=" => LessThanEqualOperator,
    "<" => LessThanOperator,
    ">=" => GreaterThanEqualOperator,
    ">" => GreaterThanOperator,
    "startswith" => StartsWithOperator,
    "in" => InOperator,
    "not" => NotOperator,
    "exists" => ExistsOperator,
    "not empty" => NotEmptyOperator,
    "empty" => EmptyOperator,
    "==" => EqualOperator,
    "!=" => NotEqualOperator,
    "=~" => RegexpMatchOperator,
    "!~" => NotRegexpMatchOperator,
    "<INVALID_OP>" => InvalidOperator,
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

/// BinaryExpr use binary operators act on two operands in an expression.
/// BinaryExpr includes relational and arithmetic operators
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[serde(rename = "BinaryExpression")]
#[allow(missing_docs)]
pub struct BinaryExpr {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub operator: Operator,
    pub left: Expression,
    pub right: Expression,
}

/// UnaryExpr use operators act on a single operand in an expression.
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[serde(rename = "UnaryExpression")]
#[allow(missing_docs)]
pub struct UnaryExpr {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub operator: Operator,
    pub argument: Expression,
}

/// LogicalOperator are used with boolean (logical) values.
#[derive(Debug, PartialEq, Eq, Clone)]
#[allow(missing_docs)]
pub enum LogicalOperator {
    AndOperator,
    OrOperator,
}

from_to_str! {
    LogicalOperator,
    "and" => AndOperator,
    "or" => OrOperator,
}

impl fmt::Display for LogicalOperator {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        f.write_str(self.as_str())
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

/// LogicalExpr represent the rule conditions that collectively evaluate to either true or false.
/// `or` expressions compute the disjunction of two boolean expressions and return boolean values.
/// `and`` expressions compute the conjunction of two boolean expressions and return boolean values.
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct LogicalExpr {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub operator: LogicalOperator,
    pub left: Expression,
    pub right: Expression,
}

/// ArrayExpr is used to create and directly specify the elements of an array object
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct ArrayItem {
    #[serde(default)]
    #[serde(flatten)]
    pub expression: Expression,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[serde(default)]
    pub comma: Vec<Comment>,
}

/// ArrayExpr is used to create and directly specify the elements of an array object
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct ArrayExpr {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[serde(default)]
    pub lbrack: Vec<Comment>,
    #[serde(deserialize_with = "deserialize_default_from_null")]
    pub elements: Vec<ArrayItem>,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[serde(default)]
    pub rbrack: Vec<Comment>,
}

/// DictExpr represents a dictionary literal
#[allow(missing_docs)]
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
pub struct DictExpr {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[serde(default)]
    pub lbrack: Vec<Comment>,
    #[serde(deserialize_with = "deserialize_default_from_null")]
    pub elements: Vec<DictItem>,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[serde(default)]
    pub rbrack: Vec<Comment>,
}

#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct DictItem {
    pub key: Expression,
    pub val: Expression,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[serde(default)]
    pub comma: Vec<Comment>,
}

#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct WithSource {
    #[serde(default)]
    #[serde(flatten)]
    pub source: Identifier,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[serde(default)]
    pub with: Vec<Comment>,
}

/// ObjectExpr allows the declaration of an anonymous object within a declaration.
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct ObjectExpr {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[serde(default)]
    pub lbrace: Vec<Comment>,
    #[serde(skip_serializing_if = "Option::is_none")]
    #[serde(default)]
    pub with: Option<WithSource>,
    #[serde(deserialize_with = "deserialize_default_from_null")]
    pub properties: Vec<Property>,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[serde(default)]
    pub rbrace: Vec<Comment>,
}

/// ConditionalExpr selects one of two expressions, `Alternate` or `Consequent`
/// depending on a third, boolean, expression, `Test`.
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct ConditionalExpr {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[serde(default)]
    pub tk_if: Vec<Comment>,
    pub test: Expression,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[serde(default)]
    pub tk_then: Vec<Comment>,
    pub consequent: Expression,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[serde(default)]
    pub tk_else: Vec<Comment>,
    pub alternate: Expression,
}

/// BadExpr is a malformed expression that contains the reason why in `text`.
/// It can contain another expression, so that the parser can make a chained list of bad expressions.
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct BadExpr {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub text: String,
    pub expression: Option<Expression>,
}

/// Property is the value associated with a key.
/// A property's key can be either an identifier or string literal.
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
#[allow(missing_docs)]
pub struct Property {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub key: PropertyKey,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[serde(default)]
    pub separator: Vec<Comment>,
    // `value` is optional, because of the shortcut: {a} <--> {a: a}
    pub value: Option<Expression>,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[serde(default)]
    pub comma: Vec<Comment>,
}

/// Identifier represents a name that identifies a unique Node
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct Identifier {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub name: String,
}

/// PipeLit represents an specialized literal value, indicating the left hand value of a pipe expression.
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct PipeLit {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
}

/// StringLit expressions begin and end with double quote marks.
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct StringLit {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub value: String,
}

/// Boolean represent boolean values
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct BooleanLit {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub value: bool,
}

/// Represent floating point numbers according to the double representations
/// defined by [IEEE-754-1985](https://en.wikipedia.org/wiki/IEEE_754-1985).
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct FloatLit {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub value: NotNan<f64>,
}

/// Represents integer numbers.
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct IntegerLit {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    #[serde(serialize_with = "serialize_to_string")]
    #[serde(deserialize_with = "deserialize_str_i64")]
    pub value: i64,
}

/// Represents integer numbers.
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct UintLit {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    #[serde(serialize_with = "serialize_to_string")]
    #[serde(deserialize_with = "deserialize_str_u64")]
    pub value: u64,
}

/// LabelLit represents a label. Used to specify specific record fields.
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct LabelLit {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,

    pub value: String,
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

/// RegexpLit expressions begin and end with `/` and are regular expressions with syntax accepted by RE2.
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct RegexpLit {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub value: String,
}

/// DurationLit is a pair consisting of length of time and the unit of time measured.
/// It is the atomic unit from which all duration literals are composed.
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[serde(rename = "Duration")]
#[allow(missing_docs)]
pub struct Duration {
    pub magnitude: i64,
    pub unit: String,
}

/// DurationLit represents the elapsed time between two instants as an
/// int64 nanosecond count with syntax of [golang's time.Duration].
///
/// [golang's time.Duration]: https://golang.org/pkg/time/#Duration
// TODO: this may be better as a class initialization
// All magnitudes in Duration vector should have the same sign
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct DurationLit {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub values: Vec<Duration>,
}

/// DateTimeLit represents an instant in time with nanosecond precision using
/// the syntax of golang's RFC3339 Nanosecond variant.
// TODO: we need a "duration from" that takes a time and a durationliteral, and gives an exact time.DurationLit instead of an approximation
// TODO: this may be better as a class initialization
#[derive(Debug, PartialEq, Eq, Clone, Serialize, Deserialize)]
#[allow(missing_docs)]
pub struct DateTimeLit {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    #[serde(default)]
    #[serde(flatten)]
    pub base: BaseNode,
    pub value: chrono::DateTime<FixedOffset>,
}

// Re-implementation of https://github.com/vityafx/serde-aux/blob/c6f8482f51da7f187ecea62931c8f38edcf355c9/src/field_attributes.rs#L676
// so we do not need to pull in an entire crate
fn deserialize_default_from_null<'de, T, D>(d: D) -> Result<T, D::Error>
where
    D: Deserializer<'de>,
    T: Deserialize<'de> + Default,
{
    Ok(Option::<T>::deserialize(d)?.unwrap_or_default())
}

// The tests code exports a few helpers for writing AST related tests.
// We make it public so other tests can consume those helpers.
#[cfg(test)]
pub mod tests;
