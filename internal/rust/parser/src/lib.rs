extern crate libc;

use libc::{c_char, c_int};
use std::ffi::CString;
use std::str;

extern "C" {
    fn scan(
        p: *mut *const c_char,
        data: *const c_char,
        pe: *const c_char,
        eof: *const c_char,
        token: *mut c_int,
        token_start: *mut c_int,
        token_len: *mut c_int,
    );
}

pub struct Scanner<'a> {
    data: &'a CString,
    ps: *const c_char,
    p: *const c_char,
    pe: *const c_char,
    eof: *const c_char,
    token: c_int,
    ts: c_int,
    te: c_int,
}

#[derive(Debug, PartialEq)]
pub struct Token<'a> {
    pub code: c_int,
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
            token: 0,
        };
    }

    pub fn scan(&mut self) -> Token {
        if self.p == self.eof {
            return Token { code: 1, lit: "" };
        }
        unsafe {
            scan(
                &mut self.p as *mut *const c_char,
                self.ps as *const c_char,
                self.pe as *const c_char,
                self.eof as *const c_char,
                &mut self.token as *mut c_int,
                &mut self.ts as *mut c_int,
                &mut self.te as *mut c_int,
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
                code: 17,
                lit: "from"
            }
        );
        assert_eq!(s.scan(), Token { code: 39, lit: "(" });
        assert_eq!(
            s.scan(),
            Token {
                code: 17,
                lit: "bucket"
            }
        );
        assert_eq!(s.scan(), Token { code: 47, lit: ":" });
        assert_eq!(
            s.scan(),
            Token {
                code: 20,
                lit: "\"foo\""
            }
        );
        assert_eq!(s.scan(), Token { code: 40, lit: ")" });
        assert_eq!(
            s.scan(),
            Token {
                code: 48,
                lit: "|>"
            }
        );
        assert_eq!(
            s.scan(),
            Token {
                code: 17,
                lit: "range"
            }
        );
        assert_eq!(s.scan(), Token { code: 39, lit: "(" });
        assert_eq!(
            s.scan(),
            Token {
                code: 17,
                lit: "start"
            }
        );
        assert_eq!(s.scan(), Token { code: 47, lit: ":" });
        assert_eq!(s.scan(), Token { code: 25, lit: "-" });
        assert_eq!(
            s.scan(),
            Token {
                code: 23,
                lit: "1m"
            }
        );
        assert_eq!(s.scan(), Token { code: 40, lit: ")" });
        assert_eq!(s.scan(), Token { code: 1, lit: "" });
    }
}
