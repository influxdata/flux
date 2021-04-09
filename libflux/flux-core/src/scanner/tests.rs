use super::*;

#[test]
fn test_scan() {
    let text = "from(bucket:\"foo\") |> range(start: -1m)";
    let mut s = Scanner::new(text);
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Ident,
            lit: String::from("from"),
            start_offset: 0,
            end_offset: 4,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 5 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::LParen,
            lit: String::from("("),
            start_offset: 4,
            end_offset: 5,
            start_pos: Position { line: 1, column: 5 },
            end_pos: Position { line: 1, column: 6 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Ident,
            lit: String::from("bucket"),
            start_offset: 5,
            end_offset: 11,
            start_pos: Position { line: 1, column: 6 },
            end_pos: Position {
                line: 1,
                column: 12
            },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Colon,
            lit: String::from(":"),
            start_offset: 11,
            end_offset: 12,
            start_pos: Position {
                line: 1,
                column: 12
            },
            end_pos: Position {
                line: 1,
                column: 13
            },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::String,
            lit: String::from("\"foo\""),
            start_offset: 12,
            end_offset: 17,
            start_pos: Position {
                line: 1,
                column: 13
            },
            end_pos: Position {
                line: 1,
                column: 18
            },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::RParen,
            lit: String::from(")"),
            start_offset: 17,
            end_offset: 18,
            start_pos: Position {
                line: 1,
                column: 18
            },
            end_pos: Position {
                line: 1,
                column: 19
            },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::PipeForward,
            lit: String::from("|>"),
            start_offset: 19,
            end_offset: 21,
            start_pos: Position {
                line: 1,
                column: 20
            },
            end_pos: Position {
                line: 1,
                column: 22
            },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Ident,
            lit: String::from("range"),
            start_offset: 22,
            end_offset: 27,
            start_pos: Position {
                line: 1,
                column: 23
            },
            end_pos: Position {
                line: 1,
                column: 28
            },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::LParen,
            lit: String::from("("),
            start_offset: 27,
            end_offset: 28,
            start_pos: Position {
                line: 1,
                column: 28
            },
            end_pos: Position {
                line: 1,
                column: 29
            },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Ident,
            lit: String::from("start"),
            start_offset: 28,
            end_offset: 33,
            start_pos: Position {
                line: 1,
                column: 29
            },
            end_pos: Position {
                line: 1,
                column: 34
            },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Colon,
            lit: String::from(":"),
            start_offset: 33,
            end_offset: 34,
            start_pos: Position {
                line: 1,
                column: 34
            },
            end_pos: Position {
                line: 1,
                column: 35
            },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Sub,
            lit: String::from("-"),
            start_offset: 35,
            end_offset: 36,
            start_pos: Position {
                line: 1,
                column: 36
            },
            end_pos: Position {
                line: 1,
                column: 37
            },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Duration,
            lit: String::from("1m"),
            start_offset: 36,
            end_offset: 38,
            start_pos: Position {
                line: 1,
                column: 37
            },
            end_pos: Position {
                line: 1,
                column: 39
            },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::RParen,
            lit: String::from(")"),
            start_offset: 38,
            end_offset: 39,
            start_pos: Position {
                line: 1,
                column: 39
            },
            end_pos: Position {
                line: 1,
                column: 40
            },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Eof,
            lit: String::from(""),
            start_offset: 39,
            end_offset: 39,
            start_pos: Position {
                line: 1,
                column: 40
            },
            end_pos: Position {
                line: 1,
                column: 40
            },
            comments: vec![],
        }
    );
}

#[test]
fn scan_invalid_unicode_single_quotes() {
    let text = "‛some string‛";
    let mut s = Scanner::new(text);
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Illegal,
            lit: String::from("‛"),
            start_offset: 0,
            end_offset: 3,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 4 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Ident,
            lit: String::from("some"),
            start_offset: 3,
            end_offset: 7,
            start_pos: Position { line: 1, column: 4 },
            end_pos: Position { line: 1, column: 8 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Ident,
            lit: String::from("string"),
            start_offset: 8,
            end_offset: 14,
            start_pos: Position { line: 1, column: 9 },
            end_pos: Position {
                line: 1,
                column: 15
            },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Illegal,
            lit: String::from("‛"),
            start_offset: 14,
            end_offset: 17,
            start_pos: Position {
                line: 1,
                column: 15
            },
            end_pos: Position {
                line: 1,
                column: 18
            },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Eof,
            lit: String::from(""),
            start_offset: 17,
            end_offset: 17,
            start_pos: Position {
                line: 1,
                column: 18
            },
            end_pos: Position {
                line: 1,
                column: 18
            },
            comments: vec![],
        }
    );
}

#[test]
fn scan_invalid_unicode_double_quotes() {
    let text = "“some string”";
    let mut s = Scanner::new(text);
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Illegal,
            lit: String::from("“"),
            start_offset: 0,
            end_offset: 3,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 4 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Ident,
            lit: String::from("some"),
            start_offset: 3,
            end_offset: 7,
            start_pos: Position { line: 1, column: 4 },
            end_pos: Position { line: 1, column: 8 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Ident,
            lit: String::from("string"),
            start_offset: 8,
            end_offset: 14,
            start_pos: Position { line: 1, column: 9 },
            end_pos: Position {
                line: 1,
                column: 15
            },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Illegal,
            lit: String::from("”"),
            start_offset: 14,
            end_offset: 17,
            start_pos: Position {
                line: 1,
                column: 15
            },
            end_pos: Position {
                line: 1,
                column: 18
            },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Eof,
            lit: String::from(""),
            start_offset: 17,
            end_offset: 17,
            start_pos: Position {
                line: 1,
                column: 18
            },
            end_pos: Position {
                line: 1,
                column: 18
            },
            comments: vec![],
        }
    );
}

#[test]
fn scan_invalid_unicode_register() {
    let text = "®some string®";
    let mut s = Scanner::new(text);
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Illegal,
            lit: String::from("®"),
            start_offset: 0,
            end_offset: 2,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 3 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Ident,
            lit: String::from("some"),
            start_offset: 2,
            end_offset: 6,
            start_pos: Position { line: 1, column: 3 },
            end_pos: Position { line: 1, column: 7 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Ident,
            lit: String::from("string"),
            start_offset: 7,
            end_offset: 13,
            start_pos: Position { line: 1, column: 8 },
            end_pos: Position {
                line: 1,
                column: 14
            },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Illegal,
            lit: String::from("®"),
            start_offset: 13,
            end_offset: 15,
            start_pos: Position {
                line: 1,
                column: 14
            },
            end_pos: Position {
                line: 1,
                column: 16
            },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Eof,
            lit: String::from(""),
            start_offset: 15,
            end_offset: 15,
            start_pos: Position {
                line: 1,
                column: 16
            },
            end_pos: Position {
                line: 1,
                column: 16
            },
            comments: vec![],
        }
    );
}

#[test]
fn test_scan_with_regex() {
    let text = "a + b =~ /.*[0-9]/ / 2";
    let mut s = Scanner::new(text);
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TokenType::Ident,
            lit: String::from("a"),
            start_offset: 0,
            end_offset: 1,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 2 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TokenType::Add,
            lit: String::from("+"),
            start_offset: 2,
            end_offset: 3,
            start_pos: Position { line: 1, column: 3 },
            end_pos: Position { line: 1, column: 4 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TokenType::Ident,
            lit: String::from("b"),
            start_offset: 4,
            end_offset: 5,
            start_pos: Position { line: 1, column: 5 },
            end_pos: Position { line: 1, column: 6 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TokenType::RegexEq,
            lit: String::from("=~"),
            start_offset: 6,
            end_offset: 8,
            start_pos: Position { line: 1, column: 7 },
            end_pos: Position { line: 1, column: 9 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TokenType::Regex,
            lit: String::from("/.*[0-9]/"),
            start_offset: 9,
            end_offset: 18,
            start_pos: Position {
                line: 1,
                column: 10
            },
            end_pos: Position {
                line: 1,
                column: 19
            },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TokenType::Div,
            lit: String::from("/"),
            start_offset: 19,
            end_offset: 20,
            start_pos: Position {
                line: 1,
                column: 20
            },
            end_pos: Position {
                line: 1,
                column: 21
            },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TokenType::Int,
            lit: String::from("2"),
            start_offset: 21,
            end_offset: 22,
            start_pos: Position {
                line: 1,
                column: 22
            },
            end_pos: Position {
                line: 1,
                column: 23
            },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TokenType::Eof,
            lit: String::from(""),
            start_offset: 22,
            end_offset: 22,
            start_pos: Position {
                line: 1,
                column: 23
            },
            end_pos: Position {
                line: 1,
                column: 23
            },
            comments: vec![],
        }
    );
}

#[test]
fn test_scan_string_expr_simple() {
    let text = r#""${a + b}""#;
    let mut s = Scanner::new(text);
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TokenType::Quote,
            lit: String::from("\""),
            start_offset: 0,
            end_offset: 1,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 2 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TokenType::StringExpr,
            lit: String::from("${"),
            start_offset: 1,
            end_offset: 3,
            start_pos: Position { line: 1, column: 2 },
            end_pos: Position { line: 1, column: 4 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TokenType::Text,
            lit: String::from("a + b}"),
            start_offset: 3,
            end_offset: 9,
            start_pos: Position { line: 1, column: 4 },
            end_pos: Position {
                line: 1,
                column: 10
            },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TokenType::Quote,
            lit: String::from("\""),
            start_offset: 9,
            end_offset: 10,
            start_pos: Position {
                line: 1,
                column: 10
            },
            end_pos: Position {
                line: 1,
                column: 11
            },
            comments: vec![],
        }
    );
}

#[test]
fn test_scan_string_expr_start_with_text() {
    let text = r#""a + b = ${a + b}""#;
    let mut s = Scanner::new(text);
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TokenType::Quote,
            lit: String::from("\""),
            start_offset: 0,
            end_offset: 1,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 2 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TokenType::Text,
            lit: String::from("a + b = "),
            start_offset: 1,
            end_offset: 9,
            start_pos: Position { line: 1, column: 2 },
            end_pos: Position {
                line: 1,
                column: 10
            },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TokenType::StringExpr,
            lit: String::from("${"),
            start_offset: 9,
            end_offset: 11,
            start_pos: Position {
                line: 1,
                column: 10
            },
            end_pos: Position {
                line: 1,
                column: 12
            },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TokenType::Text,
            lit: String::from("a + b}"),
            start_offset: 11,
            end_offset: 17,
            start_pos: Position {
                line: 1,
                column: 12
            },
            end_pos: Position {
                line: 1,
                column: 18
            },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TokenType::Quote,
            lit: String::from("\""),
            start_offset: 17,
            end_offset: 18,
            start_pos: Position {
                line: 1,
                column: 18
            },
            end_pos: Position {
                line: 1,
                column: 19
            },
            comments: vec![],
        }
    );
}

#[test]
fn test_scan_string_expr_multiple() {
    let text = r#""a + b = ${a + b} and a - b = ${a - b}""#;
    let mut s = Scanner::new(text);
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TokenType::Quote,
            lit: String::from("\""),
            start_offset: 0,
            end_offset: 1,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 2 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TokenType::Text,
            lit: String::from("a + b = "),
            start_offset: 1,
            end_offset: 9,
            start_pos: Position { line: 1, column: 2 },
            end_pos: Position {
                line: 1,
                column: 10
            },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TokenType::StringExpr,
            lit: String::from("${"),
            start_offset: 9,
            end_offset: 11,
            start_pos: Position {
                line: 1,
                column: 10
            },
            end_pos: Position {
                line: 1,
                column: 12
            },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TokenType::Text,
            lit: String::from("a + b} and a - b = "),
            start_offset: 11,
            end_offset: 30,
            start_pos: Position {
                line: 1,
                column: 12
            },
            end_pos: Position {
                line: 1,
                column: 31
            },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TokenType::StringExpr,
            lit: String::from("${"),
            start_offset: 30,
            end_offset: 32,
            start_pos: Position {
                line: 1,
                column: 31
            },
            end_pos: Position {
                line: 1,
                column: 33
            },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TokenType::Text,
            lit: String::from("a - b}"),
            start_offset: 32,
            end_offset: 38,
            start_pos: Position {
                line: 1,
                column: 33
            },
            end_pos: Position {
                line: 1,
                column: 39
            },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TokenType::Quote,
            lit: String::from("\""),
            start_offset: 38,
            end_offset: 39,
            start_pos: Position {
                line: 1,
                column: 39
            },
            end_pos: Position {
                line: 1,
                column: 40
            },
            comments: vec![],
        }
    );
}

#[test]
fn test_scan_string_expr_end_with_text() {
    let text = r#""a + b = ${a + b} and a - b = ?""#;
    let mut s = Scanner::new(text);
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TokenType::Quote,
            lit: String::from("\""),
            start_offset: 0,
            end_offset: 1,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 2 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TokenType::Text,
            lit: String::from("a + b = "),
            start_offset: 1,
            end_offset: 9,
            start_pos: Position { line: 1, column: 2 },
            end_pos: Position {
                line: 1,
                column: 10
            },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TokenType::StringExpr,
            lit: String::from("${"),
            start_offset: 9,
            end_offset: 11,
            start_pos: Position {
                line: 1,
                column: 10
            },
            end_pos: Position {
                line: 1,
                column: 12
            },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TokenType::Text,
            lit: String::from("a + b} and a - b = ?"),
            start_offset: 11,
            end_offset: 31,
            start_pos: Position {
                line: 1,
                column: 12
            },
            end_pos: Position {
                line: 1,
                column: 32
            },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TokenType::Quote,
            lit: String::from("\""),
            start_offset: 31,
            end_offset: 32,
            start_pos: Position {
                line: 1,
                column: 32
            },
            end_pos: Position {
                line: 1,
                column: 33
            },
            comments: vec![],
        }
    );
}

#[test]
fn test_scan_string_expr_escaped_quotes() {
    let text = r#""these \"\" are escaped quotes""#;
    let mut s = Scanner::new(text);
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TokenType::Quote,
            lit: String::from("\""),
            start_offset: 0,
            end_offset: 1,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 2 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TokenType::Text,
            lit: String::from(r#"these \"\" are escaped quotes"#),
            start_offset: 1,
            end_offset: 30,
            start_pos: Position { line: 1, column: 2 },
            end_pos: Position {
                line: 1,
                column: 31
            },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TokenType::Quote,
            lit: String::from("\""),
            start_offset: 30,
            end_offset: 31,
            start_pos: Position {
                line: 1,
                column: 31
            },
            end_pos: Position {
                line: 1,
                column: 32
            },
            comments: vec![],
        }
    );
}

#[test]
fn test_scan_string_expr_not_escaped_quotes() {
    let text = r#""this " is not an escaped quote""#;
    let mut s = Scanner::new(text);
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TokenType::Quote,
            lit: String::from("\""),
            start_offset: 0,
            end_offset: 1,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 2 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TokenType::Text,
            lit: String::from("this "),
            start_offset: 1,
            end_offset: 6,
            start_pos: Position { line: 1, column: 2 },
            end_pos: Position { line: 1, column: 7 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TokenType::Quote,
            lit: String::from("\""),
            start_offset: 6,
            end_offset: 7,
            start_pos: Position { line: 1, column: 7 },
            end_pos: Position { line: 1, column: 8 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TokenType::Text,
            lit: String::from(" is not an escaped quote"),
            start_offset: 7,
            end_offset: 31,
            start_pos: Position { line: 1, column: 8 },
            end_pos: Position {
                line: 1,
                column: 32
            },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TokenType::Quote,
            lit: String::from("\""),
            start_offset: 31,
            end_offset: 32,
            start_pos: Position {
                line: 1,
                column: 32
            },
            end_pos: Position {
                line: 1,
                column: 33
            },
            comments: vec![],
        }
    );
}

#[test]
fn test_scan_unread() {
    let text = "1 / 2 / 3";
    let mut s = Scanner::new(text);
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TokenType::Int,
            lit: String::from("1"),
            start_offset: 0,
            end_offset: 1,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 2 },
            comments: vec![],
        }
    );
    s.unread();
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Int,
            lit: String::from("1"),
            start_offset: 0,
            end_offset: 1,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 2 },
            comments: vec![],
        }
    );

    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TokenType::Regex,
            lit: String::from("/ 2 /"),
            start_offset: 2,
            end_offset: 7,
            start_pos: Position { line: 1, column: 3 },
            end_pos: Position { line: 1, column: 8 },
            comments: vec![],
        }
    );
    s.unread();
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Div,
            lit: String::from("/"),
            start_offset: 2,
            end_offset: 3,
            start_pos: Position { line: 1, column: 3 },
            end_pos: Position { line: 1, column: 4 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Int,
            lit: String::from("2"),
            start_offset: 4,
            end_offset: 5,
            start_pos: Position { line: 1, column: 5 },
            end_pos: Position { line: 1, column: 6 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Div,
            lit: String::from("/"),
            start_offset: 6,
            end_offset: 7,
            start_pos: Position { line: 1, column: 7 },
            end_pos: Position { line: 1, column: 8 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Int,
            lit: String::from("3"),
            start_offset: 8,
            end_offset: 9,
            start_pos: Position { line: 1, column: 9 },
            end_pos: Position {
                line: 1,
                column: 10
            },
            comments: vec![],
        }
    );
    // test unread idempotence
    s.unread();
    s.unread();
    s.unread();
    s.unread();

    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Int,
            lit: String::from("3"),
            start_offset: 8,
            end_offset: 9,
            start_pos: Position { line: 1, column: 9 },
            end_pos: Position {
                line: 1,
                column: 10
            },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Eof,
            lit: String::from(""),
            start_offset: 9,
            end_offset: 9,
            start_pos: Position {
                line: 1,
                column: 10
            },
            end_pos: Position {
                line: 1,
                column: 10
            },
            comments: vec![],
        }
    );
}

#[test]
fn test_scan_unread_with_newlines() {
    let text = r#"regex =


/foo/"#;
    let mut s = Scanner::new(text);
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Ident,
            lit: String::from("regex"),
            start_offset: 0,
            end_offset: 5,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 6 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Assign,
            lit: String::from("="),
            start_offset: 6,
            end_offset: 7,
            start_pos: Position { line: 1, column: 7 },
            end_pos: Position { line: 1, column: 8 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Div,
            lit: String::from("/"),
            start_offset: 10,
            end_offset: 11,
            start_pos: Position { line: 4, column: 1 },
            end_pos: Position { line: 4, column: 2 },
            comments: vec![],
        }
    );
    s.unread();
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TokenType::Regex,
            lit: String::from("/foo/"),
            start_offset: 10,
            end_offset: 15,
            start_pos: Position { line: 4, column: 1 },
            end_pos: Position { line: 4, column: 6 },
            comments: vec![],
        }
    );
}

#[test]
fn test_scan_with_regex_unread() {
    // We had a bug where calling scan_with_regex() and the next token
    // is a '/' with no closing '/' like a regex would have, it would return
    // incorrect token location info.
    let text = r#"3 * / 1
         y
    "#;
    let mut s = Scanner::new(text);

    let mut toks = vec![];
    toks.push(s.scan()); // 3
    toks.push(s.scan()); // *
    toks.push(s.scan()); // /
    s.unread();
    toks.push(s.scan_with_regex()); // /
    toks.push(s.scan()); // 1
    toks.push(s.scan()); // y
    toks.push(s.scan()); // Eof
    assert_eq!(
        vec![
            Token {
                tok: TokenType::Int,
                lit: String::from("3"),
                start_offset: 0,
                end_offset: 1,
                start_pos: Position { line: 1, column: 1 },
                end_pos: Position { line: 1, column: 2 },
                comments: vec![],
            },
            Token {
                tok: TokenType::Mul,
                lit: String::from("*"),
                start_offset: 2,
                end_offset: 3,
                start_pos: Position { line: 1, column: 3 },
                end_pos: Position { line: 1, column: 4 },
                comments: vec![],
            },
            Token {
                tok: TokenType::Div,
                lit: String::from("/"),
                start_offset: 4,
                end_offset: 5,
                start_pos: Position { line: 1, column: 5 },
                end_pos: Position { line: 1, column: 6 },
                comments: vec![],
            },
            Token {
                tok: TokenType::Div,
                lit: String::from("/"),
                start_offset: 4,
                end_offset: 5,
                start_pos: Position { line: 1, column: 5 },
                end_pos: Position { line: 1, column: 6 },
                comments: vec![],
            },
            Token {
                tok: TokenType::Int,
                lit: String::from("1"),
                start_offset: 6,
                end_offset: 7,
                start_pos: Position { line: 1, column: 7 },
                end_pos: Position { line: 1, column: 8 },
                comments: vec![],
            },
            Token {
                tok: TokenType::Ident,
                lit: String::from("y"),
                start_offset: 17,
                end_offset: 18,
                start_pos: Position {
                    line: 2,
                    column: 10
                },
                end_pos: Position {
                    line: 2,
                    column: 11
                },
                comments: vec![],
            },
            Token {
                tok: TokenType::Eof,
                lit: String::new(),
                start_offset: 23,
                end_offset: 23,
                start_pos: Position { line: 3, column: 5 },
                end_pos: Position { line: 3, column: 5 },
                comments: vec![],
            },
        ],
        toks
    );
}

#[test]
fn test_unclosed_quote() {
    let text = r#"x = "foo
        bar
        baz"#;
    let mut s = Scanner::new(text);
    let mut toks = vec![];
    toks.push(s.scan()); // x
    toks.push(s.scan()); // =
    toks.push(s.scan()); // "
    toks.push(s.scan()); // foo
    toks.push(s.scan()); // bar
    toks.push(s.scan()); // baz
    toks.push(s.scan()); // eof
    assert_eq!(
        vec![
            Token {
                tok: TokenType::Ident,
                lit: String::from("x"),
                start_offset: 0,
                end_offset: 1,
                start_pos: Position { line: 1, column: 1 },
                end_pos: Position { line: 1, column: 2 },
                comments: vec![],
            },
            Token {
                tok: TokenType::Assign,
                lit: String::from("="),
                start_offset: 2,
                end_offset: 3,
                start_pos: Position { line: 1, column: 3 },
                end_pos: Position { line: 1, column: 4 },
                comments: vec![],
            },
            Token {
                tok: TokenType::Quote,
                lit: String::from("\""),
                start_offset: 4,
                end_offset: 5,
                start_pos: Position { line: 1, column: 5 },
                end_pos: Position { line: 1, column: 6 },
                comments: vec![],
            },
            Token {
                tok: TokenType::Ident,
                lit: String::from("foo"),
                start_offset: 5,
                end_offset: 8,
                start_pos: Position { line: 1, column: 6 },
                end_pos: Position { line: 1, column: 9 },
                comments: vec![],
            },
            Token {
                tok: TokenType::Ident,
                lit: String::from("bar"),
                start_offset: 17,
                end_offset: 20,
                start_pos: Position { line: 2, column: 9 },
                end_pos: Position {
                    line: 2,
                    column: 12,
                },
                comments: vec![],
            },
            Token {
                tok: TokenType::Ident,
                lit: String::from("baz"),
                start_offset: 29,
                end_offset: 32,
                start_pos: Position { line: 3, column: 9 },
                end_pos: Position {
                    line: 3,
                    column: 12,
                },
                comments: vec![],
            },
            Token {
                tok: TokenType::Eof,
                lit: String::from(""),
                start_offset: 32,
                end_offset: 32,
                start_pos: Position {
                    line: 3,
                    column: 12
                },
                end_pos: Position {
                    line: 3,
                    column: 12
                },
                comments: vec![],
            }
        ],
        toks
    );
}

#[test]
fn test_scan_comments() {
    let text = r#"// this is a comment.
a
// comment with // nested comment.
// one more.
// last but not least.
1
// ok, that's it."#;
    let mut s = Scanner::new(text);
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Ident,
            lit: String::from("a"),
            start_offset: 22,
            end_offset: 23,
            start_pos: Position { line: 2, column: 1 },
            end_pos: Position { line: 2, column: 2 },
            comments: vec![Comment {
                text: String::from("// this is a comment.\n"),
            }],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Int,
            lit: String::from("1"),
            start_offset: 95,
            end_offset: 96,
            start_pos: Position { line: 6, column: 1 },
            end_pos: Position { line: 6, column: 2 },
            comments: vec![
                Comment {
                    text: String::from("// comment with // nested comment.\n"),
                },
                Comment {
                    text: String::from("// one more.\n"),
                },
                Comment {
                    text: String::from("// last but not least.\n"),
                },
            ]
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Eof,
            lit: String::from(""),
            start_offset: 114,
            end_offset: 114,
            start_pos: Position {
                line: 7,
                column: 18
            },
            end_pos: Position {
                line: 7,
                column: 18
            },
            comments: vec![Comment {
                text: String::from("// ok, that\'s it."),
            }]
        }
    );

    // with regex
    let mut s = Scanner::new(text);
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TokenType::Ident,
            lit: String::from("a"),
            start_offset: 22,
            end_offset: 23,
            start_pos: Position { line: 2, column: 1 },
            end_pos: Position { line: 2, column: 2 },
            comments: vec![Comment {
                text: String::from("// this is a comment.\n"),
            }],
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TokenType::Int,
            lit: String::from("1"),
            start_offset: 95,
            end_offset: 96,
            start_pos: Position { line: 6, column: 1 },
            end_pos: Position { line: 6, column: 2 },
            comments: vec![
                Comment {
                    text: String::from("// comment with // nested comment.\n"),
                },
                Comment {
                    text: String::from("// one more.\n"),
                },
                Comment {
                    text: String::from("// last but not least.\n"),
                },
            ]
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TokenType::Eof,
            lit: String::from(""),
            start_offset: 114,
            end_offset: 114,
            start_pos: Position {
                line: 7,
                column: 18
            },
            end_pos: Position {
                line: 7,
                column: 18
            },
            comments: vec![Comment {
                text: String::from("// ok, that\'s it."),
            }]
        }
    );
}

#[test]
fn test_scan_eof() {
    let text = r#""#;
    let mut s = Scanner::new(text);
    // idempotence with and without regex.
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Eof,
            lit: String::from(""),
            start_offset: 0,
            end_offset: 0,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 1 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TokenType::Eof,
            lit: String::from(""),
            start_offset: 0,
            end_offset: 0,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 1 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Eof,
            lit: String::from(""),
            start_offset: 0,
            end_offset: 0,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 1 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TokenType::Eof,
            lit: String::from(""),
            start_offset: 0,
            end_offset: 0,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 1 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Eof,
            lit: String::from(""),
            start_offset: 0,
            end_offset: 0,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 1 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TokenType::Eof,
            lit: String::from(""),
            start_offset: 0,
            end_offset: 0,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 1 },
            comments: vec![],
        }
    );
}

#[test]
fn test_scan_eof_trailing_spaces() {
    let mut text = String::new();
    text.push(' ');
    text.push('\t');
    text.push('\n');
    text.push('\t');
    text.push(' ');
    text.push('\t');
    text.push('\t');
    let mut s = Scanner::new(&text);
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Eof,
            lit: String::from(""),
            start_offset: 7,
            end_offset: 7,
            start_pos: Position { line: 2, column: 5 },
            end_pos: Position { line: 2, column: 5 },
            comments: vec![],
        }
    );

    let mut s = Scanner::new(&text);
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TokenType::Eof,
            lit: String::from(""),
            start_offset: 7,
            end_offset: 7,
            start_pos: Position { line: 2, column: 5 },
            end_pos: Position { line: 2, column: 5 },
            comments: vec![],
        }
    );
}

#[test]
fn test_illegal() {
    let text = r#"legal @ illegal"#;
    let mut s = Scanner::new(text);
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Ident,
            lit: String::from("legal"),
            start_offset: 0,
            end_offset: 5,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 6 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Illegal,
            lit: String::from("@"),
            start_offset: 6,
            end_offset: 7,
            start_pos: Position { line: 1, column: 7 },
            end_pos: Position { line: 1, column: 8 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Ident,
            lit: String::from("illegal"),
            start_offset: 8,
            end_offset: 15,
            start_pos: Position { line: 1, column: 9 },
            end_pos: Position {
                line: 1,
                column: 16
            },
            comments: vec![],
        }
    );

    // unread
    let mut s = Scanner::new(text);
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Ident,
            lit: String::from("legal"),
            start_offset: 0,
            end_offset: 5,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 6 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Illegal,
            lit: String::from("@"),
            start_offset: 6,
            end_offset: 7,
            start_pos: Position { line: 1, column: 7 },
            end_pos: Position { line: 1, column: 8 },
            comments: vec![],
        }
    );
    s.unread();
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Illegal,
            lit: String::from("@"),
            start_offset: 6,
            end_offset: 7,
            start_pos: Position { line: 1, column: 7 },
            end_pos: Position { line: 1, column: 8 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Ident,
            lit: String::from("illegal"),
            start_offset: 8,
            end_offset: 15,
            start_pos: Position { line: 1, column: 9 },
            end_pos: Position {
                line: 1,
                column: 16
            },
            comments: vec![],
        }
    );

    // with regex
    let mut s = Scanner::new(text);
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TokenType::Ident,
            lit: String::from("legal"),
            start_offset: 0,
            end_offset: 5,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 6 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TokenType::Illegal,
            lit: String::from("@"),
            start_offset: 6,
            end_offset: 7,
            start_pos: Position { line: 1, column: 7 },
            end_pos: Position { line: 1, column: 8 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TokenType::Ident,
            lit: String::from("illegal"),
            start_offset: 8,
            end_offset: 15,
            start_pos: Position { line: 1, column: 9 },
            end_pos: Position {
                line: 1,
                column: 16
            },
            comments: vec![],
        }
    );

    // unread
    let mut s = Scanner::new(text);
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TokenType::Ident,
            lit: String::from("legal"),
            start_offset: 0,
            end_offset: 5,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 6 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TokenType::Illegal,
            lit: String::from("@"),
            start_offset: 6,
            end_offset: 7,
            start_pos: Position { line: 1, column: 7 },
            end_pos: Position { line: 1, column: 8 },
            comments: vec![],
        }
    );
    s.unread();
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TokenType::Illegal,
            lit: String::from("@"),
            start_offset: 6,
            end_offset: 7,
            start_pos: Position { line: 1, column: 7 },
            end_pos: Position { line: 1, column: 8 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TokenType::Ident,
            lit: String::from("illegal"),
            start_offset: 8,
            end_offset: 15,
            start_pos: Position { line: 1, column: 9 },
            end_pos: Position {
                line: 1,
                column: 16
            },
            comments: vec![],
        }
    );
}

#[test]
fn test_scan_duration() {
    let text = r#"dur = 1y3mo2w1d4h1m30s1ms2µs70ns"#;
    let mut s = Scanner::new(text);
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Ident,
            lit: String::from("dur"),
            start_offset: 0,
            end_offset: 3,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 4 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Assign,
            lit: String::from("="),
            start_offset: 4,
            end_offset: 5,
            start_pos: Position { line: 1, column: 5 },
            end_pos: Position { line: 1, column: 6 },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Duration,
            lit: String::from("1y3mo2w1d4h1m30s1ms2µs70ns"),
            start_offset: 6,
            end_offset: 33,
            start_pos: Position { line: 1, column: 7 },
            end_pos: Position {
                line: 1,
                column: 34
            },
            comments: vec![],
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Eof,
            lit: String::from(""),
            start_offset: 33,
            end_offset: 33,
            start_pos: Position {
                line: 1,
                column: 34
            },
            end_pos: Position {
                line: 1,
                column: 34
            },
            comments: vec![],
        }
    );
}

#[test]
fn test_scan_offset() {
    let text = r#"ms = "multiline
string
"

// comment

c = 1 + 2




"#;
    let mut s = Scanner::new(text);
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Ident,
            lit: String::from("ms"),
            start_offset: 0,
            end_offset: 2,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 3 },
            comments: vec![],
        }
    );
    assert_eq!(0, s.offset(&Position { line: 1, column: 1 }));
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Assign,
            lit: String::from("="),
            start_offset: 3,
            end_offset: 4,
            start_pos: Position { line: 1, column: 4 },
            end_pos: Position { line: 1, column: 5 },
            comments: vec![],
        }
    );
    assert_eq!(3, s.offset(&Position { line: 1, column: 4 }));
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::String,
            lit: String::from("\"multiline\nstring\n\""),
            start_offset: 5,
            end_offset: 24,
            start_pos: Position { line: 1, column: 6 },
            end_pos: Position { line: 3, column: 2 },
            comments: vec![],
        }
    );
    assert_eq!(5, s.offset(&Position { line: 1, column: 6 }));
    assert_eq!(24, s.offset(&Position { line: 3, column: 2 }));
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Ident,
            lit: String::from("c"),
            start_offset: 38,
            end_offset: 39,
            start_pos: Position { line: 7, column: 1 },
            end_pos: Position { line: 7, column: 2 },
            comments: vec![Comment {
                text: String::from("// comment\n"),
            }]
        }
    );
    assert_eq!(38, s.offset(&Position { line: 7, column: 1 }));
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Assign,
            lit: String::from("="),
            start_offset: 40,
            end_offset: 41,
            start_pos: Position { line: 7, column: 3 },
            end_pos: Position { line: 7, column: 4 },
            comments: vec![],
        }
    );
    assert_eq!(40, s.offset(&Position { line: 7, column: 3 }));
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Int,
            lit: String::from("1"),
            start_offset: 42,
            end_offset: 43,
            start_pos: Position { line: 7, column: 5 },
            end_pos: Position { line: 7, column: 6 },
            comments: vec![],
        }
    );
    assert_eq!(42, s.offset(&Position { line: 7, column: 5 }));
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Add,
            lit: String::from("+"),
            start_offset: 44,
            end_offset: 45,
            start_pos: Position { line: 7, column: 7 },
            end_pos: Position { line: 7, column: 8 },
            comments: vec![],
        }
    );
    assert_eq!(44, s.offset(&Position { line: 7, column: 7 }));
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Int,
            lit: String::from("2"),
            start_offset: 46,
            end_offset: 47,
            start_pos: Position { line: 7, column: 9 },
            end_pos: Position {
                line: 7,
                column: 10
            },
            comments: vec![],
        }
    );
    assert_eq!(46, s.offset(&Position { line: 7, column: 9 }));
    assert_eq!(
        s.scan(),
        Token {
            tok: TokenType::Eof,
            lit: String::from(""),
            start_offset: 52,
            end_offset: 52,
            start_pos: Position {
                line: 12,
                column: 1
            },
            end_pos: Position {
                line: 12,
                column: 1
            },
            comments: vec![],
        }
    );
    assert_eq!(
        47,
        s.offset(&Position {
            line: 7,
            column: 10
        })
    );
    assert_eq!(
        52,
        s.offset(&Position {
            line: 12,
            column: 1
        })
    );

    // Ok, now re-assert every offset without scanning.
    // The scanner should keep the position unchanged.
    assert_eq!(0, s.offset(&Position { line: 1, column: 1 }));
    assert_eq!(3, s.offset(&Position { line: 1, column: 4 }));
    assert_eq!(5, s.offset(&Position { line: 1, column: 6 }));
    assert_eq!(24, s.offset(&Position { line: 3, column: 2 }));
    assert_eq!(38, s.offset(&Position { line: 7, column: 1 }));
    assert_eq!(40, s.offset(&Position { line: 7, column: 3 }));
    assert_eq!(42, s.offset(&Position { line: 7, column: 5 }));
    assert_eq!(44, s.offset(&Position { line: 7, column: 7 }));
    assert_eq!(46, s.offset(&Position { line: 7, column: 9 }));
    assert_eq!(
        52,
        s.offset(&Position {
            line: 12,
            column: 1,
        })
    );
}
