extern crate chrono;
#[macro_use]
extern crate serde_derive;

use std::time::SystemTime;
use std::vec::Vec;

use chrono::{TimeZone, Utc};
use chrono::prelude::DateTime;
use serde::ser::{Serialize, Serializer};

// serialize_to_string serializes an object that implements ToString to its string representation.
fn serialize_to_string<T, S>(field: &T, ser: S) -> Result<S::Ok, S::Error> where S: Serializer, T: ToString {
    let s = field.to_string();
    ser.serialize_str(s.as_str())
}

// TODO(affo): this enums do not match ast.go because recursive types have infinite size in Rust.
//  We can fix that by adding indirection (&) when, for instance, an Expression contains an Expression.

#[derive(Debug, PartialEq, Clone, Serialize)]
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
    Pipe(Box<PipeExpression>),
    Call(Box<CallExpression>),
    Cond(Box<ConditionalExpression>),

    Int(IntegerLiteral),
    Flt(FloatLiteral),
    Str(StringLiteral),
    Dur(DurationLiteral),
    Uint(UnsignedIntegerLiteral),
    Bool(BooleanLiteral),
    Time(DateTimeLiteral),
    Regexp(RegexpLiteral),
    PipeLit(PipeLiteral),
}

#[derive(Debug, PartialEq, Clone, Serialize)]
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

#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(untagged)]
pub enum Assignment {
    Variable(VariableAssignment),
    Member(MemberAssignment),
}

#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(untagged)]
pub enum PropertyKey {
    Identifier(Identifier),
    StringLiteral(StringLiteral),
}

// This matches the grammar, and not ast.go:
//  ParenExpression                = "(" Expression ")" .
//  FunctionExpressionSuffix       = "=>" FunctionBodyExpression .
//  FunctionBodyExpression         = Block | Expression .
#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(untagged)]
pub enum FunctionBody {
    Block(Block),
    Expr(Expression),
}

// BaseNode holds the attributes every expression or statement should have
#[derive(Debug, Default, PartialEq, Clone, Serialize)]
#[serde(tag = "type")]
pub struct BaseNode {
    //pub  Loc   : *SourceLocation,
    pub errors: Vec<String>,
}

impl BaseNode {
    pub fn is_empty(&self) -> bool {
        self.errors.is_empty()
    }
}

// Package represents a complete package source tree
#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(tag = "type")]
pub struct Package {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    pub base: BaseNode,
    #[serde(skip_serializing_if = "String::is_empty")]
    pub path: String,
    pub package: String,
    pub files: Vec<File>,
}

// File represents a source from a single file
#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(tag = "type")]
pub struct File {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    pub base: BaseNode,
    #[serde(skip_serializing_if = "String::is_empty")]
    pub name: String,
    pub package: Option<PackageClause>,
    pub imports: Vec<ImportDeclaration>,
    pub body: Vec<Statement>,
}

// PackageClause defines the current package identifier.
#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(tag = "type")]
pub struct PackageClause {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    pub base: BaseNode,
    pub name: Identifier,
}

// ImportDeclaration declares a single import
#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(tag = "type")]
pub struct ImportDeclaration {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    pub base: BaseNode,
    #[serde(rename(serialize = "as"))]
    pub alias: Option<Identifier>,
    pub path: StringLiteral,
}

// Block is a set of statements
#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(tag = "type")]
pub struct Block {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    pub base: BaseNode,
    pub body: Vec<Statement>,
}

// BadStatement is a placeholder for statements for which no correct statement nodes
// can be created.
#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(tag = "type")]
pub struct BadStatement {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    pub base: BaseNode,
    pub text: String,
}

// ExpressionStatement may consist of an expression that does not return a value and is executed solely for its side-effects.
#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(tag = "type")]
pub struct ExpressionStatement {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    pub base: BaseNode,
    pub expression: Expression,
}

// ReturnStatement defines an Expression to return
#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(tag = "type")]
pub struct ReturnStatement {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    pub base: BaseNode,
    pub argument: Expression,
}

// OptionStatement syntactically is a single variable declaration
#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(tag = "type")]
pub struct OptionStatement {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    pub base: BaseNode,
    pub assignment: Assignment,
}

// BuiltinStatement declares a builtin identifier and its struct
#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(tag = "type")]
pub struct BuiltinStatement {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    pub base: BaseNode,
    pub id: Identifier,
}

// TestStatement declares a Flux test case
#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(tag = "type")]
pub struct TestStatement {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    pub base: BaseNode,
    pub assignment: VariableAssignment,
}

// VariableAssignment represents the declaration of a variable
#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(tag = "type")]
pub struct VariableAssignment {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    pub base: BaseNode,
    pub id: Identifier,
    pub init: Expression,
}

#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(tag = "type")]
pub struct MemberAssignment {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    pub base: BaseNode,
    pub member: MemberExpression,
    pub init: Expression,
}

// CallExpression represents a function call
#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(tag = "type")]
pub struct CallExpression {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    pub base: BaseNode,
    pub callee: Expression,
    pub arguments: Vec<Expression>,
}

#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(tag = "type")]
pub struct PipeExpression {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    pub base: BaseNode,
    pub argument: Expression,
    pub call: CallExpression,
}

// MemberExpression represents calling a property of a CallExpression
#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(tag = "type")]
pub struct MemberExpression {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    pub base: BaseNode,
    pub object: Expression,
    pub property: PropertyKey,
}

// IndexExpression represents indexing into an array
#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(tag = "type")]
pub struct IndexExpression {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    pub base: BaseNode,
    pub array: Expression,
    pub index: Expression,
}

#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(tag = "type")]
pub struct FunctionExpression {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
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
}

impl ToString for OperatorKind {
    fn to_string(&self) -> String {
        match self {
            OperatorKind::MultiplicationOperator => "*".to_string(),
            OperatorKind::DivisionOperator => "/".to_string(),
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
        }
    }
}

impl Serialize for OperatorKind {
    fn serialize<S>(&self, serializer: S) -> Result<<S as Serializer>::Ok, <S as Serializer>::Error> where
        S: Serializer {
        serialize_to_string(self, serializer)
    }
}

// BinaryExpression use binary operators act on two operands in an expression.
// BinaryExpression includes relational and arithmetic operators
#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(tag = "type")]
pub struct BinaryExpression {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    pub base: BaseNode,
    pub operator: OperatorKind,
    pub left: Expression,
    pub right: Expression,
}

// UnaryExpression use operators act on a single operand in an expression.
#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(tag = "type")]
pub struct UnaryExpression {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
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
    fn serialize<S>(&self, serializer: S) -> Result<<S as Serializer>::Ok, <S as Serializer>::Error> where
        S: Serializer {
        serialize_to_string(self, serializer)
    }
}

// LogicalExpression represent the rule conditions that collectively evaluate to either true or false.
// `or` expressions compute the disjunction of two boolean expressions and return boolean values.
// `and`` expressions compute the conjunction of two boolean expressions and return boolean values.
#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(tag = "type")]
pub struct LogicalExpression {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    pub base: BaseNode,
    pub operator: LogicalOperatorKind,
    pub left: Expression,
    pub right: Expression,
}

// ArrayExpression is used to create and directly specify the elements of an array object
#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(tag = "type")]
pub struct ArrayExpression {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    pub base: BaseNode,
    pub elements: Vec<Expression>,
}

// ObjectExpression allows the declaration of an anonymous object within a declaration.
#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(tag = "type")]
pub struct ObjectExpression {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    pub base: BaseNode,
    pub properties: Vec<Property>,
}

// ConditionalExpression selects one of two expressions, `Alternate` or `Consequent`
// depending on a third, boolean, expression, `Test`.
#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(tag = "type")]
pub struct ConditionalExpression {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    pub base: BaseNode,
    pub test: Expression,
    pub consequent: Expression,
    pub alternate: Expression,
}

// Property is the value associated with a key.
// A property's key can be either an identifier or string literal.
#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(tag = "type")]
pub struct Property {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    pub base: BaseNode,
    pub key: PropertyKey,
    // `value` is optional, because of the shortcut: {a} <--> {a: a}
    pub value: Option<Expression>,
}

// Identifier represents a name that identifies a unique Node
#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(tag = "type")]
pub struct Identifier {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    pub base: BaseNode,
    pub name: String,
}

// PipeLiteral represents an specialized literal value, indicating the left hand value of a pipe expression.
#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(tag = "type")]
pub struct PipeLiteral {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    pub base: BaseNode,
}

// StringLiteral expressions begin and end with double quote marks.
#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(tag = "type")]
pub struct StringLiteral {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    pub base: BaseNode,
    pub value: String,
}

// BooleanLiteral represent boolean values
#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(tag = "type")]
pub struct BooleanLiteral {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    pub base: BaseNode,
    pub value: bool,
}

// FloatLiteral  represent floating point numbers according to the double representations defined by the IEEE-754-1985
#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(tag = "type")]
pub struct FloatLiteral {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    pub base: BaseNode,
    pub value: f64,
}

// IntegerLiteral represent integer numbers.
#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(tag = "type")]
pub struct IntegerLiteral {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    pub base: BaseNode,
    #[serde(serialize_with = "serialize_to_string")]
    pub value: i64,
}

// UnsignedIntegerLiteral represent integer numbers.
#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(tag = "type")]
pub struct UnsignedIntegerLiteral {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    pub base: BaseNode,
    #[serde(serialize_with = "serialize_to_string")]
    pub value: u64,
}

// RegexpLiteral expressions begin and end with `/` and are regular expressions with syntax accepted by RE2
#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(tag = "type")]
pub struct RegexpLiteral {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    pub base: BaseNode,
    pub value: String,
}

// Duration is a pair consisting of length of time and the unit of time measured.
// It is the atomic unit from which all duration literals are composed.
#[derive(Debug, PartialEq, Clone, Serialize)]
pub struct Duration {
    pub magnitude: i64,
    pub unit: String,
}

// DurationLiteral represents the elapsed time between two instants as an
// int64 nanosecond count with syntax of golang's time.Duration
// TODO: this may be better as a class initialization
#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(tag = "type")]
pub struct DurationLiteral {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    pub base: BaseNode,
    pub values: Vec<Duration>,
}

// TODO: we need a "duration from" that takes a time and a durationliteral, and gives an exact time.Duration instead of an approximation
//
// DateTimeLiteral represents an instant in time with nanosecond precision using
// the syntax of golang's RFC3339 Nanosecond variant
// TODO: this may be better as a class initialization
#[derive(Debug, PartialEq, Clone, Serialize)]
#[serde(tag = "type")]
pub struct DateTimeLiteral {
    #[serde(skip_serializing_if = "BaseNode::is_empty")]
    pub base: BaseNode,
    pub value: DateTime<Utc>,
}

// NOTE: These test cases directly match ast/json_test.go.
// Every test is preceded by the correspondent test case in golang.
#[cfg(test)]
mod tests {
    use serde_json::error::ErrorCode::ExpectedColon;

    use super::*;

    /*
    {
        name: "simple package",
        node: &ast.Package{
            Package: "foo",
        },
        want: `{"type":"Package","package":"foo","files":null}`,
    },
    */
    #[test] // NOTE: adapted for non-nullable files.
    fn test_json_simple_package() {
        let n = Package {
            base: BaseNode::default(),
            path: String::new(),
            package: "foo".to_string(),
            files: Vec::new(),
        };
        let serialized = serde_json::to_string(&n).unwrap();
        assert_eq!(serialized, r#"{"type":"Package","package":"foo","files":[]}"#)
    }
    /*
    {
        name: "package path",
        node: &ast.Package{
            Path:    "bar/foo",
            Package: "foo",
        },
        want: `{"type":"Package","path":"bar/foo","package":"foo","files":null}`,
    },
    */
    #[test] // NOTE: adapted for non-nullable files.
    fn test_json_package_path() {
        let n = Package {
            base: BaseNode::default(),
            path: "bar/foo".to_string(),
            package: "foo".to_string(),
            files: Vec::new(),
        };
        let serialized = serde_json::to_string(&n).unwrap();
        assert_eq!(serialized, r#"{"type":"Package","path":"bar/foo","package":"foo","files":[]}"#)
    }
    /*
    {
        name: "simple file",
        node: &ast.File{
            Body: []ast.Statement{
                &ast.ExpressionStatement{
                    Expression: &ast.StringLiteral{Value: "hello"},
                },
            },
        },
        want: `{"type":"File","package":null,"imports":null,"body":[{"type":"ExpressionStatement","expression":{"type":"StringLiteral","value":"hello"}}]}`,
    },
    */
    #[test] // NOTE: adapted for non-nullable imports.
    fn test_json_simple_file() {
        let n = File {
            base: BaseNode::default(),
            package: Option::None,
            imports: Vec::new(),
            name: String::new(),
            body: vec![
                Statement::Expression(ExpressionStatement {
                    base: BaseNode::default(),
                    expression: Expression::String(StringLiteral {
                        base: Default::default(),
                        value: "hello".to_string(),
                    }),
                }),
            ],
        };
        let serialized = serde_json::to_string(&n).unwrap();
        assert_eq!(serialized, r#"{"type":"File","package":null,"imports":[],"body":[{"type":"ExpressionStatement","expression":{"type":"StringLiteral","value":"hello"}}]}"#)
    }
    /*
    {
        name: "file",
        node: &ast.File{
            Package: &ast.PackageClause{
                Name: &ast.Identifier{Name: "foo"},
            },
            Imports: []*ast.ImportDeclaration{{
                As:   &ast.Identifier{Name: "b"},
                Path: &ast.StringLiteral{Value: "path/bar"},
            }},
            Body: []ast.Statement{
                &ast.ExpressionStatement{
                    Expression: &ast.StringLiteral{Value: "hello"},
                },
            },
        },
        want: `{"type":"File","package":{"type":"PackageClause","name":{"type":"Identifier","name":"foo"}},"imports":[{"type":"ImportDeclaration","as":{"type":"Identifier","name":"b"},"path":{"type":"StringLiteral","value":"path/bar"}}],"body":[{"type":"ExpressionStatement","expression":{"type":"StringLiteral","value":"hello"}}]}`,
    },
	*/
    #[test]
    fn test_json_file() {
        let n = File {
            base: BaseNode::default(),
            package: Some(PackageClause {
                base: BaseNode::default(),
                name: Identifier {
                    base: Default::default(),
                    name: "foo".to_string(),
                },
            }),
            imports: vec![
                ImportDeclaration {
                    base: BaseNode::default(),
                    alias: Some(Identifier {
                        base: Default::default(),
                        name: "b".to_string(),
                    }),
                    path: StringLiteral {
                        base: BaseNode::default(),
                        value: "path/bar".to_string(),
                    },
                }
            ],
            name: String::new(),
            body: vec![
                Statement::Expression(ExpressionStatement {
                    base: BaseNode::default(),
                    expression: Expression::String(StringLiteral {
                        base: Default::default(),
                        value: "hello".to_string(),
                    }),
                }),
            ],
        };
        let serialized = serde_json::to_string(&n).unwrap();
        assert_eq!(serialized, r#"{"type":"File","package":{"type":"PackageClause","name":{"type":"Identifier","name":"foo"}},"imports":[{"type":"ImportDeclaration","as":{"type":"Identifier","name":"b"},"path":{"type":"StringLiteral","value":"path/bar"}}],"body":[{"type":"ExpressionStatement","expression":{"type":"StringLiteral","value":"hello"}}]}"#)
    }
    /*
    {
        name: "block",
        node: &ast.Block{
            Body: []ast.Statement{
                &ast.ExpressionStatement{
                    Expression: &ast.StringLiteral{Value: "hello"},
                },
            },
        },
        want: `{"type":"Block","body":[{"type":"ExpressionStatement","expression":{"type":"StringLiteral","value":"hello"}}]}`,
    },
	*/
    #[test]
    fn test_json_block() {
        let n = Block {
            base: BaseNode::default(),
            body: vec![
                Statement::Expression(ExpressionStatement {
                    base: BaseNode::default(),
                    expression: Expression::String(StringLiteral {
                        base: Default::default(),
                        value: "hello".to_string(),
                    }),
                }),
            ],
        };
        let serialized = serde_json::to_string(&n).unwrap();
        assert_eq!(serialized, r#"{"type":"Block","body":[{"type":"ExpressionStatement","expression":{"type":"StringLiteral","value":"hello"}}]}"#)
    }
    /*
    {
        name: "expression statement",
        node: &ast.ExpressionStatement{
            Expression: &ast.StringLiteral{Value: "hello"},
        },
        want: `{"type":"ExpressionStatement","expression":{"type":"StringLiteral","value":"hello"}}`,
    },
	*/
    #[test]
    fn test_json_expression_statement() {
        let n = ExpressionStatement {
            base: BaseNode::default(),
            expression: Expression::String(StringLiteral {
                base: BaseNode::default(),
                value: "hello".to_string(),
            }),
        };
        let serialized = serde_json::to_string(&n).unwrap();
        assert_eq!(serialized, r#"{"type":"ExpressionStatement","expression":{"type":"StringLiteral","value":"hello"}}"#)
    }
    /*
    {
        name: "return statement",
        node: &ast.ReturnStatement{
            Argument: &ast.StringLiteral{Value: "hello"},
        },
        want: `{"type":"ReturnStatement","argument":{"type":"StringLiteral","value":"hello"}}`,
    },
	*/
    #[test]
    fn test_json_return_statement() {
        let n = ReturnStatement {
            base: BaseNode::default(),
            argument: Expression::String(StringLiteral {
                base: BaseNode::default(),
                value: "hello".to_string(),
            }),
        };
        let serialized = serde_json::to_string(&n).unwrap();
        assert_eq!(serialized, r#"{"type":"ReturnStatement","argument":{"type":"StringLiteral","value":"hello"}}"#)
    }
    /*
    {
        name: "option statement",
        node: &ast.OptionStatement{
            Assignment: &ast.VariableAssignment{
                ID: &ast.Identifier{Name: "task"},
                Init: &ast.ObjectExpression{
                    Properties: []*ast.Property{
                        {
                            Key:   &ast.Identifier{Name: "name"},
                            Value: &ast.StringLiteral{Value: "foo"},
                        },
                        {
                            Key: &ast.Identifier{Name: "every"},
                            Value: &ast.DurationLiteral{
                                Values: []ast.Duration{
                                    {
                                        Magnitude: 1,
                                        Unit:      "h",
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
        want: `{"type":"OptionStatement","assignment":{"type":"VariableAssignment","id":{"type":"Identifier","name":"task"},"init":{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"Identifier","name":"name"},"value":{"type":"StringLiteral","value":"foo"}},{"type":"Property","key":{"type":"Identifier","name":"every"},"value":{"type":"DurationLiteral","values":[{"magnitude":1,"unit":"h"}]}}]}}}`,
    },
	*/
    #[test]
    fn test_json_option_statement() {
        let n = OptionStatement {
            base: BaseNode::default(),
            assignment: Assignment::Variable(VariableAssignment {
                base: BaseNode::default(),
                id: Identifier {
                    base: BaseNode::default(),
                    name: "task".to_string(),
                },
                init: Expression::Object(ObjectExpression {
                    base: BaseNode::default(),
                    properties: vec![
                        Property {
                            base: BaseNode::default(),
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode::default(),
                                name: "name".to_string(),
                            }),
                            value: Some(Expression::String(StringLiteral {
                                base: Default::default(),
                                value: "foo".to_string(),
                            })),
                        },
                        Property {
                            base: BaseNode::default(),
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode::default(),
                                name: "every".to_string(),
                            }),
                            value: Some(Expression::Duration(DurationLiteral {
                                base: Default::default(),
                                values: vec![
                                    Duration {
                                        magnitude: 1,
                                        unit: "h".to_string(),
                                    }
                                ],
                            })),
                        },
                    ],
                }),
            }),
        };
        let serialized = serde_json::to_string(&n).unwrap();
        assert_eq!(serialized, r#"{"type":"OptionStatement","assignment":{"type":"VariableAssignment","id":{"type":"Identifier","name":"task"},"init":{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"Identifier","name":"name"},"value":{"type":"StringLiteral","value":"foo"}},{"type":"Property","key":{"type":"Identifier","name":"every"},"value":{"type":"DurationLiteral","values":[{"magnitude":1,"unit":"h"}]}}]}}}"#)
    }
    /*
    {
        name: "builtin statement",
        node: &ast.BuiltinStatement{
            ID: &ast.Identifier{Name: "task"},
        },
        want: `{"type":"BuiltinStatement","id":{"type":"Identifier","name":"task"}}`,
    },
	*/
    #[test]
    fn test_json_builtin_statement() {
        let n = BuiltinStatement {
            base: BaseNode::default(),
            id: Identifier {
                base: BaseNode::default(),
                name: "task".to_string(),
            },
        };
        let serialized = serde_json::to_string(&n).unwrap();
        assert_eq!(serialized, r#"{"type":"BuiltinStatement","id":{"type":"Identifier","name":"task"}}"#)
    }
    /*
    {
        name: "test statement",
        node: &ast.TestStatement{
            Assignment: &ast.VariableAssignment{
                ID: &ast.Identifier{Name: "mean"},
                Init: &ast.ObjectExpression{
                    Properties: []*ast.Property{
                        {
                            Key: &ast.Identifier{
                                Name: "want",
                            },
                            Value: &ast.IntegerLiteral{
                                Value: 0,
                            },
                        },
                        {
                            Key: &ast.Identifier{
                                Name: "got",
                            },
                            Value: &ast.IntegerLiteral{
                                Value: 0,
                            },
                        },
                    },
                },
            },
        },
        want: `{"type":"TestStatement","assignment":{"type":"VariableAssignment","id":{"type":"Identifier","name":"mean"},"init":{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"Identifier","name":"want"},"value":{"type":"IntegerLiteral","value":"0"}},{"type":"Property","key":{"type":"Identifier","name":"got"},"value":{"type":"IntegerLiteral","value":"0"}}]}}}`,
    },
	*/
    #[test]
    fn test_json_test_statement() {
        let n = TestStatement {
            base: BaseNode::default(),
            assignment: VariableAssignment {
                base: BaseNode::default(),
                id: Identifier {
                    base: BaseNode::default(),
                    name: "mean".to_string(),
                },
                init: Expression::Object(ObjectExpression {
                    base: BaseNode::default(),
                    properties: vec![
                        Property {
                            base: BaseNode::default(),
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode::default(),
                                name: "want".to_string(),
                            }),
                            value: Some(Expression::Int(IntegerLiteral {
                                base: Default::default(),
                                value: 0,
                            })),
                        },
                        Property {
                            base: BaseNode::default(),
                            key: PropertyKey::Identifier(Identifier {
                                base: BaseNode::default(),
                                name: "got".to_string(),
                            }),
                            value: Some(Expression::Int(IntegerLiteral {
                                base: Default::default(),
                                value: 0,
                            })),
                        },
                    ],
                }),
            },
        };
        let serialized = serde_json::to_string(&n).unwrap();
        assert_eq!(serialized, r#"{"type":"TestStatement","assignment":{"type":"VariableAssignment","id":{"type":"Identifier","name":"mean"},"init":{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"Identifier","name":"want"},"value":{"type":"IntegerLiteral","value":"0"}},{"type":"Property","key":{"type":"Identifier","name":"got"},"value":{"type":"IntegerLiteral","value":"0"}}]}}}"#)
    }
    /*
    {
        name: "qualified option statement",
        node: &ast.OptionStatement{
            Assignment: &ast.MemberAssignment{
                Member: &ast.MemberExpression{
                    Object: &ast.Identifier{
                        Name: "alert",
                    },
                    Property: &ast.Identifier{
                        Name: "state",
                    },
                },
                Init: &ast.StringLiteral{
                    Value: "Warning",
                },
            },
        },
        want: `{"type":"OptionStatement","assignment":{"type":"MemberAssignment","member":{"type":"MemberExpression","object":{"type":"Identifier","name":"alert"},"property":{"type":"Identifier","name":"state"}},"init":{"type":"StringLiteral","value":"Warning"}}}`,
    },
	*/
    #[test]
    fn test_json_qualified_option_statement() {
        let n = OptionStatement {
            base: BaseNode::default(),
            assignment: Assignment::Member(MemberAssignment {
                base: BaseNode::default(),
                member: MemberExpression {
                    base: BaseNode::default(),
                    object: Expression::Identifier(Identifier {
                        base: BaseNode::default(),
                        name: "alert".to_string(),
                    }),
                    property: PropertyKey::Identifier(Identifier {
                        base: BaseNode::default(),
                        name: "state".to_string(),
                    }
                    ),
                },
                init: Expression::String(StringLiteral {
                    base: Default::default(),
                    value: "Warning".to_string(),
                }),
            }),
        };
        let serialized = serde_json::to_string(&n).unwrap();
        assert_eq!(serialized, r#"{"type":"OptionStatement","assignment":{"type":"MemberAssignment","member":{"type":"MemberExpression","object":{"type":"Identifier","name":"alert"},"property":{"type":"Identifier","name":"state"}},"init":{"type":"StringLiteral","value":"Warning"}}}"#)
    }
    /*
    {
        name: "variable assignment",
        node: &ast.VariableAssignment{
            ID:   &ast.Identifier{Name: "a"},
            Init: &ast.StringLiteral{Value: "hello"},
        },
        want: `{"type":"VariableAssignment","id":{"type":"Identifier","name":"a"},"init":{"type":"StringLiteral","value":"hello"}}`,
    },
	*/
    #[test]
    fn test_json_variable_assignment() {
        let n = VariableAssignment {
            base: BaseNode::default(),
            id: Identifier {
                base: BaseNode::default(),
                name: "a".to_string(),
            },
            init: Expression::String(StringLiteral {
                base: BaseNode::default(),
                value: "hello".to_string(),
            }),
        };
        let serialized = serde_json::to_string(&n).unwrap();
        assert_eq!(serialized, r#"{"type":"VariableAssignment","id":{"type":"Identifier","name":"a"},"init":{"type":"StringLiteral","value":"hello"}}"#)
    }
    /*
    {
        name: "call expression",
        node: &ast.CallExpression{
            Callee:    &ast.Identifier{Name: "a"},
            Arguments: []ast.Expression{&ast.StringLiteral{Value: "hello"}},
        },
        want: `{"type":"CallExpression","callee":{"type":"Identifier","name":"a"},"arguments":[{"type":"StringLiteral","value":"hello"}]}`,
    },
	*/
    #[test]
    fn test_json_call_expression() {
        let n = CallExpression {
            base: BaseNode::default(),
            callee: Expression::Identifier(Identifier {
                base: BaseNode::default(),
                name: "a".to_string(),
            }),
            arguments: vec![
                Expression::String(StringLiteral {
                    base: BaseNode::default(),
                    value: "hello".to_string(),
                }),
            ],
        };
        let serialized = serde_json::to_string(&n).unwrap();
        assert_eq!(serialized, r#"{"type":"CallExpression","callee":{"type":"Identifier","name":"a"},"arguments":[{"type":"StringLiteral","value":"hello"}]}"#)
    }
    /*
    {
        name: "pipe expression",
        node: &ast.PipeExpression{
            Argument: &ast.Identifier{Name: "a"},
            Call: &ast.CallExpression{
                Callee:    &ast.Identifier{Name: "a"},
                Arguments: []ast.Expression{&ast.StringLiteral{Value: "hello"}},
            },
        },
        want: `{"type":"PipeExpression","argument":{"type":"Identifier","name":"a"},"call":{"type":"CallExpression","callee":{"type":"Identifier","name":"a"},"arguments":[{"type":"StringLiteral","value":"hello"}]}}`,
    },
	*/
    #[test]
    fn test_json_pipe_expression() {
        let n = PipeExpression {
            base: BaseNode::default(),
            argument: Expression::Identifier(Identifier {
                base: BaseNode::default(),
                name: "a".to_string(),
            }),
            call: CallExpression {
                base: BaseNode::default(),
                callee: Expression::Identifier(Identifier {
                    base: BaseNode::default(),
                    name: "a".to_string(),
                }),
                arguments: vec![
                    Expression::String(StringLiteral {
                        base: BaseNode::default(),
                        value: "hello".to_string(),
                    }),
                ],
            },
        };
        let serialized = serde_json::to_string(&n).unwrap();
        assert_eq!(serialized, r#"{"type":"PipeExpression","argument":{"type":"Identifier","name":"a"},"call":{"type":"CallExpression","callee":{"type":"Identifier","name":"a"},"arguments":[{"type":"StringLiteral","value":"hello"}]}}"#)
    }
    /*
    {
        name: "member expression with identifier",
        node: &ast.MemberExpression{
            Object:   &ast.Identifier{Name: "a"},
            Property: &ast.Identifier{Name: "b"},
        },
        want: `{"type":"MemberExpression","object":{"type":"Identifier","name":"a"},"property":{"type":"Identifier","name":"b"}}`,
    },
	*/
    #[test]
    fn test_json_member_expression_with_identifier() {
        let n = MemberExpression {
            base: BaseNode::default(),
            object: Expression::Identifier(Identifier {
                base: BaseNode::default(),
                name: "a".to_string(),
            }),
            property: PropertyKey::Identifier(Identifier {
                base: BaseNode::default(),
                name: "b".to_string(),
            }),
        };
        let serialized = serde_json::to_string(&n).unwrap();
        assert_eq!(serialized, r#"{"type":"MemberExpression","object":{"type":"Identifier","name":"a"},"property":{"type":"Identifier","name":"b"}}"#)
    }
    /*
    {
        name: "member expression with string literal",
        node: &ast.MemberExpression{
            Object:   &ast.Identifier{Name: "a"},
            Property: &ast.StringLiteral{Value: "b"},
        },
        want: `{"type":"MemberExpression","object":{"type":"Identifier","name":"a"},"property":{"type":"StringLiteral","value":"b"}}`,
    },
	*/
    #[test]
    fn test_json_member_expression_with_string_literal() {
        let n = MemberExpression {
            base: BaseNode::default(),
            object: Expression::Identifier(Identifier {
                base: BaseNode::default(),
                name: "a".to_string(),
            }),
            property: PropertyKey::StringLiteral(StringLiteral {
                base: BaseNode::default(),
                value: "b".to_string(),
            }),
        };
        let serialized = serde_json::to_string(&n).unwrap();
        assert_eq!(serialized, r#"{"type":"MemberExpression","object":{"type":"Identifier","name":"a"},"property":{"type":"StringLiteral","value":"b"}}"#)
    }
    /*
    {
        name: "index expression",
        node: &ast.IndexExpression{
            Array: &ast.Identifier{Name: "a"},
            Index: &ast.IntegerLiteral{Value: 3},
        },
        want: `{"type":"IndexExpression","array":{"type":"Identifier","name":"a"},"index":{"type":"IntegerLiteral","value":"3"}}`,
    },
	*/
    #[test]
    fn test_json_index_expression() {
        let n = IndexExpression {
            base: BaseNode::default(),
            array: Expression::Identifier(Identifier {
                base: BaseNode::default(),
                name: "a".to_string(),
            }),
            index: Expression::Int(IntegerLiteral {
                base: BaseNode::default(),
                value: 3,
            }),
        };
        let serialized = serde_json::to_string(&n).unwrap();
        assert_eq!(serialized, r#"{"type":"IndexExpression","array":{"type":"Identifier","name":"a"},"index":{"type":"IntegerLiteral","value":"3"}}"#)
    }
    /*
    {
        name: "arrow function expression",
        node: &ast.FunctionExpression{
            Params: []*ast.Property{{Key: &ast.Identifier{Name: "a"}}},
            Body:   &ast.StringLiteral{Value: "hello"},
        },
        want: `{"type":"FunctionExpression","params":[{"type":"Property","key":{"type":"Identifier","name":"a"},"value":null}],"body":{"type":"StringLiteral","value":"hello"}}`,
    },
	*/
    #[test]
    fn test_json_arrow_function_expression() {
        let n = FunctionExpression {
            base: BaseNode::default(),
            params: vec![
                Property {
                    base: BaseNode::default(),
                    key: PropertyKey::Identifier(Identifier {
                        base: BaseNode::default(),
                        name: "a".to_string(),
                    }),
                    value: None,
                }
            ],
            body: FunctionBody::Expr(Expression::String(StringLiteral {
                base: BaseNode::default(),
                value: "hello".to_string(),
            }
            )),
        };
        let serialized = serde_json::to_string(&n).unwrap();
        assert_eq!(serialized, r#"{"type":"FunctionExpression","params":[{"type":"Property","key":{"type":"Identifier","name":"a"},"value":null}],"body":{"type":"StringLiteral","value":"hello"}}"#)
    }
    /*
    {
        name: "binary expression",
        node: &ast.BinaryExpression{
            Operator: ast.AdditionOperator,
            Left:     &ast.StringLiteral{Value: "hello"},
            Right:    &ast.StringLiteral{Value: "world"},
        },
        want: `{"type":"BinaryExpression","operator":"+","left":{"type":"StringLiteral","value":"hello"},"right":{"type":"StringLiteral","value":"world"}}`,
    },
	*/
    #[test]
    fn test_json_binary_expression() {
        let n = BinaryExpression {
            base: BaseNode::default(),
            operator: OperatorKind::AdditionOperator,
            left: Expression::String(StringLiteral {
                base: BaseNode::default(),
                value: "hello".to_string(),
            }),
            right: Expression::String(StringLiteral {
                base: BaseNode::default(),
                value: "world".to_string(),
            }),
        };
        let serialized = serde_json::to_string(&n).unwrap();
        assert_eq!(serialized, r#"{"type":"BinaryExpression","operator":"+","left":{"type":"StringLiteral","value":"hello"},"right":{"type":"StringLiteral","value":"world"}}"#)
    }
    /*
    {
        name: "unary expression",
        node: &ast.UnaryExpression{
            Operator: ast.NotOperator,
            Argument: &ast.BooleanLiteral{Value: true},
        },
        want: `{"type":"UnaryExpression","operator":"not","argument":{"type":"BooleanLiteral","value":true}}`,
    },
	*/
    #[test]
    fn test_json_unary_expression() {
        let n = UnaryExpression {
            base: BaseNode::default(),
            operator: OperatorKind::NotOperator,
            argument: Expression::Bool(BooleanLiteral {
                base: BaseNode::default(),
                value: true,
            }),
        };
        let serialized = serde_json::to_string(&n).unwrap();
        assert_eq!(serialized, r#"{"type":"UnaryExpression","operator":"not","argument":{"type":"BooleanLiteral","value":true}}"#)
    }
    /*
    {
        name: "logical expression",
        node: &ast.LogicalExpression{
            Operator: ast.OrOperator,
            Left:     &ast.BooleanLiteral{Value: false},
            Right:    &ast.BooleanLiteral{Value: true},
        },
        want: `{"type":"LogicalExpression","operator":"or","left":{"type":"BooleanLiteral","value":false},"right":{"type":"BooleanLiteral","value":true}}`,
    },
	*/
    #[test]
    fn test_json_logical_expression() {
        let n = LogicalExpression {
            base: BaseNode::default(),
            operator: LogicalOperatorKind::OrOperator,
            left: Expression::Bool(BooleanLiteral {
                base: BaseNode::default(),
                value: false,
            }),
            right: Expression::Bool(BooleanLiteral {
                base: BaseNode::default(),
                value: true,
            }),
        };
        let serialized = serde_json::to_string(&n).unwrap();
        assert_eq!(serialized, r#"{"type":"LogicalExpression","operator":"or","left":{"type":"BooleanLiteral","value":false},"right":{"type":"BooleanLiteral","value":true}}"#)
    }
    /*
    {
        name: "array expression",
        node: &ast.ArrayExpression{
            Elements: []ast.Expression{&ast.StringLiteral{Value: "hello"}},
        },
        want: `{"type":"ArrayExpression","elements":[{"type":"StringLiteral","value":"hello"}]}`,
    },
	*/
    #[test]
    fn test_json_array_expression() {
        let n = ArrayExpression {
            base: BaseNode::default(),
            elements: vec![
                Expression::String(StringLiteral {
                    base: BaseNode::default(),
                    value: "hello".to_string(),
                }),
            ],
        };
        let serialized = serde_json::to_string(&n).unwrap();
        assert_eq!(serialized, r#"{"type":"ArrayExpression","elements":[{"type":"StringLiteral","value":"hello"}]}"#)
    }
    /*
    {
        name: "object expression",
        node: &ast.ObjectExpression{
            Properties: []*ast.Property{{
                Key:   &ast.Identifier{Name: "a"},
                Value: &ast.StringLiteral{Value: "hello"},
            }},
        },
        want: `{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"Identifier","name":"a"},"value":{"type":"StringLiteral","value":"hello"}}]}`,
    },
	*/
    #[test]
    fn test_json_object_expression() {
        let n = ObjectExpression {
            base: BaseNode::default(),
            properties: vec![
                Property {
                    base: BaseNode::default(),
                    key: PropertyKey::Identifier(Identifier {
                        base: BaseNode::default(),
                        name: "a".to_string(),
                    }),
                    value: Some(Expression::String(StringLiteral {
                        base: BaseNode::default(),
                        value: "hello".to_string(),
                    })),
                }
            ],
        };
        let serialized = serde_json::to_string(&n).unwrap();
        assert_eq!(serialized, r#"{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"Identifier","name":"a"},"value":{"type":"StringLiteral","value":"hello"}}]}"#)
    }
    /*
    {
        name: "object expression with string literal key",
        node: &ast.ObjectExpression{
            Properties: []*ast.Property{{
                Key:   &ast.StringLiteral{Value: "a"},
                Value: &ast.StringLiteral{Value: "hello"},
            }},
        },
        want: `{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"StringLiteral","value":"a"},"value":{"type":"StringLiteral","value":"hello"}}]}`,
    },
	*/
    #[test]
    fn test_json_object_expression_with_string_literal_key() {
        let n = ObjectExpression {
            base: BaseNode::default(),
            properties: vec![
                Property {
                    base: BaseNode::default(),
                    key: PropertyKey::StringLiteral(StringLiteral {
                        base: BaseNode::default(),
                        value: "a".to_string(),
                    }),
                    value: Some(Expression::String(StringLiteral {
                        base: BaseNode::default(),
                        value: "hello".to_string(),
                    })),
                }
            ],
        };
        let serialized = serde_json::to_string(&n).unwrap();
        assert_eq!(serialized, r#"{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"StringLiteral","value":"a"},"value":{"type":"StringLiteral","value":"hello"}}]}"#)
    }
    /*
		{
			name: "object expression implicit keys",
			node: &ast.ObjectExpression{
				Properties: []*ast.Property{{
					Key: &ast.Identifier{Name: "a"},
				}},
			},
			want: `{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"Identifier","name":"a"},"value":null}]}`,
		},
	*/
    #[test]
    fn test_json_object_expression_implicit_keys() {
        let n = ObjectExpression {
            base: BaseNode::default(),
            properties: vec![
                Property {
                    base: BaseNode::default(),
                    key: PropertyKey::Identifier(Identifier {
                        base: BaseNode::default(),
                        name: "a".to_string(),
                    }),
                    value: None,
                }
            ],
        };
        let serialized = serde_json::to_string(&n).unwrap();
        assert_eq!(serialized, r#"{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"Identifier","name":"a"},"value":null}]}"#)
    }
    /*
    {
        name: "conditional expression",
        node: &ast.ConditionalExpression{
            Test:       &ast.BooleanLiteral{Value: true},
            Alternate:  &ast.StringLiteral{Value: "false"},
            Consequent: &ast.StringLiteral{Value: "true"},
        },
        want: `{"type":"ConditionalExpression","test":{"type":"BooleanLiteral","value":true},"consequent":{"type":"StringLiteral","value":"true"},"alternate":{"type":"StringLiteral","value":"false"}}`,
    },
	*/
    #[test]
    fn test_json_conditional_expression() {
        let n = ConditionalExpression {
            base: BaseNode::default(),
            test: Expression::Bool(BooleanLiteral {
                base: BaseNode::default(),
                value: true,
            }),
            alternate: Expression::String(StringLiteral {
                base: BaseNode::default(),
                value: "false".to_string(),
            }),
            consequent: Expression::String(StringLiteral {
                base: BaseNode::default(),
                value: "true".to_string(),
            }),
        };
        let serialized = serde_json::to_string(&n).unwrap();
        assert_eq!(serialized, r#"{"type":"ConditionalExpression","test":{"type":"BooleanLiteral","value":true},"consequent":{"type":"StringLiteral","value":"true"},"alternate":{"type":"StringLiteral","value":"false"}}"#)
    }
    /*
    {
        name: "property",
        node: &ast.Property{
            Key:   &ast.Identifier{Name: "a"},
            Value: &ast.StringLiteral{Value: "hello"},
        },
        want: `{"type":"Property","key":{"type":"Identifier","name":"a"},"value":{"type":"StringLiteral","value":"hello"}}`,
    },
	*/
    #[test]
    fn test_json_property() {
        let n = Property {
            base: BaseNode::default(),
            key: PropertyKey::Identifier(Identifier {
                base: BaseNode::default(),
                name: "a".to_string(),
            }),
            value: Some(Expression::String(StringLiteral {
                base: BaseNode::default(),
                value: "hello".to_string(),
            })),
        };
        let serialized = serde_json::to_string(&n).unwrap();
        assert_eq!(serialized, r#"{"type":"Property","key":{"type":"Identifier","name":"a"},"value":{"type":"StringLiteral","value":"hello"}}"#)
    }
    /*
    {
        name: "identifier",
        node: &ast.Identifier{
            Name: "a",
        },
        want: `{"type":"Identifier","name":"a"}`,
    },
	*/
    #[test]
    fn test_json_identifier() {
        let n = Identifier {
            base: BaseNode::default(),
            name: "a".to_string(),
        };
        let serialized = serde_json::to_string(&n).unwrap();
        assert_eq!(serialized, r#"{"type":"Identifier","name":"a"}"#)
    }
    /*
    {
        name: "string literal",
        node: &ast.StringLiteral{
            Value: "hello",
        },
        want: `{"type":"StringLiteral","value":"hello"}`,
    },
	*/
    #[test]
    fn test_json_string_literal() {
        let n = StringLiteral {
            base: BaseNode::default(),
            value: "hello".to_string(),
        };
        let serialized = serde_json::to_string(&n).unwrap();
        assert_eq!(serialized, r#"{"type":"StringLiteral","value":"hello"}"#)
    }
    /*
    {
        name: "boolean literal",
        node: &ast.BooleanLiteral{
            Value: true,
        },
        want: `{"type":"BooleanLiteral","value":true}`,
    },
	*/
    #[test]
    fn test_json_boolean_literal() {
        let n = BooleanLiteral {
            base: BaseNode::default(),
            value: true,
        };
        let serialized = serde_json::to_string(&n).unwrap();
        assert_eq!(serialized, r#"{"type":"BooleanLiteral","value":true}"#)
    }
    /*
    {
        name: "float literal",
        node: &ast.FloatLiteral{
            Value: 42.1,
        },
        want: `{"type":"FloatLiteral","value":42.1}`,
    },
	*/
    #[test]
    fn test_json_float_literal() {
        let n = FloatLiteral {
            base: BaseNode::default(),
            value: 42.1,
        };
        let serialized = serde_json::to_string(&n).unwrap();
        assert_eq!(serialized, r#"{"type":"FloatLiteral","value":42.1}"#)
    }
    /*
    {
        name: "integer literal",
        node: &ast.IntegerLiteral{
            Value: math.MaxInt64,
        },
        want: `{"type":"IntegerLiteral","value":"9223372036854775807"}`,
    },
	*/
    #[test]
    fn test_json_integer_literal() {
        let n = IntegerLiteral {
            base: BaseNode::default(),
            value: 9223372036854775807,
        };
        let serialized = serde_json::to_string(&n).unwrap();
        assert_eq!(serialized, r#"{"type":"IntegerLiteral","value":"9223372036854775807"}"#)
    }
    /*
    {
        name: "unsigned integer literal",
        node: &ast.UnsignedIntegerLiteral{
            Value: math.MaxUint64,
        },
        want: `{"type":"UnsignedIntegerLiteral","value":"18446744073709551615"}`,
    },
	*/
    #[test]
    fn test_json_unsigned_integer_literal() {
        let n = UnsignedIntegerLiteral {
            base: BaseNode::default(),
            value: 18446744073709551615,
        };
        let serialized = serde_json::to_string(&n).unwrap();
        assert_eq!(serialized, r#"{"type":"UnsignedIntegerLiteral","value":"18446744073709551615"}"#)
    }
    /*
    {
        name: "regexp literal",
        node: &ast.RegexpLiteral{
            Value: regexp.MustCompile(`.*`),
        },
        want: `{"type":"RegexpLiteral","value":".*"}`,
    },
    */
    #[test]
    fn test_json_regexp_literal() {
        let n = RegexpLiteral {
            base: BaseNode::default(),
            value: ".*".to_string(),
        };
        let serialized = serde_json::to_string(&n).unwrap();
        assert_eq!(serialized, r#"{"type":"RegexpLiteral","value":".*"}"#)
    }
    /*
    {
        name: "duration literal",
        node: &ast.DurationLiteral{
            Values: []ast.Duration{
                {
                    Magnitude: 1,
                    Unit:      "h",
                },
                {
                    Magnitude: 1,
                    Unit:      "h",
                },
            },
        },
        want: `{"type":"DurationLiteral","values":[{"magnitude":1,"unit":"h"},{"magnitude":1,"unit":"h"}]}`,
    },
    */
    #[test]
    fn test_json_duration_literal() {
        let n = DurationLiteral {
            base: BaseNode::default(),
            values: vec![
                Duration {
                    magnitude: 1,
                    unit: "h".to_string(),
                },
                Duration {
                    magnitude: 1,
                    unit: "h".to_string(),
                },
            ],
        };
        let serialized = serde_json::to_string(&n).unwrap();
        assert_eq!(serialized, r#"{"type":"DurationLiteral","values":[{"magnitude":1,"unit":"h"},{"magnitude":1,"unit":"h"}]}"#)
    }
    /*
    {
        name: "datetime literal",
        node: &ast.DateTimeLiteral{
            Value: time.Date(2017, 8, 8, 8, 8, 8, 8, time.UTC),
        },
        want: `{"type":"DateTimeLiteral","value":"2017-08-08T08:08:08.000000008Z"}`,
    },
    */
    #[test]
    fn test_json_datetime_literal() {
        let n = DateTimeLiteral {
            base: BaseNode::default(),
            value: Utc.ymd_opt(2017, 8, 8).and_hms_nano_opt(8, 8, 8, 8).unwrap(),
        };
        let serialized = serde_json::to_string(&n).unwrap();
        assert_eq!(serialized, r#"{"type":"DateTimeLiteral","value":"2017-08-08T08:08:08.000000008Z"}"#)
    }
}
