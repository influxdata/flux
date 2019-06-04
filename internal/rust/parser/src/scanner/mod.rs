include!(concat!(env!("OUT_DIR"), "/bindings.rs"));

type c_char = i8;

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
}
