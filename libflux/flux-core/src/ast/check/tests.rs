use super::*;
use crate::{
    ast::{
        BaseNode, Expression::Integer, File, Identifier, IntegerLit, Position, SourceLocation,
        Statement::Variable, VariableAssgn,
    },
    parser::parse_string,
};

#[test]
fn test_object_check() {
    let file = parse_string("object_test".to_string(), "a = 1\nb = {c: 2, a}");
    let got = check(walk::Node::File(&file));
    let want = Err(Errors::from_iter(vec![located(
        SourceLocation {
            file: Some("object_test".into()),
            start: Position { line: 2, column: 5 },
            end: Position {
                line: 2,
                column: 14,
            },
        },
        ErrorKind {
            message: String::from("cannot mix implicit and explicit properties"),
        },
    )]));
    assert_eq!(want, got);
}

#[test]
fn test_bad_stmt() {
    let file = parse_string("bad_stmt_test".to_string(), "a = 1\nb = \nc=2");
    let got = check(walk::Node::File(&file));
    let want = Err(Errors::from_iter(vec![located(
        SourceLocation {
            file: Some("bad_stmt_test".into()),
            start: Position { line: 3, column: 2 },
            end: Position { line: 3, column: 3 },
        },
        ErrorKind {
            message: String::from("invalid statement: ="),
        },
    )]));
    assert_eq!(want, got);
}

#[test]
fn test_bad_expr() {
    let file = parse_string("bad_expr_test".to_string(), "a = 3 + / 10");
    let got = check(walk::Node::File(&file));
    let want = Err(Errors::from_iter(vec![located(
        SourceLocation {
            file: Some("bad_expr_test".into()),
            start: Position { line: 1, column: 9 },
            end: Position {
                line: 1,
                column: 10,
            },
        },
        ErrorKind {
            message: String::from("invalid expression: invalid token for primary expression: DIV"),
        },
    )]));
    assert_eq!(want, got);
}

#[test]
fn test_check_ok() {
    let file = parse_string("test_ok".to_string(), "a = 1\nb=2");
    check(walk::Node::File(&file)).unwrap();
}

#[test]
fn test_check_collect_existing_error() {
    let file = File {
        base: BaseNode {
            location: SourceLocation {
                file: Some("test_check_collect_existing_error".into()),
                start: Position { line: 1, column: 1 },
                end: Position { line: 3, column: 6 },
            },
            errors: vec![String::from("error 1")],
            ..BaseNode::default()
        },
        name: String::from("test_check_collect_existing_error"),
        metadata: String::new(),
        package: None,
        imports: vec![],
        body: vec![Variable(Box::new(VariableAssgn {
            base: BaseNode {
                location: SourceLocation {
                    file: Some("test_check_collect_existing_error".into()),
                    start: Position { line: 1, column: 1 },
                    end: Position { line: 1, column: 6 },
                },
                ..BaseNode::default()
            },
            id: Identifier {
                base: BaseNode {
                    location: SourceLocation {
                        file: Some("test_check_collect_existing_error".into()),
                        start: Position { line: 1, column: 1 },
                        end: Position { line: 1, column: 2 },
                    },
                    ..BaseNode::default()
                },
                name: String::from("a"),
            },
            init: Integer(IntegerLit {
                base: BaseNode {
                    location: SourceLocation {
                        file: Some("test_check_collect_existing_error".into()),
                        start: Position { line: 1, column: 5 },
                        end: Position { line: 1, column: 6 },
                    },
                    errors: vec![String::from("error 2"), String::from("error 3")],
                    ..BaseNode::default()
                },
                value: 1,
            }),
        }))],
        eof: vec![],
    };
    let got = check(walk::Node::File(&file)).unwrap_err();
    assert_eq!(3, got.len());
    for (i, err) in got.iter().enumerate() {
        assert_eq!(err.error.message, format!("error {}", i + 1));
    }
}
