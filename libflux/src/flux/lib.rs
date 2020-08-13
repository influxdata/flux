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
use core::semantic::types::{MonoType, PolyType, Tvar, TvarKinds};
use std::error;
use std::ffi::*;
use std::os::raw::c_char;
use wasm_bindgen::prelude::*;

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

/// flux_find_var_type() is a C-compatible wrapper around the find_var_type() function below.
/// Note that Box<T> is used to indicate we are receiving/returning a C pointer and also
/// transferring ownership.
///
/// # Safety
///
/// This function is unsafe because it dereferences a raw pointer.
#[no_mangle]
#[allow(clippy::boxed_local)]
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
/// type environment, and returns a semantic package that has not been type-injected and an
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
    if let Some(e) = env {
        prelude.copy_bindings_from(&e);
    }
    let imports = match imports() {
        Some(imports) => imports,
        None => return Err(core::Error::from("missing stdlib imports")),
    };
    let builtin_importer = builtins().importer_for(&pkgpath, &mut f);

    let (env, sub) = infer_pkg_types(&mut sem_pkg, prelude, &mut f, &imports, &builtin_importer)?;
    Ok((sem_pkg, env, sub))
}

/// Given a Flux source and a variable name, find out the type of that variable in the Flux source code.
/// A type variable will be automatically generated and injected into the type environment that
/// will be used in semantic analysis. The Flux source code itself should not contain any definition
/// for that variable.
/// This version of find_var_type is aware of the prelude and builtins.
pub fn find_var_type(ast_pkg: ast::Package, var_name: String) -> Result<MonoType, Error> {
    let mut f = fresher();
    let tvar = f.fresh();
    let mut env = Environment::empty(true);
    env.add(
        var_name.clone(),
        PolyType {
            vars: Vec::new(),
            cons: TvarKinds::new(),
            expr: MonoType::Var(tvar),
        },
    );
    infer_with_env(ast_pkg, f, Some(env))
        .map(|(_, env, _)| env.lookup(var_name.as_str()).unwrap().expr.clone())
}

/// wasm version of the flux_find_var_type() API. Instead of returning a flat buffer that contains
/// the MonoType, it returns a JsValue。
#[wasm_bindgen]
pub fn wasm_find_var_type(source: &str, file_name: &str, var_name: &str) -> JsValue {
    let mut p = Parser::new(source);
    let pkg: ast::Package = p.parse_file(file_name.to_string()).into();
    let ty = find_var_type(pkg, var_name.to_string()).unwrap_or(MonoType::Var(Tvar(0)));
    JsValue::from_serde(&ty).unwrap()
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
    use crate::parser;
    use crate::{analyze, find_var_type, flux_ast_get_error, merge_packages};
    use core::ast;
    use core::ast::get_err_type_expression;
    use core::parser::Parser;
    use core::semantic::convert::convert_file;
    use core::semantic::convert::convert_polytype;
    use core::semantic::env::Environment;
    use core::semantic::fresh::Fresher;
    use core::semantic::nodes::infer_file;
    use core::semantic::types::{MonoType, Property, Record, Tvar, TvarMap};

    pub struct MonoTypeNormalizer {
        tv_map: TvarMap,
        f: Fresher,
    }

    impl MonoTypeNormalizer {
        pub fn new() -> Self {
            Self {
                tv_map: TvarMap::new(),
                f: Fresher::default(),
            }
        }

        pub fn normalize(&mut self, t: &mut MonoType) {
            match t {
                MonoType::Var(tv) => {
                    // This is to avoid using self directly inside a closure,
                    // otherwise it will be captured by that closure and the compiler
                    // will complain that closure requires unique access to `self`
                    let f = &mut self.f;
                    let v = self.tv_map.entry(*tv).or_insert_with(|| f.fresh());
                    *tv = *v;
                }
                MonoType::Arr(arr) => {
                    self.normalize(&mut arr.as_mut().0);
                }
                MonoType::Record(r) => {
                    if let Record::Extension { head, tail } = r.as_mut() {
                        self.normalize(&mut head.v);
                        self.normalize(tail);
                    }
                }
                MonoType::Fun(f) => {
                    for (_, mut v) in f.req.iter_mut() {
                        self.normalize(&mut v);
                    }
                    for (_, mut v) in f.opt.iter_mut() {
                        self.normalize(&mut v);
                    }
                    if let Some(p) = &mut f.pipe {
                        self.normalize(&mut p.v);
                    }
                    self.normalize(&mut f.retn);
                }
                _ => {}
            }
        }
    }

    #[test]
    fn monotype_normalizer() {
        let mut ty = MonoType::Record(Box::new(Record::Extension {
            head: Property {
                k: "a".to_string(),
                v: MonoType::Var(Tvar(4949)),
            },
            tail: MonoType::Record(Box::new(Record::Extension {
                head: Property {
                    k: "b".to_string(),
                    v: MonoType::Var(Tvar(4949)),
                },
                tail: MonoType::Record(Box::new(Record::Extension {
                    head: Property {
                        k: "e".to_string(),
                        v: MonoType::Var(Tvar(4957)),
                    },
                    tail: MonoType::Record(Box::new(Record::Extension {
                        head: Property {
                            k: "f".to_string(),
                            v: MonoType::Var(Tvar(4957)),
                        },
                        tail: MonoType::Record(Box::new(Record::Extension {
                            head: Property {
                                k: "g".to_string(),
                                v: MonoType::Var(Tvar(4957)),
                            },
                            tail: MonoType::Var(Tvar(4972)),
                        })),
                    })),
                })),
            })),
        }));
        assert_eq!(
            format!("{}", ty),
            "{a:t4949 | b:t4949 | e:t4957 | f:t4957 | g:t4957 | t4972}"
        );
        let mut v = MonoTypeNormalizer::new();
        v.normalize(&mut ty);
        assert_eq!(format!("{}", ty), "{a:t0 | b:t0 | e:t1 | f:t1 | g:t1 | t2}");
    }

    #[test]
    fn find_var_ref() {
        let source = r#"
vint = v.int + 2
f = (v) => v.shadow
g = () => v.sweet
x = g()
vstr = v.str + "hello"
"#;
        let mut p = Parser::new(&source);
        let pkg: ast::Package = p.parse_file("".to_string()).into();
        let mut t = find_var_type(pkg, "v".into()).expect("Should be able to get a MonoType.");
        let mut v = MonoTypeNormalizer::new();
        v.normalize(&mut t);
        assert_eq!(format!("{}", t), "{int:int | sweet:t0 | str:string | t1}");

        assert_eq!(
            serde_json::to_string_pretty(&t).unwrap(),
            r#"{
  "Record": {
    "type": "Extension",
    "head": {
      "k": "int",
      "v": "Int"
    },
    "tail": {
      "Record": {
        "type": "Extension",
        "head": {
          "k": "sweet",
          "v": {
            "Var": 0
          }
        },
        "tail": {
          "Record": {
            "type": "Extension",
            "head": {
              "k": "str",
              "v": "String"
            },
            "tail": {
              "Var": 1
            }
          }
        }
      }
    }
  }
}"#
        );
    }

    #[test]
    fn find_var_ref_non_row_type() {
        let source = r#"
vint = v + 2
"#;
        let mut p = Parser::new(&source);
        let pkg: ast::Package = p.parse_file("".to_string()).into();
        let t = find_var_type(pkg, "v".into()).expect("Should be able to get a MonoType.");
        assert_eq!(t, MonoType::Int);

        assert_eq!(serde_json::to_string_pretty(&t).unwrap(), "\"Int\"");
    }

    #[test]
    fn find_var_ref_obj_with() {
        let source = r#"
vint = v.int + 2
o = {v with x: 256}
p = o.ethan
"#;
        let mut p = Parser::new(&source);
        let pkg: ast::Package = p.parse_file("".to_string()).into();
        let mut t = find_var_type(pkg, "v".into()).expect("Should be able to get a MonoType.");
        let mut v = MonoTypeNormalizer::new();
        v.normalize(&mut t);
        assert_eq!(format!("{}", t), "{int:int | ethan:t0 | t1}");

        assert_eq!(
            serde_json::to_string_pretty(&t).unwrap(),
            r#"{
  "Record": {
    "type": "Extension",
    "head": {
      "k": "int",
      "v": "Int"
    },
    "tail": {
      "Record": {
        "type": "Extension",
        "head": {
          "k": "ethan",
          "v": {
            "Var": 0
          }
        },
        "tail": {
          "Var": 1
        }
      }
    }
  }
}"#
        );
    }

    #[test]
    fn find_var_ref_query() {
        // Test the find_var_type() function with some calls to stdlib functions.
        let source = r#"
from(bucket: v.bucket)
|> range(start: v.timeRangeStart, stop: v.timeRangeStop)
|> filter(fn: (r) => r._measurement == v.measurement or r._measurement == "cpu")
|> filter(fn: (r) => r.host == "host.local")
|> aggregateWindow(every: 30s, fn: count)
"#;
        let mut p = Parser::new(&source);
        let pkg: ast::Package = p.parse_file("".to_string()).into();
        let mut ty = find_var_type(pkg, "v".to_string()).expect("should be able to find var type");
        let mut v = MonoTypeNormalizer::new();
        v.normalize(&mut ty);
        assert_eq!(
            format!("{}", ty),
            "{measurement:t0 | timeRangeStart:t1 | timeRangeStop:t2 | bucket:string | t3}"
        );
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
        let code = "[{ C with
                _value: A
                    , _value: A
                    , _time: time
                    , _measurement: string
                    , _field: string
                    , region: B
                    }] where A: Addable, B: Equatable ";
        let mut p = parser::Parser::new(code);

        let typ_expr = p.parse_type_expression();
        let err = get_err_type_expression(typ_expr.clone());

        if err != "" {
            let msg = format!("TypeExpression parsing failed. {:?}", err);
            panic!(msg)
        }
        let want = convert_polytype(typ_expr, &mut Fresher::default()).unwrap();

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
        let code = "[{ D with
                _value: A
                    , A: B
                    , _time: time
                    , _measurement: string
                    , _field: string
                    }] where B: Equatable ";
        let mut p = parser::Parser::new(code);

        let typ_expr = p.parse_type_expression();
        let err = get_err_type_expression(typ_expr.clone());

        if err != "" {
            let msg = format!("TypeExpression parsing failed for {:?}", err);
            panic!(msg)
        }
        let want_a = convert_polytype(typ_expr, &mut Fresher::default()).unwrap();

        let code = " [{ D with
                _value: A
                    , B: B
                    , _time: time
                    , _measurement: string
                    , _field: string
                    }] where  B: Equatable";

        let mut p = parser::Parser::new(code);

        let typ_expr = p.parse_type_expression();
        let err = get_err_type_expression(typ_expr.clone());

        if err != "" {
            let msg = format!("TypeExpression parsing failed for {:?}", err);
            panic!(msg)
        }
        let want_b = convert_polytype(typ_expr, &mut Fresher::default()).unwrap();

        let code = "[{ D with
                _value: A
                    , A: B
                    , B: C
                    , _time: time
                    , _measurement: string
                    , _field: string
                    }] where B: Equatable, C: Equatable ";
        let mut p = parser::Parser::new(code);

        let typ_expr = p.parse_type_expression();
        let err = get_err_type_expression(typ_expr.clone());

        if err != "" {
            let msg = format!("TypeExpression parsing failed for {:?}", err);
            panic!(msg)
        }
        let want_c = convert_polytype(typ_expr, &mut Fresher::default()).unwrap();

        print!("want_a ==> {}\n", want_a);
        print!("want_b ==> {}\n", want_b);
        print!("want_c ==> {}\n", want_c);

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
