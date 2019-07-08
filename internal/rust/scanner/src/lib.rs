include!(concat!(env!("OUT_DIR"), "/bindings.rs"));

pub type CChar = i8;

use std::ffi::CString;
use std::str;

pub struct Scanner {
    data: CString,
    ps: *const CChar,
    p: *const CChar,
    pe: *const CChar,
    eof: *const CChar,
    checkpoint: *const CChar,
    token: T,
    ts: u32,
    te: u32,
}

#[derive(Debug, PartialEq, Clone)]
pub struct Token {
    pub tok: T,
    pub lit: String,
}

impl Scanner {
    // New creates a scanner with the provided input.
    pub fn new(data: CString) -> Scanner {
        let ptr = data.as_ptr();
        let bytes = data.as_bytes();
        let end = ((ptr as usize) + bytes.len()) as *const CChar;
        return Scanner {
            data: data,
            ps: ptr,
            p: ptr,
            pe: end,
            eof: end,
            ts: 0,
            te: 0,
            token: T_ILLEGAL,
            checkpoint: ptr,
        };
    }

    // scan produces the next token from the input.
    pub fn scan(&mut self) -> Token {
        self._scan(false)
    }

    // scan_with_regex produces the next token from the input accounting for regex.
    pub fn scan_with_regex(&mut self) -> Token {
        self._scan(true)
    }

    fn _scan(&mut self, with_regex: bool) -> Token {
        if self.p == self.eof {
            return Token {
                tok: T_EOF,
                lit: String::from(""),
            };
        }
        self.checkpoint = self.p;
        unsafe {
            let error = scan(
                if with_regex { 1 } else { 0 },
                &mut self.p as *mut *const CChar,
                self.ps as *const CChar,
                self.pe as *const CChar,
                self.eof as *const CChar,
                &mut self.token as *mut u32,
                &mut self.ts as *mut u32,
                &mut self.te as *mut u32,
            );
            if error != 0 {
                // Execution failed meaning we hit a pattern that we don't support and
                // doesn't produce a token. Use the unicode library to decode the next character
                // in the sequence so we don't break up any unicode tokens.
                let nc = std::str::from_utf8_unchecked(&self.data.as_bytes()[(self.ts as usize)..])
                    .chars()
                    .next();
                match nc {
                    Some(nc) => {
                        let size = nc.len_utf8();
                        // Advance the data pointer to after the character we just emitted.
                        self.p = self.p.offset(size as isize);
                        return Token {
                            tok: T_ILLEGAL,
                            lit: nc.to_string(),
                        };
                    }
                    // This should be impossible as we would have produced an EOF token
                    // instead, but going to handle this anyway as in this impossible scenario
                    // we would enter an infinite loop if we continued scanning past the token.
                    None => {
                        return Token {
                            tok: T_EOF,
                            lit: String::from(""),
                        }
                    }
                }
            }
            if self.token == T_ILLEGAL && self.p == self.eof {
                return Token {
                    tok: T_EOF,
                    lit: String::from(""),
                };
            }
            let t = Token {
                tok: self.token,
                lit: String::from(str::from_utf8_unchecked(
                    &self.data.as_bytes()[(self.ts as usize)..(self.te as usize)],
                )),
            };
            // skipping comments.
            // TODO(affo): return comments to attach them to nodes within the AST.
            match t {
                Token { tok: T_COMMENT, .. } => self.scan(),
                _ => t,
            }
        }
    }

    // unread will reset the Scanner to go back to the Scanner's location
    // before the last scan_with_regex or scan call. If either of the scan_with_regex methods
    // returned an EOF token, a call to unread will not unread the discarded whitespace.
    // This method is a no-op if called multiple times.
    pub fn unread(&mut self) {
        self.p = self.checkpoint;
    }
}

#[cfg(test)]
mod tests {
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
                tok: T_IDENT,
                lit: String::from("from"),
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                tok: T_LPAREN,
                lit: String::from("("),
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                tok: T_IDENT,
                lit: String::from("bucket"),
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                tok: T_COLON,
                lit: String::from(":"),
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                tok: T_STRING,
                lit: String::from("\"foo\""),
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                tok: T_RPAREN,
                lit: String::from(")"),
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                tok: T_PIPE_FORWARD,
                lit: String::from("|>"),
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                tok: T_IDENT,
                lit: String::from("range"),
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                tok: T_LPAREN,
                lit: String::from("("),
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                tok: T_IDENT,
                lit: String::from("start"),
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                tok: T_COLON,
                lit: String::from(":"),
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                tok: T_SUB,
                lit: String::from("-"),
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                tok: T_DURATION,
                lit: String::from("1m"),
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                tok: T_RPAREN,
                lit: String::from(")"),
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                tok: T_EOF,
                lit: String::from(""),
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
                tok: T_IDENT,
                lit: String::from("a"),
            }
        );
        assert_eq!(
            s.scan_with_regex(),
            Token {
                tok: T_ADD,
                lit: String::from("+"),
            }
        );
        assert_eq!(
            s.scan_with_regex(),
            Token {
                tok: T_IDENT,
                lit: String::from("b"),
            }
        );
        assert_eq!(
            s.scan_with_regex(),
            Token {
                tok: T_REGEXEQ,
                lit: String::from("=~"),
            }
        );
        assert_eq!(
            s.scan_with_regex(),
            Token {
                tok: T_REGEX,
                lit: String::from("/.*[0-9]/"),
            }
        );
        assert_eq!(
            s.scan_with_regex(),
            Token {
                tok: T_DIV,
                lit: String::from("/"),
            }
        );
        assert_eq!(
            s.scan_with_regex(),
            Token {
                tok: T_INT,
                lit: String::from("2"),
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
                tok: T_INT,
                lit: String::from("1"),
            }
        );
        s.unread();
        assert_eq!(
            s.scan(),
            Token {
                tok: T_INT,
                lit: String::from("1"),
            }
        );

        assert_eq!(
            s.scan_with_regex(),
            Token {
                tok: T_REGEX,
                lit: String::from("/ 2 /"),
            }
        );
        s.unread();
        assert_eq!(
            s.scan(),
            Token {
                tok: T_DIV,
                lit: String::from("/"),
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                tok: T_INT,
                lit: String::from("2"),
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                tok: T_DIV,
                lit: String::from("/"),
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                tok: T_INT,
                lit: String::from("3"),
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
                tok: T_INT,
                lit: String::from("3"),
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
                tok: T_IDENT,
                lit: String::from("a"),
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                tok: T_INT,
                lit: String::from("1"),
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                tok: T_EOF,
                lit: String::from(""),
            }
        );

        // with regex
        let cdata = CString::new(text).expect("CString::new failed");
        let mut s = Scanner::new(cdata);
        assert_eq!(
            s.scan_with_regex(),
            Token {
                tok: T_IDENT,
                lit: String::from("a"),
            }
        );
        assert_eq!(
            s.scan_with_regex(),
            Token {
                tok: T_INT,
                lit: String::from("1"),
            }
        );
        assert_eq!(
            s.scan_with_regex(),
            Token {
                tok: T_EOF,
                lit: String::from(""),
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
                tok: T_EOF,
                lit: String::from(""),
            }
        );
        assert_eq!(
            s.scan_with_regex(),
            Token {
                tok: T_EOF,
                lit: String::from(""),
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                tok: T_EOF,
                lit: String::from(""),
            }
        );
        assert_eq!(
            s.scan_with_regex(),
            Token {
                tok: T_EOF,
                lit: String::from(""),
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                tok: T_EOF,
                lit: String::from(""),
            }
        );
        assert_eq!(
            s.scan_with_regex(),
            Token {
                tok: T_EOF,
                lit: String::from(""),
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
                tok: T_EOF,
                lit: String::from(""),
            }
        );

        let cdata = CString::new(text.clone()).expect("CString::new failed");
        let mut s = Scanner::new(cdata);
        assert_eq!(
            s.scan_with_regex(),
            Token {
                tok: T_EOF,
                lit: String::from(""),
            }
        );
    }

    // TODO(affo): this fails.
    #[test]
    fn test_scan_duration() {
        let text = r#"dur = 1y3mo2w1d4h1m30s1ms2µs70ns"#;
        let cdata = CString::new(text).expect("CString::new failed");
        let mut s = Scanner::new(cdata);
        assert_eq!(
            s.scan(),
            Token {
                tok: T_IDENT,
                lit: String::from("dur"),
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                tok: T_ASSIGN,
                lit: String::from("="),
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                tok: T_DURATION,
                lit: String::from("1y3mo2w1d4h1m30s1ms2µs70ns"),
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                tok: T_EOF,
                lit: String::from(""),
            }
        );
    }
}
