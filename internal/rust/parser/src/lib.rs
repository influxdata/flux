extern crate libc;

include!(concat!(env!("OUT_DIR"), "/bindings.rs"));

mod ast;
use libc::c_char;
use std::ffi::CString;
use std::str;

pub struct Scanner {
    data: CString,
    ps: *const c_char,
    p: *const c_char,
    pe: *const c_char,
    eof: *const c_char,
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
        let end = ((ptr as usize) + bytes.len()) as *const c_char;
        return Scanner {
            data: data,
            ps: ptr,
            p: ptr,
            pe: end,
            eof: end,
            ts: 0,
            te: 0,
            token: T_ILLEGAL,
        };
    }

    // Scan produces the next token from the intput.
    pub fn scan(&mut self) -> Token {
        if self.p == self.eof {
            return Token {
                tok: T_EOF,
                lit: String::from(""),
            };
        }
        unsafe {
            scan(
                &mut self.p as *mut *const c_char,
                self.ps as *const c_char,
                self.pe as *const c_char,
                self.eof as *const c_char,
                &mut self.token as *mut u32,
                &mut self.ts as *mut u32,
                &mut self.te as *mut u32,
            );
            return Token {
                tok: self.token,
                lit: String::from(str::from_utf8_unchecked(
                    &self.data.as_bytes()[(self.ts as usize)..(self.te as usize)],
                )),
            };
        }
    }
}

pub struct Parser {
    s: Scanner,
    t: Option<Token>,
    errs: Vec<String>,
}

impl Parser {
    pub fn new(src: &str) -> Parser {
        let cdata = CString::new(src).expect("CString::new failed");
        let s = Scanner::new(cdata);
        Parser {
            s: s,
            t: Option::None,
            errs: Vec::new(),
        }
    }

    // scan will read the next token from the Scanner. If peek has been used,
    // this will return the peeked token and consume it.
    fn scan(&mut self) -> Token {
        match self.t.clone() {
            Some(t) => {
                self.t = Option::None;
                return t;
            }
            None => return self.s.scan(),
        }
    }

    // peek will read the next token from the Scanner and then buffer it.
    // It will return information about the token.
    fn peek(&mut self) -> Token {
        match self.t.clone() {
            Some(t) => return t,
            None => {
                let t = self.s.scan();
                self.t = Some(t.clone());
                return t;
            }
        }
    }

    // consume will consume a token that has been retrieve using peek.
    // This will panic if a token has not been buffered with peek.
    fn consume(&mut self) {
        match self.t.clone() {
            Some(_) => self.t = Option::None,
            None => panic!("called consume on an unbuffered input"),
        }
    }

    // expect will continuously scan the input until it reads the requested
    // token. If a token has been buffered by peek, then the token will
    // be read if it matches or will be discarded if it is the wrong token.
    fn expect(&mut self, exp: T) -> Token {
        loop {
            let t = self.scan();
            match t.tok {
                tok if tok == exp => return t,
                T_EOF => {
                    self.errs.push(format!("expected {}, got EOF", exp));
                    return t;
                }
                _ => self.errs.push(format!(
                    "expected {}, got {} ({}) at {}",
                    exp, t.tok, t.lit, "position",
                )),
            }
        }
    }

    fn base_node(&self) -> ast::BaseNode {
        ast::BaseNode { errors: Vec::new() }
    }

    fn parse_package_clause(&mut self) -> Option<ast::PackageClause> {
        let t = self.peek();
        if t.tok == T_PACKAGE {
            self.consume();
            let ident = self.parse_identifier();
            return Some(ast::PackageClause {
                base: self.base_node(),
                name: ident,
            });
        }
        return None;
    }
    fn parse_identifier(&mut self) -> ast::Identifier {
        let t = self.expect(T_IDENT);
        return ast::Identifier {
            base: self.base_node(),
            name: t.lit,
        };
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
    fn test_parse_package_clause() {
        let mut p = Parser::new("package foo");
        let pc = p.parse_package_clause();
        assert_eq!(
            pc,
            Some(ast::PackageClause {
                base: ast::BaseNode { errors: Vec::new() },
                name: ast::Identifier {
                    base: ast::BaseNode { errors: Vec::new() },
                    name: String::from("foo"),
                },
            })
        )
    }
}
