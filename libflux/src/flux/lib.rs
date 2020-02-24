#![cfg_attr(feature = "strict", deny(warnings, missing_docs))]
#![allow(clippy::unknown_clippy_lints)]
//! The flux crate handles the parsing and semantic analysis of flux source
//! code.
extern crate chrono;
#[macro_use]
extern crate serde_derive;
extern crate serde_aux;

pub mod ast;
pub mod formatter;
pub mod parser;
pub mod scanner;
pub mod semantic;

use std::error;
use std::ffi::*;
use std::fmt;
use std::os::raw::{c_char, c_void};

use parser::Parser;

pub use ast::DEFAULT_PACKAGE_NAME;

#[allow(non_camel_case_types, missing_docs)]
pub mod ctypes {
    include!(concat!(env!("OUT_DIR"), "/ctypes.rs"));
}
use ctypes::*;

/// An error handle designed to allow passing `Error` instances to library
/// consumers across language boundaries.
pub struct ErrorHandle {
    /// A heap-allocated `Error`
    pub err: Box<dyn error::Error>,
}

/// An error that can occur due to problems in ast generation or semantic
/// analysis.
#[derive(Debug, Clone)]
pub struct Error {
    msg: String,
}

impl fmt::Display for Error {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        f.write_str(&self.msg)
    }
}

impl error::Error for Error {
    fn source(&self) -> Option<&(dyn error::Error + 'static)> {
        None
    }
}

impl From<String> for Error {
    fn from(msg: String) -> Self {
        Error { msg }
    }
}

impl From<&str> for Error {
    fn from(msg: &str) -> Self {
        Error {
            msg: String::from(msg),
        }
    }
}

impl From<semantic::nodes::Error> for Error {
    fn from(sn_err: semantic::nodes::Error) -> Self {
        Error { msg: sn_err.msg }
    }
}

impl From<semantic::check::Error> for Error {
    fn from(err: semantic::check::Error) -> Self {
        Error {
            msg: format!("{}", err),
        }
    }
}

/// A buffer of flux source.
#[repr(C)]
pub struct flux_buffer_t {
    /// A pointer to a byte array.
    pub data: *const u8,
    /// The length of the byte array.
    pub len: usize,
}

/// # Safety
///
/// This function is unsafe because it dereferences a raw pointer passed
/// in as a parameter. For example, if that pointer is NULL, undefined behavior
/// could occur.
#[no_mangle]
pub unsafe extern "C" fn flux_parse(cstr: *mut c_char) -> *mut flux_ast_pkg_t {
    let buf = CStr::from_ptr(cstr).to_bytes(); // Unsafe
    let s = String::from_utf8(buf.to_vec()).unwrap();
    let mut p = Parser::new(&s);
    let pkg: ast::Package = p.parse_file(String::from("")).into();
    Box::into_raw(Box::new(pkg)) as *mut flux_ast_pkg_t
}

/// # Safety
///
/// This function is unsafe because it dereferences a raw pointer passed
/// in as a parameter. For example, if that pointer is NULL, undefined behavior
/// could occur.
#[no_mangle]
pub unsafe extern "C" fn flux_parse_json(
    cstr: *mut c_char,
    out_pkg: *mut *const flux_ast_pkg_t,
) -> *mut flux_error_t {
    let buf = CStr::from_ptr(cstr).to_bytes(); // Unsafe
    let res: Result<ast::Package, serde_json::error::Error> = serde_json::from_slice(buf);
    match res {
        Ok(pkg) => {
            let pkg = Box::into_raw(Box::new(pkg)) as *const flux_ast_pkg_t;
            *out_pkg = pkg;
            std::ptr::null_mut()
        }
        Err(err) => {
            let errh = ErrorHandle { err: Box::new(err) };
            Box::into_raw(Box::new(errh)) as *mut flux_error_t
        }
    }
}

/// # Safety
///
/// This function is unsafe because it dereferences raw pointers passed
/// in as parameters. For example, if that pointer is NULL, undefined behavior
/// could occur.
#[no_mangle]
pub unsafe extern "C" fn flux_ast_marshal_json(
    ast_pkg: *mut flux_ast_pkg_t,
    buf: *mut flux_buffer_t,
) -> *mut flux_error_t {
    let ast_pkg = &*(ast_pkg as *mut ast::Package) as &ast::Package; // Unsafe
    let data = match serde_json::to_vec(ast_pkg) {
        Ok(v) => v,
        Err(err) => {
            let errh = ErrorHandle { err: Box::new(err) };
            return Box::into_raw(Box::new(errh)) as *mut flux_error_t;
        }
    };

    let buffer = &mut *buf; // Unsafe
    buffer.len = data.len();
    buffer.data = Box::into_raw(data.into_boxed_slice()) as *mut u8;
    std::ptr::null_mut()
}

/// # Safety
///
/// This function is unsafe because it takes a dereferences a raw pointer passed
/// in as a parameter. For example, if that pointer is NULL, undefined behavior
/// could occur.
#[no_mangle]
pub unsafe extern "C" fn flux_ast_marshal_fb(
    ast: *mut flux_ast_pkg_t,
    buf: *mut flux_buffer_t,
) -> *mut flux_error_t {
    let pkg = &*(ast as *mut ast::Package) as &ast::Package; // Unsafe
    let (mut vec, offset) = match ast::flatbuffers::serialize(&pkg) {
        Ok(vec_offset) => vec_offset,
        Err(err) => {
            let err: Error = err.into();
            let errh = ErrorHandle { err: Box::new(err) };
            return Box::into_raw(Box::new(errh)) as *mut flux_error_t;
        }
    };

    // Note, split_off() does a copy: https://github.com/influxdata/flux/issues/2194
    let data = vec.split_off(offset);
    let buffer = &mut *buf; // Unsafe
    buffer.len = data.len();
    buffer.data = Box::into_raw(data.into_boxed_slice()) as *mut u8;
    std::ptr::null_mut()
}

/// # Safety
///
/// This function is unsafe because it takes a dereferences a raw pointer passed
/// in as a parameter. For example, if that pointer is NULL, undefined behavior
/// could occur.
#[no_mangle]
pub unsafe extern "C" fn flux_semantic_marshal_fb(
    ast: *mut flux_semantic_pkg_t,
    buf: *mut flux_buffer_t,
) -> *mut flux_error_t {
    let pkg = &*(ast as *mut semantic::nodes::Package) as &semantic::nodes::Package; // Unsafe
    let (mut vec, offset) = match semantic::flatbuffers::serialize(&pkg) {
        Ok(vec_offset) => vec_offset,
        Err(err) => {
            let err: Error = err.into();
            let errh = ErrorHandle { err: Box::new(err) };
            return Box::into_raw(Box::new(errh)) as *mut flux_error_t;
        }
    };

    // Note, split_off() does a copy: https://github.com/influxdata/flux/issues/2194
    let data = vec.split_off(offset);
    let buffer = &mut *buf; // Unsafe
    buffer.len = data.len();
    buffer.data = Box::into_raw(data.into_boxed_slice()) as *mut u8;
    std::ptr::null_mut()
}

/// # Safety
///
/// This function is unsafe because it dereferences a raw pointer passed as a
/// parameter
#[no_mangle]
pub unsafe extern "C" fn flux_error_str(err: *mut flux_error_t) -> *mut c_char {
    let e = &*(err as *mut ErrorHandle); // Unsafe
    let s = CString::new(format!("{}", e.err)).unwrap();
    s.into_raw()
}

/// # Safety
///
/// This function is unsafe because improper use may lead to memory problems.
/// For example, a double-free may occur if the function is called twice on
/// the same raw pointer.
#[no_mangle]
pub unsafe extern "C" fn flux_free(err: *mut c_void) {
    Box::from_raw(err);
}

/// # Safety
///
/// This function is unsafe because it dereferences a raw pointer passed as a
/// parameter
///
/// flux_merge_ast_pkg_files merges the files of a given input ast::Package into the file
/// vector of an output ast::Package.
#[no_mangle]
pub unsafe extern "C" fn flux_merge_ast_pkgs(
    out_pkg: *mut flux_ast_pkg_t,
    in_pkg: *mut flux_ast_pkg_t,
) -> *mut flux_error_t {
    // Do not change ownership here so that Go maintains ownership of packages
    let out_pkg = &mut *(out_pkg as *mut ast::Package);
    let in_pkg = &mut *(in_pkg as *mut ast::Package);

    match merge_packages(out_pkg, in_pkg) {
        None => std::ptr::null_mut(),
        Some(err) => {
            let err_handle = ErrorHandle { err: Box::new(err) };
            Box::into_raw(Box::new(err_handle)) as *mut flux_error_t
        }
    }
}

/// merge_packages takes an input package and an output package, checks that the package
/// clauses match and merges the files from the input package into the output package. If
/// package clauses fail validation then an option with an Error is returned.
pub fn merge_packages(out_pkg: &mut ast::Package, in_pkg: &mut ast::Package) -> Option<Error> {
    let pkg_clause = match &out_pkg.files[0].package {
        Some(clause) => &clause.name.name,
        None => return Some(Error::from("output package does not have a package clause")),
    };

    // Check that all input files have a package clause that matches the output package.
    for file in &in_pkg.files {
        let file_clause = match &file.package {
            Some(clause) => &clause.name.name,
            None => return Some(Error::from("current file does not have a package clause")),
        };

        if pkg_clause != file_clause {
            return Some(Error::from(format!(
                "file's package clause: {} does not match package output package clause: {}",
                file_clause, pkg_clause
            )));
        }
    }
    out_pkg.files.append(&mut in_pkg.files);
    None
}

#[cfg(test)]
mod tests {
    use crate::{ast, merge_packages};

    #[test]
    fn ok_merge_multi_file() {
        let in_script = "package foo\na = 1\n";
        let out_script = "package foo\nb = 2\n";

        let in_file = crate::parser::parse_string("test", in_script);
        let out_file = crate::parser::parse_string("test", out_script);
        let mut in_pkg = ast::Package {
            base: Default::default(),
            path: "./test".to_string(),
            package: "foo".to_string(),
            files: vec![in_file.clone()],
        };
        let mut out_pkg = ast::Package {
            base: Default::default(),
            path: "./test".to_string(),
            package: "foo".to_string(),
            files: vec![out_file.clone()],
        };
        merge_packages(&mut out_pkg, &mut in_pkg);
        let got = out_pkg.files;
        let want = vec![out_file, in_file];
        assert_eq!(want, got);
    }

    #[test]
    fn ok_no_in_pkg() {
        let out_script = "package foo\nb = 2\n";

        let out_file = crate::parser::parse_string("test", out_script);
        let mut in_pkg = ast::Package {
            base: Default::default(),
            path: "./test".to_string(),
            package: "foo".to_string(),
            files: vec![],
        };
        let mut out_pkg = ast::Package {
            base: Default::default(),
            path: "./test".to_string(),
            package: "foo".to_string(),
            files: vec![out_file.clone()],
        };
        merge_packages(&mut out_pkg, &mut in_pkg);
        let got = out_pkg.files;
        let want = vec![out_file];
        assert_eq!(want, got);
    }

    #[test]
    fn err_no_out_pkg_clause() {
        let in_script = "package foo\na = 1\n";
        let out_script = "";

        let in_file = crate::parser::parse_string("test", in_script);
        let out_file = crate::parser::parse_string("test", out_script);
        let mut in_pkg = ast::Package {
            base: Default::default(),
            path: "./test".to_string(),
            package: "foo".to_string(),
            files: vec![in_file.clone()],
        };
        let mut out_pkg = ast::Package {
            base: Default::default(),
            path: "./test".to_string(),
            package: "foo".to_string(),
            files: vec![out_file.clone()],
        };
        let got_err = merge_packages(&mut out_pkg, &mut in_pkg).unwrap().msg;
        let want_err = "output package does not have a package clause";
        assert_eq!(want_err, got_err.to_string());
    }

    #[test]
    fn err_no_in_pkg_clause() {
        let in_script = "";
        let out_script = "package foo\nb = 100\n";

        let in_file = crate::parser::parse_string("test", in_script);
        let out_file = crate::parser::parse_string("test", out_script);
        let mut in_pkg = ast::Package {
            base: Default::default(),
            path: "./test".to_string(),
            package: "foo".to_string(),
            files: vec![in_file.clone()],
        };
        let mut out_pkg = ast::Package {
            base: Default::default(),
            path: "./test".to_string(),
            package: "foo".to_string(),
            files: vec![out_file.clone()],
        };
        let got_err = merge_packages(&mut out_pkg, &mut in_pkg).unwrap().msg;
        let want_err = "current file does not have a package clause";
        assert_eq!(want_err, got_err.to_string());
    }
}
