use super::*;
use crate::ast::Expression::Integer;
use crate::ast::Statement::Variable;
use crate::ast::{BaseNode, File, Identifier, IntegerLit, Position, VariableAssgn};
use crate::parser::parse_string;

#[test]
fn test_object_check() {
    let file = parse_string("object_test", "a = 1\nb = {c: 2, a}");
    let got = check(walk::Node::File(&file));
    let want = vec![Error {
        location: SourceLocation {
            file: Some(String::from("object_test")),
            start: Position { line: 2, column: 5 },
            end: Position {
                line: 2,
                column: 14,
            },
            source: Some(String::from("{c: 2, a}")),
        },
        message: String::from("cannot mix implicit and explicit properties"),
    }];
    assert_eq!(want, got);
}

#[test]
fn test_bad_expr() {
    let file = parse_string("bad_expr_test", "a = 1\nb = \nc=2");
    let got = check(walk::Node::File(&file));
    let want = vec![Error {
        location: SourceLocation {
            file: Some(String::from("bad_expr_test")),
            start: Position { line: 3, column: 2 },
            end: Position { line: 3, column: 3 },
            source: Some(String::from("=")),
        },
        message: String::from("invalid statement: ="),
    }];
    assert_eq!(want, got);
}

#[test]
fn test_check_ok() {
    let file = parse_string("test_ok", "a = 1\nb=2");
    let got = check(walk::Node::File(&file));
    assert_eq!(got.len(), 0);
}

#[test]
fn test_check_collect_existing_error() {
    let file = File {
        base: BaseNode {
            location: SourceLocation {
                file: Some(String::from("test_check_collect_existing_error")),
                start: Position { line: 1, column: 1 },
                end: Position { line: 3, column: 6 },
                source: Some(String::from("a = 1\nb=2\nc=a+b")),
            },
            errors: vec![String::from("error 1")],
        },
        name: String::from("test_check_collect_existing_error"),
        metadata: String::new(),
        package: None,
        imports: vec![],
        body: vec![Variable(Box::new(VariableAssgn {
            base: BaseNode {
                location: SourceLocation {
                    file: Some(String::from("test_check_collect_existing_error")),
                    start: Position { line: 1, column: 1 },
                    end: Position { line: 1, column: 6 },
                    source: Some(String::from("a = 1")),
                },
                errors: vec![],
            },
            id: Identifier {
                base: BaseNode {
                    location: SourceLocation {
                        file: Some(String::from("test_check_collect_existing_error")),
                        start: Position { line: 1, column: 1 },
                        end: Position { line: 1, column: 2 },
                        source: Some(String::from("a")),
                    },
                    errors: vec![],
                },
                name: String::from("a"),
            },
            init: Integer(IntegerLit {
                base: BaseNode {
                    location: SourceLocation {
                        file: Some(String::from("test_check_collect_existing_error")),
                        start: Position { line: 1, column: 5 },
                        end: Position { line: 1, column: 6 },
                        source: Some(String::from("1")),
                    },
                    errors: vec![String::from("error 2"), String::from("error 3")],
                },
                value: 1,
            }),
        }))],
    };
    let got = check(walk::Node::File(&file));
    assert_eq!(3, got.len());
    for (i, err) in got.iter().enumerate() {
        assert_eq!(err.message, format!("error {}", i + 1));
    }
}
