extern crate libc;
use std::ffi::{CStr, CString};

use libc::{c_char, c_int};

#[repr(C)]
pub struct scanner_t {
    pub p: *mut c_char,
    pub pe: *mut c_char,
    pub eof: *mut c_char,
    pub ts: *mut c_char,
    pub te: *mut c_char,
    pub token: c_int,
}

extern "C" {
    pub fn init(scanner: *mut scanner_t, str: *const c_char);
    pub fn scan(scanner: *mut scanner_t);
}

fn main() {
    let text =
        CString::new("from(bucket:\"foo\") |> range(start: -1m)").expect("CString::new failed");
    let scanner: *mut scanner_t = &mut scanner_t {
        p: 0 as *mut c_char,
        pe: 0 as *mut c_char,
        eof: 0 as *mut c_char,
        ts: 0 as *mut c_char,
        te: 0 as *mut c_char,
        token: 0,
    };

    unsafe {
        init(scanner, text.as_ptr());
        while (*scanner).p != (*scanner).eof {
            scan(scanner);
            let s = CStr::from_ptr((*scanner).ts);
            println!("scan {} {}", s.to_str().unwrap(), (*scanner).token);
        }
    }
}
