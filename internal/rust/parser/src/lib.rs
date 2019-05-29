extern crate libc;

use libc::{c_char, c_int};
use std::ffi::CString;

#[repr(C)]
struct scanner_t {
    p: *mut c_char,
    pe: *mut c_char,
    eof: *mut c_char,
    ts: *mut c_char,
    te: *mut c_char,
    token: c_int,
}

extern "C" {
    fn init(scanner: *mut scanner_t, str: *const c_char);
    fn scan(scanner: *mut scanner_t);
}

pub struct Scanner {
    s: *mut scanner_t,
}

impl Scanner {
    pub fn new(data: &CString) -> Scanner {
        let s: *mut scanner_t = &mut scanner_t {
            p: 0 as *mut c_char,
            pe: 0 as *mut c_char,
            eof: 0 as *mut c_char,
            ts: 0 as *mut c_char,
            te: 0 as *mut c_char,
            token: 0,
        };
        unsafe {
            init(s, data.as_ptr());
        }
        let scanner = Scanner { s: s };
        scanner.print("new");
        return scanner;
    }

    pub fn scan(&self) -> c_int {
        self.print("scan");
        unsafe {
            //let data = CStr::from_ptr((*self.s).p);
            //println!("data {}", data.to_str().unwrap());
            scan(self.s);
            if (*self.s).p == (*self.s).eof {
                return 1; // EOF token
            }
            //let token = CStr::from_ptr((*self.s).ts);
            //println!("token {}", token.to_str().unwrap());
            return (*self.s).token;
        }
    }
    pub fn print(&self, msg: &str) {
        unsafe {
            println!(
                "{} s:{:p} p:{:p} pe:{:p} eof:{:p} ts:{:p} te:{:p} token:{}",
                msg,
                self.s,
                (*self.s).p,
                (*self.s).pe,
                (*self.s).eof,
                (*self.s).ts,
                (*self.s).te,
                (*self.s).token,
            );
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
        let s = Scanner::new(&cdata);
        s.print("test");
        //assert_eq!(s.scan(), 17);
        //assert_eq!(s.scan(), 39);
        assert_eq!(cdata.to_str().unwrap(), "");
    }
}
