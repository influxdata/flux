#![cfg_attr(feature = "strict", deny(warnings, missing_docs))]

//! This module provides the public facing API for Flux's Go runtime, including formatting,
//! parsing, and standard library analysis.

extern crate fluxcore;
extern crate serde_aux;

extern crate serde_derive;

#[cfg(test)]
#[macro_use]
extern crate pretty_assertions;

use std::{ffi::*, mem, os::raw::c_char};

use anyhow::anyhow;
use once_cell::sync::Lazy;
use thiserror::Error;

pub use fluxcore::{ast, formatter, scanner, semantic, *};
use fluxcore::{
    parser::Parser,
    semantic::{
        env::Environment,
        flatbuffers::{
            semantic_generated::fbsemantic as fb,
            types::{build_env, build_type},
        },
        import::{Importer, Packages},
        nodes::{Package, Symbol},
        sub::Substitution,
        types::{MonoType, PolyType, TvarKinds},
        Analyzer, AnalyzerConfig, PackageExports,
    },
};

/// Result type for flux
pub type Result<T, E = Error> = std::result::Result<T, E>;

/// Error type for flux
#[derive(Error, Debug)]
pub enum Error {
    /// Semantic error
    #[error(transparent)]
    Semantic(#[from] semantic::FileErrors),

    /// Other errors that do not have a dedicated variant
    #[error(transparent)]
    Other(#[from] anyhow::Error),
}

use crate::semantic::flatbuffers::semantic_generated::fbsemantic::MonoTypeHolderArgs;

/// Prelude are the names and types of values that are inscope in all Flux scripts.
pub fn prelude() -> Option<PackageExports> {
    let buf = include_bytes!(concat!(env!("OUT_DIR"), "/prelude.data"));
    flatbuffers::root::<fb::TypeEnvironment>(buf)
        .unwrap()
        .into()
}

static PRELUDE: Lazy<Option<PackageExports>> = Lazy::new(prelude);

/// Imports is a map of import path to types of packages.
pub fn imports() -> Option<Packages> {
    let buf = include_bytes!(concat!(env!("OUT_DIR"), "/stdlib.data"));
    flatbuffers::root::<fb::Packages>(buf).unwrap().into()
}

/// Creates a new analyzer that can semantically analyze Flux source code.
///
/// The analyzer is aware of the stdlib and prelude.
pub fn new_semantic_analyzer(config: AnalyzerConfig) -> Result<Analyzer<'static, Packages>> {
    let env = match &*PRELUDE {
        Some(prelude) => prelude,
        None => return Err(anyhow!("missing prelude").into()),
    };
    let importer = match imports() {
        Some(imports) => imports,
        None => return Err(anyhow!("missing stdlib inports").into()),
    };
    Ok(Analyzer::new(Environment::from(env), importer, config))
}

/// An error handle designed to allow passing `Error` instances to library
/// consumers across language boundaries.
pub struct ErrorHandle {
    /// A heap-allocated `Error` message
    message: CString,

    /// The actual error
    err: Error,
}

impl From<Error> for Box<ErrorHandle> {
    fn from(err: Error) -> Self {
        Box::new(ErrorHandle {
            message: CString::new(format!("{}", err)).unwrap(),
            err,
        })
    }
}

/// Frees a previously allocated error.
///
/// ## Memory layout
///
/// We use the memory layout pattern described in the [`std::boxed`] module,
/// wherein a pointer where ownership is being transferred is modeled as a [`Box`], and if it could be
/// null, then it's wrapped in an [`Option`].
///
/// [`std::boxed`]: https://doc.rust-lang.org/std/boxed/index.html#memory-layout
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
    let pkg = parse(fname, &src);
    Box::new(pkg)
}

/// Parse the contents of a string.
pub fn parse(fname: String, src: &str) -> ast::Package {
    let mut p = Parser::new(src);
    p.parse_file(fname).into()
}

/// Format the Flux AST.
#[no_mangle]
pub extern "C" fn flux_ast_format(
    ast_pkg: &ast::Package,
    out: &mut flux_buffer_t,
) -> Option<Box<ErrorHandle>> {
    let mut out_str = String::new();
    for file in &ast_pkg.files {
        let s = match formatter::convert_to_string(file) {
            Ok(v) => v,
            Err(e) => return Some(Error::from(e).into()),
        };
        out_str.push_str(&s);
    }

    let len = out_str.len();
    let cstr = match CString::new(out_str) {
        Ok(bytes) => bytes,
        Err(e) => return Some(Error::from(anyhow::Error::from(e)).into()),
    };
    out.data = cstr.into_raw() as *mut u8;
    out.len = len;
    None
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
    match ast::check::check(ast_pkg) {
        Err(e) => Some(Error::from(anyhow::Error::from(e)).into()),
        Ok(_) => None,
    }
}

/// Frees an AST package.
///
/// ## Memory layout
///
/// We use the memory layout pattern described in the [`std::boxed`] module,
/// wherein a pointer where ownership is being transferred is modeled as a [`Box`], and if it could be
/// null, then it's wrapped in an [`Option`].
///
/// [`std::boxed`]: https://doc.rust-lang.org/std/boxed/index.html#memory-layout
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
        Err(err) => Some(Error::from(anyhow::Error::from(err)).into()),
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
            return Some(Error::from(anyhow::Error::from(err)).into());
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
            return Some(Error::from(err).into());
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
    let (mut vec, offset) = match semantic::flatbuffers::serialize_pkg(sem_pkg) {
        Ok(vec_offset) => vec_offset,
        Err(err) => {
            return Some(Error::from(err).into());
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
pub unsafe extern "C" fn flux_error_str(errh: &ErrorHandle) -> *const c_char {
    errh.message.as_ptr()
}

/// flux_error_print prints the error message associated with the given error to stdout.
///
/// # Safety
///
/// This function is unsafe because it dereferences a raw pointer passed as a
/// parameter
#[no_mangle]
pub unsafe extern "C" fn flux_error_print(errh: &ErrorHandle) {
    match &errh.err {
        Error::Semantic(err) => err.print(),
        Error::Other(err) => println!("{}", err),
    }
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
        Ok(_) => None,
        Err(e) => Some(Error::from(e).into()),
    }
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
        Err(err) => Some(err.into()),
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
        |e| Some(Box::from(e)),
        |t| {
            let mut builder = flatbuffers::FlatBufferBuilder::new();
            let (fb_mono_type, typ_type) = build_type(&mut builder, &t);
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

fn new_stateful_analyzer() -> Result<StatefulAnalyzer> {
    let env = match prelude() {
        Some(prelude) => prelude,
        None => return Err(anyhow!("missing prelude").into()),
    };
    let imports = match imports() {
        Some(imports) => imports,
        None => return Err(anyhow!("missing stdlib inports").into()),
    };
    Ok(StatefulAnalyzer { env, imports })
}

/// StatefulAnalyzer updates its environment with the contents of any previously analyzed package.
/// This enables uses cases where analysis is performed iteratively, for example in a REPL.
pub struct StatefulAnalyzer {
    env: PackageExports,
    imports: Packages,
}

impl StatefulAnalyzer {
    fn analyze(&mut self, ast_pkg: ast::Package) -> Result<fluxcore::semantic::nodes::Package> {
        let mut analyzer =
            Analyzer::new_with_defaults(Environment::from(&self.env), mem::take(&mut self.imports));
        let (mut env, sem_pkg) = match analyzer.analyze_ast(ast_pkg) {
            Ok(r) => r,
            Err(e) => {
                // In the face of an error we need to get the imports
                // back from the analyzer.
                let (_env, imports) = analyzer.drop();
                self.imports = imports;
                return Err(e.into());
            }
        };
        // Restore the imports.
        // We restore the env below.
        let (_, imports) = analyzer.drop();
        self.imports = imports;

        // Re-export any imported names into the env.
        // Normally we do not do this but we need to remember
        // any previous import statements since
        // each line of source is analyzed independently.
        for file in &sem_pkg.files {
            for dec in &file.imports {
                let path = &dec.path.value;

                // A failure should have already happened if any of these
                // imports would have failed.
                if let Some(typ) = self.imports.import(path) {
                    env.add(dec.import_symbol.clone(), typ);
                }
            }
        }
        self.env.copy_bindings_from(&env);
        Ok(sem_pkg)
    }
}

/// Create a new semantic analyzer.
///
/// # Safety
///
/// Ths function is unsafe because it dereferences a raw pointer.
#[no_mangle]
pub unsafe extern "C" fn flux_new_stateful_analyzer() -> Box<Result<StatefulAnalyzer>> {
    Box::new(new_stateful_analyzer())
}

/// Free a previously allocated semantic analyzer
#[no_mangle]
pub extern "C" fn flux_free_stateful_analyzer(_: Option<Box<Result<StatefulAnalyzer>>>) {}

/// # Safety
///
/// Ths function is unsafe because it dereferences a raw pointer.
#[no_mangle]
#[allow(clippy::boxed_local)]
pub unsafe extern "C" fn flux_analyze_with(
    analyzer: *mut Result<StatefulAnalyzer>,
    csrc: *const c_char,
    ast_pkg: Box<ast::Package>,
    out_sem_pkg: *mut Option<Box<semantic::nodes::Package>>,
) -> Option<Box<ErrorHandle>> {
    let ast_pkg = *ast_pkg;
    let analyzer = &mut *analyzer;
    let analyzer = match analyzer {
        Ok(a) => a,
        Err(_) => {
            match mem::replace(
                analyzer,
                Err(Error::from(anyhow!("The error has already been return!"))),
            ) {
                Err(err) => {
                    return Some(err.into());
                }
                Ok(_) => unreachable!(),
            }
        }
    };

    let src = if csrc.is_null() {
        None
    } else {
        Some(std::str::from_utf8(CStr::from_ptr(csrc).to_bytes()).unwrap())
    };

    let sem_pkg = Box::new(match analyzer.analyze(ast_pkg) {
        Ok(sem_pkg) => sem_pkg,
        Err(mut err) => {
            if let Some(src) = src {
                if let Error::Semantic(err) = &mut err {
                    err.source = Some(src.into());
                }
            }
            return Some(err.into());
        }
    });

    *out_sem_pkg = Some(sem_pkg);
    None
}

/// analyze consumes the given AST package and returns a semantic package
/// that has been type-inferred.  This function is aware of the standard library
/// and prelude.
pub fn analyze(ast_pkg: ast::Package) -> Result<Package> {
    let mut analyzer = new_semantic_analyzer(AnalyzerConfig::default())?;
    let (_, sem_pkg) = analyzer.analyze_ast(ast_pkg)?;
    Ok(sem_pkg)
}

/// infer_with_env consumes the given AST package, inject the type bindings from the given
/// type environment, and returns a semantic package that has not been type-injected and an
/// inferred type environment and substitution.
/// This function is aware of the standard library and prelude.
pub fn infer_with_env(
    ast_pkg: ast::Package,
    mut sub: Substitution,
    env: Option<Environment<'static>>,
) -> Result<(Environment<'static>, Package)> {
    let prelude = match &*PRELUDE {
        Some(prelude) => prelude,
        None => return Err(anyhow!("missing prelude").into()),
    };
    let env = if let Some(mut e) = env {
        e.external = Some(prelude);
        e
    } else {
        Environment::from(prelude)
    };
    let importer = match imports() {
        Some(imports) => imports,
        None => return Err(anyhow!("missing stdlib inports").into()),
    };
    let mut analyzer = Analyzer::new_with_defaults(env, importer);
    let (_, pkg) = analyzer.analyze_ast_with_substitution(ast_pkg, &mut sub)?;
    let (env, _) = analyzer.drop();
    Ok((env, pkg))
}

/// Given a Flux source and a variable name, find out the type of that variable in the Flux source code.
/// A type variable will be automatically generated and injected into the type environment that
/// will be used in semantic analysis. The Flux source code itself should not contain any definition
/// for that variable.
/// This version of find_var_type is aware of the prelude and builtins.
pub fn find_var_type(ast_pkg: ast::Package, var_name: String) -> Result<MonoType> {
    let sub = Substitution::default();
    let tvar = sub.fresh();
    let mut env = Environment::empty(true);
    let var_name = Symbol::from(var_name);
    env.add(
        var_name.clone(),
        PolyType {
            vars: Vec::new(),
            cons: TvarKinds::new(),
            expr: MonoType::Var(tvar),
        },
    );
    infer_with_env(ast_pkg, sub, Some(Environment::new(env)))
        .map(|(env, _)| env.lookup(&var_name).unwrap().expr.clone())
}

/// # Safety
///
/// This function is unsafe because it dereferences a raw pointer.
#[no_mangle]
pub unsafe extern "C" fn flux_get_env_stdlib(buf: *mut flux_buffer_t) {
    let imports = imports().unwrap();
    let env = PackageExports::try_from(
        imports
            .into_iter()
            .map(|(k, v)| (Symbol::from(k), v.typ()))
            .collect::<Vec<_>>(),
    )
    .unwrap();
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
    use fluxcore::{
        ast,
        parser::Parser,
        semantic::{
            convert::convert_polytype,
            fresh::Fresher,
            sub::Substitution,
            types::{Label, MonoType, Property, Ptr, Record, Tvar, TvarMap},
        },
    };

    use super::{new_semantic_analyzer, AnalyzerConfig};
    use crate::{analyze, find_var_type, flux_ast_get_error, parser};

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
                    self.normalize(&mut Ptr::make_mut(arr).0);
                }
                MonoType::Record(r) => {
                    if let Record::Extension { head, tail } = Ptr::make_mut(r) {
                        self.normalize(&mut head.v);
                        self.normalize(tail);
                    }
                }
                MonoType::Fun(f) => {
                    let f = Ptr::make_mut(f);
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
        let mut ty = MonoType::from(Record::new(
            [
                Property {
                    k: Label::from("a"),
                    v: MonoType::Var(Tvar(4949)),
                },
                Property {
                    k: Label::from("b"),
                    v: MonoType::Var(Tvar(4949)),
                },
                Property {
                    k: Label::from("e"),
                    v: MonoType::Var(Tvar(4957)),
                },
                Property {
                    k: Label::from("f"),
                    v: MonoType::Var(Tvar(4957)),
                },
                Property {
                    k: Label::from("g"),
                    v: MonoType::Var(Tvar(4957)),
                },
            ],
            Some(MonoType::Var(Tvar(4972))),
        ));
        assert_eq!(
            format!("{}", ty),
            "{t4972 with a:t4949, b:t4949, e:t4957, f:t4957, g:t4957}"
        );
        let mut v = MonoTypeNormalizer::new();
        v.normalize(&mut ty);
        assert_eq!(format!("{}", ty), "{C with a:A, b:A, e:B, f:B, g:B}");
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
        assert_eq!(format!("{}", t), "{B with int:int, sweet:A, str:string}");

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
        assert_eq!(t, MonoType::INT);

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
        assert_eq!(format!("{}", t), "{B with int:int, ethan:A}");

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
            "{D with measurement:A, timeRangeStart:B, timeRangeStop:C, bucket:string}"
        );
    }

    #[test]
    fn test_ast_get_error() {
        let ast = crate::parser::parse_string("test".to_string(), "x = 3 + / 10 - \"");
        let ast = Box::into_raw(Box::new(ast.into()));
        let errh = unsafe { flux_ast_get_error(ast) };

        expect_test::expect![[r#"
            error test@1:9-1:10: invalid expression: invalid token for primary expression: DIV

            error test@1:16-1:17: got unexpected token in string expression test@1:17-1:17: EOF"#]]
        .assert_eq(&errh.unwrap().message.into_string().unwrap());
    }

    #[test]
    fn deserialize_and_infer() {
        let mut analyzer = new_semantic_analyzer(AnalyzerConfig::default()).unwrap();

        let src = r#"
            x = from(bucket: "b")
                |> filter(fn: (r) => r.region == "west")
                |> map(fn: (r) => ({r with _value: r._value + r._value}))
        "#;

        let (got, _) = analyzer
            .analyze_source("".to_string(), "main.flux".to_string(), src)
            .unwrap();

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
        if let Err(err) = ast::check::check(ast::walk::Node::TypeExpression(&typ_expr)) {
            panic!("TypeExpression parsing failed. {:?}", err);
        }
        let want = convert_polytype(typ_expr, &mut Substitution::default()).unwrap();

        assert_eq!(want, got.lookup("x").expect("'x' not found").clone());
    }

    #[test]
    fn infer_union() {
        let mut analyzer = new_semantic_analyzer(AnalyzerConfig::default()).unwrap();

        let src = r#"
            a = from(bucket: "b")
                |> filter(fn: (r) => r.A == "A")
            b = from(bucket: "b")
                |> filter(fn: (r) => r.B == "B")
            c = union(tables: [a, b])
        "#;

        let (got, _) = analyzer
            .analyze_source("".to_string(), "main.flux".to_string(), src)
            .unwrap();

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
        if let Err(err) = ast::check::check(ast::walk::Node::TypeExpression(&typ_expr)) {
            panic!("TypeExpression parsing failed for {:?}", err);
        }
        let want_a = convert_polytype(typ_expr, &mut Substitution::default()).unwrap();

        let code = " [{ D with
                _value: A
                    , B: B
                    , _time: time
                    , _measurement: string
                    , _field: string
                    }] where  B: Equatable";

        let mut p = parser::Parser::new(code);

        let typ_expr = p.parse_type_expression();
        if let Err(err) = ast::check::check(ast::walk::Node::TypeExpression(&typ_expr)) {
            panic!("TypeExpression parsing failed for {:?}", err);
        }
        let want_b = convert_polytype(typ_expr, &mut Substitution::default()).unwrap();

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
        if let Err(err) = ast::check::check(ast::walk::Node::TypeExpression(&typ_expr)) {
            panic!("TypeExpression parsing failed for {:?}", err);
        }
        let want_c = convert_polytype(typ_expr, &mut Substitution::default()).unwrap();

        assert_eq!(want_a, got.lookup("a").expect("'a' not found").clone());
        assert_eq!(want_b, got.lookup("b").expect("'b' not found").clone());
        assert_eq!(want_c, got.lookup("c").expect("'c' not found").clone());
    }

    #[test]
    fn analyze_error() {
        let ast: ast::Package = fluxcore::parser::parse_string("".to_string(), "x = ()").into();
        match analyze(ast) {
            Ok(_) => panic!("expected an error, got none"),
            Err(e) => {
                expect_test::expect![[r#"
                    error @1:5-1:7: expected ARROW, got EOF

                    error @1:7-1:7: invalid expression: invalid token for primary expression: EOF"#]].assert_eq(&e.to_string());
            }
        }
    }
}
