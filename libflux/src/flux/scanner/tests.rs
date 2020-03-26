use super::*;
use std::ffi::CString;

#[test]
fn test_scan() {
    let text = "from(bucket:\"foo\") |> range(start: -1m)";
    let cdata = CString::new(text).expect("CString::new failed");
    let mut s = Scanner::new(cdata);
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_IDENT,
            lit: String::from("from"),
            start_offset: 0,
            end_offset: 4,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 5 },
            comments: None,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_LPAREN,
            lit: String::from("("),
            start_offset: 4,
            end_offset: 5,
            start_pos: Position { line: 1, column: 5 },
            end_pos: Position { line: 1, column: 6 },
            comments: None,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_IDENT,
            lit: String::from("bucket"),
            start_offset: 5,
            end_offset: 11,
            start_pos: Position { line: 1, column: 6 },
            end_pos: Position {
                line: 1,
                column: 12
            },
            comments: None,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_COLON,
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
            comments: None,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_STRING,
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
            comments: None,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_RPAREN,
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
            comments: None,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_PIPE_FORWARD,
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
            comments: None,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_IDENT,
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
            comments: None,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_LPAREN,
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
            comments: None,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_IDENT,
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
            comments: None,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_COLON,
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
            comments: None,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_SUB,
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
            comments: None,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_DURATION,
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
            comments: None,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_RPAREN,
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
            comments: None,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_EOF,
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
            comments: None,
        }
    );
}

#[test]
fn test_scan_with_regex() {
    let text = "a + b =~ /.*[0-9]/ / 2";
    let cdata = CString::new(text).expect("CString::new failed");
    let mut s = Scanner::new(cdata);
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_IDENT,
            lit: String::from("a"),
            start_offset: 0,
            end_offset: 1,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 2 },
            comments: None,
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_ADD,
            lit: String::from("+"),
            start_offset: 2,
            end_offset: 3,
            start_pos: Position { line: 1, column: 3 },
            end_pos: Position { line: 1, column: 4 },
            comments: None,
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_IDENT,
            lit: String::from("b"),
            start_offset: 4,
            end_offset: 5,
            start_pos: Position { line: 1, column: 5 },
            end_pos: Position { line: 1, column: 6 },
            comments: None,
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_REGEXEQ,
            lit: String::from("=~"),
            start_offset: 6,
            end_offset: 8,
            start_pos: Position { line: 1, column: 7 },
            end_pos: Position { line: 1, column: 9 },
            comments: None,
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_REGEX,
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
            comments: None,
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_DIV,
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
            comments: None,
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_INT,
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
            comments: None,
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_EOF,
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
            comments: None,
        }
    );
}

#[test]
fn test_scan_string_expr_simple() {
    let text = r#""${a + b}""#;
    let cdata = CString::new(text).expect("CString::new failed");
    let mut s = Scanner::new(cdata);
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_QUOTE,
            lit: String::from("\""),
            start_offset: 0,
            end_offset: 1,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 2 },
            comments: None,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_STRINGEXPR,
            lit: String::from("${"),
            start_offset: 1,
            end_offset: 3,
            start_pos: Position { line: 1, column: 2 },
            end_pos: Position { line: 1, column: 4 },
            comments: None,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_TEXT,
            lit: String::from("a + b}"),
            start_offset: 3,
            end_offset: 9,
            start_pos: Position { line: 1, column: 4 },
            end_pos: Position {
                line: 1,
                column: 10
            },
            comments: None,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_QUOTE,
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
            comments: None,
        }
    );
}

#[test]
fn test_scan_string_expr_start_with_text() {
    let text = r#""a + b = ${a + b}""#;
    let cdata = CString::new(text).expect("CString::new failed");
    let mut s = Scanner::new(cdata);
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_QUOTE,
            lit: String::from("\""),
            start_offset: 0,
            end_offset: 1,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 2 },
            comments: None,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_TEXT,
            lit: String::from("a + b = "),
            start_offset: 1,
            end_offset: 9,
            start_pos: Position { line: 1, column: 2 },
            end_pos: Position {
                line: 1,
                column: 10
            },
            comments: None,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_STRINGEXPR,
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
            comments: None,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_TEXT,
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
            comments: None,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_QUOTE,
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
            comments: None,
        }
    );
}

#[test]
fn test_scan_string_expr_multiple() {
    let text = r#""a + b = ${a + b} and a - b = ${a - b}""#;
    let cdata = CString::new(text).expect("CString::new failed");
    let mut s = Scanner::new(cdata);
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_QUOTE,
            lit: String::from("\""),
            start_offset: 0,
            end_offset: 1,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 2 },
            comments: None,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_TEXT,
            lit: String::from("a + b = "),
            start_offset: 1,
            end_offset: 9,
            start_pos: Position { line: 1, column: 2 },
            end_pos: Position {
                line: 1,
                column: 10
            },
            comments: None,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_STRINGEXPR,
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
            comments: None,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_TEXT,
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
            comments: None,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_STRINGEXPR,
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
            comments: None,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_TEXT,
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
            comments: None,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_QUOTE,
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
            comments: None,
        }
    );
}

#[test]
fn test_scan_string_expr_end_with_text() {
    let text = r#""a + b = ${a + b} and a - b = ?""#;
    let cdata = CString::new(text).expect("CString::new failed");
    let mut s = Scanner::new(cdata);
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_QUOTE,
            lit: String::from("\""),
            start_offset: 0,
            end_offset: 1,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 2 },
            comments: None,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_TEXT,
            lit: String::from("a + b = "),
            start_offset: 1,
            end_offset: 9,
            start_pos: Position { line: 1, column: 2 },
            end_pos: Position {
                line: 1,
                column: 10
            },
            comments: None,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_STRINGEXPR,
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
            comments: None,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_TEXT,
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
            comments: None,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_QUOTE,
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
            comments: None,
        }
    );
}

#[test]
fn test_scan_string_expr_escaped_quotes() {
    let text = r#""these \"\" are escaped quotes""#;
    let cdata = CString::new(text).expect("CString::new failed");
    let mut s = Scanner::new(cdata);
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_QUOTE,
            lit: String::from("\""),
            start_offset: 0,
            end_offset: 1,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 2 },
            comments: None,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_TEXT,
            lit: String::from(r#"these \"\" are escaped quotes"#),
            start_offset: 1,
            end_offset: 30,
            start_pos: Position { line: 1, column: 2 },
            end_pos: Position {
                line: 1,
                column: 31
            },
            comments: None,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_QUOTE,
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
            comments: None,
        }
    );
}

#[test]
fn test_scan_string_expr_not_escaped_quotes() {
    let text = r#""this " is not an escaped quote""#;
    let cdata = CString::new(text).expect("CString::new failed");
    let mut s = Scanner::new(cdata);
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_QUOTE,
            lit: String::from("\""),
            start_offset: 0,
            end_offset: 1,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 2 },
            comments: None,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_TEXT,
            lit: String::from("this "),
            start_offset: 1,
            end_offset: 6,
            start_pos: Position { line: 1, column: 2 },
            end_pos: Position { line: 1, column: 7 },
            comments: None,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_QUOTE,
            lit: String::from("\""),
            start_offset: 6,
            end_offset: 7,
            start_pos: Position { line: 1, column: 7 },
            end_pos: Position { line: 1, column: 8 },
            comments: None,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_TEXT,
            lit: String::from(" is not an escaped quote"),
            start_offset: 7,
            end_offset: 31,
            start_pos: Position { line: 1, column: 8 },
            end_pos: Position {
                line: 1,
                column: 32
            },
            comments: None,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_QUOTE,
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
            comments: None,
        }
    );
}

#[test]
fn test_scan_unread() {
    let text = "1 / 2 / 3";
    let cdata = CString::new(text).expect("CString::new failed");
    let mut s = Scanner::new(cdata);
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_INT,
            lit: String::from("1"),
            start_offset: 0,
            end_offset: 1,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 2 },
            comments: None,
        }
    );
    s.unread();
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_INT,
            lit: String::from("1"),
            start_offset: 0,
            end_offset: 1,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 2 },
            comments: None,
        }
    );

    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_REGEX,
            lit: String::from("/ 2 /"),
            start_offset: 2,
            end_offset: 7,
            start_pos: Position { line: 1, column: 3 },
            end_pos: Position { line: 1, column: 8 },
            comments: None,
        }
    );
    s.unread();
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_DIV,
            lit: String::from("/"),
            start_offset: 2,
            end_offset: 3,
            start_pos: Position { line: 1, column: 3 },
            end_pos: Position { line: 1, column: 4 },
            comments: None,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_INT,
            lit: String::from("2"),
            start_offset: 4,
            end_offset: 5,
            start_pos: Position { line: 1, column: 5 },
            end_pos: Position { line: 1, column: 6 },
            comments: None,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_DIV,
            lit: String::from("/"),
            start_offset: 6,
            end_offset: 7,
            start_pos: Position { line: 1, column: 7 },
            end_pos: Position { line: 1, column: 8 },
            comments: None,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_INT,
            lit: String::from("3"),
            start_offset: 8,
            end_offset: 9,
            start_pos: Position { line: 1, column: 9 },
            end_pos: Position {
                line: 1,
                column: 10
            },
            comments: None,
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
            tok: TOK_INT,
            lit: String::from("3"),
            start_offset: 8,
            end_offset: 9,
            start_pos: Position { line: 1, column: 9 },
            end_pos: Position {
                line: 1,
                column: 10
            },
            comments: None,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_EOF,
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
            comments: None,
        }
    );
}

#[test]
fn test_scan_unread_with_newlines() {
    let text = r#"regex =


/foo/"#;
    let cdata = CString::new(text).expect("CString::new failed");
    let mut s = Scanner::new(cdata);
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_IDENT,
            lit: String::from("regex"),
            start_offset: 0,
            end_offset: 5,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 6 },
            comments: None,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_ASSIGN,
            lit: String::from("="),
            start_offset: 6,
            end_offset: 7,
            start_pos: Position { line: 1, column: 7 },
            end_pos: Position { line: 1, column: 8 },
            comments: None,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_DIV,
            lit: String::from("/"),
            start_offset: 10,
            end_offset: 11,
            start_pos: Position { line: 4, column: 1 },
            end_pos: Position { line: 4, column: 2 },
            comments: None,
        }
    );
    s.unread();
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_REGEX,
            lit: String::from("/foo/"),
            start_offset: 10,
            end_offset: 15,
            start_pos: Position { line: 4, column: 1 },
            end_pos: Position { line: 4, column: 6 },
            comments: None,
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
    let cdata = CString::new(text).expect("CString::new failed");
    let mut s = Scanner::new(cdata);

    let mut toks = vec![];
    toks.push(s.scan()); // 3
    toks.push(s.scan()); // *
    toks.push(s.scan()); // /
    s.unread();
    toks.push(s.scan_with_regex()); // /
    toks.push(s.scan()); // 1
    toks.push(s.scan()); // y
    toks.push(s.scan()); // EOF
    assert_eq!(
        vec![
            Token {
                tok: TOK_INT,
                lit: String::from("3"),
                start_offset: 0,
                end_offset: 1,
                start_pos: Position { line: 1, column: 1 },
                end_pos: Position { line: 1, column: 2 },
                comments: None,
            },
            Token {
                tok: TOK_MUL,
                lit: String::from("*"),
                start_offset: 2,
                end_offset: 3,
                start_pos: Position { line: 1, column: 3 },
                end_pos: Position { line: 1, column: 4 },
                comments: None,
            },
            Token {
                tok: TOK_DIV,
                lit: String::from("/"),
                start_offset: 4,
                end_offset: 5,
                start_pos: Position { line: 1, column: 5 },
                end_pos: Position { line: 1, column: 6 },
                comments: None,
            },
            Token {
                tok: TOK_DIV,
                lit: String::from("/"),
                start_offset: 4,
                end_offset: 5,
                start_pos: Position { line: 1, column: 5 },
                end_pos: Position { line: 1, column: 6 },
                comments: None,
            },
            Token {
                tok: TOK_INT,
                lit: String::from("1"),
                start_offset: 6,
                end_offset: 7,
                start_pos: Position { line: 1, column: 7 },
                end_pos: Position { line: 1, column: 8 },
                comments: None,
            },
            Token {
                tok: TOK_IDENT,
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
                comments: None,
            },
            Token {
                tok: TOK_EOF,
                lit: String::new(),
                start_offset: 23,
                end_offset: 23,
                start_pos: Position { line: 3, column: 5 },
                end_pos: Position { line: 3, column: 5 },
                comments: None,
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
    let cdata = CString::new(text).expect("CString::new failed");
    let mut s = Scanner::new(cdata);
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
                tok: TOK_IDENT,
                lit: String::from("x"),
                start_offset: 0,
                end_offset: 1,
                start_pos: Position { line: 1, column: 1 },
                end_pos: Position { line: 1, column: 2 },
                comments: None,
            },
            Token {
                tok: TOK_ASSIGN,
                lit: String::from("="),
                start_offset: 2,
                end_offset: 3,
                start_pos: Position { line: 1, column: 3 },
                end_pos: Position { line: 1, column: 4 },
                comments: None,
            },
            Token {
                tok: TOK_QUOTE,
                lit: String::from("\""),
                start_offset: 4,
                end_offset: 5,
                start_pos: Position { line: 1, column: 5 },
                end_pos: Position { line: 1, column: 6 },
                comments: None,
            },
            Token {
                tok: TOK_IDENT,
                lit: String::from("foo"),
                start_offset: 5,
                end_offset: 8,
                start_pos: Position { line: 1, column: 6 },
                end_pos: Position { line: 1, column: 9 },
                comments: None,
            },
            Token {
                tok: TOK_IDENT,
                lit: String::from("bar"),
                start_offset: 17,
                end_offset: 20,
                start_pos: Position { line: 2, column: 9 },
                end_pos: Position {
                    line: 2,
                    column: 12,
                },
                comments: None,
            },
            Token {
                tok: TOK_IDENT,
                lit: String::from("baz"),
                start_offset: 29,
                end_offset: 32,
                start_pos: Position { line: 3, column: 9 },
                end_pos: Position {
                    line: 3,
                    column: 12,
                },
                comments: None,
            },
            Token {
                tok: TOK_EOF,
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
                comments: None,
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
    let cdata = CString::new(text).expect("CString::new failed");
    let mut s = Scanner::new(cdata);
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_IDENT,
            lit: String::from("a"),
            start_offset: 22,
            end_offset: 23,
            start_pos: Position { line: 2, column: 1 },
            end_pos: Position { line: 2, column: 2 },
            comments: Some(Box::new(Token {
                tok: 2,
                lit: String::from("// this is a comment.\n"),
                start_offset: 0,
                end_offset: 22,
                start_pos: Position { line: 1, column: 1 },
                end_pos: Position { line: 2, column: 1 },
                comments: None
            })),
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_INT,
            lit: String::from("1"),
            start_offset: 95,
            end_offset: 96,
            start_pos: Position { line: 6, column: 1 },
            end_pos: Position { line: 6, column: 2 },
            comments: Some(Box::new(Token {
                tok: 2,
                lit: String::from("// last but not least.\n"),
                start_offset: 72,
                end_offset: 95,
                start_pos: Position { line: 5, column: 1 },
                end_pos: Position { line: 6, column: 1 },
                comments: Some(Box::new(Token {
                    tok: 2,
                    lit: String::from("// one more.\n"),
                    start_offset: 59,
                    end_offset: 72,
                    start_pos: Position { line: 4, column: 1 },
                    end_pos: Position { line: 5, column: 1 },
                    comments: Some(Box::new(Token {
                        tok: 2,
                        lit: String::from("// comment with // nested comment.\n"),
                        start_offset: 24,
                        end_offset: 59,
                        start_pos: Position { line: 3, column: 1 },
                        end_pos: Position { line: 4, column: 1 },
                        comments: None,
                    }))
                }))
            }))
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_EOF,
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
            comments: Some(Box::new(Token {
                tok: 2,
                lit: String::from("// ok, that\'s it."),
                start_offset: 97,
                end_offset: 114,
                start_pos: Position { line: 7, column: 1 },
                end_pos: Position {
                    line: 7,
                    column: 18
                },
                comments: None
            }))
        }
    );

    // with regex
    let cdata = CString::new(text).expect("CString::new failed");
    let mut s = Scanner::new(cdata);
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_IDENT,
            lit: String::from("a"),
            start_offset: 22,
            end_offset: 23,
            start_pos: Position { line: 2, column: 1 },
            end_pos: Position { line: 2, column: 2 },
            comments: Some(Box::new(Token {
                tok: 2,
                lit: String::from("// this is a comment.\n"),
                start_offset: 0,
                end_offset: 22,
                start_pos: Position { line: 1, column: 1 },
                end_pos: Position { line: 2, column: 1 },
                comments: None
            })),
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_INT,
            lit: String::from("1"),
            start_offset: 95,
            end_offset: 96,
            start_pos: Position { line: 6, column: 1 },
            end_pos: Position { line: 6, column: 2 },
            comments: Some(Box::new(Token {
                tok: 2,
                lit: String::from("// last but not least.\n"),
                start_offset: 72,
                end_offset: 95,
                start_pos: Position { line: 5, column: 1 },
                end_pos: Position { line: 6, column: 1 },
                comments: Some(Box::new(Token {
                    tok: 2,
                    lit: String::from("// one more.\n"),
                    start_offset: 59,
                    end_offset: 72,
                    start_pos: Position { line: 4, column: 1 },
                    end_pos: Position { line: 5, column: 1 },
                    comments: Some(Box::new(Token {
                        tok: 2,
                        lit: String::from("// comment with // nested comment.\n"),
                        start_offset: 24,
                        end_offset: 59,
                        start_pos: Position { line: 3, column: 1 },
                        end_pos: Position { line: 4, column: 1 },
                        comments: None
                    }))
                }))
            }))
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_EOF,
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
            comments: Some(Box::new(Token {
                tok: 2,
                lit: String::from("// ok, that\'s it."),
                start_offset: 97,
                end_offset: 114,
                start_pos: Position { line: 7, column: 1 },
                end_pos: Position {
                    line: 7,
                    column: 18
                },
                comments: None
            }))
        }
    );
}

#[test]
fn test_scan_eof() {
    let text = r#""#;
    let cdata = CString::new(text).expect("CString::new failed");
    let mut s = Scanner::new(cdata);
    // idempotence with and without regex.
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_EOF,
            lit: String::from(""),
            start_offset: 0,
            end_offset: 0,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 1 },
            comments: None,
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_EOF,
            lit: String::from(""),
            start_offset: 0,
            end_offset: 0,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 1 },
            comments: None,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_EOF,
            lit: String::from(""),
            start_offset: 0,
            end_offset: 0,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 1 },
            comments: None,
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_EOF,
            lit: String::from(""),
            start_offset: 0,
            end_offset: 0,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 1 },
            comments: None,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_EOF,
            lit: String::from(""),
            start_offset: 0,
            end_offset: 0,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 1 },
            comments: None,
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_EOF,
            lit: String::from(""),
            start_offset: 0,
            end_offset: 0,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 1 },
            comments: None,
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
    let cdata = CString::new(text.clone()).expect("CString::new failed");
    let mut s = Scanner::new(cdata);
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_EOF,
            lit: String::from(""),
            start_offset: 7,
            end_offset: 7,
            start_pos: Position { line: 2, column: 5 },
            end_pos: Position { line: 2, column: 5 },
            comments: None,
        }
    );

    let cdata = CString::new(text.clone()).expect("CString::new failed");
    let mut s = Scanner::new(cdata);
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_EOF,
            lit: String::from(""),
            start_offset: 7,
            end_offset: 7,
            start_pos: Position { line: 2, column: 5 },
            end_pos: Position { line: 2, column: 5 },
            comments: None,
        }
    );
}

#[test]
fn test_illegal() {
    let text = r#"legal @ illegal"#;
    let cdata = CString::new(text).expect("CString::new failed");
    let mut s = Scanner::new(cdata.clone());
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_IDENT,
            lit: String::from("legal"),
            start_offset: 0,
            end_offset: 5,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 6 },
            comments: None,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_ILLEGAL,
            lit: String::from("@"),
            start_offset: 6,
            end_offset: 7,
            start_pos: Position { line: 1, column: 7 },
            end_pos: Position { line: 1, column: 8 },
            comments: None,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_IDENT,
            lit: String::from("illegal"),
            start_offset: 8,
            end_offset: 15,
            start_pos: Position { line: 1, column: 9 },
            end_pos: Position {
                line: 1,
                column: 16
            },
            comments: None,
        }
    );

    // unread
    let mut s = Scanner::new(cdata.clone());
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_IDENT,
            lit: String::from("legal"),
            start_offset: 0,
            end_offset: 5,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 6 },
            comments: None,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_ILLEGAL,
            lit: String::from("@"),
            start_offset: 6,
            end_offset: 7,
            start_pos: Position { line: 1, column: 7 },
            end_pos: Position { line: 1, column: 8 },
            comments: None,
        }
    );
    s.unread();
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_ILLEGAL,
            lit: String::from("@"),
            start_offset: 6,
            end_offset: 7,
            start_pos: Position { line: 1, column: 7 },
            end_pos: Position { line: 1, column: 8 },
            comments: None,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_IDENT,
            lit: String::from("illegal"),
            start_offset: 8,
            end_offset: 15,
            start_pos: Position { line: 1, column: 9 },
            end_pos: Position {
                line: 1,
                column: 16
            },
            comments: None,
        }
    );

    // with regex
    let mut s = Scanner::new(cdata.clone());
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_IDENT,
            lit: String::from("legal"),
            start_offset: 0,
            end_offset: 5,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 6 },
            comments: None,
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_ILLEGAL,
            lit: String::from("@"),
            start_offset: 6,
            end_offset: 7,
            start_pos: Position { line: 1, column: 7 },
            end_pos: Position { line: 1, column: 8 },
            comments: None,
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_IDENT,
            lit: String::from("illegal"),
            start_offset: 8,
            end_offset: 15,
            start_pos: Position { line: 1, column: 9 },
            end_pos: Position {
                line: 1,
                column: 16
            },
            comments: None,
        }
    );

    // unread
    let mut s = Scanner::new(cdata.clone());
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_IDENT,
            lit: String::from("legal"),
            start_offset: 0,
            end_offset: 5,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 6 },
            comments: None,
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_ILLEGAL,
            lit: String::from("@"),
            start_offset: 6,
            end_offset: 7,
            start_pos: Position { line: 1, column: 7 },
            end_pos: Position { line: 1, column: 8 },
            comments: None,
        }
    );
    s.unread();
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_ILLEGAL,
            lit: String::from("@"),
            start_offset: 6,
            end_offset: 7,
            start_pos: Position { line: 1, column: 7 },
            end_pos: Position { line: 1, column: 8 },
            comments: None,
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_IDENT,
            lit: String::from("illegal"),
            start_offset: 8,
            end_offset: 15,
            start_pos: Position { line: 1, column: 9 },
            end_pos: Position {
                line: 1,
                column: 16
            },
            comments: None,
        }
    );
}

#[test]
fn test_scan_duration() {
    let text = r#"dur = 1y3mo2w1d4h1m30s1ms2µs70ns"#;
    let cdata = CString::new(text).expect("CString::new failed");
    let mut s = Scanner::new(cdata);
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_IDENT,
            lit: String::from("dur"),
            start_offset: 0,
            end_offset: 3,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 4 },
            comments: None,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_ASSIGN,
            lit: String::from("="),
            start_offset: 4,
            end_offset: 5,
            start_pos: Position { line: 1, column: 5 },
            end_pos: Position { line: 1, column: 6 },
            comments: None,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_DURATION,
            lit: String::from("1y3mo2w1d4h1m30s1ms2µs70ns"),
            start_offset: 6,
            end_offset: 33,
            start_pos: Position { line: 1, column: 7 },
            end_pos: Position {
                line: 1,
                column: 34
            },
            comments: None,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_EOF,
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
            comments: None,
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
    let cdata = CString::new(text).expect("CString::new failed");
    let mut s = Scanner::new(cdata);
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_IDENT,
            lit: String::from("ms"),
            start_offset: 0,
            end_offset: 2,
            start_pos: Position { line: 1, column: 1 },
            end_pos: Position { line: 1, column: 3 },
            comments: None,
        }
    );
    assert_eq!(0, s.offset(&Position { line: 1, column: 1 }));
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_ASSIGN,
            lit: String::from("="),
            start_offset: 3,
            end_offset: 4,
            start_pos: Position { line: 1, column: 4 },
            end_pos: Position { line: 1, column: 5 },
            comments: None,
        }
    );
    assert_eq!(3, s.offset(&Position { line: 1, column: 4 }));
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_STRING,
            lit: String::from("\"multiline\nstring\n\""),
            start_offset: 5,
            end_offset: 24,
            start_pos: Position { line: 1, column: 6 },
            end_pos: Position { line: 3, column: 2 },
            comments: None,
        }
    );
    assert_eq!(5, s.offset(&Position { line: 1, column: 6 }));
    assert_eq!(24, s.offset(&Position { line: 3, column: 2 }));
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_IDENT,
            lit: String::from("c"),
            start_offset: 38,
            end_offset: 39,
            start_pos: Position { line: 7, column: 1 },
            end_pos: Position { line: 7, column: 2 },
            comments: Some(Box::new(Token {
                tok: 2,
                lit: String::from("// comment\n"),
                start_offset: 26,
                end_offset: 37,
                start_pos: Position { line: 5, column: 1 },
                end_pos: Position { line: 6, column: 1 },
                comments: None
            }))
        }
    );
    assert_eq!(38, s.offset(&Position { line: 7, column: 1 }));
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_ASSIGN,
            lit: String::from("="),
            start_offset: 40,
            end_offset: 41,
            start_pos: Position { line: 7, column: 3 },
            end_pos: Position { line: 7, column: 4 },
            comments: None,
        }
    );
    assert_eq!(40, s.offset(&Position { line: 7, column: 3 }));
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_INT,
            lit: String::from("1"),
            start_offset: 42,
            end_offset: 43,
            start_pos: Position { line: 7, column: 5 },
            end_pos: Position { line: 7, column: 6 },
            comments: None,
        }
    );
    assert_eq!(42, s.offset(&Position { line: 7, column: 5 }));
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_ADD,
            lit: String::from("+"),
            start_offset: 44,
            end_offset: 45,
            start_pos: Position { line: 7, column: 7 },
            end_pos: Position { line: 7, column: 8 },
            comments: None,
        }
    );
    assert_eq!(44, s.offset(&Position { line: 7, column: 7 }));
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_INT,
            lit: String::from("2"),
            start_offset: 46,
            end_offset: 47,
            start_pos: Position { line: 7, column: 9 },
            end_pos: Position {
                line: 7,
                column: 10
            },
            comments: None,
        }
    );
    assert_eq!(46, s.offset(&Position { line: 7, column: 9 }));
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_EOF,
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
            comments: None,
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
