// NOTE: These test cases directly match ast/json_test.go.
// Every test is preceded by the correspondent test case in golang.
use super::*;
use crate::parser::parse_string;

fn test_walk(source: &str, want: Vec<&str>) {
    let file = parse_string("test_walk", source);
    let mut nodes = Vec::new();
    walk(
        &create_visitor(&mut |n| nodes.push(format!("{}", n))),
        Node::File(&file),
    );
    assert_eq!(want, nodes);
}

#[test]
fn test_file() {
    test_walk("", vec!["File"])
}
#[test]
fn test_package_clause() {
    test_walk("package a", vec!["File", "PackageClause", "Identifier"])
}
#[test]
fn test_import_declaration() {
    test_walk(
        "import \"a\"",
        vec!["File", "ImportDeclaration", "StringLit"],
    )
}
#[test]
fn test_ident() {
    test_walk("a", vec!["File", "ExprStmt", "Identifier"])
}
#[test]
fn test_array_expr() {
    test_walk(
        "[1,2,3]",
        vec![
            "File",
            "ExprStmt",
            "ArrayExpr",
            "IntegerLit",
            "IntegerLit",
            "IntegerLit",
        ],
    )
}
#[test]
fn test_function_expr() {
    test_walk(
        "() => 1",
        vec!["File", "ExprStmt", "FunctionExpr", "IntegerLit"],
    )
}
#[test]
fn test_function_with_args() {
    test_walk(
        "(a=1) => a",
        vec![
            "File",
            "ExprStmt",
            "FunctionExpr",
            "Property",
            "Identifier",
            "IntegerLit",
            "Identifier",
        ],
    )
}
#[test]
fn test_logical_expr() {
    test_walk(
        "true or false",
        vec![
            "File",
            "ExprStmt",
            "LogicalExpr",
            "Identifier",
            "Identifier",
        ],
    )
}
#[test]
fn test_object_expr() {
    test_walk(
        "{a:1,\"b\":false}",
        vec![
            "File",
            "ExprStmt",
            "ObjectExpr",
            "Property",
            "Identifier",
            "IntegerLit",
            "Property",
            "StringLit",
            "Identifier",
        ],
    )
}
#[test]
fn test_member_expr() {
    test_walk(
        "a.b",
        vec!["File", "ExprStmt", "MemberExpr", "Identifier", "Identifier"],
    )
}
#[test]
fn test_index_expr() {
    test_walk(
        "a[b]",
        vec!["File", "ExprStmt", "IndexExpr", "Identifier", "Identifier"],
    )
}
#[test]
fn test_binary_expr() {
    test_walk(
        "a+b",
        vec!["File", "ExprStmt", "BinaryExpr", "Identifier", "Identifier"],
    )
}
#[test]
fn test_unary_expr() {
    test_walk("-b", vec!["File", "ExprStmt", "UnaryExpr", "Identifier"])
}
#[test]
fn test_pipe_expr() {
    test_walk(
        "a|>b()",
        vec![
            "File",
            "ExprStmt",
            "PipeExpr",
            "Identifier",
            "CallExpr",
            "Identifier",
        ],
    )
}
#[test]
fn test_call_expr() {
    test_walk(
        "b(a:1)",
        vec![
            "File",
            "ExprStmt",
            "CallExpr",
            "Identifier",
            "ObjectExpr",
            "Property",
            "Identifier",
            "IntegerLit",
        ],
    )
}
#[test]
fn test_conditional_expr() {
    test_walk(
        "if x then y else z",
        vec![
            "File",
            "ExprStmt",
            "ConditionalExpr",
            "Identifier",
            "Identifier",
            "Identifier",
        ],
    )
}
#[test]
fn test_string_expr() {
    test_walk(
        "\"hello ${world}\"",
        vec![
            "File",
            "ExprStmt",
            "StringExpr",
            "TextPart",
            "InterpolatedPart",
            "Identifier",
        ],
    )
}
#[test]
fn test_paren_expr() {
    test_walk(
        "(a + b)",
        vec![
            "File",
            "ExprStmt",
            "ParenExpr",
            "BinaryExpr",
            "Identifier",
            "Identifier",
        ],
    )
}
#[test]
fn test_integer_lit() {
    test_walk("1", vec!["File", "ExprStmt", "IntegerLit"])
}
#[test]
fn test_float_lit() {
    test_walk("1.0", vec!["File", "ExprStmt", "FloatLit"])
}
#[test]
fn test_string_lit() {
    test_walk("\"a\"", vec!["File", "ExprStmt", "StringLit"])
}
#[test]
fn test_duration_lit() {
    test_walk("1m", vec!["File", "ExprStmt", "DurationLit"])
}
#[test]
fn test_datetime_lit() {
    test_walk(
        "2019-01-01T00:00:00Z",
        vec!["File", "ExprStmt", "DateTimeLit"],
    )
}
#[test]
fn test_regexp_lit() {
    test_walk("/./", vec!["File", "ExprStmt", "RegexpLit"])
}
#[test]
fn test_pipe_lit() {
    test_walk(
        "(a=<-)=>a",
        vec![
            "File",
            "ExprStmt",
            "FunctionExpr",
            "Property",
            "Identifier",
            "PipeLit",
            "Identifier",
        ],
    )
}

#[test]
fn test_option_stmt() {
    test_walk(
        "option a = b",
        vec![
            "File",
            "OptionStmt",
            "VariableAssgn",
            "Identifier",
            "Identifier",
        ],
    )
}
#[test]
fn test_return_stmt() {
    test_walk(
        "() => {return 1}",
        vec![
            "File",
            "ExprStmt",
            "FunctionExpr",
            "Block",
            "ReturnStmt",
            "IntegerLit",
        ],
    )
}
#[test]
fn test_test_stmt() {
    test_walk(
        "test a = 1",
        vec![
            "File",
            "TestStmt",
            "VariableAssgn",
            "Identifier",
            "IntegerLit",
        ],
    )
}
#[test]
fn test_builtin_stmt() {
    test_walk("builtin a", vec!["File", "BuiltinStmt", "Identifier"])
}
#[test]
fn test_variable_assgn() {
    test_walk(
        "a = b",
        vec!["File", "VariableAssgn", "Identifier", "Identifier"],
    )
}
#[test]
fn test_member_assgn() {
    test_walk(
        "option a.b = c",
        vec![
            "File",
            "OptionStmt",
            "MemberAssgn",
            "MemberExpr",
            "Identifier",
            "Identifier",
            "Identifier",
        ],
    )
}
