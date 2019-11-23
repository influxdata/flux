#![cfg_attr(feature = "strict", deny(warnings))]
// When the unsafe interfaces are documented, these ignored clippy lints should
// be re-enabled and adjusted.
#![allow(clippy::missing_safety_doc, clippy::unknown_clippy_lints)]

extern crate chrono;
#[macro_use]
extern crate serde_derive;
extern crate serde_aux;

pub mod ast;
pub mod parser;
pub mod scanner;
pub mod semantic;

use std::error::Error;
use std::ffi::*;
use std::os::raw::{c_char, c_void};

use parser::Parser;

pub const DEFAULT_PACKAGE_NAME: &str = "main";

#[allow(non_camel_case_types)]
pub mod ctypes {
    include!(concat!(env!("OUT_DIR"), "/ctypes.rs"));
}
use ctypes::*;

struct ErrorHandle {
    err: Box<dyn Error>,
}

#[repr(C)]
pub struct flux_buffer_t {
    pub data: *const u8,
    pub len: usize,
}

#[no_mangle]
pub unsafe extern "C" fn flux_parse(cstr: *mut c_char) -> *mut flux_ast_t {
    let buf = CStr::from_ptr(cstr).to_bytes();  // Unsafe
    let s = String::from_utf8(buf.to_vec()).unwrap();
    let mut p = Parser::new(&s);
    let file = p.parse_file(String::from(""));
    Box::into_raw(Box::new(file)) as *mut flux_ast_t
}

#[no_mangle]
pub unsafe extern "C" fn flux_parse_fb(src_ptr: *const c_char) -> *mut flux_buffer_t {
    let src_bytes = CStr::from_ptr(src_ptr).to_bytes();  // Unsafe
    let src = String::from_utf8(src_bytes.to_vec()).unwrap();
    let mut p = Parser::new(&src);
    let file = p.parse_file(String::from(""));
    let package_name: String;
    match &file.package {
        Some(p) => {
            package_name = p.name.name.clone();
        }
        _ => {
            package_name = DEFAULT_PACKAGE_NAME.to_string();
        }
    }
    let pkg = ast::Package {
        base: ast::BaseNode {
            ..ast::BaseNode::default()
        },
        path: String::from(""),
        package: package_name,
        files: vec![file],
    };
    let r = ast::flatbuffers::serialize(&pkg);
    match r {
        Ok((vec, offset)) => {
            let data = &vec[offset..];
            Box::into_raw(Box::new(flux_buffer_t {
                data: data.as_ptr(),
                len: data.len(),
            }))
        }
        Err(_) => 1 as *mut flux_buffer_t,
    }
}

#[no_mangle]
pub unsafe extern "C" fn flux_ast_marshal_json(
    ast: *mut flux_ast_t,
    buf: *mut flux_buffer_t,
) -> *mut flux_error_t {
    let self_ = &*(ast as *mut ast::File) as &ast::File;  // Unsafe
    let data = match serde_json::to_vec(self_) {
        Ok(v) => v,
        Err(err) => {
            let errh = ErrorHandle { err: Box::new(err) };
            return Box::into_raw(Box::new(errh)) as *mut flux_error_t;
        }
    };

    let buffer = &mut *buf;  // Unsafe
    buffer.len = data.len();
    buffer.data = Box::into_raw(data.into_boxed_slice()) as *mut u8;
    std::ptr::null_mut()
}

#[no_mangle]
pub unsafe extern "C" fn flux_error_str(err: *mut flux_error_t) -> *mut c_char {
    let e = &*(err as *mut ErrorHandle);  // Unsafe
    let s = CString::new(e.err.description()).unwrap();
    s.into_raw()
}

#[no_mangle]
pub unsafe extern "C" fn flux_free(err: *mut c_void) {
    let _ = Box::from_raw(err);
}
