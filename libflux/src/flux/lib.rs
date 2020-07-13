extern crate serde_aux;
extern crate serde_derive;

use core::parser::Parser;
use core::semantic::builtins::builtins;
use core::semantic::check;
use core::semantic::env::Environment;
use core::semantic::flatbuffers::semantic_generated::fbsemantic as fb;
use core::semantic::flatbuffers::types::{build_env, build_type};
use core::semantic::fresh::Fresher;
use core::semantic::nodes::{infer_pkg_types, inject_pkg_types, Package};
use core::semantic::sub::Substitution;
use core::semantic::Importer;

pub use core::ast;
pub use core::formatter;
pub use core::parser;
pub use core::scanner;
pub use core::semantic;
pub use core::*;

use crate::semantic::flatbuffers::semantic_generated::fbsemantic::MonoTypeHolderArgs;
use core::semantic::types::{MonoType, PolyType, TvarKinds};
use std::error;
use std::ffi::*;
use std::os::raw::c_char;

pub fn prelude() -> Option<Environment> {
    let buf = include_bytes!(concat!(env!("OUT_DIR"), "/prelude.data"));
    flatbuffers::get_root::<fb::TypeEnvironment>(buf).into()
}

pub fn imports() -> Option<Environment> {
    let buf = include_bytes!(concat!(env!("OUT_DIR"), "/stdlib.data"));
    flatbuffers::get_root::<fb::TypeEnvironment>(buf).into()
}

pub fn fresher() -> Fresher {
    let buf = include_bytes!(concat!(env!("OUT_DIR"), "/fresher.data"));
    flatbuffers::get_root::<fb::Fresher>(buf).into()
}

/// An error handle designed to allow passing `Error` instances to library
/// consumers across language boundaries.
pub struct ErrorHandle {
    /// A heap-allocated `Error`
    pub err: Box<dyn error::Error>,
}

/// Frees a previously allocated error.
///
/// Note: we use the pattern described here: https://doc.rust-lang.org/std/boxed/index.html#memory-layout
/// wherein a pointer where ownership is being transferred is modeled as a Box, and if it could be
/// null, then it's wrapped in an Option.
#[no_mangle]
pub extern "C" fn flux_free_error(_err: Option<Box<ErrorHandle>>) {}

/// Frees a pointer to characters.
///
/// # Safety
///
/// This function is unsafe because improper use may lead to
/// memory problems. For example, a double-free may occur if the
/// function is called twice on the same raw pointer.
#[no_mangle]
pub unsafe extern "C" fn flux_free_bytes(cstr: *mut c_char) {
    Box::from_raw(cstr);
}

/// A buffer of flux source.
#[repr(C)]
pub struct flux_buffer_t {
    /// A pointer to a byte array.
    pub data: *const u8,
    /// The length of the byte array.
    pub len: usize,
}

/// flux_parse parses a string containing Flux source code into an AST.
///
/// # Safety
///
/// This function is unsafe because it dereferences a raw pointer passed
/// in as a parameter. For example, if that pointer is NULL, undefined behavior
/// could occur.
#[no_mangle]
pub unsafe extern "C" fn flux_parse(
    cfname: *const c_char,
    csrc: *const c_char,
) -> Box<ast::Package> {
    let fname = String::from_utf8(CStr::from_ptr(cfname).to_bytes().to_vec()).unwrap();
    let src = String::from_utf8(CStr::from_ptr(csrc).to_bytes().to_vec()).unwrap();
    let mut p = Parser::new(&src);
    let pkg: ast::Package = p.parse_file(fname).into();
    Box::new(pkg)
}

/// flux_ast_get_error returns the first error in the given AST.
///
/// # Safety
///
/// This funtion is unsafe because it dereferences a raw pointer.
#[no_mangle]
pub unsafe extern "C" fn flux_ast_get_error(
    ast_pkg: *const ast::Package,
) -> Option<Box<ErrorHandle>> {
    let ast_pkg = ast::walk::Node::Package(&*ast_pkg);
    let mut errs = ast::check::check(ast_pkg);
    if !errs.is_empty() {
        let err = Vec::remove(&mut errs, 0);
        Some(Box::new(ErrorHandle { err: Box::new(err) }))
    } else {
        None
    }
}

/// Frees an AST package.
///
/// Note: we use the pattern described here: https://doc.rust-lang.org/std/boxed/index.html#memory-layout
/// wherein a pointer where ownership is being transferred is modeled as a Box, and if it could be
/// null, then it's wrapped in an Option.
#[no_mangle]
pub extern "C" fn flux_free_ast_pkg(_: Option<Box<ast::Package>>) {}

/// # Safety
///
/// This function is unsafe because it dereferences a raw pointer passed
/// in as a parameter. For example, if that pointer is NULL, undefined behavior
/// could occur.
#[no_mangle]
pub unsafe extern "C" fn flux_parse_json(
    cstr: *mut c_char,
    out_pkg: *mut Option<Box<ast::Package>>,
) -> Option<Box<ErrorHandle>> {
    let buf = CStr::from_ptr(cstr).to_bytes(); // Unsafe
    let res: Result<ast::Package, serde_json::error::Error> = serde_json::from_slice(buf);
    match res {
        Ok(pkg) => {
            *out_pkg = Some(Box::new(pkg));
            None
        }
        Err(err) => {
            let errh = ErrorHandle { err: Box::new(err) };
            Some(Box::new(errh))
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
    ast_pkg: *const ast::Package,
    buf: *mut flux_buffer_t,
) -> Option<Box<ErrorHandle>> {
    let ast_pkg = &*ast_pkg;
    let data = match serde_json::to_vec(ast_pkg) {
        Ok(v) => v,
        Err(err) => {
            let errh = ErrorHandle { err: Box::new(err) };
            return Some(Box::new(errh));
        }
    };

    (*buf).len = data.len();
    (*buf).data = Box::into_raw(data.into_boxed_slice()) as *mut u8;
    None
}

/// flux_ast_marshal_fb serializes the given AST package to a flatbuffer.
///
/// # Safety
///
/// This function is unsafe because it takes a dereferences a raw pointer passed
/// in as a parameter. For example, if that pointer is NULL, undefined behavior
/// could occur.
#[no_mangle]
pub unsafe extern "C" fn flux_ast_marshal_fb(
    ast_pkg: *const ast::Package,
    buf: *mut flux_buffer_t,
) -> Option<Box<ErrorHandle>> {
    let ast_pkg = &*ast_pkg;
    let (mut vec, offset) = match ast::flatbuffers::serialize(ast_pkg) {
        Ok(vec_offset) => vec_offset,
        Err(err) => {
            let errh = ErrorHandle {
                err: Box::new(Error::from(err)),
            };
            return Some(Box::new(errh));
        }
    };

    // Note, split_off() does a copy: https://github.com/influxdata/flux/issues/2194
    let data = vec.split_off(offset);
    (*buf).len = data.len();
    (*buf).data = Box::into_raw(data.into_boxed_slice()) as *mut u8;
    None
}

/// Frees a semantic package.
#[no_mangle]
pub extern "C" fn flux_free_semantic_pkg(_: Option<Box<semantic::nodes::Package>>) {}

/// flux_semantic_marshal_fb populates the supplied buffer with a FlatBuffers serialization
/// of the given AST.
///
/// # Safety
///
/// This function is unsafe because it takes a dereferences a raw pointer passed
/// in as a parameter. For example, if that pointer is NULL, undefined behavior
/// could occur.
#[no_mangle]
pub unsafe extern "C" fn flux_semantic_marshal_fb(
    sem_pkg: *const semantic::nodes::Package,
    buf: *mut flux_buffer_t,
) -> Option<Box<ErrorHandle>> {
    let sem_pkg = &*sem_pkg;
    let (mut vec, offset) = match semantic::flatbuffers::serialize(sem_pkg) {
        Ok(vec_offset) => vec_offset,
        Err(err) => {
            let errh = ErrorHandle {
                err: Box::new(Error::from(err)),
            };
            return Some(Box::new(errh));
        }
    };

    // Note, split_off() does a copy: https://github.com/influxdata/flux/issues/2194
    let data = vec.split_off(offset);
    (*buf).len = data.len();
    (*buf).data = Box::into_raw(data.into_boxed_slice()) as *mut u8;
    None
}

/// flux_error_str returns the error message associated with the given error.
///
/// # Safety
///
/// This function is unsafe because it dereferences a raw pointer passed as a
/// parameter
#[no_mangle]
pub unsafe extern "C" fn flux_error_str(errh: *const ErrorHandle) -> CString {
    let errh = &*errh;
    CString::new(format!("{}", errh.err)).unwrap()
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
    out_pkg: *mut ast::Package,
    in_pkg: *mut ast::Package,
) -> Option<Box<ErrorHandle>> {
    // Do not change ownership here so that Go maintains ownership of packages
    let out_pkg = &mut *out_pkg;
    let in_pkg = &mut *in_pkg;

    match merge_packages(out_pkg, in_pkg) {
        None => None,
        Some(err) => {
            let err_handle = ErrorHandle { err: Box::new(err) };
            Some(Box::new(err_handle))
        }
    }
}

/// merge_packages takes an input package and an output package, checks that the package
/// clauses match and merges the files from the input package into the output package. If
/// package clauses fail validation then an option with an Error is returned.
pub fn merge_packages(out_pkg: &mut ast::Package, in_pkg: &mut ast::Package) -> Option<Error> {
    let out_pkg_name = if let Some(pc) = &out_pkg.files[0].package {
        &pc.name.name
    } else {
        DEFAULT_PACKAGE_NAME
    };

    // Check that all input files have a package clause that matches the output package.
    for file in &in_pkg.files {
        match file.package.as_ref() {
            Some(pc) => {
                let in_pkg_name = &pc.name.name;
                if in_pkg_name != out_pkg_name {
                    return Some(Error::from(format!(
                        r#"error at {}: file is in package "{}", but other files are in package "{}""#,
                        pc.base.location, in_pkg_name, out_pkg_name
                    )));
                }
            }
            None => {
                if out_pkg_name != DEFAULT_PACKAGE_NAME {
                    return Some(Error::from(format!(
                        r#"error at {}: file is in default package "{}", but other files are in package "{}""#,
                        file.base.location, DEFAULT_PACKAGE_NAME, out_pkg_name
                    )));
                }
            }
        };
    }
    out_pkg.files.append(&mut in_pkg.files);
    None
}

/// flux_analyze is a C-compatible wrapper around the analyze() function below
///
/// Note that Box<T> is used to indicate we are receiving/returning a C pointer and also
/// transferring ownership.
///
/// # Safety
///
/// This function is unsafe because it dereferences a raw pointer.
#[no_mangle]
#[allow(clippy::boxed_local)]
pub unsafe extern "C" fn flux_analyze(
    ast_pkg: Box<ast::Package>,
    out_sem_pkg: *mut Option<Box<semantic::nodes::Package>>,
) -> Option<Box<ErrorHandle>> {
    match analyze(*ast_pkg) {
        Ok(sem_pkg) => {
            *out_sem_pkg = Some(Box::new(sem_pkg));
            None
        }
        Err(err) => {
            let errh = ErrorHandle { err: Box::new(err) };
            Some(Box::new(errh))
        }
    }
}

#[no_mangle]
pub unsafe extern "C" fn flux_find_var_type(
    ast_pkg: Box<ast::Package>,
    var_name: *const c_char,
    out_type: *mut flux_buffer_t,
) -> Option<Box<ErrorHandle>> {
    let buf = CStr::from_ptr(var_name).to_bytes(); // Unsafe
    let name = String::from_utf8(buf.to_vec()).unwrap();
    find_var_type(*ast_pkg, name).map_or_else(
        |e| {
            let handle = ErrorHandle { err: Box::new(e) };
            Some(Box::new(handle))
        },
        |t| {
            let mut builder = flatbuffers::FlatBufferBuilder::new();
            let (fb_mono_type, typ_type) = build_type(&mut builder, t);
            let fb_mono_type_holder = fb::MonoTypeHolder::create(
                &mut builder,
                &MonoTypeHolderArgs {
                    typ_type,
                    typ: Some(fb_mono_type),
                },
            );
            builder.finish(fb_mono_type_holder, None);
            let (mut vec, offset) = builder.collapse();
            // Note, split_off() does a copy: https://github.com/influxdata/flux/issues/2194
            let data = vec.split_off(offset);
            let out_type = &mut *out_type; // Unsafe
            out_type.len = data.len();
            out_type.data = Box::into_raw(data.into_boxed_slice()) as *mut u8;
            None
        },
    )
}

pub struct SemanticAnalyzer {
    f: Fresher,
    env: Environment,
    imports: Environment,
    importer: Box<dyn Importer>,
}

fn new_semantic_analyzer(pkgpath: &str) -> Result<SemanticAnalyzer, core::Error> {
    let env = match prelude() {
        Some(prelude) => Environment::new(prelude),
        None => return Err(core::Error::from("missing prelude")),
    };
    let imports = match imports() {
        Some(imports) => imports,
        None => return Err(core::Error::from("missing stdlib imports")),
    };
    let mut f = fresher();
    let importer = builtins().importer_for(pkgpath, &mut f);
    Ok(SemanticAnalyzer {
        f,
        env,
        imports,
        importer: Box::new(importer),
    })
}

impl SemanticAnalyzer {
    fn analyze(
        &mut self,
        ast_pkg: ast::Package,
    ) -> Result<core::semantic::nodes::Package, core::Error> {
        let errs = ast::check::check(ast::walk::Node::Package(&ast_pkg));
        if !errs.is_empty() {
            return Err(core::Error::from(format!("{}", &errs[0])));
        }

        let mut sem_pkg = core::semantic::convert::convert_with(ast_pkg, &mut self.f)?;
        check::check(&sem_pkg)?;

        // Clone the environment. The environment may not be returned but we need to maintain
        // a copy of it for this function to be re-entrant.
        let env = self.env.clone();
        let (mut env, sub) = infer_pkg_types(
            &mut sem_pkg,
            env,
            &mut self.f,
            &self.imports,
            &self.importer,
        )?;
        // TODO(jsternberg): This part is hacky and can be improved
        // by refactoring infer file so we can use the internals without
        // infering the file itself.
        // Look at the imports that were part of this semantic package
        // and re-add them to the environment.
        for file in &sem_pkg.files {
            for dec in &file.imports {
                let path = &dec.path.value;
                let name = dec.import_name();

                // A failure should have already happened if any of these
                // imports would have failed.
                let poly = self.imports.lookup(&path).unwrap();
                env.add(name.to_owned(), poly.to_owned());
            }
        }
        self.env = env;
        Ok(inject_pkg_types(sem_pkg, &sub))
    }
}

/// Create a new semantic analyzer.
///
/// # Safety
///
/// Ths function is unsafe because it dereferences a raw pointer.
#[no_mangle]
pub unsafe extern "C" fn flux_new_semantic_analyzer(
    cstr: *mut c_char,
) -> Box<Result<SemanticAnalyzer, core::Error>> {
    let buf = CStr::from_ptr(cstr).to_bytes(); // Unsafe
    let s = String::from_utf8(buf.to_vec()).unwrap();
    Box::new(new_semantic_analyzer(&s))
}

/// Free a previously allocated semantic analyzer
#[no_mangle]
pub extern "C" fn flux_free_semantic_analyzer(
    _: Option<Box<Result<SemanticAnalyzer, core::Error>>>,
) {
}

/// # Safety
///
/// Ths function is unsafe because it dereferences a raw pointer.
#[no_mangle]
#[allow(clippy::boxed_local)]
pub unsafe extern "C" fn flux_analyze_with(
    analyzer: *mut Result<SemanticAnalyzer, core::Error>,
    ast_pkg: Box<ast::Package>,
    out_sem_pkg: *mut Option<Box<semantic::nodes::Package>>,
) -> Option<Box<ErrorHandle>> {
    let ast_pkg = *ast_pkg;
    let analyzer = match &mut *analyzer {
        Ok(a) => a,
        Err(err) => {
            let errh = ErrorHandle {
                err: Box::new(err.to_owned()),
            };
            return Some(Box::new(errh));
        }
    };

    let sem_pkg = Box::new(match analyzer.analyze(ast_pkg) {
        Ok(sem_pkg) => sem_pkg,
        Err(err) => {
            let errh = ErrorHandle { err: Box::new(err) };
            return Some(Box::new(errh));
        }
    });

    *out_sem_pkg = Some(sem_pkg);
    None
}

/// analyze consumes the given AST package and returns a semantic package
/// that has been type-inferred.  This function is aware of the standard library
/// and prelude.
pub fn analyze(ast_pkg: ast::Package) -> Result<Package, Error> {
    let (sem_pkg, _, sub) = infer_with_env(ast_pkg, fresher(), None)?;
    Ok(inject_pkg_types(sem_pkg, &sub))
}

/// infer_with_env consumes the given AST package, inject the type bindings from the given
/// type environment, and returns a semantic package that has not been type-inferred and an
/// inferred type environment and substitution.
/// This function is aware of the standard library and prelude.
pub fn infer_with_env(
    ast_pkg: ast::Package,
    mut f: Fresher,
    env: Option<Environment>,
) -> Result<(Package, Environment, Substitution), Error> {
    // First check to see if there are any errors in the AST.
    let errs = ast::check::check(ast::walk::Node::Package(&ast_pkg));
    if !errs.is_empty() {
        return Err(core::Error::from(format!("{}", &errs[0])));
    }

    let pkgpath = ast_pkg.path.clone();
    let mut sem_pkg = core::semantic::convert::convert_with(ast_pkg, &mut f)?;

    check::check(&sem_pkg)?;

    let mut prelude = match prelude() {
        Some(prelude) => Environment::new(prelude),
        None => return Err(core::Error::from("missing prelude")),
    };
    env.map(|e| prelude.copy_bindings_from(&e));
    let imports = match imports() {
        Some(imports) => imports,
        None => return Err(core::Error::from("missing stdlib imports")),
    };
    let builtin_importer = builtins().importer_for(&pkgpath, &mut f);

    let (env, sub) = infer_pkg_types(&mut sem_pkg, prelude, &mut f, &imports, &builtin_importer)?;
    Ok((sem_pkg, env, sub))
}

pub fn find_var_type(ast_pkg: ast::Package, var_name: String) -> Result<MonoType, Error> {
    let mut f = fresher();
    let tvar = f.fresh();
    let mut env = Environment::empty(true);
    env.add(
        var_name.clone(),
        PolyType {
            vars: Vec::new(),
            cons: TvarKinds::new(),
            expr: MonoType::Var(tvar.clone()),
        },
    );
    infer_with_env(ast_pkg, f, Some(env))
        .map(|(_, env, _)| env.lookup(var_name.as_str()).unwrap().expr.clone())
}

/// # Safety
///
/// This function is unsafe because it dereferences a raw pointer.
#[no_mangle]
pub unsafe extern "C" fn flux_get_env_stdlib(buf: *mut flux_buffer_t) {
    let env = imports().unwrap();
    let mut builder = flatbuffers::FlatBufferBuilder::new();
    let fb_type_env = build_env(&mut builder, env);

    builder.finish(fb_type_env, None);
    let (mut vec, offset) = builder.collapse();

    // Note, split_off() does a copy: https://github.com/influxdata/flux/issues/2194
    let data = vec.split_off(offset);
    let buf = &mut *buf; // Unsafe
    buf.len = data.len();
    buf.data = Box::into_raw(data.into_boxed_slice()) as *mut u8;
}

#[cfg(test)]
mod tests {
    use crate::{analyze, find_var_type, flux_ast_get_error, merge_packages};
    use core::parser::Parser;
    use core::semantic::convert::convert_file;
    use core::semantic::env::Environment;
    use core::semantic::nodes::infer_file;
    use core::{ast, semantic};

    #[test]
    fn test_find_var_type() {
        let source = r#"
vint = v.int + 2
f = (v) => v.shadow
h = () => v.wow
g = () => v.sweet
x = g()
vstr = v.str + "hello"
"#;
        let mut p = Parser::new(&source);
        let pkg: ast::Package = p.parse_file("".to_string()).into();
        let ty = find_var_type(pkg, "v".to_string()).expect("should be able to find var type");
        println!("{}", ty);
    }

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
    fn ok_merge_one_default_pkg() {
        // Make sure we can merge one file with default "main"
        // and on explicit
        let has_clause_script = "package main\nx = 32";
        let no_clause_script = "y = 32";
        let has_clause_file = crate::parser::parse_string("has_clause.flux", has_clause_script);
        let no_clause_file = crate::parser::parse_string("no_clause.flux", no_clause_script);
        {
            let mut out_pkg: ast::Package = has_clause_file.clone().into();
            let mut in_pkg: ast::Package = no_clause_file.clone().into();
            if let Some(e) = merge_packages(&mut out_pkg, &mut in_pkg) {
                panic!(e);
            }
            let got = out_pkg.files;
            let want = vec![has_clause_file.clone(), no_clause_file.clone()];
            assert_eq!(want, got);
        }
        {
            // Same as previous test, but reverse order
            let mut out_pkg: ast::Package = no_clause_file.clone().into();
            let mut in_pkg: ast::Package = has_clause_file.clone().into();
            if let Some(e) = merge_packages(&mut out_pkg, &mut in_pkg) {
                panic!(e);
            }
            let got = out_pkg.files;
            let want = vec![no_clause_file.clone(), has_clause_file.clone()];
            assert_eq!(want, got);
        }
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

        let in_file = crate::parser::parse_string("test_in.flux", in_script);
        let out_file = crate::parser::parse_string("test_out.flux", out_script);
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
        let want_err = r#"error at test_in.flux@1:1-1:12: file is in package "foo", but other files are in package "main""#;
        assert_eq!(got_err.to_string(), want_err);
    }

    #[test]
    fn err_no_in_pkg_clause() {
        let in_script = "a = 1000\n";
        let out_script = "package foo\nb = 100\n";

        let in_file = crate::parser::parse_string("test_in.flux", in_script);
        let out_file = crate::parser::parse_string("test_out.flux", out_script);
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
        let want_err = r#"error at test_in.flux@1:1-1:9: file is in default package "main", but other files are in package "foo""#;
        assert_eq!(got_err.to_string(), want_err);
    }

    #[test]
    fn ok_no_pkg_clauses() {
        let in_script = "a = 100\n";
        let out_script = "b = a * a\n";
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
        let result = merge_packages(&mut out_pkg, &mut in_pkg);
        assert!(result.is_none());
        assert_eq!(2, out_pkg.files.len());
    }

    #[test]
    fn test_ast_get_error() {
        let ast = crate::parser::parse_string("test", "x = 3 + / 10 - \"");
        let ast = Box::into_raw(Box::new(ast.into()));
        let errh = unsafe { flux_ast_get_error(ast) };
        assert_eq!(
            "error at test@1:9-1:10: invalid expression: invalid token for primary expression: DIV",
            format!("{}", errh.unwrap().err)
        );
    }

    #[test]
    fn deserialize_and_infer() {
        let prelude = Environment::new(super::prelude().unwrap());
        let imports = super::imports().unwrap();

        let src = r#"
            x = from(bucket: "b")
                |> filter(fn: (r) => r.region == "west")
                |> map(fn: (r) => ({r with _value: r._value + r._value}))
        "#;

        let ast = core::parser::parse_string("main.flux", src);
        let mut f = super::fresher();

        let mut file = convert_file(ast, &mut f).unwrap();
        let (got, _) = infer_file(&mut file, prelude, &mut f, &imports, &None).unwrap();

        // TODO(algow): re-introduce equality constraints for binary comparison operators
        // https://github.com/influxdata/flux/issues/2466
        let want = semantic::parser::parse(
            r#"forall [t0, t1, t2] where t0: Addable, t1: Equatable [{
                _value: t0
                    | _value: t0
                    | _time: time
                    | _measurement: string
                    | _field: string
                    | region: t1
                    | t2
                    }]
            "#,
        )
        .unwrap();

        assert_eq!(want, got.lookup("x").expect("'x' not found").clone());
    }

    #[test]
    fn infer_union() {
        let prelude = Environment::new(super::prelude().unwrap());
        let imports = super::imports().unwrap();

        let src = r#"
            a = from(bucket: "b")
                |> filter(fn: (r) => r.A == "A")
            b = from(bucket: "b")
                |> filter(fn: (r) => r.B == "B")
            c = union(tables: [a, b])
        "#;

        let ast = core::parser::parse_string("main.flux", src);
        let mut f = super::fresher();

        let mut file = convert_file(ast, &mut f).unwrap();
        let (got, _) = infer_file(&mut file, prelude, &mut f, &imports, &None).unwrap();

        // TODO(algow): re-introduce equality constraints for binary comparison operators
        // https://github.com/influxdata/flux/issues/2466
        let want_a = semantic::parser::parse(
            r#"forall [t0, t1, t3] where t1: Equatable [{
                _value: t0
                    | A: t1
                    | _time: time
                    | _measurement: string
                    | _field: string
                    | t3
                    }]
            "#,
        )
        .unwrap();
        let want_b = semantic::parser::parse(
            r#"forall [t0, t1, t3] where t1: Equatable [{
                _value: t0
                    | B: t1
                    | _time: time
                    | _measurement: string
                    | _field: string
                    | t3
                    }]
            "#,
        )
        .unwrap();
        let want_c = semantic::parser::parse(
            r#"forall [t0, t1, t2, t3] where t1: Equatable, t2: Equatable [{
                _value: t0
                    | A: t1
                    | B: t2
                    | _time: time
                    | _measurement: string
                    | _field: string
                    | t3
                    }]
            "#,
        )
        .unwrap();

        assert_eq!(want_a, got.lookup("a").expect("'a' not found").clone());
        assert_eq!(want_b, got.lookup("b").expect("'b' not found").clone());
        assert_eq!(want_c, got.lookup("c").expect("'c' not found").clone());
    }

    #[test]
    fn analyze_error() {
        let ast: ast::Package = core::parser::parse_string("", "x = ()").into();
        match analyze(ast) {
            Ok(_) => panic!("expected an error, got none"),
            Err(e) => {
                let want = "error at @1:5-1:7: expected ARROW, got EOF";
                let got = format!("{}", e);
                if want != got {
                    panic!(r#"expected error "{}", got "{}""#, want, got)
                }
            }
        }
    }
}
