extern crate libc;

include!(concat!(env!("OUT_DIR"), "/bindings.rs"));

use libc::c_char;
use std::ffi::CString;
use std::str;

pub struct Scanner<'a> {
    data: &'a CString,
    ps: *const c_char,
    p: *const c_char,
    pe: *const c_char,
    eof: *const c_char,
    token: T,
    ts: u32,
    te: u32,
}

#[derive(Debug, PartialEq)]
pub struct Token<'a> {
    pub code: T,
    pub lit: &'a str,
}

impl<'a> Scanner<'a> {
    pub fn new(data: &CString) -> Scanner {
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

    pub fn scan(&mut self) -> Token {
        if self.p == self.eof {
            return Token {
                code: T_EOF,
                lit: "",
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
                code: self.token,
                lit: str::from_utf8_unchecked(
                    &self.data.as_bytes()[(self.ts as usize)..(self.te as usize)],
                ),
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
        let mut s = Scanner::new(&cdata);
        assert_eq!(
            s.scan(),
            Token {
                code: T_IDENT,
                lit: "from"
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                code: T_LPAREN,
                lit: "("
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                code: T_IDENT,
                lit: "bucket"
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                code: T_COLON,
                lit: ":"
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                code: T_STRING,
                lit: "\"foo\""
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                code: T_RPAREN,
                lit: ")"
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                code: T_PIPE_FORWARD,
                lit: "|>"
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                code: T_IDENT,
                lit: "range"
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                code: T_LPAREN,
                lit: "("
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                code: T_IDENT,
                lit: "start"
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                code: T_COLON,
                lit: ":"
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                code: T_SUB,
                lit: "-"
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                code: T_DURATION,
                lit: "1m"
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                code: T_RPAREN,
                lit: ")"
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                code: T_EOF,
                lit: ""
            }
        );
    }
}
