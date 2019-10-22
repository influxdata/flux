extern crate chrono;
#[macro_use]
extern crate serde_derive;
extern crate serde_aux;

pub mod ast;
pub mod parser;
pub mod scanner;
pub mod semantic;

use std::ffi::*;
use std::error::Error;
use std::os::raw::{c_void, c_char};

use parser::Parser;

pub mod ctypes {
    include!(concat!(env!("OUT_DIR"), "/ctypes.rs"));
}
use ctypes::*;

struct ErrorHandle {
    err: Box<dyn Error>,
}

#[no_mangle]
pub extern "C" fn flux_parse(cstr: *mut c_char) -> *mut flux_ast_t {
    let buf = unsafe {
        CStr::from_ptr(cstr).to_bytes()
    };
    let s = String::from_utf8(buf.to_vec()).unwrap();
    let mut p = Parser::new(&s);
    let file = p.parse_file(String::from(""));
    return Box::into_raw(Box::new(file)) as *mut flux_ast_t;
}

#[no_mangle]
pub extern "C" fn flux_ast_marshal_json(ast: *mut flux_ast_t, buf: *mut flux_buffer_t) -> *mut flux_error_t {
    let self_ = unsafe { &*(ast as *mut ast::File) } as &ast::File;
    let data = match serde_json::to_vec(self_) {
        Ok(v) => v,
        Err(err) => {
            let errh = ErrorHandle{ err: Box::new(err) };
            return Box::into_raw(Box::new(errh)) as *mut flux_error_t
        },
    };

    let buffer = unsafe { &mut *buf };
    buffer.len = data.len();
    buffer.data = Box::into_raw(data.into_boxed_slice()) as *mut c_void;
    return std::ptr::null_mut();
}

#[no_mangle]
pub extern "C" fn flux_error_str(err: *mut flux_error_t) -> *mut c_char {
    let e = unsafe { &*(err as *mut ErrorHandle) };
    let s = CString::new(e.err.description()).unwrap();
    return s.into_raw();
}

#[no_mangle]
pub extern "C" fn flux_free(err: *mut c_void) {
    unsafe {
        let _ = Box::from_raw(err);
    }
}
