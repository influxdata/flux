use std::time::SystemTime;
use std::vec::Vec;

#[derive(Debug, PartialEq, Clone)]
pub enum Expression {
    Identifier(Identifier),
}

#[derive(Debug, PartialEq, Clone)]
pub enum Statement {
    Expression(ExpressionStatement),
    Return(ReturnStatement),
    Variable(VariableAssignment),
}

#[derive(Debug, PartialEq, Clone)]
pub enum Assignment {
    Variable(VariableAssignment),
    Member(MemberAssignment),
}

#[derive(Debug, PartialEq, Clone)]
pub enum PropertyKey {
    Identifier(Identifier),
    StringLiteral(StringLiteral),
}

// BaseNode holds the attributes every expression or statement should have
#[derive(Debug, PartialEq, Clone)]
pub struct BaseNode {
    //pub  Loc   : *SourceLocation,
    pub errors: Vec<String>,
}

// Package represents a complete package source tree
#[derive(Debug, PartialEq, Clone)]
pub struct Package {
    pub base: BaseNode,
    pub path: String,
    pub package: String,
}

// File represents a source from a single file
#[derive(Debug, PartialEq, Clone)]
pub struct File {
    pub base: BaseNode,
    pub name: String,
    pub package: Option<PackageClause>,
    pub imports: Vec<ImportDeclaration>,
    pub body: Vec<Statement>,
}

// PackageClause defines the current package identifier.
#[derive(Debug, PartialEq, Clone)]
pub struct PackageClause {
    pub base: BaseNode,
    pub name: Identifier,
}

// ImportDeclaration declares a single import
#[derive(Debug, PartialEq, Clone)]
pub struct ImportDeclaration {
    pub base: BaseNode,
    pub alias: Option<Identifier>,
    pub path: StringLiteral,
}

// Block is a set of statements
#[derive(Debug, PartialEq, Clone)]
pub struct Block {
    pub base: BaseNode,
    pub body: Vec<Statement>,
}

// BadStatement is a placeholder for statements for which no correct statement nodes
// can be created.
#[derive(Debug, PartialEq, Clone)]
pub struct BadStatement {
    pub base: BaseNode,
    pub text: String,
}

// ExpressionStatement may consist of an expression that does not return a value and is executed solely for its side-effects.
#[derive(Debug, PartialEq, Clone)]
pub struct ExpressionStatement {
    pub base: BaseNode,
    pub expression: Expression,
}

// ReturnStatement defines an Expression to return
#[derive(Debug, PartialEq, Clone)]
pub struct ReturnStatement {
    pub base: BaseNode,
    pub argument: Expression,
}

// OptionStatement syntactically is a single variable declaration
#[derive(Debug, PartialEq, Clone)]
pub struct OptionStatement {
    pub base: BaseNode,
    pub assignment: Assignment,
}

// BuiltinStatement declares a builtin identifier and its struct
#[derive(Debug, PartialEq, Clone)]
pub struct BuiltinStatement {
    pub base: BaseNode,
    pub id: Identifier,
}

// TestStatement declares a Flux test case
#[derive(Debug, PartialEq, Clone)]
pub struct TestStatement {
    pub base: BaseNode,
    pub assignment: VariableAssignment,
}

// VariableAssignment represents the declaration of a variable
#[derive(Debug, PartialEq, Clone)]
pub struct VariableAssignment {
    pub base: BaseNode,
    pub id: Identifier,
    pub init: Expression,
}

#[derive(Debug, PartialEq, Clone)]
pub struct MemberAssignment {
    pub base: BaseNode,
    pub member: MemberExpression,
    pub init: Expression,
}

// CallExpression represents a function call
#[derive(Debug, PartialEq, Clone)]
pub struct CallExpression {
    pub base: BaseNode,
    pub callee: Expression,
    pub arguments: Vec<Expression>,
}

#[derive(Debug, PartialEq, Clone)]
pub struct PipeExpression {
    pub base: BaseNode,
    pub argument: Expression,
    pub call: CallExpression,
}

// MemberExpression represents calling a property of a CallExpression
#[derive(Debug, PartialEq, Clone)]
pub struct MemberExpression {
    pub base: BaseNode,
    pub object: Expression,
    pub property: PropertyKey,
}

// IndexExpression represents indexing into an array
#[derive(Debug, PartialEq, Clone)]
pub struct IndexExpression {
    pub base: BaseNode,
    pub array: Expression,
    pub index: Expression,
}

#[derive(Debug, PartialEq, Clone)]
pub struct FunctionExpression {
    pub base: BaseNode,
    pub params: Vec<Property>,
    //pub body: Node,
}

// OperatorKind are Equality and Arithmatic operators.
// Result of evaluating an equality operator is always of type Boolean based on whether the
// comparison is true
// Arithmetic operators take numerical values (either literals or variables) as their operands
//  and return a single numerical value.
#[derive(Debug, PartialEq, Clone)]
pub enum OperatorKind {
    MultiplicationOperator,
    DivisionOperator,
    AdditionOperator,
    SubtractionOperator,
    LessThanEqualOperator,
    LessThanOperator,
    GreaterThanEqualOperator,
    GreaterThanOperator,
    StartsWithOperator,
    InOperator,
    NotOperator,
    NotEmptyOperator,
    EmptyOperator,
    EqualOperator,
    NotEqualOperator,
    RegexpMatchOperator,
    NotRegexpMatchOperator,
}

// BinaryExpression use binary operators act on two operands in an expression.
// BinaryExpression includes relational and arithmatic operators
#[derive(Debug, PartialEq, Clone)]
pub struct BinaryExpression {
    pub base: BaseNode,
    pub operator: OperatorKind,
    pub left: Expression,
    pub right: Expression,
}

// UnaryExpression use operators act on a single operand in an expression.
#[derive(Debug, PartialEq, Clone)]
pub struct UnaryExpression {
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

// LogicalExpression represent the rule conditions that collectively evaluate to either true or false.
// `or` expressions compute the disjunction of two boolean expressions and return boolean values.
// `and`` expressions compute the conjunction of two boolean expressions and return boolean values.
#[derive(Debug, PartialEq, Clone)]
pub struct LogicalExpression {
    pub base: BaseNode,
    pub operator: LogicalOperatorKind,
    pub left: Expression,
    pub right: Expression,
}

// ArrayExpression is used to create and directly specify the elements of an array object
#[derive(Debug, PartialEq, Clone)]
pub struct ArrayExpression {
    pub base: BaseNode,
    pub elements: Vec<Expression>,
}

// ObjectExpression allows the declaration of an anonymous object within a declaration.
#[derive(Debug, PartialEq, Clone)]
pub struct ObjectExpression {
    pub base: BaseNode,
    pub properties: Vec<Property>,
}

// ConditionalExpression selects one of two expressions, `Alternate` or `Consequent`
// depending on a third, boolean, expression, `Test`.
#[derive(Debug, PartialEq, Clone)]
pub struct ConditionalExpression {
    pub base: BaseNode,
    pub test: Expression,
    pub consequent: Expression,
    pub alternate: Expression,
}

// Property is the value associated with a key.
// A property's key can be either an identifier or string literal.
#[derive(Debug, PartialEq, Clone)]
pub struct Property {
    pub base: BaseNode,
    pub key: PropertyKey,
    pub value: Expression,
}

// Identifier represents a name that identifies a unique Node
#[derive(Debug, PartialEq, Clone)]
pub struct Identifier {
    pub base: BaseNode,
    pub name: String,
}

// PipeLiteral represents an specialized literal value, indicating the left hand value of a pipe expression.
#[derive(Debug, PartialEq, Clone)]
pub struct PipeLiteral {
    pub base: BaseNode,
}

// StringLiteral expressions begin and end with double quote marks.
#[derive(Debug, PartialEq, Clone)]
pub struct StringLiteral {
    pub base: BaseNode,
    pub value: String,
}

// BooleanLiteral represent boolean values
#[derive(Debug, PartialEq, Clone)]
pub struct BooleanLiteral {
    pub base: BaseNode,
    pub value: bool,
}

// FloatLiteral  represent floating point numbers according to the double representations defined by the IEEE-754-1985
#[derive(Debug, PartialEq, Clone)]
pub struct FloatLiteral {
    pub base: BaseNode,
    pub value: f64,
}

// IntegerLiteral represent integer numbers.
#[derive(Debug, PartialEq, Clone)]
pub struct IntegerLiteral {
    pub base: BaseNode,
    pub value: i64,
}

// UnsignedIntegerLiteral represent integer numbers.
#[derive(Debug, PartialEq, Clone)]
pub struct UnsignedIntegerLiteral {
    pub base: BaseNode,
    pub value: u64,
}

// RegexpLiteral expressions begin and end with `/` and are regular expressions with syntax accepted by RE2
#[derive(Debug, PartialEq, Clone)]
pub struct RegexpLiteral {
    pub base: BaseNode,
    pub value: String,
}

// Duration is a pair consisting of length of time and the unit of time measured.
// It is the atomic unit from which all duration literals are composed.
#[derive(Debug, PartialEq, Clone)]
pub struct Duration {
    pub magnitude: i64,
    pub unit: String,
}

// DurationLiteral represents the elapsed time between two instants as an
// int64 nanosecond count with syntax of golang's time.Duration
// TODO: this may be better as a class initialization
#[derive(Debug, PartialEq, Clone)]
pub struct DurationLiteral {
    pub base: BaseNode,
    pub values: Vec<Duration>,
}

// TODO: we need a "duration from" that takes a time and a durationliteral, and gives an exact time.Duration instead of an approximation
//
// DateTimeLiteral represents an instant in time with nanosecond precision using
// the syntax of golang's RFC3339 Nanosecond variant
// TODO: this may be better as a class initialization
#[derive(Debug, PartialEq, Clone)]
pub struct DateTimeLiteral {
    pub base: BaseNode,
    pub value: SystemTime,
}
