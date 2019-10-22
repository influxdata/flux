extern crate chrono;
#[macro_use]
extern crate serde_derive;
extern crate serde_aux;

pub mod ast;
pub mod parser;
pub mod scanner;
pub mod semantic;

use std::ffi::*;
use std::os::raw::c_char;

use parser::Parser;

#[no_mangle]
pub extern "C" fn flux_parse(cstr: *mut c_char) -> *mut ast::File {
    let buf = unsafe {
        CStr::from_ptr(cstr).to_bytes()
    };
    let s = String::from_utf8(buf.to_vec()).unwrap();
    let mut p = Parser::new(&s);
    let file = p.parse_file(String::from(""));
    return Box::into_raw(Box::new(file));
}

#[no_mangle]
pub extern "C" fn flux_ast_free(ast: *mut ast::File) {
    unsafe {
        let _ = Box::from_raw(ast);
    }
}
