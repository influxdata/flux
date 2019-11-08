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
            end_pos: Position { line: 1, column: 5 }
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
            end_pos: Position { line: 1, column: 6 }
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
            }
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
            }
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
            }
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
            }
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
            }
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
            }
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
            }
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
            }
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
            }
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
            }
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
            }
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
            }
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_EOF,
            lit: String::from(""),
            start_offset: 39,
            end_offset: 39,
            start_pos: Position { line: 0, column: 0 },
            end_pos: Position { line: 0, column: 0 }
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
            end_pos: Position { line: 1, column: 2 }
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
            end_pos: Position { line: 1, column: 4 }
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
            end_pos: Position { line: 1, column: 6 }
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
            end_pos: Position { line: 1, column: 9 }
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
            }
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
            }
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
            }
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_EOF,
            lit: String::from(""),
            start_offset: 22,
            end_offset: 22,
            start_pos: Position { line: 0, column: 0 },
            end_pos: Position { line: 0, column: 0 }
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
            end_pos: Position { line: 1, column: 2 }
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
            end_pos: Position { line: 1, column: 4 }
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
            }
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
            }
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
            end_pos: Position { line: 1, column: 2 }
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
            }
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
            }
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
            }
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
            }
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
            end_pos: Position { line: 1, column: 2 }
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
            }
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
            }
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
            }
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
            }
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
            }
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
            }
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
            end_pos: Position { line: 1, column: 2 }
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
            }
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
            }
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
            }
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
            }
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
            end_pos: Position { line: 1, column: 2 }
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
            }
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
            }
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
            end_pos: Position { line: 1, column: 2 }
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
            end_pos: Position { line: 1, column: 7 }
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
            end_pos: Position { line: 1, column: 8 }
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
            }
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
            }
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
            end_pos: Position { line: 1, column: 2 }
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
            end_pos: Position { line: 1, column: 2 }
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
            end_pos: Position { line: 1, column: 8 }
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
            end_pos: Position { line: 1, column: 4 }
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
            end_pos: Position { line: 1, column: 6 }
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
            end_pos: Position { line: 1, column: 8 }
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
            }
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
            }
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_EOF,
            lit: String::from(""),
            start_offset: 9,
            end_offset: 9,
            start_pos: Position { line: 0, column: 0 },
            end_pos: Position { line: 0, column: 0 }
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
        }
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
            end_pos: Position { line: 2, column: 2 }
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
            end_pos: Position { line: 6, column: 2 }
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_EOF,
            lit: String::from(""),
            start_offset: 114,
            end_offset: 114,
            start_pos: Position { line: 0, column: 0 },
            end_pos: Position { line: 0, column: 0 }
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
            end_pos: Position { line: 2, column: 2 }
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
            end_pos: Position { line: 6, column: 2 }
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_EOF,
            lit: String::from(""),
            start_offset: 114,
            end_offset: 114,
            start_pos: Position { line: 0, column: 0 },
            end_pos: Position { line: 0, column: 0 }
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
            start_pos: Position { line: 0, column: 0 },
            end_pos: Position { line: 0, column: 0 }
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_EOF,
            lit: String::from(""),
            start_offset: 0,
            end_offset: 0,
            start_pos: Position { line: 0, column: 0 },
            end_pos: Position { line: 0, column: 0 }
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_EOF,
            lit: String::from(""),
            start_offset: 0,
            end_offset: 0,
            start_pos: Position { line: 0, column: 0 },
            end_pos: Position { line: 0, column: 0 }
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_EOF,
            lit: String::from(""),
            start_offset: 0,
            end_offset: 0,
            start_pos: Position { line: 0, column: 0 },
            end_pos: Position { line: 0, column: 0 }
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_EOF,
            lit: String::from(""),
            start_offset: 0,
            end_offset: 0,
            start_pos: Position { line: 0, column: 0 },
            end_pos: Position { line: 0, column: 0 }
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_EOF,
            lit: String::from(""),
            start_offset: 0,
            end_offset: 0,
            start_pos: Position { line: 0, column: 0 },
            end_pos: Position { line: 0, column: 0 }
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
            start_pos: Position { line: 0, column: 0 },
            end_pos: Position { line: 0, column: 0 }
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
            start_pos: Position { line: 0, column: 0 },
            end_pos: Position { line: 0, column: 0 }
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
            end_pos: Position { line: 1, column: 6 }
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
            end_pos: Position { line: 1, column: 8 }
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
            }
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
            end_pos: Position { line: 1, column: 6 }
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
            end_pos: Position { line: 1, column: 8 }
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
            end_pos: Position { line: 1, column: 8 }
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
            }
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
            end_pos: Position { line: 1, column: 6 }
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
            end_pos: Position { line: 1, column: 8 }
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
            }
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
            end_pos: Position { line: 1, column: 6 }
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
            end_pos: Position { line: 1, column: 8 }
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
            end_pos: Position { line: 1, column: 8 }
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
            }
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
            end_pos: Position { line: 1, column: 4 }
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
            end_pos: Position { line: 1, column: 6 }
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
            }
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_EOF,
            lit: String::from(""),
            start_offset: 33,
            end_offset: 33,
            start_pos: Position { line: 0, column: 0 },
            end_pos: Position { line: 0, column: 0 }
        }
    );
}

#[test]
fn test_scan_newlines() {
    let text = r#"multiline_string = "I
am
a
multiline
string.
"

// I am a
// comment.

1
2
3

4
// comment.
"#;
    let cdata = CString::new(text).expect("CString::new failed");
    let mut s = Scanner::new(cdata);
    assert_eq!(s.lines, vec![0]);
    s.scan(); // multiline_string
    s.scan(); // =
    s.scan(); // "..."
    s.scan(); // // I am a\n// comment.
    s.scan(); // "1"
    s.scan(); // "2"
    s.scan(); // "3"
    s.scan(); // "4"
    s.scan(); // // comment.\nEOF
    s.scan(); // EOF

    // we don't care of the intermediate steps for s.lines.
    // Only the final result is important.
    assert_eq!(
        s.lines,
        vec![0, 22, 25, 27, 37, 45, 47, 48, 58, 70, 71, 73, 75, 77, 78, 80, 92]
    );

    // with regex
    let cdata = CString::new(text).expect("CString::new failed");
    let mut s = Scanner::new(cdata);
    assert_eq!(s.lines, vec![0]);
    s.scan_with_regex(); // multiline_string
    s.scan_with_regex(); // =
    s.scan_with_regex(); // "..."
    s.scan_with_regex(); // // I am a\n// comment.
    s.scan_with_regex(); // "1"
    s.scan_with_regex(); // "2"
    s.scan_with_regex(); // "3"
    s.scan_with_regex(); // "4"
    s.scan_with_regex(); // // comment.\nEOF
    s.scan_with_regex(); // EOF
    assert_eq!(
        s.lines,
        vec![0, 22, 25, 27, 37, 45, 47, 48, 58, 70, 71, 73, 75, 77, 78, 80, 92]
    );
}

#[test]
fn test_scan_position() {
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
            end_pos: Position { line: 1, column: 3 }
        }
    );
    assert_eq!(s.pos(0), Position { line: 1, column: 1 });
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_ASSIGN,
            lit: String::from("="),
            start_offset: 3,
            end_offset: 4,
            start_pos: Position { line: 1, column: 4 },
            end_pos: Position { line: 1, column: 5 }
        }
    );
    assert_eq!(s.pos(3), Position { line: 1, column: 4 });
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_STRING,
            lit: String::from("\"multiline\nstring\n\""),
            start_offset: 5,
            end_offset: 24,
            start_pos: Position { line: 1, column: 6 },
            end_pos: Position { line: 3, column: 2 }
        }
    );
    assert_eq!(s.pos(5), Position { line: 1, column: 6 });
    assert_eq!(s.pos(16), Position { line: 2, column: 1 });
    assert_eq!(s.pos(20), Position { line: 2, column: 5 });
    assert_eq!(s.pos(23), Position { line: 3, column: 1 });
    assert_eq!(s.pos(24), Position { line: 3, column: 2 });
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_IDENT,
            lit: String::from("c"),
            start_offset: 38,
            end_offset: 39,
            start_pos: Position { line: 7, column: 1 },
            end_pos: Position { line: 7, column: 2 }
        }
    );
    assert_eq!(s.pos(38), Position { line: 7, column: 1 });
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_ASSIGN,
            lit: String::from("="),
            start_offset: 40,
            end_offset: 41,
            start_pos: Position { line: 7, column: 3 },
            end_pos: Position { line: 7, column: 4 }
        }
    );
    assert_eq!(s.pos(40), Position { line: 7, column: 3 });
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_INT,
            lit: String::from("1"),
            start_offset: 42,
            end_offset: 43,
            start_pos: Position { line: 7, column: 5 },
            end_pos: Position { line: 7, column: 6 }
        }
    );
    assert_eq!(s.pos(42), Position { line: 7, column: 5 });
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_ADD,
            lit: String::from("+"),
            start_offset: 44,
            end_offset: 45,
            start_pos: Position { line: 7, column: 7 },
            end_pos: Position { line: 7, column: 8 }
        }
    );
    assert_eq!(s.pos(44), Position { line: 7, column: 7 });
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
            }
        }
    );
    assert_eq!(s.pos(46), Position { line: 7, column: 9 });
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_EOF,
            lit: String::from(""),
            start_offset: 52,
            end_offset: 52,
            start_pos: Position { line: 0, column: 0 },
            end_pos: Position { line: 0, column: 0 }
        }
    );
    assert_eq!(s.pos(48), Position { line: 8, column: 1 });
    assert_eq!(s.pos(49), Position { line: 9, column: 1 });
    assert_eq!(
        s.pos(50),
        Position {
            line: 10,
            column: 1,
        }
    );
    assert_eq!(
        s.pos(51),
        Position {
            line: 11,
            column: 1,
        }
    );
    assert_eq!(
        s.pos(52),
        Position {
            line: 12,
            column: 1,
        }
    );

    // Ok, now re-assert every position without scanning.
    // The scanner should keep the position unchanged.
    assert_eq!(s.pos(0), Position { line: 1, column: 1 });
    assert_eq!(s.pos(3), Position { line: 1, column: 4 });
    assert_eq!(s.pos(5), Position { line: 1, column: 6 });
    assert_eq!(s.pos(16), Position { line: 2, column: 1 });
    assert_eq!(s.pos(20), Position { line: 2, column: 5 });
    assert_eq!(s.pos(23), Position { line: 3, column: 1 });
    assert_eq!(s.pos(24), Position { line: 3, column: 2 });
    assert_eq!(s.pos(38), Position { line: 7, column: 1 });
    assert_eq!(s.pos(40), Position { line: 7, column: 3 });
    assert_eq!(s.pos(42), Position { line: 7, column: 5 });
    assert_eq!(s.pos(44), Position { line: 7, column: 7 });
    assert_eq!(s.pos(46), Position { line: 7, column: 9 });
    assert_eq!(s.pos(48), Position { line: 8, column: 1 });
    assert_eq!(s.pos(49), Position { line: 9, column: 1 });
    assert_eq!(
        s.pos(50),
        Position {
            line: 10,
            column: 1,
        }
    );
    assert_eq!(
        s.pos(51),
        Position {
            line: 11,
            column: 1,
        }
    );
    assert_eq!(
        s.pos(52),
        Position {
            line: 12,
            column: 1,
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
            end_pos: Position { line: 1, column: 3 }
        }
    );
    assert_eq!(0, s.offset(Position { line: 1, column: 1 }));
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_ASSIGN,
            lit: String::from("="),
            start_offset: 3,
            end_offset: 4,
            start_pos: Position { line: 1, column: 4 },
            end_pos: Position { line: 1, column: 5 }
        }
    );
    assert_eq!(3, s.offset(Position { line: 1, column: 4 }));
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_STRING,
            lit: String::from("\"multiline\nstring\n\""),
            start_offset: 5,
            end_offset: 24,
            start_pos: Position { line: 1, column: 6 },
            end_pos: Position { line: 3, column: 2 }
        }
    );
    assert_eq!(5, s.offset(Position { line: 1, column: 6 }));
    assert_eq!(16, s.offset(Position { line: 2, column: 1 }));
    assert_eq!(20, s.offset(Position { line: 2, column: 5 }));
    assert_eq!(23, s.offset(Position { line: 3, column: 1 }));
    assert_eq!(24, s.offset(Position { line: 3, column: 2 }));
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_IDENT,
            lit: String::from("c"),
            start_offset: 38,
            end_offset: 39,
            start_pos: Position { line: 7, column: 1 },
            end_pos: Position { line: 7, column: 2 }
        }
    );
    assert_eq!(38, s.offset(Position { line: 7, column: 1 }));
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_ASSIGN,
            lit: String::from("="),
            start_offset: 40,
            end_offset: 41,
            start_pos: Position { line: 7, column: 3 },
            end_pos: Position { line: 7, column: 4 }
        }
    );
    assert_eq!(40, s.offset(Position { line: 7, column: 3 }));
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_INT,
            lit: String::from("1"),
            start_offset: 42,
            end_offset: 43,
            start_pos: Position { line: 7, column: 5 },
            end_pos: Position { line: 7, column: 6 }
        }
    );
    assert_eq!(42, s.offset(Position { line: 7, column: 5 }));
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_ADD,
            lit: String::from("+"),
            start_offset: 44,
            end_offset: 45,
            start_pos: Position { line: 7, column: 7 },
            end_pos: Position { line: 7, column: 8 }
        }
    );
    assert_eq!(44, s.offset(Position { line: 7, column: 7 }));
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
            }
        }
    );
    assert_eq!(46, s.offset(Position { line: 7, column: 9 }));
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_EOF,
            lit: String::from(""),
            start_offset: 52,
            end_offset: 52,
            start_pos: Position { line: 0, column: 0 },
            end_pos: Position { line: 0, column: 0 }
        }
    );
    assert_eq!(48, s.offset(Position { line: 8, column: 1 }));
    assert_eq!(49, s.offset(Position { line: 9, column: 1 }));
    assert_eq!(
        50,
        s.offset(Position {
            line: 10,
            column: 1,
        })
    );
    assert_eq!(
        51,
        s.offset(Position {
            line: 11,
            column: 1,
        })
    );
    assert_eq!(
        52,
        s.offset(Position {
            line: 12,
            column: 1,
        })
    );

    // Ok, now re-assert every offset without scanning.
    // The scanner should keep the position unchanged.
    assert_eq!(0, s.offset(Position { line: 1, column: 1 }));
    assert_eq!(3, s.offset(Position { line: 1, column: 4 }));
    assert_eq!(5, s.offset(Position { line: 1, column: 6 }));
    assert_eq!(16, s.offset(Position { line: 2, column: 1 }));
    assert_eq!(20, s.offset(Position { line: 2, column: 5 }));
    assert_eq!(23, s.offset(Position { line: 3, column: 1 }));
    assert_eq!(24, s.offset(Position { line: 3, column: 2 }));
    assert_eq!(38, s.offset(Position { line: 7, column: 1 }));
    assert_eq!(40, s.offset(Position { line: 7, column: 3 }));
    assert_eq!(42, s.offset(Position { line: 7, column: 5 }));
    assert_eq!(44, s.offset(Position { line: 7, column: 7 }));
    assert_eq!(46, s.offset(Position { line: 7, column: 9 }));
    assert_eq!(48, s.offset(Position { line: 8, column: 1 }));
    assert_eq!(49, s.offset(Position { line: 9, column: 1 }));
    assert_eq!(
        50,
        s.offset(Position {
            line: 10,
            column: 1,
        })
    );
    assert_eq!(
        51,
        s.offset(Position {
            line: 11,
            column: 1,
        })
    );
    assert_eq!(
        52,
        s.offset(Position {
            line: 12,
            column: 1,
        })
    );
}
