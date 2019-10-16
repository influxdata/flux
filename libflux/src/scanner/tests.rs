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
            pos: 0,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_LPAREN,
            lit: String::from("("),
            pos: 4,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_IDENT,
            lit: String::from("bucket"),
            pos: 5,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_COLON,
            lit: String::from(":"),
            pos: 11,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_STRING,
            lit: String::from("\"foo\""),
            pos: 12,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_RPAREN,
            lit: String::from(")"),
            pos: 17,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_PIPE_FORWARD,
            lit: String::from("|>"),
            pos: 19,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_IDENT,
            lit: String::from("range"),
            pos: 22,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_LPAREN,
            lit: String::from("("),
            pos: 27,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_IDENT,
            lit: String::from("start"),
            pos: 28,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_COLON,
            lit: String::from(":"),
            pos: 33,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_SUB,
            lit: String::from("-"),
            pos: 35,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_DURATION,
            lit: String::from("1m"),
            pos: 36,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_RPAREN,
            lit: String::from(")"),
            pos: 38,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_EOF,
            lit: String::from(""),
            pos: 39,
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
            pos: 0,
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_ADD,
            lit: String::from("+"),
            pos: 2,
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_IDENT,
            lit: String::from("b"),
            pos: 4,
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_REGEXEQ,
            lit: String::from("=~"),
            pos: 6,
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_REGEX,
            lit: String::from("/.*[0-9]/"),
            pos: 9,
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_DIV,
            lit: String::from("/"),
            pos: 19,
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_INT,
            lit: String::from("2"),
            pos: 21,
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_EOF,
            lit: String::from(""),
            pos: 22,
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
            pos: 0,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_STRINGEXPR,
            lit: String::from("${"),
            pos: 1,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_TEXT,
            lit: String::from("a + b}"),
            pos: 3,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_QUOTE,
            lit: String::from("\""),
            pos: 9,
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
            pos: 0,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_TEXT,
            lit: String::from("a + b = "),
            pos: 1,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_STRINGEXPR,
            lit: String::from("${"),
            pos: 9,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_TEXT,
            lit: String::from("a + b}"),
            pos: 11,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_QUOTE,
            lit: String::from("\""),
            pos: 17,
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
            pos: 0,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_TEXT,
            lit: String::from("a + b = "),
            pos: 1,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_STRINGEXPR,
            lit: String::from("${"),
            pos: 9,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_TEXT,
            lit: String::from("a + b} and a - b = "),
            pos: 11,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_STRINGEXPR,
            lit: String::from("${"),
            pos: 30,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_TEXT,
            lit: String::from("a - b}"),
            pos: 32,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_QUOTE,
            lit: String::from("\""),
            pos: 38,
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
            pos: 0,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_TEXT,
            lit: String::from("a + b = "),
            pos: 1,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_STRINGEXPR,
            lit: String::from("${"),
            pos: 9,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_TEXT,
            lit: String::from("a + b} and a - b = ?"),
            pos: 11,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_QUOTE,
            lit: String::from("\""),
            pos: 31,
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
            pos: 0,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_TEXT,
            lit: String::from(r#"these \"\" are escaped quotes"#),
            pos: 1,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_QUOTE,
            lit: String::from("\""),
            pos: 30,
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
            pos: 0,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_TEXT,
            lit: String::from("this "),
            pos: 1,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_QUOTE,
            lit: String::from("\""),
            pos: 6,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_TEXT,
            lit: String::from(" is not an escaped quote"),
            pos: 7,
        }
    );
    assert_eq!(
        s.scan_string_expr(),
        Token {
            tok: TOK_QUOTE,
            lit: String::from("\""),
            pos: 31,
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
            pos: 0,
        }
    );
    s.unread();
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_INT,
            lit: String::from("1"),
            pos: 0,
        }
    );

    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_REGEX,
            lit: String::from("/ 2 /"),
            pos: 2,
        }
    );
    s.unread();
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_DIV,
            lit: String::from("/"),
            pos: 2,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_INT,
            lit: String::from("2"),
            pos: 4,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_DIV,
            lit: String::from("/"),
            pos: 6,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_INT,
            lit: String::from("3"),
            pos: 8,
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
            pos: 8,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_EOF,
            lit: String::from(""),
            pos: 9,
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
            pos: 22,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_INT,
            lit: String::from("1"),
            pos: 95,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_EOF,
            lit: String::from(""),
            pos: 114,
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
            pos: 22,
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_INT,
            lit: String::from("1"),
            pos: 95,
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_EOF,
            lit: String::from(""),
            pos: 114,
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
            pos: 0,
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_EOF,
            lit: String::from(""),
            pos: 0,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_EOF,
            lit: String::from(""),
            pos: 0,
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_EOF,
            lit: String::from(""),
            pos: 0,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_EOF,
            lit: String::from(""),
            pos: 0,
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_EOF,
            lit: String::from(""),
            pos: 0,
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
            pos: 7,
        }
    );

    let cdata = CString::new(text.clone()).expect("CString::new failed");
    let mut s = Scanner::new(cdata);
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_EOF,
            lit: String::from(""),
            pos: 7,
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
            pos: 0,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_ILLEGAL,
            lit: String::from("@"),
            pos: 6,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_IDENT,
            lit: String::from("illegal"),
            pos: 8,
        }
    );

    // unread
    let mut s = Scanner::new(cdata.clone());
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_IDENT,
            lit: String::from("legal"),
            pos: 0,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_ILLEGAL,
            lit: String::from("@"),
            pos: 6,
        }
    );
    s.unread();
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_ILLEGAL,
            lit: String::from("@"),
            pos: 6,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_IDENT,
            lit: String::from("illegal"),
            pos: 8,
        }
    );

    // with regex
    let mut s = Scanner::new(cdata.clone());
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_IDENT,
            lit: String::from("legal"),
            pos: 0,
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_ILLEGAL,
            lit: String::from("@"),
            pos: 6,
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_IDENT,
            lit: String::from("illegal"),
            pos: 8,
        }
    );

    // unread
    let mut s = Scanner::new(cdata.clone());
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_IDENT,
            lit: String::from("legal"),
            pos: 0,
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_ILLEGAL,
            lit: String::from("@"),
            pos: 6,
        }
    );
    s.unread();
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_ILLEGAL,
            lit: String::from("@"),
            pos: 6,
        }
    );
    assert_eq!(
        s.scan_with_regex(),
        Token {
            tok: TOK_IDENT,
            lit: String::from("illegal"),
            pos: 8,
        }
    );
}

// TODO(affo): this fails.
#[test]
#[ignore] // See https://github.com/influxdata/flux/issues/1448
fn test_scan_duration() {
    let text = r#"dur = 1y3mo2w1d4h1m30s1ms2µs70ns"#;
    let cdata = CString::new(text).expect("CString::new failed");
    let mut s = Scanner::new(cdata);
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_IDENT,
            lit: String::from("dur"),
            pos: 0,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_ASSIGN,
            lit: String::from("="),
            pos: 4,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_DURATION,
            lit: String::from("1y3mo2w1d4h1m30s1ms2µs70ns"),
            pos: 6,
        }
    );
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_EOF,
            lit: String::from(""),
            pos: 32,
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
            pos: 0,
        }
    );
    assert_eq!(s.pos(0), Position { line: 1, column: 1 });
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_ASSIGN,
            lit: String::from("="),
            pos: 3,
        }
    );
    assert_eq!(s.pos(3), Position { line: 1, column: 4 });
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_STRING,
            lit: String::from("\"multiline\nstring\n\""),
            pos: 5,
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
            pos: 38,
        }
    );
    assert_eq!(s.pos(38), Position { line: 7, column: 1 });
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_ASSIGN,
            lit: String::from("="),
            pos: 40,
        }
    );
    assert_eq!(s.pos(40), Position { line: 7, column: 3 });
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_INT,
            lit: String::from("1"),
            pos: 42,
        }
    );
    assert_eq!(s.pos(42), Position { line: 7, column: 5 });
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_ADD,
            lit: String::from("+"),
            pos: 44,
        }
    );
    assert_eq!(s.pos(44), Position { line: 7, column: 7 });
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_INT,
            lit: String::from("2"),
            pos: 46,
        }
    );
    assert_eq!(s.pos(46), Position { line: 7, column: 9 });
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_EOF,
            lit: String::from(""),
            pos: 52,
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
            pos: 0,
        }
    );
    assert_eq!(0, s.offset(Position { line: 1, column: 1 }));
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_ASSIGN,
            lit: String::from("="),
            pos: 3,
        }
    );
    assert_eq!(3, s.offset(Position { line: 1, column: 4 }));
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_STRING,
            lit: String::from("\"multiline\nstring\n\""),
            pos: 5,
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
            pos: 38,
        }
    );
    assert_eq!(38, s.offset(Position { line: 7, column: 1 }));
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_ASSIGN,
            lit: String::from("="),
            pos: 40,
        }
    );
    assert_eq!(40, s.offset(Position { line: 7, column: 3 }));
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_INT,
            lit: String::from("1"),
            pos: 42,
        }
    );
    assert_eq!(42, s.offset(Position { line: 7, column: 5 }));
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_ADD,
            lit: String::from("+"),
            pos: 44,
        }
    );
    assert_eq!(44, s.offset(Position { line: 7, column: 7 }));
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_INT,
            lit: String::from("2"),
            pos: 46,
        }
    );
    assert_eq!(46, s.offset(Position { line: 7, column: 9 }));
    assert_eq!(
        s.scan(),
        Token {
            tok: TOK_EOF,
            lit: String::from(""),
            pos: 52,
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
